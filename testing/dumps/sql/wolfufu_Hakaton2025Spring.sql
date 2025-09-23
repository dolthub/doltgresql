-- Downloaded from: https://github.com/wolfufu/Hakaton2025Spring/blob/5a1e4699147c858c7d4ceb6b637036cad1627872/backup/backup.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 17.4
-- Dumped by pg_dump version 17.4

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
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: authenticate_user(character, character); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.authenticate_user(p_login character, p_password character) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
    stored_hash TEXT;
BEGIN
    -- Получаем хеш пароля из базы данных
    SELECT user_password INTO stored_hash
    FROM public.users
    WHERE user_login = p_login;

    -- Если пользователь не найден, возвращаем FALSE
    IF NOT FOUND THEN
        RETURN FALSE;
    END IF;

    -- Сравниваем хеш введенного пароля с хранимым хешем
    RETURN stored_hash = crypt(p_password, stored_hash);
END;
$$;


ALTER FUNCTION public.authenticate_user(p_login character, p_password character) OWNER TO postgres;

--
-- Name: register_user(character, character, character, character); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.register_user(p_login character, p_password character, p_role character DEFAULT 'user'::bpchar, p_notification character DEFAULT '0'::bpchar) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    hashed_password TEXT;
BEGIN
    -- Проверяем, существует ли уже пользователь с таким логином
    IF EXISTS (SELECT 1 FROM public.users WHERE user_login = p_login) THEN
        RAISE EXCEPTION 'Пользователь с логином "%" уже существует', p_login;
    END IF;

    -- Хешируем пароль
    hashed_password := crypt(p_password, gen_salt('bf'));

    -- Вставляем данные в таблицу users
    INSERT INTO public.users (user_login, user_password, user_role, notification)
    VALUES (p_login, hashed_password, p_role, p_notification);
END;
$$;


ALTER FUNCTION public.register_user(p_login character, p_password character, p_role character, p_notification character) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: folders; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.folders (
    id integer NOT NULL,
    userlog character(20) NOT NULL,
    name character(20),
    newsid integer
);


ALTER TABLE public.folders OWNER TO postgres;

--
-- Name: folders_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.folders_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.folders_id_seq OWNER TO postgres;

--
-- Name: folders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.folders_id_seq OWNED BY public.folders.id;


