-- Downloaded from: https://github.com/HugoTZC/OASA/blob/ae5fe56afcf199a2df8dfcc376ae06c46838d856/backup_oasa.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 16.9 (Ubuntu 16.9-0ubuntu0.24.04.1)
-- Dumped by pg_dump version 16.9 (Ubuntu 16.9-0ubuntu0.24.04.1)

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
-- Name: dbo; Type: SCHEMA; Schema: -; Owner: pg_database_owner
--

CREATE SCHEMA dbo;


ALTER SCHEMA dbo OWNER TO pg_database_owner;

--
-- Name: SCHEMA dbo; Type: COMMENT; Schema: -; Owner: pg_database_owner
--

COMMENT ON SCHEMA dbo IS 'standard public schema';


--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: dbo; Owner: hugotzc
--

CREATE FUNCTION dbo.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION dbo.update_updated_at_column() OWNER TO hugotzc;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: blog_posts; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.blog_posts (
    id bigint NOT NULL,
    title character varying(255) NOT NULL,
    slug character varying(255) NOT NULL,
    excerpt character varying(500),
    content text NOT NULL,
    author_id bigint NOT NULL,
    featured_image_url character varying(500),
    status character varying(20) DEFAULT 'draft'::character varying NOT NULL,
    published_at timestamp with time zone,
    meta_title character varying(255),
    meta_description character varying(500),
    view_count integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT blog_posts_status_check CHECK (((status)::text = ANY ((ARRAY['draft'::character varying, 'published'::character varying, 'archived'::character varying])::text[])))
);


ALTER TABLE dbo.blog_posts OWNER TO hugotzc;

--
-- Name: blog_posts_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.blog_posts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.blog_posts_id_seq OWNER TO hugotzc;

--
-- Name: blog_posts_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.blog_posts_id_seq OWNED BY dbo.blog_posts.id;


