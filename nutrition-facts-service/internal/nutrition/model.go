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
	IngredientID string  `json:"ingredient_id" bson:"ingredient_id"`
	Unit         string  `json:"unit" bson:"unit"`
	Amount       float64 `json:"amount" bson:"amount"`
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

type Recipe struct {
	ID             string             `json:"id" bson:"_id,omitempty"`
	Name           string             `json:"name" bson:"name,omitempty"`
	CreatedBy      string             `json:"created_by" bson:"created_by,omitempty"`
	Ingredients    []RecipeIngredient `json:"ingredients,omitempty" bson:"ingredients,omitempty"`
	Steps          []Step             `json:"steps,omitempty" bson:"steps,omitempty"`
	NutritionFacts *NutritionFacts    `json:"nutrition_facts,omitempty" bson:"nutrition_facts,omitempty"`
}

type Step struct {
	Description string `json:"description" bson:"description"`
}

type FindIngredientsDTO struct {
	ID        string `json:"ingredient_id,omitempty"`
	NameQuery string `json:"name_query,omitempty"`
}

type IngredientDTO struct {
	Ingredient Ingredient `json:"ingredient,omitempty"`
	Name       string     `json:"name,omitempty"`
	ID         string     `json:"ingredient_id,omitempty"`
	NameQuery  string     `json:"name_query,omitempty"`
	Error      string     `json:"error,omitempty"`
}
