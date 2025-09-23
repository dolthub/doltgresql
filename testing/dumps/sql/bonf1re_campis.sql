-- Downloaded from: https://github.com/bonf1re/campis/blob/2711087f743616e5fc496e68a2c83f517e6c17c9/db/create.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 9.6.5
-- Dumped by pg_dump version 9.6.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: postgres; Type: COMMENT; Schema: -; Owner: postgres
--

COMMENT ON DATABASE postgres IS 'default administrative connection database';


--
-- Name: campis; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA campis;


ALTER SCHEMA campis OWNER TO postgres;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

--
-- Name: create_new_relations(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION create_new_relations() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
    BEGIN
        INSERT INTO campis.productxwarehouse (id_product, id_warehouse)
        SELECT p.id_product, NEW.id_warehouse
        FROM campis.product p;
        RETURN NEW;
    END;
$$;


ALTER FUNCTION public.create_new_relations() OWNER TO postgres;

--
-- Name: create_relations(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION create_relations() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
    BEGIN
        INSERT INTO campis.productxwarehouse (id_product, id_warehouse)
        SELECT NEW.id_product, w.id_warehouse
        FROM campis.warehouse w;
        RETURN NEW;
    END;
$$;


ALTER FUNCTION public.create_relations() OWNER TO postgres;

--
-- Name: stock_update(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION stock_update() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
    DECLARE
    productID   integer;
    BEGIN
        SELECT id_product into productID from campis.batch where id_batch = NEW.id_batch;
        IF NEW.mov_type < 2 THEN
            UPDATE campis.productxwarehouse SET p_stock = p_stock + NEW.quantity, c_stock = c_stock + NEW.quantity 
            WHERE id_product = productID AND id_warehouse = NEW.id_warehouse;
        ELSIF NEW.mov_type = 3 THEN
            UPDATE campis.productxwarehouse SET p_stock = p_stock - NEW.quantity 
            WHERE id_product = productID AND id_warehouse = NEW.id_warehouse;
        ELSIF NEW.mov_type = 4 THEN
            UPDATE campis.productxwarehouse SET p_stock = p_stock - NEW.quantity, c_stock = c_stock - NEW.quantity 
            WHERE id_product = productID AND id_warehouse = NEW.id_warehouse;
        

        END IF;
        RETURN NEW;
    END;
$$;


ALTER FUNCTION public.stock_update() OWNER TO postgres;

--
-- Name: stock_update_sale(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION stock_update_sale() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
    BEGIN
        UPDATE campis.productxwarehouse SET c_stock = c_stock - NEW.quantity 
        WHERE id_product = NEW.id_product AND id_warehouse = 1;
    RETURN NEW;
    END;
$$;


ALTER FUNCTION public.stock_update_sale() OWNER TO postgres;

SET search_path = campis, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: area; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE area (
    id_area integer NOT NULL,
    name character varying NOT NULL,
    length integer NOT NULL,
    width integer NOT NULL,
    pos_x integer NOT NULL,
    pos_y integer NOT NULL,
    id_warehouse integer NOT NULL,
    product_type integer
);


ALTER TABLE area OWNER TO postgres;

--
-- Name: area_id_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE area_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE area_id_seq OWNER TO postgres;

--
-- Name: area_id_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE area_id_seq OWNED BY area.id_area;


--
-- Name: batch; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE batch (
    id_batch integer NOT NULL,
    quantity integer NOT NULL,
    batch_cost double precision,
    arrival_date timestamp without time zone DEFAULT now() NOT NULL,
    expiration_date timestamp without time zone,
    id_product integer NOT NULL,
    type_batch integer DEFAULT 1,
    state boolean,
    id_unit integer,
    heritage character varying,
    location character varying(255)
);


ALTER TABLE batch OWNER TO postgres;

--
-- Name: batch_id_batch_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE batch_id_batch_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE batch_id_batch_seq OWNER TO postgres;

--
-- Name: batch_id_batch_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE batch_id_batch_seq OWNED BY batch.id_batch;


--
-- Name: campaign; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE campaign (
    id_campaign integer NOT NULL,
    name character varying,
    description character varying,
    initial_date timestamp without time zone DEFAULT now(),
    final_date timestamp without time zone DEFAULT now()
);


ALTER TABLE campaign OWNER TO postgres;

--
-- Name: campaign_id_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE campaign_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE campaign_id_seq OWNER TO postgres;

--
-- Name: campaign_id_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE campaign_id_seq OWNED BY campaign.id_campaign;


--
-- Name: client; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE client (
    id_client integer NOT NULL,
    name character varying NOT NULL,
    dni character varying,
    ruc character varying,
    active boolean DEFAULT true NOT NULL,
    address character varying,
    phone character varying,
    email character varying,
    id_district integer
);


ALTER TABLE client OWNER TO postgres;

--
-- Name: client_id_client_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE client_id_client_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE client_id_client_seq OWNER TO postgres;

--
-- Name: client_id_client_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE client_id_client_seq OWNED BY client.id_client;


--
-- Name: complaint; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE complaint (
    id_complaint integer NOT NULL,
    description character varying NOT NULL,
    status character varying,
    id_request_order integer
);


ALTER TABLE complaint OWNER TO postgres;

--
-- Name: disclaim_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE disclaim_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE disclaim_seq OWNER TO postgres;

--
-- Name: disclaim_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE disclaim_seq OWNED BY complaint.id_complaint;


--
-- Name: group_batch_id_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE group_batch_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE group_batch_id_seq OWNER TO postgres;

--
-- Name: dispatch_move; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE dispatch_move (
    id_dispatch_move integer DEFAULT nextval('group_batch_id_seq'::regclass) NOT NULL,
    type_owner integer,
    id_owner integer,
    mov_date timestamp without time zone DEFAULT now() NOT NULL,
    reason integer,
    id_batch integer,
    arrival_date timestamp without time zone
);


ALTER TABLE dispatch_move OWNER TO postgres;

--
-- Name: dispatch_order; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE dispatch_order (
    id_dispatch_order integer NOT NULL,
    id_request_order integer NOT NULL,
    priority integer DEFAULT 1 NOT NULL,
    status character varying,
    id_prod integer,
    quantity integer
);


ALTER TABLE dispatch_order OWNER TO postgres;

--
-- Name: dispatch_order_id_dispatch_order_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE dispatch_order_id_dispatch_order_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE dispatch_order_id_dispatch_order_seq OWNER TO postgres;

--
-- Name: dispatch_order_id_dispatch_order_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE dispatch_order_id_dispatch_order_seq OWNED BY dispatch_order.id_dispatch_order;


--
-- Name: dispatch_order_line; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE dispatch_order_line (
    id_dispatch_order_line integer NOT NULL,
    id_dispatch_order integer NOT NULL,
    id_product integer NOT NULL,
    quantity integer,
    delivered boolean DEFAULT false
);


ALTER TABLE dispatch_order_line OWNER TO postgres;

--
-- Name: dispatch_order_line_id_dispatch_order_line_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE dispatch_order_line_id_dispatch_order_line_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE dispatch_order_line_id_dispatch_order_line_seq OWNER TO postgres;

--
-- Name: dispatch_order_line_id_dispatch_order_line_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE dispatch_order_line_id_dispatch_order_line_seq OWNED BY dispatch_order_line.id_dispatch_order_line;


--
-- Name: district; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE district (
    id_district integer NOT NULL,
    name character varying,
    freight double precision
);


ALTER TABLE district OWNER TO postgres;

--
-- Name: district_id_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE district_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE district_id_seq OWNER TO postgres;

--
-- Name: district_id_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE district_id_seq OWNED BY district.id_district;


--
-- Name: document; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE document (
    id_document integer NOT NULL,
    id_request_order integer NOT NULL,
    doc_type character varying NOT NULL,
    total_amount character varying NOT NULL
);


ALTER TABLE document OWNER TO postgres;

--
-- Name: document_id_document_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE document_id_document_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE document_id_document_seq OWNER TO postgres;

--
-- Name: document_id_document_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE document_id_document_seq OWNED BY document.id_document;


--
-- Name: group_batch; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE group_batch (
    id_group_batch integer NOT NULL,
    arrival_date timestamp without time zone,
    id_owner integer,
    reason integer,
    type_owner integer
);


ALTER TABLE group_batch OWNER TO postgres;

--
-- Name: hibernate_sequence; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE hibernate_sequence
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE hibernate_sequence OWNER TO postgres;

--
-- Name: movement; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE movement (
    id_movement integer NOT NULL,
    mov_date timestamp without time zone DEFAULT now() NOT NULL,
    id_user integer NOT NULL,
    quantity integer NOT NULL,
    id_vehicle integer NOT NULL,
    mov_type integer,
    id_warehouse integer,
    id_zone integer,
    id_batch integer
);


ALTER TABLE movement OWNER TO postgres;

--
-- Name: movement_id_movement_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE movement_id_movement_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE movement_id_movement_seq OWNER TO postgres;

--
-- Name: movement_id_movement_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE movement_id_movement_seq OWNED BY movement.id_movement;


--
-- Name: movementxbatch; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE movementxbatch (
    id_batch integer NOT NULL,
    id_movement integer NOT NULL
);


ALTER TABLE movementxbatch OWNER TO postgres;

--
-- Name: movementxdispatch; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE movementxdispatch (
    id_movement integer NOT NULL,
    id_dispatch_order_line integer NOT NULL
);


ALTER TABLE movementxdispatch OWNER TO postgres;

--
-- Name: movementxdocument; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE movementxdocument (
    id_document integer NOT NULL,
    id_movement integer NOT NULL
);


ALTER TABLE movementxdocument OWNER TO postgres;

--
-- Name: parameters; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE parameters (
    igv double precision,
    dollar double precision,
    pound double precision,
    euro double precision
);


ALTER TABLE parameters OWNER TO postgres;

--
-- Name: permission_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE permission_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE permission_seq OWNER TO postgres;

--
-- Name: permission; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE permission (
    id_view integer NOT NULL,
    id_role integer NOT NULL,
    visualize boolean DEFAULT false NOT NULL,
    modify boolean DEFAULT false NOT NULL,
    id_permission integer DEFAULT nextval('permission_seq'::regclass) NOT NULL
);


ALTER TABLE permission OWNER TO postgres;

--
-- Name: product; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE product (
    id_product integer NOT NULL,
    name character varying NOT NULL,
    description character varying NOT NULL,
    p_stock integer DEFAULT 0 NOT NULL,
    c_stock integer DEFAULT 0 NOT NULL,
    weight double precision DEFAULT 0 NOT NULL,
    trademark character varying,
    base_price double precision NOT NULL,
    id_unit_of_measure integer NOT NULL,
    id_product_type integer NOT NULL,
    max_qt integer DEFAULT 40 NOT NULL,
    min_stock integer DEFAULT 0 NOT NULL
);


ALTER TABLE product OWNER TO postgres;

--
-- Name: product_id_product_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE product_id_product_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE product_id_product_seq OWNER TO postgres;

--
-- Name: product_id_product_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE product_id_product_seq OWNED BY product.id_product;


--
-- Name: product_type; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE product_type (
    id_product_type integer NOT NULL,
    description character varying NOT NULL
);


ALTER TABLE product_type OWNER TO postgres;

--
-- Name: product_type_id_product_type_seq_1; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE product_type_id_product_type_seq_1
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE product_type_id_product_type_seq_1 OWNER TO postgres;

--
-- Name: product_type_id_product_type_seq_1; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE product_type_id_product_type_seq_1 OWNED BY product_type.id_product_type;


--
-- Name: productxwarehouse; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE productxwarehouse (
    id_product integer NOT NULL,
    id_warehouse integer NOT NULL,
    p_stock integer DEFAULT 0,
    c_stock integer DEFAULT 0
);


ALTER TABLE productxwarehouse OWNER TO postgres;

--
-- Name: rack; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE rack (
    id_rack integer NOT NULL,
    id_warehouse integer NOT NULL,
    pos_x integer NOT NULL,
    pos_y integer NOT NULL,
    n_columns integer NOT NULL,
    n_floors integer NOT NULL,
    orientation integer
);


ALTER TABLE rack OWNER TO postgres;

--
-- Name: rack_id_rack_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE rack_id_rack_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE rack_id_rack_seq OWNER TO postgres;

--
-- Name: rack_id_rack_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE rack_id_rack_seq OWNED BY rack.id_rack;


--
-- Name: refund; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE refund (
    id_refund integer NOT NULL,
    id_complaint integer,
    status character varying,
    creation_date timestamp without time zone DEFAULT now(),
    type_refund character varying
);


ALTER TABLE refund OWNER TO postgres;

--
-- Name: refund_line_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE refund_line_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE refund_line_seq OWNER TO postgres;

--
-- Name: refund_line; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE refund_line (
    id_refund_line integer DEFAULT nextval('refund_line_seq'::regclass) NOT NULL,
    id_refund integer,
    quantity integer,
    id_request_order_line integer
);


ALTER TABLE refund_line OWNER TO postgres;

--
-- Name: refund_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE refund_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE refund_seq OWNER TO postgres;

--
-- Name: refund_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE refund_seq OWNED BY refund.id_refund;


--
-- Name: request_order; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE request_order (
    id_request_order integer NOT NULL,
    creation_date timestamp without time zone DEFAULT now() NOT NULL,
    delivery_date timestamp without time zone,
    base_amount double precision,
    total_amount double precision,
    status character varying,
    id_client integer NOT NULL,
    priority integer,
    id_district integer,
    address character varying
);


ALTER TABLE request_order OWNER TO postgres;

--
-- Name: request_order_id_request_order_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE request_order_id_request_order_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE request_order_id_request_order_seq OWNER TO postgres;

--
-- Name: request_order_id_request_order_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE request_order_id_request_order_seq OWNED BY request_order.id_request_order;


--
-- Name: request_order_line; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE request_order_line (
    id_request_order_line integer NOT NULL,
    quantity integer,
    cost double precision,
    id_request_order integer NOT NULL,
    id_product integer NOT NULL
);


ALTER TABLE request_order_line OWNER TO postgres;

--
-- Name: request_order_line_id_request_order_line_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE request_order_line_id_request_order_line_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE request_order_line_id_request_order_line_seq OWNER TO postgres;

--
-- Name: request_order_line_id_request_order_line_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE request_order_line_id_request_order_line_seq OWNED BY request_order_line.id_request_order_line;


--
-- Name: request_status; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE request_status (
    id_request_status integer,
    description character varying,
    name character varying(255)
);


ALTER TABLE request_status OWNER TO postgres;

--
-- Name: role; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE role (
    id_role integer NOT NULL,
    description character varying NOT NULL
);


ALTER TABLE role OWNER TO postgres;

--
-- Name: role_id_role_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE role_id_role_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE role_id_role_seq OWNER TO postgres;

--
-- Name: role_id_role_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE role_id_role_seq OWNED BY role.id_role;


--
-- Name: sale_condition; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE sale_condition (
    id_sale_condition integer NOT NULL,
    initial_date timestamp without time zone,
    final_date timestamp without time zone,
    amount double precision,
    id_sale_condition_type character varying,
    limits integer,
    id_to_take integer,
    id_campaign integer,
    n_discount integer DEFAULT 1,
    n_tocount integer DEFAULT 1
);


ALTER TABLE sale_condition OWNER TO postgres;

--
-- Name: sale_condition_id_sale_condition_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE sale_condition_id_sale_condition_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sale_condition_id_sale_condition_seq OWNER TO postgres;

--
-- Name: sale_condition_id_sale_condition_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE sale_condition_id_sale_condition_seq OWNED BY sale_condition.id_sale_condition;


--
-- Name: sale_condition_type; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE sale_condition_type (
    id_sale_condition_type integer NOT NULL,
    description character varying NOT NULL
);


ALTER TABLE sale_condition_type OWNER TO postgres;

--
-- Name: sale_condition_type_id_sale_condition_type_seq_1; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE sale_condition_type_id_sale_condition_type_seq_1
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sale_condition_type_id_sale_condition_type_seq_1 OWNER TO postgres;

--
-- Name: sale_condition_type_id_sale_condition_type_seq_1; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE sale_condition_type_id_sale_condition_type_seq_1 OWNED BY sale_condition_type.id_sale_condition_type;


--
-- Name: unit_of_measure; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE unit_of_measure (
    id_unit_of_measure integer NOT NULL,
    description character varying NOT NULL,
    descrip character varying(255)
);


ALTER TABLE unit_of_measure OWNER TO postgres;

--
-- Name: unit_of_measure_id_unit_of_measure_seq_1; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE unit_of_measure_id_unit_of_measure_seq_1
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE unit_of_measure_id_unit_of_measure_seq_1 OWNER TO postgres;

--
-- Name: unit_of_measure_id_unit_of_measure_seq_1; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE unit_of_measure_id_unit_of_measure_seq_1 OWNED BY unit_of_measure.id_unit_of_measure;


--
-- Name: users; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE users (
    id_user integer NOT NULL,
    firstname character varying NOT NULL,
    lastname character varying NOT NULL,
    password character varying NOT NULL,
    email character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    active boolean DEFAULT true NOT NULL,
    id_role integer NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    username character varying DEFAULT 'johnny1'::character varying NOT NULL
);


ALTER TABLE users OWNER TO postgres;

--
-- Name: users_id_user_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE users_id_user_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE users_id_user_seq OWNER TO postgres;

--
-- Name: users_id_user_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE users_id_user_seq OWNED BY users.id_user;


--
-- Name: vehicle; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE vehicle (
    id_vehicle integer NOT NULL,
    max_weight double precision NOT NULL,
    speed integer,
    active boolean DEFAULT true NOT NULL,
    id_warehouse integer NOT NULL,
    plate character varying
);


ALTER TABLE vehicle OWNER TO postgres;

--
-- Name: vehicle_id_vehicle_seq_1; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE vehicle_id_vehicle_seq_1
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE vehicle_id_vehicle_seq_1 OWNER TO postgres;

--
-- Name: vehicle_id_vehicle_seq_1; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE vehicle_id_vehicle_seq_1 OWNED BY vehicle.id_vehicle;


--
-- Name: view; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE view (
    id_view integer NOT NULL,
    description character varying NOT NULL
);


ALTER TABLE view OWNER TO postgres;

--
-- Name: view_id_view_seq_1; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE view_id_view_seq_1
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE view_id_view_seq_1 OWNER TO postgres;

--
-- Name: view_id_view_seq_1; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE view_id_view_seq_1 OWNED BY view.id_view;


--
-- Name: warehouse; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE warehouse (
    id_warehouse integer NOT NULL,
    name character varying NOT NULL,
    length integer NOT NULL,
    width integer NOT NULL,
    status boolean DEFAULT true NOT NULL
);


ALTER TABLE warehouse OWNER TO postgres;

--
-- Name: warehouse_id_warehouse_seq; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE warehouse_id_warehouse_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE warehouse_id_warehouse_seq OWNER TO postgres;

--
-- Name: warehouse_id_warehouse_seq; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE warehouse_id_warehouse_seq OWNED BY warehouse.id_warehouse;


--
-- Name: zone; Type: TABLE; Schema: campis; Owner: postgres
--

CREATE TABLE zone (
    id_zone integer NOT NULL,
    id_warehouse integer NOT NULL,
    id_rack integer NOT NULL,
    pos_x integer NOT NULL,
    pos_y integer NOT NULL,
    pos_z integer NOT NULL,
    free boolean DEFAULT true NOT NULL
);


ALTER TABLE zone OWNER TO postgres;

--
-- Name: zone_id_zone_seq_1; Type: SEQUENCE; Schema: campis; Owner: postgres
--

CREATE SEQUENCE zone_id_zone_seq_1
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE zone_id_zone_seq_1 OWNER TO postgres;

--
-- Name: zone_id_zone_seq_1; Type: SEQUENCE OWNED BY; Schema: campis; Owner: postgres
--

ALTER SEQUENCE zone_id_zone_seq_1 OWNED BY zone.id_zone;


--
-- Name: area id_area; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY area ALTER COLUMN id_area SET DEFAULT nextval('area_id_seq'::regclass);


--
-- Name: batch id_batch; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY batch ALTER COLUMN id_batch SET DEFAULT nextval('batch_id_batch_seq'::regclass);


--
-- Name: campaign id_campaign; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY campaign ALTER COLUMN id_campaign SET DEFAULT nextval('campaign_id_seq'::regclass);


--
-- Name: client id_client; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY client ALTER COLUMN id_client SET DEFAULT nextval('client_id_client_seq'::regclass);


--
-- Name: complaint id_complaint; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY complaint ALTER COLUMN id_complaint SET DEFAULT nextval('disclaim_seq'::regclass);


--
-- Name: dispatch_order id_dispatch_order; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY dispatch_order ALTER COLUMN id_dispatch_order SET DEFAULT nextval('dispatch_order_id_dispatch_order_seq'::regclass);


--
-- Name: dispatch_order_line id_dispatch_order_line; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY dispatch_order_line ALTER COLUMN id_dispatch_order_line SET DEFAULT nextval('dispatch_order_line_id_dispatch_order_line_seq'::regclass);


--
-- Name: district id_district; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY district ALTER COLUMN id_district SET DEFAULT nextval('district_id_seq'::regclass);


--
-- Name: document id_document; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY document ALTER COLUMN id_document SET DEFAULT nextval('document_id_document_seq'::regclass);


--
-- Name: movement id_movement; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY movement ALTER COLUMN id_movement SET DEFAULT nextval('movement_id_movement_seq'::regclass);


--
-- Name: product id_product; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY product ALTER COLUMN id_product SET DEFAULT nextval('product_id_product_seq'::regclass);


--
-- Name: product_type id_product_type; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY product_type ALTER COLUMN id_product_type SET DEFAULT nextval('product_type_id_product_type_seq_1'::regclass);


--
-- Name: rack id_rack; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY rack ALTER COLUMN id_rack SET DEFAULT nextval('rack_id_rack_seq'::regclass);


--
-- Name: refund id_refund; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY refund ALTER COLUMN id_refund SET DEFAULT nextval('refund_seq'::regclass);


--
-- Name: request_order id_request_order; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY request_order ALTER COLUMN id_request_order SET DEFAULT nextval('request_order_id_request_order_seq'::regclass);


--
-- Name: request_order_line id_request_order_line; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY request_order_line ALTER COLUMN id_request_order_line SET DEFAULT nextval('request_order_line_id_request_order_line_seq'::regclass);


--
-- Name: role id_role; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY role ALTER COLUMN id_role SET DEFAULT nextval('role_id_role_seq'::regclass);


--
-- Name: sale_condition id_sale_condition; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY sale_condition ALTER COLUMN id_sale_condition SET DEFAULT nextval('sale_condition_id_sale_condition_seq'::regclass);


--
-- Name: sale_condition_type id_sale_condition_type; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY sale_condition_type ALTER COLUMN id_sale_condition_type SET DEFAULT nextval('sale_condition_type_id_sale_condition_type_seq_1'::regclass);


--
-- Name: unit_of_measure id_unit_of_measure; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY unit_of_measure ALTER COLUMN id_unit_of_measure SET DEFAULT nextval('unit_of_measure_id_unit_of_measure_seq_1'::regclass);


--
-- Name: users id_user; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY users ALTER COLUMN id_user SET DEFAULT nextval('users_id_user_seq'::regclass);


--
-- Name: vehicle id_vehicle; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY vehicle ALTER COLUMN id_vehicle SET DEFAULT nextval('vehicle_id_vehicle_seq_1'::regclass);


--
-- Name: view id_view; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY view ALTER COLUMN id_view SET DEFAULT nextval('view_id_view_seq_1'::regclass);


--
-- Name: warehouse id_warehouse; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY warehouse ALTER COLUMN id_warehouse SET DEFAULT nextval('warehouse_id_warehouse_seq'::regclass);


--
-- Name: zone id_zone; Type: DEFAULT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY zone ALTER COLUMN id_zone SET DEFAULT nextval('zone_id_zone_seq_1'::regclass);


--
-- Name: area area_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY area
    ADD CONSTRAINT area_pk PRIMARY KEY (id_area);


--
-- Name: batch batch_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY batch
    ADD CONSTRAINT batch_pk PRIMARY KEY (id_batch);


--
-- Name: client client_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY client
    ADD CONSTRAINT client_pk PRIMARY KEY (id_client);


--
-- Name: complaint complaint_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY complaint
    ADD CONSTRAINT complaint_pk PRIMARY KEY (id_complaint);


--
-- Name: dispatch_order_line dispatch_order_line_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY dispatch_order_line
    ADD CONSTRAINT dispatch_order_line_pk PRIMARY KEY (id_dispatch_order_line);


--
-- Name: dispatch_order dispatch_order_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY dispatch_order
    ADD CONSTRAINT dispatch_order_pk PRIMARY KEY (id_dispatch_order);


--
-- Name: district district_pkey; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY district
    ADD CONSTRAINT district_pkey PRIMARY KEY (id_district);


--
-- Name: document document_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY document
    ADD CONSTRAINT document_pk PRIMARY KEY (id_document);


--
-- Name: group_batch group_batch_pkey; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY group_batch
    ADD CONSTRAINT group_batch_pkey PRIMARY KEY (id_group_batch);


--
-- Name: dispatch_move idpk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY dispatch_move
    ADD CONSTRAINT idpk PRIMARY KEY (id_dispatch_move);


--
-- Name: movement movement_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY movement
    ADD CONSTRAINT movement_pk PRIMARY KEY (id_movement);


--
-- Name: movementxbatch movementxbatch_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY movementxbatch
    ADD CONSTRAINT movementxbatch_pk PRIMARY KEY (id_batch, id_movement);


--
-- Name: movementxdispatch movementxdispatch_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY movementxdispatch
    ADD CONSTRAINT movementxdispatch_pk PRIMARY KEY (id_movement, id_dispatch_order_line);


--
-- Name: movementxdocument movementxdocument_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY movementxdocument
    ADD CONSTRAINT movementxdocument_pk PRIMARY KEY (id_document, id_movement);


--
-- Name: permission permission_id_permission_key; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY permission
    ADD CONSTRAINT permission_id_permission_key UNIQUE (id_permission);


--
-- Name: permission permission_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY permission
    ADD CONSTRAINT permission_pk PRIMARY KEY (id_view, id_role);


--
-- Name: product product_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY product
    ADD CONSTRAINT product_pk PRIMARY KEY (id_product);


--
-- Name: product_type product_type_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY product_type
    ADD CONSTRAINT product_type_pk PRIMARY KEY (id_product_type);


--
-- Name: rack rack_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY rack
    ADD CONSTRAINT rack_pk PRIMARY KEY (id_rack);


--
-- Name: refund_line refund_line_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY refund_line
    ADD CONSTRAINT refund_line_pk PRIMARY KEY (id_refund_line);


--
-- Name: refund refund_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY refund
    ADD CONSTRAINT refund_pk PRIMARY KEY (id_refund);


--
-- Name: request_order_line request_order_line_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY request_order_line
    ADD CONSTRAINT request_order_line_pk PRIMARY KEY (id_request_order_line);


--
-- Name: request_order request_order_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY request_order
    ADD CONSTRAINT request_order_pk PRIMARY KEY (id_request_order);


--
-- Name: role role_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY role
    ADD CONSTRAINT role_pk PRIMARY KEY (id_role);


--
-- Name: sale_condition sale_condition_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY sale_condition
    ADD CONSTRAINT sale_condition_pk PRIMARY KEY (id_sale_condition);


--
-- Name: sale_condition_type sale_condition_type_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY sale_condition_type
    ADD CONSTRAINT sale_condition_type_pk PRIMARY KEY (id_sale_condition_type);


--
-- Name: unit_of_measure unit_of_measure_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY unit_of_measure
    ADD CONSTRAINT unit_of_measure_pk PRIMARY KEY (id_unit_of_measure);


--
-- Name: users users_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY users
    ADD CONSTRAINT users_pk PRIMARY KEY (id_user);


--
-- Name: vehicle vehicle_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY vehicle
    ADD CONSTRAINT vehicle_pk PRIMARY KEY (id_vehicle);


--
-- Name: view view_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY view
    ADD CONSTRAINT view_pk PRIMARY KEY (id_view);


--
-- Name: warehouse warehouse_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY warehouse
    ADD CONSTRAINT warehouse_pk PRIMARY KEY (id_warehouse);


--
-- Name: zone zone_pk; Type: CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY zone
    ADD CONSTRAINT zone_pk PRIMARY KEY (id_zone);


--
-- Name: warehouse product_new_relations; Type: TRIGGER; Schema: campis; Owner: postgres
--

CREATE TRIGGER product_new_relations AFTER INSERT ON warehouse FOR EACH ROW EXECUTE PROCEDURE public.create_new_relations();


--
-- Name: product product_relations; Type: TRIGGER; Schema: campis; Owner: postgres
--

CREATE TRIGGER product_relations AFTER INSERT ON product FOR EACH ROW EXECUTE PROCEDURE public.create_relations();


--
-- Name: request_order_line stock_update_sale_t; Type: TRIGGER; Schema: campis; Owner: postgres
--

CREATE TRIGGER stock_update_sale_t AFTER INSERT ON request_order_line FOR EACH ROW EXECUTE PROCEDURE public.stock_update_sale();


--
-- Name: movement stock_update_t; Type: TRIGGER; Schema: campis; Owner: postgres
--

CREATE TRIGGER stock_update_t AFTER INSERT ON movement FOR EACH ROW EXECUTE PROCEDURE public.stock_update();


--
-- Name: area area_id_typeproduct_fkey; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY area
    ADD CONSTRAINT area_id_typeproduct_fkey FOREIGN KEY (product_type) REFERENCES product_type(id_product_type) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: movementxbatch batch_movementxbatch_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY movementxbatch
    ADD CONSTRAINT batch_movementxbatch_fk FOREIGN KEY (id_batch) REFERENCES batch(id_batch) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: batch batch_product_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY batch
    ADD CONSTRAINT batch_product_fk FOREIGN KEY (id_product) REFERENCES product(id_product) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: request_order client_request_order_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY request_order
    ADD CONSTRAINT client_request_order_fk FOREIGN KEY (id_client) REFERENCES client(id_client) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: complaint complaint_id_request_fkey; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY complaint
    ADD CONSTRAINT complaint_id_request_fkey FOREIGN KEY (id_request_order) REFERENCES request_order(id_request_order) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: dispatch_order_line dispatch_order_dispatch_order_line_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY dispatch_order_line
    ADD CONSTRAINT dispatch_order_dispatch_order_line_fk FOREIGN KEY (id_dispatch_order) REFERENCES dispatch_order(id_dispatch_order) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: movementxdispatch dispatch_order_line_movementxdispatch_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY movementxdispatch
    ADD CONSTRAINT dispatch_order_line_movementxdispatch_fk FOREIGN KEY (id_dispatch_order_line) REFERENCES dispatch_order_line(id_dispatch_order_line) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: request_order district_request_fkey; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY request_order
    ADD CONSTRAINT district_request_fkey FOREIGN KEY (id_district) REFERENCES district(id_district) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: movementxdocument document_movementxdocument_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY movementxdocument
    ADD CONSTRAINT document_movementxdocument_fk FOREIGN KEY (id_document) REFERENCES document(id_document) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: movementxbatch movement_movementxbatch_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY movementxbatch
    ADD CONSTRAINT movement_movementxbatch_fk FOREIGN KEY (id_movement) REFERENCES movement(id_movement) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: movementxdispatch movement_movementxdispatch_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY movementxdispatch
    ADD CONSTRAINT movement_movementxdispatch_fk FOREIGN KEY (id_movement) REFERENCES movement(id_movement) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: movementxdocument movement_movementxdocument_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY movementxdocument
    ADD CONSTRAINT movement_movementxdocument_fk FOREIGN KEY (id_movement) REFERENCES movement(id_movement) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: dispatch_order_line product_dispatch_order_line_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY dispatch_order_line
    ADD CONSTRAINT product_dispatch_order_line_fk FOREIGN KEY (id_product) REFERENCES product(id_product) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: request_order_line product_request_order_line_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY request_order_line
    ADD CONSTRAINT product_request_order_line_fk FOREIGN KEY (id_product) REFERENCES product(id_product) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: product product_type_product_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY product
    ADD CONSTRAINT product_type_product_fk FOREIGN KEY (id_product_type) REFERENCES product_type(id_product_type) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: productxwarehouse productxwarehouse_id_product_fkey; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY productxwarehouse
    ADD CONSTRAINT productxwarehouse_id_product_fkey FOREIGN KEY (id_product) REFERENCES product(id_product) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: productxwarehouse productxwarehouse_id_warehouse_fkey; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY productxwarehouse
    ADD CONSTRAINT productxwarehouse_id_warehouse_fkey FOREIGN KEY (id_warehouse) REFERENCES warehouse(id_warehouse) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: zone rack_zone_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY zone
    ADD CONSTRAINT rack_zone_fk FOREIGN KEY (id_rack) REFERENCES rack(id_rack) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: refund refund_id_complaint_fkey; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY refund
    ADD CONSTRAINT refund_id_complaint_fkey FOREIGN KEY (id_complaint) REFERENCES complaint(id_complaint) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: refund_line refund_id_refundline_fkey; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY refund_line
    ADD CONSTRAINT refund_id_refundline_fkey FOREIGN KEY (id_refund) REFERENCES refund(id_refund) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: dispatch_order request_order_dispatch_order_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY dispatch_order
    ADD CONSTRAINT request_order_dispatch_order_fk FOREIGN KEY (id_request_order) REFERENCES request_order(id_request_order) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: document request_order_document_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY document
    ADD CONSTRAINT request_order_document_fk FOREIGN KEY (id_request_order) REFERENCES request_order(id_request_order) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: request_order_line request_order_request_order_line_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY request_order_line
    ADD CONSTRAINT request_order_request_order_line_fk FOREIGN KEY (id_request_order) REFERENCES request_order(id_request_order) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: permission role_permission_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY permission
    ADD CONSTRAINT role_permission_fk FOREIGN KEY (id_role) REFERENCES role(id_role) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: users role_user_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY users
    ADD CONSTRAINT role_user_fk FOREIGN KEY (id_role) REFERENCES role(id_role) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: product unit_of_measure_product_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY product
    ADD CONSTRAINT unit_of_measure_product_fk FOREIGN KEY (id_unit_of_measure) REFERENCES unit_of_measure(id_unit_of_measure) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: movement user_movement_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY movement
    ADD CONSTRAINT user_movement_fk FOREIGN KEY (id_user) REFERENCES users(id_user) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: movement vehicle_movement_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY movement
    ADD CONSTRAINT vehicle_movement_fk FOREIGN KEY (id_vehicle) REFERENCES vehicle(id_vehicle) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: permission view_permission_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY permission
    ADD CONSTRAINT view_permission_fk FOREIGN KEY (id_view) REFERENCES view(id_view) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: area warehouse_area_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY area
    ADD CONSTRAINT warehouse_area_fk FOREIGN KEY (id_warehouse) REFERENCES warehouse(id_warehouse) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: rack warehouse_rack_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY rack
    ADD CONSTRAINT warehouse_rack_fk FOREIGN KEY (id_warehouse) REFERENCES warehouse(id_warehouse) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: vehicle warehouse_vehicle_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY vehicle
    ADD CONSTRAINT warehouse_vehicle_fk FOREIGN KEY (id_warehouse) REFERENCES warehouse(id_warehouse) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: zone warehouse_zone_fk; Type: FK CONSTRAINT; Schema: campis; Owner: postgres
--

ALTER TABLE ONLY zone
    ADD CONSTRAINT warehouse_zone_fk FOREIGN KEY (id_warehouse) REFERENCES warehouse(id_warehouse) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

