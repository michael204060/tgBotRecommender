CREATE TABLE IF NOT EXISTS messages (
                                        id SERIAL PRIMARY KEY,
                                        content TEXT NOT NULL,
                                        user_id INTEGER NOT NULL,
                                        priority INTEGER NOT NULL,
                                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_priority ON messages (user_id, priority);