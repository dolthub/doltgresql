-- Downloaded from: https://github.com/artygg/Data-Processing-Goida/blob/514eddeb7aaf6a532acd2cbfcc5194fcc022e02c/sql/Netflix.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 17.2
-- Dumped by pg_dump version 17.2

-- Started on 2025-01-10 11:09:32 CET

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
-- TOC entry 268 (class 1255 OID 16884)
-- Name: fn_audit(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fn_audit() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        INSERT INTO audit_log (table_name, operation, changed_by, old_data)
        VALUES (TG_TABLE_NAME, TG_OP, CURRENT_USER, row_to_json(OLD));
        RETURN OLD;

    ELSIF (TG_OP = 'UPDATE') THEN
        INSERT INTO audit_log (table_name, operation, changed_by, old_data, new_data)
        VALUES (
            TG_TABLE_NAME,
            TG_OP,
            CURRENT_USER,
            row_to_json(OLD),
            row_to_json(NEW)
        );
        RETURN NEW;

    ELSIF (TG_OP = 'INSERT') THEN
        INSERT INTO audit_log (table_name, operation, changed_by, new_data)
        VALUES (
            TG_TABLE_NAME,
            TG_OP,
            CURRENT_USER,
            row_to_json(NEW)
        );
        RETURN NEW;
    END IF;

    RETURN NULL;
END;
$$;


ALTER FUNCTION public.fn_audit() OWNER TO postgres;

--
-- TOC entry 250 (class 1255 OID 16865)
-- Name: fn_audit_subscriptions_delete(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fn_audit_subscriptions_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    INSERT INTO Subscriptions_log
        (subscription_id, profile_id, price_id, start_date, end_date, deleted_at, deleted_by)
    VALUES
        (OLD.id, OLD.profile_id, OLD.price_id, OLD.start_date, OLD.end_date, NOW(), CURRENT_USER);

    RETURN OLD;
END;
$$;


ALTER FUNCTION public.fn_audit_subscriptions_delete() OWNER TO postgres;

--
-- TOC entry 248 (class 1255 OID 16851)
-- Name: fn_block_on_login_faults(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fn_block_on_login_faults() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.login_faults >= 3 THEN
        -- Блокируем на 1 день
        NEW.block_end := NOW() + INTERVAL '1 day';
    ELSE
        -- Если сбросили ниже 3, убираем блокировку
        NEW.block_end := NULL;
    END IF;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.fn_block_on_login_faults() OWNER TO postgres;

--
-- TOC entry 249 (class 1255 OID 16853)
-- Name: fn_check_subscription_dates(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fn_check_subscription_dates() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.end_date IS NOT NULL 
       AND NEW.start_date IS NOT NULL
       AND NEW.end_date < NEW.start_date
    THEN
        RAISE EXCEPTION 'end_date (%) не может быть раньше start_date (%)',
            NEW.end_date, NEW.start_date;
    END IF;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.fn_check_subscription_dates() OWNER TO postgres;

--
-- TOC entry 264 (class 1255 OID 16867)
-- Name: fn_enforce_31day_subscription(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fn_enforce_31day_subscription() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Если указали start_date, но не указали end_date, 
    -- автоматически проставим end_date = start_date + 31 дней
    IF NEW.start_date IS NOT NULL
       AND NEW.end_date IS NULL
    THEN
        NEW.end_date := NEW.start_date + INTERVAL '31 day';
    END IF;

    -- Если обе даты указаны, проверим их корректность
    IF NEW.start_date IS NOT NULL 
       AND NEW.end_date IS NOT NULL
    THEN
        -- 1) end_date >= start_date
        IF NEW.end_date < NEW.start_date THEN
            RAISE EXCEPTION 'end_date (%) не может быть раньше start_date (%)',
                NEW.end_date, NEW.start_date;
        END IF;

        -- 2) Разница не должна превышать 31 день
        IF NEW.end_date > (NEW.start_date + INTERVAL '31 day') THEN
            RAISE EXCEPTION 'Период подписки не может превышать 31 день (start_date=%, end_date=%).',
                NEW.start_date, NEW.end_date;
        END IF;
    END IF;

    RETURN NEW;
END;
$$;


ALTER FUNCTION public.fn_enforce_31day_subscription() OWNER TO postgres;

--
-- TOC entry 265 (class 1255 OID 16869)
-- Name: sp_admin_create_user(character varying, character varying, boolean); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.sp_admin_create_user(IN p_email character varying, IN p_password character varying, IN p_has_referral boolean DEFAULT false)
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
BEGIN
    INSERT INTO Users (email, password, has_used_referral_link)
    VALUES (p_email, p_password, p_has_referral);

    RAISE NOTICE 'New user created with email: %', p_email;
END;
$$;


ALTER PROCEDURE public.sp_admin_create_user(IN p_email character varying, IN p_password character varying, IN p_has_referral boolean) OWNER TO postgres;

--
-- TOC entry 267 (class 1255 OID 16871)
-- Name: sp_admin_delete_user(integer); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.sp_admin_delete_user(IN p_user_id integer)
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
BEGIN
    DELETE FROM Users
     WHERE id = p_user_id;

    IF NOT FOUND THEN
      RAISE EXCEPTION 'User with id % not found', p_user_id;
    END IF;
END;
$$;


ALTER PROCEDURE public.sp_admin_delete_user(IN p_user_id integer) OWNER TO postgres;

--
-- TOC entry 266 (class 1255 OID 16870)
-- Name: sp_admin_update_user(integer, character varying, character varying, boolean); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.sp_admin_update_user(IN p_user_id integer, IN p_email character varying, IN p_password character varying, IN p_has_referral boolean)
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
BEGIN
    UPDATE Users
       SET email = p_email,
           password = p_password,
           has_used_referral_link = p_has_referral
     WHERE id = p_user_id;

    IF NOT FOUND THEN
      RAISE EXCEPTION 'User with id % not found', p_user_id;
    END IF;
END;
$$;


ALTER PROCEDURE public.sp_admin_update_user(IN p_user_id integer, IN p_email character varying, IN p_password character varying, IN p_has_referral boolean) OWNER TO postgres;

--
-- TOC entry 251 (class 1255 OID 16872)
-- Name: sp_create_content(character varying, character varying, text, character varying, double precision, character varying, integer, integer, integer); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.sp_create_content(IN p_title character varying, IN p_poster character varying, IN p_description text, IN p_video_link character varying, IN p_duration double precision, IN p_type character varying, IN p_season integer DEFAULT NULL::integer, IN p_episode_num integer DEFAULT NULL::integer, IN p_series_id integer DEFAULT NULL::integer)
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
DECLARE
    v_new_id INT;
BEGIN
    INSERT INTO Content (title, poster, description, video_link, duration, type, season, episode_number, series_id)
    VALUES (p_title, p_poster, p_description, p_video_link, p_duration, p_type, p_season, p_episode_num, p_series_id)
    RETURNING id INTO v_new_id;

    RAISE NOTICE 'Content created with id %', v_new_id;
END;
$$;


ALTER PROCEDURE public.sp_create_content(IN p_title character varying, IN p_poster character varying, IN p_description text, IN p_video_link character varying, IN p_duration double precision, IN p_type character varying, IN p_season integer, IN p_episode_num integer, IN p_series_id integer) OWNER TO postgres;

--
-- TOC entry 252 (class 1255 OID 16873)
-- Name: sp_delete_content(integer); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.sp_delete_content(IN p_content_id integer)
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
BEGIN
    DELETE FROM Content
     WHERE id = p_content_id;

    IF NOT FOUND THEN
      RAISE EXCEPTION 'Content with id % not found', p_content_id;
    END IF;
END;
$$;


ALTER PROCEDURE public.sp_delete_content(IN p_content_id integer) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 247 (class 1259 OID 16875)
-- Name: audit_log; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.audit_log (
    audit_id integer NOT NULL,
    table_name text NOT NULL,
    operation text NOT NULL,
    changed_by text NOT NULL,
    changed_at timestamp without time zone DEFAULT now() NOT NULL,
    old_data jsonb,
    new_data jsonb
);


ALTER TABLE public.audit_log OWNER TO postgres;

--
-- TOC entry 246 (class 1259 OID 16874)
-- Name: audit_log_audit_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.audit_log_audit_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.audit_log_audit_id_seq OWNER TO postgres;

--
-- TOC entry 3831 (class 0 OID 0)
-- Dependencies: 246
-- Name: audit_log_audit_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.audit_log_audit_id_seq OWNED BY public.audit_log.audit_id;


--
-- TOC entry 231 (class 1259 OID 16741)
-- Name: content; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.content (
    id integer NOT NULL,
    title character varying(255) NOT NULL,
    poster character varying(255),
    description text,
    video_link character varying(255),
    duration double precision,
    type character varying(50),
    season integer,
    episode_number integer,
    series_id integer
);


ALTER TABLE public.content OWNER TO postgres;

--
-- TOC entry 236 (class 1259 OID 16782)
-- Name: genre; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.genre (
    id integer NOT NULL,
    name character varying(100) NOT NULL
);


ALTER TABLE public.genre OWNER TO postgres;

--
-- TOC entry 237 (class 1259 OID 16788)
-- Name: genre_bridge; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.genre_bridge (
    content_id integer NOT NULL,
    genre_id integer NOT NULL
);


ALTER TABLE public.genre_bridge OWNER TO postgres;

--
-- TOC entry 235 (class 1259 OID 16781)
-- Name: genre_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.genre_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.genre_id_seq OWNER TO postgres;

--
-- TOC entry 3835 (class 0 OID 0)
-- Dependencies: 235
-- Name: genre_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.genre_id_seq OWNED BY public.genre.id;


--
-- TOC entry 228 (class 1259 OID 16709)
-- Name: preferences; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.preferences (
    profile_id integer NOT NULL,
    classification_id integer NOT NULL
);


ALTER TABLE public.preferences OWNER TO postgres;

--
-- TOC entry 219 (class 1259 OID 16640)
-- Name: prices; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.prices (
    id integer NOT NULL,
    resolution_id integer NOT NULL,
    price numeric(10,2)
);


ALTER TABLE public.prices OWNER TO postgres;

--
-- TOC entry 218 (class 1259 OID 16639)
-- Name: prices_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.prices_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.prices_id_seq OWNER TO postgres;

--
-- TOC entry 3839 (class 0 OID 0)
-- Dependencies: 218
-- Name: prices_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.prices_id_seq OWNED BY public.prices.id;


--
-- TOC entry 225 (class 1259 OID 16690)
-- Name: profile; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.profile (
    id integer NOT NULL,
    user_id integer NOT NULL,
    profile_image_link character varying(255),
    is_child boolean DEFAULT false,
    language character varying(50)
);


ALTER TABLE public.profile OWNER TO postgres;

--
-- TOC entry 224 (class 1259 OID 16689)
-- Name: profile_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.profile_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.profile_id_seq OWNER TO postgres;

--
-- TOC entry 3842 (class 0 OID 0)
-- Dependencies: 224
-- Name: profile_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.profile_id_seq OWNED BY public.profile.id;


--
-- TOC entry 239 (class 1259 OID 16813)
-- Name: quality_ranges; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.quality_ranges (
    content_id integer NOT NULL,
    resolution_id integer NOT NULL
);


ALTER TABLE public.quality_ranges OWNER TO postgres;

--
-- TOC entry 223 (class 1259 OID 16674)
-- Name: referrals; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.referrals (
    host_id integer NOT NULL,
    invited_id integer NOT NULL
);


ALTER TABLE public.referrals OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 16634)
-- Name: resolutions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.resolutions (
    resolution integer NOT NULL,
    resolution_name character varying(50) NOT NULL
);


ALTER TABLE public.resolutions OWNER TO postgres;

--
-- TOC entry 230 (class 1259 OID 16725)
-- Name: subscriptions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.subscriptions (
    id integer NOT NULL,
    profile_id integer NOT NULL,
    price_id integer NOT NULL,
    start_date date NOT NULL,
    end_date date
);


ALTER TABLE public.subscriptions OWNER TO postgres;

--
-- TOC entry 229 (class 1259 OID 16724)
-- Name: subscriptions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.subscriptions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.subscriptions_id_seq OWNER TO postgres;

--
-- TOC entry 3848 (class 0 OID 0)
-- Dependencies: 229
-- Name: subscriptions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.subscriptions_id_seq OWNED BY public.subscriptions.id;


--
-- TOC entry 245 (class 1259 OID 16856)
-- Name: subscriptions_log; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.subscriptions_log (
    log_id integer NOT NULL,
    subscription_id integer,
    profile_id integer,
    price_id integer,
    start_date date,
    end_date date,
    deleted_at timestamp without time zone DEFAULT now(),
    deleted_by text
);


ALTER TABLE public.subscriptions_log OWNER TO postgres;

--
-- TOC entry 244 (class 1259 OID 16855)
-- Name: subscriptions_log_log_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.subscriptions_log_log_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.subscriptions_log_log_id_seq OWNER TO postgres;

--
-- TOC entry 3850 (class 0 OID 0)
-- Dependencies: 244
-- Name: subscriptions_log_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.subscriptions_log_log_id_seq OWNED BY public.subscriptions_log.log_id;


--
-- TOC entry 238 (class 1259 OID 16803)
-- Name: subtitle; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.subtitle (
    content_id integer NOT NULL,
    language character varying(50) NOT NULL
);


ALTER TABLE public.subtitle OWNER TO postgres;

--
-- TOC entry 227 (class 1259 OID 16703)
-- Name: tag; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tag (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    type character varying(100)
);


ALTER TABLE public.tag OWNER TO postgres;

--
-- TOC entry 226 (class 1259 OID 16702)
-- Name: tag_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.tag_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.tag_id_seq OWNER TO postgres;

--
-- TOC entry 3853 (class 0 OID 0)
-- Dependencies: 226
-- Name: tag_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.tag_id_seq OWNED BY public.tag.id;


--
-- TOC entry 221 (class 1259 OID 16652)
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id integer NOT NULL,
    email character varying(255) NOT NULL,
    password character varying(255) NOT NULL,
    has_used_referral_link boolean DEFAULT false
);


ALTER TABLE public.users OWNER TO postgres;

--
-- TOC entry 220 (class 1259 OID 16651)
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
-- TOC entry 3856 (class 0 OID 0)
-- Dependencies: 220
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- TOC entry 243 (class 1259 OID 16846)
-- Name: v_content_overview; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.v_content_overview AS
 WITH genre_list AS (
         SELECT gb.content_id,
            string_agg((g.name)::text, ', '::text) AS genres
           FROM (public.genre_bridge gb
             JOIN public.genre g ON ((g.id = gb.genre_id)))
          GROUP BY gb.content_id
        ), subtitle_list AS (
         SELECT s.content_id,
            string_agg((s.language)::text, ', '::text) AS subtitles
           FROM public.subtitle s
          GROUP BY s.content_id
        ), resolution_list AS (
         SELECT qr.content_id,
            string_agg((r.resolution_name)::text, ', '::text) AS available_resolutions
           FROM (public.quality_ranges qr
             JOIN public.resolutions r ON ((r.resolution = qr.resolution_id)))
          GROUP BY qr.content_id
        )
 SELECT c.id AS content_id,
    c.title,
    c.poster,
    c.description,
    c.video_link,
    c.duration,
    c.type,
    c.season,
    c.episode_number,
    c.series_id,
    genre_list.genres,
    subtitle_list.subtitles,
    resolution_list.available_resolutions
   FROM (((public.content c
     LEFT JOIN genre_list ON ((genre_list.content_id = c.id)))
     LEFT JOIN subtitle_list ON ((subtitle_list.content_id = c.id)))
     LEFT JOIN resolution_list ON ((resolution_list.content_id = c.id)));


ALTER VIEW public.v_content_overview OWNER TO postgres;

--
-- TOC entry 242 (class 1259 OID 16842)
-- Name: v_profile_preferences; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.v_profile_preferences AS
 SELECT pf.id AS profile_id,
    u.id AS user_id,
    u.email AS user_email,
    t.name AS tag_name,
    t.type AS tag_type
   FROM (((public.preferences pr
     JOIN public.profile pf ON ((pf.id = pr.profile_id)))
     JOIN public.users u ON ((u.id = pf.user_id)))
     JOIN public.tag t ON ((t.id = pr.classification_id)));


ALTER VIEW public.v_profile_preferences OWNER TO postgres;

--
-- TOC entry 240 (class 1259 OID 16832)
-- Name: v_user_subscriptions; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.v_user_subscriptions AS
 SELECT s.id AS subscription_id,
    s.profile_id,
    pf.user_id,
    u.email,
    p.price,
    r.resolution_name,
    s.start_date,
    s.end_date
   FROM ((((public.subscriptions s
     JOIN public.profile pf ON ((pf.id = s.profile_id)))
     JOIN public.users u ON ((u.id = pf.user_id)))
     JOIN public.prices p ON ((p.id = s.price_id)))
     JOIN public.resolutions r ON ((r.resolution = p.resolution_id)))
  WHERE ((s.end_date IS NULL) OR (s.end_date > CURRENT_DATE));


ALTER VIEW public.v_user_subscriptions OWNER TO postgres;

--
-- TOC entry 233 (class 1259 OID 16749)
-- Name: watch_histories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.watch_histories (
    id integer NOT NULL,
    profile_id integer NOT NULL,
    content_id integer NOT NULL,
    stopped_at double precision,
    watching_times integer DEFAULT 1
);


ALTER TABLE public.watch_histories OWNER TO postgres;

--
-- TOC entry 241 (class 1259 OID 16837)
-- Name: v_watch_history_detailed; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.v_watch_history_detailed AS
 SELECT wh.id AS watch_history_id,
    pf.id AS profile_id,
    u.id AS user_id,
    u.email AS user_email,
    c.id AS content_id,
    c.title AS content_title,
    wh.stopped_at,
    wh.watching_times
   FROM (((public.watch_histories wh
     JOIN public.profile pf ON ((pf.id = wh.profile_id)))
     JOIN public.users u ON ((u.id = pf.user_id)))
     JOIN public.content c ON ((c.id = wh.content_id)));


ALTER VIEW public.v_watch_history_detailed OWNER TO postgres;

--
-- TOC entry 222 (class 1259 OID 16663)
-- Name: warnings; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.warnings (
    user_id integer NOT NULL,
    login_faults integer DEFAULT 0,
    block_end timestamp without time zone
);


ALTER TABLE public.warnings OWNER TO postgres;

--
-- TOC entry 232 (class 1259 OID 16748)
-- Name: watch_histories_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.watch_histories_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.watch_histories_id_seq OWNER TO postgres;

--
-- TOC entry 3864 (class 0 OID 0)
-- Dependencies: 232
-- Name: watch_histories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.watch_histories_id_seq OWNED BY public.watch_histories.id;


--
-- TOC entry 234 (class 1259 OID 16766)
-- Name: watch_later; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.watch_later (
    profile_id integer NOT NULL,
    content_id integer NOT NULL
);


ALTER TABLE public.watch_later OWNER TO postgres;

--
-- TOC entry 3565 (class 2604 OID 16878)
-- Name: audit_log audit_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.audit_log ALTER COLUMN audit_id SET DEFAULT nextval('public.audit_log_audit_id_seq'::regclass);


--
-- TOC entry 3562 (class 2604 OID 16785)
-- Name: genre id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.genre ALTER COLUMN id SET DEFAULT nextval('public.genre_id_seq'::regclass);


--
-- TOC entry 3552 (class 2604 OID 16643)
-- Name: prices id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.prices ALTER COLUMN id SET DEFAULT nextval('public.prices_id_seq'::regclass);


--
-- TOC entry 3556 (class 2604 OID 16693)
-- Name: profile id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.profile ALTER COLUMN id SET DEFAULT nextval('public.profile_id_seq'::regclass);


--
-- TOC entry 3559 (class 2604 OID 16728)
-- Name: subscriptions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions ALTER COLUMN id SET DEFAULT nextval('public.subscriptions_id_seq'::regclass);


--
-- TOC entry 3563 (class 2604 OID 16859)
-- Name: subscriptions_log log_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions_log ALTER COLUMN log_id SET DEFAULT nextval('public.subscriptions_log_log_id_seq'::regclass);


--
-- TOC entry 3558 (class 2604 OID 16706)
-- Name: tag id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag ALTER COLUMN id SET DEFAULT nextval('public.tag_id_seq'::regclass);


--
-- TOC entry 3553 (class 2604 OID 16655)
-- Name: users id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- TOC entry 3560 (class 2604 OID 16752)
-- Name: watch_histories id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.watch_histories ALTER COLUMN id SET DEFAULT nextval('public.watch_histories_id_seq'::regclass);


--
-- TOC entry 3819 (class 0 OID 16875)
-- Dependencies: 247
-- Data for Name: audit_log; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.audit_log (audit_id, table_name, operation, changed_by, changed_at, old_data, new_data) FROM stdin;
\.


--
-- TOC entry 3807 (class 0 OID 16741)
-- Dependencies: 231
-- Data for Name: content; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.content (id, title, poster, description, video_link, duration, type, season, episode_number, series_id) FROM stdin;
\.


--
-- TOC entry 3812 (class 0 OID 16782)
-- Dependencies: 236
-- Data for Name: genre; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.genre (id, name) FROM stdin;
\.


--
-- TOC entry 3813 (class 0 OID 16788)
-- Dependencies: 237
-- Data for Name: genre_bridge; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.genre_bridge (content_id, genre_id) FROM stdin;
\.


--
-- TOC entry 3804 (class 0 OID 16709)
-- Dependencies: 228
-- Data for Name: preferences; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.preferences (profile_id, classification_id) FROM stdin;
\.


--
-- TOC entry 3795 (class 0 OID 16640)
-- Dependencies: 219
-- Data for Name: prices; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.prices (id, resolution_id, price) FROM stdin;
\.


--
-- TOC entry 3801 (class 0 OID 16690)
-- Dependencies: 225
-- Data for Name: profile; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.profile (id, user_id, profile_image_link, is_child, language) FROM stdin;
\.


--
-- TOC entry 3815 (class 0 OID 16813)
-- Dependencies: 239
-- Data for Name: quality_ranges; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.quality_ranges (content_id, resolution_id) FROM stdin;
\.


--
-- TOC entry 3799 (class 0 OID 16674)
-- Dependencies: 223
-- Data for Name: referrals; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.referrals (host_id, invited_id) FROM stdin;
\.


--
-- TOC entry 3793 (class 0 OID 16634)
-- Dependencies: 217
-- Data for Name: resolutions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.resolutions (resolution, resolution_name) FROM stdin;
\.


--
-- TOC entry 3806 (class 0 OID 16725)
-- Dependencies: 230
-- Data for Name: subscriptions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.subscriptions (id, profile_id, price_id, start_date, end_date) FROM stdin;
\.


--
-- TOC entry 3817 (class 0 OID 16856)
-- Dependencies: 245
-- Data for Name: subscriptions_log; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.subscriptions_log (log_id, subscription_id, profile_id, price_id, start_date, end_date, deleted_at, deleted_by) FROM stdin;
\.


--
-- TOC entry 3814 (class 0 OID 16803)
-- Dependencies: 238
-- Data for Name: subtitle; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.subtitle (content_id, language) FROM stdin;
\.


--
-- TOC entry 3803 (class 0 OID 16703)
-- Dependencies: 227
-- Data for Name: tag; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tag (id, name, type) FROM stdin;
\.


--
-- TOC entry 3797 (class 0 OID 16652)
-- Dependencies: 221
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, email, password, has_used_referral_link) FROM stdin;
\.


