-- Downloaded from: https://github.com/DTOcean/dtocean-database/blob/51973ccac29ba7262826d5a81ddad84170e9e1cb/schema.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 10.5
-- Dumped by pg_dump version 10.5

-- Started on 2019-03-12 11:14:06

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 6 (class 2615 OID 30909)
-- Name: filter; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA filter;


--
-- TOC entry 11 (class 2615 OID 30910)
-- Name: project; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA project;


--
-- TOC entry 12 (class 2615 OID 30911)
-- Name: reference; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA reference;


--
-- TOC entry 1 (class 3079 OID 12924)
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- TOC entry 5578 (class 0 OID 0)
-- Dependencies: 1
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- TOC entry 3 (class 3079 OID 29364)
-- Name: fuzzystrmatch; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS fuzzystrmatch WITH SCHEMA public;


--
-- TOC entry 5579 (class 0 OID 0)
-- Dependencies: 3
-- Name: EXTENSION fuzzystrmatch; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION fuzzystrmatch IS 'determine similarities and distance between strings';


--
-- TOC entry 2 (class 3079 OID 29375)
-- Name: postgis; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS postgis WITH SCHEMA public;


--
-- TOC entry 5580 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION postgis; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION postgis IS 'PostGIS geometry, geography, and raster spatial types and functions';


