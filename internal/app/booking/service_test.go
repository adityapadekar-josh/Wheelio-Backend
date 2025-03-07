package booking_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/booking"
	emailMocks "github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/email/mocks"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/user"
	userMocks "github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/user/mocks"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/vehicle"
	vehicleMocks "github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/vehicle/mocks"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/cryptokit"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/middleware"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository"
	repositoryMocks "github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/undefinedlabs/go-mpatch"
)

type BookingServiceTestSuite struct {
	suite.Suite
	service           booking.Service
	bookingRepository *repositoryMocks.BookingRepository
	userService       *userMocks.Service
	vehicleService    *vehicleMocks.Service
	emailService      *emailMocks.Service
	patches           []*mpatch.Patch
}

func Test_BookingServiceTestSuite(t *testing.T) {
	suite.Run(t, new(BookingServiceTestSuite))
}

func (s *BookingServiceTestSuite) SetupTest() {
	s.bookingRepository = &repositoryMocks.BookingRepository{}
	s.userService = &userMocks.Service{}
	s.vehicleService = &vehicleMocks.Service{}
	s.emailService = &emailMocks.Service{}
	s.patches = make([]*mpatch.Patch, 0)

	s.service = booking.NewService(s.bookingRepository, s.userService, s.vehicleService, s.emailService)
}

func (s *BookingServiceTestSuite) TearDownTest() {
	for _, p := range s.patches {
		err := p.Unpatch()
		require.NoError(s.T(), err)
	}
	s.patches = make([]*mpatch.Patch, 0)

	s.bookingRepository.AssertExpectations(s.T())
	s.userService.AssertExpectations(s.T())
	s.vehicleService.AssertExpectations(s.T())
	s.emailService.AssertExpectations(s.T())
}

