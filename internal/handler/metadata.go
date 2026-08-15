package handler

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
	"gorm.io/gorm"
)

type metadataXML struct {
	DataName string `xml:"dataName"`
	SortList []struct {
		ID   string `xml:"id"`
		Name string `xml:"str"`
	} `xml:"SortList>StringID"`
}

type metadataPayload struct {
	DataName string                   `json:"dataName"`
	Items    []model.SiteMetadataItem `json:"items"`
}

var metadataNames = map[string]string{"music": "歌曲列表", "partner": "搭档列表", "ticket": "功能票列表", "chara": "旅行伙伴列表"}

func normalizeMetadataName(v string) (string, bool) {
	v = strings.ToLower(strings.TrimSpace(v))
	_, ok := metadataNames[v]
	return v, ok
}
func metadataJSON(w http.ResponseWriter, status int, value interface{}) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func metadataError(w http.ResponseWriter, status int, message string) {
	metadataJSON(w, status, map[string]interface{}{"success": false, "message": message})
}

func HandleAdminMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	if r.Method == http.MethodGet {
		var rows []model.SiteMetadata
		if err := database.DB.Order("data_name asc, item_id asc").Find(&rows).Error; err != nil {
			metadataError(w, 500, "读取元数据失败")
			return
		}
		result := map[string][]model.SiteMetadataItem{}
		for name := range metadataNames {
			result[name] = []model.SiteMetadataItem{}
		}
		for _, row := range rows {
			result[row.DataName] = append(result[row.DataName], model.SiteMetadataItem{ID: row.ItemID, Name: row.Name})
		}
		metadataJSON(w, 200, map[string]interface{}{"success": true, "metadata": result, "types": metadataNames})
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var payload metadataPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		metadataError(w, 400, "请求参数错误")
		return
	}
	replaceMetadata(w, payload)
}

func HandleAdminMetadataImport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(16 << 20); err != nil {
		metadataError(w, 400, "上传文件无效")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		metadataError(w, 400, "请选择 XML 文件")
		return
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, 16<<20))
	if err != nil {
		metadataError(w, 400, "读取 XML 文件失败")
		return
	}
	var parsed metadataXML
	if err := xml.Unmarshal(raw, &parsed); err != nil {
		metadataError(w, 400, "XML 格式解析失败")
		return
	}
	name, ok := normalizeMetadataName(parsed.DataName)
	if !ok {
		metadataError(w, 400, "不支持的 dataName，仅支持 music、partner、ticket、chara")
		return
	}
	items := make([]model.SiteMetadataItem, 0, len(parsed.SortList))
	seen := map[int64]bool{}
	for _, item := range parsed.SortList {
		id, idErr := strconv.ParseInt(strings.TrimSpace(item.ID), 10, 64)
		label := strings.TrimSpace(item.Name)
		if idErr != nil || id < 0 || label == "" {
			metadataError(w, 400, "XML 中存在无效的 ID 或名称")
			return
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		items = append(items, model.SiteMetadataItem{ID: id, Name: label})
	}
	if len(items) == 0 {
		metadataError(w, 400, "XML 未包含有效的 StringID 条目")
		return
	}
	replaceMetadata(w, metadataPayload{DataName: name, Items: items})
}

func replaceMetadata(w http.ResponseWriter, payload metadataPayload) {
	name, ok := normalizeMetadataName(payload.DataName)
	if !ok {
		metadataError(w, 400, "不支持的 dataName")
		return
	}
	seen := map[int64]bool{}
	clean := make([]model.SiteMetadata, 0, len(payload.Items))
	for _, item := range payload.Items {
		label := strings.TrimSpace(item.Name)
		if item.ID < 0 || label == "" {
			metadataError(w, 400, "元数据 ID 必须为非负数且名称不能为空")
			return
		}
		if seen[item.ID] {
			continue
		}
		seen[item.ID] = true
		clean = append(clean, model.SiteMetadata{DataName: name, ItemID: item.ID, Name: label})
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("data_name = ?", name).Delete(&model.SiteMetadata{}).Error; err != nil {
			return err
		}
		if len(clean) > 0 {
			return tx.Create(&clean).Error
		}
		return nil
	})
	if err != nil {
		metadataError(w, 500, "保存元数据失败")
		return
	}
	metadataJSON(w, 200, map[string]interface{}{"success": true, "message": metadataNames[name] + " 已保存", "count": len(clean)})
}

// HandleGetSiteMetadata exposes display-only mappings to the portal.
func HandleGetSiteMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var rows []model.SiteMetadata
	if err := database.DB.Order("data_name asc, item_id asc").Find(&rows).Error; err != nil {
		metadataError(w, 500, "读取元数据失败")
		return
	}
	result := map[string][]model.SiteMetadataItem{}
	for name := range metadataNames {
		result[name] = []model.SiteMetadataItem{}
	}
	for _, row := range rows {
		result[row.DataName] = append(result[row.DataName], model.SiteMetadataItem{ID: row.ItemID, Name: row.Name})
	}
	metadataJSON(w, 200, map[string]interface{}{"success": true, "metadata": result, "types": metadataNames})
}
