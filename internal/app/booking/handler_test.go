package booking_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/booking"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/booking/mocks"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/response"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BookingHandlerTestSuite struct {
	suite.Suite
	service *mocks.Service
}

func Test_BookingHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(BookingHandlerTestSuite))
}

func (s *BookingHandlerTestSuite) SetupTest() {
	s.service = &mocks.Service{}
}

func (s *BookingHandlerTestSuite) TearDownTest() {
	s.service.AssertExpectations(s.T())
}

func (s *BookingHandlerTestSuite) Test_CreateBooking() {
	now := time.Now()

	requestBody := booking.CreateBookingRequestBody{
		VehicleId:             101,
		HostId:                15,
		SeekerId:              25,
		Status:                "CONFIRMED",
		PickupLocation:        "Pune Airport",
		DropoffLocation:       "Mumbai Central",
		BookingAmount:         2500.50,
		OverdueFeeRatePerHour: 500.00,
		CancellationAllowed:   true,
		ScheduledPickupTime:   now,
		ScheduledDropoffTime:  now,
	}

	dummyBooking := booking.Booking{
		Id:                    1,
		VehicleId:             101,
		HostId:                15,
		SeekerId:              25,
		Status:                "CONFIRMED",
		PickupLocation:        "Pune Airport",
		DropoffLocation:       "Mumbai Central",
		BookingAmount:         2500.50,
		OverdueFeeRatePerHour: 500.00,
		CancellationAllowed:   true,
		ScheduledPickupTime:   now,
		ScheduledDropoffTime:  now,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	dummyBookingMap := map[string]interface{}{
		"id":                    float64(1),
		"vehicleId":             float64(101),
		"hostId":                float64(15),
		"seekerId":              float64(25),
		"status":                "CONFIRMED",
		"pickupLocation":        "Pune Airport",
		"dropoffLocation":       "Mumbai Central",
		"bookingAmount":         float64(2500.50),
		"overdueFeeRatePerHour": float64(500),
		"cancellationAllowed":   true,
		"scheduledPickupTime":   now.Format(time.RFC3339Nano),
		"scheduledDropoffTime":  now.Format(time.RFC3339Nano),
		"createdAt":             now.Format(time.RFC3339Nano),
		"updatedAt":             now.Format(time.RFC3339Nano),
	}

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
			body: requestBody,
			setup: func() {
				s.service.On("CreateBooking", mock.Anything, mock.Anything).Return(dummyBooking, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "booking added successfully",
				Data:    dummyBookingMap,
			},
		},
		{
			name:               "Invalid request body",
			body:               []byte("dwwadw"),
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidRequestBody.Error(),
			},
		},
		{
			name: "Failed to create booking",
			body: requestBody,
			setup: func() {
				s.service.On("CreateBooking", mock.Anything, mock.Anything).Return(booking.Booking{}, apperrors.ErrInternalServer)
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

			booking.CreateBooking(s.service)(recorder, request)

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

func (s *BookingHandlerTestSuite) Test_CancelBooking() {
	type testCaseStruct struct {
		name               string
		params             string
		setup              func()
		expectedStatusCode int
		expectedResult     response.Response
	}

	testCases := []testCaseStruct{
		{
			name:   "Success",
			params: "1",
			setup: func() {
				s.service.On("CancelBooking", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "booking cancelled successfully",
			},
		},
		{
			name:               "Invalid booking id",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: "invalid booking id",
			},
		},
		{
			name:   "Failed to cancel booking",
			params: "1",
			setup: func() {
				s.service.On("CancelBooking", mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
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

			request := httptest.NewRequest("POST", "/api/v1/auth/signup", nil)
			request.SetPathValue("id", tt.params)
			recorder := httptest.NewRecorder()

			booking.CancelBooking(s.service)(recorder, request)

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

func (s *BookingHandlerTestSuite) Test_ConfirmPickup() {
	type testCaseStruct struct {
		name               string
		params             string
		body               interface{}
		setup              func()
		expectedStatusCode int
		expectedResult     response.Response
	}

	testCases := []testCaseStruct{
		{
			name:   "Success",
			params: "1",
			body: booking.OtpRequestBody{
				Otp: "903232",
			},
			setup: func() {
				s.service.On("ConfirmPickup", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "booking pickup confirmed successfully",
			},
		},
		{
			name:               "Invalid booking id",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: "invalid booking id",
			},
		},
		{
			name:               "Invalid request body",
			params:             "1",
			body:               []byte("dwadwa"),
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidRequestBody.Error(),
			},
		},
		{
			name:   "Failed to conform pickup",
			params: "1",
			setup: func() {
				s.service.On("ConfirmPickup", mock.Anything, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
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
			request.SetPathValue("id", tt.params)
			recorder := httptest.NewRecorder()

			booking.ConfirmPickup(s.service)(recorder, request)

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

func (s *BookingHandlerTestSuite) Test_InitiateReturn() {
	type testCaseStruct struct {
		name               string
		params             string
		setup              func()
		expectedStatusCode int
		expectedResult     response.Response
	}

	testCases := []testCaseStruct{
		{
			name:   "Success",
			params: "1",
			setup: func() {
				s.service.On("InitiateReturn", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "booking return initiated successfully",
			},
		},
		{
			name:               "Invalid booking id",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: "invalid booking id",
			},
		},
		{
			name:   "Failed to initiate booking",
			params: "1",
			setup: func() {
				s.service.On("InitiateReturn", mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
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

			request := httptest.NewRequest("POST", "/api/v1/auth/signup", nil)
			request.SetPathValue("id", tt.params)
			recorder := httptest.NewRecorder()

			booking.InitiateReturn(s.service)(recorder, request)

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

func (s *BookingHandlerTestSuite) Test_ConfirmReturn() {
	type testCaseStruct struct {
		name               string
		params             string
		body               interface{}
		setup              func()
		expectedStatusCode int
		expectedResult     response.Response
	}

	testCases := []testCaseStruct{
		{
			name:   "Success",
			params: "1",
			body: booking.OtpRequestBody{
				Otp: "903232",
			},
			setup: func() {
				s.service.On("ConfirmReturn", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "booking return confirmed successfully",
			},
		},
		{
			name:               "Invalid booking id",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: "invalid booking id",
			},
		},
		{
			name:               "Invalid request body",
			params:             "1",
			body:               []byte("dwadwa"),
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidRequestBody.Error(),
			},
		},
		{
			name:   "Failed to conform return",
			params: "1",
			setup: func() {
				s.service.On("ConfirmReturn", mock.Anything, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
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
			request.SetPathValue("id", tt.params)
			recorder := httptest.NewRecorder()

			booking.ConfirmReturn(s.service)(recorder, request)

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

func (s *BookingHandlerTestSuite) Test_GetSeekerBookings() {
	now := time.Now()

	dummyBooking := booking.PaginatedBookingData{
		Data: []booking.BookingData{
			{
				Id:                      1,
				Status:                  "CONFIRMED",
				PickupLocation:          "Pune Airport",
				DropoffLocation:         "Pune Station",
				BookingAmount:           1500.00,
				OverdueFeeRatePerHour:   500.00,
				CancellationAllowed:     true,
				ScheduledPickupTime:     now,
				ScheduledDropoffTime:    now,
				VehicleName:             "XYZA",
				VehicleSeatCount:        4,
				VehicleFuelType:         "Petrol",
				VehicleTransmissionType: "Automatic",
				VehicleImage:            "https://example.com/vehicle1.jpg",
			},
			{
				Id:                      2,
				Status:                  "CANCELLED",
				PickupLocation:          "Mumbai Central",
				DropoffLocation:         "Lonavala",
				BookingAmount:           1800.00,
				OverdueFeeRatePerHour:   600.00,
				CancellationAllowed:     false,
				ScheduledPickupTime:     now,
				ScheduledDropoffTime:    now,
				VehicleName:             "ABCD",
				VehicleSeatCount:        5,
				VehicleFuelType:         "Diesel",
				VehicleTransmissionType: "Manual",
				VehicleImage:            "https://example.com/vehicle2.jpg",
			},
		},
		Pagination: booking.PaginationParams{
			Page:       1,
			PageSize:   10,
			TotalCount: 2,
		},
	}

	dummyBookingsMap := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{
				"id":                      float64(1),
				"status":                  "CONFIRMED",
				"pickupLocation":          "Pune Airport",
				"dropoffLocation":         "Pune Station",
				"bookingAmount":           float64(1500),
				"overdueFeeRatePerHour":   float64(500),
				"cancellationAllowed":     true,
				"scheduledPickupTime":     now.Format(time.RFC3339Nano),
				"scheduledDropoffTime":    now.Format(time.RFC3339Nano),
				"vehicleName":             "XYZA",
				"vehicleSeatCount":        float64(4),
				"vehicleFuelType":         "Petrol",
				"vehicleTransmissionType": "Automatic",
				"vehicleImage":            "https://example.com/vehicle1.jpg",
			},
			map[string]interface{}{
				"id":                      float64(2),
				"status":                  "CANCELLED",
				"pickupLocation":          "Mumbai Central",
				"dropoffLocation":         "Lonavala",
				"bookingAmount":           float64(1800),
				"overdueFeeRatePerHour":   float64(600),
				"cancellationAllowed":     false,
				"scheduledPickupTime":     now.Format(time.RFC3339Nano),
				"scheduledDropoffTime":    now.Format(time.RFC3339Nano),
				"vehicleName":             "ABCD",
				"vehicleSeatCount":        float64(5),
				"vehicleFuelType":         "Diesel",
				"vehicleTransmissionType": "Manual",
				"vehicleImage":            "https://example.com/vehicle2.jpg",
			},
		},
		"pagination": map[string]interface{}{
			"page":       float64(1),
			"pageSize":   float64(10),
			"totalCount": float64(2),
		},
	}

	type testCaseStruct struct {
		name               string
		page               string
		limit              string
		setup              func()
		expectedStatusCode int
		expectedResult     response.Response
	}

	testCases := []testCaseStruct{
		{
			name:  "Success 1",
			page:  "1",
			limit: "10",
			setup: func() {
				s.service.On("GetSeekerBookings", mock.Anything, mock.Anything, mock.Anything).Return(dummyBooking, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "bookings fetched successfully",
				Data:    dummyBookingsMap,
			},
		},
		{
			name: "Success 2",
			page: "1",
			setup: func() {
				s.service.On("GetSeekerBookings", mock.Anything, mock.Anything, mock.Anything).Return(dummyBooking, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "bookings fetched successfully",
				Data:    dummyBookingsMap,
			},
		},
		{
			name:  "Success 3",
			limit: "10",
			setup: func() {
				s.service.On("GetSeekerBookings", mock.Anything, mock.Anything, mock.Anything).Return(dummyBooking, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "bookings fetched successfully",
				Data:    dummyBookingsMap,
			},
		},
		{
			name:               "Invalid page",
			page:               "a",
			limit:              "10",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidQueryParams.Error(),
			},
		},
		{
			name:               "Invalid limit",
			page:               "1",
			limit:              "a",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidQueryParams.Error(),
			},
		},
		{
			name:  "Failed to get bookings",
			page:  "1",
			limit: "10",
			setup: func() {
				s.service.On("GetSeekerBookings", mock.Anything, mock.Anything, mock.Anything).Return(booking.PaginatedBookingData{}, apperrors.ErrInternalServer)
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

			request := httptest.NewRequest("POST", "/api/v1/auth/signup", nil)
			query := request.URL.Query()
			query.Set("page", tt.page)
			query.Set("limit", tt.limit)
			request.URL.RawQuery = query.Encode()
			recorder := httptest.NewRecorder()

			booking.GetSeekerBookings(s.service)(recorder, request)

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

func (s *BookingHandlerTestSuite) Test_GetHostBookings() {
	now := time.Now()

	dummyBooking := booking.PaginatedBookingData{
		Data: []booking.BookingData{
			{
				Id:                      1,
				Status:                  "CONFIRMED",
				PickupLocation:          "Pune Airport",
				DropoffLocation:         "Pune Station",
				BookingAmount:           1500.00,
				OverdueFeeRatePerHour:   500.00,
				CancellationAllowed:     true,
				ScheduledPickupTime:     now,
				ScheduledDropoffTime:    now,
				VehicleName:             "XYZA",
				VehicleSeatCount:        4,
				VehicleFuelType:         "Petrol",
				VehicleTransmissionType: "Automatic",
				VehicleImage:            "https://example.com/vehicle1.jpg",
			},
			{
				Id:                      2,
				Status:                  "CANCELLED",
				PickupLocation:          "Mumbai Central",
				DropoffLocation:         "Lonavala",
				BookingAmount:           1800.00,
				OverdueFeeRatePerHour:   600.00,
				CancellationAllowed:     false,
				ScheduledPickupTime:     now,
				ScheduledDropoffTime:    now,
				VehicleName:             "ABCD",
				VehicleSeatCount:        5,
				VehicleFuelType:         "Diesel",
				VehicleTransmissionType: "Manual",
				VehicleImage:            "https://example.com/vehicle2.jpg",
			},
		},
		Pagination: booking.PaginationParams{
			Page:       1,
			PageSize:   10,
			TotalCount: 2,
		},
	}

	dummyBookingsMap := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{
				"id":                      float64(1),
				"status":                  "CONFIRMED",
				"pickupLocation":          "Pune Airport",
				"dropoffLocation":         "Pune Station",
				"bookingAmount":           float64(1500),
				"overdueFeeRatePerHour":   float64(500),
				"cancellationAllowed":     true,
				"scheduledPickupTime":     now.Format(time.RFC3339Nano),
				"scheduledDropoffTime":    now.Format(time.RFC3339Nano),
				"vehicleName":             "XYZA",
				"vehicleSeatCount":        float64(4),
				"vehicleFuelType":         "Petrol",
				"vehicleTransmissionType": "Automatic",
				"vehicleImage":            "https://example.com/vehicle1.jpg",
			},
			map[string]interface{}{
				"id":                      float64(2),
				"status":                  "CANCELLED",
				"pickupLocation":          "Mumbai Central",
				"dropoffLocation":         "Lonavala",
				"bookingAmount":           float64(1800),
				"overdueFeeRatePerHour":   float64(600),
				"cancellationAllowed":     false,
				"scheduledPickupTime":     now.Format(time.RFC3339Nano),
				"scheduledDropoffTime":    now.Format(time.RFC3339Nano),
				"vehicleName":             "ABCD",
				"vehicleSeatCount":        float64(5),
				"vehicleFuelType":         "Diesel",
				"vehicleTransmissionType": "Manual",
				"vehicleImage":            "https://example.com/vehicle2.jpg",
			},
		},
		"pagination": map[string]interface{}{
			"page":       float64(1),
			"pageSize":   float64(10),
			"totalCount": float64(2),
		},
	}

	type testCaseStruct struct {
		name               string
		page               string
		limit              string
		setup              func()
		expectedStatusCode int
		expectedResult     response.Response
	}

	testCases := []testCaseStruct{
		{
			name:  "Success 1",
			page:  "1",
			limit: "10",
			setup: func() {
				s.service.On("GetHostBookings", mock.Anything, mock.Anything, mock.Anything).Return(dummyBooking, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "bookings fetched successfully",
				Data:    dummyBookingsMap,
			},
		},
		{
			name: "Success 2",
			page: "1",
			setup: func() {
				s.service.On("GetHostBookings", mock.Anything, mock.Anything, mock.Anything).Return(dummyBooking, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "bookings fetched successfully",
				Data:    dummyBookingsMap,
			},
		},
		{
			name:  "Success 3",
			limit: "10",
			setup: func() {
				s.service.On("GetHostBookings", mock.Anything, mock.Anything, mock.Anything).Return(dummyBooking, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "bookings fetched successfully",
				Data:    dummyBookingsMap,
			},
		},
		{
			name:               "Invalid page",
			page:               "a",
			limit:              "10",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidQueryParams.Error(),
			},
		},
		{
			name:               "Invalid limit",
			page:               "1",
			limit:              "a",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidQueryParams.Error(),
			},
		},
		{
			name:  "Failed to get bookings",
			page:  "1",
			limit: "10",
			setup: func() {
				s.service.On("GetHostBookings", mock.Anything, mock.Anything, mock.Anything).Return(booking.PaginatedBookingData{}, apperrors.ErrInternalServer)
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

			request := httptest.NewRequest("POST", "/api/v1/auth/signup", nil)
			query := request.URL.Query()
			query.Set("page", tt.page)
			query.Set("limit", tt.limit)
			request.URL.RawQuery = query.Encode()
			recorder := httptest.NewRecorder()

			booking.GetHostBookings(s.service)(recorder, request)

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

func (s *BookingHandlerTestSuite) Test_GetBookingDetailsById() {
	now := time.Now()

	dummyBooking := booking.BookingDetails{
		Id:                    1,
		Status:                "COMPLETED",
		PickupLocation:        "Pune Airport",
		DropoffLocation:       "Pune Station",
		BookingAmount:         2500.00,
		OverdueFeeRatePerHour: 500.00,
		CancellationAllowed:   true,
		ActualPickupTime:      &now,
		ActualDropoffTime:     &now,
		ScheduledPickupTime:   now,
		ScheduledDropoffTime:  now,
		Host: booking.BookingDetailsUser{
			Id:          10,
			Name:        "John Doe",
			Email:       "john.doe@example.com",
			PhoneNumber: "+1234567890",
		},
		Seeker: booking.BookingDetailsUser{
			Id:          20,
			Name:        "Alice Smith",
			Email:       "alice.smith@example.com",
			PhoneNumber: "+9876543210",
		},
		Vehicle: booking.BookingDetailsVehicle{
			Id:               5,
			Name:             "Toyota Corolla",
			FuelType:         "Petrol",
			SeatCount:        4,
			TransmissionType: "Automatic",
			Image:            "https://example.com/vehicle1.jpg",
		},
		Invoice: booking.BookingDetailsInvoice{
			Id:             1001,
			AdditionalFees: 200.00,
			Tax:            180.00,
			TaxRate:        9.0,
			TotalAmount:    2880.00,
		},
	}

	dummyBookingMap := map[string]interface{}{
		"id":                    float64(1),
		"status":                "COMPLETED",
		"pickupLocation":        "Pune Airport",
		"dropoffLocation":       "Pune Station",
		"bookingAmount":         float64(2500),
		"overdueFeeRatePerHour": float64(500),
		"cancellationAllowed":   true,
		"actualPickupTime":      now.Format(time.RFC3339Nano),
		"actualDropoffTime":     now.Format(time.RFC3339Nano),
		"scheduledPickupTime":   now.Format(time.RFC3339Nano),
		"scheduledDropoffTime":  now.Format(time.RFC3339Nano),
		"host": map[string]interface{}{
			"id":          float64(10),
			"name":        "John Doe",
			"email":       "john.doe@example.com",
			"phoneNumber": "+1234567890",
		},
		"seeker": map[string]interface{}{
			"id":          float64(20),
			"name":        "Alice Smith",
			"email":       "alice.smith@example.com",
			"phoneNumber": "+9876543210",
		},
		"vehicle": map[string]interface{}{
			"id":               float64(5),
			"name":             "Toyota Corolla",
			"fuelType":         "Petrol",
			"seatCount":        float64(4),
			"transmissionType": "Automatic",
			"image":            "https://example.com/vehicle1.jpg",
		},
		"invoice": map[string]interface{}{
			"id":             float64(1001),
			"additionalFees": float64(200),
			"tax":            float64(180),
			"taxRate":        float64(9),
			"totalAmount":    float64(2880),
		},
	}

	type testCaseStruct struct {
		name               string
		params             string
		setup              func()
		expectedStatusCode int
		expectedResult     response.Response
	}

	testCases := []testCaseStruct{
		{
			name:   "Success",
			params: "1",
			setup: func() {
				s.service.On("GetBookingDetailsById", mock.Anything, mock.Anything).Return(dummyBooking, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "booking details fetched successfully",
				Data:    dummyBookingMap,
			},
		},
		{
			name:               "Invalid booking id",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: "invalid booking id",
			},
		},
		{
			name:   "Failed to get booking details",
			params: "1",
			setup: func() {
				s.service.On("GetBookingDetailsById", mock.Anything, mock.Anything).Return(booking.BookingDetails{}, apperrors.ErrInternalServer)
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

			request := httptest.NewRequest("POST", "/api/v1/auth/signup", nil)
			request.SetPathValue("id", tt.params)
			recorder := httptest.NewRecorder()

			booking.GetBookingDetailsById(s.service)(recorder, request)

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
