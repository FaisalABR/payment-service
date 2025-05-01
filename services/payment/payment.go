package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	clients "payment-service/clients/midtrans"
	gcs "payment-service/common/gcs"
	"payment-service/common/util"
	configApp "payment-service/config"
	"payment-service/constants"
	errPayment "payment-service/constants/error/payment"
	"payment-service/controllers/kafka"
	"payment-service/domain/dto"
	"payment-service/domain/models"
	"payment-service/repositories"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PaymentService struct {
	repository repositories.IRepositoryRegistry
	gcs        gcs.IGCSClient
	kafka      kafka.IKafkaRegistry
	midtrans   clients.IMidtransClient
}

type IPaymentService interface {
	GetAllWithPagination(context.Context, *dto.PaymentRequestParam) (*util.PaginationResult, error)
	GetByUUID(context.Context, string) (*dto.PaymentResponse, error)
	Create(context.Context, *dto.PaymentRequest) (*dto.PaymentResponse, error)
	Webhook(context.Context, *dto.Webhook) error
}

func NewPaymentService(
	repository repositories.IRepositoryRegistry,
	gcs gcs.IGCSClient,
	kafka kafka.IKafkaRegistry,
	midtrans clients.IMidtransClient,
) IPaymentService {
	return &PaymentService{
		repository: repository,
		gcs:        gcs,
		kafka:      kafka,
		midtrans:   midtrans,
	}
}

func (s *PaymentService) GetAllWithPagination(
	ctx context.Context,
	param *dto.PaymentRequestParam,
) (*util.PaginationResult, error) {
	payments, total, err := s.repository.GetPayment().FindAllWithPagination(ctx, param)
	if err != nil {
		return nil, err
	}

	paymentResults := make([]dto.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		paymentResults = append(paymentResults, dto.PaymentResponse{
			UUID:          payment.UUID,
			OrderID:       payment.OrderID,
			Amount:        payment.Amount,
			Status:        payment.Status.GetStatusString(),
			PaymentLink:   payment.PaymentLink,
			InvoiceLink:   payment.InvoiceLink,
			TransactionID: payment.TransactionID,
			VANumber:      payment.VANumber,
			Bank:          payment.Bank,
			Acquirer:      payment.Acquirer,
			Description:   payment.Description,
			PaidAt:        payment.PaidAt,
			CreatedAt:     payment.CreatedAt,
			UpdatedAt:     payment.UpdatedAt,
			ExpiredAt:     payment.ExpiredAt,
		})
	}

	pagination := &util.PaginationParam{
		Page:  param.Page,
		Limit: param.Limit,
		Count: total,
		Data:  paymentResults,
	}

	response := util.GeneratePagination(*pagination)

	return &response, nil

}

func (s *PaymentService) GetByUUID(ctx context.Context, uuid string) (*dto.PaymentResponse, error) {
	payment, err := s.repository.GetPayment().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	response := dto.PaymentResponse{
		UUID:          payment.UUID,
		OrderID:       payment.OrderID,
		Amount:        payment.Amount,
		Status:        payment.Status.GetStatusString(),
		PaymentLink:   payment.PaymentLink,
		InvoiceLink:   payment.InvoiceLink,
		TransactionID: payment.TransactionID,
		VANumber:      payment.VANumber,
		Bank:          payment.Bank,
		Acquirer:      payment.Acquirer,
		Description:   payment.Description,
		PaidAt:        payment.PaidAt,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
		ExpiredAt:     payment.ExpiredAt,
	}

	return &response, nil
}

