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

//! The character set encodings that Postgres converts text to and from, with Postgres' names and numbers.

use crate::error::{PgError, Result, code};

/// NAMES are the canonical names of the encodings, by Postgres' number.
const NAMES: [&str; 42] = [
    "SQL_ASCII",
    "EUC_JP",
    "EUC_CN",
    "EUC_KR",
    "EUC_TW",
    "EUC_JIS_2004",
    "UTF8",
    "MULE_INTERNAL",
    "LATIN1",
    "LATIN2",
    "LATIN3",
    "LATIN4",
    "LATIN5",
    "LATIN6",
    "LATIN7",
    "LATIN8",
    "LATIN9",
    "LATIN10",
    "WIN1256",
    "WIN1258",
    "WIN866",
    "WIN874",
    "KOI8R",
    "WIN1251",
    "WIN1252",
    "ISO_8859_5",
    "ISO_8859_6",
    "ISO_8859_7",
    "ISO_8859_8",
    "WIN1250",
    "WIN1253",
    "WIN1254",
    "WIN1255",
    "WIN1257",
    "KOI8U",
    "SJIS",
    "BIG5",
    "GBK",
    "UHC",
    "GB18030",
    "JOHAB",
    "SHIFT_JIS_2004",
];

/// ALIASES are the other names that Postgres accepts, lowercase without punctuation, with their encodings' numbers.
const ALIASES: &[(&str, usize)] = &[
    ("abc", 19),
    ("alt", 20),
    ("iso88591", 8),
    ("iso88592", 9),
    ("iso88593", 10),
    ("iso88594", 11),
    ("iso88599", 12),
    ("iso885910", 13),
    ("iso885913", 14),
    ("iso885914", 15),
    ("iso885915", 16),
    ("iso885916", 17),
    ("koi8", 22),
    ("mskanji", 35),
    ("shiftjis", 35),
    ("tcvn", 19),
    ("tcvn5712", 19),
    ("unicode", 6),
    ("vscii", 19),
    ("win", 23),
    ("win932", 35),
    ("win936", 37),
    ("win949", 38),
    ("win950", 36),
    ("windows866", 20),
    ("windows874", 21),
    ("windows932", 35),
    ("windows936", 37),
    ("windows949", 38),
    ("windows950", 36),
];

/// Encoding is a character set encoding, by Postgres' number.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct Encoding(usize);

/// UTF8 is the server's encoding.
pub const UTF8: Encoding = Encoding(6);

impl Encoding {
    /// lookup returns the encoding of a name, ignoring case and punctuation as Postgres does.
    pub fn lookup(name: &str) -> Option<Encoding> {
        let clean: String = name.chars().filter(char::is_ascii_alphanumeric).collect::<String>().to_ascii_lowercase();
        if let Some(number) = NAMES.iter().position(|n| n.replace('_', "").eq_ignore_ascii_case(&clean)) {
            return Some(Encoding(number));
        }
        if let Some(rest) = clean.strip_prefix("windows")
            && let Some(number) = NAMES.iter().position(|n| n.eq_ignore_ascii_case(&format!("win{rest}")))
        {
            return Some(Encoding(number));
        }
        ALIASES.iter().find(|(alias, _)| *alias == clean).map(|(_, number)| Encoding(*number))
    }

    /// from_number returns the encoding of Postgres' number.
    pub fn from_number(number: i32) -> Option<Encoding> {
        usize::try_from(number).ok().filter(|&n| n < NAMES.len()).map(Encoding)
    }

    /// number returns Postgres' number of the encoding.
    pub fn number(self) -> i32 {
        self.0 as i32
    }

