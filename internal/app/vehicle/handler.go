package vehicle

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/response"
)

func CreateVehicle(vehicleService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var requestBody VehicleRequestBody
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(apperrors.ErrFailedMarshal.Error(), "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		vehicleData, err := vehicleService.CreateVehicle(ctx, requestBody)
		if err != nil {
			slog.Error("failed to create new vehicle", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "vehicle added successfully", vehicleData)
	}
}

func UpdateVehicle(vehicleService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vehicleId := r.PathValue("id")
		parsedVehicleId, err := strconv.Atoi(vehicleId)
		if err != nil {
			slog.Error("invalid vehicle id", "error", err)
			response.WriteJson(w, http.StatusBadRequest, "invalid vehicle id", nil)
			return
		}

		var requestBody VehicleRequestBody
		err = json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(apperrors.ErrFailedMarshal.Error(), "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		vehicleData, err := vehicleService.UpdateVehicle(ctx, requestBody, parsedVehicleId)
		if err != nil {
			slog.Error("failed to update vehicle", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "vehicle updated successfully", vehicleData)
	}
}

func SoftDeleteVehicle(vehicleService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vehicleId := r.PathValue("id")
		parsedVehicleId, err := strconv.Atoi(vehicleId)
		if err != nil {
			slog.Error("invalid vehicle id", "error", err)
			response.WriteJson(w, http.StatusBadRequest, "invalid vehicle id", nil)
			return
		}

		err = vehicleService.SoftDeleteVehicle(ctx, parsedVehicleId)
		if err != nil {
			slog.Error("failed to soft delete vehicle", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "vehicle deleted successfully", nil)
	}
}

func GenerateSignedVehicleImageUploadURL(vehicleService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		mimetype := r.URL.Query().Get("mimetype")

		signedUrl, accessUrl, err := vehicleService.GenerateSignedVehicleImageUploadURL(ctx, mimetype)
		if err != nil {
			slog.Error("failed to generate signed url for vehicle image upload", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		signedUrlResponse := GenerateSignedURLResponseBody{
			SignedUrl: signedUrl,
			AccessUrl: accessUrl,
		}
		response.WriteJson(w, http.StatusOK, "signed url generated successfully", signedUrlResponse)
	}
}

func GetVehicleById(vehicleService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vehicleId := r.PathValue("id")
		parsedVehicleId, err := strconv.ParseInt(vehicleId, 10, 64)
		if err != nil {
			slog.Error("invalid vehicle id", "error", err)
			response.WriteJson(w, http.StatusBadRequest, "invalid vehicle id", nil)
			return
		}

		vehicle, err := vehicleService.GetVehicleById(ctx, int(parsedVehicleId))
		if err != nil {
			slog.Error("failed to fetch vehicle", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "vehicle fetched successfully", vehicle)
	}
}

func GetVehicles(vehicleService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		page, err := parseQueryParamToInt(r, "page", 1)
		if err != nil {
			slog.Error("failed to parse page number to int", "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidQueryParams.Error(), nil)
			return
		}

		limit, err := parseQueryParamToInt(r, "limit", 10)
		if err != nil {
			slog.Error("failed to parse page limit to int", "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidQueryParams.Error(), nil)
			return
		}

		city := r.URL.Query().Get("city")
		if city == "" {
			slog.Error("no city provided in query", "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidQueryParams.Error(), nil)
			return
		}

		pickup, dropoff, err := parsePickupDropoffTimeStamp(r)
		if err != nil {
			slog.Error("failed to pickup/dropoff timestamp", "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidQueryParams.Error(), nil)
			return
		}

		params := GetVehiclesParams{
			City:             city,
			PickupTimestamp:  pickup,
			DropoffTimestamp: dropoff,
			Page:             page,
			Limit:            limit,
		}
		vehicles, err := vehicleService.GetVehicles(ctx, params)
		if err != nil {
			slog.Error("failed to fetch vehicles", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "vehicles fetched successfully", vehicles)
	}
}

func GetVehiclesForHost(vehicleService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		page, err := parseQueryParamToInt(r, "page", 1)
		if err != nil {
			slog.Error("failed to parse page number to int", "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidQueryParams.Error(), nil)
			return
		}

		limit, err := parseQueryParamToInt(r, "limit", 10)
		if err != nil {
			slog.Error("failed to parse page limit to int", "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidQueryParams.Error(), nil)
			return
		}

		vehicles, err := vehicleService.GetVehiclesForHost(ctx, int(page), int(limit))
		if err != nil {
			slog.Error("failed to fetch vehicles", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "vehicles fetched successfully", vehicles)
	}
}
