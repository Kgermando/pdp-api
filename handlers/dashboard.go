package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/pdp-api/database"
	"github.com/kgermando/pdp-api/models"
	"github.com/kgermando/pdp-api/utils"
)

// GetDashboardStats godoc
// GET /api/v1/dashboard
func GetDashboardStats(c *fiber.Ctx) error {
	province := c.Query("province")

	// Users
	var totalUsers, newUsers int64
	database.DB.Model(&models.User{}).Count(&totalUsers)
	database.DB.Model(&models.User{}).Where("created_at >= NOW() - INTERVAL '30 days'").Count(&newUsers)

	// Proposals
	var totalProposals, publishedProposals, submittedProposals int64
	propQuery := database.DB.Model(&models.Proposal{})
	if province != "" {
		propQuery = propQuery.Where("province = ?", province)
	}
	propQuery.Count(&totalProposals)
	database.DB.Model(&models.Proposal{}).Where("status = ?", models.ProposalPublished).Count(&publishedProposals)
	database.DB.Model(&models.Proposal{}).Where("status = ?", models.ProposalSubmitted).Count(&submittedProposals)

	// Reports
	var totalReports, pendingReports, resolvedReports int64
	reportQuery := database.DB.Model(&models.Report{})
	if province != "" {
		reportQuery = reportQuery.Where("province = ?", province)
	}
	reportQuery.Count(&totalReports)
	database.DB.Model(&models.Report{}).Where("status = ?", models.ReportPending).Count(&pendingReports)
	database.DB.Model(&models.Report{}).Where("status = ?", models.ReportResolved).Count(&resolvedReports)

	// Evaluations
	var totalEvaluations int64
	database.DB.Model(&models.Evaluation{}).Count(&totalEvaluations)

	// Reports by province (top 10)
	type ProvinceCount struct {
		Province string `json:"province"`
		Count    int64  `json:"count"`
	}
	var reportsByProvince []ProvinceCount
	database.DB.Model(&models.Report{}).
		Select("province, COUNT(*) as count").
		Group("province").
		Order("count DESC").
		Limit(10).
		Scan(&reportsByProvince)

	// Recent proposals
	var recentProposals []models.Proposal
	database.DB.Preload("Category").Preload("Author").
		Where("status = ?", models.ProposalPublished).
		Order("created_at DESC").
		Limit(5).
		Find(&recentProposals)

	// Recent reports
	var recentReports []models.Report
	database.DB.Preload("Category").
		Order("created_at DESC").
		Limit(5).
		Find(&recentReports)

	return utils.Success(c, fiber.Map{
		"users": fiber.Map{
			"total": totalUsers,
			"new":   newUsers,
		},
		"proposals": fiber.Map{
			"total":     totalProposals,
			"published": publishedProposals,
			"submitted": submittedProposals,
		},
		"reports": fiber.Map{
			"total":    totalReports,
			"pending":  pendingReports,
			"resolved": resolvedReports,
		},
		"evaluations": fiber.Map{
			"total": totalEvaluations,
		},
		"reports_by_province": reportsByProvince,
		"recent_proposals":    recentProposals,
		"recent_reports":      recentReports,
	}, "Statistiques du tableau de bord")
}

// GetUsers godoc
// GET /api/v1/users  (admin only)
func GetUsers(c *fiber.Ctx) error {
	p := utils.GetPagination(c)
	role := c.Query("role")
	search := c.Query("search")

	query := database.DB.Model(&models.User{})
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if search != "" {
		query = query.Where("full_name ILIKE ? OR email ILIKE ? OR phone ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var users []models.User
	query.Order("created_at DESC").Limit(p.PerPage).Offset(p.Offset).Find(&users)

	return utils.SuccessWithMeta(c, users, utils.BuildMeta(total, p), "Utilisateurs récupérés")
}

// UpdateUserRole godoc
// PATCH /api/v1/users/:id/role  (admin only)
func UpdateUserRole(c *fiber.Ctx) error {
	var user models.User
	if err := database.DB.First(&user, "id = ?", c.Params("id")).Error; err != nil {
		return utils.NotFound(c, "Utilisateur non trouvé")
	}

	var input struct {
		Role string `json:"role"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "Données invalides")
	}

	validRoles := map[string]bool{
		"citizen": true, "moderator": true, "representative": true, "admin": true,
	}
	if !validRoles[input.Role] {
		return utils.BadRequest(c, "Rôle invalide")
	}

	database.DB.Model(&user).Update("role", input.Role)
	return utils.Success(c, user, "Rôle mis à jour")
}
