package models

import (
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
	"gorm.io/gorm"
)

type Analytics struct {
	gorm.Model
	ID        uint       `gorm:"primaryKey;autoIncrement"`
	Post      string     `gorm:"column:post;not null;size:255"`
	Data      string     `gorm:"column:data;type:text"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"column:deleted"`
}

func ApplyAnalyticsCollectionSchema(c *core.Collection) {
	c.Fields.Add(
		&core.TextField{Name: "post"},
		&core.JSONField{Name: "data"},
	)

	authReadRule := `@request.auth.id != ""`
	c.ListRule = types.Pointer(authReadRule)
	c.ViewRule = types.Pointer(authReadRule)
	c.CreateRule = nil
	c.UpdateRule = nil
	c.DeleteRule = nil
}
