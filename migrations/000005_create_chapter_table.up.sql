CREATE TABLE IF NOT EXISTS chapters (
    id bigserial PRIMARY KEY,
    chapter_no bigint NOT NULL DEFAULT 0,
    title text NOT NULL,
    description text,
    content text,
    book_id bigint NOT NULL REFERENCES books ON DELETE CASCADE,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    version integer NOT NULL DEFAULT 1
);