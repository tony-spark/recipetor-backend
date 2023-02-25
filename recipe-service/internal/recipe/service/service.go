package service

import (
	"context"
	"errors"
	"fmt"

	apperror "github.com/tony-spark/recipetor-backend/recipe-service/internal/errors"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/recipe"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/recipe/storage"
)

type Service interface {
	Create(ctx context.Context, dto recipe.CreateRecipeDTO) (string, error)
	Update(ctx context.Context, dto recipe.UpdateRecipeDTO) error
	GetByID(ctx context.Context, id string) (recipe.Recipe, error)
	GetAllByUser(ctx context.Context, userID string) ([]recipe.Recipe, error)
	GetAll(ctx context.Context, start int64, limit int64) ([]recipe.Recipe, error)
	FindByIngredients(ctx context.Context, ingredientIDs []string) ([]recipe.Recipe, error)
}

type service struct {
	storage storage.Storage
}

func NewService(storage storage.Storage) Service {
	return service{
		storage: storage,
	}
}

func (s service) Create(ctx context.Context, dto recipe.CreateRecipeDTO) (string, error) {
	r := recipe.Recipe{
		Name:        dto.Name,
		CreatedBy:   dto.CreatedBy,
		Ingredients: dto.Ingredients,
		Steps:       dto.Steps,
	}

	id, err := s.storage.Create(ctx, r)
	if err != nil {
		return "", fmt.Errorf("could not create recipe: %w", err)
	}
	return id, nil
}

func (s service) Update(ctx context.Context, dto recipe.UpdateRecipeDTO) error {
	r := recipe.Recipe{
		ID:          dto.ID,
		Name:        dto.Name,
		Ingredients: dto.Ingredients,
		Steps:       dto.Steps,
	}
	err := s.storage.Update(ctx, r)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return err
		}
		return fmt.Errorf("could not update recipe: %w", err)
	}
	return nil
}

func (s service) GetByID(ctx context.Context, id string) (r recipe.Recipe, err error) {
	r, err = s.storage.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return
		}
		return r, fmt.Errorf("could not get recipe: %w", err)
	}
	return
}

func (s service) GetAllByUser(ctx context.Context, userID string) (rs []recipe.Recipe, err error) {
	rs, err = s.storage.GetAllByUser(ctx, userID)
	return
}

func (s service) GetAll(ctx context.Context, start int64, limit int64) (rs []recipe.Recipe, err error) {
	rs, err = s.storage.GetAll(ctx, start, limit)
	return
}

func (s service) FindByIngredients(ctx context.Context, ingredientIDs []string) ([]recipe.Recipe, error) {
	// TODO implement me
	panic("implement me")
}
