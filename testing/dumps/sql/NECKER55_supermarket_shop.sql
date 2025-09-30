-- Downloaded from: https://github.com/NECKER55/supermarket_shop/blob/f091e7fe8b35fb201ec7e91c4d989ef5e456b0e0/dump_29669A.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 13.11 (Debian 13.11-1.pgdg110+1)
-- Dumped by pg_dump version 15.13 (Debian 15.13-0+deb12u1)

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

ALTER TABLE IF EXISTS ONLY "società".tessera DROP CONSTRAINT IF EXISTS tessera_negozio_fkey;
ALTER TABLE IF EXISTS ONLY "società".tessera DROP CONSTRAINT IF EXISTS tessera_cf_cliente_fkey;
ALTER TABLE IF EXISTS ONLY "società".storicotessere DROP CONSTRAINT IF EXISTS storicotessere_cf_cliente_fkey;
ALTER TABLE IF EXISTS ONLY "società".ordinecontiene DROP CONSTRAINT IF EXISTS ordinecontiene_codice_prodotto_fkey;
ALTER TABLE IF EXISTS ONLY "società".ordinecontiene DROP CONSTRAINT IF EXISTS ordinecontiene_codice_ordine_fkey;
ALTER TABLE IF EXISTS ONLY "società".ordine DROP CONSTRAINT IF EXISTS ordine_codice_negozio_fkey;
ALTER TABLE IF EXISTS ONLY "società".ordine DROP CONSTRAINT IF EXISTS ordine_codice_fornitore_fkey;
ALTER TABLE IF EXISTS ONLY "società".negoziopossiede DROP CONSTRAINT IF EXISTS negoziopossiede_codice_prodotto_fkey;
ALTER TABLE IF EXISTS ONLY "società".negoziopossiede DROP CONSTRAINT IF EXISTS negoziopossiede_codice_negozio_fkey;
ALTER TABLE IF EXISTS ONLY "società".negozio DROP CONSTRAINT IF EXISTS negozio_cf_responsabile_fkey;
ALTER TABLE IF EXISTS ONLY "società".fornitorepossiede DROP CONSTRAINT IF EXISTS fornitorepossiede_codice_prodotto_fkey;
ALTER TABLE IF EXISTS ONLY "società".fornitorepossiede DROP CONSTRAINT IF EXISTS fornitorepossiede_codice_fornitore_fkey;
ALTER TABLE IF EXISTS ONLY "società".fatturacontiene DROP CONSTRAINT IF EXISTS fatturacontiene_codice_prodotto_fkey;
ALTER TABLE IF EXISTS ONLY "società".fatturacontiene DROP CONSTRAINT IF EXISTS fatturacontiene_codice_fattura_fkey;
ALTER TABLE IF EXISTS ONLY "società".fattura DROP CONSTRAINT IF EXISTS fattura_codice_negozio_fkey;
ALTER TABLE IF EXISTS ONLY "società".fattura DROP CONSTRAINT IF EXISTS fattura_cf_cliente_fkey;
ALTER TABLE IF EXISTS ONLY "società".credenziali DROP CONSTRAINT IF EXISTS credenziali_cf_persona_fkey;
DROP TRIGGER IF EXISTS trigger_utenti_piu_300_punti ON "società".tessera;
DROP TRIGGER IF EXISTS trigger_update_storico_tessere ON "società".storicotessere;
DROP TRIGGER IF EXISTS trigger_mantieni_storico_tessere ON "società".negozio;
DROP TRIGGER IF EXISTS trigger_controllo_responsabile ON "società".negozio;
DROP TRIGGER IF EXISTS trigger_applica_sconto_fattura ON "società".fattura;
DROP TRIGGER IF EXISTS trigger_aggiorna_totale_ordine ON "società".ordinecontiene;
DROP TRIGGER IF EXISTS trigger_aggiorna_totale_fattura ON "società".fatturacontiene;
DROP TRIGGER IF EXISTS trigger_aggiorna_scorte_negozio ON "società".fatturacontiene;
DROP TRIGGER IF EXISTS trigger_aggiorna_scorte_da_ordine ON "società".ordinecontiene;
DROP TRIGGER IF EXISTS trigger_aggiorna_punti_tessera ON "società".fattura;
DROP TRIGGER IF EXISTS trigger_aggiorna_disponibilita_fornitore ON "società".ordinecontiene;
DROP INDEX IF EXISTS "società".idx_tessera_punti;
DROP INDEX IF EXISTS "società".idx_ordine_negozio;
DROP INDEX IF EXISTS "società".idx_ordine_data_consegna;
DROP INDEX IF EXISTS "società".idx_negozio_possiede_quantita;
DROP INDEX IF EXISTS "società".idx_fattura_negozio;
DROP INDEX IF EXISTS "società".idx_fattura_data_acquisto;
DROP INDEX IF EXISTS "società".idx_fattura_cf_cliente;
ALTER TABLE IF EXISTS ONLY "società".tessera DROP CONSTRAINT IF EXISTS tessera_pkey;
ALTER TABLE IF EXISTS ONLY "società".storicotessere DROP CONSTRAINT IF EXISTS storicotessere_pkey;
ALTER TABLE IF EXISTS ONLY "società".prodotto DROP CONSTRAINT IF EXISTS prodotto_pkey;
ALTER TABLE IF EXISTS ONLY "società".persona DROP CONSTRAINT IF EXISTS persona_pkey;
ALTER TABLE IF EXISTS ONLY "società".ordinecontiene DROP CONSTRAINT IF EXISTS ordinecontiene_pkey;
ALTER TABLE IF EXISTS ONLY "società".ordine DROP CONSTRAINT IF EXISTS ordine_pkey;
ALTER TABLE IF EXISTS ONLY "società".negoziopossiede DROP CONSTRAINT IF EXISTS negoziopossiede_pkey;
ALTER TABLE IF EXISTS ONLY "società".negozio DROP CONSTRAINT IF EXISTS negozio_pkey;
ALTER TABLE IF EXISTS ONLY "società".fornitorepossiede DROP CONSTRAINT IF EXISTS fornitorepossiede_pkey;
ALTER TABLE IF EXISTS ONLY "società".fornitore DROP CONSTRAINT IF EXISTS fornitore_pkey;
ALTER TABLE IF EXISTS ONLY "società".fatturacontiene DROP CONSTRAINT IF EXISTS fatturacontiene_pkey;
ALTER TABLE IF EXISTS ONLY "società".fattura DROP CONSTRAINT IF EXISTS fattura_pkey;
ALTER TABLE IF EXISTS ONLY "società".credenziali DROP CONSTRAINT IF EXISTS credenziali_pkey;
ALTER TABLE IF EXISTS ONLY "società".credenziali DROP CONSTRAINT IF EXISTS credenziali_cf_persona_manager_key;
ALTER TABLE IF EXISTS "società".prodotto ALTER COLUMN codice DROP DEFAULT;
ALTER TABLE IF EXISTS "società".ordine ALTER COLUMN codice DROP DEFAULT;
ALTER TABLE IF EXISTS "società".negozio ALTER COLUMN codice DROP DEFAULT;
ALTER TABLE IF EXISTS "società".fattura ALTER COLUMN codice DROP DEFAULT;
DROP SEQUENCE IF EXISTS "società".prodotto_codice_seq;
DROP TABLE IF EXISTS "società".prodotto;
DROP TABLE IF EXISTS "società".ordinecontiene;
DROP SEQUENCE IF EXISTS "società".ordine_codice_seq;
DROP TABLE IF EXISTS "società".negoziopossiede;
DROP SEQUENCE IF EXISTS "società".negozio_codice_seq;
DROP TABLE IF EXISTS "società".negozio;
DROP MATERIALIZED VIEW IF EXISTS "società".materialized_view_utenti_piu_300_punti;
DROP TABLE IF EXISTS "società".tessera;
DROP MATERIALIZED VIEW IF EXISTS "società".materialized_view_storico_tessere;
DROP TABLE IF EXISTS "società".storicotessere;
DROP TABLE IF EXISTS "società".persona;
DROP TABLE IF EXISTS "società".fornitorepossiede;
DROP TABLE IF EXISTS "società".fornitore;
DROP TABLE IF EXISTS "società".fatturacontiene;
DROP SEQUENCE IF EXISTS "società".fattura_codice_seq;
DROP TABLE IF EXISTS "società".fattura;
DROP TABLE IF EXISTS "società".credenziali;
DROP FUNCTION IF EXISTS "società".verifica_sconto_disponibile(cf_cliente_param character);
DROP FUNCTION IF EXISTS "società".update_utenti_piu_300_punti();
DROP FUNCTION IF EXISTS "società".update_storico_tessere();
DROP FUNCTION IF EXISTS "società".trova_fornitore_economico(cod_prodotto integer, quantita_richiesta integer);
DROP FUNCTION IF EXISTS "società".punti_necessari_sconto(percentuale_sconto integer);
DROP FUNCTION IF EXISTS "società".mantieni_storico_tessere();
DROP FUNCTION IF EXISTS "società".get_storico_cliente(cf_cliente_input character);
DROP FUNCTION IF EXISTS "società".get_prodotti_negozio(cod_negozio integer);
DROP FUNCTION IF EXISTS "società".get_ordini_fornitore(fornitore character);
DROP TABLE IF EXISTS "società".ordine;
DROP FUNCTION IF EXISTS "società".controllo_inserimento_responsabile();
DROP FUNCTION IF EXISTS "società".applica_sconto_fattura();
DROP FUNCTION IF EXISTS "società".aggiorna_totale_ordine();
DROP FUNCTION IF EXISTS "società".aggiorna_totale_fattura();
DROP FUNCTION IF EXISTS "società".aggiorna_scorte_negozio();
DROP FUNCTION IF EXISTS "società".aggiorna_scorte_da_ordine();
DROP FUNCTION IF EXISTS "società".aggiorna_punti_tessera();
DROP FUNCTION IF EXISTS "società".aggiorna_disponibilita_fornitore();
DROP SCHEMA IF EXISTS "società";
-- *not* dropping schema, since initdb creates it
DROP SCHEMA IF EXISTS andrea_veneroni1;
--
-- Name: andrea_veneroni1; Type: SCHEMA; Schema: -; Owner: andrea_veneroni1
--

