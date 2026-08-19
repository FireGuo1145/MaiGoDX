package model

import "gorm.io/gorm"

// ChuniUser stores the latest CHUNITHM profile payload for an Aime user.
// The game carries version-specific fields, so the complete payload is kept
// alongside the fields needed for preview responses.
type ChuniUser struct {
	gorm.Model
	UserID       int64  `gorm:"uniqueIndex;not null"`
	UserName     string `gorm:"size:80"`
	PlayerRating int
	ProfileJSON  string `gorm:"type:text"`
}

type ChuniMusicDetail struct {
	gorm.Model
	UserID     int64  `gorm:"index:idx_chuni_music;not null"`
	MusicID    int    `gorm:"index:idx_chuni_music;not null"`
	Level      int    `gorm:"index:idx_chuni_music;not null"`
	DetailJSON string `gorm:"type:text"`
}

type ChuniPlaylog struct {
	gorm.Model
	UserID      int64 `gorm:"index;not null"`
	MusicID     int   `gorm:"index"`
	Level       int
	PlaylogJSON string `gorm:"type:text"`
}

// GameEvent is a globally configured maimai event returned by GetGameEventApi.
type GameEvent struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Type        int    `json:"type"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	DisableArea string `json:"disableArea"`
}

// GameCharge is a globally configured maimai charge item returned by GetGameChargeApi.
type GameCharge struct {
	ChargeID  int64  `gorm:"primaryKey" json:"chargeId"`
	OrderID   int64  `gorm:"index" json:"orderId"`
	Price     int    `json:"price"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

// GameSellingCard is a CardMaker card definition delivered by CMGetSellingCardApi.
// Its field names and date semantics match AquaDX Mai2GameSellingCard.
type GameSellingCard struct {
	CardID          int    `gorm:"primaryKey" json:"cardId"`
	StartDate       string `json:"startDate"`
	EndDate         string `json:"endDate"`
	NoticeStartDate string `json:"noticeStartDate"`
	NoticeEndDate   string `json:"noticeEndDate"`
}
