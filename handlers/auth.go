package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/pdp-api/database"
	"github.com/kgermando/pdp-api/models"
	"github.com/kgermando/pdp-api/utils"
)

type RegisterInput struct {
	FullName  string `json:"full_name" validate:"required,min=3"`
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone"`
	Password  string `json:"password" validate:"required,min=8"`
	Province  string `json:"province"`
	City      string `json:"city"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Register godoc
// POST /api/v1/auth/register
func Register(c *fiber.Ctx) error {
	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "Données invalides")
	}

	if len(input.FullName) < 3 {
		return utils.BadRequest(c, "Le nom complet doit avoir au moins 3 caractères")
	}
	if len(input.Password) < 8 {
		return utils.BadRequest(c, "Le mot de passe doit avoir au moins 8 caractères")
	}

	// Check duplicate email
	var existing models.User
	if result := database.DB.Where("email = ?", input.Email).First(&existing); result.Error == nil {
		return utils.BadRequest(c, "Cet email est déjà utilisé")
	}
	if input.Phone != "" {
		if result := database.DB.Where("phone = ?", input.Phone).First(&existing); result.Error == nil {
			return utils.BadRequest(c, "Ce numéro de téléphone est déjà utilisé")
		}
	}

	hashed, err := utils.HashPassword(input.Password)
	if err != nil {
		return utils.InternalError(c, "Erreur lors du chiffrement du mot de passe")
	}

	user := models.User{
		FullName: input.FullName,
		Email:    input.Email,
		Phone:    input.Phone,
		Password: hashed,
		Role:     models.RoleCitizen,
		Province: input.Province,
		City:     input.City,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return utils.InternalError(c, "Erreur lors de la création du compte")
	}

	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return utils.InternalError(c, "Erreur lors de la génération du token")
	}

	return utils.Created(c, fiber.Map{
		"token": token,
		"user":  user,
	}, "Compte créé avec succès")
}

// Login godoc
// POST /api/v1/auth/login
func Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "Données invalides")
	}

	var user models.User
	if err := database.DB.Where("email = ? AND is_active = true", input.Email).First(&user).Error; err != nil {
		return utils.Unauthorized(c, "Email ou mot de passe incorrect")
	}

	if !utils.CheckPassword(input.Password, user.Password) {
		return utils.Unauthorized(c, "Email ou mot de passe incorrect")
	}

	now := time.Now()
	database.DB.Model(&user).Update("last_login_at", &now)

	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return utils.InternalError(c, "Erreur lors de la génération du token")
	}

	return utils.Success(c, fiber.Map{
		"token": token,
		"user":  user,
	}, "Connexion réussie")
}

// GetProfile godoc
// GET /api/v1/auth/me
func GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	var user models.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		return utils.NotFound(c, "Utilisateur non trouvé")
	}
	return utils.Success(c, user, "Profil récupéré")
}

// UpdateProfile godoc
// PUT /api/v1/auth/me
func UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	var user models.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		return utils.NotFound(c, "Utilisateur non trouvé")
	}

	var input struct {
		FullName     string `json:"full_name"`
		Phone        string `json:"phone"`
		Province     string `json:"province"`
		City         string `json:"city"`
		Neighborhood string `json:"neighborhood"`
		Bio          string `json:"bio"`
		PrefersSMS   *bool  `json:"prefers_sms"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "Données invalides")
	}

	updates := map[string]interface{}{}
	if input.FullName != "" {
		updates["full_name"] = input.FullName
	}
	if input.Phone != "" {
		updates["phone"] = input.Phone
	}
	if input.Province != "" {
		updates["province"] = input.Province
	}
	if input.City != "" {
		updates["city"] = input.City
	}
	if input.Neighborhood != "" {
		updates["neighborhood"] = input.Neighborhood
	}
	if input.Bio != "" {
		updates["bio"] = input.Bio
	}
	if input.PrefersSMS != nil {
		updates["prefers_sms"] = *input.PrefersSMS
	}

	database.DB.Model(&user).Updates(updates)
	return utils.Success(c, user, "Profil mis à jour")
}

// ChangePassword godoc
// POST /api/v1/auth/change-password
func ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	var user models.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		return utils.NotFound(c, "Utilisateur non trouvé")
	}

	var input struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "Données invalides")
	}

	if !utils.CheckPassword(input.OldPassword, user.Password) {
		return utils.BadRequest(c, "Ancien mot de passe incorrect")
	}
	if len(input.NewPassword) < 8 {
		return utils.BadRequest(c, "Le nouveau mot de passe doit avoir au moins 8 caractères")
	}

	hashed, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return utils.InternalError(c, "Erreur lors du chiffrement")
	}

	database.DB.Model(&user).Update("password", hashed)
	return utils.Success(c, nil, "Mot de passe modifié avec succès")
}
