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

// Payment processes a payment order for a user.
// @Summary Process a payment order
// @Description Processes a payment order, requiring a valid JWT token for authentication. The request body should contain payment details.
// @Tags payment
// @Accept json
// @Produce json
// @Param payment_request body entity.PaymentPayload true "Payment details"
// @Security ApiKeyAuth
// @Success 200 {object} entity.ResponseOK "Payment processed successfully"
// @Failure 400 {object} entity.ResponseError "Invalid request"
// @Failure 401 {object} entity.ResponseError "Unauthorized access"
// @Failure 500 {object} entity.ResponseError "Internal server error"
// @Router /api/order/payment [post]
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
		Data:    response,
	})
}
