-- Downloaded from: https://github.com/Mistral-war2ru/PG-connect/blob/ba0e4f62b6b9b5799cf645fdc072d59f727b9ffe/lab9/TestDB.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 15.2
-- Dumped by pg_dump version 15.2

-- Started on 2023-04-19 19:40:00

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 3370 (class 1262 OID 25411)
-- Name: StudentsDB; Type: DATABASE; Schema: -; Owner: postgres
--

CREATE DATABASE "StudentsDB" WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'Russian_Russia.1251';


ALTER DATABASE "StudentsDB" OWNER TO postgres;

\connect "StudentsDB"

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 3371 (class 0 OID 0)
-- Name: StudentsDB; Type: DATABASE PROPERTIES; Schema: -; Owner: postgres
--

ALTER ROLE postgres IN DATABASE "StudentsDB" SET effective_cache_size TO '65536';


\connect "StudentsDB"

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 223 (class 1255 OID 25412)
-- Name: get_fiit_2019(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_fiit_2019() RETURNS text
    LANGUAGE plpgsql
    AS $$
declare
ret text default '';
rec  record;
curs refcursor;
begin
--call proc1(curs, 2019, 'ФИИТ');
call proc1(curs, 2018, 'ФИИТ');
loop
	fetch curs into rec;
	exit when not found;
	ret := ret || '{' || rec.SpecialtyName || ',' || rec.SetName|| ',' || rec.GroupName || ',' || rec.SetYear|| '}';
end loop;
close curs;
return ret;
end; $$;


ALTER FUNCTION public.get_fiit_2019() OWNER TO postgres;

--
-- TOC entry 224 (class 1255 OID 25413)
-- Name: proc1(refcursor); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.proc1(INOUT curs refcursor)
    LANGUAGE plpgsql
    AS $$
DECLARE
lcurs refcursor;
BEGIN
open lcurs for
select SpecialtyName, SetName, GroupName, SetYear
from Groups inner join (Sets inner join Specialities on Sets.SpecialtyID = Specialities.SpecialtyID) on Sets.SetID = Groups.SetID
where Sets.SetYear = 2019 AND Specialities.SpecialtyCode = 'ФИИТ';
curs = lcurs;
END $$;


ALTER PROCEDURE public.proc1(INOUT curs refcursor) OWNER TO postgres;

--
-- TOC entry 225 (class 1255 OID 25414)
-- Name: proc1(refcursor, integer, text); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.proc1(INOUT curs refcursor, IN yr integer, IN spec text)
    LANGUAGE plpgsql
    AS $$
DECLARE
lcurs refcursor;
BEGIN
open lcurs for
select SpecialtyName, SetName, GroupName, SetYear
from Groups inner join (Sets inner join Specialities on Sets.SpecialtyID = Specialities.SpecialtyID) on Sets.SetID = Groups.SetID
where Sets.SetYear = yr AND Specialities.SpecialtyCode = spec;
curs = lcurs;
END $$;


ALTER PROCEDURE public.proc1(INOUT curs refcursor, IN yr integer, IN spec text) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 214 (class 1259 OID 25415)
-- Name: groups; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.groups (
    groupid integer NOT NULL,
    setid integer NOT NULL,
    groupname text NOT NULL,
    teachformid integer NOT NULL
);


ALTER TABLE public.groups OWNER TO postgres;

--
-- TOC entry 215 (class 1259 OID 25420)
-- Name: groups_groupid_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.groups ALTER COLUMN groupid ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.groups_groupid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 221 (class 1259 OID 25462)
-- Name: labs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.labs (
    "ID" integer NOT NULL,
    "Name" text NOT NULL
);


ALTER TABLE public.labs OWNER TO postgres;

--
-- TOC entry 222 (class 1259 OID 25469)
-- Name: labs_ID_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.labs ALTER COLUMN "ID" ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public."labs_ID_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 216 (class 1259 OID 25421)
-- Name: sets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sets (
    setid integer NOT NULL,
    setname text NOT NULL,
    setyear integer NOT NULL,
    specialtyid integer NOT NULL
);


ALTER TABLE public.sets OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 25426)
-- Name: sets_setid_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.sets ALTER COLUMN setid ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.sets_setid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 218 (class 1259 OID 25427)
-- Name: specialities; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.specialities (
    specialtyid integer NOT NULL,
    specialtyname text NOT NULL,
    specialtycode text NOT NULL,
    descript text
);


