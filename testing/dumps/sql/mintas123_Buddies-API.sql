-- Downloaded from: https://github.com/mintas123/Buddies-API/blob/9242d73958c75da6f9d4275a13b5f9e5597aeb9b/src/dbinit/old_init.sql
--TODO rethink database schema
--
-- PostgreSQL database dump
--

-- Dumped from database version 12.3
-- Dumped by pg_dump version 12.4

-- Started on 2020-11-16 13:28:14

SET
statement_timeout = 0;
SET
lock_timeout = 0;
SET
idle_in_transaction_session_timeout = 0;
SET
client_encoding = 'UTF8';
SET
standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET
check_function_bodies = false;
SET
xmloption = content;
SET
client_min_messages = warning;
SET
row_security = off;

CREATE FUNCTION public.distance(lat double precision, lng double precision, db_lat double precision,
                                db_lng double precision) RETURNS double precision
    LANGUAGE plpgsql
AS
$$
begin
return
        6371 * acos(
                    cos(radians(lat)) * cos(radians(db_lat))
                    *
                    cos(radians(db_lng) - radians(lng))
                +
                    sin(radians(lat))
                        *
                    sin(radians(db_lat))
        );


end;
$$;


ALTER FUNCTION public.distance(lat double precision, lng double precision, db_lat double precision, db_lng double precision) OWNER TO postgres;

SET
default_tablespace = '';

SET
default_table_access_method = heap;

--
-- TOC entry 202 (class 1259 OID 16402)
-- Name: account; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.account
(
    account_id      uuid             NOT NULL,
    email           character varying(63),
    phone           character varying(31),
    hashed_password character varying(255),
    name            character varying(63),
    last_name       character varying(63),
    birthday        date,
    gender          character varying(15),
    description     character varying(2000),
    age_pref        character varying(15),
    count_pref      character varying(15),
    gender_pref     character varying(15),
    location_str    character varying(255),
    location_lat    double precision NOT NULL,
    location_lng    double precision NOT NULL,
    photo_url       character varying(255)
);


ALTER TABLE public.account
    OWNER TO postgres;

--
-- TOC entry 203 (class 1259 OID 16411)
-- Name: account_nay; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.account_nay
(
    account_id uuid NOT NULL,
    nay_tags   character varying(255)
);


ALTER TABLE public.account_nay
    OWNER TO postgres;

--
-- TOC entry 204 (class 1259 OID 16414)
-- Name: account_tags; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.account_tags
(
    account_id uuid NOT NULL,
    about_tags character varying(255)
);


ALTER TABLE public.account_tags
    OWNER TO postgres;

--
-- TOC entry 205 (class 1259 OID 16417)
-- Name: account_yay; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.account_yay
(
    account_id uuid NOT NULL,
    yay_tags   character varying(255)
);


ALTER TABLE public.account_yay
    OWNER TO postgres;

--
-- TOC entry 213 (class 1259 OID 24599)
-- Name: chat_message; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.chat_message
(
    id             uuid NOT NULL,
    chat_id        character varying(255),
    content        character varying(500),
    recipient_id   uuid,
    recipient_name character varying(255),
    sender_id      uuid,
    sender_name    character varying(255),
    status         character varying(255),
    "timestamp"    timestamp without time zone
);


ALTER TABLE public.chat_message
    OWNER TO postgres;

--
-- TOC entry 212 (class 1259 OID 24594)
-- Name: chat_room; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.chat_room
(
    id           uuid NOT NULL,
    chat_id      character varying(255),
    recipient_id uuid,
    sender_id    uuid
);


ALTER TABLE public.chat_room
    OWNER TO postgres;

--
-- TOC entry 206 (class 1259 OID 16423)
-- Name: fav_rental; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.fav_rental
(
    account_id uuid NOT NULL,
    rental_id  uuid NOT NULL
);


ALTER TABLE public.fav_rental
    OWNER TO postgres;

--
-- TOC entry 207 (class 1259 OID 16426)
-- Name: feature_tags; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.feature_tags
(
    rental_id    uuid NOT NULL,
    feature_tags character varying(255)
);


ALTER TABLE public.feature_tags
    OWNER TO postgres;

--
-- TOC entry 208 (class 1259 OID 16429)
-- Name: friends_table; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.friends_table
(
    account_id uuid NOT NULL,
    friend_id  uuid NOT NULL
);


ALTER TABLE public.friends_table
    OWNER TO postgres;

--
-- TOC entry 209 (class 1259 OID 16432)
-- Name: hibernate_sequence; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.hibernate_sequence
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE CACHE 1;


ALTER TABLE public.hibernate_sequence
    OWNER TO postgres;

--
-- TOC entry 210 (class 1259 OID 16437)
-- Name: rental; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.rental
(
    rental_id     uuid             NOT NULL,
    creator_id    uuid,
    title         character varying(100),
    is_negotiable boolean          NOT NULL,
    description   character varying(3000),
    build_year    integer          NOT NULL,
    price         integer          NOT NULL,
    deposit       integer          NOT NULL,
    price_m_sq    double precision,
    rent_date     date,
    size          double precision NOT NULL,
    rooms         integer          NOT NULL,
    floor         integer          NOT NULL,
    location_lat  double precision NOT NULL,
    location_lng  double precision NOT NULL,
    location_str  character varying(255)
);


