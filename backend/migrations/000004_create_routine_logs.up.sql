CREATE TABLE routine_logs (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    routine_id  UUID NOT NULL REFERENCES routines(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    logged_at   DATE NOT NULL,
    count       INTEGER NOT NULL DEFAULT 1,
    note        TEXT,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_routine_logs_unique ON routine_logs(routine_id, logged_at);
CREATE INDEX idx_routine_logs_user_id ON routine_logs(user_id);
CREATE INDEX idx_routine_logs_logged_at ON routine_logs(logged_at);