ALTER TABLE public.specialities OWNER TO postgres;

--
-- TOC entry 219 (class 1259 OID 25432)
-- Name: specialities_specialtyid_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.specialities ALTER COLUMN specialtyid ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.specialities_specialtyid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 220 (class 1259 OID 25458)
-- Name: v1; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.v1 AS
 SELECT groups.groupname,
    sets.setname,
    sets.setyear,
    specialities.specialtycode
   FROM (public.groups
     JOIN (public.sets
     JOIN public.specialities ON ((specialities.specialtyid = sets.specialtyid))) ON ((sets.setid = groups.setid)));


ALTER TABLE public.v1 OWNER TO postgres;

--
-- TOC entry 3357 (class 0 OID 25415)
-- Dependencies: 214
-- Data for Name: groups; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.groups OVERRIDING SYSTEM VALUE VALUES (3, 2, 'ПИ-20-а', 1);
INSERT INTO public.groups OVERRIDING SYSTEM VALUE VALUES (4, 2, 'ПИ-20-б', 1);
INSERT INTO public.groups OVERRIDING SYSTEM VALUE VALUES (7, 4, 'ФИИТ-16-а', 1);
INSERT INTO public.groups OVERRIDING SYSTEM VALUE VALUES (8, 4, 'ФИИТ-16-б', 1);
INSERT INTO public.groups OVERRIDING SYSTEM VALUE VALUES (9, 5, 'ФИИТ-18-а', 1);
INSERT INTO public.groups OVERRIDING SYSTEM VALUE VALUES (10, 5, 'ФИИТ-18-б', 1);
INSERT INTO public.groups OVERRIDING SYSTEM VALUE VALUES (12, 6, 'ФИИТ-20-а', 1);
INSERT INTO public.groups OVERRIDING SYSTEM VALUE VALUES (13, 6, 'ФИИТ-20-б', 1);
INSERT INTO public.groups OVERRIDING SYSTEM VALUE VALUES (11, 5, 'ФИИТ-18-в', 2);
INSERT INTO public.groups OVERRIDING SYSTEM VALUE VALUES (14, 6, 'ФИИТ-20-в', 2);
INSERT INTO public.groups OVERRIDING SYSTEM VALUE VALUES (15, 1, 'ПМ-20-1', 2);
INSERT INTO public.groups OVERRIDING SYSTEM VALUE VALUES (16, 1, 'ПМ-20-2', 2);
INSERT INTO public.groups OVERRIDING SYSTEM VALUE VALUES (18, 2, 'test', 2);


--
-- TOC entry 3363 (class 0 OID 25462)
-- Dependencies: 221
-- Data for Name: labs; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.labs OVERRIDING SYSTEM VALUE VALUES (1, 'lab1');
INSERT INTO public.labs OVERRIDING SYSTEM VALUE VALUES (2, 'lab2');


--
-- TOC entry 3359 (class 0 OID 25421)
-- Dependencies: 216
-- Data for Name: sets; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.sets OVERRIDING SYSTEM VALUE VALUES (1, 'ПИ-15', 2015, 1);
INSERT INTO public.sets OVERRIDING SYSTEM VALUE VALUES (2, 'ПИ-20', 2020, 2);
INSERT INTO public.sets OVERRIDING SYSTEM VALUE VALUES (3, 'ФИИТ-15', 2015, 3);
INSERT INTO public.sets OVERRIDING SYSTEM VALUE VALUES (4, 'ФИИТ-16', 2016, 3);
INSERT INTO public.sets OVERRIDING SYSTEM VALUE VALUES (5, 'ФИИТ-18', 2018, 3);
INSERT INTO public.sets OVERRIDING SYSTEM VALUE VALUES (6, 'ФИИТ-20', 2019, 3);
INSERT INTO public.sets OVERRIDING SYSTEM VALUE VALUES (7, 'name12', 2016, 10);


