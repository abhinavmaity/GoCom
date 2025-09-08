package models

// NOTE: You already have full domain models (Product, SKU, Category, Media, Review).
// These are lightweight DTOs / response shapes used by Discovery services/handlers.

type ProductRow struct {
	ID          uint    `json:"id"`
	Title       string  `json:"title"`
	Brand       string  `json:"brand"`
	PriceMin    float64 `json:"price_min"`
	PriceMax    float64 `json:"price_max"`
	RatingAvg   float64 `json:"rating_avg"`
	RatingCount int64   `json:"rating_count"`
	MediaURL    *string `json:"media_url"`
}

type ProductDetailResp struct {
	ID          uint     `json:"id"`
	Title       string   `json:"title"`
	Brand       string   `json:"brand"`
	Description string   `json:"description"`
	CategoryID  uint     `json:"category_id"`
	Skus        []SKURow `json:"skus"`
	Media       []Media  `json:"media"`
	RatingAvg   float64  `json:"rating_avg"`
	RatingCount int64    `json:"rating_count"`
}

type SKURow struct {
	ID        uint    `json:"id"`
	SKUCode   string  `json:"sku_code"`
	PriceMRP  float64 `json:"price_mrp"`
	PriceSell float64 `json:"price_sell"`
	TaxPct    float64 `json:"tax_pct"`
	Barcode   string  `json:"barcode"`
	StockQty  int64   `json:"stock_qty"`
}

type CategoryRow struct {
	ID       uint   `json:"id"`
	ParentID *uint  `json:"parent_id"`
	Name     string `json:"name"`
	SEOSlug  string `json:"seo_slug"`
}
