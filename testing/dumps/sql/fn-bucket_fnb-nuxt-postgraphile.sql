-- Downloaded from: https://github.com/fn-bucket/fnb-nuxt-postgraphile/blob/cbf414ed971338c4cf9c40db64dbadc36f1f6353/db/fnb.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 15.0 (Debian 15.0-1.pgdg110+1)
-- Dumped by pg_dump version 15.1

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
-- Name: app; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA app;


ALTER SCHEMA app OWNER TO postgres;

--
-- Name: auth_bootstrap; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA auth_bootstrap;


ALTER SCHEMA auth_bootstrap OWNER TO postgres;

--
-- Name: auth_fn; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA auth_fn;


ALTER SCHEMA auth_fn OWNER TO postgres;

--
-- Name: auth_fn_private; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA auth_fn_private;


ALTER SCHEMA auth_fn_private OWNER TO postgres;

--
-- Name: msg; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA msg;


ALTER SCHEMA msg OWNER TO postgres;

--
-- Name: msg_fn; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA msg_fn;


ALTER SCHEMA msg_fn OWNER TO postgres;

--
-- Name: postgraphile_watch; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA postgraphile_watch;


ALTER SCHEMA postgraphile_watch OWNER TO postgres;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: postgres
--

-- *not* creating schema, since initdb creates it


ALTER SCHEMA public OWNER TO postgres;

--
-- Name: shard_1; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA shard_1;


ALTER SCHEMA shard_1 OWNER TO postgres;

--
-- Name: app_route_menu_behavior; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.app_route_menu_behavior AS ENUM (
    'none',
    'module',
    'navbar'
);


ALTER TYPE app.app_route_menu_behavior OWNER TO postgres;

--
-- Name: app_tenant_payment_status; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.app_tenant_payment_status AS ENUM (
    'current',
    'warning',
    'delinquent'
);


ALTER TYPE app.app_tenant_payment_status OWNER TO postgres;

--
-- Name: app_tenant_payment_status_summary_result; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.app_tenant_payment_status_summary_result AS (
	status app.app_tenant_payment_status
);


ALTER TYPE app.app_tenant_payment_status_summary_result OWNER TO postgres;

--
-- Name: app_tenant_type; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.app_tenant_type AS ENUM (
    'anchor',
    'customer',
    'subsidiary',
    'test',
    'demo',
    'tutorial',
    'pending'
);


ALTER TYPE app.app_tenant_type OWNER TO postgres;

--
-- Name: application_setting; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.application_setting AS (
	key text,
	name text,
	value text
);


ALTER TYPE app.application_setting OWNER TO postgres;

--
-- Name: application_setting_scope; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.application_setting_scope AS ENUM (
    'applicaton',
    'license_pack_anchor',
    'app_tenant',
    'app_user'
);


ALTER TYPE app.application_setting_scope OWNER TO postgres;

--
-- Name: application_setting_type; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.application_setting_type AS ENUM (
    'string',
    'number',
    'array'
);


ALTER TYPE app.application_setting_type OWNER TO postgres;

--
-- Name: setting_accepted_value; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.setting_accepted_value AS (
	label text,
	value text
);


ALTER TYPE app.setting_accepted_value OWNER TO postgres;

--
-- Name: application_setting_config; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.application_setting_config AS (
	key text,
	name text,
	scope app.application_setting_scope,
	type app.application_setting_type,
	accepted_values app.setting_accepted_value[],
	ordinal integer,
	default_value text,
	value text,
	tenant_edit_key text
);


ALTER TYPE app.application_setting_config OWNER TO postgres;

--
-- Name: contact_status; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.contact_status AS ENUM (
    'active',
    'inactive'
);


ALTER TYPE app.contact_status OWNER TO postgres;

--
-- Name: contact_type; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.contact_type AS ENUM (
    'individual',
    'group'
);


ALTER TYPE app.contact_type OWNER TO postgres;

--
-- Name: error_report_status; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.error_report_status AS ENUM (
    'captured',
    'working',
    'addressed'
);


ALTER TYPE app.error_report_status OWNER TO postgres;

--
-- Name: expiration_interval; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.expiration_interval AS ENUM (
    'none',
    'minute',
    'hour',
    'day',
    'week',
    'month',
    'quarter',
    'year',
    'explicit'
);


ALTER TYPE app.expiration_interval OWNER TO postgres;

--
-- Name: license_pack_availability; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.license_pack_availability AS ENUM (
    'published',
    'draft',
    'discontinued'
);


ALTER TYPE app.license_pack_availability OWNER TO postgres;

--
-- Name: license_pack_type; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.license_pack_type AS ENUM (
    'anchor',
    'addon'
);


ALTER TYPE app.license_pack_type OWNER TO postgres;

--
-- Name: license_type_upgrade; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.license_type_upgrade AS (
	source_license_type_key text,
	target_license_type_key text
);


ALTER TYPE app.license_type_upgrade OWNER TO postgres;

--
-- Name: renewal_frequency; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.renewal_frequency AS ENUM (
    'never',
    'weekly',
    'monthly',
    'quarterly',
    'yearly',
    'expires'
);


ALTER TYPE app.renewal_frequency OWNER TO postgres;

--
-- Name: upgrade_path; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.upgrade_path AS (
	license_pack_key text,
	license_type_upgrades app.license_type_upgrade[]
);


ALTER TYPE app.upgrade_path OWNER TO postgres;

--
-- Name: upgrade_config; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.upgrade_config AS (
	upgrade_paths app.upgrade_path[]
);


ALTER TYPE app.upgrade_config OWNER TO postgres;

--
-- Name: id_generator(integer); Type: FUNCTION; Schema: shard_1; Owner: postgres
--

CREATE FUNCTION shard_1.id_generator(shard_id integer DEFAULT 1) RETURNS text
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
    our_epoch bigint := 1314220021721;
    seq_id bigint;
    now_millis bigint;
    presult bigint;
    result text;
BEGIN
    SELECT nextval('shard_1.global_id_sequence') % 1024 INTO seq_id;

    SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000) INTO now_millis;
    presult := (now_millis - our_epoch) << 23;
    presult := presult | (shard_id << 10);
    presult := presult | (seq_id);
    result := presult::text;
    return result;
END;
$$;


ALTER FUNCTION shard_1.id_generator(shard_id integer) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: license_pack; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.license_pack (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    key text NOT NULL,
    name text NOT NULL,
    availability app.license_pack_availability DEFAULT 'draft'::app.license_pack_availability NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    published_at timestamp with time zone,
    discontinued_at timestamp with time zone,
    type app.license_pack_type NOT NULL,
    renewal_frequency app.renewal_frequency DEFAULT 'monthly'::app.renewal_frequency NOT NULL,
    expiration_interval app.expiration_interval DEFAULT 'none'::app.expiration_interval NOT NULL,
    expiration_interval_multiplier integer DEFAULT 1 NOT NULL,
    explicit_expiration_date date,
    price numeric(10,2) DEFAULT 0 NOT NULL,
    upgrade_config app.upgrade_config DEFAULT ROW(ARRAY[]::app.upgrade_path[])::app.upgrade_config NOT NULL,
    available_add_on_keys text[] DEFAULT '{}'::text[] NOT NULL,
    coupon_code text,
    is_public_offering boolean,
    application_settings app.application_setting[],
    implicit_add_on_keys text[] DEFAULT '{}'::text[] NOT NULL
);


ALTER TABLE app.license_pack OWNER TO postgres;

--
-- Name: license_pack_sibling_set; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.license_pack_sibling_set AS (
	published app.license_pack,
	draft app.license_pack,
	discontinued app.license_pack[]
);


ALTER TYPE app.license_pack_sibling_set OWNER TO postgres;

--
-- Name: license_status; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.license_status AS ENUM (
    'active',
    'inactive',
    'expired',
    'void'
);


ALTER TYPE app.license_status OWNER TO postgres;

--
-- Name: license_status_reason; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.license_status_reason AS ENUM (
    'initial',
    'adjustment',
    'account_deletion',
    'consolidation',
    'subscription_deactivation'
);


ALTER TYPE app.license_status_reason OWNER TO postgres;

--
-- Name: order_direction; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.order_direction AS ENUM (
    'asc',
    'desc'
);


ALTER TYPE app.order_direction OWNER TO postgres;

--
-- Name: permission_key; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.permission_key AS ENUM (
    'Admin',
    'SuperAdmin',
    'Demon',
    'Tenant',
    'Support',
    'Demo',
    'User'
);


ALTER TYPE app.permission_key OWNER TO postgres;

--
-- Name: license_pack_license_type; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.license_pack_license_type (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    license_type_key text NOT NULL,
    license_pack_id text NOT NULL,
    license_count integer DEFAULT 1 NOT NULL,
    assign_upon_subscription boolean DEFAULT false NOT NULL,
    unlimited_provision boolean DEFAULT false NOT NULL,
    expiration_interval app.expiration_interval DEFAULT 'none'::app.expiration_interval NOT NULL,
    expiration_interval_multiplier integer DEFAULT 1 NOT NULL,
    explicit_expiration_date date
);


ALTER TABLE app.license_pack_license_type OWNER TO postgres;

--
-- Name: subscription_available_license_type; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.subscription_available_license_type AS (
	license_pack_license_type app.license_pack_license_type,
	provisioned_count integer,
	active_count integer,
	inactive_count integer,
	available_count integer,
	can_provision boolean,
	void_count integer,
	expired_count integer
);


ALTER TYPE app.subscription_available_license_type OWNER TO postgres;

--
-- Name: subscription_renewal_behavior; Type: TYPE; Schema: app; Owner: postgres
--

CREATE TYPE app.subscription_renewal_behavior AS ENUM (
    'renew',
    'expire',
    'ask_admin'
);


ALTER TYPE app.subscription_renewal_behavior OWNER TO postgres;

--
-- Name: app_tenant_auth0_info; Type: TYPE; Schema: auth_fn; Owner: postgres
--

CREATE TYPE auth_fn.app_tenant_auth0_info AS (
	id text,
	name text
);


ALTER TYPE auth_fn.app_tenant_auth0_info OWNER TO postgres;

--
-- Name: app_user_auth0_info; Type: TYPE; Schema: auth_fn; Owner: postgres
--

CREATE TYPE auth_fn.app_user_auth0_info AS (
	permission_key text,
	username text,
	inactive boolean,
	app_tenant_name text,
	app_tenant_id text,
	parent_app_tenant_id text,
	subsidiaries auth_fn.app_tenant_auth0_info[],
	app_user_id text,
	preferred_timezone text,
	contact_id text,
	first_name text,
	last_name text,
	recovery_email text,
	app_role text,
	permissions text[],
	home_path text,
	licensing_scope text[],
	ext_auth_id text,
	ext_auth_blocked boolean
);


ALTER TYPE auth_fn.app_user_auth0_info OWNER TO postgres;

--
-- Name: jwt_token_bootstrap; Type: TYPE; Schema: auth_bootstrap; Owner: postgres
--

CREATE TYPE auth_bootstrap.jwt_token_bootstrap AS (
	role text,
	app_user_id text,
	app_tenant_id text,
	permissions text,
	current_app_user auth_fn.app_user_auth0_info
);


ALTER TYPE auth_bootstrap.jwt_token_bootstrap OWNER TO postgres;

--
-- Name: app_tenant; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.app_tenant (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    name text NOT NULL,
    identifier text,
    organization_id text,
    registration_identifier text DEFAULT shard_1.id_generator() NOT NULL,
    registration_complete boolean DEFAULT false NOT NULL,
    settings jsonb DEFAULT '{}'::jsonb NOT NULL,
    type app.app_tenant_type DEFAULT 'customer'::app.app_tenant_type NOT NULL,
    parent_app_tenant_id text,
    anchor_subscription_id text,
    billing_topic_id text,
    CONSTRAINT app_tenant_check CHECK (((type = 'anchor'::app.app_tenant_type) OR (parent_app_tenant_id IS NOT NULL)))
);


ALTER TABLE app.app_tenant OWNER TO postgres;

--
-- Name: init_demo_result; Type: TYPE; Schema: auth_fn; Owner: postgres
--

CREATE TYPE auth_fn.init_demo_result AS (
	license_pack_key text,
	app_user_auth0_info auth_fn.app_user_auth0_info,
	demo_app_tenant app.app_tenant
);


ALTER TYPE auth_fn.init_demo_result OWNER TO postgres;

--
-- Name: init_subsidiary_admin_result; Type: TYPE; Schema: auth_fn; Owner: postgres
--

CREATE TYPE auth_fn.init_subsidiary_admin_result AS (
	app_user_auth0_info auth_fn.app_user_auth0_info,
	subsidiary_app_tenant app.app_tenant
);


ALTER TYPE auth_fn.init_subsidiary_admin_result OWNER TO postgres;

--
-- Name: jwt_token; Type: TYPE; Schema: auth_fn; Owner: postgres
--

CREATE TYPE auth_fn.jwt_token AS (
	role text,
	app_user_id text,
	app_tenant_id text
);


ALTER TYPE auth_fn.jwt_token OWNER TO postgres;

--
-- Name: email_status; Type: TYPE; Schema: msg; Owner: postgres
--

CREATE TYPE msg.email_status AS ENUM (
    'requested',
    'sent',
    'received',
    'error'
);


ALTER TYPE msg.email_status OWNER TO postgres;

--
-- Name: message_status; Type: TYPE; Schema: msg; Owner: postgres
--

CREATE TYPE msg.message_status AS ENUM (
    'draft',
    'sent',
    'deleted'
);


ALTER TYPE msg.message_status OWNER TO postgres;

--
-- Name: subscription_status; Type: TYPE; Schema: msg; Owner: postgres
--

CREATE TYPE msg.subscription_status AS ENUM (
    'active',
    'inactive',
    'blocked'
);


ALTER TYPE msg.subscription_status OWNER TO postgres;

--
-- Name: email_request_info; Type: TYPE; Schema: msg_fn; Owner: postgres
--

CREATE TYPE msg_fn.email_request_info AS (
	subject text,
	content text,
	from_address text,
	to_addresses text,
	cc_addresses text,
	bcc_addresses text,
	options jsonb
);


ALTER TYPE msg_fn.email_request_info OWNER TO postgres;

--
-- Name: message_info; Type: TYPE; Schema: msg_fn; Owner: postgres
--

CREATE TYPE msg_fn.message_info AS (
	id text,
	topic_id text,
	content text,
	tags text[]
);


ALTER TYPE msg_fn.message_info OWNER TO postgres;

--
-- Name: subscription_info; Type: TYPE; Schema: msg_fn; Owner: postgres
--

CREATE TYPE msg_fn.subscription_info AS (
	topic_id text,
	subscriber_contact_id text
);


ALTER TYPE msg_fn.subscription_info OWNER TO postgres;

--
-- Name: topic_info; Type: TYPE; Schema: msg_fn; Owner: postgres
--

CREATE TYPE msg_fn.topic_info AS (
	id text,
	name text,
	identifier text
);


ALTER TYPE msg_fn.topic_info OWNER TO postgres;

--
-- Name: app_user; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.app_user (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    app_tenant_id text NOT NULL,
    ext_auth_id text,
    ext_crm_id text,
    contact_id text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    username text NOT NULL,
    recovery_email text NOT NULL,
    inactive boolean DEFAULT false NOT NULL,
    password_reset_required boolean DEFAULT false NOT NULL,
    permission_key app.permission_key NOT NULL,
    is_support boolean DEFAULT false NOT NULL,
    preferred_timezone text DEFAULT 'PST8PDT'::text NOT NULL,
    settings jsonb DEFAULT '{}'::jsonb NOT NULL,
    ext_auth_blocked boolean DEFAULT false NOT NULL,
    language_id text DEFAULT 'en'::text NOT NULL
);


ALTER TABLE app.app_user OWNER TO postgres;

--
-- Name: app_tenant_active_guest_users(app.app_tenant); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.app_tenant_active_guest_users(_app_tenant app.app_tenant) RETURNS SETOF app.app_user
    LANGUAGE plpgsql STABLE
    AS $$
  DECLARE
  BEGIN
    return query
      select distinct
        u.*
      from app.app_user u
      join app.license l on l.assigned_to_app_user_id = u.id
      where u.app_tenant_id = _app_tenant.id
      and u.inactive = false
      and l.license_type_key like '%-guest-user'
      and u.permission_key in ('SuperAdmin', 'Admin', 'User')
    ;

  END
  $$;


ALTER FUNCTION app.app_tenant_active_guest_users(_app_tenant app.app_tenant) OWNER TO postgres;

--
-- Name: app_tenant_subscription; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.app_tenant_subscription (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    created_date date DEFAULT (CURRENT_TIMESTAMP)::date NOT NULL,
    expiration_date date,
    renewal_behavior app.subscription_renewal_behavior DEFAULT 'ask_admin'::app.subscription_renewal_behavior NOT NULL,
    is_anchor_subscription boolean DEFAULT true NOT NULL,
    inactive boolean DEFAULT false NOT NULL,
    app_tenant_id text NOT NULL,
    license_pack_id text NOT NULL,
    payment_processor_info jsonb DEFAULT '{}'::jsonb NOT NULL,
    parent_app_tenant_subscription_id text
);


ALTER TABLE app.app_tenant_subscription OWNER TO postgres;

--
-- Name: app_tenant_active_subscriptions(app.app_tenant); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.app_tenant_active_subscriptions(_app_tenant app.app_tenant) RETURNS SETOF app.app_tenant_subscription
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
  _app_tenant_id text;
BEGIN
  if _app_tenant.type = 'subsidiary' then
    _app_tenant_id = _app_tenant.parent_app_tenant_id;
  else
    _app_tenant_id = _app_tenant.id;
  end if;

  return query
  select ats.*
  from app.app_tenant_subscription ats
  where app_tenant_id = _app_tenant_id
  and inactive = false
  ;
END
$$;


ALTER FUNCTION app.app_tenant_active_subscriptions(_app_tenant app.app_tenant) OWNER TO postgres;

--
-- Name: app_tenant_active_users(app.app_tenant); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.app_tenant_active_users(_app_tenant app.app_tenant) RETURNS SETOF app.app_user
    LANGUAGE plpgsql STABLE
    AS $$
  DECLARE
  BEGIN
    return query
      select distinct
        u.*
      from app.app_user u
      left join app.license l on l.assigned_to_app_user_id = u.id
      where u.app_tenant_id = _app_tenant.id
      and u.inactive = false
      and (l.license_type_key is null or l.license_type_key != '%-guest-user')
      and u.permission_key in ('SuperAdmin', 'Admin', 'User')
    ;

  END
  $$;


ALTER FUNCTION app.app_tenant_active_users(_app_tenant app.app_tenant) OWNER TO postgres;

--
-- Name: app_tenant_available_licenses(app.app_tenant); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.app_tenant_available_licenses(_app_tenant app.app_tenant) RETURNS SETOF app.subscription_available_license_type
    LANGUAGE plpgsql STABLE SECURITY DEFINER
    AS $$
DECLARE
  _app_tenant_id text;
  _subscription app.app_tenant_subscription;
  _result app.subscription_available_license_type;
  _all_results app.subscription_available_license_type[];
BEGIN
  _all_results := '{}'::app.subscription_available_license_type[];

  -- need to first pull the anchor subscription for the app tenant, then do the following with the app_tenant_id from that
  select * into _subscription from app.app_tenant_subscription where id = _app_tenant.anchor_subscription_id;

  for _subscription in
    select * from app.app_tenant_subscription where app_tenant_id = _subscription.app_tenant_id
  loop
    for _result in
      (select * from app.app_tenant_subscription_available_licenses(_subscription))
    loop
      _all_results := array_append(_all_results, _result);
    end loop;
  end loop;

  return query
    select * from unnest(_all_results)
  ;
END
$$;


ALTER FUNCTION app.app_tenant_available_licenses(_app_tenant app.app_tenant) OWNER TO postgres;

--
-- Name: app_tenant_flags(app.app_tenant); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.app_tenant_flags(_app_tenant app.app_tenant) RETURNS jsonb
    LANGUAGE plpgsql STABLE
    AS $$
  DECLARE
    _app_flags jsonb;
    _app_user_id text;
    _app_tenant_is_linked boolean;
    _is_admin_or_therapist boolean;
    _err_context text;
  BEGIN
    _app_user_id := (select auth_fn.current_app_user()->>'app_user_id');

    _is_admin_or_therapist := (
      (
        select count(*) 
        from app.license 
        where assigned_to_app_user_id = _app_user_id
        and inactive = false 
        and (position('admin' in license_type_key::text) > 0 or position('therapist' in license_type_key::text) > 0)
        or (select auth_fn.app_user_has_permission('p:app-tenant-scope'))
      ) > 0
    );

    _app_flags := jsonb_build_object(
      -- 'whateverYouWant', _some_value_or flag
    );

    return _app_flags;

    exception
      when others then
        GET STACKED DIAGNOSTICS _err_context = PG_EXCEPTION_CONTEXT;
        if position('FB' in SQLSTATE::text) = 0 then
          _err_context := 'app.app_tenant_flags:::' || SQLSTATE::text || ':::' || SQLERRM::text || ':::' || _err_context;
          raise exception '%', _err_context using errcode = 'FB500';
        end if;
        raise;
  END
  $$;


ALTER FUNCTION app.app_tenant_flags(_app_tenant app.app_tenant) OWNER TO postgres;

