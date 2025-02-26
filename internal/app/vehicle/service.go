package vehicle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/firebase"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/middleware"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository"
	"github.com/google/uuid"
)

type service struct {
	vehicleRepository repository.VehicleRepository
	firebaseService   firebase.Service
}

type Service interface {
	CreateVehicle(ctx context.Context, vehicleData VehicleRequestBody) (Vehicle, error)
	UpdateVehicle(ctx context.Context, vehicleData VehicleRequestBody, vehicleId int) (Vehicle, error)
	SoftDeleteVehicle(ctx context.Context, vehicleId int) (err error)
	GenerateSignedVehicleImageUploadURL(ctx context.Context, mimetype string) (signedUrl, accessUrl string, err error)
	GetVehicleById(ctx context.Context, vehicleId int) (vehicle Vehicle, err error)
	GetVehicles(ctx context.Context, params GetVehiclesParams) (vehicles PaginatedVehicleOverview, err error)
	GetVehiclesForHost(ctx context.Context, page, limit int) (vehicles PaginatedVehicleOverview, err error)
}

func NewService(vehicleRepository repository.VehicleRepository, firebaseService firebase.Service) Service {
	return &service{
		vehicleRepository: vehicleRepository,
		firebaseService:   firebaseService,
	}
}

func (s *service) CreateVehicle(ctx context.Context, vehicleData VehicleRequestBody) (newVehicle Vehicle, err error) {
	userId, ok := ctx.Value(middleware.RequestContextUserIdKey).(int)
	if !ok {
		slog.Error("failed to retrieve user id from context")
		return Vehicle{}, apperrors.ErrInternalServer
	}

	err = vehicleData.validate()
	if err != nil {
		slog.Error("vehicle details validation failed", "error", err)
		return Vehicle{}, apperrors.ErrInvalidRequestBody
	}

	tx, err := s.vehicleRepository.BeginTx(ctx)
	if err != nil {
		slog.Error("failed to start user creation", "error", err)
		return Vehicle{}, err
	}

	defer func() {
		if txErr := s.vehicleRepository.HandleTransaction(ctx, tx, err); txErr != nil {
			slog.Error("failed to handle transaction", "error", txErr)
			err = txErr
		}
	}()

	createVehicleData := mapVehicleRequestBodyToCreateUserRequestBodyRepo(vehicleData)
	createVehicleData.HostId = userId
	vehicle, err := s.vehicleRepository.CreateVehicle(ctx, tx, createVehicleData)
	if err != nil {
		slog.Error("failed to create new vehicle", "error", err)
		return Vehicle{}, err
	}

	var vehicleImages []repository.VehicleImage
	for _, vehicleImage := range vehicleData.Images {
		vehicleImageData := repository.CreateVehicleImageData{
			VehicleId: vehicle.Id,
			Url:       vehicleImage.Url,
			Featured:  vehicleImage.Featured,
		}
		createdVehicleImage, err := s.vehicleRepository.CreateVehicleImage(ctx, tx, vehicleImageData)
		if err != nil {
			slog.Error("failed to link image with vehicle", "error", err)
			if errors.Is(err, apperrors.ErrInvalidImageToLink) {
				return Vehicle{}, apperrors.ErrInvalidRequestBody
			}
			return Vehicle{}, err
		}
		vehicleImages = append(vehicleImages, createdVehicleImage)
	}

	return mapVehicleRepoAndVehicleImageRepoToVehicle(vehicle, vehicleImages), nil
}

func (s *service) UpdateVehicle(ctx context.Context, vehicleData VehicleRequestBody, vehicleId int) (newVehicle Vehicle, err error) {
	err = vehicleData.validate()
	if err != nil {
		slog.Error("vehicle details validation failed", "error", err)
		return Vehicle{}, apperrors.ErrInvalidRequestBody
	}

	tx, err := s.vehicleRepository.BeginTx(ctx)
	if err != nil {
		slog.Error("failed to start user updating", "error", err)
		return Vehicle{}, err
	}

	defer func() {
		if txErr := s.vehicleRepository.HandleTransaction(ctx, tx, err); txErr != nil {
			slog.Error("failed to handle transaction", "error", txErr)
			err = txErr
		}
	}()

	editVehicleData := mapVehicleRequestBodyToEditUserRequestBodyRepo(vehicleData)
	editVehicleData.Id = vehicleId
	vehicle, err := s.vehicleRepository.UpdateVehicle(ctx, tx, editVehicleData)
	if err != nil {
		slog.Error("failed to update vehicle", "error", err)
		return Vehicle{}, err
	}

	err = s.vehicleRepository.DeleteAllImagesForVehicle(ctx, tx, vehicleId)
	if err != nil {
		slog.Error("failed to delete images for vehicle", "error", err)
		return Vehicle{}, err
	}

	var vehicleImages []repository.VehicleImage
	for _, vehicleImage := range vehicleData.Images {
		vehicleImageData := repository.CreateVehicleImageData{
			VehicleId: vehicle.Id,
			Url:       vehicleImage.Url,
			Featured:  vehicleImage.Featured,
		}
		createdVehicleImage, err := s.vehicleRepository.CreateVehicleImage(ctx, tx, vehicleImageData)
		if err != nil {
			slog.Error("failed to link image with vehicle", "error", err)
			if errors.Is(err, apperrors.ErrInvalidImageToLink) {
				return Vehicle{}, apperrors.ErrInvalidRequestBody
			}
			return Vehicle{}, err
		}
		vehicleImages = append(vehicleImages, createdVehicleImage)
	}

	return mapVehicleRepoAndVehicleImageRepoToVehicle(vehicle, vehicleImages), nil
}

