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

		var requestBody VehicleWithImages
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(apperrors.ErrFailedMarshal.Error(), "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		vehicleData, err := vehicleService.CreateVehicle(ctx, requestBody)
		if err != nil {
			slog.Error("failed to create new vehicle", "error", err.Error())
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
		parseVehicleId, err := strconv.ParseInt(vehicleId, 10, 64)
		if err != nil {
			slog.Error("invalid vehicle id", "error", err.Error())
			response.WriteJson(w, http.StatusBadRequest, "invalid vehicle id", nil)
			return
		}

		var requestBody VehicleWithImages
		err = json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(apperrors.ErrFailedMarshal.Error(), "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		vehicleData, err := vehicleService.UpdateVehicle(ctx, requestBody, int(parseVehicleId))
		if err != nil {
			slog.Error("failed to update vehicle", "error", err.Error())
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
		parseVehicleId, err := strconv.ParseInt(vehicleId, 10, 64)
		if err != nil {
			slog.Error("invalid vehicle id", "error", err.Error())
			response.WriteJson(w, http.StatusBadRequest, "invalid vehicle id", nil)
			return
		}

		err = vehicleService.SoftDeleteVehicle(ctx, int(parseVehicleId))
		if err != nil {
			slog.Error("failed to soft delete vehicle", "error", err.Error())
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "vehicle deleted successfully", nil)
	}
}
