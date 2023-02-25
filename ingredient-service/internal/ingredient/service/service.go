package service

import (
	"context"
	"errors"
	"fmt"
	apperror "github.com/tony-spark/recipetor-backend/ingredient-service/internal/errors"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient/storage"
)

type Service interface {
	Create(ctx context.Context, dto ingredient.CreateIngredientDTO) (string, error)
	GetByID(ctx context.Context, id string) (ingredient.Ingredient, error)
	SearchByName(ctx context.Context, nameQuery string) ([]ingredient.Ingredient, error)
}

type service struct {
	storage storage.Storage
}

func NewService(storage storage.Storage) Service {
	return service{
		storage: storage,
	}
}

func (s service) Create(ctx context.Context, dto ingredient.CreateIngredientDTO) (string, error) {
	ingr := ingredient.Ingredient{
		Name:           dto.Name,
		BaseUnit:       dto.BaseUnit,
		NutritionFacts: dto.NutritionFacts,
	}
	id, err := s.storage.Create(ctx, ingr)
	if err != nil {
		if errors.Is(err, apperror.ErrDuplicate) {
			return "", err
		}
		return "", fmt.Errorf("could not create ingredient: %w", err)
	}
	return id, nil
}

func (s service) GetByID(ctx context.Context, id string) (ingr ingredient.Ingredient, err error) {
	ingr, err = s.storage.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return
		}
		return ingr, fmt.Errorf("could not get ingredient: %w", err)
	}
	return
}

func (s service) SearchByName(ctx context.Context, nameQuery string) ([]ingredient.Ingredient, error) {
	ingrs, err := s.storage.SearchByName(ctx, nameQuery)
	if err != nil {
		return nil, err
	}
	return ingrs, nil
}