--
-- Name: app_tenant_inactive_guest_users(app.app_tenant); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.app_tenant_inactive_guest_users(_app_tenant app.app_tenant) RETURNS SETOF app.app_user
    LANGUAGE plpgsql STABLE
    AS $$
  DECLARE
  BEGIN
    return query
      select distinct
        u.*
      from app.app_user u
      join app.license l on l.assigned_to_app_user_id = u.id
      where u.app_tenant_id = _app_tenant.id
      and u.inactive = true
      and l.license_type_key = '%-guest-user'
      and u.permission_key in ('SuperAdmin', 'Admin', 'User')
    ;

  END
  $$;


ALTER FUNCTION app.app_tenant_inactive_guest_users(_app_tenant app.app_tenant) OWNER TO postgres;

--
-- Name: app_tenant_inactive_users(app.app_tenant); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.app_tenant_inactive_users(_app_tenant app.app_tenant) RETURNS SETOF app.app_user
    LANGUAGE plpgsql STABLE
    AS $$
  DECLARE
  BEGIN
    return query
      select distinct
        u.*
      from app.app_user u
      left join app.license l on l.assigned_to_app_user_id = u.id
      where u.app_tenant_id = _app_tenant.id
      and u.inactive = true
      and (l.license_type_key is null or l.license_type_key != '%-guest-user')
      and u.permission_key in ('SuperAdmin', 'Admin', 'User')
    ;

  END
  $$;


ALTER FUNCTION app.app_tenant_inactive_users(_app_tenant app.app_tenant) OWNER TO postgres;

--
-- Name: app_tenant_license_type_is_available(text, text); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.app_tenant_license_type_is_available(_app_tenant_id text, _license_type_key text) RETURNS boolean
    LANGUAGE plpgsql STABLE SECURITY DEFINER
    AS $$
  DECLARE
    _license_pack_license_type app.license_pack_license_type;
    _subscription_available_license_type app.subscription_available_license_type;
    _err_context text;
  BEGIN
    with atal as (
      select
        (app.app_tenant_available_licenses(apt.*)).* al
      from app.app_tenant apt where id = _app_tenant_id
    )
    , cp as (
      select (atal).license_pack_license_type lplt
      from atal
      where can_provision = true
    )
    , lplt as (
      select
        (lplt).*
      from cp
    )
    select *
    into _license_pack_license_type
    from lplt
    where license_type_key = _license_type_key
    ;

    if _license_pack_license_type.id is null then
      return false;
    else
      return true;
    end if;
    exception
      when others then
        GET STACKED DIAGNOSTICS _err_context = PG_EXCEPTION_CONTEXT;
        if position('FB' in SQLSTATE::text) = 0 then
          _err_context := 'app_fn.app_tenant_license_type_is_available:::' || SQLSTATE::text || ':::' || SQLERRM::text || ':::' || _err_context;
          raise exception '%', _err_context using errcode = 'FB500';
        end if;
        raise;
END
$$;


ALTER FUNCTION app.app_tenant_license_type_is_available(_app_tenant_id text, _license_type_key text) OWNER TO postgres;

--
-- Name: app_tenant_payment_status_summary(app.app_tenant); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.app_tenant_payment_status_summary(_app_tenant app.app_tenant) RETURNS app.app_tenant_payment_status_summary_result
    LANGUAGE plpgsql STABLE
    AS $$
  DECLARE
    _err_context text;
    _result app.app_tenant_payment_status_summary_result;
  BEGIN
    select array_agg(bp.*)
    into _result.pastdue_payments
    from bill.payment bp
    where bp.app_tenant_id = _app_tenant.id
    and bp.status = 'pastdue'
    group by bp.billing_date
    order by bp.billing_date desc
    ;
    _result.pastdue_payments := coalesce(_result.pastdue_payments, '{}'::bill.payment[]);

    select array_agg(bp.*)
    into _result.scheduled_payments
    from bill.payment bp
    where bp.app_tenant_id = _app_tenant.id
    and bp.status = 'scheduled'
    group by bp.billing_date
    order by bp.billing_date asc
    ;
    _result.scheduled_payments := coalesce(_result.scheduled_payments, '{}'::bill.payment[]);


    _result.status = (
      select case
        when array_length(_result.pastdue_payments, 1) > 0 then
          'warning'
        when array_length(_result.pastdue_payments, 1) > 2 then
          'delinquent'
        else
          'current'
      end
    );

    return _result;

    exception
      when others then
        GET STACKED DIAGNOSTICS _err_context = PG_EXCEPTION_CONTEXT;
        if position('FB' in SQLSTATE::text) = 0 then
          _err_context := 'app.app_tenant_payment_status_summary:::' || SQLSTATE::text || ':::' || SQLERRM::text || ':::' || _err_context;
          raise exception '%', _err_context using errcode = 'FB500';
        end if;
        raise;
    END
  $$;


ALTER FUNCTION app.app_tenant_payment_status_summary(_app_tenant app.app_tenant) OWNER TO postgres;

--
-- Name: app_tenant_subscription_available_add_ons(app.app_tenant_subscription); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.app_tenant_subscription_available_add_ons(_app_tenant_subscription app.app_tenant_subscription) RETURNS SETOF app.license_pack
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
  _license_pack app.license_pack;
  _upgrade_config app.upgrade_config;
  _upgrade_path app.upgrade_path;
  _upgrade_paths app.upgrade_path[];
  _upgrade_keys text[];
BEGIN
  select * into _license_pack from app.license_pack where id = _app_tenant_subscription.license_pack_id;

  return query
  select lp.*
  from app.license_pack lp
  where lp.key = any(_license_pack.available_add_on_keys)
  and lp.availability = 'published'
  order by lp.key
  ;

END
$$;


ALTER FUNCTION app.app_tenant_subscription_available_add_ons(_app_tenant_subscription app.app_tenant_subscription) OWNER TO postgres;

--
-- Name: app_tenant_subscription_available_licenses(app.app_tenant_subscription); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.app_tenant_subscription_available_licenses(_app_tenant_subscription app.app_tenant_subscription) RETURNS SETOF app.subscription_available_license_type
    LANGUAGE plpgsql STABLE SECURITY DEFINER
    AS $$
  DECLARE
    _subscription_id text;
    _retval app.subscription_available_license_type[];
  BEGIN
    _subscription_id := _app_tenant_subscription.id;

    return query
    with lplt as (
      select lplt.*
      from app.license_pack_license_type lplt
      join app.license_pack lp on lp.id = lplt.license_pack_id
      join app.app_tenant_subscription ats on ats.license_pack_id = lp.id
      where ats.id = _subscription_id
    )
    ,p as (
      select
        lplt.license_type_key
        ,(
          select count(*)
          from app.license l
          where l.license_type_key = lplt.license_type_key
          and l.subscription_id = _subscription_id
        ) as provisioned_count
        ,(
          select count(*)
          from app.license l
          where l.license_type_key = lplt.license_type_key
          and l.subscription_id = _subscription_id
          and l.status = 'active'
        ) as active_count
        ,(
          select count(*)
          from app.license l
          where l.license_type_key = lplt.license_type_key
          and l.subscription_id = _subscription_id
          and l.status = 'inactive'
        ) as inactive_count
        ,(
          select count(*)
          from app.license l
          where l.license_type_key = lplt.license_type_key
          and l.subscription_id = _subscription_id
          and l.status = 'void'
        ) as void_count
        ,(
          select count(*)
          from app.license l
          where l.license_type_key = lplt.license_type_key
          and l.subscription_id = _subscription_id
          and l.status = 'expired'
        ) as expired_count
      from lplt
    )
    ,a as (
      select
        (lplt.*)::app.license_pack_license_type
        ,p.provisioned_count::integer
        ,p.active_count::integer
        ,p.inactive_count::integer
        ,case
          when lplt.unlimited_provision = true then null::integer
          else (lplt.license_count - p.provisioned_count)::integer
        end available_count
        ,case
          when _app_tenant_subscription.inactive = true then false
          when lplt.unlimited_provision::boolean then true
          when (lplt.license_count - p.provisioned_count)::integer > 0 then true
          else false
        end::boolean can_provision
        ,p.void_count::integer
        ,p.expired_count::integer
      from p
      join lplt on p.license_type_key = lplt.license_type_key
      and lplt.license_pack_id = _app_tenant_subscription.license_pack_id
    )
    select
      a.*
    from a
    ;
  END
  $$;


ALTER FUNCTION app.app_tenant_subscription_available_licenses(_app_tenant_subscription app.app_tenant_subscription) OWNER TO postgres;

--
-- Name: app_tenant_subscription_available_upgrade_paths(app.app_tenant_subscription); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.app_tenant_subscription_available_upgrade_paths(_app_tenant_subscription app.app_tenant_subscription) RETURNS SETOF app.license_pack
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
  _license_pack app.license_pack;
  _upgrade_config app.upgrade_config;
  _upgrade_path app.upgrade_path;
  _upgrade_paths app.upgrade_path[];
  _upgrade_keys text[];
BEGIN
  select * into _license_pack from app.license_pack where id = _app_tenant_subscription.license_pack_id;

  _upgrade_config := _license_pack.upgrade_config;

  if _upgrade_config is null then
    -- return empty set
    return query select * from app.license_pack where true = false;
  end if;

  -- reduce upgrade config to possible keys
  _upgrade_keys := '{}'::text;
  _upgrade_paths := coalesce(_upgrade_config.upgrade_paths, '{}'::app.upgrade_path[]);
  foreach _upgrade_path in array(_upgrade_paths)
  loop
    _upgrade_keys := array_append(_upgrade_keys, _upgrade_path.license_pack_key);
  end loop;

  -- return published license paths from upgrade config
  return query
  select *
  from app.license_pack
  where key = any(_upgrade_keys)
  and availability = 'published'
  ;

END
$$;


ALTER FUNCTION app.app_tenant_subscription_available_upgrade_paths(_app_tenant_subscription app.app_tenant_subscription) OWNER TO postgres;

--
-- Name: license; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.license (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    app_tenant_id text NOT NULL,
    subscription_id text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    external_id text,
    name text,
    license_type_key text NOT NULL,
    assigned_to_app_user_id text,
    inactive boolean DEFAULT false NOT NULL,
    expiration_date date,
    status app.license_status NOT NULL,
    status_reason app.license_status_reason DEFAULT 'initial'::app.license_status_reason NOT NULL,
    comment text
);


ALTER TABLE app.license OWNER TO postgres;

--
-- Name: app_user_active_licenses(app.app_user); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.app_user_active_licenses(_app_user app.app_user) RETURNS SETOF app.license
    LANGUAGE plpgsql STABLE SECURITY DEFINER
    AS $$
DECLARE
  _all_results app.subscription_available_license_type[];
BEGIN
  return query
  select *
  from app.license
  where assigned_to_app_user_id = _app_user.id
  and inactive is false
  and expiration_date > current_timestamp
  ;
END
$$;


ALTER FUNCTION app.app_user_active_licenses(_app_user app.app_user) OWNER TO postgres;

--
-- Name: app_user_home_path(app.app_user); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.app_user_home_path(_app_user app.app_user) RETURNS text
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
  _license app.license;
  _path text;
BEGIN

  select l.*
  into _license
  from app.license l
  where l.assigned_to_app_user_id = _app_user.id
  and l.inactive = false
  order by l.created_at
  limit 1
  ;

  _path := case
    when _app_user.permission_key = 'SuperAdmin' then 'AppTenants'
    else 'Home'
  end;

  return _path;
END
$$;


ALTER FUNCTION app.app_user_home_path(_app_user app.app_user) OWNER TO postgres;

--
-- Name: app_user_permissions(app.app_user, text); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.app_user_permissions(_app_user app.app_user, _current_app_tenant_id text) RETURNS text[]
    LANGUAGE plpgsql STABLE
    AS $$
  DECLARE
    _permissions text[];
    _tenant_permissions text[];
  BEGIN
    with keys as (
      select distinct lp.permission_key
      from app.license_permission lp
      join app.license l on l.id = lp.license_id
      join app.app_user u on u.id = l.assigned_to_app_user_id
      where (u.id = _app_user.id)
      and l.inactive = false
      -- and u.inactive = false
      group by lp.permission_key
    )
    select coalesce(array_agg(permission_key), '{}'::text[])
    into _permissions
    from keys
    ;

    -- permissions for app tenant licenses (vision-library, etc.)
    with keys as (
      select distinct ltp.permission_key
      from app.license_type_permission ltp
      join app.license_type lt on lt.key = ltp.license_type_key
      join app.license_pack_license_type lplt on lplt.license_type_key = lt.key
      join app.license_pack lp on lp.id = lplt.license_pack_id
      join app.app_tenant_subscription ats on ats.license_pack_id = lp.id
      where (
        ats.app_tenant_id = _current_app_tenant_id 
          or 
        ats.id = (select anchor_subscription_id from app.app_tenant where id = _app_user.app_tenant_id)
      )      
      and ats.inactive = false
      and lt.permission_key = 'Tenant'
      group by ltp.permission_key
    )
    select coalesce(array_agg(permission_key), '{}'::text[])
    into _tenant_permissions
    from keys
    where permission_key != all(_permissions)
    ;
    _permissions := array_cat(_permissions, _tenant_permissions);

    if _app_user.permission_key = 'SuperAdmin' then
      -- _permissions := array_append(_permissions, 'p:super-admin');
      _permissions := array_append(_permissions, 'm:admin');
      _permissions := array_append(_permissions, 'p:app-tenant-scope');
      _permissions := array_append(_permissions, 'p:manage-subsidiaries');
      _permissions := array_append(_permissions, 'p:demo');
      _permissions := array_append(_permissions, 'p:create-announcement');
    end if;

    if _app_user.permission_key = 'Admin' then
      _permissions := array_append(_permissions, 'p:demo');
      _permissions := array_append(_permissions, 'm:admin');
      _permissions := array_append(_permissions, 'p:create-announcement');
      _permissions := array_append(_permissions, 'p:admin-subsidiaries');
    end if;

    if _app_user.permission_key = 'Support' then
      _permissions := array_append(_permissions, 'm:admin');
      _permissions := array_append(_permissions, 'p:app-tenant-scope');
      _permissions := array_append(_permissions, 'p:support');
      _permissions := array_append(_permissions, 'p:create-announcement');
      _permissions := array_append(_permissions, 'p:admin-subsidiaries');
    end if;

    if _app_user.permission_key = 'Demo' then
      _permissions := array_append(_permissions, 'p:demo');
      _permissions := array_append(_permissions, 'm:admin');
      _permissions := array_append(_permissions, 'p:app-tenant-scope');
      _permissions := array_append(_permissions, 'p:create-announcement');
    end if;

    return _permissions;
  END
  $$;


ALTER FUNCTION app.app_user_permissions(_app_user app.app_user, _current_app_tenant_id text) OWNER TO postgres;

--
-- Name: app_user_primary_license(app.app_user); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.app_user_primary_license(_app_user app.app_user) RETURNS app.license
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
  _license app.license;
BEGIN

  raise notice '_app_user.id: %', _app_user.id;

  select l.*
  into _license
  from app.license l
  where l.assigned_to_app_user_id = _app_user.id
  and l.inactive = false
  order by l.created_at
  limit 1
  ;

  return _license;
END
$$;


ALTER FUNCTION app.app_user_primary_license(_app_user app.app_user) OWNER TO postgres;

--
-- Name: application_no_delete(); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.application_no_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   RAISE EXCEPTION 'You may not delete the application!';
END; $$;


ALTER FUNCTION app.application_no_delete() OWNER TO postgres;

--
-- Name: calculate_license_status(app.license); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.calculate_license_status(_license app.license) RETURNS app.license_status
    LANGUAGE plpgsql STABLE
    AS $$
  DECLARE
    _license_status app.license_status;
  BEGIN
    if _license.status = 'void' then
      return 'void';
    end if;

    if _license.expiration_date < current_date then
      return 'expired';
    end if;

    -- if _license.inactive or _license.status = 'inactive' then
    if _license.inactive then
      return 'inactive';
    end if;

    return 'active';
  END
  $$;


ALTER FUNCTION app.calculate_license_status(_license app.license) OWNER TO postgres;

--
-- Name: contact; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.contact (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    app_tenant_id text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    status app.contact_status DEFAULT 'active'::app.contact_status NOT NULL,
    type app.contact_type DEFAULT 'individual'::app.contact_type NOT NULL,
    organization_id text,
    location_id text,
    external_id text,
    first_name text,
    last_name text,
    email text,
    cell_phone text,
    office_phone text,
    title text,
    nickname text
);


ALTER TABLE app.contact OWNER TO postgres;

--
-- Name: contact_full_name(app.contact); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.contact_full_name(_contact app.contact) RETURNS text
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
  _full_name text;
BEGIN
  _full_name := _contact.first_name || ' ' || _contact.last_name;

  return _full_name;
END
$$;


ALTER FUNCTION app.contact_full_name(_contact app.contact) OWNER TO postgres;

--
-- Name: contact_has_unanswered_messages(app.contact); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.contact_has_unanswered_messages(_contact app.contact) RETURNS boolean
    LANGUAGE plpgsql STABLE
    AS $$
  DECLARE
    _message msg.message;
    _last_message msg.message;
    _result boolean;
  BEGIN
    for _message in
      select m.*
      from msg.subscription s
      join msg.topic t on s.topic_id = t.id
      join msg.message m on m.topic_id = t.id
      where s.status = 'active'
      and s.subscriber_contact_id = _contact.id
      and m.posted_by_contact_id = _contact.id
      order by m.created_at desc
      limit 1
    loop
      select m.*
      into _last_message
      from msg.message m
      where m.topic_id = _message.topic_id
      order by m.created_at desc
      limit 1;

      if _last_message.posted_by_contact_id = _message.posted_by_contact_id then
        return true;
      end if;

    end loop;

    return false;
  END
  $$;


ALTER FUNCTION app.contact_has_unanswered_messages(_contact app.contact) OWNER TO postgres;

--
-- Name: fn_ensure_license_status(); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.fn_ensure_license_status() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
  BEGIN
    NEW.status := app.calculate_license_status(NEW);
    -- raise exception '%', NEW;
    RETURN NEW;
  END; $$;


ALTER FUNCTION app.fn_ensure_license_status() OWNER TO postgres;

--
-- Name: fn_update_eula_trigger(); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.fn_update_eula_trigger() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
  BEGIN
    if OLD.content != NEW.content then
      raise exception 'cannot update content of eula';
    end if;

    if NEW.is_inactive != true or NEW.deactivated_at is null then
      raise exception 'the only update allowed for eula is deactivation';
    end if;

    RETURN NEW;
  END; $$;


ALTER FUNCTION app.fn_update_eula_trigger() OWNER TO postgres;

--
-- Name: license_can_activate(app.license); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.license_can_activate(_license app.license) RETURNS boolean
    LANGUAGE plpgsql STABLE
    AS $$
  DECLARE
    _can_activate boolean;
  BEGIN

    select ats.inactive = false
    into _can_activate
    from app.app_tenant_subscription ats
    where ats.id = _license.subscription_id
    ;

    return _can_activate;
  END
  $$;


ALTER FUNCTION app.license_can_activate(_license app.license) OWNER TO postgres;

--
-- Name: license_pack_allowed_actions(app.license_pack); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.license_pack_allowed_actions(_license_pack app.license_pack) RETURNS text[]
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
  _allowed_actions text[];
  _draft_exists boolean;
  _published_exists boolean;
BEGIN
  _allowed_actions := '{}'::text[];

  select
    (count(*) > 0)
  into _draft_exists
  from app.license_pack
  where key = _license_pack.key
  and availability = 'draft';

  select
    (count(*) > 0)
  into _published_exists
  from app.license_pack
  where key = _license_pack.key
  and availability = 'published';

  if _license_pack.availability = 'draft' then
    if _published_exists then
      _allowed_actions := '{"publish-confirm", "discard", "edit"}';
    else
      _allowed_actions := '{"publish", "discard", "edit"}';
    end if;
  elsif _license_pack.availability = 'published' then
    if _draft_exists = true then
      _allowed_actions := '{"discontinue"}';
    else
      _allowed_actions := '{"discontinue", "clone"}';
    end if;
  elsif _license_pack.availability = 'discontinued' then
    if _draft_exists = true then
      _allowed_actions := '{}';
    else
      _allowed_actions := '{"clone"}';
    end if;
  end if;

  return _allowed_actions;
END
$$;


ALTER FUNCTION app.license_pack_allowed_actions(_license_pack app.license_pack) OWNER TO postgres;

--
-- Name: license_pack_candidate_add_on_keys(app.license_pack); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.license_pack_candidate_add_on_keys(_license_pack app.license_pack) RETURNS text[]
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
  _candidate_add_on_keys text[];
BEGIN

  with ao as(
    select lp.key
    from app.license_pack lp
    where lp.type = 'addon'
    and lp.availability != 'discontinued'
    order by lp.key
  )
  ,fao as (
    select *
    from ao
    where key not in (select unnest(_license_pack.available_add_on_keys))
  )
  select array_agg(fao.key)
  into _candidate_add_on_keys
  from fao
  ;

  return _candidate_add_on_keys;
END
$$;


ALTER FUNCTION app.license_pack_candidate_add_on_keys(_license_pack app.license_pack) OWNER TO postgres;

--
-- Name: license_pack_candidate_license_type_keys(app.license_pack); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.license_pack_candidate_license_type_keys(_license_pack app.license_pack) RETURNS text[]
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
  _candidate_license_type_keys text[];
