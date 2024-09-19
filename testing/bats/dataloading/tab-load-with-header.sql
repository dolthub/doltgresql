BEGIN;

CREATE TABLE Regions (
   id SERIAL UNIQUE NOT NULL,
   code VARCHAR(4) UNIQUE NOT NULL,
   capital VARCHAR(10) NOT NULL,
   name VARCHAR(255) UNIQUE NOT NULL
);

COPY regions (id, code, capital, name) FROM stdin WITH (HEADER);
id  code  capital  name
1	01	97105	Guadeloupe
2	02	97209	Martinique
3	03	97302	Guyane
\.

COMMIT;
