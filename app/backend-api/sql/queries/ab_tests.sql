-- name: CreateABTest :one
INSERT INTO ab_tests (shop_id, name, description, variant_a, variant_b, metric, traffic_split)
VALUES (@shop_id, @name, @description, @variant_a, @variant_b, @metric, @traffic_split)
RETURNING *;

-- name: GetABTest :one
SELECT * FROM ab_tests WHERE id = @id AND shop_id = @shop_id;

-- name: ListActiveABTests :many
SELECT * FROM ab_tests WHERE shop_id = @shop_id AND status = 'active';

-- name: ListABTestsByShop :many
SELECT * FROM ab_tests WHERE shop_id = @shop_id ORDER BY created_at DESC;

-- name: ConcludeABTest :exec
UPDATE ab_tests SET status = 'concluded', concluded_at = NOW()
WHERE id = @id AND shop_id = @shop_id;

-- name: ListABTestsForAutoConclusion :many
SELECT * FROM ab_tests
WHERE status = 'active'
  AND created_at < NOW() - INTERVAL '14 days';
