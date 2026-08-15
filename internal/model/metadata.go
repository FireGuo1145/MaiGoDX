package model

import "gorm.io/gorm"

type SiteMetadata struct {
	gorm.Model
	DataName string `gorm:"uniqueIndex:idx_site_metadata_name_id;not null" json:"dataName"`
	ItemID   int64  `gorm:"uniqueIndex:idx_site_metadata_name_id;not null" json:"id"`
	Name     string `gorm:"not null" json:"name"`
}

type SiteMetadataItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
