-- Downloaded from: https://github.com/mostafacs/ecommerce-microservices-spring-reactive-webflux/blob/919ad2b751d095a3679b445ebad27b77e42b0cef/storage/migration/src/main/resources/flyway/migrations/m20190805__Base_Tables.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 10.7 (Ubuntu 10.7-0ubuntu0.18.04.1)
-- Dumped by pg_dump version 11.2 (Ubuntu 11.2-1.pgdg18.04+1)

-- Started on 2019-07-29 04:35:46 EET

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 4 (class 2615 OID 24710)
-- Name: ecommerce; Type: SCHEMA; Schema: -; Owner: ecommerce
--
--
-- CREATE SCHEMA ecommerce;
--
--
-- ALTER SCHEMA ecommerce OWNER TO ecommerce;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- TOC entry 198 (class 1259 OID 16389)
-- Name: product_id_seq; Type: SEQUENCE; Schema: public; Owner: ecommerce
--

CREATE SEQUENCE ecommerce.product_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE ecommerce.product_id_seq OWNER TO ecommerce;

--
-- TOC entry 197 (class 1259 OID 16386)
-- Name: product; Type: TABLE; Schema: public; Owner: ecommerce
--

CREATE TABLE ecommerce.product (
    id bigint DEFAULT nextval('ecommerce.product_id_seq'::regclass) NOT NULL,
    title character varying(250),
    sku character varying(150),
    inventory_count integer
);


ALTER TABLE ecommerce.product OWNER TO ecommerce;

--
-- TOC entry 199 (class 1259 OID 24579)
-- Name: shopping_cart_id_seq; Type: SEQUENCE; Schema: public; Owner: ecommerce
--

CREATE SEQUENCE ecommerce.shopping_cart_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE ecommerce.shopping_cart_id_seq OWNER TO ecommerce;

--
-- TOC entry 201 (class 1259 OID 24583)
-- Name: shopping_cart; Type: TABLE; Schema: public; Owner: ecommerce
--

CREATE TABLE ecommerce.shopping_cart (
    id bigint DEFAULT nextval('ecommerce.shopping_cart_id_seq'::regclass) NOT NULL,
    total_quantity integer,
    sub_total_price numeric(18,5),
    total_shipping_cost numeric(18,0),
    total_cost numeric,
    user_id bigint
);


ALTER TABLE ecommerce.shopping_cart OWNER TO ecommerce;

--
-- TOC entry 200 (class 1259 OID 24581)
-- Name: shopping_cart_item_id_seq; Type: SEQUENCE; Schema: public; Owner: ecommerce
--

CREATE SEQUENCE ecommerce.shopping_cart_item_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE ecommerce.shopping_cart_item_id_seq OWNER TO ecommerce;

--
-- TOC entry 202 (class 1259 OID 24590)
-- Name: shopping_cart_item; Type: TABLE; Schema: public; Owner: ecommerce
--

CREATE TABLE ecommerce.shopping_cart_item (
    id bigint DEFAULT nextval('ecommerce.shopping_cart_item_id_seq'::regclass) NOT NULL,
    quantity integer,
    unit_price numeric,
    total_price numeric,
    shipping_cost numeric,
    product_id bigint,
    shopping_cart_id bigint
);


ALTER TABLE ecommerce.shopping_cart_item OWNER TO ecommerce;

--
-- TOC entry 204 (class 1259 OID 24671)
-- Name: user_id_seq; Type: SEQUENCE; Schema: public; Owner: ecommerce
--

CREATE SEQUENCE ecommerce.user_id_seq
    START WITH 0
    INCREMENT BY 1
    MINVALUE 0
    NO MAXVALUE
    CACHE 1;


ALTER TABLE ecommerce.user_id_seq OWNER TO ecommerce;

--
-- TOC entry 203 (class 1259 OID 24645)
-- Name: users; Type: TABLE; Schema: public; Owner: ecommerce
--

CREATE TABLE ecommerce.users (
    id bigint DEFAULT nextval('ecommerce.user_id_seq'::regclass) NOT NULL,
    first_name character varying(100),
    last_name character varying(100),
    create_date date,
    update_date date,
    last_login date,
    email character varying(100),
    password text
);


ALTER TABLE ecommerce.users OWNER TO ecommerce;

