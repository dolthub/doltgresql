-- Downloaded from: https://github.com/bclynch/edmflare/blob/21bfb9a2dbab36586c75ab0ccc341e0e8f0b118b/data/schema.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 11.4
-- Dumped by pg_dump version 11.4

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

-- manually added these in so they would run on the new db script for RDS
-- normally are dont in the setup script locally. This will probably be overwritten...
-- needed to add graphile-worker schema too with yarn worker install-db-schema command
CREATE ROLE edm_authenticator WITH LOGIN PASSWORD '01481239a706ad342f7e6c5298bff8b3af2be9d0' NOINHERIT;
CREATE ROLE edm_visitor;
GRANT edm_visitor TO edm_authenticator;
REVOKE ALL ON DATABASE edmflare FROM PUBLIC;
GRANT CONNECT ON DATABASE edmflare TO edm;
GRANT CONNECT ON DATABASE edmflare TO edm_authenticator;
GRANT ALL ON DATABASE edmflare TO edm;

--
-- Name: edm; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA edm;


--
-- Name: edm_hidden; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA edm_hidden;


--
-- Name: edm_private; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA edm_private;


--
-- Name: citext; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;


--
-- Name: EXTENSION citext; Type: COMMENT; Schema: -; Owner: -
--



--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--



--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: -
--



--
-- Name: event_type; Type: TYPE; Schema: edm; Owner: -
--

CREATE TYPE edm.event_type AS ENUM (
    'eventbrite',
    'ticketfly',
    'ticketmaster',
    'seetickets',
    'etix',
    'other',
    'livestream'
);


--
-- Name: frequency; Type: TYPE; Schema: edm; Owner: -
--

CREATE TYPE edm.frequency AS ENUM (
    'Every day',
    'Three times a week',
    'Two times a week',
    'Once a week',
    'Once every two weeks',
    'Never'
);


--
-- Name: change_password(text, text); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm.change_password(old_password text, new_password text) RETURNS boolean
    LANGUAGE plpgsql STRICT SECURITY DEFINER
    AS $$
declare
  v_user edm.users;
  v_user_secret edm_private.user_secrets;
begin
  select users.* into v_user
  from edm.users
  where id = edm.current_user_id();

  if not (v_user is null) then
    -- Load their secrets
    select * into v_user_secret from edm_private.user_secrets
    where user_secrets.user_id = v_user.id;

    if v_user_secret.password_hash = crypt(old_password, v_user_secret.password_hash) then
      perform edm_private.assert_valid_password(new_password);
      -- Reset the password as requested
      update edm_private.user_secrets
      set
        password_hash = crypt(new_password, gen_salt('bf'))
      where user_secrets.user_id = v_user.id;
      return true;
    else
      raise exception 'Incorrect password' using errcode = 'CREDS';
    end if;
  else
    raise exception 'You must log in to change your password' using errcode = 'LOGIN';
  end if;
end;
$$;


--
-- Name: FUNCTION change_password(old_password text, new_password text); Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON FUNCTION edm.change_password(old_password text, new_password text) IS 'Enter your old password and a new password to change your password.';


--
-- Name: confirm_account_deletion(text); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm.confirm_account_deletion(token text) RETURNS boolean
    LANGUAGE plpgsql STRICT SECURITY DEFINER
    SET search_path TO '$user', 'public'
    AS $$
declare
  v_user_secret edm_private.user_secrets;
  v_token_max_duration interval = interval '3 days';
