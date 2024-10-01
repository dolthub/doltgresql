CREATE TABLE test (pk int primary key);
INSERT INTO test VALUES (0), (1);

CREATE TABLE test_info (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references test(pk));

COPY test_info FROM STDIN WITH (DELIMITER '|', HEADER);
id|info|test_pk
4|string for 4|1
5|string for 5|0
6|string for 6|0
\.
