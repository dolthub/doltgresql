-- Downloaded from: https://github.com/luizantoniocardoso/trabalho-banco-2/blob/4b6079932e0a126878cc634fd5c18195923c6e13/banco/backuptrabalho.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 15.3
-- Dumped by pg_dump version 15.3

-- Started on 2024-06-24 17:50:09

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
-- TOC entry 2 (class 3079 OID 78557)
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- TOC entry 3494 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- TOC entry 248 (class 1255 OID 78699)
-- Name: atualizar_total_manutencoes_concluidas(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.atualizar_total_manutencoes_concluidas() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.status = 'Concluido' THEN
        UPDATE equipamentos 
        SET total_manutencoes = total_manutencoes + 1
        WHERE id = NEW.equipamento_id;
    END IF;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.atualizar_total_manutencoes_concluidas() OWNER TO postgres;

--
-- TOC entry 245 (class 1255 OID 78696)
-- Name: calcular_idade_tecnico(uuid); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.calcular_idade_tecnico(tecnico_id uuid) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    idade INT;
BEGIN
    SELECT EXTRACT(YEAR FROM AGE(data_nascimento)) INTO idade
    FROM tecnicos
    WHERE id = tecnico_id;

    RETURN idade;
END;
$$;


ALTER FUNCTION public.calcular_idade_tecnico(tecnico_id uuid) OWNER TO postgres;

--
-- TOC entry 246 (class 1255 OID 78697)
-- Name: equipamento_em_manutencao(uuid); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.equipamento_em_manutencao(equip_id uuid) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
    em_manutencao BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1 
        FROM manutencoes 
        WHERE equipamento_id = equip_id AND status IN ('EmAndamento', 'Agendado')
    ) INTO em_manutencao;

    RETURN em_manutencao;
END;
$$;


ALTER FUNCTION public.equipamento_em_manutencao(equip_id uuid) OWNER TO postgres;

--
-- TOC entry 247 (class 1255 OID 78698)
-- Name: total_pecas_usadas(uuid); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.total_pecas_usadas(manut_id uuid) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    total INT;
BEGIN
    SELECT SUM(quantidade) INTO total
    FROM pecas_usadas
    WHERE manutencao_id = manut_id;

    RETURN COALESCE(total, 0);
END;
$$;


ALTER FUNCTION public.total_pecas_usadas(manut_id uuid) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 215 (class 1259 OID 78299)
-- Name: __EFMigrationsHistory; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."__EFMigrationsHistory" (
    "MigrationId" character varying(150) NOT NULL,
    "ProductVersion" character varying(32) NOT NULL
);


ALTER TABLE public."__EFMigrationsHistory" OWNER TO postgres;

--
-- TOC entry 216 (class 1259 OID 78304)
-- Name: categorias; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.categorias (
    id uuid NOT NULL,
    nome character varying(150) NOT NULL,
    descricao text
);


ALTER TABLE public.categorias OWNER TO postgres;

--
-- TOC entry 219 (class 1259 OID 78350)
-- Name: equipamentos; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.equipamentos (
    id uuid NOT NULL,
    numero_de_serie character varying(100),
    data_aquisicao date,
    modelo_id uuid,
    localizacao_id uuid,
    categoria_id uuid,
    nome character varying(150) NOT NULL,
    descricao text,
    total_manutencoes integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.equipamentos OWNER TO postgres;

--
-- TOC entry 222 (class 1259 OID 78438)
-- Name: equipamentos_pecas; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.equipamentos_pecas (
    equipamento_id uuid NOT NULL,
    peca_id uuid NOT NULL
);


ALTER TABLE public.equipamentos_pecas OWNER TO postgres;

--
-- TOC entry 227 (class 1259 OID 78496)
-- Name: especializacoes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.especializacoes (
    id uuid NOT NULL,
    nome character varying(150) NOT NULL,
    descricao text
);


ALTER TABLE public.especializacoes OWNER TO postgres;

--
-- TOC entry 223 (class 1259 OID 78454)
-- Name: ferramentas; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ferramentas (
    id uuid NOT NULL,
    nome character varying(150) NOT NULL,
    descricao text
);


ALTER TABLE public.ferramentas OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 78311)
-- Name: localizacoes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.localizacoes (
    id uuid NOT NULL,
    nome character varying(150) NOT NULL,
    descricao text
);


ALTER TABLE public.localizacoes OWNER TO postgres;

--
-- TOC entry 220 (class 1259 OID 78417)
-- Name: manutencoes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.manutencoes (
    id uuid NOT NULL,
    data_inicio date NOT NULL,
    data_conclusao date NOT NULL,
    status text NOT NULL,
    tipo_manutencao text NOT NULL,
    equipamento_id uuid NOT NULL,
    nome character varying(150) NOT NULL,
    descricao text,
    CONSTRAINT chk_status_manutencao CHECK ((status = ANY (ARRAY['Pendente'::text, 'EmAndamento'::text, 'Concluido'::text, 'Encerrado'::text]))),
    CONSTRAINT chk_tipo_manutencao CHECK ((tipo_manutencao = ANY (ARRAY['Preventiva'::text, 'Corretiva'::text])))
);


ALTER TABLE public.manutencoes OWNER TO postgres;

--
-- TOC entry 218 (class 1259 OID 78318)
-- Name: modelos; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.modelos (
    id uuid NOT NULL,
    nome character varying(150) NOT NULL,
    descricao text
);


ALTER TABLE public.modelos OWNER TO postgres;

