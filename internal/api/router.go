package api

import (
	"net/http"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/response"
)

func NewRouter(deps app.Dependencies) *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		response.WriteJson(w, http.StatusOK, "Wheelio server is up and running..", nil)
	})

	return router
}
