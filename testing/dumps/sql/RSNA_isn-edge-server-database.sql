-- Downloaded from: https://github.com/RSNA/isn-edge-server-database/blob/2051d17746da0ca8e4e0d5c2667cbd0538b85e81/rsna.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 9.5.1
-- Dumped by pg_dump version 9.5.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: rsnadb; Type: DATABASE; Schema: -; Owner: edge
--

-- For UNIX system
CREATE DATABASE rsnadb WITH TEMPLATE = template0 ENCODING = 'UTF8' LC_COLLATE = 'en_US.UTF-8' LC_CTYPE = 'en_US.UTF-8';

-- For Windows system
-- CREATE DATABASE rsnadb WITH TEMPLATE = template0 ENCODING = 'UTF8' LC_COLLATE = 'English_United States.1252' LC_CTYPE = 'English_United States.1252';

ALTER DATABASE rsnadb OWNER TO edge;

\connect rsnadb

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: rsnadb; Type: COMMENT; Schema: -; Owner: edge
--

COMMENT ON DATABASE rsnadb IS 'RSNA Edge Device Database
Authors: Wendy Zhu (Univ of Chicago) and Steve G Langer (Mayo Clinic)

Copyright (c) 2010, Radiological Society of North America
  All rights reserved.
  Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:
  Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
  Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.
  Neither the name of the RSNA nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.

  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND
  CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED
  WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
  WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A
  PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL
  THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY
  DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
  CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
  PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF
  USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
  CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
  CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
  NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE
  USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY
  OF SUCH DAMAGE.';


--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

--
-- Name: fn_exam_autosend(); Type: FUNCTION; Schema: public; Owner: edge
--

CREATE FUNCTION fn_exam_autosend() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
	v_email_address character varying(255);
	v_single_use_patient_id character varying(64);
	v_access_code character varying(64);
	v_max_retries integer;
	v_user_id integer;
	v_job_set_id integer;
	v_job_id integer;
BEGIN
	-- the exam belongs to patient with autosend flag is true
	IF (SELECT autosend FROM patients WHERE patient_id=NEW.patient_id) THEN
		SELECT email_address,single_use_patient_id,access_code
		INTO v_email_address,v_single_use_patient_id,v_access_code
		FROM job_sets
		WHERE patient_id = NEW.patient_id
		ORDER BY modified_date DESC
		FETCH FIRST 1 ROW ONLY;

		IF v_single_use_patient_id IS NOT NULL THEN
			SELECT value INTO v_max_retries FROM configurations WHERE key='max-retries';
			SELECT user_id INTO v_user_id FROM users WHERE user_login='AUTOSEND';

			INSERT INTO job_sets (patient_id,user_id,email_address,single_use_patient_id,access_code)
			VALUES (NEW.patient_id,v_user_id,v_email_address,v_single_use_patient_id,v_access_code)
			RETURNING job_set_id INTO v_job_set_id;

			INSERT INTO jobs (job_set_id,exam_id,remaining_retries)
			VALUES (v_job_set_id,NEW.exam_id,v_max_retries)
			RETURNING job_id INTO v_job_id;

			INSERT INTO transactions (job_id,status_code,comments)
			VALUES (v_job_id,1,'Queued');
		END IF;
	END IF;
	RETURN NEW;
END;
$$;


ALTER FUNCTION public.fn_exam_autosend() OWNER TO edge;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: configurations; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE configurations (
    key character varying NOT NULL,
    value character varying NOT NULL,
    modified_date timestamp with time zone DEFAULT now()
);


ALTER TABLE configurations OWNER TO edge;

--
-- Name: TABLE configurations; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE configurations IS 'This table is used to store applications specific config data as key/value pairs and takes the place of java properties files (rather then having it all aly about in plain text files);

a) paths to key things (ie dicom studies)
b) site prefix for generating RSNA ID''s
c) site delay for applying to report finalize before available to send to CH
d) Clearing House connection data
e) etc';


--
-- Name: devices; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE devices (
    device_id integer NOT NULL,
    ae_title character varying(256) NOT NULL,
    host character varying(256) NOT NULL,
    port_number character varying(10) NOT NULL,
    modified_date timestamp with time zone DEFAULT now()
);


ALTER TABLE devices OWNER TO edge;

--
-- Name: TABLE devices; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE devices IS 'Used to store DICOM connection info for mage sources, and possibly others

a) the DICOM triplet (for remote DICOM study sources)
b) ?

';


--
-- Name: devices_device_id_seq; Type: SEQUENCE; Schema: public; Owner: edge
--

CREATE SEQUENCE devices_device_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE devices_device_id_seq OWNER TO edge;

--
-- Name: devices_device_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: edge
--

ALTER SEQUENCE devices_device_id_seq OWNED BY devices.device_id;


--
-- Name: email_configurations; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE email_configurations (
    key character varying NOT NULL,
    value character varying NOT NULL,
    modified_date timestamp with time zone DEFAULT now()
);


ALTER TABLE email_configurations OWNER TO edge;

--
-- Name: TABLE email_configurations; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE email_configurations IS 'This table is used to store email configuration as key/value pairs';


--
-- Name: email_jobs; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE email_jobs (
    email_job_id integer NOT NULL,
    recipient character varying NOT NULL,
    subject character varying,
    body text,
    sent boolean DEFAULT false NOT NULL,
    failed boolean DEFAULT false NOT NULL,
    comments character varying,
    created_date timestamp with time zone NOT NULL,
    modified_date timestamp with time zone DEFAULT now()
);


ALTER TABLE email_jobs OWNER TO edge;

--
-- Name: TABLE email_jobs; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE email_jobs IS 'This table is used to store queued emails. Jobs within the queue will be handled by a worker thread which is responsible for handling any send failures and retrying failed jobs';


--
-- Name: email_jobs_email_job_id_seq; Type: SEQUENCE; Schema: public; Owner: edge
--

CREATE SEQUENCE email_jobs_email_job_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE email_jobs_email_job_id_seq OWNER TO edge;

--
-- Name: email_jobs_email_job_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: edge
--

ALTER SEQUENCE email_jobs_email_job_id_seq OWNED BY email_jobs.email_job_id;


--
-- Name: exams; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE exams (
    exam_id integer NOT NULL,
    accession_number character varying(50) NOT NULL,
    patient_id integer NOT NULL,
    exam_description character varying(256),
    modified_date timestamp with time zone DEFAULT now()
);


