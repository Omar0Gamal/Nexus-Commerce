-- name: CreateConsentLog :one
INSERT INTO consent_log (customer_id, consent_type, ip_address, user_agent, terms_version)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListConsentsByCustomer :many
SELECT * FROM consent_log
WHERE customer_id = $1
ORDER BY consented_at DESC;
