ALTER TABLE buildings
DROP CONSTRAINT buildings_pkey;

ALTER TABLE buildings
RENAME COLUMN id TO guid;

ALTER TABLE buildings
ADD COLUMN id BIGSERIAL;

ALTER TABLE buildings
ADD CONSTRAINT buildings_pkey PRIMARY KEY (id);

ALTER TABLE buildings
ALTER COLUMN guid SET NOT NULL;

ALTER TABLE buildings
ADD CONSTRAINT buildings_guid_key UNIQUE (guid);