--
-- TOC entry 3798 (class 0 OID 16663)
-- Dependencies: 222
-- Data for Name: warnings; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.warnings (user_id, login_faults, block_end) FROM stdin;
\.


--
-- TOC entry 3809 (class 0 OID 16749)
-- Dependencies: 233
-- Data for Name: watch_histories; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.watch_histories (id, profile_id, content_id, stopped_at, watching_times) FROM stdin;
\.


--
-- TOC entry 3810 (class 0 OID 16766)
-- Dependencies: 234
-- Data for Name: watch_later; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.watch_later (profile_id, content_id) FROM stdin;
\.


--
-- TOC entry 3867 (class 0 OID 0)
-- Dependencies: 246
-- Name: audit_log_audit_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.audit_log_audit_id_seq', 1, false);


--
-- TOC entry 3868 (class 0 OID 0)
-- Dependencies: 235
-- Name: genre_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.genre_id_seq', 1, false);


--
-- TOC entry 3869 (class 0 OID 0)
-- Dependencies: 218
-- Name: prices_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.prices_id_seq', 1, false);


--
-- TOC entry 3870 (class 0 OID 0)
-- Dependencies: 224
-- Name: profile_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.profile_id_seq', 1, false);


--
-- TOC entry 3871 (class 0 OID 0)
-- Dependencies: 229
-- Name: subscriptions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.subscriptions_id_seq', 1, false);


