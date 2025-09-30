-- Downloaded from: https://github.com/Dmitrytsg/onectest/blob/427db1e8ecee644c8f52ea16e7b9c608f6d3c375/onecPsqlDB.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 15.7 (Homebrew)
-- Dumped by pg_dump version 15.7 (Homebrew)

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
-- Name: update_statistics(); Type: FUNCTION; Schema: public; Owner: dmitry
--

CREATE FUNCTION public.update_statistics() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    order_date DATE;
    product_category INT;
BEGIN
    -- Извлекаем дату заказа и категорию товара
    order_date := DATE(NEW.order_time);
    SELECT category_id INTO product_category FROM products WHERE product_id = NEW.product_id;

    -- Проверяем, существует ли запись в статистике за данный день и категорию
    IF EXISTS (
        SELECT 1 FROM statistics
        WHERE date = order_date
          AND category_id = product_category
    ) THEN
        -- Если запись существует, обновляем количество товаров
        UPDATE statistics
        SET product_count = product_count + NEW.number
        WHERE date = order_date
          AND category_id = product_category;
    ELSE
        -- Если запись не существует, создаем новую
        INSERT INTO statistics (date, category_id, product_count)
        VALUES (order_date, product_category, NEW.number);
    END IF;

    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_statistics() OWNER TO dmitry;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: categories; Type: TABLE; Schema: public; Owner: dmitry
--

CREATE TABLE public.categories (
    category_id integer NOT NULL,
    category_name character varying(50) NOT NULL
);


ALTER TABLE public.categories OWNER TO dmitry;

--
-- Name: categories_category_id_seq; Type: SEQUENCE; Schema: public; Owner: dmitry
--

CREATE SEQUENCE public.categories_category_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.categories_category_id_seq OWNER TO dmitry;

--
-- Name: categories_category_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: dmitry
--

ALTER SEQUENCE public.categories_category_id_seq OWNED BY public.categories.category_id;


--
-- Name: orders; Type: TABLE; Schema: public; Owner: dmitry
--

CREATE TABLE public.orders (
    order_id integer NOT NULL,
    order_time timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    product_id integer NOT NULL,
    number integer NOT NULL
);


ALTER TABLE public.orders OWNER TO dmitry;

--
-- Name: orders_order_id_seq; Type: SEQUENCE; Schema: public; Owner: dmitry
--

CREATE SEQUENCE public.orders_order_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.orders_order_id_seq OWNER TO dmitry;

--
-- Name: orders_order_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: dmitry
--

ALTER SEQUENCE public.orders_order_id_seq OWNED BY public.orders.order_id;


--
-- Name: products; Type: TABLE; Schema: public; Owner: dmitry
--

CREATE TABLE public.products (
    product_id integer NOT NULL,
    product_name character varying(100) NOT NULL,
    category_id integer NOT NULL,
    price numeric(10,2) NOT NULL,
    description text
);


ALTER TABLE public.products OWNER TO dmitry;

--
-- Name: products_product_id_seq; Type: SEQUENCE; Schema: public; Owner: dmitry
--

CREATE SEQUENCE public.products_product_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.products_product_id_seq OWNER TO dmitry;

--
-- Name: products_product_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: dmitry
--

ALTER SEQUENCE public.products_product_id_seq OWNED BY public.products.product_id;


--
-- Name: statistics; Type: TABLE; Schema: public; Owner: dmitry
--

CREATE TABLE public.statistics (
    stat_id integer NOT NULL,
    date date NOT NULL,
    category_id integer NOT NULL,
    product_count integer NOT NULL
);


ALTER TABLE public.statistics OWNER TO dmitry;

--
-- Name: statistics_stat_id_seq; Type: SEQUENCE; Schema: public; Owner: dmitry
--

CREATE SEQUENCE public.statistics_stat_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.statistics_stat_id_seq OWNER TO dmitry;

--
-- Name: statistics_stat_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: dmitry
--

ALTER SEQUENCE public.statistics_stat_id_seq OWNED BY public.statistics.stat_id;


--
-- Name: categories category_id; Type: DEFAULT; Schema: public; Owner: dmitry
--

ALTER TABLE ONLY public.categories ALTER COLUMN category_id SET DEFAULT nextval('public.categories_category_id_seq'::regclass);


--
-- Name: orders order_id; Type: DEFAULT; Schema: public; Owner: dmitry
--

ALTER TABLE ONLY public.orders ALTER COLUMN order_id SET DEFAULT nextval('public.orders_order_id_seq'::regclass);


--
-- Name: products product_id; Type: DEFAULT; Schema: public; Owner: dmitry
--

