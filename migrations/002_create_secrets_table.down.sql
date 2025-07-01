DROP INDEX IF EXISTS idx_secrets_sync_hash;
DROP INDEX IF EXISTS idx_secrets_name;
DROP INDEX IF EXISTS idx_secrets_type;
DROP INDEX IF EXISTS idx_secrets_user_id;
DROP TABLE IF EXISTS secrets;
DROP TYPE IF EXISTS secret_type;