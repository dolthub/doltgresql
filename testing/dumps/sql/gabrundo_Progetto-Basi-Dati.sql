-- Downloaded from: https://github.com/gabrundo/Progetto-Basi-Dati/blob/f8270e116533e237f55d1ef56bcd7a0a0f94f17b/dump_uni.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 11.19
-- Dumped by pg_dump version 11.19

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
-- Name: uni; Type: SCHEMA; Schema: -; Owner: bitnami
--

CREATE SCHEMA uni;


ALTER SCHEMA uni OWNER TO bitnami;

--
-- Name: anno_insegnamento(); Type: FUNCTION; Schema: uni; Owner: bitnami
--

CREATE FUNCTION uni.anno_insegnamento() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
declare 
tipo corso_laurea.tipologia%type;
begin 
select tipologia into tipo
from corso_laurea
where nome = new.corso_laurea;

if tipo = 'Triennale' then
if not (new.anno = '1' or new.anno = '2' or new.anno = '3') then
raise exception 'Inserimento del insegnamento % non valido!', new.nome;
return null;
end if;
elsif tipo = 'Magistrale' then
if not (new.anno = '1' or new.anno = '2') then
raise exception 'Inserimento del insegnamento % non valido!', new.nome;
return null;
end if;
end if;
return new;
end;

$$;


ALTER FUNCTION uni.anno_insegnamento() OWNER TO bitnami;

--
-- Name: appelli_esami(); Type: FUNCTION; Schema: uni; Owner: bitnami
--

CREATE FUNCTION uni.appelli_esami() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
declare
anno_esame insegnamento.anno%type;
begin
select anno into anno_esame
from insegnamento
where corso_laurea = new.corso_laurea and codice = new.codice;

perform *
from insegnamento i inner join appello a on a.corso_laurea = i.corso_laurea and a.codice = i.codice
where i.corso_laurea = new.corso_laurea and i.anno = anno_esame and data = new.data;

if found then
raise exception 'Impossibile inserire appello per una sovrapposizione';
return null;
else
return new;
end if;
end;
$$;


ALTER FUNCTION uni.appelli_esami() OWNER TO bitnami;

--
-- Name: descrizione_corso_laurea(character varying); Type: FUNCTION; Schema: uni; Owner: bitnami
--

CREATE FUNCTION uni.descrizione_corso_laurea(character varying) RETURNS text
    LANGUAGE plpgsql
    AS $_$
declare
descrizione text = '';
corso corso_laurea%rowtype;
ins insegnamento%rowtype;
begin
select * into corso
from corso_laurea
where trim(lower(nome)) = trim(lower($1));

if not found then
raise exception 'Nome del corso non trovato';
else 
descrizione = descrizione || 'Nome corso di laurea: ' || corso.nome 
|| ', tipologia: ' || corso.tipologia || ', facoltà: ' || corso.segreteria || E'\n';
for ins in 
select *
from insegnamento
where trim(lower(nome)) = trim(lower($1))
order by anno asc
loop
descrizione = descrizione || 'nome esame: ' || ins.nome 
|| ', anno di erogazione : ' || ins.anno  || ', descrizione: ' || ins.descrizione
|| ', responsabile: ' || ins.responsabile || E'\n';
end loop; 
end if;
return descrizione;
end;
$_$;


ALTER FUNCTION uni.descrizione_corso_laurea(character varying) OWNER TO bitnami;

--
-- Name: iscrizione_esami(); Type: FUNCTION; Schema: uni; Owner: bitnami
--

CREATE FUNCTION uni.iscrizione_esami() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
declare 
cois propedeuticita.corso_is%type;
cdis propedeuticita.codice_is%type;
begin
perform *
from insegnamento 
where corso_laurea = new.corso_laurea and codice = new.codice;

if found then
for cois, cdis in
select corso_is, codice_is 
from propedeuticita
where corso_has = new.corso_laurea and codice_has = new.codice

loop 
if found then 
perform *
from sostiene
where corso_laurea = cois and codice = cdis 
and studente = new.studente and voto > 17 and data < new.data;

if not found then
raise exception 'Propedeuticità non rispettate per il corso %', cois;
return null;
end if;
end if;
end loop;
else
raise info 'Insegnamento % non registrato', cois;
return null;
end if;

