-- Downloaded from: https://github.com/joaoporto27/Bora-Viajar-BackEnd/blob/1e95e1abb4109e554f55b5e8faa6804b45dfc3f1/backup/bora_viajar.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 16.3
-- Dumped by pg_dump version 16.3

-- Started on 2025-06-10 13:54:12

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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 215 (class 1259 OID 17687)
-- Name: comments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.comments (
    id integer NOT NULL,
    post_id integer,
    user_id integer,
    comment text
);


ALTER TABLE public.comments OWNER TO postgres;

--
-- TOC entry 216 (class 1259 OID 17692)
-- Name: comments_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.comments_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.comments_id_seq OWNER TO postgres;

--
-- TOC entry 4845 (class 0 OID 0)
-- Dependencies: 216
-- Name: comments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.comments_id_seq OWNED BY public.comments.id;


--
-- TOC entry 217 (class 1259 OID 17693)
-- Name: feedbacks; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.feedbacks (
    id integer NOT NULL,
    user_id integer,
    feedback text NOT NULL,
    rating integer,
    CONSTRAINT feedbacks_rating_check CHECK (((rating >= 1) AND (rating <= 5)))
);


ALTER TABLE public.feedbacks OWNER TO postgres;

--
-- TOC entry 218 (class 1259 OID 17699)
-- Name: feedbacks_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.feedbacks_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.feedbacks_id_seq OWNER TO postgres;

--
-- TOC entry 4846 (class 0 OID 0)
-- Dependencies: 218
-- Name: feedbacks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.feedbacks_id_seq OWNED BY public.feedbacks.id;


--
-- TOC entry 219 (class 1259 OID 17700)
-- Name: news; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.news (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    place character varying(255) NOT NULL,
    text text NOT NULL,
    image text
);


ALTER TABLE public.news OWNER TO postgres;

--
-- TOC entry 220 (class 1259 OID 17705)
-- Name: news_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.news_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.news_id_seq OWNER TO postgres;

--
-- TOC entry 4847 (class 0 OID 0)
-- Dependencies: 220
-- Name: news_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.news_id_seq OWNED BY public.news.id;


--
-- TOC entry 221 (class 1259 OID 17706)
-- Name: posts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.posts (
    id integer NOT NULL,
    user_id integer,
    image text,
    description text NOT NULL,
    tag character varying(100),
    likes_count integer DEFAULT 0
);


ALTER TABLE public.posts OWNER TO postgres;

--
-- TOC entry 222 (class 1259 OID 17712)
-- Name: posts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.posts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.posts_id_seq OWNER TO postgres;

--
-- TOC entry 4848 (class 0 OID 0)
-- Dependencies: 222
-- Name: posts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.posts_id_seq OWNED BY public.posts.id;


--
-- TOC entry 223 (class 1259 OID 17713)
-- Name: regions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.regions (
    id integer NOT NULL,
    region character varying(70) NOT NULL,
    name character varying(255) NOT NULL,
    state character varying(255) NOT NULL,
    text text NOT NULL,
    links character varying(255),
    image text
);


ALTER TABLE public.regions OWNER TO postgres;

--
-- TOC entry 224 (class 1259 OID 17718)
-- Name: regions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.regions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.regions_id_seq OWNER TO postgres;

--
-- TOC entry 4849 (class 0 OID 0)
-- Dependencies: 224
-- Name: regions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.regions_id_seq OWNED BY public.regions.id;


--
-- TOC entry 225 (class 1259 OID 17719)
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    city character varying(120) NOT NULL,
    state character varying(120) NOT NULL,
    password character varying(15) NOT NULL,
    photo text
);


ALTER TABLE public.users OWNER TO postgres;

--
-- TOC entry 226 (class 1259 OID 17724)
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_id_seq OWNER TO postgres;

--
-- TOC entry 4850 (class 0 OID 0)
-- Dependencies: 226
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- TOC entry 4659 (class 2604 OID 17725)
-- Name: comments id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comments ALTER COLUMN id SET DEFAULT nextval('public.comments_id_seq'::regclass);


--
-- TOC entry 4660 (class 2604 OID 17726)
-- Name: feedbacks id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.feedbacks ALTER COLUMN id SET DEFAULT nextval('public.feedbacks_id_seq'::regclass);


--
-- TOC entry 4661 (class 2604 OID 17727)
-- Name: news id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.news ALTER COLUMN id SET DEFAULT nextval('public.news_id_seq'::regclass);


--
-- TOC entry 4662 (class 2604 OID 17728)
-- Name: posts id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.posts ALTER COLUMN id SET DEFAULT nextval('public.posts_id_seq'::regclass);


--
-- TOC entry 4664 (class 2604 OID 17729)
-- Name: regions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.regions ALTER COLUMN id SET DEFAULT nextval('public.regions_id_seq'::regclass);


--
-- TOC entry 4665 (class 2604 OID 17730)
-- Name: users id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- TOC entry 4828 (class 0 OID 17687)
-- Dependencies: 215
-- Data for Name: comments; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.comments VALUES (1, 1, 10, 'Esse mirante em Bonito parece imperd¡vel! Quero muito conhecer.');
INSERT INTO public.comments VALUES (2, 2, 2, 'Adorei a ideia do passeio noturno em Ouro Preto, super diferente!');
INSERT INTO public.comments VALUES (3, 3, 7, 'Sarau na FLIP ‚ tudo! Excelente dica cultural.');
INSERT INTO public.comments VALUES (4, 4, 22, 'Recife sempre surpreende com eventos criativos, amei!');
INSERT INTO public.comments VALUES (5, 5, 3, 'Boa not¡cia para quem curte pedalar em Floripa!');
INSERT INTO public.comments VALUES (6, 6, 16, 'Fervedouros do JalapÆo sÆo mesmo £nicos, recomendad¡ssimo!');
INSERT INTO public.comments VALUES (7, 7, 17, 'Kitesurf em SÆo Miguel do Gostoso? J  quero reservar!');
INSERT INTO public.comments VALUES (8, 8, 18, 'Serra do Cip¢ ‚ perfeita para recarregar as energias.');
INSERT INTO public.comments VALUES (9, 9, 4, 'Tour noturno com vinho? Parece super romƒntico!');
INSERT INTO public.comments VALUES (10, 10, 11, 'Boa saber sobre a baixa das lagoas nos Len‡¢is, vou me programar melhor.');
INSERT INTO public.comments VALUES (11, 11, 46, 'Triste pela trilha fechada, mas que bom que tem outras op‡äes!');
INSERT INTO public.comments VALUES (12, 12, 12, 'Alerta importante, principalmente para quem viaja com crian‡as.');
INSERT INTO public.comments VALUES (13, 13, 34, 'Sempre bom lembrar dos hor rios de pico para a balsa de Ilhabela.');
INSERT INTO public.comments VALUES (14, 14, 14, 'Not¡cia importante pra quem faz passeios de barco. Valeu o aviso!');
INSERT INTO public.comments VALUES (15, 15, 15, 'Calor intenso em Palmas ‚ de respeito, obrigado pelo alerta!');
INSERT INTO public.comments VALUES (16, 16, 55, 'Promo‡Æo ¢tima pra curtir Gramado sem gastar tanto!');
INSERT INTO public.comments VALUES (17, 17, 17, 'Hospedagem flutuante com desconto? Que sonho!');
INSERT INTO public.comments VALUES (18, 18, 18, 'Adoro Porto de Galinhas! Com esse desconto, fica ainda melhor.');
INSERT INTO public.comments VALUES (19, 19, 17, 'àtimo combo! Cataratas e Parque das Aves sÆo imperd¡veis.');
INSERT INTO public.comments VALUES (20, 20, 20, 'Desconto excelente na Chapada Diamantina, vou aproveitar!');
INSERT INTO public.comments VALUES (21, 21, 44, 'A travessia de balsa em Ilhabela est  com longas filas. Cuidado se for viajar no fim de semana!');
INSERT INTO public.comments VALUES (22, 22, 15, 'àtima promo‡Æo em Len‡¢is, a nova sinaliza‡Æo nas trilhas da Chapada vai garantir mais seguran‡a! #Promo‡Æo');
INSERT INTO public.comments VALUES (23, 23, 31, 'Visitei Canela e fiquei encantado com a nova ilumina‡Æo da Catedral de Pedra … noite, vale muito a pena!');
INSERT INTO public.comments VALUES (24, 24, 6, 'O tour noturno com degusta‡Æo no Vale dos Vinhedos foi sensacional! Vale a pena conferir as vin¡colas!');
INSERT INTO public.comments VALUES (25, 25, 11, 'Desconto especial para quem visitar as Cataratas e o Parque das Aves em Foz do Igua‡u! Aproveitem!');
INSERT INTO public.comments VALUES (26, 26, 3, 'O Mercado Ver-o-Peso em Bel‚m foi restaurado e est  com uma nova  rea gastron“mica incr¡vel!');
INSERT INTO public.comments VALUES (27, 27, 8, 'A orla nova de JoÆo Pessoa com ciclovia e feirinhas noturnas ficou maravilhosa. Super recomendo!');
INSERT INTO public.comments VALUES (28, 28, 17, 'Passeio de buggy pelas dunas de Natal com 20% de desconto, nÆo d  para perder!');
INSERT INTO public.comments VALUES (29, 29, 33, 'Acabei de voltar do Vale dos Vinhedos e o tour noturno com degusta‡Æo foi uma experiˆncia incr¡vel!');
INSERT INTO public.comments VALUES (30, 30, 6, 'SÆo Miguel do Gostoso ‚ um para¡so para kitesurf! Praia calma e vento constante, perfeito!');
INSERT INTO public.comments VALUES (31, 31, 45, 'A ilumina‡Æo da Catedral de Pedra em Canela ‚ ainda mais bonita … noite. Vale a pena a visita!');
INSERT INTO public.comments VALUES (32, 32, 37, 'NÆo perca a chance de fazer o passeio de buggy pelas dunas de Natal com 20% de desconto!');
INSERT INTO public.comments VALUES (33, 33, 30, 'Acesso … Praia do Farol em Arraial do Cabo est  restrito por excesso de turistas. Planeje com antecedˆncia.');
INSERT INTO public.comments VALUES (34, 34, 23, 'A orla nova de JoÆo Pessoa com ciclovia e feirinhas noturnas ficou ainda mais encantadora!');
INSERT INTO public.comments VALUES (35, 35, 51, 'Pousadas com di ria reduzida em Pipa no meio da semana! Perfeito para quem quer aproveitar mais com menos!');
INSERT INTO public.comments VALUES (36, 36, 18, 'Em Pipa, as pousadas estÆo com di ria reduzida no meio da semana. Excelente para quem quer economizar!');
INSERT INTO public.comments VALUES (37, 37, 20, 'Fui a Arraial do Cabo e adorei! Acesso … Praia do Farol est  restrito por excesso de turistas, entÆo se planeje!');
INSERT INTO public.comments VALUES (38, 38, 19, 'Desconto especial para quem visita as Cataratas e o Parque das Aves em Foz do Igua‡u!');
INSERT INTO public.comments VALUES (39, 39, 12, 'Forte calor nas trilhas do Parque do Lajeado em Palmas! Leve bastante  gua!');
INSERT INTO public.comments VALUES (40, 40, 8, 'Pousadas com di ria reduzida em Pipa no meio da semana! Excelente oportunidade!');
INSERT INTO public.comments VALUES (41, 41, 60, 'Os fervedouros de JalapÆo sÆo incr¡veis! A paisagem ‚ surreal, um lugar perfeito para aventura!');
INSERT INTO public.comments VALUES (42, 42, 5, 'O Mercado Ver-o-Peso em Bel‚m est  renovado e com uma nova  rea gastron“mica. Vale a pena visitar!');
INSERT INTO public.comments VALUES (43, 43, 40, 'Aproveite o desconto em passeios de buggy em Natal! Passeio pelas dunas imperd¡vel!');
INSERT INTO public.comments VALUES (44, 44, 23, 'A ilumina‡Æo da Catedral de Pedra em Canela est  incr¡vel … noite, imperd¡vel!');
INSERT INTO public.comments VALUES (45, 45, 26, 'Tour noturno com degusta‡Æo nas vin¡colas de Vale dos Vinhedos. Uma experiˆncia maravilhosa!');
INSERT INTO public.comments VALUES (46, 46, 10, 'Em Palmas, o calor nas trilhas do Parque do Lajeado ‚ intenso. Leve muita  gua!');
INSERT INTO public.comments VALUES (47, 47, 16, 'Desconto especial em Foz do Igua‡u para quem visita as Cataratas e o Parque das Aves. NÆo perca!');
INSERT INTO public.comments VALUES (48, 48, 4, 'A ilumina‡Æo da Catedral de Pedra em Canela est  ainda mais deslumbrante … noite. Imperd¡vel!');
INSERT INTO public.comments VALUES (49, 49, 55, 'Fervedouros de JalapÆo sÆo impressionantes! Paisagens incr¡veis para quem ama ecoturismo.');
INSERT INTO public.comments VALUES (50, 50, 57, 'Aproveite a promo‡Æo em Pipa com di rias reduzidas no meio da semana!');
INSERT INTO public.comments VALUES (51, 51, 12, 'Acesso … Praia do Farol em Arraial do Cabo est  limitado devido ao excesso de turistas. Fique atento!');
INSERT INTO public.comments VALUES (52, 52, 38, 'Em Canela, a ilumina‡Æo da Catedral de Pedra … noite ‚ um espet culo! Vale cada minuto da visita!');
INSERT INTO public.comments VALUES (53, 53, 7, 'A Serra do Cip¢ tem trilhas lindas e cachoeiras cristalinas, ¢timo para ecoturismo!');
INSERT INTO public.comments VALUES (54, 54, 11, 'A travessia de balsa em Ilhabela est  com longas filas nos fins de semana. Se planeje!');
INSERT INTO public.comments VALUES (55, 55, 46, 'Mercado Ver-o-Peso em Bel‚m foi renovado com uma  rea gastron“mica incr¡vel. Muito legal!');
INSERT INTO public.comments VALUES (56, 56, 51, 'A travessia de balsa em Ilhabela est  com longas filas nos fins de semana. Melhor evitar esse pico!');
INSERT INTO public.comments VALUES (57, 57, 42, 'Em Serra do Cip¢, as trilhas sÆo maravilhosas e as cachoeiras cristalinas sÆo imperd¡veis!');
INSERT INTO public.comments VALUES (58, 58, 18, 'Fervedouros em JalapÆo sÆo uma experiˆncia £nica. Paisagens de tirar o f“lego!');
INSERT INTO public.comments VALUES (59, 59, 53, 'Desconto de 30% para pacotes de reservas em Porto de Galinhas at‚ domingo! Aproveite!');
INSERT INTO public.comments VALUES (60, 60, 26, 'Pousadas em Pipa com di rias reduzidas durante a semana. Uma ¢tima oportunidade de aproveitar mais!');
INSERT INTO public.comments VALUES (61, 61, 12, 'Desconto em passeios de telef‚rico em Balne rio Cambori£ na temporada de inverno. NÆo perca!');
INSERT INTO public.comments VALUES (62, 82, 47, 'Promo‡Æo nos passeios de telef‚rico em Balne rio Cambori£! Aproveite o inverno com desconto!');
INSERT INTO public.comments VALUES (63, 63, 8, 'Pousadas com di ria reduzida em Pipa durante o meio da semana. Super recomendada!');
INSERT INTO public.comments VALUES (64, 64, 26, 'Desconto de 30% para pacotes de Porto de Galinhas at‚ domingo! Aproveite!');
INSERT INTO public.comments VALUES (65, 65, 19, 'Novo circuito de tirolesas em Brotas! Uma aventura e tanto!');
INSERT INTO public.comments VALUES (66, 66, 2, 'Mercado Ver-o-Peso em Bel‚m est  com nova  rea gastron“mica. Vale muito a pena!');
INSERT INTO public.comments VALUES (67, 67, 45, 'Novo circuito de tirolesas em Brotas ‚ uma diversÆo para os aventureiros!');
INSERT INTO public.comments VALUES (68, 68, 15, 'Os fervedouros de JalapÆo sÆo impressionantes. Paisagens de outro mundo!');
INSERT INTO public.comments VALUES (69, 69, 38, 'Desconto especial em Foz do Igua‡u para quem visita as Cataratas e o Parque das Aves. Aproveite!');
INSERT INTO public.comments VALUES (70, 70, 53, 'A travessia de balsa em Ilhabela est  com longas filas nos fins de semana. Se organize!');
INSERT INTO public.comments VALUES (71, 71, 41, 'O tour noturno com degusta‡Æo nas vin¡colas do Vale dos Vinhedos ‚ uma experiˆncia £nica!');
INSERT INTO public.comments VALUES (72, 72, 36, 'Pousadas com diárias reduzidas durante a semana em Pipa. Ótima oportunidade!');
INSERT INTO public.comments VALUES (86, 172, 1, 'Esse mirante da Nascente Azul em Bonito parece incr¡vel! Mal posso esperar para visitar.');
INSERT INTO public.comments VALUES (87, 173, 12, 'A rota noturna em Ouro Preto parece m gica, j  est  na minha lista!');
INSERT INTO public.comments VALUES (88, 174, 53, 'Boa dica sobre a interdi‡Æo da trilha na Chapada dos Veadeiros, obrigado pelo alerta!');
INSERT INTO public.comments VALUES (89, 175, 14, 'Sarau liter rio na FLIP deve ser maravilhoso, Paraty ‚ realmente um lugar cultural.');
INSERT INTO public.comments VALUES (90, 176, 55, 'Festival Recife na Rua ‚ um dos melhores eventos, ¢tima dica para quem ama cultura local!');
INSERT INTO public.comments VALUES (91, 177, 26, 'Obrigado pelo alerta da correnteza na Praia da Joaquina, seguran‡a em primeiro lugar.');
INSERT INTO public.comments VALUES (92, 178, 97, 'A nova ciclovia em Florian¢polis vai facilitar muito os passeios em fam¡lia.');
INSERT INTO public.comments VALUES (93, 179, 18, 'Os fervedouros do JalapÆo sÆo espetaculares, uma experiˆncia £nica!');
INSERT INTO public.comments VALUES (94, 180, 9, 'Valeu pelo aviso das filas na balsa em Ilhabela, vou planejar para sair cedo.');
INSERT INTO public.comments VALUES (95, 181, 11, 'SÆo Miguel do Gostoso parece o destino perfeito para relaxar e curtir o kitesurf.');
INSERT INTO public.comments VALUES (96, 182, 61, 'àtima promo‡Æo para quem quer aproveitar as Cataratas e o Parque das Aves juntos!');
INSERT INTO public.comments VALUES (97, 183, 82, 'Degusta‡Æo noturna no Vale dos Vinhedos deve ser uma experiˆncia inesquec¡vel!');
INSERT INTO public.comments VALUES (98, 184, 13, 'Acesso restrito na Praia do Farol ‚ importante para preservar o local, obrigado pelo aviso!');
INSERT INTO public.comments VALUES (99, 72, 36, 'Pousadas com diárias reduzidas durante a semana em Pipa. Ótima oportunidade!');
INSERT INTO public.comments VALUES (100, 2, 44, 'Viagem maravilhosa para Gramado!');


