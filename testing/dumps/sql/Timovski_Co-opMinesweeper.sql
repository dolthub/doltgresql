-- Downloaded from: https://github.com/Timovski/Co-opMinesweeper/blob/c5db38d9c758b7121c0fc01e0bfedead718e709c/db.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 11.1
-- Dumped by pg_dump version 11.1

-- Started on 2019-02-20 21:09:05

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 198 (class 1255 OID 16458)
-- Name: create_game(character varying); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.create_game(INOUT new_host_connection_id character varying)
    LANGUAGE sql
    AS $$
	WITH new_game_id_holder (new_game_id) AS (
		SELECT n.random_number
		FROM (
			SELECT LPAD(FLOOR(random() * 10000)::varchar, 4, '0') AS random_number
			FROM generate_series(1, (SELECT COUNT(*) FROM public.games) + 10)
		) AS n
		LEFT OUTER JOIN 
			public.games AS g on g.game_id = n.random_number
		WHERE g.id IS NULL
		LIMIT 1
	)
	INSERT INTO public.games (
		game_id,
		host_connection_id
	)
	VALUES ( 
		(SELECT new_game_id FROM new_game_id_holder),
		new_host_connection_id
	)
	RETURNING game_id;
$$;


ALTER PROCEDURE public.create_game(INOUT new_host_connection_id character varying) OWNER TO postgres;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- TOC entry 196 (class 1259 OID 16449)
-- Name: games; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.games (
    id bigint NOT NULL,
    game_id character varying(4) NOT NULL,
    host_connection_id character varying(50) NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.games OWNER TO postgres;

--
-- TOC entry 197 (class 1259 OID 16453)
-- Name: games_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.games_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.games_id_seq OWNER TO postgres;

--
-- TOC entry 2816 (class 0 OID 0)
-- Dependencies: 197
-- Name: games_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.games_id_seq OWNED BY public.games.id;


--
-- TOC entry 2687 (class 2604 OID 16455)
-- Name: games id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.games ALTER COLUMN id SET DEFAULT nextval('public.games_id_seq'::regclass);


--
-- TOC entry 2689 (class 2606 OID 16457)
-- Name: games games_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.games
    ADD CONSTRAINT games_pkey PRIMARY KEY (id);


-- Completed on 2019-02-20 21:09:05

--
-- PostgreSQL database dump complete
--

