-- Downloaded from: https://github.com/oknosoft/windowbuilder-planning/blob/0861330ad94cd7a4755c8ba5ab798d2baf4006f6/server/keys/init/pg.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 14.8
-- Dumped by pg_dump version 14.8

-- Started on 2023-08-12 22:07:45 MSK

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = off;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET escape_string_warning = off;
SET row_security = off;

--
-- TOC entry 2 (class 3079 OID 2899788)
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- TOC entry 3404 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- TOC entry 879 (class 1247 OID 3176553)
-- Name: key_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.key_type AS ENUM (
    'order',
    'product',
    'layer',
    'profile',
    'filling',
    'glass',
    'glunit',
    'layout',
    'other'
);


--
-- TOC entry 861 (class 1247 OID 6144761)
-- Name: keys_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.keys_type AS (
	obj uuid,
	specimen integer,
	elm integer,
	region integer
);


--
-- TOC entry 885 (class 1247 OID 3698031)
-- Name: prod_row; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.prod_row AS (
	characteristic uuid,
	quantity integer
);


--
-- TOC entry 882 (class 1247 OID 3556323)
-- Name: qinfo_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.qinfo_type AS (
	abonent uuid,
	year integer,
	branch uuid,
	barcode bigint,
	ref uuid,
	calc_order uuid,
	characteristic uuid,
	presentation character varying(200),
	specimen integer,
	elm integer,
	region integer,
	type public.key_type,
	leading_product uuid
);


--
-- TOC entry 888 (class 1247 OID 5825217)
-- Name: refs; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.refs AS ENUM (
    'doc.calc_order',
    'doc.planning_event',
    'doc.work_centers_task',
    'doc.work_centers_performance',
    'doc.purchase_order',
    'doc.debit_cash_order',
    'doc.credit_cash_order',
    'doc.credit_card_order',
    'doc.debit_bank_order',
    'doc.credit_bank_order',
    'doc.selling',
    'doc.purchase',
    'doc.nom_prices_setup'
);


--
-- TOC entry 253 (class 1255 OID 3558639)
-- Name: qinfo(character varying); Type: FUNCTION; Schema: public; Owner: -
--

-- FUNCTION: public.qinfo(character varying)

-- DROP FUNCTION IF EXISTS public.qinfo(character varying);

CREATE OR REPLACE FUNCTION public.qinfo(code character varying) RETURNS qinfo_type
    LANGUAGE 'plpgsql'
    COST 100
    VOLATILE PARALLEL UNSAFE
AS $BODY$
declare
	tmp qinfo_type;
	keys_row keys%ROWTYPE;
	cx_row characteristics%ROWTYPE;
	order_row calc_orders%ROWTYPE;
	icode bigint;
	ucode uuid;
begin
  /* ищем запись в keys */
  if(char_length(code) = 13) then
	code = substring(code, 1, 12);
  end if;
  if char_length(code) = 12 then
    icode = code;
	SELECT * INTO keys_row FROM keys WHERE barcode=icode;
  elseif char_length(code) = 36 then
    ucode = code;
    SELECT * INTO keys_row FROM keys WHERE ref=ucode;
  end if;

  /* подклеиваем заказ и прочую инфу */
  if keys_row.type is null then
	RAISE NOTICE 'null';
  elseif keys_row.type = 'order' then
	SELECT * INTO order_row FROM calc_orders WHERE ref=keys_row.obj;
  else
	SELECT * INTO cx_row FROM characteristics WHERE ref=keys_row.obj;
	SELECT * INTO order_row FROM calc_orders WHERE ref=cx_row.calc_order;
  end if;
  tmp.abonent = order_row.abonent;
  tmp.year = order_row.year;
  tmp.branch = order_row.branch;
  tmp.calc_order = order_row.ref;

  tmp.characteristic = cx_row.ref;
  tmp.leading_product = cx_row.leading_product;

  tmp.barcode = keys_row.barcode;
  tmp.ref = keys_row.ref;
  tmp.specimen = keys_row.specimen;
  tmp.elm = keys_row.elm;
  tmp.region = keys_row.region;
  tmp.type = keys_row.type;

  if keys_row.type = 'order' then
  	tmp.presentation = format('%s от %s', order_row.number_doc, order_row.date);
  else
  	tmp.presentation = cx_row.name;
  end if;
  return tmp;
end
$BODY$;

ALTER FUNCTION public.qinfo(character varying)
    OWNER TO postgres;



SET default_table_access_method = heap;