--
-- TOC entry 3872 (class 0 OID 0)
-- Dependencies: 244
-- Name: subscriptions_log_log_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.subscriptions_log_log_id_seq', 1, false);


--
-- TOC entry 3873 (class 0 OID 0)
-- Dependencies: 226
-- Name: tag_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.tag_id_seq', 1, false);


--
-- TOC entry 3874 (class 0 OID 0)
-- Dependencies: 220
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_id_seq', 1, false);


--
-- TOC entry 3875 (class 0 OID 0)
-- Dependencies: 232
-- Name: watch_histories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.watch_histories_id_seq', 1, false);


--
-- TOC entry 3604 (class 2606 OID 16883)
-- Name: audit_log audit_log_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.audit_log
    ADD CONSTRAINT audit_log_pkey PRIMARY KEY (audit_id);


--
-- TOC entry 3588 (class 2606 OID 16747)
-- Name: content content_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.content
    ADD CONSTRAINT content_pkey PRIMARY KEY (id);


--
-- TOC entry 3596 (class 2606 OID 16792)
-- Name: genre_bridge genre_bridge_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.genre_bridge
    ADD CONSTRAINT genre_bridge_pkey PRIMARY KEY (content_id, genre_id);


--
-- TOC entry 3594 (class 2606 OID 16787)
-- Name: genre genre_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.genre
    ADD CONSTRAINT genre_pkey PRIMARY KEY (id);


