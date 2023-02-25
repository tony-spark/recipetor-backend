package nutrition

type Ingredient struct {
	ID             string          `json:"id"`
	BaseUnit       string          `json:"base_unit"`
	NutritionFacts *NutritionFacts `json:"nutrition_facts"`
}

type NutritionFacts struct {
	Calories      float64 `json:"calories"`
	Proteins      float64 `json:"proteins"`
	Fats          float64 `json:"fats"`
	Carbohydrates float64 `json:"carbohydrates"`
}

type RecipeIngredient struct {
	Ingredient Ingredient `json:"ingredient"`
	Unit       string     `json:"unit"`
	Amount     float64    `json:"amount"`
}

type RecipeDTO struct {
	RecipeID    string             `json:"recipe_id"`
	Ingredients []RecipeIngredient `json:"ingredients"`
}

type RecipeNutritionsDTO struct {
	RecipeID       string         `json:"recipe_id"`
	NutritionFacts NutritionFacts `json:"nutrition_facts"`
	Inaccurate     bool           `json:"is_inaccurate"`
}
