-- Downloaded from: https://github.com/Vinicius02612/sistema_da_associacao/blob/119156f1a695cb012e6c9c5806ef1266adb48036/Backend/Arq_bd/BASE_DIR.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 16.4
-- Dumped by pg_dump version 16.4

-- Started on 2024-11-10 23:49:46

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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 217 (class 1259 OID 16444)
-- Name: administrador; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.administrador (
    id integer NOT NULL,
    nome character(30) NOT NULL
);


ALTER TABLE public.administrador OWNER TO postgres;

--
-- TOC entry 221 (class 1259 OID 16474)
-- Name: despesa; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.despesa (
    id integer NOT NULL,
    data date NOT NULL,
    valor double precision NOT NULL,
    origem character varying(100)
);


ALTER TABLE public.despesa OWNER TO postgres;

--
-- TOC entry 220 (class 1259 OID 16464)
-- Name: mensalidade; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.mensalidade (
    id integer NOT NULL,
    titulo character varying(100),
    datavalidade date,
    datareferencia date,
    valor double precision NOT NULL,
    status boolean NOT NULL,
    socio_id integer
);


ALTER TABLE public.mensalidade OWNER TO postgres;

--
-- TOC entry 219 (class 1259 OID 16459)
-- Name: projeto; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.projeto (
    id integer NOT NULL,
    titulo character varying(100) NOT NULL,
    datainicio date NOT NULL,
    datafim date,
    status character varying(50),
    nomeorganizacao character varying(100)
);


ALTER TABLE public.projeto OWNER TO postgres;

--
-- TOC entry 222 (class 1259 OID 16479)
-- Name: receita; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.receita (
    id integer NOT NULL,
    data date NOT NULL,
    valor double precision NOT NULL,
    origem character varying(100)
);


ALTER TABLE public.receita OWNER TO postgres;

--
-- TOC entry 223 (class 1259 OID 16484)
-- Name: relatorio; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.relatorio (
    id integer NOT NULL,
    titulo character varying(100) NOT NULL,
    datacriacao date,
    receita_id integer,
    despesa_id integer
);


ALTER TABLE public.relatorio OWNER TO postgres;

--
-- TOC entry 216 (class 1259 OID 16434)
-- Name: socio; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.socio (
    id integer NOT NULL,
    statuspagamento boolean,
    cargo character varying(100)
);


ALTER TABLE public.socio OWNER TO postgres;

--
-- TOC entry 218 (class 1259 OID 16454)
-- Name: solicitacao; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.solicitacao (
    id integer NOT NULL,
    status boolean NOT NULL
);


ALTER TABLE public.solicitacao OWNER TO postgres;

--
-- TOC entry 215 (class 1259 OID 16427)
-- Name: usuario; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.usuario (
    id integer NOT NULL,
    nome character varying(100) NOT NULL,
    cpf character varying(14) NOT NULL,
    email character varying(100) NOT NULL,
    senha character varying(100) NOT NULL,
    tipousuario character varying(50) NOT NULL,
    datanascimento date,
    quantidadepessoasfamilia integer
);


ALTER TABLE public.usuario OWNER TO postgres;

--
-- TOC entry 4835 (class 0 OID 16444)
-- Dependencies: 217
-- Data for Name: administrador; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.administrador (id, nome) FROM stdin;
\.


--
-- TOC entry 4839 (class 0 OID 16474)
-- Dependencies: 221
-- Data for Name: despesa; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.despesa (id, data, valor, origem) FROM stdin;
\.


--
-- TOC entry 4838 (class 0 OID 16464)
-- Dependencies: 220
-- Data for Name: mensalidade; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.mensalidade (id, titulo, datavalidade, datareferencia, valor, status, socio_id) FROM stdin;
\.


--
-- TOC entry 4837 (class 0 OID 16459)
-- Dependencies: 219
-- Data for Name: projeto; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.projeto (id, titulo, datainicio, datafim, status, nomeorganizacao) FROM stdin;
\.


--
-- TOC entry 4840 (class 0 OID 16479)
-- Dependencies: 222
-- Data for Name: receita; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.receita (id, data, valor, origem) FROM stdin;
\.


