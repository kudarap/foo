CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE archers(
    id uuid DEFAULT uuid_generate_v4(),
    first_name text,
    last_name text,
    PRIMARY KEY(id)
);

-- demo purposes
INSERT INTO archers (first_name, last_name) VALUES ('Dave', 'Grohl');