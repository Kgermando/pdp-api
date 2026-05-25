package database

import (
	"log"

	"github.com/kgermando/pdp-api/config"
	"github.com/kgermando/pdp-api/models"
	"golang.org/x/crypto/bcrypt"
)

// SeedSuperAdmin crée le compte super administrateur au démarrage
// si aucun compte super_admin n'existe encore.
func SeedSuperAdmin() {
	cfg := config.AppConfig

	if cfg.SuperAdminPassword == "" {
		log.Println("⚠️  SUPER_ADMIN_PASSWORD non défini — seed super admin ignoré")
		return
	}

	var count int64
	DB.Model(&models.User{}).Where("role = ?", models.RoleSuperAdmin).Count(&count)
	if count > 0 {
		log.Println("✅ Super admin déjà présent en base")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(cfg.SuperAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("❌ Impossible de hasher le mot de passe super admin : %v", err)
	}

	superAdmin := models.User{
		FullName:   cfg.SuperAdminName,
		Email:      cfg.SuperAdminEmail,
		Password:   string(hashed),
		Role:       models.RoleSuperAdmin,
		IsVerified: true,
		IsActive:   true,
	}

	if err := DB.Create(&superAdmin).Error; err != nil {
		log.Fatalf("❌ Impossible de créer le super admin : %v", err)
	}

	log.Printf("🚀 Super admin créé : %s (%s)", superAdmin.FullName, superAdmin.Email)
}
