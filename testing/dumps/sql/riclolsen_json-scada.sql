-- Downloaded from: https://github.com/riclolsen/json-scada/blob/03a159749f400528d9a66e3864c00caa13b0c881/sql/grafanaappdb.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 14.1
-- Dumped by pg_dump version 14.1

-- Started on 2024-05-15 16:02:51

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
-- TOC entry 3 (class 2615 OID 2200)
-- Name: public; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA IF NOT EXISTS public;


ALTER SCHEMA public OWNER TO postgres;

--
-- TOC entry 4083 (class 0 OID 0)
-- Dependencies: 3
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: postgres
--

COMMENT ON SCHEMA public IS 'standard public schema';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 245 (class 1259 OID 20526)
-- Name: alert; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.alert (
    id integer NOT NULL,
    version bigint NOT NULL,
    dashboard_id bigint NOT NULL,
    panel_id bigint NOT NULL,
    org_id bigint NOT NULL,
    name character varying(255) NOT NULL,
    message text NOT NULL,
    state character varying(190) NOT NULL,
    settings text NOT NULL,
    frequency bigint NOT NULL,
    handler bigint NOT NULL,
    severity text NOT NULL,
    silenced boolean NOT NULL,
    execution_error text NOT NULL,
    eval_data text,
    eval_date timestamp without time zone,
    new_state_date timestamp without time zone NOT NULL,
    state_changes integer NOT NULL,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    "for" bigint
);


ALTER TABLE public.alert OWNER TO postgres;

--
-- TOC entry 285 (class 1259 OID 20847)
-- Name: alert_configuration; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.alert_configuration (
    id integer NOT NULL,
    alertmanager_configuration text NOT NULL,
    configuration_version character varying(3) NOT NULL,
    created_at integer NOT NULL,
    "default" boolean DEFAULT false NOT NULL,
    org_id bigint DEFAULT 0 NOT NULL,
    configuration_hash character varying(32) DEFAULT 'not-yet-calculated'::character varying NOT NULL
);


ALTER TABLE public.alert_configuration OWNER TO postgres;

