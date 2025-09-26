-- Downloaded from: https://github.com/StronglogicSolutions/kserver/blob/79d3a3d493462ea18a0738a4d1887a946da06ff1/kiq_schema.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 13.6 (Ubuntu 13.6-0ubuntu0.21.10.1)
-- Dumped by pg_dump version 14.10 (Ubuntu 14.10-0ubuntu0.22.04.1)

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
-- Name: get_recurring_seconds(integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_recurring_seconds(n integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$                                                    
 BEGIN                                                            
 CASE n                                                           
 WHEN 0 THEN                                                      
 RETURN 0;                                                        
 WHEN 1 THEN                                                      
 RETURN 3600;                                                     
 WHEN 2 THEN                                                      
 RETURN 86400;                                                    
 WHEN 3 THEN                                                      
 RETURN 604800;                                                  
 WHEN 4 THEN
 RETURN 2419200;
 WHEN 5 THEN                                                      
 RETURN 31536000;                                                 
 END CASE;                                                        
 END;                                                             
 $$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: affiliation; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.affiliation (
    id integer NOT NULL,
    pid integer NOT NULL,
    oid integer NOT NULL
);


--
-- Name: affiliation_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.affiliation_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: affiliation_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.affiliation_id_seq OWNED BY public.affiliation.id;


--
-- Name: apps; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.apps (
    id integer NOT NULL,
    path text,
    data text,
    mask integer,
    name text,
    internal boolean DEFAULT false
);


--
-- Name: apps_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.apps_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: apps_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.apps_id_seq OWNED BY public.apps.id;


--
-- Name: file; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.file (
    id integer NOT NULL,
    name text,
    sid integer,
    type text
);


--
-- Name: file_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.file_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: file_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.file_id_seq OWNED BY public.file.id;


--
-- Name: ipc; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ipc (
    id integer NOT NULL,
    "time" integer NOT NULL,
    pid text,
    command text,
    data text,
    status integer DEFAULT 0,
    p_uuid text,
    type integer DEFAULT 0,
    recurring integer DEFAULT 0
);


--
-- Name: ipc_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.ipc_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ipc_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.ipc_id_seq OWNED BY public.ipc.id;


--
-- Name: organization; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organization (
    id integer NOT NULL,
    name text NOT NULL
);


--
-- Name: organization_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.organization_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: organization_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.organization_id_seq OWNED BY public.organization.id;


--
-- Name: person; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.person (
    id integer NOT NULL,
    name text NOT NULL
);


--
-- Name: person_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.person_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: person_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.person_id_seq OWNED BY public.person.id;


--
-- Name: platform; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.platform (
    id integer NOT NULL,
    name text,
    method text DEFAULT 'process'::text NOT NULL
);


--
-- Name: platform_affiliate_user; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.platform_affiliate_user (
    id integer NOT NULL,
    uid integer NOT NULL,
    a_uid integer NOT NULL
);


--
-- Name: platform_affiliate_user_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.platform_affiliate_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: platform_affiliate_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.platform_affiliate_user_id_seq OWNED BY public.platform_affiliate_user.id;


--
-- Name: platform_filter; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.platform_filter (
    id integer NOT NULL,
    pid integer,
    value text,
    type text,
    rpid integer
);


--
-- Name: platform_filter_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.platform_filter_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: platform_filter_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.platform_filter_id_seq OWNED BY public.platform_filter.id;


--
-- Name: platform_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.platform_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: platform_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.platform_id_seq OWNED BY public.platform.id;


--
-- Name: platform_post; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.platform_post (
    id integer NOT NULL,
    pid integer NOT NULL,
    o_pid integer,
    unique_id text NOT NULL,
    "time" integer,
    status integer DEFAULT 0 NOT NULL,
    repost boolean DEFAULT false,
    uid integer DEFAULT 0 NOT NULL
);


--
-- Name: platform_post_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.platform_post_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: platform_post_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.platform_post_id_seq OWNED BY public.platform_post.id;


--
-- Name: platform_repost; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.platform_repost (
    id integer NOT NULL,
    pid integer NOT NULL,
    r_pid integer NOT NULL
);


--
-- Name: platform_repost_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.platform_repost_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: platform_repost_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.platform_repost_id_seq OWNED BY public.platform_repost.id;


--
-- Name: platform_user; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.platform_user (
    id integer NOT NULL,
    pid integer NOT NULL,
    name text NOT NULL,
    type text,
    pers_id integer
);


--
-- Name: platform_user_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.platform_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: platform_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.platform_user_id_seq OWNED BY public.platform_user.id;


--
-- Name: process_result; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.process_result (
    id integer NOT NULL,
    aid integer NOT NULL,
    "time" integer NOT NULL,
    status integer NOT NULL
);


--
-- Name: process_result_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.process_result_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: process_result_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.process_result_id_seq OWNED BY public.process_result.id;


--
-- Name: recurring; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.recurring (
    id integer NOT NULL,
    sid integer NOT NULL,
    "time" integer
);


--
-- Name: recurring_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.recurring_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: recurring_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.recurring_id_seq OWNED BY public.recurring.id;


--
-- Name: schedule; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schedule (
    id integer NOT NULL,
    mask integer,
    flags text,
    envfile text,
    "time" integer,
    completed integer DEFAULT 0,
    recurring integer DEFAULT 0,
    notify boolean DEFAULT false,
    runtime text
);


--
-- Name: schedule_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.schedule_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: schedule_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.schedule_id_seq OWNED BY public.schedule.id;


--
-- Name: term; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.term (
    id integer NOT NULL,
    name text NOT NULL,
    type text
);


--
-- Name: term_hit; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.term_hit (
    id integer NOT NULL,
    tid integer NOT NULL,
    uid integer NOT NULL,
    "time" timestamp without time zone DEFAULT now() NOT NULL,
    sid integer
);


--
-- Name: term_hit_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.term_hit_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: term_hit_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.term_hit_id_seq OWNED BY public.term_hit.id;


--
-- Name: term_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.term_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: term_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.term_id_seq OWNED BY public.term.id;


--
-- Name: trigger_config; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.trigger_config (
    id integer NOT NULL,
    tid integer NOT NULL,
    token_name text,
    section text,
    name text
);


--
-- Name: trigger_config_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.trigger_config_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: trigger_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.trigger_config_id_seq OWNED BY public.trigger_config.id;


--
-- Name: trigger_map; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.trigger_map (
    id integer NOT NULL,
    tid integer NOT NULL,
    old text,
    new text
);


--
-- Name: trigger_map_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.trigger_map_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: trigger_map_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.trigger_map_id_seq OWNED BY public.trigger_map.id;


--
-- Name: triggers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.triggers (
    id integer NOT NULL,
    mask integer,
    trigger_mask integer,
    token_name text,
    token_value text
);


--
-- Name: triggers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.triggers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: triggers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.triggers_id_seq OWNED BY public.triggers.id;


--
-- Name: affiliation id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliation ALTER COLUMN id SET DEFAULT nextval('public.affiliation_id_seq'::regclass);


--
-- Name: apps id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.apps ALTER COLUMN id SET DEFAULT nextval('public.apps_id_seq'::regclass);


--
-- Name: file id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.file ALTER COLUMN id SET DEFAULT nextval('public.file_id_seq'::regclass);


--
-- Name: ipc id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ipc ALTER COLUMN id SET DEFAULT nextval('public.ipc_id_seq'::regclass);


--
-- Name: organization id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization ALTER COLUMN id SET DEFAULT nextval('public.organization_id_seq'::regclass);


--
-- Name: person id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.person ALTER COLUMN id SET DEFAULT nextval('public.person_id_seq'::regclass);


--
-- Name: platform id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform ALTER COLUMN id SET DEFAULT nextval('public.platform_id_seq'::regclass);


--
-- Name: platform_affiliate_user id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_affiliate_user ALTER COLUMN id SET DEFAULT nextval('public.platform_affiliate_user_id_seq'::regclass);


--
-- Name: platform_filter id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_filter ALTER COLUMN id SET DEFAULT nextval('public.platform_filter_id_seq'::regclass);


--
-- Name: platform_post id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_post ALTER COLUMN id SET DEFAULT nextval('public.platform_post_id_seq'::regclass);


--
-- Name: platform_repost id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_repost ALTER COLUMN id SET DEFAULT nextval('public.platform_repost_id_seq'::regclass);


--
-- Name: platform_user id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_user ALTER COLUMN id SET DEFAULT nextval('public.platform_user_id_seq'::regclass);


--
-- Name: process_result id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.process_result ALTER COLUMN id SET DEFAULT nextval('public.process_result_id_seq'::regclass);


--
-- Name: recurring id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recurring ALTER COLUMN id SET DEFAULT nextval('public.recurring_id_seq'::regclass);


--
-- Name: schedule id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schedule ALTER COLUMN id SET DEFAULT nextval('public.schedule_id_seq'::regclass);


--
-- Name: term id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term ALTER COLUMN id SET DEFAULT nextval('public.term_id_seq'::regclass);


--
-- Name: term_hit id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_hit ALTER COLUMN id SET DEFAULT nextval('public.term_hit_id_seq'::regclass);


--
-- Name: trigger_config id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trigger_config ALTER COLUMN id SET DEFAULT nextval('public.trigger_config_id_seq'::regclass);


--
-- Name: trigger_map id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trigger_map ALTER COLUMN id SET DEFAULT nextval('public.trigger_map_id_seq'::regclass);


--
-- Name: triggers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.triggers ALTER COLUMN id SET DEFAULT nextval('public.triggers_id_seq'::regclass);


--
-- Name: affiliation affiliation_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliation
    ADD CONSTRAINT affiliation_pkey PRIMARY KEY (id);


--
-- Name: apps apps_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.apps
    ADD CONSTRAINT apps_pkey PRIMARY KEY (id);


--
-- Name: file file_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.file
    ADD CONSTRAINT file_pkey PRIMARY KEY (id);


--
-- Name: ipc ipc_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ipc
    ADD CONSTRAINT ipc_pkey PRIMARY KEY (id);


--
-- Name: organization organization_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization
    ADD CONSTRAINT organization_pkey PRIMARY KEY (id);


--
-- Name: person person_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.person
    ADD CONSTRAINT person_pkey PRIMARY KEY (id);


--
-- Name: platform_affiliate_user platform_affiliate_user_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_affiliate_user
    ADD CONSTRAINT platform_affiliate_user_pkey PRIMARY KEY (id, uid, a_uid);


--
-- Name: platform_filter platform_filter_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_filter
    ADD CONSTRAINT platform_filter_pkey PRIMARY KEY (id);


--
-- Name: platform platform_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform
    ADD CONSTRAINT platform_pkey PRIMARY KEY (id);


--
-- Name: platform_post platform_post_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_post
    ADD CONSTRAINT platform_post_pkey PRIMARY KEY (id, pid);


--
-- Name: platform_repost platform_repost_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_repost
    ADD CONSTRAINT platform_repost_pkey PRIMARY KEY (id, pid);


--
-- Name: platform_user platform_user_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_user
    ADD CONSTRAINT platform_user_pkey PRIMARY KEY (id, pid, name);


--
-- Name: process_result process_result_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.process_result
    ADD CONSTRAINT process_result_pkey PRIMARY KEY (id, aid);


--
-- Name: recurring recurring_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recurring
    ADD CONSTRAINT recurring_pkey PRIMARY KEY (id, sid);


--
-- Name: schedule schedule_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schedule
    ADD CONSTRAINT schedule_pkey PRIMARY KEY (id);


--
-- Name: term_hit term_hit_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_hit
    ADD CONSTRAINT term_hit_pkey PRIMARY KEY (id);


--
-- Name: term term_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term
    ADD CONSTRAINT term_pkey PRIMARY KEY (id);


--
-- Name: trigger_config trigger_config_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trigger_config
    ADD CONSTRAINT trigger_config_pkey PRIMARY KEY (id, tid);


--
-- Name: trigger_map trigger_map_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trigger_map
    ADD CONSTRAINT trigger_map_pkey PRIMARY KEY (id, tid);


--
-- Name: triggers triggers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.triggers
    ADD CONSTRAINT triggers_pkey PRIMARY KEY (id);


--
-- Name: platform_user u_id_const; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_user
    ADD CONSTRAINT u_id_const UNIQUE (id);


--
-- Name: apps unique_mask; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.apps
    ADD CONSTRAINT unique_mask UNIQUE (mask);


--
-- Name: platform_post unique_user_post; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_post
    ADD CONSTRAINT unique_user_post UNIQUE (pid, unique_id, uid);


--
-- Name: file file_sid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.file
    ADD CONSTRAINT file_sid_fkey FOREIGN KEY (sid) REFERENCES public.schedule(id);


--
-- Name: process_result fk_app; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.process_result
    ADD CONSTRAINT fk_app FOREIGN KEY (aid) REFERENCES public.apps(id);


--
-- Name: platform_post fk_platform; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_post
    ADD CONSTRAINT fk_platform FOREIGN KEY (pid) REFERENCES public.platform(id);


--
-- Name: platform_repost fk_platform; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_repost
    ADD CONSTRAINT fk_platform FOREIGN KEY (pid) REFERENCES public.platform(id);


--
-- Name: platform_filter fk_platform; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_filter
    ADD CONSTRAINT fk_platform FOREIGN KEY (pid) REFERENCES public.platform(id);


--
-- Name: platform_post fk_platform_origin; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_post
    ADD CONSTRAINT fk_platform_origin FOREIGN KEY (o_pid) REFERENCES public.platform(id);


--
-- Name: platform_repost fk_platform_repost; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_repost
    ADD CONSTRAINT fk_platform_repost FOREIGN KEY (r_pid) REFERENCES public.platform(id);


--
-- Name: trigger_map fk_trigger; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trigger_map
    ADD CONSTRAINT fk_trigger FOREIGN KEY (tid) REFERENCES public.triggers(id);


--
-- Name: trigger_config fk_trigger; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trigger_config
    ADD CONSTRAINT fk_trigger FOREIGN KEY (tid) REFERENCES public.triggers(id);


--
-- Name: affiliation org_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliation
    ADD CONSTRAINT org_fkey FOREIGN KEY (oid) REFERENCES public.organization(id) NOT VALID;


--
-- Name: platform_user person_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_user
    ADD CONSTRAINT person_fk FOREIGN KEY (pers_id) REFERENCES public.person(id);


--
-- Name: affiliation person_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.affiliation
    ADD CONSTRAINT person_fkey FOREIGN KEY (pid) REFERENCES public.person(id) NOT VALID;


--
-- Name: platform_affiliate_user platform_affiliate_user_a_uid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_affiliate_user
    ADD CONSTRAINT platform_affiliate_user_a_uid_fkey FOREIGN KEY (a_uid) REFERENCES public.platform_user(id);


--
-- Name: platform_affiliate_user platform_affiliate_user_uid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_affiliate_user
    ADD CONSTRAINT platform_affiliate_user_uid_fkey FOREIGN KEY (uid) REFERENCES public.platform_user(id);


--
-- Name: platform_filter platform_filter_rpid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_filter
    ADD CONSTRAINT platform_filter_rpid_fkey FOREIGN KEY (rpid) REFERENCES public.platform(id);


--
-- Name: platform_user platform_user_pid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platform_user
    ADD CONSTRAINT platform_user_pid_fkey FOREIGN KEY (pid) REFERENCES public.platform(id);


--
-- Name: recurring recurring_sid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recurring
    ADD CONSTRAINT recurring_sid_fkey FOREIGN KEY (sid) REFERENCES public.schedule(id);


--
-- Name: term_hit schedule_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_hit
    ADD CONSTRAINT schedule_fkey FOREIGN KEY (sid) REFERENCES public.schedule(id);


--
-- Name: term_hit term_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_hit
    ADD CONSTRAINT term_fkey FOREIGN KEY (tid) REFERENCES public.term(id) NOT VALID;


--
-- Name: triggers triggers_mask_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.triggers
    ADD CONSTRAINT triggers_mask_fkey FOREIGN KEY (mask) REFERENCES public.apps(mask);


--
-- Name: triggers triggers_trigger_mask_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.triggers
    ADD CONSTRAINT triggers_trigger_mask_fkey FOREIGN KEY (trigger_mask) REFERENCES public.apps(mask);


--
-- Name: term_hit user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_hit
    ADD CONSTRAINT user_fkey FOREIGN KEY (uid) REFERENCES public.platform_user(id);


--
-- PostgreSQL database dump complete
--

