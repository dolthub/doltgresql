-- Downloaded from: https://github.com/nyfagel/klubb/blob/d96cc4192acde0453c175976e35869e5d52da555/database/klubb_postgre.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 9.1.8
-- Dumped by pg_dump version 9.2.2
-- Started on 2013-03-01 14:44:58 CET

SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

DROP DATABASE klubb;
--
-- TOC entry 2061 (class 1262 OID 16581)
-- Name: klubb; Type: DATABASE; Schema: -; Owner: klubb
--

CREATE DATABASE klubb WITH TEMPLATE = template0 ENCODING = 'UTF8' LC_COLLATE = 'sv_SE.UTF-8' LC_CTYPE = 'sv_SE.UTF-8';


ALTER DATABASE klubb OWNER TO klubb;

\connect klubb

SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- TOC entry 2062 (class 1262 OID 16581)
-- Dependencies: 2061
-- Name: klubb; Type: COMMENT; Schema: -; Owner: klubb
--

COMMENT ON DATABASE klubb IS 'Database for the local installation of Klubb.';


--
-- TOC entry 6 (class 2615 OID 2200)
-- Name: klubb; Type: SCHEMA; Schema: -; Owner: klubb
--

CREATE SCHEMA klubb;


ALTER SCHEMA klubb OWNER TO klubb;

--
-- TOC entry 2063 (class 0 OID 0)
-- Dependencies: 6
-- Name: SCHEMA klubb; Type: COMMENT; Schema: -; Owner: klubb
--

COMMENT ON SCHEMA klubb IS 'Default schema for Klubb.';


--
-- TOC entry 190 (class 3079 OID 11656)
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- TOC entry 2065 (class 0 OID 0)
-- Dependencies: 190
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = klubb, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- TOC entry 161 (class 1259 OID 16582)
-- Name: authentication; Type: TABLE; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE TABLE authentication (
    "user" integer NOT NULL,
    series character varying(255) NOT NULL,
    key character varying(255) NOT NULL,
    created integer NOT NULL
);


ALTER TABLE klubb.authentication OWNER TO klubb;

--
-- TOC entry 164 (class 1259 OID 16617)
-- Name: ci_sessions; Type: TABLE; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE TABLE ci_sessions (
    session_id character varying(40) DEFAULT '0'::character varying NOT NULL,
    ip_address character varying(45) DEFAULT '0'::character varying NOT NULL,
    user_agent character varying(120) NOT NULL,
    last_activity integer DEFAULT 0 NOT NULL,
    user_data text NOT NULL
);


ALTER TABLE klubb.ci_sessions OWNER TO klubb;

--
-- TOC entry 182 (class 1259 OID 16774)
-- Name: keys; Type: TABLE; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE TABLE keys (
    id integer NOT NULL,
    key character varying(40) NOT NULL,
    level smallint NOT NULL,
    ignore_limits boolean DEFAULT false NOT NULL,
    is_private_key boolean DEFAULT false NOT NULL,
    ip_addresses text,
    date_created integer NOT NULL
);


ALTER TABLE klubb.keys OWNER TO klubb;

--
-- TOC entry 181 (class 1259 OID 16772)
-- Name: keys_id_seq; Type: SEQUENCE; Schema: klubb; Owner: klubb
--

CREATE SEQUENCE keys_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE klubb.keys_id_seq OWNER TO klubb;

--
-- TOC entry 2066 (class 0 OID 0)
-- Dependencies: 181
-- Name: keys_id_seq; Type: SEQUENCE OWNED BY; Schema: klubb; Owner: klubb
--

ALTER SEQUENCE keys_id_seq OWNED BY keys.id;


--
-- TOC entry 186 (class 1259 OID 16798)
-- Name: limits; Type: TABLE; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE TABLE limits (
    id integer NOT NULL,
    uri character varying(255) NOT NULL,
    count integer NOT NULL,
    hour_started integer NOT NULL,
    api_key character varying(40) NOT NULL
);


ALTER TABLE klubb.limits OWNER TO klubb;

--
-- TOC entry 185 (class 1259 OID 16796)
-- Name: limits_id_seq; Type: SEQUENCE; Schema: klubb; Owner: klubb
--

CREATE SEQUENCE limits_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE klubb.limits_id_seq OWNER TO klubb;

--
-- TOC entry 2067 (class 0 OID 0)
-- Dependencies: 185
-- Name: limits_id_seq; Type: SEQUENCE OWNED BY; Schema: klubb; Owner: klubb
--

ALTER SEQUENCE limits_id_seq OWNED BY limits.id;


--
-- TOC entry 166 (class 1259 OID 16630)
-- Name: log; Type: TABLE; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE TABLE log (
    id integer NOT NULL,
    "user" integer,
    action character varying(45) NOT NULL,
    path character varying(100),
    "time" timestamp with time zone NOT NULL
);


ALTER TABLE klubb.log OWNER TO klubb;

--
-- TOC entry 165 (class 1259 OID 16628)
-- Name: log_id_seq; Type: SEQUENCE; Schema: klubb; Owner: klubb
--

CREATE SEQUENCE log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE klubb.log_id_seq OWNER TO klubb;

--
-- TOC entry 2068 (class 0 OID 0)
-- Dependencies: 165
-- Name: log_id_seq; Type: SEQUENCE OWNED BY; Schema: klubb; Owner: klubb
--

ALTER SEQUENCE log_id_seq OWNED BY log.id;


--
-- TOC entry 184 (class 1259 OID 16787)
-- Name: logs; Type: TABLE; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE TABLE logs (
    id integer NOT NULL,
    uri character varying(255) NOT NULL,
    method character varying(6) NOT NULL,
    params text,
    api_key character varying(40) NOT NULL,
    ip_address character varying(45) NOT NULL,
    "time" integer NOT NULL,
    authorized boolean NOT NULL
);


ALTER TABLE klubb.logs OWNER TO klubb;

--
-- TOC entry 183 (class 1259 OID 16785)
-- Name: logs_id_seq; Type: SEQUENCE; Schema: klubb; Owner: klubb
--

CREATE SEQUENCE logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE klubb.logs_id_seq OWNER TO klubb;

--
-- TOC entry 2069 (class 0 OID 0)
-- Dependencies: 183
-- Name: logs_id_seq; Type: SEQUENCE OWNED BY; Schema: klubb; Owner: klubb
--

ALTER SEQUENCE logs_id_seq OWNED BY logs.id;


--
-- TOC entry 168 (class 1259 OID 16653)
-- Name: member_data; Type: TABLE; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE TABLE member_data (
    id integer NOT NULL,
    notes text,
    inactive boolean DEFAULT false,
    inactive_date date
);


ALTER TABLE klubb.member_data OWNER TO klubb;

--
-- TOC entry 167 (class 1259 OID 16651)
-- Name: member_data_id_seq; Type: SEQUENCE; Schema: klubb; Owner: klubb
--

CREATE SEQUENCE member_data_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE klubb.member_data_id_seq OWNER TO klubb;

--
-- TOC entry 2070 (class 0 OID 0)
-- Dependencies: 167
-- Name: member_data_id_seq; Type: SEQUENCE OWNED BY; Schema: klubb; Owner: klubb
--

ALTER SEQUENCE member_data_id_seq OWNED BY member_data.id;


--
-- TOC entry 188 (class 1259 OID 25557)
-- Name: member_flags; Type: TABLE; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE TABLE member_flags (
    key character varying(45) NOT NULL,
    "desc" character varying(100) NOT NULL,
    id integer NOT NULL
);


ALTER TABLE klubb.member_flags OWNER TO klubb;

--
-- TOC entry 187 (class 1259 OID 25555)
-- Name: member_flags_id_seq; Type: SEQUENCE; Schema: klubb; Owner: klubb
--

CREATE SEQUENCE member_flags_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE klubb.member_flags_id_seq OWNER TO klubb;

--
-- TOC entry 2071 (class 0 OID 0)
-- Dependencies: 187
-- Name: member_flags_id_seq; Type: SEQUENCE OWNED BY; Schema: klubb; Owner: klubb
--

