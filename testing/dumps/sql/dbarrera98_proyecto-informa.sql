-- Downloaded from: https://github.com/dbarrera98/proyecto-informa/blob/f6d6e700a5f20ffa272f728a83af6eb3683c3b83/initdb/init.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 16.1
-- Dumped by pg_dump version 16.1

-- Started on 2025-05-31 16:18:14

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
-- TOC entry 224 (class 1255 OID 24614)
-- Name: actualizar_autor(integer, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.actualizar_autor(p_id integer, p_nombre character varying) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_count FROM autores a WHERE a.id = p_id;
    IF v_count = 0 THEN
        RAISE EXCEPTION 'Autor no existe';
    END IF;
    UPDATE autores a SET nombre = p_nombre WHERE a.id = p_id;
END;
$$;


ALTER FUNCTION public.actualizar_autor(p_id integer, p_nombre character varying) OWNER TO postgres;

--
-- TOC entry 225 (class 1255 OID 24610)
-- Name: actualizar_libro(integer, character varying, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.actualizar_libro(p_id integer, p_titulo character varying, p_autor_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_count FROM autores a WHERE a.id = p_autor_id;
    IF v_count = 0 THEN
        RAISE EXCEPTION 'Autor no existe';
    END IF;
    SELECT COUNT(*) INTO v_count FROM libros l WHERE l.id = p_id;
    IF v_count = 0 THEN
        RAISE EXCEPTION 'Libro no existe';
    END IF;
    UPDATE libros l SET titulo = p_titulo, autor_id = p_autor_id WHERE l.id = p_id;
END;
$$;


ALTER FUNCTION public.actualizar_libro(p_id integer, p_titulo character varying, p_autor_id integer) OWNER TO postgres;

--
-- TOC entry 223 (class 1255 OID 24616)
-- Name: consultar_autor(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.consultar_autor(p_id integer) RETURNS TABLE(id integer, nombre character varying)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY SELECT a.id, a.nombre FROM autores a WHERE a.id = p_id;
END;
$$;


ALTER FUNCTION public.consultar_autor(p_id integer) OWNER TO postgres;

--
-- TOC entry 222 (class 1255 OID 24617)
-- Name: consultar_autores(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.consultar_autores() RETURNS TABLE(id integer, nombre character varying)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT a.id, a.nombre FROM autores a;
END;
$$;


ALTER FUNCTION public.consultar_autores() OWNER TO postgres;

--
-- TOC entry 220 (class 1255 OID 24612)
-- Name: consultar_libro(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.consultar_libro(p_id integer) RETURNS TABLE(id integer, titulo character varying, autor_id integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY SELECT l.id, l.titulo, l.autor_id FROM libros l WHERE l.id = p_id;
END;
$$;


ALTER FUNCTION public.consultar_libro(p_id integer) OWNER TO postgres;

--
-- TOC entry 226 (class 1255 OID 24613)
-- Name: consultar_libros(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.consultar_libros() RETURNS TABLE(id integer, titulo character varying, autor_id integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY SELECT l.id, l.titulo, l.autor_id FROM libros l;
END;
$$;


ALTER FUNCTION public.consultar_libros() OWNER TO postgres;

--
-- TOC entry 232 (class 1255 OID 24615)
-- Name: eliminar_autor(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.eliminar_autor(p_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_count FROM autores a WHERE a.id = p_id;
    IF v_count = 0 THEN
        RAISE EXCEPTION 'Autor no existe';
    END IF;
    SELECT COUNT(*) INTO v_count FROM libros l WHERE l.autor_id = p_id;
    IF v_count > 0 THEN
        RAISE EXCEPTION 'No se puede eliminar el autor porque tiene libros asociados';
    END IF;
    DELETE FROM autores a WHERE a.id = p_id;
    RAISE NOTICE 'Autor eliminado exitosamente (id = %)', p_id;
END;
$$;


ALTER FUNCTION public.eliminar_autor(p_id integer) OWNER TO postgres;

--
-- TOC entry 221 (class 1255 OID 24611)
-- Name: eliminar_libro(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.eliminar_libro(p_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_count FROM libros l WHERE l.id = p_id;
    IF v_count = 0 THEN
        RAISE EXCEPTION 'Libro no existe';
    END IF;
    DELETE FROM libros l WHERE l.id = p_id;
END;
$$;


ALTER FUNCTION public.eliminar_libro(p_id integer) OWNER TO postgres;

--
-- TOC entry 219 (class 1255 OID 24608)
-- Name: insertar_autor(character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.insertar_autor(p_nombre character varying) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    INSERT INTO autores (nombre) VALUES (p_nombre);
END;
$$;


ALTER FUNCTION public.insertar_autor(p_nombre character varying) OWNER TO postgres;

--
-- TOC entry 227 (class 1255 OID 24609)
-- Name: insertar_libro(character varying, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.insertar_libro(p_titulo character varying, p_autor_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_count FROM autores a WHERE a.id = p_autor_id;
    IF v_count = 0 THEN
        RAISE EXCEPTION 'Autor no existe';
    END IF;
    INSERT INTO libros (titulo, autor_id) VALUES (p_titulo, p_autor_id);
END;
$$;


ALTER FUNCTION public.insertar_libro(p_titulo character varying, p_autor_id integer) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 216 (class 1259 OID 24592)
-- Name: autores; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.autores (
    id integer NOT NULL,
    nombre character varying(100) NOT NULL
);


ALTER TABLE public.autores OWNER TO postgres;

--
-- TOC entry 215 (class 1259 OID 24591)
-- Name: autores_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.autores ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.autores_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 218 (class 1259 OID 24598)
-- Name: libros; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.libros (
    id integer NOT NULL,
    titulo character varying(200) NOT NULL,
    autor_id integer NOT NULL
);


ALTER TABLE public.libros OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 24597)
-- Name: libros_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.libros ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.libros_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 4798 (class 0 OID 24592)
-- Dependencies: 216
-- Data for Name: autores; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.autores (id, nombre) FROM stdin;
8	Camilo
5	Andres
12	William
13	Jane
11	Cervantes
14	Mary
15	Laura
\.


--
-- TOC entry 4800 (class 0 OID 24598)
-- Dependencies: 218
-- Data for Name: libros; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.libros (id, titulo, autor_id) FROM stdin;
4	Don Quijote de la Mancha	11
6	Frankenstein	14
8	Los caracoles	15
\.


--
-- TOC entry 4806 (class 0 OID 0)
-- Dependencies: 215
-- Name: autores_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.autores_id_seq', 15, true);


--
-- TOC entry 4807 (class 0 OID 0)
-- Dependencies: 217
-- Name: libros_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.libros_id_seq', 8, true);


--
-- TOC entry 4650 (class 2606 OID 24596)
-- Name: autores autores_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.autores
    ADD CONSTRAINT autores_pkey PRIMARY KEY (id);


--
-- TOC entry 4652 (class 2606 OID 24602)
-- Name: libros libros_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.libros
    ADD CONSTRAINT libros_pkey PRIMARY KEY (id);


--
-- TOC entry 4653 (class 2606 OID 24603)
-- Name: libros libros_autor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.libros
    ADD CONSTRAINT libros_autor_id_fkey FOREIGN KEY (autor_id) REFERENCES public.autores(id);


-- Completed on 2025-05-31 16:18:17

--
-- PostgreSQL database dump complete
--

