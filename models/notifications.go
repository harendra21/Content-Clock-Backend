package models

import (
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
	"gorm.io/gorm"
)

type Notifications struct {
	gorm.Model
	ID         uint       `gorm:"primaryKey;autoIncrement"`
	User       string     `gorm:"column:user;not null;size:255"`
	Post       string     `gorm:"column:post;size:255"`
	Connection string     `gorm:"column:connection;size:255"`
	Type       string     `gorm:"column:type;size:255"`
	Title      string     `gorm:"column:title;size:255"`
	Message    string     `gorm:"column:message;type:text"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime"`
	ReadAt     *time.Time `gorm:"column:read_at"`
	DeletedAt  *time.Time `gorm:"column:deleted"`
}

func ApplyNotificationsCollectionSchema(c *core.Collection) {
	c.Fields.Add(
		&core.TextField{Name: "user"},
		&core.TextField{Name: "post"},
		&core.TextField{Name: "connection"},
		&core.TextField{Name: "type"},
		&core.TextField{Name: "title"},
		&core.TextField{Name: "message"},
		&core.DateField{Name: "created_at"},
		&core.DateField{Name: "read_at"},
		&core.DateField{Name: "deleted"},
	)

	ownRule := `@request.auth.id != "" && user = @request.auth.id`
	c.ListRule = types.Pointer(ownRule)
	c.ViewRule = types.Pointer(ownRule)
	c.CreateRule = types.Pointer(ownRule)
	c.UpdateRule = types.Pointer(ownRule)
	c.DeleteRule = types.Pointer(ownRule)
}
