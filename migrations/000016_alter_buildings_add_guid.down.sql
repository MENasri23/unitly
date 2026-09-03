ALTER TABLE buildings
DROP CONSTRAINT buildings_pkey;

ALTER TABLE buildings
DROP CONSTRAINT buildings_guid_key;

ALTER TABLE buildings
DROP COLUMN IF EXISTS id;

ALTER TABLE buildings
RENAME COLUMN guid TO id;

ALTER TABLE buildings
ADD CONSTRAINT buildings_pkey PRIMARY KEY (id);