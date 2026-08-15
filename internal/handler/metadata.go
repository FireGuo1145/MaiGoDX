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
	Skipped  int                      `json:"-"`
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
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		metadataError(w, 400, "上传文件无效")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		metadataError(w, 400, "请选择 XML 文件")
		return
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, 64<<20))
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
	itemsByID := make(map[int64]model.SiteMetadataItem, len(parsed.SortList))
	itemOrder := make([]int64, 0, len(parsed.SortList))
	skipped := 0
	for _, item := range parsed.SortList {
		idText := strings.TrimSpace(item.ID)
		label := strings.TrimSpace(item.Name)
		id, idErr := strconv.ParseInt(idText, 10, 64)
		// 空白占位项、不完整项和非法项跳过；重复 ID 保留最后一条名称。
		if idErr != nil || id < 0 || label == "" {
			skipped++
			continue
		}
		if _, exists := itemsByID[id]; !exists {
			itemOrder = append(itemOrder, id)
		} else {
			skipped++
		}
		itemsByID[id] = model.SiteMetadataItem{ID: id, Name: label}
	}
	items := make([]model.SiteMetadataItem, 0, len(itemOrder))
	for _, id := range itemOrder {
		items = append(items, itemsByID[id])
	}
	replaceMetadata(w, metadataPayload{DataName: name, Items: items, Skipped: skipped})
}

func replaceMetadata(w http.ResponseWriter, payload metadataPayload) {
	name, ok := normalizeMetadataName(payload.DataName)
	if !ok {
		metadataError(w, 400, "不支持的 dataName")
		return
	}
	lastByID := make(map[int64]model.SiteMetadata, len(payload.Items))
	itemOrder := make([]int64, 0, len(payload.Items))
	skipped := payload.Skipped
	for _, item := range payload.Items {
		label := strings.TrimSpace(item.Name)
		if item.ID < 0 || label == "" {
			metadataError(w, 400, "元数据 ID 必须为非负数且名称不能为空")
			return
		}
		if _, exists := lastByID[item.ID]; !exists {
			itemOrder = append(itemOrder, item.ID)
		} else {
			skipped++
		}
		lastByID[item.ID] = model.SiteMetadata{DataName: name, ItemID: item.ID, Name: label}
	}
	clean := make([]model.SiteMetadata, 0, len(itemOrder))
	for _, id := range itemOrder {
		clean = append(clean, lastByID[id])
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("data_name = ?", name).Delete(&model.SiteMetadata{}).Error; err != nil {
			return err
		}
		if len(clean) > 0 {
			return tx.CreateInBatches(clean, 500).Error
		}
		return nil
	})
	if err != nil {
		metadataError(w, 500, "保存元数据失败")
		return
	}
	metadataJSON(w, 200, map[string]interface{}{"success": true, "message": metadataNames[name] + " 已保存", "count": len(clean), "skipped": skipped})
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
