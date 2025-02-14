package repository

import (
	"encoding/json"
	"time"
)

type User struct {
	Id          int
	Name        string
	Email       string
	PhoneNumber string
	Password    string
	Role        string
	IsVerified  bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateUserRequestBody struct {
	Name        string
	Email       string
	PhoneNumber string
	Password    string
	Role        string
}

type VerificationToken struct {
	Id        int
	UserId    int
	Token     string
	Type      string
	ExpiresAt time.Time
}

type Vehicle struct {
	Id                    int
	Name                  string
	FuelType              string
	SeatCount             int
	TransmissionType      string
	Features              json.RawMessage
	RatePerHour           float64
	OverdueFeeRatePerHour float64
	Address               string
	State                 string
	City                  string
	PinCode               int
	CancellationAllowed   bool
	Available             bool
	HostId                int
	IsDeleted             bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type VehicleImage struct {
	Id        int
	VehicleId int
	Url       string
	Featured  bool
	CreatedAt time.Time
}