func (s *BookingServiceTestSuite) Test_CreateBooking() {
	now := time.Now()

	input := booking.CreateBookingRequestBody{
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
		ActualPickupTime:      nil,
		ActualDropoffTime:     nil,
		ScheduledPickupTime:   now,
		ScheduledDropoffTime:  now,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	type testCaseStruct struct {
		name           string
		input          booking.CreateBookingRequestBody
		setup          func(ctx *context.Context)
		expectedError  error
		expectedResult booking.Booking
	}

	testCases := []testCaseStruct{
		{
			name:  "Success",
			input: input,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.vehicleService.On("GetVehicleById", *ctx, mock.Anything).Return(vehicle.Vehicle{}, nil)
				s.bookingRepository.On("VehicleBookingConflictCheck", *ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				s.userService.On("GetUserById", *ctx, mock.Anything).Return(user.User{}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("CreateBooking", *ctx, tx, mock.Anything).Return(repository.Booking(dummyBooking), nil)
				s.bookingRepository.On("CreateOtpToken", *ctx, tx, mock.Anything).Return(nil)
				s.emailService.On("SendEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError:  nil,
			expectedResult: dummyBooking,
		},
		{
			name:           "No user id in context",
			input:          input,
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: booking.Booking{},
		},
		{
			name: "Failed validation 1",
			input: booking.CreateBookingRequestBody{
				VehicleId:             -101,
				HostId:                -15,
				SeekerId:              -25,
				Status:                "CONFIfeRMED",
				PickupLocation:        "",
				DropoffLocation:       "",
				BookingAmount:         2500.50,
				OverdueFeeRatePerHour: 500.00,
				CancellationAllowed:   true,
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)
			},
			expectedError:  apperrors.ErrInvalidRequestBody,
			expectedResult: booking.Booking{},
		},
		{
			name: "Failed validation 2",
			input: booking.CreateBookingRequestBody{
				VehicleId:             -101,
				HostId:                -15,
				SeekerId:              -25,
				Status:                "CONFIfeRMED",
				PickupLocation:        "",
				DropoffLocation:       "",
				BookingAmount:         2500.50,
				OverdueFeeRatePerHour: 500.00,
				CancellationAllowed:   true,
				ScheduledPickupTime:   now.Add(time.Minute),
				ScheduledDropoffTime:  now,
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)
			},
			expectedError:  apperrors.ErrInvalidRequestBody,
			expectedResult: booking.Booking{},
		},
		{
			name:  "Failed to get vehicle",
			input: input,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.vehicleService.On("GetVehicleById", *ctx, mock.Anything).Return(vehicle.Vehicle{}, apperrors.ErrVehicleNotFound)
			},
			expectedError:  apperrors.ErrVehicleNotFound,
			expectedResult: booking.Booking{},
		},
		{
			name:  "Vehicle is deleted",
			input: input,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.vehicleService.On("GetVehicleById", *ctx, mock.Anything).Return(vehicle.Vehicle{IsDeleted: true}, nil)
			},
			expectedError:  apperrors.ErrVehicleNotFound,
			expectedResult: booking.Booking{},
		},
		{
			name:  "Booking slot conflict",
			input: input,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.vehicleService.On("GetVehicleById", *ctx, mock.Anything).Return(vehicle.Vehicle{}, nil)
				s.bookingRepository.On("VehicleBookingConflictCheck", *ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(apperrors.ErrBookingConflict)
			},
			expectedError:  apperrors.ErrBookingConflict,
			expectedResult: booking.Booking{},
		},
		{
			name:  "Failed to get user",
			input: input,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.vehicleService.On("GetVehicleById", *ctx, mock.Anything).Return(vehicle.Vehicle{}, nil)
				s.bookingRepository.On("VehicleBookingConflictCheck", *ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				s.userService.On("GetUserById", *ctx, mock.Anything).Return(user.User{}, apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: booking.Booking{},
		},
		{
			name:  "Failed to begin transaction",
			input: input,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.vehicleService.On("GetVehicleById", *ctx, mock.Anything).Return(vehicle.Vehicle{}, nil)
				s.bookingRepository.On("VehicleBookingConflictCheck", *ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				s.userService.On("GetUserById", *ctx, mock.Anything).Return(user.User{}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: booking.Booking{},
		},
		{
			name:  "Failed to create booking",
			input: input,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.vehicleService.On("GetVehicleById", *ctx, mock.Anything).Return(vehicle.Vehicle{}, nil)
				s.bookingRepository.On("VehicleBookingConflictCheck", *ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				s.userService.On("GetUserById", *ctx, mock.Anything).Return(user.User{}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("CreateBooking", *ctx, tx, mock.Anything).Return(repository.Booking{}, apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: booking.Booking{},
		},
		{
			name:  "Failed to handle transaction",
			input: input,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.vehicleService.On("GetVehicleById", *ctx, mock.Anything).Return(vehicle.Vehicle{}, nil)
				s.bookingRepository.On("VehicleBookingConflictCheck", *ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				s.userService.On("GetUserById", *ctx, mock.Anything).Return(user.User{}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(apperrors.ErrInternalServer)
				s.bookingRepository.On("CreateBooking", *ctx, tx, mock.Anything).Return(repository.Booking(dummyBooking), apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: booking.Booking{},
		},
		{
			name:  "Failed to generate otp token",
			input: input,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.vehicleService.On("GetVehicleById", *ctx, mock.Anything).Return(vehicle.Vehicle{}, nil)
				s.bookingRepository.On("VehicleBookingConflictCheck", *ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				s.userService.On("GetUserById", *ctx, mock.Anything).Return(user.User{}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("CreateBooking", *ctx, tx, mock.Anything).Return(repository.Booking(dummyBooking), nil)
				p, err := mpatch.PatchMethod(cryptokit.GenerateOTP, func() (string, error) {
					return "", apperrors.ErrInternalServer
				})
				require.NoError(s.T(), err)

				s.patches = append(s.patches, p)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: booking.Booking{},
		},
		{
			name:  "Failed to create otp token",
			input: input,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.vehicleService.On("GetVehicleById", *ctx, mock.Anything).Return(vehicle.Vehicle{}, nil)
				s.bookingRepository.On("VehicleBookingConflictCheck", *ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				s.userService.On("GetUserById", *ctx, mock.Anything).Return(user.User{}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("CreateBooking", *ctx, tx, mock.Anything).Return(repository.Booking(dummyBooking), nil)
				s.bookingRepository.On("CreateOtpToken", *ctx, tx, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: booking.Booking{},
		},
		{
			name:  "Failed to send email",
			input: input,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.vehicleService.On("GetVehicleById", *ctx, mock.Anything).Return(vehicle.Vehicle{}, nil)
				s.bookingRepository.On("VehicleBookingConflictCheck", *ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				s.userService.On("GetUserById", *ctx, mock.Anything).Return(user.User{}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("CreateBooking", *ctx, tx, mock.Anything).Return(repository.Booking(dummyBooking), nil)
				s.bookingRepository.On("CreateOtpToken", *ctx, tx, mock.Anything).Return(nil)
				s.emailService.On("SendEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: booking.Booking{},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup(&ctx)
			}

			result, err := s.service.CreateBooking(ctx, tt.input)

			s.Equal(tt.expectedError, err)
			s.Equal(tt.expectedResult, result)
		})
		s.TearDownTest()
	}
}

func (s *BookingServiceTestSuite) Test_CancelBooking() {
	type testCaseStruct struct {
		name          string
		input         int
		setup         func(ctx *context.Context)
		expectedError error
	}

	testCases := []testCaseStruct{
		{
			name:  "Success 1",
			input: 1,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1}, nil)
				s.bookingRepository.On("UpdateBookingStatus", *ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "Success 2",
			input: 1,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{SeekerId: 1, CancellationAllowed: true}, nil)
				s.bookingRepository.On("UpdateBookingStatus", *ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "No user id in context",
			input:         1,
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name:  "Failed to get booking",
			input: 1,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{}, apperrors.ErrBookingNotFound)
			},
			expectedError: apperrors.ErrBookingNotFound,
		},
		{
			name:  "Neither HOST or SEEKER",
			input: 1,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{}, nil)
			},
			expectedError: apperrors.ErrActionForbidden,
		},
		{
			name:  "Seeker but cancellation not allowed",
			input: 1,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{SeekerId: 1}, nil)
			},
			expectedError: apperrors.ErrBookingCancellationNotAllowed,
		},
		{
			name:  "Failed to update booking status",
			input: 1,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1}, nil)
				s.bookingRepository.On("UpdateBookingStatus", *ctx, mock.Anything, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup(&ctx)
			}

			err := s.service.CancelBooking(ctx, tt.input)

			s.Equal(tt.expectedError, err)
		})
		s.TearDownTest()
	}
}

func (s *BookingServiceTestSuite) Test_ConfirmPickup() {

	type testCaseStruct struct {
		name          string
		params        int
		input         booking.OtpRequestBody
		setup         func(ctx *context.Context)
		expectedError error
	}

	testCases := []testCaseStruct{
		{
			name:   "Success",
			params: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.Scheduled}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 1, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateBookingStatus", *ctx, tx, mock.Anything, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateActualPickupTime", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("DeleteOtpTokenById", *ctx, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "No user id in context",
			params: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name:   "Failed to get booking id",
			params: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{}, apperrors.ErrBookingNotFound)
			},
			expectedError: apperrors.ErrBookingNotFound,
		},
		{
			name:   "Host id does not match",
			params: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 2, Status: booking.Scheduled}, nil)
			},
			expectedError: apperrors.ErrActionForbidden,
		},
		{
			name:   "Status is not scheduled",
			params: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.Cancelled}, nil)
			},
			expectedError: apperrors.ErrActionForbidden,
		},
		{
			name:   "Failed to get opt token",
			params: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.Scheduled}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{}, apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInvalidOtp,
		},
		{
			name:   "Otp token expired",
			params: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.Scheduled}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 1, ExpiresAt: time.Now().Add(-1 * time.Minute)}, nil)
			},
			expectedError: apperrors.ErrInvalidOtp,
		},
		{
			name:   "Booking id does not match",
			params: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.Scheduled}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 2, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil)
			},
			expectedError: apperrors.ErrInvalidOtp,
		},
		{
			name:   "Failed to begin transaction",
			params: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.Scheduled}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 1, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(nil, apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name:   "Failed to update booking status",
			params: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.Scheduled}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 1, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateBookingStatus", *ctx, tx, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name:   "Failed to handle transaction",
			params: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.Scheduled}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 1, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(apperrors.ErrInternalServer)
				s.bookingRepository.On("UpdateBookingStatus", *ctx, tx, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name:   "Failed to update actual pickup time",
			params: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.Scheduled}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 1, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateBookingStatus", *ctx, tx, mock.Anything, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateActualPickupTime", *ctx, tx, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name:   "Warn failed to delete otp token",
			params: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.Scheduled}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 1, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateBookingStatus", *ctx, tx, mock.Anything, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateActualPickupTime", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("DeleteOtpTokenById", *ctx, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: nil,
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup(&ctx)
			}

			err := s.service.ConfirmPickup(ctx, tt.params, tt.input)

			s.Equal(tt.expectedError, err)
		})
		s.TearDownTest()
	}
}

