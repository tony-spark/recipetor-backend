package ingredient

type Ingredient struct {
	ID             string          `json:"id" bson:"_id,omitempty"`
	Name           string          `json:"name" bson:"name,omitempty"`
	BaseUnit       string          `json:"base_unit" bson:"base_unit,omitempty"`
	NutritionFacts *NutritionFacts `json:"nutrition_facts,omitempty" bson:"nutrition_facts,omitempty"`
}

type NutritionFacts struct {
	Calories      float64 `json:"calories" bson:"calories,omitempty"`
	Proteins      float64 `json:"proteins" bson:"proteins,omitempty"`
	Fats          float64 `json:"fats" bson:"fats,omitempty"`
	Carbohydrates float64 `json:"carbohydrates" bson:"carbohydrates,omitempty"`
}

type CreateIngredientDTO struct {
	Name           string          `json:"name"`
	BaseUnit       string          `json:"base_unit"`
	NutritionFacts *NutritionFacts `json:"nutrition_facts,omitempty"`
}

type IngredientDTO struct {
	Ingredient Ingredient `json:"ingredient,omitempty"`
	Name       string     `json:"name,omitempty"`
	ID         string     `json:"ingredient_id,omitempty"`
	NameQuery  string     `json:"name_query,omitempty"`
	Error      string     `json:"error,omitempty"`
}

type FindIngredientsDTO struct {
	ID        string `json:"ingredient_id,omitempty"`
	NameQuery string `json:"name_query,omitempty"`
}