--
-- TOC entry 225 (class 1259 OID 78466)
-- Name: ordens_servicos; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ordens_servicos (
    manutencao_id uuid NOT NULL,
    tecnico_id uuid NOT NULL,
    papel_id uuid DEFAULT '00000000-0000-0000-0000-000000000000'::uuid NOT NULL
);


ALTER TABLE public.ordens_servicos OWNER TO postgres;

--
-- TOC entry 226 (class 1259 OID 78483)
-- Name: papeis; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.papeis (
    id uuid NOT NULL,
    nome character varying(150) NOT NULL,
    descricao text
);


ALTER TABLE public.papeis OWNER TO postgres;

--
-- TOC entry 221 (class 1259 OID 78431)
-- Name: pecas; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.pecas (
    id uuid NOT NULL,
    nome character varying(150) NOT NULL,
    descricao text
);


ALTER TABLE public.pecas OWNER TO postgres;

--
-- TOC entry 228 (class 1259 OID 78514)
-- Name: pecas_usadas; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.pecas_usadas (
    peca_id uuid NOT NULL,
    manutencao_id uuid NOT NULL,
    quantidade integer NOT NULL
);


ALTER TABLE public.pecas_usadas OWNER TO postgres;

--
-- TOC entry 224 (class 1259 OID 78461)
-- Name: tecnicos; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tecnicos (
    id uuid NOT NULL,
    nome character varying(150) NOT NULL,
    telefone character varying(15) NOT NULL,
    cpf character varying(11) NOT NULL,
    data_nascimento date NOT NULL,
    especializacao_id uuid
);


ALTER TABLE public.tecnicos OWNER TO postgres;

--
-- TOC entry 230 (class 1259 OID 78674)
-- Name: vw_equipamentos_completos; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.vw_equipamentos_completos AS
 SELECT e.id AS equipamento_id,
    e.nome AS equipamento_nome,
    e.descricao AS equipamento_descricao,
    e.numero_de_serie,
    e.data_aquisicao,
    c.nome AS categoria_nome,
    m.nome AS modelo_nome,
    l.nome AS localizacao_nome
   FROM (((public.equipamentos e
     JOIN public.categorias c ON ((e.categoria_id = c.id)))
     JOIN public.modelos m ON ((e.modelo_id = m.id)))
     JOIN public.localizacoes l ON ((e.localizacao_id = l.id)));


ALTER TABLE public.vw_equipamentos_completos OWNER TO postgres;

--
-- TOC entry 229 (class 1259 OID 78553)
-- Name: vw_equipamentos_pecas; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.vw_equipamentos_pecas AS
 SELECT e.nome AS equipamento,
    p.nome AS peca
   FROM ((public.equipamentos e
     JOIN public.equipamentos_pecas ep ON ((e.id = ep.equipamento_id)))
     JOIN public.pecas p ON ((ep.peca_id = p.id)));


ALTER TABLE public.vw_equipamentos_pecas OWNER TO postgres;

--
-- TOC entry 232 (class 1259 OID 78684)
-- Name: vw_equipamentos_por_categoria; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.vw_equipamentos_por_categoria AS
 SELECT c.nome AS categoria_nome,
    count(e.id) AS quantidade_equipamentos
   FROM (public.categorias c
     LEFT JOIN public.equipamentos e ON ((e.categoria_id = c.id)))
  GROUP BY c.nome;


ALTER TABLE public.vw_equipamentos_por_categoria OWNER TO postgres;

--
-- TOC entry 231 (class 1259 OID 78679)
-- Name: vw_manutencoes_concluidas; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.vw_manutencoes_concluidas AS
 SELECT m.id AS manutencao_id,
    m.nome AS manutencao_nome,
    m.descricao AS manutencao_descricao,
    m.data_inicio,
    m.data_conclusao,
    m.tipo_manutencao,
    e.nome AS equipamento_nome,
    t.nome AS tecnico_nome
   FROM (((public.manutencoes m
     JOIN public.equipamentos e ON ((m.equipamento_id = e.id)))
     JOIN public.ordens_servicos os ON ((m.id = os.manutencao_id)))
     JOIN public.tecnicos t ON ((os.tecnico_id = t.id)))
  WHERE (m.status = 'Concluido'::text);


ALTER TABLE public.vw_manutencoes_concluidas OWNER TO postgres;

--
-- TOC entry 234 (class 1259 OID 78692)
-- Name: vw_pecas_usadas_por_manutencao; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.vw_pecas_usadas_por_manutencao AS
 SELECT m.id AS manutencao_id,
    m.nome AS manutencao_nome,
    p.nome AS peca_nome,
    pu.quantidade
   FROM ((public.manutencoes m
     JOIN public.pecas_usadas pu ON ((m.id = pu.manutencao_id)))
     JOIN public.pecas p ON ((pu.peca_id = p.id)));


ALTER TABLE public.vw_pecas_usadas_por_manutencao OWNER TO postgres;

--
-- TOC entry 233 (class 1259 OID 78688)
-- Name: vw_tecnicos_especializacoes; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.vw_tecnicos_especializacoes AS
 SELECT t.id AS tecnico_id,
    t.nome AS tecnico_nome,
    t.telefone,
    t.cpf,
    t.data_nascimento,
    e.nome AS especializacao_nome
   FROM (public.tecnicos t
     JOIN public.especializacoes e ON ((t.especializacao_id = e.id)));


ALTER TABLE public.vw_tecnicos_especializacoes OWNER TO postgres;

