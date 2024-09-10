BEGIN;

-- Database schema / Schéma de la base de données

-- Regions / Régions
CREATE TABLE Regions (
   id SERIAL UNIQUE NOT NULL,
   code VARCHAR(4) UNIQUE NOT NULL,
   capital VARCHAR(10) NOT NULL, -- REFERENCES Towns (code),
   -- TODO: TEXT columns do not work correctly in Doltgres yet
   -- name TEXT UNIQUE NOT NULL
   name VARCHAR(255) UNIQUE NOT NULL
);

SET client_encoding = 'utf-8';
SET check_function_bodies = false;
SET search_path = public, pg_catalog;

COPY regions (id, code, capital, name) FROM stdin;
1	01	97105	Guadeloupe
2	02	97209	Martinique
3	03	97302	Guyane
4	04	97411	La Réunion
5	11	75056	Île-de-France
6	21	51108	Champagne-Ardenne
7	22	80021	Picardie
8	23	76540	Haute-Normandie
9	24	45234	Centre
10	25	14118	Basse-Normandie
11	26	21231	Bourgogne
12	31	59350	Nord-Pas-de-Calais
13	41	57463	Lorraine
14	42	67482	Alsace
15	43	25056	Franche-Comté
16	52	44109	Pays de la Loire
17	53	35238	Bretagne
18	54	86194	Poitou-Charentes
19	72	33063	Aquitaine
20	73	31555	Midi-Pyrénées
21	74	87085	Limousin
22	82	69123	Rhône-Alpes
23	83	63113	Auvergne
24	91	34172	Languedoc-Roussillon
25	93	13055	Provence-Alpes-Côte d'Azur
26	94	2A004	Corse
\.

COMMIT;
