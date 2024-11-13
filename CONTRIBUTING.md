# Contribution Guide for DoltgreSQL

## Introduction

Welcome to the DoltgreSQL repository!
This guide outlines how to get started, in addition to how to modify the code.
We have a standard guideline that we try to adhere to for the project, and this guide will go over all of those.
I will assume familiarity with the Go toolchain.

## Getting Set Up

1. **Star the repository**: Since you're about to contribute to the project, why not give us a star to show your support? �
2. **Install the latest [Go](https://go.dev/dl/)**: We generally attempt to stay up-to-date with the latest version of Go.
   For the specific version in use, you may check the [go.mod](https://github.com/dolthub/doltgresql/blob/main/go.mod) file.
3. **Install PSQL 15**: Although we do not use PostgreSQL, we do use the PSQL client (for [PostgreSQL 15](https://www.postgresql.org/download/)) for testing.
   In addition, it is worthwhile to have a local PostgreSQL instance installed to cross-check behavior, especially when writing tests (which should _always_ be verified by a [PostgreSQL 15](https://www.postgresql.org/download/) instance).
4. **Clone the repository**: Clone the repository to a local directory of your choice.
5. **Build the parser**: Run the `doltgresql/postgres/parser/build.sh` script.
   This creates a few files within the `doltgresql/postgres/parser/parser` directory that are necessary for parsing PostgreSQL statements.
   It is recommended to run this file every time you pull changes into your local repository, as these generated files are not included since they would cause near guaranteed merge conflicts.
6. **Run Go tests**: Before building the project, you should always run all of the tests, which can be done by running `go test ./... --count=1` from the source root directory.
   This ensures that all Go tests pass, which also ensures that your Go environment is installed and configured correctly.
7. **Build the binary**: From the source root directory, run `go build -o <bin_name> ./cmd/doltgres`, where `<bin_name>` is the name of the binary (usually `doltgres` or `doltgres.exe`).
   To run the program without creating an executable, run `go run ./cmd/doltgres`.
8. **Run Bats tests**: We make use of [Bats](https://github.com/bats-core/bats-core) for all end-user-style tests.
   Assuming you have [NPM installed](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm), you can install [Bats](https://www.npmjs.com/package/bats) by running `npm install -g bats`.
   Then, navigate to `doltgresql/testing/bats` and run `bats .`, which will run all of the Bats tests in the directory.
   An alternative is to use [BashSupport Pro](https://plugins.jetbrains.com/plugin/13841-bashsupport-pro), which is cross-platform and used by several developers.
   Additionally, our [Bats](https://github.com/bats-core/bats-core) tests assume that you have a `doltgresql` (not `doltgres`) binary on your PATH.
   For Windows users, this means that the binary should _not_ end with the `.exe` file extension.
   Remember to recompile the executable on your PATH whenever you want to re-test any [Bats](https://github.com/bats-core/bats-core) tests.
9. **Change the data directory**: This is optional but recommended.
   By default, we create databases within the `~/doltgres/databases` directory.
   For developmental purposes, you may want to change this behavior. You have two options:
   1. Set the `DOLTGRES_DATA_DIR` environment variable to a different directory. A value of `.` causes DoltgreSQL to use the current directory as the data directory, so you can have multiple data directories simply by running the program in different directories. This behavior is more consistent with [Dolt's](https://github.com/dolthub/dolt) behavior. This is the recommended option for development.
   2. Specify the directory in the `--data-dir` argument. This overrides the environment variable if it is present.

### Note for Windows Users

All of the tooling in this repository is designed for a Unix-like environment.
Some of the lead project developers are on Windows, however they all prefer the Unix commandline environment over Command Prompt and Powershell.
These are a few alternatives that are used to provide a Unix environment:

- **[Git Bash](https://git-scm.com/downloads)**: The Mintty client that may be installed alongside Git.
- **[WSL 1](https://learn.microsoft.com/en-us/windows/wsl/install)**: The Linux subsystem for Windows, specifically version 1.
  We've attempted to use version 2 with [Dolt](https://github.com/dolthub/dolt), however we ran into a few issues.
  It's possible that it works, but since we do not test or support it, we heavily suggest sticking to version 1 instead.
- **[Cygwin](https://www.cygwin.com/install.html)**: Cygwin is a full Unix-like environment for Windows.
  Unlike WSL, which runs a full Linux kernel within Windows, Cygwin emulates Unix through native Windows processes.
  This has not been extensively tested, however some have had success using it for development.

## Submitting Changes

1. **Format the repository**: You can do this by running `doltgresql/scripts/format_repo.sh`.
   This will reformat any changes to adhere to our preferred style.
   This is required, otherwise the pull request's checks will not pass.
2. **Run all tests**: You should ensure that all tests are running correctly.
   All changes are required to have passing tests before they're merged.
   Additionally, most changes should have tests.
   We will not accept submissions that do not properly test to verify their behavior.
3. **Create a Pull Request**: Create your pull request and base it against the [main](https://github.com/dolthub/doltgresql/tree/main) branch.
   All pull requests should have a description stating what was fixed, and possibly why it was broken.
   If a specific issue is being fixed, then that issue should be linked to from the PR.
   In some cases, the PR is not properly tagged from within the issue, so [refer to this link for manually linking the PR within the issue](https://docs.github.com/en/issues/tracking-your-work-with-issues/linking-a-pull-request-to-an-issue#manually-linking-a-pull-request-to-an-issue-using-the-pull-request-sidebar).
4. **Address Feedback**: If any feedback was provided on the PR, then it must be addressed before it may be merged.
   In some cases, an approving review will not be given until we've re-examined the pull request.
5. **Merge the Pull Request**: Congratulations!
   You're now an official contributor of the [DoltgreSQL](https://github.com/dolthub/doltgresql) project.

## General Style Rules

Here are a few rules that apply to the entire codebase:

1. Comment your code as thoroughly as possible.
   It is better to over-comment your code than under-comment.
   All functions require a function comment, even if it's just stating that it implements an interface.
2. Remove dead code.
   Don't just put it in a comment block.
   It will still be in the source history, so it can be retrieved at a later time if needed.
   Commenting it out will potentially confuse another developer, who may think that it has some contextual significance.
   If it does, then it should be a proper comment rather than just commented-out code.
3. File names use [snake case](https://en.wikipedia.org/wiki/Snake_case).
4. Leave a `//TODO:` comment when there is more work left to be done in an area.
   This allows us to keep track of incomplete implementations, so that we don't assume a function is "complete".
   A common source of bugs is when another integrator assumes that a function is completely implemented, and interacts with that function under that assumption.
   If they find a `//TODO:`, then it may clue them in, and/or assist with debugging efforts.

## Rules for Specific Packages

Some packages have rules that only apply to those packages.
This is generally due to the size of those packages being larger than average, or some other unique reason that doesn't necessarily apply to the rest of the codebase.

### `postgres/parser/parser`

The nested `parser` naming may be confusing, but they're two distinct levels of the parser.
The first level, `postgres/parser`, contains all of the code for parsing a SQL statement into an AST.
The second level, `postgres/parser/parser`, specifically contains the YACC file, along with the generated build files to support the grammar contained in the YACC file.

The parser has been adapted from [CockroachDB](https://github.com/cockroachdb/cockroach/tree/f559e6a494e2f4b5dd329dc451be8c9695e3f831), using the most recent commit that falls under the Apache 2.0 License (see the [BSL](https://github.com/dolthub/doltgresql/blob/main/licenses/BSL.txt)).
As such, their [README](https://github.com/dolthub/doltgresql/blob/main/postgres/parser/parser/README.md) goes over how it's all structured, and we intend to continue in their style.
This means that we should continue to add the `%HELP` comments for statements, mark statements as errors when needed, etc.
This differs from our [Vitess](https://github.com/dolthub/vitess) fork, which is used by [GMS](https://github.com/dolthub/go-mysql-server) and [Dolt](https://github.com/dolthub/dolt).

### `testing/bats`

All Bats tests must follow this general structure:

```bash
@test "file-name: test name" {
    # Test Contents
}
```

`file-name` is the name of the file, without the `.bats` file extension.
`test name` is the unique name of the test within that file.
You may surround the name portion with either single quotes `'` or double quotes `"`, although double quotes may cause issues if the test name includes a backslash.

### `testing/go`

Tests within [`testing/go`](https://github.com/dolthub/doltgresql/tree/main/testing/go) are modeled after [engine tests in GMS](https://github.com/dolthub/go-mysql-server/tree/main/enginetest).
One key deviation is the [Focus](https://github.com/dolthub/doltgresql/blob/2246d40a5ec4b92661e526480b6d4af82f232583/testing/go/framework_test.go#L54) field.
Unlike in [GMS](https://github.com/dolthub/go-mysql-server), we do not have a skipped test that a developer would locally unskip and fill in with their test details.
This is by design, as such tests may never become an actual test within a file.
To support this workflow, the [Focus](https://github.com/dolthub/doltgresql/blob/2246d40a5ec4b92661e526480b6d4af82f232583/testing/go/framework_test.go#L54) field was added.

Whenever this field is set to `true`, _only_ focused tests are run.
This means that the developer is required to write a valid test, and then they may focus on that test as though it were the only test.
Once the developer is done, they'll simply delete the [Focus](https://github.com/dolthub/doltgresql/blob/2246d40a5ec4b92661e526480b6d4af82f232583/testing/go/framework_test.go#L54) line (defaults to `false`), and all of the tests will run.

This extends beyond adding new tests, as it allows a developer to focus on failing tests too.
Let's say that two specific tests are failing, you can apply [Focus](https://github.com/dolthub/doltgresql/blob/2246d40a5ec4b92661e526480b6d4af82f232583/testing/go/framework_test.go#L54) to both of those tests, and _only_ those tests will run.
Yes, [Focus](https://github.com/dolthub/doltgresql/blob/2246d40a5ec4b92661e526480b6d4af82f232583/testing/go/framework_test.go#L54) works on more than a single test.
In fact, if [Focus](https://github.com/dolthub/doltgresql/blob/2246d40a5ec4b92661e526480b6d4af82f232583/testing/go/framework_test.go#L54) is present on any tests, then it changes the testing mode such that only tests with [Focus](https://github.com/dolthub/doltgresql/blob/2246d40a5ec4b92661e526480b6d4af82f232583/testing/go/framework_test.go#L54) are run.

This has a major benefit over commenting out other tests too, in that it's safe for use with GitHub Actions.
If a developer accidentally forgets to disable [Focus](https://github.com/dolthub/doltgresql/blob/2246d40a5ec4b92661e526480b6d4af82f232583/testing/go/framework_test.go#L54), then the actions will fail, and report that [Focus](https://github.com/dolthub/doltgresql/blob/2246d40a5ec4b92661e526480b6d4af82f232583/testing/go/framework_test.go#L54) must be disabled.
Compare this with commenting out tests, where a developer forgetting to uncomment the tests will rely on their code reviewers to catch the mistake.
If this occurs after the code review, then it may make it into the main branch, where it may remain undiscovered for years.
This has almost happened multiple times in [GMS](https://github.com/dolthub/go-mysql-server), so this method of testing is a far safer alternative.
