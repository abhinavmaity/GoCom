package services

import (
    "errors"
    "time"
    "gorm.io/gorm"
    "gocom/main/internal/common/db"
    "gocom/main/internal/models"
)

type InventoryService struct {
    DB *gorm.DB
}

func NewInventoryService() *InventoryService {
    return &InventoryService{
        DB: db.GetDB(),
    }
}


func (is *InventoryService) GetInventory(skuID uint) (*InventoryResponse, error) {
    var inventory models.Inventory
    if err := is.DB.Where("sk_uid = ?", skuID).First(&inventory).Error; err != nil {
        return nil, errors.New("inventory not found")
    }

    var sku models.SKU
    is.DB.First(&sku, skuID)

    response := &InventoryResponse{
        ID:         inventory.ID,
        SKUID:      inventory.SKUID,
        SKUCode:    sku.SKUCode,
        LocationID: inventory.LocationID,
        OnHand:     inventory.OnHand,
        Reserved:   inventory.Reserved,
        Available:  inventory.OnHand - inventory.Reserved,
        Threshold:  inventory.Threshold,
        UpdatedAt:  inventory.UpdatedAt,
        IsLowStock: inventory.OnHand <= inventory.Threshold,
    }

    return response, nil
}


func (is *InventoryService) UpdateInventory(skuID uint, req *UpdateInventoryRequest) (*models.Inventory, error) {
    var inventory models.Inventory
    if err := is.DB.Where("sk_uid = ?", skuID).First(&inventory).Error; err != nil {
        return nil, errors.New("inventory not found")
    }

    if req.OnHand != nil {
        if *req.OnHand < 0 {
            return nil, errors.New("on hand quantity cannot be negative")
        }
        inventory.OnHand = *req.OnHand
    }

    if req.Threshold != nil {
        inventory.Threshold = *req.Threshold
    }

    inventory.UpdatedAt = time.Now()
    if err := is.DB.Save(&inventory).Error; err != nil {
        return nil, err
    }

    return &inventory, nil
}


func (is *InventoryService) ReserveInventory(skuID uint, quantity int) error {
    var inventory models.Inventory
    if err := is.DB.Where("sk_uid = ?", skuID).First(&inventory).Error; err != nil {
        return errors.New("inventory not found")
    }

    available := inventory.OnHand - inventory.Reserved
    if available < quantity {
        return errors.New("insufficient inventory")
    }

    inventory.Reserved += quantity
    inventory.UpdatedAt = time.Now()
    return is.DB.Save(&inventory).Error
}

func (is *InventoryService) ReleaseInventory(skuID uint, quantity int) error {
    var inventory models.Inventory
    if err := is.DB.Where("sk_uid = ?", skuID).First(&inventory).Error; err != nil {
        return errors.New("inventory not found")
    }

    inventory.Reserved -= quantity
    if inventory.Reserved < 0 {
        inventory.Reserved = 0
    }

    inventory.UpdatedAt = time.Now()
    return is.DB.Save(&inventory).Error
}

func (is *InventoryService) GetLowStockAlerts(sellerID uint) ([]LowStockAlert, error) {
    var alerts []LowStockAlert
    
    query := `
        SELECT 
            i.id as inventory_id,
            i.sk_uid as sku_id,
            s.sku_code,
            p.title as product_title,
            i.on_hand,
            i.reserved,
            i.threshold,
            (i.on_hand - i.reserved) as available
        FROM inventories i
        JOIN skus s ON i.sk_uid = s.id
        JOIN products p ON s.product_id = p.id  
        WHERE p.seller_id = ? AND i.on_hand <= i.threshold
        ORDER BY i.on_hand ASC
    `

    if err := is.DB.Raw(query, sellerID).Scan(&alerts).Error; err != nil {
        return nil, err
    }

    return alerts, nil
}

func (is *InventoryService) BulkUpdateInventory(updates []BulkInventoryUpdate) error {
    tx := is.DB.Begin()
    
    for _, update := range updates {
        var inventory models.Inventory
        if err := tx.Where("sk_uid = ?", update.SKUID).First(&inventory).Error; err != nil {
            tx.Rollback()
            return errors.New("SKU not found: " + update.SKUCode)
        }

        inventory.OnHand = update.OnHand
        inventory.Threshold = update.Threshold
        inventory.UpdatedAt = time.Now()
        
        if err := tx.Save(&inventory).Error; err != nil {
            tx.Rollback()
            return err
        }
    }

    return tx.Commit().Error
}

func (is *InventoryService) CreateInventoryForSKU(skuID uint, locationID uint, threshold int) error {
    inventory := &models.Inventory{
        SKUID:      skuID,  
        LocationID: locationID,
        OnHand:     0,
        Reserved:   0,
        Threshold:  threshold,
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
    }

    return is.DB.Create(inventory).Error
}


func (is *InventoryService) GetInventoryBySKUs(skuIDs []uint) ([]models.Inventory, error) {
    var inventories []models.Inventory
    if err := is.DB.Where("sk_uid IN ?", skuIDs).Find(&inventories).Error; err != nil {
        return nil, err
    }

    return inventories, nil
}


type UpdateInventoryRequest struct {
    OnHand    *int `json:"on_hand,omitempty"`
    Threshold *int `json:"threshold,omitempty"`
}

type InventoryResponse struct {
    ID         uint      `json:"id"`
    SKUID      uint      `json:"sku_id"`
    SKUCode    string    `json:"sku_code"`
    LocationID uint      `json:"location_id"`
    OnHand     int       `json:"on_hand"`
    Reserved   int       `json:"reserved"`
    Available  int       `json:"available"`
    Threshold  int       `json:"threshold"`
    IsLowStock bool      `json:"is_low_stock"`
    UpdatedAt  time.Time `json:"updated_at"`
}

type LowStockAlert struct {
    InventoryID  uint   `json:"inventory_id"`
    SKUID        uint   `json:"sku_id"`
    SKUCode      string `json:"sku_code"`
    ProductTitle string `json:"product_title"`
    OnHand       int    `json:"on_hand"`
    Reserved     int    `json:"reserved"`
    Available    int    `json:"available"`
    Threshold    int    `json:"threshold"`
}

type BulkInventoryUpdate struct {
    SKUID     uint   `json:"sku_id"`
    SKUCode   string `json:"sku_code"`
    OnHand    int    `json:"on_hand"`
    Threshold int    `json:"threshold"`
}
