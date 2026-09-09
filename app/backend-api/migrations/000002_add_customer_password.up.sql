-- =============================================
-- Migration 002: Add password_hash to customers
-- Enables customer (shopper) authentication per store
-- =============================================

ALTER TABLE customers ADD COLUMN password_hash VARCHAR(255);
