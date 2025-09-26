-- Downloaded from: https://github.com/yase-search/yase-engine/blob/2a41b1234d164f034048d0d87bc9df11d6b13812/sql/yase.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 9.2.8
-- Dumped by pg_dump version 9.5.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- Name: btree_gist; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS btree_gist WITH SCHEMA public;


--
-- Name: EXTENSION btree_gist; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION btree_gist IS 'support for indexing common datatypes in GiST';


--
-- Name: hstore; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS hstore WITH SCHEMA public;


--
-- Name: EXTENSION hstore; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION hstore IS 'data type for storing sets of (key, value) pairs';


SET search_path = public, pg_catalog;

--
-- Name: http_protocol; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE http_protocol AS ENUM (
    'http',
    'https'
);


SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: errors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE errors (
    id bigint NOT NULL,
    idpage bigint NOT NULL,
    errorcode integer NOT NULL,
    date timestamp with time zone NOT NULL
);


--
-- Name: errors_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE errors_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: errors_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE errors_id_seq OWNED BY errors.id;


--
-- Name: errors_idpage_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE errors_idpage_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: errors_idpage_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE errors_idpage_seq OWNED BY errors.idpage;


--
-- Name: page_content_idpage_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE page_content_idpage_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE pages (
    id bigint NOT NULL,
    id_website bigint NOT NULL,
    url character varying(2000) NOT NULL,
    clicks bigint DEFAULT 0 NOT NULL,
    page_rank integer DEFAULT 0 NOT NULL,
    content text NOT NULL,
    title character varying(100) NOT NULL,
    description character varying(255) NOT NULL,
    crawl_date timestamp with time zone NOT NULL,
    size integer NOT NULL,
    load_time integer NOT NULL,
    locale character varying(5) NOT NULL,
    favicon character varying(1000)
);


--
-- Name: pages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE pages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE pages_id_seq OWNED BY pages.id;


--
-- Name: pages_idwebsite_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE pages_idwebsite_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pages_idwebsite_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE pages_idwebsite_seq OWNED BY pages.id_website;


--
-- Name: pages_links; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE pages_links (
    "idRefferer" bigint NOT NULL,
    "idDestination" bigint NOT NULL
);


--
-- Name: pages_links_iddestination_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE pages_links_iddestination_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pages_links_iddestination_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE pages_links_iddestination_seq OWNED BY pages_links."idDestination";


--
-- Name: pages_links_idrefferer_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE pages_links_idrefferer_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pages_links_idrefferer_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE pages_links_idrefferer_seq OWNED BY pages_links."idRefferer";


--
-- Name: pages_words; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE pages_words (
    idpage bigint NOT NULL,
    idword bigint NOT NULL,
    strength integer NOT NULL
);


--
-- Name: pages_words_idpage_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE pages_words_idpage_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pages_words_idpage_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE pages_words_idpage_seq OWNED BY pages_words.idpage;


--
-- Name: pages_words_idword_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE pages_words_idword_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pages_words_idword_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE pages_words_idword_seq OWNED BY pages_words.idword;


--
-- Name: websites; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE websites (
    id bigint NOT NULL,
    domain character varying(255) NOT NULL,
    site_rank integer NOT NULL,
    protocol http_protocol NOT NULL
);


--
-- Name: websites_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE websites_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: websites_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE websites_id_seq OWNED BY websites.id;


--
-- Name: words; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE words (
    id bigint NOT NULL,
    text character varying(100) NOT NULL
);


--
-- Name: words_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE words_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: words_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE words_id_seq OWNED BY words.id;


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY errors ALTER COLUMN id SET DEFAULT nextval('errors_id_seq'::regclass);


--
-- Name: idpage; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY errors ALTER COLUMN idpage SET DEFAULT nextval('errors_idpage_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY pages ALTER COLUMN id SET DEFAULT nextval('pages_id_seq'::regclass);


--
-- Name: id_website; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY pages ALTER COLUMN id_website SET DEFAULT nextval('pages_idwebsite_seq'::regclass);


--
-- Name: idRefferer; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY pages_links ALTER COLUMN "idRefferer" SET DEFAULT nextval('pages_links_idrefferer_seq'::regclass);


--
-- Name: idDestination; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY pages_links ALTER COLUMN "idDestination" SET DEFAULT nextval('pages_links_iddestination_seq'::regclass);


--
-- Name: idpage; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY pages_words ALTER COLUMN idpage SET DEFAULT nextval('pages_words_idpage_seq'::regclass);


--
-- Name: idword; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY pages_words ALTER COLUMN idword SET DEFAULT nextval('pages_words_idword_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY websites ALTER COLUMN id SET DEFAULT nextval('websites_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY words ALTER COLUMN id SET DEFAULT nextval('words_id_seq'::regclass);


--
-- Name: errors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY errors
    ADD CONSTRAINT errors_pkey PRIMARY KEY (id);


--
-- Name: pages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY pages
    ADD CONSTRAINT pages_pkey PRIMARY KEY (id);


--
-- Name: unique_domain; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY websites
    ADD CONSTRAINT unique_domain UNIQUE (domain);


--
-- Name: websites_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY websites
    ADD CONSTRAINT websites_pkey PRIMARY KEY (id);


--
-- Name: words_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY words
    ADD CONSTRAINT words_pkey PRIMARY KEY (id),
    ADD CONSTRAINT unique_word UNIQUE (text);


--
-- Name: words_text_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX words_text_idx ON words USING btree (text);


--
-- Name: fk_iddestination; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY pages_links
    ADD CONSTRAINT fk_iddestination FOREIGN KEY ("idDestination") REFERENCES pages(id);


--
-- Name: fk_idpage; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY pages_words
    ADD CONSTRAINT fk_idpage FOREIGN KEY (idpage) REFERENCES pages(id);


--
-- Name: fk_idpage; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY errors
    ADD CONSTRAINT fk_idpage FOREIGN KEY (idpage) REFERENCES pages(id);


--
-- Name: fk_idrefferer; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY pages_links
    ADD CONSTRAINT fk_idrefferer FOREIGN KEY ("idRefferer") REFERENCES pages(id);


--
-- Name: fk_idwebsite; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY pages
    ADD CONSTRAINT fk_idwebsite FOREIGN KEY (id_website) REFERENCES websites(id);


--
-- Name pages_url_idx; Type: INDEX; Schema: public; Owren:-
--

CREATE INDEX pages_url_idx ON PAGES USING btree(url);

--
-- Name: fk_idwork; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY pages_words
    ADD CONSTRAINT fk_idwork FOREIGN KEY (idword) REFERENCES words(id);


--
-- Name: public; Type: ACL; Schema: -; Owner: -
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

