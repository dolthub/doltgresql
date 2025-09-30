-- Downloaded from: https://github.com/Abhishek842000/DB-Performance-Comparator/blob/953e7e2352e6cf7c8d7d0ffc4a0ee63b2f06cef5/full.sql
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
-- Name: nflexon_app; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA nflexon_app;


ALTER SCHEMA nflexon_app OWNER TO postgres;

--
-- Name: ambienttempdetail_byprojectid(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ambienttempdetail_byprojectid(p_panel_project_id integer) RETURNS TABLE(tempdetailid integer, panelprojectid integer, threshold_value numeric, readingintervalinmin integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT * FROM AmbientTempDetail WHERE PanelProjectId = p_panel_project_id;
END;
$$;


ALTER FUNCTION public.ambienttempdetail_byprojectid(p_panel_project_id integer) OWNER TO postgres;

--
-- Name: ambienttempdetail_insertupdate(integer, numeric, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ambienttempdetail_insertupdate(p_panel_project_id integer, p_threshold_value numeric, p_reading_interval_in_min integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_rtn_value INTEGER;
BEGIN
    IF NOT EXISTS (SELECT * FROM AmbientTempDetail WHERE PanelProjectId = p_panel_project_id) THEN
        INSERT INTO AmbientTempDetail (PanelProjectId, Threshold_Value, ReadingIntervalInMin)
        VALUES (p_panel_project_id, p_threshold_value, p_reading_interval_in_min)
        RETURNING TempDetailId INTO v_rtn_value;
    ELSE
        UPDATE AmbientTempDetail
        SET PanelProjectId = p_panel_project_id,
            Threshold_Value = p_threshold_value,
            ReadingIntervalInMin = p_reading_interval_in_min
        WHERE PanelProjectId = p_panel_project_id;
        v_rtn_value := p_panel_project_id;
    END IF;
    
    RETURN v_rtn_value;
END;
$$;


ALTER FUNCTION public.ambienttempdetail_insertupdate(p_panel_project_id integer, p_threshold_value numeric, p_reading_interval_in_min integer) OWNER TO postgres;

--
-- Name: ambienttempsensors_byprojectid(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ambienttempsensors_byprojectid(p_panel_project_id integer) RETURNS TABLE(panelprojectid integer, ip_address character varying, notes character varying)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT * FROM AmbientTempSensors WHERE PanelProjectId = p_panel_project_id;
END;
$$;


ALTER FUNCTION public.ambienttempsensors_byprojectid(p_panel_project_id integer) OWNER TO postgres;

--
-- Name: ambienttempsensors_delete(integer, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ambienttempsensors_delete(p_panel_project_id integer, p_ip_address character varying) RETURNS integer
    LANGUAGE plpgsql
    AS $$
BEGIN
    DELETE FROM AmbientTempSensors 
    WHERE PanelProjectId = p_panel_project_id AND IP_Address = p_ip_address;
    RETURN 1;
END;
$$;


ALTER FUNCTION public.ambienttempsensors_delete(p_panel_project_id integer, p_ip_address character varying) OWNER TO postgres;

--
-- Name: ambienttempsensors_insertupdate(integer, character varying, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ambienttempsensors_insertupdate(p_panel_project_id integer, p_ip_address character varying, p_notes character varying) RETURNS integer
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NOT EXISTS(SELECT * FROM AmbientTempSensors WHERE PanelProjectId = p_panel_project_id AND IP_Address = p_ip_address) THEN
        INSERT INTO AmbientTempSensors(PanelProjectId, IP_Address, Notes)
        VALUES (p_panel_project_id, p_ip_address, p_notes);
    ELSE
        UPDATE AmbientTempSensors
        SET Notes = p_notes
        WHERE PanelProjectId = p_panel_project_id AND IP_Address = p_ip_address;
    END IF;
    
    RETURN 1;
END;
$$;


ALTER FUNCTION public.ambienttempsensors_insertupdate(p_panel_project_id integer, p_ip_address character varying, p_notes character varying) OWNER TO postgres;

--
-- Name: bundledetail_deletebybundleid(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.bundledetail_deletebybundleid(p_bundle_id integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
BEGIN
    DELETE FROM BundleDetail WHERE BundleId = p_bundle_id;
    RETURN 1;
END;
$$;


ALTER FUNCTION public.bundledetail_deletebybundleid(p_bundle_id integer) OWNER TO postgres;

--
-- Name: bundledetail_getbybundleid(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.bundledetail_getbybundleid(p_bundle_id integer) RETURNS TABLE(detailid integer, bundleid integer, panelno integer, portno integer, actiongroupid integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT * FROM BundleDetail WHERE BundleId = p_bundle_id;
END;
$$;


ALTER FUNCTION public.bundledetail_getbybundleid(p_bundle_id integer) OWNER TO postgres;

--
-- Name: bundledetail_insertupdate(integer, integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.bundledetail_insertupdate(p_detail_id integer, p_bundle_id integer, p_panel_no integer, p_port_no integer, p_action_group_id integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_rtn_value INTEGER;
BEGIN
    IF p_detail_id = 0 THEN
        INSERT INTO BundleDetail (BundleId, PanelNo, PortNo, ActionGroupId)
        VALUES (p_bundle_id, p_panel_no, p_port_no, p_action_group_id)
        RETURNING DetailId INTO v_rtn_value;
    ELSE
        UPDATE BundleDetail
        SET BundleId = p_bundle_id,
            PanelNo = p_panel_no,
            PortNo = p_port_no,
            ActionGroupId = p_action_group_id
        WHERE DetailId = p_detail_id;
        v_rtn_value := p_detail_id;
    END IF;
    
    RETURN v_rtn_value;
END;
$$;


ALTER FUNCTION public.bundledetail_insertupdate(p_detail_id integer, p_bundle_id integer, p_panel_no integer, p_port_no integer, p_action_group_id integer) OWNER TO postgres;

--
-- Name: bundlemaster_getbyid(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.bundlemaster_getbyid(p_bundle_id integer) RETURNS TABLE(bundleid integer, no_of_cables integer, temppanelno integer, threshold_value numeric, panelprojectid integer, bundlename character varying)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT * FROM BundleMaster WHERE BundleId = p_bundle_id;
END;
$$;


ALTER FUNCTION public.bundlemaster_getbyid(p_bundle_id integer) OWNER TO postgres;

--
-- Name: bundlemaster_getbyprojectid(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.bundlemaster_getbyprojectid(p_panel_project_id integer) RETURNS TABLE(bundleid integer, no_of_cables integer, temppanelno integer, threshold_value numeric, panelprojectid integer, bundlename character varying)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT * FROM BundleMaster WHERE PanelProjectId = p_panel_project_id;
END;
$$;


ALTER FUNCTION public.bundlemaster_getbyprojectid(p_panel_project_id integer) OWNER TO postgres;

--
-- Name: bundlemaster_insertupdate(integer, integer, integer, numeric, integer, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.bundlemaster_insertupdate(p_bundle_id integer, p_no_of_cables integer, p_temp_panel_no integer, p_threshold_value numeric, p_panel_project_id integer, p_bundle_name character varying) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_rtn_value INTEGER;
BEGIN
    IF p_bundle_id = 0 THEN
        INSERT INTO BundleMaster (No_Of_Cables, TempPanelNo, Threshold_Value, PanelProjectId, BundleName)
        VALUES (p_no_of_cables, p_temp_panel_no, p_threshold_value, p_panel_project_id, p_bundle_name)
        RETURNING BundleId INTO v_rtn_value;
    ELSE
        UPDATE BundleMaster
        SET No_Of_Cables = p_no_of_cables,
            TempPanelNo = p_temp_panel_no,
            Threshold_Value = p_threshold_value,
            PanelProjectId = p_panel_project_id,
            BundleName = p_bundle_name
        WHERE BundleId = p_bundle_id;
        v_rtn_value := p_bundle_id;
    END IF;
    
    RETURN v_rtn_value;
END;
$$;


ALTER FUNCTION public.bundlemaster_insertupdate(p_bundle_id integer, p_no_of_cables integer, p_temp_panel_no integer, p_threshold_value numeric, p_panel_project_id integer, p_bundle_name character varying) OWNER TO postgres;

--
-- Name: configuration_getbyprojectid(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.configuration_getbyprojectid(p_panel_project_id integer) RETURNS TABLE(id integer, smtp character varying, port integer, userid character varying, password character varying, tempunit integer, panelprojectid integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT * FROM Configuration WHERE PanelProjectId = p_panel_project_id;
END;
$$;


ALTER FUNCTION public.configuration_getbyprojectid(p_panel_project_id integer) OWNER TO postgres;

--
-- Name: configuration_insertupdate(character varying, integer, character varying, character varying, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.configuration_insertupdate(p_smtp character varying, p_port integer, p_user_id character varying, p_password character varying, p_temp_unit integer, p_panel_project_id integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_rtn_value INTEGER;
BEGIN
    IF NOT EXISTS(SELECT * FROM Configuration WHERE PanelProjectId = p_panel_project_id) THEN
        INSERT INTO Configuration (SMTP, Port, UserId, Password, TempUnit, PanelProjectId)
        VALUES (p_smtp, p_port, p_user_id, p_password, p_temp_unit, p_panel_project_id)
        RETURNING id INTO v_rtn_value;
    ELSE
        UPDATE Configuration
        SET SMTP = p_smtp,
            Port = p_port,
            UserId = p_user_id,
            Password = p_password,
            TempUnit = p_temp_unit
        WHERE PanelProjectId = p_panel_project_id;
        v_rtn_value := p_panel_project_id;
    END IF;
    
    RETURN v_rtn_value;
END;
$$;


ALTER FUNCTION public.configuration_insertupdate(p_smtp character varying, p_port integer, p_user_id character varying, p_password character varying, p_temp_unit integer, p_panel_project_id integer) OWNER TO postgres;

--
-- Name: history_insert(integer, integer, integer, integer, numeric, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.history_insert(p_panel_project_id integer, p_panel_slave_address integer, p_port_no integer, p_cable_no integer, p_ambient_temp numeric, p_port_status integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
BEGIN
    INSERT INTO History (
        PanelProjectId,
        PanelSlaveAddress,
        PortNo,
        CableNo,
        AmbientTemp,
        PortStatus,
        CreatedOn
    )
    VALUES (
        p_panel_project_id,
        p_panel_slave_address,
        p_port_no,
        p_cable_no,
        p_ambient_temp,
        p_port_status,
        CURRENT_TIMESTAMP
    );
    RETURN 1;
END;
$$;


ALTER FUNCTION public.history_insert(p_panel_project_id integer, p_panel_slave_address integer, p_port_no integer, p_cable_no integer, p_ambient_temp numeric, p_port_status integer) OWNER TO postgres;

--
-- Name: iotpanelproject_insertupdate(integer, character varying, integer, character varying, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.iotpanelproject_insertupdate(p_id integer, p_ip_address character varying, p_no_of_panels integer, p_unique_identification character varying, p_port_no integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_rtn_value INTEGER;
BEGIN
    IF p_id = 0 THEN
        INSERT INTO IoTPanelProject (IP_Address, No_Of_Panels, UniqueIdentification, PortNo)
        VALUES (p_ip_address, p_no_of_panels, p_unique_identification, p_port_no)
        RETURNING id INTO v_rtn_value;
    ELSE
        UPDATE IoTPanelProject
        SET IP_Address = p_ip_address,
            No_Of_Panels = p_no_of_panels,
            UniqueIdentification = p_unique_identification,
            PortNo = p_port_no
        WHERE Id = p_id;
        v_rtn_value := p_id;
    END IF;
    
    RETURN v_rtn_value;
END;
$$;


ALTER FUNCTION public.iotpanelproject_insertupdate(p_id integer, p_ip_address character varying, p_no_of_panels integer, p_unique_identification character varying, p_port_no integer) OWNER TO postgres;

--
-- Name: iotpanelproject_selectall(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.iotpanelproject_selectall() RETURNS TABLE(id integer, ip_address character varying, no_of_panels integer, uniqueidentification character varying, portno integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT * FROM IoTPanelProject;
END;
$$;


ALTER FUNCTION public.iotpanelproject_selectall() OWNER TO postgres;

--
-- Name: iotpanelproject_selectbyid(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.iotpanelproject_selectbyid(p_id integer) RETURNS TABLE(id integer, ip_address character varying, no_of_panels integer, uniqueidentification character varying, portno integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT * FROM IoTPanelProject WHERE Id = p_id;
END;
$$;


ALTER FUNCTION public.iotpanelproject_selectbyid(p_id integer) OWNER TO postgres;

--
-- Name: iotpanelthreshold_get(integer, integer, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.iotpanelthreshold_get(p_panel_project_id integer, p_panel_no integer, p_port_no integer, p_cable_pin integer) RETURNS TABLE(id integer, panelprojectid integer, panelno integer, portno integer, cablepin integer, threshold_value numeric)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT * FROM IoTPanelThreshold
    WHERE PanelProjectId = p_panel_project_id 
    AND PanelNo = p_panel_no 
    AND PortNo = p_port_no 
    AND CablePin = p_cable_pin;
END;
$$;


ALTER FUNCTION public.iotpanelthreshold_get(p_panel_project_id integer, p_panel_no integer, p_port_no integer, p_cable_pin integer) OWNER TO postgres;

--
-- Name: iotpanelthreshold_getbypanelno(integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.iotpanelthreshold_getbypanelno(p_panel_project_id integer, p_panel_no integer) RETURNS TABLE(id integer, panelprojectid integer, panelno integer, portno integer, cablepin integer, threshold_value numeric)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT * FROM IoTPanelThreshold
    WHERE PanelProjectId = p_panel_project_id 
    AND PanelNo = p_panel_no;
END;
$$;


ALTER FUNCTION public.iotpanelthreshold_getbypanelno(p_panel_project_id integer, p_panel_no integer) OWNER TO postgres;

--
-- Name: iotpanelthreshold_insertupdate(integer, integer, integer, integer, integer, numeric); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.iotpanelthreshold_insertupdate(p_id integer, p_panel_project_id integer, p_panel_no integer, p_port_no integer, p_cable_pin integer, p_threshold_value numeric) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_rtn_value INTEGER;
BEGIN
    IF p_id = 0 THEN
        INSERT INTO IoTPanelThreshold (PanelProjectId, PanelNo, PortNo, CablePin, Threshold_Value)
        VALUES (p_panel_project_id, p_panel_no, p_port_no, p_cable_pin, p_threshold_value)
        RETURNING id INTO v_rtn_value;
    ELSE
        UPDATE IoTPanelThreshold
        SET PanelProjectId = p_panel_project_id,
            PanelNo = p_panel_no,
            PortNo = p_port_no,
            CablePin = p_cable_pin,
            Threshold_Value = p_threshold_value
        WHERE Id = p_id;
        v_rtn_value := p_id;
    END IF;
    
    RETURN v_rtn_value;
END;
$$;


ALTER FUNCTION public.iotpanelthreshold_insertupdate(p_id integer, p_panel_project_id integer, p_panel_no integer, p_port_no integer, p_cable_pin integer, p_threshold_value numeric) OWNER TO postgres;

--
-- Name: iotpanelthreshold_save(integer, integer, integer, integer, integer, numeric); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.iotpanelthreshold_save(p_id integer, p_panel_project_id integer, p_panel_no integer, p_port_no integer, p_cable_pin integer, p_threshold_value numeric) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_rtn_value INTEGER;
BEGIN
    IF NOT EXISTS (SELECT * FROM IoTPanelThreshold
        WHERE PanelProjectId = p_panel_project_id 
        AND PanelNo = p_panel_no 
        AND PortNo = p_port_no 
        AND CablePin = p_cable_pin) THEN
        INSERT INTO IoTPanelThreshold (PanelProjectId, PanelNo, PortNo, CablePin, Threshold_Value)
        VALUES (p_panel_project_id, p_panel_no, p_port_no, p_cable_pin, p_threshold_value)
        RETURNING id INTO v_rtn_value;
    ELSE
        UPDATE IoTPanelThreshold
        SET PanelProjectId = p_panel_project_id,
            PanelNo = p_panel_no,
            PortNo = p_port_no,
            CablePin = p_cable_pin,
            Threshold_Value = p_threshold_value
        WHERE PanelProjectId = p_panel_project_id 
        AND PanelNo = p_panel_no 
        AND PortNo = p_port_no 
        AND CablePin = p_cable_pin;
        v_rtn_value := p_id;
    END IF;
    
    RETURN v_rtn_value;
END;
$$;


ALTER FUNCTION public.iotpanelthreshold_save(p_id integer, p_panel_project_id integer, p_panel_no integer, p_port_no integer, p_cable_pin integer, p_threshold_value numeric) OWNER TO postgres;

--
-- Name: notification_getbyprojectid(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.notification_getbyprojectid(p_panel_project_id integer) RETURNS TABLE(id integer, notificationto character varying, notificationcc character varying, subject character varying, panelprojectid integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT * FROM Notification WHERE PanelProjectId = p_panel_project_id;
END;
$$;


ALTER FUNCTION public.notification_getbyprojectid(p_panel_project_id integer) OWNER TO postgres;

--
-- Name: notification_insertupdate(character varying, character varying, character varying, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.notification_insertupdate(p_notification_to character varying, p_notification_cc character varying, p_subject character varying, p_panel_project_id integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_rtn_value INTEGER;
BEGIN
    IF NOT EXISTS (SELECT * FROM Notification WHERE PanelProjectId = p_panel_project_id) THEN
        INSERT INTO Notification(NotificationTo, NotificationCC, Subject, PanelProjectId)
        VALUES (p_notification_to, p_notification_cc, p_subject, p_panel_project_id)
        RETURNING id INTO v_rtn_value;
    ELSE
        UPDATE Notification
        SET NotificationTo = p_notification_to,
            NotificationCC = p_notification_cc,
            Subject = p_subject
        WHERE PanelProjectId = p_panel_project_id;
        v_rtn_value := p_panel_project_id;
    END IF;
    
    RETURN v_rtn_value;
END;
$$;


ALTER FUNCTION public.notification_insertupdate(p_notification_to character varying, p_notification_cc character varying, p_subject character varying, p_panel_project_id integer) OWNER TO postgres;

--
-- Name: notificationhistory_insertupdate(integer, character varying, character varying, character varying, integer, numeric, numeric); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.notificationhistory_insertupdate(p_id integer, p_notification_to character varying, p_notification_cc character varying, p_subject character varying, p_triggered_by_bundle integer, p_panel_temp numeric, p_threshold numeric) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_rtn_value INTEGER;
BEGIN
    IF p_id = 0 THEN
        INSERT INTO NotificationHistory(
            NotificationTo,
            NotificationCC,
            Subject,
            TriggeredByBundle,
            PanelTemp,
            Threshold,
            CreatedOn
        )
        VALUES (
            p_notification_to,
            p_notification_cc,
            p_subject,
            p_triggered_by_bundle,
            p_panel_temp,
            p_threshold,
            CURRENT_TIMESTAMP
        )
        RETURNING id INTO v_rtn_value;
    ELSE
        UPDATE NotificationHistory
        SET NotificationTo = p_notification_to,
            NotificationCC = p_notification_cc,
            Subject = p_subject,
            TriggeredByBundle = p_triggered_by_bundle,
            PanelTemp = p_panel_temp,
            Threshold = p_threshold
        WHERE Id = p_id;
        v_rtn_value := p_id;
    END IF;
    
    RETURN v_rtn_value;
END;
$$;


ALTER FUNCTION public.notificationhistory_insertupdate(p_id integer, p_notification_to character varying, p_notification_cc character varying, p_subject character varying, p_triggered_by_bundle integer, p_panel_temp numeric, p_threshold numeric) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: io_connectivity; Type: TABLE; Schema: nflexon_app; Owner: postgres
--

CREATE TABLE nflexon_app.io_connectivity (
    io_mac character varying NOT NULL,
    io_type character varying NOT NULL,
    io_port integer NOT NULL,
    device character varying,
    device_mac character varying
);


ALTER TABLE nflexon_app.io_connectivity OWNER TO postgres;

--
-- Name: io_location; Type: TABLE; Schema: nflexon_app; Owner: postgres
--

CREATE TABLE nflexon_app.io_location (
    io_type character varying,
    io_mac character varying NOT NULL,
    site character varying,
    building character varying,
    floor integer,
    room character varying,
    additional_description character varying
);


ALTER TABLE nflexon_app.io_location OWNER TO postgres;

--
-- Name: pp_connectivity; Type: TABLE; Schema: nflexon_app; Owner: postgres
--

CREATE TABLE nflexon_app.pp_connectivity (
    pp_serial_no character varying NOT NULL,
    ru integer NOT NULL,
    pp_port integer NOT NULL,
    io_mac character varying,
    io_port integer
);


ALTER TABLE nflexon_app.pp_connectivity OWNER TO postgres;

--
-- Name: pp_location; Type: TABLE; Schema: nflexon_app; Owner: postgres
--

CREATE TABLE nflexon_app.pp_location (
    pp_serial_no character varying NOT NULL,
    pp_mac character varying,
    site character varying,
    building character varying,
    floor integer,
    room character varying,
    rack character varying
);


ALTER TABLE nflexon_app.pp_location OWNER TO postgres;

--
-- Name: switch; Type: TABLE; Schema: nflexon_app; Owner: postgres
--

CREATE TABLE nflexon_app.switch (
    switch_name character varying NOT NULL,
    switch_port integer NOT NULL,
    pp_serial_no character varying,
    ru integer,
    pp_port integer
);


ALTER TABLE nflexon_app.switch OWNER TO postgres;

--
-- Name: ambienttempdetail; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ambienttempdetail (
    tempdetailid integer NOT NULL,
    panelprojectid integer,
    threshold_value numeric(10,2),
    readingintervalinmin integer
);


ALTER TABLE public.ambienttempdetail OWNER TO postgres;

--
-- Name: ambienttempdetail_tempdetailid_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.ambienttempdetail_tempdetailid_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.ambienttempdetail_tempdetailid_seq OWNER TO postgres;

--
-- Name: ambienttempdetail_tempdetailid_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ambienttempdetail_tempdetailid_seq OWNED BY public.ambienttempdetail.tempdetailid;


--
-- Name: ambienttempsensors; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ambienttempsensors (
    panelprojectid integer NOT NULL,
    ip_address character varying(100) NOT NULL,
    notes character varying(100)
);


ALTER TABLE public.ambienttempsensors OWNER TO postgres;

--
-- Name: bundledetail; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.bundledetail (
    detailid integer NOT NULL,
    bundleid integer,
    panelno integer,
    portno integer,
    actiongroupid integer
);


ALTER TABLE public.bundledetail OWNER TO postgres;

--
-- Name: bundledetail_detailid_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.bundledetail_detailid_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.bundledetail_detailid_seq OWNER TO postgres;

--
-- Name: bundledetail_detailid_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.bundledetail_detailid_seq OWNED BY public.bundledetail.detailid;


--
-- Name: bundlemaster; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.bundlemaster (
    bundleid integer NOT NULL,
    no_of_cables integer,
    temppanelno integer,
    threshold_value numeric(10,2),
    panelprojectid integer,
    bundlename character varying(100)
);


ALTER TABLE public.bundlemaster OWNER TO postgres;

--
-- Name: bundlemaster_bundleid_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.bundlemaster_bundleid_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.bundlemaster_bundleid_seq OWNER TO postgres;

--
-- Name: bundlemaster_bundleid_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.bundlemaster_bundleid_seq OWNED BY public.bundlemaster.bundleid;


--
-- Name: configuration; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.configuration (
    id integer NOT NULL,
    smtp character varying(100),
    port integer,
    userid character varying(200),
    password character varying(500),
    tempunit integer,
    panelprojectid integer
);


ALTER TABLE public.configuration OWNER TO postgres;

--
-- Name: configuration_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.configuration_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.configuration_id_seq OWNER TO postgres;

--
-- Name: configuration_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.configuration_id_seq OWNED BY public.configuration.id;


--
-- Name: history; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.history (
    panelprojectid integer NOT NULL,
    panelslaveaddress integer NOT NULL,
    portno integer,
    cableno integer,
    ambienttemp numeric(10,2),
    portstatus integer,
    createdon timestamp without time zone
);


ALTER TABLE public.history OWNER TO postgres;

--
-- Name: iotpanelproject; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.iotpanelproject (
    id integer NOT NULL,
    ip_address character varying(100),
    no_of_panels integer,
    uniqueidentification character varying(100),
    portno integer
);


ALTER TABLE public.iotpanelproject OWNER TO postgres;

--
-- Name: iotpanelproject_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.iotpanelproject_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.iotpanelproject_id_seq OWNER TO postgres;

--
-- Name: iotpanelproject_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.iotpanelproject_id_seq OWNED BY public.iotpanelproject.id;


--
-- Name: iotpanelthreshold; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.iotpanelthreshold (
    id integer NOT NULL,
    panelprojectid integer,
    panelno integer,
    portno integer,
    cablepin integer,
    threshold_value numeric(10,2)
);


ALTER TABLE public.iotpanelthreshold OWNER TO postgres;

--
-- Name: iotpanelthreshold_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.iotpanelthreshold_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.iotpanelthreshold_id_seq OWNER TO postgres;

--
-- Name: iotpanelthreshold_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.iotpanelthreshold_id_seq OWNED BY public.iotpanelthreshold.id;


--
-- Name: notification; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.notification (
    id integer NOT NULL,
    notificationto character varying(500),
    notificationcc character varying(500),
    subject character varying(500),
    panelprojectid integer
);


ALTER TABLE public.notification OWNER TO postgres;

--
-- Name: notification_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.notification_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.notification_id_seq OWNER TO postgres;

--
-- Name: notification_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.notification_id_seq OWNED BY public.notification.id;


--
-- Name: notificationhistory; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.notificationhistory (
    id integer NOT NULL,
    notificationto character varying(500),
    notificationcc character varying(500),
    subject character varying(500),
    triggeredbybundle integer,
    paneltemp numeric(10,2),
    threshold numeric(10,2),
    createdon timestamp without time zone
);


ALTER TABLE public.notificationhistory OWNER TO postgres;

--
-- Name: notificationhistory_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.notificationhistory_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.notificationhistory_id_seq OWNER TO postgres;

--
-- Name: notificationhistory_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.notificationhistory_id_seq OWNED BY public.notificationhistory.id;


--
-- Name: ambienttempdetail tempdetailid; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ambienttempdetail ALTER COLUMN tempdetailid SET DEFAULT nextval('public.ambienttempdetail_tempdetailid_seq'::regclass);


--
-- Name: bundledetail detailid; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.bundledetail ALTER COLUMN detailid SET DEFAULT nextval('public.bundledetail_detailid_seq'::regclass);


--
-- Name: bundlemaster bundleid; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.bundlemaster ALTER COLUMN bundleid SET DEFAULT nextval('public.bundlemaster_bundleid_seq'::regclass);


--
-- Name: configuration id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.configuration ALTER COLUMN id SET DEFAULT nextval('public.configuration_id_seq'::regclass);


--
-- Name: iotpanelproject id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.iotpanelproject ALTER COLUMN id SET DEFAULT nextval('public.iotpanelproject_id_seq'::regclass);


--
-- Name: iotpanelthreshold id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.iotpanelthreshold ALTER COLUMN id SET DEFAULT nextval('public.iotpanelthreshold_id_seq'::regclass);


--
-- Name: notification id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification ALTER COLUMN id SET DEFAULT nextval('public.notification_id_seq'::regclass);


--
-- Name: notificationhistory id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notificationhistory ALTER COLUMN id SET DEFAULT nextval('public.notificationhistory_id_seq'::regclass);


--
-- Data for Name: io_connectivity; Type: TABLE DATA; Schema: nflexon_app; Owner: postgres
--

COPY nflexon_app.io_connectivity (io_mac, io_type, io_port, device, device_mac) FROM stdin;
34:94:54:62:54:58	FP6	1	PTZ Camera	dd:5c:4b:90:a0:3a
34:94:54:62:54:58	FP6	2	PTZ Camera	f4:d2:81:97:85:42
34:94:54:62:54:58	FP6	3	PTZ Camera	1e:1d:0e:c2:55:bf
34:94:54:62:54:58	FP6	4	PTZ Camera	0a:10:12:67:bd:67
34:94:54:62:54:58	FP6	5	Printer	eb:dc:27:04:ab:72
34:94:54:62:54:58	FP6	6	Printer	da:44:bd:98:47:83
34:94:54:62:52:9B	FP4	1	Printer	fb:18:54:91:bb:f6
34:94:54:62:52:9B	FP4	2	Printer	5e:55:ba:de:67:14
34:94:54:62:52:9B	FP4	3	PTZ Camera	76:ae:47:42:55:2c
34:94:54:62:52:9B	FP4	4	PTZ Camera	e2:13:79:5e:af:cc
9f:50:e7:e0:bb:22	FP2	1	PTZ Camera	60:4c:fc:11:c4:ce
9f:50:e7:e0:bb:22	FP2	2	PTZ Camera	e6:8b:1c:fc:f7:34
24:6b:28:b4:3f:d4	SB4	1	Printer	d6:0a:78:eb:df:f4
24:6b:28:b4:3f:d4	SB4	2	Printer	21:87:cb:70:2e:ef
24:6b:28:b4:3f:d4	SB4	3	Printer	84:4c:e5:22:09:72
24:6b:28:b4:3f:d4	SB4	4	Printer	c6:7f:4d:c0:77:a1
c9:56:bc:8f:c4:f3	SB2	1	PTZ Camera	5f:80:ac:a5:11:81
c9:56:bc:8f:c4:f3	SB2	2	PTZ Camera	dd:ee:93:1d:a1:21
46:8d:15:d3:02:b6	SB2	1	PTZ Camera	e7:01:86:2b:26:c9
46:8d:15:d3:02:b6	SB2	2	PTZ Camera	24:c9:24:59:9b:0a
\.


--
-- Data for Name: io_location; Type: TABLE DATA; Schema: nflexon_app; Owner: postgres
--

COPY nflexon_app.io_location (io_type, io_mac, site, building, floor, room, additional_description) FROM stdin;
FP6	34:94:54:62:54:58	Allen	700 Central	4	Reception Area	Below the Main Desk
FP4	34:94:54:62:52:9B	Allen	700 Central	4	Cafeteria	On the right of the TV
FP2	9f:50:e7:e0:bb:22	Parker	350 Main	1	Conference Room 1	Rightmost below the conference table
SB4	24:6b:28:b4:3f:d4	Parker	350 Main	1	Conference Room 2	Above the Ceiling - Far Right Corner 
SB2	c9:56:bc:8f:c4:f3	Parker	350 Main	1	Cafeteria	Above the Celing - Entrance Door of Cafeteria
SB2	46:8d:15:d3:02:b6	Parker	350 Main	1	Reception Area	Above the Ceiling - Far Left Corner of Reception Area 
\.


--
-- Data for Name: pp_connectivity; Type: TABLE DATA; Schema: nflexon_app; Owner: postgres
--

COPY nflexon_app.pp_connectivity (pp_serial_no, ru, pp_port, io_mac, io_port) FROM stdin;
1A04001650314D3531382D	2	8	34:94:54:62:54:58	1
1A04001650314D3531382D	6	16	34:94:54:62:54:58	2
1A04001650314D3531382D	3	12	34:94:54:62:54:58	3
1A04001650314D3531382D	6	18	34:94:54:62:54:58	4
1A04001650314D3531382D	3	8	34:94:54:62:54:58	5
1A04001650314D3531382D	2	4	34:94:54:62:54:58	6
1A04001650314D3531382D	8	4	34:94:54:62:52:9B	1
1A04001650314D3531382D	5	14	34:94:54:62:52:9B	2
1A04001650314D3531382D	5	19	34:94:54:62:52:9B	3
1A04001650314D3531382D	8	23	34:94:54:62:52:9B	4
1E04101650314D3531382D	5	9	9f:50:e7:e0:bb:22	1
1E04101650314D3531382D	5	3	9f:50:e7:e0:bb:22	2
1E04101650314D3531382D	3	4	24:6b:28:b4:3f:d4	1
1E04101650314D3531382D	6	6	24:6b:28:b4:3f:d4	2
1E04101650314D3531382D	7	6	24:6b:28:b4:3f:d4	3
1E04101650314D3531382D	4	1	24:6b:28:b4:3f:d4	4
1E04101650314D3531382D	6	16	c9:56:bc:8f:c4:f3	1
1E04101650314D3531382D	8	12	c9:56:bc:8f:c4:f3	2
1E04101650314D3531382D	4	9	46:8d:15:d3:02:b6	1
1E04101650314D3531382D	7	4	46:8d:15:d3:02:b6	2
\.


--
-- Data for Name: pp_location; Type: TABLE DATA; Schema: nflexon_app; Owner: postgres
--

COPY nflexon_app.pp_location (pp_serial_no, pp_mac, site, building, floor, room, rack) FROM stdin;
1A04001650314D3531382D	34:94:54:62:6E:EB	Allen	700 Central	4	Equipment Room 2	2
1E04101650314D3531382D	34:94:54:62:70:37	Parker	350 Main	1	Telecom Room	1
\.


--
-- Data for Name: switch; Type: TABLE DATA; Schema: nflexon_app; Owner: postgres
--

COPY nflexon_app.switch (switch_name, switch_port, pp_serial_no, ru, pp_port) FROM stdin;
NETGEAR_M	1	1A04001650314D3531382D	3	23
NETGEAR_M	2	1A04001650314D3531382D	8	9
NETGEAR_M	3	1A04001650314D3531382D	1	22
NETGEAR_M	4	1E04101650314D3531382D	1	21
NETGEAR_M	5	1A04001650314D3531382D	2	14
NETGEAR_M	6	1A04001650314D3531382D	7	11
NETGEAR_M	7	1A04001650314D3531382D	6	1
NETGEAR_M	8	1E04101650314D3531382D	2	10
NETGEAR_M	9	1A04001650314D3531382D	4	7
NETGEAR_M	10	1E04101650314D3531382D	3	8
\.


--
-- Data for Name: ambienttempdetail; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.ambienttempdetail (tempdetailid, panelprojectid, threshold_value, readingintervalinmin) FROM stdin;
\.


--
-- Data for Name: ambienttempsensors; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.ambienttempsensors (panelprojectid, ip_address, notes) FROM stdin;
\.


--
-- Data for Name: bundledetail; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.bundledetail (detailid, bundleid, panelno, portno, actiongroupid) FROM stdin;
\.


--
-- Data for Name: bundlemaster; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.bundlemaster (bundleid, no_of_cables, temppanelno, threshold_value, panelprojectid, bundlename) FROM stdin;
\.


--
-- Data for Name: configuration; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.configuration (id, smtp, port, userid, password, tempunit, panelprojectid) FROM stdin;
\.


--
-- Data for Name: history; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.history (panelprojectid, panelslaveaddress, portno, cableno, ambienttemp, portstatus, createdon) FROM stdin;
\.


--
-- Data for Name: iotpanelproject; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.iotpanelproject (id, ip_address, no_of_panels, uniqueidentification, portno) FROM stdin;
\.


--
-- Data for Name: iotpanelthreshold; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.iotpanelthreshold (id, panelprojectid, panelno, portno, cablepin, threshold_value) FROM stdin;
\.


--
-- Data for Name: notification; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.notification (id, notificationto, notificationcc, subject, panelprojectid) FROM stdin;
\.


--
-- Data for Name: notificationhistory; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.notificationhistory (id, notificationto, notificationcc, subject, triggeredbybundle, paneltemp, threshold, createdon) FROM stdin;
\.


--
-- Name: ambienttempdetail_tempdetailid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.ambienttempdetail_tempdetailid_seq', 1, false);


--
-- Name: bundledetail_detailid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.bundledetail_detailid_seq', 1, false);


--
-- Name: bundlemaster_bundleid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.bundlemaster_bundleid_seq', 1, false);


--
-- Name: configuration_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.configuration_id_seq', 1, false);


--
-- Name: iotpanelproject_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.iotpanelproject_id_seq', 1, false);


--
-- Name: iotpanelthreshold_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.iotpanelthreshold_id_seq', 1, false);


--
-- Name: notification_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.notification_id_seq', 1, false);


--
-- Name: notificationhistory_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.notificationhistory_id_seq', 1, false);


--
-- Name: io_connectivity io_connectivity_pkey; Type: CONSTRAINT; Schema: nflexon_app; Owner: postgres
--

ALTER TABLE ONLY nflexon_app.io_connectivity
    ADD CONSTRAINT io_connectivity_pkey PRIMARY KEY (io_mac, io_type, io_port);


--
-- Name: io_location io_location_pkey; Type: CONSTRAINT; Schema: nflexon_app; Owner: postgres
--

ALTER TABLE ONLY nflexon_app.io_location
    ADD CONSTRAINT io_location_pkey PRIMARY KEY (io_mac);


--
-- Name: pp_connectivity pp_connectivity_pkey; Type: CONSTRAINT; Schema: nflexon_app; Owner: postgres
--

ALTER TABLE ONLY nflexon_app.pp_connectivity
    ADD CONSTRAINT pp_connectivity_pkey PRIMARY KEY (pp_serial_no, ru, pp_port);


--
-- Name: pp_location pp_location_pkey; Type: CONSTRAINT; Schema: nflexon_app; Owner: postgres
--

ALTER TABLE ONLY nflexon_app.pp_location
    ADD CONSTRAINT pp_location_pkey PRIMARY KEY (pp_serial_no);


--
-- Name: switch switch_pkey; Type: CONSTRAINT; Schema: nflexon_app; Owner: postgres
--

ALTER TABLE ONLY nflexon_app.switch
    ADD CONSTRAINT switch_pkey PRIMARY KEY (switch_name, switch_port);


--
-- Name: ambienttempdetail ambienttempdetail_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ambienttempdetail
    ADD CONSTRAINT ambienttempdetail_pkey PRIMARY KEY (tempdetailid);


--
-- Name: bundledetail bundledetail_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.bundledetail
    ADD CONSTRAINT bundledetail_pkey PRIMARY KEY (detailid);


--
-- Name: bundlemaster bundlemaster_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.bundlemaster
    ADD CONSTRAINT bundlemaster_pkey PRIMARY KEY (bundleid);


--
-- Name: configuration configuration_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.configuration
    ADD CONSTRAINT configuration_pkey PRIMARY KEY (id);


--
-- Name: iotpanelproject iotpanelproject_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.iotpanelproject
    ADD CONSTRAINT iotpanelproject_pkey PRIMARY KEY (id);


--
-- Name: iotpanelthreshold iotpanelthreshold_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.iotpanelthreshold
    ADD CONSTRAINT iotpanelthreshold_pkey PRIMARY KEY (id);


--
-- Name: notification notification_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT notification_pkey PRIMARY KEY (id);


--
-- Name: notificationhistory notificationhistory_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notificationhistory
    ADD CONSTRAINT notificationhistory_pkey PRIMARY KEY (id);


--
-- Name: io_connectivity io_connectivity_io_mac_fkey; Type: FK CONSTRAINT; Schema: nflexon_app; Owner: postgres
--

ALTER TABLE ONLY nflexon_app.io_connectivity
    ADD CONSTRAINT io_connectivity_io_mac_fkey FOREIGN KEY (io_mac) REFERENCES nflexon_app.io_location(io_mac);


--
-- Name: pp_connectivity pp_connectivity_io_mac_fkey; Type: FK CONSTRAINT; Schema: nflexon_app; Owner: postgres
--

ALTER TABLE ONLY nflexon_app.pp_connectivity
    ADD CONSTRAINT pp_connectivity_io_mac_fkey FOREIGN KEY (io_mac) REFERENCES nflexon_app.io_location(io_mac);


--
-- Name: pp_connectivity pp_connectivity_pp_serial_no_fkey; Type: FK CONSTRAINT; Schema: nflexon_app; Owner: postgres
--

ALTER TABLE ONLY nflexon_app.pp_connectivity
    ADD CONSTRAINT pp_connectivity_pp_serial_no_fkey FOREIGN KEY (pp_serial_no) REFERENCES nflexon_app.pp_location(pp_serial_no);


--
-- Name: switch switch_pp_serial_no_fkey; Type: FK CONSTRAINT; Schema: nflexon_app; Owner: postgres
--

ALTER TABLE ONLY nflexon_app.switch
    ADD CONSTRAINT switch_pp_serial_no_fkey FOREIGN KEY (pp_serial_no) REFERENCES nflexon_app.pp_location(pp_serial_no);


--
-- PostgreSQL database dump complete
--

