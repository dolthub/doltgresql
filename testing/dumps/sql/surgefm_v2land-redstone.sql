-- Downloaded from: https://github.com/surgefm/v2land-redstone/blob/72b40f85bfa7fba631ebcbf5a54c6ca7b2914288/scripts/db-02.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 12.1
-- Dumped by pg_dump version 12.2

-- Started on 2020-04-11 03:41:07 EDT

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
-- TOC entry 607 (class 1247 OID 37511)
-- Name: enum_auth_site; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_auth_site AS ENUM (
    'twitter',
    'weibo',
    'email'
);


ALTER TYPE public.enum_auth_site OWNER TO v2land;

CREATE TYPE public."enum_resourceLock_status" AS ENUM (
    'active',
    'unlocked',
    'expired'
);


ALTER TYPE public."enum_resourceLock_status" OWNER TO v2land;

--
-- TOC entry 695 (class 1247 OID 37518)
-- Name: enum_authorizationAccessToken_status; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public."enum_authorizationAccessToken_status" AS ENUM (
    'active',
    'revoked'
);


ALTER TYPE public."enum_authorizationAccessToken_status" OWNER TO v2land;

--
-- TOC entry 698 (class 1247 OID 37524)
-- Name: enum_client_role; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_client_role AS ENUM (
    'admin',
    'manager',
    'contributor'
);


ALTER TYPE public.enum_client_role OWNER TO v2land;

--
-- TOC entry 701 (class 1247 OID 37532)
-- Name: enum_contact_method; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_contact_method AS ENUM (
    'twitter',
    'weibo',
    'twitterAt',
    'weiboAt',
    'email',
    'emailDailyReport',
    'mobileAppNotification'
);


ALTER TYPE public.enum_contact_method OWNER TO v2land;

--
-- TOC entry 704 (class 1247 OID 37548)
-- Name: enum_contact_status; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_contact_status AS ENUM (
    'active',
    'inactive',
    'expired'
);


ALTER TYPE public.enum_contact_status OWNER TO v2land;

--
-- TOC entry 707 (class 1247 OID 37556)
-- Name: enum_contact_type; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_contact_type AS ENUM (
    'email',
    'twitter',
    'weibo',
    'telegram',
    'mobileApp'
);


ALTER TYPE public.enum_contact_type OWNER TO v2land;

--
-- TOC entry 710 (class 1247 OID 37568)
-- Name: enum_critique_status; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_critique_status AS ENUM (
    'pending',
    'admitted',
    'rejected',
    'removed'
);


ALTER TYPE public.enum_critique_status OWNER TO v2land;

--
-- TOC entry 713 (class 1247 OID 37578)
-- Name: enum_event_status; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_event_status AS ENUM (
    'pending',
    'admitted',
    'rejected',
    'hidden',
    'removed'
);


ALTER TYPE public.enum_event_status OWNER TO v2land;

--
-- TOC entry 716 (class 1247 OID 37590)
-- Name: enum_news_status; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_news_status AS ENUM (
    'pending',
    'admitted',
    'rejected',
    'removed'
);


ALTER TYPE public.enum_news_status OWNER TO v2land;

--
-- TOC entry 719 (class 1247 OID 37600)
-- Name: enum_notification_mode; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_notification_mode AS ENUM (
    'EveryNewStack',
    '30DaysSinceLatestStack',
    'new',
    '7DaysSinceLatestNews',
    'daily',
    'weekly',
    'monthly',
    'EveryFriday'
);


ALTER TYPE public.enum_notification_mode OWNER TO v2land;

--
-- TOC entry 722 (class 1247 OID 37618)
-- Name: enum_notification_status; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_notification_status AS ENUM (
    'pending',
    'ongoing',
    'complete',
    'discarded'
);


ALTER TYPE public.enum_notification_status OWNER TO v2land;

--
-- TOC entry 725 (class 1247 OID 37628)
-- Name: enum_record_action; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_record_action AS ENUM (
    'createEvent',
    'updateEventStatus',
    'updateEventDetail',
    'createEventHeaderImage',
    'updateEventHeaderImage',
    'createStack',
    'updateStackStatus',
    'updateStackDetail',
    'invalidateStack',
    'notifyNewStack',
    'createNews',
    'updateNewsStatus',
    'updateNewsDetail',
    'notifyNewNews',
    'createSubscription',
    'updateSubscription',
    'cancelSubscription',
    'addModeToSubscription',
    'createClient',
    'updateClientRole',
    'updateClientDetail',
    'updateClientPassword',
    'createClientVerificationToken',
    'authorizeThirdPartyAccount',
    'unauthorizeThirdPartyAccount',
    'notify',
    'sendEmailDailyReport',
    'sendWeeklyDailyReport',
    'sendMonthlyDailyReport',
    'addContactToSubscription',
    'removeSubscriptionContact'
);


ALTER TYPE public.enum_record_action OWNER TO v2land;

--
-- TOC entry 728 (class 1247 OID 37692)
-- Name: enum_record_model; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_record_model AS ENUM (
    'Event',
    'Stack',
    'News',
    'Client',
    'HeaderImage',
    'Subscription',
    'Auth',
    'Report',
    'Miscellaneous',
    'Contact'
);


ALTER TYPE public.enum_record_model OWNER TO v2land;

--
-- TOC entry 731 (class 1247 OID 37714)
-- Name: enum_record_operation; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_record_operation AS ENUM (
    'create',
    'update',
    'destroy'
);


ALTER TYPE public.enum_record_operation OWNER TO v2land;

--
-- TOC entry 734 (class 1247 OID 37722)
-- Name: enum_reportNotification_status; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public."enum_reportNotification_status" AS ENUM (
    'pending',
    'complete',
    'invalid'
);


ALTER TYPE public."enum_reportNotification_status" OWNER TO v2land;

--
-- TOC entry 737 (class 1247 OID 37730)
-- Name: enum_report_method; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_report_method AS ENUM (
    'email',
    'telegram'
);


ALTER TYPE public.enum_report_method OWNER TO v2land;

--
-- TOC entry 740 (class 1247 OID 37736)
-- Name: enum_report_status; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_report_status AS ENUM (
    'pending',
    'ongoing',
    'complete',
    'invalid'
);


