-- Downloaded from: https://github.com/oslabs-beta/ditto/blob/fa80cb2429f0de54951b6580ef893b2ed08ac698/dittoDB2.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 16.1
-- Dumped by pg_dump version 16.2

-- Started on 2024-06-05 19:19:25

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
-- TOC entry 858 (class 1247 OID 16517)
-- Name: status_enum; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.status_enum AS ENUM (
    'Failed',
    'Success',
    'Pending'
);


--
-- TOC entry 223 (class 1255 OID 16998)
-- Name: delete_column_value(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.delete_column_value() RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    UPDATE projects SET code = NULL, code_timestamp = NULL WHERE code_timestamp < NOW() - INTERVAL '1 hour';
END;
$$;


--
-- TOC entry 228 (class 1255 OID 16996)
-- Name: update_timestamp(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$BEGIN
    IF NEW.code IS NOT NULL THEN
        NEW.code_timestamp := CURRENT_TIMESTAMP;
    ELSE
        NEW.code_timestamp := NULL;
    END IF;
    RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 218 (class 1259 OID 16487)
-- Name: databases; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.databases (
    db_id integer NOT NULL,
    date_created timestamp with time zone DEFAULT timezone('America/New_York'::text, date_trunc('second'::text, now())) NOT NULL,
    connection_string character varying NOT NULL,
    migration_id integer[]
);


--
-- TOC entry 217 (class 1259 OID 16486)
-- Name: databases_db_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.databases_db_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 4350 (class 0 OID 0)
-- Dependencies: 217
-- Name: databases_db_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.databases_db_id_seq OWNED BY public.databases.db_id;


--
-- TOC entry 216 (class 1259 OID 16476)
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    user_id integer NOT NULL,
    username character varying NOT NULL,
    password character varying NOT NULL,
    date_created timestamp with time zone DEFAULT timezone('America/New_York'::text, date_trunc('second'::text, now())) NOT NULL
);


--
-- TOC entry 215 (class 1259 OID 16475)
-- Name: users_user_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 4351 (class 0 OID 0)
-- Dependencies: 215
-- Name: users_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_user_id_seq OWNED BY public.users.user_id;


--
-- TOC entry 219 (class 1259 OID 16510)
-- Name: migration_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.migration_logs (
    migration_id integer DEFAULT nextval('public.users_user_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    date_created timestamp with time zone DEFAULT timezone('America/New_York'::text, date_trunc('second'::text, now())) NOT NULL,
    database_id integer NOT NULL,
    description character varying,
    status public.status_enum DEFAULT 'Pending'::public.status_enum,
    script character varying,
    checksum character varying,
    executed_at character varying,
    version character varying
);


--
-- TOC entry 222 (class 1259 OID 16813)
-- Name: project_db; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.project_db (
    project_db_id integer DEFAULT nextval('public.databases_db_id_seq'::regclass) NOT NULL,
    project_id integer,
    db_id integer,
    db_name character varying
);


--
-- TOC entry 221 (class 1259 OID 16786)
-- Name: projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.projects (
    date_created timestamp with time zone DEFAULT timezone('America/New_York'::text, date_trunc('second'::text, now())),
    project_name character varying,
    project_id integer DEFAULT nextval('public.databases_db_id_seq'::regclass) NOT NULL,
    owner integer,
    code character varying,
    code_timestamp timestamp with time zone
);


--
-- TOC entry 220 (class 1259 OID 16778)
-- Name: user_projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_projects (
    user_project_id character varying DEFAULT nextval('public.databases_db_id_seq'::regclass) NOT NULL,
    project_name character varying,
    user_id integer,
    project_id integer,
    role character varying,
    date_joined timestamp with time zone DEFAULT timezone('America/New_York'::text, date_trunc('second'::text, now()))
);


--
-- TOC entry 4169 (class 2604 OID 16490)
-- Name: databases db_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.databases ALTER COLUMN db_id SET DEFAULT nextval('public.databases_db_id_seq'::regclass);


--
-- TOC entry 4167 (class 2604 OID 16479)
-- Name: users user_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN user_id SET DEFAULT nextval('public.users_user_id_seq'::regclass);


--
-- TOC entry 4182 (class 2606 OID 16624)
-- Name: databases connection_string; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.databases
    ADD CONSTRAINT connection_string UNIQUE (connection_string);


--
-- TOC entry 4184 (class 2606 OID 16492)
-- Name: databases databases_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.databases
    ADD CONSTRAINT databases_pkey PRIMARY KEY (db_id);


--
-- TOC entry 4186 (class 2606 OID 16515)
-- Name: migration_logs migration_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.migration_logs
    ADD CONSTRAINT migration_logs_pkey PRIMARY KEY (migration_id);


--
-- TOC entry 4192 (class 2606 OID 16821)
-- Name: project_db project_db_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_db
    ADD CONSTRAINT project_db_pkey PRIMARY KEY (project_db_id);


--
-- TOC entry 4190 (class 2606 OID 16802)
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (project_id);


--
-- TOC entry 4188 (class 2606 OID 16785)
-- Name: user_projects user_projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_projects
    ADD CONSTRAINT user_projects_pkey PRIMARY KEY (user_project_id);


--
-- TOC entry 4180 (class 2606 OID 16481)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


--
-- TOC entry 4199 (class 2620 OID 16997)
-- Name: projects set_timestamp; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER set_timestamp BEFORE INSERT OR UPDATE ON public.projects FOR EACH ROW EXECUTE FUNCTION public.update_timestamp();


--
-- TOC entry 4197 (class 2606 OID 16827)
-- Name: project_db db_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_db
    ADD CONSTRAINT db_id_fkey FOREIGN KEY (db_id) REFERENCES public.databases(db_id) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- TOC entry 4193 (class 2606 OID 16529)
-- Name: migration_logs migration_logs_database_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.migration_logs
    ADD CONSTRAINT migration_logs_database_id_fkey FOREIGN KEY (database_id) REFERENCES public.databases(db_id) NOT VALID;


--
-- TOC entry 4194 (class 2606 OID 16524)
-- Name: migration_logs migration_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.migration_logs
    ADD CONSTRAINT migration_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) NOT VALID;


--
-- TOC entry 4198 (class 2606 OID 16822)
-- Name: project_db project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_db
    ADD CONSTRAINT project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- TOC entry 4195 (class 2606 OID 16832)
-- Name: user_projects project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_projects
    ADD CONSTRAINT project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- TOC entry 4196 (class 2606 OID 16803)
-- Name: user_projects user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_projects
    ADD CONSTRAINT user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON UPDATE CASCADE ON DELETE CASCADE NOT VALID;


--
-- TOC entry 4343 (class 0 OID 16487)
-- Dependencies: 218
-- Name: databases; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.databases ENABLE ROW LEVEL SECURITY;

-- Completed on 2024-06-05 19:19:28

--
-- PostgreSQL database dump complete
--

