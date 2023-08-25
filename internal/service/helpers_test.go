package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/PrahaTurbo/gophermart/internal/auth"
)

func Test_amountToFloat64(t *testing.T) {
	tests := []struct {
		name   string
		amount int
		want   float64
	}{
		{
			name:   "should convert amount 100 to 1.0",
			amount: 100,
			want:   1.0,
		},
		{
			name:   "should convert amount 200 to 2.0",
			amount: 200,
			want:   2.0,
		},
		{
			name:   "should convert amount 50 to 0.5",
			amount: 50,
			want:   0.5,
		},
		{
			name:   "should convert amount 0 to 0.0",
			amount: 0,
			want:   0.0,
		},
		{
			name:   "should convert amount 999 to 9.99",
			amount: 999,
			want:   9.99,
		},
		{
			name:   "should convert amount -100 to -1.0",
			amount: -100,
			want:   -1.0,
		},
		{
			name:   "should convert amount 123456789 to 1234567.89",
			amount: 123456789,
			want:   1234567.89,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := amountToFloat64(tt.amount)
			assert.Equal(t, tt.want, result)
		})
	}
}

func Test_amountToInt(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
		want   int
	}{
		{
			name:   "should convert amount 1.0 to 100",
			amount: 1.0,
			want:   100,
		},
		{
			name:   "should convert amount 2.0 to 200",
			amount: 2.0,
			want:   200,
		},
		{
			name:   "should convert amount 0.5 to 50",
			amount: 0.5,
			want:   50,
		},
		{
			name:   "should convert amount 0.0 to 0",
			amount: 0.0,
			want:   0,
		},
		{
			name:   "should convert amount 9.99 to 999",
			amount: 9.99,
			want:   999,
		},
		{
			name:   "should convert amount  -1.0 to -100",
			amount: -1.0,
			want:   -100,
		},
		{
			name:   "should convert amount 1234567.89 to 123456789",
			amount: 1234567.89,
			want:   123456789,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := amountToInt(tt.amount)
			assert.Equal(t, tt.want, result)
		})
	}
}

func Test_extractUserIDFromCtx(t *testing.T) {
	type badContextKey string
	var badKey badContextKey = "jwt_token"

	tests := []struct {
		name    string
		ctx     context.Context
		want    int
		wantErr error
	}{
		{
			name: "should return valid user ID",
			ctx:  context.WithValue(context.Background(), auth.UserIDKey, 123),
			want: 123,
		},
		{
			name:    "should return error if invalid user ID ",
			ctx:     context.WithValue(context.Background(), auth.UserIDKey, "abc"),
			wantErr: ErrExtractFromContext,
		},
		{
			name:    "should return error if invalid key",
			ctx:     context.WithValue(context.Background(), badKey, 123),
			wantErr: ErrExtractFromContext,
		},
		{
			name:    "should return error if missing user ID value",
			ctx:     context.Background(),
			wantErr: ErrExtractFromContext,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractUserIDFromCtx(tt.ctx)

			if err != nil {
				assert.Equal(t, tt.wantErr, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
