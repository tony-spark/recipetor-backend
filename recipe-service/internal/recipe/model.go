package recipe

type Step struct {
	Description string `json:"description" bson:"description"`
}

type Ingredient struct {
	IngredientID string  `json:"ingredient_id" bson:"ingredient_id"`
	Unit         string  `json:"unit" bson:"unit"`
	Amount       float64 `json:"amount" bson:"amount"`
}

type Recipe struct {
	ID             string          `json:"id" bson:"_id,omitempty"`
	Name           string          `json:"name" bson:"name,omitempty"`
	CreatedBy      string          `json:"created_by" bson:"created_by,omitempty"`
	Ingredients    []Ingredient    `json:"ingredients,omitempty" bson:"ingredients,omitempty"`
	Steps          []Step          `json:"steps,omitempty" bson:"steps,omitempty"`
	NutritionFacts *NutritionFacts `json:"nutrition_facts,omitempty" bson:"nutrition_facts,omitempty"`
}

type CreateRecipeDTO struct {
	Name        string       `json:"name"`
	CreatedBy   string       `json:"created_by"`
	Ingredients []Ingredient `json:"ingredients,omitempty"`
	Steps       []Step       `json:"steps,omitempty"`
}

type UpdateRecipeDTO struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Ingredients    []Ingredient    `json:"ingredients,omitempty"`
	Steps          []Step          `json:"steps,omitempty"`
	NutritionFacts *NutritionFacts `json:"nutrition_facts,omitempty"`
}

type FindRecipeDTO struct {
	ID            string   `json:"recipe_id"`
	UserID        string   `json:"user_id"`
	IngredientIDs []string `json:"ingredient_ids"`
}

type RecipeDTO struct {
	Recipe Recipe `json:"recipe,omitempty"`
	ID     string `json:"recipe_id,omitempty"`
	UserID string `json:"user_id,omitempty"`
	Error  string `json:"error,omitempty"`
}

type RecipeNutritionsDTO struct {
	RecipeID       string         `json:"recipe_id"`
	NutritionFacts NutritionFacts `json:"nutrition_facts"`
	Inaccurate     bool           `json:"is_inaccurate"`
}

type NutritionFacts struct {
	Calories      float64 `json:"calories" bson:"calories"`
	Proteins      float64 `json:"proteins" bson:"proteins"`
	Fats          float64 `json:"fats" bson:"fats"`
	Carbohydrates float64 `json:"carbohydrates" bson:"carbohydrates"`
}
