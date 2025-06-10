CREATE TABLE IF NOT EXISTS books (
    id bigserial PRIMARY KEY,
    title text NOT NULL,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    cover_picture text,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    version integer NOT NULL DEFAULT 1
);