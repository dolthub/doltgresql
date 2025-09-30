-- Downloaded from: https://github.com/Clar17y/Football-Events/blob/9783f281e6e568f8c3fa7cbdfa8995229870e1a7/schema.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 16.8 (Debian 16.8-1.pgdg120+1)
-- Dumped by pg_dump version 17.4

-- Started on 2025-07-16 10:07:12

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
-- TOC entry 6 (class 2615 OID 24591)
-- Name: grassroots; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA grassroots;


ALTER SCHEMA grassroots OWNER TO postgres;

--
-- TOC entry 853 (class 1247 OID 24604)
-- Name: event_kind; Type: TYPE; Schema: grassroots; Owner: postgres
--

CREATE TYPE grassroots.event_kind AS ENUM (
    'goal',
    'assist',
    'key_pass',
    'save',
    'interception',
    'tackle',
    'foul',
    'penalty',
    'free_kick',
    'ball_out',
    'own_goal'
);


ALTER TYPE grassroots.event_kind OWNER TO postgres;

--
-- TOC entry 886 (class 1247 OID 25269)
-- Name: position_code; Type: TYPE; Schema: grassroots; Owner: postgres
--

CREATE TYPE grassroots.position_code AS ENUM (
    'GK',
    'CB',
    'RCB',
    'LCB',
    'SW',
    'RB',
    'LB',
    'RWB',
    'LWB',
    'CDM',
    'RDM',
    'LDM',
    'CM',
    'RCM',
    'LCM',
    'CAM',
    'RAM',
    'LAM',
    'RM',
    'LM',
    'RW',
    'LW',
    'RF',
    'LF',
    'CF',
    'ST',
    'SS',
    'AM',
    'DM',
    'WM',
    'WB',
    'FB',
    'SUB',
    'BENCH'
);


ALTER TYPE grassroots.position_code OWNER TO postgres;

--
-- TOC entry 880 (class 1247 OID 25247)
-- Name: user_role; Type: TYPE; Schema: grassroots; Owner: postgres
--

CREATE TYPE grassroots.user_role AS ENUM (
    'ADMIN',
    'USER'
);


ALTER TYPE grassroots.user_role OWNER TO postgres;

--
-- TOC entry 226 (class 1255 OID 25266)
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: grassroots; Owner: postgres
--

CREATE FUNCTION grassroots.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
      BEGIN
          NEW.updated_at = CURRENT_TIMESTAMP;
          RETURN NEW;
      END;
      $$;


ALTER FUNCTION grassroots.update_updated_at_column() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 221 (class 1259 OID 24668)
-- Name: awards; Type: TABLE; Schema: grassroots; Owner: postgres
--

CREATE TABLE grassroots.awards (
    award_id uuid DEFAULT gen_random_uuid() NOT NULL,
    season_id uuid NOT NULL,
    player_id uuid NOT NULL,
    category text NOT NULL,
    notes text,
    created_at timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(6) with time zone,
    created_by_user_id uuid NOT NULL,
    deleted_at timestamp(3) without time zone,
    deleted_by_user_id uuid,
    is_deleted boolean DEFAULT false NOT NULL
);


ALTER TABLE grassroots.awards OWNER TO postgres;

--
-- TOC entry 220 (class 1259 OID 24658)
-- Name: events; Type: TABLE; Schema: grassroots; Owner: postgres
--

CREATE TABLE grassroots.events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    match_id uuid NOT NULL,
    season_id uuid NOT NULL,
    created_at timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    period_number integer,
    clock_ms integer,
    kind grassroots.event_kind NOT NULL,
    team_id uuid,
    player_id uuid,
    notes text,
    sentiment integer DEFAULT 0 NOT NULL,
    updated_at timestamp(6) with time zone,
    created_by_user_id uuid NOT NULL,
    deleted_at timestamp(3) without time zone,
    deleted_by_user_id uuid,
    is_deleted boolean DEFAULT false NOT NULL
);


ALTER TABLE grassroots.events OWNER TO postgres;

--
-- TOC entry 222 (class 1259 OID 24691)
-- Name: lineup; Type: TABLE; Schema: grassroots; Owner: postgres
--

CREATE TABLE grassroots.lineup (
    match_id uuid NOT NULL,
    player_id uuid NOT NULL,
    start_min double precision DEFAULT 0 NOT NULL,
    end_min double precision,
    created_at timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(6) with time zone,
    created_by_user_id uuid NOT NULL,
    deleted_at timestamp(3) without time zone,
    deleted_by_user_id uuid,
    is_deleted boolean DEFAULT false NOT NULL,
    "position" grassroots.position_code NOT NULL
);