--
-- TOC entry 3361 (class 0 OID 25427)
-- Dependencies: 218
-- Data for Name: specialities; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.specialities OVERRIDING SYSTEM VALUE VALUES (3, 'Фундаментальная информатика и ИТ', 'ФИИТ', 'Кафедра ВТ');
INSERT INTO public.specialities OVERRIDING SYSTEM VALUE VALUES (1, 'Прикладная информатика', 'ПИ', 'Кафедра ПМИ');
INSERT INTO public.specialities OVERRIDING SYSTEM VALUE VALUES (2, 'Прикладная математика и информатика', 'ПМИ', 'Кафедра ПМИ');
INSERT INTO public.specialities OVERRIDING SYSTEM VALUE VALUES (10, 'test1', 'test2', 'test3');


--
-- TOC entry 3372 (class 0 OID 0)
-- Dependencies: 215
-- Name: groups_groupid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.groups_groupid_seq', 18, true);


--
-- TOC entry 3373 (class 0 OID 0)
-- Dependencies: 222
-- Name: labs_ID_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public."labs_ID_seq"', 10, true);


--
-- TOC entry 3374 (class 0 OID 0)
-- Dependencies: 217
-- Name: sets_setid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.sets_setid_seq', 7, true);


--
-- TOC entry 3375 (class 0 OID 0)
-- Dependencies: 219
-- Name: specialities_specialtyid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.specialities_specialtyid_seq', 10, true);


--
-- TOC entry 3196 (class 2606 OID 25434)
-- Name: groups groups_groupid_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.groups
    ADD CONSTRAINT groups_groupid_key UNIQUE (groupid);


--
-- TOC entry 3211 (class 2606 OID 25468)
-- Name: labs labs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.labs
    ADD CONSTRAINT labs_pkey PRIMARY KEY ("ID");


--
-- TOC entry 3199 (class 2606 OID 25436)
-- Name: groups pkgroups; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.groups
    ADD CONSTRAINT pkgroups PRIMARY KEY (groupid);


--
-- TOC entry 3202 (class 2606 OID 25438)
-- Name: sets pksets; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sets
    ADD CONSTRAINT pksets PRIMARY KEY (setid);


--
-- TOC entry 3206 (class 2606 OID 25440)
-- Name: specialities pkspecialities; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.specialities
    ADD CONSTRAINT pkspecialities PRIMARY KEY (specialtyid);


--
-- TOC entry 3204 (class 2606 OID 25442)
-- Name: sets sets_setid_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sets
    ADD CONSTRAINT sets_setid_key UNIQUE (setid);


--
-- TOC entry 3208 (class 2606 OID 25444)
-- Name: specialities specialities_specialtyid_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.specialities
    ADD CONSTRAINT specialities_specialtyid_key UNIQUE (specialtyid);


--
-- TOC entry 3197 (class 1259 OID 25445)
-- Name: igroupsteachformid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX igroupsteachformid ON public.groups USING btree (teachformid);


--
-- TOC entry 3200 (class 1259 OID 25446)
-- Name: isetssetyear; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX isetssetyear ON public.sets USING btree (setyear);


--
-- TOC entry 3209 (class 1259 OID 25447)
-- Name: specialtyindex; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX specialtyindex ON public.specialities USING btree (specialtyid);

ALTER TABLE public.specialities CLUSTER ON specialtyindex;


--
-- TOC entry 3212 (class 2606 OID 25448)
-- Name: groups fkgroupssets; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.groups
    ADD CONSTRAINT fkgroupssets FOREIGN KEY (setid) REFERENCES public.sets(setid) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3213 (class 2606 OID 25453)
-- Name: sets fksetsspecialities; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sets
    ADD CONSTRAINT fksetsspecialities FOREIGN KEY (specialtyid) REFERENCES public.specialities(specialtyid) ON UPDATE CASCADE;


-- Completed on 2023-04-19 19:40:00

--
-- PostgreSQL database dump complete
--

