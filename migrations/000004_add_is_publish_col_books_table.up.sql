ALTER TABLE books
ADD COLUMN is_published BOOLEAN DEFAULT FALSE,
ADD COLUMN published_at TIMESTAMP;
