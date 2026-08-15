package model

import (
	"time"

	"gorm.io/gorm"
)

// Terminal is an ALL.Net arcade cabinet identified by its full Keychip serial.
type Terminal struct {
	gorm.Model
	KeychipID       string    `gorm:"uniqueIndex;size:32;not null" json:"keychipId"`
	LastSeenKeychip string    `gorm:"size:32" json:"lastSeenKeychip"`
	Name            string    `gorm:"size:80" json:"name"`
	GameID          string    `gorm:"size:8;not null;default:SDEZ" json:"gameId"`
	GameVersion     string    `gorm:"size:24" json:"gameVersion"`
	OwnerAccountID  uint      `gorm:"index" json:"ownerAccountId"`
	IsEnabled       bool      `gorm:"not null;default:true" json:"isEnabled"`
	LastSeenAt      time.Time `json:"lastSeenAt"`
	LastSeenIP      string    `gorm:"size:64" json:"lastSeenIp"`
}

// TerminalSession records the authenticated ALL.Net PowerOn session used by /gs routes.
type TerminalSession struct {
	Token      string    `gorm:"primaryKey;size:64" json:"-"`
	TerminalID uint      `gorm:"index;not null" json:"terminalId"`
	GameID     string    `gorm:"size:8;not null" json:"gameId"`
	ExpiresAt  time.Time `gorm:"index;not null" json:"expiresAt"`
	CreatedAt  time.Time `json:"createdAt"`
}
