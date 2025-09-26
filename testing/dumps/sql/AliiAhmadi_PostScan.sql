-- Downloaded from: https://github.com/AliiAhmadi/PostScan/blob/27620d20ed609904a16cef08af33ed8f16b984f9/backup.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 16.3
-- Dumped by pg_dump version 16.3

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
-- Name: malicious_function(); Type: FUNCTION; Schema: public; Owner: testuser
--

CREATE FUNCTION public.malicious_function() RETURNS void
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
    BEGIN
      PERFORM pg_sleep(10);
      END;
      $$;


ALTER FUNCTION public.malicious_function() OWNER TO testuser;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: test; DROP TABLE users;; Type: TABLE; Schema: public; Owner: testuser
--

CREATE TABLE public."test; DROP TABLE users;" (
    id integer NOT NULL
);


ALTER TABLE public."test; DROP TABLE users;" OWNER TO testuser;

--
-- Name: test; DROP TABLE users;_id_seq; Type: SEQUENCE; Schema: public; Owner: testuser
--

CREATE SEQUENCE public."test; DROP TABLE users;_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public."test; DROP TABLE users;_id_seq" OWNER TO testuser;

--
-- Name: test; DROP TABLE users;_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: testuser
--

ALTER SEQUENCE public."test; DROP TABLE users;_id_seq" OWNED BY public."test; DROP TABLE users;".id;


--
-- Name: test_table; Type: TABLE; Schema: public; Owner: testuser
--

CREATE TABLE public.test_table (
    id integer NOT NULL
);


ALTER TABLE public.test_table OWNER TO testuser;

--
-- Name: test_table_id_seq; Type: SEQUENCE; Schema: public; Owner: testuser
--

CREATE SEQUENCE public.test_table_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.test_table_id_seq OWNER TO testuser;

--
-- Name: test_table_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: testuser
--

ALTER SEQUENCE public.test_table_id_seq OWNED BY public.test_table.id;


--
-- Name: test; DROP TABLE users; id; Type: DEFAULT; Schema: public; Owner: testuser
--

ALTER TABLE ONLY public."test; DROP TABLE users;" ALTER COLUMN id SET DEFAULT nextval('public."test; DROP TABLE users;_id_seq"'::regclass);


--
-- Name: test_table id; Type: DEFAULT; Schema: public; Owner: testuser
--

ALTER TABLE ONLY public.test_table ALTER COLUMN id SET DEFAULT nextval('public.test_table_id_seq'::regclass);


--
-- Data for Name: test; DROP TABLE users;; Type: TABLE DATA; Schema: public; Owner: testuser
--

COPY public."test; DROP TABLE users;" (id) FROM stdin;
\.


--
-- Data for Name: test_table; Type: TABLE DATA; Schema: public; Owner: testuser
--

COPY public.test_table (id) FROM stdin;
\.


--
-- Name: test; DROP TABLE users;_id_seq; Type: SEQUENCE SET; Schema: public; Owner: testuser
--

SELECT pg_catalog.setval('public."test; DROP TABLE users;_id_seq"', 1, false);


--
-- Name: test_table_id_seq; Type: SEQUENCE SET; Schema: public; Owner: testuser
--

SELECT pg_catalog.setval('public.test_table_id_seq', 1, false);


--
-- Name: test_table test_table_pkey; Type: CONSTRAINT; Schema: public; Owner: testuser
--

ALTER TABLE ONLY public.test_table
    ADD CONSTRAINT test_table_pkey PRIMARY KEY (id);


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: pg_database_owner
--

GRANT ALL ON SCHEMA public TO testuser;


--
-- Name: DEFAULT PRIVILEGES FOR TABLES; Type: DEFAULT ACL; Schema: public; Owner: postgres
--

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT ALL ON TABLES TO testuser;


--
-- PostgreSQL database dump complete
--

