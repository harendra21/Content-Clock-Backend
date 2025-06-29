package models

import (
	"time"

	"gorm.io/gorm"
)

type Connections struct {
	gorm.Model
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	UserId         string    `gorm:"column:user_id;not null;size:255"`
	Name           string    `gorm:"column:name;not null;size:255"`
	Username       string    `gorm:"column:username;size:255"`
	ConnectionName string    `gorm:"column:connection_name;not null;size:255"`
	ConnectionId   string    `gorm:"column:connection_id;not null;size:255"`
	AccessToken    string    `gorm:"column:access_token;not null;size:1024"`
	RefreshToken   string    `gorm:"column:refresh_token;size:1024"`
	MetaData       string    `gorm:"column:meta_data;size:2048"`
	ProfileImage   string    `gorm:"column:profile_image;size:1024"`
	Timezone       string    `gorm:"column:timezone;size:255"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoCreateTime;autoUpdateTime"`
	DeletedAt      *time.Time
}
