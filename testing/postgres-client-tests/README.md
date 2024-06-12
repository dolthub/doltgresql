## PostgreSQL Client Tests
We created smoke tests for Doltgres's PostgreSQL client integrations, and we run these tests through GitHub Actions
on pull requests.

These tests can be run locally using Docker. From the doltgresql directory of the repo, run:

```bash
$ docker build -t postgres-client-tests -f testing/PostgresDockerfile .
$ docker run postgres-client-tests:latest
```

The `docker build` step will take a few minutes to complete as it needs to install all the dependencies in the image.

Running the built container will produce output like:
```bash
$ docker run postgres-client-tests:latest                                 
Running postgres-client-tests:
1..2
ok 1 postgres-connector-java client
ok 2 node postgres client
```
