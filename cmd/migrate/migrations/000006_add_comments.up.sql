CREATE TABLE IF NOT EXISTS comments (
    id bigserial PRIMARY KEY,
    post_id bigserial NOT NULL,
    user_Id bigserial NOT NULL,
    content text NOT NULL,
    created_at  timestamp(0) with time zone NOT NULL DEFAULT NOW()
);