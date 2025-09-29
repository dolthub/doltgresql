-- Downloaded from: https://github.com/CravingCrates/AnkiCollab-Backend/blob/ff713df0867723b05a2e0830112b800515b5eb65/schema-database.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 17.4 (Ubuntu 17.4-1.pgdg24.04+2)
-- Dumped by pg_dump version 17.4 (Ubuntu 17.4-1.pgdg24.04+2)

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
-- Name: anki; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA anki;


ALTER SCHEMA anki OWNER TO postgres;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: postgres
--

-- *not* creating schema, since initdb creates it


ALTER SCHEMA public OWNER TO postgres;

--
-- Name: delete_deck(text); Type: FUNCTION; Schema: anki; Owner: postgres
--

CREATE FUNCTION anki.delete_deck(input_hash text) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    WITH RECURSIVE subdecks AS (
        SELECT id, human_hash
        FROM decks
        WHERE human_hash = input_hash
        UNION ALL
        SELECT d.id, d.human_hash
        FROM decks d
        JOIN subdecks s ON s.id = d.parent
    )
    DELETE from decks cascade where id in (select id from subdecks);
END;
$$;


ALTER FUNCTION anki.delete_deck(input_hash text) OWNER TO postgres;

--
-- Name: delete_deck_tree(bigint); Type: FUNCTION; Schema: anki; Owner: postgres
--

CREATE FUNCTION anki.delete_deck_tree(p_id bigint) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
   r record;
BEGIN
   DELETE FROM decks WHERE parent = p_id;

   FOR r IN SELECT id FROM decks WHERE parent = p_id LOOP
      PERFORM delete_deck_tree(r.id);
   END LOOP;

   DELETE FROM decks WHERE id = p_id;
END;
$$;


ALTER FUNCTION anki.delete_deck_tree(p_id bigint) OWNER TO postgres;

--
-- Name: delete_deck_tree_by_hash(text); Type: FUNCTION; Schema: anki; Owner: postgres
--

CREATE FUNCTION anki.delete_deck_tree_by_hash(p_hash text) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    r record;
BEGIN
    DELETE FROM decks WHERE human_hash = p_hash;

    FOR r IN (SELECT id FROM decks WHERE parent = (SELECT id FROM decks WHERE human_hash = p_hash)) LOOP
        PERFORM delete_deck_tree(r.id);
    END LOOP;
END;
$$;


ALTER FUNCTION anki.delete_deck_tree_by_hash(p_hash text) OWNER TO postgres;

--
-- Name: normalize_username(); Type: FUNCTION; Schema: anki; Owner: postgres
--

CREATE FUNCTION anki.normalize_username() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.username = TRIM(LOWER(NEW.username));
    RETURN NEW;
END;
$$;


ALTER FUNCTION anki.normalize_username() OWNER TO postgres;

