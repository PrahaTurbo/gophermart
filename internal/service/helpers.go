package service

import (
	"context"
	"fmt"

	"github.com/PrahaTurbo/gophermart/internal/auth"
	"github.com/pkg/errors"
)

const (
	OrderNew        = "NEW"
	OrderRegistered = "REGISTERED"
	OrderInvalid    = "INVALID"
	OrderProcessing = "PROCESSING"
	OrderProcessed  = "PROCESSED"
)

const amountMultiplier = 100

var (
	ErrInvalidOrderID     = errors.New("order id didn't pass luhn algorithm validation")
	ErrOrderByAnotherUser = errors.New("order was uploaded by another user")
	ErrOrderByCurrentUser = errors.New("order was uploaded by current user")

	ErrBalanceNotEnough = errors.New("not enough funds on the balance")
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
		return 0, fmt.Errorf("cannot extract userID from context")
	}

	return userID, nil
}

func amountToDecimalString(amount int) float64 {
	const amountDivider = 100

	return float64(amount) / amountDivider
}