func (s *BookingServiceTestSuite) Test_InitiateReturn() {
	type testCaseStruct struct {
		name          string
		input         int
		setup         func(ctx *context.Context)
		expectedError error
	}

	testCases := []testCaseStruct{
		{
			name:  "Success",
			input: 1,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.CheckedOut}, nil)
				s.userService.On("GetUserById", *ctx, mock.Anything).Return(user.User{}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("CreateOtpToken", *ctx, mock.Anything, mock.Anything).Return(nil)
				s.emailService.On("SendEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "No id in context",
			input:         1,
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name:  "Failed to get booking",
			input: 1,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{}, apperrors.ErrBookingNotFound)

			},
			expectedError: apperrors.ErrBookingNotFound,
		},
		{
			name:  "Host id does not match",
			input: 1,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 2, Status: booking.CheckedOut}, nil)
			},
			expectedError: apperrors.ErrActionForbidden,
		},
		{
			name:  "Status is not checkout",
			input: 1,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.Cancelled}, nil)
			},
			expectedError: apperrors.ErrActionForbidden,
		},
		{
			name:  "Failed to get user",
			input: 1,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.CheckedOut}, nil)
				s.userService.On("GetUserById", *ctx, mock.Anything).Return(user.User{}, apperrors.ErrUserNotFound)
			},
			expectedError: apperrors.ErrUserNotFound,
		},
		{
			name:  "Failed to generate otp token",
			input: 1,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.CheckedOut}, nil)
				s.userService.On("GetUserById", *ctx, mock.Anything).Return(user.User{}, nil)
				p, err := mpatch.PatchMethod(cryptokit.GenerateOTP, func() (string, error) {
					return "", apperrors.ErrInternalServer
				})
				require.NoError(s.T(), err)

				s.patches = append(s.patches, p)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name:  "Failed to begin transaction",
			input: 1,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.CheckedOut}, nil)
				s.userService.On("GetUserById", *ctx, mock.Anything).Return(user.User{}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(nil, apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name:  "Failed to create otp token",
			input: 1,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.CheckedOut}, nil)
				s.userService.On("GetUserById", *ctx, mock.Anything).Return(user.User{}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("CreateOtpToken", *ctx, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name:  "Failed to handle transaction",
			input: 1,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.CheckedOut}, nil)
				s.userService.On("GetUserById", *ctx, mock.Anything).Return(user.User{}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(apperrors.ErrInternalServer)
				s.bookingRepository.On("CreateOtpToken", *ctx, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name:  "Failed to send email",
			input: 1,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{HostId: 1, Status: booking.CheckedOut}, nil)
				s.userService.On("GetUserById", *ctx, mock.Anything).Return(user.User{}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("CreateOtpToken", *ctx, mock.Anything, mock.Anything).Return(nil)
				s.emailService.On("SendEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup(&ctx)
			}

			err := s.service.InitiateReturn(ctx, tt.input)

			s.Equal(tt.expectedError, err)
		})
		s.TearDownTest()
	}
}

