-- Downloaded from: https://github.com/paulshriner/fcc-rd-cert/blob/f6c9c03198f143fd6d3e8aa229d11927cf2edbe1/celestial-bodies/universe.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 12.17 (Ubuntu 12.17-1.pgdg22.04+1)
-- Dumped by pg_dump version 12.17 (Ubuntu 12.17-1.pgdg22.04+1)

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

DROP DATABASE universe;
--
-- Name: universe; Type: DATABASE; Schema: -; Owner: freecodecamp
--

CREATE DATABASE universe WITH TEMPLATE = template0 ENCODING = 'UTF8' LC_COLLATE = 'C.UTF-8' LC_CTYPE = 'C.UTF-8';


ALTER DATABASE universe OWNER TO freecodecamp;

\connect universe

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
-- Name: asteroids; Type: TABLE; Schema: public; Owner: freecodecamp
--

CREATE TABLE public.asteroids (
    asteroids_id integer NOT NULL,
    name character varying,
    description text NOT NULL,
    length real NOT NULL
);


ALTER TABLE public.asteroids OWNER TO freecodecamp;

--
-- Name: asteroids_asteroids_id_seq; Type: SEQUENCE; Schema: public; Owner: freecodecamp
--

CREATE SEQUENCE public.asteroids_asteroids_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.asteroids_asteroids_id_seq OWNER TO freecodecamp;

--
-- Name: asteroids_asteroids_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: freecodecamp
--

ALTER SEQUENCE public.asteroids_asteroids_id_seq OWNED BY public.asteroids.asteroids_id;


--
-- Name: galaxy; Type: TABLE; Schema: public; Owner: freecodecamp
--

CREATE TABLE public.galaxy (
    galaxy_id integer NOT NULL,
    name character varying,
    description text,
    has_stars boolean NOT NULL,
    has_planets boolean NOT NULL
);


ALTER TABLE public.galaxy OWNER TO freecodecamp;

--
-- Name: galaxy_galaxy_id_seq; Type: SEQUENCE; Schema: public; Owner: freecodecamp
--

CREATE SEQUENCE public.galaxy_galaxy_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.galaxy_galaxy_id_seq OWNER TO freecodecamp;

--
-- Name: galaxy_galaxy_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: freecodecamp
--

ALTER SEQUENCE public.galaxy_galaxy_id_seq OWNED BY public.galaxy.galaxy_id;


--
-- Name: moon; Type: TABLE; Schema: public; Owner: freecodecamp
--

CREATE TABLE public.moon (
    moon_id integer NOT NULL,
    name character varying,
    description text,
    distance_from_planet integer NOT NULL,
    is_spherical boolean NOT NULL,
    planet_id integer NOT NULL
);


ALTER TABLE public.moon OWNER TO freecodecamp;

--
-- Name: moon_moon_id_seq; Type: SEQUENCE; Schema: public; Owner: freecodecamp
--

CREATE SEQUENCE public.moon_moon_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.moon_moon_id_seq OWNER TO freecodecamp;

--
-- Name: moon_moon_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: freecodecamp
--

ALTER SEQUENCE public.moon_moon_id_seq OWNED BY public.moon.moon_id;


--
-- Name: moon_planet_id_seq; Type: SEQUENCE; Schema: public; Owner: freecodecamp
--

CREATE SEQUENCE public.moon_planet_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.moon_planet_id_seq OWNER TO freecodecamp;

--
-- Name: moon_planet_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: freecodecamp
--

ALTER SEQUENCE public.moon_planet_id_seq OWNED BY public.moon.planet_id;


--
-- Name: planet; Type: TABLE; Schema: public; Owner: freecodecamp
--

CREATE TABLE public.planet (
    planet_id integer NOT NULL,
    name character varying,
    description text,
    has_life boolean NOT NULL,
    distance_from_earth integer NOT NULL,
    gravity numeric,
    star_id integer NOT NULL
);


ALTER TABLE public.planet OWNER TO freecodecamp;