CREATE SCHEMA andrea_veneroni1;


ALTER SCHEMA andrea_veneroni1 OWNER TO andrea_veneroni1;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: postgres
--

-- *not* creating schema, since initdb creates it


ALTER SCHEMA public OWNER TO postgres;

--
-- Name: società; Type: SCHEMA; Schema: -; Owner: andrea_veneroni1
--

CREATE SCHEMA "società";


ALTER SCHEMA "società" OWNER TO andrea_veneroni1;

--
-- Name: aggiorna_disponibilita_fornitore(); Type: FUNCTION; Schema: società; Owner: andrea_veneroni1
--

CREATE FUNCTION "società".aggiorna_disponibilita_fornitore() RETURNS trigger
    LANGUAGE plpgsql
    AS $$

BEGIN

    UPDATE FornitorePossiede

    SET quantita = quantita - NEW.quantita

    WHERE codice_fornitore = (

        SELECT codice_fornitore 

        FROM Ordine 

        WHERE codice = NEW.codice_ordine

    )

    AND codice_prodotto = NEW.codice_prodotto;

    IF EXISTS (

        SELECT 1 FROM FornitorePossiede 

        WHERE codice_fornitore = (

            SELECT codice_fornitore 

            FROM Ordine 

            WHERE codice = NEW.codice_ordine

        )

        AND codice_prodotto = NEW.codice_prodotto

        AND quantita < 0

    ) THEN

        RAISE EXCEPTION 'Quantità insufficiente presso il fornitore per il prodotto %', NEW.codice_prodotto;

    END IF;

    RETURN NEW;

END;

$$;


ALTER FUNCTION "società".aggiorna_disponibilita_fornitore() OWNER TO andrea_veneroni1;

--
-- Name: aggiorna_punti_tessera(); Type: FUNCTION; Schema: società; Owner: andrea_veneroni1
--

CREATE FUNCTION "società".aggiorna_punti_tessera() RETURNS trigger
    LANGUAGE plpgsql
    AS $$

BEGIN

    -- Aggiunge punti (1 punto per ogni euro speso)

    UPDATE Tessera

    SET punti = punti + FLOOR(NEW.totale)

    WHERE cf_cliente = NEW.cf_cliente;

    RETURN NEW;

END;

$$;


ALTER FUNCTION "società".aggiorna_punti_tessera() OWNER TO andrea_veneroni1;

--
-- Name: aggiorna_scorte_da_ordine(); Type: FUNCTION; Schema: società; Owner: andrea_veneroni1
--

CREATE FUNCTION "società".aggiorna_scorte_da_ordine() RETURNS trigger
    LANGUAGE plpgsql
    AS $$

DECLARE

    prezzo_esistente DECIMAL(10,2);

    quantita_esistente INT;

BEGIN

    -- Verifica se il prodotto esiste già nel negozio

    SELECT prezzo, quantita

    INTO prezzo_esistente, quantita_esistente

    FROM NegozioPossiede

    WHERE codice_prodotto = NEW.codice_prodotto

      AND codice_negozio = (

          SELECT codice_negozio 

          FROM Ordine 

          WHERE codice = NEW.codice_ordine

      );

    IF FOUND THEN

        -- Se esiste, aggiorna solo la quantità

        UPDATE NegozioPossiede

        SET quantita = quantita + NEW.quantita

        WHERE codice_prodotto = NEW.codice_prodotto

          AND codice_negozio = (

              SELECT codice_negozio 

              FROM Ordine 

              WHERE codice = NEW.codice_ordine

          );

    ELSE

        -- Se non esiste, inserisce con markup del 30%

        INSERT INTO NegozioPossiede (codice_prodotto, codice_negozio, quantita, prezzo)

        SELECT 

            NEW.codice_prodotto,

            o.codice_negozio,

            NEW.quantita,

            NEW.prezzo * 1.3

        FROM Ordine o

        WHERE o.codice = NEW.codice_ordine;

    END IF;

    RETURN NEW;

END;

$$;


ALTER FUNCTION "società".aggiorna_scorte_da_ordine() OWNER TO andrea_veneroni1;

