package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"
)

type ReviewService struct {
	db *gorm.DB
}

func NewReviewService(db *gorm.DB) *ReviewService { return &ReviewService{db: db} }

type ReviewRow struct {
	ID        uint           `json:"id"`
	UserID    uint           `json:"user_id"`
	ProductID uint           `json:"product_id"`
	Rating    int            `json:"rating"`
	Text      sql.NullString `json:"text"`
	Media     any            `json:"media"`
	Status    int            `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
}

func (s *ReviewService) ListReviews(ctx context.Context, productID uint, page, limit int) ([]ReviewRow, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	var rows []ReviewRow
	if err := s.db.WithContext(ctx).Table("reviews").
		Where("product_id = ? AND status = 1", productID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Select("id, user_id, product_id, rating, text, media, status, created_at").
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	var total int64
	if err := s.db.WithContext(ctx).Table("reviews").
		Where("product_id = ? AND status = 1", productID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (s *ReviewService) CreateReview(ctx context.Context, userID, productID uint, rating int, text any, media any, autoApprove bool) error {
	if rating < 1 || rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}
	status := 0
	if autoApprove {
		status = 1
	}
	rev := map[string]any{
		"user_id":    userID,
		"product_id": productID,
		"rating":     rating,
		"text":       text,
		"media":      media,
		"status":     status,
		"created_at": time.Now(),
	}
	return s.db.WithContext(ctx).Table("reviews").Create(rev).Error
}

// Recompute aggregates and persist (optional)
func (s *ReviewService) RecomputeAndPersistAggregate(ctx context.Context, productID uint) (float64, int64, error) {
	var agg struct {
		Avg float64 `gorm:"column:avg"`
		Cnt int64   `gorm:"column:count"`
	}
	if err := s.db.WithContext(ctx).Raw(`
		SELECT COALESCE(AVG(rating)::float,0) AS avg, COUNT(*) AS count
		FROM reviews WHERE product_id = ? AND status = 1
	`, productID).Scan(&agg).Error; err != nil {
		return 0, 0, err
	}

	// attempt to update products.rating_avg, products.rating_count if present
	if err := s.db.WithContext(ctx).Exec(`
		UPDATE products SET rating_avg = ?, rating_count = ? WHERE id = ?
	`, agg.Avg, agg.Cnt, productID).Error; err != nil {
		// return values and the error so caller can decide (non-fatal)
		return agg.Avg, agg.Cnt, err
	}
	return agg.Avg, agg.Cnt, nil
}
