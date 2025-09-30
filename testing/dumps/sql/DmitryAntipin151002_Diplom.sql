-- Downloaded from: https://github.com/DmitryAntipin151002/Diplom/blob/350c6c1b05e8ce85abe2eca8a77acacf5b14e4d4/dump_AuthS.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 17.4
-- Dumped by pg_dump version 17.4

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
-- Name: update_activity_stats(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_activity_stats() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    UPDATE user_profiles
    SET 
        total_activities = (SELECT COUNT(*) FROM user_activities WHERE user_id = NEW.user_id),
        total_distance = (SELECT COALESCE(SUM(distance), 0) FROM user_activities WHERE user_id = NEW.user_id),
        updated_at = CURRENT_TIMESTAMP
    WHERE user_id = NEW.user_id;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_activity_stats() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: event; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.event (
    id integer NOT NULL,
    name character varying(255),
    description text,
    event_date timestamp without time zone,
    max_participants integer,
    organizer_id uuid
);


ALTER TABLE public.event OWNER TO postgres;

--
-- Name: event_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.event_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.event_id_seq OWNER TO postgres;

--
-- Name: event_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.event_id_seq OWNED BY public.event.id;


--
-- Name: event_participant; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.event_participant (
    id integer NOT NULL,
    user_id uuid,
    joined_at timestamp without time zone,
    event_id integer
);


ALTER TABLE public.event_participant OWNER TO postgres;

--
-- Name: event_participant_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.event_participant_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.event_participant_id_seq OWNER TO postgres;

--
-- Name: event_participant_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.event_participant_id_seq OWNED BY public.event_participant.id;


--
-- Name: role; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.role (
    id integer NOT NULL,
    name character varying(50) NOT NULL
);


ALTER TABLE public.role OWNER TO postgres;

--
-- Name: role_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.role_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.role_id_seq OWNER TO postgres;

--
-- Name: role_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.role_id_seq OWNED BY public.role.id;


--
-- Name: status; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.status (
    id integer NOT NULL,
    name character varying(50) NOT NULL
);


ALTER TABLE public.status OWNER TO postgres;

--
-- Name: status_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.status_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.status_id_seq OWNER TO postgres;

--
-- Name: status_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.status_id_seq OWNED BY public.status.id;


--
-- Name: user_activities; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_activities (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    activity_type character varying(50) NOT NULL,
    distance numeric(10,2),
    duration text,
    calories_burned integer,
    activity_date timestamp without time zone NOT NULL,
    external_id character varying(100),
    raw_data character varying
);


ALTER TABLE public.user_activities OWNER TO postgres;

--
-- Name: user_photos; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_photos (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    photo_url character varying(255) NOT NULL,
    is_main boolean DEFAULT false,
    uploaded_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    description text
);


ALTER TABLE public.user_photos OWNER TO postgres;

--
-- Name: user_profiles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_profiles (
    user_id uuid NOT NULL,
    avatar_url character varying(255),
    bio text,
    date_of_birth date,
    gender character varying(10),
    location character varying(100),
    sport_type character varying(50),
    fitness_level character varying(20),
    goals text,
    achievements text,
    total_activities integer DEFAULT 0,
    total_distance numeric(10,2) DEFAULT 0,
    total_wins integer DEFAULT 0,
    personal_records text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    is_verified boolean,
    last_active_at date,
    CONSTRAINT user_profiles_fitness_level_check CHECK (((fitness_level)::text = ANY (ARRAY[('Beginner'::character varying)::text, ('Intermediate'::character varying)::text, ('Advanced'::character varying)::text]))),
    CONSTRAINT user_profiles_gender_check CHECK (((gender)::text = ANY (ARRAY[('Male'::character varying)::text, ('Female'::character varying)::text, ('Other'::character varying)::text])))
);


ALTER TABLE public.user_profiles OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying(100) NOT NULL,
    phone_number character varying(15),
    encrypted_password character(60) NOT NULL,
    status bigint,
    is_first_enter boolean DEFAULT false NOT NULL,
    end_date date,
    role_id bigint NOT NULL,
    last_login date
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: verification_code; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.verification_code (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    code character varying(6) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    expires_at timestamp without time zone NOT NULL
);


ALTER TABLE public.verification_code OWNER TO postgres;

--
-- Name: event id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event ALTER COLUMN id SET DEFAULT nextval('public.event_id_seq'::regclass);


--
-- Name: event_participant id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_participant ALTER COLUMN id SET DEFAULT nextval('public.event_participant_id_seq'::regclass);


--
-- Name: role id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role ALTER COLUMN id SET DEFAULT nextval('public.role_id_seq'::regclass);


--
-- Name: status id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.status ALTER COLUMN id SET DEFAULT nextval('public.status_id_seq'::regclass);


--
-- Data for Name: event; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.event (id, name, description, event_date, max_participants, organizer_id) FROM stdin;
2	Футбол на стадионе	Дружеский матч на 10 человек	2025-04-09 17:00:00	10	6b1cf72a-60b3-4f78-ac20-cfc2e75b8512
\.


--
-- Data for Name: event_participant; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.event_participant (id, user_id, joined_at, event_id) FROM stdin;
\.


--
-- Data for Name: role; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.role (id, name) FROM stdin;
1	ADMIN
2	USER
\.


--
-- Data for Name: status; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.status (id, name) FROM stdin;
1	ACTIVE
\.


--
-- Data for Name: user_activities; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_activities (id, user_id, activity_type, distance, duration, calories_burned, activity_date, external_id, raw_data) FROM stdin;
dfa160ae-d33f-4b01-8672-8168bd744ca3	960b0c25-4568-4a4a-b45a-1edf4a6dfcfc	Running	5.00	PT30M	500	2025-04-05 20:03:26.558959	external-id-123	{"speed": "12km/h"}
\.


--
-- Data for Name: user_photos; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_photos (id, user_id, photo_url, is_main, uploaded_at, description) FROM stdin;
d5023ad9-591d-4a9f-9872-bf3209bc7166	960b0c25-4568-4a4a-b45a-1edf4a6dfcfc	http://example.com/new-avatar.jpg	t	2025-04-05 19:50:16.791349	\N
\.


--
-- Data for Name: user_profiles; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_profiles (user_id, avatar_url, bio, date_of_birth, gender, location, sport_type, fitness_level, goals, achievements, total_activities, total_distance, total_wins, personal_records, created_at, updated_at, is_verified, last_active_at) FROM stdin;
960b0c25-4568-4a4a-b45a-1edf4a6dfcfc	http://example.com/avatar.jpg	Bio of the user	1990-01-01	Male	New Sity	Outdoor	Intermediate	Run 5k	Completed 10k	1	5.00	5	5k in 30min	2025-04-05 16:09:11.481505	2025-04-05 20:03:26.561609	t	2025-04-05
6b1cf72a-60b3-4f78-ac20-cfc2e75b8512	http://example.com/avatar.jpg	Bio of the user	1990-01-01	Male	New York	Outdoor	Intermediate	Run 5k	Completed 10k	10	100.50	5	5k in 30min	2025-04-08 13:00:54.036292	2025-04-08 13:00:54.036292	t	2025-04-05
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, email, phone_number, encrypted_password, status, is_first_enter, end_date, role_id, last_login) FROM stdin;
ce2082c5-1f56-4293-aac4-4a057c6980bc	user@example.com	+1234567890	$2a$10$.PC11OLxxrEH8IodYZGY2OX4T617k4qq6vbrlGN4MdQNtnJGcI5j6	1	t	\N	1	\N
960b0c25-4568-4a4a-b45a-1edf4a6dfcfc	user@example1.com	\N	$2a$10$xYVXvwWtPQD/gqgANOVxeuMkAHK83PTcv5tm3QTM51JpAXHXP/Ake	1	f	\N	1	2025-04-05
73c398a7-6612-4764-97ec-098c5802ae97	user@example2.com	\N	$2a$10$8AP//1k2Son9d/JUdUU6VuwMHGK9pzoCEAMhD5ryVN2MiMP56Pv7y	1	t	\N	1	2025-04-05
1a54beaa-6290-4dbf-b18b-a3df94cc63b5	user@example4.com	\N	$2a$10$7E6Pt9/01D8jKK4z0i98BesuxJVVSxLbIO1VWEVkiT1jc2yH.LLAG	1	t	\N	1	2025-04-05
6b1cf72a-60b3-4f78-ac20-cfc2e75b8512	dmitry@example.com	\N	$2a$10$hEZdn2vPjZkCJdMDuMs.hO31SyLqbIFgCeYfFcexZFVRBQIZeJzeq	1	t	\N	1	2025-04-08
\.