--
-- TOC entry 4830 (class 0 OID 17693)
-- Dependencies: 217
-- Data for Name: feedbacks; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.feedbacks VALUES (1, 11, 'O Bora Viajar me ajudou muito a montar um roteiro incr¡vel pelo litoral nordestino!', 5);
INSERT INTO public.feedbacks VALUES (2, 42, 'Achei as informa‡äes sobre seguran‡a muito £teis durante minha viagem a Manaus.', 4);
INSERT INTO public.feedbacks VALUES (3, 3, 'O app travou algumas vezes durante o uso, mas no geral ‚ bom.', 3);
INSERT INTO public.feedbacks VALUES (4, 48, 'Fiquei encantado com as dicas de outros viajantes. Super recomendo!', 5);
INSERT INTO public.feedbacks VALUES (5, 1, 'Faltam informa‡äes sobre transporte local em algumas cidades.', 2);
INSERT INTO public.feedbacks VALUES (6, 56, 'Os roteiros criados pelos profissionais sÆo bem completos e f ceis de seguir.', 5);
INSERT INTO public.feedbacks VALUES (7, 53, 'Gostei bastante da interface do site, muito intuitiva.', 4);
INSERT INTO public.feedbacks VALUES (8, 27, 'Tive dificuldade em encontrar excursäes atualizadas para o JalapÆo.', 2);
INSERT INTO public.feedbacks VALUES (9, 39, 'Excelente ferramenta para planejar viagens em fam¡lia!', 5);
INSERT INTO public.feedbacks VALUES (10, 19, 'O sistema de avalia‡Æo dos destinos poderia ser mais transparente.', 3);
INSERT INTO public.feedbacks VALUES (11, 8, 'Usei o app para planejar minha lua de mel e foi perfeito!', 5);
INSERT INTO public.feedbacks VALUES (12, 15, 'Algumas informa‡äes estavam desatualizadas, principalmente sobre hospedagem.', 3);
INSERT INTO public.feedbacks VALUES (13, 21, 'Adorei poder ver relatos de outros usu rios antes de decidir meu roteiro.', 5);
INSERT INTO public.feedbacks VALUES (14, 32, 'A op‡Æo de roteiros personalizados foi o que mais me atraiu no aplicativo.', 4);
INSERT INTO public.feedbacks VALUES (15, 44, 'Precisa melhorar o suporte ao cliente dentro do app.', 2);
INSERT INTO public.feedbacks VALUES (16, 42, 'Com o Bora Viajar descobri lugares incr¡veis que nunca tinha ouvido falar.', 5);
INSERT INTO public.feedbacks VALUES (17, 49, 'Muito bom para quem viaja sozinho. As dicas sÆo bem espec¡ficas e pr ticas.', 4);
INSERT INTO public.feedbacks VALUES (18, 37, 'Fiquei decepcionada com a falta de op‡äes para o interior de alguns estados.', 2);
INSERT INTO public.feedbacks VALUES (19, 47, 'Facilitou muito meu mochilÆo pelo sul do Brasil.', 5);
INSERT INTO public.feedbacks VALUES (20, 17, 'Poderia ter integra‡Æo com apps de mapas como Google Maps ou Waze.', 3);


--
-- TOC entry 4832 (class 0 OID 17700)
-- Dependencies: 219
-- Data for Name: news; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.news VALUES (1, 'Festival de Cinema agita a Serra Gaucha com novas atracoes gastronomicas', 'Gramado (RS)', 'A charmosa cidade de Gramado esta em clima de celebracao com o aguardado Festival de Cinema de 2025. O evento, que atrai artistas e turistas de todo o Brasil, traz neste ano uma mostra dedicada ao cinema latino-americano contemporaneo. Alem disso, a prefeitura lancou a Rota do Cafe Colonial Artesanal, reunindo 18 produtores locais que oferecem experiencias unicas de sabores coloniais em ambientes rusticos e acolhedores. A iluminacao tematica nas ruas centrais completa o cenario encantador.', '1748537466171-Gramado.jpg');
INSERT INTO public.news VALUES (2, 'Novo mirante e passarelas ecologicas valorizam o ecoturismo', 'Bonito (MS)', 'Bonito continua sendo referencia em turismo sustentavel, e acaba de inaugurar o "Mirante da Nascente Azul", com passarelas suspensas que oferecem uma visao aerea das aguas cristalinas e da fauna local. Com foco na preservacao e na acessibilidade, o complexo agora conta tambem com trilhas inclusivas para cadeirantes e idosos, alem de areas de descanso com sombra e sinalizacao ambiental educativa. A cidade recebe ainda, neste mes, a Feira de Turismo Ecologico do Centro-Oeste, reunindo especialistas e operadores de turismo.', '1748537518155-Bonito.jpg');
INSERT INTO public.news VALUES (3, 'Festival de esportes aquaticos e novas trilhas de aventura', 'Jericoacoara (CE)', 'A paradisiaca vila de Jericoacoara realizou a primeira edicao do Festival de Kite e Windsurf Sustentavel, que reuniu atletas de todo o pais em competicoes que valorizam praticas de baixo impacto ambiental. Novas trilhas ecologicas para quadriciclos e caminhadas foram abertas na area da Pedra Furada, com sinalizacao bilingue e mirantes panoramicos. O por do sol nas dunas continua sendo uma atracao imperdivel, agora com apresentacoes culturais ao vivo nos fins de tarde.', '1748537572480-Jericoacara.jpg');
INSERT INTO public.news VALUES (4, 'Patrimonio cultural em destaque na celebracao dos 60 anos da FLIP', 'Paraty (RJ)', 'Paraty se transformou num grande palco literario com a comemoracao dos 60 anos da Festa Literaria Internacional. A cidade recebe autores de renome, debates sobre literatura indigena e oficinas de escrita criativa em casaroes historicos. Uma das grandes novidades e a reabertura do Forte Defensor Perpetuo, agora restaurado com um museu interativo e trilhas que oferecem vista panoramica da baia de Paraty. Restaurantes locais tambem entraram na festa com menus especiais inspirados em livros e personagens literarios.', '1748537659355-Paraty.jpg');
INSERT INTO public.news VALUES (5, 'Caribe amazonico ganha hotel flutuante e turismo sustentavel cresce', 'Alter do Chao (PA)', 'As aguas claras do Rio Tapajos estao ainda mais atrativas com a chegada do primeiro hotel flutuante ecologico da regiao, operando com energia solar, coleta de agua da chuva e cardapios que valorizam a culinaria ribeirinha. A alta temporada das praias de rio esta a todo vapor, com destaque para a Ilha do Amor, que recebeu nova sinalizacao turistica e quiosques padronizados com foco em limpeza e organizacao. Passeios guiados de canoa agora incluem oficinas sobre plantas medicinais da floresta.', '1748537769311-AlterDoChÃ£o.jpg');
INSERT INTO public.news VALUES (6, 'Concertos barrocos e nova rota noturna celebram 40 anos de patrimonio mundial', 'Ouro Preto (MG)', 'Ouro Preto comemora quatro decadas como Patrimonio Mundial da UNESCO com uma agenda intensa de atracoes. As igrejas do centro historico sediam concertos noturnos com musica barroca ao vivo, criando uma atmosfera magica nas ladeiras coloniais. A nova "Rota da Luz e Historia", um passeio noturno guiado, utiliza iluminacao cenica para valorizar os principais monumentos e casaroes. A Casa dos Contos tambem abriu uma nova exposicao interativa sobre a Inconfidencia Mineira.', '1748537831783-OuroPreto.jpg');
INSERT INTO public.news VALUES (7, 'Realidade aumentada e novas trilhas inclusivas encantam visitantes', 'Chapada dos Veadeiros (GO)', 'O Parque Nacional da Chapada dos Veadeiros acaba de ganhar um moderno centro de visitantes com experiencias em realidade aumentada, simulando voos de drone sobre os canions e cachoeiras mais iconicos da regiao. As trilhas principais foram reformadas com passarelas ecologicas e sinalizacao em braile, tornando o ecoturismo ainda mais acessivel. Restaurantes e pousadas locais estao adotando praticas sustentaveis, com destaque para a gastronomia baseada em ingredientes do cerrado.', '1748537871494-ChapadaDosVeadeiros.jpg');
INSERT INTO public.news VALUES (8, 'Recife Antigo revive com arte de rua e nova fase do Cais do Sertao', 'Recife (PE)', 'O bairro historico do Recife Antigo vive um momento vibrante com a reabertura ampliada do museu Cais do Sertao, que agora conta com salas interativas sobre o ciclo do forro e da poesia nordestina. As ruas do bairro estao tomadas pelo Festival Recife na Rua, com musica ao vivo, gastronomia regional, teatro de rua e maracatus que desfilam entre os casaroes coloniais. A ciclovia recem-inaugurada ligando o Marco Zero a Praia de Boa Viagem estimula o turismo sustentavel.', '1748537907114-Recife.jpg');
INSERT INTO public.news VALUES (9, 'Ilha da Magia promove festival vegano e amplia ciclovias litoraneas', 'Florianopolis (SC)', 'Florianopolis esta em clima de consciencia ecologica com o Floripa Veg Festival, que reune chefs renomados, produtores locais e paineis sobre alimentacao consciente e saude. O evento ocorre simultaneamente em varias praias e centros culturais da cidade. Novas ciclovias foram concluidas entre a Praia Mole, Joaquina e o centrinho da Lagoa da Conceicao, promovendo o transporte alternativo e o contato com a natureza. Trilhas autoguiadas com QR Codes foram instaladas no Morro da Cruz e no Costao do Santinho.', '1748537949689-Florianopolis.jpg');
INSERT INTO public.news VALUES (10, 'Temporada de lagoas cheias traz passaporte turistico e novos voos', 'Lencois Maranhenses (MA)', 'A beleza unica dos Lencois Maranhenses esta ainda mais acessivel com a chegada de novos voos diretos para Barreirinhas, principal porta de entrada do parque. A temporada de lagoas cheias esta deslumbrante, e os visitantes agora podem participar do programa "Passaporte dos Lencois", que premia quem explora diferentes circuitos como Lagoa Azul, Lagoa Bonita e Canto do Atins. Barqueiros e guias locais estao sendo capacitados para oferecer experiencias mais seguras e informativas aos turistas.', '1748537992180-LencoisMaranhenses.jpg');


