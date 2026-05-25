package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/pdp-api/utils"
)

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.Unauthorized(c, "Autorisation requise")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return utils.Unauthorized(c, "Format de token invalide")
		}

		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			return utils.Unauthorized(c, "Token invalide ou expiré")
		}

		c.Locals("userID", claims.UserID)
		c.Locals("userEmail", claims.Email)
		c.Locals("userRole", claims.Role)
		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("userRole").(string)
		if !ok {
			return utils.Unauthorized(c, "Non authentifié")
		}
		for _, r := range roles {
			if role == r {
				return c.Next()
			}
		}
		return utils.Forbidden(c, "Accès refusé: rôle insuffisant")
	}
}