--
-- TOC entry 3475 (class 0 OID 78299)
-- Dependencies: 215
-- Data for Name: __EFMigrationsHistory; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public."__EFMigrationsHistory" ("MigrationId", "ProductVersion") FROM stdin;
20240619200004_InitialCreate	7.0.0
20240620195745_EquipamentoAdicionado_CadastroGeralItemRemovido	7.0.0
20240620200206_AjustandoColunasEquipamento	7.0.0
20240621003700_DeBigintParaStringNumeroDeSerie	7.0.0
20240621184143_AdicionadoManutencao	7.0.0
20240623183921_AddPeca_AddEquipamentoPeca	7.0.0
20240623185925_AddFerramenta	7.0.0
20240623191957_AddOrdemServico_AddTecnico	7.0.0
20240623193821_AddPapel	7.0.0
20240623200416_addEspecializacao	7.0.0
20240623201809_AddPecaUtilizada	7.0.0
20240623205638_Setando como null coluna categoria_id em Equipamentos	7.0.0
20240623222345_Campos setados como NULL	7.0.0
20240623231422_Adicionando Indicies	7.0.0
20240624031044_Mapeando colunas Nome e Descrição em Manutencao	7.0.0
20240624113038_Adicionando indices em todos as colunas nome	7.0.0
20240624113817_Adicionando checks	7.0.0
20240624114959_data_conclusao e anulavel	7.0.0
20240624182629_Adicionado coluna com total de manutencoes	7.0.0
\.


--
-- TOC entry 3476 (class 0 OID 78304)
-- Dependencies: 216
-- Data for Name: categorias; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.categorias (id, nome, descricao) FROM stdin;
7d7d95b2-fb0c-46ec-94dc-0fb05b382d27	Máquina Leve	Equipamentos leves para operações simples
8664d544-6c50-4b89-b9df-322a12ddefe8	Máquina Pesada	Equipamentos pesados para operações robustas
1e0a3dc0-eda7-44fa-b426-802304af1f89	Máquina Robusta	Equipamentos robustos para serviços de alta demanda
438c430b-4158-48ed-b0ec-8b1d60be2ef8	Máquina Simples	Equipamentos simples e básicos para uso diário
6a0801d8-9fe4-4acd-98a0-8c22f3d3bd81	Máquina Complexa	Equipamentos complexos para tarefas especializadas
\.


--
-- TOC entry 3479 (class 0 OID 78350)
-- Dependencies: 219
-- Data for Name: equipamentos; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.equipamentos (id, numero_de_serie, data_aquisicao, modelo_id, localizacao_id, categoria_id, nome, descricao, total_manutencoes) FROM stdin;
e2c82995-235a-49e5-85b7-d4f90706b1e6	12345	2022-01-01	a0adcd91-b2d3-4d12-a820-afba6869886d	6a3738f7-c5fc-4214-93d5-503babab075b	7d7d95b2-fb0c-46ec-94dc-0fb05b382d27	Compressor InnerMax	Compressor de alta eficiência para operações leves	0
834aadb1-c784-4d9e-8b59-6f42b4420adb	67890	2022-02-01	a139b7df-d056-4db4-b9ee-524287ee416e	40d01345-e720-4a63-b275-ce10fa0bc708	8664d544-6c50-4b89-b9df-322a12ddefe8	Escavadeira TomboCore	Escavadeira robusta para terrenos difíceis	0
50084eb5-831f-48de-a798-783e721ca3ec	54321	2022-03-01	168dcbb5-1f39-40e5-a20f-d1d940dbaf8e	55080044-30b5-4e46-8fed-35858ec84326	1e0a3dc0-eda7-44fa-b426-802304af1f89	Perfuradora PowerDrive	Perfuradora de alta potência	0
15cb4721-7d35-4f9c-87a5-a9c7e1c885be	09876	2022-04-01	a9123b83-4a38-48b7-9368-bf6ed74792ea	f2a08358-0fde-4ba1-b3e3-80843c80db35	438c430b-4158-48ed-b0ec-8b1d60be2ef8	Soldadora Simplex	Soldadora básica para pequenos reparos	0
e3660625-4839-4cfb-b305-4c0ba0dab0ed	11223	2022-05-01	cdae84d3-96ad-481d-8ca4-6d0c2955f9f6	16ff2ccd-2f8d-4390-99ef-a396af10bf55	6a0801d8-9fe4-4acd-98a0-8c22f3d3bd81	Fresadora ComplexoX	Fresadora especializada para cortes complexos	0
7715cf8d-e42a-42c7-885c-68bb3ceb0224	CMP-12345	2023-01-15	a0adcd91-b2d3-4d12-a820-afba6869886d	6a3738f7-c5fc-4214-93d5-503babab075b	7d7d95b2-fb0c-46ec-94dc-0fb05b382d27	Comprressor de ar incompleto	\N	0
3ec2cab3-7816-476a-ad75-71d3c932c563	GER-67890	2023-02-20	a139b7df-d056-4db4-b9ee-524287ee416e	40d01345-e720-4a63-b275-ce10fa0bc708	8664d544-6c50-4b89-b9df-322a12ddefe8	Gerador incompleto	\N	0
76295844-5e8e-4d6e-a496-a1771f45ee5d	\N	2023-03-10	168dcbb5-1f39-40e5-a20f-d1d940dbaf8e	55080044-30b5-4e46-8fed-35858ec84326	1e0a3dc0-eda7-44fa-b426-802304af1f89	Perfuradora incompleta	Perfuradora de alta potência incompleta	0
f75dce84-a9b3-4532-acfc-4d3ceeac3b0a	SLD-09876	\N	a9123b83-4a38-48b7-9368-bf6ed74792ea	f2a08358-0fde-4ba1-b3e3-80843c80db35	438c430b-4158-48ed-b0ec-8b1d60be2ef8	Soldadora incompleta	Soldadora básica para pequenos reparos	0
188362c1-f8f3-4102-86be-b46825f80f7e	FRS-11223	2023-04-01	\N	16ff2ccd-2f8d-4390-99ef-a396af10bf55	6a0801d8-9fe4-4acd-98a0-8c22f3d3bd81	Fresadora incompleta	Fresadora para cortes complexos	0
\.


