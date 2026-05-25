package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kgermando/pdp-api/database"
	"github.com/kgermando/pdp-api/models"
	"github.com/kgermando/pdp-api/utils"
)

// GetReports godoc
// GET /api/v1/reports
func GetReports(c *fiber.Ctx) error {
	p := utils.GetPagination(c)
	province := c.Query("province")
	status := c.Query("status")
	categoryID := c.Query("category_id")
	severity := c.Query("severity")
	search := c.Query("search")

	query := database.DB.Model(&models.Report{}).
		Preload("Category").
		Preload("Evidences")

	if province != "" {
		query = query.Where("province = ?", province)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var reports []models.Report
	query.Order("created_at DESC").
		Limit(p.PerPage).
		Offset(p.Offset).
		Find(&reports)

	return utils.SuccessWithMeta(c, reports, utils.BuildMeta(total, p), "Signalements récupérés")
}

// GetReport godoc
// GET /api/v1/reports/:id
func GetReport(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	var report models.Report
	if err := database.DB.
		Preload("Category").
		Preload("Evidences").
		First(&report, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "Signalement non trouvé")
	}

	if !report.IsAnonymous {
		database.DB.Preload("Reporter").First(&report, "id = ?", id)
	}

	return utils.Success(c, report, "Signalement récupéré")
}

// GetMapReports godoc
// GET /api/v1/reports/map — returns geolocated active reports
func GetMapReports(c *fiber.Ctx) error {
	province := c.Query("province")

	query := database.DB.Model(&models.Report{}).
		Select("id, title, severity, status, latitude, longitude, province, city, category_id, occurred_at, created_at").
		Preload("Category").
		Where("latitude != 0 AND longitude != 0")

	if province != "" {
		query = query.Where("province = ?", province)
	}

	var reports []models.Report
	query.Order("created_at DESC").Limit(500).Find(&reports)
	return utils.Success(c, reports, "Carte des signalements")
}

// CreateReport godoc
// POST /api/v1/reports
func CreateReport(c *fiber.Ctx) error {
	var reporterID *uuid.UUID

	if id, ok := c.Locals("userID").(uuid.UUID); ok {
		reporterID = &id
	}

	var input struct {
		Title        string  `json:"title"`
		Description  string  `json:"description"`
		CategoryID   string  `json:"category_id"`
		IsAnonymous  bool    `json:"is_anonymous"`
		Severity     string  `json:"severity"`
		Latitude     float64 `json:"latitude"`
		Longitude    float64 `json:"longitude"`
		Province     string  `json:"province"`
		City         string  `json:"city"`
		Address      string  `json:"address"`
		OccurredAt   string  `json:"occurred_at"`
		VictimsCount int     `json:"victims_count"`
		Tags         string  `json:"tags"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "Données invalides")
	}

	if input.Title == "" || input.Description == "" || input.Province == "" || input.CategoryID == "" {
		return utils.BadRequest(c, "Titre, description, province et catégorie sont obligatoires")
	}

	catID, err := uuid.Parse(input.CategoryID)
	if err != nil {
		return utils.BadRequest(c, "ID de catégorie invalide")
	}

	severity := models.ReportSeverity(input.Severity)
	if severity == "" {
		severity = models.SeverityMedium
	}

	occurredAt := time.Now()
	if input.OccurredAt != "" {
		if t, err := time.Parse(time.RFC3339, input.OccurredAt); err == nil {
			occurredAt = t
		}
	}

	report := models.Report{
		Title:        input.Title,
		Description:  input.Description,
		CategoryID:   catID,
		IsAnonymous:  input.IsAnonymous || reporterID == nil,
		Severity:     severity,
		Latitude:     input.Latitude,
		Longitude:    input.Longitude,
		Province:     input.Province,
		City:         input.City,
		Address:      input.Address,
		OccurredAt:   occurredAt,
		VictimsCount: input.VictimsCount,
		Tags:         input.Tags,
	}

	if !input.IsAnonymous && reporterID != nil {
		report.ReporterID = *reporterID
	}

	if err := database.DB.Create(&report).Error; err != nil {
		return utils.InternalError(c, "Erreur lors du signalement")
	}

	database.DB.Preload("Category").First(&report, "id = ?", report.ID)
	return utils.Created(c, report, "Signalement créé avec succès")
}

// UpdateReportStatus godoc
// PATCH /api/v1/reports/:id/status  (moderator/admin)
func UpdateReportStatus(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	var report models.Report
	if err := database.DB.First(&report, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "Signalement non trouvé")
	}

	var input struct {
		Status        string `json:"status"`
		ModeratorNote string `json:"moderator_note"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "Données invalides")
	}

	updates := map[string]interface{}{
		"status": input.Status,
	}
	if input.ModeratorNote != "" {
		updates["moderator_note"] = input.ModeratorNote
	}
	database.DB.Model(&report).Updates(updates)
	return utils.Success(c, report, "Statut mis à jour")
}

// GetReportStats godoc
// GET /api/v1/reports/stats
func GetReportStats(c *fiber.Ctx) error {
	province := c.Query("province")

	query := database.DB.Model(&models.Report{})
	if province != "" {
		query = query.Where("province = ?", province)
	}

	var total int64
	var pending, verified, resolved int64
	var critical, high int64

	query.Count(&total)
	query.Where("status = ?", models.ReportPending).Count(&pending)
	query.Where("status = ?", models.ReportVerified).Count(&verified)
	query.Where("status = ?", models.ReportResolved).Count(&resolved)
	query.Where("severity = ?", models.SeverityCritical).Count(&critical)
	query.Where("severity = ?", models.SeverityHigh).Count(&high)

	return utils.Success(c, fiber.Map{
		"total":    total,
		"pending":  pending,
		"verified": verified,
		"resolved": resolved,
		"critical": critical,
		"high":     high,
	}, "Statistiques des signalements")
}