    /// name returns the encoding's canonical name.
    pub fn name(self) -> &'static str {
        NAMES[self.0]
    }

    /// codec returns how text converts to and from the encoding.
    fn codec(self) -> Option<Codec> {
        use encoding_rs as rs;
        Some(Codec::Table(match self.name() {
            "SQL_ASCII" | "UTF8" => return Some(Codec::Utf8),
            "LATIN1" => return Some(Codec::Latin1),
            "EUC_JP" => rs::EUC_JP,
            "EUC_CN" | "GBK" => rs::GBK,
            "EUC_KR" | "UHC" => rs::EUC_KR,
            "LATIN2" => rs::ISO_8859_2,
            "LATIN3" => rs::ISO_8859_3,
            "LATIN4" => rs::ISO_8859_4,
            "LATIN5" | "WIN1254" => rs::WINDOWS_1254,
            "LATIN6" => rs::ISO_8859_10,
            "LATIN7" => rs::ISO_8859_13,
            "LATIN8" => rs::ISO_8859_14,
            "LATIN9" => rs::ISO_8859_15,
            "LATIN10" => rs::ISO_8859_16,
            "WIN1256" => rs::WINDOWS_1256,
            "WIN1258" => rs::WINDOWS_1258,
            "WIN866" => rs::IBM866,
            "WIN874" => rs::WINDOWS_874,
            "KOI8R" => rs::KOI8_R,
            "KOI8U" => rs::KOI8_U,
            "WIN1250" => rs::WINDOWS_1250,
            "WIN1251" => rs::WINDOWS_1251,
            "WIN1252" => rs::WINDOWS_1252,
            "WIN1253" => rs::WINDOWS_1253,
            "WIN1255" => rs::WINDOWS_1255,
            "WIN1257" => rs::WINDOWS_1257,
            "ISO_8859_5" => rs::ISO_8859_5,
            "ISO_8859_6" => rs::ISO_8859_6,
            "ISO_8859_7" => rs::ISO_8859_7,
            "ISO_8859_8" => rs::ISO_8859_8,
            "SJIS" => rs::SHIFT_JIS,
            "BIG5" => rs::BIG5,
            "GB18030" => rs::GB18030,
            _ => return None,
        }))
    }

    /// unsupported returns the error for an encoding that Doltgres cannot convert text to or from.
    fn unsupported(self) -> PgError {
        PgError::unsupported(format!("conversions between UTF8 and {}", self.name()))
    }

    /// encode converts text to the encoding, failing for a character the encoding lacks.
    pub fn encode(self, text: &str) -> Result<Vec<u8>> {
        let unmappable = |c: char| {
            let mut buffer = [0; 4];
            let bytes: Vec<String> = c.encode_utf8(&mut buffer).bytes().map(|b| format!("0x{b:02x}")).collect();
            PgError::new(
                code::UNTRANSLATABLE_CHARACTER,
                format!(
                    "character with byte sequence {} in encoding \"UTF8\" has no equivalent in encoding \"{}\"",
                    bytes.join(" "),
                    self.name()
                ),
            )
        };
        match self.codec().ok_or_else(|| self.unsupported())? {
            Codec::Utf8 => Ok(text.as_bytes().to_vec()),
            Codec::Latin1 => text.chars().map(|c| u8::try_from(c as u32).map_err(|_| unmappable(c))).collect(),
            Codec::Table(table) => {
                let (bytes, _, errors) = table.encode(text);
                if errors {
                    let bad = text.chars().find(|c| table.encode(&c.to_string()).2).unwrap_or('\u{fffd}');
                    return Err(unmappable(bad));
                }
                Ok(bytes.into_owned())
            }
        }
    }

    /// decode converts bytes in the encoding to text, failing for bytes that are not a character of the encoding.
    pub fn decode(self, bytes: &[u8]) -> Result<String> {
        let invalid = |at: usize| {
            PgError::new(
                code::CHARACTER_NOT_IN_REPERTOIRE,
                format!(
                    "invalid byte sequence for encoding \"{}\": 0x{:02x}",
                    self.name(),
                    bytes.get(at).copied().unwrap_or(0)
                ),
            )
        };
        match self.codec().ok_or_else(|| self.unsupported())? {
            Codec::Utf8 => String::from_utf8(bytes.to_vec()).map_err(|e| invalid(e.utf8_error().valid_up_to())),
            Codec::Latin1 => Ok(bytes.iter().map(|&b| b as char).collect()),
            Codec::Table(table) => match table.decode_without_bom_handling_and_without_replacement(bytes) {
                Some(text) => Ok(text.into_owned()),
                None => Err(invalid(0)),
            },
        }
    }
}

/// Codec is how text converts to and from an encoding.
enum Codec {
    /// The bytes are the UTF-8 text itself.
    Utf8,
    /// Each byte is the character of the same code point.
    Latin1,
    /// encoding_rs converts the text.
    Table(&'static encoding_rs::Encoding),
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn encodings_convert_as_postgres_does() {
        let lookup = |name: &str| Encoding::lookup(name).map(Encoding::number);
        assert_eq!(lookup("WINDOWS-1252"), Some(24));
        assert_eq!(lookup("utf-8"), Some(6));
        assert_eq!(lookup("ISO_8859_1"), Some(8));
        assert_eq!(lookup("Shift_JIS"), Some(35));
        assert_eq!(lookup("bogus"), None);
        let win1252 = Encoding::lookup("WIN1252").unwrap();
        assert_eq!(win1252.encode("€").unwrap(), vec![0x80]);
        assert_eq!(win1252.decode(&[0x80]).unwrap(), "€");
        let latin1 = Encoding::lookup("LATIN1").unwrap();
        assert_eq!(latin1.encode("café").unwrap(), b"caf\xe9".to_vec());
        assert_eq!(latin1.encode("€").unwrap_err().code, "22P05");
    }
}
