USE gosocial_db;

-- 1) create a seller
INSERT INTO sellers (legal_name, display_name, gstin, pan, bank_ref, status, risk_score, created_at, updated_at)
VALUES ('ACME Private Limited', 'Acme', NULL, NULL, NULL, 1, 0, NOW(), NOW());
SET @seller_id = LAST_INSERT_ID();

-- 2) create a user for reviews (do NOT include two_f_a_enabled column)
INSERT INTO users (name, email, phone, password_hash, status, created_at, updated_at)
VALUES ('Seed User', 'seeduser@example.com', '9999999999', '', 1, NOW(), NOW());
SET @user_id = LAST_INSERT_ID();

-- 3) create a category
INSERT INTO categories (parent_id, name, attributes_schema, seo_slug, created_at)
VALUES (NULL, 'Footwear', NULL, 'footwear', NOW());
SET @category_id = LAST_INSERT_ID();

-- 4) create a product referencing @seller_id and @category_id
INSERT INTO products (seller_id, category_id, title, description, brand, status, score, created_at, updated_at)
VALUES (@seller_id, @category_id, 'Example Running Shoes', 'Lightweight running shoes', 'Acme', 1, 0, NOW(), NOW());
SET @product_id = LAST_INSERT_ID();

-- 5) create a SKU for product
INSERT INTO skus (product_id, sku_code, attributes, price_mrp, price_sell, tax_pct, barcode, created_at)
VALUES (@product_id, 'EX-RT-001', '{"size":"9","color":"black"}', 2999.00, 1999.00, 18.00, '1234567890123', NOW());
SET @sku_id = LAST_INSERT_ID();

-- 6) inventory (table name is `inventories`)
INSERT INTO inventories (sku_id, location_id, on_hand, reserved, threshold, updated_at)
VALUES (@sku_id, 1, 100, 0, 5, NOW());

-- 7) media
INSERT INTO media (entity_type, entity_id, url, type, alt_text, sort, created_at)
VALUES ('product', @product_id, 'https://placehold.co/600x400', 'image', 'Example shoe', 0, NOW());

-- 8) approved review referencing @user_id and @product_id
INSERT INTO reviews (user_id, product_id, rating, text, media, status, created_at)
VALUES (@user_id, @product_id, 5, 'Comfortable and well-priced', NULL, 1, NOW());
