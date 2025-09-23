-- Downloaded from: https://github.com/Billoxinogen18/ar_backend/blob/6cfc633fc995510afa71f887a5c7d1fd6b568f61/neon_dump.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 17.5
-- Dumped by pg_dump version 17.5 (Homebrew)

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

ALTER TABLE ONLY public.users DROP CONSTRAINT users_referrer_wallet_address_fkey;
ALTER TABLE ONLY public.user_usdt_withdrawals DROP CONSTRAINT user_usdt_withdrawals_user_wallet_address_fkey;
ALTER TABLE ONLY public.user_task_completions DROP CONSTRAINT user_task_completions_user_wallet_address_fkey;
ALTER TABLE ONLY public.user_task_completions DROP CONSTRAINT user_task_completions_task_id_fkey;
ALTER TABLE ONLY public.user_stakes DROP CONSTRAINT user_stakes_user_wallet_address_fkey;
ALTER TABLE ONLY public.user_stakes DROP CONSTRAINT user_stakes_staking_plan_id_fkey;
ALTER TABLE ONLY public.referral_rewards DROP CONSTRAINT referral_rewards_stake_id_fkey;
ALTER TABLE ONLY public.referral_rewards DROP CONSTRAINT referral_rewards_referrer_wallet_address_fkey;
ALTER TABLE ONLY public.referral_rewards DROP CONSTRAINT referral_rewards_referred_wallet_address_fkey;
ALTER TABLE ONLY public.coinflip_history DROP CONSTRAINT coinflip_history_user_wallet_address_fkey;
DROP TRIGGER users_before_insert_set_referral_code ON public.users;
DROP INDEX public.idx_users_referrer;
DROP INDEX public.idx_users_referral_code;
DROP INDEX public.idx_user_usdt_withdrawals_user;
DROP INDEX public.idx_user_usdt_withdrawals_status;
DROP INDEX public.idx_user_task_completions_user_task;
DROP INDEX public.idx_user_task_completions_status;
DROP INDEX public.idx_user_stakes_wallet_address;
DROP INDEX public.idx_user_stakes_status;
DROP INDEX public.idx_referral_rewards_status;
DROP INDEX public.idx_referral_rewards_referrer;
DROP INDEX public.idx_coinflip_history_user;
DROP INDEX public.idx_announcements_active_pinned_published;
DROP INDEX neon_auth.users_sync_deleted_at_idx;
ALTER TABLE ONLY public.users DROP CONSTRAINT users_telegram_id_key;
ALTER TABLE ONLY public.users DROP CONSTRAINT users_referral_code_key;
ALTER TABLE ONLY public.users DROP CONSTRAINT users_pkey;
ALTER TABLE ONLY public.user_usdt_withdrawals DROP CONSTRAINT user_usdt_withdrawals_pkey1;
ALTER TABLE ONLY public.user_arix_withdrawals DROP CONSTRAINT user_usdt_withdrawals_pkey;
ALTER TABLE ONLY public.user_usdt_withdrawals DROP CONSTRAINT user_usdt_withdrawals_onchain_tx_hash_key1;
ALTER TABLE ONLY public.user_arix_withdrawals DROP CONSTRAINT user_usdt_withdrawals_onchain_tx_hash_key;
ALTER TABLE ONLY public.user_task_completions DROP CONSTRAINT user_task_completions_pkey;
ALTER TABLE ONLY public.user_stakes DROP CONSTRAINT user_stakes_pkey;
ALTER TABLE ONLY public.user_stakes DROP CONSTRAINT user_stakes_onchain_unstake_tx_hash_key;
ALTER TABLE ONLY public.user_stakes DROP CONSTRAINT user_stakes_onchain_stake_tx_hash_key;
ALTER TABLE ONLY public.tasks DROP CONSTRAINT tasks_task_key_key;
ALTER TABLE ONLY public.tasks DROP CONSTRAINT tasks_pkey;
ALTER TABLE ONLY public.staking_plans DROP CONSTRAINT staking_plans_plan_key_key;
ALTER TABLE ONLY public.staking_plans DROP CONSTRAINT staking_plans_pkey;
ALTER TABLE ONLY public.referral_rewards DROP CONSTRAINT referral_rewards_pkey;
ALTER TABLE ONLY public.crash_rounds DROP CONSTRAINT crash_rounds_pkey;
ALTER TABLE ONLY public.coinflip_history DROP CONSTRAINT coinflip_history_pkey;
ALTER TABLE ONLY public.announcements DROP CONSTRAINT announcements_pkey;
ALTER TABLE ONLY neon_auth.users_sync DROP CONSTRAINT users_sync_pkey;
ALTER TABLE public.user_usdt_withdrawals ALTER COLUMN withdrawal_id DROP DEFAULT;
ALTER TABLE public.user_task_completions ALTER COLUMN completion_id DROP DEFAULT;
ALTER TABLE public.user_arix_withdrawals ALTER COLUMN withdrawal_id DROP DEFAULT;
ALTER TABLE public.tasks ALTER COLUMN task_id DROP DEFAULT;
ALTER TABLE public.staking_plans ALTER COLUMN plan_id DROP DEFAULT;
ALTER TABLE public.referral_rewards ALTER COLUMN reward_id DROP DEFAULT;
ALTER TABLE public.crash_rounds ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.coinflip_history ALTER COLUMN game_id DROP DEFAULT;
ALTER TABLE public.announcements ALTER COLUMN announcement_id DROP DEFAULT;
DROP TABLE public.users;
DROP SEQUENCE public.user_usdt_withdrawals_withdrawal_id_seq1;
DROP SEQUENCE public.user_usdt_withdrawals_withdrawal_id_seq;
DROP TABLE public.user_usdt_withdrawals;
DROP SEQUENCE public.user_task_completions_completion_id_seq;
DROP TABLE public.user_task_completions;
DROP TABLE public.user_stakes;
DROP TABLE public.user_arix_withdrawals;
DROP SEQUENCE public.tasks_task_id_seq;
DROP TABLE public.tasks;
DROP SEQUENCE public.staking_plans_plan_id_seq;
DROP TABLE public.staking_plans;
DROP SEQUENCE public.referral_rewards_reward_id_seq;
DROP TABLE public.referral_rewards;
DROP SEQUENCE public.crash_rounds_id_seq;
DROP TABLE public.crash_rounds;
DROP SEQUENCE public.coinflip_history_game_id_seq;
DROP TABLE public.coinflip_history;
DROP SEQUENCE public.announcements_announcement_id_seq;
DROP TABLE public.announcements;
DROP TABLE neon_auth.users_sync;
DROP FUNCTION public.set_referral_code();
DROP FUNCTION public.generate_referral_code();
DROP SCHEMA neon_auth;
--
-- Name: neon_auth; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA neon_auth;


