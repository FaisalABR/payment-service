package repositories

import (
	paymentRepo "payment-service/repositories/payment"
	paymentHistoryRepo "payment-service/repositories/paymenthistory"

	"gorm.io/gorm"
)

type Registry struct {
	db *gorm.DB
}

type IRepositoryRegistry interface {
	GetPayment() paymentRepo.IPaymentRepository
	GetPaymentHistory() paymentHistoryRepo.IPaymentHistoryRepository
	GetTx() *gorm.DB
}

func NewRepositoryRegistry(db *gorm.DB) IRepositoryRegistry {
	return &Registry{db: db}
}

func (r *Registry) GetPayment() paymentRepo.IPaymentRepository {
	return paymentRepo.NewPaymentRepository(r.db)
}

func (r *Registry) GetPaymentHistory() paymentHistoryRepo.IPaymentHistoryRepository {
	return paymentHistoryRepo.NewPaymentHistoryRepository(r.db)
}

func (r *Registry) GetTx() *gorm.DB {
	return r.db
}