--
-- Name: aggiorna_scorte_negozio(); Type: FUNCTION; Schema: società; Owner: andrea_veneroni1
--

CREATE FUNCTION "società".aggiorna_scorte_negozio() RETURNS trigger
    LANGUAGE plpgsql
    AS $$

BEGIN

    UPDATE NegozioPossiede

    SET quantita = quantita - NEW.quantita

    WHERE codice_prodotto = NEW.codice_prodotto

      AND codice_negozio = (

          SELECT codice_negozio 

          FROM Fattura 

          WHERE codice = NEW.codice_fattura

      );

    IF EXISTS (

        SELECT 1 FROM NegozioPossiede 

        WHERE codice_prodotto = NEW.codice_prodotto

          AND codice_negozio = (

              SELECT codice_negozio 

              FROM Fattura 

              WHERE codice = NEW.codice_fattura

          )

          AND quantita < 0

    ) THEN

        RAISE EXCEPTION 'Quantità insufficiente in negozio per il prodotto %', NEW.codice_prodotto;

    END IF;

    RETURN NEW;

END;

$$;


ALTER FUNCTION "società".aggiorna_scorte_negozio() OWNER TO andrea_veneroni1;

--
-- Name: aggiorna_totale_fattura(); Type: FUNCTION; Schema: società; Owner: andrea_veneroni1
--

CREATE FUNCTION "società".aggiorna_totale_fattura() RETURNS trigger
    LANGUAGE plpgsql
    AS $$

BEGIN

    UPDATE Fattura

    SET totale = (

        SELECT COALESCE(SUM(quantita * prezzo), 0)

        FROM FatturaContiene

        WHERE codice_fattura = COALESCE(NEW.codice_fattura, OLD.codice_fattura)

    )

    WHERE codice = COALESCE(NEW.codice_fattura, OLD.codice_fattura);

    RETURN COALESCE(NEW, OLD);

END;

$$;


ALTER FUNCTION "società".aggiorna_totale_fattura() OWNER TO andrea_veneroni1;

--
-- Name: aggiorna_totale_ordine(); Type: FUNCTION; Schema: società; Owner: andrea_veneroni1
--

CREATE FUNCTION "società".aggiorna_totale_ordine() RETURNS trigger
    LANGUAGE plpgsql
    AS $$

BEGIN

    UPDATE Ordine

    SET totale = (

        SELECT COALESCE(SUM(quantita * prezzo), 0)

        FROM OrdineContiene

        WHERE codice_ordine = COALESCE(NEW.codice_ordine, OLD.codice_ordine)

    )

    WHERE codice = COALESCE(NEW.codice_ordine, OLD.codice_ordine);

    RETURN COALESCE(NEW, OLD);

END;

$$;


ALTER FUNCTION "società".aggiorna_totale_ordine() OWNER TO andrea_veneroni1;

--
-- Name: applica_sconto_fattura(); Type: FUNCTION; Schema: società; Owner: andrea_veneroni1
--

CREATE FUNCTION "società".applica_sconto_fattura() RETURNS trigger
    LANGUAGE plpgsql
    AS $$

DECLARE

    punti_attuali INT;

    punti_necessari INT;

    sconto_euro DECIMAL(10,2);

BEGIN

    -- Se è stato applicato uno sconto, decurta i punti

    IF NEW.sconto > 0 THEN

        SELECT punti INTO punti_attuali

        FROM Tessera

        WHERE cf_cliente = NEW.cf_cliente;

        punti_necessari := punti_necessari_sconto(NEW.sconto);

        -- Controlla che il cliente abbia abbastanza punti

        IF punti_attuali < punti_necessari THEN

            RAISE EXCEPTION 'Punti insufficienti per applicare lo sconto del %: Punti disponibili: %, Punti necessari: %', 

                NEW.sconto, punti_attuali, punti_necessari;

        END IF;

        -- Calcola lo sconto in euro (massimo 100 euro)

        sconto_euro := LEAST(NEW.totale * NEW.sconto / 100.0, 100.0);

        -- Aggiorna il totale della fattura

        NEW.totale := NEW.totale - sconto_euro;

        -- Decurta i punti dalla tessera

        UPDATE Tessera

        SET punti = punti - punti_necessari

        WHERE cf_cliente = NEW.cf_cliente;

    END IF;

    RETURN NEW;

END;

$$;


ALTER FUNCTION "società".applica_sconto_fattura() OWNER TO andrea_veneroni1;

--
-- Name: controllo_inserimento_responsabile(); Type: FUNCTION; Schema: società; Owner: andrea_veneroni1
--

CREATE FUNCTION "società".controllo_inserimento_responsabile() RETURNS trigger
    LANGUAGE plpgsql
    AS $$

DECLARE

    is_manager BOOLEAN;

BEGIN

    SELECT manager

    INTO is_manager

    FROM Credenziali

    WHERE cf_persona = NEW.cf_responsabile;

    IF NOT FOUND OR is_manager IS DISTINCT FROM TRUE THEN

        RAISE EXCEPTION 'Il codice fiscale % inserito come responsabile non corrisponde a un utente amministratore.', NEW.cf_responsabile;

    END IF;

    RETURN NEW;

END;

$$;


ALTER FUNCTION "società".controllo_inserimento_responsabile() OWNER TO andrea_veneroni1;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: ordine; Type: TABLE; Schema: società; Owner: andrea_veneroni1
--

CREATE TABLE "società".ordine (
    codice integer NOT NULL,
    totale numeric(10,2) DEFAULT 0 NOT NULL,
    data_consegna timestamp without time zone NOT NULL,
    codice_negozio integer NOT NULL,
    codice_fornitore character(11) NOT NULL,
    CONSTRAINT ordine_totale_check CHECK ((totale >= (0)::numeric))
);


ALTER TABLE "società".ordine OWNER TO andrea_veneroni1;

--
-- Name: get_ordini_fornitore(character); Type: FUNCTION; Schema: società; Owner: andrea_veneroni1
--

CREATE FUNCTION "società".get_ordini_fornitore(fornitore character) RETURNS SETOF "società".ordine
    LANGUAGE plpgsql
    AS $$

BEGIN

    RETURN QUERY 

    SELECT o.*

    FROM Ordine o

    WHERE o.codice_fornitore = fornitore;

END;

$$;


ALTER FUNCTION "società".get_ordini_fornitore(fornitore character) OWNER TO andrea_veneroni1;

--
-- Name: get_prodotti_negozio(integer); Type: FUNCTION; Schema: società; Owner: andrea_veneroni1
--

CREATE FUNCTION "società".get_prodotti_negozio(cod_negozio integer) RETURNS TABLE(codice integer, nome character varying, descrizione text, prezzo numeric, quantita integer)
    LANGUAGE plpgsql
    AS $$

BEGIN

    RETURN QUERY

    SELECT p.codice, p.nome, p.descrizione, np.prezzo, np.quantita

    FROM Prodotto p

    JOIN NegozioPossiede np ON p.codice = np.codice_prodotto

    WHERE np.codice_negozio = cod_negozio AND np.quantita > 0

    ORDER BY p.nome;

END;

$$;