--
-- TOC entry 4841 (class 0 OID 16484)
-- Dependencies: 223
-- Data for Name: relatorio; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.relatorio (id, titulo, datacriacao, receita_id, despesa_id) FROM stdin;
\.


--
-- TOC entry 4834 (class 0 OID 16434)
-- Dependencies: 216
-- Data for Name: socio; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.socio (id, statuspagamento, cargo) FROM stdin;
\.


--
-- TOC entry 4836 (class 0 OID 16454)
-- Dependencies: 218
-- Data for Name: solicitacao; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.solicitacao (id, status) FROM stdin;
\.


--
-- TOC entry 4833 (class 0 OID 16427)
-- Dependencies: 215
-- Data for Name: usuario; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.usuario (id, nome, cpf, email, senha, tipousuario, datanascimento, quantidadepessoasfamilia) FROM stdin;
\.


--
-- TOC entry 4672 (class 2606 OID 16448)
-- Name: administrador administrador_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.administrador
    ADD CONSTRAINT administrador_pkey PRIMARY KEY (id);


--
-- TOC entry 4680 (class 2606 OID 16478)
-- Name: despesa despesa_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.despesa
    ADD CONSTRAINT despesa_pkey PRIMARY KEY (id);


--
-- TOC entry 4678 (class 2606 OID 16468)
-- Name: mensalidade mensalidade_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.mensalidade
    ADD CONSTRAINT mensalidade_pkey PRIMARY KEY (id);


--
-- TOC entry 4676 (class 2606 OID 16463)
-- Name: projeto projeto_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.projeto
    ADD CONSTRAINT projeto_pkey PRIMARY KEY (id);


--
-- TOC entry 4682 (class 2606 OID 16483)
-- Name: receita receita_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.receita
    ADD CONSTRAINT receita_pkey PRIMARY KEY (id);


--
-- TOC entry 4684 (class 2606 OID 16488)
-- Name: relatorio relatorio_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.relatorio
    ADD CONSTRAINT relatorio_pkey PRIMARY KEY (id);


--
-- TOC entry 4670 (class 2606 OID 16438)
-- Name: socio socio_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.socio
    ADD CONSTRAINT socio_pkey PRIMARY KEY (id);


--
-- TOC entry 4674 (class 2606 OID 16458)
-- Name: solicitacao solicitacao_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.solicitacao
    ADD CONSTRAINT solicitacao_pkey PRIMARY KEY (id);


--
-- TOC entry 4666 (class 2606 OID 16433)
-- Name: usuario usuario_cpf_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.usuario
    ADD CONSTRAINT usuario_cpf_key UNIQUE (cpf);


--
-- TOC entry 4668 (class 2606 OID 16431)
-- Name: usuario usuario_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.usuario
    ADD CONSTRAINT usuario_pkey PRIMARY KEY (id);


--
-- TOC entry 4686 (class 2606 OID 16449)
-- Name: administrador administrador_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.administrador
    ADD CONSTRAINT administrador_id_fkey FOREIGN KEY (id) REFERENCES public.usuario(id);


--
-- TOC entry 4687 (class 2606 OID 16469)
-- Name: mensalidade mensalidade_socio_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.mensalidade
    ADD CONSTRAINT mensalidade_socio_id_fkey FOREIGN KEY (socio_id) REFERENCES public.socio(id);


--
-- TOC entry 4688 (class 2606 OID 16494)
-- Name: relatorio relatorio_despesa_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.relatorio
    ADD CONSTRAINT relatorio_despesa_id_fkey FOREIGN KEY (despesa_id) REFERENCES public.despesa(id);


--
-- TOC entry 4689 (class 2606 OID 16489)
-- Name: relatorio relatorio_receita_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.relatorio
    ADD CONSTRAINT relatorio_receita_id_fkey FOREIGN KEY (receita_id) REFERENCES public.receita(id);


--
-- TOC entry 4685 (class 2606 OID 16439)
-- Name: socio socio_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.socio
    ADD CONSTRAINT socio_id_fkey FOREIGN KEY (id) REFERENCES public.usuario(id);


-- Completed on 2024-11-10 23:49:46

--
-- PostgreSQL database dump complete
--

