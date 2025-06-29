package models

import (
	"time"

	"gorm.io/gorm"
)

type SocialPosts struct {
	gorm.Model
	ID              uint      `gorm:"primaryKey;autoIncrement"`
	ConnectionId    string    `gorm:"not null;type:varchar(255)"`
	UserId          string    `gorm:"not null;type:varchar(255)"`
	Title           string    `gorm:"type:varchar(255)"`
	Description     string    `gorm:"type:text"`
	Link            string    `gorm:"type:varchar(255)"`
	Medias          string    `gorm:"type:varchar(255)"`
	Type            string    `gorm:"not null;type:varchar(255)"`
	PublishAt       string    `gorm:"column:publish_at;not null"`
	Status          string    `gorm:"not null;type:varchar(255)"`
	Logs            string    `gorm:"type:text"`
	PublishedPostId string    `gorm:"type:varchar(255)"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoCreateTime;autoUpdateTime"`
	DeletedAt       *time.Time
}
