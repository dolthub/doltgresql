-- Downloaded from: https://github.com/kapil23jani/hospitease_backend/blob/e83fe7b286e5da1c3160d969afd2a39cf42086a1/mahool.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 14.15 (Homebrew)
-- Dumped by pg_dump version 15.10 (Homebrew)

-- Started on 2025-06-15 23:56:50 IST

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
-- TOC entry 4 (class 2615 OID 2200)
-- Name: public; Type: SCHEMA; Schema: -; Owner: kapiljani
--

-- *not* creating schema, since initdb creates it


ALTER SCHEMA public OWNER TO kapiljani;

--
-- TOC entry 249 (class 1255 OID 16526)
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_updated_at_column() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 220 (class 1259 OID 16506)
-- Name: appointments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.appointments (
    id integer NOT NULL,
    patient_id integer NOT NULL,
    doctor_id integer NOT NULL,
    appointment_datetime character varying,
    problem character varying(255),
    appointment_type character varying(255),
    reason character varying(255),
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    blood_pressure character varying(255),
    pulse_rate character varying(255),
    temperature character varying(255),
    spo2 character varying(255),
    weight character varying(255),
    additional_notes text,
    advice text,
    follow_up_date character varying,
    follow_up_notes text,
    appointment_date character varying(20),
    appointment_time character varying(20),
    hospital_id integer,
    status character varying(255),
    appointment_unique_id character varying(32)
);


ALTER TABLE public.appointments OWNER TO postgres;

--
-- TOC entry 219 (class 1259 OID 16505)
-- Name: appointments_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.appointments_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.appointments_id_seq OWNER TO postgres;

--
-- TOC entry 3879 (class 0 OID 0)
-- Dependencies: 219
-- Name: appointments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.appointments_id_seq OWNED BY public.appointments.id;


--
-- TOC entry 242 (class 1259 OID 16784)
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.audit_logs (
    id integer NOT NULL,
    table_name character varying NOT NULL,
    record_id integer NOT NULL,
    action character varying NOT NULL,
    "timestamp" timestamp without time zone NOT NULL,
    user_id integer,
    old_data text,
    new_data text
);


ALTER TABLE public.audit_logs OWNER TO postgres;

--
-- TOC entry 241 (class 1259 OID 16783)
-- Name: audit_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.audit_logs_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.audit_logs_id_seq OWNER TO postgres;

--
-- TOC entry 3882 (class 0 OID 0)
-- Dependencies: 241
-- Name: audit_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.audit_logs_id_seq OWNED BY public.audit_logs.id;


--
-- TOC entry 218 (class 1259 OID 16466)
-- Name: doctors; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.doctors (
    id integer NOT NULL,
    first_name character varying(255) NOT NULL,
    last_name character varying(255) NOT NULL,
    specialization character varying(255) NOT NULL,
    phone_number character varying(20) NOT NULL,
    email character varying(255) NOT NULL,
    experience integer NOT NULL,
    is_active boolean DEFAULT true,
    hospital_id integer NOT NULL,
    title character varying(100),
    gender character varying(50),
    date_of_birth date,
    blood_group character varying(10),
    mobile_number character varying(20),
    emergency_contact character varying(20),
    address text,
    city character varying(100),
    state character varying(100),
    country character varying(100),
    zipcode character varying(20),
    medical_licence_number character varying(100),
    licence_authority character varying(255),
    license_expiry_date date,
    user_id integer
);


ALTER TABLE public.doctors OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 16465)
-- Name: doctors_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.doctors_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.doctors_id_seq OWNER TO postgres;

--
-- TOC entry 3885 (class 0 OID 0)
-- Dependencies: 217
-- Name: doctors_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.doctors_id_seq OWNED BY public.doctors.id;


--
-- TOC entry 230 (class 1259 OID 16645)
-- Name: documents; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.documents (
    id integer NOT NULL,
    document_name character varying(255) NOT NULL,
    document_type character varying(50) NOT NULL,
    upload_date character varying(100) NOT NULL,
    size character varying(50) NOT NULL,
    status character varying(50) NOT NULL,
    documentable_id integer NOT NULL,
    documentable_type character varying(50) NOT NULL
);


ALTER TABLE public.documents OWNER TO postgres;

--
-- TOC entry 229 (class 1259 OID 16644)
-- Name: documents_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.documents_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.documents_id_seq OWNER TO postgres;

--
-- TOC entry 3888 (class 0 OID 0)
-- Dependencies: 229
-- Name: documents_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.documents_id_seq OWNED BY public.documents.id;


--
-- TOC entry 234 (class 1259 OID 16677)
-- Name: family_history; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.family_history (
    id integer NOT NULL,
    appointment_id integer NOT NULL,
    relationship_to_you character varying(255),
    additional_notes character varying(255)
);


ALTER TABLE public.family_history OWNER TO postgres;

--
-- TOC entry 233 (class 1259 OID 16676)
-- Name: family_history_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.family_history_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.family_history_id_seq OWNER TO postgres;

