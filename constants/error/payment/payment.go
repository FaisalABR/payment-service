package error

import "errors"

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrExpiredAt       = errors.New("expired time must be greater than current time")
)

var PaymentErrors = []error{
	ErrPaymentNotFound,
	ErrExpiredAt,
}
