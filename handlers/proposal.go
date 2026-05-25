package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kgermando/pdp-api/database"
	"github.com/kgermando/pdp-api/models"
	"github.com/kgermando/pdp-api/utils"
)

// GetProposals godoc
// GET /api/v1/proposals
func GetProposals(c *fiber.Ctx) error {
	p := utils.GetPagination(c)
	status := c.Query("status")
	province := c.Query("province")
	categoryID := c.Query("category_id")
	search := c.Query("search")

	query := database.DB.Model(&models.Proposal{}).
		Preload("Category").
		Preload("Author")

	if status != "" {
		query = query.Where("status = ?", status)
	} else {
		query = query.Where("status = ?", models.ProposalPublished)
	}
	if province != "" {
		query = query.Where("province = ?", province)
	}
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if search != "" {
		query = query.Where("title ILIKE ? OR summary ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var proposals []models.Proposal
	query.Order("created_at DESC").
		Limit(p.PerPage).
		Offset(p.Offset).
		Find(&proposals)

	return utils.SuccessWithMeta(c, proposals, utils.BuildMeta(total, p), "Propositions récupérées")
}

// GetProposal godoc
// GET /api/v1/proposals/:id
func GetProposal(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	var proposal models.Proposal
	if err := database.DB.
		Preload("Category").
		Preload("Author").
		Preload("Comments.User").
		First(&proposal, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "Proposition non trouvée")
	}

	database.DB.Model(&proposal).UpdateColumn("views_count", proposal.ViewsCount+1)
	return utils.Success(c, proposal, "Proposition récupérée")
}

// CreateProposal godoc
// POST /api/v1/proposals
func CreateProposal(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var input struct {
		Title       string `json:"title"`
		Summary     string `json:"summary"`
		Content     string `json:"content"`
		CategoryID  string `json:"category_id"`
		Province    string `json:"province"`
		Tags        string `json:"tags"`
		IsAnonymous bool   `json:"is_anonymous"`
		IsDraft     bool   `json:"is_draft"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "Données invalides")
	}

	if input.Title == "" || input.Content == "" || input.CategoryID == "" {
		return utils.BadRequest(c, "Titre, contenu et catégorie sont obligatoires")
	}

	catID, err := uuid.Parse(input.CategoryID)
	if err != nil {
		return utils.BadRequest(c, "ID de catégorie invalide")
	}

	status := models.ProposalPending
	if input.IsDraft {
		status = models.ProposalDraft
	}

	proposal := models.Proposal{
		Title:       input.Title,
		Summary:     input.Summary,
		Content:     input.Content,
		CategoryID:  catID,
		AuthorID:    userID,
		Status:      status,
		Province:    input.Province,
		Tags:        input.Tags,
		IsAnonymous: input.IsAnonymous,
	}

	if err := database.DB.Create(&proposal).Error; err != nil {
		return utils.InternalError(c, "Erreur lors de la création")
	}

	database.DB.Preload("Category").Preload("Author").First(&proposal, "id = ?", proposal.ID)
	return utils.Created(c, proposal, "Proposition créée avec succès")
}

// UpdateProposal godoc
// PUT /api/v1/proposals/:id
func UpdateProposal(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	role := c.Locals("userRole").(string)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	var proposal models.Proposal
	if err := database.DB.First(&proposal, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "Proposition non trouvée")
	}

	if proposal.AuthorID != userID && role != "admin" && role != "moderator" {
		return utils.Forbidden(c, "Vous ne pouvez modifier que vos propres propositions")
	}

	var input struct {
		Title         string `json:"title"`
		Summary       string `json:"summary"`
		Content       string `json:"content"`
		Status        string `json:"status"`
		ModeratorNote string `json:"moderator_note"`
		Province      string `json:"province"`
		Tags          string `json:"tags"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "Données invalides")
	}

	updates := map[string]interface{}{}
	if input.Title != "" {
		updates["title"] = input.Title
	}
	if input.Summary != "" {
		updates["summary"] = input.Summary
	}
	if input.Content != "" {
		updates["content"] = input.Content
	}
	if input.Province != "" {
		updates["province"] = input.Province
	}
	if input.Tags != "" {
		updates["tags"] = input.Tags
	}
	// Only moderators/admins can change status
	if input.Status != "" && (role == "admin" || role == "moderator") {
		updates["status"] = input.Status
	}
	if input.ModeratorNote != "" && (role == "admin" || role == "moderator") {
		updates["moderator_note"] = input.ModeratorNote
	}

	database.DB.Model(&proposal).Updates(updates)
	return utils.Success(c, proposal, "Proposition mise à jour")
}

