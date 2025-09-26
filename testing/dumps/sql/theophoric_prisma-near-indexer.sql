-- Downloaded from: https://github.com/theophoric/prisma-near-indexer/blob/5f8db0245d12b8e700b45c91e9f8266b5b78cd63/lib/schema.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 11.12
-- Dumped by pg_dump version 13.4

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
-- Name: access_key_permission_kind; Type: TYPE; Schema: public; Owner: testnet
--

CREATE TYPE public.access_key_permission_kind AS ENUM (
    'FULL_ACCESS',
    'FUNCTION_CALL'
);


ALTER TYPE public.access_key_permission_kind OWNER TO testnet;

--
-- Name: action_kind; Type: TYPE; Schema: public; Owner: testnet
--

CREATE TYPE public.action_kind AS ENUM (
    'CREATE_ACCOUNT',
    'DEPLOY_CONTRACT',
    'FUNCTION_CALL',
    'TRANSFER',
    'STAKE',
    'ADD_KEY',
    'DELETE_KEY',
    'DELETE_ACCOUNT'
);


ALTER TYPE public.action_kind OWNER TO testnet;

--
-- Name: execution_outcome_status; Type: TYPE; Schema: public; Owner: testnet
--

CREATE TYPE public.execution_outcome_status AS ENUM (
    'UNKNOWN',
    'FAILURE',
    'SUCCESS_VALUE',
    'SUCCESS_RECEIPT_ID'
);


ALTER TYPE public.execution_outcome_status OWNER TO testnet;

--
-- Name: receipt_kind; Type: TYPE; Schema: public; Owner: testnet
--

CREATE TYPE public.receipt_kind AS ENUM (
    'ACTION',
    'DATA'
);


ALTER TYPE public.receipt_kind OWNER TO testnet;

--
-- Name: state_change_reason_kind; Type: TYPE; Schema: public; Owner: testnet
--

CREATE TYPE public.state_change_reason_kind AS ENUM (
    'TRANSACTION_PROCESSING',
    'ACTION_RECEIPT_PROCESSING_STARTED',
    'ACTION_RECEIPT_GAS_REWARD',
    'RECEIPT_PROCESSING',
    'POSTPONED_RECEIPT',
    'UPDATED_DELAYED_RECEIPTS',
    'VALIDATOR_ACCOUNTS_UPDATE',
    'MIGRATION'
);


ALTER TYPE public.state_change_reason_kind OWNER TO testnet;

--
-- Name: decode_or_null(bytea); Type: FUNCTION; Schema: public; Owner: testnet
--

CREATE FUNCTION public.decode_or_null(bytea) RETURNS jsonb
    LANGUAGE plpgsql
    AS $_$BEGIN
   RETURN convert_from($1, 'UTF8')::jsonb;
EXCEPTION
   WHEN others THEN
      RAISE WARNING '%', SQLERRM;
RETURN '{}'::jsonb;

END;$_$;


ALTER FUNCTION public.decode_or_null(bytea) OWNER TO testnet;

--
-- Name: diesel_manage_updated_at(regclass); Type: FUNCTION; Schema: public; Owner: testnet
--

