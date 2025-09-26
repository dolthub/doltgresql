-- Downloaded from: https://github.com/S0mbre/russtat/blob/ee53cd05152e75cce7855d73b2d78f8af083a22c/sql/dbcreate.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 13.0
-- Dumped by pg_dump version 13.0

-- Started on 2020-10-13 16:36:02

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

DROP DATABASE russtat;
--
-- TOC entry 3290 (class 1262 OID 16394)
-- Name: russtat; Type: DATABASE; Schema: -; Owner: -
--

CREATE DATABASE russtat WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE = 'Russian_Russia.1251';


\connect russtat

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
-- TOC entry 2 (class 3079 OID 16846)
-- Name: btree_gin; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS btree_gin WITH SCHEMA public;


--
-- TOC entry 3291 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION btree_gin; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION btree_gin IS 'support for indexing common datatypes in GIN';


--
-- TOC entry 328 (class 1255 OID 16561)
-- Name: add_data(text, text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.add_data(dataset_json text, time_format text DEFAULT 'YYYY-MM-DD HH24-MI-SS'::text, OUT n_added integer, OUT last_data_id bigint, OUT dataset_id integer) RETURNS record
    LANGUAGE plpgsql
    AS $$
declare
	data_json jsonb;
	prep_time timestamp with time zone;
	last_updated timestamp with time zone;
	next_update timestamp with time zone; 
	data_rec json;
	codes_id integer;
	codevals_id integer;
	units_id integer;
	periods_id integer;
	cnt1 bigint;
	cnt2 bigint;
	r text;
begin
	n_added := 0;
	last_data_id := -1;
	dataset_id := -1;
	
	data_json := dataset_json::jsonb;
	
	prep_time := to_timestamp(data_json->>'prepared', time_format);
	last_updated := to_timestamp(data_json->>'updated', time_format);
	next_update := to_timestamp(data_json->'periodicity'->>'next', time_format);
	
	-- add / update dataset
	
	/*
	raise notice 'ARGS TO add_dataset:\n'
				 'prep_time=%\nlast_updated=%\nnext_update=%\n'
				 'ds_id=%\nfullname=%\nagency_id=%\nagency_name=%\n'
				 'agency_dept=%\ncodes=%\nunit=%\n'
				 'periodicity=%\nthis_release=%\ndata_range=%\n'
				 'description=%\nclassifier_id=%\nclassifier_path=%\n'
				 'prep_by=%\nprep_contacts=%', 
				  prep_time, last_updated, next_update,
				  data_json->>'id', data_json->>'full_name',
				  data_json->>'agency_id', data_json->>'agency_name',
				  data_json->>'agency_dept', jsonb_pretty(data_json->'codes'),
				  data_json->>'unit', data_json->'periodicity'->>'value',
				  data_json->'periodicity'->>'releases',
				  array[cast(data_json->'data_range'->>0 as integer), 
						cast(data_json->'data_range'->>1 as integer)], 
				  data_json->>'methodology', 
				  data_json->'classifier'->>'id',
				  data_json->'classifier'->>'path',
				  data_json->'prepared_by'->>'name',
				  data_json->'prepared_by'->>'contacts';
	return;
	*/
	
	select into dataset_id public.add_dataset(prep_time, last_updated, next_update,
								  data_json->>'id', data_json->>'full_name',
								  data_json->>'agency_id', data_json->>'agency_name',
								  data_json->>'agency_dept', data_json->'codes',
								  data_json->>'unit', data_json->'periodicity'->>'value',
								  data_json->'periodicity'->>'releases',
								  array[cast(data_json->'data_range'->>0 as integer), 
										cast(data_json->'data_range'->>1 as integer)], 
								  data_json->>'methodology', 
								  data_json->'classifier'->>'id',
								  data_json->'classifier'->>'path',
								  data_json->'prepared_by'->>'name',
								  data_json->'prepared_by'->>'contacts');
	if dataset_id = -1 then
		raise notice '! Unable to add or update dataset!';
		return;
	end if;
	
	-- add observations
	select count(*) into cnt1 from obs;
	
	for data_rec in select * from jsonb_array_elements(data_json->'data')
	loop
		-- find code
		begin
			select id into strict codes_id from codes
			where lower(name) = lower(data_rec->>0) limit 1;
		exception
			when NO_DATA_FOUND then
			insert into codes(name) values (data_rec->>0)
				on conflict on constraint codes_unique1 do
				update set name = excluded.name 
				returning id into codes_id;
		end;
				
		-- find codeval
		begin
			select id into strict codevals_id from codevals
			where code_id = codes_id and lower(name) = lower(data_rec->>1) limit 1;
		exception 
			when NO_DATA_FOUND then
			insert into codevals(code_id, name) values (codes_id, data_rec->>1)
				on conflict on constraint codevals_unique do
				update set code_id = excluded.code_id, name = excluded.name
				returning id into codevals_id;
		end;
				
		-- find unit
		begin
			select id into strict units_id from units
			where lower(val) = lower(data_rec->>2) limit 1;
		exception
			when NO_DATA_FOUND then
			insert into units(val) values (data_rec->>2)
				on conflict on constraint units_unique do
				update set val = excluded.val 
				returning id into units_id;
		end;
				
		-- find period
		begin
			select id into strict periods_id from periods
			where lower(val) = lower(data_rec->>3) limit 1;
		exception
			when NO_DATA_FOUND then
			insert into periods(val) values (data_rec->>3)
				on conflict on constraint periods_unique do
				update set val = excluded.val 
				returning id into periods_id;
		end;
				
		-- insert data record
		
		/*
		raise notice 'dataset_id=%, code_id=%, unit_id=%, '
					 'period_id=%, obs_year=%, obs_val=%',
					 dataset_id, codevals_id, units_id, periods_id, 
					 data_rec->>4, data_rec->>5;
		continue;
		*/
		
		insert into obs(dataset_id, code_id, unit_id, period_id, obs_year, obs_val) 
			values (dataset_id, codevals_id, units_id, periods_id, 
					cast(data_rec->>4 as integer), cast(data_rec->>5 as real))
			on conflict on constraint obs_unique1 do
			update set dataset_id = excluded.dataset_id, code_id = excluded.code_id,
				unit_id = excluded.unit_id, period_id = excluded.period_id,
				obs_year = excluded.obs_year, obs_val = excluded.obs_val 
			returning id into last_data_id;
				
	end loop;
	
	select count(*) into cnt2 from obs;
	n_added := cnt2 - cnt1;
	
end;
$$;


--
-- TOC entry 329 (class 1255 OID 16554)
-- Name: add_dataset(timestamp with time zone, timestamp with time zone, timestamp with time zone, text, text, text, text, text, jsonb, text, text, text, integer[], text, text, text, text, text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.add_dataset(prep_time timestamp with time zone, last_updated timestamp with time zone, next_update timestamp with time zone, ds_id text, fullname text, agency_id text, agency_name text, agency_dept text, codes jsonb, unit text, periodicity text, this_release text, data_range integer[], description text, classifier_id text, classifier_path text, prep_by text, prep_contacts text) RETURNS integer
    LANGUAGE plpgsql
    AS $$
declare
	ag_id_ integer;
	dept_id_ integer;
	class_id_ integer;
	units_id_ integer;
	periods_id_ integer;
	codes_id_ integer;
	codename text;
	js_codeval jsonb;
	ds_id_ integer;
begin
	/*
	raise notice 'Agency ID = %\nAgency name = %\nDep = %\n'
				 'Cls ID = %\nCls path = %\nUnit = %\nPeriod = %\n'
				 'Prep = %\nUpdated = %\nNext = %\nRange = %\n'
				 'DS ID = %\nName = %\nDesc = %\n'
				 'By = %\nContact = %', agency_id, agency_name, agency_dept,
				 classifier_id, classifier_path, unit, periodicity,
				 prep_time, last_updated, next_update, data_range,
				 ds_id, fullname, description, prep_by, prep_contacts;
	return -1;
	*/

	-- agency
	insert into agencies(ag_id, name)
		values (agency_id, agency_name)
		on conflict on constraint agencies_unique1 do 
		update set name = excluded.name, ag_id = excluded.ag_id 
		returning id into ag_id_;	
	
	-- department
	insert into departments(agency_id, name)
		values (ag_id_, agency_dept)
		on conflict on constraint depts_unique1 do
		update set name = excluded.name 
		returning id into dept_id_;
		
	-- classifier
	insert into classifier(class_id, name)
		values (classifier_id, classifier_path)
		on conflict on constraint classifier_unique1 do
		update set name = excluded.name 
		returning id into class_id_;
		
	-- units
	insert into units(val) values (unit)
		on conflict on constraint units_unique do
		update set val = excluded.val 
		returning id into units_id_;
		
	-- periodicity
	insert into periods(val) values (periodicity)
		on conflict on constraint periods_unique do 
		update set val = excluded.val 
		returning id into periods_id_;
		
	-- codes
	-- {'code_id': {'name': 'code_name', 'values': [['val_id', 'name'], [...], ...]}, {...}, ...}
	for codename in select * from jsonb_object_keys(codes)
	loop
		insert into codes(name) values (codes->codename->>'name')
			on conflict on constraint codes_unique1 do
			update set name = excluded.name 
			returning id into codes_id_;

		for js_codeval in select * from jsonb_array_elements(codes->codename->'values')
		loop
			insert into codevals(code_id, val_id, name) 
				values (codes_id_, js_codeval->>0, js_codeval->>1)
				on conflict on constraint codevals_unique do
				update set name = excluded.name;
		end loop;
	end loop;
	
	-- datasets
	insert into datasets(prep_time, updated_time, next_update_time,
						ds_id, agency_id, dept_id, name, period_id,
						unit_id, range_start, range_end, class_id,
						description, prep_by, prep_contact, code_id)
		values (prep_time, last_updated, next_update,
			   ds_id, ag_id_, dept_id_, fullname, periods_id_,
			   units_id_, data_range[1], data_range[2], class_id_,
			   description, prep_by, prep_contacts, codes_id_)
		on conflict on constraint datasets_unique1 do
		update set prep_time = excluded.prep_time, updated_time = excluded.updated_time, 
			next_update_time = excluded.next_update_time, ds_id = excluded.ds_id, 
			agency_id = excluded.agency_id, dept_id = excluded.dept_id,
			name = excluded.name, period_id = excluded.period_id, unit_id = excluded.unit_id,
			range_start = excluded.range_start, range_end = excluded.range_end,
			class_id = excluded.class_id, description = excluded.description, 
			prep_by = excluded.prep_by, prep_contact = excluded.prep_contact, code_id = excluded.code_id			
		returning id into ds_id_;
		
	return ds_id_;
	
end;
$$;


--
-- TOC entry 322 (class 1255 OID 17282)
-- Name: agencies_update_insert(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.agencies_update_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
	new.search := to_tsvector('russian', coalesce(new.name, ''));
	return new;
end;
$$;


--
-- TOC entry 323 (class 1255 OID 17290)
-- Name: classifier_update_insert(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.classifier_update_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
	new.search := to_tsvector('russian', coalesce(new.name, ''));
	return new;
end;
$$;


--
-- TOC entry 333 (class 1255 OID 26809)
-- Name: clear_all(boolean); Type: PROCEDURE; Schema: public; Owner: -
--

CREATE PROCEDURE public.clear_all(fullclear boolean DEFAULT true)
    LANGUAGE plpgsql
    AS $$
begin

	-- delete views
	drop view public.all_data;
	drop view public.all_datasets;

	-- delete obs, codevals, codes, datasets
	drop table public.obs;
	drop table public.codevals;
	drop table public.codes;
	drop table public.datasets;
		
	-- recreate codes
	CREATE TABLE public.codes (
		id integer NOT NULL,
		name text NOT NULL,
		search tsvector);	
	ALTER TABLE public.codes ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
		SEQUENCE NAME public.codes_id_seq
		START WITH 1
		INCREMENT BY 1
		NO MINVALUE
		NO MAXVALUE
		CACHE 1);		
	ALTER TABLE ONLY public.codes
    	ADD CONSTRAINT codes_pkey PRIMARY KEY (id);		
	ALTER TABLE ONLY public.codes
    	ADD CONSTRAINT codes_unique1 UNIQUE (name) INCLUDE (name);		
	CREATE INDEX codes_search_idx ON public.codes USING gin (search);	
	CREATE TRIGGER codes_tr1 BEFORE INSERT OR UPDATE ON public.codes 
		FOR EACH ROW EXECUTE FUNCTION public.codes_update_insert();
	
	-- recreate codevals
	CREATE TABLE public.codevals (
		id integer NOT NULL,
		code_id integer NOT NULL,
		val_id character varying(64) DEFAULT ''::character varying NOT NULL,
		name text NOT NULL,
		search tsvector
	);
	ALTER TABLE public.codevals ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
		SEQUENCE NAME public.codevals_id_seq
		START WITH 1
		INCREMENT BY 1
		NO MINVALUE
		NO MAXVALUE
		CACHE 1);
	ALTER TABLE ONLY public.codevals
    	ADD CONSTRAINT codevals_pkey PRIMARY KEY (id);
	ALTER TABLE ONLY public.codevals
    	ADD CONSTRAINT codevals_unique UNIQUE (code_id, val_id);
	ALTER TABLE ONLY public.codevals
    	ADD CONSTRAINT codevals_fk1 FOREIGN KEY (code_id) REFERENCES public.codes(id) 
		ON UPDATE CASCADE ON DELETE CASCADE;
	CREATE INDEX codevals_search_idx ON public.codevals USING gin (search);
	CREATE TRIGGER codevals_tr1 BEFORE INSERT OR UPDATE ON public.codevals 
		FOR EACH ROW EXECUTE FUNCTION public.codevals_update_insert();
		
	-- recreate datasets
	CREATE TABLE public.datasets (
		id integer NOT NULL,
		prep_time timestamp with time zone,
		updated_time timestamp with time zone,
		next_update_time timestamp with time zone,
		ds_id text,
		agency_id integer NOT NULL,
		dept_id integer NOT NULL,
		name text NOT NULL,
		unit_id integer,
		range_start smallint,
		range_end smallint,
		class_id integer,
		description text,
		prep_by text,
		prep_contact text,
		code_id integer,
		period_id integer,
		search tsvector
	);
	ALTER TABLE public.datasets ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
		SEQUENCE NAME public.datasets_id_seq
		START WITH 1
		INCREMENT BY 1
		NO MINVALUE
		NO MAXVALUE
		CACHE 1);
	ALTER TABLE ONLY public.datasets
    	ADD CONSTRAINT datasets_pkey PRIMARY KEY (id);
	ALTER TABLE ONLY public.datasets
    	ADD CONSTRAINT datasets_unique1 UNIQUE (agency_id, name, dept_id);
	ALTER TABLE ONLY public.datasets
    	ADD CONSTRAINT datasets_fk1 FOREIGN KEY (agency_id) REFERENCES public.agencies(id) 
		ON UPDATE CASCADE ON DELETE CASCADE;
	ALTER TABLE ONLY public.datasets
    	ADD CONSTRAINT datasets_fk2 FOREIGN KEY (dept_id) REFERENCES public.departments(id) 
		ON UPDATE CASCADE ON DELETE CASCADE;
	ALTER TABLE ONLY public.datasets
    	ADD CONSTRAINT datasets_fk3 FOREIGN KEY (unit_id) REFERENCES public.units(id) 
		ON UPDATE CASCADE ON DELETE SET NULL;
	ALTER TABLE ONLY public.datasets
    	ADD CONSTRAINT datasets_fk4 FOREIGN KEY (class_id) REFERENCES public.classifier(id) 
		ON UPDATE CASCADE ON DELETE SET NULL;
	ALTER TABLE ONLY public.datasets
    	ADD CONSTRAINT datasets_fk5 FOREIGN KEY (code_id) REFERENCES public.codes(id) 
		ON UPDATE CASCADE ON DELETE SET NULL;
	ALTER TABLE ONLY public.datasets
    	ADD CONSTRAINT datasets_fk6 FOREIGN KEY (period_id) REFERENCES public.periods(id) 
		ON UPDATE CASCADE ON DELETE SET NULL;
	CREATE INDEX datasets_search_idx ON public.datasets USING gin (search);
	CREATE TRIGGER datasets_tr1 BEFORE INSERT OR UPDATE ON public.datasets 
		FOR EACH ROW EXECUTE FUNCTION public.datasets_update_insert();
	
	-- recreate obs	
	CREATE TABLE public.obs (
		id bigint NOT NULL,
		dataset_id integer NOT NULL,
		code_id integer NOT NULL,
		unit_id integer NOT NULL,
		period_id integer NOT NULL,
		obs_year integer NOT NULL,
		obs_val real NOT NULL
	);
	ALTER TABLE public.obs ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
		SEQUENCE NAME public.obs_id_seq
		START WITH 1
		INCREMENT BY 1
		NO MINVALUE
		NO MAXVALUE
		CACHE 1);
	ALTER TABLE ONLY public.obs
    	ADD CONSTRAINT obs_pkey PRIMARY KEY (id);
	ALTER TABLE ONLY public.obs
    	ADD CONSTRAINT obs_unique1 UNIQUE (dataset_id, code_id, unit_id, period_id, obs_year);
	ALTER TABLE ONLY public.obs
    	ADD CONSTRAINT obs_fk1 FOREIGN KEY (dataset_id) REFERENCES public.datasets(id) 
		ON UPDATE CASCADE ON DELETE CASCADE;
	ALTER TABLE ONLY public.obs
    	ADD CONSTRAINT obs_fk2 FOREIGN KEY (code_id) REFERENCES public.codevals(id) 
		ON UPDATE CASCADE ON DELETE CASCADE;
	ALTER TABLE ONLY public.obs
    	ADD CONSTRAINT obs_fk3 FOREIGN KEY (unit_id) REFERENCES public.units(id) 
		ON UPDATE CASCADE ON DELETE CASCADE;
	ALTER TABLE ONLY public.obs
    	ADD CONSTRAINT obs_fk4 FOREIGN KEY (period_id) REFERENCES public.periods(id) 
		ON UPDATE CASCADE ON DELETE CASCADE;
		
	-- clear other tables if "fullclear" == True
	if fullclear then
		delete from public.classifier;
		delete from public.units;
		delete from public.periods;
		delete from public.departments;
		delete from public.agencies;
		
		alter sequence classifier_id_seq start with 1;
		alter sequence units_id_seq start with 1;
		alter sequence periods_id_seq start with 1;
		alter sequence departments_id_seq start with 1;
		alter sequence agencies_id_seq start with 1;
	end if;
	
	-- recreate views
	CREATE OR REPLACE VIEW public.all_data AS
	SELECT 
		ds.name AS dataset,
		cls.name AS classifier,
		ds.description,
		ds.prep_time AS prepared,
		ds.updated_time AS updated,
		ds.next_update_time AS next_update,
		ag.name AS agency,
		dept.name AS department,
		ds.range_start AS data_start,
		ds.range_end AS data_end,
		ds.prep_by,
		ds.prep_contact,
		obs.obs_year AS year,
		per.val AS release,
		units.val AS unit,
		codes.name AS code,
		codevals.name AS category,
		obs.obs_val AS value
	FROM obs
		JOIN datasets ds ON obs.dataset_id = ds.id
		JOIN classifier cls ON ds.class_id = cls.id
		JOIN agencies ag ON ds.agency_id = ag.id
		JOIN departments dept ON ds.dept_id = dept.id
		JOIN periods per ON obs.period_id = per.id
		JOIN units ON obs.unit_id = units.id
		JOIN codevals ON obs.code_id = codevals.id
		JOIN codes ON codevals.code_id = codes.id
	ORDER BY cls.name, ds.name, codes.name, codevals.name, obs.obs_year, per.val;
  
	CREATE OR REPLACE VIEW public.all_datasets AS
	SELECT 
		ds.id,
		cls.name AS classifier,
		ds.name AS dataset,
		ds.updated_time AS updated,
		ds.prep_time AS prepared,
		ds.next_update_time AS next_update,
		ds.description,
		ag.name AS agency,
		dept.name AS department,
		ds.range_start AS data_start,
		ds.range_end AS data_end,
		ds.prep_by,
		ds.prep_contact
	FROM datasets ds
		JOIN classifier cls ON ds.class_id = cls.id
		JOIN agencies ag ON ds.agency_id = ag.id
		JOIN departments dept ON ds.dept_id = dept.id
	ORDER BY cls.name, ds.name;
	