--
-- TOC entry 3891 (class 0 OID 0)
-- Dependencies: 233
-- Name: family_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.family_history_id_seq OWNED BY public.family_history.id;


--
-- TOC entry 232 (class 1259 OID 16663)
-- Name: health_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.health_info (
    id integer NOT NULL,
    appointment_id integer,
    known_allergies text,
    reaction_severity text,
    reaction_description text,
    dietary_habits text,
    physical_activity_level text,
    sleep_avg_hours integer,
    sleep_quality text,
    substance_use_smoking text,
    substance_use_alcohol text,
    stress_level text
);


ALTER TABLE public.health_info OWNER TO postgres;

--
-- TOC entry 231 (class 1259 OID 16662)
-- Name: health_info_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.health_info_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.health_info_id_seq OWNER TO postgres;

--
-- TOC entry 3894 (class 0 OID 0)
-- Dependencies: 231
-- Name: health_info_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.health_info_id_seq OWNED BY public.health_info.id;


--
-- TOC entry 248 (class 1259 OID 16844)
-- Name: hospital_payments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.hospital_payments (
    id integer NOT NULL,
    hospital_id integer NOT NULL,
    date date NOT NULL,
    amount numeric(12,2) NOT NULL,
    payment_method character varying(50) NOT NULL,
    reference character varying(100),
    status character varying(50) NOT NULL,
    paid boolean DEFAULT false,
    remarks text
);


ALTER TABLE public.hospital_payments OWNER TO postgres;

--
-- TOC entry 247 (class 1259 OID 16843)
-- Name: hospital_payments_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.hospital_payments_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.hospital_payments_id_seq OWNER TO postgres;

--
-- TOC entry 3897 (class 0 OID 0)
-- Dependencies: 247
-- Name: hospital_payments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.hospital_payments_id_seq OWNED BY public.hospital_payments.id;


--
-- TOC entry 246 (class 1259 OID 16827)
-- Name: hospital_permissions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.hospital_permissions (
    id integer NOT NULL,
    hospital_id integer,
    permission_id integer
);


ALTER TABLE public.hospital_permissions OWNER TO postgres;

--
-- TOC entry 245 (class 1259 OID 16826)
-- Name: hospital_permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.hospital_permissions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.hospital_permissions_id_seq OWNER TO postgres;

--
-- TOC entry 3900 (class 0 OID 0)
-- Dependencies: 245
-- Name: hospital_permissions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.hospital_permissions_id_seq OWNED BY public.hospital_permissions.id;


--
-- TOC entry 214 (class 1259 OID 16428)
-- Name: hospitals; Type: TABLE; Schema: public; Owner: hospitease_admin
--

CREATE TABLE public.hospitals (
    id integer NOT NULL,
    name character varying NOT NULL,
    address character varying,
    city character varying,
    state character varying,
    country character varying,
    phone_number character varying,
    email character varying,
    admin_id integer,
    registration_number character varying(100),
    type character varying(100),
    logo_url text,
    website character varying(255),
    owner_name character varying(255),
    admin_contact_number character varying(20),
    number_of_beds integer,
    departments text[],
    specialties text[],
    facilities text[],
    ambulance_services boolean DEFAULT false,
    opening_hours jsonb,
    license_number character varying(100),
    license_expiry_date date,
    is_accredited boolean DEFAULT false,
    external_id character varying(255),
    timezone character varying(50),
    is_active boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    zipcode character varying(255)
);


ALTER TABLE public.hospitals OWNER TO hospitease_admin;

--
-- TOC entry 213 (class 1259 OID 16427)
-- Name: hospitals_id_seq; Type: SEQUENCE; Schema: public; Owner: hospitease_admin
--

CREATE SEQUENCE public.hospitals_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.hospitals_id_seq OWNER TO hospitease_admin;

--
-- TOC entry 3903 (class 0 OID 0)
-- Dependencies: 213
-- Name: hospitals_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hospitease_admin
--

ALTER SEQUENCE public.hospitals_id_seq OWNED BY public.hospitals.id;


--
-- TOC entry 236 (class 1259 OID 16691)
-- Name: medical_histories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.medical_histories (
    id integer NOT NULL,
    condition character varying NOT NULL,
    diagnosis_date date NOT NULL,
    treatment text,
    doctor character varying,
    hospital character varying,
    status character varying,
    patient_id integer NOT NULL
);


ALTER TABLE public.medical_histories OWNER TO postgres;

--
-- TOC entry 235 (class 1259 OID 16690)
-- Name: medical_histories_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.medical_histories_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.medical_histories_id_seq OWNER TO postgres;

--
-- TOC entry 3905 (class 0 OID 0)
-- Dependencies: 235
-- Name: medical_histories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.medical_histories_id_seq OWNED BY public.medical_histories.id;


--
-- TOC entry 228 (class 1259 OID 16631)
-- Name: medicines; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.medicines (
    id integer NOT NULL,
    appointment_id integer NOT NULL,
    name character varying NOT NULL,
    dosage character varying NOT NULL,
    frequency character varying NOT NULL,
    duration character varying NOT NULL,
    start_date character varying NOT NULL,
    status character varying NOT NULL,
    time_interval character varying(255),
    route character varying(255),
    quantity character varying(255),
    instruction character varying(255)
);