BEGIN

  _candidate_license_type_keys := '{}'::text[];

  select array_agg(lt.key)
  into _candidate_license_type_keys
  from app.license_type lt
  where key not in (
    select license_type_key
    from app.license_pack_license_type
    where license_pack_id = _license_pack.id
  )
  ;

  return _candidate_license_type_keys;
END
$$;


ALTER FUNCTION app.license_pack_candidate_license_type_keys(_license_pack app.license_pack) OWNER TO postgres;

--
-- Name: license_pack_candidate_upgrade_path_keys(app.license_pack); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.license_pack_candidate_upgrade_path_keys(_license_pack app.license_pack) RETURNS text[]
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
  _upgrade_path app.upgrade_path;
  _current_upgrade_path_keys text[];
  _candidate_upgrade_path_keys text[];
BEGIN

  _current_upgrade_path_keys := '{}'::text[];
  foreach _upgrade_path in array((_license_pack.upgrade_config).upgrade_paths)
  loop
    _current_upgrade_path_keys := array_append(_current_upgrade_path_keys, _upgrade_path.license_pack_key);
  end loop;

  with anchors as(
    select distinct key
    from app.license_pack
    where type = 'anchor'
    and availability != 'discontinued'
    and key != _license_pack.key
  )
  select array_agg(anchors.key)
  into _candidate_upgrade_path_keys
  from anchors
  where key not in (select unnest(_current_upgrade_path_keys))
  ;

  return _candidate_upgrade_path_keys;
END
$$;


ALTER FUNCTION app.license_pack_candidate_upgrade_path_keys(_license_pack app.license_pack) OWNER TO postgres;

--
-- Name: license_pack_discontinued_add_ons(app.license_pack); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.license_pack_discontinued_add_ons(_license_pack app.license_pack) RETURNS app.license_pack[]
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
  _discontinued_add_ons app.license_pack[];
BEGIN
  with p as (
    select lp.*
    from app.license_pack lp
    where lp.key = any(_license_pack.available_add_on_keys)
    and lp.availability in ('discontinued')
    and lp.key not in (
      select key from app.license_pack_draft_add_ons(_license_pack)
    )
    and lp.key not in (
      select key from app.license_pack_published_add_ons(_license_pack)
    )
    order by lp.key
  )
  select array_agg(p.*)
  into _discontinued_add_ons
  from p
  ;

  return _discontinued_add_ons;

END
$$;


ALTER FUNCTION app.license_pack_discontinued_add_ons(_license_pack app.license_pack) OWNER TO postgres;

--
-- Name: license_pack_draft_add_ons(app.license_pack); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.license_pack_draft_add_ons(_license_pack app.license_pack) RETURNS app.license_pack[]
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
  _draft_add_ons app.license_pack[];
BEGIN
  with p as (
    select lp.*
    from app.license_pack lp
    where lp.key = any(_license_pack.available_add_on_keys)
    and lp.availability = 'draft'
    order by lp.key
  )
  select array_agg(p.*)
  into _draft_add_ons
  from p
  ;

  return _draft_add_ons;

END
$$;


ALTER FUNCTION app.license_pack_draft_add_ons(_license_pack app.license_pack) OWNER TO postgres;

--
-- Name: license_pack_published_add_ons(app.license_pack); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.license_pack_published_add_ons(_license_pack app.license_pack) RETURNS app.license_pack[]
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
  _published_add_ons app.license_pack[];
BEGIN
  with p as (
    select lp.*
    from app.license_pack lp
    where lp.key = any(_license_pack.available_add_on_keys)
    and lp.availability = 'published'
    order by lp.key
  )
  select array_agg(p.*)
  into _published_add_ons
  from p
  ;

  return _published_add_ons;

END
$$;


ALTER FUNCTION app.license_pack_published_add_ons(_license_pack app.license_pack) OWNER TO postgres;

--
-- Name: license_pack_published_implicit_add_ons(app.license_pack); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.license_pack_published_implicit_add_ons(_license_pack app.license_pack) RETURNS app.license_pack[]
    LANGUAGE plpgsql STABLE
    AS $$
  DECLARE
    _published_implicit_add_ons app.license_pack[];
  BEGIN
    with p as (
      select lp.*
      from app.license_pack lp
      where lp.key = any(_license_pack.available_implicit_add_on_keys)
      and lp.availability = 'published'
      order by lp.key
    )
    select array_agg(p.*)
    into _published_implicit_add_ons
    from p
    ;

    return _published_implicit_add_ons;

  END
  $$;


ALTER FUNCTION app.license_pack_published_implicit_add_ons(_license_pack app.license_pack) OWNER TO postgres;

--
-- Name: license_pack_siblings(app.license_pack); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.license_pack_siblings(_license_pack app.license_pack) RETURNS app.license_pack_sibling_set
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
  _sibling_set app.license_pack_sibling_set;
  _published app.license_pack;
  _draft app.license_pack;
  _discontinued app.license_pack[];
BEGIN

  select * into _published from app.license_pack where key = _license_pack.key and availability = 'published';
  select * into _draft from app.license_pack where key = _license_pack.key and availability = 'draft';

  with d as (
    select lp.*
    from app.license_pack lp
    where lp.key = _license_pack.key
    and lp.availability = 'discontinued'
    order by discontinued_at desc
  )
  select coalesce(array_agg(d.*::app.license_pack), '{}'::app.license_pack[]) into _discontinued from d;

  _sibling_set.published = _published;
  _sibling_set.draft = _draft;
  _sibling_set.discontinued = _discontinued;

  return _sibling_set;
END
$$;


ALTER FUNCTION app.license_pack_siblings(_license_pack app.license_pack) OWNER TO postgres;

--
-- Name: set_app_tenant_setting_to_default(text); Type: FUNCTION; Schema: app; Owner: postgres
--

CREATE FUNCTION app.set_app_tenant_setting_to_default(_app_tenant_id text) RETURNS app.app_tenant
    LANGUAGE plpgsql
    AS $$
  DECLARE
    _default_settings jsonb;
    _application_setting_configs app.application_setting_config[];
    _application_setting_config app.application_setting_config;
    _asc_jsonb jsonb;
    _app_tenant app.app_tenant;
  BEGIN
    -- select to_jsonb(settings) into _default_settings from app.app_tenant where id = _app_tenant_id;
    -- _default_settings := '{}';

    -- foreach _application_setting_config in array(_application_setting_configs)
    -- loop
    --   _asc_jsonb := to_jsonb(_application_setting_config);
    --   _default_settings := _default_settings || jsonb_build_object(_asc_jsonb->>'key', _asc_jsonb->>'default_value');
    -- end loop;

    -- update app.app_tenant set settings = _default_settings where id = _app_tenant_id returning * into _app_tenant;

    return _app_tenant;
  END
  $$;


ALTER FUNCTION app.set_app_tenant_setting_to_default(_app_tenant_id text) OWNER TO postgres;

--
-- Name: authenticate_bootstrap(text); Type: FUNCTION; Schema: auth_bootstrap; Owner: postgres
--

CREATE FUNCTION auth_bootstrap.authenticate_bootstrap(_username text) RETURNS auth_bootstrap.jwt_token_bootstrap
    LANGUAGE plpgsql STABLE STRICT SECURITY DEFINER
    AS $$
  DECLARE
    _app_user app.app_user;
    _token auth_bootstrap.jwt_token_bootstrap;
  BEGIN
    select * into _app_user from app.app_user where username = _username;

    _token.current_app_user := (select auth_fn_private.do_get_app_user_info(_app_user.username, _app_user.app_tenant_id));
    _token.role := 'app_usr';
    _token.app_user_id := (_token.current_app_user).app_user_id;
    _token.app_tenant_id := (_token.current_app_user).app_tenant_id;
    _token.permissions := array_to_string((_token.current_app_user).permissions, ',');

    return _token;

  end;
  $$;


ALTER FUNCTION auth_bootstrap.authenticate_bootstrap(_username text) OWNER TO postgres;

--
-- Name: bs_users(); Type: FUNCTION; Schema: auth_bootstrap; Owner: postgres
--

CREATE FUNCTION auth_bootstrap.bs_users() RETURNS SETOF app.app_user
    LANGUAGE plpgsql STABLE STRICT SECURITY DEFINER
    AS $$
  DECLARE
    _token auth_bootstrap.jwt_token_bootstrap;
  BEGIN
    return query
    select *
    from app.app_user;
  end;
  $$;


ALTER FUNCTION auth_bootstrap.bs_users() OWNER TO postgres;

--
-- Name: app_user_has_access(text); Type: FUNCTION; Schema: auth_fn; Owner: postgres
--

CREATE FUNCTION auth_fn.app_user_has_access(_app_tenant_id text) RETURNS boolean
    LANGUAGE plpgsql STRICT SECURITY DEFINER
    AS $$
  DECLARE
    _has_super_admin boolean;
    _usr_app_tenant_id text;
    _retval boolean;
  BEGIN
    _usr_app_tenant_id := (select auth_fn.current_app_user()->>'app_tenant_id');
    _has_super_admin := (select auth_fn.app_user_has_permission('p:super-admin'));

    _retval := case
      when _has_super_admin = true then
        true
      else
        _usr_app_tenant_id = _app_tenant_id
    end;

    RETURN _retval;
  end;
  $$;


ALTER FUNCTION auth_fn.app_user_has_access(_app_tenant_id text) OWNER TO postgres;

--
-- Name: FUNCTION app_user_has_access(_app_tenant_id text); Type: COMMENT; Schema: auth_fn; Owner: postgres
--

COMMENT ON FUNCTION auth_fn.app_user_has_access(_app_tenant_id text) IS 'Verify if a user has access to an entity via the app_tenant_id';


--
-- Name: app_user_has_game_library(text); Type: FUNCTION; Schema: auth_fn; Owner: postgres
--

CREATE FUNCTION auth_fn.app_user_has_game_library(_game_id text) RETURNS boolean
    LANGUAGE plpgsql STRICT SECURITY DEFINER
    AS $$
  DECLARE
    _tenant_identifier text;
    _tenant_permissions text[];
    _game_permissions text[];
    _retval boolean;
  BEGIN
    select identifier into _tenant_identifier from app.app_tenant where id = auth_fn.current_app_user()->>'app_tenant_id';
    if _tenant_identifier = 'anchor' then
      return true;
    end if;

    with keys as (
      select distinct ltp.permission_key
      from app.license_type_permission ltp
      join app.license_type lt on lt.key = ltp.license_type_key
      join app.license_pack_license_type lplt on lplt.license_type_key = lt.key
      join app.license_pack lp on lp.id = lplt.license_pack_id
      join app.app_tenant_subscription ats on ats.license_pack_id = lp.id
      where ats.app_tenant_id = auth_fn.current_app_user()->>'app_tenant_id'
      and ats.inactive = false
      and lt.permission_key = 'Tenant'
      group by ltp.permission_key
    )
    select array_agg(permission_key)
    into _tenant_permissions
    from keys
    ;

    _retval := _tenant_permissions && _game_permissions;

    RETURN _retval;
  end;
  $$;


ALTER FUNCTION auth_fn.app_user_has_game_library(_game_id text) OWNER TO postgres;

--
-- Name: app_user_has_library(text); Type: FUNCTION; Schema: auth_fn; Owner: postgres
--

CREATE FUNCTION auth_fn.app_user_has_library(_library_id text) RETURNS boolean
    LANGUAGE plpgsql STRICT SECURITY DEFINER
    AS $$
  DECLARE
    _tenant_identifier text;
    _tenant_permissions text[];
    _library_permission text;
    _retval boolean;
  BEGIN

    _library_permission := 'l:' || _library_id;
    _retval := (regexp_split_to_array(auth_fn.current_app_user()->>'permissions',','))@>(array[_library_permission]);

    RETURN _retval;
  end;
  $$;


ALTER FUNCTION auth_fn.app_user_has_library(_library_id text) OWNER TO postgres;

--
-- Name: app_user_has_licensing_scope(text); Type: FUNCTION; Schema: auth_fn; Owner: postgres
--

