-- Downloaded from: https://github.com/InnovaTech-Official/LMS/blob/e677201ec508b302b6ed189aa60ed7319965af58/loan_db.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 17.4
-- Dumped by pg_dump version 17.4

-- Started on 2025-05-29 02:19:11

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
-- TOC entry 238 (class 1255 OID 33183)
-- Name: update_timestamp(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
            BEGIN
                NEW.updated_at = CURRENT_TIMESTAMP;
                RETURN NEW;
            END;
            $$;


ALTER FUNCTION public.update_timestamp() OWNER TO postgres;

--
-- TOC entry 237 (class 1255 OID 16832)
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = CURRENT_TIMESTAMP;
   RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_updated_at_column() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 232 (class 1259 OID 33217)
-- Name: accounts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.accounts (
    id integer NOT NULL,
    sub_account_id integer NOT NULL,
    name character varying(255) NOT NULL
);


ALTER TABLE public.accounts OWNER TO postgres;

--
-- TOC entry 228 (class 1259 OID 33198)
-- Name: accounts_head; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.accounts_head (
    id integer NOT NULL,
    name character varying(50) NOT NULL
);


ALTER TABLE public.accounts_head OWNER TO postgres;

--
-- TOC entry 227 (class 1259 OID 33197)
-- Name: accounts_head_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.accounts_head_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.accounts_head_id_seq OWNER TO postgres;

--
-- TOC entry 4974 (class 0 OID 0)
-- Dependencies: 227
-- Name: accounts_head_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.accounts_head_id_seq OWNED BY public.accounts_head.id;


--
-- TOC entry 231 (class 1259 OID 33216)
-- Name: accounts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.accounts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.accounts_id_seq OWNER TO postgres;

--
-- TOC entry 4975 (class 0 OID 0)
-- Dependencies: 231
-- Name: accounts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.accounts_id_seq OWNED BY public.accounts.id;


--
-- TOC entry 236 (class 1259 OID 33236)
-- Name: area; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.area (
    id integer NOT NULL,
    city_id integer NOT NULL,
    name character varying(100) NOT NULL
);


ALTER TABLE public.area OWNER TO postgres;

--
-- TOC entry 235 (class 1259 OID 33235)
-- Name: area_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.area_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.area_id_seq OWNER TO postgres;

--
-- TOC entry 4976 (class 0 OID 0)
-- Dependencies: 235
-- Name: area_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.area_id_seq OWNED BY public.area.id;


--
-- TOC entry 226 (class 1259 OID 33186)
-- Name: bank_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.bank_info (
    id integer NOT NULL,
    bank_name character varying(255) NOT NULL,
    branch_name character varying(255),
    branch_code character varying(50),
    account_number character varying(50) NOT NULL,
    account_title character varying(255) NOT NULL,
    iban character varying(50),
    swift_code character varying(50),
    bank_address text,
    contact_person character varying(255),
    contact_number character varying(50),
    email character varying(255),
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.bank_info OWNER TO postgres;

--
-- TOC entry 225 (class 1259 OID 33185)
-- Name: bank_info_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.bank_info_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.bank_info_id_seq OWNER TO postgres;

--
-- TOC entry 4977 (class 0 OID 0)
-- Dependencies: 225
-- Name: bank_info_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.bank_info_id_seq OWNED BY public.bank_info.id;


--
-- TOC entry 234 (class 1259 OID 33229)
-- Name: city; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.city (
    id integer NOT NULL,
    name character varying(100) NOT NULL
);


ALTER TABLE public.city OWNER TO postgres;

--
-- TOC entry 233 (class 1259 OID 33228)
-- Name: city_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.city_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.city_id_seq OWNER TO postgres;

--
-- TOC entry 4978 (class 0 OID 0)
-- Dependencies: 233
-- Name: city_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.city_id_seq OWNED BY public.city.id;


--
-- TOC entry 224 (class 1259 OID 33173)
-- Name: company_settings; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.company_settings (
    id integer NOT NULL,
    company_name character varying(100) NOT NULL,
    company_logo character varying(255),
    address_line1 character varying(100),
    address_line2 character varying(100),
    city character varying(50),
    state character varying(50),
    postal_code character varying(20),
    country character varying(50),
    phone_1 character varying(20),
    phone_2 character varying(20),
    email character varying(100),
    website character varying(100),
    tax_id character varying(50),
    registration_number character varying(50),
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.company_settings OWNER TO postgres;

--
-- TOC entry 223 (class 1259 OID 33172)
-- Name: company_settings_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.company_settings_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.company_settings_id_seq OWNER TO postgres;

--
-- TOC entry 4979 (class 0 OID 0)
-- Dependencies: 223
-- Name: company_settings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.company_settings_id_seq OWNED BY public.company_settings.id;


--
-- TOC entry 222 (class 1259 OID 16812)
-- Name: role_permissions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.role_permissions (
    id integer NOT NULL,
    role_id integer NOT NULL,
    category character varying(20) NOT NULL,
    form_name character varying(100) NOT NULL,
    sub_permission character varying(20),
    allowed boolean DEFAULT false,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT role_permissions_category_check CHECK (((category)::text = ANY ((ARRAY['Setup'::character varying, 'Entry'::character varying, 'Reports'::character varying])::text[]))),
    CONSTRAINT role_permissions_sub_permission_check CHECK (((sub_permission)::text = ANY ((ARRAY['add_record'::character varying, 'edit'::character varying, 'delete'::character varying])::text[])))
);


ALTER TABLE public.role_permissions OWNER TO postgres;

--
-- TOC entry 221 (class 1259 OID 16811)
-- Name: role_permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.role_permissions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.role_permissions_id_seq OWNER TO postgres;

--
-- TOC entry 4980 (class 0 OID 0)
-- Dependencies: 221
-- Name: role_permissions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.role_permissions_id_seq OWNED BY public.role_permissions.id;


--
-- TOC entry 218 (class 1259 OID 16785)
-- Name: roles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.roles (
    id integer NOT NULL,
    role_name character varying(50) NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.roles OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 16784)
-- Name: roles_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.roles_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.roles_id_seq OWNER TO postgres;

--
-- TOC entry 4981 (class 0 OID 0)
-- Dependencies: 217
-- Name: roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.roles_id_seq OWNED BY public.roles.id;


--
-- TOC entry 230 (class 1259 OID 33205)
-- Name: sub_accounts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sub_accounts (
    id integer NOT NULL,
    account_head_id integer NOT NULL,
    name character varying(100) NOT NULL
);


ALTER TABLE public.sub_accounts OWNER TO postgres;

--
-- TOC entry 229 (class 1259 OID 33204)
-- Name: sub_accounts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.sub_accounts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sub_accounts_id_seq OWNER TO postgres;

--
-- TOC entry 4982 (class 0 OID 0)
-- Dependencies: 229
-- Name: sub_accounts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.sub_accounts_id_seq OWNED BY public.sub_accounts.id;


--
-- TOC entry 220 (class 1259 OID 16796)
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id integer NOT NULL,
    username character varying(255) NOT NULL,
    password character varying(255) NOT NULL,
    role character varying(50) DEFAULT 'user'::character varying,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    privileges text,
    blocked boolean DEFAULT false,
    last_activity timestamp without time zone,
    email character varying(255) NOT NULL,
    is_active boolean DEFAULT true,
    first_name character varying(255) DEFAULT NULL::character varying,
    last_name character varying(255) DEFAULT NULL::character varying,
    role_id integer,
    is_admin boolean DEFAULT false
);


ALTER TABLE public.users OWNER TO postgres;

--
-- TOC entry 219 (class 1259 OID 16795)
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
-- TOC entry 4983 (class 0 OID 0)
-- Dependencies: 219
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- TOC entry 4766 (class 2604 OID 33220)
-- Name: accounts id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.accounts ALTER COLUMN id SET DEFAULT nextval('public.accounts_id_seq'::regclass);


--
-- TOC entry 4764 (class 2604 OID 33201)
-- Name: accounts_head id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.accounts_head ALTER COLUMN id SET DEFAULT nextval('public.accounts_head_id_seq'::regclass);


--
-- TOC entry 4768 (class 2604 OID 33239)
-- Name: area id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.area ALTER COLUMN id SET DEFAULT nextval('public.area_id_seq'::regclass);


--
-- TOC entry 4761 (class 2604 OID 33189)
-- Name: bank_info id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.bank_info ALTER COLUMN id SET DEFAULT nextval('public.bank_info_id_seq'::regclass);


--
-- TOC entry 4767 (class 2604 OID 33232)
-- Name: city id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.city ALTER COLUMN id SET DEFAULT nextval('public.city_id_seq'::regclass);


--
-- TOC entry 4758 (class 2604 OID 33176)
-- Name: company_settings id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.company_settings ALTER COLUMN id SET DEFAULT nextval('public.company_settings_id_seq'::regclass);


--
-- TOC entry 4754 (class 2604 OID 16815)
-- Name: role_permissions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_permissions ALTER COLUMN id SET DEFAULT nextval('public.role_permissions_id_seq'::regclass);


--
-- TOC entry 4743 (class 2604 OID 16788)
-- Name: roles id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.roles ALTER COLUMN id SET DEFAULT nextval('public.roles_id_seq'::regclass);


--
-- TOC entry 4765 (class 2604 OID 33208)
-- Name: sub_accounts id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sub_accounts ALTER COLUMN id SET DEFAULT nextval('public.sub_accounts_id_seq'::regclass);


--
-- TOC entry 4746 (class 2604 OID 16799)
-- Name: users id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- TOC entry 4964 (class 0 OID 33217)
-- Dependencies: 232
-- Data for Name: accounts; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.accounts (id, sub_account_id, name) FROM stdin;
1	1	Bank
2	1	Cheque
3	1	Cash
4	1	Online
5	30	Zahid Ghori
\.


--
-- TOC entry 4960 (class 0 OID 33198)
-- Dependencies: 228
-- Data for Name: accounts_head; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.accounts_head (id, name) FROM stdin;
1	Assets
2	Liabilities
3	Equity
4	Income
5	Expenses
\.


--
-- TOC entry 4968 (class 0 OID 33236)
-- Dependencies: 236
-- Data for Name: area; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.area (id, city_id, name) FROM stdin;
\.


--
-- TOC entry 4958 (class 0 OID 33186)
-- Dependencies: 226
-- Data for Name: bank_info; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.bank_info (id, bank_name, branch_name, branch_code, account_number, account_title, iban, swift_code, bank_address, contact_person, contact_number, email, created_at, updated_at) FROM stdin;
1	Meezan	Saadi Town	1605	4625-986533245-56	Abc Co			Main Road Saadi Town	Asif	03359876451		2025-05-07 14:51:00	2025-05-07 14:51:00
\.


--
-- TOC entry 4966 (class 0 OID 33229)
-- Dependencies: 234
-- Data for Name: city; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.city (id, name) FROM stdin;
\.


--
-- TOC entry 4956 (class 0 OID 33173)
-- Dependencies: 224
-- Data for Name: company_settings; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.company_settings (id, company_name, company_logo, address_line1, address_line2, city, state, postal_code, country, phone_1, phone_2, email, website, tax_id, registration_number, created_at, updated_at) FROM stdin;
1	InnovaTech	../../../views/setup/companysetup/uploads/company_logo/company_logo_6837603e12ded.png	Saadi Town Block 1		Karachi	Sindh	74600	Pakistan	03468918711						2025-05-29 00:13:02.077856	2025-05-29 00:28:01.526055
\.


--
-- TOC entry 4954 (class 0 OID 16812)
-- Dependencies: 222
-- Data for Name: role_permissions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.role_permissions (id, role_id, category, form_name, sub_permission, allowed, created_at, updated_at) FROM stdin;
\.


--
-- TOC entry 4950 (class 0 OID 16785)
-- Dependencies: 218
-- Data for Name: roles; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.roles (id, role_name, created_at, updated_at) FROM stdin;
1	Administrator	2025-04-29 15:05:01	2025-04-29 15:05:01
2	Manager	2025-04-29 15:05:01	2025-04-29 15:05:01
3	User	2025-04-29 15:05:01	2025-04-29 15:05:01
\.


--
-- TOC entry 4962 (class 0 OID 33205)
-- Dependencies: 230
-- Data for Name: sub_accounts; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sub_accounts (id, account_head_id, name) FROM stdin;
1	1	Narration
2	1	Prepaid Rent
3	1	Land & Building
4	1	Office Equipment
5	1	Furniture & Fixtures
6	1	Vehicles
7	1	Goodwill
8	1	Software License
9	1	Patents
10	2	Short-term Loan
11	2	Bank Loan
12	2	Lease Payable
13	2	Bonds Payable
14	3	Capital
15	3	Drawings
16	3	Retained Earnings
17	3	Reserves
18	4	Sales Revenue
19	4	Service Income
20	4	Consultancy Income
21	4	Interest Income
22	4	Rent Received
23	4	Gain on Sale of Asset
24	5	Utility
25	5	Rent Expense
26	5	Interest Expense
27	5	Loss on Asset Disposal
28	5	Depreciation
29	5	Amortization
30	1	Other Receivable
31	2	Other Payable
\.


--
-- TOC entry 4952 (class 0 OID 16796)
-- Dependencies: 220
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, username, password, role, created_at, privileges, blocked, last_activity, email, is_active, first_name, last_name, role_id, is_admin) FROM stdin;
1	admin	$2y$10$/4HECJxd.TXoqtJIxrXjEeecpb8Wxy15yi1uY96uS4ua9yi29ymNe	admin	2025-03-08 17:09:31	\N	f	2025-05-13 08:38:35	admin@gmail.com	t	Admin	\N	1	t
\.


--
-- TOC entry 4984 (class 0 OID 0)
-- Dependencies: 227
-- Name: accounts_head_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.accounts_head_id_seq', 5, true);


--
-- TOC entry 4985 (class 0 OID 0)
-- Dependencies: 231
-- Name: accounts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.accounts_id_seq', 5, true);


--
-- TOC entry 4986 (class 0 OID 0)
-- Dependencies: 235
-- Name: area_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.area_id_seq', 1, false);


--
-- TOC entry 4987 (class 0 OID 0)
-- Dependencies: 225
-- Name: bank_info_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.bank_info_id_seq', 1, true);


--
-- TOC entry 4988 (class 0 OID 0)
-- Dependencies: 233
-- Name: city_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.city_id_seq', 1, false);


--
-- TOC entry 4989 (class 0 OID 0)
-- Dependencies: 223
-- Name: company_settings_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.company_settings_id_seq', 1, true);


--
-- TOC entry 4990 (class 0 OID 0)
-- Dependencies: 221
-- Name: role_permissions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.role_permissions_id_seq', 823, true);


--
-- TOC entry 4991 (class 0 OID 0)
-- Dependencies: 217
-- Name: roles_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.roles_id_seq', 3, true);


--
-- TOC entry 4992 (class 0 OID 0)
-- Dependencies: 229
-- Name: sub_accounts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.sub_accounts_id_seq', 31, true);


--
-- TOC entry 4993 (class 0 OID 0)
-- Dependencies: 219
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_id_seq', 1, false);


--
-- TOC entry 4788 (class 2606 OID 33203)
-- Name: accounts_head accounts_head_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.accounts_head
    ADD CONSTRAINT accounts_head_pkey PRIMARY KEY (id);


--
-- TOC entry 4792 (class 2606 OID 33222)
-- Name: accounts accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);