ALTER TABLE exams OWNER TO edge;

--
-- Name: TABLE exams; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE exams IS 'A listing of all ordered DICOM exams the system knows about. The report status is not stored here';


--
-- Name: exams_exam_id_seq; Type: SEQUENCE; Schema: public; Owner: edge
--

CREATE SEQUENCE exams_exam_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE exams_exam_id_seq OWNER TO edge;

--
-- Name: exams_exam_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: edge
--

ALTER SEQUENCE exams_exam_id_seq OWNED BY exams.exam_id;


--
-- Name: hipaa_audit_accession_numbers; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE hipaa_audit_accession_numbers (
    id integer NOT NULL,
    view_id integer,
    accession_number character varying(100),
    modified_date timestamp with time zone DEFAULT now()
);


ALTER TABLE hipaa_audit_accession_numbers OWNER TO edge;

--
-- Name: TABLE hipaa_audit_accession_numbers; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE hipaa_audit_accession_numbers IS 'Part of the HIPAA tracking for edge device auditing. This table and  "audit_mrns" report up to table HIPAA_views
';


--
-- Name: hipaa_audit_accession_numbers_id_seq; Type: SEQUENCE; Schema: public; Owner: edge
--

CREATE SEQUENCE hipaa_audit_accession_numbers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE hipaa_audit_accession_numbers_id_seq OWNER TO edge;

--
-- Name: hipaa_audit_accession_numbers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: edge
--

ALTER SEQUENCE hipaa_audit_accession_numbers_id_seq OWNED BY hipaa_audit_accession_numbers.id;


--
-- Name: hipaa_audit_mrns; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE hipaa_audit_mrns (
    id integer NOT NULL,
    view_id integer,
    mrn character varying(100),
    modified_date timestamp with time zone DEFAULT now()
);


ALTER TABLE hipaa_audit_mrns OWNER TO edge;

--
-- Name: TABLE hipaa_audit_mrns; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE hipaa_audit_mrns IS 'Part of the HIPAA tracking for edge device auditing. This table and  "audit_acessions" report up to table HIPAA_views
';


--
-- Name: hipaa_audit_mrns_id_seq; Type: SEQUENCE; Schema: public; Owner: edge
--

CREATE SEQUENCE hipaa_audit_mrns_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE hipaa_audit_mrns_id_seq OWNER TO edge;

--
-- Name: hipaa_audit_mrns_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: edge
--

ALTER SEQUENCE hipaa_audit_mrns_id_seq OWNED BY hipaa_audit_mrns.id;


--
-- Name: hipaa_audit_views; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE hipaa_audit_views (
    id integer NOT NULL,
    requesting_ip character varying(15),
    requesting_username character varying(100),
    requesting_uri text,
    modified_date timestamp with time zone DEFAULT now()
);


ALTER TABLE hipaa_audit_views OWNER TO edge;

--
-- Name: TABLE hipaa_audit_views; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE hipaa_audit_views IS 'Part of the HIPAA tracking for edge device auditing. This is the top level table that tracks who asked for what from where. The HIPAA tables "audfit_accession" and "audit_mrns" report up to this table';


--
-- Name: hipaa_audit_views_id_seq; Type: SEQUENCE; Schema: public; Owner: edge
--

CREATE SEQUENCE hipaa_audit_views_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE hipaa_audit_views_id_seq OWNER TO edge;

--
-- Name: hipaa_audit_views_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: edge
--

ALTER SEQUENCE hipaa_audit_views_id_seq OWNED BY hipaa_audit_views.id;


--
-- Name: job_sets; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE job_sets (
    job_set_id integer NOT NULL,
    patient_id integer NOT NULL,
    user_id integer NOT NULL,
    email_address character varying(255),
    modified_date timestamp with time zone DEFAULT now(),
    delay_in_hrs integer DEFAULT 72,
    single_use_patient_id character varying(64) NOT NULL,
    send_on_complete boolean DEFAULT false NOT NULL,
    access_code character varying(64),
    send_to_site boolean DEFAULT false NOT NULL,
    phone_number character varying(20),
    global_id character varying(64),
    global_aa character varying(64)
);


ALTER TABLE job_sets OWNER TO edge;

--
-- Name: TABLE job_sets; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE job_sets IS 'This is one of a pair of tables that bind a patient to a edge device job, consisting of one or more exam accessions descrbing DICOM exams to send to the CH. The other table is JOBS
';


--
-- Name: job_sets_job_set_id_seq; Type: SEQUENCE; Schema: public; Owner: edge
--

CREATE SEQUENCE job_sets_job_set_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE job_sets_job_set_id_seq OWNER TO edge;

--
-- Name: job_sets_job_set_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: edge
--

ALTER SEQUENCE job_sets_job_set_id_seq OWNED BY job_sets.job_set_id;


--
-- Name: jobs; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE jobs (
    job_id integer NOT NULL,
    job_set_id integer NOT NULL,
    exam_id integer NOT NULL,
    report_id integer,
    document_id character varying(100),
    remaining_retries integer NOT NULL,
    modified_date timestamp with time zone DEFAULT now()
);


ALTER TABLE jobs OWNER TO edge;

--
-- Name: TABLE jobs; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE jobs IS 'This is one of a pair of tables that bind a patient to a edge device job, consisting of one or more exam accessions descrbing DICOM exams to send to the CH. The other table is JOB_SETS
';


--
-- Name: jobs_job_id_seq; Type: SEQUENCE; Schema: public; Owner: edge
--

CREATE SEQUENCE jobs_job_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE jobs_job_id_seq OWNER TO edge;

--
-- Name: jobs_job_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: edge
--

ALTER SEQUENCE jobs_job_id_seq OWNED BY jobs.job_id;


--
-- Name: patient_merge_events; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE patient_merge_events (
    event_id integer NOT NULL,
    old_mrn character varying(50) NOT NULL,
    new_mrn character varying(50) NOT NULL,
    old_patient_id integer NOT NULL,
    new_patient_id integer NOT NULL,
    status integer DEFAULT 0 NOT NULL,
    modified_date timestamp with time zone DEFAULT now()
);


ALTER TABLE patient_merge_events OWNER TO edge;

