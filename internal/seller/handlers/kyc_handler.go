package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gocom/main/internal/seller/services"
)

type KYCHandler struct {
	KYCService *services.KYCService
}

func NewKYCHandler() *KYCHandler {
	return &KYCHandler{KYCService: services.NewKYCService()}
}

// Upload KYC document
// POST /sellers/:id/kyc
func (kh *KYCHandler) UploadKYC(c *gin.Context) {
	sellerID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req services.UploadKYCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	kyc, err := kh.KYCService.UploadKYC(uint(sellerID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    kyc,
		"message": "KYC document uploaded successfully",
	})
}

// Get KYC documents
// GET /sellers/:id/kyc
func (kh *KYCHandler) GetKYCDocuments(c *gin.Context) {
	sellerID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	documents, err := kh.KYCService.GetKYCDocuments(uint(sellerID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    documents,
		"count":   len(documents),
	})
}

// Get specific KYC document
// GET /sellers/:id/kyc/:docId
func (kh *KYCHandler) GetKYCDocument(c *gin.Context) {
	sellerID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	docID, _ := strconv.ParseUint(c.Param("docId"), 10, 32)

	document, err := kh.KYCService.GetKYCDocument(uint(sellerID), uint(docID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": document})
}

// Delete KYC document
// DELETE /sellers/:id/kyc/:docId
func (kh *KYCHandler) DeleteKYC(c *gin.Context) {
	sellerID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	docID, _ := strconv.ParseUint(c.Param("docId"), 10, 32)

	if err := kh.KYCService.DeleteKYC(uint(sellerID), uint(docID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "KYC document deleted successfully",
	})
}