ALTER TABLE grassroots.lineup OWNER TO postgres;

--
-- TOC entry 223 (class 1259 OID 24700)
-- Name: match_awards; Type: TABLE; Schema: grassroots; Owner: postgres
--

CREATE TABLE grassroots.match_awards (
    match_award_id uuid DEFAULT gen_random_uuid() NOT NULL,
    match_id uuid NOT NULL,
    player_id uuid NOT NULL,
    category text NOT NULL,
    notes text,
    created_at timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(6) with time zone
);


ALTER TABLE grassroots.match_awards OWNER TO postgres;

--
-- TOC entry 219 (class 1259 OID 24645)
-- Name: matches; Type: TABLE; Schema: grassroots; Owner: postgres
--

CREATE TABLE grassroots.matches (
    match_id uuid DEFAULT gen_random_uuid() NOT NULL,
    season_id uuid NOT NULL,
    kickoff_ts timestamp(6) with time zone NOT NULL,
    competition text,
    home_team_id uuid NOT NULL,
    away_team_id uuid NOT NULL,
    venue text,
    duration_mins integer DEFAULT 50 NOT NULL,
    period_format text DEFAULT 'quarter'::text NOT NULL,
    our_score integer DEFAULT 0 NOT NULL,
    opponent_score integer DEFAULT 0 NOT NULL,
    notes text,
    created_at timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(6) with time zone,
    created_by_user_id uuid NOT NULL,
    deleted_at timestamp(3) without time zone,
    deleted_by_user_id uuid,
    is_deleted boolean DEFAULT false NOT NULL
);


ALTER TABLE grassroots.matches OWNER TO postgres;

--
-- TOC entry 218 (class 1259 OID 24636)
-- Name: players; Type: TABLE; Schema: grassroots; Owner: postgres
--

CREATE TABLE grassroots.players (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    full_name text NOT NULL,
    squad_number integer,
    dob date,
    notes text,
    current_team uuid,
    created_at timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(6) with time zone,
    created_by_user_id uuid NOT NULL,
    deleted_at timestamp(3) without time zone,
    deleted_by_user_id uuid,
    is_deleted boolean DEFAULT false NOT NULL,
    preferred_pos grassroots.position_code
);


ALTER TABLE grassroots.players OWNER TO postgres;

--
-- TOC entry 224 (class 1259 OID 24727)
-- Name: seasons; Type: TABLE; Schema: grassroots; Owner: postgres
--

CREATE TABLE grassroots.seasons (
    season_id uuid DEFAULT gen_random_uuid() NOT NULL,
    label text NOT NULL,
    start_date date NOT NULL,
    end_date date NOT NULL,
    is_current boolean DEFAULT false NOT NULL,
    description text,
    created_at timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(6) with time zone,
    created_by_user_id uuid NOT NULL,
    deleted_at timestamp(3) without time zone,
    deleted_by_user_id uuid,
    is_deleted boolean DEFAULT false NOT NULL
);


ALTER TABLE grassroots.seasons OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 24627)
-- Name: teams; Type: TABLE; Schema: grassroots; Owner: postgres
--

CREATE TABLE grassroots.teams (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    home_kit_primary character varying(7),
    home_kit_secondary character varying(7),
    away_kit_primary character varying(7),
    away_kit_secondary character varying(7),
    logo_url text,
    created_at timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(6) with time zone,
    created_by_user_id uuid NOT NULL,
    deleted_at timestamp(3) without time zone,
    deleted_by_user_id uuid,
    is_deleted boolean DEFAULT false NOT NULL
);


ALTER TABLE grassroots.teams OWNER TO postgres;

--
-- TOC entry 225 (class 1259 OID 25251)
-- Name: users; Type: TABLE; Schema: grassroots; Owner: postgres
--

CREATE TABLE grassroots.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    first_name text,
    last_name text,
    role grassroots.user_role DEFAULT 'USER'::grassroots.user_role NOT NULL,
    email_verified boolean DEFAULT false NOT NULL,
    created_at timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(3) without time zone NOT NULL,
    is_deleted boolean DEFAULT false NOT NULL,
    deleted_at timestamp(3) without time zone,
    deleted_by_user_id uuid
);


ALTER TABLE grassroots.users OWNER TO postgres;

