ALTER TABLE buildings ADD COLUMN version INT;

UPDATE buildings SET version = 1 WHERE version IS NULL;

ALTER TABLE buildings ALTER COLUMN version SET NOT NULL;

ALTER TABLE buildings ALTER COLUMN version SET DEFAULT 1;
