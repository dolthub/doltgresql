-- Downloaded from: https://github.com/HarukaMa/aleo-explorer/blob/d3a313bae7ece2fc3c4e7abddeeb47d3bca52d5c/pg_dump.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 15.4 (Debian 15.4-3)
-- Dumped by pg_dump version 16.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', 'explorer', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: explorer; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA explorer;


--
-- Name: argument_type; Type: TYPE; Schema: explorer; Owner: -
--

CREATE TYPE explorer.argument_type AS ENUM (
    'Plaintext',
    'Future'
);


--
-- Name: authority_type; Type: TYPE; Schema: explorer; Owner: -
--

CREATE TYPE explorer.authority_type AS ENUM (
    'Beacon',
    'Quorum'
);


--
-- Name: confirmed_transaction_type; Type: TYPE; Schema: explorer; Owner: -
--

CREATE TYPE explorer.confirmed_transaction_type AS ENUM (
    'AcceptedDeploy',
    'AcceptedExecute',
    'RejectedDeploy',
    'RejectedExecute'
);


--
-- Name: finalize_operation_type; Type: TYPE; Schema: explorer; Owner: -
--

CREATE TYPE explorer.finalize_operation_type AS ENUM (
    'InitializeMapping',
    'InsertKeyValue',
    'UpdateKeyValue',
    'RemoveKeyValue',
    'ReplaceMapping',
    'RemoveMapping'
);


--
-- Name: future_type; Type: TYPE; Schema: explorer; Owner: -
--

CREATE TYPE explorer.future_type AS ENUM (
    'Output',
    'Argument'
);


--
-- Name: ratification_type; Type: TYPE; Schema: explorer; Owner: -
--

CREATE TYPE explorer.ratification_type AS ENUM (
    'Genesis',
    'BlockReward',
    'PuzzleReward'
);


--
-- Name: transaction_type; Type: TYPE; Schema: explorer; Owner: -
--

CREATE TYPE explorer.transaction_type AS ENUM (
    'Deploy',
    'Execute',
    'Fee'
);


--
-- Name: transition_data_type; Type: TYPE; Schema: explorer; Owner: -
--

CREATE TYPE explorer.transition_data_type AS ENUM (
    'Constant',
    'Public',
    'Private',
    'Record',
    'ExternalRecord',
    'Future'
);


--
-- Name: transmission_id_type; Type: TYPE; Schema: explorer; Owner: -
--

CREATE TYPE explorer.transmission_id_type AS ENUM (
    'Ratification',
    'Solution',
    'Transaction'
);


--
-- Name: get_block_target_sum(bigint); Type: FUNCTION; Schema: explorer; Owner: -
--

