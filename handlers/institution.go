package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kgermando/pdp-api/database"
	"github.com/kgermando/pdp-api/models"
	"github.com/kgermando/pdp-api/utils"
)

// GetInstitutions godoc
// GET /api/v1/institutions
func GetInstitutions(c *fiber.Ctx) error {
	p := utils.GetPagination(c)
	instType := c.Query("type")
	province := c.Query("province")
	search := c.Query("search")

	query := database.DB.Model(&models.Institution{}).Where("is_active = true")

	if instType != "" {
		query = query.Where("type = ?", instType)
	}
	if province != "" {
		query = query.Where("province = ? OR province = ''", province)
	}
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var institutions []models.Institution
	query.Order("name ASC").Limit(p.PerPage).Offset(p.Offset).Find(&institutions)

	return utils.SuccessWithMeta(c, institutions, utils.BuildMeta(total, p), "Institutions récupérées")
}

// GetInstitution godoc
// GET /api/v1/institutions/:id
func GetInstitution(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	var institution models.Institution
	if err := database.DB.
		Preload("Evaluations").
		First(&institution, "id = ? AND is_active = true", id).Error; err != nil {
		return utils.NotFound(c, "Institution non trouvée")
	}
	return utils.Success(c, institution, "Institution récupérée")
}

// CreateInstitution godoc
// POST /api/v1/institutions (admin only)
func CreateInstitution(c *fiber.Ctx) error {
	var input struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Description string `json:"description"`
		Province    string `json:"province"`
		City        string `json:"city"`
		Website     string `json:"website"`
		Phone       string `json:"phone"`
		Email       string `json:"email"`
		Address     string `json:"address"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "Données invalides")
	}
	if input.Name == "" || input.Type == "" {
		return utils.BadRequest(c, "Nom et type sont obligatoires")
	}

	institution := models.Institution{
		Name:        input.Name,
		Type:        models.InstitutionType(input.Type),
		Description: input.Description,
		Province:    input.Province,
		City:        input.City,
		Website:     input.Website,
		Phone:       input.Phone,
		Email:       input.Email,
		Address:     input.Address,
	}
	if err := database.DB.Create(&institution).Error; err != nil {
		return utils.InternalError(c, "Erreur lors de la création")
	}
	return utils.Created(c, institution, "Institution créée")
}

// EvaluateInstitution godoc
// POST /api/v1/institutions/:id/evaluate
func EvaluateInstitution(c *fiber.Ctx) error {
	var userID *uuid.UUID
	if id, ok := c.Locals("userID").(uuid.UUID); ok {
		userID = &id
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	var institution models.Institution
	if err := database.DB.First(&institution, "id = ? AND is_active = true", id).Error; err != nil {
		return utils.NotFound(c, "Institution non trouvée")
	}

	var input struct {
		Rating         int    `json:"rating"`
		Transparency   int    `json:"transparency"`
		Accountability int    `json:"accountability"`
		Responsiveness int    `json:"responsiveness"`
		Integrity      int    `json:"integrity"`
		Comment        string `json:"comment"`
		Province       string `json:"province"`
		IsAnonymous    bool   `json:"is_anonymous"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "Données invalides")
	}
	if input.Rating < 1 || input.Rating > 5 {
		return utils.BadRequest(c, "La note doit être entre 1 et 5")
	}

	eval := models.Evaluation{
		InstitutionID:  id,
		Rating:         input.Rating,
		Transparency:   input.Transparency,
		Accountability: input.Accountability,
		Responsiveness: input.Responsiveness,
		Integrity:      input.Integrity,
		Comment:        input.Comment,
		Province:       input.Province,
		IsAnonymous:    input.IsAnonymous || userID == nil,
	}
	if !input.IsAnonymous && userID != nil {
		eval.UserID = *userID
	}

	if err := database.DB.Create(&eval).Error; err != nil {
		return utils.InternalError(c, "Erreur lors de l'évaluation")
	}

	// Recalculate average rating
	var avgRating float64
	var count int64
	database.DB.Model(&models.Evaluation{}).
		Where("institution_id = ?", id).
		Select("AVG(rating) as avg, COUNT(*) as cnt").
		Row().Scan(&avgRating, &count)

	database.DB.Model(&institution).Updates(map[string]interface{}{
		"avg_rating":    avgRating,
		"total_ratings": count,
	})

	return utils.Created(c, eval, "Évaluation enregistrée")
}

// GetInstitutionEvaluations godoc
// GET /api/v1/institutions/:id/evaluations
func GetInstitutionEvaluations(c *fiber.Ctx) error {
	p := utils.GetPagination(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	query := database.DB.Model(&models.Evaluation{}).
		Where("institution_id = ?", id)

	var total int64
	query.Count(&total)

	var evaluations []models.Evaluation
	query.Order("created_at DESC").Limit(p.PerPage).Offset(p.Offset).Find(&evaluations)

	return utils.SuccessWithMeta(c, evaluations, utils.BuildMeta(total, p), "Évaluations récupérées")
}