CREATE FUNCTION auth_fn.app_user_has_licensing_scope(_app_tenant_id text) RETURNS boolean
    LANGUAGE plpgsql STRICT SECURITY DEFINER
    AS $$
  DECLARE
    _app_permission_key app.permission_key;
    _ls_app_tenant_id_json jsonb;
    _ls_app_tenant_id text;
  BEGIN
    if auth_fn.current_app_user()->>'permission_key' in ('SuperAdmin', 'Support') then
      return true;
    end if;

    for _ls_app_tenant_id_json in (select jsonb_array_elements(auth_fn.current_app_user()->'licensing_scope'))
    loop
      _ls_app_tenant_id := (select _ls_app_tenant_id_json #>> '{}');
      if _ls_app_tenant_id = _app_tenant_id then
        return true;
      end if;
    end loop;

    return false;
  end;
  $$;


ALTER FUNCTION auth_fn.app_user_has_licensing_scope(_app_tenant_id text) OWNER TO postgres;

--
-- Name: app_user_has_permission(text); Type: FUNCTION; Schema: auth_fn; Owner: postgres
--

CREATE FUNCTION auth_fn.app_user_has_permission(_permission text) RETURNS boolean
    LANGUAGE plpgsql STRICT SECURITY DEFINER
    AS $$
  DECLARE
    _retval boolean;

  BEGIN
    _retval := position(_permission in auth_fn.current_app_user()->>'permissions') > 0;

    RETURN _retval;
  end;
  $$;


ALTER FUNCTION auth_fn.app_user_has_permission(_permission text) OWNER TO postgres;

--
-- Name: app_user_has_permission_key(app.permission_key); Type: FUNCTION; Schema: auth_fn; Owner: postgres
--

CREATE FUNCTION auth_fn.app_user_has_permission_key(_permission_key app.permission_key) RETURNS boolean
    LANGUAGE plpgsql STRICT SECURITY DEFINER
    AS $$
  DECLARE
  BEGIN
    if auth_fn.current_app_user()->>'permission_key' = _permission_key::text then
      return true;
    else
      return false;
    end if;
  end;
  $$;


ALTER FUNCTION auth_fn.app_user_has_permission_key(_permission_key app.permission_key) OWNER TO postgres;

--
-- Name: auth_0_pre_registration(text); Type: FUNCTION; Schema: auth_fn; Owner: postgres
--

CREATE FUNCTION auth_fn.auth_0_pre_registration(_email text) RETURNS auth_fn.app_user_auth0_info
    LANGUAGE plpgsql STABLE SECURITY DEFINER
    AS $$
  DECLARE
    _err_context text;
    _app_user app.app_user;
    _app_tenant_subscription app.app_tenant_subscription;
    _app_user_auth0_info auth_fn.app_user_auth0_info;
  BEGIN

    select * into _app_user from app.app_user where lower(recovery_email) = lower(trim(both from _email, ' '));

    if _app_user.id is null then
      select * into _app_tenant_subscription from app.app_tenant_subscription where id = (
        select anchor_subscription_id from app.app_tenant where type = 'anchor'
      );

       _app_user_auth0_info := (select app_fn.create_new_licensed_app_user(
        '%-guest-user'::text,
        _app_tenant_subscription.id::text,
        row(
          _email -- username text
          ,_app_tenant_subscription.app_tenant_id -- ,app_tenant_id text
          ,null -- ,ext_auth_id text
          ,null -- ,ext_crm_id text
          ,row(
            null -- organization_id text
            ,null -- ,location_id text
            ,_email-- ,email text
            ,null -- ,first_name text
            ,null -- ,last_name text
            ,null -- ,cell_phone text
            ,null -- ,office_phone text
            ,null -- ,title text
            ,null -- ,nickname text
            ,null -- ,external_id text
          )::app_fn.create_contact_input
        )::app_fn.new_app_user_info
      ));
    else
      -- prolly want to change to auth_fn_private.do_get_app_user_info
      _app_user_auth0_info := (select auth_fn.get_app_user_info(_app_user.recovery_email, _app_user.app_tenant_id));
    end if;

    return _app_user_auth0_info;
    exception
      when others then
        GET STACKED DIAGNOSTICS _err_context = PG_EXCEPTION_CONTEXT;
        if position('FB' in SQLSTATE::text) = 0 then
          _err_context := 'auth_fn.auth_0_pre_registration:::' || SQLSTATE::text || ':::' || SQLERRM::text || ':::' || _err_context;
          raise exception '%', _err_context using errcode = 'FB500';
        end if;
        raise;
  end;
  $$;


ALTER FUNCTION auth_fn.auth_0_pre_registration(_email text) OWNER TO postgres;

--
-- Name: FUNCTION auth_0_pre_registration(_email text); Type: COMMENT; Schema: auth_fn; Owner: postgres
--

COMMENT ON FUNCTION auth_fn.auth_0_pre_registration(_email text) IS '@omit';


--
-- Name: current_app_user(); Type: FUNCTION; Schema: auth_fn; Owner: postgres
--

CREATE FUNCTION auth_fn.current_app_user() RETURNS jsonb
    LANGUAGE plpgsql STABLE
    AS $$
  DECLARE
  BEGIN
  return current_setting('jwt.claims.current_app_user')::jsonb;
  end;
  $$;


ALTER FUNCTION auth_fn.current_app_user() OWNER TO postgres;

--
-- Name: current_app_user_id(); Type: FUNCTION; Schema: auth_fn; Owner: postgres
--

CREATE FUNCTION auth_fn.current_app_user_id() RETURNS text
    LANGUAGE plpgsql STRICT SECURITY DEFINER
    AS $$
  DECLARE
  BEGIN
    return current_setting('jwt.claims.app_user_id')::text;
  end;
  $$;


ALTER FUNCTION auth_fn.current_app_user_id() OWNER TO postgres;

--
-- Name: get_app_tenant_scope_permissions(text); Type: FUNCTION; Schema: auth_fn; Owner: postgres
--

CREATE FUNCTION auth_fn.get_app_tenant_scope_permissions(_app_tenant_id text) RETURNS text[]
    LANGUAGE plpgsql STABLE
    AS $$
  DECLARE
    _err_context text;
    _permissions text[];
    _tenant_permissions text[];
    _app_tenant app.app_tenant;
    _anchor_tenant_id text;
  BEGIN
    select * into _app_tenant from app.app_tenant where id = _app_tenant_id;
    select id into _anchor_tenant_id from app.app_tenant where type = 'anchor';
    if _app_tenant.parent_app_tenant_id is not null and _app_tenant.parent_app_tenant_id != _anchor_tenant_id then
      _permissions := (select auth_fn.get_app_tenant_scope_permissions(_app_tenant.parent_app_tenant_id));
      return _permissions;
    end if;

    -- with keys as (
    --   select distinct ltp.permission_key
    --   from app.license_type_permission ltp
    --   join app.license_type lt on lt.key = ltp.license_type_key
    --   join app.license_pack_license_type lplt on lplt.license_type_key = lt.key
    --   join app.license_pack lp on lp.id = lplt.license_pack_id
    --   join app.app_tenant_subscription ats on ats.license_pack_id = lp.id
    --   join app.app_tenant apt on apt.anchor_subscription_id = ats.id
    --   where apt.id = _app_tenant_id
    --   and lplt.assign_upon_subscription = true
    --   group by ltp.permission_key
    -- )
    -- select array_agg(permission_key)
    -- into _permissions
    -- from keys
    -- ;
    with keys as (
      select distinct ltp.permission_key
      from app.license_type_permission ltp
      join app.license_type lt on lt.key = ltp.license_type_key
      join app.license_pack_license_type lplt on lplt.license_type_key = lt.key
      join app.license_pack lp on lp.id = lplt.license_pack_id
      join app.app_tenant_subscription ats on ats.license_pack_id = lp.id
      where ats.app_tenant_id = _app_tenant_id
      and lp.type = 'anchor'
      and ats.inactive = false
      and lplt.assign_upon_subscription = true
      group by ltp.permission_key
    )
    select array_agg(permission_key)
    into _permissions
    from keys
    ;

    with keys as (
      select distinct ltp.permission_key
      from app.license_type_permission ltp
      join app.license_type lt on lt.key = ltp.license_type_key
      join app.license_pack_license_type lplt on lplt.license_type_key = lt.key
      join app.license_pack lp on lp.id = lplt.license_pack_id
      join app.app_tenant_subscription ats on ats.license_pack_id = lp.id
      where ats.app_tenant_id = _app_tenant_id
      and ats.inactive = false
      and lt.permission_key = 'Tenant'
      group by ltp.permission_key
    )
    select array_agg(permission_key)
    into _tenant_permissions
    from keys
    ;
    _permissions := array_cat(_permissions, _tenant_permissions);

    _permissions := array_append(_permissions, 'm:admin');

    return _permissions;
    exception
      when others then
        GET STACKED DIAGNOSTICS _err_context = PG_EXCEPTION_CONTEXT;
        if position('FB' in SQLSTATE::text) = 0 then
          _err_context := 'auth_fn.get_app_tenant_scope_permissions:::' || SQLSTATE::text || ':::' || SQLERRM::text || ':::' || _err_context;
          raise exception '%', _err_context using errcode = 'FB500';
        end if;
        raise;
  end;
  $$;


ALTER FUNCTION auth_fn.get_app_tenant_scope_permissions(_app_tenant_id text) OWNER TO postgres;

--
-- Name: FUNCTION get_app_tenant_scope_permissions(_app_tenant_id text); Type: COMMENT; Schema: auth_fn; Owner: postgres
--

COMMENT ON FUNCTION auth_fn.get_app_tenant_scope_permissions(_app_tenant_id text) IS '@omit';


--
-- Name: get_app_user_info(text, text); Type: FUNCTION; Schema: auth_fn; Owner: postgres
--

CREATE FUNCTION auth_fn.get_app_user_info(_recovery_email_or_id_or_username text, _current_app_tenant_id text) RETURNS auth_fn.app_user_auth0_info
    LANGUAGE plpgsql STABLE SECURITY DEFINER
    AS $$
  DECLARE
    _err_context text;
    _app_user_auth0_info auth_fn.app_user_auth0_info;
    _app_user app.app_user;
    _working_app_tenant_id text;
    _permission text;
    _permissions text[];
    _has_permission boolean;
  BEGIN
    -- enforce usage of this function explicitly
    --
    select * into _app_user from app.app_user usr
    where usr.recovery_email = _recovery_email_or_id_or_username or usr.id = _recovery_email_or_id_or_username or usr.username = _recovery_email_or_id_or_username
    ;
    if _app_user.id is null then
      raise exception 'no app user';
    end if;

    _has_permission := false;

    if auth_fn.current_app_user()->>'app_user_id' = _app_user.id then
      _has_permission := true;
    elsif auth_fn.app_user_has_permission('p:super-admin') or auth_fn.current_app_user()->>'permission_key' in ('SuperAdmin', 'Support') then
      _has_permission := true;
    elsif (auth_fn.app_user_has_permission('p:demo') and _recovery_email_or_id_or_username = 'fnb-demo') then
      _has_permission := true;
    elsif (auth_fn.app_user_has_permission('p:app-tenant-group-admin') and _recovery_email_or_id_or_username = 'fnb-support') then
      _has_permission := true;
    elsif auth_fn.app_user_has_permission('p:admin') then
      _has_permission := (select auth_fn.app_user_has_licensing_scope(_app_user.app_tenant_id));
    end if;

    if _has_permission = false or _has_permission is null then
      raise exception 'permission denied';
    end if;

    _app_user_auth0_info := (select auth_fn_private.do_get_app_user_info(
      _recovery_email_or_id_or_username
      ,_current_app_tenant_id
    ))
    ;
    -- raise exception '_current_app_tenant_id: %, _app_user_auth_0.app_tenant_id, %', _current_app_tenant_id, _app_user_auth0_info.app_tenant_id;
    return _app_user_auth0_info;

    exception
      when others then
        GET STACKED DIAGNOSTICS _err_context = PG_EXCEPTION_CONTEXT;
        if position('FB' in SQLSTATE::text) = 0 then
          _err_context := 'auth_fn.get_app_user_info:::' || SQLSTATE::text || ':::' || SQLERRM::text || ':::' || _err_context;
          raise exception '%', _err_context using errcode = 'FB500';
        end if;
        raise;
  end;  
  $$;


ALTER FUNCTION auth_fn.get_app_user_info(_recovery_email_or_id_or_username text, _current_app_tenant_id text) OWNER TO postgres;

--
-- Name: init_app_tenant_support(text); Type: FUNCTION; Schema: auth_fn; Owner: postgres
--

CREATE FUNCTION auth_fn.init_app_tenant_support(_app_tenant_id text) RETURNS auth_fn.app_user_auth0_info
    LANGUAGE plpgsql STABLE SECURITY DEFINER
    AS $$
  DECLARE
    _err_context text;
    _app_user_auth0_info auth_fn.app_user_auth0_info;
    _tenant_permissions text[];
  BEGIN
    if 
      auth_fn.app_user_has_permission('p:super-admin') = false and
      auth_fn.app_user_has_permission('p:app-tenant-group-admin') = false
    then
      raise exception 'permission denied';
    end if;
    -- if user is SuperAdmin, they can do whatever
    -- if user has permission 'p:app-tenant-group-admin' then _app_tenant_id must be in support scope for this tenant
    if 
      auth_fn.app_user_has_permission('p:super-admin') = false and
      auth_fn.app_user_has_permission('p:app-tenant-group-admin') = true
    then
      if _app_tenant_id not in (
        select t.id
        from app.app_tenant t
        join app.app_tenant_group_member atgm on t.id = atgm.app_tenant_id
        join app.app_tenant_group atg on atg.id = atgm.app_tenant_group_id
        join app.app_tenant_group_admin atga on atga.app_tenant_group_id = atg.id
        where atga.app_user_id = auth_fn.current_app_user()->>'app_user_id'
      ) 
      then
        raise exception 'permission denied - user is not a manager of tenant group';
      end if;
    end if;
    
    _app_user_auth0_info := auth_fn.get_app_user_info('fnb-support', _app_tenant_id);
    _tenant_permissions := (select auth_fn.get_app_tenant_scope_permissions(_app_tenant_id));
    _app_user_auth0_info.permissions := array_cat(_app_user_auth0_info.permissions, _tenant_permissions);

    _app_user_auth0_info.permissions := array_append(_app_user_auth0_info.permissions, 'p:clinic-patient');
    _app_user_auth0_info.permissions := array_append(_app_user_auth0_info.permissions, 'p:home-patient');
    -- remove duplicate permissions
    with perms as (
      select unnest(_app_user_auth0_info.permissions) as p
    )
    ,d as (
      select distinct p from perms
    )
    select array_agg(p) from d
    into _app_user_auth0_info.permissions
    ;

    return _app_user_auth0_info;
    exception
      when others then
        GET STACKED DIAGNOSTICS _err_context = PG_EXCEPTION_CONTEXT;
        if position('FB' in SQLSTATE::text) = 0 then
          _err_context := 'auth_fn.init_app_tenant_support:::' || SQLSTATE::text || ':::' || SQLERRM::text || ':::' || _err_context;
          raise exception '%', _err_context using errcode = 'FB500';
        end if;
        raise;
  end;
  $$;


ALTER FUNCTION auth_fn.init_app_tenant_support(_app_tenant_id text) OWNER TO postgres;

--
-- Name: init_demo(text); Type: FUNCTION; Schema: auth_fn; Owner: postgres
--

CREATE FUNCTION auth_fn.init_demo(_license_pack_key text) RETURNS auth_fn.init_demo_result
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
  DECLARE
    _err_context text;
    _demo_app_tenant app.app_tenant;
    _app_user_auth0_info auth_fn.app_user_auth0_info;
    _result auth_fn.init_demo_result;
    _tenant_permissions text[];
  BEGIN
    if auth_fn.app_user_has_permission('p:demo') = false then
      raise exception 'permission denied';
    end if;
    _demo_app_tenant := (select app_fn.create_demo_app_tenant(_license_pack_key));

    _app_user_auth0_info := auth_fn.get_app_user_info('fnb-demo', _demo_app_tenant.id);
    _tenant_permissions := (select auth_fn.get_app_tenant_scope_permissions(_demo_app_tenant.id));
    _app_user_auth0_info.permissions = array_cat(_app_user_auth0_info.permissions, _tenant_permissions);

    _result.demo_app_tenant = _demo_app_tenant;
    _result.app_user_auth0_info = _app_user_auth0_info;
    _result.license_pack_key = _license_pack_key;

    return _result;
    exception
      when others then
        GET STACKED DIAGNOSTICS _err_context = PG_EXCEPTION_CONTEXT;
        if position('FB' in SQLSTATE::text) = 0 then
          _err_context := 'auth_fn.init_demo:::' || SQLSTATE::text || ':::' || SQLERRM::text || ':::' || _err_context;
          raise exception '%', _err_context using errcode = 'FB500';
        end if;
        raise;
  end;
  $$;


ALTER FUNCTION auth_fn.init_demo(_license_pack_key text) OWNER TO postgres;

--
-- Name: init_subsidiary_admin(text); Type: FUNCTION; Schema: auth_fn; Owner: postgres
--

CREATE FUNCTION auth_fn.init_subsidiary_admin(_subsidiary_app_tenant_id text) RETURNS auth_fn.init_subsidiary_admin_result
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
  DECLARE
    _err_context text;
    _subsidiary_app_tenant app.app_tenant;
    _app_user_auth0_info auth_fn.app_user_auth0_info;
    _result auth_fn.init_subsidiary_admin_result;
    _tenant_permissions text[];
  BEGIN
    if auth_fn.app_user_has_permission('p:admin-subsidiaries') = false then
      raise exception 'permission denied';
    end if;

    select *
    into _subsidiary_app_tenant
    from app.app_tenant
    where id = _subsidiary_app_tenant_id
    ;

    _app_user_auth0_info := (select auth_fn.get_app_user_info(auth_fn.current_app_user()->>'username', _subsidiary_app_tenant_id));
    _tenant_permissions := (select auth_fn.get_app_tenant_scope_permissions(_subsidiary_app_tenant_id));
    _app_user_auth0_info.permissions = array_cat(_app_user_auth0_info.permissions, _tenant_permissions);

    _app_user_auth0_info.permissions := array_append(_app_user_auth0_info.permissions, 'p:clinic-patient');
    _app_user_auth0_info.permissions := array_append(_app_user_auth0_info.permissions, 'p:home-patient');
    -- remove duplicate permissions
    with perms as (
      select unnest(_app_user_auth0_info.permissions) as p
    )
    ,d as (
      select distinct p from perms
    )
    select array_agg(p) from d
    into _app_user_auth0_info.permissions
    ;

    _result.subsidiary_app_tenant = _subsidiary_app_tenant;
    _result.app_user_auth0_info = _app_user_auth0_info;

    return _result;
    exception
      when others then
        GET STACKED DIAGNOSTICS _err_context = PG_EXCEPTION_CONTEXT;
        if position('FB' in SQLSTATE::text) = 0 then
          _err_context := 'auth_fn.init_subsidiary_admin:::' || SQLSTATE::text || ':::' || SQLERRM::text || ':::' || _err_context;
          raise exception '%', _err_context using errcode = 'FB500';
        end if;
        raise;
  end;
  $$;


ALTER FUNCTION auth_fn.init_subsidiary_admin(_subsidiary_app_tenant_id text) OWNER TO postgres;

--
-- Name: do_get_app_user_info(text, text); Type: FUNCTION; Schema: auth_fn_private; Owner: postgres
--

CREATE FUNCTION auth_fn_private.do_get_app_user_info(_recovery_email_or_id_or_username text, _current_app_tenant_id text) RETURNS auth_fn.app_user_auth0_info
    LANGUAGE plpgsql STABLE SECURITY DEFINER
    AS $$
  DECLARE
    _err_context text;
    _app_user_auth0_info auth_fn.app_user_auth0_info;
    _app_user app.app_user;
    _working_app_tenant_id text;
    _permission text;
    _permissions text[];
    _has_permission boolean;
  BEGIN
    select
      usr.permission_key -- permission_key
      ,usr.username -- username
      ,usr.inactive -- inactive
      ,ten.name -- app_tenant_name
      ,coalesce(_current_app_tenant_id, ten.id) -- app_tenant_id
      -- ,ten.id -- app_tenant_id
      ,case
        when ten.type = 'subsidiary' then ten.parent_app_tenant_id
        else null
      end -- parent_app_tenant_id
      ,case
        when ten.type = 'customer' then (
          with s as (
            select
              id::text
              ,name::text
            from app.app_tenant
            where parent_app_tenant_id = ten.id
          )
          select array_agg(s::auth_fn.app_tenant_auth0_info)::auth_fn.app_tenant_auth0_info[]
          from s
        )
        else '{}'::auth_fn.app_tenant_auth0_info[]
      end -- subsidiaries
      ,usr.id -- app_user_id
      ,usr.preferred_timezone -- preferred_timezone
      ,c.id as contact_id -- contact_id
      ,c.first_name -- first_name
      ,c.last_name -- last_name
      ,usr.recovery_email -- recovery_email
      ,case
        when usr.permission_key = 'SuperAdmin' then 'app_sp_adm'
        when usr.permission_key = 'Admin' then 'app_adm'
        when usr.permission_key = 'Support' then 'app_adm'
        when usr.permission_key = 'Demo' then 'app_adm'
        when usr.permission_key = 'User' then 'app_usr'
      end -- app_role
      ,(select app.app_user_permissions(usr, _current_app_tenant_id)) -- permissions
      ,(select app.app_user_home_path(usr)) -- home_path
      ,'{}'::text[] -- licensing_scope
      ,usr.ext_auth_id -- ext_auth_id
      ,usr.ext_auth_blocked -- ext_auth_blocked
    into _app_user_auth0_info
    from app.app_user usr
    join app.app_tenant ten on ten.id = usr.app_tenant_id
    join app.contact c on usr.contact_id = c.id
    where usr.recovery_email = _recovery_email_or_id_or_username or usr.id = _recovery_email_or_id_or_username or usr.username = _recovery_email_or_id_or_username
    ;
    _working_app_tenant_id := coalesce(_current_app_tenant_id, _app_user_auth0_info.app_tenant_id);

    -- licensing scope
    _app_user_auth0_info.licensing_scope := array_append(_app_user_auth0_info.licensing_scope, _working_app_tenant_id);
    if _app_user_auth0_info.permission_key in ('Support', 'Admin') then
      _app_user_auth0_info.licensing_scope := array_cat(
        _app_user_auth0_info.licensing_scope
        ,(
          select coalesce(array_agg(id),'{}'::text[]) from app.app_tenant where parent_app_tenant_id = _working_app_tenant_id and type = 'subsidiary'
        )
      );
    end if;

    return _app_user_auth0_info;

    exception
      when others then
        GET STACKED DIAGNOSTICS _err_context = PG_EXCEPTION_CONTEXT;
        if position('FB' in SQLSTATE::text) = 0 then
          _err_context := 'auth_fn_private.do_get_app_user_info:::' || SQLSTATE::text || ':::' || SQLERRM::text || ':::' || _err_context;
          raise exception '%', _err_context using errcode = 'FB500';
        end if;
        raise;
  end;  
  $$;


ALTER FUNCTION auth_fn_private.do_get_app_user_info(_recovery_email_or_id_or_username text, _current_app_tenant_id text) OWNER TO postgres;

--
-- Name: FUNCTION do_get_app_user_info(_recovery_email_or_id_or_username text, _current_app_tenant_id text); Type: COMMENT; Schema: auth_fn_private; Owner: postgres
--

COMMENT ON FUNCTION auth_fn_private.do_get_app_user_info(_recovery_email_or_id_or_username text, _current_app_tenant_id text) IS '@OMIT';


--
-- Name: subscription; Type: TABLE; Schema: msg; Owner: postgres
--

CREATE TABLE msg.subscription (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    app_tenant_id text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    status msg.subscription_status DEFAULT 'active'::msg.subscription_status NOT NULL,
    topic_id text NOT NULL,
    subscriber_contact_id text NOT NULL,
    last_read timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE msg.subscription OWNER TO postgres;

--
-- Name: deactivate_subscription(text); Type: FUNCTION; Schema: msg_fn; Owner: postgres
--

CREATE FUNCTION msg_fn.deactivate_subscription(_subscription_id text) RETURNS msg.subscription
    LANGUAGE plpgsql
    AS $$
    DECLARE
      _subscription msg.subscription;
      _err_context text;
    BEGIN
      update msg.subscription set
        status = 'inactive'
      where id = _subscription_id
      returning *
      into _subscription
      ;

      return _subscription;

      exception
        when others then
          GET STACKED DIAGNOSTICS _err_context = PG_EXCEPTION_CONTEXT;
          if position('FB' in SQLSTATE::text) = 0 then
            _err_context := 'msg_fn.deactivate_subscription:::' || SQLSTATE::text || ':::' || SQLERRM::text || ':::' || _err_context;
            raise exception '%', _err_context using errcode = 'FB500';
          end if;
          raise;
    end;
    $$;


ALTER FUNCTION msg_fn.deactivate_subscription(_subscription_id text) OWNER TO postgres;

--
-- Name: message; Type: TABLE; Schema: msg; Owner: postgres
--

CREATE TABLE msg.message (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    app_tenant_id text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    status msg.message_status DEFAULT 'sent'::msg.message_status NOT NULL,
    topic_id text NOT NULL,
    content text NOT NULL,
    posted_by_contact_id text NOT NULL,
    tags text[] DEFAULT '{}'::text[] NOT NULL,
    acknowledged_at timestamp with time zone
);


ALTER TABLE msg.message OWNER TO postgres;

--
-- Name: upsert_message(msg_fn.message_info); Type: FUNCTION; Schema: msg_fn; Owner: postgres
--

CREATE FUNCTION msg_fn.upsert_message(_message_info msg_fn.message_info) RETURNS msg.message
    LANGUAGE plpgsql
    AS $$
    DECLARE
      _topic msg.topic;
      _message msg.message;
      _posted_by_contact_id text;
      _err_context text;
    BEGIN
      select contact_id into _posted_by_contact_id from app.app_user where id = auth_fn.current_app_user_id();

      select * into _topic from msg.topic where id = _message_info.topic_id;
      if _topic.id is null then
        raise exception 'no topic for id: %', _message_info.topic_id;
      end if;

      select * into _message from msg.message where id = _message_info.id;

      if _message.id is not null then
        update msg.message set
          content = _message_info.content
          ,tags = coalesce(_message_info.tags, '{}')
        where id = _message.id
        ;
      else
        insert into msg.message(
          app_tenant_id
          ,topic_id
          ,posted_by_contact_id
          ,content
          ,tags
        )
        select
          _topic.app_tenant_id
          ,_message_info.topic_id
          ,_posted_by_contact_id
          ,_message_info.content
          ,coalesce(_message_info.tags, '{}')
        returning *
        into _message
        ;
      end if;

      return _message;
      exception
        when others then
          GET STACKED DIAGNOSTICS _err_context = PG_EXCEPTION_CONTEXT;
          if position('FB' in SQLSTATE::text) = 0 then
            _err_context := 'msg_fn.upsert_message:::' || SQLSTATE::text || ':::' || SQLERRM::text || ':::' || _err_context;
            raise exception '%', _err_context using errcode = 'FB500';
          end if;
          raise;
    end;
    $$;


ALTER FUNCTION msg_fn.upsert_message(_message_info msg_fn.message_info) OWNER TO postgres;

--
-- Name: upsert_subscription(msg_fn.subscription_info); Type: FUNCTION; Schema: msg_fn; Owner: postgres
--

CREATE FUNCTION msg_fn.upsert_subscription(_subscription_info msg_fn.subscription_info) RETURNS msg.subscription
    LANGUAGE plpgsql
    AS $$
    DECLARE
      _topic msg.topic;
      _subscription msg.subscription;
      _err_context text;
    BEGIN
      select *
      into _topic
      from msg.topic
      where id = _subscription_info.topic_id
      ;
      if _topic.id is null then
        raise exception 'no topic for id: %', _subscription_info.topic_id;
      end if;

      select * into _subscription
      from msg.subscription
      where topic_id = _subscription_info.topic_id
      and subscriber_contact_id = _subscription_info.subscriber_contact_id
      ;

      if _subscription.id is not null then
        update msg.subscription set
          status = 'active'
        where id = _subscription.id
        ;
      else
        insert into msg.subscription(
          app_tenant_id
          ,topic_id
          ,subscriber_contact_id
        )
        select
          _topic.app_tenant_id
          ,_topic.id
          ,_subscription_info.subscriber_contact_id
        returning *
        into _subscription
        ;
      end if;

      return _subscription;
      exception
        when others then
          GET STACKED DIAGNOSTICS _err_context = PG_EXCEPTION_CONTEXT;
          if position('FB' in SQLSTATE::text) = 0 then
            _err_context := 'msg_fn.upsert_subscription:::' || SQLSTATE::text || ':::' || SQLERRM::text || ':::' || _err_context;
            raise exception '%', _err_context using errcode = 'FB500';
          end if;
          raise;
    end;
    $$;


ALTER FUNCTION msg_fn.upsert_subscription(_subscription_info msg_fn.subscription_info) OWNER TO postgres;

--
-- Name: topic; Type: TABLE; Schema: msg; Owner: postgres
--

CREATE TABLE msg.topic (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    app_tenant_id text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    name text NOT NULL,
    identifier text
);


ALTER TABLE msg.topic OWNER TO postgres;

--
-- Name: upsert_topic(msg_fn.topic_info); Type: FUNCTION; Schema: msg_fn; Owner: postgres
--

CREATE FUNCTION msg_fn.upsert_topic(_topic_info msg_fn.topic_info) RETURNS msg.topic
    LANGUAGE plpgsql
    AS $$
    DECLARE
      _topic msg.topic;
      _topic_id text;
      _err_context text;
    BEGIN
      _topic_id = coalesce(_topic_info.id, shard_1.id_generator());
      select *
        into _topic
      from msg.topic
      where (id = _topic_id or identifier = _topic_info.identifier)
      and app_tenant_id = auth_fn.current_app_user()->>'app_tenant_id'
      ;

      if _topic.id is not null then
        update msg.topic set
          name = _topic_info.name
        where id = _topic_id
        ;
      else
        insert into msg.topic(
          id
          ,app_tenant_id
          ,name
          ,identifier
        )
        select
          _topic_id
          ,auth_fn.current_app_user()->>'app_tenant_id'
          ,_topic_info.name
          ,_topic_info.identifier
        returning *
        into _topic
        ;
      end if;

      return _topic;
      exception
        when others then
          GET STACKED DIAGNOSTICS _err_context = PG_EXCEPTION_CONTEXT;
          if position('FB' in SQLSTATE::text) = 0 then
            _err_context := 'msg_fn.upsert_topic:::' || SQLSTATE::text || ':::' || SQLERRM::text || ':::' || _err_context;
            raise exception '%', _err_context using errcode = 'FB500';
          end if;
          raise;
    end;
    $$;


ALTER FUNCTION msg_fn.upsert_topic(_topic_info msg_fn.topic_info) OWNER TO postgres;

--
-- Name: notify_watchers_ddl(); Type: FUNCTION; Schema: postgraphile_watch; Owner: postgres
--

CREATE FUNCTION postgraphile_watch.notify_watchers_ddl() RETURNS event_trigger
    LANGUAGE plpgsql
    AS $$
begin
  perform pg_notify(
    'postgraphile_watch',
    json_build_object(
      'type',
      'ddl',
      'payload',
      (select json_agg(json_build_object('schema', schema_name, 'command', command_tag)) from pg_event_trigger_ddl_commands() as x)
    )::text
  );
end;
$$;


ALTER FUNCTION postgraphile_watch.notify_watchers_ddl() OWNER TO postgres;

--
-- Name: notify_watchers_drop(); Type: FUNCTION; Schema: postgraphile_watch; Owner: postgres
--

CREATE FUNCTION postgraphile_watch.notify_watchers_drop() RETURNS event_trigger
    LANGUAGE plpgsql
    AS $$
begin
  perform pg_notify(
    'postgraphile_watch',
    json_build_object(
      'type',
      'drop',
      'payload',
      (select json_agg(distinct x.schema_name) from pg_event_trigger_dropped_objects() as x)
    )::text
  );
end;
$$;


ALTER FUNCTION postgraphile_watch.notify_watchers_drop() OWNER TO postgres;

--
-- Name: app_exception; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.app_exception (
    err_code text NOT NULL,
    description text,
    svr_message_mask text,
    ui_message_mask text,
    CONSTRAINT app_exception_err_code_check CHECK ((err_code <> ''::text))
);


ALTER TABLE app.app_exception OWNER TO postgres;

--
-- Name: app_route; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.app_route (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    name text NOT NULL,
    application_key text NOT NULL,
    permission_key text,
    description text,
    path text NOT NULL,
    menu_behavior app.app_route_menu_behavior DEFAULT 'none'::app.app_route_menu_behavior NOT NULL,
    menu_parent_name text,
    menu_ordinal integer
);


ALTER TABLE app.app_route OWNER TO postgres;

--
-- Name: app_tenant_group; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.app_tenant_group (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    name text NOT NULL,
    support_email text DEFAULT 'help@ourvisualbrain.com'::text NOT NULL
);


ALTER TABLE app.app_tenant_group OWNER TO postgres;

--
-- Name: app_tenant_group_admin; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.app_tenant_group_admin (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    app_tenant_group_id text NOT NULL,
    app_user_id text NOT NULL
);


ALTER TABLE app.app_tenant_group_admin OWNER TO postgres;

--
-- Name: app_tenant_group_member; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.app_tenant_group_member (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    app_tenant_group_id text NOT NULL,
    app_tenant_id text NOT NULL
);


ALTER TABLE app.app_tenant_group_member OWNER TO postgres;

--
-- Name: application; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.application (
    key text NOT NULL,
    name text,
    setting_configs app.application_setting_config[]
);


ALTER TABLE app.application OWNER TO postgres;

--
-- Name: error_report; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.error_report (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    first_reported_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_reported_at timestamp with time zone NOT NULL,
    observed_count integer,
    message text,
    comment text,
    reported_by_app_user_id text,
    reported_as_app_user_id text,
    operation_name text NOT NULL,
    variables jsonb NOT NULL,
    status app.error_report_status DEFAULT 'captured'::app.error_report_status NOT NULL
);


ALTER TABLE app.error_report OWNER TO postgres;

--
-- Name: eula; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.eula (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    is_inactive boolean DEFAULT false NOT NULL,
    deactivated_at timestamp with time zone,
    content text NOT NULL,
    CONSTRAINT eula_check CHECK (((is_inactive = false) OR (deactivated_at IS NOT NULL)))
);


ALTER TABLE app.eula OWNER TO postgres;

--
-- Name: facility; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.facility (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    app_tenant_id text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    organization_id text,
    location_id text,
    name text,
    external_id text
);


ALTER TABLE app.facility OWNER TO postgres;

--
-- Name: license_permission; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.license_permission (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    app_tenant_id text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    license_id text NOT NULL,
    permission_key text NOT NULL
);


ALTER TABLE app.license_permission OWNER TO postgres;

--
-- Name: license_type; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.license_type (
    key text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    external_id text,
    name text,
    application_key text NOT NULL,
    permission_key app.permission_key DEFAULT 'User'::app.permission_key NOT NULL,
    sync_user_on_assignment boolean DEFAULT true NOT NULL
);


ALTER TABLE app.license_type OWNER TO postgres;

--
-- Name: license_type_permission; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.license_type_permission (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    license_type_key text NOT NULL,
    permission_key text NOT NULL
);


ALTER TABLE app.license_type_permission OWNER TO postgres;

--
-- Name: location; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.location (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    app_tenant_id text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    external_id text,
    name text,
    address1 text,
    address2 text,
    city text,
    state text,
    zip text,
    lat text,
    lon text
);


ALTER TABLE app.location OWNER TO postgres;

--
-- Name: module; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.module (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    application_key text NOT NULL,
    name text NOT NULL,
    route_name text NOT NULL,
    permission_key text NOT NULL
);


ALTER TABLE app.module OWNER TO postgres;

--
-- Name: note; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.note (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    app_tenant_id text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by_app_user_id text NOT NULL,
    content text NOT NULL
);


ALTER TABLE app.note OWNER TO postgres;

--
-- Name: organization; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.organization (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    app_tenant_id text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    external_id text,
    name text NOT NULL,
    location_id text,
    primary_contact_id text,
    CONSTRAINT organization_name_check CHECK ((name <> ''::text))
);


ALTER TABLE app.organization OWNER TO postgres;

--
-- Name: permission; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.permission (
    key text NOT NULL
);


ALTER TABLE app.permission OWNER TO postgres;

--
-- Name: registration; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.registration (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    created_at timestamp with time zone DEFAULT (CURRENT_TIMESTAMP)::date NOT NULL,
    expires_at timestamp with time zone DEFAULT ((CURRENT_TIMESTAMP)::date + '24:00:00'::interval) NOT NULL,
    registered_at timestamp with time zone,
    canceled_at timestamp with time zone,
    company_name text NOT NULL,
    email text NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    phone text NOT NULL,
    registration_info jsonb DEFAULT '{}'::jsonb NOT NULL
);


ALTER TABLE app.registration OWNER TO postgres;

--
-- Name: signed_eula; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.signed_eula (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    eula_id text NOT NULL,
    app_tenant_id text NOT NULL,
    signed_by_app_user_id text NOT NULL,
    signed_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    content text NOT NULL
);


ALTER TABLE app.signed_eula OWNER TO postgres;

--
-- Name: sub_module; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.sub_module (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    module_id text NOT NULL,
    name text NOT NULL,
    route_name text NOT NULL,
    permission_key text NOT NULL
);


ALTER TABLE app.sub_module OWNER TO postgres;

--
-- Name: supported_language; Type: TABLE; Schema: app; Owner: postgres
--

CREATE TABLE app.supported_language (
    id text NOT NULL,
    name text NOT NULL,
    inactive boolean DEFAULT true NOT NULL
);


ALTER TABLE app.supported_language OWNER TO postgres;

--
-- Name: email_request; Type: TABLE; Schema: msg; Owner: postgres
--

CREATE TABLE msg.email_request (
    id text DEFAULT shard_1.id_generator() NOT NULL,
    app_tenant_id text NOT NULL,
    sent_by_app_user_id text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    status msg.email_status DEFAULT 'sent'::msg.email_status NOT NULL,
    subject text,
    content text,
    from_address text,
    to_addresses text,
    cc_addresses text,
    bcc_addresses text,
    options text,
    ext_mail_service_result jsonb DEFAULT '{}'::jsonb NOT NULL
);


ALTER TABLE msg.email_request OWNER TO postgres;

--
-- Name: global_id_sequence; Type: SEQUENCE; Schema: shard_1; Owner: postgres
--

CREATE SEQUENCE shard_1.global_id_sequence
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE shard_1.global_id_sequence OWNER TO postgres;

--
-- Data for Name: app_exception; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.app_exception (err_code, description, svr_message_mask, ui_message_mask) FROM stdin;
\.


--
-- Data for Name: app_route; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.app_route (id, name, application_key, permission_key, description, path, menu_behavior, menu_parent_name, menu_ordinal) FROM stdin;
\.


--
-- Data for Name: app_tenant; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.app_tenant (id, created_at, name, identifier, organization_id, registration_identifier, registration_complete, settings, type, parent_app_tenant_id, anchor_subscription_id, billing_topic_id) FROM stdin;
2978135433587721825	2022-11-23 22:10:30.932856+00	Anchor Tenant	anchor	2978135433612887651	2978135433587721826	f	{}	anchor	\N	2978135433696773732	\N
2978135444492912241	2022-11-23 22:10:32.171285+00	Drainage Tenant	dng	2978135444509689459	2978135444492912242	f	{}	customer	2978135433587721825	2978135444568409716	\N
2978135452478867072	2022-11-23 22:10:33.18673+00	Address Book Tenant	address-book	2978135452478867074	2978135452478867073	f	{}	customer	2978135433587721825	2978135452478867075	\N
\.


--
-- Data for Name: app_tenant_group; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.app_tenant_group (id, name, support_email) FROM stdin;
\.


--
-- Data for Name: app_tenant_group_admin; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.app_tenant_group_admin (id, app_tenant_group_id, app_user_id) FROM stdin;
\.


--
-- Data for Name: app_tenant_group_member; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.app_tenant_group_member (id, app_tenant_group_id, app_tenant_id) FROM stdin;
\.


--
-- Data for Name: app_tenant_subscription; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.app_tenant_subscription (id, created_date, expiration_date, renewal_behavior, is_anchor_subscription, inactive, app_tenant_id, license_pack_id, payment_processor_info, parent_app_tenant_subscription_id) FROM stdin;
2978135433696773732	2022-11-23	\N	ask_admin	t	f	2978135433587721825	2978135432421705302	{}	\N
2978135444568409716	2022-11-23	\N	ask_admin	t	f	2978135444492912241	2978135433076016734	{}	\N
2978135452478867075	2022-11-23	\N	ask_admin	t	f	2978135452478867072	2978135432899855963	{}	\N
\.


--
-- Data for Name: app_user; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.app_user (id, app_tenant_id, ext_auth_id, ext_crm_id, contact_id, created_at, username, recovery_email, inactive, password_reset_required, permission_key, is_support, preferred_timezone, settings, ext_auth_blocked, language_id) FROM stdin;
2978135433931654758	2978135433587721825	\N	\N	2978135433906488933	2022-11-23 22:10:30.932856+00	appsuperadmin	app-super-admin@example.com	f	f	SuperAdmin	f	PST8PDT	{}	f	en
2978135434283976298	2978135433587721825	\N	\N	2978135434283976297	2022-11-23 22:10:30.932856+00	fnb-support	help@example.com	f	f	Support	f	PST8PDT	{}	f	en
2978135434309142126	2978135433587721825	\N	\N	2978135434309142125	2022-11-23 22:10:30.932856+00	fnb-demo	demo@example.com	f	f	Demo	f	PST8PDT	{}	f	en
2978135448360060534	2978135444492912241	\N	\N	2978135448343283317	2022-11-23 22:10:32.636648+00	dng-admin	dng-admin@example.com	f	f	User	f	PST8PDT	{}	f	en
2978135452252374652	2978135444492912241	\N	\N	2978135452235597435	2022-11-23 22:10:33.105986+00	dng-user	dng-user@example.com	f	f	User	f	PST8PDT	{}	f	en
2978135456153077381	2978135452478867072	\N	\N	2978135456136300164	2022-11-23 22:10:33.571273+00	address-book-admin	address-book-admin@example.com	f	f	User	f	PST8PDT	{}	f	en
2978135459726624395	2978135452478867072	\N	\N	2978135459709847178	2022-11-23 22:10:34.001235+00	address-book-user	address-book-user@example.com	f	f	User	f	PST8PDT	{}	f	en
\.


--
-- Data for Name: application; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.application (key, name, setting_configs) FROM stdin;
anchor	Anchor	\N
address-book	Address Book	\N
dng	Drainage	\N
\.


--
-- Data for Name: contact; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.contact (id, app_tenant_id, created_at, status, type, organization_id, location_id, external_id, first_name, last_name, email, cell_phone, office_phone, title, nickname) FROM stdin;
2978135433906488933	2978135433587721825	2022-11-23 22:10:30.932856+00	active	individual	\N	\N	appsuperadmin	Appeaus	Adminus	app-super-admin@example.com	\N	\N	\N	\N
2978135434283976297	2978135433587721825	2022-11-23 22:10:30.932856+00	active	individual	\N	\N	fnb-support	FNB	Support	help@example.com	\N	\N	\N	\N
2978135434309142125	2978135433587721825	2022-11-23 22:10:30.932856+00	active	individual	\N	\N	fnb-demo	FNB	Demo	demo@example.com	\N	\N	\N	\N
2978135448343283317	2978135444492912241	2022-11-23 22:10:32.636648+00	active	individual	\N	\N	\N	Drainage	Admin	dng-admin@example.com	\N	\N	\N	\N
2978135452235597435	2978135444492912241	2022-11-23 22:10:33.105986+00	active	individual	\N	\N	\N	Drainage	User	dng-user@example.com	\N	\N	\N	\N
2978135456136300164	2978135452478867072	2022-11-23 22:10:33.571273+00	active	individual	\N	\N	\N	Drainage	Admin	address-book-admin@example.com	\N	\N	\N	\N
2978135459709847178	2978135452478867072	2022-11-23 22:10:34.001235+00	active	individual	\N	\N	\N	Drainage	User	address-book-user@example.com	\N	\N	\N	\N
\.


--
-- Data for Name: error_report; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.error_report (id, first_reported_at, last_reported_at, observed_count, message, comment, reported_by_app_user_id, reported_as_app_user_id, operation_name, variables, status) FROM stdin;
\.


--
-- Data for Name: eula; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.eula (id, created_at, updated_at, is_inactive, deactivated_at, content) FROM stdin;
\.


--
-- Data for Name: facility; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.facility (id, app_tenant_id, created_at, organization_id, location_id, name, external_id) FROM stdin;
\.


--
-- Data for Name: license; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.license (id, app_tenant_id, subscription_id, created_at, external_id, name, license_type_key, assigned_to_app_user_id, inactive, expiration_date, status, status_reason, comment) FROM stdin;
2978135433981986407	2978135433587721825	2978135433696773732	2022-11-23 22:10:30.932856+00	\N	Anchor Super Admin	anchor-super-admin	2978135433931654758	f	\N	active	initial	\N
2978135434283976299	2978135433587721825	2978135433696773732	2022-11-23 22:10:30.932856+00	\N	Anchor Support	anchor-support	2978135434283976298	f	\N	active	initial	\N
2978135434309142127	2978135433587721825	2978135433696773732	2022-11-23 22:10:30.932856+00	\N	Anchor Demo	anchor-demo	2978135434309142126	f	\N	active	initial	\N
2978135448393614967	2978135444492912241	2978135444568409716	2022-11-23 22:10:32.636648+00	\N	Address Book Admin	dng-admin	2978135448360060534	f	\N	active	initial	\N
2978135452285929085	2978135444492912241	2978135444568409716	2022-11-23 22:10:33.105986+00	\N	Address Book User	dng-user	2978135452252374652	f	2023-11-23	active	initial	\N
2978135456178243206	2978135452478867072	2978135452478867075	2022-11-23 22:10:33.571273+00	\N	Address Book Admin	address-book-admin	2978135456153077381	f	\N	active	initial	\N
2978135459760178828	2978135452478867072	2978135452478867075	2022-11-23 22:10:34.001235+00	\N	Address Book User	address-book-user	2978135459726624395	f	2023-11-23	active	initial	\N
\.


--
-- Data for Name: license_pack; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.license_pack (id, key, name, availability, created_at, published_at, discontinued_at, type, renewal_frequency, expiration_interval, expiration_interval_multiplier, explicit_expiration_date, price, upgrade_config, available_add_on_keys, coupon_code, is_public_offering, application_settings, implicit_add_on_keys) FROM stdin;
2978135432421705302	anchor	Anchor	published	2022-11-23 22:10:30.770432+00	2022-11-23 22:10:30.845967+00	\N	anchor	never	none	1	\N	0.00	({})	{}	\N	f	{}	{}
2978135432899855963	address-book	Address Book	published	2022-11-23 22:10:30.857597+00	2022-11-23 22:10:30.869922+00	\N	anchor	never	none	1	\N	0.00	({})	{}	\N	f	{}	{}
2978135433076016734	dng	Drainage	published	2022-11-23 22:10:30.87881+00	2022-11-23 22:10:30.891708+00	\N	anchor	never	none	1	\N	0.00	({})	{}	\N	f	{}	{}
\.


--
-- Data for Name: license_pack_license_type; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.license_pack_license_type (id, license_type_key, license_pack_id, license_count, assign_upon_subscription, unlimited_provision, expiration_interval, expiration_interval_multiplier, explicit_expiration_date) FROM stdin;
2978135432488814167	anchor-super-admin	2978135432421705302	0	f	t	none	1	\N
2978135432597866072	anchor-user	2978135432421705302	0	f	t	none	1	\N
2978135432597866073	anchor-support	2978135432421705302	0	f	t	none	1	\N
2978135432597866074	anchor-demo	2978135432421705302	0	f	t	none	1	\N
2978135432899855964	address-book-admin	2978135432899855963	0	t	t	none	1	\N
2978135432908244573	address-book-user	2978135432899855963	0	f	t	year	1	\N
2978135433076016735	dng-admin	2978135433076016734	0	t	t	none	1	\N
2978135433076016736	dng-user	2978135433076016734	0	f	t	year	1	\N
\.


--
-- Data for Name: license_permission; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.license_permission (id, app_tenant_id, created_at, license_id, permission_key) FROM stdin;
2978135434099426920	2978135433587721825	2022-11-23 22:10:30.932856+00	2978135433981986407	p:super-admin
2978135434292364908	2978135433587721825	2022-11-23 22:10:30.932856+00	2978135434283976299	p:support
2978135434317530736	2978135433587721825	2022-11-23 22:10:30.932856+00	2978135434309142127	p:demo
2978135448410392184	2978135444492912241	2022-11-23 22:10:32.636648+00	2978135448393614967	p:dng-admin
2978135448418780793	2978135444492912241	2022-11-23 22:10:32.636648+00	2978135448393614967	m:dng-admin
2978135448418780794	2978135444492912241	2022-11-23 22:10:32.636648+00	2978135448393614967	m:dng
2978135452302706302	2978135444492912241	2022-11-23 22:10:33.105986+00	2978135452285929085	p:dng-user
2978135452302706303	2978135444492912241	2022-11-23 22:10:33.105986+00	2978135452285929085	m:dng
2978135456195020423	2978135452478867072	2022-11-23 22:10:33.571273+00	2978135456178243206	p:address-book-admin
2978135456195020424	2978135452478867072	2022-11-23 22:10:33.571273+00	2978135456178243206	m:address-book-admin
2978135456195020425	2978135452478867072	2022-11-23 22:10:33.571273+00	2978135456178243206	m:address-book
2978135459776956045	2978135452478867072	2022-11-23 22:10:34.001235+00	2978135459760178828	p:address-book-user
2978135459776956046	2978135452478867072	2022-11-23 22:10:34.001235+00	2978135459760178828	m:address-book
\.


--
-- Data for Name: license_type; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.license_type (key, created_at, external_id, name, application_key, permission_key, sync_user_on_assignment) FROM stdin;
anchor-super-admin	2022-11-23 22:10:29.507317+00	\N	Anchor Super Admin	anchor	SuperAdmin	t
anchor-user	2022-11-23 22:10:29.507317+00	\N	Anchor User	anchor	User	t
anchor-support	2022-11-23 22:10:29.507317+00	\N	Anchor Support	anchor	User	t
anchor-demo	2022-11-23 22:10:29.507317+00	\N	Anchor Demo	anchor	User	t
address-book-admin	2022-11-23 22:10:29.584666+00	\N	Address Book Admin	address-book	User	t
address-book-user	2022-11-23 22:10:29.584666+00	\N	Address Book User	address-book	User	t
dng-admin	2022-11-23 22:10:29.593846+00	\N	Address Book Admin	dng	User	t
dng-user	2022-11-23 22:10:29.593846+00	\N	Address Book User	dng	User	t
\.


--
-- Data for Name: license_type_permission; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.license_type_permission (id, created_at, license_type_key, permission_key) FROM stdin;
2978135422044997192	2022-11-23 22:10:29.507317+00	anchor-super-admin	p:super-admin
2978135422103717449	2022-11-23 22:10:29.507317+00	anchor-user	p:anchor-user
2978135422112106058	2022-11-23 22:10:29.507317+00	anchor-support	p:support
2978135422112106059	2022-11-23 22:10:29.507317+00	anchor-demo	p:demo
2978135422212769356	2022-11-23 22:10:29.584666+00	address-book-admin	p:address-book-admin
2978135422212769357	2022-11-23 22:10:29.584666+00	address-book-admin	m:address-book-admin
2978135422212769358	2022-11-23 22:10:29.584666+00	address-book-admin	m:address-book
2978135422212769359	2022-11-23 22:10:29.584666+00	address-book-user	p:address-book-user
2978135422212769360	2022-11-23 22:10:29.584666+00	address-book-user	m:address-book
2978135422288266833	2022-11-23 22:10:29.593846+00	dng-admin	p:dng-admin
2978135422288266834	2022-11-23 22:10:29.593846+00	dng-admin	m:dng-admin
2978135422288266835	2022-11-23 22:10:29.593846+00	dng-admin	m:dng
2978135422288266836	2022-11-23 22:10:29.593846+00	dng-user	p:dng-user
2978135422288266837	2022-11-23 22:10:29.593846+00	dng-user	m:dng
\.


--
-- Data for Name: location; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.location (id, app_tenant_id, created_at, external_id, name, address1, address2, city, state, zip, lat, lon) FROM stdin;
\.


--
-- Data for Name: module; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.module (id, application_key, name, route_name, permission_key) FROM stdin;
\.


--
-- Data for Name: note; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.note (id, app_tenant_id, created_at, updated_at, created_by_app_user_id, content) FROM stdin;
\.


--
-- Data for Name: organization; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.organization (id, app_tenant_id, created_at, external_id, name, location_id, primary_contact_id) FROM stdin;
2978135433612887651	2978135433587721825	2022-11-23 22:10:30.932856+00	anchor	Anchor Tenant	\N	\N
2978135444509689459	2978135444492912241	2022-11-23 22:10:32.171285+00	dng	Drainage Tenant	\N	\N
2978135452478867074	2978135452478867072	2022-11-23 22:10:33.18673+00	address-book	Address Book Tenant	\N	\N
\.


--
-- Data for Name: permission; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.permission (key) FROM stdin;
p:super-admin
p:anchor-user
p:support
p:demo
p:address-book-admin
m:address-book-admin
p:address-book-user
m:address-book
p:dng-admin
m:dng-admin
p:dng-user
m:dng
\.


--
-- Data for Name: registration; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.registration (id, created_at, expires_at, registered_at, canceled_at, company_name, email, first_name, last_name, phone, registration_info) FROM stdin;
\.


--
-- Data for Name: signed_eula; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.signed_eula (id, eula_id, app_tenant_id, signed_by_app_user_id, signed_at, content) FROM stdin;
\.


--
-- Data for Name: sub_module; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.sub_module (id, module_id, name, route_name, permission_key) FROM stdin;
\.


--
-- Data for Name: supported_language; Type: TABLE DATA; Schema: app; Owner: postgres
--

COPY app.supported_language (id, name, inactive) FROM stdin;
en	English	f
\.


--
-- Data for Name: email_request; Type: TABLE DATA; Schema: msg; Owner: postgres
--

COPY msg.email_request (id, app_tenant_id, sent_by_app_user_id, created_at, status, subject, content, from_address, to_addresses, cc_addresses, bcc_addresses, options, ext_mail_service_result) FROM stdin;
\.


--
-- Data for Name: message; Type: TABLE DATA; Schema: msg; Owner: postgres
--

COPY msg.message (id, app_tenant_id, created_at, status, topic_id, content, posted_by_contact_id, tags, acknowledged_at) FROM stdin;
\.


--
-- Data for Name: subscription; Type: TABLE DATA; Schema: msg; Owner: postgres
--

COPY msg.subscription (id, app_tenant_id, created_at, status, topic_id, subscriber_contact_id, last_read) FROM stdin;
\.


--
-- Data for Name: topic; Type: TABLE DATA; Schema: msg; Owner: postgres
--

COPY msg.topic (id, app_tenant_id, created_at, name, identifier) FROM stdin;
\.


--
-- Name: global_id_sequence; Type: SEQUENCE SET; Schema: shard_1; Owner: postgres
--

SELECT pg_catalog.setval('shard_1.global_id_sequence', 2748, true);


--
-- Name: app_route app_route_name_key; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_route
    ADD CONSTRAINT app_route_name_key UNIQUE (name);


--
-- Name: app_tenant app_tenant_identifier_key; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant
    ADD CONSTRAINT app_tenant_identifier_key UNIQUE (identifier);


--
-- Name: app_user app_user_username_key; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_user
    ADD CONSTRAINT app_user_username_key UNIQUE (username);


--
-- Name: app_exception pk_app_exception; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_exception
    ADD CONSTRAINT pk_app_exception PRIMARY KEY (err_code);


--
-- Name: app_route pk_app_route; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_route
    ADD CONSTRAINT pk_app_route PRIMARY KEY (id);


--
-- Name: app_tenant pk_app_tenant; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant
    ADD CONSTRAINT pk_app_tenant PRIMARY KEY (id);


--
-- Name: app_tenant_group pk_app_tenant_group; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant_group
    ADD CONSTRAINT pk_app_tenant_group PRIMARY KEY (id);


--
-- Name: app_tenant_group_admin pk_app_tenant_group_admin; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant_group_admin
    ADD CONSTRAINT pk_app_tenant_group_admin PRIMARY KEY (id);


--
-- Name: app_tenant_group_member pk_app_tenant_group_member; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant_group_member
    ADD CONSTRAINT pk_app_tenant_group_member PRIMARY KEY (id);


--
-- Name: app_tenant_subscription pk_app_tenant_subscription; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant_subscription
    ADD CONSTRAINT pk_app_tenant_subscription PRIMARY KEY (id);


--
-- Name: app_user pk_app_user; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_user
    ADD CONSTRAINT pk_app_user PRIMARY KEY (id);


--
-- Name: application pk_application; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.application
    ADD CONSTRAINT pk_application PRIMARY KEY (key);


--
-- Name: contact pk_contact; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.contact
    ADD CONSTRAINT pk_contact PRIMARY KEY (id);


--
-- Name: error_report pk_error_report; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.error_report
    ADD CONSTRAINT pk_error_report PRIMARY KEY (id);


--
-- Name: eula pk_eula; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.eula
    ADD CONSTRAINT pk_eula PRIMARY KEY (id);


--
-- Name: facility pk_facility; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.facility
    ADD CONSTRAINT pk_facility PRIMARY KEY (id);


--
-- Name: license pk_license; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license
    ADD CONSTRAINT pk_license PRIMARY KEY (id);


--
-- Name: license_pack pk_license_pack; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license_pack
    ADD CONSTRAINT pk_license_pack PRIMARY KEY (id);


--
-- Name: license_pack_license_type pk_license_pack_license_type; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license_pack_license_type
    ADD CONSTRAINT pk_license_pack_license_type PRIMARY KEY (id);


--
-- Name: license_permission pk_license_permission; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license_permission
    ADD CONSTRAINT pk_license_permission PRIMARY KEY (id);


--
-- Name: license_type pk_license_type; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license_type
    ADD CONSTRAINT pk_license_type PRIMARY KEY (key);


--
-- Name: license_type_permission pk_license_type_permission; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license_type_permission
    ADD CONSTRAINT pk_license_type_permission PRIMARY KEY (id);


--
-- Name: location pk_location; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.location
    ADD CONSTRAINT pk_location PRIMARY KEY (id);


--
-- Name: module pk_module; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.module
    ADD CONSTRAINT pk_module PRIMARY KEY (id);


--
-- Name: note pk_note; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.note
    ADD CONSTRAINT pk_note PRIMARY KEY (id);


--
-- Name: organization pk_organization; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.organization
    ADD CONSTRAINT pk_organization PRIMARY KEY (id);


--
-- Name: permission pk_permission; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.permission
    ADD CONSTRAINT pk_permission PRIMARY KEY (key);


--
-- Name: registration pk_registration; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.registration
    ADD CONSTRAINT pk_registration PRIMARY KEY (id);


--
-- Name: signed_eula pk_signed_eula; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.signed_eula
    ADD CONSTRAINT pk_signed_eula PRIMARY KEY (id);


--
-- Name: sub_module pk_sub_module; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.sub_module
    ADD CONSTRAINT pk_sub_module PRIMARY KEY (id);


--
-- Name: supported_language pk_supported_language; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.supported_language
    ADD CONSTRAINT pk_supported_language PRIMARY KEY (id);


--
-- Name: signed_eula signed_eula_signed_by_app_user_id_eula_id_key; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.signed_eula
    ADD CONSTRAINT signed_eula_signed_by_app_user_id_eula_id_key UNIQUE (signed_by_app_user_id, eula_id);


--
-- Name: app_route uq_app_route; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_route
    ADD CONSTRAINT uq_app_route UNIQUE (application_key, name);


--
-- Name: app_tenant_group uq_app_tenant_group; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant_group
    ADD CONSTRAINT uq_app_tenant_group UNIQUE (name);


--
-- Name: app_tenant_group_admin uq_app_tenant_group_admin; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant_group_admin
    ADD CONSTRAINT uq_app_tenant_group_admin UNIQUE (app_tenant_group_id, app_user_id);


--
-- Name: app_tenant_group_member uq_app_tenant_group_member; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant_group_member
    ADD CONSTRAINT uq_app_tenant_group_member UNIQUE (app_tenant_group_id, app_tenant_id);


--
-- Name: organization uq_app_tenant_name; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.organization
    ADD CONSTRAINT uq_app_tenant_name UNIQUE (app_tenant_id, name);


--
-- Name: app_tenant uq_app_tenant_organization; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant
    ADD CONSTRAINT uq_app_tenant_organization UNIQUE (organization_id);


--
-- Name: app_user uq_app_user_contact; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_user
    ADD CONSTRAINT uq_app_user_contact UNIQUE (contact_id);


--
-- Name: app_user uq_app_user_tenant_recovery_email; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_user
    ADD CONSTRAINT uq_app_user_tenant_recovery_email UNIQUE (app_tenant_id, recovery_email);


--
-- Name: contact uq_contact_app_tenant_and_email; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.contact
    ADD CONSTRAINT uq_contact_app_tenant_and_email UNIQUE (app_tenant_id, email);


--
-- Name: contact uq_contact_app_tenant_and_external_id; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.contact
    ADD CONSTRAINT uq_contact_app_tenant_and_external_id UNIQUE (app_tenant_id, external_id);


--
-- Name: facility uq_facility_app_tenant_and_organization_and_name; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.facility
    ADD CONSTRAINT uq_facility_app_tenant_and_organization_and_name UNIQUE (organization_id, name);


--
-- Name: license_pack_license_type uq_license_pack_license_type; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license_pack_license_type
    ADD CONSTRAINT uq_license_pack_license_type UNIQUE (license_pack_id, license_type_key);


--
-- Name: license_permission uq_license_permission; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license_permission
    ADD CONSTRAINT uq_license_permission UNIQUE (license_id, permission_key);


--
-- Name: license_type uq_license_type; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license_type
    ADD CONSTRAINT uq_license_type UNIQUE (application_key, key);


--
-- Name: license_type_permission uq_license_type_permission; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license_type_permission
    ADD CONSTRAINT uq_license_type_permission UNIQUE (license_type_key, permission_key);


--
-- Name: location uq_location_app_tenant_and_external_id; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.location
    ADD CONSTRAINT uq_location_app_tenant_and_external_id UNIQUE (app_tenant_id, external_id);


--
-- Name: registration uq_registration; Type: CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.registration
    ADD CONSTRAINT uq_registration UNIQUE (company_name);


--
-- Name: email_request pk_email_request; Type: CONSTRAINT; Schema: msg; Owner: postgres
--

ALTER TABLE ONLY msg.email_request
    ADD CONSTRAINT pk_email_request PRIMARY KEY (id);


--
-- Name: message pk_message; Type: CONSTRAINT; Schema: msg; Owner: postgres
--

ALTER TABLE ONLY msg.message
    ADD CONSTRAINT pk_message PRIMARY KEY (id);


--
-- Name: subscription pk_subscription; Type: CONSTRAINT; Schema: msg; Owner: postgres
--

ALTER TABLE ONLY msg.subscription
    ADD CONSTRAINT pk_subscription PRIMARY KEY (id);


--
-- Name: topic pk_topic; Type: CONSTRAINT; Schema: msg; Owner: postgres
--

ALTER TABLE ONLY msg.topic
    ADD CONSTRAINT pk_topic PRIMARY KEY (id);


--
-- Name: subscription uq_subscription; Type: CONSTRAINT; Schema: msg; Owner: postgres
--

ALTER TABLE ONLY msg.subscription
    ADD CONSTRAINT uq_subscription UNIQUE (topic_id, subscriber_contact_id);


--
-- Name: idx_app_app_tenant_subscription_parent_app_tenant_subscription_; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_app_app_tenant_subscription_parent_app_tenant_subscription_ ON app.app_tenant_subscription USING btree (parent_app_tenant_subscription_id);


--
-- Name: idx_app_error_report_reported_as_app_user_id; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_app_error_report_reported_as_app_user_id ON app.error_report USING btree (reported_as_app_user_id);


--
-- Name: idx_app_error_report_reported_by_app_user_id; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_app_error_report_reported_by_app_user_id ON app.error_report USING btree (reported_by_app_user_id);


--
-- Name: idx_app_license_assigned_to_active; Type: INDEX; Schema: app; Owner: postgres
--

CREATE UNIQUE INDEX idx_app_license_assigned_to_active ON app.license USING btree (assigned_to_app_user_id, license_type_key, inactive) WHERE (inactive = false);


--
-- Name: idx_app_license_pack_draft_key; Type: INDEX; Schema: app; Owner: postgres
--

CREATE UNIQUE INDEX idx_app_license_pack_draft_key ON app.license_pack USING btree (key, availability) WHERE (availability = 'draft'::app.license_pack_availability);


--
-- Name: idx_app_license_pack_published_key; Type: INDEX; Schema: app; Owner: postgres
--

CREATE UNIQUE INDEX idx_app_license_pack_published_key ON app.license_pack USING btree (key, availability) WHERE (availability = 'published'::app.license_pack_availability);


--
-- Name: idx_app_route_application_key; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_app_route_application_key ON app.app_route USING btree (application_key);


--
-- Name: idx_app_route_menu_parent; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_app_route_menu_parent ON app.app_route USING btree (menu_parent_name);


--
-- Name: idx_app_signed_eula_app_tenant_id; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_app_signed_eula_app_tenant_id ON app.signed_eula USING btree (app_tenant_id);


--
-- Name: idx_app_signed_eula_eula_id; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_app_signed_eula_eula_id ON app.signed_eula USING btree (eula_id);


--
-- Name: idx_app_tenant_anchor_subscription; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_app_tenant_anchor_subscription ON app.app_tenant USING btree (anchor_subscription_id);


--
-- Name: idx_app_tenant_billing_topic; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_app_tenant_billing_topic ON app.app_tenant USING btree (billing_topic_id);


--
-- Name: idx_app_tenant_parent; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_app_tenant_parent ON app.app_tenant USING btree (parent_app_tenant_id);


--
-- Name: idx_app_tenant_subscription_anchor; Type: INDEX; Schema: app; Owner: postgres
--

CREATE UNIQUE INDEX idx_app_tenant_subscription_anchor ON app.app_tenant_subscription USING btree (app_tenant_id, is_anchor_subscription, inactive) WHERE ((is_anchor_subscription = true) AND (inactive = false));


--
-- Name: idx_app_tenant_subscription_subcription; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_app_tenant_subscription_subcription ON app.app_tenant_subscription USING btree (license_pack_id);


--
-- Name: idx_app_tenant_subscription_tenant; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_app_tenant_subscription_tenant ON app.app_tenant_subscription USING btree (app_tenant_id);


--
-- Name: idx_app_user_app_tenant; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_app_user_app_tenant ON app.app_user USING btree (app_tenant_id);


--
-- Name: idx_contact_app_tenant; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_contact_app_tenant ON app.contact USING btree (app_tenant_id);


--
-- Name: idx_contact_location; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_contact_location ON app.contact USING btree (location_id);


--
-- Name: idx_contact_organization; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_contact_organization ON app.contact USING btree (organization_id);


--
-- Name: idx_facility_app_tenant; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_facility_app_tenant ON app.facility USING btree (app_tenant_id);


--
-- Name: idx_facility_location; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_facility_location ON app.facility USING btree (location_id);


--
-- Name: idx_facility_organization; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_facility_organization ON app.facility USING btree (organization_id);


--
-- Name: idx_license_app_tenant; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_license_app_tenant ON app.license USING btree (app_tenant_id);


--
-- Name: idx_license_app_user; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_license_app_user ON app.license USING btree (assigned_to_app_user_id);


--
-- Name: idx_license_license_type; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_license_license_type ON app.license USING btree (license_type_key);


--
-- Name: idx_license_permission_app_tenant; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_license_permission_app_tenant ON app.license_permission USING btree (app_tenant_id);


--
-- Name: idx_license_permission_permission_key; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_license_permission_permission_key ON app.license_permission USING btree (permission_key);


--
-- Name: idx_license_subscription; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_license_subscription ON app.license USING btree (subscription_id);


--
-- Name: idx_license_type_permission_permission_key; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_license_type_permission_permission_key ON app.license_type_permission USING btree (permission_key);


--
-- Name: idx_location_app_tenant; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_location_app_tenant ON app.location USING btree (app_tenant_id);


--
-- Name: idx_lpl_license_pack; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_lpl_license_pack ON app.license_pack_license_type USING btree (license_pack_id);


--
-- Name: idx_lpl_license_type; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_lpl_license_type ON app.license_pack_license_type USING btree (license_type_key);


--
-- Name: idx_note_app_tenant_id; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_note_app_tenant_id ON app.note USING btree (app_tenant_id);


--
-- Name: idx_note_created_by_app_user_id; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_note_created_by_app_user_id ON app.note USING btree (created_by_app_user_id);


--
-- Name: idx_organization_app_tenant; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_organization_app_tenant ON app.organization USING btree (app_tenant_id);


--
-- Name: idx_organization_location; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_organization_location ON app.organization USING btree (location_id);


--
-- Name: idx_organization_primary_contact; Type: INDEX; Schema: app; Owner: postgres
--

CREATE INDEX idx_organization_primary_contact ON app.organization USING btree (primary_contact_id);


--
-- Name: idx_uq_eula_active; Type: INDEX; Schema: app; Owner: postgres
--

CREATE UNIQUE INDEX idx_uq_eula_active ON app.eula USING btree (id) WHERE (is_inactive = false);


--
-- Name: idx_message_app_tenant_id; Type: INDEX; Schema: msg; Owner: postgres
--

CREATE INDEX idx_message_app_tenant_id ON msg.message USING btree (app_tenant_id);


--
-- Name: idx_message_posted_by_contact_id; Type: INDEX; Schema: msg; Owner: postgres
--

CREATE INDEX idx_message_posted_by_contact_id ON msg.message USING btree (posted_by_contact_id);


--
-- Name: idx_msg_email_request_app_tenant_id; Type: INDEX; Schema: msg; Owner: postgres
--

CREATE INDEX idx_msg_email_request_app_tenant_id ON msg.email_request USING btree (app_tenant_id);


--
-- Name: idx_msg_email_request_sent_by_app_user_id; Type: INDEX; Schema: msg; Owner: postgres
--

CREATE INDEX idx_msg_email_request_sent_by_app_user_id ON msg.email_request USING btree (sent_by_app_user_id);


--
-- Name: idx_msg_message_topic; Type: INDEX; Schema: msg; Owner: postgres
--

CREATE INDEX idx_msg_message_topic ON msg.message USING btree (topic_id);


--
-- Name: idx_subscription_app_tenant_id; Type: INDEX; Schema: msg; Owner: postgres
--

CREATE INDEX idx_subscription_app_tenant_id ON msg.subscription USING btree (app_tenant_id);


--
-- Name: idx_subscription_subscriber_contact_id; Type: INDEX; Schema: msg; Owner: postgres
--

CREATE INDEX idx_subscription_subscriber_contact_id ON msg.subscription USING btree (subscriber_contact_id);


--
-- Name: idx_topic_app_tenant; Type: INDEX; Schema: msg; Owner: postgres
--

CREATE INDEX idx_topic_app_tenant ON msg.topic USING btree (app_tenant_id);


--
-- Name: idx_topic_app_tenant_identifier; Type: INDEX; Schema: msg; Owner: postgres
--

CREATE UNIQUE INDEX idx_topic_app_tenant_identifier ON msg.topic USING btree (app_tenant_id, identifier);


--
-- Name: application application_no_delete; Type: TRIGGER; Schema: app; Owner: postgres
--

CREATE TRIGGER application_no_delete BEFORE DELETE ON app.application FOR EACH ROW EXECUTE FUNCTION app.application_no_delete();


--
-- Name: eula tg_before_update_eula; Type: TRIGGER; Schema: app; Owner: postgres
--

CREATE TRIGGER tg_before_update_eula BEFORE UPDATE ON app.eula FOR EACH ROW EXECUTE FUNCTION app.fn_update_eula_trigger();


--
-- Name: license tg_calculate_license_status; Type: TRIGGER; Schema: app; Owner: postgres
--

CREATE TRIGGER tg_calculate_license_status BEFORE INSERT OR UPDATE ON app.license FOR EACH ROW EXECUTE FUNCTION app.fn_ensure_license_status();


--
-- Name: app_route fk_app_route_application; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_route
    ADD CONSTRAINT fk_app_route_application FOREIGN KEY (application_key) REFERENCES app.application(key);


--
-- Name: app_route fk_app_route_menu_parent; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_route
    ADD CONSTRAINT fk_app_route_menu_parent FOREIGN KEY (menu_parent_name) REFERENCES app.app_route(name);


--
-- Name: app_tenant fk_app_tenant_anchor_subscription; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant
    ADD CONSTRAINT fk_app_tenant_anchor_subscription FOREIGN KEY (anchor_subscription_id) REFERENCES app.app_tenant_subscription(id);


--
-- Name: app_tenant fk_app_tenant_billing_topic; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant
    ADD CONSTRAINT fk_app_tenant_billing_topic FOREIGN KEY (billing_topic_id) REFERENCES msg.topic(id);


--
-- Name: app_tenant_group_admin fk_app_tenant_group_admin_app_user; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant_group_admin
    ADD CONSTRAINT fk_app_tenant_group_admin_app_user FOREIGN KEY (app_user_id) REFERENCES app.app_user(id);


--
-- Name: app_tenant_group_admin fk_app_tenant_group_admin_group; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant_group_admin
    ADD CONSTRAINT fk_app_tenant_group_admin_group FOREIGN KEY (app_tenant_group_id) REFERENCES app.app_tenant_group(id);


--
-- Name: app_tenant_group_member fk_app_tenant_group_member_app_tenant; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant_group_member
    ADD CONSTRAINT fk_app_tenant_group_member_app_tenant FOREIGN KEY (app_tenant_id) REFERENCES app.app_tenant(id);


--
-- Name: app_tenant_group_member fk_app_tenant_group_member_group; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant_group_member
    ADD CONSTRAINT fk_app_tenant_group_member_group FOREIGN KEY (app_tenant_group_id) REFERENCES app.app_tenant_group(id);


--
-- Name: app_tenant fk_app_tenant_organization; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant
    ADD CONSTRAINT fk_app_tenant_organization FOREIGN KEY (organization_id) REFERENCES app.organization(id);


--
-- Name: app_tenant fk_app_tenant_parent; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant
    ADD CONSTRAINT fk_app_tenant_parent FOREIGN KEY (parent_app_tenant_id) REFERENCES app.app_tenant(id);


--
-- Name: app_tenant_subscription fk_app_tenant_subscription_license_pack; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant_subscription
    ADD CONSTRAINT fk_app_tenant_subscription_license_pack FOREIGN KEY (license_pack_id) REFERENCES app.license_pack(id);


--
-- Name: app_tenant_subscription fk_app_tenant_subscription_parent; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant_subscription
    ADD CONSTRAINT fk_app_tenant_subscription_parent FOREIGN KEY (parent_app_tenant_subscription_id) REFERENCES app.app_tenant_subscription(id);


--
-- Name: app_tenant_subscription fk_app_tenant_subscription_tenant; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_tenant_subscription
    ADD CONSTRAINT fk_app_tenant_subscription_tenant FOREIGN KEY (app_tenant_id) REFERENCES app.app_tenant(id);


--
-- Name: app_user fk_app_user_app_tenant; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_user
    ADD CONSTRAINT fk_app_user_app_tenant FOREIGN KEY (app_tenant_id) REFERENCES app.app_tenant(id);


--
-- Name: app_user fk_app_user_contact; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_user
    ADD CONSTRAINT fk_app_user_contact FOREIGN KEY (contact_id) REFERENCES app.contact(id);


--
-- Name: app_user fk_app_user_language; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.app_user
    ADD CONSTRAINT fk_app_user_language FOREIGN KEY (language_id) REFERENCES app.supported_language(id);


--
-- Name: contact fk_contact_app_tenant; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.contact
    ADD CONSTRAINT fk_contact_app_tenant FOREIGN KEY (app_tenant_id) REFERENCES app.app_tenant(id);


--
-- Name: contact fk_contact_location; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.contact
    ADD CONSTRAINT fk_contact_location FOREIGN KEY (location_id) REFERENCES app.location(id);


--
-- Name: contact fk_contact_organization; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.contact
    ADD CONSTRAINT fk_contact_organization FOREIGN KEY (organization_id) REFERENCES app.organization(id);


--
-- Name: error_report fk_error_report_reported_as_app_user; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.error_report
    ADD CONSTRAINT fk_error_report_reported_as_app_user FOREIGN KEY (reported_as_app_user_id) REFERENCES app.app_user(id);


--
-- Name: error_report fk_error_report_reported_by_app_user; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.error_report
    ADD CONSTRAINT fk_error_report_reported_by_app_user FOREIGN KEY (reported_by_app_user_id) REFERENCES app.app_user(id);


--
-- Name: facility fk_facility_app_tenant; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.facility
    ADD CONSTRAINT fk_facility_app_tenant FOREIGN KEY (app_tenant_id) REFERENCES app.app_tenant(id);


--
-- Name: facility fk_facility_location; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.facility
    ADD CONSTRAINT fk_facility_location FOREIGN KEY (location_id) REFERENCES app.location(id);


--
-- Name: facility fk_facility_organization; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.facility
    ADD CONSTRAINT fk_facility_organization FOREIGN KEY (organization_id) REFERENCES app.organization(id);


--
-- Name: license fk_license_app_tenant; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license
    ADD CONSTRAINT fk_license_app_tenant FOREIGN KEY (app_tenant_id) REFERENCES app.app_tenant(id);


--
-- Name: license fk_license_assigned_to_app_user; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license
    ADD CONSTRAINT fk_license_assigned_to_app_user FOREIGN KEY (assigned_to_app_user_id) REFERENCES app.app_user(id);


--
-- Name: license fk_license_license_type; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license
    ADD CONSTRAINT fk_license_license_type FOREIGN KEY (license_type_key) REFERENCES app.license_type(key);


--
-- Name: license_permission fk_license_permission_app_tenant; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license_permission
    ADD CONSTRAINT fk_license_permission_app_tenant FOREIGN KEY (app_tenant_id) REFERENCES app.app_tenant(id);


--
-- Name: license_permission fk_license_permission_license; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license_permission
    ADD CONSTRAINT fk_license_permission_license FOREIGN KEY (license_id) REFERENCES app.license(id);


--
-- Name: license_permission fk_license_permission_permission; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license_permission
    ADD CONSTRAINT fk_license_permission_permission FOREIGN KEY (permission_key) REFERENCES app.permission(key);


--
-- Name: license fk_license_subscription; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license
    ADD CONSTRAINT fk_license_subscription FOREIGN KEY (subscription_id) REFERENCES app.app_tenant_subscription(id);


--
-- Name: license_type fk_license_type_application; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license_type
    ADD CONSTRAINT fk_license_type_application FOREIGN KEY (application_key) REFERENCES app.application(key);


--
-- Name: license_type_permission fk_license_type_permission_license_type; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license_type_permission
    ADD CONSTRAINT fk_license_type_permission_license_type FOREIGN KEY (license_type_key) REFERENCES app.license_type(key);


--
-- Name: license_type_permission fk_license_type_permission_permission; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license_type_permission
    ADD CONSTRAINT fk_license_type_permission_permission FOREIGN KEY (permission_key) REFERENCES app.permission(key);


--
-- Name: location fk_location_app_tenant; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.location
    ADD CONSTRAINT fk_location_app_tenant FOREIGN KEY (app_tenant_id) REFERENCES app.app_tenant(id);


--
-- Name: license_pack_license_type fk_lpl_license_pack; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license_pack_license_type
    ADD CONSTRAINT fk_lpl_license_pack FOREIGN KEY (license_pack_id) REFERENCES app.license_pack(id);


--
-- Name: license_pack_license_type fk_lpl_license_type; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.license_pack_license_type
    ADD CONSTRAINT fk_lpl_license_type FOREIGN KEY (license_type_key) REFERENCES app.license_type(key);


--
-- Name: module fk_module_application; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.module
    ADD CONSTRAINT fk_module_application FOREIGN KEY (application_key) REFERENCES app.application(key);


--
-- Name: module fk_module_permission_key; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.module
    ADD CONSTRAINT fk_module_permission_key FOREIGN KEY (permission_key) REFERENCES app.permission(key);


--
-- Name: note fk_note_app_tenant; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.note
    ADD CONSTRAINT fk_note_app_tenant FOREIGN KEY (app_tenant_id) REFERENCES app.app_tenant(id);


--
-- Name: note fk_note_app_user; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.note
    ADD CONSTRAINT fk_note_app_user FOREIGN KEY (created_by_app_user_id) REFERENCES app.app_user(id);


--
-- Name: organization fk_organization_app_tenant; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.organization
    ADD CONSTRAINT fk_organization_app_tenant FOREIGN KEY (app_tenant_id) REFERENCES app.app_tenant(id);


--
-- Name: organization fk_organization_location; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.organization
    ADD CONSTRAINT fk_organization_location FOREIGN KEY (location_id) REFERENCES app.location(id);


--
-- Name: organization fk_organization_primary_contact; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.organization
    ADD CONSTRAINT fk_organization_primary_contact FOREIGN KEY (primary_contact_id) REFERENCES app.contact(id);


--
-- Name: signed_eula fk_signed_eula_app_tenant; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.signed_eula
    ADD CONSTRAINT fk_signed_eula_app_tenant FOREIGN KEY (app_tenant_id) REFERENCES app.app_tenant(id);


--
-- Name: signed_eula fk_signed_eula_eula; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.signed_eula
    ADD CONSTRAINT fk_signed_eula_eula FOREIGN KEY (eula_id) REFERENCES app.eula(id);


--
-- Name: signed_eula fk_signed_eula_signed_by_app_user; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.signed_eula
    ADD CONSTRAINT fk_signed_eula_signed_by_app_user FOREIGN KEY (signed_by_app_user_id) REFERENCES app.app_user(id);


--
-- Name: sub_module fk_sub_module_module; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.sub_module
    ADD CONSTRAINT fk_sub_module_module FOREIGN KEY (module_id) REFERENCES app.module(id);


--
-- Name: sub_module fk_sub_module_permission_key; Type: FK CONSTRAINT; Schema: app; Owner: postgres
--

ALTER TABLE ONLY app.sub_module
    ADD CONSTRAINT fk_sub_module_permission_key FOREIGN KEY (permission_key) REFERENCES app.permission(key);


--
-- Name: email_request fk_email_request_app_tenant; Type: FK CONSTRAINT; Schema: msg; Owner: postgres
--

ALTER TABLE ONLY msg.email_request
    ADD CONSTRAINT fk_email_request_app_tenant FOREIGN KEY (app_tenant_id) REFERENCES app.app_tenant(id);


--
-- Name: email_request fk_email_request_app_user; Type: FK CONSTRAINT; Schema: msg; Owner: postgres
--

ALTER TABLE ONLY msg.email_request
    ADD CONSTRAINT fk_email_request_app_user FOREIGN KEY (sent_by_app_user_id) REFERENCES app.app_user(id);


--
-- Name: message fk_message_app_tenant; Type: FK CONSTRAINT; Schema: msg; Owner: postgres
--

ALTER TABLE ONLY msg.message
    ADD CONSTRAINT fk_message_app_tenant FOREIGN KEY (app_tenant_id) REFERENCES app.app_tenant(id);


--
-- Name: message fk_message_posted_by; Type: FK CONSTRAINT; Schema: msg; Owner: postgres
--

ALTER TABLE ONLY msg.message
    ADD CONSTRAINT fk_message_posted_by FOREIGN KEY (posted_by_contact_id) REFERENCES app.contact(id);


--
-- Name: message fk_message_topic; Type: FK CONSTRAINT; Schema: msg; Owner: postgres
--

ALTER TABLE ONLY msg.message
    ADD CONSTRAINT fk_message_topic FOREIGN KEY (topic_id) REFERENCES msg.topic(id);


--
-- Name: subscription fk_subscription_app_tenant; Type: FK CONSTRAINT; Schema: msg; Owner: postgres
--

ALTER TABLE ONLY msg.subscription
    ADD CONSTRAINT fk_subscription_app_tenant FOREIGN KEY (app_tenant_id) REFERENCES app.app_tenant(id);


--
-- Name: subscription fk_subscription_subscriber_contact; Type: FK CONSTRAINT; Schema: msg; Owner: postgres
--

ALTER TABLE ONLY msg.subscription
    ADD CONSTRAINT fk_subscription_subscriber_contact FOREIGN KEY (subscriber_contact_id) REFERENCES app.contact(id);


--
-- Name: subscription fk_subscription_topic; Type: FK CONSTRAINT; Schema: msg; Owner: postgres
--

ALTER TABLE ONLY msg.subscription
    ADD CONSTRAINT fk_subscription_topic FOREIGN KEY (topic_id) REFERENCES msg.topic(id);


--
-- Name: app_exception; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.app_exception ENABLE ROW LEVEL SECURITY;

--
-- Name: app_route; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.app_route ENABLE ROW LEVEL SECURITY;

--
-- Name: app_tenant; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.app_tenant ENABLE ROW LEVEL SECURITY;

--
-- Name: app_tenant_group; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.app_tenant_group ENABLE ROW LEVEL SECURITY;

--
-- Name: app_tenant_group_admin; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.app_tenant_group_admin ENABLE ROW LEVEL SECURITY;

--
-- Name: app_tenant_group_member; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.app_tenant_group_member ENABLE ROW LEVEL SECURITY;

--
-- Name: app_tenant_subscription; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.app_tenant_subscription ENABLE ROW LEVEL SECURITY;

--
-- Name: app_user; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.app_user ENABLE ROW LEVEL SECURITY;

--
-- Name: application; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.application ENABLE ROW LEVEL SECURITY;

--
-- Name: contact; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.contact ENABLE ROW LEVEL SECURITY;

--
-- Name: error_report; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.error_report ENABLE ROW LEVEL SECURITY;

--
-- Name: eula; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.eula ENABLE ROW LEVEL SECURITY;

--
-- Name: facility; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.facility ENABLE ROW LEVEL SECURITY;

--
-- Name: license; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.license ENABLE ROW LEVEL SECURITY;

--
-- Name: license_pack; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.license_pack ENABLE ROW LEVEL SECURITY;

--
-- Name: license_pack_license_type; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.license_pack_license_type ENABLE ROW LEVEL SECURITY;

--
-- Name: license_permission; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.license_permission ENABLE ROW LEVEL SECURITY;

--
-- Name: license_type; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.license_type ENABLE ROW LEVEL SECURITY;

--
-- Name: license_type_permission; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.license_type_permission ENABLE ROW LEVEL SECURITY;

--
-- Name: location; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.location ENABLE ROW LEVEL SECURITY;

--
-- Name: note; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.note ENABLE ROW LEVEL SECURITY;

--
-- Name: organization; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.organization ENABLE ROW LEVEL SECURITY;

--
-- Name: permission; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.permission ENABLE ROW LEVEL SECURITY;

--
-- Name: registration; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.registration ENABLE ROW LEVEL SECURITY;

--
-- Name: signed_eula; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.signed_eula ENABLE ROW LEVEL SECURITY;

--
-- Name: supported_language; Type: ROW SECURITY; Schema: app; Owner: postgres
--

ALTER TABLE app.supported_language ENABLE ROW LEVEL SECURITY;

--
-- Name: app_exception tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.app_exception FOR DELETE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_route tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.app_route FOR DELETE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_tenant tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.app_tenant FOR DELETE TO app_usr USING (auth_fn.app_user_has_access(id));


--
-- Name: app_tenant_group tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.app_tenant_group FOR DELETE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_tenant_group_admin tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.app_tenant_group_admin FOR DELETE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_tenant_group_member tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.app_tenant_group_member FOR DELETE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_tenant_subscription tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.app_tenant_subscription FOR DELETE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: app_user tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.app_user FOR DELETE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: application tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.application FOR DELETE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: contact tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.contact FOR DELETE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: error_report tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.error_report FOR DELETE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: eula tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.eula FOR DELETE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: facility tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.facility FOR DELETE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: license tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.license FOR DELETE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: license_pack tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.license_pack FOR DELETE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: license_pack_license_type tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.license_pack_license_type FOR DELETE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: license_permission tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.license_permission FOR DELETE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: license_type tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.license_type FOR DELETE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: license_type_permission tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.license_type_permission FOR DELETE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: location tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.location FOR DELETE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: note tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.note FOR DELETE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: organization tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.organization FOR DELETE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: permission tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.permission FOR DELETE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: registration tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.registration FOR DELETE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: signed_eula tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.signed_eula FOR DELETE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: supported_language tenant_delete; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_delete ON app.supported_language FOR DELETE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_exception tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.app_exception FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_route tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.app_route FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_tenant tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.app_tenant FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_access(id));


--
-- Name: app_tenant_group tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.app_tenant_group FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_tenant_group_admin tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.app_tenant_group_admin FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_tenant_group_member tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.app_tenant_group_member FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_tenant_subscription tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.app_tenant_subscription FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: app_user tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.app_user FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: application tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.application FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: contact tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.contact FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: error_report tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.error_report FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: eula tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.eula FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: facility tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.facility FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: license tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.license FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: license_pack tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.license_pack FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: license_pack_license_type tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.license_pack_license_type FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: license_permission tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.license_permission FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: license_type tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.license_type FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: license_type_permission tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.license_type_permission FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: location tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.location FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: note tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.note FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: organization tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.organization FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: permission tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.permission FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: registration tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.registration FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: signed_eula tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.signed_eula FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: supported_language tenant_insert; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_insert ON app.supported_language FOR INSERT TO app_usr WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_exception tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.app_exception FOR SELECT TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_route tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.app_route FOR SELECT TO app_usr USING ((1 = 1));


--
-- Name: app_tenant tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.app_tenant FOR SELECT TO app_usr USING (auth_fn.app_user_has_access(id));


--
-- Name: app_tenant_group tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.app_tenant_group FOR SELECT TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_tenant_group_admin tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.app_tenant_group_admin FOR SELECT TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_tenant_group_member tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.app_tenant_group_member FOR SELECT TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_tenant_subscription tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.app_tenant_subscription FOR SELECT TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: app_user tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.app_user FOR SELECT TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: application tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.application FOR SELECT TO app_usr USING ((1 = 1));


--
-- Name: contact tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.contact FOR SELECT TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: error_report tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.error_report FOR SELECT TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: eula tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.eula FOR SELECT TO app_usr USING ((1 = 1));


--
-- Name: facility tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.facility FOR SELECT TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: license tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.license FOR SELECT TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: license_pack tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.license_pack FOR SELECT TO app_usr USING ((1 = 1));


--
-- Name: license_pack_license_type tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.license_pack_license_type FOR SELECT TO app_usr USING ((1 = 1));


--
-- Name: license_permission tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.license_permission FOR SELECT TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: license_type tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.license_type FOR SELECT TO app_usr USING ((1 = 1));


--
-- Name: license_type_permission tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.license_type_permission FOR SELECT TO app_usr USING ((1 = 1));


--
-- Name: location tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.location FOR SELECT TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: note tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.note FOR SELECT TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: organization tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.organization FOR SELECT TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: permission tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.permission FOR SELECT TO app_usr USING ((1 = 1));


--
-- Name: registration tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.registration FOR SELECT TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: signed_eula tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.signed_eula FOR SELECT TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: supported_language tenant_select; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_select ON app.supported_language FOR SELECT TO app_usr USING ((1 = 1));


--
-- Name: app_exception tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.app_exception FOR UPDATE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text)) WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_route tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.app_route FOR UPDATE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text)) WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_tenant tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.app_tenant FOR UPDATE TO app_usr USING (auth_fn.app_user_has_access(id)) WITH CHECK (auth_fn.app_user_has_access(id));