func (s *BookingServiceTestSuite) Test_ConfirmReturn() {
	type testCaseStruct struct {
		name          string
		param         int
		input         booking.OtpRequestBody
		setup         func(ctx *context.Context)
		expectedError error
	}

	testCases := []testCaseStruct{
		{
			name:  "Success 1",
			param: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{SeekerId: 1, Status: booking.CheckedOut}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 1, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateBookingStatus", *ctx, tx, mock.Anything, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateActualDropoffTime", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("CreateInvoice", *ctx, tx, mock.Anything).Return(repository.Invoice{}, nil)
				s.bookingRepository.On("DeleteOtpTokenById", *ctx, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "Success 2",
			param: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{SeekerId: 1, Status: booking.CheckedOut, ScheduledDropoffTime: time.Now().Add(50 * time.Minute)}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 1, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateBookingStatus", *ctx, tx, mock.Anything, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateActualDropoffTime", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("CreateInvoice", *ctx, tx, mock.Anything).Return(repository.Invoice{}, nil)
				s.bookingRepository.On("DeleteOtpTokenById", *ctx, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "No id in context",
			param: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name:  "Failed to get booking",
			param: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{}, apperrors.ErrBookingNotFound)
			},
			expectedError: apperrors.ErrBookingNotFound,
		},
		{
			name:  "Seeker id does not match",
			param: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{SeekerId: 2, Status: booking.CheckedOut}, nil)
			},
			expectedError: apperrors.ErrActionForbidden,
		},
		{
			name:  "Status does not match",
			param: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{SeekerId: 1, Status: booking.Cancelled}, nil)
			},
			expectedError: apperrors.ErrActionForbidden,
		},
		{
			name:  "Failed to get otp token",
			param: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{SeekerId: 1, Status: booking.CheckedOut}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{}, apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInvalidOtp,
		},
		{
			name:  "Token expired",
			param: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{SeekerId: 1, Status: booking.CheckedOut}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 1, ExpiresAt: time.Now().Add(-10 * time.Minute)}, nil)
			},
			expectedError: apperrors.ErrInvalidOtp,
		},
		{
			name:  "Booking id does not match",
			param: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{SeekerId: 1, Status: booking.CheckedOut}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 2, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil)
			},
			expectedError: apperrors.ErrInvalidOtp,
		},
		{
			name:  "Failed to start transaction",
			param: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{SeekerId: 1, Status: booking.CheckedOut}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 1, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(nil, apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name:  "Failed to update booking status",
			param: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{SeekerId: 1, Status: booking.CheckedOut}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 1, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateBookingStatus", *ctx, tx, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name:  "Failed to handle transactions",
			param: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{SeekerId: 1, Status: booking.CheckedOut}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 1, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(apperrors.ErrInternalServer)
				s.bookingRepository.On("UpdateBookingStatus", *ctx, tx, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name:  "Failed to update actual dropoff time stamp",
			param: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{SeekerId: 1, Status: booking.CheckedOut}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 1, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateBookingStatus", *ctx, tx, mock.Anything, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateActualDropoffTime", *ctx, tx, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name:  "Failed to create invoice",
			param: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{SeekerId: 1, Status: booking.CheckedOut}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 1, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateBookingStatus", *ctx, tx, mock.Anything, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateActualDropoffTime", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("CreateInvoice", *ctx, tx, mock.Anything).Return(repository.Invoice{}, apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name:  "Warn failed to delete otp tokens",
			param: 1,
			input: booking.OtpRequestBody{
				Otp: "123123",
			},
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				tx := &sql.Tx{}
				s.bookingRepository.On("GetBookingById", *ctx, mock.Anything, mock.Anything).Return(repository.Booking{SeekerId: 1, Status: booking.CheckedOut}, nil)
				s.bookingRepository.On("GetOtpToken", *ctx, mock.Anything, mock.Anything).Return(repository.OtpToken{BookingId: 1, ExpiresAt: time.Now().Add(10 * time.Minute)}, nil)
				s.bookingRepository.On("BeginTx", *ctx).Return(tx, nil)
				s.bookingRepository.On("HandleTransaction", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateBookingStatus", *ctx, tx, mock.Anything, mock.Anything).Return(nil)
				s.bookingRepository.On("UpdateActualDropoffTime", *ctx, tx, mock.Anything).Return(nil)
				s.bookingRepository.On("CreateInvoice", *ctx, tx, mock.Anything).Return(repository.Invoice{}, nil)
				s.bookingRepository.On("DeleteOtpTokenById", *ctx, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: nil,
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup(&ctx)
			}

			err := s.service.ConfirmReturn(ctx, tt.param, tt.input)

			s.Equal(tt.expectedError, err)
		})
		s.TearDownTest()
	}
}

func (s *BookingServiceTestSuite) Test_GetSeekerBookings() {
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

	dummyBookingRepo := []repository.BookingData{
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
	}

	type testCaseStruct struct {
		name           string
		page           int
		limit          int
		setup          func(ctx *context.Context)
		expectedError  error
		expectedResult booking.PaginatedBookingData
	}

	testCases := []testCaseStruct{
		{
			name:  "Success",
			page:  1,
			limit: 10,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetSeekerBookings", *ctx, mock.Anything, mock.Anything).Return(dummyBookingRepo, 2, nil)
			},
			expectedError:  nil,
			expectedResult: dummyBooking,
		},
		{
			name:           "No id in context",
			page:           1,
			limit:          10,
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: booking.PaginatedBookingData{},
		},
		{
			name:  "Invalid page",
			page:  -1,
			limit: 10,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)
			},
			expectedError:  apperrors.ErrInvalidPagination,
			expectedResult: booking.PaginatedBookingData{},
		},
		{
			name:  "Invalid limit",
			page:  1,
			limit: -10,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)
			},
			expectedError:  apperrors.ErrInvalidPagination,
			expectedResult: booking.PaginatedBookingData{},
		},
		{
			name:  "Failed to get bookings",
			page:  1,
			limit: 10,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetSeekerBookings", *ctx, mock.Anything, mock.Anything).Return([]repository.BookingData{}, 0, apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: booking.PaginatedBookingData{},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup(&ctx)
			}

			result, err := s.service.GetSeekerBookings(ctx, tt.page, tt.limit)

			s.Equal(tt.expectedError, err)
			s.Equal(tt.expectedResult, result)
		})
		s.TearDownTest()
	}
}