ALTER TABLE public.rental
    OWNER TO postgres;

--
-- TOC entry 211 (class 1259 OID 16443)
-- Name: rental_pic_urls; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.rental_pic_urls
(
    rental_id  uuid NOT NULL,
    photo_urls character varying(255)
);


ALTER TABLE public.rental_pic_urls
    OWNER TO postgres;

--
-- TOC entry 3727 (class 2606 OID 16449)
-- Name: account account_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.account
    ADD CONSTRAINT account_pkey PRIMARY KEY (account_id);


--
-- TOC entry 3737 (class 2606 OID 24606)
-- Name: chat_message chat_message_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chat_message
    ADD CONSTRAINT chat_message_pkey PRIMARY KEY (id);


--
-- TOC entry 3735 (class 2606 OID 24598)
-- Name: chat_room chat_room_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chat_room
    ADD CONSTRAINT chat_room_pkey PRIMARY KEY (id);


--
-- TOC entry 3729 (class 2606 OID 16453)
-- Name: fav_rental fav_rental_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.fav_rental
    ADD CONSTRAINT fav_rental_pkey PRIMARY KEY (account_id, rental_id);


--
-- TOC entry 3731 (class 2606 OID 16455)
-- Name: friends_table friends_table_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.friends_table
    ADD CONSTRAINT friends_table_pkey PRIMARY KEY (friend_id, account_id);


--
-- TOC entry 3733 (class 2606 OID 16459)
-- Name: rental rental_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.rental
    ADD CONSTRAINT rental_pkey PRIMARY KEY (rental_id);


--
-- TOC entry 3738 (class 2606 OID 16465)
-- Name: account_nay fk395putnww4q5y1ydsuup5kuf1; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.account_nay
    ADD CONSTRAINT fk395putnww4q5y1ydsuup5kuf1 FOREIGN KEY (account_id) REFERENCES public.account (account_id);


--
-- TOC entry 3743 (class 2606 OID 16470)
-- Name: feature_tags fk459dt6ppn8jq0m0rbsskup6kb; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.feature_tags
    ADD CONSTRAINT fk459dt6ppn8jq0m0rbsskup6kb FOREIGN KEY (rental_id) REFERENCES public.rental (rental_id);


--
-- TOC entry 3744 (class 2606 OID 16480)
-- Name: friends_table fk62af6kydamwnr4nmjggm043vq; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.friends_table
    ADD CONSTRAINT fk62af6kydamwnr4nmjggm043vq FOREIGN KEY (friend_id) REFERENCES public.account (account_id);


--
-- TOC entry 3741 (class 2606 OID 16485)
-- Name: fav_rental fk6tsxjh4c1c1q3atjew3gu5894; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.fav_rental
    ADD CONSTRAINT fk6tsxjh4c1c1q3atjew3gu5894 FOREIGN KEY (account_id) REFERENCES public.account (account_id);


--
-- TOC entry 3747 (class 2606 OID 16495)
-- Name: rental_pic_urls fka3s615dh158g0v69ly4ly3a5b; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.rental_pic_urls
    ADD CONSTRAINT fka3s615dh158g0v69ly4ly3a5b FOREIGN KEY (rental_id) REFERENCES public.rental (rental_id);


--
-- TOC entry 3745 (class 2606 OID 16505)
-- Name: friends_table fkb1gokokap499c6hqmmrnmi6rh; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.friends_table
    ADD CONSTRAINT fkb1gokokap499c6hqmmrnmi6rh FOREIGN KEY (account_id) REFERENCES public.account (account_id);


--
-- TOC entry 3746 (class 2606 OID 16510)
-- Name: rental fkb2oywyfdhfppwhi3os4pqxgd5; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.rental
    ADD CONSTRAINT fkb2oywyfdhfppwhi3os4pqxgd5 FOREIGN KEY (creator_id) REFERENCES public.account (account_id);


--
-- TOC entry 3740 (class 2606 OID 16515)
-- Name: account_yay fkets0f0cq42gqwu9k2m863qvob; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.account_yay
    ADD CONSTRAINT fkets0f0cq42gqwu9k2m863qvob FOREIGN KEY (account_id) REFERENCES public.account (account_id);


--
-- TOC entry 3742 (class 2606 OID 16520)
-- Name: fav_rental fkl50993o2ia1m2cnvylqdwgld4; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.fav_rental
    ADD CONSTRAINT fkl50993o2ia1m2cnvylqdwgld4 FOREIGN KEY (rental_id) REFERENCES public.rental (rental_id);


--
-- TOC entry 3739 (class 2606 OID 16525)
-- Name: account_tags fkp3bn2fkyg67xosnh8srrff7vn; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.account_tags
    ADD CONSTRAINT fkp3bn2fkyg67xosnh8srrff7vn FOREIGN KEY (account_id) REFERENCES public.account (account_id);


--
-- TOC entry 3879 (class 0 OID 0)
-- Dependencies: 3
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT
ALL
ON SCHEMA public TO postgres;
GRANT ALL
ON SCHEMA public TO PUBLIC;


-- Completed on 2020-11-16 13:28:18

--
-- PostgreSQL database dump complete
--

