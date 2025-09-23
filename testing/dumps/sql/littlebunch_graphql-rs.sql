-- Downloaded from: https://github.com/littlebunch/graphql-rs/blob/f5169da015094669fd8058c11706bbe68b56c18b/database/pg/up.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 13.0
-- Dumped by pg_dump version 13.0

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
-- Name: diesel_manage_updated_at(regclass); Type: FUNCTION; Schema: public; Owner: gmoore
--

CREATE FUNCTION public.diesel_manage_updated_at(_tbl regclass) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    EXECUTE format('CREATE TRIGGER set_updated_at BEFORE UPDATE ON %s
                    FOR EACH ROW EXECUTE PROCEDURE diesel_set_updated_at()', _tbl);
END;
$$;


ALTER FUNCTION public.diesel_manage_updated_at(_tbl regclass) OWNER TO gmoore;

--
-- Name: diesel_set_updated_at(); Type: FUNCTION; Schema: public; Owner: gmoore
--

CREATE FUNCTION public.diesel_set_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF (
        NEW IS DISTINCT FROM OLD AND
        NEW.updated_at IS NOT DISTINCT FROM OLD.updated_at
    ) THEN
        NEW.updated_at := current_timestamp;
    END IF;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.diesel_set_updated_at() OWNER TO gmoore;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: __diesel_schema_migrations; Type: TABLE; Schema: public; Owner: gmoore
--

CREATE TABLE public.__diesel_schema_migrations (
    version character varying(50) NOT NULL,
    run_on timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.__diesel_schema_migrations OWNER TO gmoore;

--
-- Name: derivations; Type: TABLE; Schema: public; Owner: gmoore
--

CREATE TABLE public.derivations (
    id integer NOT NULL,
    code character varying(255) NOT NULL,
    description text NOT NULL
);


ALTER TABLE public.derivations OWNER TO gmoore;

--
-- Name: derivations_id_seq; Type: SEQUENCE; Schema: public; Owner: gmoore
--

CREATE SEQUENCE public.derivations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.derivations_id_seq OWNER TO gmoore;

--
-- Name: derivations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: gmoore
--

ALTER SEQUENCE public.derivations_id_seq OWNED BY public.derivations.id;


--
-- Name: food_groups; Type: TABLE; Schema: public; Owner: gmoore
--

CREATE TABLE public.food_groups (
    id integer NOT NULL,
    description character varying(255) DEFAULT ''::character varying NOT NULL
);


ALTER TABLE public.food_groups OWNER TO gmoore;

--
-- Name: foods; Type: TABLE; Schema: public; Owner: gmoore
--

CREATE TABLE public.foods (
    id integer NOT NULL,
    publication_date timestamp with time zone NOT NULL,
    modified_date timestamp with time zone NOT NULL,
    available_date timestamp with time zone NOT NULL,
    upc character varying(24) NOT NULL,
    fdc_id character varying(24) NOT NULL,
    description character varying(255) NOT NULL,
    food_group_id integer DEFAULT 0 NOT NULL,
    manufacturer_id integer DEFAULT 0 NOT NULL,
    datasource character varying(8) NOT NULL,
    serving_size double precision,
    serving_unit character varying(24) DEFAULT NULL::character varying,
    serving_description character varying(256) DEFAULT NULL::character varying,
    country character varying(24) DEFAULT NULL::character varying,
    ingredients text,
    kw_tsvector tsvector GENERATED ALWAYS AS (to_tsvector('english'::regconfig, (((COALESCE(description, ''::character varying))::text || ' '::text) || COALESCE(ingredients, ''::text)))) STORED
);


ALTER TABLE public.foods OWNER TO gmoore;

--
-- Name: manufacturers; Type: TABLE; Schema: public; Owner: gmoore
--

CREATE TABLE public.manufacturers (
    id integer NOT NULL,
    name character varying(255) DEFAULT ''::character varying NOT NULL
);


ALTER TABLE public.manufacturers OWNER TO gmoore;

--
-- Name: food_groups_id_seq; Type: SEQUENCE; Schema: public; Owner: gmoore
--

CREATE SEQUENCE public.food_groups_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.food_groups_id_seq OWNER TO gmoore;

--
-- Name: food_groups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: gmoore
--

ALTER SEQUENCE public.food_groups_id_seq OWNED BY public.food_groups.id;


--
-- Name: foods_id_seq; Type: SEQUENCE; Schema: public; Owner: gmoore
--

CREATE SEQUENCE public.foods_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.foods_id_seq OWNER TO gmoore;

--
-- Name: foods_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: gmoore
--

ALTER SEQUENCE public.foods_id_seq OWNED BY public.foods.id;


--
-- Name: manufacturers_id_seq; Type: SEQUENCE; Schema: public; Owner: gmoore
--

CREATE SEQUENCE public.manufacturers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.manufacturers_id_seq OWNER TO gmoore;

--
-- Name: manufacturers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: gmoore
--

ALTER SEQUENCE public.manufacturers_id_seq OWNED BY public.manufacturers.id;


--
-- Name: nutrient_data; Type: TABLE; Schema: public; Owner: gmoore
--

CREATE TABLE public.nutrient_data (
    id integer NOT NULL,
    value double precision DEFAULT '0'::double precision NOT NULL,
    standard_error double precision,
    minimum double precision,
    maximum double precision,
    median double precision,
    derivation_id integer DEFAULT 0 NOT NULL,
    nutrient_id integer DEFAULT 0 NOT NULL,
    food_id integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.nutrient_data OWNER TO gmoore;

--
-- Name: nutrient_data_id_seq; Type: SEQUENCE; Schema: public; Owner: gmoore
--

CREATE SEQUENCE public.nutrient_data_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nutrient_data_id_seq OWNER TO gmoore;

--
-- Name: nutrient_data_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: gmoore
--

ALTER SEQUENCE public.nutrient_data_id_seq OWNED BY public.nutrient_data.id;


--
-- Name: nutrients; Type: TABLE; Schema: public; Owner: gmoore
--

CREATE TABLE public.nutrients (
    id integer NOT NULL,
    nutrientno character varying(12) NOT NULL,
    description character varying(255) NOT NULL,
    unit character varying(24) NOT NULL
);


ALTER TABLE public.nutrients OWNER TO gmoore;

--
-- Name: derivations id; Type: DEFAULT; Schema: public; Owner: gmoore
--

ALTER TABLE ONLY public.derivations ALTER COLUMN id SET DEFAULT nextval('public.derivations_id_seq'::regclass);


--
-- Name: food_groups id; Type: DEFAULT; Schema: public; Owner: gmoore
--

ALTER TABLE ONLY public.food_groups ALTER COLUMN id SET DEFAULT nextval('public.food_groups_id_seq'::regclass);


--
-- Name: foods id; Type: DEFAULT; Schema: public; Owner: gmoore
--

ALTER TABLE ONLY public.foods ALTER COLUMN id SET DEFAULT nextval('public.foods_id_seq'::regclass);


--
-- Name: manufacturers id; Type: DEFAULT; Schema: public; Owner: gmoore
--

ALTER TABLE ONLY public.manufacturers ALTER COLUMN id SET DEFAULT nextval('public.manufacturers_id_seq'::regclass);


--
-- Name: nutrient_data id; Type: DEFAULT; Schema: public; Owner: gmoore
--

ALTER TABLE ONLY public.nutrient_data ALTER COLUMN id SET DEFAULT nextval('public.nutrient_data_id_seq'::regclass);


--
-- Name: __diesel_schema_migrations __diesel_schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: gmoore
--

ALTER TABLE ONLY public.__diesel_schema_migrations
    ADD CONSTRAINT __diesel_schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: derivations derivations_pkey; Type: CONSTRAINT; Schema: public; Owner: gmoore
--

ALTER TABLE ONLY public.derivations
    ADD CONSTRAINT derivations_pkey PRIMARY KEY (id);


--
-- Name: food_groups food_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: gmoore
--

ALTER TABLE ONLY public.food_groups
    ADD CONSTRAINT food_groups_pkey PRIMARY KEY (id);


--
-- Name: foods foods_pkey; Type: CONSTRAINT; Schema: public; Owner: gmoore
--

ALTER TABLE ONLY public.foods
    ADD CONSTRAINT foods_pkey PRIMARY KEY (id);


--
-- Name: manufacturers manufacturers_pkey; Type: CONSTRAINT; Schema: public; Owner: gmoore
--

ALTER TABLE ONLY public.manufacturers
    ADD CONSTRAINT manufacturers_pkey PRIMARY KEY (id);


--
-- Name: nutrient_data nutrient_data_pkey; Type: CONSTRAINT; Schema: public; Owner: gmoore
--

ALTER TABLE ONLY public.nutrient_data
    ADD CONSTRAINT nutrient_data_pkey PRIMARY KEY (id);


--
-- Name: nutrients nutrients_pkey; Type: CONSTRAINT; Schema: public; Owner: gmoore
--

ALTER TABLE ONLY public.nutrients
    ADD CONSTRAINT nutrients_pkey PRIMARY KEY (id);


--
-- Name: idx_16458_foods_description_idx; Type: INDEX; Schema: public; Owner: gmoore
--

CREATE INDEX idx_16458_foods_description_idx ON public.foods USING btree (description);


--
-- Name: idx_16458_foods_fdc_id_idx; Type: INDEX; Schema: public; Owner: gmoore
--

CREATE INDEX idx_16458_foods_fdc_id_idx ON public.foods USING btree (fdc_id);


--
-- Name: idx_16458_foods_fk; Type: INDEX; Schema: public; Owner: gmoore
--

CREATE INDEX idx_16458_foods_fk ON public.foods USING btree (manufacturer_id);


--
-- Name: idx_16458_foods_food_group_id_idx; Type: INDEX; Schema: public; Owner: gmoore
--

CREATE INDEX idx_16458_foods_food_group_id_idx ON public.foods USING btree (food_group_id);


--
-- Name: idx_16458_foods_manufacturer_id_idx; Type: INDEX; Schema: public; Owner: gmoore
--

CREATE INDEX idx_16458_foods_manufacturer_id_idx ON public.foods USING btree (manufacturer_id);


--
-- Name: idx_16458_foods_upc_idx; Type: INDEX; Schema: public; Owner: gmoore
--

CREATE INDEX idx_16458_foods_upc_idx ON public.foods USING btree (upc);


--
-- Name: idx_16472_food_groups_description_idx; Type: INDEX; Schema: public; Owner: gmoore
--

CREATE INDEX idx_16472_food_groups_description_idx ON public.food_groups USING btree (description);


--
-- Name: idx_16479_manufacturers_name_idx; Type: INDEX; Schema: public; Owner: gmoore
--

CREATE INDEX idx_16479_manufacturers_name_idx ON public.manufacturers USING btree (name);


--
-- Name: idx_16484_nutrientno; Type: INDEX; Schema: public; Owner: gmoore
--

CREATE UNIQUE INDEX idx_16484_nutrientno ON public.nutrients USING btree (nutrientno);


--
-- Name: idx_16489_nutrient_data_fk; Type: INDEX; Schema: public; Owner: gmoore
--

CREATE INDEX idx_16489_nutrient_data_fk ON public.nutrient_data USING btree (nutrient_id);


--
-- Name: idx_16489_nutrient_data_fk_1; Type: INDEX; Schema: public; Owner: gmoore
--

CREATE INDEX idx_16489_nutrient_data_fk_1 ON public.nutrient_data USING btree (derivation_id);


--
-- Name: idx_16489_nutrient_data_food_id_idx; Type: INDEX; Schema: public; Owner: gmoore
--

CREATE INDEX idx_16489_nutrient_data_food_id_idx ON public.nutrient_data USING btree (food_id);


--
-- Name: kw_tsvector_idx; Type: INDEX; Schema: public; Owner: gmoore
--

CREATE INDEX kw_tsvector_idx ON public.foods USING gin (kw_tsvector);


--
-- Name: foods foods_fk; Type: FK CONSTRAINT; Schema: public; Owner: gmoore
--

ALTER TABLE ONLY public.foods
    ADD CONSTRAINT foods_fk FOREIGN KEY (manufacturer_id) REFERENCES public.manufacturers(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: foods foods_fk_1; Type: FK CONSTRAINT; Schema: public; Owner: gmoore
--

ALTER TABLE ONLY public.foods
    ADD CONSTRAINT foods_fk_1 FOREIGN KEY (food_group_id) REFERENCES public.food_groups(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: nutrient_data nutrient_data_fk; Type: FK CONSTRAINT; Schema: public; Owner: gmoore
--

ALTER TABLE ONLY public.nutrient_data
    ADD CONSTRAINT nutrient_data_fk FOREIGN KEY (nutrient_id) REFERENCES public.nutrients(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: nutrient_data nutrient_data_fk_1; Type: FK CONSTRAINT; Schema: public; Owner: gmoore
--

ALTER TABLE ONLY public.nutrient_data
    ADD CONSTRAINT nutrient_data_fk_1 FOREIGN KEY (derivation_id) REFERENCES public.derivations(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: nutrient_data nutrient_data_food_fk; Type: FK CONSTRAINT; Schema: public; Owner: gmoore
--

ALTER TABLE ONLY public.nutrient_data
    ADD CONSTRAINT nutrient_data_food_fk FOREIGN KEY (food_id) REFERENCES public.foods(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- PostgreSQL database dump complete
--

