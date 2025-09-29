-- Downloaded from: https://github.com/openlawnz/openlawnz-data-processor/blob/bc26156a6e29d8ca2360d46af43aa33f7a93df4a/data/db.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 10.13
-- Dumped by pg_dump version 13.1

-- Started on 2021-01-08 11:22:51

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
-- TOC entry 4036 (class 1262 OID 17857)
-- Name: dev; Type: DATABASE; Schema: -; Owner: -
--

CREATE DATABASE dev WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE = 'en_US.UTF-8';


\connect dev

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
-- TOC entry 6 (class 2615 OID 17859)
-- Name: main; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA main;


--
-- TOC entry 10 (class 2615 OID 17858)
-- Name: ugc; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA ugc;


--
-- TOC entry 2 (class 3079 OID 1133125)
-- Name: tsm_system_rows; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS tsm_system_rows WITH SCHEMA public;


--
-- TOC entry 4037 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION tsm_system_rows; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION tsm_system_rows IS 'TABLESAMPLE method which accepts number of rows as a limit';


--
-- TOC entry 3 (class 3079 OID 1132697)
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- TOC entry 4038 (class 0 OID 0)
-- Dependencies: 3
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- TOC entry 574 (class 1247 OID 17868)
-- Name: party_type; Type: TYPE; Schema: main; Owner: -
--

CREATE TYPE main.party_type AS ENUM (
    'applicant',
    'plaintiff',
    'appellant',
    'respondent',
    'defendant'
);


--
-- TOC entry 571 (class 1247 OID 17863)
-- Name: facet_type; Type: TYPE; Schema: ugc; Owner: -
--

CREATE TYPE ugc.facet_type AS ENUM (
    'boolean',
    'date'
);


--
-- TOC entry 258 (class 1255 OID 17886)
-- Name: f_boolean_facet_values(text); Type: FUNCTION; Schema: main; Owner: -
--

CREATE FUNCTION main.f_boolean_facet_values(facetname text) RETURNS TABLE(case_id text, value boolean)
    LANGUAGE sql
    AS $$
	SELECT
		FVM.case_id,
		DFV.value
	FROM funnel.boolean_facet_values DFV
	LEFT JOIN funnel.facet_value_metadata FVM on FVM.id = DFV.metadata_id
	LEFT JOIN funnel.facets F on F.id = FVM.facet_id
	WHERE F.name ILIKE facetName
	AND F.type = 'boolean';
$$;


--
-- TOC entry 267 (class 1255 OID 17887)
-- Name: f_case_text(text[], text); Type: FUNCTION; Schema: main; Owner: -
--

CREATE FUNCTION main.f_case_text(keywords text[], regex text) RETURNS TABLE(case_id text, keywords_used text[], matches_found boolean[])
    LANGUAGE sql
    AS $$
		SELECT 
			inner_C.id,
			ARRAY_AGG(inner_M.lexeme),
			ARRAY_AGG(
				(inner_C.case_text ILIKE CONCAT('%', inner_M.lexeme, '%')) OR --needs to be keyword match or regex match because some posix regex operators look like punctuation
			    (inner_c.case_text ~* inner_M.lexeme)
			)
		FROM
			main.cases inner_C
		CROSS JOIN 
			main.f_get_lexemes(
				(CASE WHEN regex IS NOT NULL 
				THEN ARRAY_APPEND(keywords, regex)
				ELSE keywords
				END)
			) inner_M
		WHERE inner_C.case_text <> ''
		GROUP BY
			inner_C.id
	$$;


--
-- TOC entry 268 (class 1255 OID 17888)
-- Name: f_date_facet_values(text); Type: FUNCTION; Schema: main; Owner: -
--

CREATE FUNCTION main.f_date_facet_values(facetname text) RETURNS TABLE(case_id text, value date)
    LANGUAGE sql
    AS $$
	SELECT
		FVM.case_id,
		DFV.value
	FROM funnel.date_facet_values DFV
	LEFT JOIN funnel.facet_value_metadata FVM on FVM.id = DFV.metadata_id
	LEFT JOIN funnel.facets F on F.id = FVM.facet_id
	WHERE F.name ILIKE facetName
	AND F.type = 'date';
$$;


--
-- TOC entry 269 (class 1255 OID 17889)
-- Name: f_get_lexemes(text[]); Type: FUNCTION; Schema: main; Owner: -
--

CREATE FUNCTION main.f_get_lexemes(keywords text[]) RETURNS TABLE(lexemes text[], lexeme text)
    LANGUAGE sql
    AS $$
		SELECT DISTINCT
				inner_L.lexemes,
				UNNEST(inner_L.lexemes) AS lexeme
			FROM (
				SELECT
					ARRAY_PREPEND (
						inner_S.stripped_keyword,												
						(CASE WHEN inner_S.keyword ILIKE '%*%' 
						THEN TS_LEXIZE('english_stem', inner_S.stripped_keyword)
						ELSE ARRAY[]::text[]
						END)
				   ) as lexemes						
				FROM (
					SELECT 
						keyword,
						REPLACE(keyword, '*', '') as stripped_keyword
					FROM (
						SELECT UNNEST(keywords) as keyword
					) inner_K						
				) inner_S
			) inner_L;
	$$;


SET default_tablespace = '';

--
-- TOC entry 201 (class 1259 OID 17890)
-- Name: legislation; Type: TABLE; Schema: main; Owner: -
--