--
-- Name: truncate_tables(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.truncate_tables() RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    statements CURSOR FOR
        SELECT tablename FROM pg_tables
        WHERE tableowner = 'postgres' AND schemaname = 'anki' AND tablename != 'users';
BEGIN
    FOR stmt IN statements LOOP
        EXECUTE 'TRUNCATE TABLE ' || quote_ident(stmt.tablename) || ' CASCADE;';
    END LOOP;
END;
$$;


ALTER FUNCTION public.truncate_tables() OWNER TO postgres;

--
-- Name: truncate_tables(character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.truncate_tables(username character varying) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    statements CURSOR FOR
        SELECT tablename FROM pg_tables
        WHERE tableowner = username AND schemaname = 'public';
BEGIN
    FOR stmt IN statements LOOP
        EXECUTE 'TRUNCATE TABLE ' || quote_ident(stmt.tablename) || ' CASCADE;';
    END LOOP;
END;
$$;


ALTER FUNCTION public.truncate_tables(username character varying) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: auth_tokens; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.auth_tokens (
    id integer NOT NULL,
    user_id integer NOT NULL,
    token_hash bytea NOT NULL,
    refresh_token_hash bytea,
    expires_at timestamp with time zone NOT NULL,
    refresh_expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    last_used_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE anki.auth_tokens OWNER TO postgres;

--
-- Name: auth_tokens_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

ALTER TABLE anki.auth_tokens ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME anki.auth_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: calculated_stats; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.calculated_stats (
    note_id bigint NOT NULL,
    sample_size integer,
    retention real,
    lapses real,
    reps real
);


ALTER TABLE anki.calculated_stats OWNER TO postgres;

--
-- Name: card_deletion_suggestions; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.card_deletion_suggestions (
    id integer NOT NULL,
    note bigint,
    creator_ip character varying(255) NOT NULL,
    commit integer
);


ALTER TABLE anki.card_deletion_suggestions OWNER TO postgres;

--
-- Name: card_deletion_suggestions_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.card_deletion_suggestions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.card_deletion_suggestions_id_seq OWNER TO postgres;

--
-- Name: card_deletion_suggestions_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.card_deletion_suggestions_id_seq OWNED BY anki.card_deletion_suggestions.id;


--
-- Name: changelogs; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.changelogs (
    id bigint NOT NULL,
    deck bigint,
    message text,
    "timestamp" timestamp with time zone
);


ALTER TABLE anki.changelogs OWNER TO postgres;

--
-- Name: changelogs_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.changelogs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.changelogs_id_seq OWNER TO postgres;

--
-- Name: changelogs_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.changelogs_id_seq OWNED BY anki.changelogs.id;


--
-- Name: commits; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.commits (
    commit_id integer NOT NULL,
    rationale integer,
    ip character varying(255),
    "timestamp" timestamp with time zone,
    deck bigint,
    info character varying(255) DEFAULT ''::character varying,
    user_id integer
);


ALTER TABLE anki.commits OWNER TO postgres;

--
-- Name: commits_commit_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.commits_commit_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.commits_commit_id_seq OWNER TO postgres;

--
-- Name: commits_commit_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.commits_commit_id_seq OWNED BY anki.commits.commit_id;


--
-- Name: decks; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.decks (
    id bigint NOT NULL,
    name text DEFAULT NULL::character varying,
    description text DEFAULT NULL::character varying,
    owner integer,
    last_update timestamp with time zone,
    parent bigint,
    crowdanki_uuid text,
    human_hash character varying(255) DEFAULT NULL::character varying,
    creator_ip character varying(255),
    full_path text,
    private boolean DEFAULT true,
    stats_enabled boolean DEFAULT false,
    retention real,
    notes_with_stats_count integer DEFAULT 0,
    restrict_notetypes boolean DEFAULT true,
    restrict_subdecks boolean DEFAULT false
);


ALTER TABLE anki.decks OWNER TO postgres;

--
-- Name: notes; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.notes (
    id bigint NOT NULL,
    guid character varying(255) NOT NULL,
    notetype bigint NOT NULL,
    deck bigint NOT NULL,
    last_update timestamp with time zone,
    reviewed boolean DEFAULT false NOT NULL,
    creator_ip character varying(255),
    deleted boolean DEFAULT false
);


ALTER TABLE anki.notes OWNER TO postgres;

--
-- Name: deck_stats; Type: MATERIALIZED VIEW; Schema: anki; Owner: postgres
--

CREATE MATERIALIZED VIEW anki.deck_stats AS
 WITH RECURSIVE deck_tree AS (
         SELECT decks.id,
            decks.parent,
            decks.id AS root_id
           FROM anki.decks
          WHERE (decks.parent IS NULL)
        UNION ALL
         SELECT d_1.id,
            d_1.parent,
            dt_1.root_id
           FROM (anki.decks d_1
             JOIN deck_tree dt_1 ON ((d_1.parent = dt_1.id)))
        )
 SELECT d.id,
    d.name,
    d.description,
    d.human_hash,
    d.owner,
    d.last_update,
    d.private,
    d.stats_enabled,
        CASE
            WHEN (count(n.*) = 0) THEN '0'::text
            WHEN (count(n.*) < 100) THEN '<100'::text
            WHEN (count(n.*) < 1000) THEN '<1k'::text
            ELSE ((count(n.*) / 1000) || 'k'::text)
        END AS note_count
   FROM ((anki.decks d
     LEFT JOIN deck_tree dt ON ((dt.root_id = d.id)))
     LEFT JOIN anki.notes n ON (((n.deck = dt.id) AND (n.deleted = false))))
  WHERE (d.parent IS NULL)
  GROUP BY d.id, d.name, d.description, d.human_hash, d.owner, d.last_update, d.private, d.stats_enabled
  WITH NO DATA;


ALTER MATERIALIZED VIEW anki.deck_stats OWNER TO postgres;

--
-- Name: decks_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.decks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.decks_id_seq OWNER TO postgres;

--
-- Name: decks_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.decks_id_seq OWNED BY anki.decks.id;


--
-- Name: fields; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.fields (
    id bigint NOT NULL,
    note bigint NOT NULL,
    "position" oid NOT NULL,
    content text,
    reviewed boolean DEFAULT false NOT NULL,
    creator_ip character varying(255),
    commit integer
);


ALTER TABLE anki.fields OWNER TO postgres;

--
-- Name: fields_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.fields_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.fields_id_seq OWNER TO postgres;

--
-- Name: fields_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.fields_id_seq OWNED BY anki.fields.id;


--
-- Name: maintainers; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.maintainers (
    id integer NOT NULL,
    deck bigint,
    user_id integer
);


ALTER TABLE anki.maintainers OWNER TO postgres;

--
-- Name: maintainers_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.maintainers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.maintainers_id_seq OWNER TO postgres;

--
-- Name: maintainers_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.maintainers_id_seq OWNED BY anki.maintainers.id;


--
-- Name: media; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.media (
    id integer NOT NULL,
    filename text NOT NULL,
    deck bigint NOT NULL
);


ALTER TABLE anki.media OWNER TO postgres;

--
-- Name: media_bulk_uploads; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.media_bulk_uploads (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    metadata jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE anki.media_bulk_uploads OWNER TO postgres;

--
-- Name: media_files; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.media_files (
    id bigint NOT NULL,
    hash character varying(64) NOT NULL,
    file_size bigint DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE anki.media_files OWNER TO postgres;

--
-- Name: media_files_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

ALTER TABLE anki.media_files ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME anki.media_files_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: media_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.media_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.media_id_seq OWNER TO postgres;

--
-- Name: media_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.media_id_seq OWNED BY anki.media.id;


--
-- Name: media_operations_log; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.media_operations_log (
    id bigint NOT NULL,
    operation_type integer NOT NULL,
    user_id integer,
    ip_address character varying(45) NOT NULL,
    "timestamp" timestamp with time zone NOT NULL,
    file_hash character varying(64),
    file_name character varying(255),
    file_size bigint
);


ALTER TABLE anki.media_operations_log OWNER TO postgres;

--
-- Name: media_operations_log_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

ALTER TABLE anki.media_operations_log ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME anki.media_operations_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: media_references; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.media_references (
    id bigint NOT NULL,
    media_id bigint NOT NULL,
    note_id bigint,
    file_name character varying(255) NOT NULL
);


ALTER TABLE anki.media_references OWNER TO postgres;

--
-- Name: media_references_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

ALTER TABLE anki.media_references ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME anki.media_references_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: mediafolders; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.mediafolders (
    id integer NOT NULL,
    url character varying(255) NOT NULL,
    deck bigint NOT NULL
);


ALTER TABLE anki.mediafolders OWNER TO postgres;

--
-- Name: mediafolders_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.mediafolders_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.mediafolders_id_seq OWNER TO postgres;

--
-- Name: mediafolders_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.mediafolders_id_seq OWNED BY anki.mediafolders.id;


--
-- Name: note_move_suggestions; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.note_move_suggestions (
    id integer NOT NULL,
    original_deck bigint,
    target_deck bigint,
    note bigint,
    creator_ip character varying(255) NOT NULL,
    commit integer
);


ALTER TABLE anki.note_move_suggestions OWNER TO postgres;

--
-- Name: note_move_suggestions_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.note_move_suggestions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.note_move_suggestions_id_seq OWNER TO postgres;

--
-- Name: note_move_suggestions_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.note_move_suggestions_id_seq OWNED BY anki.note_move_suggestions.id;


--
-- Name: note_stats; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.note_stats (
    id integer NOT NULL,
    note_id bigint,
    user_hash character varying(64),
    retention integer,
    lapses integer,
    reps integer
);


ALTER TABLE anki.note_stats OWNER TO postgres;

--
-- Name: note_stats_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.note_stats_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.note_stats_id_seq OWNER TO postgres;

--
-- Name: note_stats_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.note_stats_id_seq OWNED BY anki.note_stats.id;


--
-- Name: notes_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.notes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.notes_id_seq OWNER TO postgres;

--
-- Name: notes_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.notes_id_seq OWNED BY anki.notes.id;


--
-- Name: notetype; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.notetype (
    id bigint NOT NULL,
    guid text NOT NULL,
    css text,
    latex_post text,
    latex_pre text,
    latex_svg boolean,
    name text,
    type integer DEFAULT 0,
    owner integer,
    original_stock_kind integer DEFAULT 0 NOT NULL,
    sortf integer DEFAULT 0,
    req text
);


ALTER TABLE anki.notetype OWNER TO postgres;

--
-- Name: notetype_field; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.notetype_field (
    id bigint NOT NULL,
    notetype bigint NOT NULL,
    description text,
    font text,
    name text,
    ord integer,
    rtl boolean,
    size integer,
    sticky boolean,
    "position" oid NOT NULL,
    protected boolean DEFAULT false,
    anki_id bigint DEFAULT 0 NOT NULL,
    tag integer DEFAULT 0 NOT NULL
);


ALTER TABLE anki.notetype_field OWNER TO postgres;

--
-- Name: notetype_field_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.notetype_field_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.notetype_field_id_seq OWNER TO postgres;

--
-- Name: notetype_field_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.notetype_field_id_seq OWNED BY anki.notetype_field.id;


--
-- Name: notetype_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.notetype_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.notetype_id_seq OWNER TO postgres;

--
-- Name: notetype_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.notetype_id_seq OWNED BY anki.notetype.id;


--
-- Name: notetype_template; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.notetype_template (
    id bigint NOT NULL,
    notetype bigint NOT NULL,
    qfmt text,
    afmt text,
    bqfmt text,
    bafmt text,
    bfont text,
    bsize integer,
    name text,
    "position" oid NOT NULL,
    anki_id bigint DEFAULT 0 NOT NULL,
    ord integer DEFAULT 0
);


ALTER TABLE anki.notetype_template OWNER TO postgres;

--
-- Name: notetype_template_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.notetype_template_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.notetype_template_id_seq OWNER TO postgres;

--
-- Name: notetype_template_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.notetype_template_id_seq OWNED BY anki.notetype_template.id;


--
-- Name: optional_tags; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.optional_tags (
    id integer NOT NULL,
    deck bigint,
    tag_group character varying(255)
);


ALTER TABLE anki.optional_tags OWNER TO postgres;

--
-- Name: optional_tags_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.optional_tags_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.optional_tags_id_seq OWNER TO postgres;

--
-- Name: optional_tags_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.optional_tags_id_seq OWNED BY anki.optional_tags.id;


--
-- Name: service_accounts; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.service_accounts (
    id integer NOT NULL,
    google_data jsonb NOT NULL,
    folder_id character varying(33) NOT NULL,
    deck bigint NOT NULL
);


ALTER TABLE anki.service_accounts OWNER TO postgres;

--
-- Name: service_accounts_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.service_accounts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.service_accounts_id_seq OWNER TO postgres;

--
-- Name: service_accounts_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.service_accounts_id_seq OWNED BY anki.service_accounts.id;


--
-- Name: subscriptions; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.subscriptions (
    id integer NOT NULL,
    user_hash character varying(255) NOT NULL,
    deck_id bigint NOT NULL
);


ALTER TABLE anki.subscriptions OWNER TO postgres;

--
-- Name: subscriptions_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.subscriptions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.subscriptions_id_seq OWNER TO postgres;

--
-- Name: subscriptions_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.subscriptions_id_seq OWNED BY anki.subscriptions.id;


--
-- Name: tags; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.tags (
    id bigint NOT NULL,
    note bigint NOT NULL,
    content text,
    reviewed boolean DEFAULT false NOT NULL,
    creator_ip character varying(255),
    action boolean DEFAULT true,
    commit integer
);


ALTER TABLE anki.tags OWNER TO postgres;

--
-- Name: tags_id_seq; Type: SEQUENCE; Schema: anki; Owner: postgres
--

CREATE SEQUENCE anki.tags_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE anki.tags_id_seq OWNER TO postgres;

--
-- Name: tags_id_seq; Type: SEQUENCE OWNED BY; Schema: anki; Owner: postgres
--

ALTER SEQUENCE anki.tags_id_seq OWNED BY anki.tags.id;


--
-- Name: user_quotas; Type: TABLE; Schema: anki; Owner: postgres
--

CREATE TABLE anki.user_quotas (
    user_id integer NOT NULL,
    storage_used bigint DEFAULT 0 NOT NULL,
    upload_count integer DEFAULT 0 NOT NULL,
    download_count integer DEFAULT 0 NOT NULL,
    last_reset timestamp with time zone NOT NULL
);


ALTER TABLE anki.user_quotas OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id integer NOT NULL,
    username character varying(254) NOT NULL,
    password character varying(255) NOT NULL,
    is_admin boolean DEFAULT false
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_id_seq OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: card_deletion_suggestions id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.card_deletion_suggestions ALTER COLUMN id SET DEFAULT nextval('anki.card_deletion_suggestions_id_seq'::regclass);


--
-- Name: changelogs id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.changelogs ALTER COLUMN id SET DEFAULT nextval('anki.changelogs_id_seq'::regclass);


--
-- Name: commits commit_id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.commits ALTER COLUMN commit_id SET DEFAULT nextval('anki.commits_commit_id_seq'::regclass);


--
-- Name: decks id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.decks ALTER COLUMN id SET DEFAULT nextval('anki.decks_id_seq'::regclass);


--
-- Name: fields id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.fields ALTER COLUMN id SET DEFAULT nextval('anki.fields_id_seq'::regclass);


--
-- Name: maintainers id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.maintainers ALTER COLUMN id SET DEFAULT nextval('anki.maintainers_id_seq'::regclass);


--
-- Name: media id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.media ALTER COLUMN id SET DEFAULT nextval('anki.media_id_seq'::regclass);


--
-- Name: mediafolders id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.mediafolders ALTER COLUMN id SET DEFAULT nextval('anki.mediafolders_id_seq'::regclass);


--
-- Name: note_move_suggestions id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.note_move_suggestions ALTER COLUMN id SET DEFAULT nextval('anki.note_move_suggestions_id_seq'::regclass);


--
-- Name: note_stats id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.note_stats ALTER COLUMN id SET DEFAULT nextval('anki.note_stats_id_seq'::regclass);


--
-- Name: notes id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.notes ALTER COLUMN id SET DEFAULT nextval('anki.notes_id_seq'::regclass);


--
-- Name: notetype id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.notetype ALTER COLUMN id SET DEFAULT nextval('anki.notetype_id_seq'::regclass);


--
-- Name: notetype_field id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.notetype_field ALTER COLUMN id SET DEFAULT nextval('anki.notetype_field_id_seq'::regclass);


--
-- Name: notetype_template id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.notetype_template ALTER COLUMN id SET DEFAULT nextval('anki.notetype_template_id_seq'::regclass);


--
-- Name: optional_tags id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.optional_tags ALTER COLUMN id SET DEFAULT nextval('anki.optional_tags_id_seq'::regclass);


--
-- Name: service_accounts id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.service_accounts ALTER COLUMN id SET DEFAULT nextval('anki.service_accounts_id_seq'::regclass);


--
-- Name: subscriptions id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.subscriptions ALTER COLUMN id SET DEFAULT nextval('anki.subscriptions_id_seq'::regclass);


--
-- Name: tags id; Type: DEFAULT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.tags ALTER COLUMN id SET DEFAULT nextval('anki.tags_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: auth_tokens auth_tokens_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.auth_tokens
    ADD CONSTRAINT auth_tokens_pkey PRIMARY KEY (id);


--
-- Name: calculated_stats calculated_stats_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.calculated_stats
    ADD CONSTRAINT calculated_stats_pkey PRIMARY KEY (note_id);


--
-- Name: card_deletion_suggestions card_deletion_suggestions_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.card_deletion_suggestions
    ADD CONSTRAINT card_deletion_suggestions_pkey PRIMARY KEY (id);


--
-- Name: changelogs changelogs_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.changelogs
    ADD CONSTRAINT changelogs_pkey PRIMARY KEY (id);


--
-- Name: commits commits_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.commits
    ADD CONSTRAINT commits_pkey PRIMARY KEY (commit_id);


--
-- Name: decks hash_unique; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.decks
    ADD CONSTRAINT hash_unique UNIQUE (human_hash);


--
-- Name: notes idunique; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.notes
    ADD CONSTRAINT idunique UNIQUE (id);


--
-- Name: decks idx_16391_primary; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.decks
    ADD CONSTRAINT idx_16391_primary PRIMARY KEY (id);


--
-- Name: fields idx_16401_primary; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.fields
    ADD CONSTRAINT idx_16401_primary PRIMARY KEY (id);


--
-- Name: notes idx_16409_primary; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.notes
    ADD CONSTRAINT idx_16409_primary PRIMARY KEY (id);


--
-- Name: notetype idx_16415_primary; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.notetype
    ADD CONSTRAINT idx_16415_primary PRIMARY KEY (id);


--
-- Name: notetype_field idx_16422_primary; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.notetype_field
    ADD CONSTRAINT idx_16422_primary PRIMARY KEY (id);


--
-- Name: notetype_template idx_16429_primary; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.notetype_template
    ADD CONSTRAINT idx_16429_primary PRIMARY KEY (id);


--
-- Name: tags idx_16436_primary; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.tags
    ADD CONSTRAINT idx_16436_primary PRIMARY KEY (id);


--
-- Name: maintainers maintainers_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.maintainers
    ADD CONSTRAINT maintainers_pkey PRIMARY KEY (id);


--
-- Name: media_bulk_uploads media_bulk_uploads_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.media_bulk_uploads
    ADD CONSTRAINT media_bulk_uploads_pkey PRIMARY KEY (id);


--
-- Name: media_files media_files_hash_key; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.media_files
    ADD CONSTRAINT media_files_hash_key UNIQUE (hash);


--
-- Name: media_files media_files_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.media_files
    ADD CONSTRAINT media_files_pkey PRIMARY KEY (id);


--
-- Name: media_operations_log media_operations_log_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.media_operations_log
    ADD CONSTRAINT media_operations_log_pkey PRIMARY KEY (id);


--
-- Name: media media_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.media
    ADD CONSTRAINT media_pkey PRIMARY KEY (id);


--
-- Name: media_references media_references_media_id_note_id_key; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.media_references
    ADD CONSTRAINT media_references_media_id_note_id_key UNIQUE (media_id, note_id);


--
-- Name: media_references media_references_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.media_references
    ADD CONSTRAINT media_references_pkey PRIMARY KEY (id);


--
-- Name: mediafolders mediafolders_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.mediafolders
    ADD CONSTRAINT mediafolders_pkey PRIMARY KEY (id);


--
-- Name: mediafolders mediafolders_unique_deck; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.mediafolders
    ADD CONSTRAINT mediafolders_unique_deck UNIQUE (deck);


--
-- Name: mediafolders mediafolders_url_deck_key; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.mediafolders
    ADD CONSTRAINT mediafolders_url_deck_key UNIQUE (url, deck);


--
-- Name: note_move_suggestions note_move_suggestions_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.note_move_suggestions
    ADD CONSTRAINT note_move_suggestions_pkey PRIMARY KEY (id);


--
-- Name: note_stats note_stats_note_id_user_hash_key; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.note_stats
    ADD CONSTRAINT note_stats_note_id_user_hash_key UNIQUE (note_id, user_hash);


--
-- Name: note_stats note_stats_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.note_stats
    ADD CONSTRAINT note_stats_pkey PRIMARY KEY (id);


--
-- Name: card_deletion_suggestions note_unique; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.card_deletion_suggestions
    ADD CONSTRAINT note_unique UNIQUE (note);


--
-- Name: optional_tags optional_tags_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.optional_tags
    ADD CONSTRAINT optional_tags_pkey PRIMARY KEY (id);


--
-- Name: service_accounts service_accounts_deck_key; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.service_accounts
    ADD CONSTRAINT service_accounts_deck_key UNIQUE (deck);


--
-- Name: service_accounts service_accounts_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.service_accounts
    ADD CONSTRAINT service_accounts_pkey PRIMARY KEY (id);


--
-- Name: subscriptions subscriptions_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (id);


--
-- Name: subscriptions unique_ip_deck; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.subscriptions
    ADD CONSTRAINT unique_ip_deck UNIQUE (user_hash, deck_id);


--
-- Name: user_quotas user_quotas_pkey; Type: CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.user_quotas
    ADD CONSTRAINT user_quotas_pkey PRIMARY KEY (user_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (username);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_16391_crowdanki_uuid; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE UNIQUE INDEX idx_16391_crowdanki_uuid ON anki.decks USING btree (crowdanki_uuid);


--
-- Name: idx_16391_owner; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX idx_16391_owner ON anki.decks USING btree (owner);


--
-- Name: idx_16401_note; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX idx_16401_note ON anki.fields USING btree (note);


--
-- Name: idx_16409_deck; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX idx_16409_deck ON anki.notes USING btree (deck);


--
-- Name: idx_16422_notetype; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX idx_16422_notetype ON anki.notetype_field USING btree (notetype);


--
-- Name: idx_16429_notetype; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX idx_16429_notetype ON anki.notetype_template USING btree (notetype);


--
-- Name: idx_16436_note; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX idx_16436_note ON anki.tags USING btree (note);


--
-- Name: idx_auth_tokens_expires_at; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX idx_auth_tokens_expires_at ON anki.auth_tokens USING btree (expires_at);


--
-- Name: idx_auth_tokens_refresh_token_hash; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX idx_auth_tokens_refresh_token_hash ON anki.auth_tokens USING btree (refresh_token_hash);


--
-- Name: idx_auth_tokens_token_hash; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX idx_auth_tokens_token_hash ON anki.auth_tokens USING btree (token_hash);


--
-- Name: idx_auth_tokens_user_id; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE UNIQUE INDEX idx_auth_tokens_user_id ON anki.auth_tokens USING btree (user_id);


--
-- Name: idx_deck_lastupdate; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX idx_deck_lastupdate ON anki.decks USING btree (last_update);


--
-- Name: idx_fields_note_position_reviewed; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE UNIQUE INDEX idx_fields_note_position_reviewed ON anki.fields USING btree (note, "position") WHERE (reviewed = true);


--
-- Name: idx_fields_reviewed_position; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX idx_fields_reviewed_position ON anki.fields USING btree (reviewed, "position");


--
-- Name: idx_guid_owner; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE UNIQUE INDEX idx_guid_owner ON anki.notetype USING btree (guid, owner);


--
-- Name: idx_notes_notetype; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX idx_notes_notetype ON anki.notes USING btree (notetype);


--
-- Name: idx_notetype_field_protected; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX idx_notetype_field_protected ON anki.notetype_field USING btree (protected);


--
-- Name: idx_notetype_field_protected_notetype; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX idx_notetype_field_protected_notetype ON anki.notetype_field USING btree (protected, notetype);


--
-- Name: idx_tags_search; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX idx_tags_search ON anki.tags USING btree (note, reviewed, content);


--
-- Name: media_deck_idx; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX media_deck_idx ON anki.media USING btree (deck);


--
-- Name: media_filename_deck_key; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE UNIQUE INDEX media_filename_deck_key ON anki.media USING btree (md5(filename), deck);


--
-- Name: media_files_hash_idx; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX media_files_hash_idx ON anki.media_files USING btree (hash);


--
-- Name: media_references_file_name_idx; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX media_references_file_name_idx ON anki.media_references USING btree (file_name);


--
-- Name: media_references_media_idx; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX media_references_media_idx ON anki.media_references USING btree (media_id);


--
-- Name: media_references_note_id_idx; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX media_references_note_id_idx ON anki.media_references USING btree (note_id);


--
-- Name: notes_guid_idx; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX notes_guid_idx ON anki.notes USING btree (guid);


--
-- Name: subscriptions_deck_id_idx; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX subscriptions_deck_id_idx ON anki.subscriptions USING btree (deck_id);


--
-- Name: ~card_deletion_suggestions-baa66cc5; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX "~card_deletion_suggestions-baa66cc5" ON anki.card_deletion_suggestions USING btree (commit);


--
-- Name: ~decks-f001e354; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX "~decks-f001e354" ON anki.decks USING btree (parent);


--
-- Name: ~fields-25ac0f75; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX "~fields-25ac0f75" ON anki.fields USING btree (commit) WHERE (NOT reviewed);


--
-- Name: ~fields-921e8dd2; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX "~fields-921e8dd2" ON anki.fields USING btree (note) WHERE (NOT reviewed);


--
-- Name: ~fields-baa66cc5; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX "~fields-baa66cc5" ON anki.fields USING btree (commit);


--
-- Name: ~maintainers-6d3da866; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX "~maintainers-6d3da866" ON anki.maintainers USING btree (user_id);


--
-- Name: ~note_move_suggestions-xx; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX "~note_move_suggestions-xx" ON anki.note_move_suggestions USING btree (commit);


--
-- Name: ~notes-323fd8e9; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX "~notes-323fd8e9" ON anki.notes USING btree (deck, last_update) WHERE reviewed;


--
-- Name: ~notetype-6de94a39; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX "~notetype-6de94a39" ON anki.notetype USING btree (owner);


--
-- Name: ~tags-25ac0f75; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX "~tags-25ac0f75" ON anki.tags USING btree (commit) WHERE (NOT reviewed);


--
-- Name: ~tags-921e8dd2; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX "~tags-921e8dd2" ON anki.tags USING btree (note) WHERE (NOT reviewed);


--
-- Name: ~tags-baa66cc5; Type: INDEX; Schema: anki; Owner: postgres
--

CREATE INDEX "~tags-baa66cc5" ON anki.tags USING btree (commit);


--
-- Name: users normalize_username_trigger; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER normalize_username_trigger BEFORE INSERT OR UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION anki.normalize_username();


--
-- Name: auth_tokens auth_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.auth_tokens
    ADD CONSTRAINT auth_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: calculated_stats calculated_stats_note_id_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.calculated_stats
    ADD CONSTRAINT calculated_stats_note_id_fkey FOREIGN KEY (note_id) REFERENCES anki.notes(id) ON DELETE CASCADE;


--
-- Name: card_deletion_suggestions card_deletion_suggestions_commit_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.card_deletion_suggestions
    ADD CONSTRAINT card_deletion_suggestions_commit_fkey FOREIGN KEY (commit) REFERENCES anki.commits(commit_id) ON DELETE CASCADE;


--
-- Name: card_deletion_suggestions card_deletion_suggestions_note_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.card_deletion_suggestions
    ADD CONSTRAINT card_deletion_suggestions_note_fkey FOREIGN KEY (note) REFERENCES anki.notes(id) ON DELETE CASCADE;


--
-- Name: changelogs changelogs_deck_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.changelogs
    ADD CONSTRAINT changelogs_deck_fkey FOREIGN KEY (deck) REFERENCES anki.decks(id) ON DELETE CASCADE;


--
-- Name: commits commits_deck_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.commits
    ADD CONSTRAINT commits_deck_fkey FOREIGN KEY (deck) REFERENCES anki.decks(id) ON DELETE CASCADE;


--
-- Name: commits commits_user_id_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.commits
    ADD CONSTRAINT commits_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: decks deck_owner_constraint; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.decks
    ADD CONSTRAINT deck_owner_constraint FOREIGN KEY (owner) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: decks decks_parent_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.decks
    ADD CONSTRAINT decks_parent_fkey FOREIGN KEY (parent) REFERENCES anki.decks(id) ON DELETE CASCADE;


--
-- Name: fields fields_commit_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.fields
    ADD CONSTRAINT fields_commit_fkey FOREIGN KEY (commit) REFERENCES anki.commits(commit_id) ON DELETE CASCADE;


--
-- Name: fields fields_note_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.fields
    ADD CONSTRAINT fields_note_fkey FOREIGN KEY (note) REFERENCES anki.notes(id) ON DELETE CASCADE;


--
-- Name: maintainers maintainers_deck_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.maintainers
    ADD CONSTRAINT maintainers_deck_fkey FOREIGN KEY (deck) REFERENCES anki.decks(id) ON DELETE CASCADE;


--
-- Name: maintainers maintainers_user_id_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.maintainers
    ADD CONSTRAINT maintainers_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: media media_deck_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.media
    ADD CONSTRAINT media_deck_fkey FOREIGN KEY (deck) REFERENCES anki.decks(id) ON DELETE CASCADE;


--
-- Name: media_operations_log media_operations_log_user_id_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.media_operations_log
    ADD CONSTRAINT media_operations_log_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: media_references media_references_media_id_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.media_references
    ADD CONSTRAINT media_references_media_id_fkey FOREIGN KEY (media_id) REFERENCES anki.media_files(id) ON DELETE CASCADE;


--
-- Name: media_references media_references_note_id_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.media_references
    ADD CONSTRAINT media_references_note_id_fkey FOREIGN KEY (note_id) REFERENCES anki.notes(id) ON DELETE CASCADE;


--
-- Name: mediafolders mediafolders_deck_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.mediafolders
    ADD CONSTRAINT mediafolders_deck_fkey FOREIGN KEY (deck) REFERENCES anki.decks(id) ON DELETE CASCADE;


--
-- Name: note_move_suggestions note_move_suggestions_commit_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.note_move_suggestions
    ADD CONSTRAINT note_move_suggestions_commit_fkey FOREIGN KEY (commit) REFERENCES anki.commits(commit_id) ON DELETE CASCADE;


--
-- Name: note_move_suggestions note_move_suggestions_note_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.note_move_suggestions
    ADD CONSTRAINT note_move_suggestions_note_fkey FOREIGN KEY (note) REFERENCES anki.notes(id) ON DELETE CASCADE;


--
-- Name: note_move_suggestions note_move_suggestions_original_deck_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.note_move_suggestions
    ADD CONSTRAINT note_move_suggestions_original_deck_fkey FOREIGN KEY (original_deck) REFERENCES anki.decks(id) ON DELETE CASCADE;


--
-- Name: note_move_suggestions note_move_suggestions_target_deck_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.note_move_suggestions
    ADD CONSTRAINT note_move_suggestions_target_deck_fkey FOREIGN KEY (target_deck) REFERENCES anki.decks(id) ON DELETE CASCADE;


--
-- Name: note_stats note_stats_note_id_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.note_stats
    ADD CONSTRAINT note_stats_note_id_fkey FOREIGN KEY (note_id) REFERENCES anki.notes(id) ON DELETE CASCADE;


--
-- Name: notes notes_deck_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.notes
    ADD CONSTRAINT notes_deck_fkey FOREIGN KEY (deck) REFERENCES anki.decks(id) ON DELETE CASCADE;


--
-- Name: notetype_field notetype_field_notetype_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.notetype_field
    ADD CONSTRAINT notetype_field_notetype_fkey FOREIGN KEY (notetype) REFERENCES anki.notetype(id) ON DELETE CASCADE;


--
-- Name: notes notetype_note_constraint; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.notes
    ADD CONSTRAINT notetype_note_constraint FOREIGN KEY (notetype) REFERENCES anki.notetype(id) ON DELETE CASCADE;


--
-- Name: notetype notetype_owner_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.notetype
    ADD CONSTRAINT notetype_owner_fkey FOREIGN KEY (owner) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: notetype_template notetype_template_notetype_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.notetype_template
    ADD CONSTRAINT notetype_template_notetype_fkey FOREIGN KEY (notetype) REFERENCES anki.notetype(id) ON DELETE CASCADE;


--
-- Name: optional_tags optional_tags_deck_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.optional_tags
    ADD CONSTRAINT optional_tags_deck_fkey FOREIGN KEY (deck) REFERENCES anki.decks(id) ON DELETE CASCADE;


--
-- Name: service_accounts service_accounts_deck_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.service_accounts
    ADD CONSTRAINT service_accounts_deck_fkey FOREIGN KEY (deck) REFERENCES anki.decks(id) ON DELETE CASCADE;


--
-- Name: subscriptions subscriptions_deck_id_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.subscriptions
    ADD CONSTRAINT subscriptions_deck_id_fkey FOREIGN KEY (deck_id) REFERENCES anki.decks(id) ON DELETE CASCADE;


--
-- Name: tags tags_commit_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.tags
    ADD CONSTRAINT tags_commit_fkey FOREIGN KEY (commit) REFERENCES anki.commits(commit_id) ON DELETE CASCADE;


--
-- Name: tags tags_note_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.tags
    ADD CONSTRAINT tags_note_fkey FOREIGN KEY (note) REFERENCES anki.notes(id) ON DELETE CASCADE;


--
-- Name: user_quotas user_quotas_user_id_fkey; Type: FK CONSTRAINT; Schema: anki; Owner: postgres
--

ALTER TABLE ONLY anki.user_quotas
    ADD CONSTRAINT user_quotas_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE USAGE ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

