-- Downloaded from: https://github.com/joec05/social-media-app-pgsql/blob/a25959893494eaffd03deac61dabb08582cabf56/users_chats.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 16rc1
-- Dumped by pg_dump version 16rc1

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
-- Name: group_messages; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA group_messages;


ALTER SCHEMA group_messages OWNER to postgres;

--
-- Name: group_profile; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA group_profile;


ALTER SCHEMA group_profile OWNER to postgres;

--
-- Name: private_messages; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA private_messages;


ALTER SCHEMA private_messages OWNER to postgres;

--
-- Name: users_chats; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA users_chats;


ALTER SCHEMA users_chats OWNER to postgres;

--
-- Name: dblink; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS dblink WITH SCHEMA public;


--
-- Name: EXTENSION dblink; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION dblink IS 'connect to other PostgreSQL databases from within a database';


--
-- Name: fetch_user_chats(text, integer, integer, text, text, integer, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fetch_user_chats(currentid text, currentlength integer, paginationlimit integer, username text, ip text, port integer, password text) RETURNS TABLE(chat_data json)
    LANGUAGE plpgsql
    AS $_$
declare 
begin
	return query select row_to_json(f) from users_chats.chats_history as f
	where f.user_id = $1 and f.deleted = false and (f.recipient = '' or
	(f.recipient != '' and not is_blocked_user($1, f.recipient, username, ip, port, password)
	and not is_blocked_user(f.recipient, $1, username, ip, port, password) 
	and is_exists_user($1, f.recipient, username, ip, port, password)))
	offset $2 limit $3;
end;
$_$;


ALTER FUNCTION public.fetch_user_chats(currentid text, currentlength integer, paginationlimit integer, username text, ip text, port integer, password text) OWNER TO postgres;

--
-- Name: is_blocked_user(text, text, text, text, integer, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.is_blocked_user(checkingid text, checkedid text, username text, ip text, port integer, password text) RETURNS boolean
    LANGUAGE plpgsql
    AS $_$
declare 
res bool;
users_profiles_path text := 'dbname=users_profiles user='||username||' hostaddr='||ip||' port='||port||' password='||password;
begin	
	if checkingid = checkedid then return false; end if;
	select exists (
		select * from dblink(users_profiles_path, 'select * from blocked_users.block_history') as b(user_id text, blocked_id text)        
		where b.user_id = $1 and b.blocked_id = $2
	) into res;
	return res;
end;
$_$;


ALTER FUNCTION public.is_blocked_user(checkingid text, checkedid text, username text, ip text, port integer, password text) OWNER TO postgres;

--
-- Name: is_exists_user(text, text, text, text, integer, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.is_exists_user(checkingid text, checkedid text, username text, ip text, port integer, password text) RETURNS boolean
    LANGUAGE plpgsql
    AS $_$
declare 
res bool;
users_profiles_path text := 'dbname=users_profiles user='||username||' hostaddr='||ip||' port='||port||' password='||password;
begin	
	if checkingid = checkedid then return true; end if;
	select exists (
		select * from dblink(users_profiles_path, 'select user_id, deleted, suspended from basic_data.user_profile') as pr(user_id text, deleted bool, suspended bool)        
		where pr.user_id = $2 and pr.deleted = false and pr.suspended = false
	) into res;
	return res;
end;
$_$;


ALTER FUNCTION public.is_exists_user(checkingid text, checkedid text, username text, ip text, port integer, password text) OWNER TO postgres;

--
-- Name: is_muted_user(text, text, text, text, integer, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.is_muted_user(checkingid text, checkedid text, username text, ip text, port integer, password text) RETURNS boolean
    LANGUAGE plpgsql
    AS $_$
declare 
res bool;
users_profiles_path text := 'dbname=users_profiles user='||username||' hostaddr='||ip||' port='||port||' password='||password;
begin	
	if checkingid = checkedid then return false; end if;
	select exists (
		select * from dblink(users_profiles_path, 'select * from muted_users.mute_history') as m(user_id text, muted_id text)
		where m.user_id = $1 and m.muted_id = $2
	) into res;
	return res;
end;
$_$;


ALTER FUNCTION public.is_muted_user(checkingid text, checkedid text, username text, ip text, port integer, password text) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: messages_history; Type: TABLE; Schema: group_messages; Owner: postgres
--

CREATE TABLE group_messages.messages_history (
    chat_id text NOT NULL,
    message_id text NOT NULL,
    type text NOT NULL,
    content text NOT NULL,
    sender text NOT NULL,
    upload_time text NOT NULL,
    medias_datas text NOT NULL,
    deleted_list text[] NOT NULL
);


ALTER TABLE group_messages.messages_history OWNER TO postgres;

--
-- Name: group_info; Type: TABLE; Schema: group_profile; Owner: postgres
--

CREATE TABLE group_profile.group_info (
    chat_id text NOT NULL,
    name text NOT NULL,
    profile_pic_link text NOT NULL,
    description text NOT NULL,
    members text[] NOT NULL
);


ALTER TABLE group_profile.group_info OWNER TO postgres;

--
-- Name: messages_history; Type: TABLE; Schema: private_messages; Owner: postgres
--

CREATE TABLE private_messages.messages_history (
    chat_id text NOT NULL,
    message_id text NOT NULL,
    type text NOT NULL,
    content text NOT NULL,
    sender text NOT NULL,
    upload_time text NOT NULL,
    medias_datas text NOT NULL,
    deleted_list text[] NOT NULL
);


ALTER TABLE private_messages.messages_history OWNER TO postgres;

--
-- Name: chats_history; Type: TABLE; Schema: users_chats; Owner: postgres
--

CREATE TABLE users_chats.chats_history (
    user_id text NOT NULL,
    chat_id text NOT NULL,
    type text NOT NULL,
    recipient text NOT NULL,
    deleted boolean NOT NULL
);


ALTER TABLE users_chats.chats_history OWNER TO postgres;

--
-- Data for Name: messages_history; Type: TABLE DATA; Schema: group_messages; Owner: postgres
--

COPY group_messages.messages_history (chat_id, message_id, type, content, sender, upload_time, medias_datas, deleted_list) FROM stdin;
\.


--
-- Data for Name: group_info; Type: TABLE DATA; Schema: group_profile; Owner: postgres
--

COPY group_profile.group_info (chat_id, name, profile_pic_link, description, members) FROM stdin;
\.


--
-- Data for Name: messages_history; Type: TABLE DATA; Schema: private_messages; Owner: postgres
--

COPY private_messages.messages_history (chat_id, message_id, type, content, sender, upload_time, medias_datas, deleted_list) FROM stdin;
\.


--
-- Data for Name: chats_history; Type: TABLE DATA; Schema: users_chats; Owner: postgres
--

COPY users_chats.chats_history (user_id, chat_id, type, recipient, deleted) FROM stdin;
\.


--
-- Name: messages_history group_messages_constraint; Type: CONSTRAINT; Schema: group_messages; Owner: postgres
--

ALTER TABLE ONLY group_messages.messages_history
    ADD CONSTRAINT group_messages_constraint UNIQUE (message_id);


--
-- Name: messages_history messages_history_pkey; Type: CONSTRAINT; Schema: group_messages; Owner: postgres
--

ALTER TABLE ONLY group_messages.messages_history
    ADD CONSTRAINT messages_history_pkey PRIMARY KEY (message_id);


--
-- Name: group_info group_info_constraints; Type: CONSTRAINT; Schema: group_profile; Owner: postgres
--

ALTER TABLE ONLY group_profile.group_info
    ADD CONSTRAINT group_info_constraints UNIQUE (chat_id);


--
-- Name: group_info group_info_pkey; Type: CONSTRAINT; Schema: group_profile; Owner: postgres
--

ALTER TABLE ONLY group_profile.group_info
    ADD CONSTRAINT group_info_pkey PRIMARY KEY (chat_id);


--
-- Name: messages_history messages_history_pkey; Type: CONSTRAINT; Schema: private_messages; Owner: postgres
--

ALTER TABLE ONLY private_messages.messages_history
    ADD CONSTRAINT messages_history_pkey PRIMARY KEY (message_id);


--
-- Name: messages_history private_messages_constraints; Type: CONSTRAINT; Schema: private_messages; Owner: postgres
--

ALTER TABLE ONLY private_messages.messages_history
    ADD CONSTRAINT private_messages_constraints UNIQUE (message_id);


--
-- Name: chats_history chats_constraint; Type: CONSTRAINT; Schema: users_chats; Owner: postgres
--

ALTER TABLE ONLY users_chats.chats_history
    ADD CONSTRAINT chats_constraint UNIQUE (user_id, chat_id);


--
-- Name: chats_history chats_primary_key; Type: CONSTRAINT; Schema: users_chats; Owner: postgres
--

ALTER TABLE ONLY users_chats.chats_history
    ADD CONSTRAINT chats_primary_key PRIMARY KEY (user_id, chat_id);


--
-- PostgreSQL database dump complete
--