--
-- TOC entry 293 (class 1259 OID 20891)
-- Name: alert_configuration_history; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.alert_configuration_history (
    id integer NOT NULL,
    org_id bigint DEFAULT 0 NOT NULL,
    alertmanager_configuration text NOT NULL,
    configuration_hash character varying(32) DEFAULT 'not-yet-calculated'::character varying NOT NULL,
    configuration_version character varying(3) NOT NULL,
    created_at integer NOT NULL,
    "default" boolean DEFAULT false NOT NULL,
    last_applied integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.alert_configuration_history OWNER TO postgres;

--
-- TOC entry 292 (class 1259 OID 20890)
-- Name: alert_configuration_history_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.alert_configuration_history_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.alert_configuration_history_id_seq OWNER TO postgres;

--
-- TOC entry 4084 (class 0 OID 0)
-- Dependencies: 292
-- Name: alert_configuration_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.alert_configuration_history_id_seq OWNED BY public.alert_configuration_history.id;


--
-- TOC entry 284 (class 1259 OID 20846)
-- Name: alert_configuration_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.alert_configuration_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.alert_configuration_id_seq OWNER TO postgres;

--
-- TOC entry 4085 (class 0 OID 0)
-- Dependencies: 284
-- Name: alert_configuration_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.alert_configuration_id_seq OWNED BY public.alert_configuration.id;


--
-- TOC entry 244 (class 1259 OID 20525)
-- Name: alert_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.alert_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.alert_id_seq OWNER TO postgres;

--
-- TOC entry 4086 (class 0 OID 0)
-- Dependencies: 244
-- Name: alert_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.alert_id_seq OWNED BY public.alert.id;


--
-- TOC entry 291 (class 1259 OID 20881)
-- Name: alert_image; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.alert_image (
    id integer NOT NULL,
    token character varying(190) NOT NULL,
    path character varying(190) NOT NULL,
    url character varying(2048) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    expires_at timestamp without time zone NOT NULL
);


ALTER TABLE public.alert_image OWNER TO postgres;

--
-- TOC entry 290 (class 1259 OID 20880)
-- Name: alert_image_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.alert_image_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.alert_image_id_seq OWNER TO postgres;

--
-- TOC entry 4087 (class 0 OID 0)
-- Dependencies: 290
-- Name: alert_image_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.alert_image_id_seq OWNED BY public.alert_image.id;


--
-- TOC entry 279 (class 1259 OID 20792)
-- Name: alert_instance; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.alert_instance (
    rule_org_id bigint NOT NULL,
    rule_uid character varying(40) DEFAULT 0 NOT NULL,
    labels text NOT NULL,
    labels_hash character varying(190) NOT NULL,
    current_state character varying(190) NOT NULL,
    current_state_since bigint NOT NULL,
    last_eval_time bigint NOT NULL,
    current_state_end bigint DEFAULT 0 NOT NULL,
    current_reason character varying(190)
);


ALTER TABLE public.alert_instance OWNER TO postgres;

--
-- TOC entry 249 (class 1259 OID 20550)
-- Name: alert_notification; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.alert_notification (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    name character varying(190) NOT NULL,
    type character varying(255) NOT NULL,
    settings text NOT NULL,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    is_default boolean DEFAULT false NOT NULL,
    frequency bigint,
    send_reminder boolean DEFAULT false,
    disable_resolve_message boolean DEFAULT false NOT NULL,
    uid character varying(40),
    secure_settings text
);


ALTER TABLE public.alert_notification OWNER TO postgres;

--
-- TOC entry 248 (class 1259 OID 20549)
-- Name: alert_notification_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.alert_notification_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.alert_notification_id_seq OWNER TO postgres;

--
-- TOC entry 4088 (class 0 OID 0)
-- Dependencies: 248
-- Name: alert_notification_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.alert_notification_id_seq OWNED BY public.alert_notification.id;


--
-- TOC entry 251 (class 1259 OID 20573)
-- Name: alert_notification_state; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.alert_notification_state (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    alert_id bigint NOT NULL,
    notifier_id bigint NOT NULL,
    state character varying(50) NOT NULL,
    version bigint NOT NULL,
    updated_at bigint NOT NULL,
    alert_rule_state_updated_version bigint NOT NULL
);


ALTER TABLE public.alert_notification_state OWNER TO postgres;

--
-- TOC entry 250 (class 1259 OID 20572)
-- Name: alert_notification_state_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.alert_notification_state_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.alert_notification_state_id_seq OWNER TO postgres;

--
-- TOC entry 4089 (class 0 OID 0)
-- Dependencies: 250
-- Name: alert_notification_state_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.alert_notification_state_id_seq OWNED BY public.alert_notification_state.id;


--
-- TOC entry 281 (class 1259 OID 20806)
-- Name: alert_rule; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.alert_rule (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    title character varying(190) NOT NULL,
    condition character varying(190) NOT NULL,
    data text NOT NULL,
    updated timestamp without time zone NOT NULL,
    interval_seconds bigint DEFAULT 60 NOT NULL,
    version integer DEFAULT 0 NOT NULL,
    uid character varying(40) DEFAULT 0 NOT NULL,
    namespace_uid character varying(40) NOT NULL,
    rule_group character varying(190) NOT NULL,
    no_data_state character varying(15) DEFAULT 'NoData'::character varying NOT NULL,
    exec_err_state character varying(15) DEFAULT 'Alerting'::character varying NOT NULL,
    "for" bigint DEFAULT 0 NOT NULL,
    annotations text,
    labels text,
    dashboard_uid character varying(40),
    panel_id bigint,
    rule_group_idx integer DEFAULT 1 NOT NULL,
    is_paused boolean DEFAULT false NOT NULL
);


ALTER TABLE public.alert_rule OWNER TO postgres;

--
-- TOC entry 280 (class 1259 OID 20805)
-- Name: alert_rule_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.alert_rule_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.alert_rule_id_seq OWNER TO postgres;

--
-- TOC entry 4090 (class 0 OID 0)
-- Dependencies: 280
-- Name: alert_rule_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.alert_rule_id_seq OWNED BY public.alert_rule.id;


--
-- TOC entry 247 (class 1259 OID 20542)
-- Name: alert_rule_tag; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.alert_rule_tag (
    id integer NOT NULL,
    alert_id bigint NOT NULL,
    tag_id bigint NOT NULL
);


ALTER TABLE public.alert_rule_tag OWNER TO postgres;

--
-- TOC entry 246 (class 1259 OID 20541)
-- Name: alert_rule_tag_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.alert_rule_tag_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.alert_rule_tag_id_seq OWNER TO postgres;

--
-- TOC entry 4091 (class 0 OID 0)
-- Dependencies: 246
-- Name: alert_rule_tag_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.alert_rule_tag_id_seq OWNED BY public.alert_rule_tag.id;


--
-- TOC entry 283 (class 1259 OID 20829)
-- Name: alert_rule_version; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.alert_rule_version (
    id integer NOT NULL,
    rule_org_id bigint NOT NULL,
    rule_uid character varying(40) DEFAULT 0 NOT NULL,
    rule_namespace_uid character varying(40) NOT NULL,
    rule_group character varying(190) NOT NULL,
    parent_version integer NOT NULL,
    restored_from integer NOT NULL,
    version integer NOT NULL,
    created timestamp without time zone NOT NULL,
    title character varying(190) NOT NULL,
    condition character varying(190) NOT NULL,
    data text NOT NULL,
    interval_seconds bigint NOT NULL,
    no_data_state character varying(15) DEFAULT 'NoData'::character varying NOT NULL,
    exec_err_state character varying(15) DEFAULT 'Alerting'::character varying NOT NULL,
    "for" bigint DEFAULT 0 NOT NULL,
    annotations text,
    labels text,
    rule_group_idx integer DEFAULT 1 NOT NULL,
    is_paused boolean DEFAULT false NOT NULL
);


ALTER TABLE public.alert_rule_version OWNER TO postgres;

--
-- TOC entry 282 (class 1259 OID 20828)
-- Name: alert_rule_version_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.alert_rule_version_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.alert_rule_version_id_seq OWNER TO postgres;

--
-- TOC entry 4092 (class 0 OID 0)
-- Dependencies: 282
-- Name: alert_rule_version_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.alert_rule_version_id_seq OWNED BY public.alert_rule_version.id;


--
-- TOC entry 253 (class 1259 OID 20584)
-- Name: annotation; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.annotation (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    alert_id bigint,
    user_id bigint,
    dashboard_id bigint,
    panel_id bigint,
    category_id bigint,
    type character varying(25) NOT NULL,
    title text NOT NULL,
    text text NOT NULL,
    metric character varying(255),
    prev_state character varying(25) NOT NULL,
    new_state character varying(25) NOT NULL,
    data text NOT NULL,
    epoch bigint NOT NULL,
    region_id bigint DEFAULT 0,
    tags character varying(4096),
    created bigint DEFAULT 0,
    updated bigint DEFAULT 0,
    epoch_end bigint DEFAULT 0 NOT NULL
);


ALTER TABLE public.annotation OWNER TO postgres;

--
-- TOC entry 252 (class 1259 OID 20583)
-- Name: annotation_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.annotation_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.annotation_id_seq OWNER TO postgres;

--
-- TOC entry 4093 (class 0 OID 0)
-- Dependencies: 252
-- Name: annotation_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.annotation_id_seq OWNED BY public.annotation.id;


--
-- TOC entry 255 (class 1259 OID 20604)
-- Name: annotation_tag; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.annotation_tag (
    id integer NOT NULL,
    annotation_id bigint NOT NULL,
    tag_id bigint NOT NULL
);


ALTER TABLE public.annotation_tag OWNER TO postgres;

--
-- TOC entry 254 (class 1259 OID 20603)
-- Name: annotation_tag_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.annotation_tag_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.annotation_tag_id_seq OWNER TO postgres;

--
-- TOC entry 4094 (class 0 OID 0)
-- Dependencies: 254
-- Name: annotation_tag_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.annotation_tag_id_seq OWNED BY public.annotation_tag.id;


--
-- TOC entry 230 (class 1259 OID 20426)
-- Name: api_key; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.api_key (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    name character varying(190) NOT NULL,
    key character varying(190) NOT NULL,
    role character varying(255) NOT NULL,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    expires bigint,
    service_account_id bigint,
    last_used_at timestamp without time zone,
    is_revoked boolean DEFAULT false
);


ALTER TABLE public.api_key OWNER TO postgres;

--
-- TOC entry 229 (class 1259 OID 20425)
-- Name: api_key_id_seq1; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.api_key_id_seq1
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.api_key_id_seq1 OWNER TO postgres;

--
-- TOC entry 4095 (class 0 OID 0)
-- Dependencies: 229
-- Name: api_key_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.api_key_id_seq1 OWNED BY public.api_key.id;


--
-- TOC entry 312 (class 1259 OID 20992)
-- Name: builtin_role; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.builtin_role (
    id integer NOT NULL,
    role character varying(190) NOT NULL,
    role_id bigint NOT NULL,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    org_id bigint DEFAULT 0 NOT NULL
);


ALTER TABLE public.builtin_role OWNER TO postgres;

--
-- TOC entry 311 (class 1259 OID 20991)
-- Name: builtin_role_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.builtin_role_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.builtin_role_id_seq OWNER TO postgres;

--
-- TOC entry 4096 (class 0 OID 0)
-- Dependencies: 311
-- Name: builtin_role_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.builtin_role_id_seq OWNED BY public.builtin_role.id;


--
-- TOC entry 276 (class 1259 OID 20729)
-- Name: cache_data; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cache_data (
    cache_key character varying(168) NOT NULL,
    data bytea NOT NULL,
    expires integer NOT NULL,
    created_at integer NOT NULL
);


ALTER TABLE public.cache_data OWNER TO postgres;

--
-- TOC entry 318 (class 1259 OID 21053)
-- Name: correlation; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.correlation (
    uid character varying(40) NOT NULL,
    source_uid character varying(40) NOT NULL,
    target_uid character varying(40),
    label text NOT NULL,
    description text NOT NULL,
    config text
);


ALTER TABLE public.correlation OWNER TO postgres;

--
-- TOC entry 224 (class 1259 OID 20332)
-- Name: dashboard; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.dashboard (
    id integer NOT NULL,
    version integer NOT NULL,
    slug character varying(189) NOT NULL,
    title character varying(189) NOT NULL,
    data text NOT NULL,
    org_id bigint NOT NULL,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    updated_by integer,
    created_by integer,
    gnet_id bigint,
    plugin_id character varying(189),
    folder_id bigint DEFAULT 0 NOT NULL,
    is_folder boolean DEFAULT false NOT NULL,
    has_acl boolean DEFAULT false NOT NULL,
    uid character varying(40),
    is_public boolean DEFAULT false NOT NULL
);


ALTER TABLE public.dashboard OWNER TO postgres;

--
-- TOC entry 265 (class 1259 OID 20658)
-- Name: dashboard_acl; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.dashboard_acl (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    dashboard_id bigint NOT NULL,
    user_id bigint,
    team_id bigint,
    permission smallint DEFAULT 4 NOT NULL,
    role character varying(20),
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL
);


ALTER TABLE public.dashboard_acl OWNER TO postgres;

--
-- TOC entry 264 (class 1259 OID 20657)
-- Name: dashboard_acl_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.dashboard_acl_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.dashboard_acl_id_seq OWNER TO postgres;

--
-- TOC entry 4097 (class 0 OID 0)
-- Dependencies: 264
-- Name: dashboard_acl_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.dashboard_acl_id_seq OWNED BY public.dashboard_acl.id;


--
-- TOC entry 223 (class 1259 OID 20331)
-- Name: dashboard_id_seq1; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.dashboard_id_seq1
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.dashboard_id_seq1 OWNER TO postgres;

--
-- TOC entry 4098 (class 0 OID 0)
-- Dependencies: 223
-- Name: dashboard_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.dashboard_id_seq1 OWNED BY public.dashboard.id;


--
-- TOC entry 226 (class 1259 OID 20372)
-- Name: dashboard_provisioning; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.dashboard_provisioning (
    id integer NOT NULL,
    dashboard_id bigint,
    name character varying(150) NOT NULL,
    external_id text NOT NULL,
    updated integer DEFAULT 0 NOT NULL,
    check_sum character varying(32)
);


ALTER TABLE public.dashboard_provisioning OWNER TO postgres;

--
-- TOC entry 225 (class 1259 OID 20371)
-- Name: dashboard_provisioning_id_seq1; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.dashboard_provisioning_id_seq1
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.dashboard_provisioning_id_seq1 OWNER TO postgres;

--
-- TOC entry 4099 (class 0 OID 0)
-- Dependencies: 225
-- Name: dashboard_provisioning_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.dashboard_provisioning_id_seq1 OWNED BY public.dashboard_provisioning.id;


--
-- TOC entry 321 (class 1259 OID 21089)
-- Name: dashboard_public; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.dashboard_public (
    uid character varying(40) NOT NULL,
    dashboard_uid character varying(40) NOT NULL,
    org_id bigint NOT NULL,
    time_settings text,
    template_variables text,
    access_token character varying(32) NOT NULL,
    created_by integer NOT NULL,
    updated_by integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone,
    is_enabled boolean DEFAULT false NOT NULL,
    annotations_enabled boolean DEFAULT false NOT NULL,
    time_selection_enabled boolean DEFAULT false NOT NULL,
    share character varying(64) DEFAULT 'public'::character varying NOT NULL
);


ALTER TABLE public.dashboard_public OWNER TO postgres;

--
-- TOC entry 232 (class 1259 OID 20450)
-- Name: dashboard_snapshot; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.dashboard_snapshot (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    key character varying(190) NOT NULL,
    delete_key character varying(190) NOT NULL,
    org_id bigint NOT NULL,
    user_id bigint NOT NULL,
    external boolean NOT NULL,
    external_url character varying(255) NOT NULL,
    dashboard text NOT NULL,
    expires timestamp without time zone NOT NULL,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    external_delete_url character varying(255),
    dashboard_encrypted bytea
);


ALTER TABLE public.dashboard_snapshot OWNER TO postgres;

--
-- TOC entry 231 (class 1259 OID 20449)
-- Name: dashboard_snapshot_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.dashboard_snapshot_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.dashboard_snapshot_id_seq OWNER TO postgres;

--
-- TOC entry 4100 (class 0 OID 0)
-- Dependencies: 231
-- Name: dashboard_snapshot_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.dashboard_snapshot_id_seq OWNED BY public.dashboard_snapshot.id;


--
-- TOC entry 222 (class 1259 OID 20324)
-- Name: dashboard_tag; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.dashboard_tag (
    id integer NOT NULL,
    dashboard_id bigint NOT NULL,
    term character varying(50) NOT NULL
);


ALTER TABLE public.dashboard_tag OWNER TO postgres;

--
-- TOC entry 221 (class 1259 OID 20323)
-- Name: dashboard_tag_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.dashboard_tag_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.dashboard_tag_id_seq OWNER TO postgres;

--
-- TOC entry 4101 (class 0 OID 0)
-- Dependencies: 221
-- Name: dashboard_tag_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.dashboard_tag_id_seq OWNED BY public.dashboard_tag.id;


--
-- TOC entry 259 (class 1259 OID 20628)
-- Name: dashboard_version; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.dashboard_version (
    id integer NOT NULL,
    dashboard_id bigint NOT NULL,
    parent_version integer NOT NULL,
    restored_from integer NOT NULL,
    version integer NOT NULL,
    created timestamp without time zone NOT NULL,
    created_by bigint NOT NULL,
    message text NOT NULL,
    data text NOT NULL
);


ALTER TABLE public.dashboard_version OWNER TO postgres;

--
-- TOC entry 258 (class 1259 OID 20627)
-- Name: dashboard_version_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.dashboard_version_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.dashboard_version_id_seq OWNER TO postgres;

--
-- TOC entry 4102 (class 0 OID 0)
-- Dependencies: 258
-- Name: dashboard_version_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.dashboard_version_id_seq OWNED BY public.dashboard_version.id;


--
-- TOC entry 298 (class 1259 OID 20923)
-- Name: data_keys; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.data_keys (
    name character varying(100) NOT NULL,
    active boolean NOT NULL,
    scope character varying(30) NOT NULL,
    provider character varying(50) NOT NULL,
    encrypted_data bytea NOT NULL,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    label character varying(100) DEFAULT ''::character varying NOT NULL
);


ALTER TABLE public.data_keys OWNER TO postgres;

--
-- TOC entry 228 (class 1259 OID 20398)
-- Name: data_source; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.data_source (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    version integer NOT NULL,
    type character varying(255) NOT NULL,
    name character varying(190) NOT NULL,
    access character varying(255) NOT NULL,
    url character varying(255) NOT NULL,
    password character varying(255),
    "user" character varying(255),
    database character varying(255),
    basic_auth boolean NOT NULL,
    basic_auth_user character varying(255),
    basic_auth_password character varying(255),
    is_default boolean NOT NULL,
    json_data text,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    with_credentials boolean DEFAULT false NOT NULL,
    secure_json_data text,
    read_only boolean,
    uid character varying(40) DEFAULT 0 NOT NULL
);


ALTER TABLE public.data_source OWNER TO postgres;

--
-- TOC entry 227 (class 1259 OID 20397)
-- Name: data_source_id_seq1; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.data_source_id_seq1
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.data_source_id_seq1 OWNER TO postgres;

--
-- TOC entry 4103 (class 0 OID 0)
-- Dependencies: 227
-- Name: data_source_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.data_source_id_seq1 OWNED BY public.data_source.id;


--
-- TOC entry 320 (class 1259 OID 21063)
-- Name: entity_event; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.entity_event (
    id integer NOT NULL,
    entity_id character varying(1024) NOT NULL,
    event_type character varying(8) NOT NULL,
    created bigint NOT NULL
);


ALTER TABLE public.entity_event OWNER TO postgres;

--
-- TOC entry 319 (class 1259 OID 21062)
-- Name: entity_event_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.entity_event_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.entity_event_id_seq OWNER TO postgres;

--
-- TOC entry 4104 (class 0 OID 0)
-- Dependencies: 319
-- Name: entity_event_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.entity_event_id_seq OWNED BY public.entity_event.id;


--
-- TOC entry 322 (class 1259 OID 21103)
-- Name: file; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.file (
    path character varying(1024) NOT NULL COLLATE pg_catalog."C",
    path_hash character varying(64) NOT NULL,
    parent_folder_path_hash character varying(64) NOT NULL,
    contents bytea NOT NULL,
    etag character varying(32) NOT NULL,
    cache_control character varying(128) NOT NULL,
    content_disposition character varying(128) NOT NULL,
    updated timestamp without time zone NOT NULL,
    created timestamp without time zone NOT NULL,
    size bigint NOT NULL,
    mime_type character varying(255) NOT NULL
);


ALTER TABLE public.file OWNER TO postgres;

--
-- TOC entry 323 (class 1259 OID 21110)
-- Name: file_meta; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.file_meta (
    path_hash character varying(64) NOT NULL,
    key character varying(191) NOT NULL,
    value character varying(1024) NOT NULL
);


ALTER TABLE public.file_meta OWNER TO postgres;

--
-- TOC entry 326 (class 1259 OID 21135)
-- Name: folder; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.folder (
    id integer NOT NULL,
    uid character varying(40) NOT NULL,
    org_id bigint NOT NULL,
    title character varying(189) NOT NULL,
    description character varying(255),
    parent_uid character varying(40),
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL
);


ALTER TABLE public.folder OWNER TO postgres;

--
-- TOC entry 325 (class 1259 OID 21134)
-- Name: folder_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.folder_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.folder_id_seq OWNER TO postgres;

--
-- TOC entry 4105 (class 0 OID 0)
-- Dependencies: 325
-- Name: folder_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.folder_id_seq OWNED BY public.folder.id;


--
-- TOC entry 302 (class 1259 OID 20941)
-- Name: kv_store; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.kv_store (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    namespace character varying(190) NOT NULL,
    key character varying(190) NOT NULL,
    value text NOT NULL,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL
);


ALTER TABLE public.kv_store OWNER TO postgres;

--
-- TOC entry 301 (class 1259 OID 20940)
-- Name: kv_store_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.kv_store_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.kv_store_id_seq OWNER TO postgres;

--
-- TOC entry 4106 (class 0 OID 0)
-- Dependencies: 301
-- Name: kv_store_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.kv_store_id_seq OWNED BY public.kv_store.id;


--
-- TOC entry 295 (class 1259 OID 20905)
-- Name: library_element; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.library_element (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    folder_id bigint NOT NULL,
    uid character varying(40) NOT NULL,
    name character varying(150) NOT NULL,
    kind bigint NOT NULL,
    type character varying(40) NOT NULL,
    description character varying(2048) NOT NULL,
    model text NOT NULL,
    created timestamp without time zone NOT NULL,
    created_by bigint NOT NULL,
    updated timestamp without time zone NOT NULL,
    updated_by bigint NOT NULL,
    version bigint NOT NULL
);


ALTER TABLE public.library_element OWNER TO postgres;

--
-- TOC entry 297 (class 1259 OID 20915)
-- Name: library_element_connection; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.library_element_connection (
    id integer NOT NULL,
    element_id bigint NOT NULL,
    kind bigint NOT NULL,
    connection_id bigint NOT NULL,
    created timestamp without time zone NOT NULL,
    created_by bigint NOT NULL
);


ALTER TABLE public.library_element_connection OWNER TO postgres;

--
-- TOC entry 296 (class 1259 OID 20914)
-- Name: library_element_connection_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.library_element_connection_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.library_element_connection_id_seq OWNER TO postgres;

--
-- TOC entry 4107 (class 0 OID 0)
-- Dependencies: 296
-- Name: library_element_connection_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.library_element_connection_id_seq OWNED BY public.library_element_connection.id;


--
-- TOC entry 294 (class 1259 OID 20904)
-- Name: library_element_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.library_element_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.library_element_id_seq OWNER TO postgres;

--
-- TOC entry 4108 (class 0 OID 0)
-- Dependencies: 294
-- Name: library_element_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.library_element_id_seq OWNED BY public.library_element.id;


--
-- TOC entry 269 (class 1259 OID 20689)
-- Name: login_attempt; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.login_attempt (
    id integer NOT NULL,
    username character varying(190) NOT NULL,
    ip_address character varying(30) NOT NULL,
    created integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.login_attempt OWNER TO postgres;

--
-- TOC entry 268 (class 1259 OID 20688)
-- Name: login_attempt_id_seq1; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.login_attempt_id_seq1
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.login_attempt_id_seq1 OWNER TO postgres;

--
-- TOC entry 4109 (class 0 OID 0)
-- Dependencies: 268
-- Name: login_attempt_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.login_attempt_id_seq1 OWNED BY public.login_attempt.id;


--
-- TOC entry 210 (class 1259 OID 20216)
-- Name: migration_log; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.migration_log (
    id integer NOT NULL,
    migration_id character varying(255) NOT NULL,
    sql text NOT NULL,
    success boolean NOT NULL,
    error text NOT NULL,
    "timestamp" timestamp without time zone NOT NULL
);


ALTER TABLE public.migration_log OWNER TO postgres;

--
-- TOC entry 209 (class 1259 OID 20215)
-- Name: migration_log_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.migration_log_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.migration_log_id_seq OWNER TO postgres;

--
-- TOC entry 4110 (class 0 OID 0)
-- Dependencies: 209
-- Name: migration_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.migration_log_id_seq OWNED BY public.migration_log.id;


--
-- TOC entry 287 (class 1259 OID 20860)
-- Name: ngalert_configuration; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ngalert_configuration (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    alertmanagers text,
    created_at integer NOT NULL,
    updated_at integer NOT NULL,
    send_alerts_to smallint DEFAULT 0 NOT NULL
);


ALTER TABLE public.ngalert_configuration OWNER TO postgres;

--
-- TOC entry 286 (class 1259 OID 20859)
-- Name: ngalert_configuration_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.ngalert_configuration_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ngalert_configuration_id_seq OWNER TO postgres;

--
-- TOC entry 4111 (class 0 OID 0)
-- Dependencies: 286
-- Name: ngalert_configuration_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ngalert_configuration_id_seq OWNED BY public.ngalert_configuration.id;


--
-- TOC entry 218 (class 1259 OID 20292)
-- Name: org; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.org (
    id integer NOT NULL,
    version integer NOT NULL,
    name character varying(190) NOT NULL,
    address1 character varying(255),
    address2 character varying(255),
    city character varying(255),
    state character varying(255),
    zip_code character varying(50),
    country character varying(255),
    billing_email character varying(255),
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL
);


ALTER TABLE public.org OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 20291)
-- Name: org_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.org_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.org_id_seq OWNER TO postgres;

--
-- TOC entry 4112 (class 0 OID 0)
-- Dependencies: 217
-- Name: org_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.org_id_seq OWNED BY public.org.id;


--
-- TOC entry 220 (class 1259 OID 20302)
-- Name: org_user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.org_user (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    user_id bigint NOT NULL,
    role character varying(20) NOT NULL,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL
);


ALTER TABLE public.org_user OWNER TO postgres;

--
-- TOC entry 219 (class 1259 OID 20301)
-- Name: org_user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.org_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.org_user_id_seq OWNER TO postgres;

--
-- TOC entry 4113 (class 0 OID 0)
-- Dependencies: 219
-- Name: org_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.org_user_id_seq OWNED BY public.org_user.id;


--
-- TOC entry 304 (class 1259 OID 20951)
-- Name: permission; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.permission (
    id integer NOT NULL,
    role_id bigint NOT NULL,
    action character varying(190) NOT NULL,
    scope character varying(190) NOT NULL,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL
);


ALTER TABLE public.permission OWNER TO postgres;

--
-- TOC entry 303 (class 1259 OID 20950)
-- Name: permission_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.permission_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.permission_id_seq OWNER TO postgres;

--
-- TOC entry 4114 (class 0 OID 0)
-- Dependencies: 303
-- Name: permission_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.permission_id_seq OWNED BY public.permission.id;


--
-- TOC entry 239 (class 1259 OID 20492)
-- Name: playlist; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.playlist (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    "interval" character varying(255) NOT NULL,
    org_id bigint NOT NULL,
    uid character varying(80) DEFAULT 0 NOT NULL
);


ALTER TABLE public.playlist OWNER TO postgres;

--
-- TOC entry 238 (class 1259 OID 20491)
-- Name: playlist_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.playlist_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.playlist_id_seq OWNER TO postgres;

--
-- TOC entry 4115 (class 0 OID 0)
-- Dependencies: 238
-- Name: playlist_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.playlist_id_seq OWNED BY public.playlist.id;


--
-- TOC entry 241 (class 1259 OID 20504)
-- Name: playlist_item; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.playlist_item (
    id integer NOT NULL,
    playlist_id bigint NOT NULL,
    type character varying(255) NOT NULL,
    value text NOT NULL,
    title text NOT NULL,
    "order" integer NOT NULL
);


ALTER TABLE public.playlist_item OWNER TO postgres;

--
-- TOC entry 240 (class 1259 OID 20501)
-- Name: playlist_item_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.playlist_item_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.playlist_item_id_seq OWNER TO postgres;

--
-- TOC entry 4116 (class 0 OID 0)
-- Dependencies: 240
-- Name: playlist_item_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.playlist_item_id_seq OWNED BY public.playlist_item.id;


--
-- TOC entry 236 (class 1259 OID 20474)
-- Name: plugin_setting; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.plugin_setting (
    id integer NOT NULL,
    org_id bigint,
    plugin_id character varying(190) NOT NULL,
    enabled boolean NOT NULL,
    pinned boolean NOT NULL,
    json_data text,
    secure_json_data text,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    plugin_version character varying(50)
);


ALTER TABLE public.plugin_setting OWNER TO postgres;

--
-- TOC entry 235 (class 1259 OID 20473)
-- Name: plugin_setting_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.plugin_setting_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.plugin_setting_id_seq OWNER TO postgres;

--
-- TOC entry 4117 (class 0 OID 0)
-- Dependencies: 235
-- Name: plugin_setting_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.plugin_setting_id_seq OWNED BY public.plugin_setting.id;


--
-- TOC entry 243 (class 1259 OID 20515)
-- Name: preferences; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.preferences (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    user_id bigint NOT NULL,
    version integer NOT NULL,
    home_dashboard_id bigint NOT NULL,
    timezone character varying(50) NOT NULL,
    theme character varying(20) NOT NULL,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    team_id bigint,
    week_start character varying(10),
    json_data text
);


ALTER TABLE public.preferences OWNER TO postgres;

--
-- TOC entry 242 (class 1259 OID 20514)
-- Name: preferences_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.preferences_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.preferences_id_seq OWNER TO postgres;

--
-- TOC entry 4118 (class 0 OID 0)
-- Dependencies: 242
-- Name: preferences_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.preferences_id_seq OWNED BY public.preferences.id;


--
-- TOC entry 289 (class 1259 OID 20871)
-- Name: provenance_type; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.provenance_type (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    record_key character varying(190) NOT NULL,
    record_type character varying(190) NOT NULL,
    provenance character varying(190) NOT NULL
);


ALTER TABLE public.provenance_type OWNER TO postgres;

--
-- TOC entry 288 (class 1259 OID 20870)
-- Name: provenance_type_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.provenance_type_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.provenance_type_id_seq OWNER TO postgres;

--
-- TOC entry 4119 (class 0 OID 0)
-- Dependencies: 288
-- Name: provenance_type_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.provenance_type_id_seq OWNED BY public.provenance_type.id;


--
-- TOC entry 315 (class 1259 OID 21012)
-- Name: query_history; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.query_history (
    id integer NOT NULL,
    uid character varying(40) NOT NULL,
    org_id bigint NOT NULL,
    datasource_uid character varying(40) NOT NULL,
    created_by bigint NOT NULL,
    created_at integer NOT NULL,
    comment text NOT NULL,
    queries text NOT NULL
);


ALTER TABLE public.query_history OWNER TO postgres;

--
-- TOC entry 314 (class 1259 OID 21011)
-- Name: query_history_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.query_history_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.query_history_id_seq OWNER TO postgres;

--
-- TOC entry 4120 (class 0 OID 0)
-- Dependencies: 314
-- Name: query_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.query_history_id_seq OWNED BY public.query_history.id;


--
-- TOC entry 317 (class 1259 OID 21039)
-- Name: query_history_star; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.query_history_star (
    id integer NOT NULL,
    query_uid character varying(40) NOT NULL,
    user_id bigint NOT NULL,
    org_id bigint DEFAULT 1 NOT NULL
);


ALTER TABLE public.query_history_star OWNER TO postgres;

--
-- TOC entry 316 (class 1259 OID 21038)
-- Name: query_history_star_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.query_history_star_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.query_history_star_id_seq OWNER TO postgres;

--
-- TOC entry 4121 (class 0 OID 0)
-- Dependencies: 316
-- Name: query_history_star_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.query_history_star_id_seq OWNED BY public.query_history_star.id;


--
-- TOC entry 234 (class 1259 OID 20465)
-- Name: quota; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.quota (
    id integer NOT NULL,
    org_id bigint,
    user_id bigint,
    target character varying(190) NOT NULL,
    "limit" bigint NOT NULL,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL
);


ALTER TABLE public.quota OWNER TO postgres;

--
-- TOC entry 233 (class 1259 OID 20464)
-- Name: quota_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.quota_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.quota_id_seq OWNER TO postgres;

--
-- TOC entry 4122 (class 0 OID 0)
-- Dependencies: 233
-- Name: quota_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.quota_id_seq OWNED BY public.quota.id;


--
-- TOC entry 306 (class 1259 OID 20960)
-- Name: role; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.role (
    id integer NOT NULL,
    name character varying(190) NOT NULL,
    description text,
    version bigint NOT NULL,
    org_id bigint NOT NULL,
    uid character varying(40) NOT NULL,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    display_name character varying(190),
    group_name character varying(190),
    hidden boolean DEFAULT false NOT NULL
);


ALTER TABLE public.role OWNER TO postgres;

--
-- TOC entry 305 (class 1259 OID 20959)
-- Name: role_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.role_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.role_id_seq OWNER TO postgres;

--
-- TOC entry 4123 (class 0 OID 0)
-- Dependencies: 305
-- Name: role_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.role_id_seq OWNED BY public.role.id;


--
-- TOC entry 300 (class 1259 OID 20931)
-- Name: secrets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.secrets (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    namespace character varying(255) NOT NULL,
    type character varying(255) NOT NULL,
    value text,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL
);


ALTER TABLE public.secrets OWNER TO postgres;

--
-- TOC entry 299 (class 1259 OID 20930)
-- Name: secrets_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.secrets_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.secrets_id_seq OWNER TO postgres;

--
-- TOC entry 4124 (class 0 OID 0)
-- Dependencies: 299
-- Name: secrets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.secrets_id_seq OWNED BY public.secrets.id;


--
-- TOC entry 313 (class 1259 OID 21004)
-- Name: seed_assignment; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.seed_assignment (
    builtin_role character varying(190) NOT NULL,
    role_name character varying(190),
    action character varying(190),
    scope character varying(190),
    id integer NOT NULL
);


ALTER TABLE public.seed_assignment OWNER TO postgres;

--
-- TOC entry 324 (class 1259 OID 21122)
-- Name: seed_assignment_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.seed_assignment_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.seed_assignment_id_seq OWNER TO postgres;

--
-- TOC entry 4125 (class 0 OID 0)
-- Dependencies: 324
-- Name: seed_assignment_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.seed_assignment_id_seq OWNED BY public.seed_assignment.id;


--
-- TOC entry 273 (class 1259 OID 20710)
-- Name: server_lock; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.server_lock (
    id integer NOT NULL,
    operation_uid character varying(100) NOT NULL,
    version bigint NOT NULL,
    last_execution bigint NOT NULL
);


ALTER TABLE public.server_lock OWNER TO postgres;

--
-- TOC entry 272 (class 1259 OID 20709)
-- Name: server_lock_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.server_lock_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.server_lock_id_seq OWNER TO postgres;

--
-- TOC entry 4126 (class 0 OID 0)
-- Dependencies: 272
-- Name: server_lock_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.server_lock_id_seq OWNED BY public.server_lock.id;


--
-- TOC entry 237 (class 1259 OID 20484)
-- Name: session; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.session (
    key character(16) NOT NULL,
    data bytea NOT NULL,
    expiry integer NOT NULL
);


ALTER TABLE public.session OWNER TO postgres;

--
-- TOC entry 278 (class 1259 OID 20738)
-- Name: short_url; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.short_url (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    uid character varying(40) NOT NULL,
    path text NOT NULL,
    created_by bigint NOT NULL,
    created_at integer NOT NULL,
    last_seen_at integer
);


ALTER TABLE public.short_url OWNER TO postgres;

--
-- TOC entry 277 (class 1259 OID 20737)
-- Name: short_url_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.short_url_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.short_url_id_seq OWNER TO postgres;

--
-- TOC entry 4127 (class 0 OID 0)
-- Dependencies: 277
-- Name: short_url_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.short_url_id_seq OWNED BY public.short_url.id;


--
-- TOC entry 216 (class 1259 OID 20284)
-- Name: star; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.star (
    id integer NOT NULL,
    user_id bigint NOT NULL,
    dashboard_id bigint NOT NULL
);


ALTER TABLE public.star OWNER TO postgres;

--
-- TOC entry 215 (class 1259 OID 20283)
-- Name: star_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.star_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.star_id_seq OWNER TO postgres;

--
-- TOC entry 4128 (class 0 OID 0)
-- Dependencies: 215
-- Name: star_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.star_id_seq OWNED BY public.star.id;


--
-- TOC entry 267 (class 1259 OID 20673)
-- Name: tag; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tag (
    id integer NOT NULL,
    key character varying(100) NOT NULL,
    value character varying(100) NOT NULL
);


ALTER TABLE public.tag OWNER TO postgres;

--
-- TOC entry 266 (class 1259 OID 20672)
-- Name: tag_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.tag_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.tag_id_seq OWNER TO postgres;

--
-- TOC entry 4129 (class 0 OID 0)
-- Dependencies: 266
-- Name: tag_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.tag_id_seq OWNED BY public.tag.id;


--
-- TOC entry 261 (class 1259 OID 20639)
-- Name: team; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.team (
    id integer NOT NULL,
    name character varying(190) NOT NULL,
    org_id bigint NOT NULL,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    email character varying(190)
);


ALTER TABLE public.team OWNER TO postgres;

--
-- TOC entry 260 (class 1259 OID 20638)
-- Name: team_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.team_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.team_id_seq OWNER TO postgres;

--
-- TOC entry 4130 (class 0 OID 0)
-- Dependencies: 260
-- Name: team_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.team_id_seq OWNED BY public.team.id;


--
-- TOC entry 263 (class 1259 OID 20648)
-- Name: team_member; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.team_member (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    team_id bigint NOT NULL,
    user_id bigint NOT NULL,
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    external boolean,
    permission smallint
);


ALTER TABLE public.team_member OWNER TO postgres;

--
-- TOC entry 262 (class 1259 OID 20647)
-- Name: team_member_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.team_member_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.team_member_id_seq OWNER TO postgres;

--
-- TOC entry 4131 (class 0 OID 0)
-- Dependencies: 262
-- Name: team_member_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.team_member_id_seq OWNED BY public.team_member.id;


--
-- TOC entry 308 (class 1259 OID 20972)
-- Name: team_role; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.team_role (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    team_id bigint NOT NULL,
    role_id bigint NOT NULL,
    created timestamp without time zone NOT NULL
);


ALTER TABLE public.team_role OWNER TO postgres;

--
-- TOC entry 307 (class 1259 OID 20971)
-- Name: team_role_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.team_role_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.team_role_id_seq OWNER TO postgres;

--
-- TOC entry 4132 (class 0 OID 0)
-- Dependencies: 307
-- Name: team_role_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.team_role_id_seq OWNED BY public.team_role.id;


--
-- TOC entry 214 (class 1259 OID 20269)
-- Name: temp_user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.temp_user (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    version integer NOT NULL,
    email character varying(190) NOT NULL,
    name character varying(255),
    role character varying(20),
    code character varying(190) NOT NULL,
    status character varying(20) NOT NULL,
    invited_by_user_id bigint,
    email_sent boolean NOT NULL,
    email_sent_on timestamp without time zone,
    remote_addr character varying(255),
    created integer DEFAULT 0 NOT NULL,
    updated integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.temp_user OWNER TO postgres;

--
-- TOC entry 213 (class 1259 OID 20268)
-- Name: temp_user_id_seq1; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.temp_user_id_seq1
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.temp_user_id_seq1 OWNER TO postgres;

--
-- TOC entry 4133 (class 0 OID 0)
-- Dependencies: 213
-- Name: temp_user_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.temp_user_id_seq1 OWNED BY public.temp_user.id;


--
-- TOC entry 257 (class 1259 OID 20621)
-- Name: test_data; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.test_data (
    id integer NOT NULL,
    metric1 character varying(20),
    metric2 character varying(150),
    value_big_int bigint,
    value_double double precision,
    value_float real,
    value_int integer,
    time_epoch bigint NOT NULL,
    time_date_time timestamp without time zone NOT NULL,
    time_time_stamp timestamp without time zone NOT NULL
);


ALTER TABLE public.test_data OWNER TO postgres;

--
-- TOC entry 256 (class 1259 OID 20620)
-- Name: test_data_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.test_data_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.test_data_id_seq OWNER TO postgres;

--
-- TOC entry 4134 (class 0 OID 0)
-- Dependencies: 256
-- Name: test_data_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.test_data_id_seq OWNED BY public.test_data.id;


--
-- TOC entry 212 (class 1259 OID 20236)
-- Name: user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."user" (
    id integer NOT NULL,
    version integer NOT NULL,
    login character varying(190) NOT NULL,
    email character varying(190) NOT NULL,
    name character varying(255),
    password character varying(255),
    salt character varying(50),
    rands character varying(50),
    company character varying(255),
    org_id bigint NOT NULL,
    is_admin boolean NOT NULL,
    email_verified boolean,
    theme character varying(255),
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    help_flags1 bigint DEFAULT 0 NOT NULL,
    last_seen_at timestamp without time zone,
    is_disabled boolean DEFAULT false NOT NULL,
    is_service_account boolean DEFAULT false
);


ALTER TABLE public."user" OWNER TO postgres;

--
-- TOC entry 271 (class 1259 OID 20698)
-- Name: user_auth; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_auth (
    id integer NOT NULL,
    user_id bigint NOT NULL,
    auth_module character varying(190) NOT NULL,
    auth_id character varying(190) NOT NULL,
    created timestamp without time zone NOT NULL,
    o_auth_access_token text,
    o_auth_refresh_token text,
    o_auth_token_type text,
    o_auth_expiry timestamp without time zone,
    o_auth_id_token text
);


ALTER TABLE public.user_auth OWNER TO postgres;

--
-- TOC entry 270 (class 1259 OID 20697)
-- Name: user_auth_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_auth_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_auth_id_seq OWNER TO postgres;

--
-- TOC entry 4135 (class 0 OID 0)
-- Dependencies: 270
-- Name: user_auth_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_auth_id_seq OWNED BY public.user_auth.id;


--
-- TOC entry 275 (class 1259 OID 20718)
-- Name: user_auth_token; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_auth_token (
    id integer NOT NULL,
    user_id bigint NOT NULL,
    auth_token character varying(100) NOT NULL,
    prev_auth_token character varying(100) NOT NULL,
    user_agent character varying(255) NOT NULL,
    client_ip character varying(255) NOT NULL,
    auth_token_seen boolean NOT NULL,
    seen_at integer,
    rotated_at integer NOT NULL,
    created_at integer NOT NULL,
    updated_at integer NOT NULL,
    revoked_at integer
);


ALTER TABLE public.user_auth_token OWNER TO postgres;

--
-- TOC entry 274 (class 1259 OID 20717)
-- Name: user_auth_token_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_auth_token_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_auth_token_id_seq OWNER TO postgres;

--
-- TOC entry 4136 (class 0 OID 0)
-- Dependencies: 274
-- Name: user_auth_token_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_auth_token_id_seq OWNED BY public.user_auth_token.id;


--
-- TOC entry 211 (class 1259 OID 20235)
-- Name: user_id_seq1; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_id_seq1
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_id_seq1 OWNER TO postgres;

--
-- TOC entry 4137 (class 0 OID 0)
-- Dependencies: 211
-- Name: user_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_id_seq1 OWNED BY public."user".id;


--
-- TOC entry 310 (class 1259 OID 20982)
-- Name: user_role; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_role (
    id integer NOT NULL,
    org_id bigint NOT NULL,
    user_id bigint NOT NULL,
    role_id bigint NOT NULL,
    created timestamp without time zone NOT NULL
);


ALTER TABLE public.user_role OWNER TO postgres;

--
-- TOC entry 309 (class 1259 OID 20981)
-- Name: user_role_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_role_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_role_id_seq OWNER TO postgres;

--
-- TOC entry 4138 (class 0 OID 0)
-- Dependencies: 309
-- Name: user_role_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_role_id_seq OWNED BY public.user_role.id;


--
-- TOC entry 3497 (class 2604 OID 20529)
-- Name: alert id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert ALTER COLUMN id SET DEFAULT nextval('public.alert_id_seq'::regclass);


--
-- TOC entry 3541 (class 2604 OID 20850)
-- Name: alert_configuration id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert_configuration ALTER COLUMN id SET DEFAULT nextval('public.alert_configuration_id_seq'::regclass);


--
-- TOC entry 3549 (class 2604 OID 20894)
-- Name: alert_configuration_history id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert_configuration_history ALTER COLUMN id SET DEFAULT nextval('public.alert_configuration_history_id_seq'::regclass);


--
-- TOC entry 3548 (class 2604 OID 20884)
-- Name: alert_image id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert_image ALTER COLUMN id SET DEFAULT nextval('public.alert_image_id_seq'::regclass);


--
-- TOC entry 3499 (class 2604 OID 20553)
-- Name: alert_notification id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert_notification ALTER COLUMN id SET DEFAULT nextval('public.alert_notification_id_seq'::regclass);


--
-- TOC entry 3503 (class 2604 OID 20576)
-- Name: alert_notification_state id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert_notification_state ALTER COLUMN id SET DEFAULT nextval('public.alert_notification_state_id_seq'::regclass);


--
-- TOC entry 3525 (class 2604 OID 20809)
-- Name: alert_rule id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert_rule ALTER COLUMN id SET DEFAULT nextval('public.alert_rule_id_seq'::regclass);


--
-- TOC entry 3498 (class 2604 OID 20545)
-- Name: alert_rule_tag id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert_rule_tag ALTER COLUMN id SET DEFAULT nextval('public.alert_rule_tag_id_seq'::regclass);


--
-- TOC entry 3534 (class 2604 OID 20832)
-- Name: alert_rule_version id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert_rule_version ALTER COLUMN id SET DEFAULT nextval('public.alert_rule_version_id_seq'::regclass);


--
-- TOC entry 3504 (class 2604 OID 20587)
-- Name: annotation id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.annotation ALTER COLUMN id SET DEFAULT nextval('public.annotation_id_seq'::regclass);


--
-- TOC entry 3509 (class 2604 OID 20607)
-- Name: annotation_tag id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.annotation_tag ALTER COLUMN id SET DEFAULT nextval('public.annotation_tag_id_seq'::regclass);


--
-- TOC entry 3488 (class 2604 OID 20429)
-- Name: api_key id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_key ALTER COLUMN id SET DEFAULT nextval('public.api_key_id_seq1'::regclass);


--
-- TOC entry 3564 (class 2604 OID 20995)
-- Name: builtin_role id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.builtin_role ALTER COLUMN id SET DEFAULT nextval('public.builtin_role_id_seq'::regclass);


--
-- TOC entry 3478 (class 2604 OID 20335)
-- Name: dashboard id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dashboard ALTER COLUMN id SET DEFAULT nextval('public.dashboard_id_seq1'::regclass);


--
-- TOC entry 3514 (class 2604 OID 20661)
-- Name: dashboard_acl id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dashboard_acl ALTER COLUMN id SET DEFAULT nextval('public.dashboard_acl_id_seq'::regclass);


--
-- TOC entry 3483 (class 2604 OID 20375)
-- Name: dashboard_provisioning id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dashboard_provisioning ALTER COLUMN id SET DEFAULT nextval('public.dashboard_provisioning_id_seq1'::regclass);


--
-- TOC entry 3490 (class 2604 OID 20453)
-- Name: dashboard_snapshot id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dashboard_snapshot ALTER COLUMN id SET DEFAULT nextval('public.dashboard_snapshot_id_seq'::regclass);


--
-- TOC entry 3477 (class 2604 OID 20327)
-- Name: dashboard_tag id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dashboard_tag ALTER COLUMN id SET DEFAULT nextval('public.dashboard_tag_id_seq'::regclass);


--
-- TOC entry 3511 (class 2604 OID 20631)
-- Name: dashboard_version id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dashboard_version ALTER COLUMN id SET DEFAULT nextval('public.dashboard_version_id_seq'::regclass);


--
-- TOC entry 3485 (class 2604 OID 20401)
-- Name: data_source id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.data_source ALTER COLUMN id SET DEFAULT nextval('public.data_source_id_seq1'::regclass);


--
-- TOC entry 3570 (class 2604 OID 21066)
-- Name: entity_event id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.entity_event ALTER COLUMN id SET DEFAULT nextval('public.entity_event_id_seq'::regclass);


--
-- TOC entry 3575 (class 2604 OID 21138)
-- Name: folder id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.folder ALTER COLUMN id SET DEFAULT nextval('public.folder_id_seq'::regclass);


--
-- TOC entry 3558 (class 2604 OID 20944)
-- Name: kv_store id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.kv_store ALTER COLUMN id SET DEFAULT nextval('public.kv_store_id_seq'::regclass);


--
-- TOC entry 3554 (class 2604 OID 20908)
-- Name: library_element id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.library_element ALTER COLUMN id SET DEFAULT nextval('public.library_element_id_seq'::regclass);


--
-- TOC entry 3555 (class 2604 OID 20918)
-- Name: library_element_connection id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.library_element_connection ALTER COLUMN id SET DEFAULT nextval('public.library_element_connection_id_seq'::regclass);


--
-- TOC entry 3517 (class 2604 OID 20692)
-- Name: login_attempt id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.login_attempt ALTER COLUMN id SET DEFAULT nextval('public.login_attempt_id_seq1'::regclass);


--
-- TOC entry 3466 (class 2604 OID 20219)
-- Name: migration_log id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.migration_log ALTER COLUMN id SET DEFAULT nextval('public.migration_log_id_seq'::regclass);


--
-- TOC entry 3545 (class 2604 OID 20863)
-- Name: ngalert_configuration id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ngalert_configuration ALTER COLUMN id SET DEFAULT nextval('public.ngalert_configuration_id_seq'::regclass);


--
-- TOC entry 3475 (class 2604 OID 20295)
-- Name: org id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.org ALTER COLUMN id SET DEFAULT nextval('public.org_id_seq'::regclass);


--
-- TOC entry 3476 (class 2604 OID 20305)
-- Name: org_user id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.org_user ALTER COLUMN id SET DEFAULT nextval('public.org_user_id_seq'::regclass);


--
-- TOC entry 3559 (class 2604 OID 20954)
-- Name: permission id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.permission ALTER COLUMN id SET DEFAULT nextval('public.permission_id_seq'::regclass);


--
-- TOC entry 3493 (class 2604 OID 20495)
-- Name: playlist id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.playlist ALTER COLUMN id SET DEFAULT nextval('public.playlist_id_seq'::regclass);


--
-- TOC entry 3495 (class 2604 OID 20507)
-- Name: playlist_item id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.playlist_item ALTER COLUMN id SET DEFAULT nextval('public.playlist_item_id_seq'::regclass);


--
-- TOC entry 3492 (class 2604 OID 20477)
-- Name: plugin_setting id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.plugin_setting ALTER COLUMN id SET DEFAULT nextval('public.plugin_setting_id_seq'::regclass);


--
-- TOC entry 3496 (class 2604 OID 20518)
-- Name: preferences id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.preferences ALTER COLUMN id SET DEFAULT nextval('public.preferences_id_seq'::regclass);


--
-- TOC entry 3547 (class 2604 OID 20874)
-- Name: provenance_type id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.provenance_type ALTER COLUMN id SET DEFAULT nextval('public.provenance_type_id_seq'::regclass);


--
-- TOC entry 3567 (class 2604 OID 21015)
-- Name: query_history id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.query_history ALTER COLUMN id SET DEFAULT nextval('public.query_history_id_seq'::regclass);


--
-- TOC entry 3568 (class 2604 OID 21042)
-- Name: query_history_star id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.query_history_star ALTER COLUMN id SET DEFAULT nextval('public.query_history_star_id_seq'::regclass);


--
-- TOC entry 3491 (class 2604 OID 20468)
-- Name: quota id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quota ALTER COLUMN id SET DEFAULT nextval('public.quota_id_seq'::regclass);


--
-- TOC entry 3560 (class 2604 OID 20963)
-- Name: role id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role ALTER COLUMN id SET DEFAULT nextval('public.role_id_seq'::regclass);


--
-- TOC entry 3557 (class 2604 OID 20934)
-- Name: secrets id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.secrets ALTER COLUMN id SET DEFAULT nextval('public.secrets_id_seq'::regclass);


--
-- TOC entry 3566 (class 2604 OID 21123)
-- Name: seed_assignment id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.seed_assignment ALTER COLUMN id SET DEFAULT nextval('public.seed_assignment_id_seq'::regclass);


--
-- TOC entry 3520 (class 2604 OID 20713)
-- Name: server_lock id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.server_lock ALTER COLUMN id SET DEFAULT nextval('public.server_lock_id_seq'::regclass);


--
-- TOC entry 3522 (class 2604 OID 20741)
-- Name: short_url id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.short_url ALTER COLUMN id SET DEFAULT nextval('public.short_url_id_seq'::regclass);


--
-- TOC entry 3474 (class 2604 OID 20287)
-- Name: star id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.star ALTER COLUMN id SET DEFAULT nextval('public.star_id_seq'::regclass);


--
-- TOC entry 3516 (class 2604 OID 20676)
-- Name: tag id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag ALTER COLUMN id SET DEFAULT nextval('public.tag_id_seq'::regclass);


--
-- TOC entry 3512 (class 2604 OID 20642)
-- Name: team id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.team ALTER COLUMN id SET DEFAULT nextval('public.team_id_seq'::regclass);


--
-- TOC entry 3513 (class 2604 OID 20651)
-- Name: team_member id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.team_member ALTER COLUMN id SET DEFAULT nextval('public.team_member_id_seq'::regclass);


--
-- TOC entry 3562 (class 2604 OID 20975)
-- Name: team_role id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.team_role ALTER COLUMN id SET DEFAULT nextval('public.team_role_id_seq'::regclass);


--
-- TOC entry 3471 (class 2604 OID 20272)
-- Name: temp_user id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.temp_user ALTER COLUMN id SET DEFAULT nextval('public.temp_user_id_seq1'::regclass);


--
-- TOC entry 3510 (class 2604 OID 20624)
-- Name: test_data id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.test_data ALTER COLUMN id SET DEFAULT nextval('public.test_data_id_seq'::regclass);


--
-- TOC entry 3467 (class 2604 OID 20239)
-- Name: user id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."user" ALTER COLUMN id SET DEFAULT nextval('public.user_id_seq1'::regclass);


--
-- TOC entry 3519 (class 2604 OID 20701)
-- Name: user_auth id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_auth ALTER COLUMN id SET DEFAULT nextval('public.user_auth_id_seq'::regclass);


--
-- TOC entry 3521 (class 2604 OID 20721)
-- Name: user_auth_token id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_auth_token ALTER COLUMN id SET DEFAULT nextval('public.user_auth_token_id_seq'::regclass);


--
-- TOC entry 3563 (class 2604 OID 20985)
-- Name: user_role id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_role ALTER COLUMN id SET DEFAULT nextval('public.user_role_id_seq'::regclass);


--
-- TOC entry 3996 (class 0 OID 20526)
-- Dependencies: 245
-- Data for Name: alert; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.alert (id, version, dashboard_id, panel_id, org_id, name, message, state, settings, frequency, handler, severity, silenced, execution_error, eval_data, eval_date, new_state_date, state_changes, created, updated, "for") FROM stdin;
\.


--
-- TOC entry 4036 (class 0 OID 20847)
-- Dependencies: 285
-- Data for Name: alert_configuration; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.alert_configuration (id, alertmanager_configuration, configuration_version, created_at, "default", org_id, configuration_hash) FROM stdin;
1	{\n\t"alertmanager_config": {\n\t\t"route": {\n\t\t\t"receiver": "grafana-default-email",\n\t\t\t"group_by": ["grafana_folder", "alertname"]\n\t\t},\n\t\t"receivers": [{\n\t\t\t"name": "grafana-default-email",\n\t\t\t"grafana_managed_receiver_configs": [{\n\t\t\t\t"uid": "",\n\t\t\t\t"name": "email receiver",\n\t\t\t\t"type": "email",\n\t\t\t\t"isDefault": true,\n\t\t\t\t"settings": {\n\t\t\t\t\t"addresses": "<example@email.com>"\n\t\t\t\t}\n\t\t\t}]\n\t\t}]\n\t}\n}\n	v1	1715798117	t	1	e0528a75784033ae7b15c40851d89484
\.


--
-- TOC entry 4044 (class 0 OID 20891)
-- Dependencies: 293
-- Data for Name: alert_configuration_history; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.alert_configuration_history (id, org_id, alertmanager_configuration, configuration_hash, configuration_version, created_at, "default", last_applied) FROM stdin;
1	1	{\n\t"alertmanager_config": {\n\t\t"route": {\n\t\t\t"receiver": "grafana-default-email",\n\t\t\t"group_by": ["grafana_folder", "alertname"]\n\t\t},\n\t\t"receivers": [{\n\t\t\t"name": "grafana-default-email",\n\t\t\t"grafana_managed_receiver_configs": [{\n\t\t\t\t"uid": "",\n\t\t\t\t"name": "email receiver",\n\t\t\t\t"type": "email",\n\t\t\t\t"isDefault": true,\n\t\t\t\t"settings": {\n\t\t\t\t\t"addresses": "<example@email.com>"\n\t\t\t\t}\n\t\t\t}]\n\t\t}]\n\t}\n}\n	e0528a75784033ae7b15c40851d89484	v1	1715798117	t	1715798273
\.


--
-- TOC entry 4042 (class 0 OID 20881)
-- Dependencies: 291
-- Data for Name: alert_image; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.alert_image (id, token, path, url, created_at, expires_at) FROM stdin;
\.


--
-- TOC entry 4030 (class 0 OID 20792)
-- Dependencies: 279
-- Data for Name: alert_instance; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.alert_instance (rule_org_id, rule_uid, labels, labels_hash, current_state, current_state_since, last_eval_time, current_state_end, current_reason) FROM stdin;
\.


--
-- TOC entry 4000 (class 0 OID 20550)
-- Dependencies: 249
-- Data for Name: alert_notification; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.alert_notification (id, org_id, name, type, settings, created, updated, is_default, frequency, send_reminder, disable_resolve_message, uid, secure_settings) FROM stdin;
\.


--
-- TOC entry 4002 (class 0 OID 20573)
-- Dependencies: 251
-- Data for Name: alert_notification_state; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.alert_notification_state (id, org_id, alert_id, notifier_id, state, version, updated_at, alert_rule_state_updated_version) FROM stdin;
\.


--
-- TOC entry 4032 (class 0 OID 20806)
-- Dependencies: 281
-- Data for Name: alert_rule; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.alert_rule (id, org_id, title, condition, data, updated, interval_seconds, version, uid, namespace_uid, rule_group, no_data_state, exec_err_state, "for", annotations, labels, dashboard_uid, panel_id, rule_group_idx, is_paused) FROM stdin;
\.


--
-- TOC entry 3998 (class 0 OID 20542)
-- Dependencies: 247
-- Data for Name: alert_rule_tag; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.alert_rule_tag (id, alert_id, tag_id) FROM stdin;
\.


--
-- TOC entry 4034 (class 0 OID 20829)
-- Dependencies: 283
-- Data for Name: alert_rule_version; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.alert_rule_version (id, rule_org_id, rule_uid, rule_namespace_uid, rule_group, parent_version, restored_from, version, created, title, condition, data, interval_seconds, no_data_state, exec_err_state, "for", annotations, labels, rule_group_idx, is_paused) FROM stdin;
\.


--
-- TOC entry 4004 (class 0 OID 20584)
-- Dependencies: 253
-- Data for Name: annotation; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.annotation (id, org_id, alert_id, user_id, dashboard_id, panel_id, category_id, type, title, text, metric, prev_state, new_state, data, epoch, region_id, tags, created, updated, epoch_end) FROM stdin;
\.


--
-- TOC entry 4006 (class 0 OID 20604)
-- Dependencies: 255
-- Data for Name: annotation_tag; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.annotation_tag (id, annotation_id, tag_id) FROM stdin;
\.


--
-- TOC entry 3981 (class 0 OID 20426)
-- Dependencies: 230
-- Data for Name: api_key; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.api_key (id, org_id, name, key, role, created, updated, expires, service_account_id, last_used_at, is_revoked) FROM stdin;
\.


--
-- TOC entry 4063 (class 0 OID 20992)
-- Dependencies: 312
-- Data for Name: builtin_role; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.builtin_role (id, role, role_id, created, updated, org_id) FROM stdin;
1	Editor	2	2024-05-15 15:43:22	2024-05-15 15:43:22	1
2	Viewer	3	2024-05-15 15:43:22	2024-05-15 15:43:22	1
\.


--
-- TOC entry 4027 (class 0 OID 20729)
-- Dependencies: 276
-- Data for Name: cache_data; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.cache_data (cache_key, data, expires, created_at) FROM stdin;
auth-proxy-sync-ttl:324a6bd28d83d94f7081447a94746b34	\\x31	3600	1715798281
\.


--
-- TOC entry 4069 (class 0 OID 21053)
-- Dependencies: 318
-- Data for Name: correlation; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.correlation (uid, source_uid, target_uid, label, description, config) FROM stdin;
\.


--
-- TOC entry 3975 (class 0 OID 20332)
-- Dependencies: 224
-- Data for Name: dashboard; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.dashboard (id, version, slug, title, data, org_id, created, updated, updated_by, created_by, gnet_id, plugin_id, folder_id, is_folder, has_acl, uid, is_public) FROM stdin;
2	3	json-scada-history-analog	JSON SCADA History Analog	{"annotations":{"list":[{"builtIn":1,"datasource":{"type":"datasource","uid":"grafana"},"enable":true,"hide":true,"iconColor":"rgba(0, 211, 255, 1)","name":"Annotations \\u0026 Alerts","target":{"limit":100,"matchAny":false,"tags":[],"type":"dashboard"},"type":"dashboard"}]},"editable":true,"fiscalYearStartMonth":0,"graphTooltip":0,"id":2,"links":[],"liveNow":false,"panels":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"fieldConfig":{"defaults":{"color":{"mode":"palette-classic"},"custom":{"axisCenteredZero":false,"axisColorMode":"text","axisLabel":"","axisPlacement":"auto","barAlignment":0,"drawStyle":"line","fillOpacity":10,"gradientMode":"none","hideFrom":{"legend":false,"tooltip":false,"viz":false},"lineInterpolation":"stepAfter","lineWidth":1,"pointSize":4,"scaleDistribution":{"type":"linear"},"showPoints":"always","spanNulls":true,"stacking":{"group":"A","mode":"none"},"thresholdsStyle":{"mode":"off"}},"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]},"unit":"short"},"overrides":[]},"gridPos":{"h":11,"w":24,"x":0,"y":0},"id":2,"options":{"legend":{"calcs":[],"displayMode":"list","placement":"bottom","showLegend":false},"tooltip":{"mode":"single","sort":"none"}},"pluginVersion":"8.0.1","targets":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  \\"time\\" AS \\"time\\",\\n  value\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") and metric IN ('$point_tag')  \\n\\nUNION\\n(\\nSELECT\\n  ($__timeFrom())::timestamp with time zone AS \\"time\\",\\n  value\\nFROM grafana_hist \\nWHERE\\n  time \\u003c (($__timeFrom())::timestamp with time zone) and metric IN ('$point_tag') order by grafana_hist.time desc limit 1\\n)  \\norder by 1 ","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Plot","type":"timeseries"},{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"fieldConfig":{"defaults":{"color":{"mode":"thresholds"},"custom":{"align":"auto","cellOptions":{"type":"auto"},"filterable":true,"inspect":false},"decimals":2,"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]}},"overrides":[{"matcher":{"id":"byName","options":"absv"},"properties":[{"id":"custom.cellOptions","value":{"mode":"basic","type":"gauge"}},{"id":"min","value":0},{"id":"displayName","value":"bar"},{"id":"mappings","value":[{"options":{"from":-1e+22,"result":{"index":0,"text":"."},"to":1e+22},"type":"range"}]},{"id":"decimals","value":0},{"id":"color","value":{"fixedColor":"green","mode":"fixed"}}]},{"matcher":{"id":"byName","options":"time"},"properties":[{"id":"custom.width","value":177}]},{"matcher":{"id":"byName","options":"value"},"properties":[{"id":"custom.width","value":130}]},{"matcher":{"id":"byName","options":"metric"},"properties":[{"id":"custom.width","value":198}]},{"matcher":{"id":"byName","options":"bar"},"properties":[{"id":"custom.width","value":1412}]}]},"gridPos":{"h":14,"w":24,"x":0,"y":11},"id":5,"options":{"cellHeight":"sm","footer":{"countRows":false,"fields":"","reducer":["sum"],"show":false},"showHeader":true,"sortBy":[]},"pluginVersion":"9.5.18","targets":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"format":"table","group":[],"metricColumn":"none","rawQuery":true,"rawSql":"SELECT\\n  metric,\\n  \\"time\\" AS \\"time\\",\\n  value,\\n  abs(value) as absv,\\n  case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") and metric IN ('$point_tag')  \\n\\nUNION\\n(\\nSELECT\\n  metric,\\n  ($__timeFrom())::timestamp with time zone  AS \\"time\\",\\n  value,\\n  abs(value) as absv,\\n  case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist \\nWHERE\\n  time \\u003c (($__timeFrom())::timestamp with time zone) and metric IN ('$point_tag') order by grafana_hist.time desc limit 1\\n)  \\norder by 2 desc\\n","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"}]}],"title":"Table","type":"table"}],"refresh":"10s","schemaVersion":38,"style":"dark","tags":[],"templating":{"list":[ { "current": { "text": "", "value": "" }, "hide": 2, "label": "Tags", "name": "point_tag", "options": [], "query": "", "type": "custom" } ]},"time":{"from":"now-1h","to":"now"},"timepicker":{"refresh_intervals":["10s","30s","1m","5m","15m","30m","1h","2h","1d"]},"timezone":"","title":"JSON SCADA History Analog","uid":"78X6BmvMk","version":3,"weekStart":""}	1	2024-05-15 15:43:49	2024-05-15 15:51:22	1	1	0		0	f	f	78X6BmvMk	f
1	2	json-scada-history-digital	JSON SCADA History Digital	{"annotations":{"list":[{"builtIn":1,"datasource":{"type":"datasource","uid":"grafana"},"enable":true,"hide":true,"iconColor":"rgba(0, 211, 255, 1)","name":"Annotations \\u0026 Alerts","target":{"limit":100,"matchAny":false,"tags":[],"type":"dashboard"},"type":"dashboard"}]},"editable":true,"fiscalYearStartMonth":0,"graphTooltip":0,"id":1,"links":[],"liveNow":false,"panels":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"fieldConfig":{"defaults":{"color":{"mode":"thresholds"},"custom":{"fillOpacity":70,"lineWidth":1,"spanNulls":false},"decimals":0,"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":1}]},"unit":"short"},"overrides":[]},"gridPos":{"h":11,"w":24,"x":0,"y":0},"id":2,"options":{"alignValue":"left","legend":{"displayMode":"list","placement":"bottom","showLegend":false},"mergeValues":false,"rowHeight":0.9,"showValue":"auto","tooltip":{"mode":"single","sort":"none"}},"pluginVersion":"8.0.1","targets":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  -- metric,\\n  \\"time\\" AS \\"time\\",\\n  value,\\n  value_json-\\u003e\\u003e's' as state\\n  -- , case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") and metric IN ('$point_tag')  \\n\\nUNION\\n(\\nSELECT\\n  -- metric,\\n  ($__timeFrom())::timestamp with time zone AS \\"time\\",\\n  value,\\n  value_json-\\u003e\\u003e's' as state\\n  --, case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist \\nWHERE\\n  time \\u003c (($__timeFrom())::timestamp with time zone) and metric IN ('$point_tag') order by grafana_hist.time desc limit 1\\n)  \\norder by 1\\n","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Plot","type":"state-timeline"},{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"fieldConfig":{"defaults":{"color":{"mode":"thresholds"},"custom":{"align":"auto","cellOptions":{"type":"auto"},"filterable":true,"inspect":false},"decimals":2,"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]}},"overrides":[{"matcher":{"id":"byName","options":"time"},"properties":[{"id":"custom.width","value":177}]},{"matcher":{"id":"byName","options":"metric"},"properties":[{"id":"custom.width","value":198}]}]},"gridPos":{"h":14,"w":24,"x":0,"y":11},"id":5,"options":{"cellHeight":"sm","footer":{"countRows":false,"fields":"","reducer":["sum"],"show":false},"showHeader":true,"sortBy":[]},"pluginVersion":"9.5.18","targets":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"format":"table","group":[],"metricColumn":"none","rawQuery":true,"rawSql":"SELECT\\n  metric,\\n  \\"time\\" AS \\"time\\",\\n  value_json-\\u003e\\u003e's' as state,\\n  case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") and metric IN ('$point_tag')  \\n\\nUNION\\n(\\nSELECT\\n  metric,\\n  ($__timeFrom())::timestamp with time zone AS \\"time\\",\\n  value_json-\\u003e\\u003e's' as state,\\n  case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist \\nWHERE\\n  time \\u003c (($__timeFrom())::timestamp with time zone) and metric IN ('$point_tag') order by grafana_hist.time desc limit 1\\n)  \\norder by 2 desc\\n","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"}]}],"title":"Table","type":"table"}],"refresh":"10s","schemaVersion":38,"style":"dark","tags":[],"templating":{"list":[ { "current": { "text": "", "value": "" }, "hide": 2, "label": "Tags", "name": "point_tag", "options": [], "query": "", "type": "custom" } ]},"time":{"from":"now-3h","to":"now"},"timepicker":{"refresh_intervals":["10s","30s","1m","5m","15m","30m","1h","2h","1d"]},"timezone":"","title":"JSON SCADA History Digital","uid":"LsXOaz47z","version":2,"weekStart":""}	1	2024-05-15 15:43:22	2024-05-15 15:52:30	1	1	0		0	f	f	LsXOaz47z	f
3	5	json-scada-filtered-group1-2	JSON SCADA Filtered Group1/2	{"annotations":{"list":[{"builtIn":1,"datasource":{"type":"datasource","uid":"grafana"},"enable":true,"hide":true,"iconColor":"rgba(0, 211, 255, 1)","name":"Annotations \\u0026 Alerts","target":{"limit":100,"matchAny":false,"tags":[],"type":"dashboard"},"type":"dashboard"}]},"editable":true,"fiscalYearStartMonth":0,"graphTooltip":0,"id":3,"links":[],"liveNow":false,"panels":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"fieldConfig":{"defaults":{"color":{"mode":"palette-classic"},"custom":{"axisCenteredZero":false,"axisColorMode":"text","axisLabel":"","axisPlacement":"auto","barAlignment":0,"drawStyle":"line","fillOpacity":10,"gradientMode":"none","hideFrom":{"legend":false,"tooltip":false,"viz":false},"lineInterpolation":"stepAfter","lineWidth":1,"pointSize":5,"scaleDistribution":{"type":"linear"},"showPoints":"never","spanNulls":true,"stacking":{"group":"A","mode":"none"},"thresholdsStyle":{"mode":"off"}},"decimals":1,"links":[],"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]},"unit":"short"},"overrides":[]},"gridPos":{"h":11,"w":20,"x":0,"y":0},"id":2,"options":{"legend":{"calcs":[],"displayMode":"list","placement":"bottom","showLegend":true},"tooltip":{"mode":"single","sort":"none"}},"pluginVersion":"8.0.4","targets":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  \\"time\\" AS \\"time\\",\\n  metric AS metric,\\n  value\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") AND\\n  metric IN ( $point_tag )  \\nORDER BY 1,2","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Metrics","type":"timeseries"},{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"fieldConfig":{"defaults":{"decimals":1,"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]}},"overrides":[]},"gridPos":{"h":11,"w":20,"x":0,"y":11},"id":3,"options":{"orientation":"auto","reduceOptions":{"calcs":["lastNotNull"],"fields":"","values":false},"showThresholdLabels":false,"showThresholdMarkers":true,"text":{}},"pluginVersion":"9.5.18","targets":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  \\"time\\" AS \\"time\\",\\n  metric AS metric,\\n  value\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") AND\\n  metric IN ( $point_tag )  \\nORDER BY 1","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Metrics","type":"gauge"}],"refresh":"10s","schemaVersion":38,"style":"dark","tags":[],"templating":{"list":[{"current":{"selected":false,"text":"_SysMongoDB","value":"_SysMongoDB"},"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"definition":"select distinct json_data-\\u003e\\u003e'group1' as group1 from realtime_data order by 1","hide":0,"includeAll":false,"label":"Group1","multi":true,"name":"group1","options":[],"query":"select distinct json_data-\\u003e\\u003e'group1' as group1 from realtime_data order by 1","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false},{"current":{"selected":false,"text":"_SysMongoDB | WinDev2305Eval","value":"_SysMongoDB | WinDev2305Eval"},"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"definition":"select distinct concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') as group2 from realtime_data where json_data-\\u003e\\u003e'group1' in([[group1]])  order by 1","hide":0,"includeAll":false,"label":"Group2","multi":true,"name":"group2","options":[],"query":"select distinct concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') as group2 from realtime_data where json_data-\\u003e\\u003e'group1' in([[group1]])  order by 1","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false},{"current":{"selected":false,"text":"_SysMongoDB~WinDev2305Eval~active_reads","value":"_SysMongoDB~WinDev2305Eval~active_reads"},"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"definition":"select tag as point_tag from realtime_data where concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') in([[group2]]) order by 1","hide":0,"includeAll":false,"label":"Tags","multi":true,"name":"point_tag","options":[],"query":"select tag as point_tag from realtime_data where concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') in([[group2]]) order by 1","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false}]},"time":{"from":"now-1h","to":"now"},"timepicker":{"refresh_intervals":["10s","30s","1m","5m","15m","30m","1h","2h","1d"]},"timezone":"","title":"JSON SCADA Filtered Group1/2","uid":"zUqcvfZ7z","version":5,"weekStart":""}	1	2024-05-15 15:44:09	2024-05-15 15:59:17	1	1	0		0	f	f	zUqcvfZ7z	f
\.


