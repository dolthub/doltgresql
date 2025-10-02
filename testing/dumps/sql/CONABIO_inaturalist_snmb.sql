-- Downloaded from: https://github.com/CONABIO/inaturalist_snmb/blob/516ef220826c2ae150827dbf0da69142d32b0c18/db/structure.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 9.5.6
-- Dumped by pg_dump version 9.5.6

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- Name: postgis; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS postgis WITH SCHEMA public;


--
-- Name: EXTENSION postgis; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION postgis IS 'PostGIS geometry, geography, and raster spatial types and functions';


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


SET search_path = public, pg_catalog;

--
-- Name: _final_median(numeric[]); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION _final_median(numeric[]) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $_$
   SELECT AVG(val)
   FROM (
     SELECT val
     FROM unnest($1) val
     ORDER BY 1
     LIMIT  2 - MOD(array_upper($1, 1), 2)
     OFFSET CEIL(array_upper($1, 1) / 2.0) - 1
   ) sub;
$_$;


--
-- Name: _final_median(anyarray); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION _final_median(anyarray) RETURNS double precision
    LANGUAGE sql IMMUTABLE
    AS $_$ 
        WITH q AS
        (
           SELECT val
           FROM unnest($1) val
           WHERE VAL IS NOT NULL
           ORDER BY 1
        ),
        cnt AS
        (
          SELECT COUNT(*) AS c FROM q
        )
        SELECT AVG(val)::float8
        FROM 
        (
          SELECT val FROM q
          LIMIT  2 - MOD((SELECT c FROM cnt), 2)
          OFFSET GREATEST(CEIL((SELECT c FROM cnt) / 2.0) - 1,0)  
        ) q2;
      $_$;


--
-- Name: cleangeometry(geometry); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION cleangeometry(geom geometry) RETURNS geometry
    LANGUAGE plpgsql
    AS $_$
          DECLARE
          inGeom ALIAS for $1;
          outGeom geometry;
          tmpLinestring geometry;
          sqlString text;

      BEGIN

          outGeom := NULL;

          -- Clean Polygons --
          IF (ST_GeometryType(inGeom) = 'ST_Polygon' OR ST_GeometryType(inGeom) = 'ST_MultiPolygon') THEN

              -- Check if it needs fixing
              IF NOT ST_IsValid(inGeom) THEN

                  sqlString := '
                      -- separate multipolygon into 1 polygon per row
                      WITH split_multi (geom, poly) AS (
                          SELECT
                              (ST_Dump($1)).geom,
                              (ST_Dump($1)).path[1] -- polygon number
                      ),
                      -- break each polygon into linestrings
                      split_line (geom, poly, line) AS (
                          SELECT
                              ST_Boundary((ST_DumpRings(geom)).geom),
                              poly,
                              (ST_DumpRings(geom)).path[1] -- line number
                          FROM split_multi
                      ),
                      -- get the linestrings that make up the exterior of each polygon
                      line_exterior (geom, poly) AS (
                          SELECT
                              geom,
                              poly
                          FROM split_line
                          WHERE line = 0
                      ),
                      -- get an array of all the linestrings that make up the interior of each polygon
                      line_interior (geom, poly) AS (
                          SELECT
                              array_agg(geom ORDER BY line),
                              poly
                          FROM split_line
                          WHERE line > 0
                          GROUP BY poly
                      ),
                      -- use MakePolygon to rebuild the polygons
                      poly_geom (geom, poly) AS (
                          SELECT
                              CASE WHEN line_interior.geom IS NULL
                                  THEN ST_Buffer(ST_MakePolygon(line_exterior.geom), 0)
                                  ELSE ST_Buffer(ST_MakePolygon(line_exterior.geom, line_interior.geom), 0)
                              END,
                              line_exterior.poly
                          FROM line_exterior
                          LEFT JOIN line_interior USING (poly)
                      )
                  ';

                  IF (ST_GeometryType(inGeom) = 'ST_Polygon') THEN
                      sqlString := sqlString || '
                          SELECT geom
                          FROM poly_geom
                      ';
                  ELSE
                      sqlString := sqlString || '
                          , -- if its a multipolygon combine the polygons back together
                          multi_geom (geom) AS (
                              SELECT
                                  ST_Multi(ST_Collect(geom ORDER BY poly))
                              FROM poly_geom
                          )
                          SELECT geom
                          FROM multi_geom
                      ';
                  END IF;

                  EXECUTE sqlString INTO outGeom USING inGeom;

                  RETURN outGeom;
              ELSE
                  RETURN inGeom;
              END IF;

          -- Clean Lines --
          ELSIF (ST_GeometryType(inGeom) = 'ST_Linestring') THEN

              outGeom := ST_Union(ST_Multi(inGeom), ST_PointN(inGeom, 1));
              RETURN outGeom;
          ELSIF (ST_GeometryType(inGeom) = 'ST_MultiLinestring') THEN
              outGeom := ST_Multi(ST_Union(ST_Multi(inGeom), ST_PointN(inGeom, 1)));
              RETURN outGeom;
          ELSE
              RAISE NOTICE 'The input type % is not supported',ST_GeometryType(inGeom);
              RETURN inGeom;
          END IF;
      END;
      $_$;


