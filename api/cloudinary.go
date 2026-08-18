package api

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

	"github.com/gin-gonic/gin"
	"github.com/MyTeleProject2026/Slotopol-server/config"
)

// CloudinaryImage represents the response from Cloudinary upload
type CloudinaryImage struct {
	PublicID     string `json:"public_id"`
	Version      string `json:"version"`
	Signature    string `json:"signature"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Format       string `json:"format"`
	ResourceType string `json:"resource_type"`
	CreatedAt    string `json:"created_at"`
	Bytes        int64  `json:"bytes"`
	Type         string `json:"type"`
	URL          string `json:"url"`
	SecureURL    string `json:"secure_url"`
	Etag         string `json:"etag"`
}

// CloudinaryUploadParams holds upload configuration
type CloudinaryUploadParams struct {
	File           multipart.File
	Filename       string
	PublicID       string
	Folder         string
	Tags           []string
	Transformation string
}

// UploadToCloudinary uploads a file to Cloudinary
func UploadToCloudinary(params CloudinaryUploadParams) (*CloudinaryImage, error) {
	cloudName := Cfg.Cloudinary.CloudName
	apiKey := Cfg.Cloudinary.APIKey
	apiSecret := Cfg.Cloudinary.APISecret
	uploadFolder := Cfg.Cloudinary.UploadFolder

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("Cloudinary credentials not configured")
	}

	// Prepare multipart form
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add file
	part, err := w.CreateFormFile("file", params.Filename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, params.File); err != nil {
		return nil, err
	}

	// Add parameters
	if err := w.WriteField("upload_preset", "ml_default"); err != nil {
		return nil, err
	}
	if params.PublicID != "" {
		if err := w.WriteField("public_id", params.PublicID); err != nil {
			return nil, err
		}
	}
	if uploadFolder != "" {
		folder := uploadFolder
		if params.Folder != "" {
			folder = uploadFolder + "/" + params.Folder
		}
		if err := w.WriteField("folder", folder); err != nil {
			return nil, err
		}
	}
	if len(params.Tags) > 0 {
		if err := w.WriteField("tags", strings.Join(params.Tags, ",")); err != nil {
			return nil, err
		}
	}
	if params.Transformation != "" {
		if err := w.WriteField("transformation", params.Transformation); err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	// Build Cloudinary upload URL
	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", cloudName)

	// Create request
	req, err := http.NewRequest("POST", uploadURL, &b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.SetBasicAuth(apiKey, apiSecret)

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("Cloudinary upload failed: %s", errResp.Error.Message)
		}
		return nil, fmt.Errorf("Cloudinary upload failed with status %d", resp.StatusCode)
	}

	var result CloudinaryImage
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ApiUploadImage handles image upload to Cloudinary
func ApiUploadImage(c *gin.Context) {
	var err error
	var arg struct {
		Folder string `json:"folder" form:"folder"`
		Tags   string `json:"tags" form:"tags"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		Ret400(c, 0, fmt.Errorf("file is required"))
		return
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		Ret500(c, 0, err)
		return
	}
	defer src.Close()

	// Check file size
	if file.Size > Cfg.Uploads.MaxFileSize {
		Ret400(c, 0, fmt.Errorf("file size exceeds maximum of %d bytes", Cfg.Uploads.MaxFileSize))
		return
	}

	// Check file type
	allowedTypes := Cfg.Uploads.AllowedTypes
	contentType := file.Header.Get("Content-Type")
	allowed := false
	for _, t := range allowedTypes {
		if contentType == t {
			allowed = true
			break
		}
	}
	if !allowed {
		Ret400(c, 0, fmt.Errorf("file type %s is not allowed", contentType))
		return
	}

	// Generate public ID
	publicID := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	publicID = strings.ReplaceAll(publicID, " ", "_")

	// Upload to Cloudinary
	params := CloudinaryUploadParams{
		File:     src,
		Filename: file.Filename,
		PublicID: publicID,
		Folder:   arg.Folder,
		Tags:     strings.Split(arg.Tags, ","),
	}

	result, err := UploadToCloudinary(params)
	if err != nil {
		Ret500(c, 0, err)
		return
	}

	// Save to database
	user := c.MustGet(userKey).(*User)
	if err := SafeTransaction(XormStorage, func(session *Session) error {
		_, err := session.Exec(
			`INSERT INTO cloudinary_images (public_id, url, secure_url, format, width, height, bytes, uploaded_by, folder, tags)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			result.PublicID, result.URL, result.SecureURL, result.Format,
			result.Width, result.Height, result.Bytes, user.UID,
			arg.Folder, arg.Tags,
		)
		return err
	}); err != nil {
		Ret500(c, 0, err)
		return
	}

	RetOk(c, gin.H{
		"success": true,
		"image":   result,
	})
}

// ApiGetImages retrieves uploaded images from Cloudinary
func ApiGetImages(c *gin.Context) {
	var arg struct {
		Folder string `json:"folder" form:"folder"`
	}

	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	var images []CloudinaryImage
	var query = XormStorage.Where("1=1")
	if arg.Folder != "" {
		query = query.Where("folder = ?", arg.Folder)
	}

	if err := query.OrderBy("created_at DESC").Find(&images); err != nil {
		Ret500(c, 0, err)
		return
	}

	RetOk(c, gin.H{
		"success": true,
		"images":  images,
	})
}

// ApiDeleteImage removes an image from Cloudinary
func ApiDeleteImage(c *gin.Context) {
	var arg struct {
		PublicID string `json:"public_id" binding:"required"`
	}

	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	cloudName := Cfg.Cloudinary.CloudName
	apiKey := Cfg.Cloudinary.APIKey
	apiSecret := Cfg.Cloudinary.APISecret

	// Build Cloudinary delete URL
	deleteURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/destroy", cloudName)

	// Prepare request body
	body := map[string]string{
		"public_id": arg.PublicID,
	}
	bodyJSON, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", deleteURL, bytes.NewReader(bodyJSON))
	if err != nil {
		Ret500(c, 0, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(apiKey, apiSecret)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		Ret500(c, 0, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		Ret500(c, 0, fmt.Errorf("Cloudinary delete failed with status %d", resp.StatusCode))
		return
	}

	// Delete from database
	if _, err := XormStorage.Exec("DELETE FROM cloudinary_images WHERE public_id = ?", arg.PublicID); err != nil {
		Ret500(c, 0, err)
		return
	}

	RetOk(c, gin.H{
		"success": true,
		"message": "Image deleted successfully",
	})
}
