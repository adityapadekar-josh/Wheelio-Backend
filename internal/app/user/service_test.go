package user_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	emailMocks "github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/email/mocks"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/user"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/cryptokit"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/middleware"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository"
	repositoryMocks "github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository/mocks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/undefinedlabs/go-mpatch"
)

type UserServiceTestSuite struct {
	suite.Suite
	service        user.Service
	emailService   *emailMocks.Service
	userRepository *repositoryMocks.UserRepository
	patches        []*mpatch.Patch
}

func Test_UserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (s *UserServiceTestSuite) SetupTest() {
	s.userRepository = &repositoryMocks.UserRepository{}
	s.emailService = &emailMocks.Service{}
	s.patches = make([]*mpatch.Patch, 0)

	s.service = user.NewService(s.userRepository, s.emailService)
}

func (s *UserServiceTestSuite) TearDownTest() {
	for _, p := range s.patches {
		err := p.Unpatch()
		require.NoError(s.T(), err)
	}
	s.patches = make([]*mpatch.Patch, 0)

	s.userRepository.AssertExpectations(s.T())
	s.emailService.AssertExpectations(s.T())

}

func (s *UserServiceTestSuite) Test_SignUpUser() {
	type testCaseStruct struct {
		name          string
		input         user.CreateUserRequestBody
		setup         func(ctx context.Context)
		expectedError error
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			input: user.CreateUserRequestBody{
				Name:        "Aditya",
				Email:       "adityarpadekar@gmail.com",
				PhoneNumber: "+919022722650",
				Password:    "aditya",
				Role:        user.Host,
			},
			setup: func(ctx context.Context) {
				tx := &sql.Tx{}

				s.userRepository.On("BeginTx", ctx).Return(tx, nil)
				s.userRepository.On("HandleTransaction", ctx, tx, mock.Anything).Return(nil)
				s.userRepository.On("CreateUser", ctx, tx, mock.Anything).Return(repository.User{}, nil)
				s.userRepository.On("CreateVerificationToken", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(repository.VerificationToken{}, nil)
				s.emailService.On("SendEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "Validation failed 1",
			input:         user.CreateUserRequestBody{},
			expectedError: apperrors.ErrInvalidRequestBody,
		},
		{
			name: "Validation failed 2",
			input: user.CreateUserRequestBody{
				Name:        "Aditya",
				Email:       "adityarpadekar",
				PhoneNumber: "+91902272265d0",
				Password:    "aditya",
				Role:        "DWA",
			},
			expectedError: apperrors.ErrInvalidRequestBody,
		},
		{
			name: "Db transaction creation failed",
			input: user.CreateUserRequestBody{
				Name:        "Aditya",
				Email:       "adityarpadekar@gmail.com",
				PhoneNumber: "+919022722650",
				Password:    "aditya",
				Role:        user.Host,
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("BeginTx", ctx).Return(nil, apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name: "Db transaction handling failed",
			input: user.CreateUserRequestBody{
				Name:        "Aditya",
				Email:       "adityarpadekar@gmail.com",
				PhoneNumber: "+919022722650",
				Password:    "aditya",
				Role:        user.Host,
			},
			setup: func(ctx context.Context) {
				tx := &sql.Tx{}
				s.userRepository.On("BeginTx", ctx).Return(tx, nil)
				s.userRepository.On("CreateUser", ctx, tx, mock.Anything).Return(repository.User{}, apperrors.ErrInternalServer)
				s.userRepository.On("HandleTransaction", ctx, tx, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name: "User creation failed",
			input: user.CreateUserRequestBody{
				Name:        "Aditya",
				Email:       "adityarpadekar@gmail.com",
				PhoneNumber: "+919022722650",
				Password:    "aditya",
				Role:        user.Host,
			},
			setup: func(ctx context.Context) {
				tx := &sql.Tx{}
				s.userRepository.On("BeginTx", ctx).Return(tx, nil)
				s.userRepository.On("CreateUser", ctx, tx, mock.Anything).Return(repository.User{}, apperrors.ErrInternalServer)
				s.userRepository.On("HandleTransaction", ctx, tx, mock.Anything).Return(nil)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name: "Email already registered",
			input: user.CreateUserRequestBody{
				Name:        "Aditya",
				Email:       "adityarpadekar@gmail.com",
				PhoneNumber: "+919022722650",
				Password:    "aditya",
				Role:        user.Host,
			},
			setup: func(ctx context.Context) {
				tx := &sql.Tx{}
				s.userRepository.On("BeginTx", ctx).Return(tx, nil)
				s.userRepository.On("CreateUser", ctx, tx, mock.Anything).Return(repository.User{}, apperrors.ErrEmailAlreadyRegistered)
				s.userRepository.On("HandleTransaction", ctx, tx, mock.Anything).Return(nil)
			},
			expectedError: apperrors.ErrEmailAlreadyRegistered,
		},
		{
			name: "Failed to generate secure token",
			input: user.CreateUserRequestBody{
				Name:        "Aditya",
				Email:       "adityarpadekar@gmail.com",
				PhoneNumber: "+919022722650",
				Password:    "aditya",
				Role:        user.Host,
			},
			setup: func(ctx context.Context) {
				tx := &sql.Tx{}
				s.userRepository.On("BeginTx", ctx).Return(tx, nil)
				s.userRepository.On("CreateUser", ctx, tx, mock.Anything).Return(repository.User{}, nil)
				s.userRepository.On("HandleTransaction", ctx, tx, mock.Anything).Return(nil)
				p, err := mpatch.PatchMethod(cryptokit.GenerateSecureToken, func(length int) (string, error) {
					return "", apperrors.ErrInternalServer
				})
				require.NoError(s.T(), err)

				s.patches = append(s.patches, p)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name: "Verification token creation failed",
			input: user.CreateUserRequestBody{
				Name:        "Aditya",
				Email:       "adityarpadekar@gmail.com",
				PhoneNumber: "+919022722650",
				Password:    "aditya",
				Role:        user.Host,
			},
			setup: func(ctx context.Context) {
				tx := &sql.Tx{}
				s.userRepository.On("BeginTx", ctx).Return(tx, nil)
				s.userRepository.On("CreateUser", ctx, tx, mock.Anything).Return(repository.User{}, nil)
				s.userRepository.On("CreateVerificationToken", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(repository.VerificationToken{}, apperrors.ErrInternalServer)
				s.userRepository.On("HandleTransaction", ctx, tx, mock.Anything).Return(nil)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name: "Email sending failed",
			input: user.CreateUserRequestBody{
				Name:        "Aditya",
				Email:       "adityarpadekar@gmail.com",
				PhoneNumber: "+919022722650",
				Password:    "aditya",
				Role:        user.Host,
			},
			setup: func(ctx context.Context) {
				tx := &sql.Tx{}
				s.userRepository.On("BeginTx", ctx).Return(tx, nil)
				s.userRepository.On("CreateUser", ctx, tx, mock.Anything).Return(repository.User{}, nil)
				s.userRepository.On("CreateVerificationToken", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(repository.VerificationToken{}, nil)
				s.emailService.On("SendEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(apperrors.ErrEmailSendFailed)
				s.userRepository.On("HandleTransaction", ctx, tx, mock.Anything).Return(nil)
			},
			expectedError: apperrors.ErrEmailSendFailed,
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup(ctx)
			}

			err := s.service.RegisterUser(ctx, tt.input)

			s.Equal(tt.expectedError, err)
		})
		s.TearDownTest()
	}
}

func (s *UserServiceTestSuite) Test_SignInUser() {
	type testCaseStruct struct {
		name          string
		input         user.LoginUserRequestBody
		setup         func(ctx context.Context)
		expectedError error
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			input: user.LoginUserRequestBody{
				Email:    "adityarpadekar@gmail.com",
				Password: "aditya",
			},
			setup: func(ctx context.Context) {
				hashedPassword, _ := cryptokit.HashPassword("aditya")

				s.userRepository.On("GetUserByEmail", ctx, mock.Anything, mock.Anything).Return(repository.User{IsVerified: true, Password: hashedPassword}, nil)
			},
			expectedError: nil,
		},
		{
			name:          "Validation failed 1",
			input:         user.LoginUserRequestBody{},
			expectedError: apperrors.ErrInvalidRequestBody,
		},
		{
			name: "Validation failed 2",
			input: user.LoginUserRequestBody{
				Email:    "adityarpadekar",
				Password: "aditya",
			},
			expectedError: apperrors.ErrInvalidRequestBody,
		},
		{
			name: "User not verified",
			input: user.LoginUserRequestBody{
				Email:    "adityarpadekar@gmail.com",
				Password: "aditya",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetUserByEmail", ctx, mock.Anything, mock.Anything).Return(repository.User{IsVerified: false}, nil)

			},
			expectedError: apperrors.ErrUserNotVerified,
		},
		{
			name: "Incorrect password",
			input: user.LoginUserRequestBody{
				Email:    "adityarpadekar@gmail.com",
				Password: "aditya",
			},
			setup: func(ctx context.Context) {
				hashedPassword, _ := cryptokit.HashPassword("wadw")

				s.userRepository.On("GetUserByEmail", ctx, mock.Anything, mock.Anything).Return(repository.User{IsVerified: true, Password: hashedPassword}, nil)

			},
			expectedError: apperrors.ErrInvalidLoginCredentials,
		},
		{
			name: "Incorrect email",
			input: user.LoginUserRequestBody{
				Email:    "adityarpadekar@gmail.com",
				Password: "aditya",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetUserByEmail", ctx, mock.Anything, mock.Anything).Return(repository.User{}, apperrors.ErrInternalServer)

			},
			expectedError: apperrors.ErrInvalidLoginCredentials,
		},
		{
			name: "Incorrect email",
			input: user.LoginUserRequestBody{
				Email:    "adityarpadekar@gmail.com",
				Password: "aditya",
			},
			setup: func(ctx context.Context) {
				hashedPassword, _ := cryptokit.HashPassword("aditya")
				s.userRepository.On("GetUserByEmail", ctx, mock.Anything, mock.Anything).Return(repository.User{IsVerified: true, Password: hashedPassword}, nil)

				p, err := mpatch.PatchMethod(cryptokit.CreateJWTToken, func(data jwt.MapClaims) (string, error) {
					return "", apperrors.ErrInternalServer
				})
				require.NoError(s.T(), err)

				s.patches = append(s.patches, p)

			},
			expectedError: apperrors.ErrJWTCreationFailed,
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup(ctx)
			}

			result, err := s.service.LoginUser(ctx, tt.input)

			s.Equal(tt.expectedError, err)
			if tt.expectedError == nil {
				s.NotEmpty(result)
			}
		})
		s.TearDownTest()
	}
}

func (s *UserServiceTestSuite) Test_VerifyEmail() {
	type testCaseStruct struct {
		name          string
		input         user.Token
		setup         func(ctx context.Context)
		expectedError error
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			input: user.Token{
				Token: "xyz",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetVerificationTokenByToken", ctx, mock.Anything, mock.Anything).Return(repository.VerificationToken{Type: user.EmailVerification, ExpiresAt: time.Now().Add(1 * time.Minute)}, nil)
				s.userRepository.On("UpdateUserEmailVerifiedStatus", ctx, mock.Anything, mock.Anything).Return(nil)
				s.userRepository.On("DeleteVerificationTokenById", ctx, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "Validation failed",
			input:         user.Token{},
			expectedError: apperrors.ErrInvalidRequestBody,
		},
		{
			name: "Token of wrong type",
			input: user.Token{
				Token: "xyz",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetVerificationTokenByToken", ctx, mock.Anything, mock.Anything).Return(repository.VerificationToken{Type: user.PasswordReset, ExpiresAt: time.Now().Add(1 * time.Minute)}, nil)
			},
			expectedError: apperrors.ErrInvalidToken,
		},
		{
			name: "Token if expired",
			input: user.Token{
				Token: "xyz",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetVerificationTokenByToken", ctx, mock.Anything, mock.Anything).Return(repository.VerificationToken{Type: user.EmailVerification, ExpiresAt: time.Now().Add(-1 * time.Minute)}, nil)
			},
			expectedError: apperrors.ErrInvalidToken,
		},
		{
			name: "Failed to update user verified status",
			input: user.Token{
				Token: "xyz",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetVerificationTokenByToken", ctx, mock.Anything, mock.Anything).Return(repository.VerificationToken{Type: user.EmailVerification, ExpiresAt: time.Now().Add(1 * time.Minute)}, nil)
				s.userRepository.On("UpdateUserEmailVerifiedStatus", ctx, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name: "Failed to delete verification token",
			input: user.Token{
				Token: "xyz",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetVerificationTokenByToken", ctx, mock.Anything, mock.Anything).Return(repository.VerificationToken{Type: user.EmailVerification, ExpiresAt: time.Now().Add(1 * time.Minute)}, nil)
				s.userRepository.On("UpdateUserEmailVerifiedStatus", ctx, mock.Anything, mock.Anything).Return(nil)
				s.userRepository.On("DeleteVerificationTokenById", ctx, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: nil,
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup(ctx)
			}

			err := s.service.VerifyEmail(ctx, tt.input)

			s.Equal(tt.expectedError, err)
		})
		s.TearDownTest()
	}
}

func (s *UserServiceTestSuite) Test_ForgotPassword() {
	type testCaseStruct struct {
		name          string
		input         user.Email
		setup         func(ctx context.Context)
		expectedError error
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			input: user.Email{
				Email: "aditya@gmail.com",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetUserByEmail", ctx, mock.Anything, mock.Anything).Return(repository.User{}, nil)
				s.userRepository.On("CreateVerificationToken", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(repository.VerificationToken{}, nil)
				s.emailService.On("SendEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "Validation failed 1",
			input:         user.Email{},
			expectedError: apperrors.ErrInvalidRequestBody,
		},
		{
			name:          "Validation failed 2",
			input:         user.Email{Email: "dwdwa"},
			expectedError: apperrors.ErrInvalidRequestBody,
		},
		{
			name: "User not found",
			input: user.Email{
				Email: "aditya@gmail.com",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetUserByEmail", ctx, mock.Anything, mock.Anything).Return(repository.User{}, apperrors.ErrUserNotFound)
			},
			expectedError: nil,
		},
		{
			name: "Failed to create secure token",
			input: user.Email{
				Email: "aditya@gmail.com",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetUserByEmail", ctx, mock.Anything, mock.Anything).Return(repository.User{}, nil)

				p, err := mpatch.PatchMethod(cryptokit.GenerateSecureToken, func(length int) (string, error) {
					return "", apperrors.ErrInternalServer
				})
				require.NoError(s.T(), err)

				s.patches = append(s.patches, p)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name: "Verification token creation failed",
			input: user.Email{
				Email: "aditya@gmail.com",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetUserByEmail", ctx, mock.Anything, mock.Anything).Return(repository.User{}, nil)
				s.userRepository.On("CreateVerificationToken", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(repository.VerificationToken{}, apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name: "Failed to send email",
			input: user.Email{
				Email: "aditya@gmail.com",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetUserByEmail", ctx, mock.Anything, mock.Anything).Return(repository.User{}, nil)
				s.userRepository.On("CreateVerificationToken", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(repository.VerificationToken{}, nil)
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
				tt.setup(ctx)
			}

			err := s.service.ForgotPassword(ctx, tt.input)

			s.Equal(tt.expectedError, err)
		})
		s.TearDownTest()
	}
}

func (s *UserServiceTestSuite) Test_ResetPassword() {
	type testCaseStruct struct {
		name          string
		input         user.ResetPasswordRequestBody
		setup         func(ctx context.Context)
		expectedError error
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			input: user.ResetPasswordRequestBody{
				Token:    "xyz",
				Password: "aditya",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetVerificationTokenByToken", ctx, mock.Anything, mock.Anything).Return(repository.VerificationToken{Type: user.PasswordReset, ExpiresAt: time.Now().Add(1 * time.Minute)}, nil)
				s.userRepository.On("GetUserById", ctx, mock.Anything, mock.Anything).Return(repository.User{}, nil)
				s.userRepository.On("UpdateUserPassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				s.userRepository.On("DeleteVerificationTokenById", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "Validation failed",
			input:         user.ResetPasswordRequestBody{},
			expectedError: apperrors.ErrInvalidRequestBody,
		},
		{
			name: "Invalid token type",
			input: user.ResetPasswordRequestBody{
				Token:    "xyz",
				Password: "aditya",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetVerificationTokenByToken", ctx, mock.Anything, mock.Anything).Return(repository.VerificationToken{Type: user.EmailVerification, ExpiresAt: time.Now().Add(1 * time.Minute)}, nil)
			},
			expectedError: apperrors.ErrInvalidToken,
		},
		{
			name: "Expired token",
			input: user.ResetPasswordRequestBody{
				Token:    "xyz",
				Password: "aditya",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetVerificationTokenByToken", ctx, mock.Anything, mock.Anything).Return(repository.VerificationToken{Type: user.PasswordReset, ExpiresAt: time.Now().Add(-1 * time.Minute)}, nil)
			},
			expectedError: apperrors.ErrInvalidToken,
		},
		{
			name: "Failed to get user",
			input: user.ResetPasswordRequestBody{
				Token:    "xyz",
				Password: "aditya",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetVerificationTokenByToken", ctx, mock.Anything, mock.Anything).Return(repository.VerificationToken{Type: user.PasswordReset, ExpiresAt: time.Now().Add(1 * time.Minute)}, nil)
				s.userRepository.On("GetUserById", ctx, mock.Anything, mock.Anything).Return(repository.User{}, apperrors.ErrUserNotFound)
			},
			expectedError: apperrors.ErrInvalidToken,
		},
		{
			name: "Failed to update user password",
			input: user.ResetPasswordRequestBody{
				Token:    "xyz",
				Password: "aditya",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetVerificationTokenByToken", ctx, mock.Anything, mock.Anything).Return(repository.VerificationToken{Type: user.PasswordReset, ExpiresAt: time.Now().Add(1 * time.Minute)}, nil)
				s.userRepository.On("GetUserById", ctx, mock.Anything, mock.Anything).Return(repository.User{}, nil)
				s.userRepository.On("UpdateUserPassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name: "Failed to delete verification token",
			input: user.ResetPasswordRequestBody{
				Token:    "xyz",
				Password: "aditya",
			},
			setup: func(ctx context.Context) {
				s.userRepository.On("GetVerificationTokenByToken", ctx, mock.Anything, mock.Anything).Return(repository.VerificationToken{Type: user.PasswordReset, ExpiresAt: time.Now().Add(1 * time.Minute)}, nil)
				s.userRepository.On("GetUserById", ctx, mock.Anything, mock.Anything).Return(repository.User{}, nil)
				s.userRepository.On("UpdateUserPassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				s.userRepository.On("DeleteVerificationTokenById", mock.Anything, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
			},
			expectedError: nil,
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup(ctx)
			}

			err := s.service.ResetPassword(ctx, tt.input)

			s.Equal(tt.expectedError, err)
		})
		s.TearDownTest()
	}
}

func (s *UserServiceTestSuite) Test_GetLoggedInUser() {
	dummyUser := user.User{
		Id:          1,
		Name:        "John Doe",
		Email:       "johndoe@example.com",
		PhoneNumber: "+1234567890",
		Password:    "hashedpassword123",
		Role:        "ADMIN",
		IsVerified:  true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	type testCaseStruct struct {
		name           string
		setup          func(ctx *context.Context)
		expectedError  error
		expectedResult user.User
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)
				s.userRepository.On("GetUserById", *ctx, mock.Anything, mock.Anything).Return(repository.User(dummyUser), nil)
			},
			expectedError:  nil,
			expectedResult: dummyUser,
		},
		{
			name:           "Failed to retrieve user id from context",
			expectedError:  apperrors.ErrInternalServer,
			expectedResult: user.User{},
		},
		{
			name: "Failed to get user",
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)
				s.userRepository.On("GetUserById", *ctx, mock.Anything, mock.Anything).Return(repository.User{}, apperrors.ErrUserNotFound)
			},
			expectedError:  apperrors.ErrUserNotFound,
			expectedResult: user.User{},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup(&ctx)
			}

			result, err := s.service.GetLoggedInUser(ctx)

			s.Equal(tt.expectedError, err)
			s.Equal(tt.expectedResult, result)
		})
		s.TearDownTest()
	}
}

func (s *UserServiceTestSuite) Test_UpgradeUserRoleToHost() {
	type testCaseStruct struct {
		name          string
		setup         func(ctx *context.Context)
		expectedError error
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)
				s.userRepository.On("UpdateUserRole", *ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "Failed to retrieve user id from context",
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name: "Failed to upgrade user to HOST",
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)
				s.userRepository.On("UpdateUserRole", *ctx, mock.Anything, mock.Anything, mock.Anything).Return(apperrors.ErrInternalServer)
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

			err := s.service.UpgradeUserRoleToHost(ctx)

			s.Equal(tt.expectedError, err)
		})
		s.TearDownTest()
	}
}

func (s *UserServiceTestSuite) Test_GetUserById() {
	dummyUser := user.User{
		Id:          1,
		Name:        "John Doe",
		Email:       "johndoe@example.com",
		PhoneNumber: "+1234567890",
		Password:    "hashedpassword123",
		Role:        "ADMIN",
		IsVerified:  true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	type testCaseStruct struct {
		name           string
		input          int
		setup          func(ctx context.Context)
		expectedError  error
		expectedResult user.User
	}

	testCases := []testCaseStruct{
		{
			name:  "Success",
			input: 1,
			setup: func(ctx context.Context) {
				s.userRepository.On("GetUserById", ctx, mock.Anything, mock.Anything).Return(repository.User(dummyUser), nil)
			},
			expectedError:  nil,
			expectedResult: dummyUser,
		},
		{
			name:  "User not found",
			input: 2,
			setup: func(ctx context.Context) {
				s.userRepository.On("GetUserById", ctx, mock.Anything, mock.Anything).Return(repository.User{}, apperrors.ErrUserNotFound)
			},
			expectedError:  apperrors.ErrUserNotFound,
			expectedResult: user.User{},
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup(ctx)
			}

			result, err := s.service.GetUserById(ctx, tt.input)

			s.Equal(tt.expectedError, err)
			s.Equal(tt.expectedResult, result)
		})
		s.TearDownTest()
	}
}

func (s *UserServiceTestSuite) Test_RefreshAccessToken() {
	type testCaseStruct struct {
		name          string
		setup         func(ctx *context.Context)
		expectedError error
	}

	testCases := []testCaseStruct{
		{
			name: "Success",
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)
				s.userRepository.On("GetUserById", *ctx, mock.Anything, mock.Anything).Return(repository.User{}, nil)
			},
			expectedError: nil,
		},
		{
			name:          "No is in context",
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name: "User not found",
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)
				s.userRepository.On("GetUserById", *ctx, mock.Anything, mock.Anything).Return(repository.User{}, apperrors.ErrUserNotFound)
			},
			expectedError: apperrors.ErrInternalServer,
		},
		{
			name: "Failed to create jwt",
			setup: func(ctx *context.Context) {
				*ctx = context.WithValue(*ctx, middleware.RequestContextUserIdKey, 1)
				s.userRepository.On("GetUserById", *ctx, mock.Anything, mock.Anything).Return(repository.User{}, nil)

				p, err := mpatch.PatchMethod(cryptokit.CreateJWTToken, func(data jwt.MapClaims) (string, error) {
					return "", apperrors.ErrInternalServer
				})
				require.NoError(s.T(), err)

				s.patches = append(s.patches, p)

			},
			expectedError: apperrors.ErrJWTCreationFailed,
		},
	}

	for _, tt := range testCases {
		s.SetupTest()
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.setup != nil {
				tt.setup(&ctx)
			}

			result, err := s.service.RefreshAccessToken(ctx)

			s.Equal(tt.expectedError, err)
			if tt.expectedError == nil {
				s.NotEmpty(result)
			}
		})
		s.TearDownTest()
	}
}