--
-- TOC entry 3482 (class 0 OID 78438)
-- Dependencies: 222
-- Data for Name: equipamentos_pecas; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.equipamentos_pecas (equipamento_id, peca_id) FROM stdin;
e2c82995-235a-49e5-85b7-d4f90706b1e6	f4e898e1-49a7-4b7a-8ed1-b465560570c6
834aadb1-c784-4d9e-8b59-6f42b4420adb	5181d1fe-1038-47d2-933d-dd336597d2cb
50084eb5-831f-48de-a798-783e721ca3ec	87306f9b-4084-420a-93c9-b26ebba284d5
15cb4721-7d35-4f9c-87a5-a9c7e1c885be	99facdbc-e73a-452e-bd0d-24dcff076656
e3660625-4839-4cfb-b305-4c0ba0dab0ed	6d8bb300-55f3-46f6-8475-556eb3b32d7c
\.


--
-- TOC entry 3487 (class 0 OID 78496)
-- Dependencies: 227
-- Data for Name: especializacoes; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.especializacoes (id, nome, descricao) FROM stdin;
885f8523-3812-43a2-865a-762cf4222777	Mecânica Geral	Especialização em mecânica geral de equipamentos
ec6e25a4-13f2-450d-9a60-be19cb15696e	Hidráulica	Especialização em sistemas hidráulicos
e063926f-c903-4228-9ceb-50115712e368	Elétrica	Especialização em sistemas elétricos
84954fe0-4372-480a-a14d-cbf579f5c131	Soldagem	Especialização em soldagem de metais
8cf8cf58-e81b-43fc-802f-e1741f9d2752	Manutenção Preventiva	Especialização em manutenção preventiva de equipamentos
\.


--
-- TOC entry 3483 (class 0 OID 78454)
-- Dependencies: 223
-- Data for Name: ferramentas; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.ferramentas (id, nome, descricao) FROM stdin;
00c7de3b-e6ba-4a55-9cd1-bf55e6ed347d	Martelo	Ferramenta básica para trabalhos gerais
51537e7b-a6c9-4bb3-bd5b-2fc086d1f0b7	Serra de Cortar	Serra elétrica para cortes precisos
8117a90f-119e-4ae8-b304-3a2c5ef31c83	Chave de Fenda	Ferramenta para apertar e soltar parafusos
811f4556-3aee-42ac-bc12-b36d11b24b09	Alicate	Ferramenta para segurar e cortar objetos
6b8e642d-6a48-454e-8102-73cc112dffd9	Esmeril	Ferramenta para afiar e desbastar materiais
\.


--
-- TOC entry 3477 (class 0 OID 78311)
-- Dependencies: 217
-- Data for Name: localizacoes; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.localizacoes (id, nome, descricao) FROM stdin;
6a3738f7-c5fc-4214-93d5-503babab075b	Galpão Sul	Localização no galpão sul da fábrica
40d01345-e720-4a63-b275-ce10fa0bc708	Galpão Norte	Localização no galpão norte da fábrica
55080044-30b5-4e46-8fed-35858ec84326	Galpão Leste	Localização no galpão leste da fábrica
f2a08358-0fde-4ba1-b3e3-80843c80db35	Galpão Oeste	Localização no galpão oeste da fábrica
16ff2ccd-2f8d-4390-99ef-a396af10bf55	Armazém Central	Localização no armazém central da fábrica
\.


