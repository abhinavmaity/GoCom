package services

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gocom/main/internal/common/db"
	"gocom/main/internal/models"
)

type SellerService struct {
	DB *gorm.DB
}

func NewSellerService() *SellerService {
	return &SellerService{
		DB: db.GetDB(),
	}
}

// Register new seller
func (ss *SellerService) CreateSeller(userID uint, req *CreateSellerRequest) (*models.Seller, error) {
	// Check if user already has a seller account
	var existingSeller models.Seller
	if err := ss.DB.Where("legal_name = ? OR pan = ?", req.LegalName, req.PAN).First(&existingSeller).Error; err == nil {
		return nil, errors.New("seller with this name or PAN already exists")
	}

	// Create seller
	seller := &models.Seller{
		LegalName:   req.LegalName,
		DisplayName: req.DisplayName,
		GSTIN:       req.GSTIN,
		PAN:         req.PAN,
		BankRef:     req.BankRef,
		Status:      0, // Pending approval
		RiskScore:   ss.calculateRiskScore(req),
	}

	// Begin transaction
	tx := ss.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Save seller
	if err := tx.Create(seller).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create seller-user mapping
	sellerUser := &models.SellerUser{
		SellerID: seller.ID,
		UserID:   userID,
		Role:     "owner",
		Status:   1,
	}

	if err := tx.Create(sellerUser).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return seller, nil
}

// Get seller by ID
func (ss *SellerService) GetSeller(sellerID, userID uint) (*SellerResponse, error) {
	// Verify user has access to this seller
	if !ss.hasSellerAccess(userID, sellerID) {
		return nil, errors.New("unauthorized access to seller")
	}

	var seller models.Seller
	if err := ss.DB.First(&seller, sellerID).Error; err != nil {
		return nil, err
	}

	// Get KYC status
	var kycCount int64
	var approvedKycCount int64
	ss.DB.Model(&models.KYC{}).Where("seller_id = ?", sellerID).Count(&kycCount)
	ss.DB.Model(&models.KYC{}).Where("seller_id = ? AND status = 1", sellerID).Count(&approvedKycCount)

	// Get product count
	var productCount int64
	ss.DB.Model(&models.Product{}).Where("seller_id = ?", sellerID).Count(&productCount)

	return &SellerResponse{
		ID:          seller.ID,
		LegalName:   seller.LegalName,
		DisplayName: seller.DisplayName,
		GSTIN:       seller.GSTIN,
		PAN:         seller.PAN,
		Status:      seller.Status,
		RiskScore:   seller.RiskScore,
		CreatedAt:   seller.CreatedAt,
		UpdatedAt:   seller.UpdatedAt,
		Stats: SellerStats{
			KYCDocuments:    int(kycCount),
			ApprovedKYC:     int(approvedKycCount),
			ProductCount:    int(productCount),
			CompletionRate:  ss.calculateCompletionRate(&seller, int(approvedKycCount)),
		},
	}, nil
}

// Update seller profile
func (ss *SellerService) UpdateSeller(sellerID, userID uint, req *UpdateSellerRequest) (*models.Seller, error) {
	// Verify user has access
	if !ss.hasSellerAccess(userID, sellerID) {
		return nil, errors.New("unauthorized access to seller")
	}

	var seller models.Seller
	if err := ss.DB.First(&seller, sellerID).Error; err != nil {
		return nil, err
	}

	// Update fields
	if req.DisplayName != nil {
		seller.DisplayName = *req.DisplayName
	}
	if req.GSTIN != nil {
		seller.GSTIN = *req.GSTIN
	}
	if req.BankRef != nil {
		seller.BankRef = *req.BankRef
	}

	// Save changes
	if err := ss.DB.Save(&seller).Error; err != nil {
		return nil, err
	}

	return &seller, nil
}

// Check if user has access to seller
func (ss *SellerService) hasSellerAccess(userID, sellerID uint) bool {
	var count int64
	ss.DB.Model(&models.SellerUser{}).
		Where("user_id = ? AND seller_id = ? AND status = 1", userID, sellerID).
		Count(&count)
	return count > 0
}

// Calculate risk score based on provided information
func (ss *SellerService) calculateRiskScore(req *CreateSellerRequest) int {
	score := 100 // Start with perfect score

	// Reduce score for missing information
	if req.GSTIN == "" {
		score -= 20
	}
	if req.BankRef == "" {
		score -= 15
	}
	if req.DisplayName == "" {
		score -= 10
	}

	// Basic validation checks
	if len(req.PAN) != 10 {
		score -= 25
	}

	if score < 0 {
		score = 0
	}

	return score
}

// Calculate profile completion rate
func (ss *SellerService) calculateCompletionRate(seller *models.Seller, approvedKYC int) int {
	completion := 0

	// Basic info (40%)
	if seller.LegalName != "" {
		completion += 10
	}
	if seller.DisplayName != "" {
		completion += 10
	}
	if seller.PAN != "" {
		completion += 10
	}
	if seller.GSTIN != "" {
		completion += 10
	}

	// KYC completion (40%)
	if approvedKYC >= 2 { // Assuming minimum 2 KYC docs needed
		completion += 40
	} else if approvedKYC >= 1 {
		completion += 20
	}

	// Bank details (20%)
	if seller.BankRef != "" {
		completion += 20
	}

	return completion
}

// Request/Response DTOs
type CreateSellerRequest struct {
	LegalName   string `json:"legal_name" binding:"required,min=3,max=100"`
	DisplayName string `json:"display_name" binding:"max=50"`
	GSTIN       string `json:"gstin" binding:"omitempty,len=15"`
	PAN         string `json:"pan" binding:"required,len=10"`
	BankRef     string `json:"bank_ref" binding:"omitempty"`
}

type UpdateSellerRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	GSTIN       *string `json:"gstin,omitempty"`
	BankRef     *string `json:"bank_ref,omitempty"`
}

type SellerResponse struct {
	ID          uint        `json:"id"`
	LegalName   string      `json:"legal_name"`
	DisplayName string      `json:"display_name"`
	GSTIN       string      `json:"gstin"`
	PAN         string      `json:"pan"`
	Status      int         `json:"status"`
	RiskScore   int         `json:"risk_score"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Stats       SellerStats `json:"stats"`
}

type SellerStats struct {
	KYCDocuments   int `json:"kyc_documents"`
	ApprovedKYC    int `json:"approved_kyc"`
	ProductCount   int `json:"product_count"`
	CompletionRate int `json:"completion_rate"`
}

