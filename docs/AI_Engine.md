# AI Engine

The AI engine is an optional Python service for inference-heavy capabilities.

## Scope

Primary workloads:
- embeddings
- image analysis
- sentiment analysis
- search helpers
- forecasting and prediction routes

Text generation is orchestrated from backend AI modules and may use provider APIs depending on feature/configuration.

## Runtime Characteristics

- Framework: FastAPI
- Startup behavior: model registry loads in application lifespan
- Access model: backend-to-engine service communication
- Auth model: bearer key via `AI_ENGINE_API_KEY`

Important behavior:
- when `AI_ENGINE_API_KEY` is not set, auth check is bypassed for local development
- production deployments should always set the key and keep network access private

## Exposed Endpoints (Service Side)

Health:
- `GET /health`
- `GET /v1/health`

Inference routes are exposed under `/v1/*` via route modules:
- `embeddings`
- `image`
- `sentiment`
- `search`
- `forecast`
- `predict`

Some endpoints return `503` when the requested model slot is not ready.

## Backend Integration

Backend integration points:
- API routes mounted under `/api/v1/ai`
- feature flags gate access by plan
- provider and engine clients initialized from environment config

The backend remains the business-control layer:
- plan checks
- quota checks
- usage accounting
- response shaping for storefront/admin consumers

## Operational Guidance

- keep AI engine network-private
- enforce API key auth in non-dev environments
- cap concurrency and model memory usage
- monitor latency, error rates, and timeout rates

## Related Source

- `app/ai-engine/app/main.py`
- `app/ai-engine/app/auth.py`
- `app/ai-engine/app/routes/`
- `app/backend-api/internal/ai/`
