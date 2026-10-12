#!/usr/bin/env python3
# Copyright 2026 Dolthub, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Stands in for dolt on the oracle tests' Dolt side and records what Dolt does.

Every command runs dolt itself (`init` through dolt.sh), and is appended to the JSON lines file that ORACLE_RECORD_LOG
names, with the suite that ORACLE_SUITE names. A `sql` session runs with a marker query after each statement and its
results in CSV, so that the record holds each statement's own result, its columns and rows with NULL as None or its
error, and the session's results are then written in the format it asked for, without the markers.
"""

import csv
import io
import json
import os
import re
import subprocess
import sys

DOLT = os.environ["DOLT_REAL"]
HERE = os.path.dirname(os.path.abspath(__file__))
MARKER = "oracle_marker"


def split_statements(sql):
    """Returns the statements of a MySQL script, split at semicolons outside quotes and comments."""
    statements, current, quote, i = [], [], None, 0
    while i < len(sql):
        c = sql[i]
        if quote:
            current.append(c)
            if c == "\\" and quote != "`" and i + 1 < len(sql):
                current.append(sql[i + 1])
                i += 1
            elif c == quote:
                quote = None
        elif c in "'\"`":
            quote = c
            current.append(c)
        elif sql.startswith("--", i) or c == "#":
            end = sql.find("\n", i)
            i = len(sql) if end < 0 else end
            continue
        elif c == ";":
            statements.append("".join(current).strip())
            current = []
        else:
            current.append(c)
        i += 1
    statements.append("".join(current).strip())
    return [s for s in statements if s]


def read_csv(text):
    """Returns the rows of Dolt's CSV output, with an unquoted empty field as None."""
    rows, row, field, quoted, i = [], [], [], False, 0
    pending = False
    while i < len(text):
        c = text[i]
        if quoted:
            if c == '"' and text.startswith('""', i):
                field.append('"')
                i += 1
            elif c == '"':
                quoted = False
            else:
                field.append(c)
        elif c == '"':
            quoted = True
            pending = True
        elif c == ",":
            row.append("".join(field) if field or pending else None)
            field, pending = [], False
        elif c == "\n":
            row.append("".join(field) if field or pending else None)
            rows.append(row)
            row, field, pending = [], [], False
        else:
            field.append(c)
        i += 1
    if field or pending or row:
        row.append("".join(field) if field or pending else None)
        rows.append(row)
    return rows


def marked_sql(statements):
    """Returns a session's statements with a marker query after each, and the line each statement starts on."""
    text, starts, line = [], [], 1
    for n, statement in enumerate(statements):
        starts.append(line)
        part = f"{statement};\nSELECT 'm{n}' AS {MARKER};\n"
        text.append(part)
        line += part.count("\n")
    return "".join(text), starts


def results(statements, starts, stdout, stderr):
    """Returns each statement's result from a marked session's output, matching errors to statements by line."""
    errors = {}
    for match in re.finditer(r"error on line (\d+) for query .*?: (.*?)(?=\nerror on line \d+ for query |\Z)", stderr, re.S):
        line = int(match.group(1))
        at = max((n for n, start in enumerate(starts) if start <= line), default=0)
        errors.setdefault(at, match.group(2).strip())
    lines = read_csv(stdout)
    out, i = [], 0
    for n, statement in enumerate(statements):
        sets = []
        while i < len(lines) and not (lines[i] == [MARKER] and i + 1 < len(lines) and lines[i + 1] == [f"m{n}"]):
            header = lines[i]
            i += 1
            rows = []
            while i < len(lines) and len(lines[i]) == len(header) and lines[i] != [MARKER]:
                rows.append(lines[i])
                i += 1
            if header != [None] or rows:
                sets.append({"columns": header, "rows": rows})
        out.append({"sql": statement, "error": errors.get(n), "results": sets})
        if i >= len(lines):
            break
        i += 2
    return out


def run(args, stdin):
    """Runs dolt, passing its output through, and returns its exit code and output."""
    command = [os.path.join(HERE, "dolt.sh")] + args if args[:1] == ["init"] else [DOLT] + args
    done = subprocess.run(command, input=stdin, stdin=None if stdin is not None else subprocess.DEVNULL,
                          capture_output=True)
    sys.stdout.buffer.write(done.stdout)
    sys.stderr.buffer.write(done.stderr)
    return done.returncode, done.stdout.decode(errors="replace"), done.stderr.decode(errors="replace")


def session(args, stdin):
    """Returns the global options, the options, the result format, whether the SQL came from -q, and the SQL of a sql
    session, or None for another command."""
    if "sql" not in args:
        return None
    at = args.index("sql")
    rest, query, result_format, i = [], None, "tabular", at + 1
    while i < len(args):
        if args[i] in ("-q", "--query") and i + 1 < len(args):
            query = args[i + 1]
            i += 2
        elif args[i] in ("-r", "--result-format") and i + 1 < len(args):
            result_format = args[i + 1]
            i += 2
        else:
            rest.append(args[i])
            i += 1
    sql = query if query is not None else (stdin or b"").decode(errors="replace")
    return args[:at], rest, result_format, query is not None, sql


def render(sets, result_format):
    """Writes result sets as dolt writes them in a result format: CSV, or tables otherwise."""
    out = io.StringIO()
    for columns, rows in sets:
        columns = ["" if c is None else c for c in columns]
        if result_format == "csv":
            writer = csv.writer(out, lineterminator="\n")
            writer.writerow(columns)
            for row in rows:
                writer.writerow(["" if v is None else v for v in row])
            continue
        cells = [["NULL" if v is None else v for v in row] for row in rows]
        widths = [max([len(c)] + [len(r[i]) for r in cells]) for i, c in enumerate(columns)]
        border = "+" + "+".join("-" * (w + 2) for w in widths) + "+\n"
        line = lambda values: "|" + "|".join(f" {v.ljust(w)} " for v, w in zip(values, widths)) + "|\n"
        out.write(border + line(columns) + border + "".join(line(r) for r in cells) + border + "\n")
    return out.getvalue()


def main():
    args = sys.argv[1:]
    reads = "sql" in args and not any(a in ("-q", "--query") for a in args) and not sys.stdin.isatty()
    stdin = sys.stdin.buffer.read() if reads else None
    parsed = session(args, stdin)
    entry = {"suite": os.environ.get("ORACLE_SUITE", ""), "cwd": os.getcwd(), "args": args}
    if parsed is None:
        rc, stdout, stderr = run(args, stdin)
        entry.update(rc=rc, stdout=stdout, stderr=stderr)
    else:
        before, flags, result_format, queried, sql = parsed
        statements = split_statements(sql)
        marked, starts = marked_sql(statements)
        command = [DOLT] + before + ["sql", "-r", "csv"] + flags
        done = subprocess.run(command + (["-q", marked] if queried else []),
                              input=None if queried else marked.encode(),
                              stdin=subprocess.DEVNULL if queried else None, capture_output=True)
        stdout, stderr = done.stdout.decode(errors="replace"), done.stderr.decode(errors="replace")
        recorded = results(statements, starts, stdout, stderr)
        rc = done.returncode
        entry.update(rc=rc, sql=sql, statements=recorded)
    with open(os.environ["ORACLE_RECORD_LOG"], "a") as log:
        log.write(json.dumps(entry) + "\n")
    if parsed is not None:
        sets = [(r["columns"], r["rows"]) for statement in recorded for r in statement["results"]]
        try:
            sys.stdout.write(render(sets, result_format))
        except Exception:
            sys.stdout.write(stdout)
        sys.stderr.write(stderr)
    sys.exit(rc)


if __name__ == "__main__":
    main()