--
-- TOC entry 3293 (class 2606 OID 24676)
-- Name: awards awards_pkey; Type: CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.awards
    ADD CONSTRAINT awards_pkey PRIMARY KEY (award_id);


--
-- TOC entry 3291 (class 2606 OID 24667)
-- Name: events events_pkey; Type: CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (id);


--
-- TOC entry 3295 (class 2606 OID 24699)
-- Name: lineup lineup_pkey; Type: CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.lineup
    ADD CONSTRAINT lineup_pkey PRIMARY KEY (match_id, player_id, start_min);


--
-- TOC entry 3298 (class 2606 OID 24708)
-- Name: match_awards match_awards_pkey; Type: CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.match_awards
    ADD CONSTRAINT match_awards_pkey PRIMARY KEY (match_award_id);


--
-- TOC entry 3289 (class 2606 OID 24657)
-- Name: matches matches_pkey; Type: CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.matches
    ADD CONSTRAINT matches_pkey PRIMARY KEY (match_id);


--
-- TOC entry 3287 (class 2606 OID 24644)
-- Name: players players_pkey; Type: CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.players
    ADD CONSTRAINT players_pkey PRIMARY KEY (id);


--
-- TOC entry 3301 (class 2606 OID 24736)
-- Name: seasons seasons_pkey; Type: CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.seasons
    ADD CONSTRAINT seasons_pkey PRIMARY KEY (season_id);


--
-- TOC entry 3284 (class 2606 OID 24635)
-- Name: teams teams_pkey; Type: CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.teams
    ADD CONSTRAINT teams_pkey PRIMARY KEY (id);


--
-- TOC entry 3303 (class 2606 OID 25345)
-- Name: users users_email_key; Type: CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- TOC entry 3305 (class 2606 OID 25263)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- TOC entry 3296 (class 1259 OID 24739)
-- Name: match_awards_match_id_category_key; Type: INDEX; Schema: grassroots; Owner: postgres
--

CREATE UNIQUE INDEX match_awards_match_id_category_key ON grassroots.match_awards USING btree (match_id, category);


--
-- TOC entry 3285 (class 1259 OID 24738)
-- Name: players_fullname_team_unique; Type: INDEX; Schema: grassroots; Owner: postgres
--

CREATE UNIQUE INDEX players_fullname_team_unique ON grassroots.players USING btree (full_name, current_team);


--
-- TOC entry 3299 (class 1259 OID 24740)
-- Name: seasons_label_key; Type: INDEX; Schema: grassroots; Owner: postgres
--

CREATE UNIQUE INDEX seasons_label_key ON grassroots.seasons USING btree (label);


--
-- TOC entry 3282 (class 1259 OID 24737)
-- Name: teams_name_key; Type: INDEX; Schema: grassroots; Owner: postgres
--

CREATE UNIQUE INDEX teams_name_key ON grassroots.teams USING btree (name);


--
-- TOC entry 3332 (class 2620 OID 25267)
-- Name: users update_users_updated_at; Type: TRIGGER; Schema: grassroots; Owner: postgres
--

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON grassroots.users FOR EACH ROW EXECUTE FUNCTION grassroots.update_updated_at_column();