return new;
end;
$$;


ALTER FUNCTION uni.iscrizione_esami() OWNER TO bitnami;

--
-- Name: numero_insegnamenti_docente(); Type: FUNCTION; Schema: uni; Owner: bitnami
--

CREATE FUNCTION uni.numero_insegnamenti_docente() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
declare
resp insegnamento.responsabile%type;
numero numeric;
begin
select responsabile, count(*) into resp, numero
from insegnamento
group by responsabile
having count(*) > 3;

if found then
raise exception 'Il docente, con identificativo %, non può gestire altri corsi!', resp;
delete from insegnamento where corso_laurea = new.corso_laurea and codice = new.codice;
end if;
return null;
end;
$$;


ALTER FUNCTION uni.numero_insegnamenti_docente() OWNER TO bitnami;

--
-- Name: str_studente_carriera(); Type: FUNCTION; Schema: uni; Owner: bitnami
--

CREATE FUNCTION uni.str_studente_carriera() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
    begin
        insert into str_studente (matricola, email, nome, cognome, corso_laurea) 
        values (old.matricola, old.email, old.nome, old.cognome, old.corso_laurea);

        insert into str_sostiene 
        select *
        from sostiene
        where studente = old.matricola;

delete
from sostiene
where studente=old.matricola;

        return old;
    end;
$$;


ALTER FUNCTION uni.str_studente_carriera() OWNER TO bitnami;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: appello; Type: TABLE; Schema: uni; Owner: bitnami
--

CREATE TABLE uni.appello (
    corso_laurea character varying NOT NULL,
    codice character(3) NOT NULL,
    data date NOT NULL
);


ALTER TABLE uni.appello OWNER TO bitnami;

--
-- Name: insegnamento; Type: TABLE; Schema: uni; Owner: bitnami
--

CREATE TABLE uni.insegnamento (
    corso_laurea character varying NOT NULL,
    codice character(3) NOT NULL,
    nome character varying NOT NULL,
    anno character(1) NOT NULL,
    descrizione text NOT NULL,
    responsabile character varying
);


ALTER TABLE uni.insegnamento OWNER TO bitnami;

--
-- Name: sostiene; Type: TABLE; Schema: uni; Owner: bitnami
--

CREATE TABLE uni.sostiene (
    studente character(6) NOT NULL,
    corso_laurea character varying NOT NULL,
    codice character(3) NOT NULL,
    data date NOT NULL,
    voto smallint,
    CONSTRAINT voto_valido CHECK (((0 <= voto) AND (voto <= 30)))
);


ALTER TABLE uni.sostiene OWNER TO bitnami;

--
-- Name: carriera_completa; Type: VIEW; Schema: uni; Owner: bitnami
--

CREATE VIEW uni.carriera_completa AS
 SELECT s.studente,
    i.nome,
    a.corso_laurea,
    i.anno,
    a.data,
    s.voto
   FROM ((uni.appello a
     JOIN uni.sostiene s ON ((((a.corso_laurea)::text = (s.corso_laurea)::text) AND (a.codice = s.codice) AND (a.data = s.data))))
     JOIN uni.insegnamento i ON ((((a.corso_laurea)::text = (i.corso_laurea)::text) AND (a.codice = i.codice))));


ALTER TABLE uni.carriera_completa OWNER TO bitnami;

--
-- Name: str_sostiene; Type: TABLE; Schema: uni; Owner: bitnami
--

CREATE TABLE uni.str_sostiene (
    studente character(6) NOT NULL,
    corso_laurea character varying NOT NULL,
    codice character(3) NOT NULL,
    data date NOT NULL,
    voto smallint
);


ALTER TABLE uni.str_sostiene OWNER TO bitnami;

--
-- Name: carriera_completa_globale; Type: VIEW; Schema: uni; Owner: bitnami
--

