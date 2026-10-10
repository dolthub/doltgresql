// Copyright 2026 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//! Command-line style arguments of Dolt's procedures, parsed as Dolt's argparser does.

use std::collections::HashMap;

use crate::error::{PgError, Result};

/// Kind is how an option takes a value.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Kind {
    Flag,
    Value,
    /// A value that may be left out, which takes the next argument only when it is not an option.
    OptionalValue,
}

/// Opt is an option: its long name, its abbreviation or an empty string, and how it takes a value.
pub type Opt = (&'static str, &'static str, Kind);

/// Parser parses a command's arguments.
pub struct Parser {
    pub command: &'static str,
    pub options: &'static [Opt],
    /// The most positional arguments the command takes, if limited.
    pub max_args: Option<usize>,
}

/// Parsed is the options and positional arguments of a command.
#[derive(Debug, Default)]
pub struct Parsed {
    named: HashMap<&'static str, String>,
    pub args: Vec<String>,
    /// Where `--` separated the positional arguments, if it did.
    pub separator: Option<usize>,
}

impl Parsed {
    /// has reports whether the option was given.
    pub fn has(&self, name: &str) -> bool {
        self.named.contains_key(name)
    }

    /// value returns the option's value, if it was given.
    pub fn value(&self, name: &str) -> Option<&str> {
        self.named.get(name).map(String::as_str)
    }
}

/// error returns a Dolt procedure's error.
pub fn error(message: impl std::fmt::Display) -> PgError {
    PgError::internal(message)
}

/// argument_count returns go-mysql-server's error for a table function given the wrong number of arguments, which
/// Doltgres reports as an undefined function.
pub fn argument_count(function: &str, expected: impl std::fmt::Display, received: usize) -> PgError {
    PgError::new(
        crate::error::code::UNDEFINED_FUNCTION,
        format!("function '{function}' expected {expected} arguments, {received} received"),
    )
}

impl Parser {
    /// names returns the option names and abbreviations of one kind, longest first.
    fn names(&self, flags: bool) -> Vec<(&'static str, &'static Opt)> {
        let mut names: Vec<(&'static str, &'static Opt)> = self
            .options
            .iter()
            .filter(|o| (o.2 == Kind::Flag) == flags)
            .flat_map(|o| [(o.0, o), (o.1, o)])
            .filter(|(n, _)| !n.is_empty())
            .collect();
        names.sort_by(|a, b| b.0.len().cmp(&a.0.len()).then(a.0.cmp(b.0)));
        names
    }

    /// is_option reports whether an argument names one of the options.
    fn is_option(&self, arg: &str) -> bool {
        let Some(name) = arg.strip_prefix('-') else { return false };
        let name = name.strip_prefix('-').unwrap_or(name);
        self.options.iter().any(|o| o.0 == name || (!o.1.is_empty() && o.1 == name))
    }

    /// parse parses the arguments.
    pub fn parse(&self, args: &[String]) -> Result<Parsed> {
        let mut parsed = Parsed::default();
        let mut only_positional = false;
        let mut index = 0;
        while index < args.len() {
            let arg = &args[index];
            if arg.is_empty() || !arg.starts_with('-') || only_positional {
                parsed.args.push(arg.clone());
            } else if arg == "--" {
                only_positional = true;
                parsed.separator = Some(parsed.args.len());
            } else {
                index = self.parse_token(args, index, &mut parsed)?;
            }
            index += 1;
        }
        if let Some(max) = self.max_args
            && parsed.args.len() > max
        {
            let found = parsed.args.join(", ");
            return Err(error(if max == 0 {
                format!(
                    "error: {} does not take positional arguments, but found {}: {found}",
                    self.command,
                    parsed.args.len()
                )
            } else {
                format!(
                    "error: {} has too many positional arguments. Expected at most {max}, found {}: {found}",
                    self.command,
                    parsed.args.len()
                )
            }));
        }
        Ok(parsed)
    }