--
-- TOC entry 3480 (class 0 OID 78417)
-- Dependencies: 220
-- Data for Name: manutencoes; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.manutencoes (id, data_inicio, data_conclusao, status, tipo_manutencao, equipamento_id, nome, descricao) FROM stdin;
ce7937a3-97ef-4e66-b520-a4f7f6bf24e3	2023-01-01	2023-01-02	Concluido	Preventiva	e2c82995-235a-49e5-85b7-d4f90706b1e6	Troca de Óleo Compressor	Troca de óleo do compressor InnerMax
e0ff20de-979f-4b1e-b9ca-ef22854ce78f	2023-02-01	2023-02-02	EmAndamento	Corretiva	834aadb1-c784-4d9e-8b59-6f42b4420adb	Reparo Escavadeira	Reparo na escavadeira TomboCore
4c71c3fe-a73f-4f79-9d8b-efc62b872e30	2023-03-01	2023-03-02	Encerrado	Preventiva	50084eb5-831f-48de-a798-783e721ca3ec	Manutenção Perfuradora	Manutenção geral da perfuradora PowerDrive
87fe38ee-8c97-4b7f-9f44-d11e6d428de4	2023-04-01	2023-04-02	Concluido	Corretiva	15cb4721-7d35-4f9c-87a5-a9c7e1c885be	Ajuste Soldadora	Ajuste na soldadora Simplex
0e935eda-812f-4765-becd-d9969abac77e	2023-05-01	2023-05-02	EmAndamento	Preventiva	e3660625-4839-4cfb-b305-4c0ba0dab0ed	Calibração Fresadora	Calibração da fresadora ComplexoX
498f5c7c-7b54-48f2-8ce5-91e1ffab8f7d	2024-06-14	2024-06-15	Concluido	Preventiva	e2c82995-235a-49e5-85b7-d4f90706b1e6	Revisão Preventiva Compressor	Revisão preventiva do compressor InnerMax
cf82ac58-f9a0-4818-9091-e8d0864cec49	2024-06-04	2024-06-05	Concluido	Corretiva	76295844-5e8e-4d6e-a496-a1771f45ee5d	Troca de Peças Perfuradora	Troca de peças desgastadas da perfuradora
c1e0bdbb-6d40-49ac-84d6-a91bfff8b0d6	2024-06-09	2024-06-10	Concluido	Preventiva	188362c1-f8f3-4102-86be-b46825f80f7e	Ajuste de Precisão Fresadora	Ajuste de precisão da fresadora ComplexoX
d4837917-7f89-42aa-b70b-13478c8e34ae	2023-06-08	2023-06-08	Concluido	Corretiva	e2c82995-235a-49e5-85b7-d4f90706b1e6	Substituição Completa de Filtros	Substituição completa dos filtros do compressor
c5f1533a-d8fb-431c-a924-98039f4bec00	2023-06-01	2023-06-01	Concluido	Preventiva	e2c82995-235a-49e5-85b7-d4f90706b1e6	Revisão Geral Compressor	Revisão completa do compressor InnerMax
d39ddb50-296f-4448-b1e8-83e68e533fcf	2023-06-02	2023-06-02	Concluido	Preventiva	3ec2cab3-7816-476a-ad75-71d3c932c563	Troca de Óleo Gerador	Troca de óleo do gerador
0cbaa8bc-aef0-4599-b2cc-a4f17ffe0ae2	2023-06-03	2023-06-03	Concluido	Corretiva	76295844-5e8e-4d6e-a496-a1771f45ee5d	Substituição de Filtros	Substituição de filtros da perfuradora
8a1cfb52-3124-4276-a8de-9e44194bfc48	2023-06-04	2023-06-04	Concluido	Corretiva	f75dce84-a9b3-4532-acfc-4d3ceeac3b0a	Ajuste de Soldadora	Ajuste de potência da soldadora
2f7277f0-5225-4d04-9f54-4d2906642846	2023-06-05	2023-06-05	Concluido	Preventiva	188362c1-f8f3-4102-86be-b46825f80f7e	Calibração de Fresadora	Calibração de precisão da fresadora
bb0f772d-68f9-4b63-b1f1-b562b71c1218	2023-06-06	2023-06-06	Concluido	Preventiva	e2c82995-235a-49e5-85b7-d4f90706b1e6	Inspeção de Segurança	Inspeção de segurança do compressor
68ec0505-6e7f-438f-a008-d3a5032ff59b	2023-06-07	2023-06-07	Concluido	Corretiva	3ec2cab3-7816-476a-ad75-71d3c932c563	Reparo de Circuito Elétrico	Reparo do circuito elétrico do gerador
\.


--
-- TOC entry 3478 (class 0 OID 78318)
-- Dependencies: 218
-- Data for Name: modelos; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.modelos (id, nome, descricao) FROM stdin;
a139b7df-d056-4db4-b9ee-524287ee416e	TomboCore	Modelo avançado para operações pesadas
a0adcd91-b2d3-4d12-a820-afba6869886d	InnerMax	Modelo eficiente e leve para operações simples
168dcbb5-1f39-40e5-a20f-d1d940dbaf8e	PowerDrive	Modelo robusto para alta performance
a9123b83-4a38-48b7-9368-bf6ed74792ea	Simplex	Modelo básico para tarefas diárias
cdae84d3-96ad-481d-8ca4-6d0c2955f9f6	ComplexoX	Modelo especializado para tarefas complexas
\.


--
-- TOC entry 3485 (class 0 OID 78466)
-- Dependencies: 225
-- Data for Name: ordens_servicos; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.ordens_servicos (manutencao_id, tecnico_id, papel_id) FROM stdin;
c5f1533a-d8fb-431c-a924-98039f4bec00	98b498db-e47d-48bd-9609-010cf634cf67	f39c6cc7-6dd9-49a8-b588-153e1a625956
d39ddb50-296f-4448-b1e8-83e68e533fcf	98b498db-e47d-48bd-9609-010cf634cf67	f39c6cc7-6dd9-49a8-b588-153e1a625956
0cbaa8bc-aef0-4599-b2cc-a4f17ffe0ae2	98b498db-e47d-48bd-9609-010cf634cf67	f39c6cc7-6dd9-49a8-b588-153e1a625956
8a1cfb52-3124-4276-a8de-9e44194bfc48	98b498db-e47d-48bd-9609-010cf634cf67	f39c6cc7-6dd9-49a8-b588-153e1a625956
2f7277f0-5225-4d04-9f54-4d2906642846	98b498db-e47d-48bd-9609-010cf634cf67	f39c6cc7-6dd9-49a8-b588-153e1a625956
bb0f772d-68f9-4b63-b1f1-b562b71c1218	98b498db-e47d-48bd-9609-010cf634cf67	f39c6cc7-6dd9-49a8-b588-153e1a625956
68ec0505-6e7f-438f-a008-d3a5032ff59b	98b498db-e47d-48bd-9609-010cf634cf67	f39c6cc7-6dd9-49a8-b588-153e1a625956
ce7937a3-97ef-4e66-b520-a4f7f6bf24e3	98b498db-e47d-48bd-9609-010cf634cf67	f39c6cc7-6dd9-49a8-b588-153e1a625956
e0ff20de-979f-4b1e-b9ca-ef22854ce78f	f6bec301-a328-416c-9df2-62fcd0cd97fe	9171f9ce-f7c3-43db-96a1-4db6c15e8da9
4c71c3fe-a73f-4f79-9d8b-efc62b872e30	aadabdfc-d081-42a4-8ec5-ca5ed6227d57	0f8ca4fa-4e9c-47db-ac56-80d110b41922
87fe38ee-8c97-4b7f-9f44-d11e6d428de4	7c1e5dd5-686c-4ef4-bcc9-5cce6fd949d6	dcea40d2-bb55-4e30-99d1-dcf1594c01b1
0e935eda-812f-4765-becd-d9969abac77e	d3434182-6dee-40e5-88d9-b9a42256138b	d2865d30-7cbb-4437-a745-1f271e87618b
\.