CREATE TABLE main.legislation (
    title character varying(255),
    link character varying(255),
    year character varying(255),
    alerts text,
    id text NOT NULL,
    last_modified date DEFAULT CURRENT_DATE,
    parsers_version text
);


--
-- TOC entry 4039 (class 0 OID 0)
-- Dependencies: 201
-- Name: TABLE legislation; Type: COMMENT; Schema: main; Owner: -
--

COMMENT ON TABLE main.legislation IS '@name legislationTable
@omit create,update,delete';


--
-- TOC entry 270 (class 1255 OID 17896)
-- Name: get_legislation_by_title(character varying, integer); Type: FUNCTION; Schema: main; Owner: -
--

CREATE FUNCTION main.get_legislation_by_title(title character varying, id integer) RETURNS main.legislation
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
legislation main.legislation;
BEGIN
  IF title IS NOT NULL THEN
 	SELECT *
	INTO legislation
	FROM main.legislation_title(title)
	LIMIT 1;
  END IF;
  RETURN legislation;
  END;
$$;


--
-- TOC entry 4040 (class 0 OID 0)
-- Dependencies: 270
-- Name: FUNCTION get_legislation_by_title(title character varying, id integer); Type: COMMENT; Schema: main; Owner: -
--

COMMENT ON FUNCTION main.get_legislation_by_title(title character varying, id integer) IS '@name legislation';


--
-- TOC entry 271 (class 1255 OID 17897)
-- Name: get_search_results(character varying, integer, integer); Type: FUNCTION; Schema: main; Owner: -
--

CREATE FUNCTION main.get_search_results(search character varying, start integer, "end" integer) RETURNS TABLE(case_id integer, case_name character varying, citation character varying, case_date date, rank real, total_matches bigint)
    LANGUAGE plpgsql
    AS $$
begin
	return query
	select 
		s.case_id,
		s.case_name,
		s.citation,
		s.case_date,
		ts_rank_cd('{0.5,0.7,0.9,1.0}',"document",query) as rank,
		count(*) over ()
	from main.cases.case_search_documents s, plainto_tsquery('english',"search") query
	where query @@ "document" 
	order by rank desc
	offset start
	limit ("end"-start);
end; $$;


--
-- TOC entry 250 (class 1255 OID 17898)
-- Name: legislation_title(character varying); Type: FUNCTION; Schema: main; Owner: -
--

CREATE FUNCTION main.legislation_title(_title character varying) RETURNS main.legislation
    LANGUAGE sql STABLE
    AS $$
  SELECT *
  FROM main.legislation l
  WHERE l.title ILIKE _title
  ORDER BY (l.title = _title) DESC, l.title
  LIMIT(1)
$$;


--
-- TOC entry 4041 (class 0 OID 0)
-- Dependencies: 250
-- Name: FUNCTION legislation_title(_title character varying); Type: COMMENT; Schema: main; Owner: -
--

COMMENT ON FUNCTION main.legislation_title(_title character varying) IS '@omit execute';


--
-- TOC entry 272 (class 1255 OID 17899)
-- Name: researcher_query(text, date, date, text[], text); Type: FUNCTION; Schema: main; Owner: -
--

CREATE FUNCTION main.researcher_query(inputcategory text, startdate date, enddate date, keywords text[], regex text) RETURNS TABLE(id text, name text, date date, citation character varying[], court text, cases_cited character varying[], legislation_referenced text[], case_text text, pdf_url text, acc_date_of_injury date, acc_date_of_claim date, acc_sufficient_evidence boolean, keywords_used text[], matches_found boolean[])
    LANGUAGE sql
    AS $$
	SELECT
		C.id,
		C.case_name,
		C.case_date,			
		CC.citation,
		C.court_id,
		--CO.acronym,
		CCC.citations,
		L.legislation_referenced,
		C.case_text,
		CP.pdf_url,		
		DFV.value,
		DFV2.value,
		BFV.value,
		K.keywords_used,
		K.matches_found
	FROM
		main.cases C
	LEFT JOIN main.v_case_citations CC ON CC.case_id = C.id
	LEFT JOIN main.v_pdfs CP ON CP.pdf_id = C.pdf_id
	LEFT JOIN main.v_cited_cases_citations CCC ON CCC.case_origin = C.id
	LEFT JOIN main.v_legislation L ON L.case_id = C.id
	LEFT JOIN main.f_case_text(keywords, regex) K on C.id = K.case_id
	LEFT JOIN main.v_categories C2 on C2.case_id = C.id
	LEFT JOIN main.f_date_facet_values('date of injury') DFV on DFV.case_id = C.id
	LEFT JOIN main.f_date_facet_values('date of claim') DFV2 on DFV2.case_id = C.id
	LEFT JOIN main.f_boolean_facet_values('sufficient evidence') BFV on BFV.case_id = C.id
	WHERE
		(startDate IS NULL OR C.case_date >= startDate)
	AND (endDate IS NULL OR C.case_date <= endDate)
	AND (inputCategory IS NULL OR inputCategory ILIKE C2.category)
	GROUP BY 
		C.id,
		C.case_name,
		C.case_date,
		CC.citation,
		--CO.acronym,
		CCC.citations,
		L.legislation_referenced,
		C.case_text,
		CP.pdf_url,
		DFV.value,
		DFV2.value,
		BFV.value,
		K.keywords_used,
		K.matches_found	
	ORDER BY
		C.id;
$$;


--
-- TOC entry 200 (class 1259 OID 17879)
-- Name: cases; Type: TABLE; Schema: main; Owner: -
--