end;
$$;


--
-- TOC entry 324 (class 1255 OID 17295)
-- Name: codes_update_insert(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.codes_update_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
	new.search := to_tsvector('russian', coalesce(new.name, ''));
	return new;
end;
$$;


--
-- TOC entry 325 (class 1255 OID 17301)
-- Name: codevals_update_insert(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.codevals_update_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
	new.search := to_tsvector('russian', coalesce(new.name, ''));
	return new;
end;
$$;


--
-- TOC entry 331 (class 1255 OID 17655)
-- Name: datasets_update_insert(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.datasets_update_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
	new.description := public.entity2char(trim(replace(new.description, '&nbsp;', ' ')));
	new.search := setweight(to_tsvector('russian', coalesce(new.name, '')), 'A') ||
			      setweight(to_tsvector('russian', coalesce(new.description, '')), 'B');
	return new;
end;
$$;


--
-- TOC entry 326 (class 1255 OID 17660)
-- Name: departments_update_insert(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.departments_update_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
	new.search := to_tsvector('russian', coalesce(new.name, ''));
	return new;
end;
$$;


--
-- TOC entry 335 (class 1255 OID 27842)
-- Name: disable_triggers(); Type: PROCEDURE; Schema: public; Owner: -
--

CREATE PROCEDURE public.disable_triggers()
    LANGUAGE plpgsql
    AS $$
begin
	alter table only public.agencies disable trigger agencies_tr1;
	alter table only public.classifier disable trigger classifier_tr1;
	alter table only public.codes disable trigger codes_tr1;
	alter table only public.codevals disable trigger codevals_tr1;
	alter table only public.datasets disable trigger datasets_tr1;
	alter table only public.departments disable trigger departments_tr1;
	alter table only public.periods disable trigger periods_tr1;
	alter table only public.units disable trigger units_tr1;
end;
$$;


--
-- TOC entry 336 (class 1255 OID 27843)
-- Name: enable_triggers(boolean); Type: PROCEDURE; Schema: public; Owner: -
--

CREATE PROCEDURE public.enable_triggers(recalc boolean DEFAULT true)
    LANGUAGE plpgsql
    AS $$
begin
	alter table only public.agencies enable trigger agencies_tr1;
	alter table only public.classifier enable trigger classifier_tr1;
	alter table only public.codes enable trigger codes_tr1;
	alter table only public.codevals enable trigger codevals_tr1;
	alter table only public.datasets enable trigger datasets_tr1;
	alter table only public.departments enable trigger departments_tr1;
	alter table only public.periods enable trigger periods_tr1;
	alter table only public.units enable trigger units_tr1;
	
	if recalc then
	
		update public.agencies
		set search = to_tsvector('russian', coalesce(name, ''))
		where search is null;
		
		update public.classifier
		set search = to_tsvector('russian', coalesce(name, ''))
		where search is null;
		
		update public.codes
		set search = to_tsvector('russian', coalesce(name, ''))
		where search is null;
		
		update public.codevals
		set search = to_tsvector('russian', coalesce(name, ''))
		where search is null;
		
		update public.datasets
		set description = public.entity2char(trim(replace(description, '&nbsp;', ' ')));
		
		update public.datasets
		set search = setweight(to_tsvector('russian', coalesce(name, '')), 'A') ||
			         setweight(to_tsvector('russian', coalesce(description, '')), 'B')
		where search is null;
		
		update public.departments
		set search = to_tsvector('russian', coalesce(name, ''))
		where search is null;
		
		update public.periods
		set search = to_tsvector('russian', coalesce(val, ''))
		where search is null;
		
		update public.units
		set search = to_tsvector('russian', coalesce(val, ''))
		where search is null;
		
	end if;
end;
$$;


--
-- TOC entry 330 (class 1255 OID 26535)
-- Name: entity2char(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.entity2char(t text) RETURNS text
    LANGUAGE plpgsql IMMUTABLE
    AS $$
declare
    r record;
begin
    for r in
        select distinct ce.ch, ce.name
        from
            character_entity ce
            inner join (
                select name[1] "name"
                from regexp_matches(t, '&([A-Za-z]+?);', 'g') r(name)
            ) s on ce.name = s.name
    loop
        t := replace(t, '&' || r.name || ';', r.ch);
    end loop;

    for r in
        select distinct
            hex[1] hex,
            ('x' || repeat('0', 8 - length(hex[1])) || hex[1])::bit(32)::int codepoint
        from regexp_matches(t, '&#x([0-9a-f]{1,8}?);', 'gi') s(hex)
    loop
        t := regexp_replace(t, '&#x' || r.hex || ';', chr(r.codepoint), 'gi');
    end loop;

    for r in
        select distinct
            chr(codepoint[1]::int) ch,
            codepoint[1] codepoint
        from regexp_matches(t, '&#([0-9]{1,10}?);', 'g') s(codepoint)
    loop
        t := replace(t, '&#' || r.codepoint || ';', r.ch);
    end loop;

    return t;
end;
$$;


--
-- TOC entry 327 (class 1255 OID 17666)
-- Name: periods_update_insert(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.periods_update_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
	new.search := to_tsvector('russian', coalesce(new.val, ''));
	return new;
end;
$$;


--
-- TOC entry 233 (class 1255 OID 16530)
-- Name: resort_classifier_fn(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.resort_classifier_fn() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
declare
	class_row RECORD;
	last_pos integer;
	parent_part text;
	pid integer;
begin
	-- raise notice 'Searching all records with parent_id = -1...';
	for class_row in 
		select id as id_, name as name_ 
		from classifier where parent_id = -1
	loop
		-- raise notice 'Found item: %', class_row.name_;
		last_pos := length(class_row.name_) - position('/' in reverse(class_row.name_));
		parent_part := rtrim(left(class_row.name_, last_pos));
		if length(parent_part) = length(class_row.name_) then
			update classifier set parent_id = null 
			where classifier.id = class_row.id_;
			continue;
		end if;
		select id into pid from classifier where name = parent_part;
		if found then
			-- raise notice 'Found parent: %', pid;
			update classifier set parent_id = pid 
			where classifier.id = class_row.id_;
		end if;
	end loop;	
	
	return null;
end;
$$;


--
-- TOC entry 334 (class 1255 OID 27831)
-- Name: search_data(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.search_data(pattern text) RETURNS TABLE(id bigint, classificator text, dsname text, description text, preptime timestamp with time zone, updated timestamp with time zone, nextupdate timestamp with time zone, agency text, department text, startyr smallint, endyr smallint, prepby text, contact text, obsyear integer, obsperiod character varying, obsunit character varying, obscode text, obscodeval text, value real, ranking real)
    LANGUAGE plpgsql
    AS $$
begin 
	RETURN QUERY 
		SELECT obs.id, cls.name, ds.name, ds.description, ds.prep_time, ds.updated_time, ds.next_update_time,
				ag.name, dept.name, ds.range_start, ds.range_end, ds.prep_by, ds.prep_contact,
				obs.obs_year, per.val, units.val, codes.name, codevals.name, obs.obs_val::real,
				ts_rank((ds.search || cls.search || ag.search || dept.search || per.search || units.search || codevals.search || codes.search), 
						to_tsquery('russian', pattern)) as ranking
		FROM public.obs
			JOIN public.datasets ds ON obs.dataset_id = ds.id
			JOIN public.classifier cls ON ds.class_id = cls.id
			JOIN public.agencies ag ON ds.agency_id = ag.id
			JOIN public.departments dept ON ds.dept_id = dept.id
			JOIN public.periods per ON obs.period_id = per.id
			JOIN public.units ON obs.unit_id = units.id
			JOIN public.codevals ON obs.code_id = codevals.id
			JOIN public.codes ON codevals.code_id = codes.id
		WHERE 
			(ds.search || cls.search || ag.search || dept.search || per.search || units.search || codevals.search || codes.search) @@ to_tsquery('russian', pattern)
		ORDER BY ranking, cls.name, ds.name, codes.name, codevals.name, obs.obs_year, per.val;
  	
end;
$$;


--
-- TOC entry 332 (class 1255 OID 27829)
-- Name: search_datasets(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.search_datasets(pattern text) RETURNS TABLE(id integer, classificator text, dsname text, updated timestamp with time zone, preptime timestamp with time zone, nextupdate timestamp with time zone, description text, agency text, department text, startyr smallint, endyr smallint, prepby text, contact text, ranking real)
    LANGUAGE plpgsql
    AS $$
begin 
	RETURN QUERY
	SELECT ds.id, cls.name, ds.name, ds.updated_time, ds.prep_time, ds.next_update_time,
			ds.description, ag.name, dept.name, ds.range_start, ds.range_end,
			ds.prep_by, ds.prep_contact,
			ts_rank((ds.search || cls.search || ag.search || dept.search), 
					to_tsquery('russian', pattern)) as Ranking
	FROM public.datasets ds
		JOIN public.classifier cls ON ds.class_id = cls.id
		JOIN public.agencies ag ON ds.agency_id = ag.id
		JOIN public.departments dept ON ds.dept_id = dept.id
	WHERE 
		(ds.search || cls.search || ag.search || dept.search) @@ to_tsquery('russian', pattern)
	ORDER BY 
		Ranking desc, cls.name, ds.name;
end;
$$;


--
-- TOC entry 234 (class 1255 OID 16559)
-- Name: txt2jsonb(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.txt2jsonb(text_json text, OUT jsonb_obj jsonb) RETURNS jsonb
    LANGUAGE plpgsql
    AS $$
begin
	jsonb_obj := to_jsonb(text_json);
end;
$$;


SET default_table_access_method = heap;

--
-- TOC entry 202 (class 1259 OID 16397)
-- Name: agencies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.agencies (
    id integer NOT NULL,
    ag_id character varying(32),
    name text NOT NULL,
    search tsvector
);


--
-- TOC entry 201 (class 1259 OID 16395)
-- Name: agencies_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.agencies ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.agencies_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 206 (class 1259 OID 16419)
-- Name: classifier; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.classifier (
    id integer NOT NULL,
    class_id character varying(32),
    name text NOT NULL,
    parent_id integer DEFAULT '-1'::integer,
    search tsvector
);


--
-- TOC entry 211 (class 1259 OID 24912)
-- Name: codes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.codes (
    id integer NOT NULL,
    name text NOT NULL,
    search tsvector
);


--
-- TOC entry 213 (class 1259 OID 24920)
-- Name: codevals; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.codevals (
    id integer NOT NULL,
    code_id integer NOT NULL,
    val_id character varying(64) DEFAULT ''::character varying NOT NULL,
    name text NOT NULL,
    search tsvector
);


--
-- TOC entry 215 (class 1259 OID 24946)
-- Name: datasets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.datasets (
    id integer NOT NULL,
    prep_time timestamp with time zone,
    updated_time timestamp with time zone,
    next_update_time timestamp with time zone,
    ds_id text,
    agency_id integer NOT NULL,
    dept_id integer NOT NULL,
    name text NOT NULL,
    unit_id integer,
    range_start smallint,
    range_end smallint,
    class_id integer,
    description text,
    prep_by text,
    prep_contact text,
    code_id integer,
    period_id integer,
    search tsvector
);


--
-- TOC entry 204 (class 1259 OID 16407)
-- Name: departments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.departments (
    id integer NOT NULL,
    agency_id integer NOT NULL,
    name text NOT NULL,
    search tsvector
);


--
-- TOC entry 217 (class 1259 OID 24980)
-- Name: obs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.obs (
    id bigint NOT NULL,
    dataset_id integer NOT NULL,
    code_id integer NOT NULL,
    unit_id integer NOT NULL,
    period_id integer NOT NULL,
    obs_year integer NOT NULL,
    obs_val real NOT NULL
);


--
-- TOC entry 210 (class 1259 OID 16455)
-- Name: periods; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.periods (
    id integer NOT NULL,
    val character varying(256) NOT NULL,
    search tsvector
);


--
-- TOC entry 208 (class 1259 OID 16448)
-- Name: units; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.units (
    id integer NOT NULL,
    val character varying(256) NOT NULL,
    search tsvector
);


--
-- TOC entry 221 (class 1259 OID 28942)
-- Name: all_data; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.all_data AS
 SELECT ds.name AS dataset,
    cls.name AS classifier,
    ds.description,
    ds.prep_time AS prepared,
    ds.updated_time AS updated,
    ds.next_update_time AS next_update,
    ag.name AS agency,
    dept.name AS department,
    ds.range_start AS data_start,
    ds.range_end AS data_end,
    ds.prep_by,
    ds.prep_contact,
    obs.obs_year AS year,
    per.val AS release,
    units.val AS unit,
    codes.name AS code,
    codevals.name AS category,
    obs.obs_val AS value
   FROM ((((((((public.obs
     JOIN public.datasets ds ON ((obs.dataset_id = ds.id)))
     JOIN public.classifier cls ON ((ds.class_id = cls.id)))
     JOIN public.agencies ag ON ((ds.agency_id = ag.id)))
     JOIN public.departments dept ON ((ds.dept_id = dept.id)))
     JOIN public.periods per ON ((obs.period_id = per.id)))
     JOIN public.units ON ((obs.unit_id = units.id)))
     JOIN public.codevals ON ((obs.code_id = codevals.id)))
     JOIN public.codes ON ((codevals.code_id = codes.id)))
  ORDER BY cls.name, ds.name, codes.name, codevals.name, obs.obs_year, per.val;


--
-- TOC entry 220 (class 1259 OID 28935)
-- Name: all_datasets; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.all_datasets AS
 SELECT ds.id,
    cls.name AS classifier,
    ds.name AS dataset,
    ds.updated_time AS updated,
    ds.prep_time AS prepared,
    ds.next_update_time AS next_update,
    ds.description,
    ag.name AS agency,
    dept.name AS department,
    ds.range_start AS data_start,
    ds.range_end AS data_end,
    ds.prep_by,
    ds.prep_contact
   FROM (((public.datasets ds
     JOIN public.classifier cls ON ((ds.class_id = cls.id)))
     JOIN public.agencies ag ON ((ds.agency_id = ag.id)))
     JOIN public.departments dept ON ((ds.dept_id = dept.id)))
  ORDER BY cls.name, ds.name;


--
-- TOC entry 219 (class 1259 OID 26525)
-- Name: character_entity; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.character_entity (
    name text NOT NULL,
    ch character(1)
);


--
-- TOC entry 205 (class 1259 OID 16417)
-- Name: classifier_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.classifier ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.classifier_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 212 (class 1259 OID 24918)
-- Name: codes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.codes ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.codes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 214 (class 1259 OID 24927)
-- Name: codevals_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.codevals ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.codevals_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 216 (class 1259 OID 24952)
-- Name: datasets_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.datasets ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.datasets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 203 (class 1259 OID 16405)
-- Name: departments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.departments ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.departments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 218 (class 1259 OID 24983)
-- Name: obs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.obs ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.obs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 209 (class 1259 OID 16453)
-- Name: periods_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.periods ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.periods_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 207 (class 1259 OID 16446)
-- Name: units_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.units ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.units_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 3085 (class 2606 OID 16404)
-- Name: agencies agencies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agencies
    ADD CONSTRAINT agencies_pkey PRIMARY KEY (id);


--
-- TOC entry 3088 (class 2606 OID 16626)
-- Name: agencies agencies_unique1; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agencies
    ADD CONSTRAINT agencies_unique1 UNIQUE (name) INCLUDE (name);


--
-- TOC entry 3129 (class 2606 OID 26534)
-- Name: character_entity character_entity_ch_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.character_entity
    ADD CONSTRAINT character_entity_ch_key UNIQUE (ch);


--
-- TOC entry 3131 (class 2606 OID 26532)
-- Name: character_entity character_entity_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.character_entity
    ADD CONSTRAINT character_entity_pkey PRIMARY KEY (name);


--
-- TOC entry 3095 (class 2606 OID 16426)
-- Name: classifier classifier_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.classifier
    ADD CONSTRAINT classifier_pkey PRIMARY KEY (id);


--
-- TOC entry 3098 (class 2606 OID 16628)
-- Name: classifier classifier_unique1; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.classifier
    ADD CONSTRAINT classifier_unique1 UNIQUE (name) INCLUDE (name);


--
-- TOC entry 3110 (class 2606 OID 24930)
-- Name: codes codes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.codes
    ADD CONSTRAINT codes_pkey PRIMARY KEY (id);


--
-- TOC entry 3113 (class 2606 OID 24932)
-- Name: codes codes_unique1; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.codes
    ADD CONSTRAINT codes_unique1 UNIQUE (name) INCLUDE (name);


--
-- TOC entry 3115 (class 2606 OID 24936)
-- Name: codevals codevals_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.codevals
    ADD CONSTRAINT codevals_pkey PRIMARY KEY (id);


--
-- TOC entry 3118 (class 2606 OID 26813)
-- Name: codevals codevals_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.codevals
    ADD CONSTRAINT codevals_unique UNIQUE (code_id, val_id);


--
-- TOC entry 3120 (class 2606 OID 24955)
-- Name: datasets datasets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.datasets
    ADD CONSTRAINT datasets_pkey PRIMARY KEY (id);


--
-- TOC entry 3123 (class 2606 OID 54694)
-- Name: datasets datasets_unique1; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.datasets
    ADD CONSTRAINT datasets_unique1 UNIQUE (agency_id, name, dept_id);


--
-- TOC entry 3090 (class 2606 OID 16411)
-- Name: departments departments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.departments
    ADD CONSTRAINT departments_pkey PRIMARY KEY (id);


--
-- TOC entry 3093 (class 2606 OID 16635)
-- Name: departments depts_unique1; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.departments
    ADD CONSTRAINT depts_unique1 UNIQUE (agency_id, name);


--
-- TOC entry 3125 (class 2606 OID 24986)
-- Name: obs obs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.obs
    ADD CONSTRAINT obs_pkey PRIMARY KEY (id);


--
-- TOC entry 3127 (class 2606 OID 24988)
-- Name: obs obs_unique1; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.obs
    ADD CONSTRAINT obs_unique1 UNIQUE (dataset_id, code_id, unit_id, period_id, obs_year);


--
-- TOC entry 3105 (class 2606 OID 16459)
-- Name: periods periods_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.periods
    ADD CONSTRAINT periods_pkey PRIMARY KEY (id);


--
-- TOC entry 3108 (class 2606 OID 16640)
-- Name: periods periods_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.periods
    ADD CONSTRAINT periods_unique UNIQUE (val) INCLUDE (val);


--
-- TOC entry 3100 (class 2606 OID 16452)
-- Name: units units_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.units
    ADD CONSTRAINT units_pkey PRIMARY KEY (id);


--
-- TOC entry 3103 (class 2606 OID 16642)
-- Name: units units_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.units
    ADD CONSTRAINT units_unique UNIQUE (val) INCLUDE (val);


--
-- TOC entry 3086 (class 1259 OID 17284)
-- Name: agencies_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX agencies_search_idx ON public.agencies USING gin (search);


--
-- TOC entry 3096 (class 1259 OID 17287)
-- Name: classifier_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX classifier_search_idx ON public.classifier USING gin (search);


--
-- TOC entry 3111 (class 1259 OID 24933)
-- Name: codes_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX codes_search_idx ON public.codes USING gin (search);


--
-- TOC entry 3116 (class 1259 OID 24944)
-- Name: codevals_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX codevals_search_idx ON public.codevals USING gin (search);


--
-- TOC entry 3121 (class 1259 OID 24978)
-- Name: datasets_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX datasets_search_idx ON public.datasets USING gin (search);


--
-- TOC entry 3091 (class 1259 OID 17657)
-- Name: departments_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX departments_search_idx ON public.departments USING gin (search);


--
-- TOC entry 3106 (class 1259 OID 17665)
-- Name: periods_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX periods_search_idx ON public.periods USING gin (search);


--
-- TOC entry 3101 (class 1259 OID 17671)
-- Name: units_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX units_search_idx ON public.units USING gin (search);


--
-- TOC entry 3144 (class 2620 OID 17283)
-- Name: agencies agencies_tr1; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER agencies_tr1 BEFORE INSERT OR UPDATE ON public.agencies FOR EACH ROW EXECUTE FUNCTION public.agencies_update_insert();


--
-- TOC entry 3146 (class 2620 OID 16531)
-- Name: classifier classifier_on_update; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER classifier_on_update AFTER INSERT OR UPDATE ON public.classifier FOR EACH ROW EXECUTE FUNCTION public.resort_classifier_fn();

ALTER TABLE public.classifier DISABLE TRIGGER classifier_on_update;


--
-- TOC entry 3147 (class 2620 OID 17291)
-- Name: classifier classifier_tr1; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER classifier_tr1 BEFORE INSERT OR UPDATE ON public.classifier FOR EACH ROW EXECUTE FUNCTION public.classifier_update_insert();


--
-- TOC entry 3150 (class 2620 OID 24934)
-- Name: codes codes_tr1; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER codes_tr1 BEFORE INSERT OR UPDATE ON public.codes FOR EACH ROW EXECUTE FUNCTION public.codes_update_insert();


--
-- TOC entry 3151 (class 2620 OID 24945)
-- Name: codevals codevals_tr1; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER codevals_tr1 BEFORE INSERT OR UPDATE ON public.codevals FOR EACH ROW EXECUTE FUNCTION public.codevals_update_insert();


--
-- TOC entry 3152 (class 2620 OID 24979)
-- Name: datasets datasets_tr1; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER datasets_tr1 BEFORE INSERT OR UPDATE ON public.datasets FOR EACH ROW EXECUTE FUNCTION public.datasets_update_insert();


--
-- TOC entry 3145 (class 2620 OID 17661)
-- Name: departments departments_tr1; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER departments_tr1 BEFORE INSERT OR UPDATE ON public.departments FOR EACH ROW EXECUTE FUNCTION public.departments_update_insert();


--
-- TOC entry 3149 (class 2620 OID 17667)
-- Name: periods periods_tr1; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER periods_tr1 BEFORE INSERT OR UPDATE ON public.periods FOR EACH ROW EXECUTE FUNCTION public.periods_update_insert();


--
-- TOC entry 3148 (class 2620 OID 17672)
-- Name: units units_tr1; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER units_tr1 BEFORE INSERT OR UPDATE ON public.units FOR EACH ROW EXECUTE FUNCTION public.periods_update_insert();


--
-- TOC entry 3133 (class 2606 OID 58980)
-- Name: codevals codevals_fk1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.codevals
    ADD CONSTRAINT codevals_fk1 FOREIGN KEY (code_id) REFERENCES public.codes(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3136 (class 2606 OID 58955)
-- Name: datasets datasets_fk1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.datasets
    ADD CONSTRAINT datasets_fk1 FOREIGN KEY (agency_id) REFERENCES public.agencies(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3137 (class 2606 OID 58960)
-- Name: datasets datasets_fk2; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.datasets
    ADD CONSTRAINT datasets_fk2 FOREIGN KEY (dept_id) REFERENCES public.departments(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3134 (class 2606 OID 24968)
-- Name: datasets datasets_fk3; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.datasets
    ADD CONSTRAINT datasets_fk3 FOREIGN KEY (unit_id) REFERENCES public.units(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 3135 (class 2606 OID 24973)
-- Name: datasets datasets_fk4; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.datasets
    ADD CONSTRAINT datasets_fk4 FOREIGN KEY (class_id) REFERENCES public.classifier(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 3138 (class 2606 OID 58965)
-- Name: datasets datasets_fk5; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.datasets
    ADD CONSTRAINT datasets_fk5 FOREIGN KEY (code_id) REFERENCES public.codes(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 3139 (class 2606 OID 58970)
-- Name: datasets datasets_fk6; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.datasets
    ADD CONSTRAINT datasets_fk6 FOREIGN KEY (period_id) REFERENCES public.periods(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 3132 (class 2606 OID 58985)
-- Name: departments departments_fk1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.departments
    ADD CONSTRAINT departments_fk1 FOREIGN KEY (agency_id) REFERENCES public.agencies(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3140 (class 2606 OID 24989)
-- Name: obs obs_fk1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.obs
    ADD CONSTRAINT obs_fk1 FOREIGN KEY (dataset_id) REFERENCES public.datasets(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3141 (class 2606 OID 58990)
-- Name: obs obs_fk2; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.obs
    ADD CONSTRAINT obs_fk2 FOREIGN KEY (code_id) REFERENCES public.codevals(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3142 (class 2606 OID 58995)
-- Name: obs obs_fk3; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.obs
    ADD CONSTRAINT obs_fk3 FOREIGN KEY (unit_id) REFERENCES public.units(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 3143 (class 2606 OID 59000)
-- Name: obs obs_fk4; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.obs
    ADD CONSTRAINT obs_fk4 FOREIGN KEY (period_id) REFERENCES public.periods(id) ON UPDATE CASCADE ON DELETE CASCADE;


-- Completed on 2020-10-13 16:36:03

--
-- PostgreSQL database dump complete
--