--
-- Name: planet_planet_id_seq; Type: SEQUENCE; Schema: public; Owner: freecodecamp
--

CREATE SEQUENCE public.planet_planet_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.planet_planet_id_seq OWNER TO freecodecamp;

--
-- Name: planet_planet_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: freecodecamp
--

ALTER SEQUENCE public.planet_planet_id_seq OWNED BY public.planet.planet_id;


--
-- Name: planet_star_id_seq; Type: SEQUENCE; Schema: public; Owner: freecodecamp
--

CREATE SEQUENCE public.planet_star_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.planet_star_id_seq OWNER TO freecodecamp;

--
-- Name: planet_star_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: freecodecamp
--

ALTER SEQUENCE public.planet_star_id_seq OWNED BY public.planet.star_id;


--
-- Name: star; Type: TABLE; Schema: public; Owner: freecodecamp
--

CREATE TABLE public.star (
    star_id integer NOT NULL,
    name character varying,
    description text,
    length numeric NOT NULL,
    is_spherical boolean NOT NULL,
    galaxy_id integer NOT NULL
);


ALTER TABLE public.star OWNER TO freecodecamp;

--
-- Name: star_galaxy_id_seq; Type: SEQUENCE; Schema: public; Owner: freecodecamp
--

CREATE SEQUENCE public.star_galaxy_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.star_galaxy_id_seq OWNER TO freecodecamp;

--
-- Name: star_galaxy_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: freecodecamp
--

ALTER SEQUENCE public.star_galaxy_id_seq OWNED BY public.star.galaxy_id;


--
-- Name: star_star_id_seq; Type: SEQUENCE; Schema: public; Owner: freecodecamp
--

CREATE SEQUENCE public.star_star_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.star_star_id_seq OWNER TO freecodecamp;

--
-- Name: star_star_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: freecodecamp
--

ALTER SEQUENCE public.star_star_id_seq OWNED BY public.star.star_id;


--
-- Name: asteroids asteroids_id; Type: DEFAULT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.asteroids ALTER COLUMN asteroids_id SET DEFAULT nextval('public.asteroids_asteroids_id_seq'::regclass);


--
-- Name: galaxy galaxy_id; Type: DEFAULT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.galaxy ALTER COLUMN galaxy_id SET DEFAULT nextval('public.galaxy_galaxy_id_seq'::regclass);


--
-- Name: moon moon_id; Type: DEFAULT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.moon ALTER COLUMN moon_id SET DEFAULT nextval('public.moon_moon_id_seq'::regclass);


--
-- Name: moon planet_id; Type: DEFAULT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.moon ALTER COLUMN planet_id SET DEFAULT nextval('public.moon_planet_id_seq'::regclass);


--
-- Name: planet planet_id; Type: DEFAULT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.planet ALTER COLUMN planet_id SET DEFAULT nextval('public.planet_planet_id_seq'::regclass);


--
-- Name: planet star_id; Type: DEFAULT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.planet ALTER COLUMN star_id SET DEFAULT nextval('public.planet_star_id_seq'::regclass);


--
-- Name: star star_id; Type: DEFAULT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.star ALTER COLUMN star_id SET DEFAULT nextval('public.star_star_id_seq'::regclass);


--
-- Name: star galaxy_id; Type: DEFAULT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.star ALTER COLUMN galaxy_id SET DEFAULT nextval('public.star_galaxy_id_seq'::regclass);


--
-- Data for Name: asteroids; Type: TABLE DATA; Schema: public; Owner: freecodecamp
--

INSERT INTO public.asteroids VALUES (1, '16 Psyche', '16 Psyche is named after the Greek goddess of soul. The 16 in its name refers to the fact that it was the sixteenth asteroid discovered.', 220);
INSERT INTO public.asteroids VALUES (2, 'Dimorphos', 'Dimorphos is part of a binary system of asteroids. It orbits around a slightly larger asteroid named Didymos and has a near-Earth orbit.', 0.177);
INSERT INTO public.asteroids VALUES (3, 'Bennu', 'Bennu, or more accurately, 101955 Bennu is one of the most important asteroids out there.', 0.49);


