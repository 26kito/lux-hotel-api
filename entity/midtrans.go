package entity

type MidtransResponse struct {
	StatusCode        string `json:"status_code"`
	StatusMessage     string `json:"status_message"`
	TransactionID     string `json:"transaction_id"`
	OrderID           string `json:"order_id"`
	MerchantID        string `json:"merchant_id"`
	GrossAmount       string `json:"gross_amount"`
	Currency          string `json:"currency"`
	PaymentType       string `json:"payment_type"`
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	ExpiryTime        string `json:"expiry_time"`
	VANumbers         []struct {
		Bank     string `json:"bank"`
		VANumber string `json:"va_number"`
	} `json:"va_numbers"`
}

type MidtransPaymentPayload struct {
	PaymentType       string `json:"payment_type"`
	TransactionDetail struct {
		OrderID     string  `json:"order_id"`
		GrossAmount float64 `json:"gross_amount"`
	} `json:"transaction_details"`
	CustomerDetail struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
	} `json:"customer_details"`
	ItemDetails []struct {
		ID       string  `json:"id"`
		Price    float64 `json:"price"`
		Quantity int     `json:"quantity"`
		Name     string  `json:"name"`
	} `json:"item_details"`
	BankTransfer struct {
		Bank string `json:"bank"`
	} `json:"bank_transfer"`
}

type MidtransCallbackResponse struct {
	VA []struct {
		VANumber string `json:"va_number"`
		Bank     string `json:"bank"`
	} `json:"va_numbers"`
	TransactionTime   string        `json:"transaction_time"`
	TransactionStatus string        `json:"transaction_status"`
	TransactionID     string        `json:"transaction_id"`
	StatusMessage     string        `json:"status_message"`
	StatusCode        string        `json:"status_code"`
	SignatureKey      string        `json:"signature_key"`
	SettlementTime    string        `json:"settlement_time"`
	PaymentType       string        `json:"payment_type"`
	PaymentAmounts    []interface{} `json:"payment_amounts"`
	OrderID           string        `json:"order_id"`
	MerchantID        string        `json:"merchant_id"`
	GrossAmount       string        `json:"gross_amount"`
	FraudStatus       string        `json:"fraud_status"`
	Currency          string        `json:"currency"`
}