--
-- TOC entry 3584 (class 2606 OID 16713)
-- Name: preferences preferences_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.preferences
    ADD CONSTRAINT preferences_pkey PRIMARY KEY (profile_id, classification_id);


--
-- TOC entry 3570 (class 2606 OID 16645)
-- Name: prices prices_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.prices
    ADD CONSTRAINT prices_pkey PRIMARY KEY (id);


--
-- TOC entry 3580 (class 2606 OID 16696)
-- Name: profile profile_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.profile
    ADD CONSTRAINT profile_pkey PRIMARY KEY (id);


--
-- TOC entry 3600 (class 2606 OID 16817)
-- Name: quality_ranges quality_ranges_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quality_ranges
    ADD CONSTRAINT quality_ranges_pkey PRIMARY KEY (content_id, resolution_id);


--
-- TOC entry 3578 (class 2606 OID 16678)
-- Name: referrals referrals_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.referrals
    ADD CONSTRAINT referrals_pkey PRIMARY KEY (host_id, invited_id);


--
-- TOC entry 3568 (class 2606 OID 16638)
-- Name: resolutions resolutions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.resolutions
    ADD CONSTRAINT resolutions_pkey PRIMARY KEY (resolution);


--
-- TOC entry 3602 (class 2606 OID 16864)
-- Name: subscriptions_log subscriptions_log_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions_log
    ADD CONSTRAINT subscriptions_log_pkey PRIMARY KEY (log_id);