CREATE FUNCTION public.diesel_manage_updated_at(_tbl regclass) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    EXECUTE format('CREATE TRIGGER set_updated_at BEFORE UPDATE ON %s
                    FOR EACH ROW EXECUTE PROCEDURE diesel_set_updated_at()', _tbl);
END;
$$;


ALTER FUNCTION public.diesel_manage_updated_at(_tbl regclass) OWNER TO testnet;

--
-- Name: diesel_set_updated_at(); Type: FUNCTION; Schema: public; Owner: testnet
--

CREATE FUNCTION public.diesel_set_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF (
        NEW IS DISTINCT FROM OLD AND
        NEW.updated_at IS NOT DISTINCT FROM OLD.updated_at
    ) THEN
        NEW.updated_at := current_timestamp;
    END IF;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.diesel_set_updated_at() OWNER TO testnet;

SET default_tablespace = '';

--
-- Name: __diesel_schema_migrations; Type: TABLE; Schema: public; Owner: testnet
--

CREATE TABLE public.__diesel_schema_migrations (
    version character varying(50) NOT NULL,
    run_on timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.__diesel_schema_migrations OWNER TO testnet;

--
-- Name: access_keys; Type: TABLE; Schema: public; Owner: testnet
--

CREATE TABLE public.access_keys (
    public_key text NOT NULL,
    account_id text NOT NULL,
    created_by_receipt_id text,
    deleted_by_receipt_id text,
    permission_kind public.access_key_permission_kind NOT NULL,
    last_update_block_height numeric(20,0) NOT NULL
);


ALTER TABLE public.access_keys OWNER TO testnet;

--
-- Name: account_changes; Type: TABLE; Schema: public; Owner: testnet
--

CREATE TABLE public.account_changes (
    id bigint NOT NULL,
    affected_account_id text NOT NULL,
    changed_in_block_timestamp numeric(20,0) NOT NULL,
    changed_in_block_hash text NOT NULL,
    caused_by_transaction_hash text,
    caused_by_receipt_id text,
    update_reason public.state_change_reason_kind NOT NULL,
    affected_account_nonstaked_balance numeric(45,0) NOT NULL,
    affected_account_staked_balance numeric(45,0) NOT NULL,
    affected_account_storage_usage numeric(20,0) NOT NULL,
    index_in_block integer NOT NULL
);


ALTER TABLE public.account_changes OWNER TO testnet;

--
-- Name: account_changes_id_seq; Type: SEQUENCE; Schema: public; Owner: testnet
--

CREATE SEQUENCE public.account_changes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.account_changes_id_seq OWNER TO testnet;

--
-- Name: account_changes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: testnet
--

ALTER SEQUENCE public.account_changes_id_seq OWNED BY public.account_changes.id;


--
-- Name: accounts; Type: TABLE; Schema: public; Owner: testnet
--

CREATE TABLE public.accounts (
    id bigint NOT NULL,
    account_id text NOT NULL,
    created_by_receipt_id text,
    deleted_by_receipt_id text,
    last_update_block_height numeric(20,0) NOT NULL
);


ALTER TABLE public.accounts OWNER TO testnet;

--
-- Name: accounts_id_seq; Type: SEQUENCE; Schema: public; Owner: testnet
--

CREATE SEQUENCE public.accounts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.accounts_id_seq OWNER TO testnet;

--
-- Name: accounts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: testnet
--

ALTER SEQUENCE public.accounts_id_seq OWNED BY public.accounts.id;


--
-- Name: action_receipt_actions; Type: TABLE; Schema: public; Owner: testnet
--

CREATE TABLE public.action_receipt_actions (
    receipt_id text NOT NULL,
    index_in_action_receipt integer NOT NULL,
    action_kind public.action_kind NOT NULL,
    args jsonb NOT NULL,
    receipt_predecessor_account_id text NOT NULL,
    receipt_receiver_account_id text NOT NULL,
    receipt_included_in_block_timestamp numeric(20,0) NOT NULL
);


ALTER TABLE public.action_receipt_actions OWNER TO testnet;

--
-- Name: action_receipt_input_data; Type: TABLE; Schema: public; Owner: testnet
--

CREATE TABLE public.action_receipt_input_data (
    input_data_id text NOT NULL,
    input_to_receipt_id text NOT NULL
);


ALTER TABLE public.action_receipt_input_data OWNER TO testnet;

--
-- Name: action_receipt_output_data; Type: TABLE; Schema: public; Owner: testnet
--

CREATE TABLE public.action_receipt_output_data (
    output_data_id text NOT NULL,
    output_from_receipt_id text NOT NULL,
    receiver_account_id text NOT NULL
);


ALTER TABLE public.action_receipt_output_data OWNER TO testnet;

--
-- Name: action_receipts; Type: TABLE; Schema: public; Owner: testnet
--

CREATE TABLE public.action_receipts (
    receipt_id text NOT NULL,
    signer_account_id text NOT NULL,
    signer_public_key text NOT NULL,
    gas_price numeric(45,0) NOT NULL
);


ALTER TABLE public.action_receipts OWNER TO testnet;

--
-- Name: aggregated__circulating_supply; Type: TABLE; Schema: public; Owner: testnet
--

CREATE TABLE public.aggregated__circulating_supply (
    computed_at_block_timestamp numeric(20,0) NOT NULL,
    computed_at_block_hash text NOT NULL,
    circulating_tokens_supply numeric(45,0) NOT NULL,
    total_tokens_supply numeric(45,0) NOT NULL,
    total_lockup_contracts_count integer NOT NULL,
    unfinished_lockup_contracts_count integer NOT NULL,
    foundation_locked_tokens numeric(45,0) NOT NULL,
    lockups_locked_tokens numeric(45,0) NOT NULL
);


ALTER TABLE public.aggregated__circulating_supply OWNER TO testnet;

--
-- Name: blocks; Type: TABLE; Schema: public; Owner: testnet
--

CREATE TABLE public.blocks (
    block_height numeric(20,0) NOT NULL,
    block_hash text NOT NULL,
    prev_block_hash text NOT NULL,
    block_timestamp numeric(20,0) NOT NULL,
    total_supply numeric(45,0) NOT NULL,
    gas_price numeric(45,0) NOT NULL,
    author_account_id text NOT NULL
);


ALTER TABLE public.blocks OWNER TO testnet;

--
-- Name: receipts; Type: TABLE; Schema: public; Owner: testnet
--

CREATE TABLE public.receipts (
    receipt_id text NOT NULL,
    included_in_block_hash text NOT NULL,
    included_in_chunk_hash text NOT NULL,
    index_in_chunk integer NOT NULL,
    included_in_block_timestamp numeric(20,0) NOT NULL,
    predecessor_account_id text NOT NULL,
    receiver_account_id text NOT NULL,
    receipt_kind public.receipt_kind NOT NULL,
    originated_from_transaction_hash text NOT NULL
);


ALTER TABLE public.receipts OWNER TO testnet;

--
-- Name: aggregated__lockups; Type: VIEW; Schema: public; Owner: testnet
--

CREATE VIEW public.aggregated__lockups AS
 SELECT accounts.account_id,
    blocks_start.block_height AS creation_block_height,
    blocks_end.block_height AS deletion_block_height
   FROM ((((public.accounts
     LEFT JOIN public.receipts receipts_start ON ((accounts.created_by_receipt_id = receipts_start.receipt_id)))
     LEFT JOIN public.blocks blocks_start ON ((receipts_start.included_in_block_hash = blocks_start.block_hash)))
     LEFT JOIN public.receipts receipts_end ON ((accounts.deleted_by_receipt_id = receipts_end.receipt_id)))
     LEFT JOIN public.blocks blocks_end ON ((receipts_end.included_in_block_hash = blocks_end.block_hash)))
  WHERE (accounts.account_id ~~ '%.lockup.near'::text);


ALTER TABLE public.aggregated__lockups OWNER TO testnet;

--
-- Name: chunks; Type: TABLE; Schema: public; Owner: testnet
--

CREATE TABLE public.chunks (
    included_in_block_hash text NOT NULL,
    chunk_hash text NOT NULL,
    shard_id numeric(20,0) NOT NULL,
    signature text NOT NULL,
    gas_limit numeric(20,0) NOT NULL,
    gas_used numeric(20,0) NOT NULL,
    author_account_id text NOT NULL
);


ALTER TABLE public.chunks OWNER TO testnet;

--
-- Name: data_receipts; Type: TABLE; Schema: public; Owner: testnet
--

CREATE TABLE public.data_receipts (
    data_id text NOT NULL,
    receipt_id text NOT NULL,
    data bytea
);


ALTER TABLE public.data_receipts OWNER TO testnet;

--
-- Name: execution_outcome_receipts; Type: TABLE; Schema: public; Owner: testnet
--

CREATE TABLE public.execution_outcome_receipts (
    executed_receipt_id text NOT NULL,
    index_in_execution_outcome integer NOT NULL,
    produced_receipt_id text NOT NULL
);


ALTER TABLE public.execution_outcome_receipts OWNER TO testnet;

--
-- Name: execution_outcomes; Type: TABLE; Schema: public; Owner: testnet
--

CREATE TABLE public.execution_outcomes (
    receipt_id text NOT NULL,
    executed_in_block_hash text NOT NULL,
    executed_in_block_timestamp numeric(20,0) NOT NULL,
    index_in_chunk integer NOT NULL,
    gas_burnt numeric(20,0) NOT NULL,
    tokens_burnt numeric(45,0) NOT NULL,
    executor_account_id text NOT NULL,
    status public.execution_outcome_status NOT NULL,
    shard_id numeric(20,0) NOT NULL
);


ALTER TABLE public.execution_outcomes OWNER TO testnet;

--
-- Name: transaction_actions; Type: TABLE; Schema: public; Owner: testnet
--

CREATE TABLE public.transaction_actions (
    transaction_hash text NOT NULL,
    index_in_transaction integer NOT NULL,
    action_kind public.action_kind NOT NULL,
    args jsonb NOT NULL
);


ALTER TABLE public.transaction_actions OWNER TO testnet;

--
-- Name: transactions; Type: TABLE; Schema: public; Owner: testnet
--

CREATE TABLE public.transactions (
    transaction_hash text NOT NULL,
    included_in_block_hash text NOT NULL,
    included_in_chunk_hash text NOT NULL,
    index_in_chunk integer NOT NULL,
    block_timestamp numeric(20,0) NOT NULL,
    signer_account_id text NOT NULL,
    signer_public_key text NOT NULL,
    nonce numeric(20,0) NOT NULL,
    receiver_account_id text NOT NULL,
    signature text NOT NULL,
    status public.execution_outcome_status NOT NULL,
    converted_into_receipt_id text NOT NULL,
    receipt_conversion_gas_burnt numeric(20,0),
    receipt_conversion_tokens_burnt numeric(45,0)
);


ALTER TABLE public.transactions OWNER TO testnet;

--
-- Name: account_changes id; Type: DEFAULT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.account_changes ALTER COLUMN id SET DEFAULT nextval('public.account_changes_id_seq'::regclass);


--
-- Name: accounts id; Type: DEFAULT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.accounts ALTER COLUMN id SET DEFAULT nextval('public.accounts_id_seq'::regclass);


--
-- Name: __diesel_schema_migrations __diesel_schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.__diesel_schema_migrations
    ADD CONSTRAINT __diesel_schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: access_keys access_keys_pk; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.access_keys
    ADD CONSTRAINT access_keys_pk PRIMARY KEY (public_key, account_id);


--
-- Name: account_changes account_changes_pkey; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.account_changes
    ADD CONSTRAINT account_changes_pkey PRIMARY KEY (id);


--
-- Name: accounts accounts_account_id_key; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_account_id_key UNIQUE (account_id);


--
-- Name: accounts accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);


--
-- Name: action_receipt_input_data action_input_pk; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.action_receipt_input_data
    ADD CONSTRAINT action_input_pk PRIMARY KEY (input_data_id, input_to_receipt_id);


--
-- Name: action_receipt_output_data action_output_pk; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.action_receipt_output_data
    ADD CONSTRAINT action_output_pk PRIMARY KEY (output_data_id, output_from_receipt_id);


--
-- Name: aggregated__circulating_supply aggregated__circulating_supply_pkey; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.aggregated__circulating_supply
    ADD CONSTRAINT aggregated__circulating_supply_pkey PRIMARY KEY (computed_at_block_hash);


--
-- Name: blocks blocks_pkey; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.blocks
    ADD CONSTRAINT blocks_pkey PRIMARY KEY (block_hash);


--
-- Name: chunks chunks_pkey; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.chunks
    ADD CONSTRAINT chunks_pkey PRIMARY KEY (chunk_hash);


--
-- Name: execution_outcome_receipts execution_outcome_receipt_pk; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.execution_outcome_receipts
    ADD CONSTRAINT execution_outcome_receipt_pk PRIMARY KEY (executed_receipt_id, index_in_execution_outcome, produced_receipt_id);


--
-- Name: execution_outcomes execution_outcomes_pkey; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.execution_outcomes
    ADD CONSTRAINT execution_outcomes_pkey PRIMARY KEY (receipt_id);


--
-- Name: action_receipt_actions receipt_action_action_pk; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.action_receipt_actions
    ADD CONSTRAINT receipt_action_action_pk PRIMARY KEY (receipt_id, index_in_action_receipt);


--
-- Name: action_receipts receipt_actions_pkey; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.action_receipts
    ADD CONSTRAINT receipt_actions_pkey PRIMARY KEY (receipt_id);


--
-- Name: data_receipts receipt_data_pkey; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.data_receipts
    ADD CONSTRAINT receipt_data_pkey PRIMARY KEY (data_id);


--
-- Name: receipts receipts_pkey; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.receipts
    ADD CONSTRAINT receipts_pkey PRIMARY KEY (receipt_id);


--
-- Name: transaction_actions transaction_action_pk; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.transaction_actions
    ADD CONSTRAINT transaction_action_pk PRIMARY KEY (transaction_hash, index_in_transaction);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (transaction_hash);


--
-- Name: access_keys_account_id_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX access_keys_account_id_idx ON public.access_keys USING btree (account_id);


--
-- Name: access_keys_last_update_block_height_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX access_keys_last_update_block_height_idx ON public.access_keys USING btree (last_update_block_height);


--
-- Name: access_keys_public_key_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX access_keys_public_key_idx ON public.access_keys USING btree (public_key);


--
-- Name: account_changes_affected_account_id_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX account_changes_affected_account_id_idx ON public.account_changes USING btree (affected_account_id);


--
-- Name: account_changes_changed_in_block_hash_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX account_changes_changed_in_block_hash_idx ON public.account_changes USING btree (changed_in_block_hash);


--
-- Name: account_changes_changed_in_block_timestamp_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX account_changes_changed_in_block_timestamp_idx ON public.account_changes USING btree (changed_in_block_timestamp);


--
-- Name: account_changes_changed_in_caused_by_receipt_id_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX account_changes_changed_in_caused_by_receipt_id_idx ON public.account_changes USING btree (caused_by_receipt_id);


--
-- Name: account_changes_changed_in_caused_by_transaction_hash_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX account_changes_changed_in_caused_by_transaction_hash_idx ON public.account_changes USING btree (caused_by_transaction_hash);


--
-- Name: account_changes_null_uni_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE UNIQUE INDEX account_changes_null_uni_idx ON public.account_changes USING btree (affected_account_id, changed_in_block_hash, update_reason, affected_account_nonstaked_balance, affected_account_staked_balance, affected_account_storage_usage) WHERE ((caused_by_transaction_hash IS NULL) AND (caused_by_receipt_id IS NULL));


--
-- Name: account_changes_receipt_uni_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE UNIQUE INDEX account_changes_receipt_uni_idx ON public.account_changes USING btree (affected_account_id, changed_in_block_hash, caused_by_receipt_id, update_reason, affected_account_nonstaked_balance, affected_account_staked_balance, affected_account_storage_usage) WHERE ((caused_by_transaction_hash IS NULL) AND (caused_by_receipt_id IS NOT NULL));


--
-- Name: account_changes_sorting_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX account_changes_sorting_idx ON public.account_changes USING btree (changed_in_block_timestamp, index_in_block);


--
-- Name: account_changes_transaction_uni_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE UNIQUE INDEX account_changes_transaction_uni_idx ON public.account_changes USING btree (affected_account_id, changed_in_block_hash, caused_by_transaction_hash, update_reason, affected_account_nonstaked_balance, affected_account_staked_balance, affected_account_storage_usage) WHERE ((caused_by_transaction_hash IS NOT NULL) AND (caused_by_receipt_id IS NULL));


--
-- Name: accounts_last_update_block_height_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX accounts_last_update_block_height_idx ON public.accounts USING btree (last_update_block_height);


--
-- Name: action_receipt_actions_args_amount_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX action_receipt_actions_args_amount_idx ON public.action_receipt_actions USING btree ((((args -> 'args_json'::text) ->> 'amount'::text))) WHERE ((action_kind = 'FUNCTION_CALL'::public.action_kind) AND ((args ->> 'args_json'::text) IS NOT NULL));


--
-- Name: action_receipt_actions_args_function_call_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX action_receipt_actions_args_function_call_idx ON public.action_receipt_actions USING btree (((args ->> 'method_name'::text))) WHERE (action_kind = 'FUNCTION_CALL'::public.action_kind);


--
-- Name: action_receipt_actions_args_receiver_id_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX action_receipt_actions_args_receiver_id_idx ON public.action_receipt_actions USING btree ((((args -> 'args_json'::text) ->> 'receiver_id'::text))) WHERE ((action_kind = 'FUNCTION_CALL'::public.action_kind) AND ((args ->> 'args_json'::text) IS NOT NULL));


--
-- Name: action_receipt_actions_receipt_included_in_block_timestamp_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX action_receipt_actions_receipt_included_in_block_timestamp_idx ON public.action_receipt_actions USING btree (receipt_included_in_block_timestamp);


--
-- Name: action_receipt_actions_receipt_predecessor_account_id_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX action_receipt_actions_receipt_predecessor_account_id_idx ON public.action_receipt_actions USING btree (receipt_predecessor_account_id);


--
-- Name: action_receipt_actions_receipt_receiver_account_id_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX action_receipt_actions_receipt_receiver_account_id_idx ON public.action_receipt_actions USING btree (receipt_receiver_account_id);


--
-- Name: action_receipt_input_data_input_data_id_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX action_receipt_input_data_input_data_id_idx ON public.action_receipt_input_data USING btree (input_data_id);


--
-- Name: action_receipt_input_data_input_to_receipt_id_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX action_receipt_input_data_input_to_receipt_id_idx ON public.action_receipt_input_data USING btree (input_to_receipt_id);


--
-- Name: action_receipt_output_data_output_data_id_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX action_receipt_output_data_output_data_id_idx ON public.action_receipt_output_data USING btree (output_data_id);


--
-- Name: action_receipt_output_data_output_from_receipt_id_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX action_receipt_output_data_output_from_receipt_id_idx ON public.action_receipt_output_data USING btree (output_from_receipt_id);


--
-- Name: action_receipt_output_data_receiver_account_id_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX action_receipt_output_data_receiver_account_id_idx ON public.action_receipt_output_data USING btree (receiver_account_id);


--
-- Name: action_receipt_signer_account_id_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX action_receipt_signer_account_id_idx ON public.action_receipts USING btree (signer_account_id);


--
-- Name: aggregated__circulating_supply_timestamp_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX aggregated__circulating_supply_timestamp_idx ON public.aggregated__circulating_supply USING btree (computed_at_block_timestamp);


--
-- Name: blocks_height_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX blocks_height_idx ON public.blocks USING btree (block_height);


--
-- Name: blocks_prev_hash_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX blocks_prev_hash_idx ON public.blocks USING btree (prev_block_hash);


--
-- Name: blocks_timestamp_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX blocks_timestamp_idx ON public.blocks USING btree (block_timestamp);


--
-- Name: chunks_included_in_block_hash_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX chunks_included_in_block_hash_idx ON public.chunks USING btree (included_in_block_hash);


--
-- Name: data_receipts_receipt_id_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX data_receipts_receipt_id_idx ON public.data_receipts USING btree (receipt_id);


--
-- Name: execution_outcome_executed_in_block_timestamp; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX execution_outcome_executed_in_block_timestamp ON public.execution_outcomes USING btree (executed_in_block_timestamp);


--
-- Name: execution_outcome_receipts_produced_receipt_id; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX execution_outcome_receipts_produced_receipt_id ON public.execution_outcome_receipts USING btree (produced_receipt_id);


--
-- Name: execution_outcomes_block_hash_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX execution_outcomes_block_hash_idx ON public.execution_outcomes USING btree (executed_in_block_hash);


--
-- Name: execution_outcomes_receipt_id_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX execution_outcomes_receipt_id_idx ON public.execution_outcomes USING btree (receipt_id);


--
-- Name: receipts_included_in_block_hash_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX receipts_included_in_block_hash_idx ON public.receipts USING btree (included_in_block_hash);


--
-- Name: receipts_included_in_chunk_hash_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX receipts_included_in_chunk_hash_idx ON public.receipts USING btree (included_in_chunk_hash);


--
-- Name: receipts_predecessor_account_id_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX receipts_predecessor_account_id_idx ON public.receipts USING btree (predecessor_account_id);


--
-- Name: receipts_receiver_account_id_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX receipts_receiver_account_id_idx ON public.receipts USING btree (receiver_account_id);


--
-- Name: receipts_timestamp_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX receipts_timestamp_idx ON public.receipts USING btree (included_in_block_timestamp);


--
-- Name: transactions_converted_into_receipt_id_dx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX transactions_converted_into_receipt_id_dx ON public.transactions USING btree (converted_into_receipt_id);


--
-- Name: transactions_included_in_block_hash_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX transactions_included_in_block_hash_idx ON public.transactions USING btree (included_in_block_hash);


--
-- Name: transactions_included_in_block_timestamp_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX transactions_included_in_block_timestamp_idx ON public.transactions USING btree (block_timestamp);


--
-- Name: transactions_included_in_chunk_hash_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX transactions_included_in_chunk_hash_idx ON public.transactions USING btree (included_in_chunk_hash);


--
-- Name: transactions_signer_account_id_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX transactions_signer_account_id_idx ON public.transactions USING btree (signer_account_id);


--
-- Name: transactions_signer_public_key_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX transactions_signer_public_key_idx ON public.transactions USING btree (signer_public_key);


--
-- Name: transactions_sorting_idx; Type: INDEX; Schema: public; Owner: testnet
--

CREATE INDEX transactions_sorting_idx ON public.transactions USING btree (block_timestamp, index_in_chunk);


--
-- Name: account_changes account_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.account_changes
    ADD CONSTRAINT account_id_fk FOREIGN KEY (affected_account_id) REFERENCES public.accounts(account_id) ON DELETE CASCADE;


--
-- Name: action_receipt_actions action_receipt_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.action_receipt_actions
    ADD CONSTRAINT action_receipt_fk FOREIGN KEY (receipt_id) REFERENCES public.receipts(receipt_id) ON DELETE CASCADE;


--
-- Name: aggregated__circulating_supply aggregated__circulating_supply_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.aggregated__circulating_supply
    ADD CONSTRAINT aggregated__circulating_supply_fk FOREIGN KEY (computed_at_block_hash) REFERENCES public.blocks(block_hash) ON DELETE CASCADE;


--
-- Name: execution_outcomes block_hash_execution_outcome_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.execution_outcomes
    ADD CONSTRAINT block_hash_execution_outcome_fk FOREIGN KEY (executed_in_block_hash) REFERENCES public.blocks(block_hash) ON DELETE CASCADE;


--
-- Name: account_changes block_hash_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.account_changes
    ADD CONSTRAINT block_hash_fk FOREIGN KEY (changed_in_block_hash) REFERENCES public.blocks(block_hash) ON DELETE CASCADE;


--
-- Name: receipts block_receipts_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.receipts
    ADD CONSTRAINT block_receipts_fk FOREIGN KEY (included_in_block_hash) REFERENCES public.blocks(block_hash) ON DELETE CASCADE;


--
-- Name: transactions block_tx_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT block_tx_fk FOREIGN KEY (included_in_block_hash) REFERENCES public.blocks(block_hash) ON DELETE CASCADE;


--
-- Name: receipts chunk_receipts_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.receipts
    ADD CONSTRAINT chunk_receipts_fk FOREIGN KEY (included_in_chunk_hash) REFERENCES public.chunks(chunk_hash) ON DELETE CASCADE;


--
-- Name: transactions chunk_tx_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT chunk_tx_fk FOREIGN KEY (included_in_chunk_hash) REFERENCES public.chunks(chunk_hash) ON DELETE CASCADE;


--
-- Name: chunks chunks_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.chunks
    ADD CONSTRAINT chunks_fk FOREIGN KEY (included_in_block_hash) REFERENCES public.blocks(block_hash) ON DELETE CASCADE;


--
-- Name: access_keys created_by_receipt_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.access_keys
    ADD CONSTRAINT created_by_receipt_fk FOREIGN KEY (created_by_receipt_id) REFERENCES public.receipts(receipt_id) ON DELETE CASCADE;


--
-- Name: accounts created_receipt_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT created_receipt_fk FOREIGN KEY (created_by_receipt_id) REFERENCES public.receipts(receipt_id) ON DELETE CASCADE;


--
-- Name: access_keys deleted_by_receipt_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.access_keys
    ADD CONSTRAINT deleted_by_receipt_fk FOREIGN KEY (deleted_by_receipt_id) REFERENCES public.receipts(receipt_id) ON DELETE CASCADE;


--
-- Name: accounts deleted_receipt_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT deleted_receipt_fk FOREIGN KEY (deleted_by_receipt_id) REFERENCES public.receipts(receipt_id) ON DELETE CASCADE;


--
-- Name: execution_outcome_receipts execution_outcome_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.execution_outcome_receipts
    ADD CONSTRAINT execution_outcome_fk FOREIGN KEY (executed_receipt_id) REFERENCES public.execution_outcomes(receipt_id) ON DELETE CASCADE;


--
-- Name: execution_outcomes receipt_execution_outcome_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.execution_outcomes
    ADD CONSTRAINT receipt_execution_outcome_fk FOREIGN KEY (receipt_id) REFERENCES public.receipts(receipt_id) ON DELETE CASCADE;


--
-- Name: data_receipts receipt_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.data_receipts
    ADD CONSTRAINT receipt_fk FOREIGN KEY (receipt_id) REFERENCES public.receipts(receipt_id) ON DELETE CASCADE;


--
-- Name: action_receipts receipt_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.action_receipts
    ADD CONSTRAINT receipt_fk FOREIGN KEY (receipt_id) REFERENCES public.receipts(receipt_id) ON DELETE CASCADE;


--
-- Name: action_receipt_output_data receipt_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.action_receipt_output_data
    ADD CONSTRAINT receipt_fk FOREIGN KEY (output_from_receipt_id) REFERENCES public.receipts(receipt_id) ON DELETE CASCADE;


--
-- Name: action_receipt_input_data receipt_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.action_receipt_input_data
    ADD CONSTRAINT receipt_fk FOREIGN KEY (input_to_receipt_id) REFERENCES public.receipts(receipt_id) ON DELETE CASCADE;


--
-- Name: account_changes receipt_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.account_changes
    ADD CONSTRAINT receipt_id_fk FOREIGN KEY (caused_by_receipt_id) REFERENCES public.receipts(receipt_id) ON DELETE CASCADE;


--
-- Name: execution_outcome_receipts receipts_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.execution_outcome_receipts
    ADD CONSTRAINT receipts_fk FOREIGN KEY (executed_receipt_id) REFERENCES public.receipts(receipt_id) ON DELETE CASCADE;


--
-- Name: account_changes transaction_hash_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.account_changes
    ADD CONSTRAINT transaction_hash_fk FOREIGN KEY (caused_by_transaction_hash) REFERENCES public.transactions(transaction_hash) ON DELETE CASCADE;


--
-- Name: transaction_actions tx_action_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.transaction_actions
    ADD CONSTRAINT tx_action_fk FOREIGN KEY (transaction_hash) REFERENCES public.transactions(transaction_hash) ON DELETE CASCADE;


--
-- Name: receipts tx_receipt_fk; Type: FK CONSTRAINT; Schema: public; Owner: testnet
--

ALTER TABLE ONLY public.receipts
    ADD CONSTRAINT tx_receipt_fk FOREIGN KEY (originated_from_transaction_hash) REFERENCES public.transactions(transaction_hash) ON DELETE CASCADE;


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: cloudsqlsuperuser
--

REVOKE ALL ON SCHEMA public FROM cloudsqladmin;
REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO cloudsqlsuperuser;
GRANT USAGE ON SCHEMA public TO PUBLIC;
GRANT USAGE ON SCHEMA public TO explorer;
GRANT USAGE ON SCHEMA public TO wallet;
GRANT USAGE ON SCHEMA public TO jupyter;
GRANT USAGE ON SCHEMA public TO robertyan;
GRANT USAGE ON SCHEMA public TO public_readonly;
GRANT USAGE ON SCHEMA public TO readonly;


--
-- Name: TABLE __diesel_schema_migrations; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.__diesel_schema_migrations TO explorer;
GRANT SELECT ON TABLE public.__diesel_schema_migrations TO wallet;
GRANT SELECT ON TABLE public.__diesel_schema_migrations TO jupyter;
GRANT SELECT ON TABLE public.__diesel_schema_migrations TO robertyan;
GRANT SELECT ON TABLE public.__diesel_schema_migrations TO partner_token_terminal;
GRANT SELECT ON TABLE public.__diesel_schema_migrations TO public_readonly;
GRANT SELECT ON TABLE public.__diesel_schema_migrations TO readonly;


--
-- Name: TABLE access_keys; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.access_keys TO explorer;
GRANT SELECT ON TABLE public.access_keys TO wallet;
GRANT SELECT ON TABLE public.access_keys TO jupyter;
GRANT SELECT ON TABLE public.access_keys TO robertyan;
GRANT SELECT ON TABLE public.access_keys TO partner_token_terminal;
GRANT SELECT ON TABLE public.access_keys TO public_readonly;
GRANT SELECT ON TABLE public.access_keys TO readonly;


--
-- Name: TABLE account_changes; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.account_changes TO explorer;
GRANT SELECT ON TABLE public.account_changes TO wallet;
GRANT SELECT ON TABLE public.account_changes TO jupyter;
GRANT SELECT ON TABLE public.account_changes TO robertyan;
GRANT SELECT ON TABLE public.account_changes TO partner_token_terminal;
GRANT SELECT ON TABLE public.account_changes TO public_readonly;
GRANT SELECT ON TABLE public.account_changes TO readonly;


--
-- Name: TABLE accounts; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.accounts TO explorer;
GRANT SELECT ON TABLE public.accounts TO wallet;
GRANT SELECT ON TABLE public.accounts TO jupyter;
GRANT SELECT ON TABLE public.accounts TO robertyan;
GRANT SELECT ON TABLE public.accounts TO partner_token_terminal;
GRANT SELECT ON TABLE public.accounts TO public_readonly;
GRANT SELECT ON TABLE public.accounts TO readonly;


--
-- Name: TABLE action_receipt_actions; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.action_receipt_actions TO explorer;
GRANT SELECT ON TABLE public.action_receipt_actions TO wallet;
GRANT SELECT ON TABLE public.action_receipt_actions TO jupyter;
GRANT SELECT ON TABLE public.action_receipt_actions TO robertyan;
GRANT SELECT ON TABLE public.action_receipt_actions TO partner_token_terminal;
GRANT SELECT ON TABLE public.action_receipt_actions TO public_readonly;
GRANT SELECT ON TABLE public.action_receipt_actions TO readonly;


--
-- Name: TABLE action_receipt_input_data; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.action_receipt_input_data TO explorer;
GRANT SELECT ON TABLE public.action_receipt_input_data TO wallet;
GRANT SELECT ON TABLE public.action_receipt_input_data TO jupyter;
GRANT SELECT ON TABLE public.action_receipt_input_data TO robertyan;
GRANT SELECT ON TABLE public.action_receipt_input_data TO partner_token_terminal;
GRANT SELECT ON TABLE public.action_receipt_input_data TO public_readonly;
GRANT SELECT ON TABLE public.action_receipt_input_data TO readonly;


--
-- Name: TABLE action_receipt_output_data; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.action_receipt_output_data TO explorer;
GRANT SELECT ON TABLE public.action_receipt_output_data TO wallet;
GRANT SELECT ON TABLE public.action_receipt_output_data TO jupyter;
GRANT SELECT ON TABLE public.action_receipt_output_data TO robertyan;
GRANT SELECT ON TABLE public.action_receipt_output_data TO partner_token_terminal;
GRANT SELECT ON TABLE public.action_receipt_output_data TO public_readonly;
GRANT SELECT ON TABLE public.action_receipt_output_data TO readonly;


--
-- Name: TABLE action_receipts; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.action_receipts TO explorer;
GRANT SELECT ON TABLE public.action_receipts TO wallet;
GRANT SELECT ON TABLE public.action_receipts TO jupyter;
GRANT SELECT ON TABLE public.action_receipts TO robertyan;
GRANT SELECT ON TABLE public.action_receipts TO partner_token_terminal;
GRANT SELECT ON TABLE public.action_receipts TO public_readonly;
GRANT SELECT ON TABLE public.action_receipts TO readonly;


--
-- Name: TABLE aggregated__circulating_supply; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.aggregated__circulating_supply TO readonly;


--
-- Name: TABLE blocks; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.blocks TO explorer;
GRANT SELECT ON TABLE public.blocks TO wallet;
GRANT SELECT ON TABLE public.blocks TO jupyter;
GRANT SELECT ON TABLE public.blocks TO robertyan;
GRANT SELECT ON TABLE public.blocks TO partner_token_terminal;
GRANT SELECT ON TABLE public.blocks TO public_readonly;
GRANT SELECT ON TABLE public.blocks TO readonly;


--
-- Name: TABLE receipts; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.receipts TO explorer;
GRANT SELECT ON TABLE public.receipts TO wallet;
GRANT SELECT ON TABLE public.receipts TO jupyter;
GRANT SELECT ON TABLE public.receipts TO robertyan;
GRANT SELECT ON TABLE public.receipts TO partner_token_terminal;
GRANT SELECT ON TABLE public.receipts TO public_readonly;
GRANT SELECT ON TABLE public.receipts TO readonly;


--
-- Name: TABLE aggregated__lockups; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.aggregated__lockups TO readonly;


--
-- Name: TABLE chunks; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.chunks TO explorer;
GRANT SELECT ON TABLE public.chunks TO wallet;
GRANT SELECT ON TABLE public.chunks TO jupyter;
GRANT SELECT ON TABLE public.chunks TO robertyan;
GRANT SELECT ON TABLE public.chunks TO partner_token_terminal;
GRANT SELECT ON TABLE public.chunks TO public_readonly;
GRANT SELECT ON TABLE public.chunks TO readonly;


--
-- Name: TABLE data_receipts; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.data_receipts TO explorer;
GRANT SELECT ON TABLE public.data_receipts TO wallet;
GRANT SELECT ON TABLE public.data_receipts TO jupyter;
GRANT SELECT ON TABLE public.data_receipts TO robertyan;
GRANT SELECT ON TABLE public.data_receipts TO partner_token_terminal;
GRANT SELECT ON TABLE public.data_receipts TO public_readonly;
GRANT SELECT ON TABLE public.data_receipts TO readonly;


--
-- Name: TABLE execution_outcome_receipts; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.execution_outcome_receipts TO explorer;
GRANT SELECT ON TABLE public.execution_outcome_receipts TO wallet;
GRANT SELECT ON TABLE public.execution_outcome_receipts TO jupyter;
GRANT SELECT ON TABLE public.execution_outcome_receipts TO robertyan;
GRANT SELECT ON TABLE public.execution_outcome_receipts TO partner_token_terminal;
GRANT SELECT ON TABLE public.execution_outcome_receipts TO public_readonly;
GRANT SELECT ON TABLE public.execution_outcome_receipts TO readonly;


--
-- Name: TABLE execution_outcomes; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.execution_outcomes TO explorer;
GRANT SELECT ON TABLE public.execution_outcomes TO wallet;
GRANT SELECT ON TABLE public.execution_outcomes TO jupyter;
GRANT SELECT ON TABLE public.execution_outcomes TO robertyan;
GRANT SELECT ON TABLE public.execution_outcomes TO partner_token_terminal;
GRANT SELECT ON TABLE public.execution_outcomes TO public_readonly;
GRANT SELECT ON TABLE public.execution_outcomes TO readonly;


--
-- Name: TABLE transaction_actions; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.transaction_actions TO explorer;
GRANT SELECT ON TABLE public.transaction_actions TO wallet;
GRANT SELECT ON TABLE public.transaction_actions TO jupyter;
GRANT SELECT ON TABLE public.transaction_actions TO robertyan;
GRANT SELECT ON TABLE public.transaction_actions TO partner_token_terminal;
GRANT SELECT ON TABLE public.transaction_actions TO public_readonly;
GRANT SELECT ON TABLE public.transaction_actions TO readonly;


--
-- Name: TABLE transactions; Type: ACL; Schema: public; Owner: testnet
--

GRANT SELECT ON TABLE public.transactions TO explorer;
GRANT SELECT ON TABLE public.transactions TO wallet;
GRANT SELECT ON TABLE public.transactions TO jupyter;
GRANT SELECT ON TABLE public.transactions TO robertyan;
GRANT SELECT ON TABLE public.transactions TO partner_token_terminal;
GRANT SELECT ON TABLE public.transactions TO public_readonly;
GRANT SELECT ON TABLE public.transactions TO readonly;


--
-- Name: DEFAULT PRIVILEGES FOR TABLES; Type: DEFAULT ACL; Schema: public; Owner: testnet
--

ALTER DEFAULT PRIVILEGES FOR ROLE testnet IN SCHEMA public REVOKE ALL ON TABLES  FROM testnet;
ALTER DEFAULT PRIVILEGES FOR ROLE testnet IN SCHEMA public GRANT SELECT ON TABLES  TO readonly;


--
-- PostgreSQL database dump complete
--

