CREATE TABLE urls (
                      id VARCHAR(255) PRIMARY KEY,
                      original_url TEXT NOT NULL,
                      user_id VARCHAR(255),
                      is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
                      UNIQUE(original_url, user_id)
);