func (s *PaymentService) Create(ctx context.Context, req *dto.PaymentRequest) (*dto.PaymentResponse, error) {
	var (
		txErr, err error
		payment    *models.Payment
		response   *dto.PaymentResponse
		midtrans   *clients.MidtransData
	)

	err = s.repository.GetTx().Transaction(func(tx *gorm.DB) error {
		if !req.ExpiredAt.After(time.Now()) {
			return errPayment.ErrPaymentNotFound
		}

		midtrans, txErr = s.midtrans.CreatePaymentLink(req)
		if txErr != nil {
			return txErr
		}

		paymentRequest := &dto.PaymentRequest{
			OrderID:     req.OrderID,
			Amount:      req.Amount,
			Description: req.Description,
			ExpiredAt:   req.ExpiredAt,
			PaymentLink: midtrans.RedirectURL,
		}

		payment, txErr = s.repository.GetPayment().Create(ctx, tx, paymentRequest)
		if txErr != nil {
			return txErr
		}

		txErr = s.repository.GetPaymentHistory().Create(ctx, tx, &dto.PaymentHistoryRequest{
			PaymentID: uint(payment.ID),
			Status:    payment.Status.GetStatusString(),
		})
		if txErr != nil {
			return txErr
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	response = &dto.PaymentResponse{
		UUID:        payment.UUID,
		OrderID:     payment.OrderID,
		Amount:      payment.Amount,
		Status:      payment.Status.GetStatusString(),
		PaymentLink: payment.PaymentLink,
		Description: payment.Description,
	}

	return response, nil
}

// buat function convertToIndonesiaMonth
func (p *PaymentService) convertToIndonesiaMonth(englishMonth string) string {
	mapIndonesiaMonth := map[string]string{
		"January":   "Januari",
		"February":  "Februari",
		"March":     "Maret",
		"April":     "April",
		"May":       "Mei",
		"June":      "Juni",
		"July":      "Juli",
		"August":    "Agustus",
		"September": "September",
		"October":   "Oktober",
		"November":  "November",
		"December":  "Desember",
	}

	_, ok := mapIndonesiaMonth[englishMonth]
	if !ok {
		return errors.New("invalid month").Error()
	}

	return mapIndonesiaMonth[englishMonth]
}

func (p *PaymentService) generatePDF(req *dto.InvoiceRequest) ([]byte, error) {
	htmlTemplatePath := "template/invoice.html"
	htmlTemplate, err := os.ReadFile(htmlTemplatePath)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	jsonData, _ := json.Marshal(req)
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}

	pdf, err := util.GeneratePDFFromHTML(string(htmlTemplate), data)
	if err != nil {
		return nil, err
	}

	return pdf, nil
}

// buat function uploadToGCS
func (p *PaymentService) uploadToGCS(ctx context.Context, invoiceNumber string, pdf []byte) (string, error) {
	invoiceNumberReplace := strings.ReplaceAll(invoiceNumber, "/", "-")
	filename := fmt.Sprintf("%s.pdf", invoiceNumberReplace)
	url, err := p.gcs.UploadFile(ctx, filename, pdf)
	if err != nil {
		return "", err
	}
	return url, nil
}

// buat function randomNumber

func (p *PaymentService) randomNumber() int {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	number := random.Intn(900000) + 100000
	return number
}

func (p *PaymentService) mapTransactionStatusToEvent(status constants.PaymentStatusString) string {
	var paymentStatus string
	switch status {
	case constants.PendingString:
		paymentStatus = strings.ToUpper(string(constants.PendingString))
	case constants.SettlementString:
		paymentStatus = strings.ToUpper(string(constants.SettlementString))
	case constants.ExpiredString:
		paymentStatus = strings.ToUpper(string(constants.ExpiredString))
	}

	return paymentStatus
}

func (s *PaymentService) produceToKafka(
	req *dto.Webhook,
	payment *models.Payment,
	paidAt *time.Time,
) error {
	event := dto.KafkaEvent{
		Name: s.mapTransactionStatusToEvent(req.TransactionStatus),
	}

	metadata := dto.KafkaMetadata{
		Sender:    "payment-service",
		SendingAt: time.Now().Format(time.RFC3339),
	}

	body := dto.KafkaBody{
		Type: "JSON",
		Data: &dto.KafkaData{
			OrderID:   payment.OrderID,
			PaymentID: payment.UUID,
			Status:    string(req.TransactionStatus),
			PaidAt:    paidAt,
			ExpiredAt: *payment.ExpiredAt,
		},
	}

	kafkaMessage := dto.KafkaMessage{
		Event:    event,
		Metadata: metadata,
		Body:     body,
	}

	topic := configApp.Config.Kafka.Topic
	kafkaMessageJSON, _ := json.Marshal(kafkaMessage)
	err := s.kafka.GetKafkaProducer().Produce(topic, kafkaMessageJSON)
	if err != nil {
		return err
	}

	return nil
}

func (s *PaymentService) Webhook(ctx context.Context, req *dto.Webhook) error {
	var (
		txErr, err         error
		paymentAfterUpdate *models.Payment
		paidAt             *time.Time
		invoiceLink        string
		pdf                []byte
	)

	err = s.repository.GetTx().Transaction(func(tx *gorm.DB) error {
		_, txErr = s.repository.GetPayment().FindByOrderID(ctx, req.OrderID.String())
		if txErr != nil {
			return txErr
		}

		if req.TransactionStatus == constants.SettlementString {
			now := time.Now()
			paidAt = &now
		}

		status := req.TransactionStatus.GetStatus()
		vaNumber := req.VANumbers[0].VaNumber
		bank := req.VANumbers[0].Bank
		_, txErr = s.repository.GetPayment().Update(ctx, tx, req.OrderID.String(), &dto.UpdatePaymentRequest{
			TransactionID: &req.TransactionID,
			Status:        &status,
			VANumber:      &vaNumber,
			Bank:          &bank,
			Acquirer:      req.Acquirer,
		})
		if txErr != nil {
			return txErr
		}

		paymentAfterUpdate, txErr = s.repository.GetPayment().FindByOrderID(ctx, req.OrderID.String())
		if txErr != nil {
			return txErr
		}

		txErr = s.repository.GetPaymentHistory().Create(ctx, tx, &dto.PaymentHistoryRequest{
			PaymentID: uint(paymentAfterUpdate.ID),
			Status:    paymentAfterUpdate.Status.GetStatusString(),
		})

		if req.TransactionStatus == constants.SettlementString {
			paidDay := paidAt.Format("02")
			paidMonth := s.convertToIndonesiaMonth(paidAt.Format("January"))
			paidYear := paidAt.Format("2006")
			invoiceNumber := fmt.Sprintf("INV/%s/ORD/%d", time.Now().Format(time.DateOnly), s.randomNumber())
			total := util.FormatRupiah(&paymentAfterUpdate.Amount)
			invoiceRequest := &dto.InvoiceRequest{
				InvoiceNumber: invoiceNumber,
				Data: dto.InvoiceData{
					PaymentDetail: dto.InvoicePaymentDetail{
						BankName:      *paymentAfterUpdate.Bank,
						PaymentMethod: req.PaymentType,
						VANumber:      *paymentAfterUpdate.VANumber,
						Date:          fmt.Sprintf("%s %s %s", paidDay, paidMonth, paidYear),
						IsPaid:        true,
					},
					Items: []dto.InvoiceItem{
						{
							Description: *paymentAfterUpdate.Description,
							Price:       total,
						},
					},
					Total: total,
				},
			}

			pdf, txErr = s.generatePDF(invoiceRequest)
			if txErr != nil {
				return txErr
			}

			invoiceLink, txErr = s.uploadToGCS(ctx, invoiceNumber, pdf)
			if txErr != nil {
				return txErr
			}

			_, txErr = s.repository.GetPayment().Update(ctx, tx, req.OrderID.String(), &dto.UpdatePaymentRequest{
				InvoiceLink: &invoiceLink,
			})
			if txErr != nil {
				return txErr
			}

		}
		return nil

	})
	if err != nil {
		return nil
	}

	err = s.produceToKafka(req, paymentAfterUpdate, paidAt)
	if err != nil {
		return err
	}

	return nil
}
