package user_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/user"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/user/mocks"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/response"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UserHandlerTestSuite struct {
	suite.Suite
	service *mocks.Service
}

func Test_UserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}

func (s *UserHandlerTestSuite) SetupTest() {
	s.service = &mocks.Service{}
}

func (s *UserHandlerTestSuite) TearDownTest() {
	s.service.AssertExpectations(s.T())
}

func (s *UserHandlerTestSuite) Test_SignUpUser() {
	type testCaseStruct struct {
		name               string
		body               interface{}
		setup              func()
		expectedStatusCode int
		expectedResult     response.Response
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			body: user.CreateUserRequestBody{
				Name:        "Aditya",
				Email:       "adityarpadekar@gmail.com",
				PhoneNumber: "+919022722650",
				Password:    "aditya",
				Role:        user.Host,
			},
			setup: func() {
				s.service.On("RegisterUser", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "user registered successfully. please check your email to verify your account.",
			},
		},
		{
			name:               "Failed marshal",
			body:               []byte(`jdwodw`),
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidRequestBody.Error(),
			},
		},
		{
			name: "Invalid request body",
			body: user.CreateUserRequestBody{
				Name:        "Aditya",
				Email:       "adityarpadek",
				PhoneNumber: "+91902272dwaud2650",
				Password:    "aditya",
				Role:        "dwadwd",
			},
			setup: func() {
				s.service.On("RegisterUser", mock.Anything, mock.Anything).Return(apperrors.ErrInvalidRequestBody)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidRequestBody.Error(),
			},
		},
		{
			name: "Email already registered",
			body: user.CreateUserRequestBody{
				Name:        "Aditya",
				Email:       "adityarpadekar@gmail.com",
				PhoneNumber: "+919022722650",
				Password:    "aditya",
				Role:        user.Host,
			},
			setup: func() {
				s.service.On("RegisterUser", mock.Anything, mock.Anything).Return(apperrors.ErrEmailAlreadyRegistered)
			},
			expectedStatusCode: http.StatusConflict,
			expectedResult: response.Response{
				Message: apperrors.ErrEmailAlreadyRegistered.Error(),
			},
		},
		{
			name: "Internal server error",
			body: user.CreateUserRequestBody{
				Name:        "Aditya",
				Email:       "adityarpadekar@gmail.com",
				PhoneNumber: "+919022722650",
				Password:    "aditya",
				Role:        user.Host,
			},
			setup: func() {
				s.service.On("RegisterUser", mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResult: response.Response{
				Message: apperrors.ErrInternalServer.Error(),
			},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(s.T(), err)

			request := httptest.NewRequest("POST", "/api/v1/auth/signup", bytes.NewBuffer(bodyBytes))
			recorder := httptest.NewRecorder()

			user.SignUpUser(s.service)(recorder, request)

			result := recorder.Result()
			defer result.Body.Close()

			var responseBody response.Response
			err = json.NewDecoder(result.Body).Decode(&responseBody)
			require.NoError(s.T(), err)

			s.Equal(tt.expectedStatusCode, result.StatusCode)
			s.Equal(tt.expectedResult, responseBody)
		})
		s.TearDownTest()
	}
}

