-- Downloaded from: https://github.com/Boluwatife-AJB/backend-in-node/blob/302739ec5fb1880b77d9cb51636834bdacff16ed/sample.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 15.8 (Homebrew)
-- Dumped by pg_dump version 15.8 (Homebrew)

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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: companies; Type: TABLE; Schema: public; Owner: USER
--

CREATE TABLE public.companies (
    id uuid NOT NULL,
    name character varying
);


ALTER TABLE public.companies OWNER TO "USER";

--
-- Name: employees; Type: TABLE; Schema: public; Owner: USER
--

CREATE TABLE public.employees (
    id uuid NOT NULL,
    first_name character varying,
    last_name character varying,
    email character varying
);


ALTER TABLE public.employees OWNER TO "USER";

--
-- Data for Name: companies; Type: TABLE DATA; Schema: public; Owner: USER
--

COPY public.companies (id, name) FROM stdin;
f5c41428-5f90-4ca6-b167-9c6f5a41bae0	valhalla
7874572b-e4ca-4add-958e-3f611649c9bf	ghost road
\.


--
-- Data for Name: employees; Type: TABLE DATA; Schema: public; Owner: USER
--

COPY public.employees (id, first_name, last_name, email) FROM stdin;
ce4b3817-2ea9-4f4c-9760-8ad30aa807be	max	maximus	max@valhallah.org
ce4b3817-2ea9-4f4c-9760-8ad30aa347be	mcDonalds	williams	williams@ghost-road.org
\.


--
-- Name: companies companies_pkey; Type: CONSTRAINT; Schema: public; Owner: USER
--

ALTER TABLE ONLY public.companies
    ADD CONSTRAINT companies_pkey PRIMARY KEY (id);


--
-- Name: employees employees_pkey; Type: CONSTRAINT; Schema: public; Owner: USER
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_pkey PRIMARY KEY (id);


--
-- PostgreSQL database dump complete
--

