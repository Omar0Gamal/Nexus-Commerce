UPDATE plans SET features = features - ARRAY['ai_text_generation','ai_image_analysis','ai_embeddings','ai_monthly_text_quota','ai_monthly_image_quota']
WHERE name IN ('basic', 'professional', 'premium');
