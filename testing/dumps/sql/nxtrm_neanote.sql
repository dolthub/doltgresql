-- Downloaded from: https://github.com/nxtrm/neanote/blob/5df6e89644d61dc346590d4cc87015944c628249/dbexport.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 16.4
-- Dumped by pg_dump version 16.4

-- Started on 2025-04-18 09:56:13

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
-- TOC entry 2 (class 3079 OID 16403)
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- TOC entry 4930 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- TOC entry 867 (class 1247 OID 16426)
-- Name: note_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.note_type AS ENUM (
    'memory',
    'task',
    'goal',
    'event',
    'habit',
    'note'
);


ALTER TYPE public.note_type OWNER TO postgres;

--
-- TOC entry 242 (class 1255 OID 32768)
-- Name: cosine_similarity(double precision[], double precision[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cosine_similarity(vec1 double precision[], vec2 double precision[]) RETURNS double precision
    LANGUAGE plpgsql
    AS $$
DECLARE
    dot_product double precision := 0;
    norm_a double precision := 0;
    norm_b double precision := 0;
    i int;
BEGIN
    IF vec1 IS NULL OR vec2 IS NULL THEN
        RETURN 0;
    END IF;

    FOR i IN array_lower(vec1, 1)..array_upper(vec1, 1) LOOP
        dot_product := dot_product + (vec1[i] * vec2[i]);
        norm_a := norm_a + (vec1[i] * vec1[i]);
        norm_b := norm_b + (vec2[i] * vec2[i]);
    END LOOP;
    IF norm_a = 0 OR norm_b = 0 THEN
        RETURN 0;
    END IF;
    RETURN dot_product / (sqrt(norm_a) * sqrt(norm_b));
END;
$$;


ALTER FUNCTION public.cosine_similarity(vec1 double precision[], vec2 double precision[]) OWNER TO postgres;

--
-- TOC entry 241 (class 1255 OID 24576)
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_updated_at_column() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 230 (class 1259 OID 73732)
-- Name: goalhistory; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.goalhistory (
    goal_id uuid,
    id uuid NOT NULL
);


ALTER TABLE public.goalhistory OWNER TO postgres;

--
-- TOC entry 221 (class 1259 OID 16501)
-- Name: goals; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.goals (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    note_id uuid,
    completion_timestamp timestamp without time zone,
    due_date timestamp without time zone
);


ALTER TABLE public.goals OWNER TO postgres;

--
-- TOC entry 224 (class 1259 OID 16575)
-- Name: habitcompletion; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.habitcompletion (
    habit_id uuid,
    completion_date date
);


ALTER TABLE public.habitcompletion OWNER TO postgres;

--
-- TOC entry 223 (class 1259 OID 16563)
-- Name: habits; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.habits (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    note_id uuid,
    reminder_time time without time zone,
    streak integer DEFAULT 0,
    repetition character varying(7)
);


ALTER TABLE public.habits OWNER TO postgres;

--
-- TOC entry 225 (class 1259 OID 16583)
-- Name: habittasks; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.habittasks (
    habit_id uuid,
    task_id uuid
);


ALTER TABLE public.habittasks OWNER TO postgres;

--
-- TOC entry 222 (class 1259 OID 16512)
-- Name: milestones; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.milestones (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    goal_id uuid,
    description text,
    ms_index integer,
    completed boolean DEFAULT false
);


ALTER TABLE public.milestones OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 16448)
-- Name: notes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.notes (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    title character varying(100),
    content text,
    type public.note_type NOT NULL,
    archived boolean DEFAULT false,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    vector double precision[]
);


ALTER TABLE public.notes OWNER TO postgres;

--
-- TOC entry 227 (class 1259 OID 16604)
-- Name: notetags; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.notetags (
    note_id uuid NOT NULL,
    tag_id uuid NOT NULL
);


ALTER TABLE public.notetags OWNER TO postgres;

--
-- TOC entry 219 (class 1259 OID 16476)
-- Name: subtasks; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.subtasks (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    task_id uuid,
    description text,
    completed boolean DEFAULT false,
    st_index integer
);


ALTER TABLE public.subtasks OWNER TO postgres;

--
-- TOC entry 226 (class 1259 OID 16596)
-- Name: tags; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tags (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(50) NOT NULL,
    color character varying(7) NOT NULL,
    user_id uuid
);


ALTER TABLE public.tags OWNER TO postgres;

--
-- TOC entry 218 (class 1259 OID 16464)
-- Name: tasks; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tasks (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    note_id uuid,
    completed boolean DEFAULT false,
    due_date timestamp without time zone,
    completion_timestamp timestamp without time zone
);


ALTER TABLE public.tasks OWNER TO postgres;

--
-- TOC entry 220 (class 1259 OID 16490)
-- Name: taskstatistics; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.taskstatistics (
    user_id uuid,
    total_completed_tasks integer DEFAULT 0,
    weekly_completed_tasks integer DEFAULT 0,
    monthly_completed_tasks integer DEFAULT 0
);


ALTER TABLE public.taskstatistics OWNER TO postgres;

--
-- TOC entry 229 (class 1259 OID 57348)
-- Name: user_widgets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_widgets (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    data_source_id uuid,
    configuration json,
    data_source_type character varying,
    widget_id character varying,
    title character varying
);


ALTER TABLE public.user_widgets OWNER TO postgres;

--
-- TOC entry 216 (class 1259 OID 16437)
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    username character varying(50) NOT NULL,
    email character varying(100) NOT NULL,
    password character varying(255) NOT NULL,
    preferences json DEFAULT '{}'::json
);