ALTER TYPE public.enum_report_status OWNER TO v2land;

--
-- TOC entry 743 (class 1247 OID 37746)
-- Name: enum_report_type; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_report_type AS ENUM (
    'daily',
    'weekly',
    'monthly'
);


ALTER TYPE public.enum_report_type OWNER TO v2land;

--
-- TOC entry 746 (class 1247 OID 37754)
-- Name: enum_stack_status; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_stack_status AS ENUM (
    'pending',
    'admitted',
    'invalid',
    'rejected',
    'hidden',
    'removed'
);


ALTER TYPE public.enum_stack_status OWNER TO v2land;

--
-- TOC entry 749 (class 1247 OID 37768)
-- Name: enum_subscription_mode; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_subscription_mode AS ENUM (
    'EveryNewStack',
    '30DaysSinceLatestStack',
    'new',
    '7DaysSinceLatestNews',
    'daily',
    'weekly',
    'monthly',
    'EveryFriday'
);


ALTER TYPE public.enum_subscription_mode OWNER TO v2land;

--
-- TOC entry 752 (class 1247 OID 37786)
-- Name: enum_subscription_status; Type: TYPE; Schema: public; Owner: v2land
--

CREATE TYPE public.enum_subscription_status AS ENUM (
    'active',
    'unsubscribed'
);


ALTER TYPE public.enum_subscription_status OWNER TO v2land;

--
-- TOC entry 840 (class 1247 OID 38237)
-- Name: enum_tag_status; Type: TYPE; Schema: public; Owner: zehuali
--

CREATE TYPE public.enum_tag_status AS ENUM (
    'visible',
    'hidden'
);


ALTER TYPE public.enum_tag_status OWNER TO zehuali;

--
-- TOC entry 269 (class 1255 OID 37791)
-- Name: sails_session_store_clear(); Type: FUNCTION; Schema: public; Owner: v2land
--

CREATE FUNCTION public.sails_session_store_clear() RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  DELETE FROM sails_session_store;
END;
$$;


ALTER FUNCTION public.sails_session_store_clear() OWNER TO v2land;

--
-- TOC entry 270 (class 1255 OID 37792)
-- Name: sails_session_store_destroy(text); Type: FUNCTION; Schema: public; Owner: v2land
--

CREATE FUNCTION public.sails_session_store_destroy(sid_in text) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  DELETE FROM sails_session_store WHERE sid = sid_in;
END;
$$;


ALTER FUNCTION public.sails_session_store_destroy(sid_in text) OWNER TO v2land;

--
-- TOC entry 271 (class 1255 OID 37793)
-- Name: sails_session_store_get(text); Type: FUNCTION; Schema: public; Owner: v2land
--

CREATE FUNCTION public.sails_session_store_get(sid_in text, OUT data_out json) RETURNS json
    LANGUAGE plpgsql
    AS $$
BEGIN
  SELECT data FROM sails_session_store WHERE sid = sid_in INTO data_out;
END;
$$;


ALTER FUNCTION public.sails_session_store_get(sid_in text, OUT data_out json) OWNER TO v2land;

--
-- TOC entry 272 (class 1255 OID 37794)
-- Name: sails_session_store_length(); Type: FUNCTION; Schema: public; Owner: v2land
--

