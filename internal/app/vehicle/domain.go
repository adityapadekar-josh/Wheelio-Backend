package vehicle

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository"
)

type VehicleWithImages struct {
	Id                    int             `json:"id"`
	Name                  string          `json:"name"`
	FuelType              string          `json:"fuelType"`
	SeatCount             int             `json:"seatCount"`
	TransmissionType      string          `json:"transmissionType"`
	Features              json.RawMessage `json:"features"`
	RatePerHour           float64         `json:"ratePerHour"`
	OverdueFeeRatePerHour float64         `json:"overdueFeeRatePerHour"`
	Address               string          `json:"address"`
	State                 string          `json:"state"`
	City                  string          `json:"city"`
	PinCode               int             `json:"pinCode"`
	CancellationAllowed   bool            `json:"cancellationAllowed"`
	Images                []VehicleImage  `json:"images"`
	Available             bool            `json:"available"`
	HostId                int             `json:"hostId"`
	IsDeleted             bool            `json:"isDeleted"`
	CreatedAt             time.Time       `json:"createdAt"`
	UpdatedAt             time.Time       `json:"updatedAt"`
}

type Vehicle struct {
	Id                    int             `json:"id"`
	Name                  string          `json:"name"`
	FuelType              string          `json:"fuelType"`
	SeatCount             int             `json:"seatCount"`
	TransmissionType      string          `json:"transmissionType"`
	Features              json.RawMessage `json:"features"`
	RatePerHour           float64         `json:"ratePerHour"`
	OverdueFeeRatePerHour float64         `json:"overdueFeeRatePerHour"`
	Address               string          `json:"address"`
	State                 string          `json:"state"`
	City                  string          `json:"city"`
	PinCode               int             `json:"pinCode"`
	CancellationAllowed   bool            `json:"cancellationAllowed"`
	Available             bool            `json:"available"`
	HostId                int             `json:"hostId"`
	IsDeleted             bool            `json:"isDeleted"`
	CreatedAt             time.Time       `json:"createdAt"`
	UpdatedAt             time.Time       `json:"updatedAt"`
}

type VehicleImage struct {
	Id        int       `json:"id"`
	VehicleId int       `json:"-"`
	Url       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

var AvailableFuelType = map[string]interface{}{
	"Petrol":   nil,
	"Diesel":   nil,
	"Electric": nil,
	"Hybrid":   nil,
}

var AvailableTransmissionType = map[string]interface{}{
	"Manual":    nil,
	"Automatic": nil,
}

func (v VehicleWithImages) validate() error {
	var validationErrors []string

	if strings.TrimSpace(v.Name) == "" {
		validationErrors = append(validationErrors, "name is required")
	}

	if strings.TrimSpace(v.FuelType) == "" {
		validationErrors = append(validationErrors, "fuel type is required")
	} else if _, ok := AvailableFuelType[v.FuelType]; !ok {
		validationErrors = append(validationErrors, "fuel type is invalid")
	}

	if v.SeatCount <= 0 {
		validationErrors = append(validationErrors, "seat count should be positive integer")
	}

	if strings.TrimSpace(v.TransmissionType) == "" {
		validationErrors = append(validationErrors, "transmission type is required")
	} else if _, ok := AvailableTransmissionType[v.TransmissionType]; !ok {
		validationErrors = append(validationErrors, "transmission type is invalid")
	}

	if v.RatePerHour < 0 {
		validationErrors = append(validationErrors, "rate per hour cannot be negative")
	}

	if v.OverdueFeeRatePerHour < 0 {
		validationErrors = append(validationErrors, "overdue fee rate per hour cannot be negative")
	}

	if strings.TrimSpace(v.Address) == "" {
		validationErrors = append(validationErrors, "address is required")
	}

	if strings.TrimSpace(v.State) == "" {
		validationErrors = append(validationErrors, "state is required")
	}

	if strings.TrimSpace(v.City) == "" {
		validationErrors = append(validationErrors, "city is required")
	}

	if v.PinCode < 100000 || v.PinCode > 999999 {
		validationErrors = append(validationErrors, "pin code must be a 6-digit integer")
	}

	if len(v.Images) > 0 {
		for i, img := range v.Images {
			if img.Id <= 0 {
				validationErrors = append(validationErrors, fmt.Sprintf("image at index %d must have a valid id", i))
			}
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

func MapVehicleWithImagesToVehicleRepo(vehicle VehicleWithImages) repository.Vehicle {
	mappedVehicle := repository.Vehicle{
		Id:                    vehicle.Id,
		Name:                  vehicle.Name,
		FuelType:              vehicle.FuelType,
		SeatCount:             vehicle.SeatCount,
		TransmissionType:      vehicle.TransmissionType,
		Features:              vehicle.Features,
		RatePerHour:           vehicle.RatePerHour,
		OverdueFeeRatePerHour: vehicle.OverdueFeeRatePerHour,
		Address:               vehicle.Address,
		State:                 vehicle.State,
		City:                  vehicle.City,
		PinCode:               vehicle.PinCode,
		CancellationAllowed:   vehicle.CancellationAllowed,
		Available:             vehicle.Available,
		HostId:                vehicle.HostId,
		IsDeleted:             vehicle.IsDeleted,
		CreatedAt:             vehicle.CreatedAt,
		UpdatedAt:             vehicle.UpdatedAt,
	}

	return mappedVehicle
}

func MapVehicleRepoAndVehicleImageRepoToVehicleWithImages(vehicle repository.Vehicle, images []repository.VehicleImage) VehicleWithImages {
	convertedImages := make([]VehicleImage, len(images))
	for i, img := range images {
		convertedImages[i] = VehicleImage(img)
	}

	mappedVehicle := VehicleWithImages{
		Id:                    vehicle.Id,
		Name:                  vehicle.Name,
		FuelType:              vehicle.FuelType,
		SeatCount:             vehicle.SeatCount,
		TransmissionType:      vehicle.TransmissionType,
		Features:              vehicle.Features,
		RatePerHour:           vehicle.RatePerHour,
		OverdueFeeRatePerHour: vehicle.OverdueFeeRatePerHour,
		Address:               vehicle.Address,
		State:                 vehicle.State,
		City:                  vehicle.City,
		PinCode:               vehicle.PinCode,
		CancellationAllowed:   vehicle.CancellationAllowed,
		Images:                convertedImages,
		Available:             vehicle.Available,
		HostId:                vehicle.HostId,
		IsDeleted:             vehicle.IsDeleted,
		CreatedAt:             vehicle.CreatedAt,
		UpdatedAt:             vehicle.UpdatedAt,
	}

	return mappedVehicle
}
