package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/pdp-api/services"
	"github.com/kgermando/pdp-api/utils"
)

// UploadImage godoc
// POST /api/v1/upload/image
// Authentifié — retourne l'URL publique de l'image uploadée sur Backblaze B2
func UploadImage(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return utils.BadRequest(c, "Fichier manquant dans la requête (champ: file)")
	}

	if err := services.ValidateImageFile(file); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	b2, err := services.NewBackblazeService()
	if err != nil {
		return utils.InternalError(c, "Service de stockage indisponible")
	}

	url, err := b2.UploadImage(file)
	if err != nil {
		return utils.InternalError(c, "Échec de l'upload de l'image")
	}

	return utils.Success(c, fiber.Map{
		"url":      url,
		"filename": file.Filename,
		"size":     file.Size,
	}, "Image uploadée avec succès")
}

// UploadVideo godoc
// POST /api/v1/upload/video
// Authentifié — retourne l'URL publique de la vidéo uploadée sur Backblaze B2
func UploadVideo(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return utils.BadRequest(c, "Fichier manquant dans la requête (champ: file)")
	}

	if err := services.ValidateVideoFile(file); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	b2, err := services.NewBackblazeService()
	if err != nil {
		return utils.InternalError(c, "Service de stockage indisponible")
	}

	url, err := b2.UploadVideo(file)
	if err != nil {
		return utils.InternalError(c, "Échec de l'upload de la vidéo")
	}

	return utils.Success(c, fiber.Map{
		"url":      url,
		"filename": file.Filename,
		"size":     file.Size,
	}, "Vidéo uploadée avec succès")
}
