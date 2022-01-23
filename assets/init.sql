CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    balance real NOT NULL
);

CREATE TABLE IF NOT EXISTS  transactions (
    user_uuid UUID,
    trx_uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trx_date date NOT NULL DEFAULT current_date,
    u_timestamp integer DEFAULT extract(epoch from now()),
    trx_time time with time zone NOT NULL DEFAULT current_time,
    who text,
    description text,
    amount real,
    currency text
);