--
-- TOC entry 4016 (class 0 OID 20658)
-- Dependencies: 265
-- Data for Name: dashboard_acl; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.dashboard_acl (id, org_id, dashboard_id, user_id, team_id, permission, role, created, updated) FROM stdin;
1	-1	-1	\N	\N	1	Viewer	2017-06-20 00:00:00	2017-06-20 00:00:00
2	-1	-1	\N	\N	2	Editor	2017-06-20 00:00:00	2017-06-20 00:00:00
\.


--
-- TOC entry 3977 (class 0 OID 20372)
-- Dependencies: 226
-- Data for Name: dashboard_provisioning; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.dashboard_provisioning (id, dashboard_id, name, external_id, updated, check_sum) FROM stdin;
\.


--
-- TOC entry 4072 (class 0 OID 21089)
-- Dependencies: 321
-- Data for Name: dashboard_public; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.dashboard_public (uid, dashboard_uid, org_id, time_settings, template_variables, access_token, created_by, updated_by, created_at, updated_at, is_enabled, annotations_enabled, time_selection_enabled, share) FROM stdin;
\.


--
-- TOC entry 3983 (class 0 OID 20450)
-- Dependencies: 232
-- Data for Name: dashboard_snapshot; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.dashboard_snapshot (id, name, key, delete_key, org_id, user_id, external, external_url, dashboard, expires, created, updated, external_delete_url, dashboard_encrypted) FROM stdin;
\.


--
-- TOC entry 3973 (class 0 OID 20324)
-- Dependencies: 222
-- Data for Name: dashboard_tag; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.dashboard_tag (id, dashboard_id, term) FROM stdin;
\.


--
-- TOC entry 4010 (class 0 OID 20628)
-- Dependencies: 259
-- Data for Name: dashboard_version; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.dashboard_version (id, dashboard_id, parent_version, restored_from, version, created, created_by, message, data) FROM stdin;
1	1	2	0	1	2024-05-15 15:43:22	1		{"annotations":{"list":[{"builtIn":1,"datasource":{"type":"datasource","uid":"grafana"},"enable":true,"hide":true,"iconColor":"rgba(0, 211, 255, 1)","name":"Annotations \\u0026 Alerts","target":{"limit":100,"matchAny":false,"tags":[],"type":"dashboard"},"type":"dashboard"}]},"editable":true,"fiscalYearStartMonth":0,"graphTooltip":0,"id":null,"links":[],"liveNow":false,"panels":[{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"fieldConfig":{"defaults":{"color":{"mode":"thresholds"},"custom":{"fillOpacity":70,"lineWidth":1,"spanNulls":false},"decimals":0,"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":1}]},"unit":"short"},"overrides":[]},"gridPos":{"h":11,"w":24,"x":0,"y":0},"id":2,"options":{"alignValue":"left","legend":{"displayMode":"list","placement":"bottom","showLegend":false},"mergeValues":false,"rowHeight":0.9,"showValue":"auto","tooltip":{"mode":"single","sort":"none"}},"pluginVersion":"8.0.1","targets":[{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  -- metric,\\n  \\"time\\" AS \\"time\\",\\n  value,\\n  value_json-\\u003e\\u003e's' as state\\n  -- , case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") and metric IN ('$point_tag')  \\n\\nUNION\\n(\\nSELECT\\n  -- metric,\\n  ($__timeFrom())::timestamp with time zone AS \\"time\\",\\n  value,\\n  value_json-\\u003e\\u003e's' as state\\n  --, case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist \\nWHERE\\n  time \\u003c (($__timeFrom())::timestamp with time zone) and metric IN ('$point_tag') order by grafana_hist.time desc limit 1\\n)  \\norder by 1\\n","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Plot","type":"state-timeline"},{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"fieldConfig":{"defaults":{"color":{"mode":"thresholds"},"custom":{"align":"auto","cellOptions":{"type":"auto"},"filterable":true,"inspect":false},"decimals":2,"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]}},"overrides":[{"matcher":{"id":"byName","options":"time"},"properties":[{"id":"custom.width","value":177}]},{"matcher":{"id":"byName","options":"metric"},"properties":[{"id":"custom.width","value":198}]}]},"gridPos":{"h":14,"w":24,"x":0,"y":11},"id":5,"options":{"cellHeight":"sm","footer":{"countRows":false,"fields":"","reducer":["sum"],"show":false},"showHeader":true,"sortBy":[]},"pluginVersion":"9.5.18","targets":[{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"format":"table","group":[],"metricColumn":"none","rawQuery":true,"rawSql":"SELECT\\n  metric,\\n  \\"time\\" AS \\"time\\",\\n  value_json-\\u003e\\u003e's' as state,\\n  case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") and metric IN ('$point_tag')  \\n\\nUNION\\n(\\nSELECT\\n  metric,\\n  ($__timeFrom())::timestamp with time zone AS \\"time\\",\\n  value_json-\\u003e\\u003e's' as state,\\n  case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist \\nWHERE\\n  time \\u003c (($__timeFrom())::timestamp with time zone) and metric IN ('$point_tag') order by grafana_hist.time desc limit 1\\n)  \\norder by 2 desc\\n","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"}]}],"title":"Table","type":"table"}],"refresh":"10s","schemaVersion":38,"style":"dark","tags":[],"templating":{"list":[{"hide":2,"label":"Tags","name":"point_tag","query":"KAW2KPR21MTVA--------C","skipUrlSync":false,"type":"constant"}]},"time":{"from":"now-3h","to":"now"},"timepicker":{"refresh_intervals":["10s","30s","1m","5m","15m","30m","1h","2h","1d"]},"timezone":"","title":"JSON SCADA History Digital","uid":"LsXOaz47z","version":1,"weekStart":""}
2	2	1	0	1	2024-05-15 15:43:49	1		{"annotations":{"list":[{"builtIn":1,"datasource":{"type":"datasource","uid":"grafana"},"enable":true,"hide":true,"iconColor":"rgba(0, 211, 255, 1)","name":"Annotations \\u0026 Alerts","target":{"limit":100,"matchAny":false,"tags":[],"type":"dashboard"},"type":"dashboard"}]},"editable":true,"fiscalYearStartMonth":0,"graphTooltip":0,"id":null,"links":[],"liveNow":false,"panels":[{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"fieldConfig":{"defaults":{"color":{"mode":"palette-classic"},"custom":{"axisCenteredZero":false,"axisColorMode":"text","axisLabel":"","axisPlacement":"auto","barAlignment":0,"drawStyle":"line","fillOpacity":10,"gradientMode":"none","hideFrom":{"legend":false,"tooltip":false,"viz":false},"lineInterpolation":"stepAfter","lineWidth":1,"pointSize":4,"scaleDistribution":{"type":"linear"},"showPoints":"always","spanNulls":true,"stacking":{"group":"A","mode":"none"},"thresholdsStyle":{"mode":"off"}},"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]},"unit":"short"},"overrides":[]},"gridPos":{"h":11,"w":24,"x":0,"y":0},"id":2,"options":{"legend":{"calcs":[],"displayMode":"list","placement":"bottom","showLegend":false},"tooltip":{"mode":"single","sort":"none"}},"pluginVersion":"8.0.1","targets":[{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  \\"time\\" AS \\"time\\",\\n  value\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") and metric IN ('$point_tag')  \\n\\nUNION\\n(\\nSELECT\\n  ($__timeFrom())::timestamp with time zone AS \\"time\\",\\n  value\\nFROM grafana_hist \\nWHERE\\n  time \\u003c (($__timeFrom())::timestamp with time zone) and metric IN ('$point_tag') order by grafana_hist.time desc limit 1\\n)  \\norder by 1 ","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Plot","type":"timeseries"},{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"fieldConfig":{"defaults":{"color":{"mode":"thresholds"},"custom":{"align":"auto","cellOptions":{"type":"auto"},"filterable":true,"inspect":false},"decimals":2,"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]}},"overrides":[{"matcher":{"id":"byName","options":"absv"},"properties":[{"id":"custom.cellOptions","value":{"mode":"basic","type":"gauge"}},{"id":"min","value":0},{"id":"displayName","value":"bar"},{"id":"mappings","value":[{"options":{"from":-1e+22,"result":{"index":0,"text":"."},"to":1e+22},"type":"range"}]},{"id":"decimals","value":0},{"id":"color","value":{"fixedColor":"green","mode":"fixed"}}]},{"matcher":{"id":"byName","options":"time"},"properties":[{"id":"custom.width","value":177}]},{"matcher":{"id":"byName","options":"value"},"properties":[{"id":"custom.width","value":130}]},{"matcher":{"id":"byName","options":"metric"},"properties":[{"id":"custom.width","value":198}]},{"matcher":{"id":"byName","options":"bar"},"properties":[{"id":"custom.width","value":1412}]}]},"gridPos":{"h":14,"w":24,"x":0,"y":11},"id":5,"options":{"cellHeight":"sm","footer":{"countRows":false,"fields":"","reducer":["sum"],"show":false},"showHeader":true,"sortBy":[]},"pluginVersion":"9.5.18","targets":[{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"format":"table","group":[],"metricColumn":"none","rawQuery":true,"rawSql":"SELECT\\n  metric,\\n  \\"time\\" AS \\"time\\",\\n  value,\\n  abs(value) as absv,\\n  case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") and metric IN ('$point_tag')  \\n\\nUNION\\n(\\nSELECT\\n  metric,\\n  ($__timeFrom())::timestamp with time zone  AS \\"time\\",\\n  value,\\n  abs(value) as absv,\\n  case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist \\nWHERE\\n  time \\u003c (($__timeFrom())::timestamp with time zone) and metric IN ('$point_tag') order by grafana_hist.time desc limit 1\\n)  \\norder by 2 desc\\n","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"}]}],"title":"Table","type":"table"}],"refresh":"10s","schemaVersion":38,"style":"dark","tags":[],"templating":{"list":[{"hide":2,"label":"Tags","name":"point_tag","query":"KAW2KPR21MTVA--------C","skipUrlSync":false,"type":"constant"}]},"time":{"from":"now-1h","to":"now"},"timepicker":{"refresh_intervals":["10s","30s","1m","5m","15m","30m","1h","2h","1d"]},"timezone":"","title":"JSON SCADA History Analog","uid":"78X6BmvMk","version":1,"weekStart":""}
3	3	1	0	1	2024-05-15 15:44:09	1		{"annotations":{"list":[{"builtIn":1,"datasource":{"type":"datasource","uid":"grafana"},"enable":true,"hide":true,"iconColor":"rgba(0, 211, 255, 1)","name":"Annotations \\u0026 Alerts","target":{"limit":100,"matchAny":false,"tags":[],"type":"dashboard"},"type":"dashboard"}]},"editable":true,"fiscalYearStartMonth":0,"graphTooltip":0,"id":null,"links":[],"liveNow":false,"panels":[{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"fieldConfig":{"defaults":{"color":{"mode":"palette-classic"},"custom":{"axisCenteredZero":false,"axisColorMode":"text","axisLabel":"","axisPlacement":"auto","barAlignment":0,"drawStyle":"line","fillOpacity":10,"gradientMode":"none","hideFrom":{"legend":false,"tooltip":false,"viz":false},"lineInterpolation":"stepAfter","lineWidth":1,"pointSize":5,"scaleDistribution":{"type":"linear"},"showPoints":"never","spanNulls":true,"stacking":{"group":"A","mode":"none"},"thresholdsStyle":{"mode":"off"}},"decimals":1,"links":[],"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]},"unit":"short"},"overrides":[]},"gridPos":{"h":11,"w":20,"x":0,"y":0},"id":2,"options":{"legend":{"calcs":[],"displayMode":"list","placement":"bottom","showLegend":true},"tooltip":{"mode":"single","sort":"none"}},"pluginVersion":"8.0.4","targets":[{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  \\"time\\" AS \\"time\\",\\n  metric AS metric,\\n  value\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") AND\\n  metric IN ( $point_tag )  \\nORDER BY 1,2","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Metrics","type":"timeseries"},{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"fieldConfig":{"defaults":{"decimals":1,"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]}},"overrides":[]},"gridPos":{"h":11,"w":20,"x":0,"y":11},"id":3,"options":{"orientation":"auto","reduceOptions":{"calcs":["lastNotNull"],"fields":"","values":false},"showThresholdLabels":false,"showThresholdMarkers":true,"text":{}},"pluginVersion":"9.5.18","targets":[{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  \\"time\\" AS \\"time\\",\\n  metric AS metric,\\n  value\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") AND\\n  metric IN ( $point_tag )  \\nORDER BY 1","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Metrics","type":"gauge"}],"refresh":"10s","schemaVersion":38,"style":"dark","tags":[],"templating":{"list":[{"current":{"selected":false,"text":"_SysMongoDB","value":"_SysMongoDB"},"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"definition":"select distinct json_data-\\u003e\\u003e'group1' as group1 from realtime_data order by 1","hide":0,"includeAll":false,"label":"Group1","multi":true,"name":"group1","options":[],"query":"select distinct json_data-\\u003e\\u003e'group1' as group1 from realtime_data order by 1","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false},{"current":{"selected":false,"text":"_SysMongoDB | WinDev2305Eval","value":"_SysMongoDB | WinDev2305Eval"},"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"definition":"select distinct concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') as group2 from realtime_data where json_data-\\u003e\\u003e'group1' in([[group1]])  order by 1","hide":0,"includeAll":false,"label":"Group2","multi":true,"name":"group2","options":[],"query":"select distinct concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') as group2 from realtime_data where json_data-\\u003e\\u003e'group1' in([[group1]])  order by 1","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false},{"current":{"selected":false,"text":"_SysMongoDB~WinDev2305Eval~active_reads","value":"_SysMongoDB~WinDev2305Eval~active_reads"},"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"definition":"select tag as point_tag from realtime_data where concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') in([[group2]]) order by 1","hide":0,"includeAll":false,"label":"Tags","multi":true,"name":"point_tag","options":[],"query":"select tag as point_tag from realtime_data where concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') in([[group2]]) order by 1","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false}]},"time":{"from":"now-1h","to":"now"},"timepicker":{"refresh_intervals":["10s","30s","1m","5m","15m","30m","1h","2h","1d"]},"timezone":"","title":"JSON SCADA Filtered Group1/2","uid":"zUqcvfZ7z","version":1,"weekStart":""}
4	2	1	0	2	2024-05-15 15:46:37	1		{"annotations":{"list":[{"builtIn":1,"datasource":{"type":"datasource","uid":"grafana"},"enable":true,"hide":true,"iconColor":"rgba(0, 211, 255, 1)","name":"Annotations \\u0026 Alerts","target":{"limit":100,"matchAny":false,"tags":[],"type":"dashboard"},"type":"dashboard"}]},"editable":true,"fiscalYearStartMonth":0,"graphTooltip":0,"id":2,"links":[],"liveNow":false,"panels":[{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"fieldConfig":{"defaults":{"color":{"mode":"palette-classic"},"custom":{"axisCenteredZero":false,"axisColorMode":"text","axisLabel":"","axisPlacement":"auto","barAlignment":0,"drawStyle":"line","fillOpacity":10,"gradientMode":"none","hideFrom":{"legend":false,"tooltip":false,"viz":false},"lineInterpolation":"stepAfter","lineWidth":1,"pointSize":4,"scaleDistribution":{"type":"linear"},"showPoints":"always","spanNulls":true,"stacking":{"group":"A","mode":"none"},"thresholdsStyle":{"mode":"off"}},"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]},"unit":"short"},"overrides":[]},"gridPos":{"h":11,"w":24,"x":0,"y":0},"id":2,"options":{"legend":{"calcs":[],"displayMode":"list","placement":"bottom","showLegend":false},"tooltip":{"mode":"single","sort":"none"}},"pluginVersion":"8.0.1","targets":[{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  \\"time\\" AS \\"time\\",\\n  value\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") and metric IN ('$point_tag')  \\n\\nUNION\\n(\\nSELECT\\n  ($__timeFrom())::timestamp with time zone AS \\"time\\",\\n  value\\nFROM grafana_hist \\nWHERE\\n  time \\u003c (($__timeFrom())::timestamp with time zone) and metric IN ('$point_tag') order by grafana_hist.time desc limit 1\\n)  \\norder by 1 ","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Plot","type":"timeseries"},{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"fieldConfig":{"defaults":{"color":{"mode":"thresholds"},"custom":{"align":"auto","cellOptions":{"type":"auto"},"filterable":true,"inspect":false},"decimals":2,"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]}},"overrides":[{"matcher":{"id":"byName","options":"absv"},"properties":[{"id":"custom.cellOptions","value":{"mode":"basic","type":"gauge"}},{"id":"min","value":0},{"id":"displayName","value":"bar"},{"id":"mappings","value":[{"options":{"from":-1e+22,"result":{"index":0,"text":"."},"to":1e+22},"type":"range"}]},{"id":"decimals","value":0},{"id":"color","value":{"fixedColor":"green","mode":"fixed"}}]},{"matcher":{"id":"byName","options":"time"},"properties":[{"id":"custom.width","value":177}]},{"matcher":{"id":"byName","options":"value"},"properties":[{"id":"custom.width","value":130}]},{"matcher":{"id":"byName","options":"metric"},"properties":[{"id":"custom.width","value":198}]},{"matcher":{"id":"byName","options":"bar"},"properties":[{"id":"custom.width","value":1412}]}]},"gridPos":{"h":14,"w":24,"x":0,"y":11},"id":5,"options":{"cellHeight":"sm","footer":{"countRows":false,"fields":"","reducer":["sum"],"show":false},"showHeader":true,"sortBy":[]},"pluginVersion":"9.5.18","targets":[{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"format":"table","group":[],"metricColumn":"none","rawQuery":true,"rawSql":"SELECT\\n  metric,\\n  \\"time\\" AS \\"time\\",\\n  value,\\n  abs(value) as absv,\\n  case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") and metric IN ('$point_tag')  \\n\\nUNION\\n(\\nSELECT\\n  metric,\\n  ($__timeFrom())::timestamp with time zone  AS \\"time\\",\\n  value,\\n  abs(value) as absv,\\n  case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist \\nWHERE\\n  time \\u003c (($__timeFrom())::timestamp with time zone) and metric IN ('$point_tag') order by grafana_hist.time desc limit 1\\n)  \\norder by 2 desc\\n","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"}]}],"title":"Table","type":"table"}],"refresh":"10s","schemaVersion":38,"style":"dark","tags":[],"templating":{"list":[{"hide":2,"label":"Tags","name":"point_tag","query":"KAW2KPR21MTVA--------C","skipUrlSync":false,"type":"constant"}]},"time":{"from":"now-1h","to":"now"},"timepicker":{"refresh_intervals":["10s","30s","1m","5m","15m","30m","1h","2h","1d"]},"timezone":"","title":"JSON SCADA History Analog","uid":"78X6BmvMk","version":2,"weekStart":""}
5	3	1	0	2	2024-05-15 15:49:28	1		{"annotations":{"list":[{"builtIn":1,"datasource":{"type":"datasource","uid":"grafana"},"enable":true,"hide":true,"iconColor":"rgba(0, 211, 255, 1)","name":"Annotations \\u0026 Alerts","target":{"limit":100,"matchAny":false,"tags":[],"type":"dashboard"},"type":"dashboard"}]},"editable":true,"fiscalYearStartMonth":0,"graphTooltip":0,"id":3,"links":[],"liveNow":false,"panels":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"fieldConfig":{"defaults":{"color":{"mode":"palette-classic"},"custom":{"axisCenteredZero":false,"axisColorMode":"text","axisLabel":"","axisPlacement":"auto","barAlignment":0,"drawStyle":"line","fillOpacity":10,"gradientMode":"none","hideFrom":{"legend":false,"tooltip":false,"viz":false},"lineInterpolation":"stepAfter","lineWidth":1,"pointSize":5,"scaleDistribution":{"type":"linear"},"showPoints":"never","spanNulls":true,"stacking":{"group":"A","mode":"none"},"thresholdsStyle":{"mode":"off"}},"decimals":1,"links":[],"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]},"unit":"short"},"overrides":[]},"gridPos":{"h":11,"w":20,"x":0,"y":0},"id":2,"options":{"legend":{"calcs":[],"displayMode":"list","placement":"bottom","showLegend":true},"tooltip":{"mode":"single","sort":"none"}},"pluginVersion":"8.0.4","targets":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  \\"time\\" AS \\"time\\",\\n  metric AS metric,\\n  value\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") AND\\n  metric IN ( $point_tag )  \\nORDER BY 1,2","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Metrics","type":"timeseries"},{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"fieldConfig":{"defaults":{"decimals":1,"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]}},"overrides":[]},"gridPos":{"h":11,"w":20,"x":0,"y":11},"id":3,"options":{"orientation":"auto","reduceOptions":{"calcs":["lastNotNull"],"fields":"","values":false},"showThresholdLabels":false,"showThresholdMarkers":true,"text":{}},"pluginVersion":"9.5.18","targets":[{"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  \\"time\\" AS \\"time\\",\\n  metric AS metric,\\n  value\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") AND\\n  metric IN ( $point_tag )  \\nORDER BY 1","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Metrics","type":"gauge"}],"refresh":"10s","schemaVersion":38,"style":"dark","tags":[],"templating":{"list":[{"current":{"selected":false,"text":"_SysMongoDB","value":"_SysMongoDB"},"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"definition":"select distinct json_data-\\u003e\\u003e'group1' as group1 from realtime_data order by 1","hide":0,"includeAll":false,"label":"Group1","multi":true,"name":"group1","options":[],"query":"select distinct json_data-\\u003e\\u003e'group1' as group1 from realtime_data order by 1","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false},{"current":{"selected":false,"text":"_SysMongoDB | WinDev2305Eval","value":"_SysMongoDB | WinDev2305Eval"},"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"definition":"select distinct concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') as group2 from realtime_data where json_data-\\u003e\\u003e'group1' in([[group1]])  order by 1","hide":0,"includeAll":false,"label":"Group2","multi":true,"name":"group2","options":[],"query":"select distinct concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') as group2 from realtime_data where json_data-\\u003e\\u003e'group1' in([[group1]])  order by 1","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false},{"current":{"selected":false,"text":"_SysMongoDB~WinDev2305Eval~active_reads","value":"_SysMongoDB~WinDev2305Eval~active_reads"},"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"definition":"select tag as point_tag from realtime_data where concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') in([[group2]]) order by 1","hide":0,"includeAll":false,"label":"Tags","multi":true,"name":"point_tag","options":[],"query":"select tag as point_tag from realtime_data where concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') in([[group2]]) order by 1","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false}]},"time":{"from":"now-1h","to":"now"},"timepicker":{"refresh_intervals":["10s","30s","1m","5m","15m","30m","1h","2h","1d"]},"timezone":"","title":"JSON SCADA Filtered Group1/2","uid":"zUqcvfZ7z","version":2,"weekStart":""}
6	2	2	0	3	2024-05-15 15:51:22	1		{"annotations":{"list":[{"builtIn":1,"datasource":{"type":"datasource","uid":"grafana"},"enable":true,"hide":true,"iconColor":"rgba(0, 211, 255, 1)","name":"Annotations \\u0026 Alerts","target":{"limit":100,"matchAny":false,"tags":[],"type":"dashboard"},"type":"dashboard"}]},"editable":true,"fiscalYearStartMonth":0,"graphTooltip":0,"id":2,"links":[],"liveNow":false,"panels":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"fieldConfig":{"defaults":{"color":{"mode":"palette-classic"},"custom":{"axisCenteredZero":false,"axisColorMode":"text","axisLabel":"","axisPlacement":"auto","barAlignment":0,"drawStyle":"line","fillOpacity":10,"gradientMode":"none","hideFrom":{"legend":false,"tooltip":false,"viz":false},"lineInterpolation":"stepAfter","lineWidth":1,"pointSize":4,"scaleDistribution":{"type":"linear"},"showPoints":"always","spanNulls":true,"stacking":{"group":"A","mode":"none"},"thresholdsStyle":{"mode":"off"}},"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]},"unit":"short"},"overrides":[]},"gridPos":{"h":11,"w":24,"x":0,"y":0},"id":2,"options":{"legend":{"calcs":[],"displayMode":"list","placement":"bottom","showLegend":false},"tooltip":{"mode":"single","sort":"none"}},"pluginVersion":"8.0.1","targets":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  \\"time\\" AS \\"time\\",\\n  value\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") and metric IN ('$point_tag')  \\n\\nUNION\\n(\\nSELECT\\n  ($__timeFrom())::timestamp with time zone AS \\"time\\",\\n  value\\nFROM grafana_hist \\nWHERE\\n  time \\u003c (($__timeFrom())::timestamp with time zone) and metric IN ('$point_tag') order by grafana_hist.time desc limit 1\\n)  \\norder by 1 ","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Plot","type":"timeseries"},{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"fieldConfig":{"defaults":{"color":{"mode":"thresholds"},"custom":{"align":"auto","cellOptions":{"type":"auto"},"filterable":true,"inspect":false},"decimals":2,"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]}},"overrides":[{"matcher":{"id":"byName","options":"absv"},"properties":[{"id":"custom.cellOptions","value":{"mode":"basic","type":"gauge"}},{"id":"min","value":0},{"id":"displayName","value":"bar"},{"id":"mappings","value":[{"options":{"from":-1e+22,"result":{"index":0,"text":"."},"to":1e+22},"type":"range"}]},{"id":"decimals","value":0},{"id":"color","value":{"fixedColor":"green","mode":"fixed"}}]},{"matcher":{"id":"byName","options":"time"},"properties":[{"id":"custom.width","value":177}]},{"matcher":{"id":"byName","options":"value"},"properties":[{"id":"custom.width","value":130}]},{"matcher":{"id":"byName","options":"metric"},"properties":[{"id":"custom.width","value":198}]},{"matcher":{"id":"byName","options":"bar"},"properties":[{"id":"custom.width","value":1412}]}]},"gridPos":{"h":14,"w":24,"x":0,"y":11},"id":5,"options":{"cellHeight":"sm","footer":{"countRows":false,"fields":"","reducer":["sum"],"show":false},"showHeader":true,"sortBy":[]},"pluginVersion":"9.5.18","targets":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"format":"table","group":[],"metricColumn":"none","rawQuery":true,"rawSql":"SELECT\\n  metric,\\n  \\"time\\" AS \\"time\\",\\n  value,\\n  abs(value) as absv,\\n  case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") and metric IN ('$point_tag')  \\n\\nUNION\\n(\\nSELECT\\n  metric,\\n  ($__timeFrom())::timestamp with time zone  AS \\"time\\",\\n  value,\\n  abs(value) as absv,\\n  case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist \\nWHERE\\n  time \\u003c (($__timeFrom())::timestamp with time zone) and metric IN ('$point_tag') order by grafana_hist.time desc limit 1\\n)  \\norder by 2 desc\\n","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"}]}],"title":"Table","type":"table"}],"refresh":"10s","schemaVersion":38,"style":"dark","tags":[],"templating":{"list":[{"hide":2,"label":"Tags","name":"point_tag","query":"KAW2KPR21MTVA--------C","skipUrlSync":false,"type":"constant"}]},"time":{"from":"now-1h","to":"now"},"timepicker":{"refresh_intervals":["10s","30s","1m","5m","15m","30m","1h","2h","1d"]},"timezone":"","title":"JSON SCADA History Analog","uid":"78X6BmvMk","version":3,"weekStart":""}
7	1	1	0	2	2024-05-15 15:52:30	1		{"annotations":{"list":[{"builtIn":1,"datasource":{"type":"datasource","uid":"grafana"},"enable":true,"hide":true,"iconColor":"rgba(0, 211, 255, 1)","name":"Annotations \\u0026 Alerts","target":{"limit":100,"matchAny":false,"tags":[],"type":"dashboard"},"type":"dashboard"}]},"editable":true,"fiscalYearStartMonth":0,"graphTooltip":0,"id":1,"links":[],"liveNow":false,"panels":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"fieldConfig":{"defaults":{"color":{"mode":"thresholds"},"custom":{"fillOpacity":70,"lineWidth":1,"spanNulls":false},"decimals":0,"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":1}]},"unit":"short"},"overrides":[]},"gridPos":{"h":11,"w":24,"x":0,"y":0},"id":2,"options":{"alignValue":"left","legend":{"displayMode":"list","placement":"bottom","showLegend":false},"mergeValues":false,"rowHeight":0.9,"showValue":"auto","tooltip":{"mode":"single","sort":"none"}},"pluginVersion":"8.0.1","targets":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  -- metric,\\n  \\"time\\" AS \\"time\\",\\n  value,\\n  value_json-\\u003e\\u003e's' as state\\n  -- , case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") and metric IN ('$point_tag')  \\n\\nUNION\\n(\\nSELECT\\n  -- metric,\\n  ($__timeFrom())::timestamp with time zone AS \\"time\\",\\n  value,\\n  value_json-\\u003e\\u003e's' as state\\n  --, case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist \\nWHERE\\n  time \\u003c (($__timeFrom())::timestamp with time zone) and metric IN ('$point_tag') order by grafana_hist.time desc limit 1\\n)  \\norder by 1\\n","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Plot","type":"state-timeline"},{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"fieldConfig":{"defaults":{"color":{"mode":"thresholds"},"custom":{"align":"auto","cellOptions":{"type":"auto"},"filterable":true,"inspect":false},"decimals":2,"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]}},"overrides":[{"matcher":{"id":"byName","options":"time"},"properties":[{"id":"custom.width","value":177}]},{"matcher":{"id":"byName","options":"metric"},"properties":[{"id":"custom.width","value":198}]}]},"gridPos":{"h":14,"w":24,"x":0,"y":11},"id":5,"options":{"cellHeight":"sm","footer":{"countRows":false,"fields":"","reducer":["sum"],"show":false},"showHeader":true,"sortBy":[]},"pluginVersion":"9.5.18","targets":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"format":"table","group":[],"metricColumn":"none","rawQuery":true,"rawSql":"SELECT\\n  metric,\\n  \\"time\\" AS \\"time\\",\\n  value_json-\\u003e\\u003e's' as state,\\n  case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") and metric IN ('$point_tag')  \\n\\nUNION\\n(\\nSELECT\\n  metric,\\n  ($__timeFrom())::timestamp with time zone AS \\"time\\",\\n  value_json-\\u003e\\u003e's' as state,\\n  case when (flags \\u0026 B'10000000') = B'10000000' then 'F' else '' end as flags\\nFROM grafana_hist \\nWHERE\\n  time \\u003c (($__timeFrom())::timestamp with time zone) and metric IN ('$point_tag') order by grafana_hist.time desc limit 1\\n)  \\norder by 2 desc\\n","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"}]}],"title":"Table","type":"table"}],"refresh":"10s","schemaVersion":38,"style":"dark","tags":[],"templating":{"list":[{"hide":2,"label":"Tags","name":"point_tag","query":"KAW2KPR21MTVA--------C","skipUrlSync":false,"type":"constant"}]},"time":{"from":"now-3h","to":"now"},"timepicker":{"refresh_intervals":["10s","30s","1m","5m","15m","30m","1h","2h","1d"]},"timezone":"","title":"JSON SCADA History Digital","uid":"LsXOaz47z","version":2,"weekStart":""}
8	3	2	0	3	2024-05-15 15:54:15	1		{"annotations":{"list":[{"builtIn":1,"datasource":{"type":"datasource","uid":"grafana"},"enable":true,"hide":true,"iconColor":"rgba(0, 211, 255, 1)","name":"Annotations \\u0026 Alerts","target":{"limit":100,"matchAny":false,"tags":[],"type":"dashboard"},"type":"dashboard"}]},"editable":true,"fiscalYearStartMonth":0,"graphTooltip":0,"id":3,"links":[],"liveNow":false,"panels":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"fieldConfig":{"defaults":{"color":{"mode":"palette-classic"},"custom":{"axisCenteredZero":false,"axisColorMode":"text","axisLabel":"","axisPlacement":"auto","barAlignment":0,"drawStyle":"line","fillOpacity":10,"gradientMode":"none","hideFrom":{"legend":false,"tooltip":false,"viz":false},"lineInterpolation":"stepAfter","lineWidth":1,"pointSize":5,"scaleDistribution":{"type":"linear"},"showPoints":"never","spanNulls":true,"stacking":{"group":"A","mode":"none"},"thresholdsStyle":{"mode":"off"}},"decimals":1,"links":[],"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]},"unit":"short"},"overrides":[]},"gridPos":{"h":11,"w":20,"x":0,"y":0},"id":2,"options":{"legend":{"calcs":[],"displayMode":"list","placement":"bottom","showLegend":true},"tooltip":{"mode":"single","sort":"none"}},"pluginVersion":"8.0.4","targets":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  \\"time\\" AS \\"time\\",\\n  metric AS metric,\\n  value\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") AND\\n  metric IN ( $point_tag )  \\nORDER BY 1,2","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Metrics","type":"timeseries"},{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"fieldConfig":{"defaults":{"decimals":1,"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]}},"overrides":[]},"gridPos":{"h":11,"w":20,"x":0,"y":11},"id":3,"options":{"orientation":"auto","reduceOptions":{"calcs":["lastNotNull"],"fields":"","values":false},"showThresholdLabels":false,"showThresholdMarkers":true,"text":{}},"pluginVersion":"9.5.18","targets":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  \\"time\\" AS \\"time\\",\\n  metric AS metric,\\n  value\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") AND\\n  metric IN ( $point_tag )  \\nORDER BY 1","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Metrics","type":"gauge"}],"refresh":"10s","schemaVersion":38,"style":"dark","tags":[],"templating":{"list":[{"current":{"selected":false,"text":"_SysMongoDB","value":"_SysMongoDB"},"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"definition":"select distinct json_data-\\u003e\\u003e'group1' as group1 from realtime_data order by 1","hide":0,"includeAll":false,"label":"Group1","multi":true,"name":"group1","options":[],"query":"select distinct json_data-\\u003e\\u003e'group1' as group1 from realtime_data order by 1","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false},{"current":{"selected":false,"text":"_SysMongoDB | WinDev2305Eval","value":"_SysMongoDB | WinDev2305Eval"},"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"definition":"select distinct concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') as group2 from realtime_data where json_data-\\u003e\\u003e'group1' in([[group1]])  order by 1","hide":0,"includeAll":false,"label":"Group2","multi":true,"name":"group2","options":[],"query":"select distinct concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') as group2 from realtime_data where json_data-\\u003e\\u003e'group1' in([[group1]])  order by 1","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false},{"current":{"selected":false,"text":"_SysMongoDB~WinDev2305Eval~active_reads","value":"_SysMongoDB~WinDev2305Eval~active_reads"},"datasource":{"type":"postgres","uid":"4kXFdV4Gk"},"definition":"select tag as point_tag from realtime_data where concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') in([[group2]]) order by 1","hide":0,"includeAll":false,"label":"Tags","multi":true,"name":"point_tag","options":[],"query":"select tag as point_tag from realtime_data where concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') in([[group2]]) order by 1","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false}]},"time":{"from":"now-1h","to":"now"},"timepicker":{"refresh_intervals":["10s","30s","1m","5m","15m","30m","1h","2h","1d"]},"timezone":"","title":"JSON SCADA Filtered Group1/2","uid":"zUqcvfZ7z","version":3,"weekStart":""}
9	3	3	0	4	2024-05-15 15:55:43	1		{"annotations":{"list":[{"builtIn":1,"datasource":{"type":"datasource","uid":"grafana"},"enable":true,"hide":true,"iconColor":"rgba(0, 211, 255, 1)","name":"Annotations \\u0026 Alerts","target":{"limit":100,"matchAny":false,"tags":[],"type":"dashboard"},"type":"dashboard"}]},"editable":true,"fiscalYearStartMonth":0,"graphTooltip":0,"id":3,"links":[],"liveNow":false,"panels":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"fieldConfig":{"defaults":{"color":{"mode":"palette-classic"},"custom":{"axisCenteredZero":false,"axisColorMode":"text","axisLabel":"","axisPlacement":"auto","barAlignment":0,"drawStyle":"line","fillOpacity":10,"gradientMode":"none","hideFrom":{"legend":false,"tooltip":false,"viz":false},"lineInterpolation":"stepAfter","lineWidth":1,"pointSize":5,"scaleDistribution":{"type":"linear"},"showPoints":"never","spanNulls":true,"stacking":{"group":"A","mode":"none"},"thresholdsStyle":{"mode":"off"}},"decimals":1,"links":[],"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]},"unit":"short"},"overrides":[]},"gridPos":{"h":11,"w":20,"x":0,"y":0},"id":2,"options":{"legend":{"calcs":[],"displayMode":"list","placement":"bottom","showLegend":true},"tooltip":{"mode":"single","sort":"none"}},"pluginVersion":"8.0.4","targets":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  \\"time\\" AS \\"time\\",\\n  metric AS metric,\\n  value\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") AND\\n  metric IN ( $point_tag )  \\nORDER BY 1,2","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Metrics","type":"timeseries"},{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"fieldConfig":{"defaults":{"decimals":1,"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]}},"overrides":[]},"gridPos":{"h":11,"w":20,"x":0,"y":11},"id":3,"options":{"orientation":"auto","reduceOptions":{"calcs":["lastNotNull"],"fields":"","values":false},"showThresholdLabels":false,"showThresholdMarkers":true,"text":{}},"pluginVersion":"9.5.18","targets":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  \\"time\\" AS \\"time\\",\\n  metric AS metric,\\n  value\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") AND\\n  metric IN ( $point_tag )  \\nORDER BY 1","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Metrics","type":"gauge"}],"refresh":"10s","schemaVersion":38,"style":"dark","tags":[],"templating":{"list":[{"current":{"selected":false,"text":"_SysMongoDB","value":"_SysMongoDB"},"definition":"select distinct json_data-\\u003e\\u003e'group1' as group1 from realtime_data order by 1","hide":0,"includeAll":false,"label":"Group1","multi":true,"name":"group1","options":[],"query":"select distinct json_data-\\u003e\\u003e'group1' as group1 from realtime_data order by 1","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false},{"current":{"selected":false,"text":"_SysMongoDB | WinDev2305Eval","value":"_SysMongoDB | WinDev2305Eval"},"definition":"select distinct concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') as group2 from realtime_data where json_data-\\u003e\\u003e'group1' in([[group1]])  order by 1","hide":0,"includeAll":false,"label":"Group2","multi":true,"name":"group2","options":[],"query":"select distinct concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') as group2 from realtime_data where json_data-\\u003e\\u003e'group1' in([[group1]])  order by 1","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false},{"current":{"selected":false,"text":"_SysMongoDB~WinDev2305Eval~active_reads","value":"_SysMongoDB~WinDev2305Eval~active_reads"},"definition":"select tag as point_tag from realtime_data where concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') in([[group2]]) order by 1","hide":0,"includeAll":false,"label":"Tags","multi":true,"name":"point_tag","options":[],"query":"","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false}]},"time":{"from":"now-1h","to":"now"},"timepicker":{"refresh_intervals":["10s","30s","1m","5m","15m","30m","1h","2h","1d"]},"timezone":"","title":"JSON SCADA Filtered Group1/2","uid":"zUqcvfZ7z","version":4,"weekStart":""}
10	3	4	0	5	2024-05-15 15:59:17	1		{"annotations":{"list":[{"builtIn":1,"datasource":{"type":"datasource","uid":"grafana"},"enable":true,"hide":true,"iconColor":"rgba(0, 211, 255, 1)","name":"Annotations \\u0026 Alerts","target":{"limit":100,"matchAny":false,"tags":[],"type":"dashboard"},"type":"dashboard"}]},"editable":true,"fiscalYearStartMonth":0,"graphTooltip":0,"id":3,"links":[],"liveNow":false,"panels":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"fieldConfig":{"defaults":{"color":{"mode":"palette-classic"},"custom":{"axisCenteredZero":false,"axisColorMode":"text","axisLabel":"","axisPlacement":"auto","barAlignment":0,"drawStyle":"line","fillOpacity":10,"gradientMode":"none","hideFrom":{"legend":false,"tooltip":false,"viz":false},"lineInterpolation":"stepAfter","lineWidth":1,"pointSize":5,"scaleDistribution":{"type":"linear"},"showPoints":"never","spanNulls":true,"stacking":{"group":"A","mode":"none"},"thresholdsStyle":{"mode":"off"}},"decimals":1,"links":[],"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]},"unit":"short"},"overrides":[]},"gridPos":{"h":11,"w":20,"x":0,"y":0},"id":2,"options":{"legend":{"calcs":[],"displayMode":"list","placement":"bottom","showLegend":true},"tooltip":{"mode":"single","sort":"none"}},"pluginVersion":"8.0.4","targets":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  \\"time\\" AS \\"time\\",\\n  metric AS metric,\\n  value\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") AND\\n  metric IN ( $point_tag )  \\nORDER BY 1,2","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Metrics","type":"timeseries"},{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"fieldConfig":{"defaults":{"decimals":1,"mappings":[],"thresholds":{"mode":"absolute","steps":[{"color":"green","value":null},{"color":"red","value":80}]}},"overrides":[]},"gridPos":{"h":11,"w":20,"x":0,"y":11},"id":3,"options":{"orientation":"auto","reduceOptions":{"calcs":["lastNotNull"],"fields":"","values":false},"showThresholdLabels":false,"showThresholdMarkers":true,"text":{}},"pluginVersion":"9.5.18","targets":[{"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"format":"time_series","group":[],"metricColumn":"metric","rawQuery":true,"rawSql":"SELECT\\n  \\"time\\" AS \\"time\\",\\n  metric AS metric,\\n  value\\nFROM grafana_hist\\nWHERE\\n  $__timeFilter(\\"time\\") AND\\n  metric IN ( $point_tag )  \\nORDER BY 1","refId":"A","select":[[{"params":["value"],"type":"column"}]],"table":"grafana_hist","timeColumn":"\\"time\\"","timeColumnType":"timestamp","where":[{"name":"$__timeFilter","params":[],"type":"macro"},{"datatype":"text","name":"","params":["metric","IN","$point_tag"],"type":"expression"}]}],"title":"Metrics","type":"gauge"}],"refresh":"10s","schemaVersion":38,"style":"dark","tags":[],"templating":{"list":[{"current":{"selected":false,"text":"_SysMongoDB","value":"_SysMongoDB"},"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"definition":"select distinct json_data-\\u003e\\u003e'group1' as group1 from realtime_data order by 1","hide":0,"includeAll":false,"label":"Group1","multi":true,"name":"group1","options":[],"query":"select distinct json_data-\\u003e\\u003e'group1' as group1 from realtime_data order by 1","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false},{"current":{"selected":false,"text":"_SysMongoDB | WinDev2305Eval","value":"_SysMongoDB | WinDev2305Eval"},"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"definition":"select distinct concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') as group2 from realtime_data where json_data-\\u003e\\u003e'group1' in([[group1]])  order by 1","hide":0,"includeAll":false,"label":"Group2","multi":true,"name":"group2","options":[],"query":"select distinct concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') as group2 from realtime_data where json_data-\\u003e\\u003e'group1' in([[group1]])  order by 1","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false},{"current":{"selected":false,"text":"_SysMongoDB~WinDev2305Eval~active_reads","value":"_SysMongoDB~WinDev2305Eval~active_reads"},"datasource":{"type":"postgres","uid":"f4dfcd5e-956e-4186-b0ed-abc6496b08b0"},"definition":"select tag as point_tag from realtime_data where concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') in([[group2]]) order by 1","hide":0,"includeAll":false,"label":"Tags","multi":true,"name":"point_tag","options":[],"query":"select tag as point_tag from realtime_data where concat(json_data-\\u003e\\u003e'group1', ' | ', json_data-\\u003e\\u003e'group2') in([[group2]]) order by 1","refresh":1,"regex":"","skipUrlSync":false,"sort":0,"tagValuesQuery":"","tagsQuery":"","type":"query","useTags":false}]},"time":{"from":"now-1h","to":"now"},"timepicker":{"refresh_intervals":["10s","30s","1m","5m","15m","30m","1h","2h","1d"]},"timezone":"","title":"JSON SCADA Filtered Group1/2","uid":"zUqcvfZ7z","version":5,"weekStart":""}
\.


