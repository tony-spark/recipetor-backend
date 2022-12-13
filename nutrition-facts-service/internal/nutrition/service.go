package nutrition

type Service interface {
	CalcRecipeNutritions(recipe RecipeDTO) RecipeNutritionsDTO
}
