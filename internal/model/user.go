package model

import "gorm.io/gorm"

// UserDetail 对应玩家详细资料 (Mai2UserDetail)
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

// UserOption 对应玩家游戏设置选项 (Mai2UserOption)
type UserOption struct {
	gorm.Model
	UserID      int64 `json:"userId"`
	JudgeDisp   int   `json:"judgeDisp"`
	NoteSpeed   int   `json:"noteSpeed"`
	TapDesign   int   `json:"tapDesign"`
	HoldDesign  int   `json:"holdDesign"`
	SlideDesign int   `json:"slideDesign"`
}

// UserExtend 对应玩家扩展数据 (Mai2UserExtend)
type UserExtend struct {
	gorm.Model
	UserID            int64 `json:"userId"`
	TotalSelect       int   `json:"totalSelect"`
	TotalPlayCount    int   `json:"totalPlayCount"`
	CardId            int64 `json:"cardId"`
	EventWatchedDate  string `json:"eventWatchedDate"`
}

// UserPlaylog 对应玩家游戏成绩记录 (Mai2UserPlaylog)
type UserPlaylog struct {
	gorm.Model
	UserID      int64  `json:"userId"`
	MusicID     int    `json:"musicId"`
	Level       int    `json:"level"`
	Achievement int    `json:"achievement"`
	Score       int    `json:"score"`
	CreateDate  string `json:"createDate"`
}

// UserCharacter 对应玩家角色数据 (Mai2UserCharacter)
type UserCharacter struct {
	gorm.Model
	UserID      int64 `json:"userId"`
	CharacterID int   `json:"characterId"`
	Level       int   `json:"level"`
	Count       int   `json:"count"`
	Awakening   int   `json:"awakening"`
}

// UserItem 对应玩家道具数据 (Mai2UserItem)
type UserItem struct {
	gorm.Model
	UserID   int64 `json:"userId"`
	ItemKind int   `json:"itemKind"`
	ItemID   int   `json:"itemId"`
	Stock    int   `json:"stock"`
}

// UserMap 对应玩家地图进度数据 (Mai2UserMap)
type UserMap struct {
	gorm.Model
	UserID     int64 `json:"userId"`
	MapID      int   `json:"mapId"`
	Distance   int   `json:"distance"`
	IsComplete bool  `json:"isComplete"`
}

// UserFavorite 对应玩家收藏项 (Mai2UserFavorite)
type UserFavorite struct {
	gorm.Model
	UserID   int64 `json:"userId"`
	ItemKind int   `json:"itemKind"`
	ItemID   int   `json:"itemId"`
}

// UpsertUserAllRequest 对应客户端全量上传请求结构
type UpsertUserAllRequest struct {
	UserID          int64           `json:"userId"`
	UpsertUserAll   struct {
		UserData          []UserDetail    `json:"userData"`
		UserOption        []UserOption    `json:"userOption"`
		UserExtend        []UserExtend    `json:"userExtend"`
		UserPlaylogList   []UserPlaylog   `json:"userPlaylogList"`
		UserCharacterList []UserCharacter `json:"userCharacterList"`
		UserItemList      []UserItem      `json:"userItemList"`
		UserMapList       []UserMap       `json:"userMapList"`
		UserFavoriteList  []UserFavorite  `json:"userFavoriteList"`
	} `json:"upsertUserAll"`
}