ALTER TABLE public.users OWNER TO postgres;

--
-- TOC entry 228 (class 1259 OID 49156)
-- Name: widgets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.widgets (
    id uuid NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    allowed_data_sources public.note_type[]
);


ALTER TABLE public.widgets OWNER TO postgres;

--
-- TOC entry 4924 (class 0 OID 73732)
-- Dependencies: 230
-- Data for Name: goalhistory; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.goalhistory (goal_id, id) FROM stdin;
\.


--
-- TOC entry 4915 (class 0 OID 16501)
-- Dependencies: 221
-- Data for Name: goals; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.goals (id, note_id, completion_timestamp, due_date) FROM stdin;
\.


--
-- TOC entry 4918 (class 0 OID 16575)
-- Dependencies: 224
-- Data for Name: habitcompletion; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.habitcompletion (habit_id, completion_date) FROM stdin;
\.


--
-- TOC entry 4917 (class 0 OID 16563)
-- Dependencies: 223
-- Data for Name: habits; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.habits (id, note_id, reminder_time, streak, repetition) FROM stdin;

\.


--
-- TOC entry 4919 (class 0 OID 16583)
-- Dependencies: 225
-- Data for Name: habittasks; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.habittasks (habit_id, task_id) FROM stdin;
\.


--
-- TOC entry 4916 (class 0 OID 16512)
-- Dependencies: 222
-- Data for Name: milestones; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.milestones (id, goal_id, description, ms_index, completed) FROM stdin;
\.


--
-- TOC entry 4911 (class 0 OID 16448)
-- Dependencies: 217
-- Data for Name: notes; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.notes (id, user_id, title, content, type, archived, created_at, updated_at, vector) FROM stdin;
\.


--
-- TOC entry 4921 (class 0 OID 16604)
-- Dependencies: 227
-- Data for Name: notetags; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.notetags (note_id, tag_id) FROM stdin;
\.


--
-- TOC entry 4913 (class 0 OID 16476)
-- Dependencies: 219
-- Data for Name: subtasks; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.subtasks (id, task_id, description, completed, st_index) FROM stdin;
\.


--
-- TOC entry 4920 (class 0 OID 16596)
-- Dependencies: 226
-- Data for Name: tags; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tags (id, name, color, user_id) FROM stdin;
\.


--
-- TOC entry 4912 (class 0 OID 16464)
-- Dependencies: 218
-- Data for Name: tasks; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tasks (id, note_id, completed, due_date, completion_timestamp) FROM stdin;
\.


--
-- TOC entry 4914 (class 0 OID 16490)
-- Dependencies: 220
-- Data for Name: taskstatistics; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.taskstatistics (user_id, total_completed_tasks, weekly_completed_tasks, monthly_completed_tasks) FROM stdin;
\.


--
-- TOC entry 4923 (class 0 OID 57348)
-- Dependencies: 229
-- Data for Name: user_widgets; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_widgets (id, user_id, data_source_id, configuration, data_source_type, widget_id, title) FROM stdin;
\.


--
-- TOC entry 4910 (class 0 OID 16437)
-- Dependencies: 216
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, username, email, password, preferences) FROM stdin;
\.


--
-- TOC entry 4922 (class 0 OID 49156)
-- Dependencies: 228
-- Data for Name: widgets; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.widgets (id, name, description, allowed_data_sources) FROM stdin;
\.


--
-- TOC entry 4752 (class 2606 OID 73736)
-- Name: goalhistory goalhistory_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.goalhistory
    ADD CONSTRAINT goalhistory_pkey PRIMARY KEY (id);


--
-- TOC entry 4736 (class 2606 OID 16506)
-- Name: goals goals_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.goals
    ADD CONSTRAINT goals_pkey PRIMARY KEY (id);