--
-- Name: news; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.news (
    id integer NOT NULL,
    type_news boolean DEFAULT true,
    title text NOT NULL,
    content text NOT NULL,
    tag integer,
    source integer,
    date timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.news OWNER TO postgres;

--
-- Name: news_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.news_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.news_id_seq OWNER TO postgres;

--
-- Name: news_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.news_id_seq OWNED BY public.news.id;


--
-- Name: offers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.offers (
    id integer NOT NULL,
    user_id integer NOT NULL,
    link text NOT NULL
);


ALTER TABLE public.offers OWNER TO postgres;

--
-- Name: offers_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.offers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.offers_id_seq OWNER TO postgres;

--
-- Name: offers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.offers_id_seq OWNED BY public.offers.id;


--
-- Name: sources; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sources (
    id integer NOT NULL,
    name character(20) NOT NULL,
    link text NOT NULL
);


ALTER TABLE public.sources OWNER TO postgres;

--
-- Name: sources_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.sources_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sources_id_seq OWNER TO postgres;

--
-- Name: sources_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.sources_id_seq OWNED BY public.sources.id;


--
-- Name: tags; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tags (
    id integer NOT NULL,
    name character(20) NOT NULL
);


ALTER TABLE public.tags OWNER TO postgres;

--
-- Name: tags_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.tags_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.tags_id_seq OWNER TO postgres;

--
-- Name: tags_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.tags_id_seq OWNED BY public.tags.id;


--
-- Name: tags_news; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tags_news (
    id integer NOT NULL,
    tagid integer NOT NULL,
    newsid integer NOT NULL
);


ALTER TABLE public.tags_news OWNER TO postgres;

--
-- Name: tags_news_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.tags_news_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.tags_news_id_seq OWNER TO postgres;

--
-- Name: tags_news_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.tags_news_id_seq OWNED BY public.tags_news.id;


--
-- Name: user_source_subscriptions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_source_subscriptions (
    user_id integer NOT NULL,
    source_id integer NOT NULL
);


ALTER TABLE public.user_source_subscriptions OWNER TO postgres;

--
-- Name: user_tag_subscriptions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_tag_subscriptions (
    user_id integer NOT NULL,
    tag_id integer NOT NULL
);


ALTER TABLE public.user_tag_subscriptions OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id integer NOT NULL,
    user_login character(20) NOT NULL,
    user_password character(100) NOT NULL,
    user_role character(15) NOT NULL,
    notification character(20) DEFAULT '0'::bpchar,
    tag_subscription integer,
    sources_subsc integer,
    telegram_id character(100) DEFAULT 1000000000 NOT NULL
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_id_seq OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: folders id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.folders ALTER COLUMN id SET DEFAULT nextval('public.folders_id_seq'::regclass);


--
-- Name: news id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.news ALTER COLUMN id SET DEFAULT nextval('public.news_id_seq'::regclass);


--
-- Name: offers id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.offers ALTER COLUMN id SET DEFAULT nextval('public.offers_id_seq'::regclass);


--
-- Name: sources id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sources ALTER COLUMN id SET DEFAULT nextval('public.sources_id_seq'::regclass);


--
-- Name: tags id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tags ALTER COLUMN id SET DEFAULT nextval('public.tags_id_seq'::regclass);


--
-- Name: tags_news id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tags_news ALTER COLUMN id SET DEFAULT nextval('public.tags_news_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Data for Name: folders; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.folders (id, userlog, name, newsid) FROM stdin;
1	regular_user        	Избранное           	1
2	regular_user        	Для чтения          	2
3	verified_user       	Мои статьи          	3
4	verified_user       	Интересное          	1
5	admin               	Админские           	3
\.


--
-- Data for Name: news; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.news (id, type_news, title, content, tag, source, date) FROM stdin;
1	t	Новые технологии в IT	Содержание статьи о новых технологиях...	2	2	2025-04-05 15:14:55.008649
2	t	Выборы 2023	Последние новости о выборах...	1	1	2025-04-05 15:14:55.008649
3	f	Корпоративные новости	Внутренние новости нашей организации...	3	4	2025-04-05 15:14:55.008649
4	t	Спортивные достижения	Новые рекорды в спорте...	4	3	2025-04-05 15:14:55.008649
5	t	Советы по здоровью	Как сохранить здоровье...	5	1	2025-04-05 15:14:55.008649
\.


--
-- Data for Name: offers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.offers (id, user_id, link) FROM stdin;
2	7	source_suggestion:https://наука.рф/news/dmitriy-chernyshenko-otkryl-filial-yuzhnogo-federalnogo-universiteta-v-gavane/
\.


--
-- Data for Name: sources; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sources (id, name, link) FROM stdin;
1	BBC News            	https://www.bbc.com
2	TechCrunch          	https://techcrunch.com
3	Reuters             	https://www.reuters.com
4	Local News          	https://local-news.example.com
\.


--
-- Data for Name: tags; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tags (id, name) FROM stdin;
1	Политика            
2	Технологии          
3	Экономика           
4	Спорт               
5	Здоровье            
\.


--
-- Data for Name: tags_news; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tags_news (id, tagid, newsid) FROM stdin;
1	2	1
2	3	1
3	1	2
4	3	2
5	3	3
6	4	4
7	5	5
\.


--
-- Data for Name: user_source_subscriptions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_source_subscriptions (user_id, source_id) FROM stdin;
2	2
3	1
3	3
4	1
4	2
4	3
\.


--
-- Data for Name: user_tag_subscriptions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_tag_subscriptions (user_id, tag_id) FROM stdin;
2	2
3	4
3	5
4	1
4	3
7	2
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, user_login, user_password, user_role, notification, tag_subscription, sources_subsc, telegram_id) FROM stdin;
7	mila                	$2a$06$WRGVLQgJlFixwI9qUErV5u0ULG3zAK/A4tDtHSeJZDFG1ZKBcptlK                                        	user           	0                   	\N	\N	1000000000                                                                                          
1	admin               	$2a$06$n1nEo8oPkP1ikm/M0aPiMeF2QilIsL/2cJe2cOUNPCq49you62LaS                                        	admin          	1                   	\N	\N	1000000000                                                                                          
4	moderator           	$2a$06$ie.5VYHvXoEdVCaggscI7.mvrn9mQ9NCQDmZKbl.JHNPlSihgzyle                                        	verified       	1                   	\N	\N	1000000000                                                                                          
3	regular_user        	$2a$06$wHtkV8uL1k0VpP3AuGC15.l5j7OCaGlXJhTn9zo8ShhZpu/R.lRUO                                        	user           	0                   	2	1	1000000000                                                                                          
6	test_user           	$2a$06$S4gQYZG06fDNiyDjFdBpD.QVJKISp/s8bQ40t9FoR7R.E6BbvpGnW                                        	user           	1                   	\N	\N	1000000000                                                                                          
2	verified_user       	$2a$06$HA1LVz/4aMB.5/7EGtFyV.FZkUrIuExC/ew925Rl7xEZ9syC5m2r2                                        	verified       	1                   	1	2	1000000000                                                                                          
\.


--
-- Name: folders_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.folders_id_seq', 5, true);


--
-- Name: news_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.news_id_seq', 5, true);


--
-- Name: offers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.offers_id_seq', 2, true);


--
-- Name: sources_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.sources_id_seq', 4, true);


--
-- Name: tags_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.tags_id_seq', 5, true);


--
-- Name: tags_news_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.tags_news_id_seq', 7, true);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_id_seq', 7, true);