ALTER TABLE ONLY public.products ALTER COLUMN product_id SET DEFAULT nextval('public.products_product_id_seq'::regclass);


--
-- Name: statistics stat_id; Type: DEFAULT; Schema: public; Owner: dmitry
--

ALTER TABLE ONLY public.statistics ALTER COLUMN stat_id SET DEFAULT nextval('public.statistics_stat_id_seq'::regclass);


--
-- Data for Name: categories; Type: TABLE DATA; Schema: public; Owner: dmitry
--

COPY public.categories (category_id, category_name) FROM stdin;
1	Electronics
2	Books
3	Clothing
\.


--
-- Data for Name: orders; Type: TABLE DATA; Schema: public; Owner: dmitry
--

COPY public.orders (order_id, order_time, product_id, number) FROM stdin;
1	2024-06-20 15:53:18.844677	15	3
2	2024-06-20 15:54:59.915415	14	7
3	2024-06-20 15:54:59.915415	18	4
4	2024-06-20 15:54:59.915415	22	10
5	2024-06-20 15:54:59.915415	21	6
\.


--
-- Data for Name: products; Type: TABLE DATA; Schema: public; Owner: dmitry
--

COPY public.products (product_id, product_name, category_id, price, description) FROM stdin;
13	Smartphone	1	599.99	Latest model with advanced features
14	Laptop	1	999.99	High performance laptop with 16GB RAM
15	Headphones	1	199.99	Noise-cancelling over-ear headphones
16	Tablet	1	299.99	10-inch tablet with high-resolution display
17	E-reader	2	129.99	Portable e-reader with backlit display
18	Novel	2	19.99	Bestselling fiction novel
19	Cookbook	2	24.99	Collection of gourmet recipes
20	Biography	2	29.99	Biography of a famous personality
21	T-shirt	3	14.99	Cotton T-shirt with a graphic print
22	Jeans	3	49.99	Denim jeans with a slim fit
23	Jacket	3	89.99	Waterproof outdoor jacket
24	Sneakers	3	74.99	Comfortable and stylish sneakers
\.


--
-- Data for Name: statistics; Type: TABLE DATA; Schema: public; Owner: dmitry
--

COPY public.statistics (stat_id, date, category_id, product_count) FROM stdin;
1	2024-06-20	1	10
2	2024-06-20	2	4
3	2024-06-20	3	16
\.


--
-- Name: categories_category_id_seq; Type: SEQUENCE SET; Schema: public; Owner: dmitry
--

SELECT pg_catalog.setval('public.categories_category_id_seq', 3, true);


--
-- Name: orders_order_id_seq; Type: SEQUENCE SET; Schema: public; Owner: dmitry
--

SELECT pg_catalog.setval('public.orders_order_id_seq', 5, true);


--
-- Name: products_product_id_seq; Type: SEQUENCE SET; Schema: public; Owner: dmitry
--

SELECT pg_catalog.setval('public.products_product_id_seq', 24, true);


--
-- Name: statistics_stat_id_seq; Type: SEQUENCE SET; Schema: public; Owner: dmitry
--

SELECT pg_catalog.setval('public.statistics_stat_id_seq', 3, true);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: dmitry
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (category_id);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: dmitry
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (order_id);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: dmitry
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (product_id);


--
-- Name: statistics statistics_pkey; Type: CONSTRAINT; Schema: public; Owner: dmitry
--

ALTER TABLE ONLY public.statistics
    ADD CONSTRAINT statistics_pkey PRIMARY KEY (stat_id);


--
-- Name: orders after_order_insert; Type: TRIGGER; Schema: public; Owner: dmitry
--

CREATE TRIGGER after_order_insert AFTER INSERT ON public.orders FOR EACH ROW EXECUTE FUNCTION public.update_statistics();


--
-- Name: products fk_category_id; Type: FK CONSTRAINT; Schema: public; Owner: dmitry
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT fk_category_id FOREIGN KEY (category_id) REFERENCES public.categories(category_id) ON UPDATE SET NULL ON DELETE SET NULL;


--
-- Name: statistics fk_category_id; Type: FK CONSTRAINT; Schema: public; Owner: dmitry
--

ALTER TABLE ONLY public.statistics
    ADD CONSTRAINT fk_category_id FOREIGN KEY (category_id) REFERENCES public.categories(category_id) ON UPDATE SET NULL ON DELETE SET NULL;


--
-- Name: orders fk_product_id; Type: FK CONSTRAINT; Schema: public; Owner: dmitry
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT fk_product_id FOREIGN KEY (product_id) REFERENCES public.products(product_id) ON UPDATE SET NULL ON DELETE SET NULL;


--
-- PostgreSQL database dump complete
--