ALTER TABLE public.medicines OWNER TO postgres;

--
-- TOC entry 227 (class 1259 OID 16630)
-- Name: medicines_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.medicines_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.medicines_id_seq OWNER TO postgres;

--
-- TOC entry 3908 (class 0 OID 0)
-- Dependencies: 227
-- Name: medicines_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.medicines_id_seq OWNED BY public.medicines.id;


--
-- TOC entry 216 (class 1259 OID 16442)
-- Name: patients; Type: TABLE; Schema: public; Owner: hospitease_admin
--

CREATE TABLE public.patients (
    id integer NOT NULL,
    first_name character varying NOT NULL,
    middle_name character varying,
    last_name character varying NOT NULL,
    date_of_birth character varying,
    gender character varying NOT NULL,
    phone_number character varying,
    landline character varying,
    address character varying,
    landmark character varying,
    city character varying,
    state character varying,
    country character varying,
    blood_group character varying,
    email character varying,
    occupation character varying,
    is_dialysis_patient boolean,
    hospital_id integer NOT NULL,
    marital_status character varying(255),
    zipcode character varying(20),
    patient_unique_id character varying(255) DEFAULT 'DUMMY_ID'::character varying NOT NULL
);


ALTER TABLE public.patients OWNER TO hospitease_admin;

--
-- TOC entry 215 (class 1259 OID 16441)
-- Name: patients_id_seq; Type: SEQUENCE; Schema: public; Owner: hospitease_admin
--

CREATE SEQUENCE public.patients_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.patients_id_seq OWNER TO hospitease_admin;

--
-- TOC entry 3911 (class 0 OID 0)
-- Dependencies: 215
-- Name: patients_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hospitease_admin
--

ALTER SEQUENCE public.patients_id_seq OWNED BY public.patients.id;


--
-- TOC entry 243 (class 1259 OID 16815)
-- Name: permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.permissions_id_seq OWNER TO postgres;

--
-- TOC entry 244 (class 1259 OID 16816)
-- Name: permissions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.permissions (
    id integer DEFAULT nextval('public.permissions_id_seq'::regclass) NOT NULL,
    name character varying(100) NOT NULL,
    description character varying(255),
    amount numeric(12,2)
);


ALTER TABLE public.permissions OWNER TO postgres;

--
-- TOC entry 239 (class 1259 OID 16747)
-- Name: receipt_line_items_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.receipt_line_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.receipt_line_items_id_seq OWNER TO postgres;

--
-- TOC entry 240 (class 1259 OID 16748)
-- Name: receipt_line_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.receipt_line_items (
    id integer DEFAULT nextval('public.receipt_line_items_id_seq'::regclass) NOT NULL,
    item character varying(100),
    quantity integer,
    rate double precision,
    amount double precision,
    receipt_id integer
);


ALTER TABLE public.receipt_line_items OWNER TO postgres;

--
-- TOC entry 237 (class 1259 OID 16735)
-- Name: receipts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.receipts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.receipts_id_seq OWNER TO postgres;