--
-- TOC entry 4796 (class 2606 OID 33241)
-- Name: area area_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.area
    ADD CONSTRAINT area_pkey PRIMARY KEY (id);


--
-- TOC entry 4786 (class 2606 OID 33195)
-- Name: bank_info bank_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.bank_info
    ADD CONSTRAINT bank_info_pkey PRIMARY KEY (id);


--
-- TOC entry 4794 (class 2606 OID 33234)
-- Name: city city_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.city
    ADD CONSTRAINT city_pkey PRIMARY KEY (id);


--
-- TOC entry 4784 (class 2606 OID 33182)
-- Name: company_settings company_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.company_settings
    ADD CONSTRAINT company_settings_pkey PRIMARY KEY (id);


--
-- TOC entry 4781 (class 2606 OID 16822)
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (id);


--
-- TOC entry 4772 (class 2606 OID 16792)
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- TOC entry 4774 (class 2606 OID 16794)
-- Name: roles roles_role_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_role_name_key UNIQUE (role_name);


--
-- TOC entry 4790 (class 2606 OID 33210)
-- Name: sub_accounts sub_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sub_accounts
    ADD CONSTRAINT sub_accounts_pkey PRIMARY KEY (id);


--
-- TOC entry 4776 (class 2606 OID 16810)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- TOC entry 4777 (class 1259 OID 16826)
-- Name: idx_allowed; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_allowed ON public.role_permissions USING btree (allowed);


