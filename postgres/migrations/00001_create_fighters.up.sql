CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE fighters (
    id uuid DEFAULT uuid_generate_v4(),
    first_name text,
    last_name text,
    PRIMARY KEY(id)
);

-- demo purposes
INSERT INTO fighters (first_name, last_name) VALUES ('Dave', 'Grohl');