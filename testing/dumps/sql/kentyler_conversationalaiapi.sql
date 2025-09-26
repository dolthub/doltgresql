-- Downloaded from: https://github.com/kentyler/conversationalaiapi/blob/78e742330b401288aa971ac1bcbab487a3c8305e/neon_schema.sql
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
-- Name: vector; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA public;


--
-- Name: EXTENSION vector; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION vector IS 'vector data type and ivfflat and hnsw access methods';


--
-- Name: create_missing_sequences(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.create_missing_sequences() RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    table_rec RECORD;
    column_rec RECORD;
    seq_name TEXT;
    max_val BIGINT;
BEGIN
    -- Loop through all tables in the public schema
    FOR table_rec IN 
        SELECT c.relname AS table_name
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE c.relkind = 'r'  -- only tables
        AND n.nspname = 'public'
    LOOP
        -- For each table, find the primary key column of type BIGINT
        FOR column_rec IN
            SELECT a.attname AS column_name
            FROM pg_attribute a
            JOIN pg_constraint c ON a.attrelid = c.conrelid AND a.attnum = ANY(c.conkey)
            WHERE c.contype = 'p'  -- primary key
            AND a.attrelid = (SELECT oid FROM pg_class WHERE relname = table_rec.table_name)
            AND a.atttypid = 20  -- bigint type
        LOOP
            -- Create sequence name based on table and column
            seq_name := table_rec.table_name || '_' || column_rec.column_name || '_seq';
            
            -- Check if sequence already exists
            IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'S' AND relname = seq_name) THEN
                -- Get max value from the table for the sequence start
                EXECUTE 'SELECT COALESCE(MAX(' || column_rec.column_name || '), 0) FROM ' || table_rec.table_name INTO max_val;
                
                -- Create the sequence
                EXECUTE 'CREATE SEQUENCE IF NOT EXISTS public.' || seq_name ||
                        ' START WITH ' || (max_val + 1) ||
                        ' INCREMENT BY 1' ||
                        ' NO MINVALUE' ||
                        ' NO MAXVALUE' ||
                        ' CACHE 1';
                
                -- Set the default value for the column
                EXECUTE 'ALTER TABLE public.' || table_rec.table_name || 
                        ' ALTER COLUMN ' || column_rec.column_name || 
                        ' SET DEFAULT nextval(''public.' || seq_name || ''')';
                
                -- Set the ownership of sequence to the column
                EXECUTE 'ALTER SEQUENCE public.' || seq_name || 
                        ' OWNED BY public.' || table_rec.table_name || '.' || column_rec.column_name;
                
                RAISE NOTICE 'Created sequence % for table %.%', 
                      seq_name, table_rec.table_name, column_rec.column_name;
            ELSE
                RAISE NOTICE 'Sequence % already exists', seq_name;
            END IF;
        END LOOP;
    END LOOP;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: avatar_event_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.avatar_event_types (
    id integer NOT NULL,
    name text NOT NULL,
    description text NOT NULL
);


--
-- Name: avatar_event_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.avatar_event_types_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: avatar_event_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.avatar_event_types_id_seq OWNED BY public.avatar_event_types.id;


--
-- Name: avatar_scopes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.avatar_scopes (
    id integer NOT NULL,
    name text NOT NULL,
    description text
);


--
-- Name: avatar_scopes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.avatar_scopes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: avatar_scopes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.avatar_scopes_id_seq OWNED BY public.avatar_scopes.id;


--
-- Name: avatars; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.avatars (
    id bigint NOT NULL,
    name text NOT NULL,
    instruction_set text,
    created_at timestamp with time zone DEFAULT now(),
    avatar_scope_id integer DEFAULT 1 NOT NULL
);


--
-- Name: avatars_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.avatars_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: avatars_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.avatars_id_seq OWNED BY public.avatars.id;


--
-- Name: file_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.file_types (
    id bigint NOT NULL,
    name text NOT NULL,
    description text
);


--
-- Name: file_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.file_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: file_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.file_types_id_seq OWNED BY public.file_types.id;


--
-- Name: group_conversation_avatar_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.group_conversation_avatar_events (
    id integer NOT NULL,
    group_conversation_id bigint NOT NULL,
    avatar_id bigint NOT NULL,
    avatar_event_type_id integer NOT NULL,
    performed_by_avatar_id bigint NOT NULL,
    details jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: group_conversation_avatar_events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.group_conversation_avatar_events_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: group_conversation_avatar_events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.group_conversation_avatar_events_id_seq OWNED BY public.group_conversation_avatar_events.id;


--
-- Name: group_conversation_avatar_turn_relationships; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.group_conversation_avatar_turn_relationships (
    id bigint NOT NULL,
    turn_id bigint NOT NULL,
    target_turn_id bigint NOT NULL,
    relationship_type text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    turn_relationship_type_id integer DEFAULT 1 NOT NULL
);


--
-- Name: group_conversation_avatar_turn_relationships_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.group_conversation_avatar_turn_relationships_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: group_conversation_avatar_turn_relationships_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.group_conversation_avatar_turn_relationships_id_seq OWNED BY public.group_conversation_avatar_turn_relationships.id;


--
-- Name: group_conversation_avatar_turns; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.group_conversation_avatar_turns (
    id bigint NOT NULL,
    group_conversation_id bigint NOT NULL,
    avatar_id bigint NOT NULL,
    turn_index integer NOT NULL,
    content_text text NOT NULL,
    content_vector public.vector(1536),
    created_at timestamp with time zone DEFAULT now(),
    turn_kind_id integer NOT NULL
);


--
-- Name: group_conversation_avatar_turns_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.group_conversation_avatar_turns_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: group_conversation_avatar_turns_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.group_conversation_avatar_turns_id_seq OWNED BY public.group_conversation_avatar_turns.id;


--
-- Name: group_conversation_avatars; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.group_conversation_avatars (
    group_conversation_id bigint NOT NULL,
    avatar_id bigint NOT NULL,
    added_at timestamp with time zone DEFAULT now()
);


--
-- Name: group_conversation_avatars_avatar_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.group_conversation_avatars_avatar_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: group_conversation_avatars_avatar_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.group_conversation_avatars_avatar_id_seq OWNED BY public.group_conversation_avatars.avatar_id;


--
-- Name: group_conversation_avatars_group_conversation_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.group_conversation_avatars_group_conversation_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: group_conversation_avatars_group_conversation_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.group_conversation_avatars_group_conversation_id_seq OWNED BY public.group_conversation_avatars.group_conversation_id;


--
-- Name: group_conversation_upload_vectors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.group_conversation_upload_vectors (
    id bigint NOT NULL,
    upload_id bigint NOT NULL,
    chunk_index integer NOT NULL,
    content_text text NOT NULL,
    content_vector public.vector(1536),
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: group_conversation_upload_vectors_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.group_conversation_upload_vectors_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: group_conversation_upload_vectors_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.group_conversation_upload_vectors_id_seq OWNED BY public.group_conversation_upload_vectors.id;


--
-- Name: group_conversation_uploads; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.group_conversation_uploads (
    id bigint NOT NULL,
    group_conversation_id bigint NOT NULL,
    turn_id bigint NOT NULL,
    filename text NOT NULL,
    mime_type text,
    file_path text NOT NULL,
    uploaded_at timestamp with time zone DEFAULT now()
);


--
-- Name: group_conversation_uploads_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.group_conversation_uploads_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: group_conversation_uploads_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.group_conversation_uploads_id_seq OWNED BY public.group_conversation_uploads.id;


--
-- Name: group_conversations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.group_conversations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: group_conversations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.group_conversations (
    id bigint DEFAULT nextval('public.group_conversations_id_seq'::regclass) NOT NULL,
    group_id bigint,
    name text NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: groups_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.groups_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: groups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.groups (
    id bigint DEFAULT nextval('public.groups_id_seq'::regclass) NOT NULL,
    name text NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: participant_groups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.participant_groups (
    participant_id bigint NOT NULL,
    group_id bigint NOT NULL,
    role text,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: participant_groups_group_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.participant_groups_group_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: participant_groups_group_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.participant_groups_group_id_seq OWNED BY public.participant_groups.group_id;


--
-- Name: participant_groups_participant_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.participant_groups_participant_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: participant_groups_participant_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.participant_groups_participant_id_seq OWNED BY public.participant_groups.participant_id;


--
-- Name: participants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.participants (
    id bigint NOT NULL,
    name text NOT NULL,
    email text NOT NULL,
    password text NOT NULL,
    current_avatar_id bigint,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: participants_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.participants_id_seq
    START WITH 2
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: participants_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.participants_id_seq OWNED BY public.participants.id;


--
-- Name: turn_relationship_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.turn_relationship_types (
    id bigint NOT NULL,
    name text NOT NULL,
    description text
);


--
-- Name: relationship_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.relationship_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: relationship_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.relationship_types_id_seq OWNED BY public.turn_relationship_types.id;


--
-- Name: turn_kinds; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.turn_kinds (
    id integer NOT NULL,
    name text NOT NULL,
    description text
);


--
-- Name: turn_kinds_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.turn_kinds_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: turn_kinds_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.turn_kinds_id_seq OWNED BY public.turn_kinds.id;


--
-- Name: avatar_event_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.avatar_event_types ALTER COLUMN id SET DEFAULT nextval('public.avatar_event_types_id_seq'::regclass);


--
-- Name: avatar_scopes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.avatar_scopes ALTER COLUMN id SET DEFAULT nextval('public.avatar_scopes_id_seq'::regclass);


--
-- Name: avatars id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.avatars ALTER COLUMN id SET DEFAULT nextval('public.avatars_id_seq'::regclass);


--
-- Name: file_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.file_types ALTER COLUMN id SET DEFAULT nextval('public.file_types_id_seq'::regclass);


--
-- Name: group_conversation_avatar_events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_avatar_events ALTER COLUMN id SET DEFAULT nextval('public.group_conversation_avatar_events_id_seq'::regclass);


--
-- Name: group_conversation_avatar_turn_relationships id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_avatar_turn_relationships ALTER COLUMN id SET DEFAULT nextval('public.group_conversation_avatar_turn_relationships_id_seq'::regclass);


--
-- Name: group_conversation_avatar_turns id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_avatar_turns ALTER COLUMN id SET DEFAULT nextval('public.group_conversation_avatar_turns_id_seq'::regclass);


--
-- Name: group_conversation_avatars group_conversation_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_avatars ALTER COLUMN group_conversation_id SET DEFAULT nextval('public.group_conversation_avatars_group_conversation_id_seq'::regclass);


--
-- Name: group_conversation_avatars avatar_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_avatars ALTER COLUMN avatar_id SET DEFAULT nextval('public.group_conversation_avatars_avatar_id_seq'::regclass);


--
-- Name: group_conversation_upload_vectors id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_upload_vectors ALTER COLUMN id SET DEFAULT nextval('public.group_conversation_upload_vectors_id_seq'::regclass);


--
-- Name: group_conversation_uploads id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_uploads ALTER COLUMN id SET DEFAULT nextval('public.group_conversation_uploads_id_seq'::regclass);


--
-- Name: participant_groups participant_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.participant_groups ALTER COLUMN participant_id SET DEFAULT nextval('public.participant_groups_participant_id_seq'::regclass);


--
-- Name: participant_groups group_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.participant_groups ALTER COLUMN group_id SET DEFAULT nextval('public.participant_groups_group_id_seq'::regclass);


--
-- Name: participants id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.participants ALTER COLUMN id SET DEFAULT nextval('public.participants_id_seq'::regclass);


--
-- Name: turn_kinds id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.turn_kinds ALTER COLUMN id SET DEFAULT nextval('public.turn_kinds_id_seq'::regclass);


--
-- Name: turn_relationship_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.turn_relationship_types ALTER COLUMN id SET DEFAULT nextval('public.relationship_types_id_seq'::regclass);


--
-- Name: avatar_event_types avatar_event_types_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.avatar_event_types
    ADD CONSTRAINT avatar_event_types_name_key UNIQUE (name);


--
-- Name: avatar_event_types avatar_event_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.avatar_event_types
    ADD CONSTRAINT avatar_event_types_pkey PRIMARY KEY (id);


--
-- Name: avatar_scopes avatar_scopes_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.avatar_scopes
    ADD CONSTRAINT avatar_scopes_name_key UNIQUE (name);


--
-- Name: avatar_scopes avatar_scopes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.avatar_scopes
    ADD CONSTRAINT avatar_scopes_pkey PRIMARY KEY (id);


--
-- Name: avatars avatars_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.avatars
    ADD CONSTRAINT avatars_name_key UNIQUE (name);


--
-- Name: avatars avatars_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.avatars
    ADD CONSTRAINT avatars_pkey PRIMARY KEY (id);


--
-- Name: file_types file_types_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.file_types
    ADD CONSTRAINT file_types_name_key UNIQUE (name);


--
-- Name: file_types file_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.file_types
    ADD CONSTRAINT file_types_pkey PRIMARY KEY (id);


--
-- Name: group_conversation_avatar_turn_relationships group_conversation_avatar_tur_turn_id_target_turn_id_relati_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_avatar_turn_relationships
    ADD CONSTRAINT group_conversation_avatar_tur_turn_id_target_turn_id_relati_key UNIQUE (turn_id, target_turn_id, relationship_type);


--
-- Name: group_conversation_avatar_turn_relationships group_conversation_avatar_turn_relationships_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_avatar_turn_relationships
    ADD CONSTRAINT group_conversation_avatar_turn_relationships_pkey PRIMARY KEY (id);


--
-- Name: group_conversation_avatar_turns group_conversation_avatar_turns_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_avatar_turns
    ADD CONSTRAINT group_conversation_avatar_turns_pkey PRIMARY KEY (id);


--
-- Name: group_conversation_avatars group_conversation_avatars_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_avatars
    ADD CONSTRAINT group_conversation_avatars_pkey PRIMARY KEY (group_conversation_id, avatar_id);


--
-- Name: group_conversation_upload_vectors group_conversation_upload_vectors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_upload_vectors
    ADD CONSTRAINT group_conversation_upload_vectors_pkey PRIMARY KEY (id);


--
-- Name: group_conversation_uploads group_conversation_uploads_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_uploads
    ADD CONSTRAINT group_conversation_uploads_pkey PRIMARY KEY (id);


--
-- Name: group_conversations group_conversations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversations
    ADD CONSTRAINT group_conversations_pkey PRIMARY KEY (id);


--
-- Name: groups groups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.groups
    ADD CONSTRAINT groups_pkey PRIMARY KEY (id);


--
-- Name: participant_groups participant_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.participant_groups
    ADD CONSTRAINT participant_groups_pkey PRIMARY KEY (participant_id, group_id);


--
-- Name: participants participants_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.participants
    ADD CONSTRAINT participants_email_key UNIQUE (email);


--
-- Name: participants participants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.participants
    ADD CONSTRAINT participants_pkey PRIMARY KEY (id);


--
-- Name: turn_relationship_types relationship_types_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.turn_relationship_types
    ADD CONSTRAINT relationship_types_name_key UNIQUE (name);


--
-- Name: turn_relationship_types relationship_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.turn_relationship_types
    ADD CONSTRAINT relationship_types_pkey PRIMARY KEY (id);


--
-- Name: group_conversation_avatar_events session_avatar_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_avatar_events
    ADD CONSTRAINT session_avatar_events_pkey PRIMARY KEY (id);


--
-- Name: turn_kinds turn_kinds_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.turn_kinds
    ADD CONSTRAINT turn_kinds_name_key UNIQUE (name);


--
-- Name: turn_kinds turn_kinds_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.turn_kinds
    ADD CONSTRAINT turn_kinds_pkey PRIMARY KEY (id);


--
-- Name: idx_gcavtr_target; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_gcavtr_target ON public.group_conversation_avatar_turn_relationships USING btree (target_turn_id);


--
-- Name: avatars fk_avatars_scope; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.avatars
    ADD CONSTRAINT fk_avatars_scope FOREIGN KEY (avatar_scope_id) REFERENCES public.avatar_scopes(id) ON DELETE RESTRICT;


--
-- Name: participants fk_participants_current_avatar; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.participants
    ADD CONSTRAINT fk_participants_current_avatar FOREIGN KEY (current_avatar_id) REFERENCES public.avatars(id);


--
-- Name: group_conversation_avatar_events group_conversation_avatar_events_avatar_event_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_avatar_events
    ADD CONSTRAINT group_conversation_avatar_events_avatar_event_type_id_fkey FOREIGN KEY (avatar_event_type_id) REFERENCES public.avatar_event_types(id) ON DELETE RESTRICT;


--
-- Name: group_conversation_avatar_events group_conversation_avatar_events_avatar_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_avatar_events
    ADD CONSTRAINT group_conversation_avatar_events_avatar_id_fkey FOREIGN KEY (avatar_id) REFERENCES public.avatars(id) ON DELETE CASCADE;


--
-- Name: group_conversation_avatar_events group_conversation_avatar_events_group_conversation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_avatar_events
    ADD CONSTRAINT group_conversation_avatar_events_group_conversation_id_fkey FOREIGN KEY (group_conversation_id) REFERENCES public.group_conversations(id) ON DELETE CASCADE;


--
-- Name: group_conversation_avatar_events group_conversation_avatar_events_performed_by_avatar_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_avatar_events
    ADD CONSTRAINT group_conversation_avatar_events_performed_by_avatar_id_fkey FOREIGN KEY (performed_by_avatar_id) REFERENCES public.avatars(id) ON DELETE SET NULL;


--
-- Name: group_conversation_avatar_turn_relationships group_conversation_avatar_turn_relationship_target_turn_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_avatar_turn_relationships
    ADD CONSTRAINT group_conversation_avatar_turn_relationship_target_turn_id_fkey FOREIGN KEY (target_turn_id) REFERENCES public.group_conversation_avatar_turns(id) ON DELETE CASCADE;


--
-- Name: group_conversation_avatar_turn_relationships group_conversation_avatar_turn_relationships_turn_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_avatar_turn_relationships
    ADD CONSTRAINT group_conversation_avatar_turn_relationships_turn_id_fkey FOREIGN KEY (turn_id) REFERENCES public.group_conversation_avatar_turns(id) ON DELETE CASCADE;


--
-- Name: group_conversation_avatar_turns group_conversation_avatar_turns_turn_kind_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_conversation_avatar_turns
    ADD CONSTRAINT group_conversation_avatar_turns_turn_kind_id_fkey FOREIGN KEY (turn_kind_id) REFERENCES public.turn_kinds(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- PostgreSQL database dump complete
--