--
-- Name: app_tenant_group tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.app_tenant_group FOR UPDATE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text)) WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_tenant_group_admin tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.app_tenant_group_admin FOR UPDATE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text)) WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_tenant_group_member tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.app_tenant_group_member FOR UPDATE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text)) WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: app_tenant_subscription tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.app_tenant_subscription FOR UPDATE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id)) WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: app_user tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.app_user FOR UPDATE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id)) WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: application tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.application FOR UPDATE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text)) WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: contact tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.contact FOR UPDATE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id)) WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: error_report tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.error_report FOR UPDATE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text)) WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: eula tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.eula FOR UPDATE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text)) WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: facility tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.facility FOR UPDATE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id)) WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: license tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.license FOR UPDATE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id)) WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: license_pack tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.license_pack FOR UPDATE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text)) WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: license_pack_license_type tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.license_pack_license_type FOR UPDATE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text)) WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: license_permission tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.license_permission FOR UPDATE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id)) WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: license_type tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.license_type FOR UPDATE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text)) WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: license_type_permission tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.license_type_permission FOR UPDATE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text)) WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: location tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.location FOR UPDATE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id)) WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: note tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.note FOR UPDATE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id)) WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: organization tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.organization FOR UPDATE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id)) WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: permission tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.permission FOR UPDATE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text)) WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: registration tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.registration FOR UPDATE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text)) WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: signed_eula tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.signed_eula FOR UPDATE TO app_usr USING (auth_fn.app_user_has_access(app_tenant_id)) WITH CHECK (auth_fn.app_user_has_access(app_tenant_id));