--
-- TOC entry 231 (class 1259 OID 6296209)
-- Name: areg_needs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.areg_needs (
    register uuid NOT NULL,
    register_type public.refs NOT NULL,
    row_num bigint NOT NULL,
    period timestamp without time zone,
    sign smallint DEFAULT 1,
    calc_order uuid,
    nom uuid,
    characteristic uuid,
    stage uuid,
    planing_key uuid,
    quantity numeric(15,3) DEFAULT 0
);


--
-- TOC entry 3405 (class 0 OID 0)
-- Dependencies: 231
-- Name: COLUMN areg_needs.register; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.areg_needs.register IS 'Регистратор';


--
-- TOC entry 3406 (class 0 OID 0)
-- Dependencies: 231
-- Name: COLUMN areg_needs.register_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.areg_needs.register_type IS 'Тип регистратора';


--
-- TOC entry 3407 (class 0 OID 0)
-- Dependencies: 231
-- Name: COLUMN areg_needs.row_num; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.areg_needs.row_num IS 'Номер строки';


--
-- TOC entry 3408 (class 0 OID 0)
-- Dependencies: 231
-- Name: COLUMN areg_needs.period; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.areg_needs.period IS 'Период';


--
-- TOC entry 3409 (class 0 OID 0)
-- Dependencies: 231
-- Name: COLUMN areg_needs.nom; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.areg_needs.nom IS 'Номенклатура';


--
-- TOC entry 3410 (class 0 OID 0)
-- Dependencies: 231
-- Name: COLUMN areg_needs.characteristic; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.areg_needs.characteristic IS 'Характеристика';


--
-- TOC entry 3411 (class 0 OID 0)
-- Dependencies: 231
-- Name: COLUMN areg_needs.stage; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.areg_needs.stage IS 'Этап производства';


--
-- TOC entry 3412 (class 0 OID 0)
-- Dependencies: 231
-- Name: COLUMN areg_needs.planing_key; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.areg_needs.planing_key IS 'Ключ планирования';


--
-- TOC entry 3413 (class 0 OID 0)
-- Dependencies: 231
-- Name: COLUMN areg_needs.quantity; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.areg_needs.quantity IS 'Количество';


--
-- TOC entry 227 (class 1259 OID 3061481)
-- Name: calc_orders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.calc_orders (
    ref uuid NOT NULL,
    abonent uuid,
    branch uuid,
    year integer,
    date timestamp without time zone,
    number_doc character(11),
    partner uuid,
    organization uuid,
    author uuid,
    department uuid,
    production json
);


--
-- TOC entry 225 (class 1259 OID 2900139)
-- Name: characteristics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.characteristics (
    ref uuid NOT NULL,
    calc_order uuid,
    leading_product uuid,
    name character varying(200)
);


--
-- TOC entry 224 (class 1259 OID 2899806)
-- Name: keys; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.keys (
    ref uuid DEFAULT public.uuid_generate_v1mc() NOT NULL,
    obj uuid NOT NULL,
    specimen integer DEFAULT 1,
    elm integer DEFAULT 0,
    region integer DEFAULT 0,
    barcode bigint DEFAULT 0,
    type public.key_type
);


--
-- TOC entry 226 (class 1259 OID 2900714)
-- Name: settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.settings (
    param character varying(100) NOT NULL,
    value json NOT NULL
);


--
-- TOC entry 3247 (class 1259 OID 3061357)
-- Name: address; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX address ON public.keys USING btree (obj, specimen, elm, region);


--
-- TOC entry 3258 (class 2606 OID 6296217)
-- Name: areg_needs areg_needs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.areg_needs
    ADD CONSTRAINT areg_needs_pkey PRIMARY KEY (register, register_type, row_num);


--
-- TOC entry 3252 (class 2606 OID 2900143)
-- Name: characteristics characteristics_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.characteristics
    ADD CONSTRAINT characteristics_pkey PRIMARY KEY (ref);


--
-- TOC entry 3250 (class 2606 OID 2899811)
-- Name: keys keys_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.keys
    ADD CONSTRAINT keys_pkey PRIMARY KEY (ref);


--
-- TOC entry 3256 (class 2606 OID 3061485)
-- Name: calc_orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.calc_orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (ref);


--
-- TOC entry 3254 (class 2606 OID 2900720)
-- Name: settings settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.settings
    ADD CONSTRAINT settings_pkey PRIMARY KEY (param);


--
-- TOC entry 3248 (class 1259 OID 3494689)
-- Name: barcode; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX barcode ON public.keys USING btree (barcode);


--
-- TOC entry 3259 (class 2606 OID 3061494)
-- Name: characteristics order; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.characteristics
    ADD CONSTRAINT "order" FOREIGN KEY (calc_order) REFERENCES public.calc_orders(ref) NOT VALID;


-- Completed on 2023-08-12 22:07:45 MSK

--
-- PostgreSQL database dump complete
--

