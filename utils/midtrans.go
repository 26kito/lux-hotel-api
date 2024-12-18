package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"lux-hotel/entity"
	"os"

	"github.com/go-resty/resty/v2"
)

func MidtransPaymentHandler(payload entity.MidtransPaymentPayload) (*entity.MidtransResponse, error) {
	var response entity.MidtransResponse

	client := resty.New()

	midtransServerKey := os.Getenv("MIDTRANS_SERVER_KEY")
	encodedKey := base64.StdEncoding.EncodeToString([]byte(midtransServerKey))

	url := os.Getenv("MIDTRANS_BASE_URL") + "/charge"

	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Basic %s", encodedKey)).
		SetBody(payload).
		Post(url)

	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("500 | %v", err)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		log.Println("Error unmarshalling response:", err)
		return nil, err
	}

	return &response, nil
}

func MidtransTransactionStatusHandler(orderID string) (*entity.MidtransResponse, error) {
	var response entity.MidtransResponse

	client := resty.New()

	midtransServerKey := os.Getenv("MIDTRANS_SERVER_KEY")
	encodedKey := base64.StdEncoding.EncodeToString([]byte(midtransServerKey))

	url := os.Getenv("MIDTRANS_BASE_URL") + "/" + orderID + "/status"

	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Basic %s", encodedKey)).
		Get(url)

	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("500 | %v", err)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		log.Println("Error unmarshalling response:", err)
		return nil, err
	}

	return &response, nil
}
