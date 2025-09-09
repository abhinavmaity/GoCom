#!/bin/bash

# =============================================================================
# Complete Seller API Testing Script
# =============================================================================

set -e  # Exit on any error

# Configuration
BASE_URL="http://localhost:8080"
API_BASE="$BASE_URL/v1"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Global variables to store IDs
ACCESS_TOKEN=""
REFRESH_TOKEN=""
USER_ID=""
SELLER_ID=""
PRODUCT_ID=""
SKU_ID=""
ADDRESS_ID=""
KYC_ID=""
CATEGORY_ID=1  # Assuming category exists

# Helper functions
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}🧪 $1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_test() {
    echo -e "\n${YELLOW}📋 Testing: $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Check if jq is installed
check_dependencies() {
    if ! command -v jq &> /dev/null; then
        print_error "jq is required but not installed. Please install jq first."
        echo "Install with: sudo apt-get install jq (Ubuntu) or brew install jq (Mac)"
        exit 1
    fi
    
    if ! command -v curl &> /dev/null; then
        print_error "curl is required but not installed."
        exit 1
    fi
}

# Test server health
test_health() {
    print_test "Server Health Check"
    response=$(curl -s -w "%{http_code}" -o /tmp/health_response.json "$BASE_URL/health")
    http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        print_success "Server is healthy"
        cat /tmp/health_response.json | jq '.'
    else
        print_error "Server health check failed (HTTP $http_code)"
        exit 1
    fi
}

# =============================================================================
# AUTHENTICATION TESTS
# =============================================================================

test_auth() {
    print_header "AUTHENTICATION TESTS"
    
    # Test 1: User Registration
    print_test "User Registration"
    register_response=$(curl -s -w "%{http_code}" -o /tmp/register_response.json \
        -X POST "$API_BASE/auth/register" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "Test Seller User",
            "email": "seller@test.com", 
            "phone": "9876543210",
            "password": "password123"
        }')
    
    register_http_code="${register_response: -3}"
    if [ "$register_http_code" = "201" ] || [ "$register_http_code" = "409" ]; then
        print_success "User registration successful or user already exists"
        cat /tmp/register_response.json | jq '.'
    else
        print_error "User registration failed (HTTP $register_http_code)"
        cat /tmp/register_response.json
    fi
    
    # Test 2: User Login
    print_test "User Login"
    login_response=$(curl -s -w "%{http_code}" -o /tmp/login_response.json \
        -X POST "$API_BASE/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "seller@test.com",
            "password": "password123"
        }')
    
    login_http_code="${login_response: -3}"
    if [ "$login_http_code" = "200" ]; then
        print_success "User login successful"
        ACCESS_TOKEN=$(cat /tmp/login_response.json | jq -r '.data.access_token')
        REFRESH_TOKEN=$(cat /tmp/login_response.json | jq -r '.data.refresh_token')
        USER_ID=$(cat /tmp/login_response.json | jq -r '.data.user.id')
        print_info "Access Token: ${ACCESS_TOKEN:0:50}..."
        print_info "User ID: $USER_ID"
    else
        print_error "User login failed (HTTP $login_http_code)"
        cat /tmp/login_response.json
        exit 1
    fi
    
    # Test 3: Get User Profile
    print_test "Get User Profile"
    profile_response=$(curl -s -w "%{http_code}" -o /tmp/profile_response.json \
        -X GET "$API_BASE/auth/me" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    profile_http_code="${profile_response: -3}"
    if [ "$profile_http_code" = "200" ]; then
        print_success "User profile retrieved successfully"
        cat /tmp/profile_response.json | jq '.'
    else
        print_error "User profile retrieval failed (HTTP $profile_http_code)"
    fi
    
    # Test 4: Refresh Token
    print_test "Refresh Token"
    refresh_response=$(curl -s -w "%{http_code}" -o /tmp/refresh_response.json \
        -X POST "$API_BASE/auth/refresh" \
        -H "Content-Type: application/json" \
        -d "{\"refresh_token\": \"$REFRESH_TOKEN\"}")
    
    refresh_http_code="${refresh_response: -3}"
    if [ "$refresh_http_code" = "200" ]; then
        print_success "Token refresh successful"
        NEW_TOKEN=$(cat /tmp/refresh_response.json | jq -r '.data.access_token')
        print_info "New Access Token: ${NEW_TOKEN:0:50}..."
    else
        print_error "Token refresh failed (HTTP $refresh_http_code)"
    fi
}

# =============================================================================
# SELLER TESTS
# =============================================================================

test_sellers() {
    print_header "SELLER MANAGEMENT TESTS"
    
    # Test 1: Create Seller
    print_test "Create Seller"
    create_seller_response=$(curl -s -w "%{http_code}" -o /tmp/create_seller_response.json \
        -X POST "$API_BASE/sellers" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "legal_name": "Test Electronics Pvt Ltd",
            "display_name": "Test Electronics", 
            "gstin": "22AAAAA0000A1Z5",
            "pan": "ABCDE1234F",
            "bank_ref": "ICICI001234567890"
        }')
    
    create_seller_http_code="${create_seller_response: -3}"
    if [ "$create_seller_http_code" = "201" ]; then
        print_success "Seller created successfully"
        SELLER_ID=$(cat /tmp/create_seller_response.json | jq -r '.data.id')
        print_info "Seller ID: $SELLER_ID"
        cat /tmp/create_seller_response.json | jq '.'
    else
        print_error "Seller creation failed (HTTP $create_seller_http_code)"
        cat /tmp/create_seller_response.json
        # Try to continue with existing seller if creation failed due to duplicate
        if [ "$create_seller_http_code" = "400" ]; then
            SELLER_ID=1
            print_info "Using default Seller ID: $SELLER_ID"
        fi
    fi
    
    # Test 2: Get Seller Profile
    print_test "Get Seller Profile"
    get_seller_response=$(curl -s -w "%{http_code}" -o /tmp/get_seller_response.json \
        -X GET "$API_BASE/sellers/$SELLER_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    get_seller_http_code="${get_seller_response: -3}"
    if [ "$get_seller_http_code" = "200" ]; then
        print_success "Seller profile retrieved successfully"
        cat /tmp/get_seller_response.json | jq '.'
    else
        print_error "Seller profile retrieval failed (HTTP $get_seller_http_code)"
    fi
    
    # Test 3: Update Seller
    print_test "Update Seller"
    update_seller_response=$(curl -s -w "%{http_code}" -o /tmp/update_seller_response.json \
        -X PATCH "$API_BASE/sellers/$SELLER_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "display_name": "Updated Electronics Store",
            "bank_ref": "HDFC001234567890"
        }')
    
    update_seller_http_code="${update_seller_response: -3}"
    if [ "$update_seller_http_code" = "200" ]; then
        print_success "Seller updated successfully"
        cat /tmp/update_seller_response.json | jq '.'
    else
        print_error "Seller update failed (HTTP $update_seller_http_code)"
    fi
}

# =============================================================================
# KYC TESTS  
# =============================================================================

test_kyc() {
    print_header "KYC MANAGEMENT TESTS"
    
    # Test 1: Upload KYC Document
    print_test "Upload KYC Document"
    upload_kyc_response=$(curl -s -w "%{http_code}" -o /tmp/upload_kyc_response.json \
        -X POST "$API_BASE/sellers/$SELLER_ID/kyc" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "type": "PAN",
            "document_url": "/documents/pan_document.pdf"
        }')
    
    upload_kyc_http_code="${upload_kyc_response: -3}"
    if [ "$upload_kyc_http_code" = "201" ]; then
        print_success "KYC document uploaded successfully"
        KYC_ID=$(cat /tmp/upload_kyc_response.json | jq -r '.data.id')
        print_info "KYC ID: $KYC_ID"
        cat /tmp/upload_kyc_response.json | jq '.'
    else
        print_error "KYC document upload failed (HTTP $upload_kyc_http_code)"
        cat /tmp/upload_kyc_response.json
    fi
    
    # Test 2: Upload Another KYC Document
    print_test "Upload GSTIN KYC Document" 
    curl -s -w "%{http_code}" -o /tmp/upload_gstin_response.json \
        -X POST "$API_BASE/sellers/$SELLER_ID/kyc" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "type": "GSTIN", 
            "document_url": "/documents/gstin_document.pdf"
        }' > /dev/null
    
    # Test 3: Get All KYC Documents
    print_test "Get All KYC Documents"
    get_kyc_response=$(curl -s -w "%{http_code}" -o /tmp/get_kyc_response.json \
        -X GET "$API_BASE/sellers/$SELLER_ID/kyc" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    get_kyc_http_code="${get_kyc_response: -3}"
    if [ "$get_kyc_http_code" = "200" ]; then
        print_success "KYC documents retrieved successfully"
        cat /tmp/get_kyc_response.json | jq '.'
    else
        print_error "KYC documents retrieval failed (HTTP $get_kyc_http_code)"
    fi
    
    # Test 4: Get Specific KYC Document
    if [ -n "$KYC_ID" ] && [ "$KYC_ID" != "null" ]; then
        print_test "Get Specific KYC Document"
        get_kyc_doc_response=$(curl -s -w "%{http_code}" -o /tmp/get_kyc_doc_response.json \
            -X GET "$API_BASE/sellers/$SELLER_ID/kyc/$KYC_ID" \
            -H "Authorization: Bearer $ACCESS_TOKEN")
        
        get_kyc_doc_http_code="${get_kyc_doc_response: -3}"
        if [ "$get_kyc_doc_http_code" = "200" ]; then
            print_success "Specific KYC document retrieved successfully"
            cat /tmp/get_kyc_doc_response.json | jq '.'
        else
            print_error "Specific KYC document retrieval failed (HTTP $get_kyc_doc_http_code)"
        fi
    fi
}

# =============================================================================
# PRODUCT TESTS
# =============================================================================

test_products() {
    print_header "PRODUCT MANAGEMENT TESTS"
    
    # First, let's create a category if needed
    print_test "Setup: Create Category (Direct DB)"
    echo "Note: You may need to manually create categories in your database first"
    
    # Test 1: Create Product
    print_test "Create Product"
    create_product_response=$(curl -s -w "%{http_code}" -o /tmp/create_product_response.json \
        -X POST "$API_BASE/sellers/$SELLER_ID/products" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "category_id": 1,
            "title": "iPhone 15 Pro Max",
            "description": "Latest Apple iPhone with advanced camera system and A17 Pro chip. Features titanium design and Action button.",
            "brand": "Apple"
        }')
    
    create_product_http_code="${create_product_response: -3}"
    if [ "$create_product_http_code" = "201" ]; then
        print_success "Product created successfully"
        PRODUCT_ID=$(cat /tmp/create_product_response.json | jq -r '.data.id')
        print_info "Product ID: $PRODUCT_ID"
        cat /tmp/create_product_response.json | jq '.'
    else
        print_error "Product creation failed (HTTP $create_product_http_code)"
        cat /tmp/create_product_response.json
        # Set a default product ID to continue testing
        PRODUCT_ID=1
    fi
    
    # Test 2: List Seller Products
    print_test "List Seller Products"
    list_products_response=$(curl -s -w "%{http_code}" -o /tmp/list_products_response.json \
        -X GET "$API_BASE/sellers/$SELLER_ID/products" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    list_products_http_code="${list_products_response: -3}"
    if [ "$list_products_http_code" = "200" ]; then
        print_success "Products listed successfully"
        cat /tmp/list_products_response.json | jq '.'
    else
        print_error "Products listing failed (HTTP $list_products_http_code)"
    fi
    
    # Test 3: Get Product Details
    print_test "Get Product Details"
    get_product_response=$(curl -s -w "%{http_code}" -o /tmp/get_product_response.json \
        -X GET "$API_BASE/products/$PRODUCT_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    get_product_http_code="${get_product_response: -3}"
    if [ "$get_product_http_code" = "200" ]; then
        print_success "Product details retrieved successfully"
        cat /tmp/get_product_response.json | jq '.'
    else
        print_error "Product details retrieval failed (HTTP $get_product_http_code)"
    fi
    
    # Test 4: Update Product
    print_test "Update Product"
    update_product_response=$(curl -s -w "%{http_code}" -o /tmp/update_product_response.json \
        -X PATCH "$API_BASE/products/$PRODUCT_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "title": "iPhone 15 Pro Max - Updated",
            "description": "Updated description with more features and specifications."
        }')
    
    update_product_http_code="${update_product_response: -3}"
    if [ "$update_product_http_code" = "200" ]; then
        print_success "Product updated successfully"
        cat /tmp/update_product_response.json | jq '.'
    else
        print_error "Product update failed (HTTP $update_product_http_code)"
    fi
    
    # Test 5: Search Products with Filters
    print_test "Search Products with Filters"
    search_products_response=$(curl -s -w "%{http_code}" -o /tmp/search_products_response.json \
        -X GET "$API_BASE/sellers/$SELLER_ID/products?search=iPhone&status=0" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    search_products_http_code="${search_products_response: -3}"
    if [ "$search_products_http_code" = "200" ]; then
        print_success "Product search completed successfully"
        cat /tmp/search_products_response.json | jq '.'
    else
        print_error "Product search failed (HTTP $search_products_http_code)"
    fi
}

# =============================================================================
# SKU TESTS
# =============================================================================

test_skus() {
    print_header "SKU MANAGEMENT TESTS"
    
    # Test 1: Create SKU
    print_test "Create SKU"
    create_sku_response=$(curl -s -w "%{http_code}" -o /tmp/create_sku_response.json \
        -X POST "$API_BASE/products/$PRODUCT_ID/skus" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "sku_code": "IPHONE15-256GB-BLUE",
            "attributes": {
                "color": "Deep Blue",
                "storage": "256GB", 
                "model": "A3108"
            },
            "price_mrp": "134900.00",
            "price_sell": "129900.00", 
            "tax_pct": "18.0",
            "barcode": "194253001645",
            "threshold": 10
        }')
    
    create_sku_http_code="${create_sku_response: -3}"
    if [ "$create_sku_http_code" = "201" ]; then
        print_success "SKU created successfully"
        SKU_ID=$(cat /tmp/create_sku_response.json | jq -r '.data.id')
        print_info "SKU ID: $SKU_ID"
        cat /tmp/create_sku_response.json | jq '.'
    else
        print_error "SKU creation failed (HTTP $create_sku_http_code)"
        cat /tmp/create_sku_response.json
        SKU_ID=1
    fi
    
    # Test 2: Create Another SKU
    print_test "Create Another SKU (512GB)"
    curl -s -w "%{http_code}" -o /tmp/create_sku2_response.json \
        -X POST "$API_BASE/products/$PRODUCT_ID/skus" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "sku_code": "IPHONE15-512GB-BLUE",
            "attributes": {
                "color": "Deep Blue",
                "storage": "512GB",
                "model": "A3108"
            },
            "price_mrp": "154900.00",
            "price_sell": "149900.00",
            "tax_pct": "18.0", 
            "barcode": "194253001646",
            "threshold": 5
        }' > /dev/null
    
    # Test 3: Get Product SKUs
    print_test "Get Product SKUs"
    get_skus_response=$(curl -s -w "%{http_code}" -o /tmp/get_skus_response.json \
        -X GET "$API_BASE/products/$PRODUCT_ID/skus" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    get_skus_http_code="${get_skus_response: -3}"
    if [ "$get_skus_http_code" = "200" ]; then
        print_success "Product SKUs retrieved successfully"
        cat /tmp/get_skus_response.json | jq '.'
    else
        print_error "Product SKUs retrieval failed (HTTP $get_skus_http_code)"
    fi
    
    # Test 4: Update SKU
    print_test "Update SKU"
    update_sku_response=$(curl -s -w "%{http_code}" -o /tmp/update_sku_response.json \
        -X PATCH "$API_BASE/skus/$SKU_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "price_sell": "127900.00",
            "barcode": "194253001645-UPDATED"
        }')
    
    update_sku_http_code="${update_sku_response: -3}"
    if [ "$update_sku_http_code" = "200" ]; then
        print_success "SKU updated successfully"
        cat /tmp/update_sku_response.json | jq '.'
    else
        print_error "SKU update failed (HTTP $update_sku_http_code)"
    fi
}

# =============================================================================
# INVENTORY TESTS
# =============================================================================

test_inventory() {
    print_header "INVENTORY MANAGEMENT TESTS"
    
    # Test 1: Get SKU Inventory
    print_test "Get SKU Inventory"
    get_inventory_response=$(curl -s -w "%{http_code}" -o /tmp/get_inventory_response.json \
        -X GET "$API_BASE/skus/$SKU_ID/inventory" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    get_inventory_http_code="${get_inventory_response: -3}"
    if [ "$get_inventory_http_code" = "200" ]; then
        print_success "SKU inventory retrieved successfully"
        cat /tmp/get_inventory_response.json | jq '.'
    else
        print_error "SKU inventory retrieval failed (HTTP $get_inventory_http_code)"
    fi
    
    # Test 2: Update Inventory
    print_test "Update Inventory"
    update_inventory_response=$(curl -s -w "%{http_code}" -o /tmp/update_inventory_response.json \
        -X PATCH "$API_BASE/skus/$SKU_ID/inventory" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "on_hand": 50,
            "threshold": 5
        }')
    
    update_inventory_http_code="${update_inventory_response: -3}"
    if [ "$update_inventory_http_code" = "200" ]; then
        print_success "Inventory updated successfully"
        cat /tmp/update_inventory_response.json | jq '.'
    else
        print_error "Inventory update failed (HTTP $update_inventory_http_code)"
    fi
    
    # Test 3: Get Low Stock Alerts
    print_test "Get Low Stock Alerts"
    get_alerts_response=$(curl -s -w "%{http_code}" -o /tmp/get_alerts_response.json \
        -X GET "$API_BASE/sellers/$SELLER_ID/inventory/alerts" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    get_alerts_http_code="${get_alerts_response: -3}"
    if [ "$get_alerts_http_code" = "200" ]; then
        print_success "Low stock alerts retrieved successfully"
        cat /tmp/get_alerts_response.json | jq '.'
    else
        print_error "Low stock alerts retrieval failed (HTTP $get_alerts_http_code)"
    fi
    
    # Test 4: Bulk Inventory Update
    print_test "Bulk Inventory Update"
    bulk_update_response=$(curl -s -w "%{http_code}" -o /tmp/bulk_update_response.json \
        -X POST "$API_BASE/inventory/bulk-update" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d "[
            {
                \"sku_id\": $SKU_ID,
                \"sku_code\": \"IPHONE15-256GB-BLUE\",
                \"on_hand\": 100,
                \"threshold\": 10
            }
        ]")
    
    bulk_update_http_code="${bulk_update_response: -3}"
    if [ "$bulk_update_http_code" = "200" ]; then
        print_success "Bulk inventory update successful"
        cat /tmp/bulk_update_response.json | jq '.'
    else
        print_error "Bulk inventory update failed (HTTP $bulk_update_http_code)"
    fi
}

# =============================================================================
# ADDRESS TESTS
# =============================================================================

test_addresses() {
    print_header "ADDRESS MANAGEMENT TESTS"
    
    # Test 1: Add Seller Address
    print_test "Add Seller Address"
    add_address_response=$(curl -s -w "%{http_code}" -o /tmp/add_address_response.json \
        -X POST "$API_BASE/sellers/$SELLER_ID/addresses" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "line1": "123, Tech Park",
            "line2": "Phase 1, Whitefield",
            "city": "Bangalore",
            "state": "Karnataka", 
            "country": "India",
            "pin": "560066"
        }')
    
    add_address_http_code="${add_address_response: -3}"
    if [ "$add_address_http_code" = "201" ]; then
        print_success "Address added successfully"
        ADDRESS_ID=$(cat /tmp/add_address_response.json | jq -r '.data.id')
        print_info "Address ID: $ADDRESS_ID"
        cat /tmp/add_address_response.json | jq '.'
    else
        print_error "Address addition failed (HTTP $add_address_http_code)"
        cat /tmp/add_address_response.json
        ADDRESS_ID=1
    fi
    
    # Test 2: Get Seller Addresses
    print_test "Get Seller Addresses"
    get_addresses_response=$(curl -s -w "%{http_code}" -o /tmp/get_addresses_response.json \
        -X GET "$API_BASE/sellers/$SELLER_ID/addresses" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    get_addresses_http_code="${get_addresses_response: -3}"
    if [ "$get_addresses_http_code" = "200" ]; then
        print_success "Seller addresses retrieved successfully"
        cat /tmp/get_addresses_response.json | jq '.'
    else
        print_error "Seller addresses retrieval failed (HTTP $get_addresses_http_code)"
    fi
    
    # Test 3: Update Address
    print_test "Update Address"
    update_address_response=$(curl -s -w "%{http_code}" -o /tmp/update_address_response.json \
        -X PATCH "$API_BASE/addresses/$ADDRESS_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "line2": "Phase 2, Whitefield - Updated",
            "pin": "560067"
        }')
    
    update_address_http_code="${update_address_response: -3}"
    if [ "$update_address_http_code" = "200" ]; then
        print_success "Address updated successfully"
        cat /tmp/update_address_response.json | jq '.'
    else
        print_error "Address update failed (HTTP $update_address_http_code)"
    fi
}

# =============================================================================
# PRODUCT PUBLISHING TEST
# =============================================================================

test_product_publishing() {
    print_header "PRODUCT PUBLISHING TESTS"
    
    # Test 1: Attempt to Publish Product (should fail due to missing requirements)
    print_test "Attempt to Publish Product (Expected to Fail)"
    publish_response=$(curl -s -w "%{http_code}" -o /tmp/publish_response.json \
        -X POST "$API_BASE/products/$PRODUCT_ID/publish" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    publish_http_code="${publish_response: -3}"
    if [ "$publish_http_code" = "200" ]; then
        print_success "Product published successfully"
        cat /tmp/publish_response.json | jq '.'
    else
        print_error "Product publishing failed as expected (HTTP $publish_http_code)"
        cat /tmp/publish_response.json | jq '.'
    fi
}

# =============================================================================
# ERROR HANDLING TESTS
# =============================================================================

test_error_cases() {
    print_header "ERROR HANDLING TESTS"
    
    # Test 1: Invalid JWT Token
    print_test "Test Invalid JWT Token"
    invalid_token_response=$(curl -s -w "%{http_code}" -o /tmp/invalid_token_response.json \
        -X GET "$API_BASE/sellers/$SELLER_ID" \
        -H "Authorization: Bearer invalid_token_here")
    
    invalid_token_http_code="${invalid_token_response: -3}"
    if [ "$invalid_token_http_code" = "401" ]; then
        print_success "Invalid JWT token properly rejected"
    else
        print_error "Invalid JWT token not handled correctly"
    fi
    
    # Test 2: Access Non-existent Resource
    print_test "Test Non-existent Resource"
    nonexistent_response=$(curl -s -w "%{http_code}" -o /tmp/nonexistent_response.json \
        -X GET "$API_BASE/sellers/99999" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    nonexistent_http_code="${nonexistent_response: -3}"
    if [ "$nonexistent_http_code" = "404" ] || [ "$nonexistent_http_code" = "403" ]; then
        print_success "Non-existent resource properly handled"
    else
        print_error "Non-existent resource not handled correctly"
    fi
    
    # Test 3: Invalid JSON Body
    print_test "Test Invalid JSON Body"
    invalid_json_response=$(curl -s -w "%{http_code}" -o /tmp/invalid_json_response.json \
        -X POST "$API_BASE/sellers/$SELLER_ID/kyc" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{invalid_json}')
    
    invalid_json_http_code="${invalid_json_response: -3}"
    if [ "$invalid_json_http_code" = "400" ]; then
        print_success "Invalid JSON properly rejected"
    else
        print_error "Invalid JSON not handled correctly"
    fi
    
    # Test 4: Missing Required Fields
    print_test "Test Missing Required Fields"
    missing_fields_response=$(curl -s -w "%{http_code}" -o /tmp/missing_fields_response.json \
        -X POST "$API_BASE/sellers/$SELLER_ID/kyc" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"type": ""}')
    
    missing_fields_http_code="${missing_fields_response: -3}"
    if [ "$missing_fields_http_code" = "400" ]; then
        print_success "Missing required fields properly handled"
    else
        print_error "Missing required fields not handled correctly"
    fi
}

# =============================================================================
# CLEANUP TESTS (Optional)
# =============================================================================

test_cleanup() {
    print_header "CLEANUP TESTS (OPTIONAL)"
    
    # Test 1: Delete Address (if ID exists)
    if [ -n "$ADDRESS_ID" ] && [ "$ADDRESS_ID" != "null" ]; then
        print_test "Delete Address"
        delete_address_response=$(curl -s -w "%{http_code}" -o /tmp/delete_address_response.json \
            -X DELETE "$API_BASE/addresses/$ADDRESS_ID" \
            -H "Authorization: Bearer $ACCESS_TOKEN")
        
        delete_address_http_code="${delete_address_response: -3}"
        if [ "$delete_address_http_code" = "200" ]; then
            print_success "Address deleted successfully"
        else
            print_error "Address deletion failed (HTTP $delete_address_http_code)"
        fi
    fi
    
    # Test 2: Delete SKU (if ID exists)
    if [ -n "$SKU_ID" ] && [ "$SKU_ID" != "null" ]; then
        print_test "Delete SKU"
        delete_sku_response=$(curl -s -w "%{http_code}" -o /tmp/delete_sku_response.json \
            -X DELETE "$API_BASE/skus/$SKU_ID" \
            -H "Authorization: Bearer $ACCESS_TOKEN")
        
        delete_sku_http_code="${delete_sku_response: -3}"
        if [ "$delete_sku_http_code" = "200" ]; then
            print_success "SKU deleted successfully"
        else
            print_error "SKU deletion failed (HTTP $delete_sku_http_code)"
        fi
    fi
    
    # Test 3: Delete KYC Document (if ID exists)  
    if [ -n "$KYC_ID" ] && [ "$KYC_ID" != "null" ]; then
        print_test "Delete KYC Document"
        delete_kyc_response=$(curl -s -w "%{http_code}" -o /tmp/delete_kyc_response.json \
            -X DELETE "$API_BASE/sellers/$SELLER_ID/kyc/$KYC_ID" \
            -H "Authorization: Bearer $ACCESS_TOKEN")
        
        delete_kyc_http_code="${delete_kyc_response: -3}"
        if [ "$delete_kyc_http_code" = "200" ]; then
            print_success "KYC document deleted successfully"
        else
            print_error "KYC document deletion failed (HTTP $delete_kyc_http_code)"
        fi
    fi
    
    # Test 4: Delete Product (if ID exists)
    if [ -n "$PRODUCT_ID" ] && [ "$PRODUCT_ID" != "null" ]; then
        print_test "Delete Product"
        delete_product_response=$(curl -s -w "%{http_code}" -o /tmp/delete_product_response.json \
            -X DELETE "$API_BASE/products/$PRODUCT_ID" \
            -H "Authorization: Bearer $ACCESS_TOKEN")
        
        delete_product_http_code="${delete_product_response: -3}"
        if [ "$delete_product_http_code" = "200" ]; then
            print_success "Product deleted successfully"
        else
            print_error "Product deletion failed (HTTP $delete_product_http_code)"
        fi
    fi
}

# =============================================================================
# TEST SUMMARY
# =============================================================================

generate_summary() {
    print_header "TEST SUMMARY REPORT"
    
    echo -e "\n${BLUE}📊 Test Execution Summary${NC}"
    echo -e "${BLUE}=========================${NC}"
    
    # Count temp files to get rough idea of tests run
    test_count=$(ls /tmp/*_response.json 2>/dev/null | wc -l)
    echo -e "${GREEN}Total API calls made: $test_count${NC}"
    
    echo -e "\n${BLUE}🔑 Generated Test Data:${NC}"
    echo -e "User ID: $USER_ID"
    echo -e "Seller ID: $SELLER_ID" 
    echo -e "Product ID: $PRODUCT_ID"
    echo -e "SKU ID: $SKU_ID"
    echo -e "Address ID: $ADDRESS_ID"
    echo -e "KYC ID: $KYC_ID"
    
    echo -e "\n${BLUE}📁 Response Files:${NC}"
    echo -e "All response files are saved in /tmp/ with *_response.json pattern"
    echo -e "You can inspect them for detailed API responses"
    
    echo -e "\n${GREEN}✅ API Testing Complete!${NC}"
    echo -e "Check the output above for any failed tests (marked with ❌)"
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    print_header "🚀 COMPLETE SELLER API TESTING SUITE"
    echo -e "${BLUE}Starting comprehensive API testing...${NC}\n"
    
    # Check dependencies
    check_dependencies
    
    # Run all test suites
    test_health
    test_auth
    test_sellers
    test_kyc
    test_products
    test_skus
    test_inventory
    test_addresses
    test_product_publishing
    test_error_cases
    
    # Optional cleanup
    read -p "Do you want to run cleanup tests (delete created data)? [y/N]: " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        test_cleanup
    fi
    
    # Generate summary
    generate_summary
}

# Run the main function
main "$@"
