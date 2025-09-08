package services

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"gocom/main/internal/models"
)

// ProductService returns product details, skus and media.
type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService { return &ProductService{db: db} }

// GetProductDetail returns product basic fields and rating aggregates.
func (s *ProductService) GetProductDetail(ctx context.Context, productID uint) (*models.ProductDetailResp, error) {
	var p struct {
		ID          uint
		Title       string
		Brand       string
		Description string
		CategoryID  uint
		RatingAvg   float64
		RatingCount int64
	}
	// aggregate rating via subquery join
	err := s.db.WithContext(ctx).Table("products p").
		Select(`p.id, p.title, p.brand, p.description, p.category_id,
		        COALESCE(avg_rev.rating_avg,0) AS rating_avg, COALESCE(avg_rev.rating_cnt,0) AS rating_count`).
		Joins(`LEFT JOIN (
			SELECT product_id, AVG(rating)::float AS rating_avg, COUNT(*) AS rating_cnt
			FROM reviews WHERE status = 1 GROUP BY product_id
		) avg_rev ON avg_rev.product_id = p.id`).
		Where("p.id = ? AND p.status = 1", productID).
		Scan(&p).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product %d not found", productID)
		}
		return nil, err
	}

	// fetch SKUs with stock (cast numeric to double precision)
	var skus []models.SKURow
	err = s.db.WithContext(ctx).Raw(`
		SELECT s.id, s.sku_code,
		       COALESCE(ROUND(CAST(s.price_mrp AS double precision),2),0)::double precision AS price_mrp,
		       COALESCE(ROUND(CAST(s.price_sell AS double precision),2),0)::double precision AS price_sell,
		       COALESCE(CAST(s.tax_pct AS double precision),0) AS tax_pct,
		       COALESCE(s.barcode,'') AS barcode,
		       COALESCE(inv.on_hand,0) - COALESCE(inv.reserved,0) AS stock_qty
		FROM skus s
		LEFT JOIN inventory inv ON inv.sku_id = s.id
		WHERE s.product_id = ?
		ORDER BY s.id ASC
	`, productID).Scan(&skus).Error
	if err != nil {
		return nil, err
	}

	// fetch media
	var media []models.Media
	if err := s.db.WithContext(ctx).Table("media").
		Select("id, url, type, alt_text, sort").
		Where("entity_type = 'product' AND entity_id = ?", productID).
		Order("sort ASC").
		Scan(&media).Error; err != nil {
		// non-fatal: return product but with empty media
		media = []models.Media{}
	}

	out := &models.ProductDetailResp{
		ID:          p.ID,
		Title:       p.Title,
		Brand:       p.Brand,
		Description: p.Description,
		CategoryID:  p.CategoryID,
		Skus:        skus,
		Media:       media,
		RatingAvg:   p.RatingAvg,
		RatingCount: p.RatingCount,
	}
	return out, nil
}