CREATE TABLE main.cases (
    case_date date,
    case_text text,
    case_name character varying(1000),
    is_valid boolean,
    id text NOT NULL,
    pdf_id text,
    court_id text,
    lawreport_id text,
    location text,
    conversion_engine text,
    court_filing_number text,
    last_modified date DEFAULT CURRENT_DATE,
    parsers_version text
);


--
-- TOC entry 4042 (class 0 OID 0)
-- Dependencies: 200
-- Name: TABLE cases; Type: COMMENT; Schema: main; Owner: -
--

COMMENT ON TABLE main.cases IS '@omit create, update, delete';


--
-- TOC entry 273 (class 1255 OID 17900)
-- Name: search_by_legislation_title(main.legislation); Type: FUNCTION; Schema: main; Owner: -
--

CREATE FUNCTION main.search_by_legislation_title(u main.legislation) RETURNS SETOF main.cases
    LANGUAGE sql STABLE
    AS $$
 select cases.*
 from main.cases
 inner join main.legislation_to_cases
 on (cases.id = legislation_to_cases.case_id)
 inner join main.legislation
 on (legislation.id = legislation_to_cases.legislation_id)
 where u.title = legislation.title
$$;


--
-- TOC entry 202 (class 1259 OID 17901)
-- Name: legislation_to_cases; Type: TABLE; Schema: main; Owner: -
--

CREATE TABLE main.legislation_to_cases (
    section character varying(255) NOT NULL,
    count bigint,
    legislation_id text NOT NULL,
    case_id text NOT NULL,
    extraction_confidence smallint,
    last_modified date DEFAULT CURRENT_DATE,
    parsers_version text
);


--
-- TOC entry 4043 (class 0 OID 0)
-- Dependencies: 202
-- Name: TABLE legislation_to_cases; Type: COMMENT; Schema: main; Owner: -
--

COMMENT ON TABLE main.legislation_to_cases IS '@omit create, update, delete';


--
-- TOC entry 274 (class 1255 OID 17907)
-- Name: search_by_section(main.legislation_to_cases); Type: FUNCTION; Schema: main; Owner: -
--

CREATE FUNCTION main.search_by_section(u main.legislation_to_cases) RETURNS SETOF main.cases
    LANGUAGE sql STABLE
    AS $$
 select cases.*
 from main.cases
 left join main.legislation_to_cases
 on (cases.id = legislation_to_cases.case_id)
 where legislation_to_cases.section = u.section
$$;


