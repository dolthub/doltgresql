-- Downloaded from: https://github.com/Enesuygurs/steamcafe/blob/b81055cc25896fb8df41b297b011683aa30c201c/sql%20dump.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 17.2
-- Dumped by pg_dump version 17.2

-- Started on 2024-12-24 07:33:27

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
-- TOC entry 249 (class 1255 OID 18471)
-- Name: bos_masalar_listele(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.bos_masalar_listele() RETURNS TABLE(masano integer, kapasite integer, durum character varying)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT m.masaNo, m.kapasite, m.durum
    FROM masalar m
    WHERE m.durum = 'bos';
END;
$$;


ALTER FUNCTION public.bos_masalar_listele() OWNER TO postgres;

--
-- TOC entry 250 (class 1255 OID 18516)
-- Name: degerlendirme_ortalama(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.degerlendirme_ortalama() RETURNS numeric
    LANGUAGE plpgsql
    AS $$
DECLARE
    ortalama NUMERIC;
BEGIN
    SELECT AVG(d.puan)
    INTO ortalama
    FROM Degerlendirmeler d;

    RETURN COALESCE(ortalama, 0);
END;
$$;


ALTER FUNCTION public.degerlendirme_ortalama() OWNER TO postgres;

--
-- TOC entry 265 (class 1255 OID 18451)
-- Name: kritik_stok_ekle(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.kritik_stok_ekle() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.stok < 5 THEN
        INSERT INTO KritikStok (urunNo, ad, mevcutStok, kategoriNo, tarih)
        VALUES (NEW.urunNo, NEW.ad, NEW.stok, NEW.kategoriNo, CURRENT_TIMESTAMP)
        ON CONFLICT (urunNo) DO UPDATE
        SET mevcutStok = EXCLUDED.mevcutStok,
            tarih = CURRENT_TIMESTAMP;
    ELSE
        DELETE FROM KritikStok WHERE urunNo = NEW.urunNo;
    END IF;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.kritik_stok_ekle() OWNER TO postgres;

--
-- TOC entry 246 (class 1255 OID 18437)
-- Name: masa_durum_guncelle(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.masa_durum_guncelle() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   
    IF TG_OP = 'INSERT' THEN
        UPDATE masalar
        SET durum = 'dolu'
        WHERE masaNo = NEW.masaNo;
        RETURN NEW;
    END IF;

    IF TG_OP = 'DELETE' THEN
        UPDATE masalar
        SET durum = 'bos'
        WHERE masaNo = OLD.masaNo;
        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$;


ALTER FUNCTION public.masa_durum_guncelle() OWNER TO postgres;

--
-- TOC entry 267 (class 1255 OID 18525)
-- Name: odeme(integer, timestamp without time zone, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.odeme(siparis_id integer, tarih timestamp without time zone, odeme_turu character varying) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    toplam_tutar NUMERIC(10,2); 
BEGIN
    SELECT SUM(u.fiyat * su.miktar)
    INTO toplam_tutar
    FROM siparisurunleri su
    JOIN urunler u ON su.urunNo = u.urunNo
    WHERE su.siparisNo = siparis_id;

    INSERT INTO Odemeler (siparisNo, tutar, tarih, odemeTuru)
    VALUES (siparis_id, toplam_tutar, tarih, odeme_turu);
END;
$$;


ALTER FUNCTION public.odeme(siparis_id integer, tarih timestamp without time zone, odeme_turu character varying) OWNER TO postgres;

--
-- TOC entry 253 (class 1255 OID 18708)
-- Name: rezervasyon_iptal(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.rezervasyon_iptal(p_rezervasyonno integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    DELETE FROM Rezervasyonlar WHERE rezervasyonNo = p_rezervasyonNo;
END;
$$;


ALTER FUNCTION public.rezervasyon_iptal(p_rezervasyonno integer) OWNER TO postgres;

--
-- TOC entry 251 (class 1255 OID 18517)
-- Name: siparis_stok_azalt(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.siparis_stok_azalt() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    PERFORM stok_guncelle(NEW.urunNo, NEW.miktar, false);
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.siparis_stok_azalt() OWNER TO postgres;

--
-- TOC entry 247 (class 1255 OID 18439)
-- Name: siparis_urun_temizle(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.siparis_urun_temizle() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    DELETE FROM siparisurunleri
    WHERE siparisNo = OLD.siparisNo;
    RETURN OLD;
END;
$$;


ALTER FUNCTION public.siparis_urun_temizle() OWNER TO postgres;

--
-- TOC entry 266 (class 1255 OID 18467)
-- Name: stok_artir(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.stok_artir() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    UPDATE urunler
    SET stok = stok + OLD.miktar
    WHERE urunNo = OLD.urunNo;
    
    RETURN OLD;
END;
$$;


ALTER FUNCTION public.stok_artir() OWNER TO postgres;

--
-- TOC entry 248 (class 1255 OID 18469)
-- Name: stok_guncelle(integer, integer, boolean); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.stok_guncelle(urun_id integer, miktar integer, artirma boolean) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF artirma THEN
        UPDATE urunler
        SET stok = stok + miktar
        WHERE urunNo = urun_id;
    ELSE
        UPDATE urunler
        SET stok = stok - miktar
        WHERE urunNo = urun_id;
    END IF;
END;
$$;


ALTER FUNCTION public.stok_guncelle(urun_id integer, miktar integer, artirma boolean) OWNER TO postgres;

--
-- TOC entry 245 (class 1255 OID 18429)
-- Name: toplam_tutar_hesapla(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.toplam_tutar_hesapla(siparis_id integer) RETURNS numeric
    LANGUAGE plpgsql
    AS $$
DECLARE
    toplam NUMERIC;
BEGIN
    SELECT SUM(u.fiyat * su.miktar)
    INTO toplam
    FROM siparisurunleri su
    JOIN urunler u ON su.urunNo = u.urunNo
    WHERE su.siparisNo = siparis_id;

    RETURN toplam;
END;
$$;


ALTER FUNCTION public.toplam_tutar_hesapla(siparis_id integer) OWNER TO postgres;

--
-- TOC entry 252 (class 1255 OID 18707)
-- Name: yeni_kisi_ekle(integer, character varying, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.yeni_kisi_ekle(p_kisino integer, p_ad character varying, p_kisituru character varying) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    INSERT INTO Kisiler (kisiNo, ad, kisiTuru) VALUES (p_kisiNo, p_ad, p_kisiTuru);
END;
$$;


ALTER FUNCTION public.yeni_kisi_ekle(p_kisino integer, p_ad character varying, p_kisituru character varying) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 243 (class 1259 OID 18495)
-- Name: calisanlar; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.calisanlar (
    kisino integer NOT NULL,
    rol character varying(30) NOT NULL,
    iletisim character varying(50) NOT NULL,
    performans integer,
    vardiya integer,
    kullaniciadi character varying(100) NOT NULL,
    sifre character varying(100) NOT NULL
);


ALTER TABLE public.calisanlar OWNER TO postgres;

--
-- TOC entry 223 (class 1259 OID 18314)
-- Name: degerlendirmeler; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.degerlendirmeler (
    degerlendirmeno integer NOT NULL,
    kisino integer,
    puan integer
);


ALTER TABLE public.degerlendirmeler OWNER TO postgres;

--
-- TOC entry 222 (class 1259 OID 18313)
-- Name: degerlendirmeler_degerlendirmeno_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.degerlendirmeler_degerlendirmeno_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.degerlendirmeler_degerlendirmeno_seq OWNER TO postgres;

--
-- TOC entry 4978 (class 0 OID 0)
-- Dependencies: 222
-- Name: degerlendirmeler_degerlendirmeno_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.degerlendirmeler_degerlendirmeno_seq OWNED BY public.degerlendirmeler.degerlendirmeno;


--
-- TOC entry 218 (class 1259 OID 18263)
-- Name: kisiler; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.kisiler (
    kisino integer NOT NULL,
    ad character varying(50) NOT NULL,
    kisituru character varying(20)
);


ALTER TABLE public.kisiler OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 18262)
-- Name: kisiler_kisino_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.kisiler_kisino_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.kisiler_kisino_seq OWNER TO postgres;

--
-- TOC entry 4979 (class 0 OID 0)
-- Dependencies: 217
-- Name: kisiler_kisino_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.kisiler_kisino_seq OWNED BY public.kisiler.kisino;


--
-- TOC entry 240 (class 1259 OID 18443)
-- Name: kritikstok; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.kritikstok (
    urunno integer NOT NULL,
    ad text NOT NULL,
    mevcutstok integer NOT NULL,
    kategorino integer,
    tarih timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.kritikstok OWNER TO postgres;

--
-- TOC entry 238 (class 1259 OID 18408)
-- Name: malzemeler; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.malzemeler (
    malzemeno integer NOT NULL,
    ad character varying(50),
    stok integer
);


ALTER TABLE public.malzemeler OWNER TO postgres;

--
-- TOC entry 237 (class 1259 OID 18407)
-- Name: malzemeler_malzemeno_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.malzemeler_malzemeno_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.malzemeler_malzemeno_seq OWNER TO postgres;

--
-- TOC entry 4980 (class 0 OID 0)
-- Dependencies: 237
-- Name: malzemeler_malzemeno_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.malzemeler_malzemeno_seq OWNED BY public.malzemeler.malzemeno;


--
-- TOC entry 225 (class 1259 OID 18326)
-- Name: masalar; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.masalar (
    masano integer NOT NULL,
    kapasite integer,
    durum character varying(20)
);


ALTER TABLE public.masalar OWNER TO postgres;

--
-- TOC entry 224 (class 1259 OID 18325)
-- Name: masalar_masano_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.masalar_masano_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.masalar_masano_seq OWNER TO postgres;

--
-- TOC entry 4981 (class 0 OID 0)
-- Dependencies: 224
-- Name: masalar_masano_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.masalar_masano_seq OWNED BY public.masalar.masano;


--
-- TOC entry 233 (class 1259 OID 18374)
-- Name: menu; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.menu (
    kategorino integer NOT NULL,
    kategoriad character varying(50)
);


ALTER TABLE public.menu OWNER TO postgres;

--
-- TOC entry 232 (class 1259 OID 18373)
-- Name: menu_kategorino_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.menu_kategorino_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.menu_kategorino_seq OWNER TO postgres;

--
-- TOC entry 4982 (class 0 OID 0)
-- Dependencies: 232
-- Name: menu_kategorino_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.menu_kategorino_seq OWNED BY public.menu.kategorino;


--
-- TOC entry 219 (class 1259 OID 18269)
-- Name: musteriler; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.musteriler (
    kisino integer NOT NULL
);


ALTER TABLE public.musteriler OWNER TO postgres;

--
-- TOC entry 231 (class 1259 OID 18362)
-- Name: odemeler; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.odemeler (
    odemeno integer NOT NULL,
    siparisno integer,
    tutar numeric(10,2),
    tarih timestamp without time zone,
    odemeturu character varying(20),
    durum character varying(1),
    CONSTRAINT durum CHECK ((upper((durum)::text) = ANY (ARRAY['A'::text, 'P'::text])))
);


ALTER TABLE public.odemeler OWNER TO postgres;

--
-- TOC entry 230 (class 1259 OID 18361)
-- Name: odemeler_odemeno_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.odemeler_odemeno_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.odemeler_odemeno_seq OWNER TO postgres;

--
-- TOC entry 4983 (class 0 OID 0)
-- Dependencies: 230
-- Name: odemeler_odemeno_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.odemeler_odemeno_seq OWNED BY public.odemeler.odemeno;


--
-- TOC entry 242 (class 1259 OID 18488)
-- Name: performans; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.performans (
    performansno integer NOT NULL,
    puan character varying(10)
);


ALTER TABLE public.performans OWNER TO postgres;

--
-- TOC entry 241 (class 1259 OID 18487)
-- Name: performans_performansno_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.performans_performansno_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.performans_performansno_seq OWNER TO postgres;

--
-- TOC entry 4984 (class 0 OID 0)
-- Dependencies: 241
-- Name: performans_performansno_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.performans_performansno_seq OWNED BY public.performans.performansno;


--
-- TOC entry 227 (class 1259 OID 18333)
-- Name: rezervasyonlar; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.rezervasyonlar (
    rezervasyonno integer NOT NULL,
    kisino integer,
    masano integer,
    ad character varying(50),
    kisisayisi integer,
    zaman timestamp without time zone
);


ALTER TABLE public.rezervasyonlar OWNER TO postgres;

--
-- TOC entry 226 (class 1259 OID 18332)
-- Name: rezervasyonlar_rezervasyonno_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.rezervasyonlar_rezervasyonno_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.rezervasyonlar_rezervasyonno_seq OWNER TO postgres;

--
-- TOC entry 4985 (class 0 OID 0)
-- Dependencies: 226
-- Name: rezervasyonlar_rezervasyonno_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.rezervasyonlar_rezervasyonno_seq OWNED BY public.rezervasyonlar.rezervasyonno;


--
-- TOC entry 229 (class 1259 OID 18350)
-- Name: siparisler; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.siparisler (
    siparisno integer NOT NULL,
    masano integer,
    zaman timestamp without time zone,
    durum character(1),
    CONSTRAINT durum CHECK ((upper((durum)::text) = ANY (ARRAY['A'::text, 'P'::text])))
);


ALTER TABLE public.siparisler OWNER TO postgres;

--
-- TOC entry 228 (class 1259 OID 18349)
-- Name: siparisler_siparisno_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.siparisler_siparisno_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.siparisler_siparisno_seq OWNER TO postgres;

--
-- TOC entry 4986 (class 0 OID 0)
-- Dependencies: 228
-- Name: siparisler_siparisno_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.siparisler_siparisno_seq OWNED BY public.siparisler.siparisno;


--
-- TOC entry 236 (class 1259 OID 18392)
-- Name: siparisurunleri; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.siparisurunleri (
    siparisno integer NOT NULL,
    urunno integer NOT NULL,
    miktar integer
);


ALTER TABLE public.siparisurunleri OWNER TO postgres;

--
-- TOC entry 235 (class 1259 OID 18381)
-- Name: urunler; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.urunler (
    urunno integer NOT NULL,
    kategorino integer,
    ad character varying(50),
    fiyat numeric(10,2),
    stok integer,
    resim character varying(500)
);


ALTER TABLE public.urunler OWNER TO postgres;

--
-- TOC entry 234 (class 1259 OID 18380)
-- Name: urunler_urunno_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.urunler_urunno_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.urunler_urunno_seq OWNER TO postgres;

--
-- TOC entry 4987 (class 0 OID 0)
-- Dependencies: 234
-- Name: urunler_urunno_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.urunler_urunno_seq OWNED BY public.urunler.urunno;


--
-- TOC entry 239 (class 1259 OID 18414)
-- Name: urunmalzemeleri; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.urunmalzemeleri (
    urunno integer NOT NULL,
    malzemeno integer NOT NULL,
    stok integer
);


ALTER TABLE public.urunmalzemeleri OWNER TO postgres;

--
-- TOC entry 221 (class 1259 OID 18302)
-- Name: vardiyalar; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.vardiyalar (
    vardiyano integer NOT NULL,
    vardiya character varying(20)
);


ALTER TABLE public.vardiyalar OWNER TO postgres;

--
-- TOC entry 220 (class 1259 OID 18301)
-- Name: vardiyalar_vardiyano_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.vardiyalar_vardiyano_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.vardiyalar_vardiyano_seq OWNER TO postgres;

--
-- TOC entry 4988 (class 0 OID 0)
-- Dependencies: 220
-- Name: vardiyalar_vardiyano_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.vardiyalar_vardiyano_seq OWNED BY public.vardiyalar.vardiyano;


--
-- TOC entry 244 (class 1259 OID 18587)
-- Name: yoneticiler; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.yoneticiler (
    kisino integer NOT NULL,
    kullaniciadi character varying(100) NOT NULL,
    sifre character varying(100) NOT NULL
);


ALTER TABLE public.yoneticiler OWNER TO postgres;

--
-- TOC entry 4729 (class 2604 OID 18317)
-- Name: degerlendirmeler degerlendirmeno; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.degerlendirmeler ALTER COLUMN degerlendirmeno SET DEFAULT nextval('public.degerlendirmeler_degerlendirmeno_seq'::regclass);


--
-- TOC entry 4727 (class 2604 OID 18266)
-- Name: kisiler kisino; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.kisiler ALTER COLUMN kisino SET DEFAULT nextval('public.kisiler_kisino_seq'::regclass);


--
-- TOC entry 4736 (class 2604 OID 18411)
-- Name: malzemeler malzemeno; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.malzemeler ALTER COLUMN malzemeno SET DEFAULT nextval('public.malzemeler_malzemeno_seq'::regclass);


--
-- TOC entry 4730 (class 2604 OID 18329)
-- Name: masalar masano; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.masalar ALTER COLUMN masano SET DEFAULT nextval('public.masalar_masano_seq'::regclass);


--
-- TOC entry 4734 (class 2604 OID 18377)
-- Name: menu kategorino; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.menu ALTER COLUMN kategorino SET DEFAULT nextval('public.menu_kategorino_seq'::regclass);


--
-- TOC entry 4733 (class 2604 OID 18365)
-- Name: odemeler odemeno; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.odemeler ALTER COLUMN odemeno SET DEFAULT nextval('public.odemeler_odemeno_seq'::regclass);


--
-- TOC entry 4738 (class 2604 OID 18491)
-- Name: performans performansno; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.performans ALTER COLUMN performansno SET DEFAULT nextval('public.performans_performansno_seq'::regclass);


--
-- TOC entry 4731 (class 2604 OID 18336)
-- Name: rezervasyonlar rezervasyonno; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.rezervasyonlar ALTER COLUMN rezervasyonno SET DEFAULT nextval('public.rezervasyonlar_rezervasyonno_seq'::regclass);


--
-- TOC entry 4732 (class 2604 OID 18353)
-- Name: siparisler siparisno; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.siparisler ALTER COLUMN siparisno SET DEFAULT nextval('public.siparisler_siparisno_seq'::regclass);


--
-- TOC entry 4735 (class 2604 OID 18384)
-- Name: urunler urunno; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.urunler ALTER COLUMN urunno SET DEFAULT nextval('public.urunler_urunno_seq'::regclass);


--
-- TOC entry 4728 (class 2604 OID 18305)
-- Name: vardiyalar vardiyano; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.vardiyalar ALTER COLUMN vardiyano SET DEFAULT nextval('public.vardiyalar_vardiyano_seq'::regclass);


--
-- TOC entry 4971 (class 0 OID 18495)
-- Dependencies: 243
-- Data for Name: calisanlar; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.calisanlar (kisino, rol, iletisim, performans, vardiya, kullaniciadi, sifre) FROM stdin;
4	Garson	05552341111	4	2	osman.dere	osman123
\.


--
-- TOC entry 4951 (class 0 OID 18314)
-- Dependencies: 223
-- Data for Name: degerlendirmeler; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.degerlendirmeler (degerlendirmeno, kisino, puan) FROM stdin;
2	4	2
3	3	3
\.


--
-- TOC entry 4946 (class 0 OID 18263)
-- Dependencies: 218
-- Data for Name: kisiler; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.kisiler (kisino, ad, kisituru) FROM stdin;
4	Osman Dere	Müşteri
3	Mehmet Şimşek	Garson
2	Ali Veli	Müşteri
\.


--
-- TOC entry 4968 (class 0 OID 18443)
-- Dependencies: 240
-- Data for Name: kritikstok; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.kritikstok (urunno, ad, mevcutstok, kategorino, tarih) FROM stdin;
8	Makarna	0	1	2024-12-24 05:56:49.246661
\.


--
-- TOC entry 4966 (class 0 OID 18408)
-- Dependencies: 238
-- Data for Name: malzemeler; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.malzemeler (malzemeno, ad, stok) FROM stdin;
1	Köfte	200
2	Tavuk	250
\.


--
-- TOC entry 4953 (class 0 OID 18326)
-- Dependencies: 225
-- Data for Name: masalar; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.masalar (masano, kapasite, durum) FROM stdin;
1	4	dolu
2	2	bos
\.


--
-- TOC entry 4961 (class 0 OID 18374)
-- Dependencies: 233
-- Data for Name: menu; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.menu (kategorino, kategoriad) FROM stdin;
1	Yiyecek
2	İçecek
3	Alkollü İçecekler
\.


--
-- TOC entry 4947 (class 0 OID 18269)
-- Dependencies: 219
-- Data for Name: musteriler; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.musteriler (kisino) FROM stdin;
4
3
\.


--
-- TOC entry 4959 (class 0 OID 18362)
-- Dependencies: 231
-- Data for Name: odemeler; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.odemeler (odemeno, siparisno, tutar, tarih, odemeturu, durum) FROM stdin;
1	1	500.00	2024-12-20 21:56:23.241128	Nakit	A
2	1	550.00	2024-12-20 21:56:23.241128	Nakit	A
3	1	1666.00	2024-12-21 04:49:17.200269	nakit	a
4	1	400.00	2024-12-21 17:00:38.250875	Kredi Kartı	P
17	43	900.00	2024-12-23 13:42:06.259374	Kredi Kartı	P
18	44	240.00	2024-12-23 13:46:37.873869	Kredi Kartı	P
19	45	1200.00	2024-12-23 13:47:28.534848	Kredi Kartı	P
20	46	240.00	2024-12-23 13:48:42.432571	Kredi Kartı	P
22	49	750.00	2024-12-23 13:58:39.280175	Nakit	P
25	52	240.00	2024-12-23 14:10:18.836109	Kredi Kartı	P
26	53	120.00	2024-12-23 14:12:57.904578	Kredi Kartı	P
28	55	120.00	2024-12-23 14:22:55.090914	Kredi Kartı	P
29	56	150.00	2024-12-23 14:25:07.92635	Kredi Kartı	P
30	57	120.00	2024-12-23 14:26:05.730262	Kredi Kartı	P
31	58	150.00	2024-12-23 14:27:26.704508	Kredi Kartı	P
34	61	120.00	2024-12-23 14:30:22.058589	Kredi Kartı	P
36	63	120.00	2024-12-23 14:33:27.003986	Kredi Kartı	P
37	64	150.00	2024-12-23 14:37:48.578311	Kredi Kartı	P
39	66	120.00	2024-12-23 14:43:22.480975	Kredi Kartı	P
42	69	200.00	2024-12-23 14:51:35.667451	Kredi Kartı	P
45	72	120.00	2024-12-23 14:57:11.07861	Kredi Kartı	P
46	73	120.00	2024-12-23 15:14:48.723948	Kredi Kartı	P
49	76	150.00	2024-12-23 16:25:10.742613	Kredi Kartı	P
52	79	470.00	2024-12-23 16:35:55.593395	Kredi Kartı	P
53	80	620.00	2024-12-23 16:43:04.020095	Kredi Kartı	P
54	81	270.00	2024-12-23 16:44:24.009965	Kredi Kartı	P
56	83	350.00	2024-12-23 17:00:29.277662	Kredi Kartı	P
57	1	200.00	2024-12-16 14:30:00	Kredi Kartı	\N
58	1	200.00	2024-12-16 14:30:00	Kredi Kartıxx	\N
59	5	450.00	2024-12-16 14:30:00	Kredi Kartı	\N
\.


--
-- TOC entry 4970 (class 0 OID 18488)
-- Dependencies: 242
-- Data for Name: performans; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.performans (performansno, puan) FROM stdin;
5	🤩
4	😆
3	😅
2	😀
1	🙂
\.


--
-- TOC entry 4955 (class 0 OID 18333)
-- Dependencies: 227
-- Data for Name: rezervasyonlar; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.rezervasyonlar (rezervasyonno, kisino, masano, ad, kisisayisi, zaman) FROM stdin;
\.


--
-- TOC entry 4957 (class 0 OID 18350)
-- Dependencies: 229
-- Data for Name: siparisler; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.siparisler (siparisno, masano, zaman, durum) FROM stdin;
1	1	2024-12-18 19:08:54.11116	A
5	1	2024-12-18 20:40:47.147819	P
3	1	2024-12-18 20:04:32.761642	P
4	1	2024-12-18 20:05:30.429649	P
43	1	2024-12-23 13:42:05.852597	A
44	1	2024-12-23 13:46:37.43247	A
45	1	2024-12-23 13:47:27.869714	A
46	1	2024-12-23 13:48:42.064064	A
49	1	2024-12-23 13:58:38.60436	A
52	1	2024-12-23 14:10:18.238961	A
53	1	2024-12-23 14:12:57.657206	A
55	1	2024-12-23 14:22:54.917594	A
56	1	2024-12-23 14:25:07.499393	A
57	1	2024-12-23 14:26:05.292943	A
58	1	2024-12-23 14:27:26.0806	A
61	1	2024-12-23 14:30:21.814115	A
63	1	2024-12-23 14:33:26.239576	A
64	1	2024-12-23 14:37:48.412433	A
66	1	2024-12-23 14:43:20.582563	A
69	1	2024-12-23 14:51:35.517761	A
72	1	2024-12-23 14:57:10.893214	A
73	1	2024-12-23 15:14:46.770771	A
76	1	2024-12-23 16:25:10.606109	A
79	1	2024-12-23 16:35:55.239404	A
80	1	2024-12-23 16:43:03.597879	A
81	1	2024-12-23 16:44:23.894774	A
83	1	2024-12-23 17:00:28.946267	A
6	1	2024-12-18 19:08:54.11116	A
\.


--
-- TOC entry 4964 (class 0 OID 18392)
-- Dependencies: 236
-- Data for Name: siparisurunleri; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.siparisurunleri (siparisno, urunno, miktar) FROM stdin;
43	6	9
44	3	9
45	6	16
1	1	1
3	2	1
4	2	1
5	2	3
46	3	9
49	4	5
52	5	2
53	5	1
55	5	1
56	4	1
57	5	1
58	4	1
61	5	1
63	5	1
64	4	1
66	5	1
69	1	1
72	5	1
73	5	1
76	4	1
79	4	1
79	5	1
79	1	1
80	2	1
80	4	1
80	5	1
80	1	1
81	4	1
81	5	1
83	2	1
83	1	1
6	5	3
\.


--
-- TOC entry 4963 (class 0 OID 18381)
-- Dependencies: 235
-- Data for Name: urunler; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.urunler (urunno, kategorino, ad, fiyat, stok, resim) FROM stdin;
8	1	Makarna	200.00	0	\N
1	1	Pasta	200.00	28	https://static.vecteezy.com/system/resources/previews/025/268/608/large_2x/german-chocolate-cake-with-ai-generated-free-png.png
4	2	Latte	150.00	18	https://static.vecteezy.com/system/resources/previews/023/742/327/large_2x/latte-coffee-isolated-illustration-ai-generative-free-png.png
6	3	Viski	300.00	33	\N
3	2	Milkshake	80.00	27	https://static.vecteezy.com/system/resources/previews/029/107/695/large_2x/chocolate-milkshake-with-toppings-on-a-transparent-background-ai-generative-free-png.png
5	2	Mocha	120.00	35	https://static.vecteezy.com/system/resources/previews/021/049/268/large_2x/3d-realistic-coffee-cup-coffee-cup-cartoon-free-png.png
10	1	Tavuk Döner	150.00	6	
9	2	Çorba	75.00	30	https://static.vecteezy.com/system/resources/previews/025/270/145/large_2x/onion-soup-with-ai-generated-free-png.png
2	1	Menemen	150.00	20	\N
\.


--
-- TOC entry 4967 (class 0 OID 18414)
-- Dependencies: 239
-- Data for Name: urunmalzemeleri; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.urunmalzemeleri (urunno, malzemeno, stok) FROM stdin;
\.


--
-- TOC entry 4949 (class 0 OID 18302)
-- Dependencies: 221
-- Data for Name: vardiyalar; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.vardiyalar (vardiyano, vardiya) FROM stdin;
1	8-5
2	5-12
\.


--
-- TOC entry 4972 (class 0 OID 18587)
-- Dependencies: 244
-- Data for Name: yoneticiler; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.yoneticiler (kisino, kullaniciadi, sifre) FROM stdin;
\.


--
-- TOC entry 4989 (class 0 OID 0)
-- Dependencies: 222
-- Name: degerlendirmeler_degerlendirmeno_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.degerlendirmeler_degerlendirmeno_seq', 1, false);


--
-- TOC entry 4990 (class 0 OID 0)
-- Dependencies: 217
-- Name: kisiler_kisino_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.kisiler_kisino_seq', 3, true);


--
-- TOC entry 4991 (class 0 OID 0)
-- Dependencies: 237
-- Name: malzemeler_malzemeno_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.malzemeler_malzemeno_seq', 1, false);


--
-- TOC entry 4992 (class 0 OID 0)
-- Dependencies: 224
-- Name: masalar_masano_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.masalar_masano_seq', 1, false);


--
-- TOC entry 4993 (class 0 OID 0)
-- Dependencies: 232
-- Name: menu_kategorino_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.menu_kategorino_seq', 1, false);


--
-- TOC entry 4994 (class 0 OID 0)
-- Dependencies: 230
-- Name: odemeler_odemeno_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.odemeler_odemeno_seq', 59, true);


--
-- TOC entry 4995 (class 0 OID 0)
-- Dependencies: 241
-- Name: performans_performansno_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.performans_performansno_seq', 1, false);


--
-- TOC entry 4996 (class 0 OID 0)
-- Dependencies: 226
-- Name: rezervasyonlar_rezervasyonno_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.rezervasyonlar_rezervasyonno_seq', 1, false);


--
-- TOC entry 4997 (class 0 OID 0)
-- Dependencies: 228
-- Name: siparisler_siparisno_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.siparisler_siparisno_seq', 83, true);


--
-- TOC entry 4998 (class 0 OID 0)
-- Dependencies: 234
-- Name: urunler_urunno_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.urunler_urunno_seq', 1, true);


--
-- TOC entry 4999 (class 0 OID 0)
-- Dependencies: 220
-- Name: vardiyalar_vardiyano_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.vardiyalar_vardiyano_seq', 1, false);


--
-- TOC entry 4774 (class 2606 OID 18499)
-- Name: calisanlar calisanlar_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.calisanlar
    ADD CONSTRAINT calisanlar_pkey PRIMARY KEY (kisino);


--
-- TOC entry 4748 (class 2606 OID 18319)
-- Name: degerlendirmeler degerlendirmeler_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.degerlendirmeler
    ADD CONSTRAINT degerlendirmeler_pkey PRIMARY KEY (degerlendirmeno);


--
-- TOC entry 4750 (class 2606 OID 18705)
-- Name: degerlendirmeler kisiNo_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.degerlendirmeler
    ADD CONSTRAINT "kisiNo_unique" UNIQUE (kisino);


--
-- TOC entry 4742 (class 2606 OID 18268)
-- Name: kisiler kisiler_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.kisiler
    ADD CONSTRAINT kisiler_pkey PRIMARY KEY (kisino);


--
-- TOC entry 4770 (class 2606 OID 18450)
-- Name: kritikstok kritikstok_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.kritikstok
    ADD CONSTRAINT kritikstok_pkey PRIMARY KEY (urunno);


--
-- TOC entry 4766 (class 2606 OID 18413)
-- Name: malzemeler malzemeler_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.malzemeler
    ADD CONSTRAINT malzemeler_pkey PRIMARY KEY (malzemeno);


--
-- TOC entry 4752 (class 2606 OID 18331)
-- Name: masalar masalar_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.masalar
    ADD CONSTRAINT masalar_pkey PRIMARY KEY (masano);


--
-- TOC entry 4760 (class 2606 OID 18379)
-- Name: menu menu_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.menu
    ADD CONSTRAINT menu_pkey PRIMARY KEY (kategorino);


--
-- TOC entry 4744 (class 2606 OID 18273)
-- Name: musteriler musteriler_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.musteriler
    ADD CONSTRAINT musteriler_pkey PRIMARY KEY (kisino);


--
-- TOC entry 4758 (class 2606 OID 18367)
-- Name: odemeler odemeler_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.odemeler
    ADD CONSTRAINT odemeler_pkey PRIMARY KEY (odemeno);


--
-- TOC entry 4772 (class 2606 OID 18493)
-- Name: performans performans_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.performans
    ADD CONSTRAINT performans_pkey PRIMARY KEY (performansno);


--
-- TOC entry 4754 (class 2606 OID 18338)
-- Name: rezervasyonlar rezervasyonlar_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.rezervasyonlar
    ADD CONSTRAINT rezervasyonlar_pkey PRIMARY KEY (rezervasyonno);


--
-- TOC entry 4756 (class 2606 OID 18355)
-- Name: siparisler siparisler_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.siparisler
    ADD CONSTRAINT siparisler_pkey PRIMARY KEY (siparisno);


--
-- TOC entry 4764 (class 2606 OID 18396)
-- Name: siparisurunleri siparisurunleri_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.siparisurunleri
    ADD CONSTRAINT siparisurunleri_pkey PRIMARY KEY (siparisno, urunno);


--
-- TOC entry 4762 (class 2606 OID 18386)
-- Name: urunler urunler_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.urunler
    ADD CONSTRAINT urunler_pkey PRIMARY KEY (urunno);


--
-- TOC entry 4768 (class 2606 OID 18418)
-- Name: urunmalzemeleri urunmalzemeleri_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.urunmalzemeleri
    ADD CONSTRAINT urunmalzemeleri_pkey PRIMARY KEY (urunno, malzemeno);


--
-- TOC entry 4746 (class 2606 OID 18307)
-- Name: vardiyalar vardiyalar_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.vardiyalar
    ADD CONSTRAINT vardiyalar_pkey PRIMARY KEY (vardiyano);


--
-- TOC entry 4776 (class 2606 OID 18591)
-- Name: yoneticiler yonetici_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.yoneticiler
    ADD CONSTRAINT yonetici_pkey PRIMARY KEY (kisino);


--
-- TOC entry 4797 (class 2620 OID 18452)
-- Name: urunler kritik_stok_trigger; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER kritik_stok_trigger AFTER UPDATE ON public.urunler FOR EACH ROW EXECUTE FUNCTION public.kritik_stok_ekle();


--
-- TOC entry 4793 (class 2620 OID 18710)
-- Name: rezervasyonlar masa_durum_guncelle; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER masa_durum_guncelle AFTER INSERT OR DELETE ON public.rezervasyonlar FOR EACH ROW EXECUTE FUNCTION public.masa_durum_guncelle();


--
-- TOC entry 4794 (class 2620 OID 18709)
-- Name: rezervasyonlar masa_durum_tetikleyici; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER masa_durum_tetikleyici AFTER INSERT OR DELETE ON public.rezervasyonlar FOR EACH ROW EXECUTE FUNCTION public.masa_durum_guncelle();


--
-- TOC entry 4795 (class 2620 OID 18438)
-- Name: rezervasyonlar masa_durum_trigger; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER masa_durum_trigger AFTER INSERT ON public.rezervasyonlar FOR EACH ROW EXECUTE FUNCTION public.masa_durum_guncelle();


--
-- TOC entry 4796 (class 2620 OID 18440)
-- Name: siparisler siparis_sil_trigger; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER siparis_sil_trigger BEFORE DELETE ON public.siparisler FOR EACH ROW EXECUTE FUNCTION public.siparis_urun_temizle();


--
-- TOC entry 4798 (class 2620 OID 18518)
-- Name: siparisurunleri siparis_stok_trigger; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER siparis_stok_trigger AFTER INSERT ON public.siparisurunleri FOR EACH ROW EXECUTE FUNCTION public.siparis_stok_azalt();


--
-- TOC entry 4799 (class 2620 OID 18468)
-- Name: siparisurunleri stok_artir_trigger; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER stok_artir_trigger AFTER DELETE ON public.siparisurunleri FOR EACH ROW EXECUTE FUNCTION public.stok_artir();


--
-- TOC entry 4788 (class 2606 OID 18692)
-- Name: calisanlar calisanlar_kisino_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.calisanlar
    ADD CONSTRAINT calisanlar_kisino_fkey FOREIGN KEY (kisino) REFERENCES public.kisiler(kisino) ON DELETE CASCADE;


--
-- TOC entry 4789 (class 2606 OID 18505)
-- Name: calisanlar calisanlar_performans_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.calisanlar
    ADD CONSTRAINT calisanlar_performans_fkey FOREIGN KEY (performans) REFERENCES public.performans(performansno);


--
-- TOC entry 4790 (class 2606 OID 18510)
-- Name: calisanlar calisanlar_vardiya_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.calisanlar
    ADD CONSTRAINT calisanlar_vardiya_fkey FOREIGN KEY (vardiya) REFERENCES public.vardiyalar(vardiyano);


--
-- TOC entry 4778 (class 2606 OID 18599)
-- Name: degerlendirmeler degerlendirmeler_kisino_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.degerlendirmeler
    ADD CONSTRAINT degerlendirmeler_kisino_fkey FOREIGN KEY (kisino) REFERENCES public.musteriler(kisino) ON DELETE CASCADE;


--
-- TOC entry 4791 (class 2606 OID 18677)
-- Name: calisanlar fk_calisanlar_kisiler; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.calisanlar
    ADD CONSTRAINT fk_calisanlar_kisiler FOREIGN KEY (kisino) REFERENCES public.kisiler(kisino) ON DELETE CASCADE;


--
-- TOC entry 4777 (class 2606 OID 18697)
-- Name: musteriler musteriler_kisino_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.musteriler
    ADD CONSTRAINT musteriler_kisino_fkey FOREIGN KEY (kisino) REFERENCES public.kisiler(kisino) ON DELETE CASCADE;


--
-- TOC entry 4782 (class 2606 OID 18629)
-- Name: odemeler odemeler_siparisno_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.odemeler
    ADD CONSTRAINT odemeler_siparisno_fkey FOREIGN KEY (siparisno) REFERENCES public.siparisler(siparisno) ON DELETE CASCADE;


--
-- TOC entry 4779 (class 2606 OID 18604)
-- Name: rezervasyonlar rezervasyonlar_kisino_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.rezervasyonlar
    ADD CONSTRAINT rezervasyonlar_kisino_fkey FOREIGN KEY (kisino) REFERENCES public.musteriler(kisino) ON DELETE CASCADE;


--
-- TOC entry 4780 (class 2606 OID 18619)
-- Name: rezervasyonlar rezervasyonlar_masano_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.rezervasyonlar
    ADD CONSTRAINT rezervasyonlar_masano_fkey FOREIGN KEY (masano) REFERENCES public.masalar(masano) ON DELETE CASCADE;


--
-- TOC entry 4781 (class 2606 OID 18624)
-- Name: siparisler siparisler_masano_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.siparisler
    ADD CONSTRAINT siparisler_masano_fkey FOREIGN KEY (masano) REFERENCES public.masalar(masano) ON DELETE CASCADE;


--
-- TOC entry 4784 (class 2606 OID 18634)
-- Name: siparisurunleri siparisurunleri_siparisno_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.siparisurunleri
    ADD CONSTRAINT siparisurunleri_siparisno_fkey FOREIGN KEY (siparisno) REFERENCES public.siparisler(siparisno) ON DELETE CASCADE;


--
-- TOC entry 4785 (class 2606 OID 18644)
-- Name: siparisurunleri siparisurunleri_urunno_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.siparisurunleri
    ADD CONSTRAINT siparisurunleri_urunno_fkey FOREIGN KEY (urunno) REFERENCES public.urunler(urunno);


--
-- TOC entry 4783 (class 2606 OID 18387)
-- Name: urunler urunler_kategorino_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.urunler
    ADD CONSTRAINT urunler_kategorino_fkey FOREIGN KEY (kategorino) REFERENCES public.menu(kategorino);


--
-- TOC entry 4786 (class 2606 OID 18424)
-- Name: urunmalzemeleri urunmalzemeleri_malzemeno_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.urunmalzemeleri
    ADD CONSTRAINT urunmalzemeleri_malzemeno_fkey FOREIGN KEY (malzemeno) REFERENCES public.malzemeler(malzemeno);


--
-- TOC entry 4787 (class 2606 OID 18649)
-- Name: urunmalzemeleri urunmalzemeleri_urunno_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.urunmalzemeleri
    ADD CONSTRAINT urunmalzemeleri_urunno_fkey FOREIGN KEY (urunno) REFERENCES public.urunler(urunno) ON DELETE CASCADE;


--
-- TOC entry 4792 (class 2606 OID 18687)
-- Name: yoneticiler yonetici_kisino_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.yoneticiler
    ADD CONSTRAINT yonetici_kisino_fkey FOREIGN KEY (kisino) REFERENCES public.kisiler(kisino) ON DELETE CASCADE;


-- Completed on 2024-12-24 07:33:27

--
-- PostgreSQL database dump complete
--

