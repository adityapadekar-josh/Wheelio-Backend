package booking

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository"
)

const (
	// Booking Status
	Scheduled  = "SCHEDULED"
	CheckedOut = "CHECKED_OUT"
	Returned   = "RETURNED"
	Cancelled  = "CANCELLED"

	// Tax rate
	taxRate = 0.18
)

const (
	checkoutOtpEmailContent       = "Hello %s,\n\nThank you for choosing Wheelio! To proceed with your vehicle checkout, please provide the following OTP to the vehicle owner:\n\nOTP: %s\n\nEnsure you share this OTP with the owner before the expiration time to complete the rental process.\n\nBest regards,\nThe Wheelio Team"
	initiateReturnOtpEmailContent = "Hello %s,\n\nThank you for choosing Wheelio! To proceed with your vehicle return, please provide the following OTP to the vehicle seeker:\n\nOTP: %s\n\nThis OTP will expire in 20 minutes.\n\nEnsure you share this OTP with the seeker before the expiration time to complete the vehicle return process.\n\nBest regards,\nThe Wheelio Team"
)

type Booking struct {
	Id                    int        `json:"id"`
	VehicleId             int        `json:"vehicleId"`
	HostId                int        `json:"hostId"`
	SeekerId              int        `json:"seekerId"`
	Status                string     `json:"status"`
	PickupLocation        string     `json:"pickupLocation"`
	DropoffLocation       string     `json:"dropoffLocation"`
	BookingAmount         float64    `json:"bookingAmount"`
	OverdueFeeRatePerHour float64    `json:"overdueFeeRatePerHour"`
	CancellationAllowed   bool       `json:"cancellationAllowed"`
	ActualPickupTime      *time.Time `json:"actualPickupTime,omitempty"`
	ActualDropoffTime     *time.Time `json:"actualDropoffTime,omitempty"`
	ScheduledPickupTime   time.Time  `json:"scheduledPickupTime"`
	ScheduledDropoffTime  time.Time  `json:"scheduledDropoffTime"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
}

type CreateBookingRequestBody struct {
	VehicleId             int       `json:"vehicleId"`
	HostId                int       `json:"hostId"`
	SeekerId              int       `json:"seekerId"`
	Status                string    `json:"status"`
	PickupLocation        string    `json:"pickupLocation"`
	DropoffLocation       string    `json:"dropoffLocation"`
	BookingAmount         float64   `json:"bookingAmount"`
	OverdueFeeRatePerHour float64   `json:"overdueFeeRatePerHour"`
	CancellationAllowed   bool      `json:"cancellationAllowed"`
	ScheduledPickupTime   time.Time `json:"scheduledPickupTime"`
	ScheduledDropoffTime  time.Time `json:"scheduledDropoffTime"`
}

type OtpToken struct {
	Id        int
	BookingId int
	Otp       string
	ExpiresAt time.Time
}

type OtpRequestBody struct {
	Otp string `json:"otp"`
}

type BookingData struct {
	Id                      int       `json:"id"`
	Status                  string    `json:"status"`
	PickupLocation          string    `json:"pickupLocation"`
	DropoffLocation         string    `json:"dropoffLocation"`
	BookingAmount           float64   `json:"bookingAmount"`
	OverdueFeeRatePerHour   float64   `json:"overdueFeeRatePerHour"`
	CancellationAllowed     bool      `json:"cancellationAllowed"`
	ScheduledPickupTime     time.Time `json:"scheduledPickupTime"`
	ScheduledDropoffTime    time.Time `json:"scheduledDropoffTime"`
	VehicleName             string    `json:"vehicleName"`
	VehicleSeatCount        int       `json:"vehicleSeatCount"`
	VehicleFuelType         string    `json:"vehicleFuelType"`
	VehicleTransmissionType string    `json:"vehicleTransmissionType"`
	VehicleImage            string    `json:"vehicleImage"`
}

type PaginationParams struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalCount int `json:"totalCount"`
}

type PaginatedBookingData struct {
	Data       []BookingData    `json:"data"`
	Pagination PaginationParams `json:"pagination"`
}

type BookingDetails struct {
	Id                    int                   `json:"id"`
	Status                string                `json:"status"`
	PickupLocation        string                `json:"pickupLocation"`
	DropoffLocation       string                `json:"dropoffLocation"`
	BookingAmount         float64               `json:"bookingAmount"`
	OverdueFeeRatePerHour float64               `json:"overdueFeeRatePerHour"`
	CancellationAllowed   bool                  `json:"cancellationAllowed"`
	ActualPickupTime      *time.Time            `json:"actualPickupTime,omitempty"`
	ActualDropoffTime     *time.Time            `json:"actualDropoffTime,omitempty"`
	ScheduledPickupTime   time.Time             `json:"scheduledPickupTime"`
	ScheduledDropoffTime  time.Time             `json:"scheduledDropoffTime"`
	Host                  BookingDetailsUser    `json:"host"`
	Seeker                BookingDetailsUser    `json:"seeker"`
	Vehicle               BookingDetailsVehicle `json:"vehicle"`
	Invoice               bookingDetailsInvoice `json:"invoice"`
}

type BookingDetailsUser struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
}

type BookingDetailsVehicle struct {
	Id               int    `json:"id"`
	Name             string `json:"name"`
	FuelType         string `json:"fuelType"`
	SeatCount        int    `json:"seatCount"`
	TransmissionType string `json:"transmissionType"`
	Image            string `json:"image"`
}

type bookingDetailsInvoice struct {
	Id             int     `json:"id"`
	AdditionalFees float64 `json:"additionalFees"`
	Tax            float64 `json:"tax"`
	TaxRate        float64 `json:"taxRate"`
	TotalAmount    float64 `json:"totalAmount"`
}

func (c CreateBookingRequestBody) validate() error {
	var validationErrors []string

	if c.VehicleId <= 0 {
		validationErrors = append(validationErrors, "vehicleId must be greater than 0")
	}

	if strings.TrimSpace(c.PickupLocation) == "" {
		validationErrors = append(validationErrors, "pickupLocation is required")
	}

	if strings.TrimSpace(c.DropoffLocation) == "" {
		validationErrors = append(validationErrors, "dropoffLocation is required")
	}

	if c.ScheduledPickupTime.IsZero() {
		validationErrors = append(validationErrors, "scheduledPickupTime is required")
	}

	now := time.Now().UTC()
	earliestAllowedTime := now.Add(-21 * time.Hour)

	if c.ScheduledPickupTime.Before(earliestAllowedTime) {
		validationErrors = append(validationErrors, "scheduledPickupTime must not be in past")
	}

	if c.ScheduledDropoffTime.IsZero() {
		validationErrors = append(validationErrors, "scheduledDropoffTime is required")
	}

	if !c.ScheduledPickupTime.IsZero() && !c.ScheduledDropoffTime.IsZero() && c.ScheduledPickupTime.After(c.ScheduledDropoffTime) {
		validationErrors = append(validationErrors, "scheduledPickupTime must be before scheduledDropoffTime")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

func parseQueryParamToInt(r *http.Request, param string, defaultValue int) (int, error) {
	query := r.URL.Query().Get(param)
	if query == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(query)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func mapBookingDetailsRepoToBookingDetails(bookingDetails repository.BookingDetails) BookingDetails {
	booking := BookingDetails{
		Id:                    bookingDetails.Id,
		Status:                bookingDetails.Status,
		PickupLocation:        bookingDetails.PickupLocation,
		DropoffLocation:       bookingDetails.DropoffLocation,
		BookingAmount:         bookingDetails.BookingAmount,
		OverdueFeeRatePerHour: bookingDetails.OverdueFeeRatePerHour,
		CancellationAllowed:   bookingDetails.CancellationAllowed,
		ActualPickupTime:      bookingDetails.ActualPickupTime,
		ActualDropoffTime:     bookingDetails.ActualDropoffTime,
		ScheduledPickupTime:   bookingDetails.ScheduledPickupTime,
		ScheduledDropoffTime:  bookingDetails.ScheduledDropoffTime,
		Host:                  BookingDetailsUser(bookingDetails.Host),
		Seeker:                BookingDetailsUser(bookingDetails.Seeker),
		Vehicle:               BookingDetailsVehicle(bookingDetails.Vehicle),
		Invoice:               bookingDetailsInvoice(bookingDetails.Invoice),
	}

	return booking
}
