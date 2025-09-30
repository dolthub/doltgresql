-- Downloaded from: https://github.com/erlitx/sql_final/blob/f86f8e20939272b6a2da92409f522d399ac5eae3/backup.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 16.6 (Ubuntu 16.6-0ubuntu0.24.04.1)
-- Dumped by pg_dump version 16.6 (Ubuntu 16.6-0ubuntu0.24.04.1)

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
-- Name: calculate_total_sales(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.calculate_total_sales(product_id_input integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    total_sales INT;
BEGIN
    SELECT COALESCE(SUM(qty), 0) INTO total_sales
    FROM sale_order_line
    WHERE product_id = product_id_input;

    RETURN total_sales;
END;
$$;


ALTER FUNCTION public.calculate_total_sales(product_id_input integer) OWNER TO postgres;

--
-- Name: get_subordinates(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_subordinates(employee_id integer) RETURNS TABLE(subordinate_id integer, subordinate_name character varying)
    LANGUAGE plpgsql
    AS $$
BEGIN
	RETURN QUERY
    SELECT id, user_name
    FROM company_user
    WHERE report_to = employee_id;
END;
$$;


ALTER FUNCTION public.get_subordinates(employee_id integer) OWNER TO postgres;

--
-- Name: update_order_status(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_order_status() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Проверяем, отгружены ли все позиции заказа
    IF (
        SELECT COUNT(*) 
        FROM sale_order_line sol
        WHERE sol.sale_order_id = (
            SELECT sm.sale_order_id
            FROM stock_move sm
            WHERE sm.id = NEW.stock_move_id
        )
        AND NOT EXISTS (
            SELECT 1
            FROM stock_move_line sml
            WHERE sml.stock_move_id = NEW.stock_move_id
              AND sml.product_id = sol.product_id
              AND sml.qty >= sol.qty
        )
    ) = 0 THEN
        -- Если все позиции отгружены, обновляем статус заказа
        UPDATE sale_order
        SET status = 'Завершен'
        WHERE id = (
            SELECT sm.sale_order_id
            FROM stock_move sm
            WHERE sm.id = NEW.stock_move_id
        );
    END IF;

    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_order_status() OWNER TO postgres;

--
-- Name: validate_partner_data(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.validate_partner_data() RETURNS trigger
    LANGUAGE plpgsql
    AS $_$
BEGIN
    -- Проверяем формат email
    IF NEW.email IS NOT NULL AND NOT NEW.email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$' THEN
        RAISE EXCEPTION 'Некорректный email адрес: %', NEW.email;
    END IF;

    -- Проверяем формат телефона
    IF NEW.phone IS NOT NULL AND NOT NEW.phone ~ '^\+\d{1,3}-\d{3}-\d{3}-\d{4}$' THEN
        RAISE EXCEPTION 'Некорректный номер телефона: %', NEW.phone;
    END IF;

    RETURN NEW;
END;
$_$;


ALTER FUNCTION public.validate_partner_data() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: company_user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.company_user (
    id integer NOT NULL,
    user_name character varying(255) NOT NULL,
    report_to integer
);


ALTER TABLE public.company_user OWNER TO postgres;

--
-- Name: TABLE company_user; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.company_user IS 'Сотрудники компании';


--
-- Name: company_user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.company_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.company_user_id_seq OWNER TO postgres;

--
-- Name: company_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.company_user_id_seq OWNED BY public.company_user.id;


--
-- Name: location; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.location (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    type character varying(100),
    CONSTRAINT location_type_check CHECK (((type)::text = ANY ((ARRAY['Внутренни'::character varying, 'Покупатель'::character varying, 'Поставщик'::character varying])::text[])))
);


ALTER TABLE public.location OWNER TO postgres;

--
-- Name: TABLE location; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.location IS 'Склады, ';


--
-- Name: location_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.location_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.location_id_seq OWNER TO postgres;

--
-- Name: location_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.location_id_seq OWNED BY public.location.id;


--
-- Name: partner; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.partner (
    id integer NOT NULL,
    partner_name character varying(255) NOT NULL,
    email character varying(255),
    phone character varying(50),
    type character varying(20),
    CONSTRAINT partner_type_check CHECK (((type)::text = ANY ((ARRAY['физлицо'::character varying, 'юрлицо'::character varying])::text[])))
);


ALTER TABLE public.partner OWNER TO postgres;

--
-- Name: TABLE partner; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.partner IS 'Партнеры компании - покупатели';


--
-- Name: partner_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.partner_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.partner_id_seq OWNER TO postgres;

--
-- Name: partner_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.partner_id_seq OWNED BY public.partner.id;


--
-- Name: product; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.product (
    id integer NOT NULL,
    article character varying(50) NOT NULL,
    product_name character varying(255) NOT NULL,
    category_id integer,
    created_date timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.product OWNER TO postgres;

--
-- Name: TABLE product; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.product IS 'Содержит информацию о товарах';


--
-- Name: product_category; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.product_category (
    id integer NOT NULL,
    name character varying(255) NOT NULL
);


ALTER TABLE public.product_category OWNER TO postgres;

--
-- Name: TABLE product_category; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.product_category IS 'Категории товаров';


--
-- Name: product_category_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.product_category_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.product_category_id_seq OWNER TO postgres;

--
-- Name: product_category_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.product_category_id_seq OWNED BY public.product_category.id;


--
-- Name: product_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.product_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.product_id_seq OWNER TO postgres;

--
-- Name: product_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.product_id_seq OWNED BY public.product.id;


--
-- Name: sale_channel; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sale_channel (
    id integer NOT NULL,
    name character varying(255) NOT NULL
);


ALTER TABLE public.sale_channel OWNER TO postgres;

--
-- Name: TABLE sale_channel; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.sale_channel IS 'Каналы продаж';


--
-- Name: sale_channel_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.sale_channel_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sale_channel_id_seq OWNER TO postgres;

--
-- Name: sale_channel_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.sale_channel_id_seq OWNED BY public.sale_channel.id;


--
-- Name: sale_order; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sale_order (
    id integer NOT NULL,
    partner_id integer,
    user_id integer,
    sale_channel_id integer,
    amount numeric(10,2) NOT NULL,
    status character varying(50),
    created_date timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.sale_order OWNER TO postgres;

--
-- Name: TABLE sale_order; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.sale_order IS 'Заказы продаж';


--
-- Name: sale_order_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.sale_order_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sale_order_id_seq OWNER TO postgres;

--
-- Name: sale_order_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.sale_order_id_seq OWNED BY public.sale_order.id;


--
-- Name: sale_order_line; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sale_order_line (
    id integer NOT NULL,
    sale_order_id integer,
    product_id integer,
    qty integer NOT NULL,
    unit_price numeric(10,2) NOT NULL
);


ALTER TABLE public.sale_order_line OWNER TO postgres;

--
-- Name: TABLE sale_order_line; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.sale_order_line IS 'Товары в заказе';


--
-- Name: sale_order_line_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.sale_order_line_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sale_order_line_id_seq OWNER TO postgres;

--
-- Name: sale_order_line_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.sale_order_line_id_seq OWNED BY public.sale_order_line.id;


--
-- Name: stock_move; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.stock_move (
    id integer NOT NULL,
    sale_order_id integer,
    "timestamp" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    user_id integer,
    from_location integer,
    to_location integer
);


ALTER TABLE public.stock_move OWNER TO postgres;

--
-- Name: TABLE stock_move; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.stock_move IS 'Отгрузки товаров';


--
-- Name: stock_move_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.stock_move_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.stock_move_id_seq OWNER TO postgres;

--
-- Name: stock_move_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.stock_move_id_seq OWNED BY public.stock_move.id;


--
-- Name: stock_move_line; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.stock_move_line (
    id integer NOT NULL,
    stock_move_id integer,
    product_id integer,
    qty integer NOT NULL
);


ALTER TABLE public.stock_move_line OWNER TO postgres;

--
-- Name: TABLE stock_move_line; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.stock_move_line IS 'Товарные позиции в перемещениях';


--
-- Name: stock_move_line_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.stock_move_line_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.stock_move_line_id_seq OWNER TO postgres;

--
-- Name: stock_move_line_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.stock_move_line_id_seq OWNED BY public.stock_move_line.id;


--
-- Name: company_user id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.company_user ALTER COLUMN id SET DEFAULT nextval('public.company_user_id_seq'::regclass);


--
-- Name: location id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.location ALTER COLUMN id SET DEFAULT nextval('public.location_id_seq'::regclass);


--
-- Name: partner id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.partner ALTER COLUMN id SET DEFAULT nextval('public.partner_id_seq'::regclass);


--
-- Name: product id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product ALTER COLUMN id SET DEFAULT nextval('public.product_id_seq'::regclass);


--
-- Name: product_category id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_category ALTER COLUMN id SET DEFAULT nextval('public.product_category_id_seq'::regclass);


--
-- Name: sale_channel id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sale_channel ALTER COLUMN id SET DEFAULT nextval('public.sale_channel_id_seq'::regclass);


--
-- Name: sale_order id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sale_order ALTER COLUMN id SET DEFAULT nextval('public.sale_order_id_seq'::regclass);


--
-- Name: sale_order_line id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sale_order_line ALTER COLUMN id SET DEFAULT nextval('public.sale_order_line_id_seq'::regclass);


--
-- Name: stock_move id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_move ALTER COLUMN id SET DEFAULT nextval('public.stock_move_id_seq'::regclass);


--
-- Name: stock_move_line id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_move_line ALTER COLUMN id SET DEFAULT nextval('public.stock_move_line_id_seq'::regclass);


--
-- Data for Name: company_user; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.company_user (id, user_name, report_to) FROM stdin;
1	Иван Иванов	\N
2	Пётр Петров	1
3	Александр Сидоров	1
4	Мария Смирнова	2
5	Екатерина Орлова	2
6	Дмитрий Кузнецов	3
7	Ольга Ковалева	3
8	Сергей Васильев	4
9	Наталья Фёдорова	4
10	Анна Семёнова	5
\.


--
-- Data for Name: location; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.location (id, name, type) FROM stdin;
1	Склад Москва	Внутренни
2	Склад Спб	Внутренни
3	Склад Покупателя А	Покупатель
4	Склад Поставщика Б	Поставщик
\.


--
-- Data for Name: partner; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.partner (id, partner_name, email, phone, type) FROM stdin;
1	ООО Альфа	info@alfa.com	+7-900-123-4567	юрлицо
2	ИП Петров	petrov@business.ru	+7-901-234-5678	физлицо
3	ЗАО Бета	contact@beta.org	+7-902-345-6789	юрлицо
4	Иван Иванов	ivanov@mail.ru	+7-903-456-7890	физлицо
5	Мария Смирнова	m.smirnova@mail.ru	+7-904-567-8901	физлицо
6	ООО Гамма	sales@gamma.com	+7-905-678-9012	юрлицо
7	ИП Сидоров	sidrov@shop.biz	+7-906-789-0123	физлицо
8	ЗАО Дельта	info@delta.com	+7-907-890-1234	юрлицо
9	ООО Эпсилон	contact@epsilon.ru	+7-908-901-2345	юрлицо
10	ИП Орлова	orlova@mail.com	+7-909-012-3456	физлицо
\.


--
-- Data for Name: product; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.product (id, article, product_name, category_id, created_date) FROM stdin;
1	A001	Смартфон X	1	2024-12-23 08:07:30.430187
2	A002	Ноутбук Y	1	2024-12-23 08:07:30.430187
3	B001	Холодильник Z	2	2024-12-23 08:07:30.430187
4	C001	Куртка зимняя	3	2024-12-23 08:07:30.430187
5	C002	Футболка летняя	3	2024-12-23 08:07:30.430187
6	D001	Кроссовки спортивные	4	2024-12-23 08:07:30.430187
7	E001	Шкаф деревянный	5	2024-12-23 08:07:30.430187
8	F001	Машинка для детей	6	2024-12-23 08:07:30.430187
9	G001	Велосипед горный	7	2024-12-23 08:07:30.430187
10	H001	Тетрадь А4	8	2024-12-23 08:07:30.430187
\.


--
-- Data for Name: product_category; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.product_category (id, name) FROM stdin;
1	Электроника
2	Бытовая техника
3	Одежда
4	Обувь
5	Мебель
6	Игрушки
7	Спорттовары
8	Канцелярия
9	Продукты питания
10	Косметика
\.


--
-- Data for Name: sale_channel; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sale_channel (id, name) FROM stdin;
1	онлайн-магазин
2	WB
3	OZON
4	оффлайн-магазин
\.


--
-- Data for Name: sale_order; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sale_order (id, partner_id, user_id, sale_channel_id, amount, status, created_date) FROM stdin;
1	1	2	1	50000.00	Открыт	2024-01-01 10:00:00
2	2	3	2	12000.00	Завершен	2024-01-02 11:30:00
3	3	4	3	34000.00	Ожидает оплаты	2024-01-03 12:45:00
4	4	5	4	2700.00	Отменен	2024-01-04 14:15:00
5	5	6	1	8000.00	Завершен	2024-01-05 09:00:00
6	6	7	2	9500.00	Открыт	2024-01-06 16:20:00
7	7	8	3	5600.00	Завершен	2024-01-07 13:40:00
8	8	9	4	12300.00	Открыт	2024-01-08 17:00:00
9	9	10	1	7800.00	Ожидает отгрузки	2024-01-09 10:30:00
10	10	1	2	43000.00	Завершен	2024-01-10 15:45:00
\.


--
-- Data for Name: sale_order_line; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sale_order_line (id, sale_order_id, product_id, qty, unit_price) FROM stdin;
1	1	1	2	25000.00
2	2	2	1	12000.00
3	3	3	1	34000.00
4	4	4	1	2700.00
5	5	5	4	2000.00
6	6	6	2	4750.00
7	7	7	1	5600.00
8	8	8	2	6150.00
9	9	9	3	2600.00
10	10	10	10	4300.00
\.


--
-- Data for Name: stock_move; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.stock_move (id, sale_order_id, "timestamp", user_id, from_location, to_location) FROM stdin;
1	1	2024-01-01 15:00:00	2	4	1
2	2	2024-01-02 16:00:00	3	4	1
3	3	2024-01-03 17:00:00	4	4	1
4	4	2024-01-04 18:00:00	5	4	1
5	5	2024-01-05 19:00:00	6	4	1
6	6	2024-01-06 20:00:00	7	4	1
7	7	2024-01-07 21:00:00	8	4	1
8	8	2024-01-08 22:00:00	9	4	1
9	9	2024-01-09 23:00:00	10	4	1
10	10	2024-01-10 08:00:00	1	4	1
11	\N	2024-01-11 08:00:00	1	1	3
\.


--
-- Data for Name: stock_move_line; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.stock_move_line (id, stock_move_id, product_id, qty) FROM stdin;
1	1	1	2
2	2	2	1
3	3	3	1
4	4	4	1
5	5	5	4
6	6	6	2
7	7	7	1
8	8	8	2
9	9	9	3
10	10	10	10
11	11	10	3
\.


--
-- Name: company_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.company_user_id_seq', 1, false);


--
-- Name: location_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.location_id_seq', 4, true);


--
-- Name: partner_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.partner_id_seq', 10, true);


--
-- Name: product_category_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.product_category_id_seq', 10, true);


--
-- Name: product_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.product_id_seq', 10, true);


--
-- Name: sale_channel_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.sale_channel_id_seq', 4, true);


--
-- Name: sale_order_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.sale_order_id_seq', 10, true);


--
-- Name: sale_order_line_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.sale_order_line_id_seq', 10, true);


--
-- Name: stock_move_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.stock_move_id_seq', 11, true);


--
-- Name: stock_move_line_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.stock_move_line_id_seq', 11, true);


--
-- Name: company_user company_user_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.company_user
    ADD CONSTRAINT company_user_pkey PRIMARY KEY (id);


--
-- Name: location location_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.location
    ADD CONSTRAINT location_pkey PRIMARY KEY (id);


--
-- Name: partner partner_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.partner
    ADD CONSTRAINT partner_pkey PRIMARY KEY (id);


--
-- Name: product product_article_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product
    ADD CONSTRAINT product_article_key UNIQUE (article);


--
-- Name: product_category product_category_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_category
    ADD CONSTRAINT product_category_pkey PRIMARY KEY (id);


--
-- Name: product product_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product
    ADD CONSTRAINT product_pkey PRIMARY KEY (id);


--
-- Name: sale_channel sale_channel_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sale_channel
    ADD CONSTRAINT sale_channel_pkey PRIMARY KEY (id);


--
-- Name: sale_order_line sale_order_line_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sale_order_line
    ADD CONSTRAINT sale_order_line_pkey PRIMARY KEY (id);


--
-- Name: sale_order sale_order_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sale_order
    ADD CONSTRAINT sale_order_pkey PRIMARY KEY (id);


--
-- Name: stock_move_line stock_move_line_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_move_line
    ADD CONSTRAINT stock_move_line_pkey PRIMARY KEY (id);


--
-- Name: stock_move stock_move_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_move
    ADD CONSTRAINT stock_move_pkey PRIMARY KEY (id);


--
-- Name: stock_move_line after_stock_move_line_insert; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER after_stock_move_line_insert AFTER INSERT ON public.stock_move_line FOR EACH ROW EXECUTE FUNCTION public.update_order_status();


--
-- Name: partner trigger_validate_partner_data; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trigger_validate_partner_data BEFORE INSERT OR UPDATE ON public.partner FOR EACH ROW EXECUTE FUNCTION public.validate_partner_data();


--
-- Name: company_user company_user_report_to_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.company_user
    ADD CONSTRAINT company_user_report_to_fkey FOREIGN KEY (report_to) REFERENCES public.company_user(id);


--
-- Name: product product_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product
    ADD CONSTRAINT product_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.product_category(id);


--
-- Name: sale_order_line sale_order_line_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sale_order_line
    ADD CONSTRAINT sale_order_line_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.product(id);


--
-- Name: sale_order_line sale_order_line_sale_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sale_order_line
    ADD CONSTRAINT sale_order_line_sale_order_id_fkey FOREIGN KEY (sale_order_id) REFERENCES public.sale_order(id);


--
-- Name: sale_order sale_order_partner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sale_order
    ADD CONSTRAINT sale_order_partner_id_fkey FOREIGN KEY (partner_id) REFERENCES public.partner(id);


--
-- Name: sale_order sale_order_sale_channel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sale_order
    ADD CONSTRAINT sale_order_sale_channel_id_fkey FOREIGN KEY (sale_channel_id) REFERENCES public.sale_channel(id);


--
-- Name: sale_order sale_order_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sale_order
    ADD CONSTRAINT sale_order_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.company_user(id);


--
-- Name: stock_move stock_move_from_location_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_move
    ADD CONSTRAINT stock_move_from_location_fkey FOREIGN KEY (from_location) REFERENCES public.location(id);


--
-- Name: stock_move_line stock_move_line_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_move_line
    ADD CONSTRAINT stock_move_line_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.product(id);


--
-- Name: stock_move_line stock_move_line_stock_move_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_move_line
    ADD CONSTRAINT stock_move_line_stock_move_id_fkey FOREIGN KEY (stock_move_id) REFERENCES public.stock_move(id);


--
-- Name: stock_move stock_move_sale_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_move
    ADD CONSTRAINT stock_move_sale_order_id_fkey FOREIGN KEY (sale_order_id) REFERENCES public.sale_order(id);


--
-- Name: stock_move stock_move_to_location_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_move
    ADD CONSTRAINT stock_move_to_location_fkey FOREIGN KEY (to_location) REFERENCES public.location(id);


--
-- Name: stock_move stock_move_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_move
    ADD CONSTRAINT stock_move_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.company_user(id);


--
-- PostgreSQL database dump complete
--

