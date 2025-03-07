package vehicle_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	firebaseMocks "github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/firebase/mocks"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/vehicle"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/middleware"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository"
	repositoryMocks "github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type VehicleServiceTestSuite struct {
	suite.Suite
	service           vehicle.Service
	firebaseService   *firebaseMocks.Service
	vehicleRepository *repositoryMocks.VehicleRepository
}

func Test_VehicleServiceTestSuite(t *testing.T) {
	suite.Run(t, new(VehicleServiceTestSuite))
}

func (s *VehicleServiceTestSuite) SetupTest() {
	s.vehicleRepository = &repositoryMocks.VehicleRepository{}
	s.firebaseService = &firebaseMocks.Service{}

	s.service = vehicle.NewService(s.vehicleRepository, s.firebaseService)
}

func (s *VehicleServiceTestSuite) TearDownTest() {
	s.vehicleRepository.AssertExpectations(s.T())
	s.firebaseService.AssertExpectations(s.T())
}

func (s *VehicleServiceTestSuite) Test_CreateVehicle() {
	now := time.Now()

	input := vehicle.VehicleRequestBody{
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

	dummyRepoVehicle := repository.Vehicle{
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
		Available:             true,
		HostId:                101,
		IsDeleted:             false,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	dummyRepoImages := []repository.VehicleImage{
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
	}

	type testCaseStruct struct {
		name           string
		input          vehicle.VehicleRequestBody
		setup          func(ctx *context.Context)
		expectedError  error
		expectedResult vehicle.Vehicle
	}

	testCases := []testCaseStruct{
		{
			name:  "Success",
			input: input,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}

				s.vehicleRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.vehicleRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)

				s.vehicleRepository.On("CreateVehicle", *ctx, tx, mock.Anything).Return((dummyRepoVehicle), nil)

				var callIndex int
				s.vehicleRepository.On("CreateVehicleImage", *ctx, tx, mock.Anything).Run(func(args mock.Arguments) {
					callIndex++
				}).
					Return(func(ctx context.Context, tx *sql.Tx, data repository.CreateVehicleImageData) (repository.VehicleImage, error) {
						return dummyRepoImages[callIndex-1], nil
					})
			},
			expectedError:  nil,
			expectedResult: dummyVehicle,
		},
		{
			name:           "Error retrieving id from context",
			input:          input,
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: vehicle.Vehicle{},
		},
		{
			name:  "Empty request body",
			input: vehicle.VehicleRequestBody{},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)
			},
			expectedError:  apperrors.ErrInvalidRequestBody,
			expectedResult: vehicle.Vehicle{},
		},
		{
			name: "Invalid request body",
			input: vehicle.VehicleRequestBody{
				Name:                  "Toyota Prius",
				FuelType:              "how dumb",
				SeatCount:             5,
				TransmissionType:      "nothing much",
				Features:              json.RawMessage(`["Air Conditioning", "Bluetooth", "GPS", "Heated Seats"]`),
				RatePerHour:           -25.5,
				OverdueFeeRatePerHour: -10,
				Address:               "123 Main St",
				State:                 "Maharashtra",
				City:                  "Pune",
				PinCode:               4110901,
				CancellationAllowed:   true,
				Images: []vehicle.VehicleImage{
					{Url: "https://example.com/image2.jpg", Featured: false},
				},
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)
			},
			expectedError:  apperrors.ErrInvalidRequestBody,
			expectedResult: vehicle.Vehicle{},
		},
		{
			name:  "Db transaction creation failed",
			input: input,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.vehicleRepository.On("BeginTx", *ctx).Return(nil, apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: vehicle.Vehicle{},
		},
		{
			name:  "Db transaction handling failed",
			input: input,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.vehicleRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.vehicleRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(apperrors.ErrInternalServer)

				s.vehicleRepository.On("CreateVehicle", *ctx, tx, mock.Anything).Return(repository.Vehicle{}, apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: vehicle.Vehicle{},
		},
		{
			name:  "Failed to create vehicle",
			input: input,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.vehicleRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.vehicleRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)

				s.vehicleRepository.On("CreateVehicle", *ctx, tx, mock.Anything).Return(repository.Vehicle{}, apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: vehicle.Vehicle{},
		},
		{
			name:  "Failed to create vehicle images 1",
			input: input,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.vehicleRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.vehicleRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)

				s.vehicleRepository.On("CreateVehicle", *ctx, tx, mock.Anything).Return(dummyRepoVehicle, nil)

				s.vehicleRepository.On("CreateVehicleImage", *ctx, tx, mock.Anything).Return(repository.VehicleImage{}, apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: vehicle.Vehicle{},
		},
		{
			name:  "Failed to create vehicle images 2",
			input: input,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.vehicleRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.vehicleRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)

				s.vehicleRepository.On("CreateVehicle", *ctx, tx, mock.Anything).Return(dummyRepoVehicle, nil)

				s.vehicleRepository.On("CreateVehicleImage", *ctx, tx, mock.Anything).Return(repository.VehicleImage{}, apperrors.ErrInvalidImageToLink)
			},
			expectedError:  apperrors.ErrInvalidRequestBody,
			expectedResult: vehicle.Vehicle{},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup(&ctx)
			}

			result, err := s.service.CreateVehicle(ctx, tt.input)

			s.Equal(tt.expectedError, err)
			s.Equal(tt.expectedResult, result)
		})
		s.TearDownTest()
	}
}