--
-- TOC entry 4740 (class 2606 OID 16569)
-- Name: habits habits_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.habits
    ADD CONSTRAINT habits_pkey PRIMARY KEY (id);


--
-- TOC entry 4738 (class 2606 OID 16520)
-- Name: milestones milestones_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.milestones
    ADD CONSTRAINT milestones_pkey PRIMARY KEY (id);


--
-- TOC entry 4730 (class 2606 OID 16458)
-- Name: notes notes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notes
    ADD CONSTRAINT notes_pkey PRIMARY KEY (id);


--
-- TOC entry 4746 (class 2606 OID 16608)
-- Name: notetags notetags_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notetags
    ADD CONSTRAINT notetags_pkey PRIMARY KEY (note_id, tag_id);


--
-- TOC entry 4734 (class 2606 OID 16484)
-- Name: subtasks subtasks_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subtasks
    ADD CONSTRAINT subtasks_pkey PRIMARY KEY (id);


--
-- TOC entry 4742 (class 2606 OID 16603)
-- Name: tags tags_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_name_key UNIQUE (name);


--
-- TOC entry 4744 (class 2606 OID 16601)
-- Name: tags tags_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (id);


--
-- TOC entry 4732 (class 2606 OID 16470)
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);


--
-- TOC entry 4750 (class 2606 OID 57354)
-- Name: user_widgets user_widgets_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_widgets
    ADD CONSTRAINT user_widgets_pkey PRIMARY KEY (id);


--
-- TOC entry 4726 (class 2606 OID 16447)
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- TOC entry 4728 (class 2606 OID 16445)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- TOC entry 4748 (class 2606 OID 49162)
-- Name: widgets widgets_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.widgets
    ADD CONSTRAINT widgets_pkey PRIMARY KEY (id);


--
-- TOC entry 4766 (class 2620 OID 24577)
-- Name: notes update_notes_updated_at; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_notes_updated_at BEFORE UPDATE ON public.notes FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- TOC entry 4757 (class 2606 OID 16507)
-- Name: goals goals_note_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.goals
    ADD CONSTRAINT goals_note_id_fkey FOREIGN KEY (note_id) REFERENCES public.notes(id);


--
-- TOC entry 4760 (class 2606 OID 16578)
-- Name: habitcompletion habitcompletion_habit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.habitcompletion
    ADD CONSTRAINT habitcompletion_habit_id_fkey FOREIGN KEY (habit_id) REFERENCES public.habits(id);


--
-- TOC entry 4759 (class 2606 OID 16570)
-- Name: habits habits_note_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.habits
    ADD CONSTRAINT habits_note_id_fkey FOREIGN KEY (note_id) REFERENCES public.notes(id);


--
-- TOC entry 4761 (class 2606 OID 16586)
-- Name: habittasks habittasks_habit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.habittasks
    ADD CONSTRAINT habittasks_habit_id_fkey FOREIGN KEY (habit_id) REFERENCES public.habits(id);


--
-- TOC entry 4762 (class 2606 OID 16591)
-- Name: habittasks habittasks_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.habittasks
    ADD CONSTRAINT habittasks_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id);


--
-- TOC entry 4758 (class 2606 OID 16521)
-- Name: milestones milestones_goal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.milestones
    ADD CONSTRAINT milestones_goal_id_fkey FOREIGN KEY (goal_id) REFERENCES public.goals(id);


--
-- TOC entry 4753 (class 2606 OID 16459)
-- Name: notes notes_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notes
    ADD CONSTRAINT notes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- TOC entry 4764 (class 2606 OID 16609)
-- Name: notetags notetags_note_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notetags
    ADD CONSTRAINT notetags_note_id_fkey FOREIGN KEY (note_id) REFERENCES public.notes(id);


--
-- TOC entry 4765 (class 2606 OID 16614)
-- Name: notetags notetags_tag_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notetags
    ADD CONSTRAINT notetags_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES public.tags(id);


--
-- TOC entry 4755 (class 2606 OID 16485)
-- Name: subtasks subtasks_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subtasks
    ADD CONSTRAINT subtasks_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id);


--
-- TOC entry 4763 (class 2606 OID 16620)
-- Name: tags tags_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- TOC entry 4754 (class 2606 OID 16471)
-- Name: tasks tasks_note_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_note_id_fkey FOREIGN KEY (note_id) REFERENCES public.notes(id);


--
-- TOC entry 4756 (class 2606 OID 16496)
-- Name: taskstatistics taskstatistics_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.taskstatistics
    ADD CONSTRAINT taskstatistics_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


-- Completed on 2025-04-18 09:56:13

--
-- PostgreSQL database dump complete
--

