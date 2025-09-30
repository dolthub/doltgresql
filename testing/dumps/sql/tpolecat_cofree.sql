-- Downloaded from: https://github.com/tpolecat/cofree/blob/afe11f23ab5d11b94c180347b47ddfd3c59b6c85/create.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 9.5.3
-- Dumped by pg_dump version 9.5.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

--
-- Name: fn_array_agg_notnull(anyarray, anyelement); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION fn_array_agg_notnull(a anyarray, b anyelement) RETURNS anyarray
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN

    IF b IS NOT NULL THEN
        a := array_append(a, b);
    END IF;

    RETURN a;

END;
$$;


ALTER FUNCTION public.fn_array_agg_notnull(a anyarray, b anyelement) OWNER TO postgres;

--
-- Name: array_agg_notnull(anyelement); Type: AGGREGATE; Schema: public; Owner: postgres
--

CREATE AGGREGATE array_agg_notnull(anyelement) (
    SFUNC = fn_array_agg_notnull,
    STYPE = anyarray,
    INITCOND = '{}'
);


ALTER AGGREGATE public.array_agg_notnull(anyelement) OWNER TO postgres;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: prof; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE prof (
    id integer NOT NULL,
    parent integer,
    name character varying NOT NULL,
    year integer NOT NULL
);


ALTER TABLE prof OWNER TO postgres;

--
-- Name: prof_closure; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE prof_closure (
    id integer,
    parent integer,
    name character varying,
    uni character varying,
    year integer,
    students integer[]
);

ALTER TABLE ONLY prof_closure REPLICA IDENTITY NOTHING;


ALTER TABLE prof_closure OWNER TO postgres;

--
-- Name: prof_node; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE prof_node (
    id integer NOT NULL,
    parent integer,
    name character varying NOT NULL,
    uni character varying NOT NULL,
    year integer NOT NULL
);


ALTER TABLE prof_node OWNER TO postgres;

--
-- Name: prof_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE prof_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE prof_id_seq OWNER TO postgres;

--
-- Name: prof_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE prof_id_seq OWNED BY prof_node.id;


--
-- Name: prof_node_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE prof_node_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE prof_node_id_seq OWNER TO postgres;

--
-- Name: prof_node_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE prof_node_id_seq OWNED BY prof.id;


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY prof ALTER COLUMN id SET DEFAULT nextval('prof_node_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY prof_node ALTER COLUMN id SET DEFAULT nextval('prof_id_seq'::regclass);


--
-- Data for Name: prof; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY prof (id, parent, name, year) FROM stdin;
\.


--
-- Name: prof_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('prof_id_seq', 537, true);


--
-- Data for Name: prof_node; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY prof_node (id, parent, name, uni, year) FROM stdin;
\.


--
-- Name: prof_node_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('prof_node_id_seq', 23, true);


--
-- Name: prof_node_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY prof
    ADD CONSTRAINT prof_node_pkey PRIMARY KEY (id);


--
-- Name: prof_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY prof_node
    ADD CONSTRAINT prof_pkey PRIMARY KEY (id);


--
-- Name: _RETURN; Type: RULE; Schema: public; Owner: postgres
--

CREATE RULE "_RETURN" AS
    ON SELECT TO prof_closure DO INSTEAD  SELECT p.id,
    p.parent,
    p.name,
    p.uni,
    p.year,
    array_agg_notnull(c.id) AS students
   FROM (prof_node p
     LEFT JOIN prof_node c ON ((c.parent = p.id)))
  GROUP BY p.id;


--
-- Name: prof_parent_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY prof_node
    ADD CONSTRAINT prof_parent_fkey FOREIGN KEY (parent) REFERENCES prof_node(id);


--
-- Name: public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

