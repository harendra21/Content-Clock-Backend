package models

import (
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
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

func ApplyPostsCollectionSchema(c *core.Collection) {
	c.Fields.Add(
		&core.TextField{Name: "title"},
		&core.TextField{Name: "content"},
		&core.TextField{Name: "link"},
		&core.FileField{Name: "images", MaxSelect: 10},
		&core.TextField{Name: "status"},
		&core.TextField{Name: "type"},
		&core.TextField{Name: "group_id"},
		&core.TextField{Name: "logs"},
		&core.TextField{Name: "published_post_id"},
		&core.TextField{Name: "connection"},
		&core.TextField{Name: "user"},
		&core.DateField{Name: "publish_at"},
		&core.DateField{Name: "deleted"},
	)

	ownRule := `@request.auth.id != "" && user = @request.auth.id`
	c.ListRule = types.Pointer(ownRule)
	c.ViewRule = types.Pointer(ownRule)
	c.CreateRule = types.Pointer(ownRule)
	c.UpdateRule = types.Pointer(ownRule)
	c.DeleteRule = types.Pointer(ownRule)
}