--
-- TOC entry 3486 (class 0 OID 78483)
-- Dependencies: 226
-- Data for Name: papeis; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.papeis (id, nome, descricao) FROM stdin;
f39c6cc7-6dd9-49a8-b588-153e1a625956	Supervisor	Responsável pela supervisão da manutenção
9171f9ce-f7c3-43db-96a1-4db6c15e8da9	Técnico de Campo	Técnico que realiza a manutenção no local
0f8ca4fa-4e9c-47db-ac56-80d110b41922	Auxiliar Técnico	Auxiliar que ajuda nas manutenções
dcea40d2-bb55-4e30-99d1-dcf1594c01b1	Especialista	Especialista que realiza manutenções complexas
d2865d30-7cbb-4437-a745-1f271e87618b	Coordenador	Coordenador das atividades de manutenção
\.


--
-- TOC entry 3481 (class 0 OID 78431)
-- Dependencies: 221
-- Data for Name: pecas; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pecas (id, nome, descricao) FROM stdin;
f4e898e1-49a7-4b7a-8ed1-b465560570c6	Biela	Peça fundamental para a transmissão de movimento
5181d1fe-1038-47d2-933d-dd336597d2cb	Pistão	Componente do motor responsável pela compressão
87306f9b-4084-420a-93c9-b26ebba284d5	Anel de Vedação	Peça utilizada para vedação em sistemas hidráulicos
99facdbc-e73a-452e-bd0d-24dcff076656	Cilindro	Componente do motor onde ocorre a combustão
6d8bb300-55f3-46f6-8475-556eb3b32d7c	Correia	Peça utilizada para transmissão de força mecânica
\.


--
-- TOC entry 3488 (class 0 OID 78514)
-- Dependencies: 228
-- Data for Name: pecas_usadas; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pecas_usadas (peca_id, manutencao_id, quantidade) FROM stdin;
f4e898e1-49a7-4b7a-8ed1-b465560570c6	ce7937a3-97ef-4e66-b520-a4f7f6bf24e3	5
5181d1fe-1038-47d2-933d-dd336597d2cb	e0ff20de-979f-4b1e-b9ca-ef22854ce78f	10
87306f9b-4084-420a-93c9-b26ebba284d5	4c71c3fe-a73f-4f79-9d8b-efc62b872e30	15
99facdbc-e73a-452e-bd0d-24dcff076656	87fe38ee-8c97-4b7f-9f44-d11e6d428de4	20
6d8bb300-55f3-46f6-8475-556eb3b32d7c	0e935eda-812f-4765-becd-d9969abac77e	25
\.


--
-- TOC entry 3484 (class 0 OID 78461)
-- Dependencies: 224
-- Data for Name: tecnicos; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tecnicos (id, nome, telefone, cpf, data_nascimento, especializacao_id) FROM stdin;
98b498db-e47d-48bd-9609-010cf634cf67	João Silva	123456789	12345678901	1980-01-01	885f8523-3812-43a2-865a-762cf4222777
f6bec301-a328-416c-9df2-62fcd0cd97fe	Maria Souza	987654321	10987654321	1982-02-01	ec6e25a4-13f2-450d-9a60-be19cb15696e
aadabdfc-d081-42a4-8ec5-ca5ed6227d57	Carlos Lima	456123789	45612378901	1984-03-01	e063926f-c903-4228-9ceb-50115712e368
7c1e5dd5-686c-4ef4-bcc9-5cce6fd949d6	Ana Pereira	789456123	78945612301	1986-04-01	84954fe0-4372-480a-a14d-cbf579f5c131
d3434182-6dee-40e5-88d9-b9a42256138b	Pedro Almeida	321654987	32165498701	1988-05-01	8cf8cf58-e81b-43fc-802f-e1741f9d2752
008e1e59-2b16-4974-a450-3bb17bbebd88	Lucas Martins	1122334455	98765432100	1990-08-10	885f8523-3812-43a2-865a-762cf4222777
8eb41cde-ae4c-44bd-8870-f33eb3f768e0	Fernanda Costa	2233445566	12345678902	1992-09-15	885f8523-3812-43a2-865a-762cf4222777
0d66469c-f4fa-4e80-ae1e-2ac8299e90bf	Rafael Almeida	3344556677	45678912300	1988-07-22	885f8523-3812-43a2-865a-762cf4222777
\.


--
-- TOC entry 3268 (class 2606 OID 78303)
-- Name: __EFMigrationsHistory PK___EFMigrationsHistory; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."__EFMigrationsHistory"
    ADD CONSTRAINT "PK___EFMigrationsHistory" PRIMARY KEY ("MigrationId");


--
-- TOC entry 3270 (class 2606 OID 78310)
-- Name: categorias PK_categorias; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.categorias
    ADD CONSTRAINT "PK_categorias" PRIMARY KEY (id);