--
-- TOC entry 3586 (class 2606 OID 16730)
-- Name: subscriptions subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (id);


--
-- TOC entry 3598 (class 2606 OID 16807)
-- Name: subtitle subtitle_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subtitle
    ADD CONSTRAINT subtitle_pkey PRIMARY KEY (content_id, language);


--
-- TOC entry 3582 (class 2606 OID 16708)
-- Name: tag tag_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag
    ADD CONSTRAINT tag_pkey PRIMARY KEY (id);


--
-- TOC entry 3572 (class 2606 OID 16662)
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- TOC entry 3574 (class 2606 OID 16660)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- TOC entry 3576 (class 2606 OID 16668)
-- Name: warnings warnings_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.warnings
    ADD CONSTRAINT warnings_pkey PRIMARY KEY (user_id);


--
-- TOC entry 3590 (class 2606 OID 16755)
-- Name: watch_histories watch_histories_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.watch_histories
    ADD CONSTRAINT watch_histories_pkey PRIMARY KEY (id);


--
-- TOC entry 3592 (class 2606 OID 16770)
-- Name: watch_later watch_later_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.watch_later
    ADD CONSTRAINT watch_later_pkey PRIMARY KEY (profile_id, content_id);


--
-- TOC entry 3636 (class 2620 OID 16901)
-- Name: content trg_audit_content; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_content AFTER INSERT OR DELETE OR UPDATE ON public.content FOR EACH ROW EXECUTE FUNCTION public.fn_audit();


--
-- TOC entry 3639 (class 2620 OID 16894)
-- Name: genre trg_audit_genre; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_genre AFTER INSERT OR DELETE OR UPDATE ON public.genre FOR EACH ROW EXECUTE FUNCTION public.fn_audit();


