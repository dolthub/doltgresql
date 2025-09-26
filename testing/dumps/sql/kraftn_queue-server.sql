-- Downloaded from: https://github.com/kraftn/queue-server/blob/f88a3bc49507714647e953b70818dcd0cb1c80df/database/backup.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 11.4
-- Dumped by pg_dump version 11.4

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
-- Name: get_top(character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_top(find_type character varying) RETURNS TABLE(id bigint, kind bigint, status character varying, input text, output text)
    LANGUAGE sql
    AS $$select * from main_queue 
where id = (select min(id) from main_queue
where kind in (select id from task_types 
where kind = find_type) and status = 'В очереди')$$;


ALTER FUNCTION public.get_top(find_type character varying) OWNER TO postgres;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: main_queue; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.main_queue (
    id bigint NOT NULL,
    kind bigint NOT NULL,
    status character varying(100) NOT NULL,
    input text NOT NULL,
    output text
);


ALTER TABLE public.main_queue OWNER TO postgres;

--
-- Name: main_queue_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.main_queue_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.main_queue_id_seq OWNER TO postgres;

--
-- Name: main_queue_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.main_queue_id_seq OWNED BY public.main_queue.id;


--
-- Name: task_types; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.task_types (
    id bigint NOT NULL,
    kind character varying(100) NOT NULL
);


ALTER TABLE public.task_types OWNER TO postgres;

--
-- Name: task_types_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.task_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.task_types_id_seq OWNER TO postgres;

--
-- Name: task_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.task_types_id_seq OWNED BY public.task_types.id;


--
-- Name: main_queue id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.main_queue ALTER COLUMN id SET DEFAULT nextval('public.main_queue_id_seq'::regclass);


--
-- Name: task_types id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.task_types ALTER COLUMN id SET DEFAULT nextval('public.task_types_id_seq'::regclass);


--
-- Data for Name: main_queue; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.main_queue (id, kind, status, input, output) FROM stdin;
\.


--
-- Data for Name: task_types; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.task_types (id, kind) FROM stdin;
11	E-mail
7	Квадратное уравнение
10	Перевод
\.


--
-- Name: main_queue_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.main_queue_id_seq', 14693, true);


--
-- Name: task_types_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.task_types_id_seq', 14, true);


--
-- Name: main_queue main_queue_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.main_queue
    ADD CONSTRAINT main_queue_pkey PRIMARY KEY (id);


--
-- Name: task_types task_types_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.task_types
    ADD CONSTRAINT task_types_pkey PRIMARY KEY (id);


--
-- Name: main_queue main_queue_type_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.main_queue
    ADD CONSTRAINT main_queue_type_fkey FOREIGN KEY (kind) REFERENCES public.task_types(id);


--
-- PostgreSQL database dump complete
--

