-- Downloaded from: https://github.com/HalfCoke/blog_img/blob/c619e88b64e4f2ac7aa15975f672fa7cabe789b6/img/init_0218.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 16.1
-- Dumped by pg_dump version 16.1

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
-- Name: ssodb; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA ssodb;


ALTER SCHEMA ssodb OWNER TO postgres;

--
-- Name: add_partition_for_table(text, timestamp without time zone, timestamp without time zone); Type: FUNCTION; Schema: ssodb; Owner: ssouser
--

CREATE FUNCTION ssodb.add_partition_for_table(_table_name text, _start_date timestamp without time zone, _date timestamp without time zone) RETURNS void
    LANGUAGE plpgsql
    AS $$
declare
    _partition_name text;
begin
    -- 为分区表创建分区
    _partition_name = concat(_table_name, '_', to_char(_start_date, 'YYYYMMDD'));
    execute 'create table ' || _partition_name || ' partition of ' || _table_name || ' for values from (''' ||
            _start_date || ''') to (''' || _date || ''')';
end;
$$;


ALTER FUNCTION ssodb.add_partition_for_table(_table_name text, _start_date timestamp without time zone, _date timestamp without time zone) OWNER TO ssouser;

--
-- Name: add_partition_for_table_for_now_init(text); Type: FUNCTION; Schema: ssodb; Owner: ssouser
--

CREATE FUNCTION ssodb.add_partition_for_table_for_now_init(_table_name text) RETURNS void
    LANGUAGE plpgsql
    AS $$
begin
    -- 初始化历史分区，环境部署时，需要执行
    perform add_partition_for_table(_table_name, ('2023-01-01'), current_date);
    perform add_partition_for_table(_table_name, current_date, current_date + interval '7 day');
end;
$$;


ALTER FUNCTION ssodb.add_partition_for_table_for_now_init(_table_name text) OWNER TO ssouser;

--
-- Name: add_partition_for_table_if_not_exist_week(text); Type: FUNCTION; Schema: ssodb; Owner: ssouser
--

CREATE FUNCTION ssodb.add_partition_for_table_if_not_exist_week(_table_name text) RETURNS void
    LANGUAGE plpgsql
    AS $$
declare
    _last_date text;
begin
    -- 获取上一个分区开始日期
    select split_part(c.relname, '_', -1)
    into _last_date
    from pg_class c
             join pg_inherits i on i.inhrelid = c.oid
             join pg_class d on d.oid = i.inhparent
    where d.relname = _table_name
    order by c.relname desc
    limit 1;
    -- 构建下一周的分区
    if to_date(_last_date, 'YYYYMMDD') - current_date < 0  then
        perform add_partition_for_table(_table_name, to_date(_last_date, 'YYYYMMDD'),
                                        to_date(_last_date, 'YYYYMMDD') + interval '7 day');
    end if;
end;
$$;


ALTER FUNCTION ssodb.add_partition_for_table_if_not_exist_week(_table_name text) OWNER TO ssouser;

--
-- Name: any_like(jsonb, text); Type: FUNCTION; Schema: ssodb; Owner: ssouser
--

CREATE FUNCTION ssodb.any_like(array_value jsonb, target_value text) RETURNS boolean
    LANGUAGE sql
    AS $$
    select exists(select from jsonb_array_elements_text(array_value) where value like target_value);
$$;


ALTER FUNCTION ssodb.any_like(array_value jsonb, target_value text) OWNER TO ssouser;

--
-- Name: any_like(text, text); Type: FUNCTION; Schema: ssodb; Owner: ssouser
--

CREATE FUNCTION ssodb.any_like(single_value text, target_value text) RETURNS boolean
    LANGUAGE sql
    AS $$
select single_value like target_value;
$$;


ALTER FUNCTION ssodb.any_like(single_value text, target_value text) OWNER TO ssouser;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: t_account; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_account (
    id integer NOT NULL,
    username character varying(255),
    password character varying(255),
    password_last_modified_date timestamp without time zone,
    userid integer,
    email character varying(255),
    phone character varying(40),
    is_locked text,
    last_login_date timestamp without time zone,
    mfa character varying(1024),
    created_by character varying(255),
    created_date timestamp without time zone,
    last_modified_by character varying(255),
    last_modified_date timestamp without time zone,
    password_status character varying(40),
    recent_address character varying(255)
);


ALTER TABLE ssodb.t_account OWNER TO ssouser;

--
-- Name: t_system_property; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_system_property (
    id integer NOT NULL,
    key character varying(255),
    value character varying(255),
    status integer,
    created_date timestamp without time zone
);


ALTER TABLE ssodb.t_system_property OWNER TO ssouser;

--
-- Name: t_user; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_user (
    id integer NOT NULL,
    schoolid character varying(255),
    name character varying(255),
    ic_type character varying(40),
    ic_number character varying(255),
    department jsonb,
    type character varying(40),
    static_tag jsonb,
    dynamic_tag jsonb,
    roles character varying(1024),
    sex character varying(40),
    political character varying(40),
    candidate_number character varying(255),
    admission_number character varying(255),
    start_date timestamp without time zone,
    end_date timestamp without time zone,
    status character varying(40),
    is_registered character varying(255),
    is_disabled text,
    defer_date timestamp without time zone,
    modify_comment character varying(255),
    last_modified_by character varying(255),
    last_modified_date timestamp without time zone,
    is_face_disabled text
);


ALTER TABLE ssodb.t_user OWNER TO ssouser;

--
-- Name: COLUMN t_user.schoolid; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user.schoolid IS '学工号';


--
-- Name: COLUMN t_user.name; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user.name IS '姓名';


--
-- Name: COLUMN t_user.ic_type; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user.ic_type IS '身份证件类型';


--
-- Name: COLUMN t_user.ic_number; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user.ic_number IS '身份证件号';


--
-- Name: COLUMN t_user.department; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user.department IS '所在单位';


--
-- Name: COLUMN t_user.type; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user.type IS '人员类型';


--
-- Name: COLUMN t_user.static_tag; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user.static_tag IS '静态标签';


--
-- Name: COLUMN t_user.dynamic_tag; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user.dynamic_tag IS '动态标签';


--
-- Name: COLUMN t_user.sex; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user.sex IS '性别';


--
-- Name: COLUMN t_user.political; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user.political IS '政治面貌';


--
-- Name: COLUMN t_user.admission_number; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user.admission_number IS '录取通知书号';


--
-- Name: COLUMN t_user.is_registered; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user.is_registered IS '是否注册，如果已注册则填入账号username';


--
-- Name: COLUMN t_user.defer_date; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user.defer_date IS '延期时间';


--
-- Name: i_user_account_status; Type: VIEW; Schema: ssodb; Owner: ssouser
--

CREATE VIEW ssodb.i_user_account_status AS
 SELECT tu.id AS userid,
    ta.id AS accountid,
        CASE
            WHEN (ta.id IS NULL) THEN NULL::text
            WHEN ((ta.password_last_modified_date + '10 years'::interval) < now()) THEN '2'::text
            WHEN ((passwdttl.value)::integer = 0) THEN '0'::text
            WHEN ((ta.password_last_modified_date + (((passwdttl.value)::integer)::double precision * '1 day'::interval)) < now()) THEN '1'::text
            ELSE '0'::text
        END AS is_expired,
        CASE
            WHEN (ta.id IS NULL) THEN NULL::text
            WHEN ((ta.is_locked = '0'::text) AND ((lockunuse.value)::integer <> 0) AND ((ta.last_login_date + (((lockunuse.value)::integer)::double precision * '1 day'::interval)) < now())) THEN '1'::text
            WHEN (ta.is_locked = '1'::text) THEN '1'::text
            WHEN (ta.is_locked = '2'::text) THEN '2'::text
            ELSE '0'::text
        END AS is_locked,
        CASE
            WHEN (tu.id IS NULL) THEN NULL::text
            WHEN (tu.is_disabled = '0'::text) THEN '0'::text
            WHEN ((tu.is_disabled = '1'::text) AND (tu.defer_date IS NOT NULL) AND ((tu.defer_date + '1 day'::interval) > now())) THEN '0'::text
            ELSE '1'::text
        END AS is_disabled
   FROM (((ssodb.t_user tu
     LEFT JOIN ssodb.t_account ta ON ((tu.id = ta.userid)))
     LEFT JOIN ssodb.t_system_property passwdttl ON (((passwdttl.status = 1) AND ((passwdttl.key)::text = 'password_modify_cycle'::text))))
     LEFT JOIN ssodb.t_system_property lockunuse ON (((lockunuse.status = 1) AND ((lockunuse.key)::text = 'account_lock_unuse'::text))));


ALTER VIEW ssodb.i_user_account_status OWNER TO ssouser;

--
-- Name: v_user_attributes; Type: VIEW; Schema: ssodb; Owner: ssouser
--

CREATE VIEW ssodb.v_user_attributes AS
 SELECT ta.username,
    ta.password,
    tu.schoolid,
    tu.schoolid AS employee_number,
    tu.schoolid AS uid,
    tu.candidate_number,
    tu.admission_number,
    vuas.is_expired,
    vuas.is_locked,
    vuas.is_disabled,
    tu.name,
    tu.ic_type,
    tu.ic_number,
    tu.department,
    tu.type,
    tu.static_tag AS tag,
    ta.mfa,
    tu.sex,
    tu.political,
    ta.phone,
    ta.email,
    tu.schoolid AS edu_person_principal_name,
    tu.type AS edu_person_scoped_affiliation,
    tu.status
   FROM ((ssodb.t_user tu
     LEFT JOIN ssodb.t_account ta ON ((tu.id = ta.userid)))
     LEFT JOIN ssodb.i_user_account_status vuas ON ((tu.id = vuas.userid)));


ALTER VIEW ssodb.v_user_attributes OWNER TO ssouser;

--
-- Name: COLUMN v_user_attributes.schoolid; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.v_user_attributes.schoolid IS '学工号';


--
-- Name: COLUMN v_user_attributes.admission_number; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.v_user_attributes.admission_number IS '录取通知书号';


--
-- Name: COLUMN v_user_attributes.name; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.v_user_attributes.name IS '姓名';


--
-- Name: COLUMN v_user_attributes.ic_type; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.v_user_attributes.ic_type IS '身份证件类型';


--
-- Name: COLUMN v_user_attributes.ic_number; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.v_user_attributes.ic_number IS '身份证件号';


--
-- Name: COLUMN v_user_attributes.department; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.v_user_attributes.department IS '所在单位';


--
-- Name: COLUMN v_user_attributes.type; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.v_user_attributes.type IS '人员类型';


--
-- Name: COLUMN v_user_attributes.sex; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.v_user_attributes.sex IS '性别';


--
-- Name: COLUMN v_user_attributes.political; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.v_user_attributes.political IS '政治面貌';


--
-- Name: get_attributes(text); Type: FUNCTION; Schema: ssodb; Owner: ssouser
--

CREATE FUNCTION ssodb.get_attributes(in_uid text) RETURNS SETOF ssodb.v_user_attributes
    LANGUAGE sql
    AS $$
    select * from ssodb.v_user_attributes where schoolid = in_uid
    union
    select * from ssodb.v_user_attributes where username = in_uid
$$;


ALTER FUNCTION ssodb.get_attributes(in_uid text) OWNER TO ssouser;

--
-- Name: get_partition(text, text, text, text); Type: FUNCTION; Schema: ssodb; Owner: ssouser
--

CREATE FUNCTION ssodb.get_partition(schemaname text, tablename text, _start_date text, _end_date text) RETURNS SETOF text
    LANGUAGE plpgsql
    AS $$
declare
    firstPart text;
    lastPart  text;
    sqlString text;
begin
    -- 如果没有分区，返回null
    if ((select count(1)
         from pg_class c
                  join pg_inherits i on i.inhrelid = c.oid
                  join pg_class d on d.oid = i.inhparent
                  join pg_namespace n on c.relnamespace = n.oid
         where d.relname = tableName
           and n.nspname = schemaName) = 0) then
        return;
    end if;
    select c.relname
    into firstPart
    from pg_class c
             join pg_inherits i on i.inhrelid = c.oid
             join pg_class d on d.oid = i.inhparent
             join pg_namespace n on c.relnamespace = n.oid
    where d.relname = tableName
      and n.nspname = schemaName
      and c.relname <= concat(tableName, '_', _start_date)
    order by c.relname desc
    limit 1;
    select c.relname
    into lastPart
    from pg_class c
             join pg_inherits i on i.inhrelid = c.oid
             join pg_class d on d.oid = i.inhparent
             join pg_namespace n on c.relnamespace = n.oid
    where d.relname = tableName
      and n.nspname = schemaName
      and c.relname <= concat(tableName, '_', _end_date)
    order by c.relname desc
    limit 1;


    sqlString = 'select c.relname::text from pg_class c
                     join pg_inherits i on i.inhrelid = c.oid
                     join pg_class d on d.oid = i.inhparent
                     join pg_namespace n on c.relnamespace = n.oid
                 where d.relname = ''' || tableName || ''' and n.nspname = ''' || schemaName || '''';

    if (firstPart is not null) then
        sqlString = sqlString || ' and c.relname >= ''' || firstPart || '''';
    end if;
    if (lastPart is not null) then
        sqlString = sqlString || ' and c.relname <= ''' || lastPart || '''';
    end if;
    return query execute sqlString;
    return ;
end;
$$;


ALTER FUNCTION ssodb.get_partition(schemaname text, tablename text, _start_date text, _end_date text) OWNER TO ssouser;

--
-- Name: t_account_totp; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_account_totp (
    id integer NOT NULL,
    registration_date timestamp without time zone,
    secret_key character varying(2048),
    accountid integer,
    validation_code integer
);


ALTER TABLE ssodb.t_account_totp OWNER TO ssouser;

--
-- Name: google_authenticator_registration_record; Type: VIEW; Schema: ssodb; Owner: ssouser
--

CREATE VIEW ssodb.google_authenticator_registration_record AS
 SELECT tat.id,
    tat.registration_date,
    ta.last_login_date AS last_used_date_time,
    tat.secret_key,
    tu.schoolid AS username,
    tu.schoolid AS name,
    tat.validation_code
   FROM ((ssodb.t_account_totp tat
     LEFT JOIN ssodb.t_account ta ON ((tat.accountid = ta.id)))
     LEFT JOIN ssodb.t_user tu ON ((ta.userid = tu.id)));


ALTER VIEW ssodb.google_authenticator_registration_record OWNER TO ssouser;

--
-- Name: t_account_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_account_id_seq
    AS integer
    START WITH 100000
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_account_id_seq OWNER TO ssouser;

--
-- Name: t_account_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_account_id_seq OWNED BY ssodb.t_account.id;


--
-- Name: t_account_totp_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_account_totp_id_seq
    AS integer
    START WITH 100000
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_account_totp_id_seq OWNER TO ssouser;

--
-- Name: t_account_totp_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_account_totp_id_seq OWNED BY ssodb.t_account_totp.id;


--
-- Name: t_admin; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_admin (
    id integer NOT NULL,
    account character varying(255),
    password character varying(255),
    userid integer,
    created_date timestamp without time zone,
    created_by character varying(255),
    last_modified_date timestamp without time zone,
    last_modified_by character varying(255)
);


ALTER TABLE ssodb.t_admin OWNER TO ssouser;

--
-- Name: t_admin_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_admin_id_seq
    AS integer
    START WITH 100000
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_admin_id_seq OWNER TO ssouser;

--
-- Name: t_admin_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_admin_id_seq OWNED BY ssodb.t_admin.id;


--
-- Name: t_app; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_app (
    id integer NOT NULL,
    name character varying(255),
    description text,
    is_disabled smallint,
    appid character varying(255),
    app_secret character varying(255),
    "authorization" json,
    security_access_ip character varying(255),
    created_by character varying(255),
    created_date timestamp without time zone,
    last_modified_by character varying(255),
    last_modified_date timestamp without time zone
);


ALTER TABLE ssodb.t_app OWNER TO ssouser;

--
-- Name: t_app_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_app_id_seq
    AS integer
    START WITH 100000
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_app_id_seq OWNER TO ssouser;

--
-- Name: t_app_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_app_id_seq OWNED BY ssodb.t_app.id;


--
-- Name: t_app_token; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_app_token (
    id integer NOT NULL,
    appid character varying(255),
    token character varying(128),
    created_date timestamp without time zone
);


ALTER TABLE ssodb.t_app_token OWNER TO ssouser;

--
-- Name: t_app_token_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_app_token_id_seq
    AS integer
    START WITH 100000
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_app_token_id_seq OWNER TO ssouser;

--
-- Name: t_app_token_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_app_token_id_seq OWNED BY ssodb.t_app_token.id;


--
-- Name: t_attribute; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_attribute (
    id integer NOT NULL,
    name_en character varying(255),
    name_zh character varying(255),
    type character varying(255),
    category character varying(255),
    is_editabled integer,
    is_cas_attribute integer,
    is_ldap_attribute integer,
    is_filter_operation integer,
    is_index integer,
    created_by character varying(255),
    created_date timestamp without time zone,
    last_modified_by character varying(255),
    last_modified_date timestamp without time zone,
    is_desensitization integer,
    desensitization_mode character varying(255)
);


ALTER TABLE ssodb.t_attribute OWNER TO ssouser;

--
-- Name: t_attribute_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_attribute_id_seq
    AS integer
    START WITH 100000
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_attribute_id_seq OWNER TO ssouser;

--
-- Name: t_attribute_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_attribute_id_seq OWNED BY ssodb.t_attribute.id;


--
-- Name: t_code; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_code (
    id integer NOT NULL,
    key character varying(40),
    value character varying(255),
    category character varying(40),
    status smallint,
    created_by character varying(255),
    created_date timestamp without time zone
);


ALTER TABLE ssodb.t_code OWNER TO ssouser;

--
-- Name: t_code_department; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_code_department (
    id character varying(40) NOT NULL,
    name character varying(255),
    parent_id character varying(40),
    level smallint,
    is_disabled smallint,
    priority smallint,
    path character varying(255)
);


ALTER TABLE ssodb.t_code_department OWNER TO ssouser;

--
-- Name: t_code_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_code_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_code_id_seq OWNER TO ssouser;

--
-- Name: t_code_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_code_id_seq OWNED BY ssodb.t_code.id;


--
-- Name: t_code_userrole; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_code_userrole (
    id integer NOT NULL,
    name character varying(255),
    description character varying(255),
    "authorization" json
);


ALTER TABLE ssodb.t_code_userrole OWNER TO ssouser;

--
-- Name: t_code_userrole_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_code_userrole_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_code_userrole_id_seq OWNER TO ssouser;

--
-- Name: t_code_userrole_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_code_userrole_id_seq OWNED BY ssodb.t_code_userrole.id;


--
-- Name: t_code_usertype; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_code_usertype (
    id character varying(40) NOT NULL,
    name character varying(255),
    parent_id character varying(40),
    level smallint,
    is_disabled smallint,
    priority smallint,
    path character varying(255)
);


ALTER TABLE ssodb.t_code_usertype OWNER TO ssouser;

--
-- Name: t_export_job; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_export_job (
    id integer NOT NULL,
    created_by character varying(255),
    created_status smallint,
    start_date timestamp without time zone,
    end_date timestamp without time zone,
    created_comment character varying(255),
    download_count integer,
    download_path character varying(255),
    download_date timestamp without time zone,
    download_by character varying(255),
    remove_by character varying(255),
    remove_date timestamp without time zone,
    remove_status smallint,
    execute_param json
);


ALTER TABLE ssodb.t_export_job OWNER TO ssouser;

--
-- Name: t_export_job_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_export_job_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_export_job_id_seq OWNER TO ssouser;

--
-- Name: t_export_job_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_export_job_id_seq OWNED BY ssodb.t_export_job.id;


--
-- Name: t_file_baseinfo; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_file_baseinfo (
    id integer NOT NULL,
    file_path character varying(255),
    file_name character varying(255),
    file_length bigint,
    created_date timestamp without time zone
);


ALTER TABLE ssodb.t_file_baseinfo OWNER TO ssouser;

--
-- Name: t_file_baseinfo_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_file_baseinfo_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_file_baseinfo_id_seq OWNER TO ssouser;

--
-- Name: t_file_baseinfo_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_file_baseinfo_id_seq OWNED BY ssodb.t_file_baseinfo.id;


--
-- Name: t_import_backup; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_import_backup (
    backup_id integer,
    backup_type character varying(255),
    backup_data text,
    created_by character varying(255),
    created_date timestamp without time zone
);


ALTER TABLE ssodb.t_import_backup OWNER TO ssouser;

--
-- Name: TABLE t_import_backup; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON TABLE ssodb.t_import_backup IS 't_import_backup';


--
-- Name: t_import_job; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_import_job (
    id integer NOT NULL,
    operate_by character varying(255),
    operate_path character varying(255),
    operate_status smallint,
    operate_date_start timestamp without time zone,
    operate_date_end timestamp without time zone,
    operate_comment character varying(255),
    remove_by character varying(255),
    remove_date timestamp without time zone,
    remove_status smallint,
    execute_param json
);


ALTER TABLE ssodb.t_import_job OWNER TO ssouser;

--
-- Name: t_import_job_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_import_job_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_import_job_id_seq OWNER TO ssouser;

--
-- Name: t_import_job_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_import_job_id_seq OWNED BY ssodb.t_import_job.id;


--
-- Name: t_log_admin; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_log_admin (
    id integer NOT NULL,
    account character varying(255),
    ip character varying(255),
    created_date timestamp without time zone,
    result smallint,
    ua character varying(1024)
);


ALTER TABLE ssodb.t_log_admin OWNER TO ssouser;

--
-- Name: t_log_admin_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_log_admin_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_log_admin_id_seq OWNER TO ssouser;

--
-- Name: t_log_admin_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_log_admin_id_seq OWNED BY ssodb.t_log_admin.id;


--
-- Name: t_log_app_access; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_log_app_access (
    uuid character varying(40) NOT NULL,
    status smallint,
    appid character varying(255),
    apiid character varying(40),
    ip character varying(255),
    start_date timestamp without time zone,
    end_date timestamp without time zone,
    url character varying(1024),
    ua character varying(1024),
    parameter text,
    response text,
    exception text
);


ALTER TABLE ssodb.t_log_app_access OWNER TO ssouser;

--
-- Name: t_log_app_verify; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_log_app_verify (
    id integer NOT NULL,
    appid character varying(255),
    ip character varying(255),
    created_date timestamp without time zone,
    result smallint,
    ua character varying(1024)
);


ALTER TABLE ssodb.t_log_app_verify OWNER TO ssouser;

--
-- Name: t_log_app_verify_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_log_app_verify_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_log_app_verify_id_seq OWNER TO ssouser;

--
-- Name: t_log_app_verify_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_log_app_verify_id_seq OWNED BY ssodb.t_log_app_verify.id;


--
-- Name: t_log_audit; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_log_audit (
    id integer NOT NULL,
    operator_schoolid character varying(255),
    operator_ip character varying(255),
    operated_schoolid character varying(255),
    operation_type character varying(255),
    table_name character varying(255),
    table_columns text,
    operation_before text,
    operation_after text,
    operation_date timestamp without time zone
);


ALTER TABLE ssodb.t_log_audit OWNER TO ssouser;

--
-- Name: t_log_audit_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_log_audit_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_log_audit_id_seq OWNER TO ssouser;

--
-- Name: t_log_audit_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_log_audit_id_seq OWNED BY ssodb.t_log_audit.id;


--
-- Name: t_log_auth; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_log_auth (
    id integer NOT NULL,
    class text,
    method text,
    input text,
    output text,
    exec_time integer,
    remark text,
    code integer,
    ip text,
    geo text,
    msg text,
    flow_key text,
    created_date timestamp without time zone DEFAULT now(),
    username text,
    service_id integer,
    login_type text
)
PARTITION BY RANGE (created_date);


ALTER TABLE ssodb.t_log_auth OWNER TO ssouser;

--
-- Name: COLUMN t_log_auth.class; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_auth.class IS '系统类名';


--
-- Name: COLUMN t_log_auth.method; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_auth.method IS '系统类的方法';


--
-- Name: COLUMN t_log_auth.input; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_auth.input IS '入参数 json';


--
-- Name: COLUMN t_log_auth.output; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_auth.output IS '返回值 json';


--
-- Name: COLUMN t_log_auth.exec_time; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_auth.exec_time IS '执行用时';


--
-- Name: COLUMN t_log_auth.remark; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_auth.remark IS '中文标记，mfa时说明时mfa登陆';


--
-- Name: COLUMN t_log_auth.code; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_auth.code IS '返回状态code';


--
-- Name: COLUMN t_log_auth.geo; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_auth.geo IS 'ip信息';


--
-- Name: COLUMN t_log_auth.msg; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_auth.msg IS '返回的msg信息';


--
-- Name: COLUMN t_log_auth.flow_key; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_auth.flow_key IS '流水Id';


--
-- Name: COLUMN t_log_auth.created_date; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_auth.created_date IS '创建时间';


--
-- Name: COLUMN t_log_auth.username; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_auth.username IS '未认证前是传过来的username参数，认证后是查到的uid';


--
-- Name: t_log_auth_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_log_auth_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_log_auth_id_seq OWNER TO ssouser;

--
-- Name: t_log_auth_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_log_auth_id_seq OWNED BY ssodb.t_log_auth.id;


--
-- Name: t_log_auth_history; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_log_auth_history (
    id integer DEFAULT nextval('ssodb.t_log_auth_id_seq'::regclass) NOT NULL,
    class text,
    method text,
    input text,
    output text,
    exec_time integer,
    remark text,
    code integer,
    ip text,
    geo text,
    msg text,
    flow_key text,
    created_date timestamp without time zone DEFAULT now(),
    username text,
    service_id integer,
    login_type text
);


ALTER TABLE ssodb.t_log_auth_history OWNER TO ssouser;

--
-- Name: t_log_online; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_log_online (
    id integer NOT NULL,
    tgt character varying,
    status smallint,
    username character varying,
    principal character varying,
    login_type character varying,
    ip character varying,
    geo character varying,
    ua character varying,
    created_date timestamp without time zone,
    slo_type smallint,
    device_uuid character varying,
    uid character varying
);


ALTER TABLE ssodb.t_log_online OWNER TO ssouser;

--
-- Name: TABLE t_log_online; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON TABLE ssodb.t_log_online IS '存储tgt用户过期回调';


--
-- Name: COLUMN t_log_online.status; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_online.status IS '  0 offline离线, 1 online在线';


--
-- Name: COLUMN t_log_online.username; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_online.username IS '用户名，也可能时其他用户标识';


--
-- Name: COLUMN t_log_online.principal; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_online.principal IS 'tgt关联用户的内容';


--
-- Name: COLUMN t_log_online.login_type; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_online.login_type IS '登陆类型';


--
-- Name: COLUMN t_log_online.ip; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_online.ip IS 'ip地址';


--
-- Name: COLUMN t_log_online.geo; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_online.geo IS '地理位置';


--
-- Name: COLUMN t_log_online.ua; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_online.ua IS '浏览器ua';


--
-- Name: COLUMN t_log_online.created_date; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_online.created_date IS '创建时间，timestamp';


--
-- Name: COLUMN t_log_online.slo_type; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_online.slo_type IS '3x(32,31,30)超过tgt数量触发退出  2x(21,20)用户主动slo退出  1需要slo系统配置单点登录  0不需要slo';


--
-- Name: COLUMN t_log_online.device_uuid; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_online.device_uuid IS '设备唯一标识';


--
-- Name: t_log_online_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_log_online_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_log_online_id_seq OWNER TO ssouser;

--
-- Name: t_log_online_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_log_online_id_seq OWNED BY ssodb.t_log_online.id;


--
-- Name: t_log_task; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_log_task (
    id integer NOT NULL,
    taskid integer,
    start_date timestamp without time zone,
    end_date timestamp without time zone,
    executed_result_content text,
    executed_result_status smallint
);


ALTER TABLE ssodb.t_log_task OWNER TO ssouser;

--
-- Name: t_log_task_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_log_task_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_log_task_id_seq OWNER TO ssouser;

--
-- Name: t_log_task_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_log_task_id_seq OWNED BY ssodb.t_log_task.id;


--
-- Name: t_log_trusted_device; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_log_trusted_device (
    id bigint NOT NULL,
    device_uuid character varying,
    uid character varying,
    device_name character varying,
    ua character varying,
    geo character varying,
    ip character varying,
    created_date timestamp without time zone
);


ALTER TABLE ssodb.t_log_trusted_device OWNER TO ssouser;

--
-- Name: TABLE t_log_trusted_device; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON TABLE ssodb.t_log_trusted_device IS '常用设备';


--
-- Name: COLUMN t_log_trusted_device.device_name; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_log_trusted_device.device_name IS '别名，用户可自定义';


--
-- Name: t_log_trusted_device_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_log_trusted_device_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_log_trusted_device_id_seq OWNER TO ssouser;

--
-- Name: t_log_trusted_device_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_log_trusted_device_id_seq OWNED BY ssodb.t_log_trusted_device.id;


--
-- Name: t_news; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_news (
    id integer NOT NULL,
    news_url character varying(255),
    news_title character varying(255),
    news_summary character varying(1024),
    news_pictureid character varying(255),
    created_by character varying(255),
    created_date timestamp without time zone,
    last_modified_by character varying(255),
    last_modified_date timestamp without time zone,
    release_index integer,
    release_date_from timestamp without time zone,
    release_date_to timestamp without time zone
);


ALTER TABLE ssodb.t_news OWNER TO ssouser;

--
-- Name: t_news_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_news_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_news_id_seq OWNER TO ssouser;

--
-- Name: t_news_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_news_id_seq OWNED BY ssodb.t_news.id;


--
-- Name: t_service; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_service (
    id integer NOT NULL,
    priority integer,
    service_name character varying(255),
    type character varying(255),
    url_whitelist json,
    ip_whitelist json,
    expiration_policy json,
    attribute_policy json,
    access_policy json,
    mfa_policy json,
    access_key character varying(255),
    secret_key character varying(255),
    contacts json,
    prompt character varying(255),
    slo_enabled smallint
);


ALTER TABLE ssodb.t_service OWNER TO ssouser;

--
-- Name: COLUMN t_service.priority; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_service.priority IS '优先级（越小优先级越高）';


--
-- Name: COLUMN t_service.service_name; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_service.service_name IS '第三方服务名称';


--
-- Name: COLUMN t_service.type; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_service.type IS 'cas-web（默认）、cas-restful、oauth20、samal20等';


--
-- Name: COLUMN t_service.url_whitelist; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_service.url_whitelist IS 'ant/regexp';


--
-- Name: COLUMN t_service.ip_whitelist; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_service.ip_whitelist IS 'json ，如：["127.0.0.1","192.168.1.0/24"]';


--
-- Name: COLUMN t_service.expiration_policy; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_service.expiration_policy IS '过期策略  长期有效，指定时间内有效 {"start":"","end":""}';


--
-- Name: COLUMN t_service.attribute_policy; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_service.attribute_policy IS '用户授权字段';


--
-- Name: COLUMN t_service.access_policy; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_service.access_policy IS '准入黑白明单';


--
-- Name: COLUMN t_service.mfa_policy; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_service.mfa_policy IS 'mfa策略 开启（强制开启、指定标签tagId） 关闭';


--
-- Name: COLUMN t_service.access_key; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_service.access_key IS 'oauth时的access key，ldap时的用户名';


--
-- Name: COLUMN t_service.secret_key; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_service.secret_key IS 'oauth时的secret key，ldap时的密码';


--
-- Name: COLUMN t_service.contacts; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_service.contacts IS '联系人';


--
-- Name: COLUMN t_service.prompt; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_service.prompt IS '提醒：none为不提醒，consent为提醒用户同意';


--
-- Name: COLUMN t_service.slo_enabled; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_service.slo_enabled IS '0关闭， 1开启  单点登出是否开启';


--
-- Name: t_service_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_service_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_service_id_seq OWNER TO ssouser;

--
-- Name: t_service_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_service_id_seq OWNED BY ssodb.t_service.id;


--
-- Name: t_service_ldap; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_service_ldap (
    id integer NOT NULL,
    username character varying(255),
    password character varying(255),
    expiration_date timestamp without time zone,
    comment character varying(255),
    allowed_tags json,
    allowed_attributes json
);


ALTER TABLE ssodb.t_service_ldap OWNER TO ssouser;

--
-- Name: t_service_ldap_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_service_ldap_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_service_ldap_id_seq OWNER TO ssouser;

--
-- Name: t_service_ldap_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_service_ldap_id_seq OWNED BY ssodb.t_service_ldap.id;


--
-- Name: t_system_across_region; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_system_across_region (
    id integer NOT NULL,
    is_disabled integer,
    white_id text,
    created_by character varying(255),
    created_date timestamp without time zone
);


ALTER TABLE ssodb.t_system_across_region OWNER TO ssouser;

--
-- Name: t_system_across_region_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_system_across_region_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_system_across_region_id_seq OWNER TO ssouser;

--
-- Name: t_system_across_region_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_system_across_region_id_seq OWNED BY ssodb.t_system_across_region.id;


--
-- Name: t_system_password_weak; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_system_password_weak (
    id integer NOT NULL,
    filter text,
    created_by character varying(255),
    created_date timestamp without time zone,
    is_disabled integer,
    last_modified_by character varying(255),
    last_modified_date timestamp without time zone
);


ALTER TABLE ssodb.t_system_password_weak OWNER TO ssouser;

--
-- Name: t_system_password_weak_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_system_password_weak_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_system_password_weak_id_seq OWNER TO ssouser;

--
-- Name: t_system_password_weak_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_system_password_weak_id_seq OWNED BY ssodb.t_system_password_weak.id;


--
-- Name: t_system_property_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_system_property_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_system_property_id_seq OWNER TO ssouser;

--
-- Name: t_system_property_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_system_property_id_seq OWNED BY ssodb.t_system_property.id;


--
-- Name: t_system_resource; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_system_resource (
    id integer NOT NULL,
    key character varying(255),
    value text,
    status integer,
    description text,
    last_modified_by character varying(255),
    last_modified_date timestamp without time zone
);


ALTER TABLE ssodb.t_system_resource OWNER TO ssouser;

--
-- Name: t_system_resource_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_system_resource_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_system_resource_id_seq OWNER TO ssouser;

--
-- Name: t_system_resource_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_system_resource_id_seq OWNED BY ssodb.t_system_resource.id;


--
-- Name: t_tag; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_tag (
    id integer NOT NULL,
    name character varying(255),
    type character varying(40),
    priority integer,
    filterid integer,
    last_modified_by character varying(255),
    last_modified_date timestamp without time zone
);


ALTER TABLE ssodb.t_tag OWNER TO ssouser;

--
-- Name: t_tag_filter; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_tag_filter (
    id integer NOT NULL,
    name character varying(255),
    priority integer,
    expression_name text,
    expression_content json,
    last_modified_by character varying(255),
    last_modified_date timestamp without time zone
);


ALTER TABLE ssodb.t_tag_filter OWNER TO ssouser;

--
-- Name: t_tag_filter_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_tag_filter_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_tag_filter_id_seq OWNER TO ssouser;

--
-- Name: t_tag_filter_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_tag_filter_id_seq OWNED BY ssodb.t_tag_filter.id;


--
-- Name: t_tag_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_tag_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_tag_id_seq OWNER TO ssouser;

--
-- Name: t_tag_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_tag_id_seq OWNED BY ssodb.t_tag.id;


--
-- Name: t_tag_user; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_tag_user (
    id integer NOT NULL,
    tagid integer,
    userid integer,
    last_modified_by character varying(255),
    last_modified_date timestamp without time zone
);


ALTER TABLE ssodb.t_tag_user OWNER TO ssouser;

--
-- Name: t_tag_user_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_tag_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_tag_user_id_seq OWNER TO ssouser;

--
-- Name: t_tag_user_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_tag_user_id_seq OWNED BY ssodb.t_tag_user.id;


--
-- Name: t_task; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_task (
    id integer NOT NULL,
    name character varying(255),
    description text,
    task_type character varying(40),
    repeat_type character varying(40),
    executed_date character varying(255),
    executed_path text,
    executed_content text,
    enabled_status integer,
    executed_status integer,
    created_by character varying(255),
    created_date timestamp without time zone,
    last_modified_by character varying(255),
    last_modified_date timestamp without time zone
);


ALTER TABLE ssodb.t_task OWNER TO ssouser;

--
-- Name: t_task_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_task_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_task_id_seq OWNER TO ssouser;

--
-- Name: t_task_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_task_id_seq OWNED BY ssodb.t_task.id;


--
-- Name: t_user_attribute; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_user_attribute (
    userid integer NOT NULL,
    longtext text,
    datetype timestamp without time zone
);


ALTER TABLE ssodb.t_user_attribute OWNER TO ssouser;

--
-- Name: t_user_conf_notification; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_user_conf_notification (
    id integer NOT NULL,
    userid integer,
    login json,
    risk json,
    change json,
    last_modified_by character varying(255),
    last_modified_date timestamp without time zone
);


ALTER TABLE ssodb.t_user_conf_notification OWNER TO ssouser;

--
-- Name: t_user_conf_notification_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_user_conf_notification_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_user_conf_notification_id_seq OWNER TO ssouser;

--
-- Name: t_user_conf_notification_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_user_conf_notification_id_seq OWNED BY ssodb.t_user_conf_notification.id;


--
-- Name: t_user_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_user_id_seq
    AS integer
    START WITH 100000
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_user_id_seq OWNER TO ssouser;

--
-- Name: t_user_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_user_id_seq OWNED BY ssodb.t_user.id;


--
-- Name: t_user_image; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_user_image (
    id integer NOT NULL,
    userid integer,
    fileid integer,
    status integer,
    created_by character varying(255),
    created_date timestamp without time zone
);


ALTER TABLE ssodb.t_user_image OWNER TO ssouser;

--
-- Name: t_user_image_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_user_image_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_user_image_id_seq OWNER TO ssouser;

--
-- Name: t_user_image_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_user_image_id_seq OWNED BY ssodb.t_user_image.id;


--
-- Name: t_user_role; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_user_role (
    id integer NOT NULL,
    userid integer,
    roleid integer,
    allow_decentralization integer,
    allow_decentralization_from_userid integer
);


ALTER TABLE ssodb.t_user_role OWNER TO ssouser;

--
-- Name: t_user_role_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_user_role_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_user_role_id_seq OWNER TO ssouser;

--
-- Name: t_user_role_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_user_role_id_seq OWNED BY ssodb.t_user_role.id;


--
-- Name: t_user_surrogate; Type: TABLE; Schema: ssodb; Owner: ssouser
--

CREATE TABLE ssodb.t_user_surrogate (
    id integer NOT NULL,
    schoolid character varying(255),
    surrogate_schoolid character varying(255),
    start_date timestamp without time zone,
    end_date timestamp without time zone,
    is_disabled smallint,
    last_modified_by character varying(255),
    last_modified_date timestamp without time zone
);


ALTER TABLE ssodb.t_user_surrogate OWNER TO ssouser;

--
-- Name: COLUMN t_user_surrogate.schoolid; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user_surrogate.schoolid IS '学工号';


--
-- Name: COLUMN t_user_surrogate.surrogate_schoolid; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user_surrogate.surrogate_schoolid IS '委托人学工号';


--
-- Name: COLUMN t_user_surrogate.start_date; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user_surrogate.start_date IS '委托开始时间';


--
-- Name: COLUMN t_user_surrogate.end_date; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user_surrogate.end_date IS '委托结束时间';


--
-- Name: COLUMN t_user_surrogate.is_disabled; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user_surrogate.is_disabled IS '停用状态，0为可用，1为停用';


--
-- Name: COLUMN t_user_surrogate.last_modified_by; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user_surrogate.last_modified_by IS '修改者username';


--
-- Name: COLUMN t_user_surrogate.last_modified_date; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.t_user_surrogate.last_modified_date IS '修改时间';


--
-- Name: t_user_surrogate_id_seq; Type: SEQUENCE; Schema: ssodb; Owner: ssouser
--

CREATE SEQUENCE ssodb.t_user_surrogate_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ssodb.t_user_surrogate_id_seq OWNER TO ssouser;

--
-- Name: t_user_surrogate_id_seq; Type: SEQUENCE OWNED BY; Schema: ssodb; Owner: ssouser
--

ALTER SEQUENCE ssodb.t_user_surrogate_id_seq OWNED BY ssodb.t_user_surrogate.id;


--
-- Name: v_account; Type: VIEW; Schema: ssodb; Owner: ssouser
--

CREATE VIEW ssodb.v_account AS
 SELECT ta.username,
    ta.password,
    vuas.is_disabled,
    vuas.is_locked,
    vuas.is_expired
   FROM (ssodb.t_account ta
     LEFT JOIN ssodb.i_user_account_status vuas ON ((vuas.accountid = ta.id)))
  WHERE (ta.username IS NOT NULL)
UNION ALL
 SELECT tu.schoolid AS username,
    ta.password,
    vuas.is_disabled,
    vuas.is_locked,
    vuas.is_expired
   FROM ((ssodb.t_account ta
     LEFT JOIN ssodb.t_user tu ON ((ta.userid = tu.id)))
     LEFT JOIN ssodb.i_user_account_status vuas ON ((vuas.accountid = ta.id)))
  WHERE (ta.username IS NULL);


ALTER VIEW ssodb.v_account OWNER TO ssouser;

--
-- Name: v_user_attribute_cas; Type: VIEW; Schema: ssodb; Owner: ssouser
--

CREATE VIEW ssodb.v_user_attribute_cas AS
 SELECT t1.department,
    t1.id AS userid,
    t3.datetype,
    t3.longtext,
    t1.ic_number
   FROM ((ssodb.t_user t1
     LEFT JOIN ssodb.t_account t2 ON ((t2.userid = t1.id)))
     LEFT JOIN ssodb.t_user_attribute t3 ON ((t3.userid = t1.id)));


ALTER VIEW ssodb.v_user_attribute_cas OWNER TO ssouser;

--
-- Name: v_user_attribute_ldap; Type: VIEW; Schema: ssodb; Owner: ssouser
--

CREATE VIEW ssodb.v_user_attribute_ldap AS
 SELECT t1.department,
    t1.id AS userid,
    t3.datetype,
    t3.longtext,
    t1.ic_number
   FROM ((ssodb.t_user t1
     LEFT JOIN ssodb.t_account t2 ON ((t2.userid = t1.id)))
     LEFT JOIN ssodb.t_user_attribute t3 ON ((t3.userid = t1.id)));


ALTER VIEW ssodb.v_user_attribute_ldap OWNER TO ssouser;

--
-- Name: COLUMN v_user_attribute_ldap.department; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.v_user_attribute_ldap.department IS '所在单位';


--
-- Name: COLUMN v_user_attribute_ldap.ic_number; Type: COMMENT; Schema: ssodb; Owner: ssouser
--

COMMENT ON COLUMN ssodb.v_user_attribute_ldap.ic_number IS '身份证件号';


--
-- Name: v_user_attributes_derived; Type: VIEW; Schema: ssodb; Owner: ssouser
--

CREATE VIEW ssodb.v_user_attributes_derived AS
 SELECT tu.schoolid AS employee_number,
    tu.schoolid AS uid,
    vuas.is_expired,
    vuas.is_locked,
    vuas.is_disabled,
    tu.static_tag AS tag,
    tu.schoolid AS edu_person_principal_name,
    tu.type AS edu_person_scoped_affiliation
   FROM ((ssodb.t_user tu
     LEFT JOIN ssodb.t_account ta ON ((tu.id = ta.userid)))
     LEFT JOIN ssodb.i_user_account_status vuas ON ((tu.id = vuas.userid)));


ALTER VIEW ssodb.v_user_attributes_derived OWNER TO ssouser;

--
-- Name: t_log_auth_history; Type: TABLE ATTACH; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_log_auth ATTACH PARTITION ssodb.t_log_auth_history FOR VALUES FROM ('2000-01-01 00:00:00') TO ('2049-01-01 00:00:00');


--
-- Name: t_account id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_account ALTER COLUMN id SET DEFAULT nextval('ssodb.t_account_id_seq'::regclass);


--
-- Name: t_account_totp id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_account_totp ALTER COLUMN id SET DEFAULT nextval('ssodb.t_account_totp_id_seq'::regclass);


--
-- Name: t_admin id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_admin ALTER COLUMN id SET DEFAULT nextval('ssodb.t_admin_id_seq'::regclass);


--
-- Name: t_app id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_app ALTER COLUMN id SET DEFAULT nextval('ssodb.t_app_id_seq'::regclass);


--
-- Name: t_app_token id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_app_token ALTER COLUMN id SET DEFAULT nextval('ssodb.t_app_token_id_seq'::regclass);


--
-- Name: t_attribute id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_attribute ALTER COLUMN id SET DEFAULT nextval('ssodb.t_attribute_id_seq'::regclass);


--
-- Name: t_code id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_code ALTER COLUMN id SET DEFAULT nextval('ssodb.t_code_id_seq'::regclass);


--
-- Name: t_code_userrole id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_code_userrole ALTER COLUMN id SET DEFAULT nextval('ssodb.t_code_userrole_id_seq'::regclass);


--
-- Name: t_export_job id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_export_job ALTER COLUMN id SET DEFAULT nextval('ssodb.t_export_job_id_seq'::regclass);


--
-- Name: t_file_baseinfo id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_file_baseinfo ALTER COLUMN id SET DEFAULT nextval('ssodb.t_file_baseinfo_id_seq'::regclass);


--
-- Name: t_import_job id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_import_job ALTER COLUMN id SET DEFAULT nextval('ssodb.t_import_job_id_seq'::regclass);


--
-- Name: t_log_admin id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_log_admin ALTER COLUMN id SET DEFAULT nextval('ssodb.t_log_admin_id_seq'::regclass);


--
-- Name: t_log_app_verify id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_log_app_verify ALTER COLUMN id SET DEFAULT nextval('ssodb.t_log_app_verify_id_seq'::regclass);


--
-- Name: t_log_audit id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_log_audit ALTER COLUMN id SET DEFAULT nextval('ssodb.t_log_audit_id_seq'::regclass);


--
-- Name: t_log_auth id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_log_auth ALTER COLUMN id SET DEFAULT nextval('ssodb.t_log_auth_id_seq'::regclass);


--
-- Name: t_log_online id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_log_online ALTER COLUMN id SET DEFAULT nextval('ssodb.t_log_online_id_seq'::regclass);


--
-- Name: t_log_task id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_log_task ALTER COLUMN id SET DEFAULT nextval('ssodb.t_log_task_id_seq'::regclass);


--
-- Name: t_log_trusted_device id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_log_trusted_device ALTER COLUMN id SET DEFAULT nextval('ssodb.t_log_trusted_device_id_seq'::regclass);


--
-- Name: t_news id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_news ALTER COLUMN id SET DEFAULT nextval('ssodb.t_news_id_seq'::regclass);


--
-- Name: t_service id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_service ALTER COLUMN id SET DEFAULT nextval('ssodb.t_service_id_seq'::regclass);


--
-- Name: t_service_ldap id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_service_ldap ALTER COLUMN id SET DEFAULT nextval('ssodb.t_service_ldap_id_seq'::regclass);


--
-- Name: t_system_across_region id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_system_across_region ALTER COLUMN id SET DEFAULT nextval('ssodb.t_system_across_region_id_seq'::regclass);


--
-- Name: t_system_password_weak id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_system_password_weak ALTER COLUMN id SET DEFAULT nextval('ssodb.t_system_password_weak_id_seq'::regclass);


--
-- Name: t_system_property id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_system_property ALTER COLUMN id SET DEFAULT nextval('ssodb.t_system_property_id_seq'::regclass);


--
-- Name: t_system_resource id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_system_resource ALTER COLUMN id SET DEFAULT nextval('ssodb.t_system_resource_id_seq'::regclass);


--
-- Name: t_tag id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_tag ALTER COLUMN id SET DEFAULT nextval('ssodb.t_tag_id_seq'::regclass);


--
-- Name: t_tag_filter id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_tag_filter ALTER COLUMN id SET DEFAULT nextval('ssodb.t_tag_filter_id_seq'::regclass);


--
-- Name: t_tag_user id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_tag_user ALTER COLUMN id SET DEFAULT nextval('ssodb.t_tag_user_id_seq'::regclass);


--
-- Name: t_task id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_task ALTER COLUMN id SET DEFAULT nextval('ssodb.t_task_id_seq'::regclass);


--
-- Name: t_user id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_user ALTER COLUMN id SET DEFAULT nextval('ssodb.t_user_id_seq'::regclass);


--
-- Name: t_user_conf_notification id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_user_conf_notification ALTER COLUMN id SET DEFAULT nextval('ssodb.t_user_conf_notification_id_seq'::regclass);


--
-- Name: t_user_image id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_user_image ALTER COLUMN id SET DEFAULT nextval('ssodb.t_user_image_id_seq'::regclass);


--
-- Name: t_user_role id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_user_role ALTER COLUMN id SET DEFAULT nextval('ssodb.t_user_role_id_seq'::regclass);


--
-- Name: t_user_surrogate id; Type: DEFAULT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_user_surrogate ALTER COLUMN id SET DEFAULT nextval('ssodb.t_user_surrogate_id_seq'::regclass);


--
-- Data for Name: t_account; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--

INSERT INTO ssodb.t_account VALUES (1, 'ssotest', '{SSHA}sf2TbEWAcnkG9eEyQrf4sqLPVVC/FlnCsxavvg==', '2024-02-15 00:48:54.639014', 1, NULL, '15324289963', '0', '2023-01-12 08:17:56', ',1110,1108,1107,1,', '13061111', NULL, 'admin', '2024-02-15 00:48:54.639014', NULL, NULL);


--
-- Data for Name: t_account_totp; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_admin; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--

INSERT INTO ssodb.t_admin VALUES (1, 'admin', '{NONE}a1b2c3', 0, '2023-03-23 14:26:00.550424', '0', NULL, '67348');


--
-- Data for Name: t_app; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_app_token; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_attribute; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--

INSERT INTO ssodb.t_attribute VALUES (1, 'sex', '性别', 'VARCHAR', 'NATIVE', 0, 0, 0, 1, 0, NULL, NULL, NULL, NULL, NULL, NULL);
INSERT INTO ssodb.t_attribute VALUES (2, 'type', '人员类别', 'VARCHAR', 'NATIVE', 0, 0, 0, 1, 0, NULL, NULL, NULL, NULL, NULL, NULL);
INSERT INTO ssodb.t_attribute VALUES (3, 'status', '在校状态', 'VARCHAR', 'NATIVE', 0, 0, 0, 1, 0, NULL, NULL, NULL, NULL, NULL, NULL);
INSERT INTO ssodb.t_attribute VALUES (4, 'department', '组织机构', 'VARCHAR', 'NATIVE', 0, 1, 1, 1, 0, 'admin', '2023-04-10 22:44:13.601935', 'admin', '2023-04-10 22:44:13.601935', NULL, NULL);
INSERT INTO ssodb.t_attribute VALUES (5, 'ic_type', '证件类别', 'VARCHAR', 'NATIVE', 0, 0, 0, 1, 0, NULL, NULL, NULL, NULL, NULL, NULL);
INSERT INTO ssodb.t_attribute VALUES (6, 'political', '政治面貌', 'VARCHAR', 'NATIVE', 0, 0, 0, 1, 0, NULL, NULL, NULL, NULL, NULL, NULL);
INSERT INTO ssodb.t_attribute VALUES (7, 'ic_number', '身份证号码', 'VARCHAR', 'NATIVE', 0, 1, 1, 1, 0, 'admin', '2023-04-20 20:18:11.445799', 'admin', '2023-04-20 20:18:11.445799', 1, 'id');


--
-- Data for Name: t_code; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--

INSERT INTO ssodb.t_code VALUES (1, '01', '在校', 'atSchoolStatus', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (2, '99', '离校', 'atSchoolStatus', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (3, '1', '男', 'gender', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (4, '2', '女', 'gender', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (5, '0', '未知的性别', 'gender', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (6, '9', '未说明的性别', 'gender', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (7, '1', '居民身份证', 'id', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (8, '2', '军官证', 'id', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (9, '3', '士兵证', 'id', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (10, '4', '文职干部证', 'id', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (11, '5', '部队离退休证', 'id', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (12, '6', '香港特区护照/身份证明', 'id', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (13, '7', '澳门特区护照/身份证明', 'id', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (14, '8', '台湾居民来往大陆通行证（台胞证）', 'id', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (15, '9', '境外永久居住证', 'id', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (16, 'A', '护照', 'id', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (17, 'B', '户口薄', 'id', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (18, 'C', '港澳居民来往内地通行证（回乡证）', 'id', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (19, 'Z', '其他', 'id', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (20, '0', '启用', 'isDisabled', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (21, '1', '停用', 'isDisabled', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (22, '0', '未锁定', 'isLocked', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (23, '1', '风险锁定', 'isLocked', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (24, '2', '管理员锁定', 'isLocked', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (25, '0', '未注册', 'isRegistered', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (26, '1', '已注册', 'isRegistered', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (27, '0', '未闲置', 'loginUnuse', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (28, '1', '闲置', 'loginUnuse', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (29, '1', '启用', 'operateType', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (30, '2', '停用', 'operateType', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (31, '3', '定期停用', 'operateType', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (32, '4', '导入', 'operateType', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (33, '5', '修改', 'operateType', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (34, '6', '删除', 'operateType', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (35, '7', '角色', 'operateType', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (36, '01', '中国共产党党员', 'political', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (37, '02', '中国共产党预备党员', 'political', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (38, '03', '中国共产主义青年团团员', 'political', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (39, '04', '中国国民党革命委员会会员', 'political', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (40, '05', '中国民主同盟盟员', 'political', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (41, '06', '中国民主建国会会员', 'political', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (42, '07', '中国民主促进会会员', 'political', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (43, '08', '中国民主农工党党员', 'political', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (44, '09', '中国致公党党员', 'political', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (45, '10', '九三学社社员', 'political', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (46, '11', '台湾民主自治同盟盟员', 'political', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (47, '12', '无党派民主人士', 'political', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (48, '13', '群众', 'political', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (49, '99', '未说明', 'political', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (50, '1', '学生', 'usertype', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (51, '2', '教职工', 'usertype', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (52, '3', '临时人员', 'usertype', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (53, '0', '否', 'yesOrNo', 1, NULL, NULL);
INSERT INTO ssodb.t_code VALUES (54, '1', '是', 'yesOrNo', 1, NULL, NULL);


--
-- Data for Name: t_code_department; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--

INSERT INTO ssodb.t_code_department VALUES ('000000', '单点登录大学', NULL, 1, 0, 1, '/000000/');


--
-- Data for Name: t_code_userrole; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--

INSERT INTO ssodb.t_code_userrole VALUES (40, '全部权限', NULL, '{"app": {"privilege": "*"}, "log": {"privilege": "*"}, "user": [{"scope": {"type": ["L01"], "depart": "010100"}, "privilege": [["query"], ["modify"], ["account/query"], ["account/modify"]]}], "other": {"privilege": "*"}, "backup": {"privilege": "*"}, "config": {"privilege": "*"}, "groups": {"privilege": "*"}, "publish": {"privilege": "*"}, "service": {"privilege": "*"}, "authorization": {"privilege": "*"}, "statistic/query": {"privilege": "*"}}');
INSERT INTO ssodb.t_code_userrole VALUES (41, '系统管理员', NULL, '{"app": {"privilege": "*"}, "log": {"privilege": "*"}, "user": [{"scope": {"type": ["L01"], "depart": "yjs-2021A100102"}, "privilege": "*"}, {"scope": {"type": ["L0102"], "depart": "010100"}, "privilege": "*"}], "other": {"privilege": "*"}, "backup": {"privilege": "*"}, "config": {"privilege": "*"}, "groups": {"privilege": "*"}, "publish": {"privilege": "*"}, "service": {"privilege": "*"}, "authorization": {"privilege": "*"}, "statistic/query": {"privilege": "*"}}');


--
-- Data for Name: t_code_usertype; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--

INSERT INTO ssodb.t_code_usertype VALUES ('L01', '人员', NULL, 1, 0, 1, '/L01/');


--
-- Data for Name: t_export_job; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_file_baseinfo; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_import_backup; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_import_job; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_log_admin; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--


--
-- Data for Name: t_log_app_access; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_log_app_verify; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_log_audit; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--


--
-- Data for Name: t_log_auth_history; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--


--
-- Data for Name: t_log_online; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_log_task; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_log_trusted_device; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_news; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_service; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--

INSERT INTO ssodb.t_service VALUES (1, 10001, 'restful测试', 'cas-restful', '{"type":"ant","url":"*"}', '["127.0.0.1"]', '{"start":"1970-01-01 08:00:00.000","end":"2031-04-22 03:58:10.123"}', '{"phone":"phone","schoolid":"schoolid","employee_number":"employeeNumber"}', '{"blackTags":[{"@type":"Static","tagId":"1"},{"@type":"Static","tagId":"1"}],"whiteTags":[{"@type":"Dynamic","tagId":"2","condition":null},{"@type":"Dynamic","tagId":"2","condition":null}]}', '{"method":"totp","tags":[{"@type":"Dynamic","tagId":"2"},{"@type":"Static","tagId":"y"}]}', '', NULL, '[{"name":"张三","phone":"1352019999","email":null,"schoolid":null,"department":null}]', 'none', 0);
INSERT INTO ssodb.t_service VALUES (2, 10100, 'web测试', 'cas-web', '{"type":"regexp","url":"http://*"}', NULL, '{"end":"2024-12-27 00:00:00.000"}', '{"ic_number":"ic_number","longtext":"longtext","abcdx":"abcdx"}', '{"blackTags":[{"@type":"Static","tagId":"44"},{"@type":"Dynamic","tagId":"45","condition":null}],"whiteTags":[{"@type":"Dynamic","tagId":"42","condition":null},{"@type":"Static","tagId":"43"}]}', '{"method":"sms","tags":[{"@type":"Dynamic","tagId":"3"}]}', NULL, NULL, '[{"name":"a1","phone":"a1","email":"a1","schoolid":"a1","department":"a1"}]', NULL, NULL);


--
-- Data for Name: t_service_ldap; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_system_across_region; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_system_password_weak; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_system_property; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--

INSERT INTO ssodb.t_system_property VALUES (10, 'syslog_config', '12312312,12312312222', 0, '2023-04-20 21:31:17.977475');
INSERT INTO ssodb.t_system_property VALUES (11, 'syslog_config', '12312312,12312312222,12,1,2', 1, '2023-04-20 21:31:33.528511');
INSERT INTO ssodb.t_system_property VALUES (13, 'mfa_white_providers', 'mfa-gauth,mfa-simple', 1, '2023-06-07 10:04:28');
INSERT INTO ssodb.t_system_property VALUES (18, 'mfa_login_user_geo_location_sql', 'SELECT ep.`value` clientip FROM t_cas_event ce LEFT JOIN t_event_property ep ON ce.id = ep.event_id WHERE ce.principal_id = ? AND ep.`name` = ''clientip'' ORDER BY ce.creation_time desc LIMIT 2', 1, '2023-06-07 11:44:55');
INSERT INTO ssodb.t_system_property VALUES (22, 'mfa_type', 'sms', 1, '2024-01-01 13:04:21');
INSERT INTO ssodb.t_system_property VALUES (9, 'syslog_config', '127.0.0.1:7788', 0, '2023-04-20 14:16:03.824022');
INSERT INTO ssodb.t_system_property VALUES (14, 'mfa_white_login_types', 'mc-qr,campus-wy', 1, '2023-06-07 10:41:03');
INSERT INTO ssodb.t_system_property VALUES (15, 'mfa_white_users', 'zhangsan', 1, '2023-06-07 10:53:28');
INSERT INTO ssodb.t_system_property VALUES (17, 'mfa_white_locations', 'Shanghai,Beijing', 1, '2023-06-07 11:09:17');
INSERT INTO ssodb.t_system_property VALUES (16, 'mfa_white_ips', '127.0.0.1,202.206.28.145', 1, '2023-06-07 10:59:23');
INSERT INTO ssodb.t_system_property VALUES (12, 'mfa_black_times', '{"startTime":"20:30:01","endTime":"08:40:01","mfaProvider":"mfa-simple,mfa-gauth"}', 1, '2023-06-01 17:27:51');
INSERT INTO ssodb.t_system_property VALUES (19, 'password_regex', '^(?=.*[0-9])(?=.*[a-z])(?=.*[A-Z])(?=.*[^a-zA-Z0-9])([0-9]|[a-z]|[A-Z]|[!@#$%^&*()]){12,16}$', 1, '2023-12-21 18:27:16');
INSERT INTO ssodb.t_system_property VALUES (20, 'password_tips', '密码长度在8~20位之间，且必须包含小写字母、大写字母、数字、特殊字符（+、-、*、/）。', 1, '2023-12-21 18:31:23');
INSERT INTO ssodb.t_system_property VALUES (21, 'password_tips_en', '密码长度在8~20位之间，且必须包含小写字母、大写字母、数字、特殊字符（+、-、*、/）。', 1, '2023-12-21 18:31:24');
INSERT INTO ssodb.t_system_property VALUES (6, 'access_policy_api_regex', '.', 1, '2023-04-20 14:03:48.68501');
INSERT INTO ssodb.t_system_property VALUES (7, 'access_policy_api_type', 'A02', 1, '2023-04-20 14:03:48.68501');


--
-- Data for Name: t_system_resource; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_tag; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_tag_filter; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_tag_user; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_task; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_user; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--

INSERT INTO ssodb.t_user VALUES (1, 'ssotest', '测试账号', '1', '010101200001010011', '["/000000/"]', '/L01/', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'ssotest', '0', NULL, NULL, 'admin', '2024-02-15 00:48:54.639014', '0');


--
-- Data for Name: t_user_attribute; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_user_conf_notification; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--

INSERT INTO ssodb.t_user_conf_notification VALUES (1, NULL, '{"0":0}', '{"crack":0,"weakpass":0}', '{"password":0,"phone":0,"email":0}', 'default', '2023-04-14 02:48:30.136614');


--
-- Data for Name: t_user_image; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_user_role; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Data for Name: t_user_surrogate; Type: TABLE DATA; Schema: ssodb; Owner: ssouser
--



--
-- Name: t_account_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_account_id_seq', 1, false);


--
-- Name: t_account_totp_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_account_totp_id_seq', 1, false);


--
-- Name: t_admin_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_admin_id_seq', 1, false);


--
-- Name: t_app_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_app_id_seq', 1, false);


--
-- Name: t_app_token_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_app_token_id_seq', 1, false);


--
-- Name: t_attribute_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_attribute_id_seq', 7, true);


--
-- Name: t_code_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_code_id_seq', 1, false);


--
-- Name: t_code_userrole_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_code_userrole_id_seq', 1, false);


--
-- Name: t_export_job_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_export_job_id_seq', 1, false);


--
-- Name: t_file_baseinfo_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_file_baseinfo_id_seq', 1, false);


--
-- Name: t_import_job_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_import_job_id_seq', 1, false);


--
-- Name: t_log_admin_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_log_admin_id_seq', 2, true);


--
-- Name: t_log_app_verify_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_log_app_verify_id_seq', 1, false);


--
-- Name: t_log_audit_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_log_audit_id_seq', 2, true);


--
-- Name: t_log_auth_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_log_auth_id_seq', 2, true);


--
-- Name: t_log_online_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_log_online_id_seq', 1, false);


--
-- Name: t_log_task_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_log_task_id_seq', 1, false);


--
-- Name: t_log_trusted_device_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_log_trusted_device_id_seq', 1, false);


--
-- Name: t_news_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_news_id_seq', 1, false);


--
-- Name: t_service_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_service_id_seq', 1, false);


--
-- Name: t_service_ldap_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_service_ldap_id_seq', 1, false);


--
-- Name: t_system_across_region_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_system_across_region_id_seq', 1, false);


--
-- Name: t_system_password_weak_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_system_password_weak_id_seq', 1, false);


--
-- Name: t_system_property_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_system_property_id_seq', 1, false);


--
-- Name: t_system_resource_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_system_resource_id_seq', 1, false);


--
-- Name: t_tag_filter_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_tag_filter_id_seq', 1, false);


--
-- Name: t_tag_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_tag_id_seq', 1, false);


--
-- Name: t_tag_user_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_tag_user_id_seq', 1, false);


--
-- Name: t_task_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_task_id_seq', 1, false);


--
-- Name: t_user_conf_notification_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_user_conf_notification_id_seq', 1, false);


--
-- Name: t_user_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_user_id_seq', 1, false);


--
-- Name: t_user_image_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_user_image_id_seq', 1, false);


--
-- Name: t_user_role_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_user_role_id_seq', 1, false);


--
-- Name: t_user_surrogate_id_seq; Type: SEQUENCE SET; Schema: ssodb; Owner: ssouser
--

SELECT pg_catalog.setval('ssodb.t_user_surrogate_id_seq', 1, false);


--
-- Name: t_account t_account_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_account
    ADD CONSTRAINT t_account_pk PRIMARY KEY (id);


--
-- Name: t_account_totp t_account_totp_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_account_totp
    ADD CONSTRAINT t_account_totp_pk PRIMARY KEY (id);


--
-- Name: t_admin t_admin_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_admin
    ADD CONSTRAINT t_admin_pk PRIMARY KEY (id);


--
-- Name: t_app t_app_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_app
    ADD CONSTRAINT t_app_pk PRIMARY KEY (id);


--
-- Name: t_app_token t_app_token_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_app_token
    ADD CONSTRAINT t_app_token_pk PRIMARY KEY (id);


--
-- Name: t_attribute t_attribute_pkey; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_attribute
    ADD CONSTRAINT t_attribute_pkey PRIMARY KEY (id);


--
-- Name: t_code_department t_code_department_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_code_department
    ADD CONSTRAINT t_code_department_pk PRIMARY KEY (id);


--
-- Name: t_code t_code_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_code
    ADD CONSTRAINT t_code_pk PRIMARY KEY (id);


--
-- Name: t_code_userrole t_code_userrole_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_code_userrole
    ADD CONSTRAINT t_code_userrole_pk PRIMARY KEY (id);


--
-- Name: t_code_usertype t_code_usertype_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_code_usertype
    ADD CONSTRAINT t_code_usertype_pk PRIMARY KEY (id);


--
-- Name: t_export_job t_export_job_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_export_job
    ADD CONSTRAINT t_export_job_pk PRIMARY KEY (id);


--
-- Name: t_file_baseinfo t_file_baseinfo_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_file_baseinfo
    ADD CONSTRAINT t_file_baseinfo_pk PRIMARY KEY (id);


--
-- Name: t_import_job t_import_job_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_import_job
    ADD CONSTRAINT t_import_job_pk PRIMARY KEY (id);


--
-- Name: t_log_admin t_log_admin_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_log_admin
    ADD CONSTRAINT t_log_admin_pk PRIMARY KEY (id);


--
-- Name: t_log_app_access t_log_app_access_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_log_app_access
    ADD CONSTRAINT t_log_app_access_pk PRIMARY KEY (uuid);


--
-- Name: t_log_app_verify t_log_app_verify_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_log_app_verify
    ADD CONSTRAINT t_log_app_verify_pk PRIMARY KEY (id);


--
-- Name: t_log_audit t_log_audit_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_log_audit
    ADD CONSTRAINT t_log_audit_pk PRIMARY KEY (id);


--
-- Name: t_log_online t_log_online_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_log_online
    ADD CONSTRAINT t_log_online_pk PRIMARY KEY (id);


--
-- Name: t_log_task t_log_task_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_log_task
    ADD CONSTRAINT t_log_task_pk PRIMARY KEY (id);


--
-- Name: t_log_trusted_device t_log_trusted_device_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_log_trusted_device
    ADD CONSTRAINT t_log_trusted_device_pk PRIMARY KEY (id);


--
-- Name: t_news t_news_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_news
    ADD CONSTRAINT t_news_pk PRIMARY KEY (id);


--
-- Name: t_service_ldap t_service_ldap_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_service_ldap
    ADD CONSTRAINT t_service_ldap_pk PRIMARY KEY (id);


--
-- Name: t_service t_service_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_service
    ADD CONSTRAINT t_service_pk PRIMARY KEY (id);


--
-- Name: t_system_across_region t_system_acrossregion_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_system_across_region
    ADD CONSTRAINT t_system_acrossregion_pk PRIMARY KEY (id);


--
-- Name: t_system_password_weak t_system_passwordweak_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_system_password_weak
    ADD CONSTRAINT t_system_passwordweak_pk PRIMARY KEY (id);


--
-- Name: t_system_property t_system_properties_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_system_property
    ADD CONSTRAINT t_system_properties_pk PRIMARY KEY (id);


--
-- Name: t_system_resource t_system_resources_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_system_resource
    ADD CONSTRAINT t_system_resources_pk PRIMARY KEY (id);


--
-- Name: t_tag_filter t_tag_filter_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_tag_filter
    ADD CONSTRAINT t_tag_filter_pk PRIMARY KEY (id);


--
-- Name: t_tag t_tag_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_tag
    ADD CONSTRAINT t_tag_pk PRIMARY KEY (id);


--
-- Name: t_tag_user t_tag_user_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_tag_user
    ADD CONSTRAINT t_tag_user_pk PRIMARY KEY (id);


--
-- Name: t_task t_task_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_task
    ADD CONSTRAINT t_task_pk PRIMARY KEY (id);


--
-- Name: t_user_attribute t_user_attribute_pkey; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_user_attribute
    ADD CONSTRAINT t_user_attribute_pkey PRIMARY KEY (userid);


--
-- Name: t_user_conf_notification t_user_conf_notification_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_user_conf_notification
    ADD CONSTRAINT t_user_conf_notification_pk PRIMARY KEY (id);


--
-- Name: t_user_image t_user_image_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_user_image
    ADD CONSTRAINT t_user_image_pk PRIMARY KEY (id);


--
-- Name: t_user t_user_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_user
    ADD CONSTRAINT t_user_pk PRIMARY KEY (id);


--
-- Name: t_user_role t_user_roles_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_user_role
    ADD CONSTRAINT t_user_roles_pk PRIMARY KEY (id);


--
-- Name: t_user_surrogate t_user_surrogate_pk; Type: CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_user_surrogate
    ADD CONSTRAINT t_user_surrogate_pk PRIMARY KEY (id);


--
-- Name: t_account_totp_accountid_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_account_totp_accountid_index ON ssodb.t_account_totp USING btree (accountid);


--
-- Name: t_account_userid_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_account_userid_index ON ssodb.t_account USING btree (userid);


--
-- Name: t_account_username_uindex; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE UNIQUE INDEX t_account_username_uindex ON ssodb.t_account USING btree (username);


--
-- Name: t_log_app_access_appid_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_log_app_access_appid_index ON ssodb.t_log_app_access USING btree (appid);


--
-- Name: t_log_auth_code_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_log_auth_code_index ON ONLY ssodb.t_log_auth USING btree (code);


--
-- Name: t_log_auth_created_date_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_log_auth_created_date_index ON ONLY ssodb.t_log_auth USING btree (created_date);


--
-- Name: t_log_auth_history_code_idx; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_log_auth_history_code_idx ON ssodb.t_log_auth_history USING btree (code);


--
-- Name: t_log_auth_history_created_date_idx; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_log_auth_history_created_date_idx ON ssodb.t_log_auth_history USING btree (created_date);


--
-- Name: t_log_auth_id_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_log_auth_id_index ON ONLY ssodb.t_log_auth USING btree (id);


--
-- Name: t_log_auth_history_id_idx; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_log_auth_history_id_idx ON ssodb.t_log_auth_history USING btree (id);


--
-- Name: t_log_auth_service_id_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_log_auth_service_id_index ON ONLY ssodb.t_log_auth USING btree (service_id);


--
-- Name: t_log_auth_history_service_id_idx; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_log_auth_history_service_id_idx ON ssodb.t_log_auth_history USING btree (service_id);


--
-- Name: t_log_auth_username_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_log_auth_username_index ON ONLY ssodb.t_log_auth USING btree (username);


--
-- Name: t_log_auth_history_username_idx; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_log_auth_history_username_idx ON ssodb.t_log_auth_history USING btree (username);


--
-- Name: t_log_online_ip_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_log_online_ip_index ON ssodb.t_log_online USING btree (ip);


--
-- Name: t_log_online_tgt_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_log_online_tgt_index ON ssodb.t_log_online USING btree (tgt);


--
-- Name: t_log_online_username_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_log_online_username_index ON ssodb.t_log_online USING btree (username);


--
-- Name: t_news_news_pictureid_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_news_news_pictureid_index ON ssodb.t_news USING btree (news_pictureid);


--
-- Name: t_service_ldap_username_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_service_ldap_username_index ON ssodb.t_service_ldap USING btree (username);


--
-- Name: t_tag_filterid_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_tag_filterid_index ON ssodb.t_tag USING btree (filterid);


--
-- Name: t_tag_user_tagid_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_tag_user_tagid_index ON ssodb.t_tag_user USING btree (tagid);


--
-- Name: t_tag_user_userid_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_tag_user_userid_index ON ssodb.t_tag_user USING btree (userid);


--
-- Name: t_user_admission_number_uindex; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE UNIQUE INDEX t_user_admission_number_uindex ON ssodb.t_user USING btree (admission_number);


--
-- Name: t_user_candidate_number_uindex; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE UNIQUE INDEX t_user_candidate_number_uindex ON ssodb.t_user USING btree (candidate_number);


--
-- Name: t_user_conf_notification_userid_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_user_conf_notification_userid_index ON ssodb.t_user_conf_notification USING btree (userid);


--
-- Name: t_user_ic_number_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_user_ic_number_index ON ssodb.t_user USING btree (ic_number);


--
-- Name: t_user_image_fileid_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_user_image_fileid_index ON ssodb.t_user_image USING btree (fileid);


--
-- Name: t_user_image_userid_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_user_image_userid_index ON ssodb.t_user_image USING btree (userid);


--
-- Name: t_user_name_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_user_name_index ON ssodb.t_user USING btree (name);


--
-- Name: t_user_role_roleid_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_user_role_roleid_index ON ssodb.t_user_role USING btree (roleid);


--
-- Name: t_user_role_userid_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_user_role_userid_index ON ssodb.t_user_role USING btree (userid);


--
-- Name: t_user_schoolid_uindex; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE UNIQUE INDEX t_user_schoolid_uindex ON ssodb.t_user USING btree (schoolid);


--
-- Name: t_user_surrogate_schoolid_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_user_surrogate_schoolid_index ON ssodb.t_user_surrogate USING btree (schoolid);


--
-- Name: t_user_surrogate_surrogate_schoolid_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_user_surrogate_surrogate_schoolid_index ON ssodb.t_user_surrogate USING btree (surrogate_schoolid);


--
-- Name: t_user_type_index; Type: INDEX; Schema: ssodb; Owner: ssouser
--

CREATE INDEX t_user_type_index ON ssodb.t_user USING btree (type);


--
-- Name: t_log_auth_history_code_idx; Type: INDEX ATTACH; Schema: ssodb; Owner: ssouser
--

ALTER INDEX ssodb.t_log_auth_code_index ATTACH PARTITION ssodb.t_log_auth_history_code_idx;


--
-- Name: t_log_auth_history_created_date_idx; Type: INDEX ATTACH; Schema: ssodb; Owner: ssouser
--

ALTER INDEX ssodb.t_log_auth_created_date_index ATTACH PARTITION ssodb.t_log_auth_history_created_date_idx;


--
-- Name: t_log_auth_history_id_idx; Type: INDEX ATTACH; Schema: ssodb; Owner: ssouser
--

ALTER INDEX ssodb.t_log_auth_id_index ATTACH PARTITION ssodb.t_log_auth_history_id_idx;


--
-- Name: t_log_auth_history_service_id_idx; Type: INDEX ATTACH; Schema: ssodb; Owner: ssouser
--

ALTER INDEX ssodb.t_log_auth_service_id_index ATTACH PARTITION ssodb.t_log_auth_history_service_id_idx;


--
-- Name: t_log_auth_history_username_idx; Type: INDEX ATTACH; Schema: ssodb; Owner: ssouser
--

ALTER INDEX ssodb.t_log_auth_username_index ATTACH PARTITION ssodb.t_log_auth_history_username_idx;


--
-- Name: t_account t_account_t_user_id_fk; Type: FK CONSTRAINT; Schema: ssodb; Owner: ssouser
--

ALTER TABLE ONLY ssodb.t_account
    ADD CONSTRAINT t_account_t_user_id_fk FOREIGN KEY (userid) REFERENCES ssodb.t_user(id);


--
-- Name: SCHEMA ssodb; Type: ACL; Schema: -; Owner: postgres
--

GRANT ALL ON SCHEMA ssodb TO ssouser;


--
-- Name: DEFAULT PRIVILEGES FOR SEQUENCES; Type: DEFAULT ACL; Schema: ssodb; Owner: postgres
--

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA ssodb GRANT ALL ON SEQUENCES TO ssouser;


--
-- Name: DEFAULT PRIVILEGES FOR FUNCTIONS; Type: DEFAULT ACL; Schema: ssodb; Owner: postgres
--

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA ssodb GRANT ALL ON FUNCTIONS TO ssouser;


--
-- Name: DEFAULT PRIVILEGES FOR TABLES; Type: DEFAULT ACL; Schema: ssodb; Owner: postgres
--

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA ssodb GRANT ALL ON TABLES TO ssouser;


--
-- PostgreSQL database dump complete
--

