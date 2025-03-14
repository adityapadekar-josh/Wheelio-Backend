package vehicle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository"
)

const (
	SignedURLExpiry = 15 * time.Minute
	AccessURLFormat = "https://firebasestorage.googleapis.com/v0/b/wheelio-2f2fa.firebasestorage.app/o/%s?alt=media"
)

var AvailableFuelType = map[string]struct{}{
	"Petrol":   {},
	"Diesel":   {},
	"Electric": {},
	"Hybrid":   {},
}

var AvailableTransmissionType = map[string]struct{}{
	"Manual":    {},
	"Automatic": {},
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
	Images                []VehicleImage  `json:"images,omitempty"`
	Available             bool            `json:"available"`
	HostId                int             `json:"hostId"`
	IsDeleted             bool            `json:"isDeleted"`
	CreatedAt             time.Time       `json:"createdAt"`
	UpdatedAt             time.Time       `json:"updatedAt"`
}

type VehicleImage struct {
	Id        int       `json:"id,omitempty"`
	VehicleId int       `json:"-"`
	Url       string    `json:"url"`
	Featured  bool      `json:"featured"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

type VehicleRequestBody struct {
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
	Images                []VehicleImage  `json:"images,omitempty"`
}

type GenerateSignedURLResponseBody struct {
	SignedUrl string `json:"signedUrl"`
	AccessUrl string `json:"accessUrl"`
}

type VehicleOverview struct {
	Id               int     `json:"id"`
	Name             string  `json:"name"`
	FuelType         string  `json:"fuelType"`
	SeatCount        int     `json:"seatCount"`
	TransmissionType string  `json:"transmissionType"`
	Image            string  `json:"image"`
	RatePerHour      float64 `json:"ratePerHour"`
	Address          string  `json:"address"`
	PinCode          int     `json:"pinCode"`
}

type PaginationParams struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalCount int `json:"totalCount"`
}

type PaginatedVehicleOverview struct {
	Data       []VehicleOverview `json:"data"`
	Pagination PaginationParams  `json:"pagination"`
}

type GetVehiclesParams struct {
	City             string
	PickupTimestamp  time.Time
	DropoffTimestamp time.Time
	Page             int
	Limit            int
}

func (v VehicleRequestBody) validate() error {
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

	if len(v.Images) == 0 {
		validationErrors = append(validationErrors, "at least one image is required")
	} else {
		featuredCount := 0
		for _, img := range v.Images {
			if img.Featured {
				featuredCount++
			}
		}
		if featuredCount != 1 {
			validationErrors = append(validationErrors, "exactly one image must have the featured flag set to true")
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

func mapVehicleRequestBodyToCreateUserRequestBodyRepo(vehicleRequestBody VehicleRequestBody) repository.CreateVehicleRequestBody {
	mappedVehicle := repository.CreateVehicleRequestBody{
		Name:                  vehicleRequestBody.Name,
		FuelType:              vehicleRequestBody.FuelType,
		SeatCount:             vehicleRequestBody.SeatCount,
		TransmissionType:      vehicleRequestBody.TransmissionType,
		Features:              vehicleRequestBody.Features,
		RatePerHour:           vehicleRequestBody.RatePerHour,
		OverdueFeeRatePerHour: vehicleRequestBody.OverdueFeeRatePerHour,
		Address:               vehicleRequestBody.Address,
		State:                 vehicleRequestBody.State,
		City:                  vehicleRequestBody.City,
		PinCode:               vehicleRequestBody.PinCode,
		CancellationAllowed:   vehicleRequestBody.CancellationAllowed,
	}

	return mappedVehicle
}

func mapVehicleRequestBodyToEditUserRequestBodyRepo(vehicleRequestBody VehicleRequestBody) repository.EditVehicleRequestBody {
	mappedVehicle := repository.EditVehicleRequestBody{
		Name:                  vehicleRequestBody.Name,
		FuelType:              vehicleRequestBody.FuelType,
		SeatCount:             vehicleRequestBody.SeatCount,
		TransmissionType:      vehicleRequestBody.TransmissionType,
		Features:              vehicleRequestBody.Features,
		RatePerHour:           vehicleRequestBody.RatePerHour,
		OverdueFeeRatePerHour: vehicleRequestBody.OverdueFeeRatePerHour,
		Address:               vehicleRequestBody.Address,
		State:                 vehicleRequestBody.State,
		City:                  vehicleRequestBody.City,
		PinCode:               vehicleRequestBody.PinCode,
		CancellationAllowed:   vehicleRequestBody.CancellationAllowed,
	}

	return mappedVehicle
}

func mapVehicleRepoAndVehicleImageRepoToVehicle(vehicle repository.Vehicle, images []repository.VehicleImage) Vehicle {
	convertedImages := make([]VehicleImage, len(images))
	for i, img := range images {
		convertedImages[i] = VehicleImage(img)
	}

	mappedVehicle := Vehicle{
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

func parsePickupDropoffTimeStamp(r *http.Request) (time.Time, time.Time, error) {
	pickupQuery := r.URL.Query().Get("pickup")
	dropoffQuery := r.URL.Query().Get("dropoff")

	if (pickupQuery == "" && dropoffQuery != "") || (pickupQuery != "" && dropoffQuery == "") {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid query parameters: pickup and dropoff must both be provided or both omitted")
	}

	today := time.Now().UTC()
	pickup := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	dropoff := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 999999999, time.UTC)
	timestampLayout := time.RFC3339Nano
	var err error

	if pickupQuery != "" {
		pickup, err = time.Parse(timestampLayout, pickupQuery)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("failed to parse pickup time: %w", err)
		}
		if pickup.Location() != time.UTC {
			return time.Time{}, time.Time{}, fmt.Errorf("pickup timestamp must be in UTC")
		}
	}

	if dropoffQuery != "" {
		dropoff, err = time.Parse(timestampLayout, dropoffQuery)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("failed to parse dropoff time: %w", err)
		}
		if dropoff.Location() != time.UTC {
			return time.Time{}, time.Time{}, fmt.Errorf("dropoff timestamp must be in UTC")
		}
	}

	return pickup, dropoff, nil
}
