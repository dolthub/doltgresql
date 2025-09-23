-- Downloaded from: https://github.com/iangow/pg_functions/blob/2351aca566c2d08b7cf996cd6467cfd150ce7c58/public.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 9.4beta1
-- Dumped by pg_dump version 9.4beta1
-- Started on 2014-07-18 09:14:39 EDT

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

SET search_path = public, pg_catalog;

--
-- TOC entry 1105 (class 1247 OID 16490)
-- Name: fileinfo; Type: TYPE; Schema: public; Owner: igow
--

CREATE TYPE fileinfo AS (
	filename text,
	filesize bigint,
	ctime abstime,
	mtime abstime,
	atime abstime
);


ALTER TYPE fileinfo OWNER TO igow;

--
-- TOC entry 1108 (class 1247 OID 16493)
-- Name: fog_stats; Type: TYPE; Schema: public; Owner: igow
--

CREATE TYPE fog_stats AS (
	fog double precision,
	num_words integer,
	percent_complex double precision,
	num_sentences integer
);


ALTER TYPE fog_stats OWNER TO igow;

--
-- TOC entry 1111 (class 1247 OID 16496)
-- Name: parsed_name; Type: TYPE; Schema: public; Owner: activism
--

CREATE TYPE parsed_name AS (
	prefix text,
	first_name text,
	middle_initial text,
	last_name text,
	suffix text
);


ALTER TYPE parsed_name OWNER TO activism;

--
-- TOC entry 672 (class 1255 OID 16505)
-- Name: agrep(text, text); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION agrep(x text, y text) RETURNS integer
    LANGUAGE plr
    AS $$
        agrep(x, y)         
    $$;


ALTER FUNCTION public.agrep(x text, y text) OWNER TO igow;

--
-- TOC entry 673 (class 1255 OID 16506)
-- Name: anyleast(anyarray); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION anyleast(VARIADIC anyarray) RETURNS anyelement
    LANGUAGE sql
    AS $_$
    SELECT sum($1[i]) 
    FROM generate_subscripts($1, 1) g(i)
    WHERE $1[i] IS NOT NULL;
$_$;


ALTER FUNCTION public.anyleast(VARIADIC anyarray) OWNER TO igow;

--
-- TOC entry 674 (class 1255 OID 16507)
-- Name: array_diff(anyarray, anyarray); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION array_diff(anyarray, anyarray) RETURNS anyarray
    LANGUAGE sql
    AS $_$ 
    SELECT array(SELECT x from unnest($1) AS x 
    WHERE x NOT IN (SELECT * FROM unnest($2)))
$_$;


ALTER FUNCTION public.array_diff(anyarray, anyarray) OWNER TO igow;

--
-- TOC entry 675 (class 1255 OID 16508)
-- Name: array_max(integer[]); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION array_max(an_array integer[]) RETURNS integer
    LANGUAGE plpython2u
    AS $$
    if an_array is None:
        return None
    return max(an_array)
$$;


ALTER FUNCTION public.array_max(an_array integer[]) OWNER TO igow;

--
-- TOC entry 676 (class 1255 OID 16509)
-- Name: array_min(integer[]); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION array_min(an_array integer[]) RETURNS integer
    LANGUAGE plpython2u
    AS $$
    if an_array is None:
        return None
    return min(an_array)
$$;


ALTER FUNCTION public.array_min(an_array integer[]) OWNER TO igow;

--
-- TOC entry 677 (class 1255 OID 16510)
-- Name: array_rem(anyarray, anyelement); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION array_rem(anyarray, anyelement) RETURNS anyarray
    LANGUAGE sql
    AS $_$ SELECT array(SELECT x from unnest($1) AS x where x <> $2)
  $_$;


ALTER FUNCTION public.array_rem(anyarray, anyelement) OWNER TO igow;

--
-- TOC entry 678 (class 1255 OID 16511)
-- Name: bomonth(timestamp without time zone); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION bomonth(timestamp without time zone) RETURNS date
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$
    SELECT date_trunc('MONTH', $1)::date
  $_$;


ALTER FUNCTION public.bomonth(timestamp without time zone) OWNER TO igow;

--
-- TOC entry 679 (class 1255 OID 16512)
-- Name: calc_evol(anyarray); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION calc_evol(data anyarray) RETURNS double precision
    LANGUAGE plr
    AS $$
    parsed.data <- data
    vol <- NA 
    len <- length(parsed.data)
    
    return(max(data[,2]))       
    if (len >= 9) {
        y <- parsed.data[,1]
        x <- parsed.data[,2]
        reg.data <- data.frame(y,x)
        try ({
            vol <- var(resid(lm(y ~ x, data=reg.data, na.action=na.exclude)))
        }, silent=TRUE)
    }
    return(vol)
$$;


ALTER FUNCTION public.calc_evol(data anyarray) OWNER TO igow;

--
-- TOC entry 680 (class 1255 OID 16513)
-- Name: calc_evol(anyelement, anyelement); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION calc_evol(anyelement, anyelement) RETURNS anyarray
    LANGUAGE plr WINDOW
    AS $$
    vol <- NA
    if ( length(farg1) >= 9 && length(farg2)==length(farg1)) {
        reg.data <- data.frame(y=farg1, x=farg2)
        try ({
            vol <- var(resid(lm(y ~ x, data=reg.data, na.action=na.exclude)))
        }, silent=TRUE)
    }
    return(farg2)
