-- Downloaded from: https://github.com/StrangeGoofy/Yota_game/blob/9877bad67b39a3661fc74a08af21a86ddd3e52dc/dump.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 16.4
-- Dumped by pg_dump version 16.4

-- Started on 2025-03-18 01:12:38

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
-- TOC entry 2610 (class 2615 OID 35447)
-- Name: s314500; Type: SCHEMA; Schema: -; Owner: s314500
--

CREATE SCHEMA s314500;


ALTER SCHEMA s314500 OWNER TO s314500;

--
-- TOC entry 50491 (class 1255 OID 1371782)
-- Name: check_card_validity(integer, integer, integer, integer); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.check_card_validity(p_lobby integer, p_x integer, p_y integer, p_type integer) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE 
	x_shapes text;
	x_numbers smallint; 
	x_colors text;

	y_shapes text;
	y_numbers smallint; 
	y_colors text;

	x_count integer; 
	y_count integer;

	shape text;
	number smallint;
	color text;
BEGIN

	SELECT t.shape, t.number, t.color from cards c 
	JOIN cards_types t ON c.card_type_id = t.id
	WHERE c.card_id = p_type
	INTO shape, number, color;

	-- Если место уже занято
	IF EXISTS (SELECT 1 FROM get_cards_on_table(p_lobby) WHERE x = p_x AND y = p_y) THEN
		RAISE NOTICE 'Клетка уже занята';
		RETURN FALSE;
	END IF;

	-- Если не найдено соседних клеток
	IF NOT EXISTS (SELECT 1 FROM get_adjacent_cards(p_lobby, p_x, p_y)) THEN 
		RAISE NOTICE 'У клетки должен быть хоть один сосед';
		RETURN FALSE;
	END IF;

	SELECT COUNT (1) FROM get_adjacent_cards(p_lobby, p_x, p_y) a
	WHERE a.x = p_x 
	INTO x_count;

	IF x_count > 3 THEN
		RAISE NOTICE 'Слишком много соседей по горизонтали';
		RETURN FALSE;
	END IF;

	SELECT COUNT (1) FROM get_adjacent_cards(p_lobby, p_x, p_y) a
	WHERE a.y = p_y 
	INTO y_count;
	
	IF y_count > 3 THEN
		RAISE NOTICE 'Слишком много соседей по вертикали';
		RETURN FALSE;
	END IF;

	IF NOT (p_type IN (SELECT id FROM get_possible_cards(p_lobby, p_x, p_y))) THEN 
		RAISE NOTICE 'Карта не подходит';
		RETURN FALSE;
	END IF;

	RETURN TRUE;

END; 
$$;


ALTER FUNCTION s314500.check_card_validity(p_lobby integer, p_x integer, p_y integer, p_type integer) OWNER TO s314500;

--
-- TOC entry 51873 (class 1255 OID 1217449)
-- Name: checktoken(integer); Type: PROCEDURE; Schema: s314500; Owner: s314500
--

CREATE PROCEDURE s314500.checktoken(IN tk integer)
    LANGUAGE plpgsql
    AS $$
DECLARE
    tokenExists BOOLEAN;
BEGIN
    -- Проверка существования токена
    SELECT EXISTS (SELECT 1 FROM Tokens WHERE token = tk) INTO tokenExists;
    RAISE NOTICE 'isValid: %', tokenExists;
END;
$$;


ALTER PROCEDURE s314500.checktoken(IN tk integer) OWNER TO s314500;

--
-- TOC entry 54425 (class 1255 OID 1217448)
-- Name: cleartokens(); Type: PROCEDURE; Schema: s314500; Owner: s314500
--

CREATE PROCEDURE s314500.cleartokens()
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Удаление старых токенов (больше чем 7 дней)
    DELETE FROM Tokens WHERE created < NOW() - INTERVAL '7 days';
END;
$$;


ALTER PROCEDURE s314500.cleartokens() OWNER TO s314500;

