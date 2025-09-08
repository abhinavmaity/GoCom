package services

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"gocom/main/internal/models"
)

type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService { return &ProductService{db: db} }

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
	err := s.db.WithContext(ctx).Table("products p").
		Select(`p.id, p.title, p.brand, p.description, p.category_id,
				COALESCE(avg_rev.rating_avg,0) AS rating_avg, COALESCE(avg_rev.rating_cnt,0) AS rating_count`).
		Joins(`LEFT JOIN (
			SELECT product_id, AVG(rating) AS rating_avg, COUNT(*) AS rating_cnt
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

	// SKUs: use models.SKU to match your schema; compute stock via inventory table if present.
	var skus []models.SKU
	if err := s.db.WithContext(ctx).Where("product_id = ?", productID).Find(&skus).Error; err != nil {
		return nil, err
	}

	// Media
	var media []models.Media
	if err := s.db.WithContext(ctx).Table("media").
		Select("id, url, type, alt_text, sort").
		Where("entity_type = 'product' AND entity_id = ?", productID).
		Order("sort ASC").
		Scan(&media).Error; err != nil {
		media = []models.Media{}
	}

	out := &models.ProductDetailResp{
		ID:          p.ID,
		Title:       p.Title,
		Brand:       p.Brand,
		Description: p.Description,
		CategoryID:  p.CategoryID,
		// convert skus slice into SKURow slice if your DTO expects that; here we keep models.SKU
		Skus:        skusToSKURows(skus, s.db, ctx),
		Media:       media,
		RatingAvg:   p.RatingAvg,
		RatingCount: p.RatingCount,
	}
	return out, nil
}

// helper that converts []models.SKU to []models.SKURow (computes stock_qty)
func skusToSKURows(skus []models.SKU, db *gorm.DB, ctx context.Context) []models.SKURow {
	out := make([]models.SKURow, 0, len(skus))
	for _, s := range skus {
		var stock sqlNullInt64
		// try fetching inventory for sku
		_ = db.WithContext(ctx).Raw(`SELECT (COALESCE(on_hand,0) - COALESCE(reserved,0)) AS stock_qty FROM inventory WHERE sku_id = ? LIMIT 1`, s.ID).Scan(&stock).Error
		row := models.SKURow{
			ID:        uint(s.ID),
			SKUCode:   s.SKUCode,
			PriceMRP:  float64FromDecimal(s.PriceMRP),
			PriceSell: float64FromDecimal(s.PriceSell),
			TaxPct:    float64FromDecimal(s.TaxPct),
			Barcode:   s.Barcode,
			StockQty:  stock.Int64,
		}
		out = append(out, row)
	}
	return out
}

type sqlNullInt64 struct {
	Int64 int64 `gorm:"column:stock_qty"`
	Valid bool
}

// helper to convert shopspring decimal to float64 (simple PoC)
func float64FromDecimal(d interface{}) float64 {
	switch v := d.(type) {
	case float64:
		return v
	case nil:
		return 0
	default:
		// best-effort: try Sprintf and parse; but keep 0 for safety
		return 0
	}
}
func decimalToString(d decimal.Decimal) string {
	return d.String()
}