CREATE VIEW uni.carriera_completa_globale AS
 WITH sostiene_globale AS (
         SELECT sostiene.studente,
            sostiene.corso_laurea,
            sostiene.codice,
            sostiene.data,
            sostiene.voto
           FROM uni.sostiene
        UNION
         SELECT str_sostiene.studente,
            str_sostiene.corso_laurea,
            str_sostiene.codice,
            str_sostiene.data,
            str_sostiene.voto
           FROM uni.str_sostiene
        )
 SELECT s.studente,
    i.nome,
    a.corso_laurea,
    i.anno,
    a.data,
    s.voto
   FROM ((uni.appello a
     JOIN sostiene_globale s ON ((((a.corso_laurea)::text = (s.corso_laurea)::text) AND (a.codice = s.codice) AND (a.data = s.data))))
     JOIN uni.insegnamento i ON ((((a.corso_laurea)::text = (i.corso_laurea)::text) AND (a.codice = i.codice))));


ALTER TABLE uni.carriera_completa_globale OWNER TO bitnami;

--
-- Name: carriera_valida; Type: VIEW; Schema: uni; Owner: bitnami
--

CREATE VIEW uni.carriera_valida AS
 WITH esami_recenti AS (
         SELECT carriera_completa.studente,
            carriera_completa.nome,
            carriera_completa.corso_laurea,
            carriera_completa.anno,
            max(carriera_completa.data) AS data_recente
           FROM uni.carriera_completa
          WHERE (carriera_completa.voto > 17)
          GROUP BY carriera_completa.studente, carriera_completa.nome, carriera_completa.corso_laurea, carriera_completa.anno
        )
 SELECT e.studente,
    e.nome,
    e.corso_laurea,
    e.anno,
    e.data_recente,
    c.voto
   FROM (esami_recenti e
     JOIN uni.carriera_completa c ON (((e.studente = c.studente) AND ((e.nome)::text = (c.nome)::text) AND ((e.corso_laurea)::text = (c.corso_laurea)::text) AND (e.data_recente = c.data))));


ALTER TABLE uni.carriera_valida OWNER TO bitnami;

--
-- Name: carriera_valida_globale; Type: VIEW; Schema: uni; Owner: bitnami
--

CREATE VIEW uni.carriera_valida_globale AS
 WITH esami_recenti_globali AS (
         SELECT carriera_completa_globale.studente,
            carriera_completa_globale.nome,
            carriera_completa_globale.corso_laurea,
            carriera_completa_globale.anno,
            max(carriera_completa_globale.data) AS data_recente
           FROM uni.carriera_completa_globale
          WHERE (carriera_completa_globale.voto > 17)
          GROUP BY carriera_completa_globale.studente, carriera_completa_globale.nome, carriera_completa_globale.corso_laurea, carriera_completa_globale.anno
        )
 SELECT e.studente,
    e.nome,
    e.corso_laurea,
    e.anno,
    e.data_recente,
    c.voto
   FROM (esami_recenti_globali e
     JOIN uni.carriera_completa_globale c ON (((e.studente = c.studente) AND ((e.nome)::text = (c.nome)::text) AND ((e.corso_laurea)::text = (c.corso_laurea)::text) AND (e.data_recente = c.data))));


ALTER TABLE uni.carriera_valida_globale OWNER TO bitnami;

--
-- Name: corso_laurea; Type: TABLE; Schema: uni; Owner: bitnami
--

CREATE TABLE uni.corso_laurea (
    nome character varying NOT NULL,
    tipologia character(10) NOT NULL,
    segreteria character varying
);


ALTER TABLE uni.corso_laurea OWNER TO bitnami;

--
-- Name: docente; Type: TABLE; Schema: uni; Owner: bitnami
--

CREATE TABLE uni.docente (
    email character varying NOT NULL,
    password character varying NOT NULL,
    nome character varying(50) NOT NULL,
    cognome character varying(50) NOT NULL
);


ALTER TABLE uni.docente OWNER TO bitnami;

--
-- Name: propedeuticita; Type: TABLE; Schema: uni; Owner: bitnami
--

CREATE TABLE uni.propedeuticita (
    corso_is character varying NOT NULL,
    codice_is character(3) NOT NULL,
    corso_has character varying NOT NULL,
    codice_has character(3) NOT NULL
);


ALTER TABLE uni.propedeuticita OWNER TO bitnami;

--
-- Name: segretario; Type: TABLE; Schema: uni; Owner: bitnami
--

