package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kgermando/pdp-api/config"
)

// BackblazeService gère les uploads vers Backblaze B2
type BackblazeService struct {
	KeyID          string
	ApplicationKey string
	BucketID       string
	BucketName     string
	AuthToken      string
	APIURL         string
	DownloadURL    string
	UploadURL      string
	UploadToken    string
}

type b2AuthResponse struct {
	AccountID          string `json:"accountId"`
	AuthorizationToken string `json:"authorizationToken"`
	APIURL             string `json:"apiUrl"`
	DownloadURL        string `json:"downloadUrl"`
}

type b2UploadURLResponse struct {
	BucketID           string `json:"bucketId"`
	UploadURL          string `json:"uploadUrl"`
	AuthorizationToken string `json:"authorizationToken"`
}

type b2UploadResponse struct {
	FileID   string `json:"fileId"`
	FileName string `json:"fileName"`
}

// NewBackblazeService crée et initialise le service Backblaze B2
func NewBackblazeService() (*BackblazeService, error) {
	cfg := config.AppConfig

	if cfg.B2KeyID == "" || cfg.B2ApplicationKey == "" || cfg.B2BucketID == "" || cfg.B2BucketName == "" {
		return nil, fmt.Errorf("identifiants Backblaze B2 manquants dans les variables d'environnement")
	}

	svc := &BackblazeService{
		KeyID:          cfg.B2KeyID,
		ApplicationKey: cfg.B2ApplicationKey,
		BucketID:       cfg.B2BucketID,
		BucketName:     cfg.B2BucketName,
	}

	if err := svc.authorize(); err != nil {
		return nil, fmt.Errorf("échec d'authentification Backblaze: %w", err)
	}

	return svc, nil
}

// authorize s'authentifie auprès de l'API Backblaze B2
func (b *BackblazeService) authorize() error {
	req, err := http.NewRequest("GET", "https://api.backblazeb2.com/b2api/v2/b2_authorize_account", nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(b.KeyID, b.ApplicationKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("autorisation refusée: %s", string(body))
	}

	var authResp b2AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return err
	}

	b.AuthToken = authResp.AuthorizationToken
	b.APIURL = authResp.APIURL
	b.DownloadURL = authResp.DownloadURL

	return nil
}

// getUploadURL obtient une URL d'upload temporaire
func (b *BackblazeService) getUploadURL() error {
	url := b.APIURL + "/b2api/v2/b2_get_upload_url"

	payload := map[string]string{"bucketId": b.BucketID}
	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", b.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("impossible d'obtenir l'URL d'upload: %s", string(body))
	}

	var uploadResp b2UploadURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return err
	}

	b.UploadURL = uploadResp.UploadURL
	b.UploadToken = uploadResp.AuthorizationToken

	return nil
}

// UploadFile envoie un fichier vers Backblaze B2 dans le dossier spécifié
func (b *BackblazeService) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	// Renouveler l'URL d'upload à chaque appel (les tokens expirent)
	if err := b.getUploadURL(); err != nil {
		// Ré-autorisation si le token principal a expiré
		if authErr := b.authorize(); authErr != nil {
			return "", authErr
		}
		if err = b.getUploadURL(); err != nil {
			return "", err
		}
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	filename := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	req, err := http.NewRequest("POST", b.UploadURL, bytes.NewBuffer(fileBytes))
	if err != nil {
		return "", err
	}

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	req.Header.Set("Authorization", b.UploadToken)
	req.Header.Set("X-Bz-File-Name", filename)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Bz-Content-Sha1", "do_not_verify")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("échec de l'upload: %s", string(body))
	}

	var uploadResp b2UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return "", err
	}

	// URL publique du fichier
	fileURL := fmt.Sprintf("%s/file/%s/%s", b.DownloadURL, b.BucketName, filename)
	return fileURL, nil
}

// UploadImage envoie une image vers le dossier "images"
func (b *BackblazeService) UploadImage(file *multipart.FileHeader) (string, error) {
	return b.UploadFile(file, "images")
}

// UploadVideo envoie une vidéo vers le dossier "videos"
func (b *BackblazeService) UploadVideo(file *multipart.FileHeader) (string, error) {
	return b.UploadFile(file, "videos")
}

// UploadDocument envoie un document vers le dossier "documents"
func (b *BackblazeService) UploadDocument(file *multipart.FileHeader) (string, error) {
	return b.UploadFile(file, "documents")
}

// DeleteFile supprime un fichier de Backblaze B2 par son nom
func (b *BackblazeService) DeleteFile(fileID, fileName string) error {
	url := b.APIURL + "/b2api/v2/b2_delete_file_version"

	payload := map[string]string{
		"fileId":   fileID,
		"fileName": fileName,
	}
	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", b.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("échec de la suppression: %s", string(body))
	}

	return nil
}

// ValidateImageFile valide le type et la taille d'une image
func ValidateImageFile(file *multipart.FileHeader) error {
	const maxSize = 5 * 1024 * 1024 // 5 MB
	if file.Size > maxSize {
		return fmt.Errorf("l'image dépasse la limite de 5 Mo")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true,
		".gif": true, ".webp": true,
	}
	if !allowed[ext] {
		return fmt.Errorf("format d'image non autorisé. Formats acceptés: jpg, jpeg, png, gif, webp")
	}

	return nil
}

// ValidateVideoFile valide le type et la taille d'une vidéo
func ValidateVideoFile(file *multipart.FileHeader) error {
	const maxSize = 100 * 1024 * 1024 // 100 MB
	if file.Size > maxSize {
		return fmt.Errorf("la vidéo dépasse la limite de 100 Mo")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{
		".mp4": true, ".mov": true, ".avi": true,
		".webm": true, ".mkv": true,
	}
	if !allowed[ext] {
		return fmt.Errorf("format vidéo non autorisé. Formats acceptés: mp4, mov, avi, webm, mkv")
	}

	return nil
}