begin
  if edm.current_user_id() is null then
    raise exception 'You must log in to delete your account' using errcode = 'LOGIN';
  end if;

  select * into v_user_secret
    from edm_private.user_secrets
    where user_secrets.user_id = edm.current_user_id();

  if v_user_secret is null then
    -- Success: they're already deleted
    return true;
  end if;

  -- Check the token
  if v_user_secret.delete_account_token = token then
    -- Token passes; delete their account :(
    delete from edm.users where id = edm.current_user_id();
    return true;
  end if;

  raise exception 'The supplied token was incorrect - perhaps you''re logged in to the wrong account, or the token has expired?' using errcode = 'DNIED';
end;
$$;


--
-- Name: FUNCTION confirm_account_deletion(token text); Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON FUNCTION edm.confirm_account_deletion(token text) IS 'If you''re certain you want to delete your account, use `requestAccountDeletion` to request an account deletion token, and then supply the token through this mutation to complete account deletion.';


--
-- Name: current_session_id(); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm.current_session_id() RETURNS uuid
    LANGUAGE sql STABLE
    AS $$
  select nullif(pg_catalog.current_setting('jwt.claims.session_id', true), '')::uuid;
$$;


--
-- Name: FUNCTION current_session_id(); Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON FUNCTION edm.current_session_id() IS 'Handy method to get the current session ID.';


SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: users; Type: TABLE; Schema: edm; Owner: -
--

CREATE TABLE edm.users (
    id integer NOT NULL,
    username public.citext NOT NULL,
    name text,
    profile_photo text,
    is_admin boolean DEFAULT false NOT NULL,
    notification_frequency edm.frequency DEFAULT 'Never'::edm.frequency NOT NULL,
    push_notification boolean DEFAULT false NOT NULL,
    email_notification boolean DEFAULT false NOT NULL,
    is_verified boolean DEFAULT false NOT NULL,
    is_setup boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT users_profile_photo_check CHECK ((profile_photo ~ '^https?://[^/]+'::text)),
    CONSTRAINT users_username_check CHECK (((length((username)::text) >= 2) AND (length((username)::text) <= 24) AND (username OPERATOR(public.~) '^[a-zA-Z]([a-zA-Z0-9][_]?)+$'::public.citext)))
);


--
-- Name: TABLE users; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON TABLE edm.users IS 'A user who can log in to the application.';


--
-- Name: COLUMN users.id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.users.id IS 'Unique identifier for the user.';


--
-- Name: COLUMN users.username; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.users.username IS 'Public-facing username (or ''handle'') of the user.';


--
-- Name: COLUMN users.name; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.users.name IS 'Public-facing name (or pseudonym) of the user.';


--
-- Name: COLUMN users.profile_photo; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.users.profile_photo IS 'Optional avatar URL.';


--
-- Name: COLUMN users.is_admin; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.users.is_admin IS 'If true, the user has elevated privileges.';


--
-- Name: COLUMN users.notification_frequency; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.users.notification_frequency IS 'Designates notification frequency';


--
-- Name: COLUMN users.push_notification; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.users.push_notification IS 'Boolean yes or no for push notifications';


--
-- Name: COLUMN users.email_notification; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.users.email_notification IS 'Boolean yes or no for email notifications';


--
-- Name: COLUMN users.is_verified; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.users.is_verified IS 'Boolean that indicates whether user has verified their email or not';


--
-- Name: COLUMN users.is_setup; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.users.is_setup IS 'Boolean that indicates whether user has successfully went through / submitted the user setup process. If not then they get an option in their nav dropdown to do so.';


--
-- Name: current_user(); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm."current_user"() RETURNS edm.users
    LANGUAGE sql STABLE
    AS $$
  select users.* from edm.users where id = edm.current_user_id();
$$;


--
-- Name: FUNCTION "current_user"(); Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON FUNCTION edm."current_user"() IS 'The currently logged in user (or null if not logged in).';


--
-- Name: current_user_id(); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm.current_user_id() RETURNS integer
    LANGUAGE sql STABLE SECURITY DEFINER
    SET search_path TO '$user', 'public'
    AS $$
  select user_id from edm_private.sessions where uuid = edm.current_session_id();
$$;


--
-- Name: FUNCTION current_user_id(); Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON FUNCTION edm.current_user_id() IS 'Handy method to get the current user ID for use in RLS policies, etc; in GraphQL, use `currentUser{id}` instead.';


--
-- Name: forgot_password(public.citext); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm.forgot_password(email public.citext) RETURNS void
    LANGUAGE plpgsql STRICT SECURITY DEFINER
    SET search_path TO '$user', 'public'
    AS $$
declare
  v_user_email edm.user_emails;
  v_token text;
  v_token_min_duration_between_emails interval = interval '3 minutes';
  v_token_max_duration interval = interval '3 days';
  v_now timestamptz = clock_timestamp(); -- Function can be called multiple during transaction
  v_latest_attempt timestamptz;
begin
  -- Find the matching user_email:
  select user_emails.* into v_user_email
  from edm.user_emails
  where user_emails.email = forgot_password.email
  order by is_verified desc, id desc;

  -- If there is no match:
  if v_user_email is null then
    -- This email doesn't exist in the system; trigger an email stating as much.

    -- We do not allow this email to be triggered more than once every 15
    -- minutes, so we need to track it:
    insert into edm_private.unregistered_email_password_resets (email, latest_attempt)
      values (forgot_password.email, v_now)
      on conflict on constraint unregistered_email_pkey
      do update
        set latest_attempt = v_now, attempts = unregistered_email_password_resets.attempts + 1
        where unregistered_email_password_resets.latest_attempt < v_now - interval '15 minutes'
      returning latest_attempt into v_latest_attempt;

    if v_latest_attempt = v_now then
      perform graphile_worker.add_job(
        'user__forgot_password_unregistered_email',
        json_build_object('email', forgot_password.email::text)
      );
    end if;

    -- TODO: we should clear out the unregistered_email_password_resets table periodically.

    return;
  end if;

  -- There was a match.
  -- See if we've triggered a reset recently:
  if exists(
    select 1
    from edm_private.user_email_secrets
    where user_email_id = v_user_email.id
    and password_reset_email_sent_at is not null
    and password_reset_email_sent_at > v_now - v_token_min_duration_between_emails
  ) then
    -- If so, take no action.
    return;
  end if;

  -- Fetch or generate reset token:
  update edm_private.user_secrets
  set
    reset_password_token = (
      case
      when reset_password_token is null or reset_password_token_generated < v_now - v_token_max_duration
      then encode(gen_random_bytes(7), 'hex')
      else reset_password_token
      end
    ),
    reset_password_token_generated = (
      case
      when reset_password_token is null or reset_password_token_generated < v_now - v_token_max_duration
      then v_now
      else reset_password_token_generated
      end
    )
  where user_id = v_user_email.user_id
  returning reset_password_token into v_token;

  -- Don't allow spamming an email:
  update edm_private.user_email_secrets
  set password_reset_email_sent_at = v_now
  where user_email_id = v_user_email.id;

  -- Trigger email send:
  perform graphile_worker.add_job(
    'user__forgot_password',
    json_build_object('id', v_user_email.user_id, 'email', v_user_email.email::text, 'token', v_token)
  );

end;
$$;


--
-- Name: FUNCTION forgot_password(email public.citext); Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON FUNCTION edm.forgot_password(email public.citext) IS 'If you''ve forgotten your password, give us one of your email addresses and we''ll send you a reset token. Note this only works if you have added an email address!';


--
-- Name: logout(); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm.logout() RETURNS void
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO '$user', 'public'
    AS $$
begin
  -- Delete the session
  delete from edm_private.sessions where uuid = edm.current_session_id();
  -- Clear the identifier from the transaction
  perform set_config('jwt.claims.session_id', '', true);
end;
$$;


--
-- Name: user_emails; Type: TABLE; Schema: edm; Owner: -
--

CREATE TABLE edm.user_emails (
    id integer NOT NULL,
    user_id integer DEFAULT edm.current_user_id() NOT NULL,
    email public.citext NOT NULL,
    is_verified boolean DEFAULT false NOT NULL,
    is_primary boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT user_emails_email_check CHECK ((email OPERATOR(public.~) '[^@]+@[^@]+\.[^@]+'::public.citext)),
    CONSTRAINT user_emails_must_be_verified_to_be_primary CHECK (((is_primary IS FALSE) OR (is_verified IS TRUE)))
);


--
-- Name: TABLE user_emails; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON TABLE edm.user_emails IS 'Information about a user''s email address.';


--
-- Name: COLUMN user_emails.email; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.user_emails.email IS 'The users email address, in `a@b.c` format.';


--
-- Name: COLUMN user_emails.is_verified; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.user_emails.is_verified IS 'True if the user has is_verified their email address (by clicking the link in the email we sent them, or logging in with a social login provider), false otherwise.';


--
-- Name: make_email_primary(integer); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm.make_email_primary(email_id integer) RETURNS edm.user_emails
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
declare
  v_user_email edm.user_emails;
begin
  select * into v_user_email from edm.user_emails where id = email_id and user_id = edm.current_user_id();
  if v_user_email is null then
    raise exception 'That''s not your email' using errcode = 'DNIED';
    return null;
  end if;
  if v_user_email.is_verified is false then
    raise exception 'You may not make an unverified email primary' using errcode = 'VRIFY';
  end if;
  update edm.user_emails set is_primary = false where user_id = edm.current_user_id() and is_primary is true and id <> email_id;
  update edm.user_emails set is_primary = true where user_id = edm.current_user_id() and is_primary is not true and id = email_id returning * into v_user_email;
  return v_user_email;
end;
$$;


--
-- Name: FUNCTION make_email_primary(email_id integer); Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON FUNCTION edm.make_email_primary(email_id integer) IS 'Your primary email is where we''ll notify of account events; other emails may be used for discovery or login. Use this when you''re changing your email address.';


--
-- Name: request_account_deletion(); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm.request_account_deletion() RETURNS boolean
    LANGUAGE plpgsql STRICT SECURITY DEFINER
    SET search_path TO '$user', 'public'
    AS $$
declare
  v_user_email edm.user_emails;
  v_token text;
  v_token_max_duration interval = interval '3 days';
begin
  if edm.current_user_id() is null then
    raise exception 'You must log in to delete your account' using errcode = 'LOGIN';
  end if;

  -- Get the email to send account deletion token to
  select * into v_user_email
    from edm.user_emails
    where user_id = edm.current_user_id()
    and is_primary is true;

  -- Fetch or generate token
  update edm_private.user_secrets
  set
    delete_account_token = (
      case
      when delete_account_token is null or delete_account_token_generated < NOW() - v_token_max_duration
      then encode(gen_random_bytes(7), 'hex')
      else delete_account_token
      end
    ),
    delete_account_token_generated = (
      case
      when delete_account_token is null or delete_account_token_generated < NOW() - v_token_max_duration
      then now()
      else delete_account_token_generated
      end
    )
  where user_id = edm.current_user_id()
  returning delete_account_token into v_token;

  -- Trigger email send
  perform graphile_worker.add_job('user__send_delete_account_email', json_build_object('email', v_user_email.email::text, 'token', v_token));
  return true;
end;
$$;


--
-- Name: FUNCTION request_account_deletion(); Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON FUNCTION edm.request_account_deletion() IS 'Begin the account deletion flow by requesting the confirmation email';


--
-- Name: resend_email_verification_code(integer); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm.resend_email_verification_code(email_id integer) RETURNS boolean
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
begin
  if exists(
    select 1
    from edm.user_emails
    where user_emails.id = email_id
    and user_id = edm.current_user_id()
    and is_verified is false
  ) then
    perform graphile_worker.add_job('user_emails__send_verification', json_build_object('id', email_id));
    return true;
  end if;
  return false;
end;
$$;


--
-- Name: FUNCTION resend_email_verification_code(email_id integer); Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON FUNCTION edm.resend_email_verification_code(email_id integer) IS 'If you didn''t receive the verification code for this email, we can resend it. We silently cap the rate of resends on the backend, so calls to this function may not result in another email being sent if it has been called recently.';


--
-- Name: reset_password(integer, text, text); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm.reset_password(user_id integer, reset_token text, new_password text) RETURNS boolean
    LANGUAGE plpgsql STRICT SECURITY DEFINER
    SET search_path TO '$user', 'public'
    AS $$
declare
  v_user edm.users;
  v_user_secret edm_private.user_secrets;
  v_token_max_duration interval = interval '3 days';
begin
  select users.* into v_user
  from edm.users
  where id = user_id;

  if not (v_user is null) then
    -- Load their secrets
    select * into v_user_secret from edm_private.user_secrets
    where user_secrets.user_id = v_user.id;

    -- Have there been too many reset attempts?
    if (
      v_user_secret.first_failed_reset_password_attempt is not null
    and
      v_user_secret.first_failed_reset_password_attempt > NOW() - v_token_max_duration
    and
      v_user_secret.failed_reset_password_attempts >= 20
    ) then
      raise exception 'Password reset locked - too many reset attempts' using errcode = 'LOCKD';
    end if;

    -- Not too many reset attempts, let's check the token
    if v_user_secret.reset_password_token = reset_token then
      -- Excellent - they're legit
      perform edm_private.assert_valid_password(new_password);
      -- Let's reset the password as requested
      update edm_private.user_secrets
      set
        password_hash = crypt(new_password, gen_salt('bf')),
        failed_password_attempts = 0,
        first_failed_password_attempt = null,
        reset_password_token = null,
        reset_password_token_generated = null,
        failed_reset_password_attempts = 0,
        first_failed_reset_password_attempt = null
      where user_secrets.user_id = v_user.id;
      return true;
    else
      -- Wrong token, bump all the attempt tracking figures
      update edm_private.user_secrets
      set
        failed_reset_password_attempts = (case when first_failed_reset_password_attempt is null or first_failed_reset_password_attempt < now() - v_token_max_duration then 1 else failed_reset_password_attempts + 1 end),
        first_failed_reset_password_attempt = (case when first_failed_reset_password_attempt is null or first_failed_reset_password_attempt < now() - v_token_max_duration then now() else first_failed_reset_password_attempt end)
      where user_secrets.user_id = v_user.id;
      return null;
    end if;
  else
    -- No user with that id was found
    return null;
  end if;
end;
$$;


--
-- Name: FUNCTION reset_password(user_id integer, reset_token text, new_password text); Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON FUNCTION edm.reset_password(user_id integer, reset_token text, new_password text) IS 'After triggering forgotPassword, you''ll be sent a reset token. Combine this with your user ID and a new password to reset your password.';


--
-- Name: event; Type: TABLE; Schema: edm; Owner: -
--

CREATE TABLE edm.event (
    id text NOT NULL,
    venue text NOT NULL,
    city integer NOT NULL,
    region text,
    country text,
    name text,
    description text,
    type edm.event_type,
    start_date bigint NOT NULL,
    end_date bigint,
    ticketproviderid text,
    ticketproviderurl text,
    banner text,
    approved boolean DEFAULT false,
    contributor integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT event_name_check CHECK ((char_length(name) < 512))
);


--
-- Name: TABLE event; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON TABLE edm.event IS 'A genre in the application.';


--
-- Name: COLUMN event.id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.event.id IS 'Primary key and id of event.';


--
-- Name: COLUMN event.venue; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.event.venue IS 'Ref to venue where event takes place.';


--
-- Name: COLUMN event.city; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.event.city IS 'Ref to city where event takes place.';


--
-- Name: COLUMN event.region; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.event.region IS 'Ref to region where event takes place.';


--
-- Name: COLUMN event.country; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.event.country IS 'Ref to country where event takes place.';


--
-- Name: COLUMN event.name; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.event.name IS 'Name of event.';


--
-- Name: COLUMN event.description; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.event.description IS 'Description of event.';


--
-- Name: COLUMN event.type; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.event.type IS 'Type of event.';


--
-- Name: COLUMN event.start_date; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.event.start_date IS 'Start date of event.';


--
-- Name: COLUMN event.end_date; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.event.end_date IS 'End date of event.';


--
-- Name: COLUMN event.ticketproviderid; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.event.ticketproviderid IS 'Id by the ticket provider useful for affiliate links.';


--
-- Name: COLUMN event.ticketproviderurl; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.event.ticketproviderurl IS 'URL by the ticket provider useful for affiliate links.';


--
-- Name: COLUMN event.banner; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.event.banner IS 'Banner of event page.';


--
-- Name: COLUMN event.approved; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.event.approved IS 'Whether to display event if it has been approved.';


--
-- Name: COLUMN event.contributor; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.event.contributor IS 'Who submitted the event.';


--
-- Name: search_events_by_city(text, integer); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm.search_events_by_city(query text, cityid integer) RETURNS SETOF edm.event
    LANGUAGE sql STABLE
    AS $$
  select event.*
  from edm.event as event
  where event.city = cityId
  and (
    event.name ilike ('%' || query || '%')
    or event.venue ilike ('%' || query || '%')
  )
  ORDER BY event.start_date ASC
$$;


--
-- Name: FUNCTION search_events_by_city(query text, cityid integer); Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON FUNCTION edm.search_events_by_city(query text, cityid integer) IS 'Returns events by city containing a given query term.';


--
-- Name: search_events_by_region(text, text); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm.search_events_by_region(query text, region_name text) RETURNS SETOF edm.event
    LANGUAGE sql STABLE
    AS $$
  select event.*
  from edm.event as event
  where event.region = region_name
  and (
    event.name ilike ('%' || query || '%')
    or event.venue ilike ('%' || query || '%')
  )
  ORDER BY event.start_date ASC;
$$;


--
-- Name: FUNCTION search_events_by_region(query text, region_name text); Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON FUNCTION edm.search_events_by_region(query text, region_name text) IS 'Returns events by region containing a given query term.';


--
-- Name: tg__graphql_subscription(); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm.tg__graphql_subscription() RETURNS trigger
    LANGUAGE plpgsql
    AS $_$
declare
  v_process_new bool = (TG_OP = 'INSERT' OR TG_OP = 'UPDATE');
  v_process_old bool = (TG_OP = 'UPDATE' OR TG_OP = 'DELETE');
  v_event text = TG_ARGV[0];
  v_topic_template text = TG_ARGV[1];
  v_attribute text = TG_ARGV[2];
  v_record record;
  v_sub text;
  v_topic text;
  v_i int = 0;
  v_last_topic text;
begin
  for v_i in 0..1 loop
    if (v_i = 0) and v_process_new is true then
      v_record = new;
    elsif (v_i = 1) and v_process_old is true then
      v_record = old;
    else
      continue;
    end if;
     if v_attribute is not null then
      execute 'select $1.' || quote_ident(v_attribute)
        using v_record
        into v_sub;
    end if;
    if v_sub is not null then
      v_topic = replace(v_topic_template, '$1', v_sub);
    else
      v_topic = v_topic_template;
    end if;
    if v_topic is distinct from v_last_topic then
      -- This if statement prevents us from triggering the same notification twice
      v_last_topic = v_topic;
      perform pg_notify(v_topic, json_build_object(
        'event', v_event,
        'subject', v_sub
      )::text);
    end if;
  end loop;
  return v_record;
end;
$_$;


--
-- Name: FUNCTION tg__graphql_subscription(); Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON FUNCTION edm.tg__graphql_subscription() IS 'This function enables the creation of simple focussed GraphQL subscriptions using database triggers. Read more here: https://www.graphile.org/postgraphile/subscriptions/#custom-subscriptions';


--
-- Name: tg_user_emails__forbid_if_verified(); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm.tg_user_emails__forbid_if_verified() RETURNS trigger
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO '$user', 'public'
    AS $$
begin
  if exists(select 1 from edm.user_emails where email = NEW.email and is_verified is true) then
    raise exception 'An account using that email address has already been created.' using errcode='EMTKN';
  end if;
  return NEW;
end;
$$;


--
-- Name: tg_user_emails__verify_account_on_verified(); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm.tg_user_emails__verify_account_on_verified() RETURNS trigger
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
begin
  update edm.users set is_verified = true where id = new.user_id and is_verified is false;
  return new;
end;
$$;


--
-- Name: users_has_password(edm.users); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm.users_has_password(u edm.users) RETURNS boolean
    LANGUAGE sql STABLE SECURITY DEFINER
    AS $$
  select (password_hash is not null) from edm_private.user_secrets where user_secrets.user_id = u.id and u.id = edm.current_user_id();
$$;


--
-- Name: verify_email(integer, text); Type: FUNCTION; Schema: edm; Owner: -
--

CREATE FUNCTION edm.verify_email(user_email_id integer, token text) RETURNS boolean
    LANGUAGE plpgsql STRICT SECURITY DEFINER
    AS $$
begin
  update edm.user_emails
  set
    is_verified = true,
    is_primary = is_primary or not exists(
      select 1 from edm.user_emails other_email where other_email.user_id = user_emails.user_id and other_email.is_primary is true
    )
  where id = user_email_id
  and exists(
    select 1 from edm_private.user_email_secrets where user_email_secrets.user_email_id = user_emails.id and verification_token = token
  );
  return found;
end;
$$;


--
-- Name: FUNCTION verify_email(user_email_id integer, token text); Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON FUNCTION edm.verify_email(user_email_id integer, token text) IS 'Once you have received a verification token for your email, you may call this mutation with that token to make your email verified.';


--
-- Name: assert_valid_password(text); Type: FUNCTION; Schema: edm_private; Owner: -
--

CREATE FUNCTION edm_private.assert_valid_password(new_password text) RETURNS void
    LANGUAGE plpgsql
    AS $$
begin
  -- TODO: add better assertions!
  if length(new_password) < 8 then
    raise exception 'Password is too weak' using errcode = 'WEAKP';
  end if;
end;
$$;


--
-- Name: link_or_register_user(integer, character varying, character varying, json, json); Type: FUNCTION; Schema: edm_private; Owner: -
--

CREATE FUNCTION edm_private.link_or_register_user(f_user_id integer, f_service character varying, f_identifier character varying, f_profile json, f_auth_details json) RETURNS edm.users
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO '$user', 'public'
    AS $$
declare
  v_matched_user_id int;
  v_matched_authentication_id int;
  v_email citext;
  v_name text;
  v_profile_photo text;
  v_user edm.users;
  v_user_email edm.user_emails;
begin
  -- See if a user account already matches these details
  select id, user_id
    into v_matched_authentication_id, v_matched_user_id
    from edm.user_authentications
    where service = f_service
    and identifier = f_identifier
    limit 1;

  if v_matched_user_id is not null and f_user_id is not null and v_matched_user_id <> f_user_id then
    raise exception 'A different user already has this account linked.' using errcode = 'TAKEN';
  end if;

  v_email = f_profile ->> 'email';
  v_name := f_profile ->> 'name';
  v_profile_photo := f_profile ->> 'profile_photo';

  if v_matched_authentication_id is null then
    if f_user_id is not null then
      -- Link new account to logged in user account
      insert into edm.user_authentications (user_id, service, identifier, details) values
        (f_user_id, f_service, f_identifier, f_profile) returning id, user_id into v_matched_authentication_id, v_matched_user_id;
      insert into edm_private.user_authentication_secrets (user_authentication_id, details) values
        (v_matched_authentication_id, f_auth_details);
    elsif v_email is not null then
      -- See if the email is registered
      select * into v_user_email from edm.user_emails where email = v_email and is_verified is true;
      if v_user_email is not null then
        -- User exists!
        insert into edm.user_authentications (user_id, service, identifier, details) values
          (v_user_email.user_id, f_service, f_identifier, f_profile) returning id, user_id into v_matched_authentication_id, v_matched_user_id;
        insert into edm_private.user_authentication_secrets (user_authentication_id, details) values
          (v_matched_authentication_id, f_auth_details);
      end if;
    end if;
  end if;
  if v_matched_user_id is null and f_user_id is null and v_matched_authentication_id is null then
    -- Create and return a new user account
    return edm_private.register_user(f_service, f_identifier, f_profile, f_auth_details, true);
  else
    if v_matched_authentication_id is not null then
      update edm.user_authentications
        set details = f_profile
        where id = v_matched_authentication_id;
      update edm_private.user_authentication_secrets
        set details = f_auth_details
        where user_authentication_id = v_matched_authentication_id;
      update edm.users
        set
          name = coalesce(users.name, v_name),
          profile_photo = coalesce(users.profile_photo, v_profile_photo)
        where id = v_matched_user_id
        returning  * into v_user;
      return v_user;
    else
      -- v_matched_authentication_id is null
      -- -> v_matched_user_id is null (they're paired)
      -- -> f_user_id is not null (because the if clause above)
      -- -> v_matched_authentication_id is not null (because of the separate if block above creating a user_authentications)
      -- -> contradiction.
      raise exception 'This should not occur';
    end if;
  end if;
end;
$$;


--
-- Name: FUNCTION link_or_register_user(f_user_id integer, f_service character varying, f_identifier character varying, f_profile json, f_auth_details json); Type: COMMENT; Schema: edm_private; Owner: -
--

COMMENT ON FUNCTION edm_private.link_or_register_user(f_user_id integer, f_service character varying, f_identifier character varying, f_profile json, f_auth_details json) IS 'If you''re logged in, this will link an additional OAuth login to your account if necessary. If you''re logged out it may find if an account already exists (based on OAuth details or email address) and return that, or create a new user account if necessary.';


--
-- Name: sessions; Type: TABLE; Schema: edm_private; Owner: -
--

CREATE TABLE edm_private.sessions (
    uuid uuid DEFAULT public.gen_random_uuid() NOT NULL,
    user_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    last_active timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: login(public.citext, text); Type: FUNCTION; Schema: edm_private; Owner: -
--

CREATE FUNCTION edm_private.login(username public.citext, password text) RETURNS edm_private.sessions
    LANGUAGE plpgsql STRICT SECURITY DEFINER
    AS $$
declare
  v_user edm.users;
  v_user_secret edm_private.user_secrets;
  v_login_attempt_window_duration interval = interval '5 minutes';
  v_session edm_private.sessions;
begin
  if username like '%@%' then
    -- It's an email
    select users.* into v_user
    from edm.users
    inner join edm.user_emails
    on (user_emails.user_id = users.id)
    where user_emails.email = login.username
    order by
      user_emails.is_verified desc, -- Prefer verified email
      user_emails.created_at asc -- Failing that, prefer the first registered (unverified users _should_ verify before logging in)
    limit 1;
  else
    -- It's a username
    select users.* into v_user
    from edm.users
    where users.username = login.username;
  end if;

  if not (v_user is null) then
    -- Load their secrets
    select * into v_user_secret from edm_private.user_secrets
    where user_secrets.user_id = v_user.id;

    -- Have there been too many login attempts?
    if (
      v_user_secret.first_failed_password_attempt is not null
    and
      v_user_secret.first_failed_password_attempt > NOW() - v_login_attempt_window_duration
    and
      v_user_secret.failed_password_attempts >= 3
    ) then
      raise exception 'User account locked - too many login attempts. Try again after 5 minutes.' using errcode = 'LOCKD';
    end if;

    -- Not too many login attempts, let's check the password.
    -- NOTE: `password_hash` could be null, this is fine since `NULL = NULL` is null, and null is falsy.
    if v_user_secret.password_hash = crypt(password, v_user_secret.password_hash) then
      -- Excellent - they're logged in! Let's reset the attempt tracking
      update edm_private.user_secrets
      set failed_password_attempts = 0, first_failed_password_attempt = null, last_login_at = now()
      where user_id = v_user.id;
      -- Create a session for the user
      insert into edm_private.sessions (user_id) values (v_user.id) returning * into v_session;
      -- And finally return the session
      return v_session;
    else
      -- Wrong password, bump all the attempt tracking figures
      update edm_private.user_secrets
      set
        failed_password_attempts = (case when first_failed_password_attempt is null or first_failed_password_attempt < now() - v_login_attempt_window_duration then 1 else failed_password_attempts + 1 end),
        first_failed_password_attempt = (case when first_failed_password_attempt is null or first_failed_password_attempt < now() - v_login_attempt_window_duration then now() else first_failed_password_attempt end)
      where user_id = v_user.id;
      return null; -- Must not throw otherwise transaction will be aborted and attempts won't be recorded
    end if;
  else
    -- No user with that email/username was found
    return null;
  end if;
end;
$$;


--
-- Name: FUNCTION login(username public.citext, password text); Type: COMMENT; Schema: edm_private; Owner: -
--

COMMENT ON FUNCTION edm_private.login(username public.citext, password text) IS 'Returns a user that matches the username/password combo, or null on failure.';


--
-- Name: really_create_user(public.citext, text, boolean, text, text, text); Type: FUNCTION; Schema: edm_private; Owner: -
--

CREATE FUNCTION edm_private.really_create_user(username public.citext, email text, email_is_verified boolean, name text, profile_photo text, password text DEFAULT NULL::text) RETURNS edm.users
    LANGUAGE plpgsql
    SET search_path TO '$user', 'public'
    AS $$
declare
  v_user edm.users;
  v_username citext = username;
begin
  if password is not null then
    perform edm_private.assert_valid_password(password);
  end if;
  if email is null then
    raise exception 'Email is required' using errcode = 'MODAT';
  end if;

  -- Insert the new user
  insert into edm.users (username, name, profile_photo) values
    (v_username, name, profile_photo)
    returning * into v_user;

	-- Add the user's email
  insert into edm.user_emails (user_id, email, is_verified, is_primary)
  values (v_user.id, email, email_is_verified, email_is_verified);

  -- Store the password
  if password is not null then
    update edm_private.user_secrets
    set password_hash = crypt(password, gen_salt('bf'))
    where user_id = v_user.id;
  end if;

  -- Refresh the user
  select * into v_user from edm.users where id = v_user.id;

  return v_user;
end;
$$;


--
-- Name: FUNCTION really_create_user(username public.citext, email text, email_is_verified boolean, name text, profile_photo text, password text); Type: COMMENT; Schema: edm_private; Owner: -
--

COMMENT ON FUNCTION edm_private.really_create_user(username public.citext, email text, email_is_verified boolean, name text, profile_photo text, password text) IS 'Creates a user account. All arguments are optional, it trusts the calling method to perform sanitisation.';


--
-- Name: register_user(character varying, character varying, json, json, boolean); Type: FUNCTION; Schema: edm_private; Owner: -
--

CREATE FUNCTION edm_private.register_user(f_service character varying, f_identifier character varying, f_profile json, f_auth_details json, f_email_is_verified boolean DEFAULT false) RETURNS edm.users
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO '$user', 'public'
    AS $$
declare
  v_user edm.users;
  v_email citext;
  v_name text;
  v_username citext;
  v_profile_photo text;
  v_user_authentication_id int;
begin
  -- Extract data from the user’s OAuth profile data.
  v_email := f_profile ->> 'email';
  v_name := f_profile ->> 'name';
  v_username := f_profile ->> 'username';
  v_profile_photo := f_profile ->> 'profile_photo';

  -- Sanitise the username, and make it unique if necessary.
  if v_username is null then
    v_username = coalesce(v_name, 'user');
  end if;
  v_username = regexp_replace(v_username, '^[^a-z]+', '', 'i');
  v_username = regexp_replace(v_username, '[^a-z0-9]+', '_', 'i');
  if v_username is null or length(v_username) < 3 then
    v_username = 'user';
  end if;
  select (
    case
    when i = 0 then v_username
    else v_username || i::text
    end
  ) into v_username from generate_series(0, 1000) i
  where not exists(
    select 1
    from edm.users
    where users.username = (
      case
      when i = 0 then v_username
      else v_username || i::text
      end
    )
  )
  limit 1;

  -- Create the user account
  v_user = edm_private.really_create_user(
    username => v_username,
    email => v_email,
    email_is_verified => f_email_is_verified,
    name => v_name,
    profile_photo => v_profile_photo
  );

  -- Insert the user’s private account data (e.g. OAuth tokens)
  insert into edm.user_authentications (user_id, service, identifier, details) values
    (v_user.id, f_service, f_identifier, f_profile) returning id into v_user_authentication_id;
  insert into edm_private.user_authentication_secrets (user_authentication_id, details) values
    (v_user_authentication_id, f_auth_details);

  return v_user;
end;
$$;


--
-- Name: FUNCTION register_user(f_service character varying, f_identifier character varying, f_profile json, f_auth_details json, f_email_is_verified boolean); Type: COMMENT; Schema: edm_private; Owner: -
--

COMMENT ON FUNCTION edm_private.register_user(f_service character varying, f_identifier character varying, f_profile json, f_auth_details json, f_email_is_verified boolean) IS 'Used to register a user from information gleaned from OAuth. Primarily used by link_or_register_user';


--
-- Name: tg__add_job(); Type: FUNCTION; Schema: edm_private; Owner: -
--

CREATE FUNCTION edm_private.tg__add_job() RETURNS trigger
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO '$user', 'public'
    AS $$
begin
  perform graphile_worker.add_job(tg_argv[0], json_build_object('id', NEW.id), coalesce(tg_argv[1], public.gen_random_uuid()::text));
  return NEW;
end;
$$;


--
-- Name: FUNCTION tg__add_job(); Type: COMMENT; Schema: edm_private; Owner: -
--

COMMENT ON FUNCTION edm_private.tg__add_job() IS 'Useful shortcut to create a job on insert/update. Pass the task name as the first trigger argument, and optionally the queue name as the second argument. The record id will automatically be available on the JSON payload.';


--
-- Name: tg__timestamps(); Type: FUNCTION; Schema: edm_private; Owner: -
--

CREATE FUNCTION edm_private.tg__timestamps() RETURNS trigger
    LANGUAGE plpgsql
    SET search_path TO '$user', 'public'
    AS $$
begin
  NEW.created_at = (case when TG_OP = 'INSERT' then NOW() else OLD.created_at end);
  NEW.updated_at = (case when TG_OP = 'UPDATE' and OLD.updated_at >= NOW() then OLD.updated_at + interval '1 millisecond' else NOW() end);
  return NEW;
end;
$$;


--
-- Name: FUNCTION tg__timestamps(); Type: COMMENT; Schema: edm_private; Owner: -
--

COMMENT ON FUNCTION edm_private.tg__timestamps() IS 'This trigger should be called on all tables with created_at, updated_at - it ensures that they cannot be manipulated and that updated_at will always be larger than the previous updated_at.';


--
-- Name: tg_user_email_secrets__insert_with_user_email(); Type: FUNCTION; Schema: edm_private; Owner: -
--

CREATE FUNCTION edm_private.tg_user_email_secrets__insert_with_user_email() RETURNS trigger
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO '$user', 'public'
    AS $$
declare
  v_verification_token text;
begin
  if NEW.is_verified is false then
    v_verification_token = encode(gen_random_bytes(7), 'hex');
  end if;
  insert into edm_private.user_email_secrets(user_email_id, verification_token) values(NEW.id, v_verification_token);
  return NEW;
end;
$$;


--
-- Name: FUNCTION tg_user_email_secrets__insert_with_user_email(); Type: COMMENT; Schema: edm_private; Owner: -
--

COMMENT ON FUNCTION edm_private.tg_user_email_secrets__insert_with_user_email() IS 'Ensures that every user_email record has an associated user_email_secret record.';


--
-- Name: tg_user_secrets__insert_with_user(); Type: FUNCTION; Schema: edm_private; Owner: -
--

CREATE FUNCTION edm_private.tg_user_secrets__insert_with_user() RETURNS trigger
    LANGUAGE plpgsql
    SET search_path TO '$user', 'public'
    AS $$
begin
  insert into edm_private.user_secrets(user_id) values(NEW.id);
  return NEW;
end;
$$;


--
-- Name: FUNCTION tg_user_secrets__insert_with_user(); Type: COMMENT; Schema: edm_private; Owner: -
--

COMMENT ON FUNCTION edm_private.tg_user_secrets__insert_with_user() IS 'Ensures that every user record has an associated user_secret record.';


--
-- Name: tg_users__make_first_user_admin(); Type: FUNCTION; Schema: edm_private; Owner: -
--

CREATE FUNCTION edm_private.tg_users__make_first_user_admin() RETURNS trigger
    LANGUAGE plpgsql
    SET search_path TO '$user', 'public'
    AS $$
begin
  NEW.is_admin = true;
  return NEW;
end;
$$;


--
-- Name: artist; Type: TABLE; Schema: edm; Owner: -
--

CREATE TABLE edm.artist (
    name text NOT NULL,
    description text,
    photo text,
    twitter_username text,
    twitter_url text,
    facebook_username text,
    facebook_url text,
    instagram_username text,
    instagram_url text,
    soundcloud_username text,
    soundcloud_url text,
    youtube_username text,
    youtube_url text,
    spotify_url text,
    homepage text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE artist; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON TABLE edm.artist IS 'An artist in the application.';


--
-- Name: COLUMN artist.name; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist.name IS 'Primary key and name of artist.';


--
-- Name: COLUMN artist.description; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist.description IS 'Description of the artist.';


--
-- Name: COLUMN artist.photo; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist.photo IS 'Photo of the artist.';


--
-- Name: COLUMN artist.twitter_username; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist.twitter_username IS 'Twitter username of the artist.';


--
-- Name: COLUMN artist.twitter_url; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist.twitter_url IS 'Twitter url of the artist.';


--
-- Name: COLUMN artist.facebook_username; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist.facebook_username IS 'Facebook username of the artist.';


--
-- Name: COLUMN artist.facebook_url; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist.facebook_url IS 'Facebook url of the artist.';


--
-- Name: COLUMN artist.instagram_username; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist.instagram_username IS 'Instagram username of the artist.';


--
-- Name: COLUMN artist.instagram_url; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist.instagram_url IS 'Instagram url of the artist.';


--
-- Name: COLUMN artist.soundcloud_username; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist.soundcloud_username IS 'Soundcloud username of the artist.';


--
-- Name: COLUMN artist.soundcloud_url; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist.soundcloud_url IS 'Soundcloud url of the artist.';


--
-- Name: COLUMN artist.youtube_username; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist.youtube_username IS 'Youtube username of the artist.';


--
-- Name: COLUMN artist.youtube_url; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist.youtube_url IS 'Youtube url of the artist.';


--
-- Name: COLUMN artist.spotify_url; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist.spotify_url IS 'Spotify url of the artist.';


--
-- Name: COLUMN artist.homepage; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist.homepage IS 'Homepage url of the artist.';


--
-- Name: artist_to_event; Type: TABLE; Schema: edm; Owner: -
--

CREATE TABLE edm.artist_to_event (
    id integer NOT NULL,
    artist_id text NOT NULL,
    event_id text NOT NULL
);


--
-- Name: TABLE artist_to_event; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON TABLE edm.artist_to_event IS 'A join table for artists at an event.';


--
-- Name: COLUMN artist_to_event.id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist_to_event.id IS 'Primary key and id of row.';


--
-- Name: COLUMN artist_to_event.artist_id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist_to_event.artist_id IS 'Ref to artist.';


--
-- Name: COLUMN artist_to_event.event_id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.artist_to_event.event_id IS 'Ref to event.';


--
-- Name: artist_to_event_id_seq; Type: SEQUENCE; Schema: edm; Owner: -
--

CREATE SEQUENCE edm.artist_to_event_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: artist_to_event_id_seq; Type: SEQUENCE OWNED BY; Schema: edm; Owner: -
--

ALTER SEQUENCE edm.artist_to_event_id_seq OWNED BY edm.artist_to_event.id;


--
-- Name: city; Type: TABLE; Schema: edm; Owner: -
--

CREATE TABLE edm.city (
    id integer NOT NULL,
    name text,
    description text,
    photo text,
    region text,
    country text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE city; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON TABLE edm.city IS 'A city in the application.';


--
-- Name: COLUMN city.id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.city.id IS 'Primary key and id for city.';


--
-- Name: COLUMN city.name; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.city.name IS 'Name for city.';


--
-- Name: COLUMN city.description; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.city.description IS 'Description of the genre.';


--
-- Name: COLUMN city.photo; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.city.photo IS 'Photo for city.';


--
-- Name: COLUMN city.region; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.city.region IS 'Region ref for city.';


--
-- Name: COLUMN city.country; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.city.country IS 'Region ref for country.';


--
-- Name: city_sequence; Type: SEQUENCE; Schema: edm; Owner: -
--

CREATE SEQUENCE edm.city_sequence
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: city_sequence; Type: SEQUENCE OWNED BY; Schema: edm; Owner: -
--

ALTER SEQUENCE edm.city_sequence OWNED BY edm.city.id;


--
-- Name: country; Type: TABLE; Schema: edm; Owner: -
--

CREATE TABLE edm.country (
    code text NOT NULL,
    name text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE country; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON TABLE edm.country IS 'A country in the application.';


--
-- Name: COLUMN country.code; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.country.code IS 'Primary key and code for country.';


--
-- Name: COLUMN country.name; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.country.name IS 'Name for country.';


--
-- Name: follow_list; Type: TABLE; Schema: edm; Owner: -
--

CREATE TABLE edm.follow_list (
    id integer NOT NULL,
    user_id integer NOT NULL,
    artist_id text,
    venue_id text
);


--
-- Name: TABLE follow_list; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON TABLE edm.follow_list IS 'Join table for followed artists or venues by a user.';


--
-- Name: COLUMN follow_list.id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.follow_list.id IS 'Primary key and id of row.';


--
-- Name: COLUMN follow_list.user_id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.follow_list.user_id IS 'Ref to user.';


--
-- Name: COLUMN follow_list.artist_id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.follow_list.artist_id IS 'Ref to artist.';


--
-- Name: COLUMN follow_list.venue_id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.follow_list.venue_id IS 'Ref to venue.';


--
-- Name: follow_list_id_seq; Type: SEQUENCE; Schema: edm; Owner: -
--

CREATE SEQUENCE edm.follow_list_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: follow_list_id_seq; Type: SEQUENCE OWNED BY; Schema: edm; Owner: -
--

ALTER SEQUENCE edm.follow_list_id_seq OWNED BY edm.follow_list.id;


--
-- Name: genre; Type: TABLE; Schema: edm; Owner: -
--

CREATE TABLE edm.genre (
    name text NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT genre_description_check CHECK ((char_length(description) < 2400)),
    CONSTRAINT genre_name_check CHECK ((char_length(name) < 80))
);


--
-- Name: TABLE genre; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON TABLE edm.genre IS 'A genre in the application.';


--
-- Name: COLUMN genre.name; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.genre.name IS 'Primary key and name of genre.';


--
-- Name: genre_to_artist; Type: TABLE; Schema: edm; Owner: -
--

CREATE TABLE edm.genre_to_artist (
    id integer NOT NULL,
    genre_id text NOT NULL,
    artist_id text NOT NULL
);


--
-- Name: TABLE genre_to_artist; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON TABLE edm.genre_to_artist IS 'A join table between genres and artists.';


--
-- Name: COLUMN genre_to_artist.id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.genre_to_artist.id IS 'Id of the row.';


--
-- Name: COLUMN genre_to_artist.genre_id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.genre_to_artist.genre_id IS 'Ref to the genre.';


--
-- Name: COLUMN genre_to_artist.artist_id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.genre_to_artist.artist_id IS 'Ref to the artist.';


--
-- Name: genre_to_artist_id_seq; Type: SEQUENCE; Schema: edm; Owner: -
--

CREATE SEQUENCE edm.genre_to_artist_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: genre_to_artist_id_seq; Type: SEQUENCE OWNED BY; Schema: edm; Owner: -
--

ALTER SEQUENCE edm.genre_to_artist_id_seq OWNED BY edm.genre_to_artist.id;


--
-- Name: push_subscription; Type: TABLE; Schema: edm; Owner: -
--

CREATE TABLE edm.push_subscription (
    id integer NOT NULL,
    user_id integer NOT NULL,
    endpoint text NOT NULL,
    expiration_time timestamp without time zone,
    p256dh text NOT NULL,
    auth text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE push_subscription; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON TABLE edm.push_subscription IS 'A table with push subscription info.';


--
-- Name: COLUMN push_subscription.id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.push_subscription.id IS 'Unique identifier for the push subscription.';


--
-- Name: COLUMN push_subscription.user_id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.push_subscription.user_id IS 'Reference to the account this belongs to.';


--
-- Name: COLUMN push_subscription.endpoint; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.push_subscription.endpoint IS 'This contains a unique URL to a Firebase Cloud Messaging endpoint. This url is a public but unguessable endpoint to the Browser Push Service used by the application server to send push notifications to this subscription.';


--
-- Name: COLUMN push_subscription.expiration_time; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.push_subscription.expiration_time IS 'This is useful in certain cases, for example, if a message might contain an authentication code that expires after 1 minute.';


--
-- Name: COLUMN push_subscription.p256dh; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.push_subscription.p256dh IS 'An encryption key that our server will use to encrypt the message.';


--
-- Name: COLUMN push_subscription.auth; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.push_subscription.auth IS 'An authentication secret, which is one of the inputs of the message content encryption process.';


--
-- Name: push_subscription_id_seq; Type: SEQUENCE; Schema: edm; Owner: -
--

CREATE SEQUENCE edm.push_subscription_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: push_subscription_id_seq; Type: SEQUENCE OWNED BY; Schema: edm; Owner: -
--

ALTER SEQUENCE edm.push_subscription_id_seq OWNED BY edm.push_subscription.id;


--
-- Name: region; Type: TABLE; Schema: edm; Owner: -
--

CREATE TABLE edm.region (
    name text NOT NULL,
    description text,
    photo text,
    country text,
    lat numeric,
    lon numeric,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE region; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON TABLE edm.region IS 'A table with regions.';


--
-- Name: COLUMN region.name; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.region.name IS 'Name and primary key of region.';


--
-- Name: COLUMN region.description; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.region.description IS 'Description of the region.';


--
-- Name: COLUMN region.photo; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.region.photo IS 'Photo of the region.';


--
-- Name: COLUMN region.country; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.region.country IS 'Country ref region belongs to.';


--
-- Name: COLUMN region.lat; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.region.lat IS 'Latitude location of the region.';


--
-- Name: COLUMN region.lon; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.region.lon IS 'Longitude location of the region.';


--
-- Name: user_authentications; Type: TABLE; Schema: edm; Owner: -
--

CREATE TABLE edm.user_authentications (
    id integer NOT NULL,
    user_id integer NOT NULL,
    service text NOT NULL,
    identifier text NOT NULL,
    details jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE user_authentications; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON TABLE edm.user_authentications IS 'Contains information about the login providers this user has used, so that they may disconnect them should they wish.';


--
-- Name: COLUMN user_authentications.service; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.user_authentications.service IS 'The login service used, e.g. `twitter` or `github`.';


--
-- Name: COLUMN user_authentications.identifier; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.user_authentications.identifier IS 'A unique identifier for the user within the login service.';


--
-- Name: COLUMN user_authentications.details; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.user_authentications.details IS 'Additional profile details extracted from this login method';


--
-- Name: user_authentications_id_seq; Type: SEQUENCE; Schema: edm; Owner: -
--

CREATE SEQUENCE edm.user_authentications_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_authentications_id_seq; Type: SEQUENCE OWNED BY; Schema: edm; Owner: -
--

ALTER SEQUENCE edm.user_authentications_id_seq OWNED BY edm.user_authentications.id;


--
-- Name: user_emails_id_seq; Type: SEQUENCE; Schema: edm; Owner: -
--

CREATE SEQUENCE edm.user_emails_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_emails_id_seq; Type: SEQUENCE OWNED BY; Schema: edm; Owner: -
--

ALTER SEQUENCE edm.user_emails_id_seq OWNED BY edm.user_emails.id;


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: edm; Owner: -
--

CREATE SEQUENCE edm.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: edm; Owner: -
--

ALTER SEQUENCE edm.users_id_seq OWNED BY edm.users.id;


--
-- Name: venue; Type: TABLE; Schema: edm; Owner: -
--

CREATE TABLE edm.venue (
    name text NOT NULL,
    description text,
    lat numeric,
    lon numeric,
    city integer NOT NULL,
    address text,
    photo text,
    logo text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT venue_address_check CHECK ((char_length(address) < 512)),
    CONSTRAINT venue_description_check CHECK ((char_length(description) < 2400)),
    CONSTRAINT venue_name_check CHECK ((char_length(name) < 256))
);


--
-- Name: TABLE venue; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON TABLE edm.venue IS 'A venue in the application.';


--
-- Name: COLUMN venue.name; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.venue.name IS 'Primary key and name of venue.';


--
-- Name: COLUMN venue.description; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.venue.description IS 'Description of venue.';


--
-- Name: COLUMN venue.lat; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.venue.lat IS 'Latitude of venue.';


--
-- Name: COLUMN venue.lon; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.venue.lon IS 'Longitude of venue.';


--
-- Name: COLUMN venue.city; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.venue.city IS 'Ref to city of venue.';


--
-- Name: COLUMN venue.address; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.venue.address IS 'Address of venue.';


--
-- Name: COLUMN venue.photo; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.venue.photo IS 'Photo of venue.';


--
-- Name: COLUMN venue.logo; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.venue.logo IS 'Logo of venue.';


--
-- Name: watch_list; Type: TABLE; Schema: edm; Owner: -
--

CREATE TABLE edm.watch_list (
    id integer NOT NULL,
    user_id integer NOT NULL,
    event_id text NOT NULL
);


--
-- Name: TABLE watch_list; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON TABLE edm.watch_list IS 'Join table for events watched by a user.';


--
-- Name: COLUMN watch_list.id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.watch_list.id IS 'Primary key and id of row.';


--
-- Name: COLUMN watch_list.user_id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.watch_list.user_id IS 'Ref to user.';


--
-- Name: COLUMN watch_list.event_id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.watch_list.event_id IS 'Ref to event.';


--
-- Name: watch_list_id_seq; Type: SEQUENCE; Schema: edm; Owner: -
--

CREATE SEQUENCE edm.watch_list_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: watch_list_id_seq; Type: SEQUENCE OWNED BY; Schema: edm; Owner: -
--

ALTER SEQUENCE edm.watch_list_id_seq OWNED BY edm.watch_list.id;


--
-- Name: watched_to_account; Type: TABLE; Schema: edm; Owner: -
--

CREATE TABLE edm.watched_to_account (
    id integer NOT NULL,
    user_id integer NOT NULL,
    region text,
    city_id integer
);


--
-- Name: TABLE watched_to_account; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON TABLE edm.watched_to_account IS 'A join table for watched location to an account.';


--
-- Name: COLUMN watched_to_account.id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.watched_to_account.id IS 'Id of the row.';


--
-- Name: COLUMN watched_to_account.user_id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.watched_to_account.user_id IS 'Ref to user account.';


--
-- Name: COLUMN watched_to_account.region; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.watched_to_account.region IS 'Ref to region.';


--
-- Name: COLUMN watched_to_account.city_id; Type: COMMENT; Schema: edm; Owner: -
--

COMMENT ON COLUMN edm.watched_to_account.city_id IS 'Ref to city.';


--
-- Name: watched_to_account_id_seq; Type: SEQUENCE; Schema: edm; Owner: -
--

CREATE SEQUENCE edm.watched_to_account_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: watched_to_account_id_seq; Type: SEQUENCE OWNED BY; Schema: edm; Owner: -
--

ALTER SEQUENCE edm.watched_to_account_id_seq OWNED BY edm.watched_to_account.id;


--
-- Name: connect_pg_simple_sessions; Type: TABLE; Schema: edm_private; Owner: -
--

CREATE TABLE edm_private.connect_pg_simple_sessions (
    sid character varying NOT NULL,
    sess json NOT NULL,
    expire timestamp without time zone NOT NULL
);


--
-- Name: unregistered_email_password_resets; Type: TABLE; Schema: edm_private; Owner: -
--

CREATE TABLE edm_private.unregistered_email_password_resets (
    email public.citext NOT NULL,
    attempts integer DEFAULT 1 NOT NULL,
    latest_attempt timestamp with time zone NOT NULL
);


--
-- Name: TABLE unregistered_email_password_resets; Type: COMMENT; Schema: edm_private; Owner: -
--

COMMENT ON TABLE edm_private.unregistered_email_password_resets IS 'If someone tries to recover the password for an email that is not registered in our system, this table enables us to rate-limit outgoing emails to avoid spamming.';


--
-- Name: COLUMN unregistered_email_password_resets.attempts; Type: COMMENT; Schema: edm_private; Owner: -
--

COMMENT ON COLUMN edm_private.unregistered_email_password_resets.attempts IS 'We store the number of attempts to help us detect accounts being attacked.';


--
-- Name: COLUMN unregistered_email_password_resets.latest_attempt; Type: COMMENT; Schema: edm_private; Owner: -
--

COMMENT ON COLUMN edm_private.unregistered_email_password_resets.latest_attempt IS 'We store the time the last password reset was sent to this email to prevent the email getting flooded.';


--
-- Name: user_authentication_secrets; Type: TABLE; Schema: edm_private; Owner: -
--

CREATE TABLE edm_private.user_authentication_secrets (
    user_authentication_id integer NOT NULL,
    details jsonb DEFAULT '{}'::jsonb NOT NULL
);


--
-- Name: user_email_secrets; Type: TABLE; Schema: edm_private; Owner: -
--

CREATE TABLE edm_private.user_email_secrets (
    user_email_id integer NOT NULL,
    verification_token text,
    verification_email_sent_at timestamp with time zone,
    password_reset_email_sent_at timestamp with time zone
);


--
-- Name: TABLE user_email_secrets; Type: COMMENT; Schema: edm_private; Owner: -
--

COMMENT ON TABLE edm_private.user_email_secrets IS 'The contents of this table should never be visible to the user. Contains data mostly related to email verification and avoiding spamming users.';


--
-- Name: COLUMN user_email_secrets.password_reset_email_sent_at; Type: COMMENT; Schema: edm_private; Owner: -
--

COMMENT ON COLUMN edm_private.user_email_secrets.password_reset_email_sent_at IS 'We store the time the last password reset was sent to this email to prevent the email getting flooded.';


--
-- Name: user_secrets; Type: TABLE; Schema: edm_private; Owner: -
--

CREATE TABLE edm_private.user_secrets (
    user_id integer NOT NULL,
    password_hash text,
    last_login_at timestamp with time zone DEFAULT now() NOT NULL,
    failed_password_attempts integer DEFAULT 0 NOT NULL,
    first_failed_password_attempt timestamp with time zone,
    reset_password_token text,
    reset_password_token_generated timestamp with time zone,
    failed_reset_password_attempts integer DEFAULT 0 NOT NULL,
    first_failed_reset_password_attempt timestamp with time zone,
    delete_account_token text,
    delete_account_token_generated timestamp with time zone
);


--
-- Name: TABLE user_secrets; Type: COMMENT; Schema: edm_private; Owner: -
--

COMMENT ON TABLE edm_private.user_secrets IS 'The contents of this table should never be visible to the user. Contains data mostly related to authentication.';


--
-- Name: migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.migrations (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    run_on timestamp without time zone NOT NULL
);


--
-- Name: migrations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.migrations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: migrations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.migrations_id_seq OWNED BY public.migrations.id;


--
-- Name: artist_to_event id; Type: DEFAULT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.artist_to_event ALTER COLUMN id SET DEFAULT nextval('edm.artist_to_event_id_seq'::regclass);


--
-- Name: city id; Type: DEFAULT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.city ALTER COLUMN id SET DEFAULT nextval('edm.city_sequence'::regclass);


--
-- Name: follow_list id; Type: DEFAULT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.follow_list ALTER COLUMN id SET DEFAULT nextval('edm.follow_list_id_seq'::regclass);


--
-- Name: genre_to_artist id; Type: DEFAULT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.genre_to_artist ALTER COLUMN id SET DEFAULT nextval('edm.genre_to_artist_id_seq'::regclass);


--
-- Name: push_subscription id; Type: DEFAULT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.push_subscription ALTER COLUMN id SET DEFAULT nextval('edm.push_subscription_id_seq'::regclass);


--
-- Name: user_authentications id; Type: DEFAULT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.user_authentications ALTER COLUMN id SET DEFAULT nextval('edm.user_authentications_id_seq'::regclass);


--
-- Name: user_emails id; Type: DEFAULT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.user_emails ALTER COLUMN id SET DEFAULT nextval('edm.user_emails_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.users ALTER COLUMN id SET DEFAULT nextval('edm.users_id_seq'::regclass);


--
-- Name: watch_list id; Type: DEFAULT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.watch_list ALTER COLUMN id SET DEFAULT nextval('edm.watch_list_id_seq'::regclass);


--
-- Name: watched_to_account id; Type: DEFAULT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.watched_to_account ALTER COLUMN id SET DEFAULT nextval('edm.watched_to_account_id_seq'::regclass);


--
-- Name: migrations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.migrations ALTER COLUMN id SET DEFAULT nextval('public.migrations_id_seq'::regclass);


--
-- Name: artist artist_pkey; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.artist
    ADD CONSTRAINT artist_pkey PRIMARY KEY (name);


--
-- Name: artist_to_event artist_to_event_pkey; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.artist_to_event
    ADD CONSTRAINT artist_to_event_pkey PRIMARY KEY (id);


--
-- Name: city city_pkey; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.city
    ADD CONSTRAINT city_pkey PRIMARY KEY (id);


--
-- Name: country country_pkey; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.country
    ADD CONSTRAINT country_pkey PRIMARY KEY (code);


--
-- Name: event event_pkey; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.event
    ADD CONSTRAINT event_pkey PRIMARY KEY (id);


--
-- Name: follow_list follow_list_pkey; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.follow_list
    ADD CONSTRAINT follow_list_pkey PRIMARY KEY (id);


--
-- Name: genre genre_pkey; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.genre
    ADD CONSTRAINT genre_pkey PRIMARY KEY (name);


--
-- Name: genre_to_artist genre_to_artist_pkey; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.genre_to_artist
    ADD CONSTRAINT genre_to_artist_pkey PRIMARY KEY (id);


--
-- Name: push_subscription push_subscription_pkey; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.push_subscription
    ADD CONSTRAINT push_subscription_pkey PRIMARY KEY (id);


--
-- Name: region region_pkey; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.region
    ADD CONSTRAINT region_pkey PRIMARY KEY (name);


--
-- Name: user_authentications uniq_user_authentications; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.user_authentications
    ADD CONSTRAINT uniq_user_authentications UNIQUE (service, identifier);


--
-- Name: user_authentications user_authentications_pkey; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.user_authentications
    ADD CONSTRAINT user_authentications_pkey PRIMARY KEY (id);


--
-- Name: user_emails user_emails_pkey; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.user_emails
    ADD CONSTRAINT user_emails_pkey PRIMARY KEY (id);


--
-- Name: user_emails user_emails_user_id_email_key; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.user_emails
    ADD CONSTRAINT user_emails_user_id_email_key UNIQUE (user_id, email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: venue venue_pkey; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.venue
    ADD CONSTRAINT venue_pkey PRIMARY KEY (name);


--
-- Name: watch_list watch_list_pkey; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.watch_list
    ADD CONSTRAINT watch_list_pkey PRIMARY KEY (id);


--
-- Name: watched_to_account watched_to_account_pkey; Type: CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.watched_to_account
    ADD CONSTRAINT watched_to_account_pkey PRIMARY KEY (id);


--
-- Name: connect_pg_simple_sessions session_pkey; Type: CONSTRAINT; Schema: edm_private; Owner: -
--

ALTER TABLE ONLY edm_private.connect_pg_simple_sessions
    ADD CONSTRAINT session_pkey PRIMARY KEY (sid);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: edm_private; Owner: -
--

ALTER TABLE ONLY edm_private.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (uuid);


--
-- Name: unregistered_email_password_resets unregistered_email_pkey; Type: CONSTRAINT; Schema: edm_private; Owner: -
--

ALTER TABLE ONLY edm_private.unregistered_email_password_resets
    ADD CONSTRAINT unregistered_email_pkey PRIMARY KEY (email);


--
-- Name: user_authentication_secrets user_authentication_secrets_pkey; Type: CONSTRAINT; Schema: edm_private; Owner: -
--

ALTER TABLE ONLY edm_private.user_authentication_secrets
    ADD CONSTRAINT user_authentication_secrets_pkey PRIMARY KEY (user_authentication_id);


--
-- Name: user_email_secrets user_email_secrets_pkey; Type: CONSTRAINT; Schema: edm_private; Owner: -
--

ALTER TABLE ONLY edm_private.user_email_secrets
    ADD CONSTRAINT user_email_secrets_pkey PRIMARY KEY (user_email_id);


--
-- Name: user_secrets user_secrets_pkey; Type: CONSTRAINT; Schema: edm_private; Owner: -
--

ALTER TABLE ONLY edm_private.user_secrets
    ADD CONSTRAINT user_secrets_pkey PRIMARY KEY (user_id);


--
-- Name: migrations migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.migrations
    ADD CONSTRAINT migrations_pkey PRIMARY KEY (id);


--
-- Name: idx_artisttoevent_artist; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_artisttoevent_artist ON edm.artist_to_event USING btree (artist_id);


--
-- Name: idx_artisttoevent_event; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_artisttoevent_event ON edm.artist_to_event USING btree (event_id);


--
-- Name: idx_city_country; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_city_country ON edm.city USING btree (country);


--
-- Name: idx_city_region; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_city_region ON edm.city USING btree (region);


--
-- Name: idx_event_city; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_event_city ON edm.event USING btree (city);


--
-- Name: idx_event_contributor; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_event_contributor ON edm.event USING btree (contributor);


--
-- Name: idx_event_country; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_event_country ON edm.event USING btree (country);


--
-- Name: idx_event_created; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_event_created ON edm.event USING btree (created_at);


--
-- Name: idx_event_name; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_event_name ON edm.event USING btree (name);


--
-- Name: idx_event_region; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_event_region ON edm.event USING btree (region);


--
-- Name: idx_event_start; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_event_start ON edm.event USING btree (start_date);


--
-- Name: idx_event_venue; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_event_venue ON edm.event USING btree (venue);


--
-- Name: idx_followlist_account; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_followlist_account ON edm.follow_list USING btree (user_id);


--
-- Name: idx_followlist_artist; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_followlist_artist ON edm.follow_list USING btree (artist_id);


--
-- Name: idx_followlist_venue; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_followlist_venue ON edm.follow_list USING btree (venue_id);


--
-- Name: idx_genretoartist_artist; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_genretoartist_artist ON edm.genre_to_artist USING btree (artist_id);


--
-- Name: idx_genretoartist_genre; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_genretoartist_genre ON edm.genre_to_artist USING btree (genre_id);


--
-- Name: idx_pushsubscription_account; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_pushsubscription_account ON edm.push_subscription USING btree (user_id);


--
-- Name: idx_region_country; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_region_country ON edm.region USING btree (country);


--
-- Name: idx_user_emails_primary; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_user_emails_primary ON edm.user_emails USING btree (is_primary, user_id);


--
-- Name: idx_venue_city; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_venue_city ON edm.venue USING btree (city);


--
-- Name: idx_watchedtoaccount_account; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_watchedtoaccount_account ON edm.watched_to_account USING btree (user_id);


--
-- Name: idx_watchedtoaccount_city; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_watchedtoaccount_city ON edm.watched_to_account USING btree (city_id);


--
-- Name: idx_watchedtoaccount_region; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_watchedtoaccount_region ON edm.watched_to_account USING btree (region);


--
-- Name: idx_watchlist_account; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_watchlist_account ON edm.watch_list USING btree (user_id);


--
-- Name: idx_watchlist_event; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX idx_watchlist_event ON edm.watch_list USING btree (event_id);


--
-- Name: uniq_user_emails_primary_email; Type: INDEX; Schema: edm; Owner: -
--

CREATE UNIQUE INDEX uniq_user_emails_primary_email ON edm.user_emails USING btree (user_id) WHERE (is_primary IS TRUE);


--
-- Name: uniq_user_emails_verified_email; Type: INDEX; Schema: edm; Owner: -
--

CREATE UNIQUE INDEX uniq_user_emails_verified_email ON edm.user_emails USING btree (email) WHERE (is_verified IS TRUE);


--
-- Name: user_authentications_user_id_idx; Type: INDEX; Schema: edm; Owner: -
--

CREATE INDEX user_authentications_user_id_idx ON edm.user_authentications USING btree (user_id);


--
-- Name: users _100_timestamps; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _100_timestamps BEFORE INSERT OR UPDATE ON edm.users FOR EACH ROW EXECUTE PROCEDURE edm_private.tg__timestamps();


--
-- Name: user_emails _100_timestamps; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _100_timestamps BEFORE INSERT OR UPDATE ON edm.user_emails FOR EACH ROW EXECUTE PROCEDURE edm_private.tg__timestamps();


--
-- Name: user_authentications _100_timestamps; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _100_timestamps BEFORE INSERT OR UPDATE ON edm.user_authentications FOR EACH ROW EXECUTE PROCEDURE edm_private.tg__timestamps();


--
-- Name: push_subscription _100_timestamps; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _100_timestamps BEFORE INSERT OR UPDATE ON edm.push_subscription FOR EACH ROW EXECUTE PROCEDURE edm_private.tg__timestamps();


--
-- Name: country _100_timestamps; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _100_timestamps BEFORE INSERT OR UPDATE ON edm.country FOR EACH ROW EXECUTE PROCEDURE edm_private.tg__timestamps();


--
-- Name: region _100_timestamps; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _100_timestamps BEFORE INSERT OR UPDATE ON edm.region FOR EACH ROW EXECUTE PROCEDURE edm_private.tg__timestamps();


--
-- Name: city _100_timestamps; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _100_timestamps BEFORE INSERT OR UPDATE ON edm.city FOR EACH ROW EXECUTE PROCEDURE edm_private.tg__timestamps();


--
-- Name: venue _100_timestamps; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _100_timestamps BEFORE INSERT OR UPDATE ON edm.venue FOR EACH ROW EXECUTE PROCEDURE edm_private.tg__timestamps();


--
-- Name: genre _100_timestamps; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _100_timestamps BEFORE INSERT OR UPDATE ON edm.genre FOR EACH ROW EXECUTE PROCEDURE edm_private.tg__timestamps();


--
-- Name: artist _100_timestamps; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _100_timestamps BEFORE INSERT OR UPDATE ON edm.artist FOR EACH ROW EXECUTE PROCEDURE edm_private.tg__timestamps();


--
-- Name: event _100_timestamps; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _100_timestamps BEFORE INSERT OR UPDATE ON edm.event FOR EACH ROW EXECUTE PROCEDURE edm_private.tg__timestamps();


--
-- Name: user_emails _200_forbid_existing_email; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _200_forbid_existing_email BEFORE INSERT ON edm.user_emails FOR EACH ROW EXECUTE PROCEDURE edm.tg_user_emails__forbid_if_verified();


--
-- Name: users _200_make_first_user_admin; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _200_make_first_user_admin BEFORE INSERT ON edm.users FOR EACH ROW WHEN ((new.id = 1)) EXECUTE PROCEDURE edm_private.tg_users__make_first_user_admin();


--
-- Name: users _500_gql_update; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _500_gql_update AFTER UPDATE ON edm.users FOR EACH ROW EXECUTE PROCEDURE edm.tg__graphql_subscription('userChanged', 'graphql:user:$1', 'id');


--
-- Name: users _500_insert_secrets; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _500_insert_secrets AFTER INSERT ON edm.users FOR EACH ROW EXECUTE PROCEDURE edm_private.tg_user_secrets__insert_with_user();


--
-- Name: user_emails _500_insert_secrets; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _500_insert_secrets AFTER INSERT ON edm.user_emails FOR EACH ROW EXECUTE PROCEDURE edm_private.tg_user_email_secrets__insert_with_user_email();


--
-- Name: user_emails _500_verify_account_on_verified; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _500_verify_account_on_verified AFTER INSERT OR UPDATE OF is_verified ON edm.user_emails FOR EACH ROW WHEN ((new.is_verified IS TRUE)) EXECUTE PROCEDURE edm.tg_user_emails__verify_account_on_verified();


--
-- Name: user_emails _900_send_verification_email; Type: TRIGGER; Schema: edm; Owner: -
--

CREATE TRIGGER _900_send_verification_email AFTER INSERT ON edm.user_emails FOR EACH ROW WHEN ((new.is_verified IS FALSE)) EXECUTE PROCEDURE edm_private.tg__add_job('user_emails__send_verification');


--
-- Name: artist_to_event artist_to_event_artist_id_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.artist_to_event
    ADD CONSTRAINT artist_to_event_artist_id_fkey FOREIGN KEY (artist_id) REFERENCES edm.artist(name) ON DELETE CASCADE;


--
-- Name: artist_to_event artist_to_event_event_id_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.artist_to_event
    ADD CONSTRAINT artist_to_event_event_id_fkey FOREIGN KEY (event_id) REFERENCES edm.event(id) ON DELETE CASCADE;


--
-- Name: city city_country_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.city
    ADD CONSTRAINT city_country_fkey FOREIGN KEY (country) REFERENCES edm.country(code) ON DELETE CASCADE;


--
-- Name: city city_region_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.city
    ADD CONSTRAINT city_region_fkey FOREIGN KEY (region) REFERENCES edm.region(name) ON DELETE CASCADE;


--
-- Name: event event_city_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.event
    ADD CONSTRAINT event_city_fkey FOREIGN KEY (city) REFERENCES edm.city(id) ON DELETE CASCADE;


--
-- Name: event event_contributor_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.event
    ADD CONSTRAINT event_contributor_fkey FOREIGN KEY (contributor) REFERENCES edm.users(id) ON DELETE CASCADE;


--
-- Name: event event_country_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.event
    ADD CONSTRAINT event_country_fkey FOREIGN KEY (country) REFERENCES edm.country(code) ON DELETE CASCADE;


--
-- Name: event event_region_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.event
    ADD CONSTRAINT event_region_fkey FOREIGN KEY (region) REFERENCES edm.region(name) ON DELETE CASCADE;


--
-- Name: event event_venue_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.event
    ADD CONSTRAINT event_venue_fkey FOREIGN KEY (venue) REFERENCES edm.venue(name) ON DELETE CASCADE;


--
-- Name: follow_list follow_list_artist_id_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.follow_list
    ADD CONSTRAINT follow_list_artist_id_fkey FOREIGN KEY (artist_id) REFERENCES edm.artist(name) ON DELETE CASCADE;


--
-- Name: follow_list follow_list_user_id_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.follow_list
    ADD CONSTRAINT follow_list_user_id_fkey FOREIGN KEY (user_id) REFERENCES edm.users(id) ON DELETE CASCADE;


--
-- Name: follow_list follow_list_venue_id_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.follow_list
    ADD CONSTRAINT follow_list_venue_id_fkey FOREIGN KEY (venue_id) REFERENCES edm.venue(name) ON DELETE CASCADE;


--
-- Name: genre_to_artist genre_to_artist_artist_id_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.genre_to_artist
    ADD CONSTRAINT genre_to_artist_artist_id_fkey FOREIGN KEY (artist_id) REFERENCES edm.artist(name) ON DELETE CASCADE;


--
-- Name: genre_to_artist genre_to_artist_genre_id_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.genre_to_artist
    ADD CONSTRAINT genre_to_artist_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES edm.genre(name) ON DELETE CASCADE;


--
-- Name: push_subscription push_subscription_user_id_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.push_subscription
    ADD CONSTRAINT push_subscription_user_id_fkey FOREIGN KEY (user_id) REFERENCES edm.users(id) ON DELETE CASCADE;


--
-- Name: region region_country_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.region
    ADD CONSTRAINT region_country_fkey FOREIGN KEY (country) REFERENCES edm.country(code) ON DELETE CASCADE;


--
-- Name: user_authentications user_authentications_user_id_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.user_authentications
    ADD CONSTRAINT user_authentications_user_id_fkey FOREIGN KEY (user_id) REFERENCES edm.users(id) ON DELETE CASCADE;


--
-- Name: user_emails user_emails_user_id_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.user_emails
    ADD CONSTRAINT user_emails_user_id_fkey FOREIGN KEY (user_id) REFERENCES edm.users(id) ON DELETE CASCADE;


--
-- Name: venue venue_city_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.venue
    ADD CONSTRAINT venue_city_fkey FOREIGN KEY (city) REFERENCES edm.city(id) ON DELETE CASCADE;


--
-- Name: watch_list watch_list_event_id_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.watch_list
    ADD CONSTRAINT watch_list_event_id_fkey FOREIGN KEY (event_id) REFERENCES edm.event(id) ON DELETE CASCADE;


--
-- Name: watch_list watch_list_user_id_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.watch_list
    ADD CONSTRAINT watch_list_user_id_fkey FOREIGN KEY (user_id) REFERENCES edm.users(id) ON DELETE CASCADE;


--
-- Name: watched_to_account watched_to_account_city_id_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.watched_to_account
    ADD CONSTRAINT watched_to_account_city_id_fkey FOREIGN KEY (city_id) REFERENCES edm.city(id) ON DELETE CASCADE;


--
-- Name: watched_to_account watched_to_account_region_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.watched_to_account
    ADD CONSTRAINT watched_to_account_region_fkey FOREIGN KEY (region) REFERENCES edm.region(name) ON DELETE CASCADE;


--
-- Name: watched_to_account watched_to_account_user_id_fkey; Type: FK CONSTRAINT; Schema: edm; Owner: -
--

ALTER TABLE ONLY edm.watched_to_account
    ADD CONSTRAINT watched_to_account_user_id_fkey FOREIGN KEY (user_id) REFERENCES edm.users(id) ON DELETE CASCADE;


--
-- Name: sessions sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: edm_private; Owner: -
--

ALTER TABLE ONLY edm_private.sessions
    ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES edm.users(id) ON DELETE CASCADE;


--
-- Name: user_authentication_secrets user_authentication_secrets_user_authentication_id_fkey; Type: FK CONSTRAINT; Schema: edm_private; Owner: -
--

ALTER TABLE ONLY edm_private.user_authentication_secrets
    ADD CONSTRAINT user_authentication_secrets_user_authentication_id_fkey FOREIGN KEY (user_authentication_id) REFERENCES edm.user_authentications(id) ON DELETE CASCADE;


--
-- Name: user_email_secrets user_email_secrets_user_email_id_fkey; Type: FK CONSTRAINT; Schema: edm_private; Owner: -
--

ALTER TABLE ONLY edm_private.user_email_secrets
    ADD CONSTRAINT user_email_secrets_user_email_id_fkey FOREIGN KEY (user_email_id) REFERENCES edm.user_emails(id) ON DELETE CASCADE;


--
-- Name: user_secrets user_secrets_user_id_fkey; Type: FK CONSTRAINT; Schema: edm_private; Owner: -
--

ALTER TABLE ONLY edm_private.user_secrets
    ADD CONSTRAINT user_secrets_user_id_fkey FOREIGN KEY (user_id) REFERENCES edm.users(id) ON DELETE CASCADE;


--
-- Name: artist; Type: ROW SECURITY; Schema: edm; Owner: -
--

ALTER TABLE edm.artist ENABLE ROW LEVEL SECURITY;

--
-- Name: city; Type: ROW SECURITY; Schema: edm; Owner: -
--

ALTER TABLE edm.city ENABLE ROW LEVEL SECURITY;

--
-- Name: country; Type: ROW SECURITY; Schema: edm; Owner: -
--

ALTER TABLE edm.country ENABLE ROW LEVEL SECURITY;

--
-- Name: follow_list delete_follow_list; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY delete_follow_list ON edm.follow_list FOR DELETE USING ((user_id = edm.current_user_id()));


--
-- Name: user_emails delete_own; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY delete_own ON edm.user_emails FOR DELETE USING ((user_id = edm.current_user_id()));


--
-- Name: user_authentications delete_own; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY delete_own ON edm.user_authentications FOR DELETE USING ((user_id = edm.current_user_id()));


--
-- Name: push_subscription delete_push_subscription; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY delete_push_subscription ON edm.push_subscription FOR DELETE USING ((user_id = edm.current_user_id()));


--
-- Name: users delete_self; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY delete_self ON edm.users FOR DELETE USING ((id = edm.current_user_id()));


--
-- Name: watch_list delete_watch_list; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY delete_watch_list ON edm.watch_list FOR DELETE USING ((user_id = edm.current_user_id()));


--
-- Name: watched_to_account delete_watched_to_account; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY delete_watched_to_account ON edm.watched_to_account FOR DELETE USING ((user_id = edm.current_user_id()));


--
-- Name: event; Type: ROW SECURITY; Schema: edm; Owner: -
--

ALTER TABLE edm.event ENABLE ROW LEVEL SECURITY;

--
-- Name: follow_list; Type: ROW SECURITY; Schema: edm; Owner: -
--

ALTER TABLE edm.follow_list ENABLE ROW LEVEL SECURITY;

--
-- Name: genre; Type: ROW SECURITY; Schema: edm; Owner: -
--

ALTER TABLE edm.genre ENABLE ROW LEVEL SECURITY;

--
-- Name: artist insert_artist; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY insert_artist ON edm.artist FOR INSERT WITH CHECK (true);


--
-- Name: city insert_city; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY insert_city ON edm.city FOR INSERT WITH CHECK (true);


--
-- Name: country insert_country; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY insert_country ON edm.country FOR INSERT WITH CHECK (true);


--
-- Name: event insert_event; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY insert_event ON edm.event FOR SELECT USING (true);


--
-- Name: follow_list insert_follow_list; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY insert_follow_list ON edm.follow_list FOR INSERT WITH CHECK ((user_id = edm.current_user_id()));


--
-- Name: genre insert_genre; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY insert_genre ON edm.genre FOR INSERT WITH CHECK (true);


--
-- Name: user_emails insert_own; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY insert_own ON edm.user_emails FOR INSERT WITH CHECK ((user_id = edm.current_user_id()));


--
-- Name: push_subscription insert_push_subscription; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY insert_push_subscription ON edm.push_subscription FOR INSERT WITH CHECK ((user_id = edm.current_user_id()));


--
-- Name: region insert_region; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY insert_region ON edm.region FOR INSERT WITH CHECK (true);


--
-- Name: venue insert_venue; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY insert_venue ON edm.venue FOR INSERT WITH CHECK (true);


--
-- Name: watch_list insert_watch_list; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY insert_watch_list ON edm.watch_list FOR INSERT WITH CHECK ((user_id = edm.current_user_id()));


--
-- Name: watched_to_account insert_watched_to_account; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY insert_watched_to_account ON edm.watched_to_account FOR INSERT WITH CHECK ((user_id = edm.current_user_id()));


--
-- Name: push_subscription; Type: ROW SECURITY; Schema: edm; Owner: -
--

ALTER TABLE edm.push_subscription ENABLE ROW LEVEL SECURITY;

--
-- Name: region; Type: ROW SECURITY; Schema: edm; Owner: -
--

ALTER TABLE edm.region ENABLE ROW LEVEL SECURITY;

--
-- Name: users select_all; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY select_all ON edm.users FOR SELECT USING (true);


--
-- Name: artist select_artist; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY select_artist ON edm.artist FOR SELECT USING (true);


--
-- Name: city select_city; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY select_city ON edm.city FOR SELECT USING (true);


--
-- Name: country select_country; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY select_country ON edm.country FOR SELECT USING (true);


--
-- Name: event select_event; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY select_event ON edm.event FOR SELECT USING (true);


--
-- Name: follow_list select_follow_list; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY select_follow_list ON edm.follow_list FOR SELECT USING (true);


--
-- Name: genre select_genre; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY select_genre ON edm.genre FOR SELECT USING (true);


--
-- Name: user_emails select_own; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY select_own ON edm.user_emails FOR SELECT USING ((user_id = edm.current_user_id()));


--
-- Name: user_authentications select_own; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY select_own ON edm.user_authentications FOR SELECT USING ((user_id = edm.current_user_id()));


--
-- Name: push_subscription select_push_subscription; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY select_push_subscription ON edm.push_subscription FOR SELECT USING (true);


--
-- Name: region select_region; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY select_region ON edm.region FOR SELECT USING (true);


--
-- Name: venue select_venue; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY select_venue ON edm.venue FOR SELECT USING (true);


--
-- Name: watch_list select_watch_list; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY select_watch_list ON edm.watch_list FOR SELECT USING (true);


--
-- Name: watched_to_account select_watched_to_account; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY select_watched_to_account ON edm.watched_to_account FOR SELECT USING (true);


--
-- Name: event update_event; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY update_event ON edm.event FOR UPDATE USING ((contributor = edm.current_user_id()));


--
-- Name: follow_list update_follow_list; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY update_follow_list ON edm.follow_list FOR UPDATE USING ((user_id = edm.current_user_id()));


--
-- Name: push_subscription update_push_subscription; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY update_push_subscription ON edm.push_subscription FOR UPDATE USING ((user_id = edm.current_user_id()));


--
-- Name: users update_self; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY update_self ON edm.users FOR UPDATE USING ((id = edm.current_user_id()));


--
-- Name: watch_list update_watch_list; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY update_watch_list ON edm.watch_list FOR UPDATE USING ((user_id = edm.current_user_id()));


--
-- Name: watched_to_account update_watched_to_account; Type: POLICY; Schema: edm; Owner: -
--

CREATE POLICY update_watched_to_account ON edm.watched_to_account FOR UPDATE USING ((user_id = edm.current_user_id()));


--
-- Name: user_authentications; Type: ROW SECURITY; Schema: edm; Owner: -
--

ALTER TABLE edm.user_authentications ENABLE ROW LEVEL SECURITY;

--
-- Name: user_emails; Type: ROW SECURITY; Schema: edm; Owner: -
--

ALTER TABLE edm.user_emails ENABLE ROW LEVEL SECURITY;

--
-- Name: users; Type: ROW SECURITY; Schema: edm; Owner: -
--

ALTER TABLE edm.users ENABLE ROW LEVEL SECURITY;

--
-- Name: venue; Type: ROW SECURITY; Schema: edm; Owner: -
--

ALTER TABLE edm.venue ENABLE ROW LEVEL SECURITY;

--
-- Name: watch_list; Type: ROW SECURITY; Schema: edm; Owner: -
--

ALTER TABLE edm.watch_list ENABLE ROW LEVEL SECURITY;

--
-- Name: watched_to_account; Type: ROW SECURITY; Schema: edm; Owner: -
--

ALTER TABLE edm.watched_to_account ENABLE ROW LEVEL SECURITY;

--
-- Name: connect_pg_simple_sessions; Type: ROW SECURITY; Schema: edm_private; Owner: -
--

ALTER TABLE edm_private.connect_pg_simple_sessions ENABLE ROW LEVEL SECURITY;

--
-- Name: sessions; Type: ROW SECURITY; Schema: edm_private; Owner: -
--

ALTER TABLE edm_private.sessions ENABLE ROW LEVEL SECURITY;

--
-- Name: user_authentication_secrets; Type: ROW SECURITY; Schema: edm_private; Owner: -
--

ALTER TABLE edm_private.user_authentication_secrets ENABLE ROW LEVEL SECURITY;

--
-- Name: user_email_secrets; Type: ROW SECURITY; Schema: edm_private; Owner: -
--

ALTER TABLE edm_private.user_email_secrets ENABLE ROW LEVEL SECURITY;

--
-- Name: user_secrets; Type: ROW SECURITY; Schema: edm_private; Owner: -
--

ALTER TABLE edm_private.user_secrets ENABLE ROW LEVEL SECURITY;

--
-- Name: SCHEMA edm; Type: ACL; Schema: -; Owner: -
--

GRANT USAGE ON SCHEMA edm TO edm_visitor;


--
-- Name: SCHEMA edm_hidden; Type: ACL; Schema: -; Owner: -
--

GRANT USAGE ON SCHEMA edm_hidden TO edm_visitor;


--
-- Name: FUNCTION change_password(old_password text, new_password text); Type: ACL; Schema: edm; Owner: -
--

GRANT ALL ON FUNCTION edm.change_password(old_password text, new_password text) TO edm_visitor;


--
-- Name: TABLE users; Type: ACL; Schema: edm; Owner: -
--

GRANT SELECT ON TABLE edm.users TO edm_visitor;


--
-- Name: COLUMN users.username; Type: ACL; Schema: edm; Owner: -
--

GRANT UPDATE(username) ON TABLE edm.users TO edm_visitor;


--
-- Name: COLUMN users.name; Type: ACL; Schema: edm; Owner: -
--

GRANT UPDATE(name) ON TABLE edm.users TO edm_visitor;


--
-- Name: COLUMN users.profile_photo; Type: ACL; Schema: edm; Owner: -
--

GRANT UPDATE(profile_photo) ON TABLE edm.users TO edm_visitor;


--
-- Name: COLUMN users.notification_frequency; Type: ACL; Schema: edm; Owner: -
--

GRANT UPDATE(notification_frequency) ON TABLE edm.users TO edm_visitor;


--
-- Name: COLUMN users.push_notification; Type: ACL; Schema: edm; Owner: -
--

GRANT UPDATE(push_notification) ON TABLE edm.users TO edm_visitor;


--
-- Name: COLUMN users.email_notification; Type: ACL; Schema: edm; Owner: -
--

GRANT UPDATE(email_notification) ON TABLE edm.users TO edm_visitor;


--
-- Name: COLUMN users.is_setup; Type: ACL; Schema: edm; Owner: -
--

GRANT UPDATE(is_setup) ON TABLE edm.users TO edm_visitor;


--
-- Name: TABLE user_emails; Type: ACL; Schema: edm; Owner: -
--

GRANT SELECT,DELETE ON TABLE edm.user_emails TO edm_visitor;


--
-- Name: COLUMN user_emails.email; Type: ACL; Schema: edm; Owner: -
--

GRANT INSERT(email) ON TABLE edm.user_emails TO edm_visitor;


--
-- Name: TABLE event; Type: ACL; Schema: edm; Owner: -
--

GRANT ALL ON TABLE edm.event TO edm_visitor;


--
-- Name: TABLE artist; Type: ACL; Schema: edm; Owner: -
--

GRANT ALL ON TABLE edm.artist TO edm_visitor;


--
-- Name: TABLE artist_to_event; Type: ACL; Schema: edm; Owner: -
--

GRANT ALL ON TABLE edm.artist_to_event TO edm_visitor;


--
-- Name: SEQUENCE artist_to_event_id_seq; Type: ACL; Schema: edm; Owner: -
--

GRANT SELECT,USAGE ON SEQUENCE edm.artist_to_event_id_seq TO edm_visitor;


--
-- Name: TABLE city; Type: ACL; Schema: edm; Owner: -
--

GRANT ALL ON TABLE edm.city TO edm_visitor;


--
-- Name: SEQUENCE city_sequence; Type: ACL; Schema: edm; Owner: -
--

GRANT SELECT,USAGE ON SEQUENCE edm.city_sequence TO edm_visitor;


--
-- Name: TABLE country; Type: ACL; Schema: edm; Owner: -
--

GRANT ALL ON TABLE edm.country TO edm_visitor;


--
-- Name: TABLE follow_list; Type: ACL; Schema: edm; Owner: -
--

GRANT ALL ON TABLE edm.follow_list TO edm_visitor;


--
-- Name: SEQUENCE follow_list_id_seq; Type: ACL; Schema: edm; Owner: -
--

GRANT SELECT,USAGE ON SEQUENCE edm.follow_list_id_seq TO edm_visitor;


--
-- Name: TABLE genre; Type: ACL; Schema: edm; Owner: -
--

GRANT ALL ON TABLE edm.genre TO edm_visitor;


--
-- Name: TABLE genre_to_artist; Type: ACL; Schema: edm; Owner: -
--

GRANT ALL ON TABLE edm.genre_to_artist TO edm_visitor;


--
-- Name: SEQUENCE genre_to_artist_id_seq; Type: ACL; Schema: edm; Owner: -
--

GRANT SELECT,USAGE ON SEQUENCE edm.genre_to_artist_id_seq TO edm_visitor;


--
-- Name: TABLE push_subscription; Type: ACL; Schema: edm; Owner: -
--

GRANT ALL ON TABLE edm.push_subscription TO edm_visitor;


--
-- Name: SEQUENCE push_subscription_id_seq; Type: ACL; Schema: edm; Owner: -
--

GRANT SELECT,USAGE ON SEQUENCE edm.push_subscription_id_seq TO edm_visitor;


--
-- Name: TABLE region; Type: ACL; Schema: edm; Owner: -
--

GRANT ALL ON TABLE edm.region TO edm_visitor;


--
-- Name: TABLE user_authentications; Type: ACL; Schema: edm; Owner: -
--

GRANT SELECT,DELETE ON TABLE edm.user_authentications TO edm_visitor;


--
-- Name: SEQUENCE user_authentications_id_seq; Type: ACL; Schema: edm; Owner: -
--

GRANT SELECT,USAGE ON SEQUENCE edm.user_authentications_id_seq TO edm_visitor;


--
-- Name: SEQUENCE user_emails_id_seq; Type: ACL; Schema: edm; Owner: -
--

GRANT SELECT,USAGE ON SEQUENCE edm.user_emails_id_seq TO edm_visitor;


--
-- Name: SEQUENCE users_id_seq; Type: ACL; Schema: edm; Owner: -
--

GRANT SELECT,USAGE ON SEQUENCE edm.users_id_seq TO edm_visitor;


--
-- Name: TABLE venue; Type: ACL; Schema: edm; Owner: -
--

GRANT ALL ON TABLE edm.venue TO edm_visitor;


--
-- Name: TABLE watch_list; Type: ACL; Schema: edm; Owner: -
--

GRANT ALL ON TABLE edm.watch_list TO edm_visitor;


--
-- Name: SEQUENCE watch_list_id_seq; Type: ACL; Schema: edm; Owner: -
--

GRANT SELECT,USAGE ON SEQUENCE edm.watch_list_id_seq TO edm_visitor;


--
-- Name: TABLE watched_to_account; Type: ACL; Schema: edm; Owner: -
--

GRANT ALL ON TABLE edm.watched_to_account TO edm_visitor;


--
-- Name: SEQUENCE watched_to_account_id_seq; Type: ACL; Schema: edm; Owner: -
--

GRANT SELECT,USAGE ON SEQUENCE edm.watched_to_account_id_seq TO edm_visitor;


--
-- Name: DEFAULT PRIVILEGES FOR SEQUENCES; Type: DEFAULT ACL; Schema: edm; Owner: -
--

ALTER DEFAULT PRIVILEGES FOR ROLE edm IN SCHEMA edm REVOKE ALL ON SEQUENCES  FROM edm;
ALTER DEFAULT PRIVILEGES FOR ROLE edm IN SCHEMA edm GRANT SELECT,USAGE ON SEQUENCES  TO edm_visitor;


--
-- Name: DEFAULT PRIVILEGES FOR SEQUENCES; Type: DEFAULT ACL; Schema: edm_hidden; Owner: -
--

ALTER DEFAULT PRIVILEGES FOR ROLE edm IN SCHEMA edm_hidden REVOKE ALL ON SEQUENCES  FROM edm;
ALTER DEFAULT PRIVILEGES FOR ROLE edm IN SCHEMA edm_hidden GRANT SELECT,USAGE ON SEQUENCES  TO edm_visitor;


--
-- PostgreSQL database dump complete
--

