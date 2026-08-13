package model

import "gorm.io/gorm"

// UserDetail 对应玩家详细资料
type UserDetail struct {
	gorm.Model
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
	gorm.Model
	UserID      int64 `json:"userId"`
	JudgeDisp   int   `json:"judgeDisp"`
	NoteSpeed   int   `json:"noteSpeed"`
	TapDesign   int   `json:"tapDesign"`
	HoldDesign  int   `json:"holdDesign"`
	SlideDesign int   `json:"slideDesign"`
}

// UserPlaylog 对应玩家游戏成绩记录
type UserPlaylog struct {
	gorm.Model
	UserID      int64  `json:"userId"`
	MusicID     int    `json:"musicId"`
	Level       int    `json:"level"`
	Achievement int    `json:"achievement"`
	Score       int    `json:"score"`
	CreateDate  string `json:"createDate"`
}

// UserCharacter 对应玩家角色数据
type UserCharacter struct {
	gorm.Model
	UserID      int64 `json:"userId"`
	CharacterID int   `json:"characterId"`
	Level       int   `json:"level"`
	Count       int   `json:"count"`
}

// UserItem 对应玩家道具数据
type UserItem struct {
	gorm.Model
	UserID   int64 `json:"userId"`
	ItemKind int   `json:"itemKind"`
	ItemID   int   `json:"itemId"`
	Stock    int   `json:"stock"`
}

// UserMap 对应玩家地图进度数据
type UserMap struct {
	gorm.Model
	UserID int64 `json:"userId"`
	MapID  int   `json:"mapId"`
	Distance int `json:"distance"`
	IsComplete bool `json:"isComplete"`
}

// UpsertUserAllRequest 对应客户端全量上传请求结构
type UpsertUserAllRequest struct {
	UserID          int64           `json:"userId"`
	UpsertUserAll   struct {
		UserData        []UserDetail    `json:"userData"`
		UserOption      []UserOption    `json:"userOption"`
		UserPlaylogList []UserPlaylog   `json:"userPlaylogList"`
		UserCharacterList []UserCharacter `json:"userCharacterList"`
		UserItemList    []UserItem      `json:"userItemList"`
		UserMapList     []UserMap       `json:"userMapList"`
	} `json:"upsertUserAll"`
}
