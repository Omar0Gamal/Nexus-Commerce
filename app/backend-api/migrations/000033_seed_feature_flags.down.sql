-- migrations/000033_seed_feature_flags.down.sql
-- Safe rollback: reset features to an empty object. Does NOT delete plans.
UPDATE plans SET features = '{}'::jsonb WHERE name IN ('basic', 'professional', 'premium');