--
-- Name: TABLE patient_merge_events; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE patient_merge_events IS 'When it''s required to swap a patient to a new ID (say a john doe) this tracks the old and new MRN for rollback/auditing
';


--
-- Name: patient_merge_events_event_id_seq; Type: SEQUENCE; Schema: public; Owner: edge
--

CREATE SEQUENCE patient_merge_events_event_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE patient_merge_events_event_id_seq OWNER TO edge;

--
-- Name: patient_merge_events_event_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: edge
--

ALTER SEQUENCE patient_merge_events_event_id_seq OWNED BY patient_merge_events.event_id;


--
-- Name: patients; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE patients (
    patient_id integer NOT NULL,
    mrn character varying(50) NOT NULL,
    patient_name character varying,
    dob date,
    sex character(1),
    street character varying,
    city character varying(50),
    state character varying(30),
    zip_code character varying(30),
    email_address character varying(255),
    rsna_id character varying(64),
    modified_date timestamp with time zone DEFAULT now(),
    consent_timestamp timestamp with time zone,
    autosend boolean DEFAULT false
);


ALTER TABLE patients OWNER TO edge;

--
-- Name: TABLE patients; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE patients IS 'a list of all patient demog sent via the HL7 MIRTH channel';


--
-- Name: COLUMN patients.patient_id; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON COLUMN patients.patient_id IS 'just the dbase created primary key';


--
-- Name: COLUMN patients.mrn; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON COLUMN patients.mrn IS 'the actual medical recrod number from the medical center';


--
-- Name: patients_patient_id_seq; Type: SEQUENCE; Schema: public; Owner: edge
--

CREATE SEQUENCE patients_patient_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE patients_patient_id_seq OWNER TO edge;

--
-- Name: patients_patient_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: edge
--

ALTER SEQUENCE patients_patient_id_seq OWNED BY patients.patient_id;


--
-- Name: reports; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE reports (
    report_id integer NOT NULL,
    exam_id integer NOT NULL,
    proc_code character varying,
    status character varying NOT NULL,
    status_timestamp timestamp with time zone NOT NULL,
    report_text text,
    signer character varying,
    dictator character varying,
    transcriber character varying,
    modified_date timestamp with time zone DEFAULT now()
);


ALTER TABLE reports OWNER TO edge;

--
-- Name: TABLE reports; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE reports IS 'This table contains exam report and exam status as sent from the MIRTH HL7 channel. Combined with the Exams table, this provides all info needed to determine exam staus and location to create a job to send to the CH';


--
-- Name: reports_report_id_seq; Type: SEQUENCE; Schema: public; Owner: edge
--

CREATE SEQUENCE reports_report_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE reports_report_id_seq OWNER TO edge;

--
-- Name: reports_report_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: edge
--

ALTER SEQUENCE reports_report_id_seq OWNED BY reports.report_id;


--
-- Name: roles; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE roles (
    role_id integer NOT NULL,
    role_description character varying(50) NOT NULL,
    modified_date time with time zone DEFAULT now()
);


ALTER TABLE roles OWNER TO edge;

--
-- Name: TABLE roles; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE roles IS 'Combined with table Users, this table defines a user''s privelages
';


--
-- Name: schema_version; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE schema_version (
    id integer NOT NULL,
    version character varying,
    modified_date timestamp with time zone DEFAULT now()
);


ALTER TABLE schema_version OWNER TO edge;

--
-- Name: TABLE schema_version; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE schema_version IS 'Store database schema version';


--
-- Name: schema_version_id_seq; Type: SEQUENCE; Schema: public; Owner: edge
--

CREATE SEQUENCE schema_version_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE schema_version_id_seq OWNER TO edge;

--
-- Name: schema_version_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: edge
--

ALTER SEQUENCE schema_version_id_seq OWNED BY schema_version.id;


--
-- Name: sms_configurations; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE sms_configurations (
    key character varying NOT NULL,
    value character varying NOT NULL,
    modified_date timestamp with time zone DEFAULT now()
);


ALTER TABLE sms_configurations OWNER TO edge;

--
-- Name: TABLE sms_configurations; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE sms_configurations IS 'This table is used to store SMS configuration as key/value pairs';


--
-- Name: sms_jobs; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE sms_jobs (
    sms_job_id integer NOT NULL,
    recipient character varying NOT NULL,
    message text,
    sent boolean DEFAULT false NOT NULL,
    failed boolean DEFAULT false NOT NULL,
    comments character varying,
    created_date timestamp with time zone NOT NULL,
    modified_date timestamp with time zone DEFAULT now()
);


ALTER TABLE sms_jobs OWNER TO edge;

--
-- Name: TABLE sms_jobs; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE sms_jobs IS 'This table is used to store queued SMS messages. Jobs within the queue will be handled by a worker thread which is responsible for handling any send failures and retrying failed jobs';


--
-- Name: sms_jobs_sms_job_id_seq; Type: SEQUENCE; Schema: public; Owner: edge
--

CREATE SEQUENCE sms_jobs_sms_job_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sms_jobs_sms_job_id_seq OWNER TO edge;

--
-- Name: sms_jobs_sms_job_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: edge
--

ALTER SEQUENCE sms_jobs_sms_job_id_seq OWNED BY sms_jobs.sms_job_id;


--
-- Name: status_codes; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE status_codes (
    status_code integer NOT NULL,
    description character varying(255),
    modified_date timestamp with time zone DEFAULT now()
);


ALTER TABLE status_codes OWNER TO edge;

--
-- Name: TABLE status_codes; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE status_codes IS 'Maps a job status number to a human readable format.

Values in the 20s are owned by the COntent-prep app

Values in the 30s are owned by the Content-send app';


--
-- Name: studies; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE studies (
    study_id integer NOT NULL,
    study_uid character varying(255) NOT NULL,
    exam_id integer NOT NULL,
    study_description character varying(255),
    study_date timestamp without time zone,
    modified_date timestamp with time zone DEFAULT now()
);


ALTER TABLE studies OWNER TO edge;

--
-- Name: TABLE studies; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE studies IS 'DICOM uid info for exams listed by accession in table Exams';


--
-- Name: studies_study_id_seq; Type: SEQUENCE; Schema: public; Owner: edge
--

CREATE SEQUENCE studies_study_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE studies_study_id_seq OWNER TO edge;

--
-- Name: studies_study_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: edge
--

