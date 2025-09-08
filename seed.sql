CREATE DATABASE IF NOT EXISTS gosocial_db CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE gosocial_db;

INSERT INTO categories (parent_id, name, attributes_schema, seo_slug, created_at)
VALUES (NULL, 'Footwear', NULL, 'footwear', NOW());

INSERT INTO products (seller_id, category_id, title, description, brand, status, created_at, updated_at)
VALUES (1, 1, 'Example Running Shoes', 'Lightweight running shoes', 'Acme', 1, NOW(), NOW());

INSERT INTO skus (product_id, sku_code, attributes, price_mrp, price_sell, tax_pct, barcode, created_at)
VALUES (1, 'EX-RT-001', '{"size":"9","color":"black"}', 2999.00, 1999.00, 18.00, '1234567890123', NOW());

INSERT INTO inventory (sku_id, location_id, on_hand, reserved, threshold, updated_at)
VALUES (1, 1, 100, 0, 5, NOW());

INSERT INTO media (entity_type, entity_id, url, type, alt_text, sort, created_at)
VALUES ('product', 1, 'https://placehold.co/600x400', 'image', 'Example shoe', 0, NOW());

INSERT INTO reviews (user_id, product_id, rating, text, media, status, created_at)
VALUES (1, 1, 4, 'Comfortable and well-priced', NULL, 1, NOW());
