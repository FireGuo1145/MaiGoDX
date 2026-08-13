package model

// Response 定义 maimai 接口通用返回结构
type Response struct {
	ReturnCode int         `json:"returnCode"`
	ApiName    string      `json:"apiName"`
	Data       interface{} `json:"data,omitempty"`
	Message    string      `json:"message,omitempty"`
}

// UserPreviewData 对应玩家预览数据
type UserPreviewData struct {
	UserID   int64  `json:"userId"`
	UserName string `json:"userName"`
	IsLogin  bool   `json:"isLogin"`
	LastData string `json:"lastData"`
}

// UserDetailData 对应玩家详细资料
type UserDetailData struct {
	UserID             int64  `json:"userId"`
	UserName           string `json:"userName"`
	EquipGlassesID     int    `json:"equipGlassesID"`
	EquipBackGroundID  int    `json:"equipBackGroundID"`
	EquipNamePlateID   int    `json:"equipNamePlateID"`
	EquipFrameID       int    `json:"equipFrameID"`
	EquipIconID        int    `json:"equipIconID"`
	Rating             int    `json:"rating"`
	MaxRating          int    `json:"maxRating"`
	TotalPoint         int64  `json:"totalPoint"`
}
