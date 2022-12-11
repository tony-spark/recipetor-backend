package recipe

type Recipe struct {
	ID          string `json:"id" bson:"_id"`
	Name        string `json:"name" bson:"name"`
	CreatedBy   string `json:"created_by" bson:"created_by"`
	Ingredients []struct {
		IngredientID string  `json:"ingredient" bson:"ingredientId"`
		Unit         string  `json:"unit" bson:"unit"`
		Amount       float64 `json:"amount" bson:"amount"`
	} `json:"ingredients,omitempty" bson:"ingredients"`
	Steps []struct {
		Description string `json:"description" bson:"description"`
	} `json:"steps,omitempty" bson:"steps"`
}

type CreateRecipeDTO struct {
	Name        string `json:"name" bson:"name"`
	CreatedBy   string `json:"created_by" bson:"created_by"`
	Ingredients []struct {
		IngredientID string  `json:"ingredient" bson:"ingredientId"`
		Unit         string  `json:"unit" bson:"unit"`
		Amount       float64 `json:"amount" bson:"amount"`
	} `json:"ingredients,omitempty" bson:"ingredients"`
	Steps []struct {
		Description string `json:"description" bson:"description"`
	} `json:"steps,omitempty" bson:"steps"`
}

type UpdateRecipeDTO struct {
	ID          string `json:"id" bson:"_id"`
	Name        string `json:"name" bson:"name"`
	Ingredients []struct {
		IngredientID string  `json:"ingredient" bson:"ingredientId"`
		Unit         string  `json:"unit" bson:"unit"`
		Amount       float64 `json:"amount" bson:"amount"`
	} `json:"ingredients,omitempty" bson:"ingredients"`
	Steps []struct {
		Description string `json:"description" bson:"description"`
	} `json:"steps,omitempty" bson:"steps"`
}