--
-- Data for Name: galaxy; Type: TABLE DATA; Schema: public; Owner: freecodecamp
--

INSERT INTO public.galaxy VALUES (1, 'Milky Way', 'The Milky Way is the galaxy that includes the Solar System, with the name describing the galaxy''s appearance from Earth.', true, true);
INSERT INTO public.galaxy VALUES (2, 'Andromeda', 'The Andromeda Galaxy is a barred spiral galaxy and is the nearest major galaxy to the Milky Way.', true, true);
INSERT INTO public.galaxy VALUES (3, 'Antennae', 'The Antennae Galaxies are a pair of interacting galaxies in the constellation Corvus.', true, true);
INSERT INTO public.galaxy VALUES (4, 'NGC 4622', 'NGC 4622 is a face-on unbarred spiral galaxy with a very prominent ring structure located in the constellation Centaurus.', true, true);
INSERT INTO public.galaxy VALUES (5, 'NGC 6822', 'NGC 6822 is a barred irregular galaxy approximately 1.6 million light-years away in the constellation Sagittarius.', true, true);
INSERT INTO public.galaxy VALUES (6, 'NGC 2537', 'NGC 2537 is a blue compact dwarf galaxy in the constellation Lynx, located around 3 degrees NNW of 31 Lyncis.', true, true);


--
-- Data for Name: moon; Type: TABLE DATA; Schema: public; Owner: freecodecamp
--

INSERT INTO public.moon VALUES (1, 'Moon', 'The Moon is Earth''s only natural satellite.', 384399, true, 3);
INSERT INTO public.moon VALUES (2, 'Phobos', 'Phobos is the innermost and larger of the two natural satellites of Mars.', 6000, false, 4);
INSERT INTO public.moon VALUES (3, 'Deimos', 'Deimos is the smaller and outer of the two natural satellites of Mars.', 23460, false, 4);
INSERT INTO public.moon VALUES (4, 'Metis', 'Metis is Jupiter''s closest moon.', 128000, false, 5);
INSERT INTO public.moon VALUES (5, 'Adrastea', 'Adrastea is Jupiter''s second closest moon.', 129000, false, 5);
INSERT INTO public.moon VALUES (6, 'Amalthea', 'Amalthea is Jupiter''s third closest moon.', 181400, false, 5);
INSERT INTO public.moon VALUES (7, 'Thebe', 'Thebe is Jupiter''s fourth closest moon.', 221900, false, 5);
INSERT INTO public.moon VALUES (8, 'Pan', 'Pan is the innermost named moon of Saturn.', 133600, false, 6);
INSERT INTO public.moon VALUES (9, 'Daphnis', 'Daphnis is an inner satellite of Saturn.', 136500, false, 6);
INSERT INTO public.moon VALUES (10, 'Atlas', 'Atlas is an inner satellite of Saturn.', 137700, false, 6);
INSERT INTO public.moon VALUES (11, 'Prometheus', 'Prometheus is an inner satellite of Saturn.', 139400, false, 6);
INSERT INTO public.moon VALUES (12, 'Cordelia', 'Cordelia is the innermost known moon of Uranus.', 49800, false, 7);
INSERT INTO public.moon VALUES (13, 'Ophelia', 'Ophelia is a moon of Uranus.', 53800, false, 7);
INSERT INTO public.moon VALUES (14, 'Bianca', 'Bianca is an inner satellite of Uranus.', 59200, false, 7);
INSERT INTO public.moon VALUES (15, 'Cressida', 'Cressida is an inner satellite of Uranus.', 61800, false, 7);
INSERT INTO public.moon VALUES (16, 'Naiad', 'Naiad is the innermost satellite of Neptune and the nearest to the center of any gas giant.', 48224, false, 8);
INSERT INTO public.moon VALUES (17, 'Thalassa', 'Thalassa is the second-innermost satellite of Neptune.', 50074, false, 8);
INSERT INTO public.moon VALUES (18, 'Charon', 'Charon is the largest known natural satellite of Pluto.', 19570, true, 9);
INSERT INTO public.moon VALUES (19, 'Styx', 'Styx is a small natural satellite of Pluto.', 42000, false, 9);
INSERT INTO public.moon VALUES (20, 'Nix', 'Nix is a natural satellite of Pluto.', 48708, false, 9);


