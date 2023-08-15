package models

import (
	"github.com/shopspring/decimal"
)

type UserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type OrderResponse struct {
	ID         string          `json:"number"`
	Accrual    decimal.Decimal `json:"accrual,omitempty"`
	Status     string          `json:"status,omitempty"`
	UploadedAt string          `json:"uploaded_at"`
}

type AccrualResponse struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual int    `json:"accrual"`
}

type BalanceResponse struct {
	Current   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

type WithdrawRequest struct {
	Order string `json:"order"`
	Sum   int    `json:"sum"`
}

type WithdrawalsResponse struct {
	Order       string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt string          `json:"processed_at"`
}