--
-- TOC entry 4049 (class 0 OID 20923)
-- Dependencies: 298
-- Data for Name: data_keys; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.data_keys (name, active, scope, provider, encrypted_data, created, updated, label) FROM stdin;
d4dff665-e7b4-4b78-a8b3-43fe9dfabbc8	t	root	secretKey.v1	\\x2a5957567a4c574e6d59672a5431785a3461444a1abe195482b461b217da603ad9e5036547dcf797744b7fb8fcdb3d073da14c36	2024-05-15 15:38:36	2024-05-15 15:38:36	2024-05-15/root@secretKey.v1
\.


--
-- TOC entry 3979 (class 0 OID 20398)
-- Dependencies: 228
-- Data for Name: data_source; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.data_source (id, org_id, version, type, name, access, url, password, "user", database, basic_auth, basic_auth_user, basic_auth_password, is_default, json_data, created, updated, with_credentials, secure_json_data, read_only, uid) FROM stdin;
1	1	2	postgres	PostgreSQL-JSON_SCADA	proxy	127.0.0.1:5432		grafana		f			t	{"connMaxLifetime":14400,"database":"json_scada","maxIdleConns":100,"maxIdleConnsAuto":true,"maxOpenConns":100,"postgresVersion":1400,"sslmode":"disable","timescaledb":true}	2024-05-15 15:38:36	2024-05-15 15:42:11	f	{}	f	f4dfcd5e-956e-4186-b0ed-abc6496b08b0
\.


--
-- TOC entry 4071 (class 0 OID 21063)
-- Dependencies: 320
-- Data for Name: entity_event; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.entity_event (id, entity_id, event_type, created) FROM stdin;
\.


--
-- TOC entry 4073 (class 0 OID 21103)
-- Dependencies: 322
-- Data for Name: file; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.file (path, path_hash, parent_folder_path_hash, contents, etag, cache_control, content_disposition, updated, created, size, mime_type) FROM stdin;
\.


--
-- TOC entry 4074 (class 0 OID 21110)
-- Dependencies: 323
-- Data for Name: file_meta; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.file_meta (path_hash, key, value) FROM stdin;
\.


--
-- TOC entry 4077 (class 0 OID 21135)
-- Dependencies: 326
-- Data for Name: folder; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.folder (id, uid, org_id, title, description, parent_uid, created, updated) FROM stdin;
\.


--
-- TOC entry 4053 (class 0 OID 20941)
-- Dependencies: 302
-- Data for Name: kv_store; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.kv_store (id, org_id, namespace, key, value, created, updated) FROM stdin;
1	0	datasource	secretMigrationStatus	compatible	2024-05-15 15:35:35	2024-05-15 15:35:35
2	1	alertmanager	notifications		2024-05-15 15:37:09	2024-05-15 15:37:09
3	1	alertmanager	silences		2024-05-15 15:37:09	2024-05-15 15:37:09
\.


--
-- TOC entry 4046 (class 0 OID 20905)
-- Dependencies: 295
-- Data for Name: library_element; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.library_element (id, org_id, folder_id, uid, name, kind, type, description, model, created, created_by, updated, updated_by, version) FROM stdin;
\.


--
-- TOC entry 4048 (class 0 OID 20915)
-- Dependencies: 297
-- Data for Name: library_element_connection; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.library_element_connection (id, element_id, kind, connection_id, created, created_by) FROM stdin;
\.


--
-- TOC entry 4020 (class 0 OID 20689)
-- Dependencies: 269
-- Data for Name: login_attempt; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.login_attempt (id, username, ip_address, created) FROM stdin;
\.


