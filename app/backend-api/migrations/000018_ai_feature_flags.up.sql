-- Add AI feature keys to the existing plans.
-- Basic: no AI features.
-- Professional: 50 text generations/month, 20 image analyses/month.
-- Premium: 500 text generations/month, 200 image analyses/month, unlimited embeddings.
UPDATE plans SET features = features ||
    '{"ai_text_generation": false, "ai_image_analysis": false, "ai_embeddings": false}'::jsonb
WHERE name = 'basic';

UPDATE plans SET features = features ||
    '{"ai_text_generation": true, "ai_image_analysis": true, "ai_embeddings": false,
      "ai_monthly_text_quota": 50, "ai_monthly_image_quota": 20}'::jsonb
WHERE name = 'professional';

UPDATE plans SET features = features ||
    '{"ai_text_generation": true, "ai_image_analysis": true, "ai_embeddings": true,
      "ai_monthly_text_quota": 500, "ai_monthly_image_quota": 200}'::jsonb
WHERE name = 'premium';
