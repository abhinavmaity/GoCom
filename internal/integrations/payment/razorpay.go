// internal/integrations/payment/razorpay.go (shared with other teams)
package payment

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
)

type RazorpayClient struct {
	keyID     string
	keySecret string
	baseURL   string
}

func NewRazorpayClient() *RazorpayClient {
	return &RazorpayClient{
		keyID:     os.Getenv("RAZORPAY_KEY_ID"),
		keySecret: os.Getenv("RAZORPAY_KEY_SECRET"),
		baseURL:   "https://api.razorpay.com/v1",
	}
}

type CreateOrderRequest struct {
	Amount   int    `json:"amount"` // in paise
	Currency string `json:"currency"`
	Receipt  string `json:"receipt"`
}

type PaymentOrder struct {
	ID       string `json:"id"`
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
	Receipt  string `json:"receipt"`
	Status   string `json:"status"`
}

func (r *RazorpayClient) CreateOrder(amount int, receipt string) (*PaymentOrder, error) {
	req := CreateOrderRequest{
		Amount:   amount,
		Currency: "INR",
		Receipt:  receipt,
	}

	body, _ := json.Marshal(req)

	httpReq, _ := http.NewRequest("POST", r.baseURL+"/orders", bytes.NewReader(body))
	httpReq.SetBasicAuth(r.keyID, r.keySecret)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PaymentOrder
	json.NewDecoder(resp.Body).Decode(&result)
	return &result, nil
}

func (r *RazorpayClient) VerifyPayment(paymentID, orderID, signature string) bool {
	body := orderID + "|" + paymentID
	expectedSignature := r.generateSignature(body, r.keySecret)
	return expectedSignature == signature
}

func (r *RazorpayClient) generateSignature(body, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(body))
	return hex.EncodeToString(h.Sum(nil))
}
