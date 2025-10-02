-- Downloaded from: https://github.com/AlexTransit/venderctl/blob/5a4426d96e78edbf76b8157e42af2508dc7449bd/sql/db.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 11.15 (Debian 11.15-1.pgdg110+1)
-- Dumped by pg_dump version 11.1

-- Started on 2022-04-27 18:41:37

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

ALTER TABLE IF EXISTS ONLY public.trans DROP CONSTRAINT IF EXISTS trans_tax_job_id_fkey;
DROP TRIGGER IF EXISTS trans_tax ON public.trans;
DROP TRIGGER IF EXISTS tax_job_modified ON public.tax_job;
DROP TRIGGER IF EXISTS tax_job_maint_before ON public.tax_job;
DROP TRIGGER IF EXISTS tax_job_maint_after ON public.tax_job;
DROP INDEX IF EXISTS public.trans_executer;
DROP INDEX IF EXISTS public.tgchat_idx2;
DROP INDEX IF EXISTS public.tgchat_idx1;
DROP INDEX IF EXISTS public.tgchat_idx;
DROP INDEX IF EXISTS public.idx_trans_vmtime;
DROP INDEX IF EXISTS public.idx_trans_vmid_vmtime;
DROP INDEX IF EXISTS public.idx_tax_job_sched;
DROP INDEX IF EXISTS public.idx_tax_job_help;
DROP INDEX IF EXISTS public.idx_state_vmid_state_received;
DROP INDEX IF EXISTS public.idx_inventory_vmid_service;
DROP INDEX IF EXISTS public.idx_inventory_vmid_not_service;
DROP INDEX IF EXISTS public.idx_ingest_received;
DROP INDEX IF EXISTS public.idx_error_vmid_vmtime_code;
DROP INDEX IF EXISTS public.idx_catalog_vmid_code_name;
DROP INDEX IF EXISTS public.cashless_vmid_payment_id_order_id_key;
DROP INDEX IF EXISTS public.cashless_idx;
ALTER TABLE IF EXISTS ONLY public.tg_user DROP CONSTRAINT IF EXISTS tg_user_pkey;
ALTER TABLE IF EXISTS ONLY public.tax_job DROP CONSTRAINT IF EXISTS tax_job_pkey;
ALTER TABLE IF EXISTS ONLY public.state DROP CONSTRAINT IF EXISTS state_vmid_key;
ALTER TABLE IF EXISTS ONLY public.robot DROP CONSTRAINT IF EXISTS robot_serial_num_key;
ALTER TABLE IF EXISTS ONLY public.robot DROP CONSTRAINT IF EXISTS "robot-key";
ALTER TABLE IF EXISTS public.tax_job ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.tg_user_user_id_seq;
DROP TABLE IF EXISTS public.tg_user;
DROP TABLE IF EXISTS public.tg_chat;
DROP SEQUENCE IF EXISTS public.tax_job_id_seq;
DROP VIEW IF EXISTS public.tax_job_help;
DROP TABLE IF EXISTS public.state;
DROP TABLE IF EXISTS public.robot;
DROP TABLE IF EXISTS public.old_state;
DROP TABLE IF EXISTS public.inventory;
DROP TABLE IF EXISTS public.ingest;
DROP TABLE IF EXISTS public.error;
DROP TABLE IF EXISTS public.catalog;
DROP TABLE IF EXISTS public.cashless;
DROP FUNCTION IF EXISTS public.vmstate(s integer);
DROP FUNCTION IF EXISTS public.trans_tax_trigger();
DROP FUNCTION IF EXISTS public.tax_job_trans(t public.trans);
DROP TABLE IF EXISTS public.trans;
DROP FUNCTION IF EXISTS public.tax_job_take(arg_worker text);
DROP TABLE IF EXISTS public.tax_job;
DROP FUNCTION IF EXISTS public.tax_job_modified();
DROP FUNCTION IF EXISTS public.tax_job_maint_before();
DROP FUNCTION IF EXISTS public.tax_job_maint_after();
DROP FUNCTION IF EXISTS public.state_update(arg_vmid integer, arg_state integer);
DROP FUNCTION IF EXISTS public.connect_update(arg_vmid integer, arg_connect boolean);
DROP TYPE IF EXISTS public.tax_job_state;
DROP TYPE IF EXISTS public.cashless_state;
DROP EXTENSION IF EXISTS hstore;
--
-- TOC entry 2 (class 3079 OID 24642)
-- Name: hstore; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS hstore WITH SCHEMA public;


