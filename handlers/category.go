package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kgermando/pdp-api/database"
	"github.com/kgermando/pdp-api/models"
	"github.com/kgermando/pdp-api/utils"
)

// GetCategories godoc
// GET /api/v1/categories
func GetCategories(c *fiber.Ctx) error {
	catType := c.Query("type")

	query := database.DB.Model(&models.Category{}).Where("is_active = true")
	if catType != "" {
		query = query.Where("type = ?", catType)
	}

	var categories []models.Category
	query.Order("name ASC").Find(&categories)
	return utils.Success(c, categories, "Catégories récupérées")
}

// CreateCategory godoc
// POST /api/v1/categories (admin only)
func CreateCategory(c *fiber.Ctx) error {
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		Color       string `json:"color"`
		Type        string `json:"type"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "Données invalides")
	}
	if input.Name == "" || input.Type == "" {
		return utils.BadRequest(c, "Nom et type sont obligatoires")
	}

	category := models.Category{
		Name:        input.Name,
		Description: input.Description,
		Icon:        input.Icon,
		Color:       input.Color,
		Type:        models.CategoryType(input.Type),
	}
	if err := database.DB.Create(&category).Error; err != nil {
		return utils.InternalError(c, "Erreur lors de la création")
	}
	return utils.Created(c, category, "Catégorie créée")
}

// UpdateCategory godoc
// PUT /api/v1/categories/:id (admin only)
func UpdateCategory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	var category models.Category
	if err := database.DB.First(&category, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "Catégorie non trouvée")
	}

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		Color       string `json:"color"`
		IsActive    *bool  `json:"is_active"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "Données invalides")
	}

	updates := map[string]interface{}{}
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Description != "" {
		updates["description"] = input.Description
	}
	if input.Icon != "" {
		updates["icon"] = input.Icon
	}
	if input.Color != "" {
		updates["color"] = input.Color
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}

	database.DB.Model(&category).Updates(updates)
	return utils.Success(c, category, "Catégorie mise à jour")
}
