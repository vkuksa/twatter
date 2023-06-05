CREATE DATABASE IF NOT EXISTS twatter;

USE twatter;

CREATE TABLE IF NOT EXISTS public.messages (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    content STRING NULL,
    created_at TIME NULL DEFAULT current_time():::TIME,
    CONSTRAINT messages_pkey PRIMARY KEY (id ASC)
);

