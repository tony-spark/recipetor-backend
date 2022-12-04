package ingredient

import "context"

type Service interface {
	Create(ctx context.Context) (string, error)
	GetById(ctx context.Context, id string) (Ingredient, error)
	SearchByName(ctx context.Context, nameQuery string) ([]Ingredient, error)
}

type CreateIngredientDTO struct {
	Name           string          `json:"name" bson:"name"`
	NutritionFacts *NutritionFacts `json:"nutrition_facts,omitempty" bson:"nutrition_facts,omitempty"`
}