--
-- TOC entry 3320 (class 2606 OID 25410)
-- Name: awards awards_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.awards
    ADD CONSTRAINT awards_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES grassroots.users(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- TOC entry 3321 (class 2606 OID 25415)
-- Name: awards awards_deleted_by_user_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.awards
    ADD CONSTRAINT awards_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES grassroots.users(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 3322 (class 2606 OID 24776)
-- Name: awards awards_player_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.awards
    ADD CONSTRAINT awards_player_id_fkey FOREIGN KEY (player_id) REFERENCES grassroots.players(id) ON DELETE CASCADE;


--
-- TOC entry 3323 (class 2606 OID 24781)
-- Name: awards awards_season_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.awards
    ADD CONSTRAINT awards_season_id_fkey FOREIGN KEY (season_id) REFERENCES grassroots.seasons(season_id) ON DELETE CASCADE;


--
-- TOC entry 3316 (class 2606 OID 25400)
-- Name: events events_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.events
    ADD CONSTRAINT events_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES grassroots.users(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- TOC entry 3317 (class 2606 OID 25405)
-- Name: events events_deleted_by_user_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.events
    ADD CONSTRAINT events_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES grassroots.users(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 3318 (class 2606 OID 24766)
-- Name: events events_match_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.events
    ADD CONSTRAINT events_match_id_fkey FOREIGN KEY (match_id) REFERENCES grassroots.matches(match_id) ON DELETE CASCADE;


--
-- TOC entry 3319 (class 2606 OID 24771)
-- Name: events events_team_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.events
    ADD CONSTRAINT events_team_id_fkey FOREIGN KEY (team_id) REFERENCES grassroots.teams(id);


--
-- TOC entry 3324 (class 2606 OID 25420)
-- Name: lineup lineup_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.lineup
    ADD CONSTRAINT lineup_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES grassroots.users(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- TOC entry 3325 (class 2606 OID 25425)
-- Name: lineup lineup_deleted_by_user_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.lineup
    ADD CONSTRAINT lineup_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES grassroots.users(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 3326 (class 2606 OID 24786)
-- Name: lineup lineup_match_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.lineup
    ADD CONSTRAINT lineup_match_id_fkey FOREIGN KEY (match_id) REFERENCES grassroots.matches(match_id) ON DELETE CASCADE;


--
-- TOC entry 3327 (class 2606 OID 24791)
-- Name: lineup lineup_player_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.lineup
    ADD CONSTRAINT lineup_player_id_fkey FOREIGN KEY (player_id) REFERENCES grassroots.players(id) ON DELETE CASCADE;


--
-- TOC entry 3328 (class 2606 OID 24801)
-- Name: match_awards match_awards_match_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.match_awards
    ADD CONSTRAINT match_awards_match_id_fkey FOREIGN KEY (match_id) REFERENCES grassroots.matches(match_id) ON DELETE CASCADE;


--
-- TOC entry 3329 (class 2606 OID 24806)
-- Name: match_awards match_awards_player_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.match_awards
    ADD CONSTRAINT match_awards_player_id_fkey FOREIGN KEY (player_id) REFERENCES grassroots.players(id) ON DELETE CASCADE;


--
-- TOC entry 3311 (class 2606 OID 24751)
-- Name: matches matches_away_team_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.matches
    ADD CONSTRAINT matches_away_team_id_fkey FOREIGN KEY (away_team_id) REFERENCES grassroots.teams(id) ON DELETE CASCADE;


--
-- TOC entry 3312 (class 2606 OID 25390)
-- Name: matches matches_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.matches
    ADD CONSTRAINT matches_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES grassroots.users(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- TOC entry 3313 (class 2606 OID 25395)
-- Name: matches matches_deleted_by_user_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.matches
    ADD CONSTRAINT matches_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES grassroots.users(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 3314 (class 2606 OID 24756)
-- Name: matches matches_home_team_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.matches
    ADD CONSTRAINT matches_home_team_id_fkey FOREIGN KEY (home_team_id) REFERENCES grassroots.teams(id) ON DELETE CASCADE;


--
-- TOC entry 3315 (class 2606 OID 24761)
-- Name: matches matches_season_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.matches
    ADD CONSTRAINT matches_season_id_fkey FOREIGN KEY (season_id) REFERENCES grassroots.seasons(season_id) ON DELETE CASCADE;


--
-- TOC entry 3308 (class 2606 OID 25380)
-- Name: players players_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.players
    ADD CONSTRAINT players_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES grassroots.users(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- TOC entry 3309 (class 2606 OID 24741)
-- Name: players players_current_team_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.players
    ADD CONSTRAINT players_current_team_fkey FOREIGN KEY (current_team) REFERENCES grassroots.teams(id);


--
-- TOC entry 3310 (class 2606 OID 25385)
-- Name: players players_deleted_by_user_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.players
    ADD CONSTRAINT players_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES grassroots.users(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 3330 (class 2606 OID 25435)
-- Name: seasons seasons_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.seasons
    ADD CONSTRAINT seasons_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES grassroots.users(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- TOC entry 3331 (class 2606 OID 25440)
-- Name: seasons seasons_deleted_by_user_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.seasons
    ADD CONSTRAINT seasons_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES grassroots.users(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 3306 (class 2606 OID 25370)
-- Name: teams teams_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.teams
    ADD CONSTRAINT teams_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES grassroots.users(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- TOC entry 3307 (class 2606 OID 25375)
-- Name: teams teams_deleted_by_user_id_fkey; Type: FK CONSTRAINT; Schema: grassroots; Owner: postgres
--

ALTER TABLE ONLY grassroots.teams
    ADD CONSTRAINT teams_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES grassroots.users(id) ON UPDATE CASCADE ON DELETE SET NULL;


-- Completed on 2025-07-16 10:07:12

--
-- PostgreSQL database dump complete
--

