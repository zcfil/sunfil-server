package global

import (
	"time"

	"gorm.io/gorm"
)

type ZC_MODEL struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP(3)"`
	UpdatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP(3)"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
