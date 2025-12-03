CREATE TABLE urls (
                      id TEXT PRIMARY KEY,
                      original_url TEXT NOT NULL,
                      user_id TEXT,
                      is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
                      UNIQUE(original_url, user_id)
);
