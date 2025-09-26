-- Downloaded from: https://github.com/SoniaMarrocco/Library-Database/blob/64479330d19881cd3c3faadc1ceb1b87f89e95f4/bookDBdump.sql
--
-- PostgreSQL database dump
--

-- Dumped from database version 17.0
-- Dumped by pg_dump version 17.0

-- Started on 2024-11-22 18:28:24 EST

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 229 (class 1255 OID 24758)
-- Name: validate_isbn10(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.validate_isbn10() RETURNS trigger
    LANGUAGE plpgsql
    AS $_$
BEGIN
    IF NEW.isbn10 !~ '^[0-9]{9}[0-9X]$' THEN
        RAISE EXCEPTION 'Invalid ISBN-10 format. Must be 10 digits or 9 digits followed by X.';
    END IF;
    RETURN NEW;
END;
$_$;


ALTER FUNCTION public.validate_isbn10() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 224 (class 1259 OID 24728)
-- Name: edition; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.edition (
    isbn10 character(10) NOT NULL,
    wid character varying(20) NOT NULL,
    publish_date character varying(50)
);


ALTER TABLE public.edition OWNER TO postgres;

--
-- TOC entry 223 (class 1259 OID 24715)
-- Name: rating; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.rating (
    rid integer NOT NULL,
    wid character varying(20) NOT NULL,
    cnt integer DEFAULT 0,
    avg_rating numeric(3,2) DEFAULT NULL::numeric
);


ALTER TABLE public.rating OWNER TO postgres;

--
-- TOC entry 218 (class 1259 OID 24679)
-- Name: work; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.work (
    wid character varying(20) NOT NULL,
    title character varying(255) NOT NULL,
    first_publish_date character varying(50)
);


ALTER TABLE public.work OWNER TO postgres;

--
-- TOC entry 227 (class 1259 OID 24773)
-- Name: admin_view; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.admin_view AS
 SELECT w.wid,
    w.title,
    w.first_publish_date,
    r.avg_rating,
    r.cnt,
    e.isbn10
   FROM ((public.work w
     LEFT JOIN public.rating r ON (((w.wid)::text = (r.wid)::text)))
     LEFT JOIN public.edition e ON (((w.wid)::text = (e.wid)::text)));


ALTER VIEW public.admin_view OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 24674)
-- Name: authors; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.authors (
    aid character varying(20) NOT NULL,
    name character varying(255) NOT NULL
);


ALTER TABLE public.authors OWNER TO postgres;

--
-- TOC entry 221 (class 1259 OID 24700)
-- Name: bio; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.bio (
    aid character varying(20) NOT NULL,
    bid integer NOT NULL,
    b_text text,
    source character varying(10) NOT NULL
);


ALTER TABLE public.bio OWNER TO postgres;

--
-- TOC entry 220 (class 1259 OID 24699)
-- Name: bio_bid_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.bio ALTER COLUMN bid ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.bio_bid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 225 (class 1259 OID 24738)
-- Name: digital; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.digital (
    isbn10 character(10) NOT NULL,
    form character varying(30)
);


ALTER TABLE public.digital OWNER TO postgres;

--
-- TOC entry 226 (class 1259 OID 24748)
-- Name: physical; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.physical (
    isbn10 character(10) NOT NULL,
    type character varying(30)
);


ALTER TABLE public.physical OWNER TO postgres;

--
-- TOC entry 222 (class 1259 OID 24714)
-- Name: rating_rid_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.rating_rid_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.rating_rid_seq OWNER TO postgres;

--
-- TOC entry 3682 (class 0 OID 0)
-- Dependencies: 222
-- Name: rating_rid_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.rating_rid_seq OWNED BY public.rating.rid;


--
-- TOC entry 228 (class 1259 OID 24777)
-- Name: user_view; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.user_view AS
 SELECT w.title,
    w.first_publish_date,
    r.avg_rating
   FROM ((public.work w
     LEFT JOIN public.edition e ON (((w.wid)::text = (e.wid)::text)))
     LEFT JOIN public.rating r ON (((e.wid)::text = (r.wid)::text)));


ALTER VIEW public.user_view OWNER TO postgres;

--
-- TOC entry 219 (class 1259 OID 24684)
-- Name: work_authors; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.work_authors (
    wid character varying(20) NOT NULL,
    aid character varying(20) NOT NULL
);


ALTER TABLE public.work_authors OWNER TO postgres;

--
-- TOC entry 3488 (class 2604 OID 24718)
-- Name: rating rid; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.rating ALTER COLUMN rid SET DEFAULT nextval('public.rating_rid_seq'::regclass);


--
-- TOC entry 3667 (class 0 OID 24674)
-- Dependencies: 217
-- Data for Name: authors; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.authors (aid, name) FROM stdin;
OL39307A	Dan Brown
OL19981A	Stephen King
OL229268A	Ken Follett
OL39329A	John Grisham
OL28257A	Michael Crichton
OL2032671A	Thomas Harris
OL575390A	Peter Benchley
OL34047A	Daphne du Maurier
OL709121A	Maj Sjöwall
OL1460559A	Per Wahlöö
OL8654905A	Ona Rius Piqué
OL2783112A	Tom Weiner
OL7342508A	Martin Lexell
OL9182880A	Elda García
OL33909A	P. D. James
OL40671A	James Ellroy
OL1433006A	Gillian Flynn
OL22258A	James Patterson
OL5113109A	Garet Rogers
OL25714A	Nelson De Mille
OL28165A	David Baldacci
OL228669A	Michael Ledwidge
OL10370207A	Michael Ledwidg
OL896513A	Justin Cronin
OL7945315A	Erik Singer
OL1515334A	Howard Roughan
OL6474824A	John Fowles
OL2631878A	Andrew Gross
OL2101074A	John le Carré
OL22019A	V. C. Andrews
OL2632116A	Carlos Ruiz Zafón
OL3123066A	Copyright Paperback Collection (Library of Congress) Staff
OL3353071A	Brunonia Barry
OL765158A	Maxine Paetro
OL539397A	Eric Ambler
OL34328A	Lee Child
OL31818A	Joy Fielding
OL25277A	Tom Clancy
OL20243A	Graham Greene
OL22261A	Sandra Brown
OL28577A	Patricia Highsmith
OL713259A	Desmond Bagley
OL3288848A	David Ellis
OL6925017A	Hammond Innes
OL2660362A	Tana French
OL76547A	Martin Cruz Smith
OL7442916A	A. J. Finn
OL45364A	Henning Mankell
OL2631898A	Joe Hill
OL1391085A	Stephenie Meyer
OL1390877A	Denise Mina
OL1372814A	Marshall Karp
OL26369A	Sarah Willson
OL7287124A	Fiona Barton
OL23073A	Hans Ulrich Treichel
OL444013A	Pete Dexter
OL39821A	Harlan Coben
OL7871059A	January LaVoy
OL2757233A	Mark Sullivan
OL1433920A	Blake Crouch
OL7566140A	Max DiLallo
OL30714A	Linda Howard
OL29079A	Clive Cussler
OL2660314A	Dirk Cussler
\.