--
-- Data for Name: planet; Type: TABLE DATA; Schema: public; Owner: freecodecamp
--

INSERT INTO public.planet VALUES (1, 'Mercury', 'Mercury is the closest planet to the Sun.', false, 91691000, 3.72, 1);
INSERT INTO public.planet VALUES (2, 'Venus', 'Venus is the second planet from the Sun and the third brightest object in Earth''s sky.', false, 41400000, 8.82, 1);
INSERT INTO public.planet VALUES (3, 'Earth', 'Earth is the third planet from the Sun and largest of the terrestrial planets.', true, 0, 9.8, 1);
INSERT INTO public.planet VALUES (4, 'Mars', 'Mars is the fourth planet from the Sun and last of the terrestrial planets.', false, 78340000, 3.63, 1);
INSERT INTO public.planet VALUES (5, 'Jupiter', 'Named after the Roman king of the gods, Jupiter is fitting of its name.', false, 550390000, 24.79, 1);
INSERT INTO public.planet VALUES (6, 'Saturn', 'Saturn is the sixth planet from the Sun and second largest planet of the Solar System in terms of diameter and mass.', false, 119666000, 10.44, 1);
INSERT INTO public.planet VALUES (7, 'Uranus', 'Uranus, named after the the father of the Roman god Saturn.', false, 264561000, 8.69, 1);
INSERT INTO public.planet VALUES (8, 'Neptune', 'Neptune is the eighth planet from the Sun and last of the known planets.', false, 427306000, 11.15, 1);
INSERT INTO public.planet VALUES (9, 'Pluto', 'Pluto is a dwarf planet in the Solar System.', false, 590000000, 0.658, 1);
INSERT INTO public.planet VALUES (10, 'Ceres', 'Discovered in 1801, it was considered a planet for a year.', false, 239356593, 0.284, 2);
INSERT INTO public.planet VALUES (11, 'Haumea', 'Haumea is the fastest rotating dwarf planet with the most interesting/controversial shape.', false, 766000000, 0.441, 3);
INSERT INTO public.planet VALUES (12, 'Eris', 'Eris is the most distant dwarf planet, located beyond the orbit of Neptune.', false, 142000000, 0.89, 4);


--
-- Data for Name: star; Type: TABLE DATA; Schema: public; Owner: freecodecamp
--

INSERT INTO public.star VALUES (1, 'Sun', 'The Sun is the star at the centre of the Solar System.', 1392684, true, 1);
INSERT INTO public.star VALUES (2, 'Alpha Centauri', 'Alpha Centauri is a triple star system in the southern constellation of Centaurus.', 78200000000000, true, 1);
INSERT INTO public.star VALUES (3, 'Sigma Octantis', 'Sigma Octantis is a solitary star in the Octans constellation that forms the pole star of the Southern Hemisphere.', 6110400, true, 1);
INSERT INTO public.star VALUES (4, 'Mira', 'Mira is a red-giant star.', 231072000, true, 1);
INSERT INTO public.star VALUES (5, 'Sirius', 'Sirius is the brightest star in the night sky.', 1192728, true, 1);
INSERT INTO public.star VALUES (6, 'Barnard''s Star', 'Barnard''s Star is a small red dwarf star in the constellation of Ophiuchus.', 130012, true, 1);


--
-- Name: asteroids_asteroids_id_seq; Type: SEQUENCE SET; Schema: public; Owner: freecodecamp
--

