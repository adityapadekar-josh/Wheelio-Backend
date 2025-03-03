package booking

import (
	"fmt"
	"strings"
	"time"
)

const (
	// Booking Status
	Scheduled  = "SCHEDULED"
	CheckedOut = "CHECKED_OUT"
	Returned   = "RETURNED"
	Cancelled  = "CANCELLED"
)

const (
	checkoutOtpEmailContent = "Hello %s,\n\nThank you for choosing Wheelio! To proceed with your vehicle checkout, please provide the following OTP to the vehicle owner:\n\nOTP: %s\n\nThis OTP will expire on %s.\n\nEnsure you share this OTP with the owner before the expiration time to complete the rental process.\n\nBest regards,\nThe Wheelio Team"
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
	ScheduledPickupTime   time.Time `json:"scheduledPickupTime"`
	ScheduledDropoffTime  time.Time `json:"scheduledDropoffTime"`
}

type OtpToken struct {
	Id        int
	BookingId int
	Otp       string
	ExpiresAt time.Time
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

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UTC()

	if c.ScheduledPickupTime.Before(today) {
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