func (s *BookingServiceTestSuite) Test_GetHostBookings() {
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

	dummyBookingRepo := []repository.BookingData{
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
	}

	type testCaseStruct struct {
		name           string
		page           int
		limit          int
		setup          func(ctx *context.Context)
		expectedError  error
		expectedResult booking.PaginatedBookingData
	}

	testCases := []testCaseStruct{
		{
			name:  "Success",
			page:  1,
			limit: 10,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetHostBookings", *ctx, mock.Anything, mock.Anything).Return(dummyBookingRepo, 2, nil)
			},
			expectedError:  nil,
			expectedResult: dummyBooking,
		},
		{
			name:           "No id in context",
			page:           1,
			limit:          10,
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: booking.PaginatedBookingData{},
		},
		{
			name:  "Invalid page",
			page:  -1,
			limit: 10,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)
			},
			expectedError:  apperrors.ErrInvalidPagination,
			expectedResult: booking.PaginatedBookingData{},
		},
		{
			name:  "Invalid limit",
			page:  1,
			limit: -10,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)
			},
			expectedError:  apperrors.ErrInvalidPagination,
			expectedResult: booking.PaginatedBookingData{},
		},
		{
			name:  "Failed to get bookings",
			page:  1,
			limit: 10,
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)

				s.bookingRepository.On("GetHostBookings", *ctx, mock.Anything, mock.Anything).Return([]repository.BookingData{}, 0, apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: booking.PaginatedBookingData{},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup(&ctx)
			}

			result, err := s.service.GetHostBookings(ctx, tt.page, tt.limit)

			s.Equal(tt.expectedError, err)
			s.Equal(tt.expectedResult, result)
		})
		s.TearDownTest()
	}
}