--
-- Data for Name: verification_code; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.verification_code (id, user_id, code, created_at, expires_at) FROM stdin;
3b9a4c0b-b319-4639-ba60-ad064bc4596b	ce2082c5-1f56-4293-aac4-4a057c6980bc	262495	2025-03-19 15:05:26.162304	2025-03-19 15:10:26.162304
39c31998-35f3-4e8e-aeeb-eb3c1a877a56	960b0c25-4568-4a4a-b45a-1edf4a6dfcfc	341250	2025-04-05 14:59:50.149167	2025-04-05 15:04:50.149167
3ecd4af4-f5e4-476b-83cf-cc4906d9efc4	6b1cf72a-60b3-4f78-ac20-cfc2e75b8512	226002	2025-04-08 13:00:08.830408	2025-04-08 13:05:08.830408
\.


--
-- Name: event_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.event_id_seq', 2, true);


--
-- Name: event_participant_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.event_participant_id_seq', 1, false);


--
-- Name: role_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.role_id_seq', 1, false);


--
-- Name: status_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.status_id_seq', 1, false);


--
-- Name: event_participant event_participant_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_participant
    ADD CONSTRAINT event_participant_pkey PRIMARY KEY (id);


--
-- Name: event event_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event
    ADD CONSTRAINT event_pkey PRIMARY KEY (id);