--
-- TOC entry 1674 (class 1255 OID 30912)
-- Name: sp_build_tables(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_tables() RETURNS void
    LANGUAGE plpgsql
    AS $$

DECLARE
   arr varchar[] := array['bathymetry',
			  'bathymetry_layer',
			  'cable_corridor_bathymetry',
			  'cable_corridor_bathymetry_layer',
			  'cable_corridor_constraint',
			  'constraint',
			  'device_shared',
			  'device_floating',
			  'device_tidal',
			  'device_tidal_power_performance',
			  'device_wave',
			  'lease_area',
			  'sub_systems_access',
			  'sub_systems_economic',
			  'sub_systems_install',
			  'sub_systems_inspection',
			  'sub_systems_maintenance',
			  'sub_systems_operation_weightings',
			  'sub_systems_replace',
			  'time_series_energy_tidal',
			  'time_series_energy_wave',
			  'time_series_om_tidal',
			  'time_series_om_wave',
			  'time_series_om_wind'
			  ];
   y TEXT;
   x TEXT;
   r RECORD;
BEGIN
   FOREACH x IN ARRAY arr
   LOOP
      y := 'DROP TABLE filter.' || x;
      RAISE NOTICE '%', y;
      BEGIN
        EXECUTE y;
      EXCEPTION
        WHEN SQLSTATE '42P01' THEN NULL;
      END;
      y := 'CREATE TABLE filter.' || x || ' (LIKE project.' || x || ')';
      RAISE NOTICE '%', y;
      EXECUTE y;
   END LOOP;
END

$$;


--
-- TOC entry 1675 (class 1255 OID 30913)
-- Name: sp_build_view_bathymetry_layer(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_view_bathymetry_layer() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW filter.view_bathymetry_layer AS 
 SELECT 
    bathymetry.utm_point,
    bathymetry.depth,
    bathymetry.mannings_no,
    bathymetry_layer.layer_order,
    bathymetry_layer.initial_depth,
    soil_type.description AS sediment_type
   FROM filter.bathymetry
     LEFT JOIN filter.bathymetry_layer
         ON bathymetry.id = bathymetry_layer.fk_bathymetry_id
     LEFT JOIN reference.soil_type
         ON bathymetry_layer.fk_soil_type_id = soil_type.id;
$$;


--
-- TOC entry 1676 (class 1255 OID 30914)
-- Name: sp_build_view_cable_corridor_bathymetry_layer(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_view_cable_corridor_bathymetry_layer() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW filter.view_cable_corridor_bathymetry_layer AS 
 SELECT cable_corridor_bathymetry.utm_point,
    cable_corridor_bathymetry.depth,
    cable_corridor_bathymetry_layer.layer_order,
    cable_corridor_bathymetry_layer.initial_depth,
    soil_type.description AS sediment_type
   FROM filter.cable_corridor_bathymetry
     LEFT JOIN filter.cable_corridor_bathymetry_layer
         ON cable_corridor_bathymetry.id = cable_corridor_bathymetry_layer.fk_bathymetry_id
     LEFT JOIN reference.soil_type
         ON cable_corridor_bathymetry_layer.fk_soil_type_id = soil_type.id;
$$;


--
-- TOC entry 1677 (class 1255 OID 30915)
-- Name: sp_build_view_control_system_access(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_view_control_system_access() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW filter.view_control_system_access AS 
  SELECT 
    sub_system,
    operation_duration,
    max_hs,
    max_tp,
    max_ws,
    max_cs
  FROM
    "filter"."sub_systems_access" INNER JOIN "project"."sub_systems"
    ON ("filter"."sub_systems_access"."fk_sub_system_id" = "project"."sub_systems"."id")
  WHERE sub_system in ('Control System');
$$;


--
-- TOC entry 1678 (class 1255 OID 30916)
-- Name: sp_build_view_control_system_economic(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_view_control_system_economic() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW filter.view_control_system_economic AS 
  SELECT 
    sub_system,
    cost,
    failure_rate
  FROM
    "filter"."sub_systems_economic" INNER JOIN "project"."sub_systems"
    ON ("filter"."sub_systems_economic"."fk_sub_system_id" = "project"."sub_systems"."id")
  WHERE sub_system in ('Control System');
$$;


--
-- TOC entry 1679 (class 1255 OID 30917)
-- Name: sp_build_view_control_system_inspection(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_view_control_system_inspection() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW filter.view_control_system_inspection AS 
  SELECT 
    sub_system,
    sub_systems_inspection.operation_duration,
    sub_systems_inspection.crew_lead_time,
    sub_systems_inspection.other_lead_time,
    sub_systems_inspection.n_specialists,
    sub_systems_inspection.n_technicians,
    sub_systems_inspection.max_hs,
    sub_systems_inspection.max_tp,
    sub_systems_inspection.max_ws,
    sub_systems_inspection.max_cs
  FROM
    filter.sub_systems_inspection INNER JOIN "project"."sub_systems"
    ON ("filter"."sub_systems_inspection"."fk_sub_system_id" = "project"."sub_systems"."id")
  WHERE sub_system in ('Control System');
$$;


--
-- TOC entry 1680 (class 1255 OID 30918)
-- Name: sp_build_view_control_system_installation(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_view_control_system_installation() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW filter.view_control_system_installation AS 
  SELECT 
    sub_system,
    length,
    width,
    height,
    dry_mass,
    max_hs,
    max_tp,
    max_ws,
    max_cs
  FROM
    filter.sub_systems_install INNER JOIN "project"."sub_systems"
    ON ("filter"."sub_systems_install"."fk_sub_system_id" = "project"."sub_systems"."id")
    WHERE sub_system in ('Control System');
$$;


--
-- TOC entry 1681 (class 1255 OID 30919)
-- Name: sp_build_view_control_system_maintenance(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_view_control_system_maintenance() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW filter.view_control_system_maintenance AS 
  SELECT
    sub_system,
    operation_duration,
    interruptible,
    parts_length,
    parts_width,
    parts_height,
    parts_dry_mass,
    assembly_lead_time,
    crew_lead_time,
    other_lead_time,
    n_specialists,
    n_technicians,
    max_hs,
    max_tp,
    max_ws,
    max_cs
  FROM
    filter.sub_systems_maintenance INNER JOIN "project"."sub_systems"
    ON ("filter"."sub_systems_maintenance"."fk_sub_system_id" = "project"."sub_systems"."id")
  WHERE sub_system in ('Control System');
$$;


--
-- TOC entry 1682 (class 1255 OID 30920)
-- Name: sp_build_view_control_system_operation_weightings(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_view_control_system_operation_weightings() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW filter.view_control_system_operation_weightings AS 
  SELECT
    sub_system,
    maintenance,
    replacement,
    inspection
  FROM
    filter.sub_systems_operation_weightings INNER JOIN "project"."sub_systems"
    ON ("filter"."sub_systems_operation_weightings"."fk_sub_system_id" = "project"."sub_systems"."id")
  WHERE sub_system in ('Control System');
$$;


--
-- TOC entry 1654 (class 1255 OID 30921)
-- Name: sp_build_view_control_system_replace(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_view_control_system_replace() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW filter.view_control_system_replace AS 
  SELECT
    sub_system,
    operation_duration,
    interruptible,
    assembly_lead_time,
    crew_lead_time,
    other_lead_time,
    n_specialists,
    n_technicians
  FROM
    filter.sub_systems_replace INNER JOIN "project"."sub_systems"
    ON ("filter"."sub_systems_replace"."fk_sub_system_id" = "project"."sub_systems"."id")
  WHERE sub_system in ('Control System');
$$;


--
-- TOC entry 1663 (class 1255 OID 30922)
-- Name: sp_build_view_sub_systems_access(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_view_sub_systems_access() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW filter.view_sub_systems_access AS 
  SELECT 
    sub_system,
    operation_duration,
    max_hs,
    max_tp,
    max_ws,
    max_cs
  FROM
    "filter"."sub_systems_access" INNER JOIN "project"."sub_systems"
    ON ("filter"."sub_systems_access"."fk_sub_system_id" = "project"."sub_systems"."id")
  WHERE sub_system in ('Prime Mover', 'PTO', 'Support Structure');
$$;


--
-- TOC entry 1671 (class 1255 OID 30923)
-- Name: sp_build_view_sub_systems_economic(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_view_sub_systems_economic() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW filter.view_sub_systems_economic AS 
  SELECT 
    sub_system,
    cost,
    failure_rate
  FROM
    "filter"."sub_systems_economic" INNER JOIN "project"."sub_systems"
    ON ("filter"."sub_systems_economic"."fk_sub_system_id" = "project"."sub_systems"."id")
  WHERE sub_system in ('Prime Mover', 'PTO', 'Support Structure');
$$;


--
-- TOC entry 1683 (class 1255 OID 30924)
-- Name: sp_build_view_sub_systems_inspection(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_view_sub_systems_inspection() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW filter.view_sub_systems_inspection AS 
  SELECT 
    sub_system,
    sub_systems_inspection.operation_duration,
    sub_systems_inspection.crew_lead_time,
    sub_systems_inspection.other_lead_time,
    sub_systems_inspection.n_specialists,
    sub_systems_inspection.n_technicians,
    sub_systems_inspection.max_hs,
    sub_systems_inspection.max_tp,
    sub_systems_inspection.max_ws,
    sub_systems_inspection.max_cs
  FROM
    filter.sub_systems_inspection INNER JOIN "project"."sub_systems"
    ON ("filter"."sub_systems_inspection"."fk_sub_system_id" = "project"."sub_systems"."id")
  WHERE sub_system in ('Prime Mover', 'PTO', 'Support Structure');
$$;


--
-- TOC entry 1684 (class 1255 OID 30925)
-- Name: sp_build_view_sub_systems_installation(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_view_sub_systems_installation() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW filter.view_sub_systems_installation AS 
  SELECT 
    sub_system,
    length,
    width,
    height,
    dry_mass,
    max_hs,
    max_tp,
    max_ws,
    max_cs
  FROM
    filter.sub_systems_install INNER JOIN "project"."sub_systems"
    ON ("filter"."sub_systems_install"."fk_sub_system_id" = "project"."sub_systems"."id")
    WHERE sub_system in ('Prime Mover', 'PTO', 'Support Structure');
$$;


--
-- TOC entry 1685 (class 1255 OID 30926)
-- Name: sp_build_view_sub_systems_maintenance(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_view_sub_systems_maintenance() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW filter.view_sub_systems_maintenance AS 
  SELECT
    sub_system,
    operation_duration,
    interruptible,
    parts_length,
    parts_width,
    parts_height,
    parts_dry_mass,
    assembly_lead_time,
    crew_lead_time,
    other_lead_time,
    n_specialists,
    n_technicians,
    max_hs,
    max_tp,
    max_ws,
    max_cs
  FROM
    filter.sub_systems_maintenance INNER JOIN "project"."sub_systems"
    ON ("filter"."sub_systems_maintenance"."fk_sub_system_id" = "project"."sub_systems"."id")
  WHERE sub_system in ('Prime Mover', 'PTO', 'Support Structure');
$$;


--
-- TOC entry 1686 (class 1255 OID 30927)
-- Name: sp_build_view_sub_systems_operation_weightings(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_view_sub_systems_operation_weightings() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW filter.view_sub_systems_operation_weightings AS 
  SELECT
    sub_system,
    maintenance,
    replacement,
    inspection
  FROM
    filter.sub_systems_operation_weightings INNER JOIN "project"."sub_systems"
    ON ("filter"."sub_systems_operation_weightings"."fk_sub_system_id" = "project"."sub_systems"."id")
  WHERE sub_system in ('Prime Mover', 'PTO', 'Support Structure');
$$;


--
-- TOC entry 1687 (class 1255 OID 30928)
-- Name: sp_build_view_sub_systems_replace(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_view_sub_systems_replace() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW filter.view_sub_systems_replace AS 
  SELECT
    sub_system,
    operation_duration,
    interruptible,
    assembly_lead_time,
    crew_lead_time,
    other_lead_time,
    n_specialists,
    n_technicians
  FROM
    filter.sub_systems_replace INNER JOIN "project"."sub_systems"
    ON ("filter"."sub_systems_replace"."fk_sub_system_id" = "project"."sub_systems"."id")
  WHERE sub_system in ('Prime Mover', 'PTO', 'Support Structure');
$$;


--
-- TOC entry 1688 (class 1255 OID 30929)
-- Name: sp_build_view_time_series_energy_tidal(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_view_time_series_energy_tidal() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW filter.view_time_series_energy_tidal AS 
 SELECT bathymetry.utm_point,
    time_series_energy_tidal.measure_date,
    time_series_energy_tidal.measure_time,
    time_series_energy_tidal.u,
    time_series_energy_tidal.v,
    time_series_energy_tidal.turbulence_intensity,
    time_series_energy_tidal.ssh
   FROM filter.bathymetry
     JOIN filter.time_series_energy_tidal ON bathymetry.id = time_series_energy_tidal.fk_bathymetry_id;
$$;


--
-- TOC entry 1689 (class 1255 OID 30930)
-- Name: sp_build_views(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_build_views() RETURNS void
    LANGUAGE sql
    AS $$
SELECT filter.sp_build_view_bathymetry_layer();
SELECT filter.sp_build_view_cable_corridor_bathymetry_layer();
SELECT filter.sp_build_view_time_series_energy_tidal();
SELECT filter.sp_build_view_sub_systems_installation();
SELECT filter.sp_build_view_sub_systems_access();
SELECT filter.sp_build_view_sub_systems_inspection();
SELECT filter.sp_build_view_sub_systems_maintenance();
SELECT filter.sp_build_view_sub_systems_replace();
SELECT filter.sp_build_view_sub_systems_economic();
SELECT filter.sp_build_view_sub_systems_operation_weightings();
SELECT filter.sp_build_view_control_system_installation();
SELECT filter.sp_build_view_control_system_access();
SELECT filter.sp_build_view_control_system_economic();
SELECT filter.sp_build_view_control_system_inspection();
SELECT filter.sp_build_view_control_system_maintenance();
SELECT filter.sp_build_view_control_system_operation_weightings();
SELECT filter.sp_build_view_control_system_replace();
$$;


--
-- TOC entry 1690 (class 1255 OID 30931)
-- Name: sp_drop_views(); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_drop_views() RETURNS void
    LANGUAGE sql
    AS $$
DROP VIEW filter.view_bathymetry_layer;
DROP VIEW filter.view_cable_corridor_bathymetry_layer;
DROP VIEW filter.view_time_series_energy_tidal;
DROP VIEW filter.view_sub_systems_installation;
DROP VIEW filter.view_control_system_installation;
DROP VIEW filter.view_sub_systems_access;
DROP VIEW filter.view_sub_systems_inspection;
DROP VIEW filter.view_sub_systems_maintenance;
DROP VIEW filter.view_sub_systems_replace;
DROP VIEW filter.view_sub_systems_economic;
DROP VIEW filter.view_sub_systems_operation_weightings;
DROP VIEW filter.view_control_system_access;
DROP VIEW filter.view_control_system_economic;
DROP VIEW filter.view_control_system_inspection;
DROP VIEW filter.view_control_system_maintenance;
DROP VIEW filter.view_control_system_operation_weightings;
DROP VIEW filter.view_control_system_replace;
$$;


--
-- TOC entry 1691 (class 1255 OID 30932)
-- Name: sp_filter_cable_corridor_constraint(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_cable_corridor_constraint(site_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."cable_corridor_constraint" 
  SELECT * FROM "project"."cable_corridor_constraint"
  WHERE "project"."cable_corridor_constraint".fk_site_id = "site_id"; 
END;
$$;


--
-- TOC entry 1692 (class 1255 OID 30933)
-- Name: sp_filter_cable_corridor_site_bathymetry(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_cable_corridor_site_bathymetry(site_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."cable_corridor_bathymetry" 
  SELECT * FROM "project"."cable_corridor_bathymetry"
  WHERE "project"."cable_corridor_bathymetry".fk_site_id = "site_id";  
END;
$$;


--
-- TOC entry 1693 (class 1255 OID 30934)
-- Name: sp_filter_cable_corridor_site_bathymetry_layer(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_cable_corridor_site_bathymetry_layer(site_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."cable_corridor_bathymetry_layer" 
  SELECT "cable_corridor_bathymetry_layer".*
  FROM
     "project"."cable_corridor_bathymetry"
     INNER JOIN "project"."cable_corridor_bathymetry_layer"
     ON ("project"."cable_corridor_bathymetry_layer"."fk_bathymetry_id" = "project"."cable_corridor_bathymetry"."id")
     WHERE 
       "project"."cable_corridor_bathymetry"."fk_site_id" = site_id;
 END;
$$;


--
-- TOC entry 1694 (class 1255 OID 30935)
-- Name: sp_filter_constraint(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_constraint(site_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."constraint" 
  SELECT * FROM "project"."constraint"
  WHERE "project"."constraint".fk_site_id = "site_id"; 
END;
$$;


--
-- TOC entry 1695 (class 1255 OID 30936)
-- Name: sp_filter_device_data(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_device_data(device_id integer) RETURNS void
    LANGUAGE sql
    AS $$
TRUNCATE filter.device_shared;
SELECT * FROM filter.sp_filter_device_shared(device_id);

TRUNCATE filter.device_floating;
SELECT * FROM filter.sp_filter_device_floating(device_id);

TRUNCATE filter.device_tidal;
SELECT * FROM filter.sp_filter_device_tidal(device_id);

TRUNCATE filter.device_tidal_power_performance;
SELECT * FROM filter.sp_filter_device_tidal_power_performance(device_id);

TRUNCATE filter.device_wave;
SELECT * FROM filter.sp_filter_device_wave(device_id);

TRUNCATE filter.sub_systems_install;
SELECT * FROM filter.sp_filter_sub_systems_install(device_id);

TRUNCATE filter.sub_systems_access;
SELECT * FROM filter.sp_filter_sub_systems_access(device_id);

TRUNCATE filter.sub_systems_economic;
SELECT * FROM filter.sp_filter_sub_systems_economic(device_id);

TRUNCATE filter.sub_systems_inspection;
SELECT * FROM filter.sp_filter_sub_systems_inspection(device_id);

TRUNCATE filter.sub_systems_maintenance;
SELECT * FROM filter.sp_filter_sub_systems_maintenance(device_id);

TRUNCATE filter.sub_systems_operation_weightings;
SELECT * FROM filter.sp_filter_sub_systems_operation_weightings(device_id);

TRUNCATE filter.sub_systems_replace;
SELECT * FROM filter.sp_filter_sub_systems_replace(device_id);
$$;


--
-- TOC entry 1696 (class 1255 OID 30937)
-- Name: sp_filter_device_floating(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_device_floating(device_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."device_floating" 
  SELECT * FROM "project"."device_floating"
  WHERE "project"."device_floating".fk_device_id = "device_id"; 
END;
$$;


--
-- TOC entry 1697 (class 1255 OID 30938)
-- Name: sp_filter_device_shared(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_device_shared(device_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."device_shared" 
  SELECT * FROM "project"."device_shared"
  WHERE "project"."device_shared".fk_device_id = "device_id"; 
END;
$$;


--
-- TOC entry 1698 (class 1255 OID 30939)
-- Name: sp_filter_device_tidal(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_device_tidal(device_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."device_tidal" 
  SELECT * FROM "project"."device_tidal"
  WHERE "project"."device_tidal".fk_device_id = "device_id"; 
END;
$$;


--
-- TOC entry 1699 (class 1255 OID 30940)
-- Name: sp_filter_device_tidal_power_performance(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_device_tidal_power_performance(device_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."device_tidal_power_performance" 
  SELECT * FROM "project"."device_tidal_power_performance"
  WHERE "project"."device_tidal_power_performance".fk_device_id = "device_id"; 
END;
$$;


--
-- TOC entry 1700 (class 1255 OID 30941)
-- Name: sp_filter_device_wave(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_device_wave(device_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."device_wave" 
  SELECT * FROM "project"."device_wave"
  WHERE "project"."device_wave".fk_device_id = "device_id"; 
END;
$$;


--
-- TOC entry 1701 (class 1255 OID 30942)
-- Name: sp_filter_lease_area(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_lease_area(site_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."lease_area" 
  SELECT * FROM "project"."lease_area"
  WHERE "project"."lease_area".fk_site_id = "site_id"; 
END;
$$;


--
-- TOC entry 1702 (class 1255 OID 30943)
-- Name: sp_filter_site_bathymetry(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_site_bathymetry(site_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."bathymetry" 
  SELECT * FROM "project"."bathymetry"
  WHERE "project"."bathymetry".fk_site_id = "site_id"; 
END;
$$;


--
-- TOC entry 1703 (class 1255 OID 30944)
-- Name: sp_filter_site_bathymetry_layer(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_site_bathymetry_layer(site_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."bathymetry_layer" 
  SELECT "bathymetry_layer".*
  FROM
     "project"."bathymetry"
     INNER JOIN "project"."bathymetry_layer"
     ON ("project"."bathymetry_layer"."fk_bathymetry_id" = "project"."bathymetry"."id")
     WHERE 
       "project"."bathymetry"."fk_site_id" = site_id; 
END;
$$;


--
-- TOC entry 1704 (class 1255 OID 30945)
-- Name: sp_filter_site_data(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_site_data(site_id integer) RETURNS void
    LANGUAGE sql
    AS $$
TRUNCATE filter.lease_area;
SELECT * FROM filter.sp_filter_lease_area(site_id);

TRUNCATE filter.constraint;
SELECT * FROM filter.sp_filter_constraint(site_id);

TRUNCATE filter.cable_corridor_constraint;
SELECT * FROM filter.sp_filter_cable_corridor_constraint(site_id);

TRUNCATE filter.bathymetry;
SELECT * FROM filter.sp_filter_site_bathymetry(site_id);

TRUNCATE filter.bathymetry_layer;
SELECT * FROM filter.sp_filter_site_bathymetry_layer(site_id);

TRUNCATE filter.cable_corridor_bathymetry;
SELECT * FROM filter.sp_filter_cable_corridor_site_bathymetry(site_id);

TRUNCATE filter.cable_corridor_bathymetry_layer;
SELECT * FROM filter.sp_filter_cable_corridor_site_bathymetry_layer(site_id);

-- Time Series Data
TRUNCATE filter.time_series_energy_tidal;
SELECT * FROM filter.sp_filter_site_time_series_energy_tidal(site_id);

TRUNCATE filter.time_series_energy_wave;
SELECT * FROM filter.sp_filter_site_time_series_energy_wave(site_id);

TRUNCATE filter.time_series_om_tidal;
SELECT * FROM filter.sp_filter_site_time_series_om_tidal(site_id);

TRUNCATE filter.time_series_om_wave;
SELECT * FROM filter.sp_filter_site_time_series_om_wave(site_id);

TRUNCATE filter.time_series_om_wind;
SELECT * FROM filter.sp_filter_site_time_series_om_wind(site_id);
$$;


--
-- TOC entry 1705 (class 1255 OID 30946)
-- Name: sp_filter_site_time_series_energy_tidal(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_site_time_series_energy_tidal(site_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO filter.time_series_energy_tidal 
  SELECT project.time_series_energy_tidal.*
  FROM
     project.bathymetry
     INNER JOIN project.time_series_energy_tidal 
     ON (project.time_series_energy_tidal.fk_bathymetry_id = project.bathymetry.id)
     WHERE 
       project.bathymetry.fk_site_id = site_id; 
  
END;
$$;


--
-- TOC entry 1706 (class 1255 OID 30947)
-- Name: sp_filter_site_time_series_energy_wave(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_site_time_series_energy_wave(site_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."time_series_energy_wave" 
  SELECT * FROM "project"."time_series_energy_wave"
  WHERE "project"."time_series_energy_wave"."fk_site_id" = "site_id"; 
END;
$$;


--
-- TOC entry 1707 (class 1255 OID 30948)
-- Name: sp_filter_site_time_series_om_tidal(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_site_time_series_om_tidal(site_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."time_series_om_tidal" 
  SELECT * FROM "project"."time_series_om_tidal"
  WHERE "project"."time_series_om_tidal"."fk_site_id" = "site_id"; 
END;
$$;


--
-- TOC entry 1708 (class 1255 OID 30949)
-- Name: sp_filter_site_time_series_om_wave(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_site_time_series_om_wave(site_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."time_series_om_wave" 
  SELECT * FROM "project"."time_series_om_wave"
  WHERE "project"."time_series_om_wave"."fk_site_id" = "site_id";
END;
$$;


--
-- TOC entry 1613 (class 1255 OID 30950)
-- Name: sp_filter_site_time_series_om_wind(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_site_time_series_om_wind(site_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."time_series_om_wind" 
  SELECT * FROM "project"."time_series_om_wind"
  WHERE "project"."time_series_om_wind"."fk_site_id" = "site_id"; 
END;
$$;


--
-- TOC entry 1709 (class 1255 OID 30951)
-- Name: sp_filter_sub_systems_access(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_sub_systems_access(device_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."sub_systems_access" 
  SELECT
     "sub_systems_access".*
  FROM
     "project"."sub_systems"
     INNER JOIN "project"."sub_systems_access"
     ON ("project"."sub_systems_access"."fk_sub_system_id" = "project"."sub_systems"."id")
     WHERE 
       "project"."sub_systems"."fk_device_id" = device_id; 
END;
$$;


--
-- TOC entry 1710 (class 1255 OID 30952)
-- Name: sp_filter_sub_systems_economic(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_sub_systems_economic(device_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."sub_systems_economic" 
  SELECT
     "sub_systems_economic".*
  FROM
     "project"."sub_systems"
     INNER JOIN "project"."sub_systems_economic"
     ON ("project"."sub_systems_economic"."fk_sub_system_id" = "project"."sub_systems"."id")
     WHERE 
       "project"."sub_systems"."fk_device_id" = device_id; 
END;
$$;


--
-- TOC entry 1711 (class 1255 OID 30953)
-- Name: sp_filter_sub_systems_inspection(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_sub_systems_inspection(device_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."sub_systems_inspection" 
  SELECT
     "sub_systems_inspection".*
  FROM
     "project"."sub_systems"
     INNER JOIN "project"."sub_systems_inspection"
     ON ("project"."sub_systems_inspection"."fk_sub_system_id" = "project"."sub_systems"."id")
     WHERE 
       "project"."sub_systems"."fk_device_id" = device_id; 
END;
$$;


--
-- TOC entry 1712 (class 1255 OID 30954)
-- Name: sp_filter_sub_systems_install(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_sub_systems_install(device_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."sub_systems_install" 
  SELECT
     "sub_systems_install".*
  FROM
     "project"."sub_systems"
     INNER JOIN "project"."sub_systems_install"
     ON ("project"."sub_systems_install"."fk_sub_system_id" = "project"."sub_systems"."id")
     WHERE 
       "project"."sub_systems"."fk_device_id" = device_id; 
END;
$$;


--
-- TOC entry 1713 (class 1255 OID 30955)
-- Name: sp_filter_sub_systems_maintenance(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_sub_systems_maintenance(device_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."sub_systems_maintenance" 
  SELECT
     "sub_systems_maintenance".*
  FROM
     "project"."sub_systems"
     INNER JOIN "project"."sub_systems_maintenance"
     ON ("project"."sub_systems_maintenance"."fk_sub_system_id" = "project"."sub_systems"."id")
     WHERE 
       "project"."sub_systems"."fk_device_id" = device_id; 
END;
$$;


--
-- TOC entry 1714 (class 1255 OID 30956)
-- Name: sp_filter_sub_systems_operation_weightings(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_sub_systems_operation_weightings(device_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."sub_systems_operation_weightings" 
  SELECT
     "sub_systems_operation_weightings".*
  FROM
     "project"."sub_systems"
     INNER JOIN "project"."sub_systems_operation_weightings"
     ON ("project"."sub_systems_operation_weightings"."fk_sub_system_id" = "project"."sub_systems"."id")
     WHERE 
       "project"."sub_systems"."fk_device_id" = device_id; 
END;
$$;


--
-- TOC entry 1715 (class 1255 OID 30957)
-- Name: sp_filter_sub_systems_replace(integer); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_filter_sub_systems_replace(device_id integer) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
  INSERT INTO "filter"."sub_systems_replace" 
  SELECT
     "sub_systems_replace".*
  FROM
     "project"."sub_systems"
     INNER JOIN "project"."sub_systems_replace"
     ON ("project"."sub_systems_replace"."fk_sub_system_id" = "project"."sub_systems"."id")
     WHERE 
       "project"."sub_systems"."fk_device_id" = device_id; 
END;
$$;


--
-- TOC entry 1716 (class 1255 OID 30958)
-- Name: sp_select_bathymetry_by_polygon(text); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_select_bathymetry_by_polygon(polystring text) RETURNS TABLE(utm_point public.geometry, depth double precision, mannings_no double precision, layer_order smallint, initial_depth double precision, sediment_type character varying)
    LANGUAGE sql ROWS 100000
    AS $$
SELECT 
  utm_point,
  depth,
  mannings_no,
  layer_order,
  initial_depth,
  sediment_type
FROM filter.view_bathymetry_layer
WHERE 
(ST_Covers(ST_GeomFromText('POLYGON(('|| polystring || '))', 0), filter.view_bathymetry_layer.utm_point));
$$;


--
-- TOC entry 1717 (class 1255 OID 30959)
-- Name: sp_select_cable_corridor_bathymetry_by_polygon(text); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_select_cable_corridor_bathymetry_by_polygon(polystring text) RETURNS TABLE(utm_point public.geometry, depth double precision, layer_order smallint, initial_depth double precision, sediment_type character varying)
    LANGUAGE sql ROWS 100000
    AS $$
SELECT 
  utm_point,
  depth,
  layer_order,
  initial_depth,
  sediment_type
FROM filter.view_cable_corridor_bathymetry_layer
WHERE 
(ST_Covers(ST_GeomFromText('POLYGON(('|| polystring || '))', -1), filter.view_cable_corridor_bathymetry_layer.utm_point));
$$;


--
-- TOC entry 1718 (class 1255 OID 30960)
-- Name: sp_select_tidal_energy_time_series_by_polygon(text); Type: FUNCTION; Schema: filter; Owner: -
--

CREATE FUNCTION filter.sp_select_tidal_energy_time_series_by_polygon(polystring text) RETURNS TABLE(utm_point public.geometry, measure_date date, measure_time time without time zone, u double precision, v double precision, turbulence_intensity double precision, ssh double precision)
    LANGUAGE sql ROWS 5
    AS $$
SELECT
  utm_point,
  measure_date,
  measure_time,
  u,
  v,
  turbulence_intensity,
  ssh
FROM filter.view_time_series_energy_tidal
WHERE 
(ST_Covers(ST_GeomFromText('POLYGON(('|| polystring || '))', 0), filter.view_time_series_energy_tidal.utm_point));
$$;


--
-- TOC entry 1719 (class 1255 OID 30961)
-- Name: sp_build_view_component_cable_dynamic(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_component_cable_dynamic() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_component_cable_dynamic AS 
 SELECT component.id AS component_id,
    component.description,
    component_continuous.diameter,
    component_continuous.dry_mass_per_unit_length,
    component_continuous.wet_mass_per_unit_length,
    minimum_breaking_load,
    minimum_bend_radius,
    number_conductors,
    number_fibre_channels,
    resistance_dc_20,
    resistance_ac_90,
    inductive_reactance,
    capacitance,
    rated_current_air,
    rated_current_buried,
    rated_current_jtube,
    rated_voltage_u0,
    operational_temp_max,
    component_continuous.cost_per_unit_length,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM reference.component_cable
   JOIN reference.component_continuous ON component_cable.fk_component_continuous_id = component_continuous.id
   JOIN reference.component_shared ON component_continuous.fk_component_id = component_shared.fk_component_id
   JOIN reference.component ON component_continuous.fk_component_id = component.id
   JOIN reference.component_type ON component_cable.fk_component_type_id = component_type.id
   WHERE component_type.description::text = 'cable dynamic'::text;
  $$;


--
-- TOC entry 1720 (class 1255 OID 30962)
-- Name: sp_build_view_component_cable_static(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_component_cable_static() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_component_cable_static AS 
 SELECT component.id AS component_id,
    component.description,
    component_continuous.diameter,
    component_continuous.dry_mass_per_unit_length,
    component_continuous.wet_mass_per_unit_length,
    minimum_breaking_load,
    minimum_bend_radius,
    number_conductors,
    number_fibre_channels,
    resistance_dc_20,
    resistance_ac_90,
    inductive_reactance,
    capacitance,
    rated_current_air,
    rated_current_buried,
    rated_current_jtube,
    rated_voltage_u0,
    operational_temp_max,
    component_continuous.cost_per_unit_length,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM reference.component_cable
   JOIN reference.component_continuous ON component_cable.fk_component_continuous_id = component_continuous.id
   JOIN reference.component_shared ON component_continuous.fk_component_id = component_shared.fk_component_id
   JOIN reference.component ON component_continuous.fk_component_id = component.id
   JOIN reference.component_type ON component_cable.fk_component_type_id = component_type.id
   WHERE component_type.description::text = 'cable static'::text;
$$;


--
-- TOC entry 1721 (class 1255 OID 30963)
-- Name: sp_build_view_component_collection_point(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_component_collection_point() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_component_collection_point AS 
 SELECT component.id AS component_id,
    component.description,
    component_discrete.width,
    component_discrete.length AS depth,
    component_discrete.height,
    component_discrete.dry_mass,
    component_discrete.wet_mass,
    wet_frontal_area,
    dry_frontal_area,
    wet_beam_area,
    dry_beam_area,
    maximum_water_depth,
    orientation_angle,
    input_lines,
    output_lines,
    input_connector_type,
    output_connector_type,
    number_fibre_channels,
    voltage_primary_winding,
    voltage_secondary_winding,
    rated_operating_current,
    operational_temp_min,
    operational_temp_max,
    foundation_locations,
    centre_of_gravity,
    component_discrete.cost,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM reference.component_collection_point
   JOIN reference.component_discrete ON component_collection_point.fk_component_discrete_id = component_discrete.id
   JOIN reference.component_shared ON component_discrete.fk_component_id = component_shared.fk_component_id
   JOIN reference.component ON component_discrete.fk_component_id = component.id;
$$;


--
-- TOC entry 1722 (class 1255 OID 30964)
-- Name: sp_build_view_component_connector_drymate(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_component_connector_drymate() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_component_connector_drymate AS 
 SELECT component.id AS component_id,
    component.description,
    component_discrete.width,
    component_discrete.length AS depth,
    component_discrete.height,
    component_discrete.dry_mass,
    component_discrete.wet_mass,
    maximum_water_depth,
    number_contacts,
    number_fibre_channels,
    mating_force,
    demating_force,
    rated_voltage_u0,
    rated_current,
    cable_area_min,
    cable_area_max,
    operational_temp_min,
    operational_temp_max,
    component_discrete.cost,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM reference.component_connector
   JOIN reference.component_discrete ON component_connector.fk_component_discrete_id = component_discrete.id
   JOIN reference.component_shared ON component_discrete.fk_component_id = component_shared.fk_component_id
   JOIN reference.component ON component_discrete.fk_component_id = component.id
   JOIN reference.component_type ON component_connector.fk_component_type_id = component_type.id
   WHERE component_type.description::text = 'connector dry-mate'::text;
$$;


--
-- TOC entry 1723 (class 1255 OID 30965)
-- Name: sp_build_view_component_connector_wetmate(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_component_connector_wetmate() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_component_connector_wetmate AS 
 SELECT component.id AS component_id,
    component.description,
    component_discrete.width,
    component_discrete.length AS depth,
    component_discrete.height,
    component_discrete.dry_mass,
    component_discrete.wet_mass,
    maximum_water_depth,
    number_contacts,
    number_fibre_channels,
    mating_force,
    demating_force,
    rated_voltage_u0,
    rated_current,
    cable_area_min,
    cable_area_max,
    operational_temp_min,
    operational_temp_max,
    component_discrete.cost,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM reference.component_connector
   JOIN reference.component_discrete ON component_connector.fk_component_discrete_id = component_discrete.id
   JOIN reference.component_shared ON component_discrete.fk_component_id = component_shared.fk_component_id
   JOIN reference.component ON component_discrete.fk_component_id = component.id
   JOIN reference.component_type ON component_connector.fk_component_type_id = component_type.id
   WHERE component_type.description::text = 'connector wet-mate'::text;
$$;


--
-- TOC entry 1725 (class 1255 OID 30966)
-- Name: sp_build_view_component_foundations_anchor(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_component_foundations_anchor() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_component_foundations_anchor AS 
 SELECT component.id AS component_id,
    component.description,
    component_discrete.width,
    component_discrete.length AS depth,
    component_discrete.height,
    component_discrete.dry_mass,
    component_discrete.wet_mass,
    connecting_size,
    minimum_breaking_load,
    axial_stiffness,
    component_discrete.cost,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM reference.component_anchor
   JOIN reference.component_discrete ON component_anchor.fk_component_discrete_id = component_discrete.id
   JOIN reference.component_shared ON component_discrete.fk_component_id = component_shared.fk_component_id
   JOIN reference.component ON component_discrete.fk_component_id = component.id;
$$;


--
-- TOC entry 1726 (class 1255 OID 30967)
-- Name: sp_build_view_component_foundations_anchor_coefs(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_component_foundations_anchor_coefs() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_component_foundations_anchor_coefs AS 
 SELECT component.id AS component_id,
    soft_holding_cap_coef_1,
    soft_holding_cap_coef_2,
    soft_penetration_coef_1,
    soft_penetration_coef_2,
    sand_holding_cap_coef_1,
    sand_holding_cap_coef_2,
    sand_penetration_coef_1,
    sand_penetration_coef_2
   FROM reference.component_anchor
   JOIN reference.component_discrete ON component_anchor.fk_component_discrete_id = component_discrete.id
   JOIN reference.component ON component_discrete.fk_component_id = component.id;
$$;


--
-- TOC entry 1727 (class 1255 OID 30968)
-- Name: sp_build_view_component_foundations_pile(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_component_foundations_pile() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_component_foundations_pile AS 
 SELECT component.id AS component_id,
    component.description,
    component_continuous.diameter,
    component_continuous.dry_mass_per_unit_length,
    component_continuous.wet_mass_per_unit_length,
    wall_thickness,
    yield_stress,
    youngs_modulus,
    component_continuous.cost_per_unit_length,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM reference.component_pile
   JOIN reference.component_continuous ON component_pile.fk_component_continuous_id = component_continuous.id
   JOIN reference.component_shared ON component_continuous.fk_component_id = component_shared.fk_component_id
   JOIN reference.component ON component_continuous.fk_component_id = component.id;
$$;


--
-- TOC entry 1728 (class 1255 OID 30969)
-- Name: sp_build_view_component_moorings_chain(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_component_moorings_chain() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_component_moorings_chain AS 
 SELECT component.id AS component_id,
    component.description,
    component_continuous.diameter,
    component_continuous.dry_mass_per_unit_length,
    component_continuous.wet_mass_per_unit_length,
    connecting_length,
    minimum_breaking_load,
    axial_stiffness,
    component_continuous.cost_per_unit_length,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM reference.component_mooring_continuous
   JOIN reference.component_continuous ON component_mooring_continuous.fk_component_continuous_id = component_continuous.id
   JOIN reference.component_shared ON component_continuous.fk_component_id = component_shared.fk_component_id
   JOIN reference.component ON component_continuous.fk_component_id = component.id
   JOIN reference.component_type ON component_mooring_continuous.fk_component_type_id = component_type.id
   WHERE component_type.description::text = 'chain'::text;
$$;


--
-- TOC entry 1729 (class 1255 OID 30970)
-- Name: sp_build_view_component_moorings_forerunner(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_component_moorings_forerunner() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_component_moorings_forerunner AS 
 SELECT component.id AS component_id,
    component.description,
    component_continuous.diameter,
    component_continuous.dry_mass_per_unit_length,
    component_continuous.wet_mass_per_unit_length,
    connecting_length,
    minimum_breaking_load,
    axial_stiffness,
    component_continuous.cost_per_unit_length,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM reference.component_mooring_continuous
   JOIN reference.component_continuous ON component_mooring_continuous.fk_component_continuous_id = component_continuous.id
   JOIN reference.component_shared ON component_continuous.fk_component_id = component_shared.fk_component_id
   JOIN reference.component ON component_continuous.fk_component_id = component.id
   JOIN reference.component_type ON component_mooring_continuous.fk_component_type_id = component_type.id
   WHERE component_type.description::text = 'forerunner'::text;
$$;


--
-- TOC entry 1730 (class 1255 OID 30971)
-- Name: sp_build_view_component_moorings_rope(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_component_moorings_rope() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_component_moorings_rope AS 
 SELECT component.id AS component_id,
    component.description,
    component_continuous.diameter,
    component_continuous.dry_mass_per_unit_length,
    component_continuous.wet_mass_per_unit_length,
    material,
    minimum_breaking_load,
    rope_stiffness_curve,
    component_continuous.cost_per_unit_length,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM reference.component_rope
   JOIN reference.component_continuous ON component_rope.fk_component_continuous_id = component_continuous.id
   JOIN reference.component_shared ON component_continuous.fk_component_id = component_shared.fk_component_id
   JOIN reference.component ON component_continuous.fk_component_id = component.id;
$$;


--
-- TOC entry 1731 (class 1255 OID 30972)
-- Name: sp_build_view_component_moorings_shackle(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_component_moorings_shackle() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_component_moorings_shackle AS 
 SELECT component.id AS component_id,
    component.description,
    component_discrete.width,
    component_discrete.length AS depth,
    component_discrete.height,
    component_discrete.dry_mass,
    component_discrete.wet_mass,
    nominal_diameter,
    connecting_length,
    minimum_breaking_load,
    axial_stiffness,
    component_discrete.cost,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM reference.component_mooring_discrete
   JOIN reference.component_discrete ON component_mooring_discrete.fk_component_discrete_id = component_discrete.id
   JOIN reference.component_shared ON component_discrete.fk_component_id = component_shared.fk_component_id
   JOIN reference.component ON component_discrete.fk_component_id = component.id
   JOIN reference.component_type ON component_mooring_discrete.fk_component_type_id = component_type.id
   WHERE component_type.description::text = 'shackle'::text;
$$;


--
-- TOC entry 1732 (class 1255 OID 30973)
-- Name: sp_build_view_component_moorings_swivel(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_component_moorings_swivel() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_component_moorings_swivel AS 
 SELECT component.id AS component_id,
    component.description,
    component_discrete.width,
    component_discrete.length AS depth,
    component_discrete.height,
    component_discrete.dry_mass,
    component_discrete.wet_mass,
    nominal_diameter,
    connecting_length,
    minimum_breaking_load,
    axial_stiffness,
    component_discrete.cost,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM reference.component_mooring_discrete
   JOIN reference.component_discrete ON component_mooring_discrete.fk_component_discrete_id = component_discrete.id
   JOIN reference.component_shared ON component_discrete.fk_component_id = component_shared.fk_component_id
   JOIN reference.component ON component_discrete.fk_component_id = component.id
   JOIN reference.component_type ON component_mooring_discrete.fk_component_type_id = component_type.id
   WHERE component_type.description::text = 'swivel'::text;
$$;


--
-- TOC entry 1733 (class 1255 OID 30974)
-- Name: sp_build_view_component_transformer(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_component_transformer() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_component_transformer AS 
 SELECT component.id AS component_id,
    component.description,
    component_discrete.width,
    component_discrete.length AS depth,
    component_discrete.height,
    component_discrete.dry_mass,
    component_discrete.wet_mass,
    maximum_water_depth,
    power_rating,
    impedance,
    windings,
    voltage_primary_winding,
    voltage_secondary_winding,
    voltage_tertiary_winding,
    operational_temp_min,
    operational_temp_max,
    component_discrete.cost,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM reference.component_transformer
   JOIN reference.component_discrete ON component_transformer.fk_component_discrete_id = component_discrete.id
   JOIN reference.component_shared ON component_discrete.fk_component_id = component_shared.fk_component_id
   JOIN reference.component ON component_discrete.fk_component_id = component.id;
$$;


--
-- TOC entry 1734 (class 1255 OID 30975)
-- Name: sp_build_view_operations_limit_cs(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_operations_limit_cs() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_operations_limit_cs AS 
  SELECT
    operations_type.description AS operations_type,
    cs_limit
  FROM reference.operations_limit_cs
  JOIN reference.operations_type ON operations_limit_cs.fk_operations_id = operations_type.id;
$$;


--
-- TOC entry 1735 (class 1255 OID 30976)
-- Name: sp_build_view_operations_limit_hs(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_operations_limit_hs() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_operations_limit_hs AS 
  SELECT
    operations_type.description AS operations_type,
    hs_limit
  FROM reference.operations_limit_hs
  JOIN reference.operations_type ON operations_limit_hs.fk_operations_id = operations_type.id;
$$;


--
-- TOC entry 1736 (class 1255 OID 30977)
-- Name: sp_build_view_operations_limit_tp(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_operations_limit_tp() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_operations_limit_tp AS 
  SELECT
    operations_type.description AS operations_type,
    tp_limit
  FROM reference.operations_limit_tp
  JOIN reference.operations_type ON operations_limit_tp.fk_operations_id = operations_type.id;
$$;


--
-- TOC entry 1737 (class 1255 OID 30978)
-- Name: sp_build_view_operations_limit_ws(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_operations_limit_ws() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_operations_limit_ws AS 
  SELECT
    operations_type.description AS operations_type,
    ws_limit
  FROM reference.operations_limit_ws
  JOIN reference.operations_type ON operations_limit_ws.fk_operations_id = operations_type.id;
$$;


--
-- TOC entry 1738 (class 1255 OID 30979)
-- Name: sp_build_view_soil_type_geotechnical_properties(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_soil_type_geotechnical_properties() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_soil_type_geotechnical_properties AS 
  SELECT
    soil_type.description AS soil_type,
    drained_soil_friction_angle,
    relative_soil_density,
    buoyant_unit_weight_of_soil,
    undrained_soil_shear_strength_constant_term,
    undrained_soil_shear_strength_depth_dependent_term,
    effective_drained_cohesion,
    seafloor_friction_coefficient,
    soil_sensitivity,
    rock_compressive_strength
  FROM reference.soil_type_geotechnical_properties
  JOIN reference.soil_type ON soil_type_geotechnical_properties.fk_soil_type_id = soil_type.id;
$$;


--
-- TOC entry 1724 (class 1255 OID 30980)
-- Name: sp_build_view_vehicle_helicopter(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_vehicle_helicopter() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_vehicle_helicopter AS 
 SELECT
    vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    deck_space,
    max_deck_load_pressure,
    max_cargo_mass,
    crane_max_load_mass,
    external_personel,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM reference.vehicle_helicopter
   JOIN reference.vehicle_shared ON vehicle_helicopter.fk_vehicle_id = vehicle_shared.fk_vehicle_id
   JOIN reference.vehicle ON vehicle_shared.fk_vehicle_id = vehicle.id;
$$;


--
-- TOC entry 1740 (class 1255 OID 30981)
-- Name: sp_build_view_vehicle_vessel_ahts(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_vehicle_vessel_ahts() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_vehicle_vessel_ahts AS 
 SELECT
    vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    beam,
    max_draft,
    vehicle_shared.consumption,
    consumption_towing,
    vehicle_shared.transit_speed,
    deck_space,
    max_deck_load_pressure,
    max_cargo_mass,
    crane_max_load_mass,
    bollard_pull,
    anchor_handling_drum_capacity,
    anchor_handling_winch_rated_pull,
    external_personel,
    towing_max_hs,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM reference.vehicle_vessel_anchor_handling
   JOIN reference.vehicle_shared ON vehicle_vessel_anchor_handling.fk_vehicle_id = vehicle_shared.fk_vehicle_id
   JOIN reference.vehicle ON vehicle_shared.fk_vehicle_id = vehicle.id
   JOIN reference.vehicle_type ON vehicle_vessel_anchor_handling.fk_vehicle_type_id = vehicle_type.id
   WHERE vehicle_type.description::text = 'anchor handling tug supply vessel'::text;
$$;


--
-- TOC entry 1741 (class 1255 OID 30982)
-- Name: sp_build_view_vehicle_vessel_barge(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_vehicle_vessel_barge() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_vehicle_vessel_barge AS 
 SELECT
    vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    beam,
    max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    deck_space,
    max_deck_load_pressure,
    max_cargo_mass,
    crane_max_load_mass,
    dynamic_positioning_capabilities,
    external_personel,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM reference.vehicle_vessel_cargo
   JOIN reference.vehicle_shared ON vehicle_vessel_cargo.fk_vehicle_id = vehicle_shared.fk_vehicle_id
   JOIN reference.vehicle ON vehicle_shared.fk_vehicle_id = vehicle.id
   JOIN reference.vehicle_type ON vehicle_vessel_cargo.fk_vehicle_type_id = vehicle_type.id
   WHERE vehicle_type.description::text = 'barge'::text;
$$;


--
-- TOC entry 1742 (class 1255 OID 30983)
-- Name: sp_build_view_vehicle_vessel_clb(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_vehicle_vessel_clb() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_vehicle_vessel_clb AS 
 SELECT
    vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    beam,
    max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    deck_space,
    max_deck_load_pressure,
    max_cargo_mass,
    crane_max_load_mass,
    number_turntables,
    turntable_max_load_mass,
    turntable_inner_diameter,
    cable_splice_capabilities,
    dynamic_positioning_capabilities,
    external_personel,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM reference.vehicle_vessel_cable_laying
   JOIN reference.vehicle_shared ON vehicle_vessel_cable_laying.fk_vehicle_id = vehicle_shared.fk_vehicle_id
   JOIN reference.vehicle ON vehicle_shared.fk_vehicle_id = vehicle.id
   JOIN reference.vehicle_type ON vehicle_vessel_cable_laying.fk_vehicle_type_id = vehicle_type.id
   WHERE vehicle_type.description::text = 'cable laying barge'::text;
$$;


--
-- TOC entry 1743 (class 1255 OID 30984)
-- Name: sp_build_view_vehicle_vessel_clv(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_vehicle_vessel_clv() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_vehicle_vessel_clv AS 
 SELECT
    vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    beam,
    max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    deck_space,
    max_deck_load_pressure,
    max_cargo_mass,
    crane_max_load_mass,
    bollard_pull,
    number_turntables,
    turntable_max_load_mass,
    turntable_inner_diameter,
    cable_splice_capabilities,
    dynamic_positioning_capabilities,
    external_personel,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM reference.vehicle_vessel_cable_laying
   JOIN reference.vehicle_shared ON vehicle_vessel_cable_laying.fk_vehicle_id = vehicle_shared.fk_vehicle_id
   JOIN reference.vehicle ON vehicle_shared.fk_vehicle_id = vehicle.id
   JOIN reference.vehicle_type ON vehicle_vessel_cable_laying.fk_vehicle_type_id = vehicle_type.id
   WHERE vehicle_type.description::text = 'cable laying vessel'::text;
$$;


--
-- TOC entry 1744 (class 1255 OID 30985)
-- Name: sp_build_view_vehicle_vessel_crane_barge(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_vehicle_vessel_crane_barge() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_vehicle_vessel_crane_barge AS 
 SELECT
    vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    beam,
    max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    deck_space,
    max_deck_load_pressure,
    max_cargo_mass,
    crane_max_load_mass,
    dynamic_positioning_capabilities,
    external_personel,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM reference.vehicle_vessel_cargo
   JOIN reference.vehicle_shared ON vehicle_vessel_cargo.fk_vehicle_id = vehicle_shared.fk_vehicle_id
   JOIN reference.vehicle ON vehicle_shared.fk_vehicle_id = vehicle.id
   JOIN reference.vehicle_type ON vehicle_vessel_cargo.fk_vehicle_type_id = vehicle_type.id
   WHERE vehicle_type.description::text = 'crane barge'::text;
$$;


--
-- TOC entry 1745 (class 1255 OID 30986)
-- Name: sp_build_view_vehicle_vessel_crane_vessel(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_vehicle_vessel_crane_vessel() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_vehicle_vessel_crane_vessel AS 
 SELECT
    vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    beam,
    max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    deck_space,
    max_deck_load_pressure,
    max_cargo_mass,
    crane_max_load_mass,
    dynamic_positioning_capabilities,
    external_personel,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM reference.vehicle_vessel_cargo
   JOIN reference.vehicle_shared ON vehicle_vessel_cargo.fk_vehicle_id = vehicle_shared.fk_vehicle_id
   JOIN reference.vehicle ON vehicle_shared.fk_vehicle_id = vehicle.id
   JOIN reference.vehicle_type ON vehicle_vessel_cargo.fk_vehicle_type_id = vehicle_type.id
   WHERE vehicle_type.description::text = 'crane vessel'::text;
$$;


--
-- TOC entry 1746 (class 1255 OID 30987)
-- Name: sp_build_view_vehicle_vessel_csv(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_vehicle_vessel_csv() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_vehicle_vessel_csv AS 
 SELECT
    vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    beam,
    max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    deck_space,
    max_deck_load_pressure,
    max_cargo_mass,
    crane_max_load_mass,
    dynamic_positioning_capabilities,
    external_personel,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM reference.vehicle_vessel_cargo
   JOIN reference.vehicle_shared ON vehicle_vessel_cargo.fk_vehicle_id = vehicle_shared.fk_vehicle_id
   JOIN reference.vehicle ON vehicle_shared.fk_vehicle_id = vehicle.id
   JOIN reference.vehicle_type ON vehicle_vessel_cargo.fk_vehicle_type_id = vehicle_type.id
   WHERE vehicle_type.description::text = 'construction support vessel'::text;
$$;


--
-- TOC entry 1747 (class 1255 OID 30988)
-- Name: sp_build_view_vehicle_vessel_ctv(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_vehicle_vessel_ctv() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_vehicle_vessel_ctv AS 
 SELECT
    vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    beam,
    max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    deck_space,
    max_deck_load_pressure,
    max_cargo_mass,
    crane_max_load_mass,
    dynamic_positioning_capabilities,
    external_personel,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM reference.vehicle_vessel_cargo
   JOIN reference.vehicle_shared ON vehicle_vessel_cargo.fk_vehicle_id = vehicle_shared.fk_vehicle_id
   JOIN reference.vehicle ON vehicle_shared.fk_vehicle_id = vehicle.id
   JOIN reference.vehicle_type ON vehicle_vessel_cargo.fk_vehicle_type_id = vehicle_type.id
   WHERE vehicle_type.description::text = 'crew transfer vessel'::text;
$$;


--
-- TOC entry 1748 (class 1255 OID 30989)
-- Name: sp_build_view_vehicle_vessel_jackup_barge(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_vehicle_vessel_jackup_barge() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_vehicle_vessel_jackup_barge AS 
 SELECT
    vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    beam,
    max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    deck_space,
    max_deck_load_pressure,
    max_cargo_mass,
    crane_max_load_mass,
    dynamic_positioning_capabilities,
    external_personel,
    jackup_max_water_depth,
    jackup_speed_down,
    jackup_max_payload_mass,
    jacking_max_hs,
    jacking_max_tp,
    jacking_max_cs,
    jacking_max_ws,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM reference.vehicle_vessel_jackup
   JOIN reference.vehicle_shared ON vehicle_vessel_jackup.fk_vehicle_id = vehicle_shared.fk_vehicle_id
   JOIN reference.vehicle ON vehicle_shared.fk_vehicle_id = vehicle.id
   JOIN reference.vehicle_type ON vehicle_vessel_jackup.fk_vehicle_type_id = vehicle_type.id
   WHERE vehicle_type.description::text = 'jackup barge'::text;
$$;


--
-- TOC entry 1749 (class 1255 OID 30990)
-- Name: sp_build_view_vehicle_vessel_jackup_vessel(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_vehicle_vessel_jackup_vessel() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_vehicle_vessel_jackup_vessel AS 
 SELECT
    vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    beam,
    max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    deck_space,
    max_deck_load_pressure,
    max_cargo_mass,
    crane_max_load_mass,
    dynamic_positioning_capabilities,
    external_personel,
    jackup_max_water_depth,
    jackup_speed_down,
    jackup_max_payload_mass,
    jacking_max_hs,
    jacking_max_tp,
    jacking_max_cs,
    jacking_max_ws,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM reference.vehicle_vessel_jackup
   JOIN reference.vehicle_shared ON vehicle_vessel_jackup.fk_vehicle_id = vehicle_shared.fk_vehicle_id
   JOIN reference.vehicle ON vehicle_shared.fk_vehicle_id = vehicle.id
   JOIN reference.vehicle_type ON vehicle_vessel_jackup.fk_vehicle_type_id = vehicle_type.id
   WHERE vehicle_type.description::text = 'jackup vessel'::text;
$$;


--
-- TOC entry 1750 (class 1255 OID 30991)
-- Name: sp_build_view_vehicle_vessel_multicat(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_vehicle_vessel_multicat() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_vehicle_vessel_multicat AS 
 SELECT
    vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    beam,
    max_draft,
    vehicle_shared.consumption,
    consumption_towing,
    vehicle_shared.transit_speed,
    deck_space,
    max_deck_load_pressure,
    max_cargo_mass,
    crane_max_load_mass,
    bollard_pull,
    anchor_handling_drum_capacity,
    anchor_handling_winch_rated_pull,
    external_personel,
    towing_max_hs,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM reference.vehicle_vessel_anchor_handling
   JOIN reference.vehicle_shared ON vehicle_vessel_anchor_handling.fk_vehicle_id = vehicle_shared.fk_vehicle_id
   JOIN reference.vehicle ON vehicle_shared.fk_vehicle_id = vehicle.id
   JOIN reference.vehicle_type ON vehicle_vessel_anchor_handling.fk_vehicle_type_id = vehicle_type.id
   WHERE vehicle_type.description::text = 'multicat'::text;
$$;


--
-- TOC entry 1751 (class 1255 OID 30992)
-- Name: sp_build_view_vehicle_vessel_tugboat(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_view_vehicle_vessel_tugboat() RETURNS void
    LANGUAGE sql
    AS $$
CREATE OR REPLACE VIEW reference.view_vehicle_vessel_tugboat AS 
 SELECT
    vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    beam,
    max_draft,
    vehicle_shared.consumption,
    consumption_towing,
    vehicle_shared.transit_speed,
    bollard_pull,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM reference.vehicle_vessel_tugboat
   JOIN reference.vehicle_shared ON vehicle_vessel_tugboat.fk_vehicle_id = vehicle_shared.fk_vehicle_id
   JOIN reference.vehicle ON vehicle_shared.fk_vehicle_id = vehicle.id;
$$;


--
-- TOC entry 1739 (class 1255 OID 30993)
-- Name: sp_build_views(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_build_views() RETURNS void
    LANGUAGE sql
    AS $$
SELECT reference.sp_build_view_component_cable_dynamic();
SELECT reference.sp_build_view_component_cable_static();
SELECT reference.sp_build_view_component_collection_point();
SELECT reference.sp_build_view_component_connector_drymate();
SELECT reference.sp_build_view_component_connector_wetmate();
SELECT reference.sp_build_view_component_foundations_anchor();
SELECT reference.sp_build_view_component_foundations_anchor_coefs();
SELECT reference.sp_build_view_component_foundations_pile();
SELECT reference.sp_build_view_component_moorings_chain();
SELECT reference.sp_build_view_component_moorings_forerunner();
SELECT reference.sp_build_view_component_moorings_rope();
SELECT reference.sp_build_view_component_moorings_shackle();
SELECT reference.sp_build_view_component_moorings_swivel();
SELECT reference.sp_build_view_component_transformer();
SELECT reference.sp_build_view_operations_limit_hs();
SELECT reference.sp_build_view_operations_limit_tp();
SELECT reference.sp_build_view_operations_limit_ws();
SELECT reference.sp_build_view_operations_limit_cs();
SELECT reference.sp_build_view_soil_type_geotechnical_properties();
SELECT reference.sp_build_view_vehicle_helicopter();
SELECT reference.sp_build_view_vehicle_vessel_ahts();
SELECT reference.sp_build_view_vehicle_vessel_barge();
SELECT reference.sp_build_view_vehicle_vessel_clb();
SELECT reference.sp_build_view_vehicle_vessel_clv();
SELECT reference.sp_build_view_vehicle_vessel_crane_barge();
SELECT reference.sp_build_view_vehicle_vessel_crane_vessel();
SELECT reference.sp_build_view_vehicle_vessel_csv();
SELECT reference.sp_build_view_vehicle_vessel_ctv();
SELECT reference.sp_build_view_vehicle_vessel_jackup_barge();
SELECT reference.sp_build_view_vehicle_vessel_jackup_vessel();
SELECT reference.sp_build_view_vehicle_vessel_multicat();
SELECT reference.sp_build_view_vehicle_vessel_tugboat();
$$;


--
-- TOC entry 1752 (class 1255 OID 30994)
-- Name: sp_drop_views(); Type: FUNCTION; Schema: reference; Owner: -
--

CREATE FUNCTION reference.sp_drop_views() RETURNS void
    LANGUAGE sql
    AS $$
DROP VIEW reference.view_component_cable_dynamic;
DROP VIEW reference.view_component_cable_static;
DROP VIEW reference.view_component_collection_point;
DROP VIEW reference.view_component_connector_drymate;
DROP VIEW reference.view_component_connector_wetmate;
DROP VIEW reference.view_component_foundations_anchor;
DROP VIEW reference.view_component_foundations_anchor_coefs;
DROP VIEW reference.view_component_foundations_pile;
DROP VIEW reference.view_component_moorings_chain;
DROP VIEW reference.view_component_moorings_forerunner;
DROP VIEW reference.view_component_moorings_rope;
DROP VIEW reference.view_component_moorings_shackle;
DROP VIEW reference.view_component_moorings_swivel;
DROP VIEW reference.view_component_transformer;
DROP VIEW reference.view_operations_limit_hs;
DROP VIEW reference.view_operations_limit_tp;
DROP VIEW reference.view_operations_limit_ws;
DROP VIEW reference.view_operations_limit_cs;
DROP VIEW reference.view_soil_type_geotechnical_properties;
DROP VIEW reference.view_vehicle_helicopter;
DROP VIEW reference.view_vehicle_vessel_ahts;
DROP VIEW reference.view_vehicle_vessel_barge;
DROP VIEW reference.view_vehicle_vessel_clb;
DROP VIEW reference.view_vehicle_vessel_clv;
DROP VIEW reference.view_vehicle_vessel_crane_barge;
DROP VIEW reference.view_vehicle_vessel_crane_vessel;
DROP VIEW reference.view_vehicle_vessel_csv;
DROP VIEW reference.view_vehicle_vessel_ctv;
DROP VIEW reference.view_vehicle_vessel_jackup_barge;
DROP VIEW reference.view_vehicle_vessel_jackup_vessel;
DROP VIEW reference.view_vehicle_vessel_multicat;
DROP VIEW reference.view_vehicle_vessel_tugboat;
$$;


SET default_tablespace = '';

SET default_with_oids = false;

--
-- TOC entry 216 (class 1259 OID 30995)
-- Name: bathymetry; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.bathymetry (
    id bigint NOT NULL,
    fk_site_id smallint,
    utm_point public.geometry,
    depth double precision,
    mannings_no double precision
);


--
-- TOC entry 217 (class 1259 OID 31001)
-- Name: bathymetry_layer; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.bathymetry_layer (
    id bigint NOT NULL,
    fk_bathymetry_id bigint,
    fk_soil_type_id integer,
    layer_order smallint,
    initial_depth double precision
);


--
-- TOC entry 218 (class 1259 OID 31004)
-- Name: cable_corridor_bathymetry; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.cable_corridor_bathymetry (
    id bigint NOT NULL,
    fk_site_id integer,
    utm_point public.geometry,
    depth double precision
);


--
-- TOC entry 219 (class 1259 OID 31010)
-- Name: cable_corridor_bathymetry_layer; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.cable_corridor_bathymetry_layer (
    id bigint NOT NULL,
    fk_bathymetry_id bigint,
    fk_soil_type_id integer,
    layer_order smallint,
    initial_depth double precision
);


--
-- TOC entry 220 (class 1259 OID 31013)
-- Name: cable_corridor_constraint; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.cable_corridor_constraint (
    id integer NOT NULL,
    fk_site_id integer,
    description text,
    boundary public.geometry(Polygon)
);


--
-- TOC entry 221 (class 1259 OID 31019)
-- Name: constraint; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter."constraint" (
    id integer NOT NULL,
    fk_site_id integer,
    description text,
    boundary public.geometry(Polygon)
);


--
-- TOC entry 222 (class 1259 OID 31025)
-- Name: device_floating; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.device_floating (
    id integer NOT NULL,
    fk_device_id integer,
    draft double precision,
    maximum_displacement double precision[],
    depth_variation_permitted boolean,
    fairlead_locations double precision[],
    umbilical_connection_point double precision[],
    prescribed_mooring_system character varying(50),
    prescribed_umbilical_type character varying(50)
);


--
-- TOC entry 223 (class 1259 OID 31031)
-- Name: device_shared; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.device_shared (
    id integer NOT NULL,
    fk_device_id integer,
    height double precision,
    width double precision,
    length double precision,
    displaced_volume double precision,
    wet_frontal_area double precision,
    dry_frontal_area double precision,
    wet_beam_area double precision,
    dry_beam_area double precision,
    centre_of_gravity double precision[],
    mass double precision,
    profile character varying(12),
    surface_roughness double precision,
    yaw double precision,
    prescribed_footprint_radius double precision,
    footprint_corner_coords double precision[],
    installation_depth_max double precision,
    installation_depth_min double precision,
    minimum_distance_x double precision,
    minimum_distance_y double precision,
    prescribed_foundation_system character varying(50),
    foundation_locations double precision[],
    rated_power double precision,
    rated_voltage_u0 double precision,
    connector_type character varying(8),
    constant_power_factor double precision,
    variable_power_factor double precision[],
    assembly_duration double precision,
    connect_duration double precision,
    disconnect_duration double precision,
    load_out_method character varying(10),
    transportation_method character varying(4),
    bollard_pull double precision,
    two_stage_assembly boolean,
    cost double precision
);


--
-- TOC entry 224 (class 1259 OID 31037)
-- Name: device_tidal; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.device_tidal (
    id integer NOT NULL,
    fk_device_id integer,
    cut_in_velocity double precision,
    cut_out_velocity double precision,
    hub_height double precision,
    turbine_diameter double precision,
    two_ways_flow boolean,
    turbine_interdistance double precision
);


--
-- TOC entry 225 (class 1259 OID 31040)
-- Name: device_tidal_power_performance; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.device_tidal_power_performance (
    id integer NOT NULL,
    fk_device_id integer,
    velocity double precision NOT NULL,
    thrust_coefficient double precision,
    power_coefficient double precision
);


--
-- TOC entry 226 (class 1259 OID 31043)
-- Name: device_wave; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.device_wave (
    id integer NOT NULL,
    fk_device_id integer,
    wave_data_directory character varying(200)
);


--
-- TOC entry 227 (class 1259 OID 31046)
-- Name: lease_area; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.lease_area (
    id integer NOT NULL,
    fk_site_id integer,
    blockage_ratio double precision,
    tidal_occurrence_point public.geometry(Point),
    wave_spectrum_type character varying(22),
    wave_spectrum_gamma double precision,
    wave_spectrum_spreading_parameter double precision,
    surface_current_flow_velocity double precision,
    current_flow_direction double precision,
    moor_found_current_profile character varying(20),
    significant_wave_height double precision,
    peak_wave_period double precision,
    predominant_wave_direction double precision,
    jonswap_gamma double precision,
    mean_wind_speed double precision,
    predominant_wind_direction double precision,
    max_wind_gust_speed double precision,
    wind_gust_direction double precision,
    water_level_max double precision,
    water_level_min double precision,
    soil_sensitivity double precision,
    has_helipad boolean
);


--
-- TOC entry 228 (class 1259 OID 31052)
-- Name: site_infrastructure; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.site_infrastructure (
    id integer NOT NULL,
    fk_site_id integer,
    has_helipad boolean
);


--
-- TOC entry 229 (class 1259 OID 31055)
-- Name: sub_systems_access; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.sub_systems_access (
    id integer NOT NULL,
    fk_sub_system_id integer,
    operation_duration double precision,
    max_hs double precision,
    max_tp double precision,
    max_ws double precision,
    max_cs double precision
);


--
-- TOC entry 230 (class 1259 OID 31058)
-- Name: sub_systems_economic; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.sub_systems_economic (
    id integer NOT NULL,
    fk_sub_system_id integer,
    cost double precision,
    failure_rate double precision
);


--
-- TOC entry 231 (class 1259 OID 31061)
-- Name: sub_systems_inspection; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.sub_systems_inspection (
    id integer NOT NULL,
    fk_sub_system_id integer,
    operation_duration double precision,
    crew_lead_time double precision,
    other_lead_time double precision,
    n_specialists integer,
    n_technicians integer,
    max_hs double precision,
    max_tp double precision,
    max_ws double precision,
    max_cs double precision
);


--
-- TOC entry 232 (class 1259 OID 31064)
-- Name: sub_systems_install; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.sub_systems_install (
    id integer NOT NULL,
    fk_sub_system_id integer,
    length double precision,
    width double precision,
    height double precision,
    dry_mass double precision,
    max_hs double precision,
    max_tp double precision,
    max_ws double precision,
    max_cs double precision
);


--
-- TOC entry 233 (class 1259 OID 31067)
-- Name: sub_systems_maintenance; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.sub_systems_maintenance (
    id integer NOT NULL,
    fk_sub_system_id integer,
    operation_duration double precision,
    interruptible boolean,
    parts_length double precision,
    parts_width double precision,
    parts_height double precision,
    parts_dry_mass double precision,
    assembly_lead_time double precision,
    crew_lead_time double precision,
    other_lead_time double precision,
    n_specialists integer,
    n_technicians integer,
    max_hs double precision,
    max_tp double precision,
    max_ws double precision,
    max_cs double precision
);


--
-- TOC entry 234 (class 1259 OID 31070)
-- Name: sub_systems_operation_weightings; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.sub_systems_operation_weightings (
    id integer NOT NULL,
    fk_sub_system_id integer,
    maintenance double precision,
    replacement double precision,
    inspection double precision
);


--
-- TOC entry 235 (class 1259 OID 31073)
-- Name: sub_systems_replace; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.sub_systems_replace (
    id integer NOT NULL,
    fk_sub_system_id integer,
    operation_duration double precision,
    interruptible boolean,
    assembly_lead_time double precision,
    crew_lead_time double precision,
    other_lead_time double precision,
    n_specialists integer,
    n_technicians integer
);


--
-- TOC entry 236 (class 1259 OID 31076)
-- Name: time_series_energy_tidal; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.time_series_energy_tidal (
    id bigint NOT NULL,
    fk_bathymetry_id bigint,
    measure_date date,
    measure_time time(6) without time zone,
    u double precision,
    v double precision,
    turbulence_intensity double precision,
    ssh double precision
);


--
-- TOC entry 237 (class 1259 OID 31079)
-- Name: time_series_energy_wave; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.time_series_energy_wave (
    id bigint NOT NULL,
    fk_site_id integer,
    measure_date date,
    measure_time time(6) without time zone,
    height double precision,
    te double precision,
    direction double precision
);


--
-- TOC entry 238 (class 1259 OID 31082)
-- Name: time_series_om_tidal; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.time_series_om_tidal (
    id bigint NOT NULL,
    fk_site_id bigint,
    measure_date date,
    measure_time time(6) without time zone,
    current_speed double precision
);


--
-- TOC entry 239 (class 1259 OID 31085)
-- Name: time_series_om_wave; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.time_series_om_wave (
    id bigint NOT NULL,
    fk_site_id integer,
    measure_date date,
    measure_time time(6) without time zone,
    period_tp double precision,
    height_hs double precision
);


--
-- TOC entry 240 (class 1259 OID 31088)
-- Name: time_series_om_wind; Type: TABLE; Schema: filter; Owner: -
--

CREATE TABLE filter.time_series_om_wind (
    id bigint NOT NULL,
    fk_site_id integer,
    measure_date date,
    measure_time time(6) without time zone,
    wind_speed double precision
);


--
-- TOC entry 241 (class 1259 OID 31091)
-- Name: soil_type; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.soil_type (
    id integer NOT NULL,
    description character varying(40)
);


--
-- TOC entry 242 (class 1259 OID 31094)
-- Name: view_bathymetry_layer; Type: VIEW; Schema: filter; Owner: -
--

CREATE VIEW filter.view_bathymetry_layer AS
 SELECT bathymetry.utm_point,
    bathymetry.depth,
    bathymetry.mannings_no,
    bathymetry_layer.layer_order,
    bathymetry_layer.initial_depth,
    soil_type.description AS sediment_type
   FROM ((filter.bathymetry
     LEFT JOIN filter.bathymetry_layer ON ((bathymetry.id = bathymetry_layer.fk_bathymetry_id)))
     LEFT JOIN reference.soil_type ON ((bathymetry_layer.fk_soil_type_id = soil_type.id)));


--
-- TOC entry 243 (class 1259 OID 31098)
-- Name: view_cable_corridor_bathymetry_layer; Type: VIEW; Schema: filter; Owner: -
--

CREATE VIEW filter.view_cable_corridor_bathymetry_layer AS
 SELECT cable_corridor_bathymetry.utm_point,
    cable_corridor_bathymetry.depth,
    cable_corridor_bathymetry_layer.layer_order,
    cable_corridor_bathymetry_layer.initial_depth,
    soil_type.description AS sediment_type
   FROM ((filter.cable_corridor_bathymetry
     LEFT JOIN filter.cable_corridor_bathymetry_layer ON ((cable_corridor_bathymetry.id = cable_corridor_bathymetry_layer.fk_bathymetry_id)))
     LEFT JOIN reference.soil_type ON ((cable_corridor_bathymetry_layer.fk_soil_type_id = soil_type.id)));


--
-- TOC entry 244 (class 1259 OID 31102)
-- Name: sub_systems; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.sub_systems (
    id integer NOT NULL,
    fk_device_id integer,
    sub_system character varying(20),
    CONSTRAINT sub_systems_sub_system_check CHECK (((sub_system)::text = ANY (ARRAY[('Control System'::character varying)::text, ('Prime Mover'::character varying)::text, ('PTO'::character varying)::text, ('Support Structure'::character varying)::text])))
);


--
-- TOC entry 245 (class 1259 OID 31106)
-- Name: view_control_system_access; Type: VIEW; Schema: filter; Owner: -
--

CREATE VIEW filter.view_control_system_access AS
 SELECT sub_systems.sub_system,
    sub_systems_access.operation_duration,
    sub_systems_access.max_hs,
    sub_systems_access.max_tp,
    sub_systems_access.max_ws,
    sub_systems_access.max_cs
   FROM (filter.sub_systems_access
     JOIN project.sub_systems ON ((sub_systems_access.fk_sub_system_id = sub_systems.id)))
  WHERE ((sub_systems.sub_system)::text = 'Control System'::text);


--
-- TOC entry 246 (class 1259 OID 31110)
-- Name: view_control_system_economic; Type: VIEW; Schema: filter; Owner: -
--

CREATE VIEW filter.view_control_system_economic AS
 SELECT sub_systems.sub_system,
    sub_systems_economic.cost,
    sub_systems_economic.failure_rate
   FROM (filter.sub_systems_economic
     JOIN project.sub_systems ON ((sub_systems_economic.fk_sub_system_id = sub_systems.id)))
  WHERE ((sub_systems.sub_system)::text = 'Control System'::text);


--
-- TOC entry 247 (class 1259 OID 31114)
-- Name: view_control_system_inspection; Type: VIEW; Schema: filter; Owner: -
--

CREATE VIEW filter.view_control_system_inspection AS
 SELECT sub_systems.sub_system,
    sub_systems_inspection.operation_duration,
    sub_systems_inspection.crew_lead_time,
    sub_systems_inspection.other_lead_time,
    sub_systems_inspection.n_specialists,
    sub_systems_inspection.n_technicians,
    sub_systems_inspection.max_hs,
    sub_systems_inspection.max_tp,
    sub_systems_inspection.max_ws,
    sub_systems_inspection.max_cs
   FROM (filter.sub_systems_inspection
     JOIN project.sub_systems ON ((sub_systems_inspection.fk_sub_system_id = sub_systems.id)))
  WHERE ((sub_systems.sub_system)::text = 'Control System'::text);


--
-- TOC entry 248 (class 1259 OID 31118)
-- Name: view_control_system_installation; Type: VIEW; Schema: filter; Owner: -
--

CREATE VIEW filter.view_control_system_installation AS
 SELECT sub_systems.sub_system,
    sub_systems_install.length,
    sub_systems_install.width,
    sub_systems_install.height,
    sub_systems_install.dry_mass,
    sub_systems_install.max_hs,
    sub_systems_install.max_tp,
    sub_systems_install.max_ws,
    sub_systems_install.max_cs
   FROM (filter.sub_systems_install
     JOIN project.sub_systems ON ((sub_systems_install.fk_sub_system_id = sub_systems.id)))
  WHERE ((sub_systems.sub_system)::text = 'Control System'::text);


--
-- TOC entry 249 (class 1259 OID 31122)
-- Name: view_control_system_maintenance; Type: VIEW; Schema: filter; Owner: -
--

CREATE VIEW filter.view_control_system_maintenance AS
 SELECT sub_systems.sub_system,
    sub_systems_maintenance.operation_duration,
    sub_systems_maintenance.interruptible,
    sub_systems_maintenance.parts_length,
    sub_systems_maintenance.parts_width,
    sub_systems_maintenance.parts_height,
    sub_systems_maintenance.parts_dry_mass,
    sub_systems_maintenance.assembly_lead_time,
    sub_systems_maintenance.crew_lead_time,
    sub_systems_maintenance.other_lead_time,
    sub_systems_maintenance.n_specialists,
    sub_systems_maintenance.n_technicians,
    sub_systems_maintenance.max_hs,
    sub_systems_maintenance.max_tp,
    sub_systems_maintenance.max_ws,
    sub_systems_maintenance.max_cs
   FROM (filter.sub_systems_maintenance
     JOIN project.sub_systems ON ((sub_systems_maintenance.fk_sub_system_id = sub_systems.id)))
  WHERE ((sub_systems.sub_system)::text = 'Control System'::text);


--
-- TOC entry 250 (class 1259 OID 31127)
-- Name: view_control_system_operation_weightings; Type: VIEW; Schema: filter; Owner: -
--

CREATE VIEW filter.view_control_system_operation_weightings AS
 SELECT sub_systems.sub_system,
    sub_systems_operation_weightings.maintenance,
    sub_systems_operation_weightings.replacement,
    sub_systems_operation_weightings.inspection
   FROM (filter.sub_systems_operation_weightings
     JOIN project.sub_systems ON ((sub_systems_operation_weightings.fk_sub_system_id = sub_systems.id)))
  WHERE ((sub_systems.sub_system)::text = 'Control System'::text);


--
-- TOC entry 251 (class 1259 OID 31131)
-- Name: view_control_system_replace; Type: VIEW; Schema: filter; Owner: -
--

CREATE VIEW filter.view_control_system_replace AS
 SELECT sub_systems.sub_system,
    sub_systems_replace.operation_duration,
    sub_systems_replace.interruptible,
    sub_systems_replace.assembly_lead_time,
    sub_systems_replace.crew_lead_time,
    sub_systems_replace.other_lead_time,
    sub_systems_replace.n_specialists,
    sub_systems_replace.n_technicians
   FROM (filter.sub_systems_replace
     JOIN project.sub_systems ON ((sub_systems_replace.fk_sub_system_id = sub_systems.id)))
  WHERE ((sub_systems.sub_system)::text = 'Control System'::text);


--
-- TOC entry 252 (class 1259 OID 31135)
-- Name: view_sub_systems_access; Type: VIEW; Schema: filter; Owner: -
--

CREATE VIEW filter.view_sub_systems_access AS
 SELECT sub_systems.sub_system,
    sub_systems_access.operation_duration,
    sub_systems_access.max_hs,
    sub_systems_access.max_tp,
    sub_systems_access.max_ws,
    sub_systems_access.max_cs
   FROM (filter.sub_systems_access
     JOIN project.sub_systems ON ((sub_systems_access.fk_sub_system_id = sub_systems.id)))
  WHERE ((sub_systems.sub_system)::text = ANY (ARRAY[('Prime Mover'::character varying)::text, ('PTO'::character varying)::text, ('Support Structure'::character varying)::text]));


--
-- TOC entry 253 (class 1259 OID 31139)
-- Name: view_sub_systems_economic; Type: VIEW; Schema: filter; Owner: -
--

CREATE VIEW filter.view_sub_systems_economic AS
 SELECT sub_systems.sub_system,
    sub_systems_economic.cost,
    sub_systems_economic.failure_rate
   FROM (filter.sub_systems_economic
     JOIN project.sub_systems ON ((sub_systems_economic.fk_sub_system_id = sub_systems.id)))
  WHERE ((sub_systems.sub_system)::text = ANY (ARRAY[('Prime Mover'::character varying)::text, ('PTO'::character varying)::text, ('Support Structure'::character varying)::text]));


--
-- TOC entry 254 (class 1259 OID 31143)
-- Name: view_sub_systems_inspection; Type: VIEW; Schema: filter; Owner: -
--

CREATE VIEW filter.view_sub_systems_inspection AS
 SELECT sub_systems.sub_system,
    sub_systems_inspection.operation_duration,
    sub_systems_inspection.crew_lead_time,
    sub_systems_inspection.other_lead_time,
    sub_systems_inspection.n_specialists,
    sub_systems_inspection.n_technicians,
    sub_systems_inspection.max_hs,
    sub_systems_inspection.max_tp,
    sub_systems_inspection.max_ws,
    sub_systems_inspection.max_cs
   FROM (filter.sub_systems_inspection
     JOIN project.sub_systems ON ((sub_systems_inspection.fk_sub_system_id = sub_systems.id)))
  WHERE ((sub_systems.sub_system)::text = ANY (ARRAY[('Prime Mover'::character varying)::text, ('PTO'::character varying)::text, ('Support Structure'::character varying)::text]));


--
-- TOC entry 255 (class 1259 OID 31148)
-- Name: view_sub_systems_installation; Type: VIEW; Schema: filter; Owner: -
--

CREATE VIEW filter.view_sub_systems_installation AS
 SELECT sub_systems.sub_system,
    sub_systems_install.length,
    sub_systems_install.width,
    sub_systems_install.height,
    sub_systems_install.dry_mass,
    sub_systems_install.max_hs,
    sub_systems_install.max_tp,
    sub_systems_install.max_ws,
    sub_systems_install.max_cs
   FROM (filter.sub_systems_install
     JOIN project.sub_systems ON ((sub_systems_install.fk_sub_system_id = sub_systems.id)))
  WHERE ((sub_systems.sub_system)::text = ANY (ARRAY[('Prime Mover'::character varying)::text, ('PTO'::character varying)::text, ('Support Structure'::character varying)::text]));


--
-- TOC entry 256 (class 1259 OID 31152)
-- Name: view_sub_systems_maintenance; Type: VIEW; Schema: filter; Owner: -
--

CREATE VIEW filter.view_sub_systems_maintenance AS
 SELECT sub_systems.sub_system,
    sub_systems_maintenance.operation_duration,
    sub_systems_maintenance.interruptible,
    sub_systems_maintenance.parts_length,
    sub_systems_maintenance.parts_width,
    sub_systems_maintenance.parts_height,
    sub_systems_maintenance.parts_dry_mass,
    sub_systems_maintenance.assembly_lead_time,
    sub_systems_maintenance.crew_lead_time,
    sub_systems_maintenance.other_lead_time,
    sub_systems_maintenance.n_specialists,
    sub_systems_maintenance.n_technicians,
    sub_systems_maintenance.max_hs,
    sub_systems_maintenance.max_tp,
    sub_systems_maintenance.max_ws,
    sub_systems_maintenance.max_cs
   FROM (filter.sub_systems_maintenance
     JOIN project.sub_systems ON ((sub_systems_maintenance.fk_sub_system_id = sub_systems.id)))
  WHERE ((sub_systems.sub_system)::text = ANY (ARRAY[('Prime Mover'::character varying)::text, ('PTO'::character varying)::text, ('Support Structure'::character varying)::text]));


--
-- TOC entry 257 (class 1259 OID 31157)
-- Name: view_sub_systems_operation_weightings; Type: VIEW; Schema: filter; Owner: -
--

CREATE VIEW filter.view_sub_systems_operation_weightings AS
 SELECT sub_systems.sub_system,
    sub_systems_operation_weightings.maintenance,
    sub_systems_operation_weightings.replacement,
    sub_systems_operation_weightings.inspection
   FROM (filter.sub_systems_operation_weightings
     JOIN project.sub_systems ON ((sub_systems_operation_weightings.fk_sub_system_id = sub_systems.id)))
  WHERE ((sub_systems.sub_system)::text = ANY (ARRAY[('Prime Mover'::character varying)::text, ('PTO'::character varying)::text, ('Support Structure'::character varying)::text]));


--
-- TOC entry 258 (class 1259 OID 31161)
-- Name: view_sub_systems_replace; Type: VIEW; Schema: filter; Owner: -
--

CREATE VIEW filter.view_sub_systems_replace AS
 SELECT sub_systems.sub_system,
    sub_systems_replace.operation_duration,
    sub_systems_replace.interruptible,
    sub_systems_replace.assembly_lead_time,
    sub_systems_replace.crew_lead_time,
    sub_systems_replace.other_lead_time,
    sub_systems_replace.n_specialists,
    sub_systems_replace.n_technicians
   FROM (filter.sub_systems_replace
     JOIN project.sub_systems ON ((sub_systems_replace.fk_sub_system_id = sub_systems.id)))
  WHERE ((sub_systems.sub_system)::text = ANY (ARRAY[('Prime Mover'::character varying)::text, ('PTO'::character varying)::text, ('Support Structure'::character varying)::text]));


--
-- TOC entry 259 (class 1259 OID 31166)
-- Name: view_time_series_energy_tidal; Type: VIEW; Schema: filter; Owner: -
--

CREATE VIEW filter.view_time_series_energy_tidal AS
 SELECT bathymetry.utm_point,
    time_series_energy_tidal.measure_date,
    time_series_energy_tidal.measure_time,
    time_series_energy_tidal.u,
    time_series_energy_tidal.v,
    time_series_energy_tidal.turbulence_intensity,
    time_series_energy_tidal.ssh
   FROM (filter.bathymetry
     JOIN filter.time_series_energy_tidal ON ((bathymetry.id = time_series_energy_tidal.fk_bathymetry_id)));


--
-- TOC entry 260 (class 1259 OID 31170)
-- Name: bathymetry; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.bathymetry (
    id bigint NOT NULL,
    fk_site_id smallint,
    utm_point public.geometry,
    depth double precision,
    mannings_no double precision
);


--
-- TOC entry 261 (class 1259 OID 31176)
-- Name: bathymetry_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.bathymetry_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5581 (class 0 OID 0)
-- Dependencies: 261
-- Name: bathymetry_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.bathymetry_id_seq OWNED BY project.bathymetry.id;


--
-- TOC entry 262 (class 1259 OID 31178)
-- Name: bathymetry_layer_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.bathymetry_layer_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 263 (class 1259 OID 31180)
-- Name: bathymetry_layer; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.bathymetry_layer (
    id bigint DEFAULT nextval('project.bathymetry_layer_id_seq'::regclass) NOT NULL,
    fk_bathymetry_id bigint,
    fk_soil_type_id integer,
    layer_order smallint,
    initial_depth double precision
);


--
-- TOC entry 264 (class 1259 OID 31184)
-- Name: cable_corridor_bathymetry; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.cable_corridor_bathymetry (
    id bigint NOT NULL,
    fk_site_id integer,
    utm_point public.geometry,
    depth double precision
);


--
-- TOC entry 265 (class 1259 OID 31190)
-- Name: cable_corridor_bathymetry_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.cable_corridor_bathymetry_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5582 (class 0 OID 0)
-- Dependencies: 265
-- Name: cable_corridor_bathymetry_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.cable_corridor_bathymetry_id_seq OWNED BY project.cable_corridor_bathymetry.id;


--
-- TOC entry 266 (class 1259 OID 31192)
-- Name: cable_corridor_bathymetry_layer; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.cable_corridor_bathymetry_layer (
    id bigint NOT NULL,
    fk_bathymetry_id bigint,
    fk_soil_type_id integer,
    layer_order smallint,
    initial_depth double precision
);


--
-- TOC entry 267 (class 1259 OID 31195)
-- Name: cable_corridor_bathymetry_layer_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.cable_corridor_bathymetry_layer_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5583 (class 0 OID 0)
-- Dependencies: 267
-- Name: cable_corridor_bathymetry_layer_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.cable_corridor_bathymetry_layer_id_seq OWNED BY project.cable_corridor_bathymetry_layer.id;


--
-- TOC entry 268 (class 1259 OID 31197)
-- Name: cable_corridor_constraint; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.cable_corridor_constraint (
    id integer NOT NULL,
    fk_site_id integer,
    description text,
    boundary public.geometry(Polygon)
);


--
-- TOC entry 269 (class 1259 OID 31203)
-- Name: cable_corridor_constraint_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.cable_corridor_constraint_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5584 (class 0 OID 0)
-- Dependencies: 269
-- Name: cable_corridor_constraint_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.cable_corridor_constraint_id_seq OWNED BY project.cable_corridor_constraint.id;


--
-- TOC entry 270 (class 1259 OID 31205)
-- Name: constraint; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project."constraint" (
    id integer NOT NULL,
    fk_site_id integer,
    description text,
    boundary public.geometry(Polygon)
);


--
-- TOC entry 271 (class 1259 OID 31211)
-- Name: constraint_type_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.constraint_type_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 32767
    CACHE 1;


--
-- TOC entry 272 (class 1259 OID 31213)
-- Name: device; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.device (
    id integer NOT NULL,
    description character varying(200),
    device_type character varying(50),
    image bytea,
    CONSTRAINT device_device_type_check CHECK (((device_type)::text = ANY (ARRAY[('Tidal Fixed'::character varying)::text, ('Tidal Floating'::character varying)::text, ('Wave Fixed'::character varying)::text, ('Wave Floating'::character varying)::text])))
);


--
-- TOC entry 273 (class 1259 OID 31220)
-- Name: device_floating; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.device_floating (
    id integer NOT NULL,
    fk_device_id integer,
    draft double precision,
    maximum_displacement double precision[],
    depth_variation_permitted boolean,
    fairlead_locations double precision[],
    umbilical_connection_point double precision[],
    prescribed_mooring_system character varying(50),
    prescribed_umbilical_type character varying(50)
);


--
-- TOC entry 274 (class 1259 OID 31226)
-- Name: device_floating_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.device_floating_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5585 (class 0 OID 0)
-- Dependencies: 274
-- Name: device_floating_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.device_floating_id_seq OWNED BY project.device_floating.id;


--
-- TOC entry 275 (class 1259 OID 31228)
-- Name: device_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.device_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5586 (class 0 OID 0)
-- Dependencies: 275
-- Name: device_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.device_id_seq OWNED BY project.device.id;


--
-- TOC entry 276 (class 1259 OID 31230)
-- Name: device_power_performance_tidal_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.device_power_performance_tidal_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 2147483647
    CACHE 1;


--
-- TOC entry 277 (class 1259 OID 31232)
-- Name: device_shared; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.device_shared (
    id integer NOT NULL,
    fk_device_id integer,
    height double precision,
    width double precision,
    length double precision,
    displaced_volume double precision,
    wet_frontal_area double precision,
    dry_frontal_area double precision,
    wet_beam_area double precision,
    dry_beam_area double precision,
    centre_of_gravity double precision[],
    mass double precision,
    profile character varying(12),
    surface_roughness double precision,
    yaw double precision,
    prescribed_footprint_radius double precision,
    footprint_corner_coords double precision[],
    installation_depth_max double precision,
    installation_depth_min double precision,
    minimum_distance_x double precision,
    minimum_distance_y double precision,
    prescribed_foundation_system character varying(50),
    foundation_locations double precision[],
    rated_power double precision,
    rated_voltage_u0 double precision,
    connector_type character varying(8),
    constant_power_factor double precision,
    variable_power_factor double precision[],
    assembly_duration double precision,
    connect_duration double precision,
    disconnect_duration double precision,
    load_out_method character varying(10),
    transportation_method character varying(4),
    bollard_pull double precision,
    two_stage_assembly boolean,
    cost double precision,
    CONSTRAINT device_shared_connector_type_check CHECK (((connector_type)::text = ANY (ARRAY[('Wet-Mate'::character varying)::text, ('Dry-Mate'::character varying)::text]))),
    CONSTRAINT device_shared_load_out_method_check CHECK (((load_out_method)::text = ANY (ARRAY[('Skidded'::character varying)::text, ('Trailer'::character varying)::text, ('Float Away'::character varying)::text, ('Lift Away'::character varying)::text]))),
    CONSTRAINT device_shared_profile_check CHECK (((profile)::text = ANY (ARRAY[('Cylindrical'::character varying)::text, ('Rectangular'::character varying)::text]))),
    CONSTRAINT device_shared_transportation_method_check CHECK (((transportation_method)::text = ANY (ARRAY[('Deck'::character varying)::text, ('Tow'::character varying)::text])))
);


--
-- TOC entry 278 (class 1259 OID 31242)
-- Name: device_shared_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.device_shared_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5587 (class 0 OID 0)
-- Dependencies: 278
-- Name: device_shared_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.device_shared_id_seq OWNED BY project.device_shared.id;


--
-- TOC entry 279 (class 1259 OID 31244)
-- Name: device_tidal; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.device_tidal (
    id integer NOT NULL,
    fk_device_id integer,
    cut_in_velocity double precision,
    cut_out_velocity double precision,
    hub_height double precision,
    turbine_diameter double precision,
    two_ways_flow boolean,
    turbine_interdistance double precision
);


--
-- TOC entry 280 (class 1259 OID 31247)
-- Name: device_tidal_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.device_tidal_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5588 (class 0 OID 0)
-- Dependencies: 280
-- Name: device_tidal_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.device_tidal_id_seq OWNED BY project.device_tidal.id;


--
-- TOC entry 281 (class 1259 OID 31249)
-- Name: device_tidal_power_performance; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.device_tidal_power_performance (
    id integer NOT NULL,
    fk_device_id integer,
    velocity double precision NOT NULL,
    thrust_coefficient double precision,
    power_coefficient double precision
);


--
-- TOC entry 282 (class 1259 OID 31252)
-- Name: device_tidal_power_performance_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.device_tidal_power_performance_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5589 (class 0 OID 0)
-- Dependencies: 282
-- Name: device_tidal_power_performance_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.device_tidal_power_performance_id_seq OWNED BY project.device_tidal_power_performance.id;


--
-- TOC entry 283 (class 1259 OID 31254)
-- Name: device_wave; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.device_wave (
    id integer NOT NULL,
    fk_device_id integer,
    wave_data_directory character varying(200)
);


--
-- TOC entry 284 (class 1259 OID 31257)
-- Name: device_wave_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.device_wave_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5590 (class 0 OID 0)
-- Dependencies: 284
-- Name: device_wave_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.device_wave_id_seq OWNED BY project.device_wave.id;


--
-- TOC entry 285 (class 1259 OID 31259)
-- Name: lease_area; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.lease_area (
    id integer NOT NULL,
    fk_site_id integer,
    blockage_ratio double precision,
    tidal_occurrence_point public.geometry(Point),
    wave_spectrum_type character varying(22),
    wave_spectrum_gamma double precision,
    wave_spectrum_spreading_parameter double precision,
    surface_current_flow_velocity double precision,
    current_flow_direction double precision,
    moor_found_current_profile character varying(20),
    significant_wave_height double precision,
    peak_wave_period double precision,
    predominant_wave_direction double precision,
    jonswap_gamma double precision,
    mean_wind_speed double precision,
    predominant_wind_direction double precision,
    max_wind_gust_speed double precision,
    wind_gust_direction double precision,
    water_level_max double precision,
    water_level_min double precision,
    soil_sensitivity double precision,
    has_helipad boolean,
    CONSTRAINT lease_area_moor_found_current_profile_check CHECK (((moor_found_current_profile)::text = ANY (ARRAY[('Uniform'::character varying)::text, ('1/7 Power Law'::character varying)::text]))),
    CONSTRAINT lease_area_wave_spectrum_type_check CHECK (((wave_spectrum_type)::text = ANY (ARRAY[('Regular'::character varying)::text, ('Pierson-Moskowitz'::character varying)::text, ('JONSWAP'::character varying)::text, ('Bretschneider'::character varying)::text, ('Modified Bretschneider'::character varying)::text])))
);


--
-- TOC entry 286 (class 1259 OID 31267)
-- Name: lease_area_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.lease_area_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5591 (class 0 OID 0)
-- Dependencies: 286
-- Name: lease_area_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.lease_area_id_seq OWNED BY project.lease_area.id;


--
-- TOC entry 287 (class 1259 OID 31269)
-- Name: site; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.site (
    id integer NOT NULL,
    site_name character varying(20),
    lease_area_proj4_string character varying(100),
    site_boundary public.geometry(Polygon,4326),
    lease_boundary public.geometry(Polygon),
    corridor_boundary public.geometry(Polygon),
    cable_landing_location public.geometry(Point)
);


--
-- TOC entry 288 (class 1259 OID 31275)
-- Name: site_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.site_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5592 (class 0 OID 0)
-- Dependencies: 288
-- Name: site_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.site_id_seq OWNED BY project.site.id;


--
-- TOC entry 289 (class 1259 OID 31277)
-- Name: sub_systems_access; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.sub_systems_access (
    id integer NOT NULL,
    fk_sub_system_id integer,
    operation_duration double precision,
    max_hs double precision,
    max_tp double precision,
    max_ws double precision,
    max_cs double precision
);


--
-- TOC entry 290 (class 1259 OID 31280)
-- Name: sub_systems_access_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.sub_systems_access_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5593 (class 0 OID 0)
-- Dependencies: 290
-- Name: sub_systems_access_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.sub_systems_access_id_seq OWNED BY project.sub_systems_access.id;


--
-- TOC entry 291 (class 1259 OID 31282)
-- Name: sub_systems_economic; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.sub_systems_economic (
    id integer NOT NULL,
    fk_sub_system_id integer,
    cost double precision,
    failure_rate double precision
);


--
-- TOC entry 292 (class 1259 OID 31285)
-- Name: sub_systems_economic_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.sub_systems_economic_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5594 (class 0 OID 0)
-- Dependencies: 292
-- Name: sub_systems_economic_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.sub_systems_economic_id_seq OWNED BY project.sub_systems_economic.id;


--
-- TOC entry 293 (class 1259 OID 31287)
-- Name: sub_systems_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.sub_systems_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5595 (class 0 OID 0)
-- Dependencies: 293
-- Name: sub_systems_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.sub_systems_id_seq OWNED BY project.sub_systems.id;


--
-- TOC entry 294 (class 1259 OID 31289)
-- Name: sub_systems_inspection; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.sub_systems_inspection (
    id integer NOT NULL,
    fk_sub_system_id integer,
    operation_duration double precision,
    crew_lead_time double precision,
    other_lead_time double precision,
    n_specialists integer,
    n_technicians integer,
    max_hs double precision,
    max_tp double precision,
    max_ws double precision,
    max_cs double precision
);


--
-- TOC entry 295 (class 1259 OID 31292)
-- Name: sub_systems_inspection_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.sub_systems_inspection_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5596 (class 0 OID 0)
-- Dependencies: 295
-- Name: sub_systems_inspection_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.sub_systems_inspection_id_seq OWNED BY project.sub_systems_inspection.id;


--
-- TOC entry 296 (class 1259 OID 31294)
-- Name: sub_systems_install; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.sub_systems_install (
    id integer NOT NULL,
    fk_sub_system_id integer,
    length double precision,
    width double precision,
    height double precision,
    dry_mass double precision,
    max_hs double precision,
    max_tp double precision,
    max_ws double precision,
    max_cs double precision
);


--
-- TOC entry 297 (class 1259 OID 31297)
-- Name: sub_systems_install_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.sub_systems_install_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5597 (class 0 OID 0)
-- Dependencies: 297
-- Name: sub_systems_install_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.sub_systems_install_id_seq OWNED BY project.sub_systems_install.id;


--
-- TOC entry 298 (class 1259 OID 31299)
-- Name: sub_systems_maintenance; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.sub_systems_maintenance (
    id integer NOT NULL,
    fk_sub_system_id integer,
    operation_duration double precision,
    interruptible boolean,
    parts_length double precision,
    parts_width double precision,
    parts_height double precision,
    parts_dry_mass double precision,
    assembly_lead_time double precision,
    crew_lead_time double precision,
    other_lead_time double precision,
    n_specialists integer,
    n_technicians integer,
    max_hs double precision,
    max_tp double precision,
    max_ws double precision,
    max_cs double precision
);


--
-- TOC entry 299 (class 1259 OID 31302)
-- Name: sub_systems_maintenance_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.sub_systems_maintenance_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5598 (class 0 OID 0)
-- Dependencies: 299
-- Name: sub_systems_maintenance_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.sub_systems_maintenance_id_seq OWNED BY project.sub_systems_maintenance.id;


--
-- TOC entry 300 (class 1259 OID 31304)
-- Name: sub_systems_operation_weightings; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.sub_systems_operation_weightings (
    id integer NOT NULL,
    fk_sub_system_id integer,
    maintenance double precision,
    replacement double precision,
    inspection double precision
);


--
-- TOC entry 301 (class 1259 OID 31307)
-- Name: sub_systems_operation_weightings_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.sub_systems_operation_weightings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5599 (class 0 OID 0)
-- Dependencies: 301
-- Name: sub_systems_operation_weightings_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.sub_systems_operation_weightings_id_seq OWNED BY project.sub_systems_operation_weightings.id;


--
-- TOC entry 302 (class 1259 OID 31309)
-- Name: sub_systems_replace; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.sub_systems_replace (
    id integer NOT NULL,
    fk_sub_system_id integer,
    operation_duration double precision,
    interruptible boolean,
    assembly_lead_time double precision,
    crew_lead_time double precision,
    other_lead_time double precision,
    n_specialists integer,
    n_technicians integer
);


--
-- TOC entry 303 (class 1259 OID 31312)
-- Name: sub_systems_replace_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.sub_systems_replace_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5600 (class 0 OID 0)
-- Dependencies: 303
-- Name: sub_systems_replace_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.sub_systems_replace_id_seq OWNED BY project.sub_systems_replace.id;


--
-- TOC entry 304 (class 1259 OID 31314)
-- Name: time_series_energy_tidal; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.time_series_energy_tidal (
    id bigint NOT NULL,
    fk_bathymetry_id bigint,
    measure_date date,
    measure_time time(6) without time zone,
    u double precision,
    v double precision,
    turbulence_intensity double precision,
    ssh double precision
);


--
-- TOC entry 305 (class 1259 OID 31317)
-- Name: time_series_energy_tidal_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.time_series_energy_tidal_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5601 (class 0 OID 0)
-- Dependencies: 305
-- Name: time_series_energy_tidal_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.time_series_energy_tidal_id_seq OWNED BY project.time_series_energy_tidal.id;


--
-- TOC entry 306 (class 1259 OID 31319)
-- Name: time_series_energy_wave; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.time_series_energy_wave (
    id bigint NOT NULL,
    fk_site_id integer,
    measure_date date,
    measure_time time(6) without time zone,
    height double precision,
    te double precision,
    direction double precision
);


--
-- TOC entry 307 (class 1259 OID 31322)
-- Name: time_series_energy_wave_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.time_series_energy_wave_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5602 (class 0 OID 0)
-- Dependencies: 307
-- Name: time_series_energy_wave_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.time_series_energy_wave_id_seq OWNED BY project.time_series_energy_wave.id;


--
-- TOC entry 308 (class 1259 OID 31324)
-- Name: time_series_om_tidal; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.time_series_om_tidal (
    id bigint NOT NULL,
    fk_site_id bigint,
    measure_date date,
    measure_time time(6) without time zone,
    current_speed double precision
);


--
-- TOC entry 309 (class 1259 OID 31327)
-- Name: time_series_om_tidal_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.time_series_om_tidal_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5603 (class 0 OID 0)
-- Dependencies: 309
-- Name: time_series_om_tidal_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.time_series_om_tidal_id_seq OWNED BY project.time_series_om_tidal.id;


--
-- TOC entry 310 (class 1259 OID 31329)
-- Name: time_series_om_wave; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.time_series_om_wave (
    id bigint NOT NULL,
    fk_site_id integer,
    measure_date date,
    measure_time time(6) without time zone,
    period_tp double precision,
    height_hs double precision
);


--
-- TOC entry 311 (class 1259 OID 31332)
-- Name: time_series_om_wave_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.time_series_om_wave_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5604 (class 0 OID 0)
-- Dependencies: 311
-- Name: time_series_om_wave_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.time_series_om_wave_id_seq OWNED BY project.time_series_om_wave.id;


--
-- TOC entry 312 (class 1259 OID 31334)
-- Name: time_series_om_wind; Type: TABLE; Schema: project; Owner: -
--

CREATE TABLE project.time_series_om_wind (
    id bigint NOT NULL,
    fk_site_id integer,
    measure_date date,
    measure_time time(6) without time zone,
    wind_speed double precision
);


--
-- TOC entry 313 (class 1259 OID 31337)
-- Name: time_series_om_wind_id_seq; Type: SEQUENCE; Schema: project; Owner: -
--

CREATE SEQUENCE project.time_series_om_wind_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5605 (class 0 OID 0)
-- Dependencies: 313
-- Name: time_series_om_wind_id_seq; Type: SEQUENCE OWNED BY; Schema: project; Owner: -
--

ALTER SEQUENCE project.time_series_om_wind_id_seq OWNED BY project.time_series_om_wind.id;


--
-- TOC entry 314 (class 1259 OID 31339)
-- Name: component; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.component (
    id bigint NOT NULL,
    description character varying(200)
);


--
-- TOC entry 315 (class 1259 OID 31342)
-- Name: component_anchor; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.component_anchor (
    id integer NOT NULL,
    fk_component_discrete_id bigint,
    fk_component_type_id smallint,
    connecting_size double precision,
    minimum_breaking_load double precision,
    axial_stiffness double precision,
    soft_holding_cap_coef_1 double precision,
    soft_holding_cap_coef_2 double precision,
    soft_penetration_coef_1 double precision,
    soft_penetration_coef_2 double precision,
    sand_holding_cap_coef_1 double precision,
    sand_holding_cap_coef_2 double precision,
    sand_penetration_coef_1 double precision,
    sand_penetration_coef_2 double precision,
    CONSTRAINT component_anchor_fk_component_type_id_check CHECK ((fk_component_type_id = 1))
);


--
-- TOC entry 316 (class 1259 OID 31346)
-- Name: component_anchor_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.component_anchor_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5606 (class 0 OID 0)
-- Dependencies: 316
-- Name: component_anchor_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.component_anchor_id_seq OWNED BY reference.component_anchor.id;


--
-- TOC entry 317 (class 1259 OID 31348)
-- Name: component_cable; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.component_cable (
    id integer NOT NULL,
    fk_component_continuous_id bigint,
    fk_component_type_id smallint,
    minimum_breaking_load double precision,
    minimum_bend_radius double precision,
    number_conductors smallint,
    number_fibre_channels smallint,
    resistance_dc_20 double precision,
    resistance_ac_90 double precision,
    inductive_reactance double precision,
    capacitance double precision,
    rated_current_air double precision,
    rated_current_buried double precision,
    rated_current_jtube double precision,
    rated_voltage_u0 double precision,
    operational_temp_max double precision,
    CONSTRAINT component_cable_fk_component_type_id_check CHECK ((fk_component_type_id = ANY (ARRAY[2, 3])))
);


--
-- TOC entry 318 (class 1259 OID 31352)
-- Name: component_cable_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.component_cable_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5607 (class 0 OID 0)
-- Dependencies: 318
-- Name: component_cable_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.component_cable_id_seq OWNED BY reference.component_cable.id;


--
-- TOC entry 319 (class 1259 OID 31354)
-- Name: component_collection_point; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.component_collection_point (
    id integer NOT NULL,
    fk_component_discrete_id bigint,
    fk_component_type_id smallint,
    wet_frontal_area double precision,
    dry_frontal_area double precision,
    wet_beam_area double precision,
    dry_beam_area double precision,
    maximum_water_depth double precision,
    orientation_angle double precision,
    input_lines integer,
    output_lines integer,
    input_connector_type character varying(8),
    output_connector_type character varying(8),
    number_fibre_channels integer,
    voltage_primary_winding double precision,
    voltage_secondary_winding double precision,
    rated_operating_current double precision,
    operational_temp_min double precision,
    operational_temp_max double precision,
    foundation_locations double precision[],
    centre_of_gravity double precision[],
    CONSTRAINT component_collection_point_fk_component_type_id_check CHECK ((fk_component_type_id = 5)),
    CONSTRAINT component_collection_point_input_connector_type_check CHECK (((input_connector_type)::text = ANY (ARRAY[('wet-mate'::character varying)::text, ('dry-mate'::character varying)::text]))),
    CONSTRAINT component_collection_point_output_connector_type_check CHECK (((output_connector_type)::text = ANY (ARRAY[('wet-mate'::character varying)::text, ('dry-mate'::character varying)::text])))
);


--
-- TOC entry 320 (class 1259 OID 31363)
-- Name: component_collection_point_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.component_collection_point_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5608 (class 0 OID 0)
-- Dependencies: 320
-- Name: component_collection_point_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.component_collection_point_id_seq OWNED BY reference.component_collection_point.id;


--
-- TOC entry 321 (class 1259 OID 31365)
-- Name: component_connector; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.component_connector (
    id integer NOT NULL,
    fk_component_discrete_id bigint,
    fk_component_type_id smallint,
    maximum_water_depth double precision,
    number_contacts integer,
    number_fibre_channels integer,
    mating_force double precision,
    demating_force double precision,
    rated_voltage_u0 double precision,
    rated_current double precision,
    cable_area_min double precision,
    cable_area_max double precision,
    operational_temp_min double precision,
    operational_temp_max double precision,
    CONSTRAINT component_connector_fk_component_type_id_check CHECK ((fk_component_type_id = ANY (ARRAY[6, 7])))
);


--
-- TOC entry 322 (class 1259 OID 31369)
-- Name: component_connector_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.component_connector_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5609 (class 0 OID 0)
-- Dependencies: 322
-- Name: component_connector_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.component_connector_id_seq OWNED BY reference.component_connector.id;


--
-- TOC entry 323 (class 1259 OID 31371)
-- Name: component_continuous; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.component_continuous (
    id bigint NOT NULL,
    fk_component_id bigint,
    diameter double precision,
    dry_mass_per_unit_length double precision,
    wet_mass_per_unit_length double precision,
    cost_per_unit_length double precision
);


--
-- TOC entry 324 (class 1259 OID 31374)
-- Name: component_continuous_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.component_continuous_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5610 (class 0 OID 0)
-- Dependencies: 324
-- Name: component_continuous_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.component_continuous_id_seq OWNED BY reference.component_continuous.id;


--
-- TOC entry 325 (class 1259 OID 31376)
-- Name: component_discrete; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.component_discrete (
    id bigint NOT NULL,
    fk_component_id bigint,
    length double precision,
    width double precision,
    height double precision,
    dry_mass double precision,
    wet_mass double precision,
    cost double precision
);


--
-- TOC entry 326 (class 1259 OID 31379)
-- Name: component_discrete_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.component_discrete_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5611 (class 0 OID 0)
-- Dependencies: 326
-- Name: component_discrete_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.component_discrete_id_seq OWNED BY reference.component_discrete.id;


--
-- TOC entry 327 (class 1259 OID 31381)
-- Name: component_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.component_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5612 (class 0 OID 0)
-- Dependencies: 327
-- Name: component_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.component_id_seq OWNED BY reference.component.id;


--
-- TOC entry 328 (class 1259 OID 31383)
-- Name: component_mooring_continuous; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.component_mooring_continuous (
    id integer NOT NULL,
    fk_component_continuous_id bigint,
    fk_component_type_id smallint,
    connecting_length double precision,
    minimum_breaking_load double precision,
    axial_stiffness double precision,
    CONSTRAINT component_mooring_continuous_fk_component_type_id_check CHECK ((fk_component_type_id = ANY (ARRAY[4, 8])))
);


--
-- TOC entry 329 (class 1259 OID 31387)
-- Name: component_mooring_continuous_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.component_mooring_continuous_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5613 (class 0 OID 0)
-- Dependencies: 329
-- Name: component_mooring_continuous_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.component_mooring_continuous_id_seq OWNED BY reference.component_mooring_continuous.id;


--
-- TOC entry 330 (class 1259 OID 31389)
-- Name: component_mooring_discrete; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.component_mooring_discrete (
    id integer NOT NULL,
    fk_component_discrete_id bigint,
    fk_component_type_id smallint,
    nominal_diameter double precision,
    connecting_length double precision,
    minimum_breaking_load double precision,
    axial_stiffness double precision,
    CONSTRAINT component_mooring_discrete_fk_component_type_id_check CHECK ((fk_component_type_id = ANY (ARRAY[11, 12])))
);


--
-- TOC entry 331 (class 1259 OID 31393)
-- Name: component_mooring_discrete_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.component_mooring_discrete_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5614 (class 0 OID 0)
-- Dependencies: 331
-- Name: component_mooring_discrete_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.component_mooring_discrete_id_seq OWNED BY reference.component_mooring_discrete.id;


--
-- TOC entry 332 (class 1259 OID 31395)
-- Name: component_pile; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.component_pile (
    id bigint NOT NULL,
    fk_component_continuous_id bigint,
    fk_component_type_id smallint,
    wall_thickness double precision,
    yield_stress double precision,
    youngs_modulus double precision,
    CONSTRAINT component_pile_fk_component_type_id_check CHECK ((fk_component_type_id = 9))
);


--
-- TOC entry 333 (class 1259 OID 31399)
-- Name: component_pile_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.component_pile_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5615 (class 0 OID 0)
-- Dependencies: 333
-- Name: component_pile_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.component_pile_id_seq OWNED BY reference.component_pile.id;


--
-- TOC entry 334 (class 1259 OID 31401)
-- Name: component_rope; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.component_rope (
    id integer NOT NULL,
    fk_component_continuous_id bigint,
    fk_component_type_id smallint,
    material character varying(10),
    minimum_breaking_load double precision,
    rope_stiffness_curve double precision[],
    CONSTRAINT component_rope_fk_component_type_id_check CHECK ((fk_component_type_id = 10)),
    CONSTRAINT component_rope_material_check CHECK (((material)::text = ANY (ARRAY[('polyester'::character varying)::text, ('nylon'::character varying)::text, ('hmpe'::character varying)::text, ('steelite'::character varying)::text])))
);


--
-- TOC entry 335 (class 1259 OID 31409)
-- Name: component_rope_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.component_rope_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5616 (class 0 OID 0)
-- Dependencies: 335
-- Name: component_rope_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.component_rope_id_seq OWNED BY reference.component_rope.id;


--
-- TOC entry 336 (class 1259 OID 31411)
-- Name: component_shared; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.component_shared (
    id bigint NOT NULL,
    fk_component_id bigint,
    preparation_person_hours double precision,
    inspection_person_hours double precision,
    maintenance_person_hours double precision,
    replacement_person_hours double precision,
    ncfr_lower_bound double precision,
    ncfr_mean double precision,
    ncfr_upper_bound double precision,
    cfr_lower_bound double precision,
    cfr_mean double precision,
    cfr_upper_bound double precision,
    environmental_impact character varying(100)
);


--
-- TOC entry 337 (class 1259 OID 31414)
-- Name: component_shared_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.component_shared_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5617 (class 0 OID 0)
-- Dependencies: 337
-- Name: component_shared_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.component_shared_id_seq OWNED BY reference.component_shared.id;


--
-- TOC entry 338 (class 1259 OID 31416)
-- Name: component_transformer; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.component_transformer (
    id integer NOT NULL,
    fk_component_discrete_id bigint,
    fk_component_type_id smallint,
    maximum_water_depth double precision,
    power_rating double precision,
    impedance double precision,
    windings integer,
    voltage_primary_winding double precision,
    voltage_secondary_winding double precision,
    voltage_tertiary_winding double precision,
    operational_temp_min double precision,
    operational_temp_max double precision,
    CONSTRAINT component_transformer_fk_component_type_id_check CHECK ((fk_component_type_id = 13))
);


--
-- TOC entry 339 (class 1259 OID 31420)
-- Name: component_transformer_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.component_transformer_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5618 (class 0 OID 0)
-- Dependencies: 339
-- Name: component_transformer_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.component_transformer_id_seq OWNED BY reference.component_transformer.id;


--
-- TOC entry 340 (class 1259 OID 31422)
-- Name: component_type; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.component_type (
    id smallint NOT NULL,
    description character varying(20)
);


--
-- TOC entry 341 (class 1259 OID 31425)
-- Name: component_type_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.component_type_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5619 (class 0 OID 0)
-- Dependencies: 341
-- Name: component_type_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.component_type_id_seq OWNED BY reference.component_type.id;


--
-- TOC entry 342 (class 1259 OID 31427)
-- Name: constants; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.constants (
    lock character(1) NOT NULL,
    gravity double precision,
    sea_water_density double precision,
    air_density double precision,
    steel_density double precision,
    concrete_density double precision,
    grout_density double precision,
    grout_strength double precision,
    CONSTRAINT constants_lock_check CHECK ((lock = 'X'::bpchar))
);


--
-- TOC entry 5620 (class 0 OID 0)
-- Dependencies: 342
-- Name: COLUMN constants.lock; Type: COMMENT; Schema: reference; Owner: -
--

COMMENT ON COLUMN reference.constants.lock IS 'Ensures table always has a single row. Value should be "X".';


--
-- TOC entry 343 (class 1259 OID 31431)
-- Name: constraint_type_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.constraint_type_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 32767
    CACHE 1;


--
-- TOC entry 344 (class 1259 OID 31433)
-- Name: equipment_cable_burial; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.equipment_cable_burial (
    id integer NOT NULL,
    description character varying(200),
    width double precision,
    length double precision,
    height double precision,
    dry_mass double precision,
    max_operating_depth double precision,
    tow_force_required double precision,
    jetting_capability boolean,
    ploughing_capability boolean,
    cutting_capability boolean,
    jetting_trench_depth double precision,
    ploughing_trench_depth double precision,
    cutting_trench_depth double precision,
    max_cable_diameter double precision,
    min_cable_bend_radius double precision,
    additional_equipment_footprint double precision,
    additional_equipment_mass double precision,
    equipment_day_rate double precision,
    personnel_day_rate double precision
);


--
-- TOC entry 345 (class 1259 OID 31436)
-- Name: equipment_cable_burial_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.equipment_cable_burial_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5621 (class 0 OID 0)
-- Dependencies: 345
-- Name: equipment_cable_burial_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.equipment_cable_burial_id_seq OWNED BY reference.equipment_cable_burial.id;


--
-- TOC entry 346 (class 1259 OID 31438)
-- Name: equipment_divers; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.equipment_divers (
    id integer NOT NULL,
    description character varying(200),
    max_operating_depth double precision,
    deployment_eq_footprint double precision,
    deployment_eq_mass double precision,
    total_day_rate double precision
);


--
-- TOC entry 347 (class 1259 OID 31441)
-- Name: equipment_divers_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.equipment_divers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5622 (class 0 OID 0)
-- Dependencies: 347
-- Name: equipment_divers_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.equipment_divers_id_seq OWNED BY reference.equipment_divers.id;


--
-- TOC entry 348 (class 1259 OID 31443)
-- Name: equipment_drilling_rigs; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.equipment_drilling_rigs (
    id integer NOT NULL,
    description character varying(200),
    diameter double precision,
    length double precision,
    dry_mass double precision,
    max_water_depth double precision,
    max_drilling_depth double precision,
    drilling_diameter_range double precision,
    additional_equipment_footprint double precision,
    additional_equipment_mass double precision,
    equipment_day_rate double precision,
    personnel_day_rate double precision
);


--
-- TOC entry 349 (class 1259 OID 31446)
-- Name: equipment_drilling_rigs_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.equipment_drilling_rigs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5623 (class 0 OID 0)
-- Dependencies: 349
-- Name: equipment_drilling_rigs_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.equipment_drilling_rigs_id_seq OWNED BY reference.equipment_drilling_rigs.id;


--
-- TOC entry 350 (class 1259 OID 31448)
-- Name: equipment_excavating; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.equipment_excavating (
    id integer NOT NULL,
    description character varying(200),
    width double precision,
    height double precision,
    dry_mass double precision,
    depth_rating double precision,
    equipment_day_rate double precision,
    personnel_day_rate double precision
);


--
-- TOC entry 351 (class 1259 OID 31451)
-- Name: equipment_excavating_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.equipment_excavating_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5624 (class 0 OID 0)
-- Dependencies: 351
-- Name: equipment_excavating_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.equipment_excavating_id_seq OWNED BY reference.equipment_excavating.id;


--
-- TOC entry 352 (class 1259 OID 31453)
-- Name: equipment_hammer; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.equipment_hammer (
    id integer NOT NULL,
    description character varying(200),
    length double precision,
    dry_mass double precision,
    depth_rating double precision,
    min_pile_diameter double precision,
    max_pile_diameter double precision,
    additional_equipment_footprint double precision,
    additional_equipment_mass double precision,
    equipment_day_rate double precision,
    personnel_day_rate double precision
);


--
-- TOC entry 353 (class 1259 OID 31456)
-- Name: equipment_hammer_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.equipment_hammer_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5625 (class 0 OID 0)
-- Dependencies: 353
-- Name: equipment_hammer_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.equipment_hammer_id_seq OWNED BY reference.equipment_hammer.id;


--
-- TOC entry 354 (class 1259 OID 31458)
-- Name: equipment_mattress; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.equipment_mattress (
    id integer NOT NULL,
    description character varying(200),
    width double precision,
    length double precision,
    thickness double precision,
    dry_mass double precision,
    cost double precision
);


--
-- TOC entry 355 (class 1259 OID 31461)
-- Name: equipment_mattress_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.equipment_mattress_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5626 (class 0 OID 0)
-- Dependencies: 355
-- Name: equipment_mattress_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.equipment_mattress_id_seq OWNED BY reference.equipment_mattress.id;


--
-- TOC entry 356 (class 1259 OID 31463)
-- Name: equipment_rock_filter_bags; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.equipment_rock_filter_bags (
    id integer NOT NULL,
    description character varying(200),
    diameter double precision,
    height double precision,
    dry_mass double precision,
    cost double precision
);


--
-- TOC entry 357 (class 1259 OID 31466)
-- Name: equipment_rock_filter_bags_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.equipment_rock_filter_bags_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5627 (class 0 OID 0)
-- Dependencies: 357
-- Name: equipment_rock_filter_bags_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.equipment_rock_filter_bags_id_seq OWNED BY reference.equipment_rock_filter_bags.id;


--
-- TOC entry 358 (class 1259 OID 31468)
-- Name: equipment_rov; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.equipment_rov (
    id integer NOT NULL,
    description character varying(200),
    rov_class character varying(16),
    width double precision,
    length double precision,
    height double precision,
    dry_mass double precision,
    depth_rating double precision,
    payload double precision,
    manipulator_grip_force double precision,
    additional_equipment_footprint double precision,
    additional_equipment_mass double precision,
    additional_equipment_supervisors integer,
    additional_equipment_technicians integer,
    equipment_day_rate double precision,
    supervisor_day_rate double precision,
    technician_day_rate double precision,
    CONSTRAINT equipment_rov_rov_class_check CHECK (((rov_class)::text = ANY (ARRAY[('Inspection class'::character varying)::text, ('Workclass'::character varying)::text])))
);


--
-- TOC entry 359 (class 1259 OID 31472)
-- Name: equipment_rov_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.equipment_rov_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5628 (class 0 OID 0)
-- Dependencies: 359
-- Name: equipment_rov_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.equipment_rov_id_seq OWNED BY reference.equipment_rov.id;


--
-- TOC entry 360 (class 1259 OID 31474)
-- Name: equipment_soil_lay_rates; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.equipment_soil_lay_rates (
    equipment_type character varying(100) NOT NULL,
    soil_ls double precision,
    soil_ms double precision,
    soil_ds double precision,
    soil_vsc double precision,
    soil_sc double precision,
    soil_fc double precision,
    soil_stc double precision,
    soil_hgt double precision,
    soil_cm double precision,
    soil_src double precision,
    soil_hr double precision,
    soil_gc double precision
);


--
-- TOC entry 361 (class 1259 OID 31477)
-- Name: equipment_soil_penet_rates; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.equipment_soil_penet_rates (
    equipment_type character varying(100) NOT NULL,
    soil_ls double precision,
    soil_ms double precision,
    soil_ds double precision,
    soil_vsc double precision,
    soil_sc double precision,
    soil_fc double precision,
    soil_stc double precision,
    soil_hgt double precision,
    soil_cm double precision,
    soil_src double precision,
    soil_hr double precision,
    soil_gc double precision
);


--
-- TOC entry 362 (class 1259 OID 31480)
-- Name: equipment_split_pipe; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.equipment_split_pipe (
    id integer NOT NULL,
    description character varying(200),
    length double precision,
    cost double precision
);


--
-- TOC entry 363 (class 1259 OID 31483)
-- Name: equipment_split_pipe_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.equipment_split_pipe_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5629 (class 0 OID 0)
-- Dependencies: 363
-- Name: equipment_split_pipe_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.equipment_split_pipe_id_seq OWNED BY reference.equipment_split_pipe.id;


--
-- TOC entry 364 (class 1259 OID 31485)
-- Name: equipment_vibro_driver; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.equipment_vibro_driver (
    id integer NOT NULL,
    description character varying(200),
    width double precision,
    length double precision,
    height double precision,
    vibro_driver_mass double precision,
    clamp_mass double precision,
    min_pile_diameter double precision,
    max_pile_diameter double precision,
    max_pile_mass double precision,
    additional_equipment_footprint double precision,
    additional_equipment_mass double precision,
    equipment_day_rate double precision,
    personnel_day_rate double precision
);


--
-- TOC entry 365 (class 1259 OID 31488)
-- Name: equipment_vibro_driver_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.equipment_vibro_driver_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5630 (class 0 OID 0)
-- Dependencies: 365
-- Name: equipment_vibro_driver_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.equipment_vibro_driver_id_seq OWNED BY reference.equipment_vibro_driver.id;


--
-- TOC entry 366 (class 1259 OID 31490)
-- Name: operations_limit_cs; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.operations_limit_cs (
    id smallint NOT NULL,
    fk_operations_id smallint,
    cs_limit double precision
);


--
-- TOC entry 367 (class 1259 OID 31493)
-- Name: operations_limit_cs_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.operations_limit_cs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5631 (class 0 OID 0)
-- Dependencies: 367
-- Name: operations_limit_cs_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.operations_limit_cs_id_seq OWNED BY reference.operations_limit_cs.id;


--
-- TOC entry 368 (class 1259 OID 31495)
-- Name: operations_limit_hs; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.operations_limit_hs (
    id smallint NOT NULL,
    fk_operations_id smallint,
    hs_limit double precision
);


--
-- TOC entry 369 (class 1259 OID 31498)
-- Name: operations_limit_hs_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.operations_limit_hs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5632 (class 0 OID 0)
-- Dependencies: 369
-- Name: operations_limit_hs_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.operations_limit_hs_id_seq OWNED BY reference.operations_limit_hs.id;


--
-- TOC entry 370 (class 1259 OID 31500)
-- Name: operations_limit_tp; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.operations_limit_tp (
    id smallint NOT NULL,
    fk_operations_id smallint,
    tp_limit double precision
);


--
-- TOC entry 371 (class 1259 OID 31503)
-- Name: operations_limit_tp_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.operations_limit_tp_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5633 (class 0 OID 0)
-- Dependencies: 371
-- Name: operations_limit_tp_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.operations_limit_tp_id_seq OWNED BY reference.operations_limit_tp.id;


--
-- TOC entry 372 (class 1259 OID 31505)
-- Name: operations_limit_ws; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.operations_limit_ws (
    id smallint NOT NULL,
    fk_operations_id smallint,
    ws_limit double precision
);


--
-- TOC entry 373 (class 1259 OID 31508)
-- Name: operations_limit_ws_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.operations_limit_ws_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5634 (class 0 OID 0)
-- Dependencies: 373
-- Name: operations_limit_ws_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.operations_limit_ws_id_seq OWNED BY reference.operations_limit_ws.id;


--
-- TOC entry 374 (class 1259 OID 31510)
-- Name: operations_type; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.operations_type (
    id smallint NOT NULL,
    description character varying(150)
);


--
-- TOC entry 375 (class 1259 OID 31513)
-- Name: operations_type_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.operations_type_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5635 (class 0 OID 0)
-- Dependencies: 375
-- Name: operations_type_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.operations_type_id_seq OWNED BY reference.operations_type.id;


--
-- TOC entry 376 (class 1259 OID 31515)
-- Name: ports; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ports (
    id integer NOT NULL,
    name character varying(100),
    country character varying(100),
    type_of_terminal character varying(8),
    entrance_width double precision,
    terminal_length double precision,
    terminal_load_bearing double precision,
    terminal_draught double precision,
    terminal_area double precision,
    max_gantry_crane_lift_capacity double precision,
    max_tower_crane_lift_capacity double precision,
    jacking_capability boolean,
    point_location public.geometry(Point,4326),
    CONSTRAINT ports_type_of_terminal_check CHECK (((type_of_terminal)::text = ANY (ARRAY[('Quay'::character varying)::text, ('Dry-dock'::character varying)::text])))
);


--
-- TOC entry 377 (class 1259 OID 31522)
-- Name: ports_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.ports_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5636 (class 0 OID 0)
-- Dependencies: 377
-- Name: ports_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.ports_id_seq OWNED BY reference.ports.id;


--
-- TOC entry 378 (class 1259 OID 31524)
-- Name: ref_current_drag_coef_rect; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ref_current_drag_coef_rect (
    width_length double precision NOT NULL,
    thickness_width double precision
);


--
-- TOC entry 379 (class 1259 OID 31527)
-- Name: ref_drag_coef_cyl; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ref_drag_coef_cyl (
    reynolds_number double precision NOT NULL,
    smooth double precision,
    roughness_1e_5 double precision,
    roughness_1e_2 double precision
);


--
-- TOC entry 380 (class 1259 OID 31530)
-- Name: ref_drift_coef_float_rect; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ref_drift_coef_float_rect (
    wavenumber_draft double precision NOT NULL,
    reflection_coefficient double precision
);


--
-- TOC entry 381 (class 1259 OID 31533)
-- Name: ref_holding_capacity_factors_plate_anchors; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ref_holding_capacity_factors_plate_anchors (
    relative_embedment_depth double precision NOT NULL,
    drained_friction_angle_20deg double precision,
    drained_friction_angle_25deg double precision,
    drained_friction_angle_30deg double precision,
    drained_friction_angle_35deg double precision,
    drained_friction_angle_40deg double precision
);


--
-- TOC entry 382 (class 1259 OID 31536)
-- Name: ref_line_bcf; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ref_line_bcf (
    soil_friction_angle double precision NOT NULL,
    bearing_capacity_factor double precision
);


--
-- TOC entry 383 (class 1259 OID 31539)
-- Name: ref_pile_deflection_coefficients; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ref_pile_deflection_coefficients (
    depth_coefficient double precision NOT NULL,
    coefficient_ay double precision,
    coefficient_by double precision
);


--
-- TOC entry 384 (class 1259 OID 31542)
-- Name: ref_pile_limiting_values_noncalcareous; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ref_pile_limiting_values_noncalcareous (
    soil_friction_angle double precision NOT NULL,
    friction_angle_sand_pile double precision,
    bearing_capacity_factor double precision,
    max_unit_skin_friction double precision,
    max_end_bearing_capacity double precision
);


--
-- TOC entry 385 (class 1259 OID 31545)
-- Name: ref_pile_moment_coefficient_sam; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ref_pile_moment_coefficient_sam (
    depth_coefficient double precision NOT NULL,
    pile_length_relative_soil_pile_stiffness_10 double precision,
    pile_length_relative_soil_pile_stiffness_5 double precision,
    pile_length_relative_soil_pile_stiffness_4 double precision,
    pile_length_relative_soil_pile_stiffness_3 double precision,
    pile_length_relative_soil_pile_stiffness_2 double precision
);


--
-- TOC entry 386 (class 1259 OID 31548)
-- Name: ref_pile_moment_coefficient_sbm; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ref_pile_moment_coefficient_sbm (
    depth_coefficient double precision NOT NULL,
    pile_length_relative_soil_pile_stiffness_10 double precision,
    pile_length_relative_soil_pile_stiffness_5 double precision,
    pile_length_relative_soil_pile_stiffness_4 double precision,
    pile_length_relative_soil_pile_stiffness_3 double precision,
    pile_length_relative_soil_pile_stiffness_2 double precision
);


--
-- TOC entry 387 (class 1259 OID 31551)
-- Name: ref_rectangular_wave_inertia; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ref_rectangular_wave_inertia (
    "width/length" double precision NOT NULL,
    inertia_coefficients double precision
);


--
-- TOC entry 388 (class 1259 OID 31554)
-- Name: ref_subgrade_reaction_coefficient_cohesionless; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ref_subgrade_reaction_coefficient_cohesionless (
    allowable_deflection_diameter double precision NOT NULL,
    relative_density_35 double precision,
    relative_density_50 double precision,
    relative_density_65 double precision,
    relative_density_85 double precision
);


--
-- TOC entry 389 (class 1259 OID 31557)
-- Name: ref_subgrade_reaction_coefficient_k1_cohesive; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ref_subgrade_reaction_coefficient_k1_cohesive (
    allowable_deflection_diameter double precision NOT NULL,
    softclay double precision,
    stiffclay double precision
);


--
-- TOC entry 390 (class 1259 OID 31560)
-- Name: ref_superline_nylon; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ref_superline_nylon (
    extension double precision NOT NULL,
    load_mbl double precision
);


--
-- TOC entry 391 (class 1259 OID 31563)
-- Name: ref_superline_polyester; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ref_superline_polyester (
    extension double precision NOT NULL,
    load_mbl double precision
);


--
-- TOC entry 392 (class 1259 OID 31566)
-- Name: ref_superline_steelite; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ref_superline_steelite (
    extension double precision NOT NULL,
    load_mbl double precision
);


--
-- TOC entry 393 (class 1259 OID 31569)
-- Name: ref_wake_amplification_factor_cyl; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ref_wake_amplification_factor_cyl (
    kc_steady_drag_coefficient double precision NOT NULL,
    amplification_factor_for_smooth_cylinders double precision,
    amplification_factor_for_rough_cylinders double precision
);


--
-- TOC entry 394 (class 1259 OID 31572)
-- Name: ref_wind_drag_coef_rect; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.ref_wind_drag_coef_rect (
    width_length double precision NOT NULL,
    height_breadth_between_0_1 double precision,
    height_breadth_less_1 double precision,
    height_breadth_less_2 double precision,
    height_breadth_less_4 double precision,
    height_breadth_less_6 double precision,
    height_breadth_less_10 double precision,
    height_breadth_less_20 double precision
);


--
-- TOC entry 395 (class 1259 OID 31575)
-- Name: soil_type_geotechnical_properties; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.soil_type_geotechnical_properties (
    fk_soil_type_id integer NOT NULL,
    drained_soil_friction_angle double precision,
    relative_soil_density double precision,
    buoyant_unit_weight_of_soil double precision,
    effective_drained_cohesion double precision,
    seafloor_friction_coefficient double precision,
    soil_sensitivity double precision,
    rock_compressive_strength double precision,
    undrained_soil_shear_strength_constant_term double precision,
    undrained_soil_shear_strength_depth_dependent_term double precision
);


--
-- TOC entry 396 (class 1259 OID 31578)
-- Name: soil_type_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.soil_type_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5637 (class 0 OID 0)
-- Dependencies: 396
-- Name: soil_type_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.soil_type_id_seq OWNED BY reference.soil_type.id;


--
-- TOC entry 397 (class 1259 OID 31580)
-- Name: vehicle; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.vehicle (
    id bigint NOT NULL,
    description character varying(200)
);


--
-- TOC entry 398 (class 1259 OID 31583)
-- Name: vehicle_helicopter; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.vehicle_helicopter (
    id integer NOT NULL,
    fk_vehicle_id bigint,
    fk_vehicle_type_id smallint,
    deck_space double precision,
    max_deck_load_pressure double precision,
    max_cargo_mass double precision,
    crane_max_load_mass double precision,
    external_personel integer,
    CONSTRAINT vehicle_helicopter_fk_vehicle_type_id_check CHECK ((fk_vehicle_type_id = 9))
);


--
-- TOC entry 399 (class 1259 OID 31587)
-- Name: vehicle_helicopter_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.vehicle_helicopter_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5638 (class 0 OID 0)
-- Dependencies: 399
-- Name: vehicle_helicopter_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.vehicle_helicopter_id_seq OWNED BY reference.vehicle_helicopter.id;


--
-- TOC entry 400 (class 1259 OID 31589)
-- Name: vehicle_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.vehicle_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5639 (class 0 OID 0)
-- Dependencies: 400
-- Name: vehicle_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.vehicle_id_seq OWNED BY reference.vehicle.id;


--
-- TOC entry 401 (class 1259 OID 31591)
-- Name: vehicle_shared; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.vehicle_shared (
    id bigint NOT NULL,
    fk_vehicle_id bigint,
    gross_tonnage double precision,
    length double precision,
    consumption double precision,
    transit_speed double precision,
    transit_max_hs double precision,
    transit_max_tp double precision,
    transit_max_cs double precision,
    transit_max_ws double precision,
    mobilisation_time double precision,
    mobilisation_percentage_cost double precision,
    min_day_rate double precision,
    max_day_rate double precision
);


--
-- TOC entry 402 (class 1259 OID 31594)
-- Name: vehicle_shared_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.vehicle_shared_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5640 (class 0 OID 0)
-- Dependencies: 402
-- Name: vehicle_shared_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.vehicle_shared_id_seq OWNED BY reference.vehicle_shared.id;


--
-- TOC entry 403 (class 1259 OID 31596)
-- Name: vehicle_type; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.vehicle_type (
    id smallint NOT NULL,
    description character varying(40)
);


--
-- TOC entry 404 (class 1259 OID 31599)
-- Name: vehicle_type_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.vehicle_type_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5641 (class 0 OID 0)
-- Dependencies: 404
-- Name: vehicle_type_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.vehicle_type_id_seq OWNED BY reference.vehicle_type.id;


--
-- TOC entry 405 (class 1259 OID 31601)
-- Name: vehicle_vessel_anchor_handling; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.vehicle_vessel_anchor_handling (
    id integer NOT NULL,
    fk_vehicle_id bigint,
    fk_vehicle_type_id smallint,
    beam double precision,
    max_draft double precision,
    consumption_towing double precision,
    deck_space double precision,
    max_deck_load_pressure double precision,
    max_cargo_mass double precision,
    crane_max_load_mass double precision,
    bollard_pull double precision,
    anchor_handling_drum_capacity double precision,
    anchor_handling_winch_rated_pull double precision,
    external_personel integer,
    towing_max_hs double precision,
    CONSTRAINT vehicle_vessel_anchor_handling_fk_vehicle_type_id_check CHECK ((fk_vehicle_type_id = ANY (ARRAY[1, 12])))
);


--
-- TOC entry 406 (class 1259 OID 31605)
-- Name: vehicle_vessel_anchor_handling_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.vehicle_vessel_anchor_handling_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5642 (class 0 OID 0)
-- Dependencies: 406
-- Name: vehicle_vessel_anchor_handling_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.vehicle_vessel_anchor_handling_id_seq OWNED BY reference.vehicle_vessel_anchor_handling.id;


--
-- TOC entry 407 (class 1259 OID 31607)
-- Name: vehicle_vessel_cable_laying; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.vehicle_vessel_cable_laying (
    id integer NOT NULL,
    fk_vehicle_id bigint,
    fk_vehicle_type_id smallint,
    beam double precision,
    max_draft double precision,
    deck_space double precision,
    max_deck_load_pressure double precision,
    max_cargo_mass double precision,
    crane_max_load_mass double precision,
    bollard_pull double precision,
    number_turntables integer,
    turntable_max_load_mass double precision,
    turntable_inner_diameter double precision,
    cable_splice_capabilities boolean,
    dynamic_positioning_capabilities boolean,
    external_personel integer,
    CONSTRAINT vehicle_vessel_cable_laying_fk_vehicle_type_id_check CHECK ((fk_vehicle_type_id = ANY (ARRAY[3, 4])))
);


--
-- TOC entry 408 (class 1259 OID 31611)
-- Name: vehicle_vessel_cable_laying_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.vehicle_vessel_cable_laying_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5643 (class 0 OID 0)
-- Dependencies: 408
-- Name: vehicle_vessel_cable_laying_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.vehicle_vessel_cable_laying_id_seq OWNED BY reference.vehicle_vessel_cable_laying.id;


--
-- TOC entry 409 (class 1259 OID 31613)
-- Name: vehicle_vessel_cargo; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.vehicle_vessel_cargo (
    id integer NOT NULL,
    fk_vehicle_id bigint,
    fk_vehicle_type_id smallint,
    beam double precision,
    max_draft double precision,
    deck_space double precision,
    max_deck_load_pressure double precision,
    max_cargo_mass double precision,
    crane_max_load_mass double precision,
    dynamic_positioning_capabilities boolean,
    external_personel integer,
    CONSTRAINT vehicle_vessel_cargo_fk_vehicle_type_id_check CHECK ((fk_vehicle_type_id = ANY (ARRAY[2, 5, 6, 7, 8])))
);


--
-- TOC entry 410 (class 1259 OID 31617)
-- Name: vehicle_vessel_cargo_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.vehicle_vessel_cargo_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5644 (class 0 OID 0)
-- Dependencies: 410
-- Name: vehicle_vessel_cargo_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.vehicle_vessel_cargo_id_seq OWNED BY reference.vehicle_vessel_cargo.id;


--
-- TOC entry 411 (class 1259 OID 31619)
-- Name: vehicle_vessel_jackup; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.vehicle_vessel_jackup (
    id integer NOT NULL,
    fk_vehicle_id bigint,
    fk_vehicle_type_id smallint,
    beam double precision,
    max_draft double precision,
    deck_space double precision,
    max_deck_load_pressure double precision,
    max_cargo_mass double precision,
    crane_max_load_mass double precision,
    dynamic_positioning_capabilities boolean,
    jackup_max_water_depth double precision,
    jackup_speed_down double precision,
    jackup_max_payload_mass double precision,
    external_personel integer,
    jacking_max_hs double precision,
    jacking_max_tp double precision,
    jacking_max_cs double precision,
    jacking_max_ws double precision,
    CONSTRAINT vehicle_vessel_jackup_fk_vehicle_type_id_check CHECK ((fk_vehicle_type_id = ANY (ARRAY[10, 11])))
);


--
-- TOC entry 412 (class 1259 OID 31623)
-- Name: vehicle_vessel_jackup_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.vehicle_vessel_jackup_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5645 (class 0 OID 0)
-- Dependencies: 412
-- Name: vehicle_vessel_jackup_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.vehicle_vessel_jackup_id_seq OWNED BY reference.vehicle_vessel_jackup.id;


--
-- TOC entry 413 (class 1259 OID 31625)
-- Name: vehicle_vessel_tugboat; Type: TABLE; Schema: reference; Owner: -
--

CREATE TABLE reference.vehicle_vessel_tugboat (
    id integer NOT NULL,
    fk_vehicle_id bigint,
    fk_vehicle_type_id smallint,
    beam double precision,
    max_draft double precision,
    consumption_towing double precision,
    bollard_pull double precision,
    CONSTRAINT vehicle_vessel_tugboat_fk_vehicle_type_id_check CHECK ((fk_vehicle_type_id = 13))
);


--
-- TOC entry 414 (class 1259 OID 31629)
-- Name: vehicle_vessel_tugboat_id_seq; Type: SEQUENCE; Schema: reference; Owner: -
--

CREATE SEQUENCE reference.vehicle_vessel_tugboat_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5646 (class 0 OID 0)
-- Dependencies: 414
-- Name: vehicle_vessel_tugboat_id_seq; Type: SEQUENCE OWNED BY; Schema: reference; Owner: -
--

ALTER SEQUENCE reference.vehicle_vessel_tugboat_id_seq OWNED BY reference.vehicle_vessel_tugboat.id;


--
-- TOC entry 415 (class 1259 OID 31631)
-- Name: view_component_cable_dynamic; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_component_cable_dynamic AS
 SELECT component.id AS component_id,
    component.description,
    component_continuous.diameter,
    component_continuous.dry_mass_per_unit_length,
    component_continuous.wet_mass_per_unit_length,
    component_cable.minimum_breaking_load,
    component_cable.minimum_bend_radius,
    component_cable.number_conductors,
    component_cable.number_fibre_channels,
    component_cable.resistance_dc_20,
    component_cable.resistance_ac_90,
    component_cable.inductive_reactance,
    component_cable.capacitance,
    component_cable.rated_current_air,
    component_cable.rated_current_buried,
    component_cable.rated_current_jtube,
    component_cable.rated_voltage_u0,
    component_cable.operational_temp_max,
    component_continuous.cost_per_unit_length,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM ((((reference.component_cable
     JOIN reference.component_continuous ON ((component_cable.fk_component_continuous_id = component_continuous.id)))
     JOIN reference.component_shared ON ((component_continuous.fk_component_id = component_shared.fk_component_id)))
     JOIN reference.component ON ((component_continuous.fk_component_id = component.id)))
     JOIN reference.component_type ON ((component_cable.fk_component_type_id = component_type.id)))
  WHERE ((component_type.description)::text = 'cable dynamic'::text);


--
-- TOC entry 416 (class 1259 OID 31636)
-- Name: view_component_cable_static; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_component_cable_static AS
 SELECT component.id AS component_id,
    component.description,
    component_continuous.diameter,
    component_continuous.dry_mass_per_unit_length,
    component_continuous.wet_mass_per_unit_length,
    component_cable.minimum_breaking_load,
    component_cable.minimum_bend_radius,
    component_cable.number_conductors,
    component_cable.number_fibre_channels,
    component_cable.resistance_dc_20,
    component_cable.resistance_ac_90,
    component_cable.inductive_reactance,
    component_cable.capacitance,
    component_cable.rated_current_air,
    component_cable.rated_current_buried,
    component_cable.rated_current_jtube,
    component_cable.rated_voltage_u0,
    component_cable.operational_temp_max,
    component_continuous.cost_per_unit_length,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM ((((reference.component_cable
     JOIN reference.component_continuous ON ((component_cable.fk_component_continuous_id = component_continuous.id)))
     JOIN reference.component_shared ON ((component_continuous.fk_component_id = component_shared.fk_component_id)))
     JOIN reference.component ON ((component_continuous.fk_component_id = component.id)))
     JOIN reference.component_type ON ((component_cable.fk_component_type_id = component_type.id)))
  WHERE ((component_type.description)::text = 'cable static'::text);


--
-- TOC entry 417 (class 1259 OID 31641)
-- Name: view_component_collection_point; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_component_collection_point AS
 SELECT component.id AS component_id,
    component.description,
    component_discrete.width,
    component_discrete.length AS depth,
    component_discrete.height,
    component_discrete.dry_mass,
    component_discrete.wet_mass,
    component_collection_point.wet_frontal_area,
    component_collection_point.dry_frontal_area,
    component_collection_point.wet_beam_area,
    component_collection_point.dry_beam_area,
    component_collection_point.maximum_water_depth,
    component_collection_point.orientation_angle,
    component_collection_point.input_lines,
    component_collection_point.output_lines,
    component_collection_point.input_connector_type,
    component_collection_point.output_connector_type,
    component_collection_point.number_fibre_channels,
    component_collection_point.voltage_primary_winding,
    component_collection_point.voltage_secondary_winding,
    component_collection_point.rated_operating_current,
    component_collection_point.operational_temp_min,
    component_collection_point.operational_temp_max,
    component_collection_point.foundation_locations,
    component_collection_point.centre_of_gravity,
    component_discrete.cost,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM (((reference.component_collection_point
     JOIN reference.component_discrete ON ((component_collection_point.fk_component_discrete_id = component_discrete.id)))
     JOIN reference.component_shared ON ((component_discrete.fk_component_id = component_shared.fk_component_id)))
     JOIN reference.component ON ((component_discrete.fk_component_id = component.id)));


--
-- TOC entry 418 (class 1259 OID 31646)
-- Name: view_component_connector_drymate; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_component_connector_drymate AS
 SELECT component.id AS component_id,
    component.description,
    component_discrete.width,
    component_discrete.length AS depth,
    component_discrete.height,
    component_discrete.dry_mass,
    component_discrete.wet_mass,
    component_connector.maximum_water_depth,
    component_connector.number_contacts,
    component_connector.number_fibre_channels,
    component_connector.mating_force,
    component_connector.demating_force,
    component_connector.rated_voltage_u0,
    component_connector.rated_current,
    component_connector.cable_area_min,
    component_connector.cable_area_max,
    component_connector.operational_temp_min,
    component_connector.operational_temp_max,
    component_discrete.cost,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM ((((reference.component_connector
     JOIN reference.component_discrete ON ((component_connector.fk_component_discrete_id = component_discrete.id)))
     JOIN reference.component_shared ON ((component_discrete.fk_component_id = component_shared.fk_component_id)))
     JOIN reference.component ON ((component_discrete.fk_component_id = component.id)))
     JOIN reference.component_type ON ((component_connector.fk_component_type_id = component_type.id)))
  WHERE ((component_type.description)::text = 'connector dry-mate'::text);


--
-- TOC entry 419 (class 1259 OID 31651)
-- Name: view_component_connector_wetmate; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_component_connector_wetmate AS
 SELECT component.id AS component_id,
    component.description,
    component_discrete.width,
    component_discrete.length AS depth,
    component_discrete.height,
    component_discrete.dry_mass,
    component_discrete.wet_mass,
    component_connector.maximum_water_depth,
    component_connector.number_contacts,
    component_connector.number_fibre_channels,
    component_connector.mating_force,
    component_connector.demating_force,
    component_connector.rated_voltage_u0,
    component_connector.rated_current,
    component_connector.cable_area_min,
    component_connector.cable_area_max,
    component_connector.operational_temp_min,
    component_connector.operational_temp_max,
    component_discrete.cost,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM ((((reference.component_connector
     JOIN reference.component_discrete ON ((component_connector.fk_component_discrete_id = component_discrete.id)))
     JOIN reference.component_shared ON ((component_discrete.fk_component_id = component_shared.fk_component_id)))
     JOIN reference.component ON ((component_discrete.fk_component_id = component.id)))
     JOIN reference.component_type ON ((component_connector.fk_component_type_id = component_type.id)))
  WHERE ((component_type.description)::text = 'connector wet-mate'::text);


--
-- TOC entry 420 (class 1259 OID 31656)
-- Name: view_component_foundations_anchor; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_component_foundations_anchor AS
 SELECT component.id AS component_id,
    component.description,
    component_discrete.width,
    component_discrete.length AS depth,
    component_discrete.height,
    component_discrete.dry_mass,
    component_discrete.wet_mass,
    component_anchor.connecting_size,
    component_anchor.minimum_breaking_load,
    component_anchor.axial_stiffness,
    component_discrete.cost,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM (((reference.component_anchor
     JOIN reference.component_discrete ON ((component_anchor.fk_component_discrete_id = component_discrete.id)))
     JOIN reference.component_shared ON ((component_discrete.fk_component_id = component_shared.fk_component_id)))
     JOIN reference.component ON ((component_discrete.fk_component_id = component.id)));


--
-- TOC entry 421 (class 1259 OID 31661)
-- Name: view_component_foundations_anchor_coefs; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_component_foundations_anchor_coefs AS
 SELECT component.id AS component_id,
    component_anchor.soft_holding_cap_coef_1,
    component_anchor.soft_holding_cap_coef_2,
    component_anchor.soft_penetration_coef_1,
    component_anchor.soft_penetration_coef_2,
    component_anchor.sand_holding_cap_coef_1,
    component_anchor.sand_holding_cap_coef_2,
    component_anchor.sand_penetration_coef_1,
    component_anchor.sand_penetration_coef_2
   FROM ((reference.component_anchor
     JOIN reference.component_discrete ON ((component_anchor.fk_component_discrete_id = component_discrete.id)))
     JOIN reference.component ON ((component_discrete.fk_component_id = component.id)));


--
-- TOC entry 422 (class 1259 OID 31666)
-- Name: view_component_foundations_pile; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_component_foundations_pile AS
 SELECT component.id AS component_id,
    component.description,
    component_continuous.diameter,
    component_continuous.dry_mass_per_unit_length,
    component_continuous.wet_mass_per_unit_length,
    component_pile.wall_thickness,
    component_pile.yield_stress,
    component_pile.youngs_modulus,
    component_continuous.cost_per_unit_length,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM (((reference.component_pile
     JOIN reference.component_continuous ON ((component_pile.fk_component_continuous_id = component_continuous.id)))
     JOIN reference.component_shared ON ((component_continuous.fk_component_id = component_shared.fk_component_id)))
     JOIN reference.component ON ((component_continuous.fk_component_id = component.id)));


--
-- TOC entry 423 (class 1259 OID 31671)
-- Name: view_component_moorings_chain; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_component_moorings_chain AS
 SELECT component.id AS component_id,
    component.description,
    component_continuous.diameter,
    component_continuous.dry_mass_per_unit_length,
    component_continuous.wet_mass_per_unit_length,
    component_mooring_continuous.connecting_length,
    component_mooring_continuous.minimum_breaking_load,
    component_mooring_continuous.axial_stiffness,
    component_continuous.cost_per_unit_length,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM ((((reference.component_mooring_continuous
     JOIN reference.component_continuous ON ((component_mooring_continuous.fk_component_continuous_id = component_continuous.id)))
     JOIN reference.component_shared ON ((component_continuous.fk_component_id = component_shared.fk_component_id)))
     JOIN reference.component ON ((component_continuous.fk_component_id = component.id)))
     JOIN reference.component_type ON ((component_mooring_continuous.fk_component_type_id = component_type.id)))
  WHERE ((component_type.description)::text = 'chain'::text);


--
-- TOC entry 424 (class 1259 OID 31676)
-- Name: view_component_moorings_forerunner; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_component_moorings_forerunner AS
 SELECT component.id AS component_id,
    component.description,
    component_continuous.diameter,
    component_continuous.dry_mass_per_unit_length,
    component_continuous.wet_mass_per_unit_length,
    component_mooring_continuous.connecting_length,
    component_mooring_continuous.minimum_breaking_load,
    component_mooring_continuous.axial_stiffness,
    component_continuous.cost_per_unit_length,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM ((((reference.component_mooring_continuous
     JOIN reference.component_continuous ON ((component_mooring_continuous.fk_component_continuous_id = component_continuous.id)))
     JOIN reference.component_shared ON ((component_continuous.fk_component_id = component_shared.fk_component_id)))
     JOIN reference.component ON ((component_continuous.fk_component_id = component.id)))
     JOIN reference.component_type ON ((component_mooring_continuous.fk_component_type_id = component_type.id)))
  WHERE ((component_type.description)::text = 'forerunner'::text);


--
-- TOC entry 425 (class 1259 OID 31681)
-- Name: view_component_moorings_rope; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_component_moorings_rope AS
 SELECT component.id AS component_id,
    component.description,
    component_continuous.diameter,
    component_continuous.dry_mass_per_unit_length,
    component_continuous.wet_mass_per_unit_length,
    component_rope.material,
    component_rope.minimum_breaking_load,
    component_rope.rope_stiffness_curve,
    component_continuous.cost_per_unit_length,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM (((reference.component_rope
     JOIN reference.component_continuous ON ((component_rope.fk_component_continuous_id = component_continuous.id)))
     JOIN reference.component_shared ON ((component_continuous.fk_component_id = component_shared.fk_component_id)))
     JOIN reference.component ON ((component_continuous.fk_component_id = component.id)));


--
-- TOC entry 426 (class 1259 OID 31686)
-- Name: view_component_moorings_shackle; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_component_moorings_shackle AS
 SELECT component.id AS component_id,
    component.description,
    component_discrete.width,
    component_discrete.length AS depth,
    component_discrete.height,
    component_discrete.dry_mass,
    component_discrete.wet_mass,
    component_mooring_discrete.nominal_diameter,
    component_mooring_discrete.connecting_length,
    component_mooring_discrete.minimum_breaking_load,
    component_mooring_discrete.axial_stiffness,
    component_discrete.cost,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM ((((reference.component_mooring_discrete
     JOIN reference.component_discrete ON ((component_mooring_discrete.fk_component_discrete_id = component_discrete.id)))
     JOIN reference.component_shared ON ((component_discrete.fk_component_id = component_shared.fk_component_id)))
     JOIN reference.component ON ((component_discrete.fk_component_id = component.id)))
     JOIN reference.component_type ON ((component_mooring_discrete.fk_component_type_id = component_type.id)))
  WHERE ((component_type.description)::text = 'shackle'::text);


--
-- TOC entry 427 (class 1259 OID 31691)
-- Name: view_component_moorings_swivel; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_component_moorings_swivel AS
 SELECT component.id AS component_id,
    component.description,
    component_discrete.width,
    component_discrete.length AS depth,
    component_discrete.height,
    component_discrete.dry_mass,
    component_discrete.wet_mass,
    component_mooring_discrete.nominal_diameter,
    component_mooring_discrete.connecting_length,
    component_mooring_discrete.minimum_breaking_load,
    component_mooring_discrete.axial_stiffness,
    component_discrete.cost,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM ((((reference.component_mooring_discrete
     JOIN reference.component_discrete ON ((component_mooring_discrete.fk_component_discrete_id = component_discrete.id)))
     JOIN reference.component_shared ON ((component_discrete.fk_component_id = component_shared.fk_component_id)))
     JOIN reference.component ON ((component_discrete.fk_component_id = component.id)))
     JOIN reference.component_type ON ((component_mooring_discrete.fk_component_type_id = component_type.id)))
  WHERE ((component_type.description)::text = 'swivel'::text);


--
-- TOC entry 428 (class 1259 OID 31696)
-- Name: view_component_transformer; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_component_transformer AS
 SELECT component.id AS component_id,
    component.description,
    component_discrete.width,
    component_discrete.length AS depth,
    component_discrete.height,
    component_discrete.dry_mass,
    component_discrete.wet_mass,
    component_transformer.maximum_water_depth,
    component_transformer.power_rating,
    component_transformer.impedance,
    component_transformer.windings,
    component_transformer.voltage_primary_winding,
    component_transformer.voltage_secondary_winding,
    component_transformer.voltage_tertiary_winding,
    component_transformer.operational_temp_min,
    component_transformer.operational_temp_max,
    component_discrete.cost,
    component_shared.preparation_person_hours,
    component_shared.inspection_person_hours,
    component_shared.maintenance_person_hours,
    component_shared.replacement_person_hours,
    component_shared.ncfr_lower_bound,
    component_shared.ncfr_mean,
    component_shared.ncfr_upper_bound,
    component_shared.cfr_lower_bound,
    component_shared.cfr_mean,
    component_shared.cfr_upper_bound,
    component_shared.environmental_impact
   FROM (((reference.component_transformer
     JOIN reference.component_discrete ON ((component_transformer.fk_component_discrete_id = component_discrete.id)))
     JOIN reference.component_shared ON ((component_discrete.fk_component_id = component_shared.fk_component_id)))
     JOIN reference.component ON ((component_discrete.fk_component_id = component.id)));


--
-- TOC entry 429 (class 1259 OID 31701)
-- Name: view_operations_limit_cs; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_operations_limit_cs AS
 SELECT operations_type.description AS operations_type,
    operations_limit_cs.cs_limit
   FROM (reference.operations_limit_cs
     JOIN reference.operations_type ON ((operations_limit_cs.fk_operations_id = operations_type.id)));


--
-- TOC entry 430 (class 1259 OID 31705)
-- Name: view_operations_limit_hs; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_operations_limit_hs AS
 SELECT operations_type.description AS operations_type,
    operations_limit_hs.hs_limit
   FROM (reference.operations_limit_hs
     JOIN reference.operations_type ON ((operations_limit_hs.fk_operations_id = operations_type.id)));


--
-- TOC entry 431 (class 1259 OID 31709)
-- Name: view_operations_limit_tp; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_operations_limit_tp AS
 SELECT operations_type.description AS operations_type,
    operations_limit_tp.tp_limit
   FROM (reference.operations_limit_tp
     JOIN reference.operations_type ON ((operations_limit_tp.fk_operations_id = operations_type.id)));


--
-- TOC entry 432 (class 1259 OID 31713)
-- Name: view_operations_limit_ws; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_operations_limit_ws AS
 SELECT operations_type.description AS operations_type,
    operations_limit_ws.ws_limit
   FROM (reference.operations_limit_ws
     JOIN reference.operations_type ON ((operations_limit_ws.fk_operations_id = operations_type.id)));


--
-- TOC entry 433 (class 1259 OID 31717)
-- Name: view_soil_type_geotechnical_properties; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_soil_type_geotechnical_properties AS
 SELECT soil_type.description AS soil_type,
    soil_type_geotechnical_properties.drained_soil_friction_angle,
    soil_type_geotechnical_properties.relative_soil_density,
    soil_type_geotechnical_properties.buoyant_unit_weight_of_soil,
    soil_type_geotechnical_properties.undrained_soil_shear_strength_constant_term,
    soil_type_geotechnical_properties.undrained_soil_shear_strength_depth_dependent_term,
    soil_type_geotechnical_properties.effective_drained_cohesion,
    soil_type_geotechnical_properties.seafloor_friction_coefficient,
    soil_type_geotechnical_properties.soil_sensitivity,
    soil_type_geotechnical_properties.rock_compressive_strength
   FROM (reference.soil_type_geotechnical_properties
     JOIN reference.soil_type ON ((soil_type_geotechnical_properties.fk_soil_type_id = soil_type.id)));


--
-- TOC entry 434 (class 1259 OID 31721)
-- Name: view_vehicle_helicopter; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_vehicle_helicopter AS
 SELECT vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    vehicle_helicopter.deck_space,
    vehicle_helicopter.max_deck_load_pressure,
    vehicle_helicopter.max_cargo_mass,
    vehicle_helicopter.crane_max_load_mass,
    vehicle_helicopter.external_personel,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM ((reference.vehicle_helicopter
     JOIN reference.vehicle_shared ON ((vehicle_helicopter.fk_vehicle_id = vehicle_shared.fk_vehicle_id)))
     JOIN reference.vehicle ON ((vehicle_shared.fk_vehicle_id = vehicle.id)));


--
-- TOC entry 435 (class 1259 OID 31726)
-- Name: view_vehicle_vessel_ahts; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_vehicle_vessel_ahts AS
 SELECT vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    vehicle_vessel_anchor_handling.beam,
    vehicle_vessel_anchor_handling.max_draft,
    vehicle_shared.consumption,
    vehicle_vessel_anchor_handling.consumption_towing,
    vehicle_shared.transit_speed,
    vehicle_vessel_anchor_handling.deck_space,
    vehicle_vessel_anchor_handling.max_deck_load_pressure,
    vehicle_vessel_anchor_handling.max_cargo_mass,
    vehicle_vessel_anchor_handling.crane_max_load_mass,
    vehicle_vessel_anchor_handling.bollard_pull,
    vehicle_vessel_anchor_handling.anchor_handling_drum_capacity,
    vehicle_vessel_anchor_handling.anchor_handling_winch_rated_pull,
    vehicle_vessel_anchor_handling.external_personel,
    vehicle_vessel_anchor_handling.towing_max_hs,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM (((reference.vehicle_vessel_anchor_handling
     JOIN reference.vehicle_shared ON ((vehicle_vessel_anchor_handling.fk_vehicle_id = vehicle_shared.fk_vehicle_id)))
     JOIN reference.vehicle ON ((vehicle_shared.fk_vehicle_id = vehicle.id)))
     JOIN reference.vehicle_type ON ((vehicle_vessel_anchor_handling.fk_vehicle_type_id = vehicle_type.id)))
  WHERE ((vehicle_type.description)::text = 'anchor handling tug supply vessel'::text);


--
-- TOC entry 436 (class 1259 OID 31731)
-- Name: view_vehicle_vessel_barge; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_vehicle_vessel_barge AS
 SELECT vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    vehicle_vessel_cargo.beam,
    vehicle_vessel_cargo.max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    vehicle_vessel_cargo.deck_space,
    vehicle_vessel_cargo.max_deck_load_pressure,
    vehicle_vessel_cargo.max_cargo_mass,
    vehicle_vessel_cargo.crane_max_load_mass,
    vehicle_vessel_cargo.dynamic_positioning_capabilities,
    vehicle_vessel_cargo.external_personel,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM (((reference.vehicle_vessel_cargo
     JOIN reference.vehicle_shared ON ((vehicle_vessel_cargo.fk_vehicle_id = vehicle_shared.fk_vehicle_id)))
     JOIN reference.vehicle ON ((vehicle_shared.fk_vehicle_id = vehicle.id)))
     JOIN reference.vehicle_type ON ((vehicle_vessel_cargo.fk_vehicle_type_id = vehicle_type.id)))
  WHERE ((vehicle_type.description)::text = 'barge'::text);


--
-- TOC entry 437 (class 1259 OID 31736)
-- Name: view_vehicle_vessel_clb; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_vehicle_vessel_clb AS
 SELECT vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    vehicle_vessel_cable_laying.beam,
    vehicle_vessel_cable_laying.max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    vehicle_vessel_cable_laying.deck_space,
    vehicle_vessel_cable_laying.max_deck_load_pressure,
    vehicle_vessel_cable_laying.max_cargo_mass,
    vehicle_vessel_cable_laying.crane_max_load_mass,
    vehicle_vessel_cable_laying.number_turntables,
    vehicle_vessel_cable_laying.turntable_max_load_mass,
    vehicle_vessel_cable_laying.turntable_inner_diameter,
    vehicle_vessel_cable_laying.cable_splice_capabilities,
    vehicle_vessel_cable_laying.dynamic_positioning_capabilities,
    vehicle_vessel_cable_laying.external_personel,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM (((reference.vehicle_vessel_cable_laying
     JOIN reference.vehicle_shared ON ((vehicle_vessel_cable_laying.fk_vehicle_id = vehicle_shared.fk_vehicle_id)))
     JOIN reference.vehicle ON ((vehicle_shared.fk_vehicle_id = vehicle.id)))
     JOIN reference.vehicle_type ON ((vehicle_vessel_cable_laying.fk_vehicle_type_id = vehicle_type.id)))
  WHERE ((vehicle_type.description)::text = 'cable laying barge'::text);


--
-- TOC entry 438 (class 1259 OID 31741)
-- Name: view_vehicle_vessel_clv; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_vehicle_vessel_clv AS
 SELECT vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    vehicle_vessel_cable_laying.beam,
    vehicle_vessel_cable_laying.max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    vehicle_vessel_cable_laying.deck_space,
    vehicle_vessel_cable_laying.max_deck_load_pressure,
    vehicle_vessel_cable_laying.max_cargo_mass,
    vehicle_vessel_cable_laying.crane_max_load_mass,
    vehicle_vessel_cable_laying.bollard_pull,
    vehicle_vessel_cable_laying.number_turntables,
    vehicle_vessel_cable_laying.turntable_max_load_mass,
    vehicle_vessel_cable_laying.turntable_inner_diameter,
    vehicle_vessel_cable_laying.cable_splice_capabilities,
    vehicle_vessel_cable_laying.dynamic_positioning_capabilities,
    vehicle_vessel_cable_laying.external_personel,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM (((reference.vehicle_vessel_cable_laying
     JOIN reference.vehicle_shared ON ((vehicle_vessel_cable_laying.fk_vehicle_id = vehicle_shared.fk_vehicle_id)))
     JOIN reference.vehicle ON ((vehicle_shared.fk_vehicle_id = vehicle.id)))
     JOIN reference.vehicle_type ON ((vehicle_vessel_cable_laying.fk_vehicle_type_id = vehicle_type.id)))
  WHERE ((vehicle_type.description)::text = 'cable laying vessel'::text);


--
-- TOC entry 439 (class 1259 OID 31746)
-- Name: view_vehicle_vessel_crane_barge; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_vehicle_vessel_crane_barge AS
 SELECT vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    vehicle_vessel_cargo.beam,
    vehicle_vessel_cargo.max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    vehicle_vessel_cargo.deck_space,
    vehicle_vessel_cargo.max_deck_load_pressure,
    vehicle_vessel_cargo.max_cargo_mass,
    vehicle_vessel_cargo.crane_max_load_mass,
    vehicle_vessel_cargo.dynamic_positioning_capabilities,
    vehicle_vessel_cargo.external_personel,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM (((reference.vehicle_vessel_cargo
     JOIN reference.vehicle_shared ON ((vehicle_vessel_cargo.fk_vehicle_id = vehicle_shared.fk_vehicle_id)))
     JOIN reference.vehicle ON ((vehicle_shared.fk_vehicle_id = vehicle.id)))
     JOIN reference.vehicle_type ON ((vehicle_vessel_cargo.fk_vehicle_type_id = vehicle_type.id)))
  WHERE ((vehicle_type.description)::text = 'crane barge'::text);


--
-- TOC entry 440 (class 1259 OID 31751)
-- Name: view_vehicle_vessel_crane_vessel; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_vehicle_vessel_crane_vessel AS
 SELECT vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    vehicle_vessel_cargo.beam,
    vehicle_vessel_cargo.max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    vehicle_vessel_cargo.deck_space,
    vehicle_vessel_cargo.max_deck_load_pressure,
    vehicle_vessel_cargo.max_cargo_mass,
    vehicle_vessel_cargo.crane_max_load_mass,
    vehicle_vessel_cargo.dynamic_positioning_capabilities,
    vehicle_vessel_cargo.external_personel,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM (((reference.vehicle_vessel_cargo
     JOIN reference.vehicle_shared ON ((vehicle_vessel_cargo.fk_vehicle_id = vehicle_shared.fk_vehicle_id)))
     JOIN reference.vehicle ON ((vehicle_shared.fk_vehicle_id = vehicle.id)))
     JOIN reference.vehicle_type ON ((vehicle_vessel_cargo.fk_vehicle_type_id = vehicle_type.id)))
  WHERE ((vehicle_type.description)::text = 'crane vessel'::text);


--
-- TOC entry 441 (class 1259 OID 31756)
-- Name: view_vehicle_vessel_csv; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_vehicle_vessel_csv AS
 SELECT vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    vehicle_vessel_cargo.beam,
    vehicle_vessel_cargo.max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    vehicle_vessel_cargo.deck_space,
    vehicle_vessel_cargo.max_deck_load_pressure,
    vehicle_vessel_cargo.max_cargo_mass,
    vehicle_vessel_cargo.crane_max_load_mass,
    vehicle_vessel_cargo.dynamic_positioning_capabilities,
    vehicle_vessel_cargo.external_personel,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM (((reference.vehicle_vessel_cargo
     JOIN reference.vehicle_shared ON ((vehicle_vessel_cargo.fk_vehicle_id = vehicle_shared.fk_vehicle_id)))
     JOIN reference.vehicle ON ((vehicle_shared.fk_vehicle_id = vehicle.id)))
     JOIN reference.vehicle_type ON ((vehicle_vessel_cargo.fk_vehicle_type_id = vehicle_type.id)))
  WHERE ((vehicle_type.description)::text = 'construction support vessel'::text);


--
-- TOC entry 442 (class 1259 OID 31761)
-- Name: view_vehicle_vessel_ctv; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_vehicle_vessel_ctv AS
 SELECT vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    vehicle_vessel_cargo.beam,
    vehicle_vessel_cargo.max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    vehicle_vessel_cargo.deck_space,
    vehicle_vessel_cargo.max_deck_load_pressure,
    vehicle_vessel_cargo.max_cargo_mass,
    vehicle_vessel_cargo.crane_max_load_mass,
    vehicle_vessel_cargo.dynamic_positioning_capabilities,
    vehicle_vessel_cargo.external_personel,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM (((reference.vehicle_vessel_cargo
     JOIN reference.vehicle_shared ON ((vehicle_vessel_cargo.fk_vehicle_id = vehicle_shared.fk_vehicle_id)))
     JOIN reference.vehicle ON ((vehicle_shared.fk_vehicle_id = vehicle.id)))
     JOIN reference.vehicle_type ON ((vehicle_vessel_cargo.fk_vehicle_type_id = vehicle_type.id)))
  WHERE ((vehicle_type.description)::text = 'crew transfer vessel'::text);


--
-- TOC entry 443 (class 1259 OID 31766)
-- Name: view_vehicle_vessel_jackup_barge; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_vehicle_vessel_jackup_barge AS
 SELECT vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    vehicle_vessel_jackup.beam,
    vehicle_vessel_jackup.max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    vehicle_vessel_jackup.deck_space,
    vehicle_vessel_jackup.max_deck_load_pressure,
    vehicle_vessel_jackup.max_cargo_mass,
    vehicle_vessel_jackup.crane_max_load_mass,
    vehicle_vessel_jackup.dynamic_positioning_capabilities,
    vehicle_vessel_jackup.external_personel,
    vehicle_vessel_jackup.jackup_max_water_depth,
    vehicle_vessel_jackup.jackup_speed_down,
    vehicle_vessel_jackup.jackup_max_payload_mass,
    vehicle_vessel_jackup.jacking_max_hs,
    vehicle_vessel_jackup.jacking_max_tp,
    vehicle_vessel_jackup.jacking_max_cs,
    vehicle_vessel_jackup.jacking_max_ws,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM (((reference.vehicle_vessel_jackup
     JOIN reference.vehicle_shared ON ((vehicle_vessel_jackup.fk_vehicle_id = vehicle_shared.fk_vehicle_id)))
     JOIN reference.vehicle ON ((vehicle_shared.fk_vehicle_id = vehicle.id)))
     JOIN reference.vehicle_type ON ((vehicle_vessel_jackup.fk_vehicle_type_id = vehicle_type.id)))
  WHERE ((vehicle_type.description)::text = 'jackup barge'::text);


--
-- TOC entry 444 (class 1259 OID 31771)
-- Name: view_vehicle_vessel_jackup_vessel; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_vehicle_vessel_jackup_vessel AS
 SELECT vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    vehicle_vessel_jackup.beam,
    vehicle_vessel_jackup.max_draft,
    vehicle_shared.consumption,
    vehicle_shared.transit_speed,
    vehicle_vessel_jackup.deck_space,
    vehicle_vessel_jackup.max_deck_load_pressure,
    vehicle_vessel_jackup.max_cargo_mass,
    vehicle_vessel_jackup.crane_max_load_mass,
    vehicle_vessel_jackup.dynamic_positioning_capabilities,
    vehicle_vessel_jackup.external_personel,
    vehicle_vessel_jackup.jackup_max_water_depth,
    vehicle_vessel_jackup.jackup_speed_down,
    vehicle_vessel_jackup.jackup_max_payload_mass,
    vehicle_vessel_jackup.jacking_max_hs,
    vehicle_vessel_jackup.jacking_max_tp,
    vehicle_vessel_jackup.jacking_max_cs,
    vehicle_vessel_jackup.jacking_max_ws,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM (((reference.vehicle_vessel_jackup
     JOIN reference.vehicle_shared ON ((vehicle_vessel_jackup.fk_vehicle_id = vehicle_shared.fk_vehicle_id)))
     JOIN reference.vehicle ON ((vehicle_shared.fk_vehicle_id = vehicle.id)))
     JOIN reference.vehicle_type ON ((vehicle_vessel_jackup.fk_vehicle_type_id = vehicle_type.id)))
  WHERE ((vehicle_type.description)::text = 'jackup vessel'::text);


--
-- TOC entry 445 (class 1259 OID 31776)
-- Name: view_vehicle_vessel_multicat; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_vehicle_vessel_multicat AS
 SELECT vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    vehicle_vessel_anchor_handling.beam,
    vehicle_vessel_anchor_handling.max_draft,
    vehicle_shared.consumption,
    vehicle_vessel_anchor_handling.consumption_towing,
    vehicle_shared.transit_speed,
    vehicle_vessel_anchor_handling.deck_space,
    vehicle_vessel_anchor_handling.max_deck_load_pressure,
    vehicle_vessel_anchor_handling.max_cargo_mass,
    vehicle_vessel_anchor_handling.crane_max_load_mass,
    vehicle_vessel_anchor_handling.bollard_pull,
    vehicle_vessel_anchor_handling.anchor_handling_drum_capacity,
    vehicle_vessel_anchor_handling.anchor_handling_winch_rated_pull,
    vehicle_vessel_anchor_handling.external_personel,
    vehicle_vessel_anchor_handling.towing_max_hs,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM (((reference.vehicle_vessel_anchor_handling
     JOIN reference.vehicle_shared ON ((vehicle_vessel_anchor_handling.fk_vehicle_id = vehicle_shared.fk_vehicle_id)))
     JOIN reference.vehicle ON ((vehicle_shared.fk_vehicle_id = vehicle.id)))
     JOIN reference.vehicle_type ON ((vehicle_vessel_anchor_handling.fk_vehicle_type_id = vehicle_type.id)))
  WHERE ((vehicle_type.description)::text = 'multicat'::text);


--
-- TOC entry 446 (class 1259 OID 31781)
-- Name: view_vehicle_vessel_tugboat; Type: VIEW; Schema: reference; Owner: -
--

CREATE VIEW reference.view_vehicle_vessel_tugboat AS
 SELECT vehicle.description,
    vehicle_shared.gross_tonnage,
    vehicle_shared.length,
    vehicle_vessel_tugboat.beam,
    vehicle_vessel_tugboat.max_draft,
    vehicle_shared.consumption,
    vehicle_vessel_tugboat.consumption_towing,
    vehicle_shared.transit_speed,
    vehicle_vessel_tugboat.bollard_pull,
    vehicle_shared.transit_max_hs,
    vehicle_shared.transit_max_tp,
    vehicle_shared.transit_max_cs,
    vehicle_shared.transit_max_ws,
    vehicle_shared.mobilisation_time,
    vehicle_shared.mobilisation_percentage_cost,
    vehicle_shared.min_day_rate,
    vehicle_shared.max_day_rate
   FROM ((reference.vehicle_vessel_tugboat
     JOIN reference.vehicle_shared ON ((vehicle_vessel_tugboat.fk_vehicle_id = vehicle_shared.fk_vehicle_id)))
     JOIN reference.vehicle ON ((vehicle_shared.fk_vehicle_id = vehicle.id)));


--
-- TOC entry 4999 (class 2604 OID 31786)
-- Name: bathymetry id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.bathymetry ALTER COLUMN id SET DEFAULT nextval('project.bathymetry_id_seq'::regclass);


--
-- TOC entry 5001 (class 2604 OID 31787)
-- Name: cable_corridor_bathymetry id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.cable_corridor_bathymetry ALTER COLUMN id SET DEFAULT nextval('project.cable_corridor_bathymetry_id_seq'::regclass);


--
-- TOC entry 5002 (class 2604 OID 31788)
-- Name: cable_corridor_bathymetry_layer id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.cable_corridor_bathymetry_layer ALTER COLUMN id SET DEFAULT nextval('project.cable_corridor_bathymetry_layer_id_seq'::regclass);


--
-- TOC entry 5003 (class 2604 OID 31789)
-- Name: cable_corridor_constraint id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.cable_corridor_constraint ALTER COLUMN id SET DEFAULT nextval('project.cable_corridor_constraint_id_seq'::regclass);


--
-- TOC entry 5004 (class 2604 OID 31790)
-- Name: device id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.device ALTER COLUMN id SET DEFAULT nextval('project.device_id_seq'::regclass);


--
-- TOC entry 5006 (class 2604 OID 31791)
-- Name: device_floating id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.device_floating ALTER COLUMN id SET DEFAULT nextval('project.device_floating_id_seq'::regclass);


--
-- TOC entry 5007 (class 2604 OID 31792)
-- Name: device_shared id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.device_shared ALTER COLUMN id SET DEFAULT nextval('project.device_shared_id_seq'::regclass);


--
-- TOC entry 5012 (class 2604 OID 31793)
-- Name: device_tidal id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.device_tidal ALTER COLUMN id SET DEFAULT nextval('project.device_tidal_id_seq'::regclass);


--
-- TOC entry 5013 (class 2604 OID 31794)
-- Name: device_tidal_power_performance id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.device_tidal_power_performance ALTER COLUMN id SET DEFAULT nextval('project.device_tidal_power_performance_id_seq'::regclass);


--
-- TOC entry 5014 (class 2604 OID 31795)
-- Name: device_wave id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.device_wave ALTER COLUMN id SET DEFAULT nextval('project.device_wave_id_seq'::regclass);


--
-- TOC entry 5015 (class 2604 OID 31796)
-- Name: lease_area id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.lease_area ALTER COLUMN id SET DEFAULT nextval('project.lease_area_id_seq'::regclass);


--
-- TOC entry 5018 (class 2604 OID 31797)
-- Name: site id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.site ALTER COLUMN id SET DEFAULT nextval('project.site_id_seq'::regclass);


--
-- TOC entry 4997 (class 2604 OID 31798)
-- Name: sub_systems id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems ALTER COLUMN id SET DEFAULT nextval('project.sub_systems_id_seq'::regclass);


--
-- TOC entry 5019 (class 2604 OID 31799)
-- Name: sub_systems_access id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_access ALTER COLUMN id SET DEFAULT nextval('project.sub_systems_access_id_seq'::regclass);


--
-- TOC entry 5020 (class 2604 OID 31800)
-- Name: sub_systems_economic id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_economic ALTER COLUMN id SET DEFAULT nextval('project.sub_systems_economic_id_seq'::regclass);


--
-- TOC entry 5021 (class 2604 OID 31801)
-- Name: sub_systems_inspection id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_inspection ALTER COLUMN id SET DEFAULT nextval('project.sub_systems_inspection_id_seq'::regclass);


--
-- TOC entry 5022 (class 2604 OID 31802)
-- Name: sub_systems_install id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_install ALTER COLUMN id SET DEFAULT nextval('project.sub_systems_install_id_seq'::regclass);


--
-- TOC entry 5023 (class 2604 OID 31803)
-- Name: sub_systems_maintenance id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_maintenance ALTER COLUMN id SET DEFAULT nextval('project.sub_systems_maintenance_id_seq'::regclass);


--
-- TOC entry 5024 (class 2604 OID 31804)
-- Name: sub_systems_operation_weightings id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_operation_weightings ALTER COLUMN id SET DEFAULT nextval('project.sub_systems_operation_weightings_id_seq'::regclass);


--
-- TOC entry 5025 (class 2604 OID 31805)
-- Name: sub_systems_replace id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_replace ALTER COLUMN id SET DEFAULT nextval('project.sub_systems_replace_id_seq'::regclass);


--
-- TOC entry 5026 (class 2604 OID 31806)
-- Name: time_series_energy_tidal id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.time_series_energy_tidal ALTER COLUMN id SET DEFAULT nextval('project.time_series_energy_tidal_id_seq'::regclass);


--
-- TOC entry 5027 (class 2604 OID 31807)
-- Name: time_series_energy_wave id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.time_series_energy_wave ALTER COLUMN id SET DEFAULT nextval('project.time_series_energy_wave_id_seq'::regclass);


--
-- TOC entry 5028 (class 2604 OID 31808)
-- Name: time_series_om_tidal id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.time_series_om_tidal ALTER COLUMN id SET DEFAULT nextval('project.time_series_om_tidal_id_seq'::regclass);


--
-- TOC entry 5029 (class 2604 OID 31809)
-- Name: time_series_om_wave id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.time_series_om_wave ALTER COLUMN id SET DEFAULT nextval('project.time_series_om_wave_id_seq'::regclass);


--
-- TOC entry 5030 (class 2604 OID 31810)
-- Name: time_series_om_wind id; Type: DEFAULT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.time_series_om_wind ALTER COLUMN id SET DEFAULT nextval('project.time_series_om_wind_id_seq'::regclass);


--
-- TOC entry 5031 (class 2604 OID 31811)
-- Name: component id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component ALTER COLUMN id SET DEFAULT nextval('reference.component_id_seq'::regclass);


--
-- TOC entry 5032 (class 2604 OID 31812)
-- Name: component_anchor id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_anchor ALTER COLUMN id SET DEFAULT nextval('reference.component_anchor_id_seq'::regclass);


--
-- TOC entry 5034 (class 2604 OID 31813)
-- Name: component_cable id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_cable ALTER COLUMN id SET DEFAULT nextval('reference.component_cable_id_seq'::regclass);


--
-- TOC entry 5036 (class 2604 OID 31814)
-- Name: component_collection_point id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_collection_point ALTER COLUMN id SET DEFAULT nextval('reference.component_collection_point_id_seq'::regclass);


--
-- TOC entry 5040 (class 2604 OID 31815)
-- Name: component_connector id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_connector ALTER COLUMN id SET DEFAULT nextval('reference.component_connector_id_seq'::regclass);


--
-- TOC entry 5042 (class 2604 OID 31816)
-- Name: component_continuous id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_continuous ALTER COLUMN id SET DEFAULT nextval('reference.component_continuous_id_seq'::regclass);


--
-- TOC entry 5043 (class 2604 OID 31817)
-- Name: component_discrete id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_discrete ALTER COLUMN id SET DEFAULT nextval('reference.component_discrete_id_seq'::regclass);


--
-- TOC entry 5044 (class 2604 OID 31818)
-- Name: component_mooring_continuous id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_mooring_continuous ALTER COLUMN id SET DEFAULT nextval('reference.component_mooring_continuous_id_seq'::regclass);


--
-- TOC entry 5046 (class 2604 OID 31819)
-- Name: component_mooring_discrete id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_mooring_discrete ALTER COLUMN id SET DEFAULT nextval('reference.component_mooring_discrete_id_seq'::regclass);


--
-- TOC entry 5048 (class 2604 OID 31820)
-- Name: component_pile id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_pile ALTER COLUMN id SET DEFAULT nextval('reference.component_pile_id_seq'::regclass);


--
-- TOC entry 5050 (class 2604 OID 31821)
-- Name: component_rope id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_rope ALTER COLUMN id SET DEFAULT nextval('reference.component_rope_id_seq'::regclass);


--
-- TOC entry 5053 (class 2604 OID 31822)
-- Name: component_shared id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_shared ALTER COLUMN id SET DEFAULT nextval('reference.component_shared_id_seq'::regclass);


--
-- TOC entry 5054 (class 2604 OID 31823)
-- Name: component_transformer id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_transformer ALTER COLUMN id SET DEFAULT nextval('reference.component_transformer_id_seq'::regclass);


--
-- TOC entry 5056 (class 2604 OID 31824)
-- Name: component_type id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_type ALTER COLUMN id SET DEFAULT nextval('reference.component_type_id_seq'::regclass);


--
-- TOC entry 5058 (class 2604 OID 31825)
-- Name: equipment_cable_burial id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_cable_burial ALTER COLUMN id SET DEFAULT nextval('reference.equipment_cable_burial_id_seq'::regclass);


--
-- TOC entry 5059 (class 2604 OID 31826)
-- Name: equipment_divers id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_divers ALTER COLUMN id SET DEFAULT nextval('reference.equipment_divers_id_seq'::regclass);


--
-- TOC entry 5060 (class 2604 OID 31827)
-- Name: equipment_drilling_rigs id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_drilling_rigs ALTER COLUMN id SET DEFAULT nextval('reference.equipment_drilling_rigs_id_seq'::regclass);


--
-- TOC entry 5061 (class 2604 OID 31828)
-- Name: equipment_excavating id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_excavating ALTER COLUMN id SET DEFAULT nextval('reference.equipment_excavating_id_seq'::regclass);


--
-- TOC entry 5062 (class 2604 OID 31829)
-- Name: equipment_hammer id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_hammer ALTER COLUMN id SET DEFAULT nextval('reference.equipment_hammer_id_seq'::regclass);


--
-- TOC entry 5063 (class 2604 OID 31830)
-- Name: equipment_mattress id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_mattress ALTER COLUMN id SET DEFAULT nextval('reference.equipment_mattress_id_seq'::regclass);


--
-- TOC entry 5064 (class 2604 OID 31831)
-- Name: equipment_rock_filter_bags id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_rock_filter_bags ALTER COLUMN id SET DEFAULT nextval('reference.equipment_rock_filter_bags_id_seq'::regclass);


--
-- TOC entry 5065 (class 2604 OID 31832)
-- Name: equipment_rov id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_rov ALTER COLUMN id SET DEFAULT nextval('reference.equipment_rov_id_seq'::regclass);


--
-- TOC entry 5067 (class 2604 OID 31833)
-- Name: equipment_split_pipe id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_split_pipe ALTER COLUMN id SET DEFAULT nextval('reference.equipment_split_pipe_id_seq'::regclass);


--
-- TOC entry 5068 (class 2604 OID 31834)
-- Name: equipment_vibro_driver id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_vibro_driver ALTER COLUMN id SET DEFAULT nextval('reference.equipment_vibro_driver_id_seq'::regclass);


--
-- TOC entry 5069 (class 2604 OID 31835)
-- Name: operations_limit_cs id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.operations_limit_cs ALTER COLUMN id SET DEFAULT nextval('reference.operations_limit_cs_id_seq'::regclass);


--
-- TOC entry 5070 (class 2604 OID 31836)
-- Name: operations_limit_hs id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.operations_limit_hs ALTER COLUMN id SET DEFAULT nextval('reference.operations_limit_hs_id_seq'::regclass);


--
-- TOC entry 5071 (class 2604 OID 31837)
-- Name: operations_limit_tp id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.operations_limit_tp ALTER COLUMN id SET DEFAULT nextval('reference.operations_limit_tp_id_seq'::regclass);


--
-- TOC entry 5072 (class 2604 OID 31838)
-- Name: operations_limit_ws id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.operations_limit_ws ALTER COLUMN id SET DEFAULT nextval('reference.operations_limit_ws_id_seq'::regclass);


--
-- TOC entry 5073 (class 2604 OID 31839)
-- Name: operations_type id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.operations_type ALTER COLUMN id SET DEFAULT nextval('reference.operations_type_id_seq'::regclass);


--
-- TOC entry 5074 (class 2604 OID 31840)
-- Name: ports id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ports ALTER COLUMN id SET DEFAULT nextval('reference.ports_id_seq'::regclass);


--
-- TOC entry 4996 (class 2604 OID 31841)
-- Name: soil_type id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.soil_type ALTER COLUMN id SET DEFAULT nextval('reference.soil_type_id_seq'::regclass);


--
-- TOC entry 5076 (class 2604 OID 31842)
-- Name: vehicle id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle ALTER COLUMN id SET DEFAULT nextval('reference.vehicle_id_seq'::regclass);


--
-- TOC entry 5077 (class 2604 OID 31843)
-- Name: vehicle_helicopter id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_helicopter ALTER COLUMN id SET DEFAULT nextval('reference.vehicle_helicopter_id_seq'::regclass);


--
-- TOC entry 5079 (class 2604 OID 31844)
-- Name: vehicle_shared id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_shared ALTER COLUMN id SET DEFAULT nextval('reference.vehicle_shared_id_seq'::regclass);


--
-- TOC entry 5080 (class 2604 OID 31845)
-- Name: vehicle_type id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_type ALTER COLUMN id SET DEFAULT nextval('reference.vehicle_type_id_seq'::regclass);


--
-- TOC entry 5081 (class 2604 OID 31846)
-- Name: vehicle_vessel_anchor_handling id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_anchor_handling ALTER COLUMN id SET DEFAULT nextval('reference.vehicle_vessel_anchor_handling_id_seq'::regclass);


--
-- TOC entry 5083 (class 2604 OID 31847)
-- Name: vehicle_vessel_cable_laying id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_cable_laying ALTER COLUMN id SET DEFAULT nextval('reference.vehicle_vessel_cable_laying_id_seq'::regclass);


--
-- TOC entry 5085 (class 2604 OID 31848)
-- Name: vehicle_vessel_cargo id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_cargo ALTER COLUMN id SET DEFAULT nextval('reference.vehicle_vessel_cargo_id_seq'::regclass);


--
-- TOC entry 5087 (class 2604 OID 31849)
-- Name: vehicle_vessel_jackup id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_jackup ALTER COLUMN id SET DEFAULT nextval('reference.vehicle_vessel_jackup_id_seq'::regclass);


--
-- TOC entry 5089 (class 2604 OID 31850)
-- Name: vehicle_vessel_tugboat id; Type: DEFAULT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_tugboat ALTER COLUMN id SET DEFAULT nextval('reference.vehicle_vessel_tugboat_id_seq'::regclass);


--
-- TOC entry 5102 (class 2606 OID 31852)
-- Name: bathymetry_layer bathymetry_layer_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.bathymetry_layer
    ADD CONSTRAINT bathymetry_layer_pkey PRIMARY KEY (id);


--
-- TOC entry 5098 (class 2606 OID 31854)
-- Name: bathymetry bathymetry_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.bathymetry
    ADD CONSTRAINT bathymetry_pkey PRIMARY KEY (id);


--
-- TOC entry 5109 (class 2606 OID 31856)
-- Name: cable_corridor_bathymetry_layer cable_corridor_bathymetry_layer_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.cable_corridor_bathymetry_layer
    ADD CONSTRAINT cable_corridor_bathymetry_layer_pkey PRIMARY KEY (id);


--
-- TOC entry 5105 (class 2606 OID 31858)
-- Name: cable_corridor_bathymetry cable_corridor_bathymetry_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.cable_corridor_bathymetry
    ADD CONSTRAINT cable_corridor_bathymetry_pkey PRIMARY KEY (id);


--
-- TOC entry 5112 (class 2606 OID 31860)
-- Name: cable_corridor_constraint cable_corridor_constraint_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.cable_corridor_constraint
    ADD CONSTRAINT cable_corridor_constraint_pkey PRIMARY KEY (id);


--
-- TOC entry 5115 (class 2606 OID 31862)
-- Name: constraint constraint_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project."constraint"
    ADD CONSTRAINT constraint_pkey PRIMARY KEY (id);


--
-- TOC entry 5120 (class 2606 OID 31864)
-- Name: device_floating device_floating_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.device_floating
    ADD CONSTRAINT device_floating_pkey PRIMARY KEY (id);


--
-- TOC entry 5117 (class 2606 OID 31866)
-- Name: device device_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.device
    ADD CONSTRAINT device_pkey PRIMARY KEY (id);


--
-- TOC entry 5123 (class 2606 OID 31868)
-- Name: device_shared device_shared_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.device_shared
    ADD CONSTRAINT device_shared_pkey PRIMARY KEY (id);


--
-- TOC entry 5126 (class 2606 OID 31870)
-- Name: device_tidal device_tidal_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.device_tidal
    ADD CONSTRAINT device_tidal_pkey PRIMARY KEY (id);


--
-- TOC entry 5129 (class 2606 OID 31872)
-- Name: device_tidal_power_performance device_tidal_power_performance_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.device_tidal_power_performance
    ADD CONSTRAINT device_tidal_power_performance_pkey PRIMARY KEY (id);


--
-- TOC entry 5132 (class 2606 OID 31874)
-- Name: device_wave device_wave_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.device_wave
    ADD CONSTRAINT device_wave_pkey PRIMARY KEY (id);


--
-- TOC entry 5135 (class 2606 OID 31876)
-- Name: lease_area lease_area_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.lease_area
    ADD CONSTRAINT lease_area_pkey PRIMARY KEY (id);


--
-- TOC entry 5137 (class 2606 OID 31878)
-- Name: site site_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.site
    ADD CONSTRAINT site_pkey PRIMARY KEY (id);


--
-- TOC entry 5140 (class 2606 OID 31880)
-- Name: sub_systems_access sub_systems_access_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_access
    ADD CONSTRAINT sub_systems_access_pkey PRIMARY KEY (id);


--
-- TOC entry 5143 (class 2606 OID 31882)
-- Name: sub_systems_economic sub_systems_economic_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_economic
    ADD CONSTRAINT sub_systems_economic_pkey PRIMARY KEY (id);


--
-- TOC entry 5146 (class 2606 OID 31884)
-- Name: sub_systems_inspection sub_systems_inspection_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_inspection
    ADD CONSTRAINT sub_systems_inspection_pkey PRIMARY KEY (id);


--
-- TOC entry 5149 (class 2606 OID 31886)
-- Name: sub_systems_install sub_systems_install_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_install
    ADD CONSTRAINT sub_systems_install_pkey PRIMARY KEY (id);


--
-- TOC entry 5152 (class 2606 OID 31888)
-- Name: sub_systems_maintenance sub_systems_maintenance_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_maintenance
    ADD CONSTRAINT sub_systems_maintenance_pkey PRIMARY KEY (id);


--
-- TOC entry 5155 (class 2606 OID 31890)
-- Name: sub_systems_operation_weightings sub_systems_operation_weightings_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_operation_weightings
    ADD CONSTRAINT sub_systems_operation_weightings_pkey PRIMARY KEY (id);


--
-- TOC entry 5095 (class 2606 OID 31892)
-- Name: sub_systems sub_systems_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems
    ADD CONSTRAINT sub_systems_pkey PRIMARY KEY (id);


--
-- TOC entry 5158 (class 2606 OID 31894)
-- Name: sub_systems_replace sub_systems_replace_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_replace
    ADD CONSTRAINT sub_systems_replace_pkey PRIMARY KEY (id);


--
-- TOC entry 5161 (class 2606 OID 31896)
-- Name: time_series_energy_tidal time_series_energy_tidal_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.time_series_energy_tidal
    ADD CONSTRAINT time_series_energy_tidal_pkey PRIMARY KEY (id);


--
-- TOC entry 5164 (class 2606 OID 31898)
-- Name: time_series_energy_wave time_series_energy_wave_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.time_series_energy_wave
    ADD CONSTRAINT time_series_energy_wave_pkey PRIMARY KEY (id);


--
-- TOC entry 5167 (class 2606 OID 31900)
-- Name: time_series_om_tidal time_series_om_tidal_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.time_series_om_tidal
    ADD CONSTRAINT time_series_om_tidal_pkey PRIMARY KEY (id);


--
-- TOC entry 5170 (class 2606 OID 31902)
-- Name: time_series_om_wave time_series_om_wave_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.time_series_om_wave
    ADD CONSTRAINT time_series_om_wave_pkey PRIMARY KEY (id);


--
-- TOC entry 5173 (class 2606 OID 31904)
-- Name: time_series_om_wind time_series_om_wind_pkey; Type: CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.time_series_om_wind
    ADD CONSTRAINT time_series_om_wind_pkey PRIMARY KEY (id);


--
-- TOC entry 5179 (class 2606 OID 31906)
-- Name: component_anchor component_anchor_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_anchor
    ADD CONSTRAINT component_anchor_pkey PRIMARY KEY (id);


--
-- TOC entry 5183 (class 2606 OID 31908)
-- Name: component_cable component_cable_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_cable
    ADD CONSTRAINT component_cable_pkey PRIMARY KEY (id);


--
-- TOC entry 5187 (class 2606 OID 31910)
-- Name: component_collection_point component_collection_point_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_collection_point
    ADD CONSTRAINT component_collection_point_pkey PRIMARY KEY (id);


--
-- TOC entry 5191 (class 2606 OID 31912)
-- Name: component_connector component_connector_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_connector
    ADD CONSTRAINT component_connector_pkey PRIMARY KEY (id);


--
-- TOC entry 5194 (class 2606 OID 31914)
-- Name: component_continuous component_continuous_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_continuous
    ADD CONSTRAINT component_continuous_pkey PRIMARY KEY (id);


--
-- TOC entry 5197 (class 2606 OID 31916)
-- Name: component_discrete component_discrete_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_discrete
    ADD CONSTRAINT component_discrete_pkey PRIMARY KEY (id);


--
-- TOC entry 5201 (class 2606 OID 31918)
-- Name: component_mooring_continuous component_mooring_continuous_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_mooring_continuous
    ADD CONSTRAINT component_mooring_continuous_pkey PRIMARY KEY (id);


--
-- TOC entry 5205 (class 2606 OID 31920)
-- Name: component_mooring_discrete component_mooring_discrete_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_mooring_discrete
    ADD CONSTRAINT component_mooring_discrete_pkey PRIMARY KEY (id);


--
-- TOC entry 5209 (class 2606 OID 31922)
-- Name: component_pile component_pile_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_pile
    ADD CONSTRAINT component_pile_pkey PRIMARY KEY (id);


--
-- TOC entry 5175 (class 2606 OID 31924)
-- Name: component component_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component
    ADD CONSTRAINT component_pkey PRIMARY KEY (id);


--
-- TOC entry 5213 (class 2606 OID 31926)
-- Name: component_rope component_rope_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_rope
    ADD CONSTRAINT component_rope_pkey PRIMARY KEY (id);


--
-- TOC entry 5216 (class 2606 OID 31928)
-- Name: component_shared component_shared_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_shared
    ADD CONSTRAINT component_shared_pkey PRIMARY KEY (id);


--
-- TOC entry 5220 (class 2606 OID 31930)
-- Name: component_transformer component_transformer_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_transformer
    ADD CONSTRAINT component_transformer_pkey PRIMARY KEY (id);


--
-- TOC entry 5222 (class 2606 OID 31932)
-- Name: component_type component_type_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_type
    ADD CONSTRAINT component_type_pkey PRIMARY KEY (id);


--
-- TOC entry 5224 (class 2606 OID 31934)
-- Name: constants constants_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.constants
    ADD CONSTRAINT constants_pkey PRIMARY KEY (lock);


--
-- TOC entry 5226 (class 2606 OID 31936)
-- Name: equipment_cable_burial equipment_cable_burial_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_cable_burial
    ADD CONSTRAINT equipment_cable_burial_pkey PRIMARY KEY (id);


--
-- TOC entry 5228 (class 2606 OID 31938)
-- Name: equipment_divers equipment_divers_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_divers
    ADD CONSTRAINT equipment_divers_pkey PRIMARY KEY (id);


--
-- TOC entry 5230 (class 2606 OID 31940)
-- Name: equipment_drilling_rigs equipment_drilling_rigs_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_drilling_rigs
    ADD CONSTRAINT equipment_drilling_rigs_pkey PRIMARY KEY (id);


--
-- TOC entry 5232 (class 2606 OID 31942)
-- Name: equipment_excavating equipment_excavating_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_excavating
    ADD CONSTRAINT equipment_excavating_pkey PRIMARY KEY (id);


--
-- TOC entry 5234 (class 2606 OID 31944)
-- Name: equipment_hammer equipment_hammer_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_hammer
    ADD CONSTRAINT equipment_hammer_pkey PRIMARY KEY (id);


--
-- TOC entry 5236 (class 2606 OID 31946)
-- Name: equipment_mattress equipment_mattress_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_mattress
    ADD CONSTRAINT equipment_mattress_pkey PRIMARY KEY (id);


--
-- TOC entry 5238 (class 2606 OID 31948)
-- Name: equipment_rock_filter_bags equipment_rock_filter_bags_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_rock_filter_bags
    ADD CONSTRAINT equipment_rock_filter_bags_pkey PRIMARY KEY (id);


--
-- TOC entry 5240 (class 2606 OID 31950)
-- Name: equipment_rov equipment_rov_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_rov
    ADD CONSTRAINT equipment_rov_pkey PRIMARY KEY (id);


--
-- TOC entry 5242 (class 2606 OID 31952)
-- Name: equipment_soil_lay_rates equipment_soil_lay_rates_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_soil_lay_rates
    ADD CONSTRAINT equipment_soil_lay_rates_pkey PRIMARY KEY (equipment_type);


--
-- TOC entry 5244 (class 2606 OID 31954)
-- Name: equipment_soil_penet_rates equipment_soil_penet_rates_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_soil_penet_rates
    ADD CONSTRAINT equipment_soil_penet_rates_pkey PRIMARY KEY (equipment_type);


--
-- TOC entry 5246 (class 2606 OID 31956)
-- Name: equipment_split_pipe equipment_split_pipe_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_split_pipe
    ADD CONSTRAINT equipment_split_pipe_pkey PRIMARY KEY (id);


--
-- TOC entry 5248 (class 2606 OID 31958)
-- Name: equipment_vibro_driver equipment_vibro_driver_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.equipment_vibro_driver
    ADD CONSTRAINT equipment_vibro_driver_pkey PRIMARY KEY (id);


--
-- TOC entry 5250 (class 2606 OID 31960)
-- Name: operations_limit_cs operations_limit_cs_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.operations_limit_cs
    ADD CONSTRAINT operations_limit_cs_pkey PRIMARY KEY (id);


--
-- TOC entry 5252 (class 2606 OID 31962)
-- Name: operations_limit_hs operations_limit_hs_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.operations_limit_hs
    ADD CONSTRAINT operations_limit_hs_pkey PRIMARY KEY (id);


--
-- TOC entry 5254 (class 2606 OID 31964)
-- Name: operations_limit_tp operations_limit_tp_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.operations_limit_tp
    ADD CONSTRAINT operations_limit_tp_pkey PRIMARY KEY (id);


--
-- TOC entry 5256 (class 2606 OID 31966)
-- Name: operations_limit_ws operations_limit_ws_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.operations_limit_ws
    ADD CONSTRAINT operations_limit_ws_pkey PRIMARY KEY (id);


--
-- TOC entry 5258 (class 2606 OID 31968)
-- Name: operations_type operations_type_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.operations_type
    ADD CONSTRAINT operations_type_pkey PRIMARY KEY (id);


--
-- TOC entry 5260 (class 2606 OID 31970)
-- Name: ports ports_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ports
    ADD CONSTRAINT ports_pkey PRIMARY KEY (id);


--
-- TOC entry 5262 (class 2606 OID 31972)
-- Name: ref_current_drag_coef_rect ref_current_drag_coef_rect_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ref_current_drag_coef_rect
    ADD CONSTRAINT ref_current_drag_coef_rect_pkey PRIMARY KEY (width_length);


--
-- TOC entry 5264 (class 2606 OID 31974)
-- Name: ref_drag_coef_cyl ref_drag_coef_cyl_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ref_drag_coef_cyl
    ADD CONSTRAINT ref_drag_coef_cyl_pkey PRIMARY KEY (reynolds_number);


--
-- TOC entry 5266 (class 2606 OID 31976)
-- Name: ref_drift_coef_float_rect ref_drift_coef_float_rect_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ref_drift_coef_float_rect
    ADD CONSTRAINT ref_drift_coef_float_rect_pkey PRIMARY KEY (wavenumber_draft);


--
-- TOC entry 5268 (class 2606 OID 31978)
-- Name: ref_holding_capacity_factors_plate_anchors ref_holding_capacity_factors_plate_anchors_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ref_holding_capacity_factors_plate_anchors
    ADD CONSTRAINT ref_holding_capacity_factors_plate_anchors_pkey PRIMARY KEY (relative_embedment_depth);


--
-- TOC entry 5270 (class 2606 OID 31980)
-- Name: ref_line_bcf ref_line_bcf_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ref_line_bcf
    ADD CONSTRAINT ref_line_bcf_pkey PRIMARY KEY (soil_friction_angle);


--
-- TOC entry 5272 (class 2606 OID 31982)
-- Name: ref_pile_deflection_coefficients ref_pile_deflection_coefficients_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ref_pile_deflection_coefficients
    ADD CONSTRAINT ref_pile_deflection_coefficients_pkey PRIMARY KEY (depth_coefficient);


--
-- TOC entry 5274 (class 2606 OID 31984)
-- Name: ref_pile_limiting_values_noncalcareous ref_pile_limiting_values_noncalcareous_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ref_pile_limiting_values_noncalcareous
    ADD CONSTRAINT ref_pile_limiting_values_noncalcareous_pkey PRIMARY KEY (soil_friction_angle);


--
-- TOC entry 5276 (class 2606 OID 31986)
-- Name: ref_pile_moment_coefficient_sam ref_pile_moment_coefficient_sam_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ref_pile_moment_coefficient_sam
    ADD CONSTRAINT ref_pile_moment_coefficient_sam_pkey PRIMARY KEY (depth_coefficient);


--
-- TOC entry 5278 (class 2606 OID 31988)
-- Name: ref_pile_moment_coefficient_sbm ref_pile_moment_coefficient_sbm_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ref_pile_moment_coefficient_sbm
    ADD CONSTRAINT ref_pile_moment_coefficient_sbm_pkey PRIMARY KEY (depth_coefficient);


--
-- TOC entry 5280 (class 2606 OID 31990)
-- Name: ref_rectangular_wave_inertia ref_rectangular_wave_inertia_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ref_rectangular_wave_inertia
    ADD CONSTRAINT ref_rectangular_wave_inertia_pkey PRIMARY KEY ("width/length");


--
-- TOC entry 5282 (class 2606 OID 31992)
-- Name: ref_subgrade_reaction_coefficient_cohesionless ref_subgrade_reaction_coefficient_cohesionless_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ref_subgrade_reaction_coefficient_cohesionless
    ADD CONSTRAINT ref_subgrade_reaction_coefficient_cohesionless_pkey PRIMARY KEY (allowable_deflection_diameter);


--
-- TOC entry 5284 (class 2606 OID 31994)
-- Name: ref_subgrade_reaction_coefficient_k1_cohesive ref_subgrade_reaction_coefficient_k1_cohesive_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ref_subgrade_reaction_coefficient_k1_cohesive
    ADD CONSTRAINT ref_subgrade_reaction_coefficient_k1_cohesive_pkey PRIMARY KEY (allowable_deflection_diameter);


--
-- TOC entry 5286 (class 2606 OID 31996)
-- Name: ref_superline_nylon ref_superline_nylon_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ref_superline_nylon
    ADD CONSTRAINT ref_superline_nylon_pkey PRIMARY KEY (extension);


--
-- TOC entry 5288 (class 2606 OID 31998)
-- Name: ref_superline_polyester ref_superline_polyester_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ref_superline_polyester
    ADD CONSTRAINT ref_superline_polyester_pkey PRIMARY KEY (extension);


--
-- TOC entry 5290 (class 2606 OID 32000)
-- Name: ref_superline_steelite ref_superline_steelite_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ref_superline_steelite
    ADD CONSTRAINT ref_superline_steelite_pkey PRIMARY KEY (extension);


--
-- TOC entry 5292 (class 2606 OID 32002)
-- Name: ref_wake_amplification_factor_cyl ref_wake_amplification_factor_cyl_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ref_wake_amplification_factor_cyl
    ADD CONSTRAINT ref_wake_amplification_factor_cyl_pkey PRIMARY KEY (kc_steady_drag_coefficient);


--
-- TOC entry 5294 (class 2606 OID 32004)
-- Name: ref_wind_drag_coef_rect ref_wind_drag_coef_rect_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.ref_wind_drag_coef_rect
    ADD CONSTRAINT ref_wind_drag_coef_rect_pkey PRIMARY KEY (width_length);


--
-- TOC entry 5296 (class 2606 OID 32006)
-- Name: soil_type_geotechnical_properties soil_type_geotechnical_properties_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.soil_type_geotechnical_properties
    ADD CONSTRAINT soil_type_geotechnical_properties_pkey PRIMARY KEY (fk_soil_type_id);


--
-- TOC entry 5092 (class 2606 OID 32008)
-- Name: soil_type soil_type_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.soil_type
    ADD CONSTRAINT soil_type_pkey PRIMARY KEY (id);


--
-- TOC entry 5302 (class 2606 OID 32010)
-- Name: vehicle_helicopter vehicle_helicopter_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_helicopter
    ADD CONSTRAINT vehicle_helicopter_pkey PRIMARY KEY (id);


--
-- TOC entry 5298 (class 2606 OID 32012)
-- Name: vehicle vehicle_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle
    ADD CONSTRAINT vehicle_pkey PRIMARY KEY (id);


--
-- TOC entry 5305 (class 2606 OID 32014)
-- Name: vehicle_shared vehicle_shared_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_shared
    ADD CONSTRAINT vehicle_shared_pkey PRIMARY KEY (id);


--
-- TOC entry 5307 (class 2606 OID 32016)
-- Name: vehicle_type vehicle_type_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_type
    ADD CONSTRAINT vehicle_type_pkey PRIMARY KEY (id);


--
-- TOC entry 5311 (class 2606 OID 32018)
-- Name: vehicle_vessel_anchor_handling vehicle_vessel_anchor_handling_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_anchor_handling
    ADD CONSTRAINT vehicle_vessel_anchor_handling_pkey PRIMARY KEY (id);


--
-- TOC entry 5315 (class 2606 OID 32020)
-- Name: vehicle_vessel_cable_laying vehicle_vessel_cable_laying_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_cable_laying
    ADD CONSTRAINT vehicle_vessel_cable_laying_pkey PRIMARY KEY (id);


--
-- TOC entry 5319 (class 2606 OID 32022)
-- Name: vehicle_vessel_cargo vehicle_vessel_cargo_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_cargo
    ADD CONSTRAINT vehicle_vessel_cargo_pkey PRIMARY KEY (id);


--
-- TOC entry 5323 (class 2606 OID 32024)
-- Name: vehicle_vessel_jackup vehicle_vessel_jackup_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_jackup
    ADD CONSTRAINT vehicle_vessel_jackup_pkey PRIMARY KEY (id);


--
-- TOC entry 5327 (class 2606 OID 32026)
-- Name: vehicle_vessel_tugboat vehicle_vessel_tugboat_pkey; Type: CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_tugboat
    ADD CONSTRAINT vehicle_vessel_tugboat_pkey PRIMARY KEY (id);


--
-- TOC entry 5096 (class 1259 OID 32027)
-- Name: bathymetry_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX bathymetry_fk_idx ON project.bathymetry USING btree (fk_site_id);


--
-- TOC entry 5099 (class 1259 OID 32028)
-- Name: bathymetry_layer_fk_bathymetry_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX bathymetry_layer_fk_bathymetry_idx ON project.bathymetry_layer USING btree (fk_bathymetry_id);


--
-- TOC entry 5100 (class 1259 OID 32029)
-- Name: bathymetry_layer_fk_soil_type_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX bathymetry_layer_fk_soil_type_idx ON project.bathymetry_layer USING btree (fk_soil_type_id);


--
-- TOC entry 5103 (class 1259 OID 32030)
-- Name: cable_corridor_bathymetry_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX cable_corridor_bathymetry_fk_idx ON project.cable_corridor_bathymetry USING btree (fk_site_id);


--
-- TOC entry 5106 (class 1259 OID 32031)
-- Name: cable_corridor_bathymetry_fk_soil_type_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX cable_corridor_bathymetry_fk_soil_type_idx ON project.cable_corridor_bathymetry_layer USING btree (fk_soil_type_id);


--
-- TOC entry 5107 (class 1259 OID 32032)
-- Name: cable_corridor_bathymetry_layer_fk_bathymetry_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX cable_corridor_bathymetry_layer_fk_bathymetry_idx ON project.cable_corridor_bathymetry_layer USING btree (fk_bathymetry_id);


--
-- TOC entry 5110 (class 1259 OID 32033)
-- Name: cable_corridor_constraint_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX cable_corridor_constraint_fk_idx ON project.cable_corridor_constraint USING btree (fk_site_id);


--
-- TOC entry 5113 (class 1259 OID 32034)
-- Name: constraint_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX constraint_fk_idx ON project."constraint" USING btree (fk_site_id);


--
-- TOC entry 5118 (class 1259 OID 32035)
-- Name: device_floating_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX device_floating_fk_idx ON project.device_floating USING btree (fk_device_id);


--
-- TOC entry 5121 (class 1259 OID 32036)
-- Name: device_shared_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX device_shared_fk_idx ON project.device_shared USING btree (fk_device_id);


--
-- TOC entry 5124 (class 1259 OID 32037)
-- Name: device_tidal_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX device_tidal_fk_idx ON project.device_tidal USING btree (fk_device_id);


--
-- TOC entry 5127 (class 1259 OID 32038)
-- Name: device_tidal_power_performance_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX device_tidal_power_performance_fk_idx ON project.device_tidal_power_performance USING btree (fk_device_id);


--
-- TOC entry 5130 (class 1259 OID 32039)
-- Name: device_wave_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX device_wave_fk_idx ON project.device_wave USING btree (fk_device_id);


--
-- TOC entry 5133 (class 1259 OID 32040)
-- Name: lease_area_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX lease_area_fk_idx ON project.lease_area USING btree (fk_site_id);


--
-- TOC entry 5138 (class 1259 OID 32041)
-- Name: sub_systems_access_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX sub_systems_access_fk_idx ON project.sub_systems_access USING btree (fk_sub_system_id);


--
-- TOC entry 5141 (class 1259 OID 32042)
-- Name: sub_systems_economic_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX sub_systems_economic_fk_idx ON project.sub_systems_economic USING btree (fk_sub_system_id);


--
-- TOC entry 5093 (class 1259 OID 32043)
-- Name: sub_systems_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX sub_systems_fk_idx ON project.sub_systems USING btree (fk_device_id);


--
-- TOC entry 5144 (class 1259 OID 32044)
-- Name: sub_systems_inspection_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX sub_systems_inspection_fk_idx ON project.sub_systems_inspection USING btree (fk_sub_system_id);


--
-- TOC entry 5147 (class 1259 OID 32045)
-- Name: sub_systems_install_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX sub_systems_install_fk_idx ON project.sub_systems_install USING btree (fk_sub_system_id);


--
-- TOC entry 5150 (class 1259 OID 32046)
-- Name: sub_systems_maintenance_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX sub_systems_maintenance_fk_idx ON project.sub_systems_maintenance USING btree (fk_sub_system_id);


--
-- TOC entry 5153 (class 1259 OID 32047)
-- Name: sub_systems_operation_weightings_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX sub_systems_operation_weightings_fk_idx ON project.sub_systems_operation_weightings USING btree (fk_sub_system_id);


--
-- TOC entry 5156 (class 1259 OID 32048)
-- Name: sub_systems_replace_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX sub_systems_replace_fk_idx ON project.sub_systems_replace USING btree (fk_sub_system_id);


--
-- TOC entry 5159 (class 1259 OID 32049)
-- Name: time_series_energy_tidal_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX time_series_energy_tidal_fk_idx ON project.time_series_energy_tidal USING btree (fk_bathymetry_id);


--
-- TOC entry 5162 (class 1259 OID 32050)
-- Name: time_series_energy_wave_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX time_series_energy_wave_fk_idx ON project.time_series_energy_wave USING btree (fk_site_id);


--
-- TOC entry 5165 (class 1259 OID 32051)
-- Name: time_series_om_tidal_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX time_series_om_tidal_fk_idx ON project.time_series_om_tidal USING btree (fk_site_id);


--
-- TOC entry 5168 (class 1259 OID 32052)
-- Name: time_series_om_wave_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX time_series_om_wave_fk_idx ON project.time_series_om_wave USING btree (fk_site_id);


--
-- TOC entry 5171 (class 1259 OID 32053)
-- Name: time_series_om_wind_fk_idx; Type: INDEX; Schema: project; Owner: -
--

CREATE INDEX time_series_om_wind_fk_idx ON project.time_series_om_wind USING btree (fk_site_id);


--
-- TOC entry 5176 (class 1259 OID 32054)
-- Name: component_anchor_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_anchor_fk_idx ON reference.component_anchor USING btree (fk_component_discrete_id);


--
-- TOC entry 5177 (class 1259 OID 32055)
-- Name: component_anchor_fk_type_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_anchor_fk_type_idx ON reference.component_anchor USING btree (fk_component_type_id);


--
-- TOC entry 5180 (class 1259 OID 32056)
-- Name: component_cable_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_cable_fk_idx ON reference.component_cable USING btree (fk_component_continuous_id);


--
-- TOC entry 5181 (class 1259 OID 32057)
-- Name: component_cable_fk_type_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_cable_fk_type_idx ON reference.component_cable USING btree (fk_component_type_id);


--
-- TOC entry 5184 (class 1259 OID 32058)
-- Name: component_collection_point_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_collection_point_fk_idx ON reference.component_collection_point USING btree (fk_component_discrete_id);


--
-- TOC entry 5185 (class 1259 OID 32059)
-- Name: component_collection_point_fk_type_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_collection_point_fk_type_idx ON reference.component_collection_point USING btree (fk_component_type_id);


--
-- TOC entry 5202 (class 1259 OID 32060)
-- Name: component_component_mooring_discrete_fk_type_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_component_mooring_discrete_fk_type_idx ON reference.component_mooring_discrete USING btree (fk_component_type_id);


--
-- TOC entry 5188 (class 1259 OID 32061)
-- Name: component_connector_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_connector_fk_idx ON reference.component_connector USING btree (fk_component_discrete_id);


--
-- TOC entry 5189 (class 1259 OID 32062)
-- Name: component_connector_fk_type_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_connector_fk_type_idx ON reference.component_connector USING btree (fk_component_type_id);


--
-- TOC entry 5192 (class 1259 OID 32063)
-- Name: component_continuous_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_continuous_fk_idx ON reference.component_continuous USING btree (fk_component_id);


--
-- TOC entry 5195 (class 1259 OID 32064)
-- Name: component_discrete_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_discrete_fk_idx ON reference.component_discrete USING btree (fk_component_id);


--
-- TOC entry 5198 (class 1259 OID 32065)
-- Name: component_mooring_continuous_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_mooring_continuous_fk_idx ON reference.component_mooring_continuous USING btree (fk_component_continuous_id);


--
-- TOC entry 5199 (class 1259 OID 32066)
-- Name: component_mooring_continuous_fk_type_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_mooring_continuous_fk_type_idx ON reference.component_mooring_continuous USING btree (fk_component_type_id);


--
-- TOC entry 5203 (class 1259 OID 32067)
-- Name: component_mooring_discrete_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_mooring_discrete_fk_idx ON reference.component_mooring_discrete USING btree (fk_component_discrete_id);


--
-- TOC entry 5206 (class 1259 OID 32068)
-- Name: component_pile_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_pile_fk_idx ON reference.component_pile USING btree (fk_component_continuous_id);


--
-- TOC entry 5207 (class 1259 OID 32069)
-- Name: component_pile_fk_type_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_pile_fk_type_idx ON reference.component_pile USING btree (fk_component_type_id);


--
-- TOC entry 5210 (class 1259 OID 32070)
-- Name: component_rope_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_rope_fk_idx ON reference.component_rope USING btree (fk_component_continuous_id);


--
-- TOC entry 5211 (class 1259 OID 32071)
-- Name: component_rope_fk_type_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_rope_fk_type_idx ON reference.component_rope USING btree (fk_component_type_id);


--
-- TOC entry 5214 (class 1259 OID 32072)
-- Name: component_shared_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_shared_fk_idx ON reference.component_shared USING btree (fk_component_id);


--
-- TOC entry 5217 (class 1259 OID 32073)
-- Name: component_transformer_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_transformer_fk_idx ON reference.component_transformer USING btree (fk_component_discrete_id);


--
-- TOC entry 5218 (class 1259 OID 32074)
-- Name: component_transformer_fk_type_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX component_transformer_fk_type_idx ON reference.component_transformer USING btree (fk_component_type_id);


--
-- TOC entry 5299 (class 1259 OID 32075)
-- Name: vehicle_helicopter_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX vehicle_helicopter_fk_idx ON reference.vehicle_helicopter USING btree (fk_vehicle_id);


--
-- TOC entry 5300 (class 1259 OID 32076)
-- Name: vehicle_helicopter_fk_type_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX vehicle_helicopter_fk_type_idx ON reference.vehicle_helicopter USING btree (fk_vehicle_type_id);


--
-- TOC entry 5303 (class 1259 OID 32077)
-- Name: vehicle_shared_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX vehicle_shared_fk_idx ON reference.vehicle_shared USING btree (fk_vehicle_id);


--
-- TOC entry 5308 (class 1259 OID 32078)
-- Name: vehicle_vessel_anchor_handling_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX vehicle_vessel_anchor_handling_fk_idx ON reference.vehicle_vessel_anchor_handling USING btree (fk_vehicle_id);


--
-- TOC entry 5309 (class 1259 OID 32079)
-- Name: vehicle_vessel_anchor_handling_fk_type_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX vehicle_vessel_anchor_handling_fk_type_idx ON reference.vehicle_vessel_anchor_handling USING btree (fk_vehicle_type_id);


--
-- TOC entry 5312 (class 1259 OID 32080)
-- Name: vehicle_vessel_cable_laying_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX vehicle_vessel_cable_laying_fk_idx ON reference.vehicle_vessel_cable_laying USING btree (fk_vehicle_id);


--
-- TOC entry 5313 (class 1259 OID 32081)
-- Name: vehicle_vessel_cable_laying_fk_type_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX vehicle_vessel_cable_laying_fk_type_idx ON reference.vehicle_vessel_cable_laying USING btree (fk_vehicle_type_id);


--
-- TOC entry 5316 (class 1259 OID 32082)
-- Name: vehicle_vessel_cargo_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX vehicle_vessel_cargo_fk_idx ON reference.vehicle_vessel_cargo USING btree (fk_vehicle_id);


--
-- TOC entry 5317 (class 1259 OID 32083)
-- Name: vehicle_vessel_cargo_fk_type_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX vehicle_vessel_cargo_fk_type_idx ON reference.vehicle_vessel_cargo USING btree (fk_vehicle_type_id);


--
-- TOC entry 5320 (class 1259 OID 32084)
-- Name: vehicle_vessel_jackup_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX vehicle_vessel_jackup_fk_idx ON reference.vehicle_vessel_jackup USING btree (fk_vehicle_id);


--
-- TOC entry 5321 (class 1259 OID 32085)
-- Name: vehicle_vessel_jackup_fk_type_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX vehicle_vessel_jackup_fk_type_idx ON reference.vehicle_vessel_jackup USING btree (fk_vehicle_type_id);


--
-- TOC entry 5324 (class 1259 OID 32086)
-- Name: vehicle_vessel_tugboat_fk_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX vehicle_vessel_tugboat_fk_idx ON reference.vehicle_vessel_tugboat USING btree (fk_vehicle_id);


--
-- TOC entry 5325 (class 1259 OID 32087)
-- Name: vehicle_vessel_tugboat_fk_type_idx; Type: INDEX; Schema: reference; Owner: -
--

CREATE INDEX vehicle_vessel_tugboat_fk_type_idx ON reference.vehicle_vessel_tugboat USING btree (fk_vehicle_type_id);


--
-- TOC entry 5329 (class 2606 OID 32088)
-- Name: bathymetry bathymetry_fk_site_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.bathymetry
    ADD CONSTRAINT bathymetry_fk_site_id_fkey FOREIGN KEY (fk_site_id) REFERENCES project.site(id);


--
-- TOC entry 5330 (class 2606 OID 32093)
-- Name: bathymetry_layer bathymetry_layer_fk_bathymetry_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.bathymetry_layer
    ADD CONSTRAINT bathymetry_layer_fk_bathymetry_id_fkey FOREIGN KEY (fk_bathymetry_id) REFERENCES project.bathymetry(id);


--
-- TOC entry 5331 (class 2606 OID 32098)
-- Name: bathymetry_layer bathymetry_layer_fk_soil_type_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.bathymetry_layer
    ADD CONSTRAINT bathymetry_layer_fk_soil_type_id_fkey FOREIGN KEY (fk_soil_type_id) REFERENCES reference.soil_type(id);


--
-- TOC entry 5332 (class 2606 OID 32103)
-- Name: cable_corridor_bathymetry cable_corridor_bathymetry_fk_site_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.cable_corridor_bathymetry
    ADD CONSTRAINT cable_corridor_bathymetry_fk_site_id_fkey FOREIGN KEY (fk_site_id) REFERENCES project.site(id);


--
-- TOC entry 5333 (class 2606 OID 32108)
-- Name: cable_corridor_bathymetry_layer cable_corridor_bathymetry_layer_fk_bathymetry_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.cable_corridor_bathymetry_layer
    ADD CONSTRAINT cable_corridor_bathymetry_layer_fk_bathymetry_id_fkey FOREIGN KEY (fk_bathymetry_id) REFERENCES project.cable_corridor_bathymetry(id);


--
-- TOC entry 5334 (class 2606 OID 32113)
-- Name: cable_corridor_bathymetry_layer cable_corridor_bathymetry_layer_fk_soil_type_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.cable_corridor_bathymetry_layer
    ADD CONSTRAINT cable_corridor_bathymetry_layer_fk_soil_type_id_fkey FOREIGN KEY (fk_soil_type_id) REFERENCES reference.soil_type(id);


--
-- TOC entry 5335 (class 2606 OID 32118)
-- Name: cable_corridor_constraint cable_corridor_constraint_fk_site_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.cable_corridor_constraint
    ADD CONSTRAINT cable_corridor_constraint_fk_site_id_fkey FOREIGN KEY (fk_site_id) REFERENCES project.site(id);


--
-- TOC entry 5336 (class 2606 OID 32123)
-- Name: constraint constraint_fk_site_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project."constraint"
    ADD CONSTRAINT constraint_fk_site_id_fkey FOREIGN KEY (fk_site_id) REFERENCES project.site(id);


--
-- TOC entry 5337 (class 2606 OID 32128)
-- Name: device_floating device_floating_fk_device_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.device_floating
    ADD CONSTRAINT device_floating_fk_device_id_fkey FOREIGN KEY (fk_device_id) REFERENCES project.device(id);


--
-- TOC entry 5338 (class 2606 OID 32133)
-- Name: device_shared device_shared_fk_device_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.device_shared
    ADD CONSTRAINT device_shared_fk_device_id_fkey FOREIGN KEY (fk_device_id) REFERENCES project.device(id);


--
-- TOC entry 5339 (class 2606 OID 32138)
-- Name: device_tidal device_tidal_fk_device_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.device_tidal
    ADD CONSTRAINT device_tidal_fk_device_id_fkey FOREIGN KEY (fk_device_id) REFERENCES project.device(id);


--
-- TOC entry 5340 (class 2606 OID 32143)
-- Name: device_tidal_power_performance device_tidal_power_performance_fk_device_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.device_tidal_power_performance
    ADD CONSTRAINT device_tidal_power_performance_fk_device_id_fkey FOREIGN KEY (fk_device_id) REFERENCES project.device(id);


--
-- TOC entry 5341 (class 2606 OID 32148)
-- Name: device_wave device_wave_fk_device_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.device_wave
    ADD CONSTRAINT device_wave_fk_device_id_fkey FOREIGN KEY (fk_device_id) REFERENCES project.device(id);


--
-- TOC entry 5342 (class 2606 OID 32153)
-- Name: lease_area lease_area_fk_site_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.lease_area
    ADD CONSTRAINT lease_area_fk_site_id_fkey FOREIGN KEY (fk_site_id) REFERENCES project.site(id);


--
-- TOC entry 5343 (class 2606 OID 32158)
-- Name: sub_systems_access sub_systems_access_fk_sub_system_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_access
    ADD CONSTRAINT sub_systems_access_fk_sub_system_id_fkey FOREIGN KEY (fk_sub_system_id) REFERENCES project.sub_systems(id);


--
-- TOC entry 5344 (class 2606 OID 32163)
-- Name: sub_systems_economic sub_systems_economic_fk_sub_system_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_economic
    ADD CONSTRAINT sub_systems_economic_fk_sub_system_id_fkey FOREIGN KEY (fk_sub_system_id) REFERENCES project.sub_systems(id);


--
-- TOC entry 5328 (class 2606 OID 32168)
-- Name: sub_systems sub_systems_fk_device_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems
    ADD CONSTRAINT sub_systems_fk_device_id_fkey FOREIGN KEY (fk_device_id) REFERENCES project.device(id);


--
-- TOC entry 5345 (class 2606 OID 32173)
-- Name: sub_systems_inspection sub_systems_inspection_fk_sub_system_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_inspection
    ADD CONSTRAINT sub_systems_inspection_fk_sub_system_id_fkey FOREIGN KEY (fk_sub_system_id) REFERENCES project.sub_systems(id);


--
-- TOC entry 5346 (class 2606 OID 32178)
-- Name: sub_systems_install sub_systems_install_fk_sub_system_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_install
    ADD CONSTRAINT sub_systems_install_fk_sub_system_id_fkey FOREIGN KEY (fk_sub_system_id) REFERENCES project.sub_systems(id);


--
-- TOC entry 5347 (class 2606 OID 32183)
-- Name: sub_systems_maintenance sub_systems_maintenance_fk_sub_system_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_maintenance
    ADD CONSTRAINT sub_systems_maintenance_fk_sub_system_id_fkey FOREIGN KEY (fk_sub_system_id) REFERENCES project.sub_systems(id);


--
-- TOC entry 5348 (class 2606 OID 32188)
-- Name: sub_systems_operation_weightings sub_systems_operation_weightings_fk_sub_system_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_operation_weightings
    ADD CONSTRAINT sub_systems_operation_weightings_fk_sub_system_id_fkey FOREIGN KEY (fk_sub_system_id) REFERENCES project.sub_systems(id);


--
-- TOC entry 5349 (class 2606 OID 32193)
-- Name: sub_systems_replace sub_systems_replace_fk_sub_system_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.sub_systems_replace
    ADD CONSTRAINT sub_systems_replace_fk_sub_system_id_fkey FOREIGN KEY (fk_sub_system_id) REFERENCES project.sub_systems(id);


--
-- TOC entry 5350 (class 2606 OID 32198)
-- Name: time_series_energy_tidal time_series_energy_tidal_fk_bathymetry_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.time_series_energy_tidal
    ADD CONSTRAINT time_series_energy_tidal_fk_bathymetry_id_fkey FOREIGN KEY (fk_bathymetry_id) REFERENCES project.bathymetry(id);


--
-- TOC entry 5351 (class 2606 OID 32203)
-- Name: time_series_energy_wave time_series_energy_wave_fk_site_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.time_series_energy_wave
    ADD CONSTRAINT time_series_energy_wave_fk_site_id_fkey FOREIGN KEY (fk_site_id) REFERENCES project.site(id);


--
-- TOC entry 5352 (class 2606 OID 32208)
-- Name: time_series_om_tidal time_series_om_tidal_fk_site_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.time_series_om_tidal
    ADD CONSTRAINT time_series_om_tidal_fk_site_id_fkey FOREIGN KEY (fk_site_id) REFERENCES project.site(id);


--
-- TOC entry 5353 (class 2606 OID 32213)
-- Name: time_series_om_wave time_series_om_wave_fk_site_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.time_series_om_wave
    ADD CONSTRAINT time_series_om_wave_fk_site_id_fkey FOREIGN KEY (fk_site_id) REFERENCES project.site(id);


--
-- TOC entry 5354 (class 2606 OID 32218)
-- Name: time_series_om_wind time_series_om_wind_fk_site_id_fkey; Type: FK CONSTRAINT; Schema: project; Owner: -
--

ALTER TABLE ONLY project.time_series_om_wind
    ADD CONSTRAINT time_series_om_wind_fk_site_id_fkey FOREIGN KEY (fk_site_id) REFERENCES project.site(id);


--
-- TOC entry 5355 (class 2606 OID 32223)
-- Name: component_anchor component_anchor_fk_component_discrete_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_anchor
    ADD CONSTRAINT component_anchor_fk_component_discrete_id_fkey FOREIGN KEY (fk_component_discrete_id) REFERENCES reference.component_discrete(id);


--
-- TOC entry 5356 (class 2606 OID 32228)
-- Name: component_anchor component_anchor_fk_component_type_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_anchor
    ADD CONSTRAINT component_anchor_fk_component_type_id_fkey FOREIGN KEY (fk_component_type_id) REFERENCES reference.component_type(id);


--
-- TOC entry 5357 (class 2606 OID 32233)
-- Name: component_cable component_cable_fk_component_continuous_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_cable
    ADD CONSTRAINT component_cable_fk_component_continuous_id_fkey FOREIGN KEY (fk_component_continuous_id) REFERENCES reference.component_continuous(id);


--
-- TOC entry 5358 (class 2606 OID 32238)
-- Name: component_cable component_cable_fk_component_type_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_cable
    ADD CONSTRAINT component_cable_fk_component_type_id_fkey FOREIGN KEY (fk_component_type_id) REFERENCES reference.component_type(id);


--
-- TOC entry 5359 (class 2606 OID 32243)
-- Name: component_collection_point component_collection_point_fk_component_discrete_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_collection_point
    ADD CONSTRAINT component_collection_point_fk_component_discrete_id_fkey FOREIGN KEY (fk_component_discrete_id) REFERENCES reference.component_discrete(id);


--
-- TOC entry 5360 (class 2606 OID 32248)
-- Name: component_collection_point component_collection_point_fk_component_type_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_collection_point
    ADD CONSTRAINT component_collection_point_fk_component_type_id_fkey FOREIGN KEY (fk_component_type_id) REFERENCES reference.component_type(id);


--
-- TOC entry 5361 (class 2606 OID 32253)
-- Name: component_connector component_connector_fk_component_discrete_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_connector
    ADD CONSTRAINT component_connector_fk_component_discrete_fkey FOREIGN KEY (fk_component_discrete_id) REFERENCES reference.component_discrete(id);


--
-- TOC entry 5362 (class 2606 OID 32258)
-- Name: component_connector component_connector_fk_component_type_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_connector
    ADD CONSTRAINT component_connector_fk_component_type_id_fkey FOREIGN KEY (fk_component_type_id) REFERENCES reference.component_type(id);


--
-- TOC entry 5363 (class 2606 OID 32263)
-- Name: component_continuous component_continuous_fk_component_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_continuous
    ADD CONSTRAINT component_continuous_fk_component_id_fkey FOREIGN KEY (fk_component_id) REFERENCES reference.component(id);


--
-- TOC entry 5364 (class 2606 OID 32268)
-- Name: component_discrete component_discrete_fk_component_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_discrete
    ADD CONSTRAINT component_discrete_fk_component_id_fkey FOREIGN KEY (fk_component_id) REFERENCES reference.component(id);


--
-- TOC entry 5365 (class 2606 OID 32273)
-- Name: component_mooring_continuous component_mooring_continuous_fk_component_continuous_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_mooring_continuous
    ADD CONSTRAINT component_mooring_continuous_fk_component_continuous_id_fkey FOREIGN KEY (fk_component_continuous_id) REFERENCES reference.component_continuous(id);


--
-- TOC entry 5366 (class 2606 OID 32278)
-- Name: component_mooring_continuous component_mooring_continuous_fk_component_type_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_mooring_continuous
    ADD CONSTRAINT component_mooring_continuous_fk_component_type_id_fkey FOREIGN KEY (fk_component_type_id) REFERENCES reference.component_type(id);


--
-- TOC entry 5367 (class 2606 OID 32283)
-- Name: component_mooring_discrete component_mooring_discrete_fk_component_discrete_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_mooring_discrete
    ADD CONSTRAINT component_mooring_discrete_fk_component_discrete_id_fkey FOREIGN KEY (fk_component_discrete_id) REFERENCES reference.component_discrete(id);


--
-- TOC entry 5368 (class 2606 OID 32288)
-- Name: component_mooring_discrete component_mooring_discrete_fk_component_type_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_mooring_discrete
    ADD CONSTRAINT component_mooring_discrete_fk_component_type_id_fkey FOREIGN KEY (fk_component_type_id) REFERENCES reference.component_type(id);


--
-- TOC entry 5369 (class 2606 OID 32293)
-- Name: component_pile component_pile_fk_component_continuous_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_pile
    ADD CONSTRAINT component_pile_fk_component_continuous_id_fkey FOREIGN KEY (fk_component_continuous_id) REFERENCES reference.component_continuous(id);


--
-- TOC entry 5370 (class 2606 OID 32298)
-- Name: component_pile component_pile_fk_component_type_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_pile
    ADD CONSTRAINT component_pile_fk_component_type_id_fkey FOREIGN KEY (fk_component_type_id) REFERENCES reference.component_type(id);


--
-- TOC entry 5371 (class 2606 OID 32303)
-- Name: component_rope component_rope_fk_component_continuous_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_rope
    ADD CONSTRAINT component_rope_fk_component_continuous_id_fkey FOREIGN KEY (fk_component_continuous_id) REFERENCES reference.component_continuous(id);


--
-- TOC entry 5372 (class 2606 OID 32308)
-- Name: component_rope component_rope_fk_component_type_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_rope
    ADD CONSTRAINT component_rope_fk_component_type_id_fkey FOREIGN KEY (fk_component_type_id) REFERENCES reference.component_type(id);


--
-- TOC entry 5373 (class 2606 OID 32313)
-- Name: component_shared component_shared_fk_component_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_shared
    ADD CONSTRAINT component_shared_fk_component_id_fkey FOREIGN KEY (fk_component_id) REFERENCES reference.component(id);


--
-- TOC entry 5374 (class 2606 OID 32318)
-- Name: component_transformer component_transformer_fk_component_discrete_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_transformer
    ADD CONSTRAINT component_transformer_fk_component_discrete_id_fkey FOREIGN KEY (fk_component_discrete_id) REFERENCES reference.component_discrete(id);


--
-- TOC entry 5375 (class 2606 OID 32323)
-- Name: component_transformer component_transformer_fk_component_type_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.component_transformer
    ADD CONSTRAINT component_transformer_fk_component_type_id_fkey FOREIGN KEY (fk_component_type_id) REFERENCES reference.component_type(id);


--
-- TOC entry 5376 (class 2606 OID 32328)
-- Name: operations_limit_cs operations_limit_cs_fk_operations_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.operations_limit_cs
    ADD CONSTRAINT operations_limit_cs_fk_operations_id_fkey FOREIGN KEY (fk_operations_id) REFERENCES reference.operations_type(id);


--
-- TOC entry 5377 (class 2606 OID 32333)
-- Name: operations_limit_hs operations_limit_hs_fk_operations_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.operations_limit_hs
    ADD CONSTRAINT operations_limit_hs_fk_operations_id_fkey FOREIGN KEY (fk_operations_id) REFERENCES reference.operations_type(id);


--
-- TOC entry 5378 (class 2606 OID 32338)
-- Name: operations_limit_tp operations_limit_tp_fk_operations_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.operations_limit_tp
    ADD CONSTRAINT operations_limit_tp_fk_operations_id_fkey FOREIGN KEY (fk_operations_id) REFERENCES reference.operations_type(id);


--
-- TOC entry 5379 (class 2606 OID 32343)
-- Name: operations_limit_ws operations_limit_ws_fk_operations_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.operations_limit_ws
    ADD CONSTRAINT operations_limit_ws_fk_operations_id_fkey FOREIGN KEY (fk_operations_id) REFERENCES reference.operations_type(id);


--
-- TOC entry 5380 (class 2606 OID 32348)
-- Name: soil_type_geotechnical_properties soil_type_geotechnical_properties_fk_soil_type_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.soil_type_geotechnical_properties
    ADD CONSTRAINT soil_type_geotechnical_properties_fk_soil_type_id_fkey FOREIGN KEY (fk_soil_type_id) REFERENCES reference.soil_type(id);


--
-- TOC entry 5381 (class 2606 OID 32353)
-- Name: vehicle_helicopter vehicle_helicopter_fk_vehicle_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_helicopter
    ADD CONSTRAINT vehicle_helicopter_fk_vehicle_id_fkey FOREIGN KEY (fk_vehicle_id) REFERENCES reference.vehicle(id);


--
-- TOC entry 5382 (class 2606 OID 32358)
-- Name: vehicle_helicopter vehicle_helicopter_fk_vehicle_type_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_helicopter
    ADD CONSTRAINT vehicle_helicopter_fk_vehicle_type_id_fkey FOREIGN KEY (fk_vehicle_type_id) REFERENCES reference.vehicle_type(id);


--
-- TOC entry 5383 (class 2606 OID 32363)
-- Name: vehicle_shared vehicle_shared_fk_vehicle_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_shared
    ADD CONSTRAINT vehicle_shared_fk_vehicle_id_fkey FOREIGN KEY (fk_vehicle_id) REFERENCES reference.vehicle(id);


--
-- TOC entry 5384 (class 2606 OID 32368)
-- Name: vehicle_vessel_anchor_handling vehicle_vessel_anchor_handling_fk_vehicle_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_anchor_handling
    ADD CONSTRAINT vehicle_vessel_anchor_handling_fk_vehicle_id_fkey FOREIGN KEY (fk_vehicle_id) REFERENCES reference.vehicle(id);


--
-- TOC entry 5385 (class 2606 OID 32373)
-- Name: vehicle_vessel_anchor_handling vehicle_vessel_anchor_handling_fk_vehicle_type_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_anchor_handling
    ADD CONSTRAINT vehicle_vessel_anchor_handling_fk_vehicle_type_id_fkey FOREIGN KEY (fk_vehicle_type_id) REFERENCES reference.vehicle_type(id);


--
-- TOC entry 5386 (class 2606 OID 32378)
-- Name: vehicle_vessel_cable_laying vehicle_vessel_cable_laying_fk_vehicle_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_cable_laying
    ADD CONSTRAINT vehicle_vessel_cable_laying_fk_vehicle_id_fkey FOREIGN KEY (fk_vehicle_id) REFERENCES reference.vehicle(id);


--
-- TOC entry 5387 (class 2606 OID 32383)
-- Name: vehicle_vessel_cable_laying vehicle_vessel_cable_laying_fk_vehicle_type_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_cable_laying
    ADD CONSTRAINT vehicle_vessel_cable_laying_fk_vehicle_type_id_fkey FOREIGN KEY (fk_vehicle_type_id) REFERENCES reference.vehicle_type(id);


--
-- TOC entry 5388 (class 2606 OID 32388)
-- Name: vehicle_vessel_cargo vehicle_vessel_cargo_fk_vehicle_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_cargo
    ADD CONSTRAINT vehicle_vessel_cargo_fk_vehicle_id_fkey FOREIGN KEY (fk_vehicle_id) REFERENCES reference.vehicle(id);


--
-- TOC entry 5389 (class 2606 OID 32393)
-- Name: vehicle_vessel_cargo vehicle_vessel_cargo_fk_vehicle_type_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_cargo
    ADD CONSTRAINT vehicle_vessel_cargo_fk_vehicle_type_id_fkey FOREIGN KEY (fk_vehicle_type_id) REFERENCES reference.vehicle_type(id);


--
-- TOC entry 5390 (class 2606 OID 32398)
-- Name: vehicle_vessel_jackup vehicle_vessel_jackup_fk_vehicle_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_jackup
    ADD CONSTRAINT vehicle_vessel_jackup_fk_vehicle_id_fkey FOREIGN KEY (fk_vehicle_id) REFERENCES reference.vehicle(id);


--
-- TOC entry 5391 (class 2606 OID 32403)
-- Name: vehicle_vessel_jackup vehicle_vessel_jackup_fk_vehicle_type_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_jackup
    ADD CONSTRAINT vehicle_vessel_jackup_fk_vehicle_type_id_fkey FOREIGN KEY (fk_vehicle_type_id) REFERENCES reference.vehicle_type(id);


--
-- TOC entry 5392 (class 2606 OID 32408)
-- Name: vehicle_vessel_tugboat vehicle_vessel_tugboat_fk_vehicle_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_tugboat
    ADD CONSTRAINT vehicle_vessel_tugboat_fk_vehicle_id_fkey FOREIGN KEY (fk_vehicle_id) REFERENCES reference.vehicle(id);


--
-- TOC entry 5393 (class 2606 OID 32413)
-- Name: vehicle_vessel_tugboat vehicle_vessel_tugboat_fk_vehicle_type_id_fkey; Type: FK CONSTRAINT; Schema: reference; Owner: -
--

ALTER TABLE ONLY reference.vehicle_vessel_tugboat
    ADD CONSTRAINT vehicle_vessel_tugboat_fk_vehicle_type_id_fkey FOREIGN KEY (fk_vehicle_type_id) REFERENCES reference.vehicle_type(id);


-- Completed on 2019-03-12 11:14:07

--
-- PostgreSQL database dump complete
--

