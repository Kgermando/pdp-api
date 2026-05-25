package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kgermando/pdp-api/database"
	"github.com/kgermando/pdp-api/models"
	"github.com/kgermando/pdp-api/utils"
	"gorm.io/gorm"
)

// GetForums godoc
// GET /api/v1/forums
func GetForums(c *fiber.Ctx) error {
	p := utils.GetPagination(c)
	category := c.Query("category")
	province := c.Query("province")
	search := c.Query("search")

	query := database.DB.Model(&models.Forum{}).Preload("Author")

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if province != "" {
		query = query.Where("province = ? OR province = ''", province)
	}
	if search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ? OR tags ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var forums []models.Forum
	query.Order("is_pinned DESC, last_activity_at DESC NULLS LAST, created_at DESC").
		Limit(p.PerPage).
		Offset(p.Offset).
		Find(&forums)

	return utils.SuccessWithMeta(c, forums, utils.BuildMeta(total, p), "Forums récupérés")
}

// GetTrendingForums godoc
// GET /api/v1/forums/trending
func GetTrendingForums(c *fiber.Ctx) error {
	var forums []models.Forum
	database.DB.Model(&models.Forum{}).
		Preload("Author").
		Where("is_locked = false").
		Order("message_count DESC, views_count DESC, likes_count DESC").
		Limit(10).
		Find(&forums)
	return utils.Success(c, forums, "Forums tendance récupérés")
}

// GetForum godoc
// GET /api/v1/forums/:id
func GetForum(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	var forum models.Forum
	if err := database.DB.
		Preload("Author").
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Where("parent_id IS NULL").Order("created_at ASC")
		}).
		Preload("Messages.Author").
		Preload("Messages.Replies", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("Messages.Replies.Author").
		First(&forum, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "Forum non trouvé")
	}

	database.DB.Model(&forum).UpdateColumn("views_count", gorm.Expr("views_count + 1"))
	return utils.Success(c, forum, "Forum récupéré")
}

// CreateForum godoc
// POST /api/v1/forums
func CreateForum(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Category    string `json:"category"`
		Province    string `json:"province"`
		Tags        string `json:"tags"`
		IsAnonymous bool   `json:"is_anonymous"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "Données invalides")
	}
	if input.Title == "" || input.Category == "" {
		return utils.BadRequest(c, "Titre et catégorie sont obligatoires")
	}

	now := time.Now()
	forum := models.Forum{
		Title:          input.Title,
		Description:    input.Description,
		Category:       models.ForumCategory(input.Category),
		AuthorID:       userID,
		Province:       input.Province,
		Tags:           input.Tags,
		IsAnonymous:    input.IsAnonymous,
		LastActivityAt: &now,
	}

	if err := database.DB.Create(&forum).Error; err != nil {
		return utils.InternalError(c, "Erreur lors de la création")
	}

	database.DB.Preload("Author").First(&forum, "id = ?", forum.ID)
	return utils.Created(c, forum, "Salon de discussion créé")
}

// AddForumMessage godoc
// POST /api/v1/forums/:id/messages
func AddForumMessage(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	forumID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	var forum models.Forum
	if err := database.DB.First(&forum, "id = ?", forumID).Error; err != nil {
		return utils.NotFound(c, "Forum non trouvé")
	}
	if forum.IsLocked {
		return utils.BadRequest(c, "Ce salon est verrouillé par les modérateurs")
	}

	var input struct {
		Content     string  `json:"content"`
		ParentID    *string `json:"parent_id"`
		IsAnonymous bool    `json:"is_anonymous"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "Données invalides")
	}
	if input.Content == "" {
		return utils.BadRequest(c, "Le message ne peut pas être vide")
	}

	var parentID *uuid.UUID
	if input.ParentID != nil && *input.ParentID != "" {
		pid, err := uuid.Parse(*input.ParentID)
		if err == nil {
			parentID = &pid
		}
	}

	// Official badge for admins, moderators, institution accounts
	var user models.User
	database.DB.First(&user, "id = ?", userID)
	isOfficial := user.Role == "admin" || user.Role == "moderator" || user.Role == "institution"

	msg := models.ForumMessage{
		ForumID:     forumID,
		ParentID:    parentID,
		AuthorID:    userID,
		Content:     input.Content,
		IsAnonymous: input.IsAnonymous,
		IsOfficial:  isOfficial,
	}

	if err := database.DB.Create(&msg).Error; err != nil {
		return utils.InternalError(c, "Erreur lors de l'envoi du message")
	}

	now := time.Now()
	database.DB.Model(&forum).Updates(map[string]interface{}{
		"message_count":    gorm.Expr("message_count + 1"),
		"last_activity_at": now,
	})

	database.DB.Preload("Author").First(&msg, "id = ?", msg.ID)
	return utils.Created(c, msg, "Message envoyé")
}