--
-- Name: supported_language tenant_update; Type: POLICY; Schema: app; Owner: postgres
--

CREATE POLICY tenant_update ON app.supported_language FOR UPDATE TO app_usr USING (auth_fn.app_user_has_permission('p:super-admin'::text)) WITH CHECK (auth_fn.app_user_has_permission('p:super-admin'::text));


--
-- Name: SCHEMA app; Type: ACL; Schema: -; Owner: postgres
--

GRANT USAGE ON SCHEMA app TO app_usr;


--
-- Name: SCHEMA auth_bootstrap; Type: ACL; Schema: -; Owner: postgres
--

GRANT USAGE ON SCHEMA auth_bootstrap TO app_anon;
GRANT USAGE ON SCHEMA auth_bootstrap TO app_usr;


--
-- Name: SCHEMA auth_fn; Type: ACL; Schema: -; Owner: postgres
--

GRANT USAGE ON SCHEMA auth_fn TO app_usr;


--
-- Name: SCHEMA auth_fn_private; Type: ACL; Schema: -; Owner: postgres
--

GRANT USAGE ON SCHEMA auth_fn_private TO app_usr;


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE USAGE ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO app_auth;


--
-- Name: SCHEMA shard_1; Type: ACL; Schema: -; Owner: postgres
--

