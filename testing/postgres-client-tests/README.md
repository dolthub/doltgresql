## MySQL Client Tests
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
Running mysql-client-tests:
1..1
ok 1 mysql-connector-java client
```