--
-- Name: crc32(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION crc32(word text) RETURNS bigint
    LANGUAGE plpgsql IMMUTABLE
    AS $$
          DECLARE tmp bigint;
          DECLARE i int;
          DECLARE j int;
          DECLARE byte_length int;
          DECLARE word_array bytea;
          BEGIN
            IF COALESCE(word, '') = '' THEN
              return 0;
            END IF;

            i = 0;
            tmp = 4294967295;
            byte_length = bit_length(word) / 8;
            word_array = decode(replace(word, E'\\', E'\\\\'), 'escape');
            LOOP
              tmp = (tmp # get_byte(word_array, i))::bigint;
              i = i + 1;
              j = 0;
              LOOP
                tmp = ((tmp >> 1) # (3988292384 * (tmp & 1)))::bigint;
                j = j + 1;
                IF j >= 8 THEN
                  EXIT;
                END IF;
              END LOOP;
              IF i >= byte_length THEN
                EXIT;
              END IF;
            END LOOP;
            return (tmp # 4294967295);
          END
        $$;


--
-- Name: st_aslatlontext(geometry); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION st_aslatlontext(geometry) RETURNS text
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$ SELECT ST_AsLatLonText($1, '') $_$;


--
-- Name: median(anyelement); Type: AGGREGATE; Schema: public; Owner: -
--

CREATE AGGREGATE median(anyelement) (
    SFUNC = array_append,
    STYPE = anyarray,
    INITCOND = '{}',
    FINALFUNC = public._final_median
);


--
-- Name: median(numeric); Type: AGGREGATE; Schema: public; Owner: -
--

CREATE AGGREGATE median(numeric) (
    SFUNC = array_append,
    STYPE = numeric[],
    INITCOND = '{}',
    FINALFUNC = public._final_median
);


SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: annotations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE annotations (
    id integer NOT NULL,
    uuid uuid DEFAULT uuid_generate_v4(),
    resource_id integer,
    resource_type character varying,
    controlled_attribute_id integer,
    controlled_value_id integer,
    user_id integer,
    observation_field_value_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: annotations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE annotations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: annotations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE annotations_id_seq OWNED BY annotations.id;


--
-- Name: announcements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE announcements (
    id integer NOT NULL,
    placement character varying(255),
    start timestamp without time zone,
    "end" timestamp without time zone,
    body text,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    locales text[] DEFAULT '{}'::text[],
    site_id integer
);


--
-- Name: announcements_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE announcements_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: announcements_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE announcements_id_seq OWNED BY announcements.id;


--
-- Name: api_endpoint_caches; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE api_endpoint_caches (
    id integer NOT NULL,
    api_endpoint_id integer,
    request_url character varying,
    request_began_at timestamp without time zone,
    request_completed_at timestamp without time zone,
    success boolean,
    response text,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: api_endpoint_caches_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE api_endpoint_caches_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: api_endpoint_caches_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE api_endpoint_caches_id_seq OWNED BY api_endpoint_caches.id;


--
-- Name: api_endpoints; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE api_endpoints (
    id integer NOT NULL,
    title character varying NOT NULL,
    description text,
    documentation_url character varying,
    base_url character varying,
    cache_hours integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: api_endpoints_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE api_endpoints_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: api_endpoints_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE api_endpoints_id_seq OWNED BY api_endpoints.id;


--
-- Name: assessment_sections; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE assessment_sections (
    id integer NOT NULL,
    assessment_id integer,
    user_id integer,
    title character varying(255),
    body text,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    display_order integer
);


--
-- Name: assessment_sections_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE assessment_sections_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: assessment_sections_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE assessment_sections_id_seq OWNED BY assessment_sections.id;


--
-- Name: assessments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE assessments (
    id integer NOT NULL,
    taxon_id integer,
    project_id integer,
    user_id integer,
    description text,
    completed_at timestamp without time zone,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: assessments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE assessments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: assessments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE assessments_id_seq OWNED BY assessments.id;


--
-- Name: atlas_alterations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE atlas_alterations (
    id integer NOT NULL,
    atlas_id integer,
    user_id integer,
    place_id integer,
    action character varying,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: atlas_alterations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE atlas_alterations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: atlas_alterations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE atlas_alterations_id_seq OWNED BY atlas_alterations.id;


--
-- Name: atlases; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE atlases (
    id integer NOT NULL,
    user_id integer,
    taxon_id integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    is_active boolean DEFAULT false,
    is_marked boolean DEFAULT false
);


--
-- Name: atlases_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE atlases_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: atlases_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE atlases_id_seq OWNED BY atlases.id;


--
-- Name: colors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE colors (
    id integer NOT NULL,
    value character varying(255)
);


--
-- Name: colors_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE colors_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: colors_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE colors_id_seq OWNED BY colors.id;


--
-- Name: colors_taxa; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE colors_taxa (
    color_id integer,
    taxon_id integer
);


--
-- Name: comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE comments (
    id integer NOT NULL,
    user_id integer,
    parent_id integer,
    parent_type character varying(255),
    body text,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    uuid uuid DEFAULT uuid_generate_v4()
);


--
-- Name: comments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE comments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: comments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE comments_id_seq OWNED BY comments.id;


--
-- Name: complete_sets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE complete_sets (
    id integer NOT NULL,
    user_id integer,
    taxon_id integer,
    place_id integer,
    description text,
    source_id integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    is_active boolean DEFAULT false
);


--
-- Name: complete_sets_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE complete_sets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: complete_sets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE complete_sets_id_seq OWNED BY complete_sets.id;


--
-- Name: conservation_statuses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE conservation_statuses (
    id integer NOT NULL,
    taxon_id integer,
    user_id integer,
    place_id integer,
    source_id integer,
    authority character varying(255),
    status character varying(255),
    url character varying(512),
    description text,
    geoprivacy character varying(255) DEFAULT 'obscured'::character varying,
    iucn integer DEFAULT 20,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: conservation_statuses_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE conservation_statuses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: conservation_statuses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE conservation_statuses_id_seq OWNED BY conservation_statuses.id;


--
-- Name: controlled_term_labels; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE controlled_term_labels (
    id integer NOT NULL,
    controlled_term_id integer,
    locale character varying,
    valid_within_clade integer,
    label character varying,
    definition character varying,
    icon_file_name character varying,
    icon_content_type character varying,
    icon_file_size character varying,
    icon_updated_at character varying,
    user_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: controlled_term_labels_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE controlled_term_labels_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: controlled_term_labels_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE controlled_term_labels_id_seq OWNED BY controlled_term_labels.id;


--
-- Name: controlled_term_values; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE controlled_term_values (
    id integer NOT NULL,
    controlled_attribute_id integer,
    controlled_value_id integer
);


--
-- Name: controlled_term_values_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE controlled_term_values_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: controlled_term_values_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE controlled_term_values_id_seq OWNED BY controlled_term_values.id;


--
-- Name: controlled_terms; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE controlled_terms (
    id integer NOT NULL,
    ontology_uri text,
    uri text,
    valid_within_clade integer,
    is_value boolean DEFAULT false,
    active boolean DEFAULT false,
    multivalued boolean DEFAULT false,
    user_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: controlled_terms_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE controlled_terms_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: controlled_terms_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE controlled_terms_id_seq OWNED BY controlled_terms.id;


--
-- Name: counties_simplified_01; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE counties_simplified_01 (
    id integer NOT NULL,
    place_geometry_id integer,
    place_id integer,
    geom geometry(MultiPolygon) NOT NULL
);


--
-- Name: counties_simplified_01_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE counties_simplified_01_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: counties_simplified_01_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE counties_simplified_01_id_seq OWNED BY counties_simplified_01.id;


--
-- Name: countries_simplified_1; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE countries_simplified_1 (
    id integer NOT NULL,
    place_geometry_id integer,
    place_id integer,
    geom geometry(MultiPolygon) NOT NULL
);


--
-- Name: countries_simplified_1_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE countries_simplified_1_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: countries_simplified_1_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE countries_simplified_1_id_seq OWNED BY countries_simplified_1.id;


--
-- Name: custom_projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE custom_projects (
    id integer NOT NULL,
    head text,
    side text,
    css text,
    project_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: custom_projects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE custom_projects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: custom_projects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE custom_projects_id_seq OWNED BY custom_projects.id;


--
-- Name: delayed_jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE delayed_jobs (
    id integer NOT NULL,
    priority integer DEFAULT 0,
    attempts integer DEFAULT 0,
    handler text,
    last_error text,
    run_at timestamp without time zone,
    locked_at timestamp without time zone,
    failed_at timestamp without time zone,
    locked_by character varying(255),
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    queue character varying(255),
    unique_hash character varying
);


--
-- Name: delayed_jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE delayed_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: delayed_jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE delayed_jobs_id_seq OWNED BY delayed_jobs.id;


--
-- Name: deleted_observations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE deleted_observations (
    id integer NOT NULL,
    user_id integer,
    observation_id integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: deleted_observations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE deleted_observations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: deleted_observations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE deleted_observations_id_seq OWNED BY deleted_observations.id;


--
-- Name: deleted_photos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE deleted_photos (
    id integer NOT NULL,
    user_id integer,
    photo_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    removed_from_s3 boolean DEFAULT false NOT NULL,
    orphan boolean DEFAULT false NOT NULL
);


--
-- Name: deleted_photos_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE deleted_photos_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: deleted_photos_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE deleted_photos_id_seq OWNED BY deleted_photos.id;


--
-- Name: deleted_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE deleted_users (
    id integer NOT NULL,
    user_id integer,
    login character varying(255),
    email character varying(255),
    user_created_at timestamp without time zone,
    user_updated_at timestamp without time zone,
    observations_count integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: deleted_users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE deleted_users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: deleted_users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE deleted_users_id_seq OWNED BY deleted_users.id;


--
-- Name: exploded_atlas_places; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE exploded_atlas_places (
    id integer NOT NULL,
    atlas_id integer,
    place_id integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: exploded_atlas_places_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE exploded_atlas_places_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exploded_atlas_places_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE exploded_atlas_places_id_seq OWNED BY exploded_atlas_places.id;


--
-- Name: flags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE flags (
    id integer NOT NULL,
    flag character varying(255),
    comment character varying(255),
    created_at timestamp without time zone NOT NULL,
    flaggable_id integer DEFAULT 0 NOT NULL,
    flaggable_type character varying(15) NOT NULL,
    user_id integer DEFAULT 0 NOT NULL,
    resolver_id integer,
    resolved boolean DEFAULT false,
    updated_at timestamp without time zone
);


--
-- Name: flags_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE flags_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: flags_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE flags_id_seq OWNED BY flags.id;


--
-- Name: flickr_identities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE flickr_identities (
    id integer NOT NULL,
    flickr_username character varying(255),
    frob character varying(255),
    token character varying(255),
    token_created_at timestamp without time zone,
    auto_import integer DEFAULT 0,
    auto_imported_at timestamp without time zone,
    user_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    flickr_user_id character varying(255),
    secret character varying(255)
);


--
-- Name: flickr_identities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE flickr_identities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: flickr_identities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE flickr_identities_id_seq OWNED BY flickr_identities.id;


--
-- Name: flow_task_resources; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE flow_task_resources (
    id integer NOT NULL,
    flow_task_id integer,
    resource_type character varying(255),
    resource_id integer,
    type character varying(255),
    file_file_name character varying(255),
    file_content_type character varying(255),
    file_file_size integer,
    file_updated_at timestamp without time zone,
    extra text,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: flow_task_resources_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE flow_task_resources_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: flow_task_resources_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE flow_task_resources_id_seq OWNED BY flow_task_resources.id;


--
-- Name: flow_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE flow_tasks (
    id integer NOT NULL,
    type character varying(255),
    options text,
    command character varying(255),
    error character varying(255),
    started_at timestamp without time zone,
    finished_at timestamp without time zone,
    user_id integer,
    redirect_url character varying(255),
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    exception text
);


--
-- Name: flow_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE flow_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: flow_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE flow_tasks_id_seq OWNED BY flow_tasks.id;


--
-- Name: friendly_id_slugs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE friendly_id_slugs (
    id integer NOT NULL,
    slug character varying(255),
    sluggable_id integer,
    sequence integer DEFAULT 1 NOT NULL,
    sluggable_type character varying(40),
    scope character varying(255),
    created_at timestamp without time zone
);


--
-- Name: friendly_id_slugs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE friendly_id_slugs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: friendly_id_slugs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE friendly_id_slugs_id_seq OWNED BY friendly_id_slugs.id;


--
-- Name: friendships; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE friendships (
    id integer NOT NULL,
    user_id integer,
    friend_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: friendships_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE friendships_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: friendships_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE friendships_id_seq OWNED BY friendships.id;


--
-- Name: goal_contributions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE goal_contributions (
    id integer NOT NULL,
    contribution_id integer,
    contribution_type character varying(255),
    goal_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    goal_participant_id integer
);


--
-- Name: goal_contributions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE goal_contributions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: goal_contributions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE goal_contributions_id_seq OWNED BY goal_contributions.id;


--
-- Name: goal_participants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE goal_participants (
    id integer NOT NULL,
    goal_id integer,
    user_id integer,
    goal_completed integer DEFAULT 0,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: goal_participants_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE goal_participants_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: goal_participants_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE goal_participants_id_seq OWNED BY goal_participants.id;


--
-- Name: goal_rules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE goal_rules (
    id integer NOT NULL,
    goal_id integer,
    operator character varying(255),
    operator_class character varying(255),
    arguments character varying(255),
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: goal_rules_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE goal_rules_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: goal_rules_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE goal_rules_id_seq OWNED BY goal_rules.id;


--
-- Name: goals; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE goals (
    id integer NOT NULL,
    description text,
    number_of_contributions_required integer,
    goal_type character varying(255),
    ends_at timestamp without time zone,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    completed boolean DEFAULT false
);


--
-- Name: goals_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE goals_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: goals_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE goals_id_seq OWNED BY goals.id;


--
-- Name: guide_photos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE guide_photos (
    id integer NOT NULL,
    guide_taxon_id integer,
    title character varying(255),
    description character varying(255),
    photo_id integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    "position" integer DEFAULT 0
);


--
-- Name: guide_photos_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE guide_photos_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: guide_photos_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE guide_photos_id_seq OWNED BY guide_photos.id;


--
-- Name: guide_ranges; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE guide_ranges (
    id integer NOT NULL,
    guide_taxon_id integer,
    medium_url character varying(512),
    thumb_url character varying(512),
    original_url character varying(512),
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    license character varying(255),
    source_url character varying(512),
    rights_holder character varying(255),
    source_id integer,
    file_file_name character varying(255),
    file_content_type character varying(255),
    file_file_size integer,
    file_updated_at timestamp without time zone,
    "position" integer DEFAULT 0
);


--
-- Name: guide_ranges_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE guide_ranges_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: guide_ranges_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE guide_ranges_id_seq OWNED BY guide_ranges.id;


--
-- Name: guide_sections; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE guide_sections (
    id integer NOT NULL,
    guide_taxon_id integer,
    title character varying(255),
    description text,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    "position" integer DEFAULT 0,
    license character varying(255),
    source_url character varying(255),
    rights_holder character varying(255),
    source_id integer,
    creator_id integer,
    updater_id integer
);


--
-- Name: guide_sections_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE guide_sections_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: guide_sections_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE guide_sections_id_seq OWNED BY guide_sections.id;


--
-- Name: guide_taxa; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE guide_taxa (
    id integer NOT NULL,
    guide_id integer,
    taxon_id integer,
    name character varying(255),
    display_name character varying(255),
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    "position" integer DEFAULT 0,
    source_identifier character varying(255)
);


--
-- Name: guide_taxa_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE guide_taxa_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: guide_taxa_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE guide_taxa_id_seq OWNED BY guide_taxa.id;


--
-- Name: guide_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE guide_users (
    id integer NOT NULL,
    guide_id integer,
    user_id integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: guide_users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE guide_users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: guide_users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE guide_users_id_seq OWNED BY guide_users.id;


--
-- Name: guides; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE guides (
    id integer NOT NULL,
    title character varying(255),
    description text,
    published_at timestamp without time zone,
    latitude numeric,
    longitude numeric,
    user_id integer,
    place_id integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    license character varying(255) DEFAULT 'CC-BY-SA'::character varying,
    icon_file_name character varying(255),
    icon_content_type character varying(255),
    icon_file_size integer,
    icon_updated_at timestamp without time zone,
    map_type character varying(255) DEFAULT 'terrain'::character varying,
    zoom_level integer,
    taxon_id integer,
    source_url character varying(255),
    downloadable boolean DEFAULT false,
    ngz_file_name character varying(255),
    ngz_content_type character varying(255),
    ngz_file_size integer,
    ngz_updated_at timestamp without time zone
);


--
-- Name: guides_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE guides_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: guides_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE guides_id_seq OWNED BY guides.id;


--
-- Name: identifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE identifications (
    id integer NOT NULL,
    observation_id integer,
    taxon_id integer,
    user_id integer,
    type character varying(255),
    body text,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    current boolean DEFAULT true,
    taxon_change_id integer,
    category character varying,
    uuid uuid DEFAULT uuid_generate_v4(),
    blind boolean,
    previous_observation_taxon_id integer
);


--
-- Name: identifications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE identifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: identifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE identifications_id_seq OWNED BY identifications.id;


--
-- Name: invites; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE invites (
    id integer NOT NULL,
    user_id integer,
    invite_address character varying(255),
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: invites_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE invites_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: invites_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE invites_id_seq OWNED BY invites.id;


--
-- Name: list_rules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE list_rules (
    id integer NOT NULL,
    list_id integer,
    operator character varying(255),
    operand_id integer,
    operand_type character varying(255),
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: list_rules_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE list_rules_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: list_rules_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE list_rules_id_seq OWNED BY list_rules.id;


--
-- Name: listed_taxa; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE listed_taxa (
    id integer NOT NULL,
    taxon_id integer,
    list_id integer,
    last_observation_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    taxon_ancestor_ids character varying(255),
    place_id integer,
    description text,
    comments_count integer DEFAULT 0,
    user_id integer,
    updater_id integer,
    occurrence_status_level integer,
    establishment_means character varying(32),
    first_observation_id integer,
    observations_count integer DEFAULT 0,
    observations_month_counts character varying(255),
    taxon_range_id integer,
    source_id integer,
    manually_added boolean DEFAULT false,
    primary_listing boolean DEFAULT true
);


--
-- Name: listed_taxa_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE listed_taxa_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: listed_taxa_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE listed_taxa_id_seq OWNED BY listed_taxa.id;


--
-- Name: listed_taxon_alterations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE listed_taxon_alterations (
    id integer NOT NULL,
    taxon_id integer,
    user_id integer,
    place_id integer,
    action character varying,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: listed_taxon_alterations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE listed_taxon_alterations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: listed_taxon_alterations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE listed_taxon_alterations_id_seq OWNED BY listed_taxon_alterations.id;


--
-- Name: lists; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE lists (
    id integer NOT NULL,
    title character varying(255),
    description text,
    user_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    type character varying(255),
    comprehensive boolean DEFAULT false,
    taxon_id integer,
    last_synced_at timestamp without time zone,
    place_id integer,
    project_id integer,
    source_id integer,
    show_obs_photos boolean DEFAULT true
);


--
-- Name: lists_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE lists_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: lists_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE lists_id_seq OWNED BY lists.id;


--
-- Name: messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE messages (
    id integer NOT NULL,
    user_id integer,
    from_user_id integer,
    to_user_id integer,
    thread_id integer,
    subject character varying(255),
    body text,
    read_at timestamp without time zone,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: messages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE messages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: messages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE messages_id_seq OWNED BY messages.id;


--
-- Name: model_attribute_changes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE model_attribute_changes (
    id integer NOT NULL,
    model_type character varying,
    model_id integer,
    field_name character varying,
    changed_at timestamp without time zone
);


--
-- Name: model_attribute_changes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE model_attribute_changes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: model_attribute_changes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE model_attribute_changes_id_seq OWNED BY model_attribute_changes.id;


--
-- Name: oauth_access_grants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE oauth_access_grants (
    id integer NOT NULL,
    resource_owner_id integer NOT NULL,
    application_id integer NOT NULL,
    token character varying(255) NOT NULL,
    expires_in integer NOT NULL,
    redirect_uri character varying(255) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    revoked_at timestamp without time zone,
    scopes character varying(255)
);


--
-- Name: oauth_access_grants_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE oauth_access_grants_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: oauth_access_grants_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE oauth_access_grants_id_seq OWNED BY oauth_access_grants.id;


--
-- Name: oauth_access_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE oauth_access_tokens (
    id integer NOT NULL,
    resource_owner_id integer,
    application_id integer NOT NULL,
    token character varying(255) NOT NULL,
    refresh_token character varying(255),
    expires_in integer,
    revoked_at timestamp without time zone,
    created_at timestamp without time zone NOT NULL,
    scopes character varying(255)
);


--
-- Name: oauth_access_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE oauth_access_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: oauth_access_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE oauth_access_tokens_id_seq OWNED BY oauth_access_tokens.id;


--
-- Name: oauth_applications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE oauth_applications (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    uid character varying(255) NOT NULL,
    secret character varying(255) NOT NULL,
    redirect_uri character varying(255) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    owner_id integer,
    owner_type character varying(255),
    trusted boolean DEFAULT false,
    image_file_name character varying(255),
    image_content_type character varying(255),
    image_file_size integer,
    image_updated_at timestamp without time zone,
    url character varying(255),
    description text,
    scopes character varying DEFAULT ''::character varying NOT NULL
);


--
-- Name: oauth_applications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE oauth_applications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: oauth_applications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE oauth_applications_id_seq OWNED BY oauth_applications.id;


--
-- Name: observation_field_values; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_field_values (
    id integer NOT NULL,
    observation_id integer,
    observation_field_id integer,
    value character varying(2048),
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    user_id integer,
    updater_id integer,
    uuid uuid DEFAULT uuid_generate_v4()
);


--
-- Name: observation_field_values_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE observation_field_values_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: observation_field_values_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE observation_field_values_id_seq OWNED BY observation_field_values.id;


--
-- Name: observation_fields; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_fields (
    id integer NOT NULL,
    name character varying(255),
    datatype character varying(255),
    user_id integer,
    description character varying(255),
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    allowed_values text,
    values_count integer,
    users_count integer
);


--
-- Name: observation_fields_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE observation_fields_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: observation_fields_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE observation_fields_id_seq OWNED BY observation_fields.id;


--
-- Name: observation_links; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_links (
    id integer NOT NULL,
    observation_id integer,
    rel character varying(255),
    href character varying(255),
    href_name character varying(255),
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: observation_links_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE observation_links_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: observation_links_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE observation_links_id_seq OWNED BY observation_links.id;


--
-- Name: observation_photos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_photos (
    id integer NOT NULL,
    observation_id integer NOT NULL,
    photo_id integer NOT NULL,
    "position" integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    uuid character varying(255)
);


--
-- Name: observation_photos_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE observation_photos_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: observation_photos_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE observation_photos_id_seq OWNED BY observation_photos.id;


--
-- Name: observation_reviews; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_reviews (
    id integer NOT NULL,
    user_id integer,
    observation_id integer,
    reviewed boolean DEFAULT true,
    user_added boolean DEFAULT false,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: observation_reviews_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE observation_reviews_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: observation_reviews_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE observation_reviews_id_seq OWNED BY observation_reviews.id;


--
-- Name: observation_sounds; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_sounds (
    id integer NOT NULL,
    observation_id integer,
    sound_id integer
);


--
-- Name: observation_sounds_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE observation_sounds_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: observation_sounds_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE observation_sounds_id_seq OWNED BY observation_sounds.id;


--
-- Name: observation_zooms_10; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_10 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observation_zooms_11; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_11 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observation_zooms_12; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_12 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observation_zooms_125; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_125 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observation_zooms_2; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_2 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observation_zooms_2000; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_2000 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observation_zooms_250; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_250 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observation_zooms_3; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_3 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observation_zooms_4; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_4 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observation_zooms_4000; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_4000 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observation_zooms_5; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_5 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observation_zooms_500; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_500 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observation_zooms_6; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_6 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observation_zooms_63; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_63 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observation_zooms_7; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_7 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observation_zooms_8; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_8 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observation_zooms_9; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_9 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observation_zooms_990; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observation_zooms_990 (
    taxon_id integer,
    geom geometry,
    count integer NOT NULL
);


--
-- Name: observations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observations (
    id integer NOT NULL,
    observed_on date,
    description text,
    latitude numeric(15,10),
    longitude numeric(15,10),
    map_scale integer,
    timeframe text,
    species_guess character varying(255),
    user_id integer,
    taxon_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    place_guess character varying(255),
    id_please boolean DEFAULT false,
    observed_on_string character varying(255),
    iconic_taxon_id integer,
    num_identification_agreements integer DEFAULT 0,
    num_identification_disagreements integer DEFAULT 0,
    time_observed_at timestamp without time zone,
    time_zone character varying(255),
    location_is_exact boolean DEFAULT false,
    delta boolean DEFAULT false,
    positional_accuracy integer,
    private_latitude numeric(15,10),
    private_longitude numeric(15,10),
    private_positional_accuracy integer,
    geoprivacy character varying(255),
    quality_grade character varying DEFAULT 'casual'::character varying,
    user_agent character varying(255),
    positioning_method character varying(255),
    positioning_device character varying(255),
    out_of_range boolean,
    license character varying(255),
    uri character varying(255),
    observation_photos_count integer DEFAULT 0,
    comments_count integer DEFAULT 0,
    geom geometry(Point),
    cached_tag_list character varying(768) DEFAULT NULL::character varying,
    zic_time_zone character varying(255),
    oauth_application_id integer,
    observation_sounds_count integer DEFAULT 0,
    identifications_count integer DEFAULT 0,
    private_geom geometry(Point),
    community_taxon_id integer,
    captive boolean DEFAULT false,
    site_id integer,
    uuid character varying(255),
    public_positional_accuracy integer,
    mappable boolean DEFAULT false,
    cached_votes_total integer DEFAULT 0,
    last_indexed_at timestamp without time zone,
    private_place_guess character varying
);


--
-- Name: observations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE observations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: observations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE observations_id_seq OWNED BY observations.id;


--
-- Name: observations_places; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observations_places (
    id integer NOT NULL,
    observation_id integer NOT NULL,
    place_id integer NOT NULL
);


--
-- Name: observations_places_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE observations_places_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: observations_places_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE observations_places_id_seq OWNED BY observations_places.id;


--
-- Name: observations_posts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE observations_posts (
    observation_id integer NOT NULL,
    post_id integer NOT NULL
);


--
-- Name: passwords; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE passwords (
    id integer NOT NULL,
    user_id integer,
    reset_code character varying(255),
    expiration_date timestamp without time zone,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: passwords_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE passwords_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: passwords_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE passwords_id_seq OWNED BY passwords.id;


--
-- Name: photos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE photos (
    id integer NOT NULL,
    user_id integer,
    native_photo_id character varying(255),
    square_url character varying(512),
    thumb_url character varying(512),
    small_url character varying(512),
    medium_url character varying(512),
    large_url character varying(512),
    original_url character varying(512),
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    native_page_url character varying(512),
    native_username character varying(255),
    native_realname character varying(255),
    license integer,
    type character varying(255),
    file_content_type character varying(255),
    file_file_name character varying(255),
    file_file_size integer,
    file_processing boolean,
    mobile boolean DEFAULT false,
    file_updated_at timestamp without time zone,
    metadata text,
    subtype character varying(255),
    native_original_image_url character varying(512)
);


--
-- Name: photos_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE photos_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: photos_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE photos_id_seq OWNED BY photos.id;


--
-- Name: picasa_identities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE picasa_identities (
    id integer NOT NULL,
    user_id integer,
    token character varying(255),
    token_created_at timestamp without time zone,
    picasa_user_id character varying(255),
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: picasa_identities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE picasa_identities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: picasa_identities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE picasa_identities_id_seq OWNED BY picasa_identities.id;


--
-- Name: place_geometries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE place_geometries (
    id integer NOT NULL,
    place_id integer,
    source_name character varying(255),
    source_identifier character varying(255),
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    source_filename character varying(255),
    geom geometry(MultiPolygon) NOT NULL,
    source_id integer
);


--
-- Name: place_geometries_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE place_geometries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: place_geometries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE place_geometries_id_seq OWNED BY place_geometries.id;


--
-- Name: place_taxon_names; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE place_taxon_names (
    id integer NOT NULL,
    place_id integer,
    taxon_name_id integer,
    "position" integer DEFAULT 0
);


--
-- Name: place_taxon_names_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE place_taxon_names_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: place_taxon_names_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE place_taxon_names_id_seq OWNED BY place_taxon_names.id;


--
-- Name: places; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE places (
    id integer NOT NULL,
    name character varying(255),
    display_name character varying(255),
    code character varying(255),
    latitude numeric(15,10),
    longitude numeric(15,10),
    swlat numeric(15,10),
    swlng numeric(15,10),
    nelat numeric(15,10),
    nelng numeric(15,10),
    woeid integer,
    parent_id integer,
    check_list_id integer,
    place_type integer,
    bbox_area double precision,
    source_name character varying(255),
    source_identifier character varying(255),
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    delta boolean DEFAULT false,
    user_id integer,
    source_filename character varying(255),
    ancestry character varying(255),
    slug character varying(255),
    source_id integer,
    admin_level integer
);


--
-- Name: places_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE places_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: places_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE places_id_seq OWNED BY places.id;


--
-- Name: posts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE posts (
    id integer NOT NULL,
    parent_id integer NOT NULL,
    parent_type character varying(255) NOT NULL,
    user_id integer NOT NULL,
    published_at timestamp without time zone,
    title character varying(255) NOT NULL,
    body text,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    type character varying(255),
    start_time timestamp without time zone,
    stop_time timestamp without time zone,
    place_id integer,
    latitude numeric(15,10),
    longitude numeric(15,10),
    radius integer
);


--
-- Name: posts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE posts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: posts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE posts_id_seq OWNED BY posts.id;


--
-- Name: preferences; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE preferences (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    owner_id integer NOT NULL,
    owner_type character varying(255) NOT NULL,
    group_id integer,
    group_type character varying(255),
    value text,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: preferences_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE preferences_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: preferences_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE preferences_id_seq OWNED BY preferences.id;


--
-- Name: project_assets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE project_assets (
    id integer NOT NULL,
    project_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    asset_file_name character varying(255),
    asset_content_type character varying(255),
    asset_file_size integer,
    asset_updated_at timestamp without time zone
);


--
-- Name: project_assets_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE project_assets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: project_assets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE project_assets_id_seq OWNED BY project_assets.id;


--
-- Name: project_invitations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE project_invitations (
    id integer NOT NULL,
    project_id integer,
    user_id integer,
    observation_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: project_invitations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE project_invitations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: project_invitations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE project_invitations_id_seq OWNED BY project_invitations.id;


--
-- Name: project_observation_fields; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE project_observation_fields (
    id integer NOT NULL,
    project_id integer,
    observation_field_id integer,
    required boolean,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    "position" integer
);


--
-- Name: project_observation_fields_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE project_observation_fields_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: project_observation_fields_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE project_observation_fields_id_seq OWNED BY project_observation_fields.id;


--
-- Name: project_observations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE project_observations (
    id integer NOT NULL,
    project_id integer,
    observation_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    curator_identification_id integer,
    tracking_code character varying(255),
    user_id integer,
    uuid uuid DEFAULT uuid_generate_v4()
);


--
-- Name: project_observations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE project_observations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: project_observations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE project_observations_id_seq OWNED BY project_observations.id;


--
-- Name: project_user_invitations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE project_user_invitations (
    id integer NOT NULL,
    user_id integer,
    invited_user_id integer,
    project_id integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: project_user_invitations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE project_user_invitations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: project_user_invitations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE project_user_invitations_id_seq OWNED BY project_user_invitations.id;


--
-- Name: project_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE project_users (
    id integer NOT NULL,
    project_id integer,
    user_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    role character varying(255),
    observations_count integer DEFAULT 0,
    taxa_count integer DEFAULT 0
);


--
-- Name: project_users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE project_users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: project_users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE project_users_id_seq OWNED BY project_users.id;


--
-- Name: projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE projects (
    id integer NOT NULL,
    user_id integer,
    title character varying(255),
    description text,
    terms text,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    icon_file_name character varying(255),
    icon_content_type character varying(255),
    icon_file_size integer,
    icon_updated_at timestamp without time zone,
    project_type character varying(255),
    slug character varying(255),
    observed_taxa_count integer DEFAULT 0,
    featured_at timestamp without time zone,
    source_url character varying(255),
    tracking_codes character varying(255),
    delta boolean DEFAULT false,
    place_id integer,
    map_type character varying(255) DEFAULT 'terrain'::character varying,
    latitude numeric(15,10),
    longitude numeric(15,10),
    zoom_level integer,
    cover_file_name character varying(255),
    cover_content_type character varying(255),
    cover_file_size integer,
    cover_updated_at timestamp without time zone,
    event_url character varying(255),
    start_time timestamp without time zone,
    end_time timestamp without time zone,
    trusted boolean DEFAULT false,
    "group" character varying(255),
    last_aggregated_at timestamp without time zone
);


--
-- Name: projects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE projects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: projects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE projects_id_seq OWNED BY projects.id;


--
-- Name: provider_authorizations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE provider_authorizations (
    id integer NOT NULL,
    provider_name character varying(255) NOT NULL,
    provider_uid text,
    token text,
    user_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    scope character varying(255),
    secret character varying(255),
    refresh_token character varying
);


--
-- Name: provider_authorizations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE provider_authorizations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: provider_authorizations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE provider_authorizations_id_seq OWNED BY provider_authorizations.id;


--
-- Name: quality_metrics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE quality_metrics (
    id integer NOT NULL,
    user_id integer,
    observation_id integer,
    metric character varying(255),
    agree boolean DEFAULT true,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: quality_metrics_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE quality_metrics_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: quality_metrics_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE quality_metrics_id_seq OWNED BY quality_metrics.id;


--
-- Name: roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE roles (
    id integer NOT NULL,
    name character varying(255)
);


--
-- Name: roles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE roles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE roles_id_seq OWNED BY roles.id;


--
-- Name: roles_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE roles_users (
    role_id integer,
    user_id integer
);


--
-- Name: rules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE rules (
    id integer NOT NULL,
    type character varying(255),
    ruler_type character varying(255),
    ruler_id integer,
    operand_type character varying(255),
    operand_id integer,
    operator character varying(255),
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: rules_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE rules_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: rules_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE rules_id_seq OWNED BY rules.id;


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE schema_migrations (
    version character varying(255) NOT NULL
);


--
-- Name: site_admins; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE site_admins (
    id integer NOT NULL,
    user_id integer,
    site_id integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: site_admins_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE site_admins_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: site_admins_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE site_admins_id_seq OWNED BY site_admins.id;


--
-- Name: site_statistics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE site_statistics (
    id integer NOT NULL,
    created_at timestamp without time zone,
    data json
);


--
-- Name: site_statistics_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE site_statistics_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: site_statistics_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE site_statistics_id_seq OWNED BY site_statistics.id;


--
-- Name: sites; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE sites (
    id integer NOT NULL,
    name character varying(255),
    url character varying(255),
    place_id integer,
    source_id integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    logo_file_name character varying(255),
    logo_content_type character varying(255),
    logo_file_size integer,
    logo_updated_at timestamp without time zone,
    logo_square_file_name character varying(255),
    logo_square_content_type character varying(255),
    logo_square_file_size integer,
    logo_square_updated_at timestamp without time zone,
    stylesheet_file_name character varying(255),
    stylesheet_content_type character varying(255),
    stylesheet_file_size integer,
    stylesheet_updated_at timestamp without time zone,
    draft boolean DEFAULT false,
    homepage_data text,
    logo_email_banner_file_name character varying,
    logo_email_banner_content_type character varying,
    logo_email_banner_file_size integer,
    logo_email_banner_updated_at timestamp without time zone
);


--
-- Name: sites_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE sites_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sites_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE sites_id_seq OWNED BY sites.id;


--
-- Name: soundcloud_identities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE soundcloud_identities (
    id integer NOT NULL,
    native_username character varying(255),
    native_realname character varying(255),
    user_id integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: soundcloud_identities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE soundcloud_identities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: soundcloud_identities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE soundcloud_identities_id_seq OWNED BY soundcloud_identities.id;


--
-- Name: sounds; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE sounds (
    id integer NOT NULL,
    user_id integer,
    native_username character varying(255),
    native_realname character varying(255),
    native_sound_id character varying(255),
    native_page_url character varying(255),
    license integer,
    type character varying(255),
    sound_url character varying(255),
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    native_response text
);


--
-- Name: sounds_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE sounds_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sounds_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE sounds_id_seq OWNED BY sounds.id;


--
-- Name: sources; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE sources (
    id integer NOT NULL,
    in_text character varying(255),
    citation character varying(512),
    url character varying(512),
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    title character varying(255),
    user_id integer
);


--
-- Name: sources_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE sources_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sources_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE sources_id_seq OWNED BY sources.id;


--
-- Name: states_simplified_1; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE states_simplified_1 (
    id integer NOT NULL,
    place_geometry_id integer,
    place_id integer,
    geom geometry(MultiPolygon) NOT NULL
);


--
-- Name: states_simplified_1_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE states_simplified_1_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: states_simplified_1_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE states_simplified_1_id_seq OWNED BY states_simplified_1.id;


--
-- Name: subscriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE subscriptions (
    id integer NOT NULL,
    user_id integer,
    resource_type character varying(255),
    resource_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    taxon_id integer
);


--
-- Name: subscriptions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE subscriptions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: subscriptions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE subscriptions_id_seq OWNED BY subscriptions.id;


--
-- Name: taggings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE taggings (
    id integer NOT NULL,
    tag_id integer,
    taggable_id integer,
    taggable_type character varying(255),
    created_at timestamp without time zone,
    tagger_id integer,
    tagger_type character varying,
    context character varying(128)
);


--
-- Name: taggings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE taggings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: taggings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE taggings_id_seq OWNED BY taggings.id;


--
-- Name: tags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE tags (
    id integer NOT NULL,
    name character varying(255),
    taggings_count integer DEFAULT 0
);


--
-- Name: tags_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE tags_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tags_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE tags_id_seq OWNED BY tags.id;


--
-- Name: taxa; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE taxa (
    id integer NOT NULL,
    name character varying(255),
    rank character varying(255),
    source_identifier character varying(255),
    source_url character varying(255),
    parent_id integer,
    source_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    iconic_taxon_id integer,
    is_iconic boolean DEFAULT false,
    auto_photos boolean DEFAULT true,
    auto_description boolean DEFAULT true,
    version integer,
    name_provider character varying(255),
    delta boolean DEFAULT false,
    creator_id integer,
    updater_id integer,
    observations_count integer DEFAULT 0,
    listed_taxa_count integer DEFAULT 0,
    rank_level integer,
    unique_name character varying(255),
    wikipedia_summary text,
    wikipedia_title character varying(255),
    featured_at timestamp without time zone,
    ancestry character varying(255),
    conservation_status integer,
    conservation_status_source_id integer,
    locked boolean DEFAULT false NOT NULL,
    conservation_status_source_identifier integer,
    is_active boolean DEFAULT true NOT NULL
);


--
-- Name: taxa_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE taxa_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: taxa_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE taxa_id_seq OWNED BY taxa.id;


--
-- Name: taxon_ancestors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE taxon_ancestors (
    taxon_id integer NOT NULL,
    ancestor_taxon_id integer NOT NULL
);


--
-- Name: taxon_change_taxa; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE taxon_change_taxa (
    id integer NOT NULL,
    taxon_change_id integer,
    taxon_id integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: taxon_change_taxa_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE taxon_change_taxa_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: taxon_change_taxa_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE taxon_change_taxa_id_seq OWNED BY taxon_change_taxa.id;


--
-- Name: taxon_changes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE taxon_changes (
    id integer NOT NULL,
    description text,
    taxon_id integer,
    source_id integer,
    user_id integer,
    type character varying(255),
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    committed_on date,
    change_group character varying(255),
    committer_id integer
);


--
-- Name: taxon_changes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE taxon_changes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: taxon_changes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE taxon_changes_id_seq OWNED BY taxon_changes.id;


--
-- Name: taxon_descriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE taxon_descriptions (
    id integer NOT NULL,
    taxon_id integer,
    locale character varying(255),
    body text
);


--
-- Name: taxon_descriptions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE taxon_descriptions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: taxon_descriptions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE taxon_descriptions_id_seq OWNED BY taxon_descriptions.id;


--
-- Name: taxon_links; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE taxon_links (
    id integer NOT NULL,
    url character varying(255) NOT NULL,
    site_title character varying(255),
    taxon_id integer NOT NULL,
    show_for_descendent_taxa boolean DEFAULT false,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    user_id integer,
    place_id integer,
    species_only boolean DEFAULT false,
    short_title character varying(10)
);


--
-- Name: taxon_links_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE taxon_links_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: taxon_links_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE taxon_links_id_seq OWNED BY taxon_links.id;


--
-- Name: taxon_names; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE taxon_names (
    id integer NOT NULL,
    name character varying(255),
    is_valid boolean,
    lexicon character varying(255),
    source_identifier character varying(255),
    source_url character varying(255),
    taxon_id integer,
    source_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    name_provider character varying(255),
    creator_id integer,
    updater_id integer,
    "position" integer DEFAULT 0
);


--
-- Name: taxon_names_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE taxon_names_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: taxon_names_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE taxon_names_id_seq OWNED BY taxon_names.id;


--
-- Name: taxon_photos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE taxon_photos (
    id integer NOT NULL,
    taxon_id integer NOT NULL,
    photo_id integer NOT NULL,
    "position" integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: taxon_photos_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE taxon_photos_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: taxon_photos_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE taxon_photos_id_seq OWNED BY taxon_photos.id;


--
-- Name: taxon_ranges; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE taxon_ranges (
    id integer NOT NULL,
    taxon_id integer,
    source character varying(255),
    start_month integer,
    end_month integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    range_type character varying(255),
    range_content_type character varying(255),
    range_file_name character varying(255),
    range_file_size integer,
    description text,
    source_id integer,
    source_identifier integer,
    range_updated_at timestamp without time zone,
    geom geometry(MultiPolygon),
    url character varying(255)
);


--
-- Name: taxon_ranges_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE taxon_ranges_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: taxon_ranges_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE taxon_ranges_id_seq OWNED BY taxon_ranges.id;


--
-- Name: taxon_scheme_taxa; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE taxon_scheme_taxa (
    id integer NOT NULL,
    taxon_scheme_id integer,
    taxon_id integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    source_identifier character varying(255),
    taxon_name_id integer
);


--
-- Name: taxon_scheme_taxa_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE taxon_scheme_taxa_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: taxon_scheme_taxa_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE taxon_scheme_taxa_id_seq OWNED BY taxon_scheme_taxa.id;


--
-- Name: taxon_schemes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE taxon_schemes (
    id integer NOT NULL,
    title character varying(255),
    description text,
    source_id integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: taxon_schemes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE taxon_schemes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: taxon_schemes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE taxon_schemes_id_seq OWNED BY taxon_schemes.id;


--
-- Name: taxon_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE taxon_versions (
    id integer NOT NULL,
    taxon_id integer,
    version integer,
    name character varying(255),
    rank character varying(255),
    source_identifier character varying(255),
    source_url character varying(255),
    parent_id integer,
    source_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    iconic_taxon_id integer,
    is_iconic boolean DEFAULT false,
    auto_photos boolean DEFAULT true,
    auto_description boolean DEFAULT true,
    lft integer,
    rgt integer,
    name_provider character varying(255),
    delta boolean DEFAULT false,
    creator_id integer,
    updater_id integer,
    rank_level integer
);


--
-- Name: taxon_versions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE taxon_versions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: taxon_versions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE taxon_versions_id_seq OWNED BY taxon_versions.id;


--
-- Name: trip_purposes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE trip_purposes (
    id integer NOT NULL,
    trip_id integer,
    purpose character varying(255),
    resource_type character varying(255),
    resource_id integer,
    success boolean,
    complete boolean DEFAULT false,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: trip_purposes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE trip_purposes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: trip_purposes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE trip_purposes_id_seq OWNED BY trip_purposes.id;


--
-- Name: trip_taxa; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE trip_taxa (
    id integer NOT NULL,
    taxon_id integer,
    trip_id integer,
    observed boolean,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: trip_taxa_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE trip_taxa_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: trip_taxa_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE trip_taxa_id_seq OWNED BY trip_taxa.id;


--
-- Name: update_actions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE update_actions (
    id integer NOT NULL,
    resource_id integer,
    resource_type character varying,
    notifier_type character varying,
    notifier_id integer,
    notification character varying,
    resource_owner_id integer,
    created_at timestamp without time zone
);


--
-- Name: update_actions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE update_actions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: update_actions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE update_actions_id_seq OWNED BY update_actions.id;


--
-- Name: update_subscribers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE update_subscribers (
    id integer NOT NULL,
    update_action_id integer,
    subscriber_id integer,
    viewed_at timestamp without time zone
);


--
-- Name: update_subscribers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE update_subscribers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: update_subscribers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE update_subscribers_id_seq OWNED BY update_subscribers.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE users (
    id integer NOT NULL,
    login character varying(40),
    name character varying(100),
    email character varying(100),
    encrypted_password character varying(128) DEFAULT ''::character varying NOT NULL,
    password_salt character varying(255) DEFAULT ''::character varying NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    remember_token character varying(40),
    remember_token_expires_at timestamp without time zone,
    confirmation_token character varying(255),
    confirmed_at timestamp without time zone,
    state character varying(255) DEFAULT 'passive'::character varying,
    deleted_at timestamp without time zone,
    time_zone character varying(255),
    description text,
    life_list_id integer,
    observations_count integer DEFAULT 0,
    identifications_count integer DEFAULT 0,
    journal_posts_count integer DEFAULT 0,
    life_list_taxa_count integer DEFAULT 0,
    old_preferences text,
    icon_url character varying(255),
    last_ip character varying(255),
    confirmation_sent_at timestamp without time zone,
    reset_password_token character varying(255),
    reset_password_sent_at timestamp without time zone,
    remember_created_at timestamp without time zone,
    suspended_at timestamp without time zone,
    suspension_reason character varying(255),
    uri character varying(255),
    locale character varying(255),
    site_id integer,
    place_id integer,
    spammer boolean,
    spam_count integer DEFAULT 0,
    last_active date,
    subscriptions_suspended_at timestamp without time zone,
    test_groups character varying,
    latitude double precision,
    longitude double precision,
    lat_lon_acc_admin_level integer,
    icon_file_name character varying,
    icon_content_type character varying,
    icon_file_size integer,
    icon_updated_at timestamp without time zone
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE users_id_seq OWNED BY users.id;


--
-- Name: votes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE votes (
    id integer NOT NULL,
    votable_id integer,
    votable_type character varying,
    voter_id integer,
    voter_type character varying,
    vote_flag boolean,
    vote_scope character varying,
    vote_weight integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: votes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE votes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: votes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE votes_id_seq OWNED BY votes.id;


--
-- Name: wiki_page_attachments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE wiki_page_attachments (
    id integer NOT NULL,
    page_id integer NOT NULL,
    wiki_page_attachment_file_name character varying(255),
    wiki_page_attachment_content_type character varying(255),
    wiki_page_attachment_file_size integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: wiki_page_attachments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE wiki_page_attachments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: wiki_page_attachments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE wiki_page_attachments_id_seq OWNED BY wiki_page_attachments.id;


--
-- Name: wiki_page_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE wiki_page_versions (
    id integer NOT NULL,
    page_id integer NOT NULL,
    updator_id integer,
    number integer,
    comment character varying(255),
    path character varying(255),
    title character varying(255),
    content text,
    updated_at timestamp without time zone
);


--
-- Name: wiki_page_versions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE wiki_page_versions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: wiki_page_versions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE wiki_page_versions_id_seq OWNED BY wiki_page_versions.id;


--
-- Name: wiki_pages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE wiki_pages (
    id integer NOT NULL,
    creator_id integer,
    updator_id integer,
    path character varying(255),
    title character varying(255),
    content text,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: wiki_pages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE wiki_pages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: wiki_pages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE wiki_pages_id_seq OWNED BY wiki_pages.id;


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY annotations ALTER COLUMN id SET DEFAULT nextval('annotations_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY announcements ALTER COLUMN id SET DEFAULT nextval('announcements_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY api_endpoint_caches ALTER COLUMN id SET DEFAULT nextval('api_endpoint_caches_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY api_endpoints ALTER COLUMN id SET DEFAULT nextval('api_endpoints_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY assessment_sections ALTER COLUMN id SET DEFAULT nextval('assessment_sections_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY assessments ALTER COLUMN id SET DEFAULT nextval('assessments_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY atlas_alterations ALTER COLUMN id SET DEFAULT nextval('atlas_alterations_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY atlases ALTER COLUMN id SET DEFAULT nextval('atlases_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY colors ALTER COLUMN id SET DEFAULT nextval('colors_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY comments ALTER COLUMN id SET DEFAULT nextval('comments_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY complete_sets ALTER COLUMN id SET DEFAULT nextval('complete_sets_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY conservation_statuses ALTER COLUMN id SET DEFAULT nextval('conservation_statuses_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY controlled_term_labels ALTER COLUMN id SET DEFAULT nextval('controlled_term_labels_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY controlled_term_values ALTER COLUMN id SET DEFAULT nextval('controlled_term_values_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY controlled_terms ALTER COLUMN id SET DEFAULT nextval('controlled_terms_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY counties_simplified_01 ALTER COLUMN id SET DEFAULT nextval('counties_simplified_01_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY countries_simplified_1 ALTER COLUMN id SET DEFAULT nextval('countries_simplified_1_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY custom_projects ALTER COLUMN id SET DEFAULT nextval('custom_projects_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY delayed_jobs ALTER COLUMN id SET DEFAULT nextval('delayed_jobs_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY deleted_observations ALTER COLUMN id SET DEFAULT nextval('deleted_observations_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY deleted_photos ALTER COLUMN id SET DEFAULT nextval('deleted_photos_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY deleted_users ALTER COLUMN id SET DEFAULT nextval('deleted_users_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY exploded_atlas_places ALTER COLUMN id SET DEFAULT nextval('exploded_atlas_places_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY flags ALTER COLUMN id SET DEFAULT nextval('flags_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY flickr_identities ALTER COLUMN id SET DEFAULT nextval('flickr_identities_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY flow_task_resources ALTER COLUMN id SET DEFAULT nextval('flow_task_resources_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY flow_tasks ALTER COLUMN id SET DEFAULT nextval('flow_tasks_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY friendly_id_slugs ALTER COLUMN id SET DEFAULT nextval('friendly_id_slugs_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY friendships ALTER COLUMN id SET DEFAULT nextval('friendships_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY goal_contributions ALTER COLUMN id SET DEFAULT nextval('goal_contributions_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY goal_participants ALTER COLUMN id SET DEFAULT nextval('goal_participants_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY goal_rules ALTER COLUMN id SET DEFAULT nextval('goal_rules_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY goals ALTER COLUMN id SET DEFAULT nextval('goals_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY guide_photos ALTER COLUMN id SET DEFAULT nextval('guide_photos_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY guide_ranges ALTER COLUMN id SET DEFAULT nextval('guide_ranges_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY guide_sections ALTER COLUMN id SET DEFAULT nextval('guide_sections_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY guide_taxa ALTER COLUMN id SET DEFAULT nextval('guide_taxa_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY guide_users ALTER COLUMN id SET DEFAULT nextval('guide_users_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY guides ALTER COLUMN id SET DEFAULT nextval('guides_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY identifications ALTER COLUMN id SET DEFAULT nextval('identifications_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY invites ALTER COLUMN id SET DEFAULT nextval('invites_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY list_rules ALTER COLUMN id SET DEFAULT nextval('list_rules_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY listed_taxa ALTER COLUMN id SET DEFAULT nextval('listed_taxa_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY listed_taxon_alterations ALTER COLUMN id SET DEFAULT nextval('listed_taxon_alterations_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY lists ALTER COLUMN id SET DEFAULT nextval('lists_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY messages ALTER COLUMN id SET DEFAULT nextval('messages_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY model_attribute_changes ALTER COLUMN id SET DEFAULT nextval('model_attribute_changes_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY oauth_access_grants ALTER COLUMN id SET DEFAULT nextval('oauth_access_grants_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY oauth_access_tokens ALTER COLUMN id SET DEFAULT nextval('oauth_access_tokens_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY oauth_applications ALTER COLUMN id SET DEFAULT nextval('oauth_applications_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY observation_field_values ALTER COLUMN id SET DEFAULT nextval('observation_field_values_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY observation_fields ALTER COLUMN id SET DEFAULT nextval('observation_fields_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY observation_links ALTER COLUMN id SET DEFAULT nextval('observation_links_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY observation_photos ALTER COLUMN id SET DEFAULT nextval('observation_photos_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY observation_reviews ALTER COLUMN id SET DEFAULT nextval('observation_reviews_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY observation_sounds ALTER COLUMN id SET DEFAULT nextval('observation_sounds_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY observations ALTER COLUMN id SET DEFAULT nextval('observations_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY observations_places ALTER COLUMN id SET DEFAULT nextval('observations_places_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY passwords ALTER COLUMN id SET DEFAULT nextval('passwords_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY photos ALTER COLUMN id SET DEFAULT nextval('photos_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY picasa_identities ALTER COLUMN id SET DEFAULT nextval('picasa_identities_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY place_geometries ALTER COLUMN id SET DEFAULT nextval('place_geometries_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY place_taxon_names ALTER COLUMN id SET DEFAULT nextval('place_taxon_names_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY places ALTER COLUMN id SET DEFAULT nextval('places_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY posts ALTER COLUMN id SET DEFAULT nextval('posts_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY preferences ALTER COLUMN id SET DEFAULT nextval('preferences_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY project_assets ALTER COLUMN id SET DEFAULT nextval('project_assets_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY project_invitations ALTER COLUMN id SET DEFAULT nextval('project_invitations_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY project_observation_fields ALTER COLUMN id SET DEFAULT nextval('project_observation_fields_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY project_observations ALTER COLUMN id SET DEFAULT nextval('project_observations_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY project_user_invitations ALTER COLUMN id SET DEFAULT nextval('project_user_invitations_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY project_users ALTER COLUMN id SET DEFAULT nextval('project_users_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY projects ALTER COLUMN id SET DEFAULT nextval('projects_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY provider_authorizations ALTER COLUMN id SET DEFAULT nextval('provider_authorizations_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY quality_metrics ALTER COLUMN id SET DEFAULT nextval('quality_metrics_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY roles ALTER COLUMN id SET DEFAULT nextval('roles_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY rules ALTER COLUMN id SET DEFAULT nextval('rules_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY site_admins ALTER COLUMN id SET DEFAULT nextval('site_admins_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY site_statistics ALTER COLUMN id SET DEFAULT nextval('site_statistics_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY sites ALTER COLUMN id SET DEFAULT nextval('sites_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY soundcloud_identities ALTER COLUMN id SET DEFAULT nextval('soundcloud_identities_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY sounds ALTER COLUMN id SET DEFAULT nextval('sounds_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY sources ALTER COLUMN id SET DEFAULT nextval('sources_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY states_simplified_1 ALTER COLUMN id SET DEFAULT nextval('states_simplified_1_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY subscriptions ALTER COLUMN id SET DEFAULT nextval('subscriptions_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY taggings ALTER COLUMN id SET DEFAULT nextval('taggings_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY tags ALTER COLUMN id SET DEFAULT nextval('tags_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxa ALTER COLUMN id SET DEFAULT nextval('taxa_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_change_taxa ALTER COLUMN id SET DEFAULT nextval('taxon_change_taxa_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_changes ALTER COLUMN id SET DEFAULT nextval('taxon_changes_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_descriptions ALTER COLUMN id SET DEFAULT nextval('taxon_descriptions_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_links ALTER COLUMN id SET DEFAULT nextval('taxon_links_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_names ALTER COLUMN id SET DEFAULT nextval('taxon_names_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_photos ALTER COLUMN id SET DEFAULT nextval('taxon_photos_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_ranges ALTER COLUMN id SET DEFAULT nextval('taxon_ranges_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_scheme_taxa ALTER COLUMN id SET DEFAULT nextval('taxon_scheme_taxa_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_schemes ALTER COLUMN id SET DEFAULT nextval('taxon_schemes_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_versions ALTER COLUMN id SET DEFAULT nextval('taxon_versions_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY trip_purposes ALTER COLUMN id SET DEFAULT nextval('trip_purposes_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY trip_taxa ALTER COLUMN id SET DEFAULT nextval('trip_taxa_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY update_actions ALTER COLUMN id SET DEFAULT nextval('update_actions_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY update_subscribers ALTER COLUMN id SET DEFAULT nextval('update_subscribers_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY users ALTER COLUMN id SET DEFAULT nextval('users_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY votes ALTER COLUMN id SET DEFAULT nextval('votes_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY wiki_page_attachments ALTER COLUMN id SET DEFAULT nextval('wiki_page_attachments_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY wiki_page_versions ALTER COLUMN id SET DEFAULT nextval('wiki_page_versions_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY wiki_pages ALTER COLUMN id SET DEFAULT nextval('wiki_pages_id_seq'::regclass);


--
-- Name: annotations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY annotations
    ADD CONSTRAINT annotations_pkey PRIMARY KEY (id);


--
-- Name: announcements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY announcements
    ADD CONSTRAINT announcements_pkey PRIMARY KEY (id);


--
-- Name: api_endpoint_caches_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY api_endpoint_caches
    ADD CONSTRAINT api_endpoint_caches_pkey PRIMARY KEY (id);


--
-- Name: api_endpoints_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY api_endpoints
    ADD CONSTRAINT api_endpoints_pkey PRIMARY KEY (id);


--
-- Name: assessment_sections_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY assessment_sections
    ADD CONSTRAINT assessment_sections_pkey PRIMARY KEY (id);


--
-- Name: assessments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY assessments
    ADD CONSTRAINT assessments_pkey PRIMARY KEY (id);


--
-- Name: atlas_alterations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY atlas_alterations
    ADD CONSTRAINT atlas_alterations_pkey PRIMARY KEY (id);


--
-- Name: atlases_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY atlases
    ADD CONSTRAINT atlases_pkey PRIMARY KEY (id);


--
-- Name: colors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY colors
    ADD CONSTRAINT colors_pkey PRIMARY KEY (id);


--
-- Name: comments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);


--
-- Name: complete_sets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY complete_sets
    ADD CONSTRAINT complete_sets_pkey PRIMARY KEY (id);


--
-- Name: conservation_statuses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY conservation_statuses
    ADD CONSTRAINT conservation_statuses_pkey PRIMARY KEY (id);


--
-- Name: controlled_term_labels_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY controlled_term_labels
    ADD CONSTRAINT controlled_term_labels_pkey PRIMARY KEY (id);


--
-- Name: controlled_term_values_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY controlled_term_values
    ADD CONSTRAINT controlled_term_values_pkey PRIMARY KEY (id);


--
-- Name: controlled_terms_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY controlled_terms
    ADD CONSTRAINT controlled_terms_pkey PRIMARY KEY (id);


--
-- Name: counties_simplified_01_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY counties_simplified_01
    ADD CONSTRAINT counties_simplified_01_pkey PRIMARY KEY (id);


--
-- Name: countries_simplified_1_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY countries_simplified_1
    ADD CONSTRAINT countries_simplified_1_pkey PRIMARY KEY (id);


--
-- Name: custom_projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY custom_projects
    ADD CONSTRAINT custom_projects_pkey PRIMARY KEY (id);


--
-- Name: delayed_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY delayed_jobs
    ADD CONSTRAINT delayed_jobs_pkey PRIMARY KEY (id);


--
-- Name: deleted_observations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY deleted_observations
    ADD CONSTRAINT deleted_observations_pkey PRIMARY KEY (id);


--
-- Name: deleted_photos_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY deleted_photos
    ADD CONSTRAINT deleted_photos_pkey PRIMARY KEY (id);


--
-- Name: deleted_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY deleted_users
    ADD CONSTRAINT deleted_users_pkey PRIMARY KEY (id);


--
-- Name: exploded_atlas_places_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY exploded_atlas_places
    ADD CONSTRAINT exploded_atlas_places_pkey PRIMARY KEY (id);


--
-- Name: flags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY flags
    ADD CONSTRAINT flags_pkey PRIMARY KEY (id);


--
-- Name: flickr_identities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY flickr_identities
    ADD CONSTRAINT flickr_identities_pkey PRIMARY KEY (id);


--
-- Name: flow_task_resources_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY flow_task_resources
    ADD CONSTRAINT flow_task_resources_pkey PRIMARY KEY (id);


--
-- Name: flow_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY flow_tasks
    ADD CONSTRAINT flow_tasks_pkey PRIMARY KEY (id);


--
-- Name: friendly_id_slugs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY friendly_id_slugs
    ADD CONSTRAINT friendly_id_slugs_pkey PRIMARY KEY (id);


--
-- Name: friendships_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY friendships
    ADD CONSTRAINT friendships_pkey PRIMARY KEY (id);


--
-- Name: goal_contributions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY goal_contributions
    ADD CONSTRAINT goal_contributions_pkey PRIMARY KEY (id);


--
-- Name: goal_participants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY goal_participants
    ADD CONSTRAINT goal_participants_pkey PRIMARY KEY (id);


--
-- Name: goal_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY goal_rules
    ADD CONSTRAINT goal_rules_pkey PRIMARY KEY (id);


--
-- Name: goals_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY goals
    ADD CONSTRAINT goals_pkey PRIMARY KEY (id);


--
-- Name: guide_photos_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY guide_photos
    ADD CONSTRAINT guide_photos_pkey PRIMARY KEY (id);


--
-- Name: guide_ranges_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY guide_ranges
    ADD CONSTRAINT guide_ranges_pkey PRIMARY KEY (id);


--
-- Name: guide_sections_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY guide_sections
    ADD CONSTRAINT guide_sections_pkey PRIMARY KEY (id);


--
-- Name: guide_taxa_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY guide_taxa
    ADD CONSTRAINT guide_taxa_pkey PRIMARY KEY (id);


--
-- Name: guide_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY guide_users
    ADD CONSTRAINT guide_users_pkey PRIMARY KEY (id);


--
-- Name: guides_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY guides
    ADD CONSTRAINT guides_pkey PRIMARY KEY (id);


--
-- Name: identifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY identifications
    ADD CONSTRAINT identifications_pkey PRIMARY KEY (id);


--
-- Name: invites_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY invites
    ADD CONSTRAINT invites_pkey PRIMARY KEY (id);


--
-- Name: list_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY list_rules
    ADD CONSTRAINT list_rules_pkey PRIMARY KEY (id);


--
-- Name: listed_taxa_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY listed_taxa
    ADD CONSTRAINT listed_taxa_pkey PRIMARY KEY (id);


--
-- Name: listed_taxon_alterations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY listed_taxon_alterations
    ADD CONSTRAINT listed_taxon_alterations_pkey PRIMARY KEY (id);


--
-- Name: lists_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY lists
    ADD CONSTRAINT lists_pkey PRIMARY KEY (id);


--
-- Name: messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);


--
-- Name: model_attribute_changes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY model_attribute_changes
    ADD CONSTRAINT model_attribute_changes_pkey PRIMARY KEY (id);


--
-- Name: oauth_access_grants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY oauth_access_grants
    ADD CONSTRAINT oauth_access_grants_pkey PRIMARY KEY (id);


--
-- Name: oauth_access_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY oauth_access_tokens
    ADD CONSTRAINT oauth_access_tokens_pkey PRIMARY KEY (id);


--
-- Name: oauth_applications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY oauth_applications
    ADD CONSTRAINT oauth_applications_pkey PRIMARY KEY (id);


--
-- Name: observation_field_values_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY observation_field_values
    ADD CONSTRAINT observation_field_values_pkey PRIMARY KEY (id);


--
-- Name: observation_fields_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY observation_fields
    ADD CONSTRAINT observation_fields_pkey PRIMARY KEY (id);


--
-- Name: observation_links_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY observation_links
    ADD CONSTRAINT observation_links_pkey PRIMARY KEY (id);


--
-- Name: observation_photos_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY observation_photos
    ADD CONSTRAINT observation_photos_pkey PRIMARY KEY (id);


--
-- Name: observation_reviews_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY observation_reviews
    ADD CONSTRAINT observation_reviews_pkey PRIMARY KEY (id);


--
-- Name: observations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY observations
    ADD CONSTRAINT observations_pkey PRIMARY KEY (id);


--
-- Name: observations_places_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY observations_places
    ADD CONSTRAINT observations_places_pkey PRIMARY KEY (id);


--
-- Name: observations_sounds_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY observation_sounds
    ADD CONSTRAINT observations_sounds_pkey PRIMARY KEY (id);


--
-- Name: passwords_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY passwords
    ADD CONSTRAINT passwords_pkey PRIMARY KEY (id);


--
-- Name: photos_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY photos
    ADD CONSTRAINT photos_pkey PRIMARY KEY (id);


--
-- Name: picasa_identities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY picasa_identities
    ADD CONSTRAINT picasa_identities_pkey PRIMARY KEY (id);


--
-- Name: place_geometries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY place_geometries
    ADD CONSTRAINT place_geometries_pkey PRIMARY KEY (id);


--
-- Name: place_taxon_names_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY place_taxon_names
    ADD CONSTRAINT place_taxon_names_pkey PRIMARY KEY (id);


--
-- Name: places_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY places
    ADD CONSTRAINT places_pkey PRIMARY KEY (id);


--
-- Name: posts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY posts
    ADD CONSTRAINT posts_pkey PRIMARY KEY (id);


--
-- Name: preferences_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY preferences
    ADD CONSTRAINT preferences_pkey PRIMARY KEY (id);


--
-- Name: project_assets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY project_assets
    ADD CONSTRAINT project_assets_pkey PRIMARY KEY (id);


--
-- Name: project_invitations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY project_invitations
    ADD CONSTRAINT project_invitations_pkey PRIMARY KEY (id);


--
-- Name: project_observation_fields_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY project_observation_fields
    ADD CONSTRAINT project_observation_fields_pkey PRIMARY KEY (id);


--
-- Name: project_observations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY project_observations
    ADD CONSTRAINT project_observations_pkey PRIMARY KEY (id);


--
-- Name: project_user_invitations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY project_user_invitations
    ADD CONSTRAINT project_user_invitations_pkey PRIMARY KEY (id);


--
-- Name: project_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY project_users
    ADD CONSTRAINT project_users_pkey PRIMARY KEY (id);


--
-- Name: projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);


--
-- Name: provider_authorizations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY provider_authorizations
    ADD CONSTRAINT provider_authorizations_pkey PRIMARY KEY (id);


--
-- Name: quality_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY quality_metrics
    ADD CONSTRAINT quality_metrics_pkey PRIMARY KEY (id);


--
-- Name: roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: rules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY rules
    ADD CONSTRAINT rules_pkey PRIMARY KEY (id);


--
-- Name: site_admins_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY site_admins
    ADD CONSTRAINT site_admins_pkey PRIMARY KEY (id);


--
-- Name: site_statistics_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY site_statistics
    ADD CONSTRAINT site_statistics_pkey PRIMARY KEY (id);


--
-- Name: sites_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY sites
    ADD CONSTRAINT sites_pkey PRIMARY KEY (id);


--
-- Name: soundcloud_identities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY soundcloud_identities
    ADD CONSTRAINT soundcloud_identities_pkey PRIMARY KEY (id);


--
-- Name: sounds_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY sounds
    ADD CONSTRAINT sounds_pkey PRIMARY KEY (id);


--
-- Name: sources_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY sources
    ADD CONSTRAINT sources_pkey PRIMARY KEY (id);


--
-- Name: states_simplified_1_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY states_simplified_1
    ADD CONSTRAINT states_simplified_1_pkey PRIMARY KEY (id);


--
-- Name: subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (id);


--
-- Name: taggings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY taggings
    ADD CONSTRAINT taggings_pkey PRIMARY KEY (id);


--
-- Name: tags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (id);


--
-- Name: taxa_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxa
    ADD CONSTRAINT taxa_pkey PRIMARY KEY (id);


--
-- Name: taxon_change_taxa_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_change_taxa
    ADD CONSTRAINT taxon_change_taxa_pkey PRIMARY KEY (id);


--
-- Name: taxon_changes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_changes
    ADD CONSTRAINT taxon_changes_pkey PRIMARY KEY (id);


--
-- Name: taxon_descriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_descriptions
    ADD CONSTRAINT taxon_descriptions_pkey PRIMARY KEY (id);


--
-- Name: taxon_links_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_links
    ADD CONSTRAINT taxon_links_pkey PRIMARY KEY (id);


--
-- Name: taxon_names_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_names
    ADD CONSTRAINT taxon_names_pkey PRIMARY KEY (id);


--
-- Name: taxon_photos_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_photos
    ADD CONSTRAINT taxon_photos_pkey PRIMARY KEY (id);


--
-- Name: taxon_ranges_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_ranges
    ADD CONSTRAINT taxon_ranges_pkey PRIMARY KEY (id);


--
-- Name: taxon_scheme_taxa_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_scheme_taxa
    ADD CONSTRAINT taxon_scheme_taxa_pkey PRIMARY KEY (id);


--
-- Name: taxon_schemes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_schemes
    ADD CONSTRAINT taxon_schemes_pkey PRIMARY KEY (id);


--
-- Name: taxon_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY taxon_versions
    ADD CONSTRAINT taxon_versions_pkey PRIMARY KEY (id);


--
-- Name: trip_purposes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY trip_purposes
    ADD CONSTRAINT trip_purposes_pkey PRIMARY KEY (id);


--
-- Name: trip_taxa_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY trip_taxa
    ADD CONSTRAINT trip_taxa_pkey PRIMARY KEY (id);


--
-- Name: update_actions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY update_actions
    ADD CONSTRAINT update_actions_pkey PRIMARY KEY (id);


--
-- Name: update_subscribers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY update_subscribers
    ADD CONSTRAINT update_subscribers_pkey PRIMARY KEY (id);


--
-- Name: users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: votes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY votes
    ADD CONSTRAINT votes_pkey PRIMARY KEY (id);


--
-- Name: wiki_page_attachments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY wiki_page_attachments
    ADD CONSTRAINT wiki_page_attachments_pkey PRIMARY KEY (id);


--
-- Name: wiki_page_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY wiki_page_versions
    ADD CONSTRAINT wiki_page_versions_pkey PRIMARY KEY (id);


--
-- Name: wiki_pages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY wiki_pages
    ADD CONSTRAINT wiki_pages_pkey PRIMARY KEY (id);


--
-- Name: fk_flags_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX fk_flags_user ON flags USING btree (user_id);


--
-- Name: index_annotations_on_resource_id_and_resource_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_annotations_on_resource_id_and_resource_type ON annotations USING btree (resource_id, resource_type);


--
-- Name: index_announcements_on_site_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_announcements_on_site_id ON announcements USING btree (site_id);


--
-- Name: index_announcements_on_start_and_end; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_announcements_on_start_and_end ON announcements USING btree (start, "end");


--
-- Name: index_api_endpoint_caches_on_api_endpoint_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_api_endpoint_caches_on_api_endpoint_id ON api_endpoint_caches USING btree (api_endpoint_id);


--
-- Name: index_api_endpoint_caches_on_request_url; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_api_endpoint_caches_on_request_url ON api_endpoint_caches USING btree (request_url);


--
-- Name: index_api_endpoints_on_title; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_api_endpoints_on_title ON api_endpoints USING btree (title);


--
-- Name: index_assessment_sections_on_assessment_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_assessment_sections_on_assessment_id ON assessment_sections USING btree (assessment_id);


--
-- Name: index_assessment_sections_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_assessment_sections_on_user_id ON assessment_sections USING btree (user_id);


--
-- Name: index_assessments_on_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_assessments_on_project_id ON assessments USING btree (project_id);


--
-- Name: index_assessments_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_assessments_on_taxon_id ON assessments USING btree (taxon_id);


--
-- Name: index_assessments_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_assessments_on_user_id ON assessments USING btree (user_id);


--
-- Name: index_atlas_alterations_on_atlas_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_atlas_alterations_on_atlas_id ON atlas_alterations USING btree (atlas_id);


--
-- Name: index_atlas_alterations_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_atlas_alterations_on_place_id ON atlas_alterations USING btree (place_id);


--
-- Name: index_atlas_alterations_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_atlas_alterations_on_user_id ON atlas_alterations USING btree (user_id);


--
-- Name: index_atlases_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_atlases_on_taxon_id ON atlases USING btree (taxon_id);


--
-- Name: index_atlases_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_atlases_on_user_id ON atlases USING btree (user_id);


--
-- Name: index_colors_taxa_on_taxon_id_and_color_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_colors_taxa_on_taxon_id_and_color_id ON colors_taxa USING btree (taxon_id, color_id);


--
-- Name: index_comments_on_parent_type_and_parent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_comments_on_parent_type_and_parent_id ON comments USING btree (parent_type, parent_id);


--
-- Name: index_comments_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_comments_on_user_id ON comments USING btree (user_id);


--
-- Name: index_comments_on_uuid; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_comments_on_uuid ON comments USING btree (uuid);


--
-- Name: index_complete_sets_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_complete_sets_on_place_id ON complete_sets USING btree (place_id);


--
-- Name: index_complete_sets_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_complete_sets_on_taxon_id ON complete_sets USING btree (taxon_id);


--
-- Name: index_complete_sets_on_taxon_id_and_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_complete_sets_on_taxon_id_and_place_id ON complete_sets USING btree (taxon_id, place_id);


--
-- Name: index_complete_sets_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_complete_sets_on_user_id ON complete_sets USING btree (user_id);


--
-- Name: index_conservation_statuses_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_conservation_statuses_on_place_id ON conservation_statuses USING btree (place_id);


--
-- Name: index_conservation_statuses_on_source_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_conservation_statuses_on_source_id ON conservation_statuses USING btree (source_id);


--
-- Name: index_conservation_statuses_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_conservation_statuses_on_taxon_id ON conservation_statuses USING btree (taxon_id);


--
-- Name: index_conservation_statuses_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_conservation_statuses_on_user_id ON conservation_statuses USING btree (user_id);


--
-- Name: index_counties_simplified_01_on_geom; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_counties_simplified_01_on_geom ON counties_simplified_01 USING gist (geom);


--
-- Name: index_counties_simplified_01_on_place_geometry_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_counties_simplified_01_on_place_geometry_id ON counties_simplified_01 USING btree (place_geometry_id);


--
-- Name: index_counties_simplified_01_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_counties_simplified_01_on_place_id ON counties_simplified_01 USING btree (place_id);


--
-- Name: index_countries_simplified_1_on_geom; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_countries_simplified_1_on_geom ON countries_simplified_1 USING gist (geom);


--
-- Name: index_countries_simplified_1_on_place_geometry_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_countries_simplified_1_on_place_geometry_id ON countries_simplified_1 USING btree (place_geometry_id);


--
-- Name: index_countries_simplified_1_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_countries_simplified_1_on_place_id ON countries_simplified_1 USING btree (place_id);


--
-- Name: index_custom_projects_on_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_custom_projects_on_project_id ON custom_projects USING btree (project_id);


--
-- Name: index_delayed_jobs_on_unique_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_delayed_jobs_on_unique_hash ON delayed_jobs USING btree (unique_hash);


--
-- Name: index_deleted_observations_on_user_id_and_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_deleted_observations_on_user_id_and_created_at ON deleted_observations USING btree (user_id, created_at);


--
-- Name: index_deleted_photos_on_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_deleted_photos_on_created_at ON deleted_photos USING btree (created_at);


--
-- Name: index_deleted_users_on_login; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_deleted_users_on_login ON deleted_users USING btree (login);


--
-- Name: index_deleted_users_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_deleted_users_on_user_id ON deleted_users USING btree (user_id);


--
-- Name: index_exploded_atlas_places_on_atlas_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_exploded_atlas_places_on_atlas_id ON exploded_atlas_places USING btree (atlas_id);


--
-- Name: index_exploded_atlas_places_on_atlas_id_and_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_exploded_atlas_places_on_atlas_id_and_place_id ON exploded_atlas_places USING btree (atlas_id, place_id);


--
-- Name: index_exploded_atlas_places_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_exploded_atlas_places_on_place_id ON exploded_atlas_places USING btree (place_id);


--
-- Name: index_flickr_photos_on_flickr_native_photo_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_flickr_photos_on_flickr_native_photo_id ON photos USING btree (native_photo_id);


--
-- Name: index_flow_task_resources_on_flow_task_id_and_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_flow_task_resources_on_flow_task_id_and_type ON flow_task_resources USING btree (flow_task_id, type);


--
-- Name: index_flow_task_resources_on_resource_type_and_resource_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_flow_task_resources_on_resource_type_and_resource_id ON flow_task_resources USING btree (resource_type, resource_id);


--
-- Name: index_flow_tasks_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_flow_tasks_on_user_id ON flow_tasks USING btree (user_id);


--
-- Name: index_friendships_on_user_id_and_friend_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_friendships_on_user_id_and_friend_id ON friendships USING btree (user_id, friend_id);


--
-- Name: index_guide_photos_on_guide_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_guide_photos_on_guide_taxon_id ON guide_photos USING btree (guide_taxon_id);


--
-- Name: index_guide_photos_on_photo_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_guide_photos_on_photo_id ON guide_photos USING btree (photo_id);


--
-- Name: index_guide_ranges_on_guide_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_guide_ranges_on_guide_taxon_id ON guide_ranges USING btree (guide_taxon_id);


--
-- Name: index_guide_ranges_on_source_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_guide_ranges_on_source_id ON guide_ranges USING btree (source_id);


--
-- Name: index_guide_sections_on_creator_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_guide_sections_on_creator_id ON guide_sections USING btree (creator_id);


--
-- Name: index_guide_sections_on_guide_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_guide_sections_on_guide_taxon_id ON guide_sections USING btree (guide_taxon_id);


--
-- Name: index_guide_sections_on_source_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_guide_sections_on_source_id ON guide_sections USING btree (source_id);


--
-- Name: index_guide_sections_on_updater_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_guide_sections_on_updater_id ON guide_sections USING btree (updater_id);


--
-- Name: index_guide_taxa_on_guide_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_guide_taxa_on_guide_id ON guide_taxa USING btree (guide_id);


--
-- Name: index_guide_taxa_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_guide_taxa_on_taxon_id ON guide_taxa USING btree (taxon_id);


--
-- Name: index_guide_users_on_guide_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_guide_users_on_guide_id ON guide_users USING btree (guide_id);


--
-- Name: index_guide_users_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_guide_users_on_user_id ON guide_users USING btree (user_id);


--
-- Name: index_guides_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_guides_on_place_id ON guides USING btree (place_id);


--
-- Name: index_guides_on_source_url; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_guides_on_source_url ON guides USING btree (source_url);


--
-- Name: index_guides_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_guides_on_taxon_id ON guides USING btree (taxon_id);


--
-- Name: index_guides_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_guides_on_user_id ON guides USING btree (user_id);


--
-- Name: index_identifications_on_category; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_identifications_on_category ON identifications USING btree (category);


--
-- Name: index_identifications_on_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_identifications_on_created_at ON identifications USING btree (created_at);


--
-- Name: index_identifications_on_observation_id_and_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_identifications_on_observation_id_and_created_at ON identifications USING btree (observation_id, created_at);


--
-- Name: index_identifications_on_previous_observation_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_identifications_on_previous_observation_taxon_id ON identifications USING btree (previous_observation_taxon_id);


--
-- Name: index_identifications_on_taxon_change_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_identifications_on_taxon_change_id ON identifications USING btree (taxon_change_id);


--
-- Name: index_identifications_on_user_id_and_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_identifications_on_user_id_and_created_at ON identifications USING btree (user_id, created_at);


--
-- Name: index_identifications_on_user_id_and_current; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_identifications_on_user_id_and_current ON identifications USING btree (user_id, current);


--
-- Name: index_identifications_on_uuid; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_identifications_on_uuid ON identifications USING btree (uuid);


--
-- Name: index_list_rules_on_list_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_list_rules_on_list_id ON list_rules USING btree (list_id);


--
-- Name: index_list_rules_on_operand_type_and_operand_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_list_rules_on_operand_type_and_operand_id ON list_rules USING btree (operand_type, operand_id);


--
-- Name: index_listed_taxa_on_first_observation_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_listed_taxa_on_first_observation_id ON listed_taxa USING btree (first_observation_id);


--
-- Name: index_listed_taxa_on_last_observation_id_and_list_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_listed_taxa_on_last_observation_id_and_list_id ON listed_taxa USING btree (last_observation_id, list_id);


--
-- Name: index_listed_taxa_on_list_id_and_taxon_ancestor_ids_and_taxon_i; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_listed_taxa_on_list_id_and_taxon_ancestor_ids_and_taxon_i ON listed_taxa USING btree (list_id, taxon_ancestor_ids, taxon_id);


--
-- Name: index_listed_taxa_on_list_id_and_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_listed_taxa_on_list_id_and_taxon_id ON listed_taxa USING btree (list_id, taxon_id);


--
-- Name: index_listed_taxa_on_place_id_and_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_listed_taxa_on_place_id_and_created_at ON listed_taxa USING btree (place_id, created_at);


--
-- Name: index_listed_taxa_on_place_id_and_observations_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_listed_taxa_on_place_id_and_observations_count ON listed_taxa USING btree (place_id, observations_count);


--
-- Name: index_listed_taxa_on_place_id_and_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_listed_taxa_on_place_id_and_taxon_id ON listed_taxa USING btree (place_id, taxon_id);


--
-- Name: index_listed_taxa_on_source_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_listed_taxa_on_source_id ON listed_taxa USING btree (source_id);


--
-- Name: index_listed_taxa_on_taxon_ancestor_ids; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_listed_taxa_on_taxon_ancestor_ids ON listed_taxa USING btree (taxon_ancestor_ids);


--
-- Name: index_listed_taxa_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_listed_taxa_on_taxon_id ON listed_taxa USING btree (taxon_id);


--
-- Name: index_listed_taxa_on_taxon_range_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_listed_taxa_on_taxon_range_id ON listed_taxa USING btree (taxon_range_id);


--
-- Name: index_listed_taxa_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_listed_taxa_on_user_id ON listed_taxa USING btree (user_id);


--
-- Name: index_listed_taxon_alterations_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_listed_taxon_alterations_on_place_id ON listed_taxon_alterations USING btree (place_id);


--
-- Name: index_listed_taxon_alterations_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_listed_taxon_alterations_on_taxon_id ON listed_taxon_alterations USING btree (taxon_id);


--
-- Name: index_listed_taxon_alterations_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_listed_taxon_alterations_on_user_id ON listed_taxon_alterations USING btree (user_id);


--
-- Name: index_lists_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_lists_on_place_id ON lists USING btree (place_id);


--
-- Name: index_lists_on_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_lists_on_project_id ON lists USING btree (project_id);


--
-- Name: index_lists_on_source_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_lists_on_source_id ON lists USING btree (source_id);


--
-- Name: index_lists_on_type_and_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_lists_on_type_and_id ON lists USING btree (type, id);


--
-- Name: index_lists_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_lists_on_user_id ON lists USING btree (user_id);


--
-- Name: index_messages_on_user_id_and_from_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_messages_on_user_id_and_from_user_id ON messages USING btree (user_id, from_user_id);


--
-- Name: index_messages_on_user_id_and_to_user_id_and_read_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_messages_on_user_id_and_to_user_id_and_read_at ON messages USING btree (user_id, to_user_id, read_at);


--
-- Name: index_model_attribute_changes_on_changed_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_model_attribute_changes_on_changed_at ON model_attribute_changes USING btree (changed_at);


--
-- Name: index_model_attribute_changes_on_model_id_and_field_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_model_attribute_changes_on_model_id_and_field_name ON model_attribute_changes USING btree (model_id, field_name);


--
-- Name: index_oauth_access_grants_on_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_oauth_access_grants_on_token ON oauth_access_grants USING btree (token);


--
-- Name: index_oauth_access_tokens_on_refresh_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_oauth_access_tokens_on_refresh_token ON oauth_access_tokens USING btree (refresh_token);


--
-- Name: index_oauth_access_tokens_on_resource_owner_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_oauth_access_tokens_on_resource_owner_id ON oauth_access_tokens USING btree (resource_owner_id);


--
-- Name: index_oauth_access_tokens_on_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_oauth_access_tokens_on_token ON oauth_access_tokens USING btree (token);


--
-- Name: index_oauth_applications_on_owner_id_and_owner_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_oauth_applications_on_owner_id_and_owner_type ON oauth_applications USING btree (owner_id, owner_type);


--
-- Name: index_oauth_applications_on_uid; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_oauth_applications_on_uid ON oauth_applications USING btree (uid);


--
-- Name: index_observation_field_values_on_observation_field_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_field_values_on_observation_field_id ON observation_field_values USING btree (observation_field_id);


--
-- Name: index_observation_field_values_on_observation_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_field_values_on_observation_id ON observation_field_values USING btree (observation_id);


--
-- Name: index_observation_field_values_on_updater_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_field_values_on_updater_id ON observation_field_values USING btree (updater_id);


--
-- Name: index_observation_field_values_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_field_values_on_user_id ON observation_field_values USING btree (user_id);


--
-- Name: index_observation_field_values_on_uuid; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_field_values_on_uuid ON observation_field_values USING btree (uuid);


--
-- Name: index_observation_fields_on_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_fields_on_name ON observation_fields USING btree (name);


--
-- Name: index_observation_links_on_observation_id_and_href; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_links_on_observation_id_and_href ON observation_links USING btree (observation_id, href);


--
-- Name: index_observation_photos_on_observation_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_photos_on_observation_id ON observation_photos USING btree (observation_id);


--
-- Name: index_observation_photos_on_photo_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_photos_on_photo_id ON observation_photos USING btree (photo_id);


--
-- Name: index_observation_photos_on_uuid; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_photos_on_uuid ON observation_photos USING btree (uuid);


--
-- Name: index_observation_reviews_on_observation_id_and_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_observation_reviews_on_observation_id_and_user_id ON observation_reviews USING btree (observation_id, user_id);


--
-- Name: index_observation_reviews_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_reviews_on_user_id ON observation_reviews USING btree (user_id);


--
-- Name: index_observation_zooms_10_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_10_on_taxon_id ON observation_zooms_10 USING btree (taxon_id);


--
-- Name: index_observation_zooms_11_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_11_on_taxon_id ON observation_zooms_11 USING btree (taxon_id);


--
-- Name: index_observation_zooms_125_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_125_on_taxon_id ON observation_zooms_125 USING btree (taxon_id);


--
-- Name: index_observation_zooms_12_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_12_on_taxon_id ON observation_zooms_12 USING btree (taxon_id);


--
-- Name: index_observation_zooms_2000_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_2000_on_taxon_id ON observation_zooms_2000 USING btree (taxon_id);


--
-- Name: index_observation_zooms_250_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_250_on_taxon_id ON observation_zooms_250 USING btree (taxon_id);


--
-- Name: index_observation_zooms_2_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_2_on_taxon_id ON observation_zooms_2 USING btree (taxon_id);


--
-- Name: index_observation_zooms_3_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_3_on_taxon_id ON observation_zooms_3 USING btree (taxon_id);


--
-- Name: index_observation_zooms_4000_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_4000_on_taxon_id ON observation_zooms_4000 USING btree (taxon_id);


--
-- Name: index_observation_zooms_4_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_4_on_taxon_id ON observation_zooms_4 USING btree (taxon_id);


--
-- Name: index_observation_zooms_500_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_500_on_taxon_id ON observation_zooms_500 USING btree (taxon_id);


--
-- Name: index_observation_zooms_5_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_5_on_taxon_id ON observation_zooms_5 USING btree (taxon_id);


--
-- Name: index_observation_zooms_63_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_63_on_taxon_id ON observation_zooms_63 USING btree (taxon_id);


--
-- Name: index_observation_zooms_6_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_6_on_taxon_id ON observation_zooms_6 USING btree (taxon_id);


--
-- Name: index_observation_zooms_7_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_7_on_taxon_id ON observation_zooms_7 USING btree (taxon_id);


--
-- Name: index_observation_zooms_8_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_8_on_taxon_id ON observation_zooms_8 USING btree (taxon_id);


--
-- Name: index_observation_zooms_990_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_990_on_taxon_id ON observation_zooms_990 USING btree (taxon_id);


--
-- Name: index_observation_zooms_9_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observation_zooms_9_on_taxon_id ON observation_zooms_9 USING btree (taxon_id);


--
-- Name: index_observations_on_cached_votes_total; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_cached_votes_total ON observations USING btree (cached_votes_total);


--
-- Name: index_observations_on_captive; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_captive ON observations USING btree (captive);


--
-- Name: index_observations_on_comments_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_comments_count ON observations USING btree (comments_count);


--
-- Name: index_observations_on_community_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_community_taxon_id ON observations USING btree (community_taxon_id);


--
-- Name: index_observations_on_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_created_at ON observations USING btree (created_at);


--
-- Name: index_observations_on_geom; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_geom ON observations USING gist (geom);


--
-- Name: index_observations_on_last_indexed_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_last_indexed_at ON observations USING btree (last_indexed_at);


--
-- Name: index_observations_on_mappable; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_mappable ON observations USING btree (mappable);


--
-- Name: index_observations_on_oauth_application_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_oauth_application_id ON observations USING btree (oauth_application_id);


--
-- Name: index_observations_on_observed_on; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_observed_on ON observations USING btree (observed_on);


--
-- Name: index_observations_on_observed_on_and_time_observed_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_observed_on_and_time_observed_at ON observations USING btree (observed_on, time_observed_at);


--
-- Name: index_observations_on_out_of_range; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_out_of_range ON observations USING btree (out_of_range);


--
-- Name: index_observations_on_photos_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_photos_count ON observations USING btree (observation_photos_count);


--
-- Name: index_observations_on_private_geom; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_private_geom ON observations USING gist (private_geom);


--
-- Name: index_observations_on_quality_grade; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_quality_grade ON observations USING btree (quality_grade);


--
-- Name: index_observations_on_site_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_site_id ON observations USING btree (site_id);


--
-- Name: index_observations_on_taxon_id_and_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_taxon_id_and_user_id ON observations USING btree (taxon_id, user_id);


--
-- Name: index_observations_on_uri; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_uri ON observations USING btree (uri);


--
-- Name: index_observations_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_user_id ON observations USING btree (user_id);


--
-- Name: index_observations_on_uuid; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_on_uuid ON observations USING btree (uuid);


--
-- Name: index_observations_places_on_observation_id_and_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_observations_places_on_observation_id_and_place_id ON observations_places USING btree (observation_id, place_id);


--
-- Name: index_observations_places_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_places_on_place_id ON observations_places USING btree (place_id);


--
-- Name: index_observations_posts_on_observation_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_posts_on_observation_id ON observations_posts USING btree (observation_id);


--
-- Name: index_observations_posts_on_post_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_posts_on_post_id ON observations_posts USING btree (post_id);


--
-- Name: index_observations_sounds_on_observation_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_sounds_on_observation_id ON observation_sounds USING btree (observation_id);


--
-- Name: index_observations_sounds_on_sound_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_sounds_on_sound_id ON observation_sounds USING btree (sound_id);


--
-- Name: index_observations_user_datetime; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_observations_user_datetime ON observations USING btree (user_id, observed_on, time_observed_at);


--
-- Name: index_photos_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_photos_on_user_id ON photos USING btree (user_id);


--
-- Name: index_picasa_identities_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_picasa_identities_on_user_id ON picasa_identities USING btree (user_id);


--
-- Name: index_place_geometries_on_geom; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_place_geometries_on_geom ON place_geometries USING gist (geom);


--
-- Name: index_place_geometries_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_place_geometries_on_place_id ON place_geometries USING btree (place_id);


--
-- Name: index_place_geometries_on_source_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_place_geometries_on_source_id ON place_geometries USING btree (source_id);


--
-- Name: index_place_taxon_names_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_place_taxon_names_on_place_id ON place_taxon_names USING btree (place_id);


--
-- Name: index_place_taxon_names_on_taxon_name_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_place_taxon_names_on_taxon_name_id ON place_taxon_names USING btree (taxon_name_id);


--
-- Name: index_places_on_admin_level; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_places_on_admin_level ON places USING btree (admin_level);


--
-- Name: index_places_on_ancestry; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_places_on_ancestry ON places USING btree (ancestry text_pattern_ops);


--
-- Name: index_places_on_bbox_area; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_places_on_bbox_area ON places USING btree (bbox_area);


--
-- Name: index_places_on_check_list_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_places_on_check_list_id ON places USING btree (check_list_id);


--
-- Name: index_places_on_latitude_and_longitude; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_places_on_latitude_and_longitude ON places USING btree (latitude, longitude);


--
-- Name: index_places_on_parent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_places_on_parent_id ON places USING btree (parent_id);


--
-- Name: index_places_on_place_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_places_on_place_type ON places USING btree (place_type);


--
-- Name: index_places_on_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_places_on_slug ON places USING btree (slug);


--
-- Name: index_places_on_source_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_places_on_source_id ON places USING btree (source_id);


--
-- Name: index_places_on_swlat_and_swlng_and_nelat_and_nelng; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_places_on_swlat_and_swlng_and_nelat_and_nelng ON places USING btree (swlat, swlng, nelat, nelng);


--
-- Name: index_places_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_places_on_user_id ON places USING btree (user_id);


--
-- Name: index_posts_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_posts_on_place_id ON posts USING btree (place_id);


--
-- Name: index_posts_on_published_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_posts_on_published_at ON posts USING btree (published_at);


--
-- Name: index_preferences_on_owner_and_name_and_preference; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_preferences_on_owner_and_name_and_preference ON preferences USING btree (owner_id, owner_type, name, group_id, group_type);


--
-- Name: index_project_assets_on_asset_content_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_project_assets_on_asset_content_type ON project_assets USING btree (asset_content_type);


--
-- Name: index_project_assets_on_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_project_assets_on_project_id ON project_assets USING btree (project_id);


--
-- Name: index_project_invitations_on_observation_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_project_invitations_on_observation_id ON project_invitations USING btree (observation_id);


--
-- Name: index_project_observation_fields_on_observation_field_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_project_observation_fields_on_observation_field_id ON project_observation_fields USING btree (observation_field_id);


--
-- Name: index_project_observations_on_curator_identification_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_project_observations_on_curator_identification_id ON project_observations USING btree (curator_identification_id);


--
-- Name: index_project_observations_on_observation_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_project_observations_on_observation_id ON project_observations USING btree (observation_id);


--
-- Name: index_project_observations_on_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_project_observations_on_project_id ON project_observations USING btree (project_id);


--
-- Name: index_project_observations_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_project_observations_on_user_id ON project_observations USING btree (user_id);


--
-- Name: index_project_observations_on_uuid; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_project_observations_on_uuid ON project_observations USING btree (uuid);


--
-- Name: index_project_user_invitations_on_invited_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_project_user_invitations_on_invited_user_id ON project_user_invitations USING btree (invited_user_id);


--
-- Name: index_project_user_invitations_on_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_project_user_invitations_on_project_id ON project_user_invitations USING btree (project_id);


--
-- Name: index_project_user_invitations_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_project_user_invitations_on_user_id ON project_user_invitations USING btree (user_id);


--
-- Name: index_project_users_on_project_id_and_taxa_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_project_users_on_project_id_and_taxa_count ON project_users USING btree (project_id, taxa_count);


--
-- Name: index_project_users_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_project_users_on_user_id ON project_users USING btree (user_id);


--
-- Name: index_projects_on_cached_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_projects_on_cached_slug ON projects USING btree (slug);


--
-- Name: index_projects_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_projects_on_place_id ON projects USING btree (place_id);


--
-- Name: index_projects_on_source_url; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_projects_on_source_url ON projects USING btree (source_url);


--
-- Name: index_projects_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_projects_on_user_id ON projects USING btree (user_id);


--
-- Name: index_provider_authorizations_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_provider_authorizations_on_user_id ON provider_authorizations USING btree (user_id);


--
-- Name: index_quality_metrics_on_observation_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_quality_metrics_on_observation_id ON quality_metrics USING btree (observation_id);


--
-- Name: index_quality_metrics_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_quality_metrics_on_user_id ON quality_metrics USING btree (user_id);


--
-- Name: index_roles_users_on_role_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_roles_users_on_role_id ON roles_users USING btree (role_id);


--
-- Name: index_roles_users_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_roles_users_on_user_id ON roles_users USING btree (user_id);


--
-- Name: index_site_admins_on_site_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_site_admins_on_site_id ON site_admins USING btree (site_id);


--
-- Name: index_site_admins_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_site_admins_on_user_id ON site_admins USING btree (user_id);


--
-- Name: index_sites_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_sites_on_place_id ON sites USING btree (place_id);


--
-- Name: index_sites_on_source_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_sites_on_source_id ON sites USING btree (source_id);


--
-- Name: index_slugs_on_n_s_s_and_s; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_slugs_on_n_s_s_and_s ON friendly_id_slugs USING btree (slug, sluggable_type, sequence, scope);


--
-- Name: index_slugs_on_sluggable_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_slugs_on_sluggable_id ON friendly_id_slugs USING btree (sluggable_id);


--
-- Name: index_sounds_on_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_sounds_on_type ON sounds USING btree (type);


--
-- Name: index_sounds_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_sounds_on_user_id ON sounds USING btree (user_id);


--
-- Name: index_sources_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_sources_on_user_id ON sources USING btree (user_id);


--
-- Name: index_states_simplified_1_on_geom; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_states_simplified_1_on_geom ON states_simplified_1 USING gist (geom);


--
-- Name: index_states_simplified_1_on_place_geometry_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_states_simplified_1_on_place_geometry_id ON states_simplified_1 USING btree (place_geometry_id);


--
-- Name: index_states_simplified_1_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_states_simplified_1_on_place_id ON states_simplified_1 USING btree (place_id);


--
-- Name: index_subscriptions_on_resource_type_and_resource_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_subscriptions_on_resource_type_and_resource_id ON subscriptions USING btree (resource_type, resource_id);


--
-- Name: index_subscriptions_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_subscriptions_on_taxon_id ON subscriptions USING btree (taxon_id);


--
-- Name: index_subscriptions_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_subscriptions_on_user_id ON subscriptions USING btree (user_id);


--
-- Name: index_taggings_on_taggable_id_and_taggable_type_and_context; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taggings_on_taggable_id_and_taggable_type_and_context ON taggings USING btree (taggable_id, taggable_type, context);


--
-- Name: index_tags_on_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_tags_on_name ON tags USING btree (name);


--
-- Name: index_taxa_on_ancestry; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxa_on_ancestry ON taxa USING btree (ancestry text_pattern_ops);


--
-- Name: index_taxa_on_conservation_status_source_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxa_on_conservation_status_source_id ON taxa USING btree (conservation_status_source_id);


--
-- Name: index_taxa_on_featured_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxa_on_featured_at ON taxa USING btree (featured_at);


--
-- Name: index_taxa_on_is_iconic; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxa_on_is_iconic ON taxa USING btree (is_iconic);


--
-- Name: index_taxa_on_listed_taxa_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxa_on_listed_taxa_count ON taxa USING btree (listed_taxa_count);


--
-- Name: index_taxa_on_locked; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxa_on_locked ON taxa USING btree (locked);


--
-- Name: index_taxa_on_lower_name_and_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxa_on_lower_name_and_id ON taxa USING btree (lower((name)::text), id);


--
-- Name: index_taxa_on_observations_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxa_on_observations_count ON taxa USING btree (observations_count);


--
-- Name: index_taxa_on_parent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxa_on_parent_id ON taxa USING btree (parent_id);


--
-- Name: index_taxa_on_rank_level; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxa_on_rank_level ON taxa USING btree (rank_level);


--
-- Name: index_taxa_on_unique_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxa_on_unique_name ON taxa USING btree (unique_name);


--
-- Name: index_taxon_ancestors_on_ancestor_taxon_id_and_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_taxon_ancestors_on_ancestor_taxon_id_and_taxon_id ON taxon_ancestors USING btree (ancestor_taxon_id, taxon_id);


--
-- Name: index_taxon_ancestors_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_ancestors_on_taxon_id ON taxon_ancestors USING btree (taxon_id);


--
-- Name: index_taxon_change_taxa_on_taxon_change_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_change_taxa_on_taxon_change_id ON taxon_change_taxa USING btree (taxon_change_id);


--
-- Name: index_taxon_change_taxa_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_change_taxa_on_taxon_id ON taxon_change_taxa USING btree (taxon_id);


--
-- Name: index_taxon_changes_on_committer_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_changes_on_committer_id ON taxon_changes USING btree (committer_id);


--
-- Name: index_taxon_changes_on_source_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_changes_on_source_id ON taxon_changes USING btree (source_id);


--
-- Name: index_taxon_changes_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_changes_on_taxon_id ON taxon_changes USING btree (taxon_id);


--
-- Name: index_taxon_changes_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_changes_on_user_id ON taxon_changes USING btree (user_id);


--
-- Name: index_taxon_descriptions_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_descriptions_on_taxon_id ON taxon_descriptions USING btree (taxon_id);


--
-- Name: index_taxon_links_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_links_on_place_id ON taxon_links USING btree (place_id);


--
-- Name: index_taxon_links_on_taxon_id_and_show_for_descendent_taxa; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_links_on_taxon_id_and_show_for_descendent_taxa ON taxon_links USING btree (taxon_id, show_for_descendent_taxa);


--
-- Name: index_taxon_links_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_links_on_user_id ON taxon_links USING btree (user_id);


--
-- Name: index_taxon_names_on_lexicon; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_names_on_lexicon ON taxon_names USING btree (lexicon);


--
-- Name: index_taxon_names_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_names_on_taxon_id ON taxon_names USING btree (taxon_id);


--
-- Name: index_taxon_photos_on_photo_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_photos_on_photo_id ON taxon_photos USING btree (photo_id);


--
-- Name: index_taxon_photos_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_photos_on_taxon_id ON taxon_photos USING btree (taxon_id);


--
-- Name: index_taxon_ranges_on_geom; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_ranges_on_geom ON taxon_ranges USING gist (geom);


--
-- Name: index_taxon_ranges_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_ranges_on_taxon_id ON taxon_ranges USING btree (taxon_id);


--
-- Name: index_taxon_scheme_taxa_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_scheme_taxa_on_taxon_id ON taxon_scheme_taxa USING btree (taxon_id);


--
-- Name: index_taxon_scheme_taxa_on_taxon_name_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_scheme_taxa_on_taxon_name_id ON taxon_scheme_taxa USING btree (taxon_name_id);


--
-- Name: index_taxon_scheme_taxa_on_taxon_scheme_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_scheme_taxa_on_taxon_scheme_id ON taxon_scheme_taxa USING btree (taxon_scheme_id);


--
-- Name: index_taxon_schemes_on_source_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taxon_schemes_on_source_id ON taxon_schemes USING btree (source_id);


--
-- Name: index_trip_purposes_on_resource_type_and_resource_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_trip_purposes_on_resource_type_and_resource_id ON trip_purposes USING btree (resource_type, resource_id);


--
-- Name: index_trip_purposes_on_trip_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_trip_purposes_on_trip_id ON trip_purposes USING btree (trip_id);


--
-- Name: index_trip_taxa_on_taxon_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_trip_taxa_on_taxon_id ON trip_taxa USING btree (taxon_id);


--
-- Name: index_trip_taxa_on_trip_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_trip_taxa_on_trip_id ON trip_taxa USING btree (trip_id);


--
-- Name: index_update_actions_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_update_actions_unique ON update_actions USING btree (resource_id, notifier_id, resource_type, notifier_type, notification, resource_owner_id);


--
-- Name: index_update_subscribers_on_update_action_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_update_subscribers_on_update_action_id ON update_subscribers USING btree (update_action_id);


--
-- Name: index_users_on_identifications_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_identifications_count ON users USING btree (identifications_count);


--
-- Name: index_users_on_journal_posts_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_journal_posts_count ON users USING btree (journal_posts_count);


--
-- Name: index_users_on_life_list_taxa_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_life_list_taxa_count ON users USING btree (life_list_taxa_count);


--
-- Name: index_users_on_login; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_users_on_login ON users USING btree (login);


--
-- Name: index_users_on_lower_login; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_lower_login ON users USING btree (lower((login)::text));


--
-- Name: index_users_on_observations_count; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_observations_count ON users USING btree (observations_count);


--
-- Name: index_users_on_place_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_place_id ON users USING btree (place_id);


--
-- Name: index_users_on_site_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_site_id ON users USING btree (site_id);


--
-- Name: index_users_on_spammer; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_spammer ON users USING btree (spammer);


--
-- Name: index_users_on_state; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_state ON users USING btree (state);


--
-- Name: index_users_on_uri; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_uri ON users USING btree (uri);


--
-- Name: index_votes_on_votable_id_and_votable_type_and_vote_scope; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_votes_on_votable_id_and_votable_type_and_vote_scope ON votes USING btree (votable_id, votable_type, vote_scope);


--
-- Name: index_votes_on_voter_id_and_voter_type_and_vote_scope; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_votes_on_voter_id_and_voter_type_and_vote_scope ON votes USING btree (voter_id, voter_type, vote_scope);


--
-- Name: index_wiki_page_attachments_on_page_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_wiki_page_attachments_on_page_id ON wiki_page_attachments USING btree (page_id);


--
-- Name: index_wiki_page_versions_on_page_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_wiki_page_versions_on_page_id ON wiki_page_versions USING btree (page_id);


--
-- Name: index_wiki_page_versions_on_updator_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_wiki_page_versions_on_updator_id ON wiki_page_versions USING btree (updator_id);


--
-- Name: index_wiki_pages_on_creator_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_wiki_pages_on_creator_id ON wiki_pages USING btree (creator_id);


--
-- Name: index_wiki_pages_on_path; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_wiki_pages_on_path ON wiki_pages USING btree (path);


--
-- Name: pof_projid_ofid; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX pof_projid_ofid ON project_observation_fields USING btree (project_id, observation_field_id);


--
-- Name: pof_projid_pos; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX pof_projid_pos ON project_observation_fields USING btree (project_id, "position");


--
-- Name: taggings_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX taggings_idx ON taggings USING btree (tag_id, taggable_id, taggable_type, context, tagger_id, tagger_type);


--
-- Name: taxon_names_lower_name_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX taxon_names_lower_name_index ON taxon_names USING btree (lower((name)::text));


--
-- Name: unique_schema_migrations; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX unique_schema_migrations ON schema_migrations USING btree (version);


--
-- PostgreSQL database dump complete
--

SET search_path TO "$user", public;

INSERT INTO schema_migrations (version) VALUES ('20090820033338');

INSERT INTO schema_migrations (version) VALUES ('20090920043428');

INSERT INTO schema_migrations (version) VALUES ('20091005055004');

INSERT INTO schema_migrations (version) VALUES ('20091023222943');

INSERT INTO schema_migrations (version) VALUES ('20091024022010');

INSERT INTO schema_migrations (version) VALUES ('20091123044434');

INSERT INTO schema_migrations (version) VALUES ('20091216052325');

INSERT INTO schema_migrations (version) VALUES ('20091221195909');

INSERT INTO schema_migrations (version) VALUES ('20091223030137');

INSERT INTO schema_migrations (version) VALUES ('20100119024356');

INSERT INTO schema_migrations (version) VALUES ('20100610052004');

INSERT INTO schema_migrations (version) VALUES ('20100709225557');

INSERT INTO schema_migrations (version) VALUES ('20100807184336');

INSERT INTO schema_migrations (version) VALUES ('20100807184524');

INSERT INTO schema_migrations (version) VALUES ('20100807184540');

INSERT INTO schema_migrations (version) VALUES ('20100815222147');

INSERT INTO schema_migrations (version) VALUES ('20101002052112');

INSERT INTO schema_migrations (version) VALUES ('20101010224648');

INSERT INTO schema_migrations (version) VALUES ('20101017010641');

INSERT INTO schema_migrations (version) VALUES ('20101120231112');

INSERT INTO schema_migrations (version) VALUES ('20101128052201');

INSERT INTO schema_migrations (version) VALUES ('20101203223538');

INSERT INTO schema_migrations (version) VALUES ('20101218044932');

INSERT INTO schema_migrations (version) VALUES ('20101226171854');

INSERT INTO schema_migrations (version) VALUES ('20110107064406');

INSERT INTO schema_migrations (version) VALUES ('20110112061527');

INSERT INTO schema_migrations (version) VALUES ('20110202063613');

INSERT INTO schema_migrations (version) VALUES ('20110228043741');

INSERT INTO schema_migrations (version) VALUES ('20110316040303');

INSERT INTO schema_migrations (version) VALUES ('20110326195224');

INSERT INTO schema_migrations (version) VALUES ('20110330050657');

INSERT INTO schema_migrations (version) VALUES ('20110331173629');

INSERT INTO schema_migrations (version) VALUES ('20110331174611');

INSERT INTO schema_migrations (version) VALUES ('20110401221815');

INSERT INTO schema_migrations (version) VALUES ('20110402222428');

INSERT INTO schema_migrations (version) VALUES ('20110405041648');

INSERT INTO schema_migrations (version) VALUES ('20110405041654');

INSERT INTO schema_migrations (version) VALUES ('20110405041659');

INSERT INTO schema_migrations (version) VALUES ('20110408005124');

INSERT INTO schema_migrations (version) VALUES ('20110409064704');

INSERT INTO schema_migrations (version) VALUES ('20110414202308');

INSERT INTO schema_migrations (version) VALUES ('20110415221429');

INSERT INTO schema_migrations (version) VALUES ('20110415225622');

INSERT INTO schema_migrations (version) VALUES ('20110415230149');

INSERT INTO schema_migrations (version) VALUES ('20110428074115');

INSERT INTO schema_migrations (version) VALUES ('20110429004856');

INSERT INTO schema_migrations (version) VALUES ('20110429075345');

INSERT INTO schema_migrations (version) VALUES ('20110502182056');

INSERT INTO schema_migrations (version) VALUES ('20110502221926');

INSERT INTO schema_migrations (version) VALUES ('20110505040504');

INSERT INTO schema_migrations (version) VALUES ('20110513230256');

INSERT INTO schema_migrations (version) VALUES ('20110514221925');

INSERT INTO schema_migrations (version) VALUES ('20110526205447');

INSERT INTO schema_migrations (version) VALUES ('20110529052159');

INSERT INTO schema_migrations (version) VALUES ('20110531065431');

INSERT INTO schema_migrations (version) VALUES ('20110610193807');

INSERT INTO schema_migrations (version) VALUES ('20110709200352');

INSERT INTO schema_migrations (version) VALUES ('20110714185244');

INSERT INTO schema_migrations (version) VALUES ('20110731201217');

INSERT INTO schema_migrations (version) VALUES ('20110801001844');

INSERT INTO schema_migrations (version) VALUES ('20110805044702');

INSERT INTO schema_migrations (version) VALUES ('20110807035642');

INSERT INTO schema_migrations (version) VALUES ('20110809064402');

INSERT INTO schema_migrations (version) VALUES ('20110809064437');

INSERT INTO schema_migrations (version) VALUES ('20110811040139');

INSERT INTO schema_migrations (version) VALUES ('20110905185019');

INSERT INTO schema_migrations (version) VALUES ('20110913060143');

INSERT INTO schema_migrations (version) VALUES ('20111003210305');

INSERT INTO schema_migrations (version) VALUES ('20111014181723');

INSERT INTO schema_migrations (version) VALUES ('20111014182046');

INSERT INTO schema_migrations (version) VALUES ('20111027041911');

INSERT INTO schema_migrations (version) VALUES ('20111027211849');

INSERT INTO schema_migrations (version) VALUES ('20111028190803');

INSERT INTO schema_migrations (version) VALUES ('20111102210429');

INSERT INTO schema_migrations (version) VALUES ('20111108184751');

INSERT INTO schema_migrations (version) VALUES ('20111202065742');

INSERT INTO schema_migrations (version) VALUES ('20111209033826');

INSERT INTO schema_migrations (version) VALUES ('20111212052205');

INSERT INTO schema_migrations (version) VALUES ('20111226210945');

INSERT INTO schema_migrations (version) VALUES ('20120102213824');

INSERT INTO schema_migrations (version) VALUES ('20120105232343');

INSERT INTO schema_migrations (version) VALUES ('20120106222437');

INSERT INTO schema_migrations (version) VALUES ('20120109221839');

INSERT INTO schema_migrations (version) VALUES ('20120109221956');

INSERT INTO schema_migrations (version) VALUES ('20120119183954');

INSERT INTO schema_migrations (version) VALUES ('20120119184143');

INSERT INTO schema_migrations (version) VALUES ('20120120232035');

INSERT INTO schema_migrations (version) VALUES ('20120123001206');

INSERT INTO schema_migrations (version) VALUES ('20120123190202');

INSERT INTO schema_migrations (version) VALUES ('20120214200727');

INSERT INTO schema_migrations (version) VALUES ('20120413012920');

INSERT INTO schema_migrations (version) VALUES ('20120413013521');

INSERT INTO schema_migrations (version) VALUES ('20120416221933');

INSERT INTO schema_migrations (version) VALUES ('20120425042326');

INSERT INTO schema_migrations (version) VALUES ('20120427014202');

INSERT INTO schema_migrations (version) VALUES ('20120504214431');

INSERT INTO schema_migrations (version) VALUES ('20120521225005');

INSERT INTO schema_migrations (version) VALUES ('20120524173746');

INSERT INTO schema_migrations (version) VALUES ('20120525190526');

INSERT INTO schema_migrations (version) VALUES ('20120529181631');

INSERT INTO schema_migrations (version) VALUES ('20120609003704');

INSERT INTO schema_migrations (version) VALUES ('20120628014940');

INSERT INTO schema_migrations (version) VALUES ('20120628014948');

INSERT INTO schema_migrations (version) VALUES ('20120628015126');

INSERT INTO schema_migrations (version) VALUES ('20120629011843');

INSERT INTO schema_migrations (version) VALUES ('20120702194230');

INSERT INTO schema_migrations (version) VALUES ('20120702224519');

INSERT INTO schema_migrations (version) VALUES ('20120704055118');

INSERT INTO schema_migrations (version) VALUES ('20120711053525');

INSERT INTO schema_migrations (version) VALUES ('20120711053620');

INSERT INTO schema_migrations (version) VALUES ('20120712040410');

INSERT INTO schema_migrations (version) VALUES ('20120713074557');

INSERT INTO schema_migrations (version) VALUES ('20120717184355');

INSERT INTO schema_migrations (version) VALUES ('20120719171324');

INSERT INTO schema_migrations (version) VALUES ('20120725194234');

INSERT INTO schema_migrations (version) VALUES ('20120801204921');

INSERT INTO schema_migrations (version) VALUES ('20120808224842');

INSERT INTO schema_migrations (version) VALUES ('20120810053551');

INSERT INTO schema_migrations (version) VALUES ('20120821195023');

INSERT INTO schema_migrations (version) VALUES ('20120830020828');

INSERT INTO schema_migrations (version) VALUES ('20120902210558');

INSERT INTO schema_migrations (version) VALUES ('20120904064231');

INSERT INTO schema_migrations (version) VALUES ('20120906014934');

INSERT INTO schema_migrations (version) VALUES ('20120919201617');

INSERT INTO schema_migrations (version) VALUES ('20120926220539');

INSERT INTO schema_migrations (version) VALUES ('20120929003044');

INSERT INTO schema_migrations (version) VALUES ('20121011181051');

INSERT INTO schema_migrations (version) VALUES ('20121031200130');

INSERT INTO schema_migrations (version) VALUES ('20121101180101');

INSERT INTO schema_migrations (version) VALUES ('20121115043256');

INSERT INTO schema_migrations (version) VALUES ('20121116214553');

INSERT INTO schema_migrations (version) VALUES ('20121119073505');

INSERT INTO schema_migrations (version) VALUES ('20121128022641');

INSERT INTO schema_migrations (version) VALUES ('20121224231303');

INSERT INTO schema_migrations (version) VALUES ('20121227214513');

INSERT INTO schema_migrations (version) VALUES ('20121230023106');

INSERT INTO schema_migrations (version) VALUES ('20121230210148');

INSERT INTO schema_migrations (version) VALUES ('20130102225500');

INSERT INTO schema_migrations (version) VALUES ('20130103065755');

INSERT INTO schema_migrations (version) VALUES ('20130108182219');

INSERT INTO schema_migrations (version) VALUES ('20130108182802');

INSERT INTO schema_migrations (version) VALUES ('20130116165914');

INSERT INTO schema_migrations (version) VALUES ('20130116225224');

INSERT INTO schema_migrations (version) VALUES ('20130131001533');

INSERT INTO schema_migrations (version) VALUES ('20130131061500');

INSERT INTO schema_migrations (version) VALUES ('20130201224839');

INSERT INTO schema_migrations (version) VALUES ('20130205052838');

INSERT INTO schema_migrations (version) VALUES ('20130206192217');

INSERT INTO schema_migrations (version) VALUES ('20130208003925');

INSERT INTO schema_migrations (version) VALUES ('20130208222855');

INSERT INTO schema_migrations (version) VALUES ('20130226064319');

INSERT INTO schema_migrations (version) VALUES ('20130227211137');

INSERT INTO schema_migrations (version) VALUES ('20130301222959');

INSERT INTO schema_migrations (version) VALUES ('20130304024311');

INSERT INTO schema_migrations (version) VALUES ('20130306020925');

INSERT INTO schema_migrations (version) VALUES ('20130311061913');

INSERT INTO schema_migrations (version) VALUES ('20130312070047');

INSERT INTO schema_migrations (version) VALUES ('20130313192420');

INSERT INTO schema_migrations (version) VALUES ('20130403235431');

INSERT INTO schema_migrations (version) VALUES ('20130409225631');

INSERT INTO schema_migrations (version) VALUES ('20130411225629');

INSERT INTO schema_migrations (version) VALUES ('20130418190210');

INSERT INTO schema_migrations (version) VALUES ('20130429215442');

INSERT INTO schema_migrations (version) VALUES ('20130501005855');

INSERT INTO schema_migrations (version) VALUES ('20130502190619');

INSERT INTO schema_migrations (version) VALUES ('20130514012017');

INSERT INTO schema_migrations (version) VALUES ('20130514012037');

INSERT INTO schema_migrations (version) VALUES ('20130514012051');

INSERT INTO schema_migrations (version) VALUES ('20130514012105');

INSERT INTO schema_migrations (version) VALUES ('20130514012120');

INSERT INTO schema_migrations (version) VALUES ('20130516200016');

INSERT INTO schema_migrations (version) VALUES ('20130521001431');

INSERT INTO schema_migrations (version) VALUES ('20130523203022');

INSERT INTO schema_migrations (version) VALUES ('20130603221737');

INSERT INTO schema_migrations (version) VALUES ('20130603234330');

INSERT INTO schema_migrations (version) VALUES ('20130604012213');

INSERT INTO schema_migrations (version) VALUES ('20130607221500');

INSERT INTO schema_migrations (version) VALUES ('20130611025612');

INSERT INTO schema_migrations (version) VALUES ('20130613223707');

INSERT INTO schema_migrations (version) VALUES ('20130624022309');

INSERT INTO schema_migrations (version) VALUES ('20130628035929');

INSERT INTO schema_migrations (version) VALUES ('20130701224024');

INSERT INTO schema_migrations (version) VALUES ('20130704010119');

INSERT INTO schema_migrations (version) VALUES ('20130708233246');

INSERT INTO schema_migrations (version) VALUES ('20130708235548');

INSERT INTO schema_migrations (version) VALUES ('20130709005451');

INSERT INTO schema_migrations (version) VALUES ('20130709212550');

INSERT INTO schema_migrations (version) VALUES ('20130711181857');

INSERT INTO schema_migrations (version) VALUES ('20130721235136');

INSERT INTO schema_migrations (version) VALUES ('20130730200246');

INSERT INTO schema_migrations (version) VALUES ('20130814211257');

INSERT INTO schema_migrations (version) VALUES ('20130903235202');

INSERT INTO schema_migrations (version) VALUES ('20130910053330');

INSERT INTO schema_migrations (version) VALUES ('20130917071826');

INSERT INTO schema_migrations (version) VALUES ('20130926224132');

INSERT INTO schema_migrations (version) VALUES ('20130926233023');

INSERT INTO schema_migrations (version) VALUES ('20130929024857');

INSERT INTO schema_migrations (version) VALUES ('20131008061545');

INSERT INTO schema_migrations (version) VALUES ('20131011234030');

INSERT INTO schema_migrations (version) VALUES ('20131023224910');

INSERT INTO schema_migrations (version) VALUES ('20131024045916');

INSERT INTO schema_migrations (version) VALUES ('20131031160647');

INSERT INTO schema_migrations (version) VALUES ('20131031171349');

INSERT INTO schema_migrations (version) VALUES ('20131119214722');

INSERT INTO schema_migrations (version) VALUES ('20131123022658');

INSERT INTO schema_migrations (version) VALUES ('20131128214012');

INSERT INTO schema_migrations (version) VALUES ('20131128234236');

INSERT INTO schema_migrations (version) VALUES ('20131204211450');

INSERT INTO schema_migrations (version) VALUES ('20131220044313');

INSERT INTO schema_migrations (version) VALUES ('20140101210916');

INSERT INTO schema_migrations (version) VALUES ('20140104202529');

INSERT INTO schema_migrations (version) VALUES ('20140113145150');

INSERT INTO schema_migrations (version) VALUES ('20140114210551');

INSERT INTO schema_migrations (version) VALUES ('20140124190652');

INSERT INTO schema_migrations (version) VALUES ('20140205200914');

INSERT INTO schema_migrations (version) VALUES ('20140220201532');

INSERT INTO schema_migrations (version) VALUES ('20140225074921');

INSERT INTO schema_migrations (version) VALUES ('20140307003642');

INSERT INTO schema_migrations (version) VALUES ('20140313030123');

INSERT INTO schema_migrations (version) VALUES ('20140416193430');

INSERT INTO schema_migrations (version) VALUES ('20140604055610');

INSERT INTO schema_migrations (version) VALUES ('20140611180054');

INSERT INTO schema_migrations (version) VALUES ('20140620021223');

INSERT INTO schema_migrations (version) VALUES ('20140701212522');

INSERT INTO schema_migrations (version) VALUES ('20140704062909');

INSERT INTO schema_migrations (version) VALUES ('20140731201815');

INSERT INTO schema_migrations (version) VALUES ('20140820152353');

INSERT INTO schema_migrations (version) VALUES ('20140904004901');

INSERT INTO schema_migrations (version) VALUES ('20140912201349');

INSERT INTO schema_migrations (version) VALUES ('20141003193707');

INSERT INTO schema_migrations (version) VALUES ('20141015212020');

INSERT INTO schema_migrations (version) VALUES ('20141015213053');

INSERT INTO schema_migrations (version) VALUES ('20141112011137');

INSERT INTO schema_migrations (version) VALUES ('20141201211037');

INSERT INTO schema_migrations (version) VALUES ('20141203024242');

INSERT INTO schema_migrations (version) VALUES ('20141204224856');

INSERT INTO schema_migrations (version) VALUES ('20141213001622');

INSERT INTO schema_migrations (version) VALUES ('20141213195804');

INSERT INTO schema_migrations (version) VALUES ('20141229185357');

INSERT INTO schema_migrations (version) VALUES ('20141231210447');

INSERT INTO schema_migrations (version) VALUES ('20150104021132');

INSERT INTO schema_migrations (version) VALUES ('20150104033219');

INSERT INTO schema_migrations (version) VALUES ('20150126194129');

INSERT INTO schema_migrations (version) VALUES ('20150128225554');

INSERT INTO schema_migrations (version) VALUES ('20150203174741');

INSERT INTO schema_migrations (version) VALUES ('20150226010539');

INSERT INTO schema_migrations (version) VALUES ('20150304201738');

INSERT INTO schema_migrations (version) VALUES ('20150313171312');

INSERT INTO schema_migrations (version) VALUES ('20150319205049');

INSERT INTO schema_migrations (version) VALUES ('20150324004401');

INSERT INTO schema_migrations (version) VALUES ('20150404012836');

INSERT INTO schema_migrations (version) VALUES ('20150406181841');

INSERT INTO schema_migrations (version) VALUES ('20150409021334');

INSERT INTO schema_migrations (version) VALUES ('20150409031504');

INSERT INTO schema_migrations (version) VALUES ('20150412200608');

INSERT INTO schema_migrations (version) VALUES ('20150413222254');

INSERT INTO schema_migrations (version) VALUES ('20150421155510');

INSERT INTO schema_migrations (version) VALUES ('20150504184529');

INSERT INTO schema_migrations (version) VALUES ('20150509225733');

INSERT INTO schema_migrations (version) VALUES ('20150512222753');

INSERT INTO schema_migrations (version) VALUES ('20150524000620');

INSERT INTO schema_migrations (version) VALUES ('20150611215738');

INSERT INTO schema_migrations (version) VALUES ('20150614212053');

INSERT INTO schema_migrations (version) VALUES ('20150619231829');

INSERT INTO schema_migrations (version) VALUES ('20150622201252');

INSERT INTO schema_migrations (version) VALUES ('20150625230227');

INSERT INTO schema_migrations (version) VALUES ('20150701222736');

INSERT INTO schema_migrations (version) VALUES ('20150902052821');

INSERT INTO schema_migrations (version) VALUES ('20150916164339');

INSERT INTO schema_migrations (version) VALUES ('20150922154000');

INSERT INTO schema_migrations (version) VALUES ('20150922215548');

INSERT INTO schema_migrations (version) VALUES ('20151006230511');

INSERT INTO schema_migrations (version) VALUES ('20151014213826');

INSERT INTO schema_migrations (version) VALUES ('20151026184104');

INSERT INTO schema_migrations (version) VALUES ('20151030205931');

INSERT INTO schema_migrations (version) VALUES ('20151104175231');

INSERT INTO schema_migrations (version) VALUES ('20151117005737');

INSERT INTO schema_migrations (version) VALUES ('20151117221028');

INSERT INTO schema_migrations (version) VALUES ('20151228144302');

INSERT INTO schema_migrations (version) VALUES ('20160104200015');

INSERT INTO schema_migrations (version) VALUES ('20160317211729');

INSERT INTO schema_migrations (version) VALUES ('20160323182801');

INSERT INTO schema_migrations (version) VALUES ('20160324184344');

INSERT INTO schema_migrations (version) VALUES ('20160325152944');

INSERT INTO schema_migrations (version) VALUES ('20160406233849');

INSERT INTO schema_migrations (version) VALUES ('20160531181652');

INSERT INTO schema_migrations (version) VALUES ('20160531215755');

INSERT INTO schema_migrations (version) VALUES ('20160611140606');

INSERT INTO schema_migrations (version) VALUES ('20160613200151');

INSERT INTO schema_migrations (version) VALUES ('20160613202854');

INSERT INTO schema_migrations (version) VALUES ('20160624205645');

INSERT INTO schema_migrations (version) VALUES ('20160627194031');

INSERT INTO schema_migrations (version) VALUES ('20160629221454');

INSERT INTO schema_migrations (version) VALUES ('20160630024035');

INSERT INTO schema_migrations (version) VALUES ('20160701031842');

INSERT INTO schema_migrations (version) VALUES ('20160701042751');

INSERT INTO schema_migrations (version) VALUES ('20160726191620');

INSERT INTO schema_migrations (version) VALUES ('20160808154245');

INSERT INTO schema_migrations (version) VALUES ('20160809221731');

INSERT INTO schema_migrations (version) VALUES ('20160809221754');

INSERT INTO schema_migrations (version) VALUES ('20160815154039');

INSERT INTO schema_migrations (version) VALUES ('20160818234437');

INSERT INTO schema_migrations (version) VALUES ('20160913224325');

INSERT INTO schema_migrations (version) VALUES ('20160920151846');

INSERT INTO schema_migrations (version) VALUES ('20160929155608');

INSERT INTO schema_migrations (version) VALUES ('20161012202458');

INSERT INTO schema_migrations (version) VALUES ('20161012202803');

INSERT INTO schema_migrations (version) VALUES ('20161012204604');

INSERT INTO schema_migrations (version) VALUES ('20161020190217');

INSERT INTO schema_migrations (version) VALUES ('20161110221032');

INSERT INTO schema_migrations (version) VALUES ('20161210081605');

INSERT INTO schema_migrations (version) VALUES ('20161216041939');

INSERT INTO schema_migrations (version) VALUES ('20161220213126');

INSERT INTO schema_migrations (version) VALUES ('20170110025430');

INSERT INTO schema_migrations (version) VALUES ('20170110025450');

INSERT INTO schema_migrations (version) VALUES ('20170110185648');

INSERT INTO schema_migrations (version) VALUES ('20170113211950');

INSERT INTO schema_migrations (version) VALUES ('20170309003500');

INSERT INTO schema_migrations (version) VALUES ('20170317183900');