--
-- Name: generate_referral_code(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.generate_referral_code() RETURNS character varying
    LANGUAGE plpgsql
    AS $$
DECLARE
  chars TEXT[] := '{A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z,0,1,2,3,4,5,6,7,8,9}';
  result TEXT := '';
  i INTEGER;
BEGIN
  FOR i IN 1..8 LOOP -- Generate an 8-character code
    result := result || chars[1+random()*(array_length(chars, 1)-1)];
  END LOOP;
  RETURN result;
END;
$$;


--
-- Name: set_referral_code(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.set_referral_code() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  IF NEW.referral_code IS NULL THEN
    LOOP
      NEW.referral_code := generate_referral_code();
      EXIT WHEN NOT EXISTS (SELECT 1 FROM users WHERE referral_code = NEW.referral_code);
    END LOOP;
  END IF;
  RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: users_sync; Type: TABLE; Schema: neon_auth; Owner: -
--

CREATE TABLE neon_auth.users_sync (
    raw_json jsonb NOT NULL,
    id text GENERATED ALWAYS AS ((raw_json ->> 'id'::text)) STORED NOT NULL,
    name text GENERATED ALWAYS AS ((raw_json ->> 'display_name'::text)) STORED,
    email text GENERATED ALWAYS AS ((raw_json ->> 'primary_email'::text)) STORED,
    created_at timestamp with time zone GENERATED ALWAYS AS (to_timestamp((trunc((((raw_json ->> 'signed_up_at_millis'::text))::bigint)::double precision) / (1000)::double precision))) STORED,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: announcements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.announcements (
    announcement_id integer NOT NULL,
    title character varying(255) NOT NULL,
    content text NOT NULL,
    type character varying(50) DEFAULT 'info'::character varying,
    image_url text,
    action_url text,
    action_text character varying(100),
    is_pinned boolean DEFAULT false,
    is_active boolean DEFAULT true,
    published_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    expires_at timestamp with time zone
);


--
-- Name: announcements_announcement_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.announcements_announcement_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: announcements_announcement_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.announcements_announcement_id_seq OWNED BY public.announcements.announcement_id;


--
-- Name: coinflip_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.coinflip_history (
    game_id integer NOT NULL,
    user_wallet_address character varying(68) NOT NULL,
    bet_amount_arix numeric(20,9) NOT NULL,
    choice character varying(10) NOT NULL,
    server_coin_side character varying(10) NOT NULL,
    outcome character varying(10) NOT NULL,
    amount_delta_arix numeric(20,9) NOT NULL,
    played_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: coinflip_history_game_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.coinflip_history_game_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: coinflip_history_game_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.coinflip_history_game_id_seq OWNED BY public.coinflip_history.game_id;


--
-- Name: crash_rounds; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.crash_rounds (
    id integer NOT NULL,
    crash_multiplier numeric(10,2) NOT NULL,
    server_seed character varying(255),
    public_hash character varying(255),
    status character varying(20) DEFAULT 'waiting'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: crash_rounds_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.crash_rounds_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: crash_rounds_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.crash_rounds_id_seq OWNED BY public.crash_rounds.id;


--
-- Name: referral_rewards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.referral_rewards (
    reward_id integer NOT NULL,
    stake_id uuid,
    referrer_wallet_address character varying(68) NOT NULL,
    referred_wallet_address character varying(68) NOT NULL,
    level integer NOT NULL,
    reward_type character varying(50) NOT NULL,
    reward_amount_usdt numeric(20,6) NOT NULL,
    status character varying(20) DEFAULT 'credited'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: referral_rewards_reward_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.referral_rewards_reward_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: referral_rewards_reward_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.referral_rewards_reward_id_seq OWNED BY public.referral_rewards.reward_id;


--
-- Name: staking_plans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.staking_plans (
    plan_id integer NOT NULL,
    plan_key character varying(50) NOT NULL,
    title character varying(100) NOT NULL,
    duration_days integer NOT NULL,
    fixed_usdt_apr_percent numeric(5,2) NOT NULL,
    arix_early_unstake_penalty_percent numeric(5,2) NOT NULL,
    min_stake_usdt numeric(10,2) DEFAULT 0,
    max_stake_usdt numeric(10,2),
    referral_l1_invest_percent numeric(5,2) DEFAULT 0,
    referral_l2_invest_percent numeric(5,2) DEFAULT 0,
    referral_l2_commission_on_l1_bonus_percent numeric(5,2) DEFAULT 0,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: staking_plans_plan_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.staking_plans_plan_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: staking_plans_plan_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.staking_plans_plan_id_seq OWNED BY public.staking_plans.plan_id;


--
-- Name: tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tasks (
    task_id integer NOT NULL,
    task_key character varying(50) NOT NULL,
    title character varying(255) NOT NULL,
    description text,
    reward_arix_amount numeric(20,9) DEFAULT 0,
    task_type character varying(50) DEFAULT 'social'::character varying,
    validation_type character varying(50) DEFAULT 'manual'::character varying,
    action_url text,
    is_active boolean DEFAULT true,
    is_repeatable boolean DEFAULT false,
    max_completions_user integer DEFAULT 1,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: tasks_task_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tasks_task_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tasks_task_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tasks_task_id_seq OWNED BY public.tasks.task_id;


--
-- Name: user_arix_withdrawals; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_arix_withdrawals (
    withdrawal_id integer NOT NULL,
    user_wallet_address character varying(68) NOT NULL,
    amount_arix numeric(20,9) NOT NULL,
    status character varying(20) DEFAULT 'pending_payout'::character varying NOT NULL,
    onchain_tx_hash character varying(64),
    requested_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    processed_at timestamp with time zone
);


--
-- Name: user_stakes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_stakes (
    stake_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_wallet_address character varying(68) NOT NULL,
    staking_plan_id integer NOT NULL,
    arix_amount_staked numeric(20,9) NOT NULL,
    reference_usdt_value_at_stake_time numeric(20,6) NOT NULL,
    stake_timestamp timestamp with time zone NOT NULL,
    unlock_timestamp timestamp with time zone NOT NULL,
    onchain_stake_tx_boc text,
    onchain_stake_tx_hash character varying(64),
    status character varying(30) DEFAULT 'pending_confirmation'::character varying NOT NULL,
    usdt_reward_accrued_total numeric(20,6) DEFAULT 0.00,
    last_usdt_reward_calc_timestamp timestamp with time zone,
    arix_penalty_applied numeric(20,9) DEFAULT 0,
    arix_final_reward_calculated numeric(20,9) DEFAULT 0,
    onchain_unstake_tx_boc text,
    onchain_unstake_tx_hash character varying(64),
    notes text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: user_task_completions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_task_completions (
    completion_id integer NOT NULL,
    user_wallet_address character varying(68) NOT NULL,
    task_id integer NOT NULL,
    status character varying(30) DEFAULT 'pending_verification'::character varying NOT NULL,
    submission_data jsonb,
    completed_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    verified_at timestamp with time zone,
    reward_credited_at timestamp with time zone,
    notes text
);


--
-- Name: user_task_completions_completion_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_task_completions_completion_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_task_completions_completion_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_task_completions_completion_id_seq OWNED BY public.user_task_completions.completion_id;


--
-- Name: user_usdt_withdrawals; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_usdt_withdrawals (
    withdrawal_id integer NOT NULL,
    user_wallet_address character varying(68) NOT NULL,
    amount_usdt numeric(20,6) NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    onchain_tx_hash character varying(64),
    notes text,
    requested_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    processed_at timestamp with time zone
);


--
-- Name: user_usdt_withdrawals_withdrawal_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_usdt_withdrawals_withdrawal_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_usdt_withdrawals_withdrawal_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_usdt_withdrawals_withdrawal_id_seq OWNED BY public.user_arix_withdrawals.withdrawal_id;


--
-- Name: user_usdt_withdrawals_withdrawal_id_seq1; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_usdt_withdrawals_withdrawal_id_seq1
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_usdt_withdrawals_withdrawal_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_usdt_withdrawals_withdrawal_id_seq1 OWNED BY public.user_usdt_withdrawals.withdrawal_id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    wallet_address character varying(68) NOT NULL,
    telegram_id bigint,
    username character varying(255),
    referral_code character varying(10),
    referrer_wallet_address character varying(68),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    claimable_usdt_balance numeric(20,6) DEFAULT 0.00 NOT NULL,
    claimable_arix_rewards numeric(20,9) DEFAULT 0.00 NOT NULL
);


--
-- Name: announcements announcement_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.announcements ALTER COLUMN announcement_id SET DEFAULT nextval('public.announcements_announcement_id_seq'::regclass);


--
-- Name: coinflip_history game_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.coinflip_history ALTER COLUMN game_id SET DEFAULT nextval('public.coinflip_history_game_id_seq'::regclass);


--
-- Name: crash_rounds id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crash_rounds ALTER COLUMN id SET DEFAULT nextval('public.crash_rounds_id_seq'::regclass);


--
-- Name: referral_rewards reward_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_rewards ALTER COLUMN reward_id SET DEFAULT nextval('public.referral_rewards_reward_id_seq'::regclass);


--
-- Name: staking_plans plan_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.staking_plans ALTER COLUMN plan_id SET DEFAULT nextval('public.staking_plans_plan_id_seq'::regclass);


--
-- Name: tasks task_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks ALTER COLUMN task_id SET DEFAULT nextval('public.tasks_task_id_seq'::regclass);


--
-- Name: user_arix_withdrawals withdrawal_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_arix_withdrawals ALTER COLUMN withdrawal_id SET DEFAULT nextval('public.user_usdt_withdrawals_withdrawal_id_seq'::regclass);


--
-- Name: user_task_completions completion_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_task_completions ALTER COLUMN completion_id SET DEFAULT nextval('public.user_task_completions_completion_id_seq'::regclass);


--
-- Name: user_usdt_withdrawals withdrawal_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_usdt_withdrawals ALTER COLUMN withdrawal_id SET DEFAULT nextval('public.user_usdt_withdrawals_withdrawal_id_seq1'::regclass);


--
-- Data for Name: users_sync; Type: TABLE DATA; Schema: neon_auth; Owner: -
--

COPY neon_auth.users_sync (raw_json, updated_at, deleted_at) FROM stdin;
\.


--
-- Data for Name: announcements; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.announcements (announcement_id, title, content, type, image_url, action_url, action_text, is_pinned, is_active, published_at, expires_at) FROM stdin;
\.


--
-- Data for Name: coinflip_history; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.coinflip_history (game_id, user_wallet_address, bet_amount_arix, choice, server_coin_side, outcome, amount_delta_arix, played_at) FROM stdin;
\.


--
-- Data for Name: crash_rounds; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.crash_rounds (id, crash_multiplier, server_seed, public_hash, status, created_at, updated_at) FROM stdin;
1	1.21	\N	\N	running	2025-06-13 16:38:03.017992+00	2025-06-13 16:38:03.017992+00
2	1.13	\N	\N	crashed	2025-06-13 16:42:54.356463+00	2025-06-13 16:42:54.356463+00
3	1.82	\N	\N	running	2025-06-13 16:51:08.252454+00	2025-06-13 16:51:08.252454+00
4	5.64	\N	\N	running	2025-06-13 17:36:57.200568+00	2025-06-13 17:36:57.200568+00
5	2.32	\N	\N	crashed	2025-06-13 18:23:38.735437+00	2025-06-13 18:23:38.735437+00
6	1.18	\N	\N	crashed	2025-06-13 18:24:23.515218+00	2025-06-13 18:24:23.515218+00
7	1.54	\N	\N	crashed	2025-06-13 18:24:45.510766+00	2025-06-13 18:24:45.510766+00
8	3.42	\N	\N	crashed	2025-06-13 18:26:49.071351+00	2025-06-13 18:26:49.071351+00
9	1.55	\N	\N	crashed	2025-06-13 18:32:48.800299+00	2025-06-13 18:32:48.800299+00
10	1.12	\N	\N	crashed	2025-06-13 18:35:48.783438+00	2025-06-13 18:35:48.783438+00
\.


--
-- Data for Name: referral_rewards; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.referral_rewards (reward_id, stake_id, referrer_wallet_address, referred_wallet_address, level, reward_type, reward_amount_usdt, status, created_at) FROM stdin;
\.


--
-- Data for Name: staking_plans; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.staking_plans (plan_id, plan_key, title, duration_days, fixed_usdt_apr_percent, arix_early_unstake_penalty_percent, min_stake_usdt, max_stake_usdt, referral_l1_invest_percent, referral_l2_invest_percent, referral_l2_commission_on_l1_bonus_percent, is_active, created_at) FROM stdin;
1	STARTER	Starter Plan	30	6.00	7.00	100.00	500.00	5.00	1.00	0.00	t	2025-06-03 22:14:19.786755+00
2	BUILDER	Builder Plan	60	7.50	7.00	500.00	1000.00	7.00	2.00	0.00	t	2025-06-03 22:14:19.786755+00
3	ADVANCED	Advanced Plan	90	9.00	7.00	1000.00	5000.00	10.00	0.00	3.00	t	2025-06-03 22:14:19.786755+00
4	VIP	VIP Plan	120	12.00	7.00	5000.00	\N	12.00	0.00	5.00	t	2025-06-03 22:14:19.786755+00
\.


--
-- Data for Name: tasks; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.tasks (task_id, task_key, title, description, reward_arix_amount, task_type, validation_type, action_url, is_active, is_repeatable, max_completions_user, start_date, end_date, created_at) FROM stdin;
1	TWITTER_FOLLOW_ARIX	Follow ARIX on Twitter	Follow our official ARIX Twitter account and verify.	100.000000000	social	manual	https://twitter.com/ARIX_PROJECT_X	t	f	1	\N	\N	2025-06-03 22:14:20.089542+00
2	TELEGRAM_JOIN_ARIX	Join ARIX Telegram Channel	Join our main Telegram channel for updates.	50.000000000	social	manual	https://t.me/ARIX_CHANEL	t	f	1	\N	\N	2025-06-03 22:14:20.089542+00
3	FIRST_STAKE_BONUS	First Stake Bonus	Make your first ARIX stake on any plan and get a bonus!	200.000000000	engagement	auto_approve_on_stake	\N	t	f	1	\N	\N	2025-06-03 22:14:20.089542+00
\.


--
-- Data for Name: user_arix_withdrawals; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.user_arix_withdrawals (withdrawal_id, user_wallet_address, amount_arix, status, onchain_tx_hash, requested_at, processed_at) FROM stdin;
\.


--
-- Data for Name: user_stakes; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.user_stakes (stake_id, user_wallet_address, staking_plan_id, arix_amount_staked, reference_usdt_value_at_stake_time, stake_timestamp, unlock_timestamp, onchain_stake_tx_boc, onchain_stake_tx_hash, status, usdt_reward_accrued_total, last_usdt_reward_calc_timestamp, arix_penalty_applied, arix_final_reward_calculated, onchain_unstake_tx_boc, onchain_unstake_tx_hash, notes, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: user_task_completions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.user_task_completions (completion_id, user_wallet_address, task_id, status, submission_data, completed_at, verified_at, reward_credited_at, notes) FROM stdin;
2	0:43083987ff40670469d1483c27f99f31bd307d6ebebfdeab5b32e7b98d180d2d	3	pending_verification	\N	2025-06-05 15:05:01.130086+00	\N	\N	\N
3	0:43083987ff40670469d1483c27f99f31bd307d6ebebfdeab5b32e7b98d180d2d	2	pending_verification	\N	2025-06-05 15:05:07.995873+00	\N	\N	\N
4	0:43083987ff40670469d1483c27f99f31bd307d6ebebfdeab5b32e7b98d180d2d	1	pending_verification	\N	2025-06-05 19:34:38.065079+00	\N	\N	\N
\.


--
-- Data for Name: user_usdt_withdrawals; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.user_usdt_withdrawals (withdrawal_id, user_wallet_address, amount_usdt, status, onchain_tx_hash, notes, requested_at, processed_at) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.users (wallet_address, telegram_id, username, referral_code, referrer_wallet_address, created_at, updated_at, claimable_usdt_balance, claimable_arix_rewards) FROM stdin;
0:43083987ff40670469d1483c27f99f31bd307d6ebebfdeab5b32e7b98d180d2d	\N	\N	PF7EIWXV	\N	2025-06-04 11:55:12.260376+00	2025-06-04 11:55:12.260376+00	0.000000	0.000000000
0:d42060eeef8d8163c37d4d96a56dd3d6a9448ee71c0b87329f536c5a9c5ca321	691258888	babubig	CS6JJGPS	\N	2025-06-04 01:47:02.276714+00	2025-06-04 14:01:17.878099+00	0.000000	0.000000000
0:67c517996903dce7917688ffdf9c3ba1702d5c90987cb94ed01908de554180e3	\N	\N	BHDANQHH	\N	2025-06-05 00:50:57.575671+00	2025-06-05 00:50:57.575671+00	0.000000	0.000000000
0:bea285b414bf75c07f193dee2bcb69782c0d637a22098b447e2d8d68398e2839	102085319	Yazdan_1374	85WM74RG	\N	2025-06-05 21:18:14.665842+00	2025-06-05 21:32:10.879301+00	0.000000	0.000000000
0:7b66ff85bb9f5aa67450f9224193301acda1d4679d87ade1969a1270a9812460	1897468368	jane_rose_admin	F0831XL3	\N	2025-06-04 03:40:46.395507+00	2025-06-13 14:25:59.901806+00	0.000000	0.000000000
0:2a68ab785b2894ecbde25c302161daaa5b98a1c594fa8117e355a2220a488d56	7290455517	sunwukongdev	RQM4VOBQ	\N	2025-06-04 01:38:05.332298+00	2025-06-10 15:52:09.448019+00	0.000000	0.000000000
0:49525c4124e95aed883e463e44653a305c5f7698376b8bcaf46796aeb01c1bf5	7674252805	bigthingezra	45VVZUJ4	\N	2025-06-04 10:44:57.279505+00	2025-06-11 12:06:24.000506+00	0.000000	0.000000000
\.


--
-- Name: announcements_announcement_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.announcements_announcement_id_seq', 1, false);


--
-- Name: coinflip_history_game_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.coinflip_history_game_id_seq', 1, false);


--
-- Name: crash_rounds_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.crash_rounds_id_seq', 10, true);


--
-- Name: referral_rewards_reward_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.referral_rewards_reward_id_seq', 1, false);


--
-- Name: staking_plans_plan_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.staking_plans_plan_id_seq', 4, true);


--
-- Name: tasks_task_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.tasks_task_id_seq', 3, true);


--
-- Name: user_task_completions_completion_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.user_task_completions_completion_id_seq', 4, true);


--
-- Name: user_usdt_withdrawals_withdrawal_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.user_usdt_withdrawals_withdrawal_id_seq', 1, false);


--
-- Name: user_usdt_withdrawals_withdrawal_id_seq1; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.user_usdt_withdrawals_withdrawal_id_seq1', 1, false);


--
-- Name: users_sync users_sync_pkey; Type: CONSTRAINT; Schema: neon_auth; Owner: -
--

ALTER TABLE ONLY neon_auth.users_sync
    ADD CONSTRAINT users_sync_pkey PRIMARY KEY (id);


--
-- Name: announcements announcements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.announcements
    ADD CONSTRAINT announcements_pkey PRIMARY KEY (announcement_id);


--
-- Name: coinflip_history coinflip_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.coinflip_history
    ADD CONSTRAINT coinflip_history_pkey PRIMARY KEY (game_id);


--
-- Name: crash_rounds crash_rounds_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crash_rounds
    ADD CONSTRAINT crash_rounds_pkey PRIMARY KEY (id);


--
-- Name: referral_rewards referral_rewards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_rewards
    ADD CONSTRAINT referral_rewards_pkey PRIMARY KEY (reward_id);


--
-- Name: staking_plans staking_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.staking_plans
    ADD CONSTRAINT staking_plans_pkey PRIMARY KEY (plan_id);


--
-- Name: staking_plans staking_plans_plan_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.staking_plans
    ADD CONSTRAINT staking_plans_plan_key_key UNIQUE (plan_key);


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (task_id);


--
-- Name: tasks tasks_task_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_task_key_key UNIQUE (task_key);


--
-- Name: user_stakes user_stakes_onchain_stake_tx_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_stakes
    ADD CONSTRAINT user_stakes_onchain_stake_tx_hash_key UNIQUE (onchain_stake_tx_hash);


--
-- Name: user_stakes user_stakes_onchain_unstake_tx_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_stakes
    ADD CONSTRAINT user_stakes_onchain_unstake_tx_hash_key UNIQUE (onchain_unstake_tx_hash);


--
-- Name: user_stakes user_stakes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_stakes
    ADD CONSTRAINT user_stakes_pkey PRIMARY KEY (stake_id);


--
-- Name: user_task_completions user_task_completions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_task_completions
    ADD CONSTRAINT user_task_completions_pkey PRIMARY KEY (completion_id);


--
-- Name: user_arix_withdrawals user_usdt_withdrawals_onchain_tx_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_arix_withdrawals
    ADD CONSTRAINT user_usdt_withdrawals_onchain_tx_hash_key UNIQUE (onchain_tx_hash);


--
-- Name: user_usdt_withdrawals user_usdt_withdrawals_onchain_tx_hash_key1; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_usdt_withdrawals
    ADD CONSTRAINT user_usdt_withdrawals_onchain_tx_hash_key1 UNIQUE (onchain_tx_hash);


--
-- Name: user_arix_withdrawals user_usdt_withdrawals_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_arix_withdrawals
    ADD CONSTRAINT user_usdt_withdrawals_pkey PRIMARY KEY (withdrawal_id);


--
-- Name: user_usdt_withdrawals user_usdt_withdrawals_pkey1; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_usdt_withdrawals
    ADD CONSTRAINT user_usdt_withdrawals_pkey1 PRIMARY KEY (withdrawal_id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (wallet_address);


--
-- Name: users users_referral_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_referral_code_key UNIQUE (referral_code);


--
-- Name: users users_telegram_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_telegram_id_key UNIQUE (telegram_id);


--
-- Name: users_sync_deleted_at_idx; Type: INDEX; Schema: neon_auth; Owner: -
--

CREATE INDEX users_sync_deleted_at_idx ON neon_auth.users_sync USING btree (deleted_at);


--
-- Name: idx_announcements_active_pinned_published; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_announcements_active_pinned_published ON public.announcements USING btree (is_active, is_pinned DESC, published_at DESC);


--
-- Name: idx_coinflip_history_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_coinflip_history_user ON public.coinflip_history USING btree (user_wallet_address);


--
-- Name: idx_referral_rewards_referrer; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_referral_rewards_referrer ON public.referral_rewards USING btree (referrer_wallet_address);


--
-- Name: idx_referral_rewards_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_referral_rewards_status ON public.referral_rewards USING btree (status);


--
-- Name: idx_user_stakes_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_stakes_status ON public.user_stakes USING btree (status);


--
-- Name: idx_user_stakes_wallet_address; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_stakes_wallet_address ON public.user_stakes USING btree (user_wallet_address);


--
-- Name: idx_user_task_completions_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_task_completions_status ON public.user_task_completions USING btree (status);


--
-- Name: idx_user_task_completions_user_task; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_task_completions_user_task ON public.user_task_completions USING btree (user_wallet_address, task_id);


--
-- Name: idx_user_usdt_withdrawals_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_usdt_withdrawals_status ON public.user_arix_withdrawals USING btree (status);


--
-- Name: idx_user_usdt_withdrawals_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_usdt_withdrawals_user ON public.user_arix_withdrawals USING btree (user_wallet_address);


--
-- Name: idx_users_referral_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_referral_code ON public.users USING btree (referral_code);


--
-- Name: idx_users_referrer; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_referrer ON public.users USING btree (referrer_wallet_address);


--
-- Name: users users_before_insert_set_referral_code; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER users_before_insert_set_referral_code BEFORE INSERT ON public.users FOR EACH ROW EXECUTE FUNCTION public.set_referral_code();


--
-- Name: coinflip_history coinflip_history_user_wallet_address_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.coinflip_history
    ADD CONSTRAINT coinflip_history_user_wallet_address_fkey FOREIGN KEY (user_wallet_address) REFERENCES public.users(wallet_address);


--
-- Name: referral_rewards referral_rewards_referred_wallet_address_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_rewards
    ADD CONSTRAINT referral_rewards_referred_wallet_address_fkey FOREIGN KEY (referred_wallet_address) REFERENCES public.users(wallet_address);


--
-- Name: referral_rewards referral_rewards_referrer_wallet_address_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_rewards
    ADD CONSTRAINT referral_rewards_referrer_wallet_address_fkey FOREIGN KEY (referrer_wallet_address) REFERENCES public.users(wallet_address);


--
-- Name: referral_rewards referral_rewards_stake_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_rewards
    ADD CONSTRAINT referral_rewards_stake_id_fkey FOREIGN KEY (stake_id) REFERENCES public.user_stakes(stake_id) ON DELETE SET NULL;


--
-- Name: user_stakes user_stakes_staking_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_stakes
    ADD CONSTRAINT user_stakes_staking_plan_id_fkey FOREIGN KEY (staking_plan_id) REFERENCES public.staking_plans(plan_id);


--
-- Name: user_stakes user_stakes_user_wallet_address_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_stakes
    ADD CONSTRAINT user_stakes_user_wallet_address_fkey FOREIGN KEY (user_wallet_address) REFERENCES public.users(wallet_address);


--
-- Name: user_task_completions user_task_completions_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_task_completions
    ADD CONSTRAINT user_task_completions_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(task_id);


--
-- Name: user_task_completions user_task_completions_user_wallet_address_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_task_completions
    ADD CONSTRAINT user_task_completions_user_wallet_address_fkey FOREIGN KEY (user_wallet_address) REFERENCES public.users(wallet_address);


--
-- Name: user_usdt_withdrawals user_usdt_withdrawals_user_wallet_address_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_usdt_withdrawals
    ADD CONSTRAINT user_usdt_withdrawals_user_wallet_address_fkey FOREIGN KEY (user_wallet_address) REFERENCES public.users(wallet_address);


--
-- Name: users users_referrer_wallet_address_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_referrer_wallet_address_fkey FOREIGN KEY (referrer_wallet_address) REFERENCES public.users(wallet_address) ON DELETE SET NULL;


--
-- PostgreSQL database dump complete
--