ALTER SEQUENCE studies_study_id_seq OWNED BY studies.study_id;


--
-- Name: transactions; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE transactions (
    transaction_id integer NOT NULL,
    job_id integer NOT NULL,
    status_code integer NOT NULL,
    comments character varying,
    modified_date timestamp with time zone DEFAULT now()
);


ALTER TABLE transactions OWNER TO edge;

--
-- Name: TABLE transactions; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE transactions IS 'status logging/auditing for jobs defined in table Jobs. The java apps come here to determine their work by looking at the value status';


--
-- Name: COLUMN transactions.status_code; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON COLUMN transactions.status_code IS 'WHen a job is created by the GUI Token app, the row is created with value 1

Prepare Content looks for value of one and promites status to 2 on exit

Content transfer looks for status 2 and promotes to 3 on exit

 ';


--
-- Name: transactions_transaction_id_seq; Type: SEQUENCE; Schema: public; Owner: edge
--

CREATE SEQUENCE transactions_transaction_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE transactions_transaction_id_seq OWNER TO edge;

--
-- Name: transactions_transaction_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: edge
--

ALTER SEQUENCE transactions_transaction_id_seq OWNED BY transactions.transaction_id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: edge
--

CREATE TABLE users (
    user_id integer NOT NULL,
    user_login character varying(40) DEFAULT NULL::character varying,
    user_name character varying(100) DEFAULT ''::character varying,
    email character varying(100) DEFAULT NULL::character varying,
    crypted_password character varying(40) DEFAULT NULL::character varying,
    salt character varying(40) DEFAULT NULL::character varying,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    remember_token character varying(40) DEFAULT NULL::character varying,
    remember_token_expires_at timestamp with time zone,
    role_id integer NOT NULL,
    modified_date timestamp with time zone DEFAULT now(),
    active boolean DEFAULT true
);


ALTER TABLE users OWNER TO edge;

--
-- Name: TABLE users; Type: COMMENT; Schema: public; Owner: edge
--

COMMENT ON TABLE users IS 'Combined with table Roles, this table defines who can do what on the Edge appliacne Web GUI';


--
-- Name: users_user_id_seq; Type: SEQUENCE; Schema: public; Owner: edge
--

CREATE SEQUENCE users_user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE users_user_id_seq OWNER TO edge;

--
-- Name: users_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: edge
--

ALTER SEQUENCE users_user_id_seq OWNED BY users.user_id;


--
-- Name: v_consented; Type: VIEW; Schema: public; Owner: edge
--

CREATE VIEW v_consented AS
 SELECT patients.patient_id,
    patients.mrn,
    patients.patient_name,
    patients.dob,
    patients.sex,
    patients.street,
    patients.city,
    patients.state,
    patients.zip_code,
    patients.email_address,
    patients.rsna_id,
    patients.modified_date,
    patients.consent_timestamp
   FROM patients
  WHERE (patients.consent_timestamp IS NOT NULL);


ALTER TABLE v_consented OWNER TO edge;

--
-- Name: v_exam_status; Type: VIEW; Schema: public; Owner: edge
--

CREATE VIEW v_exam_status AS
 SELECT p.patient_id,
    p.mrn,
    p.patient_name,
    p.dob,
    p.sex,
    p.street,
    p.city,
    p.state,
    p.zip_code,
    p.email_address,
    e.exam_id,
    e.accession_number,
    e.exam_description,
    r.report_id,
    r.status,
    r.status_timestamp,
    r.report_text,
    r.dictator,
    r.transcriber,
    r.signer
   FROM ((patients p
     JOIN exams e ON ((p.patient_id = e.patient_id)))
     JOIN ( SELECT r1.report_id,
            r1.exam_id,
            r1.proc_code,
            r1.status,
            r1.status_timestamp,
            r1.report_text,
            r1.signer,
            r1.dictator,
            r1.transcriber,
            r1.modified_date
           FROM reports r1
          WHERE (r1.report_id = ( SELECT r2.report_id
                   FROM reports r2
                  WHERE (r2.exam_id = r1.exam_id)
                  ORDER BY r2.report_id DESC
                 LIMIT 1))) r ON ((e.exam_id = r.exam_id)));


ALTER TABLE v_exam_status OWNER TO edge;

--
-- Name: v_exams_sent; Type: VIEW; Schema: public; Owner: edge
--

CREATE VIEW v_exams_sent AS
 SELECT transactions.transaction_id,
    transactions.job_id,
    transactions.status_code,
    transactions.comments,
    transactions.modified_date
   FROM transactions
  WHERE (transactions.status_code = 40);


ALTER TABLE v_exams_sent OWNER TO edge;

--
-- Name: v_job_status; Type: VIEW; Schema: public; Owner: edge
--

CREATE VIEW v_job_status AS
 SELECT js.job_set_id,
    j.job_id,
    j.exam_id,
    js.delay_in_hrs,
    t.status,
    t.status_message,
    t.modified_date AS last_transaction_timestamp,
    js.single_use_patient_id,
    js.email_address,
    js.phone_number,
    t.comments,
    js.send_on_complete,
    js.access_code,
    j.remaining_retries,
    js.send_to_site
   FROM ((jobs j
     JOIN job_sets js ON ((j.job_set_id = js.job_set_id)))
     JOIN ( SELECT t1.job_id,
            t1.status_code AS status,
            sc.description AS status_message,
            t1.comments,
            t1.modified_date
           FROM (transactions t1
             JOIN status_codes sc ON ((t1.status_code = sc.status_code)))
          WHERE (t1.modified_date = ( SELECT max(t2.modified_date) AS max
                   FROM transactions t2
                  WHERE (t2.job_id = t1.job_id)))) t ON ((j.job_id = t.job_id)));


ALTER TABLE v_job_status OWNER TO edge;

--
-- Name: v_patients_sent; Type: VIEW; Schema: public; Owner: edge
--

CREATE VIEW v_patients_sent AS
 SELECT DISTINCT job_sets.patient_id
   FROM transactions,
    jobs,
    job_sets
  WHERE ((transactions.status_code = 40) AND (transactions.job_id = jobs.job_id) AND (jobs.job_set_id = job_sets.job_set_id));


ALTER TABLE v_patients_sent OWNER TO edge;

--
-- Name: device_id; Type: DEFAULT; Schema: public; Owner: edge
--