--
-- TOC entry 3092 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION hstore; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION hstore IS 'data type for storing sets of (key, value) pairs';


--
-- TOC entry 692 (class 1247 OID 65591)
-- Name: cashless_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.cashless_state AS ENUM (
    'order_start',
    'order_prepay',
    'order_complete',
    'order_cancel'
);


--
-- TOC entry 707 (class 1247 OID 26134)
-- Name: tax_job_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.tax_job_state AS ENUM (
    'sched',
    'busy',
    'final',
    'help'
);


--
-- TOC entry 294 (class 1255 OID 55071)
-- Name: connect_update(integer, boolean); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.connect_update(arg_vmid integer, arg_connect boolean) RETURNS integer
    LANGUAGE plpgsql
    AS '
BEGIN
	INSERT INTO state (vmid, state, received, connected, contime)  VALUES (arg_vmid, 0, CURRENT_TIMESTAMP, arg_connect, CURRENT_TIMESTAMP)
    ON CONFLICT (vmid) DO UPDATE 
    SET connected = excluded.connected, contime = CURRENT_TIMESTAMP;
    return null;
END;
';


--
-- TOC entry 290 (class 1255 OID 25797)
-- Name: state_update(integer, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.state_update(arg_vmid integer, arg_state integer) RETURNS integer
    LANGUAGE plpgsql
    AS '
DECLARE
    old_state int4 = NULL;
BEGIN
    SELECT
        state INTO old_state
    FROM
        state
    WHERE
        vmid = arg_vmid
    LIMIT 1
    FOR UPDATE;
    INSERT INTO state (vmid, state, received)
        VALUES (arg_vmid, arg_state, CURRENT_TIMESTAMP)
    ON CONFLICT (vmid)
        DO UPDATE SET
            state = excluded.state, received = excluded.received;
    RETURN old_state;
END;
';


--
-- TOC entry 291 (class 1255 OID 26171)
-- Name: tax_job_maint_after(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.tax_job_maint_after() RETURNS trigger
    LANGUAGE plpgsql
    AS '
BEGIN
    CASE new.state
    WHEN ''final'' THEN
        NOTIFY tax_job_final;
    WHEN ''help'' THEN
        NOTIFY tax_job_help;
    WHEN ''sched'' THEN
        NOTIFY tax_job_sched;
    ELSE
        NULL;
    END CASE;
    RETURN NEW;
END;
';


--
-- TOC entry 292 (class 1255 OID 26173)
-- Name: tax_job_maint_before(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.tax_job_maint_before() RETURNS trigger
    LANGUAGE plpgsql
    AS '
BEGIN
    IF new.state = ''final'' THEN
        new.scheduled = NULL;
    END IF;
    RETURN NEW;
END;
';


--
-- TOC entry 293 (class 1255 OID 26175)
-- Name: tax_job_modified(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.tax_job_modified() RETURNS trigger
    LANGUAGE plpgsql
    AS '
BEGIN
    new.modified := CURRENT_TIMESTAMP;
    RETURN NEW;
END;
';


SET default_tablespace = '';

SET default_with_oids = false;

--
-- TOC entry 209 (class 1259 OID 26219)
-- Name: tax_job; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tax_job (
    id bigint NOT NULL,
    state public.tax_job_state NOT NULL,
    created timestamp with time zone NOT NULL,
    modified timestamp with time zone NOT NULL,
    scheduled timestamp with time zone,
    worker text,
    processor text,
    ext_id text,
    data jsonb,
    gross integer,
    notes text[],
    ops jsonb,
    CONSTRAINT tax_job_check CHECK ((NOT ((state = 'sched'::public.tax_job_state) AND (scheduled IS NULL)))),
    CONSTRAINT tax_job_check1 CHECK ((NOT ((state = 'busy'::public.tax_job_state) AND (worker IS NULL))))
);


--
-- TOC entry 295 (class 1255 OID 26249)
-- Name: tax_job_take(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.tax_job_take(arg_worker text) RETURNS SETOF public.tax_job
    LANGUAGE sql
    AS '
    UPDATE
        tax_job
    SET
        state = ''busy'',
        worker = arg_worker
    WHERE
        state = ''sched''
        AND scheduled <= CURRENT_TIMESTAMP
        AND id = (
            SELECT
                id
            FROM
                tax_job
            WHERE
                state = ''sched''
                AND scheduled <= CURRENT_TIMESTAMP
            ORDER BY
                scheduled,
                modified
            LIMIT 1
            FOR UPDATE
                SKIP LOCKED)
    RETURNING
        *;

';


--
-- TOC entry 210 (class 1259 OID 26232)
-- Name: trans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.trans (
    vmid integer NOT NULL,
    vmtime timestamp with time zone,
    received timestamp with time zone NOT NULL,
    menu_code text NOT NULL,
    options integer[],
    price integer NOT NULL,
    method integer NOT NULL,
    tax_job_id bigint,
    executer bigint,
    exeputer_type integer,
    executer_str text
);


--
-- TOC entry 296 (class 1255 OID 26250)
-- Name: tax_job_trans(public.trans); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.tax_job_trans(t public.trans) RETURNS public.tax_job
    LANGUAGE plpgsql
    AS '
    # print_strict_params ON
DECLARE
    tjd jsonb;
    ops jsonb;
    tj tax_job;
    name text;
BEGIN
    -- lock trans row
    PERFORM
        1
    FROM
        trans
    WHERE (vmid, vmtime) = (t.vmid,
        t.vmtime)
LIMIT 1
FOR UPDATE;
    -- if trans already has tax_job assigned, just return it
    IF t.tax_job_id IS NOT NULL THEN
        SELECT
            * INTO STRICT tj
        FROM
            tax_job
        WHERE
            id = t.tax_job_id;
        RETURN tj;
    END IF;
    -- op code to human friendly name via catalog
    SELECT
        catalog.name INTO name
    FROM
        catalog
    WHERE (vmid, code) = (t.vmid,
        t.menu_code);
    IF NOT found THEN
        name := ''#'' || t.menu_code;
    END IF;
    ops := jsonb_build_array (jsonb_build_object(''vmid'', t.vmid, ''time'', t.vmtime, ''name'', name, ''code'', t.menu_code, ''amount'', 1, ''price'', t.price, ''method'', t.method));
    INSERT INTO tax_job (state, created, modified, scheduled, processor, ops, gross)
        VALUES (''sched'', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ''ru2019'', ops, t.price)
    RETURNING
        * INTO STRICT tj;
    UPDATE
        trans
    SET
        tax_job_id = tj.id
    WHERE (vmid, vmtime) = (t.vmid,
        t.vmtime);
    RETURN tj;
END;
';


--
-- TOC entry 289 (class 1255 OID 26177)
-- Name: trans_tax_trigger(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.trans_tax_trigger() RETURNS trigger
    LANGUAGE plpgsql
    AS '
BEGIN
	IF (NEW.vmid = (SELECT vmid from robot where robot.vmid = NEW.vmid and robot.work = TRUE) and (NEW.method = 1 or NEW.method = 2)) THEN
	    PERFORM
       	tax_job_trans (new);
    END IF;
    RETURN new;
END;
';


--
-- TOC entry 297 (class 1255 OID 26492)
-- Name: vmstate(integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.vmstate(s integer) RETURNS text
    LANGUAGE sql IMMUTABLE STRICT
    AS '
    -- TODO generate from tele.proto
    -- Invalid = 0;
    -- Boot = 1;
    -- Nominal = 2;
    -- Disconnected = 3;
    -- Problem = 4;
    -- Service = 5;
    -- Lock = 6;
    SELECT
        CASE WHEN s = 0 THEN
            ''Invalid''
        WHEN s = 1 THEN
            ''Boot''
        WHEN s = 2 THEN
            ''Nominal''
        WHEN s = 3 THEN
            ''Disconnected''
        WHEN s = 4 THEN
            ''Problem''
        WHEN s = 5 THEN
            ''Service''
        WHEN s = 6 THEN
            ''Lock''
        ELSE
            ''unknown:'' || s
        END
';


--
-- TOC entry 219 (class 1259 OID 65639)
-- Name: cashless; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cashless (
    state public.cashless_state DEFAULT 'order_start'::public.cashless_state NOT NULL,
    vmid integer NOT NULL,
    create_date timestamp with time zone NOT NULL,
    credit_date timestamp with time zone,
    finish_date timestamp with time zone,
    payment_id character varying(20) NOT NULL,
    order_id character varying NOT NULL,
    amount integer NOT NULL,
    credited integer DEFAULT 0 NOT NULL,
    bank_commission integer DEFAULT 0 NOT NULL,
    terminal text
);
ALTER TABLE ONLY public.cashless ALTER COLUMN credit_date SET STATISTICS 0;
ALTER TABLE ONLY public.cashless ALTER COLUMN payment_id SET STATISTICS 0;
ALTER TABLE ONLY public.cashless ALTER COLUMN credited SET STATISTICS 0;


--
-- TOC entry 212 (class 1259 OID 26503)
-- Name: catalog; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.catalog (
    vmid integer NOT NULL,
    code text NOT NULL,
    name text NOT NULL
);


--
-- TOC entry 206 (class 1259 OID 25437)
-- Name: error; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.error (
    vmid integer NOT NULL,
    vmtime timestamp with time zone NOT NULL,
    received timestamp with time zone NOT NULL,
    code integer,
    message text NOT NULL,
    count integer,
    app_version text
);


--
-- TOC entry 205 (class 1259 OID 25417)
-- Name: ingest; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ingest (
    received timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    vmid integer NOT NULL,
    done boolean DEFAULT false NOT NULL,
    raw bytea NOT NULL
);


--
-- TOC entry 207 (class 1259 OID 25482)
-- Name: inventory; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.inventory (
    vmid integer NOT NULL,
    at_service boolean NOT NULL,
    vmtime timestamp with time zone NOT NULL,
    received timestamp with time zone NOT NULL,
    inventory public.hstore,
    cashbox_bill public.hstore,
    cashbox_coin public.hstore,
    change_bill public.hstore,
    change_coin public.hstore
);


--
-- TOC entry 214 (class 1259 OID 55050)
-- Name: old_state; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.old_state (
    state integer
);


--
-- TOC entry 213 (class 1259 OID 26578)
-- Name: robot; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.robot (
    vmid integer NOT NULL,
    vmnum integer NOT NULL,
    description text,
    location text,
    bunkers public.hstore,
    "mobile-number" numeric(10,0),
    serial_num numeric(7,0) NOT NULL,
    work boolean DEFAULT true NOT NULL,
    in_robo public.hstore,
    to_robo public.hstore
);


--
-- TOC entry 3093 (class 0 OID 0)
-- Dependencies: 213
-- Name: COLUMN robot.in_robo; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.robot.in_robo IS 'inventoiry inside robo';


--
-- TOC entry 215 (class 1259 OID 55059)
-- Name: state; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.state (
    vmid integer NOT NULL,
    state integer NOT NULL,
    received timestamp with time zone NOT NULL,
    connected boolean DEFAULT false NOT NULL,
    contime timestamp with time zone DEFAULT now() NOT NULL
);


--
-- TOC entry 211 (class 1259 OID 26245)
-- Name: tax_job_help; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.tax_job_help AS
 SELECT tax_job.id,
    tax_job.state,
    tax_job.created,
    tax_job.modified,
    tax_job.scheduled,
    tax_job.worker,
    tax_job.processor,
    tax_job.ext_id,
    tax_job.data,
    tax_job.gross,
    tax_job.notes
   FROM public.tax_job
  WHERE (tax_job.state = 'help'::public.tax_job_state)
  ORDER BY tax_job.modified;


--
-- TOC entry 208 (class 1259 OID 26217)
-- Name: tax_job_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tax_job_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 3094 (class 0 OID 0)
-- Dependencies: 208
-- Name: tax_job_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tax_job_id_seq OWNED BY public.tax_job.id;


--
-- TOC entry 218 (class 1259 OID 65014)
-- Name: tg_chat; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tg_chat (
    create_date timestamp(0) without time zone DEFAULT now() NOT NULL,
    messageid integer NOT NULL,
    fromid bigint NOT NULL,
    toid bigint NOT NULL,
    date integer NOT NULL,
    text text,
    changedate integer,
    changetext text
);
ALTER TABLE ONLY public.tg_chat ALTER COLUMN messageid SET STATISTICS 0;
ALTER TABLE ONLY public.tg_chat ALTER COLUMN fromid SET STATISTICS 0;
ALTER TABLE ONLY public.tg_chat ALTER COLUMN toid SET STATISTICS 0;
ALTER TABLE ONLY public.tg_chat ALTER COLUMN text SET STATISTICS 0;


--
-- TOC entry 217 (class 1259 OID 64971)
-- Name: tg_user; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tg_user (
    ban boolean DEFAULT false,
    userid bigint NOT NULL,
    name text,
    firstname text,
    lastname text,
    phonenumber text,
    balance integer,
    credit integer,
    registerdate integer,
    diskont integer DEFAULT 3
);
ALTER TABLE ONLY public.tg_user ALTER COLUMN name SET STATISTICS 0;
ALTER TABLE ONLY public.tg_user ALTER COLUMN phonenumber SET STATISTICS 0;


--
-- TOC entry 216 (class 1259 OID 64969)
-- Name: tg_user_user_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tg_user_user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 3095 (class 0 OID 0)
-- Dependencies: 216
-- Name: tg_user_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tg_user_user_id_seq OWNED BY public.tg_user.userid;


--
-- TOC entry 2922 (class 2604 OID 26222)
-- Name: tax_job id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tax_job ALTER COLUMN id SET DEFAULT nextval('public.tax_job_id_seq'::regclass);


--
-- TOC entry 2947 (class 2606 OID 26586)
-- Name: robot robot-key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.robot
    ADD CONSTRAINT "robot-key" PRIMARY KEY (vmid);


--
-- TOC entry 2949 (class 2606 OID 26588)
-- Name: robot robot_serial_num_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.robot
    ADD CONSTRAINT robot_serial_num_key UNIQUE (serial_num);


--
-- TOC entry 2952 (class 2606 OID 55075)
-- Name: state state_vmid_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.state
    ADD CONSTRAINT state_vmid_key UNIQUE (vmid);


--
-- TOC entry 2941 (class 2606 OID 26229)
-- Name: tax_job tax_job_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tax_job
    ADD CONSTRAINT tax_job_pkey PRIMARY KEY (id);


--
-- TOC entry 2954 (class 2606 OID 64984)
-- Name: tg_user tg_user_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tg_user
    ADD CONSTRAINT tg_user_pkey PRIMARY KEY (userid);


--
-- TOC entry 2958 (class 1259 OID 65648)
-- Name: cashless_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX cashless_idx ON public.cashless USING btree (payment_id, order_id);


--
-- TOC entry 2959 (class 1259 OID 65649)
-- Name: cashless_vmid_payment_id_order_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX cashless_vmid_payment_id_order_id_key ON public.cashless USING btree (vmid, payment_id, order_id);


--
-- TOC entry 2945 (class 1259 OID 26509)
-- Name: idx_catalog_vmid_code_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_catalog_vmid_code_name ON public.catalog USING btree (vmid, code, name);


--
-- TOC entry 2935 (class 1259 OID 26132)
-- Name: idx_error_vmid_vmtime_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_error_vmid_vmtime_code ON public.error USING btree (vmid, vmtime DESC) INCLUDE (code);


--
-- TOC entry 2934 (class 1259 OID 26128)
-- Name: idx_ingest_received; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ingest_received ON public.ingest USING btree (received) WHERE (NOT done);


--
-- TOC entry 2936 (class 1259 OID 26131)
-- Name: idx_inventory_vmid_not_service; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_inventory_vmid_not_service ON public.inventory USING btree (vmid) WITH (fillfactor='10') WHERE (NOT at_service);


--
-- TOC entry 2937 (class 1259 OID 26130)
-- Name: idx_inventory_vmid_service; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_inventory_vmid_service ON public.inventory USING btree (vmid) WITH (fillfactor='10') WHERE at_service;


--
-- TOC entry 2950 (class 1259 OID 55062)
-- Name: idx_state_vmid_state_received; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_state_vmid_state_received ON public.state USING btree (vmid, state, received) WITH (fillfactor='10');


--
-- TOC entry 2938 (class 1259 OID 26231)
-- Name: idx_tax_job_help; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tax_job_help ON public.tax_job USING btree (modified) WHERE (state = 'help'::public.tax_job_state);


--
-- TOC entry 2939 (class 1259 OID 26230)
-- Name: idx_tax_job_sched; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tax_job_sched ON public.tax_job USING btree (scheduled, modified) WHERE (state = 'sched'::public.tax_job_state);


--
-- TOC entry 2942 (class 1259 OID 26244)
-- Name: idx_trans_vmid_vmtime; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_trans_vmid_vmtime ON public.trans USING btree (vmid, vmtime);


--
-- TOC entry 2943 (class 1259 OID 26243)
-- Name: idx_trans_vmtime; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_trans_vmtime ON public.trans USING btree (vmtime);


--
-- TOC entry 2955 (class 1259 OID 65021)
-- Name: tgchat_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX tgchat_idx ON public.tg_chat USING btree (messageid, fromid, toid, date);


--
-- TOC entry 2956 (class 1259 OID 65022)
-- Name: tgchat_idx1; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX tgchat_idx1 ON public.tg_chat USING btree (fromid);


--
-- TOC entry 2957 (class 1259 OID 65023)
-- Name: tgchat_idx2; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX tgchat_idx2 ON public.tg_chat USING btree (toid);


--
-- TOC entry 2944 (class 1259 OID 64901)
-- Name: trans_executer; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX trans_executer ON public.trans USING btree (executer);


--
-- TOC entry 2961 (class 2620 OID 26251)
-- Name: tax_job tax_job_maint_after; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER tax_job_maint_after AFTER INSERT OR UPDATE ON public.tax_job FOR EACH ROW EXECUTE PROCEDURE public.tax_job_maint_after();


--
-- TOC entry 2962 (class 2620 OID 26252)
-- Name: tax_job tax_job_maint_before; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER tax_job_maint_before BEFORE INSERT OR UPDATE ON public.tax_job FOR EACH ROW EXECUTE PROCEDURE public.tax_job_maint_before();


--
-- TOC entry 2963 (class 2620 OID 26253)
-- Name: tax_job tax_job_modified; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER tax_job_modified BEFORE UPDATE ON public.tax_job FOR EACH ROW WHEN (((new.ext_id IS DISTINCT FROM old.ext_id) OR (new.data IS DISTINCT FROM old.data) OR (new.notes IS DISTINCT FROM old.notes))) EXECUTE PROCEDURE public.tax_job_modified();


--
-- TOC entry 2964 (class 2620 OID 26254)
-- Name: trans trans_tax; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trans_tax AFTER INSERT ON public.trans FOR EACH ROW EXECUTE PROCEDURE public.trans_tax_trigger();


--
-- TOC entry 2960 (class 2606 OID 26238)
-- Name: trans trans_tax_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trans
    ADD CONSTRAINT trans_tax_job_id_fkey FOREIGN KEY (tax_job_id) REFERENCES public.tax_job(id) ON UPDATE RESTRICT ON DELETE SET NULL;


-- Completed on 2022-04-27 18:41:38

--
-- PostgreSQL database dump complete
--