--
-- TOC entry 3640 (class 2620 OID 16895)
-- Name: genre_bridge trg_audit_genre_bridge; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_genre_bridge AFTER INSERT OR DELETE OR UPDATE ON public.genre_bridge FOR EACH ROW EXECUTE FUNCTION public.fn_audit();


--
-- TOC entry 3631 (class 2620 OID 16890)
-- Name: preferences trg_audit_preferences; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_preferences AFTER INSERT OR DELETE OR UPDATE ON public.preferences FOR EACH ROW EXECUTE FUNCTION public.fn_audit();


--
-- TOC entry 3624 (class 2620 OID 16887)
-- Name: prices trg_audit_prices; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_prices AFTER INSERT OR DELETE OR UPDATE ON public.prices FOR EACH ROW EXECUTE FUNCTION public.fn_audit();


--
-- TOC entry 3629 (class 2620 OID 16899)
-- Name: profile trg_audit_profile; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_profile AFTER INSERT OR DELETE OR UPDATE ON public.profile FOR EACH ROW EXECUTE FUNCTION public.fn_audit();


--
-- TOC entry 3642 (class 2620 OID 16897)
-- Name: quality_ranges trg_audit_quality_ranges; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_quality_ranges AFTER INSERT OR DELETE OR UPDATE ON public.quality_ranges FOR EACH ROW EXECUTE FUNCTION public.fn_audit();


--
-- TOC entry 3628 (class 2620 OID 16889)
-- Name: referrals trg_audit_referrals; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_referrals AFTER INSERT OR DELETE OR UPDATE ON public.referrals FOR EACH ROW EXECUTE FUNCTION public.fn_audit();


--
-- TOC entry 3623 (class 2620 OID 16886)
-- Name: resolutions trg_audit_resolutions; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_resolutions AFTER INSERT OR DELETE OR UPDATE ON public.resolutions FOR EACH ROW EXECUTE FUNCTION public.fn_audit();


--
-- TOC entry 3632 (class 2620 OID 16900)
-- Name: subscriptions trg_audit_subscriptions; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_subscriptions AFTER INSERT OR DELETE OR UPDATE ON public.subscriptions FOR EACH ROW EXECUTE FUNCTION public.fn_audit();


--
-- TOC entry 3633 (class 2620 OID 16866)
-- Name: subscriptions trg_audit_subscriptions_delete; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_subscriptions_delete AFTER DELETE ON public.subscriptions FOR EACH ROW EXECUTE FUNCTION public.fn_audit_subscriptions_delete();


--
-- TOC entry 3643 (class 2620 OID 16902)
-- Name: subscriptions_log trg_audit_subscriptions_log; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_subscriptions_log AFTER INSERT OR DELETE OR UPDATE ON public.subscriptions_log FOR EACH ROW EXECUTE FUNCTION public.fn_audit();


--
-- TOC entry 3641 (class 2620 OID 16896)
-- Name: subtitle trg_audit_subtitle; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_subtitle AFTER INSERT OR DELETE OR UPDATE ON public.subtitle FOR EACH ROW EXECUTE FUNCTION public.fn_audit();


--
-- TOC entry 3630 (class 2620 OID 16891)
-- Name: tag trg_audit_tag; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_tag AFTER INSERT OR DELETE OR UPDATE ON public.tag FOR EACH ROW EXECUTE FUNCTION public.fn_audit();


--
-- TOC entry 3625 (class 2620 OID 16898)
-- Name: users trg_audit_users; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_users AFTER INSERT OR DELETE OR UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.fn_audit();


--
-- TOC entry 3626 (class 2620 OID 16888)
-- Name: warnings trg_audit_warnings; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_warnings AFTER INSERT OR DELETE OR UPDATE ON public.warnings FOR EACH ROW EXECUTE FUNCTION public.fn_audit();


--
-- TOC entry 3637 (class 2620 OID 16892)
-- Name: watch_histories trg_audit_watch_histories; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_watch_histories AFTER INSERT OR DELETE OR UPDATE ON public.watch_histories FOR EACH ROW EXECUTE FUNCTION public.fn_audit();


--
-- TOC entry 3638 (class 2620 OID 16893)
-- Name: watch_later trg_audit_watch_later; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_audit_watch_later AFTER INSERT OR DELETE OR UPDATE ON public.watch_later FOR EACH ROW EXECUTE FUNCTION public.fn_audit();


--
-- TOC entry 3627 (class 2620 OID 16852)
-- Name: warnings trg_block_on_login_faults; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_block_on_login_faults BEFORE INSERT OR UPDATE ON public.warnings FOR EACH ROW EXECUTE FUNCTION public.fn_block_on_login_faults();


--
-- TOC entry 3634 (class 2620 OID 16854)
-- Name: subscriptions trg_check_subscription_dates; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_check_subscription_dates BEFORE INSERT OR UPDATE ON public.subscriptions FOR EACH ROW EXECUTE FUNCTION public.fn_check_subscription_dates();


--
-- TOC entry 3635 (class 2620 OID 16868)
-- Name: subscriptions trg_enforce_31day_subscription; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_enforce_31day_subscription BEFORE INSERT OR UPDATE ON public.subscriptions FOR EACH ROW EXECUTE FUNCTION public.fn_enforce_31day_subscription();


--
-- TOC entry 3618 (class 2606 OID 16793)
-- Name: genre_bridge fk_gb_content; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.genre_bridge
    ADD CONSTRAINT fk_gb_content FOREIGN KEY (content_id) REFERENCES public.content(id) ON DELETE CASCADE;


--
-- TOC entry 3619 (class 2606 OID 16798)
-- Name: genre_bridge fk_gb_genre; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.genre_bridge
    ADD CONSTRAINT fk_gb_genre FOREIGN KEY (genre_id) REFERENCES public.genre(id) ON DELETE CASCADE;


--
-- TOC entry 3610 (class 2606 OID 16714)
-- Name: preferences fk_preferences_profile; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.preferences
    ADD CONSTRAINT fk_preferences_profile FOREIGN KEY (profile_id) REFERENCES public.profile(id) ON DELETE CASCADE;


