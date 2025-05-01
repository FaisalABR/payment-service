package controllers

import (
	controllerPayment "payment-service/controllers/http/payment"
	"payment-service/services"
)

type Registry struct {
	services services.IServiceRegistry
}

type IControllerRegistry interface {
	GetPayment() controllerPayment.IPaymentController
}

func NewControllerRegistry(services services.IServiceRegistry) IControllerRegistry {
	return &Registry{
		services: services,
	}
}

func (r *Registry) GetPayment() controllerPayment.IPaymentController {
	return controllerPayment.NewPaymentController(r.services)
}