--
-- Name: folders folders_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.folders
    ADD CONSTRAINT folders_pkey PRIMARY KEY (id);


--
-- Name: news news_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.news
    ADD CONSTRAINT news_pkey PRIMARY KEY (id);


--
-- Name: offers offers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.offers
    ADD CONSTRAINT offers_pkey PRIMARY KEY (id);


--
-- Name: sources sources_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sources
    ADD CONSTRAINT sources_pkey PRIMARY KEY (id);


--
-- Name: tags_news tags_news_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tags_news
    ADD CONSTRAINT tags_news_pkey PRIMARY KEY (id);


--
-- Name: tags_news tags_news_tagid_newsid_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tags_news
    ADD CONSTRAINT tags_news_tagid_newsid_key UNIQUE (tagid, newsid);


--
-- Name: tags tags_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (id);


--
-- Name: user_source_subscriptions user_source_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_source_subscriptions
    ADD CONSTRAINT user_source_subscriptions_pkey PRIMARY KEY (user_id, source_id);


--
-- Name: user_tag_subscriptions user_tag_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_tag_subscriptions
    ADD CONSTRAINT user_tag_subscriptions_pkey PRIMARY KEY (user_id, tag_id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_user_login_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_user_login_key UNIQUE (user_login);


--
-- Name: folders folders_newsid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.folders
    ADD CONSTRAINT folders_newsid_fkey FOREIGN KEY (newsid) REFERENCES public.news(id) ON DELETE SET NULL;


--
-- Name: folders folders_userlog_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.folders
    ADD CONSTRAINT folders_userlog_fkey FOREIGN KEY (userlog) REFERENCES public.users(user_login) ON DELETE CASCADE;


--
-- Name: news news_source_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.news
    ADD CONSTRAINT news_source_fkey FOREIGN KEY (source) REFERENCES public.sources(id) ON DELETE SET NULL;


--
-- Name: news news_tag_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.news
    ADD CONSTRAINT news_tag_fkey FOREIGN KEY (tag) REFERENCES public.tags(id) ON DELETE SET NULL;


--
-- Name: offers offers_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.offers
    ADD CONSTRAINT offers_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: tags_news tags_news_newsid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tags_news
    ADD CONSTRAINT tags_news_newsid_fkey FOREIGN KEY (newsid) REFERENCES public.news(id) ON DELETE CASCADE;


--
-- Name: tags_news tags_news_tagid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tags_news
    ADD CONSTRAINT tags_news_tagid_fkey FOREIGN KEY (tagid) REFERENCES public.tags(id) ON DELETE CASCADE;


--
-- Name: user_source_subscriptions user_source_subscriptions_source_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_source_subscriptions
    ADD CONSTRAINT user_source_subscriptions_source_id_fkey FOREIGN KEY (source_id) REFERENCES public.sources(id) ON DELETE CASCADE;


--
-- Name: user_source_subscriptions user_source_subscriptions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_source_subscriptions
    ADD CONSTRAINT user_source_subscriptions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_tag_subscriptions user_tag_subscriptions_tag_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_tag_subscriptions
    ADD CONSTRAINT user_tag_subscriptions_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES public.tags(id) ON DELETE CASCADE;


--
-- Name: user_tag_subscriptions user_tag_subscriptions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_tag_subscriptions
    ADD CONSTRAINT user_tag_subscriptions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

