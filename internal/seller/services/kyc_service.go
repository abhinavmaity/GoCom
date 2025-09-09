package services

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gocom/main/internal/common/db"
	"gocom/main/internal/models"
)

type KYCService struct {
	DB *gorm.DB
}

func NewKYCService() *KYCService {
	return &KYCService{DB: db.GetDB()}
}

// Upload KYC document
func (ks *KYCService) UploadKYC(sellerID uint, req *UploadKYCRequest) (*models.KYC, error) {
	// Verify seller exists
	var seller models.Seller
	if err := ks.DB.First(&seller, sellerID).Error; err != nil {
		return nil, errors.New("seller not found")
	}

	// Check if document type already exists
	var existing models.KYC
	if err := ks.DB.Where("seller_id = ? AND type = ?", sellerID, req.Type).
		First(&existing).Error; err == nil {
		return nil, errors.New("document of this type already exists")
	}

	kyc := &models.KYC{
		SellerID:    sellerID,
		Type:        req.Type,
		DocumentURL: req.DocumentURL,
		Status:      0, // Pending
		CreatedAt:   time.Now(),
	}

	if err := ks.DB.Create(kyc).Error; err != nil {
		return nil, err
	}

	return kyc, nil
}

// Get all KYC documents for seller
func (ks *KYCService) GetKYCDocuments(sellerID uint) ([]KYCResponse, error) {
	var documents []models.KYC
	if err := ks.DB.Where("seller_id = ?", sellerID).
		Order("created_at DESC").Find(&documents).Error; err != nil {
		return nil, err
	}

	var response []KYCResponse
	for _, doc := range documents {
		response = append(response, KYCResponse{
			ID:          doc.ID,
			Type:        doc.Type,
			DocumentURL: doc.DocumentURL,
			Status:      doc.Status,
			StatusText:  ks.getStatusText(doc.Status),
			Remarks:     doc.Remarks,
			CreatedAt:   doc.CreatedAt,
		})
	}

	return response, nil
}

// Get specific KYC document
func (ks *KYCService) GetKYCDocument(sellerID, docID uint) (*KYCResponse, error) {
	var document models.KYC
	if err := ks.DB.Where("id = ? AND seller_id = ?", docID, sellerID).
		First(&document).Error; err != nil {
		return nil, errors.New("document not found")
	}

	response := &KYCResponse{
		ID:          document.ID,
		Type:        document.Type,
		DocumentURL: document.DocumentURL,
		Status:      document.Status,
		StatusText:  ks.getStatusText(document.Status),
		Remarks:     document.Remarks,
		CreatedAt:   document.CreatedAt,
	}

	return response, nil
}

// Delete KYC document
func (ks *KYCService) DeleteKYC(sellerID, docID uint) error {
	result := ks.DB.Where("id = ? AND seller_id = ?", docID, sellerID).
		Delete(&models.KYC{})
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return errors.New("document not found")
	}

	return nil
}

func (ks *KYCService) getStatusText(status int) string {
	switch status {
	case 0:
		return "Pending"
	case 1:
		return "Approved"
	case 2:
		return "Rejected"
	default:
		return "Unknown"
	}
}

// DTOs
type UploadKYCRequest struct {
	Type        string `json:"type" binding:"required"`        // PAN, GSTIN, etc.
	DocumentURL string `json:"document_url" binding:"required"` // MinIO/S3 URL
}

type KYCResponse struct {
    ID          uint      `json:"id"`           // Was getting null
    Type        string    `json:"type"`
    DocumentURL string    `json:"document_url"`
    Status      int       `json:"status"`
    StatusText  string    `json:"status_text"`
    Remarks     string    `json:"remarks"`
    CreatedAt   time.Time `json:"created_at"`
}

