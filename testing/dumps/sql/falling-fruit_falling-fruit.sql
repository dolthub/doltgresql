-- Downloaded from: https://github.com/falling-fruit/falling-fruit/blob/29caf62b55d2367fca5cb966647aea925f656339/db/schema.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 14.7 (Homebrew)
-- Dumped by pg_dump version 14.7 (Homebrew)

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
-- Name: intarray; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS intarray WITH SCHEMA public;


--
-- Name: EXTENSION intarray; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION intarray IS 'functions, operators, and index support for 1-D arrays of integers';


--
-- Name: postgis; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS postgis WITH SCHEMA public;


--
-- Name: EXTENSION postgis; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION postgis IS 'PostGIS geometry and geography spatial types and functions';


--
-- Name: add_observation_photo(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.add_observation_photo() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
      BEGIN
        INSERT INTO photos (
          observation_id,
          user_id,
          observation_order,
          thumb,
          medium,
          original
        )
        VALUES (
          NEW.id,
          NEW.user_id,
          1,
          'https://ff-production.s3.us-west-2.amazonaws.com/observations/photos/' || substring(lpad(NEW.id::text, 9, '0') from 1 for 3) || '/' || substring(lpad(NEW.id::text, 9, '0') from 4 for 3) || '/' || substring(lpad(NEW.id::text, 9, '0') from 7 for 3) || '/thumb/' || NEW.photo_file_name,
          'https://ff-production.s3.us-west-2.amazonaws.com/observations/photos/' || substring(lpad(NEW.id::text, 9, '0') from 1 for 3) || '/' || substring(lpad(NEW.id::text, 9, '0') from 4 for 3) || '/' || substring(lpad(NEW.id::text, 9, '0') from 7 for 3) || '/medium/' || NEW.photo_file_name,
          'https://ff-production.s3.us-west-2.amazonaws.com/observations/photos/' || substring(lpad(NEW.id::text, 9, '0') from 1 for 3) || '/' || substring(lpad(NEW.id::text, 9, '0') from 4 for 3) || '/' || substring(lpad(NEW.id::text, 9, '0') from 7 for 3) || '/original/' || NEW.photo_file_name
        );
        RETURN NEW;
      END;
      $$;


--
-- Name: st_buffer_meters(public.geometry, double precision); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.st_buffer_meters(public.geometry, double precision) RETURNS public.geometry
    LANGUAGE plpgsql IMMUTABLE
    AS $_$
DECLARE
orig_srid int;
utm_srid int;

BEGIN
orig_srid:= ST_SRID($1);
utm_srid:= utmzone(ST_Centroid($1));

RETURN ST_transform(ST_Buffer(ST_transform($1, utm_srid), $2), orig_srid);
END;
$_$;


--
-- Name: utmzone(public.geometry); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.utmzone(public.geometry) RETURNS integer
    LANGUAGE plpgsql IMMUTABLE
    AS $_$
DECLARE
geomgeog geometry;
zone int;
pref int;

BEGIN
geomgeog:= ST_Transform($1,4326);

IF (ST_Y(geomgeog))>0 THEN
pref:=32600;
ELSE
pref:=32700;
END IF;

zone:=floor((ST_X(geomgeog)+180)/6)+1;
RETURN zone+pref;
END;
$_$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: api_keys; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.api_keys (
    id integer NOT NULL,
    api_key character varying(255),
    version integer DEFAULT 0 NOT NULL,
    api_type character varying(255),
    name character varying(255),
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: api_keys_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.api_keys_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: api_keys_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.api_keys_id_seq OWNED BY public.api_keys.id;


--
-- Name: api_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.api_logs (
    id integer NOT NULL,
    n integer,
    endpoint character varying(255),
    request_method character varying(255),
    params text,
    ip_address character varying(255),
    api_key character varying(255),
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: api_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.api_logs_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: api_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.api_logs_id_seq OWNED BY public.api_logs.id;


--
-- Name: changes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.changes (
    id integer NOT NULL,
    location_id integer,
    remote_ip character varying(255),
    description text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    user_id integer,
    observation_id integer,
    author character varying(255),
    description_patch text,
    former_type_ids integer[] DEFAULT '{}'::integer[],
    former_type_others character varying(255)[] DEFAULT '{}'::character varying[],
    former_location public.geography(Point,4326),
    spam boolean DEFAULT false,
    location json,
    review json
);


--
-- Name: changes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.changes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: changes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.changes_id_seq OWNED BY public.changes.id;


--
-- Name: clusters; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.clusters (
    id integer NOT NULL,
    geohash text NOT NULL,
    muni boolean NOT NULL,
    x real NOT NULL,
    y real NOT NULL,
    count integer NOT NULL,
    zoom integer NOT NULL,
    type_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: clusters_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.clusters_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: clusters_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.clusters_id_seq OWNED BY public.clusters.id;


--
-- Name: imports; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.imports (
    id integer NOT NULL,
    url character varying(255),
    name character varying(255) NOT NULL,
    comments text,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    autoload boolean DEFAULT true NOT NULL,
    muni boolean DEFAULT false,
    license text,
    auto_cluster boolean DEFAULT false,
    reverse_geocode boolean DEFAULT false,
    default_category_mask integer DEFAULT 0
);


--
-- Name: imports_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.imports_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: imports_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.imports_id_seq OWNED BY public.imports.id;


--
-- Name: invasives; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invasives (
    id integer NOT NULL,
    regions public.geography(MultiPolygon,4326),
    type_id integer,
    source character varying(255)
);


--
-- Name: invasives_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.invasives_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: invasives_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.invasives_id_seq OWNED BY public.invasives.id;


--
-- Name: locations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.locations (
    id integer NOT NULL,
    lat double precision,
    lng double precision,
    author character varying(255),
    description text,
    season_start integer,
    season_stop integer,
    no_season boolean,
    address text,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    unverified boolean DEFAULT false,
    access integer,
    import_id integer,
    location public.geography(Point,4326),
    client character varying(255) DEFAULT 'web'::character varying,
    city character varying(255),
    state character varying(255),
    country character varying(255),
    user_id integer,
    type_ids integer[],
    muni boolean DEFAULT false,
    original_ids character varying[],
    invasive boolean DEFAULT false,
    inaturalist_id integer,
    hidden boolean DEFAULT false
);


--
-- Name: locations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.locations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: locations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.locations_id_seq OWNED BY public.locations.id;


--
-- Name: locations_routes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.locations_routes (
    id integer NOT NULL,
    location_id integer,
    route_id integer,
    "position" integer,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: locations_routes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.locations_routes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: locations_routes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.locations_routes_id_seq OWNED BY public.locations_routes.id;


--
-- Name: observations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.observations (
    id integer NOT NULL,
    location_id integer,
    comment text,
    observed_on date,
    photo_file_name character varying(255),
    photo_content_type character varying(255),
    photo_file_size integer,
    photo_updated_at timestamp without time zone,
    fruiting integer,
    quality_rating integer,
    yield_rating integer,
    user_id integer,
    remote_ip character varying(255),
    author character varying(255),
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    photo_caption text,
    graft boolean DEFAULT false
);


--
-- Name: observations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.observations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: observations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.observations_id_seq OWNED BY public.observations.id;


--
-- Name: photos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.photos (
    id integer NOT NULL,
    observation_id integer,
    user_id integer,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    thumb text NOT NULL,
    medium text NOT NULL,
    original text NOT NULL,
    observation_order integer
);


--
-- Name: photos_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.photos_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: photos_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.photos_id_seq OWNED BY public.photos.id;


--
-- Name: problems; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.problems (
    id integer NOT NULL,
    problem_code integer,
    comment text,
    resolution_code integer,
    response text,
    reporter_id integer,
    responder_id integer,
    email character varying(255) NOT NULL,
    name character varying(255),
    location_id integer,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: problems_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.problems_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: problems_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.problems_id_seq OWNED BY public.problems.id;


--
-- Name: refresh_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.refresh_tokens (
    id integer NOT NULL,
    user_id integer NOT NULL,
    jti text NOT NULL,
    exp integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: refresh_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.refresh_tokens_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: refresh_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.refresh_tokens_id_seq OWNED BY public.refresh_tokens.id;


--
-- Name: routes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.routes (
    id integer NOT NULL,
    name character varying(255),
    user_id integer,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    is_public boolean DEFAULT true NOT NULL,
    access_key character varying(255),
    transport_type integer DEFAULT 0
);


--
-- Name: routes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.routes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: routes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.routes_id_seq OWNED BY public.routes.id;


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying(255) NOT NULL
);


--
-- Name: sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sessions (
    id integer NOT NULL,
    session_id character varying(255) NOT NULL,
    data text,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sessions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sessions_id_seq OWNED BY public.sessions.id;


--
-- Name: types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.types (
    id integer NOT NULL,
    en_name character varying(255),
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    scientific_name character varying(255),
    usda_symbol character varying(255),
    wikipedia_url character varying(255),
    edibility character varying(255),
    notes text,
    en_synonyms character varying(255),
    scientific_synonyms character varying(255),
    urban_mushrooms_url character varying(255),
    fruitipedia_url character varying(255),
    eat_the_weeds_url character varying(255),
    foraging_texas_url character varying(255),
    parent_id integer,
    taxonomic_rank integer,
    es_name character varying(255),
    he_name character varying(255),
    pl_name character varying(255),
    category_mask integer DEFAULT 1,
    fr_name character varying(255),
    de_name character varying(255),
    pending boolean DEFAULT true,
    it_name character varying(255) DEFAULT NULL::character varying,
    el_name character varying(255) DEFAULT NULL::character varying,
    sv_name character varying(255) DEFAULT NULL::character varying,
    tr_name character varying(255) DEFAULT NULL::character varying,
    nl_name character varying(255) DEFAULT NULL::character varying,
    ar_name character varying(255) DEFAULT NULL::character varying,
    sk_name character varying(255) DEFAULT NULL::character varying,
    vi_name text,
    pt_name text,
    zh_hant_name text,
    zh_hans_name text,
    uk_name text
);


--
-- Name: types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.types_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.types_id_seq OWNED BY public.types.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id integer NOT NULL,
    email character varying(255) DEFAULT ''::character varying NOT NULL,
    encrypted_password character varying(255) DEFAULT ''::character varying NOT NULL,
    reset_password_token character varying(255),
    reset_password_sent_at timestamp without time zone,
    remember_created_at timestamp without time zone,
    sign_in_count integer DEFAULT 0,
    current_sign_in_at timestamp without time zone,
    last_sign_in_at timestamp without time zone,
    current_sign_in_ip character varying(255),
    last_sign_in_ip character varying(255),
    confirmation_token character varying(255),
    confirmed_at timestamp without time zone,
    confirmation_sent_at timestamp without time zone,
    unconfirmed_email character varying(255),
    authentication_token character varying(255),
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    range public.geography(Polygon,4326),
    name character varying(255),
    bio text,
    roles_mask integer DEFAULT 10 NOT NULL,
    range_updates_email boolean DEFAULT false NOT NULL,
    add_anonymously boolean DEFAULT false NOT NULL,
    announcements_email boolean DEFAULT true,
    address text,
    lat numeric,
    lng numeric,
    range_radius numeric,
    range_radius_unit character varying(255),
    location public.geography(Point,4326),
    roles text[] GENERATED ALWAYS AS (
CASE
    WHEN ((roles_mask & ('0001'::"bit")::integer) > 0) THEN ARRAY['user'::text, 'admin'::text]
    ELSE ARRAY['user'::text]
END) STORED NOT NULL
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: api_keys id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_keys ALTER COLUMN id SET DEFAULT nextval('public.api_keys_id_seq'::regclass);


--
-- Name: api_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_logs ALTER COLUMN id SET DEFAULT nextval('public.api_logs_id_seq'::regclass);


--
-- Name: changes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changes ALTER COLUMN id SET DEFAULT nextval('public.changes_id_seq'::regclass);


--
-- Name: clusters id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clusters ALTER COLUMN id SET DEFAULT nextval('public.clusters_id_seq'::regclass);


--
-- Name: imports id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.imports ALTER COLUMN id SET DEFAULT nextval('public.imports_id_seq'::regclass);


--
-- Name: invasives id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invasives ALTER COLUMN id SET DEFAULT nextval('public.invasives_id_seq'::regclass);


--
-- Name: locations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locations ALTER COLUMN id SET DEFAULT nextval('public.locations_id_seq'::regclass);


--
-- Name: locations_routes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locations_routes ALTER COLUMN id SET DEFAULT nextval('public.locations_routes_id_seq'::regclass);


--
-- Name: observations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.observations ALTER COLUMN id SET DEFAULT nextval('public.observations_id_seq'::regclass);


--
-- Name: photos id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.photos ALTER COLUMN id SET DEFAULT nextval('public.photos_id_seq'::regclass);


--
-- Name: problems id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.problems ALTER COLUMN id SET DEFAULT nextval('public.problems_id_seq'::regclass);


--
-- Name: refresh_tokens id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.refresh_tokens ALTER COLUMN id SET DEFAULT nextval('public.refresh_tokens_id_seq'::regclass);


--
-- Name: routes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.routes ALTER COLUMN id SET DEFAULT nextval('public.routes_id_seq'::regclass);


--
-- Name: sessions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions ALTER COLUMN id SET DEFAULT nextval('public.sessions_id_seq'::regclass);


--
-- Name: types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.types ALTER COLUMN id SET DEFAULT nextval('public.types_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: api_keys api_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);


--
-- Name: api_logs api_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_logs
    ADD CONSTRAINT api_logs_pkey PRIMARY KEY (id);


--
-- Name: changes changes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changes
    ADD CONSTRAINT changes_pkey PRIMARY KEY (id);


--
-- Name: clusters clusters_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clusters
    ADD CONSTRAINT clusters_pkey PRIMARY KEY (id);


--
-- Name: imports imports_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.imports
    ADD CONSTRAINT imports_pkey PRIMARY KEY (id);


--
-- Name: invasives invasives_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invasives
    ADD CONSTRAINT invasives_pkey PRIMARY KEY (id);


--
-- Name: locations locations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locations
    ADD CONSTRAINT locations_pkey PRIMARY KEY (id);


--
-- Name: locations_routes locations_routes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locations_routes
    ADD CONSTRAINT locations_routes_pkey PRIMARY KEY (id);


--
-- Name: observations observations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.observations
    ADD CONSTRAINT observations_pkey PRIMARY KEY (id);


--
-- Name: photos photos_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.photos
    ADD CONSTRAINT photos_pkey PRIMARY KEY (id);


--
-- Name: problems problems_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.problems
    ADD CONSTRAINT problems_pkey PRIMARY KEY (id);


--
-- Name: refresh_tokens refresh_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_pkey PRIMARY KEY (id);


--
-- Name: routes routes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.routes
    ADD CONSTRAINT routes_pkey PRIMARY KEY (id);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: types types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.types
    ADD CONSTRAINT types_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: changes_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX changes_created_at_idx ON public.changes USING btree (created_at DESC);


--
-- Name: changes_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX changes_user_id_idx ON public.changes USING btree (user_id);


--
-- Name: index_changes_on_former_location; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_changes_on_former_location ON public.changes USING gist (former_location);


--
-- Name: index_clusters_on_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_clusters_on_type_id ON public.clusters USING btree (type_id);


--
-- Name: index_invasives_on_regions; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_invasives_on_regions ON public.invasives USING gist (regions);


--
-- Name: index_invasives_on_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_invasives_on_type_id ON public.invasives USING btree (type_id);


--
-- Name: index_locations_on_hidden; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_locations_on_hidden ON public.locations USING btree (hidden);


--
-- Name: index_locations_on_location; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_locations_on_location ON public.locations USING gist (location);


--
-- Name: index_locations_on_muni_updated_lng_lat; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_locations_on_muni_updated_lng_lat ON public.locations USING btree (muni, updated_at DESC, lng, lat) WHERE (NOT hidden);


--
-- Name: index_locations_routes_on_location_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_locations_routes_on_location_id ON public.locations_routes USING btree (location_id);


--
-- Name: index_locations_routes_on_route_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_locations_routes_on_route_id ON public.locations_routes USING btree (route_id);


--
-- Name: index_problems_on_location_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_problems_on_location_id ON public.problems USING btree (location_id);


--
-- Name: index_problems_on_reporter_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_problems_on_reporter_id ON public.problems USING btree (reporter_id);


--
-- Name: index_problems_on_responder_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_problems_on_responder_id ON public.problems USING btree (responder_id);


--
-- Name: index_routes_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_routes_on_user_id ON public.routes USING btree (user_id);


--
-- Name: index_sessions_on_session_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_sessions_on_session_id ON public.sessions USING btree (session_id);


--
-- Name: index_sessions_on_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_sessions_on_updated_at ON public.sessions USING btree (updated_at);


--
-- Name: index_users_on_authentication_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_users_on_authentication_token ON public.users USING btree (authentication_token);


--
-- Name: index_users_on_confirmation_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_users_on_confirmation_token ON public.users USING btree (confirmation_token);


--
-- Name: index_users_on_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_users_on_email ON public.users USING btree (email);


--
-- Name: index_users_on_location; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_location ON public.users USING gist (location);


--
-- Name: index_users_on_range; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_range ON public.users USING gist (range);


--
-- Name: index_users_on_reset_password_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_users_on_reset_password_token ON public.users USING btree (reset_password_token);


--
-- Name: locations_import_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX locations_import_idx ON public.locations USING btree (import_id);


--
-- Name: locations_muni_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX locations_muni_idx ON public.locations USING btree (muni);


--
-- Name: locations_type_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX locations_type_idx ON public.locations USING btree (type_ids);


--
-- Name: locations_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX locations_user_id_idx ON public.locations USING btree (user_id);


--
-- Name: observations_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX observations_user_id_idx ON public.observations USING btree (user_id);


--
-- Name: photos_observation_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX photos_observation_id_idx ON public.photos USING btree (observation_id);


--
-- Name: photos_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX photos_user_id_idx ON public.photos USING btree (user_id);


--
-- Name: refresh_tokens_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX refresh_tokens_user_id_idx ON public.refresh_tokens USING btree (user_id);


--
-- Name: unique_schema_migrations; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX unique_schema_migrations ON public.schema_migrations USING btree (version);


--
-- Name: observations add_observation_photo_trigger; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER add_observation_photo_trigger AFTER INSERT ON public.observations FOR EACH ROW WHEN ((new.photo_file_name IS NOT NULL)) EXECUTE FUNCTION public.add_observation_photo();


--
-- Name: changes changes_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changes
    ADD CONSTRAINT changes_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.locations(id) ON DELETE CASCADE;


--
-- Name: changes changes_observation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changes
    ADD CONSTRAINT changes_observation_id_fkey FOREIGN KEY (observation_id) REFERENCES public.observations(id) ON DELETE CASCADE;


--
-- Name: changes changes_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.changes
    ADD CONSTRAINT changes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: locations locations_import_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locations
    ADD CONSTRAINT locations_import_id_fkey FOREIGN KEY (import_id) REFERENCES public.imports(id) ON DELETE SET NULL;


--
-- Name: locations locations_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locations
    ADD CONSTRAINT locations_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: observations observations_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.observations
    ADD CONSTRAINT observations_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.locations(id) ON DELETE CASCADE;


--
-- Name: observations observations_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.observations
    ADD CONSTRAINT observations_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: photos photos_observation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.photos
    ADD CONSTRAINT photos_observation_id_fkey FOREIGN KEY (observation_id) REFERENCES public.observations(id) ON DELETE CASCADE;


--
-- Name: photos photos_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.photos
    ADD CONSTRAINT photos_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: problems problems_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.problems
    ADD CONSTRAINT problems_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.locations(id) ON DELETE SET NULL;


--
-- Name: problems problems_reporter_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.problems
    ADD CONSTRAINT problems_reporter_id_fkey FOREIGN KEY (reporter_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: problems problems_responder_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.problems
    ADD CONSTRAINT problems_responder_id_fkey FOREIGN KEY (responder_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: refresh_tokens refresh_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: types types_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.types
    ADD CONSTRAINT types_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.types(id) ON DELETE SET NULL;


--
-- PostgreSQL database dump complete
--

--
-- PostgreSQL database dump
--

-- Dumped from database version 14.7 (Homebrew)
-- Dumped by pg_dump version 14.7 (Homebrew)

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
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.schema_migrations (version) FROM stdin;
20121024002119
20121024002244
20121024002531
20130125221414
20130125221849
20130125231339
20130126194149
20130126195410
20130130011504
20130202175557
20130202180415
20130202204513
20130205200822
20130205232437
20130213164353
20130217163517
20130220005519
20130226175338
20130428192429
20130428222654
20130428234854
20130503190834
20130503191902
20130509020018
20130531015908
20130603222010
20130730182103
20130809160242
20130811164904
20130829171108
20130829171415
20130910180201
20130910215544
20130913193714
20131109161337
20131110212841
20131110213005
20131111171522
20131111173614
20131124204626
20131124212425
20131206141457
20140210231416
20140317143110
20140318211550
20140321171442
20140325153906
20140327194704
20140407232224
20140417205031
20140513173219
20140804172708
20140924145737
20140925170731
20140925191944
20140925201738
20141009034457
20141226194649
20150227163444
20150303154024
20150402221746
20150418004920
20150425195210
20150426233816
20150723232010
20150823011715
20150914031653
20150920184333
20160411162746
20160425203759
20160501155720
20160906192251
20170113003103
20170828190311
20170828190638
20170904062642
20170904112847
20170904113628
20170904114744
20190816051810
20190816052605
20190816053525
20191224111156
20200512162439
20201230162701
20230305115612
20230305201225
20230305201521
20230305202032
20230305202357
20230305203436
20230305203802
20230305204429
20230305204742
20230305205110
20230305205220
20230305205334
20230305210534
20230305210917
20240229142208
20240703090432
20241211202028
20250410072058
20250410072400
20250410073355
20250410073503
\.


--
-- PostgreSQL database dump complete
--