--
-- TOC entry 2969 (class 0 OID 0)
-- Dependencies: 198
-- Name: product_id_seq; Type: SEQUENCE SET; Schema: public; Owner: ecommerce
--

SELECT pg_catalog.setval('ecommerce.product_id_seq', 3, true);


--
-- TOC entry 2970 (class 0 OID 0)
-- Dependencies: 199
-- Name: shopping_cart_id_seq; Type: SEQUENCE SET; Schema: public; Owner: ecommerce
--

SELECT pg_catalog.setval('ecommerce.shopping_cart_id_seq', 24, true);


--
-- TOC entry 2971 (class 0 OID 0)
-- Dependencies: 200
-- Name: shopping_cart_item_id_seq; Type: SEQUENCE SET; Schema: public; Owner: ecommerce
--

SELECT pg_catalog.setval('ecommerce.shopping_cart_item_id_seq', 11, true);


--
-- TOC entry 2972 (class 0 OID 0)
-- Dependencies: 204
-- Name: user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: ecommerce
--

SELECT pg_catalog.setval('ecommerce.user_id_seq', 2, true);


--
-- TOC entry 2823 (class 2606 OID 24598)
-- Name: shopping_cart_item cart_item_pk; Type: CONSTRAINT; Schema: public; Owner: ecommerce
--

ALTER TABLE ONLY ecommerce.shopping_cart_item
    ADD CONSTRAINT cart_item_pk PRIMARY KEY (id);


--
-- TOC entry 2821 (class 2606 OID 24606)
-- Name: shopping_cart cart_pk; Type: CONSTRAINT; Schema: public; Owner: ecommerce
--

ALTER TABLE ONLY ecommerce.shopping_cart
    ADD CONSTRAINT cart_pk PRIMARY KEY (id);


--
-- TOC entry 2819 (class 2606 OID 16392)
-- Name: product product_pk; Type: CONSTRAINT; Schema: public; Owner: ecommerce
--

ALTER TABLE ONLY ecommerce.product
    ADD CONSTRAINT product_pk PRIMARY KEY (id);


--
-- TOC entry 2827 (class 2606 OID 24649)
-- Name: users user_pk; Type: CONSTRAINT; Schema: public; Owner: ecommerce
--

ALTER TABLE ONLY ecommerce.users
    ADD CONSTRAINT user_pk PRIMARY KEY (id);


--
-- TOC entry 2824 (class 1259 OID 24617)
-- Name: fki_cart_item_cart_fk; Type: INDEX; Schema: public; Owner: ecommerce
--

CREATE INDEX fki_cart_item_cart_fk ON ecommerce.shopping_cart_item USING btree (shopping_cart_id);


--
-- TOC entry 2825 (class 1259 OID 24604)
-- Name: fki_cart_item_product_fk; Type: INDEX; Schema: public; Owner: ecommerce
--

CREATE INDEX fki_cart_item_product_fk ON ecommerce.shopping_cart_item USING btree (product_id);


--
-- TOC entry 2833 (class 2606 OID 24612)
-- Name: shopping_cart_item cart_item_cart_fk; Type: FK CONSTRAINT; Schema: public; Owner: ecommerce
--

ALTER TABLE ONLY ecommerce.shopping_cart_item
    ADD CONSTRAINT cart_item_cart_fk FOREIGN KEY (shopping_cart_id) REFERENCES ecommerce.shopping_cart(id);


--
-- TOC entry 2832 (class 2606 OID 24599)
-- Name: shopping_cart_item cart_item_product_fk; Type: FK CONSTRAINT; Schema: public; Owner: ecommerce
--

ALTER TABLE ONLY ecommerce.shopping_cart_item
    ADD CONSTRAINT cart_item_product_fk FOREIGN KEY (product_id) REFERENCES ecommerce.product(id);


--
-- TOC entry 2831 (class 2606 OID 24650)
-- Name: shopping_cart user_fk; Type: FK CONSTRAINT; Schema: public; Owner: ecommerce
--

ALTER TABLE ONLY ecommerce.shopping_cart
    ADD CONSTRAINT user_fk FOREIGN KEY (user_id) REFERENCES ecommerce.users(id);


-- Completed on 2019-07-29 04:35:50 EET

--
-- PostgreSQL database dump complete
--

