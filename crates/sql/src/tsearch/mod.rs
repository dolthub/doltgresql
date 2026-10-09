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

//! Text search configurations and dictionaries: the default parser's tokens run through the simple and English
//! dictionaries, as Postgres' parsetext, to_tsvector, and to_tsquery family do.

mod english;
pub(crate) mod parser;

use std::collections::HashSet;
use std::sync::OnceLock;

use crate::basetypes::textsearch::{Builder, Lexeme, MAX_POSITION, Mode, Operator, Query, parse_query};
use crate::error::{PgError, Result, code};

/// MAX_LEXEME is the length in bytes of the longest lexeme that a vector can hold.
pub(crate) const MAX_LEXEME: usize = 2047;

/// MAX_POSITIONS is how many positions to_tsvector keeps for a lexeme.
const MAX_POSITIONS: usize = 255;

/// Dictionary is a built-in text search dictionary that Doltgres implements.
#[derive(Clone, Copy, Debug, PartialEq)]
pub(crate) enum Dictionary {
    Simple,
    EnglishStem,
}

/// english_stop_words returns the English stop words, from the english.stop file that Postgres ships.
fn english_stop_words() -> &'static HashSet<&'static str> {
    static WORDS: OnceLock<HashSet<&'static str>> = OnceLock::new();
    WORDS.get_or_init(|| include_str!("english.stop").lines().map(str::trim_end).filter(|w| !w.is_empty()).collect())
}

/// lowercase lowercases text as Postgres' lowerstr does in the C locale, which changes only ASCII letters.
fn lowercase(text: &str) -> String {
    text.to_ascii_lowercase()
}

impl Dictionary {
    /// named returns the dictionary with a name, failing for the dictionaries that Doltgres does not implement.
    pub(crate) fn named(name: &str) -> Result<Dictionary> {
        match name {
            "simple" => Ok(Dictionary::Simple),
            "english_stem" => Ok(Dictionary::EnglishStem),
            _ => Err(PgError::unsupported(format!("text search dictionary \"{name}\""))),
        }
    }

    /// name returns the dictionary's name.
    pub(crate) fn name(self) -> &'static str {
        match self {
            Dictionary::Simple => "simple",
            Dictionary::EnglishStem => "english_stem",
        }
    }

    /// lexize returns the lexemes of a token, which are none for a stop word, as Postgres' dsimple_lexize and
    /// dsnowball_lexize do.
    pub(crate) fn lexize(self, token: &str) -> Vec<String> {
        let word = lowercase(token);
        match self {
            Dictionary::Simple if word.is_empty() => Vec::new(),
            Dictionary::Simple => vec![word],
            Dictionary::EnglishStem if token.len() > 1000 => vec![word],
            Dictionary::EnglishStem if word.is_empty() || english_stop_words().contains(word.as_str()) => Vec::new(),
            Dictionary::EnglishStem => vec![english::stem(&word)],
        }
    }
}

/// Config is a built-in text search configuration that Doltgres implements.
#[derive(Clone, Copy, Debug, PartialEq)]
pub(crate) enum Config {
    Simple,
    English,
}

impl Config {
    /// named returns the configuration with a name, which may name the pg_catalog schema, failing for the
    /// configurations that Doltgres does not implement.
    pub(crate) fn named(name: &str) -> Result<Config> {
        match name.strip_prefix("pg_catalog.").unwrap_or(name) {
            "simple" => Ok(Config::Simple),
            "english" => Ok(Config::English),
            _ => Err(PgError::unsupported(format!("text search configuration \"{name}\""))),
        }
    }

    /// dictionaries returns the dictionaries that tokens of a type go to, as pg_ts_config_map lists them.
    pub(crate) fn dictionaries(self, token_type: u8) -> &'static [Dictionary] {
        use parser::*;
        match (self, token_type) {
            (_, SPACE | TAG | PROTOCOL | XML_ENTITY) => &[],
            (Config::Simple, _) => &[Dictionary::Simple],
            (Config::English, ASCII_WORD | ASCII_PART_HWORD | ASCII_HWORD | WORD | PART_HWORD | HWORD) => {
                &[Dictionary::EnglishStem]
            }
            (Config::English, _) => &[Dictionary::Simple],
        }
    }
}

