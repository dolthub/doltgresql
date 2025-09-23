-- Downloaded from: https://github.com/by-zah/universityScheduleBot/blob/c8e0034ca23d8f0cac3c0ae6d5f4968ebb4be207/emptyDump.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 12.3 (Ubuntu 12.3-1.pgdg16.04+1)
-- Dumped by pg_dump version 12.4

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

DROP DATABASE IF EXISTS d7tmbubts2qa9a;
--
-- Name: d7tmbubts2qa9a; Type: DATABASE; Schema: -; Owner: hayrdyszkbigyi
--

CREATE DATABASE d7tmbubts2qa9a WITH TEMPLATE = template0 ENCODING = 'UTF8' LC_COLLATE = 'en_US.UTF-8' LC_CTYPE = 'en_US.UTF-8';


ALTER DATABASE d7tmbubts2qa9a OWNER TO hayrdyszkbigyi;

\connect d7tmbubts2qa9a

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
-- Name: public; Type: SCHEMA; Schema: -; Owner: hayrdyszkbigyi
--

CREATE SCHEMA public;


ALTER SCHEMA public OWNER TO hayrdyszkbigyi;

--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: hayrdyszkbigyi
--

COMMENT ON SCHEMA public IS 'standard public schema';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: classes; Type: TABLE; Schema: public; Owner: hayrdyszkbigyi
--

CREATE TABLE public.classes (
    index integer NOT NULL,
    group_name character varying(255) NOT NULL,
    day integer NOT NULL,
    building character varying(255),
    name character varying(255),
    room_number character varying(255)
);


ALTER TABLE public.classes OWNER TO hayrdyszkbigyi;

--
-- Name: groups; Type: TABLE; Schema: public; Owner: hayrdyszkbigyi
--

CREATE TABLE public.groups (
    name character varying(255) NOT NULL,
    owner_id integer
);


ALTER TABLE public.groups OWNER TO hayrdyszkbigyi;

--
-- Name: schedule; Type: TABLE; Schema: public; Owner: hayrdyszkbigyi
--

CREATE TABLE public.schedule (
    index integer NOT NULL,
    end_hour integer,
    end_min integer,
    start_hour integer,
    start_min integer
);


ALTER TABLE public.schedule OWNER TO hayrdyszkbigyi;

--
-- Name: subscriptions; Type: TABLE; Schema: public; Owner: hayrdyszkbigyi
--

CREATE TABLE public.subscriptions (
    user_chat_id bigint NOT NULL,
    "group" character varying(255) NOT NULL
);


ALTER TABLE public.subscriptions OWNER TO hayrdyszkbigyi;

--
-- Name: users; Type: TABLE; Schema: public; Owner: hayrdyszkbigyi
--

CREATE TABLE public.users (
    id integer NOT NULL,
    chat_id bigint NOT NULL,
    interfaculty_discipline character varying(255),
    local character varying(255),
    is_supper boolean DEFAULT false
);


ALTER TABLE public.users OWNER TO hayrdyszkbigyi;

--
-- Data for Name: classes; Type: TABLE DATA; Schema: public; Owner: hayrdyszkbigyi
--

COPY public.classes (index, group_name, day, building, name, room_number) FROM stdin;
\.


--
-- Data for Name: groups; Type: TABLE DATA; Schema: public; Owner: hayrdyszkbigyi
--

COPY public.groups (name, owner_id) FROM stdin;
\.


--
-- Data for Name: schedule; Type: TABLE DATA; Schema: public; Owner: hayrdyszkbigyi
--

COPY public.schedule (index, end_hour, end_min, start_hour, start_min) FROM stdin;
\.


--
-- Data for Name: subscriptions; Type: TABLE DATA; Schema: public; Owner: hayrdyszkbigyi
--

COPY public.subscriptions (user_chat_id, "group") FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: hayrdyszkbigyi
--

COPY public.users (id, chat_id, interfaculty_discipline, local, is_supper) FROM stdin;
\.


--
-- Name: classes classes_pkey; Type: CONSTRAINT; Schema: public; Owner: hayrdyszkbigyi
--

ALTER TABLE ONLY public.classes
    ADD CONSTRAINT classes_pkey PRIMARY KEY (index, group_name, day);


--
-- Name: groups groups_pkey; Type: CONSTRAINT; Schema: public; Owner: hayrdyszkbigyi
--

ALTER TABLE ONLY public.groups
    ADD CONSTRAINT groups_pkey PRIMARY KEY (name);


--
-- Name: schedule schedule_pkey; Type: CONSTRAINT; Schema: public; Owner: hayrdyszkbigyi
--

ALTER TABLE ONLY public.schedule
    ADD CONSTRAINT schedule_pkey PRIMARY KEY (index);


--
-- Name: subscriptions subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: hayrdyszkbigyi
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (user_chat_id, "group");


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: hayrdyszkbigyi
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users_chat_id_uindex; Type: INDEX; Schema: public; Owner: hayrdyszkbigyi
--

CREATE UNIQUE INDEX users_chat_id_uindex ON public.users USING btree (chat_id);


--
-- Name: classes classes_groups_name_fk; Type: FK CONSTRAINT; Schema: public; Owner: hayrdyszkbigyi
--

ALTER TABLE ONLY public.classes
    ADD CONSTRAINT classes_groups_name_fk FOREIGN KEY (group_name) REFERENCES public.groups(name) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: classes classes_schedule_index_fk; Type: FK CONSTRAINT; Schema: public; Owner: hayrdyszkbigyi
--

ALTER TABLE ONLY public.classes
    ADD CONSTRAINT classes_schedule_index_fk FOREIGN KEY (index) REFERENCES public.schedule(index);


--
-- Name: groups groups_users_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: hayrdyszkbigyi
--

ALTER TABLE ONLY public.groups
    ADD CONSTRAINT groups_users_id_fk FOREIGN KEY (owner_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: subscriptions subscriptions_groups_name_fk; Type: FK CONSTRAINT; Schema: public; Owner: hayrdyszkbigyi
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_groups_name_fk FOREIGN KEY ("group") REFERENCES public.groups(name) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: subscriptions subscriptions_users_chat_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: hayrdyszkbigyi
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_users_chat_id_fk FOREIGN KEY (user_chat_id) REFERENCES public.users(chat_id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: DATABASE d7tmbubts2qa9a; Type: ACL; Schema: -; Owner: hayrdyszkbigyi
--

REVOKE CONNECT,TEMPORARY ON DATABASE d7tmbubts2qa9a FROM PUBLIC;


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: hayrdyszkbigyi
--

REVOKE ALL ON SCHEMA public FROM postgres;
REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO hayrdyszkbigyi;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- Name: LANGUAGE plpgsql; Type: ACL; Schema: -; Owner: postgres
--

GRANT ALL ON LANGUAGE plpgsql TO hayrdyszkbigyi;


--
-- PostgreSQL database dump complete
--

