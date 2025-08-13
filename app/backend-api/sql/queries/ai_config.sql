-- ==================== AI Config ====================

-- name: GetAIConfig :one
SELECT * FROM ai_config WHERE id = 1;

-- name: UpsertAIConfig :one
INSERT INTO ai_config (
    id, primary_provider, primary_api_key_encrypted, primary_model,
    fallback_provider, fallback_api_key_encrypted, fallback_model,
    engine_base_url, engine_api_key_encrypted, max_tokens_per_request, updated_at
) VALUES (
    1, @primary_provider, @primary_api_key_encrypted, @primary_model,
    @fallback_provider, @fallback_api_key_encrypted, @fallback_model,
    @engine_base_url, @engine_api_key_encrypted, @max_tokens_per_request, NOW()
)
ON CONFLICT (id) DO UPDATE SET
    primary_provider           = EXCLUDED.primary_provider,
    primary_api_key_encrypted  = EXCLUDED.primary_api_key_encrypted,
    primary_model              = EXCLUDED.primary_model,
    fallback_provider          = EXCLUDED.fallback_provider,
    fallback_api_key_encrypted = EXCLUDED.fallback_api_key_encrypted,
    fallback_model             = EXCLUDED.fallback_model,
    engine_base_url            = EXCLUDED.engine_base_url,
    engine_api_key_encrypted   = EXCLUDED.engine_api_key_encrypted,
    max_tokens_per_request     = EXCLUDED.max_tokens_per_request,
    updated_at                 = NOW()
RETURNING *;

-- name: InsertAIUsageLog :exec
INSERT INTO ai_usage_log (
    shop_id, feature, provider, model,
    input_tokens, output_tokens, latency_ms, cost_usd_micros, succeeded
) VALUES (
    @shop_id, @feature, @provider, @model,
    @input_tokens, @output_tokens, @latency_ms, @cost_usd_micros, @succeeded
);

-- name: CountAIUsage :one
SELECT COUNT(*) FROM ai_usage_log
WHERE shop_id = @shop_id
  AND feature = @feature
  AND created_at >= @period_start;
