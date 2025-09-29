-- Downloaded from: https://github.com/kirooha/adtech-simple/blob/a1cf7c0607bd2aea3e6aa2d042c9fe2b4e28e797/db/structure.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 13.8
-- Dumped by pg_dump version 16.0

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
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

-- *not* creating schema, since initdb creates it


--
-- Name: set_updated_at_column(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.set_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = now() at time zone 'utc';
RETURN NEW;
END;
$$;


SET default_table_access_method = heap;

--
-- Name: campaigns; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.campaigns (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    adserver_id uuid,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL
);


--
-- Name: goose_db_version; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goose_db_version (
    id integer NOT NULL,
    version_id bigint NOT NULL,
    is_applied boolean NOT NULL,
    tstamp timestamp without time zone DEFAULT now()
);


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.goose_db_version_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.goose_db_version_id_seq OWNED BY public.goose_db_version.id;


--
-- Name: gue_jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gue_jobs (
    job_id text NOT NULL,
    priority smallint NOT NULL,
    run_at timestamp with time zone NOT NULL,
    job_type text NOT NULL,
    args bytea NOT NULL,
    error_count integer DEFAULT 0 NOT NULL,
    last_error text,
    queue text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: goose_db_version id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.goose_db_version ALTER COLUMN id SET DEFAULT nextval('public.goose_db_version_id_seq'::regclass);


--
-- Name: goose_db_version goose_db_version_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- Name: gue_jobs gue_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gue_jobs
    ADD CONSTRAINT gue_jobs_pkey PRIMARY KEY (job_id);


--
-- Name: campaigns__adserver_id__uidx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX campaigns__adserver_id__uidx ON public.campaigns USING btree (adserver_id);


--
-- Name: campaigns__name__uidx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX campaigns__name__uidx ON public.campaigns USING btree (name);


--
-- Name: idx_gue_jobs_selector; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_gue_jobs_selector ON public.gue_jobs USING btree (queue, run_at, priority);


--
-- Name: campaigns update_campaign_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_campaign_updated_at BEFORE UPDATE ON public.campaigns FOR EACH ROW EXECUTE FUNCTION public.set_updated_at_column();


--
-- PostgreSQL database dump complete
--