$$;


ALTER FUNCTION public.calc_evol(anyelement, anyelement) OWNER TO igow;

--
-- TOC entry 681 (class 1255 OID 16514)
-- Name: calc_evol(character, integer); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION calc_evol(gvkey character, fyear integer) RETURNS double precision
    LANGUAGE plr
    AS $_$
    vol <- NA
    num.years <- 10L
    min.year <- fyear - num.years + 1

        sql <- paste("SELECT * FROM (SELECT eps, lag(eps) OVER w AS lag_eps ",
    	    "FROM gtv.stuff WHERE gvkey='", gvkey, "' AND ",
    		"fyear >= ", min.year," AND fyear <= ", fyear,
    		" AND eps IS NOT NULL",
    		" WINDOW w AS (PARTITION BY gvkey ORDER BY fyear)) AS a ",
    		"WHERE eps IS NOT NULL AND lag_eps IS NOT NULL", sep="")							
    data <- pg.spi.exec(sql)

    try({
        if(dim(data)[1] >= 9) {
            lm.EPS <- lm(eps ~ lag_eps, data=data, na.action=na.exclude) 
            vol <- var(lm.EPS$residuals)
        }
    }, silent=TRUE)

    return(vol)

$_$;


ALTER FUNCTION public.calc_evol(gvkey character, fyear integer) OWNER TO igow;

