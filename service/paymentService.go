package service

import (
	"lux-hotel/entity"
	"lux-hotel/repository"
	"strconv"

	"github.com/labstack/echo/v4"
)

type PaymentService interface {
	Payment(c echo.Context) error
}

type paymentService struct {
	PaymentRepository repository.PaymentRepository
}

func NewPaymentService(paymentRepository repository.PaymentRepository) PaymentService {
	return &paymentService{PaymentRepository: paymentRepository}
}

func (ps *paymentService) Payment(c echo.Context) error {
	var payload entity.PaymentPayload

	if err := c.Bind(&payload); err != nil {
		return c.JSON(400, entity.ResponseError{
			Status:  400,
			Message: "Invalid request",
		})
	}

	response, err := ps.PaymentRepository.Payment(payload)

	if err != nil {
		errCode, _ := strconv.Atoi(err.Error()[:3])
		errMessage := err.Error()[6:]

		return c.JSON(errCode, entity.ResponseError{
			Status:  errCode,
			Message: errMessage,
		})
	}

	return c.JSON(200, entity.ResponseOK{
		Status:  200,
		Message: "Success",
		Data: map[string]interface{}{
			"transaction_id": response.TransactionID,
			"status":         response.TransactionStatus,
			"amount":         response.GrossAmount,
			"payment_type":   response.PaymentType,
			"bank":           response.VANumbers[0].Bank,
			"va_number":      response.VANumbers[0].VANumber,
		},
	})
}