--
-- TOC entry 3611 (class 2606 OID 16719)
-- Name: preferences fk_preferences_tag; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.preferences
    ADD CONSTRAINT fk_preferences_tag FOREIGN KEY (classification_id) REFERENCES public.tag(id) ON DELETE CASCADE;


--
-- TOC entry 3605 (class 2606 OID 16646)
-- Name: prices fk_prices_resolution; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.prices
    ADD CONSTRAINT fk_prices_resolution FOREIGN KEY (resolution_id) REFERENCES public.resolutions(resolution) ON DELETE RESTRICT;


--
-- TOC entry 3609 (class 2606 OID 16697)
-- Name: profile fk_profile_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.profile
    ADD CONSTRAINT fk_profile_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 3621 (class 2606 OID 16818)
-- Name: quality_ranges fk_quality_content; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quality_ranges
    ADD CONSTRAINT fk_quality_content FOREIGN KEY (content_id) REFERENCES public.content(id) ON DELETE CASCADE;


--
-- TOC entry 3622 (class 2606 OID 16823)
-- Name: quality_ranges fk_quality_resolution; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quality_ranges
    ADD CONSTRAINT fk_quality_resolution FOREIGN KEY (resolution_id) REFERENCES public.resolutions(resolution) ON DELETE RESTRICT;


