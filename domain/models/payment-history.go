package models

import (
	"payment-service/constants"
	"time"
)

type PaymentHistory struct {
	ID        uint                          `gorm:"primaryKey;autoIncrement"`
	PaymentID uint                          `gorm:"type:bigint;not null"`
	Status    constants.PaymentStatusString `gorm:"type:varchar(30);not null"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}