--
-- TOC entry 3282 (class 2606 OID 78392)
-- Name: equipamentos PK_equipamentos; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.equipamentos
    ADD CONSTRAINT "PK_equipamentos" PRIMARY KEY (id);


--
-- TOC entry 3293 (class 2606 OID 78442)
-- Name: equipamentos_pecas PK_equipamentos_pecas; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.equipamentos_pecas
    ADD CONSTRAINT "PK_equipamentos_pecas" PRIMARY KEY (equipamento_id, peca_id);


--
-- TOC entry 3309 (class 2606 OID 78502)
-- Name: especializacoes PK_especializacoes; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.especializacoes
    ADD CONSTRAINT "PK_especializacoes" PRIMARY KEY (id);


--
-- TOC entry 3295 (class 2606 OID 78460)
-- Name: ferramentas PK_ferramentas; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ferramentas
    ADD CONSTRAINT "PK_ferramentas" PRIMARY KEY (id);


--
-- TOC entry 3273 (class 2606 OID 78317)
-- Name: localizacoes PK_localizacoes; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.localizacoes
    ADD CONSTRAINT "PK_localizacoes" PRIMARY KEY (id);


--
-- TOC entry 3286 (class 2606 OID 78423)
-- Name: manutencoes PK_manutencoes; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.manutencoes
    ADD CONSTRAINT "PK_manutencoes" PRIMARY KEY (id);


--
-- TOC entry 3276 (class 2606 OID 78324)
-- Name: modelos PK_modelos; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.modelos
    ADD CONSTRAINT "PK_modelos" PRIMARY KEY (id);


--
-- TOC entry 3304 (class 2606 OID 78470)
-- Name: ordens_servicos PK_ordens_servicos; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ordens_servicos
    ADD CONSTRAINT "PK_ordens_servicos" PRIMARY KEY (manutencao_id, tecnico_id);


--
-- TOC entry 3306 (class 2606 OID 78489)
-- Name: papeis PK_papeis; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.papeis
    ADD CONSTRAINT "PK_papeis" PRIMARY KEY (id);


--
-- TOC entry 3289 (class 2606 OID 78437)
-- Name: pecas PK_pecas; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pecas
    ADD CONSTRAINT "PK_pecas" PRIMARY KEY (id);


--
-- TOC entry 3313 (class 2606 OID 78518)
-- Name: pecas_usadas PK_pecas_usadas; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pecas_usadas
    ADD CONSTRAINT "PK_pecas_usadas" PRIMARY KEY (peca_id, manutencao_id);


--
-- TOC entry 3299 (class 2606 OID 78465)
-- Name: tecnicos PK_tecnicos; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tecnicos
    ADD CONSTRAINT "PK_tecnicos" PRIMARY KEY (id);


--
-- TOC entry 3278 (class 1259 OID 78372)
-- Name: IX_equipamentos_categoria_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IX_equipamentos_categoria_id" ON public.equipamentos USING btree (categoria_id);


--
-- TOC entry 3279 (class 1259 OID 78373)
-- Name: IX_equipamentos_localizacao_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IX_equipamentos_localizacao_id" ON public.equipamentos USING btree (localizacao_id);


--
-- TOC entry 3280 (class 1259 OID 78374)
-- Name: IX_equipamentos_modelo_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IX_equipamentos_modelo_id" ON public.equipamentos USING btree (modelo_id);


--
-- TOC entry 3291 (class 1259 OID 78453)
-- Name: IX_equipamentos_pecas_peca_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IX_equipamentos_pecas_peca_id" ON public.equipamentos_pecas USING btree (peca_id);


--
-- TOC entry 3284 (class 1259 OID 78429)
-- Name: IX_manutencoes_equipamento_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IX_manutencoes_equipamento_id" ON public.manutencoes USING btree (equipamento_id);


--
-- TOC entry 3301 (class 1259 OID 78490)
-- Name: IX_ordens_servicos_papel_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IX_ordens_servicos_papel_id" ON public.ordens_servicos USING btree (papel_id);


--
-- TOC entry 3302 (class 1259 OID 78481)
-- Name: IX_ordens_servicos_tecnico_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IX_ordens_servicos_tecnico_id" ON public.ordens_servicos USING btree (tecnico_id);


--
-- TOC entry 3311 (class 1259 OID 78529)
-- Name: IX_pecas_usadas_manutencao_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IX_pecas_usadas_manutencao_id" ON public.pecas_usadas USING btree (manutencao_id);


--
-- TOC entry 3297 (class 1259 OID 78503)
-- Name: IX_tecnicos_especializacao_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "IX_tecnicos_especializacao_id" ON public.tecnicos USING btree (especializacao_id);


--
-- TOC entry 3287 (class 1259 OID 78575)
-- Name: idx_manutencoes_nome; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_manutencoes_nome ON public.manutencoes USING btree (nome);


--
-- TOC entry 3271 (class 1259 OID 78548)
-- Name: ix_categorias_nome; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_categorias_nome ON public.categorias USING btree (nome);


--
-- TOC entry 3283 (class 1259 OID 78547)
-- Name: ix_equipamentos_nome; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_equipamentos_nome ON public.equipamentos USING btree (nome);


--
-- TOC entry 3310 (class 1259 OID 78580)
-- Name: ix_especializacoes_nome; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_especializacoes_nome ON public.especializacoes USING btree (nome);


--
-- TOC entry 3296 (class 1259 OID 78579)
-- Name: ix_ferramentas_nome; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_ferramentas_nome ON public.ferramentas USING btree (nome);


--
-- TOC entry 3274 (class 1259 OID 78546)
-- Name: ix_localizacoes_nome; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_localizacoes_nome ON public.localizacoes USING btree (nome);


