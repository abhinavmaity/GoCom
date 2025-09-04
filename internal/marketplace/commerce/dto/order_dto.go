// internal/marketplace/commerce/dto/order_dto.go
package dto

import "github.com/shopspring/decimal"

type CheckoutRequest struct {
	CartID    uint `json:"cart_id" binding:"required"`
	AddressID uint `json:"address_id" binding:"required"`
}

type CheckoutResponse struct {
	OrderID    uint            `json:"order_id"`
	Total      decimal.Decimal `json:"total"`
	Tax        decimal.Decimal `json:"tax"`
	Shipping   decimal.Decimal `json:"shipping"`
	GrandTotal decimal.Decimal `json:"grand_total"`
	Items      []OrderItemDto  `json:"items"`
}

type OrderItemDto struct {
	SKUID    uint            `json:"sku_id"`
	SKUCode  string          `json:"sku_code"`
	Title    string          `json:"title"`
	Qty      int             `json:"qty"`
	Price    decimal.Decimal `json:"price"`
	Tax      decimal.Decimal `json:"tax"`
	SellerID uint            `json:"seller_id"`
}
