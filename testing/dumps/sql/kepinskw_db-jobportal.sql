-- Downloaded from: https://github.com/kepinskw/db-jobportal/blob/6999ce3ca894c62abcd2cdf5d5a94618d26a7441/schema.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 14.13 (Ubuntu 14.13-1.pgdg22.04+1)
-- Dumped by pg_dump version 17.0 (Ubuntu 17.0-1.pgdg22.04+1)

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
-- Name: pg_cron; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_cron WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION pg_cron; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pg_cron IS 'Job scheduler for PostgreSQL';


--
-- Name: public; Type: SCHEMA; Schema: -; Owner: postgres
--

-- *not* creating schema, since initdb creates it


ALTER SCHEMA public OWNER TO postgres;

--
-- Name: add_new_application(integer, integer, character varying); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.add_new_application(IN p_job_id integer, IN p_user_id integer, IN p_status character varying)
    LANGUAGE plpgsql
    AS $$
DECLARE
    new_app_id INT;
BEGIN
    SELECT COALESCE(MAX(app_id), 0) + 1 INTO new_app_id FROM applications;

    INSERT INTO applications (app_id, job_id, user_id, status)
    VALUES (new_app_id, p_job_id, p_user_id, p_status);
    
    RAISE NOTICE 'New application ID: % was added', new_app_id;
END;
$$;


ALTER PROCEDURE public.add_new_application(IN p_job_id integer, IN p_user_id integer, IN p_status character varying) OWNER TO postgres;

--
-- Name: add_new_employer(character varying, character varying); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.add_new_employer(IN p_employername character varying, IN p_password character varying)
    LANGUAGE plpgsql
    AS $$
DECLARE
    new_employer_id INT;
BEGIN
    SELECT COALESCE(MAX(employer_id), 0) + 1 INTO new_employer_id FROM employer;

    INSERT INTO employer (employer_id, employername, password)
    VALUES (new_employer_id, p_employername, p_password);
    
    RAISE NOTICE 'New employer: % ID: % was added', p_employername, new_employer_id;
END;
$$;


ALTER PROCEDURE public.add_new_employer(IN p_employername character varying, IN p_password character varying) OWNER TO postgres;

--
-- Name: add_new_interview(integer, integer, timestamp without time zone, character varying); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.add_new_interview(IN p_user_id integer, IN p_employer_id integer, IN p_timestamp timestamp without time zone, IN p_text character varying)
    LANGUAGE plpgsql
    AS $$
DECLARE
    new_interview_id INT;
BEGIN
    SELECT COALESCE(MAX(interview_id), 0) + 1 INTO new_interview_id FROM interviewscheduler;

    INSERT INTO interviewscheduler (interview_id, user_id, employer_id, timestamp, text)
    VALUES (new_interview_id, p_user_id, p_employer_id, p_timestamp, p_text);
    
    RAISE NOTICE 'New interview ID: % was added', new_interview_id;
END;
$$;


ALTER PROCEDURE public.add_new_interview(IN p_user_id integer, IN p_employer_id integer, IN p_timestamp timestamp without time zone, IN p_text character varying) OWNER TO postgres;

--
-- Name: add_new_job_offer(integer, character varying, character varying, character varying, real, real); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.add_new_job_offer(IN p_employer_id integer, IN p_title character varying, IN p_description character varying, IN p_location character varying, IN p_salaryfrom real, IN p_salaryupto real)
    LANGUAGE plpgsql
    AS $$
DECLARE
    new_job_id INT;
BEGIN
    SELECT COALESCE(MAX(job_id), 0) + 1 INTO new_job_id FROM joboffers;

    INSERT INTO joboffers (job_id, employer_id, title, description, location, salary_from, salary_upto)
    VALUES (new_job_id, p_employer_id, p_title, p_description, p_location, p_salaryfrom, p_salaryupto);
    
    RAISE NOTICE 'New joboffer ID: % was added', new_job_id;
END;
$$;


ALTER PROCEDURE public.add_new_job_offer(IN p_employer_id integer, IN p_title character varying, IN p_description character varying, IN p_location character varying, IN p_salaryfrom real, IN p_salaryupto real) OWNER TO postgres;

--
-- Name: add_new_message(integer, integer, character varying); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.add_new_message(IN p_sender_id integer, IN p_receiver_id integer, IN p_text character varying)
    LANGUAGE plpgsql
    AS $$
DECLARE
    new_message_id INT;