CREATE FUNCTION explorer.get_block_target_sum(block_height bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
SELECT SUM(target) FROM explorer.partial_solution ps
JOIN explorer.coinbase_solution cs ON cs.id = ps.coinbase_solution_id
JOIN explorer.block b ON b.id = cs.block_id
WHERE height = block_height
$$;


--
-- Name: get_confirmed_transactions(integer); Type: FUNCTION; Schema: explorer; Owner: -
--

CREATE FUNCTION explorer.get_confirmed_transactions(block_db_id integer) RETURNS TABLE(confirmed_transaction_id integer, confirmed_transaction_type explorer.confirmed_transaction_type, index integer, reject_reason text, transaction_id text, transaction_type explorer.transaction_type, transaction_deploy_id integer, edition integer, verifying_keys bytea, program_id text, owner text, transaction_execute_id integer, global_state_root text, proof text, fee_id integer, fee_global_state_root text, fee_proof text)
    LANGUAGE plpgsql
    AS $$
declare
    transaction_db_id transaction.id%type;
    confirmed_transaction_db_id confirmed_transaction.id%type;
begin
    for confirmed_transaction_db_id, confirmed_transaction_type, index, reject_reason in
        select t.id, t.type, t.index, t.reject_reason from confirmed_transaction t where t.block_id = block_db_id
        loop
            confirmed_transaction_id := confirmed_transaction_db_id;
            select t.id, t.transaction_id, t.type from transaction t where t.confirmed_transaction_id = confirmed_transaction_db_id into transaction_db_id, transaction_id, transaction_type;
            if confirmed_transaction_type = 'AcceptedDeploy' or confirmed_transaction_type = 'RejectedDeploy' then
                if confirmed_transaction_type = 'RejectedDeploy' then
                    select t.id, t.edition, t.verifying_keys, t.program_id, t.owner from transaction_deploy t where t.transaction_id = transaction_db_id order by id limit 1 into transaction_deploy_id, edition, verifying_keys, program_id, owner;
                else
                    select t.id, t.edition, t.verifying_keys from transaction_deploy t where t.transaction_id = transaction_db_id into transaction_deploy_id, edition, verifying_keys;
                end if;
                select t.id, t.global_state_root, t.proof from fee t where t.transaction_id = transaction_db_id order by id limit 1 into fee_id, fee_global_state_root, fee_proof;
                return next;
            elsif confirmed_transaction_type = 'AcceptedExecute' or confirmed_transaction_type = 'RejectedExecute' then
                select t.id, t.global_state_root, t.proof from transaction_execute t where t.transaction_id = transaction_db_id order by id limit 1 into transaction_execute_id, global_state_root, proof;
                select t.id, t.global_state_root, t.proof from fee t where t.transaction_id = transaction_db_id order by id limit 1 into fee_id, fee_global_state_root, fee_proof;
                return next;
            end if;
        end loop;
end;
$$;


--
-- Name: get_finalize_operations(integer); Type: FUNCTION; Schema: explorer; Owner: -
--

CREATE FUNCTION explorer.get_finalize_operations(confirmed_transaction_db_id integer) RETURNS TABLE(type explorer.finalize_operation_type, index integer, mapping_id text, key_id text, value_id text)
    LANGUAGE plpgsql
    AS $$
declare
    finalize_operation_db_id finalize_operation.id%type;
begin
    for finalize_operation_db_id, type, index in
        select t.id, t.type, t.index from finalize_operation t where t.confirmed_transaction_id = confirmed_transaction_db_id order by id
        loop
            if type = 'InitializeMapping' then
                select t.mapping_id from finalize_operation_initialize_mapping t where t.finalize_operation_id = finalize_operation_db_id into mapping_id;
                return next;
            elsif type = 'InsertKeyValue' then
                select t.mapping_id, t.key_id, t.value_id from finalize_operation_insert_kv t where t.finalize_operation_id = finalize_operation_db_id into mapping_id, key_id, value_id;
                return next;
            elsif type = 'UpdateKeyValue' then
                select t.mapping_id, t.key_id, t.value_id from finalize_operation_update_kv t where t.finalize_operation_id = finalize_operation_db_id into mapping_id, key_id, value_id;
                return next;
            elsif type = 'RemoveKeyValue' then
                select t.mapping_id, t.key_id from finalize_operation_remove_kv t where t.finalize_operation_id = finalize_operation_db_id into mapping_id, key_id;
                return next;
            elsif type = 'RemoveMapping' then
                select t.mapping_id from finalize_operation_remove_mapping t where t.finalize_operation_id = finalize_operation_db_id into mapping_id;
                return next;
            else
                raise exception 'unsupported finalize operation type: %', type;
            end if;
        end loop;
end;
$$;


--
-- Name: get_transition_inputs(integer); Type: FUNCTION; Schema: explorer; Owner: -
--

CREATE FUNCTION explorer.get_transition_inputs(transition_db_id integer) RETURNS TABLE(type explorer.transition_data_type, index integer, plaintext_hash text, plaintext bytea, ciphertext_hash text, ciphertext text, serial_number text, tag text, commitment text)
    LANGUAGE plpgsql
    AS $$
declare
    transition_input_db_id transition_input.id%type;
begin
    for transition_input_db_id, type, index in
        select id, t.type, t.index from transition_input t where transition_id = transition_db_id order by id
        loop
            if type = 'Public' then
                select t.plaintext_hash, t.plaintext from transition_input_public t where transition_input_id = transition_input_db_id into plaintext_hash, plaintext;
                return next;
            elsif type = 'Private' then
                select t.ciphertext_hash, t.ciphertext from transition_input_private t where transition_input_id = transition_input_db_id into ciphertext_hash, ciphertext;
                return next;
            elsif type = 'Record' then
                select t.serial_number, t.tag from transition_input_record t where transition_input_id = transition_input_db_id into serial_number, tag;
                return next;
            elsif type = 'ExternalRecord' then
                select t.commitment from transition_input_external_record t where transition_input_id = transition_input_db_id into commitment;
                return next;
            else
                raise exception 'unsupported transition input type: %', type;
            end if;
        end loop;
end;
$$;


--
-- Name: get_transition_outputs(integer); Type: FUNCTION; Schema: explorer; Owner: -
--

CREATE FUNCTION explorer.get_transition_outputs(transition_db_id integer) RETURNS TABLE(type explorer.transition_data_type, index integer, plaintext_hash text, plaintext bytea, ciphertext_hash text, ciphertext text, record_commitment text, checksum text, record_ciphertext text, external_record_commitment text, future_id integer, future_hash text)
    LANGUAGE plpgsql
    AS $$
declare
    transition_output_db_id transition_output.id%type;
begin
    for transition_output_db_id, type, index in
        select id, t.type, t.index from transition_output t where transition_id = transition_db_id order by id
        loop
            if type = 'Public' then
                select t.plaintext_hash, t.plaintext from transition_output_public t where transition_output_id = transition_output_db_id into plaintext_hash, plaintext;
                return next;
            elsif type = 'Private' then
                select t.ciphertext_hash, t.ciphertext from transition_output_private t where transition_output_id = transition_output_db_id into ciphertext_hash, ciphertext;
                return next;
            elsif type = 'Record' then
                select t.commitment, t.checksum, t.record_ciphertext from transition_output_record t where transition_output_id = transition_output_db_id into record_commitment, checksum, record_ciphertext;
                return next;
            elsif type = 'ExternalRecord' then
                select t.commitment from transition_output_external_record t where transition_output_id = transition_output_db_id into external_record_commitment;
                return next;
            elsif type = 'Future' then
                select t.id, t.future_hash from transition_output_future t where transition_output_id = transition_output_db_id into future_id, future_hash;
                return next;
            else
                raise exception 'unsupported transition output type: %', type;
            end if;
        end loop;
end;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: _dirty_flag; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer._dirty_flag (
    dirty boolean DEFAULT false NOT NULL
);


--
-- Name: _migration; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer._migration (
    migrated_id integer NOT NULL
);


--
-- Name: address_fee_history; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.address_fee_history (
    id bigint NOT NULL,
    height integer NOT NULL,
    address text NOT NULL,
    fee numeric(20,0) NOT NULL,
    previous_id bigint
);


--
-- Name: address_fee_history_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.address_fee_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: address_fee_history_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.address_fee_history_id_seq OWNED BY explorer.address_fee_history.id;


--
-- Name: address_fee_history_last_id; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.address_fee_history_last_id (
    address text NOT NULL,
    last_history_id bigint
);


--
-- Name: address_puzzle_reward_history; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.address_puzzle_reward_history (
    id bigint NOT NULL,
    height integer NOT NULL,
    address text NOT NULL,
    puzzle_reward numeric(20,0) NOT NULL,
    previous_id bigint
);


--
-- Name: address_puzzle_reward_history_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.address_puzzle_reward_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: address_puzzle_reward_history_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.address_puzzle_reward_history_id_seq OWNED BY explorer.address_puzzle_reward_history.id;


--
-- Name: address_puzzle_reward_history_last_id; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.address_puzzle_reward_history_last_id (
    address text NOT NULL,
    last_history_id bigint NOT NULL
);


--
-- Name: address_stake_reward; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.address_stake_reward (
    id integer NOT NULL,
    address text NOT NULL,
    stake_reward numeric(20,0) NOT NULL
);


--
-- Name: address_stake_reward_history; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.address_stake_reward_history (
    id bigint NOT NULL,
    height integer NOT NULL,
    content jsonb NOT NULL
);


--
-- Name: address_stake_reward_history_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.address_stake_reward_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: address_stake_reward_history_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.address_stake_reward_history_id_seq OWNED BY explorer.address_stake_reward_history.id;


--
-- Name: address_stake_reward_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.address_stake_reward_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: address_stake_reward_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.address_stake_reward_id_seq OWNED BY explorer.address_stake_reward.id;


--
-- Name: address_tag; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.address_tag (
    id integer NOT NULL,
    address text NOT NULL,
    tag text NOT NULL
);


--
-- Name: address_tag_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.address_tag_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: address_tag_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.address_tag_id_seq OWNED BY explorer.address_tag.id;


--
-- Name: address_transfer_in_history; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.address_transfer_in_history (
    id bigint NOT NULL,
    height integer NOT NULL,
    address text NOT NULL,
    transfer_in numeric(20,0) NOT NULL,
    previous_id bigint
);


--
-- Name: address_transfer_in_history_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.address_transfer_in_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: address_transfer_in_history_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.address_transfer_in_history_id_seq OWNED BY explorer.address_transfer_in_history.id;


--
-- Name: address_transfer_in_history_last_id; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.address_transfer_in_history_last_id (
    address text NOT NULL,
    last_history_id bigint NOT NULL
);


--
-- Name: address_transfer_out_history; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.address_transfer_out_history (
    id bigint NOT NULL,
    height integer NOT NULL,
    address text NOT NULL,
    transfer_out numeric(20,0) NOT NULL,
    previous_id bigint
);


--
-- Name: address_transfer_out_history_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.address_transfer_out_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: address_transfer_out_history_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.address_transfer_out_history_id_seq OWNED BY explorer.address_transfer_out_history.id;


--
-- Name: address_transfer_out_history_last_id; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.address_transfer_out_history_last_id (
    address text NOT NULL,
    last_history_id bigint NOT NULL
);


--
-- Name: address_transition; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.address_transition (
    address text NOT NULL,
    transition_id integer NOT NULL
);


--
-- Name: authority; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.authority (
    id integer NOT NULL,
    block_id integer NOT NULL,
    type explorer.authority_type NOT NULL,
    signature text
);


--
-- Name: authority_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.authority_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: authority_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.authority_id_seq OWNED BY explorer.authority.id;


--
-- Name: block; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.block (
    id integer NOT NULL,
    height bigint NOT NULL,
    block_hash text NOT NULL,
    previous_hash text NOT NULL,
    previous_state_root text NOT NULL,
    transactions_root text NOT NULL,
    finalize_root text NOT NULL,
    ratifications_root text NOT NULL,
    solutions_root text NOT NULL,
    subdag_root text NOT NULL,
    round numeric(20,0) NOT NULL,
    cumulative_weight numeric(40,0) NOT NULL,
    cumulative_proof_target numeric(40,0) NOT NULL,
    coinbase_target numeric(20,0) NOT NULL,
    proof_target numeric(20,0) NOT NULL,
    last_coinbase_target numeric(20,0) NOT NULL,
    last_coinbase_timestamp bigint NOT NULL,
    "timestamp" bigint NOT NULL,
    block_reward numeric(20,0) NOT NULL,
    coinbase_reward numeric(20,0) NOT NULL,
    total_supply numeric(40,0) NOT NULL,
    confirm_timestamp bigint NOT NULL
);


--
-- Name: block_aborted_solution_id; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.block_aborted_solution_id (
    id integer NOT NULL,
    block_id integer NOT NULL,
    solution_id text NOT NULL
);


--
-- Name: block_aborted_solution_id_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.block_aborted_solution_id_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: block_aborted_solution_id_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.block_aborted_solution_id_id_seq OWNED BY explorer.block_aborted_solution_id.id;


--
-- Name: block_aborted_transaction_id; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.block_aborted_transaction_id (
    id integer NOT NULL,
    block_id integer NOT NULL,
    transaction_id text NOT NULL
);


--
-- Name: block_aborted_transaction_id_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.block_aborted_transaction_id_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: block_aborted_transaction_id_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.block_aborted_transaction_id_id_seq OWNED BY explorer.block_aborted_transaction_id.id;


--
-- Name: block_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.block_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: block_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.block_id_seq OWNED BY explorer.block.id;


--
-- Name: block_validator; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.block_validator (
    id bigint NOT NULL,
    block_id integer NOT NULL,
    validator text NOT NULL
);


--
-- Name: block_validator_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.block_validator_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: block_validator_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.block_validator_id_seq OWNED BY explorer.block_validator.id;


--
-- Name: puzzle_solution; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.puzzle_solution (
    id integer NOT NULL,
    block_id integer NOT NULL,
    target_sum numeric(20,0) DEFAULT 0 NOT NULL
);


--
-- Name: coinbase_solution_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.coinbase_solution_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: coinbase_solution_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.coinbase_solution_id_seq OWNED BY explorer.puzzle_solution.id;


--
-- Name: committee_history; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.committee_history (
    id integer NOT NULL,
    height bigint NOT NULL,
    starting_round numeric(20,0) NOT NULL,
    total_stake numeric(20,0) NOT NULL,
    committee_id text NOT NULL
);


--
-- Name: committee_history_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.committee_history_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: committee_history_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.committee_history_id_seq OWNED BY explorer.committee_history.id;


--
-- Name: committee_history_member; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.committee_history_member (
    id integer NOT NULL,
    committee_id integer NOT NULL,
    address text NOT NULL,
    stake numeric(20,0) NOT NULL,
    is_open boolean NOT NULL,
    commission integer NOT NULL
);


--
-- Name: committee_history_member_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.committee_history_member_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: committee_history_member_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.committee_history_member_id_seq OWNED BY explorer.committee_history_member.id;


--
-- Name: confirmed_transaction; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.confirmed_transaction (
    id integer NOT NULL,
    block_id integer,
    index integer NOT NULL,
    type explorer.confirmed_transaction_type NOT NULL,
    reject_reason text
);


--
-- Name: confirmed_transaction_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.confirmed_transaction_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: confirmed_transaction_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.confirmed_transaction_id_seq OWNED BY explorer.confirmed_transaction.id;


--
-- Name: dag_vertex; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.dag_vertex (
    id bigint NOT NULL,
    authority_id integer NOT NULL,
    round numeric(20,0) NOT NULL,
    batch_id text NOT NULL,
    author text NOT NULL,
    "timestamp" bigint NOT NULL,
    author_signature text NOT NULL,
    index integer NOT NULL,
    committee_id text NOT NULL
);


--
-- Name: dag_vertex_adjacency; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.dag_vertex_adjacency (
    id bigint NOT NULL,
    vertex_id bigint NOT NULL,
    previous_vertex_id bigint NOT NULL,
    index integer NOT NULL
);


--
-- Name: dag_vertex_adjacency_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.dag_vertex_adjacency_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: dag_vertex_adjacency_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.dag_vertex_adjacency_id_seq OWNED BY explorer.dag_vertex_adjacency.id;


--
-- Name: dag_vertex_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.dag_vertex_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: dag_vertex_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.dag_vertex_id_seq OWNED BY explorer.dag_vertex.id;


--
-- Name: dag_vertex_signature; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.dag_vertex_signature (
    id bigint NOT NULL,
    vertex_id bigint NOT NULL,
    signature text NOT NULL,
    "timestamp" bigint,
    index integer NOT NULL
);


--
-- Name: dag_vertex_signature_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.dag_vertex_signature_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: dag_vertex_signature_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.dag_vertex_signature_id_seq OWNED BY explorer.dag_vertex_signature.id;


--
-- Name: dag_vertex_transmission_id; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.dag_vertex_transmission_id (
    id bigint NOT NULL,
    vertex_id bigint NOT NULL,
    type explorer.transmission_id_type NOT NULL,
    index integer NOT NULL,
    commitment text,
    transaction_id text
);


--
-- Name: dag_vertex_transmission_id_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.dag_vertex_transmission_id_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: dag_vertex_transmission_id_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.dag_vertex_transmission_id_id_seq OWNED BY explorer.dag_vertex_transmission_id.id;


--
-- Name: transaction_execute; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.transaction_execute (
    id integer NOT NULL,
    transaction_id integer NOT NULL,
    global_state_root text NOT NULL,
    proof text
);


--
-- Name: execute_transaction_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.execute_transaction_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: execute_transaction_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.execute_transaction_id_seq OWNED BY explorer.transaction_execute.id;


--
-- Name: fee; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.fee (
    id integer NOT NULL,
    transaction_id integer NOT NULL,
    global_state_root text NOT NULL,
    proof text
);


--
-- Name: fee_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.fee_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: fee_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.fee_id_seq OWNED BY explorer.fee.id;


--
-- Name: feedback; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.feedback (
    id integer NOT NULL,
    contact text NOT NULL,
    content text NOT NULL,
    "timestamp" bigint DEFAULT EXTRACT(epoch FROM now()) NOT NULL
);


--
-- Name: feedback_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.feedback_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: feedback_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.feedback_id_seq OWNED BY explorer.feedback.id;


--
-- Name: finalize_operation; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.finalize_operation (
    id integer NOT NULL,
    confirmed_transaction_id integer NOT NULL,
    type explorer.finalize_operation_type NOT NULL,
    index integer NOT NULL
);


--
-- Name: finalize_operation_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.finalize_operation_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: finalize_operation_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.finalize_operation_id_seq OWNED BY explorer.finalize_operation.id;


--
-- Name: finalize_operation_initialize_mapping; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.finalize_operation_initialize_mapping (
    id integer NOT NULL,
    finalize_operation_id integer NOT NULL,
    mapping_id text NOT NULL
);


--
-- Name: finalize_operation_initialize_mapping_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.finalize_operation_initialize_mapping_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: finalize_operation_initialize_mapping_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.finalize_operation_initialize_mapping_id_seq OWNED BY explorer.finalize_operation_initialize_mapping.id;


--
-- Name: finalize_operation_insert_kv; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.finalize_operation_insert_kv (
    id integer NOT NULL,
    finalize_operation_id integer NOT NULL,
    mapping_id text NOT NULL,
    key_id text NOT NULL,
    value_id text NOT NULL
);


--
-- Name: finalize_operation_insert_kv_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.finalize_operation_insert_kv_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: finalize_operation_insert_kv_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.finalize_operation_insert_kv_id_seq OWNED BY explorer.finalize_operation_insert_kv.id;


--
-- Name: finalize_operation_remove_kv; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.finalize_operation_remove_kv (
    id integer NOT NULL,
    finalize_operation_id integer NOT NULL,
    mapping_id text NOT NULL,
    key_id text NOT NULL
);


--
-- Name: finalize_operation_remove_kv_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.finalize_operation_remove_kv_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: finalize_operation_remove_kv_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.finalize_operation_remove_kv_id_seq OWNED BY explorer.finalize_operation_remove_kv.id;


--
-- Name: finalize_operation_remove_mapping; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.finalize_operation_remove_mapping (
    id integer NOT NULL,
    finalize_operation_id integer NOT NULL,
    mapping_id text NOT NULL
);


--
-- Name: finalize_operation_remove_mapping_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.finalize_operation_remove_mapping_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: finalize_operation_remove_mapping_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.finalize_operation_remove_mapping_id_seq OWNED BY explorer.finalize_operation_remove_mapping.id;


--
-- Name: finalize_operation_replace_mapping; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.finalize_operation_replace_mapping (
    id integer NOT NULL,
    finalize_operation_id integer NOT NULL,
    mapping_id text NOT NULL
);


--
-- Name: finalize_operation_replace_mapping_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.finalize_operation_replace_mapping_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: finalize_operation_replace_mapping_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.finalize_operation_replace_mapping_id_seq OWNED BY explorer.finalize_operation_replace_mapping.id;


--
-- Name: finalize_operation_update_kv; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.finalize_operation_update_kv (
    id integer NOT NULL,
    finalize_operation_id integer NOT NULL,
    mapping_id text NOT NULL,
    key_id text NOT NULL,
    value_id text NOT NULL
);


--
-- Name: finalize_operation_update_kv_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.finalize_operation_update_kv_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: finalize_operation_update_kv_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.finalize_operation_update_kv_id_seq OWNED BY explorer.finalize_operation_update_kv.id;


--
-- Name: future; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.future (
    id integer NOT NULL,
    type explorer.future_type NOT NULL,
    transition_output_future_id integer,
    future_argument_id integer,
    program_id text NOT NULL,
    function_name text NOT NULL
);


--
-- Name: future_argument; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.future_argument (
    id integer NOT NULL,
    future_id integer NOT NULL,
    type explorer.argument_type NOT NULL,
    plaintext bytea
);


--
-- Name: future_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.future_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: future_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.future_id_seq OWNED BY explorer.future.id;


--
-- Name: mapping; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.mapping (
    id integer NOT NULL,
    mapping_id text NOT NULL,
    program_id text NOT NULL,
    mapping text NOT NULL
);


--
-- Name: mapping_bonded_history; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.mapping_bonded_history (
    id integer NOT NULL,
    height bigint NOT NULL,
    content jsonb NOT NULL
);


--
-- Name: mapping_bonded_history_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.mapping_bonded_history_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: mapping_bonded_history_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.mapping_bonded_history_id_seq OWNED BY explorer.mapping_bonded_history.id;


--
-- Name: mapping_bonded_value; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.mapping_bonded_value (
    id integer NOT NULL,
    key_id text NOT NULL,
    key bytea NOT NULL,
    value bytea NOT NULL
);


--
-- Name: mapping_bonded_value_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.mapping_bonded_value_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: mapping_bonded_value_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.mapping_bonded_value_id_seq OWNED BY explorer.mapping_bonded_value.id;


--
-- Name: mapping_committee_history; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.mapping_committee_history (
    id integer NOT NULL,
    height bigint NOT NULL,
    content jsonb NOT NULL
);


--
-- Name: mapping_committee_history_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.mapping_committee_history_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: mapping_committee_history_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.mapping_committee_history_id_seq OWNED BY explorer.mapping_committee_history.id;


--
-- Name: mapping_delegated_history; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.mapping_delegated_history (
    id integer NOT NULL,
    height bigint NOT NULL,
    content jsonb NOT NULL
);


--
-- Name: mapping_delegated_history_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.mapping_delegated_history_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: mapping_delegated_history_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.mapping_delegated_history_id_seq OWNED BY explorer.mapping_delegated_history.id;


--
-- Name: mapping_history; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.mapping_history (
    id bigint NOT NULL,
    mapping_id integer NOT NULL,
    height integer NOT NULL,
    key_id text NOT NULL,
    key bytea NOT NULL,
    value bytea,
    from_transaction boolean NOT NULL,
    previous_id bigint
);
ALTER TABLE ONLY explorer.mapping_history ALTER COLUMN key_id SET STATISTICS 10000;


--
-- Name: mapping_history_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.mapping_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: mapping_history_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.mapping_history_id_seq OWNED BY explorer.mapping_history.id;


--
-- Name: mapping_history_last_id; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.mapping_history_last_id (
    key_id text NOT NULL,
    last_history_id bigint NOT NULL
);


--
-- Name: mapping_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.mapping_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: mapping_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.mapping_id_seq OWNED BY explorer.mapping.id;


--
-- Name: mapping_value; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.mapping_value (
    id integer NOT NULL,
    mapping_id integer NOT NULL,
    key_id text NOT NULL,
    value_id text NOT NULL,
    key bytea NOT NULL,
    value bytea NOT NULL
);


--
-- Name: mapping_value_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.mapping_value_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: mapping_value_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.mapping_value_id_seq OWNED BY explorer.mapping_value.id;


--
-- Name: solution; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.solution (
    id bigint NOT NULL,
    puzzle_solution_id integer NOT NULL,
    address text NOT NULL,
    counter numeric(20,0) NOT NULL,
    target numeric(20,0) NOT NULL,
    reward integer NOT NULL,
    epoch_hash text NOT NULL,
    solution_id text
);


--
-- Name: partial_solution_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.partial_solution_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: partial_solution_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.partial_solution_id_seq OWNED BY explorer.solution.id;


--
-- Name: transition_input_private; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.transition_input_private (
    id integer NOT NULL,
    transition_input_id integer NOT NULL,
    ciphertext_hash text NOT NULL,
    ciphertext text
);


--
-- Name: private_transition_input_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.private_transition_input_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: private_transition_input_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.private_transition_input_id_seq OWNED BY explorer.transition_input_private.id;


--
-- Name: program; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.program (
    id integer NOT NULL,
    transaction_deploy_id integer,
    program_id text NOT NULL,
    import text[],
    mapping text[],
    interface text[],
    record text[],
    closure text[],
    function text[],
    raw_data bytea NOT NULL,
    is_helloworld boolean DEFAULT false NOT NULL,
    feature_hash bytea NOT NULL,
    owner text,
    signature text,
    leo_source text,
    address text NOT NULL
);


--
-- Name: program_filter_hash; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.program_filter_hash (
    hash bytea NOT NULL
);


--
-- Name: program_function; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.program_function (
    id integer NOT NULL,
    program_id integer NOT NULL,
    name text NOT NULL,
    input text[] NOT NULL,
    input_mode text[] NOT NULL,
    output text[] NOT NULL,
    output_mode text[] NOT NULL,
    finalize text[] NOT NULL,
    called bigint DEFAULT 0 NOT NULL
);


--
-- Name: program_function_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.program_function_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: program_function_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.program_function_id_seq OWNED BY explorer.program_function.id;


--
-- Name: program_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.program_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: program_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.program_id_seq OWNED BY explorer.program.id;


--
-- Name: ratification; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.ratification (
    id integer NOT NULL,
    block_id integer NOT NULL,
    index integer NOT NULL,
    type explorer.ratification_type NOT NULL,
    amount numeric(20,0)
);


--
-- Name: ratification_genesis_balance; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.ratification_genesis_balance (
    id integer NOT NULL,
    address text NOT NULL,
    amount numeric(20,0) NOT NULL
);


--
-- Name: ratification_genesis_balance_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.ratification_genesis_balance_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ratification_genesis_balance_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.ratification_genesis_balance_id_seq OWNED BY explorer.ratification_genesis_balance.id;


--
-- Name: ratification_genesis_bonded; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.ratification_genesis_bonded (
    id integer NOT NULL,
    staker text NOT NULL,
    validator text NOT NULL,
    withdrawal text NOT NULL,
    amount numeric(20,0) NOT NULL
);


--
-- Name: ratification_genesis_bonded_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.ratification_genesis_bonded_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ratification_genesis_bonded_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.ratification_genesis_bonded_id_seq OWNED BY explorer.ratification_genesis_bonded.id;


--
-- Name: ratification_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.ratification_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ratification_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.ratification_id_seq OWNED BY explorer.ratification.id;


--
-- Name: stats; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.stats (
    name text NOT NULL,
    value numeric(20,0) NOT NULL
);


--
-- Name: transaction; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.transaction (
    id integer NOT NULL,
    confirmed_transaction_id integer,
    transaction_id text NOT NULL,
    type explorer.transaction_type NOT NULL,
    first_seen bigint DEFAULT EXTRACT(epoch FROM now()),
    original_transaction_id text,
    aborted boolean DEFAULT false NOT NULL
);


--
-- Name: transaction_deploy; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.transaction_deploy (
    id integer NOT NULL,
    transaction_id integer NOT NULL,
    edition integer NOT NULL,
    verifying_keys bytea NOT NULL,
    program_id text,
    owner text
);


--
-- Name: transaction_deployment_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.transaction_deployment_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transaction_deployment_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.transaction_deployment_id_seq OWNED BY explorer.transaction_deploy.id;


--
-- Name: transaction_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.transaction_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transaction_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.transaction_id_seq OWNED BY explorer.transaction.id;


--
-- Name: transition; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.transition (
    id integer NOT NULL,
    transition_id text NOT NULL,
    transaction_execute_id integer,
    fee_id integer,
    program_id text NOT NULL,
    function_name text NOT NULL,
    tpk text NOT NULL,
    tcm text NOT NULL,
    index integer NOT NULL,
    scm text NOT NULL
);


--
-- Name: transition_finalize_future_argument_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.transition_finalize_future_argument_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transition_finalize_future_argument_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.transition_finalize_future_argument_id_seq OWNED BY explorer.future_argument.id;


--
-- Name: transition_output_future; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.transition_output_future (
    id integer NOT NULL,
    transition_output_id integer NOT NULL,
    future_hash text NOT NULL
);


--
-- Name: transition_finalize_future_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.transition_finalize_future_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transition_finalize_future_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.transition_finalize_future_id_seq OWNED BY explorer.transition_output_future.id;


--
-- Name: transition_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.transition_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transition_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.transition_id_seq OWNED BY explorer.transition.id;


--
-- Name: transition_input; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.transition_input (
    id integer NOT NULL,
    transition_id integer NOT NULL,
    type explorer.transition_data_type NOT NULL,
    index integer NOT NULL
);


--
-- Name: transition_input_external_record; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.transition_input_external_record (
    id integer NOT NULL,
    transition_input_id integer NOT NULL,
    commitment text NOT NULL
);


--
-- Name: transition_input_external_record_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.transition_input_external_record_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transition_input_external_record_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.transition_input_external_record_id_seq OWNED BY explorer.transition_input_external_record.id;


--
-- Name: transition_input_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.transition_input_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transition_input_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.transition_input_id_seq OWNED BY explorer.transition_input.id;


--
-- Name: transition_input_public; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.transition_input_public (
    id integer NOT NULL,
    transition_input_id integer NOT NULL,
    plaintext_hash text NOT NULL,
    plaintext bytea
);


--
-- Name: transition_input_public_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.transition_input_public_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transition_input_public_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.transition_input_public_id_seq OWNED BY explorer.transition_input_public.id;


--
-- Name: transition_input_record; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.transition_input_record (
    id integer NOT NULL,
    transition_input_id integer NOT NULL,
    serial_number text NOT NULL,
    tag text NOT NULL
);


--
-- Name: transition_input_record_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.transition_input_record_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transition_input_record_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.transition_input_record_id_seq OWNED BY explorer.transition_input_record.id;


--
-- Name: transition_output; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.transition_output (
    id integer NOT NULL,
    transition_id integer NOT NULL,
    type explorer.transition_data_type NOT NULL,
    index integer NOT NULL
);


--
-- Name: transition_output_external_record; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.transition_output_external_record (
    id integer NOT NULL,
    transition_output_id integer NOT NULL,
    commitment text NOT NULL
);


--
-- Name: transition_output_external_record_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.transition_output_external_record_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transition_output_external_record_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.transition_output_external_record_id_seq OWNED BY explorer.transition_output_external_record.id;


--
-- Name: transition_output_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.transition_output_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transition_output_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.transition_output_id_seq OWNED BY explorer.transition_output.id;


--
-- Name: transition_output_private; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.transition_output_private (
    id integer NOT NULL,
    transition_output_id integer NOT NULL,
    ciphertext_hash text NOT NULL,
    ciphertext text
);


--
-- Name: transition_output_private_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.transition_output_private_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transition_output_private_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.transition_output_private_id_seq OWNED BY explorer.transition_output_private.id;


--
-- Name: transition_output_public; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.transition_output_public (
    id integer NOT NULL,
    transition_output_id integer NOT NULL,
    plaintext_hash text NOT NULL,
    plaintext bytea
);


--
-- Name: transition_output_public_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.transition_output_public_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transition_output_public_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.transition_output_public_id_seq OWNED BY explorer.transition_output_public.id;


--
-- Name: transition_output_record; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.transition_output_record (
    id integer NOT NULL,
    transition_output_id integer NOT NULL,
    commitment text NOT NULL,
    checksum text NOT NULL,
    record_ciphertext text
);


--
-- Name: transition_output_record_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.transition_output_record_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transition_output_record_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.transition_output_record_id_seq OWNED BY explorer.transition_output_record.id;


--
-- Name: validator_info; Type: TABLE; Schema: explorer; Owner: -
--

CREATE TABLE explorer.validator_info (
    id integer NOT NULL,
    address text NOT NULL,
    website text,
    logo text
);


--
-- Name: validator_info_id_seq; Type: SEQUENCE; Schema: explorer; Owner: -
--

CREATE SEQUENCE explorer.validator_info_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: validator_info_id_seq; Type: SEQUENCE OWNED BY; Schema: explorer; Owner: -
--

ALTER SEQUENCE explorer.validator_info_id_seq OWNED BY explorer.validator_info.id;


--
-- Name: address_fee_history id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_fee_history ALTER COLUMN id SET DEFAULT nextval('explorer.address_fee_history_id_seq'::regclass);


--
-- Name: address_puzzle_reward_history id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_puzzle_reward_history ALTER COLUMN id SET DEFAULT nextval('explorer.address_puzzle_reward_history_id_seq'::regclass);


--
-- Name: address_stake_reward id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_stake_reward ALTER COLUMN id SET DEFAULT nextval('explorer.address_stake_reward_id_seq'::regclass);


--
-- Name: address_stake_reward_history id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_stake_reward_history ALTER COLUMN id SET DEFAULT nextval('explorer.address_stake_reward_history_id_seq'::regclass);


--
-- Name: address_tag id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_tag ALTER COLUMN id SET DEFAULT nextval('explorer.address_tag_id_seq'::regclass);


--
-- Name: address_transfer_in_history id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_transfer_in_history ALTER COLUMN id SET DEFAULT nextval('explorer.address_transfer_in_history_id_seq'::regclass);


--
-- Name: address_transfer_out_history id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_transfer_out_history ALTER COLUMN id SET DEFAULT nextval('explorer.address_transfer_out_history_id_seq'::regclass);


--
-- Name: authority id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.authority ALTER COLUMN id SET DEFAULT nextval('explorer.authority_id_seq'::regclass);


--
-- Name: block id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.block ALTER COLUMN id SET DEFAULT nextval('explorer.block_id_seq'::regclass);


--
-- Name: block_aborted_solution_id id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.block_aborted_solution_id ALTER COLUMN id SET DEFAULT nextval('explorer.block_aborted_solution_id_id_seq'::regclass);


--
-- Name: block_aborted_transaction_id id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.block_aborted_transaction_id ALTER COLUMN id SET DEFAULT nextval('explorer.block_aborted_transaction_id_id_seq'::regclass);


--
-- Name: block_validator id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.block_validator ALTER COLUMN id SET DEFAULT nextval('explorer.block_validator_id_seq'::regclass);


--
-- Name: committee_history id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.committee_history ALTER COLUMN id SET DEFAULT nextval('explorer.committee_history_id_seq'::regclass);


--
-- Name: committee_history_member id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.committee_history_member ALTER COLUMN id SET DEFAULT nextval('explorer.committee_history_member_id_seq'::regclass);


--
-- Name: confirmed_transaction id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.confirmed_transaction ALTER COLUMN id SET DEFAULT nextval('explorer.confirmed_transaction_id_seq'::regclass);


--
-- Name: dag_vertex id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.dag_vertex ALTER COLUMN id SET DEFAULT nextval('explorer.dag_vertex_id_seq'::regclass);


--
-- Name: dag_vertex_adjacency id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.dag_vertex_adjacency ALTER COLUMN id SET DEFAULT nextval('explorer.dag_vertex_adjacency_id_seq'::regclass);


--
-- Name: dag_vertex_signature id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.dag_vertex_signature ALTER COLUMN id SET DEFAULT nextval('explorer.dag_vertex_signature_id_seq'::regclass);


--
-- Name: dag_vertex_transmission_id id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.dag_vertex_transmission_id ALTER COLUMN id SET DEFAULT nextval('explorer.dag_vertex_transmission_id_id_seq'::regclass);


--
-- Name: fee id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.fee ALTER COLUMN id SET DEFAULT nextval('explorer.fee_id_seq'::regclass);


--
-- Name: feedback id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.feedback ALTER COLUMN id SET DEFAULT nextval('explorer.feedback_id_seq'::regclass);


--
-- Name: finalize_operation id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation ALTER COLUMN id SET DEFAULT nextval('explorer.finalize_operation_id_seq'::regclass);


--
-- Name: finalize_operation_initialize_mapping id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_initialize_mapping ALTER COLUMN id SET DEFAULT nextval('explorer.finalize_operation_initialize_mapping_id_seq'::regclass);


--
-- Name: finalize_operation_insert_kv id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_insert_kv ALTER COLUMN id SET DEFAULT nextval('explorer.finalize_operation_insert_kv_id_seq'::regclass);


--
-- Name: finalize_operation_remove_kv id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_remove_kv ALTER COLUMN id SET DEFAULT nextval('explorer.finalize_operation_remove_kv_id_seq'::regclass);


--
-- Name: finalize_operation_remove_mapping id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_remove_mapping ALTER COLUMN id SET DEFAULT nextval('explorer.finalize_operation_remove_mapping_id_seq'::regclass);


--
-- Name: finalize_operation_replace_mapping id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_replace_mapping ALTER COLUMN id SET DEFAULT nextval('explorer.finalize_operation_replace_mapping_id_seq'::regclass);


--
-- Name: finalize_operation_update_kv id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_update_kv ALTER COLUMN id SET DEFAULT nextval('explorer.finalize_operation_update_kv_id_seq'::regclass);


--
-- Name: future id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.future ALTER COLUMN id SET DEFAULT nextval('explorer.future_id_seq'::regclass);


--
-- Name: future_argument id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.future_argument ALTER COLUMN id SET DEFAULT nextval('explorer.transition_finalize_future_argument_id_seq'::regclass);


--
-- Name: mapping id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping ALTER COLUMN id SET DEFAULT nextval('explorer.mapping_id_seq'::regclass);


--
-- Name: mapping_bonded_history id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping_bonded_history ALTER COLUMN id SET DEFAULT nextval('explorer.mapping_bonded_history_id_seq'::regclass);


--
-- Name: mapping_bonded_value id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping_bonded_value ALTER COLUMN id SET DEFAULT nextval('explorer.mapping_bonded_value_id_seq'::regclass);


--
-- Name: mapping_committee_history id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping_committee_history ALTER COLUMN id SET DEFAULT nextval('explorer.mapping_committee_history_id_seq'::regclass);


--
-- Name: mapping_delegated_history id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping_delegated_history ALTER COLUMN id SET DEFAULT nextval('explorer.mapping_delegated_history_id_seq'::regclass);


--
-- Name: mapping_history id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping_history ALTER COLUMN id SET DEFAULT nextval('explorer.mapping_history_id_seq'::regclass);


--
-- Name: mapping_value id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping_value ALTER COLUMN id SET DEFAULT nextval('explorer.mapping_value_id_seq'::regclass);


--
-- Name: program id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.program ALTER COLUMN id SET DEFAULT nextval('explorer.program_id_seq'::regclass);


--
-- Name: program_function id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.program_function ALTER COLUMN id SET DEFAULT nextval('explorer.program_function_id_seq'::regclass);


--
-- Name: puzzle_solution id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.puzzle_solution ALTER COLUMN id SET DEFAULT nextval('explorer.coinbase_solution_id_seq'::regclass);


--
-- Name: ratification id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.ratification ALTER COLUMN id SET DEFAULT nextval('explorer.ratification_id_seq'::regclass);


--
-- Name: ratification_genesis_balance id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.ratification_genesis_balance ALTER COLUMN id SET DEFAULT nextval('explorer.ratification_genesis_balance_id_seq'::regclass);


--
-- Name: ratification_genesis_bonded id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.ratification_genesis_bonded ALTER COLUMN id SET DEFAULT nextval('explorer.ratification_genesis_bonded_id_seq'::regclass);


--
-- Name: solution id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.solution ALTER COLUMN id SET DEFAULT nextval('explorer.partial_solution_id_seq'::regclass);


--
-- Name: transaction id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transaction ALTER COLUMN id SET DEFAULT nextval('explorer.transaction_id_seq'::regclass);


--
-- Name: transaction_deploy id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transaction_deploy ALTER COLUMN id SET DEFAULT nextval('explorer.transaction_deployment_id_seq'::regclass);


--
-- Name: transaction_execute id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transaction_execute ALTER COLUMN id SET DEFAULT nextval('explorer.execute_transaction_id_seq'::regclass);


--
-- Name: transition id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition ALTER COLUMN id SET DEFAULT nextval('explorer.transition_id_seq'::regclass);


--
-- Name: transition_input id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_input ALTER COLUMN id SET DEFAULT nextval('explorer.transition_input_id_seq'::regclass);


--
-- Name: transition_input_external_record id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_input_external_record ALTER COLUMN id SET DEFAULT nextval('explorer.transition_input_external_record_id_seq'::regclass);


--
-- Name: transition_input_private id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_input_private ALTER COLUMN id SET DEFAULT nextval('explorer.private_transition_input_id_seq'::regclass);


--
-- Name: transition_input_public id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_input_public ALTER COLUMN id SET DEFAULT nextval('explorer.transition_input_public_id_seq'::regclass);


--
-- Name: transition_input_record id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_input_record ALTER COLUMN id SET DEFAULT nextval('explorer.transition_input_record_id_seq'::regclass);


--
-- Name: transition_output id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output ALTER COLUMN id SET DEFAULT nextval('explorer.transition_output_id_seq'::regclass);


--
-- Name: transition_output_external_record id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output_external_record ALTER COLUMN id SET DEFAULT nextval('explorer.transition_output_external_record_id_seq'::regclass);


--
-- Name: transition_output_future id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output_future ALTER COLUMN id SET DEFAULT nextval('explorer.transition_finalize_future_id_seq'::regclass);


--
-- Name: transition_output_private id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output_private ALTER COLUMN id SET DEFAULT nextval('explorer.transition_output_private_id_seq'::regclass);


--
-- Name: transition_output_public id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output_public ALTER COLUMN id SET DEFAULT nextval('explorer.transition_output_public_id_seq'::regclass);


--
-- Name: transition_output_record id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output_record ALTER COLUMN id SET DEFAULT nextval('explorer.transition_output_record_id_seq'::regclass);


--
-- Name: validator_info id; Type: DEFAULT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.validator_info ALTER COLUMN id SET DEFAULT nextval('explorer.validator_info_id_seq'::regclass);


--
-- Name: address_fee_history_last_id address_fee_history_last_id_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_fee_history_last_id
    ADD CONSTRAINT address_fee_history_last_id_pk PRIMARY KEY (address);


--
-- Name: address_fee_history address_fee_history_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_fee_history
    ADD CONSTRAINT address_fee_history_pk PRIMARY KEY (id);


--
-- Name: address_puzzle_reward_history_last_id address_puzzle_reward_history_last_id_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_puzzle_reward_history_last_id
    ADD CONSTRAINT address_puzzle_reward_history_last_id_pk PRIMARY KEY (address);


--
-- Name: address_puzzle_reward_history address_puzzle_reward_history_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_puzzle_reward_history
    ADD CONSTRAINT address_puzzle_reward_history_pk PRIMARY KEY (id);


--
-- Name: address_stake_reward_history address_stake_reward_history_pkey; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_stake_reward_history
    ADD CONSTRAINT address_stake_reward_history_pkey PRIMARY KEY (id);


--
-- Name: address_stake_reward address_stake_reward_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_stake_reward
    ADD CONSTRAINT address_stake_reward_pk PRIMARY KEY (id);


--
-- Name: address_stake_reward address_stake_reward_pk_2; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_stake_reward
    ADD CONSTRAINT address_stake_reward_pk_2 UNIQUE (address);


--
-- Name: address_tag address_tag_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_tag
    ADD CONSTRAINT address_tag_pk PRIMARY KEY (id);


--
-- Name: address_transfer_in_history_last_id address_transfer_in_history_last_id_pkey; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_transfer_in_history_last_id
    ADD CONSTRAINT address_transfer_in_history_last_id_pkey PRIMARY KEY (address);


--
-- Name: address_transfer_in_history address_transfer_in_history_pkey; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_transfer_in_history
    ADD CONSTRAINT address_transfer_in_history_pkey PRIMARY KEY (id);


--
-- Name: address_transfer_out_history_last_id address_transfer_out_history_last_id_pkey; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_transfer_out_history_last_id
    ADD CONSTRAINT address_transfer_out_history_last_id_pkey PRIMARY KEY (address);


--
-- Name: address_transfer_out_history address_transfer_out_history_pkey; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_transfer_out_history
    ADD CONSTRAINT address_transfer_out_history_pkey PRIMARY KEY (id);


--
-- Name: authority authority_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.authority
    ADD CONSTRAINT authority_pk PRIMARY KEY (id);


--
-- Name: block_aborted_solution_id block_aborted_solution_id_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.block_aborted_solution_id
    ADD CONSTRAINT block_aborted_solution_id_pk PRIMARY KEY (id);


--
-- Name: block_aborted_transaction_id block_aborted_transaction_id_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.block_aborted_transaction_id
    ADD CONSTRAINT block_aborted_transaction_id_pk PRIMARY KEY (id);


--
-- Name: block block_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.block
    ADD CONSTRAINT block_pk PRIMARY KEY (id);


--
-- Name: block_validator block_validator_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.block_validator
    ADD CONSTRAINT block_validator_pk PRIMARY KEY (id);


--
-- Name: committee_history_member committee_history_member_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.committee_history_member
    ADD CONSTRAINT committee_history_member_pk PRIMARY KEY (id);


--
-- Name: committee_history committee_history_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.committee_history
    ADD CONSTRAINT committee_history_pk PRIMARY KEY (id);


--
-- Name: confirmed_transaction confirmed_transaction_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.confirmed_transaction
    ADD CONSTRAINT confirmed_transaction_pk PRIMARY KEY (id);


--
-- Name: dag_vertex_adjacency dag_vertex_adjacency_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.dag_vertex_adjacency
    ADD CONSTRAINT dag_vertex_adjacency_pk PRIMARY KEY (id);


--
-- Name: dag_vertex dag_vertex_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.dag_vertex
    ADD CONSTRAINT dag_vertex_pk PRIMARY KEY (id);


--
-- Name: dag_vertex_signature dag_vertex_signature_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.dag_vertex_signature
    ADD CONSTRAINT dag_vertex_signature_pk PRIMARY KEY (id);


--
-- Name: dag_vertex_transmission_id dag_vertex_transmission_id_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.dag_vertex_transmission_id
    ADD CONSTRAINT dag_vertex_transmission_id_pk PRIMARY KEY (id);


--
-- Name: fee fee_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.fee
    ADD CONSTRAINT fee_pk PRIMARY KEY (id);


--
-- Name: feedback feedback_pkey; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.feedback
    ADD CONSTRAINT feedback_pkey PRIMARY KEY (id);


--
-- Name: finalize_operation_initialize_mapping finalize_operation_initialize_mapping_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_initialize_mapping
    ADD CONSTRAINT finalize_operation_initialize_mapping_pk PRIMARY KEY (id);


--
-- Name: finalize_operation_insert_kv finalize_operation_insert_kv_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_insert_kv
    ADD CONSTRAINT finalize_operation_insert_kv_pk PRIMARY KEY (id);


--
-- Name: finalize_operation finalize_operation_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation
    ADD CONSTRAINT finalize_operation_pk PRIMARY KEY (id);


--
-- Name: finalize_operation_remove_kv finalize_operation_remove_kv_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_remove_kv
    ADD CONSTRAINT finalize_operation_remove_kv_pk PRIMARY KEY (id);


--
-- Name: finalize_operation_remove_mapping finalize_operation_remove_mapping_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_remove_mapping
    ADD CONSTRAINT finalize_operation_remove_mapping_pk PRIMARY KEY (id);


--
-- Name: finalize_operation_replace_mapping finalize_operation_replace_mapping_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_replace_mapping
    ADD CONSTRAINT finalize_operation_replace_mapping_pk PRIMARY KEY (id);


--
-- Name: finalize_operation_update_kv finalize_operation_update_kv_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_update_kv
    ADD CONSTRAINT finalize_operation_update_kv_pk PRIMARY KEY (id);


--
-- Name: future future_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.future
    ADD CONSTRAINT future_pk PRIMARY KEY (id);


--
-- Name: mapping_bonded_history mapping_bonded_history_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping_bonded_history
    ADD CONSTRAINT mapping_bonded_history_pk PRIMARY KEY (id);


--
-- Name: mapping_bonded_value mapping_bonded_value_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping_bonded_value
    ADD CONSTRAINT mapping_bonded_value_pk PRIMARY KEY (id);


--
-- Name: mapping_committee_history mapping_committee_history_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping_committee_history
    ADD CONSTRAINT mapping_committee_history_pk PRIMARY KEY (id);


--
-- Name: mapping_delegated_history mapping_delegated_history_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping_delegated_history
    ADD CONSTRAINT mapping_delegated_history_pk PRIMARY KEY (id);


--
-- Name: mapping_history_last_id mapping_history_last_id_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping_history_last_id
    ADD CONSTRAINT mapping_history_last_id_pk PRIMARY KEY (key_id);


--
-- Name: mapping_history mapping_history_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping_history
    ADD CONSTRAINT mapping_history_pk PRIMARY KEY (id);


--
-- Name: mapping mapping_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping
    ADD CONSTRAINT mapping_pk PRIMARY KEY (id);


--
-- Name: mapping mapping_pk2; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping
    ADD CONSTRAINT mapping_pk2 UNIQUE (mapping_id);


--
-- Name: mapping mapping_pk3; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping
    ADD CONSTRAINT mapping_pk3 UNIQUE (program_id, mapping);


--
-- Name: mapping_value mapping_value_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping_value
    ADD CONSTRAINT mapping_value_pk PRIMARY KEY (id);


--
-- Name: mapping_value mapping_value_pk2; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping_value
    ADD CONSTRAINT mapping_value_pk2 UNIQUE (mapping_id, key_id);


--
-- Name: program_function program_function_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.program_function
    ADD CONSTRAINT program_function_pk PRIMARY KEY (id);


--
-- Name: program program_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.program
    ADD CONSTRAINT program_pk PRIMARY KEY (id);


--
-- Name: program program_pk2; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.program
    ADD CONSTRAINT program_pk2 UNIQUE (program_id);


--
-- Name: puzzle_solution puzzle_solution_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.puzzle_solution
    ADD CONSTRAINT puzzle_solution_pk PRIMARY KEY (id);


--
-- Name: ratification ratification_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.ratification
    ADD CONSTRAINT ratification_pk PRIMARY KEY (id);


--
-- Name: solution solution_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.solution
    ADD CONSTRAINT solution_pk PRIMARY KEY (id);


--
-- Name: stats stats_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.stats
    ADD CONSTRAINT stats_pk PRIMARY KEY (name);


--
-- Name: transaction_deploy transaction_deployment_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transaction_deploy
    ADD CONSTRAINT transaction_deployment_pk PRIMARY KEY (id);


--
-- Name: transaction_execute transaction_execute_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transaction_execute
    ADD CONSTRAINT transaction_execute_pk PRIMARY KEY (id);


--
-- Name: transaction transaction_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transaction
    ADD CONSTRAINT transaction_pk PRIMARY KEY (id);


--
-- Name: future_argument transition_finalize_future_argument_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.future_argument
    ADD CONSTRAINT transition_finalize_future_argument_pk PRIMARY KEY (id);


--
-- Name: transition_input_external_record transition_input_external_record_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_input_external_record
    ADD CONSTRAINT transition_input_external_record_pk PRIMARY KEY (id);


--
-- Name: transition_input transition_input_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_input
    ADD CONSTRAINT transition_input_pk PRIMARY KEY (id);


--
-- Name: transition_input_private transition_input_private_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_input_private
    ADD CONSTRAINT transition_input_private_pk PRIMARY KEY (id);


--
-- Name: transition_input_public transition_input_public_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_input_public
    ADD CONSTRAINT transition_input_public_pk PRIMARY KEY (id);


--
-- Name: transition_input_record transition_input_record_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_input_record
    ADD CONSTRAINT transition_input_record_pk PRIMARY KEY (id);


--
-- Name: transition_output_external_record transition_output_external_record_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output_external_record
    ADD CONSTRAINT transition_output_external_record_pk PRIMARY KEY (id);


--
-- Name: transition_output_future transition_output_future_id_uindex; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output_future
    ADD CONSTRAINT transition_output_future_id_uindex PRIMARY KEY (id);


--
-- Name: transition_output transition_output_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output
    ADD CONSTRAINT transition_output_pk PRIMARY KEY (id);


--
-- Name: transition_output_private transition_output_private_pkey; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output_private
    ADD CONSTRAINT transition_output_private_pkey PRIMARY KEY (id);


--
-- Name: transition_output_public transition_output_public_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output_public
    ADD CONSTRAINT transition_output_public_pk PRIMARY KEY (id);


--
-- Name: transition_output_record transition_output_record_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output_record
    ADD CONSTRAINT transition_output_record_pk PRIMARY KEY (id);


--
-- Name: transition transition_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition
    ADD CONSTRAINT transition_pk PRIMARY KEY (id);


--
-- Name: validator_info validator_info_pk; Type: CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.validator_info
    ADD CONSTRAINT validator_info_pk PRIMARY KEY (id);


--
-- Name: address_fee_history_address_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX address_fee_history_address_index ON explorer.address_fee_history USING btree (address);


--
-- Name: address_fee_history_height_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX address_fee_history_height_index ON explorer.address_fee_history USING btree (height);


--
-- Name: address_fee_history_previous_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX address_fee_history_previous_id_index ON explorer.address_fee_history USING btree (previous_id);


--
-- Name: address_puzzle_reward_history_address_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX address_puzzle_reward_history_address_index ON explorer.address_puzzle_reward_history USING btree (address);


--
-- Name: address_puzzle_reward_history_height_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX address_puzzle_reward_history_height_index ON explorer.address_puzzle_reward_history USING btree (height);


--
-- Name: address_puzzle_reward_history_previous_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX address_puzzle_reward_history_previous_id_index ON explorer.address_puzzle_reward_history USING btree (previous_id);


--
-- Name: address_stake_reward_history_content_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX address_stake_reward_history_content_index ON explorer.address_stake_reward_history USING gin (content);


--
-- Name: address_stake_reward_history_height_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX address_stake_reward_history_height_index ON explorer.address_stake_reward_history USING btree (height);


--
-- Name: address_stake_reward_stake_reward_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX address_stake_reward_stake_reward_index ON explorer.address_stake_reward USING btree (stake_reward);


--
-- Name: address_tag_address_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE UNIQUE INDEX address_tag_address_index ON explorer.address_tag USING btree (address);


--
-- Name: address_tag_tag_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE UNIQUE INDEX address_tag_tag_index ON explorer.address_tag USING btree (tag);


--
-- Name: address_transfer_in_history_address_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX address_transfer_in_history_address_index ON explorer.address_transfer_in_history USING btree (address);


--
-- Name: address_transfer_in_history_height_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX address_transfer_in_history_height_index ON explorer.address_transfer_in_history USING btree (height);


--
-- Name: address_transfer_in_history_previous_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX address_transfer_in_history_previous_id_index ON explorer.address_transfer_in_history USING btree (previous_id);


--
-- Name: address_transfer_out_history_address_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX address_transfer_out_history_address_index ON explorer.address_transfer_out_history USING btree (address);


--
-- Name: address_transfer_out_history_height_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX address_transfer_out_history_height_index ON explorer.address_transfer_out_history USING btree (height);


--
-- Name: address_transfer_out_history_previous_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX address_transfer_out_history_previous_id_index ON explorer.address_transfer_out_history USING btree (previous_id);


--
-- Name: address_transition_address_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX address_transition_address_index ON explorer.address_transition USING btree (address text_pattern_ops);


--
-- Name: address_transition_transition_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX address_transition_transition_id_index ON explorer.address_transition USING btree (transition_id);


--
-- Name: authority_block_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX authority_block_id_index ON explorer.authority USING btree (block_id);


--
-- Name: authority_type_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX authority_type_index ON explorer.authority USING btree (type);


--
-- Name: block_aborted_solution_id_block_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX block_aborted_solution_id_block_id_index ON explorer.block_aborted_solution_id USING btree (block_id);


--
-- Name: block_aborted_solution_id_solution_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX block_aborted_solution_id_solution_id_index ON explorer.block_aborted_solution_id USING btree (solution_id text_pattern_ops);


--
-- Name: block_aborted_transaction_id_block_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX block_aborted_transaction_id_block_id_index ON explorer.block_aborted_transaction_id USING btree (block_id);


--
-- Name: block_block_hash_uindex; Type: INDEX; Schema: explorer; Owner: -
--

CREATE UNIQUE INDEX block_block_hash_uindex ON explorer.block USING btree (block_hash text_pattern_ops);


--
-- Name: block_height_uindex; Type: INDEX; Schema: explorer; Owner: -
--

CREATE UNIQUE INDEX block_height_uindex ON explorer.block USING btree (height);


--
-- Name: block_timestamp_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX block_timestamp_index ON explorer.block USING btree ("timestamp");


--
-- Name: block_validator_block_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX block_validator_block_id_index ON explorer.block_validator USING btree (block_id);


--
-- Name: block_validator_validator_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX block_validator_validator_index ON explorer.block_validator USING btree (validator);


--
-- Name: committee_history_height_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX committee_history_height_index ON explorer.committee_history USING btree (height);


--
-- Name: committee_history_member_address_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX committee_history_member_address_index ON explorer.committee_history_member USING btree (address);


--
-- Name: committee_history_member_committee_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX committee_history_member_committee_id_index ON explorer.committee_history_member USING btree (committee_id);


--
-- Name: confirmed_transaction_block_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX confirmed_transaction_block_id_index ON explorer.confirmed_transaction USING btree (block_id);


--
-- Name: confirmed_transaction_type_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX confirmed_transaction_type_index ON explorer.confirmed_transaction USING btree (type);


--
-- Name: dag_vertex_adjacency_end_vertex_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX dag_vertex_adjacency_end_vertex_index ON explorer.dag_vertex_adjacency USING btree (previous_vertex_id);


--
-- Name: dag_vertex_adjacency_index_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX dag_vertex_adjacency_index_index ON explorer.dag_vertex_adjacency USING btree (index);


--
-- Name: dag_vertex_adjacency_vertex_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX dag_vertex_adjacency_vertex_id_index ON explorer.dag_vertex_adjacency USING btree (vertex_id);


--
-- Name: dag_vertex_author_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX dag_vertex_author_index ON explorer.dag_vertex USING btree (author);


--
-- Name: dag_vertex_authority_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX dag_vertex_authority_id_index ON explorer.dag_vertex USING btree (authority_id);


--
-- Name: dag_vertex_round_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX dag_vertex_round_index ON explorer.dag_vertex USING btree (round);


--
-- Name: dag_vertex_signature_vertex_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX dag_vertex_signature_vertex_id_index ON explorer.dag_vertex_signature USING btree (vertex_id);


--
-- Name: dag_vertex_transmission_id_vertex_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX dag_vertex_transmission_id_vertex_id_index ON explorer.dag_vertex_transmission_id USING btree (vertex_id);


--
-- Name: fee_transaction_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX fee_transaction_id_index ON explorer.fee USING btree (transaction_id);


--
-- Name: finalize_operation_confirmed_transaction_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX finalize_operation_confirmed_transaction_id_index ON explorer.finalize_operation USING btree (confirmed_transaction_id);


--
-- Name: finalize_operation_initialize_mapping_finalize_operation_id_ind; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX finalize_operation_initialize_mapping_finalize_operation_id_ind ON explorer.finalize_operation_initialize_mapping USING btree (finalize_operation_id);


--
-- Name: finalize_operation_initialize_mapping_mapping_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX finalize_operation_initialize_mapping_mapping_id_index ON explorer.finalize_operation_initialize_mapping USING btree (mapping_id);


--
-- Name: finalize_operation_insert_kv_finalize_operation_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX finalize_operation_insert_kv_finalize_operation_id_index ON explorer.finalize_operation_insert_kv USING btree (finalize_operation_id);


--
-- Name: finalize_operation_insert_kv_mapping_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX finalize_operation_insert_kv_mapping_id_index ON explorer.finalize_operation_insert_kv USING btree (mapping_id);


--
-- Name: finalize_operation_remove_kv_finalize_operation_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX finalize_operation_remove_kv_finalize_operation_id_index ON explorer.finalize_operation_remove_kv USING btree (finalize_operation_id);


--
-- Name: finalize_operation_remove_kv_mapping_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX finalize_operation_remove_kv_mapping_id_index ON explorer.finalize_operation_remove_kv USING btree (mapping_id);


--
-- Name: finalize_operation_remove_mapping_finalize_operation_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX finalize_operation_remove_mapping_finalize_operation_id_index ON explorer.finalize_operation_remove_mapping USING btree (finalize_operation_id);


--
-- Name: finalize_operation_remove_mapping_mapping_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX finalize_operation_remove_mapping_mapping_id_index ON explorer.finalize_operation_remove_mapping USING btree (mapping_id);


--
-- Name: finalize_operation_replace_mapping_finalize_operation_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX finalize_operation_replace_mapping_finalize_operation_id_index ON explorer.finalize_operation_replace_mapping USING btree (finalize_operation_id);


--
-- Name: finalize_operation_replace_mapping_mapping_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX finalize_operation_replace_mapping_mapping_id_index ON explorer.finalize_operation_replace_mapping USING btree (mapping_id);


--
-- Name: finalize_operation_type_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX finalize_operation_type_index ON explorer.finalize_operation USING btree (type);


--
-- Name: finalize_operation_update_kv_finalize_operation_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX finalize_operation_update_kv_finalize_operation_id_index ON explorer.finalize_operation_update_kv USING btree (finalize_operation_id);


--
-- Name: finalize_operation_update_kv_mapping_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX finalize_operation_update_kv_mapping_id_index ON explorer.finalize_operation_update_kv USING btree (mapping_id);


--
-- Name: future_argument_future_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX future_argument_future_id_index ON explorer.future_argument USING btree (future_id);


--
-- Name: future_future_argument_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX future_future_argument_id_index ON explorer.future USING btree (future_argument_id);


--
-- Name: future_program_id_function_name_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX future_program_id_function_name_index ON explorer.future USING btree (program_id, function_name);


--
-- Name: future_transition_output_future_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX future_transition_output_future_id_index ON explorer.future USING btree (transition_output_future_id);


--
-- Name: future_type_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX future_type_index ON explorer.future USING btree (type);


--
-- Name: mapping_bonded_history_content_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX mapping_bonded_history_content_index ON explorer.mapping_bonded_history USING gin (content);


--
-- Name: mapping_bonded_history_height_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX mapping_bonded_history_height_index ON explorer.mapping_bonded_history USING btree (height);


--
-- Name: mapping_bonded_value_key_id_uindex; Type: INDEX; Schema: explorer; Owner: -
--

CREATE UNIQUE INDEX mapping_bonded_value_key_id_uindex ON explorer.mapping_bonded_value USING btree (key_id);


--
-- Name: mapping_committee_history_content_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX mapping_committee_history_content_index ON explorer.mapping_committee_history USING gin (content);


--
-- Name: mapping_committee_history_height_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX mapping_committee_history_height_index ON explorer.mapping_committee_history USING btree (height);


--
-- Name: mapping_delegated_history_content_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX mapping_delegated_history_content_index ON explorer.mapping_delegated_history USING gin (content);


--
-- Name: mapping_delegated_history_height_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX mapping_delegated_history_height_index ON explorer.mapping_delegated_history USING btree (height);


--
-- Name: mapping_history_height_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX mapping_history_height_index ON explorer.mapping_history USING btree (height);


--
-- Name: mapping_history_key_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX mapping_history_key_id_index ON explorer.mapping_history USING btree (key_id);


--
-- Name: mapping_history_mapping_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX mapping_history_mapping_id_index ON explorer.mapping_history USING btree (mapping_id);


--
-- Name: mapping_history_previous_id_uindex; Type: INDEX; Schema: explorer; Owner: -
--

CREATE UNIQUE INDEX mapping_history_previous_id_uindex ON explorer.mapping_history USING btree (previous_id);


--
-- Name: mapping_value_key_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX mapping_value_key_id_index ON explorer.mapping_value USING btree (key_id);


--
-- Name: mapping_value_mapping_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX mapping_value_mapping_id_index ON explorer.mapping_value USING btree (mapping_id);


--
-- Name: program_address_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX program_address_index ON explorer.program USING btree (address text_pattern_ops);


--
-- Name: program_feature_hash_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX program_feature_hash_index ON explorer.program USING btree (feature_hash);


--
-- Name: program_feature_hash_index2; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX program_feature_hash_index2 ON explorer.program USING hash (feature_hash);


--
-- Name: program_filter_hash_hash_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX program_filter_hash_hash_index ON explorer.program_filter_hash USING btree (hash);


--
-- Name: program_function_name_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX program_function_name_index ON explorer.program_function USING btree (name);


--
-- Name: program_function_program_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX program_function_program_id_index ON explorer.program_function USING btree (program_id);


--
-- Name: program_import_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX program_import_index ON explorer.program USING gin (import);


--
-- Name: program_is_helloworld_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX program_is_helloworld_index ON explorer.program USING btree (is_helloworld);


--
-- Name: program_owner_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX program_owner_index ON explorer.program USING btree (owner);


--
-- Name: program_transaction_deploy_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX program_transaction_deploy_id_index ON explorer.program USING btree (transaction_deploy_id);


--
-- Name: puzzle_solution_block_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX puzzle_solution_block_id_index ON explorer.puzzle_solution USING btree (block_id);


--
-- Name: ratification_block_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX ratification_block_id_index ON explorer.ratification USING btree (block_id);


--
-- Name: ratification_type_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX ratification_type_index ON explorer.ratification USING btree (type);


--
-- Name: solution_address_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX solution_address_index ON explorer.solution USING btree (address text_pattern_ops);


--
-- Name: solution_puzzle_solution_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX solution_puzzle_solution_id_index ON explorer.solution USING btree (solution_id text_pattern_ops);


--
-- Name: transaction_confirmed_transaction_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transaction_confirmed_transaction_id_index ON explorer.transaction USING btree (confirmed_transaction_id);


--
-- Name: transaction_deployment_transaction_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transaction_deployment_transaction_id_index ON explorer.transaction_deploy USING btree (transaction_id);


--
-- Name: transaction_execute_transaction_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transaction_execute_transaction_id_index ON explorer.transaction_execute USING btree (transaction_id);


--
-- Name: transaction_first_seen_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transaction_first_seen_index ON explorer.transaction USING btree (first_seen);


--
-- Name: transaction_original_transaction_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transaction_original_transaction_id_index ON explorer.transaction USING btree (original_transaction_id text_pattern_ops);


--
-- Name: transaction_transaction_id_uindex; Type: INDEX; Schema: explorer; Owner: -
--

CREATE UNIQUE INDEX transaction_transaction_id_uindex ON explorer.transaction USING btree (transaction_id text_pattern_ops);


--
-- Name: transition_fee_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transition_fee_id_index ON explorer.transition USING btree (fee_id);


--
-- Name: transition_finalize_future_argument_type_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transition_finalize_future_argument_type_index ON explorer.future_argument USING btree (type);


--
-- Name: transition_function_name_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transition_function_name_index ON explorer.transition USING btree (function_name);


--
-- Name: transition_input_external_record_transition_input_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transition_input_external_record_transition_input_id_index ON explorer.transition_input_external_record USING btree (transition_input_id);


--
-- Name: transition_input_private_transition_input_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transition_input_private_transition_input_id_index ON explorer.transition_input_private USING btree (transition_input_id);


--
-- Name: transition_input_public_transition_input_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transition_input_public_transition_input_id_index ON explorer.transition_input_public USING btree (transition_input_id);


--
-- Name: transition_input_record_transition_input_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transition_input_record_transition_input_id_index ON explorer.transition_input_record USING btree (transition_input_id);


--
-- Name: transition_input_transition_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transition_input_transition_id_index ON explorer.transition_input USING btree (transition_id);


--
-- Name: transition_output_external_record_transition_output_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transition_output_external_record_transition_output_id_index ON explorer.transition_output_external_record USING btree (transition_output_id);


--
-- Name: transition_output_future_transition_output_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transition_output_future_transition_output_id_index ON explorer.transition_output_future USING btree (transition_output_id);


--
-- Name: transition_output_private_transition_output_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transition_output_private_transition_output_id_index ON explorer.transition_output_private USING btree (transition_output_id);


--
-- Name: transition_output_public_transition_output_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transition_output_public_transition_output_id_index ON explorer.transition_output_public USING btree (transition_output_id);


--
-- Name: transition_output_record_transition_output_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transition_output_record_transition_output_id_index ON explorer.transition_output_record USING btree (transition_output_id);


--
-- Name: transition_output_transition_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transition_output_transition_id_index ON explorer.transition_output USING btree (transition_id);


--
-- Name: transition_program_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transition_program_id_index ON explorer.transition USING btree (program_id);


--
-- Name: transition_transaction_execute_id_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE INDEX transition_transaction_execute_id_index ON explorer.transition USING btree (transaction_execute_id);


--
-- Name: transition_transition_id_uindex; Type: INDEX; Schema: explorer; Owner: -
--

CREATE UNIQUE INDEX transition_transition_id_uindex ON explorer.transition USING btree (transition_id text_pattern_ops);


--
-- Name: validator_info_address_index; Type: INDEX; Schema: explorer; Owner: -
--

CREATE UNIQUE INDEX validator_info_address_index ON explorer.validator_info USING btree (address);


--
-- Name: address_transition address_stats_transition_transition_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.address_transition
    ADD CONSTRAINT address_stats_transition_transition_id_fk FOREIGN KEY (transition_id) REFERENCES explorer.transition(id) ON DELETE CASCADE;


--
-- Name: authority authority_block_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.authority
    ADD CONSTRAINT authority_block_id_fk FOREIGN KEY (block_id) REFERENCES explorer.block(id) ON DELETE CASCADE;


--
-- Name: block_aborted_solution_id block_aborted_solution_id_block_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.block_aborted_solution_id
    ADD CONSTRAINT block_aborted_solution_id_block_id_fk FOREIGN KEY (block_id) REFERENCES explorer.block(id) ON DELETE CASCADE;


--
-- Name: block_aborted_transaction_id block_aborted_transaction_id_block_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.block_aborted_transaction_id
    ADD CONSTRAINT block_aborted_transaction_id_block_id_fk FOREIGN KEY (block_id) REFERENCES explorer.block(id) ON DELETE CASCADE;


--
-- Name: block_validator block_validator_block_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.block_validator
    ADD CONSTRAINT block_validator_block_id_fk FOREIGN KEY (block_id) REFERENCES explorer.block(id) ON DELETE CASCADE;


--
-- Name: committee_history_member committee_history_member_committee_history_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.committee_history_member
    ADD CONSTRAINT committee_history_member_committee_history_id_fk FOREIGN KEY (committee_id) REFERENCES explorer.committee_history(id) ON DELETE CASCADE;


--
-- Name: confirmed_transaction confirmed_transaction_block_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.confirmed_transaction
    ADD CONSTRAINT confirmed_transaction_block_id_fk FOREIGN KEY (block_id) REFERENCES explorer.block(id) ON DELETE CASCADE;


--
-- Name: dag_vertex_adjacency dag_vertex_adjacency_dag_vertex_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.dag_vertex_adjacency
    ADD CONSTRAINT dag_vertex_adjacency_dag_vertex_id_fk FOREIGN KEY (vertex_id) REFERENCES explorer.dag_vertex(id) ON DELETE CASCADE;


--
-- Name: dag_vertex_adjacency dag_vertex_adjacency_dag_vertex_id_fk2; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.dag_vertex_adjacency
    ADD CONSTRAINT dag_vertex_adjacency_dag_vertex_id_fk2 FOREIGN KEY (previous_vertex_id) REFERENCES explorer.dag_vertex(id) ON DELETE CASCADE;


--
-- Name: dag_vertex dag_vertex_authority_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.dag_vertex
    ADD CONSTRAINT dag_vertex_authority_id_fk FOREIGN KEY (authority_id) REFERENCES explorer.authority(id) ON DELETE CASCADE;


--
-- Name: dag_vertex_signature dag_vertex_signature_dag_vertex_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.dag_vertex_signature
    ADD CONSTRAINT dag_vertex_signature_dag_vertex_id_fk FOREIGN KEY (vertex_id) REFERENCES explorer.dag_vertex(id) ON DELETE CASCADE;


--
-- Name: dag_vertex_transmission_id dag_vertex_transmission_id_dag_vertex_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.dag_vertex_transmission_id
    ADD CONSTRAINT dag_vertex_transmission_id_dag_vertex_id_fk FOREIGN KEY (vertex_id) REFERENCES explorer.dag_vertex(id) ON DELETE CASCADE;


--
-- Name: fee fee_transaction_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.fee
    ADD CONSTRAINT fee_transaction_id_fk FOREIGN KEY (transaction_id) REFERENCES explorer.transaction(id) ON DELETE CASCADE;


--
-- Name: finalize_operation finalize_operation_confirmed_transaction_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation
    ADD CONSTRAINT finalize_operation_confirmed_transaction_id_fk FOREIGN KEY (confirmed_transaction_id) REFERENCES explorer.confirmed_transaction(id) ON DELETE CASCADE;


--
-- Name: finalize_operation_initialize_mapping finalize_operation_initialize_mapping_finalize_operation_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_initialize_mapping
    ADD CONSTRAINT finalize_operation_initialize_mapping_finalize_operation_id_fk FOREIGN KEY (finalize_operation_id) REFERENCES explorer.finalize_operation(id) ON DELETE CASCADE;


--
-- Name: finalize_operation_insert_kv finalize_operation_insert_kv_finalize_operation_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_insert_kv
    ADD CONSTRAINT finalize_operation_insert_kv_finalize_operation_id_fk FOREIGN KEY (finalize_operation_id) REFERENCES explorer.finalize_operation(id) ON DELETE CASCADE;


--
-- Name: finalize_operation_remove_kv finalize_operation_remove_kv_finalize_operation_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_remove_kv
    ADD CONSTRAINT finalize_operation_remove_kv_finalize_operation_id_fk FOREIGN KEY (finalize_operation_id) REFERENCES explorer.finalize_operation(id) ON DELETE CASCADE;


--
-- Name: finalize_operation_remove_mapping finalize_operation_remove_mapping_finalize_operation_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_remove_mapping
    ADD CONSTRAINT finalize_operation_remove_mapping_finalize_operation_id_fk FOREIGN KEY (finalize_operation_id) REFERENCES explorer.finalize_operation(id) ON DELETE CASCADE;


--
-- Name: finalize_operation_replace_mapping finalize_operation_replace_mapping_finalize_operation_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_replace_mapping
    ADD CONSTRAINT finalize_operation_replace_mapping_finalize_operation_id_fk FOREIGN KEY (finalize_operation_id) REFERENCES explorer.finalize_operation(id) ON DELETE CASCADE;


--
-- Name: finalize_operation_update_kv finalize_operation_update_kv_finalize_operation_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.finalize_operation_update_kv
    ADD CONSTRAINT finalize_operation_update_kv_finalize_operation_id_fk FOREIGN KEY (finalize_operation_id) REFERENCES explorer.finalize_operation(id) ON DELETE CASCADE;


--
-- Name: future_argument future_argument_future_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.future_argument
    ADD CONSTRAINT future_argument_future_id_fk FOREIGN KEY (future_id) REFERENCES explorer.future(id) ON DELETE CASCADE;


--
-- Name: future future_future_argument_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.future
    ADD CONSTRAINT future_future_argument_id_fk FOREIGN KEY (future_argument_id) REFERENCES explorer.future_argument(id) ON DELETE CASCADE;


--
-- Name: future future_transition_output_future_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.future
    ADD CONSTRAINT future_transition_output_future_id_fk FOREIGN KEY (transition_output_future_id) REFERENCES explorer.transition_output_future(id) ON DELETE CASCADE;


--
-- Name: mapping_history mapping_history_mapping_history_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping_history
    ADD CONSTRAINT mapping_history_mapping_history_id_fk FOREIGN KEY (previous_id) REFERENCES explorer.mapping_history(id) ON UPDATE RESTRICT ON DELETE CASCADE;


--
-- Name: mapping_history mapping_history_mapping_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping_history
    ADD CONSTRAINT mapping_history_mapping_id_fk FOREIGN KEY (mapping_id) REFERENCES explorer.mapping(id) ON DELETE CASCADE;


--
-- Name: mapping_value mapping_value_mapping_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.mapping_value
    ADD CONSTRAINT mapping_value_mapping_id_fk FOREIGN KEY (mapping_id) REFERENCES explorer.mapping(id) ON DELETE CASCADE;


--
-- Name: program_function program_function_program_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.program_function
    ADD CONSTRAINT program_function_program_id_fk FOREIGN KEY (program_id) REFERENCES explorer.program(id) ON DELETE CASCADE;


--
-- Name: program program_transaction_deployment_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.program
    ADD CONSTRAINT program_transaction_deployment_id_fk FOREIGN KEY (transaction_deploy_id) REFERENCES explorer.transaction_deploy(id) ON DELETE CASCADE;


--
-- Name: puzzle_solution puzzle_solution_block_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.puzzle_solution
    ADD CONSTRAINT puzzle_solution_block_id_fk FOREIGN KEY (block_id) REFERENCES explorer.block(id) ON DELETE CASCADE;


--
-- Name: ratification ratification_block_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.ratification
    ADD CONSTRAINT ratification_block_id_fk FOREIGN KEY (block_id) REFERENCES explorer.block(id) ON DELETE CASCADE;


--
-- Name: solution solution_puzzle_solution_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.solution
    ADD CONSTRAINT solution_puzzle_solution_id_fk FOREIGN KEY (puzzle_solution_id) REFERENCES explorer.puzzle_solution(id) ON DELETE CASCADE;


--
-- Name: transaction transaction_confirmed_transaction_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transaction
    ADD CONSTRAINT transaction_confirmed_transaction_id_fk FOREIGN KEY (confirmed_transaction_id) REFERENCES explorer.confirmed_transaction(id) ON DELETE CASCADE;


--
-- Name: transaction_deploy transaction_deployment_transaction_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transaction_deploy
    ADD CONSTRAINT transaction_deployment_transaction_id_fk FOREIGN KEY (transaction_id) REFERENCES explorer.transaction(id) ON DELETE CASCADE;


--
-- Name: transaction_execute transaction_execute_transaction_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transaction_execute
    ADD CONSTRAINT transaction_execute_transaction_id_fk FOREIGN KEY (transaction_id) REFERENCES explorer.transaction(id) ON DELETE CASCADE;


--
-- Name: transition transition_fee_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition
    ADD CONSTRAINT transition_fee_id_fk FOREIGN KEY (fee_id) REFERENCES explorer.fee(id) ON DELETE CASCADE;


--
-- Name: transition_input_external_record transition_input_external_record_transition_input_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_input_external_record
    ADD CONSTRAINT transition_input_external_record_transition_input_id_fk FOREIGN KEY (transition_input_id) REFERENCES explorer.transition_input(id) ON DELETE CASCADE;


--
-- Name: transition_input_private transition_input_private_transition_input_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_input_private
    ADD CONSTRAINT transition_input_private_transition_input_id_fk FOREIGN KEY (transition_input_id) REFERENCES explorer.transition_input(id) ON DELETE CASCADE;


--
-- Name: transition_input_public transition_input_public_transition_input_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_input_public
    ADD CONSTRAINT transition_input_public_transition_input_id_fk FOREIGN KEY (transition_input_id) REFERENCES explorer.transition_input(id) ON DELETE CASCADE;


--
-- Name: transition_input_record transition_input_record_transition_input_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_input_record
    ADD CONSTRAINT transition_input_record_transition_input_id_fk FOREIGN KEY (transition_input_id) REFERENCES explorer.transition_input(id) ON DELETE CASCADE;


--
-- Name: transition_input transition_input_transition_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_input
    ADD CONSTRAINT transition_input_transition_id_fk FOREIGN KEY (transition_id) REFERENCES explorer.transition(id) ON DELETE CASCADE;


--
-- Name: transition_output_external_record transition_output_external_record_transition_output_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output_external_record
    ADD CONSTRAINT transition_output_external_record_transition_output_id_fk FOREIGN KEY (transition_output_id) REFERENCES explorer.transition_output(id) ON DELETE CASCADE;


--
-- Name: transition_output_future transition_output_future_transition_output_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output_future
    ADD CONSTRAINT transition_output_future_transition_output_id_fk FOREIGN KEY (transition_output_id) REFERENCES explorer.transition_output(id) ON DELETE CASCADE;


--
-- Name: transition_output_private transition_output_private_transition_output_id_fkey; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output_private
    ADD CONSTRAINT transition_output_private_transition_output_id_fkey FOREIGN KEY (transition_output_id) REFERENCES explorer.transition_output(id) ON DELETE CASCADE;


--
-- Name: transition_output_public transition_output_public_transition_output_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output_public
    ADD CONSTRAINT transition_output_public_transition_output_id_fk FOREIGN KEY (transition_output_id) REFERENCES explorer.transition_output(id) ON DELETE CASCADE;


--
-- Name: transition_output_record transition_output_record_transition_output_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output_record
    ADD CONSTRAINT transition_output_record_transition_output_id_fk FOREIGN KEY (transition_output_id) REFERENCES explorer.transition_output(id) ON DELETE CASCADE;


--
-- Name: transition_output transition_output_transition_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition_output
    ADD CONSTRAINT transition_output_transition_id_fk FOREIGN KEY (transition_id) REFERENCES explorer.transition(id) ON DELETE CASCADE;


--
-- Name: transition transition_transaction_execute_id_fk; Type: FK CONSTRAINT; Schema: explorer; Owner: -
--

ALTER TABLE ONLY explorer.transition
    ADD CONSTRAINT transition_transaction_execute_id_fk FOREIGN KEY (transaction_execute_id) REFERENCES explorer.transaction_execute(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