--
-- TOC entry 3607 (class 2606 OID 16679)
-- Name: referrals fk_referrals_host; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.referrals
    ADD CONSTRAINT fk_referrals_host FOREIGN KEY (host_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 3608 (class 2606 OID 16684)
-- Name: referrals fk_referrals_invited; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.referrals
    ADD CONSTRAINT fk_referrals_invited FOREIGN KEY (invited_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 3612 (class 2606 OID 16736)
-- Name: subscriptions fk_subscriptions_price; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT fk_subscriptions_price FOREIGN KEY (price_id) REFERENCES public.prices(id) ON DELETE RESTRICT;


--
-- TOC entry 3613 (class 2606 OID 16731)
-- Name: subscriptions fk_subscriptions_profile; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT fk_subscriptions_profile FOREIGN KEY (profile_id) REFERENCES public.profile(id) ON DELETE CASCADE;


--
-- TOC entry 3620 (class 2606 OID 16808)
-- Name: subtitle fk_subtitle_content; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subtitle
    ADD CONSTRAINT fk_subtitle_content FOREIGN KEY (content_id) REFERENCES public.content(id) ON DELETE CASCADE;


--
-- TOC entry 3606 (class 2606 OID 16669)
-- Name: warnings fk_warnings_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.warnings
    ADD CONSTRAINT fk_warnings_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 3614 (class 2606 OID 16761)
-- Name: watch_histories fk_watch_hist_content; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.watch_histories
    ADD CONSTRAINT fk_watch_hist_content FOREIGN KEY (content_id) REFERENCES public.content(id) ON DELETE CASCADE;


--
-- TOC entry 3615 (class 2606 OID 16756)
-- Name: watch_histories fk_watch_hist_profile; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.watch_histories
    ADD CONSTRAINT fk_watch_hist_profile FOREIGN KEY (profile_id) REFERENCES public.profile(id) ON DELETE CASCADE;


--
-- TOC entry 3616 (class 2606 OID 16776)
-- Name: watch_later fk_watch_later_content; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.watch_later
    ADD CONSTRAINT fk_watch_later_content FOREIGN KEY (content_id) REFERENCES public.content(id) ON DELETE CASCADE;


--
-- TOC entry 3617 (class 2606 OID 16771)
-- Name: watch_later fk_watch_later_profile; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.watch_later
    ADD CONSTRAINT fk_watch_later_profile FOREIGN KEY (profile_id) REFERENCES public.profile(id) ON DELETE CASCADE;


--
-- TOC entry 3825 (class 0 OID 0)
-- Dependencies: 5
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: pg_database_owner
--

GRANT ALL ON SCHEMA public TO app_admin;


--
-- TOC entry 3826 (class 0 OID 0)
-- Dependencies: 265
-- Name: PROCEDURE sp_admin_create_user(IN p_email character varying, IN p_password character varying, IN p_has_referral boolean); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON PROCEDURE public.sp_admin_create_user(IN p_email character varying, IN p_password character varying, IN p_has_referral boolean) TO app_admin;


--
-- TOC entry 3827 (class 0 OID 0)
-- Dependencies: 267
-- Name: PROCEDURE sp_admin_delete_user(IN p_user_id integer); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON PROCEDURE public.sp_admin_delete_user(IN p_user_id integer) TO app_admin;


--
-- TOC entry 3828 (class 0 OID 0)
-- Dependencies: 266
-- Name: PROCEDURE sp_admin_update_user(IN p_user_id integer, IN p_email character varying, IN p_password character varying, IN p_has_referral boolean); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON PROCEDURE public.sp_admin_update_user(IN p_user_id integer, IN p_email character varying, IN p_password character varying, IN p_has_referral boolean) TO app_admin;


--
-- TOC entry 3829 (class 0 OID 0)
-- Dependencies: 251
-- Name: PROCEDURE sp_create_content(IN p_title character varying, IN p_poster character varying, IN p_description text, IN p_video_link character varying, IN p_duration double precision, IN p_type character varying, IN p_season integer, IN p_episode_num integer, IN p_series_id integer); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON PROCEDURE public.sp_create_content(IN p_title character varying, IN p_poster character varying, IN p_description text, IN p_video_link character varying, IN p_duration double precision, IN p_type character varying, IN p_season integer, IN p_episode_num integer, IN p_series_id integer) TO content_manager;


--
-- TOC entry 3830 (class 0 OID 0)
-- Dependencies: 252
-- Name: PROCEDURE sp_delete_content(IN p_content_id integer); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON PROCEDURE public.sp_delete_content(IN p_content_id integer) TO content_manager;


--
-- TOC entry 3832 (class 0 OID 0)
-- Dependencies: 231
-- Name: TABLE content; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.content TO app_admin;
GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.content TO content_manager;
GRANT SELECT ON TABLE public.content TO analytics_viewer;


--
-- TOC entry 3833 (class 0 OID 0)
-- Dependencies: 236
-- Name: TABLE genre; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.genre TO app_admin;
GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.genre TO content_manager;
GRANT SELECT ON TABLE public.genre TO analytics_viewer;


--
-- TOC entry 3834 (class 0 OID 0)
-- Dependencies: 237
-- Name: TABLE genre_bridge; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.genre_bridge TO app_admin;
GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.genre_bridge TO content_manager;
GRANT SELECT ON TABLE public.genre_bridge TO analytics_viewer;


--
-- TOC entry 3836 (class 0 OID 0)
-- Dependencies: 235
-- Name: SEQUENCE genre_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.genre_id_seq TO app_admin;


--
-- TOC entry 3837 (class 0 OID 0)
-- Dependencies: 228
-- Name: TABLE preferences; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.preferences TO app_admin;
GRANT SELECT ON TABLE public.preferences TO analytics_viewer;


--
-- TOC entry 3838 (class 0 OID 0)
-- Dependencies: 219
-- Name: TABLE prices; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.prices TO app_admin;
GRANT SELECT ON TABLE public.prices TO analytics_viewer;


--
-- TOC entry 3840 (class 0 OID 0)
-- Dependencies: 218
-- Name: SEQUENCE prices_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.prices_id_seq TO app_admin;


--
-- TOC entry 3841 (class 0 OID 0)
-- Dependencies: 225
-- Name: TABLE profile; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.profile TO app_admin;
GRANT SELECT ON TABLE public.profile TO analytics_viewer;


--
-- TOC entry 3843 (class 0 OID 0)
-- Dependencies: 224
-- Name: SEQUENCE profile_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.profile_id_seq TO app_admin;


--
-- TOC entry 3844 (class 0 OID 0)
-- Dependencies: 239
-- Name: TABLE quality_ranges; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.quality_ranges TO app_admin;
GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.quality_ranges TO content_manager;
GRANT SELECT ON TABLE public.quality_ranges TO analytics_viewer;


--
-- TOC entry 3845 (class 0 OID 0)
-- Dependencies: 223
-- Name: TABLE referrals; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.referrals TO app_admin;
GRANT SELECT ON TABLE public.referrals TO analytics_viewer;


--
-- TOC entry 3846 (class 0 OID 0)
-- Dependencies: 217
-- Name: TABLE resolutions; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.resolutions TO app_admin;
GRANT SELECT ON TABLE public.resolutions TO analytics_viewer;


--
-- TOC entry 3847 (class 0 OID 0)
-- Dependencies: 230
-- Name: TABLE subscriptions; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.subscriptions TO app_admin;
GRANT SELECT ON TABLE public.subscriptions TO analytics_viewer;


--
-- TOC entry 3849 (class 0 OID 0)
-- Dependencies: 229
-- Name: SEQUENCE subscriptions_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.subscriptions_id_seq TO app_admin;


--
-- TOC entry 3851 (class 0 OID 0)
-- Dependencies: 238
-- Name: TABLE subtitle; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.subtitle TO app_admin;
GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.subtitle TO content_manager;
GRANT SELECT ON TABLE public.subtitle TO analytics_viewer;


--
-- TOC entry 3852 (class 0 OID 0)
-- Dependencies: 227
-- Name: TABLE tag; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.tag TO app_admin;
GRANT SELECT ON TABLE public.tag TO analytics_viewer;


--
-- TOC entry 3854 (class 0 OID 0)
-- Dependencies: 226
-- Name: SEQUENCE tag_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.tag_id_seq TO app_admin;


--
-- TOC entry 3855 (class 0 OID 0)
-- Dependencies: 221
-- Name: TABLE users; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.users TO app_admin;
GRANT SELECT ON TABLE public.users TO analytics_viewer;


--
-- TOC entry 3857 (class 0 OID 0)
-- Dependencies: 220
-- Name: SEQUENCE users_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.users_id_seq TO app_admin;


--
-- TOC entry 3858 (class 0 OID 0)
-- Dependencies: 243
-- Name: TABLE v_content_overview; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.v_content_overview TO app_admin;
GRANT SELECT ON TABLE public.v_content_overview TO analytics_viewer;


--
-- TOC entry 3859 (class 0 OID 0)
-- Dependencies: 242
-- Name: TABLE v_profile_preferences; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.v_profile_preferences TO app_admin;
GRANT SELECT ON TABLE public.v_profile_preferences TO analytics_viewer;


--
-- TOC entry 3860 (class 0 OID 0)
-- Dependencies: 240
-- Name: TABLE v_user_subscriptions; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.v_user_subscriptions TO app_admin;
GRANT SELECT ON TABLE public.v_user_subscriptions TO analytics_viewer;


--
-- TOC entry 3861 (class 0 OID 0)
-- Dependencies: 233
-- Name: TABLE watch_histories; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.watch_histories TO app_admin;
GRANT SELECT ON TABLE public.watch_histories TO analytics_viewer;


--
-- TOC entry 3862 (class 0 OID 0)
-- Dependencies: 241
-- Name: TABLE v_watch_history_detailed; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.v_watch_history_detailed TO app_admin;
GRANT SELECT ON TABLE public.v_watch_history_detailed TO analytics_viewer;


--
-- TOC entry 3863 (class 0 OID 0)
-- Dependencies: 222
-- Name: TABLE warnings; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.warnings TO app_admin;
GRANT SELECT ON TABLE public.warnings TO analytics_viewer;


--
-- TOC entry 3865 (class 0 OID 0)
-- Dependencies: 232
-- Name: SEQUENCE watch_histories_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.watch_histories_id_seq TO app_admin;


--
-- TOC entry 3866 (class 0 OID 0)
-- Dependencies: 234
-- Name: TABLE watch_later; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.watch_later TO app_admin;
GRANT SELECT ON TABLE public.watch_later TO analytics_viewer;


-- Completed on 2025-01-10 11:09:32 CET

--
-- PostgreSQL database dump complete
--

