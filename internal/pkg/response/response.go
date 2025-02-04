package response

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
)

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func WriteJson(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	response := Response{
		Message: message,
		Data:    data,
	}

	marshaledResponse, err := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		slog.Error("failed to marshal response", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(fmt.Sprintf(`{ "message" : "%s" }`, apperrors.ErrInternalServer.Error())))
		if err != nil {
			slog.Error("error occurred while writing response", "error", err.Error())
		}
		return
	}

	w.WriteHeader(statusCode)
	w.Write(marshaledResponse)
}