CREATE TABLE uni.segretario (
    email character varying NOT NULL,
    password character varying NOT NULL,
    nome character varying(50) NOT NULL,
    cognome character varying(50) NOT NULL,
    segreteria character varying
);


ALTER TABLE uni.segretario OWNER TO bitnami;

--
-- Name: segreteria; Type: TABLE; Schema: uni; Owner: bitnami
--

CREATE TABLE uni.segreteria (
    indirizzo character varying NOT NULL
);


ALTER TABLE uni.segreteria OWNER TO bitnami;

--
-- Name: str_studente; Type: TABLE; Schema: uni; Owner: bitnami
--

CREATE TABLE uni.str_studente (
    matricola character(6) NOT NULL,
    email character varying NOT NULL,
    nome character varying(50) NOT NULL,
    cognome character varying(50) NOT NULL,
    corso_laurea character varying
);


ALTER TABLE uni.str_studente OWNER TO bitnami;

--
-- Name: studente; Type: TABLE; Schema: uni; Owner: bitnami
--

CREATE TABLE uni.studente (
    matricola character(6) NOT NULL,
    email character varying NOT NULL,
    password character varying NOT NULL,
    nome character varying(50) NOT NULL,
    cognome character varying(50) NOT NULL,
    corso_laurea character varying
);


ALTER TABLE uni.studente OWNER TO bitnami;

--
-- Data for Name: appello; Type: TABLE DATA; Schema: uni; Owner: bitnami
--

COPY uni.appello (corso_laurea, codice, data) FROM stdin;
Informatica	001	2023-06-15
Informatica	001	2023-06-30
Informatica	002	2023-07-10
Informatica	004	2023-07-10
Informatica	004	2023-07-24
Informatica	003	2023-07-17
\.


--
-- Data for Name: corso_laurea; Type: TABLE DATA; Schema: uni; Owner: bitnami
--

COPY uni.corso_laurea (nome, tipologia, segreteria) FROM stdin;
Informatica	Triennale 	Scienze e Tecnologie
Matematica	Magistrale	Scienze e Tecnologie
\.


--
-- Data for Name: docente; Type: TABLE DATA; Schema: uni; Owner: bitnami
--

COPY uni.docente (email, password, nome, cognome) FROM stdin;
luca.bianchi@example.com	6a4dc9133d5f3b6d9fff778aff361961	Luca	Bianchi
giulia.verdi@example.com	6a4dc9133d5f3b6d9fff778aff361961	Giulia	Verdi
mario.rossi@example.com	6a4dc9133d5f3b6d9fff778aff361961	Mario	Rossi
giovanni.russo@example.com	6a4dc9133d5f3b6d9fff778aff361961	Giovanni	Russo
\.


--
-- Data for Name: insegnamento; Type: TABLE DATA; Schema: uni; Owner: bitnami
--

COPY uni.insegnamento (corso_laurea, codice, nome, anno, descrizione, responsabile) FROM stdin;
Informatica	001	Programmazione	1	Corso introduttivo alla programmazione	mario.rossi@example.com
Informatica	002	Basi di Dati	2	Intorduzione alle basi di dati	luca.bianchi@example.com
Matematica	001	Analisi Matematica	1	Corso di analisi matematica avanzata	giulia.verdi@example.com
Informatica	004	Logica Matematica	1	Introduzione alla logica matematica	giulia.verdi@example.com
Informatica	003	Matematica del continuo	1	Corso introduttivo di analisi matematica	giulia.verdi@example.com
Informatica	005	Architettura degli Elaboratori	1	Corso introduttivo alla architettura degli elaboratori	giovanni.russo@example.com
\.


--
-- Data for Name: propedeuticita; Type: TABLE DATA; Schema: uni; Owner: bitnami
--

COPY uni.propedeuticita (corso_is, codice_is, corso_has, codice_has) FROM stdin;
Informatica	001	Informatica	002
\.


--
-- Data for Name: segretario; Type: TABLE DATA; Schema: uni; Owner: bitnami
--

