ALTER TABLE users
ALTER COLUMN password_hash DROP NOT NULL;

ALTER TABLE users
DROP CONSTRAINT users_password_hash_check;