package api

import (
	"net/http"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/user"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/constant"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/middleware"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/response"
)

func NewRouter(deps app.Dependencies) *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		response.WriteJson(w, http.StatusOK, "Wheelio server is up and running..", nil)
	})

	router.HandleFunc("POST /api/v1/auth/signup", user.SignUpUser(deps.UserService))
	router.HandleFunc("POST /api/v1/auth/signin", user.SignInUser(deps.UserService))
	router.HandleFunc("POST /api/v1/auth/email/verify", user.VerifyEmail(deps.UserService))
	router.HandleFunc("POST /api/v1/auth/password/forgot", user.ForgotPassword(deps.UserService))
	router.HandleFunc("POST /api/v1/auth/password/reset", user.ResetPassword(deps.UserService))
	router.HandleFunc(
		"GET /api/v1/auth/user",
		middleware.ChainMiddleware(
			user.GetLoggedInUser(deps.UserService),
			middleware.AuthenticationMiddleware,
		),
	)
	router.HandleFunc("POST /api/v1/auth/host/singup", user.HostSignUpUser(deps.UserService))
	router.HandleFunc(
		"PATCH /api/v1/auth/host/upgrade",
		middleware.ChainMiddleware(
			user.UpgradeUserRoleToHost(deps.UserService),
			middleware.AuthorizationMiddleware(constant.SEEKER),
			middleware.AuthenticationMiddleware,
		),
	)

	return router
}