COPY uni.segretario (email, password, nome, cognome, segreteria) FROM stdin;
laura.rosa@example.com	6a4dc9133d5f3b6d9fff778aff361961	Laura	Rosa	Studi Umanistici
paolo.neri@example.com	6a4dc9133d5f3b6d9fff778aff361961	Paolo	Neri	Scienze e Tecnologie
\.


--
-- Data for Name: segreteria; Type: TABLE DATA; Schema: uni; Owner: bitnami
--

COPY uni.segreteria (indirizzo) FROM stdin;
Scienze e Tecnologie
Studi Umanistici
\.


--
-- Data for Name: sostiene; Type: TABLE DATA; Schema: uni; Owner: bitnami
--

COPY uni.sostiene (studente, corso_laurea, codice, data, voto) FROM stdin;
789012	Informatica	001	2023-06-30	25
789012	Informatica	002	2023-07-10	27
123456	Informatica	001	2023-06-15	28
123456	Informatica	001	2023-06-30	30
123456	Informatica	002	2023-07-10	\N
123456	Informatica	004	2023-07-10	20
123456	Informatica	003	2023-07-17	27
\.


--
-- Data for Name: str_sostiene; Type: TABLE DATA; Schema: uni; Owner: bitnami
--

COPY uni.str_sostiene (studente, corso_laurea, codice, data, voto) FROM stdin;
\.


--
-- Data for Name: str_studente; Type: TABLE DATA; Schema: uni; Owner: bitnami
--

COPY uni.str_studente (matricola, email, nome, cognome, corso_laurea) FROM stdin;
345678	sara.rossi@example.com	Sara	Rossi	Matematica
\.


--
-- Data for Name: studente; Type: TABLE DATA; Schema: uni; Owner: bitnami
--

COPY uni.studente (matricola, email, password, nome, cognome, corso_laurea) FROM stdin;
789012	marco.bianchi@example.com	6a4dc9133d5f3b6d9fff778aff361961	Marco	Bianchi	Informatica
123456	giuseppe.verdi@example.com	6a4dc9133d5f3b6d9fff778aff361961	Giuseppe	Verdi	Informatica
234567	luca.bianchi@example.com	6a4dc9133d5f3b6d9fff778aff361961	Luca	Bianchi	Informatica
\.


--
-- Name: appello appello_pkey; Type: CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.appello
    ADD CONSTRAINT appello_pkey PRIMARY KEY (corso_laurea, codice, data);


--
-- Name: corso_laurea corso_laurea_pkey; Type: CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.corso_laurea
    ADD CONSTRAINT corso_laurea_pkey PRIMARY KEY (nome);


--
-- Name: docente docente_pkey; Type: CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.docente
    ADD CONSTRAINT docente_pkey PRIMARY KEY (email);


--
-- Name: insegnamento insegnamento_pkey; Type: CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.insegnamento
    ADD CONSTRAINT insegnamento_pkey PRIMARY KEY (corso_laurea, codice);


--
-- Name: propedeuticita propedeuticita_pkey; Type: CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.propedeuticita
    ADD CONSTRAINT propedeuticita_pkey PRIMARY KEY (codice_is, corso_is, corso_has, codice_has);


--
-- Name: segretario segretario_pkey; Type: CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.segretario
    ADD CONSTRAINT segretario_pkey PRIMARY KEY (email);


--
-- Name: segreteria segreteria_pkey; Type: CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.segreteria
    ADD CONSTRAINT segreteria_pkey PRIMARY KEY (indirizzo);


--
-- Name: sostiene sostiene_pkey; Type: CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.sostiene
    ADD CONSTRAINT sostiene_pkey PRIMARY KEY (studente, corso_laurea, codice, data);


--
-- Name: str_sostiene str_sostiene_pkey; Type: CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.str_sostiene
    ADD CONSTRAINT str_sostiene_pkey PRIMARY KEY (studente, corso_laurea, codice, data);


--
-- Name: str_studente str_studente_email_key; Type: CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.str_studente
    ADD CONSTRAINT str_studente_email_key UNIQUE (email);


--
-- Name: str_studente str_studente_pkey; Type: CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.str_studente
    ADD CONSTRAINT str_studente_pkey PRIMARY KEY (matricola);


--
-- Name: studente studente_email_key; Type: CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.studente
    ADD CONSTRAINT studente_email_key UNIQUE (email);


