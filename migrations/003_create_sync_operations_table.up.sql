CREATE TYPE operation_type AS ENUM ('create', 'update', 'delete');

CREATE TABLE IF NOT EXISTS sync_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    secret_id UUID,
    operation operation_type NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sync_operations_user_id ON sync_operations(user_id);
CREATE INDEX IF NOT EXISTS idx_sync_operations_timestamp ON sync_operations(timestamp);
CREATE INDEX IF NOT EXISTS idx_sync_operations_secret_id ON sync_operations(secret_id);