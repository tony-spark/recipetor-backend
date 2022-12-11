package ingredient

type Ingredient struct {
	ID             string          `json:"id" bson:"_id"`
	Name           string          `json:"name" bson:"name"`
	BaseUnit       string          `json:"base_unit" bson:"base_unit"`
	NutritionFacts *NutritionFacts `json:"nutrition_facts,omitempty" bson:"nutrition_facts,omitempty"`
}

type NutritionFacts struct {
	Calories      float64 `json:"calories" bson:"calories"`
	Proteins      float64 `json:"proteins" bson:"proteins"`
	Fats          float64 `json:"fats" bson:"fats"`
	Carbohydrates float64 `json:"carbohydrates" bson:"carbohydrates"`
}

type CreateIngredientDTO struct {
	Name           string          `json:"name" bson:"name"`
	NutritionFacts *NutritionFacts `json:"nutrition_facts,omitempty" bson:"nutrition_facts,omitempty"`
}