BEGIN
    SELECT COALESCE(MAX(message_id), 0) + 1 INTO new_message_id FROM messages;

    INSERT INTO messages (message_id, sender_id, receiver_id, text, timestamp)
    VALUES (new_message_id, p_sender_id, p_receiver_id, p_text, CURRENT_TIMESTAMP);
    
    RAISE NOTICE 'New message ID: % was added', new_message_id;
END;
$$;


ALTER PROCEDURE public.add_new_message(IN p_sender_id integer, IN p_receiver_id integer, IN p_text character varying) OWNER TO postgres;

--
-- Name: add_new_portfolio(integer, character varying, character varying); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.add_new_portfolio(IN p_user_id integer, IN p_text character varying, IN p_website character varying)
    LANGUAGE plpgsql
    AS $$
DECLARE
    new_portfolio_id INT;
BEGIN
    SELECT COALESCE(MAX(portfolio_id), 0) + 1 INTO new_portfolio_id FROM portfolio;

    INSERT INTO portfolio (portfolio_id, user_id, text, website)
    VALUES (new_portfolio_id, p_user_id, p_text, p_website);
    
    RAISE NOTICE 'New portfolio ID: % was added', new_portfolio_id;
END;
$$;


ALTER PROCEDURE public.add_new_portfolio(IN p_user_id integer, IN p_text character varying, IN p_website character varying) OWNER TO postgres;

--
-- Name: add_new_profile(integer, character varying, character varying); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.add_new_profile(IN p_user_id integer, IN p_headline character varying, IN p_summary character varying)
    LANGUAGE plpgsql
    AS $$
DECLARE
    new_profile_id INT;
BEGIN
    SELECT COALESCE(MAX(profile_id), 0) + 1 INTO new_profile_id FROM profile;

    INSERT INTO profile (profile_id, user_id, headline, summary)
    VALUES (new_profile_id, p_user_id, p_headline, p_summary);
    
    RAISE NOTICE 'New profile ID: % was added', new_profile_id;
END;
$$;


ALTER PROCEDURE public.add_new_profile(IN p_user_id integer, IN p_headline character varying, IN p_summary character varying) OWNER TO postgres;

--
-- Name: add_new_user(character varying, character varying); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.add_new_user(IN p_username character varying, IN p_password character varying)
    LANGUAGE plpgsql
    AS $$
DECLARE
    new_user_id INT;
BEGIN
    SELECT COALESCE(MAX(user_id), 0) + 1 INTO new_user_id FROM users;

    INSERT INTO users (user_id, username, password)
    VALUES (new_user_id, p_username, p_password);
    
    RAISE NOTICE 'New user: % ID: % was added', p_username, new_user_id;
END;
$$;


ALTER PROCEDURE public.add_new_user(IN p_username character varying, IN p_password character varying) OWNER TO postgres;

--
-- Name: apply_for_job(integer, integer, character varying); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.apply_for_job(IN p_user_id integer, IN p_job_id integer, IN p_message character varying)
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_employer_id INT;
    v_job_title VARCHAR(100);
    v_app_id INT;
    v_message_id INT;
BEGIN
    -- Sprawdzenie, czy użytkownik istnieje
    IF NOT EXISTS (SELECT 1 FROM users WHERE user_id = p_user_id) THEN
        RAISE EXCEPTION 'User with ID % does not exist', p_user_id;
    END IF;
    -- Sprawdzenie, czy oferta pracy istnieje i pobranie informacji o niej
    SELECT employer_id, title INTO v_employer_id, v_job_title
    FROM joboffers
    WHERE job_id = p_job_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Job offer with ID % does not exist', p_job_id;
    END IF;
    -- Sprawdzenie, czy użytkownik już nie aplikował na tę ofertę
    IF EXISTS (SELECT 1 FROM applications WHERE user_id = p_user_id AND job_id = p_job_id) THEN
        RAISE EXCEPTION 'User has already applied for this job';
    END IF;
    -- Dodanie nowej aplikacji
	SELECT COALESCE(MAX(app_id),0) + 1 INTO v_app_id FROM applications;
    INSERT INTO applications (app_id, job_id, user_id, status)
    VALUES (v_app_id, p_job_id, p_user_id, 'pending');
    -- Wysłanie wiadomości
	SELECT COALESCE(MAX(message_id),0) + 1 INTO v_message_id FROM messages;
    INSERT INTO messages (message_id, sender_id, receiver_id, text, timestamp)
    VALUES (v_message_id, p_user_id, v_employer_id, p_message, CURRENT_TIMESTAMP);
    -- Jeśli wszystko się powiodło, zatwierdzamy transakcję
    RAISE NOTICE 'Application submitted successfully. Application ID: %, Message ID: %', v_app_id, v_message_id;
