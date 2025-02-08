package vehicle

import (
	"context"
	"errors"
	"log/slog"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/middleware"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository"
)

type service struct {
	vehicleRepository repository.VehicleRepository
}

type Service interface {
	CreateVehicle(ctx context.Context, vehicleData VehicleWithImages) (VehicleWithImages, error)
	UpdateVehicle(ctx context.Context, vehicleData VehicleWithImages, vehicleId int) (VehicleWithImages, error)
	SoftDeleteVehicle(ctx context.Context, vehicleId int) error
}

func NewService(vehicleRepository repository.VehicleRepository) Service {
	return &service{
		vehicleRepository: vehicleRepository,
	}
}

func (s *service) CreateVehicle(ctx context.Context, vehicleData VehicleWithImages) (newVehicle VehicleWithImages, err error) {
	userId, ok := ctx.Value(middleware.RequestContextUserIdKey).(int)
	if !ok {
		slog.Error("failed to retrieve user id from context")
		return newVehicle, apperrors.ErrInternalServer
	}

	err = vehicleData.validate()
	if err != nil {
		slog.Error("vehicle details validation failed", "error", err)
		return newVehicle, apperrors.ErrInvalidRequestBody
	}

	tx, err := s.vehicleRepository.BeginTx(ctx)
	if err != nil {
		slog.Error("failed to start user creation", "error", err.Error())
		return newVehicle, err
	}

	defer func() {
		if txErr := s.vehicleRepository.HandleTransaction(ctx, tx, err); txErr != nil {
			slog.Error("failed to handle transaction", "error", txErr)
		}
	}()

	vehicle, err := s.vehicleRepository.CreateVehicle(ctx, tx, MapVehicleWithImagesToVehicleRepo(vehicleData), userId)
	if err != nil {
		slog.Error("failed to create new vehicle", "error", err)
		return newVehicle, err
	}

	var vehicleImages []repository.VehicleImage
	for _, vehicleImage := range vehicleData.Images {
		updatedVehicleImage, err := s.vehicleRepository.LinkVehicleImage(ctx, tx, vehicleImage.Id, vehicle.Id)
		if err != nil {
			slog.Error("failed to link image with vehicle", "error", err)
			if errors.Is(err, apperrors.ErrInvalidImageToLink) {
				return newVehicle, apperrors.ErrInvalidRequestBody
			}
			return newVehicle, err
		}
		vehicleImages = append(vehicleImages, updatedVehicleImage)
	}

	return MapVehicleRepoAndVehicleImageRepoToVehicleWithImages(vehicle, vehicleImages), nil
}

func (s *service) UpdateVehicle(ctx context.Context, vehicleData VehicleWithImages, vehicleId int) (newVehicle VehicleWithImages, err error) {
	err = vehicleData.validate()
	if err != nil {
		slog.Error("vehicle details validation failed", "error", err)
		return newVehicle, apperrors.ErrInvalidRequestBody
	}

	tx, err := s.vehicleRepository.BeginTx(ctx)
	if err != nil {
		slog.Error("failed to start user creation", "error", err.Error())
		return newVehicle, err
	}

	defer func() {
		if txErr := s.vehicleRepository.HandleTransaction(ctx, tx, err); txErr != nil {
			slog.Error("failed to handle transaction", "error", txErr)
		}
	}()

	vehicle, err := s.vehicleRepository.UpdateVehicle(ctx, tx, MapVehicleWithImagesToVehicleRepo(vehicleData), vehicleId)
	if err != nil {
		slog.Error("failed to update vehicle", "error", err)
		return newVehicle, err
	}

	err = s.vehicleRepository.DeleteAllImagesForVehicle(ctx, tx, vehicleId)
	if err != nil {
		slog.Error("failed to delete images for vehicle", "error", err)
		return newVehicle, err
	}

	var vehicleImages []repository.VehicleImage
	for _, vehicleImage := range vehicleData.Images {
		updatedVehicleImage, err := s.vehicleRepository.LinkVehicleImage(ctx, tx, vehicleImage.Id, vehicle.Id)
		if err != nil {
			slog.Error("failed to link image with vehicle", "error", err)
			if errors.Is(err, apperrors.ErrInvalidImageToLink) {
				return newVehicle, apperrors.ErrInvalidRequestBody
			}
			return newVehicle, err
		}
		vehicleImages = append(vehicleImages, updatedVehicleImage)
	}

	return MapVehicleRepoAndVehicleImageRepoToVehicleWithImages(vehicle, vehicleImages), nil
}

func (s *service) SoftDeleteVehicle(ctx context.Context, vehicleId int) error {
	err := s.vehicleRepository.SoftDeleteVehicle(ctx, nil, vehicleId)
	if err != nil {
		slog.Error("failed to soft delete vehicle", "error", err)
		return err
	}

	return nil
}
