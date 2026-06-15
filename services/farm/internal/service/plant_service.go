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

type PlantService struct {
	plants repository.PlantRepository
}

func NewPlantService(plants repository.PlantRepository) *PlantService {
	return &PlantService{plants: plants}
}

type PlantInput struct {
	Name         string
	Variety      string
	PlantingDate time.Time
}

type PlantView struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Variety      string    `json:"variety"`
	PlantingDate time.Time `json:"planting_date"`
	GrowingDays  int       `json:"growing_days"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (s *PlantService) Create(ctx context.Context, userID string, input PlantInput) (*PlantView, error) {
	if err := validatePlantInput(input); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	plant := model.Plant{
		UserID:       userID,
		Name:         strings.TrimSpace(input.Name),
		Variety:      strings.TrimSpace(input.Variety),
		PlantingDate: input.PlantingDate,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.plants.Create(ctx, &plant); err != nil {
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "could not create plant")
	}

	view := toPlantView(plant)
	return &view, nil
}

func (s *PlantService) List(ctx context.Context, userID string) ([]PlantView, error) {
	plants, err := s.plants.ListByUser(ctx, userID)
	if err != nil {
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "could not list plants")
	}

	views := make([]PlantView, 0, len(plants))
	for _, plant := range plants {
		views = append(views, toPlantView(plant))
	}

	return views, nil
}

func (s *PlantService) Get(ctx context.Context, userID, id string) (*PlantView, error) {
	plant, err := s.plants.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	view := toPlantView(*plant)
	return &view, nil
}

func (s *PlantService) Update(ctx context.Context, userID, id string, input PlantInput) (*PlantView, error) {
	if err := validatePlantInput(input); err != nil {
		return nil, err
	}

	plant, err := s.plants.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	plant.Name = strings.TrimSpace(input.Name)
	plant.Variety = strings.TrimSpace(input.Variety)
	plant.PlantingDate = input.PlantingDate
	plant.UpdatedAt = time.Now().UTC()

	if err := s.plants.Update(ctx, plant); err != nil {
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "could not update plant")
	}

	view := toPlantView(*plant)
	return &view, nil
}

func (s *PlantService) Delete(ctx context.Context, userID, id string) error {
	return s.plants.Delete(ctx, userID, id)
}

func validatePlantInput(input PlantInput) error {
	if strings.TrimSpace(input.Name) == "" || input.PlantingDate.IsZero() {
		return apperrors.ErrBadRequest
	}
	return nil
}

func toPlantView(plant model.Plant) PlantView {
	plantingDate := plant.PlantingDate.UTC().Truncate(24 * time.Hour)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	growingDays := int(today.Sub(plantingDate).Hours() / 24)
	if growingDays < 0 {
		growingDays = 0
	}

	return PlantView{
		ID:           plant.ID,
		Name:         plant.Name,
		Variety:      plant.Variety,
		PlantingDate: plant.PlantingDate,
		GrowingDays:  growingDays,
		CreatedAt:    plant.CreatedAt,
		UpdatedAt:    plant.UpdatedAt,
	}
}
