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

type CreateVehicleRequestBody struct {
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
	HostId                int
}

type EditVehicleRequestBody struct {
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
}

type CreateVehicleImageData struct {
	VehicleId int
	Url       string
	Featured  bool
}

type VehicleOverview struct {
	Id               int
	Name             string
	FuelType         string
	SeatCount        int
	TransmissionType string
	Image            string
	RatePerHour      float64
	Address          string
	PinCode          int
}

type GetVehiclesParams struct {
	City             string
	PickupTimestamp  time.Time
	DropoffTimestamp time.Time
	Offset           int
	Limit            int
}

type GetVehiclesForHostParams struct {
	HostId int
	Offset int
	Limit  int
}
