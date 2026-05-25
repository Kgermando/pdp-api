package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/pdp-api/handlers"
	"github.com/kgermando/pdp-api/middleware"
)

func Setup(app *fiber.App) {
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "Parlement Digital du Peuple API"})
	})

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", handlers.Register)
	auth.Post("/login", handlers.Login)
	auth.Get("/me", middleware.Protected(), handlers.GetProfile)
	auth.Put("/me", middleware.Protected(), handlers.UpdateProfile)
	auth.Post("/change-password", middleware.Protected(), handlers.ChangePassword)

	// Categories (public read)
	categories := api.Group("/categories")
	categories.Get("/", handlers.GetCategories)
	categories.Post("/", middleware.Protected(), middleware.RequireRole("admin"), handlers.CreateCategory)
	categories.Put("/:id", middleware.Protected(), middleware.RequireRole("admin"), handlers.UpdateCategory)

	// Proposals
	proposals := api.Group("/proposals")
	proposals.Get("/", handlers.GetProposals)
	proposals.Get("/:id", handlers.GetProposal)
	proposals.Post("/", middleware.Protected(), handlers.CreateProposal)
	proposals.Put("/:id", middleware.Protected(), handlers.UpdateProposal)
	proposals.Delete("/:id", middleware.Protected(), handlers.DeleteProposal)
	proposals.Post("/:id/vote", middleware.Protected(), handlers.VoteProposal)
	proposals.Post("/:id/comments", middleware.Protected(), handlers.AddComment)

	// Reports / Alerts
	reports := api.Group("/reports")
	reports.Get("/", handlers.GetReports)
	reports.Get("/map", handlers.GetMapReports)
	reports.Get("/stats", handlers.GetReportStats)
	reports.Get("/:id", handlers.GetReport)
	reports.Post("/", handlers.CreateReport) // Anonymous allowed
	reports.Patch("/:id/status", middleware.Protected(), middleware.RequireRole("admin", "moderator"), handlers.UpdateReportStatus)

	// Institutions & Evaluations
	institutions := api.Group("/institutions")
	institutions.Get("/", handlers.GetInstitutions)
	institutions.Get("/:id", handlers.GetInstitution)
	institutions.Post("/", middleware.Protected(), middleware.RequireRole("admin"), handlers.CreateInstitution)
	institutions.Post("/:id/evaluate", handlers.EvaluateInstitution) // Anonymous allowed
	institutions.Get("/:id/evaluations", handlers.GetInstitutionEvaluations)

	// Forums / Salons de discussion citoyenne
	forums := api.Group("/forums")
	forums.Get("/", handlers.GetForums)
	forums.Get("/trending", handlers.GetTrendingForums)
	forums.Get("/:id", handlers.GetForum)
	forums.Post("/", middleware.Protected(), handlers.CreateForum)
	forums.Delete("/:id", middleware.Protected(), handlers.DeleteForum)
	forums.Patch("/:id", middleware.Protected(), middleware.RequireRole("admin", "moderator"), handlers.PatchForum)
	forums.Post("/:id/messages", middleware.Protected(), handlers.AddForumMessage)
	forums.Post("/:id/like", middleware.Protected(), handlers.LikeForum)
	forums.Post("/messages/:id/like", middleware.Protected(), handlers.LikeForumMessage)

	// Notifications (authenticated)
	notifications := api.Group("/notifications", middleware.Protected())
	notifications.Get("/", handlers.GetNotifications)
	notifications.Get("/unread-count", handlers.GetUnreadCount)
	notifications.Patch("/read-all", handlers.MarkAllNotificationsRead)
	notifications.Patch("/:id/read", handlers.MarkNotificationRead)

	// Admin / Dashboard
	admin := api.Group("/admin", middleware.Protected(), middleware.RequireRole("admin", "moderator"))
	admin.Get("/dashboard", handlers.GetDashboardStats)
	admin.Get("/users", middleware.RequireRole("admin"), handlers.GetUsers)
	admin.Patch("/users/:id/role", middleware.RequireRole("admin"), handlers.UpdateUserRole)

	// Upload (images & vidéos → Backblaze B2)
	upload := api.Group("/upload", middleware.Protected())
	upload.Post("/image", handlers.UploadImage)
	upload.Post("/video", handlers.UploadVideo)
}
