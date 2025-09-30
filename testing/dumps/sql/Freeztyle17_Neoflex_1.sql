-- Downloaded from: https://github.com/Freeztyle17/Neoflex_1/blob/76834a5e4eca2ffb6131a4bb82631d03463f371c/dump2-projectDE.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 16.1
-- Dumped by pg_dump version 16.1

-- Started on 2024-07-25 17:30:40

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
-- TOC entry 9 (class 2615 OID 17741)
-- Name: dm; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA dm;


ALTER SCHEMA dm OWNER TO postgres;

--
-- TOC entry 6 (class 2615 OID 16920)
-- Name: ds; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA ds;


ALTER SCHEMA ds OWNER TO postgres;

--
-- TOC entry 7 (class 2615 OID 16921)
-- Name: logs; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA logs;


ALTER SCHEMA logs OWNER TO postgres;

--
-- TOC entry 4 (class 2615 OID 2200)
-- Name: public; Type: SCHEMA; Schema: -; Owner: pg_database_owner
--

CREATE SCHEMA public;


ALTER SCHEMA public OWNER TO pg_database_owner;

--
-- TOC entry 4959 (class 0 OID 0)
-- Dependencies: 4
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: pg_database_owner
--

COMMENT ON SCHEMA public IS 'standard public schema';


--
-- TOC entry 8 (class 2615 OID 17166)
-- Name: raw; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA raw;


ALTER SCHEMA raw OWNER TO postgres;

--
-- TOC entry 271 (class 1255 OID 17744)
-- Name: fill_account_turnover_f(date); Type: PROCEDURE; Schema: dm; Owner: postgres
--

CREATE PROCEDURE dm.fill_account_turnover_f(IN i_ondate date)
    LANGUAGE plpgsql
    AS $$
declare
    v_RowCount int;
    v_StartDate date := date_trunc('month', i_OnDate); 
    v_EndDate date := date_trunc('month', i_OnDate) + interval '1 month - 1 day'; 
    v_CurrentDate date;