--
-- TOC entry 683 (class 1255 OID 16515)
-- Name: clean_name(text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION clean_name(name_orig text) RETURNS text
    LANGUAGE plr IMMUTABLE STRICT
    AS $_$
	temp <- name_orig
	temp <- gsub("(?i)(THE) $","",temp,perl=TRUE)	# Remove the word (THE) near the end
	temp <- gsub("(?i)^THE ","",temp,perl=TRUE)		# Remove the word THE at the start
	temp <- gsub("[-/]"," ",temp,perl=TRUE)			# Replace hyphens and slashes with spaces
	temp <- gsub("[\\.,]","",temp,perl=TRUE)		# Remove periods and commas
	temp <- gsub("\\s+\\/.*$","",temp,perl=TRUE)	# Remove backslashes preceded by spaces
	temp <- gsub("\\s{2,}+"," ",temp,perl=TRUE)		# Replace multiple spaces with singles
	temp <- toupper(temp)							# Convert to upper case
	return(temp)
  $_$;


ALTER FUNCTION public.clean_name(name_orig text) OWNER TO postgres;

--
-- TOC entry 684 (class 1255 OID 16516)
-- Name: clean_name_sql(text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION clean_name_sql(text) RETURNS text
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$
  SELECT regexp_replace(temp, E'[-/]',E' ') FROM
	(SELECT regexp_replace($1, E'((?i)(THE) $|(?i)^THE)|[\\.,]|\\s+\\/.*$)',E'') AS temp) AS a
  -- temp <- gsub("[-/]"," ",temp,perl=TRUE)
  $_$;


ALTER FUNCTION public.clean_name_sql(text) OWNER TO postgres;

--
-- TOC entry 685 (class 1255 OID 16517)
-- Name: clean_proxy_names(text); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION clean_proxy_names(names text) RETURNS text
    LANGUAGE plperl
    AS $_X$
    $temp = $_[0];
    $temp =~ s/ {2,}/ /g;
    $temp =~ s/,?\s+and\s+/\n/gi;
    $temp =~ s/\s*\n\s*/;/g;
    $temp =~ s/\s*\(*\d+[.)]*\s*//g;
    $temp =~ s/,\s*$//;
    return $temp;
$_X$;


ALTER FUNCTION public.clean_proxy_names(names text) OWNER TO igow;

--
-- TOC entry 750 (class 1255 OID 16518)
-- Name: clean_tickers(text); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION clean_tickers(ticker text) RETURNS text
    LANGUAGE plperl
    AS $_X$
  # Remove any asterisks
  $_[0] =~ s/\*//g;

  # Remove trailing .A
  $_[0] =~ s/\.A$//g;

  return $_[0];
$_X$;


ALTER FUNCTION public.clean_tickers(ticker text) OWNER TO igow;

--
-- TOC entry 686 (class 1255 OID 16519)
-- Name: create_name_table(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION create_name_table() RETURNS integer
    LANGUAGE plr COST 1000000
    AS $_$

  # Establish a connection to a PostgreSQL database
  library(RPostgreSQL)
  drv <- dbDriver("PostgreSQL")
  pgr <- dbConnect(drv, dbname = "crsp")

  library(RPostgreSQL)
  
  # A function to scrub company names of unnecessary details
  cleanNames <- function(names) {
    temp <- gsub("(?i)(THE) $","",names,perl=TRUE)  # Remove the word (THE) near the end
    temp <- gsub("(?i)^THE ","",temp,perl=TRUE)     # Remove the word THE at the start
    temp <- gsub("[-/]"," ",temp,perl=TRUE)         # Replace hyphens and slashes with spaces
    temp <- gsub("[\\.,]","",temp,perl=TRUE)        # Remove periods and commas
    temp <- gsub("\\s+\\/.*$","",temp,perl=TRUE)    # Remove backslashes preceded by spaces
    temp <- gsub("\\s{2,}+"," ",temp,perl=TRUE)     # Replace multiple spaces with singles
    temp <- toupper(temp)                           # Convert to upper case
    temp
  }

  # Construct a table of company name-CIK pairs
  # This is a surprisingly big set (422,123 on 2011-05-07)
  name.ciks <-
    dbGetQuery(pgr,"SELECT company_name, cik FROM filings.filings LIMIT 100")

  # Clean up names to merge tables... 
  name.ciks$name_mod <- cleanNames(name.ciks$company_name)

  rs <- dbWriteTable(pgr,"cleaned_names",
    name.ciks, overwrite=TRUE, row.names=FALSE)

  return(0)

  $_$;


ALTER FUNCTION public.create_name_table() OWNER TO postgres;

--
-- TOC entry 687 (class 1255 OID 16520)
-- Name: cusip_to_permno(text); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION cusip_to_permno(text) RETURNS integer
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$
    SELECT DISTINCT PERMNO FROM crsp.stocknames WHERE cusip=$1;
  $_$;


ALTER FUNCTION public.cusip_to_permno(text) OWNER TO igow;

--
-- TOC entry 688 (class 1255 OID 16521)
-- Name: date_quarter(date); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION date_quarter(date) RETURNS date
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$
        SELECT ((extract(year FROM $1)-1 ||'-12-31')::date +  
                    interval '1 month' *  (extract(quarter FROM $1)*3))::date
    $_$;


ALTER FUNCTION public.date_quarter(date) OWNER TO igow;

--
-- TOC entry 689 (class 1255 OID 16522)
-- Name: direct_id(text); Type: FUNCTION; Schema: public; Owner: igow
--


CREATE FUNCTION director_id(text) RETURNS integer
    LANGUAGE sql
    AS $_$
    SELECT CASE WHEN $1 != ''
    THEN regexp_replace($1, E'^.*\\.(\\d+)$', E'\\1')::integer
    ELSE NULL END
  $_$;

CREATE FUNCTION eomonth(date) RETURNS date
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$
    SELECT (date_trunc('MONTH', $1) + INTERVAL '1 month - 1 day')::date;
  $_$;

CREATE FUNCTION eomonth(timestamp without time zone) RETURNS date
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$
    SELECT (date_trunc('MONTH', $1) + INTERVAL '1 month - 1 day')::date;
  $_$;

CREATE FUNCTION equilar_id(text) RETURNS integer
    LANGUAGE sql
    AS $_$
        SELECT regexp_replace($1, E'^(\\d+)\\..*$', E'\\1')::integer
    $_$;


ALTER FUNCTION public.equilar_id(text) OWNER TO igow;

CREATE FUNCTION evol(double precision, double precision) RETURNS double precision
    LANGUAGE plr WINDOW
    AS $_$
    vol <- NA
    y <- farg1
    x <- farg2 
    if (fnumrows==9) try (vol <- lm(y ~ x)$coefficients[2])
    return(vol)
$_$;

CREATE FUNCTION extract_products(raw_list text[], participants text[]) RETURNS text[]
    LANGUAGE plpython2u
    AS $$
    "Function to extract product names from text"
    import re
    
    raw_text = ' '.join(raw_list)

    # gazetteers is a corpus containing place names
    from nltk.corpus import gazetteers
    
    # Find capitalized words that do not follow sentence-ending punctuation
    regex = r'(?<![\.\?\!-])\s\b([A-Z][A-Za-z0-9]+)'	
    matches = list(set(re.findall(regex, raw_text)))

    # A list of words that should excluded
    exclude = ' '.join(participants).split() + gazetteers.words() + \
       ['EBITDA', 'EBIT', 'IT', 'Capex', 'CapEx', 'Board', 'IFRS', 'GAAP',
        'EPS', 'Group', 'OpEx'] + ['January', 'February', 'March',
        'April', 'June',
        'July', 'August', 'September', 'October', 'November',
        'December'] + ['North', 'East', 'West', 'South'] + \
       ['Europe', 'European', 'Asia', 'Asian', 'American', 'Africa', 'African'] + \
       ['Q1', 'Q2', 'Q3', 'Q4'] 

    return [match for match in matches if match not in exclude]
$$;

CREATE FUNCTION extract_suffixes(text) RETURNS text[]
    LANGUAGE plperl
    AS $_X$
    $temp = $_[0];
    $suffix_list = '\(ret\.\)|Esq\.|C[PF]A|M\.?P\.?H\.?|M\.?D\.?|[SJ][Rr].?|IV|III?\.?|P[Hh]\.?D\.?';
    $suffix_list .= '|M\.D\., Ph\.D\.';
    $prefix_list = '(?:Prof\.|Mr\.|Mrs\.|Dr\.|Captain)';
    $temp =~ /^-?($prefix_list)?(?:,|\s)?(.*?)(?:\s+|,)?($suffix_list)?,?$/;
    $prefix = $1;
    $name = $2;
    $suffix = $3;
    return [$prefix , $name , $suffix];
$$;

CREATE FUNCTION ff_alpha(double precision, double precision, double precision, double precision) RETURNS double precision
LANGUAGE plr WINDOW AS
$_$
    alpha <- NA
    try (alpha <- lm(farg1 ~ farg2 + farg3 + farg4)$coefficients[1])
    return(alpha)
$_$;

CREATE FUNCTION fiscal_year(date) RETURNS integer
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$
   SELECT (extract(year from $1))::integer - 
    (extract(month from $1)::integer < 6)::integer;
  $_$;

CREATE FUNCTION fix_cusip(cusip_orig text) RETURNS text
    LANGUAGE plr IMMUTABLE STRICT
    AS $$
  # Remove spaces from CUSIPs
	cusip <- gsub("\\s+","",cusip_orig)

	# First fix CUSIPs with the letter "E" in them. These get interpreted by Excel
	# as numbers, with "00101E20" being understood as 1.01E22 (or 1.01 * 10^22). 
	# We need to convert "1.01E22" back to "00101E20".
	#                     1234567
	#  This is achieved by removing the "." adjusting the 20 to 22 (adding
	#  back the number of characters between "." and "E").
	if (grepl("[[Ee]\\+",cusip)) {

		# Get some parameters to convert the defective CUSIP
		point <- regexpr("\\.",cusip,perl=TRUE)[[1]] 	# Location of ".", 2 in the example
		e <- regexpr("[Ee]",cusip,perl=TRUE)[[1]]			# Location of "E", 5 in the example
		exponent <- substr(cusip,e+2,nchar(cusip)) 		# Uncorrected exponent
		exponent <- as.double(exponent) - (e-point-1) # Corrected exponent

		# Combine everything to create a valid CUSIP
		cusip.new <- paste(substr(cusip,1,point-1), substr(cusip,point+1,e-1),
												"E",exponent,sep="")
	} else {
		cusip.new <- cusip # Seems the CUSIP is OK
	}

	# If leading zeros are needed
	if (8-nchar(cusip.new)>0) { 
		# create a string of leading zeros...
		leading.zeros <- paste(rep("0",8-nchar(cusip.new)),collapse="") 

		# ... and add them to the CUSIP
		cusip.new <- paste(leading.zeros,cusip.new,sep="")
	} 

	# Return the fixed CUSIP
	cusip.new
  $$;

CREATE FUNCTION fix_cusip9(cusip_orig text) RETURNS text
    LANGUAGE plr IMMUTABLE STRICT
    AS $$
  # Remove spaces from CUSIPs
	cusip <- gsub("\\s+","",cusip_orig)

	# First fix CUSIPs with the letter "E" in them. These get interpreted by Excel
	# as numbers, with "00101E20" being understood as 1.01E22 (or 1.01 * 10^22). 
	# We need to convert "1.01E22" back to "00101E20".
	#                     1234567
	#  This is achieved by removing the "." adjusting the 20 to 22 (adding
	#  back the number of characters between "." and "E").
	if (grepl("[[Ee]\\+",cusip)) {

		# Get some parameters to convert the defective CUSIP
		point <- regexpr("\\.",cusip,perl=TRUE)[[1]] 	# Location of ".", 2 in the example
		e <- regexpr("[Ee]",cusip,perl=TRUE)[[1]]			# Location of "E", 5 in the example
		exponent <- substr(cusip,e+2,nchar(cusip)) 		# Uncorrected exponent
		exponent <- as.double(exponent) - (e-point-1) # Corrected exponent

		# Combine everything to create a valid CUSIP
		cusip.new <- paste(substr(cusip,1,point-1), substr(cusip,point+1,e-1),
												"E",exponent,sep="")
	} else {
		cusip.new <- cusip # Seems the CUSIP is OK
	}

	# If leading zeros are needed
	if (9-nchar(cusip.new)>0) { 
		# create a string of leading zeros...
		leading.zeros <- paste(rep("0",9-nchar(cusip.new)),collapse="") 

		# ... and add them to the CUSIP
		cusip.new <- paste(leading.zeros,cusip.new,sep="")
	} 

	# Return the fixed CUSIP
	cusip.new
  $$;



CREATE OR REPLACE FUNCTION fog(text) RETURNS double precision
    LANGUAGE plperl AS 
$$ 

  # Load Perl modules that calculate fog, etc.
  use Lingua::EN::Fathom;
  use Lingua::EN::Sentence qw( get_sentences add_acronyms );

  my $text = new Lingua::EN::Fathom;
  if (defined($_[0])) {
    $text->analyse_block($_[0]);
    return($text->fog);
  }

$$;


ALTER FUNCTION public.fog(text) OWNER TO igow;

--
-- TOC entry 703 (class 1255 OID 16539)
-- Name: fog_data(text); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION fog_data(text) RETURNS fog_stats
    LANGUAGE plperlu AS
$$ 

  # Load Perl module that calculates fog, etc.
  use Lingua::EN::Fathom;
  use Lingua::EN::Sentence qw( get_sentences add_acronyms );

  my $text = new Lingua::EN::Fathom;
  if (defined($_[0])) {
    $text->analyse_block($_[0]);
    $fog   = $text->fog;
    $num_words = $text->num_words;
    $percent_complex = $text->percent_complex_words;  
    $num_sentences = $text->num_sentences;
  }

  return {fog => $fog, num_words => $num_words, percent_complex => $percent_complex, 
	num_sentences => $num_sentences};

$;

CREATE FUNCTION intersection(anyarray, anyarray) RETURNS anyarray
    LANGUAGE sql
    AS $_$
SELECT ARRAY(
    SELECT $1[i]
    FROM generate_series( array_lower($1, 1), array_upper($1, 1) ) i
    WHERE ARRAY[$1[i]] && $2
);
$_$;

CREATE FUNCTION kls_domains(raw_text text[]) RETURNS text[]
    LANGUAGE plpython2u
    AS $$
	import unicodedata, re
	from nltk.tokenize import wordpunct_tokenize
	
	word_list = {
		'market': ['market', 'marketplace', 'environment', 'segment', 'sector'],
		'competition': ['market', 'marketplace', 'environment', 'customer', 'channel', 'value', 
			'first-mover', 'technology', 'alliance', 'partnership', 'venture', 
			'regulation', 'litigation'],
		'industry_structure': ['industry', 'entrant', 'supplier', 'buyer', 'substitute',
			'scale', 'product', 'brand', 'switching', 'capital', 'access', 'cost', 
			'rivalry', 'capacity', 'concentration', 'exit', 'barrier', 'price', 'profit', 
			'quality', 'input', 'volume', 'purchase', 'integration', 
			'power'],
		'strategic_intent': ['strategy', 'strategic', 'value', 'sales', 'revenue',
			'share', 'profit', 'profitability', 'product', 'service', 'lead', 
			'leader', 'quality', 'customer', 'buyer', 'growth', 'opportunity',
			 'risk', 'resource'],
		'innovation_and_r_d': ['research and development', 'r&d', 'patent', 'discovery',
			'license', 'licensing', 'regulation', 'regulatory', 'trial', 'monitor', 
			'innovate', 'innovation', 'competence'],
		'mode_of_entry': ['entry', 'cost', 'business', 'complementary', 'green-field', 
			'venture', 'investment', 'capital', 'solution', 'price'],
		'business_model': ['model', 'best', 'lowest', 'low', 'highest', 'high',
			 'supplier', 'distribution'],
		'partnerships': ['partner', 'alliance', 'merger', 'acquisition', 'joint', 
			'venture', 'relationship', 'equity', 'asset'],
		'leadership': ['leader', 'leadership', 'record', 'value', 'culture', 
			'responsibility', 'goal', 'objective'],
		'management_quality': ['management', 'quality', 'best', 'proven', 
			'experience', 'teamwork'],
		'governance': ['recruitment', 'development', 'governance', 'corporate',
			 'board', 'incentive', 'owner', 'ownership', 'compensation'],
		'disclosure': ['disclosure', 'transparent', 'transparency', 'information',
			'audit', 'auditing', 'oversight', 'assurance', 'regulation', 'mandate',
			'mandated'],
		'measures': ['up', 'down', 'better', 'worse', 'recover', 'advance', 
			'advancing', 'progress', 'progressing', 'expand', 'expanding', 'improve',
			 'improving', 'reduce', 'reducing', 'reduction', 'decline', 'declining', 
			'retain', 'retention', 'profit', 'profitability', 'feedback', 'scorecard',
				 'growth', 'growing', 'performance', 'projected', 'projections'],
		'customer': ['customer', 'satisfaction', 'feedback', 'trust'],
		'brand': ['brand', 'image', 'name', 'trademark', 'recognition', 'stretch', 
			'quality', 'awareness'],
		'media': ['radio', 'television', 'newspaper', 'internet', 'promotion', 
			'media spend', 'announcements', 'release', 'media budget'],
		'advertising': ['advertising', 'ad', 'direct', 'channel', 'advertising', 'spend',
			'ad spend', 'advertising allocation', 'budget', 'ad budget'],
		'corporate_image': ['corporate image', 'reputation', 'integrity', 'community', 
			'trust', 'trusted name', 'confidence', 'durability', 'strength', 'character'],
		'financial_performance': ['gross', 'net', 'return on investment', 'return on sales', 
			'return on assets', 'return on equity', 'ROI', 'ROA', 'ROE', 'profit', 
			'earnings', 'margin', 'capital', 'debt', 'sales', 'EBITDA', 'ratings',
			 'leverage', 'valuation', 'cost of capital'],
		'forecasting': ['forecast', 'forecasting', 'cash flow', 'prospectus', 'quarterly'],
		'insider_stock_transactions': ['insider buy', 'insider sell'],
		'regulation': ['regulation', 'federal', 'state', 'securities and exchange', 'commission',
			 'commerce', 'legislation', 'congress', 'law', 'legal', 'hearings',
			 'enacted', 'pending', 'sec', 'medicare', 'medicaid', 'FDA'],
		'special_interest_groups': ['lobby', 'lobbyists', 'special', 'interest', 'expert', 
			'testimony', 'industry', 'watchdog', 'consumer rights', 'patient rights'] 
			}

	lower_text = ' '.join(raw_text).decode('utf8').lower()
	
	domains = [(domain, term) for domain in ['industry_structure', 'competition', 'financial_performance',
                       'innovation_and_r_d', 'mode_of_entry', 'brand'] # word_list.keys() 
		for term in word_list[domain] if re.search(term, lower_text)]
	return list(set(domains))
    $$;

CREATE FUNCTION liwc_counts(the_text text) RETURNS json
    LANGUAGE plpythonu
    AS $_$

    """Function to return number of matches against a LIWC category in a text"""
    if 're' in SD:
        re = SD['re']
        json = SD['json']
    else:
        import re, json
        SD['re'] = re
        SD['json'] = json

    if SD.has_key("regex_list"):
        regex_list = SD["regex_list"]
        categories = SD["categories"]
    else:
        rv = plpy.execute("SELECT category FROM personality.word_list")

        categories = [ (r["category"]) for r in rv]

        # Implement Robin's suggestion to convert *s to regular expressions
        # outside the loop. And a
        plan = plpy.prepare("""
            SELECT word_list
            FROM personality.word_list
            WHERE category = $1""", ["text"])
        mod_word_list = {}
        for cat in categories:
            rows = list(plpy.cursor(plan, [cat]))
            word_list = rows[0]['word_list']
            mod_word_list[cat] = [re.sub('\*(?:\s*$)?', '[a-z]*', word.lower())
                                    for word in word_list]

        # Pre-compile regular expressions.
        regex_list = {}
        for key in mod_word_list.keys():
            regex = '\\b(?:' + '|'.join(mod_word_list[key]) + ')\\b'
            regex_list[key] = re.compile(regex)
        SD["regex_list"] = regex_list
        SD["categories"] = categories

    # rest of function

    # Construct a counter of the words and return as JSON
    text = re.sub(u'\u2019', "'", the_text).lower()
    the_dict = {cat: len(re.findall(regex_list[cat], text)) for cat in categories}
    return json.dumps(the_dict)

$_$;

CREATE FUNCTION ordered_set(the_list integer[]) RETURNS integer[]
    LANGUAGE plpython2u
    AS $$
    ordered_set = []
    for x in the_list:
        if x not in ordered_set:
            ordered_set.append(x)
    return ordered_set
$$;

CREATE FUNCTION ordered_set(the_list text[]) RETURNS text[]
    LANGUAGE plpython2u
    AS $$
    ordered_set = []
    for x in the_list:
        if x not in ordered_set:
            ordered_set.append(x)
    return ordered_set
$$;

CREATE FUNCTION plr_array_accum(double precision[], double precision[]) RETURNS double precision[]
    LANGUAGE c
    AS '$libdir/plr', 'plr_array_accum';


ALTER FUNCTION public.plr_array_accum(double precision[], double precision[]) OWNER TO igow;

CREATE FUNCTION plr_array_accum(text[], text) RETURNS text[]
    LANGUAGE c
    AS '$libdir/plr', 'plr_array_accum';


ALTER FUNCTION public.plr_array_accum(text[], text) OWNER TO igow;

CREATE FUNCTION plr_array_accum_float(double precision[], double precision[]) RETURNS double precision[]
    LANGUAGE c
    AS '$libdir/plr', 'plr_array_accum';


ALTER FUNCTION public.plr_array_accum_float(double precision[], double precision[]) OWNER TO igow;

--
-- TOC entry 713 (class 1255 OID 16554)
-- Name: plr_array_accum_float(text[], double precision[]); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION plr_array_accum_float(text[], double precision[]) RETURNS text[]
    LANGUAGE c
    AS '$libdir/plr', 'plr_array_accum';


ALTER FUNCTION public.plr_array_accum_float(text[], double precision[]) OWNER TO igow;

--
-- TOC entry 714 (class 1255 OID 16555)
-- Name: plr_array_accum_float(text[], text); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION plr_array_accum_float(text[], text) RETURNS text[]
    LANGUAGE c
    AS '$libdir/plr', 'plr_array_accum';


ALTER FUNCTION public.plr_array_accum_float(text[], text) OWNER TO igow;

--
-- TOC entry 715 (class 1255 OID 16556)
-- Name: pos_tag(text); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION pos_tag(string text) RETURNS text[]
    LANGUAGE plpython2u AS
$$
    import nltk
    return  [nltk.pos_tag(nltk.word_tokenize(sentence))
        for sentence in nltk.sent_tokenize(string)]
$$;

CREATE FUNCTION prep_gtv(text) RETURNS text
    LANGUAGE plr
    AS $$
  gtv <<- pg.spi.prepare(arg1, c(20, 18));
  print("OK")
$$;


ALTER FUNCTION public.prep_gtv(text) OWNER TO igow;

--
-- TOC entry 717 (class 1255 OID 16558)
-- Name: quarters_between(date, date); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION quarters_between(date, date) RETURNS integer
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$
     SELECT quarters_of($1) - quarters_of($2)
  $_$;


ALTER FUNCTION public.quarters_between(date, date) OWNER TO igow;

--
-- TOC entry 718 (class 1255 OID 16559)
-- Name: quarters_of(date); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION quarters_of(date) RETURNS integer
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$
    SELECT extract(years FROM $1)::int * 4 + extract(quarter FROM $1)::int
  $_$;


ALTER FUNCTION public.quarters_of(date) OWNER TO igow;

--
-- TOC entry 719 (class 1255 OID 16560)
-- Name: r_median(double precision[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION r_median(double precision[]) RETURNS double precision
    LANGUAGE plr
    AS $$
  median(arg1)
$$;


ALTER FUNCTION public.r_median(double precision[]) OWNER TO postgres;

--
-- TOC entry 720 (class 1255 OID 16561)
-- Name: r_median_window(double precision); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION r_median_window(double precision) RETURNS double precision
    LANGUAGE plr WINDOW
    AS $$
  median(farg1)
$$;


ALTER FUNCTION public.r_median_window(double precision) OWNER TO igow;

--
-- TOC entry 721 (class 1255 OID 16562)
-- Name: r_quantile(double precision[], double precision); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION r_quantile(double precision[], double precision) RETURNS double precision
    LANGUAGE plr STRICT
    AS $$quantile(arg1, arg2, na.rm=TRUE)$$;


ALTER FUNCTION public.r_quantile(double precision[], double precision) OWNER TO igow;

--
-- TOC entry 722 (class 1255 OID 16563)
-- Name: r_quintile_rank(double precision[]); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION r_quintile_rank(data double precision[]) RETURNS double precision[]
    LANGUAGE plr
    AS $$
    as.integer(cut(data,as.vector(quantile(data,probs=seq(0,1,length.out=6),na.rm=TRUE)), include.lowest=TRUE))
$$;


ALTER FUNCTION public.r_quintile_rank(data double precision[]) OWNER TO igow;

--
-- TOC entry 723 (class 1255 OID 16564)
-- Name: r_quintile_rank(numeric); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION r_quintile_rank(data numeric) RETURNS double precision[]
    LANGUAGE plr
    AS $$
    as.integer(cut(data,quantile(data,probs=seq(0,1,length.out=5),na.rm=TRUE)))
$$;


ALTER FUNCTION public.r_quintile_rank(data numeric) OWNER TO igow;

--
-- TOC entry 724 (class 1255 OID 16565)
-- Name: r_regr_slope(double precision, double precision); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION r_regr_slope(double precision, double precision) RETURNS double precision
    LANGUAGE plr WINDOW
    AS $_$
  slope <- NA
  y <- farg1
  x <- farg2
  try (slope <- lm(y ~ x)$coefficients[2])
  return(slope)
$_$;


ALTER FUNCTION public.r_regr_slope(double precision, double precision) OWNER TO igow;

--
-- TOC entry 725 (class 1255 OID 16566)
-- Name: r_sum(double precision[]); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION r_sum(x double precision[]) RETURNS double precision
    LANGUAGE plr
    AS $$
    sum(x, na.rm=TRUE) 
$$;


ALTER FUNCTION public.r_sum(x double precision[]) OWNER TO igow;

--
-- TOC entry 726 (class 1255 OID 16567)
-- Name: rel_freq(double precision); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION rel_freq(double precision) RETURNS double precision
    LANGUAGE plr WINDOW
    AS $$
  var <- as.vector(farg1)
  return((var/sum(var ))[prownum])
$$;


ALTER FUNCTION public.rel_freq(double precision) OWNER TO igow;

--
-- TOC entry 733 (class 1255 OID 16568)
-- Name: remove_trailing_q(text); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION remove_trailing_q(ticker text) RETURNS text
    LANGUAGE plperl
    AS $_X$
  # Remove trailing Qs
  $_[0] =~ s/Q$//g;

  return $_[0];
$_X$;


ALTER FUNCTION public.remove_trailing_q(ticker text) OWNER TO igow;

--
-- TOC entry 727 (class 1255 OID 16569)
-- Name: row_sum(anyarray); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION row_sum(VARIADIC anyarray) RETURNS double precision
    LANGUAGE sql
    AS $_$
    SELECT r_sum(ARRAY[$1]) 
$_$;


ALTER FUNCTION public.row_sum(VARIADIC anyarray) OWNER TO igow;

--
-- TOC entry 728 (class 1255 OID 16570)
-- Name: row_sum_alt(double precision[]); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION row_sum_alt(VARIADIC x double precision[]) RETURNS double precision
    LANGUAGE plr
    AS $$
    sum(x, na.rm=TRUE) 
$$;


ALTER FUNCTION public.row_sum_alt(VARIADIC x double precision[]) OWNER TO igow;

--
-- TOC entry 732 (class 1255 OID 16571)
-- Name: sent_count(text); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION sent_count(raw text) RETURNS integer
    LANGUAGE plpythonu
    AS $$
    """Function to count the number of sentences in a passage."""
    import nltk, sys
    if (sys.version_info.major==2):
        text= raw.decode('utf-8')
    sent_tokenizer = nltk.data.load('tokenizers/punkt/english.pickle')
    sents = sent_tokenizer.tokenize(text)
    return len(sents)
$$;


ALTER FUNCTION public.sent_count(raw text) OWNER TO igow;

--
-- TOC entry 729 (class 1255 OID 16572)
-- Name: table_file_info(text, text); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION table_file_info(schemaname text, tablename text) RETURNS SETOF fileinfo
    LANGUAGE plpythonu
    AS $_$
    import datetime, glob, os
    db_info = plpy.execute("""
       SELECT datname AS database_name,
         current_setting('data_directory') || '/base/' || db.oid AS data_directory
       FROM pg_database db
       WHERE datname= current_database()""")

    table_info_plan = plpy.prepare("""
      SELECT nspname AS schemaname,
        relname AS tablename,
        relfilenode AS filename
      FROM pg_class c
      JOIN pg_namespace ns
      ON c.relnamespace=ns.oid
      WHERE nspname=$1 AND relname =$2;""", ['text', 'text'])

    table_info = plpy.execute(table_info_plan, [schemaname, tablename])
    filemask = '%s/%s*' % (db_info[0]['data_directory'], table_info[0]['filename'])
    res = []
    for filename in glob.glob(filemask):
        fstat = os.stat(filename)
        res.append((
          filename,
          fstat.st_size,
          datetime.datetime.fromtimestamp(fstat.st_ctime).isoformat(),
          datetime.datetime.fromtimestamp(fstat.st_mtime).isoformat(),
          datetime.datetime.fromtimestamp(fstat.st_atime).isoformat()
        ))
return res
$_$;


ALTER FUNCTION public.table_file_info(schemaname text, tablename text) OWNER TO igow;

--
-- TOC entry 730 (class 1255 OID 16573)
-- Name: test_spi_execp(text, text, text); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION test_spi_execp(text, text, text) RETURNS SETOF record
    LANGUAGE plr
    AS $$
  pg.spi.execp(pg.reval(arg1), list(arg2,arg3))
$$;


ALTER FUNCTION public.test_spi_execp(text, text, text) OWNER TO igow;

--
-- TOC entry 731 (class 1255 OID 16574)
-- Name: test_spi_prep(text); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION test_spi_prep(text) RETURNS text
    LANGUAGE plr
    AS $$
  sp <<- pg.spi.prepare(arg1, c(NAMEOID, NAMEOID));
  print("OK")
$$;


ALTER FUNCTION public.test_spi_prep(text) OWNER TO igow;

--
-- TOC entry 734 (class 1255 OID 16575)
-- Name: tone_count(text); Type: FUNCTION; Schema: public; Owner: igow
--

CREATE FUNCTION tone_count(the_text text) RETURNS json
    LANGUAGE plpythonu
    AS $_$
    if 're' in SD:
        re = SD['re']
        json = SD['json']
    else:
        import re, json
        SD['re'] = re
        SD['json'] = json

    if SD.has_key("regex_list"):
        regex_list = SD["regex_list"]
        categories = SD["categories"]
    else:

        rv = plpy.execute("SELECT category FROM bgt.lm_tone")
       
        categories = [ (r["category"]) for r in rv]

        # Implement Robin's suggestion to convert *s to regular expressions
        # outside the loop. And a
        plan = plpy.prepare("""
            SELECT word_list
            FROM bgt.lm_tone 
            WHERE category = $1""", ["text"])
        mod_word_list = {}
        for cat in categories:
            rows = list(plpy.cursor(plan, [cat]))
            word_list = rows[0]['word_list']
            mod_word_list[cat] = [word.lower() for word in word_list]

        # Pre-compile regular expressions.
        regex_list = {}
        for key in mod_word_list.keys():
            regex = '\\b(?:' + '|'.join(mod_word_list[key]) + ')\\b'
            regex_list[key] = re.compile(regex)
        SD["regex_list"] = regex_list
        SD["categories"] = categories

    # rest of function
    """Function to return number of matches against a LIWC category in a text"""
    text = re.sub(u'\u2019', "'", the_text).lower()
    the_dict = {category: len(re.findall(regex_list[category], text)) for category in categories}
    return json.dumps(the_dict)
    
$_$;

CREATE FUNCTION var_resid(double precision, double precision) RETURNS double precision
    LANGUAGE plr WINDOW
    AS $$
        var_resid <- NA
        try (var_resid <- var(resid(lm(farg1 ~ farg2))))
        return(var_resid)
    $$;

CREATE FUNCTION winsorize(double precision, double precision) RETURNS double precision
    LANGUAGE plr WINDOW
    AS $$
	library(psych)
	return(winsor(as.vector(farg1), arg2)[prownum])
	# return(farg1)
$$;


CREATE FUNCTION word_count(raw text) RETURNS integer
    LANGUAGE plpythonu
    AS $$
    """Function to count the number of words in a passage."""
    import nltk, sys
    if (sys.version_info.major==2):
        text = raw.decode('utf-8')
    tokens = nltk.word_tokenize(text)
    return len(tokens)
$$;

CREATE FUNCTION word_tokenize_r(raw_text text) RETURNS text[]
    LANGUAGE plr
    AS $$
      require('RWeka')
      words <- WordTokenizer(raw_text)
      return(words)
    $$;


CREATE AGGREGATE array_accum(anyelement) (
    SFUNC = array_append,
    STYPE = anyarray,
    INITCOND = '{}'
);


CREATE AGGREGATE evol_new(anyarray) (
    SFUNC = array_cat,
    STYPE = anyarray,
    INITCOND = '{{NULL,NULL}}',
    FINALFUNC = public.calc_evol
);


CREATE AGGREGATE median(double precision) (
    SFUNC = public.plr_array_accum,
    STYPE = double precision[],
    FINALFUNC = r_median
);


ALTER AGGREGATE public.median(double precision) OWNER TO postgres;

--
-- TOC entry 2287 (class 1255 OID 16588)
-- Name: product(double precision); Type: AGGREGATE; Schema: public; Owner: igow
--

CREATE AGGREGATE product(double precision) (
    SFUNC = float8mul,
    STYPE = double precision
);


ALTER AGGREGATE public.product(double precision) OWNER TO igow;

--
-- TOC entry 2288 (class 1255 OID 16589)
-- Name: quintile_rank(double precision); Type: AGGREGATE; Schema: public; Owner: igow
--

CREATE AGGREGATE quintile_rank(double precision) (
    SFUNC = public.plr_array_accum,
    STYPE = double precision[],
    FINALFUNC = public.r_quintile_rank
);

GRANT USAGE ON SCHEMA public TO crsp_basic;


-- Completed on 2014-07-18 09:14:40 EDT

--
-- PostgreSQL database dump complete
--