// VoteProposal godoc
// POST /api/v1/proposals/:id/vote
func VoteProposal(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	var proposal models.Proposal
	if err := database.DB.First(&proposal, "id = ? AND status = ?", id, models.ProposalPublished).Error; err != nil {
		return utils.NotFound(c, "Proposition non trouvée ou non publiée")
	}

	var input struct {
		VoteType string `json:"vote_type"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "Données invalides")
	}
	if input.VoteType != "for" && input.VoteType != "against" && input.VoteType != "abstain" {
		return utils.BadRequest(c, "Type de vote invalide (for, against, abstain)")
	}

	// Upsert vote
	var existingVote models.ProposalVote
	result := database.DB.Where("proposal_id = ? AND user_id = ?", id, userID).First(&existingVote)

	if result.Error == nil {
		// Update existing vote - adjust counters
		oldType := existingVote.VoteType
		existingVote.VoteType = models.VoteType(input.VoteType)
		database.DB.Save(&existingVote)

		updates := map[string]interface{}{}
		switch oldType {
		case models.VoteFor:
			updates["votes_for"] = proposal.VotesFor - 1
		case models.VoteAgainst:
			updates["votes_against"] = proposal.VotesAgainst - 1
		case models.VoteAbstain:
			updates["votes_abstain"] = proposal.VotesAbstain - 1
		}
		switch models.VoteType(input.VoteType) {
		case models.VoteFor:
			updates["votes_for"] = proposal.VotesFor + 1
		case models.VoteAgainst:
			updates["votes_against"] = proposal.VotesAgainst + 1
		case models.VoteAbstain:
			updates["votes_abstain"] = proposal.VotesAbstain + 1
		}
		database.DB.Model(&proposal).Updates(updates)
	} else {
		vote := models.ProposalVote{
			ProposalID: id,
			UserID:     userID,
			VoteType:   models.VoteType(input.VoteType),
		}
		database.DB.Create(&vote)

		switch models.VoteType(input.VoteType) {
		case models.VoteFor:
			database.DB.Model(&proposal).UpdateColumn("votes_for", proposal.VotesFor+1)
		case models.VoteAgainst:
			database.DB.Model(&proposal).UpdateColumn("votes_against", proposal.VotesAgainst+1)
		case models.VoteAbstain:
			database.DB.Model(&proposal).UpdateColumn("votes_abstain", proposal.VotesAbstain+1)
		}
	}

	database.DB.First(&proposal, "id = ?", id)
	return utils.Success(c, proposal, "Vote enregistré")
}

// AddComment godoc
// POST /api/v1/proposals/:id/comments
func AddComment(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	var input struct {
		Content string `json:"content"`
	}
	if err := c.BodyParser(&input); err != nil || input.Content == "" {
		return utils.BadRequest(c, "Contenu du commentaire requis")
	}

	comment := models.ProposalComment{
		ProposalID: id,
		UserID:     userID,
		Content:    input.Content,
	}
	if err := database.DB.Create(&comment).Error; err != nil {
		return utils.InternalError(c, "Erreur lors de la création du commentaire")
	}

	database.DB.Model(&models.Proposal{}).Where("id = ?", id).
		UpdateColumn("comments_count", database.DB.Model(&models.ProposalComment{}).Where("proposal_id = ?", id))

	database.DB.Preload("User").First(&comment, "id = ?", comment.ID)
	return utils.Created(c, comment, "Commentaire ajouté")
}

// DeleteProposal godoc
// DELETE /api/v1/proposals/:id
func DeleteProposal(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	role := c.Locals("userRole").(string)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.BadRequest(c, "ID invalide")
	}

	var proposal models.Proposal
	if err := database.DB.First(&proposal, "id = ?", id).Error; err != nil {
		return utils.NotFound(c, "Proposition non trouvée")
	}

	if proposal.AuthorID != userID && role != "admin" {
		return utils.Forbidden(c, "Accès refusé")
	}

	database.DB.Delete(&proposal)
	return utils.Success(c, nil, "Proposition supprimée")
}
