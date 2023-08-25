package service

import (
	"context"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/PrahaTurbo/gophermart/internal/auth"
)

const (
	OrderNew        = "NEW"
	OrderRegistered = "REGISTERED"
	OrderInvalid    = "INVALID"
	OrderProcessing = "PROCESSING"
	OrderProcessed  = "PROCESSED"
)

var (
	ErrInvalidOrderID     = errors.New("order id didn't pass luhn algorithm validation")
	ErrOrderByAnotherUser = errors.New("order was uploaded by another user")
	ErrOrderByCurrentUser = errors.New("order was uploaded by current user")

	ErrBalanceNotEnough = errors.New("not enough funds on the balance")

	ErrExtractFromContext = errors.New("cannot extract userID from context")
)

func validateLuhn(s string) bool {
	n := len(s)
	sum := 0
	double := false
	for i := n - 1; i >= 0; i-- {
		c := int(s[i] - '0')

		if double {
			c = c * 2
			if c > 9 {
				c = c - 9
			}
		}
		double = !double

		sum += c
	}

	return sum%10 == 0
}

func extractUserIDFromCtx(ctx context.Context) (int, error) {
	userIDVal := ctx.Value(auth.UserIDKey)
	userID, ok := userIDVal.(int)
	if !ok {
		return 0, ErrExtractFromContext
	}

	return userID, nil
}

func amountToFloat64(amount int) float64 {
	const amountDivider = 100

	d := decimal.NewFromInt(int64(amount)).Div(decimal.NewFromInt(amountDivider)).Round(2)
	result, ok := d.Float64()

	if !ok {
		return float64(amount) / amountDivider
	}

	return result
}

func amountToInt(amount float64) int {
	const amountMultiplier = 100

	d := decimal.NewFromFloat(amount).Mul(decimal.NewFromInt(amountMultiplier))
	n := d.IntPart()

	return int(n)
}
