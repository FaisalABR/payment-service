package routes

import (
	"payment-service/clients"
	"payment-service/constants"
	controllers "payment-service/controllers/http"
	"payment-service/middlewares"

	"github.com/gin-gonic/gin"
)

type PaymentRoutes struct {
	controller controllers.IControllerRegistry
	client     clients.IClientRegistry
	group      *gin.RouterGroup
}

type IPaymentRoutes interface {
	Run()
}

func NewPaymentRoutes(
	controller controllers.IControllerRegistry,
	client clients.IClientRegistry,
	group *gin.RouterGroup,
) IPaymentRoutes {
	return &PaymentRoutes{
		controller: controller,
		client:     client,
		group:      group,
	}
}

func (p *PaymentRoutes) Run() {
	group := p.group.Group("/payment")
	group.POST("/webhook", p.controller.GetPayment().Webhook)
	group.Use(middlewares.Authenticate())
	group.GET("", middlewares.CheckRole(
		[]string{
			constants.Admin,
			constants.Customer,
		}, p.client),
		p.controller.GetPayment().GetAllWithPagination)
	group.GET("/:uuid", middlewares.CheckRole(
		[]string{
			constants.Admin,
			constants.Customer,
		}, p.client),
		p.controller.GetPayment().GetByUUID)
	group.POST("", middlewares.CheckRole(
		[]string{
			constants.Customer,
		}, p.client),
		p.controller.GetPayment().Create)
}