ALTER FUNCTION "società".get_prodotti_negozio(cod_negozio integer) OWNER TO andrea_veneroni1;

--
-- Name: get_storico_cliente(character); Type: FUNCTION; Schema: società; Owner: andrea_veneroni1
--

CREATE FUNCTION "società".get_storico_cliente(cf_cliente_input character) RETURNS TABLE(codice_fattura integer, data_acquisto timestamp without time zone, totale numeric, sconto integer, negozio_indirizzo character varying)
    LANGUAGE plpgsql
    AS $$

BEGIN

    RETURN QUERY

    SELECT f.codice, f.data_acquisto, f.totale, f.sconto, n.indirizzo

    FROM Fattura f

    JOIN Negozio n ON f.codice_negozio = n.codice

    WHERE f.cf_cliente = cf_cliente_input

    ORDER BY f.data_acquisto DESC;

END;

$$;


ALTER FUNCTION "società".get_storico_cliente(cf_cliente_input character) OWNER TO andrea_veneroni1;

--
-- Name: mantieni_storico_tessere(); Type: FUNCTION; Schema: società; Owner: andrea_veneroni1
--

CREATE FUNCTION "società".mantieni_storico_tessere() RETURNS trigger
    LANGUAGE plpgsql
    AS $$

BEGIN

    INSERT INTO StoricoTessere (cf_cliente, codice_negozio_eliminato, punti_al_momento_eliminazione, data_richiesta)

    SELECT cf_cliente, OLD.codice, punti, data_richiesta

    FROM Tessera

    WHERE negozio = OLD.codice;

    RETURN OLD;

END;

$$;


ALTER FUNCTION "società".mantieni_storico_tessere() OWNER TO andrea_veneroni1;

--
-- Name: punti_necessari_sconto(integer); Type: FUNCTION; Schema: società; Owner: andrea_veneroni1
--