GRANT USAGE ON SCHEMA shard_1 TO app_usr;


--
-- Name: FUNCTION id_generator(shard_id integer); Type: ACL; Schema: shard_1; Owner: postgres
--

REVOKE ALL ON FUNCTION shard_1.id_generator(shard_id integer) FROM PUBLIC;
GRANT ALL ON FUNCTION shard_1.id_generator(shard_id integer) TO app_usr;


--
-- Name: TABLE license_pack; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.license_pack TO app_usr;


--
-- Name: TABLE license_pack_license_type; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.license_pack_license_type TO app_usr;


--
-- Name: TABLE app_tenant; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.app_tenant TO app_usr;


--
-- Name: TABLE app_user; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.app_user TO app_usr;


--
-- Name: FUNCTION app_tenant_active_guest_users(_app_tenant app.app_tenant); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.app_tenant_active_guest_users(_app_tenant app.app_tenant) FROM PUBLIC;
GRANT ALL ON FUNCTION app.app_tenant_active_guest_users(_app_tenant app.app_tenant) TO app_usr;


--
-- Name: TABLE app_tenant_subscription; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.app_tenant_subscription TO app_usr;


--
-- Name: FUNCTION app_tenant_active_subscriptions(_app_tenant app.app_tenant); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.app_tenant_active_subscriptions(_app_tenant app.app_tenant) FROM PUBLIC;
GRANT ALL ON FUNCTION app.app_tenant_active_subscriptions(_app_tenant app.app_tenant) TO app_usr;


