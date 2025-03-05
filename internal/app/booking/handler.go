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

func ConfirmPickup(bookingService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bookingId := r.PathValue("id")
		parsedBookingId, err := strconv.Atoi(bookingId)
		if err != nil {
			slog.Error("invalid booking id", "error", err)
			response.WriteJson(w, http.StatusBadRequest, "invalid booking id", nil)
			return
		}

		var requestBody OtpRequestBody
		err = json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(apperrors.ErrFailedMarshal.Error(), "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		err = bookingService.ConfirmPickup(ctx, parsedBookingId, requestBody)
		if err != nil {
			slog.Error("failed to confirm pickup", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "booking pickup confirmed successfully", nil)
	}
}

func InitiateReturn(bookingService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bookingId := r.PathValue("id")
		parsedBookingId, err := strconv.Atoi(bookingId)
		if err != nil {
			slog.Error("invalid booking id", "error", err)
			response.WriteJson(w, http.StatusBadRequest, "invalid booking id", nil)
			return
		}

		err = bookingService.InitiateReturn(ctx, parsedBookingId)
		if err != nil {
			slog.Error("failed to initiate booking", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "booking return initiated successfully", nil)
	}
}

func ConfirmReturn(bookingService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bookingId := r.PathValue("id")
		parsedBookingId, err := strconv.Atoi(bookingId)
		if err != nil {
			slog.Error("invalid booking id", "error", err)
			response.WriteJson(w, http.StatusBadRequest, "invalid booking id", nil)
			return
		}

		var requestBody OtpRequestBody
		err = json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(apperrors.ErrFailedMarshal.Error(), "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		err = bookingService.ConfirmReturn(ctx, parsedBookingId, requestBody)
		if err != nil {
			slog.Error("failed to confirm booking return", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "booking return confirmed successfully", nil)
	}
}

func GetSeekerBookings(bookingService Service) http.HandlerFunc {
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

		bookings, err := bookingService.GetSeekerBookings(ctx, int(page), int(limit))
		if err != nil {
			slog.Error("failed to fetch bookings", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "bookings fetched successfully", bookings)
	}
}

func GetHostBookings(bookingService Service) http.HandlerFunc {
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

		bookings, err := bookingService.GetHostBookings(ctx, int(page), int(limit))
		if err != nil {
			slog.Error("failed to fetch bookings", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "bookings fetched successfully", bookings)
	}
}

func GetBookingDetailsById(bookingService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bookingId := r.PathValue("id")
		parsedBookingId, err := strconv.ParseInt(bookingId, 10, 64)
		if err != nil {
			slog.Error("invalid booking id", "error", err)
			response.WriteJson(w, http.StatusBadRequest, "invalid booking id", nil)
			return
		}

		bookingDetails, err := bookingService.GetBookingDetailsById(ctx, int(parsedBookingId))
		if err != nil {
			slog.Error("failed to fetch booking details", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "booking details fetched successfully", bookingDetails)
	}
}
