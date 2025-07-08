ALTER TABLE products
    DROP COLUMN IF EXISTS stock_quantity,
    DROP COLUMN IF EXISTS low_stock_threshold;