--
-- Name: FUNCTION app_tenant_active_users(_app_tenant app.app_tenant); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.app_tenant_active_users(_app_tenant app.app_tenant) FROM PUBLIC;
GRANT ALL ON FUNCTION app.app_tenant_active_users(_app_tenant app.app_tenant) TO app_usr;


--
-- Name: FUNCTION app_tenant_available_licenses(_app_tenant app.app_tenant); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.app_tenant_available_licenses(_app_tenant app.app_tenant) FROM PUBLIC;
GRANT ALL ON FUNCTION app.app_tenant_available_licenses(_app_tenant app.app_tenant) TO app_usr;


--
-- Name: FUNCTION app_tenant_flags(_app_tenant app.app_tenant); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.app_tenant_flags(_app_tenant app.app_tenant) FROM PUBLIC;
GRANT ALL ON FUNCTION app.app_tenant_flags(_app_tenant app.app_tenant) TO app_usr;


--
-- Name: FUNCTION app_tenant_inactive_guest_users(_app_tenant app.app_tenant); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.app_tenant_inactive_guest_users(_app_tenant app.app_tenant) FROM PUBLIC;
GRANT ALL ON FUNCTION app.app_tenant_inactive_guest_users(_app_tenant app.app_tenant) TO app_usr;


--
-- Name: FUNCTION app_tenant_inactive_users(_app_tenant app.app_tenant); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.app_tenant_inactive_users(_app_tenant app.app_tenant) FROM PUBLIC;
GRANT ALL ON FUNCTION app.app_tenant_inactive_users(_app_tenant app.app_tenant) TO app_usr;


--
-- Name: FUNCTION app_tenant_license_type_is_available(_app_tenant_id text, _license_type_key text); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.app_tenant_license_type_is_available(_app_tenant_id text, _license_type_key text) FROM PUBLIC;
GRANT ALL ON FUNCTION app.app_tenant_license_type_is_available(_app_tenant_id text, _license_type_key text) TO app_usr;


--
-- Name: FUNCTION app_tenant_payment_status_summary(_app_tenant app.app_tenant); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.app_tenant_payment_status_summary(_app_tenant app.app_tenant) FROM PUBLIC;
GRANT ALL ON FUNCTION app.app_tenant_payment_status_summary(_app_tenant app.app_tenant) TO app_usr;


--
-- Name: FUNCTION app_tenant_subscription_available_add_ons(_app_tenant_subscription app.app_tenant_subscription); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.app_tenant_subscription_available_add_ons(_app_tenant_subscription app.app_tenant_subscription) FROM PUBLIC;
GRANT ALL ON FUNCTION app.app_tenant_subscription_available_add_ons(_app_tenant_subscription app.app_tenant_subscription) TO app_usr;


--
-- Name: FUNCTION app_tenant_subscription_available_licenses(_app_tenant_subscription app.app_tenant_subscription); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.app_tenant_subscription_available_licenses(_app_tenant_subscription app.app_tenant_subscription) FROM PUBLIC;
GRANT ALL ON FUNCTION app.app_tenant_subscription_available_licenses(_app_tenant_subscription app.app_tenant_subscription) TO app_usr;


--
-- Name: FUNCTION app_tenant_subscription_available_upgrade_paths(_app_tenant_subscription app.app_tenant_subscription); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.app_tenant_subscription_available_upgrade_paths(_app_tenant_subscription app.app_tenant_subscription) FROM PUBLIC;
GRANT ALL ON FUNCTION app.app_tenant_subscription_available_upgrade_paths(_app_tenant_subscription app.app_tenant_subscription) TO app_usr;


--
-- Name: TABLE license; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.license TO app_usr;


--
-- Name: FUNCTION app_user_active_licenses(_app_user app.app_user); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.app_user_active_licenses(_app_user app.app_user) FROM PUBLIC;
GRANT ALL ON FUNCTION app.app_user_active_licenses(_app_user app.app_user) TO app_usr;


--
-- Name: FUNCTION app_user_home_path(_app_user app.app_user); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.app_user_home_path(_app_user app.app_user) FROM PUBLIC;
GRANT ALL ON FUNCTION app.app_user_home_path(_app_user app.app_user) TO app_usr;


--
-- Name: FUNCTION app_user_permissions(_app_user app.app_user, _current_app_tenant_id text); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.app_user_permissions(_app_user app.app_user, _current_app_tenant_id text) FROM PUBLIC;
GRANT ALL ON FUNCTION app.app_user_permissions(_app_user app.app_user, _current_app_tenant_id text) TO app_usr;


--
-- Name: FUNCTION app_user_primary_license(_app_user app.app_user); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.app_user_primary_license(_app_user app.app_user) FROM PUBLIC;
GRANT ALL ON FUNCTION app.app_user_primary_license(_app_user app.app_user) TO app_usr;


--
-- Name: FUNCTION application_no_delete(); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.application_no_delete() FROM PUBLIC;
GRANT ALL ON FUNCTION app.application_no_delete() TO app_usr;


--
-- Name: FUNCTION calculate_license_status(_license app.license); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.calculate_license_status(_license app.license) FROM PUBLIC;
GRANT ALL ON FUNCTION app.calculate_license_status(_license app.license) TO app_usr;


--
-- Name: TABLE contact; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.contact TO app_usr;


--
-- Name: FUNCTION contact_full_name(_contact app.contact); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.contact_full_name(_contact app.contact) FROM PUBLIC;
GRANT ALL ON FUNCTION app.contact_full_name(_contact app.contact) TO app_usr;


--
-- Name: FUNCTION contact_has_unanswered_messages(_contact app.contact); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.contact_has_unanswered_messages(_contact app.contact) FROM PUBLIC;
GRANT ALL ON FUNCTION app.contact_has_unanswered_messages(_contact app.contact) TO app_usr;


--
-- Name: FUNCTION fn_ensure_license_status(); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.fn_ensure_license_status() FROM PUBLIC;
GRANT ALL ON FUNCTION app.fn_ensure_license_status() TO app_usr;


--
-- Name: FUNCTION fn_update_eula_trigger(); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.fn_update_eula_trigger() FROM PUBLIC;
GRANT ALL ON FUNCTION app.fn_update_eula_trigger() TO app_usr;


--
-- Name: FUNCTION license_can_activate(_license app.license); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.license_can_activate(_license app.license) FROM PUBLIC;
GRANT ALL ON FUNCTION app.license_can_activate(_license app.license) TO app_usr;


--
-- Name: FUNCTION license_pack_allowed_actions(_license_pack app.license_pack); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.license_pack_allowed_actions(_license_pack app.license_pack) FROM PUBLIC;
GRANT ALL ON FUNCTION app.license_pack_allowed_actions(_license_pack app.license_pack) TO app_usr;


--
-- Name: FUNCTION license_pack_candidate_add_on_keys(_license_pack app.license_pack); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.license_pack_candidate_add_on_keys(_license_pack app.license_pack) FROM PUBLIC;
GRANT ALL ON FUNCTION app.license_pack_candidate_add_on_keys(_license_pack app.license_pack) TO app_usr;


--
-- Name: FUNCTION license_pack_candidate_license_type_keys(_license_pack app.license_pack); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.license_pack_candidate_license_type_keys(_license_pack app.license_pack) FROM PUBLIC;
GRANT ALL ON FUNCTION app.license_pack_candidate_license_type_keys(_license_pack app.license_pack) TO app_usr;


--
-- Name: FUNCTION license_pack_candidate_upgrade_path_keys(_license_pack app.license_pack); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.license_pack_candidate_upgrade_path_keys(_license_pack app.license_pack) FROM PUBLIC;
GRANT ALL ON FUNCTION app.license_pack_candidate_upgrade_path_keys(_license_pack app.license_pack) TO app_usr;


--
-- Name: FUNCTION license_pack_discontinued_add_ons(_license_pack app.license_pack); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.license_pack_discontinued_add_ons(_license_pack app.license_pack) FROM PUBLIC;
GRANT ALL ON FUNCTION app.license_pack_discontinued_add_ons(_license_pack app.license_pack) TO app_usr;


--
-- Name: FUNCTION license_pack_draft_add_ons(_license_pack app.license_pack); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.license_pack_draft_add_ons(_license_pack app.license_pack) FROM PUBLIC;
GRANT ALL ON FUNCTION app.license_pack_draft_add_ons(_license_pack app.license_pack) TO app_usr;


--
-- Name: FUNCTION license_pack_published_add_ons(_license_pack app.license_pack); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.license_pack_published_add_ons(_license_pack app.license_pack) FROM PUBLIC;
GRANT ALL ON FUNCTION app.license_pack_published_add_ons(_license_pack app.license_pack) TO app_usr;


--
-- Name: FUNCTION license_pack_published_implicit_add_ons(_license_pack app.license_pack); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.license_pack_published_implicit_add_ons(_license_pack app.license_pack) FROM PUBLIC;
GRANT ALL ON FUNCTION app.license_pack_published_implicit_add_ons(_license_pack app.license_pack) TO app_usr;


--
-- Name: FUNCTION license_pack_siblings(_license_pack app.license_pack); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.license_pack_siblings(_license_pack app.license_pack) FROM PUBLIC;
GRANT ALL ON FUNCTION app.license_pack_siblings(_license_pack app.license_pack) TO app_usr;


--
-- Name: FUNCTION set_app_tenant_setting_to_default(_app_tenant_id text); Type: ACL; Schema: app; Owner: postgres
--

REVOKE ALL ON FUNCTION app.set_app_tenant_setting_to_default(_app_tenant_id text) FROM PUBLIC;
GRANT ALL ON FUNCTION app.set_app_tenant_setting_to_default(_app_tenant_id text) TO app_usr;


--
-- Name: FUNCTION authenticate_bootstrap(_username text); Type: ACL; Schema: auth_bootstrap; Owner: postgres
--

GRANT ALL ON FUNCTION auth_bootstrap.authenticate_bootstrap(_username text) TO app_anon;


--
-- Name: FUNCTION app_user_has_access(_app_tenant_id text); Type: ACL; Schema: auth_fn; Owner: postgres
--

REVOKE ALL ON FUNCTION auth_fn.app_user_has_access(_app_tenant_id text) FROM PUBLIC;
GRANT ALL ON FUNCTION auth_fn.app_user_has_access(_app_tenant_id text) TO app_usr;


--
-- Name: FUNCTION app_user_has_game_library(_game_id text); Type: ACL; Schema: auth_fn; Owner: postgres
--

REVOKE ALL ON FUNCTION auth_fn.app_user_has_game_library(_game_id text) FROM PUBLIC;
GRANT ALL ON FUNCTION auth_fn.app_user_has_game_library(_game_id text) TO app_usr;


--
-- Name: FUNCTION app_user_has_library(_library_id text); Type: ACL; Schema: auth_fn; Owner: postgres
--

REVOKE ALL ON FUNCTION auth_fn.app_user_has_library(_library_id text) FROM PUBLIC;
GRANT ALL ON FUNCTION auth_fn.app_user_has_library(_library_id text) TO app_usr;


--
-- Name: FUNCTION app_user_has_licensing_scope(_app_tenant_id text); Type: ACL; Schema: auth_fn; Owner: postgres
--

REVOKE ALL ON FUNCTION auth_fn.app_user_has_licensing_scope(_app_tenant_id text) FROM PUBLIC;
GRANT ALL ON FUNCTION auth_fn.app_user_has_licensing_scope(_app_tenant_id text) TO app_usr;


--
-- Name: FUNCTION app_user_has_permission(_permission text); Type: ACL; Schema: auth_fn; Owner: postgres
--

REVOKE ALL ON FUNCTION auth_fn.app_user_has_permission(_permission text) FROM PUBLIC;
GRANT ALL ON FUNCTION auth_fn.app_user_has_permission(_permission text) TO app_usr;


--
-- Name: FUNCTION app_user_has_permission_key(_permission_key app.permission_key); Type: ACL; Schema: auth_fn; Owner: postgres
--

REVOKE ALL ON FUNCTION auth_fn.app_user_has_permission_key(_permission_key app.permission_key) FROM PUBLIC;
GRANT ALL ON FUNCTION auth_fn.app_user_has_permission_key(_permission_key app.permission_key) TO app_usr;


--
-- Name: FUNCTION auth_0_pre_registration(_email text); Type: ACL; Schema: auth_fn; Owner: postgres
--

REVOKE ALL ON FUNCTION auth_fn.auth_0_pre_registration(_email text) FROM PUBLIC;
GRANT ALL ON FUNCTION auth_fn.auth_0_pre_registration(_email text) TO app_usr;


--
-- Name: FUNCTION current_app_user(); Type: ACL; Schema: auth_fn; Owner: postgres
--

REVOKE ALL ON FUNCTION auth_fn.current_app_user() FROM PUBLIC;
GRANT ALL ON FUNCTION auth_fn.current_app_user() TO app_usr;


--
-- Name: FUNCTION current_app_user_id(); Type: ACL; Schema: auth_fn; Owner: postgres
--

REVOKE ALL ON FUNCTION auth_fn.current_app_user_id() FROM PUBLIC;
GRANT ALL ON FUNCTION auth_fn.current_app_user_id() TO app_usr;


--
-- Name: FUNCTION get_app_tenant_scope_permissions(_app_tenant_id text); Type: ACL; Schema: auth_fn; Owner: postgres
--

REVOKE ALL ON FUNCTION auth_fn.get_app_tenant_scope_permissions(_app_tenant_id text) FROM PUBLIC;
GRANT ALL ON FUNCTION auth_fn.get_app_tenant_scope_permissions(_app_tenant_id text) TO app_usr;


--
-- Name: FUNCTION get_app_user_info(_recovery_email_or_id_or_username text, _current_app_tenant_id text); Type: ACL; Schema: auth_fn; Owner: postgres
--

REVOKE ALL ON FUNCTION auth_fn.get_app_user_info(_recovery_email_or_id_or_username text, _current_app_tenant_id text) FROM PUBLIC;
GRANT ALL ON FUNCTION auth_fn.get_app_user_info(_recovery_email_or_id_or_username text, _current_app_tenant_id text) TO app_usr;


--
-- Name: FUNCTION init_app_tenant_support(_app_tenant_id text); Type: ACL; Schema: auth_fn; Owner: postgres
--

REVOKE ALL ON FUNCTION auth_fn.init_app_tenant_support(_app_tenant_id text) FROM PUBLIC;
GRANT ALL ON FUNCTION auth_fn.init_app_tenant_support(_app_tenant_id text) TO app_usr;


--
-- Name: FUNCTION init_demo(_license_pack_key text); Type: ACL; Schema: auth_fn; Owner: postgres
--

REVOKE ALL ON FUNCTION auth_fn.init_demo(_license_pack_key text) FROM PUBLIC;
GRANT ALL ON FUNCTION auth_fn.init_demo(_license_pack_key text) TO app_usr;


--
-- Name: FUNCTION init_subsidiary_admin(_subsidiary_app_tenant_id text); Type: ACL; Schema: auth_fn; Owner: postgres
--

REVOKE ALL ON FUNCTION auth_fn.init_subsidiary_admin(_subsidiary_app_tenant_id text) FROM PUBLIC;
GRANT ALL ON FUNCTION auth_fn.init_subsidiary_admin(_subsidiary_app_tenant_id text) TO app_usr;


--
-- Name: FUNCTION do_get_app_user_info(_recovery_email_or_id_or_username text, _current_app_tenant_id text); Type: ACL; Schema: auth_fn_private; Owner: postgres
--

REVOKE ALL ON FUNCTION auth_fn_private.do_get_app_user_info(_recovery_email_or_id_or_username text, _current_app_tenant_id text) FROM PUBLIC;
GRANT ALL ON FUNCTION auth_fn_private.do_get_app_user_info(_recovery_email_or_id_or_username text, _current_app_tenant_id text) TO app_usr;


--
-- Name: TABLE app_exception; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.app_exception TO app_usr;


--
-- Name: TABLE app_route; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.app_route TO app_usr;


--
-- Name: TABLE app_tenant_group; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.app_tenant_group TO app_usr;


--
-- Name: TABLE app_tenant_group_admin; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.app_tenant_group_admin TO app_usr;


--
-- Name: TABLE app_tenant_group_member; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.app_tenant_group_member TO app_usr;


--
-- Name: TABLE application; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.application TO app_usr;


--
-- Name: TABLE error_report; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.error_report TO app_usr;


--
-- Name: TABLE eula; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.eula TO app_usr;


--
-- Name: TABLE facility; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.facility TO app_usr;


--
-- Name: TABLE license_permission; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.license_permission TO app_usr;


--
-- Name: TABLE license_type; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.license_type TO app_usr;


--
-- Name: TABLE license_type_permission; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.license_type_permission TO app_usr;


--
-- Name: TABLE location; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.location TO app_usr;


--
-- Name: TABLE note; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.note TO app_usr;


--
-- Name: TABLE organization; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.organization TO app_usr;


--
-- Name: TABLE permission; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.permission TO app_usr;


--
-- Name: TABLE registration; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.registration TO app_usr;


--
-- Name: TABLE signed_eula; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.signed_eula TO app_usr;


--
-- Name: TABLE supported_language; Type: ACL; Schema: app; Owner: postgres
--

GRANT ALL ON TABLE app.supported_language TO app_usr;


--
-- Name: postgraphile_watch_ddl; Type: EVENT TRIGGER; Schema: -; Owner: postgres
--

CREATE EVENT TRIGGER postgraphile_watch_ddl ON ddl_command_end
         WHEN TAG IN ('ALTER AGGREGATE', 'ALTER DOMAIN', 'ALTER EXTENSION', 'ALTER FOREIGN TABLE', 'ALTER FUNCTION', 'ALTER POLICY', 'ALTER SCHEMA', 'ALTER TABLE', 'ALTER TYPE', 'ALTER VIEW', 'COMMENT', 'CREATE AGGREGATE', 'CREATE DOMAIN', 'CREATE EXTENSION', 'CREATE FOREIGN TABLE', 'CREATE FUNCTION', 'CREATE INDEX', 'CREATE POLICY', 'CREATE RULE', 'CREATE SCHEMA', 'CREATE TABLE', 'CREATE TABLE AS', 'CREATE VIEW', 'DROP AGGREGATE', 'DROP DOMAIN', 'DROP EXTENSION', 'DROP FOREIGN TABLE', 'DROP FUNCTION', 'DROP INDEX', 'DROP OWNED', 'DROP POLICY', 'DROP RULE', 'DROP SCHEMA', 'DROP TABLE', 'DROP TYPE', 'DROP VIEW', 'GRANT', 'REVOKE', 'SELECT INTO')
   EXECUTE FUNCTION postgraphile_watch.notify_watchers_ddl();


ALTER EVENT TRIGGER postgraphile_watch_ddl OWNER TO postgres;

--
-- Name: postgraphile_watch_drop; Type: EVENT TRIGGER; Schema: -; Owner: postgres
--

CREATE EVENT TRIGGER postgraphile_watch_drop ON sql_drop
   EXECUTE FUNCTION postgraphile_watch.notify_watchers_drop();


ALTER EVENT TRIGGER postgraphile_watch_drop OWNER TO postgres;

--
-- PostgreSQL database dump complete
--