--
-- TOC entry 4778 (class 1259 OID 16825)
-- Name: idx_category; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_category ON public.role_permissions USING btree (category);


--
-- TOC entry 4779 (class 1259 OID 16824)
-- Name: idx_role_form; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_role_form ON public.role_permissions USING btree (role_id, category, form_name);


--
-- TOC entry 4782 (class 1259 OID 16823)
-- Name: unique_permission; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX unique_permission ON public.role_permissions USING btree (role_id, category, form_name, sub_permission);


--
-- TOC entry 4803 (class 2620 OID 33196)
-- Name: bank_info set_updated_at; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER set_updated_at BEFORE UPDATE ON public.bank_info FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- TOC entry 4801 (class 2620 OID 16833)
-- Name: role_permissions trg_update_updated_at; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_update_updated_at BEFORE UPDATE ON public.role_permissions FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- TOC entry 4802 (class 2620 OID 33184)
-- Name: company_settings update_company_settings_timestamp; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_company_settings_timestamp BEFORE UPDATE ON public.company_settings FOR EACH ROW EXECUTE FUNCTION public.update_timestamp();


--
-- TOC entry 4799 (class 2606 OID 33223)
-- Name: accounts accounts_sub_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_sub_account_id_fkey FOREIGN KEY (sub_account_id) REFERENCES public.sub_accounts(id);


--
-- TOC entry 4800 (class 2606 OID 33242)
-- Name: area area_city_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.area
    ADD CONSTRAINT area_city_id_fkey FOREIGN KEY (city_id) REFERENCES public.city(id) ON DELETE CASCADE;


--
-- TOC entry 4797 (class 2606 OID 16827)
-- Name: role_permissions fk_role_permissions_role_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT fk_role_permissions_role_id FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;


--
-- TOC entry 4798 (class 2606 OID 33211)
-- Name: sub_accounts sub_accounts_account_head_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sub_accounts
    ADD CONSTRAINT sub_accounts_account_head_id_fkey FOREIGN KEY (account_head_id) REFERENCES public.accounts_head(id);


-- Completed on 2025-05-29 02:19:11

--
-- PostgreSQL database dump complete
--