func (s *service) SoftDeleteVehicle(ctx context.Context, vehicleId int) (err error) {
	err = s.vehicleRepository.SoftDeleteVehicle(ctx, nil, vehicleId)
	if err != nil {
		slog.Error("failed to soft delete vehicle", "error", err)
		return err
	}

	return nil
}

func (s *service) GenerateSignedVehicleImageUploadURL(ctx context.Context, mimetype string) (signedUrl, accessUrl string, err error) {
	timestamp := time.Now().UnixNano()
	randomStr := uuid.New().String()

	objectPath := fmt.Sprintf("vehicles/%d-%s",
		timestamp,
		randomStr,
	)

	if mimetype == "" {
		mimetype = "image/jpeg"
	}

	signedUrl, err = s.firebaseService.GenerateSignedURL(ctx, objectPath, mimetype, SignedURLExpiry)
	if err != nil {
		slog.Error("failed to generate signed url for vehicle image upload", "error", err)
		return "", "", err
	}

	accessUrl = fmt.Sprintf(AccessURLFormat, fmt.Sprintf("vehicles%%2F%d-%s", timestamp, randomStr))

	return signedUrl, accessUrl, nil
}

func (s *service) GetVehicleById(ctx context.Context, vehicleId int) (vehicle Vehicle, err error) {
	vehicleDetails, err := s.vehicleRepository.GetVehicleById(ctx, nil, vehicleId)
	if err != nil {
		slog.Error("failed to get vehicle details", "error", err)
		return Vehicle{}, err
	}

	vehicleImages, err := s.vehicleRepository.GetVehicleImagesByVehicleId(ctx, nil, vehicleId)
	if err != nil {
		slog.Error("failed to get vehicle images", "error", err)
		return Vehicle{}, err
	}

	return mapVehicleRepoAndVehicleImageRepoToVehicle(vehicleDetails, vehicleImages), nil
}

func (s *service) GetVehicles(ctx context.Context, params GetVehiclesParams) (vehicles PaginatedVehicleOverview, err error) {
	if params.PickupTimestamp.After(params.DropoffTimestamp) {
		slog.Error("pickup time after the dropoff time", "error", err)
		return PaginatedVehicleOverview{}, apperrors.ErrInvalidPickupDropoff
	}

	if params.Page <= 0 {
		slog.Error("invalid page number provided", "page", params.Page)
		return PaginatedVehicleOverview{}, apperrors.ErrInvalidPagination
	}

	if params.Limit <= 0 {
		slog.Error("invalid limit value provided", "limit", params.Limit)
		return PaginatedVehicleOverview{}, apperrors.ErrInvalidPagination
	}

	offset := params.Limit * (params.Page - 1)

	repoParams := repository.GetVehiclesParams{
		City:             params.City,
		PickupTimestamp:  params.PickupTimestamp,
		DropoffTimestamp: params.DropoffTimestamp,
		Offset:           offset,
		Limit:            params.Limit,
	}
	vehicleList, totalVehicles, err := s.vehicleRepository.GetVehicles(ctx, nil, repoParams)
	if err != nil {
		slog.Error("failed to get vehicle list", "error", err)
		return PaginatedVehicleOverview{}, err
	}

	vehicleData := make([]VehicleOverview, len(vehicleList))
	for i, v := range vehicleList {
		vehicleData[i] = VehicleOverview(v)
	}

	return PaginatedVehicleOverview{
		Data: vehicleData,
		Pagination: PaginationParams{
			Page:       params.Page,
			PageSize:   params.Limit,
			TotalCount: totalVehicles,
		}}, nil
}

func (s *service) GetVehiclesForHost(ctx context.Context, page, limit int) (vehicles PaginatedVehicleOverview, err error) {
	userId, ok := ctx.Value(middleware.RequestContextUserIdKey).(int)

	if !ok {
		slog.Error("failed to retrieve user id from context")
		return PaginatedVehicleOverview{}, apperrors.ErrInternalServer
	}

	if page <= 0 {
		slog.Error("invalid page number provided", "page", page)
		return PaginatedVehicleOverview{}, apperrors.ErrInvalidPagination
	}

	if limit <= 0 {
		slog.Error("invalid limit value provided", "limit", limit)
		return PaginatedVehicleOverview{}, apperrors.ErrInvalidPagination
	}

	offset := limit * (page - 1)

	repoParams := repository.GetVehiclesForHostParams{
		HostId: userId,
		Offset: offset,
		Limit:  limit,
	}
	vehicleList, totalVehicles, err := s.vehicleRepository.GetVehiclesForHost(ctx, nil, repoParams)
	if err != nil {
		slog.Error("failed to get vehicle list for host", "error", err)
		return PaginatedVehicleOverview{}, err
	}

	vehicleData := make([]VehicleOverview, len(vehicleList))
	for i, v := range vehicleList {
		vehicleData[i] = VehicleOverview(v)
	}

	return PaginatedVehicleOverview{
		Data: vehicleData,
		Pagination: PaginationParams{
			Page:       page,
			PageSize:   limit,
			TotalCount: totalVehicles,
		}}, nil
}