--
-- TOC entry 3961 (class 0 OID 20216)
-- Dependencies: 210
-- Data for Name: migration_log; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.migration_log (id, migration_id, sql, success, error, "timestamp") FROM stdin;
1	create migration_log table	CREATE TABLE IF NOT EXISTS "migration_log" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "migration_id" VARCHAR(255) NOT NULL\n, "sql" TEXT NOT NULL\n, "success" BOOL NOT NULL\n, "error" TEXT NOT NULL\n, "timestamp" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:40
2	create user table	CREATE TABLE IF NOT EXISTS "user" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "version" INTEGER NOT NULL\n, "login" VARCHAR(190) NOT NULL\n, "email" VARCHAR(190) NOT NULL\n, "name" VARCHAR(255) NULL\n, "password" VARCHAR(255) NULL\n, "salt" VARCHAR(50) NULL\n, "rands" VARCHAR(50) NULL\n, "company" VARCHAR(255) NULL\n, "account_id" BIGINT NOT NULL\n, "is_admin" BOOL NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:41
3	add unique index user.login	CREATE UNIQUE INDEX "UQE_user_login" ON "user" ("login");	t		2024-05-15 15:34:41
4	add unique index user.email	CREATE UNIQUE INDEX "UQE_user_email" ON "user" ("email");	t		2024-05-15 15:34:41
5	drop index UQE_user_login - v1	DROP INDEX "UQE_user_login" CASCADE	t		2024-05-15 15:34:41
6	drop index UQE_user_email - v1	DROP INDEX "UQE_user_email" CASCADE	t		2024-05-15 15:34:41
7	Rename table user to user_v1 - v1	ALTER TABLE "user" RENAME TO "user_v1"	t		2024-05-15 15:34:41
8	create user table v2	CREATE TABLE IF NOT EXISTS "user" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "version" INTEGER NOT NULL\n, "login" VARCHAR(190) NOT NULL\n, "email" VARCHAR(190) NOT NULL\n, "name" VARCHAR(255) NULL\n, "password" VARCHAR(255) NULL\n, "salt" VARCHAR(50) NULL\n, "rands" VARCHAR(50) NULL\n, "company" VARCHAR(255) NULL\n, "org_id" BIGINT NOT NULL\n, "is_admin" BOOL NOT NULL\n, "email_verified" BOOL NULL\n, "theme" VARCHAR(255) NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:41
9	create index UQE_user_login - v2	CREATE UNIQUE INDEX "UQE_user_login" ON "user" ("login");	t		2024-05-15 15:34:41
10	create index UQE_user_email - v2	CREATE UNIQUE INDEX "UQE_user_email" ON "user" ("email");	t		2024-05-15 15:34:41
11	copy data_source v1 to v2	INSERT INTO "user" ("updated"\n, "salt"\n, "org_id"\n, "is_admin"\n, "created"\n, "email"\n, "name"\n, "password"\n, "version"\n, "id"\n, "login"\n, "rands"\n, "company") SELECT "updated"\n, "salt"\n, "account_id"\n, "is_admin"\n, "created"\n, "email"\n, "name"\n, "password"\n, "version"\n, "id"\n, "login"\n, "rands"\n, "company" FROM "user_v1"	t		2024-05-15 15:34:41
12	Drop old table user_v1	DROP TABLE IF EXISTS "user_v1"	t		2024-05-15 15:34:41
13	Add column help_flags1 to user table	alter table "user" ADD COLUMN "help_flags1" BIGINT NOT NULL DEFAULT 0 	t		2024-05-15 15:34:41
14	Update user table charset	ALTER TABLE "user" ALTER "login" TYPE VARCHAR(190), ALTER "email" TYPE VARCHAR(190), ALTER "name" TYPE VARCHAR(255), ALTER "password" TYPE VARCHAR(255), ALTER "salt" TYPE VARCHAR(50), ALTER "rands" TYPE VARCHAR(50), ALTER "company" TYPE VARCHAR(255), ALTER "theme" TYPE VARCHAR(255);	t		2024-05-15 15:34:41
15	Add last_seen_at column to user	alter table "user" ADD COLUMN "last_seen_at" TIMESTAMP NULL 	t		2024-05-15 15:34:41
16	Add missing user data	code migration	t		2024-05-15 15:34:41
17	Add is_disabled column to user	alter table "user" ADD COLUMN "is_disabled" BOOL NOT NULL DEFAULT false 	t		2024-05-15 15:34:41
18	Add index user.login/user.email	CREATE INDEX "IDX_user_login_email" ON "user" ("login","email");	t		2024-05-15 15:34:41
19	Add is_service_account column to user	alter table "user" ADD COLUMN "is_service_account" BOOL NOT NULL DEFAULT false 	t		2024-05-15 15:34:41
20	Update is_service_account column to nullable	ALTER TABLE `user` ALTER COLUMN is_service_account DROP NOT NULL;	t		2024-05-15 15:34:41
21	create temp user table v1-7	CREATE TABLE IF NOT EXISTS "temp_user" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "version" INTEGER NOT NULL\n, "email" VARCHAR(190) NOT NULL\n, "name" VARCHAR(255) NULL\n, "role" VARCHAR(20) NULL\n, "code" VARCHAR(190) NOT NULL\n, "status" VARCHAR(20) NOT NULL\n, "invited_by_user_id" BIGINT NULL\n, "email_sent" BOOL NOT NULL\n, "email_sent_on" TIMESTAMP NULL\n, "remote_addr" VARCHAR(255) NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:41
22	create index IDX_temp_user_email - v1-7	CREATE INDEX "IDX_temp_user_email" ON "temp_user" ("email");	t		2024-05-15 15:34:41
23	create index IDX_temp_user_org_id - v1-7	CREATE INDEX "IDX_temp_user_org_id" ON "temp_user" ("org_id");	t		2024-05-15 15:34:41
24	create index IDX_temp_user_code - v1-7	CREATE INDEX "IDX_temp_user_code" ON "temp_user" ("code");	t		2024-05-15 15:34:41
25	create index IDX_temp_user_status - v1-7	CREATE INDEX "IDX_temp_user_status" ON "temp_user" ("status");	t		2024-05-15 15:34:41
26	Update temp_user table charset	ALTER TABLE "temp_user" ALTER "email" TYPE VARCHAR(190), ALTER "name" TYPE VARCHAR(255), ALTER "role" TYPE VARCHAR(20), ALTER "code" TYPE VARCHAR(190), ALTER "status" TYPE VARCHAR(20), ALTER "remote_addr" TYPE VARCHAR(255);	t		2024-05-15 15:34:41
27	drop index IDX_temp_user_email - v1	DROP INDEX "IDX_temp_user_email" CASCADE	t		2024-05-15 15:34:41
28	drop index IDX_temp_user_org_id - v1	DROP INDEX "IDX_temp_user_org_id" CASCADE	t		2024-05-15 15:34:41
29	drop index IDX_temp_user_code - v1	DROP INDEX "IDX_temp_user_code" CASCADE	t		2024-05-15 15:34:42
30	drop index IDX_temp_user_status - v1	DROP INDEX "IDX_temp_user_status" CASCADE	t		2024-05-15 15:34:42
31	Rename table temp_user to temp_user_tmp_qwerty - v1	ALTER TABLE "temp_user" RENAME TO "temp_user_tmp_qwerty"	t		2024-05-15 15:34:42
32	create temp_user v2	CREATE TABLE IF NOT EXISTS "temp_user" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "version" INTEGER NOT NULL\n, "email" VARCHAR(190) NOT NULL\n, "name" VARCHAR(255) NULL\n, "role" VARCHAR(20) NULL\n, "code" VARCHAR(190) NOT NULL\n, "status" VARCHAR(20) NOT NULL\n, "invited_by_user_id" BIGINT NULL\n, "email_sent" BOOL NOT NULL\n, "email_sent_on" TIMESTAMP NULL\n, "remote_addr" VARCHAR(255) NULL\n, "created" INTEGER NOT NULL DEFAULT 0\n, "updated" INTEGER NOT NULL DEFAULT 0\n);	t		2024-05-15 15:34:42
33	create index IDX_temp_user_email - v2	CREATE INDEX "IDX_temp_user_email" ON "temp_user" ("email");	t		2024-05-15 15:34:42
34	create index IDX_temp_user_org_id - v2	CREATE INDEX "IDX_temp_user_org_id" ON "temp_user" ("org_id");	t		2024-05-15 15:34:42
35	create index IDX_temp_user_code - v2	CREATE INDEX "IDX_temp_user_code" ON "temp_user" ("code");	t		2024-05-15 15:34:42
36	create index IDX_temp_user_status - v2	CREATE INDEX "IDX_temp_user_status" ON "temp_user" ("status");	t		2024-05-15 15:34:42
37	copy temp_user v1 to v2	INSERT INTO "temp_user" ("id"\n, "version"\n, "name"\n, "invited_by_user_id"\n, "status"\n, "email_sent"\n, "email_sent_on"\n, "remote_addr"\n, "org_id"\n, "email"\n, "role"\n, "code") SELECT "id"\n, "version"\n, "name"\n, "invited_by_user_id"\n, "status"\n, "email_sent"\n, "email_sent_on"\n, "remote_addr"\n, "org_id"\n, "email"\n, "role"\n, "code" FROM "temp_user_tmp_qwerty"	t		2024-05-15 15:34:42
38	drop temp_user_tmp_qwerty	DROP TABLE IF EXISTS "temp_user_tmp_qwerty"	t		2024-05-15 15:34:42
39	Set created for temp users that will otherwise prematurely expire	code migration	t		2024-05-15 15:34:42
165	Add column team_id in preferences	alter table "preferences" ADD COLUMN "team_id" BIGINT NULL 	t		2024-05-15 15:34:46
40	create star table	CREATE TABLE IF NOT EXISTS "star" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "user_id" BIGINT NOT NULL\n, "dashboard_id" BIGINT NOT NULL\n);	t		2024-05-15 15:34:42
41	add unique index star.user_id_dashboard_id	CREATE UNIQUE INDEX "UQE_star_user_id_dashboard_id" ON "star" ("user_id","dashboard_id");	t		2024-05-15 15:34:42
42	create org table v1	CREATE TABLE IF NOT EXISTS "org" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "version" INTEGER NOT NULL\n, "name" VARCHAR(190) NOT NULL\n, "address1" VARCHAR(255) NULL\n, "address2" VARCHAR(255) NULL\n, "city" VARCHAR(255) NULL\n, "state" VARCHAR(255) NULL\n, "zip_code" VARCHAR(50) NULL\n, "country" VARCHAR(255) NULL\n, "billing_email" VARCHAR(255) NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:42
43	create index UQE_org_name - v1	CREATE UNIQUE INDEX "UQE_org_name" ON "org" ("name");	t		2024-05-15 15:34:42
44	create org_user table v1	CREATE TABLE IF NOT EXISTS "org_user" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "user_id" BIGINT NOT NULL\n, "role" VARCHAR(20) NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:42
45	create index IDX_org_user_org_id - v1	CREATE INDEX "IDX_org_user_org_id" ON "org_user" ("org_id");	t		2024-05-15 15:34:42
46	create index UQE_org_user_org_id_user_id - v1	CREATE UNIQUE INDEX "UQE_org_user_org_id_user_id" ON "org_user" ("org_id","user_id");	t		2024-05-15 15:34:42
47	create index IDX_org_user_user_id - v1	CREATE INDEX "IDX_org_user_user_id" ON "org_user" ("user_id");	t		2024-05-15 15:34:42
48	Update org table charset	ALTER TABLE "org" ALTER "name" TYPE VARCHAR(190), ALTER "address1" TYPE VARCHAR(255), ALTER "address2" TYPE VARCHAR(255), ALTER "city" TYPE VARCHAR(255), ALTER "state" TYPE VARCHAR(255), ALTER "zip_code" TYPE VARCHAR(50), ALTER "country" TYPE VARCHAR(255), ALTER "billing_email" TYPE VARCHAR(255);	t		2024-05-15 15:34:42
49	Update org_user table charset	ALTER TABLE "org_user" ALTER "role" TYPE VARCHAR(20);	t		2024-05-15 15:34:42
50	Migrate all Read Only Viewers to Viewers	UPDATE org_user SET role = 'Viewer' WHERE role = 'Read Only Editor'	t		2024-05-15 15:34:42
51	create dashboard table	CREATE TABLE IF NOT EXISTS "dashboard" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "version" INTEGER NOT NULL\n, "slug" VARCHAR(189) NOT NULL\n, "title" VARCHAR(255) NOT NULL\n, "data" TEXT NOT NULL\n, "account_id" BIGINT NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:42
52	add index dashboard.account_id	CREATE INDEX "IDX_dashboard_account_id" ON "dashboard" ("account_id");	t		2024-05-15 15:34:42
53	add unique index dashboard_account_id_slug	CREATE UNIQUE INDEX "UQE_dashboard_account_id_slug" ON "dashboard" ("account_id","slug");	t		2024-05-15 15:34:42
54	create dashboard_tag table	CREATE TABLE IF NOT EXISTS "dashboard_tag" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "dashboard_id" BIGINT NOT NULL\n, "term" VARCHAR(50) NOT NULL\n);	t		2024-05-15 15:34:42
55	add unique index dashboard_tag.dasboard_id_term	CREATE UNIQUE INDEX "UQE_dashboard_tag_dashboard_id_term" ON "dashboard_tag" ("dashboard_id","term");	t		2024-05-15 15:34:42
56	drop index UQE_dashboard_tag_dashboard_id_term - v1	DROP INDEX "UQE_dashboard_tag_dashboard_id_term" CASCADE	t		2024-05-15 15:34:43
57	Rename table dashboard to dashboard_v1 - v1	ALTER TABLE "dashboard" RENAME TO "dashboard_v1"	t		2024-05-15 15:34:43
58	create dashboard v2	CREATE TABLE IF NOT EXISTS "dashboard" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "version" INTEGER NOT NULL\n, "slug" VARCHAR(189) NOT NULL\n, "title" VARCHAR(255) NOT NULL\n, "data" TEXT NOT NULL\n, "org_id" BIGINT NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:43
59	create index IDX_dashboard_org_id - v2	CREATE INDEX "IDX_dashboard_org_id" ON "dashboard" ("org_id");	t		2024-05-15 15:34:43
60	create index UQE_dashboard_org_id_slug - v2	CREATE UNIQUE INDEX "UQE_dashboard_org_id_slug" ON "dashboard" ("org_id","slug");	t		2024-05-15 15:34:43
61	copy dashboard v1 to v2	INSERT INTO "dashboard" ("id"\n, "version"\n, "slug"\n, "title"\n, "data"\n, "org_id"\n, "created"\n, "updated") SELECT "id"\n, "version"\n, "slug"\n, "title"\n, "data"\n, "account_id"\n, "created"\n, "updated" FROM "dashboard_v1"	t		2024-05-15 15:34:43
62	drop table dashboard_v1	DROP TABLE IF EXISTS "dashboard_v1"	t		2024-05-15 15:34:43
63	alter dashboard.data to mediumtext v1	SELECT 0;	t		2024-05-15 15:34:43
64	Add column updated_by in dashboard - v2	alter table "dashboard" ADD COLUMN "updated_by" INTEGER NULL 	t		2024-05-15 15:34:43
65	Add column created_by in dashboard - v2	alter table "dashboard" ADD COLUMN "created_by" INTEGER NULL 	t		2024-05-15 15:34:43
66	Add column gnetId in dashboard	alter table "dashboard" ADD COLUMN "gnet_id" BIGINT NULL 	t		2024-05-15 15:34:43
67	Add index for gnetId in dashboard	CREATE INDEX "IDX_dashboard_gnet_id" ON "dashboard" ("gnet_id");	t		2024-05-15 15:34:43
68	Add column plugin_id in dashboard	alter table "dashboard" ADD COLUMN "plugin_id" VARCHAR(189) NULL 	t		2024-05-15 15:34:43
69	Add index for plugin_id in dashboard	CREATE INDEX "IDX_dashboard_org_id_plugin_id" ON "dashboard" ("org_id","plugin_id");	t		2024-05-15 15:34:43
70	Add index for dashboard_id in dashboard_tag	CREATE INDEX "IDX_dashboard_tag_dashboard_id" ON "dashboard_tag" ("dashboard_id");	t		2024-05-15 15:34:43
71	Update dashboard table charset	ALTER TABLE "dashboard" ALTER "slug" TYPE VARCHAR(189), ALTER "title" TYPE VARCHAR(255), ALTER "plugin_id" TYPE VARCHAR(189), ALTER "data" TYPE TEXT;	t		2024-05-15 15:34:43
72	Update dashboard_tag table charset	ALTER TABLE "dashboard_tag" ALTER "term" TYPE VARCHAR(50);	t		2024-05-15 15:34:43
73	Add column folder_id in dashboard	alter table "dashboard" ADD COLUMN "folder_id" BIGINT NOT NULL DEFAULT 0 	t		2024-05-15 15:34:43
74	Add column isFolder in dashboard	alter table "dashboard" ADD COLUMN "is_folder" BOOL NOT NULL DEFAULT false 	t		2024-05-15 15:34:43
75	Add column has_acl in dashboard	alter table "dashboard" ADD COLUMN "has_acl" BOOL NOT NULL DEFAULT false 	t		2024-05-15 15:34:43
76	Add column uid in dashboard	alter table "dashboard" ADD COLUMN "uid" VARCHAR(40) NULL 	t		2024-05-15 15:34:43
77	Update uid column values in dashboard	UPDATE dashboard SET uid=lpad('' || id::text,9,'0') WHERE uid IS NULL;	t		2024-05-15 15:34:43
78	Add unique index dashboard_org_id_uid	CREATE UNIQUE INDEX "UQE_dashboard_org_id_uid" ON "dashboard" ("org_id","uid");	t		2024-05-15 15:34:43
79	Remove unique index org_id_slug	DROP INDEX "UQE_dashboard_org_id_slug" CASCADE	t		2024-05-15 15:34:43
80	Update dashboard title length	ALTER TABLE "dashboard" ALTER "title" TYPE VARCHAR(189);	t		2024-05-15 15:34:44
81	Add unique index for dashboard_org_id_title_folder_id	CREATE UNIQUE INDEX "UQE_dashboard_org_id_folder_id_title" ON "dashboard" ("org_id","folder_id","title");	t		2024-05-15 15:34:44
82	create dashboard_provisioning	CREATE TABLE IF NOT EXISTS "dashboard_provisioning" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "dashboard_id" BIGINT NULL\n, "name" VARCHAR(150) NOT NULL\n, "external_id" TEXT NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:44
83	Rename table dashboard_provisioning to dashboard_provisioning_tmp_qwerty - v1	ALTER TABLE "dashboard_provisioning" RENAME TO "dashboard_provisioning_tmp_qwerty"	t		2024-05-15 15:34:44
84	create dashboard_provisioning v2	CREATE TABLE IF NOT EXISTS "dashboard_provisioning" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "dashboard_id" BIGINT NULL\n, "name" VARCHAR(150) NOT NULL\n, "external_id" TEXT NOT NULL\n, "updated" INTEGER NOT NULL DEFAULT 0\n);	t		2024-05-15 15:34:44
85	create index IDX_dashboard_provisioning_dashboard_id - v2	CREATE INDEX "IDX_dashboard_provisioning_dashboard_id" ON "dashboard_provisioning" ("dashboard_id");	t		2024-05-15 15:34:44
86	create index IDX_dashboard_provisioning_dashboard_id_name - v2	CREATE INDEX "IDX_dashboard_provisioning_dashboard_id_name" ON "dashboard_provisioning" ("dashboard_id","name");	t		2024-05-15 15:34:44
87	copy dashboard_provisioning v1 to v2	INSERT INTO "dashboard_provisioning" ("name"\n, "external_id"\n, "id"\n, "dashboard_id") SELECT "name"\n, "external_id"\n, "id"\n, "dashboard_id" FROM "dashboard_provisioning_tmp_qwerty"	t		2024-05-15 15:34:44
88	drop dashboard_provisioning_tmp_qwerty	DROP TABLE IF EXISTS "dashboard_provisioning_tmp_qwerty"	t		2024-05-15 15:34:44
89	Add check_sum column	alter table "dashboard_provisioning" ADD COLUMN "check_sum" VARCHAR(32) NULL 	t		2024-05-15 15:34:44
90	Add index for dashboard_title	CREATE INDEX "IDX_dashboard_title" ON "dashboard" ("title");	t		2024-05-15 15:34:44
91	delete tags for deleted dashboards	DELETE FROM dashboard_tag WHERE dashboard_id NOT IN (SELECT id FROM dashboard)	t		2024-05-15 15:34:44
92	delete stars for deleted dashboards	DELETE FROM star WHERE dashboard_id NOT IN (SELECT id FROM dashboard)	t		2024-05-15 15:34:44
93	Add index for dashboard_is_folder	CREATE INDEX "IDX_dashboard_is_folder" ON "dashboard" ("is_folder");	t		2024-05-15 15:34:44
94	Add isPublic for dashboard	alter table "dashboard" ADD COLUMN "is_public" BOOL NOT NULL DEFAULT false 	t		2024-05-15 15:34:44
95	create data_source table	CREATE TABLE IF NOT EXISTS "data_source" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "account_id" BIGINT NOT NULL\n, "version" INTEGER NOT NULL\n, "type" VARCHAR(255) NOT NULL\n, "name" VARCHAR(190) NOT NULL\n, "access" VARCHAR(255) NOT NULL\n, "url" VARCHAR(255) NOT NULL\n, "password" VARCHAR(255) NULL\n, "user" VARCHAR(255) NULL\n, "database" VARCHAR(255) NULL\n, "basic_auth" BOOL NOT NULL\n, "basic_auth_user" VARCHAR(255) NULL\n, "basic_auth_password" VARCHAR(255) NULL\n, "is_default" BOOL NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:44
96	add index data_source.account_id	CREATE INDEX "IDX_data_source_account_id" ON "data_source" ("account_id");	t		2024-05-15 15:34:44
97	add unique index data_source.account_id_name	CREATE UNIQUE INDEX "UQE_data_source_account_id_name" ON "data_source" ("account_id","name");	t		2024-05-15 15:34:44
98	drop index IDX_data_source_account_id - v1	DROP INDEX "IDX_data_source_account_id" CASCADE	t		2024-05-15 15:34:44
99	drop index UQE_data_source_account_id_name - v1	DROP INDEX "UQE_data_source_account_id_name" CASCADE	t		2024-05-15 15:34:44
100	Rename table data_source to data_source_v1 - v1	ALTER TABLE "data_source" RENAME TO "data_source_v1"	t		2024-05-15 15:34:44
101	create data_source table v2	CREATE TABLE IF NOT EXISTS "data_source" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "version" INTEGER NOT NULL\n, "type" VARCHAR(255) NOT NULL\n, "name" VARCHAR(190) NOT NULL\n, "access" VARCHAR(255) NOT NULL\n, "url" VARCHAR(255) NOT NULL\n, "password" VARCHAR(255) NULL\n, "user" VARCHAR(255) NULL\n, "database" VARCHAR(255) NULL\n, "basic_auth" BOOL NOT NULL\n, "basic_auth_user" VARCHAR(255) NULL\n, "basic_auth_password" VARCHAR(255) NULL\n, "is_default" BOOL NOT NULL\n, "json_data" TEXT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:44
102	create index IDX_data_source_org_id - v2	CREATE INDEX "IDX_data_source_org_id" ON "data_source" ("org_id");	t		2024-05-15 15:34:44
103	create index UQE_data_source_org_id_name - v2	CREATE UNIQUE INDEX "UQE_data_source_org_id_name" ON "data_source" ("org_id","name");	t		2024-05-15 15:34:44
104	Drop old table data_source_v1 #2	DROP TABLE IF EXISTS "data_source_v1"	t		2024-05-15 15:34:44
105	Add column with_credentials	alter table "data_source" ADD COLUMN "with_credentials" BOOL NOT NULL DEFAULT false 	t		2024-05-15 15:34:44
106	Add secure json data column	alter table "data_source" ADD COLUMN "secure_json_data" TEXT NULL 	t		2024-05-15 15:34:44
107	Update data_source table charset	ALTER TABLE "data_source" ALTER "type" TYPE VARCHAR(255), ALTER "name" TYPE VARCHAR(190), ALTER "access" TYPE VARCHAR(255), ALTER "url" TYPE VARCHAR(255), ALTER "password" TYPE VARCHAR(255), ALTER "user" TYPE VARCHAR(255), ALTER "database" TYPE VARCHAR(255), ALTER "basic_auth_user" TYPE VARCHAR(255), ALTER "basic_auth_password" TYPE VARCHAR(255), ALTER "json_data" TYPE TEXT, ALTER "secure_json_data" TYPE TEXT;	t		2024-05-15 15:34:44
108	Update initial version to 1	UPDATE data_source SET version = 1 WHERE version = 0	t		2024-05-15 15:34:44
109	Add read_only data column	alter table "data_source" ADD COLUMN "read_only" BOOL NULL 	t		2024-05-15 15:34:44
110	Migrate logging ds to loki ds	UPDATE data_source SET type = 'loki' WHERE type = 'logging'	t		2024-05-15 15:34:44
111	Update json_data with nulls	UPDATE data_source SET json_data = '{}' WHERE json_data is null	t		2024-05-15 15:34:44
112	Add uid column	alter table "data_source" ADD COLUMN "uid" VARCHAR(40) NOT NULL DEFAULT 0 	t		2024-05-15 15:34:44
113	Update uid value	UPDATE data_source SET uid=lpad('' || id::text,9,'0');	t		2024-05-15 15:34:44
114	Add unique index datasource_org_id_uid	CREATE UNIQUE INDEX "UQE_data_source_org_id_uid" ON "data_source" ("org_id","uid");	t		2024-05-15 15:34:44
115	add unique index datasource_org_id_is_default	CREATE INDEX "IDX_data_source_org_id_is_default" ON "data_source" ("org_id","is_default");	t		2024-05-15 15:34:44
116	create api_key table	CREATE TABLE IF NOT EXISTS "api_key" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "account_id" BIGINT NOT NULL\n, "name" VARCHAR(190) NOT NULL\n, "key" VARCHAR(64) NOT NULL\n, "role" VARCHAR(255) NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:44
117	add index api_key.account_id	CREATE INDEX "IDX_api_key_account_id" ON "api_key" ("account_id");	t		2024-05-15 15:34:45
118	add index api_key.key	CREATE UNIQUE INDEX "UQE_api_key_key" ON "api_key" ("key");	t		2024-05-15 15:34:45
119	add index api_key.account_id_name	CREATE UNIQUE INDEX "UQE_api_key_account_id_name" ON "api_key" ("account_id","name");	t		2024-05-15 15:34:45
120	drop index IDX_api_key_account_id - v1	DROP INDEX "IDX_api_key_account_id" CASCADE	t		2024-05-15 15:34:45
121	drop index UQE_api_key_key - v1	DROP INDEX "UQE_api_key_key" CASCADE	t		2024-05-15 15:34:45
122	drop index UQE_api_key_account_id_name - v1	DROP INDEX "UQE_api_key_account_id_name" CASCADE	t		2024-05-15 15:34:45
123	Rename table api_key to api_key_v1 - v1	ALTER TABLE "api_key" RENAME TO "api_key_v1"	t		2024-05-15 15:34:45
164	Update preferences table charset	ALTER TABLE "preferences" ALTER "timezone" TYPE VARCHAR(50), ALTER "theme" TYPE VARCHAR(20);	t		2024-05-15 15:34:46
124	create api_key table v2	CREATE TABLE IF NOT EXISTS "api_key" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "name" VARCHAR(190) NOT NULL\n, "key" VARCHAR(190) NOT NULL\n, "role" VARCHAR(255) NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:45
125	create index IDX_api_key_org_id - v2	CREATE INDEX "IDX_api_key_org_id" ON "api_key" ("org_id");	t		2024-05-15 15:34:45
126	create index UQE_api_key_key - v2	CREATE UNIQUE INDEX "UQE_api_key_key" ON "api_key" ("key");	t		2024-05-15 15:34:45
127	create index UQE_api_key_org_id_name - v2	CREATE UNIQUE INDEX "UQE_api_key_org_id_name" ON "api_key" ("org_id","name");	t		2024-05-15 15:34:45
128	copy api_key v1 to v2	INSERT INTO "api_key" ("name"\n, "key"\n, "role"\n, "created"\n, "updated"\n, "id"\n, "org_id") SELECT "name"\n, "key"\n, "role"\n, "created"\n, "updated"\n, "id"\n, "account_id" FROM "api_key_v1"	t		2024-05-15 15:34:45
129	Drop old table api_key_v1	DROP TABLE IF EXISTS "api_key_v1"	t		2024-05-15 15:34:45
130	Update api_key table charset	ALTER TABLE "api_key" ALTER "name" TYPE VARCHAR(190), ALTER "key" TYPE VARCHAR(190), ALTER "role" TYPE VARCHAR(255);	t		2024-05-15 15:34:45
131	Add expires to api_key table	alter table "api_key" ADD COLUMN "expires" BIGINT NULL 	t		2024-05-15 15:34:45
132	Add service account foreign key	alter table "api_key" ADD COLUMN "service_account_id" BIGINT NULL 	t		2024-05-15 15:34:45
133	set service account foreign key to nil if 0	UPDATE api_key SET service_account_id = NULL WHERE service_account_id = 0;	t		2024-05-15 15:34:45
134	Add last_used_at to api_key table	alter table "api_key" ADD COLUMN "last_used_at" TIMESTAMP NULL 	t		2024-05-15 15:34:45
135	Add is_revoked column to api_key table	alter table "api_key" ADD COLUMN "is_revoked" BOOL NULL DEFAULT false 	t		2024-05-15 15:34:45
136	create dashboard_snapshot table v4	CREATE TABLE IF NOT EXISTS "dashboard_snapshot" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "name" VARCHAR(255) NOT NULL\n, "key" VARCHAR(190) NOT NULL\n, "dashboard" TEXT NOT NULL\n, "expires" TIMESTAMP NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:45
137	drop table dashboard_snapshot_v4 #1	DROP TABLE IF EXISTS "dashboard_snapshot"	t		2024-05-15 15:34:45
138	create dashboard_snapshot table v5 #2	CREATE TABLE IF NOT EXISTS "dashboard_snapshot" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "name" VARCHAR(255) NOT NULL\n, "key" VARCHAR(190) NOT NULL\n, "delete_key" VARCHAR(190) NOT NULL\n, "org_id" BIGINT NOT NULL\n, "user_id" BIGINT NOT NULL\n, "external" BOOL NOT NULL\n, "external_url" VARCHAR(255) NOT NULL\n, "dashboard" TEXT NOT NULL\n, "expires" TIMESTAMP NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:45
139	create index UQE_dashboard_snapshot_key - v5	CREATE UNIQUE INDEX "UQE_dashboard_snapshot_key" ON "dashboard_snapshot" ("key");	t		2024-05-15 15:34:45
140	create index UQE_dashboard_snapshot_delete_key - v5	CREATE UNIQUE INDEX "UQE_dashboard_snapshot_delete_key" ON "dashboard_snapshot" ("delete_key");	t		2024-05-15 15:34:45
141	create index IDX_dashboard_snapshot_user_id - v5	CREATE INDEX "IDX_dashboard_snapshot_user_id" ON "dashboard_snapshot" ("user_id");	t		2024-05-15 15:34:45
142	alter dashboard_snapshot to mediumtext v2	SELECT 0;	t		2024-05-15 15:34:45
143	Update dashboard_snapshot table charset	ALTER TABLE "dashboard_snapshot" ALTER "name" TYPE VARCHAR(255), ALTER "key" TYPE VARCHAR(190), ALTER "delete_key" TYPE VARCHAR(190), ALTER "external_url" TYPE VARCHAR(255), ALTER "dashboard" TYPE TEXT;	t		2024-05-15 15:34:45
144	Add column external_delete_url to dashboard_snapshots table	alter table "dashboard_snapshot" ADD COLUMN "external_delete_url" VARCHAR(255) NULL 	t		2024-05-15 15:34:45
145	Add encrypted dashboard json column	alter table "dashboard_snapshot" ADD COLUMN "dashboard_encrypted" BYTEA NULL 	t		2024-05-15 15:34:45
146	Change dashboard_encrypted column to MEDIUMBLOB	SELECT 0;	t		2024-05-15 15:34:46
147	create quota table v1	CREATE TABLE IF NOT EXISTS "quota" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NULL\n, "user_id" BIGINT NULL\n, "target" VARCHAR(190) NOT NULL\n, "limit" BIGINT NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:46
148	create index UQE_quota_org_id_user_id_target - v1	CREATE UNIQUE INDEX "UQE_quota_org_id_user_id_target" ON "quota" ("org_id","user_id","target");	t		2024-05-15 15:34:46
149	Update quota table charset	ALTER TABLE "quota" ALTER "target" TYPE VARCHAR(190);	t		2024-05-15 15:34:46
150	create plugin_setting table	CREATE TABLE IF NOT EXISTS "plugin_setting" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NULL\n, "plugin_id" VARCHAR(190) NOT NULL\n, "enabled" BOOL NOT NULL\n, "pinned" BOOL NOT NULL\n, "json_data" TEXT NULL\n, "secure_json_data" TEXT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:46
151	create index UQE_plugin_setting_org_id_plugin_id - v1	CREATE UNIQUE INDEX "UQE_plugin_setting_org_id_plugin_id" ON "plugin_setting" ("org_id","plugin_id");	t		2024-05-15 15:34:46
152	Add column plugin_version to plugin_settings	alter table "plugin_setting" ADD COLUMN "plugin_version" VARCHAR(50) NULL 	t		2024-05-15 15:34:46
153	Update plugin_setting table charset	ALTER TABLE "plugin_setting" ALTER "plugin_id" TYPE VARCHAR(190), ALTER "json_data" TYPE TEXT, ALTER "secure_json_data" TYPE TEXT, ALTER "plugin_version" TYPE VARCHAR(50);	t		2024-05-15 15:34:46
154	create session table	CREATE TABLE IF NOT EXISTS "session" (\n"key" CHAR(16) PRIMARY KEY NOT NULL\n, "data" BYTEA NOT NULL\n, "expiry" INTEGER NOT NULL\n);	t		2024-05-15 15:34:46
155	Drop old table playlist table	DROP TABLE IF EXISTS "playlist"	t		2024-05-15 15:34:46
156	Drop old table playlist_item table	DROP TABLE IF EXISTS "playlist_item"	t		2024-05-15 15:34:46
157	create playlist table v2	CREATE TABLE IF NOT EXISTS "playlist" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "name" VARCHAR(255) NOT NULL\n, "interval" VARCHAR(255) NOT NULL\n, "org_id" BIGINT NOT NULL\n);	t		2024-05-15 15:34:46
158	create playlist item table v2	CREATE TABLE IF NOT EXISTS "playlist_item" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "playlist_id" BIGINT NOT NULL\n, "type" VARCHAR(255) NOT NULL\n, "value" TEXT NOT NULL\n, "title" TEXT NOT NULL\n, "order" INTEGER NOT NULL\n);	t		2024-05-15 15:34:46
159	Update playlist table charset	ALTER TABLE "playlist" ALTER "name" TYPE VARCHAR(255), ALTER "interval" TYPE VARCHAR(255);	t		2024-05-15 15:34:46
160	Update playlist_item table charset	ALTER TABLE "playlist_item" ALTER "type" TYPE VARCHAR(255), ALTER "value" TYPE TEXT, ALTER "title" TYPE TEXT;	t		2024-05-15 15:34:46
161	drop preferences table v2	DROP TABLE IF EXISTS "preferences"	t		2024-05-15 15:34:46
162	drop preferences table v3	DROP TABLE IF EXISTS "preferences"	t		2024-05-15 15:34:46
163	create preferences table v3	CREATE TABLE IF NOT EXISTS "preferences" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "user_id" BIGINT NOT NULL\n, "version" INTEGER NOT NULL\n, "home_dashboard_id" BIGINT NOT NULL\n, "timezone" VARCHAR(50) NOT NULL\n, "theme" VARCHAR(20) NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:46
248	alter dashboard_version.data to mediumtext v1	SELECT 0;	t		2024-05-15 15:34:54
166	Update team_id column values in preferences	UPDATE preferences SET team_id=0 WHERE team_id IS NULL;	t		2024-05-15 15:34:46
167	Add column week_start in preferences	alter table "preferences" ADD COLUMN "week_start" VARCHAR(10) NULL 	t		2024-05-15 15:34:46
168	Add column preferences.json_data	alter table "preferences" ADD COLUMN "json_data" TEXT NULL 	t		2024-05-15 15:34:46
169	alter preferences.json_data to mediumtext v1	SELECT 0;	t		2024-05-15 15:34:46
170	Add preferences index org_id	CREATE INDEX "IDX_preferences_org_id" ON "preferences" ("org_id");	t		2024-05-15 15:34:46
171	Add preferences index user_id	CREATE INDEX "IDX_preferences_user_id" ON "preferences" ("user_id");	t		2024-05-15 15:34:46
172	create alert table v1	CREATE TABLE IF NOT EXISTS "alert" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "version" BIGINT NOT NULL\n, "dashboard_id" BIGINT NOT NULL\n, "panel_id" BIGINT NOT NULL\n, "org_id" BIGINT NOT NULL\n, "name" VARCHAR(255) NOT NULL\n, "message" TEXT NOT NULL\n, "state" VARCHAR(190) NOT NULL\n, "settings" TEXT NOT NULL\n, "frequency" BIGINT NOT NULL\n, "handler" BIGINT NOT NULL\n, "severity" TEXT NOT NULL\n, "silenced" BOOL NOT NULL\n, "execution_error" TEXT NOT NULL\n, "eval_data" TEXT NULL\n, "eval_date" TIMESTAMP NULL\n, "new_state_date" TIMESTAMP NOT NULL\n, "state_changes" INTEGER NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:46
173	add index alert org_id & id 	CREATE INDEX "IDX_alert_org_id_id" ON "alert" ("org_id","id");	t		2024-05-15 15:34:46
174	add index alert state	CREATE INDEX "IDX_alert_state" ON "alert" ("state");	t		2024-05-15 15:34:46
175	add index alert dashboard_id	CREATE INDEX "IDX_alert_dashboard_id" ON "alert" ("dashboard_id");	t		2024-05-15 15:34:46
176	Create alert_rule_tag table v1	CREATE TABLE IF NOT EXISTS "alert_rule_tag" (\n"alert_id" BIGINT NOT NULL\n, "tag_id" BIGINT NOT NULL\n);	t		2024-05-15 15:34:46
177	Add unique index alert_rule_tag.alert_id_tag_id	CREATE UNIQUE INDEX "UQE_alert_rule_tag_alert_id_tag_id" ON "alert_rule_tag" ("alert_id","tag_id");	t		2024-05-15 15:34:46
178	drop index UQE_alert_rule_tag_alert_id_tag_id - v1	DROP INDEX "UQE_alert_rule_tag_alert_id_tag_id" CASCADE	t		2024-05-15 15:34:46
179	Rename table alert_rule_tag to alert_rule_tag_v1 - v1	ALTER TABLE "alert_rule_tag" RENAME TO "alert_rule_tag_v1"	t		2024-05-15 15:34:46
180	Create alert_rule_tag table v2	CREATE TABLE IF NOT EXISTS "alert_rule_tag" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "alert_id" BIGINT NOT NULL\n, "tag_id" BIGINT NOT NULL\n);	t		2024-05-15 15:34:46
181	create index UQE_alert_rule_tag_alert_id_tag_id - Add unique index alert_rule_tag.alert_id_tag_id V2	CREATE UNIQUE INDEX "UQE_alert_rule_tag_alert_id_tag_id" ON "alert_rule_tag" ("alert_id","tag_id");	t		2024-05-15 15:34:47
182	copy alert_rule_tag v1 to v2	INSERT INTO "alert_rule_tag" ("alert_id"\n, "tag_id") SELECT "alert_id"\n, "tag_id" FROM "alert_rule_tag_v1"	t		2024-05-15 15:34:47
183	drop table alert_rule_tag_v1	DROP TABLE IF EXISTS "alert_rule_tag_v1"	t		2024-05-15 15:34:47
184	create alert_notification table v1	CREATE TABLE IF NOT EXISTS "alert_notification" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "name" VARCHAR(190) NOT NULL\n, "type" VARCHAR(255) NOT NULL\n, "settings" TEXT NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:47
185	Add column is_default	alter table "alert_notification" ADD COLUMN "is_default" BOOL NOT NULL DEFAULT false 	t		2024-05-15 15:34:47
186	Add column frequency	alter table "alert_notification" ADD COLUMN "frequency" BIGINT NULL 	t		2024-05-15 15:34:47
187	Add column send_reminder	alter table "alert_notification" ADD COLUMN "send_reminder" BOOL NULL DEFAULT false 	t		2024-05-15 15:34:47
188	Add column disable_resolve_message	alter table "alert_notification" ADD COLUMN "disable_resolve_message" BOOL NOT NULL DEFAULT false 	t		2024-05-15 15:34:47
189	add index alert_notification org_id & name	CREATE UNIQUE INDEX "UQE_alert_notification_org_id_name" ON "alert_notification" ("org_id","name");	t		2024-05-15 15:34:47
190	Update alert table charset	ALTER TABLE "alert" ALTER "name" TYPE VARCHAR(255), ALTER "message" TYPE TEXT, ALTER "state" TYPE VARCHAR(190), ALTER "settings" TYPE TEXT, ALTER "severity" TYPE TEXT, ALTER "execution_error" TYPE TEXT, ALTER "eval_data" TYPE TEXT;	t		2024-05-15 15:34:47
191	Update alert_notification table charset	ALTER TABLE "alert_notification" ALTER "name" TYPE VARCHAR(190), ALTER "type" TYPE VARCHAR(255), ALTER "settings" TYPE TEXT;	t		2024-05-15 15:34:47
192	create notification_journal table v1	CREATE TABLE IF NOT EXISTS "alert_notification_journal" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "alert_id" BIGINT NOT NULL\n, "notifier_id" BIGINT NOT NULL\n, "sent_at" BIGINT NOT NULL\n, "success" BOOL NOT NULL\n);	t		2024-05-15 15:34:48
193	add index notification_journal org_id & alert_id & notifier_id	CREATE INDEX "IDX_alert_notification_journal_org_id_alert_id_notifier_id" ON "alert_notification_journal" ("org_id","alert_id","notifier_id");	t		2024-05-15 15:34:48
194	drop alert_notification_journal	DROP TABLE IF EXISTS "alert_notification_journal"	t		2024-05-15 15:34:48
195	create alert_notification_state table v1	CREATE TABLE IF NOT EXISTS "alert_notification_state" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "alert_id" BIGINT NOT NULL\n, "notifier_id" BIGINT NOT NULL\n, "state" VARCHAR(50) NOT NULL\n, "version" BIGINT NOT NULL\n, "updated_at" BIGINT NOT NULL\n, "alert_rule_state_updated_version" BIGINT NOT NULL\n);	t		2024-05-15 15:34:51
196	add index alert_notification_state org_id & alert_id & notifier_id	CREATE UNIQUE INDEX "UQE_alert_notification_state_org_id_alert_id_notifier_id" ON "alert_notification_state" ("org_id","alert_id","notifier_id");	t		2024-05-15 15:34:52
197	Add for to alert table	alter table "alert" ADD COLUMN "for" BIGINT NULL 	t		2024-05-15 15:34:52
198	Add column uid in alert_notification	alter table "alert_notification" ADD COLUMN "uid" VARCHAR(40) NULL 	t		2024-05-15 15:34:52
199	Update uid column values in alert_notification	UPDATE alert_notification SET uid=lpad('' || id::text,9,'0') WHERE uid IS NULL;	t		2024-05-15 15:34:52
200	Add unique index alert_notification_org_id_uid	CREATE UNIQUE INDEX "UQE_alert_notification_org_id_uid" ON "alert_notification" ("org_id","uid");	t		2024-05-15 15:34:52
201	Remove unique index org_id_name	DROP INDEX "UQE_alert_notification_org_id_name" CASCADE	t		2024-05-15 15:34:52
202	Add column secure_settings in alert_notification	alter table "alert_notification" ADD COLUMN "secure_settings" TEXT NULL 	t		2024-05-15 15:34:52
203	alter alert.settings to mediumtext	SELECT 0;	t		2024-05-15 15:34:52
204	Add non-unique index alert_notification_state_alert_id	CREATE INDEX "IDX_alert_notification_state_alert_id" ON "alert_notification_state" ("alert_id");	t		2024-05-15 15:34:52
205	Add non-unique index alert_rule_tag_alert_id	CREATE INDEX "IDX_alert_rule_tag_alert_id" ON "alert_rule_tag" ("alert_id");	t		2024-05-15 15:34:52
206	Drop old annotation table v4	DROP TABLE IF EXISTS "annotation"	t		2024-05-15 15:34:53
291	add unique index user_auth_token.auth_token	CREATE UNIQUE INDEX "UQE_user_auth_token_auth_token" ON "user_auth_token" ("auth_token");	t		2024-05-15 15:34:56
207	create annotation table v5	CREATE TABLE IF NOT EXISTS "annotation" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "alert_id" BIGINT NULL\n, "user_id" BIGINT NULL\n, "dashboard_id" BIGINT NULL\n, "panel_id" BIGINT NULL\n, "category_id" BIGINT NULL\n, "type" VARCHAR(25) NOT NULL\n, "title" TEXT NOT NULL\n, "text" TEXT NOT NULL\n, "metric" VARCHAR(255) NULL\n, "prev_state" VARCHAR(25) NOT NULL\n, "new_state" VARCHAR(25) NOT NULL\n, "data" TEXT NOT NULL\n, "epoch" BIGINT NOT NULL\n);	t		2024-05-15 15:34:53
208	add index annotation 0 v3	CREATE INDEX "IDX_annotation_org_id_alert_id" ON "annotation" ("org_id","alert_id");	t		2024-05-15 15:34:53
209	add index annotation 1 v3	CREATE INDEX "IDX_annotation_org_id_type" ON "annotation" ("org_id","type");	t		2024-05-15 15:34:53
210	add index annotation 2 v3	CREATE INDEX "IDX_annotation_org_id_category_id" ON "annotation" ("org_id","category_id");	t		2024-05-15 15:34:53
211	add index annotation 3 v3	CREATE INDEX "IDX_annotation_org_id_dashboard_id_panel_id_epoch" ON "annotation" ("org_id","dashboard_id","panel_id","epoch");	t		2024-05-15 15:34:53
212	add index annotation 4 v3	CREATE INDEX "IDX_annotation_org_id_epoch" ON "annotation" ("org_id","epoch");	t		2024-05-15 15:34:53
213	Update annotation table charset	ALTER TABLE "annotation" ALTER "type" TYPE VARCHAR(25), ALTER "title" TYPE TEXT, ALTER "text" TYPE TEXT, ALTER "metric" TYPE VARCHAR(255), ALTER "prev_state" TYPE VARCHAR(25), ALTER "new_state" TYPE VARCHAR(25), ALTER "data" TYPE TEXT;	t		2024-05-15 15:34:53
214	Add column region_id to annotation table	alter table "annotation" ADD COLUMN "region_id" BIGINT NULL DEFAULT 0 	t		2024-05-15 15:34:53
215	Drop category_id index	DROP INDEX "IDX_annotation_org_id_category_id" CASCADE	t		2024-05-15 15:34:53
216	Add column tags to annotation table	alter table "annotation" ADD COLUMN "tags" VARCHAR(500) NULL 	t		2024-05-15 15:34:53
217	Create annotation_tag table v2	CREATE TABLE IF NOT EXISTS "annotation_tag" (\n"annotation_id" BIGINT NOT NULL\n, "tag_id" BIGINT NOT NULL\n);	t		2024-05-15 15:34:53
218	Add unique index annotation_tag.annotation_id_tag_id	CREATE UNIQUE INDEX "UQE_annotation_tag_annotation_id_tag_id" ON "annotation_tag" ("annotation_id","tag_id");	t		2024-05-15 15:34:53
219	drop index UQE_annotation_tag_annotation_id_tag_id - v2	DROP INDEX "UQE_annotation_tag_annotation_id_tag_id" CASCADE	t		2024-05-15 15:34:53
220	Rename table annotation_tag to annotation_tag_v2 - v2	ALTER TABLE "annotation_tag" RENAME TO "annotation_tag_v2"	t		2024-05-15 15:34:53
221	Create annotation_tag table v3	CREATE TABLE IF NOT EXISTS "annotation_tag" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "annotation_id" BIGINT NOT NULL\n, "tag_id" BIGINT NOT NULL\n);	t		2024-05-15 15:34:53
222	create index UQE_annotation_tag_annotation_id_tag_id - Add unique index annotation_tag.annotation_id_tag_id V3	CREATE UNIQUE INDEX "UQE_annotation_tag_annotation_id_tag_id" ON "annotation_tag" ("annotation_id","tag_id");	t		2024-05-15 15:34:53
223	copy annotation_tag v2 to v3	INSERT INTO "annotation_tag" ("annotation_id"\n, "tag_id") SELECT "annotation_id"\n, "tag_id" FROM "annotation_tag_v2"	t		2024-05-15 15:34:53
224	drop table annotation_tag_v2	DROP TABLE IF EXISTS "annotation_tag_v2"	t		2024-05-15 15:34:53
225	Update alert annotations and set TEXT to empty	UPDATE annotation SET TEXT = '' WHERE alert_id > 0	t		2024-05-15 15:34:53
226	Add created time to annotation table	alter table "annotation" ADD COLUMN "created" BIGINT NULL DEFAULT 0 	t		2024-05-15 15:34:53
227	Add updated time to annotation table	alter table "annotation" ADD COLUMN "updated" BIGINT NULL DEFAULT 0 	t		2024-05-15 15:34:53
228	Add index for created in annotation table	CREATE INDEX "IDX_annotation_org_id_created" ON "annotation" ("org_id","created");	t		2024-05-15 15:34:53
229	Add index for updated in annotation table	CREATE INDEX "IDX_annotation_org_id_updated" ON "annotation" ("org_id","updated");	t		2024-05-15 15:34:54
230	Convert existing annotations from seconds to milliseconds	UPDATE annotation SET epoch = (epoch*1000) where epoch < 9999999999	t		2024-05-15 15:34:54
231	Add epoch_end column	alter table "annotation" ADD COLUMN "epoch_end" BIGINT NOT NULL DEFAULT 0 	t		2024-05-15 15:34:54
232	Add index for epoch_end	CREATE INDEX "IDX_annotation_org_id_epoch_epoch_end" ON "annotation" ("org_id","epoch","epoch_end");	t		2024-05-15 15:34:54
233	Make epoch_end the same as epoch	UPDATE annotation SET epoch_end = epoch	t		2024-05-15 15:34:54
234	Move region to single row	code migration	t		2024-05-15 15:34:54
235	Remove index org_id_epoch from annotation table	DROP INDEX "IDX_annotation_org_id_epoch" CASCADE	t		2024-05-15 15:34:54
236	Remove index org_id_dashboard_id_panel_id_epoch from annotation table	DROP INDEX "IDX_annotation_org_id_dashboard_id_panel_id_epoch" CASCADE	t		2024-05-15 15:34:54
237	Add index for org_id_dashboard_id_epoch_end_epoch on annotation table	CREATE INDEX "IDX_annotation_org_id_dashboard_id_epoch_end_epoch" ON "annotation" ("org_id","dashboard_id","epoch_end","epoch");	t		2024-05-15 15:34:54
238	Add index for org_id_epoch_end_epoch on annotation table	CREATE INDEX "IDX_annotation_org_id_epoch_end_epoch" ON "annotation" ("org_id","epoch_end","epoch");	t		2024-05-15 15:34:54
239	Remove index org_id_epoch_epoch_end from annotation table	DROP INDEX "IDX_annotation_org_id_epoch_epoch_end" CASCADE	t		2024-05-15 15:34:54
240	Add index for alert_id on annotation table	CREATE INDEX "IDX_annotation_alert_id" ON "annotation" ("alert_id");	t		2024-05-15 15:34:54
241	Increase tags column to length 4096	ALTER TABLE annotation ALTER COLUMN tags TYPE VARCHAR(4096);	t		2024-05-15 15:34:54
242	create test_data table	CREATE TABLE IF NOT EXISTS "test_data" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "metric1" VARCHAR(20) NULL\n, "metric2" VARCHAR(150) NULL\n, "value_big_int" BIGINT NULL\n, "value_double" DOUBLE PRECISION NULL\n, "value_float" REAL NULL\n, "value_int" INTEGER NULL\n, "time_epoch" BIGINT NOT NULL\n, "time_date_time" TIMESTAMP NOT NULL\n, "time_time_stamp" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:54
243	create dashboard_version table v1	CREATE TABLE IF NOT EXISTS "dashboard_version" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "dashboard_id" BIGINT NOT NULL\n, "parent_version" INTEGER NOT NULL\n, "restored_from" INTEGER NOT NULL\n, "version" INTEGER NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "created_by" BIGINT NOT NULL\n, "message" TEXT NOT NULL\n, "data" TEXT NOT NULL\n);	t		2024-05-15 15:34:54
244	add index dashboard_version.dashboard_id	CREATE INDEX "IDX_dashboard_version_dashboard_id" ON "dashboard_version" ("dashboard_id");	t		2024-05-15 15:34:54
245	add unique index dashboard_version.dashboard_id and dashboard_version.version	CREATE UNIQUE INDEX "UQE_dashboard_version_dashboard_id_version" ON "dashboard_version" ("dashboard_id","version");	t		2024-05-15 15:34:54
246	Set dashboard version to 1 where 0	UPDATE dashboard SET version = 1 WHERE version = 0	t		2024-05-15 15:34:54
247	save existing dashboard data in dashboard_version table v1	INSERT INTO dashboard_version\n(\n\tdashboard_id,\n\tversion,\n\tparent_version,\n\trestored_from,\n\tcreated,\n\tcreated_by,\n\tmessage,\n\tdata\n)\nSELECT\n\tdashboard.id,\n\tdashboard.version,\n\tdashboard.version,\n\tdashboard.version,\n\tdashboard.updated,\n\tCOALESCE(dashboard.updated_by, -1),\n\t'',\n\tdashboard.data\nFROM dashboard;	t		2024-05-15 15:34:54
249	create team table	CREATE TABLE IF NOT EXISTS "team" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "name" VARCHAR(190) NOT NULL\n, "org_id" BIGINT NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:54
250	add index team.org_id	CREATE INDEX "IDX_team_org_id" ON "team" ("org_id");	t		2024-05-15 15:34:54
251	add unique index team_org_id_name	CREATE UNIQUE INDEX "UQE_team_org_id_name" ON "team" ("org_id","name");	t		2024-05-15 15:34:54
252	create team member table	CREATE TABLE IF NOT EXISTS "team_member" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "team_id" BIGINT NOT NULL\n, "user_id" BIGINT NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:54
253	add index team_member.org_id	CREATE INDEX "IDX_team_member_org_id" ON "team_member" ("org_id");	t		2024-05-15 15:34:54
254	add unique index team_member_org_id_team_id_user_id	CREATE UNIQUE INDEX "UQE_team_member_org_id_team_id_user_id" ON "team_member" ("org_id","team_id","user_id");	t		2024-05-15 15:34:54
255	add index team_member.team_id	CREATE INDEX "IDX_team_member_team_id" ON "team_member" ("team_id");	t		2024-05-15 15:34:55
256	Add column email to team table	alter table "team" ADD COLUMN "email" VARCHAR(190) NULL 	t		2024-05-15 15:34:55
257	Add column external to team_member table	alter table "team_member" ADD COLUMN "external" BOOL NULL 	t		2024-05-15 15:34:55
258	Add column permission to team_member table	alter table "team_member" ADD COLUMN "permission" SMALLINT NULL 	t		2024-05-15 15:34:55
259	create dashboard acl table	CREATE TABLE IF NOT EXISTS "dashboard_acl" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "dashboard_id" BIGINT NOT NULL\n, "user_id" BIGINT NULL\n, "team_id" BIGINT NULL\n, "permission" SMALLINT NOT NULL DEFAULT 4\n, "role" VARCHAR(20) NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:55
260	add index dashboard_acl_dashboard_id	CREATE INDEX "IDX_dashboard_acl_dashboard_id" ON "dashboard_acl" ("dashboard_id");	t		2024-05-15 15:34:55
261	add unique index dashboard_acl_dashboard_id_user_id	CREATE UNIQUE INDEX "UQE_dashboard_acl_dashboard_id_user_id" ON "dashboard_acl" ("dashboard_id","user_id");	t		2024-05-15 15:34:55
262	add unique index dashboard_acl_dashboard_id_team_id	CREATE UNIQUE INDEX "UQE_dashboard_acl_dashboard_id_team_id" ON "dashboard_acl" ("dashboard_id","team_id");	t		2024-05-15 15:34:55
263	add index dashboard_acl_user_id	CREATE INDEX "IDX_dashboard_acl_user_id" ON "dashboard_acl" ("user_id");	t		2024-05-15 15:34:55
264	add index dashboard_acl_team_id	CREATE INDEX "IDX_dashboard_acl_team_id" ON "dashboard_acl" ("team_id");	t		2024-05-15 15:34:55
265	add index dashboard_acl_org_id_role	CREATE INDEX "IDX_dashboard_acl_org_id_role" ON "dashboard_acl" ("org_id","role");	t		2024-05-15 15:34:55
266	add index dashboard_permission	CREATE INDEX "IDX_dashboard_acl_permission" ON "dashboard_acl" ("permission");	t		2024-05-15 15:34:55
267	save default acl rules in dashboard_acl table	\nINSERT INTO dashboard_acl\n\t(\n\t\torg_id,\n\t\tdashboard_id,\n\t\tpermission,\n\t\trole,\n\t\tcreated,\n\t\tupdated\n\t)\n\tVALUES\n\t\t(-1,-1, 1,'Viewer','2017-06-20','2017-06-20'),\n\t\t(-1,-1, 2,'Editor','2017-06-20','2017-06-20')\n\t	t		2024-05-15 15:34:55
268	delete acl rules for deleted dashboards and folders	DELETE FROM dashboard_acl WHERE dashboard_id NOT IN (SELECT id FROM dashboard) AND dashboard_id != -1	t		2024-05-15 15:34:55
269	create tag table	CREATE TABLE IF NOT EXISTS "tag" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "key" VARCHAR(100) NOT NULL\n, "value" VARCHAR(100) NOT NULL\n);	t		2024-05-15 15:34:55
270	add index tag.key_value	CREATE UNIQUE INDEX "UQE_tag_key_value" ON "tag" ("key","value");	t		2024-05-15 15:34:56
271	create login attempt table	CREATE TABLE IF NOT EXISTS "login_attempt" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "username" VARCHAR(190) NOT NULL\n, "ip_address" VARCHAR(30) NOT NULL\n, "created" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:56
272	add index login_attempt.username	CREATE INDEX "IDX_login_attempt_username" ON "login_attempt" ("username");	t		2024-05-15 15:34:56
273	drop index IDX_login_attempt_username - v1	DROP INDEX "IDX_login_attempt_username" CASCADE	t		2024-05-15 15:34:56
274	Rename table login_attempt to login_attempt_tmp_qwerty - v1	ALTER TABLE "login_attempt" RENAME TO "login_attempt_tmp_qwerty"	t		2024-05-15 15:34:56
275	create login_attempt v2	CREATE TABLE IF NOT EXISTS "login_attempt" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "username" VARCHAR(190) NOT NULL\n, "ip_address" VARCHAR(30) NOT NULL\n, "created" INTEGER NOT NULL DEFAULT 0\n);	t		2024-05-15 15:34:56
276	create index IDX_login_attempt_username - v2	CREATE INDEX "IDX_login_attempt_username" ON "login_attempt" ("username");	t		2024-05-15 15:34:56
277	copy login_attempt v1 to v2	INSERT INTO "login_attempt" ("username"\n, "ip_address"\n, "id") SELECT "username"\n, "ip_address"\n, "id" FROM "login_attempt_tmp_qwerty"	t		2024-05-15 15:34:56
278	drop login_attempt_tmp_qwerty	DROP TABLE IF EXISTS "login_attempt_tmp_qwerty"	t		2024-05-15 15:34:56
279	create user auth table	CREATE TABLE IF NOT EXISTS "user_auth" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "user_id" BIGINT NOT NULL\n, "auth_module" VARCHAR(190) NOT NULL\n, "auth_id" VARCHAR(100) NOT NULL\n, "created" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:34:56
280	create index IDX_user_auth_auth_module_auth_id - v1	CREATE INDEX "IDX_user_auth_auth_module_auth_id" ON "user_auth" ("auth_module","auth_id");	t		2024-05-15 15:34:56
281	alter user_auth.auth_id to length 190	ALTER TABLE user_auth ALTER COLUMN auth_id TYPE VARCHAR(190);	t		2024-05-15 15:34:56
282	Add OAuth access token to user_auth	alter table "user_auth" ADD COLUMN "o_auth_access_token" TEXT NULL 	t		2024-05-15 15:34:56
283	Add OAuth refresh token to user_auth	alter table "user_auth" ADD COLUMN "o_auth_refresh_token" TEXT NULL 	t		2024-05-15 15:34:56
284	Add OAuth token type to user_auth	alter table "user_auth" ADD COLUMN "o_auth_token_type" TEXT NULL 	t		2024-05-15 15:34:56
285	Add OAuth expiry to user_auth	alter table "user_auth" ADD COLUMN "o_auth_expiry" TIMESTAMP NULL 	t		2024-05-15 15:34:56
286	Add index to user_id column in user_auth	CREATE INDEX "IDX_user_auth_user_id" ON "user_auth" ("user_id");	t		2024-05-15 15:34:56
287	Add OAuth ID token to user_auth	alter table "user_auth" ADD COLUMN "o_auth_id_token" TEXT NULL 	t		2024-05-15 15:34:56
288	create server_lock table	CREATE TABLE IF NOT EXISTS "server_lock" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "operation_uid" VARCHAR(100) NOT NULL\n, "version" BIGINT NOT NULL\n, "last_execution" BIGINT NOT NULL\n);	t		2024-05-15 15:34:56
289	add index server_lock.operation_uid	CREATE UNIQUE INDEX "UQE_server_lock_operation_uid" ON "server_lock" ("operation_uid");	t		2024-05-15 15:34:56
290	create user auth token table	CREATE TABLE IF NOT EXISTS "user_auth_token" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "user_id" BIGINT NOT NULL\n, "auth_token" VARCHAR(100) NOT NULL\n, "prev_auth_token" VARCHAR(100) NOT NULL\n, "user_agent" VARCHAR(255) NOT NULL\n, "client_ip" VARCHAR(255) NOT NULL\n, "auth_token_seen" BOOL NOT NULL\n, "seen_at" INTEGER NULL\n, "rotated_at" INTEGER NOT NULL\n, "created_at" INTEGER NOT NULL\n, "updated_at" INTEGER NOT NULL\n);	t		2024-05-15 15:34:56
292	add unique index user_auth_token.prev_auth_token	CREATE UNIQUE INDEX "UQE_user_auth_token_prev_auth_token" ON "user_auth_token" ("prev_auth_token");	t		2024-05-15 15:34:56
293	add index user_auth_token.user_id	CREATE INDEX "IDX_user_auth_token_user_id" ON "user_auth_token" ("user_id");	t		2024-05-15 15:34:57
294	Add revoked_at to the user auth token	alter table "user_auth_token" ADD COLUMN "revoked_at" INTEGER NULL 	t		2024-05-15 15:34:58
295	create cache_data table	CREATE TABLE IF NOT EXISTS "cache_data" (\n"cache_key" VARCHAR(168) PRIMARY KEY NOT NULL\n, "data" BYTEA NOT NULL\n, "expires" INTEGER NOT NULL\n, "created_at" INTEGER NOT NULL\n);	t		2024-05-15 15:34:58
296	add unique index cache_data.cache_key	CREATE UNIQUE INDEX "UQE_cache_data_cache_key" ON "cache_data" ("cache_key");	t		2024-05-15 15:34:58
297	create short_url table v1	CREATE TABLE IF NOT EXISTS "short_url" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "uid" VARCHAR(40) NOT NULL\n, "path" TEXT NOT NULL\n, "created_by" INTEGER NOT NULL\n, "created_at" INTEGER NOT NULL\n, "last_seen_at" INTEGER NULL\n);	t		2024-05-15 15:34:58
298	add index short_url.org_id-uid	CREATE UNIQUE INDEX "UQE_short_url_org_id_uid" ON "short_url" ("org_id","uid");	t		2024-05-15 15:34:58
299	alter table short_url alter column created_by type to bigint	ALTER TABLE short_url ALTER COLUMN created_by TYPE BIGINT;	t		2024-05-15 15:34:58
300	delete alert_definition table	DROP TABLE IF EXISTS "alert_definition"	t		2024-05-15 15:34:59
301	recreate alert_definition table	CREATE TABLE IF NOT EXISTS "alert_definition" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "title" VARCHAR(190) NOT NULL\n, "condition" VARCHAR(190) NOT NULL\n, "data" TEXT NOT NULL\n, "updated" TIMESTAMP NOT NULL\n, "interval_seconds" BIGINT NOT NULL DEFAULT 60\n, "version" INTEGER NOT NULL DEFAULT 0\n, "uid" VARCHAR(40) NOT NULL DEFAULT 0\n);	t		2024-05-15 15:34:59
302	add index in alert_definition on org_id and title columns	CREATE INDEX "IDX_alert_definition_org_id_title" ON "alert_definition" ("org_id","title");	t		2024-05-15 15:34:59
303	add index in alert_definition on org_id and uid columns	CREATE INDEX "IDX_alert_definition_org_id_uid" ON "alert_definition" ("org_id","uid");	t		2024-05-15 15:35:01
304	alter alert_definition table data column to mediumtext in mysql	SELECT 0;	t		2024-05-15 15:35:01
305	drop index in alert_definition on org_id and title columns	DROP INDEX "IDX_alert_definition_org_id_title" CASCADE	t		2024-05-15 15:35:01
306	drop index in alert_definition on org_id and uid columns	DROP INDEX "IDX_alert_definition_org_id_uid" CASCADE	t		2024-05-15 15:35:01
307	add unique index in alert_definition on org_id and title columns	CREATE UNIQUE INDEX "UQE_alert_definition_org_id_title" ON "alert_definition" ("org_id","title");	t		2024-05-15 15:35:01
308	add unique index in alert_definition on org_id and uid columns	CREATE UNIQUE INDEX "UQE_alert_definition_org_id_uid" ON "alert_definition" ("org_id","uid");	t		2024-05-15 15:35:01
309	Add column paused in alert_definition	alter table "alert_definition" ADD COLUMN "paused" BOOL NOT NULL DEFAULT false 	t		2024-05-15 15:35:02
310	drop alert_definition table	DROP TABLE IF EXISTS "alert_definition"	t		2024-05-15 15:35:02
311	delete alert_definition_version table	DROP TABLE IF EXISTS "alert_definition_version"	t		2024-05-15 15:35:02
312	recreate alert_definition_version table	CREATE TABLE IF NOT EXISTS "alert_definition_version" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "alert_definition_id" BIGINT NOT NULL\n, "alert_definition_uid" VARCHAR(40) NOT NULL DEFAULT 0\n, "parent_version" INTEGER NOT NULL\n, "restored_from" INTEGER NOT NULL\n, "version" INTEGER NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "title" VARCHAR(190) NOT NULL\n, "condition" VARCHAR(190) NOT NULL\n, "data" TEXT NOT NULL\n, "interval_seconds" BIGINT NOT NULL\n);	t		2024-05-15 15:35:02
313	add index in alert_definition_version table on alert_definition_id and version columns	CREATE UNIQUE INDEX "UQE_alert_definition_version_alert_definition_id_version" ON "alert_definition_version" ("alert_definition_id","version");	t		2024-05-15 15:35:03
314	add index in alert_definition_version table on alert_definition_uid and version columns	CREATE UNIQUE INDEX "UQE_alert_definition_version_alert_definition_uid_version" ON "alert_definition_version" ("alert_definition_uid","version");	t		2024-05-15 15:35:03
315	alter alert_definition_version table data column to mediumtext in mysql	SELECT 0;	t		2024-05-15 15:35:03
316	drop alert_definition_version table	DROP TABLE IF EXISTS "alert_definition_version"	t		2024-05-15 15:35:03
317	create alert_instance table	CREATE TABLE IF NOT EXISTS "alert_instance" (\n"def_org_id" BIGINT NOT NULL\n, "def_uid" VARCHAR(40) NOT NULL DEFAULT 0\n, "labels" TEXT NOT NULL\n, "labels_hash" VARCHAR(190) NOT NULL\n, "current_state" VARCHAR(190) NOT NULL\n, "current_state_since" BIGINT NOT NULL\n, "last_eval_time" BIGINT NOT NULL\n, PRIMARY KEY ( "def_org_id","def_uid","labels_hash" ));	t		2024-05-15 15:35:03
318	add index in alert_instance table on def_org_id, def_uid and current_state columns	CREATE INDEX "IDX_alert_instance_def_org_id_def_uid_current_state" ON "alert_instance" ("def_org_id","def_uid","current_state");	t		2024-05-15 15:35:03
319	add index in alert_instance table on def_org_id, current_state columns	CREATE INDEX "IDX_alert_instance_def_org_id_current_state" ON "alert_instance" ("def_org_id","current_state");	t		2024-05-15 15:35:04
320	add column current_state_end to alert_instance	alter table "alert_instance" ADD COLUMN "current_state_end" BIGINT NOT NULL DEFAULT 0 	t		2024-05-15 15:35:04
321	remove index def_org_id, def_uid, current_state on alert_instance	DROP INDEX "IDX_alert_instance_def_org_id_def_uid_current_state" CASCADE	t		2024-05-15 15:35:04
322	remove index def_org_id, current_state on alert_instance	DROP INDEX "IDX_alert_instance_def_org_id_current_state" CASCADE	t		2024-05-15 15:35:04
323	rename def_org_id to rule_org_id in alert_instance	ALTER TABLE alert_instance RENAME COLUMN def_org_id TO rule_org_id;	t		2024-05-15 15:35:04
324	rename def_uid to rule_uid in alert_instance	ALTER TABLE alert_instance RENAME COLUMN def_uid TO rule_uid;	t		2024-05-15 15:35:04
325	add index rule_org_id, rule_uid, current_state on alert_instance	CREATE INDEX "IDX_alert_instance_rule_org_id_rule_uid_current_state" ON "alert_instance" ("rule_org_id","rule_uid","current_state");	t		2024-05-15 15:35:04
326	add index rule_org_id, current_state on alert_instance	CREATE INDEX "IDX_alert_instance_rule_org_id_current_state" ON "alert_instance" ("rule_org_id","current_state");	t		2024-05-15 15:35:05
327	add current_reason column related to current_state	alter table "alert_instance" ADD COLUMN "current_reason" VARCHAR(190) NULL 	t		2024-05-15 15:35:05
328	create alert_rule table	CREATE TABLE IF NOT EXISTS "alert_rule" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "title" VARCHAR(190) NOT NULL\n, "condition" VARCHAR(190) NOT NULL\n, "data" TEXT NOT NULL\n, "updated" TIMESTAMP NOT NULL\n, "interval_seconds" BIGINT NOT NULL DEFAULT 60\n, "version" INTEGER NOT NULL DEFAULT 0\n, "uid" VARCHAR(40) NOT NULL DEFAULT 0\n, "namespace_uid" VARCHAR(40) NOT NULL\n, "rule_group" VARCHAR(190) NOT NULL\n, "no_data_state" VARCHAR(15) NOT NULL DEFAULT 'NoData'\n, "exec_err_state" VARCHAR(15) NOT NULL DEFAULT 'Alerting'\n);	t		2024-05-15 15:35:05
329	add index in alert_rule on org_id and title columns	CREATE UNIQUE INDEX "UQE_alert_rule_org_id_title" ON "alert_rule" ("org_id","title");	t		2024-05-15 15:35:06
330	add index in alert_rule on org_id and uid columns	CREATE UNIQUE INDEX "UQE_alert_rule_org_id_uid" ON "alert_rule" ("org_id","uid");	t		2024-05-15 15:35:06
331	add index in alert_rule on org_id, namespace_uid, group_uid columns	CREATE INDEX "IDX_alert_rule_org_id_namespace_uid_rule_group" ON "alert_rule" ("org_id","namespace_uid","rule_group");	t		2024-05-15 15:35:06
332	alter alert_rule table data column to mediumtext in mysql	SELECT 0;	t		2024-05-15 15:35:06
333	add column for to alert_rule	alter table "alert_rule" ADD COLUMN "for" BIGINT NOT NULL DEFAULT 0 	t		2024-05-15 15:35:06
334	add column annotations to alert_rule	alter table "alert_rule" ADD COLUMN "annotations" TEXT NULL 	t		2024-05-15 15:35:06
335	add column labels to alert_rule	alter table "alert_rule" ADD COLUMN "labels" TEXT NULL 	t		2024-05-15 15:35:06
336	remove unique index from alert_rule on org_id, title columns	DROP INDEX "UQE_alert_rule_org_id_title" CASCADE	t		2024-05-15 15:35:06
337	add index in alert_rule on org_id, namespase_uid and title columns	CREATE UNIQUE INDEX "UQE_alert_rule_org_id_namespace_uid_title" ON "alert_rule" ("org_id","namespace_uid","title");	t		2024-05-15 15:35:06
338	add dashboard_uid column to alert_rule	alter table "alert_rule" ADD COLUMN "dashboard_uid" VARCHAR(40) NULL 	t		2024-05-15 15:35:06
339	add panel_id column to alert_rule	alter table "alert_rule" ADD COLUMN "panel_id" BIGINT NULL 	t		2024-05-15 15:35:06
340	add index in alert_rule on org_id, dashboard_uid and panel_id columns	CREATE INDEX "IDX_alert_rule_org_id_dashboard_uid_panel_id" ON "alert_rule" ("org_id","dashboard_uid","panel_id");	t		2024-05-15 15:35:06
341	add rule_group_idx column to alert_rule	alter table "alert_rule" ADD COLUMN "rule_group_idx" INTEGER NOT NULL DEFAULT 1 	t		2024-05-15 15:35:06
342	add is_paused column to alert_rule table	alter table "alert_rule" ADD COLUMN "is_paused" BOOL NOT NULL DEFAULT false 	t		2024-05-15 15:35:06
343	fix is_paused column for alert_rule table	ALTER TABLE alert_rule ALTER COLUMN is_paused SET DEFAULT false;\nUPDATE alert_rule SET is_paused = false;	t		2024-05-15 15:35:06
344	create alert_rule_version table	CREATE TABLE IF NOT EXISTS "alert_rule_version" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "rule_org_id" BIGINT NOT NULL\n, "rule_uid" VARCHAR(40) NOT NULL DEFAULT 0\n, "rule_namespace_uid" VARCHAR(40) NOT NULL\n, "rule_group" VARCHAR(190) NOT NULL\n, "parent_version" INTEGER NOT NULL\n, "restored_from" INTEGER NOT NULL\n, "version" INTEGER NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "title" VARCHAR(190) NOT NULL\n, "condition" VARCHAR(190) NOT NULL\n, "data" TEXT NOT NULL\n, "interval_seconds" BIGINT NOT NULL\n, "no_data_state" VARCHAR(15) NOT NULL DEFAULT 'NoData'\n, "exec_err_state" VARCHAR(15) NOT NULL DEFAULT 'Alerting'\n);	t		2024-05-15 15:35:06
345	add index in alert_rule_version table on rule_org_id, rule_uid and version columns	CREATE UNIQUE INDEX "UQE_alert_rule_version_rule_org_id_rule_uid_version" ON "alert_rule_version" ("rule_org_id","rule_uid","version");	t		2024-05-15 15:35:06
346	add index in alert_rule_version table on rule_org_id, rule_namespace_uid and rule_group columns	CREATE INDEX "IDX_alert_rule_version_rule_org_id_rule_namespace_uid_rule_group" ON "alert_rule_version" ("rule_org_id","rule_namespace_uid","rule_group");	t		2024-05-15 15:35:06
347	alter alert_rule_version table data column to mediumtext in mysql	SELECT 0;	t		2024-05-15 15:35:06
348	add column for to alert_rule_version	alter table "alert_rule_version" ADD COLUMN "for" BIGINT NOT NULL DEFAULT 0 	t		2024-05-15 15:35:06
349	add column annotations to alert_rule_version	alter table "alert_rule_version" ADD COLUMN "annotations" TEXT NULL 	t		2024-05-15 15:35:06
350	add column labels to alert_rule_version	alter table "alert_rule_version" ADD COLUMN "labels" TEXT NULL 	t		2024-05-15 15:35:06
351	add rule_group_idx column to alert_rule_version	alter table "alert_rule_version" ADD COLUMN "rule_group_idx" INTEGER NOT NULL DEFAULT 1 	t		2024-05-15 15:35:06
352	add is_paused column to alert_rule_versions table	alter table "alert_rule_version" ADD COLUMN "is_paused" BOOL NOT NULL DEFAULT false 	t		2024-05-15 15:35:06
353	fix is_paused column for alert_rule_version table	ALTER TABLE alert_rule_version ALTER COLUMN is_paused SET DEFAULT false;\nUPDATE alert_rule_version SET is_paused = false;	t		2024-05-15 15:35:06
354	create_alert_configuration_table	CREATE TABLE IF NOT EXISTS "alert_configuration" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "alertmanager_configuration" TEXT NOT NULL\n, "configuration_version" VARCHAR(3) NOT NULL\n, "created_at" INTEGER NOT NULL\n);	t		2024-05-15 15:35:06
355	Add column default in alert_configuration	alter table "alert_configuration" ADD COLUMN "default" BOOL NOT NULL DEFAULT false 	t		2024-05-15 15:35:06
356	alert alert_configuration alertmanager_configuration column from TEXT to MEDIUMTEXT if mysql	SELECT 0;	t		2024-05-15 15:35:06
357	add column org_id in alert_configuration	alter table "alert_configuration" ADD COLUMN "org_id" BIGINT NOT NULL DEFAULT 0 	t		2024-05-15 15:35:06
358	add index in alert_configuration table on org_id column	CREATE INDEX "IDX_alert_configuration_org_id" ON "alert_configuration" ("org_id");	t		2024-05-15 15:35:06
359	add configuration_hash column to alert_configuration	alter table "alert_configuration" ADD COLUMN "configuration_hash" VARCHAR(32) NOT NULL DEFAULT 'not-yet-calculated' 	t		2024-05-15 15:35:06
360	create_ngalert_configuration_table	CREATE TABLE IF NOT EXISTS "ngalert_configuration" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "alertmanagers" TEXT NULL\n, "created_at" INTEGER NOT NULL\n, "updated_at" INTEGER NOT NULL\n);	t		2024-05-15 15:35:06
361	add index in ngalert_configuration on org_id column	CREATE UNIQUE INDEX "UQE_ngalert_configuration_org_id" ON "ngalert_configuration" ("org_id");	t		2024-05-15 15:35:07
362	add column send_alerts_to in ngalert_configuration	alter table "ngalert_configuration" ADD COLUMN "send_alerts_to" SMALLINT NOT NULL DEFAULT 0 	t		2024-05-15 15:35:07
363	create provenance_type table	CREATE TABLE IF NOT EXISTS "provenance_type" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "record_key" VARCHAR(190) NOT NULL\n, "record_type" VARCHAR(190) NOT NULL\n, "provenance" VARCHAR(190) NOT NULL\n);	t		2024-05-15 15:35:07
364	add index to uniquify (record_key, record_type, org_id) columns	CREATE UNIQUE INDEX "UQE_provenance_type_record_type_record_key_org_id" ON "provenance_type" ("record_type","record_key","org_id");	t		2024-05-15 15:35:07
365	create alert_image table	CREATE TABLE IF NOT EXISTS "alert_image" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "token" VARCHAR(190) NOT NULL\n, "path" VARCHAR(190) NOT NULL\n, "url" VARCHAR(190) NOT NULL\n, "created_at" TIMESTAMP NOT NULL\n, "expires_at" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:35:07
366	add unique index on token to alert_image table	CREATE UNIQUE INDEX "UQE_alert_image_token" ON "alert_image" ("token");	t		2024-05-15 15:35:07
367	support longer URLs in alert_image table	ALTER TABLE alert_image ALTER COLUMN url TYPE VARCHAR(2048);	t		2024-05-15 15:35:07
368	create_alert_configuration_history_table	CREATE TABLE IF NOT EXISTS "alert_configuration_history" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL DEFAULT 0\n, "alertmanager_configuration" TEXT NOT NULL\n, "configuration_hash" VARCHAR(32) NOT NULL DEFAULT 'not-yet-calculated'\n, "configuration_version" VARCHAR(3) NOT NULL\n, "created_at" INTEGER NOT NULL\n, "default" BOOL NOT NULL DEFAULT false\n);	t		2024-05-15 15:35:07
369	drop non-unique orgID index on alert_configuration	DROP INDEX "IDX_alert_configuration_org_id" CASCADE	t		2024-05-15 15:35:08
370	drop unique orgID index on alert_configuration if exists	DROP INDEX "UQE_alert_configuration_org_id" CASCADE	t		2024-05-15 15:35:08
371	extract alertmanager configuration history to separate table	code migration	t		2024-05-15 15:35:08
372	add unique index on orgID to alert_configuration	CREATE UNIQUE INDEX "UQE_alert_configuration_org_id" ON "alert_configuration" ("org_id");	t		2024-05-15 15:35:08
373	add last_applied column to alert_configuration_history	alter table "alert_configuration_history" ADD COLUMN "last_applied" INTEGER NOT NULL DEFAULT 0 	t		2024-05-15 15:35:08
374	move dashboard alerts to unified alerting	code migration	t		2024-05-15 15:35:08
375	create library_element table v1	CREATE TABLE IF NOT EXISTS "library_element" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "folder_id" BIGINT NOT NULL\n, "uid" VARCHAR(40) NOT NULL\n, "name" VARCHAR(150) NOT NULL\n, "kind" BIGINT NOT NULL\n, "type" VARCHAR(40) NOT NULL\n, "description" VARCHAR(255) NOT NULL\n, "model" TEXT NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "created_by" BIGINT NOT NULL\n, "updated" TIMESTAMP NOT NULL\n, "updated_by" BIGINT NOT NULL\n, "version" BIGINT NOT NULL\n);	t		2024-05-15 15:35:08
376	add index library_element org_id-folder_id-name-kind	CREATE UNIQUE INDEX "UQE_library_element_org_id_folder_id_name_kind" ON "library_element" ("org_id","folder_id","name","kind");	t		2024-05-15 15:35:08
377	create library_element_connection table v1	CREATE TABLE IF NOT EXISTS "library_element_connection" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "element_id" BIGINT NOT NULL\n, "kind" BIGINT NOT NULL\n, "connection_id" BIGINT NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "created_by" BIGINT NOT NULL\n);	t		2024-05-15 15:35:08
378	add index library_element_connection element_id-kind-connection_id	CREATE UNIQUE INDEX "UQE_library_element_connection_element_id_kind_connection_id" ON "library_element_connection" ("element_id","kind","connection_id");	t		2024-05-15 15:35:08
379	add unique index library_element org_id_uid	CREATE UNIQUE INDEX "UQE_library_element_org_id_uid" ON "library_element" ("org_id","uid");	t		2024-05-15 15:35:08
380	increase max description length to 2048	ALTER TABLE "library_element" ALTER "description" TYPE VARCHAR(2048);	t		2024-05-15 15:35:08
381	alter library_element model to mediumtext	SELECT 0;	t		2024-05-15 15:35:08
382	clone move dashboard alerts to unified alerting	code migration	t		2024-05-15 15:35:08
383	create data_keys table	CREATE TABLE IF NOT EXISTS "data_keys" (\n"name" VARCHAR(100) PRIMARY KEY NOT NULL\n, "active" BOOL NOT NULL\n, "scope" VARCHAR(30) NOT NULL\n, "provider" VARCHAR(50) NOT NULL\n, "encrypted_data" BYTEA NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:35:08
384	create secrets table	CREATE TABLE IF NOT EXISTS "secrets" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "namespace" VARCHAR(255) NOT NULL\n, "type" VARCHAR(255) NOT NULL\n, "value" TEXT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:35:08
385	rename data_keys name column to id	ALTER TABLE "data_keys" RENAME COLUMN "name" TO "id"	t		2024-05-15 15:35:08
386	add name column into data_keys	alter table "data_keys" ADD COLUMN "name" VARCHAR(100) NOT NULL DEFAULT '' 	t		2024-05-15 15:35:08
387	copy data_keys id column values into name	UPDATE data_keys SET name = id	t		2024-05-15 15:35:08
388	rename data_keys name column to label	ALTER TABLE "data_keys" RENAME COLUMN "name" TO "label"	t		2024-05-15 15:35:08
389	rename data_keys id column back to name	ALTER TABLE "data_keys" RENAME COLUMN "id" TO "name"	t		2024-05-15 15:35:08
390	create kv_store table v1	CREATE TABLE IF NOT EXISTS "kv_store" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "namespace" VARCHAR(190) NOT NULL\n, "key" VARCHAR(190) NOT NULL\n, "value" TEXT NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:35:08
391	add index kv_store.org_id-namespace-key	CREATE UNIQUE INDEX "UQE_kv_store_org_id_namespace_key" ON "kv_store" ("org_id","namespace","key");	t		2024-05-15 15:35:08
392	update dashboard_uid and panel_id from existing annotations	set dashboard_uid and panel_id migration	t		2024-05-15 15:35:09
393	create permission table	CREATE TABLE IF NOT EXISTS "permission" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "role_id" BIGINT NOT NULL\n, "action" VARCHAR(190) NOT NULL\n, "scope" VARCHAR(190) NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:35:09
394	add unique index permission.role_id	CREATE INDEX "IDX_permission_role_id" ON "permission" ("role_id");	t		2024-05-15 15:35:09
395	add unique index role_id_action_scope	CREATE UNIQUE INDEX "UQE_permission_role_id_action_scope" ON "permission" ("role_id","action","scope");	t		2024-05-15 15:35:09
396	create role table	CREATE TABLE IF NOT EXISTS "role" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "name" VARCHAR(190) NOT NULL\n, "description" TEXT NULL\n, "version" BIGINT NOT NULL\n, "org_id" BIGINT NOT NULL\n, "uid" VARCHAR(40) NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:35:09
397	add column display_name	alter table "role" ADD COLUMN "display_name" VARCHAR(190) NULL 	t		2024-05-15 15:35:10
398	add column group_name	alter table "role" ADD COLUMN "group_name" VARCHAR(190) NULL 	t		2024-05-15 15:35:10
399	add index role.org_id	CREATE INDEX "IDX_role_org_id" ON "role" ("org_id");	t		2024-05-15 15:35:10
400	add unique index role_org_id_name	CREATE UNIQUE INDEX "UQE_role_org_id_name" ON "role" ("org_id","name");	t		2024-05-15 15:35:10
401	add index role_org_id_uid	CREATE UNIQUE INDEX "UQE_role_org_id_uid" ON "role" ("org_id","uid");	t		2024-05-15 15:35:10
402	create team role table	CREATE TABLE IF NOT EXISTS "team_role" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "team_id" BIGINT NOT NULL\n, "role_id" BIGINT NOT NULL\n, "created" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:35:10
403	add index team_role.org_id	CREATE INDEX "IDX_team_role_org_id" ON "team_role" ("org_id");	t		2024-05-15 15:35:10
404	add unique index team_role_org_id_team_id_role_id	CREATE UNIQUE INDEX "UQE_team_role_org_id_team_id_role_id" ON "team_role" ("org_id","team_id","role_id");	t		2024-05-15 15:35:11
405	add index team_role.team_id	CREATE INDEX "IDX_team_role_team_id" ON "team_role" ("team_id");	t		2024-05-15 15:35:11
406	create user role table	CREATE TABLE IF NOT EXISTS "user_role" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "org_id" BIGINT NOT NULL\n, "user_id" BIGINT NOT NULL\n, "role_id" BIGINT NOT NULL\n, "created" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:35:11
407	add index user_role.org_id	CREATE INDEX "IDX_user_role_org_id" ON "user_role" ("org_id");	t		2024-05-15 15:35:11
408	add unique index user_role_org_id_user_id_role_id	CREATE UNIQUE INDEX "UQE_user_role_org_id_user_id_role_id" ON "user_role" ("org_id","user_id","role_id");	t		2024-05-15 15:35:11
409	add index user_role.user_id	CREATE INDEX "IDX_user_role_user_id" ON "user_role" ("user_id");	t		2024-05-15 15:35:11
410	create builtin role table	CREATE TABLE IF NOT EXISTS "builtin_role" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "role" VARCHAR(190) NOT NULL\n, "role_id" BIGINT NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:35:11
411	add index builtin_role.role_id	CREATE INDEX "IDX_builtin_role_role_id" ON "builtin_role" ("role_id");	t		2024-05-15 15:35:11
412	add index builtin_role.name	CREATE INDEX "IDX_builtin_role_role" ON "builtin_role" ("role");	t		2024-05-15 15:35:11
413	Add column org_id to builtin_role table	alter table "builtin_role" ADD COLUMN "org_id" BIGINT NOT NULL DEFAULT 0 	t		2024-05-15 15:35:12
414	add index builtin_role.org_id	CREATE INDEX "IDX_builtin_role_org_id" ON "builtin_role" ("org_id");	t		2024-05-15 15:35:12
415	add unique index builtin_role_org_id_role_id_role	CREATE UNIQUE INDEX "UQE_builtin_role_org_id_role_id_role" ON "builtin_role" ("org_id","role_id","role");	t		2024-05-15 15:35:12
416	Remove unique index role_org_id_uid	DROP INDEX "UQE_role_org_id_uid" CASCADE	t		2024-05-15 15:35:12
417	add unique index role.uid	CREATE UNIQUE INDEX "UQE_role_uid" ON "role" ("uid");	t		2024-05-15 15:35:12
418	create seed assignment table	CREATE TABLE IF NOT EXISTS "seed_assignment" (\n"builtin_role" VARCHAR(190) NOT NULL\n, "role_name" VARCHAR(190) NOT NULL\n);	t		2024-05-15 15:35:12
419	add unique index builtin_role_role_name	CREATE UNIQUE INDEX "UQE_seed_assignment_builtin_role_role_name" ON "seed_assignment" ("builtin_role","role_name");	t		2024-05-15 15:35:12
420	add column hidden to role table	alter table "role" ADD COLUMN "hidden" BOOL NOT NULL DEFAULT false 	t		2024-05-15 15:35:12
421	create query_history table v1	CREATE TABLE IF NOT EXISTS "query_history" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "uid" VARCHAR(40) NOT NULL\n, "org_id" BIGINT NOT NULL\n, "datasource_uid" VARCHAR(40) NOT NULL\n, "created_by" INTEGER NOT NULL\n, "created_at" INTEGER NOT NULL\n, "comment" TEXT NOT NULL\n, "queries" TEXT NOT NULL\n);	t		2024-05-15 15:35:12
422	add index query_history.org_id-created_by-datasource_uid	CREATE INDEX "IDX_query_history_org_id_created_by_datasource_uid" ON "query_history" ("org_id","created_by","datasource_uid");	t		2024-05-15 15:35:13
423	alter table query_history alter column created_by type to bigint	ALTER TABLE query_history ALTER COLUMN created_by TYPE BIGINT;	t		2024-05-15 15:35:13
424	teams permissions migration	code migration	t		2024-05-15 15:35:13
425	dashboard permissions	code migration	t		2024-05-15 15:35:13
426	dashboard permissions uid scopes	code migration	t		2024-05-15 15:35:13
427	drop managed folder create actions	code migration	t		2024-05-15 15:35:13
428	alerting notification permissions	code migration	t		2024-05-15 15:35:13
429	create query_history_star table v1	CREATE TABLE IF NOT EXISTS "query_history_star" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "query_uid" VARCHAR(40) NOT NULL\n, "user_id" INTEGER NOT NULL\n);	t		2024-05-15 15:35:13
430	add index query_history.user_id-query_uid	CREATE UNIQUE INDEX "UQE_query_history_star_user_id_query_uid" ON "query_history_star" ("user_id","query_uid");	t		2024-05-15 15:35:13
431	add column org_id in query_history_star	alter table "query_history_star" ADD COLUMN "org_id" BIGINT NOT NULL DEFAULT 1 	t		2024-05-15 15:35:13
432	alter table query_history_star_mig column user_id type to bigint	ALTER TABLE query_history_star ALTER COLUMN user_id TYPE BIGINT;	t		2024-05-15 15:35:13
433	create correlation table v1	CREATE TABLE IF NOT EXISTS "correlation" (\n"uid" VARCHAR(40) NOT NULL\n, "source_uid" VARCHAR(40) NOT NULL\n, "target_uid" VARCHAR(40) NULL\n, "label" TEXT NOT NULL\n, "description" TEXT NOT NULL\n, PRIMARY KEY ( "uid","source_uid" ));	t		2024-05-15 15:35:13
434	add index correlations.uid	CREATE INDEX "IDX_correlation_uid" ON "correlation" ("uid");	t		2024-05-15 15:35:13
435	add index correlations.source_uid	CREATE INDEX "IDX_correlation_source_uid" ON "correlation" ("source_uid");	t		2024-05-15 15:35:13
436	add correlation config column	alter table "correlation" ADD COLUMN "config" TEXT NULL 	t		2024-05-15 15:35:13
437	create entity_events table	CREATE TABLE IF NOT EXISTS "entity_event" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "entity_id" VARCHAR(1024) NOT NULL\n, "event_type" VARCHAR(8) NOT NULL\n, "created" BIGINT NOT NULL\n);	t		2024-05-15 15:35:13
438	create dashboard public config v1	CREATE TABLE IF NOT EXISTS "dashboard_public_config" (\n"uid" VARCHAR(40) PRIMARY KEY NOT NULL\n, "dashboard_uid" VARCHAR(40) NOT NULL\n, "org_id" BIGINT NOT NULL\n, "time_settings" TEXT NOT NULL\n, "refresh_rate" INTEGER NOT NULL DEFAULT 30\n, "template_variables" TEXT NULL\n);	t		2024-05-15 15:35:13
439	drop index UQE_dashboard_public_config_uid - v1	DROP INDEX "UQE_dashboard_public_config_uid" CASCADE	t		2024-05-15 15:35:13
440	drop index IDX_dashboard_public_config_org_id_dashboard_uid - v1	DROP INDEX "IDX_dashboard_public_config_org_id_dashboard_uid" CASCADE	t		2024-05-15 15:35:13
441	Drop old dashboard public config table	DROP TABLE IF EXISTS "dashboard_public_config"	t		2024-05-15 15:35:13
442	recreate dashboard public config v1	CREATE TABLE IF NOT EXISTS "dashboard_public_config" (\n"uid" VARCHAR(40) PRIMARY KEY NOT NULL\n, "dashboard_uid" VARCHAR(40) NOT NULL\n, "org_id" BIGINT NOT NULL\n, "time_settings" TEXT NOT NULL\n, "refresh_rate" INTEGER NOT NULL DEFAULT 30\n, "template_variables" TEXT NULL\n);	t		2024-05-15 15:35:13
443	create index UQE_dashboard_public_config_uid - v1	CREATE UNIQUE INDEX "UQE_dashboard_public_config_uid" ON "dashboard_public_config" ("uid");	t		2024-05-15 15:35:13
444	create index IDX_dashboard_public_config_org_id_dashboard_uid - v1	CREATE INDEX "IDX_dashboard_public_config_org_id_dashboard_uid" ON "dashboard_public_config" ("org_id","dashboard_uid");	t		2024-05-15 15:35:13
445	drop index UQE_dashboard_public_config_uid - v2	DROP INDEX "UQE_dashboard_public_config_uid" CASCADE	t		2024-05-15 15:35:14
446	drop index IDX_dashboard_public_config_org_id_dashboard_uid - v2	DROP INDEX "IDX_dashboard_public_config_org_id_dashboard_uid" CASCADE	t		2024-05-15 15:35:14
447	Drop public config table	DROP TABLE IF EXISTS "dashboard_public_config"	t		2024-05-15 15:35:14
448	Recreate dashboard public config v2	CREATE TABLE IF NOT EXISTS "dashboard_public_config" (\n"uid" VARCHAR(40) PRIMARY KEY NOT NULL\n, "dashboard_uid" VARCHAR(40) NOT NULL\n, "org_id" BIGINT NOT NULL\n, "time_settings" TEXT NULL\n, "template_variables" TEXT NULL\n, "access_token" VARCHAR(32) NOT NULL\n, "created_by" INTEGER NOT NULL\n, "updated_by" INTEGER NULL\n, "created_at" TIMESTAMP NOT NULL\n, "updated_at" TIMESTAMP NULL\n, "is_enabled" BOOL NOT NULL DEFAULT false\n);	t		2024-05-15 15:35:14
449	create index UQE_dashboard_public_config_uid - v2	CREATE UNIQUE INDEX "UQE_dashboard_public_config_uid" ON "dashboard_public_config" ("uid");	t		2024-05-15 15:35:14
450	create index IDX_dashboard_public_config_org_id_dashboard_uid - v2	CREATE INDEX "IDX_dashboard_public_config_org_id_dashboard_uid" ON "dashboard_public_config" ("org_id","dashboard_uid");	t		2024-05-15 15:35:14
451	create index UQE_dashboard_public_config_access_token - v2	CREATE UNIQUE INDEX "UQE_dashboard_public_config_access_token" ON "dashboard_public_config" ("access_token");	t		2024-05-15 15:35:14
452	Rename table dashboard_public_config to dashboard_public - v2	ALTER TABLE "dashboard_public_config" RENAME TO "dashboard_public"	t		2024-05-15 15:35:14
453	add annotations_enabled column	alter table "dashboard_public" ADD COLUMN "annotations_enabled" BOOL NOT NULL DEFAULT false 	t		2024-05-15 15:35:14
454	add time_selection_enabled column	alter table "dashboard_public" ADD COLUMN "time_selection_enabled" BOOL NOT NULL DEFAULT false 	t		2024-05-15 15:35:14
455	delete orphaned public dashboards	DELETE FROM dashboard_public WHERE dashboard_uid NOT IN (SELECT uid FROM dashboard)	t		2024-05-15 15:35:14
456	add share column	alter table "dashboard_public" ADD COLUMN "share" VARCHAR(64) NOT NULL DEFAULT 'public' 	t		2024-05-15 15:35:14
457	backfill empty share column fields with default of public	UPDATE dashboard_public SET share='public' WHERE share=''	t		2024-05-15 15:35:14
458	create default alerting folders	code migration	t		2024-05-15 15:35:14
459	create file table	CREATE TABLE IF NOT EXISTS "file" (\n"path" VARCHAR(1024) NOT NULL\n, "path_hash" VARCHAR(64) NOT NULL\n, "parent_folder_path_hash" VARCHAR(64) NOT NULL\n, "contents" BYTEA NOT NULL\n, "etag" VARCHAR(32) NOT NULL\n, "cache_control" VARCHAR(128) NOT NULL\n, "content_disposition" VARCHAR(128) NOT NULL\n, "updated" TIMESTAMP NOT NULL\n, "created" TIMESTAMP NOT NULL\n, "size" BIGINT NOT NULL\n, "mime_type" VARCHAR(255) NOT NULL\n);	t		2024-05-15 15:35:14
460	file table idx: path natural pk	CREATE UNIQUE INDEX "UQE_file_path_hash" ON "file" ("path_hash");	t		2024-05-15 15:35:14
461	file table idx: parent_folder_path_hash fast folder retrieval	CREATE INDEX "IDX_file_parent_folder_path_hash" ON "file" ("parent_folder_path_hash");	t		2024-05-15 15:35:14
462	create file_meta table	CREATE TABLE IF NOT EXISTS "file_meta" (\n"path_hash" VARCHAR(64) NOT NULL\n, "key" VARCHAR(191) NOT NULL\n, "value" VARCHAR(1024) NOT NULL\n);	t		2024-05-15 15:35:15
463	file table idx: path key	CREATE UNIQUE INDEX "UQE_file_meta_path_hash_key" ON "file_meta" ("path_hash","key");	t		2024-05-15 15:35:15
464	set path collation in file table	ALTER TABLE file ALTER COLUMN path TYPE VARCHAR(1024) COLLATE "C";	t		2024-05-15 15:35:15
465	managed permissions migration	code migration	t		2024-05-15 15:35:15
466	managed folder permissions alert actions migration	code migration	t		2024-05-15 15:35:15
467	RBAC action name migrator	code migration	t		2024-05-15 15:35:15
468	Add UID column to playlist	alter table "playlist" ADD COLUMN "uid" VARCHAR(80) NOT NULL DEFAULT 0 	t		2024-05-15 15:35:15
469	Update uid column values in playlist	UPDATE playlist SET uid=id::text;	t		2024-05-15 15:35:15
470	Add index for uid in playlist	CREATE UNIQUE INDEX "UQE_playlist_org_id_uid" ON "playlist" ("org_id","uid");	t		2024-05-15 15:35:15
471	update group index for alert rules	code migration	t		2024-05-15 15:35:15
472	managed folder permissions alert actions repeated migration	code migration	t		2024-05-15 15:35:15
473	admin only folder/dashboard permission	code migration	t		2024-05-15 15:35:15
474	add action column to seed_assignment	alter table "seed_assignment" ADD COLUMN "action" VARCHAR(190) NULL 	t		2024-05-15 15:35:15
475	add scope column to seed_assignment	alter table "seed_assignment" ADD COLUMN "scope" VARCHAR(190) NULL 	t		2024-05-15 15:35:15
476	remove unique index builtin_role_role_name before nullable update	DROP INDEX "UQE_seed_assignment_builtin_role_role_name" CASCADE	t		2024-05-15 15:35:15
477	update seed_assignment role_name column to nullable	ALTER TABLE `seed_assignment` ALTER COLUMN role_name DROP NOT NULL;	t		2024-05-15 15:35:15
478	add unique index builtin_role_name back	CREATE UNIQUE INDEX "UQE_seed_assignment_builtin_role_role_name" ON "seed_assignment" ("builtin_role","role_name");	t		2024-05-15 15:35:15
479	add unique index builtin_role_action_scope	CREATE UNIQUE INDEX "UQE_seed_assignment_builtin_role_action_scope" ON "seed_assignment" ("builtin_role","action","scope");	t		2024-05-15 15:35:15
480	add primary key to seed_assigment	code migration	t		2024-05-15 15:35:15
481	managed folder permissions alert actions repeated fixed migration	code migration	t		2024-05-15 15:35:15
482	migrate external alertmanagers to datsourcse	migrate external alertmanagers to datasource	t		2024-05-15 15:35:15
483	create folder table	CREATE TABLE IF NOT EXISTS "folder" (\n"id" SERIAL PRIMARY KEY  NOT NULL\n, "uid" VARCHAR(40) NOT NULL\n, "org_id" BIGINT NOT NULL\n, "title" VARCHAR(255) NOT NULL\n, "description" VARCHAR(255) NULL\n, "parent_uid" VARCHAR(40) NULL\n, "created" TIMESTAMP NOT NULL\n, "updated" TIMESTAMP NOT NULL\n);	t		2024-05-15 15:35:15
484	Add index for parent_uid	CREATE INDEX "IDX_folder_parent_uid_org_id" ON "folder" ("parent_uid","org_id");	t		2024-05-15 15:35:16
485	Add unique index for folder.uid and folder.org_id	CREATE UNIQUE INDEX "UQE_folder_uid_org_id" ON "folder" ("uid","org_id");	t		2024-05-15 15:35:16
486	Update folder title length	ALTER TABLE "folder" ALTER "title" TYPE VARCHAR(189);	t		2024-05-15 15:35:16
487	Add unique index for folder.title and folder.parent_uid	CREATE UNIQUE INDEX "UQE_folder_title_parent_uid" ON "folder" ("title","parent_uid");	t		2024-05-15 15:35:16
\.