--
-- TOC entry 54756 (class 1255 OID 1371779)
-- Name: createlobby(integer, character varying, integer); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.createlobby(tk integer, pw character varying, turnt integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    userLogin VARCHAR(64);
    userId INT;
    lobbyId INT;
BEGIN
    -- Получаем логин по токену
    SELECT login INTO userLogin FROM tokens WHERE token = tk;
    IF userLogin IS NULL THEN
        RAISE EXCEPTION 'Невалидный токен';
    END IF;

    -- Получаем userId
    SELECT id INTO userId FROM users WHERE login = userLogin;

    -- Создаем лобби
    INSERT INTO lobbies (password, turn_time, host_id, state)
    VALUES (pw, turnT, userId, 'Start')
    RETURNING id INTO lobbyId;

    -- Добавляем пользователя в лобби
    INSERT INTO players (id,login, lobby_id, is_ready)
    VALUES (userId, userLogin, lobbyId, true);

    RETURN lobbyId;
END;
$$;


ALTER FUNCTION s314500.createlobby(tk integer, pw character varying, turnt integer) OWNER TO s314500;

--
-- TOC entry 50936 (class 1255 OID 1371803)
-- Name: createlobby(integer, character varying, character varying, integer); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.createlobby(tk integer, p_nickname character varying, pw character varying, turnt integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    userLogin VARCHAR(64);
    userId INT;
    lobbyId INT;
BEGIN
    -- Получаем логин по токену
    SELECT login INTO userLogin FROM tokens WHERE token = tk;
    IF userLogin IS NULL THEN
        RAISE EXCEPTION 'Невалидный токен';
    END IF;

    -- Получаем userId
    SELECT id INTO userId FROM users WHERE login = userLogin;

    -- Создаем лобби
    INSERT INTO lobbies (password, turn_time, host_id, state)
    VALUES (pw, turnT, userId, 'Start')
    RETURNING id INTO lobbyId;

    -- Добавляем пользователя в лобби
    INSERT INTO players (login, nickname, lobby_id, is_ready)
    VALUES (userLogin, p_nickname, lobbyId, true);

    RETURN lobbyId;
END;
$$;


ALTER FUNCTION s314500.createlobby(tk integer, p_nickname character varying, pw character varying, turnt integer) OWNER TO s314500;

--
-- TOC entry 50376 (class 1255 OID 1217458)
-- Name: enterlobby(integer, integer, character varying); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.enterlobby(tk integer, lobbyid integer, inputpassword character varying) RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE
    userLogin VARCHAR(64);
    userId INT;
    actualPassword VARCHAR(10);
BEGIN
    -- Получаем логин по токену
    SELECT login INTO userLogin FROM tokens WHERE token = tk;
    IF userLogin IS NULL THEN
        RETURN 'Невалидный токен';
    END IF;

    -- Получаем userId
    SELECT id INTO userId FROM users WHERE login = userLogin;

    -- Проверка, что пользователь не в лобби
    IF EXISTS (SELECT 1 FROM players WHERE user_id = userId AND lobby_id = lobbyId) THEN
        RETURN 'Пользователь уже в лобби';
    END IF;

    -- Проверка на пароль
    SELECT password INTO actualPassword FROM lobbies WHERE id = lobbyId;
    IF actualPassword IS NOT NULL AND inputPassword != actualPassword THEN
        RETURN 'Неверный пароль';
    END IF;

    -- Проверка на максимальное количество игроков
    IF (SELECT COUNT(*) FROM players WHERE lobby_id = lobbyId) = 4 THEN
        RETURN 'Лобби полное';
    END IF;

    -- Вход в лобби
    INSERT INTO players (user_id, lobby_id) VALUES (userId, lobbyId);
    RETURN 'Вход в лобби выполнен';
END;
$$;


ALTER FUNCTION s314500.enterlobby(tk integer, lobbyid integer, inputpassword character varying) OWNER TO s314500;

--
-- TOC entry 52142 (class 1255 OID 1376656)
-- Name: get_adjacent_cards(integer, integer, integer); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.get_adjacent_cards(p_lobby integer, p_x integer, p_y integer) RETURNS TABLE(x smallint, y smallint, shape text, color text, number smallint)
    LANGUAGE plpgsql
    AS $$
BEGIN 

    RETURN QUERY 
    WITH RECURSIVE 
    left_x_neighbors AS (
        -- Находим ближайшее существующее значение x, которое меньше param
        (SELECT t.x, t.y, t.shape, t.color, t.number FROM get_cards_on_table(p_lobby) t WHERE t.x = p_x - 1 AND t.y = p_y)
        UNION ALL
        -- Добавляем предыдущее значение, если оно есть
        SELECT t.x, t.y, t.shape, t.color, t.number FROM get_cards_on_table(p_lobby) t
        JOIN left_x_neighbors n ON t.x = n.x - 1
        WHERE t.y = p_y
    ),
    
    right_x_neighbors AS (
        -- Находим ближайшее существующее значение x, которое больше param
        (SELECT t.x, t.y, t.shape, t.color, t.number FROM get_cards_on_table(p_lobby) t WHERE t.x = p_x + 1 AND t.y = p_y)
        UNION ALL
        -- Добавляем следующее значение, если оно есть
        SELECT t.x, t.y, t.shape, t.color, t.number FROM get_cards_on_table(p_lobby) t
        JOIN right_x_neighbors n ON t.x = n.x + 1
        WHERE t.y = p_y
    ),
    
    left_y_neighbors AS (
        -- Находим ближайшее существующее значение y, которое меньше param
        (SELECT t.x, t.y, t.shape, t.color, t.number FROM get_cards_on_table(p_lobby) t WHERE t.y = p_y - 1 AND t.x = p_x)
        UNION ALL
        -- Добавляем предыдущее значение, если оно есть
        SELECT t.x, t.y, t.shape, t.color, t.number FROM get_cards_on_table(p_lobby) t
        JOIN left_y_neighbors n ON t.y = n.y - 1
        WHERE t.x = p_x
    ),

    right_y_neighbors AS (
        -- Находим ближайшее существующее значение y, которое больше param
        (SELECT t.x, t.y, t.shape, t.color, t.number FROM get_cards_on_table(p_lobby) t WHERE t.y = p_y + 1 AND t.x = p_x)
        UNION ALL
        -- Добавляем следующее значение, если оно есть
        SELECT t.x, t.y, t.shape, t.color, t.number FROM get_cards_on_table(p_lobby) t
        JOIN right_y_neighbors n ON t.y = n.y + 1
        WHERE t.x = p_x
    )

    SELECT lx.x, lx.y, lx.shape, lx.color, lx.number FROM left_x_neighbors lx
    UNION 
    SELECT rx.x, rx.y, rx.shape, rx.color, rx.number FROM right_x_neighbors rx
    UNION 
    SELECT ly.x, ly.y, ly.shape, ly.color, ly.number FROM left_y_neighbors ly
    UNION 
    SELECT ry.x, ry.y, ry.shape, ry.color, ry.number FROM right_y_neighbors ry;

END 
$$;


ALTER FUNCTION s314500.get_adjacent_cards(p_lobby integer, p_x integer, p_y integer) OWNER TO s314500;

--
-- TOC entry 52953 (class 1255 OID 1371353)
-- Name: get_cards_on_table(integer); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.get_cards_on_table(lobby integer) RETURNS TABLE(x smallint, y smallint, shape text, color text, number smallint)
    LANGUAGE plpgsql
    AS $$BEGIN 

RETURN QUERY
SELECT pos.x, pos.y, c.shape, c.color, c.number FROM cards_on_table pos 
LEFT JOIN cards ON pos.card_id = cards.card_id 
LEFT JOIN cards_types c ON cards.card_id = c.id
WHERE pos.lobby_id = lobby;

END;$$;


ALTER FUNCTION s314500.get_cards_on_table(lobby integer) OWNER TO s314500;

--
-- TOC entry 50727 (class 1255 OID 1376860)
-- Name: get_possible_cards(integer, integer, integer); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.get_possible_cards(p_lobby integer, p_x integer, p_y integer) RETURNS TABLE(id integer, possible_color text, possible_shape text, possible_number smallint)
    LANGUAGE plpgsql
    AS $$
DECLARE 
	
	y_count integer;
	x_count integer;
	
BEGIN

SELECT COUNT(DISTINCT(x)) FROM get_adjacent_cards(p_lobby, p_x, p_y) 
WHERE y = p_y
INTO x_count;

SELECT COUNT(DISTINCT(y)) FROM get_adjacent_cards(p_lobby, p_x, p_y) 
WHERE x = p_x
INTO y_count;

RETURN QUERY
WITH
distinct_color_x AS 
(
	SELECT DISTINCT(c.color) FROM get_adjacent_cards(p_lobby, p_x, p_y) c 
	WHERE y = p_y
),
distinct_color_y AS
(
	SELECT DISTINCT(c.color) FROM get_adjacent_cards(p_lobby, p_x, p_y) c 
	WHERE x = p_x
),
possible_colors AS 
	(
		(
			(
			SELECT c.color FROM distinct_color_x c  
			WHERE (SELECT COUNT(1) FROM distinct_color_x) = 1
			)
			UNION
			(
			SELECT t.name as color FROM t_card_color t
			WHERE (SELECT COUNT(1) FROM distinct_color_x) = x_count
			EXCEPT SELECT color FROM distinct_color_x  
			)
		)
		INTERSECT
		(
			(
			SELECT c.color FROM distinct_color_y c  
			WHERE (SELECT COUNT(1) FROM distinct_color_y) = 1
			)
			UNION
			(
			SELECT t.name as color FROM t_card_color t
			WHERE (SELECT COUNT(1) FROM distinct_color_y) = y_count
			EXCEPT SELECT c.color FROM distinct_color_y c
			)
		)		
	),
distinct_shape_x AS 
	(
	SELECT DISTINCT(c.shape) FROM get_adjacent_cards(p_lobby, p_x, p_y) c 
	WHERE y = p_y
	),
distinct_shape_y AS
(
	SELECT DISTINCT(c.shape) FROM get_adjacent_cards(p_lobby, p_x, p_y) c 
	WHERE x = p_x
),
possible_shapes AS 
	(
		(
			(
			SELECT c.shape FROM distinct_shape_x c  
			WHERE (SELECT COUNT(1) FROM distinct_shape_x) = 1
			)
			UNION
			(
			SELECT t.name as shape FROM t_card_shape t
			WHERE (SELECT COUNT(1) FROM distinct_shape_x) = x_count
			EXCEPT SELECT shape FROM distinct_shape_x  
			)
		)
		INTERSECT
		(
			(
			SELECT c.shape FROM distinct_shape_y c  
			WHERE (SELECT COUNT(1) FROM distinct_shape_y) = 1
			)
			UNION
			(
			SELECT t.name as shape FROM t_card_shape t
			WHERE (SELECT COUNT(1) FROM distinct_shape_y) = y_count
			EXCEPT SELECT c.shape FROM distinct_shape_y c
			)
		)		
	),
distinct_number_x AS 
	(
	SELECT DISTINCT(c.number) FROM get_adjacent_cards(p_lobby, p_x, p_y) c 
	WHERE y = p_y
	),
distinct_number_y AS
(
	SELECT DISTINCT(c.number) FROM get_adjacent_cards(p_lobby, p_x, p_y) c 
	WHERE x = p_x
),
possible_numbers AS 
	(
		(
			(
			SELECT c.number FROM distinct_number_x c  
			WHERE (SELECT COUNT(1) FROM distinct_number_x) = 1
			)
			UNION
			(
			SELECT t.number as number FROM t_card_number t
			WHERE (SELECT COUNT(1) FROM distinct_number_x) = x_count
			EXCEPT SELECT number FROM distinct_number_x  
			)
		)
		INTERSECT
		(
			(
			SELECT c.number FROM distinct_number_y c  
			WHERE (SELECT COUNT(1) FROM distinct_number_y) = 1
			)
			UNION
			(
			SELECT t.number as number FROM t_card_number t
			WHERE (SELECT COUNT(1) FROM distinct_number_y) = y_count
			EXCEPT SELECT c.number FROM distinct_number_y c
			)
		)		
	)

	SELECT t.id, t.color, t.shape, t.number FROM cards_types t
	WHERE 
	t.color IN (SELECT color FROM possible_colors) AND 
	t.shape IN (SELECT shape FROM possible_shapes) AND 
	t.number IN (SELECT number FROM possible_numbers);

	-- SELECT 0, p.color, '', 0::smallint FROM possible_colors p;

END;
$$;


ALTER FUNCTION s314500.get_possible_cards(p_lobby integer, p_x integer, p_y integer) OWNER TO s314500;

--
-- TOC entry 53243 (class 1255 OID 1217452)
-- Name: getcurrentgames(integer); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.getcurrentgames(tk integer) RETURNS TABLE(id integer, usercount integer, hostlogin character varying)
    LANGUAGE plpgsql
    AS $$
DECLARE
    userLogin VARCHAR(64);
    userId INT;
BEGIN
    -- Получаем логин по токену
    SELECT login INTO userLogin FROM Tokens WHERE token = tk;
    IF userLogin IS NULL THEN
        RAISE EXCEPTION 'Невалидный токен';
    END IF;

    -- Получаем userId
    SELECT id INTO userId FROM Users WHERE login = userLogin;

    -- Запрос
    RETURN QUERY
    SELECT gl.id,
           COUNT(ul2.user_id) AS userCount,
           u.login AS hostLogin
    FROM GameLobbies AS gl
    JOIN UsersInLobby AS ul ON gl.id = ul.lobby_id
    LEFT JOIN UsersInLobby AS ul2 ON gl.id = ul2.lobby_id
    LEFT JOIN Users u ON gl.host_id = u.id
    WHERE ul.user_id = userId
    GROUP BY gl.id, u.login;
END;
$$;


ALTER FUNCTION s314500.getcurrentgames(tk integer) OWNER TO s314500;

--
-- TOC entry 54609 (class 1255 OID 1217454)
-- Name: gethost(integer); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.gethost(lobbyid integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN (SELECT host_id FROM lobbies WHERE id = lobbyid);
END;
$$;


ALTER FUNCTION s314500.gethost(lobbyid integer) OWNER TO s314500;

--
-- TOC entry 53004 (class 1255 OID 1371783)
-- Name: getlobbysettings(integer); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.getlobbysettings(lobbyid integer) RETURNS TABLE(haspassword boolean, turn_time integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        CASE WHEN password IS NOT NULL THEN TRUE ELSE FALSE END AS hasPassword,
        turn_time
    FROM lobbies
    WHERE id = lobbyId;
END;
$$;


ALTER FUNCTION s314500.getlobbysettings(lobbyid integer) OWNER TO s314500;

--
-- TOC entry 50629 (class 1255 OID 1217500)
-- Name: getuserid(integer); Type: PROCEDURE; Schema: s314500; Owner: s314500
--

CREATE PROCEDURE s314500.getuserid(IN tk integer)
    LANGUAGE plpgsql
    AS $$
DECLARE
    userLogin VARCHAR(64) := getUserLoginByToken(tk);
BEGIN
    -- Получение id пользователя по логину
    SELECT id FROM Users WHERE login = userLogin;
END;
$$;


ALTER PROCEDURE s314500.getuserid(IN tk integer) OWNER TO s314500;

--
-- TOC entry 52924 (class 1255 OID 1217450)
-- Name: getuserloginbytoken(integer); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.getuserloginbytoken(tk integer) RETURNS character varying
    LANGUAGE plpgsql
    AS $$
DECLARE
    userLogin VARCHAR(64);
BEGIN
    -- Получение логина по токену
    SELECT login INTO userLogin FROM tokens WHERE token = tk;
    RETURN userLogin;
END;
$$;


ALTER FUNCTION s314500.getuserloginbytoken(tk integer) OWNER TO s314500;

--
-- TOC entry 53948 (class 1255 OID 1217453)
-- Name: getusersinlobby(integer, integer); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.getusersinlobby(tk integer, lobbyid integer) RETURNS TABLE(user_id integer, login character varying, win_count integer, is_ready boolean)
    LANGUAGE plpgsql
    AS $$
DECLARE
    userLogin VARCHAR(64);
    userId INT;
BEGIN
    -- Получаем логин по токену
    SELECT login INTO userLogin FROM tokens WHERE token = tk;
    IF userLogin IS NULL THEN
        RAISE EXCEPTION 'Невалидный токен';
    END IF;

    -- Получаем userId
    SELECT id INTO userId FROM Users WHERE login = userLogin;

    -- Проверка на существование лобби
    IF NOT EXISTS (SELECT 1 FROM lobbies WHERE id = lobbyId) THEN
        RAISE EXCEPTION 'Лобби не существует';
    END IF;

    -- Проверка на то, что пользователь в лобби
    IF NOT EXISTS (SELECT 1 FROM players WHERE user_id = userId AND lobby_id = lobbyId) THEN
        RAISE EXCEPTION 'Пользователь не в лобби';
    END IF;

    -- Запрос
    RETURN QUERY
    SELECT u.id AS user_id, u.login, p.is_ready
    FROM players p
    JOIN users u ON p.user_id = u.id
    WHERE p.lobby_id = lobbyId;
END;
$$;


ALTER FUNCTION s314500.getusersinlobby(tk integer, lobbyid integer) OWNER TO s314500;

--
-- TOC entry 54217 (class 1255 OID 1217465)
-- Name: hashpassword(character varying); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.hashpassword(password character varying) RETURNS character varying
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Хэширование пароля с добавлением соли 'megasalt'
    RETURN encode(convert_to(CONCAT(password, 'megasalt'), 'UTF8'), 'hex');
END;
$$;


ALTER FUNCTION s314500.hashpassword(password character varying) OWNER TO s314500;

--
-- TOC entry 55466 (class 1255 OID 1217748)
-- Name: hello_world(); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.hello_world() RETURNS text
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN 'Hello, World!';
END;
$$;


ALTER FUNCTION s314500.hello_world() OWNER TO s314500;

--
-- TOC entry 53786 (class 1255 OID 1371335)
-- Name: initgame(integer); Type: PROCEDURE; Schema: s314500; Owner: s314500
--

CREATE PROCEDURE s314500.initgame(IN id_lobby integer)
    LANGUAGE plpgsql
    AS $$
DECLARE
    x integer;
BEGIN
    -- Вызов mashup
    CALL mashup(id_lobby);

    -- Получаем card_id
    SELECT card_id INTO x FROM Cards_in_deck LIMIT 1;

    -- Вставляем карту на стол
    INSERT INTO CardsOnTable 
    SELECT * FROM Cards_in_deck WHERE card_id = x;

    -- Удаляем карту из колоды
    DELETE FROM Cards_in_deck WHERE card_id = x;

    -- Создаём места в лобби
    CALL makePlaces(id_lobby, 4);

    -- Определяем случайного игрока, который начнёт ход
    INSERT INTO Current_Turn (player_id)
    SELECT player_id FROM Players 
    WHERE lobby_id = id_lobby 
    ORDER BY RANDOM() 
    LIMIT 1;
END;
$$;


ALTER PROCEDURE s314500.initgame(IN id_lobby integer) OWNER TO s314500;

--
-- TOC entry 54206 (class 1255 OID 1217457)
-- Name: isgamestarted(integer); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.isgamestarted(lobbyid integer) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM CurrentTurn
        WHERE turn_player_id IN (SELECT id FROM Players WHERE lobby_id = lobbyId)
    );
END;
$$;


ALTER FUNCTION s314500.isgamestarted(lobbyid integer) OWNER TO s314500;

--
-- TOC entry 53391 (class 1255 OID 1217460)
-- Name: leavelobby(integer, integer); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.leavelobby(tk integer, lobbyid integer) RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE
    userLogin VARCHAR(64);
    userId INT;
    currentHostId INT;
BEGIN
    -- Получаем логин по токену
    SELECT login INTO userLogin FROM tokens WHERE token = tk;
    IF userLogin IS NULL THEN
        RETURN 'Невалидный токен';
    END IF;

    -- Получаем userId
    SELECT id INTO userId FROM users WHERE login = userLogin;

    -- Проверка, является ли пользователь хостом
    SELECT host_id INTO currentHostId FROM lobbies WHERE id = lobbyId;
    IF currentHostId = userId THEN
        IF (SELECT COUNT(*) FROM players WHERE lobby_id = lobbyId) > 1 THEN
            UPDATE lobbies
            SET host_id = (
                SELECT user_id
                FROM players
                WHERE lobby_id = lobbyId
                AND user_id != userId
                ORDER BY RANDOM()
                LIMIT 1
            )
            WHERE id = lobbyId;
        ELSE
            DELETE FROM lobbies WHERE id = lobbyId;
        END IF;
    END IF;

    DELETE FROM players WHERE user_id = userId AND lobby_id = lobbyId;
    RETURN 'Выход из лобби выполнен';
END;
$$;


ALTER FUNCTION s314500.leavelobby(tk integer, lobbyid integer) OWNER TO s314500;

--
-- TOC entry 55024 (class 1255 OID 1217495)
-- Name: login(character varying, character varying); Type: PROCEDURE; Schema: s314500; Owner: s314500
--

CREATE PROCEDURE s314500.login(IN lg character varying, IN pw character varying)
    LANGUAGE plpgsql
    AS $$
DECLARE
    tk bigint;
BEGIN
    -- Генерация случайного токена
    tk := floor(random() * 4000000000) + 1;

    -- Проверка пароля
    IF hashPassword(pw) = (SELECT password FROM Users WHERE login = lg LIMIT 1) THEN
        -- Очистка токенов (предполагается, что процедура clearTokens() уже создана)
        CALL clearTokens();

        -- Вставка нового токена
        INSERT INTO Tokens (token, login) VALUES (tk, lg);

        -- Возвращаем id пользователя и токен

        PERFORM
    (SELECT id FROM Users WHERE login = lg LIMIT 1),
    (SELECT tk FROM Users WHERE login = lg LIMIT 1);
    ELSE
        -- Если логин или пароль неверный
        RAISE EXCEPTION 'Пароль или логин неверный';
    END IF;
END;
$$;


ALTER PROCEDURE s314500.login(IN lg character varying, IN pw character varying) OWNER TO s314500;

--
-- TOC entry 55754 (class 1255 OID 1217445)
-- Name: logout(integer); Type: PROCEDURE; Schema: s314500; Owner: s314500
--

CREATE PROCEDURE s314500.logout(IN tk integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Удаление токена
    DELETE FROM Tokens WHERE token = tk;

    -- Проверка, были ли удалены строки
    IF NOT FOUND THEN
        -- Если токен не был найден
        RAISE EXCEPTION 'Невалидный токен';
    ELSE
        -- Если токен был успешно удален
        RAISE NOTICE 'Вы успешно вышли из аккаунта';
    END IF;
END;
$$;


ALTER PROCEDURE s314500.logout(IN tk integer) OWNER TO s314500;

--
-- TOC entry 53158 (class 1255 OID 1332580)
-- Name: make_places(integer, integer); Type: PROCEDURE; Schema: s314500; Owner: s314500
--

CREATE PROCEDURE s314500.make_places(IN id_lobby integer, IN count_cards integer)
    LANGUAGE plpgsql
    AS $$
DECLARE
    total_cards INT;
    count_players INT;
BEGIN
    -- Создание временной таблицы для игроков
    CREATE TEMP TABLE tmp (
        n SERIAL PRIMARY KEY,
        id INT
    ) ON COMMIT DROP;

    -- Заполняем временную таблицу случайным порядком игроков из лобби
    INSERT INTO tmp(id)
    SELECT player_id FROM Players WHERE lobby_id = id_lobby ORDER BY RANDOM();

    -- Подсчитываем количество игроков
    SELECT COUNT(*) INTO count_players FROM tmp;

    -- Вычисляем общее количество карт
    total_cards := count_players * count_cards;

    -- Создание временной таблицы для карт
    CREATE TEMP TABLE tmpCards (
        n SERIAL PRIMARY KEY,
        id_card INT
    ) ON COMMIT DROP;

    -- Заполняем временную таблицу случайными картами из колоды, ограничивая по total_cards
    INSERT INTO tmpCards(id_card)
    SELECT card_id FROM Cards_in_Deck WHERE lobby_id = id_lobby ORDER BY RANDOM() LIMIT total_cards;

    -- Распределяем карты между игроками
    INSERT INTO Cards_in_hand (player_id, card_id)
    SELECT tmp.id AS player_id, tmpCards.id_card
    FROM tmpCards
    JOIN tmp ON tmp.n = (tmpCards.n % count_players) + 1;
	
END $$;


ALTER PROCEDURE s314500.make_places(IN id_lobby integer, IN count_cards integer) OWNER TO s314500;

--
-- TOC entry 52249 (class 1255 OID 1371336)
-- Name: makeplaces(integer, integer); Type: PROCEDURE; Schema: s314500; Owner: s314500
--

CREATE PROCEDURE s314500.makeplaces(IN id_lobby integer, IN count_cards integer)
    LANGUAGE plpgsql
    AS $$
DECLARE
    total_cards integer;
    count_players integer;
BEGIN
    -- Создание временной таблицы для игроков (случайный порядок)
    CREATE TEMP TABLE tmp (
        n SERIAL PRIMARY KEY, 
        id integer
    ) ON COMMIT DROP;

    -- Заполнение временной таблицы случайно отсортированными игроками
    INSERT INTO tmp(id)
    SELECT player_id FROM Players WHERE lobby_id = id_lobby ORDER BY RANDOM();

    -- Подсчёт количества игроков
    SELECT COUNT(*) INTO count_players FROM tmp;
    
    -- Вычисление общего количества карт
    total_cards := count_players * count_cards;

    CREATE TEMP TABLE tmpCards (
        n SERIAL PRIMARY KEY, 
        id_card integer
    ) ON COMMIT DROP;

    -- Заполнение картами (случайный порядок)
    INSERT INTO tmpCards(id_card)
    SELECT card_id FROM Cards_in_Deck WHERE lobby_id = id_lobby ORDER BY RANDOM() LIMIT total_cards;

    -- Раздача карт игрокам
    INSERT INTO Cards_in_hand (player_id, card_id)
    SELECT tmp.id AS player_id, tmpCards.id_card 
    FROM tmpCards 
    JOIN tmp ON tmp.n = (tmpCards.n % count_players) + 1;

END;
$$;


ALTER PROCEDURE s314500.makeplaces(IN id_lobby integer, IN count_cards integer) OWNER TO s314500;

--
-- TOC entry 54287 (class 1255 OID 1371334)
-- Name: mashup(integer); Type: PROCEDURE; Schema: s314500; Owner: s314500
--

CREATE PROCEDURE s314500.mashup(IN id_lobby integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Вставляем случайно отсортированные карты из cards в cards_in_tables
    INSERT INTO cards_on_table (card_id, lobby_id)
    SELECT card_id, lobby_id FROM cards
    ORDER BY RANDOM();
END;
$$;


ALTER PROCEDURE s314500.mashup(IN id_lobby integer) OWNER TO s314500;

--
-- TOC entry 50560 (class 1255 OID 1217466)
-- Name: register(character varying, character varying); Type: PROCEDURE; Schema: s314500; Owner: s314500
--

CREATE PROCEDURE s314500.register(IN login character varying, IN password character varying)
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Проверка на минимальную длину пароля и наличие как минимум одной буквы и одной цифры
    IF LENGTH(password) < 6 OR password !~ '[0-9]' OR password !~ '[a-zA-Z]' THEN
        RAISE EXCEPTION 'Пароль должен быть длиной не менее 6 символов и содержать как минимум одну букву и одну цифру';
    END IF;

    -- Вставка пользователя, если логин уникален
    BEGIN
        INSERT INTO Users(login, password) VALUES (login, hashPassword(password));
    EXCEPTION WHEN unique_violation THEN
        RAISE EXCEPTION 'Такой логин уже занят';
    END;

    -- Вызов функции для входа пользователя
    CALL login(login, password);
END;
$$;


ALTER PROCEDURE s314500.register(IN login character varying, IN password character varying) OWNER TO s314500;

--
-- TOC entry 50991 (class 1255 OID 1217459)
-- Name: setready(integer, integer, boolean); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.setready(tk integer, lobbyid integer, state boolean) RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE
    userLogin VARCHAR(64);
    userId INT;
BEGIN
    -- Получаем логин по токену
    SELECT login INTO userLogin FROM tokens WHERE token = tk;
    IF userLogin IS NULL THEN
        RETURN 'Невалидный токен';
    END IF;

    -- Получаем userId
    SELECT id INTO userId FROM users WHERE login = userLogin;

    -- Проверка, находится ли пользователь в лобби
    IF NOT EXISTS (SELECT 1 FROM players WHERE user_id = userId AND lobby_id = lobbyId) THEN
        RETURN 'Пользователь не в лобби';
    END IF;

    -- Обновление статуса готовности
    UPDATE players
    SET is_ready = state
    WHERE user_id = userId AND lobby_id = lobbyId;

    -- Возвращаем обновленную информацию о лобби
    PERFORM getUsersInLobby(tk, lobbyId);
    RETURN 'Готовность обновлена';
END;
$$;


ALTER FUNCTION s314500.setready(tk integer, lobbyid integer, state boolean) OWNER TO s314500;

--
-- TOC entry 51028 (class 1255 OID 1217451)
-- Name: showavailablegames(integer); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.showavailablegames(tk integer) RETURNS TABLE(id integer, usercount integer, haspassword boolean, hostlogin character varying)
    LANGUAGE plpgsql
    AS $$
DECLARE
    userLogin VARCHAR(64);
    userId INT;
BEGIN
    -- Получаем логин по токену
    SELECT login INTO userLogin FROM Tokens WHERE token = tk;
    IF userLogin IS NULL THEN
        RAISE EXCEPTION 'Невалидный токен';
    END IF;

    -- Получаем userId
    SELECT id INTO userId FROM Users WHERE login = userLogin;

    -- Запрос
    RETURN QUERY
    SELECT p.lobby_id AS id,
           COUNT(*) AS userCount,
           CASE WHEN l.password IS NOT NULL THEN TRUE ELSE FALSE END AS hasPassword,
           u.login AS hostLogin
    FROM players p
    LEFT JOIN players p2 ON p2.lobby_id = p.lobby_id AND p2.user_id = userId
    INNER JOIN lobbies l ON p.lobby_id = l.id
    LEFT JOIN users u ON l.host_id = u.id
    WHERE p2.user_id IS NULL
    GROUP BY p.lobby_id, l.password, u.login
    HAVING COUNT(*) < 4;
END;
$$;


ALTER FUNCTION s314500.showavailablegames(tk integer) OWNER TO s314500;

--
-- TOC entry 52005 (class 1255 OID 1217496)
-- Name: showuserinfo(integer); Type: PROCEDURE; Schema: s314500; Owner: s314500
--

CREATE PROCEDURE s314500.showuserinfo(IN tk integer)
    LANGUAGE plpgsql
    AS $$
DECLARE
    userLogin VARCHAR(64);
BEGIN
    -- Проверка на валидность токена
    SELECT login INTO userLogin FROM Tokens WHERE token = tk;

    IF userLogin IS NULL THEN
        -- Если токен не найден
        RAISE EXCEPTION 'Невалидный токен';
    ELSE
        -- Запрос информации о пользователе
        PERFORM (SELECT login FROM Users WHERE login = userLogin);
    END IF;
END;
$$;


ALTER PROCEDURE s314500.showuserinfo(IN tk integer) OWNER TO s314500;

--
-- TOC entry 50785 (class 1255 OID 1217499)
-- Name: showuserpl(); Type: PROCEDURE; Schema: s314500; Owner: s314500
--

CREATE PROCEDURE s314500.showuserpl()
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Запрос информации о всех пользователях
    PERFORM (SELECT login FROM Users);
END;
$$;


ALTER PROCEDURE s314500.showuserpl() OWNER TO s314500;

--
-- TOC entry 54430 (class 1255 OID 1217462)
-- Name: startgame(integer); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.startgame(lobbyid integer) RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE
    currentHostId INT;
BEGIN
    -- Получаем хоста лобби
    SELECT host_id INTO currentHostId FROM lobbies WHERE id = lobbyId;

    -- Проверка, что хост запускает игру
    IF currentHostId != (SELECT id FROM users WHERE login = userLogin) THEN
        RETURN 'Только хост может начать игру';
    END IF;

    -- Запуск игры
    UPDATE lobbies SET state = 'inProgress' WHERE id = lobbyId;
    
    RETURN 'Игра началась';
END;
$$;


ALTER FUNCTION s314500.startgame(lobbyid integer) OWNER TO s314500;

--
-- TOC entry 51443 (class 1255 OID 1376881)
-- Name: type_of_card(integer); Type: FUNCTION; Schema: s314500; Owner: s314500
--

CREATE FUNCTION s314500.type_of_card(p_card integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$DECLARE 

	out integer;

BEGIN 

SELECT card_type_id FROM cards WHERE card_id = p_card 
INTO out;

RETURN out;

END;$$;


ALTER FUNCTION s314500.type_of_card(p_card integer) OWNER TO s314500;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 47333 (class 1259 OID 1217335)
-- Name: cards; Type: TABLE; Schema: s314500; Owner: s314500
--

CREATE TABLE s314500.cards (
    card_id integer NOT NULL,
    card_type_id integer NOT NULL
);


ALTER TABLE s314500.cards OWNER TO s314500;

--
-- TOC entry 47332 (class 1259 OID 1217334)
-- Name: cards_card_id_seq; Type: SEQUENCE; Schema: s314500; Owner: s314500
--

CREATE SEQUENCE s314500.cards_card_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE s314500.cards_card_id_seq OWNER TO s314500;

--
-- TOC entry 99952 (class 0 OID 0)
-- Dependencies: 47332
-- Name: cards_card_id_seq; Type: SEQUENCE OWNED BY; Schema: s314500; Owner: s314500
--

ALTER SEQUENCE s314500.cards_card_id_seq OWNED BY s314500.cards.card_id;


--
-- TOC entry 47336 (class 1259 OID 1217378)
-- Name: cards_in_deck; Type: TABLE; Schema: s314500; Owner: s314500
--

CREATE TABLE s314500.cards_in_deck (
    card_id integer NOT NULL,
    lobby_id integer NOT NULL
);


ALTER TABLE s314500.cards_in_deck OWNER TO s314500;

--
-- TOC entry 47334 (class 1259 OID 1217346)
-- Name: cards_in_hand; Type: TABLE; Schema: s314500; Owner: s314500
--

CREATE TABLE s314500.cards_in_hand (
    card_id integer NOT NULL,
    player_id integer NOT NULL
);


ALTER TABLE s314500.cards_in_hand OWNER TO s314500;

--
-- TOC entry 47335 (class 1259 OID 1217361)
-- Name: cards_on_table; Type: TABLE; Schema: s314500; Owner: s314500
--

CREATE TABLE s314500.cards_on_table (
    card_id integer NOT NULL,
    lobby_id integer NOT NULL,
    x smallint NOT NULL,
    y smallint NOT NULL
);


ALTER TABLE s314500.cards_on_table OWNER TO s314500;

--
-- TOC entry 49114 (class 1259 OID 1330823)
-- Name: cardstypes_id_seq; Type: SEQUENCE; Schema: s314500; Owner: s314500
--

CREATE SEQUENCE s314500.cardstypes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 66
    CACHE 1;


ALTER SEQUENCE s314500.cardstypes_id_seq OWNER TO s314500;

--
-- TOC entry 49115 (class 1259 OID 1330824)
-- Name: cards_types; Type: TABLE; Schema: s314500; Owner: s314500
--

CREATE TABLE s314500.cards_types (
    id integer DEFAULT nextval('s314500.cardstypes_id_seq'::regclass) NOT NULL,
    shape text,
    number smallint,
    color text,
    CONSTRAINT cardstypes_color_check CHECK ((color = ANY (ARRAY['Blue'::text, 'Yellow'::text, 'Red'::text, 'Green'::text]))),
    CONSTRAINT cardstypes_number_check CHECK ((number >= 0)),
    CONSTRAINT cardstypes_shape_check CHECK ((shape = ANY (ARRAY['Square'::text, 'Triangle'::text, 'Circle'::text, 'Cross'::text])))
);


ALTER TABLE s314500.cards_types OWNER TO s314500;

--
-- TOC entry 47331 (class 1259 OID 1217310)
-- Name: current_turn; Type: TABLE; Schema: s314500; Owner: s314500
--

CREATE TABLE s314500.current_turn (
    player_id integer NOT NULL,
    start_time timestamp without time zone NOT NULL
);


ALTER TABLE s314500.current_turn OWNER TO s314500;

--
-- TOC entry 47327 (class 1259 OID 1217160)
-- Name: lobbies; Type: TABLE; Schema: s314500; Owner: s314500
--

CREATE TABLE s314500.lobbies (
    id integer NOT NULL,
    password character varying(10),
    turn_time integer DEFAULT 30 NOT NULL,
    host_id integer,
    state character varying DEFAULT 'waiting'::character varying NOT NULL
);


ALTER TABLE s314500.lobbies OWNER TO s314500;

--
-- TOC entry 47326 (class 1259 OID 1217159)
-- Name: lobbies_id_seq; Type: SEQUENCE; Schema: s314500; Owner: s314500
--

CREATE SEQUENCE s314500.lobbies_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE s314500.lobbies_id_seq OWNER TO s314500;

--
-- TOC entry 99953 (class 0 OID 0)
-- Dependencies: 47326
-- Name: lobbies_id_seq; Type: SEQUENCE OWNED BY; Schema: s314500; Owner: s314500
--

ALTER SEQUENCE s314500.lobbies_id_seq OWNED BY s314500.lobbies.id;


--
-- TOC entry 47329 (class 1259 OID 1217276)
-- Name: players; Type: TABLE; Schema: s314500; Owner: s314500
--

CREATE TABLE s314500.players (
    id integer NOT NULL,
    lobby_id integer NOT NULL,
    login character varying(20) NOT NULL,
    nickname character varying(15) NOT NULL,
    points integer DEFAULT 0 NOT NULL,
    is_ready boolean DEFAULT false NOT NULL
);


ALTER TABLE s314500.players OWNER TO s314500;

--
-- TOC entry 47328 (class 1259 OID 1217275)
-- Name: players_id_seq; Type: SEQUENCE; Schema: s314500; Owner: s314500
--

CREATE SEQUENCE s314500.players_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE s314500.players_id_seq OWNER TO s314500;

--
-- TOC entry 99954 (class 0 OID 0)
-- Dependencies: 47328
-- Name: players_id_seq; Type: SEQUENCE OWNED BY; Schema: s314500; Owner: s314500
--

ALTER SEQUENCE s314500.players_id_seq OWNED BY s314500.players.id;


--
-- TOC entry 50294 (class 1259 OID 1376837)
-- Name: t_card_color; Type: TABLE; Schema: s314500; Owner: s314500
--

CREATE TABLE s314500.t_card_color (
    id smallint NOT NULL,
    name text NOT NULL
);


ALTER TABLE s314500.t_card_color OWNER TO s314500;

--
-- TOC entry 50296 (class 1259 OID 1376851)
-- Name: t_card_number; Type: TABLE; Schema: s314500; Owner: s314500
--

CREATE TABLE s314500.t_card_number (
    id smallint NOT NULL,
    number smallint NOT NULL
);


ALTER TABLE s314500.t_card_number OWNER TO s314500;

--
-- TOC entry 50295 (class 1259 OID 1376844)
-- Name: t_card_shape; Type: TABLE; Schema: s314500; Owner: s314500
--

CREATE TABLE s314500.t_card_shape (
    id smallint NOT NULL,
    name text NOT NULL
);


ALTER TABLE s314500.t_card_shape OWNER TO s314500;

--
-- TOC entry 50102 (class 1259 OID 1371608)
-- Name: test_table; Type: TABLE; Schema: s314500; Owner: s314500
--

CREATE TABLE s314500.test_table (
    x integer NOT NULL
);


ALTER TABLE s314500.test_table OWNER TO s314500;

--
-- TOC entry 47337 (class 1259 OID 1217393)
-- Name: tokens; Type: TABLE; Schema: s314500; Owner: s314500
--

CREATE TABLE s314500.tokens (
    login character varying(64) NOT NULL,
    token bigint NOT NULL,
    created timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE s314500.tokens OWNER TO s314500;

--
-- TOC entry 47325 (class 1259 OID 1217154)
-- Name: users; Type: TABLE; Schema: s314500; Owner: s314500
--

CREATE TABLE s314500.users (
    login character varying(20) NOT NULL,
    password character varying(50) NOT NULL,
    id integer NOT NULL
);


ALTER TABLE s314500.users OWNER TO s314500;

--
-- TOC entry 47342 (class 1259 OID 1217480)
-- Name: users_id_seq; Type: SEQUENCE; Schema: s314500; Owner: s314500
--

CREATE SEQUENCE s314500.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE s314500.users_id_seq OWNER TO s314500;

--
-- TOC entry 99955 (class 0 OID 0)
-- Dependencies: 47342
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: s314500; Owner: s314500
--

ALTER SEQUENCE s314500.users_id_seq OWNED BY s314500.users.id;


--
-- TOC entry 47330 (class 1259 OID 1217295)
-- Name: users_in_lobby; Type: TABLE; Schema: s314500; Owner: s314500
--

CREATE TABLE s314500.users_in_lobby (
    lobby_id integer NOT NULL,
    login character varying(20) NOT NULL
);


ALTER TABLE s314500.users_in_lobby OWNER TO s314500;

--
-- TOC entry 99550 (class 2604 OID 1217338)
-- Name: cards card_id; Type: DEFAULT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.cards ALTER COLUMN card_id SET DEFAULT nextval('s314500.cards_card_id_seq'::regclass);


--
-- TOC entry 99544 (class 2604 OID 1217163)
-- Name: lobbies id; Type: DEFAULT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.lobbies ALTER COLUMN id SET DEFAULT nextval('s314500.lobbies_id_seq'::regclass);


--
-- TOC entry 99547 (class 2604 OID 1217279)
-- Name: players id; Type: DEFAULT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.players ALTER COLUMN id SET DEFAULT nextval('s314500.players_id_seq'::regclass);


--
-- TOC entry 99543 (class 2604 OID 1217481)
-- Name: users id; Type: DEFAULT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.users ALTER COLUMN id SET DEFAULT nextval('s314500.users_id_seq'::regclass);


--
-- TOC entry 99935 (class 0 OID 1217335)
-- Dependencies: 47333
-- Data for Name: cards; Type: TABLE DATA; Schema: s314500; Owner: s314500
--

COPY s314500.cards (card_id, card_type_id) FROM stdin;
1	1
2	2
3	3
4	4
5	5
6	6
7	7
8	8
9	9
10	10
11	11
12	12
13	13
14	14
15	15
16	16
17	17
18	18
19	19
20	20
21	21
22	22
23	23
24	24
25	25
26	26
27	27
28	28
29	29
30	30
31	31
32	32
33	33
34	34
35	35
36	36
37	37
38	38
39	39
40	40
41	41
42	42
43	43
44	44
45	45
46	46
47	47
48	48
49	49
50	50
51	51
52	52
53	53
54	54
55	55
56	56
57	57
58	58
59	59
60	60
61	61
62	62
63	63
64	64
65	65
66	66
\.


--
-- TOC entry 99938 (class 0 OID 1217378)
-- Dependencies: 47336
-- Data for Name: cards_in_deck; Type: TABLE DATA; Schema: s314500; Owner: s314500
--

COPY s314500.cards_in_deck (card_id, lobby_id) FROM stdin;
\.


--
-- TOC entry 99936 (class 0 OID 1217346)
-- Dependencies: 47334
-- Data for Name: cards_in_hand; Type: TABLE DATA; Schema: s314500; Owner: s314500
--

COPY s314500.cards_in_hand (card_id, player_id) FROM stdin;
\.


--
-- TOC entry 99937 (class 0 OID 1217361)
-- Dependencies: 47335
-- Data for Name: cards_on_table; Type: TABLE DATA; Schema: s314500; Owner: s314500
--

COPY s314500.cards_on_table (card_id, lobby_id, x, y) FROM stdin;
1	1	0	0
2	1	1	0
3	1	2	0
4	1	3	0
5	1	2	1
\.


--
-- TOC entry 99942 (class 0 OID 1330824)
-- Dependencies: 49115
-- Data for Name: cards_types; Type: TABLE DATA; Schema: s314500; Owner: s314500
--

COPY s314500.cards_types (id, shape, number, color) FROM stdin;
1	Square	1	Blue
2	Square	1	Yellow
3	Square	1	Red
4	Square	1	Green
5	Square	2	Blue
6	Square	2	Yellow
7	Square	2	Red
8	Square	2	Green
9	Square	3	Blue
10	Square	3	Yellow
11	Square	3	Red
12	Square	3	Green
13	Square	4	Blue
14	Square	4	Yellow
15	Square	4	Red
16	Square	4	Green
17	Triangle	1	Blue
18	Triangle	1	Yellow
19	Triangle	1	Red
20	Triangle	1	Green
21	Triangle	2	Blue
22	Triangle	2	Yellow
23	Triangle	2	Red
24	Triangle	2	Green
25	Triangle	3	Blue
26	Triangle	3	Yellow
27	Triangle	3	Red
28	Triangle	3	Green
29	Triangle	4	Blue
30	Triangle	4	Yellow
31	Triangle	4	Red
32	Triangle	4	Green
33	Circle	1	Blue
34	Circle	1	Yellow
35	Circle	1	Red
36	Circle	1	Green
37	Circle	2	Blue
38	Circle	2	Yellow
39	Circle	2	Red
40	Circle	2	Green
41	Circle	3	Blue
42	Circle	3	Yellow
43	Circle	3	Red
44	Circle	3	Green
45	Circle	4	Blue
46	Circle	4	Yellow
47	Circle	4	Red
48	Circle	4	Green
49	Cross	1	Blue
50	Cross	1	Yellow
51	Cross	1	Red
52	Cross	1	Green
53	Cross	2	Blue
54	Cross	2	Yellow
55	Cross	2	Red
56	Cross	2	Green
57	Cross	3	Blue
58	Cross	3	Yellow
59	Cross	3	Red
60	Cross	3	Green
61	Cross	4	Blue
62	Cross	4	Yellow
63	Cross	4	Red
64	Cross	4	Green
65	\N	\N	\N
66	\N	\N	\N
\.


--
-- TOC entry 99933 (class 0 OID 1217310)
-- Dependencies: 47331
-- Data for Name: current_turn; Type: TABLE DATA; Schema: s314500; Owner: s314500
--

COPY s314500.current_turn (player_id, start_time) FROM stdin;
\.


--
-- TOC entry 99929 (class 0 OID 1217160)
-- Dependencies: 47327
-- Data for Name: lobbies; Type: TABLE DATA; Schema: s314500; Owner: s314500
--

COPY s314500.lobbies (id, password, turn_time, host_id, state) FROM stdin;
1	12345678	60	\N	waiting
11	qwerty	60	8	Start
17	qwerty	60	8	Start
\.


--
-- TOC entry 99931 (class 0 OID 1217276)
-- Dependencies: 47329
-- Data for Name: players; Type: TABLE DATA; Schema: s314500; Owner: s314500
--

COPY s314500.players (id, lobby_id, login, nickname, points, is_ready) FROM stdin;
8	11	TestUser	Tester	0	t
1	17	TestUser	Tester	0	t
\.


--
-- TOC entry 99944 (class 0 OID 1376837)
-- Dependencies: 50294
-- Data for Name: t_card_color; Type: TABLE DATA; Schema: s314500; Owner: s314500
--

COPY s314500.t_card_color (id, name) FROM stdin;
1	Yellow
2	Green
3	Red
4	Blue
\.


--
-- TOC entry 99946 (class 0 OID 1376851)
-- Dependencies: 50296
-- Data for Name: t_card_number; Type: TABLE DATA; Schema: s314500; Owner: s314500
--

COPY s314500.t_card_number (id, number) FROM stdin;
1	1
2	2
3	3
4	4
\.


--
-- TOC entry 99945 (class 0 OID 1376844)
-- Dependencies: 50295
-- Data for Name: t_card_shape; Type: TABLE DATA; Schema: s314500; Owner: s314500
--

COPY s314500.t_card_shape (id, name) FROM stdin;
1	Circle
2	Triangle
3	Square
4	Cross
\.


--
-- TOC entry 99943 (class 0 OID 1371608)
-- Dependencies: 50102
-- Data for Name: test_table; Type: TABLE DATA; Schema: s314500; Owner: s314500
--

COPY s314500.test_table (x) FROM stdin;
1
2
3
4
5
6
8
9
12
13
14
21
22
23
\.


--
-- TOC entry 99939 (class 0 OID 1217393)
-- Dependencies: 47337
-- Data for Name: tokens; Type: TABLE DATA; Schema: s314500; Owner: s314500
--

COPY s314500.tokens (login, token, created) FROM stdin;
TestUser	1076758434	2025-03-16 22:53:09.790649
TestUser2	357352692	2025-03-16 22:54:17.350104
TestUser3	466720616	2025-03-16 22:57:47.070035
TestUser4	1154052900	2025-03-16 23:06:17.176119
\.


--
-- TOC entry 99927 (class 0 OID 1217154)
-- Dependencies: 47325
-- Data for Name: users; Type: TABLE DATA; Schema: s314500; Owner: s314500
--

COPY s314500.users (login, password, id) FROM stdin;
vasya	7661733132336d65676173616c74	5
TestUser	7177657274793132336d65676173616c74	8
TestUser2	7177657274793132336d65676173616c74	13
TestUser3	7177657274793132336d65676173616c74	21
TestUser4	717765727479313233346d65676173616c74	42
\.


--
-- TOC entry 99932 (class 0 OID 1217295)
-- Dependencies: 47330
-- Data for Name: users_in_lobby; Type: TABLE DATA; Schema: s314500; Owner: s314500
--

COPY s314500.users_in_lobby (lobby_id, login) FROM stdin;
\.


--
-- TOC entry 99956 (class 0 OID 0)
-- Dependencies: 47332
-- Name: cards_card_id_seq; Type: SEQUENCE SET; Schema: s314500; Owner: s314500
--

SELECT pg_catalog.setval('s314500.cards_card_id_seq', 1, false);


--
-- TOC entry 99957 (class 0 OID 0)
-- Dependencies: 49114
-- Name: cardstypes_id_seq; Type: SEQUENCE SET; Schema: s314500; Owner: s314500
--

SELECT pg_catalog.setval('s314500.cardstypes_id_seq', 1, false);


--
-- TOC entry 99958 (class 0 OID 0)
-- Dependencies: 47326
-- Name: lobbies_id_seq; Type: SEQUENCE SET; Schema: s314500; Owner: s314500
--

SELECT pg_catalog.setval('s314500.lobbies_id_seq', 19, true);


--
-- TOC entry 99959 (class 0 OID 0)
-- Dependencies: 47328
-- Name: players_id_seq; Type: SEQUENCE SET; Schema: s314500; Owner: s314500
--

SELECT pg_catalog.setval('s314500.players_id_seq', 3, true);


--
-- TOC entry 99960 (class 0 OID 0)
-- Dependencies: 47342
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: s314500; Owner: s314500
--

SELECT pg_catalog.setval('s314500.users_id_seq', 42, true);


--
-- TOC entry 99569 (class 2606 OID 1217340)
-- Name: cards cards_pkey; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.cards
    ADD CONSTRAINT cards_pkey PRIMARY KEY (card_id);


--
-- TOC entry 99577 (class 2606 OID 1217382)
-- Name: cards_in_deck cardsindeck_pkey; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.cards_in_deck
    ADD CONSTRAINT cardsindeck_pkey PRIMARY KEY (card_id);


--
-- TOC entry 99571 (class 2606 OID 1217350)
-- Name: cards_in_hand cardsinhand_pkey; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.cards_in_hand
    ADD CONSTRAINT cardsinhand_pkey PRIMARY KEY (card_id);


--
-- TOC entry 99573 (class 2606 OID 1217365)
-- Name: cards_on_table cardsontable_pkey; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.cards_on_table
    ADD CONSTRAINT cardsontable_pkey PRIMARY KEY (card_id);


--
-- TOC entry 99581 (class 2606 OID 1330834)
-- Name: cards_types cardstypes_pkey; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.cards_types
    ADD CONSTRAINT cardstypes_pkey PRIMARY KEY (id);


--
-- TOC entry 99567 (class 2606 OID 1217314)
-- Name: current_turn currentturn_pkey; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.current_turn
    ADD CONSTRAINT currentturn_pkey PRIMARY KEY (player_id);


--
-- TOC entry 99559 (class 2606 OID 1217166)
-- Name: lobbies lobbies_pkey; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.lobbies
    ADD CONSTRAINT lobbies_pkey PRIMARY KEY (id);


--
-- TOC entry 99561 (class 2606 OID 1217282)
-- Name: players players_pkey; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.players
    ADD CONSTRAINT players_pkey PRIMARY KEY (id);


--
-- TOC entry 99587 (class 2606 OID 1376843)
-- Name: t_card_color t_card_color_pkey; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.t_card_color
    ADD CONSTRAINT t_card_color_pkey PRIMARY KEY (id);


--
-- TOC entry 99591 (class 2606 OID 1376855)
-- Name: t_card_number t_card_number_pkey; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.t_card_number
    ADD CONSTRAINT t_card_number_pkey PRIMARY KEY (id);


--
-- TOC entry 99589 (class 2606 OID 1376850)
-- Name: t_card_shape t_card_shape_pkey; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.t_card_shape
    ADD CONSTRAINT t_card_shape_pkey PRIMARY KEY (id);


--
-- TOC entry 99585 (class 2606 OID 1371612)
-- Name: test_table test_table_pkey; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.test_table
    ADD CONSTRAINT test_table_pkey PRIMARY KEY (x);


--
-- TOC entry 99579 (class 2606 OID 1217487)
-- Name: tokens tokens_pkey; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.tokens
    ADD CONSTRAINT tokens_pkey PRIMARY KEY (token);


--
-- TOC entry 99575 (class 2606 OID 1217367)
-- Name: cards_on_table unique_lobby_xy; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.cards_on_table
    ADD CONSTRAINT unique_lobby_xy UNIQUE (lobby_id, x, y);


--
-- TOC entry 99563 (class 2606 OID 1217284)
-- Name: players unique_login_lobby; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.players
    ADD CONSTRAINT unique_login_lobby UNIQUE (login, lobby_id);


--
-- TOC entry 99583 (class 2606 OID 1330836)
-- Name: cards_types unique_shape_number_color; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.cards_types
    ADD CONSTRAINT unique_shape_number_color UNIQUE (shape, number, color);


--
-- TOC entry 99557 (class 2606 OID 1217158)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (login);


--
-- TOC entry 99565 (class 2606 OID 1217299)
-- Name: users_in_lobby usersinlobby_pkey; Type: CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.users_in_lobby
    ADD CONSTRAINT usersinlobby_pkey PRIMARY KEY (lobby_id, login);


--
-- TOC entry 99601 (class 2606 OID 1217383)
-- Name: cards_in_deck cardsindeck_card_id_fkey; Type: FK CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.cards_in_deck
    ADD CONSTRAINT cardsindeck_card_id_fkey FOREIGN KEY (card_id) REFERENCES s314500.cards(card_id) ON DELETE RESTRICT;


--
-- TOC entry 99602 (class 2606 OID 1217388)
-- Name: cards_in_deck cardsindeck_lobby_id_fkey; Type: FK CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.cards_in_deck
    ADD CONSTRAINT cardsindeck_lobby_id_fkey FOREIGN KEY (lobby_id) REFERENCES s314500.lobbies(id) ON DELETE CASCADE;


--
-- TOC entry 99597 (class 2606 OID 1217351)
-- Name: cards_in_hand cardsinhand_card_id_fkey; Type: FK CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.cards_in_hand
    ADD CONSTRAINT cardsinhand_card_id_fkey FOREIGN KEY (card_id) REFERENCES s314500.cards(card_id) ON DELETE RESTRICT;


--
-- TOC entry 99598 (class 2606 OID 1217356)
-- Name: cards_in_hand cardsinhand_player_id_fkey; Type: FK CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.cards_in_hand
    ADD CONSTRAINT cardsinhand_player_id_fkey FOREIGN KEY (player_id) REFERENCES s314500.players(id) ON DELETE CASCADE;


--
-- TOC entry 99599 (class 2606 OID 1217368)
-- Name: cards_on_table cardsontable_card_id_fkey; Type: FK CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.cards_on_table
    ADD CONSTRAINT cardsontable_card_id_fkey FOREIGN KEY (card_id) REFERENCES s314500.cards(card_id) ON DELETE RESTRICT;


--
-- TOC entry 99600 (class 2606 OID 1217373)
-- Name: cards_on_table cardsontable_lobby_id_fkey; Type: FK CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.cards_on_table
    ADD CONSTRAINT cardsontable_lobby_id_fkey FOREIGN KEY (lobby_id) REFERENCES s314500.lobbies(id) ON DELETE CASCADE;


--
-- TOC entry 99596 (class 2606 OID 1217315)
-- Name: current_turn currentturn_player_id_fkey; Type: FK CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.current_turn
    ADD CONSTRAINT currentturn_player_id_fkey FOREIGN KEY (player_id) REFERENCES s314500.players(id) ON DELETE CASCADE;


--
-- TOC entry 99592 (class 2606 OID 1217290)
-- Name: players players_lobby_id_fkey; Type: FK CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.players
    ADD CONSTRAINT players_lobby_id_fkey FOREIGN KEY (lobby_id) REFERENCES s314500.lobbies(id) ON DELETE CASCADE;


--
-- TOC entry 99593 (class 2606 OID 1217285)
-- Name: players players_login_fkey; Type: FK CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.players
    ADD CONSTRAINT players_login_fkey FOREIGN KEY (login) REFERENCES s314500.users(login) ON DELETE RESTRICT;


--
-- TOC entry 99603 (class 2606 OID 1217399)
-- Name: tokens tokens_login_fkey; Type: FK CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.tokens
    ADD CONSTRAINT tokens_login_fkey FOREIGN KEY (login) REFERENCES s314500.users(login) ON DELETE CASCADE;


--
-- TOC entry 99594 (class 2606 OID 1217300)
-- Name: users_in_lobby usersinlobby_lobby_id_fkey; Type: FK CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.users_in_lobby
    ADD CONSTRAINT usersinlobby_lobby_id_fkey FOREIGN KEY (lobby_id) REFERENCES s314500.lobbies(id) ON DELETE CASCADE;


--
-- TOC entry 99595 (class 2606 OID 1217305)
-- Name: users_in_lobby usersinlobby_login_fkey; Type: FK CONSTRAINT; Schema: s314500; Owner: s314500
--

ALTER TABLE ONLY s314500.users_in_lobby
    ADD CONSTRAINT usersinlobby_login_fkey FOREIGN KEY (login) REFERENCES s314500.users(login) ON DELETE CASCADE;


-- Completed on 2025-03-18 01:12:42

--
-- PostgreSQL database dump complete
--

