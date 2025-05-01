package dto

import (
	"time"

	"github.com/google/uuid"
)

type KafkaEvent struct {
	Name string `json:"name"`
}

type KafkaMetadata struct {
	Sender    string `json:"sender"`
	SendingAt string `json:"sendingAt"`
}

type KafkaData struct {
	OrderID   uuid.UUID  `json:"orderID"`
	PaymentID uuid.UUID  `json:"paymentID"`
	Status    string     `json:"status"`
	PaidAt    *time.Time `json:"paidAt"`
	ExpiredAt time.Time  `json:"expiredAt"`
}

type KafkaBody struct {
	Type string     `json:"type"`
	Data *KafkaData `json:"data"`
}

type KafkaMessage struct {
	Event    KafkaEvent    `json:"event"`
	Metadata KafkaMetadata `json:"metadata"`
	Body     KafkaBody     `json:"body"`
}
