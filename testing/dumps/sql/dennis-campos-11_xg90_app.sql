-- Downloaded from: https://github.com/dennis-campos-11/xg90_app/blob/6987535ab3c9cc19cf58221c2f8a00b0f9e6973b/db/structure.sql
\restrict EdWRZenb7mFdNu2UgQVaersivDeqiv9OblwKlhJaxPhij60hrIzkbhoMS6kduPo

-- Dumped from database version 15.14 (Postgres.app)
-- Dumped by pg_dump version 15.14 (Postgres.app)

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
-- Name: process_fixture_list_data(jsonb); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.process_fixture_list_data(params jsonb) RETURNS TABLE(team_id integer, fixture_id integer, team_location integer, processed_stats jsonb, processed_facts jsonb)
    LANGUAGE sql STABLE
    AS $$
      WITH single_fixture_list AS MATERIALIZED (
        SELECT
          params->>'id' AS id,
          params->>'name' AS name,
          (params->>'total_matches')::int AS total_matches,
          (params->>'season_index')::int AS season_index,
          (params->>'home_location')::int AS home_location,
          (params->>'away_location')::int AS away_location,
          (params->>'settings')::jsonb AS settings,
          (params->>'fixture_date')::int AS fixture_date,
          COALESCE(params->'fixture_list_fields_attributes', params->'fixture_list_fields') AS fixture_list_fields,
          COALESCE(params->'fixture_list_competitions_attributes', params->'fixture_list_competitions') AS fixture_list_competitions,
          params->'sort' AS sort
      ),
      sort_params AS MATERIALIZED (
        SELECT
          sort->>'field_code' AS field_code,
          sort->>'field_type' AS field_type,
          sort->>'metric' AS metric,
          (sort->>'location')::int AS location,
          COALESCE(sort->>'direction', 'asc') AS direction
        FROM single_fixture_list
      ),
      total_matches_value AS MATERIALIZED (
        SELECT total_matches FROM single_fixture_list
      ),
      fixture_date_value AS MATERIALIZED (
        SELECT fixture_date FROM single_fixture_list
      ),
      selected_competitions AS MATERIALIZED (
        SELECT (comp->>'competition_id')::int AS competition_id
        FROM single_fixture_list sfl,
            LATERAL jsonb_array_elements(sfl.fixture_list_competitions) comp
      ),
      upcoming_fixtures AS MATERIALIZED (
        SELECT id, starting_at, home_id, away_id, competition_id
        FROM fixtures
        WHERE 
          starting_at >= CASE (SELECT fixture_date FROM fixture_date_value)
            WHEN 0 THEN now()
            WHEN 1 THEN date_trunc('day', now())
            WHEN 2 THEN date_trunc('day', now() + interval '1 day')
            WHEN 3 THEN date_trunc('day', now() + interval '2 day')
            WHEN 4 THEN date_trunc('day', now() + interval '3 day')
            ELSE now()
          END
          AND starting_at < CASE (SELECT fixture_date FROM fixture_date_value)
            WHEN 0 THEN 'infinity'::timestamp
            WHEN 1 THEN date_trunc('day', now()) + interval '1 day'
            WHEN 2 THEN date_trunc('day', now() + interval '1 day') + interval '1 day'
            WHEN 3 THEN date_trunc('day', now() + interval '2 day') + interval '1 day'
            WHEN 4 THEN date_trunc('day', now() + interval '3 day') + interval '1 day'
            ELSE 'infinity'::timestamp
          END
          AND (
            EXISTS (SELECT 1 FROM selected_competitions)
              AND competition_id IN (SELECT competition_id FROM selected_competitions)
            OR
            NOT EXISTS (SELECT 1 FROM selected_competitions)
          )
      ),
      input_teams AS MATERIALIZED (
        SELECT DISTINCT
          f.home_id AS team_id,
          f.starting_at AS starting_at,
          1 AS team_location,
          sfl.home_location AS data_location,
          f.id AS fixture_id,
          f.competition_id
        FROM upcoming_fixtures f
        CROSS JOIN single_fixture_list sfl

        UNION

        SELECT DISTINCT
          f.away_id AS team_id,
          f.starting_at AS starting_at,
          2 AS team_location,
          sfl.away_location AS data_location,
          f.id AS fixture_id,
          f.competition_id
        FROM upcoming_fixtures f
        CROSS JOIN single_fixture_list sfl
      ),
      fields AS MATERIALIZED (
        SELECT
          COALESCE(fl->>'field_code', df.code) AS field_code,
          COALESCE((fl->>'field_type')::NUMERIC, df.field_type) AS field_type,
          (fl->>'filters')::JSONB AS filters
        FROM single_fixture_list sfl
        CROSS JOIN LATERAL jsonb_array_elements(
          sfl.fixture_list_fields ||
          '[{"field_code":"minutes_on_field_ft","field_type":1,"filters":{}},
            {"field_code":"minutes_on_field_1h","field_type":1,"filters":{}},
            {"field_code":"minutes_on_field_2h","field_type":1,"filters":{}}]'::jsonb
        ) AS fl
        LEFT JOIN data_fields df
          ON df.id = (fl->>'data_field_id')::INT
          OR df.code = fl->>'field_code'
      ),
      recent_fixtures AS MATERIALIZED (
        SELECT
          t.team_id,
          t.team_location,
          t.fixture_id,
          f.starting_at,
          CASE WHEN t.team_id = f.home_id THEN f.home_stats ELSE f.away_stats END AS stats,
          CASE WHEN t.team_id = f.home_id THEN f.home_facts ELSE f.away_facts END AS facts,
          CASE WHEN t.team_id = f.home_id THEN f.away_stats ELSE f.home_stats END AS opponent_stats,
          CASE WHEN t.team_id = f.home_id THEN f.away_facts ELSE f.home_facts END AS opponent_facts,
          CASE WHEN t.team_id = f.home_id THEN 'home' ELSE 'away' END AS location,
          CASE WHEN t.team_id = f.home_id THEN f.away_id ELSE f.home_id END AS opponent_id,
          opponent.name AS opponent_name
        FROM input_teams t
        JOIN LATERAL (
          SELECT fi.*
          FROM (
            SELECT fi.*
            FROM fixtures fi
            JOIN seasons s ON fi.season_id = s.id
            CROSS JOIN single_fixture_list sfl
            WHERE fi.status = 1
              AND (sfl.season_index IS NULL OR s.index <= sfl.season_index)
              AND fi.home_id = t.team_id
              AND ((sfl.settings->'general'->>'only_current_competition')::boolean IS NOT TRUE OR fi.competition_id = t.competition_id)

            UNION ALL

            SELECT fi.*
            FROM fixtures fi
            JOIN seasons s ON fi.season_id = s.id
            CROSS JOIN single_fixture_list sfl
            WHERE fi.status = 1
              AND (sfl.season_index IS NULL OR s.index <= sfl.season_index)
              AND fi.away_id = t.team_id
              AND ((sfl.settings->'general'->>'only_current_competition')::boolean IS NOT TRUE OR fi.competition_id = t.competition_id)
          ) fi
          WHERE (t.data_location IS NULL OR t.data_location NOT IN (1,2)) 
                OR (t.data_location = 1 AND fi.home_id = t.team_id)
                OR (t.data_location = 2 AND fi.away_id = t.team_id)
                AND (
                  NOT (SELECT (settings->'general'->>'only_current_competition')::boolean FROM single_fixture_list)
                  OR fi.competition_id = t.competition_id
                )
          ORDER BY fi.starting_at DESC
          LIMIT (SELECT total_matches FROM total_matches_value)
        ) f ON true
        JOIN teams opponent ON opponent.id = CASE WHEN t.team_id = f.home_id THEN f.away_id ELSE f.home_id END
      ),
      stats_agg AS MATERIALIZED (
        SELECT
          rf.team_id,
          rf.team_location,
          rf.fixture_id,
          f1.field_code,
          COUNT(*) AS games_played,
          ROUND(AVG((rf.stats ->> f1.field_code)::NUMERIC), 2) AS average,
          ROUND(SUM((rf.stats ->> f1.field_code)::NUMERIC), 2) AS total,
          ROUND((
            SUM((rf.stats ->> f1.field_code)::NUMERIC) /
            (
              SUM((rf.stats ->> ('minutes_on_field_' || RIGHT(f1.field_code, 2)))::NUMERIC) /
              CASE
                WHEN RIGHT(f1.field_code, 2) = 'ft' THEN 90.0
                WHEN RIGHT(f1.field_code, 2) IN ('1h', '2h') THEN 45.0
                ELSE 1.0
              END
            )
          ), 2) AS average_by_period
        FROM recent_fixtures rf
        JOIN fields f1 ON f1.field_type = 1
        GROUP BY rf.team_id, rf.team_location, rf.fixture_id, f1.field_code
      ),
      stats_agg_extras AS MATERIALIZED (
        SELECT
          s.fixture_id,
          s.field_code,
          ROUND(SUM(average), 2) AS overall,
          ROUND(SUM(average_by_period), 2) AS overall_by_period
        FROM stats_agg s
        WHERE s.fixture_id IS NOT NULL
        GROUP BY s.fixture_id, s.field_code
      ),
      filtered_stats AS (
        SELECT s.*, se.overall, se.overall_by_period
        FROM stats_agg s
        JOIN stats_agg_extras se
          ON se.fixture_id = s.fixture_id AND s.field_code = se.field_code
        LEFT JOIN fields f ON s.field_code = f.field_code AND f.field_type = 1
        WHERE
          se.overall BETWEEN COALESCE(CASE s.team_location WHEN 1 THEN NULLIF(f.filters->'overall'->'home'->>'from', '')::NUMERIC WHEN 2 THEN NULLIF(f.filters->'overall'->'away'->>'from', '')::NUMERIC ELSE NULL END, 0)
                          AND COALESCE(CASE s.team_location WHEN 1 THEN NULLIF(f.filters->'overall'->'home'->>'to', '')::NUMERIC WHEN 2 THEN NULLIF(f.filters->'overall'->'away'->>'to', '')::NUMERIC ELSE NULL END, 'infinity')
          AND se.overall_by_period BETWEEN COALESCE(CASE s.team_location WHEN 1 THEN NULLIF(f.filters->'overall_by_period'->'home'->>'from', '')::NUMERIC WHEN 2 THEN NULLIF(f.filters->'overall_by_period'->'away'->>'from', '')::NUMERIC ELSE NULL END, 0)
                                    AND COALESCE(CASE s.team_location WHEN 1 THEN NULLIF(f.filters->'overall_by_period'->'home'->>'to', '')::NUMERIC WHEN 2 THEN NULLIF(f.filters->'overall_by_period'->'away'->>'to', '')::NUMERIC ELSE NULL END, 'infinity')
          AND s.average BETWEEN COALESCE(CASE s.team_location WHEN 1 THEN NULLIF(f.filters->'average'->'home'->>'from', '')::NUMERIC WHEN 2 THEN NULLIF(f.filters->'average'->'away'->>'from', '')::NUMERIC ELSE NULL END, 0)
                          AND COALESCE(CASE s.team_location WHEN 1 THEN NULLIF(f.filters->'average'->'home'->>'to', '')::NUMERIC WHEN 2 THEN NULLIF(f.filters->'average'->'away'->>'to', '')::NUMERIC ELSE NULL END, 'infinity')
          AND s.average_by_period BETWEEN COALESCE(CASE s.team_location WHEN 1 THEN NULLIF(f.filters->'average_by_period'->'home'->>'from', '')::NUMERIC WHEN 2 THEN NULLIF(f.filters->'average_by_period'->'away'->>'from', '')::NUMERIC ELSE NULL END, 0)
                                    AND COALESCE(CASE s.team_location WHEN 1 THEN NULLIF(f.filters->'average_by_period'->'home'->>'to', '')::NUMERIC WHEN 2 THEN NULLIF(f.filters->'average_by_period'->'away'->>'to', '')::NUMERIC ELSE NULL END, 'infinity')
      ),
      facts_agg AS MATERIALIZED (
        SELECT
          rf.team_id,
          rf.team_location,
          rf.fixture_id,
          f2.field_code,
          COUNT(*) AS games_played,
          SUM((rf.facts ->> f2.field_code)::NUMERIC) AS total,
          ROUND(AVG((rf.facts ->> f2.field_code)::NUMERIC * 100), 0) AS percentage,
          COALESCE(
            ARRAY_POSITION(
              ARRAY_APPEND(ARRAY_AGG((rf.facts ->> f2.field_code)::int ORDER BY rf.starting_at ASC), 0),
              0
            ) - 1, 0
          ) AS streak
        FROM recent_fixtures rf
        JOIN fields f2 ON f2.field_type = 2
        GROUP BY rf.team_id, rf.team_location, rf.fixture_id, f2.field_code
      ),
      facts_agg_extras AS MATERIALIZED (
        SELECT
          f.fixture_id,
          f.field_code,
          ROUND(AVG(percentage), 2) AS average
        FROM facts_agg f
        WHERE f.fixture_id IS NOT NULL
        GROUP BY f.fixture_id, f.field_code
      ),
      filtered_facts AS MATERIALIZED (
        SELECT fa.*, fe.average
        FROM facts_agg fa
        JOIN facts_agg_extras fe
          ON fe.fixture_id = fa.fixture_id AND fa.field_code = fe.field_code
        LEFT JOIN fields f ON fa.field_code = f.field_code AND f.field_type = 2
        WHERE
          fa.percentage BETWEEN COALESCE(CASE fa.team_location WHEN 1 THEN NULLIF(f.filters->'percentage'->'home'->>'from', '')::NUMERIC WHEN 2 THEN NULLIF(f.filters->'percentage'->'away'->>'from', '')::NUMERIC ELSE NULL END, 0)
                          AND COALESCE(CASE fa.team_location WHEN 1 THEN NULLIF(f.filters->'percentage'->'home'->>'to', '')::NUMERIC WHEN 2 THEN NULLIF(f.filters->'percentage'->'away'->>'to', '')::NUMERIC ELSE NULL END, 'infinity')
          AND fa.total BETWEEN COALESCE(CASE fa.team_location WHEN 1 THEN NULLIF(f.filters->'total'->'home'->>'from', '')::NUMERIC WHEN 2 THEN NULLIF(f.filters->'total'->'away'->>'from', '')::NUMERIC ELSE NULL END, 0)
                          AND COALESCE(CASE fa.team_location WHEN 1 THEN NULLIF(f.filters->'total'->'home'->>'to', '')::NUMERIC WHEN 2 THEN NULLIF(f.filters->'total'->'away'->>'to', '')::NUMERIC ELSE NULL END, 'infinity')
          AND fa.streak BETWEEN COALESCE(CASE fa.team_location WHEN 1 THEN NULLIF(f.filters->'streak'->'home'->>'from', '')::NUMERIC WHEN 2 THEN NULLIF(f.filters->'streak'->'away'->>'from', '')::NUMERIC ELSE NULL END, 0)
                          AND COALESCE(CASE fa.team_location WHEN 1 THEN NULLIF(f.filters->'streak'->'home'->>'to', '')::NUMERIC WHEN 2 THEN NULLIF(f.filters->'streak'->'away'->>'to', '')::NUMERIC ELSE NULL END, 'infinity')
      ),
      field_count AS MATERIALIZED (
        SELECT COUNT(DISTINCT field_code) AS total_fields FROM fields
      ),
      stats_max AS MATERIALIZED (
        SELECT
          s.field_code,
          MAX(s.average) AS max_average,
          MAX(s.overall) AS max_overall,
          MAX(s.overall_by_period) AS max_overall_by_period,
          MAX(s.average_by_period) AS max_average_by_period,
          MAX(s.total) AS max_total
        FROM filtered_stats s
        GROUP BY s.field_code
      ),
      facts_max AS MATERIALIZED (
        SELECT
          f.field_code,
          MAX(f.average) AS max_average,
          MAX(f.percentage) AS max_percentage,
          MAX(f.total) AS max_total,
          MAX(f.streak) AS max_streak
        FROM filtered_facts f
        GROUP BY f.field_code
      ),
      final_data AS MATERIALIZED (
        SELECT
          t.starting_at,
          t.team_id,
          t.fixture_id,
          COALESCE(s.team_location, f.team_location) AS team_location,
          JSONB_AGG(DISTINCT JSONB_BUILD_OBJECT(
            'field_code', s.field_code,
			      'team_location', s.team_location,
            'games_played', s.games_played,
            'overall', s.overall::FLOAT8,
            'overall_by_period', s.overall_by_period::FLOAT8,
            'total', s.total::FLOAT8,
            'average', s.average::FLOAT8,
            'average_by_period', s.average_by_period::FLOAT8,
            'max', JSONB_BUILD_OBJECT(
              'overall', sm.max_overall::FLOAT8,
              'overall_by_period', sm.max_overall_by_period::FLOAT8,
              'total', sm.max_total::FLOAT8,
              'average', sm.max_average::FLOAT8,
              'average_by_period', sm.max_average_by_period::FLOAT8
            )
          )) AS processed_stats,
          JSONB_AGG(DISTINCT JSONB_BUILD_OBJECT(
            'field_code', f.field_code,
			      'team_location', f.team_location,
            'games_played', f.games_played,
            'average', f.average::FLOAT8,
            'total', f.total::FLOAT8,
            'percentage', f.percentage::FLOAT8,
            'streak', f.streak::FLOAT8,
            'max', JSONB_BUILD_OBJECT(
              'average', fm.max_average::FLOAT8,
              'total', fm.max_total::FLOAT8,
              'percentage', fm.max_percentage::FLOAT8,
              'streak', fm.max_streak::FLOAT8
            )
          )) AS processed_facts
        FROM input_teams t
        LEFT JOIN filtered_stats s 
          ON s.team_id = t.team_id AND s.field_code IS NOT NULL
        LEFT JOIN stats_max sm 
          ON sm.field_code = s.field_code
        LEFT JOIN filtered_facts f 
          ON f.team_id = t.team_id AND f.field_code IS NOT NULL
        LEFT JOIN facts_max fm 
          ON fm.field_code = f.field_code
        GROUP BY t.starting_at, t.fixture_id, t.team_id, s.team_location, f.team_location
        HAVING 
          (COUNT(DISTINCT s.field_code) + COUNT(DISTINCT f.field_code)) = (SELECT total_fields FROM field_count)
      )
      SELECT
        team_id,
        fixture_id,
        team_location,
        processed_stats,
        processed_facts
      FROM final_data
      $$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: ar_internal_metadata; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ar_internal_metadata (
    key character varying NOT NULL,
    value character varying,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: competitions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.competitions (
    id bigint NOT NULL,
    name character varying,
    external_ws_id bigint,
    country_id bigint NOT NULL,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: competitions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.competitions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: competitions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.competitions_id_seq OWNED BY public.competitions.id;


--
-- Name: countries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.countries (
    id bigint NOT NULL,
    name character varying,
    external_ws_id bigint,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: countries_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.countries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: countries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.countries_id_seq OWNED BY public.countries.id;


--
-- Name: data_fields; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.data_fields (
    id bigint NOT NULL,
    code character varying,
    field_type integer,
    half_type integer,
    settings jsonb DEFAULT '{}'::jsonb,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: data_fields_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.data_fields_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: data_fields_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.data_fields_id_seq OWNED BY public.data_fields.id;


--
-- Name: fixture_list_competitions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.fixture_list_competitions (
    id bigint NOT NULL,
    fixture_list_id bigint NOT NULL,
    competition_id bigint NOT NULL,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: fixture_list_competitions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.fixture_list_competitions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: fixture_list_competitions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.fixture_list_competitions_id_seq OWNED BY public.fixture_list_competitions.id;


--
-- Name: fixture_list_fields; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.fixture_list_fields (
    id bigint NOT NULL,
    fixture_list_id bigint NOT NULL,
    data_field_id bigint NOT NULL,
    filters jsonb DEFAULT '{}'::jsonb,
    index integer,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: fixture_list_fields_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.fixture_list_fields_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: fixture_list_fields_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.fixture_list_fields_id_seq OWNED BY public.fixture_list_fields.id;


--
-- Name: fixture_lists; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.fixture_lists (
    id bigint NOT NULL,
    name character varying,
    home_location integer,
    away_location integer,
    total_matches integer,
    season_index integer,
    sort jsonb DEFAULT '{}'::jsonb,
    settings jsonb DEFAULT '{}'::jsonb,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: fixture_lists_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.fixture_lists_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: fixture_lists_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.fixture_lists_id_seq OWNED BY public.fixture_lists.id;


--
-- Name: fixtures; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.fixtures (
    id bigint NOT NULL,
    competition_id bigint,
    season_id bigint,
    home_id bigint,
    away_id bigint,
    external_ws_id bigint,
    starting_at timestamp(6) without time zone,
    status integer,
    home_stats jsonb DEFAULT '{}'::jsonb,
    away_stats jsonb DEFAULT '{}'::jsonb,
    home_facts jsonb DEFAULT '{}'::jsonb,
    away_facts jsonb DEFAULT '{}'::jsonb,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: fixtures_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.fixtures_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: fixtures_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.fixtures_id_seq OWNED BY public.fixtures.id;


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying NOT NULL
);


--
-- Name: season_teams; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.season_teams (
    id bigint NOT NULL,
    season_id bigint NOT NULL,
    team_id bigint NOT NULL,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: season_teams_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.season_teams_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: season_teams_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.season_teams_id_seq OWNED BY public.season_teams.id;


--
-- Name: seasons; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.seasons (
    id bigint NOT NULL,
    name character varying,
    external_ws_id bigint,
    competition_id bigint NOT NULL,
    index integer,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: seasons_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.seasons_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: seasons_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.seasons_id_seq OWNED BY public.seasons.id;


--
-- Name: teams; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.teams (
    id bigint NOT NULL,
    name character varying,
    common_name character varying,
    short_name character varying,
    external_ws_id bigint,
    primary_color character varying,
    text_color character varying,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: teams_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.teams_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: teams_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.teams_id_seq OWNED BY public.teams.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    email character varying,
    password_digest character varying,
    first_name character varying,
    last_name character varying,
    language character varying DEFAULT 'en'::character varying,
    appearance character varying DEFAULT 'light'::character varying,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL,
    reset_password_token character varying,
    reset_password_sent_at timestamp(6) without time zone
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: competitions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.competitions ALTER COLUMN id SET DEFAULT nextval('public.competitions_id_seq'::regclass);


--
-- Name: countries id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.countries ALTER COLUMN id SET DEFAULT nextval('public.countries_id_seq'::regclass);


--
-- Name: data_fields id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.data_fields ALTER COLUMN id SET DEFAULT nextval('public.data_fields_id_seq'::regclass);


--
-- Name: fixture_list_competitions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fixture_list_competitions ALTER COLUMN id SET DEFAULT nextval('public.fixture_list_competitions_id_seq'::regclass);


--
-- Name: fixture_list_fields id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fixture_list_fields ALTER COLUMN id SET DEFAULT nextval('public.fixture_list_fields_id_seq'::regclass);


--
-- Name: fixture_lists id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fixture_lists ALTER COLUMN id SET DEFAULT nextval('public.fixture_lists_id_seq'::regclass);


--
-- Name: fixtures id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fixtures ALTER COLUMN id SET DEFAULT nextval('public.fixtures_id_seq'::regclass);


--
-- Name: season_teams id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.season_teams ALTER COLUMN id SET DEFAULT nextval('public.season_teams_id_seq'::regclass);


--
-- Name: seasons id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seasons ALTER COLUMN id SET DEFAULT nextval('public.seasons_id_seq'::regclass);


--
-- Name: teams id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams ALTER COLUMN id SET DEFAULT nextval('public.teams_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: ar_internal_metadata ar_internal_metadata_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ar_internal_metadata
    ADD CONSTRAINT ar_internal_metadata_pkey PRIMARY KEY (key);


--
-- Name: competitions competitions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.competitions
    ADD CONSTRAINT competitions_pkey PRIMARY KEY (id);


--
-- Name: countries countries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.countries
    ADD CONSTRAINT countries_pkey PRIMARY KEY (id);


--
-- Name: data_fields data_fields_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.data_fields
    ADD CONSTRAINT data_fields_pkey PRIMARY KEY (id);


--
-- Name: fixture_list_competitions fixture_list_competitions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fixture_list_competitions
    ADD CONSTRAINT fixture_list_competitions_pkey PRIMARY KEY (id);


--
-- Name: fixture_list_fields fixture_list_fields_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fixture_list_fields
    ADD CONSTRAINT fixture_list_fields_pkey PRIMARY KEY (id);


--
-- Name: fixture_lists fixture_lists_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fixture_lists
    ADD CONSTRAINT fixture_lists_pkey PRIMARY KEY (id);


--
-- Name: fixtures fixtures_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fixtures
    ADD CONSTRAINT fixtures_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: season_teams season_teams_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.season_teams
    ADD CONSTRAINT season_teams_pkey PRIMARY KEY (id);


--
-- Name: seasons seasons_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seasons
    ADD CONSTRAINT seasons_pkey PRIMARY KEY (id);


--
-- Name: teams teams_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_on_fixture_list_id_competition_id_7e158515e7; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_on_fixture_list_id_competition_id_7e158515e7 ON public.fixture_list_competitions USING btree (fixture_list_id, competition_id);


--
-- Name: index_competitions_on_country_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_competitions_on_country_id ON public.competitions USING btree (country_id);


--
-- Name: index_competitions_on_external_ws_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_competitions_on_external_ws_id ON public.competitions USING btree (external_ws_id);


--
-- Name: index_countries_on_external_ws_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_countries_on_external_ws_id ON public.countries USING btree (external_ws_id);


--
-- Name: index_data_fields_on_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_data_fields_on_code ON public.data_fields USING btree (code);


--
-- Name: index_data_fields_on_code_and_field_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_data_fields_on_code_and_field_type ON public.data_fields USING btree (code, field_type);


--
-- Name: index_fixture_list_competitions_on_competition_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_fixture_list_competitions_on_competition_id ON public.fixture_list_competitions USING btree (competition_id);


--
-- Name: index_fixture_list_competitions_on_fixture_list_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_fixture_list_competitions_on_fixture_list_id ON public.fixture_list_competitions USING btree (fixture_list_id);


--
-- Name: index_fixture_list_fields_on_data_field_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_fixture_list_fields_on_data_field_id ON public.fixture_list_fields USING btree (data_field_id);


--
-- Name: index_fixture_list_fields_on_fixture_list_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_fixture_list_fields_on_fixture_list_id ON public.fixture_list_fields USING btree (fixture_list_id);


--
-- Name: index_fixture_list_fields_on_fixture_list_id_and_data_field_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_fixture_list_fields_on_fixture_list_id_and_data_field_id ON public.fixture_list_fields USING btree (fixture_list_id, data_field_id);


--
-- Name: index_fixtures_on_away_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_fixtures_on_away_id ON public.fixtures USING btree (away_id);


--
-- Name: index_fixtures_on_away_id_and_starting_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_fixtures_on_away_id_and_starting_at ON public.fixtures USING btree (away_id, starting_at);


--
-- Name: index_fixtures_on_competition_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_fixtures_on_competition_id ON public.fixtures USING btree (competition_id);


--
-- Name: index_fixtures_on_competition_id_and_starting_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_fixtures_on_competition_id_and_starting_at ON public.fixtures USING btree (competition_id, starting_at);


--
-- Name: index_fixtures_on_external_ws_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_fixtures_on_external_ws_id ON public.fixtures USING btree (external_ws_id);


--
-- Name: index_fixtures_on_home_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_fixtures_on_home_id ON public.fixtures USING btree (home_id);


--
-- Name: index_fixtures_on_home_id_and_starting_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_fixtures_on_home_id_and_starting_at ON public.fixtures USING btree (home_id, starting_at);


--
-- Name: index_fixtures_on_season_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_fixtures_on_season_id ON public.fixtures USING btree (season_id);


--
-- Name: index_fixtures_on_season_id_and_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_fixtures_on_season_id_and_status ON public.fixtures USING btree (season_id, status);


--
-- Name: index_fixtures_on_starting_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_fixtures_on_starting_at ON public.fixtures USING btree (starting_at);


--
-- Name: index_fixtures_on_status_and_away_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_fixtures_on_status_and_away_id ON public.fixtures USING btree (status, away_id);


--
-- Name: index_fixtures_on_status_and_home_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_fixtures_on_status_and_home_id ON public.fixtures USING btree (status, home_id);


--
-- Name: index_season_teams_on_season_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_season_teams_on_season_id ON public.season_teams USING btree (season_id);


--
-- Name: index_season_teams_on_season_id_and_team_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_season_teams_on_season_id_and_team_id ON public.season_teams USING btree (season_id, team_id);


--
-- Name: index_season_teams_on_team_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_season_teams_on_team_id ON public.season_teams USING btree (team_id);


--
-- Name: index_seasons_on_competition_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_seasons_on_competition_id ON public.seasons USING btree (competition_id);


--
-- Name: index_seasons_on_external_ws_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_seasons_on_external_ws_id ON public.seasons USING btree (external_ws_id);


--
-- Name: index_seasons_on_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_seasons_on_index ON public.seasons USING btree (index);


--
-- Name: index_teams_on_external_ws_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_teams_on_external_ws_id ON public.teams USING btree (external_ws_id);


--
-- Name: fixtures fk_rails_140b10c8ba; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fixtures
    ADD CONSTRAINT fk_rails_140b10c8ba FOREIGN KEY (season_id) REFERENCES public.seasons(id);


--
-- Name: fixtures fk_rails_1f42e7a792; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fixtures
    ADD CONSTRAINT fk_rails_1f42e7a792 FOREIGN KEY (competition_id) REFERENCES public.competitions(id);


--
-- Name: season_teams fk_rails_2155f0dc86; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.season_teams
    ADD CONSTRAINT fk_rails_2155f0dc86 FOREIGN KEY (season_id) REFERENCES public.seasons(id);


--
-- Name: fixture_list_competitions fk_rails_285890b3a6; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fixture_list_competitions
    ADD CONSTRAINT fk_rails_285890b3a6 FOREIGN KEY (competition_id) REFERENCES public.competitions(id);


--
-- Name: fixtures fk_rails_306eb56476; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fixtures
    ADD CONSTRAINT fk_rails_306eb56476 FOREIGN KEY (home_id) REFERENCES public.teams(id);


--
-- Name: competitions fk_rails_3b0b6c07a2; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.competitions
    ADD CONSTRAINT fk_rails_3b0b6c07a2 FOREIGN KEY (country_id) REFERENCES public.countries(id);


--
-- Name: fixture_list_fields fk_rails_4c4ea6f1fc; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fixture_list_fields
    ADD CONSTRAINT fk_rails_4c4ea6f1fc FOREIGN KEY (fixture_list_id) REFERENCES public.fixture_lists(id);


--
-- Name: seasons fk_rails_5426847078; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seasons
    ADD CONSTRAINT fk_rails_5426847078 FOREIGN KEY (competition_id) REFERENCES public.competitions(id);


--
-- Name: season_teams fk_rails_91293a5993; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.season_teams
    ADD CONSTRAINT fk_rails_91293a5993 FOREIGN KEY (team_id) REFERENCES public.teams(id);


--
-- Name: fixture_list_fields fk_rails_9782f6ad4c; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fixture_list_fields
    ADD CONSTRAINT fk_rails_9782f6ad4c FOREIGN KEY (data_field_id) REFERENCES public.data_fields(id);


--
-- Name: fixtures fk_rails_a7807906ad; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fixtures
    ADD CONSTRAINT fk_rails_a7807906ad FOREIGN KEY (away_id) REFERENCES public.teams(id);


--
-- Name: fixture_list_competitions fk_rails_fcc1601ccf; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fixture_list_competitions
    ADD CONSTRAINT fk_rails_fcc1601ccf FOREIGN KEY (fixture_list_id) REFERENCES public.fixture_lists(id);


--
-- PostgreSQL database dump complete
--

\unrestrict EdWRZenb7mFdNu2UgQVaersivDeqiv9OblwKlhJaxPhij60hrIzkbhoMS6kduPo

SET search_path TO "$user", public;

INSERT INTO "schema_migrations" (version) VALUES
('20250811193804'),
('20250811161817'),
('20250714184422'),
('20250714184408'),
('20250714184356'),
('20250714184341'),
('20250714184331'),
('20250714184313'),
('20250714184306'),
('20250714184257'),
('20250714184246'),
('20250714184159'),
('20250713022556'),
('20250711194725');