--
-- Name: brands; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.brands (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    slug character varying(255) NOT NULL,
    description text,
    logo_url character varying(500),
    website_url character varying(500),
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE dbo.brands OWNER TO hugotzc;

--
-- Name: brands_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.brands_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.brands_id_seq OWNER TO hugotzc;

--
-- Name: brands_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.brands_id_seq OWNED BY dbo.brands.id;


--
-- Name: cart_items; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.cart_items (
    id bigint NOT NULL,
    user_id bigint,
    session_id character varying(255),
    product_id bigint NOT NULL,
    quantity integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE dbo.cart_items OWNER TO hugotzc;

--
-- Name: cart_items_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.cart_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.cart_items_id_seq OWNER TO hugotzc;

--
-- Name: cart_items_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.cart_items_id_seq OWNED BY dbo.cart_items.id;


--
-- Name: categories; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.categories (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    slug character varying(255) NOT NULL,
    description text,
    icon character varying(100),
    color character varying(20),
    parent_id bigint,
    display_order integer DEFAULT 0,
    is_active boolean DEFAULT true,
    meta_title character varying(255),
    meta_description character varying(500),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE dbo.categories OWNER TO hugotzc;

--
-- Name: categories_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.categories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.categories_id_seq OWNER TO hugotzc;

--
-- Name: categories_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.categories_id_seq OWNED BY dbo.categories.id;


--
-- Name: category_showcases; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.category_showcases (
    id bigint NOT NULL,
    category_id bigint NOT NULL,
    title character varying(255),
    description character varying(500),
    image_url character varying(500),
    display_order integer DEFAULT 0,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE dbo.category_showcases OWNER TO hugotzc;

--
-- Name: category_showcases_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.category_showcases_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.category_showcases_id_seq OWNER TO hugotzc;

--
-- Name: category_showcases_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.category_showcases_id_seq OWNED BY dbo.category_showcases.id;


--
-- Name: client_feature_access; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.client_feature_access (
    id bigint NOT NULL,
    client_id character varying(100) NOT NULL,
    feature_id bigint NOT NULL,
    is_enabled boolean NOT NULL,
    feature_limit integer,
    override_reason text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE dbo.client_feature_access OWNER TO hugotzc;

--
-- Name: client_feature_access_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.client_feature_access_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.client_feature_access_id_seq OWNER TO hugotzc;

--
-- Name: client_feature_access_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.client_feature_access_id_seq OWNED BY dbo.client_feature_access.id;


--
-- Name: client_subscriptions; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.client_subscriptions (
    id bigint NOT NULL,
    client_id character varying(100) NOT NULL,
    plan_id bigint NOT NULL,
    status character varying(50) DEFAULT 'active'::character varying,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT client_subscriptions_status_check CHECK (((status)::text = ANY ((ARRAY['active'::character varying, 'suspended'::character varying, 'cancelled'::character varying, 'expired'::character varying])::text[])))
);


ALTER TABLE dbo.client_subscriptions OWNER TO hugotzc;

--
-- Name: client_subscriptions_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.client_subscriptions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.client_subscriptions_id_seq OWNER TO hugotzc;

--
-- Name: client_subscriptions_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.client_subscriptions_id_seq OWNED BY dbo.client_subscriptions.id;


--
-- Name: coupons; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.coupons (
    id bigint NOT NULL,
    code character varying(50) NOT NULL,
    name character varying(255) NOT NULL,
    description character varying(500),
    type character varying(20) NOT NULL,
    value numeric(10,2) NOT NULL,
    minimum_amount numeric(10,2),
    maximum_discount numeric(10,2),
    usage_limit integer,
    used_count integer DEFAULT 0,
    is_active boolean DEFAULT true,
    starts_at timestamp with time zone,
    expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT coupons_type_check CHECK (((type)::text = ANY ((ARRAY['percentage'::character varying, 'fixed_amount'::character varying])::text[])))
);


ALTER TABLE dbo.coupons OWNER TO hugotzc;

--
-- Name: coupons_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.coupons_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.coupons_id_seq OWNER TO hugotzc;

--
-- Name: coupons_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.coupons_id_seq OWNED BY dbo.coupons.id;


--
-- Name: featured_product_sections; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.featured_product_sections (
    id bigint NOT NULL,
    product_id bigint NOT NULL,
    section_name character varying(100) NOT NULL,
    display_order integer DEFAULT 0,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE dbo.featured_product_sections OWNER TO hugotzc;

--
-- Name: featured_product_sections_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.featured_product_sections_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.featured_product_sections_id_seq OWNER TO hugotzc;

--
-- Name: featured_product_sections_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.featured_product_sections_id_seq OWNED BY dbo.featured_product_sections.id;


--
-- Name: features; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.features (
    id bigint NOT NULL,
    feature_key character varying(100) NOT NULL,
    feature_name character varying(255) NOT NULL,
    description text,
    category character varying(100) DEFAULT 'general'::character varying,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE dbo.features OWNER TO hugotzc;

--
-- Name: features_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.features_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.features_id_seq OWNER TO hugotzc;

--
-- Name: features_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.features_id_seq OWNED BY dbo.features.id;


--
-- Name: hero_slides; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.hero_slides (
    id bigint NOT NULL,
    title character varying(255) NOT NULL,
    subtitle character varying(255),
    description text,
    cta_text character varying(100),
    cta_url character varying(500),
    image_url character varying(500),
    background_color character varying(50),
    text_color character varying(50),
    display_order integer DEFAULT 0,
    is_active boolean DEFAULT true,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE dbo.hero_slides OWNER TO hugotzc;

--
-- Name: hero_slides_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.hero_slides_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.hero_slides_id_seq OWNER TO hugotzc;

--
-- Name: hero_slides_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.hero_slides_id_seq OWNED BY dbo.hero_slides.id;


--
-- Name: newsletter_subscribers; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.newsletter_subscribers (
    id bigint NOT NULL,
    email character varying(255) NOT NULL,
    name character varying(255),
    status character varying(20) DEFAULT 'active'::character varying NOT NULL,
    source character varying(100),
    subscribed_at timestamp with time zone DEFAULT now() NOT NULL,
    unsubscribed_at timestamp with time zone,
    CONSTRAINT newsletter_subscribers_status_check CHECK (((status)::text = ANY ((ARRAY['active'::character varying, 'unsubscribed'::character varying])::text[])))
);


ALTER TABLE dbo.newsletter_subscribers OWNER TO hugotzc;

--
-- Name: newsletter_subscribers_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.newsletter_subscribers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.newsletter_subscribers_id_seq OWNER TO hugotzc;

--
-- Name: newsletter_subscribers_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.newsletter_subscribers_id_seq OWNED BY dbo.newsletter_subscribers.id;


--
-- Name: order_items; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.order_items (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    product_id bigint NOT NULL,
    product_name character varying(255) NOT NULL,
    product_sku character varying(100) NOT NULL,
    quantity integer NOT NULL,
    unit_price numeric(10,2) NOT NULL,
    total_price numeric(10,2) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE dbo.order_items OWNER TO hugotzc;

--
-- Name: order_items_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.order_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.order_items_id_seq OWNER TO hugotzc;

--
-- Name: order_items_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.order_items_id_seq OWNED BY dbo.order_items.id;


--
-- Name: orders; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.orders (
    id bigint NOT NULL,
    user_id bigint,
    order_number character varying(50) NOT NULL,
    status character varying(50) DEFAULT 'pending'::character varying NOT NULL,
    total_amount numeric(10,2) NOT NULL,
    subtotal numeric(10,2) NOT NULL,
    tax_amount numeric(10,2) DEFAULT 0,
    shipping_amount numeric(10,2) DEFAULT 0,
    discount_amount numeric(10,2) DEFAULT 0,
    billing_name character varying(255) NOT NULL,
    billing_email character varying(255) NOT NULL,
    billing_phone character varying(50),
    billing_company character varying(255),
    billing_address_line1 character varying(255) NOT NULL,
    billing_address_line2 character varying(255),
    billing_city character varying(100) NOT NULL,
    billing_state character varying(100) NOT NULL,
    billing_postal_code character varying(20) NOT NULL,
    billing_country character varying(100) NOT NULL,
    shipping_name character varying(255),
    shipping_phone character varying(50),
    shipping_company character varying(255),
    shipping_address_line1 character varying(255),
    shipping_address_line2 character varying(255),
    shipping_city character varying(100),
    shipping_state character varying(100),
    shipping_postal_code character varying(20),
    shipping_country character varying(100),
    payment_method character varying(50),
    payment_status character varying(50) DEFAULT 'pending'::character varying,
    notes text,
    shipped_at timestamp with time zone,
    delivered_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT orders_payment_status_check CHECK (((payment_status)::text = ANY ((ARRAY['pending'::character varying, 'paid'::character varying, 'failed'::character varying, 'refunded'::character varying])::text[]))),
    CONSTRAINT orders_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'processing'::character varying, 'shipped'::character varying, 'delivered'::character varying, 'cancelled'::character varying, 'refunded'::character varying])::text[])))
);


ALTER TABLE dbo.orders OWNER TO hugotzc;

--
-- Name: orders_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.orders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.orders_id_seq OWNER TO hugotzc;

--
-- Name: orders_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.orders_id_seq OWNED BY dbo.orders.id;


--
-- Name: plan_features; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.plan_features (
    id bigint NOT NULL,
    plan_id bigint NOT NULL,
    feature_id bigint NOT NULL,
    is_enabled boolean DEFAULT true,
    feature_limit integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE dbo.plan_features OWNER TO hugotzc;

--
-- Name: plan_features_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.plan_features_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.plan_features_id_seq OWNER TO hugotzc;

--
-- Name: plan_features_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.plan_features_id_seq OWNED BY dbo.plan_features.id;


--
-- Name: product_attributes; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.product_attributes (
    id bigint NOT NULL,
    product_id bigint NOT NULL,
    attribute_name character varying(255) NOT NULL,
    attribute_value character varying(500) NOT NULL,
    display_order integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE dbo.product_attributes OWNER TO hugotzc;

--
-- Name: product_attributes_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.product_attributes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.product_attributes_id_seq OWNER TO hugotzc;

--
-- Name: product_attributes_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.product_attributes_id_seq OWNED BY dbo.product_attributes.id;


--
-- Name: product_images; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.product_images (
    id bigint NOT NULL,
    product_id bigint NOT NULL,
    url character varying(500) NOT NULL,
    alt_text character varying(255),
    display_order integer DEFAULT 0,
    is_primary boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE dbo.product_images OWNER TO hugotzc;

--
-- Name: product_images_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.product_images_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.product_images_id_seq OWNER TO hugotzc;

--
-- Name: product_images_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.product_images_id_seq OWNED BY dbo.product_images.id;


--
-- Name: product_reviews; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.product_reviews (
    id bigint NOT NULL,
    product_id bigint NOT NULL,
    user_id bigint NOT NULL,
    rating integer NOT NULL,
    title character varying(255),
    comment text,
    is_verified_purchase boolean DEFAULT false,
    is_approved boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT product_reviews_rating_check CHECK (((rating >= 1) AND (rating <= 5)))
);


ALTER TABLE dbo.product_reviews OWNER TO hugotzc;

--
-- Name: product_reviews_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.product_reviews_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.product_reviews_id_seq OWNER TO hugotzc;

--
-- Name: product_reviews_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.product_reviews_id_seq OWNED BY dbo.product_reviews.id;


--
-- Name: products; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.products (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    slug character varying(255) NOT NULL,
    sku character varying(100) NOT NULL,
    description text,
    short_description character varying(500),
    category_id bigint NOT NULL,
    brand_id bigint,
    price numeric(10,2) NOT NULL,
    original_price numeric(10,2),
    cost_price numeric(10,2),
    stock_quantity integer DEFAULT 0 NOT NULL,
    min_stock_level integer DEFAULT 0,
    weight numeric(8,2),
    dimensions_length numeric(8,2),
    dimensions_width numeric(8,2),
    dimensions_height numeric(8,2),
    is_featured boolean DEFAULT false,
    is_active boolean DEFAULT true,
    is_digital boolean DEFAULT false,
    requires_shipping boolean DEFAULT true,
    meta_title character varying(255),
    meta_description character varying(500),
    rating_average numeric(3,2) DEFAULT 0,
    rating_count integer DEFAULT 0,
    view_count integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE dbo.products OWNER TO hugotzc;

--
-- Name: products_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.products_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.products_id_seq OWNER TO hugotzc;

--
-- Name: products_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.products_id_seq OWNED BY dbo.products.id;


--
-- Name: site_settings; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.site_settings (
    id bigint NOT NULL,
    setting_key character varying(100) NOT NULL,
    setting_value text,
    setting_type character varying(50) DEFAULT 'text'::character varying,
    description character varying(500),
    is_public boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE dbo.site_settings OWNER TO hugotzc;

--
-- Name: site_settings_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.site_settings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.site_settings_id_seq OWNER TO hugotzc;

--
-- Name: site_settings_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.site_settings_id_seq OWNED BY dbo.site_settings.id;


--
-- Name: subscription_plans; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.subscription_plans (
    id bigint NOT NULL,
    plan_key character varying(100) NOT NULL,
    plan_name character varying(255) NOT NULL,
    description text,
    price_monthly numeric(10,2),
    price_yearly numeric(10,2),
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE dbo.subscription_plans OWNER TO hugotzc;

--
-- Name: subscription_plans_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.subscription_plans_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.subscription_plans_id_seq OWNER TO hugotzc;

--
-- Name: subscription_plans_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.subscription_plans_id_seq OWNED BY dbo.subscription_plans.id;


--
-- Name: user_profiles; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.user_profiles (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    company_name character varying(255),
    tax_id character varying(50),
    address_line1 character varying(255),
    address_line2 character varying(255),
    city character varying(100),
    state character varying(100),
    postal_code character varying(20),
    country character varying(100) DEFAULT 'México'::character varying,
    birthdate date,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE dbo.user_profiles OWNER TO hugotzc;

--
-- Name: user_profiles_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.user_profiles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.user_profiles_id_seq OWNER TO hugotzc;

--
-- Name: user_profiles_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.user_profiles_id_seq OWNED BY dbo.user_profiles.id;


--
-- Name: user_sessions; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.user_sessions (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    token character varying(255) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE dbo.user_sessions OWNER TO hugotzc;

--
-- Name: user_sessions_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.user_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.user_sessions_id_seq OWNER TO hugotzc;

--
-- Name: user_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.user_sessions_id_seq OWNED BY dbo.user_sessions.id;


--
-- Name: users; Type: TABLE; Schema: dbo; Owner: hugotzc
--

CREATE TABLE dbo.users (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    password_hash character varying(255) NOT NULL,
    role character varying(20) NOT NULL,
    status character varying(20) DEFAULT 'active'::character varying NOT NULL,
    email_verified boolean DEFAULT false,
    phone character varying(50),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT users_role_check CHECK (((role)::text = ANY ((ARRAY['admin'::character varying, 'customer'::character varying])::text[]))),
    CONSTRAINT users_status_check CHECK (((status)::text = ANY ((ARRAY['active'::character varying, 'inactive'::character varying, 'suspended'::character varying])::text[])))
);


ALTER TABLE dbo.users OWNER TO hugotzc;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: dbo; Owner: hugotzc
--

CREATE SEQUENCE dbo.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dbo.users_id_seq OWNER TO hugotzc;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: dbo; Owner: hugotzc
--

ALTER SEQUENCE dbo.users_id_seq OWNED BY dbo.users.id;


--
-- Name: blog_posts id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.blog_posts ALTER COLUMN id SET DEFAULT nextval('dbo.blog_posts_id_seq'::regclass);


--
-- Name: brands id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.brands ALTER COLUMN id SET DEFAULT nextval('dbo.brands_id_seq'::regclass);


--
-- Name: cart_items id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.cart_items ALTER COLUMN id SET DEFAULT nextval('dbo.cart_items_id_seq'::regclass);


--
-- Name: categories id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.categories ALTER COLUMN id SET DEFAULT nextval('dbo.categories_id_seq'::regclass);


--
-- Name: category_showcases id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.category_showcases ALTER COLUMN id SET DEFAULT nextval('dbo.category_showcases_id_seq'::regclass);


--
-- Name: client_feature_access id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.client_feature_access ALTER COLUMN id SET DEFAULT nextval('dbo.client_feature_access_id_seq'::regclass);


--
-- Name: client_subscriptions id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.client_subscriptions ALTER COLUMN id SET DEFAULT nextval('dbo.client_subscriptions_id_seq'::regclass);


--
-- Name: coupons id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.coupons ALTER COLUMN id SET DEFAULT nextval('dbo.coupons_id_seq'::regclass);


--
-- Name: featured_product_sections id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.featured_product_sections ALTER COLUMN id SET DEFAULT nextval('dbo.featured_product_sections_id_seq'::regclass);


--
-- Name: features id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.features ALTER COLUMN id SET DEFAULT nextval('dbo.features_id_seq'::regclass);


--
-- Name: hero_slides id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.hero_slides ALTER COLUMN id SET DEFAULT nextval('dbo.hero_slides_id_seq'::regclass);


--
-- Name: newsletter_subscribers id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.newsletter_subscribers ALTER COLUMN id SET DEFAULT nextval('dbo.newsletter_subscribers_id_seq'::regclass);


--
-- Name: order_items id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.order_items ALTER COLUMN id SET DEFAULT nextval('dbo.order_items_id_seq'::regclass);


--
-- Name: orders id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.orders ALTER COLUMN id SET DEFAULT nextval('dbo.orders_id_seq'::regclass);


--
-- Name: plan_features id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.plan_features ALTER COLUMN id SET DEFAULT nextval('dbo.plan_features_id_seq'::regclass);


--
-- Name: product_attributes id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.product_attributes ALTER COLUMN id SET DEFAULT nextval('dbo.product_attributes_id_seq'::regclass);


--
-- Name: product_images id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.product_images ALTER COLUMN id SET DEFAULT nextval('dbo.product_images_id_seq'::regclass);


--
-- Name: product_reviews id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.product_reviews ALTER COLUMN id SET DEFAULT nextval('dbo.product_reviews_id_seq'::regclass);


--
-- Name: products id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.products ALTER COLUMN id SET DEFAULT nextval('dbo.products_id_seq'::regclass);


--
-- Name: site_settings id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.site_settings ALTER COLUMN id SET DEFAULT nextval('dbo.site_settings_id_seq'::regclass);


--
-- Name: subscription_plans id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.subscription_plans ALTER COLUMN id SET DEFAULT nextval('dbo.subscription_plans_id_seq'::regclass);


--
-- Name: user_profiles id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.user_profiles ALTER COLUMN id SET DEFAULT nextval('dbo.user_profiles_id_seq'::regclass);


--
-- Name: user_sessions id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.user_sessions ALTER COLUMN id SET DEFAULT nextval('dbo.user_sessions_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.users ALTER COLUMN id SET DEFAULT nextval('dbo.users_id_seq'::regclass);


--
-- Data for Name: blog_posts; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.blog_posts (id, title, slug, excerpt, content, author_id, featured_image_url, status, published_at, meta_title, meta_description, view_count, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: brands; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.brands (id, name, slug, description, logo_url, website_url, is_active, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: cart_items; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.cart_items (id, user_id, session_id, product_id, quantity, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: categories; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.categories (id, name, slug, description, icon, color, parent_id, display_order, is_active, meta_title, meta_description, created_at, updated_at) FROM stdin;
1	Gases Industriales	gases-industriales	Oxígeno, argón, acetileno y más gases para uso industrial	Settings	bg-blue-800	\N	1	t	\N	\N	2025-07-10 20:54:46.171555-06	2025-07-10 20:54:46.171555-06
2	Equipos de Soldadura	equipos-soldadura	Soldadoras, electrodos y accesorios profesionales	Zap	bg-blue-800	\N	2	t	\N	\N	2025-07-10 20:54:46.171555-06	2025-07-10 20:54:46.171555-06
3	Herramientas	herramientas	Herramientas manuales y eléctricas de calidad	Wrench	bg-blue-800	\N	3	t	\N	\N	2025-07-10 20:54:46.171555-06	2025-07-10 20:54:46.171555-06
5	Equipos Neumáticos	equipos-neumaticos	Compresores, pistolas neumáticas y accesorios	Fuel	bg-blue-800	\N	5	t	\N	\N	2025-07-10 20:54:46.171555-06	2025-07-10 20:54:46.171555-06
6	Seguridad Industrial	seguridad-industrial	Equipos de protección personal y seguridad laboral	Shield	bg-blue-800	\N	6	t	\N	\N	2025-07-10 20:54:46.171555-06	2025-07-10 20:54:46.171555-06
\.


--
-- Data for Name: category_showcases; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.category_showcases (id, category_id, title, description, image_url, display_order, is_active, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: client_feature_access; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.client_feature_access (id, client_id, feature_id, is_enabled, feature_limit, override_reason, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: client_subscriptions; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.client_subscriptions (id, client_id, plan_id, status, started_at, expires_at, created_at, updated_at) FROM stdin;
1	oasa-default	2	active	2025-07-11 21:46:47.733018-06	\N	2025-07-11 21:46:47.733018-06	2025-07-11 21:46:47.733018-06
\.


--
-- Data for Name: coupons; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.coupons (id, code, name, description, type, value, minimum_amount, maximum_discount, usage_limit, used_count, is_active, starts_at, expires_at, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: featured_product_sections; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.featured_product_sections (id, product_id, section_name, display_order, is_active, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: features; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.features (id, feature_key, feature_name, description, category, is_active, created_at, updated_at) FROM stdin;
1	shopping_cart	Shopping Cart	Enable shopping cart functionality	ecommerce	t	2025-07-11 21:46:47.688231-06	2025-07-11 21:46:47.688231-06
2	product_pricing	Product Pricing	Display product prices	ecommerce	t	2025-07-11 21:46:47.688231-06	2025-07-11 21:46:47.688231-06
3	add_to_cart	Add to Cart	Allow products to be added to cart	ecommerce	t	2025-07-11 21:46:47.688231-06	2025-07-11 21:46:47.688231-06
4	checkout_process	Checkout Process	Enable checkout and payment processing	ecommerce	t	2025-07-11 21:46:47.688231-06	2025-07-11 21:46:47.688231-06
5	user_accounts	User Accounts	Enable user registration and login	users	t	2025-07-11 21:46:47.688231-06	2025-07-11 21:46:47.688231-06
6	admin_panel	Admin Panel	Access to administration interface	admin	t	2025-07-11 21:46:47.688231-06	2025-07-11 21:46:47.688231-06
7	analytics	Analytics	Advanced analytics and reporting	analytics	t	2025-07-11 21:46:47.688231-06	2025-07-11 21:46:47.688231-06
8	multi_language	Multi Language	Multiple language support	content	t	2025-07-11 21:46:47.688231-06	2025-07-11 21:46:47.688231-06
9	custom_branding	Custom Branding	Custom colors, logos, and themes	branding	t	2025-07-11 21:46:47.688231-06	2025-07-11 21:46:47.688231-06
10	email_marketing	Email Marketing	Email campaigns and newsletters	marketing	t	2025-07-11 21:46:47.688231-06	2025-07-11 21:46:47.688231-06
11	seo_tools	SEO Tools	Search engine optimization tools	marketing	t	2025-07-11 21:46:47.688231-06	2025-07-11 21:46:47.688231-06
12	inventory_management	Inventory Management	Stock tracking and management	inventory	t	2025-07-11 21:46:47.688231-06	2025-07-11 21:46:47.688231-06
13	order_management	Order Management	Order processing and tracking	orders	t	2025-07-11 21:46:47.688231-06	2025-07-11 21:46:47.688231-06
14	payment_gateways	Payment Gateways	Multiple payment methods	payments	t	2025-07-11 21:46:47.688231-06	2025-07-11 21:46:47.688231-06
15	shipping_integration	Shipping Integration	Shipping providers integration	shipping	t	2025-07-11 21:46:47.688231-06	2025-07-11 21:46:47.688231-06
\.


--
-- Data for Name: hero_slides; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.hero_slides (id, title, subtitle, description, cta_text, cta_url, image_url, background_color, text_color, display_order, is_active, start_date, end_date, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: newsletter_subscribers; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.newsletter_subscribers (id, email, name, status, source, subscribed_at, unsubscribed_at) FROM stdin;
\.


--
-- Data for Name: order_items; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.order_items (id, order_id, product_id, product_name, product_sku, quantity, unit_price, total_price, created_at) FROM stdin;
\.


--
-- Data for Name: orders; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.orders (id, user_id, order_number, status, total_amount, subtotal, tax_amount, shipping_amount, discount_amount, billing_name, billing_email, billing_phone, billing_company, billing_address_line1, billing_address_line2, billing_city, billing_state, billing_postal_code, billing_country, shipping_name, shipping_phone, shipping_company, shipping_address_line1, shipping_address_line2, shipping_city, shipping_state, shipping_postal_code, shipping_country, payment_method, payment_status, notes, shipped_at, delivered_at, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: plan_features; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.plan_features (id, plan_id, feature_id, is_enabled, feature_limit, created_at) FROM stdin;
1	1	2	t	\N	2025-07-11 21:46:47.713699-06
2	1	5	t	\N	2025-07-11 21:46:47.713699-06
3	1	9	t	\N	2025-07-11 21:46:47.713699-06
4	2	1	t	\N	2025-07-11 21:46:47.713699-06
5	2	2	t	\N	2025-07-11 21:46:47.713699-06
6	2	3	t	\N	2025-07-11 21:46:47.713699-06
7	2	4	t	\N	2025-07-11 21:46:47.713699-06
8	2	5	t	\N	2025-07-11 21:46:47.713699-06
9	2	6	t	\N	2025-07-11 21:46:47.713699-06
10	2	9	t	\N	2025-07-11 21:46:47.713699-06
11	2	10	t	\N	2025-07-11 21:46:47.713699-06
12	2	12	t	\N	2025-07-11 21:46:47.713699-06
13	2	13	t	\N	2025-07-11 21:46:47.713699-06
14	2	14	t	\N	2025-07-11 21:46:47.713699-06
15	3	1	t	\N	2025-07-11 21:46:47.713699-06
16	3	2	t	\N	2025-07-11 21:46:47.713699-06
17	3	3	t	\N	2025-07-11 21:46:47.713699-06
18	3	4	t	\N	2025-07-11 21:46:47.713699-06
19	3	5	t	\N	2025-07-11 21:46:47.713699-06
20	3	6	t	\N	2025-07-11 21:46:47.713699-06
21	3	7	t	\N	2025-07-11 21:46:47.713699-06
22	3	8	t	\N	2025-07-11 21:46:47.713699-06
23	3	9	t	\N	2025-07-11 21:46:47.713699-06
24	3	10	t	\N	2025-07-11 21:46:47.713699-06
25	3	11	t	\N	2025-07-11 21:46:47.713699-06
26	3	12	t	\N	2025-07-11 21:46:47.713699-06
27	3	13	t	\N	2025-07-11 21:46:47.713699-06
28	3	14	t	\N	2025-07-11 21:46:47.713699-06
29	3	15	t	\N	2025-07-11 21:46:47.713699-06
\.


--
-- Data for Name: product_attributes; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.product_attributes (id, product_id, attribute_name, attribute_value, display_order, created_at) FROM stdin;
\.


--
-- Data for Name: product_images; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.product_images (id, product_id, url, alt_text, display_order, is_primary, created_at) FROM stdin;
\.


--
-- Data for Name: product_reviews; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.product_reviews (id, product_id, user_id, rating, title, comment, is_verified_purchase, is_approved, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: products; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.products (id, name, slug, sku, description, short_description, category_id, brand_id, price, original_price, cost_price, stock_quantity, min_stock_level, weight, dimensions_length, dimensions_width, dimensions_height, is_featured, is_active, is_digital, requires_shipping, meta_title, meta_description, rating_average, rating_count, view_count, created_at, updated_at) FROM stdin;
1	Tanque de Oxígeno Industrial 50L	tanque-oxigeno-industrial-50l	OXI-50L-001	Tanque de oxígeno de alta calidad para uso industrial y médico. Fabricado con materiales resistentes y certificado para uso profesional. Capacidad de 50 litros con presión máxima de 200 bar.	Tanque de oxígeno industrial de 50 litros para uso profesional.	1	\N	1250.00	1400.00	\N	15	0	\N	\N	\N	\N	t	t	f	t	\N	\N	4.50	23	0	2025-07-15 22:12:59.476643-06	2025-07-15 22:12:59.476643-06
2	Soldadora MIG/MAG 200A	soldadora-mig-mag-200a	SOL-MIG-200	Soldadora MIG/MAG profesional de 200 amperios. Incluye antorcha, cables y accesorios. Ideal para trabajos de soldadura semi-automática en acero y aluminio.	Soldadora MIG/MAG profesional de 200A con accesorios completos.	2	\N	8500.00	\N	\N	5	0	\N	\N	\N	\N	t	t	f	t	\N	\N	4.80	45	0	2025-07-15 22:12:59.508113-06	2025-07-15 22:12:59.508113-06
3	Kit Herramientas Mecánico 150 Pzs	kit-herramientas-mecanico-150pzs	KIT-MEC-150	Kit completo de herramientas mecánicas con 150 piezas. Incluye llaves, dados, destornilladores y accesorios en maletín resistente.	Kit completo de 150 herramientas mecánicas en maletín.	3	\N	2100.00	2400.00	\N	10	0	\N	\N	\N	\N	t	t	f	t	\N	\N	4.30	156	0	2025-07-15 22:12:59.523535-06	2025-07-15 22:12:59.523535-06
4	Compresor de Aire 100L 3HP	compresor-aire-100l-3hp	COM-AIR-100	Compresor de aire de 100 litros con motor de 3HP. Presión máxima 8 bar. Ideal para uso profesional e industrial.	Compresor de aire profesional de 100L y 3HP.	5	\N	12500.00	14000.00	\N	3	0	\N	\N	\N	\N	t	t	f	t	\N	\N	4.70	28	0	2025-07-15 22:12:59.536448-06	2025-07-15 22:12:59.536448-06
5	Regulador de Presión Argón	regulador-presion-argon	REG-ARG-001	Regulador de presión profesional para argón. Construcción robusta en latón con manómetros de alta precisión. Ideal para soldadura TIG y MIG.	Regulador de presión profesional para argón con manómetros de precisión.	1	\N	950.00	\N	\N	8	0	\N	\N	\N	\N	f	t	f	t	\N	\N	4.60	34	0	2025-07-15 22:12:59.550786-06	2025-07-15 22:12:59.550786-06
6	Electrodo 6013 3.2mm (5kg)	electrodo-6013-32mm-5kg	ELE-6013-32	Electrodos de soldadura 6013 de 3.2mm. Paquete de 5kg ideal para soldadura de mantenimiento y reparación general.	Electrodos 6013 de 3.2mm, paquete de 5kg para soldadura general.	2	\N	320.00	380.00	\N	25	0	\N	\N	\N	\N	f	t	f	t	\N	\N	4.20	67	0	2025-07-15 22:12:59.564032-06	2025-07-15 22:12:59.564032-06
\.


--
-- Data for Name: site_settings; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.site_settings (id, setting_key, setting_value, setting_type, description, is_public, created_at, updated_at) FROM stdin;
1	site_name	OASA	text	Company name	t	2025-07-10 20:54:46.171555-06	2025-07-10 20:54:46.171555-06
2	site_tagline	Tu proveedor de confianza en equipos industriales	text	Company tagline	t	2025-07-10 20:54:46.171555-06	2025-07-10 20:54:46.171555-06
3	contact_email	contacto@oasa.com	text	Main contact email	t	2025-07-10 20:54:46.171555-06	2025-07-10 20:54:46.171555-06
4	contact_phone	+52 55 1234 5678	text	Main contact phone	t	2025-07-10 20:54:46.171555-06	2025-07-10 20:54:46.171555-06
5	currency	MXN	text	Default currency	t	2025-07-10 20:54:46.171555-06	2025-07-10 20:54:46.171555-06
6	tax_rate	16	number	Tax rate percentage	f	2025-07-10 20:54:46.171555-06	2025-07-10 20:54:46.171555-06
7	enable_shopping	true	boolean	Enable/disable shopping cart and e-commerce features	t	2025-07-11 07:43:36.958967-06	2025-07-11 21:36:45.96476-06
8	enable_pricing	true	boolean	Show/hide product prices	t	2025-07-11 07:43:36.958967-06	2025-07-11 21:36:45.96476-06
9	enable_add_to_cart	true	boolean	Enable/disable add to cart functionality	t	2025-07-11 07:43:36.958967-06	2025-07-11 21:36:45.96476-06
10	enable_checkout	true	boolean	Enable/disable checkout process	f	2025-07-11 07:43:36.958967-06	2025-07-11 21:36:45.96476-06
11	shopping_mode	full	text	Shopping mode: full, catalog, or disabled	t	2025-07-11 07:43:36.958967-06	2025-07-11 21:36:45.96476-06
\.


--
-- Data for Name: subscription_plans; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.subscription_plans (id, plan_key, plan_name, description, price_monthly, price_yearly, is_active, created_at, updated_at) FROM stdin;
1	basic	Basic Plan	Essential features for small businesses	29.99	299.99	t	2025-07-11 21:46:47.700498-06	2025-07-11 21:46:47.700498-06
2	professional	Professional Plan	Advanced features for growing businesses	79.99	799.99	t	2025-07-11 21:46:47.700498-06	2025-07-11 21:46:47.700498-06
3	enterprise	Enterprise Plan	Full featured plan for large businesses	199.99	1999.99	t	2025-07-11 21:46:47.700498-06	2025-07-11 21:46:47.700498-06
4	custom	Custom Plan	Tailored features for specific needs	\N	\N	t	2025-07-11 21:46:47.700498-06	2025-07-11 21:46:47.700498-06
\.


--
-- Data for Name: user_profiles; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.user_profiles (id, user_id, company_name, tax_id, address_line1, address_line2, city, state, postal_code, country, birthdate, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: user_sessions; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.user_sessions (id, user_id, token, expires_at, created_at) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: dbo; Owner: hugotzc
--

COPY dbo.users (id, name, email, password_hash, role, status, email_verified, phone, created_at, updated_at) FROM stdin;
1	Administrador	admin@oasa.com	hashed_password_here	admin	active	t	\N	2025-07-10 20:54:46.171555-06	2025-07-10 20:54:46.171555-06
\.


--
-- Name: blog_posts_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.blog_posts_id_seq', 1, false);


--
-- Name: brands_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.brands_id_seq', 1, false);


--
-- Name: cart_items_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.cart_items_id_seq', 1, false);


--
-- Name: categories_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.categories_id_seq', 6, true);


--
-- Name: category_showcases_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.category_showcases_id_seq', 1, false);


--
-- Name: client_feature_access_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.client_feature_access_id_seq', 1, false);


--
-- Name: client_subscriptions_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.client_subscriptions_id_seq', 1, true);


--
-- Name: coupons_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.coupons_id_seq', 1, false);


--
-- Name: featured_product_sections_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.featured_product_sections_id_seq', 1, false);


--
-- Name: features_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.features_id_seq', 15, true);


--
-- Name: hero_slides_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.hero_slides_id_seq', 1, false);


--
-- Name: newsletter_subscribers_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.newsletter_subscribers_id_seq', 1, false);


--
-- Name: order_items_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.order_items_id_seq', 1, false);


--
-- Name: orders_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.orders_id_seq', 1, false);


--
-- Name: plan_features_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.plan_features_id_seq', 29, true);


--
-- Name: product_attributes_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.product_attributes_id_seq', 1, false);


--
-- Name: product_images_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.product_images_id_seq', 1, false);


--
-- Name: product_reviews_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.product_reviews_id_seq', 1, false);


--
-- Name: products_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.products_id_seq', 6, true);


--
-- Name: site_settings_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.site_settings_id_seq', 16, true);


--
-- Name: subscription_plans_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.subscription_plans_id_seq', 4, true);


--
-- Name: user_profiles_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.user_profiles_id_seq', 1, false);


--
-- Name: user_sessions_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.user_sessions_id_seq', 1, false);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: dbo; Owner: hugotzc
--

SELECT pg_catalog.setval('dbo.users_id_seq', 1, true);


--
-- Name: blog_posts blog_posts_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.blog_posts
    ADD CONSTRAINT blog_posts_pkey PRIMARY KEY (id);


--
-- Name: blog_posts blog_posts_slug_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.blog_posts
    ADD CONSTRAINT blog_posts_slug_key UNIQUE (slug);


--
-- Name: brands brands_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.brands
    ADD CONSTRAINT brands_pkey PRIMARY KEY (id);


--
-- Name: brands brands_slug_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.brands
    ADD CONSTRAINT brands_slug_key UNIQUE (slug);


--
-- Name: cart_items cart_items_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.cart_items
    ADD CONSTRAINT cart_items_pkey PRIMARY KEY (id);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: categories categories_slug_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.categories
    ADD CONSTRAINT categories_slug_key UNIQUE (slug);


--
-- Name: category_showcases category_showcases_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.category_showcases
    ADD CONSTRAINT category_showcases_pkey PRIMARY KEY (id);


--
-- Name: client_feature_access client_feature_access_client_id_feature_id_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.client_feature_access
    ADD CONSTRAINT client_feature_access_client_id_feature_id_key UNIQUE (client_id, feature_id);


--
-- Name: client_feature_access client_feature_access_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.client_feature_access
    ADD CONSTRAINT client_feature_access_pkey PRIMARY KEY (id);


--
-- Name: client_subscriptions client_subscriptions_client_id_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.client_subscriptions
    ADD CONSTRAINT client_subscriptions_client_id_key UNIQUE (client_id);


--
-- Name: client_subscriptions client_subscriptions_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.client_subscriptions
    ADD CONSTRAINT client_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: coupons coupons_code_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.coupons
    ADD CONSTRAINT coupons_code_key UNIQUE (code);


--
-- Name: coupons coupons_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.coupons
    ADD CONSTRAINT coupons_pkey PRIMARY KEY (id);


--
-- Name: featured_product_sections featured_product_sections_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.featured_product_sections
    ADD CONSTRAINT featured_product_sections_pkey PRIMARY KEY (id);


--
-- Name: features features_feature_key_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.features
    ADD CONSTRAINT features_feature_key_key UNIQUE (feature_key);


--
-- Name: features features_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.features
    ADD CONSTRAINT features_pkey PRIMARY KEY (id);


--
-- Name: hero_slides hero_slides_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.hero_slides
    ADD CONSTRAINT hero_slides_pkey PRIMARY KEY (id);


--
-- Name: newsletter_subscribers newsletter_subscribers_email_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.newsletter_subscribers
    ADD CONSTRAINT newsletter_subscribers_email_key UNIQUE (email);


--
-- Name: newsletter_subscribers newsletter_subscribers_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.newsletter_subscribers
    ADD CONSTRAINT newsletter_subscribers_pkey PRIMARY KEY (id);


--
-- Name: order_items order_items_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);


--
-- Name: orders orders_order_number_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.orders
    ADD CONSTRAINT orders_order_number_key UNIQUE (order_number);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: plan_features plan_features_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.plan_features
    ADD CONSTRAINT plan_features_pkey PRIMARY KEY (id);


--
-- Name: plan_features plan_features_plan_id_feature_id_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.plan_features
    ADD CONSTRAINT plan_features_plan_id_feature_id_key UNIQUE (plan_id, feature_id);


--
-- Name: product_attributes product_attributes_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.product_attributes
    ADD CONSTRAINT product_attributes_pkey PRIMARY KEY (id);


--
-- Name: product_images product_images_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.product_images
    ADD CONSTRAINT product_images_pkey PRIMARY KEY (id);


--
-- Name: product_reviews product_reviews_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.product_reviews
    ADD CONSTRAINT product_reviews_pkey PRIMARY KEY (id);


--
-- Name: product_reviews product_reviews_product_id_user_id_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.product_reviews
    ADD CONSTRAINT product_reviews_product_id_user_id_key UNIQUE (product_id, user_id);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: products products_sku_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.products
    ADD CONSTRAINT products_sku_key UNIQUE (sku);


--
-- Name: products products_slug_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.products
    ADD CONSTRAINT products_slug_key UNIQUE (slug);


--
-- Name: site_settings site_settings_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.site_settings
    ADD CONSTRAINT site_settings_pkey PRIMARY KEY (id);


--
-- Name: site_settings site_settings_setting_key_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.site_settings
    ADD CONSTRAINT site_settings_setting_key_key UNIQUE (setting_key);


--
-- Name: subscription_plans subscription_plans_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.subscription_plans
    ADD CONSTRAINT subscription_plans_pkey PRIMARY KEY (id);


--
-- Name: subscription_plans subscription_plans_plan_key_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.subscription_plans
    ADD CONSTRAINT subscription_plans_plan_key_key UNIQUE (plan_key);


--
-- Name: user_profiles user_profiles_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.user_profiles
    ADD CONSTRAINT user_profiles_pkey PRIMARY KEY (id);


--
-- Name: user_profiles user_profiles_user_id_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.user_profiles
    ADD CONSTRAINT user_profiles_user_id_key UNIQUE (user_id);


--
-- Name: user_sessions user_sessions_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.user_sessions
    ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);


--
-- Name: user_sessions user_sessions_token_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.user_sessions
    ADD CONSTRAINT user_sessions_token_key UNIQUE (token);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_cart_session_id; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_cart_session_id ON dbo.cart_items USING btree (session_id);


--
-- Name: idx_cart_user_id; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_cart_user_id ON dbo.cart_items USING btree (user_id);


--
-- Name: idx_categories_is_active; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_categories_is_active ON dbo.categories USING btree (is_active);


--
-- Name: idx_categories_parent_id; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_categories_parent_id ON dbo.categories USING btree (parent_id);


--
-- Name: idx_categories_slug; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_categories_slug ON dbo.categories USING btree (slug);


--
-- Name: idx_client_feature_access_client_id; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_client_feature_access_client_id ON dbo.client_feature_access USING btree (client_id);


--
-- Name: idx_client_subscriptions_client_id; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_client_subscriptions_client_id ON dbo.client_subscriptions USING btree (client_id);


--
-- Name: idx_client_subscriptions_status; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_client_subscriptions_status ON dbo.client_subscriptions USING btree (status);


--
-- Name: idx_features_category; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_features_category ON dbo.features USING btree (category);


--
-- Name: idx_features_is_active; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_features_is_active ON dbo.features USING btree (is_active);


--
-- Name: idx_orders_created_at; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_orders_created_at ON dbo.orders USING btree (created_at);


--
-- Name: idx_orders_order_number; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_orders_order_number ON dbo.orders USING btree (order_number);


--
-- Name: idx_orders_status; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_orders_status ON dbo.orders USING btree (status);


--
-- Name: idx_orders_user_id; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_orders_user_id ON dbo.orders USING btree (user_id);


--
-- Name: idx_plan_features_plan_id; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_plan_features_plan_id ON dbo.plan_features USING btree (plan_id);


--
-- Name: idx_products_brand_id; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_products_brand_id ON dbo.products USING btree (brand_id);


--
-- Name: idx_products_category_id; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_products_category_id ON dbo.products USING btree (category_id);


--
-- Name: idx_products_is_active; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_products_is_active ON dbo.products USING btree (is_active);


--
-- Name: idx_products_is_featured; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_products_is_featured ON dbo.products USING btree (is_featured);


--
-- Name: idx_products_price; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_products_price ON dbo.products USING btree (price);


--
-- Name: idx_products_sku; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_products_sku ON dbo.products USING btree (sku);


--
-- Name: idx_products_slug; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_products_slug ON dbo.products USING btree (slug);


--
-- Name: idx_products_stock_quantity; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_products_stock_quantity ON dbo.products USING btree (stock_quantity);


--
-- Name: idx_reviews_product_id; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_reviews_product_id ON dbo.product_reviews USING btree (product_id);


--
-- Name: idx_reviews_rating; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_reviews_rating ON dbo.product_reviews USING btree (rating);


--
-- Name: idx_reviews_user_id; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_reviews_user_id ON dbo.product_reviews USING btree (user_id);


--
-- Name: idx_users_email; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_users_email ON dbo.users USING btree (email);


--
-- Name: idx_users_role; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_users_role ON dbo.users USING btree (role);


--
-- Name: idx_users_status; Type: INDEX; Schema: dbo; Owner: hugotzc
--

CREATE INDEX idx_users_status ON dbo.users USING btree (status);


--
-- Name: blog_posts update_blog_posts_updated_at; Type: TRIGGER; Schema: dbo; Owner: hugotzc
--

CREATE TRIGGER update_blog_posts_updated_at BEFORE UPDATE ON dbo.blog_posts FOR EACH ROW EXECUTE FUNCTION dbo.update_updated_at_column();


--
-- Name: brands update_brands_updated_at; Type: TRIGGER; Schema: dbo; Owner: hugotzc
--

CREATE TRIGGER update_brands_updated_at BEFORE UPDATE ON dbo.brands FOR EACH ROW EXECUTE FUNCTION dbo.update_updated_at_column();


--
-- Name: cart_items update_cart_items_updated_at; Type: TRIGGER; Schema: dbo; Owner: hugotzc
--

CREATE TRIGGER update_cart_items_updated_at BEFORE UPDATE ON dbo.cart_items FOR EACH ROW EXECUTE FUNCTION dbo.update_updated_at_column();


--
-- Name: categories update_categories_updated_at; Type: TRIGGER; Schema: dbo; Owner: hugotzc
--

CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON dbo.categories FOR EACH ROW EXECUTE FUNCTION dbo.update_updated_at_column();


--
-- Name: category_showcases update_category_showcases_updated_at; Type: TRIGGER; Schema: dbo; Owner: hugotzc
--

CREATE TRIGGER update_category_showcases_updated_at BEFORE UPDATE ON dbo.category_showcases FOR EACH ROW EXECUTE FUNCTION dbo.update_updated_at_column();


--
-- Name: coupons update_coupons_updated_at; Type: TRIGGER; Schema: dbo; Owner: hugotzc
--

CREATE TRIGGER update_coupons_updated_at BEFORE UPDATE ON dbo.coupons FOR EACH ROW EXECUTE FUNCTION dbo.update_updated_at_column();


--
-- Name: featured_product_sections update_featured_product_sections_updated_at; Type: TRIGGER; Schema: dbo; Owner: hugotzc
--

CREATE TRIGGER update_featured_product_sections_updated_at BEFORE UPDATE ON dbo.featured_product_sections FOR EACH ROW EXECUTE FUNCTION dbo.update_updated_at_column();


--
-- Name: hero_slides update_hero_slides_updated_at; Type: TRIGGER; Schema: dbo; Owner: hugotzc
--

CREATE TRIGGER update_hero_slides_updated_at BEFORE UPDATE ON dbo.hero_slides FOR EACH ROW EXECUTE FUNCTION dbo.update_updated_at_column();


--
-- Name: orders update_orders_updated_at; Type: TRIGGER; Schema: dbo; Owner: hugotzc
--

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON dbo.orders FOR EACH ROW EXECUTE FUNCTION dbo.update_updated_at_column();


--
-- Name: product_reviews update_product_reviews_updated_at; Type: TRIGGER; Schema: dbo; Owner: hugotzc
--

CREATE TRIGGER update_product_reviews_updated_at BEFORE UPDATE ON dbo.product_reviews FOR EACH ROW EXECUTE FUNCTION dbo.update_updated_at_column();


--
-- Name: products update_products_updated_at; Type: TRIGGER; Schema: dbo; Owner: hugotzc
--

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON dbo.products FOR EACH ROW EXECUTE FUNCTION dbo.update_updated_at_column();


--
-- Name: site_settings update_site_settings_updated_at; Type: TRIGGER; Schema: dbo; Owner: hugotzc
--

CREATE TRIGGER update_site_settings_updated_at BEFORE UPDATE ON dbo.site_settings FOR EACH ROW EXECUTE FUNCTION dbo.update_updated_at_column();


--
-- Name: user_profiles update_user_profiles_updated_at; Type: TRIGGER; Schema: dbo; Owner: hugotzc
--

CREATE TRIGGER update_user_profiles_updated_at BEFORE UPDATE ON dbo.user_profiles FOR EACH ROW EXECUTE FUNCTION dbo.update_updated_at_column();


--
-- Name: users update_users_updated_at; Type: TRIGGER; Schema: dbo; Owner: hugotzc
--

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON dbo.users FOR EACH ROW EXECUTE FUNCTION dbo.update_updated_at_column();


--
-- Name: blog_posts blog_posts_author_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.blog_posts
    ADD CONSTRAINT blog_posts_author_id_fkey FOREIGN KEY (author_id) REFERENCES dbo.users(id);


--
-- Name: cart_items cart_items_product_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.cart_items
    ADD CONSTRAINT cart_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES dbo.products(id) ON DELETE CASCADE;


--
-- Name: cart_items cart_items_user_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.cart_items
    ADD CONSTRAINT cart_items_user_id_fkey FOREIGN KEY (user_id) REFERENCES dbo.users(id) ON DELETE CASCADE;


--
-- Name: categories categories_parent_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.categories
    ADD CONSTRAINT categories_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES dbo.categories(id) ON DELETE SET NULL;


--
-- Name: category_showcases category_showcases_category_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.category_showcases
    ADD CONSTRAINT category_showcases_category_id_fkey FOREIGN KEY (category_id) REFERENCES dbo.categories(id) ON DELETE CASCADE;


--
-- Name: client_feature_access client_feature_access_feature_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.client_feature_access
    ADD CONSTRAINT client_feature_access_feature_id_fkey FOREIGN KEY (feature_id) REFERENCES dbo.features(id) ON DELETE CASCADE;


--
-- Name: client_subscriptions client_subscriptions_plan_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.client_subscriptions
    ADD CONSTRAINT client_subscriptions_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES dbo.subscription_plans(id);


--
-- Name: featured_product_sections featured_product_sections_product_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.featured_product_sections
    ADD CONSTRAINT featured_product_sections_product_id_fkey FOREIGN KEY (product_id) REFERENCES dbo.products(id) ON DELETE CASCADE;


--
-- Name: order_items order_items_order_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.order_items
    ADD CONSTRAINT order_items_order_id_fkey FOREIGN KEY (order_id) REFERENCES dbo.orders(id) ON DELETE CASCADE;


--
-- Name: order_items order_items_product_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.order_items
    ADD CONSTRAINT order_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES dbo.products(id);


--
-- Name: orders orders_user_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.orders
    ADD CONSTRAINT orders_user_id_fkey FOREIGN KEY (user_id) REFERENCES dbo.users(id) ON DELETE SET NULL;


--
-- Name: plan_features plan_features_feature_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.plan_features
    ADD CONSTRAINT plan_features_feature_id_fkey FOREIGN KEY (feature_id) REFERENCES dbo.features(id) ON DELETE CASCADE;


--
-- Name: plan_features plan_features_plan_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.plan_features
    ADD CONSTRAINT plan_features_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES dbo.subscription_plans(id) ON DELETE CASCADE;


--
-- Name: product_attributes product_attributes_product_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.product_attributes
    ADD CONSTRAINT product_attributes_product_id_fkey FOREIGN KEY (product_id) REFERENCES dbo.products(id) ON DELETE CASCADE;


--
-- Name: product_images product_images_product_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.product_images
    ADD CONSTRAINT product_images_product_id_fkey FOREIGN KEY (product_id) REFERENCES dbo.products(id) ON DELETE CASCADE;


--
-- Name: product_reviews product_reviews_product_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.product_reviews
    ADD CONSTRAINT product_reviews_product_id_fkey FOREIGN KEY (product_id) REFERENCES dbo.products(id) ON DELETE CASCADE;


--
-- Name: product_reviews product_reviews_user_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.product_reviews
    ADD CONSTRAINT product_reviews_user_id_fkey FOREIGN KEY (user_id) REFERENCES dbo.users(id) ON DELETE CASCADE;


--
-- Name: products products_brand_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.products
    ADD CONSTRAINT products_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES dbo.brands(id) ON DELETE SET NULL;


--
-- Name: products products_category_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.products
    ADD CONSTRAINT products_category_id_fkey FOREIGN KEY (category_id) REFERENCES dbo.categories(id);


--
-- Name: user_profiles user_profiles_user_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.user_profiles
    ADD CONSTRAINT user_profiles_user_id_fkey FOREIGN KEY (user_id) REFERENCES dbo.users(id) ON DELETE CASCADE;


--
-- Name: user_sessions user_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: dbo; Owner: hugotzc
--

ALTER TABLE ONLY dbo.user_sessions
    ADD CONSTRAINT user_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES dbo.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