// LikeForum godoc
// POST /api/v1/forums/:id/like
func LikeForum(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	forumID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	var existing models.ForumLike
	result := database.DB.Where("forum_id = ? AND user_id = ? AND message_id IS NULL", forumID, userID).First(&existing)
	if result.Error == nil {
		database.DB.Delete(&existing)
		database.DB.Model(&models.Forum{}).Where("id = ?", forumID).
			UpdateColumn("likes_count", gorm.Expr("GREATEST(likes_count - 1, 0)"))
		return utils.Success(c, fiber.Map{"liked": false}, "Like retiré")
	}

	like := models.ForumLike{ForumID: &forumID, UserID: userID}
	database.DB.Create(&like)
	database.DB.Model(&models.Forum{}).Where("id = ?", forumID).
		UpdateColumn("likes_count", gorm.Expr("likes_count + 1"))
	return utils.Success(c, fiber.Map{"liked": true}, "Like ajouté")
}

// LikeForumMessage godoc
// POST /api/v1/forums/messages/:id/like
func LikeForumMessage(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	msgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	var existing models.ForumLike
	result := database.DB.Where("message_id = ? AND user_id = ?", msgID, userID).First(&existing)
	if result.Error == nil {
		database.DB.Delete(&existing)
		database.DB.Model(&models.ForumMessage{}).Where("id = ?", msgID).
			UpdateColumn("likes_count", gorm.Expr("GREATEST(likes_count - 1, 0)"))
		return utils.Success(c, fiber.Map{"liked": false}, "Like retiré")
	}

	like := models.ForumLike{MessageID: &msgID, UserID: userID}
	database.DB.Create(&like)
	database.DB.Model(&models.ForumMessage{}).Where("id = ?", msgID).
		UpdateColumn("likes_count", gorm.Expr("likes_count + 1"))
	return utils.Success(c, fiber.Map{"liked": true}, "Like ajouté")
}

// DeleteForum godoc
// DELETE /api/v1/forums/:id
func DeleteForum(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	role := c.Locals("userRole").(string)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	var forum models.Forum
	if err := database.DB.First(&forum, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "Forum non trouvé")
	}

	if forum.AuthorID != userID && role != "admin" && role != "moderator" {
		return utils.Forbidden(c, "Vous n'êtes pas autorisé à supprimer ce forum")
	}

	database.DB.Delete(&forum)
	return utils.Success(c, nil, "Forum supprimé")
}

// PatchForum godoc (admin/moderator only — pin/lock)
// PATCH /api/v1/forums/:id
func PatchForum(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	var forum models.Forum
	if err := database.DB.First(&forum, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "Forum non trouvé")
	}

	var input struct {
		IsPinned *bool `json:"is_pinned"`
		IsLocked *bool `json:"is_locked"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "Données invalides")
	}

	updates := map[string]interface{}{}
	if input.IsPinned != nil {
		updates["is_pinned"] = *input.IsPinned
	}
	if input.IsLocked != nil {
		updates["is_locked"] = *input.IsLocked
	}
	database.DB.Model(&forum).Updates(updates)
	return utils.Success(c, forum, "Forum mis à jour")
}
