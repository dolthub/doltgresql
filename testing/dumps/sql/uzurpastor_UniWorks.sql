-- Downloaded from: https://github.com/uzurpastor/UniWorks/blob/3024fee80cdfb379ffa4b1f96069bce492e5e11a/dump-uniworks.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 13.7 (Debian 13.7-0+deb11u1)
-- Dumped by pg_dump version 13.7 (Debian 13.7-0+deb11u1)

-- Started on 2022-08-18 11:56:29 EEST

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
-- TOC entry 3114 (class 1262 OID 16417)
-- Name: lr_psql; Type: DATABASE; Schema: -; Owner: postgres
--

CREATE DATABASE lr_psql WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE = 'uk_UA.UTF-8';


ALTER DATABASE lr_psql OWNER TO postgres;

\connect lr_psql

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
-- TOC entry 3 (class 2615 OID 2200)
-- Name: public; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA public;


ALTER SCHEMA public OWNER TO postgres;

--
-- TOC entry 3115 (class 0 OID 0)
-- Dependencies: 3
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: postgres
--

COMMENT ON SCHEMA public IS 'standard public schema';


--
-- TOC entry 220 (class 1255 OID 16418)
-- Name: get_film_titles(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_film_titles() RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE 
    titles TEXT DEFAULT '';
    rec_film   RECORD;
    cur_films CURSOR 
       FOR SELECT s.pk_sbj, s."name"
       FROM subject s;
BEGIN
   -- Open the cursor
   OPEN cur_films;
   
   LOOP
    -- fetch row into the film
      FETCH cur_films INTO rec_film;
    -- exit when no more row to fetch
      EXIT WHEN NOT FOUND;
 
    -- build the output
         titles := titles || ',' || rec_film.pk_sbj || ':' || rec_film."name" || '
' ;
   END LOOP;
  
   -- Close the cursor
   CLOSE cur_films;
 
   RETURN titles;
END; $$;


ALTER FUNCTION public.get_film_titles() OWNER TO postgres;

--
-- TOC entry 221 (class 1255 OID 16419)
-- Name: project_employee(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.project_employee() RETURNS TABLE(project_id integer, work_id integer, subject text, employee_info text)
    LANGUAGE plpgsql
    AS $$

declare 	
	project_id 		int;
	work_id 		int;
	subject 		text;
	employee_info 	text;
begin
	select  
		pe.pk_pemp		::int 	project_id,
		pe.fk_work		::int 	work_id,
		s."name"		::text 	subject,
		e.employee_info	::text 	employee_info
	from project_emp pe
	left join  subject s
	on s.pk_sbj = pe.fk_sbj 
	left join (
		select 
			pk_emp,
			(fname || " " || lname || " " || dname)::text as info
		from employee) e 
	on e.pk_emp = pe.fk_emp;
end;
$$;


ALTER FUNCTION public.project_employee() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 200 (class 1259 OID 16420)
-- Name: employee; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.employee (
    pk_emp integer NOT NULL,
    fname character(255) DEFAULT NULL::bpchar,
    lname character(255) DEFAULT NULL::bpchar,
    dname character(255) DEFAULT NULL::bpchar
);


ALTER TABLE public.employee OWNER TO postgres;

--
-- TOC entry 201 (class 1259 OID 16429)
-- Name: employee_pk_emp_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.employee_pk_emp_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.employee_pk_emp_seq OWNER TO postgres;

--
-- TOC entry 3116 (class 0 OID 0)
-- Dependencies: 201
-- Name: employee_pk_emp_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.employee_pk_emp_seq OWNED BY public.employee.pk_emp;


--
-- TOC entry 219 (class 1259 OID 16659)
-- Name: group; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."group" (
    pk_g integer NOT NULL,
    name text
);


ALTER TABLE public."group" OWNER TO postgres;

--
-- TOC entry 218 (class 1259 OID 16657)
-- Name: group_pk_g_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.group_pk_g_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.group_pk_g_seq OWNER TO postgres;

--
-- TOC entry 3117 (class 0 OID 0)
-- Dependencies: 218
-- Name: group_pk_g_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.group_pk_g_seq OWNED BY public."group".pk_g;


--
-- TOC entry 202 (class 1259 OID 16434)
-- Name: project_emp; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.project_emp (
    pk_pemp integer NOT NULL,
    fk_work integer,
    fk_sbj integer,
    fk_emp integer
);


ALTER TABLE public.project_emp OWNER TO postgres;

--
-- TOC entry 203 (class 1259 OID 16437)
-- Name: project_emp_pk_pemp_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.project_emp_pk_pemp_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.project_emp_pk_pemp_seq OWNER TO postgres;

--
-- TOC entry 3118 (class 0 OID 0)
-- Dependencies: 203
-- Name: project_emp_pk_pemp_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.project_emp_pk_pemp_seq OWNED BY public.project_emp.pk_pemp;


--
-- TOC entry 204 (class 1259 OID 16439)
-- Name: subject; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.subject (
    pk_sbj integer NOT NULL,
    name character(255) DEFAULT NULL::bpchar
);


ALTER TABLE public.subject OWNER TO postgres;

--
-- TOC entry 205 (class 1259 OID 16443)
-- Name: project_employee; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.project_employee AS
 SELECT pe.pk_pemp AS project_id,
    pe.fk_work AS work_id,
    s.name AS subject,
    e.info
   FROM ((public.project_emp pe
     LEFT JOIN public.subject s ON ((s.pk_sbj = pe.fk_sbj)))
     LEFT JOIN ( SELECT employee.pk_emp,
            (((((employee.fname)::text || ' '::text) || (employee.lname)::text) || ' '::text) || (employee.dname)::text) AS info
           FROM public.employee) e ON ((e.pk_emp = pe.fk_emp)));


ALTER TABLE public.project_employee OWNER TO postgres;

--
-- TOC entry 206 (class 1259 OID 16448)
-- Name: project_st; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.project_st (
    pk_pst integer NOT NULL,
    fk_st integer,
    fk_work integer
);


ALTER TABLE public.project_st OWNER TO postgres;

--
-- TOC entry 207 (class 1259 OID 16451)
-- Name: project_st_pk_pst_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.project_st_pk_pst_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.project_st_pk_pst_seq OWNER TO postgres;

--
-- TOC entry 3119 (class 0 OID 0)
-- Dependencies: 207
-- Name: project_st_pk_pst_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.project_st_pk_pst_seq OWNED BY public.project_st.pk_pst;


--
-- TOC entry 208 (class 1259 OID 16453)
-- Name: project_tl; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.project_tl (
    pk_ptl integer NOT NULL,
    fk_work integer,
    fk_tl integer,
    comment text
);


ALTER TABLE public.project_tl OWNER TO postgres;

--
-- TOC entry 209 (class 1259 OID 16459)
-- Name: project_tl_pk_ptl_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.project_tl_pk_ptl_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.project_tl_pk_ptl_seq OWNER TO postgres;

--
-- TOC entry 3120 (class 0 OID 0)
-- Dependencies: 209
-- Name: project_tl_pk_ptl_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.project_tl_pk_ptl_seq OWNED BY public.project_tl.pk_ptl;


--
-- TOC entry 210 (class 1259 OID 16461)
-- Name: student; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.student (
    pk_st integer NOT NULL,
    fname character(255) DEFAULT NULL::bpchar,
    lname character(255) DEFAULT NULL::bpchar,
    dname character(255) DEFAULT NULL::bpchar,
    "group" integer
);


ALTER TABLE public.student OWNER TO postgres;

--
-- TOC entry 211 (class 1259 OID 16471)
-- Name: student_pk_st_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.student_pk_st_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.student_pk_st_seq OWNER TO postgres;

--
-- TOC entry 3121 (class 0 OID 0)
-- Dependencies: 211
-- Name: student_pk_st_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.student_pk_st_seq OWNED BY public.student.pk_st;


--
-- TOC entry 212 (class 1259 OID 16473)
-- Name: subject_pk_sbj_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.subject_pk_sbj_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.subject_pk_sbj_seq OWNER TO postgres;

--
-- TOC entry 3122 (class 0 OID 0)
-- Dependencies: 212
-- Name: subject_pk_sbj_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.subject_pk_sbj_seq OWNED BY public.subject.pk_sbj;


--
-- TOC entry 213 (class 1259 OID 16475)
-- Name: work; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.work (
    pk_w integer NOT NULL,
    name character(255) NOT NULL,
    article text NOT NULL,
    realise_year date NOT NULL
);


ALTER TABLE public.work OWNER TO postgres;

--
-- TOC entry 214 (class 1259 OID 16481)
-- Name: summary; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.summary AS
 SELECT w.pk_w,
    w.name,
    w.article,
    w.realise_year,
    ps.fk_st
   FROM (public.work w
     JOIN public.project_st ps ON ((w.pk_w = ps.fk_work)));


ALTER TABLE public.summary OWNER TO postgres;

--
-- TOC entry 215 (class 1259 OID 16485)
-- Name: tool; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tool (
    pk_tl integer NOT NULL,
    name character(255) DEFAULT NULL::bpchar
);


ALTER TABLE public.tool OWNER TO postgres;

--
-- TOC entry 216 (class 1259 OID 16489)
-- Name: tool_pk_tl_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.tool_pk_tl_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.tool_pk_tl_seq OWNER TO postgres;

--
-- TOC entry 3123 (class 0 OID 0)
-- Dependencies: 216
-- Name: tool_pk_tl_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.tool_pk_tl_seq OWNED BY public.tool.pk_tl;


--
-- TOC entry 217 (class 1259 OID 16491)
-- Name: work_pk_w_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.work_pk_w_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.work_pk_w_seq OWNER TO postgres;

--
-- TOC entry 3124 (class 0 OID 0)
-- Dependencies: 217
-- Name: work_pk_w_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.work_pk_w_seq OWNED BY public.work.pk_w;


--
-- TOC entry 2919 (class 2604 OID 16559)
-- Name: employee pk_emp; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employee ALTER COLUMN pk_emp SET DEFAULT nextval('public.employee_pk_emp_seq'::regclass);


--
-- TOC entry 2932 (class 2604 OID 16662)
-- Name: group pk_g; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."group" ALTER COLUMN pk_g SET DEFAULT nextval('public.group_pk_g_seq'::regclass);


--
-- TOC entry 2920 (class 2604 OID 16560)
-- Name: project_emp pk_pemp; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_emp ALTER COLUMN pk_pemp SET DEFAULT nextval('public.project_emp_pk_pemp_seq'::regclass);


--
-- TOC entry 2923 (class 2604 OID 16561)
-- Name: project_st pk_pst; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_st ALTER COLUMN pk_pst SET DEFAULT nextval('public.project_st_pk_pst_seq'::regclass);


--
-- TOC entry 2924 (class 2604 OID 16562)
-- Name: project_tl pk_ptl; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_tl ALTER COLUMN pk_ptl SET DEFAULT nextval('public.project_tl_pk_ptl_seq'::regclass);


--
-- TOC entry 2928 (class 2604 OID 16563)
-- Name: student pk_st; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student ALTER COLUMN pk_st SET DEFAULT nextval('public.student_pk_st_seq'::regclass);


--
-- TOC entry 2922 (class 2604 OID 16564)
-- Name: subject pk_sbj; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subject ALTER COLUMN pk_sbj SET DEFAULT nextval('public.subject_pk_sbj_seq'::regclass);


--
-- TOC entry 2931 (class 2604 OID 16565)
-- Name: tool pk_tl; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tool ALTER COLUMN pk_tl SET DEFAULT nextval('public.tool_pk_tl_seq'::regclass);


--
-- TOC entry 2929 (class 2604 OID 16566)
-- Name: work pk_w; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.work ALTER COLUMN pk_w SET DEFAULT nextval('public.work_pk_w_seq'::regclass);


--
-- TOC entry 3091 (class 0 OID 16420)
-- Dependencies: 200
-- Data for Name: employee; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.employee (pk_emp, fname, lname, dname) FROM stdin;
2	pash                                                                                                                                                                                                                                                           	opor                                                                                                                                                                                                                                                           	lohhh                                                                                                                                                                                                                                                          
1	sasha                                                                                                                                                                                                                                                          	fffff                                                                                                                                                                                                                                                          	dadov                                                                                                                                                                                                                                                          
\.


--
-- TOC entry 3108 (class 0 OID 16659)
-- Dependencies: 219
-- Data for Name: group; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public."group" (pk_g, name) FROM stdin;
1	xxx
2	dfdf
\.


--
-- TOC entry 3093 (class 0 OID 16434)
-- Dependencies: 202
-- Data for Name: project_emp; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.project_emp (pk_pemp, fk_work, fk_sbj, fk_emp) FROM stdin;
1	4	3	2
2	5	2	1
\.


--
-- TOC entry 3096 (class 0 OID 16448)
-- Dependencies: 206
-- Data for Name: project_st; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.project_st (pk_pst, fk_st, fk_work) FROM stdin;
1	1	4
2	2	5
\.


--
-- TOC entry 3098 (class 0 OID 16453)
-- Dependencies: 208
-- Data for Name: project_tl; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.project_tl (pk_ptl, fk_work, fk_tl, comment) FROM stdin;
1	4	1	some
2	5	2	other
\.


--
-- TOC entry 3100 (class 0 OID 16461)
-- Dependencies: 210
-- Data for Name: student; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.student (pk_st, fname, lname, dname, "group") FROM stdin;
1	andrey                                                                                                                                                                                                                                                         	shevch                                                                                                                                                                                                                                                         	roman                                                                                                                                                                                                                                                          	2
2	sacha                                                                                                                                                                                                                                                          	pupkin                                                                                                                                                                                                                                                         	loh                                                                                                                                                                                                                                                            	1
\.


--
-- TOC entry 3095 (class 0 OID 16439)
-- Dependencies: 204
-- Data for Name: subject; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.subject (pk_sbj, name) FROM stdin;
1	math                                                                                                                                                                                                                                                           
2	oop                                                                                                                                                                                                                                                            
3	fp                                                                                                                                                                                                                                                             
\.


--
-- TOC entry 3104 (class 0 OID 16485)
-- Dependencies: 215
-- Data for Name: tool; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tool (pk_tl, name) FROM stdin;
1	ruby postgres                                                                                                                                                                                                                                                  
2	rust qt                                                                                                                                                                                                                                                        
3	java                                                                                                                                                                                                                                                           
\.


--
-- TOC entry 3103 (class 0 OID 16475)
-- Dependencies: 213
-- Data for Name: work; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.work (pk_w, name, article, realise_year) FROM stdin;
3	sport                                                                                                                                                                                                                                                          	describe exersice in gym	2022-05-27
4	shop                                                                                                                                                                                                                                                           	describe sexshop	2022-05-07
5	terrarium                                                                                                                                                                                                                                                      	describe animal classes in zoo	2022-05-11
\.


--
-- TOC entry 3125 (class 0 OID 0)
-- Dependencies: 201
-- Name: employee_pk_emp_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.employee_pk_emp_seq', 2, true);


--
-- TOC entry 3126 (class 0 OID 0)
-- Dependencies: 218
-- Name: group_pk_g_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.group_pk_g_seq', 1, false);


--
-- TOC entry 3127 (class 0 OID 0)
-- Dependencies: 203
-- Name: project_emp_pk_pemp_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.project_emp_pk_pemp_seq', 2, true);


--
-- TOC entry 3128 (class 0 OID 0)
-- Dependencies: 207
-- Name: project_st_pk_pst_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.project_st_pk_pst_seq', 2, true);


--
-- TOC entry 3129 (class 0 OID 0)
-- Dependencies: 209
-- Name: project_tl_pk_ptl_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.project_tl_pk_ptl_seq', 2, true);


--
-- TOC entry 3130 (class 0 OID 0)
-- Dependencies: 211
-- Name: student_pk_st_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.student_pk_st_seq', 2, true);


--
-- TOC entry 3131 (class 0 OID 0)
-- Dependencies: 212
-- Name: subject_pk_sbj_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.subject_pk_sbj_seq', 3, true);


--
-- TOC entry 3132 (class 0 OID 0)
-- Dependencies: 216
-- Name: tool_pk_tl_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.tool_pk_tl_seq', 3, true);


--
-- TOC entry 3133 (class 0 OID 0)
-- Dependencies: 217
-- Name: work_pk_w_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.work_pk_w_seq', 5, true);


--
-- TOC entry 2934 (class 2606 OID 16502)
-- Name: employee employee_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employee
    ADD CONSTRAINT employee_pkey PRIMARY KEY (pk_emp);


--
-- TOC entry 2950 (class 2606 OID 16667)
-- Name: group group_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."group"
    ADD CONSTRAINT group_pkey PRIMARY KEY (pk_g);


--
-- TOC entry 2936 (class 2606 OID 16506)
-- Name: project_emp project_emp_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_emp
    ADD CONSTRAINT project_emp_pkey PRIMARY KEY (pk_pemp);


--
-- TOC entry 2940 (class 2606 OID 16508)
-- Name: project_st project_st_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_st
    ADD CONSTRAINT project_st_pkey PRIMARY KEY (pk_pst);


--
-- TOC entry 2942 (class 2606 OID 16510)
-- Name: project_tl project_tl_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_tl
    ADD CONSTRAINT project_tl_pkey PRIMARY KEY (pk_ptl);


--
-- TOC entry 2944 (class 2606 OID 16512)
-- Name: student student_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student
    ADD CONSTRAINT student_pkey PRIMARY KEY (pk_st);


--
-- TOC entry 2938 (class 2606 OID 16514)
-- Name: subject subject_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subject
    ADD CONSTRAINT subject_pkey PRIMARY KEY (pk_sbj);


--
-- TOC entry 2948 (class 2606 OID 16516)
-- Name: tool tool_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tool
    ADD CONSTRAINT tool_pkey PRIMARY KEY (pk_tl);


--
-- TOC entry 2946 (class 2606 OID 16518)
-- Name: work work_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.work
    ADD CONSTRAINT work_pkey PRIMARY KEY (pk_w);


--
-- TOC entry 2951 (class 2606 OID 16519)
-- Name: project_emp project_emp_fk_emp_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_emp
    ADD CONSTRAINT project_emp_fk_emp_fkey FOREIGN KEY (fk_emp) REFERENCES public.employee(pk_emp);


--
-- TOC entry 2952 (class 2606 OID 16524)
-- Name: project_emp project_emp_fk_sbj_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_emp
    ADD CONSTRAINT project_emp_fk_sbj_fkey FOREIGN KEY (fk_sbj) REFERENCES public.subject(pk_sbj);


--
-- TOC entry 2953 (class 2606 OID 16529)
-- Name: project_emp project_emp_fk_work_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_emp
    ADD CONSTRAINT project_emp_fk_work_fkey FOREIGN KEY (fk_work) REFERENCES public.work(pk_w);


--
-- TOC entry 2954 (class 2606 OID 16534)
-- Name: project_st project_st_fk_st_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_st
    ADD CONSTRAINT project_st_fk_st_fkey FOREIGN KEY (fk_st) REFERENCES public.student(pk_st);


--
-- TOC entry 2955 (class 2606 OID 16539)
-- Name: project_st project_st_fk_work_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_st
    ADD CONSTRAINT project_st_fk_work_fkey FOREIGN KEY (fk_work) REFERENCES public.work(pk_w);


--
-- TOC entry 2956 (class 2606 OID 16544)
-- Name: project_tl project_tl_fk_tl_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_tl
    ADD CONSTRAINT project_tl_fk_tl_fkey FOREIGN KEY (fk_tl) REFERENCES public.tool(pk_tl);


--
-- TOC entry 2957 (class 2606 OID 16549)
-- Name: project_tl project_tl_fk_work_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_tl
    ADD CONSTRAINT project_tl_fk_work_fkey FOREIGN KEY (fk_work) REFERENCES public.work(pk_w);


--
-- TOC entry 2958 (class 2606 OID 16673)
-- Name: student student_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student
    ADD CONSTRAINT student_fk FOREIGN KEY ("group") REFERENCES public."group"(pk_g);


-- Completed on 2022-08-18 11:56:29 EEST

--
-- PostgreSQL database dump complete
--

