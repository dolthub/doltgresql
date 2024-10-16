BEGIN;

CREATE TABLE Regions (
   "Id" SERIAL UNIQUE NOT NULL,
   "Code" VARCHAR(4) UNIQUE NOT NULL,
   "Capital" VARCHAR(10) NOT NULL,
   "Name" VARCHAR(255) UNIQUE NOT NULL
);

COPY regions ("Id", "Code", "Capital", "Name") FROM stdin;
1	01	97105	Guadeloupe
2	02	97209	Martinique
3	03	97302	Guyane
\.

COMMIT;
