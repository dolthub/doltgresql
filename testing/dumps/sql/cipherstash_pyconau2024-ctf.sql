-- Downloaded from: https://github.com/cipherstash/pyconau2024-ctf/blob/eefb622852c0b3282092dd7520c17f04f046a6d5/db/postgres.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 16.2 (Debian 16.2-1.pgdg120+2)
-- Dumped by pg_dump version 17.0

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
-- Name: _cs_config_check_cast(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_config_check_cast(val jsonb) RETURNS boolean
    LANGUAGE plpgsql
    AS $_$
	BEGIN
    IF EXISTS (SELECT jsonb_array_elements_text(jsonb_path_query_array(val, '$.tables.*.*.cast_as')) = ANY('{text, int, small_int, big_int, real, double, boolean, date, jsonb}')) THEN
      RETURN true;
    END IF;
    RAISE 'Configuration has an invalid cast_as (%). Cast should be one of {text, int, small_int, big_int, real, double, boolean, date, jsonb}', val;
  END;
$_$;


ALTER FUNCTION public._cs_config_check_cast(val jsonb) OWNER TO postgres;

--
-- Name: _cs_config_check_indexes(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_config_check_indexes(val jsonb) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
	BEGIN
    IF (SELECT EXISTS (SELECT _cs_extract_indexes(val)))  THEN
      IF (SELECT bool_and(index = ANY('{match, ore, unique, ste_vec}')) FROM _cs_extract_indexes(val) AS index) THEN
        RETURN true;
      END IF;
      RAISE 'Configuration has an invalid index (%). Index should be one of {match, ore, unique, ste_vec}', val;
    END IF;
    RETURN true;
  END;
$$;


ALTER FUNCTION public._cs_config_check_indexes(val jsonb) OWNER TO postgres;

--
-- Name: _cs_config_check_tables(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_config_check_tables(val jsonb) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
	BEGIN
    IF (val ? 'tables') AND (val->'tables' <> '{}'::jsonb) THEN
      RETURN true;
    END IF;
    RAISE 'Configuration missing tables (tables) field: %', val;
  END;
$$;


ALTER FUNCTION public._cs_config_check_tables(val jsonb) OWNER TO postgres;

--
-- Name: _cs_config_check_v(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_config_check_v(val jsonb) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
	BEGIN
    IF (val ? 'v') THEN
      RETURN true;
    END IF;
    RAISE 'Configuration missing version (v) field: %', val;
  END;
$$;


ALTER FUNCTION public._cs_config_check_v(val jsonb) OWNER TO postgres;

--
-- Name: cs_configuration_data_v1; Type: DOMAIN; Schema: public; Owner: postgres
--

CREATE DOMAIN public.cs_configuration_data_v1 AS jsonb
	CONSTRAINT cs_configuration_data_v1_check CHECK ((public._cs_config_check_v(VALUE) AND public._cs_config_check_tables(VALUE) AND public._cs_config_check_cast(VALUE) AND public._cs_config_check_indexes(VALUE)));


ALTER DOMAIN public.cs_configuration_data_v1 OWNER TO postgres;

--
-- Name: cs_configuration_state_v1; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.cs_configuration_state_v1 AS ENUM (
    'active',
    'inactive',
    'encrypting',
    'pending'
);


ALTER TYPE public.cs_configuration_state_v1 OWNER TO postgres;

--
-- Name: _cs_encrypted_check_i(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_encrypted_check_i(val jsonb) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
	BEGIN
    IF val ? 'i' THEN
      RETURN true;
    END IF;
    RAISE 'Encrypted column missing ident (i) field: %', val;
  END;
$$;


ALTER FUNCTION public._cs_encrypted_check_i(val jsonb) OWNER TO postgres;

--
-- Name: _cs_encrypted_check_k(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_encrypted_check_k(val jsonb) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
	BEGIN
    IF (val->>'k' = ANY('{ct, sv}')) THEN
      RETURN true;
    END IF;
    RAISE 'Invalid kind (%) in Encrypted column. Kind should be one of {ct, sv}', val;
  END;
$$;


ALTER FUNCTION public._cs_encrypted_check_k(val jsonb) OWNER TO postgres;

--
-- Name: _cs_encrypted_check_k_ct(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_encrypted_check_k_ct(val jsonb) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
	BEGIN
    IF (val->>'k' = 'ct') THEN
      IF (val ? 'c') THEN
        RETURN true;
      END IF;
      RAISE 'Encrypted column kind (k) of "ct" missing data field (c):  %', val;
    END IF;
    RETURN true;
  END;
$$;


ALTER FUNCTION public._cs_encrypted_check_k_ct(val jsonb) OWNER TO postgres;

--
-- Name: _cs_encrypted_check_k_sv(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_encrypted_check_k_sv(val jsonb) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
	BEGIN
    IF (val->>'k' = 'sv') THEN
      IF (val ? 'sv') THEN
        RETURN true;
      END IF;
      RAISE 'Encrypted column kind (k) of "sv" missing data field (sv):  %', val;
    END IF;
    RETURN true;
  END;
$$;


ALTER FUNCTION public._cs_encrypted_check_k_sv(val jsonb) OWNER TO postgres;

--
-- Name: _cs_encrypted_check_p(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_encrypted_check_p(val jsonb) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
	BEGIN
    IF NOT val ? 'p' THEN
      RETURN true;
    END IF;
    RAISE 'Encrypted column includes plaintext (p) field: %', val;
  END;
$$;


ALTER FUNCTION public._cs_encrypted_check_p(val jsonb) OWNER TO postgres;

--
-- Name: _cs_encrypted_check_q(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_encrypted_check_q(val jsonb) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
	BEGIN
    IF val ? 'q' THEN
      RAISE 'Encrypted column includes query (q) field: %', val;
    END IF;
    RETURN true;
  END;
$$;


ALTER FUNCTION public._cs_encrypted_check_q(val jsonb) OWNER TO postgres;

--
-- Name: _cs_encrypted_check_v(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_encrypted_check_v(val jsonb) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
	BEGIN
    IF (val ? 'v') THEN
      RETURN true;
    END IF;
    RAISE 'Encrypted column missing version (v) field: %', val;
  END;
$$;


ALTER FUNCTION public._cs_encrypted_check_v(val jsonb) OWNER TO postgres;

--
-- Name: cs_check_encrypted_v1(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_check_encrypted_v1(val jsonb) RETURNS boolean
    LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE
    BEGIN ATOMIC
 RETURN (public._cs_encrypted_check_v(val) AND public._cs_encrypted_check_i(val) AND public._cs_encrypted_check_k(val) AND public._cs_encrypted_check_k_ct(val) AND public._cs_encrypted_check_k_sv(val) AND public._cs_encrypted_check_q(val) AND public._cs_encrypted_check_p(val));
END;


ALTER FUNCTION public.cs_check_encrypted_v1(val jsonb) OWNER TO postgres;

--
-- Name: cs_encrypted_v1; Type: DOMAIN; Schema: public; Owner: postgres
--

CREATE DOMAIN public.cs_encrypted_v1 AS jsonb
	CONSTRAINT cs_encrypted_v1_check CHECK (public.cs_check_encrypted_v1(VALUE));


ALTER DOMAIN public.cs_encrypted_v1 OWNER TO postgres;

--
-- Name: cs_match_index_v1; Type: DOMAIN; Schema: public; Owner: postgres
--

CREATE DOMAIN public.cs_match_index_v1 AS smallint[];


ALTER DOMAIN public.cs_match_index_v1 OWNER TO postgres;

--
-- Name: cs_ste_vec_encrypted_term_v1; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.cs_ste_vec_encrypted_term_v1 AS (
	bytes bytea
);


ALTER TYPE public.cs_ste_vec_encrypted_term_v1 OWNER TO postgres;

--
-- Name: cs_ste_vec_v1_entry; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.cs_ste_vec_v1_entry AS (
	tokenized_selector text,
	term public.cs_ste_vec_encrypted_term_v1,
	ciphertext text
);


ALTER TYPE public.cs_ste_vec_v1_entry OWNER TO postgres;

--
-- Name: cs_ste_vec_index_v1; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.cs_ste_vec_index_v1 AS (
	entries public.cs_ste_vec_v1_entry[]
);


ALTER TYPE public.cs_ste_vec_index_v1 OWNER TO postgres;

--
-- Name: cs_unique_index_v1; Type: DOMAIN; Schema: public; Owner: postgres
--

CREATE DOMAIN public.cs_unique_index_v1 AS text;


ALTER DOMAIN public.cs_unique_index_v1 OWNER TO postgres;

--
-- Name: ore_64_8_v1_term; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.ore_64_8_v1_term AS (
	bytes bytea
);


ALTER TYPE public.ore_64_8_v1_term OWNER TO postgres;

--
-- Name: ore_64_8_v1; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.ore_64_8_v1 AS (
	terms public.ore_64_8_v1_term[]
);


ALTER TYPE public.ore_64_8_v1 OWNER TO postgres;

--
-- Name: ore_cllw_8_v1; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.ore_cllw_8_v1 AS (
	bytes bytea
);


ALTER TYPE public.ore_cllw_8_v1 OWNER TO postgres;

--
-- Name: ore_cllw_8_variable_v1; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.ore_cllw_8_variable_v1 AS (
	bytes bytea
);


ALTER TYPE public.ore_cllw_8_variable_v1 OWNER TO postgres;

--
-- Name: jsonb_to_cs_ste_vec_index_v1(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.jsonb_to_cs_ste_vec_index_v1(input jsonb) RETURNS public.cs_ste_vec_index_v1
    LANGUAGE plpgsql
    AS $$
DECLARE
    vec_entry cs_ste_vec_v1_entry;
    entry_array cs_ste_vec_v1_entry[];
    entry_json jsonb;
    entry_json_array jsonb[];
    entry_array_length int;
    i int;
BEGIN
    FOR entry_json IN SELECT * FROM jsonb_array_elements(input)
    LOOP
        vec_entry := ROW(
           entry_json->>0,
           ROW(decode(entry_json->>1, 'hex'))::cs_ste_vec_encrypted_term_v1,
           entry_json->>2
        )::cs_ste_vec_v1_entry;
        entry_array := array_append(entry_array, vec_entry);
    END LOOP;

    RETURN ROW(entry_array)::cs_ste_vec_index_v1;
END;
$$;


ALTER FUNCTION public.jsonb_to_cs_ste_vec_index_v1(input jsonb) OWNER TO postgres;

--
-- Name: CAST (jsonb AS public.cs_ste_vec_index_v1); Type: CAST; Schema: -; Owner: -
--

CREATE CAST (jsonb AS public.cs_ste_vec_index_v1) WITH FUNCTION public.jsonb_to_cs_ste_vec_index_v1(jsonb) AS IMPLICIT;


--
-- Name: _cs_text_to_ore_64_8_v1_term_v1_0(text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_text_to_ore_64_8_v1_term_v1_0(t text) RETURNS public.ore_64_8_v1_term
    LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE
    BEGIN ATOMIC
 RETURN (t)::bytea;
END;


ALTER FUNCTION public._cs_text_to_ore_64_8_v1_term_v1_0(t text) OWNER TO postgres;

--
-- Name: CAST (text AS public.ore_64_8_v1_term); Type: CAST; Schema: -; Owner: -
--

CREATE CAST (text AS public.ore_64_8_v1_term) WITH FUNCTION public._cs_text_to_ore_64_8_v1_term_v1_0(text) AS IMPLICIT;


--
-- Name: __bytea_ct_eq(bytea, bytea); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.__bytea_ct_eq(a bytea, b bytea) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
    result boolean;
    differing bytea;
BEGIN
    -- Check if the bytea values are the same length
    IF LENGTH(a) != LENGTH(b) THEN
        RETURN false;
    END IF;

    -- Compare each byte in the bytea values
    result := true;
    FOR i IN 1..LENGTH(a) LOOP
        IF SUBSTRING(a FROM i FOR 1) != SUBSTRING(b FROM i FOR 1) THEN
            result := result AND false;
        END IF;
    END LOOP;

    RETURN result;
END;
$$;


ALTER FUNCTION public.__bytea_ct_eq(a bytea, b bytea) OWNER TO postgres;

--
-- Name: __compare_inner_ore_cllw_8_v1(bytea, bytea); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.__compare_inner_ore_cllw_8_v1(a bytea, b bytea) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    len_a INT;
    x BYTEA;
    y BYTEA;
    i INT;
    differing RECORD;
BEGIN
    len_a := LENGTH(a);

    -- Iterate over each byte and compare them
    FOR i IN 1..len_a LOOP
        x := SUBSTRING(a FROM i FOR 1);
        y := SUBSTRING(b FROM i FOR 1);

        -- Check if there's a difference
        IF x != y THEN
            differing := (x, y);
            EXIT;
        END IF;
    END LOOP;

    -- If a difference is found, compare the bytes as in Rust logic
    IF differing IS NOT NULL THEN
        IF (get_byte(y, 0) + 1) % 256 = get_byte(x, 0) THEN
            RETURN 1;
        ELSE
            RETURN -1;
        END IF;
    ELSE
        RETURN 0;
    END IF;
END;
$$;


ALTER FUNCTION public.__compare_inner_ore_cllw_8_v1(a bytea, b bytea) OWNER TO postgres;

--
-- Name: _cs_config_add_cast(text, text, text, jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_config_add_cast(table_name text, column_name text, cast_as text, config jsonb) RETURNS jsonb
    LANGUAGE plpgsql IMMUTABLE PARALLEL SAFE
    AS $$
  BEGIN
    SELECT jsonb_set(config, array['tables', table_name, column_name, 'cast_as'], to_jsonb(cast_as)) INTO config;
    RETURN config;
  END;
$$;


ALTER FUNCTION public._cs_config_add_cast(table_name text, column_name text, cast_as text, config jsonb) OWNER TO postgres;

--
-- Name: _cs_config_add_column(text, text, jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_config_add_column(table_name text, column_name text, config jsonb) RETURNS jsonb
    LANGUAGE plpgsql IMMUTABLE PARALLEL SAFE
    AS $$
  DECLARE
    col jsonb;
  BEGIN
    IF NOT config #> array['tables', table_name] ? column_name THEN
      SELECT jsonb_build_object('indexes', jsonb_build_object()) into col;
      SELECT jsonb_set(config, array['tables', table_name, column_name], col) INTO config;
    END IF;
    RETURN config;
  END;
$$;


ALTER FUNCTION public._cs_config_add_column(table_name text, column_name text, config jsonb) OWNER TO postgres;

--
-- Name: _cs_config_add_index(text, text, text, jsonb, jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_config_add_index(table_name text, column_name text, index_name text, opts jsonb, config jsonb) RETURNS jsonb
    LANGUAGE plpgsql IMMUTABLE PARALLEL SAFE
    AS $$
  BEGIN
    SELECT jsonb_insert(config, array['tables', table_name, column_name, 'indexes', index_name], opts) INTO config;
    RETURN config;
  END;
$$;


ALTER FUNCTION public._cs_config_add_index(table_name text, column_name text, index_name text, opts jsonb, config jsonb) OWNER TO postgres;

--
-- Name: _cs_config_add_table(text, jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_config_add_table(table_name text, config jsonb) RETURNS jsonb
    LANGUAGE plpgsql
    AS $$
  DECLARE
    tbl jsonb;
  BEGIN
    IF NOT config #> array['tables'] ? table_name THEN
      SELECT jsonb_insert(config, array['tables', table_name], jsonb_build_object()) INTO config;
    END IF;
    RETURN config;
  END;
$$;


ALTER FUNCTION public._cs_config_add_table(table_name text, config jsonb) OWNER TO postgres;

--
-- Name: _cs_config_default(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_config_default(config jsonb) RETURNS jsonb
    LANGUAGE plpgsql IMMUTABLE PARALLEL SAFE
    AS $$
  BEGIN
    IF config IS NULL THEN
      SELECT jsonb_build_object('v', 1, 'tables', jsonb_build_object()) INTO config;
    END IF;
    RETURN config;
  END;
$$;


ALTER FUNCTION public._cs_config_default(config jsonb) OWNER TO postgres;

--
-- Name: _cs_config_match_default(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_config_match_default() RETURNS jsonb
    LANGUAGE sql STRICT PARALLEL SAFE
    BEGIN ATOMIC
 SELECT jsonb_build_object('k', 6, 'm', 2048, 'include_original', true, 'tokenizer', json_build_object('kind', 'ngram', 'token_length', 3), 'token_filters', json_build_array(json_build_object('kind', 'downcase'))) AS jsonb_build_object;
END;


ALTER FUNCTION public._cs_config_match_default() OWNER TO postgres;

--
-- Name: _cs_diff_config_v1(jsonb, jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_diff_config_v1(a jsonb, b jsonb) RETURNS TABLE(table_name text, column_name text)
    LANGUAGE plpgsql IMMUTABLE STRICT PARALLEL SAFE
    AS $$
  BEGIN
    RETURN QUERY
    WITH table_keys AS (
      SELECT jsonb_object_keys(a->'tables') AS key
      UNION
      SELECT jsonb_object_keys(b->'tables') AS key
    ),
    column_keys AS (
      SELECT tk.key AS table_key, jsonb_object_keys(a->'tables'->tk.key) AS column_key
      FROM table_keys tk
      UNION
      SELECT tk.key AS table_key, jsonb_object_keys(b->'tables'->tk.key) AS column_key
      FROM table_keys tk
    )
    SELECT
      ck.table_key AS table_name,
      ck.column_key AS column_name
    FROM
      column_keys ck
    WHERE
      (a->'tables'->ck.table_key->ck.column_key IS DISTINCT FROM b->'tables'->ck.table_key->ck.column_key);
  END;
$$;


ALTER FUNCTION public._cs_diff_config_v1(a jsonb, b jsonb) OWNER TO postgres;

--
-- Name: _cs_encrypted_check_i_ct(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_encrypted_check_i_ct(val jsonb) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
	BEGIN
    IF (val->'i' ?& array['t', 'c']) THEN
      RETURN true;
    END IF;
    RAISE 'Encrypted column ident (i) missing table (t) or column (c) fields: %', val;
  END;
$$;


ALTER FUNCTION public._cs_encrypted_check_i_ct(val jsonb) OWNER TO postgres;

--
-- Name: _cs_extract_indexes(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_extract_indexes(val jsonb) RETURNS SETOF text
    LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE
    BEGIN ATOMIC
 SELECT jsonb_object_keys(jsonb_path_query(val, '$."tables".*.*."indexes"'::jsonpath)) AS jsonb_object_keys;
END;


ALTER FUNCTION public._cs_extract_indexes(val jsonb) OWNER TO postgres;

--
-- Name: _cs_first_grouped_value(jsonb, jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public._cs_first_grouped_value(jsonb, jsonb) RETURNS jsonb
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT COALESCE($1, $2);
$_$;


ALTER FUNCTION public._cs_first_grouped_value(jsonb, jsonb) OWNER TO postgres;

--
-- Name: compare_lex_ore_cllw_8_v1(public.ore_cllw_8_variable_v1, public.ore_cllw_8_variable_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.compare_lex_ore_cllw_8_v1(a public.ore_cllw_8_variable_v1, b public.ore_cllw_8_variable_v1) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    len_a INT;
    len_b INT;
    -- length of the common part of the two bytea values
    common_len INT;
    cmp_result INT;
BEGIN
    -- Get the lengths of both bytea inputs
    len_a := LENGTH(a.bytes);
    len_b := LENGTH(b.bytes);

    -- Handle empty cases
    IF len_a = 0 AND len_b = 0 THEN
        RETURN 0;
    ELSIF len_a = 0 THEN
        RETURN -1;
    ELSIF len_b = 0 THEN
        RETURN 1;
    END IF;

    -- Find the length of the shorter bytea
    IF len_a < len_b THEN
        common_len := len_a;
    ELSE
        common_len := len_b;
    END IF;

    -- Use the compare_bytea function to compare byte by byte
    cmp_result := __compare_inner_ore_cllw_8_v1(
      SUBSTRING(a.bytes FROM 1 FOR common_len),
      SUBSTRING(b.bytes FROM 1 FOR common_len)
    );

    -- If the comparison returns 'less' or 'greater', return that result
    IF cmp_result = -1 THEN
        RETURN -1;
    ELSIF cmp_result = 1 THEN
        RETURN 1;
    END IF;

    -- If the bytea comparison is 'equal', compare lengths
    IF len_a < len_b THEN
        RETURN -1;
    ELSIF len_a > len_b THEN
        RETURN 1;
    ELSE
        RETURN 0;
    END IF;
END;
$$;


ALTER FUNCTION public.compare_lex_ore_cllw_8_v1(a public.ore_cllw_8_variable_v1, b public.ore_cllw_8_variable_v1) OWNER TO postgres;

--
-- Name: compare_ore_64_8_v1(public.ore_64_8_v1, public.ore_64_8_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.compare_ore_64_8_v1(a public.ore_64_8_v1, b public.ore_64_8_v1) RETURNS integer
    LANGUAGE plpgsql
    AS $$
  DECLARE
    cmp_result integer;
  BEGIN
    -- Recursively compare blocks bailing as soon as we can make a decision
    RETURN compare_ore_array(a.terms, b.terms);
  END
$$;


ALTER FUNCTION public.compare_ore_64_8_v1(a public.ore_64_8_v1, b public.ore_64_8_v1) OWNER TO postgres;

--
-- Name: compare_ore_64_8_v1_term(public.ore_64_8_v1_term, public.ore_64_8_v1_term); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.compare_ore_64_8_v1_term(a public.ore_64_8_v1_term, b public.ore_64_8_v1_term) RETURNS integer
    LANGUAGE plpgsql
    AS $$
  DECLARE
    eq boolean := true;
    unequal_block smallint := 0;
    hash_key bytea;
    target_block bytea;

    left_block_size CONSTANT smallint := 16;
    right_block_size CONSTANT smallint := 32;
    right_offset CONSTANT smallint := 136; -- 8 * 17

    indicator smallint := 0;
  BEGIN
    IF a IS NULL AND b IS NULL THEN
      RETURN 0;
    END IF;

    IF a IS NULL THEN
      RETURN -1;
    END IF;

    IF b IS NULL THEN
      RETURN 1;
    END IF;

    IF bit_length(a.bytes) != bit_length(b.bytes) THEN
      RAISE EXCEPTION 'Ciphertexts are different lengths';
    END IF;

    FOR block IN 0..7 LOOP
      -- Compare each PRP (byte from the first 8 bytes) and PRF block (8 byte
      -- chunks of the rest of the value).
      -- NOTE:
      -- * Substr is ordinally indexed (hence 1 and not 0, and 9 and not 8).
      -- * We are not worrying about timing attacks here; don't fret about
      --   the OR or !=.
      IF
        substr(a.bytes, 1 + block, 1) != substr(b.bytes, 1 + block, 1)
        OR substr(a.bytes, 9 + left_block_size * block, left_block_size) != substr(b.bytes, 9 + left_block_size * BLOCK, left_block_size)
      THEN
        -- set the first unequal block we find
        IF eq THEN
          unequal_block := block;
        END IF;
        eq = false;
      END IF;
    END LOOP;

    IF eq THEN
      RETURN 0::integer;
    END IF;

    -- Hash key is the IV from the right CT of b
    hash_key := substr(b.bytes, right_offset + 1, 16);

    -- first right block is at right offset + nonce_size (ordinally indexed)
    target_block := substr(b.bytes, right_offset + 17 + (unequal_block * right_block_size), right_block_size);

    indicator := (
      get_bit(
        encrypt(
          substr(a.bytes, 9 + (left_block_size * unequal_block), left_block_size),
          hash_key,
          'aes-ecb'
        ),
        0
      ) + get_bit(target_block, get_byte(a.bytes, unequal_block))) % 2;

    IF indicator = 1 THEN
      RETURN 1::integer;
    ELSE
      RETURN -1::integer;
    END IF;
  END;
$$;


ALTER FUNCTION public.compare_ore_64_8_v1_term(a public.ore_64_8_v1_term, b public.ore_64_8_v1_term) OWNER TO postgres;

--
-- Name: compare_ore_array(public.ore_64_8_v1_term[], public.ore_64_8_v1_term[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.compare_ore_array(a public.ore_64_8_v1_term[], b public.ore_64_8_v1_term[]) RETURNS integer
    LANGUAGE plpgsql
    AS $$
  DECLARE
    cmp_result integer;
  BEGIN
    IF (array_length(a, 1) = 0 OR a IS NULL) AND (array_length(b, 1) = 0 OR b IS NULL) THEN
      RETURN 0;
    END IF;
    IF array_length(a, 1) = 0 OR a IS NULL THEN
      RETURN -1;
    END IF;
    IF array_length(b, 1) = 0 OR a IS NULL THEN
      RETURN 1;
    END IF;

    cmp_result := compare_ore_64_8_v1_term(a[1], b[1]);
    IF cmp_result = 0 THEN
    -- Removes the first element in the array, and calls this fn again to compare the next element/s in the array.
      RETURN compare_ore_array(a[2:array_length(a,1)], b[2:array_length(b,1)]);
    END IF;

    RETURN cmp_result;
  END
$$;


ALTER FUNCTION public.compare_ore_array(a public.ore_64_8_v1_term[], b public.ore_64_8_v1_term[]) OWNER TO postgres;

--
-- Name: compare_ore_cllw_8_v1(public.ore_cllw_8_v1, public.ore_cllw_8_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.compare_ore_cllw_8_v1(a public.ore_cllw_8_v1, b public.ore_cllw_8_v1) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    len_a INT;
    len_b INT;
    x BYTEA;
    y BYTEA;
    i INT;
    differing RECORD;
BEGIN
    -- Check if the lengths of the two bytea arguments are the same
    len_a := LENGTH(a.bytes);
    len_b := LENGTH(b.bytes);

    IF len_a != len_b THEN
      RAISE EXCEPTION 'Numeric ORE comparison requires bytea values of the same length';
    END IF;

    RETURN __compare_inner_ore_cllw_8_v1(a.bytes, b.bytes);
END;
$$;


ALTER FUNCTION public.compare_ore_cllw_8_v1(a public.ore_cllw_8_v1, b public.ore_cllw_8_v1) OWNER TO postgres;

--
-- Name: compare_ste_vec_encrypted_term_v1(public.cs_ste_vec_encrypted_term_v1, public.cs_ste_vec_encrypted_term_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.compare_ste_vec_encrypted_term_v1(a public.cs_ste_vec_encrypted_term_v1, b public.cs_ste_vec_encrypted_term_v1) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
  header_a INT;
  header_b INT;
  body_a BYTEA;
  body_b BYTEA;
BEGIN
  -- `get_byte` is 0-indexed
  header_a := get_byte(a.bytes, 0);
  header_b := get_byte(b.bytes, 0);

  IF header_a != header_b THEN
    RAISE EXCEPTION 'compare_ste_vec_encrypted_term_v1: expected equal header bytes';
  END IF;

  -- `substr` is 1-indexed (yes, `subtr` starts at 1 and `get_byte` starts at 0).
  body_a := substr(a.bytes, 2);
  body_b := substr(b.bytes, 2);

  CASE header_a
    WHEN 0 THEN
      RAISE EXCEPTION 'compare_ste_vec_encrypted_term_v1: can not compare MAC terms';
    WHEN 1 THEN
      RETURN compare_ore_cllw_8_v1(ROW(body_a)::ore_cllw_8_v1, ROW(body_b)::ore_cllw_8_v1);
    WHEN 2 THEN
      RETURN compare_lex_ore_cllw_8_v1(ROW(body_a)::ore_cllw_8_variable_v1, ROW(body_b)::ore_cllw_8_variable_v1);
    ELSE
      RAISE EXCEPTION 'compare_ste_vec_encrypted_term_v1: invalid header for cs_ste_vec_encrypted_term_v1: header "%", body "%', header_a, body_a;
  END CASE;
END;
$$;


ALTER FUNCTION public.compare_ste_vec_encrypted_term_v1(a public.cs_ste_vec_encrypted_term_v1, b public.cs_ste_vec_encrypted_term_v1) OWNER TO postgres;

--
-- Name: cs_activate_v1(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_activate_v1() RETURNS boolean
    LANGUAGE plpgsql
    AS $$
	BEGIN

	  IF EXISTS (SELECT FROM cs_configuration_v1 c WHERE c.state = 'encrypting') THEN
	  	UPDATE cs_configuration_v1 SET state = 'inactive' WHERE state = 'active';
			UPDATE cs_configuration_v1 SET state = 'active' WHERE state = 'encrypting';
			RETURN true;
		ELSE
			RAISE EXCEPTION 'No encrypting configuration exists to activate';
		END IF;
  END;
$$;


ALTER FUNCTION public.cs_activate_v1() OWNER TO postgres;

--
-- Name: cs_add_column_v1(text, text, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_add_column_v1(table_name text, column_name text, cast_as text DEFAULT 'text'::text) RETURNS jsonb
    LANGUAGE plpgsql
    AS $$
  DECLARE
    key text;
    _config jsonb;
  BEGIN
    -- set the active config
    SELECT data INTO _config FROM cs_configuration_v1 WHERE state = 'active' OR state = 'pending' ORDER BY state DESC;

    -- set default config
    SELECT _cs_config_default(_config) INTO _config;

    -- if index exists
    IF _config #> array['tables', table_name] ?  column_name THEN
      RAISE EXCEPTION 'Config exists for column: % %', table_name, column_name;
    END IF;

    SELECT _cs_config_add_table(table_name, _config) INTO _config;

    SELECT _cs_config_add_column(table_name, column_name, _config) INTO _config;

    SELECT _cs_config_add_cast(table_name, column_name, cast_as, _config) INTO _config;

    --  create a new pending record if we don't have one
    INSERT INTO cs_configuration_v1 (state, data) VALUES ('pending', _config)
    ON CONFLICT (state)
      WHERE state = 'pending'
    DO UPDATE
      SET data = _config;

    -- exeunt
    RETURN _config;
  END;
$$;


ALTER FUNCTION public.cs_add_column_v1(table_name text, column_name text, cast_as text) OWNER TO postgres;

--
-- Name: cs_add_index_v1(text, text, text, text, jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_add_index_v1(table_name text, column_name text, index_name text, cast_as text DEFAULT 'text'::text, opts jsonb DEFAULT '{}'::jsonb) RETURNS jsonb
    LANGUAGE plpgsql
    AS $$
  DECLARE
    o jsonb;
    _config jsonb;
  BEGIN

    -- set the active config
    SELECT data INTO _config FROM cs_configuration_v1 WHERE state = 'active' OR state = 'pending' ORDER BY state DESC;

    -- if index exists
    IF _config #> array['tables', table_name, column_name, 'indexes'] ?  index_name THEN
      RAISE EXCEPTION '% index exists for column: % %', index_name, table_name, column_name;
    END IF;

    IF NOT cast_as = ANY('{text, int, small_int, big_int, real, double, boolean, date, jsonb}') THEN
      RAISE EXCEPTION '% is not a valid cast type', cast_as;
    END IF;

    -- set default config
    SELECT _cs_config_default(_config) INTO _config;

    SELECT _cs_config_add_table(table_name, _config) INTO _config;

    SELECT _cs_config_add_column(table_name, column_name, _config) INTO _config;

    SELECT _cs_config_add_cast(table_name, column_name, cast_as, _config) INTO _config;

    -- set default options for index if opts empty
    IF index_name = 'match' AND opts = '{}' THEN
      SELECT _cs_config_match_default() INTO opts;
    END IF;

    SELECT _cs_config_add_index(table_name, column_name, index_name, opts, _config) INTO _config;

    --  create a new pending record if we don't have one
    INSERT INTO cs_configuration_v1 (state, data) VALUES ('pending', _config)
    ON CONFLICT (state)
      WHERE state = 'pending'
    DO UPDATE
      SET data = _config;

    -- exeunt
    RETURN _config;
  END;
$$;


ALTER FUNCTION public.cs_add_index_v1(table_name text, column_name text, index_name text, cast_as text, opts jsonb) OWNER TO postgres;

--
-- Name: cs_ciphertext_v1_v0_0(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ciphertext_v1_v0_0(val jsonb) RETURNS text
    LANGUAGE plpgsql IMMUTABLE STRICT PARALLEL SAFE
    AS $$
	BEGIN
    IF val ? 'c' THEN
      RETURN val->>'c';
    END IF;
    RAISE 'Expected a ciphertext (c) value in json: %', val;
  END;
$$;


ALTER FUNCTION public.cs_ciphertext_v1_v0_0(val jsonb) OWNER TO postgres;

--
-- Name: cs_ciphertext_v1(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ciphertext_v1(val jsonb) RETURNS text
    LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE
    BEGIN ATOMIC
 RETURN public.cs_ciphertext_v1_v0_0(val);
END;


ALTER FUNCTION public.cs_ciphertext_v1(val jsonb) OWNER TO postgres;

--
-- Name: cs_ciphertext_v1_v0(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ciphertext_v1_v0(val jsonb) RETURNS text
    LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE
    BEGIN ATOMIC
 RETURN public.cs_ciphertext_v1_v0_0(val);
END;


ALTER FUNCTION public.cs_ciphertext_v1_v0(val jsonb) OWNER TO postgres;

--
-- Name: cs_count_encrypted_with_active_config_v1(text, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_count_encrypted_with_active_config_v1(table_name text, column_name text) RETURNS bigint
    LANGUAGE plpgsql
    AS $$
DECLARE
  result BIGINT;
BEGIN
	EXECUTE format(
        'SELECT COUNT(%I) FROM %s t WHERE %I->>%L = (SELECT id::TEXT FROM cs_configuration_v1 WHERE state = %L)',
        column_name, table_name, column_name, 'v', 'active'
    )
	INTO result;
  	RETURN result;
END;
$$;


ALTER FUNCTION public.cs_count_encrypted_with_active_config_v1(table_name text, column_name text) OWNER TO postgres;

--
-- Name: cs_create_encrypted_columns_v1(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_create_encrypted_columns_v1() RETURNS TABLE(table_name text, column_name text)
    LANGUAGE plpgsql
    AS $$
	BEGIN
    FOR table_name, column_name IN
      SELECT c.table_name, (c.column_name || '_encrypted') FROM cs_select_target_columns_v1() AS c WHERE c.target_column IS NULL
    LOOP
		  EXECUTE format('ALTER TABLE %I ADD column %I cs_encrypted_v1', table_name, column_name);
      RETURN NEXT;
    END LOOP;
	END;
$$;


ALTER FUNCTION public.cs_create_encrypted_columns_v1() OWNER TO postgres;

--
-- Name: cs_discard_v1(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_discard_v1() RETURNS boolean
    LANGUAGE plpgsql
    AS $$
  BEGIN
    IF EXISTS (SELECT FROM cs_configuration_v1 c WHERE c.state = 'pending') THEN
        DELETE FROM cs_configuration_v1 WHERE state = 'pending';
      RETURN true;
    ELSE
      RAISE EXCEPTION 'No pending configuration exists to discard';
    END IF;
  END;
$$;


ALTER FUNCTION public.cs_discard_v1() OWNER TO postgres;

--
-- Name: cs_encrypt_v1(boolean); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_encrypt_v1(force boolean DEFAULT false) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
	BEGIN

    IF EXISTS (SELECT FROM cs_configuration_v1 c WHERE c.state = 'encrypting') THEN
      RAISE EXCEPTION 'An encryption is already in progress';
    END IF;

		IF NOT EXISTS (SELECT FROM cs_configuration_v1 c WHERE c.state = 'pending') THEN
			RAISE EXCEPTION 'No pending configuration exists to encrypt';
		END IF;

    IF NOT force THEN
      IF NOT cs_ready_for_encryption_v1() THEN
        RAISE EXCEPTION 'Some pending columns do not have an encrypted target';
      END IF;
    END IF;

    UPDATE cs_configuration_v1 SET state = 'encrypting' WHERE state = 'pending';
		RETURN true;
  END;
$$;


ALTER FUNCTION public.cs_encrypt_v1(force boolean) OWNER TO postgres;

--
-- Name: cs_match_v1_v0_0(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_match_v1_v0_0(val jsonb) RETURNS public.cs_match_index_v1
    LANGUAGE plpgsql IMMUTABLE STRICT PARALLEL SAFE
    AS $$
	BEGIN
    IF val ? 'm' THEN
      RETURN ARRAY(SELECT jsonb_array_elements(val->'m'))::cs_match_index_v1;
    END IF;
    RAISE 'Expected a match index (m) value in json: %', val;
  END;
$$;


ALTER FUNCTION public.cs_match_v1_v0_0(val jsonb) OWNER TO postgres;

--
-- Name: cs_match_v1(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_match_v1(val jsonb) RETURNS public.cs_match_index_v1
    LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE
    BEGIN ATOMIC
 RETURN public.cs_match_v1_v0_0(val);
END;


ALTER FUNCTION public.cs_match_v1(val jsonb) OWNER TO postgres;

--
-- Name: cs_match_v1_v0(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_match_v1_v0(val jsonb) RETURNS public.cs_match_index_v1
    LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE
    BEGIN ATOMIC
 RETURN public.cs_match_v1_v0_0(val);
END;


ALTER FUNCTION public.cs_match_v1_v0(val jsonb) OWNER TO postgres;

--
-- Name: cs_modify_index_v1(text, text, text, text, jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_modify_index_v1(table_name text, column_name text, index_name text, cast_as text DEFAULT 'text'::text, opts jsonb DEFAULT '{}'::jsonb) RETURNS jsonb
    LANGUAGE plpgsql
    AS $$
  BEGIN
    PERFORM cs_remove_index_v1(table_name, column_name, index_name);
    RETURN cs_add_index_v1(table_name, column_name, index_name, cast_as, opts);
  END;
$$;


ALTER FUNCTION public.cs_modify_index_v1(table_name text, column_name text, index_name text, cast_as text, opts jsonb) OWNER TO postgres;

--
-- Name: cs_ore_64_8_v1_v0_0(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ore_64_8_v1_v0_0(val jsonb) RETURNS public.ore_64_8_v1
    LANGUAGE plpgsql IMMUTABLE STRICT PARALLEL SAFE
    AS $$
	BEGIN
    IF val ? 'o' THEN
      RETURN (val->>'o')::ore_64_8_v1;
    END IF;
    RAISE 'Expected an ore index (o) value in json: %', val;
  END;
$$;


ALTER FUNCTION public.cs_ore_64_8_v1_v0_0(val jsonb) OWNER TO postgres;

--
-- Name: cs_ore_64_8_v1(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ore_64_8_v1(val jsonb) RETURNS public.ore_64_8_v1
    LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE
    BEGIN ATOMIC
 RETURN public.cs_ore_64_8_v1_v0_0(val);
END;


ALTER FUNCTION public.cs_ore_64_8_v1(val jsonb) OWNER TO postgres;

--
-- Name: cs_ore_64_8_v1_v0(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ore_64_8_v1_v0(val jsonb) RETURNS public.ore_64_8_v1
    LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE
    BEGIN ATOMIC
 RETURN public.cs_ore_64_8_v1_v0_0(val);
END;


ALTER FUNCTION public.cs_ore_64_8_v1_v0(val jsonb) OWNER TO postgres;

--
-- Name: cs_ready_for_encryption_v1(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ready_for_encryption_v1() RETURNS boolean
    LANGUAGE sql STABLE STRICT PARALLEL SAFE
    AS $$
	SELECT EXISTS (
	  SELECT *
	  FROM cs_select_target_columns_v1() AS c
	  WHERE c.target_column IS NOT NULL);
$$;


ALTER FUNCTION public.cs_ready_for_encryption_v1() OWNER TO postgres;

--
-- Name: cs_refresh_encrypt_config(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_refresh_encrypt_config() RETURNS void
    LANGUAGE sql STRICT PARALLEL SAFE
    BEGIN ATOMIC
 RETURN NULL::text;
END;


ALTER FUNCTION public.cs_refresh_encrypt_config() OWNER TO postgres;

--
-- Name: cs_remove_column_v1(text, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_remove_column_v1(table_name text, column_name text) RETURNS jsonb
    LANGUAGE plpgsql
    AS $$
  DECLARE
    key text;
    _config jsonb;
  BEGIN
     -- set the active config
    SELECT data INTO _config FROM cs_configuration_v1 WHERE state = 'active' OR state = 'pending' ORDER BY state DESC;

    -- if no config
    IF _config IS NULL THEN
      RAISE EXCEPTION 'No active or pending configuration exists';
    END IF;

    -- if the table doesn't exist
    IF NOT _config #> array['tables'] ? table_name THEN
      RAISE EXCEPTION 'No configuration exists for table: %', table_name;
    END IF;

    -- if the column does not exist
    IF NOT _config #> array['tables', table_name] ?  column_name THEN
      RAISE EXCEPTION 'No configuration exists for column: % %', table_name, column_name;
    END IF;

    --  create a new pending record if we don't have one
    INSERT INTO cs_configuration_v1 (state, data) VALUES ('pending', _config)
    ON CONFLICT (state)
      WHERE state = 'pending'
    DO NOTHING;

    -- remove the column
    SELECT _config #- array['tables', table_name, column_name] INTO _config;

    -- if table  is now empty, remove the table
    IF _config #> array['tables', table_name] = '{}' THEN
      SELECT _config #- array['tables', table_name] INTO _config;
    END IF;

    -- if config empty delete
    -- or update the config
    IF _config #> array['tables'] = '{}' THEN
      DELETE FROM cs_configuration_v1 WHERE state = 'pending';
    ELSE
      UPDATE cs_configuration_v1 SET data = _config WHERE state = 'pending';
    END IF;

    -- exeunt
    RETURN _config;

  END;
$$;


ALTER FUNCTION public.cs_remove_column_v1(table_name text, column_name text) OWNER TO postgres;

--
-- Name: cs_remove_index_v1(text, text, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_remove_index_v1(table_name text, column_name text, index_name text) RETURNS jsonb
    LANGUAGE plpgsql
    AS $$
  DECLARE
    _config jsonb;
  BEGIN

    -- set the active config
    SELECT data INTO _config FROM cs_configuration_v1 WHERE state = 'active' OR state = 'pending' ORDER BY state DESC;

    -- if no config
    IF _config IS NULL THEN
      RAISE EXCEPTION 'No active or pending configuration exists';
    END IF;

    -- if the table doesn't exist
    IF NOT _config #> array['tables'] ? table_name THEN
      RAISE EXCEPTION 'No configuration exists for table: %', table_name;
    END IF;

    -- if the index does not exist
    -- IF NOT _config->key ? index_name THEN
    IF NOT _config #> array['tables', table_name] ?  column_name THEN
      RAISE EXCEPTION 'No % index exists for column: % %', index_name, table_name, column_name;
    END IF;

    --  create a new pending record if we don't have one
    INSERT INTO cs_configuration_v1 (state, data) VALUES ('pending', _config)
    ON CONFLICT (state)
      WHERE state = 'pending'
    DO NOTHING;

    -- remove the index
    SELECT _config #- array['tables', table_name, column_name, 'indexes', index_name] INTO _config;

    -- if column is now empty, remove the column
    IF _config #> array['tables', table_name, column_name, 'indexes'] = '{}' THEN
      SELECT _config #- array['tables', table_name, column_name] INTO _config;
    END IF;

    -- if table  is now empty, remove the table
    IF _config #> array['tables', table_name] = '{}' THEN
      SELECT _config #- array['tables', table_name] INTO _config;
    END IF;

    -- if config empty delete
    -- or update the config
    IF _config #> array['tables'] = '{}' THEN
      DELETE FROM cs_configuration_v1 WHERE state = 'pending';
    ELSE
      UPDATE cs_configuration_v1 SET data = _config WHERE state = 'pending';
    END IF;

    -- exeunt
    RETURN _config;
  END;
$$;


ALTER FUNCTION public.cs_remove_index_v1(table_name text, column_name text, index_name text) OWNER TO postgres;

--
-- Name: cs_rename_encrypted_columns_v1(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_rename_encrypted_columns_v1() RETURNS TABLE(table_name text, column_name text, target_column text)
    LANGUAGE plpgsql
    AS $$
	BEGIN
    FOR table_name, column_name, target_column IN
      SELECT * FROM cs_select_target_columns_v1() as c WHERE c.target_column = c.column_name || '_encrypted'
    LOOP
		  EXECUTE format('ALTER TABLE %I RENAME %I TO %I;', table_name, column_name, column_name || '_plaintext');
		  EXECUTE format('ALTER TABLE %I RENAME %I TO %I;', table_name, target_column, column_name);
      RETURN NEXT;
    END LOOP;
	END;
$$;


ALTER FUNCTION public.cs_rename_encrypted_columns_v1() OWNER TO postgres;

--
-- Name: cs_select_pending_columns_v1(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_select_pending_columns_v1() RETURNS TABLE(table_name text, column_name text)
    LANGUAGE plpgsql
    AS $$
	DECLARE
		active JSONB;
		pending JSONB;
		config_id BIGINT;
	BEGIN
		SELECT data INTO active FROM cs_configuration_v1 WHERE state = 'active';

		-- set default config
    IF active IS NULL THEN
      active := '{}';
    END IF;

		SELECT id, data INTO config_id, pending FROM cs_configuration_v1 WHERE state = 'pending';

		-- set default config
		IF config_id IS NULL THEN
			RAISE EXCEPTION 'No pending configuration exists to encrypt';
		END IF;

		RETURN QUERY
		SELECT d.table_name, d.column_name FROM _cs_diff_config_v1(active, pending) as d;
	END;
$$;


ALTER FUNCTION public.cs_select_pending_columns_v1() OWNER TO postgres;

--
-- Name: cs_select_target_columns_v1(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_select_target_columns_v1() RETURNS TABLE(table_name text, column_name text, target_column text)
    LANGUAGE sql STABLE STRICT PARALLEL SAFE
    AS $$
  SELECT
    c.table_name,
    c.column_name,
    s.column_name as target_column
  FROM
    cs_select_pending_columns_v1() c
  LEFT JOIN information_schema.columns s ON
    s.table_name = c.table_name AND
    (s.column_name = c.column_name OR s.column_name = c.column_name || '_encrypted') AND
    (s.domain_name = 'cs_encrypted_v1' OR s.data_type = 'jsonb');
$$;


ALTER FUNCTION public.cs_select_target_columns_v1() OWNER TO postgres;

--
-- Name: cs_ste_vec_encrypted_term_eq(public.cs_ste_vec_encrypted_term_v1, public.cs_ste_vec_encrypted_term_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ste_vec_encrypted_term_eq(a public.cs_ste_vec_encrypted_term_v1, b public.cs_ste_vec_encrypted_term_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT __bytea_ct_eq(a.bytes, b.bytes)
$$;


ALTER FUNCTION public.cs_ste_vec_encrypted_term_eq(a public.cs_ste_vec_encrypted_term_v1, b public.cs_ste_vec_encrypted_term_v1) OWNER TO postgres;

--
-- Name: cs_ste_vec_encrypted_term_gt(public.cs_ste_vec_encrypted_term_v1, public.cs_ste_vec_encrypted_term_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ste_vec_encrypted_term_gt(a public.cs_ste_vec_encrypted_term_v1, b public.cs_ste_vec_encrypted_term_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ste_vec_encrypted_term_v1(a, b) = 1
$$;


ALTER FUNCTION public.cs_ste_vec_encrypted_term_gt(a public.cs_ste_vec_encrypted_term_v1, b public.cs_ste_vec_encrypted_term_v1) OWNER TO postgres;

--
-- Name: cs_ste_vec_encrypted_term_gte(public.cs_ste_vec_encrypted_term_v1, public.cs_ste_vec_encrypted_term_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ste_vec_encrypted_term_gte(a public.cs_ste_vec_encrypted_term_v1, b public.cs_ste_vec_encrypted_term_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ste_vec_encrypted_term_v1(a, b) != -1
$$;


ALTER FUNCTION public.cs_ste_vec_encrypted_term_gte(a public.cs_ste_vec_encrypted_term_v1, b public.cs_ste_vec_encrypted_term_v1) OWNER TO postgres;

--
-- Name: cs_ste_vec_encrypted_term_lt(public.cs_ste_vec_encrypted_term_v1, public.cs_ste_vec_encrypted_term_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ste_vec_encrypted_term_lt(a public.cs_ste_vec_encrypted_term_v1, b public.cs_ste_vec_encrypted_term_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ste_vec_encrypted_term_v1(a, b) = -1
$$;


ALTER FUNCTION public.cs_ste_vec_encrypted_term_lt(a public.cs_ste_vec_encrypted_term_v1, b public.cs_ste_vec_encrypted_term_v1) OWNER TO postgres;

--
-- Name: cs_ste_vec_encrypted_term_lte(public.cs_ste_vec_encrypted_term_v1, public.cs_ste_vec_encrypted_term_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ste_vec_encrypted_term_lte(a public.cs_ste_vec_encrypted_term_v1, b public.cs_ste_vec_encrypted_term_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ste_vec_encrypted_term_v1(a, b) != 1
$$;


ALTER FUNCTION public.cs_ste_vec_encrypted_term_lte(a public.cs_ste_vec_encrypted_term_v1, b public.cs_ste_vec_encrypted_term_v1) OWNER TO postgres;

--
-- Name: cs_ste_vec_encrypted_term_neq(public.cs_ste_vec_encrypted_term_v1, public.cs_ste_vec_encrypted_term_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ste_vec_encrypted_term_neq(a public.cs_ste_vec_encrypted_term_v1, b public.cs_ste_vec_encrypted_term_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT not __bytea_ct_eq(a.bytes, b.bytes)
$$;


ALTER FUNCTION public.cs_ste_vec_encrypted_term_neq(a public.cs_ste_vec_encrypted_term_v1, b public.cs_ste_vec_encrypted_term_v1) OWNER TO postgres;

--
-- Name: cs_ste_vec_term_v1(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ste_vec_term_v1(col jsonb) RETURNS public.cs_ste_vec_encrypted_term_v1
    LANGUAGE plpgsql
    AS $$
DECLARE
  ste_vec_index cs_ste_vec_index_v1;
BEGIN
  ste_vec_index := cs_ste_vec_v1(col);

  IF ste_vec_index IS NULL THEN
    RETURN NULL;
  END IF;

  RETURN ste_vec_index.entries[1].term;
END;
$$;


ALTER FUNCTION public.cs_ste_vec_term_v1(col jsonb) OWNER TO postgres;

--
-- Name: cs_ste_vec_term_v1(jsonb, jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ste_vec_term_v1(col jsonb, selector jsonb) RETURNS public.cs_ste_vec_encrypted_term_v1
    LANGUAGE plpgsql
    AS $$
DECLARE
  ste_vec_index cs_ste_vec_index_v1;
  target_selector text;
  found cs_ste_vec_encrypted_term_v1;
  ignored cs_ste_vec_encrypted_term_v1;
  i integer;
BEGIN
  ste_vec_index := cs_ste_vec_v1(col);

  IF ste_vec_index IS NULL THEN
    RETURN NULL;
  END IF;

  target_selector := selector->>'svs';

  FOR i IN 1..array_length(ste_vec_index.entries, 1) LOOP
      -- The ELSE part is to help ensure constant time operation.
      -- The result is thrown away.
      IF ste_vec_index.entries[i].tokenized_selector = target_selector THEN
        found := ste_vec_index.entries[i].term;
      ELSE
        ignored := ste_vec_index.entries[i].term;
      END IF;
  END LOOP;

  RETURN found;
END;
$$;


ALTER FUNCTION public.cs_ste_vec_term_v1(col jsonb, selector jsonb) OWNER TO postgres;

--
-- Name: cs_ste_vec_terms_v1(jsonb, jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ste_vec_terms_v1(col jsonb, selector jsonb) RETURNS public.cs_ste_vec_encrypted_term_v1[]
    LANGUAGE plpgsql
    AS $$
DECLARE
  ste_vec_index cs_ste_vec_index_v1;
  target_selector text;
  found cs_ste_vec_encrypted_term_v1;
  ignored cs_ste_vec_encrypted_term_v1;
  i integer;
  term_array cs_ste_vec_encrypted_term_v1[];
BEGIN
  ste_vec_index := cs_ste_vec_v1(col);

  IF ste_vec_index IS NULL THEN
    RETURN NULL;
  END IF;

  target_selector := selector->>'svs';

  FOR i IN 1..array_length(ste_vec_index.entries, 1) LOOP
      -- The ELSE part is to help ensure constant time operation.
      -- The result is thrown away.
      IF ste_vec_index.entries[i].tokenized_selector = target_selector THEN
        found := ste_vec_index.entries[i].term;
        term_array := array_append(term_array, found);
      ELSE
        ignored := ste_vec_index.entries[i].term;
      END IF;
  END LOOP;

  RETURN term_array;
END;
$$;


ALTER FUNCTION public.cs_ste_vec_terms_v1(col jsonb, selector jsonb) OWNER TO postgres;

--
-- Name: cs_ste_vec_v1_v0_0(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ste_vec_v1_v0_0(val jsonb) RETURNS public.cs_ste_vec_index_v1
    LANGUAGE plpgsql IMMUTABLE STRICT PARALLEL SAFE
    AS $$
	BEGIN
    IF val ? 'sv' THEN
      RETURN (val->'sv')::cs_ste_vec_index_v1;
    END IF;
    RAISE 'Expected a structured vector index (sv) value in json: %', val;
  END;
$$;


ALTER FUNCTION public.cs_ste_vec_v1_v0_0(val jsonb) OWNER TO postgres;

--
-- Name: cs_ste_vec_v1(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ste_vec_v1(val jsonb) RETURNS public.cs_ste_vec_index_v1
    LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE
    BEGIN ATOMIC
 RETURN public.cs_ste_vec_v1_v0_0(val);
END;


ALTER FUNCTION public.cs_ste_vec_v1(val jsonb) OWNER TO postgres;

--
-- Name: cs_ste_vec_v1_entry_array_contains_entry(public.cs_ste_vec_v1_entry[], public.cs_ste_vec_v1_entry); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ste_vec_v1_entry_array_contains_entry(a public.cs_ste_vec_v1_entry[], b public.cs_ste_vec_v1_entry) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
    result boolean;
    intermediate_result boolean;
BEGIN
    IF array_length(a, 1) IS NULL THEN
        RETURN false;
    END IF;

    result := false;
    FOR i IN 1..array_length(a, 1) LOOP
        intermediate_result := a[i].tokenized_selector = b.tokenized_selector AND a[i].term = b.term;
        result := result OR intermediate_result;
    END LOOP;
    RETURN result;
END;
$$;


ALTER FUNCTION public.cs_ste_vec_v1_entry_array_contains_entry(a public.cs_ste_vec_v1_entry[], b public.cs_ste_vec_v1_entry) OWNER TO postgres;

--
-- Name: cs_ste_vec_v1_entry_eq(public.cs_ste_vec_v1_entry, public.cs_ste_vec_v1_entry); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ste_vec_v1_entry_eq(a public.cs_ste_vec_v1_entry, b public.cs_ste_vec_v1_entry) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
    sel_cmp int;
    term_cmp int;
BEGIN
    -- Constant time comparison
    IF a.tokenized_selector = b.tokenized_selector THEN
        sel_cmp := 1;
    ELSE
        sel_cmp := 0;
    END IF;
    IF a.term = b.term THEN
        term_cmp := 1;
    ELSE
        term_cmp := 0;
    END IF;
    RETURN (sel_cmp # term_cmp) = 0;
END;
$$;


ALTER FUNCTION public.cs_ste_vec_v1_entry_eq(a public.cs_ste_vec_v1_entry, b public.cs_ste_vec_v1_entry) OWNER TO postgres;

--
-- Name: cs_ste_vec_v1_v0(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ste_vec_v1_v0(val jsonb) RETURNS public.cs_ste_vec_index_v1
    LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE
    BEGIN ATOMIC
 RETURN public.cs_ste_vec_v1_v0_0(val);
END;


ALTER FUNCTION public.cs_ste_vec_v1_v0(val jsonb) OWNER TO postgres;

--
-- Name: cs_ste_vec_value_v1(jsonb, jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_ste_vec_value_v1(col jsonb, selector jsonb) RETURNS jsonb
    LANGUAGE plpgsql
    AS $$
DECLARE
  ste_vec_index cs_ste_vec_index_v1;
  target_selector text;
  found text;
  ignored text;
  i integer;
BEGIN
  ste_vec_index := cs_ste_vec_v1(col);

  IF ste_vec_index IS NULL THEN
    RETURN NULL;
  END IF;

  target_selector := selector->>'svs';

  FOR i IN 1..array_length(ste_vec_index.entries, 1) LOOP
      -- The ELSE part is to help ensure constant time operation.
      -- The result is thrown away.
      IF ste_vec_index.entries[i].tokenized_selector = target_selector THEN
        found := ste_vec_index.entries[i].ciphertext;
      ELSE
        ignored := ste_vec_index.entries[i].ciphertext;
      END IF;
  END LOOP;

  IF found IS NOT NULL THEN
    RETURN jsonb_build_object(
      'k', 'ct',
      'c', found,
      'o', NULL,
      'm', NULL,
      'u', NULL,
      'i', col->'i',
      'v', 1
    );
  ELSE
    RETURN NULL;
  END IF;
END;
$$;


ALTER FUNCTION public.cs_ste_vec_value_v1(col jsonb, selector jsonb) OWNER TO postgres;

--
-- Name: cs_unique_v1_v0_0(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_unique_v1_v0_0(val jsonb) RETURNS public.cs_unique_index_v1
    LANGUAGE plpgsql IMMUTABLE STRICT PARALLEL SAFE
    AS $$
	BEGIN
    IF val ? 'u' THEN
      RETURN val->>'u';
    END IF;
    RAISE 'Expected a unique index (u) value in json: %', val;
  END;
$$;


ALTER FUNCTION public.cs_unique_v1_v0_0(val jsonb) OWNER TO postgres;

--
-- Name: cs_unique_v1(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_unique_v1(val jsonb) RETURNS public.cs_unique_index_v1
    LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE
    BEGIN ATOMIC
 RETURN public.cs_unique_v1_v0_0(val);
END;


ALTER FUNCTION public.cs_unique_v1(val jsonb) OWNER TO postgres;

--
-- Name: cs_unique_v1_v0(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.cs_unique_v1_v0(val jsonb) RETURNS public.cs_unique_index_v1
    LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE
    BEGIN ATOMIC
 RETURN public.cs_unique_v1_v0_0(val);
END;


ALTER FUNCTION public.cs_unique_v1_v0(val jsonb) OWNER TO postgres;

--
-- Name: ore_64_8_v1_eq(public.ore_64_8_v1, public.ore_64_8_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_64_8_v1_eq(a public.ore_64_8_v1, b public.ore_64_8_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ore_64_8_v1(a, b) = 0
$$;


ALTER FUNCTION public.ore_64_8_v1_eq(a public.ore_64_8_v1, b public.ore_64_8_v1) OWNER TO postgres;

--
-- Name: ore_64_8_v1_gt(public.ore_64_8_v1, public.ore_64_8_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_64_8_v1_gt(a public.ore_64_8_v1, b public.ore_64_8_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ore_64_8_v1(a, b) = 1
$$;


ALTER FUNCTION public.ore_64_8_v1_gt(a public.ore_64_8_v1, b public.ore_64_8_v1) OWNER TO postgres;

--
-- Name: ore_64_8_v1_gte(public.ore_64_8_v1, public.ore_64_8_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_64_8_v1_gte(a public.ore_64_8_v1, b public.ore_64_8_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ore_64_8_v1(a, b) != -1
$$;


ALTER FUNCTION public.ore_64_8_v1_gte(a public.ore_64_8_v1, b public.ore_64_8_v1) OWNER TO postgres;

--
-- Name: ore_64_8_v1_lt(public.ore_64_8_v1, public.ore_64_8_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_64_8_v1_lt(a public.ore_64_8_v1, b public.ore_64_8_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ore_64_8_v1(a, b) = -1
$$;


ALTER FUNCTION public.ore_64_8_v1_lt(a public.ore_64_8_v1, b public.ore_64_8_v1) OWNER TO postgres;

--
-- Name: ore_64_8_v1_lte(public.ore_64_8_v1, public.ore_64_8_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_64_8_v1_lte(a public.ore_64_8_v1, b public.ore_64_8_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ore_64_8_v1(a, b) != 1
$$;


ALTER FUNCTION public.ore_64_8_v1_lte(a public.ore_64_8_v1, b public.ore_64_8_v1) OWNER TO postgres;

--
-- Name: ore_64_8_v1_neq(public.ore_64_8_v1, public.ore_64_8_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_64_8_v1_neq(a public.ore_64_8_v1, b public.ore_64_8_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ore_64_8_v1(a, b) <> 0
$$;


ALTER FUNCTION public.ore_64_8_v1_neq(a public.ore_64_8_v1, b public.ore_64_8_v1) OWNER TO postgres;

--
-- Name: ore_64_8_v1_term_eq(public.ore_64_8_v1_term, public.ore_64_8_v1_term); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_64_8_v1_term_eq(a public.ore_64_8_v1_term, b public.ore_64_8_v1_term) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ore_64_8_v1_term(a, b) = 0
$$;


ALTER FUNCTION public.ore_64_8_v1_term_eq(a public.ore_64_8_v1_term, b public.ore_64_8_v1_term) OWNER TO postgres;

--
-- Name: ore_64_8_v1_term_gt(public.ore_64_8_v1_term, public.ore_64_8_v1_term); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_64_8_v1_term_gt(a public.ore_64_8_v1_term, b public.ore_64_8_v1_term) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ore_64_8_v1_term(a, b) = 1
$$;


ALTER FUNCTION public.ore_64_8_v1_term_gt(a public.ore_64_8_v1_term, b public.ore_64_8_v1_term) OWNER TO postgres;

--
-- Name: ore_64_8_v1_term_gte(public.ore_64_8_v1_term, public.ore_64_8_v1_term); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_64_8_v1_term_gte(a public.ore_64_8_v1_term, b public.ore_64_8_v1_term) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ore_64_8_v1_term(a, b) != -1
$$;


ALTER FUNCTION public.ore_64_8_v1_term_gte(a public.ore_64_8_v1_term, b public.ore_64_8_v1_term) OWNER TO postgres;

--
-- Name: ore_64_8_v1_term_lt(public.ore_64_8_v1_term, public.ore_64_8_v1_term); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_64_8_v1_term_lt(a public.ore_64_8_v1_term, b public.ore_64_8_v1_term) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ore_64_8_v1_term(a, b) = -1
$$;


ALTER FUNCTION public.ore_64_8_v1_term_lt(a public.ore_64_8_v1_term, b public.ore_64_8_v1_term) OWNER TO postgres;

--
-- Name: ore_64_8_v1_term_lte(public.ore_64_8_v1_term, public.ore_64_8_v1_term); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_64_8_v1_term_lte(a public.ore_64_8_v1_term, b public.ore_64_8_v1_term) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ore_64_8_v1_term(a, b) != 1
$$;


ALTER FUNCTION public.ore_64_8_v1_term_lte(a public.ore_64_8_v1_term, b public.ore_64_8_v1_term) OWNER TO postgres;

--
-- Name: ore_64_8_v1_term_neq(public.ore_64_8_v1_term, public.ore_64_8_v1_term); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_64_8_v1_term_neq(a public.ore_64_8_v1_term, b public.ore_64_8_v1_term) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ore_64_8_v1_term(a, b) <> 0
$$;


ALTER FUNCTION public.ore_64_8_v1_term_neq(a public.ore_64_8_v1_term, b public.ore_64_8_v1_term) OWNER TO postgres;

--
-- Name: ore_cllw_8_v1_eq(public.ore_cllw_8_v1, public.ore_cllw_8_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_cllw_8_v1_eq(a public.ore_cllw_8_v1, b public.ore_cllw_8_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT __bytea_ct_eq(a.bytes, b.bytes)
$$;


ALTER FUNCTION public.ore_cllw_8_v1_eq(a public.ore_cllw_8_v1, b public.ore_cllw_8_v1) OWNER TO postgres;

--
-- Name: ore_cllw_8_v1_gt(public.ore_cllw_8_v1, public.ore_cllw_8_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_cllw_8_v1_gt(a public.ore_cllw_8_v1, b public.ore_cllw_8_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ore_cllw_8_v1(a, b) = 1
$$;


ALTER FUNCTION public.ore_cllw_8_v1_gt(a public.ore_cllw_8_v1, b public.ore_cllw_8_v1) OWNER TO postgres;

--
-- Name: ore_cllw_8_v1_gt_lex(public.ore_cllw_8_variable_v1, public.ore_cllw_8_variable_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_cllw_8_v1_gt_lex(a public.ore_cllw_8_variable_v1, b public.ore_cllw_8_variable_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_lex_ore_cllw_8_v1(a, b) = 1
$$;


ALTER FUNCTION public.ore_cllw_8_v1_gt_lex(a public.ore_cllw_8_variable_v1, b public.ore_cllw_8_variable_v1) OWNER TO postgres;

--
-- Name: ore_cllw_8_v1_gte(public.ore_cllw_8_v1, public.ore_cllw_8_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_cllw_8_v1_gte(a public.ore_cllw_8_v1, b public.ore_cllw_8_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ore_cllw_8_v1(a, b) != -1
$$;


ALTER FUNCTION public.ore_cllw_8_v1_gte(a public.ore_cllw_8_v1, b public.ore_cllw_8_v1) OWNER TO postgres;

--
-- Name: ore_cllw_8_v1_gte_lex(public.ore_cllw_8_variable_v1, public.ore_cllw_8_variable_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_cllw_8_v1_gte_lex(a public.ore_cllw_8_variable_v1, b public.ore_cllw_8_variable_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_lex_ore_cllw_8_v1(a, b) != -1
$$;


ALTER FUNCTION public.ore_cllw_8_v1_gte_lex(a public.ore_cllw_8_variable_v1, b public.ore_cllw_8_variable_v1) OWNER TO postgres;

--
-- Name: ore_cllw_8_v1_lt(public.ore_cllw_8_v1, public.ore_cllw_8_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_cllw_8_v1_lt(a public.ore_cllw_8_v1, b public.ore_cllw_8_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ore_cllw_8_v1(a, b) = -1
$$;


ALTER FUNCTION public.ore_cllw_8_v1_lt(a public.ore_cllw_8_v1, b public.ore_cllw_8_v1) OWNER TO postgres;

--
-- Name: ore_cllw_8_v1_lt_lex(public.ore_cllw_8_variable_v1, public.ore_cllw_8_variable_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_cllw_8_v1_lt_lex(a public.ore_cllw_8_variable_v1, b public.ore_cllw_8_variable_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_lex_ore_cllw_8_v1(a, b) = -1
$$;


ALTER FUNCTION public.ore_cllw_8_v1_lt_lex(a public.ore_cllw_8_variable_v1, b public.ore_cllw_8_variable_v1) OWNER TO postgres;

--
-- Name: ore_cllw_8_v1_lte(public.ore_cllw_8_v1, public.ore_cllw_8_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_cllw_8_v1_lte(a public.ore_cllw_8_v1, b public.ore_cllw_8_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_ore_cllw_8_v1(a, b) != 1
$$;


ALTER FUNCTION public.ore_cllw_8_v1_lte(a public.ore_cllw_8_v1, b public.ore_cllw_8_v1) OWNER TO postgres;

--
-- Name: ore_cllw_8_v1_lte_lex(public.ore_cllw_8_variable_v1, public.ore_cllw_8_variable_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_cllw_8_v1_lte_lex(a public.ore_cllw_8_variable_v1, b public.ore_cllw_8_variable_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT compare_lex_ore_cllw_8_v1(a, b) != 1
$$;


ALTER FUNCTION public.ore_cllw_8_v1_lte_lex(a public.ore_cllw_8_variable_v1, b public.ore_cllw_8_variable_v1) OWNER TO postgres;

--
-- Name: ore_cllw_8_v1_neq(public.ore_cllw_8_v1, public.ore_cllw_8_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_cllw_8_v1_neq(a public.ore_cllw_8_v1, b public.ore_cllw_8_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT not __bytea_ct_eq(a.bytes, b.bytes)
$$;


ALTER FUNCTION public.ore_cllw_8_v1_neq(a public.ore_cllw_8_v1, b public.ore_cllw_8_v1) OWNER TO postgres;

--
-- Name: ore_cllw_8_variable_v1_eq(public.ore_cllw_8_variable_v1, public.ore_cllw_8_variable_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_cllw_8_variable_v1_eq(a public.ore_cllw_8_variable_v1, b public.ore_cllw_8_variable_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT __bytea_ct_eq(a.bytes, b.bytes)
$$;


ALTER FUNCTION public.ore_cllw_8_variable_v1_eq(a public.ore_cllw_8_variable_v1, b public.ore_cllw_8_variable_v1) OWNER TO postgres;

--
-- Name: ore_cllw_8_variable_v1_neq(public.ore_cllw_8_variable_v1, public.ore_cllw_8_variable_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ore_cllw_8_variable_v1_neq(a public.ore_cllw_8_variable_v1, b public.ore_cllw_8_variable_v1) RETURNS boolean
    LANGUAGE sql
    AS $$
  SELECT not __bytea_ct_eq(a.bytes, b.bytes)
$$;


ALTER FUNCTION public.ore_cllw_8_variable_v1_neq(a public.ore_cllw_8_variable_v1, b public.ore_cllw_8_variable_v1) OWNER TO postgres;

--
-- Name: ste_vec_v1_logical_contains(public.cs_ste_vec_index_v1, public.cs_ste_vec_index_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ste_vec_v1_logical_contains(a public.cs_ste_vec_index_v1, b public.cs_ste_vec_index_v1) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
    result boolean;
    intermediate_result boolean;
BEGIN
    result := true;
    IF array_length(b.entries, 1) IS NULL THEN
        RETURN result;
    END IF;
    FOR i IN 1..array_length(b.entries, 1) LOOP
        intermediate_result := cs_ste_vec_v1_entry_array_contains_entry(a.entries, b.entries[i]);
        result := result AND intermediate_result;
    END LOOP;
    RETURN result;
END;
$$;


ALTER FUNCTION public.ste_vec_v1_logical_contains(a public.cs_ste_vec_index_v1, b public.cs_ste_vec_index_v1) OWNER TO postgres;

--
-- Name: ste_vec_v1_logical_is_contained(public.cs_ste_vec_index_v1, public.cs_ste_vec_index_v1); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ste_vec_v1_logical_is_contained(a public.cs_ste_vec_index_v1, b public.cs_ste_vec_index_v1) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN ste_vec_v1_logical_contains(b, a);
END;
$$;


ALTER FUNCTION public.ste_vec_v1_logical_is_contained(a public.cs_ste_vec_index_v1, b public.cs_ste_vec_index_v1) OWNER TO postgres;

--
-- Name: cs_grouped_value_v1(jsonb); Type: AGGREGATE; Schema: public; Owner: postgres
--

CREATE AGGREGATE public.cs_grouped_value_v1(jsonb) (
    SFUNC = public._cs_first_grouped_value,
    STYPE = jsonb
);


ALTER AGGREGATE public.cs_grouped_value_v1(jsonb) OWNER TO postgres;

--
-- Name: <; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.< (
    FUNCTION = public.ore_64_8_v1_term_lt,
    LEFTARG = public.ore_64_8_v1_term,
    RIGHTARG = public.ore_64_8_v1_term,
    COMMUTATOR = OPERATOR(public.>),
    NEGATOR = OPERATOR(public.>=),
    RESTRICT = scalarltsel,
    JOIN = scalarltjoinsel
);


ALTER OPERATOR public.< (public.ore_64_8_v1_term, public.ore_64_8_v1_term) OWNER TO postgres;

--
-- Name: <; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.< (
    FUNCTION = public.ore_64_8_v1_lt,
    LEFTARG = public.ore_64_8_v1,
    RIGHTARG = public.ore_64_8_v1,
    COMMUTATOR = OPERATOR(public.>),
    NEGATOR = OPERATOR(public.>=),
    RESTRICT = scalarltsel,
    JOIN = scalarltjoinsel
);


ALTER OPERATOR public.< (public.ore_64_8_v1, public.ore_64_8_v1) OWNER TO postgres;

--
-- Name: <; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.< (
    FUNCTION = public.ore_cllw_8_v1_lt,
    LEFTARG = public.ore_cllw_8_v1,
    RIGHTARG = public.ore_cllw_8_v1,
    NEGATOR = OPERATOR(public.>=),
    MERGES,
    HASHES,
    RESTRICT = scalargtsel,
    JOIN = scalargtjoinsel
);


ALTER OPERATOR public.< (public.ore_cllw_8_v1, public.ore_cllw_8_v1) OWNER TO postgres;

--
-- Name: <; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.< (
    FUNCTION = public.ore_cllw_8_v1_lt_lex,
    LEFTARG = public.ore_cllw_8_variable_v1,
    RIGHTARG = public.ore_cllw_8_variable_v1,
    NEGATOR = OPERATOR(public.>=),
    MERGES,
    HASHES,
    RESTRICT = scalargtsel,
    JOIN = scalargtjoinsel
);


ALTER OPERATOR public.< (public.ore_cllw_8_variable_v1, public.ore_cllw_8_variable_v1) OWNER TO postgres;

--
-- Name: <; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.< (
    FUNCTION = public.cs_ste_vec_encrypted_term_lt,
    LEFTARG = public.cs_ste_vec_encrypted_term_v1,
    RIGHTARG = public.cs_ste_vec_encrypted_term_v1,
    NEGATOR = OPERATOR(public.>=),
    MERGES,
    HASHES,
    RESTRICT = scalargtsel,
    JOIN = scalargtjoinsel
);


ALTER OPERATOR public.< (public.cs_ste_vec_encrypted_term_v1, public.cs_ste_vec_encrypted_term_v1) OWNER TO postgres;

--
-- Name: <=; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.<= (
    FUNCTION = public.ore_64_8_v1_term_lte,
    LEFTARG = public.ore_64_8_v1_term,
    RIGHTARG = public.ore_64_8_v1_term,
    COMMUTATOR = OPERATOR(public.>=),
    NEGATOR = OPERATOR(public.>),
    RESTRICT = scalarlesel,
    JOIN = scalarlejoinsel
);


ALTER OPERATOR public.<= (public.ore_64_8_v1_term, public.ore_64_8_v1_term) OWNER TO postgres;

--
-- Name: <=; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.<= (
    FUNCTION = public.ore_64_8_v1_lte,
    LEFTARG = public.ore_64_8_v1,
    RIGHTARG = public.ore_64_8_v1,
    COMMUTATOR = OPERATOR(public.>=),
    NEGATOR = OPERATOR(public.>),
    RESTRICT = scalarlesel,
    JOIN = scalarlejoinsel
);


ALTER OPERATOR public.<= (public.ore_64_8_v1, public.ore_64_8_v1) OWNER TO postgres;

--
-- Name: <=; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.<= (
    FUNCTION = public.ore_cllw_8_v1_lte,
    LEFTARG = public.ore_cllw_8_v1,
    RIGHTARG = public.ore_cllw_8_v1,
    NEGATOR = OPERATOR(public.>),
    MERGES,
    HASHES,
    RESTRICT = scalargtsel,
    JOIN = scalargtjoinsel
);


ALTER OPERATOR public.<= (public.ore_cllw_8_v1, public.ore_cllw_8_v1) OWNER TO postgres;

--
-- Name: <=; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.<= (
    FUNCTION = public.ore_cllw_8_v1_lte_lex,
    LEFTARG = public.ore_cllw_8_variable_v1,
    RIGHTARG = public.ore_cllw_8_variable_v1,
    NEGATOR = OPERATOR(public.>),
    MERGES,
    HASHES,
    RESTRICT = scalargtsel,
    JOIN = scalargtjoinsel
);


ALTER OPERATOR public.<= (public.ore_cllw_8_variable_v1, public.ore_cllw_8_variable_v1) OWNER TO postgres;

--
-- Name: <=; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.<= (
    FUNCTION = public.cs_ste_vec_encrypted_term_lte,
    LEFTARG = public.cs_ste_vec_encrypted_term_v1,
    RIGHTARG = public.cs_ste_vec_encrypted_term_v1,
    NEGATOR = OPERATOR(public.>),
    MERGES,
    HASHES,
    RESTRICT = scalargtsel,
    JOIN = scalargtjoinsel
);


ALTER OPERATOR public.<= (public.cs_ste_vec_encrypted_term_v1, public.cs_ste_vec_encrypted_term_v1) OWNER TO postgres;

--
-- Name: <>; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.<> (
    FUNCTION = public.ore_64_8_v1_term_neq,
    LEFTARG = public.ore_64_8_v1_term,
    RIGHTARG = public.ore_64_8_v1_term,
    NEGATOR = OPERATOR(public.=),
    MERGES,
    HASHES,
    RESTRICT = eqsel,
    JOIN = eqjoinsel
);


ALTER OPERATOR public.<> (public.ore_64_8_v1_term, public.ore_64_8_v1_term) OWNER TO postgres;

--
-- Name: <>; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.<> (
    FUNCTION = public.ore_64_8_v1_neq,
    LEFTARG = public.ore_64_8_v1,
    RIGHTARG = public.ore_64_8_v1,
    NEGATOR = OPERATOR(public.=),
    MERGES,
    HASHES,
    RESTRICT = eqsel,
    JOIN = eqjoinsel
);


ALTER OPERATOR public.<> (public.ore_64_8_v1, public.ore_64_8_v1) OWNER TO postgres;

--
-- Name: <>; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.<> (
    FUNCTION = public.ore_cllw_8_v1_neq,
    LEFTARG = public.ore_cllw_8_v1,
    RIGHTARG = public.ore_cllw_8_v1,
    NEGATOR = OPERATOR(public.=),
    MERGES,
    HASHES,
    RESTRICT = eqsel,
    JOIN = eqjoinsel
);


ALTER OPERATOR public.<> (public.ore_cllw_8_v1, public.ore_cllw_8_v1) OWNER TO postgres;

--
-- Name: <>; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.<> (
    FUNCTION = public.ore_cllw_8_variable_v1_neq,
    LEFTARG = public.ore_cllw_8_variable_v1,
    RIGHTARG = public.ore_cllw_8_variable_v1,
    NEGATOR = OPERATOR(public.=),
    MERGES,
    HASHES,
    RESTRICT = eqsel,
    JOIN = eqjoinsel
);


ALTER OPERATOR public.<> (public.ore_cllw_8_variable_v1, public.ore_cllw_8_variable_v1) OWNER TO postgres;

--
-- Name: <>; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.<> (
    FUNCTION = public.cs_ste_vec_encrypted_term_neq,
    LEFTARG = public.cs_ste_vec_encrypted_term_v1,
    RIGHTARG = public.cs_ste_vec_encrypted_term_v1,
    NEGATOR = OPERATOR(public.=),
    MERGES,
    HASHES,
    RESTRICT = eqsel,
    JOIN = eqjoinsel
);


ALTER OPERATOR public.<> (public.cs_ste_vec_encrypted_term_v1, public.cs_ste_vec_encrypted_term_v1) OWNER TO postgres;

--
-- Name: <@; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.<@ (
    FUNCTION = public.ste_vec_v1_logical_is_contained,
    LEFTARG = public.cs_ste_vec_index_v1,
    RIGHTARG = public.cs_ste_vec_index_v1,
    COMMUTATOR = OPERATOR(public.@>)
);


ALTER OPERATOR public.<@ (public.cs_ste_vec_index_v1, public.cs_ste_vec_index_v1) OWNER TO postgres;

--
-- Name: =; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.= (
    FUNCTION = public.ore_64_8_v1_term_eq,
    LEFTARG = public.ore_64_8_v1_term,
    RIGHTARG = public.ore_64_8_v1_term,
    NEGATOR = OPERATOR(public.<>),
    MERGES,
    HASHES,
    RESTRICT = eqsel,
    JOIN = eqjoinsel
);


ALTER OPERATOR public.= (public.ore_64_8_v1_term, public.ore_64_8_v1_term) OWNER TO postgres;

--
-- Name: =; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.= (
    FUNCTION = public.ore_64_8_v1_eq,
    LEFTARG = public.ore_64_8_v1,
    RIGHTARG = public.ore_64_8_v1,
    NEGATOR = OPERATOR(public.<>),
    MERGES,
    HASHES,
    RESTRICT = eqsel,
    JOIN = eqjoinsel
);


ALTER OPERATOR public.= (public.ore_64_8_v1, public.ore_64_8_v1) OWNER TO postgres;

--
-- Name: =; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.= (
    FUNCTION = public.ore_cllw_8_v1_eq,
    LEFTARG = public.ore_cllw_8_v1,
    RIGHTARG = public.ore_cllw_8_v1,
    NEGATOR = OPERATOR(public.<>),
    MERGES,
    HASHES,
    RESTRICT = eqsel,
    JOIN = eqjoinsel
);


ALTER OPERATOR public.= (public.ore_cllw_8_v1, public.ore_cllw_8_v1) OWNER TO postgres;

--
-- Name: =; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.= (
    FUNCTION = public.ore_cllw_8_variable_v1_eq,
    LEFTARG = public.ore_cllw_8_variable_v1,
    RIGHTARG = public.ore_cllw_8_variable_v1,
    NEGATOR = OPERATOR(public.<>),
    MERGES,
    HASHES,
    RESTRICT = eqsel,
    JOIN = eqjoinsel
);


ALTER OPERATOR public.= (public.ore_cllw_8_variable_v1, public.ore_cllw_8_variable_v1) OWNER TO postgres;

--
-- Name: =; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.= (
    FUNCTION = public.cs_ste_vec_encrypted_term_eq,
    LEFTARG = public.cs_ste_vec_encrypted_term_v1,
    RIGHTARG = public.cs_ste_vec_encrypted_term_v1,
    NEGATOR = OPERATOR(public.<>),
    MERGES,
    HASHES,
    RESTRICT = eqsel,
    JOIN = eqjoinsel
);


ALTER OPERATOR public.= (public.cs_ste_vec_encrypted_term_v1, public.cs_ste_vec_encrypted_term_v1) OWNER TO postgres;

--
-- Name: >; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.> (
    FUNCTION = public.ore_64_8_v1_term_gt,
    LEFTARG = public.ore_64_8_v1_term,
    RIGHTARG = public.ore_64_8_v1_term,
    COMMUTATOR = OPERATOR(public.<),
    NEGATOR = OPERATOR(public.<=),
    RESTRICT = scalargtsel,
    JOIN = scalargtjoinsel
);


ALTER OPERATOR public.> (public.ore_64_8_v1_term, public.ore_64_8_v1_term) OWNER TO postgres;

--
-- Name: >; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.> (
    FUNCTION = public.ore_64_8_v1_gt,
    LEFTARG = public.ore_64_8_v1,
    RIGHTARG = public.ore_64_8_v1,
    COMMUTATOR = OPERATOR(public.<),
    NEGATOR = OPERATOR(public.<=),
    RESTRICT = scalargtsel,
    JOIN = scalargtjoinsel
);


ALTER OPERATOR public.> (public.ore_64_8_v1, public.ore_64_8_v1) OWNER TO postgres;

--
-- Name: >; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.> (
    FUNCTION = public.ore_cllw_8_v1_gt,
    LEFTARG = public.ore_cllw_8_v1,
    RIGHTARG = public.ore_cllw_8_v1,
    NEGATOR = OPERATOR(public.<=),
    MERGES,
    HASHES,
    RESTRICT = scalarltsel,
    JOIN = scalarltjoinsel
);


ALTER OPERATOR public.> (public.ore_cllw_8_v1, public.ore_cllw_8_v1) OWNER TO postgres;

--
-- Name: >; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.> (
    FUNCTION = public.ore_cllw_8_v1_gt_lex,
    LEFTARG = public.ore_cllw_8_variable_v1,
    RIGHTARG = public.ore_cllw_8_variable_v1,
    NEGATOR = OPERATOR(public.<=),
    MERGES,
    HASHES,
    RESTRICT = scalarltsel,
    JOIN = scalarltjoinsel
);


ALTER OPERATOR public.> (public.ore_cllw_8_variable_v1, public.ore_cllw_8_variable_v1) OWNER TO postgres;

--
-- Name: >; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.> (
    FUNCTION = public.cs_ste_vec_encrypted_term_gt,
    LEFTARG = public.cs_ste_vec_encrypted_term_v1,
    RIGHTARG = public.cs_ste_vec_encrypted_term_v1,
    NEGATOR = OPERATOR(public.<=),
    MERGES,
    HASHES,
    RESTRICT = scalarltsel,
    JOIN = scalarltjoinsel
);


ALTER OPERATOR public.> (public.cs_ste_vec_encrypted_term_v1, public.cs_ste_vec_encrypted_term_v1) OWNER TO postgres;

--
-- Name: >=; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.>= (
    FUNCTION = public.ore_64_8_v1_term_gte,
    LEFTARG = public.ore_64_8_v1_term,
    RIGHTARG = public.ore_64_8_v1_term,
    COMMUTATOR = OPERATOR(public.<=),
    NEGATOR = OPERATOR(public.<),
    RESTRICT = scalarlesel,
    JOIN = scalarlejoinsel
);


ALTER OPERATOR public.>= (public.ore_64_8_v1_term, public.ore_64_8_v1_term) OWNER TO postgres;

--
-- Name: >=; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.>= (
    FUNCTION = public.ore_64_8_v1_gte,
    LEFTARG = public.ore_64_8_v1,
    RIGHTARG = public.ore_64_8_v1,
    COMMUTATOR = OPERATOR(public.<=),
    NEGATOR = OPERATOR(public.<),
    RESTRICT = scalarlesel,
    JOIN = scalarlejoinsel
);


ALTER OPERATOR public.>= (public.ore_64_8_v1, public.ore_64_8_v1) OWNER TO postgres;

--
-- Name: >=; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.>= (
    FUNCTION = public.ore_cllw_8_v1_gte,
    LEFTARG = public.ore_cllw_8_v1,
    RIGHTARG = public.ore_cllw_8_v1,
    NEGATOR = OPERATOR(public.<),
    MERGES,
    HASHES,
    RESTRICT = scalarltsel,
    JOIN = scalarltjoinsel
);


ALTER OPERATOR public.>= (public.ore_cllw_8_v1, public.ore_cllw_8_v1) OWNER TO postgres;

--
-- Name: >=; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.>= (
    FUNCTION = public.ore_cllw_8_v1_gte_lex,
    LEFTARG = public.ore_cllw_8_variable_v1,
    RIGHTARG = public.ore_cllw_8_variable_v1,
    NEGATOR = OPERATOR(public.<),
    MERGES,
    HASHES,
    RESTRICT = scalarltsel,
    JOIN = scalarltjoinsel
);


ALTER OPERATOR public.>= (public.ore_cllw_8_variable_v1, public.ore_cllw_8_variable_v1) OWNER TO postgres;

--
-- Name: >=; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.>= (
    FUNCTION = public.cs_ste_vec_encrypted_term_gte,
    LEFTARG = public.cs_ste_vec_encrypted_term_v1,
    RIGHTARG = public.cs_ste_vec_encrypted_term_v1,
    NEGATOR = OPERATOR(public.<),
    MERGES,
    HASHES,
    RESTRICT = scalarltsel,
    JOIN = scalarltjoinsel
);


ALTER OPERATOR public.>= (public.cs_ste_vec_encrypted_term_v1, public.cs_ste_vec_encrypted_term_v1) OWNER TO postgres;

--
-- Name: @>; Type: OPERATOR; Schema: public; Owner: postgres
--

CREATE OPERATOR public.@> (
    FUNCTION = public.ste_vec_v1_logical_contains,
    LEFTARG = public.cs_ste_vec_index_v1,
    RIGHTARG = public.cs_ste_vec_index_v1,
    COMMUTATOR = OPERATOR(public.<@)
);


ALTER OPERATOR public.@> (public.cs_ste_vec_index_v1, public.cs_ste_vec_index_v1) OWNER TO postgres;

--
-- Name: cs_ste_vec_encrypted_term_v1_btree_ops; Type: OPERATOR FAMILY; Schema: public; Owner: postgres
--

CREATE OPERATOR FAMILY public.cs_ste_vec_encrypted_term_v1_btree_ops USING btree;


ALTER OPERATOR FAMILY public.cs_ste_vec_encrypted_term_v1_btree_ops USING btree OWNER TO postgres;

--
-- Name: cs_ste_vec_encrypted_term_v1_btree_ops; Type: OPERATOR CLASS; Schema: public; Owner: postgres
--

CREATE OPERATOR CLASS public.cs_ste_vec_encrypted_term_v1_btree_ops
    DEFAULT FOR TYPE public.cs_ste_vec_encrypted_term_v1 USING btree FAMILY public.cs_ste_vec_encrypted_term_v1_btree_ops AS
    OPERATOR 1 public.<(public.cs_ste_vec_encrypted_term_v1,public.cs_ste_vec_encrypted_term_v1) ,
    OPERATOR 2 public.<=(public.cs_ste_vec_encrypted_term_v1,public.cs_ste_vec_encrypted_term_v1) ,
    OPERATOR 3 public.=(public.cs_ste_vec_encrypted_term_v1,public.cs_ste_vec_encrypted_term_v1) ,
    OPERATOR 4 public.>=(public.cs_ste_vec_encrypted_term_v1,public.cs_ste_vec_encrypted_term_v1) ,
    OPERATOR 5 public.>(public.cs_ste_vec_encrypted_term_v1,public.cs_ste_vec_encrypted_term_v1) ,
    FUNCTION 1 (public.cs_ste_vec_encrypted_term_v1, public.cs_ste_vec_encrypted_term_v1) public.compare_ste_vec_encrypted_term_v1(public.cs_ste_vec_encrypted_term_v1,public.cs_ste_vec_encrypted_term_v1);


ALTER OPERATOR CLASS public.cs_ste_vec_encrypted_term_v1_btree_ops USING btree OWNER TO postgres;

--
-- Name: ore_64_8_v1_btree_ops; Type: OPERATOR FAMILY; Schema: public; Owner: postgres
--

CREATE OPERATOR FAMILY public.ore_64_8_v1_btree_ops USING btree;


ALTER OPERATOR FAMILY public.ore_64_8_v1_btree_ops USING btree OWNER TO postgres;

--
-- Name: ore_64_8_v1_btree_ops; Type: OPERATOR CLASS; Schema: public; Owner: postgres
--

CREATE OPERATOR CLASS public.ore_64_8_v1_btree_ops
    DEFAULT FOR TYPE public.ore_64_8_v1 USING btree FAMILY public.ore_64_8_v1_btree_ops AS
    OPERATOR 1 public.<(public.ore_64_8_v1,public.ore_64_8_v1) ,
    OPERATOR 2 public.<=(public.ore_64_8_v1,public.ore_64_8_v1) ,
    OPERATOR 3 public.=(public.ore_64_8_v1,public.ore_64_8_v1) ,
    OPERATOR 4 public.>=(public.ore_64_8_v1,public.ore_64_8_v1) ,
    OPERATOR 5 public.>(public.ore_64_8_v1,public.ore_64_8_v1) ,
    FUNCTION 1 (public.ore_64_8_v1, public.ore_64_8_v1) public.compare_ore_64_8_v1(public.ore_64_8_v1,public.ore_64_8_v1);


ALTER OPERATOR CLASS public.ore_64_8_v1_btree_ops USING btree OWNER TO postgres;

--
-- Name: ore_64_8_v1_term_btree_ops; Type: OPERATOR FAMILY; Schema: public; Owner: postgres
--

CREATE OPERATOR FAMILY public.ore_64_8_v1_term_btree_ops USING btree;


ALTER OPERATOR FAMILY public.ore_64_8_v1_term_btree_ops USING btree OWNER TO postgres;

--
-- Name: ore_64_8_v1_term_btree_ops; Type: OPERATOR CLASS; Schema: public; Owner: postgres
--

CREATE OPERATOR CLASS public.ore_64_8_v1_term_btree_ops
    DEFAULT FOR TYPE public.ore_64_8_v1_term USING btree FAMILY public.ore_64_8_v1_term_btree_ops AS
    OPERATOR 1 public.<(public.ore_64_8_v1_term,public.ore_64_8_v1_term) ,
    OPERATOR 2 public.<=(public.ore_64_8_v1_term,public.ore_64_8_v1_term) ,
    OPERATOR 3 public.=(public.ore_64_8_v1_term,public.ore_64_8_v1_term) ,
    OPERATOR 4 public.>=(public.ore_64_8_v1_term,public.ore_64_8_v1_term) ,
    OPERATOR 5 public.>(public.ore_64_8_v1_term,public.ore_64_8_v1_term) ,
    FUNCTION 1 (public.ore_64_8_v1_term, public.ore_64_8_v1_term) public.compare_ore_64_8_v1_term(public.ore_64_8_v1_term,public.ore_64_8_v1_term);


ALTER OPERATOR CLASS public.ore_64_8_v1_term_btree_ops USING btree OWNER TO postgres;

--
-- Name: ore_cllw_8_v1_btree_ops; Type: OPERATOR FAMILY; Schema: public; Owner: postgres
--

CREATE OPERATOR FAMILY public.ore_cllw_8_v1_btree_ops USING btree;


ALTER OPERATOR FAMILY public.ore_cllw_8_v1_btree_ops USING btree OWNER TO postgres;

--
-- Name: ore_cllw_8_v1_btree_ops; Type: OPERATOR CLASS; Schema: public; Owner: postgres
--

CREATE OPERATOR CLASS public.ore_cllw_8_v1_btree_ops
    DEFAULT FOR TYPE public.ore_cllw_8_v1 USING btree FAMILY public.ore_cllw_8_v1_btree_ops AS
    OPERATOR 1 public.<(public.ore_cllw_8_v1,public.ore_cllw_8_v1) ,
    OPERATOR 2 public.<=(public.ore_cllw_8_v1,public.ore_cllw_8_v1) ,
    OPERATOR 3 public.=(public.ore_cllw_8_v1,public.ore_cllw_8_v1) ,
    OPERATOR 4 public.>=(public.ore_cllw_8_v1,public.ore_cllw_8_v1) ,
    OPERATOR 5 public.>(public.ore_cllw_8_v1,public.ore_cllw_8_v1) ,
    FUNCTION 1 (public.ore_cllw_8_v1, public.ore_cllw_8_v1) public.compare_ore_cllw_8_v1(public.ore_cllw_8_v1,public.ore_cllw_8_v1);


ALTER OPERATOR CLASS public.ore_cllw_8_v1_btree_ops USING btree OWNER TO postgres;

--
-- Name: ore_cllw_8_v1_variable_btree_ops; Type: OPERATOR FAMILY; Schema: public; Owner: postgres
--

CREATE OPERATOR FAMILY public.ore_cllw_8_v1_variable_btree_ops USING btree;


ALTER OPERATOR FAMILY public.ore_cllw_8_v1_variable_btree_ops USING btree OWNER TO postgres;

--
-- Name: ore_cllw_8_v1_variable_btree_ops; Type: OPERATOR CLASS; Schema: public; Owner: postgres
--

CREATE OPERATOR CLASS public.ore_cllw_8_v1_variable_btree_ops
    DEFAULT FOR TYPE public.ore_cllw_8_variable_v1 USING btree FAMILY public.ore_cllw_8_v1_variable_btree_ops AS
    OPERATOR 1 public.<(public.ore_cllw_8_variable_v1,public.ore_cllw_8_variable_v1) ,
    OPERATOR 2 public.<=(public.ore_cllw_8_variable_v1,public.ore_cllw_8_variable_v1) ,
    OPERATOR 3 public.=(public.ore_cllw_8_variable_v1,public.ore_cllw_8_variable_v1) ,
    OPERATOR 4 public.>=(public.ore_cllw_8_variable_v1,public.ore_cllw_8_variable_v1) ,
    OPERATOR 5 public.>(public.ore_cllw_8_variable_v1,public.ore_cllw_8_variable_v1) ,
    FUNCTION 1 (public.ore_cllw_8_variable_v1, public.ore_cllw_8_variable_v1) public.compare_lex_ore_cllw_8_v1(public.ore_cllw_8_variable_v1,public.ore_cllw_8_variable_v1);


ALTER OPERATOR CLASS public.ore_cllw_8_v1_variable_btree_ops USING btree OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: cs_configuration_v1; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cs_configuration_v1 (
    id bigint NOT NULL,
    state public.cs_configuration_state_v1 DEFAULT 'pending'::public.cs_configuration_state_v1 NOT NULL,
    data public.cs_configuration_data_v1,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.cs_configuration_v1 OWNER TO postgres;

--
-- Name: cs_configuration_v1_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.cs_configuration_v1 ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.cs_configuration_v1_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: pycon_cta; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.pycon_cta (
    id integer NOT NULL,
    key public.cs_encrypted_v1
);


ALTER TABLE public.pycon_cta OWNER TO postgres;

--
-- Name: pycon_cta_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.pycon_cta_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.pycon_cta_id_seq OWNER TO postgres;

--
-- Name: pycon_cta_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.pycon_cta_id_seq OWNED BY public.pycon_cta.id;


--
-- Name: pycon_cta id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pycon_cta ALTER COLUMN id SET DEFAULT nextval('public.pycon_cta_id_seq'::regclass);


--
-- Data for Name: cs_configuration_v1; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.cs_configuration_v1 (id, state, data, created_at) FROM stdin;
1	active	{"v": 1, "tables": {"pycon_cta": {"key": {"cast_as": "text", "indexes": {"match": {"tokenizer": {"kind": "ngram", "token_length": 3}, "token_filters": [{"kind": "downcase"}]}}}}}}	2024-11-19 17:14:25.679181+00
\.


--
-- Data for Name: pycon_cta; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pycon_cta (id, key) FROM stdin;
1	{"c": "mBbKu90y1ViL{q%eI+{V`PntZKwD0VP-O+wBX_m&IJ5}#$pezU!JIWc;G=a)-F%om4O3y};yarKk$1ci0Y2CwTBkv+JiJ@aks5}}Zn=2xs>C3mQ!4GM4=tem^%nhYI~EPm;XIey=%#$b?$L?vuz?JtYh`)B", "i": {"c": "key", "t": "pycon_cta"}, "k": "ct", "m": [813, 192, 971, 1255, 909, 1012, 1456, 1678, 823, 1606, 1013, 1543, 1409, 80, 1636, 2040, 1816, 457, 327, 509, 1492, 54, 1206, 1935, 691, 1263, 1198, 1815, 1908, 1227, 724, 1272, 1069, 1146, 1021, 439, 1917, 1313, 359, 1091, 191, 1190, 1594, 1163, 686, 1035, 1614, 1117, 597, 272, 23, 1883, 1326, 1082, 1073, 90, 871, 450, 1124, 492, 1592, 1461, 57, 1841, 1722, 1241, 153, 145, 1256, 518, 933, 1706, 181, 790, 1639, 1345, 745, 770, 523, 1881, 590, 1998, 490, 1739, 371, 1637, 1308, 473, 1491, 903, 1310, 2018, 225, 1921, 1750, 651, 268, 692, 1250, 1134, 1087, 696, 1977, 1853, 1862, 789, 477, 1628, 66, 531, 1161, 1577, 1659, 278, 121, 547, 1365, 584, 1674, 551, 1608, 274, 1213, 129, 336, 398, 1509, 316, 1222, 1842, 1164, 884, 2003, 1240, 1174, 1104, 1611, 63, 300, 1700, 806, 1458, 422, 917, 716, 1289, 608, 1824, 1138, 1880, 131, 261, 687, 623, 1399, 1618, 1066, 1056, 1956, 1397, 680, 1666, 1529, 1656, 1645, 1781, 273, 1695, 1349, 1534, 1912, 1680, 751, 501, 1889, 1501, 1817, 1705, 840, 659, 2029, 2, 2014, 469, 136, 923, 202, 1224, 1508, 786, 835, 1922, 1297, 1763, 1643, 1070, 1488, 838, 70, 1281, 761, 1179, 1975, 1968, 1619, 936, 1242, 305, 48, 646, 1370, 780, 1305, 982, 855, 1676, 1201, 843, 1893], "o": null, "u": null, "v": 1}
\.


--
-- Name: cs_configuration_v1_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.cs_configuration_v1_id_seq', 1, true);


--
-- Name: pycon_cta_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.pycon_cta_id_seq', 1, true);


--
-- Name: cs_configuration_v1 cs_configuration_v1_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cs_configuration_v1
    ADD CONSTRAINT cs_configuration_v1_pkey PRIMARY KEY (id);


--
-- Name: pycon_cta pycon_cta_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pycon_cta
    ADD CONSTRAINT pycon_cta_pkey PRIMARY KEY (id);


--
-- Name: cs_configuration_v1_index_active; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX cs_configuration_v1_index_active ON public.cs_configuration_v1 USING btree (state) WHERE (state = 'active'::public.cs_configuration_state_v1);


--
-- Name: cs_configuration_v1_index_encrypting; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX cs_configuration_v1_index_encrypting ON public.cs_configuration_v1 USING btree (state) WHERE (state = 'encrypting'::public.cs_configuration_state_v1);


--
-- Name: cs_configuration_v1_index_pending; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX cs_configuration_v1_index_pending ON public.cs_configuration_v1 USING btree (state) WHERE (state = 'pending'::public.cs_configuration_state_v1);


--
-- PostgreSQL database dump complete
--

