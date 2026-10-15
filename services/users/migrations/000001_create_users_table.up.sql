CREATE TYPE user_status AS ENUM (
    'ACTIVE',
    'BLOCKED'
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    email VARCHAR(255),
    phone VARCHAR(16) NOT NULL,

    first_name VARCHAR(100),
    last_name VARCHAR(100),
    middle_name VARCHAR(100),
    full_name VARCHAR(300),

    birth_date DATE,
    avatar_file_id UUID,

    status user_status NOT NULL DEFAULT 'ACTIVE',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_phone_unique UNIQUE (phone)
);

CREATE UNIQUE INDEX users_email_unique
    ON users (LOWER(email))
    WHERE email IS NOT NULL;