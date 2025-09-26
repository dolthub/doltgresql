-- Downloaded from: https://github.com/rmarquez123/titans/blob/f498987acd658bfd6555a2026b4f0c3ae24872b6/docker/terrabyte.database/titans.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 10.6
-- Dumped by pg_dump version 15.2

-- Started on 2024-01-23 17:49:56

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
-- TOC entry 10 (class 2615 OID 1128245)
-- Name: postgis; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA postgis;


ALTER SCHEMA postgis OWNER TO postgres;

--
-- TOC entry 11 (class 2615 OID 1137946)
-- Name: projects; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA projects;


ALTER SCHEMA projects OWNER TO postgres;

--
-- TOC entry 7 (class 2615 OID 2200)
-- Name: public; Type: SCHEMA; Schema: -; Owner: postgres
--

-- *not* creating schema, since initdb creates it


ALTER SCHEMA public OWNER TO postgres;

--
-- TOC entry 9 (class 2615 OID 1125424)
-- Name: users; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA users;


ALTER SCHEMA users OWNER TO postgres;

--
-- TOC entry 2 (class 3079 OID 1128246)
-- Name: postgis; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS postgis WITH SCHEMA postgis;


--
-- TOC entry 4338 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION postgis; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION postgis IS 'PostGIS geometry, geography, and raster spatial types and functions';


SET default_tablespace = '';

--
-- TOC entry 231 (class 1259 OID 1137947)
-- Name: project; Type: TABLE; Schema: projects; Owner: postgres
--

CREATE TABLE projects.project (
    project_id integer NOT NULL,
    name character varying NOT NULL
);


ALTER TABLE projects.project OWNER TO postgres;

--
-- TOC entry 232 (class 1259 OID 1137955)
-- Name: project_envelope; Type: TABLE; Schema: projects; Owner: postgres
--

CREATE TABLE projects.project_envelope (
    project_id integer NOT NULL,
    lowerleft point NOT NULL,
    upperright point NOT NULL,
    srid integer NOT NULL
);


ALTER TABLE projects.project_envelope OWNER TO postgres;

--
-- TOC entry 233 (class 1259 OID 1137965)
-- Name: projectdatasource; Type: TABLE; Schema: projects; Owner: postgres
--

CREATE TABLE projects.projectdatasource (
    project_id integer NOT NULL,
    rastergroup_id integer NOT NULL
);


ALTER TABLE projects.projectdatasource OWNER TO postgres;

--
-- TOC entry 207 (class 1259 OID 1125447)
-- Name: raster_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.raster_id_seq
    START WITH 0
    INCREMENT BY 1
    MINVALUE 0
    MAXVALUE 1000000
    CACHE 1;


ALTER TABLE public.raster_id_seq OWNER TO postgres;

--
-- TOC entry 208 (class 1259 OID 1125450)
-- Name: rastergeomproperties_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.rastergeomproperties_id_seq
    START WITH 0
    INCREMENT BY 1
    MINVALUE 0
    MAXVALUE 1000000
    CACHE 1;


ALTER TABLE public.rastergeomproperties_id_seq OWNER TO postgres;

--
-- TOC entry 202 (class 1259 OID 1125392)
-- Name: raster; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.raster (
    raster_id integer DEFAULT nextval('public.raster_id_seq'::regclass) NOT NULL,
    rastertype_id integer NOT NULL,
    source_id integer NOT NULL,
    rastergeomproperties_id integer DEFAULT nextval('public.rastergeomproperties_id_seq'::regclass) NOT NULL
);


ALTER TABLE public.raster OWNER TO postgres;

--
-- TOC entry 203 (class 1259 OID 1125397)
-- Name: rastergeomproperties; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.rastergeomproperties (
    rastergeomproperties_id integer DEFAULT nextval('public.rastergeomproperties_id_seq'::regclass) NOT NULL,
    dx double precision NOT NULL,
    dy double precision NOT NULL,
    lowerleft point NOT NULL,
    upperright point NOT NULL,
    srid integer NOT NULL
);


ALTER TABLE public.rastergeomproperties OWNER TO postgres;

--
-- TOC entry 209 (class 1259 OID 1125455)
-- Name: rastergroup_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.rastergroup_id_seq
    START WITH 0
    INCREMENT BY 1
    MINVALUE 0
    MAXVALUE 1000000
    CACHE 1;