    /// parse_token parses the option argument at the index, returning the index of the last argument it used.
    fn parse_token(&self, args: &[String], mut index: usize, parsed: &mut Parsed) -> Result<usize> {
        let long = args[index].starts_with("--");
        let arg = args[index].trim_start_matches('-');
        let values = self.names(false);
        let mut rest = arg;
        let mut flags = self.names(true);
        let mut matched = false;
        'outer: loop {
            for (name, _) in &values {
                if rest == *name || rest.starts_with(&format!("{name}=")) {
                    break 'outer;
                }
            }
            for i in 0..flags.len() {
                let (name, opt) = flags[i];
                if rest.starts_with(name) {
                    rest = &rest[name.len()..];
                    if parsed.named.insert(opt.0, String::new()).is_some() {
                        return Err(error(format!("error: multiple values provided for `{}'", opt.0)));
                    }
                    flags.retain(|(_, o)| o.0 != opt.0);
                    matched = true;
                    continue 'outer;
                }
            }
            break;
        }
        let found = values.iter().find(|(name, _)| rest.starts_with(name));
        let Some(&(name, opt)) = found else {
            if rest.is_empty() {
                return Ok(index);
            }
            if matched {
                parsed.args.push(rest.to_string());
                return Ok(index);
            }
            return Err(error(format!("error: unknown option `{arg}'")));
        };
        let mut value = &rest[name.len()..];
        if !value.is_empty() && !" =:".contains(&value[..1]) && long {
            return Err(error(format!("error: unknown option `{arg}'")));
        }
        value = value.trim_start_matches([' ', '=', ':']);
        if parsed.named.contains_key(opt.0) {
            return Err(error(format!("error: multiple values provided for `{}'", opt.0)));
        }
        let value = if !value.is_empty() {
            value.to_string()
        } else if index + 1 >= args.len() {
            if opt.2 != Kind::OptionalValue {
                return Err(error(format!("error: no value for option `{}'", opt.0)));
            }
            String::new()
        } else if opt.2 == Kind::OptionalValue && (args[index + 1] == "--" || self.is_option(&args[index + 1])) {
            String::new()
        } else {
            index += 1;
            args[index].clone()
        };
        parsed.named.insert(opt.0, value);
        Ok(index)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    const COMMIT: Parser = Parser {
        command: "commit",
        options: &[
            ("message", "m", Kind::Value),
            ("all", "a", Kind::Flag),
            ("ALL", "A", Kind::Flag),
            ("allow-empty", "", Kind::Flag),
            ("author", "", Kind::Value),
        ],
        max_args: Some(0),
    };

    /// strings converts arguments to owned strings.
    fn strings(args: &[&str]) -> Vec<String> {
        args.iter().map(|s| s.to_string()).collect()
    }

    #[test]
    fn arguments_parse_as_dolt_parses_them() {
        let parsed = COMMIT.parse(&strings(&["-Am", "first commit"])).unwrap();
        assert!(parsed.has("ALL") && !parsed.has("all"));
        assert_eq!(parsed.value("message"), Some("first commit"));
        let parsed = COMMIT.parse(&strings(&["--allow-empty", "--message=x", "--author", "A <a@b.c>"])).unwrap();
        assert!(parsed.has("allow-empty"));
        assert_eq!(parsed.value("message"), Some("x"));
        assert_eq!(parsed.value("author"), Some("A <a@b.c>"));
        assert_eq!(COMMIT.parse(&strings(&["-mhello"])).unwrap().value("message"), Some("hello"));
        let err = COMMIT.parse(&strings(&["-z"])).unwrap_err();
        assert_eq!(err.message, "error: unknown option `z'");
        let err = COMMIT.parse(&strings(&["-m"])).unwrap_err();
        assert_eq!(err.message, "error: no value for option `message'");
        let err = COMMIT.parse(&strings(&["-m", "a", "b"])).unwrap_err();
        assert_eq!(err.message, "error: commit does not take positional arguments, but found 1: b");
    }
}
