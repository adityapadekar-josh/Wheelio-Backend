package vehicle_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/vehicle"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/vehicle/mocks"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/response"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type VehicleHandlerTestSuite struct {
	suite.Suite
	service *mocks.Service
}

func Test_VehicleHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(VehicleHandlerTestSuite))
}

func (s *VehicleHandlerTestSuite) SetupTest() {
	s.service = &mocks.Service{}
}

func (s *VehicleHandlerTestSuite) TearDownTest() {
	s.service.AssertExpectations(s.T())
}

func (s *VehicleHandlerTestSuite) Test_CreateVehicle() {
	now := time.Now()

	requestBody := vehicle.VehicleRequestBody{
		Name:                  "Toyota Prius",
		FuelType:              "Hybrid",
		SeatCount:             5,
		TransmissionType:      "Automatic",
		Features:              json.RawMessage(`["Air Conditioning", "Bluetooth", "GPS", "Heated Seats"]`),
		RatePerHour:           25.5,
		OverdueFeeRatePerHour: 10.0,
		Address:               "123 Main St",
		State:                 "Maharashtra",
		City:                  "Pune",
		PinCode:               411001,
		CancellationAllowed:   true,
		Images: []vehicle.VehicleImage{
			{Url: "https://example.com/image1.jpg", Featured: true},
			{Url: "https://example.com/image2.jpg", Featured: false},
		},
	}

	dummyVehicle := vehicle.Vehicle{
		Id:                    1,
		Name:                  "Toyota Prius",
		FuelType:              "Hybrid",
		SeatCount:             5,
		TransmissionType:      "Automatic",
		Features:              json.RawMessage(`["Air Conditioning", "Bluetooth", "GPS", "Heated Seats"]`),
		RatePerHour:           25.5,
		OverdueFeeRatePerHour: 10.0,
		Address:               "123 Main St",
		State:                 "Maharashtra",
		City:                  "Pune",
		PinCode:               411001,
		CancellationAllowed:   true,
		Images: []vehicle.VehicleImage{
			{
				Id:        1,
				VehicleId: 1,
				Url:       "https://example.com/image1.jpg",
				Featured:  true,
				CreatedAt: now,
			},
			{
				Id:        2,
				VehicleId: 1,
				Url:       "https://example.com/image2.jpg",
				Featured:  false,
				CreatedAt: now,
			},
		},
		Available: true,
		HostId:    101,
		IsDeleted: false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	dummyVehicleMap := map[string]interface{}{
		"id":                    float64(1),
		"name":                  "Toyota Prius",
		"fuelType":              "Hybrid",
		"seatCount":             float64(5),
		"transmissionType":      "Automatic",
		"features":              []interface{}{"Air Conditioning", "Bluetooth", "GPS", "Heated Seats"},
		"ratePerHour":           float64(25.5),
		"overdueFeeRatePerHour": float64(10),
		"address":               "123 Main St",
		"state":                 "Maharashtra",
		"city":                  "Pune",
		"pinCode":               float64(411001),
		"cancellationAllowed":   true,
		"images": []interface{}{
			map[string]interface{}{
				"id":        float64(1),
				"url":       "https://example.com/image1.jpg",
				"featured":  true,
				"createdAt": now.Format(time.RFC3339Nano),
			},
			map[string]interface{}{
				"id":        float64(2),
				"url":       "https://example.com/image2.jpg",
				"featured":  false,
				"createdAt": now.Format(time.RFC3339Nano),
			},
		},
		"available": true,
		"hostId":    float64(101),
		"isDeleted": false,
		"createdAt": now.Format(time.RFC3339Nano),
		"updatedAt": now.Format(time.RFC3339Nano),
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
				s.service.On("CreateVehicle", mock.Anything, mock.Anything).Return(dummyVehicle, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "vehicle added successfully",
				Data:    dummyVehicleMap,
			},
		},
		{
			name: "Empty request body",
			body: nil,
			setup: func() {
				s.service.On("CreateVehicle", mock.Anything, mock.Anything).Return(vehicle.Vehicle{}, apperrors.ErrInvalidRequestBody)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidRequestBody.Error(),
			},
		},
		{
			name:               "Invalid request body",
			body:               []byte("dhwudhwudw"),
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidRequestBody.Error(),
			},
		},
		{
			name: "Internal server error",
			body: requestBody,
			setup: func() {
				s.service.On("CreateVehicle", mock.Anything, mock.Anything).Return(vehicle.Vehicle{}, apperrors.ErrInternalServer)
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

			request := httptest.NewRequest("POST", "/api/v1/vehicles", bytes.NewBuffer(bodyBytes))
			recorder := httptest.NewRecorder()

			vehicle.CreateVehicle(s.service)(recorder, request)

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

func (s *VehicleHandlerTestSuite) Test_UpdateVehicle() {
	now := time.Now()

	requestBody := vehicle.VehicleRequestBody{
		Name:                  "Toyota Prius",
		FuelType:              "Hybrid",
		SeatCount:             5,
		TransmissionType:      "Automatic",
		Features:              json.RawMessage(`["Air Conditioning", "Bluetooth", "GPS", "Heated Seats"]`),
		RatePerHour:           25.5,
		OverdueFeeRatePerHour: 10.0,
		Address:               "123 Main St",
		State:                 "Maharashtra",
		City:                  "Pune",
		PinCode:               411001,
		CancellationAllowed:   true,
		Images: []vehicle.VehicleImage{
			{Url: "https://example.com/image1.jpg", Featured: true},
			{Url: "https://example.com/image2.jpg", Featured: false},
		},
	}

	dummyVehicle := vehicle.Vehicle{
		Id:                    1,
		Name:                  "Toyota Prius",
		FuelType:              "Hybrid",
		SeatCount:             5,
		TransmissionType:      "Automatic",
		Features:              json.RawMessage(`["Air Conditioning", "Bluetooth", "GPS", "Heated Seats"]`),
		RatePerHour:           25.5,
		OverdueFeeRatePerHour: 10.0,
		Address:               "123 Main St",
		State:                 "Maharashtra",
		City:                  "Pune",
		PinCode:               411001,
		CancellationAllowed:   true,
		Images: []vehicle.VehicleImage{
			{
				Id:        1,
				VehicleId: 1,
				Url:       "https://example.com/image1.jpg",
				Featured:  true,
				CreatedAt: now,
			},
			{
				Id:        2,
				VehicleId: 1,
				Url:       "https://example.com/image2.jpg",
				Featured:  false,
				CreatedAt: now,
			},
		},
		Available: true,
		HostId:    101,
		IsDeleted: false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	dummyVehicleMap := map[string]interface{}{
		"id":                    float64(1),
		"name":                  "Toyota Prius",
		"fuelType":              "Hybrid",
		"seatCount":             float64(5),
		"transmissionType":      "Automatic",
		"features":              []interface{}{"Air Conditioning", "Bluetooth", "GPS", "Heated Seats"},
		"ratePerHour":           float64(25.5),
		"overdueFeeRatePerHour": float64(10),
		"address":               "123 Main St",
		"state":                 "Maharashtra",
		"city":                  "Pune",
		"pinCode":               float64(411001),
		"cancellationAllowed":   true,
		"images": []interface{}{
			map[string]interface{}{
				"id":        float64(1),
				"url":       "https://example.com/image1.jpg",
				"featured":  true,
				"createdAt": now.Format(time.RFC3339Nano),
			},
			map[string]interface{}{
				"id":        float64(2),
				"url":       "https://example.com/image2.jpg",
				"featured":  false,
				"createdAt": now.Format(time.RFC3339Nano),
			},
		},
		"available": true,
		"hostId":    float64(101),
		"isDeleted": false,
		"createdAt": now.Format(time.RFC3339Nano),
		"updatedAt": now.Format(time.RFC3339Nano),
	}

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
			body:   requestBody,
			setup: func() {
				s.service.On("UpdateVehicle", mock.Anything, mock.Anything, mock.Anything).Return(dummyVehicle, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "vehicle updated successfully",
				Data:    dummyVehicleMap,
			},
		},
		{
			name:   "Empty request body",
			params: "1",
			body:   nil,
			setup: func() {
				s.service.On("UpdateVehicle", mock.Anything, mock.Anything, mock.Anything).Return(vehicle.Vehicle{}, apperrors.ErrInvalidRequestBody)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidRequestBody.Error(),
			},
		},
		{
			name:               "Invalid request body",
			params:             "1",
			body:               []byte("dhwudhwudw"),
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidRequestBody.Error(),
			},
		},
		{
			name:               "Invalid vehicle param id",
			body:               requestBody,
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: "invalid vehicle id",
			},
		},
		{
			name:   "Internal server error",
			params: "1",
			body:   requestBody,
			setup: func() {
				s.service.On("UpdateVehicle", mock.Anything, mock.Anything, mock.Anything).Return(vehicle.Vehicle{}, apperrors.ErrInternalServer)
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

			request := httptest.NewRequest("POST", "/api/v1/vehicles/1", bytes.NewBuffer(bodyBytes))
			request.SetPathValue("id", tt.params)
			recorder := httptest.NewRecorder()

			vehicle.UpdateVehicle(s.service)(recorder, request)

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

func (s *VehicleHandlerTestSuite) Test_SoftDeleteVehicle() {
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
				s.service.On("SoftDeleteVehicle", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "vehicle deleted successfully",
			},
		},
		{
			name:               "Invalid vehicle param id",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: "invalid vehicle id",
			},
		},
		{
			name:   "Internal server error",
			params: "1",
			setup: func() {
				s.service.On("SoftDeleteVehicle", mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
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

			request := httptest.NewRequest("POST", "/api/v1/vehicles/1", nil)
			request.SetPathValue("id", tt.params)
			recorder := httptest.NewRecorder()

			vehicle.SoftDeleteVehicle(s.service)(recorder, request)

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

func (s *VehicleHandlerTestSuite) Test_GenerateSignedVehicleImageUploadURL() {
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
				s.service.On("GenerateSignedVehicleImageUploadURL", mock.Anything, mock.Anything).Return("xyz", "abc", nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "signed url generated successfully",
				Data: map[string]interface{}{
					"signedUrl": "xyz",
					"accessUrl": "abc",
				},
			},
		},
		{
			name: "Internal server error",
			setup: func() {
				s.service.On("GenerateSignedVehicleImageUploadURL", mock.Anything, mock.Anything).Return("", "", apperrors.ErrInternalServer)
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

			request := httptest.NewRequest("POST", "/api/v1/vehicles/1", nil)
			recorder := httptest.NewRecorder()

			vehicle.GenerateSignedVehicleImageUploadURL(s.service)(recorder, request)

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

func (s *VehicleHandlerTestSuite) Test_GetVehicleById() {
	now := time.Now()

	dummyVehicle := vehicle.Vehicle{
		Id:                    1,
		Name:                  "Toyota Prius",
		FuelType:              "Hybrid",
		SeatCount:             5,
		TransmissionType:      "Automatic",
		Features:              json.RawMessage(`["Air Conditioning", "Bluetooth", "GPS", "Heated Seats"]`),
		RatePerHour:           25.5,
		OverdueFeeRatePerHour: 10.0,
		Address:               "123 Main St",
		State:                 "Maharashtra",
		City:                  "Pune",
		PinCode:               411001,
		CancellationAllowed:   true,
		Images: []vehicle.VehicleImage{
			{
				Id:        1,
				VehicleId: 1,
				Url:       "https://example.com/image1.jpg",
				Featured:  true,
				CreatedAt: now,
			},
			{
				Id:        2,
				VehicleId: 1,
				Url:       "https://example.com/image2.jpg",
				Featured:  false,
				CreatedAt: now,
			},
		},
		Available: true,
		HostId:    101,
		IsDeleted: false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	dummyVehicleMap := map[string]interface{}{
		"id":                    float64(1),
		"name":                  "Toyota Prius",
		"fuelType":              "Hybrid",
		"seatCount":             float64(5),
		"transmissionType":      "Automatic",
		"features":              []interface{}{"Air Conditioning", "Bluetooth", "GPS", "Heated Seats"},
		"ratePerHour":           float64(25.5),
		"overdueFeeRatePerHour": float64(10),
		"address":               "123 Main St",
		"state":                 "Maharashtra",
		"city":                  "Pune",
		"pinCode":               float64(411001),
		"cancellationAllowed":   true,
		"images": []interface{}{
			map[string]interface{}{
				"id":        float64(1),
				"url":       "https://example.com/image1.jpg",
				"featured":  true,
				"createdAt": now.Format(time.RFC3339Nano),
			},
			map[string]interface{}{
				"id":        float64(2),
				"url":       "https://example.com/image2.jpg",
				"featured":  false,
				"createdAt": now.Format(time.RFC3339Nano),
			},
		},
		"available": true,
		"hostId":    float64(101),
		"isDeleted": false,
		"createdAt": now.Format(time.RFC3339Nano),
		"updatedAt": now.Format(time.RFC3339Nano),
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
				s.service.On("GetVehicleById", mock.Anything, mock.Anything).Return(dummyVehicle, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "vehicle fetched successfully",
				Data:    dummyVehicleMap,
			},
		},
		{
			name:               "Invalid vehicle param id",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: "invalid vehicle id",
			},
		},
		{
			name:   "Internal server error",
			params: "1",
			setup: func() {
				s.service.On("GetVehicleById", mock.Anything, mock.Anything).Return(vehicle.Vehicle{}, apperrors.ErrInternalServer)
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

			request := httptest.NewRequest("POST", "/api/v1/vehicles/1", nil)
			request.SetPathValue("id", tt.params)
			recorder := httptest.NewRecorder()

			vehicle.GetVehicleById(s.service)(recorder, request)

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

func (s *VehicleHandlerTestSuite) Test_GetVehicles() {
	paginatedVehicles := vehicle.PaginatedVehicleOverview{
		Data: []vehicle.VehicleOverview{
			{
				Id:               1,
				Name:             "Toyota Prius",
				FuelType:         "Hybrid",
				SeatCount:        5,
				TransmissionType: "Automatic",
				Image:            "https://example.com/images/toyota-prius.jpg",
				RatePerHour:      500.0,
				Address:          "123 Main St, Pune",
				PinCode:          411003,
			},
			{
				Id:               2,
				Name:             "Honda City",
				FuelType:         "Petrol",
				SeatCount:        5,
				TransmissionType: "Manual",
				Image:            "https://example.com/images/honda-city.jpg",
				RatePerHour:      400.0,
				Address:          "456 Market Rd, Pune",
				PinCode:          411003,
			},
			{
				Id:               3,
				Name:             "Hyundai Creta",
				FuelType:         "Diesel",
				SeatCount:        5,
				TransmissionType: "Automatic",
				Image:            "https://example.com/images/hyundai-creta.jpg",
				RatePerHour:      550.0,
				Address:          "789 Central Ave, Pune",
				PinCode:          411003,
			},
		},
		Pagination: vehicle.PaginationParams{
			Page:       1,
			PageSize:   10,
			TotalCount: 3,
		},
	}

	paginatedVehiclesMap := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{
				"id":               float64(1),
				"name":             "Toyota Prius",
				"fuelType":         "Hybrid",
				"seatCount":        float64(5),
				"transmissionType": "Automatic",
				"image":            "https://example.com/images/toyota-prius.jpg",
				"ratePerHour":      float64(500),
				"address":          "123 Main St, Pune",
				"pinCode":          float64(411003),
			},
			map[string]interface{}{
				"id":               float64(2),
				"name":             "Honda City",
				"fuelType":         "Petrol",
				"seatCount":        float64(5),
				"transmissionType": "Manual",
				"image":            "https://example.com/images/honda-city.jpg",
				"ratePerHour":      float64(400),
				"address":          "456 Market Rd, Pune",
				"pinCode":          float64(411003),
			},
			map[string]interface{}{
				"id":               float64(3),
				"name":             "Hyundai Creta",
				"fuelType":         "Diesel",
				"seatCount":        float64(5),
				"transmissionType": "Automatic",
				"image":            "https://example.com/images/hyundai-creta.jpg",
				"ratePerHour":      float64(550),
				"address":          "789 Central Ave, Pune",
				"pinCode":          float64(411003),
			},
		},
		"pagination": map[string]interface{}{
			"page":       float64(1),
			"pageSize":   float64(10),
			"totalCount": float64(3),
		},
	}

	type testCaseStruct struct {
		name               string
		page               string
		limit              string
		city               string
		pickup             string
		dropoff            string
		setup              func()
		expectedStatusCode int
		expectedResult     response.Response
	}

	testCases := []testCaseStruct{
		{
			name:    "Success 1",
			page:    "1",
			limit:   "10",
			city:    "Pune",
			pickup:  "2025-03-06T08:30:00Z",
			dropoff: "2025-03-06T12:45:00Z",
			setup: func() {
				s.service.On("GetVehicles", mock.Anything, mock.Anything).Return(paginatedVehicles, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "vehicles fetched successfully",
				Data:    paginatedVehiclesMap,
			},
		},
		{
			name:  "Success 2",
			page:  "1",
			limit: "10",
			city:  "Pune",
			setup: func() {
				s.service.On("GetVehicles", mock.Anything, mock.Anything).Return(paginatedVehicles, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "vehicles fetched successfully",
				Data:    paginatedVehiclesMap,
			},
		},
		{
			name:  "Success 3",
			limit: "10",
			city:  "Pune",
			setup: func() {
				s.service.On("GetVehicles", mock.Anything, mock.Anything).Return(paginatedVehicles, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "vehicles fetched successfully",
				Data:    paginatedVehiclesMap,
			},
		},
		{
			name: "Success 4",
			city: "Pune",
			setup: func() {
				s.service.On("GetVehicles", mock.Anything, mock.Anything).Return(paginatedVehicles, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "vehicles fetched successfully",
				Data:    paginatedVehiclesMap,
			},
		},
		{
			name:               "Invalid page 1",
			page:               "a",
			city:               "Pune",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidQueryParams.Error(),
			},
		},
		{
			name:               "Invalid limit 2",
			limit:              "a",
			city:               "Pune",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidQueryParams.Error(),
			},
		},
		{
			name:               "No city provided",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidQueryParams.Error(),
			},
		},
		{
			name:               "No only one timestamp 1",
			city:               "Pune",
			dropoff:            "2025-03-06T12:45:00Z",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidQueryParams.Error(),
			},
		},
		{
			name:               "No only one timestamp 2",
			city:               "Pune",
			pickup:             "2025-03-06T08:30:00Z",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidQueryParams.Error(),
			},
		},
		{
			name:               "Invalid timestamp 1",
			city:               "Pune",
			pickup:             "2025-03dwa-06T08:30:00Z",
			dropoff:            "2025-03-06T12:45:00Z",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidQueryParams.Error(),
			},
		},
		{
			name:               "Invalid timestamp 2",
			city:               "Pune",
			pickup:             "2025-03-06T08:30:00Z",
			dropoff:            "2025-03-06dwadawT12:45:00Z",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidQueryParams.Error(),
			},
		},
		{
			name:               "Invalid timestamp 3",
			city:               "Pune",
			pickup:             "2025-03-06T14:00:00+05:30",
			dropoff:            "2025-03-06T12:45:00Z",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidQueryParams.Error(),
			},
		},
		{
			name:               "Invalid timestamp 4",
			city:               "Pune",
			pickup:             "2025-03-06T08:30:00Z",
			dropoff:            "2025-03-06T18:15:00+05:30",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidQueryParams.Error(),
			},
		},
		{
			name:    "Internal server error",
			page:    "1",
			limit:   "10",
			city:    "Pune",
			pickup:  "2025-03-06T08:30:00Z",
			dropoff: "2025-03-06T12:45:00Z",
			setup: func() {
				s.service.On("GetVehicles", mock.Anything, mock.Anything).Return(vehicle.PaginatedVehicleOverview{}, apperrors.ErrInternalServer)
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

			request := httptest.NewRequest("POST", "/api/v1/vehicles/1", nil)
			query := request.URL.Query()
			query.Set("page", tt.page)
			query.Set("limit", tt.limit)
			query.Set("city", tt.city)
			query.Set("pickup", tt.pickup)
			query.Set("dropoff", tt.dropoff)
			request.URL.RawQuery = query.Encode()
			recorder := httptest.NewRecorder()

			vehicle.GetVehicles(s.service)(recorder, request)

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

func (s *VehicleHandlerTestSuite) Test_GetVehiclesForHost() {
	paginatedVehicles := vehicle.PaginatedVehicleOverview{
		Data: []vehicle.VehicleOverview{
			{
				Id:               1,
				Name:             "Toyota Prius",
				FuelType:         "Hybrid",
				SeatCount:        5,
				TransmissionType: "Automatic",
				Image:            "https://example.com/images/toyota-prius.jpg",
				RatePerHour:      500.0,
				Address:          "123 Main St, Pune",
				PinCode:          411003,
			},
			{
				Id:               2,
				Name:             "Honda City",
				FuelType:         "Petrol",
				SeatCount:        5,
				TransmissionType: "Manual",
				Image:            "https://example.com/images/honda-city.jpg",
				RatePerHour:      400.0,
				Address:          "456 Market Rd, Pune",
				PinCode:          411003,
			},
			{
				Id:               3,
				Name:             "Hyundai Creta",
				FuelType:         "Diesel",
				SeatCount:        5,
				TransmissionType: "Automatic",
				Image:            "https://example.com/images/hyundai-creta.jpg",
				RatePerHour:      550.0,
				Address:          "789 Central Ave, Pune",
				PinCode:          411003,
			},
		},
		Pagination: vehicle.PaginationParams{
			Page:       1,
			PageSize:   10,
			TotalCount: 3,
		},
	}

	paginatedVehiclesMap := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{
				"id":               float64(1),
				"name":             "Toyota Prius",
				"fuelType":         "Hybrid",
				"seatCount":        float64(5),
				"transmissionType": "Automatic",
				"image":            "https://example.com/images/toyota-prius.jpg",
				"ratePerHour":      float64(500),
				"address":          "123 Main St, Pune",
				"pinCode":          float64(411003),
			},
			map[string]interface{}{
				"id":               float64(2),
				"name":             "Honda City",
				"fuelType":         "Petrol",
				"seatCount":        float64(5),
				"transmissionType": "Manual",
				"image":            "https://example.com/images/honda-city.jpg",
				"ratePerHour":      float64(400),
				"address":          "456 Market Rd, Pune",
				"pinCode":          float64(411003),
			},
			map[string]interface{}{
				"id":               float64(3),
				"name":             "Hyundai Creta",
				"fuelType":         "Diesel",
				"seatCount":        float64(5),
				"transmissionType": "Automatic",
				"image":            "https://example.com/images/hyundai-creta.jpg",
				"ratePerHour":      float64(550),
				"address":          "789 Central Ave, Pune",
				"pinCode":          float64(411003),
			},
		},
		"pagination": map[string]interface{}{
			"page":       float64(1),
			"pageSize":   float64(10),
			"totalCount": float64(3),
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
				s.service.On("GetVehiclesForHost", mock.Anything, mock.Anything,mock.Anything).Return(paginatedVehicles, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "vehicles fetched successfully",
				Data:    paginatedVehiclesMap,
			},
		},
		{
			name: "Success 2",
			page: "1",
			setup: func() {
				s.service.On("GetVehiclesForHost", mock.Anything, mock.Anything,mock.Anything).Return(paginatedVehicles, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "vehicles fetched successfully",
				Data:    paginatedVehiclesMap,
			},
		},
		{
			name:  "Success 3",
			limit: "10",
			setup: func() {
				s.service.On("GetVehiclesForHost", mock.Anything, mock.Anything,mock.Anything).Return(paginatedVehicles, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResult: response.Response{
				Message: "vehicles fetched successfully",
				Data:    paginatedVehiclesMap,
			},
		},
		{
			name:               "Invalid page 1",
			page:               "a",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidQueryParams.Error(),
			},
		},
		{
			name:               "Invalid limit 2",
			limit:              "a",
			expectedStatusCode: http.StatusBadRequest,
			expectedResult: response.Response{
				Message: apperrors.ErrInvalidQueryParams.Error(),
			},
		},
		{
			name:  "Internal server error",
			page:  "1",
			limit: "10",
			setup: func() {
				s.service.On("GetVehiclesForHost", mock.Anything, mock.Anything,mock.Anything).Return(vehicle.PaginatedVehicleOverview{}, apperrors.ErrInternalServer)
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

			request := httptest.NewRequest("POST", "/api/v1/vehicles/1", nil)
			query := request.URL.Query()
			query.Set("page", tt.page)
			query.Set("limit", tt.limit)
			request.URL.RawQuery = query.Encode()
			recorder := httptest.NewRecorder()

			vehicle.GetVehiclesForHost(s.service)(recorder, request)

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
