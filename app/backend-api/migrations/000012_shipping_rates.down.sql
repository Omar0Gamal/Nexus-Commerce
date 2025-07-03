ALTER TABLE orders DROP COLUMN IF EXISTS shipping_fee;
DROP TABLE IF EXISTS shipping_rates;
DROP TYPE IF EXISTS shipping_rate_type;
