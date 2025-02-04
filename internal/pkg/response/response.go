package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func WriteJson(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	response := Response{
		Message: message,
		Result:  data,
	}

	marshaledResponse, err := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		slog.Error("Failed to marshal response", slog.String("Error:", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message" : "Failed to marshal response" }`))
		return
	}

	w.WriteHeader(statusCode)
	w.Write(marshaledResponse)
}