begin
    
    for v_CurrentDate in (select generate_series(i_OnDate::date, date_trunc('month', i_ondate) + interval '1 month' - interval '1 day'::interval, '1 day'::interval)) loop
    
        call dm.writelog('[BEGIN] fill(i_OnDate => date ''' 
             || to_char(v_CurrentDate, 'yyyy-mm-dd') 
             || ''');', 1
           );

        call dm.writelog('delete on_date = ' 
             || to_char(v_CurrentDate, 'yyyy-mm-dd'), 1
           );

        delete from dm.dm_account_turnover_f f
        where f.on_date = v_CurrentDate;

        call dm.writelog('insert', 1);

        insert into dm.dm_account_turnover_f
            (on_date, account_rk, credit_amount, credit_amount_rub, debet_amount, debet_amount_rub)
        with wt_turn as (
            select p.credit_account_rk as account_rk,
                   p.credit_amount,
                   p.credit_amount * nullif(er.reduced_cource, 1) as credit_amount_rub,
                   cast(null as numeric) as debet_amount,
                   cast(null as numeric) as debet_amount_rub
            from ds.ft_posting_f p
            join ds.md_account_d a on a.account_rk = p.credit_account_rk
            left join ds.md_exchange_rate_d er on er.currency_rk = a.currency_rk
                                               and v_CurrentDate between er.data_actual_date and er.data_actual_end_date
            where p.oper_date = v_CurrentDate
              and v_CurrentDate between a.data_actual_date and a.data_actual_end_date
              and a.data_actual_date between v_StartDate and v_EndDate
            union all
            select p.debet_account_rk as account_rk,
                   cast(null as numeric) as credit_amount,
                   cast(null as numeric) as credit_amount_rub,
                   p.debet_amount,
                   p.debet_amount * nullif(er.reduced_cource, 1) as debet_amount_rub
            from ds.ft_posting_f p
            join ds.md_account_d a on a.account_rk = p.debet_account_rk
            left join ds.md_exchange_rate_d er on er.currency_rk = a.currency_rk
                                               and v_CurrentDate between er.data_actual_date and er.data_actual_end_date
            where p.oper_date = v_CurrentDate
              and v_CurrentDate between a.data_actual_date and a.data_actual_end_date
              and a.data_actual_date between v_StartDate and v_EndDate
        )
        select v_CurrentDate as on_date,
               t.account_rk,
               sum(t.credit_amount) as credit_amount,
               sum(t.credit_amount_rub) as credit_amount_rub,
               sum(t.debet_amount) as debet_amount,
               sum(t.debet_amount_rub) as debet_amount_rub
        from wt_turn t
        group by t.account_rk;

        GET DIAGNOSTICS v_RowCount = ROW_COUNT;
        call dm.writelog('[END] inserted ' || to_char(v_RowCount, 'FM99999999') || ' rows.', 1);
        
        -- Коммит после каждой итерации, чтобы избежать переполнения транзакционных логов
        commit;

    end loop;

    -- Коммит в конце процедуры
    commit;
end;
$$;


ALTER PROCEDURE dm.fill_account_turnover_f(IN i_ondate date) OWNER TO postgres;

--
-- TOC entry 268 (class 1255 OID 17775)
-- Name: fill_f101_round_f(date); Type: PROCEDURE; Schema: dm; Owner: postgres
--

CREATE PROCEDURE dm.fill_f101_round_f(IN i_ondate date)
    LANGUAGE plpgsql
    AS $$
declare
	v_RowCount int;
begin
    call dm.writelog( '[BEGIN] fill(i_OnDate => date ''' 
         || to_char(i_OnDate, 'yyyy-mm-dd') 
         || ''');', 1
       );
    
    call dm.writelog( 'delete on_date = ' 
         || to_char(i_OnDate, 'yyyy-mm-dd'), 1
       );

    delete
      from dm.DM_F101_ROUND_F f
     where from_date = date_trunc('month', i_OnDate)  
       and to_date = (date_trunc('MONTH', i_OnDate) + INTERVAL '1 MONTH - 1 day');
   
    call dm.writelog('insert', 1);
   
    insert 
      into dm.dm_f101_round_f
           ( from_date         
           , to_date           
           , chapter           
           , ledger_account    
           , characteristic    
           , balance_in_rub    
           , balance_in_val    
           , balance_in_total  
           , turn_deb_rub      
           , turn_deb_val      
           , turn_deb_total    
           , turn_cre_rub      
           , turn_cre_val      
           , turn_cre_total    
           , balance_out_rub  
           , balance_out_val   
           , balance_out_total 
           )
    select  date_trunc('month', i_OnDate)        as from_date,
           (date_trunc('MONTH', i_OnDate) + INTERVAL '1 MONTH - 1 day')  as to_date,
           s.chapter                             as chapter,
           substr(acc_d.account_number, 1, 5)    as ledger_account,
           acc_d.char_type                       as characteristic,
           -- RUB balance
           sum( case 
                  when cur.currency_code in ('643', '810')
                  then b.balance_out
                  else 0
                 end
              )                                  as balance_in_rub,
          -- VAL balance converted to rub
          sum( case 
                 when cur.currency_code not in ('643', '810')
                 then b.balance_out * exch_r.reduced_cource
                 else 0
                end
             )                                   as balance_in_val,
          -- Total: RUB balance + VAL converted to rub
          sum(  case 
                 when cur.currency_code in ('643', '810')
                 then b.balance_out
                 else b.balance_out * exch_r.reduced_cource
               end
             )                                   as balance_in_total  ,
           -- RUB debet turnover
           sum(case 
                 when cur.currency_code in ('643', '810')
                 then at.debet_amount_rub
                 else 0
               end
           )                                     as turn_deb_rub,
           -- VAL debet turnover converted
           sum(case 
                 when cur.currency_code not in ('643', '810')
                 then at.debet_amount_rub
                 else 0
               end
           )                                     as turn_deb_val,
           -- SUM = RUB debet turnover + VAL debet turnover converted
           sum(at.debet_amount_rub)              as turn_deb_total,
           -- RUB credit turnover
           sum(case 
                 when cur.currency_code in ('643', '810')
                 then at.credit_amount_rub
                 else 0
               end
              )                                  as turn_cre_rub,
           -- VAL credit turnover converted
           sum(case 
                 when cur.currency_code not in ('643', '810')
                 then at.credit_amount_rub
                 else 0
               end
              )                                  as turn_cre_val,
           -- SUM = RUB credit turnover + VAL credit turnover converted
           sum(at.credit_amount_rub)             as turn_cre_total,
			 case
			    when acc_d.char_type = 'A' and acc_d.currency_code in ('643', '810')
			    then sum(b.balance_out) - sum(at.credit_amount_rub) + sum(at.debet_amount_rub)
			    when acc_d.char_type = 'P' and acc_d.currency_code in ('643', '810')
			    then sum(b.balance_out) + sum(at.credit_amount_rub) - sum(at.debet_amount_rub)
			    else null
			  end as balance_out_rub,
			  case
			    when acc_d.char_type = 'A' and acc_d.currency_code not in ('643', '810')
			    then sum(b.balance_out * exch_r.reduced_cource) - sum(at.credit_amount_rub) + sum(at.debet_amount_rub)
			    when acc_d.char_type = 'P' and acc_d.currency_code not in ('643', '810')
			    then sum(b.balance_out * exch_r.reduced_cource) + sum(at.credit_amount_rub) - sum(at.debet_amount_rub)
			    else null
			  end as balance_out_val,
			  case
			    when acc_d.char_type in ('A', 'P')
			    then coalesce(
			           case when acc_d.char_type = 'A' then sum(b.balance_out) else 0 end +
			           case when acc_d.char_type = 'P' then sum(b.balance_out * exch_r.reduced_cource) else 0 end
			         , 0) - coalesce(sum(at.credit_amount_rub), 0) + coalesce(sum(at.debet_amount_rub), 0)
			    else null
		  	end as balance_out_total
      from ds.md_ledger_account_s s
      join ds.md_account_d acc_d
        on substr(acc_d.account_number, 1, 5) = to_char(s.ledger_account, 'FM99999999')
      join ds.md_currency_d cur
        on cur.currency_rk = acc_d.currency_rk
      left 
      join ds.ft_balance_f b
        on b.account_rk = acc_d.account_rk
       and b.on_date  = (date_trunc('month', i_OnDate) - INTERVAL '1 day')
      left 
      join ds.md_exchange_rate_d exch_r
        on exch_r.currency_rk = acc_d.currency_rk
       and i_OnDate between exch_r.data_actual_date and exch_r.data_actual_end_date
      left 
      join dm.dm_account_turnover_f at
        on at.account_rk = acc_d.account_rk
       and at.on_date between date_trunc('month', i_OnDate) and (date_trunc('MONTH', i_OnDate) + INTERVAL '1 MONTH - 1 day')
     where i_OnDate between s.start_date and s.end_date
       and i_OnDate between acc_d.data_actual_date and acc_d.data_actual_end_date
       and i_OnDate between cur.data_actual_date and cur.data_actual_end_date
     group by s.chapter,
           substr(acc_d.account_number, 1, 5),
           acc_d.char_type,
           acc_d.currency_code;
	
	GET DIAGNOSTICS v_RowCount = ROW_COUNT;
	call dm.writelog('[END] inserted ' ||  to_char(v_RowCount,'FM99999999') || ' rows.', 1);

    commit;
    
  end;$$;


ALTER PROCEDURE dm.fill_f101_round_f(IN i_ondate date) OWNER TO postgres;

--
-- TOC entry 269 (class 1255 OID 17846)
-- Name: writelog(character varying, integer); Type: PROCEDURE; Schema: dm; Owner: postgres
--

CREATE PROCEDURE dm.writelog(IN i_message character varying, IN i_messagetype integer)
    LANGUAGE plpgsql
    AS $$
DECLARE
    log_NOTICE            CONSTANT INT := 1;
    log_WARNING           CONSTANT INT := 2;
    log_ERROR             CONSTANT INT := 3;
    log_DEBUG             CONSTANT INT := 4;

    c_splitToTable        CONSTANT INT := 4000;

    v_logDate             TIMESTAMP;
    v_message             VARCHAR;
BEGIN
    v_logDate := NOW();
    v_message := i_message;
    i_messageType := log_NOTICE; -- Присвоение жесткого значения log_NOTICE, можно изменить на i_messageType, если требуется

    WHILE LENGTH(v_message) > 0 LOOP
        INSERT INTO dm.lg_messages (
            date_time,
            pid,
            message,
            message_type,
            usename,
            datname,
            client_addr,
            application_name,
            backend_start
        )
        SELECT
            NOW(),
            pid,
            SUBSTR(v_message, 1, c_splitToTable),
            i_messageType,
            usename,
            datname,
            client_addr,
            application_name,
            backend_start
        FROM pg_stat_activity
        WHERE pid = pg_backend_pid();
        
        v_message := SUBSTR(v_message, c_splitToTable + 1);
    END LOOP;

    COMMIT;
END;
$$;


ALTER PROCEDURE dm.writelog(IN i_message character varying, IN i_messagetype integer) OWNER TO postgres;

--
-- TOC entry 270 (class 1255 OID 17792)
-- Name: fill_account_turnover_f(date); Type: PROCEDURE; Schema: ds; Owner: postgres
--

CREATE PROCEDURE ds.fill_account_turnover_f(IN i_ondate date)
    LANGUAGE plpgsql
    AS $$
declare
    v_RowCount int;
    v_StartDate date;
    v_EndDate date;
    v_CurrentDate date;
begin
    v_StartDate := date_trunc('month', i_OnDate);  -- начало месяца i_OnDate
    v_EndDate := v_StartDate + INTERVAL '1 MONTH - 1 day';  -- конец месяца i_OnDate

    v_CurrentDate := v_StartDate;

    while v_CurrentDate <= v_EndDate loop
        delete
          from dm.dm_account_turnover_f f
         where f.on_date = v_CurrentDate;

        insert
          into dm.dm_account_turnover_f
               ( on_date
               , account_rk
               , credit_amount
               , credit_amount_rub
               , debet_amount
               , debet_amount_rub
               )
        with wt_turn as
        ( select p.credit_account_rk                  as account_rk
               , p.credit_amount                      as credit_amount
               , p.credit_amount * nullif(er.reduced_cource, 1)         as credit_amount_rub
               , cast(null as numeric)                 as debet_amount
               , cast(null as numeric)                 as debet_amount_rub
            from ds.ft_posting_f p
            join ds.md_account_d a
              on a.account_rk = p.credit_account_rk
            left
            join ds.md_exchange_rate_d er
              on er.currency_rk = a.currency_rk
             and v_CurrentDate between er.data_actual_date and er.data_actual_end_date
           where p.oper_date = v_CurrentDate
           	 and v_CurrentDate between a.data_actual_date and a.data_actual_end_date
             and a.data_actual_date between date_trunc('month', v_CurrentDate) and (date_trunc('MONTH', v_CurrentDate) + INTERVAL '1 MONTH - 1 day')
           union all
          select p.debet_account_rk                   as account_rk
               , cast(null as numeric)                 as credit_amount
               , cast(null as numeric)                 as credit_amount_rub
               , p.debet_amount                       as debet_amount
               , p.debet_amount * nullif(er.reduced_cource, 1)          as debet_amount_rub
            from ds.ft_posting_f p
            join ds.md_account_d a
              on a.account_rk = p.debet_account_rk
            left 
            join ds.md_exchange_rate_d er
              on er.currency_rk = a.currency_rk
             and v_CurrentDate between er.data_actual_date and er.data_actual_end_date
           where p.oper_date = v_CurrentDate
           	 and v_CurrentDate between a.data_actual_date and a.data_actual_end_date
             and a.data_actual_date between date_trunc('month', v_CurrentDate) and (date_trunc('MONTH', v_CurrentDate) + INTERVAL '1 MONTH - 1 day')
        )
        select v_CurrentDate                        as on_date
             , t.account_rk
             , sum(t.credit_amount)                   as credit_amount
             , sum(t.credit_amount_rub)               as credit_amount_rub
             , sum(t.debet_amount)                    as debet_amount
             , sum(t.debet_amount_rub)                as debet_amount_rub
          from wt_turn t
         group by t.account_rk;

        GET DIAGNOSTICS v_RowCount = ROW_COUNT;

        v_CurrentDate := v_CurrentDate + INTERVAL '1 day';
    end loop;

    commit;
    
end;$$;


ALTER PROCEDURE ds.fill_account_turnover_f(IN i_ondate date) OWNER TO postgres;

--
-- TOC entry 267 (class 1255 OID 17826)
-- Name: get_credit_debet_summary(date); Type: FUNCTION; Schema: ds; Owner: postgres
--

CREATE FUNCTION ds.get_credit_debet_summary(input_date date) RETURNS TABLE(date date, max_credit numeric, min_credit numeric, max_debet numeric, min_debet numeric)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT
        input_date AS date,
        MAX(CASE WHEN credit_amount > 0 THEN credit_amount ELSE 0 END)::NUMERIC AS max_credit,
        MIN(CASE WHEN credit_amount > 0 THEN credit_amount ELSE 0 END)::NUMERIC AS min_credit,
        MAX(CASE WHEN debet_amount > 0 THEN debet_amount ELSE 0 END)::NUMERIC AS max_debet,
        MIN(CASE WHEN debet_amount > 0 THEN debet_amount ELSE 0 END)::NUMERIC AS min_debet
    FROM
        ds.ft_posting_f p
    WHERE
        p.oper_date = input_date;
END;
$$;


ALTER FUNCTION ds.get_credit_debet_summary(input_date date) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 219 (class 1259 OID 17745)
-- Name: dm_account_turnover_f; Type: TABLE; Schema: dm; Owner: postgres
--

CREATE TABLE dm.dm_account_turnover_f (
    on_date date,
    account_rk integer,
    credit_amount double precision,
    credit_amount_rub double precision,
    debet_amount double precision,
    debet_amount_rub double precision
);


ALTER TABLE dm.dm_account_turnover_f OWNER TO postgres;

--
-- TOC entry 220 (class 1259 OID 17751)
-- Name: dm_f101_round_f; Type: TABLE; Schema: dm; Owner: postgres
--

CREATE TABLE dm.dm_f101_round_f (
    from_date date,
    to_date date,
    chapter character(1),
    ledger_account character(5),
    characteristic character(1),
    balance_in_rub numeric(23,8),
    r_balance_in_rub numeric(23,8),
    balance_in_val numeric(23,8),
    r_balance_in_val numeric(23,8),
    balance_in_total numeric(23,8),
    r_balance_in_total numeric(23,8),
    turn_deb_rub numeric(23,8),
    r_turn_deb_rub numeric(23,8),
    turn_deb_val numeric(23,8),
    r_turn_deb_val numeric(23,8),
    turn_deb_total numeric(23,8),
    r_turn_deb_total numeric(23,8),
    turn_cre_rub numeric(23,8),
    r_turn_cre_rub numeric(23,8),
    turn_cre_val numeric(23,8),
    r_turn_cre_val numeric(23,8),
    turn_cre_total numeric(23,8),
    r_turn_cre_total numeric(23,8),
    balance_out_rub numeric(23,8),
    r_balance_out_rub numeric(23,8),
    balance_out_val numeric(23,8),
    r_balance_out_val numeric(23,8),
    balance_out_total numeric(23,8),
    r_balance_out_total numeric(23,8)
);


ALTER TABLE dm.dm_f101_round_f OWNER TO postgres;

--
-- TOC entry 223 (class 1259 OID 17838)
-- Name: lg_messages; Type: TABLE; Schema: dm; Owner: postgres
--

CREATE TABLE dm.lg_messages (
    record_id integer NOT NULL,
    date_time timestamp without time zone NOT NULL,
    pid integer NOT NULL,
    message character varying(4000) NOT NULL,
    message_type integer NOT NULL,
    usename character varying,
    datname character varying,
    client_addr character varying,
    application_name character varying,
    backend_start timestamp without time zone
);


ALTER TABLE dm.lg_messages OWNER TO postgres;

--
-- TOC entry 222 (class 1259 OID 17837)
-- Name: lg_messages_record_id_seq; Type: SEQUENCE; Schema: dm; Owner: postgres
--

CREATE SEQUENCE dm.lg_messages_record_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE dm.lg_messages_record_id_seq OWNER TO postgres;

--
-- TOC entry 4960 (class 0 OID 0)
-- Dependencies: 222
-- Name: lg_messages_record_id_seq; Type: SEQUENCE OWNED BY; Schema: dm; Owner: postgres
--

ALTER SEQUENCE dm.lg_messages_record_id_seq OWNED BY dm.lg_messages.record_id;


--
-- TOC entry 233 (class 1259 OID 20030)
-- Name: ft_balance_f; Type: TABLE; Schema: ds; Owner: postgres
--

CREATE TABLE ds.ft_balance_f (
    id integer NOT NULL,
    on_date date NOT NULL,
    account_rk integer NOT NULL,
    currency_rk integer,
    balance_out double precision
);


ALTER TABLE ds.ft_balance_f OWNER TO postgres;

--
-- TOC entry 232 (class 1259 OID 20029)
-- Name: ft_balance_f_id_seq; Type: SEQUENCE; Schema: ds; Owner: postgres
--

CREATE SEQUENCE ds.ft_balance_f_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ds.ft_balance_f_id_seq OWNER TO postgres;

--
-- TOC entry 4961 (class 0 OID 0)
-- Dependencies: 232
-- Name: ft_balance_f_id_seq; Type: SEQUENCE OWNED BY; Schema: ds; Owner: postgres
--

ALTER SEQUENCE ds.ft_balance_f_id_seq OWNED BY ds.ft_balance_f.id;


--
-- TOC entry 235 (class 1259 OID 20037)
-- Name: ft_posting_f; Type: TABLE; Schema: ds; Owner: postgres
--

CREATE TABLE ds.ft_posting_f (
    id integer NOT NULL,
    oper_date date NOT NULL,
    credit_account_rk integer NOT NULL,
    debet_account_rk integer NOT NULL,
    credit_amount double precision,
    debet_amount double precision
);


ALTER TABLE ds.ft_posting_f OWNER TO postgres;

--
-- TOC entry 234 (class 1259 OID 20036)
-- Name: ft_posting_f_id_seq; Type: SEQUENCE; Schema: ds; Owner: postgres
--

CREATE SEQUENCE ds.ft_posting_f_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ds.ft_posting_f_id_seq OWNER TO postgres;

--
-- TOC entry 4962 (class 0 OID 0)
-- Dependencies: 234
-- Name: ft_posting_f_id_seq; Type: SEQUENCE OWNED BY; Schema: ds; Owner: postgres
--

ALTER SEQUENCE ds.ft_posting_f_id_seq OWNED BY ds.ft_posting_f.id;


--
-- TOC entry 237 (class 1259 OID 20044)
-- Name: md_account_d; Type: TABLE; Schema: ds; Owner: postgres
--

CREATE TABLE ds.md_account_d (
    id integer NOT NULL,
    data_actual_date date NOT NULL,
    data_actual_end_date date NOT NULL,
    account_rk integer NOT NULL,
    account_number character varying(20) NOT NULL,
    char_type character varying(1) NOT NULL,
    currency_rk integer NOT NULL,
    currency_code character varying(3) NOT NULL
);


ALTER TABLE ds.md_account_d OWNER TO postgres;

--
-- TOC entry 236 (class 1259 OID 20043)
-- Name: md_account_d_id_seq; Type: SEQUENCE; Schema: ds; Owner: postgres
--

CREATE SEQUENCE ds.md_account_d_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ds.md_account_d_id_seq OWNER TO postgres;

--
-- TOC entry 4963 (class 0 OID 0)
-- Dependencies: 236
-- Name: md_account_d_id_seq; Type: SEQUENCE OWNED BY; Schema: ds; Owner: postgres
--

ALTER SEQUENCE ds.md_account_d_id_seq OWNED BY ds.md_account_d.id;


--
-- TOC entry 239 (class 1259 OID 20051)
-- Name: md_currency_d; Type: TABLE; Schema: ds; Owner: postgres
--

CREATE TABLE ds.md_currency_d (
    id integer NOT NULL,
    currency_rk integer NOT NULL,
    data_actual_date date NOT NULL,
    data_actual_end_date date,
    currency_code character varying(3),
    code_iso_char character varying(3) NOT NULL
);


ALTER TABLE ds.md_currency_d OWNER TO postgres;

--
-- TOC entry 238 (class 1259 OID 20050)
-- Name: md_currency_d_id_seq; Type: SEQUENCE; Schema: ds; Owner: postgres
--

CREATE SEQUENCE ds.md_currency_d_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ds.md_currency_d_id_seq OWNER TO postgres;

--
-- TOC entry 4964 (class 0 OID 0)
-- Dependencies: 238
-- Name: md_currency_d_id_seq; Type: SEQUENCE OWNED BY; Schema: ds; Owner: postgres
--

ALTER SEQUENCE ds.md_currency_d_id_seq OWNED BY ds.md_currency_d.id;


--
-- TOC entry 241 (class 1259 OID 20058)
-- Name: md_exchange_rate_d; Type: TABLE; Schema: ds; Owner: postgres
--

CREATE TABLE ds.md_exchange_rate_d (
    id integer NOT NULL,
    data_actual_date date NOT NULL,
    data_actual_end_date date,
    currency_rk integer NOT NULL,
    reduced_cource integer NOT NULL,
    code_iso_num character varying(3)
);


ALTER TABLE ds.md_exchange_rate_d OWNER TO postgres;

--
-- TOC entry 240 (class 1259 OID 20057)
-- Name: md_exchange_rate_d_id_seq; Type: SEQUENCE; Schema: ds; Owner: postgres
--

CREATE SEQUENCE ds.md_exchange_rate_d_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ds.md_exchange_rate_d_id_seq OWNER TO postgres;

--
-- TOC entry 4965 (class 0 OID 0)
-- Dependencies: 240
-- Name: md_exchange_rate_d_id_seq; Type: SEQUENCE OWNED BY; Schema: ds; Owner: postgres
--

ALTER SEQUENCE ds.md_exchange_rate_d_id_seq OWNED BY ds.md_exchange_rate_d.id;


--
-- TOC entry 243 (class 1259 OID 20065)
-- Name: md_ledger_account_s; Type: TABLE; Schema: ds; Owner: postgres
--

CREATE TABLE ds.md_ledger_account_s (
    id integer NOT NULL,
    chapter character varying(1),
    chapter_name character varying(16),
    section_number integer,
    section_name character varying(22),
    subsection_name character varying(21),
    ledger1_account integer,
    ledger1_account_name character varying(47),
    ledger_account integer NOT NULL,
    ledger_account_name character varying(153),
    characteristic character varying(1),
    is_resident integer,
    is_reserve integer,
    is_reserved integer,
    is_loan integer,
    is_reserved_assets integer,
    is_overdue integer,
    is_interest integer,
    pair_account character varying(5),
    start_date date NOT NULL,
    end_date date,
    is_rub_only integer,
    min_term character varying(1),
    min_term_measure character varying(1),
    max_term character varying(1),
    max_term_measure character varying(1),
    ledger_acc_full_name_translit character varying(16),
    is_revaluation character varying(1),
    is_correct character varying(1)
);


ALTER TABLE ds.md_ledger_account_s OWNER TO postgres;

--
-- TOC entry 242 (class 1259 OID 20064)
-- Name: md_ledger_account_s_id_seq; Type: SEQUENCE; Schema: ds; Owner: postgres
--

CREATE SEQUENCE ds.md_ledger_account_s_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE ds.md_ledger_account_s_id_seq OWNER TO postgres;

--
-- TOC entry 4966 (class 0 OID 0)
-- Dependencies: 242
-- Name: md_ledger_account_s_id_seq; Type: SEQUENCE OWNED BY; Schema: ds; Owner: postgres
--

ALTER SEQUENCE ds.md_ledger_account_s_id_seq OWNED BY ds.md_ledger_account_s.id;


--
-- TOC entry 231 (class 1259 OID 19593)
-- Name: etl_log; Type: TABLE; Schema: logs; Owner: postgres
--

CREATE TABLE logs.etl_log (
    id bigint NOT NULL,
    description character varying(255),
    object_of_oper character varying(255),
    time_of_oper timestamp(6) without time zone
);


ALTER TABLE logs.etl_log OWNER TO postgres;

--
-- TOC entry 230 (class 1259 OID 19592)
-- Name: etl_log_id_seq; Type: SEQUENCE; Schema: logs; Owner: postgres
--

ALTER TABLE logs.etl_log ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME logs.etl_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 227 (class 1259 OID 17924)
-- Name: etl_logs; Type: TABLE; Schema: logs; Owner: postgres
--

CREATE TABLE logs.etl_logs (
    log_id integer NOT NULL,
    process_name character varying(255) NOT NULL,
    oper_time timestamp without time zone NOT NULL,
    status character varying(255) NOT NULL
);


ALTER TABLE logs.etl_logs OWNER TO postgres;

--
-- TOC entry 226 (class 1259 OID 17923)
-- Name: etl_logs_log_id_seq; Type: SEQUENCE; Schema: logs; Owner: postgres
--

CREATE SEQUENCE logs.etl_logs_log_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE logs.etl_logs_log_id_seq OWNER TO postgres;

--
-- TOC entry 4967 (class 0 OID 0)
-- Dependencies: 226
-- Name: etl_logs_log_id_seq; Type: SEQUENCE OWNED BY; Schema: logs; Owner: postgres
--

ALTER SEQUENCE logs.etl_logs_log_id_seq OWNED BY logs.etl_logs.log_id;


--
-- TOC entry 229 (class 1259 OID 19509)
-- Name: dm_f101_round_f; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.dm_f101_round_f (
    id bigint NOT NULL,
    balance_in_rub double precision,
    balance_in_total double precision,
    balance_in_val double precision,
    balance_out_rub double precision,
    balance_out_total double precision,
    balance_out_val double precision,
    chapter character varying(255),
    characteristic character varying(255),
    from_date timestamp(6) without time zone,
    ledger_account character varying(255),
    r_balance_in_rub double precision,
    r_balance_in_total double precision,
    r_balance_in_val double precision,
    r_balance_out_rub double precision,
    r_balance_out_total double precision,
    r_balance_out_val double precision,
    r_turn_cre_rub double precision,
    r_turn_cre_total double precision,
    r_turn_cre_val double precision,
    r_turn_deb_rub double precision,
    r_turn_deb_total double precision,
    r_turn_deb_val double precision,
    to_date timestamp(6) without time zone,
    turn_cre_rub double precision,
    turn_cre_total double precision,
    turn_cre_val double precision,
    turn_deb_rub double precision,
    turn_deb_total double precision,
    turn_deb_val double precision
);


ALTER TABLE public.dm_f101_round_f OWNER TO postgres;

--
-- TOC entry 228 (class 1259 OID 19508)
-- Name: dm_f101_round_f_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.dm_f101_round_f ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.dm_f101_round_f_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 221 (class 1259 OID 17827)
-- Name: posting_summary; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.posting_summary (
    date timestamp(6) without time zone NOT NULL,
    max_credit numeric(38,2),
    max_debit numeric(38,2),
    min_credit numeric(38,2),
    min_debit numeric(38,2),
    max_debet numeric(38,2),
    min_debet numeric(38,2)
);


ALTER TABLE public.posting_summary OWNER TO postgres;

--
-- TOC entry 245 (class 1259 OID 20072)
-- Name: ft_balance_f; Type: TABLE; Schema: raw; Owner: postgres
--

CREATE TABLE raw.ft_balance_f (
    id integer NOT NULL,
    on_date timestamp(6) without time zone,
    account_rk integer,
    currency_rk integer,
    balance_out double precision
);


ALTER TABLE raw.ft_balance_f OWNER TO postgres;

--
-- TOC entry 244 (class 1259 OID 20071)
-- Name: ft_balance_f_id_seq; Type: SEQUENCE; Schema: raw; Owner: postgres
--

CREATE SEQUENCE raw.ft_balance_f_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE raw.ft_balance_f_id_seq OWNER TO postgres;

--
-- TOC entry 4968 (class 0 OID 0)
-- Dependencies: 244
-- Name: ft_balance_f_id_seq; Type: SEQUENCE OWNED BY; Schema: raw; Owner: postgres
--

ALTER SEQUENCE raw.ft_balance_f_id_seq OWNED BY raw.ft_balance_f.id;


--
-- TOC entry 247 (class 1259 OID 20077)
-- Name: ft_posting_f; Type: TABLE; Schema: raw; Owner: postgres
--

CREATE TABLE raw.ft_posting_f (
    id bigint NOT NULL,
    oper_date date,
    credit_account_rk bigint,
    debet_account_rk bigint,
    credit_amount double precision,
    debet_amount double precision
);


ALTER TABLE raw.ft_posting_f OWNER TO postgres;

--
-- TOC entry 246 (class 1259 OID 20076)
-- Name: ft_posting_f_id_seq; Type: SEQUENCE; Schema: raw; Owner: postgres
--

CREATE SEQUENCE raw.ft_posting_f_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE raw.ft_posting_f_id_seq OWNER TO postgres;

--
-- TOC entry 4969 (class 0 OID 0)
-- Dependencies: 246
-- Name: ft_posting_f_id_seq; Type: SEQUENCE OWNED BY; Schema: raw; Owner: postgres
--

ALTER SEQUENCE raw.ft_posting_f_id_seq OWNED BY raw.ft_posting_f.id;


--
-- TOC entry 249 (class 1259 OID 20082)
-- Name: md_account_d; Type: TABLE; Schema: raw; Owner: postgres
--

CREATE TABLE raw.md_account_d (
    id bigint NOT NULL,
    data_actual_date date,
    data_actual_end_date date,
    account_rk bigint,
    account_number character varying(255),
    char_type character varying(255),
    currency_rk bigint,
    currency_code character varying(255)
);


ALTER TABLE raw.md_account_d OWNER TO postgres;

--
-- TOC entry 248 (class 1259 OID 20081)
-- Name: md_account_d_id_seq; Type: SEQUENCE; Schema: raw; Owner: postgres
--

CREATE SEQUENCE raw.md_account_d_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE raw.md_account_d_id_seq OWNER TO postgres;

--
-- TOC entry 4970 (class 0 OID 0)
-- Dependencies: 248
-- Name: md_account_d_id_seq; Type: SEQUENCE OWNED BY; Schema: raw; Owner: postgres
--

ALTER SEQUENCE raw.md_account_d_id_seq OWNED BY raw.md_account_d.id;


--
-- TOC entry 251 (class 1259 OID 20087)
-- Name: md_currency_d; Type: TABLE; Schema: raw; Owner: postgres
--

CREATE TABLE raw.md_currency_d (
    id bigint NOT NULL,
    currency_rk bigint,
    data_actual_date date,
    data_actual_end_date date,
    currency_code character varying(255),
    code_iso_char character varying(255)
);


ALTER TABLE raw.md_currency_d OWNER TO postgres;

--
-- TOC entry 250 (class 1259 OID 20086)
-- Name: md_currency_d_id_seq; Type: SEQUENCE; Schema: raw; Owner: postgres
--

CREATE SEQUENCE raw.md_currency_d_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE raw.md_currency_d_id_seq OWNER TO postgres;

--
-- TOC entry 4971 (class 0 OID 0)
-- Dependencies: 250
-- Name: md_currency_d_id_seq; Type: SEQUENCE OWNED BY; Schema: raw; Owner: postgres
--

ALTER SEQUENCE raw.md_currency_d_id_seq OWNED BY raw.md_currency_d.id;


--
-- TOC entry 253 (class 1259 OID 20092)
-- Name: md_exchange_rate_d; Type: TABLE; Schema: raw; Owner: postgres
--

CREATE TABLE raw.md_exchange_rate_d (
    id bigint NOT NULL,
    data_actual_date date,
    data_actual_end_date date,
    currency_rk bigint,
    reduced_cource double precision,
    code_iso_num character varying(255)
);


ALTER TABLE raw.md_exchange_rate_d OWNER TO postgres;

--
-- TOC entry 252 (class 1259 OID 20091)
-- Name: md_exchange_rate_d_id_seq; Type: SEQUENCE; Schema: raw; Owner: postgres
--

CREATE SEQUENCE raw.md_exchange_rate_d_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE raw.md_exchange_rate_d_id_seq OWNER TO postgres;

--
-- TOC entry 4972 (class 0 OID 0)
-- Dependencies: 252
-- Name: md_exchange_rate_d_id_seq; Type: SEQUENCE OWNED BY; Schema: raw; Owner: postgres
--

ALTER SEQUENCE raw.md_exchange_rate_d_id_seq OWNED BY raw.md_exchange_rate_d.id;


--
-- TOC entry 255 (class 1259 OID 20097)
-- Name: md_ledger_account_s; Type: TABLE; Schema: raw; Owner: postgres
--

CREATE TABLE raw.md_ledger_account_s (
    id bigint NOT NULL,
    chapter character varying(255),
    chapter_name character varying(255),
    section_number bigint,
    section_name character varying(255),
    subsection_name character varying(255),
    ledger1_account bigint,
    ledger1_account_name character varying(255),
    ledger_account bigint,
    ledger_account_name character varying(255),
    characteristic character varying(255),
    is_resident integer,
    is_reserve integer,
    is_reserved integer,
    is_loan integer,
    is_reserved_assets integer,
    is_overdue integer,
    is_interest integer,
    pair_account character varying(255),
    start_date date,
    end_date date,
    is_rub_only integer,
    min_term character varying(255),
    min_term_measure character varying(255),
    max_term character varying(255),
    max_term_measure character varying(255),
    ledger_acc_full_name_translit character varying(255),
    is_revaluation character varying(255),
    is_correct character varying(255),
    column1 integer
);


ALTER TABLE raw.md_ledger_account_s OWNER TO postgres;

--
-- TOC entry 254 (class 1259 OID 20096)
-- Name: md_ledger_account_s_id_seq; Type: SEQUENCE; Schema: raw; Owner: postgres
--

CREATE SEQUENCE raw.md_ledger_account_s_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE raw.md_ledger_account_s_id_seq OWNER TO postgres;

--
-- TOC entry 4973 (class 0 OID 0)
-- Dependencies: 254
-- Name: md_ledger_account_s_id_seq; Type: SEQUENCE OWNED BY; Schema: raw; Owner: postgres
--

ALTER SEQUENCE raw.md_ledger_account_s_id_seq OWNED BY raw.md_ledger_account_s.id;


--
-- TOC entry 225 (class 1259 OID 17850)
-- Name: rejected_data; Type: TABLE; Schema: raw; Owner: postgres
--

CREATE TABLE raw.rejected_data (
    rejected_id integer NOT NULL,
    table_name character varying(100) NOT NULL,
    error_message text NOT NULL,
    rejected_data text
);


ALTER TABLE raw.rejected_data OWNER TO postgres;

--
-- TOC entry 224 (class 1259 OID 17849)
-- Name: rejected_data_rejected_id_seq; Type: SEQUENCE; Schema: raw; Owner: postgres
--

CREATE SEQUENCE raw.rejected_data_rejected_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE raw.rejected_data_rejected_id_seq OWNER TO postgres;

--
-- TOC entry 4974 (class 0 OID 0)
-- Dependencies: 224
-- Name: rejected_data_rejected_id_seq; Type: SEQUENCE OWNED BY; Schema: raw; Owner: postgres
--

ALTER SEQUENCE raw.rejected_data_rejected_id_seq OWNED BY raw.rejected_data.rejected_id;


--
-- TOC entry 4735 (class 2604 OID 17841)
-- Name: lg_messages record_id; Type: DEFAULT; Schema: dm; Owner: postgres
--

ALTER TABLE ONLY dm.lg_messages ALTER COLUMN record_id SET DEFAULT nextval('dm.lg_messages_record_id_seq'::regclass);


--
-- TOC entry 4738 (class 2604 OID 20033)
-- Name: ft_balance_f id; Type: DEFAULT; Schema: ds; Owner: postgres
--

ALTER TABLE ONLY ds.ft_balance_f ALTER COLUMN id SET DEFAULT nextval('ds.ft_balance_f_id_seq'::regclass);


--
-- TOC entry 4739 (class 2604 OID 20040)
-- Name: ft_posting_f id; Type: DEFAULT; Schema: ds; Owner: postgres
--

ALTER TABLE ONLY ds.ft_posting_f ALTER COLUMN id SET DEFAULT nextval('ds.ft_posting_f_id_seq'::regclass);


--
-- TOC entry 4740 (class 2604 OID 20047)
-- Name: md_account_d id; Type: DEFAULT; Schema: ds; Owner: postgres
--

ALTER TABLE ONLY ds.md_account_d ALTER COLUMN id SET DEFAULT nextval('ds.md_account_d_id_seq'::regclass);


--
-- TOC entry 4741 (class 2604 OID 20054)
-- Name: md_currency_d id; Type: DEFAULT; Schema: ds; Owner: postgres
--

ALTER TABLE ONLY ds.md_currency_d ALTER COLUMN id SET DEFAULT nextval('ds.md_currency_d_id_seq'::regclass);


--
-- TOC entry 4742 (class 2604 OID 20061)
-- Name: md_exchange_rate_d id; Type: DEFAULT; Schema: ds; Owner: postgres
--

ALTER TABLE ONLY ds.md_exchange_rate_d ALTER COLUMN id SET DEFAULT nextval('ds.md_exchange_rate_d_id_seq'::regclass);


--
-- TOC entry 4743 (class 2604 OID 20068)
-- Name: md_ledger_account_s id; Type: DEFAULT; Schema: ds; Owner: postgres
--

ALTER TABLE ONLY ds.md_ledger_account_s ALTER COLUMN id SET DEFAULT nextval('ds.md_ledger_account_s_id_seq'::regclass);


--
-- TOC entry 4737 (class 2604 OID 17927)
-- Name: etl_logs log_id; Type: DEFAULT; Schema: logs; Owner: postgres
--

ALTER TABLE ONLY logs.etl_logs ALTER COLUMN log_id SET DEFAULT nextval('logs.etl_logs_log_id_seq'::regclass);


--
-- TOC entry 4744 (class 2604 OID 20075)
-- Name: ft_balance_f id; Type: DEFAULT; Schema: raw; Owner: postgres
--

ALTER TABLE ONLY raw.ft_balance_f ALTER COLUMN id SET DEFAULT nextval('raw.ft_balance_f_id_seq'::regclass);


--
-- TOC entry 4745 (class 2604 OID 20104)
-- Name: ft_posting_f id; Type: DEFAULT; Schema: raw; Owner: postgres
--

ALTER TABLE ONLY raw.ft_posting_f ALTER COLUMN id SET DEFAULT nextval('raw.ft_posting_f_id_seq'::regclass);


--
-- TOC entry 4746 (class 2604 OID 20114)
-- Name: md_account_d id; Type: DEFAULT; Schema: raw; Owner: postgres
--

ALTER TABLE ONLY raw.md_account_d ALTER COLUMN id SET DEFAULT nextval('raw.md_account_d_id_seq'::regclass);


--
-- TOC entry 4747 (class 2604 OID 20128)
-- Name: md_currency_d id; Type: DEFAULT; Schema: raw; Owner: postgres
--

ALTER TABLE ONLY raw.md_currency_d ALTER COLUMN id SET DEFAULT nextval('raw.md_currency_d_id_seq'::regclass);


--
-- TOC entry 4748 (class 2604 OID 20139)
-- Name: md_exchange_rate_d id; Type: DEFAULT; Schema: raw; Owner: postgres
--

ALTER TABLE ONLY raw.md_exchange_rate_d ALTER COLUMN id SET DEFAULT nextval('raw.md_exchange_rate_d_id_seq'::regclass);


--
-- TOC entry 4749 (class 2604 OID 20149)
-- Name: md_ledger_account_s id; Type: DEFAULT; Schema: raw; Owner: postgres
--

ALTER TABLE ONLY raw.md_ledger_account_s ALTER COLUMN id SET DEFAULT nextval('raw.md_ledger_account_s_id_seq'::regclass);


--
-- TOC entry 4736 (class 2604 OID 17853)
-- Name: rejected_data rejected_id; Type: DEFAULT; Schema: raw; Owner: postgres
--

ALTER TABLE ONLY raw.rejected_data ALTER COLUMN rejected_id SET DEFAULT nextval('raw.rejected_data_rejected_id_seq'::regclass);


--
-- TOC entry 4917 (class 0 OID 17745)
-- Dependencies: 219
-- Data for Name: dm_account_turnover_f; Type: TABLE DATA; Schema: dm; Owner: postgres
--

-- данные


--
-- TOC entry 4918 (class 0 OID 17751)
-- Dependencies: 220
-- Data for Name: dm_f101_round_f; Type: TABLE DATA; Schema: dm; Owner: postgres
--

-- данные

--
-- TOC entry 4921 (class 0 OID 17838)
-- Dependencies: 223
-- Data for Name: lg_messages; Type: TABLE DATA; Schema: dm; Owner: postgres
--

-- данные

--
-- TOC entry 4931 (class 0 OID 20030)
-- Dependencies: 233
-- Data for Name: ft_balance_f; Type: TABLE DATA; Schema: ds; Owner: postgres
--

-- данные


--
-- TOC entry 4933 (class 0 OID 20037)
-- Dependencies: 235
-- Data for Name: ft_posting_f; Type: TABLE DATA; Schema: ds; Owner: postgres
--


-- данные


--
-- TOC entry 4935 (class 0 OID 20044)
-- Dependencies: 237
-- Data for Name: md_account_d; Type: TABLE DATA; Schema: ds; Owner: postgres
--

-- данные


--
-- TOC entry 4937 (class 0 OID 20051)
-- Dependencies: 239
-- Data for Name: md_currency_d; Type: TABLE DATA; Schema: ds; Owner: postgres
--

-- данные


--
-- TOC entry 4939 (class 0 OID 20058)
-- Dependencies: 241
-- Data for Name: md_exchange_rate_d; Type: TABLE DATA; Schema: ds; Owner: postgres
--

-- данные

--
-- TOC entry 4941 (class 0 OID 20065)
-- Dependencies: 243
-- Data for Name: md_ledger_account_s; Type: TABLE DATA; Schema: ds; Owner: postgres
--

-- данные


--
-- TOC entry 4929 (class 0 OID 19593)
-- Dependencies: 231
-- Data for Name: etl_log; Type: TABLE DATA; Schema: logs; Owner: postgres
--



--
-- TOC entry 4925 (class 0 OID 17924)
-- Dependencies: 227
-- Data for Name: etl_logs; Type: TABLE DATA; Schema: logs; Owner: postgres
--

-- данные

--
-- TOC entry 4927 (class 0 OID 19509)
-- Dependencies: 229
-- Data for Name: dm_f101_round_f; Type: TABLE DATA; Schema: public; Owner: postgres
--

-- данные

--
-- TOC entry 4919 (class 0 OID 17827)
-- Dependencies: 221
-- Data for Name: posting_summary; Type: TABLE DATA; Schema: public; Owner: postgres
--

-- данные

--
-- TOC entry 4943 (class 0 OID 20072)
-- Dependencies: 245
-- Data for Name: ft_balance_f; Type: TABLE DATA; Schema: raw; Owner: postgres
--

-- данные

--
-- TOC entry 4945 (class 0 OID 20077)
-- Dependencies: 247
-- Data for Name: ft_posting_f; Type: TABLE DATA; Schema: raw; Owner: postgres
--

-- данные


--
-- TOC entry 4947 (class 0 OID 20082)
-- Dependencies: 249
-- Data for Name: md_account_d; Type: TABLE DATA; Schema: raw; Owner: postgres
--

-- данные


--
-- TOC entry 4949 (class 0 OID 20087)
-- Dependencies: 251
-- Data for Name: md_currency_d; Type: TABLE DATA; Schema: raw; Owner: postgres
--

-- данные


--
-- TOC entry 4951 (class 0 OID 20092)
-- Dependencies: 253
-- Data for Name: md_exchange_rate_d; Type: TABLE DATA; Schema: raw; Owner: postgres
--


-- данные


--
-- TOC entry 4953 (class 0 OID 20097)
-- Dependencies: 255
-- Data for Name: md_ledger_account_s; Type: TABLE DATA; Schema: raw; Owner: postgres
--

-- данные

--
-- TOC entry 4923 (class 0 OID 17850)
-- Dependencies: 225
-- Data for Name: rejected_data; Type: TABLE DATA; Schema: raw; Owner: postgres
--



--
-- TOC entry 4975 (class 0 OID 0)
-- Dependencies: 222
-- Name: lg_messages_record_id_seq; Type: SEQUENCE SET; Schema: dm; Owner: postgres
--

SELECT pg_catalog.setval('dm.lg_messages_record_id_seq', 768, true);


--
-- TOC entry 4976 (class 0 OID 0)
-- Dependencies: 232
-- Name: ft_balance_f_id_seq; Type: SEQUENCE SET; Schema: ds; Owner: postgres
--

SELECT pg_catalog.setval('ds.ft_balance_f_id_seq', 1, false);


--
-- TOC entry 4977 (class 0 OID 0)
-- Dependencies: 234
-- Name: ft_posting_f_id_seq; Type: SEQUENCE SET; Schema: ds; Owner: postgres
--

SELECT pg_catalog.setval('ds.ft_posting_f_id_seq', 1, false);


--
-- TOC entry 4978 (class 0 OID 0)
-- Dependencies: 236
-- Name: md_account_d_id_seq; Type: SEQUENCE SET; Schema: ds; Owner: postgres
--

SELECT pg_catalog.setval('ds.md_account_d_id_seq', 1, false);


--
-- TOC entry 4979 (class 0 OID 0)
-- Dependencies: 238
-- Name: md_currency_d_id_seq; Type: SEQUENCE SET; Schema: ds; Owner: postgres
--

SELECT pg_catalog.setval('ds.md_currency_d_id_seq', 1, false);


--
-- TOC entry 4980 (class 0 OID 0)
-- Dependencies: 240
-- Name: md_exchange_rate_d_id_seq; Type: SEQUENCE SET; Schema: ds; Owner: postgres
--

SELECT pg_catalog.setval('ds.md_exchange_rate_d_id_seq', 1, false);


--
-- TOC entry 4981 (class 0 OID 0)
-- Dependencies: 242
-- Name: md_ledger_account_s_id_seq; Type: SEQUENCE SET; Schema: ds; Owner: postgres
--

SELECT pg_catalog.setval('ds.md_ledger_account_s_id_seq', 1, false);


--
-- TOC entry 4982 (class 0 OID 0)
-- Dependencies: 230
-- Name: etl_log_id_seq; Type: SEQUENCE SET; Schema: logs; Owner: postgres
--

SELECT pg_catalog.setval('logs.etl_log_id_seq', 1, false);


--
-- TOC entry 4983 (class 0 OID 0)
-- Dependencies: 226
-- Name: etl_logs_log_id_seq; Type: SEQUENCE SET; Schema: logs; Owner: postgres
--

SELECT pg_catalog.setval('logs.etl_logs_log_id_seq', 327, true);


--
-- TOC entry 4984 (class 0 OID 0)
-- Dependencies: 228
-- Name: dm_f101_round_f_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.dm_f101_round_f_id_seq', 1, false);


--
-- TOC entry 4985 (class 0 OID 0)
-- Dependencies: 244
-- Name: ft_balance_f_id_seq; Type: SEQUENCE SET; Schema: raw; Owner: postgres
--

SELECT pg_catalog.setval('raw.ft_balance_f_id_seq', 112, true);


--
-- TOC entry 4986 (class 0 OID 0)
-- Dependencies: 246
-- Name: ft_posting_f_id_seq; Type: SEQUENCE SET; Schema: raw; Owner: postgres
--

SELECT pg_catalog.setval('raw.ft_posting_f_id_seq', 33892, true);


--
-- TOC entry 4987 (class 0 OID 0)
-- Dependencies: 248
-- Name: md_account_d_id_seq; Type: SEQUENCE SET; Schema: raw; Owner: postgres
--

SELECT pg_catalog.setval('raw.md_account_d_id_seq', 112, true);


--
-- TOC entry 4988 (class 0 OID 0)
-- Dependencies: 250
-- Name: md_currency_d_id_seq; Type: SEQUENCE SET; Schema: raw; Owner: postgres
--

SELECT pg_catalog.setval('raw.md_currency_d_id_seq', 50, true);


--
-- TOC entry 4989 (class 0 OID 0)
-- Dependencies: 252
-- Name: md_exchange_rate_d_id_seq; Type: SEQUENCE SET; Schema: raw; Owner: postgres
--

SELECT pg_catalog.setval('raw.md_exchange_rate_d_id_seq', 9614, true);


--
-- TOC entry 4990 (class 0 OID 0)
-- Dependencies: 254
-- Name: md_ledger_account_s_id_seq; Type: SEQUENCE SET; Schema: raw; Owner: postgres
--

SELECT pg_catalog.setval('raw.md_ledger_account_s_id_seq', 36, true);


--
-- TOC entry 4991 (class 0 OID 0)
-- Dependencies: 224
-- Name: rejected_data_rejected_id_seq; Type: SEQUENCE SET; Schema: raw; Owner: postgres
--

SELECT pg_catalog.setval('raw.rejected_data_rejected_id_seq', 1, false);


--
-- TOC entry 4753 (class 2606 OID 17845)
-- Name: lg_messages lg_messages_pkey; Type: CONSTRAINT; Schema: dm; Owner: postgres
--

ALTER TABLE ONLY dm.lg_messages
    ADD CONSTRAINT lg_messages_pkey PRIMARY KEY (record_id);


--
-- TOC entry 4763 (class 2606 OID 20035)
-- Name: ft_balance_f ft_balance_f_pkey; Type: CONSTRAINT; Schema: ds; Owner: postgres
--

ALTER TABLE ONLY ds.ft_balance_f
    ADD CONSTRAINT ft_balance_f_pkey PRIMARY KEY (id, on_date, account_rk);


--
-- TOC entry 4765 (class 2606 OID 20042)
-- Name: ft_posting_f ft_posting_f_pkey; Type: CONSTRAINT; Schema: ds; Owner: postgres
--

ALTER TABLE ONLY ds.ft_posting_f
    ADD CONSTRAINT ft_posting_f_pkey PRIMARY KEY (id, oper_date, credit_account_rk, debet_account_rk);


--
-- TOC entry 4767 (class 2606 OID 20049)
-- Name: md_account_d md_account_d_pkey; Type: CONSTRAINT; Schema: ds; Owner: postgres
--

ALTER TABLE ONLY ds.md_account_d
    ADD CONSTRAINT md_account_d_pkey PRIMARY KEY (id, data_actual_date, account_rk);


--
-- TOC entry 4769 (class 2606 OID 20056)
-- Name: md_currency_d md_currency_d_pkey; Type: CONSTRAINT; Schema: ds; Owner: postgres
--

ALTER TABLE ONLY ds.md_currency_d
    ADD CONSTRAINT md_currency_d_pkey PRIMARY KEY (id, currency_rk, data_actual_date);


--
-- TOC entry 4771 (class 2606 OID 20063)
-- Name: md_exchange_rate_d md_exchange_rate_d_pkey; Type: CONSTRAINT; Schema: ds; Owner: postgres
--

ALTER TABLE ONLY ds.md_exchange_rate_d
    ADD CONSTRAINT md_exchange_rate_d_pkey PRIMARY KEY (id, data_actual_date, currency_rk);


--
-- TOC entry 4773 (class 2606 OID 20070)
-- Name: md_ledger_account_s md_ledger_account_s_pkey; Type: CONSTRAINT; Schema: ds; Owner: postgres
--

ALTER TABLE ONLY ds.md_ledger_account_s
    ADD CONSTRAINT md_ledger_account_s_pkey PRIMARY KEY (id, ledger_account, start_date);


--
-- TOC entry 4761 (class 2606 OID 19599)
-- Name: etl_log etl_log_pkey; Type: CONSTRAINT; Schema: logs; Owner: postgres
--

ALTER TABLE ONLY logs.etl_log
    ADD CONSTRAINT etl_log_pkey PRIMARY KEY (id);


--
-- TOC entry 4757 (class 2606 OID 17929)
-- Name: etl_logs etl_logs_pkey; Type: CONSTRAINT; Schema: logs; Owner: postgres
--

ALTER TABLE ONLY logs.etl_logs
    ADD CONSTRAINT etl_logs_pkey PRIMARY KEY (log_id);


--
-- TOC entry 4759 (class 2606 OID 19515)
-- Name: dm_f101_round_f dm_f101_round_f_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_f101_round_f
    ADD CONSTRAINT dm_f101_round_f_pkey PRIMARY KEY (id);


--
-- TOC entry 4751 (class 2606 OID 17831)
-- Name: posting_summary posting_summary_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.posting_summary
    ADD CONSTRAINT posting_summary_pkey PRIMARY KEY (date);


--
-- TOC entry 4755 (class 2606 OID 17857)
-- Name: rejected_data rejected_data_pkey; Type: CONSTRAINT; Schema: raw; Owner: postgres
--

ALTER TABLE ONLY raw.rejected_data
    ADD CONSTRAINT rejected_data_pkey PRIMARY KEY (rejected_id);


-- Completed on 2024-07-25 17:30:41

--
-- PostgreSQL database dump complete
--

