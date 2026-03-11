package models

import (
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
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

func ApplyConnectionsCollectionSchema(c *core.Collection) {
	c.Fields.Add(
		&core.TextField{Name: "name"},
		&core.TextField{Name: "username"},
		&core.TextField{Name: "connection_name"},
		&core.TextField{Name: "connection_id"},
		&core.TextField{Name: "access_token"},
		&core.TextField{Name: "refresh_token"},
		&core.JSONField{Name: "meta_data"},
		&core.TextField{Name: "timezone"},
		&core.TextField{Name: "user"},
		&core.TextField{Name: "profile_image_url"},
		&core.FileField{Name: "profile_image", MaxSelect: 1},
		&core.DateField{Name: "deleted"},
	)

	ownRule := `@request.auth.id != "" && user = @request.auth.id`
	c.ListRule = types.Pointer(ownRule)
	c.ViewRule = types.Pointer(ownRule)
	c.CreateRule = types.Pointer(ownRule)
	c.UpdateRule = types.Pointer(ownRule)
	c.DeleteRule = types.Pointer(ownRule)
}
