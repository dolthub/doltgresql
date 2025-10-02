-- Downloaded from: https://github.com/linvivian7/fe-react-16-demo/blob/6ababddceccfc0320ae1739fc0152335740bbadc/vocab.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 9.6.2
-- Dumped by pg_dump version 9.6.2

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
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

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: vocabulary; Type: TABLE; Schema: public; Owner: Admin
--

CREATE TABLE vocabulary (
    id integer NOT NULL,
    jp_word character varying(50) NOT NULL,
    en_word character varying(50) NOT NULL
);


ALTER TABLE vocabulary OWNER TO "Admin";

--
-- Name: vocabulary_id_seq; Type: SEQUENCE; Schema: public; Owner: Admin
--

CREATE SEQUENCE vocabulary_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE vocabulary_id_seq OWNER TO "Admin";

--
-- Name: vocabulary_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: Admin
--

ALTER SEQUENCE vocabulary_id_seq OWNED BY vocabulary.id;


--
-- Name: vocabulary id; Type: DEFAULT; Schema: public; Owner: Admin
--

ALTER TABLE ONLY vocabulary ALTER COLUMN id SET DEFAULT nextval('vocabulary_id_seq'::regclass);


--
-- Data for Name: vocabulary; Type: TABLE DATA; Schema: public; Owner: Admin
--

COPY vocabulary (id, jp_word, en_word) FROM stdin;
1	内訳	itemization
2	用件	things to be done
3	抗争	resistance
4	感染	infect
5	勝る	to excel
6	告白	confess
7	華やか	showy, brilliant
8	軽蔑	scorn, disdain
9	築く	to build
10	同居	coliving
11	 独立	independent
12	沿い	along
13	花粉	pollen
14	一頃	once
15	発作	spasm
16	器官	organs
18	液	liquid
19	先代	predecessor
\.


--
-- Name: vocabulary_id_seq; Type: SEQUENCE SET; Schema: public; Owner: Admin
--

SELECT pg_catalog.setval('vocabulary_id_seq', 19, true);


--
-- Name: vocabulary vocabulary_pkey; Type: CONSTRAINT; Schema: public; Owner: Admin
--

ALTER TABLE ONLY vocabulary
    ADD CONSTRAINT vocabulary_pkey PRIMARY KEY (id);


--
-- PostgreSQL database dump complete
--