ALTER TABLE public.rastergroup_id_seq OWNER TO postgres;

--
-- TOC entry 204 (class 1259 OID 1125419)
-- Name: rastergroup; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.rastergroup (
    rastergroup_id integer DEFAULT nextval('public.rastergroup_id_seq'::regclass) NOT NULL,
    name character varying(200) NOT NULL
);


ALTER TABLE public.rastergroup OWNER TO postgres;

--
-- TOC entry 210 (class 1259 OID 1125462)
-- Name: rastergroup_by_user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.rastergroup_by_user_id_seq
    START WITH 0
    INCREMENT BY 1
    MINVALUE 0
    MAXVALUE 1000000
    CACHE 1;


ALTER TABLE public.rastergroup_by_user_id_seq OWNER TO postgres;

--
-- TOC entry 206 (class 1259 OID 1125430)
-- Name: rastergroup_by_user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.rastergroup_by_user (
    rastergroup_by_user_id integer DEFAULT nextval('public.rastergroup_by_user_id_seq'::regclass) NOT NULL,
    rastergroup_id integer NOT NULL,
    user_id integer NOT NULL
);


ALTER TABLE public.rastergroup_by_user OWNER TO postgres;

--
-- TOC entry 214 (class 1259 OID 1125476)
-- Name: rastergroup_raster_link; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.rastergroup_raster_link (
    rastergroup_id integer NOT NULL,
    raster_id integer NOT NULL
);


ALTER TABLE public.rastergroup_raster_link OWNER TO postgres;

--
-- TOC entry 211 (class 1259 OID 1125465)
-- Name: rastertype_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.rastertype_id_seq
    START WITH 0
    INCREMENT BY 1
    MINVALUE 0
    MAXVALUE 1000000
    CACHE 1;


ALTER TABLE public.rastertype_id_seq OWNER TO postgres;

--
-- TOC entry 200 (class 1259 OID 1125372)
-- Name: rastertype; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.rastertype (
    rastertype_id integer DEFAULT nextval('public.rastertype_id_seq'::regclass) NOT NULL,
    name character varying(200) NOT NULL
);


ALTER TABLE public.rastertype OWNER TO postgres;

--
-- TOC entry 212 (class 1259 OID 1125468)
-- Name: source_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.source_id_seq
    START WITH 0
    INCREMENT BY 1
    MINVALUE 0
    MAXVALUE 1000000
    CACHE 1;


ALTER TABLE public.source_id_seq OWNER TO postgres;

--
-- TOC entry 201 (class 1259 OID 1125377)
-- Name: source; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.source (
    source_id integer DEFAULT nextval('public.source_id_seq'::regclass) NOT NULL,
    title character varying(200) NOT NULL,
    description character varying(200)
);


ALTER TABLE public.source OWNER TO postgres;

--
-- TOC entry 230 (class 1259 OID 1129754)
-- Name: authentication; Type: TABLE; Schema: users; Owner: postgres
--

CREATE TABLE users.authentication (
    user_id integer NOT NULL,
    key character varying(200) NOT NULL
);


ALTER TABLE users.authentication OWNER TO postgres;

--
-- TOC entry 213 (class 1259 OID 1125473)
-- Name: user_id_seq; Type: SEQUENCE; Schema: users; Owner: postgres
--

CREATE SEQUENCE users.user_id_seq
    START WITH 0
    INCREMENT BY 1
    MINVALUE 0
    MAXVALUE 1000000
    CACHE 1;


ALTER TABLE users.user_id_seq OWNER TO postgres;

--
-- TOC entry 205 (class 1259 OID 1125425)
-- Name: user; Type: TABLE; Schema: users; Owner: postgres
--

CREATE TABLE users."user" (
    user_id integer DEFAULT nextval('users.user_id_seq'::regclass) NOT NULL,
    name character varying(200) NOT NULL,
    email character varying(200)
);


ALTER TABLE users."user" OWNER TO postgres;

--
-- TOC entry 4337 (class 0 OID 0)
-- Dependencies: 7
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE USAGE ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO PUBLIC;


-- Completed on 2024-01-23 17:49:56

--
-- PostgreSQL database dump complete
--