func (s *UserHandlerTestSuite) Test_SignInUser() {
	type testCaseStruct struct {
		name               string
		body               interface{}
		setup              func()
		expectedStatusCode int
		expectedResult     response.Response
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			body: user.LoginUserRequestBody{
				Email:    "adityarpadekar@gmail.com",
				Password: "aditya",
			},
			setup: func() {
				s.service.On("LoginUser", mock.Anything, mock.Anything).Return(user.AccessToken{
					AccessToken: "9320193103910391",
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "login successful",
				Data: map[string]interface{}{
					"accessToken": "9320193103910391",
				},
			},
		},
		{
			name:               "Failed marshal",
			body:               []byte(`jdwodw`),
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidRequestBody.Error(),
			},
		},
		{
			name: "Invalid request body",
			body: user.LoginUserRequestBody{
				Email:    "adityarpadek",
				Password: "aditya",
			},
			setup: func() {
				s.service.On("LoginUser", mock.Anything, mock.Anything).Return(user.AccessToken{}, apperrors.ErrInvalidRequestBody)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidRequestBody.Error(),
			},
		},
		{
			name: "Invalid credentials",
			body: user.LoginUserRequestBody{
				Email:    "adityarpadekar@gmail.com",
				Password: "aditya1",
			},
			setup: func() {
				s.service.On("LoginUser", mock.Anything, mock.Anything).Return(user.AccessToken{}, apperrors.ErrInvalidLoginCredentials)
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidLoginCredentials.Error(),
			},
		},
		{
			name: "Internal server error",
			body: user.LoginUserRequestBody{
				Email:    "adityarpadekar@gmail.com",
				Password: "aditya",
			},
			setup: func() {
				s.service.On("LoginUser", mock.Anything, mock.Anything).Return(user.AccessToken{}, apperrors.ErrInternalServer)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResult: response.Response{
				Message: apperrors.ErrInternalServer.Error(),
			},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(s.T(), err)

			request := httptest.NewRequest("POST", "/api/v1/auth/signin", bytes.NewBuffer(bodyBytes))
			recorder := httptest.NewRecorder()

			user.SignInUser(s.service)(recorder, request)

			result := recorder.Result()
			defer result.Body.Close()

			var responseBody response.Response
			err = json.NewDecoder(result.Body).Decode(&responseBody)
			require.NoError(s.T(), err)

			s.Equal(tt.expectedStatusCode, result.StatusCode)
			s.Equal(tt.expectedResult, responseBody)
		})
		s.TearDownTest()
	}
}

func (s *UserHandlerTestSuite) Test_VerifyEmail() {
	type testCaseStruct struct {
		name               string
		body               interface{}
		setup              func()
		expectedStatusCode int
		expectedResult     response.Response
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			body: user.Token{
				Token: "token",
			},
			setup: func() {
				s.service.On("VerifyEmail", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "email verification successful",
			},
		},
		{
			name:               "Failed marshal",
			body:               []byte(`jdwodw`),
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidRequestBody.Error(),
			},
		},
		{
			name: "Invalid request body",
			body: user.LoginUserRequestBody{
				Email:    "adityarpadek",
				Password: "aditya",
			},
			setup: func() {
				s.service.On("VerifyEmail", mock.Anything, mock.Anything).Return(apperrors.ErrInvalidRequestBody)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidRequestBody.Error(),
			},
		},
		{
			name: "Invalid token",
			body: user.Token{
				Token: "token",
			},
			setup: func() {
				s.service.On("VerifyEmail", mock.Anything, mock.Anything).Return(apperrors.ErrInvalidToken)
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidToken.Error(),
			},
		},
		{
			name: "Internal server error",
			body: user.Token{
				Token: "token",
			},
			setup: func() {
				s.service.On("VerifyEmail", mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResult: response.Response{
				Message: apperrors.ErrInternalServer.Error(),
			},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(s.T(), err)

			request := httptest.NewRequest("POST", "/api/v1/auth/email/verify", bytes.NewBuffer(bodyBytes))
			recorder := httptest.NewRecorder()

			user.VerifyEmail(s.service)(recorder, request)

			result := recorder.Result()
			defer result.Body.Close()

			var responseBody response.Response
			err = json.NewDecoder(result.Body).Decode(&responseBody)
			require.NoError(s.T(), err)

			s.Equal(tt.expectedStatusCode, result.StatusCode)
			s.Equal(tt.expectedResult, responseBody)
		})
		s.TearDownTest()
	}
}

