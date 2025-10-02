-- Downloaded from: https://github.com/jeffchang001/ee-midd/blob/888fbe7137741eefd67414ab9d580964282d1256/sql/schema.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 14.13
-- Dumped by pg_dump version 14.13

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
SET search_path TO public, pg_catalog;

--
-- Name: find_approval_manager(character varying, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.find_approval_manager(p_employee_no character varying, p_tree_type character varying DEFAULT '2'::character varying) RETURNS TABLE(manager_employee_no character varying, manager_full_name character varying, manager_org_code character varying, manager_org_name character varying)
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_org_code VARCHAR;
    v_current_org_code VARCHAR;  -- 用來儲存當前檢查的組織編號
    v_employee_no VARCHAR;
    v_manager_employee_no VARCHAR;
    v_manager_full_name VARCHAR;
    v_manager_org_code VARCHAR;
    v_manager_org_name VARCHAR;
BEGIN
    -- Find the employee's organization code based on the tree type
    SELECT 
        CASE 
            WHEN p_tree_type = '0' THEN formula_org_code
            WHEN p_tree_type = '1' THEN function_org_code
            WHEN p_tree_type = '2' THEN form_org_code
        END
    INTO v_org_code
    FROM api_employee_info
    WHERE employee_no = p_employee_no;

    -- 儲存原始組織編號
    v_current_org_code := v_org_code;

    -- 使用 LOOP 持續往上尋找直到找到不同組織的主管
    LOOP
        -- Find the manager of the current department
        SELECT aom.employee_no, aom.full_name, aom.org_code, ao.org_name
        INTO v_manager_employee_no, v_manager_full_name, v_manager_org_code, v_manager_org_name
        FROM api_organization_manager aom
        JOIN api_organization ao ON aom.org_code = ao.org_code
        WHERE aom.org_code = v_current_org_code
        LIMIT 1;

        -- 如果找不到主管，退出循環
        IF v_manager_employee_no IS NULL THEN
            EXIT;
        END IF;

        -- 如果找到的主管不是自己 且 組織編號與原始組織不同，退出循環
        IF v_manager_employee_no != p_employee_no AND v_manager_org_code != v_org_code THEN
            EXIT;
        END IF;

        -- 找上層組織編號
        SELECT aor.parent_org_code
        INTO v_current_org_code
        FROM api_organization_relation aor
        WHERE aor.org_code = v_current_org_code
        AND aor.org_tree_type = p_tree_type;

        -- 如果找不到上層組織，退出循環
        IF v_current_org_code IS NULL THEN
            EXIT;
        END IF;
    END LOOP;

    -- Return the result
    RETURN QUERY SELECT v_manager_employee_no, v_manager_full_name, v_manager_org_code, v_manager_org_name;
END;
$$;


ALTER FUNCTION public.find_approval_manager(p_employee_no character varying, p_tree_type character varying) OWNER TO postgres;

--
-- Name: get_actived_org_tree(character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_actived_org_tree(p_org_tree_type character varying) RETURNS TABLE(id bigint, company_code character varying, company_party_id bigint, created_date timestamp with time zone, data_created_date timestamp with time zone, data_created_user character varying, data_modified_date timestamp with time zone, data_modified_user character varying, org_code character varying, org_name character varying, org_tree_type character varying, organization_relation_id bigint, parent_org_code character varying, status character varying, tenant_id character varying)
    LANGUAGE sql
    AS $_$
SELECT r.*
FROM public.api_organization o
JOIN public.api_organization_relation r 
    ON o.org_code = r.org_code
WHERE 
    -- 若為空字串，則不加任何 org_tree_type 的限制
    ($1 = '' OR r.org_tree_type = $1);
$_$;


ALTER FUNCTION public.get_actived_org_tree(p_org_tree_type character varying) OWNER TO postgres;

--
-- Name: get_employee_no_by_org_hierarchy(character varying, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_employee_no_by_org_hierarchy(p_employee_no character varying, p_org_tree_type character varying) RETURNS TABLE(employee_no character varying, full_name character varying, english_name character varying, email_address character varying, hire_date timestamp with time zone, job_title_name character varying, formula_org_code character varying, formula_org_name character varying, org_code character varying, org_name character varying, parent_org_code character varying, org_level integer)
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_formula_org_code VARCHAR(20);
BEGIN
    -- 首先獲取員工的 formula_org_code
    SELECT ei.formula_org_code INTO v_formula_org_code
    FROM public.api_employee_info ei
    WHERE ei.employee_no = p_employee_no;

    -- 如果找不到員工，返回空結果
    IF v_formula_org_code IS NULL THEN
        RETURN;
    END IF;

    RETURN QUERY
    WITH RECURSIVE org_hierarchy AS (
        -- 基礎案例：從員工的 formula_org_code 開始
        SELECT 
            ar.org_code,
            ar.org_name,
            ar.parent_org_code,
            1 AS org_level
        FROM public.api_organization_relation ar
        WHERE ar.org_code = v_formula_org_code AND ar.org_tree_type = p_org_tree_type  --應該是不同 org type 要參照不同個 員工欄位

        UNION ALL

        -- 遞迴案例：查找上層組織
        SELECT 
            ar.org_code,
            ar.org_name,
            ar.parent_org_code,
            oh.org_level + 1
        FROM public.api_organization_relation ar
        JOIN org_hierarchy oh ON ar.org_code = oh.parent_org_code
        WHERE ar.org_tree_type = p_org_tree_type
    )
    SELECT 
        ei.employee_no,
        ei.full_name,
        ei.english_name,
        ei.email_address,
        ei.hire_date,
        ei.job_title_name,
        ei.formula_org_code,
        ei.formula_org_name,
        oh.org_code,
        oh.org_name,
        oh.parent_org_code,
        oh.org_level
    FROM 
        public.api_employee_info ei
    CROSS JOIN org_hierarchy oh
    WHERE 
        ei.employee_no = p_employee_no
    ORDER BY 
        oh.org_level;
END;
$$;


ALTER FUNCTION public.get_employee_no_by_org_hierarchy(p_employee_no character varying, p_org_tree_type character varying) OWNER TO postgres;

--
-- Name: get_employees_by_org_code(character varying, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_employees_by_org_code(p_org_code character varying, p_org_tree_type character varying) RETURNS TABLE(id bigint, arcidno character varying, company_code character varying, company_name character varying, company_partyid bigint, country_code character varying, created_date timestamp with time zone, data_created_date timestamp with time zone, data_created_user character varying, data_modified_date timestamp with time zone, data_modified_user character varying, date_of_birth timestamp with time zone, email_address character varying, employed_status character varying, employee_no character varying, employee_type_code character varying, employee_type_name character varying, english_name character varying, ext_no character varying, form_org_code character varying, form_org_name character varying, form_org_partyid bigint, formula_org_code character varying, formula_org_name character varying, formula_org_partyid bigint, full_name character varying, gender_code character varying, gender_name character varying, hire_date timestamp with time zone, idno character varying, job_flag character varying, job_grade_code character varying, job_grade_name character varying, job_level_code character varying, job_level_name character varying, job_title_code character varying, job_title_name character varying, mobile_phone_no character varying, office_phone character varying, party_roleid bigint, passport_no character varying, permanent_address character varying, permanent_phone_no character varying, permanent_zip_code character varying, position_code character varying, position_name character varying, present_address character varying, present_phone_no character varying, present_zip_code character varying, resignation_date timestamp with time zone, status character varying, tenantid character varying, userid character varying, function_org_code character varying, function_org_name character varying, mvpn character varying, idno_suffix character varying)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ei.id,
        ei.arcidno,
        ei.company_code,
        ei.company_name,
        ei.company_partyid,
        ei.country_code,
        ei.created_date::TIMESTAMPTZ,
        ei.data_created_date::TIMESTAMPTZ,
        ei.data_created_user,
        ei.data_modified_date::TIMESTAMPTZ,
        ei.data_modified_user,
        ei.date_of_birth::TIMESTAMPTZ,
        ei.email_address,
        ei.employed_status,
        ei.employee_no,
        ei.employee_type_code,
        ei.employee_type_name,
        ei.english_name,
        ei.ext_no,
        ei.form_org_code,
        ei.form_org_name,
        ei.form_org_partyid,
        ei.formula_org_code,
        ei.formula_org_name,
        ei.formula_org_partyid,
        ei.full_name,
        ei.gender_code,
        ei.gender_name,
        ei.hire_date::TIMESTAMPTZ,
        ei.idno,
		ei.job_flag,
        ei.job_grade_code,
        ei.job_grade_name,
        ei.job_level_code,
        ei.job_level_name,
        ei.job_title_code,
        ei.job_title_name,
        ei.mobile_phone_no,
        ei.office_phone,
        ei.party_roleid,
        ei.passport_no,
        ei.permanent_address,
        ei.permanent_phone_no,
        ei.permanent_zip_code,
        ei.position_code,
        ei.position_name,
        ei.present_address,
        ei.present_phone_no,
        ei.present_zip_code,
        ei.resignation_date::TIMESTAMPTZ,
        ei.status,
        ei.tenantid,
        ei.userid,
        ei.function_org_code,
        ei.function_org_name,
        ei.mvpn,
        ei.idno_suffix
    FROM public.api_employee_info ei
    JOIN public.api_organization_relation ar ON
        CASE
            WHEN p_org_tree_type = '0' THEN ei.formula_org_code = ar.org_code
            WHEN p_org_tree_type = '1' THEN ei.function_org_code = ar.org_code
            WHEN p_org_tree_type = '2' THEN ei.form_org_code = ar.org_code
            ELSE ei.formula_org_code = ar.org_code  -- 默認使用 formula_org_code
        END
    WHERE ar.org_code = p_org_code AND ar.org_tree_type = p_org_tree_type AND ei.employed_status = '1';
END;
$$;


ALTER FUNCTION public.get_employees_by_org_code(p_org_code character varying, p_org_tree_type character varying) OWNER TO postgres;

--
-- Name: get_materialized_view_changes_by_date(date, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_materialized_view_changes_by_date(p_date date, p_view_name character varying DEFAULT NULL::character varying) RETURNS TABLE(view_name character varying, refreshed_at timestamp with time zone, diff_count integer, diff_details jsonb)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        mvrl.view_name,
        mvrl.refreshed_at,
        mvrl.diff_count,
        mvrl.diff_details
    FROM 
        materialized_view_refresh_log mvrl
    WHERE 
        DATE(mvrl.refreshed_at) = p_date
        AND (p_view_name IS NULL OR mvrl.view_name = p_view_name)
    ORDER BY 
        mvrl.refreshed_at;
END;
$$;


ALTER FUNCTION public.get_materialized_view_changes_by_date(p_date date, p_view_name character varying) OWNER TO postgres;

--
-- Name: get_org_hierarchy_by_org_code(character varying, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_org_hierarchy_by_org_code(p_org_code character varying, p_org_tree_type character varying) RETURNS TABLE(org_code character varying, org_name character varying, org_tree_type character varying, parent_org_code character varying, level integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    WITH RECURSIVE org_hierarchy AS (
        -- 基礎案例：從輸入的 org_code 開始
        SELECT 
            ar.org_code,
            ar.org_name,
            ar.org_tree_type,
            ar.parent_org_code,
            1 AS level
        FROM public.api_organization_relation ar
        WHERE ar.org_code = p_org_code AND ar.org_tree_type = p_org_tree_type

        UNION ALL

        -- 遞迴案例：查找上層組織
        SELECT 
            ar.org_code,
            ar.org_name,
            ar.org_tree_type,
            ar.parent_org_code,
            oh.level + 1
        FROM public.api_organization_relation ar
        JOIN org_hierarchy oh ON ar.org_code = oh.parent_org_code
        WHERE ar.org_tree_type = p_org_tree_type
    )
    SELECT * FROM org_hierarchy
    ORDER BY level;
END;
$$;


ALTER FUNCTION public.get_org_hierarchy_by_org_code(p_org_code character varying, p_org_tree_type character varying) OWNER TO postgres;

--
-- Name: refresh_all_materialized_views(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.refresh_all_materialized_views() RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_views TEXT[] := ARRAY['fse7en_org_deptgradeinfo', 'fse7en_org_deptinfo', 'fse7en_org_deptstruct', 
                           'fse7en_org_jobtitle2grade', 'fse7en_org_memberinfo', 'fse7en_org_memberstruct'];
    v_view TEXT;
BEGIN
    FOREACH v_view IN ARRAY v_views
    LOOP
        PERFORM refresh_materialized_view_with_log(v_view);
    END LOOP;
END;
$$;


ALTER FUNCTION public.refresh_all_materialized_views() OWNER TO postgres;

--
-- Name: refresh_materialized_view_with_log(character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.refresh_materialized_view_with_log(p_view_name character varying) RETURNS void
    LANGUAGE plpgsql
    AS $_$
DECLARE
    v_old_data JSONB;
    v_new_data JSONB;
    v_diff JSONB;
    v_diff_count INT;
    v_snapshot_table TEXT;
    v_row_identifiers TEXT[];  -- 改為陣列，支援複合主鍵
    v_compare_condition TEXT;
    v_key_fields TEXT;
    v_key_json_old TEXT;
    v_key_json_new TEXT;
    i INT;
BEGIN
    -- 建立臨時表存儲更新前的數據快照
    v_snapshot_table := 'tmp_' || p_view_name || '_snapshot';
    
    -- 根據視圖名稱確定主鍵或唯一識別欄位（支援複合鍵）
    CASE lower(p_view_name)
        WHEN 'fse7en_org_deptgradeinfo' THEN v_row_identifiers := ARRAY['grade_id'];
        WHEN 'fse7en_org_deptinfo' THEN v_row_identifiers := ARRAY['org_code'];
        WHEN 'fse7en_org_deptstruct' THEN v_row_identifiers := ARRAY['org_code'];
        WHEN 'fse7en_org_jobtitle2grade' THEN v_row_identifiers := ARRAY['job_title_code'];
        WHEN 'fse7en_org_memberinfo' THEN v_row_identifiers := ARRAY['employee_no'];
        WHEN 'fse7en_org_memberstruct' THEN v_row_identifiers := ARRAY['employee_no', 'org_code'];  -- 複合主鍵
        ELSE v_row_identifiers := NULL;
    END CASE;
    
    -- 建立臨時表
    EXECUTE format('CREATE TEMP TABLE %I ON COMMIT DROP AS SELECT * FROM %I', v_snapshot_table, p_view_name);
    
    -- 更新實體化視圖
    EXECUTE format('REFRESH MATERIALIZED VIEW %I', p_view_name);
    
    -- 為複合主鍵構建比較條件
    IF v_row_identifiers IS NOT NULL THEN
        v_compare_condition := '';
        v_key_fields := '';
        v_key_json_old := '';
        v_key_json_new := '';
        
        FOR i IN 1..array_length(v_row_identifiers, 1) LOOP
            -- 構建比較條件
            IF i > 1 THEN
                v_compare_condition := v_compare_condition || ' AND ';
                v_key_fields := v_key_fields || ', ';
                v_key_json_old := v_key_json_old || ', ';
                v_key_json_new := v_key_json_new || ', ';
            END IF;
            
            v_compare_condition := v_compare_condition || format('o.%1$I = n.%1$I', v_row_identifiers[i]);
            v_key_fields := v_key_fields || format('o.%1$I', v_row_identifiers[i]);
            v_key_json_old := v_key_json_old || format('''%1$I'', d.old_data->>''%1$I''', v_row_identifiers[i]);
            v_key_json_new := v_key_json_new || format('''%1$I'', d.new_data->>''%1$I''', v_row_identifiers[i]);
        END LOOP;
    END IF;
    
    -- 計算差異
    IF v_row_identifiers IS NULL THEN
        -- 沒有唯一識別欄位，使用整行比較
        EXECUTE format('
            WITH diff_data AS (
                -- 刪除的記錄
                SELECT
                    ''DELETE'' as operation,
                    row_to_json(o)::jsonb as old_data,
                    NULL::jsonb as new_data
                FROM
                    %I o
                WHERE
                    NOT EXISTS (
                        SELECT 1 FROM %I n
                        WHERE row_to_json(o)::jsonb = row_to_json(n)::jsonb
                    )
                    
                UNION ALL
                
                -- 插入的記錄
                SELECT
                    ''INSERT'' as operation,
                    NULL::jsonb as old_data,
                    row_to_json(n)::jsonb as new_data
                FROM
                    %I n
                WHERE
                    NOT EXISTS (
                        SELECT 1 FROM %I o
                        WHERE row_to_json(o)::jsonb = row_to_json(n)::jsonb
                    )
            )
            SELECT
                jsonb_agg(d.old_data) FILTER (WHERE d.operation = ''DELETE''),
                jsonb_agg(d.new_data) FILTER (WHERE d.operation = ''INSERT''),
                jsonb_agg(
                    CASE
                        WHEN d.operation = ''DELETE'' THEN jsonb_build_object(''operation'', d.operation, ''data'', d.old_data)
                        WHEN d.operation = ''INSERT'' THEN jsonb_build_object(''operation'', d.operation, ''data'', d.new_data)
                    END
                ),
                COUNT(*)
            FROM
                diff_data d',
            v_snapshot_table,
            p_view_name,
            p_view_name,
            v_snapshot_table
        )
        INTO v_old_data, v_new_data, v_diff, v_diff_count;
    ELSE
        -- 使用複合鍵進行比較
        EXECUTE format('
            WITH diff_data AS (
                -- 刪除的記錄
                SELECT
                    ''DELETE'' as operation,
                    row_to_json(o)::jsonb as old_data,
                    NULL::jsonb as new_data
                FROM
                    %I o
                WHERE
                    NOT EXISTS (
                        SELECT 1 FROM %I n
                        WHERE %s
                    )
                    
                UNION ALL
                
                -- 插入的記錄
                SELECT
                    ''INSERT'' as operation,
                    NULL::jsonb as old_data,
                    row_to_json(n)::jsonb as new_data
                FROM
                    %I n
                WHERE
                    NOT EXISTS (
                        SELECT 1 FROM %I o
                        WHERE %s
                    )
                    
                UNION ALL
                
                -- 更新的記錄
                SELECT
                    ''UPDATE'' as operation,
                    row_to_json(o)::jsonb as old_data,
                    row_to_json(n)::jsonb as new_data
                FROM
                    %I o
                JOIN
                    %I n ON %s
                WHERE
                    row_to_json(o)::jsonb <> row_to_json(n)::jsonb
            )
            SELECT
                jsonb_agg(d.old_data) FILTER (WHERE d.operation IN (''DELETE'', ''UPDATE'')),
                jsonb_agg(d.new_data) FILTER (WHERE d.operation IN (''INSERT'', ''UPDATE'')),
                jsonb_agg(
                    CASE
                        WHEN d.operation = ''DELETE'' THEN jsonb_build_object(''operation'', d.operation, ''key'', jsonb_build_object(%s), ''data'', d.old_data)
                        WHEN d.operation = ''INSERT'' THEN jsonb_build_object(''operation'', d.operation, ''key'', jsonb_build_object(%s), ''data'', d.new_data)
                        ELSE jsonb_build_object(''operation'', d.operation, ''key'', jsonb_build_object(%s), ''old'', d.old_data, ''new'', d.new_data)
                    END
                ),
                COUNT(*)
            FROM
                diff_data d',
            v_snapshot_table,
            p_view_name,
            v_compare_condition,
            p_view_name,
            v_snapshot_table,
            v_compare_condition,
            v_snapshot_table,
            p_view_name,
            v_compare_condition,
            v_key_json_old,
            v_key_json_new,
            v_key_json_old
        )
        INTO v_old_data, v_new_data, v_diff, v_diff_count;
    END IF;
    
    -- 記錄變更
    INSERT INTO materialized_view_refresh_log(view_name, refreshed_at, diff_count, diff_details)
    VALUES (p_view_name, CURRENT_TIMESTAMP, COALESCE(v_diff_count, 0), v_diff);
    
END;
$_$;


ALTER FUNCTION public.refresh_materialized_view_with_log(p_view_name character varying) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: api_company; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.api_company (
    id bigint NOT NULL,
    company_code character varying(20),
    company_name character varying(50),
    company_partyid bigint,
    created_date timestamp with time zone,
    data_created_date timestamp with time zone,
    data_created_user character varying(30),
    data_modified_date timestamp with time zone,
    data_modified_user character varying(30),
    effective_date timestamp with time zone,
    status character varying(2),
    tenantid character varying(20)
);


ALTER TABLE public.api_company OWNER TO postgres;

--
-- Name: api_company_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.api_company_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.api_company_id_seq OWNER TO postgres;

--
-- Name: api_company_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.api_company_id_seq OWNED BY public.api_company.id;


--
-- Name: api_employee_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.api_employee_info (
    id bigint NOT NULL,
    arcidno character varying(20),
    company_code character varying(20),
    company_name character varying(50),
    company_partyid bigint,
    country_code character varying(10),
    created_date timestamp with time zone,
    data_created_date timestamp with time zone,
    data_created_user character varying(30),
    data_modified_date timestamp with time zone,
    data_modified_user character varying(30),
    date_of_birth timestamp with time zone,
    email_address character varying(50),
    employed_status character varying(2),
    employee_no character varying(10),
    employee_type_code character varying(10),
    employee_type_name character varying(20),
    english_name character varying(50),
    ext_no character varying(10),
    form_org_code character varying(20),
    form_org_name character varying(50),
    form_org_partyid bigint,
    formula_org_code character varying(20),
    formula_org_name character varying(50),
    formula_org_partyid bigint,
    full_name character varying(50),
    function_org_code character varying(20),
    function_org_name character varying(50),
    gender_code character varying(2),
    gender_name character varying(10),
    hire_date timestamp with time zone,
    idno character varying(20),
    idno_suffix character varying(10),
    job_flag character varying(2),
    job_grade_code character varying(10),
    job_grade_name character varying(20),
    job_level_code character varying(10),
    job_level_name character varying(20),
    job_title_code character varying(10),
    job_title_name character varying(20),
    mobile_phone_no character varying(20),
    mvpn character varying(20),
    office_phone character varying(20),
    party_roleid bigint,
    passport_no character varying(30),
    permanent_address character varying(50),
    permanent_phone_no character varying(20),
    permanent_zip_code character varying(10),
    position_code character varying(10),
    position_name character varying(20),
    present_address character varying(50),
    present_phone_no character varying(20),
    present_zip_code character varying(10),
    resignation_date timestamp with time zone,
    status character varying(2),
    tenantid character varying(10),
    userid character varying(10)
);


ALTER TABLE public.api_employee_info OWNER TO postgres;

--
-- Name: api_employee_info_action_log; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.api_employee_info_action_log (
    id bigint NOT NULL,
    action character varying(10),
    action_date timestamp with time zone,
    created_date timestamp with time zone NOT NULL,
    employee_no character varying(10) NOT NULL,
    field_name character varying(30) NOT NULL,
    new_value character varying(50),
    old_value character varying(50),
    party_roleid bigint
);


ALTER TABLE public.api_employee_info_action_log OWNER TO postgres;

--
-- Name: api_employee_info_action_log_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.api_employee_info_action_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.api_employee_info_action_log_id_seq OWNER TO postgres;

--
-- Name: api_employee_info_action_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.api_employee_info_action_log_id_seq OWNED BY public.api_employee_info_action_log.id;


--
-- Name: api_employee_info_archived; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.api_employee_info_archived (
    id bigint NOT NULL,
    arcidno character varying(20),
    company_code character varying(20),
    company_name character varying(50),
    company_partyid bigint,
    country_code character varying(10),
    created_date timestamp with time zone,
    data_created_date timestamp with time zone,
    data_created_user character varying(30),
    data_modified_date timestamp with time zone,
    data_modified_user character varying(30),
    date_of_birth timestamp with time zone,
    email_address character varying(50),
    employed_status character varying(2),
    employee_no character varying(10),
    employee_type_code character varying(10),
    employee_type_name character varying(20),
    english_name character varying(50),
    ext_no character varying(10),
    form_org_code character varying(20),
    form_org_name character varying(50),
    form_org_partyid bigint,
    formula_org_code character varying(20),
    formula_org_name character varying(50),
    formula_org_partyid bigint,
    full_name character varying(20),
    function_org_code character varying(20),
    function_org_name character varying(50),
    gender_code character varying(2),
    gender_name character varying(10),
    hire_date timestamp with time zone,
    idno character varying(20),
    idno_suffix character varying(10),
    job_flag character varying(2),
    job_grade_code character varying(10),
    job_grade_name character varying(20),
    job_level_code character varying(10),
    job_level_name character varying(20),
    job_title_code character varying(10),
    job_title_name character varying(20),
    mobile_phone_no character varying(20),
    mvpn character varying(20),
    office_phone character varying(20),
    party_roleid bigint,
    passport_no character varying(30),
    permanent_address character varying(50),
    permanent_phone_no character varying(20),
    permanent_zip_code character varying(10),
    position_code character varying(10),
    position_name character varying(20),
    present_address character varying(50),
    present_phone_no character varying(20),
    present_zip_code character varying(10),
    resignation_date timestamp with time zone,
    status character varying(2),
    tenantid character varying(10),
    userid character varying(10)
);


ALTER TABLE public.api_employee_info_archived OWNER TO postgres;

--
-- Name: api_employee_info_archived_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.api_employee_info_archived_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.api_employee_info_archived_id_seq OWNER TO postgres;

--
-- Name: api_employee_info_archived_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.api_employee_info_archived_id_seq OWNED BY public.api_employee_info_archived.id;


--
-- Name: api_employee_info_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.api_employee_info_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.api_employee_info_id_seq OWNER TO postgres;

--
-- Name: api_employee_info_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.api_employee_info_id_seq OWNED BY public.api_employee_info.id;


--
-- Name: api_organization; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.api_organization (
    id bigint NOT NULL,
    address character varying(50),
    company_code character varying(20),
    company_name character varying(50),
    company_partyid bigint,
    created_date timestamp with time zone,
    data_created_date timestamp with time zone,
    data_created_user character varying(30),
    data_modified_date timestamp with time zone,
    data_modified_user character varying(30),
    email character varying(50),
    end_date timestamp with time zone,
    english_name character varying(50),
    fax character varying(10),
    layer_code character varying(5),
    layer_name character varying(20),
    org_code character varying(20),
    org_name character varying(50),
    org_property_code character varying(2),
    organizationid bigint,
    remark character varying(20),
    sort_sequence integer,
    start_date timestamp with time zone,
    status character varying(2),
    telephone character varying(20),
    tenantid character varying(10)
);


ALTER TABLE public.api_organization OWNER TO postgres;

--
-- Name: api_organization_action_log; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.api_organization_action_log (
    id bigint NOT NULL,
    action character varying(10),
    action_date timestamp with time zone,
    created_date timestamp with time zone NOT NULL,
    field_name character varying(30) NOT NULL,
    new_value character varying(50),
    old_value character varying(50),
    org_code character varying(20) NOT NULL,
    organizationid bigint
);


ALTER TABLE public.api_organization_action_log OWNER TO postgres;

--
-- Name: api_organization_action_log_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.api_organization_action_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.api_organization_action_log_id_seq OWNER TO postgres;

--
-- Name: api_organization_action_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.api_organization_action_log_id_seq OWNED BY public.api_organization_action_log.id;


--
-- Name: api_organization_archived; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.api_organization_archived (
    id bigint NOT NULL,
    address character varying(50),
    company_code character varying(20),
    company_name character varying(50),
    company_partyid bigint,
    created_date timestamp with time zone,
    data_created_date timestamp with time zone,
    data_created_user character varying(30),
    data_modified_date timestamp with time zone,
    data_modified_user character varying(30),
    email character varying(50),
    end_date timestamp with time zone,
    english_name character varying(50),
    fax character varying(10),
    layer_code character varying(5),
    layer_name character varying(20),
    org_code character varying(20),
    org_name character varying(50),
    org_property_code character varying(2),
    organizationid bigint,
    remark character varying(20),
    sort_sequence integer,
    start_date timestamp with time zone,
    status character varying(2),
    telephone character varying(20),
    tenantid character varying(10)
);


ALTER TABLE public.api_organization_archived OWNER TO postgres;

--
-- Name: api_organization_archived_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.api_organization_archived_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.api_organization_archived_id_seq OWNER TO postgres;

--
-- Name: api_organization_archived_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.api_organization_archived_id_seq OWNED BY public.api_organization_archived.id;


--
-- Name: api_organization_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.api_organization_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.api_organization_id_seq OWNER TO postgres;

--
-- Name: api_organization_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.api_organization_id_seq OWNED BY public.api_organization.id;


--
-- Name: api_organization_manager; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.api_organization_manager (
    id bigint NOT NULL,
    company_code character varying(20),
    company_partyid bigint,
    created_date timestamp with time zone,
    data_created_date timestamp with time zone,
    data_created_user character varying(30),
    data_modified_date timestamp with time zone,
    data_modified_user character varying(30),
    effective_date timestamp with time zone,
    employee_no character varying(10),
    end_date timestamp with time zone,
    full_name character varying(20),
    is_deputy boolean,
    manager_role_type character varying(2),
    org_code character varying(20),
    organization_managerid bigint,
    status character varying(2),
    tenantid character varying(10)
);


ALTER TABLE public.api_organization_manager OWNER TO postgres;

--
-- Name: api_organization_manager_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.api_organization_manager_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.api_organization_manager_id_seq OWNER TO postgres;

--
-- Name: api_organization_manager_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.api_organization_manager_id_seq OWNED BY public.api_organization_manager.id;


--
-- Name: api_organization_relation; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.api_organization_relation (
    id bigint NOT NULL,
    company_code character varying(20),
    company_party_id bigint,
    created_date timestamp with time zone,
    data_created_date timestamp with time zone,
    data_created_user character varying(30),
    data_modified_date timestamp with time zone,
    data_modified_user character varying(30),
    org_code character varying(20),
    org_name character varying(50),
    org_tree_type character varying(2),
    organization_relation_id bigint,
    parent_org_code character varying(20),
    status character varying(2),
    tenant_id character varying(10)
);


ALTER TABLE public.api_organization_relation OWNER TO postgres;

--
-- Name: api_organization_relation_action_log; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.api_organization_relation_action_log (
    id bigint NOT NULL,
    action character varying(10) NOT NULL,
    action_date timestamp with time zone NOT NULL,
    created_date timestamp with time zone NOT NULL,
    field_name character varying(30) NOT NULL,
    is_sync boolean,
    new_value character varying(50),
    old_value character varying(50),
    org_code character varying(20) NOT NULL,
    organization_relation_id bigint
);


ALTER TABLE public.api_organization_relation_action_log OWNER TO postgres;

--
-- Name: api_organization_relation_action_log_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.api_organization_relation_action_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.api_organization_relation_action_log_id_seq OWNER TO postgres;

--
-- Name: api_organization_relation_action_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.api_organization_relation_action_log_id_seq OWNED BY public.api_organization_relation_action_log.id;


--
-- Name: api_organization_relation_archived; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.api_organization_relation_archived (
    id bigint NOT NULL,
    archived_date timestamp with time zone,
    company_code character varying(20),
    company_party_id bigint,
    created_date timestamp with time zone,
    data_created_date timestamp with time zone,
    data_created_user character varying(30),
    data_modified_date timestamp with time zone,
    data_modified_user character varying(30),
    org_code character varying(20),
    org_name character varying(50),
    org_tree_type character varying(2),
    organization_relation_id bigint,
    parent_org_code character varying(20),
    status character varying(2),
    tenant_id character varying(10)
);


ALTER TABLE public.api_organization_relation_archived OWNER TO postgres;

--
-- Name: api_organization_relation_archived_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.api_organization_relation_archived_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.api_organization_relation_archived_id_seq OWNER TO postgres;

--
-- Name: api_organization_relation_archived_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.api_organization_relation_archived_id_seq OWNED BY public.api_organization_relation_archived.id;


--
-- Name: api_organization_relation_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.api_organization_relation_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.api_organization_relation_id_seq OWNER TO postgres;

--
-- Name: api_organization_relation_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.api_organization_relation_id_seq OWNED BY public.api_organization_relation.id;


--
-- Name: fse7en_org_deptgradeinfo; Type: MATERIALIZED VIEW; Schema: public; Owner: postgres
--

CREATE MATERIALIZED VIEW public.fse7en_org_deptgradeinfo AS
 SELECT DISTINCT ('Level'::text || ( SELECT count(*) AS count
           FROM public.get_org_hierarchy_by_org_code(aor.org_code, '0'::character varying) get_org_hierarchy_by_org_code(org_code, org_name, org_tree_type, parent_org_code, level))) AS grade_id,
    (('第'::text || ( SELECT count(*) AS count
           FROM public.get_org_hierarchy_by_org_code(aor.org_code, '0'::character varying) get_org_hierarchy_by_org_code(org_code, org_name, org_tree_type, parent_org_code, level))) || '階'::text) AS display_name,
    ('-'::text || ( SELECT count(*) AS count
           FROM public.get_org_hierarchy_by_org_code(aor.org_code, '0'::character varying) get_org_hierarchy_by_org_code(org_code, org_name, org_tree_type, parent_org_code, level))) AS grade_num
   FROM (public.api_organization_relation aor
     JOIN public.api_organization ao ON (((aor.org_code)::text = (ao.org_code)::text)))
  WHERE ((aor.org_tree_type)::text = '0'::text)
  ORDER BY ('Level'::text || ( SELECT count(*) AS count
           FROM public.get_org_hierarchy_by_org_code(aor.org_code, '0'::character varying) get_org_hierarchy_by_org_code(org_code, org_name, org_tree_type, parent_org_code, level)))
  WITH NO DATA;


ALTER TABLE public.fse7en_org_deptgradeinfo OWNER TO postgres;

--
-- Name: fse7en_org_deptinfo; Type: MATERIALIZED VIEW; Schema: public; Owner: postgres
--

CREATE MATERIALIZED VIEW public.fse7en_org_deptinfo AS
 SELECT ao.org_code,
    ao.org_name
   FROM (public.api_organization ao
     JOIN public.api_organization_relation aor ON (((ao.org_code)::text = (aor.org_code)::text)))
  WHERE ((aor.org_tree_type)::text = '0'::text)
  ORDER BY ao.org_code
  WITH NO DATA;


ALTER TABLE public.fse7en_org_deptinfo OWNER TO postgres;

--
-- Name: fse7en_org_deptstruct; Type: MATERIALIZED VIEW; Schema: public; Owner: postgres
--

CREATE MATERIALIZED VIEW public.fse7en_org_deptstruct AS
 SELECT aor.org_code,
    aor.parent_org_code,
    ('Level'::text || ( SELECT count(*) AS count
           FROM public.get_org_hierarchy_by_org_code(aor.org_code, '0'::character varying) get_org_hierarchy_by_org_code(org_code, org_name, org_tree_type, parent_org_code, level))) AS grade_id,
    ('-'::text || ( SELECT count(*) AS count
           FROM public.get_org_hierarchy_by_org_code(aor.org_code, '0'::character varying) get_org_hierarchy_by_org_code(org_code, org_name, org_tree_type, parent_org_code, level))) AS grade_num,
    ( SELECT get_org_hierarchy_by_org_code.org_code
           FROM public.get_org_hierarchy_by_org_code(aor.org_code, '0'::character varying) get_org_hierarchy_by_org_code(org_code, org_name, org_tree_type, parent_org_code, level)
          WHERE (get_org_hierarchy_by_org_code.parent_org_code IS NULL)) AS top_orgtree_code
   FROM (public.api_organization_relation aor
     JOIN public.api_organization ao ON (((aor.org_code)::text = (ao.org_code)::text)))
  WHERE ((aor.org_tree_type)::text = '0'::text)
  ORDER BY aor.org_code
  WITH NO DATA;


ALTER TABLE public.fse7en_org_deptstruct OWNER TO postgres;

--
-- Name: fse7en_org_jobtitle2grade; Type: MATERIALIZED VIEW; Schema: public; Owner: postgres
--

CREATE MATERIALIZED VIEW public.fse7en_org_jobtitle2grade AS
 SELECT DISTINCT (aei.job_title_code)::character varying AS job_title_code,
    aei.job_title_name,
    regexp_replace(((aei.job_title_code)::character varying)::text, '[A-Za-z]'::text, ''::text, 'g'::text) AS job_grade
   FROM public.api_employee_info aei
  WHERE (aei.job_title_code IS NOT NULL)
  ORDER BY (aei.job_title_code)::character varying DESC
  WITH NO DATA;


ALTER TABLE public.fse7en_org_jobtitle2grade OWNER TO postgres;

--
-- Name: fse7en_org_memberinfo; Type: MATERIALIZED VIEW; Schema: public; Owner: postgres
--

CREATE MATERIALIZED VIEW public.fse7en_org_memberinfo AS
 SELECT DISTINCT api_employee_info.employee_no,
    api_employee_info.full_name,
    api_employee_info.ext_no,
    api_employee_info.email_address,
        CASE
            WHEN ((api_employee_info.employed_status)::text = '1'::text) THEN '0'::text
            ELSE '1'::text
        END AS isterminated,
    api_employee_info.hire_date,
    api_employee_info.email_address AS azureaccount,
    CASE 
        WHEN api_employee_info.job_flag IN ('1', '4') THEN 'BU'
        WHEN api_employee_info.job_flag IS NULL THEN null
        ELSE 'ADM' 
    END as job_flag
   FROM public.api_employee_info
  ORDER BY
        CASE
            WHEN ((api_employee_info.employed_status)::text = '1'::text) THEN '0'::text
            ELSE '1'::text
        END, api_employee_info.employee_no
  WITH NO DATA;


ALTER TABLE public.fse7en_org_memberinfo OWNER TO postgres;

--
-- Name: fse7en_org_memberstruct; Type: MATERIALIZED VIEW; Schema: public; Owner: postgres
--


CREATE MATERIALIZED VIEW public.fse7en_org_memberstruct AS
WITH RECURSIVE employee_managers AS (
    SELECT aei_1.employee_no,
           aei_1.full_name,
           aei_1.form_org_code,
           aei_1.form_org_name,
           aei_1.formula_org_code,
           aei_1.formula_org_name,
           aei_1.job_title_code,
           aei_1.employed_status,
           -- 修正：排除自己當主管的情況，並確保主管在職狀態
           ( SELECT aom_1.employee_no
             FROM api_organization_manager aom_1
             JOIN api_employee_info aei_mgr ON aom_1.employee_no = aei_mgr.employee_no
            WHERE aom_1.org_code::text = aei_1.form_org_code::text
              AND aom_1.employee_no::text <> aei_1.employee_no::text
              AND aei_mgr.employed_status in ('1', '2')  -- 主管在職狀態
            ORDER BY aom_1.effective_date DESC
           LIMIT 1) AS direct_manager_employee_no,
           -- 找上層組織
           ( SELECT aor.parent_org_code
             FROM api_organization_relation aor
            WHERE aor.org_code::text = aei_1.form_org_code::text
              AND aor.org_tree_type::text = '2'::text
           LIMIT 1) AS parent_org_code
    FROM api_employee_info aei_1
), org_hierarchy AS (
    -- 遞迴查找組織階層，直到找到不是自己的主管
    SELECT 
        em.employee_no,
        em.form_org_code as current_org_code,
        em.parent_org_code,
        1 as level
    FROM employee_managers em
    
    UNION ALL
    
    SELECT 
        oh.employee_no,
        oh.parent_org_code as current_org_code,
        ( SELECT aor.parent_org_code
          FROM api_organization_relation aor
         WHERE aor.org_code::text = oh.parent_org_code::text
           AND aor.org_tree_type::text = '2'::text
        LIMIT 1) as parent_org_code,
        oh.level + 1
    FROM org_hierarchy oh
    WHERE oh.parent_org_code IS NOT NULL 
      AND oh.level < 10  -- 防止無限遞迴
), employee_with_approve_right AS (
    SELECT em.employee_no,
           em.full_name,
           em.form_org_code,
           em.form_org_name,
           em.formula_org_code,
           em.formula_org_name,
           em.job_title_code,
           em.employed_status,
           CASE
               -- 先檢查是否有直屬主管（不是自己且在職）
               WHEN em.direct_manager_employee_no IS NOT NULL 
               THEN em.direct_manager_employee_no
               -- 如果沒有，則遞迴往上找組織主管（不是自己且在職）
               ELSE ( SELECT aom.employee_no
                      FROM org_hierarchy oh
                      JOIN api_organization_manager aom ON aom.org_code::text = oh.current_org_code::text
                      JOIN api_employee_info aei_mgr ON aom.employee_no = aei_mgr.employee_no
                     WHERE oh.employee_no = em.employee_no
                       AND aom.employee_no::text <> em.employee_no::text
                       AND aei_mgr.employed_status in ('1', '2')  -- 主管在職狀態
                     ORDER BY oh.level ASC, aom.effective_date DESC
                    LIMIT 1)
           END AS approval_manager_employee_no
    FROM employee_managers em
)
SELECT DISTINCT aei.employee_no,
       aei.full_name,
       aei.formula_org_code AS org_code,
       aei.formula_org_name AS org_name,
       aei.job_title_code,
       regexp_replace(aei.job_title_code::character varying::text, '[A-Za-z]'::text, ''::text, 'g'::text) AS job_grade,
       CASE
           WHEN aom.manager_role_type IS NULL THEN '1'::text
           WHEN aom.manager_role_type::text = '1'::text OR aom.manager_role_type::text = ''::text THEN '1'::text
           ELSE '0'::text
       END AS is_main_job,
       CASE
           WHEN (EXISTS ( SELECT 1
                         FROM employee_with_approve_right ewar
                        WHERE ewar.approval_manager_employee_no::text = aei.employee_no::text)) THEN '1'::text
           ELSE '0'::text
       END AS approve_right,
       CASE
           WHEN aei.employed_status::text = '1'::text THEN '1'::text
           ELSE '0'::text
       END AS enable,
       -- 修正 instructor：主管員工編號@主管的formula_org_code，確保主管在職
       ( SELECT CASE 
                   WHEN ewar.approval_manager_employee_no IS NOT NULL 
                   THEN (ewar.approval_manager_employee_no::text || '@'::text) || a2.formula_org_code::text
                   ELSE NULL 
                END
         FROM employee_with_approve_right ewar
           LEFT JOIN api_employee_info a2 ON ewar.approval_manager_employee_no::text = a2.employee_no::text
        WHERE ewar.employee_no::text = aei.employee_no::text 
          AND a2.employed_status in ('1', '2')  -- 最終確認主管在職狀態
       ORDER BY aom.effective_date DESC  -- 按生效日期排序，取最近的
       LIMIT 1) AS instructor
FROM api_employee_info aei
  LEFT JOIN api_organization_manager aom ON aei.employee_no::text = aom.employee_no::text
  LEFT JOIN api_organization ao ON aei.form_org_code::text = ao.org_code::text
WHERE
    CASE
        WHEN aom.manager_role_type IS NULL THEN '1'::text
        WHEN aom.manager_role_type::text = '1'::text OR aom.manager_role_type::text = ''::text THEN '1'::text
        ELSE '0'::text
    END = '1'::text
ORDER BY 
    aei.employee_no
WITH NO DATA;


ALTER TABLE public.fse7en_org_memberstruct OWNER TO postgres;

--
-- Name: materialized_view_refresh_log; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.materialized_view_refresh_log (
    id bigint NOT NULL,
    view_name character varying(255) NOT NULL,
    refreshed_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    diff_count integer,
    diff_details jsonb
);


ALTER TABLE public.materialized_view_refresh_log OWNER TO postgres;

--
-- Name: materialized_view_refresh_log_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.materialized_view_refresh_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.materialized_view_refresh_log_id_seq OWNER TO postgres;

--
-- Name: materialized_view_refresh_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.materialized_view_refresh_log_id_seq OWNED BY public.materialized_view_refresh_log.id;


--
-- Name: api_company id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_company ALTER COLUMN id SET DEFAULT nextval('public.api_company_id_seq'::regclass);


--
-- Name: api_employee_info id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_employee_info ALTER COLUMN id SET DEFAULT nextval('public.api_employee_info_id_seq'::regclass);


--
-- Name: api_employee_info_action_log id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_employee_info_action_log ALTER COLUMN id SET DEFAULT nextval('public.api_employee_info_action_log_id_seq'::regclass);


--
-- Name: api_employee_info_archived id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_employee_info_archived ALTER COLUMN id SET DEFAULT nextval('public.api_employee_info_archived_id_seq'::regclass);


--
-- Name: api_organization id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_organization ALTER COLUMN id SET DEFAULT nextval('public.api_organization_id_seq'::regclass);


--
-- Name: api_organization_action_log id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_organization_action_log ALTER COLUMN id SET DEFAULT nextval('public.api_organization_action_log_id_seq'::regclass);


--
-- Name: api_organization_archived id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_organization_archived ALTER COLUMN id SET DEFAULT nextval('public.api_organization_archived_id_seq'::regclass);


--
-- Name: api_organization_manager id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_organization_manager ALTER COLUMN id SET DEFAULT nextval('public.api_organization_manager_id_seq'::regclass);


--
-- Name: api_organization_relation id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_organization_relation ALTER COLUMN id SET DEFAULT nextval('public.api_organization_relation_id_seq'::regclass);


--
-- Name: api_organization_relation_action_log id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_organization_relation_action_log ALTER COLUMN id SET DEFAULT nextval('public.api_organization_relation_action_log_id_seq'::regclass);


--
-- Name: api_organization_relation_archived id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_organization_relation_archived ALTER COLUMN id SET DEFAULT nextval('public.api_organization_relation_archived_id_seq'::regclass);


--
-- Name: materialized_view_refresh_log id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.materialized_view_refresh_log ALTER COLUMN id SET DEFAULT nextval('public.materialized_view_refresh_log_id_seq'::regclass);


--
-- Name: api_company api_company_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_company
    ADD CONSTRAINT api_company_pkey PRIMARY KEY (id);


--
-- Name: api_employee_info_action_log api_employee_info_action_log_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_employee_info_action_log
    ADD CONSTRAINT api_employee_info_action_log_pkey PRIMARY KEY (id);


--
-- Name: api_employee_info_archived api_employee_info_archived_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_employee_info_archived
    ADD CONSTRAINT api_employee_info_archived_pkey PRIMARY KEY (id);


--
-- Name: api_employee_info api_employee_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_employee_info
    ADD CONSTRAINT api_employee_info_pkey PRIMARY KEY (id);


--
-- Name: api_organization_action_log api_organization_action_log_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_organization_action_log
    ADD CONSTRAINT api_organization_action_log_pkey PRIMARY KEY (id);


--
-- Name: api_organization_archived api_organization_archived_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_organization_archived
    ADD CONSTRAINT api_organization_archived_pkey PRIMARY KEY (id);


--
-- Name: api_organization_manager api_organization_manager_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_organization_manager
    ADD CONSTRAINT api_organization_manager_pkey PRIMARY KEY (id);


--
-- Name: api_organization api_organization_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_organization
    ADD CONSTRAINT api_organization_pkey PRIMARY KEY (id);


--
-- Name: api_organization_relation_action_log api_organization_relation_action_log_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_organization_relation_action_log
    ADD CONSTRAINT api_organization_relation_action_log_pkey PRIMARY KEY (id);


--
-- Name: api_organization_relation_archived api_organization_relation_archived_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_organization_relation_archived
    ADD CONSTRAINT api_organization_relation_archived_pkey PRIMARY KEY (id);


--
-- Name: api_organization_relation api_organization_relation_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_organization_relation
    ADD CONSTRAINT api_organization_relation_pkey PRIMARY KEY (id);


--
-- Name: materialized_view_refresh_log materialized_view_refresh_log_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.materialized_view_refresh_log
    ADD CONSTRAINT materialized_view_refresh_log_pkey PRIMARY KEY (id);


--
-- Name: api_organization uk_s4lkcneefc2dlhaxcbhi5a8s0; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_organization
    ADD CONSTRAINT uk_s4lkcneefc2dlhaxcbhi5a8s0 UNIQUE (org_code);


--
-- Name: idx_fse7en_org_deptgradeinfo_grade_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_fse7en_org_deptgradeinfo_grade_id ON public.fse7en_org_deptgradeinfo USING btree (grade_id);


--
-- Name: idx_fse7en_org_deptinfo_org_code; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_fse7en_org_deptinfo_org_code ON public.fse7en_org_deptinfo USING btree (org_code);


--
-- Name: idx_fse7en_org_deptstruct_org_code; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_fse7en_org_deptstruct_org_code ON public.fse7en_org_deptstruct USING btree (org_code);


--
-- Name: idx_fse7en_org_jobtitle2grade_code; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_fse7en_org_jobtitle2grade_code ON public.fse7en_org_jobtitle2grade USING btree (job_title_code);


--
-- Name: idx_fse7en_org_memberinfo_employee_no; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_fse7en_org_memberinfo_employee_no ON public.fse7en_org_memberinfo USING btree (employee_no);


--
-- Name: idx_fse7en_org_memberstruct_emp_org; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_fse7en_org_memberstruct_emp_org ON public.fse7en_org_memberstruct USING btree (employee_no, org_code);


--
-- PostgreSQL database dump complete
--


CREATE TABLE public.approved_amount_for_each_layer (
    layer_code character varying(10) NOT NULL,
    layer_description character varying(20),
    max_capital_fee NUMERIC(18, 2),
    max_expense_fee NUMERIC(18, 2),
    max_payment_penalty_fee NUMERIC(18, 2),
    max_payment_relation_fee NUMERIC(18, 2),
    max_payment_current_capital_fee NUMERIC(18, 2),
    max_payment_regular_expense_fee NUMERIC(18, 2),
    max_payment_other_capital_fee NUMERIC(18, 2),
    max_payment_other_expense_fee NUMERIC(18, 2)
);

ALTER TABLE public.approved_amount_for_each_layer OWNER TO postgres;

insert into public.approved_amount_for_each_layer values('0', '董事長', 999999999, 999999999, 999999999, 999999999, 0, 0, 999999999, 999999999);
insert into public.approved_amount_for_each_layer values('1', '總經理', 5000000, 2500000, 2500000, 300000, 0, 0, 5000000, 2500000);
insert into public.approved_amount_for_each_layer values('2', '本部主管', 1000000, 500000, 500000, 50000, 0, 0, 1000000, 500000);
insert into public.approved_amount_for_each_layer values('3', '店內店長主管.總公司部級主管', 500000, 250000, 100000, 20000, 999999999, 999999999, 500000, 250000);
insert into public.approved_amount_for_each_layer values('4', '店內部級主管.總公司課級主管', 100000, 50000, 20000, 10000, 200000, 100000, 100000, 50000);
insert into public.approved_amount_for_each_layer values('5', '店內課級主管', 0, 10000, 0, 0, 0, 20000, 0, 10000);


CREATE OR REPLACE VIEW public.view_employee_approval_amount AS
SELECT 
    DISTINCT
    emp.employee_no,
    emp.full_name,
    emp.job_title_code,
    emp.job_title_name,
    form_org.layer_code,
    form_org.layer_name,
    approvedAmount.layer_description,
    formula_org.org_code AS formula_org_code,
    formula_org.org_name AS formula_org_name,
    form_org.org_code AS form_org_code,
    form_org.org_name AS form_org_name,
    approvedAmount.max_capital_fee,
    approvedAmount.max_expense_fee,
    approvedAmount.max_payment_penalty_fee,
    approvedAmount.max_payment_relation_fee,
    approvedAmount.max_payment_current_capital_fee,
    approvedAmount.max_payment_regular_expense_fee,
    approvedAmount.max_payment_other_capital_fee,
    approvedAmount.max_payment_other_expense_fee
FROM 
    public.api_employee_info emp
LEFT JOIN 
    public.api_organization form_org ON emp.form_org_code = form_org.org_code
LEFT JOIN 
    public.api_organization formula_org ON emp.formula_org_code = formula_org.org_code
LEFT JOIN 
    public.FSe7en_Org_MemberStruct memberStruct ON emp.employee_no = memberStruct.employee_no
LEFT JOIN 
    public.approved_amount_for_each_layer approvedAmount ON form_org.layer_code = approvedAmount.layer_code
where emp.employed_status in ('1', '2') and memberStruct.approve_right='1'
order by layer_code;

REFRESH MATERIALIZED VIEW fse7en_org_deptgradeinfo;
REFRESH MATERIALIZED VIEW fse7en_org_deptinfo;
REFRESH MATERIALIZED VIEW fse7en_org_deptstruct;
REFRESH MATERIALIZED VIEW fse7en_org_jobtitle2grade;
REFRESH MATERIALIZED VIEW fse7en_org_memberinfo;
REFRESH MATERIALIZED VIEW fse7en_org_memberstruct;

--EHR_EMPLOYEE
CREATE OR REPLACE VIEW view_core_ehr_employee AS
SELECT 
  employee_no, 
  full_name, 
  formula_org_code, 
  job_title_code, 
  job_title_name,
  position_code, 
  job_grade_code, 
  job_level_code, 
  to_char(hire_date, 'yyyy-MM-dd') as hire_date, 
  to_char(resignation_date, 'yyyy-MM-dd') as resignation_date, 
  (
    SELECT 
      manager_employee_no 
    FROM 
      Find_approval_manager(employee_no, '0')
  ) AS manager_employee_no, 
  (
    SELECT 
      manager_full_name 
    FROM 
      Find_approval_manager(employee_no, '0')
  ) AS manager_full_name, 
  employed_status, 
  email_address 
FROM 
  api_employee_info ami;


--EHR_DEPARTMENT
CREATE OR REPLACE VIEW view_core_ehr_department AS
select 
  ao.org_code, 
  ao.org_name as org_abbr_name, 
  ao.org_name, 
  (
    select 
      employee_no 
    from 
      api_organization_manager 
    where 
      manager_role_type is null 
      or manager_role_type = '' 
      and org_code = aor.parent_org_code 
    order by 
      effective_date desc 
    limit 
      1
  ) as manager_employee_no, 
  (
    select 
      full_name 
    from 
      api_organization_manager 
    where 
      manager_role_type is null 
      or manager_role_type = '' 
      and org_code = aor.parent_org_code 
    order by 
      effective_date desc 
    limit 
      1
  ) as manager_full_name, 
  aor.parent_org_code as parent_org_code, 
  to_char(ao.start_date, 'yyyy-MM-dd')  as start_date, 
  to_char(ao.end_date, 'yyyy-MM-dd') as end_date, 
  false as is_virtual 
from 
  api_organization ao 
  left outer join api_organization_relation aor on ao.org_code = aor.org_code 
  and aor.org_tree_type = '0';


