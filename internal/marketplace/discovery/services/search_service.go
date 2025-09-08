package services

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"gocom/main/internal/models"
)

type SearchService struct {
	db *gorm.DB
}

func NewSearchService(db *gorm.DB) *SearchService { return &SearchService{db: db} }

type SearchParams struct {
	Query    string
	Category string
	Brand    string
	MinPrice string
	MaxPrice string
	Sort     string
	Page     int
	PageSize int
}

type SearchResponse struct {
	Products []models.ProductRow `json:"products"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	Total    int64               `json:"total"`
}

func (s *SearchService) Search(ctx context.Context, p SearchParams) (*SearchResponse, error) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}

	q := s.db.WithContext(ctx).Table("products p").
		Select(`
			p.id, p.title, p.brand,
			COALESCE(MIN(CAST(skus.price_sell AS double precision)),0) AS price_min,
			COALESCE(MAX(CAST(skus.price_sell AS double precision)),0) AS price_max,
			COALESCE(avg_rev.rating_avg,0) AS rating_avg,
			COALESCE(avg_rev.rating_cnt,0) AS rating_count,
			(SELECT m.url FROM media m WHERE m.entity_type='product' AND m.entity_id=p.id ORDER BY m.sort ASC LIMIT 1) AS media_url
		`).
		Joins("LEFT JOIN skus ON skus.product_id = p.id").
		Joins(`LEFT JOIN (
			SELECT product_id, AVG(rating)::float AS rating_avg, COUNT(*) AS rating_cnt
			FROM reviews WHERE status=1 GROUP BY product_id
		) avg_rev ON avg_rev.product_id = p.id`).
		Where("p.status = 1")

	if strings.TrimSpace(p.Query) != "" {
		q = q.Where("p.search_vector @@ plainto_tsquery('simple', ?)", p.Query)
	}
	if p.Category != "" {
		q = q.Where("p.category_id = ?", p.Category)
	}
	if p.Brand != "" {
		q = q.Where("p.brand ILIKE ?", "%"+p.Brand+"%")
	}
	if p.MinPrice != "" {
		q = q.Where("CAST(skus.price_sell AS double precision) >= ?", p.MinPrice)
	}
	if p.MaxPrice != "" {
		q = q.Where("CAST(skus.price_sell AS double precision) <= ?", p.MaxPrice)
	}

	q = q.Group("p.id, avg_rev.rating_avg, avg_rev.rating_cnt")

	switch p.Sort {
	case "price_asc":
		q = q.Order("price_min ASC")
	case "price_desc":
		q = q.Order("price_max DESC")
	case "rating":
		q = q.Order("rating_avg DESC, rating_count DESC")
	default:
		if strings.TrimSpace(p.Query) != "" {
			// safe param binding using GORM Expr
			q = q.Order(gorm.Expr("ts_rank(p.search_vector, plainto_tsquery('simple', ?)) DESC", p.Query))
		} else {
			q = q.Order("p.id DESC")
		}
	}

	// count total using subquery
	var total int64
	if err := s.db.WithContext(ctx).Table("(?) as sub", q).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("counting results: %w", err)
	}

	offset := (p.Page - 1) * p.PageSize
	q = q.Offset(offset).Limit(p.PageSize)

	var rows []models.ProductRow
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return &SearchResponse{Products: rows, Page: p.Page, PageSize: p.PageSize, Total: total}, nil
}
