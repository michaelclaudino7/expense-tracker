ALTER TABLE transactions DROP COLUMN user_id;

ALTER TABLE categories DROP CONSTRAINT categories_user_name_type_key;
ALTER TABLE categories DROP COLUMN user_id;
ALTER TABLE categories ADD CONSTRAINT categories_name_type_key UNIQUE (name, type);

DROP TABLE users;