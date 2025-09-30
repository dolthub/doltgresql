-- Downloaded from: https://github.com/julesd7/collatask/blob/659e477af286721943425883fd618038bfa94ea9/structure.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 16.4
-- Dumped by pg_dump version 16.8 (Homebrew)

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
-- Name: assign_user_to_project(integer, integer, character varying); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.assign_user_to_project(p_user_id integer, p_project_id integer, p_role character varying DEFAULT 'member'::character varying) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    INSERT INTO project_assignments (user_id, project_id, role)
    VALUES (p_user_id, p_project_id, p_role)
    ON CONFLICT (user_id, project_id) DO NOTHING;  -- Ignore si l'utilisateur est déjà assigné
END;
$$;


--
-- Name: create_default_boards(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.create_default_boards() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    INSERT INTO boards (project_id, title) VALUES (NEW.id, 'To Do');
    INSERT INTO boards (project_id, title) VALUES (NEW.id, 'In Progress');
    INSERT INTO boards (project_id, title) VALUES (NEW.id, 'Completed');
    RETURN NEW;
END;
$$;


--
-- Name: set_deleted_user(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.set_deleted_user() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    UPDATE messages
    SET sender = 'DeletedUser'
    WHERE sender = OLD.username;

    RETURN OLD;
END;
$$;


--
-- Name: update_board_timestamp(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_board_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;  -- Met à jour le champ updated_at à l'heure actuelle
    RETURN NEW;
END;
$$;


--
-- Name: update_card_last_change(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_card_last_change() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.last_change = CURRENT_TIMESTAMP;  -- Met à jour le champ last_change à l'heure actuelle
    RETURN NEW;
END;
$$;


--
-- Name: update_project_timestamp(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_project_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    UPDATE projects
    SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.project_id;
    RETURN NEW;
END;
$$;


--
-- Name: update_task_timestamp(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_task_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;  -- Met à jour le champ updated_at à l'heure actuelle
    RETURN NEW;
END;
$$;


--
-- Name: update_timestamp(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;  -- Met à jour le champ updated_at à l'heure actuelle
    RETURN NEW;
END;
$$;


--
-- Name: update_updated_at(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: boards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.boards (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    title character varying(255),
    project_id uuid
);


--
-- Name: cards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cards (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    title character varying(100) NOT NULL,
    description text,
    start_date date,
    end_date date,
    last_change timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    board_id uuid,
    project_id uuid,
    assignees_ids uuid[],
    priority character varying(2),
    CONSTRAINT cards_priority_check CHECK (((priority)::text = ANY ((ARRAY['P0'::character varying, 'P1'::character varying, 'P2'::character varying, 'P3'::character varying])::text[])))
);


--
-- Name: cards_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.cards_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.messages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    message text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    room uuid NOT NULL,
    sender text NOT NULL
);


--
-- Name: project_assignments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.project_assignments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    role character varying(20) DEFAULT 'viewer'::character varying,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    user_id uuid NOT NULL,
    project_id uuid NOT NULL
);


--
-- Name: projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.projects (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    title character varying(100) NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    owner_id uuid
);


--
-- Name: tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tasks (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    title character varying(100) NOT NULL,
    description text,
    status character varying(20) DEFAULT 'pending'::character varying,
    board_id integer NOT NULL,
    assigned_to integer,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    username character varying(50) NOT NULL,
    email character varying(100) NOT NULL,
    password character varying(255),
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    verified boolean DEFAULT false,
    verification_token character varying(255),
    reset_token character varying(255),
    last_connection timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: boards boards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.boards
    ADD CONSTRAINT boards_pkey PRIMARY KEY (id);


--
-- Name: cards cards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cards
    ADD CONSTRAINT cards_pkey PRIMARY KEY (id);


--
-- Name: project_assignments project_assignments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_assignments
    ADD CONSTRAINT project_assignments_pkey PRIMARY KEY (id);


--
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: project_assignments set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER set_updated_at BEFORE UPDATE ON public.project_assignments FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: projects trigger_create_default_boards; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trigger_create_default_boards AFTER INSERT ON public.projects FOR EACH ROW EXECUTE FUNCTION public.create_default_boards();


--
-- Name: boards trigger_update_board_timestamp; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trigger_update_board_timestamp BEFORE UPDATE ON public.boards FOR EACH ROW EXECUTE FUNCTION public.update_board_timestamp();


--
-- Name: cards trigger_update_card_last_change; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trigger_update_card_last_change BEFORE UPDATE ON public.cards FOR EACH ROW EXECUTE FUNCTION public.update_card_last_change();


--
-- Name: boards trigger_update_project_on_board_change; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trigger_update_project_on_board_change AFTER INSERT OR DELETE OR UPDATE ON public.boards FOR EACH ROW EXECUTE FUNCTION public.update_project_timestamp();


--
-- Name: cards trigger_update_project_on_card_change; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trigger_update_project_on_card_change AFTER INSERT OR DELETE OR UPDATE ON public.cards FOR EACH ROW EXECUTE FUNCTION public.update_project_timestamp();


--
-- Name: tasks trigger_update_task_timestamp; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trigger_update_task_timestamp BEFORE UPDATE ON public.tasks FOR EACH ROW EXECUTE FUNCTION public.update_task_timestamp();


--
-- Name: users user_deletion_trigger; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER user_deletion_trigger BEFORE DELETE ON public.users FOR EACH ROW EXECUTE FUNCTION public.set_deleted_user();


--
-- Name: cards fk_board_id; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cards
    ADD CONSTRAINT fk_board_id FOREIGN KEY (board_id) REFERENCES public.boards(id) ON DELETE CASCADE;


--
-- Name: projects fk_owner_id; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_owner_id FOREIGN KEY (owner_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: project_assignments fk_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_assignments
    ADD CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: project_assignments fk_project_id; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_assignments
    ADD CONSTRAINT fk_project_id FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: boards fk_project_id; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.boards
    ADD CONSTRAINT fk_project_id FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: cards fk_project_id; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cards
    ADD CONSTRAINT fk_project_id FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: project_assignments fk_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_assignments
    ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

