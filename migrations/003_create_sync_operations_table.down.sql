DROP INDEX IF EXISTS idx_sync_operations_secret_id;
DROP INDEX IF EXISTS idx_sync_operations_timestamp;
DROP INDEX IF EXISTS idx_sync_operations_user_id;
DROP TABLE IF EXISTS sync_operations;
DROP TYPE IF EXISTS operation_type;