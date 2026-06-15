package service

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/19parwiz/agripro-core/services/farm/internal/model"
	"github.com/19parwiz/agripro-core/services/farm/internal/repository"
	apperrors "github.com/19parwiz/agripro-core/shared/errors"
)

type DeviceService struct {
	devices repository.DeviceRepository
}

func NewDeviceService(devices repository.DeviceRepository) *DeviceService {
	return &DeviceService{devices: devices}
}

type DeviceInput struct {
	Name       string
	DeviceID   string
	Type       string
	Location   string
	StreamPath string
}

type DeviceView struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	DeviceID   string    `json:"device_id"`
	Type       string    `json:"type"`
	Location   string    `json:"location"`
	StreamPath string    `json:"stream_path"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (s *DeviceService) Create(ctx context.Context, userID string, input DeviceInput) (*DeviceView, error) {
	if err := validateDeviceInput(input); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	device := model.Device{
		UserID:     userID,
		Name:       strings.TrimSpace(input.Name),
		DeviceID:   strings.TrimSpace(input.DeviceID),
		Type:       strings.TrimSpace(input.Type),
		Location:   strings.TrimSpace(input.Location),
		StreamPath: strings.TrimSpace(input.StreamPath),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.devices.Create(ctx, &device); err != nil {
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "could not create device")
	}

	view := toDeviceView(device)
	return &view, nil
}

func (s *DeviceService) List(ctx context.Context, userID string) ([]DeviceView, error) {
	devices, err := s.devices.ListByUser(ctx, userID)
	if err != nil {
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "could not list devices")
	}

	views := make([]DeviceView, 0, len(devices))
	for _, device := range devices {
		views = append(views, toDeviceView(device))
	}

	return views, nil
}

func (s *DeviceService) Get(ctx context.Context, userID, id string) (*DeviceView, error) {
	device, err := s.devices.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	view := toDeviceView(*device)
	return &view, nil
}

func (s *DeviceService) Update(ctx context.Context, userID, id string, input DeviceInput) (*DeviceView, error) {
	if err := validateDeviceInput(input); err != nil {
		return nil, err
	}

	device, err := s.devices.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	device.Name = strings.TrimSpace(input.Name)
	device.DeviceID = strings.TrimSpace(input.DeviceID)
	device.Type = strings.TrimSpace(input.Type)
	device.Location = strings.TrimSpace(input.Location)
	device.StreamPath = strings.TrimSpace(input.StreamPath)
	device.UpdatedAt = time.Now().UTC()

	if err := s.devices.Update(ctx, device); err != nil {
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "could not update device")
	}

	view := toDeviceView(*device)
	return &view, nil
}

func (s *DeviceService) Delete(ctx context.Context, userID, id string) error {
	return s.devices.Delete(ctx, userID, id)
}

func validateDeviceInput(input DeviceInput) error {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.DeviceID) == "" {
		return apperrors.ErrBadRequest
	}
	return nil
}

func toDeviceView(device model.Device) DeviceView {
	return DeviceView{
		ID:         device.ID,
		Name:       device.Name,
		DeviceID:   device.DeviceID,
		Type:       device.Type,
		Location:   device.Location,
		StreamPath: device.StreamPath,
		CreatedAt:  device.CreatedAt,
		UpdatedAt:  device.UpdatedAt,
	}
}
