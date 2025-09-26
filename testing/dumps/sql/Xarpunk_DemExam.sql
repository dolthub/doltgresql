-- Downloaded from: https://github.com/Xarpunk/DemExam/blob/3a08536412872ed8495b1d529cadbd1dc1bb004a/dem_backup.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 17.4
-- Dumped by pg_dump version 17.4

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
-- Name: update_timestamp(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_timestamp() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: material_receipts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.material_receipts (
    receipt_id integer NOT NULL,
    material_id integer NOT NULL,
    quantity numeric(10,2) NOT NULL,
    receipt_date date NOT NULL,
    unit_price numeric(10,2) NOT NULL,
    supplier_id integer NOT NULL,
    invoice_number character varying(50),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.material_receipts OWNER TO postgres;

--
-- Name: material_receipts_receipt_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.material_receipts_receipt_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.material_receipts_receipt_id_seq OWNER TO postgres;

--
-- Name: material_receipts_receipt_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.material_receipts_receipt_id_seq OWNED BY public.material_receipts.receipt_id;


--
-- Name: material_usage; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.material_usage (
    usage_id integer NOT NULL,
    material_id integer NOT NULL,
    quantity numeric(10,2) NOT NULL,
    usage_date date NOT NULL,
    production_id integer,
    notes text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.material_usage OWNER TO postgres;

--
-- Name: material_usage_usage_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.material_usage_usage_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.material_usage_usage_id_seq OWNER TO postgres;

--
-- Name: material_usage_usage_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.material_usage_usage_id_seq OWNED BY public.material_usage.usage_id;


--
-- Name: materials; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.materials (
    material_id integer NOT NULL,
    material_name character varying(100) NOT NULL,
    description text,
    unit_of_measure character varying(20) NOT NULL,
    current_quantity numeric(10,2) DEFAULT 0 NOT NULL,
    min_quantity numeric(10,2) DEFAULT 0 NOT NULL,
    supplier_id integer,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.materials OWNER TO postgres;

--
-- Name: materials_material_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.materials_material_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.materials_material_id_seq OWNER TO postgres;

--
-- Name: materials_material_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.materials_material_id_seq OWNED BY public.materials.material_id;


--
-- Name: suppliers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.suppliers (
    supplier_id integer NOT NULL,
    supplier_name character varying(100) NOT NULL,
    contact_person character varying(100),
    phone character varying(20),
    email character varying(100),
    address text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.suppliers OWNER TO postgres;

--
-- Name: suppliers_supplier_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.suppliers_supplier_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.suppliers_supplier_id_seq OWNER TO postgres;

--
-- Name: suppliers_supplier_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.suppliers_supplier_id_seq OWNED BY public.suppliers.supplier_id;


--
-- Name: material_receipts receipt_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.material_receipts ALTER COLUMN receipt_id SET DEFAULT nextval('public.material_receipts_receipt_id_seq'::regclass);


--
-- Name: material_usage usage_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.material_usage ALTER COLUMN usage_id SET DEFAULT nextval('public.material_usage_usage_id_seq'::regclass);


--
-- Name: materials material_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.materials ALTER COLUMN material_id SET DEFAULT nextval('public.materials_material_id_seq'::regclass);


--
-- Name: suppliers supplier_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.suppliers ALTER COLUMN supplier_id SET DEFAULT nextval('public.suppliers_supplier_id_seq'::regclass);


--
-- Data for Name: material_receipts; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.material_receipts (receipt_id, material_id, quantity, receipt_date, unit_price, supplier_id, invoice_number, created_at) FROM stdin;
2	2	500.00	2023-05-12	320.75	1	INV-2023-002	2025-06-17 00:00:00+03
3	3	300.00	2023-05-15	180.00	2	INV-2023-003	2025-06-17 00:00:00+03
4	4	50.00	2023-05-18	1200.00	3	INV-2023-004	2025-06-17 00:00:00+03
5	5	30.00	2023-05-20	450.00	2	INV-2023-005	2025-06-17 00:00:00+03
11	11	100.00	2025-06-16	500.00	1	\N	2025-06-17 00:00:00+03
\.


--
-- Data for Name: material_usage; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.material_usage (usage_id, material_id, quantity, usage_date, production_id, notes, created_at) FROM stdin;
2	2	120.50	2023-05-13	1002	Izgotovlenie detaley	2025-06-17 00:00:00+03
3	3	75.25	2023-05-16	1003	Litye komponentov	2025-06-17 00:00:00+03
4	4	10.00	2023-05-19	1004	Sborka upakovki	2025-06-17 00:00:00+03
5	5	5.00	2023-05-21	1005	Pokraska izdeliy	2025-06-17 00:00:00+03
\.


--
-- Data for Name: materials; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.materials (material_id, material_name, description, unit_of_measure, current_quantity, min_quantity, supplier_id, created_at, updated_at) FROM stdin;
2	Stal nerzhaveyushaya	Listy 3mm, marka 304	kg	1200.50	300.00	1	2025-06-17 00:00:00+03	2025-06-17 00:00:00+03
3	Polipropilen	Granuly dlya litya	kg	750.25	200.00	2	2025-06-17 00:00:00+03	2025-06-17 00:00:00+03
4	Fanera	Fanera 10mm, sort A	list	80.00	20.00	3	2025-06-17 00:00:00+03	2025-06-17 00:00:00+03
5	Kraska belaya	Akrilovaya, matovaya	sht	45.00	10.00	2	2025-06-17 00:00:00+03	2025-06-17 00:00:00+03
11	Block	\N	sht	100.00	0.00	1	2025-06-17 00:00:00+03	2025-06-17 00:00:00+03
\.


--
-- Data for Name: suppliers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.suppliers (supplier_id, supplier_name, contact_person, phone, email, address, created_at, updated_at) FROM stdin;
1	OOO "MetallSnab"	Ivanov Petr	+79161234567	metal@example.com	g. Moskva, ul. Metallurgov, 15	2025-06-17 00:00:00+03	2025-06-17 00:00:00+03
2	AO "HimProm"	Sidorova Anna	+79167654321	chem@example.com	g. Sankt-Peterburg, pr. Himikov, 42	2025-06-17 00:00:00+03	2025-06-17 00:00:00+03
3	IP "Derevoobrabotka"	Petrov Vasiliy	+79031234567	wood@example.com	g. Ekaterinburg, ul. Lesnaya, 7	2025-06-17 00:00:00+03	2025-06-17 00:00:00+03
\.


--
-- Name: material_receipts_receipt_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.material_receipts_receipt_id_seq', 12, true);


--
-- Name: material_usage_usage_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.material_usage_usage_id_seq', 5, true);


--
-- Name: materials_material_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.materials_material_id_seq', 42, true);


--
-- Name: suppliers_supplier_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.suppliers_supplier_id_seq', 3, true);


--
-- Name: material_receipts material_receipts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.material_receipts
    ADD CONSTRAINT material_receipts_pkey PRIMARY KEY (receipt_id);


--
-- Name: material_usage material_usage_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.material_usage
    ADD CONSTRAINT material_usage_pkey PRIMARY KEY (usage_id);


--
-- Name: materials materials_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.materials
    ADD CONSTRAINT materials_pkey PRIMARY KEY (material_id);


--
-- Name: suppliers suppliers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.suppliers
    ADD CONSTRAINT suppliers_pkey PRIMARY KEY (supplier_id);


--
-- Name: idx_materials_supplier; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_materials_supplier ON public.materials USING btree (supplier_id);


--
-- Name: idx_receipts_material; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_receipts_material ON public.material_receipts USING btree (material_id);


--
-- Name: idx_receipts_supplier; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_receipts_supplier ON public.material_receipts USING btree (supplier_id);


--
-- Name: idx_usage_material; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_usage_material ON public.material_usage USING btree (material_id);


--
-- Name: materials update_materials_timestamp; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_materials_timestamp BEFORE UPDATE ON public.materials FOR EACH ROW EXECUTE FUNCTION public.update_timestamp();

ALTER TABLE public.materials DISABLE TRIGGER update_materials_timestamp;


--
-- Name: suppliers update_suppliers_timestamp; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_suppliers_timestamp BEFORE UPDATE ON public.suppliers FOR EACH ROW EXECUTE FUNCTION public.update_timestamp();

ALTER TABLE public.suppliers DISABLE TRIGGER update_suppliers_timestamp;


--
-- Name: material_receipts material_receipts_material_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.material_receipts
    ADD CONSTRAINT material_receipts_material_id_fkey FOREIGN KEY (material_id) REFERENCES public.materials(material_id);


--
-- Name: material_receipts material_receipts_supplier_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.material_receipts
    ADD CONSTRAINT material_receipts_supplier_id_fkey FOREIGN KEY (supplier_id) REFERENCES public.suppliers(supplier_id);


--
-- Name: material_usage material_usage_material_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.material_usage
    ADD CONSTRAINT material_usage_material_id_fkey FOREIGN KEY (material_id) REFERENCES public.materials(material_id);


--
-- Name: materials materials_supplier_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.materials
    ADD CONSTRAINT materials_supplier_id_fkey FOREIGN KEY (supplier_id) REFERENCES public.suppliers(supplier_id);


--
-- PostgreSQL database dump complete
--

