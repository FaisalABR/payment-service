package controllers

import (
	"net/http"
	errValidation "payment-service/common/error"
	"payment-service/common/response"
	"payment-service/domain/dto"
	"payment-service/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PaymentController struct {
	services services.IServiceRegistry
}

type IPaymentController interface {
	GetAllWithPagination(*gin.Context)
	GetByUUID(*gin.Context)
	Create(*gin.Context)
	Webhook(*gin.Context)
}

func NewPaymentController(services services.IServiceRegistry) IPaymentController {
	return &PaymentController{
		services: services,
	}
}

func (p *PaymentController) GetAllWithPagination(c *gin.Context) {
	var param dto.PaymentRequestParam
	err := c.ShouldBindQuery(&param)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code:  http.StatusBadRequest,
			Error: err,
			Gin:   c,
		})
		return
	}

	validate := validator.New()
	err = validate.Struct(param)
	if err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errResponse := errValidation.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Code:    http.StatusBadRequest,
			Message: &errMessage,
			Error:   err,
			Data:    errResponse,
			Gin:     c,
		})
		return
	}

	results, err := p.services.GetPayment().GetAllWithPagination(c, &param)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code:  http.StatusInternalServerError,
			Error: err,
			Gin:   c,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: results,
		Gin:  c,
	})
}

func (p *PaymentController) GetByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	result, err := p.services.GetPayment().GetByUUID(c, uuid)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code:  http.StatusBadRequest,
			Error: err,
			Gin:   c,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  c,
	})
}

func (p *PaymentController) Create(c *gin.Context) {
	var req dto.PaymentRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code:  http.StatusBadRequest,
			Error: err,
			Gin:   c,
		})
		return
	}

	validate := validator.New()
	err = validate.Struct(req)
	if err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errResponse := errValidation.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Code:    http.StatusBadRequest,
			Message: &errMessage,
			Data:    errResponse,
			Error:   err,
			Gin:     c,
		})
		return
	}

	result, err := p.services.GetPayment().Create(c, &req)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code:  http.StatusInternalServerError,
			Error: err,
			Gin:   c,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusCreated,
		Data: result,
		Gin:  c,
	})
}

func (p *PaymentController) Webhook(c *gin.Context) {
	var req dto.Webhook
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code:  http.StatusBadRequest,
			Error: err,
			Gin:   c,
		})
		return
	}

	validate := validator.New()
	err = validate.Struct(req)
	if err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errResponse := errValidation.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Code:    http.StatusBadRequest,
			Message: &errMessage,
			Error:   err,
			Data:    errResponse,
			Gin:     c,
		})
		return
	}

	err = p.services.GetPayment().Webhook(c, &req)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code:  http.StatusInternalServerError,
			Error: err,
			Gin:   c,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Gin:  c,
	})
}
