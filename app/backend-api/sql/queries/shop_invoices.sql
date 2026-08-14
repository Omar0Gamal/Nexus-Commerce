-- ==================== Shop Invoices ====================

-- name: CreateShopInvoice :one
INSERT INTO shop_invoices (
    shop_id, amount, status, billing_reason, hosted_invoice_url
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetShopInvoice :one
SELECT * FROM shop_invoices WHERE id = $1;

-- name: ListShopInvoices :many
SELECT * FROM shop_invoices
WHERE shop_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListShopInvoicesByStatus :many
SELECT * FROM shop_invoices
WHERE shop_id = $1 AND status = $2
ORDER BY created_at DESC;

-- name: UpdateShopInvoiceStatus :one
UPDATE shop_invoices SET status = $2 WHERE id = $1
RETURNING *;

-- name: DeleteShopInvoice :exec
DELETE FROM shop_invoices WHERE id = $1;

-- name: CountShopInvoices :one
SELECT COUNT(*) FROM shop_invoices WHERE shop_id = $1;