--
-- TOC entry 4038 (class 0 OID 20860)
-- Dependencies: 287
-- Data for Name: ngalert_configuration; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.ngalert_configuration (id, org_id, alertmanagers, created_at, updated_at, send_alerts_to) FROM stdin;
\.


--
-- TOC entry 3969 (class 0 OID 20292)
-- Dependencies: 218
-- Data for Name: org; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.org (id, version, name, address1, address2, city, state, zip_code, country, billing_email, created, updated) FROM stdin;
1	0	Main Org.							\N	2024-05-15 15:35:16	2024-05-15 15:35:16
\.


--
-- TOC entry 3971 (class 0 OID 20302)
-- Dependencies: 220
-- Data for Name: org_user; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.org_user (id, org_id, user_id, role, created, updated) FROM stdin;
1	1	1	Admin	2024-05-15 15:35:16	2024-05-15 15:35:16
\.


--
-- TOC entry 4055 (class 0 OID 20951)
-- Dependencies: 304
-- Data for Name: permission; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.permission (id, role_id, action, scope, created, updated) FROM stdin;
1	1	dashboards.permissions:read	dashboards:uid:LsXOaz47z	2024-05-15 15:43:22	2024-05-15 15:43:22
2	1	dashboards.permissions:write	dashboards:uid:LsXOaz47z	2024-05-15 15:43:22	2024-05-15 15:43:22
3	1	dashboards:read	dashboards:uid:LsXOaz47z	2024-05-15 15:43:22	2024-05-15 15:43:22
4	1	dashboards:write	dashboards:uid:LsXOaz47z	2024-05-15 15:43:22	2024-05-15 15:43:22
5	1	dashboards:delete	dashboards:uid:LsXOaz47z	2024-05-15 15:43:22	2024-05-15 15:43:22
6	2	dashboards:read	dashboards:uid:LsXOaz47z	2024-05-15 15:43:22	2024-05-15 15:43:22
7	2	dashboards:write	dashboards:uid:LsXOaz47z	2024-05-15 15:43:22	2024-05-15 15:43:22
8	2	dashboards:delete	dashboards:uid:LsXOaz47z	2024-05-15 15:43:22	2024-05-15 15:43:22
9	3	dashboards:read	dashboards:uid:LsXOaz47z	2024-05-15 15:43:22	2024-05-15 15:43:22
10	1	dashboards:read	dashboards:uid:78X6BmvMk	2024-05-15 15:43:50	2024-05-15 15:43:50
11	1	dashboards:write	dashboards:uid:78X6BmvMk	2024-05-15 15:43:50	2024-05-15 15:43:50
12	1	dashboards:delete	dashboards:uid:78X6BmvMk	2024-05-15 15:43:50	2024-05-15 15:43:50
13	1	dashboards.permissions:read	dashboards:uid:78X6BmvMk	2024-05-15 15:43:50	2024-05-15 15:43:50
14	1	dashboards.permissions:write	dashboards:uid:78X6BmvMk	2024-05-15 15:43:50	2024-05-15 15:43:50
15	2	dashboards:read	dashboards:uid:78X6BmvMk	2024-05-15 15:43:50	2024-05-15 15:43:50
16	2	dashboards:write	dashboards:uid:78X6BmvMk	2024-05-15 15:43:50	2024-05-15 15:43:50
17	2	dashboards:delete	dashboards:uid:78X6BmvMk	2024-05-15 15:43:50	2024-05-15 15:43:50
18	3	dashboards:read	dashboards:uid:78X6BmvMk	2024-05-15 15:43:50	2024-05-15 15:43:50
19	1	dashboards.permissions:write	dashboards:uid:zUqcvfZ7z	2024-05-15 15:44:09	2024-05-15 15:44:09
20	1	dashboards:read	dashboards:uid:zUqcvfZ7z	2024-05-15 15:44:09	2024-05-15 15:44:09
21	1	dashboards:write	dashboards:uid:zUqcvfZ7z	2024-05-15 15:44:09	2024-05-15 15:44:09
22	1	dashboards:delete	dashboards:uid:zUqcvfZ7z	2024-05-15 15:44:09	2024-05-15 15:44:09
23	1	dashboards.permissions:read	dashboards:uid:zUqcvfZ7z	2024-05-15 15:44:09	2024-05-15 15:44:09
24	2	dashboards:write	dashboards:uid:zUqcvfZ7z	2024-05-15 15:44:09	2024-05-15 15:44:09
25	2	dashboards:delete	dashboards:uid:zUqcvfZ7z	2024-05-15 15:44:09	2024-05-15 15:44:09
26	2	dashboards:read	dashboards:uid:zUqcvfZ7z	2024-05-15 15:44:09	2024-05-15 15:44:09
27	3	dashboards:read	dashboards:uid:zUqcvfZ7z	2024-05-15 15:44:09	2024-05-15 15:44:09
\.


