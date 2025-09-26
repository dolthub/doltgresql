-- Downloaded from: https://github.com/jwalit21/BitmapJoinDatabaseEngine/blob/d8bb2343fd89a88a6eb88503b6ccd308a2bc9cd3/R_S.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 14.9 (Debian 14.9-1.pgdg120+1)
-- Dumped by pg_dump version 14.9 (Debian 14.9-1.pgdg120+1)

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
-- Name: plpython3u; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS plpython3u WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpython3u; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpython3u IS 'PL/Python3U untrusted procedural language';


--
-- Name: random_int(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.random_int() RETURNS integer
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN floor(random() * 5 + 1)::INTEGER;
END;
$$;


ALTER FUNCTION public.random_int() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: r; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.r (
    a1 integer,
    a2 integer,
    a3 integer,
    a4 integer,
    ann character varying(10),
    i integer NOT NULL
);


ALTER TABLE public.r OWNER TO postgres;

--
-- Name: r_i_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.r_i_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.r_i_seq OWNER TO postgres;

--
-- Name: r_i_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.r_i_seq OWNED BY public.r.i;


--
-- Name: s; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.s (
    b1 integer,
    b2 integer,
    b3 integer,
    b4 integer,
    ann character varying(10),
    i integer NOT NULL
);


ALTER TABLE public.s OWNER TO postgres;

--
-- Name: s_i_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.s_i_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.s_i_seq OWNER TO postgres;

--
-- Name: s_i_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.s_i_seq OWNED BY public.s.i;


--
-- Name: r i; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.r ALTER COLUMN i SET DEFAULT nextval('public.r_i_seq'::regclass);


--
-- Name: s i; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.s ALTER COLUMN i SET DEFAULT nextval('public.s_i_seq'::regclass);


--
-- Data for Name: r; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.r (a1, a2, a3, a4, ann, i) FROM stdin;
1	5	4	4	R1	1
3	2	2	5	R2	2
1	5	3	2	R3	3
1	3	2	4	R4	4
4	4	4	2	R5	5
2	3	4	4	R6	6
2	3	1	3	R7	7
4	1	2	4	R8	8
3	5	4	2	R9	9
2	4	2	5	R10	10
4	5	1	1	R11	11
2	5	3	3	R12	12
3	5	3	2	R13	13
3	2	3	3	R14	14
1	3	2	2	R15	15
2	2	4	4	R16	16
1	2	5	4	R17	17
4	1	5	4	R18	18
4	1	4	5	R19	19
5	2	3	1	R20	20
3	5	4	1	R21	21
2	2	3	4	R22	22
2	1	3	1	R23	23
1	4	1	2	R24	24
2	2	5	3	R25	25
3	5	3	1	R26	26
4	2	3	1	R27	27
2	1	2	5	R28	28
5	5	4	5	R29	29
1	4	4	4	R30	30
3	1	3	3	R31	31
2	5	4	5	R32	32
2	4	5	5	R33	33
2	3	4	3	R34	34
2	2	3	3	R35	35
2	3	4	4	R36	36
4	2	5	2	R37	37
4	5	5	5	R38	38
1	2	4	1	R39	39
2	4	3	5	R40	40
5	5	5	2	R41	41
5	5	1	1	R42	42
3	4	3	1	R43	43
5	2	2	4	R44	44
4	1	5	1	R45	45
5	2	1	4	R46	46
5	1	3	3	R47	47
3	3	2	1	R48	48
1	2	1	2	R49	49
5	1	3	3	R50	50
1	3	3	2	R51	51
3	4	4	4	R52	52
1	3	5	3	R53	53
2	2	5	2	R54	54
1	4	3	2	R55	55
2	2	1	4	R56	56
4	1	3	4	R57	57
2	1	2	1	R58	58
3	3	3	5	R59	59
3	3	3	4	R60	60
2	1	1	5	R61	61
2	4	1	4	R62	62
3	2	5	2	R63	63
1	5	1	3	R64	64
4	5	3	5	R65	65
5	1	1	4	R66	66
2	1	1	4	R67	67
3	2	1	5	R68	68
5	2	5	3	R69	69
1	3	4	4	R70	70
3	3	2	4	R71	71
3	3	1	4	R72	72
2	3	2	5	R73	73
5	2	3	4	R74	74
2	2	1	1	R75	75
1	4	1	1	R76	76
4	4	1	1	R77	77
2	2	4	1	R78	78
1	1	4	3	R79	79
2	3	1	4	R80	80
3	2	4	1	R81	81
3	3	2	1	R82	82
2	4	3	5	R83	83
3	3	1	2	R84	84
3	3	2	1	R85	85
1	2	2	2	R86	86
5	5	4	2	R87	87
5	4	5	4	R88	88
4	2	4	4	R89	89
4	5	5	4	R90	90
2	1	2	3	R91	91
4	5	1	5	R92	92
5	5	5	3	R93	93
4	4	1	4	R94	94
2	3	5	1	R95	95
5	1	2	2	R96	96
2	4	2	5	R97	97
3	2	1	3	R98	98
4	2	2	2	R99	99
3	4	4	3	R100	100
4	4	3	5	R101	101
3	4	2	4	R102	102
3	5	3	5	R103	103
1	1	3	2	R104	104
3	1	5	1	R105	105
3	1	1	5	R106	106
3	3	4	4	R107	107
1	4	3	2	R108	108
2	2	1	4	R109	109
5	3	1	5	R110	110
2	3	3	5	R111	111
2	1	4	4	R112	112
2	1	4	4	R113	113
1	5	4	4	R114	114
5	1	4	3	R115	115
2	1	5	2	R116	116
5	2	5	5	R117	117
5	2	4	2	R118	118
3	1	5	1	R119	119
1	4	2	3	R120	120
2	2	5	1	R121	121
3	3	4	5	R122	122
2	2	5	4	R123	123
5	4	1	4	R124	124
2	2	4	2	R125	125
2	4	1	3	R126	126
5	5	2	3	R127	127
4	4	2	4	R128	128
2	3	5	5	R129	129
5	4	3	3	R130	130
4	2	3	4	R131	131
2	3	1	3	R132	132
4	4	3	5	R133	133
2	1	5	4	R134	134
5	1	3	2	R135	135
4	1	1	3	R136	136
4	1	1	2	R137	137
4	5	3	4	R138	138
5	1	1	3	R139	139
2	5	1	1	R140	140
5	4	2	1	R141	141
2	2	2	1	R142	142
2	5	3	2	R143	143
4	3	3	3	R144	144
3	4	3	5	R145	145
5	1	3	4	R146	146
1	4	1	1	R147	147
1	4	2	5	R148	148
3	2	4	2	R149	149
4	1	2	5	R150	150
1	4	4	2	R151	151
5	2	4	2	R152	152
4	1	4	2	R153	153
5	3	3	1	R154	154
3	2	2	5	R155	155
3	3	2	2	R156	156
4	4	3	3	R157	157
1	3	4	1	R158	158
2	5	2	2	R159	159
3	3	1	3	R160	160
3	5	4	1	R161	161
1	1	5	2	R162	162
2	3	2	1	R163	163
2	5	1	4	R164	164
3	2	3	3	R165	165
1	1	2	2	R166	166
3	4	4	1	R167	167
3	4	4	1	R168	168
3	5	2	2	R169	169
2	3	3	5	R170	170
5	2	5	1	R171	171
4	2	5	3	R172	172
4	5	5	2	R173	173
4	3	3	4	R174	174
4	1	1	3	R175	175
5	1	5	3	R176	176
1	3	5	3	R177	177
5	1	2	4	R178	178
5	2	1	5	R179	179
2	1	4	4	R180	180
5	4	1	1	R181	181
3	2	1	5	R182	182
2	2	4	5	R183	183
5	4	1	1	R184	184
4	3	3	5	R185	185
5	4	3	5	R186	186
3	1	1	5	R187	187
2	1	1	2	R188	188
3	3	4	1	R189	189
3	2	4	5	R190	190
2	3	1	4	R191	191
5	2	1	4	R192	192
3	2	2	3	R193	193
4	2	2	4	R194	194
3	5	2	1	R195	195
4	1	2	4	R196	196
1	3	2	1	R197	197
5	4	4	1	R198	198
1	4	4	4	R199	199
4	2	2	1	R200	200
\.


--
-- Data for Name: s; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.s (b1, b2, b3, b4, ann, i) FROM stdin;
4	2	2	5	S1	1
3	3	4	1	S2	2
5	5	5	3	S3	3
5	3	5	2	S4	4
3	4	1	4	S5	5
5	5	4	5	S6	6
3	3	2	3	S7	7
2	3	5	5	S8	8
3	1	3	4	S9	9
1	5	4	1	S10	10
3	5	1	4	S11	11
2	1	5	5	S12	12
4	1	1	5	S13	13
3	5	3	2	S14	14
3	1	4	2	S15	15
1	4	4	4	S16	16
1	2	4	3	S17	17
1	3	5	4	S18	18
1	3	3	3	S19	19
3	1	4	2	S20	20
5	2	2	5	S21	21
4	1	4	2	S22	22
4	5	2	2	S23	23
4	4	4	2	S24	24
1	2	2	3	S25	25
4	4	1	4	S26	26
2	4	4	1	S27	27
1	2	4	4	S28	28
4	3	3	4	S29	29
3	1	5	5	S30	30
3	5	5	1	S31	31
3	1	5	1	S32	32
3	3	1	2	S33	33
4	3	3	3	S34	34
5	2	4	3	S35	35
5	3	2	4	S36	36
4	3	4	1	S37	37
1	4	4	5	S38	38
2	4	1	1	S39	39
4	3	2	4	S40	40
3	4	5	1	S41	41
3	2	3	4	S42	42
5	3	4	4	S43	43
4	3	4	5	S44	44
5	5	1	3	S45	45
3	3	5	4	S46	46
5	1	2	5	S47	47
3	1	4	4	S48	48
1	3	5	4	S49	49
4	5	2	2	S50	50
3	5	4	2	S51	51
4	5	2	2	S52	52
3	3	2	3	S53	53
5	4	2	3	S54	54
3	4	3	5	S55	55
1	4	3	4	S56	56
3	3	4	1	S57	57
3	3	3	3	S58	58
4	1	5	3	S59	59
1	5	5	3	S60	60
2	3	4	1	S61	61
2	2	5	3	S62	62
1	1	5	2	S63	63
4	1	5	5	S64	64
3	5	5	5	S65	65
3	5	4	4	S66	66
1	4	5	1	S67	67
1	3	1	5	S68	68
4	3	1	3	S69	69
4	2	3	2	S70	70
4	3	3	2	S71	71
2	4	3	4	S72	72
5	3	2	3	S73	73
5	5	2	2	S74	74
2	4	5	5	S75	75
1	4	4	2	S76	76
2	2	4	4	S77	77
2	2	2	2	S78	78
2	1	3	3	S79	79
1	2	5	1	S80	80
3	4	3	5	S81	81
2	2	1	3	S82	82
5	4	5	1	S83	83
4	2	1	2	S84	84
2	1	2	2	S85	85
3	1	4	4	S86	86
5	5	5	3	S87	87
4	2	1	5	S88	88
5	1	2	2	S89	89
3	1	3	5	S90	90
4	4	5	5	S91	91
1	5	3	3	S92	92
3	2	1	5	S93	93
4	3	3	4	S94	94
5	3	5	2	S95	95
4	2	3	2	S96	96
1	1	3	4	S97	97
2	1	1	3	S98	98
3	3	4	4	S99	99
2	1	5	2	S100	100
1	2	3	1	S101	101
1	2	1	4	S102	102
1	2	1	3	S103	103
4	4	2	3	S104	104
4	2	2	4	S105	105
2	4	2	5	S106	106
3	2	1	2	S107	107
3	3	1	3	S108	108
5	4	5	3	S109	109
3	2	1	3	S110	110
3	3	1	5	S111	111
3	3	1	2	S112	112
2	3	1	3	S113	113
2	3	4	2	S114	114
4	4	3	2	S115	115
4	4	3	4	S116	116
3	3	3	3	S117	117
3	2	3	4	S118	118
5	1	2	2	S119	119
3	3	3	2	S120	120
3	5	5	4	S121	121
2	3	3	4	S122	122
1	5	2	4	S123	123
3	3	3	2	S124	124
2	2	1	1	S125	125
2	5	2	2	S126	126
5	4	4	1	S127	127
5	3	2	1	S128	128
1	5	1	2	S129	129
5	1	4	2	S130	130
3	1	2	5	S131	131
2	1	5	5	S132	132
1	5	5	5	S133	133
5	2	5	4	S134	134
4	4	2	2	S135	135
5	2	4	5	S136	136
2	1	3	5	S137	137
5	1	4	1	S138	138
4	4	3	2	S139	139
3	4	1	5	S140	140
5	4	3	2	S141	141
3	4	1	4	S142	142
5	1	2	3	S143	143
5	4	1	3	S144	144
4	1	3	5	S145	145
1	1	4	4	S146	146
4	3	2	4	S147	147
4	5	2	3	S148	148
2	3	1	1	S149	149
4	4	1	2	S150	150
3	5	1	2	S151	151
3	2	4	1	S152	152
5	2	4	3	S153	153
1	3	2	5	S154	154
2	5	2	4	S155	155
5	3	1	5	S156	156
2	2	1	2	S157	157
1	3	1	3	S158	158
4	1	3	5	S159	159
5	1	2	1	S160	160
1	1	3	1	S161	161
4	2	1	3	S162	162
3	5	3	4	S163	163
4	1	1	3	S164	164
1	4	3	4	S165	165
1	3	1	5	S166	166
3	3	5	2	S167	167
5	2	2	5	S168	168
4	2	1	3	S169	169
3	5	5	3	S170	170
2	3	2	4	S171	171
5	4	5	5	S172	172
4	3	4	5	S173	173
4	2	1	5	S174	174
1	4	5	1	S175	175
2	1	5	4	S176	176
2	3	1	5	S177	177
4	3	1	2	S178	178
5	1	4	1	S179	179
3	1	1	2	S180	180
4	2	3	2	S181	181
2	2	1	2	S182	182
2	3	1	2	S183	183
2	2	3	2	S184	184
3	3	4	4	S185	185
4	4	2	2	S186	186
4	5	3	3	S187	187
2	4	2	2	S188	188
2	2	1	3	S189	189
4	1	2	3	S190	190
3	2	3	2	S191	191
5	1	3	5	S192	192
2	2	1	5	S193	193
2	2	2	5	S194	194
1	4	4	2	S195	195
4	3	2	1	S196	196
2	2	1	5	S197	197
4	3	3	2	S198	198
5	3	3	1	S199	199
4	4	2	5	S200	200
\.


--
-- Name: r_i_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.r_i_seq', 200, true);


--
-- Name: s_i_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.s_i_seq', 200, true);


--
-- Name: r r_ann_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.r
    ADD CONSTRAINT r_ann_key UNIQUE (ann);


--
-- Name: r r_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.r
    ADD CONSTRAINT r_pkey PRIMARY KEY (i);


--
-- Name: s s_ann_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.s
    ADD CONSTRAINT s_ann_key UNIQUE (ann);


--
-- Name: s s_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.s
    ADD CONSTRAINT s_pkey PRIMARY KEY (i);


--
-- PostgreSQL database dump complete
--

