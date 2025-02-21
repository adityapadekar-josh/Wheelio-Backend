package booking

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/response"
)

func CreateBooking(bookingService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var requestBody CreateBookingRequestBody
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(apperrors.ErrFailedMarshal.Error(), "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		bookingData, err := bookingService.CreateBooking(ctx, requestBody)
		if err != nil {
			slog.Error("failed to create new booking", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "booking added successfully", bookingData)
	}
}

func CancelBooking(bookingService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bookingId := r.PathValue("id")
		parsedBookingId, err := strconv.Atoi(bookingId)
		if err != nil {
			slog.Error("invalid booking id", "error", err)
			response.WriteJson(w, http.StatusBadRequest, "invalid booking id", nil)
			return
		}

		err = bookingService.CancelBooking(ctx, parsedBookingId)
		if err != nil {
			slog.Error("failed to cancel booking", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "booking cancelled successfully", nil)
	}
}