--
-- TOC entry 3990 (class 0 OID 20492)
-- Dependencies: 239
-- Data for Name: playlist; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.playlist (id, name, "interval", org_id, uid) FROM stdin;
\.


--
-- TOC entry 3992 (class 0 OID 20504)
-- Dependencies: 241
-- Data for Name: playlist_item; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.playlist_item (id, playlist_id, type, value, title, "order") FROM stdin;
\.


--
-- TOC entry 3987 (class 0 OID 20474)
-- Dependencies: 236
-- Data for Name: plugin_setting; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.plugin_setting (id, org_id, plugin_id, enabled, pinned, json_data, secure_json_data, created, updated, plugin_version) FROM stdin;
\.


--
-- TOC entry 3994 (class 0 OID 20515)
-- Dependencies: 243
-- Data for Name: preferences; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.preferences (id, org_id, user_id, version, home_dashboard_id, timezone, theme, created, updated, team_id, week_start, json_data) FROM stdin;
\.


--
-- TOC entry 4040 (class 0 OID 20871)
-- Dependencies: 289
-- Data for Name: provenance_type; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.provenance_type (id, org_id, record_key, record_type, provenance) FROM stdin;
\.


--
-- TOC entry 4066 (class 0 OID 21012)
-- Dependencies: 315
-- Data for Name: query_history; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.query_history (id, uid, org_id, datasource_uid, created_by, created_at, comment, queries) FROM stdin;
\.


--
-- TOC entry 4068 (class 0 OID 21039)
-- Dependencies: 317
-- Data for Name: query_history_star; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.query_history_star (id, query_uid, user_id, org_id) FROM stdin;
\.


--
-- TOC entry 3985 (class 0 OID 20465)
-- Dependencies: 234
-- Data for Name: quota; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.quota (id, org_id, user_id, target, "limit", created, updated) FROM stdin;
\.


--
-- TOC entry 4057 (class 0 OID 20960)
-- Dependencies: 306
-- Data for Name: role; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.role (id, name, description, version, org_id, uid, created, updated, display_name, group_name, hidden) FROM stdin;
1	managed:users:1:permissions		0	1	cdfe6eee-c123-4ec0-a09e-39e905069d33	2024-05-15 15:43:22	2024-05-15 15:43:22			f
2	managed:builtins:editor:permissions		0	1	ed2d7b85-6681-49c9-a0bf-ddbab2d26e59	2024-05-15 15:43:22	2024-05-15 15:43:22			f
3	managed:builtins:viewer:permissions		0	1	c0acf7f5-03c7-4bc7-b948-b03e7588db7d	2024-05-15 15:43:22	2024-05-15 15:43:22			f
\.


--
-- TOC entry 4051 (class 0 OID 20931)
-- Dependencies: 300
-- Data for Name: secrets; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.secrets (id, org_id, namespace, type, value, created, updated) FROM stdin;
1	1	PostgreSQL-JSON_SCADA	datasource	I1pEUmtabVkyTmpVdFpUZGlOQzAwWWpjNExXRTRZak10TkRObVpUbGtabUZpWW1NNCMqWVdWekxXTm1ZZypWTzZFc2tBSGgMqkb2sBuD2U2akcbxWm0r/w	2024-05-15 15:38:36	2024-05-15 15:42:11
\.


--
-- TOC entry 4064 (class 0 OID 21004)
-- Dependencies: 313
-- Data for Name: seed_assignment; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.seed_assignment (builtin_role, role_name, action, scope, id) FROM stdin;
\.


--
-- TOC entry 4024 (class 0 OID 20710)
-- Dependencies: 273
-- Data for Name: server_lock; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.server_lock (id, operation_uid, version, last_execution) FROM stdin;
1	cleanup expired auth tokens	1	1715798134
4	delete old login attempts	2	1715799480
\.


--
-- TOC entry 3988 (class 0 OID 20484)
-- Dependencies: 237
-- Data for Name: session; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.session (key, data, expiry) FROM stdin;
\.


--
-- TOC entry 4029 (class 0 OID 20738)
-- Dependencies: 278
-- Data for Name: short_url; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.short_url (id, org_id, uid, path, created_by, created_at, last_seen_at) FROM stdin;
\.


--
-- TOC entry 3967 (class 0 OID 20284)
-- Dependencies: 216
-- Data for Name: star; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.star (id, user_id, dashboard_id) FROM stdin;
1	1	3
\.


--
-- TOC entry 4018 (class 0 OID 20673)
-- Dependencies: 267
-- Data for Name: tag; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tag (id, key, value) FROM stdin;
\.


--
-- TOC entry 4012 (class 0 OID 20639)
-- Dependencies: 261
-- Data for Name: team; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.team (id, name, org_id, created, updated, email) FROM stdin;
\.


--
-- TOC entry 4014 (class 0 OID 20648)
-- Dependencies: 263
-- Data for Name: team_member; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.team_member (id, org_id, team_id, user_id, created, updated, external, permission) FROM stdin;
\.


--
-- TOC entry 4059 (class 0 OID 20972)
-- Dependencies: 308
-- Data for Name: team_role; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.team_role (id, org_id, team_id, role_id, created) FROM stdin;
\.


--
-- TOC entry 3965 (class 0 OID 20269)
-- Dependencies: 214
-- Data for Name: temp_user; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.temp_user (id, org_id, version, email, name, role, code, status, invited_by_user_id, email_sent, email_sent_on, remote_addr, created, updated) FROM stdin;
\.


--
-- TOC entry 4008 (class 0 OID 20621)
-- Dependencies: 257
-- Data for Name: test_data; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.test_data (id, metric1, metric2, value_big_int, value_double, value_float, value_int, time_epoch, time_date_time, time_time_stamp) FROM stdin;
\.


--
-- TOC entry 3963 (class 0 OID 20236)
-- Dependencies: 212
-- Data for Name: user; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public."user" (id, version, login, email, name, password, salt, rands, company, org_id, is_admin, email_verified, theme, created, updated, help_flags1, last_seen_at, is_disabled, is_service_account) FROM stdin;
1	0	admin	admin@localhost		7924c804508492bf9bf2fe917131861842312ae935c49b3ac3b355bfbba053714340e0ef13eeb92abcc2a66719ac85413968	gRxyIkNDEg	OuKpg0IDu6		1	t	f		2024-05-15 15:35:16	2024-05-15 15:35:16	0	2024-05-15 15:58:17	f	f
\.


--
-- TOC entry 4022 (class 0 OID 20698)
-- Dependencies: 271
-- Data for Name: user_auth; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_auth (id, user_id, auth_module, auth_id, created, o_auth_access_token, o_auth_refresh_token, o_auth_token_type, o_auth_expiry, o_auth_id_token) FROM stdin;
1	1	authproxy	admin	2024-05-15 15:38:01				\N	
\.


--
-- TOC entry 4026 (class 0 OID 20718)
-- Dependencies: 275
-- Data for Name: user_auth_token; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_auth_token (id, user_id, auth_token, prev_auth_token, user_agent, client_ip, auth_token_seen, seen_at, rotated_at, created_at, updated_at, revoked_at) FROM stdin;
\.


--
-- TOC entry 4061 (class 0 OID 20982)
-- Dependencies: 310
-- Data for Name: user_role; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_role (id, org_id, user_id, role_id, created) FROM stdin;
1	1	1	1	2024-05-15 15:43:22
\.


--
-- TOC entry 4139 (class 0 OID 0)
-- Dependencies: 292
-- Name: alert_configuration_history_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.alert_configuration_history_id_seq', 1, true);


--
-- TOC entry 4140 (class 0 OID 0)
-- Dependencies: 284
-- Name: alert_configuration_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.alert_configuration_id_seq', 1, true);


--
-- TOC entry 4141 (class 0 OID 0)
-- Dependencies: 244
-- Name: alert_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.alert_id_seq', 1, false);


--
-- TOC entry 4142 (class 0 OID 0)
-- Dependencies: 290
-- Name: alert_image_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.alert_image_id_seq', 1, false);


--
-- TOC entry 4143 (class 0 OID 0)
-- Dependencies: 248
-- Name: alert_notification_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.alert_notification_id_seq', 1, false);


--
-- TOC entry 4144 (class 0 OID 0)
-- Dependencies: 250
-- Name: alert_notification_state_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.alert_notification_state_id_seq', 1, false);


--
-- TOC entry 4145 (class 0 OID 0)
-- Dependencies: 280
-- Name: alert_rule_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.alert_rule_id_seq', 1, false);


--
-- TOC entry 4146 (class 0 OID 0)
-- Dependencies: 246
-- Name: alert_rule_tag_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.alert_rule_tag_id_seq', 1, false);


--
-- TOC entry 4147 (class 0 OID 0)
-- Dependencies: 282
-- Name: alert_rule_version_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.alert_rule_version_id_seq', 1, false);


--
-- TOC entry 4148 (class 0 OID 0)
-- Dependencies: 252
-- Name: annotation_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.annotation_id_seq', 1, false);


--
-- TOC entry 4149 (class 0 OID 0)
-- Dependencies: 254
-- Name: annotation_tag_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.annotation_tag_id_seq', 1, false);


--
-- TOC entry 4150 (class 0 OID 0)
-- Dependencies: 229
-- Name: api_key_id_seq1; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.api_key_id_seq1', 1, false);


--
-- TOC entry 4151 (class 0 OID 0)
-- Dependencies: 311
-- Name: builtin_role_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.builtin_role_id_seq', 2, true);


--
-- TOC entry 4152 (class 0 OID 0)
-- Dependencies: 264
-- Name: dashboard_acl_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.dashboard_acl_id_seq', 2, true);


--
-- TOC entry 4153 (class 0 OID 0)
-- Dependencies: 223
-- Name: dashboard_id_seq1; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.dashboard_id_seq1', 3, true);


--
-- TOC entry 4154 (class 0 OID 0)
-- Dependencies: 225
-- Name: dashboard_provisioning_id_seq1; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.dashboard_provisioning_id_seq1', 1, false);


--
-- TOC entry 4155 (class 0 OID 0)
-- Dependencies: 231
-- Name: dashboard_snapshot_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.dashboard_snapshot_id_seq', 1, false);


--
-- TOC entry 4156 (class 0 OID 0)
-- Dependencies: 221
-- Name: dashboard_tag_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.dashboard_tag_id_seq', 1, false);


--
-- TOC entry 4157 (class 0 OID 0)
-- Dependencies: 258
-- Name: dashboard_version_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.dashboard_version_id_seq', 10, true);


--
-- TOC entry 4158 (class 0 OID 0)
-- Dependencies: 227
-- Name: data_source_id_seq1; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.data_source_id_seq1', 1, true);


--
-- TOC entry 4159 (class 0 OID 0)
-- Dependencies: 319
-- Name: entity_event_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.entity_event_id_seq', 1, false);


--
-- TOC entry 4160 (class 0 OID 0)
-- Dependencies: 325
-- Name: folder_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.folder_id_seq', 1, false);


--
-- TOC entry 4161 (class 0 OID 0)
-- Dependencies: 301
-- Name: kv_store_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.kv_store_id_seq', 3, true);


--
-- TOC entry 4162 (class 0 OID 0)
-- Dependencies: 296
-- Name: library_element_connection_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.library_element_connection_id_seq', 1, false);


--
-- TOC entry 4163 (class 0 OID 0)
-- Dependencies: 294
-- Name: library_element_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.library_element_id_seq', 1, false);


--
-- TOC entry 4164 (class 0 OID 0)
-- Dependencies: 268
-- Name: login_attempt_id_seq1; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.login_attempt_id_seq1', 1, false);


--
-- TOC entry 4165 (class 0 OID 0)
-- Dependencies: 209
-- Name: migration_log_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.migration_log_id_seq', 487, true);


--
-- TOC entry 4166 (class 0 OID 0)
-- Dependencies: 286
-- Name: ngalert_configuration_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.ngalert_configuration_id_seq', 1, false);


--
-- TOC entry 4167 (class 0 OID 0)
-- Dependencies: 217
-- Name: org_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.org_id_seq', 1, true);


--
-- TOC entry 4168 (class 0 OID 0)
-- Dependencies: 219
-- Name: org_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.org_user_id_seq', 1, true);


--
-- TOC entry 4169 (class 0 OID 0)
-- Dependencies: 303
-- Name: permission_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.permission_id_seq', 27, true);


--
-- TOC entry 4170 (class 0 OID 0)
-- Dependencies: 238
-- Name: playlist_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.playlist_id_seq', 1, false);


--
-- TOC entry 4171 (class 0 OID 0)
-- Dependencies: 240
-- Name: playlist_item_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.playlist_item_id_seq', 1, false);


--
-- TOC entry 4172 (class 0 OID 0)
-- Dependencies: 235
-- Name: plugin_setting_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.plugin_setting_id_seq', 1, false);


--
-- TOC entry 4173 (class 0 OID 0)
-- Dependencies: 242
-- Name: preferences_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.preferences_id_seq', 1, false);


--
-- TOC entry 4174 (class 0 OID 0)
-- Dependencies: 288
-- Name: provenance_type_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.provenance_type_id_seq', 1, false);


--
-- TOC entry 4175 (class 0 OID 0)
-- Dependencies: 314
-- Name: query_history_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.query_history_id_seq', 1, false);


--
-- TOC entry 4176 (class 0 OID 0)
-- Dependencies: 316
-- Name: query_history_star_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.query_history_star_id_seq', 1, false);


--
-- TOC entry 4177 (class 0 OID 0)
-- Dependencies: 233
-- Name: quota_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.quota_id_seq', 1, false);


--
-- TOC entry 4178 (class 0 OID 0)
-- Dependencies: 305
-- Name: role_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.role_id_seq', 3, true);


--
-- TOC entry 4179 (class 0 OID 0)
-- Dependencies: 299
-- Name: secrets_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.secrets_id_seq', 1, true);


--
-- TOC entry 4180 (class 0 OID 0)
-- Dependencies: 324
-- Name: seed_assignment_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.seed_assignment_id_seq', 1, false);


--
-- TOC entry 4181 (class 0 OID 0)
-- Dependencies: 272
-- Name: server_lock_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.server_lock_id_seq', 4, true);


--
-- TOC entry 4182 (class 0 OID 0)
-- Dependencies: 277
-- Name: short_url_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.short_url_id_seq', 1, false);


--
-- TOC entry 4183 (class 0 OID 0)
-- Dependencies: 215
-- Name: star_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.star_id_seq', 1, true);


--
-- TOC entry 4184 (class 0 OID 0)
-- Dependencies: 266
-- Name: tag_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.tag_id_seq', 1, false);


--
-- TOC entry 4185 (class 0 OID 0)
-- Dependencies: 260
-- Name: team_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.team_id_seq', 1, false);


--
-- TOC entry 4186 (class 0 OID 0)
-- Dependencies: 262
-- Name: team_member_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.team_member_id_seq', 1, false);


--
-- TOC entry 4187 (class 0 OID 0)
-- Dependencies: 307
-- Name: team_role_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.team_role_id_seq', 1, false);


--
-- TOC entry 4188 (class 0 OID 0)
-- Dependencies: 213
-- Name: temp_user_id_seq1; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.temp_user_id_seq1', 1, false);


--
-- TOC entry 4189 (class 0 OID 0)
-- Dependencies: 256
-- Name: test_data_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.test_data_id_seq', 1, false);


--
-- TOC entry 4190 (class 0 OID 0)
-- Dependencies: 270
-- Name: user_auth_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.user_auth_id_seq', 1, true);


--
-- TOC entry 4191 (class 0 OID 0)
-- Dependencies: 274
-- Name: user_auth_token_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.user_auth_token_id_seq', 1, false);


--
-- TOC entry 4192 (class 0 OID 0)
-- Dependencies: 211
-- Name: user_id_seq1; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.user_id_seq1', 1, true);


--
-- TOC entry 4193 (class 0 OID 0)
-- Dependencies: 309
-- Name: user_role_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.user_role_id_seq', 1, true);


--
-- TOC entry 3752 (class 2606 OID 20901)
-- Name: alert_configuration_history alert_configuration_history_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert_configuration_history
    ADD CONSTRAINT alert_configuration_history_pkey PRIMARY KEY (id);


--
-- TOC entry 3741 (class 2606 OID 20854)
-- Name: alert_configuration alert_configuration_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert_configuration
    ADD CONSTRAINT alert_configuration_pkey PRIMARY KEY (id);


--
-- TOC entry 3750 (class 2606 OID 20888)
-- Name: alert_image alert_image_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert_image
    ADD CONSTRAINT alert_image_pkey PRIMARY KEY (id);