ALTER SEQUENCE member_flags_id_seq OWNED BY member_flags.id;


--
-- TOC entry 170 (class 1259 OID 16665)
-- Name: members; Type: TABLE; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE TABLE members (
    id integer NOT NULL,
    type integer,
    firstname character varying(45) NOT NULL,
    lastname character varying(45) NOT NULL,
    ssid character varying(12) NOT NULL,
    phone character varying(12) NOT NULL,
    address character varying(45),
    zip character varying(10),
    city character varying(45),
    data integer,
    email character varying(100),
    last_update timestamp with time zone
);


ALTER TABLE klubb.members OWNER TO klubb;

--
-- TOC entry 169 (class 1259 OID 16663)
-- Name: members_id_seq; Type: SEQUENCE; Schema: klubb; Owner: klubb
--

CREATE SEQUENCE members_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE klubb.members_id_seq OWNER TO klubb;

--
-- TOC entry 2072 (class 0 OID 0)
-- Dependencies: 169
-- Name: members_id_seq; Type: SEQUENCE OWNED BY; Schema: klubb; Owner: klubb
--

ALTER SEQUENCE members_id_seq OWNED BY members.id;


--
-- TOC entry 180 (class 1259 OID 16763)
-- Name: migrations; Type: TABLE; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE TABLE migrations (
    version integer NOT NULL
);


ALTER TABLE klubb.migrations OWNER TO klubb;

--
-- TOC entry 172 (class 1259 OID 16678)
-- Name: rights; Type: TABLE; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE TABLE rights (
    id integer NOT NULL,
    role integer NOT NULL,
    add_members boolean DEFAULT false NOT NULL,
    add_users boolean DEFAULT false NOT NULL,
    use_system boolean DEFAULT true NOT NULL
);


ALTER TABLE klubb.rights OWNER TO klubb;

--
-- TOC entry 171 (class 1259 OID 16676)
-- Name: rights_id_seq; Type: SEQUENCE; Schema: klubb; Owner: klubb
--

CREATE SEQUENCE rights_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE klubb.rights_id_seq OWNER TO klubb;

--
-- TOC entry 2073 (class 0 OID 0)
-- Dependencies: 171
-- Name: rights_id_seq; Type: SEQUENCE OWNED BY; Schema: klubb; Owner: klubb
--

ALTER SEQUENCE rights_id_seq OWNED BY rights.id;


--
-- TOC entry 174 (class 1259 OID 16695)
-- Name: roles; Type: TABLE; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE TABLE roles (
    id integer NOT NULL,
    name character varying(45) NOT NULL,
    system boolean DEFAULT true NOT NULL
);


ALTER TABLE klubb.roles OWNER TO klubb;

--
-- TOC entry 179 (class 1259 OID 16754)
-- Name: role_view; Type: VIEW; Schema: klubb; Owner: klubb
--

CREATE VIEW role_view AS
    SELECT roles.id, rights.add_members, rights.add_users, rights.use_system, roles.name, roles.system AS system_role FROM roles, rights WHERE (roles.id = rights.role);


ALTER TABLE klubb.role_view OWNER TO klubb;

--
-- TOC entry 173 (class 1259 OID 16693)
-- Name: roles_id_seq; Type: SEQUENCE; Schema: klubb; Owner: klubb
--

CREATE SEQUENCE roles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE klubb.roles_id_seq OWNER TO klubb;

--
-- TOC entry 2074 (class 0 OID 0)
-- Dependencies: 173
-- Name: roles_id_seq; Type: SEQUENCE OWNED BY; Schema: klubb; Owner: klubb
--

ALTER SEQUENCE roles_id_seq OWNED BY roles.id;


--
-- TOC entry 175 (class 1259 OID 16704)
-- Name: system; Type: TABLE; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE TABLE system (
    key character varying(45) NOT NULL,
    value character varying(255) NOT NULL
);


ALTER TABLE klubb.system OWNER TO klubb;

--
-- TOC entry 177 (class 1259 OID 16711)
-- Name: types; Type: TABLE; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE TABLE types (
    id integer NOT NULL,
    name character varying(45) NOT NULL,
    plural character varying(45),
    "desc" character varying(65)
);


ALTER TABLE klubb.types OWNER TO klubb;

--
-- TOC entry 176 (class 1259 OID 16709)
-- Name: types_id_seq; Type: SEQUENCE; Schema: klubb; Owner: klubb
--

CREATE SEQUENCE types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE klubb.types_id_seq OWNER TO klubb;

--
-- TOC entry 2075 (class 0 OID 0)
-- Dependencies: 176
-- Name: types_id_seq; Type: SEQUENCE OWNED BY; Schema: klubb; Owner: klubb
--

ALTER SEQUENCE types_id_seq OWNED BY types.id;


--
-- TOC entry 189 (class 1259 OID 25565)
-- Name: types_requirements; Type: TABLE; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE TABLE types_requirements (
    fieldname character varying(45) NOT NULL,
    type integer NOT NULL,
    fieldtype character varying(45) NOT NULL,
    rule character varying(45) NOT NULL,
    rule_desc character varying(45) NOT NULL,
    sort_order integer
);


ALTER TABLE klubb.types_requirements OWNER TO klubb;

--
-- TOC entry 178 (class 1259 OID 16739)
-- Name: user_role; Type: TABLE; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE TABLE user_role (
    "user" integer NOT NULL,
    role integer NOT NULL
);


ALTER TABLE klubb.user_role OWNER TO klubb;

--
-- TOC entry 163 (class 1259 OID 16594)
-- Name: users; Type: TABLE; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE TABLE users (
    id integer NOT NULL,
    username character varying(255) NOT NULL,
    firstname character varying(100),
    lastname character varying(100),
    email character varying(255) NOT NULL,
    phone character varying(12),
    key character varying(255) NOT NULL,
    password character varying(255) NOT NULL,
    registered integer NOT NULL,
    first_login boolean DEFAULT true NOT NULL,
    loggedin boolean DEFAULT false NOT NULL
);


ALTER TABLE klubb.users OWNER TO klubb;

--
-- TOC entry 162 (class 1259 OID 16592)
-- Name: users_id_seq; Type: SEQUENCE; Schema: klubb; Owner: klubb
--

CREATE SEQUENCE users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE klubb.users_id_seq OWNER TO klubb;

--
-- TOC entry 2076 (class 0 OID 0)
-- Dependencies: 162
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: klubb; Owner: klubb
--

ALTER SEQUENCE users_id_seq OWNED BY users.id;


--
-- TOC entry 1975 (class 2604 OID 16777)
-- Name: id; Type: DEFAULT; Schema: klubb; Owner: klubb
--

ALTER TABLE ONLY keys ALTER COLUMN id SET DEFAULT nextval('keys_id_seq'::regclass);


--
-- TOC entry 1979 (class 2604 OID 16801)
-- Name: id; Type: DEFAULT; Schema: klubb; Owner: klubb
--

ALTER TABLE ONLY limits ALTER COLUMN id SET DEFAULT nextval('limits_id_seq'::regclass);


--
-- TOC entry 1964 (class 2604 OID 16633)
-- Name: id; Type: DEFAULT; Schema: klubb; Owner: klubb
--

ALTER TABLE ONLY log ALTER COLUMN id SET DEFAULT nextval('log_id_seq'::regclass);


--
-- TOC entry 1978 (class 2604 OID 16790)
-- Name: id; Type: DEFAULT; Schema: klubb; Owner: klubb
--

ALTER TABLE ONLY logs ALTER COLUMN id SET DEFAULT nextval('logs_id_seq'::regclass);


--
-- TOC entry 1965 (class 2604 OID 16656)
-- Name: id; Type: DEFAULT; Schema: klubb; Owner: klubb
--

ALTER TABLE ONLY member_data ALTER COLUMN id SET DEFAULT nextval('member_data_id_seq'::regclass);


--
-- TOC entry 1980 (class 2604 OID 25560)
-- Name: id; Type: DEFAULT; Schema: klubb; Owner: klubb
--

