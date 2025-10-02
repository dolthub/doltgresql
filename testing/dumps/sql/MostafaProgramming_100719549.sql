-- Downloaded from: https://github.com/MostafaProgramming/100719549/blob/dde5242880a860cda0275c4ccfcf5b33a9fa4e40/db_backup.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 17.0
-- Dumped by pg_dump version 17.0

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
-- Name: decrease_stock_level(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.decrease_stock_level() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  -- Ensure stock level does not go below zero
  IF (SELECT stock_level FROM product WHERE product_id = NEW.product_id) < NEW.quantity_sold THEN
    RAISE EXCEPTION 'Insufficient stock for product_id %', NEW.product_id;
  END IF;

  -- Decrease the stock level for the product
  UPDATE product
  SET stock_level = stock_level - NEW.quantity_sold
  WHERE product_id = NEW.product_id;

  RETURN NEW;
END;
$$;


ALTER FUNCTION public.decrease_stock_level() OWNER TO postgres;

--
-- Name: update_delivery_date(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_delivery_date() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  -- If status is changed to 'Completed', set delivery_date to current date
  IF NEW.order_status = 'Completed' THEN
    NEW.delivery_date = CURRENT_DATE;
  -- If status is not 'Completed', set delivery_date to NULL
  ELSE
    NEW.delivery_date = NULL;
  END IF;
  RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_delivery_date() OWNER TO postgres;

--
-- Name: update_inventory_on_purchase_order(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_inventory_on_purchase_order() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Check if the status is being changed to 'Completed'
    IF NEW.order_status = 'Completed' AND OLD.order_status != 'Completed' THEN
        -- Update the inventory table for the respective location and product
        UPDATE inventory
        SET quantity = quantity + NEW.quantity,
            last_updated = NOW()
        WHERE product_id = NEW.product_id AND location_id = NEW.location_id;
    END IF;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_inventory_on_purchase_order() OWNER TO postgres;

--
-- Name: update_inventory_on_sale(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_inventory_on_sale() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  -- Check if the product exists in the inventory for the given store
  IF EXISTS (
    SELECT 1 
    FROM inventory
    WHERE product_id = NEW.product_id AND location_id = NEW.store_id
  ) THEN
    -- Update the inventory quantity for the product at the store
    UPDATE inventory
    SET quantity = quantity - NEW.quantity_sold,
        last_updated = NOW()
    WHERE product_id = NEW.product_id AND location_id = NEW.store_id;
  ELSE
    -- If the inventory record doesn't exist for the store, raise an error
    RAISE EXCEPTION 'Product % not found in inventory for store %', NEW.product_id, NEW.store_id;
  END IF;

  RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_inventory_on_sale() OWNER TO postgres;

--
-- Name: update_inventory_on_transfer(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_inventory_on_transfer() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  -- Reduce stock at the source location
  UPDATE inventory
  SET quantity = quantity - NEW.quantity,
      last_updated = NOW()
  WHERE product_id = NEW.product_id AND location_id = NEW.source_location_id;

  -- Check if the product exists in the inventory at the destination location
  IF EXISTS (
    SELECT 1 
    FROM inventory 
    WHERE product_id = NEW.product_id AND location_id = NEW.destination_location_id
  ) THEN
    -- Increase stock at the destination location
    UPDATE inventory
    SET quantity = quantity + NEW.quantity,
        last_updated = NOW()
    WHERE product_id = NEW.product_id AND location_id = NEW.destination_location_id;
  ELSE
    -- If the product does not exist in the destination, create a new record
    INSERT INTO inventory (product_id, location_id, quantity, last_updated)
    VALUES (NEW.product_id, NEW.destination_location_id, NEW.quantity, NOW());
  END IF;

  RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_inventory_on_transfer() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: attendance; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.attendance (
    attendance_id integer NOT NULL,
    person_id integer,
    date date,
    status character varying(25)
);


ALTER TABLE public.attendance OWNER TO postgres;

--
-- Name: attendance_attendance_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.attendance_attendance_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.attendance_attendance_id_seq OWNER TO postgres;

--
-- Name: attendance_attendance_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.attendance_attendance_id_seq OWNED BY public.attendance.attendance_id;


--
-- Name: budget; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.budget (
    budget_id integer NOT NULL,
    department_id integer,
    allocated_amount numeric(10,2),
    start_date date DEFAULT CURRENT_DATE,
    end_date date,
    description character varying(256),
    category character varying(256)
);


ALTER TABLE public.budget OWNER TO postgres;

--
-- Name: budget_budget_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.budget_budget_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.budget_budget_id_seq OWNER TO postgres;

--
-- Name: budget_budget_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.budget_budget_id_seq OWNED BY public.budget.budget_id;


--
-- Name: department; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.department (
    department_id integer NOT NULL,
    department_name character varying(50),
    budget numeric(10,2),
    manager_id integer
);


ALTER TABLE public.department OWNER TO postgres;

--
-- Name: department_department_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.department_department_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.department_department_id_seq OWNER TO postgres;

--
-- Name: department_department_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.department_department_id_seq OWNED BY public.department.department_id;


--
-- Name: employee; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.employee (
    employee_id integer NOT NULL,
    person_id integer,
    absences integer
);


ALTER TABLE public.employee OWNER TO postgres;

--
-- Name: employee_employee_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.employee_employee_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.employee_employee_id_seq OWNER TO postgres;

--
-- Name: employee_employee_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.employee_employee_id_seq OWNED BY public.employee.employee_id;


--
-- Name: executive; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.executive (
    executive_id integer NOT NULL,
    person_id integer
);


ALTER TABLE public.executive OWNER TO postgres;

--
-- Name: executive_executive_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.executive_executive_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.executive_executive_id_seq OWNER TO postgres;

--
-- Name: executive_executive_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.executive_executive_id_seq OWNED BY public.executive.executive_id;


--
-- Name: expenses; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.expenses (
    expense_id integer NOT NULL,
    amount numeric(10,2),
    department_id integer,
    budget_id integer,
    category character varying(50),
    date_of_expense date,
    description character varying(100),
    person_id integer
);


ALTER TABLE public.expenses OWNER TO postgres;

--
-- Name: expenses_expense_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.expenses_expense_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.expenses_expense_id_seq OWNER TO postgres;

--
-- Name: expenses_expense_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.expenses_expense_id_seq OWNED BY public.expenses.expense_id;


--
-- Name: financial_report; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.financial_report (
    report_id integer NOT NULL,
    department_id integer,
    store_id integer,
    report_type character varying(25),
    generated_date date,
    period_start date,
    period_end date,
    total_expenses numeric(10,2),
    total_revenue numeric(10,2),
    net_profit numeric(10,2)
);


ALTER TABLE public.financial_report OWNER TO postgres;

--
-- Name: financialreport_report_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.financialreport_report_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.financialreport_report_id_seq OWNER TO postgres;

--
-- Name: financialreport_report_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.financialreport_report_id_seq OWNED BY public.financial_report.report_id;


--
-- Name: inventory; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.inventory (
    inventory_id integer NOT NULL,
    product_id integer,
    location_id integer,
    quantity integer,
    last_updated date
);


ALTER TABLE public.inventory OWNER TO postgres;

--
-- Name: inventory_inventory_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.inventory_inventory_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.inventory_inventory_id_seq OWNER TO postgres;

--
-- Name: inventory_inventory_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.inventory_inventory_id_seq OWNED BY public.inventory.inventory_id;


--
-- Name: location; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.location (
    location_id integer NOT NULL,
    location_name character varying(25),
    location character varying(100),
    manager_id integer,
    contact_number character varying(15)
);


ALTER TABLE public.location OWNER TO postgres;

--
-- Name: location_location_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.location_location_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.location_location_id_seq OWNER TO postgres;

--
-- Name: location_location_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.location_location_id_seq OWNED BY public.location.location_id;


--
-- Name: manager; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.manager (
    manager_id integer NOT NULL,
    person_id integer,
    absences integer
);


ALTER TABLE public.manager OWNER TO postgres;

--
-- Name: manager_manager_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.manager_manager_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.manager_manager_id_seq OWNER TO postgres;

--
-- Name: manager_manager_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.manager_manager_id_seq OWNED BY public.manager.manager_id;


--
-- Name: payroll; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.payroll (
    payroll_id integer NOT NULL,
    person_id integer,
    department_id integer,
    budget_id integer,
    salary numeric(10,2),
    bonuses numeric(10,2),
    deductions numeric(10,2),
    net_salary numeric(10,2),
    pay_date date
);


ALTER TABLE public.payroll OWNER TO postgres;

--
-- Name: payroll_payroll_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.payroll_payroll_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.payroll_payroll_id_seq OWNER TO postgres;

--
-- Name: payroll_payroll_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.payroll_payroll_id_seq OWNED BY public.payroll.payroll_id;


--
-- Name: person; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.person (
    person_id integer NOT NULL,
    name character varying(100),
    email character varying(100),
    role character varying(50),
    salary numeric(10,2),
    department_id integer,
    location_id integer,
    hire_date date,
    location character varying(100)
);


ALTER TABLE public.person OWNER TO postgres;

--
-- Name: person_person_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.person_person_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.person_person_id_seq OWNER TO postgres;

--
-- Name: person_person_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.person_person_id_seq OWNED BY public.person.person_id;


--
-- Name: product; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.product (
    product_id integer NOT NULL,
    product_name character varying(100),
    category character varying(50),
    price numeric(10,2),
    stock_level integer,
    reorder_level integer,
    last_purchase_date date,
    supplier_id integer,
    cost numeric(10,2)
);


ALTER TABLE public.product OWNER TO postgres;

--
-- Name: product_product_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.product_product_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.product_product_id_seq OWNER TO postgres;

--
-- Name: product_product_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.product_product_id_seq OWNED BY public.product.product_id;


--
-- Name: purchase_order; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.purchase_order (
    purchase_order_id integer NOT NULL,
    supplier_id integer,
    product_id integer,
    order_date date,
    delivery_date date,
    total_amount numeric(10,2),
    order_status character varying(25),
    quantity integer DEFAULT 1 NOT NULL,
    location_id integer,
    CONSTRAINT purchase_order_status_check CHECK (((order_status)::text = ANY ((ARRAY['Pending'::character varying, 'Completed'::character varying, 'Cancelled'::character varying, 'Awaiting Approval'::character varying, 'Late'::character varying])::text[])))
);


ALTER TABLE public.purchase_order OWNER TO postgres;

--
-- Name: purchaseorder_purchase_order_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.purchaseorder_purchase_order_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.purchaseorder_purchase_order_id_seq OWNER TO postgres;

--
-- Name: purchaseorder_purchase_order_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.purchaseorder_purchase_order_id_seq OWNED BY public.purchase_order.purchase_order_id;


--
-- Name: sales_forecast; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sales_forecast (
    forecast_id integer NOT NULL,
    store_id integer,
    product_id integer,
    start_date date,
    end_date date,
    predicted_sales integer,
    created_date date
);


ALTER TABLE public.sales_forecast OWNER TO postgres;

--
-- Name: sales_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sales_record (
    sales_id integer NOT NULL,
    store_id integer,
    manager_id integer,
    product_id integer,
    quantity_sold integer,
    date_of_sale date,
    total_amount numeric(10,2),
    payment_method character varying(25)
);


ALTER TABLE public.sales_record OWNER TO postgres;

--
-- Name: salesforecast_forecast_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.salesforecast_forecast_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.salesforecast_forecast_id_seq OWNER TO postgres;

--
-- Name: salesforecast_forecast_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.salesforecast_forecast_id_seq OWNED BY public.sales_forecast.forecast_id;


--
-- Name: salesrecord_sales_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.salesrecord_sales_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.salesrecord_sales_id_seq OWNER TO postgres;

--
-- Name: salesrecord_sales_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.salesrecord_sales_id_seq OWNED BY public.sales_record.sales_id;


--
-- Name: stock_transfer; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.stock_transfer (
    transfer_id integer NOT NULL,
    source_location_id integer,
    destination_location_id integer,
    product_id integer,
    quantity integer,
    transfer_date date,
    status character varying(25)
);


ALTER TABLE public.stock_transfer OWNER TO postgres;

--
-- Name: stocktransfer_transfer_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.stocktransfer_transfer_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.stocktransfer_transfer_id_seq OWNER TO postgres;

--
-- Name: stocktransfer_transfer_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.stocktransfer_transfer_id_seq OWNED BY public.stock_transfer.transfer_id;


--
-- Name: store; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.store (
    store_id integer NOT NULL,
    location_id integer,
    operating_hours character varying(50)
);


ALTER TABLE public.store OWNER TO postgres;

--
-- Name: store_store_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.store_store_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.store_store_id_seq OWNER TO postgres;

--
-- Name: store_store_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.store_store_id_seq OWNED BY public.store.store_id;


--
-- Name: supplier; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.supplier (
    supplier_id integer NOT NULL,
    supplier_name character varying(50),
    contact_details character varying(25),
    location character varying(25),
    contract_terms character varying(250)
);


ALTER TABLE public.supplier OWNER TO postgres;

--
-- Name: supplier_supplier_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.supplier_supplier_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.supplier_supplier_id_seq OWNER TO postgres;

--
-- Name: supplier_supplier_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.supplier_supplier_id_seq OWNED BY public.supplier.supplier_id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    user_id integer NOT NULL,
    person_id integer,
    email character varying(50),
    password character varying(255),
    role character varying(25),
    location_id integer
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: users_user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.users_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_user_id_seq OWNER TO postgres;

--
-- Name: users_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.users_user_id_seq OWNED BY public.users.user_id;


--
-- Name: warehouse; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.warehouse (
    warehouse_id integer NOT NULL,
    location_id integer,
    capacity integer
);


ALTER TABLE public.warehouse OWNER TO postgres;

--
-- Name: warehouse_warehouse_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.warehouse_warehouse_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.warehouse_warehouse_id_seq OWNER TO postgres;

--
-- Name: warehouse_warehouse_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.warehouse_warehouse_id_seq OWNED BY public.warehouse.warehouse_id;


--
-- Name: attendance attendance_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.attendance ALTER COLUMN attendance_id SET DEFAULT nextval('public.attendance_attendance_id_seq'::regclass);


--
-- Name: budget budget_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.budget ALTER COLUMN budget_id SET DEFAULT nextval('public.budget_budget_id_seq'::regclass);


--
-- Name: department department_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.department ALTER COLUMN department_id SET DEFAULT nextval('public.department_department_id_seq'::regclass);


--
-- Name: employee employee_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employee ALTER COLUMN employee_id SET DEFAULT nextval('public.employee_employee_id_seq'::regclass);


--
-- Name: executive executive_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.executive ALTER COLUMN executive_id SET DEFAULT nextval('public.executive_executive_id_seq'::regclass);


--
-- Name: expenses expense_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.expenses ALTER COLUMN expense_id SET DEFAULT nextval('public.expenses_expense_id_seq'::regclass);


--
-- Name: financial_report report_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.financial_report ALTER COLUMN report_id SET DEFAULT nextval('public.financialreport_report_id_seq'::regclass);


--
-- Name: inventory inventory_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inventory ALTER COLUMN inventory_id SET DEFAULT nextval('public.inventory_inventory_id_seq'::regclass);


--
-- Name: location location_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.location ALTER COLUMN location_id SET DEFAULT nextval('public.location_location_id_seq'::regclass);


--
-- Name: manager manager_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.manager ALTER COLUMN manager_id SET DEFAULT nextval('public.manager_manager_id_seq'::regclass);


--
-- Name: payroll payroll_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payroll ALTER COLUMN payroll_id SET DEFAULT nextval('public.payroll_payroll_id_seq'::regclass);


--
-- Name: person person_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.person ALTER COLUMN person_id SET DEFAULT nextval('public.person_person_id_seq'::regclass);


--
-- Name: product product_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product ALTER COLUMN product_id SET DEFAULT nextval('public.product_product_id_seq'::regclass);


--
-- Name: purchase_order purchase_order_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.purchase_order ALTER COLUMN purchase_order_id SET DEFAULT nextval('public.purchaseorder_purchase_order_id_seq'::regclass);


--
-- Name: sales_forecast forecast_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sales_forecast ALTER COLUMN forecast_id SET DEFAULT nextval('public.salesforecast_forecast_id_seq'::regclass);


--
-- Name: sales_record sales_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sales_record ALTER COLUMN sales_id SET DEFAULT nextval('public.salesrecord_sales_id_seq'::regclass);


--
-- Name: stock_transfer transfer_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_transfer ALTER COLUMN transfer_id SET DEFAULT nextval('public.stocktransfer_transfer_id_seq'::regclass);


--
-- Name: store store_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.store ALTER COLUMN store_id SET DEFAULT nextval('public.store_store_id_seq'::regclass);


--
-- Name: supplier supplier_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.supplier ALTER COLUMN supplier_id SET DEFAULT nextval('public.supplier_supplier_id_seq'::regclass);


--
-- Name: users user_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN user_id SET DEFAULT nextval('public.users_user_id_seq'::regclass);


--
-- Name: warehouse warehouse_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.warehouse ALTER COLUMN warehouse_id SET DEFAULT nextval('public.warehouse_warehouse_id_seq'::regclass);


--
-- Data for Name: attendance; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.attendance (attendance_id, person_id, date, status) FROM stdin;
1	3	2025-01-02	Present
2	3	2025-01-01	Absent
3	3	2024-12-31	Absent
4	3	2025-01-03	Present
5	2	2025-01-02	Present
6	3	2025-01-02	Absent
7	4	2025-01-01	Present
8	4	2025-01-02	Present
\.


--
-- Data for Name: budget; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.budget (budget_id, department_id, allocated_amount, start_date, end_date, description, category) FROM stdin;
104	2	15000.00	2025-01-01	2025-12-31	Budget for job postings, recruitment drives, and onboarding	Recruitment
105	2	20000.00	2025-01-01	2025-12-31	Budget for employee upskilling and training programs	Employee Training
106	2	10000.00	2025-01-01	2025-12-31	Budget for team-building activities and welfare programs	Employee Welfare
110	4	25000.00	2025-01-01	2025-12-31	Budget for vendor-related activities and agreements	Vendor Management
111	4	50000.00	2025-01-01	2025-12-31	Budget for warehouse and stock handling costs	Inventory Management
112	4	15000.00	2025-01-01	2025-12-31	Budget for office and operational supplies	Supplies
113	5	20000.00	2025-01-01	2025-12-31	Budget for electricity, water, and other utilities	Utilities
114	5	30000.00	2025-01-01	2025-12-31	Budget for transportation and delivery costs	Logistics
115	5	20000.00	2025-01-01	2025-12-31	Budget for equipment and infrastructure maintenance	Maintenance
107	1	30000.00	2025-01-01	2025-12-31	Travel budget for sales personnel	Travel
108	1	40000.00	2025-01-01	2025-12-31	Budget for promotional campaigns and marketing materials	Promotions
109	1	15000.00	2025-01-01	2025-12-31	Budget for client dinners and meetings	Client Entertainment
101	3	50000.00	2025-01-01	2025-12-31	Annual budget for financial research 	Financial Research
102	3	30000.00	2025-01-01	2025-12-31	Budget for getting projected forecasts	Forecasting Methods
103	3	20000.00	2025-01-01	2025-12-31	Budget for Outsourcing 	Outsourcing Finance
\.


--
-- Data for Name: department; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.department (department_id, department_name, budget, manager_id) FROM stdin;
5	Operations	700000.00	\N
1	Sales	500000.00	6
2	HR	350000.00	7
3	Finance	600000.00	8
4	Procurement	600000.00	9
\.


--
-- Data for Name: employee; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.employee (employee_id, person_id, absences) FROM stdin;
4	6	0
6	17	0
7	18	0
8	19	0
9	20	0
10	21	0
11	22	0
12	23	0
13	24	0
14	25	0
15	26	0
16	27	0
17	28	0
18	29	0
19	30	0
20	31	0
1	1	0
2	2	0
3	3	0
5	10	0
\.


--
-- Data for Name: executive; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.executive (executive_id, person_id) FROM stdin;
1	7
2	8
\.


--
-- Data for Name: expenses; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.expenses (expense_id, amount, department_id, budget_id, category, date_of_expense, description, person_id) FROM stdin;
26	5000.00	3	101	Financial Research	2025-01-15	Consultation with financial analysts	15
27	3000.00	3	102	Forecasting Methods	2025-02-01	Workshop on forecasting techniques	19
28	5000.00	3	103	Outsourcing Finance	2025-03-10	Payment to outsourced accounting firm	15
29	4000.00	2	104	Recruitment	2025-01-20	Job advertisement costs for new hires	14
30	3000.00	2	105	Employee Training	2025-02-15	Workshop for upskilling employees	14
31	2000.00	2	106	Employee Welfare	2025-03-10	Snacks and refreshments for team-building events	17
32	5000.00	1	107	Travel	2025-01-25	Travel costs for client meetings	22
33	10000.00	1	108	Promotions	2025-02-20	Social media and online campaigns	24
34	3000.00	1	109	Client Entertainment	2025-03-05	Client dinner at a corporate event	13
35	8000.00	5	110	Vendor Management	2025-01-12	Annual vendor contract renewal fees	4
36	12000.00	5	111	Inventory Management	2025-02-01	Restocking of warehouse supplies	12
37	1500.00	5	112	Supplies	2025-02-18	Office supplies for operations	5
38	2000.00	5	113	Utilities	2025-03-10	Electricity bill for March	9
39	4000.00	5	114	Logistics	2025-03-15	Transportation costs for product deliveries	11
40	5000.00	5	115	Maintenance	2025-03-25	Equipment servicing and repair	9
\.


--
-- Data for Name: financial_report; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.financial_report (report_id, department_id, store_id, report_type, generated_date, period_start, period_end, total_expenses, total_revenue, net_profit) FROM stdin;
\.


--
-- Data for Name: inventory; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.inventory (inventory_id, product_id, location_id, quantity, last_updated) FROM stdin;
1	1001	1	120	2025-01-01
2	1001	2	100	2025-01-01
3	1001	3	80	2025-01-01
4	1001	4	80	2025-01-01
5	1001	5	88	2025-01-01
6	1002	1	30	2025-01-01
7	1002	2	25	2025-01-01
8	1002	3	20	2025-01-01
10	1002	5	29	2025-01-01
12	1003	2	50	2025-01-01
13	1003	3	40	2025-01-01
15	1003	5	50	2025-01-01
16	1005	1	50	2025-01-01
17	1005	2	40	2025-01-01
18	1005	3	40	2025-01-01
19	1005	4	50	2025-01-01
21	1007	1	15	2025-01-01
22	1007	2	15	2025-01-01
23	1007	3	10	2025-01-01
24	1007	4	20	2025-01-01
25	1007	5	19	2025-01-01
26	1009	1	40	2025-01-01
27	1009	2	35	2025-01-01
28	1009	3	30	2025-01-01
29	1009	4	40	2025-01-01
30	1009	5	39	2025-01-01
31	1010	1	70	2025-01-01
32	1010	2	60	2025-01-01
33	1010	3	60	2025-01-01
34	1010	4	70	2025-01-01
35	1010	5	68	2025-01-01
36	1004	1	130	2025-01-01
37	1004	2	120	2025-01-01
38	1004	3	120	2025-01-01
39	1004	4	130	2025-01-01
41	1008	1	80	2025-01-01
42	1008	2	75	2025-01-01
43	1008	3	70	2025-01-01
44	1008	4	75	2025-01-01
45	1008	5	76	2025-01-01
47	1006	2	8	2025-01-01
48	1006	3	6	2025-01-01
49	1006	4	8	2025-01-01
50	1006	5	10	2025-01-01
46	1006	1	5	2025-01-05
11	1003	1	60	2025-01-05
14	1003	4	60	2025-01-05
9	1002	4	49	2025-01-05
40	1004	5	148	2025-01-05
20	1005	5	57	2025-01-05
\.


--
-- Data for Name: location; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.location (location_id, location_name, location, manager_id, contact_number) FROM stdin;
1	Las Vegas Store	Las Vegas, NV	1	702-555-0100
2	Sydney Store	Sydney, Australia	2	61-2-5550-1234
5	New York Warehouse	New York, NY	3	212-555-0200
3	New York Store	New York, NY	4	212-555-0200
4	Washington Warehouse	Washington	5	212-555-0200
\.


--
-- Data for Name: manager; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.manager (manager_id, person_id, absences) FROM stdin;
1	4	1
2	5	0
3	9	2
4	11	0
5	12	0
6	13	0
7	14	0
8	15	0
9	16	0
\.


--
-- Data for Name: payroll; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.payroll (payroll_id, person_id, department_id, budget_id, salary, bonuses, deductions, net_salary, pay_date) FROM stdin;
6	6	2	108	76000.00	4500.00	2000.00	78500.00	2025-01-31
7	7	5	113	58000.00	2000.00	500.00	59500.00	2025-01-31
8	8	5	113	60000.00	1000.00	800.00	60200.00	2025-01-31
9	9	5	113	55000.00	1500.00	700.00	55800.00	2025-01-31
10	10	4	106	80000.00	5000.00	3000.00	82000.00	2025-01-31
11	1	3	101	90000.00	3000.00	2000.00	91000.00	2024-12-31
12	2	3	101	95000.00	3500.00	2500.00	96000.00	2024-12-31
13	3	1	102	78000.00	3000.00	2000.00	79000.00	2024-11-30
14	4	2	105	77000.00	2500.00	3000.00	76500.00	2024-11-30
15	5	4	106	74000.00	2000.00	1000.00	75000.00	2024-10-31
16	6	2	108	76000.00	3500.00	1500.00	78000.00	2024-09-30
17	7	5	113	58000.00	1000.00	500.00	58500.00	2024-12-31
18	8	5	113	60000.00	1200.00	800.00	60400.00	2024-11-30
19	9	5	113	55000.00	1000.00	600.00	55400.00	2024-10-31
20	10	4	106	80000.00	4000.00	2500.00	81500.00	2024-08-31
21	11	1	102	85000.00	2500.00	1500.00	86000.00	2025-01-31
22	12	2	108	50000.00	2000.00	800.00	51200.00	2024-12-31
23	13	3	101	47000.00	3000.00	1000.00	49000.00	2024-12-31
24	14	5	113	75000.00	4000.00	3000.00	76000.00	2024-12-31
25	15	4	106	72000.00	3000.00	2500.00	72500.00	2024-11-30
26	16	5	113	42000.00	2000.00	500.00	43500.00	2024-10-31
27	17	4	106	75000.00	4500.00	2000.00	77500.00	2024-09-30
28	18	1	102	56000.00	2000.00	500.00	57500.00	2025-01-31
29	19	1	102	60000.00	3000.00	1500.00	61500.00	2025-01-31
30	20	5	113	43000.00	1000.00	700.00	42300.00	2024-12-31
1	1	5	101	90000.00	5000.00	2000.00	93000.00	2025-01-31
2	2	5	101	95000.00	4000.00	3000.00	96000.00	2025-01-31
3	3	5	102	78000.00	3000.00	2500.00	78500.00	2025-01-31
4	4	5	105	77000.00	4000.00	3000.00	78000.00	2025-01-31
5	5	5	106	74000.00	2000.00	1000.00	75000.00	2025-01-31
\.


--
-- Data for Name: person; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.person (person_id, name, email, role, salary, department_id, location_id, hire_date, location) FROM stdin;
7	Valentina Shevchenko	valentina.shevchenko@example.com	Executive	90000.00	\N	\N	2017-05-30	London
8	Francis Ngannou	francis.ngannou@example.com	Executive	95000.00	\N	\N	2016-04-12	London
13	Deontay Wilder	deontay.wilder@example.com	Sales Manager	78000.00	1	\N	2019-09-10	London
14	Anthony Joshua	anthony.joshua@example.com	HR Manager	77000.00	2	\N	2021-03-22	London
15	Vasiliy Lomachenko	vasiliy.lomachenko@example.com	Finance Manager	76000.00	3	\N	2017-12-05	London
16	Gervonta Davis	gervonta.davis@example.com	Procurement Manager	74000.00	4	\N	2022-01-20	London
17	Lionel Messi	lionel.messi@example.com	HR Employee	55000.00	2	\N	2021-02-10	London
18	Cristiano Ronaldo	cristiano.ronaldo@example.com	HR Employee	56000.00	2	\N	2020-11-25	London
19	Neymar Jr	neymar.jr@example.com	Finance Employee	54000.00	3	\N	2022-07-18	London
20	Kylian Mbappe	kylian.mbappe@example.com	Procurement Employee	58000.00	4	\N	2019-01-12	London
21	Robert Lewandowski	robert.lewandowski@example.com	Procurement Employee	57000.00	4	\N	2020-08-05	London
22	Kevin De Bruyne	kevin.debruyne@example.com	Sales Employee	55000.00	1	\N	2021-06-15	London
23	Virgil Van Dijk	virgil.vandijk@example.com	Sales Employee	60000.00	1	\N	2018-10-23	London
24	Mohamed Salah	mohamed.salah@example.com	Sales Employee	58000.00	1	\N	2022-02-18	London
1	Conor McGregor	conor.mcgregor@example.com	Store Employee	85000.00	5	1	2021-01-15	Las Vegas
2	Israel Adesanya	israel.adesanya@example.com	Store Employee	42000.00	5	1	2020-11-12	Las Vegas
3	Kamaru Usman	kamaru.usman@example.com	Store Employee	47000.00	5	2	2022-03-01	Sydney
4	Amanda Nunes	amanda.nunes@example.com	Store Manager	75000.00	5	1	2019-06-01	Las Vegas
5	Alexander Volkanovski	alexander.volkanovski@example.com	Store Manager	72000.00	5	2	2018-10-20	Sydney
6	Jon Jones	jon.jones@example.com	Store Employee	43000.00	5	2	2021-12-10	Sydney
9	Sean OMalley	sean.omalley@example.com	Warehouse Manager	76000.00	5	5	2020-09-15	New York
10	Dustin Poirier	dustin.poirier@example.com	Warehouse Employee	40000.00	5	5	2022-07-08	New York
11	Canelo Alvarez	canelo.alvarez@example.com	Store Manager	75000.00	5	3	2020-04-01	New York
12	Tyson Fury	tyson.fury@example.com	Warehouse Manager	80000.00	5	4	2018-06-15	Washington
25	Harry Kane	harry.kane@example.com	Store Employee	59000.00	5	3	2020-03-15	New York
26	Erling Haaland	erling.haaland@example.com	Store Employee	56000.00	5	3	2021-09-01	New York
27	Luka Modric	luka.modric@example.com	Store Employee	57000.00	5	3	2022-01-30	New York
28	Karim Benzema	karim.benzema@example.com	Store Employee	58000.00	5	3	2019-04-25	New York
29	Sadio Mane	sadio.mane@example.com	Warehouse Employee	54000.00	5	4	2021-07-20	Washington
30	Bruno Fernandes	bruno.fernandes@example.com	Warehouse Employee	55000.00	5	4	2020-11-10	Washington
31	Marcus Rashford	marcus.rashford@example.com	Warehouse Employee	60000.00	5	4	2022-05-18	Washington
\.


--
-- Data for Name: product; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.product (product_id, product_name, category, price, stock_level, reorder_level, last_purchase_date, supplier_id, cost) FROM stdin;
1001	High-Pressure Turbine Blade	Engine Component	2500.00	468	100	2024-01-10	1	900.00
1002	Combustion Chamber	Engine Component	12000.00	134	50	2024-01-15	2	4000.00
1003	Fan Blade	Engine Component	1800.00	260	75	2024-01-20	3	1200.00
1005	Compressor Disk	Engine Component	5000.00	231	75	2024-02-01	5	3200.00
1007	Ignition System	Engine Component	4500.00	79	25	2024-02-10	1	2700.00
1009	Bearing Housing	Engine Component	3500.00	184	50	2024-02-20	3	2800.00
1010	Thermal Insulation Blanket	Engine Component	1500.00	328	90	2024-02-25	4	950.00
1004	Fuel Nozzle	Engine Component	800.00	640	200	2024-01-25	4	550.00
1008	Oil Pump	Engine Component	1200.00	376	100	2024-02-15	2	700.00
1006	Thrust Reverser	Engine Component	15000.00	37	20	2024-02-05	6	12500.00
\.


--
-- Data for Name: purchase_order; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.purchase_order (purchase_order_id, supplier_id, product_id, order_date, delivery_date, total_amount, order_status, quantity, location_id) FROM stdin;
8	2	1002	2024-01-15	2024-02-05	3000.00	Completed	15	4
10	4	1004	2024-01-25	2024-02-15	4000.00	Completed	8	5
12	6	1006	2024-02-05	\N	6000.00	Pending	12	5
9	3	1003	2024-01-20	\N	7500.00	Cancelled	25	4
7	1	1002	2024-01-10	\N	5000.00	Awaiting Approval	19	4
11	5	1005	2024-02-01	2025-01-05	2000.00	Completed	6	5
13	2	1002	2025-01-08	\N	24000.00	Awaiting Approval	2	5
14	3	1003	2025-01-08	\N	3600.00	Awaiting Approval	2	5
15	1	1001	2025-01-08	\N	32500.00	Awaiting Approval	13	5
\.


--
-- Data for Name: sales_forecast; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sales_forecast (forecast_id, store_id, product_id, start_date, end_date, predicted_sales, created_date) FROM stdin;
\.


--
-- Data for Name: sales_record; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sales_record (sales_id, store_id, manager_id, product_id, quantity_sold, date_of_sale, total_amount, payment_method) FROM stdin;
1	1	1	1001	5	2025-01-05	12500.00	Credit Card
2	1	1	1003	10	2025-01-06	18000.00	Cash
3	1	1	1007	3	2025-01-07	13500.00	Debit Card
4	2	2	1002	2	2025-01-05	24000.00	Cash
5	2	2	1005	4	2025-01-06	20000.00	Credit Card
6	2	2	1008	6	2025-01-07	21000.00	Debit Card
7	3	4	1004	7	2025-01-05	5600.00	Cash
8	3	4	1006	1	2025-01-06	15000.00	Credit Card
9	3	4	1010	3	2025-01-07	4500.00	Debit Card
10	1	1	1001	10	2025-01-01	25000.00	Credit Card
11	1	1	1002	5	2025-01-02	60000.00	Cash
12	1	1	1003	8	2025-01-05	14400.00	Debit Card
13	1	1	1004	20	2025-01-10	16000.00	Credit Card
14	1	1	1005	3	2025-02-01	15000.00	Cash
15	1	1	1006	2	2025-02-15	30000.00	Credit Card
16	1	1	1007	5	2025-02-20	22500.00	Cash
17	1	1	1008	10	2025-03-01	35000.00	Credit Card
18	1	1	1009	4	2025-03-15	14000.00	Debit Card
19	1	1	1010	6	2025-03-20	9000.00	Cash
20	2	2	1001	12	2025-01-01	30000.00	Cash
21	2	2	1003	15	2025-01-12	27000.00	Debit Card
22	2	2	1004	25	2025-02-05	20000.00	Credit Card
23	2	2	1005	10	2025-02-14	50000.00	Cash
24	2	2	1006	4	2025-02-28	60000.00	Debit Card
25	2	2	1007	7	2025-03-10	31500.00	Cash
26	2	2	1008	6	2025-03-15	21000.00	Credit Card
27	2	2	1009	8	2025-03-20	28000.00	Cash
28	2	2	1010	10	2025-03-25	15000.00	Debit Card
29	2	2	1002	6	2025-03-30	72000.00	Credit Card
30	3	4	1001	5	2025-01-03	12500.00	Debit Card
31	3	4	1002	3	2025-01-15	36000.00	Cash
32	3	4	1003	7	2025-02-05	12600.00	Credit Card
33	3	4	1005	2	2025-02-12	10000.00	Cash
34	3	4	1007	6	2025-02-20	27000.00	Debit Card
35	3	4	1009	4	2025-02-25	14000.00	Credit Card
36	3	4	1010	3	2025-03-05	4500.00	Cash
37	3	4	1004	8	2025-03-10	6400.00	Debit Card
38	3	4	1008	2	2025-03-15	7000.00	Credit Card
39	3	4	1006	1	2025-03-20	15000.00	Cash
40	1	1	1006	5	2025-04-13	12500.00	Credit Card
\.


--
-- Data for Name: stock_transfer; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.stock_transfer (transfer_id, source_location_id, destination_location_id, product_id, quantity, transfer_date, status) FROM stdin;
1	1	4	1003	10	2025-01-05	Completed
\.


--
-- Data for Name: store; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.store (store_id, location_id, operating_hours) FROM stdin;
2	2	09:00-17:00
3	3	09:00-17:00
1	1	09:00-15:00
\.


--
-- Data for Name: supplier; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.supplier (supplier_id, supplier_name, contact_details, location, contract_terms) FROM stdin;
1	Jet Engine Components Co.	123-456-7890	Las Vegas, USA	Supply jet engine components for 3 years
2	Precision Engineering Ltd.	987-654-3210	New York, USA	Exclusive supplier of precision-machined parts
3	High-Tech Materials Inc.	555-666-7777	Washington, USA	Provide high-strength alloys and composites
4	Aerospace Electronics PLC	222-333-4444	Sydney, Australia	Electronics and sensors for Rolls-Royce engines
5	Advanced Manufacturing Solutions	444-555-6666	Coventry, UK	Support with additive manufacturing technologies
6	Global Aviation Supplies	777-888-9999	Istanbul, Turkey	Provide consumables and tools for aviation operations
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (user_id, person_id, email, password, role, location_id) FROM stdin;
1	4	store1@example.com	$2b$10$YFogi0fPHcFT1iM/IMzE5ePCbHIKEkkwnjZhkKycUeQn6ivDt73XS	Store Manager	\N
3	11	store3@example.com	$2b$10$aCzTVMWeLeZd1nknHMuVTe.jB5bpR2TewJC84/4vn7eg5Re5GTNQa	Store Manager	\N
4	9	warehouse1@example.com	$2b$10$cwiQEcv7yngU9oiUgEzCPersfMaWUwqpDQvMxerDOiLgw6SXm1zT2	Warehouse Manager	\N
5	12	warehouse2@example.com	$2b$10$9TbMLjz6seUrR7uY.wnthen3pK8Iu272lznYUzNRhrueOEWRVQfJu	Warehouse Manager	\N
6	13	sales@example.com	$2b$10$BW76tOwtn4kiuNC7S4/XpeGn8RJMmscIE2OlVmjXr6C3eaH.Iulhy	Sales Manager	\N
7	14	hr@example.com	$2b$10$2t8yE0Q.vLVGr91L1QK7Aut5rapNE6iIZcUfi/l/XH1BHau9979C.	HR Manager	\N
8	15	finance@example.com	$2b$10$JrDox9vWSXlIiH6vI4HvMuJyQhCqJcbg/NmByHCB8yaY4vLP3Occy	Finance Manager	\N
9	16	procurement@example.com	$2b$10$Zv.kP.kipkm9aFAxYOBbMOwa4iRkLp3MVQHm.kIDfDA21DFY.NQBq	Procurement Manager	\N
2	5	store2@example.com	$2b$10$BWQJHwNs4N88rmebcI0r4eD5oMguGhk1RnGew/RmKyM0QpMhzJ58e	Store Manager	\N
\.


--
-- Data for Name: warehouse; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.warehouse (warehouse_id, location_id, capacity) FROM stdin;
\.


--
-- Name: attendance_attendance_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.attendance_attendance_id_seq', 8, true);


--
-- Name: budget_budget_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.budget_budget_id_seq', 1, false);


--
-- Name: department_department_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.department_department_id_seq', 5, true);


--
-- Name: employee_employee_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.employee_employee_id_seq', 20, true);


--
-- Name: executive_executive_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.executive_executive_id_seq', 2, true);


--
-- Name: expenses_expense_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.expenses_expense_id_seq', 40, true);


--
-- Name: financialreport_report_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.financialreport_report_id_seq', 1, false);


--
-- Name: inventory_inventory_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.inventory_inventory_id_seq', 50, true);


--
-- Name: location_location_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.location_location_id_seq', 10, true);


--
-- Name: manager_manager_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.manager_manager_id_seq', 15, true);


--
-- Name: payroll_payroll_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.payroll_payroll_id_seq', 30, true);


--
-- Name: person_person_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.person_person_id_seq', 52, true);


--
-- Name: product_product_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.product_product_id_seq', 10, true);


--
-- Name: purchaseorder_purchase_order_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.purchaseorder_purchase_order_id_seq', 15, true);


--
-- Name: salesforecast_forecast_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.salesforecast_forecast_id_seq', 1, false);


--
-- Name: salesrecord_sales_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.salesrecord_sales_id_seq', 40, true);


--
-- Name: stocktransfer_transfer_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.stocktransfer_transfer_id_seq', 1, true);


--
-- Name: store_store_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.store_store_id_seq', 3, true);


--
-- Name: supplier_supplier_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.supplier_supplier_id_seq', 6, true);


--
-- Name: users_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_user_id_seq', 9, true);


--
-- Name: warehouse_warehouse_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.warehouse_warehouse_id_seq', 1, false);


--
-- Name: attendance attendance_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.attendance
    ADD CONSTRAINT attendance_pkey PRIMARY KEY (attendance_id);


--
-- Name: budget budget_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.budget
    ADD CONSTRAINT budget_pkey PRIMARY KEY (budget_id);


--
-- Name: department department_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.department
    ADD CONSTRAINT department_pkey PRIMARY KEY (department_id);


--
-- Name: employee employee_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employee
    ADD CONSTRAINT employee_pkey PRIMARY KEY (employee_id);


--
-- Name: executive executive_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.executive
    ADD CONSTRAINT executive_pkey PRIMARY KEY (executive_id);


--
-- Name: expenses expenses_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.expenses
    ADD CONSTRAINT expenses_pkey PRIMARY KEY (expense_id);


--
-- Name: financial_report financialreport_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.financial_report
    ADD CONSTRAINT financialreport_pkey PRIMARY KEY (report_id);


--
-- Name: inventory inventory_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inventory
    ADD CONSTRAINT inventory_pkey PRIMARY KEY (inventory_id);


--
-- Name: location location_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.location
    ADD CONSTRAINT location_pkey PRIMARY KEY (location_id);


--
-- Name: manager manager_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.manager
    ADD CONSTRAINT manager_pkey PRIMARY KEY (manager_id);


--
-- Name: payroll payroll_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payroll
    ADD CONSTRAINT payroll_pkey PRIMARY KEY (payroll_id);


--
-- Name: person person_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.person
    ADD CONSTRAINT person_email_key UNIQUE (email);


--
-- Name: person person_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.person
    ADD CONSTRAINT person_pkey PRIMARY KEY (person_id);


--
-- Name: product product_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product
    ADD CONSTRAINT product_pkey PRIMARY KEY (product_id);


--
-- Name: purchase_order purchaseorder_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.purchase_order
    ADD CONSTRAINT purchaseorder_pkey PRIMARY KEY (purchase_order_id);


--
-- Name: sales_forecast salesforecast_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sales_forecast
    ADD CONSTRAINT salesforecast_pkey PRIMARY KEY (forecast_id);


--
-- Name: sales_record salesrecord_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sales_record
    ADD CONSTRAINT salesrecord_pkey PRIMARY KEY (sales_id);


--
-- Name: stock_transfer stocktransfer_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_transfer
    ADD CONSTRAINT stocktransfer_pkey PRIMARY KEY (transfer_id);


--
-- Name: store store_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.store
    ADD CONSTRAINT store_pkey PRIMARY KEY (store_id);


--
-- Name: supplier supplier_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.supplier
    ADD CONSTRAINT supplier_pkey PRIMARY KEY (supplier_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


--
-- Name: warehouse warehouse_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.warehouse
    ADD CONSTRAINT warehouse_pkey PRIMARY KEY (warehouse_id);


--
-- Name: purchase_order set_delivery_date; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER set_delivery_date BEFORE UPDATE ON public.purchase_order FOR EACH ROW WHEN (((old.order_status)::text IS DISTINCT FROM (new.order_status)::text)) EXECUTE FUNCTION public.update_delivery_date();


--
-- Name: sales_record trigger_decrease_stock; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trigger_decrease_stock AFTER INSERT ON public.sales_record FOR EACH ROW EXECUTE FUNCTION public.decrease_stock_level();


--
-- Name: purchase_order trigger_update_inventory_on_purchase; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trigger_update_inventory_on_purchase AFTER UPDATE ON public.purchase_order FOR EACH ROW WHEN (((old.order_status)::text IS DISTINCT FROM (new.order_status)::text)) EXECUTE FUNCTION public.update_inventory_on_purchase_order();


--
-- Name: sales_record trigger_update_inventory_on_sale; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trigger_update_inventory_on_sale AFTER INSERT ON public.sales_record FOR EACH ROW EXECUTE FUNCTION public.update_inventory_on_sale();


--
-- Name: stock_transfer trigger_update_inventory_on_transfer; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trigger_update_inventory_on_transfer AFTER INSERT ON public.stock_transfer FOR EACH ROW EXECUTE FUNCTION public.update_inventory_on_transfer();


--
-- Name: attendance attendance_person_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.attendance
    ADD CONSTRAINT attendance_person_id_fkey FOREIGN KEY (person_id) REFERENCES public.person(person_id);


--
-- Name: budget budget_department_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.budget
    ADD CONSTRAINT budget_department_id_fkey FOREIGN KEY (department_id) REFERENCES public.department(department_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: employee employee_person_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employee
    ADD CONSTRAINT employee_person_id_fkey FOREIGN KEY (person_id) REFERENCES public.person(person_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: executive executive_person_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.executive
    ADD CONSTRAINT executive_person_id_fkey FOREIGN KEY (person_id) REFERENCES public.person(person_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: expenses expenses_budget_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.expenses
    ADD CONSTRAINT expenses_budget_id_fkey FOREIGN KEY (budget_id) REFERENCES public.budget(budget_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: expenses expenses_department_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.expenses
    ADD CONSTRAINT expenses_department_id_fkey FOREIGN KEY (department_id) REFERENCES public.department(department_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: expenses expenses_person_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.expenses
    ADD CONSTRAINT expenses_person_id_fkey FOREIGN KEY (person_id) REFERENCES public.person(person_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: financial_report financialreport_department_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.financial_report
    ADD CONSTRAINT financialreport_department_id_fkey FOREIGN KEY (department_id) REFERENCES public.department(department_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: financial_report financialreport_store_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.financial_report
    ADD CONSTRAINT financialreport_store_id_fkey FOREIGN KEY (store_id) REFERENCES public.store(store_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: department fk_department_manager; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.department
    ADD CONSTRAINT fk_department_manager FOREIGN KEY (manager_id) REFERENCES public.manager(manager_id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: location fk_location_manager; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.location
    ADD CONSTRAINT fk_location_manager FOREIGN KEY (manager_id) REFERENCES public.manager(manager_id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: inventory inventory_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inventory
    ADD CONSTRAINT inventory_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.location(location_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: inventory inventory_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inventory
    ADD CONSTRAINT inventory_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.product(product_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: manager manager_person_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.manager
    ADD CONSTRAINT manager_person_id_fkey FOREIGN KEY (person_id) REFERENCES public.person(person_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: payroll payroll_budget_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payroll
    ADD CONSTRAINT payroll_budget_id_fkey FOREIGN KEY (budget_id) REFERENCES public.budget(budget_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: payroll payroll_department_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payroll
    ADD CONSTRAINT payroll_department_id_fkey FOREIGN KEY (department_id) REFERENCES public.department(department_id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: payroll payroll_person_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payroll
    ADD CONSTRAINT payroll_person_id_fkey FOREIGN KEY (person_id) REFERENCES public.person(person_id) ON DELETE CASCADE;


--
-- Name: person person_department_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.person
    ADD CONSTRAINT person_department_id_fkey FOREIGN KEY (department_id) REFERENCES public.department(department_id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: person person_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.person
    ADD CONSTRAINT person_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.location(location_id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: product product_supplier_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product
    ADD CONSTRAINT product_supplier_id_fkey FOREIGN KEY (supplier_id) REFERENCES public.supplier(supplier_id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: purchase_order purchase_order_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.purchase_order
    ADD CONSTRAINT purchase_order_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.location(location_id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: purchase_order purchaseorder_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.purchase_order
    ADD CONSTRAINT purchaseorder_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.product(product_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: purchase_order purchaseorder_supplier_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.purchase_order
    ADD CONSTRAINT purchaseorder_supplier_id_fkey FOREIGN KEY (supplier_id) REFERENCES public.supplier(supplier_id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: sales_forecast salesforecast_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sales_forecast
    ADD CONSTRAINT salesforecast_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.product(product_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: sales_forecast salesforecast_store_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sales_forecast
    ADD CONSTRAINT salesforecast_store_id_fkey FOREIGN KEY (store_id) REFERENCES public.store(store_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: sales_record salesrecord_manager_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sales_record
    ADD CONSTRAINT salesrecord_manager_id_fkey FOREIGN KEY (manager_id) REFERENCES public.manager(manager_id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: sales_record salesrecord_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sales_record
    ADD CONSTRAINT salesrecord_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.product(product_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: sales_record salesrecord_store_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sales_record
    ADD CONSTRAINT salesrecord_store_id_fkey FOREIGN KEY (store_id) REFERENCES public.store(store_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: stock_transfer stocktransfer_destination_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_transfer
    ADD CONSTRAINT stocktransfer_destination_location_id_fkey FOREIGN KEY (destination_location_id) REFERENCES public.location(location_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: stock_transfer stocktransfer_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_transfer
    ADD CONSTRAINT stocktransfer_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.product(product_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: stock_transfer stocktransfer_source_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_transfer
    ADD CONSTRAINT stocktransfer_source_location_id_fkey FOREIGN KEY (source_location_id) REFERENCES public.location(location_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: store store_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.store
    ADD CONSTRAINT store_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.location(location_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: users users_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.location(location_id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: users users_person_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_person_id_fkey FOREIGN KEY (person_id) REFERENCES public.person(person_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: warehouse warehouse_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.warehouse
    ADD CONSTRAINT warehouse_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.location(location_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

