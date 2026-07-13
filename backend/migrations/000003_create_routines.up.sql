CREATE TYPE period_type AS ENUM ('daily', 'weekly', 'monthly');

CREATE TABLE routines (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title        VARCHAR(100) NOT NULL,
    description  TEXT,
    period_type  period_type NOT NULL,
    target_count INTEGER NOT NULL DEFAULT 1,
    icon         VARCHAR(50),
    color        VARCHAR(20),
    is_active    BOOLEAN NOT NULL DEFAULT true,
    sort_order   INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_routines_user_id ON routines(user_id);
CREATE INDEX idx_routines_user_active ON routines(user_id, is_active);