--
-- TOC entry 3728 (class 2606 OID 20799)
-- Name: alert_instance alert_instance_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert_instance
    ADD CONSTRAINT alert_instance_pkey PRIMARY KEY (rule_org_id, rule_uid, labels_hash);


--
-- TOC entry 3660 (class 2606 OID 20557)
-- Name: alert_notification alert_notification_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert_notification
    ADD CONSTRAINT alert_notification_pkey PRIMARY KEY (id);


--
-- TOC entry 3664 (class 2606 OID 20578)
-- Name: alert_notification_state alert_notification_state_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert_notification_state
    ADD CONSTRAINT alert_notification_state_pkey PRIMARY KEY (id);


--
-- TOC entry 3653 (class 2606 OID 20533)
-- Name: alert alert_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert
    ADD CONSTRAINT alert_pkey PRIMARY KEY (id);


--
-- TOC entry 3734 (class 2606 OID 20818)
-- Name: alert_rule alert_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert_rule
    ADD CONSTRAINT alert_rule_pkey PRIMARY KEY (id);


--
-- TOC entry 3657 (class 2606 OID 20547)
-- Name: alert_rule_tag alert_rule_tag_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert_rule_tag
    ADD CONSTRAINT alert_rule_tag_pkey PRIMARY KEY (id);


--
-- TOC entry 3738 (class 2606 OID 20839)
-- Name: alert_rule_version alert_rule_version_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.alert_rule_version
    ADD CONSTRAINT alert_rule_version_pkey PRIMARY KEY (id);


--
-- TOC entry 3673 (class 2606 OID 20591)
-- Name: annotation annotation_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.annotation
    ADD CONSTRAINT annotation_pkey PRIMARY KEY (id);


--
-- TOC entry 3676 (class 2606 OID 20609)
-- Name: annotation_tag annotation_tag_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.annotation_tag
    ADD CONSTRAINT annotation_tag_pkey PRIMARY KEY (id);


--
-- TOC entry 3626 (class 2606 OID 20433)
-- Name: api_key api_key_pkey1; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_key
    ADD CONSTRAINT api_key_pkey1 PRIMARY KEY (id);


--
-- TOC entry 3791 (class 2606 OID 20997)
-- Name: builtin_role builtin_role_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.builtin_role
    ADD CONSTRAINT builtin_role_pkey PRIMARY KEY (id);


--
-- TOC entry 3721 (class 2606 OID 20735)
-- Name: cache_data cache_data_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cache_data
    ADD CONSTRAINT cache_data_pkey PRIMARY KEY (cache_key);


--
-- TOC entry 3805 (class 2606 OID 21059)
-- Name: correlation correlation_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.correlation
    ADD CONSTRAINT correlation_pkey PRIMARY KEY (uid, source_uid);


--
-- TOC entry 3700 (class 2606 OID 20664)
-- Name: dashboard_acl dashboard_acl_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dashboard_acl
    ADD CONSTRAINT dashboard_acl_pkey PRIMARY KEY (id);


--
-- TOC entry 3611 (class 2606 OID 20339)
-- Name: dashboard dashboard_pkey1; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dashboard
    ADD CONSTRAINT dashboard_pkey1 PRIMARY KEY (id);


--
-- TOC entry 3615 (class 2606 OID 20380)
-- Name: dashboard_provisioning dashboard_provisioning_pkey1; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dashboard_provisioning
    ADD CONSTRAINT dashboard_provisioning_pkey1 PRIMARY KEY (id);


--
-- TOC entry 3812 (class 2606 OID 21096)
-- Name: dashboard_public dashboard_public_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dashboard_public
    ADD CONSTRAINT dashboard_public_config_pkey PRIMARY KEY (uid);


--
-- TOC entry 3631 (class 2606 OID 20457)
-- Name: dashboard_snapshot dashboard_snapshot_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dashboard_snapshot
    ADD CONSTRAINT dashboard_snapshot_pkey PRIMARY KEY (id);


--
-- TOC entry 3602 (class 2606 OID 20329)
-- Name: dashboard_tag dashboard_tag_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dashboard_tag
    ADD CONSTRAINT dashboard_tag_pkey PRIMARY KEY (id);


--
-- TOC entry 3682 (class 2606 OID 20635)
-- Name: dashboard_version dashboard_version_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dashboard_version
    ADD CONSTRAINT dashboard_version_pkey PRIMARY KEY (id);


--
-- TOC entry 3761 (class 2606 OID 20929)
-- Name: data_keys data_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.data_keys
    ADD CONSTRAINT data_keys_pkey PRIMARY KEY (name);


--
-- TOC entry 3621 (class 2606 OID 20405)
-- Name: data_source data_source_pkey1; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.data_source
    ADD CONSTRAINT data_source_pkey1 PRIMARY KEY (id);


--
-- TOC entry 3807 (class 2606 OID 21070)
-- Name: entity_event entity_event_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.entity_event
    ADD CONSTRAINT entity_event_pkey PRIMARY KEY (id);


--
-- TOC entry 3820 (class 2606 OID 21142)
-- Name: folder folder_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.folder
    ADD CONSTRAINT folder_pkey PRIMARY KEY (id);


--
-- TOC entry 3766 (class 2606 OID 20948)
-- Name: kv_store kv_store_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.kv_store
    ADD CONSTRAINT kv_store_pkey PRIMARY KEY (id);


--
-- TOC entry 3759 (class 2606 OID 20920)
-- Name: library_element_connection library_element_connection_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.library_element_connection
    ADD CONSTRAINT library_element_connection_pkey PRIMARY KEY (id);


--
-- TOC entry 3756 (class 2606 OID 20912)
-- Name: library_element library_element_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.library_element
    ADD CONSTRAINT library_element_pkey PRIMARY KEY (id);


--
-- TOC entry 3706 (class 2606 OID 20695)
-- Name: login_attempt login_attempt_pkey1; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.login_attempt
    ADD CONSTRAINT login_attempt_pkey1 PRIMARY KEY (id);


--
-- TOC entry 3577 (class 2606 OID 20223)
-- Name: migration_log migration_log_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.migration_log
    ADD CONSTRAINT migration_log_pkey PRIMARY KEY (id);


--
-- TOC entry 3744 (class 2606 OID 20867)
-- Name: ngalert_configuration ngalert_configuration_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ngalert_configuration
    ADD CONSTRAINT ngalert_configuration_pkey PRIMARY KEY (id);


--
-- TOC entry 3594 (class 2606 OID 20299)
-- Name: org org_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.org
    ADD CONSTRAINT org_pkey PRIMARY KEY (id);


--
-- TOC entry 3599 (class 2606 OID 20307)
-- Name: org_user org_user_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.org_user
    ADD CONSTRAINT org_user_pkey PRIMARY KEY (id);


--
-- TOC entry 3770 (class 2606 OID 20956)
-- Name: permission permission_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.permission
    ADD CONSTRAINT permission_pkey PRIMARY KEY (id);


--
-- TOC entry 3644 (class 2606 OID 20511)
-- Name: playlist_item playlist_item_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.playlist_item
    ADD CONSTRAINT playlist_item_pkey PRIMARY KEY (id);


--
-- TOC entry 3642 (class 2606 OID 20499)
-- Name: playlist playlist_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.playlist
    ADD CONSTRAINT playlist_pkey PRIMARY KEY (id);


--
-- TOC entry 3637 (class 2606 OID 20481)
-- Name: plugin_setting plugin_setting_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.plugin_setting
    ADD CONSTRAINT plugin_setting_pkey PRIMARY KEY (id);


--
-- TOC entry 3648 (class 2606 OID 20520)
-- Name: preferences preferences_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.preferences
    ADD CONSTRAINT preferences_pkey PRIMARY KEY (id);


--
-- TOC entry 3747 (class 2606 OID 20878)
-- Name: provenance_type provenance_type_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.provenance_type
    ADD CONSTRAINT provenance_type_pkey PRIMARY KEY (id);


--
-- TOC entry 3798 (class 2606 OID 21019)
-- Name: query_history query_history_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.query_history
    ADD CONSTRAINT query_history_pkey PRIMARY KEY (id);


--
-- TOC entry 3801 (class 2606 OID 21044)
-- Name: query_history_star query_history_star_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.query_history_star
    ADD CONSTRAINT query_history_star_pkey PRIMARY KEY (id);


--
-- TOC entry 3634 (class 2606 OID 20470)
-- Name: quota quota_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.quota
    ADD CONSTRAINT quota_pkey PRIMARY KEY (id);


--
-- TOC entry 3775 (class 2606 OID 20967)
-- Name: role role_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role
    ADD CONSTRAINT role_pkey PRIMARY KEY (id);


--
-- TOC entry 3763 (class 2606 OID 20938)
-- Name: secrets secrets_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.secrets
    ADD CONSTRAINT secrets_pkey PRIMARY KEY (id);


--
-- TOC entry 3795 (class 2606 OID 21125)
-- Name: seed_assignment seed_assignment_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.seed_assignment
    ADD CONSTRAINT seed_assignment_pkey PRIMARY KEY (id);


--
-- TOC entry 3713 (class 2606 OID 20715)
-- Name: server_lock server_lock_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.server_lock
    ADD CONSTRAINT server_lock_pkey PRIMARY KEY (id);


--
-- TOC entry 3639 (class 2606 OID 20490)
-- Name: session session_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.session
    ADD CONSTRAINT session_pkey PRIMARY KEY (key);


--
-- TOC entry 3724 (class 2606 OID 20746)
-- Name: short_url short_url_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.short_url
    ADD CONSTRAINT short_url_pkey PRIMARY KEY (id);


--
-- TOC entry 3591 (class 2606 OID 20289)
-- Name: star star_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.star
    ADD CONSTRAINT star_pkey PRIMARY KEY (id);


--
-- TOC entry 3703 (class 2606 OID 20678)
-- Name: tag tag_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tag
    ADD CONSTRAINT tag_pkey PRIMARY KEY (id);


--
-- TOC entry 3691 (class 2606 OID 20653)
-- Name: team_member team_member_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.team_member
    ADD CONSTRAINT team_member_pkey PRIMARY KEY (id);


--
-- TOC entry 3686 (class 2606 OID 20644)
-- Name: team team_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.team
    ADD CONSTRAINT team_pkey PRIMARY KEY (id);


--
-- TOC entry 3780 (class 2606 OID 20977)
-- Name: team_role team_role_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.team_role
    ADD CONSTRAINT team_role_pkey PRIMARY KEY (id);


--
-- TOC entry 3588 (class 2606 OID 20278)
-- Name: temp_user temp_user_pkey1; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.temp_user
    ADD CONSTRAINT temp_user_pkey1 PRIMARY KEY (id);


--
-- TOC entry 3678 (class 2606 OID 20626)
-- Name: test_data test_data_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.test_data
    ADD CONSTRAINT test_data_pkey PRIMARY KEY (id);


--
-- TOC entry 3710 (class 2606 OID 20703)
-- Name: user_auth user_auth_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_auth
    ADD CONSTRAINT user_auth_pkey PRIMARY KEY (id);


--
-- TOC entry 3718 (class 2606 OID 20725)
-- Name: user_auth_token user_auth_token_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_auth_token
    ADD CONSTRAINT user_auth_token_pkey PRIMARY KEY (id);


--
-- TOC entry 3582 (class 2606 OID 20243)
-- Name: user user_pkey1; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."user"
    ADD CONSTRAINT user_pkey1 PRIMARY KEY (id);


--
-- TOC entry 3785 (class 2606 OID 20987)
-- Name: user_role user_role_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_role
    ADD CONSTRAINT user_role_pkey PRIMARY KEY (id);


--
-- TOC entry 3649 (class 1259 OID 20536)
-- Name: IDX_alert_dashboard_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_alert_dashboard_id" ON public.alert USING btree (dashboard_id);


--
-- TOC entry 3725 (class 1259 OID 20804)
-- Name: IDX_alert_instance_rule_org_id_current_state; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_alert_instance_rule_org_id_current_state" ON public.alert_instance USING btree (rule_org_id, current_state);


--
-- TOC entry 3726 (class 1259 OID 20803)
-- Name: IDX_alert_instance_rule_org_id_rule_uid_current_state; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_alert_instance_rule_org_id_rule_uid_current_state" ON public.alert_instance USING btree (rule_org_id, rule_uid, current_state);


--
-- TOC entry 3661 (class 1259 OID 20581)
-- Name: IDX_alert_notification_state_alert_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_alert_notification_state_alert_id" ON public.alert_notification_state USING btree (alert_id);


--
-- TOC entry 3650 (class 1259 OID 20534)
-- Name: IDX_alert_org_id_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_alert_org_id_id" ON public.alert USING btree (org_id, id);


--
-- TOC entry 3729 (class 1259 OID 20824)
-- Name: IDX_alert_rule_org_id_dashboard_uid_panel_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_alert_rule_org_id_dashboard_uid_panel_id" ON public.alert_rule USING btree (org_id, dashboard_uid, panel_id);


--
-- TOC entry 3730 (class 1259 OID 20821)
-- Name: IDX_alert_rule_org_id_namespace_uid_rule_group; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_alert_rule_org_id_namespace_uid_rule_group" ON public.alert_rule USING btree (org_id, namespace_uid, rule_group);


--
-- TOC entry 3654 (class 1259 OID 20582)
-- Name: IDX_alert_rule_tag_alert_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_alert_rule_tag_alert_id" ON public.alert_rule_tag USING btree (alert_id);


--
-- TOC entry 3735 (class 1259 OID 20841)
-- Name: IDX_alert_rule_version_rule_org_id_rule_namespace_uid_rule_grou; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_alert_rule_version_rule_org_id_rule_namespace_uid_rule_grou" ON public.alert_rule_version USING btree (rule_org_id, rule_namespace_uid, rule_group);


--
-- TOC entry 3651 (class 1259 OID 20562)
-- Name: IDX_alert_state; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_alert_state" ON public.alert USING btree (state);


--
-- TOC entry 3665 (class 1259 OID 20619)
-- Name: IDX_annotation_alert_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_annotation_alert_id" ON public.annotation USING btree (alert_id);


--
-- TOC entry 3666 (class 1259 OID 20592)
-- Name: IDX_annotation_org_id_alert_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_annotation_org_id_alert_id" ON public.annotation USING btree (org_id, alert_id);


--
-- TOC entry 3667 (class 1259 OID 20613)
-- Name: IDX_annotation_org_id_created; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_annotation_org_id_created" ON public.annotation USING btree (org_id, created);


--
-- TOC entry 3668 (class 1259 OID 20617)
-- Name: IDX_annotation_org_id_dashboard_id_epoch_end_epoch; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_annotation_org_id_dashboard_id_epoch_end_epoch" ON public.annotation USING btree (org_id, dashboard_id, epoch_end, epoch);


--
-- TOC entry 3669 (class 1259 OID 20618)
-- Name: IDX_annotation_org_id_epoch_end_epoch; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_annotation_org_id_epoch_end_epoch" ON public.annotation USING btree (org_id, epoch_end, epoch);


--
-- TOC entry 3670 (class 1259 OID 20597)
-- Name: IDX_annotation_org_id_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_annotation_org_id_type" ON public.annotation USING btree (org_id, type);


--
-- TOC entry 3671 (class 1259 OID 20614)
-- Name: IDX_annotation_org_id_updated; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_annotation_org_id_updated" ON public.annotation USING btree (org_id, updated);


--
-- TOC entry 3622 (class 1259 OID 20434)
-- Name: IDX_api_key_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_api_key_org_id" ON public.api_key USING btree (org_id);


--
-- TOC entry 3786 (class 1259 OID 21001)
-- Name: IDX_builtin_role_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_builtin_role_org_id" ON public.builtin_role USING btree (org_id);


--
-- TOC entry 3787 (class 1259 OID 20999)
-- Name: IDX_builtin_role_role; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_builtin_role_role" ON public.builtin_role USING btree (role);


--
-- TOC entry 3788 (class 1259 OID 20998)
-- Name: IDX_builtin_role_role_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_builtin_role_role_id" ON public.builtin_role USING btree (role_id);


--
-- TOC entry 3802 (class 1259 OID 21061)
-- Name: IDX_correlation_source_uid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_correlation_source_uid" ON public.correlation USING btree (source_uid);


--
-- TOC entry 3803 (class 1259 OID 21060)
-- Name: IDX_correlation_uid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_correlation_uid" ON public.correlation USING btree (uid);


--
-- TOC entry 3692 (class 1259 OID 20665)
-- Name: IDX_dashboard_acl_dashboard_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_dashboard_acl_dashboard_id" ON public.dashboard_acl USING btree (dashboard_id);


--
-- TOC entry 3693 (class 1259 OID 20670)
-- Name: IDX_dashboard_acl_org_id_role; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_dashboard_acl_org_id_role" ON public.dashboard_acl USING btree (org_id, role);


--
-- TOC entry 3694 (class 1259 OID 20671)
-- Name: IDX_dashboard_acl_permission; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_dashboard_acl_permission" ON public.dashboard_acl USING btree (permission);


--
-- TOC entry 3695 (class 1259 OID 20669)
-- Name: IDX_dashboard_acl_team_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_dashboard_acl_team_id" ON public.dashboard_acl USING btree (team_id);


--
-- TOC entry 3696 (class 1259 OID 20668)
-- Name: IDX_dashboard_acl_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_dashboard_acl_user_id" ON public.dashboard_acl USING btree (user_id);


--
-- TOC entry 3603 (class 1259 OID 20342)
-- Name: IDX_dashboard_gnet_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_dashboard_gnet_id" ON public.dashboard USING btree (gnet_id);


--
-- TOC entry 3604 (class 1259 OID 20384)
-- Name: IDX_dashboard_is_folder; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_dashboard_is_folder" ON public.dashboard USING btree (is_folder);


--
-- TOC entry 3605 (class 1259 OID 20340)
-- Name: IDX_dashboard_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_dashboard_org_id" ON public.dashboard USING btree (org_id);


--
-- TOC entry 3606 (class 1259 OID 20346)
-- Name: IDX_dashboard_org_id_plugin_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_dashboard_org_id_plugin_id" ON public.dashboard USING btree (org_id, plugin_id);


--
-- TOC entry 3612 (class 1259 OID 20381)
-- Name: IDX_dashboard_provisioning_dashboard_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_dashboard_provisioning_dashboard_id" ON public.dashboard_provisioning USING btree (dashboard_id);


--
-- TOC entry 3613 (class 1259 OID 20382)
-- Name: IDX_dashboard_provisioning_dashboard_id_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_dashboard_provisioning_dashboard_id_name" ON public.dashboard_provisioning USING btree (dashboard_id, name);


--
-- TOC entry 3808 (class 1259 OID 21098)
-- Name: IDX_dashboard_public_config_org_id_dashboard_uid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_dashboard_public_config_org_id_dashboard_uid" ON public.dashboard_public USING btree (org_id, dashboard_uid);


--
-- TOC entry 3627 (class 1259 OID 20461)
-- Name: IDX_dashboard_snapshot_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_dashboard_snapshot_user_id" ON public.dashboard_snapshot USING btree (user_id);


--
-- TOC entry 3600 (class 1259 OID 20344)
-- Name: IDX_dashboard_tag_dashboard_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_dashboard_tag_dashboard_id" ON public.dashboard_tag USING btree (dashboard_id);


--
-- TOC entry 3607 (class 1259 OID 20383)
-- Name: IDX_dashboard_title; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_dashboard_title" ON public.dashboard USING btree (title);


--
-- TOC entry 3679 (class 1259 OID 20636)
-- Name: IDX_dashboard_version_dashboard_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_dashboard_version_dashboard_id" ON public.dashboard_version USING btree (dashboard_id);


--
-- TOC entry 3616 (class 1259 OID 20406)
-- Name: IDX_data_source_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_data_source_org_id" ON public.data_source USING btree (org_id);


--
-- TOC entry 3617 (class 1259 OID 20412)
-- Name: IDX_data_source_org_id_is_default; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_data_source_org_id_is_default" ON public.data_source USING btree (org_id, is_default);


--
-- TOC entry 3813 (class 1259 OID 21109)
-- Name: IDX_file_parent_folder_path_hash; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_file_parent_folder_path_hash" ON public.file USING btree (parent_folder_path_hash);


--
-- TOC entry 3816 (class 1259 OID 21143)
-- Name: IDX_folder_parent_uid_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_folder_parent_uid_org_id" ON public.folder USING btree (parent_uid, org_id);


--
-- TOC entry 3704 (class 1259 OID 20696)
-- Name: IDX_login_attempt_username; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_login_attempt_username" ON public.login_attempt USING btree (username);


--
-- TOC entry 3595 (class 1259 OID 20308)
-- Name: IDX_org_user_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_org_user_org_id" ON public.org_user USING btree (org_id);


--
-- TOC entry 3596 (class 1259 OID 20310)
-- Name: IDX_org_user_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_org_user_user_id" ON public.org_user USING btree (user_id);


--
-- TOC entry 3767 (class 1259 OID 20957)
-- Name: IDX_permission_role_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_permission_role_id" ON public.permission USING btree (role_id);


--
-- TOC entry 3645 (class 1259 OID 20523)
-- Name: IDX_preferences_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_preferences_org_id" ON public.preferences USING btree (org_id);


--
-- TOC entry 3646 (class 1259 OID 20524)
-- Name: IDX_preferences_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_preferences_user_id" ON public.preferences USING btree (user_id);


--
-- TOC entry 3796 (class 1259 OID 21024)
-- Name: IDX_query_history_org_id_created_by_datasource_uid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_query_history_org_id_created_by_datasource_uid" ON public.query_history USING btree (org_id, created_by, datasource_uid);


--
-- TOC entry 3771 (class 1259 OID 20968)
-- Name: IDX_role_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_role_org_id" ON public.role USING btree (org_id);


--
-- TOC entry 3687 (class 1259 OID 20654)
-- Name: IDX_team_member_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_team_member_org_id" ON public.team_member USING btree (org_id);


--
-- TOC entry 3688 (class 1259 OID 20656)
-- Name: IDX_team_member_team_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_team_member_team_id" ON public.team_member USING btree (team_id);


--
-- TOC entry 3683 (class 1259 OID 20645)
-- Name: IDX_team_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_team_org_id" ON public.team USING btree (org_id);


--
-- TOC entry 3776 (class 1259 OID 20978)
-- Name: IDX_team_role_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_team_role_org_id" ON public.team_role USING btree (org_id);


--
-- TOC entry 3777 (class 1259 OID 20980)
-- Name: IDX_team_role_team_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_team_role_team_id" ON public.team_role USING btree (team_id);


--
-- TOC entry 3583 (class 1259 OID 20281)
-- Name: IDX_temp_user_code; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_temp_user_code" ON public.temp_user USING btree (code);


--
-- TOC entry 3584 (class 1259 OID 20279)
-- Name: IDX_temp_user_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_temp_user_email" ON public.temp_user USING btree (email);


--
-- TOC entry 3585 (class 1259 OID 20280)
-- Name: IDX_temp_user_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_temp_user_org_id" ON public.temp_user USING btree (org_id);


--
-- TOC entry 3586 (class 1259 OID 20282)
-- Name: IDX_temp_user_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_temp_user_status" ON public.temp_user USING btree (status);


--
-- TOC entry 3707 (class 1259 OID 20705)
-- Name: IDX_user_auth_auth_module_auth_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_user_auth_auth_module_auth_id" ON public.user_auth USING btree (auth_module, auth_id);


--
-- TOC entry 3714 (class 1259 OID 20728)
-- Name: IDX_user_auth_token_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_user_auth_token_user_id" ON public.user_auth_token USING btree (user_id);


--
-- TOC entry 3708 (class 1259 OID 20708)
-- Name: IDX_user_auth_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_user_auth_user_id" ON public.user_auth USING btree (user_id);


--
-- TOC entry 3578 (class 1259 OID 20250)
-- Name: IDX_user_login_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_user_login_email" ON public."user" USING btree (login, email);


--
-- TOC entry 3781 (class 1259 OID 20988)
-- Name: IDX_user_role_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_user_role_org_id" ON public.user_role USING btree (org_id);


--
-- TOC entry 3782 (class 1259 OID 20990)
-- Name: IDX_user_role_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IDX_user_role_user_id" ON public.user_role USING btree (user_id);


--
-- TOC entry 3739 (class 1259 OID 20902)
-- Name: UQE_alert_configuration_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_alert_configuration_org_id" ON public.alert_configuration USING btree (org_id);


--
-- TOC entry 3748 (class 1259 OID 20889)
-- Name: UQE_alert_image_token; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_alert_image_token" ON public.alert_image USING btree (token);


--
-- TOC entry 3658 (class 1259 OID 20580)
-- Name: UQE_alert_notification_org_id_uid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_alert_notification_org_id_uid" ON public.alert_notification USING btree (org_id, uid);


--
-- TOC entry 3662 (class 1259 OID 20579)
-- Name: UQE_alert_notification_state_org_id_alert_id_notifier_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_alert_notification_state_org_id_alert_id_notifier_id" ON public.alert_notification_state USING btree (org_id, alert_id, notifier_id);


--
-- TOC entry 3731 (class 1259 OID 20823)
-- Name: UQE_alert_rule_org_id_namespace_uid_title; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_alert_rule_org_id_namespace_uid_title" ON public.alert_rule USING btree (org_id, namespace_uid, title);


--
-- TOC entry 3732 (class 1259 OID 20820)
-- Name: UQE_alert_rule_org_id_uid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_alert_rule_org_id_uid" ON public.alert_rule USING btree (org_id, uid);


--
-- TOC entry 3655 (class 1259 OID 20548)
-- Name: UQE_alert_rule_tag_alert_id_tag_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_alert_rule_tag_alert_id_tag_id" ON public.alert_rule_tag USING btree (alert_id, tag_id);


--
-- TOC entry 3736 (class 1259 OID 20840)
-- Name: UQE_alert_rule_version_rule_org_id_rule_uid_version; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_alert_rule_version_rule_org_id_rule_uid_version" ON public.alert_rule_version USING btree (rule_org_id, rule_uid, version);


--
-- TOC entry 3674 (class 1259 OID 20610)
-- Name: UQE_annotation_tag_annotation_id_tag_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_annotation_tag_annotation_id_tag_id" ON public.annotation_tag USING btree (annotation_id, tag_id);


--
-- TOC entry 3623 (class 1259 OID 20438)
-- Name: UQE_api_key_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_api_key_key" ON public.api_key USING btree (key);


--
-- TOC entry 3624 (class 1259 OID 20437)
-- Name: UQE_api_key_org_id_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_api_key_org_id_name" ON public.api_key USING btree (org_id, name);


--
-- TOC entry 3789 (class 1259 OID 21002)
-- Name: UQE_builtin_role_org_id_role_id_role; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_builtin_role_org_id_role_id_role" ON public.builtin_role USING btree (org_id, role_id, role);


--
-- TOC entry 3719 (class 1259 OID 20736)
-- Name: UQE_cache_data_cache_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_cache_data_cache_key" ON public.cache_data USING btree (cache_key);


--
-- TOC entry 3697 (class 1259 OID 20667)
-- Name: UQE_dashboard_acl_dashboard_id_team_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_dashboard_acl_dashboard_id_team_id" ON public.dashboard_acl USING btree (dashboard_id, team_id);


--
-- TOC entry 3698 (class 1259 OID 20666)
-- Name: UQE_dashboard_acl_dashboard_id_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_dashboard_acl_dashboard_id_user_id" ON public.dashboard_acl USING btree (dashboard_id, user_id);


--
-- TOC entry 3608 (class 1259 OID 20361)
-- Name: UQE_dashboard_org_id_folder_id_title; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_dashboard_org_id_folder_id_title" ON public.dashboard USING btree (org_id, folder_id, title);


--
-- TOC entry 3609 (class 1259 OID 20350)
-- Name: UQE_dashboard_org_id_uid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_dashboard_org_id_uid" ON public.dashboard USING btree (org_id, uid);


--
-- TOC entry 3809 (class 1259 OID 21099)
-- Name: UQE_dashboard_public_config_access_token; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_dashboard_public_config_access_token" ON public.dashboard_public USING btree (access_token);


--
-- TOC entry 3810 (class 1259 OID 21097)
-- Name: UQE_dashboard_public_config_uid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_dashboard_public_config_uid" ON public.dashboard_public USING btree (uid);


--
-- TOC entry 3628 (class 1259 OID 20463)
-- Name: UQE_dashboard_snapshot_delete_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_dashboard_snapshot_delete_key" ON public.dashboard_snapshot USING btree (delete_key);


--
-- TOC entry 3629 (class 1259 OID 20462)
-- Name: UQE_dashboard_snapshot_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_dashboard_snapshot_key" ON public.dashboard_snapshot USING btree (key);


--
-- TOC entry 3680 (class 1259 OID 20637)
-- Name: UQE_dashboard_version_dashboard_id_version; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_dashboard_version_dashboard_id_version" ON public.dashboard_version USING btree (dashboard_id, version);


--
-- TOC entry 3618 (class 1259 OID 20409)
-- Name: UQE_data_source_org_id_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_data_source_org_id_name" ON public.data_source USING btree (org_id, name);


--
-- TOC entry 3619 (class 1259 OID 20411)
-- Name: UQE_data_source_org_id_uid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_data_source_org_id_uid" ON public.data_source USING btree (org_id, uid);


--
-- TOC entry 3815 (class 1259 OID 21115)
-- Name: UQE_file_meta_path_hash_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_file_meta_path_hash_key" ON public.file_meta USING btree (path_hash, key);


--
-- TOC entry 3814 (class 1259 OID 21108)
-- Name: UQE_file_path_hash; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_file_path_hash" ON public.file USING btree (path_hash);


--
-- TOC entry 3817 (class 1259 OID 21153)
-- Name: UQE_folder_title_parent_uid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_folder_title_parent_uid" ON public.folder USING btree (title, parent_uid);


--
-- TOC entry 3818 (class 1259 OID 21144)
-- Name: UQE_folder_uid_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_folder_uid_org_id" ON public.folder USING btree (uid, org_id);


--
-- TOC entry 3764 (class 1259 OID 20949)
-- Name: UQE_kv_store_org_id_namespace_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_kv_store_org_id_namespace_key" ON public.kv_store USING btree (org_id, namespace, key);


--
-- TOC entry 3757 (class 1259 OID 20921)
-- Name: UQE_library_element_connection_element_id_kind_connection_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_library_element_connection_element_id_kind_connection_id" ON public.library_element_connection USING btree (element_id, kind, connection_id);


--
-- TOC entry 3753 (class 1259 OID 20913)
-- Name: UQE_library_element_org_id_folder_id_name_kind; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_library_element_org_id_folder_id_name_kind" ON public.library_element USING btree (org_id, folder_id, name, kind);


--
-- TOC entry 3754 (class 1259 OID 20922)
-- Name: UQE_library_element_org_id_uid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_library_element_org_id_uid" ON public.library_element USING btree (org_id, uid);


--
-- TOC entry 3742 (class 1259 OID 20868)
-- Name: UQE_ngalert_configuration_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_ngalert_configuration_org_id" ON public.ngalert_configuration USING btree (org_id);


--
-- TOC entry 3592 (class 1259 OID 20311)
-- Name: UQE_org_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_org_name" ON public.org USING btree (name);


--
-- TOC entry 3597 (class 1259 OID 20309)
-- Name: UQE_org_user_org_id_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_org_user_org_id_user_id" ON public.org_user USING btree (org_id, user_id);


--
-- TOC entry 3768 (class 1259 OID 20958)
-- Name: UQE_permission_role_id_action_scope; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_permission_role_id_action_scope" ON public.permission USING btree (role_id, action, scope);


--
-- TOC entry 3640 (class 1259 OID 21117)
-- Name: UQE_playlist_org_id_uid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_playlist_org_id_uid" ON public.playlist USING btree (org_id, uid);


--
-- TOC entry 3635 (class 1259 OID 20483)
-- Name: UQE_plugin_setting_org_id_plugin_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_plugin_setting_org_id_plugin_id" ON public.plugin_setting USING btree (org_id, plugin_id);


--
-- TOC entry 3745 (class 1259 OID 20879)
-- Name: UQE_provenance_type_record_type_record_key_org_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_provenance_type_record_type_record_key_org_id" ON public.provenance_type USING btree (record_type, record_key, org_id);


--
-- TOC entry 3799 (class 1259 OID 21047)
-- Name: UQE_query_history_star_user_id_query_uid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_query_history_star_user_id_query_uid" ON public.query_history_star USING btree (user_id, query_uid);


--
-- TOC entry 3632 (class 1259 OID 20472)
-- Name: UQE_quota_org_id_user_id_target; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_quota_org_id_user_id_target" ON public.quota USING btree (org_id, user_id, target);


--
-- TOC entry 3772 (class 1259 OID 20969)
-- Name: UQE_role_org_id_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_role_org_id_name" ON public.role USING btree (org_id, name);


--
-- TOC entry 3773 (class 1259 OID 21003)
-- Name: UQE_role_uid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_role_uid" ON public.role USING btree (uid);


--
-- TOC entry 3792 (class 1259 OID 21121)
-- Name: UQE_seed_assignment_builtin_role_action_scope; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_seed_assignment_builtin_role_action_scope" ON public.seed_assignment USING btree (builtin_role, action, scope);


--
-- TOC entry 3793 (class 1259 OID 21120)
-- Name: UQE_seed_assignment_builtin_role_role_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_seed_assignment_builtin_role_role_name" ON public.seed_assignment USING btree (builtin_role, role_name);


--
-- TOC entry 3711 (class 1259 OID 20716)
-- Name: UQE_server_lock_operation_uid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_server_lock_operation_uid" ON public.server_lock USING btree (operation_uid);


--
-- TOC entry 3722 (class 1259 OID 20748)
-- Name: UQE_short_url_org_id_uid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_short_url_org_id_uid" ON public.short_url USING btree (org_id, uid);


--
-- TOC entry 3589 (class 1259 OID 20290)
-- Name: UQE_star_user_id_dashboard_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_star_user_id_dashboard_id" ON public.star USING btree (user_id, dashboard_id);


--
-- TOC entry 3701 (class 1259 OID 20679)
-- Name: UQE_tag_key_value; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_tag_key_value" ON public.tag USING btree (key, value);


--
-- TOC entry 3689 (class 1259 OID 20655)
-- Name: UQE_team_member_org_id_team_id_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_team_member_org_id_team_id_user_id" ON public.team_member USING btree (org_id, team_id, user_id);


--
-- TOC entry 3684 (class 1259 OID 20646)
-- Name: UQE_team_org_id_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_team_org_id_name" ON public.team USING btree (org_id, name);


--
-- TOC entry 3778 (class 1259 OID 20979)
-- Name: UQE_team_role_org_id_team_id_role_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_team_role_org_id_team_id_role_id" ON public.team_role USING btree (org_id, team_id, role_id);


--
-- TOC entry 3715 (class 1259 OID 20726)
-- Name: UQE_user_auth_token_auth_token; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_user_auth_token_auth_token" ON public.user_auth_token USING btree (auth_token);


--
-- TOC entry 3716 (class 1259 OID 20727)
-- Name: UQE_user_auth_token_prev_auth_token; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_user_auth_token_prev_auth_token" ON public.user_auth_token USING btree (prev_auth_token);


--
-- TOC entry 3579 (class 1259 OID 20248)
-- Name: UQE_user_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_user_email" ON public."user" USING btree (email);


--
-- TOC entry 3580 (class 1259 OID 20247)
-- Name: UQE_user_login; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_user_login" ON public."user" USING btree (login);


--
-- TOC entry 3783 (class 1259 OID 20989)
-- Name: UQE_user_role_org_id_user_id_role_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "UQE_user_role_org_id_user_id_role_id" ON public.user_role USING btree (org_id, user_id, role_id);


-- Completed on 2024-05-15 16:02:53

--
-- PostgreSQL database dump complete
--

