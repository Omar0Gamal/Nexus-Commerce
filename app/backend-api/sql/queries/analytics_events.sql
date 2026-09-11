-- name: InsertAnalyticsEvent :one
INSERT INTO analytics_events (
    shop_id, session_id, customer_id, event_type, url_path, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetDynamicDropoffs :many
WITH SessionStats AS (
    SELECT 
        session_id,
        MAX((metadata->>'scroll_percentage')::numeric) as max_scroll_depth,
        SUM((metadata->>'dwell_time_ms')::numeric) as total_dwell_time_ms,
        bool_or(event_type = 'order_completed') as converted
    FROM analytics_events
    WHERE analytics_events.shop_id = sqlc.arg('shop_id') AND analytics_events.created_at >= sqlc.arg('created_at_start') AND analytics_events.created_at <= sqlc.arg('created_at_end')
    GROUP BY session_id
),
SessionLastEvent AS (
    SELECT DISTINCT ON (session_id)
        session_id,
        event_type as final_event,
        metadata as final_metadata
    FROM analytics_events
    WHERE analytics_events.shop_id = sqlc.arg('shop_id') AND analytics_events.created_at >= sqlc.arg('created_at_start') AND analytics_events.created_at <= sqlc.arg('created_at_end')
    ORDER BY session_id, created_at DESC
)
SELECT 
    l.final_event as dropoff_stage,
    l.final_metadata::text as context,
    AVG(s.max_scroll_depth)::numeric as avg_scroll_depth,
    AVG(s.total_dwell_time_ms)::numeric as avg_dwell_time_ms,
    COUNT(*)::bigint as count
FROM SessionStats s
JOIN SessionLastEvent l ON s.session_id = l.session_id
WHERE s.converted = false
GROUP BY l.final_event, l.final_metadata::text
HAVING COUNT(*) >= 3 -- Only return buckets with multiple occurrences to filter noise
ORDER BY count DESC
LIMIT 20;
