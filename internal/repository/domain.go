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

type Booking struct {
	Id                    int
	VehicleId             int
	HostId                int
	SeekerId              int
	Status                string
	PickupLocation        string
	DropoffLocation       string
	BookingAmount         float64
	OverdueFeeRatePerHour float64
	CancellationAllowed   bool
	ActualPickupTime      *time.Time
	ActualDropoffTime     *time.Time
	ScheduledPickupTime   time.Time
	ScheduledDropoffTime  time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type CreateBookingRequestBody struct {
	VehicleId             int
	HostId                int
	SeekerId              int
	Status                string
	PickupLocation        string
	DropoffLocation       string
	BookingAmount         float64
	OverdueFeeRatePerHour float64
	CancellationAllowed   bool
	ScheduledPickupTime   time.Time
	ScheduledDropoffTime  time.Time
}

type OtpToken struct {
	Id        int
	BookingId int
	Otp       string
	ExpiresAt time.Time
}

type Invoice struct {
	Id             int
	BookingId      int
	BookingAmount  float64
	AdditionalFees float64
	Tax            float64
	TaxRate        float64
	TotalAmount    float64
}

type BookingData struct {
	Id                      int
	Status                  string
	PickupLocation          string
	DropoffLocation         string
	BookingAmount           float64
	OverdueFeeRatePerHour   float64
	CancellationAllowed     bool
	ScheduledPickupTime     time.Time
	ScheduledDropoffTime    time.Time
	VehicleName             string
	VehicleSeatCount        int
	VehicleFuelType         string
	VehicleTransmissionType string
	VehicleImage string
}

type GetSeekerBookingsParams struct {
	SeekerId int
	Offset   int
	Limit    int
}

type GetHostBookingsParams struct {
	HostId int
	Offset int
	Limit  int
}

type BookingDetails struct {
	Id                    int
	Status                string
	PickupLocation        string
	DropoffLocation       string
	BookingAmount         float64
	OverdueFeeRatePerHour float64
	CancellationAllowed   bool
	ActualPickupTime      *time.Time
	ActualDropoffTime     *time.Time
	ScheduledPickupTime   time.Time
	ScheduledDropoffTime  time.Time
	Host                  BookingDetailsUser
	Seeker                BookingDetailsUser
	Vehicle               BookingDetailsVehicle
}

type BookingDetailsUser struct {
	Id          int
	Name        string
	Email       string
	PhoneNumber string
}

type BookingDetailsVehicle struct {
	Id                  int
	Name                string
	FuelType            string
	SeatCount           int
	TransmissionType    string
	Image    string
}
