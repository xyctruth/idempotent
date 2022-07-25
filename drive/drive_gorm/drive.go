package drive_gorm

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type drive struct {
	db *gorm.DB
}

func New(db *gorm.DB) *drive {
	return &drive{db: db}
}

func (e *drive) Init() error {
	return e.db.AutoMigrate(&Idempotent{})
}

func (e *drive) Acquire(key string, ttl time.Duration) (bool, error) {
	now := time.Now()
	r := &Idempotent{
		IdempotentKey: key,
		CreateAt:      now,
		ExpiryAt:      now.Add(ttl),
	}

	result := e.db.Clauses(clause.OnConflict{DoNothing: true}).Create(r)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return true, nil
	}

	result = e.db.Model(&Idempotent{}).
		Where("idempotent_key = ? AND expiry_at < ?", key, time.Now()).
		Update("expiry_at", r.ExpiryAt)

	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return true, nil
	}

	return false, nil
}

func (e *drive) Clear() error {
	return e.db.Where("expiry_at < ?", time.Now()).Delete(&Idempotent{}).Error
}

type Idempotent struct {
	IdempotentKey string    `gorm:"primaryKey; size:100; NOT NULL;"` //自增ID
	CreateAt      time.Time `gorm:"NOT NULL;" `                      //创建时间
	ExpiryAt      time.Time `gorm:"NOT NULL; index;"`                //过期时间
}
