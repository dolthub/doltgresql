## PostgreSQL ORM Tests
We created smoke tests for Doltgres's PostgreSQL ORM library integrations, 
and we run these tests through GitHub Actions on pull requests.

These tests can be run locally using Docker. From the doltgresql directory of the repo, run:

```bash
$ docker build -t doltgres-orm-tests -f testing/DoltgresORMDockerfile .
$ docker run doltgres-orm-tests:latest
```

The `docker build` step will take a few minutes to complete 
as it needs to install all the dependencies in the image.

Running the built container will produce output like:
```bash
$ docker run orm-tests:latest                                 
Running Doltgres orm-tests:
1..1
ok 1 drizzle smoke test
```
docker run doltgres-orm-tests:latest
