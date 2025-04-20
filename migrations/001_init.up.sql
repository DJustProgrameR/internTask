CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


CREATE TABLE IF NOT EXISTS users (
                       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       email VARCHAR(50) UNIQUE NOT NULL,
                       password TEXT NOT NULL,
                       role INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS pvz (
                     id UUID PRIMARY KEY DEFAULT uuid_generate_v4() ,
                     registration_date DATE NOT NULL,
                     city INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS receptions (
                            id UUID PRIMARY KEY DEFAULT uuid_generate_v4 () ,
                            pvz_id UUID NOT NULL,
                            date_time TIMESTAMP NOT NULL,
                            status INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS products (
                          id UUID PRIMARY KEY DEFAULT uuid_generate_v4() ,
                          reception_id UUID NOT NULL,
                          date_time TIMESTAMP NOT NULL,
                          type INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users USING hash (email);
CREATE INDEX IF NOT EXISTS idx_users_email ON receptions (date_time);