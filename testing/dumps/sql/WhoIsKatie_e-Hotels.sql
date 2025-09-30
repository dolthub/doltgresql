-- Downloaded from: https://github.com/WhoIsKatie/e-Hotels/blob/83bbd189f6f8ee312e8fca8830e33ae67b18558b/server/database.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 15.2
-- Dumped by pg_dump version 15.2

-- Started on 2023-03-31 21:08:48

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
-- TOC entry 3475 (class 1262 OID 16899)
-- Name: e-Hotels; Type: DATABASE; Schema: -; Owner: postgres
--

CREATE DATABASE "e-Hotels" WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'English_United States.1252';


ALTER DATABASE "e-Hotels" OWNER TO postgres;

\connect -reuse-previous=on "dbname='e-Hotels'"

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
-- TOC entry 238 (class 1255 OID 17123)
-- Name: employee_is_manager(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.employee_is_manager() RETURNS trigger
    LANGUAGE plpgsql
    AS $$BEGIN
	IF NOT EXISTS (select * FROM inst_role WHERE (role_sin = NEW.manager_id
												  AND role_name = 'manager'))
		THEN ROLLBACK;
	END IF;		
	RETURN NULL;
END;$$;


ALTER FUNCTION public.employee_is_manager() OWNER TO postgres;

--
-- TOC entry 239 (class 1255 OID 17350)
-- Name: employee_not_customer_or_same_hotel(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.employee_not_customer_or_same_hotel() RETURNS trigger
    LANGUAGE plpgsql
    AS $$BEGIN
	IF EXISTS (SELECT * FROM booking WHERE (booking_id = NEW.booking_id 
											AND customer_sin = NEW.employee_sin)) 
		THEN ROLLBACK;
	ELSE
		IF NOT EXISTS (SELECT * FROM booking AS b, employee AS e
				   WHERE (b.booking_id = NEW.booking_id
						  AND b.hotel_id = e.hotel_id))
			THEN ROLLBACK;
		END IF;
	END IF;
	RETURN NULL;
END;$$;


ALTER FUNCTION public.employee_not_customer_or_same_hotel() OWNER TO postgres;

--
-- TOC entry 221 (class 1259 OID 16925)
-- Name: chain_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.chain_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.chain_id_seq OWNER TO postgres;

--
-- TOC entry 235 (class 1259 OID 17127)
-- Name: hotel_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.hotel_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.hotel_id_seq OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 219 (class 1259 OID 16918)
-- Name: hotel; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.hotel (
    hotel_id integer DEFAULT nextval('public.hotel_id_seq'::regclass) NOT NULL,
    chain_id integer NOT NULL,
    street_num integer NOT NULL,
    street_name character varying(100) NOT NULL,
    unit_num integer,
    city character varying(100) NOT NULL,
    state character varying(2) NOT NULL,
    country character varying(30) NOT NULL,
    postal_code character varying(10) NOT NULL,
    rating integer NOT NULL,
    manager_id numeric(9,0) NOT NULL,
    email character varying(100) NOT NULL,
    CONSTRAINT hotel_rating_check CHECK (((rating >= 1) AND (rating <= 5)))
);


ALTER TABLE public.hotel OWNER TO postgres;

--
-- TOC entry 220 (class 1259 OID 16922)
-- Name: hotel_chain; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.hotel_chain (
    chain_id integer DEFAULT nextval('public.chain_id_seq'::regclass) NOT NULL,
    name character varying(50) NOT NULL,
    street_num integer NOT NULL,
    street_name character varying(100) NOT NULL,
    unit_num integer,
    city character varying(20) NOT NULL,
    state character(2),
    country character varying(30),
    postal_code character varying(10)
);


ALTER TABLE public.hotel_chain OWNER TO postgres;

--
-- TOC entry 233 (class 1259 OID 16959)
-- Name: room; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.room (
    room_number integer NOT NULL,
    price money,
    capacity integer,
    max_capacity integer,
    hotel_id integer NOT NULL,
    CONSTRAINT room_check CHECK (((capacity >= 1) AND (max_capacity >= capacity))),
    CONSTRAINT room_number_check CHECK ((room_number >= 0))
);


ALTER TABLE public.room OWNER TO postgres;

--
-- TOC entry 237 (class 1259 OID 17371)
-- Name: available_rooms; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.available_rooms AS
 SELECT room.room_number,
    room.price,
    room.capacity,
    room.max_capacity,
    hotel.hotel_id,
    hotel.city,
    hotel.state,
    hotel.country,
    hotel.rating,
    hotel_chain.name AS chain_name
   FROM ((public.room
     JOIN public.hotel ON ((room.hotel_id = hotel.hotel_id)))
     JOIN public.hotel_chain ON ((hotel.chain_id = hotel_chain.chain_id)));


ALTER TABLE public.available_rooms OWNER TO postgres;

--
-- TOC entry 214 (class 1259 OID 16900)
-- Name: booking; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.booking (
    booking_id bigint NOT NULL,
    checkin_date date NOT NULL,
    checkout_date date NOT NULL,
    hotel_id integer NOT NULL,
    customer_sin numeric(9,0) NOT NULL,
    room_number integer NOT NULL,
    CONSTRAINT valid_dates CHECK ((checkout_date > checkin_date))
);


ALTER TABLE public.booking OWNER TO postgres;

--
-- TOC entry 215 (class 1259 OID 16903)
-- Name: booking_booking_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.booking_booking_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.booking_booking_id_seq OWNER TO postgres;

--
-- TOC entry 3476 (class 0 OID 0)
-- Dependencies: 215
-- Name: booking_booking_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.booking_booking_id_seq OWNED BY public.booking.booking_id;


--
-- TOC entry 217 (class 1259 OID 16910)
-- Name: chain_inst_phone_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.chain_inst_phone_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.chain_inst_phone_id_seq OWNER TO postgres;

--
-- TOC entry 230 (class 1259 OID 16952)
-- Name: customer; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.customer (
    customer_sin numeric(9,0) NOT NULL,
    registration_date date NOT NULL
);


ALTER TABLE public.customer OWNER TO postgres;

--
-- TOC entry 218 (class 1259 OID 16914)
-- Name: employee; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.employee (
    sin numeric(9,0) NOT NULL,
    hotel_id integer NOT NULL,
    salary money NOT NULL
);


ALTER TABLE public.employee OWNER TO postgres;

--
-- TOC entry 236 (class 1259 OID 17366)
-- Name: hotel_capacity; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.hotel_capacity AS
 SELECT room.capacity,
    room.max_capacity,
    hotel.hotel_id
   FROM (public.room
     JOIN public.hotel ON ((room.hotel_id = hotel.hotel_id)));


ALTER TABLE public.hotel_capacity OWNER TO postgres;

--
-- TOC entry 223 (class 1259 OID 16929)
-- Name: inst_amenity; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.inst_amenity (
    amenity_id integer NOT NULL,
    amenity character varying(100) NOT NULL,
    hotel_id integer NOT NULL
);


ALTER TABLE public.inst_amenity OWNER TO postgres;

--
-- TOC entry 216 (class 1259 OID 16907)
-- Name: inst_chain_phone; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.inst_chain_phone (
    chain_phone_id integer DEFAULT nextval('public.chain_inst_phone_id_seq'::regclass) NOT NULL,
    phone_number numeric(10,0) NOT NULL
);


ALTER TABLE public.inst_chain_phone OWNER TO postgres;

--
-- TOC entry 224 (class 1259 OID 16932)
-- Name: inst_concern; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.inst_concern (
    concern_id integer NOT NULL,
    concern character varying(200) NOT NULL,
    hotel_id integer NOT NULL
);


ALTER TABLE public.inst_concern OWNER TO postgres;

--
-- TOC entry 226 (class 1259 OID 16938)
-- Name: inst_email_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.inst_email_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.inst_email_id_seq OWNER TO postgres;

--
-- TOC entry 225 (class 1259 OID 16935)
-- Name: inst_email; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.inst_email (
    email_id integer DEFAULT nextval('public.inst_email_id_seq'::regclass) NOT NULL,
    email character varying(100) NOT NULL
);


ALTER TABLE public.inst_email OWNER TO postgres;

--
-- TOC entry 222 (class 1259 OID 16926)
-- Name: inst_hotel_phone; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.inst_hotel_phone (
    phone_id integer NOT NULL,
    phone_number numeric(10,0) NOT NULL
);


ALTER TABLE public.inst_hotel_phone OWNER TO postgres;

--
-- TOC entry 227 (class 1259 OID 16939)
-- Name: inst_role; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.inst_role (
    role_sin numeric(9,0) NOT NULL,
    role_name character varying(80) NOT NULL
);


ALTER TABLE public.inst_role OWNER TO postgres;

--
-- TOC entry 228 (class 1259 OID 16942)
-- Name: manages; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.manages (
    booking_id bigint NOT NULL,
    employee_sin numeric(9,0) NOT NULL
);


ALTER TABLE public.manages OWNER TO postgres;

--
-- TOC entry 229 (class 1259 OID 16945)
-- Name: person; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.person (
    sin numeric(9,0) NOT NULL,
    first_name character varying,
    last_name character varying,
    middle_name character varying,
    street_num integer,
    street_name character varying,
    unit_num integer,
    city character varying,
    state character varying(2),
    country character varying(20),
    postal_code character varying(10),
    CONSTRAINT person_street_num_check CHECK ((street_num > 0))
);


ALTER TABLE public.person OWNER TO postgres;

--
-- TOC entry 231 (class 1259 OID 16955)
-- Name: renting; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.renting (
    booking_id bigint NOT NULL,
    cc_number numeric(16,0) NOT NULL,
    expiry_date date NOT NULL
);


ALTER TABLE public.renting OWNER TO postgres;

--
-- TOC entry 232 (class 1259 OID 16958)
-- Name: renting_booking_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.renting_booking_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.renting_booking_id_seq OWNER TO postgres;

--
-- TOC entry 3477 (class 0 OID 0)
-- Dependencies: 232
-- Name: renting_booking_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.renting_booking_id_seq OWNED BY public.renting.booking_id;


--
-- TOC entry 234 (class 1259 OID 16964)
-- Name: supervises; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.supervises (
    subordinate_sin numeric(9,0) NOT NULL,
    supervisor_sin numeric(9,0) NOT NULL
);


ALTER TABLE public.supervises OWNER TO postgres;

--
-- TOC entry 3248 (class 2604 OID 16967)
-- Name: booking booking_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.booking ALTER COLUMN booking_id SET DEFAULT nextval('public.booking_booking_id_seq'::regclass);


--
-- TOC entry 3253 (class 2604 OID 16971)
-- Name: renting booking_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.renting ALTER COLUMN booking_id SET DEFAULT nextval('public.renting_booking_id_seq'::regclass);


--
-- TOC entry 3263 (class 2606 OID 16973)
-- Name: booking booking_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.booking
    ADD CONSTRAINT booking_pkey PRIMARY KEY (booking_id);


--
-- TOC entry 3266 (class 2606 OID 16977)
-- Name: inst_chain_phone chain_inst_phone_number_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inst_chain_phone
    ADD CONSTRAINT chain_inst_phone_number_key UNIQUE (phone_number);


--
-- TOC entry 3268 (class 2606 OID 16979)
-- Name: inst_chain_phone chain_inst_phone_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inst_chain_phone
    ADD CONSTRAINT chain_inst_phone_pkey PRIMARY KEY (chain_phone_id, phone_number);


--
-- TOC entry 3298 (class 2606 OID 17356)
-- Name: customer customer_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.customer
    ADD CONSTRAINT customer_pkey PRIMARY KEY (customer_sin);


--
-- TOC entry 3270 (class 2606 OID 16983)
-- Name: employee employee_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employee
    ADD CONSTRAINT employee_pkey PRIMARY KEY (sin);


--
-- TOC entry 3280 (class 2606 OID 16985)
-- Name: hotel_chain hotel_chain_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.hotel_chain
    ADD CONSTRAINT hotel_chain_pkey PRIMARY KEY (chain_id);


--
-- TOC entry 3282 (class 2606 OID 16987)
-- Name: inst_hotel_phone hotel_inst_phone_number_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inst_hotel_phone
    ADD CONSTRAINT hotel_inst_phone_number_key UNIQUE (phone_number);


--
-- TOC entry 3275 (class 2606 OID 17121)
-- Name: hotel hotel_manager_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.hotel
    ADD CONSTRAINT hotel_manager_unique UNIQUE (manager_id);


--
-- TOC entry 3277 (class 2606 OID 17201)
-- Name: hotel hotel_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.hotel
    ADD CONSTRAINT hotel_pkey PRIMARY KEY (hotel_id);


--
-- TOC entry 3284 (class 2606 OID 16991)
-- Name: inst_amenity inst_amenity_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inst_amenity
    ADD CONSTRAINT inst_amenity_pkey PRIMARY KEY (amenity_id, amenity);


--
-- TOC entry 3286 (class 2606 OID 17132)
-- Name: inst_concern inst_concern_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inst_concern
    ADD CONSTRAINT inst_concern_pkey PRIMARY KEY (concern_id, concern);


--
-- TOC entry 3288 (class 2606 OID 16995)
-- Name: inst_email inst_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inst_email
    ADD CONSTRAINT inst_email_key UNIQUE (email);


--
-- TOC entry 3290 (class 2606 OID 16997)
-- Name: inst_email inst_email_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inst_email
    ADD CONSTRAINT inst_email_pkey PRIMARY KEY (email_id, email);


--
-- TOC entry 3292 (class 2606 OID 17130)
-- Name: inst_role inst_role_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inst_role
    ADD CONSTRAINT inst_role_pkey PRIMARY KEY (role_sin, role_name);


--
-- TOC entry 3294 (class 2606 OID 17307)
-- Name: manages manages_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.manages
    ADD CONSTRAINT manages_pkey PRIMARY KEY (booking_id);


--
-- TOC entry 3296 (class 2606 OID 17003)
-- Name: person person_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.person
    ADD CONSTRAINT person_pkey PRIMARY KEY (sin);


--
-- TOC entry 3300 (class 2606 OID 17309)
-- Name: renting renting_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.renting
    ADD CONSTRAINT renting_pkey PRIMARY KEY (booking_id, cc_number, expiry_date);


--
-- TOC entry 3303 (class 2606 OID 17242)
-- Name: room room_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.room
    ADD CONSTRAINT room_pkey PRIMARY KEY (room_number, hotel_id);


--
-- TOC entry 3259 (class 1259 OID 17346)
-- Name: booking_checkin_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX booking_checkin_index ON public.booking USING btree (checkin_date);


--
-- TOC entry 3260 (class 1259 OID 17347)
-- Name: booking_checkout_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX booking_checkout_index ON public.booking USING btree (checkout_date);


--
-- TOC entry 3261 (class 1259 OID 17345)
-- Name: booking_customer_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX booking_customer_index ON public.booking USING btree (customer_sin);


--
-- TOC entry 3264 (class 1259 OID 17348)
-- Name: booking_room_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX booking_room_index ON public.booking USING btree (room_number);


--
-- TOC entry 3271 (class 1259 OID 17294)
-- Name: hotel_chain_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX hotel_chain_index ON public.hotel USING btree (chain_id);


--
-- TOC entry 3272 (class 1259 OID 17295)
-- Name: hotel_city_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX hotel_city_index ON public.hotel USING btree (city);


--
-- TOC entry 3273 (class 1259 OID 17296)
-- Name: hotel_country_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX hotel_country_index ON public.hotel USING btree (country);


--
-- TOC entry 3278 (class 1259 OID 17297)
-- Name: hotel_rating_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX hotel_rating_index ON public.hotel USING btree (rating);


--
-- TOC entry 3301 (class 1259 OID 17293)
-- Name: room_capacity_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX room_capacity_index ON public.room USING btree (max_capacity);


--
-- TOC entry 3304 (class 1259 OID 17298)
-- Name: room_price_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX room_price_index ON public.room USING btree (price);


--
-- TOC entry 3324 (class 2620 OID 17134)
-- Name: hotel hotel_manager_constraint_trigger; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE CONSTRAINT TRIGGER hotel_manager_constraint_trigger AFTER INSERT OR UPDATE ON public.hotel DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION public.employee_is_manager();


--
-- TOC entry 3325 (class 2620 OID 17353)
-- Name: manages managing_employee_constraint_trigger; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE CONSTRAINT TRIGGER managing_employee_constraint_trigger AFTER INSERT OR UPDATE ON public.manages NOT DEFERRABLE INITIALLY IMMEDIATE FOR EACH ROW EXECUTE FUNCTION public.employee_not_customer_or_same_hotel();


--
-- TOC entry 3305 (class 2606 OID 17357)
-- Name: booking booking_customer_sin_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.booking
    ADD CONSTRAINT booking_customer_sin_fkey FOREIGN KEY (customer_sin) REFERENCES public.customer(customer_sin);


--
-- TOC entry 3306 (class 2606 OID 17325)
-- Name: booking booking_hotel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.booking
    ADD CONSTRAINT booking_hotel_id_fkey FOREIGN KEY (hotel_id) REFERENCES public.hotel(hotel_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3307 (class 2606 OID 17340)
-- Name: booking booking_hotel_id_room_number_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.booking
    ADD CONSTRAINT booking_hotel_id_room_number_fkey FOREIGN KEY (hotel_id, room_number) REFERENCES public.room(hotel_id, room_number) ON UPDATE CASCADE;


--
-- TOC entry 3308 (class 2606 OID 17030)
-- Name: inst_chain_phone chain_inst_phone_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inst_chain_phone
    ADD CONSTRAINT chain_inst_phone_id_fkey FOREIGN KEY (chain_phone_id) REFERENCES public.hotel_chain(chain_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3319 (class 2606 OID 17310)
-- Name: customer customer_sin_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.customer
    ADD CONSTRAINT customer_sin_fkey FOREIGN KEY (customer_sin) REFERENCES public.person(sin) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3309 (class 2606 OID 17258)
-- Name: employee employee_hotel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employee
    ADD CONSTRAINT employee_hotel_id_fkey FOREIGN KEY (hotel_id) REFERENCES public.hotel(hotel_id) DEFERRABLE INITIALLY DEFERRED;


--
-- TOC entry 3310 (class 2606 OID 17045)
-- Name: employee employee_sin_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employee
    ADD CONSTRAINT employee_sin_fkey FOREIGN KEY (sin) REFERENCES public.person(sin) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3311 (class 2606 OID 17050)
-- Name: hotel hotel_chain_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.hotel
    ADD CONSTRAINT hotel_chain_id_fkey FOREIGN KEY (chain_id) REFERENCES public.hotel_chain(chain_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3313 (class 2606 OID 17273)
-- Name: inst_hotel_phone hotel_inst_phone_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inst_hotel_phone
    ADD CONSTRAINT hotel_inst_phone_id_fkey FOREIGN KEY (phone_id) REFERENCES public.hotel(hotel_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3314 (class 2606 OID 17263)
-- Name: inst_amenity inst_amenity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inst_amenity
    ADD CONSTRAINT inst_amenity_id_fkey FOREIGN KEY (amenity_id, hotel_id) REFERENCES public.room(room_number, hotel_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3315 (class 2606 OID 17268)
-- Name: inst_concern inst_concern_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inst_concern
    ADD CONSTRAINT inst_concern_id_fkey FOREIGN KEY (concern_id, hotel_id) REFERENCES public.room(room_number, hotel_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3316 (class 2606 OID 17070)
-- Name: inst_email inst_email_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inst_email
    ADD CONSTRAINT inst_email_id_fkey FOREIGN KEY (email_id) REFERENCES public.hotel_chain(chain_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3317 (class 2606 OID 17075)
-- Name: inst_role inst_role_sin_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.inst_role
    ADD CONSTRAINT inst_role_sin_fkey FOREIGN KEY (role_sin) REFERENCES public.employee(sin) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3312 (class 2606 OID 17080)
-- Name: hotel manager_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.hotel
    ADD CONSTRAINT manager_id_fkey FOREIGN KEY (manager_id) REFERENCES public.employee(sin) ON UPDATE CASCADE ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED;


--
-- TOC entry 3318 (class 2606 OID 17090)
-- Name: manages manages_employee_sin_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.manages
    ADD CONSTRAINT manages_employee_sin_fkey FOREIGN KEY (employee_sin) REFERENCES public.employee(sin) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- TOC entry 3320 (class 2606 OID 17105)
-- Name: renting renting_booking_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.renting
    ADD CONSTRAINT renting_booking_id_fkey FOREIGN KEY (booking_id) REFERENCES public.booking(booking_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3321 (class 2606 OID 17288)
-- Name: room room_hotel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.room
    ADD CONSTRAINT room_hotel_id_fkey FOREIGN KEY (hotel_id) REFERENCES public.hotel(hotel_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3322 (class 2606 OID 17110)
-- Name: supervises supervises_subordinate_sin_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.supervises
    ADD CONSTRAINT supervises_subordinate_sin_fkey FOREIGN KEY (subordinate_sin) REFERENCES public.employee(sin) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3323 (class 2606 OID 17115)
-- Name: supervises supervises_supervisor_sin_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.supervises
    ADD CONSTRAINT supervises_supervisor_sin_fkey FOREIGN KEY (supervisor_sin) REFERENCES public.employee(sin) ON UPDATE CASCADE ON DELETE CASCADE;


-- Completed on 2023-03-31 21:08:49

--
-- PostgreSQL database dump complete
--