--
-- TOC entry 3671 (class 0 OID 24700)
-- Dependencies: 221
-- Data for Name: bio; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.bio (aid, bid, b_text, source) FROM stdin;
OL39307A	1	Dan Brown is an American author of thriller fiction, best known for the 2003 bestselling novel, *The Da Vinci Code*. Brown's novels, which are treasure hunts set in a 24-hour time period, feature the recurring themes of cryptography, keys, symbols, codes, and conspiracy theories. His books have been translated into over 40 languages, and as of 2009, sold over 80 million copies.\r\n\r\nBrown's novels that feature the lead character Robert Langdon also include historical themes and Christianity as recurring motifs, and as a result, have generated controversy. Brown states on his website that his books are not anti-Christian, though he is on a 'constant spiritual journey' himself, and says of his book *The Da Vinci Code* that it is simply "an entertaining story that promotes spiritual discussion and debate" and suggests that the book may be used "as a positive catalyst for introspection and exploration of our faith."\r\n\r\n([Source][1])\r\n\r\n\r\n  [1]: http://en.wikipedia.org/wiki/Dan_Brown	OpenLib
OL39307A	2	Daniel Gerhard Brown (born June 22, 1964) is an American author best known for his thriller novels, including the Robert Langdon novels Angels & Demons (2000), The Da Vinci Code (2003), The Lost Symbol (2009), Inferno (2013), and Origin (2017). His novels are treasure hunts that usually take place over a period of 24 hours. They feature recurring themes of cryptography, art, and conspiracy theories. His books have been translated into 57 languages and, as of 2012, have sold over 200 million copies. Three of them, Angels & Demons, The Da Vinci Code, and Inferno, have been adapted into films, while one of them, The Lost Symbol, was adapted into a television series.\nThe Robert Langdon novels are deeply engaged with Christian themes and historical fiction, and have generated controversy as a result. Brown states on his website that his books are not anti-Christian and he is on a "constant spiritual journey" himself. He states that his book The Da Vinci Code is "an entertaining story that promotes spiritual discussion and debate" and suggests that the book may be used "as a positive catalyst for introspection and exploration of our faith."	Wikipedia
OL19981A	7	Stephen Edwin King (born September 21, 1947) is an American author of horror, supernatural fiction, suspense, crime, science-fiction, and fantasy novels. His books have sold more than 350 million copies, and many have been adapted into films, television series, miniseries, and comic books. King has published 63 novels, including seven under the pen name Richard Bachman, and five non-fiction books. He has also written approximately 200 short stories, most of which have been published in book collections.\r\n\r\nKing has received Bram Stoker Awards, World Fantasy Awards, and British Fantasy Society Awards. In 2003, the National Book Foundation awarded him the Medal for Distinguished Contribution to American Letters. He has also received awards for his contribution to literature for his entire bibliography, such as the 2004 World Fantasy Award for Life Achievement and the 2007 Grand Master Award from the Mystery Writers of America. In 2015, he was awarded with a National Medal of Arts from the U.S. National Endowment for the Arts for his contributions to literature. He has been described as the "King of Horror", a play on his surname and a reference to his high standing in pop culture.	OpenLib
OL19981A	8	Stephen Edwin King (born September 21, 1947) is an American author. Widely known for his horror novels, he has been crowned the "King of Horror". He has also explored other genres, among them suspense, crime, science-fiction, fantasy and mystery. Though known primarily for his novels, he has written approximately 200 short stories, most of which have been published in collections. \nHis debut, Carrie (1974), established him in horror. Different Seasons (1982), a collection of four novellas, was his first major departure from the genre. Among the films adapted from King's fiction are Carrie (1976), The Shining (1980), The Dead Zone (1983), Christine (1983), Stand by Me (1986), Misery (1990), The Shawshank Redemption (1994), Dolores Claiborne (1995), The Green Mile (1999), The Mist (2007) and It (2017). He has published under the pseudonym Richard Bachman and has co-written works with other authors, notably his friend Peter Straub and sons Joe Hill and Owen King. He has also written nonfiction, notably Danse Macabre (1981) and On Writing: A Memoir of the Craft (2000).\nAmong other awards, King has won the O. Henry Award for "The Man in the Black Suit" (1994) and the Los Angeles Times Book Prize for Mystery/Thriller for 11/22/63 (2011). He has also won honors for his overall contributions to literature, including the 2003 Medal for Distinguished Contribution to American Letters, the 2007 Grand Master Award from the Mystery Writers of America and the 2014 National Medal of Arts. Joyce Carol Oates called King "a brilliantly rooted, psychologically 'realistic' writer for whom the American scene has been a continuous source of inspiration, and American popular culture a vast cornucopia of possibilities."	Wikipedia
OL229268A	9	Ken Follett is a British author of thrillers and historical novels. He has sold more then 100 million copies of his works. Four of his books have reached the number 1 ranking on the New York Times best-seller list : *The Key to Rebecca, Lie Down with Lions, Triple* and *World Without End*.	OpenLib
OL229268A	10	Kenneth Martin Follett (born 5 June 1949) is a Welsh author of thrillers and historical novels who has sold more than 160 million copies of his works.\nFollett's commercial breakthrough came with the spy thriller Eye of the Needle (1978). After writing more best-sellers in the genre in the 1980s, he branched into historical fiction with The Pillars of the Earth (1989), an epic set in medieval England which became his best-known work and the first published in the Kingsbridge series. He has continued to write in both genres, including the Century Trilogy. Many of his books have achieved high ranking on bestseller lists, including the number-one position on the New York Times Best Seller list.	Wikipedia
OL39329A	25	Long before his name became synonymous with the modern legal thriller, he was working 60-70 hours a week at a small Southaven, Mississippi, law practice, squeezing in time before going to the office and during courtroom recesses to work on his hobby—writing his first novel.\r\n\r\nBorn on February 8, 1955 in Jonesboro, Arkansas, to a construction worker and a homemaker, John Grisham as a child dreamed of being a professional baseball player. Realizing he didn’t have the right stuff for a pro career, he shifted gears and majored in accounting at Mississippi State University. After graduating from law school at Ole Miss in 1981, he went on to practice law for nearly a decade in Southaven, specializing in criminal defense and personal injury litigation. In 1983, he was elected to the state House of Representatives and served until 1990.\r\n\r\nOne day at the DeSoto County courthouse, Grisham overheard the harrowing testimony of a twelve-year-old rape victim and was inspired to start a novel exploring what would have happened if the girl’s father had murdered her assailants. Getting up at 5 a.m. every day to get in several hours of writing time before heading off to work, Grisham spent three years on A Time to Kill and finished it in 1987. Initially rejected by many publishers, it was eventually bought by Wynwood Press, who gave it a modest 5,000 copy printing and published it in June 1988.\r\n\r\nThat might have put an end to Grisham’s hobby. However, he had already begun his next book, and it would quickly turn that hobby into a new full-time career—and spark one of publishing’s greatest success stories. The day after Grisham completed A Time to Kill, he began work on another novel, the story of a hotshot young attorney lured to an apparently perfect law firm that was not what it appeared. When he sold the film rights to The Firm to Paramount Pictures for $600,000, Grisham suddenly became a hot property among publishers, and book rights were bought by Doubleday. Spending 47 weeks on The New York Times bestseller list, The Firm became the bestselling novel of 1991.\r\n\r\nThe successes of The Pelican Brief, which hit number one on the New York Times bestseller list, and The Client, which debuted at number one, confirmed Grisham’s reputation as the master of the legal thriller. Grisham’s success even renewed interest in A Time to Kill, which was republished in hardcover by Doubleday and then in paperback by Dell. This time around, it was a bestseller.\r\n\r\nSince first publishing A Time to Kill in 1988, Grisham has written one novel a year (his other books are The Firm, The Pelican Brief, The Client, The Chamber, The Rainmaker, The Runaway Jury, The Partner, The Street Lawyer, The Testament, The Brethren, A Painted House, Skipping Christmas, The Summons, The King of Torts, Bleachers, The Last Juror, The Broker, Playing for Pizza, The Appeal, and The Associate) and all of them have become international bestsellers. There are currently over 250 million John Grisham books in print worldwide, which have been translated into 29 languages. Nine of his novels have been turned into films (The Firm, The Pelican Brief, The Client, A Time to Kill, The Rainmaker, The Chamber, A Painted House, The Runaway Jury, and Skipping Christmas), as was an original screenplay, The Gingerbread Man. The Innocent Man (October 2006) marked his first foray into non-fiction, and Ford County (November 2009) was his first short story collection.\r\n\r\nGrisham lives with his wife Renee and their two children Ty and Shea. The family splits their time between their Victorian home on a farm in Mississippi and a plantation near Charlottesville, VA.\r\n\r\nGrisham took time off from writing for several months in 1996 to return, after a five-year hiatus, to the courtroom. He was honoring a commitment made before he had retired from the law to become a full-time writer: representing the family of a railroad brakeman killed when he was pinned between two cars. Preparing his case with the same passion and dedication as his books’ protagonists, Grisham successfully argued his clients’ case, earning them a jury award of $683,500—the biggest verdict of his career.\r\n\r\nWhen he’s not writing, Grisham devotes time to charitable causes, including most recently his Rebuild The Coast Fund, which raised 8.8 million dollars for Gulf Coast relief in the wake of Hurricane Katrina. He also keeps up with his greatest passion: baseball. The man who dreamed of being a professional baseball player now serves as the local Little League commissioner. The six ballfields he built on his property have played host to over 350 kids on 26 Little League teams.	OpenLib
OL39329A	26	John Ray Grisham Jr. (; born February 8, 1955) is an American novelist, lawyer, and former member of the Mississippi House of Representatives, known for his best-selling legal thrillers. According to the American Academy of Achievement, Grisham has written 37 consecutive number-one fiction bestsellers, and his books have sold 300 million copies worldwide. Along with Tom Clancy and J. K. Rowling, Grisham is one of only three anglophone authors to have sold two million copies on the first printing.\nGrisham graduated from Mississippi State University and earned a Juris Doctor from the University of Mississippi School of Law in 1981. He practiced criminal law for about a decade and served in the Mississippi House of Representatives from 1983 to 1990. Grisham's first novel, A Time to Kill, was published in June 1989, four years after he began writing it. It was later adapted into the 1996 feature film of the same name. Grisham's first bestseller, The Firm, sold more than seven million copies, and was also adapted into a 1993 feature film of the same name, starring Tom Cruise, and a 2012 TV series that continues the story ten years after the events of the film and novel. Seven of his other novels have also been adapted into films: The Chamber, The Client, A Painted House, The Pelican Brief, The Rainmaker, The Runaway Jury, and Skipping Christmas.	Wikipedia
OL28257A	27	An American writer and filmmaker.	OpenLib
OL28257A	28	John Michael Crichton (; October 23, 1942 – November 4, 2008) was an American author, screenwriter and filmmaker. His books have sold over 200 million copies worldwide, and over a dozen have been adapted into films. His literary works heavily feature technology and are usually within the science fiction, techno-thriller, and medical fiction genres. Crichton's novels often explore human technological advancement and attempted dominance over nature, both with frequently catastrophic results; many of his works are cautionary tales, especially regarding themes of biotechnology. Several of his stories center specifically around themes of genetic modification, hybridization, paleontology and/or zoology. Many feature medical or scientific underpinnings, reflective of his own medical training and scientific background.\nCrichton received an M.D. from Harvard Medical School in 1969 but did not practice medicine, choosing to focus on his writing instead. Initially writing under a pseudonym, he eventually wrote 26 novels, including: The Andromeda Strain (1969), The Terminal Man (1972), The Great Train Robbery (1975), Congo (1980), Sphere (1987), Jurassic Park (1990), Rising Sun (1992), Disclosure (1994), The Lost World (1995), Airframe (1996), Timeline (1999), Prey (2002), State of Fear (2004), and Next (2006). Several novels, in various states of completion, were published after his death in 2008.\nCrichton was also involved in the film and television industry. In 1973, he wrote and directed Westworld, the first film to use 2D computer-generated imagery. He also directed Coma (1978), The First Great Train Robbery (1978), Looker (1981), and Runaway (1984). He was the creator of the television series ER (1994–2009), and several of his novels were adapted into films, most notably the Jurassic Park franchise.	Wikipedia
OL2032671A	39	William Thomas Harris III (born September 22, 1940) is an American writer. He is the author of a series of suspense novels about Hannibal Lecter. The majority of his works have been adapted into films and television, including The Silence of the Lambs, which became only the third film in Academy Awards history to sweep the Oscars in all of the five major categories.\nHis novels have sold more than 50 million copies, with The Silence of the Lambs alone selling 10 million copies, as of 2019.	Wikipedia
OL575390A	40	Peter Bradford Benchley (May 8, 1940 – February 11, 2006) was an American author. He is best known for his bestselling novel Jaws and co-wrote its movie adaptation with Carl Gottlieb. Several more of his works were also adapted for both cinema and television, including The Deep, The Island, Beast, and White Shark.\nLater in life, Benchley expressed some regret for his writing about sharks, which he felt indulged already present fear and false belief about sharks, and he became an advocate for marine conservation.  Contrary to widespread rumor, Benchley did not believe that his writings contributed to shark depopulation, nor is there evidence that Jaws or any of his works did so.	Wikipedia
OL34047A	41	Daphne du Maurier was born on 13 May 1907 in London, England, United Kingdom, the second of three daughters of Muriel Beaumont, an actress and maternal niece of William Comyns Beaumont, and Sir Gerald du Maurier, the prominent actor-manager, son of the author and Punch cartoonist George du Maurier, who created the character of Svengali in the novel Trilby. She was also the cousin of the Llewelyn Davies boys, who served as J.M. Barrie's inspiration for the characters in the play Peter Pan, or The Boy Who Wouldn't Grow Up. As a young child, she met many of the brightest stars of the theatre, thanks to the celebrity of her father. These connections helped her in establishing her literary career, and she published some of her early stories in Beaumont's Bystander magazine. Her first novel, The Loving Spirit, was published in 1931, and she continued writing successfull gothic novels in addition to biographies and other non-fiction books. Alfred Hitchcock was a fan of her novels and short stories, and adapted some of these to films: Jamaica Inn (1939), Rebecca (1940), and The Birds (1963). Other of her works adapted were Frenchman's Creek (1942), Hungry Hill (1943), My Cousin Rachel (1951), and "Don't Look Now" (1973). She was named a Dame of the British Empire.\r\n\r\nIn 1932, she married Frederick "Boy" Browning, with whom she had three children, Tessa, Flavia and Christian. Her husband died in 1965, and she passed away on 19 April 1989 in Fowey, Cornwall. After her death, it was revealed that she was bisexual.	OpenLib
OL34047A	42	Dame Daphne du Maurier, Lady Browning,  (; 13 May 1907 – 19 April 1989) was an English novelist, biographer and playwright. Her parents were actor-manager Sir Gerald du Maurier and his wife, actress Muriel Beaumont. Her grandfather George du Maurier was a writer and cartoonist.\nAlthough du Maurier is classed as a romantic novelist, her stories have been described as "moody and resonant" with overtones of the paranormal. Her bestselling works were not at first taken seriously by critics, but they have since earned an enduring reputation for narrative craft. Many have been successfully adapted into films, including the novels Rebecca, Frenchman's Creek, My Cousin Rachel and Jamaica Inn, and the short stories "The Birds" and "Don't Look Now". Du Maurier spent much of her life in Cornwall, where most of her works are set. As her fame increased, she became more reclusive.	Wikipedia
OL709121A	43	Swedish writer. Together with her husband [Per Wahlöö](/a/OL1460559A), author of crime novels about detective Martin Beck, using the signature Sjöwall/Wahlöö.	OpenLib
OL709121A	44	Maj Sjöwall (Swedish pronunciation: [maj ˈɧø̂ːval]; 25 September 1935 – 29 April 2020) was a Swedish author and translator. She is best known for her books about police detective Martin Beck. She wrote the books in collaborative work with her partner Per Wahlöö.	Wikipedia
OL1460559A	45	Swedish writer. Together with his wife [Maj Sjöwall](/a/OL709121A), author of crime novels about detective Martin Beck, using the signature Sjöwall/Wahlöö.	OpenLib
OL1460559A	46	Per Fredrik Wahlöö (5 August 1926 – 22 June 1975) – in English translations often identified as Peter Wahloo – was a Swedish author. He is perhaps best known for the collaborative work with his partner Maj Sjöwall on a series of ten novels about the exploits of Martin Beck, a police detective in Stockholm, published between 1965 and 1975. In 1971, The Laughing Policeman (a translation of Den skrattande polisen, originally published in 1968) won an Edgar Award from the Mystery Writers of America for Best Novel. Wahlöö and Sjöwall also wrote novels separately.\nWahlöö was born in Tölö parish, Kungsbacka Municipality, Halland. Following school, he worked as a crime reporter from 1946 onwards. After long trips around the world he returned to Sweden and started working as a journalist again.\nHe had a thirteen-year relationship with Sjöwall but they never married, as he already was married. Both were Marxists.	Wikipedia
OL8654905A	47	List of people and institutions rewarded with the Creu de Sant Jordi Award, the second-highest civil distinction awarded in Catalonia (Spain).	Wikipedia
OL28165A	73	David Baldacci (born August 5, 1960) is an American novelist. An attorney by education, Baldacci writes mainly suspense novels and legal thrillers.	Wikipedia
OL2783112A	48	Anthony David Weiner ( born September 4, 1964) is an American former politician who served as the U.S. representative for New York's 9th congressional district from 1999 until his resignation in 2011. A member of the Democratic Party, he consistently carried the district with at least 60% of the vote. Weiner resigned from Congress in June 2011 after it was revealed he sent sexually suggestive photos of himself to different women, including a minor.\nA two-time candidate for Mayor of New York City, Weiner finished second in the Democratic primary in 2005. He ran again in 2013, placing fifth in the Democratic primary.\nIn 2017, Weiner pled guilty to transferring obscene material to a minor and was sentenced to 21 months in prison. He was also required to permanently register as a sex offender. Weiner began serving his federal prison sentence the same year and was released in 2019.	Wikipedia
OL7342508A	49	The sixth and final season of Cobra Kai, also known as Cobra Kai VI, consists of 15 episodes and is releasing on Netflix. Unlike previous seasons, the sixth season will be released in three parts of five episodes each. The first of these was released on July 18, 2024. A second part was released on November 15, 2024, followed by the final part set to release on February 13, 2025. The series is a direct sequel to the original four films in The Karate Kid franchise, focusing on the characters of Daniel LaRusso and Johnny Lawrence over 30 years after the titular film. The season was originally set to be released in 2023, but experienced delays as a result of the 2023 Hollywood labor disputes. The second part, however, got an early release date after originally receiving a November 28, 2024 release window. The season features 13 starring roles, all of which returned from previous seasons, including Gianni DeCenzo who was a series regular in seasons 2–4, but was demoted to a recurring cast member in the previous season.	Wikipedia
OL9182880A	50	The Odd Couple is a play by Neil Simon. Following its premiere on Broadway in 1965, the characters were revived in a successful 1968 film and 1970s television series, as well as several other derivative works and spin-offs.  The plot concerns two mismatched roommates: the neat, uptight Felix Ungar and the slovenly, easygoing Oscar Madison.  Simon adapted the play in 1985 to feature a pair of female roommates (Florence Ungar and Olive Madison) in The Female Odd Couple.  An updated version of the 1965 show appeared in 2002 with the title Oscar and Felix: A New Look at the Odd Couple.	Wikipedia
OL33909A	57	An English crime writer and Conservative life peer in the House of Lords, most famous for a series of detective novels starring policeman and poet Adam Dalgliesh.	OpenLib
OL33909A	58	Phyllis Dorothy James White, Baroness James of Holland Park (3 August 1920 – 27 November 2014), known professionally as P. D. James, was an English novelist and life peer. Her rise to fame came with her series of detective novels featuring the police commander and poet, Adam Dalgliesh.	Wikipedia
OL40671A	61	Lee Earle "James" Ellroy is an American crime fiction writer and essayist. Ellroy has become known for a telegrammatic prose style in his most recent work, wherein he frequently omits connecting words and uses only short, staccato sentences, and in particular for the novels *The Black Dahlia* (1987), *The Big Nowhere* (1988), *L.A. Confidential* (1990), *White Jazz* (1992), *American Tabloid* (1995), *The Cold Six Thousand* (2001), and *Blood's a Rover* (2009). *-- Wikipedia*	OpenLib
OL40671A	62	Lee Earle "James" Ellroy (born March 4, 1948) is an American crime fiction writer and essayist. Ellroy has become known for a telegrammatic prose style in his most recent work, wherein he frequently omits connecting words and uses only short, staccato sentences, and in particular for the novels The Black Dahlia (1987) and L.A. Confidential (1990).	Wikipedia
OL1433006A	67	Flynn, who lives in Chicago, grew up in Kansas City, Missouri. She graduated at the University of Kansas, and qualified for a Master's degree from Northwestern University.	OpenLib
OL1433006A	68	Gillian Schieber Flynn (; born February 24, 1971) is an American author, screenwriter, and producer, best known for her thriller and mystery novels Sharp Objects (2006), Dark Places (2009), and Gone Girl (2012), all of which have received critical acclaim. Her works have been translated into 40 languages, and by 2016, Gone Girl had sold over 15 million copies worldwide.\nFlynn wrote the screenplay for the 2014 film adaptation of Gone Girl, directed by David Fincher, for which she won the Critics’ Choice Movie Award for Best Adapted Screenplay and was nominated for both the Writers Guild of America and the BAFTA awards, among others.\nShe also wrote and produced the HBO limited series adaptation of Sharp Objects, for which she received nominations for the Primetime Emmy and the Writers Guild of America Award. Additionally, Flynn also co-wrote the screenplay for the 2018 film Widows alongside director Steve McQueen.\nFlynn served as showrunner, writer, and executive producer for Amazon Prime Video’s sci-fi thriller series Utopia (2020), which ran for one season. As of 2024, she is working on her fourth novel, to be published by Penguin Random House.	Wikipedia
OL22258A	69	James Brendan Patterson (born March 22, 1947) is an American author. Among his works are the Alex Cross, Michael Bennett, Women's Murder Club, Maximum Ride, Daniel X, NYPD Red, Witch & Wizard, Private and Middle School series, as well as many stand-alone thrillers, non-fiction, and romance novels. His books have sold more than 425 million copies, and he was the first person to sell 1 million e-books. In 2016, Patterson topped Forbes's list of highest-paid authors for the third consecutive year, with an income of $95 million. His total income over a decade is estimated at $700 million.\r\n\r\nIn November 2015, Patterson received the Literarian Award from the National Book Foundation. Patterson has donated millions of dollars in grants and scholarship to various universities, teachers' colleges, independent bookstores, school libraries, and college students to promote literacy. [source](https://en.wikipedia.org/wiki/James_Patterson)	OpenLib
OL22258A	70	James Brendan Patterson (born March 22, 1947) is an American author. Among his works are the Alex Cross, Michael Bennett, Women's Murder Club, Maximum Ride, Daniel X, NYPD Red, Witch & Wizard, Private and Middle School series, as well as many stand-alone thrillers, non-fiction, and romance novels. Patterson's books have sold more than 425 million copies, and he was the first person to sell one million e-books. In 2016, Patterson topped Forbes's list of highest-paid authors for the third consecutive year, with an income of $95 million. His total income over a decade is estimated at $700 million. \nIn November 2015, Patterson received the Literarian Award from the National Book Foundation. He has donated millions of dollars in grants and scholarship to various universities, teachers' colleges, independent bookstores, school libraries, and college students to promote literacy.	Wikipedia
OL5113109A	71	Roscoe Conkling "Fatty" Arbuckle (; March 24, 1887 – June 29, 1933) was an American silent film actor, director, and screenwriter. He started at the Selig Polyscope Company and eventually moved to Keystone Studios, where he worked with Mabel Normand and Harold Lloyd as well as with his nephew, Al St. John. He also mentored Charlie Chaplin, Monty Banks and Bob Hope, and brought vaudeville star Buster Keaton into the movie business. Arbuckle was one of the most popular silent stars of the 1910s and one of the highest-paid actors in Hollywood, signing a contract in 1920 with Paramount Pictures for $1,000,000 a year (equivalent to $15.2 million in 2023).\nArbuckle was the defendant in three widely publicized trials between November 1921 and April 1922 for the rape and manslaughter of actress Virginia Rappe. Rappe had fallen ill at a party hosted by Arbuckle at San Francisco's St. Francis Hotel in September 1921, and died four days later. A friend of Rappe accused Arbuckle of raping and accidentally killing her. The first two trials resulted in hung juries, but the third trial acquitted Arbuckle. The third jury took the unusual step of giving Arbuckle a written statement of apology for his treatment by the justice system.\nDespite Arbuckle's acquittal, the scandal has mostly overshadowed his legacy as a pioneering comedian. At the behest of Adolph Zukor, president of Famous Players–Lasky, his films were banned by motion picture industry censor Will H. Hays after the trial, and he was publicly ostracized. Zukor was faced with the moral outrage of various groups such as the Lord's Day Alliance, the powerful Federation of Women's Clubs and even the Federal Trade Commission to curb what they perceived as Hollywood debauchery run amok and its effect on the morals of the general public.  While Arbuckle saw a resurgence in his popularity immediately after his acquittal, Zukor decided he had to be sacrificed to keep the movie industry out of the clutches of censors and moralists. Hays lifted the ban within a year, but Arbuckle only worked sparingly through the 1920s. In their deal, Keaton promised to give him 35% of the Buster Keaton Comedies Co. profits. He later worked as a film director under the pseudonym William Goodrich. He was finally able to return to acting, making short two-reel comedies in 1932–33 for Warner Bros.\nArbuckle died in his sleep of a heart attack in 1933 at age 46, reportedly on the day that he signed a contract with Warner Bros. to make a feature film.	Wikipedia
OL25714A	72	Nelson Richard DeMille (August 23, 1943 – September 17, 2024) was an American author of action adventure and suspense novels. His novels include Plum Island, The Charm School, and The General's Daughter. DeMille also wrote under the pen names Jack Cannon, Kurt Ladner, Ellen Kay, and Brad Matthews.	Wikipedia
OL228669A	76	Michael S. Ledwidge is an American author of Irish descent. He wrote his first novel, The Narrowback, while working as the back elevator operator for a Park Avenue Coop apartment building.  His novel, Bad Connection was written while working as a lineman for the telephone company in NYC.  His most successful writing has been several books he has co-authored with the best-selling author James Patterson.	Wikipedia
OL896513A	77	Born and raised in New England, Justin Cronin is a graduate of Harvard University and the Iowa Writers’ Workshop. Awards for his fiction include the Stephen Crane Prize, a Whiting Writers’ Award, and a Pew Fellowship in the Arts. He is a professor of English at Rice University and lives with his wife and children in Houston, Texas. \r\n\r\n*From the publisher*	OpenLib
OL896513A	78	Justin Cronin (born 1962) is an American author. He has written six novels: Mary and O'Neil, The Ferryman, and The Summer Guest, as well as a vampire trilogy consisting of The Passage, The Twelve and The City of Mirrors. He has won the PEN/Hemingway Award for Debut Novel, the Stephen Crane Prize, and a Whiting Award.\nBorn and raised in New England, Cronin is a graduate of Harvard University and the Iowa Writers’ Workshop. He taught creative writing and was the "Author in-residence" at La Salle University in Philadelphia, Pennsylvania, from 1992 to 2003. He is a former professor of English at Rice University, and he lives with his wife and children in Houston, Texas.\nIn July 2017, Variety reported that Fox 2000 had bought the screen rights to Cronin's vampire trilogy. The first book of the series, The Passage, was released in June 2010. It garnered mainly favorable reviews. The book has been adapted by Fox into a television series, with Cronin credited as a co-producer.	Wikipedia
OL7945315A	81	Lê Trung Thành (born October 13, 1997), managed by V-MAS entertainment. He also known by his stage name Erik, is a Vietnamese singer and dancer. He first gained recognition competing The Voice Kids of Vietnam in 2013, in addition to having been part in 2016 of the Vietnamese boy group Monstar.	Wikipedia
OL1515334A	84	James Patterson has written or co-written many "Bookshots" or novellas, and has co-written books with many authors. The list below separates the works into four main categories: fiction written for adults, for young adults and for children, and non-fiction.	Wikipedia
OL6474824A	85	John Robert Fowles was an English novelist of international renown, critically positioned between modernism and postmodernism. His work was influenced by Jean-Paul Sartre and Albert Camus, among others.	OpenLib
OL6474824A	86	John Robert Fowles (; 31 March 1926 – 5 November 2005) was an English novelist, critically positioned between modernism and postmodernism. His work was influenced by Jean-Paul Sartre and Albert Camus, among others.\nAfter leaving Oxford University, Fowles taught English at a school on the Greek island of Spetses, a sojourn that inspired The Magus (1965), an instant best-seller that was directly in tune with 1960s "hippy" anarchism and experimental philosophy. This was followed by The French Lieutenant's Woman (1969), a Victorian-era romance with a postmodern twist that was set in Lyme Regis, Dorset, where Fowles lived for much of his life. Later fictional works include The Ebony Tower (1974), Daniel Martin (1977), Mantissa \n(1982), and A Maggot (1985).\nFowles's books have been translated into many languages, and several have been adapted as films.	Wikipedia
OL2631878A	89	Andrew Gross (born 1952) is an American author of thriller novels, including four New York Times bestsellers. He is best known for his collaborations with suspense writer James Patterson. Gross's books feature close family bonds, relationships characterized by loss or betrayal, and a large degree of emotional resonance which generally leads to wider crimes and cover-ups. The books have all been published by William Morrow, an imprint of HarperCollins.	Wikipedia
OL2101074A	91	David John Moore Cornwell (19 October 1931 – 12 December 2020), better known by his pen name John le Carré  was a British Irish author, best known for his espionage novels, many of which were successfully adapted for film or television. A "sophisticated, morally ambiguous writer",he is considered one of the greatest novelists of the postwar era. During the 1950s and 1960s, he worked for both the Security Service (MI5) and the Secret Intelligence Service (MI6). Near the end of his life, due to his strong disapproval of Brexit, he took out Irish citizenship, which was possible due to his having an Irish grandparent.\r\n\r\nLe Carré's third novel, The Spy Who Came in from the Cold (1963), became an international best-seller, was adapted as an award-winning film, and remains one of his best-known works. This success allowed him to leave MI6 to become a full-time author.[4] His novels which have been adapted for film or television include The Looking Glass War (1965), Tinker Tailor Soldier Spy (1974, 2011), Smiley's People (1979), The Little Drummer Girl (1983), The Night Manager (1993), The Tailor of Panama (1996), The Constant Gardener (2001), A Most Wanted Man (2008) and Our Kind of Traitor (2010). Philip Roth said that A Perfect Spy (1986) was "the best English novel since the war".	OpenLib
OL2101074A	92	David John Moore Cornwell (19 October 1931 – 12 December 2020), better known by his pen name John le Carré ( lə-KARR-ay), was a British author, best known for his espionage novels, many of which were successfully adapted for film or television. A "sophisticated, morally ambiguous writer", he is considered one of the greatest novelists of the postwar era. During the 1950s and 1960s, he worked for both the Security Service (MI5) and the Secret Intelligence Service (MI6). Near the end of his life, le Carré became an Irish citizen.\nLe Carré's third novel, The Spy Who Came in from the Cold (1963), became an international best-seller, was adapted as an award-winning film, and remains one of his best-known works. This success allowed him to leave MI6 to become a full-time author. His other novels that have been adapted for film or television include The Looking Glass War (1965), Tinker Tailor Soldier Spy (1974), Smiley's People (1979), The Little Drummer Girl (1983), The Russia House (1989), The Night Manager (1993), The Tailor of Panama (1996), The Constant Gardener (2001), A Most Wanted Man (2008) and Our Kind of Traitor (2010). Philip Roth said that A Perfect Spy (1986) was "the best English novel since the war".	Wikipedia
OL22019A	93	V. C. Andrews - vcandrewsbooks.com\r\n\r\n**Cleo Virginia Andrews, better known as V. C. Andrews or Virginia C. Andrews, was an American novelist.** She was born in Portsmouth, Virginia. Andrews died of breast cancer at the age of 63. Andrews' novels combine Gothic horror and family saga, revolving around family secrets and forbidden love, and sometimes include a rags-to-riches story.Wikipedia\r\nBorn:Cleo Virginia Andrews, Jun 6, 1923, Portsmouth, Virginia, U.S.\r\nDied:Dec 19, 1986, Virginia Beach, Virginia, U.S.\r\nOccupation:Novelist	OpenLib
OL22019A	94	Cleo Virginia Andrews (June 6, 1923 – December 19, 1986), better known as V. C. Andrews or Virginia C. Andrews, was an American novelist. She was best known for her 1979 novel Flowers in the Attic, which inspired two movie adaptations and four sequels. While her novels are not classified by her publisher as Young Adult, their young protagonists have made them popular among teenagers for decades. After her death in 1986, a ghostwriter who was initially hired to complete two unfinished works has continued to publish books under her name.	Wikipedia
OL2632116A	100	Mi afición a los dragones viene de largo. Barcelona es ciudad de dragones, que adornan o vigilan muchas de sus fachadas, y me temo que yo soy uno de ellos. Quizás por eso, por solidaridad con el pequeño monstruo, hace ya muchos años que los colecciono y les ofrezco refugio en mi casa, dragonera al uso. Al día de hoy ya son más de 400 criaturas dragonas las que hacen mi censo, que aumenta cada mes. Además de haber nacido en el año, por supuesto, del dragón, mis vínculos con estas bestias verdes que respiran fuego son numerosos. Somos criaturas nocturnas, aficionadas a las tinieblas, no particularmente sociables, poco amigas de hidalgos y caballeros andantes y difíciles de conocer.\r\n([source][1])\r\n\r\n\r\n  [1]: http://www.carlosruizzafon.com/es/carlos-ruiz-zafon.php	OpenLib
OL2632116A	101	Carlos Ruiz Zafón (Spanish pronunciation: [ˈkaɾlos rwiθ θaˈfon]; 25 September 1964 – 19 June 2020) was a Spanish novelist known for his 2001 novel La sombra del viento (The Shadow of the Wind). The novel sold 15 million copies and was winner of numerous awards; it was included in the list of the one hundred best books in Spanish in the last twenty-five years, made in 2007 by eighty-one Latin American and Spanish writers and critics.	Wikipedia
OL3123066A	106	The following outline is provided as an overview of and topical guide to books.	Wikipedia
OL3353071A	110	Brunonia Barry is the New York Times and international best selling author of The Lace Reader, The Map of True Places and The Fifth Petal, which was recently chosen #1 of Strand Magazine's Top 25 Books of 2017. Her work has been translated into more than thirty languages. She was the first American author to win the International Women’s Fiction Festival’s Baccante Award and was a past recipient of Ragdale Artists’ Colony’s Strnad Fellowship as well as the winner of New England Book Festival’s award for Best Fiction and Amazon’s Best of the Month. Her reviews and articles on writing have appeared in The London Times and The Washington Post. Brunonia co-chairs the Salem Athenaeum’s Writers’ Committee. She lives in Salem with her husband Gary Ward and their dog, Angel. \r\n\r\n([source][1])\r\n\r\n\r\n  [1]: https://brunonia-barry.squarespace.com/bio	OpenLib
OL3353071A	111	Brunonia Barry (born 1950 in Salem, Massachusetts) is the author of The Lace Reader and The Map of True Places.  Her third novel, The Fifth Petal: a novel, was published on January 24, 2017.  Barry, with husband Gary Ward, founded SmartGames, a game and puzzle software company.	Wikipedia
OL765158A	114	Maxine Paetro is an American author who has been published since 1979.  Paetro has collaborated with best-selling author James Patterson on the Women’s Murder Club novel series and standalone novels.	Wikipedia
OL539397A	119	Eric Clifford Ambler OBE was an influential British author of thrillers, in particular spy novels who introduced a new realism to the genre. He also worked as a screenwriter. Ambler used the pseudonym Eliot Reed for books co-written with Charles Rodda.\r\nSource: Wikipedia	OpenLib
OL539397A	120	Eric Clifford Ambler OBE (28 June 1909 – 22 October 1998) was an English author of thrillers, in particular spy novels, who introduced a new realism to the genre. Also working as a screenwriter, Ambler used the pseudonym Eliot Reed for books co-written with Charles Rodda.	Wikipedia
OL34328A	121	Lee Child was born in 1954 in Coventry, England, but spent his formative years in the nearby city of Birmingham. He went to law school in Sheffield, England, and after part-time work in the theater he joined Granada Television in Manchester for what turned out to be an eighteen-year career as a presentation director during British TV's "golden age." During his tenure his company made Brideshead Revisited, The Jewel in the Crown, Prime Suspect, and Cracker. He was fired in 1995 at the age of 40 as a result of corporate restructuring. Always a voracious reader, he decided to see an opportunity where others might have seen a crisis and bought six dollars' worth of paper and pencils and sat down to write a book, Killing Floor, the first in the (now 15 book) Jack Reacher series.\r\n\r\n-- from leechild.com	OpenLib
OL34328A	122	James Dover Grant  (born 29 October 1954), primarily known by his pen name Lee Child, is a British author who writes thriller novels, and is best known for his Jack Reacher novel series. The books follow the adventures of a former American military policeman, Jack Reacher, who wanders the United States. His first novel, Killing Floor (1997), won both the Anthony Award and the 1998 Barry Award for Best First Novel.	Wikipedia
OL31818A	123	Joy Fielding (born on March 18, 1945) is a Canadian novelist and actress.	OpenLib
OL31818A	124	Joy Fielding (née Tepperman; born March 18, 1945) is a Canadian novelist and actress. She lives in Toronto, Ontario.	Wikipedia
OL25277A	125	Thomas Leo "Tom" Clancy Jr. is an American author, best known for his technically detailed espionage and military science storylines set during and in the aftermath of the Cold War, and several video games which he did not work on, but which bear his name for licensing and promotional purposes. His name is also a brand for similar movie scripts written by ghost writers and many series of non-fiction books on military subjects and merged biographies of key leaders. He is also part-owner and Vice Chairman of Community Activities and Public Affairs of the Baltimore Orioles, a Major League Baseball team.\r\n\r\n([Source][1])\r\n\r\n\r\n  [1]: http://en.wikipedia.org/wiki/Tom_Clancy	OpenLib
OL25277A	126	Thomas Leo Clancy Jr. (April 12, 1947 – October 1, 2013) was an American novelist. He is best known for his technically detailed espionage and military-science storylines set during and after the Cold War. Seventeen of his novels have been bestsellers and more than 100 million copies of his books have been sold. His name was also used on screenplays written by ghostwriters, nonfiction books on military subjects occasionally with co-authors, and video games. He was a part-owner of his hometown Major League Baseball team, the Baltimore Orioles, and vice-chairman of their community activities and public affairs committees.\nOriginally an insurance agent, Clancy launched his literary career in 1984 when he sold his first military thriller novel The Hunt for Red October for $5,000 published by the small academic Naval Institute Press of Annapolis, Maryland.\nHis works The Hunt for Red October (1984), Patriot Games (1987), Clear and Present Danger (1989), and The Sum of All Fears (1991) have been turned into commercially successful films. Tom Clancy's works also inspired games such as the Ghost Recon, Rainbow Six, The Division, and Splinter Cell series. Since Clancy's death in 2013, the Jack Ryan series has been continued by his family estate through a series of authors.	Wikipedia
OL20243A	132	An English author, playwright and literary critic.	OpenLib
OL2660362A	158	Tana French  (born 10 May 1973) is an American-Irish writer and theatrical actress. She is a longtime resident of Dublin, Ireland. Her debut novel In the Woods (2007), a psychological mystery, won the Edgar, Anthony, Macavity, and Barry awards for best first novel. The Independent has referred to her as "the First Lady of Irish Crime".	Wikipedia
OL20243A	133	Henry Graham Greene  (2 October 1904 – 3 April 1991) was an English writer and journalist regarded by many as one of the leading novelists of the 20th century.\nCombining literary acclaim with widespread popularity, Greene acquired a reputation early in his lifetime as a major writer, both of serious Catholic novels, and of thrillers (or "entertainments" as he termed them). He was shortlisted for the Nobel Prize in Literature several times. Through 67 years of writing, which included over 25 novels, he explored the conflicting moral and political issues of the modern world. The Power and the Glory won the 1941 Hawthornden Prize and The Heart of the Matter won the 1948 James Tait Black Memorial Prize and was shortlisted for the Best of the James Tait Black.  Greene was awarded the 1968 Shakespeare Prize and the 1981 Jerusalem Prize. Several of his stories have been filmed, some more than once, and he collaborated with filmmaker Carol Reed on The Fallen Idol (1948) and The Third Man (1949).\nHe converted to Catholicism in 1926 after meeting his future wife, Vivien Dayrell-Browning. Later in life he took to calling himself a "Catholic agnostic".\nHe died in 1991, aged 86, of leukemia, and was buried in Corseaux cemetery in Switzerland. William Golding called Greene "the ultimate chronicler of twentieth-century man's consciousness and anxiety".	Wikipedia
OL22261A	134	Sandra Lynn Cox was born on March 12, 1948 in Waco, Texas and raised in Ft. Worth. She is nothing if not serious when it comes to her work. As the oldest of five daughters, she was a responsible and mature girl, and always chose to read a book rather than play with dolls. Her responsible nature stayed with Sandra as she graduated from Texas Christian University with a degree in English, and in her job as a contributing feature reporter at the nationally syndicated PM Magazine in Dallas. When the show experienced mass layoffs, however, Sandra found herself out of work.\r\n\r\nSandra married Michael Brown, former television anchorman and award-winning documentarian of Dust to Dust, and returned to Ft. Worth. They had two children, Rachel and Ryan. Though she continued in her occasional position as a showroom model in Dallas, her husband encouraged her to try fiction writing while their children were at school. He had just left a career as a news anchor and talk-show host to form his own production company, so why shouldn't she take a creative risk, too?\r\n\r\nWithin a year Sandra sold her first novel, Love's Encore, under the name Rachel Ryan (taken from the first names of her two children). Soon thereafter, she was producing a succession of books for six different publishers, culling ideas from briefs in USA Today, television shows, and her own active imagination. She wrote two boosk as Laura Jordan and several books for Harlequin under the name Erin St. Claire.\r\n\r\nSince the publication of her first novel in 1981, she has penned well over sixty books. Sandra has over fifty million copies of her books in print, and has achieved some major feats on what is perhaps the most highly regarded bestseller list of all--that of the New York Times. Since 1990, every one of Sandra's novels has appeared on the list. In total, her books have appeared on the prestigious list over thirty times.\r\n\r\nIn 1992 her novel "French Silk" was made into an ABC-TV movie.	OpenLib
OL22261A	135	Sandra Lynn Brown, née Cox (born March 12, 1948) is an American bestselling author of romantic novels and thriller suspense novels. Brown has also published works under the pen names of Rachel Ryan, Laura Jordan, and Erin St. Claire.	Wikipedia
OL28577A	136	Patricia Highsmith (January 19, 1921 – February 4, 1995) was an American novelist and short story writer widely known for her psychological thrillers, including her series of five novels featuring the character Tom Ripley.\r\n\r\nShe wrote 22 novels and numerous short stories throughout her career spanning nearly five decades, and her work has led to more than two dozen film adaptations. Her writing derived influence from existentialist literature, and questioned notions of identity and popular morality. She was dubbed "the poet of apprehension" by novelist Graham Greene.\r\n\r\nHer first novel, *Strangers on a Train*, has been adapted for stage and screen, the best known being the Alfred Hitchcock film released in 1951. Her 1955 novel *The Talented Mr. Ripley* has been adapted for film. Writing under the pseudonym **Claire Morgan**, Highsmith published the first lesbian novel with a happy ending, *The Price of Salt*, in 1952, republished 38 years later as Carol under her own name and later adapted into a 2015 film. \r\n\r\n**Source**: [Patricia Highsmith](https://en.wikipedia.org/wiki/Patricia_Highsmith) on Wikipedia	OpenLib
OL28577A	137	Patricia Highsmith (born Mary Patricia Plangman; January 19, 1921 – February 4, 1995) was an American novelist and short story writer widely known for her psychological thrillers, including her series of five novels featuring the character Tom Ripley. She wrote 22 novels and numerous short stories in a career spanning nearly five decades, and her work has led to more than two dozen film adaptations. Her writing was influenced by existentialist literature, and questioned notions of identity and popular morality. She was dubbed "the poet of apprehension" by novelist Graham Greene.\nBorn in Fort Worth, Texas, and mostly raised in her infancy by her maternal grandmother, Highsmith moved to New York City at the age of six to live with her mother and step father. After graduating college in 1942, she worked as a writer for comic books while writing her own short stories and novels in her spare time. Her literary breakthrough came with the publication of her first novel Strangers on a Train (1950) which was adapted into a 1951 film directed by Alfred Hitchcock. Her 1955 novel The Talented Mr. Ripley was well received in the United States and Europe, cementing her reputation as a major exponent of psychological thrillers.\nIn 1963, Highsmith moved to England where her critical reputation continued to grow. Following the breakdown of her relationship with a married Englishwoman, she moved to France in 1967 to try to rebuild her life. Her sales were now higher in Europe than in the United States which her agent attributed to her subversion of the conventions of American crime fiction. She moved to Switzerland in 1982 where she continued to publish new work that increasingly divided critics. The last years of her life were marked by ill health and she died of aplastic anemia and lung cancer in Switzerland in 1995.\nThe Times said of Highsmith: "she puts the suspense story in a toweringly high place in the hierarchy of fiction.": 180  Her second novel, The Price of Salt, published under a pseudonym in 1952, was ground breaking for its positive depiction of lesbian relationships and optimistic ending.: 1  She remains controversial for her antisemitic, racist and misanthropic statements.	Wikipedia
OL713259A	151	Desmond Bagley (29 October 1923 – 12 April 1983) was an English journalist and novelist known mainly for a series of bestselling thrillers. He and fellow British writers such as Hammond Innes and Alistair MacLean set conventions for the genre: a tough, resourceful, but essentially ordinary hero pitted against villains determined to sow destruction and chaos for their own ends.	Wikipedia
OL3288848A	154	David or Dave Ellis may refer to:	Wikipedia
OL6925017A	155	Ralph Hammond Innes was a British novelist who wrote over 30 novels, as well as works for children and travel books.	OpenLib
OL6925017A	156	Ralph Hammond Innes  (15 July 1913 – 10 June 1998) was a British novelist who wrote over 30 novels, as well as works for children and travel books.	Wikipedia
OL2660362A	157	Tana French (May 10, 1973) is an American-Irish novelist and theatrical actress. Her debut novel In the Woods (2007), a psychological mystery, won the Edgar, Anthony, Macavity, and Barry awards for best first novel. She lives in Dublin. She is referred to as the 'First Lady of Irish Crime'. Source: Wikipedia	OpenLib
OL76547A	159	Martin Cruz Smith, born Martin William Smith (November 3, 1942), is an American writer of mystery and suspense fiction, mostly in an international or historical setting. He is best known for his ten-novel series (to date) on Russian investigator Arkady Renko, introduced in 1981 with Gorky Park. The tenth book in the series, Independence Square, was published in May 2023.	Wikipedia
OL7442916A	165	Daniel Mallory (born 1979) is an American author who writes crime fiction under the name A. J. Finn. His 2018 novel The Woman in the Window debuted at number one on the New York Times Best Seller list. The Woman in the Window was adapted into a feature film of the same name, directed by Joe Wright and featuring Amy Adams, Julianne Moore and Gary Oldman.\nIn 2019 an article in The New Yorker stated that Mallory had frequently lied about his personal life and health. Mallory obliquely acknowledged being deceptive in a statement. Mallory attributed his actions to his struggles with bipolar depressive disorder, which drew criticism from psychiatrists. His second novel, End of Story, was published in February 2024.	Wikipedia
OL45364A	168	Henning Mankell was born in Stockholm, Sweden, the son of a judge. He grew up in the towns of Sveg and Borås. His grandfather, also called Henning Mankell (1868–1930), was a well-known composer. At the age of 20, Mankell was the assistant director at the Riks Theater in Stockholm, and he was also writing. In the 1970s he moved to Norway, where he lived with a woman who was a member of the Maoist Communist Labour Party, although he never officially joined the Party. He moved to Africa and lived in several African countries, and in 1985 he founded the Avenida Theater in Maputo, Mozambique, where he continues to spend about half of every year. In 1997 he began his most well-known series of novels, a series of murder mysteries set in Ystad, Sweden, featuring the police detective Kurt Wallander. He also established a publishing house, Leopard Förlag, to publish young talents from both Africa and Sweden. He is married to Eva Bergman, daughter of Ingmar Bergman.	OpenLib
OL45364A	169	Henning Georg Mankell (Swedish pronunciation: [ˈhɛ̂nːɪŋ ˈmǎŋːkɛl]; 3 February 1948 – 5 October 2015) was a Swedish crime writer, children's author, and dramatist, best known for a series of mystery novels starring his most noted creation, Inspector Kurt Wallander. He also wrote a number of plays and screenplays for television.\nHe was a left-wing social critic and activist. In his books and plays he constantly highlighted social inequality issues and injustices in Sweden and abroad. In 2010, Mankell was on board one of the ships in the Gaza Freedom Flotilla that was boarded by Israeli commandos. He was below deck on the MV Mavi Marmara when nine civilians were killed in international waters.\nMankell shared his time between Sweden and countries in Africa, mostly Mozambique where he started a theatre. He made considerable donations to charity organizations, mostly connected to Africa.	Wikipedia
OL2631898A	170	Hill is the second child of authors Stephen and Tabitha King. He grew up in Bangor, Maine. His younger brother Owen is also a writer. Hill has three sons.\r\n\r\nHill chose to use an abbreviated form of his given name (a reference to executed labor leader Joe Hill, for whom he was named) in 1997, out of a desire to succeed based solely on his own merits instead of as the son of Stephen King. After achieving a degree of independent success, Hill publicly confirmed his identity in 2007 after an article the previous year in Variety broke his cover (although online speculation about Hill's family background had been appearing since 2005).\r\n\r\nJoe Hill is a past recipient of the Ray Bradbury Fellowship. He has also received the William L. Crawford award for best new fantasy writer in 2006, the A. E. Coppard Long Fiction Prize in 1999 for "Better Than Home" and the 2006 World Fantasy Award for Best Novella for "Voluntary Committal". His stories have appeared in a variety of magazines, such as Subterranean Magazine, Postscripts and The High Plains Literary Review, and in many anthologies, including The Mammoth Book of Best New Horror (ed. Stephen Jones) and The Year's Best Fantasy and Horror (ed. Ellen Datlow, Kelly Link & Gavin Grant).\r\n\r\nHill's first book, the limited edition collection 20th Century Ghosts published in 2005 by PS Publishing), showcases fourteen of his short stories and won the Bram Stoker Award for Best Fiction Collection, together with the British Fantasy Award for Best Collection and Best Short Story for "Best New Horror". In October 2007, Hill's mainstream US and UK publishers reprinted 20th Century Ghosts, without the extras published in the 2005 slipcased versions, but including one new story.\r\n\r\nHill's first novel, Heart-Shaped Box, was published by William Morrow/HarperCollins on February 13, 2007 and by Victor Gollancz Ltd in UK in March 2007. Simultaneous to these two editions, a limited edition of Heart-Shaped Box was also released by Subterranean Press; it sold out several months prior to publication. The novel reached number 8 on the New York Times bestseller list on April 1, 2007.\r\n\r\nOn September 23, 2007, at the thirty-first Fantasycon, the British Fantasy Society awarded Hill the first ever Sydney J. Bounds Best Newcomer Award. Hill's first professional sale was in 1997.\r\n\r\nAmong unpublished works is one partly completed with his father, "But Only Darkness Loves Me", which is held with the Stephen King papers at the Special Collections Unit of the Raymond H Fogler Library at the University of Maine in Orono, Maine.\r\n\r\nHill is also the author of Locke & Key, a new comic book series published by IDW Publishing. The first issue, released on February 20, 2008, sold out of its initial publication run in one day. A forthcoming collection of the series in limited form from Subterranean Press sold out within 24 hours of being announced.\r\n\r\nHis only screen appearance so far was aged 10 in the film Creepshow (1982) (dir. George Romero), which co-starred and was co-written by his father.	OpenLib
OL2631898A	171	Joe Hill may refer to:\n\nJoe Hill (activist) (1879–1915),  Swedish-American labor activist and songwriter\nJoe Hills (1897–1969), English cricketer and umpire\nJoe Hill (alias of Joseph Graves Olney, 1849–1884), American rancher and outlaw\nBlind Joe Hill (1937–1998), American blues singer, guitarist, harmonica player and drummer\nJoseph Hill (a.k.a. Dusty Hill) (1949–2021), American bassist associated with the band ZZ Top\nJoe Hill (writer) (born 1972), pen name of American author Joseph Hillstrom King, son of author Stephen King\nJoe Hills (American football) (born 1987), American football wide receiver\nJoe Hill, fictional character on the TV series Blue Bloods\nJo Hill (born 1973), Australian women's basketball player\nJoe Hill (journalist), Australian television presenter associated with station ADS	Wikipedia
OL1391085A	172	an American novelist and film producer. She is best known for writing the vampire romance series Twilight.\r\nMeyer was the bestselling author of 2008 and 2009 in the U.S. \r\nMeyer received the 2009 Children's Book of the Year award from the British Book Awards for Breaking Dawn, the Twilight series finale.\r\n\r\nStephenie Morgan was born on December 24, 1973, in Hartford, Connecticut, the second of six children to financial officer Stephen Morgan and homemaker Candy Morgan. Meyer was raised in Phoenix, Arizona, and attended Chaparral High School in Scottsdale, Arizona. In 1992, Meyer won a National Merit Scholarship, which helped fund her undergraduate studies at Brigham Young University in Provo, Utah, where she received a BA in English Literature in 1997. Although she began and finished her degree at BYU, she took classes at Arizona State University in fall 1996 and spring 1997.\r\n\r\nMeyer met her future husband, Christian "Pancho" Meyer, in Arizona when they were both children. They married in 1994, when Meyer was twenty-one.Together, they have sons who Christian Meyer retired from his job as an auditor to take care of full time.\r\n\r\nBefore writing her first novel, Twilight, Meyer considered going to law school because she felt she had no chance of becoming a writer. She later noted that the birth of her oldest son, Gabe, in 1997 changed her mind: "Once I had Gabe, I just wanted to be his mom." Before becoming an author, Meyer's only professional work was as a receptionist at a property company.	OpenLib
OL1391085A	173	Stephenie Meyer (; née Morgan; born December 24, 1973) is an American author and film producer. She is best known for writing the vampire romance book series Twilight, which has sold over 160 million copies, with translations into 49 different languages. She was the bestselling author of 2008 and 2009 in the United States, having sold over 29 million books in 2008 and 26.5 million in 2009.\nAn avid young reader, she attended Brigham Young University, marrying at the age of twenty-one before graduating with a degree in English in 1997. Having no prior experience as an author, she conceived the idea for the Twilight series in a dream. Influenced by the work of Jane Austen and William Shakespeare, she wrote Twilight soon thereafter. After many rejections, Little, Brown and Company offered her a $750,000 three-book deal which led to a four-book series, two spin-off novels, a novella, and a series of commercially successful film adaptations. Aside from young adult novels, Meyer has ventured into adult novels with The Host (2008) and The Chemist (2016). Meyer has worked in film production and co-founded production company Fickle Fish Films. Meyer produced both parts of Breaking Dawn, the Twilight film series' finale, and two other novel adaptations.\nMeyer's membership in the Church of Jesus Christ of Latter-day Saints shaped her novels. Themes consistent with her religion—including agency, mortality, temptation, eternal life, and pro-life—are featured in her work. Critics have called her writing style overly simplistic, but her stories have also received praise, and she has acquired a fan following.\nMeyer was included on Time's list of the "top 100 most influential people" in 2008 and Forbes's list of the "top 100 most powerful celebrities" in 2009, with her annual earnings exceeding $50 million.	Wikipedia
OL1390877A	174	Scottish crime writer and playwright	OpenLib
OL1390877A	175	Denise Mina (born 21 August 1966) is a Scottish crime writer and playwright. She has written the Garnethill trilogy and another three novels featuring the character Patricia "Paddy" Meehan, a Glasgow journalist.  Described as an author of Tartan Noir, she has also written for comic books, including 13 issues of Hellblazer.\nMina's first Paddy Meehan novel, The Field of Blood (2005), was filmed for broadcast in 2011 by the BBC, starring  Jayd Johnson, Peter Capaldi and David Morrissey. The second, The Dead Hour, was filmed and broadcast in 2013.	Wikipedia
OL1372814A	181	James Patterson has written or co-written many "Bookshots" or novellas, and has co-written books with many authors. The list below separates the works into four main categories: fiction written for adults, for young adults and for children, and non-fiction.	Wikipedia
OL26369A	182	The Boy with the Arab Strap is the third studio album by Scottish indie pop band Belle & Sebastian, released in 1998 through Jeepster Records.	Wikipedia
OL7287124A	183	Fetch the Bolt Cutters is the fifth studio album by American singer-songwriter Fiona Apple. It was released on April 17, 2020, Apple's first release since The Idler Wheel... in 2012. The album was recorded from 2015 to 2020, largely at Apple's home in Venice Beach. It was produced and performed by Apple alongside Amy Aileen Wood, Sebastian Steinberg and Davíd Garza; the recording consisted of long, often improvised takes with unconventional percussive sounds. GarageBand was used for much of this recording, and Fiona Apple credited the album's unedited vocals and long takes to her lack of expertise with the program.\nRooted in experimentation, the album largely features unconventional percussion. While conventional instruments, such as pianos and drum sets, do appear, the album also features prominent use of non-musical found objects as percussion. Apple described the result as "percussion orchestras". These industrial-like rhythms are contrasted against traditional melodies, and the upbeat songs often subvert traditional pop structures.\nThe album explores freedom from oppression; Apple identified its core message as: "Fetch the fucking bolt cutters and get yourself out of the situation you're in". The title, a quote from TV series The Fall, reflects this idea. The album also discusses Apple's complex relationships with other women and other personal experiences, including bullying and sexual assault. It has nevertheless been referred to as Apple's most humorous album.\nFetch the Bolt Cutters was released during the COVID-19 pandemic, and many critics found its exploration of confinement timely. It received widespread acclaim from music critics, who described it as an instant classic, revolutionary, and Apple's best work to-date. The album was awarded Best Alternative Music Album at the 63rd Annual Grammy Awards, with "Shameika" winning Best Rock Performance. The album debuted at number four on the US Billboard 200 and number one on the US Top Alternative Albums and Top Rock Albums, with 44,000 equivalent album units. It also charted in the top 15 in Canada, Australia and New Zealand.	Wikipedia
OL23073A	184	Hans-Ulrich Treichel (born 12 August 1952) is a Germanist, novelist and poet. His earliest published books were collections of poetry, but prose writing has become a larger part of his output since the critical and commercial success of his first novel Der Verlorene (translated into English as Lost). Treichel has also worked as an opera librettist, most prominently in collaboration with the composer Hans Werner Henze.	Wikipedia
OL444013A	185	1988 U.S. National Book Award winner for his novel Paris Trout.	OpenLib
OL444013A	186	Pete Dexter (born July 22, 1943) is an American novelist. He won the U.S. National Book Award in 1988 for his novel Paris Trout.	Wikipedia
OL39821A	190	Harlan Coben (born c. 1962) is an American writer of mystery novels and thrillers. The plots of his novels often involve the resurfacing of unresolved or misinterpreted events in the past, murders, or fatal accidents and have multiple twists. Twelve of his novels have been adapted for film and television. \nCoben has won an Edgar Award, a Shamus Award, and an Anthony Award—the first author to receive all three. His books have been translated into 43 languages and sold over 60 million copies.	Wikipedia
OL7871059A	191	January LaVoy (born in Trumbull, Connecticut) is an American actress and audiobook narrator. As an actress, she is most recognized as Noelle Ortiz on the ABC daytime drama One Life to Live. LaVoy made her Broadway debut in the Broadway premiere of the play Enron at the Broadhurst Theatre on April 27, 2010. \nAs an audiobook narrator, she has received five Audie Awards and been a finalist for nineteen. In 2013, she won Publishers Weekly's Listen Up Award for Audiobook Narrator of the Year. In 2019, AudioFile named her a Golden Voice narrator.	Wikipedia
OL2757233A	194	Mark Sullivan may refer to:\n\nMark J. Sullivan, director of the United States Secret Service, 2006–2013\nMark Sullivan (cricketer) (born 1964), South African cricketer\nMark T. Sullivan (born 1958), American author of mystery and suspense novels\nMark Sullivan (judge) (1911–2001), justice on the New Jersey Supreme Court, 1973–1981\nMark Sullivan (public servant), former secretary of the Australian Government Department of Veterans' Affairs\nMark Sullivan, founder Snowboard Magazine\nMark Sullivan, keyboardist for several California bands including Toiling Midgets\nMark Sullivan, chief scientist for Eagle Eye Technologies, Inc, later SkyBitz\nMark Sullivan (visual effects artist), Academy Award nominated visual effects artist\nMark Sullivan (American football), American football coach\nMark Sullivan (journalist) (1874–1952), American political commentator\nMark Sullivan (runner), winner of the 1988 4 × 800 meter relay at the NCAA Division I Indoor Track and Field Championships	Wikipedia
OL1433920A	195	Blake Crouch is a bestselling novelist and screenwriter. He is the author of the forthcoming novel, Dark Matter, for which he is writing the screenplay for Sony Pictures. His international-bestselling Wayward Pines trilogy was adapted into a television series for FOX, executive produced by M. Night Shyamalan, that was Summer 2015’s #1 show. With Chad Hodge, Crouch also created Good Behavior, the TNT television show starring Michelle Dockery based on his Letty Dobesh novellas. He has written more than a dozen novels that have been translated into over thirty languages and his short fiction has appeared in numerous publications including Ellery Queen and Alfred Hitchcock Mystery Magazine. Crouch lives in Colorado with his family.	OpenLib
OL1433920A	196	William Blake Crouch (born October 15, 1978) is an American author known for books such as Dark Matter, Recursion, Upgrade, and his Wayward Pines Trilogy, which was adapted into a television series in 2015. Dark Matter was adapted for television in 2024.	Wikipedia
OL7566140A	199	James Patterson has written or co-written many "Bookshots" or novellas, and has co-written books with many authors. The list below separates the works into four main categories: fiction written for adults, for young adults and for children, and non-fiction.	Wikipedia
OL30714A	200	Linda S. was born August 3, 1950 in Gadsden, Alabama, U.S.A.. She cut her teeth on Margaret Mitchell, Robert Ruark, "and anything else that fell into my hands," she says. Whether she is reading them or writing them, books have long played a profound role in Linda's life. Linda wrote her first book when she was 10 years old. "Needless to say, it was unpublishable," she says. "It didn't even have a title. I didn't name them back then."\r\n\r\nIn the ensuing 21 years of writing for her own pleasure, following junior college Linda worked in the transportation industry, where she met Gary F. Howington, her husband. "In the company I worked for, my title was secretary to the terminal manager, but I actually did very little secretarial work," she says. "I worked in every phase of the transportation business, but my main duties were payroll, insurance, and the efficiency and production reports." Writing production reports, however, soon grew tiresome for Linda.\r\n\r\nAs she continued to write fiction, concentrating on romantic stories. "I get bored with politics and murder and mayhem," she says. She eventually worked up the courage to submit a manuscript for publication. "It made me sick literally, physically ill. It was like putting your naked baby into the mailbox. And I lost 20 pounds waiting to hear from them. I couldn't eat." Linda needn't have worried Silhouette Books bought her manuscript, beginning a career that has (so far) lasted over 10 years and earned her many awards and letters of praise from adoring fans. She has over 10 million books in print around the world, and has written more than 25 titles. Linda has written for Silhouette Special Edition and continues to write for Silhouette Sensation, and is a New York Times bestselling author for Pocket Books writing historicals.\r\n\r\nLinda Howard is a charter member of RWA, joining in 1981 shortly after it was formed. She is one of the original members of her local RWA chapter, has served as treasurer, vice president, and president of that chapter, and has twice been a RITA finalist. In addition to her wide public acclaim, Linda has also been honored by both the critics and her peers many times. She has won the B. Dalton Bestseller Award and the Romantic Times Magazine Reviewers' Choice Award for Series and the W.I.S.H. Award for hero Joe Mackenzie from her Silhouette Intimate Moments title, Mackenzie's Mission. A tie-in book, Mackenzie's Pleasure, reached number 61 on the USA Today bestseller list. A Romance Writers of America RITA and Golden Choice finalist, she is a frequent Waldenbooks bestselling author, often claiming the number-one position.\r\n\r\nNow, Linda has three grown step-children and three grandchildren. She lives in her native Alabama with her husband Gary and two golden retrievers, named Bit O'Honey and Sugar Baby. They live in in a big house that's very much a home and not a showplace. "It's a house where the kids romp, the dogs romp, and you can sit on any piece of furniture. Her husband fishes BassMaster tournament trail for a living, and she travels with him."	OpenLib
OL30714A	201	Linda S. Howington (born August 3, 1950 in Alabama, United States) is an American best-selling romance/suspense author under her pseudonym Linda Howard.	Wikipedia
OL29079A	205	Cussler began writing novels in 1965 and published his first work featuring his continuous series hero, Dirk Pitt, in 1973. His first non-fiction, The Sea Hunters, was released in 1996. The Board of Governors of the Maritime College, State University of New York, considered The Sea Hunters in lieu of a Ph.D. thesis and awarded Cussler a Doctor of Letters degree in May, 1997. It was the first time since the College was founded in 1874 that such a degree was bestowed. \r\n\r\nCussler is an internationally recognized authority on shipwrecks and the founder of the National Underwater and Marine Agency, (NUMA) a 501C3 non-profit organization (named after the fictional Federal agency in his novels) that dedicates itself to preserving American maritime and naval history. He and his crew of marine experts and NUMA volunteers have discovered more than 60 historically significant underwater wreck sites including the first submarine to sink a ship in battle, the Confederacy's Hunley, and its victim, the Union's Housatonic; the U-20, the U-boat that sank the Lusitania; the Cumberland, which was sunk by the famous ironclad, Merrimack; the renowned Confederate raider Florida; the Navy airship, Akron, the Republic of Texas Navy warship, Zavala, found under a parking lot in Galveston, and the Carpathia, which sank almost six years to-the-day after plucking Titanic's survivors from the sea. \r\n\r\nIn September, 1998, NUMA - which turns over all artifacts to state and Federal authorities, or donates them to museums and universities - launched its own web site for those wishing more information about maritime history or wishing to make donations to the organization. (www.numa.net). \r\n\r\nIn addition to being the Chairman of NUMA, Cussler is also a fellow in both the Explorers Club of New York and the Royal Geographic Society in London. He has been honored with the Lowell Thomas Award for outstanding underwater exploration. \r\n\r\nCussler's books have been published in more than 40 languages in more than 100 countries. His past international bestsellers include Pacific Vortex, Mediterranean Caper, Iceberg, Raise the Titanic, Vixen 03, Night Probe, Deep Six, Cyclops, Treasure, Dragon, Sahara, Inca Gold, Shock Wave, Flood Tide, Atlantis Found, Valhalla Rising, Trojan Odyssey and Black Wind (this last with his son, Dirk Cussler); the nonfiction books The Sea Hunters, The Sea Hunters II and Clive Cussler and Dirk Pitt r Revealed; the NUMA® Files novels Serpent, Blue Gold, Fire Ice, White Death and Lost City (written with Paul Kemprecos); and the Oregon Files novels Sacred Stone and Golden Buddha (written with Craig Dirgo) and Dark Watch (written with Jack Du Brul). \r\n\r\nTaken From Good Reads\r\nhttps://www.goodreads.com/author/show/18411.Clive_Cussler	OpenLib
OL29079A	206	Clive Eric Cussler (July 15, 1931 – February 24, 2020) was an American adventure novelist and underwater explorer. His thriller novels, many featuring the character Dirk Pitt, have been listed on The New York Times fiction best-seller list more than 20 times. Cussler was the founder and chairman of the National Underwater and Marine Agency (NUMA), which has discovered more than 60 shipwreck sites and numerous other notable underwater wrecks. He was the sole author or main author of more than 80 books. He often placed himself into his books as himself.\nHis novels have inspired various other works of fiction.	Wikipedia
OL2660314A	207	**Dirk Cussler** arbeitete nach seinem Studium in Berkeley viele Jahre lang in der Finanzwelt, bevor er sich hauptberuflich dem Schreiben widmete. Darüber hinaus nahm er an mehreren der über achtzig Expeditionen der NUMA teil.	OpenLib
OL2660314A	208	Dirk Cussler (born 1961) is an American author. He is the son of best selling author Clive Cussler and a co-author of several Dirk Pitt adventure novels, as well as being the namesake of the Pitt character.	Wikipedia
\.


--
-- TOC entry 3675 (class 0 OID 24738)
-- Dependencies: 225
-- Data for Name: digital; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.digital (isbn10, form) FROM stdin;
3785749023	audio cd
0739313118	audio cassette
1598951378	preloaded digital audio player
0739339796	audio cd
0385504217	ebook
0739307312	audio cassette
1933499834	audio cd
1428124527	preloaded digital audio player
0743539451	audio cassette
074358046X	audio cd
0743580451	audio cd
030773692X	audio cd
0385533136	ebook
1616573058	preloaded digital audio player
0804165742	audio cd
2878626176	audio cd
1797107313	audio cd
0736616497	audio cassette
0671690841	audio cassette
3785710380	Audio CD, 7
3785710372	Audio cassette, 6
3785713193	Audio CD, 12
1405091045	audio cd
0143142380	Audio CD
0743527631	Audio cassette
1508297126	audio cd
1508217114	Digital Audio
1508230749	audio cd
0743599101	audio cd
0743536304	audio cassette
0743539478	audio cd
074353946X	audio cassette
1402575440	audio cd
074353574X	audio cassette
0743535758	audio cd
0743504356	audio cassette
1508278717	audio cd
1508226636	Digital Audio
073669689X	audio cassette
0786113723	audio cassette
1427271283	audio cd
0449008606	audio cd library binding
038552885X	electronic resource
0307987582	Digital Audio
0307576132	ebook
0553470698	Audio cassette
0553502220	Audio cassette
0553712640	Audio CD
0553702203	Audio cassette
555576751X	Audio Cassette
0553745190	Audio Cassette
0375412204	ebook
1664626875	audio cd
0788730932	Audio Cassette
0788737252	Audio CD
0671582348	audio cassette
1508293562	audio cd
0671045857	audio cassette
0671045865	audio cd
0743563395	Digital Audio
3785711131	Audio CD
3785711123	Audio Cassette
1508217475	Digital Audio
9993574864	Audio Cassette
1508217505	Digital Audio
3453165802	Audio cassette
0553502123	Audio cassette
0553479180	Audio cassette
0739312170	Audio Cassette
3550101244	Audio CD
3550101252	Audio cassette
3453198638	Audio Cassette
0743527070	Audio Cassette
0743527054	Audio Cassette
067167854X	Audiobook on cassette
345319862X	Audiobook on CD
1440657750	eBook
075310833X	Audio Cassette
0745141323	Audio cassette
1433263262	audio cd
1433249065	audio cd
1844561607	audio cd
0671582356	Audio cassette
0743563352	Digital Audio
0743509870	Audio CD
0743509862	Audio cassette
0743520955	audio cassette
0743520963	Audio CD
1402518862	Audio cassette
0743563336	Digital Audio
1508218315	Digital Audio
0788711636	Audio cassette
1508218676	audio cd
0143143913	Digital Audio
039456331X	Audio Cassette
0816195765	Audio cassette
0745160638	Audio cassette
179710635X	audio cd
1508219168	Digital Audio
0743533526	Audio CD
0743533518	Audio cassette
0739346792	Digital Audio
0736618163	Audio cassette
0739323881	Audio CD
1456112503	Digital Audio
0743597311	Digital Audio
1501216775	Digital Audio
0307816508	ebook
0385366752	audio cd
1478938056	audio cd
1586215809	Audio CD
0736698647	Audio CD
1586215795	Audio Cassette
1478938269	audio cd
1600240941	Audio CD
1586217089	Audio CD
1415904936	Audio Cassette
1586217097	Audio CD
1415904944	Audio CD
1586217100	Audio cassette
1598953834	Audio CD
1594835802	Audio CD
1594835829	Audio cassette
0759569010	Electronic resource
1594835845	Audio CD
1594836272	Audio CD
1600242790	Audio CD
1594836256	Audio Cassette
159483623X	Audio CD
0345516869	ebook
1984886851	audio cd
0449806944	audio cd
0061673536	audio cd
166506417X	audio cd
0007254903	Digital Audio
0060873094	Audio CD
006087306X	Audio CD
1586217267	Audio Cassette
0755324811	Audio CD
1415907943	Audio CD
1586217275	Audio CD
1415907935	Audio Cassette
0745127703	Audio Cassette
1586214144	Audio Cassette
1586215353	Audio Cassette
1586214152	Audio CD
1586215361	Audio CD
0446598119	ebook
1600240488	Audio CD
1600240526	Audio CD
1489358676	audio cd
5552758609	Audio Cassette
067942508X	Audio cassette
0736627898	Audio cassette
5557103368	Audio cassette
0754053598	Audio CD
0745143539	Audio Cassette
075311772X	Audio Cassette
0753122316	Audio CD
5553819318	Audio Cassette
5557098127	Audio Cassette
5557103023	Audio cassette
0753122766	Audio CD
0753117738	Audio Cassette
1478963689	audio cd
1594830479	Audio CD
0755325710	Audio CD
1600242537	Audio CD
1415923248	Audio Cassette
0755325702	Audio Cassette
1600248497	audio cd
1600240577	Audio Cassette
1600240593	Audio CD
1600240550	Audio CD
1478956259	audio cd
160024226X	Audio CD
1600242308	Audio CD
159483928X	Audio CD
1594839263	Audio Cassette
0061982997	ebook
0061696099	Electronic resource
0061661554	Audio CD
1607885441	audio cd
160788545X	audio cd
1607886901	audio cd
1445007800	audio cd
1445007797	audio cassette
1607884488	audio cd
0804164223	audio cd
0804164231	audio cd
141593309X	Audio CD
0736616063	Audio Cassette
1524722871	audio cd
0307749665	audio cd
1524734403	audio cd
0553713094	Audiobook on cassette
0553713108	Audiobook on CD
3550090757	Audiobook on CD
1101912618	audio cd
159483895X	Audio CD
1594838925	Audio CD
1594838941	Audio Cassette
0743526368	audio cd
074355521X	Audio CD
1101921854	audio cd
1478976519	preloaded digital audio player
1478972955	audio cd
1478969652	audio cd
1478988150	audio cd
1478930020	audio cd
1478903546	audio cd
1478955473	audio cd
1478952482	audio cd
1478927682	audio cd
1478984236	audio cd
0062209035	audio cd
0062270737	audio cd
1549113186	preloaded digital audio player
1549194461	audio cd
1478921528	audio cd
1478998423	audio cd
0816196877	Audio Cassette
0745157653	Audio cassette
1478964111	audio cd
0736608540	Audio cassette
555367333X	Audio cassette
1449839401	audio cd
0679451617	audio cassette
0333672240	Audio Cassette
0913369683	Audio Cassette
1549168045	audio cd
1549171585	audio cd
1549171593	audio cd
0062896210	audio cd
1478952989	audio cd
0307877965	audio cd
0062314270	audio cd
0061235873	Audio CD
1415937559	Audio CD
1478906561	audio cd
1665187239	audio cd
1598870289	Audio CD
0399565035	audio cd
1984833162	audio cd
1984833170	audio cd
1478951699	audio cd
0525590161	audio cd
0147525179	audio cd
0061287350	Digital Audio
1478953632	audio cd
1455856053	audio cd
1478953829	audio cd
0525590285	audio cd
1101924470	audio cd
1101904232	eBook
1549178075	audio cd
1549178083	audio cd
1423310020	Audio CD
1423310004	Audio cassette
142331008X	Audio CD
1423309995	Audio cassette
1423310012	Audio CD
1423310071	Audio CD
147891632X	audio cd
1524702943	audio cd
1984883518	audio cd
\.


--
-- TOC entry 3674 (class 0 OID 24728)
-- Dependencies: 224
-- Data for Name: edition; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.edition (isbn10, wid, publish_date) FROM stdin;
3404154851	OL76837W	April 8, 2006
8389779110	OL76837W	2004-01-01
9100102970	OL76837W	2003
8466423117	OL76837W	Aug 30, 2017
8497870794	OL76837W	Dec 01, 2004
8497870379	OL76837W	Apr 16, 2004
849787157X	OL76837W	Nov 04, 2005
9639526770	OL76837W	Feb 03, 2004
9875808857	OL76837W	Nov 02, 2013
8804628553	OL76837W	Mar 01, 2013
8408175726	OL76837W	Aug 29, 2017
9584255185	OL76837W	2016-11
3785749023	OL76837W	May 14, 2013
9601409025	OL76837W	2004
6171247588	OL76837W	Apr 10, 2018
0739313118	OL76837W	Oct 30, 2003
1598951378	OL76837W	Apr 20, 2005
2709628538	OL76837W	May 19, 2006
8804523417	OL76837W	2005
4047914746	OL76837W	2004
4047914754	OL76837W	2004
8497081749	OL76837W	May 01, 2006
8417031235	OL76837W	Aug 30, 2017
849930026X	OL76837W	Oct 01, 2009
0739339796	OL76837W	Mar 28, 2006
0593055810	OL76837W	2005
8126412267	OL76837W	Nov 18, 2017
2709643952	OL76837W	May 02, 2013
0385504217	OL76837W	2017
0593056574	OL76837W	Dec 13, 2006
5550155184	OL76837W	Apr 24, 2003
0739307312	OL76837W	Apr 07, 2003
0552173878	OL76837W	2016
8391913139	OL76837W	2006
8373591672	OL76837W	2006
0307277674	OL76837W	2006-03
8957591052	OL76837W	2004
3763255389	OL76837W	2004
8373594213	OL76837W	2005
8408176021	OL76837W	2017-08
3404148665	OL76833W	2004-07
8804531673	OL76833W	\N
8476696957	OL76833W	Apr 01, 2005
8417031278	OL76833W	Aug 30, 2017
8492549742	OL76833W	May 01, 2009
8466423133	OL76833W	Aug 30, 2017
8497870476	OL76833W	Sep 01, 2004
8804667249	OL76833W	May 24, 2016
2266198122	OL76833W	May 28, 2009
9722514091	OL76833W	Oct 18, 2005
8599296027	OL76833W	Oct 27, 2005
8799015714	OL76833W	2004-11-05
8389779285	OL76833W	\N
8389779005	OL76833W	\N
8375088285	OL76833W	Apr 23, 2013
975210455X	OL76833W	2007
9752110800	OL76833W	Nov 03, 2009
9752106404	OL76833W	Nov 03, 2005
8492516518	OL76833W	Mar 01, 2009
9875808873	OL76833W	Nov 02, 2013
8467210117	OL76833W	Jun 10, 2004
8408227637	OL76833W	Jun 02, 2020
8408176005	OL76833W	Aug 29, 2017
8408099973	OL76833W	Jun 21, 2011
2709627914	OL76833W	2005-01-01
2253093009	OL76833W	Apr 01, 2015
0739443917	OL76833W	2003
1416529365	OL76833W	2006-11-30
1417664207	OL76833W	Jul 23, 2001
1933499834	OL76833W	Feb 09, 2009
9639526266	OL76833W	2003
1416580824	OL76833W	2009-03
9791600295	OL76833W	2006
9100110566	OL76833W	2006
1428124527	OL76833W	Nov 13, 2003
1442365471	OL76833W	Apr 16, 2013
1557048347	OL76833W	May 12, 2009
0743539451	OL76833W	Sep 01, 2003
0593057708	OL76833W	Aug 26, 2006
8184980450	OL76833W	Jan 01, 2009
074358046X	OL76833W	Mar 31, 2009
0743580451	OL76833W	Mar 31, 2009
0743597184	OL76833W	Mar 31, 2009
8408175742	OL76833W	2017-08
9584222546	OL14873315W	Apr 04, 2009
9752112676	OL14873315W	Nov 03, 2010
9049800424	OL14873315W	2010-07-01
9024589401	OL14873315W	\N
9021012812	OL14873315W	Nov 01, 2011
8466423125	OL14873315W	Aug 30, 2017
8497876806	OL14873315W	Nov 18, 2010
8497874587	OL14873315W	Nov 01, 2009
8497874579	OL14873315W	Mar 23, 2010
8497874528	OL14873315W	2009-10
8476698909	OL14873315W	Nov 01, 2009
9722526774	OL14873315W	May 01, 2013
030773692X	OL14873315W	Dec 01, 2009
8499307515	OL14873315W	Oct 07, 2015
8417031243	OL14873315W	Aug 30, 2017
8499302394	OL14873315W	Dec 16, 2010
8408099221	OL14873315W	Jan 11, 2011
8408176013	OL14873315W	Aug 29, 2017
3404160002	OL14873315W	Nov 05, 2011
3828995969	OL14873315W	Sep 25, 2010
1408460300	OL14873315W	2010
9661406707	OL14873315W	2010-01-01
0525565868	OL14873315W	Aug 21, 2018
0205133045	OL14873315W	\N
5170251955	OL14873315W	2015
0385533136	OL14873315W	2013
9722520148	OL14873315W	2009
2253134171	OL14873315W	2009
859929668X	OL14873315W	2010
5271303810	OL14873315W	2012
1616573058	OL14873315W	Oct 01, 2010
1615232168	OL14873315W	Nov 28, 2009
8983923385	OL14873315W	Jun 01, 2010
0804165742	OL14873315W	Jul 02, 2013
0132268310	OL14873315W	Sep 15, 2009
7020101577	OL14873315W	Mar 17, 2013
9601420452	OL14873315W	Mar 19, 2009
8408175750	OL14873315W	2017
7532150771	OL81634W	2013
8401499976	OL81634W	1999-05
8581052142	OL81634W	2014
1444741292	OL81634W	2011
3453051351	OL81634W	1991
2277231126	OL81634W	Jan 04, 1999
9875662690	OL81634W	2020
8868361558	OL81634W	2023
8376481118	OL81634W	Apr 14, 2009
8376486136	OL81634W	Feb 23, 2011
9024581818	OL81634W	Jun 27, 2018
8466345256	OL81634W	Oct 04, 2018
846634568X	OL81634W	Nov 08, 2018
2878626176	OL81634W	May 03, 2013
0606038590	OL81634W	Oct 01, 1990
0340390700	OL81634W	Oct 18, 1987
1797107313	OL81634W	Jan 07, 2020
1529311144	OL81634W	2019
2724247493	OL81634W	1990-01
0340923288	OL81634W	2000?
0340951435	OL81634W	Sep 27, 2007
0450417395	OL81634W	1988-11-01
1598878743	OL81634W	Apr 08, 2009
0606395687	OL81634W	Feb 28, 2017
0739476815	OL81634W	Feb 09, 1978
3548263135	OL81634W	Feb 27, 2005
0451153553	OL81634W	1988-06
5170255470	OL81634W	2004
8497595351	OL81634W	2008
3404118960	OL1914022W	1998-12
0688046592	OL1914022W	1989
8532527698	OL1914022W	Jan 01, 2012
0736616497	OL1914022W	Mar 23, 1989
033053517X	OL1914022W	Mar 23, 2010
0451168631	OL1914022W	Mar 23, 1989
151424716X	OL1914022W	Jun 05, 2015
1509886060	OL1914022W	\N
0230736076	OL1914022W	Mar 23, 1800
0330465732	OL1914022W	Mar 23, 2007
0330465724	OL1914022W	Mar 23, 1989
0671690841	OL1914022W	Nov 01, 1989
0451166892	OL1914022W	2001?
0330699814	OL1914022W	July 8, 1994
8804413476	OL1914022W	Sep 16, 1996
8496863247	OL1914022W	Jul 01, 2007
8401023866	OL1914022W	Jul 16, 2020
8401328519	OL1914022W	2000
753274843X	OL1914022W	2009
7532746852	OL1914022W	2009
3785710380	OL1914022W	November 1, 1999
3785710372	OL1914022W	November 1, 1999
3785713193	OL1914022W	June 1, 2003
0330450131	OL1914022W	2007
1405091045	OL1914022W	Oct 05, 2007
1509848495	OL1914022W	2017
1447265440	OL1914022W	2014
0330534920	OL1914022W	Aug 01, 2010
0143142380	OL1914022W	October 9, 2007
0330391984	OL1914022W	2000?
8422635755	OL1914022W	1998
0743527631	OL1914022W	July 1, 2002
0451207149	OL1914022W	2007?
9752121764	OL81613W	Nov 03, 2000
2277069043	OL81613W	1988
0451159276	OL81613W	1987-09
0451169514	OL81613W	1998?
8378858510	OL81613W	2014
8466357300	OL81613W	Oct 07, 2021
6073157061	OL81613W	Nov 12, 2014
9669931754	OL81613W	Sep 30, 2020
1623300894	OL81613W	Sep 20, 2016
9877253194	OL81613W	Mar 14, 2014
8868365626	OL81613W	2019
2253083364	OL81613W	Nov 02, 2017
9752119271	OL81613W	Jan 01, 2015
8466345345	OL81613W	Nov 23, 2017
3453504089	OL81613W	Aug 12, 2019
8466347925	OL81613W	Jun 20, 2019
1508297126	OL81613W	Jul 30, 2019
150824412X	OL81613W	Sep 05, 2017
9024516137	OL81613W	1986
9877250241	OL81613W	Mar 14, 2014
1982127791	OL81613W	2019-07
1444707868	OL81613W	2011
9500406977	OL81613W	June 1986
1501182099	OL81613W	2017-07
345343577X	OL81613W	Feb 08, 2011
345309994X	OL81613W	January 1, 1996
3548256112	OL81613W	January 1, 2003
0340951451	OL81613W	Jul 01, 2007
1508217114	OL81613W	2016-01-01
9754051518	OL81628W	Aug 01, 2004
840102143X	OL81628W	Sep 27, 2017
229034589X	OL81628W	Mar 08, 2006
8466341293	OL81628W	Jan 05, 2021
9875666270	OL81628W	Nov 03, 2011
8466336958	OL81628W	Feb 02, 2017
6073156243	OL81628W	Nov 02, 2017
9722526456	OL81628W	Dec 31, 2013
8495501627	OL81628W	Sep 16, 2000
0340723351	OL81628W	1998
0451160525	OL81628W	1989-07
8960172111	OL81628W	2009
5170770510	OL81628W	2013
8878244570	OL81628W	1994
0452279607	OL81628W	1988
0142800376	OL81628W	June 24, 2003
8440607466	OL81628W	June 1, 1998
0451210840	OL81628W	2003
0747401004	OL81628W	1989
3453029127	OL81628W	1992
0670032549	OL81628W	2003
1501143514	OL81628W	May 3, 2016
0340896213	OL81628W	2005
1508230749	OL81628W	Jun 27, 2017
1471227898	OL81628W	2013
9024526884	OL81628W	2001
902455585X	OL81628W	\N
9024516641	OL81628W	\N
8868363666	OL81628W	Jun 13, 2017
2738206581	OL81628W	1993-11
0606391622	OL81628W	2016
3453875567	OL81628W	Oct 05, 2003
3453873017	OL81628W	2003
1501166115	OL81628W	2017-07
8440220979	OL81628W	1982
836578131X	OL81628W	Jun 19, 2017
3404150554	OL76835W	2004-11
0671027387	OL76835W	2002-12
0552171360	OL76835W	2009
8408175769	OL76835W	Aug 29, 2017
9584261878	OL76835W	2017
987580889X	OL76835W	Nov 05, 2013
3404770579	OL76835W	Oct 31, 2005
3404175042	OL76835W	Sep 09, 2016
2709626411	OL76835W	2006-01
8957591567	OL76835W	2006
9752111300	OL76835W	Nov 03, 2005
9752105734	OL76835W	2007
9637318739	OL76835W	2005
0593057430	OL76835W	2005
0552151769	OL76835W	01/04/2004
837659026X	OL76835W	2005
837508199X	OL76835W	2005
8379997360	OL76835W	Apr 26, 2016
846721385X	OL76835W	2005
0743475437	OL76835W	2001
0552173541	OL76835W	2016
9100111899	OL76835W	2010
055216996X	OL76835W	2013
8791746485	OL76835W	2009
0552169978	OL76835W	2013
1846174589	OL76835W	Oct 01, 2006
1417751460	OL76835W	Apr 12, 2006
0606330178	OL76835W	Dec 27, 2004
1442365463	OL76835W	Apr 16, 2013
0743599101	OL76835W	Sep 14, 2010
0743536304	OL76835W	Jul 01, 2004
0743539478	OL76835W	Jun 01, 2004
074353946X	OL76835W	May 24, 2004
1402575440	OL76835W	Feb 12, 2004
074353574X	OL76835W	Nov 11, 2003
0743535758	OL76835W	Nov 11, 2003
1416528644	OL76835W	Feb 12, 2006
1627153128	OL81631W	\N
3453007867	OL81631W	1991
2290307769	OL81631W	Aug 30, 2000
2738204422	OL81631W	Sep 13, 1998
857302187X	OL81631W	1998
6073113870	OL81631W	Dec 04, 2014
8466357335	OL81631W	Oct 07, 2021
9877253747	OL81631W	Nov 02, 2020
0451150244	OL81631W	1987?
8401499844	OL81631W	1999-11
7020136575	OL81631W	2019
0743504356	OL81631W	Feb 01, 2001
2253151432	OL81631W	1985
3453879988	OL81631W	2004-05
1508278717	OL81631W	Feb 26, 2019
1529378303	OL81631W	2019-02-26
3453194667	OL81631W	2001-10
1444708139	OL81631W	2011
1501156705	OL81631W	2017
5170770863	OL81631W	2013
0385182449	OL81631W	1983
9754051526	OL81631W	Apr 13, 2003
1508226636	OL81631W	2019?
3453435796	OL81631W	Feb 08, 2011
3453504070	OL81631W	Mar 11, 2019
2724278488	OL81631W	1994-06
2724229916	OL81631W	1986-06
3548263100	OL81631W	2009
345321045X	OL81631W	October 1, 2002
0451139755	OL81631W	Nov 01, 1984
1529378311	OL81631W	2019
198211598X	OL81631W	2019
5170129467	OL81631W	2004
1982112395	OL81631W	Dec 04, 2018
0451162072	OL81631W	1989?
3453007042	OL81630W	1990
8466357130	OL81630W	May 20, 2021
5170720920	OL81630W	Jun 06, 2013
8374696192	OL81630W	Apr 13, 1979
9754059640	OL81630W	Nov 01, 2015
270961250X	OL81630W	1993
2709603047	OL81630W	1984
2709602180	OL81630W	1983
1501143816	OL81630W	2016
0451155750	OL81630W	1979
0451139720	OL81630W	1979
0451126661	OL81630W	1983?
0751504327	OL81630W	2000
0606019170	OL81630W	Dec 18, 1994
1786364670	OL81630W	Oct 18, 2020
1439507627	OL81630W	1980
0517321998	OL81630W	Sep 01, 1981
0812424719	OL81630W	Aug 01, 1980
9024544971	OL81630W	2002
2724278437	OL81630W	1994-06
8882742970	OL81630W	2001-09
8804297263	OL81630W	Oct 25, 1994
9024559499	OL81630W	2006-11
9024526868	OL81630W	\N
5770708417	OL81630W	1993
0451093380	OL81630W	1980-08
5170067399	OL81630W	2006
9573333511	OL81630W	2018
0340952687	OL81630W	Jul 02, 2008
780607225X	OL81630W	1998
1501144502	OL81630W	2018-08
8497871588	OL76836W	Nov 01, 2005
8497871596	OL76836W	Nov 01, 2005
8408176102	OL76836W	Aug 29, 2017
9875808903	OL76836W	Nov 02, 2013
841703126X	OL76836W	Aug 30, 2017
8499300278	OL76836W	Oct 01, 2009
0552171379	OL76836W	2009
9752111653	OL76836W	Nov 01, 2009
9752112439	OL76836W	Nov 03, 2010
8957591230	OL76836W	2005
8957591249	OL76836W	2005
8957591222	OL76836W	2005
7020101542	OL76836W	2013-12-01
404295510X	OL76836W	2009
4042955118	OL76836W	2009
8376590278	OL76836W	2008
8375081981	OL76836W	2008
8375088498	OL76836W	May 20, 2013
5170137761	OL76836W	2015
8373591729	OL76836W	2004
9953299129	OL76836W	2005
9024527910	OL76836W	2007-10
0752868926	OL76836W	Sep 26, 2004
073669689X	OL76836W	Feb 12, 2003
1439565775	OL76836W	Apr 09, 2009
9722514695	OL76836W	2006-03
8467219149	OL76836W	2006
2253127078	OL76836W	2014-04
2744192066	OL76836W	2006-06
0552151696	OL76836W	2004
0552169986	OL76836W	2013
0786113723	OL76836W	Jan 01, 2000
1427271283	OL76836W	Jun 09, 2015
1250064074	OL76836W	Jun 03, 2014
055217355X	OL76836W	2016
9024553024	OL76836W	2004
8389779013	OL76836W	2007
9754058997	OL81618W	Nov 01, 2015
2709610205	OL81618W	1991
2277223263	OL81618W	Feb 26, 2001
8581050549	OL81618W	Jan 01, 2013
0451160959	OL81618W	1987?
0451179285	OL81618W	1994?
8466357238	OL81618W	Jan 21, 2021
1623303281	OL81618W	Mar 17, 2020
1529370515	OL81618W	2021
8401498961	OL81618W	1992
340425242X	OL81618W	1995-01
3404281268	OL81618W	\N
060601828X	OL81618W	May 18, 1994
0593314018	OL81618W	2020
1529356547	OL81618W	\N
3785705859	OL81618W	\N
270961281X	OL81618W	May 19, 1993
9752121772	OL81618W	Apr 13, 2016
9752114075	OL81618W	Apr 13, 2012
0593313887	OL81618W	Dec 01, 2020
0449008606	OL81618W	Mar 06, 2012
5170276168	OL81618W	2005
0451924096	OL81618W	Aug 17, 1990
0340920955	OL81618W	Oct 24, 2006
1444720732	OL81618W	May 12, 2011
0307947300	OL81618W	2014?
143955787X	OL81618W	Oct 04, 2008
3453438183	OL81618W	Mar 08, 2016
038552885X	OL81618W	2012-08
5170109407	OL81618W	2002
0307987582	OL81618W	2012-02-14
0340358955	OL81618W	Jun 19, 1990
0606256156	OL81618W	Jun 28, 2011
0450537374	OL81618W	September 1999
8497599411	OL81618W	Nov 10, 2003
3453061187	OL77001W	1995
0385470819	OL77001W	1993-11
9022981037	OL77001W	1993
0440211727	OL77001W	1992-07
0307576132	OL77001W	2020
0922066728	OL77001W	1989
080412115X	OL77001W	2012?
0099590751	OL77001W	2013
2724289242	OL77001W	1995-08
5170702655	OL77001W	2011
5271350614	OL77001W	2011
5226037600	OL77001W	2011
5170883196	OL77001W	2015
8422655799	OL77001W	1995
2266072471	OL77001W	October 1997
8371691319	OL77001W	January 1, 1996
2266068520	OL77001W	1996
2266115499	OL77001W	2002-01
0440245915	OL77001W	2016?
8408012428	OL77001W	September 1995
8408020935	OL77001W	January 1, 1997
8408041258	OL77001W	July 2003
0385338600	OL77001W	2004 04
0385470789	OL77001W	October 1, 1993
9992191422	OL77001W	July 1992
0816155909	OL77001W	March 1993
0099201216	OL77001W	1992
0553470698	OL77001W	June 1, 1992
0553502220	OL77001W	October 6, 1998
0553712640	OL77001W	May 15, 2001
0553702203	OL77001W	May 15, 2001
8408033581	OL77001W	March 2000
0780735420	OL77001W	September 1993
555576751X	OL77001W	February 1993
5170127286	OL77001W	2003
0553745190	OL77001W	June 1, 1992
0375412204	OL46876W	1995
0034540288	OL46876W	Jul 16, 1995
140587967X	OL46876W	May 02, 2008
1664626875	OL46876W	Apr 29, 1999
8846203054	OL46876W	2003
0679419462	OL46876W	1995
0099637812	OL46876W	1996
0140816739	OL46876W	1997
034540288X	OL46876W	1996-10
274410700X	OL46876W	1997-05
902451956X	OL46876W	1997
4152079738	OL46876W	1995
4152079746	OL46876W	1995
8811602289	OL46876W	Apr 19, 2018
8576573059	OL46876W	Aug 22, 2016
1784752231	OL46876W	2015
8532506321	OL46876W	1996
0345538994	OL46876W	2012
3426193817	OL46876W	Aug 24, 1996
837169394X	OL46876W	1997
8401326508	OL46876W	1996
0788730932	OL46876W	1999
9500420104	OL46876W	October 1999
9500415844	OL46876W	1997-07
1400001129	OL46876W	January 22, 2002
2221081293	OL46876W	March 16, 1998
2266079123	OL46876W	January 1999
2266116061	OL46876W	March 2003
3442466814	OL46876W	Sep 08, 2008
8497597796	OL46876W	Jul 28, 2003
067945540X	OL46876W	April 7, 1997
0788737252	OL46876W	1999
0712676902	OL46876W	1995
0582402468	OL46876W	1998
0679765077	OL46876W	1995
8379856716	OL81609W	Apr 13, 2015
8376597434	OL81609W	\N
8376597426	OL81609W	\N
8376483730	OL81609W	Apr 13, 2010
8374698187	OL81609W	\N
6073120362	OL81609W	Mar 16, 2014
8401327431	OL81609W	1999
8447331652	OL81609W	2003
8372556474	OL81609W	2001
0613171004	OL81609W	1999
067102423X	OL81609W	1999
2226105662	OL81609W	2000
0340718196	OL81609W	1998
0684853507	OL81609W	1998
5170268580	OL81609W	2004
5170262841	OL81609W	2005
5170177305	OL81609W	2003
0671582348	OL81609W	Sep 01, 1998
1508293562	OL81609W	Jun 04, 2019
2702838022	OL81609W	1999-09
2744135933	OL81609W	2000-05
1501160451	OL81609W	\N
8573022469	OL81609W	\N
3453160819	OL81609W	November 1, 1999
3453210441	OL81609W	October 1, 2002
0340951427	OL81609W	May 31, 2007
0340921358	OL81609W	2006?
068483541X	OL81609W	Apr 06, 1999
1451678622	OL81609W	Dec 06, 2011
1982102497	OL81609W	Sep 04, 2018
1444720686	OL81609W	Aug 01, 2011
1501198890	OL81609W	2018
034071820X	OL81609W	Jun 24, 1999
8422676095	OL81609W	1998
1473695503	OL81609W	2018
8882742172	OL81609W	\N
3548268412	OL81612W	\N
8379858581	OL81612W	Apr 14, 2016
061323717X	OL81612W	2000
8820029073	OL81612W	\N
9024536448	OL81612W	Mar 21, 1999
0606183698	OL81612W	Apr 15, 2000
0340952385	OL81612W	Jul 01, 2007
3548252427	OL81612W	2002
9751016150	OL81612W	May 27, 2000
0671045857	OL81612W	Apr 01, 1999
3453441095	OL81612W	\N
3795117496	OL81612W	\N
8086278743	OL81612W	2000
3548840124	OL81612W	2004-01
9751037190	OL81612W	Apr 13, 2017
0671045865	OL81612W	Apr 01, 1999
8447331679	OL81612W	2003
5170030681	OL81612W	2006
5237029493	OL81612W	2002
5170250843	OL81612W	2004
1501157515	OL81612W	Apr 25, 2017
3548251285	OL81612W	May 1, 2001
0743563395	OL81612W	1999 04
3785711131	OL81612W	January 1, 2001
3548255744	OL81612W	October 1, 2002
3795117569	OL81612W	2000
3785711123	OL81612W	January 1, 2001
1501192280	OL81612W	May 15, 2018
2226115234	OL81612W	2000
225315136X	OL81612W	2007-11
3795117402	OL81612W	January 1, 2000
8497593677	OL81612W	2003
2290032433	OL81625W	Jul 31, 1997
2290321125	OL81625W	2004
8401021413	OL81625W	Oct 10, 2017
8466340718	OL81625W	Jan 11, 2008
9875666963	OL81625W	Nov 02, 2013
8440220995	OL81625W	1991
9752103596	OL81625W	Jan 01, 2000
345301216X	OL81625W	\N
3453875583	OL81625W	Aug 09, 2003
9722528866	OL81625W	Jan 26, 2015
8573026693	OL81625W	2005
0451173317	OL81625W	1993-01
3453053397	OL81625W	1995
0142800392	OL81625W	August 26, 2003
0452267404	OL81625W	1992
5237014852	OL81625W	1999
702005496X	OL81625W	2006
8466300228	OL81625W	2000
0451210867	OL81625W	2003-09
3453123867	OL81625W	November 1, 1997
837985585X	OL81625W	Jan 01, 2015
0340832258	OL81625W	2003
0606391649	OL81625W	May 03, 2016
1508217475	OL81625W	2016-01-01
1501143549	OL81625W	May 03, 2016
517027100X	OL81625W	2005
753214920X	OL81625W	2013
0452279623	OL81625W	1997?
093798616X	OL81625W	Sep 23, 1991
9993574864	OL81625W	January 1992
034082977X	OL81625W	2002?
8440644302	OL81625W	1994-04
2277232432	OL81625W	1992
2277250341	OL81625W	1998-03
0747411875	OL81625W	November 12, 1992
0785705430	OL81625W	October 1999
1417637137	OL81625W	September 2003
8868363690	OL81615W	Jun 13, 2017
8373590250	OL81615W	2003
8401021456	OL81615W	May 25, 2018
8466342656	OL81615W	May 10, 2017
607313021X	OL81615W	Feb 01, 2015
8499892604	OL81615W	Nov 17, 2011
0451194861	OL81615W	2001?
0452284724	OL81615W	2003-07
0451210875	OL81615W	2004?
0452279178	OL81615W	1997-11
0340702141	OL81615W	\N
2290053139	OL81615W	1999
0142800406	OL81615W	September 30, 2003
0340696613	OL81615W	1997
5237008127	OL81615W	1999
1444723472	OL81615W	2012
3453147596	OL81615W	February 1, 1999
1880418371	OL81615W	Sep 27, 1997
8379855868	OL81615W	Jan 01, 2016
0606391657	OL81615W	May 03, 2016
0340829788	OL81615W	2003
0340712937	OL81615W	Aug 25, 1997
7532149196	OL81615W	2013
1508217505	OL81615W	2016-01
0613090993	OL81615W	October 1999
3453138783	OL81615W	1997
227725035X	OL81615W	November 1, 1998
8440690134	OL81615W	November 2001
9022995593	OL77014W	2009
9022983994	OL77014W	1998
0099244926	OL77014W	1998
274413743X	OL77014W	2000
0712678212	OL77014W	1998
0385490992	OL77014W	1998-03
5170968531	OL77014W	2016
0440794234	OL77014W	January 5, 1999
9176436160	OL77014W	1998
5170205414	OL77014W	2004
3598800061	OL77014W	October 1, 2002
3453187296	OL77014W	2001
3455025013	OL77014W	February 1, 1999
3453165802	OL77014W	September 1, 1999
3453169247	OL77014W	August 2000
8466300627	OL77014W	May 31, 2001
8440694377	OL77014W	1999
8483468794	OL77014W	2009
8496546381	OL77014W	September 2006
8440681798	OL77014W	October 2000
2266107275	OL77014W	May 2001
2251443053	OL77014W	2006
2221087143	OL77014W	September 16, 1999
2266145428	OL77014W	March 2005
1568656718	OL77014W	1998
0553502123	OL77014W	February 4, 1998
0440295602	OL77014W	October 13, 1998
0440245958	OL77014W	October 26, 2010
0375433473	OL77014W	September 6, 2005
0553479180	OL77014W	February 4, 1998
8440669437	OL77014W	May 1999
0440225701	OL77014W	1999
0712679332	OL77014W	\N
5170156219	OL77014W	2003
0739312170	OL77014W	January 6, 2004
0385339097	OL77014W	April 26, 2005
8804444495	OL77014W	1998
3550101244	OL23480W	September 1, 2001
3550101252	OL23480W	September 1, 2001
2744155535	OL23480W	2002-07
2266145444	OL23480W	2005-05
6045322231	OL23480W	2014
0440206154	OL23480W	1999?
3453198638	OL23480W	September 1, 2001
3453861582	OL23480W	November 1, 2002
8440630123	OL23480W	July 1993
844063014X	OL23480W	April 1995
9994122304	OL23480W	March 1993
0743527070	OL23480W	September 1, 2002
0517494264	OL23480W	July 21, 1985
0743527054	OL23480W	September 1, 2002
2863740946	OL23480W	May 1, 1982
0708981690	OL23480W	June 1984
067167854X	OL23480W	March 15, 1989
0712653570	OL23480W	September 12, 1991
0899668771	OL23480W	June 1991
0099111519	OL23480W	March 4, 1993
2266103679	OL23480W	August 2002
8804328088	OL23480W	1993
3453025423	OL23480W	July 2002
2226120432	OL23480W	2000-11
0307344681	OL23480W	November 1, 2005
345319862X	OL23480W	2001
8439705158	OL23480W	December 2002
8497594924	OL23480W	2003
9724607046	OL23480W	1995
0553227467	OL23480W	1982-10
0552121606	OL23480W	1983
1440657750	OL23480W	2000
0370304489	OL23480W	1982
1585472379	OL23480W	2002
2253013099	OL3454854W	1976
2231002116	OL3454854W	1976-01-01
1611297222	OL3454854W	2005
055308500X	OL3454854W	1974-01-01
7020122639	OL3454854W	2018-01-01
0449219631	OL3454854W	Jul 30, 1991
8428603987	OL3454854W	1973
8428603995	OL3454854W	1974
3548033202	OL3454854W	1980-10
7805672040	OL3454854W	2000-02
8401490049	OL3454854W	June 1983
8428604762	OL3454854W	June 1984
9507426779	OL3454854W	December 1995
0345544145	OL3454854W	2013
1557046778	OL3454854W	June 14, 2005
0233965106	OL3454854W	1974
0330243829	OL3454854W	1975
0553204653	OL3454854W	July 1, 1981
1417653078	OL3454854W	July 30, 1991
060601179X	OL3454854W	1991 08
0140816682	OL3454854W	December 1999
075310833X	OL3454854W	April 1, 2000
1400064562	OL3454854W	2005
0385047711	OL3454854W	1974
0671754327	OL36626W	1970
0862251419	OL36626W	1972-01-01
2859404511	OL36626W	1996
2859405321	OL36626W	May 28, 1998
0140017232	OL36626W	1970
0385047258	OL36626W	Jun 01, 1957
0884111490	OL36626W	April 1991
081221725X	OL36626W	2000
0708980740	OL36626W	June 1982
089190154X	OL36626W	June 1977
0330245899	OL36626W	1975
009986620X	OL36626W	1992
0881844098	OL36626W	June 1988
0892330376	OL36626W	January 1, 1976
0745141323	OL36626W	June 1993
0575029218	OL36626W	January 1981
6058072735	OL3899224W	Apr 13, 2019
0007232837	OL3899224W	2006
9022912892	OL3899224W	1973
9044912895	OL3899224W	1985
349922951X	OL3899224W	August 1, 2000
2743618043	OL3899224W	2008
0752856138	OL3899224W	March 18, 2004
8490567077	OL3899224W	Feb 04, 2016
8490060096	OL3899224W	May 01, 2011
1433263262	OL3899224W	May 01, 2012
1433263270	OL3899224W	Oct 06, 2009
089190378X	OL3899224W	Jun 27, 1980
0007242972	OL3899224W	Apr 07, 2007
3499425882	OL3899224W	Jun 03, 1990
3688107241	OL3899224W	Nov 09, 2017
0553065491	OL3899224W	1974
0330233513	OL3899224W	1972
3499244411	OL3899224W	Oct 01, 2008
0007930666	OL3899224W	2012-01-01
8491875468	OL3899224W	Mar 11, 2021
8411322246	OL3899224W	Dec 15, 2022
9117820324	OL3899224W	1978
0007944543	OL3899224W	2014-10-01
0394717791	OL3899224W	March 12, 1976
8466410708	OL3899224W	May 01, 2009
1433249073	OL3899224W	Oct 14, 2008
1433249065	OL3899224W	Oct 14, 2008
8490566356	OL3899224W	Nov 12, 2015
9997531779	OL3899224W	Jun 17, 1967
2264006986	OL3899224W	1985
034075494X	OL81624W	Sep 16, 1999
0340738901	OL81624W	1999
0671042149	OL81624W	Dec 15, 2000
0756905753	OL81624W	Aug 01, 2000
1844561607	OL81624W	\N
8497592956	OL81624W	2002
0606194967	OL81624W	Oct 18, 2000
1417739738	OL81624W	Aug 01, 2000
1501160478	OL81624W	\N
0671024248	OL81624W	2000-08
0684844907	OL81624W	Sep 14, 1999
3453185668	OL81624W	February 1, 2002
0671582356	OL81624W	September 14, 1999
0606286357	OL81624W	August 2000
3453435710	OL81624W	Feb 08, 2011
0340952393	OL81624W	Jan 01, 2007
3453182669	OL81624W	2000
3453159926	OL81624W	December 31, 1998
1501195972	OL81624W	2017-11
0743436210	OL81624W	2001
0743563352	OL81624W	2001 09
8401327865	OL81624W	1999
5170095619	OL81624W	January 1, 2005
8401013011	OL81624W	July 2000
034073891X	OL81624W	2000
0743509870	OL81624W	September 1, 2001
0743509862	OL81624W	September 1, 2001
3453177487	OL81624W	October 2001
2253151408	OL81624W	2001
2226122095	OL81624W	2001-12
8401328454	OL81624W	June 26, 2000
8484503119	OL81624W	January 2002
9871138504	OL81624W	April 2004
078388737X	OL81624W	1999
0684853515	OL81624W	1999
9573258552	OL81624W	2006
0340818670	OL81624W	February 21, 2002
1400094518	OL81586W	2004-02
0743417682	OL81586W	2003-12
8374696303	OL81586W	\N
8378393550	OL81586W	2002
6073113862	OL81586W	Nov 02, 2013
9024539145	OL81586W	2002
0340770708	OL81586W	May 29, 2003
5170187289	OL81586W	2004
5170230044	OL81586W	2006
0743520955	OL81586W	Sep 01, 2002
0340952660	OL81586W	Nov 01, 2007
1439568103	OL81586W	2003
1501160435	OL81586W	\N
0743228472	OL81586W	2002
0340792345	OL81586W	2002
355008353X	OL81586W	March 1, 2002
0708949606	OL81586W	2003
1444753673	OL81586W	2011
8447331601	OL81586W	2003
2226150765	OL81586W	2004-02
9752103693	OL81586W	2003
8573025506	OL81586W	2003
8374690844	OL81586W	2003
1501192191	OL81586W	2017
0743520963	OL81586W	2002
8497930843	OL81586W	Jan 23, 2004
9875660175	OL81586W	October 2004
1400003148	OL81586W	November 19, 2002
8401329647	OL81586W	October 15, 2002
9506440182	OL81586W	December 2002
0684017148	OL81586W	September 2002
1402518862	OL81586W	2002
0743563336	OL81586W	2002 09
0743235959	OL81586W	September 2002
0613707397	OL81586W	December 2003
0340671785	OL2794726W	1996
902452606X	OL2794726W	Dec 22, 1999
6073178093	OL2794726W	Feb 12, 2019
8497595947	OL2794726W	2004
2253151505	OL2794726W	2004
034095227X	OL2794726W	May 14, 2007
2744109231	OL2794726W	1996
0451191013	OL2794726W	1997
0451191927	OL2794726W	December 1, 2002
0525941908	OL2794726W	1996 10
0613096207	OL2794726W	October 1999
9751012856	OL2794726W	1998
0786208457	OL2794726W	1996
3453186672	OL2794726W	2001
1508218315	OL2794726W	2016 02
0788711636	OL2794726W	1997
0525942246	OL2794726W	October 1, 1996
3548255019	OL2794726W	October 1, 2002
3453129601	OL2794726W	1997
2226088083	OL2794726W	1996
0451983963	OL2794726W	September 1, 1997
2290306681	OL2794726W	July 6, 2000
8882741591	OL2794726W	2001-05-22
8484502546	OL2794726W	May 2004
8447334260	OL2794726W	2007
8401474701	OL2794726W	January 1999
1501143751	OL2794726W	2016
5170890249	OL2794726W	1997
837186065X	OL2794726W	Jun 19, 1999
3453435818	OL2794726W	Feb 08, 2011
9751037514	OL2794726W	Apr 13, 2017
0340671769	OL2794726W	1996
1508218676	OL2794726W	Sep 27, 2016
1501144278	OL2794726W	2016
0788191799	OL2794726W	Dec 04, 1996
0143143913	OL2794726W	2007
9501510085	OL510879W	1990
8490708835	OL510879W	Apr 25, 2019
221302524X	OL510879W	December 1, 1990
2253048593	OL510879W	November 1991
2863742566	OL510879W	1987
229800448X	OL510879W	2007
0571229603	OL510879W	2005-01-01
8804298626	OL510879W	1987
0345430581	OL510879W	November 28, 1998
0747403465	OL510879W	October 6, 1988
0745160638	OL510879W	1991
0676971881	OL510879W	2011
039455583X	OL510879W	1986
1400096472	OL510879W	November 8, 2005
1400025109	OL510879W	September 25, 2007
0140129545	OL510879W	April 1995
0345469380	OL510879W	November 4, 2003
039456331X	OL510879W	April 1991
0886191424	OL510879W	1986
0886191440	OL510879W	1986
0816195765	OL510879W	March 1991
0446776890	OL510879W	September 1992
0816142653	OL510879W	1987
0816142661	OL510879W	1987
0722152221	OL510879W	1987
0446323527	OL510879W	1987
0571137997	OL510879W	1986
0571145701	OL510879W	1986
8486311896	OL510879W	1986
844063269X	OL510879W	October 5, 1999
3763235892	OL510879W	\N
3426306972	OL510879W	\N
3426605392	OL510879W	May 13, 1996
3828967027	OL510879W	1999-01-01
3426191997	OL510879W	1988-01-01
8804390379	OL510879W	1994
3426614383	OL510879W	February 1, 1999
3426031159	OL510879W	1988
3426624869	OL510879W	May 1, 2003
8376597221	OL81594W	Apr 14, 2013
2290332488	OL81594W	Aug 31, 2004
2290332461	OL81594W	Dec 07, 2006
0743251628	OL81594W	2005
8373591427	OL81594W	2004
3453530233	OL81594W	2004-12
9752104541	OL81594W	Apr 13, 2017
179710635X	OL81594W	Nov 19, 2019
8401335299	OL81594W	2004
8379855876	OL81594W	Jan 01, 2017
7020062369	OL81594W	2007
1508219168	OL81594W	2017?
1439565929	OL81594W	Oct 20, 2008
517023645X	OL81594W	2004
7532149358	OL81594W	2013
141651693X	OL81594W	2003
0340836156	OL81594W	2005
0340827165	OL81594W	January 3, 2005
0743561694	OL81594W	2006
0340827173	OL81594W	\N
3453874145	OL81594W	2003
1417645687	OL81594W	January 2005
9685960836	OL81594W	\N
0743533526	OL81594W	November 4, 2003
0743533518	OL81594W	November 4, 2003
1880418568	OL81594W	2003
5551280098	OL81594W	2003
0340827157	OL81594W	2003
8496581985	OL553754W	Oct 18, 2006
3548256104	OL553754W	January 1, 2003
3548250203	OL553754W	January 1, 2001
3548254896	OL553754W	August 1, 2002
8498721970	OL553754W	Mar 18, 2009
8466617027	OL553754W	August 30, 2005
8466302026	OL553754W	February 2001
9045013231	OL553754W	2006
2743615877	OL553754W	2006
2869303912	OL553754W	December 31, 1998
2869301537	OL553754W	March 16, 2001
0739346792	OL553754W	2006-08-15
8490701040	OL553754W	2015
0445405252	OL553754W	1988
0446698873	OL553754W	2006
0099498103	OL553754W	1989
071261995X	OL553754W	1988
0099419335	OL553754W	February 17, 1994
0736618163	OL553754W	September 5, 1990
0099366517	OL553754W	November 3, 2005
0446618128	OL553754W	September 1, 2006
0892962062	OL553754W	1987
0739473603	OL553754W	2006
0099508516	OL553754W	\N
8440636881	OL553754W	1993
0446674362	OL553754W	1998
9867058410	OL553754W	2006
0739323881	OL553754W	August 29, 2006
0099492164	OL553754W	February 2, 2006
0786293101	OL553754W	February 2007
8376483498	OL14917748W	Apr 12, 2010
9024564344	OL14917748W	Oct 04, 2013
8868360276	OL14917748W	Jul 03, 2013
225316979X	OL14917748W	Mar 06, 2013
8499891098	OL14917748W	May 07, 2010
9875669091	OL14917748W	Nov 02, 2013
1439148503	OL14917748W	2009
9661410259	OL14917748W	Sep 22, 2011
1501156799	OL14917748W	\N
1476735476	OL14917748W	Jun 11, 2013
0593311582	OL14917748W	Nov 16, 2021
9752113249	OL14917748W	Apr 13, 2011
0340992565	OL14917748W	2009
1594134170	OL14917748W	Jul 06, 2010
0340992573	OL14917748W	Aug 19, 2009
1442345500	OL14917748W	Nov 08, 2011
1442365498	OL14917748W	Jun 11, 2013
3453435230	OL14917748W	2011
1439149038	OL14917748W	2010-07
1476743940	OL14917748W	2013
0307741125	OL14917748W	2010-07
1410423964	OL14917748W	2009
1456112503	OL14917748W	2010 December 01
1439156972	OL14917748W	2009 November 10
0743597311	OL14917748W	2009 November
0340992581	OL14917748W	2013?
8846200039	OL46913W	1997
1501216791	OL46913W	Oct 06, 2015
0345378490	OL46913W	1993-01
9578701241	OL46913W	1995
2253030031	OL46913W	1982-06
2266068385	OL46913W	Aug 26, 1995
2264013974	OL46913W	1994-03
3499263424	OL46913W	May 1, 2001
1501216775	OL46913W	2015-10-06
0679431136	OL46913W	Nov 02, 1993
0375402977	OL46913W	April 20, 1998
8422658097	OL46913W	1995
0307816508	OL46913W	2012-05
8532512356	OL46913W	2001
0394513924	OL46913W	1980
0099544318	OL46913W	July 6, 1995
0752904299	OL46913W	1995
0099320819	OL46913W	1993
0712659110	OL46913W	1993
034541893X	OL46913W	June 23, 1997
1417617330	OL46913W	November 2003
8401492343	OL46913W	June 1994
0712676058	OL46913W	\N
081613202X	OL46913W	1981
0613100557	OL46913W	October 1999
0060541830	OL46913W	October 28, 2003
014005863X	OL46913W	1981
8484502902	OL46913W	June 1994
0713914165	OL46913W	1981
8371695209	OL46913W	1994
0708989128	OL46913W	May 2000
0061782556	OL46913W	2009 05
9500415275	OL46913W	1996
0099682710	OL46913W	August 31, 1995
038056176X	OL46913W	1981-10
9024523656	OL46913W	1995
030758836X	OL16239762W	2012
1410450953	OL16239762W	2012
0307588386	OL16239762W	2012
0553418351	OL16239762W	2014
0753827662	OL16239762W	2013
0345805461	OL16239762W	2013-05
359652072X	OL16239762W	Jul 23, 2015
3596188784	OL16239762W	Jul 21, 2014
1524763675	OL16239762W	May 22, 2018
0385366752	OL16239762W	Jul 24, 2012
1780228228	OL16239762W	2014
0297859382	OL16239762W	2012
159413605X	OL16239762W	2014
3502102228	OL16239762W	\N
3596032199	OL16239762W	Apr 22, 2014
9000914132	OL16239762W	2014
0385347774	OL16239762W	May 24, 2012
2253164917	OL16239762W	2013
2298071179	OL16239762W	\N
2355841179	OL16239762W	2012-08
1471306984	OL16239762W	2012
9022578712	OL16239762W	2017
7508639197	OL16239762W	2014
8377782979	OL16239762W	Jun 04, 2013
5389051602	OL16239762W	2013
0307588378	OL16239762W	2014
0446614319	OL167179W	2004-04
0755300297	OL167179W	2004
0446610224	OL167179W	2004-10
0739437062	OL167179W	2003
8373592091	OL167179W	2005
1478938056	OL167179W	Oct 20, 2015
0446692573	OL167179W	Sep 22, 2004
0755309774	OL167179W	Sep 22, 2003
1435291034	OL167179W	May 29, 2008
1843956292	OL167179W	2005
0755334221	OL167179W	Sep 22, 2004
9022989151	OL167179W	2005
2253123056	OL167179W	2008
0755381270	OL167179W	2014
2709626918	OL167179W	2006
2744199311	OL167179W	2006
0755349377	OL167179W	2001
0316602906	OL167179W	2003
1478938064	OL167179W	Oct 20, 2015
141766715X	OL167179W	October 2004
5558595272	OL167179W	January 2001
1586215809	OL167179W	November 17, 2003
0736698647	OL167179W	2003
075530022X	OL167179W	\N
0316743844	OL167179W	2003
1586215795	OL167179W	November 17, 2003
0755300211	OL167179W	October 13, 2003
9044313363	OL22914W	2005
9752106951	OL22914W	Nov 03, 2006
1478938269	OL22914W	Nov 17, 2015
837359311X	OL22914W	2006
2749902878	OL22914W	2005-05
1455578614	OL22914W	Mar 31, 2015
1455581771	OL22914W	Mar 31, 2015
7802252792	OL22914W	2007-05
5170391447	OL22914W	2007
1600240941	OL22914W	November 6, 2007
5558595264	OL22914W	January 2001
1586217089	OL22914W	November 1, 2004
1415904936	OL22914W	November 2004
0316858471	OL22914W	December 2, 2004
0316858498	OL22914W	2004
044617792X	OL22914W	2007 11
1586217097	OL22914W	November 1, 2004
0446576638	OL22914W	2004-11
1415904944	OL22914W	November 2004
0751531804	OL22914W	July 11, 2005
1586217100	OL22914W	November 1, 2004
0446616621	OL22914W	2005 11
0446577146	OL22914W	2004-11
0739449079	OL22914W	2004
9022989461	OL41256W	\N
902299662X	OL41256W	2010-01-01
0446698628	OL41256W	2006
8373594086	OL41256W	2008
3404164792	OL41256W	2010
3785723547	OL41256W	2008
0739475622	OL41256W	2006
0446400327	OL41256W	Apr 24, 2007
044653109X	OL41256W	2006-10
9022994643	OL41256W	2008
8498720958	OL41256W	2008-05
1447274296	OL41256W	2014
0330523511	OL41256W	2011-04
1405090111	OL41256W	Jun 02, 2006
1598953834	OL41256W	2006
0330444085	OL41256W	2007
0446615633	OL41256W	2007 09
1594835802	OL41256W	October 17, 2006
1594835829	OL41256W	October 17, 2006
0759569010	OL41256W	2006-10
1594835845	OL41256W	October 17, 2006
0446580198	OL41256W	October 17, 2006
1405089849	OL41256W	2006
846663519X	OL41256W	2007
0755330390	OL167155W	2007
0316013943	OL167155W	2007-02
0446199273	OL167155W	2008-06
1472258614	OL167155W	\N
0446698458	OL167155W	2007
5170790090	OL167155W	2013
1405649313	OL167155W	2008
0755330412	OL167155W	2008
1478963328	OL167155W	Jan 12, 2016
1594836272	OL167155W	February 12, 2007
0316017752	OL167155W	February 6, 2007
1600242790	OL167155W	June 3, 2008
1594836256	OL167155W	February 12, 2007
159483623X	OL167155W	February 12, 2007
0755330404	OL167155W	February 8, 2007
1587672316	OL15168588W	Dec 01, 2010
9023425545	OL15168588W	2016
0606321829	OL15168588W	Jul 31, 2012
0385669518	OL15168588W	Jun 08, 2010
1400026253	OL15168588W	Jul 31, 2012
0345516869	OL15168588W	2012
6050905940	OL15168588W	Oct 18, 2012
3442469376	OL15168588W	2012-01
8804606371	OL15168588W	2011-03
2266218573	OL15168588W	2013-03
2221111133	OL15168588W	2011-03
8415139292	OL15168588W	2012
0345504976	OL15168588W	2012
0385693702	OL15168588W	2018
0752883305	OL15168588W	2012?
0525618759	OL15168588W	2018
0525618740	OL15168588W	2018-12
1984886851	OL15168588W	Dec 31, 2018
0752897845	OL15168588W	2010
1409128512	OL15168588W	2012
1409190986	OL15168588W	2019
1409102300	OL15168588W	Jan 01, 2011
0449806944	OL15168588W	Sep 11, 2012
0345525221	OL15168588W	2010
140910334X	OL15168588W	2010
0752897853	OL15168588W	2010
0345528174	OL15168588W	2012-07
0345504968	OL15168588W	2010
1410432874	OL15168588W	2010
8489367876	OL15168588W	2010-10
8483469103	OL46904W	2009-03-01
0061673536	OL46904W	Nov 18, 2008
0061350206	OL46904W	2007
0732283639	OL46904W	Nov 15, 2007
0307391981	OL46904W	Feb 05, 2008
166506417X	OL46904W	Mar 09, 2021
2266242245	OL46904W	Dec 28, 2008
0060872985	OL46904W	2006
2266182048	OL46904W	2009-01-06
006222719X	OL46904W	2013
0007254903	OL46904W	2006-11-28
0007330626	OL46904W	2009-05
0007459955	OL46904W	Mar 01, 2012
8401336406	OL46904W	2007-10
0061284319	OL46904W	November 28, 2006
0007248997	OL46904W	November 28, 2006
0060873167	OL46904W	2007-12
5699247173	OL46904W	2007
0060873035	OL46904W	2006
0060873094	OL46904W	November 28, 2006
006087306X	OL46904W	November 28, 2006
8934926198	OL46904W	2007
9573261162	OL46904W	2007
8495618850	OL167160W	2005
0446613371	OL167160W	2007-01
0739449672	OL167160W	2005
8580411246	OL167160W	2013
1843959135	OL167160W	2005
2253118931	OL167160W	2007
2841877671	OL167160W	2006-01
7538271864	OL167160W	2005
3442459079	OL167160W	2006
0755305779	OL167160W	2006
0755305760	OL167160W	2005
0316710628	OL167160W	2005-02
0446696269	OL167160W	2006-01
1478963352	OL167160W	Jan 12, 2016
1586217267	OL167160W	February 1, 2005
0755305752	OL167160W	[date missing]
0316009563	OL167160W	February 14, 2005
0755324811	OL167160W	January 31, 2005
1415907943	OL167160W	May 2005
1586217275	OL167160W	February 1, 2005
1415907935	OL167160W	May 2005
0446694215	OL167160W	February 1, 2005
555859523X	OL167160W	January 2001
0330200747	OL15008W	1974
202006376X	OL15008W	February 28, 1983
8437617499	OL15008W	June 30, 2004
5702701976	OL15008W	1996
0099470470	OL15008W	February 5, 2004
9573100835	OL15008W	1989
5835202261	OL15008W	1993
009974371X	OL15008W	1998
0316290238	OL15008W	August 4, 1997
044031335X	OL15008W	\N
0330295683	OL15008W	1986
0224602179	OL15008W	1971
0586044264	OL15008W	1976
0965058484	OL15008W	1900
0745127703	OL15008W	December 1996
0330309900	OL15008W	1989
076072539X	OL15008W	2002
0606252819	OL15008W	December 1997
1850890293	OL15008W	March 1986
0440113350	OL15008W	1979
0316290963	OL15008W	Jun 01, 1963
354860224X	OL15008W	August 1, 2002
8401490316	OL15008W	June 1983
8476695845	OL15008W	June 30, 2005
0739433326	OL167161W	2003
0316147877	OL167161W	2003
0446613843	OL167161W	2004-02
0755300203	OL167161W	2004
1472258622	OL167161W	\N
3404156234	OL167161W	2007
184395222X	OL167161W	2004
0755349466	OL167161W	2011
486332720X	OL167161W	2004
5170377436	OL167161W	2007
0316602086	OL167161W	March 2003
0316602078	OL167161W	March 2003
075530019X	OL167161W	2003
0316602051	OL167161W	2003
0755300181	OL167161W	2003
1586214144	OL167161W	March 2003
061392519X	OL167161W	February 2004
1586215353	OL167161W	March 1, 2003
1586214152	OL167161W	March 2003
1586215361	OL167161W	March 1, 2003
0446177873	OL41249W	2007-11
1447272307	OL41249W	2007-01-01
1509836489	OL41249W	2016
3404160800	OL41249W	2011-10
0446598119	OL41249W	2007-11
8360192588	OL41249W	2008
849872872X	OL41249W	Sep 25, 2013
9022993418	OL41249W	2007
0230017754	OL41249W	2007
0230017797	OL41249W	2007
0446615641	OL41249W	2008-09
0446577391	OL41249W	2007-11
2749912660	OL41249W	2010-09
144727430X	OL41249W	2014
1447226577	OL41249W	Jul 31, 2011
033052352X	OL41249W	2011
0330450980	OL41249W	2008
8466637443	OL41249W	2009-06
0446195103	OL41249W	2007 11
1602522324	OL41249W	\N
1600240488	OL41249W	November 6, 2007
1600240526	OL41249W	November 6, 2007
3548288626	OL181821W	Mar 14, 2016
345308201X	OL181821W	November 1, 1994
3462022776	OL181821W	July 1, 1993
1489358676	OL181821W	Jul 28, 2016
0340937688	OL181821W	Apr 16, 2011
0241247527	OL181821W	2016
0241291259	OL181821W	2016
1524796956	OL181821W	Aug 15, 2017
0399594000	OL181821W	Mar 29, 2016
0345418301	OL181821W	June 23, 1997
0340600888	OL181821W	June 24, 1993
5552758609	OL181821W	August 1994
0345385764	OL181821W	June 1, 1994
0517137631	OL181821W	January 17, 1995
0517396165	OL181821W	May 11, 1999
067942508X	OL181821W	July 6, 1993
0736627898	OL181821W	July 1, 1994
5557103368	OL181821W	January 1993
0340766522	OL181821W	January 20, 2000
0754053598	OL181821W	\N
1858918960	OL181821W	December 12, 1996
0340597658	OL181821W	1994
0340592818	OL181821W	1993
0345480325	OL181821W	December 28, 2004
0745143539	OL181821W	October 1994
0679425136	OL181821W	1993
0679747281	OL181821W	1993
9867475534	OL181821W	2005
0670851191	OL181821W	1993
0708988148	OL181821W	1995
9032510924	OL134880W	2007
9032507133	OL134880W	1990-01-01
2277234648	OL134880W	1994-08-25
0671670689	OL134880W	1990-11
2724278143	OL134880W	1994-06
1416511849	OL134880W	2002
0671717324	OL134880W	1991
2277021520	OL134880W	1992-05
1982113197	OL134880W	2019-02
075311772X	OL134880W	July 2003
2290303909	OL134880W	2000
0743440269	OL134880W	August 5, 2002
0753122316	OL134880W	January 2003
0816151849	OL134880W	1991
0816151865	OL134880W	1991
5553819318	OL134880W	August 1991
067171550X	OL134880W	1993
5557098127	OL134880W	January 1990
8401493323	OL134880W	1991
0671670670	OL134880W	1990-11
0671717405	OL134880W	1991
0671853511	OL134880W	January 1, 1994
0833558935	OL134880W	October 1999
8401497590	OL134880W	1996 10
1416511865	OL134885W	2002
9032510932	OL134885W	2007
2277021547	OL134885W	1992-09
8401324882	OL134885W	1993-01
8401493331	OL134885W	June 1994
8401497604	OL134885W	1997
1400000734	OL134885W	October 30, 2001
3442412226	OL134885W	April 1, 1992
9754052964	OL134885W	2003
5557103023	OL134885W	January 1993
999484928X	OL134885W	November 1994
0671695126	OL134885W	1991-06
0671695134	OL134885W	1991
0671715798	OL134885W	1993
9602460881	OL134885W	1991
0816153868	OL134885W	1992
0753122766	OL134885W	February 2004
0833574256	OL134885W	October 1999
0745174825	OL134885W	1993
2277235806	OL134885W	1993-11-05
0671717472	OL134885W	1991
0743440277	OL134885W	August 5, 2002
0671717413	OL134885W	1991
2290303917	OL134885W	October 10, 2000
0753117738	OL134885W	January 2003
8489367108	OL167162W	2006
0755325672	OL167162W	2005
044661761X	OL167162W	2006-08
2253123285	OL167162W	2009-09
1472258525	OL167162W	\N
1478963689	OL167162W	Mar 15, 2016
9022991792	OL167162W	2006
0755325699	OL167162W	2006
0755325680	OL167162W	2005
031610695X	OL167162W	2005-07
0316057851	OL167162W	2005-07
1594830479	OL167162W	July 11, 2005
0755325710	OL167162W	June 20, 2005
1600242537	OL167162W	November 13, 2007
1415923248	OL167162W	July 11, 2005
0755325702	OL167162W	June 20, 2005
8377586533	OL15165640W	2014
229803009X	OL15165640W	2010-05
229802619X	OL15165640W	Jan 22, 2009
0385528701	OL15165640W	2009
8408163361	OL15165640W	Oct 11, 2016
2266194232	OL15165640W	2010-12
0307455378	OL15165640W	2013?
8804583355	OL15165640W	2008-10
0297855549	OL15165640W	Jul 03, 2009
140846134X	OL15165640W	2010
3100954009	OL15165640W	2008
3596186447	OL15165640W	Apr 25, 2010
8408086944	OL15165640W	May 19, 2008
9655176797	OL15165640W	2010
0753826445	OL15165640W	2009
0767931114	OL15165640W	2010 May
0739328492	OL15165640W	2009
030745536X	OL15165640W	2008-05
0755330315	OL167150W	2007
0739486063	OL167150W	2007
0316004316	OL167150W	2007-11
0316015059	OL167150W	2007-11
1405649291	OL167150W	2008
1600248497	OL167150W	Aug 04, 2009
0755381238	OL167150W	2010
0446198986	OL167150W	2008-10
1602523150	OL167150W	November 13, 2007
1600240577	OL167150W	November 13, 2007
1600240593	OL167150W	November 13, 2007
1600240550	OL167150W	November 13, 2007
0786282924	OL167174W	2006
0755321944	OL167174W	2006
1408432226	OL167174W	2009
1455530689	OL167174W	Oct 28, 2014
0606366261	OL167174W	Oct 28, 2014
1435233492	OL167174W	Apr 11, 2008
1478956259	OL167174W	Oct 28, 2014
7544822664	OL167174W	2012
0756982723	OL167174W	Apr 01, 2007
0316185515	OL167174W	May 30, 2007
0316067954	OL167174W	2007-04
1417750294	OL167174W	2006-05
0755321928	OL167174W	2005
0446617792	OL167174W	2006-05
5558595256	OL167174W	January 2001
160024226X	OL167174W	July 22, 2008
0316059927	OL167174W	2006
031615556X	OL167174W	2005
0316117366	OL167152W	2007-07
0739484702	OL167152W	2007
0316118826	OL167152W	2007-07
0755335724	OL167152W	2008
0446501646	OL167152W	2008-04
1472253434	OL167152W	\N
3442465982	OL167152W	2008
9186369792	OL167152W	2010
0446407054	OL167152W	2008-06
1405649348	OL167152W	2008
2253133809	OL167152W	2010
8373598138	OL167152W	2009
044619896X	OL167152W	2009
0755335708	OL167152W	2007
1600242308	OL167152W	April 1, 2008
159483928X	OL167152W	July 2, 2007
1594839263	OL167152W	July 2, 2007
0446581747	OL167152W	July 3, 2007
902345734X	OL9302808W	2010
8373598448	OL9302808W	2010
902342929X	OL9302808W	2008
0061624772	OL9302808W	2009
0007287097	OL9302808W	2008
1408460289	OL9302808W	2010
000729266X	OL9302808W	2009
097915930X	OL9302808W	2007-07
8408005146	OL9302808W	May 08, 2012
6051112219	OL9302808W	2009
0007345828	OL9302808W	2010
0061982997	OL9302808W	2008-06
0061624764	OL9302808W	2008
0061696099	OL9302808W	2008-06
1607516640	OL9302808W	2006
0061668265	OL9302808W	2008
0061661554	OL9302808W	July 29, 2008
605539555X	OL14920152W	Oct 29, 2012
8580410533	OL14920152W	2012-01-01
2253167231	OL14920152W	2011
1607885441	OL14920152W	Jun 28, 2010
0446571504	OL14920152W	Jan 01, 2010
0099594633	OL14920152W	Sep 22, 2011
0446574724	OL14920152W	2011
0316096237	OL14920152W	2010
9023458907	OL14920152W	2010
160788545X	OL14920152W	Jun 28, 2010
1607886901	OL14920152W	Feb 22, 2011
1445007819	OL14920152W	Mar 01, 2011
1445007800	OL14920152W	Mar 01, 2011
1445007797	OL14920152W	Mar 01, 2011
1616644656	OL14920152W	Sep 22, 2010
1607884488	OL14920152W	Jun 28, 2010
0099550067	OL14920152W	2011
6045913760	OL5842017W	Nov 12, 2014
986620071X	OL5842017W	2012
8580578213	OL5842017W	2015
8324025901	OL5842017W	2014
6054482793	OL5842017W	Dec 25, 2011
9022552756	OL5842017W	2010
1101902884	OL5842017W	May 22, 2018
0804164223	OL5842017W	Jul 02, 2013
2253157139	OL5842017W	2011
0553418483	OL5842017W	Jun 02, 2015
0753827034	OL5842017W	Sep 12, 2010
0606359737	OL5842017W	May 04, 2010
3596173981	OL5842017W	Sep 24, 2015
1101972483	OL5842017W	2016
1780226861	OL5842017W	2016-01-14
1615237801	OL5842017W	2009
075382759X	OL5842017W	2010
1410417751	OL5842017W	2009
2702137814	OL5842018W	2007
6054377868	OL5842018W	Dec 25, 2011
9651321024	OL5842018W	2009
8324048308	OL5842018W	Jan 01, 2018
5389052684	OL5842018W	2013
2298011389	OL5842018W	2008
0804164231	OL5842018W	Jul 02, 2013
1101902876	OL5842018W	May 22, 2018
0525576819	OL5842018W	Jul 01, 2018
052557574X	OL5842018W	2018
0606367225	OL5842018W	Jul 31, 2007
1405619864	OL5842018W	2008
1405619872	OL5842018W	2008
080413832X	OL5842018W	Mar 01, 2014
0804171769	OL5842018W	Jun 03, 2014
0525575758	OL5842018W	Jun 12, 2018
0297851535	OL5842018W	January 3, 2007
0307341550	OL5842018W	July 31, 2007
141593309X	OL5842018W	December 2006
1597224588	OL5842018W	March 7, 2007
0307341542	OL5842018W	2006
0425067688	OL3297218W	May 1, 1984
0375726772	OL3297218W	December 2, 2003
0345259114	OL3297218W	July 12, 1977
0345282558	OL3297218W	December 12, 1978
0881847178	OL3297218W	May 1991
0736616063	OL3297218W	July 1, 1989
2020133989	OL3297218W	August 30, 1991
3257205392	OL3297218W	April 1, 1999
0006158137	OL3297218W	1979
0006146449	OL3297218W	1977
0006121187	OL3297218W	1966
0434019720	OL3297218W	October 1965
0006169074	OL3297218W	May 24, 1984
0812999037	OL16806568W	2014-01-01
0425286037	OL16806568W	2016-01-01
9024561906	OL16806568W	2013
2702158560	OL16806568W	2016
8850256779	OL16806568W	2020
3734104572	OL16806568W	2017
1524722871	OL16806568W	Sep 06, 2016
039959325X	OL16806568W	2016
0399594973	OL16806568W	2016
0399593268	OL16806568W	Sep 06, 2016
0593065743	OL16806568W	2013
0307749665	OL16806568W	Sep 03, 2013
1524734403	OL16806568W	Jul 12, 2016
0440246326	OL16806568W	2014
0804121044	OL16806568W	2013
0385344341	OL16806568W	2013
8422668068	OL86274W	1997
8401468558	OL86274W	1999
2844040187	OL86274W	1998
4167254484	OL86274W	1999
8532510922	OL86274W	2000
3442054982	OL86274W	February 1, 2003
0747251215	OL86274W	1997
8882740773	OL86274W	2000
9100566853	OL86274W	1998
8761204536	OL86274W	1999
9143001483	OL86274W	1998
1568954913	OL86274W	1997
222108084X	OL86274W	1998
0747215286	OL86274W	1997
9635487967	OL86274W	1999
8401327059	OL86274W	1997
8760804920	OL86274W	1998
837132748X	OL86274W	2004
8259021056	OL86274W	1999
9100574562	OL86274W	2000
8371324081	OL86274W	2000
0770429661	OL86274W	2004
077042788X	OL86274W	1998
8820026473	OL86274W	1998
0385485212	OL86274W	1997
0786240644	OL448959W	2002
0754018415	OL448959W	2002
0425191184	OL448959W	2003
5699149376	OL448959W	2006
0399148701	OL448959W	2002
0399149147	OL448959W	2002
1410400441	OL448959W	2003
9576775361	OL448959W	2003
0141014156	OL448959W	2003
0553713094	OL448959W	August 6, 2002
0553713108	OL448959W	August 6, 2002
2226141766	OL448959W	October 15, 2003
2226141812	OL448959W	October 15, 2003
8408046446	OL448959W	March 2003
8408054031	OL448959W	November 2004
8408065068	OL448959W	September 30, 2006
8408059815	OL448959W	2005
3550090757	OL448959W	March 1, 2003
2744176303	OL448959W	2004
3453864816	OL448959W	2002
0718146077	OL448959W	Jul 04, 2002
1101912618	OL448959W	Mar 10, 2015
0736688870	OL448959W	Jan 23, 2002
8466355464	OL20126932W	Apr 08, 2021
8466351833	OL20126932W	Oct 08, 2020
9506445095	OL20126932W	Nov 05, 2019
9752126049	OL20126932W	Nov 03, 2021
3453272374	OL20126932W	Sep 09, 2019
1982110562	OL20126932W	September 10, 2019
0593311213	OL20126932W	2020
1982110597	OL20126932W	2019 09
855651085X	OL20126932W	\N
8401022355	OL20126932W	Sep 12, 2019
1432870122	OL20126932W	2019
1529355397	OL20126932W	2019
1529355419	OL20126932W	2020
2298159599	OL20126932W	Sep 24, 2020
1982110570	OL20126932W	2021
1432870130	OL20126932W	Nov 21, 2020
225310342X	OL20126932W	Aug 25, 2021
2226443274	OL20126932W	Jan 29, 2020
031611880X	OL5337429W	2007
0755330374	OL5337429W	2008
0316014796	OL5337429W	2007
9046114104	OL5337429W	2011
9022992454	OL5337429W	2008
1847821995	OL5337429W	2008
0755330358	OL5337429W	2007
0446198951	OL5337429W	2008
0446179515	OL5337429W	2008-01
159483895X	OL5337429W	May 8, 2007
1594838925	OL5337429W	May 8, 2007
1594838941	OL5337429W	May 8, 2007
014000257X	OL106064W	1971
0434305502	OL106064W	1970
0140185410	OL106064W	1977
0859974731	OL106064W	1980
8402083943	OL106064W	1981
3764501626	OL167333W	2005
0743526368	OL167333W	Oct 01, 2002
3442366089	OL167333W	Feb 05, 2007
0446527041	OL167333W	2002
9044310895	OL167333W	2004
0340827688	OL167333W	2003
5699370420	OL167333W	2009
0446613053	OL167333W	2008-11
141040160X	OL167333W	2003
0739429345	OL167333W	2002
0786243473	OL167333W	2002
0786243481	OL167333W	2002
074355521X	OL167333W	August 7, 2006
5699027092	OL167333W	2003
3257203411	OL59431W	1984
8373920501	OL59431W	2003
325706408X	OL59431W	March 1, 2002
3257234082	OL59431W	April 1, 2003
2724236254	OL59431W	1987
0802145159	OL59431W	2011
225305786X	OL59431W	October 2, 1991
0099282976	OL59431W	September 2, 1999
043433507X	OL59431W	January 1968
0871132907	OL59431W	1989
2253026360	OL59431W	1980
2266021214	OL59431W	\N
2702116515	OL59431W	April 1, 1994
0140036040	OL59431W	November 1992
1444798820	OL17116913W	Jan 29, 2015
8804647159	OL17116913W	2014
1473610419	OL17116913W	2014
1444765655	OL17116913W	2015
0345543254	OL17116913W	Aug 18, 2015
1101921854	OL17116913W	Aug 18, 2015
1101910763	OL17116913W	2015
0316017701	OL167177W	2008
1847823521	OL167177W	2008
0099514540	OL167177W	2009
1846052505	OL167177W	2008
0316004324	OL167177W	February 5, 2008
1478976519	OL19356257W	May 18, 2017
1478972955	OL19356257W	Apr 18, 2017
1509848274	OL19356257W	Aug 31, 2017
060641293X	OL19356257W	Mar 20, 2018
1538711222	OL19356257W	2017
147894546X	OL19356257W	Apr 18, 2017
1478969652	OL19356257W	Apr 18, 2017
1455586544	OL19356257W	Sep 12, 2017
1455586587	OL19356257W	2018
1447277430	OL19356257W	2017
1447277821	OL19356257W	Apr 20, 2017
5040909950	OL19356257W	2018
1455586560	OL19356257W	2017
1478988150	OL19356257W	Feb 27, 2018
1478930020	OL19356257W	Apr 18, 2017
1478972963	OL19356257W	Apr 18, 2017
1447277449	OL19356257W	Oct 31, 2017
8830448079	OL17079190W	2017
1455558524	OL17079190W	2014
1455533718	OL17079190W	2015
0099574047	OL17079190W	2015
0316410632	OL17079190W	2014
0316365386	OL17079190W	2014
1455515876	OL17079190W	2014
1478903546	OL17079190W	Apr 07, 2015
1478955473	OL17079190W	Sep 29, 2014
1455581984	OL17725006W	Apr 29, 2014
9400504446	OL17725006W	2014
1478952482	OL17725006W	Apr 22, 2014
144722535X	OL17725006W	2014
1478927682	OL17725006W	Apr 22, 2014
1478982527	OL17725006W	Apr 22, 2014
1478984236	OL17725006W	Feb 24, 2015
1455521205	OL17725006W	2014-04
840816337X	OL16708051W	Oct 11, 2016
6070737431	OL16708051W	2016-10
0062209035	OL16708051W	Jul 10, 2012
1444815954	OL16708051W	2013
7020095690	OL16708051W	2013-01-01
2221131029	OL16708051W	2012-11
0297868101	OL16708051W	Mar 20, 2012
1780223250	OL16708051W	\N
0062270737	OL16708051W	Mar 12, 2013
006222347X	OL16708051W	Feb 26, 2013
0345803302	OL16708051W	2012-05
0062206281	OL16708051W	2012
6070709845	OL16708051W	2013
006220629X	OL16708051W	2013
8804620307	OL16708051W	2012
840803121X	OL16708051W	Oct 01, 2012
1538762242	OL19356256W	2018-03
1549113186	OL19356256W	Dec 14, 2017
1509865772	OL19356256W	2018
1447277848	OL19356256W	2017
1538760282	OL19356256W	2017
1455586609	OL19356256W	2017
1478998431	OL19356256W	Nov 21, 2017
1549194461	OL19356256W	Jul 31, 2018
1478921528	OL19356256W	Nov 21, 2017
1478998423	OL19356256W	Nov 21, 2017
1447277414	OL19356256W	2018
0006126693	OL3917426W	Apr 07, 1975
0002218798	OL3917426W	1979
0006122329	OL3917426W	1970
0816196877	OL3917426W	July 1987
0854564438	OL3917426W	December 1990
184232019X	OL3917426W	September 2000
0006153976	OL3917426W	1986
9681315960	OL3917426W	1986
0060807326	OL3917426W	1984
0745157653	OL3917426W	January 1996
1455588806	OL17868937W	2015-09
1629535656	OL17868937W	2015
1455567736	OL17868937W	2016
1787460002	OL17868937W	2015
1780893027	OL17868937W	2001
0099594897	OL17868937W	Jul 14, 2016
0099594889	OL17868937W	Mar 26, 2001
1478964111	OL17868937W	Dec 27, 2016
0606389814	OL17868937W	Apr 05, 2016
0316410985	OL17868937W	2015-09
0006138950	OL1822016W	1975
0006169082	OL1822016W	1984
0330342231	OL1822016W	May 22, 1998
0002214504	OL1822016W	June 1966
0881841447	OL1822016W	November 1985
0736608540	OL1822016W	December 1, 1984
555367333X	OL1822016W	\N
0006146473	OL1822016W	\N
0854561188	OL1822016W	1972
0345274121	OL1822016W	September 12, 1978
1473654858	OL15191187W	2016-09-01
1611292859	OL15191187W	2010-01-01
1449839401	OL15191187W	Jul 13, 2010
8491870482	OL15191187W	Apr 12, 2018
1410429741	OL15191187W	2010
1444722018	OL15191187W	2011
0670021873	OL15191187W	2010
0143119494	OL15191187W	2011
2744103101	OL874049W	1996
0679451617	OL874049W	Apr 23, 1996
0333672240	OL874049W	May 24, 1996
2266094106	OL874049W	February 15, 2001
034539044X	OL874049W	June 10, 1997
034542252X	OL874049W	2000
0679426612	OL874049W	1996
067975878X	OL874049W	1996
0333664973	OL874049W	1996
0333632923	OL874049W	1996
9500817853	OL874049W	1997
0330347683	OL874049W	March 7, 1997
5702010698	OL874049W	1999
0913369683	OL874049W	April 17, 1997
344243968X	OL874049W	July 1, 1998
8408024450	OL874049W	May 1998
0434305510	OL106059W	1970
0370014251	OL106059W	1970
0140031464	OL106059W	1981
0434305049	OL106059W	1951
0099286173	OL106059W	July 5, 2001
0140185518	OL106059W	March 1, 1992
1787464490	OL19356870W	2018-11-06
0316484938	OL19356870W	Nov 08, 2018
1780895216	OL19356870W	2018
1538713810	OL19356870W	2019-03
1549168045	OL19356870W	Apr 30, 2018
0316412252	OL19356870W	2018
1549171585	OL19356870W	Apr 30, 2018
1549171593	OL19356870W	Apr 30, 2018
1538760886	OL19356870W	Nov 06, 2018
0316274046	OL19356870W	2018-04
3764506415	OL18147682W	Mar 19, 2018
6059441645	OL18147682W	Oct 31, 2018
0008234183	OL18147682W	Dec 27, 2018
0062896210	OL18147682W	Mar 05, 2019
0062678426	OL18147682W	Mar 05, 2019
0062791451	OL18147682W	Jan 02, 2018
0062678418	OL18147682W	Jan 02, 2018
2709656817	OL16809825W	2017-06
1444822780	OL16809825W	2015
1780890133	OL16809825W	2013
1780890141	OL16809825W	2013
1455585866	OL16809825W	2014
0316210919	OL16809825W	2013
1478952989	OL16809825W	Sep 30, 2014
0606357440	OL16809825W	Sep 30, 2014
9044515241	OL15679291W	2010
9044517546	OL15679291W	2010
9044518542	OL15679291W	2010
2298038635	OL15679291W	Mar 10, 2010
0307398846	OL15679291W	2012
2757825097	OL15679291W	2012
2021018784	OL15679291W	2010
8831707280	OL15679291W	2010
1846553717	OL15679291W	2011
1846553725	OL15679291W	2011
1445858185	OL15679291W	2011
1445858193	OL15679291W	2011
3552054960	OL15679291W	Feb 27, 2010
0739378112	OL15679291W	2011
0307877965	OL15679291W	Mar 29, 2011
0062314270	OL278110W	Aug 20, 2013
0061241431	OL278110W	Feb 05, 2007
0575081864	OL278110W	2008
0061944890	OL278110W	December 22, 2009
0061233242	OL278110W	2007
0061235873	OL278110W	February 13, 2007
1598208853	OL278110W	May 1, 2007
8496463869	OL278110W	2007
0752889184	OL278110W	2007
0061147931	OL278110W	February 13, 2007
006114794X	OL278110W	April 1, 2008
1415937559	OL278110W	July 2007
1945540184	OL17929924W	2017-03-07
0316505943	OL17929924W	\N
3596298938	OL17929924W	Sep 01, 2019
3651025500	OL17929924W	Nov 08, 2016
0316387835	OL17929924W	Nov 08, 2016
1478918047	OL17929924W	Nov 08, 2016
1478906561	OL17929924W	Nov 08, 2016
0316505579	OL17929924W	Nov 08, 2016
517103155X	OL17929924W	2018
8937842610	OL17929924W	2017
0751570044	OL17929924W	Mar 26, 2017
0316468134	OL17929924W	Mar 05, 2016
0553815601	OL5718957W	2007-01-01
1665187239	OL5718957W	Mar 01, 2021
1409150674	OL5718957W	2014
1409135276	OL5718957W	2012
140913704X	OL5718957W	2012
1846176174	OL5718957W	2007
0553818937	OL5718957W	2007
0316735949	OL5718957W	2006
1597223425	OL5718957W	October 18, 2006
0316003530	OL5718957W	February 12, 2008
1598870289	OL5718957W	July 6, 2006
0593051424	OL5718957W	August 1, 2006
1473616964	OL17930362W	Mar 27, 2017
0399565035	OL17930362W	Oct 24, 2017
038554412X	OL17930362W	2017
1101967706	OL17930362W	Jun 19, 2018
0385541171	OL17930362W	Oct 24, 2017
1984833162	OL17930362W	Oct 16, 2018
1984833170	OL17930362W	Oct 16, 2018
0399565191	OL17930362W	Oct 24, 2017
1455548839	OL17062554W	2013
145558407X	OL17062554W	2014
9400501161	OL17062554W	2013
144726505X	OL17062554W	Aug 31, 2013
1447225341	OL17062554W	2014
1455521248	OL17062554W	2014
1624908772	OL17062554W	2013
1447225309	OL17062554W	2013
1478951699	OL17062554W	Jul 29, 2014
1455521310	OL17062554W	2013
1784759074	OL17426934W	2014
0606365265	OL17426934W	2014
1471267393	OL17426934W	2014
147898225X	OL17426934W	Mar 24, 2014
0099574233	OL17426934W	2014
0316211230	OL17426934W	2014
0613519310	OL27641W	March 2002
0689852282	OL27641W	October 1, 2002
8551001027	OL17606520W	2017
3499271672	OL17606520W	Apr 22, 2017
3805250975	OL17606520W	May 21, 2016
1101990260	OL17606520W	2016
0399583025	OL17606520W	2016
1101990473	OL17606520W	2016
0525590161	OL17606520W	Mar 06, 2018
0593076214	OL17606520W	2016
0593076222	OL17606520W	2016
1628999233	OL17606520W	2016
0147525179	OL17606520W	Feb 16, 2016
3518395610	OL15284W	1999
3518188607	OL15284W	2005
3518409565	OL15284W	1998
000617776X	OL261852W	1989
2757821946	OL261852W	Feb 01, 2011
0394563700	OL261852W	1988
0061287350	OL261852W	2007-06-03
3596185815	OL261852W	May 01, 2010
0812987381	OL261852W	2014
0140122060	OL261852W	1989
0140156852	OL261852W	April 1, 1991
0002234246	OL261852W	1988
0006543871	OL261852W	1991?
8433911597	OL261852W	August 1993
0006545475	OL261852W	1993
2809816166	OL17306744W	2015
0099574020	OL17306744W	2014
1780890095	OL17306744W	2013
1455553352	OL17306744W	2014-04
145554566X	OL17306744W	Oct 31, 2013
1478953632	OL17306744W	Dec 16, 2014
0316210986	OL17306744W	2013-09
1471276325	OL17356815W	2014
1455856053	OL17356815W	Feb 10, 2015
1409151204	OL17356815W	2014-10-23
1409144607	OL17356815W	2014
1410466280	OL17356815W	2014
0525953493	OL17356815W	2014
0099574152	OL17283721W	2014
1455576751	OL17283721W	Jul 31, 2014
1478953829	OL17283721W	Jul 08, 2014
0316211125	OL17283721W	2013
0316211095	OL17283721W	2013
1447297571	OL17358795W	\N
3442205123	OL17358795W	Mar 27, 2017
1524763241	OL17358795W	Mar 27, 2018
1410491455	OL17358795W	2016
0525590285	OL17358795W	Mar 06, 2018
1101924470	OL17358795W	Jul 26, 2016
144729758X	OL17358795W	\N
1447297563	OL17358795W	\N
143284007X	OL17358795W	May 03, 2017
1101904232	OL17358795W	Jul 26, 2016
1101904224	OL17358795W	2016
1101904240	OL17358795W	2017
1549178075	OL19865381W	Feb 18, 2019
1549178083	OL19865381W	Feb 18, 2019
1538714868	OL19865381W	2019-11
0316532320	OL19865381W	Feb 18, 2019
1423310020	OL491818W	June 27, 2006
1423310004	OL491818W	June 27, 2006
034548651X	OL491818W	May 1, 2007
1423310047	OL491818W	June 27, 2006
0749937718	OL491818W	2007
0739326317	OL491818W	August 22, 2006
142331008X	OL491818W	April 28, 2007
0345486501	OL491818W	June 27, 2006
1423310039	OL491818W	June 27, 2006
0749936924	OL491818W	2006
1423309995	OL491818W	June 27, 2006
1423310012	OL491818W	June 27, 2006
1423310071	OL491818W	June 27, 2006
986165206X	OL491818W	2006
1629538604	OL17359906W	2016-01-01
0099594390	OL17359906W	2017
0316407089	OL17359906W	2016-08
1455585297	OL17359906W	2017-08
0316407194	OL17359906W	Aug 01, 2016
147891632X	OL17359906W	Aug 29, 2017
1780892721	OL17359906W	2016
8324164901	OL20036181W	Jun 03, 2017
0735214840	OL20036181W	2016-01-01
8850256442	OL20036181W	\N
0735218358	OL20036181W	2017-01-01
060640791X	OL20036181W	Nov 14, 2017
1524702943	OL20036181W	Nov 15, 2016
1984883518	OL20036181W	Mar 05, 2019
0399575510	OL20036181W	2016
\.


--
-- TOC entry 3676 (class 0 OID 24748)
-- Dependencies: 226
-- Data for Name: physical; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.physical (isbn10, type) FROM stdin;
3404154851	Taschenbuch
8466423117	paperback
8497870794	paperback
8497870379	hardcover
849787157X	paperback
9639526770	unknown binding
9875808857	paperback
8804628553	hardcover
8408175726	mass market paperback
9584255185	paperback
9601409025	paperback
6171247588	hardcover
2709628538	paperback
8497081749	paperback
8417031235	mass market paperback
849930026X	mass market paperback
0593055810	paperback
8126412267	paperback
2709643952	paperback
0593056574	paperback
5550155184	hardcover
0552173878	paperback
8391913139	paperback
8373591672	paperback
0307277674	paperback
8957591052	Hardcover
3763255389	Hardcover
8373594213	paperback
8408176021	Paperback
3404148665	Taschenbuch
8476696957	paperback
8417031278	mass market paperback
8492549742	mass market paperback
8466423133	paperback
8497870476	paperback
8804667249	paperback
2266198122	pocket book
9722514091	paperback
8599296027	paperback
8389779285	hardcover
8375088285	paperback
975210455X	paperback
9752110800	paperback
9752106404	paperback
8492516518	paperback
9875808873	paperback
8467210117	hardcover
8408227637	flexibound
8408176005	paperback
8408099973	paperback
2253093009	pocket book
0739443917	Hardcover
1417664207	library binding
9639526266	Paperback
1416580824	Paperback
9100110566	hardcover
1442365471	mp3 cd
1557048347	hardcover
0593057708	paperback
8184980450	paperback
0743597184	mp3 cd
8408175742	Paperback
9584222546	paperback
9752112676	paperback
9024589401	mass market paperback
9021012812	paperback
8466423125	paperback
8497876806	hardcover
8497874587	hardcover
8497874579	hardcover
8497874528	paperback
8476698909	paperback
9722526774	paperback
8499307515	hardcover
8417031243	mass market paperback
8499302394	mass market paperback
8408099221	mass market paperback
8408176013	paperback
3404160002	paperback
3828995969	perfect paperback
1408460300	hardcover
0525565868	paperback
5170251955	Hardcover
9722520148	paperback
2253134171	Mass Market Paperback
859929668X	Paperback
5271303810	Hardcover
1615232168	hardcover
8983923385	paperback
0132268310	hardcover
7020101577	paperback
9601420452	paperback
8408175750	paperback
8401499976	mass market paperback
2277231126	pocket book
9875662690	paperback
8868361558	Paperback
8376481118	paperback
8376486136	paperback
9024581818	paperback
8466345256	mass market paperback
846634568X	paperback
0606038590	library binding
0340390700	hardcover
1529311144	paperback
2724247493	Hardcover
0340923288	Mass Market Paperback
0340951435	paperback
0450417395	paperback
0451153553	Mass Market Paperback
5170255470	Hardcover
0688046592	Hardcover
8532527698	hardcover
033053517X	paperback
0451168631	paperback
151424716X	paperback
1509886060	paperback
0230736076	paperback
0330465732	paperback
0330465724	paperback
0451166892	Mass Market Paperback
0330699814	Paperback
8804413476	paperback
8496863247	paperback
8401023866	hardcover
0330450131	paperback
1509848495	paperback
1447265440	paperback
0330534920	paperback
0330391984	Paperback
8422635755	Hardcover
0451207149	Paperback
9752121764	hardcover
0451159276	mass market paperback
0451169514	mass market paperback
8466357300	mass market paperback
6073157061	paperback
9669931754	hardcover
1623300894	board book
9877253194	paperback
8868365626	Paperback
2253083364	pocket book
9752119271	paperback
8466345345	paperback
3453504089	paperback
8466347925	mass market paperback
150824412X	mp3 cd
9024516137	Paperback
9877250241	paperback
1982127791	Paperback
1444707868	Paperback
9500406977	Paperback
1501182099	Hardcover
345343577X	paperback
345309994X	Paperback
3548256112	Paperback
9754051518	paperback
840102143X	hardcover
229034589X	pocket book
8466341293	mass market paperback
9875666270	paperback
8466336958	mass market paperback
6073156243	paperback
9722526456	paperback
8495501627	paperback
0451160525	mass market paperback
8878244570	paperback
8440607466	Hardcover
0747401004	Paperback
0670032549	paperback
0340896213	Paperback
1471227898	Paperback
9024526884	Mass Market Paperback
902455585X	paperback
9024516641	paperback
8868363666	paperback
2738206581	Hardcover
0606391622	library binding
3453875567	paperback
3453873017	Paperback
1501166115	mass market paperback
836578131X	hardcover
3404150554	Taschenbuch
0671027387	paperback
0552171360	paperback
8408175769	mass market paperback
9584261878	paperback
987580889X	paperback
3404770579	paperback
3404175042	paperback
2709626411	paperback
9752111300	paperback
9752105734	paperback
9637318739	Hardcover
0593057430	paperback
0552151769	Paperback
837659026X	Paperback
837508199X	Paperback
8379997360	paperback
846721385X	Hardcover
0552173541	Mass Market Paperback
9100111899	Mass Market Paperback
055216996X	paperback
8791746485	Paperback
0552169978	paperback
1846174589	hardcover
1417751460	library binding
0606330178	paperback
1442365463	mp3 cd
1416528644	mass market paperback
3453007867	Softcover
2290307769	pocket book
2738204422	hardcover
857302187X	Paperback
6073113870	paperback
8466357335	mass market paperback
9877253747	paperback
0451150244	mass market paperback
8401499844	Paperback
1529378303	paperback
1444708139	paperback
1501156705	mass market paperback
5170770863	Paperback
0385182449	Hardcover
9754051526	paperback
3453435796	paperback
3453504070	paperback
2724278488	Hardcover
2724229916	Hardcover
3548263100	Mass Market Paperback
345321045X	Paperback
0451139755	mass market paperback
1529378311	Paperback
198211598X	paperback
5170129467	Hardcover
1982112395	hardcover
0451162072	Mass Market Paperback
3453007042	Taschenbuch
8466357130	paperback
5170720920	hardcover
8374696192	paperback
9754059640	paperback
1501143816	mass market paperback
0606019170	paperback
1786364670	hardcover
1439507627	library binding
0517321998	hardcover
0812424719	library binding
9024544971	Mass Market Paperback
2724278437	Hardcover
8882742970	Mass Market Paperback
8804297263	paperback
9024559499	paperback
9024526868	paperback
5770708417	Hardcover
0451093380	Mass Market Paperback
5170067399	Paperback
1501144502	paperback
8497871588	paperback
8497871596	hardcover
8408176102	paperback
9875808903	paperback
841703126X	mass market paperback
8499300278	mass market paperback
0552171379	paperback
9752111653	paperback
9752112439	paperback
7020101542	Paperback
8376590278	Paperback
8375081981	Paperback
8375088498	paperback
5170137761	Mass Market Paperback
8373591729	paperback
9953299129	paperback
9024527910	Paperback
0752868926	hardcover
1439565775	library binding
9722514695	Paperback
2253127078	Mass Market Paperback
2744192066	Hardcover
0552151696	Mass Market Paperback
0552169986	Mass Market Paperback
1250064074	paperback
055217355X	Paperback
9024553024	Paperback
8389779013	Paperback
9754058997	paperback
2709610205	Paperback
2277223263	paperback
8581050549	paperback
0451160959	mass market paperback
0451179285	mass market paperback
8466357238	mass market paperback
1623303281	hardcover
1529370515	paperback
340425242X	Taschenbuch
3404281268	perfect paperback
060601828X	library binding
0593314018	paperback
1529356547	paperback
3785705859	hardcover
270961281X	paperback
9752121772	hardcover
9752114075	paperback
0593313887	paperback
5170276168	Hardcover
0451924096	paperback
0340920955	paperback
1444720732	paperback
0307947300	paperback
143955787X	library binding
3453438183	paperback
5170109407	Paperback
8497599411	Paperback
3453061187	Taschenbuch
0385470819	Hardcover
9022981037	Hardcover
0440211727	Mass Market Paperback
0922066728	Hardcover
080412115X	Paperback
0099590751	Mass Market Paperback
2724289242	Hardcover
8422655799	Hardcover
2266072471	Mass Market Paperback
8371691319	Paperback
2266068520	Mass Market Paperback
2266115499	Mass Market Paperback
0440245915	Mass Market Paperback
8408012428	Paperback
8408020935	Paperback
8408041258	Paperback
0385338600	Paperback
9992191422	Paperback
0816155909	Hardcover
0099201216	Paperback
8408033581	Paperback
0780735420	Unknown Binding
0034540288	paperback
140587967X	paperback and CD
0679419462	Hardcover
034540288X	mass market paperback
274410700X	Hardcover
902451956X	Paperback
8811602289	paperback
8576573059	paperback
1784752231	Paperback
8532506321	Paperback
0345538994	mass market paperback
3426193817	hardcover
837169394X	Paperback
8401326508	Hardcover
9500420104	Paperback
9500415844	Paperback
1400001129	Mass Market Paperback
2221081293	Paperback
2266079123	Mass Market Paperback
2266116061	Mass Market Paperback
8497597796	Paperback
067945540X	Hardcover
0099637812	Mass Market Paperback
8379856716	paperback
8376597434	hardcover
8376597426	paperback
8376483730	paperback
8374698187	paperback
6073120362	paperback
0613171004	school & library binding
2226105662	Paperback
0684853507	hardcover
5170268580	Hardcover
5170262841	Paperback
5170177305	Hardcover
2702838022	Hardcover
067102423X	Mass Market Paperback
2744135933	Hardcover
1501160451	mass market paperback
8573022469	Paperback
3453160819	Paperback
3453210441	Paperback
0340921358	Paperback
0340718196	Hardcover
1501198890	paperback
034071820X	paperback
8422676095	Hardcover
8882742172	Paperback
3548268412	paperback
8379858581	paperback
9024536448	paperback
0606183698	library binding
3548252427	Paperback
9751016150	paperback
3453441095	paperback
3795117496	hardcover
8086278743	Hardcover
3548840124	paperback
9751037190	paperback
5170030681	Hardcover
5237029493	Hardcover
5170250843	Hardcover
1501157515	mass market paperback
3548251285	Paperback
3548255744	Paperback
3795117569	Softcover
2226115234	Paperback
225315136X	Mass Market Paperback
3795117402	Hardcover
2290032433	pocket book
2290321125	Mass Market Paperback
8401021413	hardcover
8466340718	mass market paperback
9875666963	paperback
9752103596	paperback
345301216X	hardcover
3453875583	paperback
9722528866	paperback
8573026693	Paperback
0451173317	mass market paperback
0452267404	Paperback
8466300228	Paperback
0451210867	Mass Market Paperback
3453123867	Paperback
837985585X	hardcover
0606391649	library binding
1501143549	paperback
517027100X	Paperback
0452279623	Paperback
093798616X	hardcover
0340832258	Paperback
034082977X	Paperback
8440644302	Paperback
2277250341	Paperback
0747411875	Paperback
0785705430	Library binding
1417637137	School & Library Binding
8868363690	paperback
8401021456	hardcover
8466342656	mass market paperback
607313021X	paperback
8499892604	mass market paperback
0451194861	paperback
0452284724	paperback
0451210875	paperback
0452279178	paperback
2290053139	pocket book
1444723472	paperback
3453147596	Paperback
1880418371	paperback
8379855868	hardcover
0606391657	library binding
0340829788	Paperback
0340712937	hardcover
0613090993	Library binding
227725035X	Paperback
8440690134	Paperback
9176436160	Mass Market Paperback
5170205414	Hardcover
3598800061	Hardcover
3455025013	Hardcover
3453169247	Paperback
8466300627	Mass Market Paperback
8440694377	Paperback
8496546381	Paperback
8440681798	Hardcover
2266107275	Mass Market Paperback
2221087143	Paperback
2266145428	Paperback
0440245958	Mass Market Paperback
0375433473	Hardcover
8440669437	Paperback
0712679332	Unknown Binding
2744155535	Hardcover
2266145444	Mass Market Paperback
0440206154	Mass Market Paperback
3453861582	Paperback
8440630123	Hardcover
844063014X	Hardcover
9994122304	Paperback
0517494264	Hardcover
2863740946	Paperback
0708981690	Library Binding
0712653570	Hardcover
0899668771	Hardcover
0099111519	Paperback
2266103679	Mass Market Paperback
3453025423	Paperback
2226120432	Paperback
0307344681	Paperback
8439705158	Paperback
8497594924	Paperback
9724607046	Paperback
0553227467	Mass Market Paperback
2253013099	paperback
1611297222	hardcover
7020122639	Paperback
0449219631	mass market paperback
8428603987	Paperback
8428603995	Hardcover
3548033202	Taschenbuch
7805672040	Paperback
8401490049	Paperback
8428604762	Paperback
9507426779	Paperback
0345544145	Trade Paperback
1557046778	Hardcover
0553204653	Mass Market Paperback
1417653078	Library Binding
060601179X	Library Binding
0330243829	Paperback
1400064562	Hardcover
0385047711	Hardcover
0671754327	Mass Market Paperback
2859405321	Mass Market Paperback
0140017232	Paperback
0385047258	hardcover
0884111490	Hardcover
0708980740	Hardcover
089190154X	Hardcover
0330245899	Paperback
009986620X	Paperback
0881844098	Paperback
0892330376	Hardcover
0575029218	Hardcover
6058072735	paperback
0007232837	Paperback
9044912895	paperback
349922951X	Paperback
0752856138	Paperback
8490567077	paperback
8490060096	paperback
1433263270	mp3 cd
089190378X	hardcover
0007242972	paperback
3499425882	paperback
3688107241	paperback
0553065491	Paperback
3499244411	paperback
8491875468	paperback
8411322246	paperback
8466410708	paperback
1433249073	mp3 cd
8490566356	paperback
9997531779	hardcover
034075494X	hardcover
0671042149	mass market paperback
0756905753	library binding
8497592956	Paperback
0606194967	library binding
1417739738	library binding
1501160478	mass market paperback
0671024248	Mass Market Paperback
0340738901	hardcover
0684844907	unbound
3453185668	Paperback
3453435710	paperback
3453159926	Hardcover
1501195972	paperback
0743436210	Mass Market Paperback
8401327865	Hardcover
5170095619	Hardcover
8401013011	Paperback
034073891X	Mass Market Paperback
3453177487	Paperback
2253151408	Mass Market Paperback
2226122095	Paperback
8401328454	Paperback
8484503119	Paperback
9871138504	Paperback
0684853515	hardcover
0340818670	Paperback
1400094518	Paperback
0743417682	Mass Market Paperback
8374696303	paperback
6073113862	paperback
9024539145	paperback
0340770708	Paperback
5170187289	Hardcover
5170230044	Paperback
0340952660	paperback
1439568103	library binding
1501160435	mass market paperback
0743228472	Hardcover
0340792345	Paperback
355008353X	Hardcover
0708949606	Hardcover
1444753673	Mass Market Paperback
2226150765	Paperback
8573025506	Paperback
1501192191	paperback
8497930843	Paperback
9875660175	Paperback
1400003148	Hardcover
8401329647	Hardcover
9506440182	Paperback
0684017148	Hardcover
0743235959	Hardcover
902452606X	paperback
6073178093	paperback
034095227X	paperback
0525941908	Hardcover
0451191013	Mass Market Paperback
0525942246	Hardcover
3548255019	Paperback
3453129601	Paperback
2226088083	Paperback
2290306681	Paperback
8882741591	Paperback
8484502546	Paperback
8447334260	Hardcover
8401474701	Paperback
1501143751	mass market paperback
837186065X	paperback
3453435818	paperback
9751037514	paperback
0340671769	Hardcover
1501144278	paperback
0788191799	hardcover
9501510085	Paperback
8490708835	mass market paperback
221302524X	Mass Market Paperback
2253048593	Mass Market Paperback
0345430581	Mass Market Paperback
0747403465	Paperback
0745160638	Talking Book
1400096472	Paperback
1400025109	Mass Market Paperback
0140129545	Hardcover
0345469380	Paperback
0446776890	Mass Market Paperback
844063269X	Paperback
3426306972	paperback
3426605392	turtleback
3426614383	Paperback
3426624869	Paperback
8376597221	paperback
2290332488	paperback
2290332461	pocket book
0743251628	Paperback
3453530233	Taschenbuch
9752104541	paperback
8379855876	hardcover
1439565929	library binding
517023645X	Hardcover
0340836156	paperback
0340827165	Paperback
0743561694	Unknown Binding
0340827173	Paperback
1417645687	School & Library Binding
9685960836	Paperback
141651693X	Paperback
1880418568	hardcover
5551280098	Mass Market Paperback
0340827157	Hardcover
8496581985	paperback
3548256104	Paperback
3548250203	Paperback
3548254896	Paperback
8498721970	paperback
8466617027	Paperback
8466302026	Mass Market Paperback
2869303912	Mass Market Paperback
2869301537	Mass Market Paperback
0445405252	Mass Market Paperback
0446698873	Paperback
0099419335	Paperback
0099366517	Paperback
0739473603	Hardcover
0099508516	Paperback
0099492164	Hardcover
0786293101	Hardcover
8376483498	paperback
9024564344	paperback
8868360276	paperback
225316979X	pocket book
8499891098	mass market paperback
9875669091	paperback
1439148503	hardcover
9661410259	hardcover
1501156799	mass market paperback
1476735476	paperback
0593311582	paperback
9752113249	paperback
0340992565	Hardcover
1594134170	paperback
0340992573	paperback
1442345500	mp3 cd
1442365498	mp3 cd
3453435230	paperback
1439149038	Trade Paperback
0307741125	Paperback
1410423964	Hardcover
1439156972	Hardcover
0340992581	Mass Market Paperback
8846200039	Paperback
1501216791	mp3 cd
0345378490	Mass Market Paperback
2253030031	Mass Market Paperback
2266068385	pocket book
2264013974	mass market paperback
3499263424	Paperback
8532512356	Paperback
0099544318	Paperback
0099320819	Paperback
034541893X	Paperback
1417617330	School & Library Binding
8401492343	Paperback
0712676058	Hardcover
0613100557	Library Binding
0060541830	Mass Market Paperback
8484502902	Paperback
0708989128	Hardcover
0061782556	Mass Market Paperback
9500415275	Paperback
0099682710	Paperback
038056176X	Mass Market Paperback
9024523656	Paperback
030758836X	Hardcover
0307588386	Epub
0553418351	Mass Market Paperback
0753827662	Mass Market Paperback
0345805461	Paperback
359652072X	hardcover
3596188784	paperback
1524763675	mass market paperback
1780228228	Mass Market Paperback
0297859382	Hardcover
159413605X	Paperback
3502102228	perfect paperback
3596032199	paperback
0385347774	paperback
2298071179	paperback
2355841179	Paperback
1471306984	Hardcover
9022578712	Paperback
7508639197	Paperback
8377782979	hardcover
0307588378	Paperback
0446614319	Mass Market Paperback
0755300297	Paperback
0739437062	Hardcover
0446692573	paperback
0755309774	hardcover
1435291034	library binding
1843956292	hardcover
0755334221	paperback
0755349377	paperback
1478938064	mp3 cd
141766715X	Library Binding
5558595272	Hardcover
0316743844	Text (large print)
0755300211	Hardcover
9044313363	Paperback
9752106951	paperback
837359311X	Paperback
2749902878	Hardcover
1455578614	mass market paperback
1455581771	paperback
7802252792	Paperback
5170391447	Paperback
5558595264	Hardcover
0316858471	Hardcover
0316858498	Paperback
044617792X	Paperback
0446576638	Hardcover
0751531804	Paperback
0446616621	Mass Market Paperback
0446577146	Hardcover
0739449079	Hardcover
9022989461	paperback
902299662X	paperback
3785723547	hardcover
0739475622	hardcover
0446400327	mass market paperback
9022994643	Mass Market Paperback
8498720958	Hardcover
1447274296	Paperback
0330523511	paperback
1405090111	paperback
0330444085	Mass Market Paperback
044653109X	Hardcover
0446615633	Mass Market Paperback
0446580198	Large Print
1405089849	Hardcover
0755330390	Hardcover
0316013943	Hardcover
0446199273	Mass Market Paperback
0755330412	[sound recording] /
1478963328	mp3 cd
0316017752	Hardcover
0755330404	Paperback
1587672316	hardcover
9023425545	Paperback
0606321829	library binding
0385669518	hardcover
1400026253	mass market paperback
6050905940	paperback
3442469376	paperback
8804606371	Paperback
2266218573	Paperback
2221111133	Paperback
0345504976	Paperback
0385693702	Paperback
0752883305	Paperback
0525618759	Paperback
0525618740	mass market paperback
0752897845	hardcover
1409128512	Paperback
1409190986	Paperback
0345525221	paperback
140910334X	paperback
0752897853	paperback
0345528174	Mass Market Paperback
0345504968	Hardcover
1410432874	hardcover
8489367876	Paperback
0061350206	paperback
0732283639	paperback
0307391981	paperback
2266242245	mass market paperback
0060872985	Hardcover
2266182048	Mass Market Paperback
006222719X	Paperback
0007330626	Epub
8401336406	Hardcover
0061284319	MP3 CD
0007248997	Paperback
0060873167	Mass Market Paperback
0060873035	Paperback
9573261162	Paperback
8495618850	Paperback
0739449672	Hardcover
0755305779	Paperback
0755305760	Paperback
1478963352	mp3 cd
0755305752	Hardcover
0316009563	Hardcover
0446694215	Paperback
555859523X	Hardcover
0330200747	Paperback
202006376X	Mass Market Paperback
8437617499	Paperback
0099470470	Paperback
0316290238	Paperback
009974371X	Paperback
0330295683	Paperback
0965058484	Hardcover
076072539X	Unknown Binding
0606252819	Turtleback
1850890293	Hardcover
0330309900	Paperback
0316290963	hardcover
354860224X	Paperback
8401490316	Paperback
8476695845	Hardcover
0739433326	Hardcover
0316147877	Hardcover
0755300203	Paperback
0755349466	paperback
0316602086	Hardcover
0316602078	Hardcover
075530019X	Paperback
061392519X	School & Library Binding
0446177873	Paperback
1509836489	paperback
3404160800	Paperback
849872872X	paperback
9022993418	Paperback
0446615641	Mass Market Paperback
0446577391	hardcover
2749912660	Paperback
144727430X	Mass Market Paperback
1447226577	paperback
0230017797	paperback
033052352X	Mass Market Paperback
0230017754	hardcover
0330450980	paperback
8466637443	Hardcover
0446195103	Hardcover
1602522324	Unknown Binding
3548288626	paperback
345308201X	Paperback
3462022776	Hardcover
0340937688	paperback
1524796956	paperback
0399594000	paperback
0345418301	Paperback
0340600888	Paperback
0345385764	Mass Market Paperback
0517137631	Hardcover
0517396165	Hardcover
0340766522	Paperback
1858918960	Paperback
0340597658	Paperback
0340592818	Hardcover
0345480325	Mass Market Paperback
0679425136	Hardcover
9032510924	library binding
2277234648	Mass Market Paperback
0671670689	Mass Market Paperback
2724278143	Hardcover
1416511849	Mass Market Paperback
0671717324	Hardcover
2277021520	Mass Market Paperback
1982113197	Paperback
2290303909	Mass Market Paperback
0743440269	Paperback
067171550X	Paperback
0671670670	Mass Market Paperback
0671853511	Paperback
0833558935	Library Binding
8401497590	Paperback
1416511865	paperback
9032510932	Library Binding
2277021547	Paperback
8401324882	Paperback
8401493331	Paperback
8401497604	Paperback
1400000734	Mass Market Paperback
3442412226	Paperback
9754052964	Paperback
999484928X	Paperback
0671695126	Mass Market Paperback
0671715798	Mass Market Paperback
0816153868	Paperback
0833574256	School & Library Binding
0745174825	Hardcover
2277235806	Mass Market Paperback
0743440277	Paperback
0671717413	paperback
2290303917	Mass Market Paperback
8489367108	Paperback
0755325672	Hardcover
0755325699	Paperback
0755325680	Paperback
031610695X	Large type
8377586533	Paperback
229803009X	Paperback
229802619X	paperback
0385528701	Hardcover
8408163361	hardcover
2266194232	mass market paperback
0307455378	Paperback
8804583355	Hardcover
0297855549	hardcover
140846134X	Hardcover
3100954009	hardcover
3596186447	paperback
8408086944	paperback
0753826445	Paperback
0767931114	Paperback
0739328492	Paperback
030745536X	Hardcover
0755330315	Hardcover
0739486063	hardcover
0316004316	Hardcover
0316015059	Hardcover
1602523150	Unknown Binding
0786282924	Hardcover
1455530689	mass market paperback
0606366261	library binding
1435233492	library binding
7544822664	Zhuan zhu,
0756982723	library binding
0316185515	paperback
0316067954	Paperback
1417750294	Unknown Binding
0755321928	Hardcover
5558595256	Hardcover
0316059927	Paperback
0316117366	Hardcover
0739484702	hardcover
0316118826	Hardcover
0755335724	Paperback
0446501646	Paperback
0446407054	mass market paperback
0755335708	Hardcover
0446581747	Paperback
902345734X	paperback
8373598448	Paperback
902342929X	Paperback
0061624772	Paperback
0007287097	Paperback
1408460289	Hardcover
000729266X	Paperback
097915930X	Paperback
8408005146	paperback
6051112219	Paperback
0007345828	Paperback
0061624764	Hardcover
1607516640	Paperback
0061668265	Paperback
605539555X	paperback
0446571504	paperback
0099594633	mass market paperback
0316096237	Hardcover
1445007819	mp3 cd
1616644656	hardcover
6045913760	paperback
8580578213	Capa mole
6054482793	paperback
1101902884	mass market paperback
0553418483	paperback
0753827034	paperback
0606359737	library binding
3596173981	paperback
1780226861	paperback
075382759X	paperback
6054377868	paperback
8324048308	paperback
1101902876	mass market paperback
0525576819	paperback
052557574X	paperback
0606367225	library binding
080413832X	mass market paperback
0297851535	Paperback
0307341550	Paperback
1597224588	Hardcover
0425067688	Paperback
0345259114	Mass Market Paperback
0345282558	Mass Market Paperback
0881847178	Paperback
2020133989	Mass Market Paperback
3257205392	Paperback
0006146449	Mass Market Paperback
0006169074	Paperback
3734104572	paperback
039959325X	paperback
0399594973	paperback
0399593268	paperback
0593065743	hardcover
3442054982	Paperback
0770429661	Mass Market Paperback
0399148701	Hardcover
0399149147	Hardcover
0141014156	Paperback
2226141766	Unknown Binding
2226141812	Unknown Binding
8408046446	Hardcover
8408054031	Paperback
8408065068	Paperback
8408059815	Paperback
8466355464	mass market paperback
8466351833	mass market paperback
9506445095	paperback
9752126049	paperback
3453272374	hardcover
0593311213	paperback
1982110597	Epub
855651085X	paperback
1982110562	Hardcover
8401022355	hardcover
1529355397	Hardcover
1529355419	B format paperback
2298159599	paperback
1982110570	mass market paperback
1432870130	paperback
225310342X	pocket book
2226443274	paperback
031611880X	[large print] :
0755330374	paperback
0316014796	Hardcover
0446179515	Paperback
014000257X	Paperback
0140185410	Paperback
3442366089	paperback
0446613053	Mass Market Paperback
0739429345	Hardcover
0786243473	[large print] /
0786243481	[large print] /
3257203411	Taschenbuch
325706408X	Hardcover
3257234082	Paperback
225305786X	Mass Market Paperback
0099282976	Paperback
043433507X	Hardcover
0871132907	Paperback
2702116515	Mass Market Paperback
0140036040	Hardcover
1444798820	paperback
0345543254	paperback
0316017701	Hardcover
0316004324	Hardcover
1509848274	paperback
060641293X	library binding
1538711222	hardcover
147894546X	paperback
1455586544	paperback
1455586587	mass market paperback
1447277430	hardcover
1447277821	paperback
1478972963	mp3 cd
1455581984	hardcover
144722535X	paperback
1478982527	mp3 cd
840816337X	hardcover
6070737431	paperback
1444815954	Hardcover
7020095690	Paperback
2221131029	Paperback
1780223250	paperback
006222347X	mass market paperback
0345803302	Paperback
0062206281	Hardcover
6070709845	paperback
006220629X	Trade Paperback
8804620307	Hardcover
840803121X	paperback
1509865772	paperback
1447277848	paperback
1478998431	mp3 cd
0006126693	paperback
0002218798	Paperback
0006122329	Paperback
0854564438	Hardcover
184232019X	Paperback
1780893027	hardcover
0099594897	paperback
0606389814	library binding
0330342231	Paperback
0002214504	Hardcover
0881841447	Paperback
0006146473	Paperback
0854561188	Unknown Binding
0345274121	Mass Market Paperback
8491870482	paperback
1410429741	Hardcover
1444722018	Paperback
0670021873	Hardcover
0143119494	Paperback
2266094106	Mass Market Paperback
034539044X	Paperback
9500817853	Paperback
0330347683	Paperback
034542252X	Mass Market Paperback
344243968X	Paperback
8408024450	Hardcover
0099286173	Paperback
0140185518	Paperback
0316484938	hardcover
1780895216	hardcover
1538713810	mass market paperback
3764506415	perfect paperback
6059441645	paperback
0008234183	paperback
0062678426	paperback
0062791451	paperback
1780890141	paperback
0606357440	library binding
2298038635	paperback
0061944890	Paperback
0061233242	Paperback
1598208853	Paperback
8496463869	Paperback
0752889184	Hardcover
0061147931	Hardcover
006114794X	Paperback
3596298938	paperback
3651025500	hardcover
0316387835	hardcover
1478918047	mp3 cd
0316505579	hardcover
0553818937	Mass Market Paperback
1597223425	Hardcover
0316735949	Hardcover
0316003530	Paperback
038554412X	hardcover
1101967706	paperback
144726505X	paperback
1447225309	hardcover
1784759074	paperback
147898225X	mp3 cd
0613519310	School & Library Binding
0689852282	Paperback
8551001027	paperback
3499271672	paperback
3805250975	perfect paperback
000617776X	paperback
2757821946	paperback
0394563700	Hardcover
3596185815	paperback
0812987381	Paperback
0140122060	Paperback
0140156852	Paperback
0006543871	Paperback
8433911597	Paperback
0006545475	Paperback
1780890095	Paperback
1455553352	mass market paperback
145554566X	paperback
1409151204	paperback
1409144607	hardcover
1455576751	paperback
1447297571	paperback
3442205123	perfect paperback
1524763241	mass market paperback
144729758X	paperback
1447297563	hardcover
1101904240	Paperback
1538714868	paperback
0316532320	paperback
034548651X	Mass Market Paperback
1423310047	MP3 CD
0749937718	Paperback
0739326317	Hardcover
0345486501	Hardcover
1423310039	MP3 CD
0749936924	Paperback
0099594390	paperback
0316407194	hardcover
8324164901	paperback
8850256442	paperback
0735218358	mass market paperback
060640791X	library binding
\.


--
-- TOC entry 3673 (class 0 OID 24715)
-- Dependencies: 223
-- Data for Name: rating; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.rating (rid, wid, cnt, avg_rating) FROM stdin;
1	OL76837W	143	3.91
2	OL76833W	217	3.60
3	OL14873315W	79	3.59
4	OL81634W	109	4.07
5	OL1914022W	57	4.18
6	OL81613W	436	4.10
7	OL81628W	40	4.18
8	OL76835W	91	3.26
9	OL81631W	163	4.05
10	OL81630W	37	4.16
11	OL76836W	29	3.55
12	OL81618W	79	4.30
13	OL77001W	23	3.70
14	OL46876W	16	4.31
15	OL81609W	28	3.96
16	OL81612W	24	3.79
17	OL81625W	75	4.17
18	OL81615W	61	4.07
19	OL77014W	14	3.79
20	OL23480W	30	4.23
21	OL3454854W	18	4.39
22	OL36626W	1	4.00
23	OL3899224W	8	3.25
24	OL81624W	27	3.96
25	OL81586W	17	3.53
26	OL2794726W	0	\N
27	OL510879W	3	4.00
28	OL81594W	177	4.05
29	OL553754W	15	3.67
30	OL14917748W	50	3.84
31	OL46913W	37	3.38
32	OL16239762W	53	3.60
33	OL167179W	9	3.56
34	OL11913975W	0	\N
35	OL22914W	1	5.00
36	OL41256W	8	4.00
37	OL167155W	8	3.63
38	OL15168588W	36	3.78
39	OL46904W	21	3.29
40	OL167160W	4	4.25
41	OL15008W	4	3.75
42	OL167161W	2	4.00
43	OL41249W	4	4.25
44	OL181821W	7	3.57
45	OL134880W	7	4.00
46	OL134885W	5	4.80
47	OL167162W	2	3.50
48	OL15165640W	12	4.17
49	OL167150W	7	3.43
50	OL167174W	21	3.90
51	OL167152W	2	5.00
52	OL9302808W	2	3.00
53	OL14920152W	6	3.17
54	OL5842017W	34	4.15
55	OL5842018W	25	3.76
56	OL3297218W	0	\N
57	OL16806568W	10	3.30
58	OL86274W	1	3.00
59	OL448959W	29	3.69
60	OL20126932W	13	3.85
61	OL5337429W	5	4.00
62	OL106064W	0	\N
63	OL167333W	5	4.40
64	OL59431W	3	4.00
65	OL17116913W	3	2.33
66	OL167177W	7	4.29
67	OL19356257W	3	4.00
68	OL17079190W	3	2.33
69	OL17725006W	9	4.00
70	OL16708051W	13	3.92
71	OL19356256W	3	4.33
72	OL3917426W	0	\N
73	OL17868937W	4	4.25
74	OL1822016W	0	\N
75	OL15191187W	13	3.77
76	OL18081210W	0	\N
77	OL874049W	1	4.00
78	OL106059W	1	3.00
79	OL19356870W	2	3.50
80	OL18147682W	8	3.75
81	OL16809825W	6	3.67
82	OL15679291W	5	4.00
83	OL278110W	13	3.54
84	OL17929924W	6	3.50
85	OL5718957W	0	\N
86	OL17930362W	5	3.20
87	OL17062554W	5	4.40
88	OL17426934W	0	\N
89	OL27641W	0	\N
90	OL17606520W	2	3.50
91	OL15284W	0	\N
92	OL261852W	2	2.50
93	OL17306744W	2	4.50
94	OL17356815W	2	4.50
95	OL17283721W	1	4.00
96	OL17358795W	69	4.06
97	OL19865381W	1	3.00
98	OL491818W	9	3.78
99	OL17359906W	2	4.00
100	OL20036181W	4	4.25
\.


--
-- TOC entry 3668 (class 0 OID 24679)
-- Dependencies: 218
-- Data for Name: work; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.work (wid, title, first_publish_date) FROM stdin;
OL76837W	The Da Vinci Code	2004
OL76833W	Angels & Demons	\N
OL14873315W	The Lost Symbol	\N
OL81634W	Misery	April 9, 2002
OL1914022W	The Pillars of the Earth	November 14, 2007
OL81613W	It	October 1, 1987
OL81628W	The Gunslinger	December 1976
OL76835W	Deception Point	\N
OL81631W	Pet Sematary	February 1996
OL81630W	The Dead Zone	August 1, 1983
OL76836W	Digital Fortress	\N
OL81618W	The Stand	January 1, 1980
OL77001W	A Time to Kill	\N
OL46876W	The Lost World	\N
OL81609W	Bag of Bones	January 1999
OL81612W	The Girl Who Loved Tom Gordon	January 1, 2000
OL81625W	The Waste Lands	April 1992
OL81615W	Wizard and Glass	November 1, 1997
OL77014W	The Street Lawyer	\N
OL23480W	Red Dragon	1981
OL3454854W	Jaws	1976
OL36626W	The scapegoat	1957
OL3899224W	Roseanna	\N
OL81624W	Hearts in Atlantis	January 1, 2000
OL81586W	From a Buick 8	December 2002
OL2794726W	The Regulators	July 17, 1997
OL510879W	A Taste for Death	1991
OL81594W	Wolves of the Calla	November 4, 2003
OL553754W	The Black Dahlia	\N
OL14917748W	Under the Dome	\N
OL46913W	Congo	\N
OL16239762W	Gone Girl	\N
OL167179W	The Big Bad Wolf	2003
OL11913975W	The jumping off place	\N
OL22914W	Night Fall	2004
OL41256W	The Collectors	\N
OL167155W	Step on a Crack	2007
OL15168588W	The Passage	\N
OL46904W	Next	\N
OL167160W	Honeymoon	2005
OL15008W	The Collector	June 1983
OL167161W	The Jester	2003
OL41249W	Stone Cold	\N
OL181821W	The Night Manager	1993
OL134880W	Dawn	1990
OL134885W	Secrets of the Morning	1991
OL167162W	Lifeguard	2005
OL15165640W	El Juego del Ángel	\N
OL167150W	Double Cross	2007
OL167174W	The Angel Experiment	2005
OL167152W	The Quickie	2007
OL9302808W	The Lace Reader	\N
OL14920152W	Private	\N
OL5842017W	Dark Places	\N
OL5842018W	Sharp Objects	\N
OL3297218W	The Night-Comers	1956
OL16806568W	NEVER GO BACK	\N
OL86274W	Missing pieces	\N
OL448959W	Red Rabbit	\N
OL20126932W	The Institute	\N
OL5337429W	The 6th Target	\N
OL106064W	It's a Battlefield	\N
OL167333W	The Crush	2002
OL59431W	The Cry of the Owl	\N
OL17116913W	Gray Mountain	\N
OL167177W	7th Heaven	2008
OL19356257W	The Fix	\N
OL17079190W	Burn	\N
OL17725006W	The Target	\N
OL16708051W	El prisionero del cielo	\N
OL19356256W	End Game	\N
OL3917426W	The Vivero Letter	\N
OL17868937W	The murder house	\N
OL1822016W	The land God gave to Cain	1958
OL15191187W	Faithful Place	\N
OL18081210W	Pouillé historique de l'archevêché de Rennes	\N
OL874049W	Rose	\N
OL106059W	England made me	\N
OL19356870W	The 17th Suspect	\N
OL18147682W	The Woman in the Window: A Novel	\N
OL16809825W	Cross My heart	\N
OL15679291W	Den orolige mannen	\N
OL278110W	Heart-Shaped Box	2007
OL17929924W	The Chemist	\N
OL5718957W	The Dead Hour	\N
OL17930362W	The Rooster Bar	\N
OL17062554W	King and Maxwell	\N
OL17426934W	Nypd Red 2	\N
OL27641W	Babies in Toyland (Rugrats)	March 2002
OL17606520W	The Widow	\N
OL15284W	Der Verlorene	1998
OL261852W	Paris Trout	1993
OL17306744W	Gone	\N
OL17356815W	Missing You	\N
OL17283721W	Private L.A.	\N
OL17358795W	Dark Matter	\N
OL19865381W	The Chef	\N
OL491818W	Cover of Night	\N
OL17359906W	Bullseye	\N
OL20036181W	Odessa Sea	\N
\.


--
-- TOC entry 3669 (class 0 OID 24684)
-- Dependencies: 219
-- Data for Name: work_authors; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.work_authors (wid, aid) FROM stdin;
OL76837W	OL39307A
OL76833W	OL39307A
OL14873315W	OL39307A
OL81634W	OL19981A
OL1914022W	OL229268A
OL81613W	OL19981A
OL81628W	OL19981A
OL76835W	OL39307A
OL81631W	OL19981A
OL81630W	OL19981A
OL76836W	OL39307A
OL81618W	OL19981A
OL77001W	OL39329A
OL46876W	OL28257A
OL81609W	OL19981A
OL81612W	OL19981A
OL81625W	OL19981A
OL81615W	OL19981A
OL77014W	OL39329A
OL23480W	OL2032671A
OL3454854W	OL575390A
OL36626W	OL34047A
OL3899224W	OL709121A
OL3899224W	OL1460559A
OL3899224W	OL8654905A
OL3899224W	OL2783112A
OL3899224W	OL7342508A
OL3899224W	OL9182880A
OL81624W	OL19981A
OL81586W	OL19981A
OL2794726W	OL19981A
OL510879W	OL33909A
OL81594W	OL19981A
OL553754W	OL40671A
OL14917748W	OL19981A
OL46913W	OL28257A
OL16239762W	OL1433006A
OL167179W	OL22258A
OL11913975W	OL5113109A
OL22914W	OL25714A
OL41256W	OL28165A
OL167155W	OL22258A
OL167155W	OL228669A
OL167155W	OL10370207A
OL15168588W	OL896513A
OL46904W	OL28257A
OL46904W	OL7945315A
OL167160W	OL22258A
OL167160W	OL1515334A
OL15008W	OL6474824A
OL167161W	OL22258A
OL167161W	OL2631878A
OL41249W	OL28165A
OL181821W	OL2101074A
OL134880W	OL22019A
OL134885W	OL22019A
OL167162W	OL22258A
OL167162W	OL2631878A
OL15165640W	OL2632116A
OL167150W	OL22258A
OL167174W	OL22258A
OL167174W	OL3123066A
OL167152W	OL22258A
OL167152W	OL228669A
OL9302808W	OL3353071A
OL14920152W	OL22258A
OL14920152W	OL765158A
OL5842017W	OL1433006A
OL5842018W	OL1433006A
OL3297218W	OL539397A
OL16806568W	OL34328A
OL86274W	OL31818A
OL448959W	OL25277A
OL20126932W	OL19981A
OL5337429W	OL22258A
OL5337429W	OL765158A
OL106064W	OL20243A
OL167333W	OL22261A
OL59431W	OL28577A
OL17116913W	OL39329A
OL167177W	OL22258A
OL167177W	OL765158A
OL19356257W	OL28165A
OL17079190W	OL22258A
OL17079190W	OL228669A
OL17725006W	OL28165A
OL16708051W	OL2632116A
OL19356256W	OL28165A
OL3917426W	OL713259A
OL17868937W	OL22258A
OL17868937W	OL3288848A
OL1822016W	OL6925017A
OL15191187W	OL2660362A
OL874049W	OL76547A
OL106059W	OL20243A
OL19356870W	OL22258A
OL19356870W	OL765158A
OL18147682W	OL7442916A
OL16809825W	OL22258A
OL15679291W	OL45364A
OL278110W	OL2631898A
OL17929924W	OL1391085A
OL5718957W	OL1390877A
OL17930362W	OL39329A
OL17062554W	OL28165A
OL17426934W	OL22258A
OL17426934W	OL1372814A
OL27641W	OL26369A
OL17606520W	OL7287124A
OL15284W	OL23073A
OL261852W	OL444013A
OL17306744W	OL22258A
OL17306744W	OL228669A
OL17356815W	OL39821A
OL17356815W	OL7871059A
OL17283721W	OL22258A
OL17283721W	OL2757233A
OL17358795W	OL1433920A
OL19865381W	OL22258A
OL19865381W	OL7566140A
OL491818W	OL30714A
OL17359906W	OL22258A
OL17359906W	OL228669A
OL20036181W	OL29079A
OL20036181W	OL2660314A
\.


--
-- TOC entry 3683 (class 0 OID 0)
-- Dependencies: 220
-- Name: bio_bid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.bio_bid_seq', 208, true);


--
-- TOC entry 3684 (class 0 OID 0)
-- Dependencies: 222
-- Name: rating_rid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.rating_rid_seq', 100, true);


--
-- TOC entry 3492 (class 2606 OID 24678)
-- Name: authors authors_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.authors
    ADD CONSTRAINT authors_pkey PRIMARY KEY (aid);


--
-- TOC entry 3500 (class 2606 OID 24708)
-- Name: bio bio_aid_source_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.bio
    ADD CONSTRAINT bio_aid_source_key UNIQUE (aid, source);


--
-- TOC entry 3502 (class 2606 OID 24706)
-- Name: bio bio_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.bio
    ADD CONSTRAINT bio_pkey PRIMARY KEY (aid, bid);


--
-- TOC entry 3509 (class 2606 OID 24742)
-- Name: digital digital_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.digital
    ADD CONSTRAINT digital_pkey PRIMARY KEY (isbn10);


--
-- TOC entry 3507 (class 2606 OID 24732)
-- Name: edition edition_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.edition
    ADD CONSTRAINT edition_pkey PRIMARY KEY (isbn10);


--
-- TOC entry 3511 (class 2606 OID 24752)
-- Name: physical physical_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.physical
    ADD CONSTRAINT physical_pkey PRIMARY KEY (isbn10);


--
-- TOC entry 3505 (class 2606 OID 24722)
-- Name: rating rating_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.rating
    ADD CONSTRAINT rating_pkey PRIMARY KEY (rid);


--
-- TOC entry 3498 (class 2606 OID 24688)
-- Name: work_authors work_authors_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.work_authors
    ADD CONSTRAINT work_authors_pkey PRIMARY KEY (wid, aid);


--
-- TOC entry 3495 (class 2606 OID 24683)
-- Name: work work_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.work
    ADD CONSTRAINT work_pkey PRIMARY KEY (wid);


--
-- TOC entry 3503 (class 1259 OID 24761)
-- Name: idx_bio_aid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_bio_aid ON public.bio USING btree (aid);


--
-- TOC entry 3496 (class 1259 OID 24760)
-- Name: idx_work_authors_wid; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_work_authors_wid ON public.work_authors USING btree (wid);


--
-- TOC entry 3493 (class 1259 OID 24762)
-- Name: idx_work_title; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_work_title ON public.work USING btree (title);


--
-- TOC entry 3519 (class 2620 OID 24759)
-- Name: edition isbn10_trigger; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER isbn10_trigger BEFORE INSERT OR UPDATE ON public.edition FOR EACH ROW EXECUTE FUNCTION public.validate_isbn10();


--
-- TOC entry 3514 (class 2606 OID 24709)
-- Name: bio bio_aid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.bio
    ADD CONSTRAINT bio_aid_fkey FOREIGN KEY (aid) REFERENCES public.authors(aid) ON DELETE CASCADE;


--
-- TOC entry 3517 (class 2606 OID 24743)
-- Name: digital digital_isbn10_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.digital
    ADD CONSTRAINT digital_isbn10_fkey FOREIGN KEY (isbn10) REFERENCES public.edition(isbn10) ON DELETE CASCADE;


--
-- TOC entry 3516 (class 2606 OID 24733)
-- Name: edition edition_wid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.edition
    ADD CONSTRAINT edition_wid_fkey FOREIGN KEY (wid) REFERENCES public.work(wid) ON DELETE CASCADE;


--
-- TOC entry 3518 (class 2606 OID 24753)
-- Name: physical physical_isbn10_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.physical
    ADD CONSTRAINT physical_isbn10_fkey FOREIGN KEY (isbn10) REFERENCES public.edition(isbn10) ON DELETE CASCADE;


--
-- TOC entry 3515 (class 2606 OID 24723)
-- Name: rating rating_wid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.rating
    ADD CONSTRAINT rating_wid_fkey FOREIGN KEY (wid) REFERENCES public.work(wid) ON DELETE CASCADE;


--
-- TOC entry 3512 (class 2606 OID 24694)
-- Name: work_authors work_authors_aid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.work_authors
    ADD CONSTRAINT work_authors_aid_fkey FOREIGN KEY (aid) REFERENCES public.authors(aid) ON DELETE CASCADE;


--
-- TOC entry 3513 (class 2606 OID 24689)
-- Name: work_authors work_authors_wid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.work_authors
    ADD CONSTRAINT work_authors_wid_fkey FOREIGN KEY (wid) REFERENCES public.work(wid) ON DELETE CASCADE;


-- Completed on 2024-11-22 18:28:24 EST

--
-- PostgreSQL database dump complete
--

