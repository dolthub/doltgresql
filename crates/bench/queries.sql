-- The benchmark's setup and queries. A `-- setup` line starts the statements that build the data, and each
-- `-- name: x` line starts a timed query, which may hold several statements. `-- extended` after a name runs the query
-- through the extended protocol. `-- write` marks a query that changes data, whose results are not compared.

-- setup
CREATE TABLE items (id INT PRIMARY KEY, category INT NOT NULL, price NUMERIC(10, 2), qty INT, name TEXT, created TIMESTAMP, flag BOOLEAN);
CREATE INDEX items_category ON items (category);
CREATE INDEX items_qty_price ON items (qty, price);
INSERT INTO items SELECT i, i % 50, (i % 997) * 1.25, i % 113, 'item ' || i, TIMESTAMP '2024-01-01' + INTERVAL '1 minute' * i::float8, i % 3 = 0 FROM generate_series(1, 50000) i;
CREATE TABLE orders (id INT PRIMARY KEY, item_id INT NOT NULL, customer_id INT NOT NULL, amount FLOAT8, status VARCHAR(10), note TEXT);
CREATE INDEX orders_item ON orders (item_id);
INSERT INTO orders SELECT i, (i * 7) % 50000 + 1, i % 2000, (i % 1000) / 7.0, CASE i % 4 WHEN 0 THEN 'new' WHEN 1 THEN 'paid' WHEN 2 THEN 'shipped' ELSE 'closed' END, md5(i::text) FROM generate_series(1, 100000) i;
CREATE TABLE customers (id INT PRIMARY KEY, name TEXT, region INT);
INSERT INTO customers SELECT i, 'customer ' || i, i % 10 FROM generate_series(0, 1999) i;
CREATE TABLE regions (id INT PRIMARY KEY, name TEXT);
INSERT INTO regions SELECT i, 'region ' || i FROM generate_series(0, 9) i;
CREATE TABLE events (ts TIMESTAMP, kind TEXT, payload JSONB);
INSERT INTO events SELECT TIMESTAMP '2024-01-01' + INTERVAL '1 second' * i::float8, 'kind' || (i % 7), ('{"n": ' || i || ', "tag": "tag' || (i % 13) || '"}')::jsonb FROM generate_series(1, 20000) i;
CREATE TABLE tree (id INT PRIMARY KEY, parent INT);
INSERT INTO tree SELECT i, i / 2 FROM generate_series(1, 5000) i;
CREATE TABLE scratch (id INT PRIMARY KEY, v TEXT);

-- name: pk_point
SELECT * FROM items WHERE id = 25000;

-- name: pk_point_extended
-- extended
SELECT * FROM items WHERE id = 25001;

-- name: pk_in_list
SELECT id, name FROM items WHERE id IN (5, 500, 5000, 7000, 12000, 18000, 25000, 33000, 41000, 49999);

-- name: pk_range
SELECT id, price FROM items WHERE id BETWEEN 1000 AND 1100;

-- name: pk_range_open
SELECT count(*) FROM items WHERE id < 100;

-- name: secondary_count
SELECT count(*) FROM items WHERE category = 7;

-- name: secondary_rows
SELECT id, name FROM items WHERE category = 7;

-- name: secondary_ranges
SELECT count(qty) FROM items WHERE qty BETWEEN 10 AND 12 OR qty BETWEEN 50 AND 52 OR qty BETWEEN 90 AND 91;

-- name: covering_index
SELECT qty, price FROM items WHERE qty = 5;

-- name: composite_prefix_range
SELECT id FROM items WHERE qty = 5 AND price > 600;

-- name: count_star
SELECT count(*) FROM orders;

-- name: scan_filter
SELECT count(*) FROM orders WHERE amount > 100 AND status = 'paid';

-- name: scan_like
SELECT count(*) FROM items WHERE name LIKE 'item 12%';

-- name: scan_sum_expr
SELECT sum(price * qty) FROM items;

-- name: scan_all_rows
SELECT * FROM customers;

-- name: group_small
SELECT category, count(*), sum(price), avg(qty) FROM items GROUP BY category ORDER BY category;

-- name: group_large
SELECT customer_id, count(*), max(amount) FROM orders GROUP BY customer_id ORDER BY customer_id;

-- name: group_having
SELECT item_id, count(*) FROM orders GROUP BY item_id HAVING count(*) > 2 ORDER BY item_id LIMIT 20;

-- name: count_distinct
SELECT count(DISTINCT customer_id) FROM orders;

-- name: order_limit_pk
SELECT * FROM items ORDER BY id LIMIT 10;

-- name: order_limit_pk_desc
SELECT * FROM items ORDER BY id DESC LIMIT 10;