EXCEPTION
    WHEN OTHERS THEN
        -- W przypadku błędu, cofamy transakcję i zgłaszamy wyjątek
        ROLLBACK;
        RAISE EXCEPTION 'Error applying for job: %', SQLERRM;
	COMMIT;
END;
$$;


ALTER PROCEDURE public.apply_for_job(IN p_user_id integer, IN p_job_id integer, IN p_message character varying) OWNER TO postgres;

--
-- Name: avg_salary_by_location(character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.avg_salary_by_location(p_location character varying) RETURNS real
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_average_salary float4;
BEGIN
    SELECT 
        ROUND(AVG((salary_from + salary_upto) / 2)::numeric, 2)
    INTO v_average_salary
    FROM joboffers
    WHERE location = p_location
    AND salary_from IS NOT NULL 
    AND salary_upto IS NOT NULL;
    
    -- Jeśli nie znaleziono ofert, zwracamy 0
    RETURN COALESCE(v_average_salary, 0);
END;
$$;


ALTER FUNCTION public.avg_salary_by_location(p_location character varying) OWNER TO postgres;

--
-- Name: decline_other_applications(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.decline_other_applications() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    other_app RECORD;
    employer_id_val INT4;
BEGIN
    IF NEW.status = 'hired' THEN
        SELECT employer_id INTO employer_id_val 
        FROM joboffers 
        WHERE job_id = NEW.job_id;
        
        FOR other_app IN 
            SELECT app_id, user_id 
            FROM applications 
            WHERE job_id = NEW.job_id 
            AND app_id != NEW.app_id 
            AND status NOT IN ('declined', 'hired','rejected')
        LOOP
            UPDATE applications 
            SET status = 'declined'
            WHERE app_id = other_app.app_id;
            
            CALL public.add_new_message(
                employer_id_val,
                other_app.user_id,
                'Thank you for your interest in our company. We regret to inform you that we have chosen another candidate for this position.'
            );
        END LOOP;
    END IF;
    
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.decline_other_applications() OWNER TO postgres;

--
-- Name: send_daily_interview_reminders(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.send_daily_interview_reminders() RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    interview_record RECORD;
BEGIN
    -- Wszystkie interview dnia
    FOR interview_record IN 
        SELECT 
            i.interview_id,
            i.user_id,
            i.employer_id,
            i.timestamp as interview_time,
            j.title as job_title,
            e.employername
        FROM interviewscheduler i
        JOIN joboffers j ON i.employer_id = j.employer_id
        JOIN employer e ON i.employer_id = e.employer_id
        WHERE DATE(i.timestamp) = CURRENT_DATE
        AND i.timestamp > CURRENT_TIMESTAMP
        -- Sprawdzenie czy nie było już powiadomienia
        AND NOT EXISTS (
            SELECT 1 
            FROM messages m 
            WHERE m.receiver_id = i.user_id 
            AND m.sender_id = i.employer_id
            AND DATE(m.timestamp) = CURRENT_DATE
            AND m.text LIKE 'You have a meeting%'
        )
        ORDER BY i.timestamp ASC
    LOOP
        
        INSERT INTO messages (
            sender_id,    
            receiver_id, 
            text,
            timestamp
        ) VALUES (
            interview_record.employer_id,  
            interview_record.user_id,      
            FORMAT(
                'You have a meeting for %s with %s at %s',
                interview_record.job_title,
                interview_record.employername,
                interview_record.interview_time
            ),
            CURRENT_TIMESTAMP
        );
        
    END LOOP;
END;
$$;


ALTER FUNCTION public.send_daily_interview_reminders() OWNER TO postgres;

--
-- Name: update_application_status(integer, character varying); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.update_application_status(IN p_app_id integer, IN p_status character varying)
    LANGUAGE plpgsql
    AS $$
DECLARE
    valid_statuses VARCHAR(25)[] := ARRAY['pending', 'under_review', 'interviewing', 'hired', 'declined', 'rejected'];
    current_status VARCHAR(25);
BEGIN
    IF NOT EXISTS (SELECT 1 FROM applications WHERE app_id = p_app_id) THEN
        RAISE EXCEPTION 'Application with ID % does not exist', p_app_id;
    END IF;

    SELECT status INTO current_status
    FROM applications
    WHERE app_id = p_app_id;

    IF NOT p_status = ANY(valid_statuses) THEN
        RAISE EXCEPTION 'Invalid status: %. Valid statuses are: %', p_status, array_to_string(valid_statuses, ', ');
    END IF;

    IF current_status IN ('hired', 'declined','rejected') THEN
        RAISE EXCEPTION 'Cannot update status: Application is already % and finalized', current_status;
    END IF;

    UPDATE applications
    SET status = p_status
    WHERE app_id = p_app_id;

    RAISE NOTICE 'Application % status updated from % to %', p_app_id, current_status, p_status;
END;
$$;


ALTER PROCEDURE public.update_application_status(IN p_app_id integer, IN p_status character varying) OWNER TO postgres;

--
-- Name: update_user_status_on_hire(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_user_status_on_hire() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_employer_id INT4;
    v_job_title VARCHAR(100);
	v_status_id INT4;
BEGIN
    
    SELECT j.employer_id, j.title INTO v_employer_id, v_job_title
    FROM joboffers j 
    WHERE j.job_id = NEW.job_id;

    SELECT COALESCE(MAX(status_id), 0) + 1 INTO v_status_id FROM status;

    IF NEW.status = 'hired' THEN
        INSERT INTO status (
			status_id,
            user_id,
            employer_id,
            position,
            date_from
        ) VALUES (
			v_status_id,
            NEW.user_id,
            v_employer_id,
            v_job_title,
            CURRENT_TIMESTAMP
        );
        
        RAISE NOTICE 'New status record_id % created for user_id % with employer_id % and position %', 
            v_status_id,NEW.user_id, v_employer_id, v_job_title;
    END IF;
    
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_user_status_on_hire() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: applications; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.applications (
    app_id integer NOT NULL,
    job_id integer NOT NULL,
    user_id integer NOT NULL,
    status character varying(25) NOT NULL
);


ALTER TABLE public.applications OWNER TO postgres;

--
-- Name: education; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.education (
    education_id integer NOT NULL,
    user_id integer NOT NULL,
    school_name character varying(200) NOT NULL,
    field_of_study character varying(150),
    degree_level character varying(50),
    start_date date NOT NULL,
    end_date date,
    is_current boolean DEFAULT false,
    grade character varying(20),
    location character varying(100)
);


ALTER TABLE public.education OWNER TO postgres;

--
-- Name: employer; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.employer (
    employer_id integer NOT NULL,
    employername character varying(100) NOT NULL,
    password character varying(255) NOT NULL
);


ALTER TABLE public.employer OWNER TO postgres;

--
-- Name: joboffers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.joboffers (
    job_id integer NOT NULL,
    employer_id integer NOT NULL,
    title character varying(100) NOT NULL,
    description character varying(1000) NOT NULL,
    location character varying(100) NOT NULL,
    salary_from real NOT NULL,
    salary_upto real NOT NULL
);


ALTER TABLE public.joboffers OWNER TO postgres;

--
-- Name: employer_applicant_stats_view; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.employer_applicant_stats_view AS
 SELECT j.job_id,
    j.title AS job_title,
    count(a.app_id) AS total_applications,
    sum(
        CASE
            WHEN ((a.status)::text = 'rejected'::text) THEN 1
            ELSE 0
        END) AS rejected_applications,
    sum(
        CASE
            WHEN ((a.status)::text = 'accepted'::text) THEN 1
            ELSE 0
        END) AS accepted_applications,
    sum(
        CASE
            WHEN ((a.status)::text = 'pending'::text) THEN 1
            ELSE 0
        END) AS pending_applications
   FROM (public.joboffers j
     LEFT JOIN public.applications a ON ((j.job_id = a.job_id)))
  GROUP BY j.job_id, j.title;


ALTER VIEW public.employer_applicant_stats_view OWNER TO postgres;

--
-- Name: profile; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.profile (
    profile_id integer NOT NULL,
    user_id integer NOT NULL,
    headline character varying(100) NOT NULL,
    summary character varying(2000) NOT NULL
);


ALTER TABLE public.profile OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    user_id integer NOT NULL,
    username character varying(50) NOT NULL,
    password character varying(255) NOT NULL
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: employer_applications_view; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.employer_applications_view AS
 SELECT a.app_id,
    a.job_id,
    a.status,
    u.username,
    p.headline,
    j.title AS job_title
   FROM (((public.applications a
     JOIN public.users u ON ((a.user_id = u.user_id)))
     JOIN public.profile p ON ((u.user_id = p.user_id)))
     JOIN public.joboffers j ON ((a.job_id = j.job_id)));


ALTER VIEW public.employer_applications_view OWNER TO postgres;

--
-- Name: event_registrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.event_registrations (
    registration_id integer NOT NULL,
    event_id integer,
    user_id integer,
    registration_date timestamp without time zone,
    status character varying(50),
    notes character varying(500),
    attendance_confirmed boolean
);


ALTER TABLE public.event_registrations OWNER TO postgres;

--
-- Name: interviewscheduler; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.interviewscheduler (
    interview_id integer NOT NULL,
    user_id integer NOT NULL,
    employer_id integer NOT NULL,
    "timestamp" timestamp without time zone NOT NULL,
    text character varying(500)
);


ALTER TABLE public.interviewscheduler OWNER TO postgres;

--
-- Name: job_events; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.job_events (
    event_id integer NOT NULL,
    title character varying(200) NOT NULL,
    description text,
    event_date date NOT NULL,
    start_time time without time zone,
    end_time time without time zone,
    location character varying(200),
    event_type character varying(50)
);


ALTER TABLE public.job_events OWNER TO postgres;

--
-- Name: jobseeker_job_offers_view; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.jobseeker_job_offers_view AS
 SELECT j.job_id,
    j.title,
    j.description,
    j.location,
    j.salary_from,
    j.salary_upto,
    count(a.app_id) AS application_count
   FROM (public.joboffers j
     LEFT JOIN public.applications a ON ((j.job_id = a.job_id)))
  GROUP BY j.job_id, j.title, j.description, j.location, j.salary_from, j.salary_upto;


ALTER VIEW public.jobseeker_job_offers_view OWNER TO postgres;

--
-- Name: portfolio; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.portfolio (
    portfolio_id integer NOT NULL,
    user_id integer NOT NULL,
    text character varying(2500) NOT NULL,
    website character varying(255)
);


ALTER TABLE public.portfolio OWNER TO postgres;

--
-- Name: jobseeker_profile_view; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.jobseeker_profile_view AS
 SELECT u.user_id,
    u.username,
    p.headline,
    p.summary,
    po.website
   FROM ((public.users u
     LEFT JOIN public.profile p ON ((u.user_id = p.user_id)))
     LEFT JOIN public.portfolio po ON ((u.user_id = po.user_id)));


ALTER VIEW public.jobseeker_profile_view OWNER TO postgres;

--
-- Name: messages; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.messages (
    message_id integer NOT NULL,
    sender_id integer NOT NULL,
    receiver_id integer NOT NULL,
    text character varying(1000) NOT NULL,
    "timestamp" timestamp without time zone NOT NULL
);


ALTER TABLE public.messages OWNER TO postgres;

--
-- Name: recruiter_job_offers_view; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.recruiter_job_offers_view AS
 SELECT count(a.app_id) AS application_count,
    j.job_id,
    j.title,
    j.description,
    j.location,
    j.salary_from,
    j.salary_upto,
    e.employername AS employer_name
   FROM ((public.joboffers j
     JOIN public.employer e ON ((j.employer_id = e.employer_id)))
     LEFT JOIN public.applications a ON ((j.job_id = a.job_id)))
  GROUP BY j.job_id, j.title, j.description, j.location, j.salary_from, j.salary_upto, e.employername
  ORDER BY (count(a.app_id)) DESC;


ALTER VIEW public.recruiter_job_offers_view OWNER TO postgres;

--
-- Name: recruiter_user_view; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.recruiter_user_view AS
 SELECT u.username,
    p.headline,
    p.summary
   FROM (public.users u
     LEFT JOIN public.profile p ON ((u.user_id = p.user_id)));


ALTER VIEW public.recruiter_user_view OWNER TO postgres;

--
-- Name: status; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.status (
    status_id integer NOT NULL,
    user_id integer NOT NULL,
    employer_id integer NOT NULL,
    "position" character varying(100) NOT NULL,
    date_from date NOT NULL,
    date_to date
);


ALTER TABLE public.status OWNER TO postgres;

--
-- Name: applications app_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.applications
    ADD CONSTRAINT app_id PRIMARY KEY (app_id);


--
-- Name: education education_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.education
    ADD CONSTRAINT education_pkey PRIMARY KEY (education_id);


--
-- Name: employer employer_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employer
    ADD CONSTRAINT employer_id PRIMARY KEY (employer_id);


--
-- Name: employer employername; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employer
    ADD CONSTRAINT employername UNIQUE (employername);


--
-- Name: event_registrations event_registrations_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_registrations
    ADD CONSTRAINT event_registrations_pk PRIMARY KEY (registration_id);


--
-- Name: interviewscheduler interview_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.interviewscheduler
    ADD CONSTRAINT interview_id PRIMARY KEY (interview_id);


--
-- Name: job_events job_events_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.job_events
    ADD CONSTRAINT job_events_pkey PRIMARY KEY (event_id);


--
-- Name: joboffers job_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.joboffers
    ADD CONSTRAINT job_id PRIMARY KEY (job_id);


--
-- Name: messages message_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT message_id PRIMARY KEY (message_id);


--
-- Name: portfolio portfolio_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.portfolio
    ADD CONSTRAINT portfolio_id PRIMARY KEY (portfolio_id);


--
-- Name: profile profile_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.profile
    ADD CONSTRAINT profile_id PRIMARY KEY (profile_id);


--
-- Name: status status_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.status
    ADD CONSTRAINT status_pkey PRIMARY KEY (status_id);


--
-- Name: event_registrations unique_event_user; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_registrations
    ADD CONSTRAINT unique_event_user UNIQUE (event_id, user_id);


--
-- Name: profile user_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.profile
    ADD CONSTRAINT user_id UNIQUE (user_id);


--
-- Name: users username; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT username UNIQUE (username);


--
-- Name: users users_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_id PRIMARY KEY (user_id);


--
-- Name: idx_applications_job_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_applications_job_id ON public.applications USING btree (job_id);


--
-- Name: idx_applications_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_applications_status ON public.applications USING btree (status);


--
-- Name: idx_applications_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_applications_user_id ON public.applications USING btree (user_id);


--
-- Name: idx_education_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_education_user_id ON public.education USING btree (user_id);


--
-- Name: idx_event_registrations_event_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_registrations_event_id ON public.event_registrations USING btree (event_id);


--
-- Name: idx_event_registrations_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_registrations_status ON public.event_registrations USING btree (status);


--
-- Name: idx_event_registrations_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_registrations_user_id ON public.event_registrations USING btree (user_id);


--
-- Name: idx_interview_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_interview_date ON public.interviewscheduler USING btree ("timestamp");


--
-- Name: idx_interviewscheduler_timestamp; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_interviewscheduler_timestamp ON public.interviewscheduler USING btree ("timestamp");


--
-- Name: idx_job_events_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_job_events_date ON public.job_events USING btree (event_date);


--
-- Name: idx_job_events_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_job_events_type ON public.job_events USING btree (event_type);


--
-- Name: idx_joboffers_employer_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_joboffers_employer_id ON public.joboffers USING btree (employer_id);


--
-- Name: idx_joboffers_location_title; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_joboffers_location_title ON public.joboffers USING btree (location, title);


--
-- Name: idx_joboffers_salary; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_joboffers_salary ON public.joboffers USING btree (salary_from, salary_upto);


--
-- Name: idx_messages_conversation; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_messages_conversation ON public.messages USING btree (sender_id, receiver_id);


--
-- Name: idx_messages_sender_receiver; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_messages_sender_receiver ON public.messages USING btree (sender_id, receiver_id, "timestamp");


--
-- Name: idx_messages_time; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_messages_time ON public.messages USING btree ("timestamp");


--
-- Name: idx_status_datefrom_to; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_status_datefrom_to ON public.status USING btree (date_from, date_to);


--
-- Name: idx_status_employer_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_status_employer_id ON public.status USING btree (employer_id);


--
-- Name: idx_status_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_status_user_id ON public.status USING btree (user_id);


--
-- Name: applications decline_others_on_hire; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER decline_others_on_hire AFTER UPDATE OF status ON public.applications FOR EACH ROW WHEN ((((old.status)::text <> 'hired'::text) AND ((new.status)::text = 'hired'::text))) EXECUTE FUNCTION public.decline_other_applications();


--
-- Name: applications user_hired_trigger; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER user_hired_trigger AFTER UPDATE ON public.applications FOR EACH ROW WHEN ((((old.status)::text <> 'hired'::text) AND ((new.status)::text = 'hired'::text))) EXECUTE FUNCTION public.update_user_status_on_hire();


--
-- Name: applications applications_joboffers_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.applications
    ADD CONSTRAINT applications_joboffers_fk FOREIGN KEY (job_id) REFERENCES public.joboffers(job_id);


--
-- Name: applications applications_users_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.applications
    ADD CONSTRAINT applications_users_fk FOREIGN KEY (user_id) REFERENCES public.users(user_id);


--
-- Name: education education_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.education
    ADD CONSTRAINT education_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id);


--
-- Name: event_registrations event_registrations_job_events_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_registrations
    ADD CONSTRAINT event_registrations_job_events_fk FOREIGN KEY (event_id) REFERENCES public.job_events(event_id);


--
-- Name: event_registrations event_registrations_users_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_registrations
    ADD CONSTRAINT event_registrations_users_fk FOREIGN KEY (user_id) REFERENCES public.users(user_id);


--
-- Name: interviewscheduler interviewscheduler_employer_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.interviewscheduler
    ADD CONSTRAINT interviewscheduler_employer_fk FOREIGN KEY (employer_id) REFERENCES public.employer(employer_id);


--
-- Name: interviewscheduler interviewscheduler_users_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.interviewscheduler
    ADD CONSTRAINT interviewscheduler_users_fk FOREIGN KEY (user_id) REFERENCES public.users(user_id);


--
-- Name: joboffers joboffers_employer_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.joboffers
    ADD CONSTRAINT joboffers_employer_fk FOREIGN KEY (employer_id) REFERENCES public.employer(employer_id);


--
-- Name: messages messages_users_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_users_fk FOREIGN KEY (sender_id) REFERENCES public.users(user_id);


--
-- Name: messages messages_users_fk_1; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_users_fk_1 FOREIGN KEY (receiver_id) REFERENCES public.users(user_id);


--
-- Name: portfolio portfolio_users_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.portfolio
    ADD CONSTRAINT portfolio_users_fk FOREIGN KEY (user_id) REFERENCES public.users(user_id);


--
-- Name: profile profile_users_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.profile
    ADD CONSTRAINT profile_users_fk FOREIGN KEY (user_id) REFERENCES public.users(user_id);


--
-- Name: status status_employer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.status
    ADD CONSTRAINT status_employer_id_fkey FOREIGN KEY (employer_id) REFERENCES public.employer(employer_id);


--
-- Name: status status_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.status
    ADD CONSTRAINT status_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id);


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE USAGE ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- Name: PROCEDURE add_new_application(IN p_job_id integer, IN p_user_id integer, IN p_status character varying); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON PROCEDURE public.add_new_application(IN p_job_id integer, IN p_user_id integer, IN p_status character varying) TO recruiter;
GRANT ALL ON PROCEDURE public.add_new_application(IN p_job_id integer, IN p_user_id integer, IN p_status character varying) TO job_seeker;


--
-- Name: PROCEDURE add_new_employer(IN p_employername character varying, IN p_password character varying); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON PROCEDURE public.add_new_employer(IN p_employername character varying, IN p_password character varying) TO employer;


--
-- Name: PROCEDURE add_new_interview(IN p_user_id integer, IN p_employer_id integer, IN p_timestamp timestamp without time zone, IN p_text character varying); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON PROCEDURE public.add_new_interview(IN p_user_id integer, IN p_employer_id integer, IN p_timestamp timestamp without time zone, IN p_text character varying) TO recruiter;
GRANT ALL ON PROCEDURE public.add_new_interview(IN p_user_id integer, IN p_employer_id integer, IN p_timestamp timestamp without time zone, IN p_text character varying) TO employer;


--
-- Name: PROCEDURE add_new_job_offer(IN p_employer_id integer, IN p_title character varying, IN p_description character varying, IN p_location character varying, IN p_salaryfrom real, IN p_salaryupto real); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON PROCEDURE public.add_new_job_offer(IN p_employer_id integer, IN p_title character varying, IN p_description character varying, IN p_location character varying, IN p_salaryfrom real, IN p_salaryupto real) TO employer;
GRANT ALL ON PROCEDURE public.add_new_job_offer(IN p_employer_id integer, IN p_title character varying, IN p_description character varying, IN p_location character varying, IN p_salaryfrom real, IN p_salaryupto real) TO recruiter;


--
-- Name: PROCEDURE add_new_message(IN p_sender_id integer, IN p_receiver_id integer, IN p_text character varying); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON PROCEDURE public.add_new_message(IN p_sender_id integer, IN p_receiver_id integer, IN p_text character varying) TO recruiter;
GRANT ALL ON PROCEDURE public.add_new_message(IN p_sender_id integer, IN p_receiver_id integer, IN p_text character varying) TO job_seeker;
GRANT ALL ON PROCEDURE public.add_new_message(IN p_sender_id integer, IN p_receiver_id integer, IN p_text character varying) TO employer;


--
-- Name: PROCEDURE add_new_portfolio(IN p_user_id integer, IN p_text character varying, IN p_website character varying); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON PROCEDURE public.add_new_portfolio(IN p_user_id integer, IN p_text character varying, IN p_website character varying) TO job_seeker;


--
-- Name: PROCEDURE add_new_profile(IN p_user_id integer, IN p_headline character varying, IN p_summary character varying); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON PROCEDURE public.add_new_profile(IN p_user_id integer, IN p_headline character varying, IN p_summary character varying) TO job_seeker;


--
-- Name: PROCEDURE add_new_user(IN p_username character varying, IN p_password character varying); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON PROCEDURE public.add_new_user(IN p_username character varying, IN p_password character varying) TO job_seeker;


--
-- Name: PROCEDURE apply_for_job(IN p_user_id integer, IN p_job_id integer, IN p_message character varying); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON PROCEDURE public.apply_for_job(IN p_user_id integer, IN p_job_id integer, IN p_message character varying) TO job_seeker;


--
-- Name: FUNCTION avg_salary_by_location(p_location character varying); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON FUNCTION public.avg_salary_by_location(p_location character varying) TO employer;
GRANT ALL ON FUNCTION public.avg_salary_by_location(p_location character varying) TO recruiter;
GRANT ALL ON FUNCTION public.avg_salary_by_location(p_location character varying) TO job_seeker;


--
-- Name: FUNCTION decline_other_applications(); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON FUNCTION public.decline_other_applications() TO employer;
GRANT ALL ON FUNCTION public.decline_other_applications() TO recruiter;


--
-- Name: PROCEDURE update_application_status(IN p_app_id integer, IN p_status character varying); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON PROCEDURE public.update_application_status(IN p_app_id integer, IN p_status character varying) TO employer;
GRANT ALL ON PROCEDURE public.update_application_status(IN p_app_id integer, IN p_status character varying) TO recruiter;


--
-- Name: FUNCTION update_user_status_on_hire(); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON FUNCTION public.update_user_status_on_hire() TO employer;
GRANT ALL ON FUNCTION public.update_user_status_on_hire() TO recruiter;


--
-- Name: TABLE applications; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,INSERT ON TABLE public.applications TO job_seeker;
GRANT SELECT ON TABLE public.applications TO employer;
GRANT SELECT,INSERT,UPDATE ON TABLE public.applications TO recruiter;


--
-- Name: TABLE employer; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,UPDATE ON TABLE public.employer TO employer;
GRANT SELECT ON TABLE public.employer TO recruiter;


--
-- Name: TABLE joboffers; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT ON TABLE public.joboffers TO job_seeker;
GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.joboffers TO employer;
GRANT SELECT,INSERT,UPDATE ON TABLE public.joboffers TO recruiter;


--
-- Name: TABLE employer_applicant_stats_view; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT ON TABLE public.employer_applicant_stats_view TO employer;


--
-- Name: TABLE profile; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,INSERT,REFERENCES,DELETE,TRIGGER,TRUNCATE,UPDATE ON TABLE public.profile TO job_seeker;
GRANT SELECT ON TABLE public.profile TO employer;


--
-- Name: TABLE users; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,UPDATE ON TABLE public.users TO job_seeker;
GRANT SELECT ON TABLE public.users TO employer;
GRANT SELECT ON TABLE public.users TO recruiter;


--
-- Name: TABLE interviewscheduler; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.interviewscheduler TO employer;


--
-- Name: TABLE portfolio; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,UPDATE ON TABLE public.portfolio TO job_seeker;
GRANT SELECT ON TABLE public.portfolio TO employer;


--
-- Name: TABLE jobseeker_profile_view; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT ON TABLE public.jobseeker_profile_view TO job_seeker;


--
-- Name: TABLE recruiter_user_view; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT ON TABLE public.recruiter_user_view TO recruiter;


--
-- PostgreSQL database dump complete
--

