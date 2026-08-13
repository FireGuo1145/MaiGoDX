package model

// UserDetail 对应 maimai DX 玩家详细资料
type UserDetail struct {
	ID                uint   `gorm:"primaryKey" json:"id"`
	UserID            int64  `json:"userId"`
	UserName          string `json:"userName"`
	EquipGlassesID    int    `json:"equipGlassesID"`
	EquipBackGroundID int    `json:"equipBackGroundID"`
	EquipNamePlateID  int    `json:"equipNamePlateID"`
	EquipFrameID      int    `json:"equipFrameID"`
	EquipIconID       int    `json:"equipIconID"`
	Rating            int    `json:"rating"`
	MaxRating         int    `json:"maxRating"`
	TotalPoint        int64  `json:"totalPoint"`
	IsNetMember       int    `json:"isNetMember"`
}

// UserOption 对应玩家游戏设置选项
type UserOption struct {
	ID          uint `gorm:"primaryKey" json:"id"`
	UserID      int64 `json:"userId"`
	JudgeDisp   int  `json:"judgeDisp"`
	NoteSpeed   int  `json:"noteSpeed"`
	TapDesign   int  `json:"tapDesign"`
	HoldDesign  int  `json:"holdDesign"`
	SlideDesign int  `json:"slideDesign"`
}

// UserPlaylog 对应玩家游戏成绩记录
type UserPlaylog struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	UserID     int64  `json:"userId"`
	MusicID    int    `json:"musicId"`
	Level      int    `json:"level"`
	Achievement int   `json:"achievement"`
	Score      int    `json:"score"`
	CreateDate string `json:"createDate"`
}

// UpsertUserAllRequest 对应客户端全量上传请求结构
type UpsertUserAllRequest struct {
	UserID       int64        `json:"userId"`
	UpsertUserAll struct {
		UserData       []UserDetail  `json:"userData"`
		UserOption     []UserOption  `json:"userOption"`
		UserPlaylogList []UserPlaylog `json:"userPlaylogList"`
	} `json:"upsertUserAll"`
}
