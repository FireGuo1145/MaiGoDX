package model

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
