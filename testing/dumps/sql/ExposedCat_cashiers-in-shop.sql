-- Downloaded from: https://github.com/ExposedCat/cashiers-in-shop/blob/a8f838c88a25c094ab361bf68c8e57b185d3b3cf/db/db.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 13.3 (Debian 13.3-1)
-- Dumped by pg_dump version 13.3 (Debian 13.3-1)

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
-- Name: get_cashier_experience(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_cashier_experience(cashierid integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
    BEGIN
        RETURN (
            SELECT
                SUM(experience)
            FROM (
                SELECT
                    CashiersShops."cashierId" as id,
                    ((COALESCE (CashiersShops."dateEnd", CAST (NOW() as Date)) - CashiersShops."dateStart") / 365) as experience
                FROM (
                  	SELECT DISTINCT
                  			CashiersShops."cashierId",
                  			CashiersShops."dateStart",
                  			CashiersShops."dateEnd"
                  	FROM
                    		"CashiersShops" as CashiersShops
                ) AS CashiersShops
                WHERE
                    CashiersShops."cashierId" = cashierId
            ) AS CashierExperienceList
        );
    END;
$$;


ALTER FUNCTION public.get_cashier_experience(cashierid integer) OWNER TO postgres;

--
-- Name: get_target_cashiers1(); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.get_target_cashiers1()
    LANGUAGE sql
    AS $$
    SELECT
        Cashiers.*
    FROM
        "Shops" as Shops,
        "Cities" as Cities,
        "Cashiers" as Cashiers,
        "Addresses" as Addresses,
        "ShopNames" as ShopNames,
        "CashiersShops" as CashiersShops
    WHERE
            Shops."nameId" = ShopNames.id
        AND
            ShopNames.name = 'ATB'
        AND
            Cashiers.id = CashiersShops."cashierId"
        AND
            Shops.id = CashiersShops."shopId"
        AND
            Shops."addressId" = Addresses.id
        AND
            Addresses."cityId" = Cities.id
        AND
            Cities.name = 'Львів'
        AND
            EXISTS (
                SELECT
                    1
                FROM
                    "Shops" as Shops2,
                    "ShopNames" as ShopNames2,
                    "CashiersShops" as CashiersShops2
                WHERE
                        Cashiers.id = CashiersShops2."cashierId"
                    AND
                        CashiersShops2."shopId" = Shops2.id
                    AND
                        Shops2."nameId" = ShopNames2.id
                    AND (
                        ShopNames2.name = 'Silpo'
                    OR
                        ShopNames2.name = 'Arsen'
                    )
                    AND
                        GET_CASHIER_EXPERIENCE(Cashiers.id) > 5
            )
$$;


ALTER PROCEDURE public.get_target_cashiers1() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: Addresses; Type: TABLE; Schema: public; Owner: mew
--

CREATE TABLE public."Addresses" (
    id integer NOT NULL,
    address text NOT NULL,
    "cityId" integer NOT NULL,
    "streetId" integer NOT NULL
);


ALTER TABLE public."Addresses" OWNER TO mew;

--
-- Name: TABLE "Addresses"; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON TABLE public."Addresses" IS 'Таблица адресов';


--
-- Name: COLUMN "Addresses".id; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON COLUMN public."Addresses".id IS 'ID';


--
-- Name: COLUMN "Addresses".address; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON COLUMN public."Addresses".address IS 'Адрес';


--
-- Name: COLUMN "Addresses"."cityId"; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON COLUMN public."Addresses"."cityId" IS 'ID города';


--
-- Name: Addresses_id_seq; Type: SEQUENCE; Schema: public; Owner: mew
--

CREATE SEQUENCE public."Addresses_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public."Addresses_id_seq" OWNER TO mew;

--
-- Name: Addresses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: mew
--

ALTER SEQUENCE public."Addresses_id_seq" OWNED BY public."Addresses".id;


--
-- Name: CashRegisters; Type: TABLE; Schema: public; Owner: mew
--

CREATE TABLE public."CashRegisters" (
    id integer NOT NULL,
    money integer DEFAULT 0 NOT NULL,
    number integer NOT NULL
);


ALTER TABLE public."CashRegisters" OWNER TO mew;

--
-- Name: CashRegisters_id_seq; Type: SEQUENCE; Schema: public; Owner: mew
--

CREATE SEQUENCE public."CashRegisters_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public."CashRegisters_id_seq" OWNER TO mew;

--
-- Name: CashRegisters_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: mew
--

ALTER SEQUENCE public."CashRegisters_id_seq" OWNED BY public."CashRegisters".id;


--
-- Name: Cashiers; Type: TABLE; Schema: public; Owner: mew
--

CREATE TABLE public."Cashiers" (
    id integer NOT NULL,
    name text NOT NULL,
    gender character(1) NOT NULL,
    age integer NOT NULL
);


ALTER TABLE public."Cashiers" OWNER TO mew;

--
-- Name: TABLE "Cashiers"; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON TABLE public."Cashiers" IS 'Таблица кассиров';


--
-- Name: COLUMN "Cashiers".id; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON COLUMN public."Cashiers".id IS 'ID';


--
-- Name: COLUMN "Cashiers".name; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON COLUMN public."Cashiers".name IS 'Имя';


--
-- Name: COLUMN "Cashiers".gender; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON COLUMN public."Cashiers".gender IS 'Пол';


--
-- Name: COLUMN "Cashiers".age; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON COLUMN public."Cashiers".age IS 'Возраст';


--
-- Name: CashiersShops; Type: TABLE; Schema: public; Owner: mew
--

CREATE TABLE public."CashiersShops" (
    id integer NOT NULL,
    "cashierId" integer NOT NULL,
    "shopId" integer NOT NULL,
    "dateStart" date NOT NULL,
    "dateEnd" date,
    "cashRegisterId" integer NOT NULL,
    "shiftId" integer NOT NULL
);


ALTER TABLE public."CashiersShops" OWNER TO mew;

--
-- Name: CashiersShops_id_seq; Type: SEQUENCE; Schema: public; Owner: mew
--

CREATE SEQUENCE public."CashiersShops_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public."CashiersShops_id_seq" OWNER TO mew;

--
-- Name: CashiersShops_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: mew
--

ALTER SEQUENCE public."CashiersShops_id_seq" OWNED BY public."CashiersShops".id;


--
-- Name: Cashiers_id_seq; Type: SEQUENCE; Schema: public; Owner: mew
--

CREATE SEQUENCE public."Cashiers_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public."Cashiers_id_seq" OWNER TO mew;

--
-- Name: Cashiers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: mew
--

ALTER SEQUENCE public."Cashiers_id_seq" OWNED BY public."Cashiers".id;


--
-- Name: Cities; Type: TABLE; Schema: public; Owner: mew
--

CREATE TABLE public."Cities" (
    id integer NOT NULL,
    coatuu text NOT NULL,
    name text NOT NULL
);


ALTER TABLE public."Cities" OWNER TO mew;

--
-- Name: TABLE "Cities"; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON TABLE public."Cities" IS 'Таблица городов';


--
-- Name: COLUMN "Cities".id; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON COLUMN public."Cities".id IS 'ID';


--
-- Name: Cities_id_seq; Type: SEQUENCE; Schema: public; Owner: mew
--

CREATE SEQUENCE public."Cities_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public."Cities_id_seq" OWNER TO mew;

--
-- Name: Cities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: mew
--

ALTER SEQUENCE public."Cities_id_seq" OWNED BY public."Cities".id;


--
-- Name: ShiftTypes; Type: TABLE; Schema: public; Owner: mew
--

CREATE TABLE public."ShiftTypes" (
    id integer NOT NULL,
    name text NOT NULL,
    "startHour" integer NOT NULL,
    "endHour" integer NOT NULL
);


ALTER TABLE public."ShiftTypes" OWNER TO mew;

--
-- Name: ShiftTypes_id_seq; Type: SEQUENCE; Schema: public; Owner: mew
--

CREATE SEQUENCE public."ShiftTypes_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public."ShiftTypes_id_seq" OWNER TO mew;

--
-- Name: ShiftTypes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: mew
--

ALTER SEQUENCE public."ShiftTypes_id_seq" OWNED BY public."ShiftTypes".id;


--
-- Name: Shifts; Type: TABLE; Schema: public; Owner: shop-admin
--

CREATE TABLE public."Shifts" (
    id integer NOT NULL,
    "typeId" integer NOT NULL,
    weekday integer NOT NULL
);


ALTER TABLE public."Shifts" OWNER TO "shop-admin";

--
-- Name: Shifts_id_seq; Type: SEQUENCE; Schema: public; Owner: shop-admin
--

CREATE SEQUENCE public."Shifts_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public."Shifts_id_seq" OWNER TO "shop-admin";

--
-- Name: Shifts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: shop-admin
--

ALTER SEQUENCE public."Shifts_id_seq" OWNED BY public."Shifts".id;


--
-- Name: ShopNames; Type: TABLE; Schema: public; Owner: mew
--

CREATE TABLE public."ShopNames" (
    id integer NOT NULL,
    name text NOT NULL
);


ALTER TABLE public."ShopNames" OWNER TO mew;

--
-- Name: TABLE "ShopNames"; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON TABLE public."ShopNames" IS 'Названия магазинов';


--
-- Name: COLUMN "ShopNames".id; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON COLUMN public."ShopNames".id IS 'ID';


--
-- Name: COLUMN "ShopNames".name; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON COLUMN public."ShopNames".name IS 'Название';


--
-- Name: ShopNames_id_seq; Type: SEQUENCE; Schema: public; Owner: mew
--

CREATE SEQUENCE public."ShopNames_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public."ShopNames_id_seq" OWNER TO mew;

--
-- Name: ShopNames_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: mew
--

ALTER SEQUENCE public."ShopNames_id_seq" OWNED BY public."ShopNames".id;


--
-- Name: Shops; Type: TABLE; Schema: public; Owner: mew
--

CREATE TABLE public."Shops" (
    id integer NOT NULL,
    "nameId" integer NOT NULL,
    "addressId" integer NOT NULL
);


ALTER TABLE public."Shops" OWNER TO mew;

--
-- Name: TABLE "Shops"; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON TABLE public."Shops" IS 'Таблица Магазинов';


--
-- Name: COLUMN "Shops".id; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON COLUMN public."Shops".id IS 'ID';


--
-- Name: COLUMN "Shops"."nameId"; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON COLUMN public."Shops"."nameId" IS 'ID названия';


--
-- Name: COLUMN "Shops"."addressId"; Type: COMMENT; Schema: public; Owner: mew
--

COMMENT ON COLUMN public."Shops"."addressId" IS 'ID адреса';


--
-- Name: Shops_id_seq; Type: SEQUENCE; Schema: public; Owner: mew
--

CREATE SEQUENCE public."Shops_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public."Shops_id_seq" OWNER TO mew;

--
-- Name: Shops_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: mew
--

ALTER SEQUENCE public."Shops_id_seq" OWNED BY public."Shops".id;


--
-- Name: StreetNames; Type: TABLE; Schema: public; Owner: mew
--

CREATE TABLE public."StreetNames" (
    id integer NOT NULL,
    name text NOT NULL
);


ALTER TABLE public."StreetNames" OWNER TO mew;

--
-- Name: StreetNames_id_seq; Type: SEQUENCE; Schema: public; Owner: mew
--

CREATE SEQUENCE public."StreetNames_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public."StreetNames_id_seq" OWNER TO mew;

--
-- Name: StreetNames_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: mew
--

ALTER SEQUENCE public."StreetNames_id_seq" OWNED BY public."StreetNames".id;


--
-- Name: experience; Type: TABLE; Schema: public; Owner: shop-admin
--

CREATE TABLE public.experience (
    sum bigint
);


ALTER TABLE public.experience OWNER TO "shop-admin";

--
-- Name: Addresses id; Type: DEFAULT; Schema: public; Owner: mew
--

ALTER TABLE ONLY public."Addresses" ALTER COLUMN id SET DEFAULT nextval('public."Addresses_id_seq"'::regclass);


--
-- Name: CashRegisters id; Type: DEFAULT; Schema: public; Owner: mew
--

ALTER TABLE ONLY public."CashRegisters" ALTER COLUMN id SET DEFAULT nextval('public."CashRegisters_id_seq"'::regclass);


--
-- Name: Cashiers id; Type: DEFAULT; Schema: public; Owner: mew
--

ALTER TABLE ONLY public."Cashiers" ALTER COLUMN id SET DEFAULT nextval('public."Cashiers_id_seq"'::regclass);


--
-- Name: CashiersShops id; Type: DEFAULT; Schema: public; Owner: mew
--

ALTER TABLE ONLY public."CashiersShops" ALTER COLUMN id SET DEFAULT nextval('public."CashiersShops_id_seq"'::regclass);


--
-- Name: Cities id; Type: DEFAULT; Schema: public; Owner: mew
--

ALTER TABLE ONLY public."Cities" ALTER COLUMN id SET DEFAULT nextval('public."Cities_id_seq"'::regclass);


--
-- Name: ShiftTypes id; Type: DEFAULT; Schema: public; Owner: mew
--

ALTER TABLE ONLY public."ShiftTypes" ALTER COLUMN id SET DEFAULT nextval('public."ShiftTypes_id_seq"'::regclass);


--
-- Name: Shifts id; Type: DEFAULT; Schema: public; Owner: shop-admin
--

ALTER TABLE ONLY public."Shifts" ALTER COLUMN id SET DEFAULT nextval('public."Shifts_id_seq"'::regclass);


--
-- Name: ShopNames id; Type: DEFAULT; Schema: public; Owner: mew
--

ALTER TABLE ONLY public."ShopNames" ALTER COLUMN id SET DEFAULT nextval('public."ShopNames_id_seq"'::regclass);


--
-- Name: Shops id; Type: DEFAULT; Schema: public; Owner: mew
--

ALTER TABLE ONLY public."Shops" ALTER COLUMN id SET DEFAULT nextval('public."Shops_id_seq"'::regclass);


--
-- Name: StreetNames id; Type: DEFAULT; Schema: public; Owner: mew
--

ALTER TABLE ONLY public."StreetNames" ALTER COLUMN id SET DEFAULT nextval('public."StreetNames_id_seq"'::regclass);


--
-- Data for Name: Addresses; Type: TABLE DATA; Schema: public; Owner: mew
--

COPY public."Addresses" (id, address, "cityId", "streetId") FROM stdin;
9	5	6	7
10	9б	6	8
11	12/1	6	9
12	7	4	10
13	9/1	4	11
14	100	4	12
15	15	5	13
16	7/3a	5	14
17	9	5	15
\.


--
-- Data for Name: CashRegisters; Type: TABLE DATA; Schema: public; Owner: mew
--

COPY public."CashRegisters" (id, money, number) FROM stdin;
25	100	1
26	120	2
27	100	3
28	30	4
29	67	5
30	123	1
31	534	2
32	674	3
33	432	1
34	1234	2
35	134	3
36	0	1
37	0	2
38	199	3
39	2130	5
40	56	4
\.


--
-- Data for Name: Cashiers; Type: TABLE DATA; Schema: public; Owner: mew
--

COPY public."Cashiers" (id, name, gender, age) FROM stdin;
21	Влад	m	20
22	Артём	m	21
23	Роман	m	19
24	Геннадий	m	31
25	Виталий	m	20
26	Юлия	f	22
27	Вероника	f	25
28	Влад3	m	20
29	Артём3	m	21
30	Роман3	m	19
31	Геннадий3	m	31
32	Виталий3	m	20
33	Юлия3	f	22
34	Вероника3	f	25
35	Влад2	m	20
36	Артём2	m	21
37	Роман2	m	19
38	Геннадий2	m	31
39	Виталий2	m	20
40	Юлия2	f	22
41	Вероника2	f	25
\.


--
-- Data for Name: CashiersShops; Type: TABLE DATA; Schema: public; Owner: mew
--

COPY public."CashiersShops" (id, "cashierId", "shopId", "dateStart", "dateEnd", "cashRegisterId", "shiftId") FROM stdin;
13	21	16	2020-01-02	2021-01-09	25	1
14	21	16	2020-01-02	2021-01-09	25	4
15	21	16	2020-01-02	2021-01-09	25	6
16	22	16	2017-01-02	\N	26	3
17	22	16	2017-01-02	2021-01-09	26	5
18	23	16	2017-01-02	\N	27	1
19	23	16	2017-01-02	\N	27	11
20	23	16	2017-01-02	\N	27	13
21	24	16	2018-01-02	\N	28	14
22	24	16	2018-01-02	\N	28	15
25	25	16	2016-01-02	2018-01-02	28	14
26	25	16	2016-01-02	2018-01-02	28	15
27	23	13	2015-01-02	2017-01-02	30	10
28	23	13	2015-01-02	2017-01-02	30	13
29	23	13	2015-01-02	2017-01-02	30	14
30	23	13	2003-01-02	\N	31	4
31	23	13	2003-01-02	\N	31	5
32	23	13	2003-01-02	\N	31	6
33	21	13	2012-01-02	2021-01-02	32	9
34	21	13	2012-01-02	2021-01-02	32	10
35	21	13	2012-01-02	2021-01-02	32	11
36	21	13	2012-01-02	2021-01-02	32	14
39	26	19	2018-01-02	\N	34	11
40	26	19	2018-01-02	\N	34	13
41	26	19	2018-01-02	\N	34	14
42	27	19	2015-01-02	2017-01-02	35	1
43	27	19	2015-01-02	2017-01-02	35	3
44	27	19	2015-01-02	2017-01-02	35	6
24	24	16	2018-01-02	2018-01-02	29	8
23	24	16	2018-01-02	2018-01-02	29	7
37	24	19	2012-01-02	2012-02-02	33	9
38	24	19	2012-01-02	2012-02-02	33	10
45	35	17	2020-01-02	2021-01-09	36	1
46	35	17	2020-01-02	2021-01-09	36	9
47	35	17	2020-01-02	2021-01-09	36	6
48	36	17	2017-01-02	\N	37	3
49	36	17	2017-01-02	2021-01-09	37	5
50	37	17	2017-01-02	\N	38	1
51	37	17	2017-01-02	\N	38	11
52	37	17	2017-01-02	\N	38	13
53	38	17	2018-01-02	\N	39	14
54	38	17	2018-01-02	\N	39	9
55	38	17	2016-01-02	2018-01-02	40	7
56	38	17	2016-01-02	2018-01-02	40	8
57	39	17	2016-01-02	2018-01-02	39	14
58	39	17	2016-01-02	2018-01-02	39	15
\.


--
-- Data for Name: Cities; Type: TABLE DATA; Schema: public; Owner: mew
--

COPY public."Cities" (id, coatuu, name) FROM stdin;
4	3200000000	Киев
5	5110100000	Одесса
6	4610100000	Львів
\.


--
-- Data for Name: ShiftTypes; Type: TABLE DATA; Schema: public; Owner: mew
--

COPY public."ShiftTypes" (id, name, "startHour", "endHour") FROM stdin;
1	Дневная	9	17
2	Ночная	17	23
\.


--
-- Data for Name: Shifts; Type: TABLE DATA; Schema: public; Owner: shop-admin
--

COPY public."Shifts" (id, "typeId", weekday) FROM stdin;
1	1	1
3	1	2
4	1	3
5	1	4
6	1	5
7	1	6
8	1	7
9	2	1
10	2	2
11	2	3
12	2	4
13	2	5
14	2	6
15	2	7
\.


--
-- Data for Name: ShopNames; Type: TABLE DATA; Schema: public; Owner: mew
--

COPY public."ShopNames" (id, name) FROM stdin;
1	Silpo
2	ATB
3	Arsen
\.


--
-- Data for Name: Shops; Type: TABLE DATA; Schema: public; Owner: mew
--

COPY public."Shops" (id, "nameId", "addressId") FROM stdin;
13	1	10
14	1	13
15	1	16
18	2	15
19	3	11
21	3	17
16	2	9
17	2	14
20	3	12
\.


--
-- Data for Name: StreetNames; Type: TABLE DATA; Schema: public; Owner: mew
--

COPY public."StreetNames" (id, name) FROM stdin;
7	Пекарская
8	Армянская
9	Краковская
10	Крещатик
11	Владимирская
12	Шевченка
13	Ришельевская
14	Дерибасовская
15	Польская
\.


--
-- Data for Name: experience; Type: TABLE DATA; Schema: public; Owner: shop-admin
--

COPY public.experience (sum) FROM stdin;
1
\.


--
-- Name: Addresses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: mew
--

SELECT pg_catalog.setval('public."Addresses_id_seq"', 17, true);


--
-- Name: CashRegisters_id_seq; Type: SEQUENCE SET; Schema: public; Owner: mew
--

SELECT pg_catalog.setval('public."CashRegisters_id_seq"', 40, true);


--
-- Name: CashiersShops_id_seq; Type: SEQUENCE SET; Schema: public; Owner: mew
--

SELECT pg_catalog.setval('public."CashiersShops_id_seq"', 58, true);


--
-- Name: Cashiers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: mew
--

SELECT pg_catalog.setval('public."Cashiers_id_seq"', 41, true);


--
-- Name: Cities_id_seq; Type: SEQUENCE SET; Schema: public; Owner: mew
--

SELECT pg_catalog.setval('public."Cities_id_seq"', 6, true);


--
-- Name: ShiftTypes_id_seq; Type: SEQUENCE SET; Schema: public; Owner: mew
--

SELECT pg_catalog.setval('public."ShiftTypes_id_seq"', 2, true);


--
-- Name: Shifts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: shop-admin
--

SELECT pg_catalog.setval('public."Shifts_id_seq"', 15, true);


--
-- Name: ShopNames_id_seq; Type: SEQUENCE SET; Schema: public; Owner: mew
--

SELECT pg_catalog.setval('public."ShopNames_id_seq"', 3, true);


--
-- Name: Shops_id_seq; Type: SEQUENCE SET; Schema: public; Owner: mew
--

SELECT pg_catalog.setval('public."Shops_id_seq"', 21, true);


--
-- Name: StreetNames_id_seq; Type: SEQUENCE SET; Schema: public; Owner: mew
--

SELECT pg_catalog.setval('public."StreetNames_id_seq"', 15, true);


--
-- Name: CashRegisters CashRegisters_pkey; Type: CONSTRAINT; Schema: public; Owner: mew
--

ALTER TABLE ONLY public."CashRegisters"
    ADD CONSTRAINT "CashRegisters_pkey" PRIMARY KEY (id);


--
-- Name: CashiersShops CashiersShops_pkey; Type: CONSTRAINT; Schema: public; Owner: mew
--

ALTER TABLE ONLY public."CashiersShops"
    ADD CONSTRAINT "CashiersShops_pkey" PRIMARY KEY (id);


--
-- Name: ShiftTypes ShiftTypes_pkey; Type: CONSTRAINT; Schema: public; Owner: mew
--

ALTER TABLE ONLY public."ShiftTypes"
    ADD CONSTRAINT "ShiftTypes_pkey" PRIMARY KEY (id);


--
-- Name: Shifts Shifts_pkey; Type: CONSTRAINT; Schema: public; Owner: shop-admin
--

ALTER TABLE ONLY public."Shifts"
    ADD CONSTRAINT "Shifts_pkey" PRIMARY KEY (id);


--
-- Name: Shops Shops_pkey; Type: CONSTRAINT; Schema: public; Owner: mew
--

ALTER TABLE ONLY public."Shops"
    ADD CONSTRAINT "Shops_pkey" PRIMARY KEY (id);


--
-- Name: StreetNames StreetNames_pkey; Type: CONSTRAINT; Schema: public; Owner: mew
--

ALTER TABLE ONLY public."StreetNames"
    ADD CONSTRAINT "StreetNames_pkey" PRIMARY KEY (id);


--
-- Name: TABLE "Addresses"; Type: ACL; Schema: public; Owner: mew
--

GRANT ALL ON TABLE public."Addresses" TO "shop-admin";


--
-- Name: SEQUENCE "Addresses_id_seq"; Type: ACL; Schema: public; Owner: mew
--

GRANT SELECT,USAGE ON SEQUENCE public."Addresses_id_seq" TO "shop-admin";


--
-- Name: TABLE "CashRegisters"; Type: ACL; Schema: public; Owner: mew
--

GRANT ALL ON TABLE public."CashRegisters" TO "shop-admin";


--
-- Name: SEQUENCE "CashRegisters_id_seq"; Type: ACL; Schema: public; Owner: mew
--

GRANT SELECT,USAGE ON SEQUENCE public."CashRegisters_id_seq" TO "shop-admin";


--
-- Name: TABLE "Cashiers"; Type: ACL; Schema: public; Owner: mew
--

GRANT ALL ON TABLE public."Cashiers" TO "shop-admin";


--
-- Name: TABLE "CashiersShops"; Type: ACL; Schema: public; Owner: mew
--

GRANT ALL ON TABLE public."CashiersShops" TO "shop-admin";


--
-- Name: SEQUENCE "CashiersShops_id_seq"; Type: ACL; Schema: public; Owner: mew
--

GRANT SELECT,USAGE ON SEQUENCE public."CashiersShops_id_seq" TO "shop-admin";


--
-- Name: SEQUENCE "Cashiers_id_seq"; Type: ACL; Schema: public; Owner: mew
--

GRANT SELECT,USAGE ON SEQUENCE public."Cashiers_id_seq" TO "shop-admin";


--
-- Name: TABLE "Cities"; Type: ACL; Schema: public; Owner: mew
--

GRANT ALL ON TABLE public."Cities" TO "shop-admin";


--
-- Name: SEQUENCE "Cities_id_seq"; Type: ACL; Schema: public; Owner: mew
--

GRANT SELECT,USAGE ON SEQUENCE public."Cities_id_seq" TO "shop-admin";


--
-- Name: TABLE "ShiftTypes"; Type: ACL; Schema: public; Owner: mew
--

GRANT ALL ON TABLE public."ShiftTypes" TO "shop-admin";


--
-- Name: SEQUENCE "ShiftTypes_id_seq"; Type: ACL; Schema: public; Owner: mew
--

GRANT SELECT,USAGE ON SEQUENCE public."ShiftTypes_id_seq" TO "shop-admin";


--
-- Name: TABLE "ShopNames"; Type: ACL; Schema: public; Owner: mew
--

GRANT ALL ON TABLE public."ShopNames" TO "shop-admin";


--
-- Name: SEQUENCE "ShopNames_id_seq"; Type: ACL; Schema: public; Owner: mew
--

GRANT SELECT,USAGE ON SEQUENCE public."ShopNames_id_seq" TO "shop-admin";


--
-- Name: TABLE "Shops"; Type: ACL; Schema: public; Owner: mew
--

GRANT ALL ON TABLE public."Shops" TO "shop-admin";


--
-- Name: SEQUENCE "Shops_id_seq"; Type: ACL; Schema: public; Owner: mew
--

GRANT SELECT,USAGE ON SEQUENCE public."Shops_id_seq" TO "shop-admin";


--
-- Name: TABLE "StreetNames"; Type: ACL; Schema: public; Owner: mew
--

GRANT ALL ON TABLE public."StreetNames" TO "shop-admin";


--
-- Name: SEQUENCE "StreetNames_id_seq"; Type: ACL; Schema: public; Owner: mew
--

GRANT SELECT,USAGE ON SEQUENCE public."StreetNames_id_seq" TO "shop-admin";


--
-- PostgreSQL database dump complete
--