--
-- TOC entry 249 (class 1255 OID 17908)
-- Name: newrow(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.newrow() RETURNS trigger
    LANGUAGE plpgsql
    AS $$  
BEGIN  
    PERFORM pg_notify('FileEventChannel', row_to_json(NEW)::text);
    RETURN NEW;
END;  
$$;


--
-- TOC entry 232 (class 1255 OID 856028)
-- Name: sync_lastmod(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.sync_lastmod() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  NEW.last_modified := NOW();

  RETURN NEW;
END;
$$;


--
-- TOC entry 246 (class 1255 OID 856069)
-- Name: sync_lastmod_case_citations(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.sync_lastmod_case_citations() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ --here
BEGIN
  NEW.last_modified := NOW();

  RETURN NEW;
END;
$$;


--
-- TOC entry 247 (class 1255 OID 856071)
-- Name: sync_lastmod_case_pdfs(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.sync_lastmod_case_pdfs() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ --here
BEGIN
  NEW.last_modified := NOW();

  RETURN NEW;
END;
$$;


--
-- TOC entry 248 (class 1255 OID 856073)
-- Name: sync_lastmod_cases_cited(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.sync_lastmod_cases_cited() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ --here
BEGIN
  NEW.last_modified := NOW();

  RETURN NEW;
END;
$$;


--
-- TOC entry 229 (class 1255 OID 856110)
-- Name: sync_lastmod_categories(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.sync_lastmod_categories() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ --here
BEGIN
  NEW.last_modified := NOW();
  RETURN NEW;
END;
$$;


--
-- TOC entry 245 (class 1255 OID 856067)
-- Name: sync_lastmod_category_to_cases(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.sync_lastmod_category_to_cases() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ --here
BEGIN
  NEW.last_modified := NOW();

  RETURN NEW;
END;
$$;


--
-- TOC entry 230 (class 1255 OID 856120)
-- Name: sync_lastmod_courts(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.sync_lastmod_courts() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ --here
BEGIN
  NEW.last_modified := NOW();
  RETURN NEW;
END;
$$;


--
-- TOC entry 231 (class 1255 OID 856130)
-- Name: sync_lastmod_legislation(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.sync_lastmod_legislation() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ --here
BEGIN
  NEW.last_modified := NOW();
  RETURN NEW;
END;
$$;


--
-- TOC entry 251 (class 1255 OID 856091)
-- Name: sync_lastmod_legislation_to_cases(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.sync_lastmod_legislation_to_cases() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ --here
BEGIN
  NEW.last_modified := NOW();

  RETURN NEW;
END;
$$;


--
-- TOC entry 252 (class 1255 OID 856100)
-- Name: sync_lastmod_party_and_representative_to_cases(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.sync_lastmod_party_and_representative_to_cases() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ --here
BEGIN
  NEW.last_modified := NOW();

  RETURN NEW;
END;
$$;


--
-- TOC entry 253 (class 1255 OID 1132740)
-- Name: project_acc_cases(); Type: FUNCTION; Schema: ugc; Owner: -
--

CREATE FUNCTION ugc.project_acc_cases() RETURNS TABLE(id text, name character varying)
    LANGUAGE plpgsql
    AS $$
BEGIN
	RETURN QUERY SELECT main.cases.id, main.cases.case_name FROM main.cases
INNER JOIN main.category_to_cases ON main.cases.id = main.category_to_cases.case_id
WHERE main.category_to_cases.category_id = 'acc';
END; $$;


--
-- TOC entry 254 (class 1255 OID 1132752)
-- Name: project_general(); Type: FUNCTION; Schema: ugc; Owner: -
--

CREATE FUNCTION ugc.project_general() RETURNS TABLE(id text, name character varying)
    LANGUAGE plpgsql
    AS $$BEGIN
	RETURN QUERY SELECT main.cases.id, main.cases.case_name FROM main.cases;
END; $$;


--
-- TOC entry 209 (class 1259 OID 17945)
-- Name: case_citations; Type: TABLE; Schema: main; Owner: -
--

CREATE TABLE main.case_citations (
    citation character varying(50) DEFAULT ''::character varying NOT NULL,
    year integer,
    id text NOT NULL,
    case_id text,
    last_modified date DEFAULT CURRENT_DATE,
    parsers_version text
);


--
-- TOC entry 4044 (class 0 OID 0)
-- Dependencies: 209
-- Name: TABLE case_citations; Type: COMMENT; Schema: main; Owner: -
--

COMMENT ON TABLE main.case_citations IS '@omit create, update, delete';


--
-- TOC entry 210 (class 1259 OID 17952)
-- Name: case_pdfs; Type: TABLE; Schema: main; Owner: -
--

CREATE TABLE main.case_pdfs (
    fetch_date date,
    pdf_provider character varying(45),
    pdf_db_key character varying(255),
    pdf_url character varying(2083),
    pdf_checksum character varying(64),
    pdf_id text NOT NULL,
    last_modified date DEFAULT CURRENT_DATE,
    parsers_version text
);


--
-- TOC entry 4045 (class 0 OID 0)
-- Dependencies: 210
-- Name: TABLE case_pdfs; Type: COMMENT; Schema: main; Owner: -
--

COMMENT ON TABLE main.case_pdfs IS '@omit create, update, delete';


--
-- TOC entry 211 (class 1259 OID 17958)
-- Name: case_pdfs_pdf_id_seq; Type: SEQUENCE; Schema: main; Owner: -
--

CREATE SEQUENCE main.case_pdfs_pdf_id_seq
    START WITH 30141
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 212 (class 1259 OID 17960)
-- Name: cases_cited; Type: TABLE; Schema: main; Owner: -
--

CREATE TABLE main.cases_cited (
    citation_count bigint NOT NULL,
    case_origin text NOT NULL,
    case_cited text NOT NULL,
    last_modified date DEFAULT CURRENT_DATE,
    parsers_version text
);


--
-- TOC entry 4046 (class 0 OID 0)
-- Dependencies: 212
-- Name: TABLE cases_cited; Type: COMMENT; Schema: main; Owner: -
--

COMMENT ON TABLE main.cases_cited IS '@omit create, update, delete';


--
-- TOC entry 213 (class 1259 OID 17966)
-- Name: cases_id_seq; Type: SEQUENCE; Schema: main; Owner: -
--

CREATE SEQUENCE main.cases_id_seq
    START WITH 30141
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 214 (class 1259 OID 17968)
-- Name: categories; Type: TABLE; Schema: main; Owner: -
--

CREATE TABLE main.categories (
    category character varying(255) NOT NULL,
    id text NOT NULL,
    last_modified date DEFAULT CURRENT_DATE
);


--
-- TOC entry 4047 (class 0 OID 0)
-- Dependencies: 214
-- Name: TABLE categories; Type: COMMENT; Schema: main; Owner: -
--

COMMENT ON TABLE main.categories IS '@omit create,update,delete';


--
-- TOC entry 215 (class 1259 OID 17974)
-- Name: category_to_cases; Type: TABLE; Schema: main; Owner: -
--

CREATE TABLE main.category_to_cases (
    case_id text,
    category_id text,
    last_modified date DEFAULT CURRENT_DATE,
    parsers_version text
);


--
-- TOC entry 4048 (class 0 OID 0)
-- Dependencies: 215
-- Name: TABLE category_to_cases; Type: COMMENT; Schema: main; Owner: -
--

COMMENT ON TABLE main.category_to_cases IS '@omit create,update,delete';


--
-- TOC entry 216 (class 1259 OID 17980)
-- Name: court_to_cases_id_seq; Type: SEQUENCE; Schema: main; Owner: -
--

CREATE SEQUENCE main.court_to_cases_id_seq
    START WITH 28237
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 217 (class 1259 OID 17982)
-- Name: courts; Type: TABLE; Schema: main; Owner: -
--

CREATE TABLE main.courts (
    name character varying(255),
    acronyms text[],
    id text NOT NULL,
    last_modified date DEFAULT CURRENT_DATE
);


--
-- TOC entry 4049 (class 0 OID 0)
-- Dependencies: 217
-- Name: TABLE courts; Type: COMMENT; Schema: main; Owner: -
--

COMMENT ON TABLE main.courts IS '@omit create, update, delete';


--
-- TOC entry 218 (class 1259 OID 17988)
-- Name: courts_id_seq; Type: SEQUENCE; Schema: main; Owner: -
--

CREATE SEQUENCE main.courts_id_seq
    START WITH 11
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 219 (class 1259 OID 17990)
-- Name: judge_titles; Type: TABLE; Schema: main; Owner: -
--

CREATE TABLE main.judge_titles (
    short_title text,
    id text NOT NULL,
    long_titles text[]
);


--
-- TOC entry 4050 (class 0 OID 0)
-- Dependencies: 219
-- Name: TABLE judge_titles; Type: COMMENT; Schema: main; Owner: -
--

COMMENT ON TABLE main.judge_titles IS '@omit create,update,delete';


--
-- TOC entry 220 (class 1259 OID 17996)
-- Name: judge_titles_id_seq; Type: SEQUENCE; Schema: main; Owner: -
--

CREATE SEQUENCE main.judge_titles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 221 (class 1259 OID 17998)
-- Name: judge_to_cases; Type: TABLE; Schema: main; Owner: -
--

CREATE TABLE main.judge_to_cases (
    title_id text,
    name text,
    case_id text,
    last_modified date DEFAULT CURRENT_DATE,
    parsers_version text
);


--
-- TOC entry 4051 (class 0 OID 0)
-- Dependencies: 221
-- Name: TABLE judge_to_cases; Type: COMMENT; Schema: main; Owner: -
--

COMMENT ON TABLE main.judge_to_cases IS '@omit create,update,delete
@name experimentalJudgeToCases';


--
-- TOC entry 222 (class 1259 OID 18004)
-- Name: judges_id_seq; Type: SEQUENCE; Schema: main; Owner: -
--

CREATE SEQUENCE main.judges_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 223 (class 1259 OID 18006)
-- Name: judges_title_relation_id_seq; Type: SEQUENCE; Schema: main; Owner: -
--

CREATE SEQUENCE main.judges_title_relation_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 224 (class 1259 OID 18008)
-- Name: law_reports; Type: TABLE; Schema: main; Owner: -
--

CREATE TABLE main.law_reports (
    name text,
    id text NOT NULL,
    acronym text
);


--
-- TOC entry 4052 (class 0 OID 0)
-- Dependencies: 224
-- Name: TABLE law_reports; Type: COMMENT; Schema: main; Owner: -
--

COMMENT ON TABLE main.law_reports IS '@omit create,update,delete';


--
-- TOC entry 225 (class 1259 OID 18014)
-- Name: legislation_id_seq; Type: SEQUENCE; Schema: main; Owner: -
--

CREATE SEQUENCE main.legislation_id_seq
    START WITH 1790
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 226 (class 1259 OID 18016)
-- Name: party_and_representative_to_cases; Type: TABLE; Schema: main; Owner: -
--

CREATE TABLE main.party_and_representative_to_cases (
    case_id text,
    party_type main.party_type,
    names text[],
    appearance text[],
    last_modified date DEFAULT CURRENT_DATE,
    parsers_version text
);


--
-- TOC entry 4053 (class 0 OID 0)
-- Dependencies: 226
-- Name: TABLE party_and_representative_to_cases; Type: COMMENT; Schema: main; Owner: -
--

COMMENT ON TABLE main.party_and_representative_to_cases IS '@omit create,update,delete
@name experimentalPartyAndRepresentative';


--
-- TOC entry 203 (class 1259 OID 17909)
-- Name: boolean_facet_values; Type: TABLE; Schema: ugc; Owner: -
--

CREATE TABLE ugc.boolean_facet_values (
    id text NOT NULL,
    metadata_id text,
    value boolean,
    not_applicable boolean,
    unsure boolean
);


--
-- TOC entry 204 (class 1259 OID 17915)
-- Name: date_facet_values; Type: TABLE; Schema: ugc; Owner: -
--

CREATE TABLE ugc.date_facet_values (
    id text NOT NULL,
    metadata_id text,
    date_day integer,
    date_month integer,
    date_year integer,
    not_applicable boolean,
    unsure boolean
);


--
-- TOC entry 205 (class 1259 OID 17921)
-- Name: facet_boolean_keywords; Type: TABLE; Schema: ugc; Owner: -
--

CREATE TABLE ugc.facet_boolean_keywords (
    id text NOT NULL,
    facet_id text,
    value text,
    whole_word boolean
);


--
-- TOC entry 206 (class 1259 OID 17927)
-- Name: facet_value_metadata; Type: TABLE; Schema: ugc; Owner: -
--

CREATE TABLE ugc.facet_value_metadata (
    id text NOT NULL,
    ugc_id text,
    user_id text,
    date_recorded timestamp without time zone,
    case_id text
);


--
-- TOC entry 207 (class 1259 OID 17933)
-- Name: facets; Type: TABLE; Schema: ugc; Owner: -
--

CREATE TABLE ugc.facets (
    id text NOT NULL,
    name character varying(255) NOT NULL,
    type ugc.facet_type NOT NULL,
    description text,
    project_id uuid
);


--
-- TOC entry 227 (class 1259 OID 1132680)
-- Name: projects; Type: TABLE; Schema: ugc; Owner: -
--

CREATE TABLE ugc.projects (
    short_description text NOT NULL,
    terms text,
    cases_stored_proc text NOT NULL,
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text,
    logo text,
    style text
);


--
-- TOC entry 208 (class 1259 OID 17939)
-- Name: random_case_sets; Type: TABLE; Schema: ugc; Owner: -
--

CREATE TABLE ugc.random_case_sets (
    id text NOT NULL,
    case_set json,
    project_id uuid
);


--
-- TOC entry 228 (class 1259 OID 1132691)
-- Name: users_to_projects; Type: TABLE; Schema: ugc; Owner: -
--

CREATE TABLE ugc.users_to_projects (
    user_id text NOT NULL,
    project_id uuid NOT NULL,
    is_owner boolean NOT NULL,
    has_agreed_to_terms boolean
);


--
-- TOC entry 3863 (class 2606 OID 24637)
-- Name: case_citations case_citations_pkey; Type: CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.case_citations
    ADD CONSTRAINT case_citations_pkey PRIMARY KEY (id);


--
-- TOC entry 3866 (class 2606 OID 24639)
-- Name: case_pdfs case_pdfs_pkey; Type: CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.case_pdfs
    ADD CONSTRAINT case_pdfs_pkey PRIMARY KEY (pdf_id);


--
-- TOC entry 3868 (class 2606 OID 24641)
-- Name: cases_cited cases_cited_pkey; Type: CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.cases_cited
    ADD CONSTRAINT cases_cited_pkey PRIMARY KEY (case_origin, case_cited);


--
-- TOC entry 3841 (class 2606 OID 24643)
-- Name: cases cases_pkey; Type: CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.cases
    ADD CONSTRAINT cases_pkey PRIMARY KEY (id);


--
-- TOC entry 3870 (class 2606 OID 24645)
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- TOC entry 3872 (class 2606 OID 24647)
-- Name: courts courts_pkey; Type: CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.courts
    ADD CONSTRAINT courts_pkey PRIMARY KEY (id);


--
-- TOC entry 3874 (class 2606 OID 24649)
-- Name: judge_titles judge_titles_pkey; Type: CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.judge_titles
    ADD CONSTRAINT judge_titles_pkey PRIMARY KEY (id);


--
-- TOC entry 3876 (class 2606 OID 24651)
-- Name: judge_titles judge_titles_short_title_key; Type: CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.judge_titles
    ADD CONSTRAINT judge_titles_short_title_key UNIQUE (short_title);


--
-- TOC entry 3878 (class 2606 OID 24653)
-- Name: law_reports law_reports_pkey; Type: CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.law_reports
    ADD CONSTRAINT law_reports_pkey PRIMARY KEY (id);


--
-- TOC entry 3845 (class 2606 OID 24655)
-- Name: legislation legislation_pkey; Type: CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.legislation
    ADD CONSTRAINT legislation_pkey PRIMARY KEY (id);


--
-- TOC entry 3847 (class 2606 OID 24657)
-- Name: legislation_to_cases legislation_to_cases_pkey; Type: CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.legislation_to_cases
    ADD CONSTRAINT legislation_to_cases_pkey PRIMARY KEY (section, legislation_id, case_id);


--
-- TOC entry 3849 (class 2606 OID 24623)
-- Name: boolean_facet_values boolean_facet_values_pkey; Type: CONSTRAINT; Schema: ugc; Owner: -
--

ALTER TABLE ONLY ugc.boolean_facet_values
    ADD CONSTRAINT boolean_facet_values_pkey PRIMARY KEY (id);


--
-- TOC entry 3851 (class 2606 OID 24625)
-- Name: date_facet_values date_facet_values_pkey; Type: CONSTRAINT; Schema: ugc; Owner: -
--

ALTER TABLE ONLY ugc.date_facet_values
    ADD CONSTRAINT date_facet_values_pkey PRIMARY KEY (id);


--
-- TOC entry 3853 (class 2606 OID 24627)
-- Name: facet_boolean_keywords facet_boolean_keywords_pkey; Type: CONSTRAINT; Schema: ugc; Owner: -
--

ALTER TABLE ONLY ugc.facet_boolean_keywords
    ADD CONSTRAINT facet_boolean_keywords_pkey PRIMARY KEY (id);


--
-- TOC entry 3855 (class 2606 OID 24629)
-- Name: facet_value_metadata facet_value_metadata_pkey; Type: CONSTRAINT; Schema: ugc; Owner: -
--

ALTER TABLE ONLY ugc.facet_value_metadata
    ADD CONSTRAINT facet_value_metadata_pkey PRIMARY KEY (id);


--
-- TOC entry 3857 (class 2606 OID 24631)
-- Name: facets facets_pkey; Type: CONSTRAINT; Schema: ugc; Owner: -
--

ALTER TABLE ONLY ugc.facets
    ADD CONSTRAINT facets_pkey PRIMARY KEY (id);


--
-- TOC entry 3880 (class 2606 OID 1132722)
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: ugc; Owner: -
--

ALTER TABLE ONLY ugc.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);


--
-- TOC entry 3859 (class 2606 OID 24633)
-- Name: random_case_sets random_case_sets_id_key; Type: CONSTRAINT; Schema: ugc; Owner: -
--

ALTER TABLE ONLY ugc.random_case_sets
    ADD CONSTRAINT random_case_sets_id_key UNIQUE (id);


--
-- TOC entry 3861 (class 2606 OID 24635)
-- Name: random_case_sets random_case_sets_pkey; Type: CONSTRAINT; Schema: ugc; Owner: -
--

ALTER TABLE ONLY ugc.random_case_sets
    ADD CONSTRAINT random_case_sets_pkey PRIMARY KEY (id);


--
-- TOC entry 3864 (class 1259 OID 24658)
-- Name: idx_80746_case_citations_index; Type: INDEX; Schema: main; Owner: -
--

CREATE INDEX idx_80746_case_citations_index ON main.case_citations USING gin (to_tsvector('simple'::regconfig, (citation)::text));


--
-- TOC entry 3842 (class 1259 OID 24659)
-- Name: idx_80761_case_text_fulltext_index; Type: INDEX; Schema: main; Owner: -
--

CREATE INDEX idx_80761_case_text_fulltext_index ON main.cases USING gin (to_tsvector('simple'::regconfig, case_text));


--
-- TOC entry 3843 (class 1259 OID 24673)
-- Name: idx_80761_idx_cases_case_text; Type: INDEX; Schema: main; Owner: -
--

CREATE INDEX idx_80761_idx_cases_case_text ON main.cases USING gin (to_tsvector('simple'::regconfig, case_text));


--
-- TOC entry 3900 (class 2620 OID 856029)
-- Name: cases sync_lastmod; Type: TRIGGER; Schema: main; Owner: -
--

CREATE TRIGGER sync_lastmod BEFORE UPDATE ON main.cases FOR EACH ROW EXECUTE PROCEDURE public.sync_lastmod();


--
-- TOC entry 3903 (class 2620 OID 856070)
-- Name: case_citations sync_lastmod_case_citations; Type: TRIGGER; Schema: main; Owner: -
--

CREATE TRIGGER sync_lastmod_case_citations BEFORE UPDATE ON main.case_citations FOR EACH ROW EXECUTE PROCEDURE public.sync_lastmod_case_citations();


--
-- TOC entry 3904 (class 2620 OID 856072)
-- Name: case_pdfs sync_lastmod_case_pdfs; Type: TRIGGER; Schema: main; Owner: -
--

CREATE TRIGGER sync_lastmod_case_pdfs BEFORE UPDATE ON main.case_pdfs FOR EACH ROW EXECUTE PROCEDURE public.sync_lastmod_case_pdfs();


--
-- TOC entry 3905 (class 2620 OID 856074)
-- Name: cases_cited sync_lastmod_cases_cited; Type: TRIGGER; Schema: main; Owner: -
--

CREATE TRIGGER sync_lastmod_cases_cited BEFORE UPDATE ON main.cases_cited FOR EACH ROW EXECUTE PROCEDURE public.sync_lastmod_cases_cited();


--
-- TOC entry 3906 (class 2620 OID 856111)
-- Name: categories sync_lastmod_categories; Type: TRIGGER; Schema: main; Owner: -
--

CREATE TRIGGER sync_lastmod_categories BEFORE UPDATE ON main.categories FOR EACH ROW EXECUTE PROCEDURE public.sync_lastmod_categories();


--
-- TOC entry 3907 (class 2620 OID 856068)
-- Name: category_to_cases sync_lastmod_category_to_cases; Type: TRIGGER; Schema: main; Owner: -
--

CREATE TRIGGER sync_lastmod_category_to_cases BEFORE UPDATE ON main.category_to_cases FOR EACH ROW EXECUTE PROCEDURE public.sync_lastmod_category_to_cases();


--
-- TOC entry 3908 (class 2620 OID 856121)
-- Name: courts sync_lastmod_courts; Type: TRIGGER; Schema: main; Owner: -
--

CREATE TRIGGER sync_lastmod_courts BEFORE UPDATE ON main.courts FOR EACH ROW EXECUTE PROCEDURE public.sync_lastmod_courts();


--
-- TOC entry 3901 (class 2620 OID 856131)
-- Name: legislation sync_lastmod_legislation; Type: TRIGGER; Schema: main; Owner: -
--

CREATE TRIGGER sync_lastmod_legislation BEFORE UPDATE ON main.legislation FOR EACH ROW EXECUTE PROCEDURE public.sync_lastmod_legislation();


--
-- TOC entry 3902 (class 2620 OID 856092)
-- Name: legislation_to_cases sync_lastmod_legislation_to_cases; Type: TRIGGER; Schema: main; Owner: -
--

CREATE TRIGGER sync_lastmod_legislation_to_cases BEFORE UPDATE ON main.legislation_to_cases FOR EACH ROW EXECUTE PROCEDURE public.sync_lastmod_legislation_to_cases();


--
-- TOC entry 3909 (class 2620 OID 856101)
-- Name: party_and_representative_to_cases sync_lastmod_party_and_representative_to_cases; Type: TRIGGER; Schema: main; Owner: -
--

CREATE TRIGGER sync_lastmod_party_and_representative_to_cases BEFORE UPDATE ON main.party_and_representative_to_cases FOR EACH ROW EXECUTE PROCEDURE public.sync_lastmod_party_and_representative_to_cases();


--
-- TOC entry 3892 (class 2606 OID 24689)
-- Name: cases_cited case_cited_fk; Type: FK CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.cases_cited
    ADD CONSTRAINT case_cited_fk FOREIGN KEY (case_cited) REFERENCES main.cases(id);


--
-- TOC entry 3891 (class 2606 OID 24694)
-- Name: case_citations case_id; Type: FK CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.case_citations
    ADD CONSTRAINT case_id FOREIGN KEY (case_id) REFERENCES main.cases(id);


--
-- TOC entry 3894 (class 2606 OID 24699)
-- Name: category_to_cases case_id_fkey; Type: FK CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.category_to_cases
    ADD CONSTRAINT case_id_fkey FOREIGN KEY (case_id) REFERENCES main.cases(id);


--
-- TOC entry 3893 (class 2606 OID 24704)
-- Name: cases_cited case_origin_fk; Type: FK CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.cases_cited
    ADD CONSTRAINT case_origin_fk FOREIGN KEY (case_origin) REFERENCES main.cases(id);


--
-- TOC entry 3895 (class 2606 OID 24709)
-- Name: category_to_cases category_id; Type: FK CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.category_to_cases
    ADD CONSTRAINT category_id FOREIGN KEY (category_id) REFERENCES main.categories(id) DEFERRABLE;


--
-- TOC entry 3881 (class 2606 OID 24714)
-- Name: cases court_id_fkey; Type: FK CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.cases
    ADD CONSTRAINT court_id_fkey FOREIGN KEY (court_id) REFERENCES main.courts(id);


--
-- TOC entry 3896 (class 2606 OID 24719)
-- Name: judge_to_cases judge_to_cases_case_fkey; Type: FK CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.judge_to_cases
    ADD CONSTRAINT judge_to_cases_case_fkey FOREIGN KEY (case_id) REFERENCES main.cases(id);


--
-- TOC entry 3897 (class 2606 OID 24724)
-- Name: judge_to_cases judge_to_cases_title_fkey; Type: FK CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.judge_to_cases
    ADD CONSTRAINT judge_to_cases_title_fkey FOREIGN KEY (title_id) REFERENCES main.judge_titles(id);


--
-- TOC entry 3882 (class 2606 OID 24729)
-- Name: cases law_report_id_fkey; Type: FK CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.cases
    ADD CONSTRAINT law_report_id_fkey FOREIGN KEY (lawreport_id) REFERENCES main.law_reports(id);


--
-- TOC entry 3884 (class 2606 OID 24734)
-- Name: legislation_to_cases legislation_to_cases_cases_fkey; Type: FK CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.legislation_to_cases
    ADD CONSTRAINT legislation_to_cases_cases_fkey FOREIGN KEY (case_id) REFERENCES main.cases(id);


--
-- TOC entry 3885 (class 2606 OID 24739)
-- Name: legislation_to_cases legislation_to_cases_legislation_fkey; Type: FK CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.legislation_to_cases
    ADD CONSTRAINT legislation_to_cases_legislation_fkey FOREIGN KEY (legislation_id) REFERENCES main.legislation(id);


--
-- TOC entry 3898 (class 2606 OID 24744)
-- Name: party_and_representative_to_cases p_and_r_to_cases_cases_fkey; Type: FK CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.party_and_representative_to_cases
    ADD CONSTRAINT p_and_r_to_cases_cases_fkey FOREIGN KEY (case_id) REFERENCES main.cases(id) NOT VALID;


--
-- TOC entry 3883 (class 2606 OID 24749)
-- Name: cases pdf_id; Type: FK CONSTRAINT; Schema: main; Owner: -
--

ALTER TABLE ONLY main.cases
    ADD CONSTRAINT pdf_id FOREIGN KEY (pdf_id) REFERENCES main.case_pdfs(pdf_id);


--
-- TOC entry 3886 (class 2606 OID 24674)
-- Name: boolean_facet_values boolean_facet_values_metadata_id_fkey; Type: FK CONSTRAINT; Schema: ugc; Owner: -
--

ALTER TABLE ONLY ugc.boolean_facet_values
    ADD CONSTRAINT boolean_facet_values_metadata_id_fkey FOREIGN KEY (metadata_id) REFERENCES ugc.facet_value_metadata(id);


--
-- TOC entry 3890 (class 2606 OID 1132741)
-- Name: random_case_sets case_set_to_project_fk; Type: FK CONSTRAINT; Schema: ugc; Owner: -
--

ALTER TABLE ONLY ugc.random_case_sets
    ADD CONSTRAINT case_set_to_project_fk FOREIGN KEY (project_id) REFERENCES ugc.projects(id) NOT VALID;


--
-- TOC entry 3887 (class 2606 OID 24679)
-- Name: date_facet_values date_facet_values_metadata_id_fkey; Type: FK CONSTRAINT; Schema: ugc; Owner: -
--

ALTER TABLE ONLY ugc.date_facet_values
    ADD CONSTRAINT date_facet_values_metadata_id_fkey FOREIGN KEY (metadata_id) REFERENCES ugc.facet_value_metadata(id);


--
-- TOC entry 3888 (class 2606 OID 24684)
-- Name: facet_boolean_keywords facet_boolean_keywords_facet_id_fkey; Type: FK CONSTRAINT; Schema: ugc; Owner: -
--

ALTER TABLE ONLY ugc.facet_boolean_keywords
    ADD CONSTRAINT facet_boolean_keywords_facet_id_fkey FOREIGN KEY (facet_id) REFERENCES ugc.facets(id);


--
-- TOC entry 3889 (class 2606 OID 1132747)
-- Name: facets facets_to_project_fk; Type: FK CONSTRAINT; Schema: ugc; Owner: -
--

ALTER TABLE ONLY ugc.facets
    ADD CONSTRAINT facets_to_project_fk FOREIGN KEY (project_id) REFERENCES ugc.projects(id) NOT VALID;


--
-- TOC entry 3899 (class 2606 OID 1132723)
-- Name: users_to_projects users_to_projects_fk; Type: FK CONSTRAINT; Schema: ugc; Owner: -
--

ALTER TABLE ONLY ugc.users_to_projects
    ADD CONSTRAINT users_to_projects_fk FOREIGN KEY (project_id) REFERENCES ugc.projects(id) NOT VALID;


-- Completed on 2021-01-08 11:23:03

--
-- PostgreSQL database dump complete
--