func (s *BookingServiceTestSuite) Test_GetBookingDetailsById() {
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

	dummyBookingRepo := repository.BookingDetails{
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
		Host: repository.BookingDetailsUser{
			Id:          10,
			Name:        "John Doe",
			Email:       "john.doe@example.com",
			PhoneNumber: "+1234567890",
		},
		Seeker: repository.BookingDetailsUser{
			Id:          20,
			Name:        "Alice Smith",
			Email:       "alice.smith@example.com",
			PhoneNumber: "+9876543210",
		},
		Vehicle: repository.BookingDetailsVehicle{
			Id:               5,
			Name:             "Toyota Corolla",
			FuelType:         "Petrol",
			SeatCount:        4,
			TransmissionType: "Automatic",
			Image:            "https://example.com/vehicle1.jpg",
		},
		Invoice: repository.BookingDetailsInvoice{
			Id:             1001,
			AdditionalFees: 200.00,
			Tax:            180.00,
			TaxRate:        9.0,
			TotalAmount:    2880.00,
		},
	}

	type testCaseStruct struct {
		name           string
		params         int
		setup          func()
		expectedError  error
		expectedResult booking.BookingDetails
	}

	testCases := []testCaseStruct{
		{
			name:   "Success",
			params: 1,
			setup: func() {
				s.bookingRepository.On("GetBookingDetailsById", mock.Anything, mock.Anything, mock.Anything).Return(dummyBookingRepo, nil)
			},
			expectedError:  nil,
			expectedResult: dummyBooking,
		},
		{
			name:   "Failed to get booking details",
			params: 1,
			setup: func() {
				s.bookingRepository.On("GetBookingDetailsById", mock.Anything, mock.Anything, mock.Anything).Return(repository.BookingDetails{}, apperrors.ErrInternalServer)
			},
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: booking.BookingDetails{},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup()
			}

			result, err := s.service.GetBookingDetailsById(ctx, tt.params)

			s.Equal(tt.expectedError, err)
			s.Equal(tt.expectedResult, result)
		})
		s.TearDownTest()
	}
}