func (s *UserHandlerTestSuite) Test_ForgotPassword() {
	type testCaseStruct struct {
		name               string
		body               interface{}
		setup              func()
		expectedStatusCode int
		expectedResult     response.Response
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			body: user.Email{
				Email: "example@gmail.com",
			},
			setup: func() {
				s.service.On("ForgotPassword", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "password reset instructions have been sent to your email",
			},
		},
		{
			name:               "Failed marshal",
			body:               []byte(`jdwodw`),
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidRequestBody.Error(),
			},
		},
		{
			name: "Invalid request body",
			body: user.LoginUserRequestBody{
				Email:    "adityarpadek",
				Password: "aditya",
			},
			setup: func() {
				s.service.On("ForgotPassword", mock.Anything, mock.Anything).Return(apperrors.ErrInvalidRequestBody)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidRequestBody.Error(),
			},
		},
		{
			name: "Internal server error",
			body: user.Email{
				Email: "example@example.com",
			},
			setup: func() {
				s.service.On("ForgotPassword", mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResult: response.Response{
				Message: apperrors.ErrInternalServer.Error(),
			},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(s.T(), err)

			request := httptest.NewRequest("POST", "/api/v1/auth/password/forgot", bytes.NewBuffer(bodyBytes))
			recorder := httptest.NewRecorder()

			user.ForgotPassword(s.service)(recorder, request)

			result := recorder.Result()
			defer result.Body.Close()

			var responseBody response.Response
			err = json.NewDecoder(result.Body).Decode(&responseBody)
			require.NoError(s.T(), err)

			s.Equal(tt.expectedStatusCode, result.StatusCode)
			s.Equal(tt.expectedResult, responseBody)
		})
		s.TearDownTest()
	}
}

func (s *UserHandlerTestSuite) Test_ResetPassword() {
	type testCaseStruct struct {
		name               string
		body               interface{}
		setup              func()
		expectedStatusCode int
		expectedResult     response.Response
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			body: user.ResetPasswordRequestBody{
				Token:    "xyz",
				Password: "aditya",
			},
			setup: func() {
				s.service.On("ResetPassword", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "password has been successfully reset",
			},
		},
		{
			name:               "Failed marshal",
			body:               []byte(`jdwodw`),
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidRequestBody.Error(),
			},
		},
		{
			name: "Invalid request body",
			body: user.LoginUserRequestBody{
				Email:    "adityarpadek",
				Password: "aditya",
			},
			setup: func() {
				s.service.On("ResetPassword", mock.Anything, mock.Anything).Return(apperrors.ErrInvalidRequestBody)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidRequestBody.Error(),
			},
		},
		{
			name: "Invalid token",
			body: user.ResetPasswordRequestBody{
				Token:    "token",
				Password: "aditya",
			},
			setup: func() {
				s.service.On("ResetPassword", mock.Anything, mock.Anything).Return(apperrors.ErrInvalidToken)
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidToken.Error(),
			},
		},
		{
			name: "Internal server error",
			body: user.ResetPasswordRequestBody{
				Token:    "xyz",
				Password: "aditya",
			},
			setup: func() {
				s.service.On("ResetPassword", mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResult: response.Response{
				Message: apperrors.ErrInternalServer.Error(),
			},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(s.T(), err)

			request := httptest.NewRequest("PATCH", "/api/v1/auth/password/reset", bytes.NewBuffer(bodyBytes))
			recorder := httptest.NewRecorder()

			user.ResetPassword(s.service)(recorder, request)

			result := recorder.Result()
			defer result.Body.Close()

			var responseBody response.Response
			err = json.NewDecoder(result.Body).Decode(&responseBody)
			require.NoError(s.T(), err)

			s.Equal(tt.expectedStatusCode, result.StatusCode)
			s.Equal(tt.expectedResult, responseBody)
		})
		s.TearDownTest()
	}
}