/// Word is a lexeme of parsed text and its position.
pub(crate) struct Word {
    pub(crate) text: String,
    pub(crate) position: u16,
}

/// long_word_notice returns the notice for a word that is too long to index.
pub(crate) fn long_word_notice() -> PgError {
    PgError {
        detail: Some(format!("Words longer than {MAX_LEXEME} characters are ignored.")),
        ..PgError::notice(code::PROGRAM_LIMIT_EXCEEDED, "word is too long to be indexed")
    }
}

/// parse_text returns the lexemes of text and their positions, counting positions on from `position`, as Postgres'
/// parsetext does, where a stop word takes a position but adds no lexeme.
pub(crate) fn parse_text(config: Config, text: &str, position: &mut u32, notices: &mut Vec<PgError>) -> Vec<Word> {
    let mut words = Vec::new();
    let mut tokens = parser::Parser::new(text);
    while let Some((token_type, token)) = tokens.next_token() {
        if token.len() > MAX_LEXEME {
            notices.push(long_word_notice());
            continue;
        }
        let Some(dictionary) = config.dictionaries(token_type).first() else { continue };
        *position += 1;
        for lexeme in dictionary.lexize(token) {
            if lexeme.len() > MAX_LEXEME {
                notices.push(long_word_notice());
                continue;
            }
            let position = (*position).min(u32::from(MAX_POSITION)) as u16;
            words.push(Word { text: lexeme, position });
        }
    }
    words
}

/// vector_lexemes merges parsed words into a vector's lexemes, as Postgres' make_tsvector and uniqueWORD do.
pub(crate) fn vector_lexemes(mut words: Vec<Word>) -> Vec<Lexeme> {
    words.sort_by(|a, b| a.text.as_bytes().cmp(b.text.as_bytes()).then(a.position.cmp(&b.position)));
    let mut lexemes: Vec<Lexeme> = Vec::new();
    for word in words {
        match lexemes.last_mut() {
            Some((text, positions)) if text.as_slice() == word.text.as_bytes() => {
                let last = *positions.last().unwrap_or(&0);
                if positions.len() < MAX_POSITIONS && last != MAX_POSITION && last != word.position {
                    positions.push(word.position);
                }
            }
            _ => lexemes.push((word.text.into_bytes(), vec![word.position])),
        }
    }
    lexemes
}

/// to_tsquery parses a query's text in a mode, passing each operand through a configuration as Postgres'
/// pushval_morph does, which joins the lexemes of an operand's words with `join` and fills the places of stop words
/// between them.
pub(crate) fn to_tsquery(
    config: Config,
    text: &str,
    mode: Mode,
    join: Operator,
) -> Result<(Option<Query>, Vec<PgError>)> {
    let mut notices = Vec::new();
    let mut morph = |builder: &mut Builder, operand: &[u8], weights: u8, prefix: bool| -> Result<()> {
        let words = parse_text(config, &String::from_utf8_lossy(operand), &mut 0, &mut notices);
        if words.is_empty() {
            builder.stop();
        }
        let mut previous = 0;
        for (i, word) in words.iter().enumerate() {
            while previous > 0 && previous + 1 < word.position {
                builder.stop();
                builder.operator(join, 1)?;
                previous += 1;
            }
            previous = word.position;
            builder.value(word.text.as_bytes(), weights, prefix)?;
            if i > 0 {
                builder.operator(join, 1)?;
            }
        }
        Ok(())
    };
    let parsed = parse_query(text, mode, &mut morph)?;
    notices.extend(parsed.notice.map(|notice| PgError::notice("00000", notice)));
    Ok((parsed.query, notices))
}
