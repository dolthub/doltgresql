-- Downloaded from: https://github.com/qqtati/diplom/blob/2be609833a470cdb3fb4a6e8f68732c779b049b6/create.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 17.4
-- Dumped by pg_dump version 17.2 (Homebrew)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: pg_database_owner
--

CREATE SCHEMA public;


ALTER SCHEMA public OWNER TO pg_database_owner;

--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: pg_database_owner
--

COMMENT ON SCHEMA public IS 'standard public schema';


--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: rw_main
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_updated_at_column() OWNER TO rw_main;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: event; Type: TABLE; Schema: public; Owner: rw_main
--

CREATE TABLE public.event (
    id integer NOT NULL,
    start_time timestamp without time zone NOT NULL,
    duration integer NOT NULL,
    teacher_id integer NOT NULL,
    price numeric(10,2) NOT NULL,
    student_id integer NOT NULL,
    description text,
    approved_by_teacher boolean DEFAULT false,
    skipped boolean DEFAULT false,
    rating integer,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.event OWNER TO rw_main;

--
-- Name: event_id_seq; Type: SEQUENCE; Schema: public; Owner: rw_main
--

CREATE SEQUENCE public.event_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.event_id_seq OWNER TO rw_main;

--
-- Name: event_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: rw_main
--

ALTER SEQUENCE public.event_id_seq OWNED BY public.event.id;


--
-- Name: homework_files; Type: TABLE; Schema: public; Owner: rw_main
--

CREATE TABLE public.homework_files (
    id integer NOT NULL,
    homework_id integer NOT NULL,
    file_name character varying(255) NOT NULL,
    file_path text NOT NULL,
    uploaded_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.homework_files OWNER TO rw_main;

--
-- Name: homework_files_id_seq; Type: SEQUENCE; Schema: public; Owner: rw_main
--

CREATE SEQUENCE public.homework_files_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.homework_files_id_seq OWNER TO rw_main;

--
-- Name: homework_files_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: rw_main
--

ALTER SEQUENCE public.homework_files_id_seq OWNED BY public.homework_files.id;


--
-- Name: homeworks; Type: TABLE; Schema: public; Owner: rw_main
--

CREATE TABLE public.homeworks (
    id integer NOT NULL,
    description text NOT NULL,
    due_date timestamp without time zone NOT NULL,
    student_id integer NOT NULL,
    teacher_id integer NOT NULL,
    rating integer,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.homeworks OWNER TO rw_main;

--
-- Name: homeworks_id_seq; Type: SEQUENCE; Schema: public; Owner: rw_main
--

CREATE SEQUENCE public.homeworks_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.homeworks_id_seq OWNER TO rw_main;

--
-- Name: homeworks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: rw_main
--

ALTER SEQUENCE public.homeworks_id_seq OWNED BY public.homeworks.id;


--
-- Name: teacher_student; Type: TABLE; Schema: public; Owner: rw_main
--

CREATE TABLE public.teacher_student (
    id integer NOT NULL,
    teacher_id integer NOT NULL,
    student_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.teacher_student OWNER TO rw_main;

--
-- Name: teacher_student_id_seq; Type: SEQUENCE; Schema: public; Owner: rw_main
--

CREATE SEQUENCE public.teacher_student_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.teacher_student_id_seq OWNER TO rw_main;

--
-- Name: teacher_student_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: rw_main
--

ALTER SEQUENCE public.teacher_student_id_seq OWNED BY public.teacher_student.id;


--
-- Name: user; Type: TABLE; Schema: public; Owner: rw_main
--

CREATE TABLE public."user" (
    id integer NOT NULL,
    username character varying(255) NOT NULL,
    password character varying(255) NOT NULL,
    role integer NOT NULL,
    invite_code character varying(255),
    name character varying(255) NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public."user" OWNER TO rw_main;

--
-- Name: user_id_seq; Type: SEQUENCE; Schema: public; Owner: rw_main
--

CREATE SEQUENCE public.user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_id_seq OWNER TO rw_main;

--
-- Name: user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: rw_main
--

ALTER SEQUENCE public.user_id_seq OWNED BY public."user".id;


--
-- Name: event id; Type: DEFAULT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public.event ALTER COLUMN id SET DEFAULT nextval('public.event_id_seq'::regclass);


--
-- Name: homework_files id; Type: DEFAULT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public.homework_files ALTER COLUMN id SET DEFAULT nextval('public.homework_files_id_seq'::regclass);


--
-- Name: homeworks id; Type: DEFAULT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public.homeworks ALTER COLUMN id SET DEFAULT nextval('public.homeworks_id_seq'::regclass);


--
-- Name: teacher_student id; Type: DEFAULT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public.teacher_student ALTER COLUMN id SET DEFAULT nextval('public.teacher_student_id_seq'::regclass);


--
-- Name: user id; Type: DEFAULT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public."user" ALTER COLUMN id SET DEFAULT nextval('public.user_id_seq'::regclass);


--
-- Data for Name: event; Type: TABLE DATA; Schema: public; Owner: rw_main
--



--
-- Data for Name: homework_files; Type: TABLE DATA; Schema: public; Owner: rw_main
--

INSERT INTO public.homework_files (id, homework_id, file_name, file_path, uploaded_at) VALUES (11, 11, 'ТИ.pdf', 'uploads/homework/11_ТИ.pdf.pdf', '2025-04-22 00:48:03.430567');


--
-- Data for Name: homeworks; Type: TABLE DATA; Schema: public; Owner: rw_main
--

INSERT INTO public.homeworks (id, description, due_date, student_id, teacher_id, rating, created_at, updated_at) VALUES (9, 'Тестовое домашнее задание', '2025-04-23 00:00:00', 96, 95, NULL, '2025-04-22 00:23:08.533939', '2025-04-22 00:23:08.533939');
INSERT INTO public.homeworks (id, description, due_date, student_id, teacher_id, rating, created_at, updated_at) VALUES (10, 'Тестовое домашнее задание', '2025-04-23 00:00:00', 99, 98, NULL, '2025-04-22 00:26:20.967737', '2025-04-22 00:26:20.967737');
INSERT INTO public.homeworks (id, description, due_date, student_id, teacher_id, rating, created_at, updated_at) VALUES (11, 'test', '2025-04-30 00:00:00', 72, 71, 3, '2025-04-22 00:32:10.366758', '2025-04-22 00:42:46.13845');


--
-- Data for Name: teacher_student; Type: TABLE DATA; Schema: public; Owner: rw_main
--

INSERT INTO public.teacher_student (id, teacher_id, student_id, created_at) VALUES (8, 71, 72, '2025-04-22 00:11:02.193972');
INSERT INTO public.teacher_student (id, teacher_id, student_id, created_at) VALUES (9, 74, 75, '2025-04-22 00:15:49.048226');
INSERT INTO public.teacher_student (id, teacher_id, student_id, created_at) VALUES (10, 77, 78, '2025-04-22 00:16:16.65874');
INSERT INTO public.teacher_student (id, teacher_id, student_id, created_at) VALUES (11, 80, 81, '2025-04-22 00:18:15.892062');
INSERT INTO public.teacher_student (id, teacher_id, student_id, created_at) VALUES (12, 83, 84, '2025-04-22 00:19:08.826555');
INSERT INTO public.teacher_student (id, teacher_id, student_id, created_at) VALUES (13, 86, 87, '2025-04-22 00:21:01.39769');
INSERT INTO public.teacher_student (id, teacher_id, student_id, created_at) VALUES (14, 89, 90, '2025-04-22 00:21:58.283115');
INSERT INTO public.teacher_student (id, teacher_id, student_id, created_at) VALUES (15, 92, 93, '2025-04-22 00:22:44.642627');
INSERT INTO public.teacher_student (id, teacher_id, student_id, created_at) VALUES (16, 95, 96, '2025-04-22 00:23:08.520639');
INSERT INTO public.teacher_student (id, teacher_id, student_id, created_at) VALUES (17, 98, 99, '2025-04-22 00:26:20.951353');


--
-- Data for Name: user; Type: TABLE DATA; Schema: public; Owner: rw_main
--

INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (72, 'test2@test.ru', '80eb26791e4fd74e71cb685f219d2e8bcf96a462f70dcac12562a99eedf5ae62dd42523022b786a82f27b43a6250a4b7f1662e9fe4378b0e341d9482d395f669', 1, 'CD57F5', 'Nikita Petrov', '2025-04-22 00:11:02.184003', '2025-04-22 00:11:02.184003');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (76, 'testuser_1745280973', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '8B25AF', 'Test User', '2025-04-22 00:16:13.592259', '2025-04-22 00:16:13.592259');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (78, 'student_1745280974', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '430758', 'Test Student', '2025-04-22 00:16:15.630775', '2025-04-22 00:16:15.630775');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (80, 'teacher_1745281093', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '26D769', 'Test Teacher', '2025-04-22 00:18:13.839757', '2025-04-22 00:18:13.839757');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (82, 'testuser_1745281145', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '170230', 'Test User', '2025-04-22 00:19:05.739213', '2025-04-22 00:19:05.739213');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (84, 'student_1745281146', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '9F5AE7', 'Test Student', '2025-04-22 00:19:07.791038', '2025-04-22 00:19:07.791038');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (88, 'testuser_1745281315', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, 'AD79FF', 'Test User', '2025-04-22 00:21:55.226097', '2025-04-22 00:21:55.226097');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (89, 'teacher_1745281316', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '3017EA', 'Test Teacher', '2025-04-22 00:21:56.241584', '2025-04-22 00:21:56.241584');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (90, 'student_1745281316', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '547F38', 'Test Student', '2025-04-22 00:21:57.256868', '2025-04-22 00:21:57.256868');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (91, 'testuser_1745281361', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '5F8FD1', 'Test User', '2025-04-22 00:22:41.562072', '2025-04-22 00:22:41.562072');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (93, 'student_1745281362', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, 'D77F73', 'Test Student', '2025-04-22 00:22:43.612883', '2025-04-22 00:22:43.612883');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (97, 'testuser_1745281577', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '469631', 'Test User', '2025-04-22 00:26:17.899461', '2025-04-22 00:26:17.899461');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (98, 'teacher_1745281578', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '9A8C99', 'Test Teacher', '2025-04-22 00:26:18.913182', '2025-04-22 00:26:18.913182');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (99, 'student_1745281578', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '8A2D2A', 'Test Student', '2025-04-22 00:26:19.92689', '2025-04-22 00:26:19.92689');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (73, 'testuser_1745280945', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '7F6203', 'Test User', '2025-04-22 00:15:45.984499', '2025-04-22 00:15:45.984499');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (74, 'teacher_1745280947', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '7129F8', 'Test Teacher', '2025-04-22 00:15:47.00502', '2025-04-22 00:15:47.00502');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (75, 'student_1745280947', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '951C2E', 'Test Student', '2025-04-22 00:15:48.018523', '2025-04-22 00:15:48.018523');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (77, 'teacher_1745280974', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '077BC8', 'Test Teacher', '2025-04-22 00:16:14.612506', '2025-04-22 00:16:14.612506');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (79, 'testuser_1745281092', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, 'DC104B', 'Test User', '2025-04-22 00:18:12.810485', '2025-04-22 00:18:12.810485');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (81, 'student_1745281093', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, 'BCA103', 'Test Student', '2025-04-22 00:18:14.863414', '2025-04-22 00:18:14.863414');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (83, 'teacher_1745281146', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, 'C824B5', 'Test Teacher', '2025-04-22 00:19:06.765978', '2025-04-22 00:19:06.765978');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (85, 'testuser_1745281258', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '4E7A4C', 'Test User', '2025-04-22 00:20:58.33461', '2025-04-22 00:20:58.33461');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (86, 'teacher_1745281259', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '7A9C92', 'Test Teacher', '2025-04-22 00:20:59.35437', '2025-04-22 00:20:59.35437');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (87, 'student_1745281259', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '38826B', 'Test Student', '2025-04-22 00:21:00.371785', '2025-04-22 00:21:00.371785');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (92, 'teacher_1745281362', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '04AA4C', 'Test Teacher', '2025-04-22 00:22:42.590414', '2025-04-22 00:22:42.590414');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (94, 'testuser_1745281385', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '219576', 'Test User', '2025-04-22 00:23:05.463139', '2025-04-22 00:23:05.463139');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (95, 'teacher_1745281386', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, 'EFB084', 'Test Teacher', '2025-04-22 00:23:06.48325', '2025-04-22 00:23:06.48325');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (96, 'student_1745281386', '0876e6c17ac41e68003af908be877b7a5178996eca7c91cac50fc825eef9ac5b018b9c862f7e3665317e60cc24d62a7cdd663ed7bfc22e782d565cc17190e101', 1, '9469D0', 'Test Student', '2025-04-22 00:23:07.493413', '2025-04-22 00:23:07.493413');
INSERT INTO public."user" (id, username, password, role, invite_code, name, created_at, updated_at) VALUES (71, 'test@test.ru', '80eb26791e4fd74e71cb685f219d2e8bcf96a462f70dcac12562a99eedf5ae62dd42523022b786a82f27b43a6250a4b7f1662e9fe4378b0e341d9482d395f669', 0, '389206', 'Иван Геннадьевич', '2025-04-22 00:10:36.334634', '2025-04-22 00:10:36.334634');


--
-- Name: event_id_seq; Type: SEQUENCE SET; Schema: public; Owner: rw_main
--

SELECT pg_catalog.setval('public.event_id_seq', 1, false);


--
-- Name: homework_files_id_seq; Type: SEQUENCE SET; Schema: public; Owner: rw_main
--

SELECT pg_catalog.setval('public.homework_files_id_seq', 11, true);


--
-- Name: homeworks_id_seq; Type: SEQUENCE SET; Schema: public; Owner: rw_main
--

SELECT pg_catalog.setval('public.homeworks_id_seq', 11, true);


--
-- Name: teacher_student_id_seq; Type: SEQUENCE SET; Schema: public; Owner: rw_main
--

SELECT pg_catalog.setval('public.teacher_student_id_seq', 17, true);


--
-- Name: user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: rw_main
--

SELECT pg_catalog.setval('public.user_id_seq', 99, true);


--
-- Name: event event_pkey; Type: CONSTRAINT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public.event
    ADD CONSTRAINT event_pkey PRIMARY KEY (id);


--
-- Name: homework_files homework_files_pkey; Type: CONSTRAINT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public.homework_files
    ADD CONSTRAINT homework_files_pkey PRIMARY KEY (id);


--
-- Name: homeworks homeworks_pkey; Type: CONSTRAINT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public.homeworks
    ADD CONSTRAINT homeworks_pkey PRIMARY KEY (id);


--
-- Name: teacher_student teacher_student_pkey; Type: CONSTRAINT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public.teacher_student
    ADD CONSTRAINT teacher_student_pkey PRIMARY KEY (id);


--
-- Name: teacher_student teacher_student_teacher_id_student_id_key; Type: CONSTRAINT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public.teacher_student
    ADD CONSTRAINT teacher_student_teacher_id_student_id_key UNIQUE (teacher_id, student_id);


--
-- Name: user user_pkey; Type: CONSTRAINT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public."user"
    ADD CONSTRAINT user_pkey PRIMARY KEY (id);


--
-- Name: user user_username_key; Type: CONSTRAINT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public."user"
    ADD CONSTRAINT user_username_key UNIQUE (username);


--
-- Name: idx_event_start_time; Type: INDEX; Schema: public; Owner: rw_main
--

CREATE INDEX idx_event_start_time ON public.event USING btree (start_time);


--
-- Name: idx_event_student_id; Type: INDEX; Schema: public; Owner: rw_main
--

CREATE INDEX idx_event_student_id ON public.event USING btree (student_id);


--
-- Name: idx_event_teacher_id; Type: INDEX; Schema: public; Owner: rw_main
--

CREATE INDEX idx_event_teacher_id ON public.event USING btree (teacher_id);


--
-- Name: idx_homework_files_homework_id; Type: INDEX; Schema: public; Owner: rw_main
--

CREATE INDEX idx_homework_files_homework_id ON public.homework_files USING btree (homework_id);


--
-- Name: idx_homeworks_student_id; Type: INDEX; Schema: public; Owner: rw_main
--

CREATE INDEX idx_homeworks_student_id ON public.homeworks USING btree (student_id);


--
-- Name: idx_homeworks_teacher_id; Type: INDEX; Schema: public; Owner: rw_main
--

CREATE INDEX idx_homeworks_teacher_id ON public.homeworks USING btree (teacher_id);


--
-- Name: idx_teacher_student_student_id; Type: INDEX; Schema: public; Owner: rw_main
--

CREATE INDEX idx_teacher_student_student_id ON public.teacher_student USING btree (student_id);


--
-- Name: idx_teacher_student_teacher_id; Type: INDEX; Schema: public; Owner: rw_main
--

CREATE INDEX idx_teacher_student_teacher_id ON public.teacher_student USING btree (teacher_id);


--
-- Name: idx_user_invite_code; Type: INDEX; Schema: public; Owner: rw_main
--

CREATE INDEX idx_user_invite_code ON public."user" USING btree (invite_code);


--
-- Name: idx_user_username; Type: INDEX; Schema: public; Owner: rw_main
--

CREATE INDEX idx_user_username ON public."user" USING btree (username);


--
-- Name: event update_event_updated_at; Type: TRIGGER; Schema: public; Owner: rw_main
--

CREATE TRIGGER update_event_updated_at BEFORE UPDATE ON public.event FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: homeworks update_homeworks_updated_at; Type: TRIGGER; Schema: public; Owner: rw_main
--

CREATE TRIGGER update_homeworks_updated_at BEFORE UPDATE ON public.homeworks FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: user update_user_updated_at; Type: TRIGGER; Schema: public; Owner: rw_main
--

CREATE TRIGGER update_user_updated_at BEFORE UPDATE ON public."user" FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: event event_student_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public.event
    ADD CONSTRAINT event_student_id_fkey FOREIGN KEY (student_id) REFERENCES public."user"(id) ON DELETE CASCADE;


--
-- Name: event event_teacher_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public.event
    ADD CONSTRAINT event_teacher_id_fkey FOREIGN KEY (teacher_id) REFERENCES public."user"(id) ON DELETE CASCADE;


--
-- Name: homework_files homework_files_homework_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public.homework_files
    ADD CONSTRAINT homework_files_homework_id_fkey FOREIGN KEY (homework_id) REFERENCES public.homeworks(id) ON DELETE CASCADE;


--
-- Name: homeworks homeworks_student_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public.homeworks
    ADD CONSTRAINT homeworks_student_id_fkey FOREIGN KEY (student_id) REFERENCES public."user"(id) ON DELETE CASCADE;


--
-- Name: homeworks homeworks_teacher_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public.homeworks
    ADD CONSTRAINT homeworks_teacher_id_fkey FOREIGN KEY (teacher_id) REFERENCES public."user"(id) ON DELETE CASCADE;


--
-- Name: teacher_student teacher_student_student_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public.teacher_student
    ADD CONSTRAINT teacher_student_student_id_fkey FOREIGN KEY (student_id) REFERENCES public."user"(id) ON DELETE CASCADE;


--
-- Name: teacher_student teacher_student_teacher_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: rw_main
--

ALTER TABLE ONLY public.teacher_student
    ADD CONSTRAINT teacher_student_teacher_id_fkey FOREIGN KEY (teacher_id) REFERENCES public."user"(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