func (s *UserHandlerTestSuite) Test_GetLoggedInUser() {
	now := time.Now()

	type testCaseStruct struct {
		name               string
		setup              func()
		expectedStatusCode int
		expectedResult     response.Response
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			setup: func() {
				s.service.On("GetLoggedInUser", mock.Anything, mock.Anything).Return(user.User{
					Id:          1,
					Name:        "John Doe",
					Email:       "john.doe@example.com",
					PhoneNumber: "+1234567890",
					Role:        "HOST",
					IsVerified:  true,
					CreatedAt:   now,
					UpdatedAt:   now,
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "user details fetched successfully",
				Data: map[string]interface{}{
					"id":          float64(1),
					"name":        "John Doe",
					"email":       "john.doe@example.com",
					"phoneNumber": "+1234567890",
					"role":        "HOST",
					"isVerified":  true,
					"createdAt":   now.Format(time.RFC3339Nano),
					"updatedAt":   now.Format(time.RFC3339Nano),
				},
			},
		},
		{
			name: "User not found",
			setup: func() {
				s.service.On("GetLoggedInUser", mock.Anything, mock.Anything).Return(user.User{}, apperrors.ErrUserNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResult: response.Response{
				Message: apperrors.ErrUserNotFound.Error(),
			},
		},
		{
			name: "Internal server error",
			setup: func() {
				s.service.On("GetLoggedInUser", mock.Anything, mock.Anything).Return(user.User{}, apperrors.ErrInternalServer)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResult: response.Response{
				Message: apperrors.ErrInternalServer.Error(),
			},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}

			request := httptest.NewRequest("GET", "/api/v1/auth/user", nil)
			recorder := httptest.NewRecorder()

			user.GetLoggedInUser(s.service)(recorder, request)

			result := recorder.Result()
			defer result.Body.Close()

			var responseBody response.Response
			err := json.NewDecoder(result.Body).Decode(&responseBody)
			require.NoError(s.T(), err)

			s.Equal(tt.expectedStatusCode, result.StatusCode)
			s.Equal(tt.expectedResult, responseBody)
		})
		s.TearDownTest()
	}
}

func (s *UserHandlerTestSuite) Test_UpgradeUserRoleToHost() {
	type testCaseStruct struct {
		name               string
		setup              func()
		expectedStatusCode int
		expectedResult     response.Response
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			setup: func() {
				s.service.On("UpgradeUserRoleToHost", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "user successfully upgraded to HOST",
			},
		},
		{
			name: "Internal server error",
			setup: func() {
				s.service.On("UpgradeUserRoleToHost", mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResult: response.Response{
				Message: apperrors.ErrInternalServer.Error(),
			},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}

			request := httptest.NewRequest("GET", "/api/v1/host/upgrade", nil)
			recorder := httptest.NewRecorder()

			user.UpgradeUserRoleToHost(s.service)(recorder, request)

			result := recorder.Result()
			defer result.Body.Close()

			var responseBody response.Response
			err := json.NewDecoder(result.Body).Decode(&responseBody)
			require.NoError(s.T(), err)

			s.Equal(tt.expectedStatusCode, result.StatusCode)
			s.Equal(tt.expectedResult, responseBody)
		})
		s.TearDownTest()
	}
}

func (s *UserHandlerTestSuite) Test_RefreshAccessToken() {
	type testCaseStruct struct {
		name               string
		setup              func()
		expectedStatusCode int
		expectedResult     response.Response
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			setup: func() {
				s.service.On("RefreshAccessToken", mock.Anything).Return(user.AccessToken{
					AccessToken: "9320193103910391",
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "access token refreshed successfully",
				Data: map[string]interface{}{
					"accessToken": "9320193103910391",
				},
			},
		},
		{
			name: "Internal server error",
			setup: func() {
				s.service.On("RefreshAccessToken", mock.Anything).Return(user.AccessToken{}, apperrors.ErrInternalServer)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResult: response.Response{
				Message: apperrors.ErrInternalServer.Error(),
			},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}

			request := httptest.NewRequest("GET", "/api/v1/host/upgrade", nil)
			recorder := httptest.NewRecorder()

			user.RefreshAccessToken(s.service)(recorder, request)

			result := recorder.Result()
			defer result.Body.Close()

			var responseBody response.Response
			err := json.NewDecoder(result.Body).Decode(&responseBody)
			require.NoError(s.T(), err)

			s.Equal(tt.expectedStatusCode, result.StatusCode)
			s.Equal(tt.expectedResult, responseBody)
		})
		s.TearDownTest()
	}
}
