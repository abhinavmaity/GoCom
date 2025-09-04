// internal/marketplace/commerce/dto/cart_dto.go
package dto

import "github.com/shopspring/decimal"

type CreateCartRequest struct {
	Currency string `json:"currency" binding:"required"`
}

type AddCartItemRequest struct {
	SKUID uint `json:"sku_id" binding:"required"`
	Qty   int  `json:"qty" binding:"required,min=1"`
}

type UpdateCartItemRequest struct {
	Qty int `json:"qty" binding:"required,min=1"`
}

type CartResponse struct {
	ID        uint               `json:"id"`
	UserID    uint               `json:"user_id"`
	Currency  string             `json:"currency"`
	Items     []CartItemResponse `json:"items"`
	Total     decimal.Decimal    `json:"total"`
	CreatedAt string             `json:"created_at"`
}

type CartItemResponse struct {
	ID       uint            `json:"id"`
	SKUID    uint            `json:"sku_id"`
	SKUCode  string          `json:"sku_code"`
	Title    string          `json:"title"`
	Price    decimal.Decimal `json:"price"`
	Qty      int             `json:"qty"`
	Subtotal decimal.Decimal `json:"subtotal"`
}