SELECT pg_catalog.setval('public.asteroids_asteroids_id_seq', 3, true);


--
-- Name: galaxy_galaxy_id_seq; Type: SEQUENCE SET; Schema: public; Owner: freecodecamp
--

SELECT pg_catalog.setval('public.galaxy_galaxy_id_seq', 6, true);


--
-- Name: moon_moon_id_seq; Type: SEQUENCE SET; Schema: public; Owner: freecodecamp
--

SELECT pg_catalog.setval('public.moon_moon_id_seq', 20, true);


--
-- Name: moon_planet_id_seq; Type: SEQUENCE SET; Schema: public; Owner: freecodecamp
--

SELECT pg_catalog.setval('public.moon_planet_id_seq', 1, false);


--
-- Name: planet_planet_id_seq; Type: SEQUENCE SET; Schema: public; Owner: freecodecamp
--

SELECT pg_catalog.setval('public.planet_planet_id_seq', 12, true);


--
-- Name: planet_star_id_seq; Type: SEQUENCE SET; Schema: public; Owner: freecodecamp
--

SELECT pg_catalog.setval('public.planet_star_id_seq', 1, false);


--
-- Name: star_galaxy_id_seq; Type: SEQUENCE SET; Schema: public; Owner: freecodecamp
--

SELECT pg_catalog.setval('public.star_galaxy_id_seq', 1, false);


--
-- Name: star_star_id_seq; Type: SEQUENCE SET; Schema: public; Owner: freecodecamp
--

SELECT pg_catalog.setval('public.star_star_id_seq', 6, true);


--
-- Name: asteroids asteroids_name_key; Type: CONSTRAINT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.asteroids
    ADD CONSTRAINT asteroids_name_key UNIQUE (name);


--
-- Name: asteroids asteroids_pkey; Type: CONSTRAINT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.asteroids
    ADD CONSTRAINT asteroids_pkey PRIMARY KEY (asteroids_id);


--
-- Name: galaxy galaxy_name_key; Type: CONSTRAINT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.galaxy
    ADD CONSTRAINT galaxy_name_key UNIQUE (name);


--
-- Name: galaxy galaxy_pkey; Type: CONSTRAINT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.galaxy
    ADD CONSTRAINT galaxy_pkey PRIMARY KEY (galaxy_id);


--
-- Name: moon moon_name_key; Type: CONSTRAINT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.moon
    ADD CONSTRAINT moon_name_key UNIQUE (name);


--
-- Name: moon moon_pkey; Type: CONSTRAINT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.moon
    ADD CONSTRAINT moon_pkey PRIMARY KEY (moon_id);


--
-- Name: planet planet_name_key; Type: CONSTRAINT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.planet
    ADD CONSTRAINT planet_name_key UNIQUE (name);


--
-- Name: planet planet_pkey; Type: CONSTRAINT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.planet
    ADD CONSTRAINT planet_pkey PRIMARY KEY (planet_id);


--
-- Name: star star_name_key; Type: CONSTRAINT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.star
    ADD CONSTRAINT star_name_key UNIQUE (name);


--
-- Name: star star_pkey; Type: CONSTRAINT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.star
    ADD CONSTRAINT star_pkey PRIMARY KEY (star_id);


--
-- Name: moon moon_planet_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.moon
    ADD CONSTRAINT moon_planet_id_fkey FOREIGN KEY (planet_id) REFERENCES public.planet(planet_id);


--
-- Name: planet planet_star_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.planet
    ADD CONSTRAINT planet_star_id_fkey FOREIGN KEY (star_id) REFERENCES public.star(star_id);


--
-- Name: star star_galaxy_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: freecodecamp
--

ALTER TABLE ONLY public.star
    ADD CONSTRAINT star_galaxy_id_fkey FOREIGN KEY (galaxy_id) REFERENCES public.galaxy(galaxy_id);


--
-- PostgreSQL database dump complete
--

