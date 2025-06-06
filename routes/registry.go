package routes

import (
	"payment-service/clients"
	controllers "payment-service/controllers/http"
	routes "payment-service/routes/payment"

	"github.com/gin-gonic/gin"
)

type Registry struct {
	controller controllers.IControllerRegistry
	client     clients.IClientRegistry
	group      *gin.RouterGroup
}

type IRoutesRegistry interface {
	Serve()
}

func NewRouteRegistry(
	controller controllers.IControllerRegistry,
	group *gin.RouterGroup,
	client clients.IClientRegistry,
) IRoutesRegistry {
	return &Registry{
		controller: controller,
		client:     client,
		group:      group,
	}
}

func (r *Registry) Serve() {
	r.paymentRoute().Run()
}

func (r *Registry) paymentRoute() routes.IPaymentRoutes {
	return routes.NewPaymentRoutes(r.controller, r.client, r.group)

}
