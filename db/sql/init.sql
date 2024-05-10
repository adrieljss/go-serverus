-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- @block
CREATE TABLE IF NOT EXISTS users (
  uid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id VARCHAR(35) UNIQUE NOT NULL,
  username VARCHAR(35) NOT NULL,
  email VARCHAR(50) UNIQUE NOT NULL,
  pfp_url TEXT,
  password TEXT NOT NULL,
  last_updated TIMESTAMP NOT NULL DEFAULT now(),
  created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- @block insert test
INSERT INTO users (user_id, username, email, password) VALUES (
  'adrielcp',
  'Adriel Jansen',
  'adrielsiahaya@gmail.com',
  crypt('password', gen_salt('bf'))
) RETURNING *;

-- @block delete test
DELETE FROM users WHERE password = crypt('password', password) RETURNING *;

-- @block select test
SELECT * FROM users WHERE password = crypt('password', password);