CREATE FUNCTION public.sails_session_store_length(OUT length integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
BEGIN
  SELECT count(*) FROM sails_session_store INTO length;
END;
$$;


ALTER FUNCTION public.sails_session_store_length(OUT length integer) OWNER TO v2land;

--
-- TOC entry 273 (class 1255 OID 37795)
-- Name: sails_session_store_set(text, json); Type: FUNCTION; Schema: public; Owner: v2land
--

CREATE FUNCTION public.sails_session_store_set(sid_in text, data_in json) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  -- delete current session data if it exists so the next insert succeeds
  DELETE FROM sails_session_store WHERE sid = sid_in;
  INSERT INTO sails_session_store(sid, data) VALUES(sid_in, data_in);
END;
$$;


ALTER FUNCTION public.sails_session_store_set(sid_in text, data_in json) OWNER TO v2land;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 266 (class 1259 OID 38360)
-- Name: Session; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public."Session" (
    sid character varying(36),
    expires timestamp with time zone,
    data text,
    "createdAt" timestamp with time zone,
    "updatedAt" timestamp with time zone
);


ALTER TABLE public."Session" OWNER TO v2land;

--
-- TOC entry 230 (class 1259 OID 37802)
-- Name: auth; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public.auth (
    site public.enum_auth_site,
    "profileId" text,
    profile jsonb,
    token text,
    "tokenSecret" text,
    "accessToken" text,
    "accessTokenSecret" text,
    "refreshToken" text,
    redirect text,
    owner integer,
    id integer NOT NULL,
    "createdAt" timestamp with time zone,
    "updatedAt" timestamp with time zone
);


ALTER TABLE public.auth OWNER TO v2land;

CREATE SEQUENCE public."site_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.site
(
    id integer NOT NULL DEFAULT nextval('public."site_id_seq"'::regclass),
    name character varying COLLATE pg_catalog."default" NOT NULL,
    domains text[] COLLATE pg_catalog."default" NOT NULL,
    "shouldParseFulltext" boolean NOT NULL DEFAULT true,
    "dynamicLoading" boolean NOT NULL DEFAULT false,
    "rssUrls" text[] COLLATE pg_catalog."default",
    "rssUrl" text COLLATE pg_catalog."default",
    "updatedAt" timestamp without time zone,
    "createdAt" timestamp without time zone,
    homepage text COLLATE pg_catalog."default",
    icon text COLLATE pg_catalog."default",
    description text COLLATE pg_catalog."default",
    PRIMARY KEY (id)
);

ALTER TABLE public.site OWNER to v2land;

CREATE SEQUENCE public."siteAccount_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public."siteAccount"
(
    id integer NOT NULL DEFAULT nextval('public."siteAccount_id_seq"'::regclass),
    "siteId" integer,
    avatar text,
    username text NOT NULL,
    homepage text,
    "createdAt" timestamp with time zone,
    "updatedAt" timestamp with time zone,
    PRIMARY KEY (id)
);

ALTER TABLE public."siteAccount" OWNER to v2land;

--
-- TOC entry 231 (class 1259 OID 37808)
-- Name: auth_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

CREATE SEQUENCE public.auth_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.auth_id_seq OWNER TO v2land;

CREATE SEQUENCE public."eventContributor_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

--
-- TOC entry 3547 (class 0 OID 0)
-- Dependencies: 231
-- Name: auth_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: v2land
--

ALTER SEQUENCE public.auth_id_seq OWNED BY public.auth.id;


--
-- TOC entry 232 (class 1259 OID 37810)
-- Name: authorizationAccessToken; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public."authorizationAccessToken" (
    id integer NOT NULL,
    token text NOT NULL,
    "refreshToken" text,
    expire timestamp with time zone,
    "authorizationClientId" integer NOT NULL,
    owner integer NOT NULL,
    status public."enum_authorizationAccessToken_status" NOT NULL,
    "createdAt" timestamp with time zone NOT NULL,
    "updatedAt" timestamp with time zone NOT NULL
);


ALTER TABLE public."authorizationAccessToken" OWNER TO v2land;

--
-- TOC entry 233 (class 1259 OID 37816)
-- Name: authorizationAccessToken_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

ALTER TABLE public."authorizationAccessToken" ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public."authorizationAccessToken_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 234 (class 1259 OID 37818)
-- Name: authorizationClient; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public."authorizationClient" (
    name text NOT NULL,
    description text,
    "redirectURI" text,
    "allowAuthorizationByCredentials" boolean NOT NULL,
    id integer NOT NULL,
    "createdAt" timestamp with time zone,
    "updatedAt" timestamp with time zone
);


ALTER TABLE public."authorizationClient" OWNER TO v2land;

--
-- TOC entry 235 (class 1259 OID 37824)
-- Name: authorizationClient_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

ALTER TABLE public."authorizationClient" ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public."authorizationClient_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 236 (class 1259 OID 37826)
-- Name: authorizationCode; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public."authorizationCode" (
    id integer NOT NULL,
    code text NOT NULL,
    url text,
    expire timestamp with time zone NOT NULL,
    owner integer NOT NULL,
    "authorizationClientId" integer NOT NULL,
    "createdAt" timestamp with time zone NOT NULL,
    "updatedAt" timestamp with time zone NOT NULL
);


ALTER TABLE public."authorizationCode" OWNER TO v2land;

--
-- TOC entry 237 (class 1259 OID 37832)
-- Name: authorizationCode_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

ALTER TABLE public."authorizationCode" ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public."authorizationCode_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 238 (class 1259 OID 37834)
-- Name: client; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public.client (
    username text,
    email text,
    password text,
    nickname text,
    description text,
    avatar text,
    id integer NOT NULL,
    "createdAt" timestamp with time zone,
    "updatedAt" timestamp with time zone,
    "emailVerified" boolean DEFAULT false,
    settings jsonb
);


ALTER TABLE public.client OWNER TO v2land;

--
-- TOC entry 239 (class 1259 OID 37846)
-- Name: client_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

CREATE SEQUENCE public.client_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.client_id_seq OWNER TO v2land;

--
-- TOC entry 3548 (class 0 OID 0)
-- Dependencies: 239
-- Name: client_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: v2land
--

ALTER SEQUENCE public.client_id_seq OWNED BY public.client.id;


--
-- TOC entry 268 (class 1259 OID 38385)
-- Name: commit; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public.commit (
    id integer NOT NULL,
    summary text,
    description text,
    data jsonb,
    diff jsonb,
    "time" timestamp with time zone,
    "parentId" integer,
    "authorId" integer,
    "eventId" integer,
    "createdAt" timestamp with time zone NOT NULL,
    "updatedAt" timestamp with time zone NOT NULL,
    "isForkCommit" boolean
);


ALTER TABLE public.commit OWNER TO v2land;

--
-- TOC entry 267 (class 1259 OID 38383)
-- Name: commit_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

CREATE SEQUENCE public.commit_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.commit_id_seq OWNER TO v2land;

--
-- TOC entry 3549 (class 0 OID 0)
-- Dependencies: 267
-- Name: commit_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: v2land
--

ALTER SEQUENCE public.commit_id_seq OWNED BY public.commit.id;


--
-- TOC entry 240 (class 1259 OID 37848)
-- Name: contact_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

CREATE SEQUENCE public.contact_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.contact_id_seq OWNER TO v2land;

--
-- TOC entry 241 (class 1259 OID 37850)
-- Name: contact; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public.contact (
    id integer DEFAULT nextval('public.contact_id_seq'::regclass) NOT NULL,
    "profileId" text,
    type public.enum_contact_type NOT NULL,
    method public.enum_contact_method NOT NULL,
    status public.enum_contact_status NOT NULL,
    "createdAt" timestamp with time zone NOT NULL,
    "updatedAt" timestamp with time zone NOT NULL,
    owner integer,
    "subscriptionId" integer,
    "authId" integer,
    "unsubscribeId" text
);


ALTER TABLE public.contact OWNER TO v2land;

--
-- TOC entry 242 (class 1259 OID 37857)
-- Name: critique_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

CREATE SEQUENCE public.critique_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.critique_id_seq OWNER TO v2land;

--
-- TOC entry 243 (class 1259 OID 37859)
-- Name: critique; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public.critique (
    id integer DEFAULT nextval('public.critique_id_seq'::regclass) NOT NULL,
    url text NOT NULL,
    source text NOT NULL,
    title text NOT NULL,
    abstract text NOT NULL,
    "time" timestamp with time zone NOT NULL,
    status public.enum_critique_status DEFAULT 'pending'::public.enum_critique_status,
    "createdAt" timestamp with time zone NOT NULL,
    "updatedAt" timestamp with time zone NOT NULL,
    "eventId" integer
);


ALTER TABLE public.critique OWNER TO v2land;

--
-- TOC entry 244 (class 1259 OID 37867)
-- Name: event; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public.event (
    name text,
    description text,
    status public.enum_event_status,
    id integer NOT NULL,
    "createdAt" timestamp with time zone,
    "updatedAt" timestamp with time zone,
    pinyin text,
    "latestAdmittedNewsId" integer,
    "ownerId" integer,
    "parentId" integer
);


ALTER TABLE public.event OWNER TO v2land;

CREATE TABLE public."eventContributor" (
    "id" integer NOT NULL,
	"eventId" integer NOT NULL,
	"contributorId" integer NOT NULL,
    "commitId" integer,
    "parentId" integer,
	"points" double precision NOT NULL,
	"createdAt" timestamp without time zone,
	"updatedAt" timestamp without time zone,
	PRIMARY KEY ("id")
);

ALTER TABLE public."eventContributor" OWNER TO v2land;

--
-- TOC entry 265 (class 1259 OID 38333)
-- Name: eventStackNews; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public."eventStackNews" (
    "eventId" integer NOT NULL,
    "newsId" integer NOT NULL,
    "stackId" integer,
    "createdAt" timestamp with time zone NOT NULL,
    "updatedAt" timestamp with time zone NOT NULL
);


ALTER TABLE public."eventStackNews" OWNER TO v2land;

--
-- TOC entry 245 (class 1259 OID 37873)
-- Name: eventTag; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public."eventTag" (
    "eventId" integer,
    "tagId" integer,
    id integer NOT NULL,
    "updatedAt" timestamp with time zone,
    "createdAt" timestamp with time zone
);


ALTER TABLE public."eventTag" OWNER TO v2land;

--
-- TOC entry 246 (class 1259 OID 37876)
-- Name: eventTag_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

CREATE SEQUENCE public."eventTag_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public."eventTag_id_seq" OWNER TO v2land;

--
-- TOC entry 3550 (class 0 OID 0)
-- Dependencies: 246
-- Name: eventTag_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: v2land
--

ALTER SEQUENCE public."eventTag_id_seq" OWNED BY public."eventTag".id;


--
-- TOC entry 247 (class 1259 OID 37878)
-- Name: event_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

CREATE SEQUENCE public.event_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.event_id_seq OWNER TO v2land;

--
-- TOC entry 3551 (class 0 OID 0)
-- Dependencies: 247
-- Name: event_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: v2land
--

ALTER SEQUENCE public.event_id_seq OWNED BY public.event.id;


--
-- TOC entry 248 (class 1259 OID 37880)
-- Name: headerImage; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public."headerImage" (
    "imageUrl" text,
    source text,
    "sourceUrl" text,
    "eventId" integer,
    id integer NOT NULL,
    "createdAt" timestamp with time zone,
    "updatedAt" timestamp with time zone
);


ALTER TABLE public."headerImage" OWNER TO v2land;

--
-- TOC entry 249 (class 1259 OID 37886)
-- Name: headerimage_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

CREATE SEQUENCE public.headerimage_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.headerimage_id_seq OWNER TO v2land;

--
-- TOC entry 3552 (class 0 OID 0)
-- Dependencies: 249
-- Name: headerimage_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: v2land
--

ALTER SEQUENCE public.headerimage_id_seq OWNED BY public."headerImage".id;


--
-- TOC entry 250 (class 1259 OID 37894)
-- Name: news; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public.news (
    url text,
    source text,
    title text,
    abstract text,
    "time" timestamp with time zone,
    status public.enum_news_status,
    comment text,
    id integer NOT NULL,
    "siteId" integer,
    "siteAccountId" integer,
    "createdAt" timestamp with time zone,
    "updatedAt" timestamp with time zone,
    "isInTemporaryStack" boolean
);


ALTER TABLE public.news OWNER TO v2land;

--
-- TOC entry 251 (class 1259 OID 37900)
-- Name: news_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

CREATE SEQUENCE public.news_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.news_id_seq OWNER TO v2land;

--
-- TOC entry 3553 (class 0 OID 0)
-- Dependencies: 251
-- Name: news_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: v2land
--

ALTER SEQUENCE public.news_id_seq OWNED BY public.news.id;


--
-- TOC entry 252 (class 1259 OID 37902)
-- Name: notification; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public.notification (
    "time" date,
    mode public.enum_notification_mode,
    "eventId" integer,
    status public.enum_notification_status,
    id integer NOT NULL,
    "createdAt" timestamp with time zone,
    "updatedAt" timestamp with time zone
);


ALTER TABLE public.notification OWNER TO v2land;

--
-- TOC entry 253 (class 1259 OID 37905)
-- Name: notification_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

CREATE SEQUENCE public.notification_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.notification_id_seq OWNER TO v2land;

--
-- TOC entry 3554 (class 0 OID 0)
-- Dependencies: 253
-- Name: notification_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: v2land
--

ALTER SEQUENCE public.notification_id_seq OWNED BY public.notification.id;


--
-- TOC entry 254 (class 1259 OID 37907)
-- Name: record; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public.record (
    model character varying(256),
    target integer,
    operation public.enum_record_operation,
    action character varying(256),
    data jsonb,
    owner integer,
    id integer NOT NULL,
    "createdAt" timestamp with time zone,
    "updatedAt" timestamp with time zone,
    before jsonb,
    subtarget integer
);


ALTER TABLE public.record OWNER TO v2land;

--
-- TOC entry 255 (class 1259 OID 37913)
-- Name: record_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

CREATE SEQUENCE public.record_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.record_id_seq OWNER TO v2land;

--
-- TOC entry 3555 (class 0 OID 0)
-- Dependencies: 255
-- Name: record_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: v2land
--

ALTER SEQUENCE public.record_id_seq OWNED BY public.record.id;


--
-- TOC entry 256 (class 1259 OID 37915)
-- Name: report_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

CREATE SEQUENCE public.report_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.report_id_seq OWNER TO v2land;

--
-- TOC entry 257 (class 1259 OID 37917)
-- Name: report; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public.report (
    id integer DEFAULT nextval('public.report_id_seq'::regclass) NOT NULL,
    "time" timestamp with time zone DEFAULT '2018-10-27 02:08:13.457+08'::timestamp with time zone NOT NULL,
    type public.enum_report_type DEFAULT 'daily'::public.enum_report_type,
    method public.enum_report_method DEFAULT 'email'::public.enum_report_method,
    status public.enum_report_status DEFAULT 'pending'::public.enum_report_status,
    "createdAt" timestamp with time zone NOT NULL,
    "updatedAt" timestamp with time zone NOT NULL,
    owner integer
);


ALTER TABLE public.report OWNER TO v2land;

--
-- TOC entry 258 (class 1259 OID 37925)
-- Name: reportNotification; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public."reportNotification" (
    status public."enum_reportNotification_status" DEFAULT 'pending'::public."enum_reportNotification_status",
    "createdAt" timestamp with time zone NOT NULL,
    "updatedAt" timestamp with time zone NOT NULL,
    "notificationId" integer NOT NULL,
    "reportId" integer NOT NULL
);


ALTER TABLE public."reportNotification" OWNER TO v2land;

CREATE SEQUENCE public."resourceLock_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public."resourceLock_id_seq" OWNER TO v2land;

--
-- TOC entry 257 (class 1259 OID 37917)
-- Name: report; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public."resourceLock" (
    id integer DEFAULT nextval('public."resourceLock_id_seq"'::regclass) NOT NULL,
    "expires" timestamp with time zone DEFAULT '2018-10-27 02:08:13.457+08'::timestamp with time zone NOT NULL,
    status public."enum_resourceLock_status" DEFAULT 'active'::public."enum_resourceLock_status",
    "createdAt" timestamp with time zone NOT NULL,
    "updatedAt" timestamp with time zone NOT NULL,
    locker integer NOT NULL,
    "eventId" integer,
    model text NOT NULL,
    "resourceId" integer NOT NULL
);


ALTER TABLE public.report OWNER TO v2land;

--
-- TOC entry 259 (class 1259 OID 37936)
-- Name: stack_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

CREATE SEQUENCE public.stack_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 2147483647
    CACHE 1;


ALTER TABLE public.stack_id_seq OWNER TO v2land;

--
-- TOC entry 260 (class 1259 OID 37938)
-- Name: stack; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public.stack (
    title text,
    description text,
    status public.enum_stack_status,
    "eventId" integer,
    "stackEventId" integer,
    "createdAt" timestamp with time zone,
    "updatedAt" timestamp with time zone,
    id integer DEFAULT nextval('public.stack_id_seq'::regclass) NOT NULL,
    "order" integer DEFAULT '-1'::integer NOT NULL,
    "time" timestamp with time zone
);


ALTER TABLE public.stack OWNER TO v2land;

--
-- TOC entry 261 (class 1259 OID 37946)
-- Name: subscription; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public.subscription (
    mode public.enum_subscription_mode,
    status public.enum_subscription_status,
    "unsubscribeId" text,
    subscriber integer,
    "eventId" integer,
    id integer NOT NULL,
    "createdAt" timestamp with time zone,
    "updatedAt" timestamp with time zone
);


ALTER TABLE public.subscription OWNER TO v2land;

--
-- TOC entry 262 (class 1259 OID 37952)
-- Name: subscription_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

CREATE SEQUENCE public.subscription_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.subscription_id_seq OWNER TO v2land;

--
-- TOC entry 3556 (class 0 OID 0)
-- Dependencies: 262
-- Name: subscription_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: v2land
--

ALTER SEQUENCE public.subscription_id_seq OWNED BY public.subscription.id;


--
-- TOC entry 263 (class 1259 OID 37954)
-- Name: tag; Type: TABLE; Schema: public; Owner: v2land
--

CREATE TABLE public.tag (
    id integer NOT NULL,
    name text NOT NULL,
    slug text,
    description text,
    "redirectToId" integer,
    "createdAt" timestamp with time zone,
    "updatedAt" timestamp with time zone,
    status public.enum_tag_status
);


ALTER TABLE public.tag OWNER TO v2land;

CREATE SEQUENCE public."tagCurator_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public."tagCurator_id_seq" OWNER TO v2land;

CREATE TABLE public."tagCurator" (
    id integer NOT NULL DEFAULT nextval('public."tagCurator_id_seq"'::regclass),
    "tagId" integer NOT NULL,
    "curatorId" integer NOT NULL,
    primary key ("id")
);


ALTER TABLE public."tagCurator" OWNER TO v2land;

--
-- TOC entry 264 (class 1259 OID 37960)
-- Name: tag_id_seq; Type: SEQUENCE; Schema: public; Owner: v2land
--

CREATE SEQUENCE public.tag_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.tag_id_seq OWNER TO v2land;

--
-- TOC entry 3557 (class 0 OID 0)
-- Dependencies: 264
-- Name: tag_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: v2land
--

ALTER SEQUENCE public.tag_id_seq OWNED BY public.tag.id;


--
-- TOC entry 3288 (class 2604 OID 37962)
-- Name: auth id; Type: DEFAULT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.auth ALTER COLUMN id SET DEFAULT nextval('public.auth_id_seq'::regclass);


--
-- TOC entry 3290 (class 2604 OID 37963)
-- Name: client id; Type: DEFAULT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.client ALTER COLUMN id SET DEFAULT nextval('public.client_id_seq'::regclass);


--
-- TOC entry 3310 (class 2604 OID 38388)
-- Name: commit id; Type: DEFAULT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.commit ALTER COLUMN id SET DEFAULT nextval('public.commit_id_seq'::regclass);


--
-- TOC entry 3294 (class 2604 OID 37965)
-- Name: event id; Type: DEFAULT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.event ALTER COLUMN id SET DEFAULT nextval('public.event_id_seq'::regclass);

ALTER TABLE ONLY public."eventContributor" ALTER COLUMN id SET DEFAULT nextval('public."eventContributor_id_seq"'::regclass);


--
-- TOC entry 3295 (class 2604 OID 37966)
-- Name: eventTag id; Type: DEFAULT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."eventTag" ALTER COLUMN id SET DEFAULT nextval('public."eventTag_id_seq"'::regclass);


--
-- TOC entry 3296 (class 2604 OID 37967)
-- Name: headerImage id; Type: DEFAULT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."headerImage" ALTER COLUMN id SET DEFAULT nextval('public.headerimage_id_seq'::regclass);


--
-- TOC entry 3297 (class 2604 OID 37968)
-- Name: news id; Type: DEFAULT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.news ALTER COLUMN id SET DEFAULT nextval('public.news_id_seq'::regclass);


--
-- TOC entry 3298 (class 2604 OID 37969)
-- Name: notification id; Type: DEFAULT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.notification ALTER COLUMN id SET DEFAULT nextval('public.notification_id_seq'::regclass);


--
-- TOC entry 3299 (class 2604 OID 37970)
-- Name: record id; Type: DEFAULT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.record ALTER COLUMN id SET DEFAULT nextval('public.record_id_seq'::regclass);


--
-- TOC entry 3308 (class 2604 OID 37971)
-- Name: subscription id; Type: DEFAULT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.subscription ALTER COLUMN id SET DEFAULT nextval('public.subscription_id_seq'::regclass);


--
-- TOC entry 3309 (class 2604 OID 37972)
-- Name: tag id; Type: DEFAULT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.tag ALTER COLUMN id SET DEFAULT nextval('public.tag_id_seq'::regclass);


--
-- TOC entry 3539 (class 0 OID 38360)
-- Dependencies: 266
-- Data for Name: Session; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public."Session" (sid, expires, data, "createdAt", "updatedAt") FROM stdin;
\.


--
-- TOC entry 3503 (class 0 OID 37802)
-- Dependencies: 230
-- Data for Name: auth; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public.auth (site, "profileId", profile, token, "tokenSecret", "accessToken", "accessTokenSecret", "refreshToken", redirect, owner, id, "createdAt", "updatedAt") FROM stdin;
\.


--
-- TOC entry 3505 (class 0 OID 37810)
-- Dependencies: 232
-- Data for Name: authorizationAccessToken; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public."authorizationAccessToken" (id, token, "refreshToken", expire, "authorizationClientId", owner, status, "createdAt", "updatedAt") FROM stdin;
\.


--
-- TOC entry 3507 (class 0 OID 37818)
-- Dependencies: 234
-- Data for Name: authorizationClient; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public."authorizationClient" (name, description, "redirectURI", "allowAuthorizationByCredentials", id, "createdAt", "updatedAt") FROM stdin;
\.


--
-- TOC entry 3509 (class 0 OID 37826)
-- Dependencies: 236
-- Data for Name: authorizationCode; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public."authorizationCode" (id, code, url, expire, owner, "authorizationClientId", "createdAt", "updatedAt") FROM stdin;
\.


--
-- TOC entry 3511 (class 0 OID 37834)
-- Dependencies: 238
-- Data for Name: client; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public.client (username, email, password, role, id, "createdAt", "updatedAt", "emailVerified", settings) FROM stdin;
\.


--
-- TOC entry 3541 (class 0 OID 38385)
-- Dependencies: 268
-- Data for Name: commit; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public.commit (id, summary, description, data, diff, "time", "parentId", "authorId", "eventId", "createdAt", "updatedAt", "isForkCommit") FROM stdin;
\.


--
-- TOC entry 3514 (class 0 OID 37850)
-- Dependencies: 241
-- Data for Name: contact; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public.contact (id, "profileId", type, method, status, "createdAt", "updatedAt", owner, "subscriptionId", "authId", "unsubscribeId") FROM stdin;
\.


--
-- TOC entry 3516 (class 0 OID 37859)
-- Dependencies: 243
-- Data for Name: critique; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public.critique (id, url, source, title, abstract, "time", status, "createdAt", "updatedAt", "eventId") FROM stdin;
\.


--
-- TOC entry 3517 (class 0 OID 37867)
-- Dependencies: 244
-- Data for Name: event; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public.event (name, description, status, id, "createdAt", "updatedAt", pinyin, "latestAdmittedNewsId", "ownerId", "parentId") FROM stdin;
\.


--
-- TOC entry 3538 (class 0 OID 38333)
-- Dependencies: 265
-- Data for Name: eventStackNews; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public."eventStackNews" ("eventId", "newsId", "stackId", "createdAt", "updatedAt") FROM stdin;
\.


--
-- TOC entry 3518 (class 0 OID 37873)
-- Dependencies: 245
-- Data for Name: eventTag; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public."eventTag" ("eventId", "tagId", id, "updatedAt", "createdAt") FROM stdin;
\.


--
-- TOC entry 3521 (class 0 OID 37880)
-- Dependencies: 248
-- Data for Name: headerImage; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public."headerImage" ("imageUrl", source, "sourceUrl", "eventId", id, "createdAt", "updatedAt") FROM stdin;
\.


--
-- TOC entry 3523 (class 0 OID 37894)
-- Dependencies: 250
-- Data for Name: news; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public.news (url, source, title, abstract, "time", status, comment, id, "createdAt", "updatedAt", "isInTemporaryStack") FROM stdin;
\.


--
-- TOC entry 3525 (class 0 OID 37902)
-- Dependencies: 252
-- Data for Name: notification; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public.notification ("time", mode, "eventId", status, id, "createdAt", "updatedAt") FROM stdin;
\.


--
-- TOC entry 3527 (class 0 OID 37907)
-- Dependencies: 254
-- Data for Name: record; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public.record (model, target, operation, action, data, owner, id, "createdAt", "updatedAt", before, subtarget) FROM stdin;
\.


--
-- TOC entry 3530 (class 0 OID 37917)
-- Dependencies: 257
-- Data for Name: report; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public.report (id, "time", type, method, status, "createdAt", "updatedAt", owner) FROM stdin;
\.


--
-- TOC entry 3531 (class 0 OID 37925)
-- Dependencies: 258
-- Data for Name: reportNotification; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public."reportNotification" (status, "createdAt", "updatedAt", "notificationId", "reportId") FROM stdin;
\.


--
-- TOC entry 3533 (class 0 OID 37938)
-- Dependencies: 260
-- Data for Name: stack; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public.stack (title, description, status, "eventId", "createdAt", "updatedAt", id, "order", "time") FROM stdin;
\.


--
-- TOC entry 3534 (class 0 OID 37946)
-- Dependencies: 261
-- Data for Name: subscription; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public.subscription (mode, status, "unsubscribeId", subscriber, "eventId", id, "createdAt", "updatedAt") FROM stdin;
\.


--
-- TOC entry 3536 (class 0 OID 37954)
-- Dependencies: 263
-- Data for Name: tag; Type: TABLE DATA; Schema: public; Owner: v2land
--

COPY public.tag (id, name, description, "createdAt", "updatedAt", status) FROM stdin;
\.


--
-- TOC entry 3558 (class 0 OID 0)
-- Dependencies: 231
-- Name: auth_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public.auth_id_seq', 1, true);

SELECT pg_catalog.setval('public."resourceLock_id_seq"', 1, true);


--
-- TOC entry 3559 (class 0 OID 0)
-- Dependencies: 233
-- Name: authorizationAccessToken_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public."authorizationAccessToken_id_seq"', 53, true);


--
-- TOC entry 3560 (class 0 OID 0)
-- Dependencies: 235
-- Name: authorizationClient_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public."authorizationClient_id_seq"', 4, true);


--
-- TOC entry 3561 (class 0 OID 0)
-- Dependencies: 237
-- Name: authorizationCode_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public."authorizationCode_id_seq"', 1, false);


--
-- TOC entry 3562 (class 0 OID 0)
-- Dependencies: 239
-- Name: client_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public.client_id_seq', 1, true);


--
-- TOC entry 3563 (class 0 OID 0)
-- Dependencies: 267
-- Name: commit_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public.commit_id_seq', 1, true);


--
-- TOC entry 3564 (class 0 OID 0)
-- Dependencies: 240
-- Name: contact_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public.contact_id_seq', 1, true);


--
-- TOC entry 3565 (class 0 OID 0)
-- Dependencies: 242
-- Name: critique_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public.critique_id_seq', 1, true);


--
-- TOC entry 3566 (class 0 OID 0)
-- Dependencies: 246
-- Name: eventTag_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public."eventTag_id_seq"', 1, true);


--
-- TOC entry 3567 (class 0 OID 0)
-- Dependencies: 247
-- Name: event_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public.event_id_seq', 1, true);


--
-- TOC entry 3568 (class 0 OID 0)
-- Dependencies: 249
-- Name: headerimage_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public.headerimage_id_seq', 1, true);


--
-- TOC entry 3569 (class 0 OID 0)
-- Dependencies: 251
-- Name: news_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public.news_id_seq', 1, true);


--
-- TOC entry 3570 (class 0 OID 0)
-- Dependencies: 253
-- Name: notification_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public.notification_id_seq', 1, true);


--
-- TOC entry 3571 (class 0 OID 0)
-- Dependencies: 255
-- Name: record_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public.record_id_seq', 1, true);


--
-- TOC entry 3572 (class 0 OID 0)
-- Dependencies: 256
-- Name: report_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public.report_id_seq', 1, true);


--
-- TOC entry 3573 (class 0 OID 0)
-- Dependencies: 259
-- Name: stack_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public.stack_id_seq', 1, true);


--
-- TOC entry 3574 (class 0 OID 0)
-- Dependencies: 262
-- Name: subscription_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public.subscription_id_seq', 1, true);


--
-- TOC entry 3575 (class 0 OID 0)
-- Dependencies: 264
-- Name: tag_id_seq; Type: SEQUENCE SET; Schema: public; Owner: v2land
--

SELECT pg_catalog.setval('public.tag_id_seq', 1, true);


--
-- TOC entry 3314 (class 2606 OID 38085)
-- Name: authorizationAccessToken AuthorizationAccessToken_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."authorizationAccessToken"
    ADD CONSTRAINT "AuthorizationAccessToken_pkey" PRIMARY KEY (id);


--
-- TOC entry 3312 (class 2606 OID 38087)
-- Name: auth auth_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.auth
    ADD CONSTRAINT auth_pkey PRIMARY KEY (id);


--
-- TOC entry 3316 (class 2606 OID 38089)
-- Name: authorizationClient authorizationClient_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."authorizationClient"
    ADD CONSTRAINT "authorizationClient_pkey" PRIMARY KEY (id);


--
-- TOC entry 3318 (class 2606 OID 38091)
-- Name: authorizationCode authorizationCode_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."authorizationCode"
    ADD CONSTRAINT "authorizationCode_pkey" PRIMARY KEY (id);


--
-- TOC entry 3320 (class 2606 OID 38093)
-- Name: client client_email_key; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.client
    ADD CONSTRAINT client_email_key UNIQUE (email);


--
-- TOC entry 3322 (class 2606 OID 38097)
-- Name: client client_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.client
    ADD CONSTRAINT client_pkey PRIMARY KEY (id);


--
-- TOC entry 3324 (class 2606 OID 38099)
-- Name: client client_username_key; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.client
    ADD CONSTRAINT client_username_key UNIQUE (username);


--
-- TOC entry 3358 (class 2606 OID 38393)
-- Name: commit commit_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.commit
    ADD CONSTRAINT commit_pkey PRIMARY KEY (id);


--
-- TOC entry 3326 (class 2606 OID 38101)
-- Name: contact contact_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.contact
    ADD CONSTRAINT contact_pkey PRIMARY KEY (id);


--
-- TOC entry 3328 (class 2606 OID 38103)
-- Name: critique critique_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.critique
    ADD CONSTRAINT critique_pkey PRIMARY KEY (id);


--
-- TOC entry 3356 (class 2606 OID 38337)
-- Name: eventStackNews eventStackNews_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."eventStackNews"
    ADD CONSTRAINT "eventStackNews_pkey" PRIMARY KEY ("eventId", "newsId");


--
-- TOC entry 3334 (class 2606 OID 38105)
-- Name: eventTag eventTag_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."eventTag"
    ADD CONSTRAINT "eventTag_pkey" PRIMARY KEY (id);


--
-- TOC entry 3330 (class 2606 OID 41455)
-- Name: event event_name_owner_key; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.event
    ADD CONSTRAINT event_name_owner_key UNIQUE (name, "ownerId");


--
-- TOC entry 3332 (class 2606 OID 38109)
-- Name: event event_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.event
    ADD CONSTRAINT event_pkey PRIMARY KEY (id);


--
-- TOC entry 3336 (class 2606 OID 38111)
-- Name: headerImage headerimage_event_key; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."headerImage"
    ADD CONSTRAINT headerimage_event_key UNIQUE ("eventId");


--
-- TOC entry 3338 (class 2606 OID 38113)
-- Name: headerImage headerimage_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."headerImage"
    ADD CONSTRAINT headerimage_pkey PRIMARY KEY (id);


--
-- TOC entry 3340 (class 2606 OID 38117)
-- Name: news news_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.news
    ADD CONSTRAINT news_pkey PRIMARY KEY (id);


--
-- TOC entry 3342 (class 2606 OID 38119)
-- Name: notification notification_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT notification_pkey PRIMARY KEY (id);


--
-- TOC entry 3344 (class 2606 OID 38121)
-- Name: record record_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.record
    ADD CONSTRAINT record_pkey PRIMARY KEY (id);


--
-- TOC entry 3348 (class 2606 OID 38123)
-- Name: reportNotification reportNotification_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."reportNotification"
    ADD CONSTRAINT "reportNotification_pkey" PRIMARY KEY ("notificationId", "reportId");


--
-- TOC entry 3346 (class 2606 OID 38125)
-- Name: report report_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.report
    ADD CONSTRAINT report_pkey PRIMARY KEY (id);


--
-- TOC entry 3350 (class 2606 OID 38129)
-- Name: stack stack_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.stack
    ADD CONSTRAINT stack_pkey PRIMARY KEY (id);


--
-- TOC entry 3352 (class 2606 OID 38131)
-- Name: subscription subscription_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.subscription
    ADD CONSTRAINT subscription_pkey PRIMARY KEY (id);


--
-- TOC entry 3354 (class 2606 OID 38133)
-- Name: tag tag_pkey; Type: CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.tag
    ADD CONSTRAINT tag_pkey PRIMARY KEY (id);


--
-- TOC entry 3361 (class 2606 OID 38134)
-- Name: authorizationCode authorizationClientId; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."authorizationCode"
    ADD CONSTRAINT "authorizationClientId" FOREIGN KEY ("authorizationClientId") REFERENCES public."authorizationClient"(id) NOT VALID;


--
-- TOC entry 3359 (class 2606 OID 38139)
-- Name: authorizationAccessToken authorizationClientId; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."authorizationAccessToken"
    ADD CONSTRAINT "authorizationClientId" FOREIGN KEY ("authorizationClientId") REFERENCES public."authorizationClient"(id) NOT VALID;


--
-- TOC entry 3375 (class 2606 OID 38399)
-- Name: commit commit_authorId_fkey; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.commit
    ADD CONSTRAINT "commit_authorId_fkey" FOREIGN KEY ("authorId") REFERENCES public.client(id) ON UPDATE CASCADE;


--
-- TOC entry 3376 (class 2606 OID 38404)
-- Name: commit commit_eventId_fkey; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.commit
    ADD CONSTRAINT "commit_eventId_fkey" FOREIGN KEY ("eventId") REFERENCES public.event(id) ON UPDATE CASCADE;


--
-- TOC entry 3374 (class 2606 OID 38394)
-- Name: commit commit_parentId_fkey; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.commit
    ADD CONSTRAINT "commit_parentId_fkey" FOREIGN KEY ("parentId") REFERENCES public.commit(id) ON UPDATE CASCADE;


--
-- TOC entry 3363 (class 2606 OID 38144)
-- Name: contact contact_authId_fkey; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.contact
    ADD CONSTRAINT "contact_authId_fkey" FOREIGN KEY ("authId") REFERENCES public.auth(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 3364 (class 2606 OID 38149)
-- Name: contact contact_owner_fkey; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.contact
    ADD CONSTRAINT contact_owner_fkey FOREIGN KEY (owner) REFERENCES public.client(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 3365 (class 2606 OID 38154)
-- Name: contact contact_subscriptionId_fkey; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.contact
    ADD CONSTRAINT "contact_subscriptionId_fkey" FOREIGN KEY ("subscriptionId") REFERENCES public.subscription(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 3366 (class 2606 OID 38159)
-- Name: critique critique_event_fkey; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.critique
    ADD CONSTRAINT critique_event_fkey FOREIGN KEY ("eventId") REFERENCES public.event(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 3371 (class 2606 OID 38342)
-- Name: eventStackNews eventStackNews_eventId_fkey; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."eventStackNews"
    ADD CONSTRAINT "eventStackNews_eventId_fkey" FOREIGN KEY ("eventId") REFERENCES public.event(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3372 (class 2606 OID 38347)
-- Name: eventStackNews eventStackNews_newsId_fkey; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."eventStackNews"
    ADD CONSTRAINT "eventStackNews_newsId_fkey" FOREIGN KEY ("newsId") REFERENCES public.news(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3373 (class 2606 OID 38352)
-- Name: eventStackNews eventStackNews_stackId_fkey; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."eventStackNews"
    ADD CONSTRAINT "eventStackNews_stackId_fkey" FOREIGN KEY ("stackId") REFERENCES public.stack(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3362 (class 2606 OID 38179)
-- Name: authorizationCode owner; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."authorizationCode"
    ADD CONSTRAINT owner FOREIGN KEY (owner) REFERENCES public.client(id) NOT VALID;


--
-- TOC entry 3360 (class 2606 OID 38184)
-- Name: authorizationAccessToken owner; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."authorizationAccessToken"
    ADD CONSTRAINT owner FOREIGN KEY (owner) REFERENCES public.client(id) NOT VALID;


--
-- TOC entry 3368 (class 2606 OID 38189)
-- Name: reportNotification reportNotification_notificationId_fkey; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."reportNotification"
    ADD CONSTRAINT "reportNotification_notificationId_fkey" FOREIGN KEY ("notificationId") REFERENCES public.notification(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3369 (class 2606 OID 38194)
-- Name: reportNotification reportNotification_reportId_fkey; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public."reportNotification"
    ADD CONSTRAINT "reportNotification_reportId_fkey" FOREIGN KEY ("reportId") REFERENCES public.report(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3367 (class 2606 OID 38199)
-- Name: report report_owner_fkey; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.report
    ADD CONSTRAINT report_owner_fkey FOREIGN KEY (owner) REFERENCES public.client(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 3370 (class 2606 OID 38204)
-- Name: stack stack_event_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: v2land
--

ALTER TABLE ONLY public.stack
    ADD CONSTRAINT stack_event_id_fk FOREIGN KEY ("eventId") REFERENCES public.event(id) ON DELETE CASCADE;


-- Completed on 2020-04-11 03:41:07 EDT

--
-- PostgreSQL database dump complete
--

