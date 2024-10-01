BEGIN;

CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));

-- NOTE: This is legacy syntax, but still in use and still supported by PostgreSQL
COPY tbl1 FROM STDIN CSV, HEADER;
pk,c1,c2
1,green,
2,"blue","a
q
u
a"
3,"brown",
4,"NULL",NULL
5,"?",""
6,"foo
\\.
bar","baz"
7,  ,' '
8," ",""
9,,''
\.

COMMIT;