CREATE FUNCTION "società".punti_necessari_sconto(percentuale_sconto integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$

BEGIN

    CASE percentuale_sconto

        WHEN 5 THEN RETURN 100;

        WHEN 15 THEN RETURN 200;

        WHEN 30 THEN RETURN 300;

        ELSE RETURN 0;

    END CASE;

END;

$$;


ALTER FUNCTION "società".punti_necessari_sconto(percentuale_sconto integer) OWNER TO andrea_veneroni1;

--
-- Name: trova_fornitore_economico(integer, integer); Type: FUNCTION; Schema: società; Owner: andrea_veneroni1
--

CREATE FUNCTION "società".trova_fornitore_economico(cod_prodotto integer, quantita_richiesta integer) RETURNS character
    LANGUAGE plpgsql
    AS $$

DECLARE

    fornitore_scelto CHAR(11);

BEGIN

    SELECT codice_fornitore

    INTO fornitore_scelto

    FROM FornitorePossiede

    WHERE codice_prodotto = cod_prodotto 

      AND quantita >= quantita_richiesta

    ORDER BY prezzo ASC

    LIMIT 1;

    IF fornitore_scelto IS NULL THEN

        RAISE EXCEPTION 'Nessun fornitore disponibile per il prodotto % con quantità %', cod_prodotto, quantita_richiesta;

    END IF;

    RETURN fornitore_scelto;

END;

$$;


ALTER FUNCTION "società".trova_fornitore_economico(cod_prodotto integer, quantita_richiesta integer) OWNER TO andrea_veneroni1;

--
-- Name: update_storico_tessere(); Type: FUNCTION; Schema: società; Owner: andrea_veneroni1
--

CREATE FUNCTION "società".update_storico_tessere() RETURNS trigger
    LANGUAGE plpgsql
    AS $$

BEGIN

    REFRESH MATERIALIZED VIEW materialized_view_storico_tessere;

    RETURN NULL;

END;

$$;


ALTER FUNCTION "società".update_storico_tessere() OWNER TO andrea_veneroni1;

--
-- Name: update_utenti_piu_300_punti(); Type: FUNCTION; Schema: società; Owner: andrea_veneroni1
--

CREATE FUNCTION "società".update_utenti_piu_300_punti() RETURNS trigger
    LANGUAGE plpgsql
    AS $$

BEGIN

    IF (TG_OP = 'DELETE') THEN

        IF (OLD.punti > 300) THEN

            REFRESH MATERIALIZED VIEW materialized_view_utenti_piu_300_punti;

        END IF;

    ELSIF (TG_OP IN ('INSERT', 'UPDATE')) THEN

        IF ((NEW.punti > 300) OR (TG_OP = 'UPDATE' AND OLD.punti > 300)) THEN

            REFRESH MATERIALIZED VIEW materialized_view_utenti_piu_300_punti;

        END IF;

    END IF;

    RETURN NULL;

END;

$$;


ALTER FUNCTION "società".update_utenti_piu_300_punti() OWNER TO andrea_veneroni1;

--
-- Name: verifica_sconto_disponibile(character); Type: FUNCTION; Schema: società; Owner: andrea_veneroni1
--

CREATE FUNCTION "società".verifica_sconto_disponibile(cf_cliente_param character) RETURNS TABLE(sconto_5 boolean, sconto_15 boolean, sconto_30 boolean, punti_attuali integer)
    LANGUAGE plpgsql
    AS $$

BEGIN

    RETURN QUERY

    SELECT 

        t.punti >= 100 AS sconto_5,

        t.punti >= 200 AS sconto_15,

        t.punti >= 300 AS sconto_30,

        t.punti AS punti_attuali

    FROM Tessera t

    WHERE t.cf_cliente = cf_cliente_param;

END;

$$;


ALTER FUNCTION "società".verifica_sconto_disponibile(cf_cliente_param character) OWNER TO andrea_veneroni1;

--
-- Name: credenziali; Type: TABLE; Schema: società; Owner: andrea_veneroni1
--

CREATE TABLE "società".credenziali (
    username character varying(50) NOT NULL,
    password character varying(128) NOT NULL,
    cf_persona character(16) NOT NULL,
    manager boolean DEFAULT false NOT NULL
);


ALTER TABLE "società".credenziali OWNER TO andrea_veneroni1;

--
-- Name: fattura; Type: TABLE; Schema: società; Owner: andrea_veneroni1
--

CREATE TABLE "società".fattura (
    codice integer NOT NULL,
    totale numeric(10,2) DEFAULT 0 NOT NULL,
    sconto integer DEFAULT 0 NOT NULL,
    codice_negozio integer NOT NULL,
    cf_cliente character(16) NOT NULL,
    data_acquisto timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fattura_sconto_check CHECK ((sconto >= 0)),
    CONSTRAINT fattura_totale_check CHECK ((totale >= (0)::numeric))
);


ALTER TABLE "società".fattura OWNER TO andrea_veneroni1;

--
-- Name: fattura_codice_seq; Type: SEQUENCE; Schema: società; Owner: andrea_veneroni1
--

CREATE SEQUENCE "società".fattura_codice_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "società".fattura_codice_seq OWNER TO andrea_veneroni1;

--
-- Name: fattura_codice_seq; Type: SEQUENCE OWNED BY; Schema: società; Owner: andrea_veneroni1
--

ALTER SEQUENCE "società".fattura_codice_seq OWNED BY "società".fattura.codice;


--
-- Name: fatturacontiene; Type: TABLE; Schema: società; Owner: andrea_veneroni1
--

CREATE TABLE "società".fatturacontiene (
    codice_fattura integer NOT NULL,
    codice_prodotto integer NOT NULL,
    prezzo numeric(10,2) NOT NULL,
    quantita integer NOT NULL,
    CONSTRAINT fatturacontiene_prezzo_check CHECK ((prezzo > (0)::numeric)),
    CONSTRAINT fatturacontiene_quantita_check CHECK ((quantita > 0))
);


ALTER TABLE "società".fatturacontiene OWNER TO andrea_veneroni1;

--
-- Name: fornitore; Type: TABLE; Schema: società; Owner: andrea_veneroni1
--

CREATE TABLE "società".fornitore (
    p_iva character(11) NOT NULL,
    indirizzo character varying(100) NOT NULL
);


ALTER TABLE "società".fornitore OWNER TO andrea_veneroni1;

--
-- Name: fornitorepossiede; Type: TABLE; Schema: società; Owner: andrea_veneroni1
--

CREATE TABLE "società".fornitorepossiede (
    codice_fornitore character(11) NOT NULL,
    codice_prodotto integer NOT NULL,
    prezzo numeric(10,2) NOT NULL,
    quantita integer NOT NULL,
    CONSTRAINT fornitorepossiede_prezzo_check CHECK ((prezzo > (0)::numeric)),
    CONSTRAINT fornitorepossiede_quantita_check CHECK ((quantita >= 0))
);


ALTER TABLE "società".fornitorepossiede OWNER TO andrea_veneroni1;

--
-- Name: persona; Type: TABLE; Schema: società; Owner: andrea_veneroni1
--

CREATE TABLE "società".persona (
    cf character(16) NOT NULL,
    nome character varying(50) NOT NULL,
    cognome character varying(50) NOT NULL
);


ALTER TABLE "società".persona OWNER TO andrea_veneroni1;

--
-- Name: storicotessere; Type: TABLE; Schema: società; Owner: andrea_veneroni1
--

CREATE TABLE "società".storicotessere (
    cf_cliente character(16) NOT NULL,
    codice_negozio_eliminato integer NOT NULL,
    punti_al_momento_eliminazione integer NOT NULL,
    data_richiesta timestamp without time zone NOT NULL,
    data_eliminazione_negozio timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE "società".storicotessere OWNER TO andrea_veneroni1;

--
-- Name: materialized_view_storico_tessere; Type: MATERIALIZED VIEW; Schema: società; Owner: andrea_veneroni1
--

CREATE MATERIALIZED VIEW "società".materialized_view_storico_tessere AS
 SELECT st.cf_cliente,
    st.codice_negozio_eliminato,
    st.punti_al_momento_eliminazione,
    st.data_richiesta,
    st.data_eliminazione_negozio,
    p.nome,
    p.cognome
   FROM ("società".storicotessere st
     JOIN "società".persona p ON ((p.cf = st.cf_cliente)))
  ORDER BY st.data_eliminazione_negozio DESC
  WITH NO DATA;


ALTER TABLE "società".materialized_view_storico_tessere OWNER TO andrea_veneroni1;

--
-- Name: tessera; Type: TABLE; Schema: società; Owner: andrea_veneroni1
--

CREATE TABLE "società".tessera (
    cf_cliente character(16) NOT NULL,
    negozio integer,
    punti integer DEFAULT 0 NOT NULL,
    data_richiesta timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT tessera_punti_check CHECK ((punti >= 0))
);


ALTER TABLE "società".tessera OWNER TO andrea_veneroni1;

--
-- Name: materialized_view_utenti_piu_300_punti; Type: MATERIALIZED VIEW; Schema: società; Owner: andrea_veneroni1
--

CREATE MATERIALIZED VIEW "società".materialized_view_utenti_piu_300_punti AS
 SELECT p.cf AS persona_cf,
    p.nome AS persona_nome,
    p.cognome AS persona_cognome,
    t.cf_cliente AS tessera_cf_cliente,
    t.negozio AS tessera_negozio,
    t.punti AS tessera_punti,
    t.data_richiesta AS tessera_data_richiesta
   FROM ("società".persona p
     JOIN "società".tessera t ON ((t.cf_cliente = p.cf)))
  WHERE (t.punti > 300)
  WITH NO DATA;


ALTER TABLE "società".materialized_view_utenti_piu_300_punti OWNER TO andrea_veneroni1;

--
-- Name: negozio; Type: TABLE; Schema: società; Owner: andrea_veneroni1
--

CREATE TABLE "società".negozio (
    codice integer NOT NULL,
    indirizzo character varying(100) NOT NULL,
    cf_responsabile character(16) NOT NULL,
    orario_apertura time without time zone NOT NULL,
    orario_chiusura time without time zone NOT NULL
);


ALTER TABLE "società".negozio OWNER TO andrea_veneroni1;

--
-- Name: negozio_codice_seq; Type: SEQUENCE; Schema: società; Owner: andrea_veneroni1
--

CREATE SEQUENCE "società".negozio_codice_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "società".negozio_codice_seq OWNER TO andrea_veneroni1;

--
-- Name: negozio_codice_seq; Type: SEQUENCE OWNED BY; Schema: società; Owner: andrea_veneroni1
--

ALTER SEQUENCE "società".negozio_codice_seq OWNED BY "società".negozio.codice;


--
-- Name: negoziopossiede; Type: TABLE; Schema: società; Owner: andrea_veneroni1
--

CREATE TABLE "società".negoziopossiede (
    codice_prodotto integer NOT NULL,
    codice_negozio integer NOT NULL,
    prezzo numeric(10,2) NOT NULL,
    quantita integer NOT NULL,
    CONSTRAINT negoziopossiede_prezzo_check CHECK ((prezzo > (0)::numeric)),
    CONSTRAINT negoziopossiede_quantita_check CHECK ((quantita >= 0))
);


ALTER TABLE "società".negoziopossiede OWNER TO andrea_veneroni1;

--
-- Name: ordine_codice_seq; Type: SEQUENCE; Schema: società; Owner: andrea_veneroni1
--

CREATE SEQUENCE "società".ordine_codice_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "società".ordine_codice_seq OWNER TO andrea_veneroni1;

--
-- Name: ordine_codice_seq; Type: SEQUENCE OWNED BY; Schema: società; Owner: andrea_veneroni1
--

ALTER SEQUENCE "società".ordine_codice_seq OWNED BY "società".ordine.codice;


--
-- Name: ordinecontiene; Type: TABLE; Schema: società; Owner: andrea_veneroni1
--

CREATE TABLE "società".ordinecontiene (
    codice_ordine integer NOT NULL,
    codice_prodotto integer NOT NULL,
    quantita integer NOT NULL,
    prezzo numeric(10,2) NOT NULL,
    CONSTRAINT ordinecontiene_prezzo_check CHECK ((prezzo > (0)::numeric)),
    CONSTRAINT ordinecontiene_quantita_check CHECK ((quantita > 0))
);


ALTER TABLE "società".ordinecontiene OWNER TO andrea_veneroni1;

--
-- Name: prodotto; Type: TABLE; Schema: società; Owner: andrea_veneroni1
--

CREATE TABLE "società".prodotto (
    codice integer NOT NULL,
    nome character varying(50) NOT NULL,
    descrizione text NOT NULL
);


ALTER TABLE "società".prodotto OWNER TO andrea_veneroni1;

--
-- Name: prodotto_codice_seq; Type: SEQUENCE; Schema: società; Owner: andrea_veneroni1
--

CREATE SEQUENCE "società".prodotto_codice_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE "società".prodotto_codice_seq OWNER TO andrea_veneroni1;

--
-- Name: prodotto_codice_seq; Type: SEQUENCE OWNED BY; Schema: società; Owner: andrea_veneroni1
--

ALTER SEQUENCE "società".prodotto_codice_seq OWNED BY "società".prodotto.codice;


--
-- Name: fattura codice; Type: DEFAULT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".fattura ALTER COLUMN codice SET DEFAULT nextval('"società".fattura_codice_seq'::regclass);


--
-- Name: negozio codice; Type: DEFAULT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".negozio ALTER COLUMN codice SET DEFAULT nextval('"società".negozio_codice_seq'::regclass);


--
-- Name: ordine codice; Type: DEFAULT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".ordine ALTER COLUMN codice SET DEFAULT nextval('"società".ordine_codice_seq'::regclass);


--
-- Name: prodotto codice; Type: DEFAULT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".prodotto ALTER COLUMN codice SET DEFAULT nextval('"società".prodotto_codice_seq'::regclass);


--
-- Data for Name: credenziali; Type: TABLE DATA; Schema: società; Owner: andrea_veneroni1
--

INSERT INTO "società".credenziali VALUES ('admin', 'admin123', 'RSSMRA80A01H501Z', true);
INSERT INTO "società".credenziali VALUES ('lverdi', 'password123', 'VRDLGI85B15F205X', false);
INSERT INTO "società".credenziali VALUES ('gbianchi', 'password123', 'BNCGIA90C25L736Y', false);
INSERT INTO "società".credenziali VALUES ('fneri', 'password123', 'NRGFRN88D12A662K', false);
INSERT INTO "società".credenziali VALUES ('lmonti', 'password123', 'MNTLRA92E20B354T', false);


--
-- Data for Name: fattura; Type: TABLE DATA; Schema: società; Owner: andrea_veneroni1
--



--
-- Data for Name: fatturacontiene; Type: TABLE DATA; Schema: società; Owner: andrea_veneroni1
--



--
-- Data for Name: fornitore; Type: TABLE DATA; Schema: società; Owner: andrea_veneroni1
--

INSERT INTO "società".fornitore VALUES ('12345678901', 'Via Industria 10, Torino');
INSERT INTO "società".fornitore VALUES ('98765432109', 'Via Commercio 5, Napoli');
INSERT INTO "società".fornitore VALUES ('11223344556', 'Via Logistica 8, Bologna');


--
-- Data for Name: fornitorepossiede; Type: TABLE DATA; Schema: società; Owner: andrea_veneroni1
--

INSERT INTO "società".fornitorepossiede VALUES ('12345678901', 1, 299.99, 50);
INSERT INTO "società".fornitorepossiede VALUES ('98765432109', 1, 289.99, 30);
INSERT INTO "società".fornitorepossiede VALUES ('12345678901', 2, 799.99, 25);
INSERT INTO "società".fornitorepossiede VALUES ('11223344556', 2, 759.99, 40);
INSERT INTO "società".fornitorepossiede VALUES ('98765432109', 3, 199.99, 60);
INSERT INTO "società".fornitorepossiede VALUES ('12345678901', 4, 89.99, 100);
INSERT INTO "società".fornitorepossiede VALUES ('11223344556', 5, 29.99, 200);


--
-- Data for Name: negozio; Type: TABLE DATA; Schema: società; Owner: andrea_veneroni1
--

INSERT INTO "società".negozio VALUES (1, 'Via Roma 1, Milano', 'RSSMRA80A01H501Z', '09:00:00', '19:00:00');
INSERT INTO "società".negozio VALUES (2, 'Via Garibaldi 15, Roma', 'RSSMRA80A01H501Z', '08:30:00', '20:00:00');


--
-- Data for Name: negoziopossiede; Type: TABLE DATA; Schema: società; Owner: andrea_veneroni1
--

INSERT INTO "società".negoziopossiede VALUES (1, 1, 329.99, 10);
INSERT INTO "società".negoziopossiede VALUES (1, 2, 319.99, 8);
INSERT INTO "società".negoziopossiede VALUES (2, 1, 849.99, 5);
INSERT INTO "società".negoziopossiede VALUES (2, 2, 829.99, 7);
INSERT INTO "società".negoziopossiede VALUES (3, 1, 229.99, 12);
INSERT INTO "società".negoziopossiede VALUES (4, 1, 99.99, 15);
INSERT INTO "società".negoziopossiede VALUES (5, 1, 34.99, 25);


--
-- Data for Name: ordine; Type: TABLE DATA; Schema: società; Owner: andrea_veneroni1
--



--
-- Data for Name: ordinecontiene; Type: TABLE DATA; Schema: società; Owner: andrea_veneroni1
--



--
-- Data for Name: persona; Type: TABLE DATA; Schema: società; Owner: andrea_veneroni1
--

INSERT INTO "società".persona VALUES ('RSSMRA80A01H501Z', 'Mario', 'Rossi');
INSERT INTO "società".persona VALUES ('VRDLGI85B15F205X', 'Luigi', 'Verdi');
INSERT INTO "società".persona VALUES ('BNCGIA90C25L736Y', 'Giulia', 'Bianchi');
INSERT INTO "società".persona VALUES ('NRGFRN88D12A662K', 'Francesco', 'Neri');
INSERT INTO "società".persona VALUES ('MNTLRA92E20B354T', 'Laura', 'Monti');


--
-- Data for Name: prodotto; Type: TABLE DATA; Schema: società; Owner: andrea_veneroni1
--

INSERT INTO "società".prodotto VALUES (1, 'Smartphone XYZ', 'Smartphone di ultima generazione con fotocamera avanzata');
INSERT INTO "società".prodotto VALUES (2, 'Laptop ABC', 'Computer portatile per uso professionale');
INSERT INTO "società".prodotto VALUES (3, 'Tablet DEF', 'Tablet per intrattenimento e lavoro');
INSERT INTO "società".prodotto VALUES (4, 'Cuffie GHI', 'Cuffie wireless con cancellazione del rumore');
INSERT INTO "società".prodotto VALUES (5, 'Mouse JKL', 'Mouse ergonomico per computer');


--
-- Data for Name: storicotessere; Type: TABLE DATA; Schema: società; Owner: andrea_veneroni1
--



--
-- Data for Name: tessera; Type: TABLE DATA; Schema: società; Owner: andrea_veneroni1
--

INSERT INTO "società".tessera VALUES ('VRDLGI85B15F205X', 1, 350, '2024-01-15 10:00:00');
INSERT INTO "società".tessera VALUES ('BNCGIA90C25L736Y', 1, 120, '2024-02-01 14:30:00');
INSERT INTO "società".tessera VALUES ('NRGFRN88D12A662K', 2, 80, '2024-01-20 09:15:00');
INSERT INTO "società".tessera VALUES ('MNTLRA92E20B354T', 2, 450, '2024-01-10 16:45:00');


--
-- Name: fattura_codice_seq; Type: SEQUENCE SET; Schema: società; Owner: andrea_veneroni1
--

SELECT pg_catalog.setval('"società".fattura_codice_seq', 1, false);


--
-- Name: negozio_codice_seq; Type: SEQUENCE SET; Schema: società; Owner: andrea_veneroni1
--

SELECT pg_catalog.setval('"società".negozio_codice_seq', 2, true);


--
-- Name: ordine_codice_seq; Type: SEQUENCE SET; Schema: società; Owner: andrea_veneroni1
--

SELECT pg_catalog.setval('"società".ordine_codice_seq', 1, false);


--
-- Name: prodotto_codice_seq; Type: SEQUENCE SET; Schema: società; Owner: andrea_veneroni1
--

SELECT pg_catalog.setval('"società".prodotto_codice_seq', 5, true);


--
-- Name: credenziali credenziali_cf_persona_manager_key; Type: CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".credenziali
    ADD CONSTRAINT credenziali_cf_persona_manager_key UNIQUE (cf_persona, manager);


--
-- Name: credenziali credenziali_pkey; Type: CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".credenziali
    ADD CONSTRAINT credenziali_pkey PRIMARY KEY (username);


--
-- Name: fattura fattura_pkey; Type: CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".fattura
    ADD CONSTRAINT fattura_pkey PRIMARY KEY (codice);


--
-- Name: fatturacontiene fatturacontiene_pkey; Type: CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".fatturacontiene
    ADD CONSTRAINT fatturacontiene_pkey PRIMARY KEY (codice_fattura, codice_prodotto);


--
-- Name: fornitore fornitore_pkey; Type: CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".fornitore
    ADD CONSTRAINT fornitore_pkey PRIMARY KEY (p_iva);


--
-- Name: fornitorepossiede fornitorepossiede_pkey; Type: CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".fornitorepossiede
    ADD CONSTRAINT fornitorepossiede_pkey PRIMARY KEY (codice_fornitore, codice_prodotto);


--
-- Name: negozio negozio_pkey; Type: CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".negozio
    ADD CONSTRAINT negozio_pkey PRIMARY KEY (codice);


--
-- Name: negoziopossiede negoziopossiede_pkey; Type: CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".negoziopossiede
    ADD CONSTRAINT negoziopossiede_pkey PRIMARY KEY (codice_prodotto, codice_negozio);


--
-- Name: ordine ordine_pkey; Type: CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".ordine
    ADD CONSTRAINT ordine_pkey PRIMARY KEY (codice);


--
-- Name: ordinecontiene ordinecontiene_pkey; Type: CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".ordinecontiene
    ADD CONSTRAINT ordinecontiene_pkey PRIMARY KEY (codice_ordine, codice_prodotto);


--
-- Name: persona persona_pkey; Type: CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".persona
    ADD CONSTRAINT persona_pkey PRIMARY KEY (cf);


--
-- Name: prodotto prodotto_pkey; Type: CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".prodotto
    ADD CONSTRAINT prodotto_pkey PRIMARY KEY (codice);


--
-- Name: storicotessere storicotessere_pkey; Type: CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".storicotessere
    ADD CONSTRAINT storicotessere_pkey PRIMARY KEY (cf_cliente, codice_negozio_eliminato, data_eliminazione_negozio);


--
-- Name: tessera tessera_pkey; Type: CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".tessera
    ADD CONSTRAINT tessera_pkey PRIMARY KEY (cf_cliente);


--
-- Name: idx_fattura_cf_cliente; Type: INDEX; Schema: società; Owner: andrea_veneroni1
--

CREATE INDEX idx_fattura_cf_cliente ON "società".fattura USING btree (cf_cliente);


--
-- Name: idx_fattura_data_acquisto; Type: INDEX; Schema: società; Owner: andrea_veneroni1
--

CREATE INDEX idx_fattura_data_acquisto ON "società".fattura USING btree (data_acquisto);


--
-- Name: idx_fattura_negozio; Type: INDEX; Schema: società; Owner: andrea_veneroni1
--

CREATE INDEX idx_fattura_negozio ON "società".fattura USING btree (codice_negozio);


--
-- Name: idx_negozio_possiede_quantita; Type: INDEX; Schema: società; Owner: andrea_veneroni1
--

CREATE INDEX idx_negozio_possiede_quantita ON "società".negoziopossiede USING btree (quantita);


--
-- Name: idx_ordine_data_consegna; Type: INDEX; Schema: società; Owner: andrea_veneroni1
--

CREATE INDEX idx_ordine_data_consegna ON "società".ordine USING btree (data_consegna);


--
-- Name: idx_ordine_negozio; Type: INDEX; Schema: società; Owner: andrea_veneroni1
--

CREATE INDEX idx_ordine_negozio ON "società".ordine USING btree (codice_negozio);


--
-- Name: idx_tessera_punti; Type: INDEX; Schema: società; Owner: andrea_veneroni1
--

CREATE INDEX idx_tessera_punti ON "società".tessera USING btree (punti);


--
-- Name: ordinecontiene trigger_aggiorna_disponibilita_fornitore; Type: TRIGGER; Schema: società; Owner: andrea_veneroni1
--

CREATE TRIGGER trigger_aggiorna_disponibilita_fornitore AFTER INSERT ON "società".ordinecontiene FOR EACH ROW EXECUTE FUNCTION "società".aggiorna_disponibilita_fornitore();


--
-- Name: fattura trigger_aggiorna_punti_tessera; Type: TRIGGER; Schema: società; Owner: andrea_veneroni1
--

CREATE TRIGGER trigger_aggiorna_punti_tessera AFTER INSERT ON "società".fattura FOR EACH ROW EXECUTE FUNCTION "società".aggiorna_punti_tessera();


--
-- Name: ordinecontiene trigger_aggiorna_scorte_da_ordine; Type: TRIGGER; Schema: società; Owner: andrea_veneroni1
--

CREATE TRIGGER trigger_aggiorna_scorte_da_ordine AFTER INSERT ON "società".ordinecontiene FOR EACH ROW EXECUTE FUNCTION "società".aggiorna_scorte_da_ordine();


--
-- Name: fatturacontiene trigger_aggiorna_scorte_negozio; Type: TRIGGER; Schema: società; Owner: andrea_veneroni1
--

CREATE TRIGGER trigger_aggiorna_scorte_negozio AFTER INSERT ON "società".fatturacontiene FOR EACH ROW EXECUTE FUNCTION "società".aggiorna_scorte_negozio();


--
-- Name: fatturacontiene trigger_aggiorna_totale_fattura; Type: TRIGGER; Schema: società; Owner: andrea_veneroni1
--

CREATE TRIGGER trigger_aggiorna_totale_fattura AFTER INSERT OR DELETE OR UPDATE ON "società".fatturacontiene FOR EACH ROW EXECUTE FUNCTION "società".aggiorna_totale_fattura();


--
-- Name: ordinecontiene trigger_aggiorna_totale_ordine; Type: TRIGGER; Schema: società; Owner: andrea_veneroni1
--

CREATE TRIGGER trigger_aggiorna_totale_ordine AFTER INSERT OR DELETE OR UPDATE ON "società".ordinecontiene FOR EACH ROW EXECUTE FUNCTION "società".aggiorna_totale_ordine();


--
-- Name: fattura trigger_applica_sconto_fattura; Type: TRIGGER; Schema: società; Owner: andrea_veneroni1
--

CREATE TRIGGER trigger_applica_sconto_fattura BEFORE INSERT ON "società".fattura FOR EACH ROW EXECUTE FUNCTION "società".applica_sconto_fattura();


--
-- Name: negozio trigger_controllo_responsabile; Type: TRIGGER; Schema: società; Owner: andrea_veneroni1
--

CREATE TRIGGER trigger_controllo_responsabile BEFORE INSERT OR UPDATE OF cf_responsabile ON "società".negozio FOR EACH ROW EXECUTE FUNCTION "società".controllo_inserimento_responsabile();


--
-- Name: negozio trigger_mantieni_storico_tessere; Type: TRIGGER; Schema: società; Owner: andrea_veneroni1
--

CREATE TRIGGER trigger_mantieni_storico_tessere BEFORE DELETE ON "società".negozio FOR EACH ROW EXECUTE FUNCTION "società".mantieni_storico_tessere();


--
-- Name: storicotessere trigger_update_storico_tessere; Type: TRIGGER; Schema: società; Owner: andrea_veneroni1
--

CREATE TRIGGER trigger_update_storico_tessere AFTER INSERT OR DELETE OR UPDATE ON "società".storicotessere FOR EACH ROW EXECUTE FUNCTION "società".update_storico_tessere();


--
-- Name: tessera trigger_utenti_piu_300_punti; Type: TRIGGER; Schema: società; Owner: andrea_veneroni1
--

CREATE TRIGGER trigger_utenti_piu_300_punti AFTER INSERT OR DELETE OR UPDATE ON "società".tessera FOR EACH ROW EXECUTE FUNCTION "società".update_utenti_piu_300_punti();


--
-- Name: credenziali credenziali_cf_persona_fkey; Type: FK CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".credenziali
    ADD CONSTRAINT credenziali_cf_persona_fkey FOREIGN KEY (cf_persona) REFERENCES "società".persona(cf) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: fattura fattura_cf_cliente_fkey; Type: FK CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".fattura
    ADD CONSTRAINT fattura_cf_cliente_fkey FOREIGN KEY (cf_cliente) REFERENCES "società".persona(cf) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: fattura fattura_codice_negozio_fkey; Type: FK CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".fattura
    ADD CONSTRAINT fattura_codice_negozio_fkey FOREIGN KEY (codice_negozio) REFERENCES "società".negozio(codice) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: fatturacontiene fatturacontiene_codice_fattura_fkey; Type: FK CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".fatturacontiene
    ADD CONSTRAINT fatturacontiene_codice_fattura_fkey FOREIGN KEY (codice_fattura) REFERENCES "società".fattura(codice) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: fatturacontiene fatturacontiene_codice_prodotto_fkey; Type: FK CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".fatturacontiene
    ADD CONSTRAINT fatturacontiene_codice_prodotto_fkey FOREIGN KEY (codice_prodotto) REFERENCES "società".prodotto(codice) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: fornitorepossiede fornitorepossiede_codice_fornitore_fkey; Type: FK CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".fornitorepossiede
    ADD CONSTRAINT fornitorepossiede_codice_fornitore_fkey FOREIGN KEY (codice_fornitore) REFERENCES "società".fornitore(p_iva) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: fornitorepossiede fornitorepossiede_codice_prodotto_fkey; Type: FK CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".fornitorepossiede
    ADD CONSTRAINT fornitorepossiede_codice_prodotto_fkey FOREIGN KEY (codice_prodotto) REFERENCES "società".prodotto(codice) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: negozio negozio_cf_responsabile_fkey; Type: FK CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".negozio
    ADD CONSTRAINT negozio_cf_responsabile_fkey FOREIGN KEY (cf_responsabile) REFERENCES "società".persona(cf) ON UPDATE CASCADE;


--
-- Name: negoziopossiede negoziopossiede_codice_negozio_fkey; Type: FK CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".negoziopossiede
    ADD CONSTRAINT negoziopossiede_codice_negozio_fkey FOREIGN KEY (codice_negozio) REFERENCES "società".negozio(codice) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: negoziopossiede negoziopossiede_codice_prodotto_fkey; Type: FK CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".negoziopossiede
    ADD CONSTRAINT negoziopossiede_codice_prodotto_fkey FOREIGN KEY (codice_prodotto) REFERENCES "società".prodotto(codice) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: ordine ordine_codice_fornitore_fkey; Type: FK CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".ordine
    ADD CONSTRAINT ordine_codice_fornitore_fkey FOREIGN KEY (codice_fornitore) REFERENCES "società".fornitore(p_iva) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: ordine ordine_codice_negozio_fkey; Type: FK CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".ordine
    ADD CONSTRAINT ordine_codice_negozio_fkey FOREIGN KEY (codice_negozio) REFERENCES "società".negozio(codice) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: ordinecontiene ordinecontiene_codice_ordine_fkey; Type: FK CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".ordinecontiene
    ADD CONSTRAINT ordinecontiene_codice_ordine_fkey FOREIGN KEY (codice_ordine) REFERENCES "società".ordine(codice) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: ordinecontiene ordinecontiene_codice_prodotto_fkey; Type: FK CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".ordinecontiene
    ADD CONSTRAINT ordinecontiene_codice_prodotto_fkey FOREIGN KEY (codice_prodotto) REFERENCES "società".prodotto(codice) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: storicotessere storicotessere_cf_cliente_fkey; Type: FK CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".storicotessere
    ADD CONSTRAINT storicotessere_cf_cliente_fkey FOREIGN KEY (cf_cliente) REFERENCES "società".persona(cf) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: tessera tessera_cf_cliente_fkey; Type: FK CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".tessera
    ADD CONSTRAINT tessera_cf_cliente_fkey FOREIGN KEY (cf_cliente) REFERENCES "società".persona(cf) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: tessera tessera_negozio_fkey; Type: FK CONSTRAINT; Schema: società; Owner: andrea_veneroni1
--

ALTER TABLE ONLY "società".tessera
    ADD CONSTRAINT tessera_negozio_fkey FOREIGN KEY (negozio) REFERENCES "società".negozio(codice) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE USAGE ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- Name: materialized_view_storico_tessere; Type: MATERIALIZED VIEW DATA; Schema: società; Owner: andrea_veneroni1
--

REFRESH MATERIALIZED VIEW "società".materialized_view_storico_tessere;


--
-- Name: materialized_view_utenti_piu_300_punti; Type: MATERIALIZED VIEW DATA; Schema: società; Owner: andrea_veneroni1
--

REFRESH MATERIALIZED VIEW "società".materialized_view_utenti_piu_300_punti;


--
-- PostgreSQL database dump complete
--