ALTER TABLE ONLY member_flags ALTER COLUMN id SET DEFAULT nextval('member_flags_id_seq'::regclass);


--
-- TOC entry 1967 (class 2604 OID 16668)
-- Name: id; Type: DEFAULT; Schema: klubb; Owner: klubb
--

ALTER TABLE ONLY members ALTER COLUMN id SET DEFAULT nextval('members_id_seq'::regclass);


--
-- TOC entry 1968 (class 2604 OID 16681)
-- Name: id; Type: DEFAULT; Schema: klubb; Owner: klubb
--

ALTER TABLE ONLY rights ALTER COLUMN id SET DEFAULT nextval('rights_id_seq'::regclass);


--
-- TOC entry 1972 (class 2604 OID 16698)
-- Name: id; Type: DEFAULT; Schema: klubb; Owner: klubb
--

ALTER TABLE ONLY roles ALTER COLUMN id SET DEFAULT nextval('roles_id_seq'::regclass);


--
-- TOC entry 1974 (class 2604 OID 16714)
-- Name: id; Type: DEFAULT; Schema: klubb; Owner: klubb
--

ALTER TABLE ONLY types ALTER COLUMN id SET DEFAULT nextval('types_id_seq'::regclass);


--
-- TOC entry 1958 (class 2604 OID 16597)
-- Name: id; Type: DEFAULT; Schema: klubb; Owner: klubb
--

ALTER TABLE ONLY users ALTER COLUMN id SET DEFAULT nextval('users_id_seq'::regclass);


--
-- TOC entry 2029 (class 0 OID 16582)
-- Dependencies: 161
-- Data for Name: authentication; Type: TABLE DATA; Schema: klubb; Owner: klubb
--

INSERT INTO authentication ("user", series, key, created) VALUES (1, 'cf19ba7b4a007bbe6be6698358dfe3e081903321868902124ec80a13b4a31214', 'a48bf47a7dd6db3b42d8f7ab21e6a82a9c3fbf8e42c2c309625a24e0315cde23', 1362135604);


--
-- TOC entry 2032 (class 0 OID 16617)
-- Dependencies: 164
-- Data for Name: ci_sessions; Type: TABLE DATA; Schema: klubb; Owner: klubb
--

INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('cf73b33d33212e7f8eb03a769803e0d2', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143510, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('07062b48c927d20c517014ab0ff69c91', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362143595, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('773d27a7b415614d20cdbcd43bc60ead', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143659, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('0ff7bd2d0b9bef60a588fdb6ecb202cd', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362143741, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('d38a519a5c8a66d5ec0f16388b5caa8d', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143869, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('702f2701f3be6175d092bddf742a85a5', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140269, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5ebab74632485f61f0cd7a2ae1698c3a', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140329, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5f2badc6ade093889e0685c618273916', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140419, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('867973ad0a3522ad92a6a83281cb1835', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140809, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('391743c8ac4fb30afb7c93737897bfd1', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362140894, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('f676a0f3a403bd0fc22ec1a6d6779580', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140479, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('df4ee9b28a6534f5bc6cbef9debfbd36', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140569, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('c4cdf48796fbd4c6688c3fcd73d4a53b', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140959, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('1408ec7f672d6d9b78b6c32b1c8f8d1e', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362141041, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('00f5c700ac700353f7562db6cc2c8236', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141109, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('d0e5ea226177235171a615245d58288d', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362141194, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('4a2b0b6213c9f169a87dde7345dc1d1d', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141259, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('96ff905c32afc7033821a7e88bfe25be', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143929, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('ad4177da1876f95f01bc859e954edd8c', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144019, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('9b1f663b19f90cd5571edecca2a0cf12', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144079, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('f0ccc47551b1e18918738b5c6922e80f', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144169, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('8a47154daf54df02fae6574b4fe8655c', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144229, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('1f3cef9098804c7694c213265ee0ed6f', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144319, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('aac53711d6d8675c27102a9614f089a9', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144380, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('61861718d5d17eaccb882d43f58d2fe2', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144469, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('b0be06d2d727ae469fec6324ec1ce5db', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140629, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5e58abf7ec65eae0a2bdcf8de4a7c986', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140719, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('23328bb3d1b3165e52685df64d983a9c', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140779, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('56c123f310bc9e45303e29dc185f17c4', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144529, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('6c017a2dce3e1ba302c60dca3b4657c3', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144649, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('37c798a70dd356d49ac58577e6a28a81', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144739, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('1024b962addf475dc500a5bb770763ee', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144799, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('00f0c7b9c75f131687c0b67a1eaacf85', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144889, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('d3f9a1fd07183f9ac167be67d7a356c3', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144949, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('7f902bc6b6e6a6203c6f67b76899ca62', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362145039, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('1d737bce1356362c9a426bc101edc151', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362145099, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('ed0e602fb21fa50cb375062b600bec04', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362145189, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('19f35502ce3d00ab2f50813027f58681', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362145249, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('18d84316130b64c6c9a1f2d3fa16d84b', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362145339, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('cbc11ab8d490afd70dc293a14c8314ff', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362145399, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('c58782e8d115ca5ec03b2c232e03eb75', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362145489, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('3ac74d023d5a3ffaf940bd37dd789f62', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143539, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('f8ded3705798d4c975a5e51b5eab7f95', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143599, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('f2c771c83f855361befa563a1816f8d7', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143689, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('999a6cc0095b6a347585e02a93a78c65', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143749, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('b98c012e903624c38f95e4d63a9ee7c5', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362140295, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('fa7781a8001105a1ce7e4d4dc23049e8', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140359, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('8bdcfa39cc4a8ce939865442193dbdda', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140839, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('18ff9fb5ca15ddfadd90ec7a86fe582e', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140899, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('f3d6902ceb42ef25f7f7dced51fe23cf', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140989, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('690fbf31361a20a2e531a908b2bf3370', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141049, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('2ac1bcc5672a6a086441e82d9f8998e9', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141139, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('bdf5302e71f8fb32fceb0b4ec2e60bb9', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141199, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('9242a214709ae826c3242a4af03646ff', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143809, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('00dba7bf55f5c040f341217f6968c6be', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362143894, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('c95ca8b06843260092240ac88dde01eb', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143959, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('978b26db93995de5e24e2082a997ce89', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362144041, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5e33e6dd7ad14288c686049d8e85abda', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144109, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('e1dbeb918de0240bcb64cb379e38431a', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362144194, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5832533d3060ab42fd9e96dd7acdd592', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144260, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5e7f93cf7e5528318fc125f6e7e83325', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362144341, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('9b3d52fc288875ab958d260ab3804457', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144409, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('10c0ee9f7ec2d2699b83785fb3d77e1f', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362144495, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('39a434c79a1f951183fb3f0f6ff0a9b8', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362140441, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('df1287f7bc11b28fb2439ea2c791618c', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140509, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('3bebd81c1df4eecb84f35d18671cf598', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362140594, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('df73fad74d56a70cefc660e297f480e1', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140659, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('7bcab884fb209f156041776fb0e3956f', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362140741, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('91bed8cc83e307c95de670426082fe4c', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144559, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('044e34a79e2071f0a33ffb110b4ce533', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144619, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('34ee9f40a7c35f9defb98529bb2def08', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144679, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('ad5e5de0fcc25fdfcdc8f49a077c576e', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144769, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('cdc1904a06594528d6dfa53c36ea6791', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144829, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('59fdeac0ecb74ccf81199ff06420093b', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144919, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('e393f2662d121fc5615cdc05671014b4', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144979, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('d66e527cc5fac09c94868e8216e83383', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362145069, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('6c151eae0f397a0e723f1256fbe4414b', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362145129, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('320eb050d7d7e50f1c53e288e0220160', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362145219, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('eee1f8e060cab5a32d9da0d594d7baf1', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362145279, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('a469e0ea19dc8e620f1183f0dffb2ff2', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362145369, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('4a7015ea559b25e4f04796acaa6f35cd', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362145429, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('c89816e9299dd41c02e1ec31c584e301', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143569, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('d73d214548643487b95da6d64a5c0591', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143629, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('3a4ad95a1784acecfd123496f663dad0', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143719, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('f9cab735a3a621f9ec516fb1a658d692', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143779, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('c3484acf9abc8df7f59214bc9860b06e', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143839, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('36f16d30cfa2fa5722964163f4527275', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140299, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('bbd106328960f61dc9f8f212ad62ceb5', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140389, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('fdb1d4ab52b8955f73444faebce606a9', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140869, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('95712f27a5738bef1863c1bbca99e4f6', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140929, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('999eafd19758316b1f1fde5b2595646a', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141019, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5fd5b7f74983f0b2bd25dfdd6c5b822d', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141079, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('d2cf8f10767b690a86b33a31a2ec46b8', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141169, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('ab23d5d19edca708faedb5143041b6c4', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141230, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('6e90aa4c8655f2ae992db63b9d0487e0', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143899, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('6f81746b16f6db827d9e7fd2fa4b22ad', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143989, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('8c41ad25ed322668c81a44f9472d012f', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144049, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('ad32c9b9877708f0ffa2295ada9ba583', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144139, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('a31167c5005a37ca0e8e72d47e5f9f30', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144199, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('09dd592a174ff0f22a6303524e2f0c0f', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144289, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('9c9a5029ca7233471572f3239e017ec3', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144349, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('e5a2bb80aee336ebd86064425182b1ef', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144439, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('a5b070bfb8ce5efb932f0a1aa82a9aee', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140449, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('8b9a0e5c57c3b7a451c3fe035cba2fa9', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140539, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('c7bd1056624587f597e0dc7c0a1dd996', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140599, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('523184493c4f0e205814d93cd24aac99', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140689, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('f7a5cbeaf98528dc618e4804935b2a14', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140749, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('20a6126b5f598f9f3cef9667fff9cd4c', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144499, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('e062c411e65b84fb0921b261ed62ad17', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144589, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('ef013dc4a6bc51f46b2d0a7ccb0e19d6', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362144641, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('f0cb431bf77657241c03a1237a12929e', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144709, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('7c13a042d020cc3fe494456027b09673', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362144794, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('fb3a0d55dc5fd34140932568ca8c6d42', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362144859, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5e0a5b611a9609dbef1a351319d4650d', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362144941, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('14bfc06e98f8aa2966fb54f5cb09cb9f', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362145009, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5bbccc1506ee74b4023ce1ab5c9f6362', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362145094, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('ab7acec3668fe72c3570f1ed8ee05b44', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362145159, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('cfc88d2bd969f6a702fafdb8f7bdd1f4', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362145242, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('4761fa8f4f56722b5f69183dd690c02a', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362145309, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('962b183d1daeb1dcf0d4f36e81ded250', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362145394, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('9ba145b8bc88b0ecd746d8b9a0f5976d', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362145459, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('0b13016dcf6f7d301627103dbdd6dd7b', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141289, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('113b1b2ff38a67011b2873ad0b96f257', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141349, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('bcdbbdd50914bf5f55d155fa48d45a79', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141439, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('e211f5afc83be5d8dfd41c890d8a43e2', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141499, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('a8e1a8638c83c224a1f2ec8973860645', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141589, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('2b07bd8a32c76746ae8f39a159f29967', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141649, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('4f692f0067a9ebe467f88f29ae0cc98e', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141739, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('6d93be9b726417543bba4021c7c7d356', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141799, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('0daaa9a294a530bbf07c99747c50e8e9', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362138341, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('c85009ae27bac7e16bd60a4be2a7fd01', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138409, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('393b3499d2d8a4d915858f7dfb3076a2', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362138494, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('4eab3fd3fdba0574250b5e68a484c1e6', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138559, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('3300dbd7fe9b1414750b874f71e4225c', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362138641, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('7e357081097c974c8b6717d265c3627a', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138709, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('ff657a8439237767646394e8d592badb', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362138794, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('2e8a8a7d74bdf6073e4b0605eef23af3', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138859, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('a02dbbbebcfede453f41d64ad710914a', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362138941, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('7171b4ec2299618400a0037736d2e209', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141889, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('02fd691362071070d6ab0ef9f292e30f', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141949, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('3974444859cd012111a3fee1cdb82585', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142039, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('06a9b91dcff6405d55a67d3ff4af15dd', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142099, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('4ea9978e6b68ea95952a38b427e79789', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142189, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('9746982cad42d5436fe95b32f346be31', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142249, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('2e86baad506902ed74fe442b5d8a0fbd', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141319, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('c8192b3a0b45e71e74eb109deca8ba15', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141380, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('0a915ed4f2378fd7cd71119ddc1bdbb3', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141470, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5d61bd189832cfc029749d3bce87c283', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141529, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('a4c07cfe92082fc0978c63aee6860747', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141620, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('3ce251772423687c992f2e3b0bf51095', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141679, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('fc81ea3134409f16d5bd8615f71b7993', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141770, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('e5d24e80c87eaa0e6521fc51f2c6b91e', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141829, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('e7c205f01047b832b72a3222360f3093', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141919, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('53c62f5299c03aa0e6d931675edbca7d', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138290, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('3557fd6d991f1ebf23f1e6b372e03ce5', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138349, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('0b5d76b9e63b231cd9f26ff858f305bc', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138439, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('96f2d5abf8d8cb26747eaadb07b6c242', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138499, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('161a9a6e7d46c31762aa811ef98d5c31', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138589, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('d198c3a51574d84b081330aedcf5abfa', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138649, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('b03e9265ff8f28f7d0e8908a7f3a42f6', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138739, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('c28255545dbb78922919140ad2529112', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138799, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('33ce6fa7ef2b1dbf764eb5d10cda6568', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138889, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5ce1f34ea1ecd952c437dbddefad06b2', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141979, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5d97a733668ad7b99bb1662abc7f9e52', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142069, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5b81b7d707dfad5413315d8b6582935e', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142129, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('22da413afebae08f95402862319fbbf4', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142219, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('50fe8fd38ec613127d76d02b91b88d66', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138949, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('f024a61e119e1ff7558fe3d2ae770280', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142279, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('f313c0fa8fcbee7eb195486269067d0e', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362141341, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('69dbe7b981141212c6039ff598cb8933', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141409, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('4e123dc2fc95470de427ee9fd75cae6d', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362141494, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('7a323f54c5d682ad346ad1c3ab82f104', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141559, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('ff41a1289753b5da1700d285b508d6fa', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362141641, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('486b7d31406a704ee648cfc8e217ea82', '90.236.108.89', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_2) AppleWebKit/536.26.17 (KHTML, like Gecko) Version/6.0.2 Safari/536.26.17', 1362144597, 'a:3:{s:9:"user_data";s:0:"";s:9:"auth_user";s:1:"1";s:13:"auth_loggedin";b:1;}');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('44a98f5c80bc23ddbf7c94f8e3d34932', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138319, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('72275f9767d5f546f539bab6b7b0fb68', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138379, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5c777a356a6140ec117c2a71ded3ec4a', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138469, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('11b22100ca2442a60fcdf60e5a7f94ee', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138530, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('67410739f4c4692ae83ea82438aab87c', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138619, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('67fe376333449a7e379481c3ec4e0747', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138679, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5bb9bc77390c233e57bd2df9ae29c6a2', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138769, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('3535bd2f431013e104699eb64abdc78e', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138829, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('ec00b7ffcac670654db0a67ed6f65290', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138919, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('93ff0f8f856fb4aec8cf1f4ab889f7d0', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141709, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('f4b3d4b3d47930a0c9b5e77c0d34d234', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362141794, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('57cf5b852a8dc3c5373497ff29538f32', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362141859, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('c8b29293ea7e70edba5ecf03b18ee13b', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362141940, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('d6c05755d9695f192f539f82c9620cd4', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142009, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('88a2fd84f6457253cc6991b1c57eb0af', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362142094, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('7fd2fcf07d67025f78ae5faa3e0b7361', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362138979, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('01c9375e684d28ed15603cdb6dc10809', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142159, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('d8f3775819ce3513f4d5cd483e3b3544', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362142241, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('f3b810406a75c9323d8d1cda2593256d', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139009, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('042dd5e000a82c8e88533096302d2edf', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362139094, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('83abf30619acd54f1d054dff0472a819', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139159, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5e9087f13415ed932b9cd3cc54c0ef06', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362139241, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('7e5782d76cdbee6236aef8bad7903271', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139309, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('f0dca9864b1069cee8ff62ea8a0a3cc9', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362139394, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('1fd2c09918490135cc379e76d5d46e6d', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139459, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('85a2581e51d181b2afeaa27c3014c741', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362139541, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('1de2316c1ef233d07fbdf5eb78d31b64', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139609, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('f3e742e773ec220a61e6437b449646a5', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362139694, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('05205e02c106136662b1ded2aa6048c4', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139759, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('e6b33e21a0b29a8eab18a273f493de2b', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142309, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('dcab6e8c7e8cf7953f9f4eb025f3ab18', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362142394, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('ac038e037d55bf0d3516df58d45eb57b', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142459, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('3192305d215d70d6e86f80882130c5dc', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362142541, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('a986786e10a5bc3095c3bc433a8a9c4e', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142609, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('e1ed26406cbdb5a7cb434d4c7864e9c5', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362142694, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('2d72b1f6d976fafcbeb6195401cc2465', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142759, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('91e5cc8a0f9e603b91fdd8b819eb7fb2', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362142841, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('c775dba01475aa5e5fff0bb6b9af52aa', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142909, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('a3179101bd2c45a73cc52eb3de047d9f', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362142994, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('b4f0b71188b8fbc60ab64b283efc8928', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143059, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('79666c443b8686e8540fdab52a0fec72', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362143141, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('9fab24fd8cacb9c7c2938673f142e9b7', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362139841, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('a8b622343617487354bd2bec3ec0d5e5', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139909, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5c8f4d8aae5a96a90fb6a22d53ce916b', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362139995, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('911cc73690e924a175b1a078ac0d8307', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140059, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('aec0dcbfac77fbf96a608e624e1f9b5f', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362140141, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('733637649ba285f3c9abda748ad148a6', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140209, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('d86c887d5b0a847f88d1a67c639678f2', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143209, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('6d301542b943f3c1e9d892e8e17796a4', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143269, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('1106cba211270db2800ea30d2b976ffa', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143299, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('54489b395f29f8b14c09376c17479439', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362143441, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('72af8599bad28a3fe4eea1a9c122f673', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139040, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('978e77351790f11025a2cd80a55328f6', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139099, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('d5823fc63bc4574cc15e448e9e23976b', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139189, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('494454f379c0bc1e332c1c0126b976c6', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139249, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5cd72aeff4375199d553618ddf691aaf', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139339, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('bfc9195a69356ab6956a9e3fa19ad219', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139399, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('70fe8779a421c2772d706c3af227a104', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139489, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('109ad4ffeb158283badada4979056006', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139550, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('825c352d874bc728620c42b65bf2e7d8', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139639, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('f42b626b3f64acfc508f65612d939e6a', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139699, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('0b21fabe28514239b8c8c082eebf60ce', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139789, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('199ab4ce8801e6d18b5c20cc36453b00', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139850, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('e287abea193ad1d468c12f2f8cf918af', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142339, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('17d268a6bfab62323546101ebde4b7c3', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142399, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('d88b4edf71ff4cfadaca2e8a03ba1dc7', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142489, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('9c0f25eed6eb9185e528d0b456a54552', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142549, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('8666b690af9610b876b810beb541fada', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142639, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('92cd7ff802dc27f66c53fffc769b37a9', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142699, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('4ce73baa184824ec8fea68b45f74fd33', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142789, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('86581213637f9908e2ca6a7a138283f8', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142850, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('86bd012d4ae6da8ee940a1f8333abb2d', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142939, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('bdbb798f6701ae512563282c24895f77', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142999, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('b7edb49a1683c03aa2891b54ec2a94e8', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143089, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('d082e02f03f70b1e62907e926c78afed', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143149, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5ebc7235fe8f8b6b2e940b7f93494ba8', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139939, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('fec25db0e7af21400fb5766c31f38c9e', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139999, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('f0778e5306cdc2c6382f8cb19a8f9e57', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140089, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('86b1aa9070e742c53b5d256dec6d0106', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140149, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('616c954125b4f27eae62c3c7bffae0b7', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140239, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('c88d54edd7de662a872ef40180b3d283', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143239, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('a684a9006a8359f4306b7c75ada0d0d1', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143329, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('9b5427254ad816f849599f167c3eb516', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143390, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('8a144b46d4ccde4d0d43f76678e04f53', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143449, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('88c97386f870989e6659d89b4f82e822', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142369, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('cce1f55422bd310ceff91b0129b03255', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142429, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('70a7b810c2ea3f46dac0382baed50a7f', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139069, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('fd5be47da56a1bc4bb16d1eb1a911785', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139129, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('61db230f4a806db9eecf583c58bb7e86', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139219, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('487b0dc17dfb22f5afc9e4705bea7cdc', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139279, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('409034749fe5c3b453e1cfaebced6c8f', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139370, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('3331d7d31f04af0b79e3a71f36850a07', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139429, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('8ad081da8861ece269c1c48f2bda1cbb', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139519, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('cce916bd98908b6f7d1463da174197ca', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139579, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('129b8ec36c59f1677aa7c0e2d09e5232', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139669, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('f3ca1752a6be722eb90de3c303891b66', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142519, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('a5a39c8b8a613355282d602bc08e8c31', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142579, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('27e6cb0c3f1a5f5332e5740c2c2dee25', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142669, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('d1712c9d354e01200ec309e49bec8fce', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142729, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('32887cac7bca27a1827adbe3ad8f6c3f', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142819, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('d17cc2ae7c96d7cea96a1847ff91f263', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142879, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('ca6a1a246767d5c4db0903ad796d06f4', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362142969, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('dadefd5d944ce28785a7a42e8b98992d', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143029, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('68b1220494d729d75404da22ee7d588f', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143119, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('6332b52884db76866da504f2510b20d9', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143179, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('7847cba03aaba5999581eaa3b96b0b8a', '192.168.160.1', 'check_http/v1.4.15 (nagios-plugins 1.4.15)', 1362143294, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('53412d79065dad16a6c41f84542ee8ba', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143359, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('2de8eb7467668ddd7aef790f358b95ac', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139729, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('3a79c523974480a4a4fe651acd0b396f', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139819, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('d4d909527c1276afa0dbedd47dce68f2', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139879, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('5884f238855fa2f3228cc03c528c9d4a', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362139969, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('9099bd020e2a557dbb7ad69e63033cd6', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140029, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('44be2bad532b479f3353d5918c91dfad', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140119, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('42b9d379331bf3651eda006a7cb37372', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362140179, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('de484bffe56a8e2a35b2459b59a42df8', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143419, '');
INSERT INTO ci_sessions (session_id, ip_address, user_agent, last_activity, user_data) VALUES ('4b8bc82500208e909710f46d6c3237f4', '10.200.57.129', 'ELB-HealthChecker/1.0', 1362143479, '');


--
-- TOC entry 2049 (class 0 OID 16774)
-- Dependencies: 182
-- Data for Name: keys; Type: TABLE DATA; Schema: klubb; Owner: klubb
--

INSERT INTO keys (id, key, level, ignore_limits, is_private_key, ip_addresses, date_created) VALUES (1, 'b84b9eb779c6706ce75584c29b8005b1', 2, false, false, NULL, 0);


--
-- TOC entry 2077 (class 0 OID 0)
-- Dependencies: 181
-- Name: keys_id_seq; Type: SEQUENCE SET; Schema: klubb; Owner: klubb
--

SELECT pg_catalog.setval('keys_id_seq', 1, true);


--
-- TOC entry 2053 (class 0 OID 16798)
-- Dependencies: 186
-- Data for Name: limits; Type: TABLE DATA; Schema: klubb; Owner: klubb
--



--
-- TOC entry 2078 (class 0 OID 0)
-- Dependencies: 185
-- Name: limits_id_seq; Type: SEQUENCE SET; Schema: klubb; Owner: klubb
--

SELECT pg_catalog.setval('limits_id_seq', 1, false);


--
-- TOC entry 2034 (class 0 OID 16630)
-- Dependencies: 166
-- Data for Name: log; Type: TABLE DATA; Schema: klubb; Owner: klubb
--



--
-- TOC entry 2079 (class 0 OID 0)
-- Dependencies: 165
-- Name: log_id_seq; Type: SEQUENCE SET; Schema: klubb; Owner: klubb
--

SELECT pg_catalog.setval('log_id_seq', 1, true);


--
-- TOC entry 2051 (class 0 OID 16787)
-- Dependencies: 184
-- Data for Name: logs; Type: TABLE DATA; Schema: klubb; Owner: klubb
--

INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (1, 'role/rights/role/2/X-API-KEY/b84b9eb779c6706ce75584c29b8005b1', 'get', 'a:2:{s:4:"role";s:1:"2";s:9:"X-API-KEY";s:32:"b84b9eb779c6706ce75584c29b8005b1";}', 'b84b9eb779c6706ce75584c29b8005b1', '10.0.1.200', 1357947309, false);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (2, 'role/rights/role/2/X-API-KEY/b84b9eb779c6706ce75584c29b8005b1', 'get', 'a:2:{s:4:"role";s:1:"2";s:9:"X-API-KEY";s:32:"b84b9eb779c6706ce75584c29b8005b1";}', 'b84b9eb779c6706ce75584c29b8005b1', '10.0.1.200', 1357947348, false);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (3, 'role/rights/role/2/X-API-KEY/b84b9eb779c6706ce75584c29b8005b1', 'get', 'a:2:{s:4:"role";s:1:"2";s:9:"X-API-KEY";s:32:"b84b9eb779c6706ce75584c29b8005b1";}', 'b84b9eb779c6706ce75584c29b8005b1', '10.0.1.200', 1357947449, false);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (4, 'role/rights/role/2/X-API-KEY/b84b9eb779c6706ce75584c29b8005b1', 'get', 'a:2:{s:4:"role";s:1:"2";s:9:"X-API-KEY";s:32:"b84b9eb779c6706ce75584c29b8005b1";}', 'b84b9eb779c6706ce75584c29b8005b1', '10.0.1.200', 1357947567, false);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (5, 'role/rights/role/2/X-API-KEY/b84b9eb779c6706ce75584c29b8005b1', 'get', 'a:2:{s:4:"role";s:1:"2";s:9:"X-API-KEY";s:32:"b84b9eb779c6706ce75584c29b8005b1";}', 'b84b9eb779c6706ce75584c29b8005b1', '10.0.1.200', 1357985995, false);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (6, 'role/rights/role/2/X-API-KEY/b84b9eb779c6706ce75584c29b8005b1', 'get', 'a:2:{s:4:"role";s:1:"2";s:9:"X-API-KEY";s:32:"b84b9eb779c6706ce75584c29b8005b1";}', 'b84b9eb779c6706ce75584c29b8005b1', '10.0.1.200', 1358034683, false);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (7, 'role/rights/role/2/X-API-KEY/b84b9eb779c6706ce75584c29b8005b1', 'get', 'a:2:{s:4:"role";s:1:"2";s:9:"X-API-KEY";s:32:"b84b9eb779c6706ce75584c29b8005b1";}', 'b84b9eb779c6706ce75584c29b8005b1', '10.0.1.200', 1358114101, false);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (8, 'role/rights/role/2/X-API-KEY/b84b9eb779c6706ce75584c29b8005b1', 'get', 'a:2:{s:4:"role";s:1:"2";s:9:"X-API-KEY";s:32:"b84b9eb779c6706ce75584c29b8005b1";}', 'b84b9eb779c6706ce75584c29b8005b1', '10.0.1.200', 1358114242, false);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (9, 'role/rights/role/2/X-API-KEY/b84b9eb779c6706ce75584c29b8005b1', 'get', 'a:2:{s:4:"role";s:1:"2";s:9:"X-API-KEY";s:32:"b84b9eb779c6706ce75584c29b8005b1";}', 'b84b9eb779c6706ce75584c29b8005b1', '10.0.1.200', 1358114258, false);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (10, 'role/rights/role/2', 'get', 'a:1:{s:4:"role";s:1:"2";}', '', '10.0.1.200', 1358114423, true);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (11, 'role/rights/role/2', 'get', 'a:1:{s:4:"role";s:1:"2";}', '', '10.0.1.200', 1358116197, true);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (12, 'role/rights/role/2', 'get', 'a:1:{s:4:"role";s:1:"2";}', '', '10.0.1.200', 1358118631, true);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (13, 'role/rights/role/2', 'get', 'a:1:{s:4:"role";s:1:"2";}', '', '10.0.1.200', 1358118734, true);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (14, 'role/rights/role/2', 'get', 'a:1:{s:4:"role";s:1:"2";}', '', '94.254.4.85', 1360153339, true);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (15, 'role/rights/role/2', 'get', 'a:1:{s:4:"role";s:1:"2";}', '', '94.254.4.85', 1360776970, true);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (16, 'role/rights/role/2', 'get', 'a:1:{s:4:"role";s:1:"2";}', '', '94.254.4.85', 1360776996, true);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (17, 'role/rights/role/2', 'get', 'a:1:{s:4:"role";s:1:"2";}', '', '46.194.178.55', 1360929401, true);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (18, 'role/rights/role/2', 'get', 'a:1:{s:4:"role";s:1:"2";}', '', '90.236.236.212', 1361276689, true);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (19, 'role/rights/role/2', 'get', 'a:1:{s:4:"role";s:1:"2";}', '', '90.236.236.212', 1361277509, true);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (20, 'role/rights/role/2', 'get', 'a:1:{s:4:"role";s:1:"2";}', '', '95.196.131.37', 1361279347, true);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (21, 'role/rights/role/2', 'get', 'a:1:{s:4:"role";s:1:"2";}', '', '95.196.131.37', 1361279945, true);
INSERT INTO logs (id, uri, method, params, api_key, ip_address, "time", authorized) VALUES (22, 'role/rights/role/2', 'get', 'a:1:{s:4:"role";s:1:"2";}', '', '95.196.131.37', 1361279995, true);


--
-- TOC entry 2080 (class 0 OID 0)
-- Dependencies: 183
-- Name: logs_id_seq; Type: SEQUENCE SET; Schema: klubb; Owner: klubb
--

SELECT pg_catalog.setval('logs_id_seq', 22, true);


--
-- TOC entry 2036 (class 0 OID 16653)
-- Dependencies: 168
-- Data for Name: member_data; Type: TABLE DATA; Schema: klubb; Owner: klubb
--



--
-- TOC entry 2081 (class 0 OID 0)
-- Dependencies: 167
-- Name: member_data_id_seq; Type: SEQUENCE SET; Schema: klubb; Owner: klubb
--

SELECT pg_catalog.setval('member_data_id_seq', 1, false);


--
-- TOC entry 2055 (class 0 OID 25557)
-- Dependencies: 188
-- Data for Name: member_flags; Type: TABLE DATA; Schema: klubb; Owner: klubb
--



--
-- TOC entry 2082 (class 0 OID 0)
-- Dependencies: 187
-- Name: member_flags_id_seq; Type: SEQUENCE SET; Schema: klubb; Owner: klubb
--

SELECT pg_catalog.setval('member_flags_id_seq', 1, false);


--
-- TOC entry 2038 (class 0 OID 16665)
-- Dependencies: 170
-- Data for Name: members; Type: TABLE DATA; Schema: klubb; Owner: klubb
--



--
-- TOC entry 2083 (class 0 OID 0)
-- Dependencies: 169
-- Name: members_id_seq; Type: SEQUENCE SET; Schema: klubb; Owner: klubb
--

SELECT pg_catalog.setval('members_id_seq', 33, true);


--
-- TOC entry 2047 (class 0 OID 16763)
-- Dependencies: 180
-- Data for Name: migrations; Type: TABLE DATA; Schema: klubb; Owner: klubb
--

INSERT INTO migrations (version) VALUES (0);


--
-- TOC entry 2040 (class 0 OID 16678)
-- Dependencies: 172
-- Data for Name: rights; Type: TABLE DATA; Schema: klubb; Owner: klubb
--

INSERT INTO rights (id, role, add_members, add_users, use_system) VALUES (1, 1, true, false, true);
INSERT INTO rights (id, role, add_members, add_users, use_system) VALUES (2, 2, false, false, true);
INSERT INTO rights (id, role, add_members, add_users, use_system) VALUES (3, 3, true, true, true);


--
-- TOC entry 2084 (class 0 OID 0)
-- Dependencies: 171
-- Name: rights_id_seq; Type: SEQUENCE SET; Schema: klubb; Owner: klubb
--

SELECT pg_catalog.setval('rights_id_seq', 3, true);


--
-- TOC entry 2042 (class 0 OID 16695)
-- Dependencies: 174
-- Data for Name: roles; Type: TABLE DATA; Schema: klubb; Owner: klubb
--

INSERT INTO roles (id, name, system) VALUES (1, 'Administratör', true);
INSERT INTO roles (id, name, system) VALUES (2, 'Användare', true);
INSERT INTO roles (id, name, system) VALUES (3, 'Superadministratör', true);


--
-- TOC entry 2085 (class 0 OID 0)
-- Dependencies: 173
-- Name: roles_id_seq; Type: SEQUENCE SET; Schema: klubb; Owner: klubb
--

SELECT pg_catalog.setval('roles_id_seq', 3, true);


--
-- TOC entry 2043 (class 0 OID 16704)
-- Dependencies: 175
-- Data for Name: system; Type: TABLE DATA; Schema: klubb; Owner: klubb
--

INSERT INTO system (key, value) VALUES ('org_name', 'Ung Cancer');
INSERT INTO system (key, value) VALUES ('app_name', 'Medlemsregistret');
INSERT INTO system (key, value) VALUES ('org_type', 'förening');
INSERT INTO system (key, value) VALUES ('inactive_title', 'Avliden');
INSERT INTO system (key, value) VALUES ('inactive_date_title', 'Datum');


--
-- TOC entry 2045 (class 0 OID 16711)
-- Dependencies: 177
-- Data for Name: types; Type: TABLE DATA; Schema: klubb; Owner: klubb
--

INSERT INTO types (id, name, plural, "desc") VALUES (1, 'Medlem', 'Medlemmar', NULL);
INSERT INTO types (id, name, plural, "desc") VALUES (2, 'Anhörigmedlem', 'Anhörigmedlemmar', NULL);
INSERT INTO types (id, name, plural, "desc") VALUES (3, 'Stödmedlem', 'Stödmedlemmar', NULL);


--
-- TOC entry 2086 (class 0 OID 0)
-- Dependencies: 176
-- Name: types_id_seq; Type: SEQUENCE SET; Schema: klubb; Owner: klubb
--

SELECT pg_catalog.setval('types_id_seq', 3, true);


--
-- TOC entry 2056 (class 0 OID 25565)
-- Dependencies: 189
-- Data for Name: types_requirements; Type: TABLE DATA; Schema: klubb; Owner: klubb
--

INSERT INTO types_requirements (fieldname, type, fieldtype, rule, rule_desc, sort_order) VALUES ('ssid', 1, 'text', 'required', 'required', 1);
INSERT INTO types_requirements (fieldname, type, fieldtype, rule, rule_desc, sort_order) VALUES ('firstname', 1, 'text', 'required', 'required', 2);
INSERT INTO types_requirements (fieldname, type, fieldtype, rule, rule_desc, sort_order) VALUES ('lastname', 1, 'text', 'required', 'required', 2);
INSERT INTO types_requirements (fieldname, type, fieldtype, rule, rule_desc, sort_order) VALUES ('phone', 1, 'tel', 'required', 'required', 3);
INSERT INTO types_requirements (fieldname, type, fieldtype, rule, rule_desc, sort_order) VALUES ('email', 1, 'email', 'required', 'required', 3);
INSERT INTO types_requirements (fieldname, type, fieldtype, rule, rule_desc, sort_order) VALUES ('address', 1, 'text', 'required', 'required', 4);
INSERT INTO types_requirements (fieldname, type, fieldtype, rule, rule_desc, sort_order) VALUES ('city', 1, 'text', 'required', 'required', 5);
INSERT INTO types_requirements (fieldname, type, fieldtype, rule, rule_desc, sort_order) VALUES ('zipcode', 1, 'text', 'required', 'required', 5);
INSERT INTO types_requirements (fieldname, type, fieldtype, rule, rule_desc, sort_order) VALUES ('diagnos', 1, 'text', 'required', 'required', 6);
INSERT INTO types_requirements (fieldname, type, fieldtype, rule, rule_desc, sort_order) VALUES ('cancer', 1, 'text', 'required', 'required', 6);
INSERT INTO types_requirements (fieldname, type, fieldtype, rule, rule_desc, sort_order) VALUES ('tell', 1, 'checkbox', 'optional', 'optional', 7);
INSERT INTO types_requirements (fieldname, type, fieldtype, rule, rule_desc, sort_order) VALUES ('talking_partner', 1, 'checkbox', 'optional', 'optional', 7);
INSERT INTO types_requirements (fieldname, type, fieldtype, rule, rule_desc, sort_order) VALUES ('participate', 1, 'checkbox', 'optional', 'optional', 7);


--
-- TOC entry 2046 (class 0 OID 16739)
-- Dependencies: 178
-- Data for Name: user_role; Type: TABLE DATA; Schema: klubb; Owner: klubb
--

INSERT INTO user_role ("user", role) VALUES (1, 3);
INSERT INTO user_role ("user", role) VALUES (2, 1);


--
-- TOC entry 2031 (class 0 OID 16594)
-- Dependencies: 163
-- Data for Name: users; Type: TABLE DATA; Schema: klubb; Owner: klubb
--

INSERT INTO users (id, username, firstname, lastname, email, phone, key, password, registered, first_login, loggedin) VALUES (20, 'johanna', 'Johanna', NULL, 'johanna@ungcancer.se', NULL, 'l4rsCxzFjTfmHCq9HQK2lsOV+/TSVIdfhytYDTBUz8KAP+c6JJ2/bupPYuC1E2kilcF/7G6c252VuAuwfAXThQ==', '$2a$08$XfaTOeuiFvD98BnFmG0Hxe2zIcGmTJlyfboSkGk8zfiAcCBMvyvYu', 1361276768, true, true);
INSERT INTO users (id, username, firstname, lastname, email, phone, key, password, registered, first_login, loggedin) VALUES (21, 'petra', 'Petra', 'Lindblom', 'petra@nyfagel.se', NULL, 'siVGWfAlYSH4Evzt2Ip4zi/uNZJoaUC1mOuWM7LCyJiHDHSjQQKHiI+nSzKXaIa8/+5bTgHnZXSzEeYeNnm9Xg==', '$2a$08$krbVg/JLdQOmLk.3aWD4CuvSfMYEB6Lx/xBSU5a6lTGUqmFUZ6zSS', 1361281567, true, false);
INSERT INTO users (id, username, firstname, lastname, email, phone, key, password, registered, first_login, loggedin) VALUES (1, 'jan', 'Jan', 'Lindblom', 'jan@nyfagel.se', '0731-509 338', 'w87p7zjWbikHvJ+yja0QqmCmWMb1GAjLNZjUYoMW/tiSKrK9iz+fpvwX4JFePKv3C1BTqE6U/2oR73RWAw7ljw==', '$2a$08$gNc6oAbrTTT3tVeHTnE9je2oSUPvv7bxnzoE3qG9mks8EZZtgEHYG', 1355861094, false, true);
INSERT INTO users (id, username, firstname, lastname, email, phone, key, password, registered, first_login, loggedin) VALUES (2, 'judith', 'Judith', 'Loménius', 'judith@ungcancer.se', NULL, 'gLr0CLylw4x6+5z/zUcmSA9eSJqfgdBAYGE2zWglXtRuuaWBjAy/lJHA+maTWpkq9rcGlGZhCxuxsLMOxmXq8A==', '$2a$08$yKhrHx8iSn6tpecXOdw1Zu/.rdOzgxJSkU3rHDGKGydtwma4BOW3y', 1357119301, false, true);
INSERT INTO users (id, username, firstname, lastname, email, phone, key, password, registered, first_login, loggedin) VALUES (19, 'pelle', 'Pelle', 'Landeberg', 'pelle@nyfagel.se', NULL, '/HOaws3hZQCb43ncd+4fVjVqO3GyaFlKdf4BxGbmM6HdTUHnfFlMXl0QXKaQ4P+4zhG9dFXao/0xy++hD7JbHA==', '$2a$08$pKfE3m3B7uN2otF5/Y3eT.CWj9h/9YoYoGZkBA/ApHDBvB6beYu22', 1361276703, true, true);


--
-- TOC entry 2087 (class 0 OID 0)
-- Dependencies: 162
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: klubb; Owner: klubb
--

SELECT pg_catalog.setval('users_id_seq', 21, true);


--
-- TOC entry 1982 (class 2606 OID 16589)
-- Name: authentication_pkey; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY authentication
    ADD CONSTRAINT authentication_pkey PRIMARY KEY ("user", series);


--
-- TOC entry 1990 (class 2606 OID 16627)
-- Name: ci_sessions_pk; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY ci_sessions
    ADD CONSTRAINT ci_sessions_pk PRIMARY KEY (session_id);


--
-- TOC entry 2008 (class 2606 OID 16805)
-- Name: keys_key_key; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY keys
    ADD CONSTRAINT keys_key_key UNIQUE (key);


--
-- TOC entry 2010 (class 2606 OID 16784)
-- Name: keys_pkey; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY keys
    ADD CONSTRAINT keys_pkey PRIMARY KEY (id);


--
-- TOC entry 2016 (class 2606 OID 16803)
-- Name: limits_pkey; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY limits
    ADD CONSTRAINT limits_pkey PRIMARY KEY (id);


--
-- TOC entry 1992 (class 2606 OID 16635)
-- Name: log_pk; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY log
    ADD CONSTRAINT log_pk PRIMARY KEY (id);


--
-- TOC entry 2013 (class 2606 OID 16795)
-- Name: logs_pkey; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY logs
    ADD CONSTRAINT logs_pkey PRIMARY KEY (id);


--
-- TOC entry 1994 (class 2606 OID 16662)
-- Name: member_data_pk; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY member_data
    ADD CONSTRAINT member_data_pk PRIMARY KEY (id);


--
-- TOC entry 2018 (class 2606 OID 25564)
-- Name: member_flags_key_key; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY member_flags
    ADD CONSTRAINT member_flags_key_key UNIQUE (key);


--
-- TOC entry 2020 (class 2606 OID 25562)
-- Name: member_flags_pkey; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY member_flags
    ADD CONSTRAINT member_flags_pkey PRIMARY KEY (id);


--
-- TOC entry 1996 (class 2606 OID 16733)
-- Name: members_pkey; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY members
    ADD CONSTRAINT members_pkey PRIMARY KEY (id);


--
-- TOC entry 1984 (class 2606 OID 16604)
-- Name: pk_id; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY users
    ADD CONSTRAINT pk_id PRIMARY KEY (id);


--
-- TOC entry 1998 (class 2606 OID 16703)
-- Name: rights_pkey; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY rights
    ADD CONSTRAINT rights_pkey PRIMARY KEY (id);


--
-- TOC entry 2000 (class 2606 OID 16700)
-- Name: roles_pkey; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- TOC entry 2002 (class 2606 OID 16708)
-- Name: system_pkey; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY system
    ADD CONSTRAINT system_pkey PRIMARY KEY (key);


--
-- TOC entry 2004 (class 2606 OID 16716)
-- Name: types_pkey; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY types
    ADD CONSTRAINT types_pkey PRIMARY KEY (id);


--
-- TOC entry 2022 (class 2606 OID 25569)
-- Name: types_requirements_pkey; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY types_requirements
    ADD CONSTRAINT types_requirements_pkey PRIMARY KEY (fieldname, type, rule);


--
-- TOC entry 1986 (class 2606 OID 16608)
-- Name: u_email; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY users
    ADD CONSTRAINT u_email UNIQUE (email);


--
-- TOC entry 1988 (class 2606 OID 16606)
-- Name: u_username; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY users
    ADD CONSTRAINT u_username UNIQUE (username);


--
-- TOC entry 2006 (class 2606 OID 16743)
-- Name: user_role_pkey; Type: CONSTRAINT; Schema: klubb; Owner: klubb; Tablespace: 
--

ALTER TABLE ONLY user_role
    ADD CONSTRAINT user_role_pkey PRIMARY KEY ("user", role);


--
-- TOC entry 2011 (class 1259 OID 16811)
-- Name: fki_api_key_fk; Type: INDEX; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE INDEX fki_api_key_fk ON logs USING btree (api_key);


--
-- TOC entry 2014 (class 1259 OID 16822)
-- Name: fki_limits_api_key_fk; Type: INDEX; Schema: klubb; Owner: klubb; Tablespace: 
--

CREATE INDEX fki_limits_api_key_fk ON limits USING btree (api_key);


--
-- TOC entry 2025 (class 2606 OID 25544)
-- Name: members_data_fkey; Type: FK CONSTRAINT; Schema: klubb; Owner: klubb
--

ALTER TABLE ONLY members
    ADD CONSTRAINT members_data_fkey FOREIGN KEY (data) REFERENCES member_data(id) ON DELETE SET NULL;


--
-- TOC entry 2026 (class 2606 OID 25549)
-- Name: members_type_fkey; Type: FK CONSTRAINT; Schema: klubb; Owner: klubb
--

ALTER TABLE ONLY members
    ADD CONSTRAINT members_type_fkey FOREIGN KEY (type) REFERENCES types(id) ON DELETE SET NULL;


--
-- TOC entry 2023 (class 2606 OID 16609)
-- Name: user_fk; Type: FK CONSTRAINT; Schema: klubb; Owner: klubb
--

ALTER TABLE ONLY authentication
    ADD CONSTRAINT user_fk FOREIGN KEY ("user") REFERENCES users(id) ON DELETE CASCADE;


--
-- TOC entry 2024 (class 2606 OID 16636)
-- Name: user_fk; Type: FK CONSTRAINT; Schema: klubb; Owner: klubb
--

ALTER TABLE ONLY log
    ADD CONSTRAINT user_fk FOREIGN KEY ("user") REFERENCES users(id) ON DELETE SET NULL;


--
-- TOC entry 2028 (class 2606 OID 16749)
-- Name: user_role_role_fkey; Type: FK CONSTRAINT; Schema: klubb; Owner: klubb
--

ALTER TABLE ONLY user_role
    ADD CONSTRAINT user_role_role_fkey FOREIGN KEY (role) REFERENCES roles(id) ON DELETE CASCADE;


--
-- TOC entry 2027 (class 2606 OID 16744)
-- Name: user_role_user_fkey; Type: FK CONSTRAINT; Schema: klubb; Owner: klubb
--

ALTER TABLE ONLY user_role
    ADD CONSTRAINT user_role_user_fkey FOREIGN KEY ("user") REFERENCES users(id) ON DELETE CASCADE;


--
-- TOC entry 2064 (class 0 OID 0)
-- Dependencies: 6
-- Name: klubb; Type: ACL; Schema: -; Owner: klubb
--

REVOKE ALL ON SCHEMA klubb FROM PUBLIC;
REVOKE ALL ON SCHEMA klubb FROM klubb;
GRANT ALL ON SCHEMA klubb TO klubb;
GRANT ALL ON SCHEMA klubb TO PUBLIC;


-- Completed on 2013-03-01 14:45:40 CET

--
-- PostgreSQL database dump complete
--