--
-- Name: role role_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role
    ADD CONSTRAINT role_name_key UNIQUE (name);


--
-- Name: role role_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role
    ADD CONSTRAINT role_pkey PRIMARY KEY (id);


--
-- Name: status status_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.status
    ADD CONSTRAINT status_name_key UNIQUE (name);


--
-- Name: status status_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.status
    ADD CONSTRAINT status_pkey PRIMARY KEY (id);


--
-- Name: user_activities user_activities_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_activities
    ADD CONSTRAINT user_activities_pkey PRIMARY KEY (id);


--
-- Name: user_photos user_photos_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_photos
    ADD CONSTRAINT user_photos_pkey PRIMARY KEY (id);


--
-- Name: user_profiles user_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT user_profiles_pkey PRIMARY KEY (user_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: verification_code verification_code_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.verification_code
    ADD CONSTRAINT verification_code_pkey PRIMARY KEY (id);


--
-- Name: idx_user_activities_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_activities_date ON public.user_activities USING btree (activity_date);


--
-- Name: idx_user_activities_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_activities_user_id ON public.user_activities USING btree (user_id);


--
-- Name: idx_user_photos_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_photos_user_id ON public.user_photos USING btree (user_id);


--
-- Name: user_activities trigger_update_activity_stats; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trigger_update_activity_stats AFTER INSERT OR DELETE OR UPDATE ON public.user_activities FOR EACH ROW EXECUTE FUNCTION public.update_activity_stats();


--
-- Name: event_participant event_participant_event_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_participant
    ADD CONSTRAINT event_participant_event_id_fkey FOREIGN KEY (event_id) REFERENCES public.event(id);


--
-- Name: user_activities user_activities_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_activities
    ADD CONSTRAINT user_activities_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.user_profiles(user_id) ON DELETE CASCADE;


--
-- Name: user_photos user_photos_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_photos
    ADD CONSTRAINT user_photos_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.user_profiles(user_id) ON DELETE CASCADE;


--
-- Name: user_profiles user_profiles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT user_profiles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: users users_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.role(id) ON DELETE CASCADE;


--
-- Name: users users_status_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_status_fkey FOREIGN KEY (status) REFERENCES public.status(id) ON DELETE SET NULL;


--
-- PostgreSQL database dump complete
--