--
-- Name: studente studente_pkey; Type: CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.studente
    ADD CONSTRAINT studente_pkey PRIMARY KEY (matricola);


--
-- Name: insegnamento gestione_anni_insegnamento; Type: TRIGGER; Schema: uni; Owner: bitnami
--

CREATE TRIGGER gestione_anni_insegnamento BEFORE INSERT ON uni.insegnamento FOR EACH ROW EXECUTE PROCEDURE uni.anno_insegnamento();


--
-- Name: appello gestione_appelli_esami; Type: TRIGGER; Schema: uni; Owner: bitnami
--

CREATE TRIGGER gestione_appelli_esami BEFORE INSERT ON uni.appello FOR EACH ROW EXECUTE PROCEDURE uni.appelli_esami();


--
-- Name: sostiene gestione_iscrizione_esami; Type: TRIGGER; Schema: uni; Owner: bitnami
--

CREATE TRIGGER gestione_iscrizione_esami BEFORE INSERT ON uni.sostiene FOR EACH ROW EXECUTE PROCEDURE uni.iscrizione_esami();


--
-- Name: insegnamento gestione_numero_insegnamenti_docente; Type: TRIGGER; Schema: uni; Owner: bitnami
--

CREATE TRIGGER gestione_numero_insegnamenti_docente AFTER INSERT ON uni.insegnamento FOR EACH ROW EXECUTE PROCEDURE uni.numero_insegnamenti_docente();


--
-- Name: studente storico_studente_carriera; Type: TRIGGER; Schema: uni; Owner: bitnami
--

CREATE TRIGGER storico_studente_carriera BEFORE DELETE ON uni.studente FOR EACH ROW EXECUTE PROCEDURE uni.str_studente_carriera();


--
-- Name: appello appello_insegnamento_fkey; Type: FK CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.appello
    ADD CONSTRAINT appello_insegnamento_fkey FOREIGN KEY (corso_laurea, codice) REFERENCES uni.insegnamento(corso_laurea, codice);


--
-- Name: corso_laurea corso_laurea_segreteria_fkey; Type: FK CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.corso_laurea
    ADD CONSTRAINT corso_laurea_segreteria_fkey FOREIGN KEY (segreteria) REFERENCES uni.segreteria(indirizzo) ON UPDATE CASCADE;


--
-- Name: propedeuticita has_prop_insegnamento_fkey; Type: FK CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.propedeuticita
    ADD CONSTRAINT has_prop_insegnamento_fkey FOREIGN KEY (corso_has, codice_has) REFERENCES uni.insegnamento(corso_laurea, codice);


--
-- Name: insegnamento insegnamento_responsabile_fkey; Type: FK CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.insegnamento
    ADD CONSTRAINT insegnamento_responsabile_fkey FOREIGN KEY (responsabile) REFERENCES uni.docente(email) ON UPDATE CASCADE;


--
-- Name: propedeuticita is_prop_insegnamento_fkey; Type: FK CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.propedeuticita
    ADD CONSTRAINT is_prop_insegnamento_fkey FOREIGN KEY (corso_is, codice_is) REFERENCES uni.insegnamento(corso_laurea, codice);


--
-- Name: segretario segretario_segreteria_fkey; Type: FK CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.segretario
    ADD CONSTRAINT segretario_segreteria_fkey FOREIGN KEY (segreteria) REFERENCES uni.segreteria(indirizzo) ON UPDATE CASCADE;


--
-- Name: sostiene sostiene_appello_fkey; Type: FK CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.sostiene
    ADD CONSTRAINT sostiene_appello_fkey FOREIGN KEY (corso_laurea, codice, data) REFERENCES uni.appello(corso_laurea, codice, data);


--
-- Name: studente studente_corso_laurea_fkey; Type: FK CONSTRAINT; Schema: uni; Owner: bitnami
--

ALTER TABLE ONLY uni.studente
    ADD CONSTRAINT studente_corso_laurea_fkey FOREIGN KEY (corso_laurea) REFERENCES uni.corso_laurea(nome) ON UPDATE CASCADE;


--
-- PostgreSQL database dump complete
--