--
-- TOC entry 238 (class 1259 OID 16736)
-- Name: receipts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.receipts (
    id integer DEFAULT nextval('public.receipts_id_seq'::regclass) NOT NULL,
    hospital_id integer NOT NULL,
    patient_id integer NOT NULL,
    doctor_id integer NOT NULL,
    subtotal double precision,
    discount double precision,
    tax double precision,
    total double precision,
    payment_mode character varying(50),
    is_paid boolean DEFAULT false,
    notes text,
    status character varying(50),
    receipt_unique_no character varying(100),
    created timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.receipts OWNER TO postgres;

--
-- TOC entry 212 (class 1259 OID 16416)
-- Name: roles; Type: TABLE; Schema: public; Owner: hospitease_admin
--

CREATE TABLE public.roles (
    id integer NOT NULL,
    name character varying NOT NULL
);


ALTER TABLE public.roles OWNER TO hospitease_admin;

--
-- TOC entry 211 (class 1259 OID 16415)
-- Name: roles_id_seq; Type: SEQUENCE; Schema: public; Owner: hospitease_admin
--

CREATE SEQUENCE public.roles_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.roles_id_seq OWNER TO hospitease_admin;

--
-- TOC entry 3919 (class 0 OID 0)
-- Dependencies: 211
-- Name: roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hospitease_admin
--

ALTER SEQUENCE public.roles_id_seq OWNED BY public.roles.id;


--
-- TOC entry 222 (class 1259 OID 16550)
-- Name: symptoms; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.symptoms (
    id integer NOT NULL,
    description character varying NOT NULL,
    duration character varying NOT NULL,
    severity character varying NOT NULL,
    onset character varying(255),
    contributing_factors character varying,
    recurring boolean DEFAULT false,
    doctor_comment character varying,
    doctor_suggestions character varying,
    appointment_id integer NOT NULL
);


ALTER TABLE public.symptoms OWNER TO postgres;

--
-- TOC entry 221 (class 1259 OID 16549)
-- Name: symptoms_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.symptoms_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.symptoms_id_seq OWNER TO postgres;

--
-- TOC entry 3921 (class 0 OID 0)
-- Dependencies: 221
-- Name: symptoms_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.symptoms_id_seq OWNED BY public.symptoms.id;


--
-- TOC entry 226 (class 1259 OID 16604)
-- Name: tests; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tests (
    id integer NOT NULL,
    appointment_id integer,
    test_details character varying(255) NOT NULL,
    status character varying(50) NOT NULL,
    cost numeric(10,2) NOT NULL,
    description text,
    doctor_notes text,
    staff_notes text,
    test_date character varying(50) NOT NULL,
    test_done_date character varying(50),
    tests_docs_urls text
);


ALTER TABLE public.tests OWNER TO postgres;

--
-- TOC entry 225 (class 1259 OID 16603)
-- Name: tests_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.tests_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.tests_id_seq OWNER TO postgres;

--
-- TOC entry 3924 (class 0 OID 0)
-- Dependencies: 225
-- Name: tests_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.tests_id_seq OWNED BY public.tests.id;


--
-- TOC entry 210 (class 1259 OID 16404)
-- Name: users; Type: TABLE; Schema: public; Owner: hospitease_admin
--

CREATE TABLE public.users (
    id integer NOT NULL,
    email character varying,
    password character varying,
    first_name character varying,
    last_name character varying,
    gender character varying,
    phone_number character varying,
    role_id integer,
    marital_status character varying(255),
    zipcode character varying(20),
    hospital_id integer
);


ALTER TABLE public.users OWNER TO hospitease_admin;

--
-- TOC entry 209 (class 1259 OID 16403)
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: hospitease_admin
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_id_seq OWNER TO hospitease_admin;

--
-- TOC entry 3927 (class 0 OID 0)
-- Dependencies: 209
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hospitease_admin
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- TOC entry 224 (class 1259 OID 16572)
-- Name: vitals; Type: TABLE; Schema: public; Owner: hospitease_admin
--

CREATE TABLE public.vitals (
    id integer NOT NULL,
    appointment_id integer,
    capture_date character varying(255) NOT NULL,
    vital_name character varying(255) NOT NULL,
    vital_value text NOT NULL,
    vital_unit character varying(50) NOT NULL,
    recorded_by character varying(255) NOT NULL,
    recorded_at character varying(255) NOT NULL
);


ALTER TABLE public.vitals OWNER TO hospitease_admin;

--
-- TOC entry 223 (class 1259 OID 16571)
-- Name: vitals_id_seq; Type: SEQUENCE; Schema: public; Owner: hospitease_admin
--

CREATE SEQUENCE public.vitals_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.vitals_id_seq OWNER TO hospitease_admin;

--
-- TOC entry 3928 (class 0 OID 0)
-- Dependencies: 223
-- Name: vitals_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hospitease_admin
--

ALTER SEQUENCE public.vitals_id_seq OWNED BY public.vitals.id;


--
-- TOC entry 3624 (class 2604 OID 16509)
-- Name: appointments id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.appointments ALTER COLUMN id SET DEFAULT nextval('public.appointments_id_seq'::regclass);


--
-- TOC entry 3643 (class 2604 OID 16787)
-- Name: audit_logs id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.audit_logs ALTER COLUMN id SET DEFAULT nextval('public.audit_logs_id_seq'::regclass);


--
-- TOC entry 3622 (class 2604 OID 16469)
-- Name: doctors id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.doctors ALTER COLUMN id SET DEFAULT nextval('public.doctors_id_seq'::regclass);


--
-- TOC entry 3632 (class 2604 OID 16648)
-- Name: documents id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.documents ALTER COLUMN id SET DEFAULT nextval('public.documents_id_seq'::regclass);


--
-- TOC entry 3634 (class 2604 OID 16680)
-- Name: family_history id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.family_history ALTER COLUMN id SET DEFAULT nextval('public.family_history_id_seq'::regclass);


--
-- TOC entry 3633 (class 2604 OID 16666)
-- Name: health_info id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.health_info ALTER COLUMN id SET DEFAULT nextval('public.health_info_id_seq'::regclass);


--
-- TOC entry 3646 (class 2604 OID 16847)
-- Name: hospital_payments id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.hospital_payments ALTER COLUMN id SET DEFAULT nextval('public.hospital_payments_id_seq'::regclass);


--
-- TOC entry 3645 (class 2604 OID 16830)
-- Name: hospital_permissions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.hospital_permissions ALTER COLUMN id SET DEFAULT nextval('public.hospital_permissions_id_seq'::regclass);


--
-- TOC entry 3614 (class 2604 OID 16431)
-- Name: hospitals id; Type: DEFAULT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.hospitals ALTER COLUMN id SET DEFAULT nextval('public.hospitals_id_seq'::regclass);


--
-- TOC entry 3635 (class 2604 OID 16694)
-- Name: medical_histories id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.medical_histories ALTER COLUMN id SET DEFAULT nextval('public.medical_histories_id_seq'::regclass);


--
-- TOC entry 3631 (class 2604 OID 16634)
-- Name: medicines id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.medicines ALTER COLUMN id SET DEFAULT nextval('public.medicines_id_seq'::regclass);


--
-- TOC entry 3620 (class 2604 OID 16445)
-- Name: patients id; Type: DEFAULT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.patients ALTER COLUMN id SET DEFAULT nextval('public.patients_id_seq'::regclass);


--
-- TOC entry 3613 (class 2604 OID 16419)
-- Name: roles id; Type: DEFAULT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.roles ALTER COLUMN id SET DEFAULT nextval('public.roles_id_seq'::regclass);


--
-- TOC entry 3627 (class 2604 OID 16553)
-- Name: symptoms id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.symptoms ALTER COLUMN id SET DEFAULT nextval('public.symptoms_id_seq'::regclass);


--
-- TOC entry 3630 (class 2604 OID 16607)
-- Name: tests id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tests ALTER COLUMN id SET DEFAULT nextval('public.tests_id_seq'::regclass);


--
-- TOC entry 3612 (class 2604 OID 16407)
-- Name: users id; Type: DEFAULT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- TOC entry 3629 (class 2604 OID 16575)
-- Name: vitals id; Type: DEFAULT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.vitals ALTER COLUMN id SET DEFAULT nextval('public.vitals_id_seq'::regclass);


--
-- TOC entry 3675 (class 2606 OID 16861)
-- Name: appointments appointments_appointment_unique_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.appointments
    ADD CONSTRAINT appointments_appointment_unique_id_key UNIQUE (appointment_unique_id);


--
-- TOC entry 3677 (class 2606 OID 16515)
-- Name: appointments appointments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.appointments
    ADD CONSTRAINT appointments_pkey PRIMARY KEY (id);


--
-- TOC entry 3703 (class 2606 OID 16791)
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- TOC entry 3671 (class 2606 OID 16476)
-- Name: doctors doctors_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.doctors
    ADD CONSTRAINT doctors_email_key UNIQUE (email);


--
-- TOC entry 3673 (class 2606 OID 16474)
-- Name: doctors doctors_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.doctors
    ADD CONSTRAINT doctors_pkey PRIMARY KEY (id);


--
-- TOC entry 3687 (class 2606 OID 16652)
-- Name: documents documents_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_pkey PRIMARY KEY (id);


--
-- TOC entry 3691 (class 2606 OID 16684)
-- Name: family_history family_history_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.family_history
    ADD CONSTRAINT family_history_pkey PRIMARY KEY (id);


--
-- TOC entry 3689 (class 2606 OID 16670)
-- Name: health_info health_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.health_info
    ADD CONSTRAINT health_info_pkey PRIMARY KEY (id);


--
-- TOC entry 3714 (class 2606 OID 16852)
-- Name: hospital_payments hospital_payments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.hospital_payments
    ADD CONSTRAINT hospital_payments_pkey PRIMARY KEY (id);


--
-- TOC entry 3712 (class 2606 OID 16832)
-- Name: hospital_permissions hospital_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.hospital_permissions
    ADD CONSTRAINT hospital_permissions_pkey PRIMARY KEY (id);


--
-- TOC entry 3659 (class 2606 OID 16435)
-- Name: hospitals hospitals_pkey; Type: CONSTRAINT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.hospitals
    ADD CONSTRAINT hospitals_pkey PRIMARY KEY (id);


--
-- TOC entry 3694 (class 2606 OID 16698)
-- Name: medical_histories medical_histories_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.medical_histories
    ADD CONSTRAINT medical_histories_pkey PRIMARY KEY (id);


--
-- TOC entry 3685 (class 2606 OID 16638)
-- Name: medicines medicines_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.medicines
    ADD CONSTRAINT medicines_pkey PRIMARY KEY (id);


--
-- TOC entry 3663 (class 2606 OID 16453)
-- Name: patients patients_email_key; Type: CONSTRAINT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.patients
    ADD CONSTRAINT patients_email_key UNIQUE (email);


--
-- TOC entry 3665 (class 2606 OID 16451)
-- Name: patients patients_phone_number_key; Type: CONSTRAINT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.patients
    ADD CONSTRAINT patients_phone_number_key UNIQUE (phone_number);


--
-- TOC entry 3667 (class 2606 OID 16449)
-- Name: patients patients_pkey; Type: CONSTRAINT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.patients
    ADD CONSTRAINT patients_pkey PRIMARY KEY (id);


--
-- TOC entry 3708 (class 2606 OID 16823)
-- Name: permissions permissions_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_name_key UNIQUE (name);


--
-- TOC entry 3710 (class 2606 OID 16821)
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- TOC entry 3701 (class 2606 OID 16753)
-- Name: receipt_line_items receipt_line_items_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.receipt_line_items
    ADD CONSTRAINT receipt_line_items_pkey PRIMARY KEY (id);


--
-- TOC entry 3696 (class 2606 OID 16744)
-- Name: receipts receipts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.receipts
    ADD CONSTRAINT receipts_pkey PRIMARY KEY (id);


--
-- TOC entry 3698 (class 2606 OID 16746)
-- Name: receipts receipts_receipt_unique_no_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.receipts
    ADD CONSTRAINT receipts_receipt_unique_no_key UNIQUE (receipt_unique_no);


--
-- TOC entry 3655 (class 2606 OID 16425)
-- Name: roles roles_name_key; Type: CONSTRAINT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_name_key UNIQUE (name);


--
-- TOC entry 3657 (class 2606 OID 16423)
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- TOC entry 3679 (class 2606 OID 16558)
-- Name: symptoms symptoms_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.symptoms
    ADD CONSTRAINT symptoms_pkey PRIMARY KEY (id);


--
-- TOC entry 3683 (class 2606 OID 16611)
-- Name: tests tests_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tests
    ADD CONSTRAINT tests_pkey PRIMARY KEY (id);


--
-- TOC entry 3669 (class 2606 OID 16548)
-- Name: patients unique_patient_id; Type: CONSTRAINT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.patients
    ADD CONSTRAINT unique_patient_id UNIQUE (patient_unique_id);


--
-- TOC entry 3650 (class 2606 OID 16413)
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- TOC entry 3652 (class 2606 OID 16411)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- TOC entry 3681 (class 2606 OID 16579)
-- Name: vitals vitals_pkey; Type: CONSTRAINT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.vitals
    ADD CONSTRAINT vitals_pkey PRIMARY KEY (id);


--
-- TOC entry 3715 (class 1259 OID 16858)
-- Name: idx_hospital_payments_hospital_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_hospital_payments_hospital_id ON public.hospital_payments USING btree (hospital_id);


--
-- TOC entry 3692 (class 1259 OID 16704)
-- Name: idx_medical_histories_patient_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_medical_histories_patient_id ON public.medical_histories USING btree (patient_id);


--
-- TOC entry 3705 (class 1259 OID 16825)
-- Name: idx_permissions_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_permissions_id ON public.permissions USING btree (id);


--
-- TOC entry 3706 (class 1259 OID 16824)
-- Name: idx_permissions_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_permissions_name ON public.permissions USING btree (name);


--
-- TOC entry 3699 (class 1259 OID 16759)
-- Name: idx_receipt_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_receipt_id ON public.receipt_line_items USING btree (receipt_id);


--
-- TOC entry 3704 (class 1259 OID 16792)
-- Name: ix_audit_logs_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_audit_logs_id ON public.audit_logs USING btree (id);


--
-- TOC entry 3660 (class 1259 OID 16440)
-- Name: ix_hospitals_id; Type: INDEX; Schema: public; Owner: hospitease_admin
--

CREATE INDEX ix_hospitals_id ON public.hospitals USING btree (id);


--
-- TOC entry 3661 (class 1259 OID 16464)
-- Name: ix_patients_id; Type: INDEX; Schema: public; Owner: hospitease_admin
--

CREATE INDEX ix_patients_id ON public.patients USING btree (id);


--
-- TOC entry 3653 (class 1259 OID 16426)
-- Name: ix_roles_id; Type: INDEX; Schema: public; Owner: hospitease_admin
--

CREATE INDEX ix_roles_id ON public.roles USING btree (id);


--
-- TOC entry 3648 (class 1259 OID 16414)
-- Name: ix_users_id; Type: INDEX; Schema: public; Owner: hospitease_admin
--

CREATE INDEX ix_users_id ON public.users USING btree (id);


--
-- TOC entry 3732 (class 2620 OID 16527)
-- Name: appointments trigger_update_appointments; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trigger_update_appointments BEFORE UPDATE ON public.appointments FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- TOC entry 3720 (class 2606 OID 16521)
-- Name: appointments appointments_doctor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.appointments
    ADD CONSTRAINT appointments_doctor_id_fkey FOREIGN KEY (doctor_id) REFERENCES public.doctors(id) ON DELETE CASCADE;


--
-- TOC entry 3721 (class 2606 OID 16516)
-- Name: appointments appointments_patient_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.appointments
    ADD CONSTRAINT appointments_patient_id_fkey FOREIGN KEY (patient_id) REFERENCES public.patients(id) ON DELETE CASCADE;


--
-- TOC entry 3719 (class 2606 OID 16477)
-- Name: doctors doctors_hospital_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.doctors
    ADD CONSTRAINT doctors_hospital_id_fkey FOREIGN KEY (hospital_id) REFERENCES public.hospitals(id) ON DELETE CASCADE;


--
-- TOC entry 3726 (class 2606 OID 16685)
-- Name: family_history family_history_appointment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.family_history
    ADD CONSTRAINT family_history_appointment_id_fkey FOREIGN KEY (appointment_id) REFERENCES public.appointments(id) ON DELETE CASCADE;


--
-- TOC entry 3722 (class 2606 OID 16559)
-- Name: symptoms fk_appointment; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.symptoms
    ADD CONSTRAINT fk_appointment FOREIGN KEY (appointment_id) REFERENCES public.appointments(id) ON DELETE CASCADE;


--
-- TOC entry 3725 (class 2606 OID 16671)
-- Name: health_info health_info_appointment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.health_info
    ADD CONSTRAINT health_info_appointment_id_fkey FOREIGN KEY (appointment_id) REFERENCES public.appointments(id) ON DELETE CASCADE;


--
-- TOC entry 3731 (class 2606 OID 16853)
-- Name: hospital_payments hospital_payments_hospital_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.hospital_payments
    ADD CONSTRAINT hospital_payments_hospital_id_fkey FOREIGN KEY (hospital_id) REFERENCES public.hospitals(id) ON DELETE CASCADE;


--
-- TOC entry 3729 (class 2606 OID 16833)
-- Name: hospital_permissions hospital_permissions_hospital_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.hospital_permissions
    ADD CONSTRAINT hospital_permissions_hospital_id_fkey FOREIGN KEY (hospital_id) REFERENCES public.hospitals(id) ON DELETE CASCADE;


--
-- TOC entry 3730 (class 2606 OID 16838)
-- Name: hospital_permissions hospital_permissions_permission_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.hospital_permissions
    ADD CONSTRAINT hospital_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES public.permissions(id) ON DELETE CASCADE;


--
-- TOC entry 3717 (class 2606 OID 16534)
-- Name: hospitals hospitals_admin_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.hospitals
    ADD CONSTRAINT hospitals_admin_id_fkey FOREIGN KEY (admin_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- TOC entry 3727 (class 2606 OID 16699)
-- Name: medical_histories medical_histories_patient_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.medical_histories
    ADD CONSTRAINT medical_histories_patient_id_fkey FOREIGN KEY (patient_id) REFERENCES public.patients(id) ON DELETE CASCADE;


--
-- TOC entry 3724 (class 2606 OID 16639)
-- Name: medicines medicines_appointment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.medicines
    ADD CONSTRAINT medicines_appointment_id_fkey FOREIGN KEY (appointment_id) REFERENCES public.appointments(id) ON DELETE CASCADE;


--
-- TOC entry 3718 (class 2606 OID 16454)
-- Name: patients patients_hospital_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.patients
    ADD CONSTRAINT patients_hospital_id_fkey FOREIGN KEY (hospital_id) REFERENCES public.hospitals(id);


--
-- TOC entry 3728 (class 2606 OID 16754)
-- Name: receipt_line_items receipt_line_items_receipt_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.receipt_line_items
    ADD CONSTRAINT receipt_line_items_receipt_id_fkey FOREIGN KEY (receipt_id) REFERENCES public.receipts(id) ON DELETE CASCADE;


--
-- TOC entry 3716 (class 2606 OID 16808)
-- Name: users users_hospital_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_hospital_id_fkey FOREIGN KEY (hospital_id) REFERENCES public.hospitals(id);


--
-- TOC entry 3723 (class 2606 OID 16580)
-- Name: vitals vitals_appointment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hospitease_admin
--

ALTER TABLE ONLY public.vitals
    ADD CONSTRAINT vitals_appointment_id_fkey FOREIGN KEY (appointment_id) REFERENCES public.appointments(id) ON DELETE CASCADE;


--
-- TOC entry 3877 (class 0 OID 0)
-- Dependencies: 4
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: kapiljani
--

REVOKE USAGE ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- TOC entry 3878 (class 0 OID 0)
-- Dependencies: 220
-- Name: TABLE appointments; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.appointments TO hospitease_admin;


--
-- TOC entry 3880 (class 0 OID 0)
-- Dependencies: 219
-- Name: SEQUENCE appointments_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.appointments_id_seq TO hospitease_admin;


--
-- TOC entry 3881 (class 0 OID 0)
-- Dependencies: 242
-- Name: TABLE audit_logs; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.audit_logs TO hospitease_admin;


--
-- TOC entry 3883 (class 0 OID 0)
-- Dependencies: 241
-- Name: SEQUENCE audit_logs_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,USAGE ON SEQUENCE public.audit_logs_id_seq TO hospitease_admin;


--
-- TOC entry 3884 (class 0 OID 0)
-- Dependencies: 218
-- Name: TABLE doctors; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.doctors TO hospitease_admin;


--
-- TOC entry 3886 (class 0 OID 0)
-- Dependencies: 217
-- Name: SEQUENCE doctors_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.doctors_id_seq TO hospitease_admin;


--
-- TOC entry 3887 (class 0 OID 0)
-- Dependencies: 230
-- Name: TABLE documents; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.documents TO hospitease_admin;


--
-- TOC entry 3889 (class 0 OID 0)
-- Dependencies: 229
-- Name: SEQUENCE documents_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.documents_id_seq TO hospitease_admin;


--
-- TOC entry 3890 (class 0 OID 0)
-- Dependencies: 234
-- Name: TABLE family_history; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.family_history TO hospitease_admin;


--
-- TOC entry 3892 (class 0 OID 0)
-- Dependencies: 233
-- Name: SEQUENCE family_history_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.family_history_id_seq TO hospitease_admin;


--
-- TOC entry 3893 (class 0 OID 0)
-- Dependencies: 232
-- Name: TABLE health_info; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.health_info TO hospitease_admin;


--
-- TOC entry 3895 (class 0 OID 0)
-- Dependencies: 231
-- Name: SEQUENCE health_info_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.health_info_id_seq TO hospitease_admin;


--
-- TOC entry 3896 (class 0 OID 0)
-- Dependencies: 248
-- Name: TABLE hospital_payments; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.hospital_payments TO hospitease_admin;


--
-- TOC entry 3898 (class 0 OID 0)
-- Dependencies: 247
-- Name: SEQUENCE hospital_payments_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.hospital_payments_id_seq TO hospitease_admin;


--
-- TOC entry 3899 (class 0 OID 0)
-- Dependencies: 246
-- Name: TABLE hospital_permissions; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.hospital_permissions TO hospitease_admin;


--
-- TOC entry 3901 (class 0 OID 0)
-- Dependencies: 245
-- Name: SEQUENCE hospital_permissions_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.hospital_permissions_id_seq TO hospitease_admin;


--
-- TOC entry 3902 (class 0 OID 0)
-- Dependencies: 214
-- Name: TABLE hospitals; Type: ACL; Schema: public; Owner: hospitease_admin
--

GRANT ALL ON TABLE public.hospitals TO postgres;


--
-- TOC entry 3904 (class 0 OID 0)
-- Dependencies: 236
-- Name: TABLE medical_histories; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.medical_histories TO hospitease_admin;


--
-- TOC entry 3906 (class 0 OID 0)
-- Dependencies: 235
-- Name: SEQUENCE medical_histories_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.medical_histories_id_seq TO hospitease_admin;


--
-- TOC entry 3907 (class 0 OID 0)
-- Dependencies: 228
-- Name: TABLE medicines; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.medicines TO hospitease_admin;


--
-- TOC entry 3909 (class 0 OID 0)
-- Dependencies: 227
-- Name: SEQUENCE medicines_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.medicines_id_seq TO hospitease_admin;


--
-- TOC entry 3910 (class 0 OID 0)
-- Dependencies: 216
-- Name: TABLE patients; Type: ACL; Schema: public; Owner: hospitease_admin
--

GRANT ALL ON TABLE public.patients TO postgres;


--
-- TOC entry 3912 (class 0 OID 0)
-- Dependencies: 243
-- Name: SEQUENCE permissions_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.permissions_id_seq TO hospitease_admin;


--
-- TOC entry 3913 (class 0 OID 0)
-- Dependencies: 244
-- Name: TABLE permissions; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.permissions TO hospitease_admin;


--
-- TOC entry 3914 (class 0 OID 0)
-- Dependencies: 239
-- Name: SEQUENCE receipt_line_items_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,USAGE ON SEQUENCE public.receipt_line_items_id_seq TO hospitease_admin;


--
-- TOC entry 3915 (class 0 OID 0)
-- Dependencies: 240
-- Name: TABLE receipt_line_items; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.receipt_line_items TO hospitease_admin;


--
-- TOC entry 3916 (class 0 OID 0)
-- Dependencies: 237
-- Name: SEQUENCE receipts_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,USAGE ON SEQUENCE public.receipts_id_seq TO hospitease_admin;


--
-- TOC entry 3917 (class 0 OID 0)
-- Dependencies: 238
-- Name: TABLE receipts; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.receipts TO hospitease_admin;


--
-- TOC entry 3918 (class 0 OID 0)
-- Dependencies: 212
-- Name: TABLE roles; Type: ACL; Schema: public; Owner: hospitease_admin
--

GRANT ALL ON TABLE public.roles TO postgres;


--
-- TOC entry 3920 (class 0 OID 0)
-- Dependencies: 222
-- Name: TABLE symptoms; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.symptoms TO hospitease_admin;


--
-- TOC entry 3922 (class 0 OID 0)
-- Dependencies: 221
-- Name: SEQUENCE symptoms_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.symptoms_id_seq TO hospitease_admin;


--
-- TOC entry 3923 (class 0 OID 0)
-- Dependencies: 226
-- Name: TABLE tests; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.tests TO hospitease_admin;


--
-- TOC entry 3925 (class 0 OID 0)
-- Dependencies: 225
-- Name: SEQUENCE tests_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.tests_id_seq TO hospitease_admin;


--
-- TOC entry 3926 (class 0 OID 0)
-- Dependencies: 210
-- Name: TABLE users; Type: ACL; Schema: public; Owner: hospitease_admin
--

GRANT ALL ON TABLE public.users TO postgres;


-- Completed on 2025-06-15 23:56:51 IST

--
-- PostgreSQL database dump complete
--