--
-- TOC entry 4834 (class 0 OID 17706)
-- Dependencies: 221
-- Data for Name: posts; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.posts VALUES (2, 88, NULL, 'PROMO€ÇO imperd¡vel em Gramado! Reserve 3 noites e ganhe um jantar colonial completo em um dos caf‚s artesanais parceiros da nova rota gourmet.', 'PROMO€ÇO', 4);
INSERT INTO public.posts VALUES (3, 7, NULL, 'PROMO€ÇO em Alter do ChÆo! Hospedagem flutuante com desconto para reservas feitas at‚ o fim do mˆs. Ideal pra quem quer dormir com vista pro rio.', 'PROMO€ÇO', 4);
INSERT INTO public.posts VALUES (4, 61, NULL, 'PROMO€ÇO em Porto de Galinhas (PE): pacotes com desconto de 30% para quem reserva at‚ domingo. àtimo para fam¡lias com crian‡as!', 'PROMO€ÇO', 4);
INSERT INTO public.posts VALUES (5, 53, NULL, 'ALERTA: Em Manaus (AM), o porto est  parcialmente fechado para embarques tur¡sticos devido … cheia do rio Negro. Verifiquem antes de comprar passeio.', 'ALERTA', 4);
INSERT INTO public.posts VALUES (6, 78, NULL, 'A Serra do Cip¢ (MG) me surpreendeu! Trilhas lindas, cachoeiras de  guas cristalinas e um clima perfeito pra quem curte natureza e sossego.', 'NOVIDADES', 8);
INSERT INTO public.posts VALUES (9, 14, NULL, 'ALERTA: A travessia de balsa est  com longas filas nos fins de semana. (Ilhabela (SP))', 'ALERTA', 8);
INSERT INTO public.posts VALUES (11, 31, NULL, 'PROMO€ÇO em Canela (RS)! Nova ilumina‡Æo na Catedral de Pedra est  deslumbrante … noite.', 'PROMO€ÇO', 8);
INSERT INTO public.posts VALUES (12, 90, NULL, 'ALERTA: Tour noturno com degusta‡Æo, vin¡colas encantadoras. (Vale dos Vinhedos (RS))', 'ALERTA', 8);
INSERT INTO public.posts VALUES (13, 56, NULL, 'PROMO€ÇO em Foz do Igua‡u (PR)! Desconto especial para quem visita as Cataratas e o Parque das Aves.', 'PROMO€ÇO', 1);
INSERT INTO public.posts VALUES (15, 48, NULL, 'PROMO€ÇO em JoÆo Pessoa (PB)! Orla nova com ciclovia e feirinhas noturnas animadas.', 'PROMO€ÇO', 1);
INSERT INTO public.posts VALUES (19, 7, NULL, 'Aproveite a PROMO€ÇO em Natal (RN): Passeio de buggy pelas dunas est  com 20% de desconto.', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (20, 30, NULL, 'Aproveite a PROMO€ÇO em Arraial do Cabo (RJ): ALERTA: Acesso limitado … Praia do Farol por excesso de turistas.', 'PROMO€ÇO', 1);
INSERT INTO public.posts VALUES (22, 51, NULL, 'ATEN€ÇO: Pousadas com di ria reduzida durante o meio da semana. (Pipa (RN))', 'ALERTA', 1);
INSERT INTO public.posts VALUES (25, 60, NULL, 'Aproveite a PROMO€ÇO em JalapÆo (TO): Os fervedouros sÆo impressionantes e a paisagem ‚ surreal.', 'PROMO€ÇO', 20);
INSERT INTO public.posts VALUES (29, 78, NULL, 'Aproveite a PROMO€ÇO em Vale dos Vinhedos (RS): Tour noturno com degusta‡Æo, vin¡colas encantadoras.', 'PROMO€ÇO', 20);
INSERT INTO public.posts VALUES (32, 7, NULL, 'Aproveite a PROMO€ÇO em Pipa (RN): Pousadas com di ria reduzida durante o meio da semana.', 'PROMO€ÇO', 10);
INSERT INTO public.posts VALUES (40, 22, NULL, 'ATEN€ÇO: PROMO€ÇO em passeios de telef‚rico na temporada de inverno. (Balne rio Cambori£ (SC))', 'ALERTA', 6);
INSERT INTO public.posts VALUES (38, 53, NULL, 'ATEN€ÇO: Pacotes com desconto de 30% para reservas at‚ domingo. (Porto de Galinhas (PE))', 'ALERTA', 5);
INSERT INTO public.posts VALUES (8, 21, NULL, 'ALERTA: Forte calor em Palmas (TO), especialmente nas trilhas do Parque Estadual do Lajeado. Leve bastante  gua e evite caminhadas no hor rio de pico.', 'ALERTA', 5);
INSERT INTO public.posts VALUES (37, 32, NULL, 'Aproveite a PROMO€ÇO em JalapÆo (TO): Os fervedouros sÆo impressionantes e a paisagem ‚ surreal.', 'PROMO€ÇO', 15);
INSERT INTO public.posts VALUES (35, 51, NULL, 'Aproveite a PROMO€ÇO em Ilhabela (SP): A travessia de balsa est  com longas filas nos fins de semana.', 'PROMO€ÇO', 16);
INSERT INTO public.posts VALUES (45, 38, NULL, 'Aproveite a PROMO€ÇO em Foz do Igua‡u (PR): Desconto especial para quem visita as Cataratas e o Parque das Aves.', 'PROMO€ÇO', 6);
INSERT INTO public.posts VALUES (48, 78, NULL, 'ATEN€ÇO: O porto est  parcialmente fechado devido … cheia do rio Negro. (Manaus (AM))', 'ALERTA', 6);
INSERT INTO public.posts VALUES (49, 29, NULL, 'Aproveite a PROMO€ÇO em Manaus (AM): O porto est  parcialmente fechado devido … cheia do rio Negro.', 'PROMO€ÇO', 6);
INSERT INTO public.posts VALUES (33, 12, NULL, 'Aproveite a PROMO€ÇO em Arraial do Cabo (RJ): ALERTA: Acesso limitado … Praia do Farol por excesso de turistas.', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (34, 31, NULL, 'ATEN€ÇO: A travessia de balsa est  com longas filas nos fins de semana. (Ilhabela (SP))', 'ALERTA', 11);
INSERT INTO public.posts VALUES (39, 33, NULL, 'ATEN€ÇO: Pousadas com di ria reduzida durante o meio da semana. (Pipa (RN))', 'ALERTA', 15);
INSERT INTO public.posts VALUES (23, 18, NULL, 'Aproveite a PROMO€ÇO em Pipa (RN): Pousadas com di ria reduzida durante o meio da semana.', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (21, 23, NULL, 'ATEN€ÇO: Orla nova com ciclovia e feirinhas noturnas animadas. (JoÆo Pessoa (PB))', 'ALERTA', 5);
INSERT INTO public.posts VALUES (14, 13, NULL, 'PROMO€ÇO em Bel‚m (PA)! Mercado Ver-o-Peso restaurado com nova  rea gastron“mica.', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (17, 8, NULL, 'Aproveite a PROMO€ÇO em SÆo Miguel do Gostoso (RN): Praia calma e vento constante, ¢timo para kitesurf.', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (18, 88, NULL, 'Aproveite a PROMO€ÇO em Canela (RS): Nova ilumina‡Æo na Catedral de Pedra est  deslumbrante … noite.', 'PROMO€ÇO', 6);
INSERT INTO public.posts VALUES (10, 10, NULL, 'PROMO€ÇO em Len‡¢is (BA)! Nova sinaliza‡Æo nas trilhas da Chapada garante mais seguran‡a.', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (16, 17, NULL, 'PROMO€ÇO em Natal (RN)! Passeio de buggy pelas dunas est  com 20% de desconto.', 'PROMO€ÇO', 2);
INSERT INTO public.posts VALUES (24, 8, NULL, 'Aproveite a PROMO€ÇO em Pipa (RN): Pousadas com di ria reduzida durante o meio da semana.', 'PROMO€ÇO', 2);
INSERT INTO public.posts VALUES (26, 5, NULL, 'ATEN€ÇO: Mercado Ver-o-Peso restaurado com nova  rea gastron“mica. (Bel‚m (PA))', 'ALERTA', 2);
INSERT INTO public.posts VALUES (27, 40, NULL, 'Aproveite a PROMO€ÇO em Natal (RN): Passeio de buggy pelas dunas est  com 20% de desconto.', 'PROMO€ÇO', 2);
INSERT INTO public.posts VALUES (28, 23, NULL, 'ATEN€ÇO: Nova ilumina‡Æo na Catedral de Pedra est  deslumbrante … noite. (Canela (RS))', 'ALERTA', 2);
INSERT INTO public.posts VALUES (30, 4, NULL, 'ATEN€ÇO: Nova ilumina‡Æo na Catedral de Pedra est  deslumbrante … noite. (Canela (RS))', 'ALERTA', 2);
INSERT INTO public.posts VALUES (31, 55, NULL, 'Aproveite a PROMO€ÇO em JalapÆo (TO): Os fervedouros sÆo impressionantes e a paisagem ‚ surreal.', 'PROMO€ÇO', 2);
INSERT INTO public.posts VALUES (42, 8, NULL, 'ATEN€ÇO: Pousadas com di ria reduzida durante o meio da semana. (Pipa (RN))', 'ALERTA', 2);
INSERT INTO public.posts VALUES (44, 2, NULL, 'Aproveite a PROMO€ÇO em Bel‚m (PA): Mercado Ver-o-Peso restaurado com nova  rea gastron“mica.', 'PROMO€ÇO', 2);
INSERT INTO public.posts VALUES (46, 21, NULL, 'ATEN€ÇO: Tour noturno com degusta‡Æo, vin¡colas encantadoras. (Vale dos Vinhedos (RS))', 'ALERTA', 2);
INSERT INTO public.posts VALUES (50, 22, NULL, 'ATEN€ÇO: Praia calma e vento constante, ¢timo para kitesurf. (SÆo Miguel do Gostoso (RN))', 'ALERTA', 2);
INSERT INTO public.posts VALUES (36, 42, NULL, 'ATEN€ÇO: Trilhas lindas e cachoeiras cristalinas, perfeito para ecoturismo. (Serra do Cip¢ (MG))', 'ALERTA', 5);
INSERT INTO public.posts VALUES (41, 37, NULL, 'Aproveite a PROMO€ÇO em Balne rio Cambori£ (SC): PROMO€ÇO em passeios de telef‚rico na temporada de inverno.', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (43, 19, NULL, 'ATEN€ÇO: Novo circuito de tirolesas foi inaugurado no parque de aventuras. (Brotas (SP))', 'ALERTA', 5);
INSERT INTO public.posts VALUES (47, 58, NULL, 'Aproveite a PROMO€ÇO em JalapÆo (TO): Os fervedouros sÆo impressionantes e a paisagem ‚ surreal.', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (105, 21, NULL, 'Aproveite a super oferta para conhecer a Rota das Emo‡äes!', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (56, 59, NULL, 'Aproveite a PROMO€ÇO em Brotas (SP): Novo circuito de tirolesas foi inaugurado no parque de aventuras.', 'PROMO€ÇO', 0);
INSERT INTO public.posts VALUES (57, 3, NULL, 'ATEN€ÇO: Passeio de buggy pelas dunas est  com 20% de desconto. (Natal (RN))', 'ALERTA', 0);
INSERT INTO public.posts VALUES (59, 27, NULL, 'Aproveite a PROMO€ÇO em JoÆo Pessoa (PB): Orla nova com ciclovia e feirinhas noturnas animadas.', 'PROMO€ÇO', 0);
INSERT INTO public.posts VALUES (61, 32, NULL, 'ATEN€ÇO: Praia calma e vento constante, ¢timo para kitesurf. (SÆo Miguel do Gostoso (RN))', 'ALERTA', 0);
INSERT INTO public.posts VALUES (63, 54, NULL, 'Aproveite a PROMO€ÇO em Len‡¢is (BA): Nova sinaliza‡Æo nas trilhas da Chapada garante mais seguran‡a.', 'PROMO€ÇO', 0);
INSERT INTO public.posts VALUES (68, 56, NULL, 'ATEN€ÇO: Tour noturno com degusta‡Æo, vin¡colas encantadoras. (Vale dos Vinhedos (RS))', 'ALERTA', 0);
INSERT INTO public.posts VALUES (69, 7, NULL, 'ATEN€ÇO: Os fervedouros sÆo impressionantes e a paisagem ‚ surreal. (JalapÆo (TO))', 'ALERTA', 0);
INSERT INTO public.posts VALUES (71, 40, NULL, 'ATEN€ÇO: PROMO€ÇO em passeios de telef‚rico na temporada de inverno. (Balne rio Cambori£ (SC))', 'ALERTA', 0);
INSERT INTO public.posts VALUES (72, 45, NULL, 'Aproveite a PROMO€ÇO em Len‡¢is (BA): Nova sinaliza‡Æo nas trilhas da Chapada garante mais seguran‡a.', 'PROMO€ÇO', 0);
INSERT INTO public.posts VALUES (77, 13, NULL, 'Aproveite a PROMO€ÇO em Palmas (TO): Forte calor nas trilhas do Parque do Lajeado, leve bastante  gua.', 'PROMO€ÇO', 0);
INSERT INTO public.posts VALUES (81, 54, NULL, 'ATEN€ÇO: Trilhas lindas e cachoeiras cristalinas, perfeito para ecoturismo. (Serra do Cip¢ (MG))', 'ALERTA', 0);
INSERT INTO public.posts VALUES (82, 48, NULL, 'Aproveite a PROMO€ÇO em Ilhabela (SP): A travessia de balsa est  com longas filas nos fins de semana.', 'PROMO€ÇO', 0);
INSERT INTO public.posts VALUES (88, 44, NULL, 'O Mercado Ver-o-Peso em Bel‚m (PA) ficou ainda melhor com a nova  rea gastron“mica. Perfeito para quem ama sabores ex¢ticos!', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (90, 19, NULL, 'ALERTA em SÆo Miguel do Gostoso (RN): O vento est  muito forte, entÆo, se vocˆ vai praticar kitesurf, tome cuidado.', 'ALERTA', 0);
INSERT INTO public.posts VALUES (99, 88, NULL, 'Descontos de at‚ 40% para viagens … Serra do Cip¢!', 'PROMO€ÇO', 0);
INSERT INTO public.posts VALUES (102, 18, NULL, 'Pacotes promocionais para conhecer o JalapÆo com guias locais.', 'PROMO€ÇO', 7);
INSERT INTO public.posts VALUES (60, 53, NULL, 'ATEN€ÇO: O porto est  parcialmente fechado devido … cheia do rio Negro. (Manaus (AM))', 'ALERTA', 3);
INSERT INTO public.posts VALUES (64, 26, NULL, 'Aproveite a PROMO€ÇO em Porto de Galinhas (PE): Pacotes com desconto de 30% para reservas at‚ domingo.', 'PROMO€ÇO', 3);
INSERT INTO public.posts VALUES (67, 23, NULL, 'ATEN€ÇO: Pousadas em Len‡¢is com at‚ 40% de desconto em junho. (Chapada Diamantina (BA))', 'ALERTA', 3);
INSERT INTO public.posts VALUES (70, 16, NULL, 'ATEN€ÇO: Nova sinaliza‡Æo nas trilhas da Chapada garante mais seguran‡a. (Len‡¢is (BA))', 'ALERTA', 5);
INSERT INTO public.posts VALUES (73, 17, NULL, 'Aproveite a PROMO€ÇO em Pipa (RN): Pousadas com di ria reduzida durante o meio da semana.', 'PROMO€ÇO', 4);
INSERT INTO public.posts VALUES (76, 13, NULL, 'ATEN€ÇO: Forte calor nas trilhas do Parque do Lajeado, leve bastante  gua. (Palmas (TO))', 'ALERTA', 4);
INSERT INTO public.posts VALUES (78, 78, NULL, 'Aproveite a PROMO€ÇO em Manaus (AM): O porto est  parcialmente fechado devido … cheia do rio Negro.', 'PROMO€ÇO', 4);
INSERT INTO public.posts VALUES (79, 57, NULL, 'ATEN€ÇO: Trilhas lindas e cachoeiras cristalinas, perfeito para ecoturismo. (Serra do Cip¢ (MG))', 'ALERTA', 4);
INSERT INTO public.posts VALUES (80, 55, NULL, 'ATEN€ÇO: ALERTA: Acesso limitado … Praia do Farol por excesso de turistas. (Arraial do Cabo (RJ))', 'ALERTA', 4);
INSERT INTO public.posts VALUES (83, 6, NULL, 'Aproveite a PROMO€ÇO em Ilhabela (SP): A travessia de balsa est  com longas filas nos fins de semana.', 'PROMO€ÇO', 4);
INSERT INTO public.posts VALUES (84, 18, NULL, 'ALERTA: Acesso … Praia do Farol, em Arraial do Cabo (RJ), est  restrito devido ao grande n£mero de turistas. Planeje com antecedˆncia.', 'ALERTA', 41);
INSERT INTO public.posts VALUES (86, 82, NULL, 'Fui a Piren¢polis (GO) no Festival de Inverno, e os shows e a comida t¡pica me conquistaram. Uma verdadeira festa!', 'NOVIDADES', 41);
INSERT INTO public.posts VALUES (89, 28, NULL, 'PROMO€ÇO no Vale dos Vinhedos (RS): Descontos em pacotes tur¡sticos para quem visita mais de duas vin¡colas no mesmo dia.', 'PROMO€ÇO', 41);
INSERT INTO public.posts VALUES (94, 10, NULL, 'ALERTA: Praia do Gunga est  com acesso limitado devido … mar‚ alta.', 'ALERTA', 41);
INSERT INTO public.posts VALUES (100, 16, NULL, 'ALERTA: Interdi‡Æo tempor ria da estrada para o Pico da Bandeira por deslizamento.', 'ALERTA', 22);
INSERT INTO public.posts VALUES (51, 44, NULL, 'Aproveite a PROMO€ÇO em Arraial do Cabo (RJ): ALERTA: Acesso limitado … Praia do Farol por excesso de turistas.', 'PROMO€ÇO', 2);
INSERT INTO public.posts VALUES (53, 13, NULL, 'ATEN€ÇO: Orla nova com ciclovia e feirinhas noturnas animadas. (JoÆo Pessoa (PB))', 'ALERTA', 2);
INSERT INTO public.posts VALUES (54, 52, NULL, 'ATEN€ÇO: Trilhas lindas e cachoeiras cristalinas, perfeito para ecoturismo. (Serra do Cip¢ (MG))', 'ALERTA', 7);
INSERT INTO public.posts VALUES (55, 1, NULL, 'Aproveite a PROMO€ÇO em JoÆo Pessoa (PB): Orla nova com ciclovia e feirinhas noturnas animadas.', 'PROMO€ÇO', 7);
INSERT INTO public.posts VALUES (58, 23, NULL, 'ATEN€ÇO: Nova sinaliza‡Æo nas trilhas da Chapada garante mais seguran‡a. (Len‡¢is (BA))', 'ALERTA', 13);
INSERT INTO public.posts VALUES (103, 19, NULL, 'ALERTA: Chuva forte e risco de enchente em Manaus. Redobre os cuidados.', 'ALERTA', 7);
INSERT INTO public.posts VALUES (104, 20, NULL, 'Lan‡amento de circuito cultural em Salvador com foco na heran‡a afro-brasileira.', 'NOVIDADES', 7);
INSERT INTO public.posts VALUES (97, 13, NULL, 'ALERTA: Forte temporal previsto para a regiÆo de Paraty neste fim de semana.', 'ALERTA', 7);
INSERT INTO public.posts VALUES (98, 14, NULL, 'Explore as belezas naturais da Chapada das Mesas com nossos novos pacotes.', 'NOVIDADES', 7);
INSERT INTO public.posts VALUES (65, 3, NULL, 'Aproveite a PROMO€ÇO em Ilhabela (SP): A travessia de balsa est  com longas filas nos fins de semana.', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (66, 96, NULL, 'ATEN€ÇO: Tour noturno com degusta‡Æo, vin¡colas encantadoras. (Vale dos Vinhedos (RS))', 'ALERTA', 5);
INSERT INTO public.posts VALUES (74, 58, NULL, 'ATEN€ÇO: Orla nova com ciclovia e feirinhas noturnas animadas. (JoÆo Pessoa (PB))', 'ALERTA', 5);
INSERT INTO public.posts VALUES (75, 82, NULL, 'Aproveite a PROMO€ÇO em Foz do Igua‡u (PR): Desconto especial para quem visita as Cataratas e o Parque das Aves.', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (85, 53, NULL, 'PROMO€ÇO em Pipa (RN): Passeios de barco com desconto para quem compra at‚ sexta-feira!', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (87, 59, NULL, 'ALERTA: O calor est  muito intenso em Palmas (TO). Cuidado com as trilhas no Parque Estadual do Lajeado.', 'ALERTA', 5);
INSERT INTO public.posts VALUES (91, 58, NULL, 'Descobri uma PROMO€ÇO imperd¡vel em Gramado (RS): Hospedagem com 25% de desconto no pr¢ximo feriado!', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (92, 8, NULL, 'Descubra a hist¢rica cidade de Goi s Velho, repleta de charme colonial.', 'NOVIDADES', 5);
INSERT INTO public.posts VALUES (93, 9, NULL, 'Promo‡Æo especial para o JalapÆo! Pacotes com at‚ 30% de desconto.', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (95, 76, NULL, 'Nova rota gastron“mica em Tiradentes promete encantar os visitantes.', 'NOVIDADES', 5);
INSERT INTO public.posts VALUES (96, 12, NULL, 'Descontos imperd¡veis para conhecer os Len‡¢is Maranhenses!', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (101, 17, NULL, 'Nova trilha ecol¢gica aberta no Parque Estadual do Ibitipoca!', 'NOVIDADES', 5);
INSERT INTO public.posts VALUES (106, 22, NULL, 'ALERTA: Temporada de algas atinge praias de Alagoas, com impacto na balneabilidade.', 'ALERTA', 0);
INSERT INTO public.posts VALUES (107, 24, NULL, 'Pacotes com desconto para visitas … Serra do Roncador dispon¡veis!', 'PROMO€ÇO', 0);
INSERT INTO public.posts VALUES (108, 25, NULL, 'ALERTA: Praia de Copacabana est  temporariamente fechada para eventos oficiais.', 'ALERTA', 0);
INSERT INTO public.posts VALUES (109, 27, NULL, 'Super promo‡Æo para excursäes escolares … Floresta Nacional do Tapaj¢s!', 'PROMO€ÇO', 0);
INSERT INTO public.posts VALUES (110, 28, NULL, 'ALERTA: Interdi‡Æo parcial da trilha da Pedra da G vea por manuten‡Æo.', 'ALERTA', 0);
INSERT INTO public.posts VALUES (112, 90, NULL, 'Descontos exclusivos para casais em pousadas da Serra Ga£cha.', 'PROMO€ÇO', 0);
INSERT INTO public.posts VALUES (113, 31, NULL, 'ALERTA: µrea de mergulho em Maragogi fechada temporariamente por preserva‡Æo.', 'ALERTA', 0);
INSERT INTO public.posts VALUES (144, 38, NULL, 'Pacotes com valores reduzidos para conhecer o Vale do Ribeira!', 'PROMO€ÇO', 0);
INSERT INTO public.posts VALUES (145, 24, NULL, 'ALERTA: Interdi‡Æo parcial do Elevador Lacerda em Salvador por manuten‡Æo el‚trica.', 'ALERTA', 0);
INSERT INTO public.posts VALUES (156, 44, NULL, 'ALERTA: Forte correnteza no litoral norte da Bahia. Cuidado com crian‡as e idosos.', 'ALERTA', 0);
INSERT INTO public.posts VALUES (157, 14, NULL, 'Reabertura de centro hist¢rico em SÆo Lu¡s ap¢s restaura‡Æo de casaräes coloniais.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (134, 27, NULL, 'ALERTA: Trecho da Estrada Real interditado por obras de revitaliza‡Æo.', 'ALERTA', 22);
INSERT INTO public.posts VALUES (127, 49, NULL, 'ALERTA: Aviso de ressaca para o litoral sul de SÆo Paulo. Cuidado com o banho de mar.', 'ALERTA', 22);
INSERT INTO public.posts VALUES (159, 76, NULL, 'ALERTA: Alta concentra‡Æo de  guas-vivas em praias do sul de Santa Catarina.', 'ALERTA', 22);
INSERT INTO public.posts VALUES (135, 13, NULL, 'Pacotes com at‚ 45% de desconto para visitar a Serra do Espinha‡o!', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (137, 19, NULL, 'Ofertas imperd¡veis para hospedagem em Fernando de Noronha neste mˆs!', 'PROMO€ÇO', 2);
INSERT INTO public.posts VALUES (140, 61, NULL, 'Descontos de baixa temporada para destinos de ecoturismo em Rond“nia.', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (162, 90, NULL, 'ALERTA: Interrup‡Æo no fornecimento de energia em pontos tur¡sticos de Natal.', 'ALERTA', 9);
INSERT INTO public.posts VALUES (111, 29, NULL, 'Reabertura do Museu do Mar em SÆo Francisco do Sul com exposi‡äes in‚ditas.', 'NOVIDADES', 5);
INSERT INTO public.posts VALUES (146, 57, NULL, 'Pacotes promocionais para o SÆo JoÆo em Caruaru j  estÆo dispon¡veis!', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (161, 55, NULL, 'Descontos em pacotes para conhecer o Delta do Parna¡ba!', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (133, 34, NULL, 'Super desconto para visitas guiadas ao Parque Nacional da Serra da Capivara!', 'PROMO€ÇO', 7);
INSERT INTO public.posts VALUES (131, 55, NULL, 'ALERTA: Interdi‡Æo parcial do bondinho do PÆo de A‡£car por manuten‡Æo preventiva.', 'ALERTA', 10);
INSERT INTO public.posts VALUES (136, 48, NULL, 'ALERTA: Possibilidade de neblina densa na regiÆo da Serra do Mar durante a madrugada.', 'ALERTA', 5);
INSERT INTO public.posts VALUES (138, 4, NULL, 'ALERTA: Trilha da Pedra Redonda fechada temporariamente por risco de deslizamento.', 'ALERTA', 5);
INSERT INTO public.posts VALUES (139, 35, NULL, 'Festival de inverno em Campos do JordÆo come‡a com atra‡äes musicais e gastron“micas.', 'NOVIDADES', 7);
INSERT INTO public.posts VALUES (141, 88, NULL, 'ALERTA: Queimadas afetam visibilidade na Chapada dos GuimarÆes.', 'ALERTA', 5);
INSERT INTO public.posts VALUES (142, 10, NULL, 'Hospedagens em Morro de SÆo Paulo com promo‡äes para grupos!', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (143, 3, NULL, 'ALERTA: Animais silvestres avistados em trilhas do Parque Nacional do Capara¢. Aten‡Æo redobrada.', 'ALERTA', 10);
INSERT INTO public.posts VALUES (148, 41, NULL, 'Inaugura‡Æo de centro de visitantes no Parque Estadual do JalapÆo.', 'NOVIDADES', 5);
INSERT INTO public.posts VALUES (150, 50, NULL, 'ALERTA: Aviso de mar‚ alta em praias do norte do Par . Acesso restrito em alguns pontos.', 'ALERTA', 5);
INSERT INTO public.posts VALUES (152, 9, NULL, 'ALERTA: Mosquitos transmissores de doen‡as tropicais em alta no litoral do Amap .', 'ALERTA', 5);
INSERT INTO public.posts VALUES (158, 31, NULL, 'Pacotes para a Chapada dos Veadeiros com at‚ 35% de desconto.', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (160, 60, NULL, 'Inaugura‡Æo de ciclofaixa tur¡stica em Florian¢polis com vista para o mar.', 'NOVIDADES', 11);
INSERT INTO public.posts VALUES (163, 22, NULL, 'Pacotes com valores especiais para turismo de base comunit ria no Xingu!', 'PROMO€ÇO', 9);
INSERT INTO public.posts VALUES (165, 1, NULL, 'Reabertura de passarela sobre o Rio Tocantins com nova ilumina‡Æo noturna.', 'NOVIDADES', 9);
INSERT INTO public.posts VALUES (166, 40, NULL, 'Hospedagens com at‚ 40% de desconto no Vale do Itaja¡!', 'PROMO€ÇO', 9);
INSERT INTO public.posts VALUES (167, 43, NULL, 'ALERTA: Forte calor no sertÆo nordestino. Hidrata‡Æo e prote‡Æo solar sÆo essenciais.', 'ALERTA', 9);
INSERT INTO public.posts VALUES (151, 21, NULL, 'Pacotes para turismo de aventura na Serra do Mar com valores reduzidos.', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (153, 6, NULL, 'Promo‡äes exclusivas para destinos hist¢ricos no interior paulista.', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (154, 18, NULL, 'ALERTA: Obras na BR-101 provocam lentidÆo no acesso a destinos do litoral sul.', 'ALERTA', 5);
INSERT INTO public.posts VALUES (155, 17, NULL, 'Hospedagens em Cara¡va com pacotes promocionais de fim de semana.', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (114, 32, NULL, 'Festival de m£sica ao ar livre chega ao Vale dos Vinhedos neste mˆs!', 'NOVIDADES', 7);
INSERT INTO public.posts VALUES (115, 33, NULL, 'Pacotes promocionais para o Pantanal durante o mˆs de junho!', 'PROMO€ÇO', 7);
INSERT INTO public.posts VALUES (116, 34, NULL, 'ALERTA: Forte calor atinge o litoral do Esp¡rito Santo. Evite exposi‡Æo prolongada.', 'ALERTA', 7);
INSERT INTO public.posts VALUES (118, 37, NULL, 'ALERTA: Acesso … Cachoeira da Fuma‡a restrito por riscos de queda de rochas.', 'ALERTA', 7);
INSERT INTO public.posts VALUES (117, 78, NULL, 'Oferta imperd¡vel para explorar a Ilha do Cardoso com desconto!', 'PROMO€ÇO', 7);
INSERT INTO public.posts VALUES (119, 39, NULL, 'Hospedagem em Alter do ChÆo com at‚ 35% de desconto por tempo limitado.', 'PROMO€ÇO', 7);
INSERT INTO public.posts VALUES (121, 41, NULL, 'Lan‡amento de roteiro sustent vel na regiÆo do JalapÆo em parceria com comunidades locais.', 'NOVIDADES', 7);
INSERT INTO public.posts VALUES (122, 42, NULL, 'Promo‡Æo especial para roteiros de aventura no Parque do CantÆo!', 'PROMO€ÇO', 7);
INSERT INTO public.posts VALUES (123, 43, NULL, 'ALERTA: Forte correnteza no Rio SÆo Francisco. Cuidado ao praticar esportes aqu ticos.', 'ALERTA', 7);
INSERT INTO public.posts VALUES (124, 45, NULL, 'Hospedagens em pousadas de charme em Paraty com pre‡os especiais!', 'PROMO€ÇO', 7);
INSERT INTO public.posts VALUES (125, 90, NULL, 'ALERTA: N¡veis de polui‡Æo do ar elevados em Belo Horizonte. Evite atividades ao ar livre.', 'ALERTA', 7);
INSERT INTO public.posts VALUES (126, 48, NULL, 'Viagens para Bonito com descontos especiais neste mˆs!', 'PROMO€ÇO', 7);
INSERT INTO public.posts VALUES (128, 51, NULL, 'Descontos de at‚ 50% para passeios na Rota das Cachoeiras em Goi s!', 'PROMO€ÇO', 7);
INSERT INTO public.posts VALUES (129, 88, NULL, 'ALERTA: Fauna silvestre vis¡vel em  reas urbanas de Florian¢polis. Mantenha distƒncia.', 'ALERTA', 7);
INSERT INTO public.posts VALUES (130, 54, NULL, 'Pacotes promocionais para conhecer o Vale Europeu em SC!', 'PROMO€ÇO', 7);
INSERT INTO public.posts VALUES (132, 56, NULL, 'Evento especial de observa‡Æo de aves no Parque Nacional do Viru .', 'NOVIDADES', 7);
INSERT INTO public.posts VALUES (147, 2, NULL, 'ALERTA: Forte vento na orla de Fortaleza. Redobre o cuidado com barracas e guarda-s¢is.', 'ALERTA', 7);
INSERT INTO public.posts VALUES (149, 16, NULL, 'Viagens para Porto Seguro com at‚ 50% OFF para reservas antecipadas!', 'PROMO€ÇO', 7);
INSERT INTO public.posts VALUES (171, 28, NULL, 'Festival ind¡gena em Tocantins celebra cultura com atividades abertas ao p£blico.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (173, 1, '1747784062208-2.jpg', 'Visitei Ouro Preto no último fim de semana e recomendo muito a nova rota noturna com luzes cênicas. Caminhar pelas ladeiras históricas à noite foi mágico!', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (175, 28, '1747784280770-4.jpg', 'Fui na FLIP em Paraty e participei de um sarau literário incrível em um dos casarões restaurados. A cidade respira cultura o tempo todo.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (181, 29, '1747784460436-10.jpg', 'Passei o feriado em São Miguel do Gostoso (RN) e foi incrível! Praia calma, vento constante, ótima pra kitesurf e descanso total.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (182, 5, '1747784488587-11.jpg', 'PROMOÇÃO rolando em Foz do Iguaçu (PR): desconto para quem visitar as Cataratas e o Parque das Aves no mesmo dia. Entrada combinada vale a pena!', 'PROMOÇÃO', 0);
INSERT INTO public.posts VALUES (184, 9, '1747784536155-13.jpg', 'Visitei Arraial do Cabo (RJ) e adorei! ALERTA: Acesso limitado à Praia do Farol por excesso de turistas.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (186, 12, '1747784612805-15.jpg', 'Visitei Palmas (TO) e adorei! Forte calor nas trilhas do Parque do Lajeado, leve bastante água.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (191, 6, '1747784742417-20.jpg', 'Visitei Porto de Galinhas (PE) e adorei! Pacotes com desconto de 30% para reservas até domingo.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (197, 9, '1747784876043-26.jpg', 'Visitei João Pessoa (PB) e adorei! Orla nova com ciclovia e feirinhas noturnas animadas.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (206, 45, '1747785060602-35.jpg', 'Rota dos Cânions é ampliada com novos mirantes e áreas de descanso.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (210, 23, '1747785132953-39.jpg', 'Nova atração turística em São João del Rei celebra a cultura mineira.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (213, 59, '1747785193792-42.jpg', 'Nova feira cultural em Olinda destaca artistas locais e artesanato regional.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (178, 26, '1747784375273-7.jpg', 'Achei ótima a nova ciclovia entre a Praia Mole e Lagoa da Conceição. Pedalei com a família inteira e foi super tranquilo!', 'NOVIDADES', 10);
INSERT INTO public.posts VALUES (201, 8, '1747784956915-30.jpg', 'Visitei Vale dos Vinhedos (RS) e adorei! Tour noturno com degustação, vinícolas encantadoras.', 'NOVIDADES', 4);
INSERT INTO public.posts VALUES (208, 76, '1747785096717-37.jpg', 'Inaugurado o novo teleférico panorâmico no Parque Nacional de Itatiaia.', 'NOVIDADES', 5);
INSERT INTO public.posts VALUES (168, 7, NULL, 'Lan‡amento de aplicativo de turismo para o interior de Minas Gerais.', 'NOVIDADES', 9);
INSERT INTO public.posts VALUES (170, 58, NULL, 'ALERTA: Interdi‡Æo tempor ria no Parque Nacional da Serra da Bocaina por manuten‡Æo.', 'ALERTA', 10);
INSERT INTO public.posts VALUES (174, 43, '1747784241712-3.jpg', 'ALERTA: Trilha principal da Chapada dos Veadeiros está com interdição parcial para manutenção das passarelas ecológicas. Só está liberado o acesso ao Vale da Lua.', 'ALERTA', 10);
INSERT INTO public.posts VALUES (176, 59, '1747784306343-5.jpg', 'Recife está um espetáculo com o Festival Recife na Rua. Tem música, teatro e comidinhas típicas em cada esquina do bairro antigo. Vale muito a visita!', 'NOVIDADES', 10);
INSERT INTO public.posts VALUES (177, 82, '1747784335809-6.jpg', 'ALERTA: Praia da Joaquina em Florianópolis com forte correnteza nesta semana. Evitem nadar em áreas sem sinalização de segurança.', 'ALERTA', 12);
INSERT INTO public.posts VALUES (179, 12, '1747784396041-8.jpg', 'Visitei o Jalapão (TO) semana passada e fiquei impressionado com os fervedouros! A água parece te empurrar pra cima. Experiência surreal!', 'NOVIDADES', 10);
INSERT INTO public.posts VALUES (180, 98, '1747784432090-9.jpg', 'ALERTA: Em Ilhabela (SP), a travessia de balsa está com longas filas nos fins de semana. Recomendo ir bem cedo pra evitar atrasos.', 'ALERTA', 10);
INSERT INTO public.posts VALUES (183, 13, '1747784510157-12.jpg', 'Estive em Vale dos Vinhedos (RS) recentemente. Tour noturno com degustação, vinícolas encantadoras.', 'NOVIDADES', 10);
INSERT INTO public.posts VALUES (185, 39, '1747784583660-14.jpg', 'ATENÇÃO: Desconto especial para quem visita as Cataratas e o Parque das Aves. (Foz do Iguaçu (PR))', 'PROMOÇÃO', 10);
INSERT INTO public.posts VALUES (187, 26, '1747784635425-16.jpg', 'Visitei Pipa (RN) e adorei! Pousadas com diária reduzida durante o meio da semana.', 'NOVIDADES', 10);
INSERT INTO public.posts VALUES (188, 58, '1747784655958-17.jpg', 'Visitei Canela (RS) e adorei! Nova iluminação na Catedral de Pedra está deslumbrante à noite.', 'NOVIDADES', 10);
INSERT INTO public.posts VALUES (189, 37, '1747784678770-18.jpg', 'Visitei Serra do Cipó (MG) e adorei! Trilhas lindas e cachoeiras cristalinas, perfeito para ecoturismo.', 'NOVIDADES', 10);
INSERT INTO public.posts VALUES (190, 56, '1747784722017-19.jpg', 'Visitei Belém (PA) e adorei! Mercado Ver-o-Peso restaurado com nova área gastronômica.', 'NOVIDADES', 10);
INSERT INTO public.posts VALUES (193, 78, '1747784782791-22.jpg', 'Visitei Chapada Diamantina (BA) e adorei! Pousadas em Lençóis com até 40% de desconto em junho.', 'NOVIDADES', 10);
INSERT INTO public.posts VALUES (192, 45, '1747784763273-21.jpg', 'Visitei Brotas (SP) e adorei! Novo circuito de tirolesas foi inaugurado no parque de aventuras.', 'NOVIDADES', 20);
INSERT INTO public.posts VALUES (194, 11, '1747784803684-23.jpg', 'Visitei Arraial do Cabo (RJ) e adorei! ALERTA: Acesso limitado à Praia do Farol por excesso de turistas.', 'NOVIDADES', 10);
INSERT INTO public.posts VALUES (195, 32, '1747784833462-24.jpg', 'Visitei Natal (RN) e adorei! Passeio de buggy pelas dunas está com 20% de desconto.', 'NOVIDADES', 10);
INSERT INTO public.posts VALUES (196, 8, '1747784853927-25.jpg', 'Visitei Manaus (AM) e adorei! O porto está parcialmente fechado devido à cheia do rio Negro.', 'NOVIDADES', 10);
INSERT INTO public.posts VALUES (198, 19, '1747784902324-27.jpg', 'Visitei Palmas (TO) e adorei! Forte calor nas trilhas do Parque do Lajeado, leve bastante água.', 'NOVIDADES', 10);
INSERT INTO public.posts VALUES (199, 80, '1747784921618-28.jpg', 'Visitei Balneário Camboriú (SC) e adorei! PROMOÇÃO em passeios de teleférico na temporada de inverno.', 'NOVIDADES', 10);
INSERT INTO public.posts VALUES (200, 53, '1747784939634-29.jpg', 'Visitei São Miguel do Gostoso (RN) e adorei! Praia calma e vento constante, ótimo para kitesurf.', 'NOVIDADES', 6);
INSERT INTO public.posts VALUES (202, 35, '1747784974954-31.jpg', 'Novo mirante aberto ao público no Parque Nacional de Aparados da Serra.', 'NOVIDADES', 4);
INSERT INTO public.posts VALUES (203, 44, '1747784992565-32.jpg', 'Novo centro de visitantes no Parque Nacional da Serra dos Órgãos oferece experiências imersivas.', 'NOVIDADES', 14);
INSERT INTO public.posts VALUES (205, 29, '1747785039361-34.jpg', 'Novo passeio de barco em Arraial do Cabo oferece experiências ao pôr do sol.', 'NOVIDADES', 14);
INSERT INTO public.posts VALUES (204, 47, '1747785015575-33.jpg', 'Reabertura da Rota do Vinho em São Roque com novos empreendimentos.', 'NOVIDADES', 28);
INSERT INTO public.posts VALUES (211, 53, '1747785155442-40.jpg', 'Abertura de nova trilha acessível para pessoas com deficiência no Parque da Tijuca.', 'NOVIDADES', 14);
INSERT INTO public.posts VALUES (209, 38, '1747785116389-38.jpg', 'Inaugurado o circuito de cicloturismo na Serra da Mantiqueira.', 'NOVIDADES', 14);
INSERT INTO public.posts VALUES (207, 12, '1747785078006-36.jpg', 'Novo roteiro de ecoturismo no Acre oferece experiências únicas na floresta.', 'NOVIDADES', 28);
INSERT INTO public.posts VALUES (212, 8, '1747785174875-41.jpg', 'Reinauguração do Museu da Imigração em São Paulo com novas exposições interativas.', 'NOVIDADES', 14);
INSERT INTO public.posts VALUES (215, 36, '1747785230790-44.jpg', 'Nova atração interativa no Instituto Inhotim atrai visitantes de todas as idades.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (216, 32, '1747785251060-45.jpg', 'Nova linha de trem turístico ligando Curitiba a Morretes é reinaugurada.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (222, 4, NULL, 'Descobri um caf‚ novo em Tiradentes com vista para as montanhas e ambiente colonial. àtima parada depois das trilhas culturais.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (223, 55, NULL, 'ALERTA: O acesso … Cachoeira do Formiga, em JalapÆo, est  restrito durante a semana para recupera‡Æo da trilha. Planejem-se para visitar aos fins de semana.', 'ALERTA', 0);
INSERT INTO public.posts VALUES (226, 23, NULL, 'PROMO€ÇO: Passeios guiados pelo centro hist¢rico de Salvador com 50% de desconto durante o mˆs de junho. Aproveite para conhecer mais da cultura baiana!', 'PROMO€ÇO', 0);
INSERT INTO public.posts VALUES (227, 38, NULL, 'ALERTA: Sobe para n¡vel m ximo o risco de incˆndios na regiÆo do JalapÆo (TO). Evite acampamentos e nÆo fa‡a fogueiras.', 'ALERTA', 0);
INSERT INTO public.posts VALUES (228, 10, NULL, 'Nova passarela panorƒmica inaugurada na Serra do Rio do Rastro! Agora d  pra ter uma vista completa dos cƒnions com seguran‡a.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (230, 61, NULL, 'ALERTA: Temporada de chuvas fortes no litoral do Esp¡rito Santo. Algumas praias estÆo com acesso restrito devido a deslizamentos.', 'ALERTA', 0);
INSERT INTO public.posts VALUES (231, 30, NULL, 'Participei de um festival de culin ria cai‡ara em Ilhabela com pratos t¡picos e m£sica ao vivo. Um ¢timo programa para quem ama gastronomia regional.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (232, 5, NULL, 'PROMO€ÇO relƒmpago para trilha guiada na Serra da Canastra: grupos com at‚ 4 pessoas pagam o pre‡o de 3!', 'PROMO€ÇO', 0);
INSERT INTO public.posts VALUES (233, 75, NULL, 'ALERTA: Observa‡Æo de baleias em Abrolhos suspensa temporariamente por conta do aumento da atividade mar¡tima irregular na  rea.', 'ALERTA', 0);
INSERT INTO public.posts VALUES (234, 18, NULL, 'NOVIDADES em Curitiba: o Jardim Botƒnico agora oferece visitas noturnas com ilumina‡Æo art¡stica e trilhas sensoriais. Uma experiˆncia £nica!', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (235, 52, NULL, 'ALERTA: A travessia do Parque Nacional de Itatiaia est  fechada por tempo indeterminado devido a deslizamentos recentes.', 'ALERTA', 0);
INSERT INTO public.posts VALUES (236, 7, NULL, 'PROMO€ÇO: City tour por Belo Horizonte com degusta‡Æo de queijos e doces mineiros incluso no pacote. Descontos v lidos at‚ o final do mˆs.', 'PROMO€ÇO', 0);
INSERT INTO public.posts VALUES (237, 33, NULL, 'Descobri um novo circuito de bike em Bento Gon‡alves entre vinhedos e paisagens incr¡veis. Ideal para quem curte pedal e enoturismo.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (239, 89, NULL, 'PROMO€ÇO: Hospedagem em hostels de SÆo Paulo com 20% OFF para reservas feitas pelo app at‚ sexta-feira.', 'PROMO€ÇO', 0);
INSERT INTO public.posts VALUES (241, 13, NULL, 'ALERTA: A travessia do Vale do Pati, na Chapada Diamantina, requer guias credenciados obrigatoriamente a partir deste mˆs. Fiscaliza‡Æo aumentada.', 'ALERTA', 0);
INSERT INTO public.posts VALUES (242, 20, '1747786139446-47.jpg', 'PROMOÇÃO em Bonito (MS): agências locais estão oferecendo combos de flutuação + trilha com 25% de desconto durante a semana.', 'PROMOÇÃO', 0);
INSERT INTO public.posts VALUES (243, 64, '1747786166030-47.jpg', 'NOVIDADES em São Luís (MA): reabertura do Teatro Arthur Azevedo com programação cultural gratuita nas noites de sexta-feira.', 'NOVIDADES', 0);
INSERT INTO public.posts VALUES (1, 38, NULL, 'ALERTA: Algumas lagoas dos Len‡¢is Maranhenses estÆo com menor volume de  gua este mˆs por causa das chuvas irregulares. Melhor confirmar com os guias antes de agendar passeio.', 'ALERTA', 1);
INSERT INTO public.posts VALUES (52, 42, NULL, 'ATEN€ÇO: Orla nova com ciclovia e feirinhas noturnas animadas. (JoÆo Pessoa (PB))', 'ALERTA', 6);
INSERT INTO public.posts VALUES (120, 40, NULL, 'ALERTA: Interdi‡Æo da travessia para Ilha de Itamarac  por manuten‡Æo na balsa.', 'ALERTA', 22);
INSERT INTO public.posts VALUES (225, 67, NULL, 'NOVIDADES em Bel‚m! O Mercado Ver-o-Peso agora conta com um espa‡o gourmet reformado, reunindo chefs locais e ingredientes t¡picos amaz“nicos.', 'NOVIDADES', 5);
INSERT INTO public.posts VALUES (250, 14, '1747786336568-55.jpg', 'ALERTA: Forte ressaca atinge a orla do Rio de Janeiro, especialmente entre o Leme e o Leblon. Evitem a faixa de areia e áreas de pedras.', 'ALERTA', 2);
INSERT INTO public.posts VALUES (240, 41, NULL, 'A rota do canga‡o em Piranhas (AL) est  com novas sinaliza‡äes e pain‚is interativos contando a hist¢ria de LampiÆo e Maria Bonita. Muito bem feita!', 'NOVIDADES', 2);
INSERT INTO public.posts VALUES (220, 36, NULL, 'Participei da Festa do Divino em Alcƒntara (MA) neste mˆs e foi emocionante. As tradi‡äes locais seguem vivas e encantadoras.', 'NOVIDADES', 2);
INSERT INTO public.posts VALUES (224, 12, NULL, 'PROMO€ÇO especial em Jericoacoara: pousadas com di ria gr tis na terceira noite durante a baixa temporada. Uma ¢tima oportunidade para relaxar!', 'PROMO€ÇO', 2);
INSERT INTO public.posts VALUES (238, 26, NULL, 'ALERTA: Forte neblina nas estradas da Serra Ga£cha. Redobrem a aten‡Æo e evitem dirigir … noite neste per¡odo.', 'ALERTA', 5);
INSERT INTO public.posts VALUES (218, 17, NULL, 'NOVIDADE em Foz do Igua‡u! O novo centro de visitantes do Parque Nacional est  incr¡vel, com experiˆncias imersivas e interativas sobre a fauna local.', 'NOVIDADES', 5);
INSERT INTO public.posts VALUES (164, 30, NULL, 'ALERTA: Restri‡Æo de visitas em  reas ind¡genas por motivos sanit rios.', 'ALERTA', 18);
INSERT INTO public.posts VALUES (169, 5, NULL, 'Promo‡Æo para pacotes de ecoturismo com foco em observa‡Æo de aves.', 'PROMO€ÇO', 31);
INSERT INTO public.posts VALUES (214, 20, '1747785211406-43.jpg', 'Nova rota de turismo religioso no interior do Paraná é lançada.', 'NOVIDADES', 5);
INSERT INTO public.posts VALUES (217, 26, '1747785276755-46.jpg', 'Nova exposição sobre biodiversidade marinha no Museu Nacional da UFRJ.', 'NOVIDADES', 5);
INSERT INTO public.posts VALUES (221, 91, NULL, 'PROMO€ÇO: Pacotes com 30% de desconto para passeios de barco pelo Delta do Parna¡ba at‚ o final deste mˆs. Imperd¡vel!', 'PROMO€ÇO', 7);
INSERT INTO public.posts VALUES (229, 47, NULL, 'PROMO€ÇO: Resort em Caldas Novas com tarifa reduzida para fam¡lias e cortesia para crian‡as menores de 10 anos durante os fins de semana.', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (244, 11, '1747786198365-49.jpg', 'ALERTA: Praia do Rosa (SC) com alta concentração de águas-vivas nesta temporada. Atenção redobrada para quem for entrar no mar.', 'ALERTA', 5);
INSERT INTO public.posts VALUES (245, 58, '1747786226526-50.jpg', 'PROMOÇÃO: Passeios de buggy nas dunas de Natal com valores promocionais para grupos a partir de 3 pessoas. Ideal para famílias!', 'PROMOÇÃO', 5);
INSERT INTO public.posts VALUES (246, 39, '1747786250746-51.jpg', 'NOVIDADE em Manaus: novo mirante do Encontro das Águas oferece vista privilegiada e acesso facilitado com estacionamento.', 'NOVIDADES', 5);
INSERT INTO public.posts VALUES (247, 6, '1747786269889-52.jpg', 'ALERTA: Interdição temporária do acesso ao Morro do Pai Inácio, na Chapada Diamantina, para manutenção nas escadarias e corrimãos.', 'ALERTA', 5);
INSERT INTO public.posts VALUES (248, 95, '1747786292444-53.jpg', 'PROMOÇÃO para amantes do frio: pousadas na Serra Catarinense com café colonial incluso e descontos de até 40% na baixa temporada.', 'PROMOÇÃO', 5);
INSERT INTO public.posts VALUES (249, 72, '1747786315640-54.jpg', 'NOVIDADES em Brasília: o Eixo Cultural Ibero-americano foi inaugurado com exposições permanentes sobre arte e arquitetura latino-americana.'', ''NOVIDADES', 'NOVIDADES', 5);
INSERT INTO public.posts VALUES (172, 34, '1747784020477-1.jpg', 'Acabei de voltar de Bonito (MS) e fiquei encantado com o novo Mirante da Nascente Azul! A passarela suspensa é incrível e a vista é de tirar o fôlego.', 'NOVIDADES', 10);
INSERT INTO public.posts VALUES (219, 22, NULL, 'ALERTA: Estrada de acesso ao Pico da Bandeira est  com trechos escorregadios devido …s chuvas intensas. Redobrem a aten‡Æo ao subir.', 'ALERTA', 5);
INSERT INTO public.posts VALUES (7, 76, NULL, 'PROMO€ÇO na Chapada Diamantina (BA): pousadas em Len‡¢is com at‚ 40% de desconto no mˆs de junho. Aproveitem antes da alta temporada!', 'PROMO€ÇO', 5);
INSERT INTO public.posts VALUES (62, 17, NULL, 'ATEN€ÇO: Pacotes com desconto de 30% para reservas at‚ domingo. (Porto de Galinhas (PE))', 'ALERTA', 5);


--
-- TOC entry 4836 (class 0 OID 17713)
-- Dependencies: 223
-- Data for Name: regions; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.regions VALUES (1, 'Sudeste', 'Sao Paulo (capital)', 'SP', 'Centro cultural e economico do pais, com atracoes como o MASP, Mercado Municipal, Parque Ibirapuera e a Avenida Paulista.', 'https://www.tripadvisor.com.br/Attractions-g303631-Activities-Sao_Paulo_State_of_Sao_Paulo.html', '1748352576513-SÃ£oPaulo(capital).jpg');
INSERT INTO public.regions VALUES (2, 'Sudeste', 'Ilhabela', 'SP', 'Ilha paradisiaca com praias intocadas, trilhas na Mata Atlantica e otimos locais para mergulho e esportes aquaticos.', NULL, '1748352608572-Ilhabela.jpg');
INSERT INTO public.regions VALUES (3, 'Sudeste', 'Campos do Jordao', 'SP', 'Conhecida como a Suica Brasileira, oferece clima ameno, arquitetura europeia e o famoso Festival de Inverno.', NULL, '1748352635888-CamposDeJordÃ£o.jpg');
INSERT INTO public.regions VALUES (4, 'Sudeste', 'Holambra', 'SP', 'Capital das Flores, com campos floridos e o maior moinho da America Latina, refletindo a heranca holandesa.', NULL, '1748352657756-Holambra.jpg');
INSERT INTO public.regions VALUES (5, 'Sudeste', 'Rio de Janeiro (capital)', 'RJ', 'Famosa mundialmente por suas belezas naturais, como o Pao de Acucar, Cristo Redentor e praias como Copacabana e Ipanema.', 'https://www.tripadvisor.com.br/Attractions-g303506-Activities-Rio_de_Janeiro_State_of_Rio_de_Janeiro.htm', '1748352701266-RioDeJaneiro(capital).jpg');
INSERT INTO public.regions VALUES (6, 'Sudeste', 'Paraty', 'RJ', 'Cidade colonial charmosa, reconhecida como Patrimonio Mundial da UNESCO, com centro historico preservado e belas praias.', NULL, '1748352728169-Paraty.jpg');
INSERT INTO public.regions VALUES (7, 'Sudeste', 'Regiao dos Lagos', 'RJ', 'Inclui destinos como Buzios, Arraial do Cabo e Cabo Frio, conhecidos por suas praias paradisiacas e excelente infraestrutura turistica.', NULL, '1748352755677-RegiÃ£oDosLagos.jpg');
INSERT INTO public.regions VALUES (8, 'Sudeste', 'Belo Horizonte (capital)', 'MG', 'Famosa por sua hospitalidade, culinaria e vida noturna animada. Ponto de partida para cidades historicas como Ouro Preto e Mariana.', 'https://www.tripadvisor.com.br/Attractions-g303370-Activities-State_of_Minas_Gerais.html', '1748352801815-BeloHorizonte.jpg');
INSERT INTO public.regions VALUES (9, 'Sudeste', 'Ouro Preto', 'MG', 'Cidade colonial que preserva a rica heranca do periodo do ouro no Brasil, com igrejas barrocas e ruas de paralelepipedos.', NULL, '1748352834816-OuroPreto.jpg');
INSERT INTO public.regions VALUES (10, 'Sudeste', 'Parques Nacionais', 'MG', 'Como o da Serra da Canastra, onde nasce o Rio Sao Francisco, e o da Serra do Cipo, conhecidos por suas belezas naturais e biodiversidade.', NULL, '1748352884576-SerraDaCanastra.jpg');
INSERT INTO public.regions VALUES (11, 'Sudeste', 'Vitoria (capital)', 'ES', 'Cidade litoranea com praias tranquilas, como a Ilha do Boi e a Ilha do Frade, alem de uma rica gastronomia baseada em frutos do mar.', 'https://www.tripadvisor.com.br/Attractions-g303308-Activities-State_of_Espirito_Santo.html', '1748352932765-VitÃ³ria(capital).jpg');
INSERT INTO public.regions VALUES (12, 'Sul', 'Curitiba (capital)', 'PR', 'Reconhecida por seu planejamento urbano, oferece atrações como o Jardim Botânico, o Museu Oscar Niemeyer e o Parque Barigui.', 'https://www.tripadvisor.com.br/Attractions-g303435-Activities-State_of_Parana.html', '1748353071149-Curitiba(capital).jpg');
INSERT INTO public.regions VALUES (13, 'Sul', 'Foz do Iguaçu', 'PR', 'Lar das impressionantes Cataratas do Iguaçu, uma das Sete Maravilhas Naturais do Mundo, e da Represa de Itaipu. A cidade também é ponto de acesso a Ciudad del Este (Paraguai) e Puerto Iguazú (Argentina).', NULL, '1748353104316-FozDoIguaÃ§u.jpg');
INSERT INTO public.regions VALUES (14, 'Sul', 'Serra do Mar', 'PR', 'Ideal para ecoturismo, com trilhas que levam a vilarejos históricos como Morretes e Antonina, onde se pode degustar o tradicional barreado.', NULL, '1748353124848-SerraDoMar.jpg');
INSERT INTO public.regions VALUES (15, 'Sul', 'Florianópolis (capital)', 'SC', 'A capital, conhecida por suas belas praias, como Joaquina e Campeche, além da Lagoa da Conceição e da Ponte Hercílio Luz.', 'https://www.tripadvisor.com.br/Attractions-g303570-Activities-State_of_Santa_Catarina.html', '1748353167619-FlorianÃ³polis(capital).jpg');
INSERT INTO public.regions VALUES (16, 'Sul', 'Balneário Camboriú', 'SC', 'Famosa por sua vida noturna vibrante e pelo Parque Unipraias, que oferece vistas panorâmicas e atividades de ecoturismo.', NULL, '1748353190148-BalneÃ¡rioCamboriÃº.jpg');
INSERT INTO public.regions VALUES (17, 'Sul', 'São Joaquim', 'SC', 'Destino ideal para quem deseja vivenciar o inverno brasileiro, com possibilidade de neve e temperaturas negativas.', NULL, '1748353207675-SÃ£oJoaquim.jpg');
INSERT INTO public.regions VALUES (18, 'Sul', 'Porto Alegre (capital)', 'RS', 'Cidade com forte influência cultural, sendo um dos principais centros urbanos do sul do Brasil. A cidade está situada às margens do Lago Guaíba e é conhecida por sua diversidade cultural, gastronomia, e seus famosos centros históricos e espaços ao ar livre.', 'https://www.tripadvisor.com.br/Attractions-g303530-Activities-State_of_Rio_Grande_do_Sul.html', '1748353281920-PortoAlegre(capital).jpg');
INSERT INTO public.regions VALUES (19, 'Sul', 'Vale dos Vinhedos', 'RS', 'Região produtora de vinhos, onde é possível visitar vinícolas e participar de degustações.', NULL, '1748353314814-ValedosVinhedos.jpg');
INSERT INTO public.regions VALUES (20, 'Sul', 'Cânion Itaimbezinho', 'RS', 'Localizado no Parque Nacional de Aparados da Serra, oferece trilhas e vistas deslumbrantes.', NULL, '1748353339153-CÃ¢nionItaimbezinho.jpg');
INSERT INTO public.regions VALUES (21, 'Norte', 'Manaus (capital)', 'AM', 'Porta de entrada para a Amazônia, a capital amazonense oferece atrações como o Teatro Amazonas, o Mercado Municipal e o Encontro das Águas, onde os rios Negro e Solimões correm lado a lado sem se misturar.', 'https://www.tripadvisor.com.br/Attractions-g303226-Activities-State_of_Amazonas.html', '1748353489684-Manaus(capital).jpg');
INSERT INTO public.regions VALUES (22, 'Norte', 'Floresta Amazônica', 'AM', 'Possibilidade de realizar passeios de barco, caminhadas na selva e observação de fauna e flora únicas.', NULL, '1748353513698-FlorestaAmazÃ´nica.jpg');
INSERT INTO public.regions VALUES (23, 'Norte', 'Belém (capital)', 'PA', 'Capital paraense, famosa por sua culinária típica, como o tacacá e o açaí, além de pontos turísticos como o Mercado Ver-o-Peso e o Museu Paraense Emílio Goeldi.', 'https://www.tripadvisor.com.br/Attractions-g303402-Activities-State_of_Para.html', '1748353548551-BelÃ©m(capital).jpg');
INSERT INTO public.regions VALUES (24, 'Norte', 'Alter do Chão', 'PA', 'Conhecida como o Caribe Amazônico, oferece praias de água doce com areia branca e águas cristalinas, sendo considerada uma das praias mais bonitas do Brasil.', NULL, '1748353775086-AlterdoChÃ£o.jpg');
INSERT INTO public.regions VALUES (25, 'Norte', 'Rio Branco (capital)', 'AC', 'Maior cidade do estado e está localizada na região norte do Brasil, próxima à fronteira com o Peru e a Bolívia. A cidade tem uma mistura de história, cultura e natureza, sendo a principal porta de entrada para quem deseja explorar a Floresta Amazônica e as belezas naturais do Acre.', 'https://www.tripadvisor.com.br/Attractions-g303199-Activities-State_of_Acre.html', '1748353815386-RioBranco(capital).jpg');
INSERT INTO public.regions VALUES (26, 'Norte', 'Geoglifos do Acre', 'AC', 'Estruturas geométricas misteriosas esculpidas no solo, com mais de 3.000 anos de idade, que despertam interesse arqueológico.', NULL, '1748353844882-GeoglifosdoAcre.jpg');
INSERT INTO public.regions VALUES (27, 'Norte', 'Boa Vista (capital)', 'RR', 'Única capital brasileira situada completamente ao norte da linha do Equador e é conhecida por sua proximidade com a natureza e por ser uma cidade com forte presença de influências indígenas e amazônicas.', 'https://www.tripadvisor.com.br/Attractions-g30362-Activities-State_of_Roraima.html', '1748353889845-BoaVista(capital).jpg');
INSERT INTO public.regions VALUES (28, 'Norte', 'Monte Roraima', 'RR', 'Uma das formações geológicas mais antigas do planeta, oferece trilhas desafiadoras e vistas deslumbrantes, sendo inspiração para obras como O Mundo Perdido de Arthur Conan Doyle.', NULL, '1748353912311-MonteRoraima.jpg');
INSERT INTO public.regions VALUES (29, 'Norte', 'Palmas (capital)', 'TO', 'Cidade planejada, inaugurada em 1989, e está situada às margens do Lago de Palmas, em uma região de grande beleza natural.', 'https://www.tripadvisor.com.br/Attractions-g303645-Activities-State_of_Tocantins.html', '1748353947029-Palmas(capital).jpg');
INSERT INTO public.regions VALUES (30, 'Norte', 'Jalapão', 'TO', 'Conhecido por suas dunas douradas, fervedouros (poços de água que borbulham devido à pressão da água subterrânea) e cachoeiras, é um destino ideal para quem busca aventura e contato com a natureza.', NULL, '1748353973263-JalapÃ£o.jpg');
INSERT INTO public.regions VALUES (31, 'Nordeste', 'Salvador (capital)', 'BA', 'Capital cultural do Brasil, famosa pelo Pelourinho, igrejas históricas e o Elevador Lacerda.', 'https://www.tripadvisor.com.br/Attractions-g303251-Activities-State_of_Bahia.html', '1748354221683-Salvador(capital).jpg');
INSERT INTO public.regions VALUES (32, 'Nordeste', 'Chapada Diamantina', 'BA', 'Parque nacional com cachoeiras, cavernas e trilhas, ideal para os amantes de ecoturismo.', NULL, '1748354244297-ChapadaDiamantina.jpg');
INSERT INTO public.regions VALUES (33, 'Nordeste', 'Ilha de Itaparica', 'BA', 'Destino tranquilo com praias de águas calmas, perfeito para relaxar.', NULL, '1748354270673-IlhaDeItaparica.jpg');
INSERT INTO public.regions VALUES (34, 'Nordeste', 'Recife (capital)', 'PE', 'Cidade com forte influência cultural, sendo um dos principais centros urbanos do sul do Brasil. A cidade está situada às margens do Lago Guaíba e é conhecida por sua divePEidade cultural, gastronomia, e seus famosos centros históricos e espaços ao ar livre.', 'https://www.tripadvisor.com.br/Attractions-g303459-Activities-State_of_Pernambuco.html', '1748354314693-Recife(capital).jpg');
INSERT INTO public.regions VALUES (35, 'Nordeste', 'Olinda', 'PE', 'Cidade histórica com igrejas barrocas e ladeiras coloridas.', NULL, '1748354363802-Olinda.jpg');
INSERT INTO public.regions VALUES (36, 'Nordeste', 'Porto de Galinhas', 'PE', 'Praias paradisíacas com piscinas naturais e águas cristalinas.', NULL, '1748354386706-PortoDeGalinhas.jpg');
INSERT INTO public.regions VALUES (37, 'Nordeste', 'Fortaleza (capital)', 'CE', 'Capital vibrante com praias urbanas como Praia do Futuro e atrações culturais como o Mercado Central.', 'https://www.tripadvisor.com.br/Attractions-g303284-Activities-State_of_Ceara.html', '1748354421294-Fortaleza(capital).jpg');
INSERT INTO public.regions VALUES (38, 'Nordeste', 'Jericoacoara', 'CE', 'Vila charmosa com dunas, lagoas e pôr do sol deslumbrante.', NULL, '1748354443236-Jericoacoara.jpg');
INSERT INTO public.regions VALUES (39, 'Nordeste', 'Canoa Quebrada', 'CE', 'Praia famosa por suas falésias vermelhas e vida noturna animada.', NULL, '1748354467668-CanoaQuebrada.jpg');
INSERT INTO public.regions VALUES (40, 'Nordeste', 'São Luís (capital)', 'MA', 'Centro histórico com casarões coloniais e manifestações culturais como o Bumba Meu Boi.', 'https://www.tripadvisor.com.br/Attractions-g303325-Activities-State_of_Maranhao.html', '1748354503952-SÃ£oLuÃ­s(capital).jpg');
INSERT INTO public.regions VALUES (41, 'Nordeste', 'Lençóis Maranhenses', 'MA', 'Parque nacional com dunas de areia branca e lagoas de água doce, criando paisagens únicas.', NULL, '1748354536584-LenÃ§Ã³isMaranhenses.jpg');
INSERT INTO public.regions VALUES (42, 'Nordeste', 'Maceió (capital)', 'AL', 'Praias urbanas como Pajuçara e Ponta Verde, além de lagoas e piscinas naturais.', 'https://www.tripadvisor.com.br/Attractions-g303208-Activities-State_of_Alagoas.html', '1748354572548-MaceiÃ³(capital).jpg');
INSERT INTO public.regions VALUES (43, 'Nordeste', 'Maragogi', 'AL', 'Conhecida como o Caribe Brasileiro, com suas águas transparentes e recifes de corais.', NULL, '1748354599452-Maragogi.jpg');
INSERT INTO public.regions VALUES (44, 'Nordeste', 'João Pessoa (capital)', 'PB', 'Capital tranquila com praias como Tambaú e Cabo Branco, além do Centro Histórico bem preservado.', 'https://www.tripadvisor.com.br/Attractions-g303422-Activities-State_of_Paraiba.html', '1748354637655-JoÃ£oPessoa(capital).jpg');
INSERT INTO public.regions VALUES (45, 'Nordeste', 'Praia de Coqueirinho', 'PB', 'Praia paradisíaca com falésias e águas calmas.', NULL, '1748354663484-PraiaDeCoqueirinho.jpg');
INSERT INTO public.regions VALUES (46, 'Nordeste', 'Teresina (capital)', 'PI', 'Capital com rica culinária e proximidade com o Parque Nacional de Sete Cidades, conhecido por suas formações rochosas.', 'https://www.tripadvisor.com.br/Attractions-g303462-Activities-State_of_Piaui.html', '1748354738836-Teresina(capital).jpg');
INSERT INTO public.regions VALUES (47, 'Nordeste', 'Delta do Parnaíba', 'PI', 'Único delta em mar aberto das Américas, com ilhas e igarapés.', NULL, '1748354759784-DeltaDoParnaÃ­ba.jpg');
INSERT INTO public.regions VALUES (48, 'Nordeste', 'Natal (capital)', 'RN', 'Capital com praias urbanas como Ponta Negra e o famoso Forte dos Reis Magos.', 'https://www.tripadvisor.com.br/Attractions-g303510-Activities-State_of_Rio_Grande_do_Norte.html', '1748354814632-Natal(capital).jpg');
INSERT INTO public.regions VALUES (49, 'Nordeste', 'Genipabu', 'RN', 'Conhecida por suas dunas móveis e passeios de buggy.', NULL, '1748354833391-Genipabu.jpg');
INSERT INTO public.regions VALUES (50, 'Nordeste', 'Aracaju (capital)', 'SE', 'Capital com praias como Atalaia e Orla de Atalaia, além de museus e mercados artesanais.', 'https://www.tripadvisor.com.br/Attractions-g303637-Activities-State_of_Sergipe.html', '1748354873633-Aracaju(capital).jpg');
INSERT INTO public.regions VALUES (51, 'Nordeste', 'Cânion do Xingó', 'SE', 'Formações rochosas impressionantes no Rio São Francisco.', NULL, '1748354901974-CÃ¢nionDoXingÃ³.jpg');
INSERT INTO public.regions VALUES (52, 'Centro-Oeste', 'Goiania (capital)', 'GO', 'Cidade moderna, planejada, e um dos principais centros urbanos da regiao Centro-Oeste do Brasil.', 'https://www.tripadvisor.com.br/Attractions-g303323-Activities-State_of_Goias.html', '1748355116965-GoiÃ¢nia(capital).jpg');
INSERT INTO public.regions VALUES (53, 'Centro-Oeste', 'Chapada dos Veadeiros', 'GO', 'Localizada no estado de Goias, a Chapada dos Veadeiros e um parque nacional famoso por suas cachoeiras, trilhas e formacoes rochosas. Destinos como a Cachoeira de Santa Barbara e a Cachoeira do Segredo sao altamente recomendados para os amantes de natureza.', NULL, '1748355142603-ChapadaDosVeadeiros.jpg');
INSERT INTO public.regions VALUES (54, 'Centro-Oeste', 'Rio Quente', 'GO', 'Situado no estado de Goias, Rio Quente e famoso por suas aguas termais, com temperaturas que chegam a 38°C. O Hot Park e um dos maiores parques aquaticos da regiao e oferece diversas atracoes para todas as idades.', NULL, '1748355165237-RioQuente.jpg');
INSERT INTO public.regions VALUES (55, 'Centro-Oeste', 'Cuiaba (capital)', 'MT', 'Conhecida como a porta de entrada para o Pantanal e tambem pela sua localizacao geografica unica: e considerada o centro da America do Sul.', 'https://www.tripadvisor.com.br/Attractions-g303346-Activities-State_of_Mato_Grosso.html', '1748355199353-CuiabÃ¡(capital).jpg');
INSERT INTO public.regions VALUES (56, 'Centro-Oeste', 'Chapada dos Guimaraes', 'MT', 'Proxima a Cuiaba, a Chapada dos Guimaraes e conhecida por suas cachoeiras, cavernas e mirantes. O Parque Nacional da Chapada dos Guimaraes e um local ideal para trilhas e observacao da fauna local.', NULL, '1748355222033-ChapadaDosGuimarÃ£es.jpg');
INSERT INTO public.regions VALUES (57, 'Centro-Oeste', 'Campo Grande (capital)', 'MS', 'Conhecida como a Cidade Morena devido a coloracao avermelhada de seu solo, Campo Grande e um importante centro urbano e economico do Centro-Oeste do Brasil. A cidade e uma mistura de culturas indigenas, sul-mato-grossense, paraguaia e pantaneira.', 'https://www.tripadvisor.com.br/Attractions-g303368-Activities-State_of_Mato_Grosso_do_Sul.html', '1748355264898-CampoGrande(capital).jpg');
INSERT INTO public.regions VALUES (58, 'Centro-Oeste', 'Bonito', 'MS', 'Famosa por suas aguas, Bonito oferece atividades como flutuacao em rios, visita a cavernas e cachoeiras. A Gruta da Lagoa Azul e o Abismo Anhumas sao atracoes renomadas.', NULL, '1748355287639-Bonito.jpg');
INSERT INTO public.regions VALUES (59, 'Centro-Oeste', 'Brasilia (distrito federal)', 'DF', 'A capital federal do Brasil e reconhecida por sua arquitetura modernista, projetada por Oscar Niemeyer. Pontos turisticos como a Catedral Metropolitana, o Congresso Nacional e o Palacio da Alvorada sao imperdiveis.', 'https://www.tripadvisor.com.br/Attractions-g303322-Activities-Brasilia_Federal_District.html', '1748355328218-BrasÃ­lia(distrito federal).jpg');


--
-- TOC entry 4838 (class 0 OID 17719)
-- Dependencies: 225
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.users VALUES (1, 'Amanda Gomes Mechi', 'amanda.mechi@gmail.com', 'Campinas', 'SP', '123456@', '1747780715193-Amanda.jpg');
INSERT INTO public.users VALUES (2, 'Andre Lucca Santos', 'andre.l.santos45@gmail.com', 'Campinas', 'SP', '234561$', '1747780753314-AndrÃ©.jpg');
INSERT INTO public.users VALUES (3, 'Flavia Regina Alexandre Mendes', 'flavia.r.mendes@gmail.com', 'Campinas', 'SP', '345612%', '1747780784485-Flavia.jpg');
INSERT INTO public.users VALUES (4, 'João Vitor Porto Sales', 'joao.v.sales15@gmail.com', 'Campinas', 'SP', '678345#', '1747780835677-JoÃ£o Vitor.jpg');
INSERT INTO public.users VALUES (5, 'Giovanna Caron de Barros', 'giovanna.c.barros@gmail.com', 'Valinhos', 'SP', '456123@', '1747780866879-Giovanna.jpg');
INSERT INTO public.users VALUES (6, 'Isabella Borin de Moraes Rosa', 'isabella.b.rosa6@gmail.com', 'Valinhos', 'SP', '567234%', '1747780956484-Isabella.jpg');
INSERT INTO public.users VALUES (7, 'Laura Ferreira Violla', 'laura.violla@gmail.com', 'Valinhos', 'SP', '789456#', '1747780983062-Laura.jpg');
INSERT INTO public.users VALUES (8, 'Carlos Silva', 'carlo.silva435@gmail.com', 'São Paulo', 'SP', '512735$', '1747781046463-Carlos Silva.jpg');
INSERT INTO public.users VALUES (9, 'Vinícius Andrade', 'vinicius.andrade101@gmail.com', 'Campinas', 'SP', '180577%', '1747781075911-VinÃ­cius Andrade.jpg');
INSERT INTO public.users VALUES (10, 'Matheus Rocha', 'matheus.rocha217@gmail.com', 'São Bernardo do Campo', 'SP', '698730&', '1747781106543-Matheus Rocha.jpg');
INSERT INTO public.users VALUES (11, 'Ana Oliveira', 'ana.oliveira2@gmail.com', 'Rio de Janeiro', 'RJ', '184767#', '1747781549319-Ana Oliveira.jpg');
INSERT INTO public.users VALUES (12, 'Marcos Souza', 'marcos.souza3@gmail.com', 'Belo Horizonte', 'MG', '941291#', '1747781584713-Marcos Souza.jpg');
INSERT INTO public.users VALUES (13, 'Juliana Lima', 'juliana.lima4@gmail.com', 'Curitiba', 'PR', '408734#', '1747781620286-Juliana Lima.jpg');
INSERT INTO public.users VALUES (14, 'Bruna Costa', 'bruna.costa205@gmail.com', 'Maringá', 'PR', '370614#', '1747781649023-Bruna Costa.jpg');
INSERT INTO public.users VALUES (15, 'André Souza', 'andre.souza206@gmail.com', 'Londrina', 'PR', '939015#', '1747781677546-AndrÃ© Souza.jpg');
INSERT INTO public.users VALUES (16, 'Rafael Costa', 'rafael.costa5@gmail.com', 'Porto Alegre', 'RS', '089987@', '1747781722599-Rafael Costa.jpg');
INSERT INTO public.users VALUES (17, 'Tatiane Silva', 'tatiane.silva207@gmail.com', 'Caxias do Sul', 'RS', '905764@', '1747781753726-Tatiane Silva.jpg');
INSERT INTO public.users VALUES (18, 'Fernanda Rocha', 'fernanda.rocha6@gmail.com', 'Recife', 'PE', '555133@', '1747781792439-Fernanda Rocha.jpg');
INSERT INTO public.users VALUES (19, 'Lucas Martins', 'lucas.martins7@gmail.com', 'Salvador', 'BA', '576843$', '1747781824271-Lucas Martins.jpg');
INSERT INTO public.users VALUES (20, 'Camila Ribeiro', 'camila.ribeiro8@gmail.com', 'Brasília', 'BA', '216441$', '1747781863584-Camila Ribeiro.jpg');
INSERT INTO public.users VALUES (21, 'Pedro Almeida', 'pedro.almeida9@gmail.com', 'Fortaleza', 'CE', '418251$', '1747781896396-Pedro Almeida.jpg');
INSERT INTO public.users VALUES (22, 'Aline Mendes', 'aline.mendes10@gmail.com', 'Manaus', 'AM', '442272#', '1747781929384-Aline Mendes.jpg');
INSERT INTO public.users VALUES (23, 'Patrícia Moraes', 'patricia.moraes102@gmail.com', 'Maceió', 'AL', '900128#', '1747781960320-PatrÃ­cia Moraes.jpg');
INSERT INTO public.users VALUES (24, 'Bruno Teixeira', 'bruno.teixeira103@gmail.com', 'João Pessoa', 'PB', '769424#', '1747781998812-Bruno Teixeira.jpg');
INSERT INTO public.users VALUES (25, 'Larissa Pires', 'larissa.pires104@gmail.com', 'São Luís', 'MA', '638515@', '1747782043429-Larissa Pires.jpg');
INSERT INTO public.users VALUES (26, 'Tiago Nunes', 'tiago.nunes105@gmail.com', 'Goiânia', 'GO', '378299@', '1747782073892-Tiago Nunes.jpg');
INSERT INTO public.users VALUES (27, 'Amanda Barreto', 'amanda.barreto201@gmail.com', 'Blumenau', 'SC', '717463@', '1747782105733-Amanda Barreto.jpg');
INSERT INTO public.users VALUES (28, 'João Vitor Sales', 'joao.sales202@gmail.com', 'Cuiabá', 'MT', '5049215%', '1747782134971-JoÃ£o Vitor Sales.jpg');
INSERT INTO public.users VALUES (29, 'Luana Neves', 'luana.neves203@gmail.com', 'Palmas', 'TO', '137840%', '1747782164469-Luana Neves.jpg');
INSERT INTO public.users VALUES (30, 'Letícia Gomes', 'leticia.gomes209@gmail.com', 'Santarém', 'PA', '304881$', '1747782196217-LetÃ­cia Gomes.jpg');
INSERT INTO public.users VALUES (31, 'Rayssa Lopes', 'rayssa.lopes213@gmail.com', 'Aracaju', 'SE', '181202#', '1747782227367-Rayssa Lopes.jpg');
INSERT INTO public.users VALUES (32, 'Caio Antunes', 'caio.antunes214@gmail.com', 'Macapá', 'AP', '641973#', '1747782264548-Caio Antunes.jpg');
INSERT INTO public.users VALUES (33, 'Yasmin Duarte', 'yasmin.duarte218@gmail.com', 'Teresina', 'PI', '390458#', '1747782300186-Yasmin Duarte.jpg');
INSERT INTO public.users VALUES (34, 'Douglas Fernandes', 'douglas.fernandes208@gmail.com', 'RibeirÆo Preto', 'SP', '425876', NULL);
INSERT INTO public.users VALUES (35, 'Gabriel Monteiro', 'gabriel.monteiro210@gmail.com', 'SÆo Jos‚ dos Campos', 'SP', '717008', NULL);
INSERT INTO public.users VALUES (36, 'F bio Pimentel', 'fabio.pimentel212@gmail.com', 'Bauru', 'SP', '382112', NULL);
INSERT INTO public.users VALUES (37, 'Murilo Batista', 'murilo.batista199@gmail.com', 'Ubatuba', 'SP', '794320', NULL);
INSERT INTO public.users VALUES (38, 'Nat lia Gomes', 'natalia.gomes196@gmail.com', 'Arraial do Cabo', 'RJ', '602513', NULL);
INSERT INTO public.users VALUES (39, 'Lorena Dias', 'lorena.dias211@gmail.com', 'Uberlƒndia', 'MG', '643917', NULL);
INSERT INTO public.users VALUES (40, 'Priscila Martins', 'priscila.martins194@gmail.com', 'Tiradentes', 'MG', '810026', NULL);
INSERT INTO public.users VALUES (41, 'Ricardo Farias', 'ricardo.farias195@gmail.com', 'Canela', 'RS', '990348', NULL);
INSERT INTO public.users VALUES (42, 'Diego Cunha', 'diego.cunha219@gmail.com', 'Caruaru', 'PE', '952316', NULL);
INSERT INTO public.users VALUES (43, 'Alexandre Duarte', 'alexandre.duarte193@gmail.com', 'Ilh‚us', 'BA', '457680', NULL);
INSERT INTO public.users VALUES (44, 'Helena Carvalho', 'helena.carvalho200@gmail.com', 'Itacar‚', 'BA', '239851', NULL);
INSERT INTO public.users VALUES (45, 'Isis Ferreira', 'isis.ferreira215@gmail.com', 'Boa Vista', 'RR', '104559', NULL);
INSERT INTO public.users VALUES (46, 'Eduarda Martins', 'eduarda.martins216@gmail.com', 'Joinville', 'SC', '178194', NULL);
INSERT INTO public.users VALUES (47, 'Rebeca Vasconcelos', 'rebeca.vasconcelos220@gmail.com', 'Juazeiro do Norte', 'CE', '820013', NULL);
INSERT INTO public.users VALUES (48, 'Felipe Soares', 'felipe.soares197@gmail.com', 'Jericoacoara', 'CE', '928471', NULL);
INSERT INTO public.users VALUES (49, 'J‚ssica Castro', 'jessica.castro198@gmail.com', 'SÆo Miguel dos Milagres', 'AL', '153709', NULL);
INSERT INTO public.users VALUES (50, 'Rafael Castro', 'rafael.castro234@gmail.com', 'Palmas', 'TO', '785412#', NULL);
INSERT INTO public.users VALUES (51, 'Juliana Paiva', 'juliana.paiva235@gmail.com', 'SÆo Paulo', 'SP', '634291#', NULL);
INSERT INTO public.users VALUES (52, 'Andr‚ Nascimento', 'andre.nascimento236@gmail.com', 'Rio de Janeiro', 'RJ', '981276#', NULL);
INSERT INTO public.users VALUES (53, 'Bruna Oliveira', 'bruna.oliveira237@gmail.com', 'Curitiba', 'PR', '453018#', NULL);
INSERT INTO public.users VALUES (54, 'Henrique Souza', 'henrique.souza238@gmail.com', 'Recife', 'PE', '719830#', NULL);
INSERT INTO public.users VALUES (55, 'Let¡cia Andrade', 'leticia.andrade239@gmail.com', 'Porto Seguro', 'BA', '168392#', NULL);
INSERT INTO public.users VALUES (56, 'Gustavo Moura', 'gustavo.moura240@gmail.com', 'JoÆo Pessoa', 'PB', '823105#', NULL);
INSERT INTO public.users VALUES (57, 'Marina Duarte', 'marina.duarte241@gmail.com', 'Aracaju', 'SE', '290186#', NULL);
INSERT INTO public.users VALUES (58, 'Eduardo Pinto', 'eduardo.pinto242@gmail.com', 'Macap ', 'AP', '331094#', NULL);
INSERT INTO public.users VALUES (59, 'Talita Ribeiro', 'talita.ribeiro243@gmail.com', 'Rio Branco', 'AC', '557319#', NULL);
INSERT INTO public.users VALUES (60, 'C‚sar Martins', 'cesar.martins244@gmail.com', 'Boa Vista', 'RR', '671042#', NULL);
INSERT INTO public.users VALUES (61, 'Mirela Cardoso', 'mirela.cardoso245@gmail.com', 'Petrolina', 'PE', '442011#', NULL);
INSERT INTO public.users VALUES (62, 'Paulo Henrique', 'paulo.henrique246@gmail.com', 'Maragogi', 'AL', '799034#', NULL);
INSERT INTO public.users VALUES (63, 'Nina Bastos', 'nina.bastos247@gmail.com', 'Chapada dos Veadeiros', 'GO', '915376#', NULL);
INSERT INTO public.users VALUES (64, 'Alana Cruz', 'alana.cruz248@gmail.com', 'Foz do Igua‡u', 'PR', '348190#', NULL);
INSERT INTO public.users VALUES (65, 'Sabrina Teles', 'sabrina.teles249@gmail.com', 'Paraty', 'RJ', '624781#', NULL);
INSERT INTO public.users VALUES (66, 'Fernando Mello', 'fernando.mello250@gmail.com', 'SÆo Lu¡s', 'MA', '813470#', NULL);
INSERT INTO public.users VALUES (67, 'Amanda Prado', 'amanda.prado251@gmail.com', 'Balne rio Cambori£', 'SC', '938624#', NULL);
INSERT INTO public.users VALUES (68, 'Tatiana Correia', 'tatiana.correia253@gmail.com', 'Vit¢ria da Conquista', 'BA', '409782#', NULL);
INSERT INTO public.users VALUES (69, 'Cristiano Neves', 'cristiano.neves254@gmail.com', 'Campina Grande', 'PB', '517209#', NULL);
INSERT INTO public.users VALUES (70, 'Elisa Martins', 'elisa.martins255@gmail.com', 'Altamira', 'PA', '374821#', NULL);
INSERT INTO public.users VALUES (71, 'Rodrigo Sales', 'rodrigo.sales256@gmail.com', 'Rondon¢polis', 'MT', '902184#', NULL);
INSERT INTO public.users VALUES (72, 'Marcela Luz', 'marcela.luz258@gmail.com', 'Teres¢polis', 'RJ', '281905@', NULL);
INSERT INTO public.users VALUES (73, 'Eduardo Lima', 'eduardo.lima259@gmail.com', 'Caxias do Sul', 'RS', '650378@', NULL);
INSERT INTO public.users VALUES (74, 'J£lia Viana', 'julia.viana260@gmail.com', 'Franca', 'SP', '195024@', NULL);
INSERT INTO public.users VALUES (75, 'Mariane Costa', 'mariane.costa262@gmail.com', 'Itabuna', 'BA', '347902@', NULL);
INSERT INTO public.users VALUES (76, 'Leandro Rocha', 'leandro.rocha252@gmail.com', 'Porto Velho', 'RO', '110348#', NULL);
INSERT INTO public.users VALUES (77, 'Daniela Furtado', 'daniela.furtado257@gmail.com', 'SÆo JoÆo del-Rei', 'MG', '764392@', NULL);
INSERT INTO public.users VALUES (78, 'Pedro Cunha', 'pedro.cunha261@gmail.com', 'Piracicaba', 'SP', '817402@', NULL);
INSERT INTO public.users VALUES (79, 'Larissa Freitas', 'larissa.freitas233@gmail.com', 'Vit¢ria', 'ES', '213904#', NULL);
INSERT INTO public.users VALUES (80, 'Ot vio Nunes', 'otavio.nunes263@gmail.com', 'Penedo', 'AL', '713289@', NULL);
INSERT INTO public.users VALUES (81, 'Cíntia Fernandes', 'cintia.fernandes190@gmail.com', 'Ilhabela', 'SP', '300742@', '1747782356940-CÃ­ntia Fernandes.jpg');
INSERT INTO public.users VALUES (82, 'Rodrigo Leal', 'rodrigo.leal183@gmail.com', 'Paraty', 'RJ', '430872@', '1747782394334-Rodrigo Leal.jpg');
INSERT INTO public.users VALUES (83, 'Marcelo Barros', 'marcelo.barros187@gmail.com', 'Ouro Preto', 'MG', '811935@', '1747782426625-Marcelo Barros.jpg');
INSERT INTO public.users VALUES (84, 'Eduardo Campos', 'eduardo.campos181@gmail.com', 'Foz do Iguaçu', 'PR', '781304@', '1747782464178-Eduardo Campos.jpg');
INSERT INTO public.users VALUES (85, 'Beatriz Tavares', 'beatriz.tavares186@gmail.com', 'Gramado', 'RS', '540161#', '1747782504371-Beatriz Tavares.jpg');
INSERT INTO public.users VALUES (86, 'Daniela Freitas', 'daniela.freitas182@gmail.com', 'Olinda', 'PE', '954702#', '1747782536309-Daniela Freitas.jpg');
INSERT INTO public.users VALUES (87, 'Tatiane Lopes', 'tatiane.lopes188@gmail.com', 'Fernando de Noronha', 'PE', '731508%', '1747782574824-Tatiane Lopes.jpg');
INSERT INTO public.users VALUES (88, 'Sérgio Araújo', 'sergio.araujo185@gmail.com', 'Lençóis', 'BA', '206187%', '1747782626830-SÃ©rgio AraÃºjo.jpg');
INSERT INTO public.users VALUES (89, 'Vanessa Prado', 'vanessa.prado184@gmail.com', 'Bonito', 'MS', '629103%', '1747782664492-Vanessa Prado.jpg');
INSERT INTO public.users VALUES (90, 'Rogério Melo', 'rogerio.melo189@gmail.com', 'Chapada dos Veadeiros', 'GO', '647091%', '1747782708815-RogÃ©rio Melo.jpg');
INSERT INTO public.users VALUES (91, 'Gustavo Reis', 'gustavo.reis191@gmail.com', 'Petrolina', 'PE', '284105%', '1747782746332-Gustavo Reis.jpg');
INSERT INTO public.users VALUES (92, 'Larissa Almeida', 'larissa.almeida123@gmail.com', 'Campinas', 'SP', '983214@', NULL);
INSERT INTO public.users VALUES (93, 'Renato Cardoso', 'renato.cardoso99@gmail.com', 'Belo Horizonte', 'MG', '612347@', NULL);
INSERT INTO public.users VALUES (94, 'Juliana Mendes', 'juliana.mendes@gmail.com', 'Curitiba', 'PR', '704128@', NULL);
INSERT INTO public.users VALUES (95, 'Tiago Oliveira', 'tiago.oliveira87@gmail.com', 'Porto Alegre', 'RS', '559876@', NULL);
INSERT INTO public.users VALUES (96, 'Bruna Souza', 'bruna.souza@gmail.com', 'Salvador', 'BA', '347281@', NULL);
INSERT INTO public.users VALUES (97, 'Marcelo Ribeiro', 'marcelo.ribeiro@hotmail.com', 'Recife', 'PE', '890173$', NULL);
INSERT INTO public.users VALUES (98, 'Diego Nascimento', 'diego.nascimento@gmail.com', 'JoÆo Pessoa', 'PB', '519283$', NULL);
INSERT INTO public.users VALUES (99, 'Camila Duarte', 'camila.duarte@gmail.com', 'Fortaleza', 'CE', '209134$', NULL);
INSERT INTO public.users VALUES (100, 'Andr‚ Lima', 'andre.lima@gmail.com', 'Goiƒnia', 'GO', '381947$', NULL);


--
-- TOC entry 4851 (class 0 OID 0)
-- Dependencies: 216
-- Name: comments_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.comments_id_seq', 101, true);


--
-- TOC entry 4852 (class 0 OID 0)
-- Dependencies: 218
-- Name: feedbacks_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.feedbacks_id_seq', 21, true);


--
-- TOC entry 4853 (class 0 OID 0)
-- Dependencies: 220
-- Name: news_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.news_id_seq', 11, true);


--
-- TOC entry 4854 (class 0 OID 0)
-- Dependencies: 222
-- Name: posts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.posts_id_seq', 251, true);


--
-- TOC entry 4855 (class 0 OID 0)
-- Dependencies: 224
-- Name: regions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.regions_id_seq', 60, true);


--
-- TOC entry 4856 (class 0 OID 0)
-- Dependencies: 226
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_id_seq', 102, true);


--
-- TOC entry 4668 (class 2606 OID 17732)
-- Name: comments comments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);


--
-- TOC entry 4670 (class 2606 OID 17734)
-- Name: feedbacks feedbacks_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.feedbacks
    ADD CONSTRAINT feedbacks_pkey PRIMARY KEY (id);


--
-- TOC entry 4672 (class 2606 OID 17736)
-- Name: news news_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.news
    ADD CONSTRAINT news_pkey PRIMARY KEY (id);


--
-- TOC entry 4674 (class 2606 OID 17738)
-- Name: posts posts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_pkey PRIMARY KEY (id);


--
-- TOC entry 4676 (class 2606 OID 17740)
-- Name: regions regions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.regions
    ADD CONSTRAINT regions_pkey PRIMARY KEY (id);


--
-- TOC entry 4678 (class 2606 OID 17742)
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- TOC entry 4680 (class 2606 OID 17744)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- TOC entry 4681 (class 2606 OID 17745)
-- Name: comments comments_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id) ON DELETE CASCADE;


--
-- TOC entry 4682 (class 2606 OID 17750)
-- Name: comments comments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 4683 (class 2606 OID 17755)
-- Name: feedbacks feedbacks_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.feedbacks
    ADD CONSTRAINT feedbacks_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 4684 (class 2606 OID 17760)
-- Name: posts posts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


-- Completed on 2025-06-10 13:54:12

--
-- PostgreSQL database dump complete
--

