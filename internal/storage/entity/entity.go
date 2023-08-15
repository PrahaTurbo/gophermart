package entity

import "time"

type User struct {
	ID           int
	Login        string
	PasswordHash string
}

type Balance struct {
	UserID    int
	Current   int
	Withdrawn int
}

type Order struct {
	ID         string
	UserID     int
	Accrual    int
	Status     string
	UploadedAt time.Time
}

type Withdraw struct {
	UserID      int
	OrderID     string
	Sum         int
	ProcessedAt time.Time
}
