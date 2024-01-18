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
6. **Run Go tests**: Before building the project, you should always run all of the tests, which can be done by running `go run test ./... --count=1` from the source root directory.
This ensures that all Go tests pass, which also ensures that your Go environment is installed and configured correctly.
7. **Build the binary**: From the source root directory, run `go build -o <bin_name> .`, where `<bin_name>` is the name of the binary (usually `doltgres` or `doltgres.exe`).
   To run the program without creating an executable, run `go run .`.
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
* **[Git Bash](https://git-scm.com/downloads)**: The Mintty client that may be installed alongside Git.
* **[WSL 1](https://learn.microsoft.com/en-us/windows/wsl/install)**: The Linux subsystem for Windows, specifically version 1.
We've attempted to use version 2 with [Dolt](https://github.com/dolthub/dolt), however we ran into a few issues.
It's possible that it works, but since we do not test or support it, we heavily suggest sticking to version 1 instead.
* **[Cygwin](https://www.cygwin.com/install.html)**: Cygwin is a full Unix-like environment for Windows.
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

## Project Structure

The project has a fairly straightforward directory structure, with each directory housing items relevant to the directory's name.

1. `licenses`: Contains all third-party licenses.
These are automatically included when releasing a new version of DoltgreSQL, so make sure to add any new licenses to this directory.
2. `postgres`: The primary directory that contains all aspects of PostgreSQL communication.
   1. `connection`: This handles the wire protocol, which writes messages to and from a connection.
   2. `messages`: Contains all of the messages that we support.
   3. `parser`: Contains all of the code pertaining to parsing SQL statements and generating an AST.
3. `scripts`: Contains all of the non-build scripts and tools.
4. `server`: The interface between the `postgres` directory and the underlying [GMS](https://github.com/dolthub/go-mysql-server) and [Dolt](https://github.com/dolthub/dolt) backends.
   1. `ast`: Specifically houses all transformations from the `postgres` AST to the [`vitess`](https://github.com/dolthub/vitess) AST.
5. `testing`: Contains all integration tests, and all things related to testing.
This will not contain _all_ tests within the repository, as functions within other directories may declare their own unit tests.
   1. `bats`: Contains all of our [Bats](https://github.com/bats-core/bats-core) tests.
   2. `generation`: Any tools that generate tests will go here. The generated tests should also go in this directory. Generated tests are useful for creating a structured set of verified tests.
   3. `go`: Contains all of the Go integration tests. These are further broken up by files, with some used for regression testing, etc.
   4. `logictest`: Contains the harness for running DoltgreSQL against the [sqllogictests](https://github.com/dolthub/sqllogictest).
6. `utils`: Contains all items that may be used across packages, such as a [Stack](https://github.com/dolthub/doltgresql/blob/main/utils/stack.go) structure.
No files here should rely on code from any other part of the project, making them safe to import without fear of cyclic references.

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

### `postgres/messages`

Each message needs an `init` function similar to the following:
```go
func init() {
	connection.InitializeDefaultMessage(MessageStruct{})
}
```
It is required that all messages are initialized, as the initialization also performs sanity checks to ensure that the message is properly formatted.
If the message contains a message header, then you need to add...
```go
connection.AddMessageHeader(MessageStruct{})
```
...after initializing the default message.
Message headers are always a `Byte1` at the very beginning of the message, although having a `Byte1` does not necessarily mean that it represents a header.
[Within the documentation](https://www.postgresql.org/docs/15/protocol-message-formats.html), they often have a description similar to _"Identifies the message as..."_.

Besides the aforementioned, the `Encode()` function for all messages _must_ make and return a copy of the default message.
Callers expect that they're able to modify the returned message, therefore returning the default message may lead to errors during runtime.

### `postgres/parser/parser`

The nested `parser` naming may be confusing, but they're two distinct levels of the parser.
The first level, `postgres/parser`, contains all of the code for parsing a SQL statement into an AST.
The second level, `postgres/parser/parser`, specifically contains the YACC file, along with the generated build files to support the grammar contained in the YACC file.

The parser has been adapted from [CockroachDB](https://github.com/cockroachdb/cockroach/tree/f559e6a494e2f4b5dd329dc451be8c9695e3f831), using the most recent commit that falls under the Apache 2.0 License (see the [BSL](https://github.com/dolthub/doltgresql/blob/main/licenses/BSL.txt)).
As such, their [README](https://github.com/dolthub/doltgresql/blob/main/postgres/parser/parser/README.md) goes over how it's all structured, and we intend to continue in their style.
This means that we should continue to add the `%HELP` comments for statements, mark statements as errors when needed, etc.
This differs from our [Vitess](https://github.com/dolthub/vitess) fork, which is used by [GMS](https://github.com/dolthub/go-mysql-server) and [Dolt](https://github.com/dolthub/dolt).

### `server/ast`

As mentioned in [Project Structure](#project-structure), this package contains all transformations that take a PostgreSQL AST (defined in [`postgres/parser/sem/tree`](https://github.com/dolthub/doltgresql/tree/main/postgres/parser/sem/tree)) and converts it to a [Vitess](https://github.com/dolthub/vitess) AST.
As DoltgreSQL is currently built on top of [GMS](https://github.com/dolthub/go-mysql-server), we need to convert our ASTs into a form that [GMS](https://github.com/dolthub/go-mysql-server) can understand.

This starts with the [`convert.go`](https://github.com/dolthub/doltgresql/blob/main/server/ast/convert.go) file, which takes in all PostgreSQL ASTs and dispatches to the correct function to start the transformation.
We refer to these ASTs as _nodes_, and it's very common that nodes embed other nodes.
Put another way, an AST is just a hierarchy of nodes.

#### Rule 1: Functions transforming nodes must start with `node`

Let's use the [CREATE TABLE](https://www.postgresql.org/docs/15/sql-createtable.html) AST `*tree.CreateTable` as example.
Our function will transform `*tree.CreateTable` into `*vitess.DDL`, which is used to represent CREATE TABLE in GMS.
As it's a transformation, the function name must start with `node`.
Next, the exact name of the PostgreSQL struct will follow.
This gives us the following signature:
```go
func nodeCreateTable(node *tree.CreateTable) (*vitess.DDL, error)
```
Even when the function does not return an error, it is better to include it and return a `nil` error, as it prevents the need for a future refactor in case the function later needs to return an error.

This naming strategy gives a few benefits:

1. It is trivial to find transformation functions.
2. Prevents function duplication.
Without a consistent naming scheme, one developer could write `transformCreateTbl` while another writes `handleTableCreation`, with both handling the same node.
3. Restricts the capabilities of the function, so that each transform only handles that transform, and does not handle additional responsibilities that would not be obvious from the name.

This also extends to functions that take the same input, but vary their output.
Let's use `tree.Expr` as an example.
```go
func nodeExpr(node tree.Expr) (vitess.Expr, error)

func nodeExprToSelectExpr(node tree.Expr) (vitess.SelectExpr, error)
```
Here we have two functions, both taking a `tree.Expr`, but both returning different values.
They have different use cases and are called from different nodes, thus they're behaviorally different.
In these cases, simply append `To<ReturnType>` to the name to differentiate the functions.
With this naming scheme, it's easy to discover all functions that transform a `tree.Expr`.

There has been debate regarding whether all such functions should be moved to a `node` subpackage, and this would be an acceptable change.
If this is something you want, then feel free to submit it as a PR.

#### Rule 2: Functions modifying nodes must start with `assign`

This is similar to the `node` rule, except that this applies to functions that modify nodes rather than transform them.
An example signature:
```go
func assignTableDef(node tree.TableDef, target *vitess.DDL) error
```
`tree.TableDef` is not a node that can be transformed into a [Vitess](https://github.com/dolthub/vitess) node.
Instead, it's purpose is to modify other nodes, which in this case is `*vitess.DDL`.

#### Rule 3: Functions modifying vitess AST nodes start with `translate`

Functions that modify a vitess expression, e.g. to perform additional or common translation logic
after other translation steps, should start with `translate`.

If there are any functions that do not fall into the categories above, and that function belongs in the `ast` package, then create an issue/PR and we will discuss it there.
If we add another function type, then this guide will also be updated.

#### Rule 4: Each node gets its own file

With very few exceptions, each node should have its own file.
While this will create many files, it also allows a developer to quickly find the file containing the functionality that they're interested in.

The main exceptions are interface nodes, such as `tree.CompositeDatum` (the other exception is listed in [Rule 4](#rule-4-node-slices-get-their-own-functions)).
All `tree.CompositeDatum` nodes are also `tree.Expr` nodes, and thus `nodeCompositeDatum` directly calls into `tree.Expr`.
In this case, `nodeCompositeDatum` is in [`expr.go`](https://github.com/dolthub/doltgresql/blob/main/server/ast/expr.go), since its implementation is essentially in [`expr.go`](https://github.com/dolthub/doltgresql/blob/main/server/ast/expr.go).
This is technically a discrepancy though, and all such exceptions should be moved to their own files if they cause any confusion.

#### Rule 5: Node slices get their own functions

Many nodes also have slice forms.
For example `tree.Expr` has `tree.Exprs`, which is defined as `type Exprs []Expr`.
In these cases, we should also include `nodeExprs` in addition to `nodeExpr`.
It may seem unnecessary, but it follows in the spirit of [Rule 1](#rule-1-functions-transforming-nodes-must-start-with-node), as it is technically a different type, and thus a different node.
In these cases, we generally do not create a new file, which technically violates [Rule 3](#rule-3-each-node-gets-its-own-file).
However, since most of these functions amount to mere convenience functions, we include them in the same file to reduce the total number of files.

#### Rule 6: All fields are accounted for

All functions must account for all fields, implementations, etc. of their respective inputs.
All interfaces must handle all integrators through a `switch` statement.
All structs must do something with every field (or add a `//TODO:` comment mentioning why the field is being ignored).
All enums must have a default case that returns an error.

For example, [`nodeTableExpr`](https://github.com/dolthub/doltgresql/blob/main/server/ast/table_expr.go) implements a transformation for the node `tree.TableExpr`, which is an interface.
Within the function body, we explicitly list all known integrators of the `tree.TableExpr` interface within a switch statement.
If any integrators are not yet supported, then we return an error explicitly stating that it's not yet supported.
In addition, we have a default case that returns an error, even though it's guaranteed to never run into the default case.
This default case allows us to properly error if we eventually add a new integrator to the `tree.TableExpr` interface, rather than continuing in some unintended state, potentially causing data corruption further in the pipeline.

Another example is [`nodeUpdate`](https://github.com/dolthub/doltgresql/blob/main/server/ast/update.go).
We do not yet support [`RETURNING` in `UPDATE`](https://www.postgresql.org/docs/15/sql-update.html), and therefore we return an error in the event that it has been specified.

The last example is [`nodeUnionClause`](https://github.com/dolthub/doltgresql/blob/main/server/ast/union_clause.go).
It contains a switch over the union type, and even though all types are accounted for, it has a default case that returns an error.

This is perhaps the most important rule of them all, as it will greatly enhance our correctness, and prevent subtle bugs from causing data corruption.
We should only handle statements that we know we can handle correctly.
There are exceptions, as some statements we do not yet support, and cannot support for quite a while, however we still need to handle them, such as the default NULL order for indexes.
In these cases, we must add a `//TODO:` comment stating what is missing and why it isn't an error.
This will at least allow us to track all such instances where we deviate from the expected behavior, which we can also document elsewhere for users of DoltgreSQL.

### `server/functions`

The `functions` package contains the functions, along with an implementation to approximate the function overloading structure (and type coercion).

The function overloading structure is defined in all files that have the `zinternal_` prefix.
Although not preferable, this was chosen as Go does not allow cyclical references between packages.
Rather than have half of the implementation in `functions`, and the other half in another package, the decision was made to include both in the `functions` package with the added prefix for distinction.

There's an `init` function in `server/functions/zinternal_catalog.go` (this is included in `server/listener.go`) that removes any conflicting GMS function names, and replaces them with the PostgreSQL equivalents.
This means that the functions that we've added behave as expected, and for others to have _some_ sort of implementation rather than outright failing.
We will eventually remove all GMS functions once all PostgreSQL functions have been implemented.
The other internal files all contribute to the generation of functions, along with their proper handling.

Each function (and all overloads) are contained in a single file.
Overloads are named according to their parameters, and prefixed by their target function name.
The set of overloads are then added to the `Catalog` within `server/functions/zinternal_catalog.go`.
To add a new function, it is as simple as creating the `Function`, adding the overloads, and adding it to the `Catalog`.

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