-- name: order_limit_top_n
SELECT id, amount FROM orders ORDER BY amount DESC, id LIMIT 10;

-- name: order_full
SELECT id, amount FROM orders WHERE customer_id < 100 ORDER BY amount, id;

-- name: order_by_index
SELECT id FROM items ORDER BY category, id LIMIT 20;

-- name: offset_page
SELECT id, name FROM items ORDER BY id LIMIT 20 OFFSET 20000;

-- name: join_pk_lookup
SELECT o.id, i.name FROM orders o JOIN items i ON i.id = o.item_id WHERE o.id BETWEEN 100 AND 200;

-- name: join_index_lookup
SELECT count(*) FROM items i JOIN orders o ON o.item_id = i.id WHERE i.category = 3;

-- name: join_hash_group
SELECT c.region, count(*) FROM orders o JOIN customers c ON c.id = o.customer_id GROUP BY c.region ORDER BY c.region;

-- name: join_three
SELECT r.name, sum(o.amount) FROM orders o JOIN customers c ON c.id = o.customer_id JOIN regions r ON r.id = c.region WHERE o.status = 'new' GROUP BY r.name ORDER BY r.name;

-- name: join_left
SELECT count(*) FROM customers c LEFT JOIN orders o ON o.customer_id = c.id AND o.status = 'new';

-- name: join_order_limit
SELECT o.id, i.name FROM orders o JOIN items i ON i.id = o.item_id ORDER BY o.id LIMIT 10;

-- name: in_subquery
SELECT count(*) FROM items WHERE id IN (SELECT item_id FROM orders WHERE customer_id = 5);

-- name: exists_correlated
SELECT count(*) FROM customers c WHERE EXISTS (SELECT 1 FROM orders o WHERE o.customer_id = c.id AND o.amount > 142);

-- name: not_exists
SELECT count(*) FROM items i WHERE NOT EXISTS (SELECT 1 FROM orders o WHERE o.item_id = i.id);

-- name: scalar_correlated
SELECT i.id, (SELECT count(*) FROM orders o WHERE o.item_id = i.id) FROM items i WHERE i.id <= 200 ORDER BY i.id;

-- name: scalar_uncorrelated
SELECT id FROM items WHERE price > (SELECT avg(price) FROM items) ORDER BY id LIMIT 10;

-- name: cte_aggregate
WITH totals AS (SELECT customer_id, sum(amount) AS total FROM orders GROUP BY customer_id) SELECT count(*) FROM totals WHERE total > 3500;

-- name: recursive_series
WITH RECURSIVE r(n) AS (SELECT 1 UNION ALL SELECT n + 1 FROM r WHERE n < 1000) SELECT sum(n) FROM r;

-- name: recursive_tree
WITH RECURSIVE sub(id) AS (SELECT id FROM tree WHERE id = 3 UNION ALL SELECT t.id FROM tree t JOIN sub ON t.parent = sub.id) SELECT count(*) FROM sub;

-- name: window_rank
SELECT id, rank() OVER (PARTITION BY category ORDER BY price, id) FROM items WHERE id <= 5000 ORDER BY id LIMIT 50;

-- name: window_running_sum
SELECT max(running) FROM (SELECT sum(amount) OVER (ORDER BY id) AS running FROM orders WHERE id <= 10000) s;

-- name: distinct_small
SELECT DISTINCT status FROM orders ORDER BY status;

-- name: union_distinct
SELECT count(*) FROM (SELECT customer_id FROM orders WHERE id < 1000 UNION SELECT id FROM customers WHERE region = 3) u;

-- name: string_functions
SELECT count(*) FROM orders WHERE upper(substr(note, 1, 2)) = 'AB' AND length(note) = 32;

-- name: jsonb_filter
SELECT count(*) FROM events WHERE payload->>'tag' = 'tag5';

-- name: date_trunc_group
SELECT date_trunc('day', created) AS day, count(*) FROM items GROUP BY day ORDER BY day LIMIT 5;

-- name: case_expression
SELECT sum(CASE WHEN qty > 50 THEN price ELSE 0 END), sum(CASE status WHEN 'x' THEN 1 ELSE 0 END) FROM items, (VALUES ('y')) v(status);

-- name: update_pk
-- write
UPDATE items SET flag = NOT flag WHERE id = 100;

-- name: update_range
-- write
UPDATE orders SET note = note || '' WHERE id BETWEEN 1 AND 100;

-- name: insert_delete
-- write
INSERT INTO scratch SELECT i, 'v' || i FROM generate_series(1, 100) i; DELETE FROM scratch;

-- name: insert_single
-- write
INSERT INTO scratch VALUES (1, 'one'); DELETE FROM scratch WHERE id = 1;
