ALTER TABLE users
ALTER COLUMN password_hash SET NOT NULL;

ALTER TABLE users
ADD CONSTRAINT users_password_hash_check
CHECK (password_hash <> '');