--
-- TOC entry 3277 (class 1259 OID 78545)
-- Name: ix_modelos_nome; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_modelos_nome ON public.modelos USING btree (nome);


--
-- TOC entry 3307 (class 1259 OID 78578)
-- Name: ix_papeis_nome; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_papeis_nome ON public.papeis USING btree (nome);


--
-- TOC entry 3290 (class 1259 OID 78577)
-- Name: ix_pecas_nome; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_pecas_nome ON public.pecas USING btree (nome);


--
-- TOC entry 3300 (class 1259 OID 78576)
-- Name: ix_tecnicos_nome; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_tecnicos_nome ON public.tecnicos USING btree (nome);


--
-- TOC entry 3326 (class 2620 OID 78700)
-- Name: manutencoes trg_atualizar_total_manutencoes_concluidas; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_atualizar_total_manutencoes_concluidas AFTER UPDATE ON public.manutencoes FOR EACH ROW EXECUTE FUNCTION public.atualizar_total_manutencoes_concluidas();


--
-- TOC entry 3314 (class 2606 OID 78530)
-- Name: equipamentos FK_equipamentos_categorias_categoria_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.equipamentos
    ADD CONSTRAINT "FK_equipamentos_categorias_categoria_id" FOREIGN KEY (categoria_id) REFERENCES public.categorias(id) ON DELETE SET NULL;


--
-- TOC entry 3315 (class 2606 OID 78535)
-- Name: equipamentos FK_equipamentos_localizacoes_localizacao_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.equipamentos
    ADD CONSTRAINT "FK_equipamentos_localizacoes_localizacao_id" FOREIGN KEY (localizacao_id) REFERENCES public.localizacoes(id) ON DELETE SET NULL;


--
-- TOC entry 3316 (class 2606 OID 78540)
-- Name: equipamentos FK_equipamentos_modelos_modelo_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.equipamentos
    ADD CONSTRAINT "FK_equipamentos_modelos_modelo_id" FOREIGN KEY (modelo_id) REFERENCES public.modelos(id) ON DELETE SET NULL;


--
-- TOC entry 3318 (class 2606 OID 78443)
-- Name: equipamentos_pecas FK_equipamentos_pecas_equipamentos_equipamento_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.equipamentos_pecas
    ADD CONSTRAINT "FK_equipamentos_pecas_equipamentos_equipamento_id" FOREIGN KEY (equipamento_id) REFERENCES public.equipamentos(id) ON DELETE RESTRICT;


--
-- TOC entry 3319 (class 2606 OID 78448)
-- Name: equipamentos_pecas FK_equipamentos_pecas_pecas_peca_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.equipamentos_pecas
    ADD CONSTRAINT "FK_equipamentos_pecas_pecas_peca_id" FOREIGN KEY (peca_id) REFERENCES public.pecas(id) ON DELETE RESTRICT;


--
-- TOC entry 3317 (class 2606 OID 78424)
-- Name: manutencoes FK_manutencoes_equipamentos_equipamento_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.manutencoes
    ADD CONSTRAINT "FK_manutencoes_equipamentos_equipamento_id" FOREIGN KEY (equipamento_id) REFERENCES public.equipamentos(id) ON DELETE CASCADE;


--
-- TOC entry 3321 (class 2606 OID 78471)
-- Name: ordens_servicos FK_ordens_servicos_manutencoes_manutencao_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ordens_servicos
    ADD CONSTRAINT "FK_ordens_servicos_manutencoes_manutencao_id" FOREIGN KEY (manutencao_id) REFERENCES public.manutencoes(id) ON DELETE RESTRICT;


--
-- TOC entry 3322 (class 2606 OID 78504)
-- Name: ordens_servicos FK_ordens_servicos_papeis_papel_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ordens_servicos
    ADD CONSTRAINT "FK_ordens_servicos_papeis_papel_id" FOREIGN KEY (papel_id) REFERENCES public.papeis(id) ON DELETE RESTRICT;


--
-- TOC entry 3323 (class 2606 OID 78476)
-- Name: ordens_servicos FK_ordens_servicos_tecnicos_tecnico_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ordens_servicos
    ADD CONSTRAINT "FK_ordens_servicos_tecnicos_tecnico_id" FOREIGN KEY (tecnico_id) REFERENCES public.tecnicos(id) ON DELETE RESTRICT;


--
-- TOC entry 3324 (class 2606 OID 78519)
-- Name: pecas_usadas FK_pecas_usadas_manutencoes_manutencao_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pecas_usadas
    ADD CONSTRAINT "FK_pecas_usadas_manutencoes_manutencao_id" FOREIGN KEY (manutencao_id) REFERENCES public.manutencoes(id) ON DELETE RESTRICT;


--
-- TOC entry 3325 (class 2606 OID 78524)
-- Name: pecas_usadas FK_pecas_usadas_pecas_peca_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pecas_usadas
    ADD CONSTRAINT "FK_pecas_usadas_pecas_peca_id" FOREIGN KEY (peca_id) REFERENCES public.pecas(id) ON DELETE RESTRICT;


--
-- TOC entry 3320 (class 2606 OID 78509)
-- Name: tecnicos FK_tecnicos_especializacoes_especializacao_id; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tecnicos
    ADD CONSTRAINT "FK_tecnicos_especializacoes_especializacao_id" FOREIGN KEY (especializacao_id) REFERENCES public.especializacoes(id) ON DELETE RESTRICT;


-- Completed on 2024-06-24 17:50:10

--
-- PostgreSQL database dump complete
--

