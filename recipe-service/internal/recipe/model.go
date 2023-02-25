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
	ID          string       `json:"id" bson:"_id,omitempty"`
	Name        string       `json:"name" bson:"name,omitempty"`
	CreatedBy   string       `json:"created_by" bson:"created_by,omitempty"`
	Ingredients []Ingredient `json:"ingredients,omitempty" bson:"ingredients,omitempty"`
	Steps       []Step       `json:"steps,omitempty" bson:"steps,omitempty"`
}

type CreateRecipeDTO struct {
	Name        string       `json:"name"`
	CreatedBy   string       `json:"created_by"`
	Ingredients []Ingredient `json:"ingredients,omitempty"`
	Steps       []Step       `json:"steps,omitempty"`
}

type UpdateRecipeDTO struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Ingredients []Ingredient `json:"ingredients,omitempty"`
	Steps       []Step       `json:"steps,omitempty"`
}
