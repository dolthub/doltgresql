-- Downloaded from: https://github.com/DRON12261/EduVault/blob/d95f919c69452719e016c3afc53eb24280c11453/backup.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 17.4
-- Dumped by pg_dump version 17.4

-- Started on 2025-05-04 21:20:55

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
-- TOC entry 4 (class 2615 OID 2200)
-- Name: public; Type: SCHEMA; Schema: -; Owner: pg_database_owner
--

CREATE SCHEMA public;


ALTER SCHEMA public OWNER TO pg_database_owner;

--
-- TOC entry 4900 (class 0 OID 0)
-- Dependencies: 4
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: pg_database_owner
--

COMMENT ON SCHEMA public IS 'standard public schema';


--
-- TOC entry 237 (class 1255 OID 16628)
-- Name: Login(character varying, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public."Login"(login_to_check character varying, password_to_check character varying) RETURNS boolean
    LANGUAGE plpgsql
    AS $$DECLARE
	permission boolean;
BEGIN
	SELECT
		EXISTS 
		(
			SELECT * FROM "Users" 
			WHERE "Users".login = login_to_check
			AND "Users".password = password_to_check
		)
	INTO permission;
	
	RETURN permission;
END;$$;


ALTER FUNCTION public."Login"(login_to_check character varying, password_to_check character varying) OWNER TO postgres;

--
-- TOC entry 238 (class 1255 OID 16636)
-- Name: test_func(character varying, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.test_func(character varying, character varying) RETURNS integer
    LANGUAGE plpgsql
    AS $$DECLARE
	kek integer;
BEGIN
	SELECT user_id from "Users" into kek;
	RETURN kek;
END;$$;


ALTER FUNCTION public.test_func(character varying, character varying) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 224 (class 1259 OID 16459)
-- Name: AccessRights; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."AccessRights" (
    accessright_id bigint NOT NULL,
    user_id bigint,
    role_id bigint,
    accessrighttype_id bigint NOT NULL,
    record_id bigint NOT NULL
);


ALTER TABLE public."AccessRights" OWNER TO postgres;

--
-- TOC entry 230 (class 1259 OID 16509)
-- Name: AccessRightsTypes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."AccessRightsTypes" (
    accessrighttype_id bigint NOT NULL,
    artname character varying NOT NULL
);


ALTER TABLE public."AccessRightsTypes" OWNER TO postgres;

--
-- TOC entry 229 (class 1259 OID 16508)
-- Name: AccessRightsTypes_art_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public."AccessRightsTypes_art_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public."AccessRightsTypes_art_id_seq" OWNER TO postgres;

--
-- TOC entry 4901 (class 0 OID 0)
-- Dependencies: 229
-- Name: AccessRightsTypes_art_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public."AccessRightsTypes_art_id_seq" OWNED BY public."AccessRightsTypes".accessrighttype_id;


--
-- TOC entry 223 (class 1259 OID 16458)
-- Name: AccessRights_accessright_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public."AccessRights_accessright_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public."AccessRights_accessright_id_seq" OWNER TO postgres;

--
-- TOC entry 4902 (class 0 OID 0)
-- Dependencies: 223
-- Name: AccessRights_accessright_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public."AccessRights_accessright_id_seq" OWNED BY public."AccessRights".accessright_id;


--
-- TOC entry 228 (class 1259 OID 16475)
-- Name: Fields; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."Fields" (
    field_id bigint NOT NULL,
    name character varying NOT NULL,
    record_id bigint NOT NULL,
    value character varying,
    filetypefield_id bigint
);


ALTER TABLE public."Fields" OWNER TO postgres;

--
-- TOC entry 226 (class 1259 OID 16466)
-- Name: FileTypes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."FileTypes" (
    filetype_id bigint NOT NULL,
    typename character varying NOT NULL
);


ALTER TABLE public."FileTypes" OWNER TO postgres;

--
-- TOC entry 234 (class 1259 OID 16540)
-- Name: FileTypesFields; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."FileTypesFields" (
    filetypefield_id bigint NOT NULL,
    filetype_id bigint NOT NULL,
    name character varying NOT NULL,
    isrequired boolean NOT NULL,
    prefilling boolean NOT NULL
);


ALTER TABLE public."FileTypesFields" OWNER TO postgres;

--
-- TOC entry 233 (class 1259 OID 16539)
-- Name: FileTypesFields_filetypefield_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public."FileTypesFields_filetypefield_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public."FileTypesFields_filetypefield_id_seq" OWNER TO postgres;

--
-- TOC entry 4903 (class 0 OID 0)
-- Dependencies: 233
-- Name: FileTypesFields_filetypefield_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public."FileTypesFields_filetypefield_id_seq" OWNED BY public."FileTypesFields".filetypefield_id;


--
-- TOC entry 225 (class 1259 OID 16465)
-- Name: FileTypes_filetype_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public."FileTypes_filetype_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public."FileTypes_filetype_id_seq" OWNER TO postgres;

--
-- TOC entry 4904 (class 0 OID 0)
-- Dependencies: 225
-- Name: FileTypes_filetype_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public."FileTypes_filetype_id_seq" OWNED BY public."FileTypes".filetype_id;


--
-- TOC entry 218 (class 1259 OID 16431)
-- Name: Records; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."Records" (
    record_id bigint NOT NULL,
    filetype_id bigint NOT NULL,
    name character varying,
    filepath character varying,
    author character varying NOT NULL
);


ALTER TABLE public."Records" OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 16430)
-- Name: Metadata_metadata_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public."Metadata_metadata_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public."Metadata_metadata_id_seq" OWNER TO postgres;

--
-- TOC entry 4905 (class 0 OID 0)
-- Dependencies: 217
-- Name: Metadata_metadata_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public."Metadata_metadata_id_seq" OWNED BY public."Records".record_id;


--
-- TOC entry 232 (class 1259 OID 16523)
-- Name: RecordsRelations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."RecordsRelations" (
    relation_id bigint NOT NULL,
    sourcerecord bigint NOT NULL,
    targetrecord bigint NOT NULL
);


ALTER TABLE public."RecordsRelations" OWNER TO postgres;

--
-- TOC entry 231 (class 1259 OID 16522)
-- Name: RecordsRelations_relation_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public."RecordsRelations_relation_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public."RecordsRelations_relation_id_seq" OWNER TO postgres;

--
-- TOC entry 4906 (class 0 OID 0)
-- Dependencies: 231
-- Name: RecordsRelations_relation_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public."RecordsRelations_relation_id_seq" OWNED BY public."RecordsRelations".relation_id;


--
-- TOC entry 222 (class 1259 OID 16449)
-- Name: Roles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."Roles" (
    role_id bigint NOT NULL,
    rolename character varying NOT NULL
);


ALTER TABLE public."Roles" OWNER TO postgres;

--
-- TOC entry 221 (class 1259 OID 16448)
-- Name: Roles_role_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public."Roles_role_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public."Roles_role_id_seq" OWNER TO postgres;

--
-- TOC entry 4907 (class 0 OID 0)
-- Dependencies: 221
-- Name: Roles_role_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public."Roles_role_id_seq" OWNED BY public."Roles".role_id;


--
-- TOC entry 227 (class 1259 OID 16474)
-- Name: UserMetadata_usermetadata_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public."UserMetadata_usermetadata_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public."UserMetadata_usermetadata_id_seq" OWNER TO postgres;

--
-- TOC entry 4908 (class 0 OID 0)
-- Dependencies: 227
-- Name: UserMetadata_usermetadata_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public."UserMetadata_usermetadata_id_seq" OWNED BY public."Fields".field_id;


--
-- TOC entry 220 (class 1259 OID 16440)
-- Name: Users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."Users" (
    user_id bigint NOT NULL,
    login character varying NOT NULL,
    password character varying NOT NULL,
    name character varying,
    usertype bigint
);


ALTER TABLE public."Users" OWNER TO postgres;

--
-- TOC entry 235 (class 1259 OID 16571)
-- Name: UsersRoles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."UsersRoles" (
    usersroles_id bigint NOT NULL,
    user_id bigint NOT NULL,
    role_id bigint NOT NULL
);


ALTER TABLE public."UsersRoles" OWNER TO postgres;

--
-- TOC entry 236 (class 1259 OID 16574)
-- Name: UsersRoles_usersroles_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public."UsersRoles_usersroles_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public."UsersRoles_usersroles_id_seq" OWNER TO postgres;

--
-- TOC entry 4909 (class 0 OID 0)
-- Dependencies: 236
-- Name: UsersRoles_usersroles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public."UsersRoles_usersroles_id_seq" OWNED BY public."UsersRoles".usersroles_id;


--
-- TOC entry 219 (class 1259 OID 16439)
-- Name: Users_user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public."Users_user_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public."Users_user_id_seq" OWNER TO postgres;

--
-- TOC entry 4910 (class 0 OID 0)
-- Dependencies: 219
-- Name: Users_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public."Users_user_id_seq" OWNED BY public."Users".user_id;


--
-- TOC entry 4691 (class 2604 OID 16562)
-- Name: AccessRights accessright_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."AccessRights" ALTER COLUMN accessright_id SET DEFAULT nextval('public."AccessRights_accessright_id_seq"'::regclass);


--
-- TOC entry 4694 (class 2604 OID 16563)
-- Name: AccessRightsTypes accessrighttype_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."AccessRightsTypes" ALTER COLUMN accessrighttype_id SET DEFAULT nextval('public."AccessRightsTypes_art_id_seq"'::regclass);


--
-- TOC entry 4693 (class 2604 OID 16569)
-- Name: Fields field_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."Fields" ALTER COLUMN field_id SET DEFAULT nextval('public."UserMetadata_usermetadata_id_seq"'::regclass);


--
-- TOC entry 4692 (class 2604 OID 16564)
-- Name: FileTypes filetype_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."FileTypes" ALTER COLUMN filetype_id SET DEFAULT nextval('public."FileTypes_filetype_id_seq"'::regclass);


--
-- TOC entry 4696 (class 2604 OID 16565)
-- Name: FileTypesFields filetypefield_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."FileTypesFields" ALTER COLUMN filetypefield_id SET DEFAULT nextval('public."FileTypesFields_filetypefield_id_seq"'::regclass);


--
-- TOC entry 4688 (class 2604 OID 16566)
-- Name: Records record_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."Records" ALTER COLUMN record_id SET DEFAULT nextval('public."Metadata_metadata_id_seq"'::regclass);


--
-- TOC entry 4695 (class 2604 OID 16567)
-- Name: RecordsRelations relation_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."RecordsRelations" ALTER COLUMN relation_id SET DEFAULT nextval('public."RecordsRelations_relation_id_seq"'::regclass);


--
-- TOC entry 4690 (class 2604 OID 16568)
-- Name: Roles role_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."Roles" ALTER COLUMN role_id SET DEFAULT nextval('public."Roles_role_id_seq"'::regclass);


--
-- TOC entry 4689 (class 2604 OID 16570)
-- Name: Users user_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."Users" ALTER COLUMN user_id SET DEFAULT nextval('public."Users_user_id_seq"'::regclass);


--
-- TOC entry 4697 (class 2604 OID 16575)
-- Name: UsersRoles usersroles_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."UsersRoles" ALTER COLUMN usersroles_id SET DEFAULT nextval('public."UsersRoles_usersroles_id_seq"'::regclass);


--
-- TOC entry 4882 (class 0 OID 16459)
-- Dependencies: 224
-- Data for Name: AccessRights; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public."AccessRights" (accessright_id, user_id, role_id, accessrighttype_id, record_id) FROM stdin;
\.


--
-- TOC entry 4888 (class 0 OID 16509)
-- Dependencies: 230
-- Data for Name: AccessRightsTypes; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public."AccessRightsTypes" (accessrighttype_id, artname) FROM stdin;
\.


--
-- TOC entry 4886 (class 0 OID 16475)
-- Dependencies: 228
-- Data for Name: Fields; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public."Fields" (field_id, name, record_id, value, filetypefield_id) FROM stdin;
\.


--
-- TOC entry 4884 (class 0 OID 16466)
-- Dependencies: 226
-- Data for Name: FileTypes; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public."FileTypes" (filetype_id, typename) FROM stdin;
\.


--
-- TOC entry 4892 (class 0 OID 16540)
-- Dependencies: 234
-- Data for Name: FileTypesFields; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public."FileTypesFields" (filetypefield_id, filetype_id, name, isrequired, prefilling) FROM stdin;
\.


--
-- TOC entry 4876 (class 0 OID 16431)
-- Dependencies: 218
-- Data for Name: Records; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public."Records" (record_id, filetype_id, name, filepath, author) FROM stdin;
\.


--
-- TOC entry 4890 (class 0 OID 16523)
-- Dependencies: 232
-- Data for Name: RecordsRelations; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public."RecordsRelations" (relation_id, sourcerecord, targetrecord) FROM stdin;
\.


--
-- TOC entry 4880 (class 0 OID 16449)
-- Dependencies: 222
-- Data for Name: Roles; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public."Roles" (role_id, rolename) FROM stdin;
\.


--
-- TOC entry 4878 (class 0 OID 16440)
-- Dependencies: 220
-- Data for Name: Users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public."Users" (user_id, login, password, name, usertype) FROM stdin;
1	test	test	TestUser	1
\.


--
-- TOC entry 4893 (class 0 OID 16571)
-- Dependencies: 235
-- Data for Name: UsersRoles; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public."UsersRoles" (usersroles_id, user_id, role_id) FROM stdin;
\.


--
-- TOC entry 4911 (class 0 OID 0)
-- Dependencies: 229
-- Name: AccessRightsTypes_art_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public."AccessRightsTypes_art_id_seq"', 1, false);


--
-- TOC entry 4912 (class 0 OID 0)
-- Dependencies: 223
-- Name: AccessRights_accessright_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public."AccessRights_accessright_id_seq"', 1, false);


--
-- TOC entry 4913 (class 0 OID 0)
-- Dependencies: 233
-- Name: FileTypesFields_filetypefield_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public."FileTypesFields_filetypefield_id_seq"', 1, false);


--
-- TOC entry 4914 (class 0 OID 0)
-- Dependencies: 225
-- Name: FileTypes_filetype_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public."FileTypes_filetype_id_seq"', 1, false);


--
-- TOC entry 4915 (class 0 OID 0)
-- Dependencies: 217
-- Name: Metadata_metadata_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public."Metadata_metadata_id_seq"', 1, false);


--
-- TOC entry 4916 (class 0 OID 0)
-- Dependencies: 231
-- Name: RecordsRelations_relation_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public."RecordsRelations_relation_id_seq"', 1, false);


--
-- TOC entry 4917 (class 0 OID 0)
-- Dependencies: 221
-- Name: Roles_role_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public."Roles_role_id_seq"', 1, false);


--
-- TOC entry 4918 (class 0 OID 0)
-- Dependencies: 227
-- Name: UserMetadata_usermetadata_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public."UserMetadata_usermetadata_id_seq"', 1, false);


--
-- TOC entry 4919 (class 0 OID 0)
-- Dependencies: 236
-- Name: UsersRoles_usersroles_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public."UsersRoles_usersroles_id_seq"', 1, false);


--
-- TOC entry 4920 (class 0 OID 0)
-- Dependencies: 219
-- Name: Users_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public."Users_user_id_seq"', 1, true);


--
-- TOC entry 4711 (class 2606 OID 16516)
-- Name: AccessRightsTypes AccessRightsTypes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."AccessRightsTypes"
    ADD CONSTRAINT "AccessRightsTypes_pkey" PRIMARY KEY (accessrighttype_id);


--
-- TOC entry 4705 (class 2606 OID 16464)
-- Name: AccessRights AccessRights_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."AccessRights"
    ADD CONSTRAINT "AccessRights_pkey" PRIMARY KEY (accessright_id);


--
-- TOC entry 4715 (class 2606 OID 16547)
-- Name: FileTypesFields FileTypesFields_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."FileTypesFields"
    ADD CONSTRAINT "FileTypesFields_pkey" PRIMARY KEY (filetypefield_id);


--
-- TOC entry 4707 (class 2606 OID 16473)
-- Name: FileTypes FileTypes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."FileTypes"
    ADD CONSTRAINT "FileTypes_pkey" PRIMARY KEY (filetype_id);


--
-- TOC entry 4699 (class 2606 OID 16438)
-- Name: Records Metadata_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."Records"
    ADD CONSTRAINT "Metadata_pkey" PRIMARY KEY (record_id);


--
-- TOC entry 4713 (class 2606 OID 16528)
-- Name: RecordsRelations RecordsRelations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."RecordsRelations"
    ADD CONSTRAINT "RecordsRelations_pkey" PRIMARY KEY (relation_id);


--
-- TOC entry 4703 (class 2606 OID 16456)
-- Name: Roles Roles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."Roles"
    ADD CONSTRAINT "Roles_pkey" PRIMARY KEY (role_id);


--
-- TOC entry 4709 (class 2606 OID 16482)
-- Name: Fields UserMetadata_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."Fields"
    ADD CONSTRAINT "UserMetadata_pkey" PRIMARY KEY (field_id);


--
-- TOC entry 4717 (class 2606 OID 16580)
-- Name: UsersRoles UsersRoles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."UsersRoles"
    ADD CONSTRAINT "UsersRoles_pkey" PRIMARY KEY (usersroles_id);


--
-- TOC entry 4701 (class 2606 OID 16447)
-- Name: Users Users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."Users"
    ADD CONSTRAINT "Users_pkey" PRIMARY KEY (user_id);


--
-- TOC entry 4719 (class 2606 OID 16517)
-- Name: AccessRights accessrights_ART; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."AccessRights"
    ADD CONSTRAINT "accessrights_ART" FOREIGN KEY (accessrighttype_id) REFERENCES public."AccessRightsTypes"(accessrighttype_id) NOT VALID;


--
-- TOC entry 4720 (class 2606 OID 16498)
-- Name: AccessRights accessrights_records; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."AccessRights"
    ADD CONSTRAINT accessrights_records FOREIGN KEY (record_id) REFERENCES public."Records"(record_id) NOT VALID;


--
-- TOC entry 4721 (class 2606 OID 16493)
-- Name: AccessRights accessrights_roles; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."AccessRights"
    ADD CONSTRAINT accessrights_roles FOREIGN KEY (role_id) REFERENCES public."Roles"(role_id) NOT VALID;


--
-- TOC entry 4722 (class 2606 OID 16488)
-- Name: AccessRights accessrights_users; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."AccessRights"
    ADD CONSTRAINT accessrights_users FOREIGN KEY (user_id) REFERENCES public."Users"(user_id) NOT VALID;


--
-- TOC entry 4723 (class 2606 OID 16591)
-- Name: Fields files_filetypesfields; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."Fields"
    ADD CONSTRAINT files_filetypesfields FOREIGN KEY (filetypefield_id) REFERENCES public."FileTypesFields"(filetypefield_id) NOT VALID;


--
-- TOC entry 4724 (class 2606 OID 16503)
-- Name: Fields files_records; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."Fields"
    ADD CONSTRAINT files_records FOREIGN KEY (record_id) REFERENCES public."Records"(record_id) NOT VALID;


--
-- TOC entry 4727 (class 2606 OID 16548)
-- Name: FileTypesFields filetypesfields_filetypes; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."FileTypesFields"
    ADD CONSTRAINT filetypesfields_filetypes FOREIGN KEY (filetype_id) REFERENCES public."FileTypes"(filetype_id);


--
-- TOC entry 4718 (class 2606 OID 16483)
-- Name: Records metadata_filetypes; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."Records"
    ADD CONSTRAINT metadata_filetypes FOREIGN KEY (filetype_id) REFERENCES public."FileTypes"(filetype_id) NOT VALID;


--
-- TOC entry 4725 (class 2606 OID 16529)
-- Name: RecordsRelations sourcerecord_records; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."RecordsRelations"
    ADD CONSTRAINT sourcerecord_records FOREIGN KEY (sourcerecord) REFERENCES public."Records"(record_id);


--
-- TOC entry 4726 (class 2606 OID 16534)
-- Name: RecordsRelations targetrecord_records; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."RecordsRelations"
    ADD CONSTRAINT targetrecord_records FOREIGN KEY (targetrecord) REFERENCES public."Records"(record_id);


--
-- TOC entry 4728 (class 2606 OID 16586)
-- Name: UsersRoles usersroles_roles; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."UsersRoles"
    ADD CONSTRAINT usersroles_roles FOREIGN KEY (role_id) REFERENCES public."Roles"(role_id) NOT VALID;


--
-- TOC entry 4729 (class 2606 OID 16581)
-- Name: UsersRoles usersroles_users; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."UsersRoles"
    ADD CONSTRAINT usersroles_users FOREIGN KEY (user_id) REFERENCES public."Users"(user_id) NOT VALID;


-- Completed on 2025-05-04 21:20:56

--
-- PostgreSQL database dump complete
--