func (s *VehicleServiceTestSuite) Test_UpdateVehicle() {
	now := time.Now()

	input := vehicle.VehicleRequestBody{
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

	dummyRepoVehicle := repository.Vehicle{
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
		Available:             true,
		HostId:                101,
		IsDeleted:             false,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	dummyRepoImages := []repository.VehicleImage{
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
	}

	type testCaseStruct struct {
		name           string
		input          vehicle.VehicleRequestBody
		setup          func()
		expectedError  error
		expectedResult vehicle.Vehicle
	}

	testCases := []testCaseStruct{
		{
			name:  "Success",
			input: input,
			setup: func() {

				tx := &sql.Tx{}

				s.vehicleRepository.On("BeginTx", mock.Anything).Return(tx, nil)
				s.vehicleRepository.On("HandleTransaction", mock.Anything, tx, mock.Anything).Return(nil)

				s.vehicleRepository.On("UpdateVehicle", mock.Anything, tx, mock.Anything).Return((dummyRepoVehicle), nil)
				s.vehicleRepository.On("DeleteAllImagesForVehicle", mock.Anything, tx, mock.Anything).Return(nil)

				var callIndex int
				s.vehicleRepository.On("CreateVehicleImage", mock.Anything, tx, mock.Anything).Run(func(args mock.Arguments) {
					callIndex++
				}).
					Return(func(ctx context.Context, tx *sql.Tx, data repository.CreateVehicleImageData) (repository.VehicleImage, error) {
						return dummyRepoImages[callIndex-1], nil
					})
			},
			expectedError:  nil,
			expectedResult: dummyVehicle,
		},
		{
			name:           "Empty request body",
			input:          vehicle.VehicleRequestBody{},
			expectedError:  apperrors.ErrInvalidRequestBody,
			expectedResult: vehicle.Vehicle{},
		},
		{
			name: "Invalid request body",
			input: vehicle.VehicleRequestBody{
				Name:                  "Toyota Prius",
				FuelType:              "how dumb",
				SeatCount:             5,
				TransmissionType:      "nothing much",
				Features:              json.RawMessage(`["Air Conditioning", "Bluetooth", "GPS", "Heated Seats"]`),
				RatePerHour:           -25.5,
				OverdueFeeRatePerHour: -10,
				Address:               "123 Main St",
				State:                 "Maharashtra",
				City:                  "Pune",
				PinCode:               4110901,
				CancellationAllowed:   true,
				Images: []vehicle.VehicleImage{
					{Url: "https://example.com/image2.jpg", Featured: false},
				},
			},
			expectedError:  apperrors.ErrInvalidRequestBody,
			expectedResult: vehicle.Vehicle{},
		},
		{
			name:  "Db transaction creation failed",
			input: input,
			setup: func() {

				s.vehicleRepository.On("BeginTx", mock.Anything).Return(nil, apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: vehicle.Vehicle{},
		},
		{
			name:  "Db transaction handling failed",
			input: input,
			setup: func() {
				tx := &sql.Tx{}
				s.vehicleRepository.On("BeginTx", mock.Anything).Return(tx, nil)
				s.vehicleRepository.On("HandleTransaction", mock.Anything, tx, mock.Anything).Return(apperrors.ErrInternalServer)

				s.vehicleRepository.On("UpdateVehicle", mock.Anything, tx, mock.Anything).Return(repository.Vehicle{}, apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: vehicle.Vehicle{},
		},
		{
			name:  "Failed to update vehicle",
			input: input,
			setup: func() {
				tx := &sql.Tx{}
				s.vehicleRepository.On("BeginTx", mock.Anything).Return(tx, nil)
				s.vehicleRepository.On("HandleTransaction", mock.Anything, tx, mock.Anything).Return(nil)

				s.vehicleRepository.On("UpdateVehicle", mock.Anything, tx, mock.Anything).Return(repository.Vehicle{}, apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: vehicle.Vehicle{},
		},
		{
			name:  "Failed to delete old vehicle images",
			input: input,
			setup: func() {
				tx := &sql.Tx{}
				s.vehicleRepository.On("BeginTx", mock.Anything).Return(tx, nil)
				s.vehicleRepository.On("HandleTransaction", mock.Anything, tx, mock.Anything).Return(nil)

				s.vehicleRepository.On("UpdateVehicle", mock.Anything, tx, mock.Anything).Return(dummyRepoVehicle, nil)
				s.vehicleRepository.On("DeleteAllImagesForVehicle", mock.Anything, tx, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: vehicle.Vehicle{},
		},
		{
			name:  "Failed to create vehicle images 1",
			input: input,
			setup: func() {
				tx := &sql.Tx{}
				s.vehicleRepository.On("BeginTx", mock.Anything).Return(tx, nil)
				s.vehicleRepository.On("HandleTransaction", mock.Anything, tx, mock.Anything).Return(nil)

				s.vehicleRepository.On("UpdateVehicle", mock.Anything, tx, mock.Anything).Return(dummyRepoVehicle, nil)
				s.vehicleRepository.On("DeleteAllImagesForVehicle", mock.Anything, tx, mock.Anything).Return(nil)
				s.vehicleRepository.On("CreateVehicleImage", mock.Anything, tx, mock.Anything).Return(repository.VehicleImage{}, apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: vehicle.Vehicle{},
		},
		{
			name:  "Failed to create vehicle images 2",
			input: input,
			setup: func() {
				tx := &sql.Tx{}
				s.vehicleRepository.On("BeginTx", mock.Anything).Return(tx, nil)
				s.vehicleRepository.On("HandleTransaction", mock.Anything, tx, mock.Anything).Return(nil)

				s.vehicleRepository.On("UpdateVehicle", mock.Anything, tx, mock.Anything).Return(dummyRepoVehicle, nil)
				s.vehicleRepository.On("DeleteAllImagesForVehicle", mock.Anything, tx, mock.Anything).Return(nil)
				s.vehicleRepository.On("CreateVehicleImage", mock.Anything, tx, mock.Anything).Return(repository.VehicleImage{}, apperrors.ErrInvalidImageToLink)
			},
			expectedError:  apperrors.ErrInvalidRequestBody,
			expectedResult: vehicle.Vehicle{},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup()
			}

			result, err := s.service.UpdateVehicle(ctx, tt.input, 1)

			s.Equal(tt.expectedError, err)
			s.Equal(tt.expectedResult, result)
		})
		s.TearDownTest()
	}
}

func (s *VehicleServiceTestSuite) Test_SoftDeleteVehicle() {
	type testCaseStruct struct {
		name          string
		setup         func()
		expectedError error
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			setup: func() {
				s.vehicleRepository.On("SoftDeleteVehicle", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Fail",
			setup: func() {
				s.vehicleRepository.On("SoftDeleteVehicle", mock.Anything, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup()
			}

			err := s.service.SoftDeleteVehicle(ctx, 1)

			s.Equal(tt.expectedError, err)
		})
		s.TearDownTest()
	}
}

func (s *VehicleServiceTestSuite) Test_GenerateSignedVehicleImageUploadURL() {
	type testCaseStruct struct {
		name           string
		setup          func()
		expectedError  error
		expectedResult string
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			setup: func() {
				s.firebaseService.On("GenerateSignedURL", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("abc", nil)
			},
			expectedError:  nil,
			expectedResult: "abc",
		},
		{
			name: "Fail",
			setup: func() {
				s.firebaseService.On("GenerateSignedURL", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: "",
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup()
			}

			result, _, err := s.service.GenerateSignedVehicleImageUploadURL(ctx, "")

			s.Equal(tt.expectedError, err)
			s.Equal(tt.expectedResult, result)
		})
		s.TearDownTest()
	}
}

func (s *VehicleServiceTestSuite) Test_GetVehicleById() {
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

	dummyRepoVehicle := repository.Vehicle{
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
		Available:             true,
		HostId:                101,
		IsDeleted:             false,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	dummyRepoImages := []repository.VehicleImage{
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
	}

	type testCaseStruct struct {
		name           string
		setup          func()
		expectedError  error
		expectedResult vehicle.Vehicle
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			setup: func() {
				s.vehicleRepository.On("GetVehicleById", mock.Anything, mock.Anything, mock.Anything).Return(dummyRepoVehicle, nil)
				s.vehicleRepository.On("GetVehicleImagesByVehicleId", mock.Anything, mock.Anything, mock.Anything).Return(dummyRepoImages, nil)
			},
			expectedError:  nil,
			expectedResult: dummyVehicle,
		},
		{
			name: "Failed to get vehicle",
			setup: func() {
				s.vehicleRepository.On("GetVehicleById", mock.Anything, mock.Anything, mock.Anything).Return(dummyRepoVehicle, apperrors.ErrVehicleNotFound)
			},
			expectedError:  apperrors.ErrVehicleNotFound,
			expectedResult: vehicle.Vehicle{},
		},
		{
			name: "Failed to get vehicle images",
			setup: func() {
				s.vehicleRepository.On("GetVehicleById", mock.Anything, mock.Anything, mock.Anything).Return(dummyRepoVehicle, nil)
				s.vehicleRepository.On("GetVehicleImagesByVehicleId", mock.Anything, mock.Anything, mock.Anything).Return([]repository.VehicleImage{}, apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: vehicle.Vehicle{},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup()
			}

			result, err := s.service.GetVehicleById(ctx, 1)

			s.Equal(tt.expectedError, err)
			s.Equal(tt.expectedResult, result)
		})
		s.TearDownTest()
	}
}

func (s *VehicleServiceTestSuite) Test_GetVehicles() {
	now := time.Now()

	dummyVehicles := vehicle.PaginatedVehicleOverview{
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

	dummyRepoVehicles := []repository.VehicleOverview{
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
	}

	type testCaseStruct struct {
		name           string
		input          vehicle.GetVehiclesParams
		setup          func()
		expectedError  error
		expectedResult vehicle.PaginatedVehicleOverview
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			input: vehicle.GetVehiclesParams{
				Page:             1,
				Limit:            10,
				City:             "Pune",
				PickupTimestamp:  now,
				DropoffTimestamp: now.Add(time.Minute),
			},
			setup: func() {
				s.vehicleRepository.On("GetVehicles", mock.Anything, mock.Anything, mock.Anything).Return(dummyRepoVehicles, 3, nil)
			},
			expectedError:  nil,
			expectedResult: dummyVehicles,
		},
		{
			name: "Invalid timestamp",
			input: vehicle.GetVehiclesParams{
				Page:             1,
				Limit:            10,
				City:             "Pune",
				PickupTimestamp:  now,
				DropoffTimestamp: now.Add(-1 * time.Minute),
			},
			expectedError:  apperrors.ErrInvalidPickupDropoff,
			expectedResult: vehicle.PaginatedVehicleOverview{},
		},
		{
			name: "Invalid page",
			input: vehicle.GetVehiclesParams{
				Page:             -1,
				Limit:            10,
				City:             "Pune",
				PickupTimestamp:  now,
				DropoffTimestamp: now.Add(time.Minute),
			},
			expectedError:  apperrors.ErrInvalidPagination,
			expectedResult: vehicle.PaginatedVehicleOverview{},
		},
		{
			name: "Invalid limit",
			input: vehicle.GetVehiclesParams{
				Page:             1,
				Limit:            -10,
				City:             "Pune",
				PickupTimestamp:  now,
				DropoffTimestamp: now.Add(time.Minute),
			},
			expectedError:  apperrors.ErrInvalidPagination,
			expectedResult: vehicle.PaginatedVehicleOverview{},
		},
		{
			name: "Failed to get vehicles",
			input: vehicle.GetVehiclesParams{
				Page:             1,
				Limit:            10,
				City:             "Pune",
				PickupTimestamp:  now,
				DropoffTimestamp: now.Add(time.Minute),
			},
			setup: func() {
				s.vehicleRepository.On("GetVehicles", mock.Anything, mock.Anything, mock.Anything).Return([]repository.VehicleOverview{}, 0, apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: vehicle.PaginatedVehicleOverview{},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup()
			}

			result, err := s.service.GetVehicles(ctx, tt.input)

			s.Equal(tt.expectedError, err)
			s.Equal(tt.expectedResult, result)
		})
		s.TearDownTest()
	}
}

func (s *VehicleServiceTestSuite) Test_GetVehiclesForHost() {
	dummyVehicles := vehicle.PaginatedVehicleOverview{
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

	dummyRepoVehicles := []repository.VehicleOverview{
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
	}

	type testCaseStruct struct {
		name           string
		page           int
		limit          int
		setup          func(ctx *context.Context)
		expectedError  error
		expectedResult vehicle.PaginatedVehicleOverview
	}

	testCases := []testCaseStruct{
		{
			name:  "Success",
			page:  1,
			limit: 10,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.vehicleRepository.On("GetVehiclesForHost", mock.Anything, mock.Anything, mock.Anything).Return(dummyRepoVehicles, 3, nil)
			},
			expectedError:  nil,
			expectedResult: dummyVehicles,
		},
		{
			name:           "Failed to get user if from context",
			page:           -1,
			limit:          10,
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: vehicle.PaginatedVehicleOverview{},
		},
		{
			name:  "Invalid page",
			page:  -1,
			limit: 10,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)
			},
			expectedError:  apperrors.ErrInvalidPagination,
			expectedResult: vehicle.PaginatedVehicleOverview{},
		},
		{
			name:  "Invalid limit",
			page:  1,
			limit: -10,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)
			},
			expectedError:  apperrors.ErrInvalidPagination,
			expectedResult: vehicle.PaginatedVehicleOverview{},
		},
		{
			name:  "Failed to get vehicles",
			page:  1,
			limit: 10,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.vehicleRepository.On("GetVehiclesForHost", mock.Anything, mock.Anything, mock.Anything).Return([]repository.VehicleOverview{}, 0, apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: vehicle.PaginatedVehicleOverview{},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup(&ctx)
			}

			result, err := s.service.GetVehiclesForHost(ctx, tt.page, tt.limit)

			s.Equal(tt.expectedError, err)
			s.Equal(tt.expectedResult, result)
		})
		s.TearDownTest()
	}
}