ALTER TABLE ONLY devices ALTER COLUMN device_id SET DEFAULT nextval('devices_device_id_seq'::regclass);


--
-- Name: email_job_id; Type: DEFAULT; Schema: public; Owner: edge
--

ALTER TABLE ONLY email_jobs ALTER COLUMN email_job_id SET DEFAULT nextval('email_jobs_email_job_id_seq'::regclass);


--
-- Name: exam_id; Type: DEFAULT; Schema: public; Owner: edge
--

ALTER TABLE ONLY exams ALTER COLUMN exam_id SET DEFAULT nextval('exams_exam_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: edge
--

ALTER TABLE ONLY hipaa_audit_accession_numbers ALTER COLUMN id SET DEFAULT nextval('hipaa_audit_accession_numbers_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: edge
--

ALTER TABLE ONLY hipaa_audit_mrns ALTER COLUMN id SET DEFAULT nextval('hipaa_audit_mrns_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: edge
--

ALTER TABLE ONLY hipaa_audit_views ALTER COLUMN id SET DEFAULT nextval('hipaa_audit_views_id_seq'::regclass);


--
-- Name: job_set_id; Type: DEFAULT; Schema: public; Owner: edge
--

ALTER TABLE ONLY job_sets ALTER COLUMN job_set_id SET DEFAULT nextval('job_sets_job_set_id_seq'::regclass);


--
-- Name: job_id; Type: DEFAULT; Schema: public; Owner: edge
--

ALTER TABLE ONLY jobs ALTER COLUMN job_id SET DEFAULT nextval('jobs_job_id_seq'::regclass);


--
-- Name: event_id; Type: DEFAULT; Schema: public; Owner: edge
--

ALTER TABLE ONLY patient_merge_events ALTER COLUMN event_id SET DEFAULT nextval('patient_merge_events_event_id_seq'::regclass);


--
-- Name: patient_id; Type: DEFAULT; Schema: public; Owner: edge
--

ALTER TABLE ONLY patients ALTER COLUMN patient_id SET DEFAULT nextval('patients_patient_id_seq'::regclass);


--
-- Name: report_id; Type: DEFAULT; Schema: public; Owner: edge
--

ALTER TABLE ONLY reports ALTER COLUMN report_id SET DEFAULT nextval('reports_report_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: edge
--

ALTER TABLE ONLY schema_version ALTER COLUMN id SET DEFAULT nextval('schema_version_id_seq'::regclass);


--
-- Name: sms_job_id; Type: DEFAULT; Schema: public; Owner: edge
--

ALTER TABLE ONLY sms_jobs ALTER COLUMN sms_job_id SET DEFAULT nextval('sms_jobs_sms_job_id_seq'::regclass);


--
-- Name: study_id; Type: DEFAULT; Schema: public; Owner: edge
--

ALTER TABLE ONLY studies ALTER COLUMN study_id SET DEFAULT nextval('studies_study_id_seq'::regclass);


--
-- Name: transaction_id; Type: DEFAULT; Schema: public; Owner: edge
--

ALTER TABLE ONLY transactions ALTER COLUMN transaction_id SET DEFAULT nextval('transactions_transaction_id_seq'::regclass);


--
-- Name: user_id; Type: DEFAULT; Schema: public; Owner: edge
--

ALTER TABLE ONLY users ALTER COLUMN user_id SET DEFAULT nextval('users_user_id_seq'::regclass);


--
-- Data for Name: configurations; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY configurations (key, value, modified_date) FROM stdin;
consent-expired-days	90	2012-03-13 15:56:06.768-05
fail-on-incomplete-study	false	2013-03-04 14:57:33.549-06
iti41-doc-endpoint-uri	https://64.28.70.198:8443/XDImgService/services/xdrreceiver	2017-11-03 13:17:12.478876-06
iti41-img-endpoint-uri	https://64.28.70.198:8443/XDImgService/services/xdrreceiver	2017-11-03 13:17:12.478876-06
iti41-socket-timeout	120	2011-04-05 12:10:46.354824-05
iti41-source-id	TBD	2017-11-03 13:17:12.478876-06
iti8-reg-uri	mllps://64.28.70.198:8444	2017-11-03 13:17:12.478876-06
iti9-pix-uri	mllps://64.28.70.198:8444	2017-11-03 13:17:12.478876-06
max-retries	10	2013-02-26 14:57:33.549-06
pdf-template	false	2017-04-03 13:04:01.177535-05
retrieve-timeout-in-secs	600	2013-03-04 14:57:33.549-06
retry-delay-in-mins	10	2013-02-26 14:57:33.549-06
rsna-assigning-authority	1.3.6.1.4.1.19376.3.840.1.1.1	2017-11-03 13:17:12.478876-06
scp-ae-title	RSNA-ISN	2011-01-10 18:35:16.668828-06
scp-idle-timeout	60000	2014-06-18 13:05:05.933-05
scp-max-receive-pdu-length	16364	2015-03-20 13:00:05.933-05
scp-max-send-pdu-length	16364	2015-03-20 13:00:05.933-05
scp-port	4104	2012-03-13 15:57:33.549-05
scp-release-timeout	5000	2012-03-13 15:57:33.549-05
scp-request-timeout	5000	2012-03-13 15:57:33.549-05
scu-ae-title	RSNA-ISN	2011-01-10 18:43:13.369949-06
search-patient-lastname	false	2014-02-21 12:05:05.933-06
secondary-capture-report-enabled	true	2014-02-21 12:05:05.933-06
site-assigning-authority	TBD	2011-11-10 13:00:00.000000-06
site_id	TBD	2015-03-31 17:35:16.668828-05
submit-stats	false	2014-10-16 15:58:33.549-05
\.


--
-- Data for Name: devices; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY devices (device_id, ae_title, host, port_number, modified_date) FROM stdin;
\.


--
-- Name: devices_device_id_seq; Type: SEQUENCE SET; Schema: public; Owner: edge
--

SELECT pg_catalog.setval('devices_device_id_seq', 1, true);


--
-- Data for Name: email_configurations; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY email_configurations (key, value, modified_date) FROM stdin;
mail.smtp.from		2015-02-13 12:05:05.933-06
mail.smtp.host		2015-02-13 12:05:05.933-06
enable_error_email	false	2014-02-21 12:05:05.933-06
enable_patient_email	false	2014-02-21 12:05:05.933-06
error_email_recipients		2014-02-21 12:05:05.933-06
username		2014-02-21 12:05:05.933-06
password		2014-02-21 12:05:05.933-06
bounce_email		2014-02-21 12:05:05.933-06
reply_to_email		2014-02-21 12:05:05.933-06
patient_email_subject	Your imaging records are ready for viewing	2017-04-03 12:17:43.758298-05
error_email_body	The following job failed to send to the clearinghouse:\r\n\r\nName: $patientname$\r\nAccession: $accession$\r\nJob ID: $jobid$\r\nStatus: $jobstatus$ ($jobstatuscode$)\r\nError Detail:\r\n\r\n$errormsg$	2017-04-03 12:17:43.758298-05
patient_email_body	<b><font size="24">RSNA Image Share</font></b><br><b><i><font size="5">Take Control of Your Medical Images</font></i></b></h2><br><br><br>Dear $patientname$,<br><br>Welcome to Image Share, a network designed to enable patients to access and control their medical imaging results. Image Share was developed by the Radiological Society of North America (RSNA) and its partners, with funding from the National Institute of Biomedical Imaging and Bioengineering.<br><br>You are receiving this message because the radiology staff at $site_id$ have sent your imaging results to a secure online data repository so you can access them.<br><br>To access your images:<br><br><ol><li><b>Create a personal health record (PHR) account on one of the participating image-enabled PHR systems.</b> If you created an account following an earlier visit, simply log in. You can create an account on one of the following participating sites:<br><br><ul><li>DICOM Grid:<a href="http://imageshare.dicomgrid.com">http://imageshare.dicomgrid.com</a></li><li>lifeIMAGE: <a href="https://cloud.lifeimage.com/rsna/phr">https://cloud.lifeimage.com/rsna/phr</a></li></ul><br><br>Each of these sites provides detailed instructions on creating an account and using it to retrieve your imaging results. Be careful to record your PHR account log in information (PHR provider, user name and password) and keep it secure, as you do with all your valuable online information.</li><br><br><li><b>Use your PHR account to access and take control of your imaging results.</b> Once you’ve created an account, you’ll just need to log in and provide two pieces of information to access your images and reports:<br><br>Your Access Code: <b>$accesscode$</b><br>Your date of birth<br><br>That’s all you need to retrieve the images and report into your PHR account. You can then use your PHR account to share information with others you trust, including care providers. They can view the images and the report anywhere Internet access is available. Some PHRs enable you to email a link to your images, so a provider can view your examinations without you needing to be present.</li></ol><br><br>User support is available during business hours at <a href="mailto:helpdesk@imsharing.org">helpdesk@imsharing.org</a> | Toll-free: 1-855-IM-SHARING (467-4274).	2017-04-03 12:19:09.864251-05
\.


--
-- Data for Name: email_jobs; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY email_jobs (email_job_id, recipient, subject, body, sent, failed, comments, created_date, modified_date) FROM stdin;
\.


--
-- Name: email_jobs_email_job_id_seq; Type: SEQUENCE SET; Schema: public; Owner: edge
--

SELECT pg_catalog.setval('email_jobs_email_job_id_seq', 1, false);


--
-- Data for Name: exams; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY exams (exam_id, accession_number, patient_id, exam_description, modified_date) FROM stdin;
\.


--
-- Name: exams_exam_id_seq; Type: SEQUENCE SET; Schema: public; Owner: edge
--

SELECT pg_catalog.setval('exams_exam_id_seq', 101, true);


--
-- Data for Name: hipaa_audit_accession_numbers; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY hipaa_audit_accession_numbers (id, view_id, accession_number, modified_date) FROM stdin;
\.


--
-- Name: hipaa_audit_accession_numbers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: edge
--

SELECT pg_catalog.setval('hipaa_audit_accession_numbers_id_seq', 388, true);


--
-- Data for Name: hipaa_audit_mrns; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY hipaa_audit_mrns (id, view_id, mrn, modified_date) FROM stdin;
\.


--
-- Name: hipaa_audit_mrns_id_seq; Type: SEQUENCE SET; Schema: public; Owner: edge
--

SELECT pg_catalog.setval('hipaa_audit_mrns_id_seq', 2220, true);


--
-- Data for Name: hipaa_audit_views; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY hipaa_audit_views (id, requesting_ip, requesting_username, requesting_uri, modified_date) FROM stdin;
\.


--
-- Name: hipaa_audit_views_id_seq; Type: SEQUENCE SET; Schema: public; Owner: edge
--

SELECT pg_catalog.setval('hipaa_audit_views_id_seq', 1662, true);


--
-- Data for Name: job_sets; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY job_sets (job_set_id, patient_id, user_id, email_address, modified_date, delay_in_hrs, single_use_patient_id, send_on_complete, access_code, send_to_site, phone_number, global_id) FROM stdin;
\.


--
-- Name: job_sets_job_set_id_seq; Type: SEQUENCE SET; Schema: public; Owner: edge
--

SELECT pg_catalog.setval('job_sets_job_set_id_seq', 112, true);


--
-- Data for Name: jobs; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY jobs (job_id, job_set_id, exam_id, report_id, document_id, remaining_retries, modified_date) FROM stdin;
\.


--
-- Name: jobs_job_id_seq; Type: SEQUENCE SET; Schema: public; Owner: edge
--

SELECT pg_catalog.setval('jobs_job_id_seq', 114, true);


--
-- Data for Name: patient_merge_events; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY patient_merge_events (event_id, old_mrn, new_mrn, old_patient_id, new_patient_id, status, modified_date) FROM stdin;
\.


--
-- Name: patient_merge_events_event_id_seq; Type: SEQUENCE SET; Schema: public; Owner: edge
--

SELECT pg_catalog.setval('patient_merge_events_event_id_seq', 1, false);


--
-- Data for Name: patients; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY patients (patient_id, mrn, patient_name, dob, sex, street, city, state, zip_code, email_address, rsna_id, modified_date, consent_timestamp, autosend) FROM stdin;
\.


--
-- Name: patients_patient_id_seq; Type: SEQUENCE SET; Schema: public; Owner: edge
--

SELECT pg_catalog.setval('patients_patient_id_seq', 90, true);


--
-- Data for Name: reports; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY reports (report_id, exam_id, proc_code, status, status_timestamp, report_text, signer, dictator, transcriber, modified_date) FROM stdin;
\.


--
-- Name: reports_report_id_seq; Type: SEQUENCE SET; Schema: public; Owner: edge
--

SELECT pg_catalog.setval('reports_report_id_seq', 186, true);


--
-- Data for Name: roles; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY roles (role_id, role_description, modified_date) FROM stdin;
\.


--
-- Data for Name: schema_version; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY schema_version (id, version, modified_date) FROM stdin;
0	5.0.0	2017-04-03 12:19:09.864251-05
\.


--
-- Name: schema_version_id_seq; Type: SEQUENCE SET; Schema: public; Owner: edge
--

SELECT pg_catalog.setval('schema_version_id_seq', 1, false);


--
-- Data for Name: sms_configurations; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY sms_configurations (key, value, modified_date) FROM stdin;
enable_sms	false	2017-04-03 12:19:09.864251-05
account_id		2017-04-03 12:19:09.864251-05
token		2017-04-03 12:19:09.864251-05
sender		2017-04-03 12:19:09.864251-05
body	Your imaging results are ready to be accessed. Your Access Code is $accesscode$. Instructions available at http://www.rsna.org/image_share.aspx.	2017-04-03 12:19:09.864251-05
proxy_host	192.203.117.40	2017-04-03 12:19:09.864251-05
proxy_port	3128	2017-04-03 12:19:09.864251-05
proxy_set	false	2017-04-03 12:19:09.864251-05
\.


--
-- Data for Name: sms_jobs; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY sms_jobs (sms_job_id, recipient, message, sent, failed, comments, created_date, modified_date) FROM stdin;
\.


--
-- Name: sms_jobs_sms_job_id_seq; Type: SEQUENCE SET; Schema: public; Owner: edge
--

SELECT pg_catalog.setval('sms_jobs_sms_job_id_seq', 1, false);


--
-- Data for Name: status_codes; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY status_codes (status_code, description, modified_date) FROM stdin;
31	Started processing input data	2010-10-22 09:58:07.496506-05
21	Waiting for report finalization	2010-10-22 11:59:15.243445-05
23	Started DICOM C-MOVE	2010-10-22 11:59:15.243445-05
30	Waiting to start transfer to clearinghouse	2010-10-22 11:59:15.243445-05
22	Waiting for job delay to expire	2010-10-22 11:59:15.243445-05
24	Waiting for exam completion	2013-02-26 14:57:33.549-06
34	Started document submission	2010-10-22 09:58:07.496506-05
40	Completed transfer to clearinghouse	2010-10-22 09:58:07.496506-05
1	Waiting to retrieve images and report	2010-10-22 09:58:07.496506-05
-23	DICOM C-MOVE failed	2010-10-22 11:59:15.243445-05
-21	Unable to find images	2010-10-22 11:59:15.243445-05
-30	Failed to transfer to clearinghouse	2010-11-02 09:39:24.901369-05
-20	Failed to prepare content	2010-10-22 09:58:07.496506-05
-34	Failed to submit documents to clearinghouse	2010-11-02 09:40:28.488821-05
-24	Exam has been canceled	2014-09-03 15:41:37.99-05
-1	No devices found	2017-04-03 12:19:09.864251-05
32	Started patient registration	2017-04-03 12:19:09.864251-05
33	Retrieving global id	2017-04-03 12:19:09.864251-05
-32	Failed to register patient with clearinghouse	2017-04-03 12:19:09.864251-05
-33	Failed to retrieve global id	2017-04-03 12:19:09.864251-05
\.


--
-- Data for Name: studies; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY studies (study_id, study_uid, exam_id, study_description, study_date, modified_date) FROM stdin;
\.


--
-- Name: studies_study_id_seq; Type: SEQUENCE SET; Schema: public; Owner: edge
--

SELECT pg_catalog.setval('studies_study_id_seq', 236, true);


--
-- Data for Name: transactions; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY transactions (transaction_id, job_id, status_code, comments, modified_date) FROM stdin;
\.


--
-- Name: transactions_transaction_id_seq; Type: SEQUENCE SET; Schema: public; Owner: edge
--

SELECT pg_catalog.setval('transactions_transaction_id_seq', 16078, true);


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: edge
--

COPY users (user_id, user_login, user_name, email, crypted_password, salt, created_at, updated_at, remember_token, remember_token_expires_at, role_id, modified_date, active) FROM stdin;
10	AUTOSEND	System AutoSend	\N	\N	\N	\N	\N	\N	\N	0	2017-04-03 12:19:09.864251-05	t
\.


--
-- Name: users_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: edge
--

SELECT pg_catalog.setval('users_user_id_seq', 10, true);


--
-- Name: pk_device_id; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY devices
    ADD CONSTRAINT pk_device_id PRIMARY KEY (device_id);


--
-- Name: pk_email_configuration_key; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY email_configurations
    ADD CONSTRAINT pk_email_configuration_key PRIMARY KEY (key);


--
-- Name: pk_email_job_id; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY email_jobs
    ADD CONSTRAINT pk_email_job_id PRIMARY KEY (email_job_id);


--
-- Name: pk_event_id; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY patient_merge_events
    ADD CONSTRAINT pk_event_id PRIMARY KEY (event_id);


--
-- Name: pk_exam_id; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY exams
    ADD CONSTRAINT pk_exam_id PRIMARY KEY (exam_id);


--
-- Name: pk_hipaa_audit_accession_number_id; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY hipaa_audit_accession_numbers
    ADD CONSTRAINT pk_hipaa_audit_accession_number_id PRIMARY KEY (id);


--
-- Name: pk_hipaa_audit_mrn_id; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY hipaa_audit_mrns
    ADD CONSTRAINT pk_hipaa_audit_mrn_id PRIMARY KEY (id);


--
-- Name: pk_hipaa_audit_view_id; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY hipaa_audit_views
    ADD CONSTRAINT pk_hipaa_audit_view_id PRIMARY KEY (id);


--
-- Name: pk_id; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY schema_version
    ADD CONSTRAINT pk_id PRIMARY KEY (id);


--
-- Name: pk_job_id; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY jobs
    ADD CONSTRAINT pk_job_id PRIMARY KEY (job_id);


--
-- Name: pk_job_set_id; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY job_sets
    ADD CONSTRAINT pk_job_set_id PRIMARY KEY (job_set_id);


--
-- Name: pk_key; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY configurations
    ADD CONSTRAINT pk_key PRIMARY KEY (key);


--
-- Name: pk_patient_id; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY patients
    ADD CONSTRAINT pk_patient_id PRIMARY KEY (patient_id);


--
-- Name: pk_report_id; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY reports
    ADD CONSTRAINT pk_report_id PRIMARY KEY (report_id);


--
-- Name: pk_role_id; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY roles
    ADD CONSTRAINT pk_role_id PRIMARY KEY (role_id);


--
-- Name: pk_sms_configuration_key; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY sms_configurations
    ADD CONSTRAINT pk_sms_configuration_key PRIMARY KEY (key);


--
-- Name: pk_sms_job_id; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY sms_jobs
    ADD CONSTRAINT pk_sms_job_id PRIMARY KEY (sms_job_id);


--
-- Name: pk_status_code; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY status_codes
    ADD CONSTRAINT pk_status_code PRIMARY KEY (status_code);


--
-- Name: pk_study_id; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY studies
    ADD CONSTRAINT pk_study_id PRIMARY KEY (study_id);


--
-- Name: pk_transaction_id; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY transactions
    ADD CONSTRAINT pk_transaction_id PRIMARY KEY (transaction_id);


--
-- Name: pk_user_id; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY users
    ADD CONSTRAINT pk_user_id PRIMARY KEY (user_id);


--
-- Name: uq_exam; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY exams
    ADD CONSTRAINT uq_exam UNIQUE (accession_number, patient_id);


--
-- Name: uq_login; Type: CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY users
    ADD CONSTRAINT uq_login UNIQUE (user_login);


--
-- Name: exams_accession_number_idx; Type: INDEX; Schema: public; Owner: edge
--

CREATE INDEX exams_accession_number_idx ON exams USING btree (accession_number);


--
-- Name: jobs_job_set_id; Type: INDEX; Schema: public; Owner: edge
--

CREATE INDEX jobs_job_set_id ON jobs USING btree (job_set_id);


--
-- Name: patient_search_idx; Type: INDEX; Schema: public; Owner: edge
--

CREATE INDEX patient_search_idx ON patients USING gin (to_tsvector('simple'::regconfig, ((((((ltrim((mrn)::text, '0'::text) || ' '::text) || (COALESCE(patient_name, ''::character varying))::text) || ' '::text) || COALESCE((date_part('year'::text, dob))::text, ''::text)) || ' '::text) || (COALESCE(email_address, ''::character varying))::text)));


--
-- Name: patients_dob_idx; Type: INDEX; Schema: public; Owner: edge
--

CREATE INDEX patients_dob_idx ON patients USING btree (dob);


--
-- Name: patients_mrn_ix; Type: INDEX; Schema: public; Owner: edge
--

CREATE UNIQUE INDEX patients_mrn_ix ON patients USING btree (mrn);


--
-- Name: patients_patient_name_idx; Type: INDEX; Schema: public; Owner: edge
--

CREATE INDEX patients_patient_name_idx ON patients USING btree (patient_name);


--
-- Name: reports_status_timestamp_idx; Type: INDEX; Schema: public; Owner: edge
--

CREATE INDEX reports_status_timestamp_idx ON reports USING btree (status_timestamp);


--
-- Name: reports_unique_status_idx; Type: INDEX; Schema: public; Owner: edge
--

CREATE UNIQUE INDEX reports_unique_status_idx ON reports USING btree (exam_id, status, status_timestamp);


--
-- Name: transactions_job_id; Type: INDEX; Schema: public; Owner: edge
--

CREATE INDEX transactions_job_id ON transactions USING btree (job_id);


--
-- Name: transactions_modified_date; Type: INDEX; Schema: public; Owner: edge
--

CREATE INDEX transactions_modified_date ON transactions USING btree (modified_date);


--
-- Name: transactions_status_code_idx; Type: INDEX; Schema: public; Owner: edge
--

CREATE INDEX transactions_status_code_idx ON transactions USING btree (status_code);


--
-- Name: trigger_exam_autosend; Type: TRIGGER; Schema: public; Owner: edge
--

CREATE TRIGGER trigger_exam_autosend AFTER INSERT ON exams FOR EACH ROW EXECUTE PROCEDURE fn_exam_autosend();


--
-- Name: fk_exam_id; Type: FK CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY jobs
    ADD CONSTRAINT fk_exam_id FOREIGN KEY (exam_id) REFERENCES exams(exam_id);


--
-- Name: fk_exam_id; Type: FK CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY studies
    ADD CONSTRAINT fk_exam_id FOREIGN KEY (exam_id) REFERENCES exams(exam_id);


--
-- Name: fk_exam_id; Type: FK CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY reports
    ADD CONSTRAINT fk_exam_id FOREIGN KEY (exam_id) REFERENCES exams(exam_id);


--
-- Name: fk_job_id; Type: FK CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY transactions
    ADD CONSTRAINT fk_job_id FOREIGN KEY (job_id) REFERENCES jobs(job_id);


--
-- Name: fk_job_set_id; Type: FK CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY jobs
    ADD CONSTRAINT fk_job_set_id FOREIGN KEY (job_set_id) REFERENCES job_sets(job_set_id);


--
-- Name: fk_patient_id; Type: FK CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY job_sets
    ADD CONSTRAINT fk_patient_id FOREIGN KEY (patient_id) REFERENCES patients(patient_id);


--
-- Name: fk_patient_id; Type: FK CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY exams
    ADD CONSTRAINT fk_patient_id FOREIGN KEY (patient_id) REFERENCES patients(patient_id);


--
-- Name: fk_report_id; Type: FK CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY jobs
    ADD CONSTRAINT fk_report_id FOREIGN KEY (report_id) REFERENCES reports(report_id);


--
-- Name: fk_status_code; Type: FK CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY transactions
    ADD CONSTRAINT fk_status_code FOREIGN KEY (status_code) REFERENCES status_codes(status_code);


--
-- Name: fk_user_id; Type: FK CONSTRAINT; Schema: public; Owner: edge
--

ALTER TABLE ONLY job_sets
    ADD CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(user_id);


--
-- Name: public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

