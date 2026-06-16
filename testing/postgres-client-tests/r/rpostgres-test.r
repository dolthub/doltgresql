library(RPostgres)
library(DBI)

args = commandArgs(trailingOnly=TRUE)

user = args[1]
port = strtoi(args[2])

conn = dbConnect(RPostgres::Postgres(),
                 host="localhost",
                 port=port,
                 user=user,
                 password="password",
                 dbname="postgres")

# check standard queries
queries = list(
    "DROP TABLE IF EXISTS test",
    "create table test (pk int, value int, primary key(pk))",
    "select * from test",
    "insert into test (pk, value) values (0,0)",
    "select * from test")

responses = list(
    NULL,
    NULL,
    data.frame(pk = integer(0), value = integer(0), stringsAsFactors = FALSE),
    NULL,
    data.frame(pk = c(as.integer(0)), value = c(as.integer(0)), stringsAsFactors = FALSE))

for (i in 1:length(queries)) {
    q = queries[[i]]
    want = responses[[i]]
    if (!is.null(want)) {
        got <- dbGetQuery(conn, q)
        if (length(want) == length(got)) {
            for (j in 1:length(want)) {
                if (!identical(want[[j]], got[[j]])) {
                    print(q)
                    print(c("want:", want[[j]], "type: ", typeof(want[[j]])))
                    print(c("got:", got[[j]], "type: ", typeof(got[[j]])))
                    quit("no", 1)
                }
            }
        }
    } else {
        invisible(dbExecute(conn, q))
    }
}

dolt_queries = list(
    "select dolt_add('-A')",
    "select dolt_commit('-m', 'my commit')",
    "select dolt_checkout('-b', 'mybranch')",
    "insert into test (pk, value) values (1,1)",
    "select dolt_commit('-a', '-m', 'my commit2')",
    "select dolt_checkout('main')",
    "select dolt_merge('mybranch')")

for (i in 1:length(dolt_queries)) {
    q = dolt_queries[[i]]
    if (startsWith(trimws(tolower(q)), "select")) {
        dbGetQuery(conn, q)
    } else {
        invisible(dbExecute(conn, q))
    }
}

count <- dbGetQuery(conn, "select COUNT(*) as c from dolt.log")
want <- data.frame(c = c(4))
ret <- all.equal(count, want)
if (!isTRUE(ret)) {
    print("Number of commits is incorrect")
    print(count)
    quit("no", 1)
}

dbDisconnect(conn)
print("RPostgres test passed")
