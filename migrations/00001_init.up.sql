CREATE TABLE test (
                      id SERIAL PRIMARY KEY,
                      name VARCHAR(100),
                      age INT,
                      email VARCHAR(100)
);

CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       login VARCHAR(255) NOT NULL,
                       password VARCHAR(255) NOT NULL
);
