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

//! The default text search parser, a port of Postgres' wparser_def.c state machine. Characters are classified as in
//! the C locale of a UTF-8 database, where every non-ASCII character is a letter, so the rows of Postgres' tables that
//! test for special characters never apply and are left out.

/// The token types of the default parser, numbered as Postgres numbers them.
pub const ASCII_WORD: u8 = 1;
pub const WORD: u8 = 2;
pub const NUM_WORD: u8 = 3;
pub const EMAIL: u8 = 4;
pub const URL: u8 = 5;
pub const HOST: u8 = 6;
pub const SCIENTIFIC: u8 = 7;
pub const VERSION: u8 = 8;
pub const NUM_PART_HWORD: u8 = 9;
pub const PART_HWORD: u8 = 10;
pub const ASCII_PART_HWORD: u8 = 11;
pub const SPACE: u8 = 12;
pub const TAG: u8 = 13;
pub const PROTOCOL: u8 = 14;
pub const NUM_HWORD: u8 = 15;
pub const ASCII_HWORD: u8 = 16;
pub const HWORD: u8 = 17;
pub const URL_PATH: u8 = 18;
pub const FILE: u8 = 19;
pub const DECIMAL: u8 = 20;
pub const SIGNED_INT: u8 = 21;
pub const UNSIGNED_INT: u8 = 22;
pub const XML_ENTITY: u8 = 23;

/// TOKEN_TYPES are the aliases and descriptions of the token types, in order of their numbers from 1.
pub const TOKEN_TYPES: &[(&str, &str)] = &[
    ("asciiword", "Word, all ASCII"),
    ("word", "Word, all letters"),
    ("numword", "Word, letters and digits"),
    ("email", "Email address"),
    ("url", "URL"),
    ("host", "Host"),
    ("sfloat", "Scientific notation"),
    ("version", "Version number"),
    ("hword_numpart", "Hyphenated word part, letters and digits"),
    ("hword_part", "Hyphenated word part, all letters"),
    ("hword_asciipart", "Hyphenated word part, all ASCII"),
    ("blank", "Space symbols"),
    ("tag", "XML tag"),
    ("protocol", "Protocol head"),
    ("numhword", "Hyphenated word, letters and digits"),
    ("asciihword", "Hyphenated word, all ASCII"),
    ("hword", "Hyphenated word, all letters"),
    ("url_path", "URL path"),
    ("file", "File or path name"),
    ("float", "Decimal notation"),
    ("int", "Signed integer"),
    ("uint", "Unsigned integer"),
    ("entity", "XML entity"),
];

/// The flags of an action, as Postgres names them: BINGO ends a token, POP returns to the state saved by the last
/// PUSH, RERUN tests the same character again, CLEAR drops the last saved state, MERGE takes the position into the
/// saved state, and CLRALL drops every saved state.
const BINGO: u8 = 0x01;
const POP: u8 = 0x02;
const PUSH: u8 = 0x04;
const RERUN: u8 = 0x08;
const CLEAR: u8 = 0x10;
const MERGE: u8 = 0x20;
const CLRALL: u8 = 0x40;

/// Test is what an action checks of the current character, as Postgres' p_is functions do.
#[derive(Clone, Copy)]
enum Test {
    Any,
    Eof,
    Eq(u8),
    Ignore,
    AsciiLetter,
    Alpha,
    Digit,
    Alnum,
    NotAlnum,
    Space,
    HexDigit,
    UrlChar,
    StopHost,
    Host,
    UrlPath,
}

/// Special is what an action does to the parser before its flags apply, as Postgres' Special functions do.
#[derive(Clone, Copy)]
enum Special {
    Tags,
    FullUrl,
    Hyphen,
    VerVersion,
}

/// Action is one row of a state's table: a test, the flags to apply when it passes, the state to move to, the type of
/// the token it ends, and its special handling.
struct Action {
    test: Test,
    flags: u8,
    to: Option<State>,
    token: u8,
    special: Option<Special>,
}

/// a builds an action.
const fn a(test: Test, flags: u8, to: Option<State>, token: u8, special: Option<Special>) -> Action {
    Action { test, flags, to, token, special }
}

/// State is a state of the parser, as Postgres' TParserState names them.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum State {
    Base,
    InNumWord,
    InAsciiWord,
    InWord,
    InUnsignedInt,
    InSignedIntFirst,
    InSignedInt,
    InSpace,
    InUDecimalFirst,
    InUDecimal,
    InDecimalFirst,
    InDecimal,
    InVerVersion,
    InSVerVersion,
    InVersionFirst,
    InVersion,
    InMantissaFirst,
    InMantissaSign,
    InMantissa,
    InXMLEntityFirst,
    InXMLEntity,
    InXMLEntityNumFirst,
    InXMLEntityNum,
    InXMLEntityHexNumFirst,
    InXMLEntityHexNum,
    InXMLEntityEnd,
    InTagFirst,
    InXMLBegin,
    InTagCloseFirst,
    InTagName,
    InTagBeginEnd,
    InTag,
    InTagEscapeK,
    InTagEscapeKK,
    InTagBackSleshed,
    InTagEnd,
    InCommentFirst,
    InCommentLast,
    InComment,
    InCloseCommentFirst,
    InCloseCommentLast,
    InCommentEnd,
    InHostFirstDomain,
    InHostDomainSecond,
    InHostDomain,
    InPortFirst,
    InPort,
    InHostFirstAN,
    InHost,
    InEmail,
    InFileFirst,
    InFileTwiddle,
    InPathFirst,
    InPathFirstFirst,
    InPathSecond,
    InFile,
    InFileNext,
    InURLPathFirst,
    InURLPathStart,
    InURLPath,
    InFURL,
    InProtocolFirst,
    InProtocolSecond,
    InProtocolEnd,
    InHyphenAsciiWordFirst,
    InHyphenAsciiWord,
    InHyphenWordFirst,
    InHyphenWord,
    InHyphenNumWordFirst,
    InHyphenNumWord,
    InHyphenDigitLookahead,
    InParseHyphen,
    InParseHyphenHyphen,
    InHyphenWordPart,
    InHyphenAsciiWordPart,
    InHyphenNumWordPart,
    InHyphenUnsignedInt,
}

/// actions returns the actions of a state, which the parser tries in order until one's test passes.
fn actions(state: State) -> &'static [Action] {
    match state {
        State::Base => {
            const {
                &[
                    a(Test::Eof, 0, None, 0, None),
                    a(Test::Eq(b'<'), PUSH, Some(State::InTagFirst), 0, None),
                    a(Test::Ignore, 0, Some(State::InSpace), 0, None),
                    a(Test::AsciiLetter, 0, Some(State::InAsciiWord), 0, None),
                    a(Test::Alpha, 0, Some(State::InWord), 0, None),
                    a(Test::Digit, 0, Some(State::InUnsignedInt), 0, None),
                    a(Test::Eq(b'-'), PUSH, Some(State::InSignedIntFirst), 0, None),
                    a(Test::Eq(b'+'), PUSH, Some(State::InSignedIntFirst), 0, None),
                    a(Test::Eq(b'&'), PUSH, Some(State::InXMLEntityFirst), 0, None),
                    a(Test::Eq(b'~'), PUSH, Some(State::InFileTwiddle), 0, None),
                    a(Test::Eq(b'/'), PUSH, Some(State::InFileFirst), 0, None),
                    a(Test::Eq(b'.'), PUSH, Some(State::InPathFirstFirst), 0, None),
                    a(Test::Any, 0, Some(State::InSpace), 0, None),
                ]
            }
        }
        State::InNumWord => {
            const {
                &[
                    a(Test::Eof, BINGO, Some(State::Base), NUM_WORD, None),
                    a(Test::Alnum, 0, Some(State::InNumWord), 0, None),
                    a(Test::Eq(b'@'), PUSH, Some(State::InEmail), 0, None),
                    a(Test::Eq(b'/'), PUSH, Some(State::InFileFirst), 0, None),
                    a(Test::Eq(b'.'), PUSH, Some(State::InFileNext), 0, None),
                    a(Test::Eq(b'-'), PUSH, Some(State::InHyphenNumWordFirst), 0, None),
                    a(Test::Any, BINGO, Some(State::Base), NUM_WORD, None),
                ]
            }
        }
        State::InAsciiWord => {
            const {
                &[
                    a(Test::Eof, BINGO, Some(State::Base), ASCII_WORD, None),
                    a(Test::AsciiLetter, 0, None, 0, None),
                    a(Test::Eq(b'.'), PUSH, Some(State::InHostFirstDomain), 0, None),
                    a(Test::Eq(b'.'), PUSH, Some(State::InFileNext), 0, None),
                    a(Test::Eq(b'-'), PUSH, Some(State::InHostFirstAN), 0, None),
                    a(Test::Eq(b'-'), PUSH, Some(State::InHyphenAsciiWordFirst), 0, None),
                    a(Test::Eq(b'_'), PUSH, Some(State::InHostFirstAN), 0, None),
                    a(Test::Eq(b'@'), PUSH, Some(State::InEmail), 0, None),
                    a(Test::Eq(b':'), PUSH, Some(State::InProtocolFirst), 0, None),
                    a(Test::Eq(b'/'), PUSH, Some(State::InFileFirst), 0, None),
                    a(Test::Digit, PUSH, Some(State::InHost), 0, None),
                    a(Test::Digit, 0, Some(State::InNumWord), 0, None),
                    a(Test::Alpha, 0, Some(State::InWord), 0, None),
                    a(Test::Any, BINGO, Some(State::Base), ASCII_WORD, None),
                ]
            }
        }
        State::InWord => {
            const {
                &[
                    a(Test::Eof, BINGO, Some(State::Base), WORD, None),
                    a(Test::Alpha, 0, None, 0, None),
                    a(Test::Digit, 0, Some(State::InNumWord), 0, None),
                    a(Test::Eq(b'-'), PUSH, Some(State::InHyphenWordFirst), 0, None),
                    a(Test::Any, BINGO, Some(State::Base), WORD, None),
                ]
            }
        }
        State::InUnsignedInt => {
            const {
                &[
                    a(Test::Eof, BINGO, Some(State::Base), UNSIGNED_INT, None),
                    a(Test::Digit, 0, None, 0, None),
                    a(Test::Eq(b'.'), PUSH, Some(State::InHostFirstDomain), 0, None),
                    a(Test::Eq(b'.'), PUSH, Some(State::InUDecimalFirst), 0, None),
                    a(Test::Eq(b'e'), PUSH, Some(State::InMantissaFirst), 0, None),
                    a(Test::Eq(b'E'), PUSH, Some(State::InMantissaFirst), 0, None),
                    a(Test::Eq(b'-'), PUSH, Some(State::InHostFirstAN), 0, None),
                    a(Test::Eq(b'_'), PUSH, Some(State::InHostFirstAN), 0, None),
                    a(Test::Eq(b'@'), PUSH, Some(State::InEmail), 0, None),
                    a(Test::AsciiLetter, PUSH, Some(State::InHost), 0, None),
                    a(Test::Alpha, 0, Some(State::InNumWord), 0, None),
                    a(Test::Eq(b'/'), PUSH, Some(State::InFileFirst), 0, None),
                    a(Test::Any, BINGO, Some(State::Base), UNSIGNED_INT, None),
                ]
            }
        }
        State::InSignedIntFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Digit, CLEAR, Some(State::InSignedInt), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InSignedInt => {
            const {
                &[
                    a(Test::Eof, BINGO, Some(State::Base), SIGNED_INT, None),
                    a(Test::Digit, 0, None, 0, None),
                    a(Test::Eq(b'.'), PUSH, Some(State::InDecimalFirst), 0, None),
                    a(Test::Eq(b'e'), PUSH, Some(State::InMantissaFirst), 0, None),
                    a(Test::Eq(b'E'), PUSH, Some(State::InMantissaFirst), 0, None),
                    a(Test::Any, BINGO, Some(State::Base), SIGNED_INT, None),
                ]
            }
        }
        State::InSpace => {
            const {
                &[
                    a(Test::Eof, BINGO, Some(State::Base), SPACE, None),
                    a(Test::Eq(b'<'), BINGO, Some(State::Base), SPACE, None),
                    a(Test::Ignore, 0, None, 0, None),
                    a(Test::Eq(b'-'), BINGO, Some(State::Base), SPACE, None),
                    a(Test::Eq(b'+'), BINGO, Some(State::Base), SPACE, None),
                    a(Test::Eq(b'&'), BINGO, Some(State::Base), SPACE, None),
                    a(Test::Eq(b'/'), BINGO, Some(State::Base), SPACE, None),
                    a(Test::NotAlnum, 0, Some(State::InSpace), 0, None),
                    a(Test::Any, BINGO, Some(State::Base), SPACE, None),
                ]
            }
        }
        State::InUDecimalFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Digit, CLEAR, Some(State::InUDecimal), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InUDecimal => {
            const {
                &[
                    a(Test::Eof, BINGO, Some(State::Base), DECIMAL, None),
                    a(Test::Digit, 0, Some(State::InUDecimal), 0, None),
                    a(Test::Eq(b'.'), PUSH, Some(State::InVersionFirst), 0, None),
                    a(Test::Eq(b'e'), PUSH, Some(State::InMantissaFirst), 0, None),
                    a(Test::Eq(b'E'), PUSH, Some(State::InMantissaFirst), 0, None),
                    a(Test::Any, BINGO, Some(State::Base), DECIMAL, None),
                ]
            }
        }
        State::InDecimalFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Digit, CLEAR, Some(State::InDecimal), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InDecimal => {
            const {
                &[
                    a(Test::Eof, BINGO, Some(State::Base), DECIMAL, None),
                    a(Test::Digit, 0, Some(State::InDecimal), 0, None),
                    a(Test::Eq(b'.'), PUSH, Some(State::InVerVersion), 0, None),
                    a(Test::Eq(b'e'), PUSH, Some(State::InMantissaFirst), 0, None),
                    a(Test::Eq(b'E'), PUSH, Some(State::InMantissaFirst), 0, None),
                    a(Test::Any, BINGO, Some(State::Base), DECIMAL, None),
                ]
            }
        }
        State::InVerVersion => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Digit, RERUN, Some(State::InSVerVersion), 0, Some(Special::VerVersion)),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InSVerVersion => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Digit, BINGO | CLRALL, Some(State::InUnsignedInt), SPACE, None),
                    a(Test::Any, 0, None, 0, None),
                ]
            }
        }
        State::InVersionFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Digit, CLEAR, Some(State::InVersion), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InVersion => {
            const {
                &[
                    a(Test::Eof, BINGO, Some(State::Base), VERSION, None),
                    a(Test::Digit, 0, Some(State::InVersion), 0, None),
                    a(Test::Eq(b'.'), PUSH, Some(State::InVersionFirst), 0, None),
                    a(Test::Any, BINGO, Some(State::Base), VERSION, None),
                ]
            }
        }
        State::InMantissaFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Digit, CLEAR, Some(State::InMantissa), 0, None),
                    a(Test::Eq(b'+'), 0, Some(State::InMantissaSign), 0, None),
                    a(Test::Eq(b'-'), 0, Some(State::InMantissaSign), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InMantissaSign => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Digit, CLEAR, Some(State::InMantissa), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InMantissa => {
            const {
                &[
                    a(Test::Eof, BINGO, Some(State::Base), SCIENTIFIC, None),
                    a(Test::Digit, 0, Some(State::InMantissa), 0, None),
                    a(Test::Any, BINGO, Some(State::Base), SCIENTIFIC, None),
                ]
            }
        }
        State::InXMLEntityFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Eq(b'#'), 0, Some(State::InXMLEntityNumFirst), 0, None),
                    a(Test::AsciiLetter, 0, Some(State::InXMLEntity), 0, None),
                    a(Test::Eq(b':'), 0, Some(State::InXMLEntity), 0, None),
                    a(Test::Eq(b'_'), 0, Some(State::InXMLEntity), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InXMLEntity => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Alnum, 0, Some(State::InXMLEntity), 0, None),
                    a(Test::Eq(b':'), 0, Some(State::InXMLEntity), 0, None),
                    a(Test::Eq(b'_'), 0, Some(State::InXMLEntity), 0, None),
                    a(Test::Eq(b'.'), 0, Some(State::InXMLEntity), 0, None),
                    a(Test::Eq(b'-'), 0, Some(State::InXMLEntity), 0, None),
                    a(Test::Eq(b';'), 0, Some(State::InXMLEntityEnd), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InXMLEntityNumFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Eq(b'x'), 0, Some(State::InXMLEntityHexNumFirst), 0, None),
                    a(Test::Eq(b'X'), 0, Some(State::InXMLEntityHexNumFirst), 0, None),
                    a(Test::Digit, 0, Some(State::InXMLEntityNum), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InXMLEntityNum => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Digit, 0, Some(State::InXMLEntityNum), 0, None),
                    a(Test::Eq(b';'), 0, Some(State::InXMLEntityEnd), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InXMLEntityHexNumFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::HexDigit, 0, Some(State::InXMLEntityHexNum), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InXMLEntityHexNum => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::HexDigit, 0, Some(State::InXMLEntityHexNum), 0, None),
                    a(Test::Eq(b';'), 0, Some(State::InXMLEntityEnd), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InXMLEntityEnd => const { &[a(Test::Any, BINGO | CLEAR, Some(State::Base), XML_ENTITY, None)] },
        State::InTagFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Eq(b'/'), PUSH, Some(State::InTagCloseFirst), 0, None),
                    a(Test::Eq(b'!'), PUSH, Some(State::InCommentFirst), 0, None),
                    a(Test::Eq(b'?'), PUSH, Some(State::InXMLBegin), 0, None),
                    a(Test::AsciiLetter, PUSH, Some(State::InTagName), 0, None),
                    a(Test::Eq(b':'), PUSH, Some(State::InTagName), 0, None),
                    a(Test::Eq(b'_'), PUSH, Some(State::InTagName), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InXMLBegin => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Eq(b'x'), 0, Some(State::InTag), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InTagCloseFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::AsciiLetter, 0, Some(State::InTagName), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InTagName => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Eq(b'/'), 0, Some(State::InTagBeginEnd), 0, None),
                    a(Test::Eq(b'>'), 0, Some(State::InTagEnd), 0, Some(Special::Tags)),
                    a(Test::Space, 0, Some(State::InTag), 0, Some(Special::Tags)),
                    a(Test::Alnum, 0, None, 0, None),
                    a(Test::Eq(b':'), 0, None, 0, None),
                    a(Test::Eq(b'_'), 0, None, 0, None),
                    a(Test::Eq(b'.'), 0, None, 0, None),
                    a(Test::Eq(b'-'), 0, None, 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InTagBeginEnd => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Eq(b'>'), 0, Some(State::InTagEnd), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InTag => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Eq(b'>'), 0, Some(State::InTagEnd), 0, Some(Special::Tags)),
                    a(Test::Eq(b'\''), 0, Some(State::InTagEscapeK), 0, None),
                    a(Test::Eq(b'"'), 0, Some(State::InTagEscapeKK), 0, None),
                    a(Test::AsciiLetter, 0, None, 0, None),
                    a(Test::Digit, 0, None, 0, None),
                    a(Test::Eq(b'='), 0, None, 0, None),
                    a(Test::Eq(b'-'), 0, None, 0, None),
                    a(Test::Eq(b'_'), 0, None, 0, None),
                    a(Test::Eq(b'#'), 0, None, 0, None),
                    a(Test::Eq(b'/'), 0, None, 0, None),
                    a(Test::Eq(b':'), 0, None, 0, None),
                    a(Test::Eq(b'.'), 0, None, 0, None),
                    a(Test::Eq(b'&'), 0, None, 0, None),
                    a(Test::Eq(b'?'), 0, None, 0, None),
                    a(Test::Eq(b'%'), 0, None, 0, None),
                    a(Test::Eq(b'~'), 0, None, 0, None),
                    a(Test::Space, 0, None, 0, Some(Special::Tags)),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InTagEscapeK => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Eq(b'\\'), PUSH, Some(State::InTagBackSleshed), 0, None),
                    a(Test::Eq(b'\''), 0, Some(State::InTag), 0, None),
                    a(Test::Any, 0, Some(State::InTagEscapeK), 0, None),
                ]
            }
        }
        State::InTagEscapeKK => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Eq(b'\\'), PUSH, Some(State::InTagBackSleshed), 0, None),
                    a(Test::Eq(b'"'), 0, Some(State::InTag), 0, None),
                    a(Test::Any, 0, Some(State::InTagEscapeKK), 0, None),
                ]
            }
        }
        State::InTagBackSleshed => const { &[a(Test::Eof, POP, None, 0, None), a(Test::Any, MERGE, None, 0, None)] },
        State::InTagEnd => const { &[a(Test::Any, BINGO | CLRALL, Some(State::Base), TAG, None)] },
        State::InCommentFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Eq(b'-'), 0, Some(State::InCommentLast), 0, None),
                    a(Test::Eq(b'D'), 0, Some(State::InTag), 0, None),
                    a(Test::Eq(b'd'), 0, Some(State::InTag), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InCommentLast => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Eq(b'-'), 0, Some(State::InComment), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InComment => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Eq(b'-'), 0, Some(State::InCloseCommentFirst), 0, None),
                    a(Test::Any, 0, None, 0, None),
                ]
            }
        }
        State::InCloseCommentFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Eq(b'-'), 0, Some(State::InCloseCommentLast), 0, None),
                    a(Test::Any, 0, Some(State::InComment), 0, None),
                ]
            }
        }
        State::InCloseCommentLast => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Eq(b'-'), 0, None, 0, None),
                    a(Test::Eq(b'>'), 0, Some(State::InCommentEnd), 0, None),
                    a(Test::Any, 0, Some(State::InComment), 0, None),
                ]
            }
        }
        State::InCommentEnd => const { &[a(Test::Any, BINGO | CLRALL, Some(State::Base), TAG, None)] },
        State::InHostFirstDomain => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::AsciiLetter, 0, Some(State::InHostDomainSecond), 0, None),
                    a(Test::Digit, 0, Some(State::InHost), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InHostDomainSecond => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::AsciiLetter, 0, Some(State::InHostDomain), 0, None),
                    a(Test::Digit, PUSH, Some(State::InHost), 0, None),
                    a(Test::Eq(b'-'), PUSH, Some(State::InHostFirstAN), 0, None),
                    a(Test::Eq(b'_'), PUSH, Some(State::InHostFirstAN), 0, None),
                    a(Test::Eq(b'.'), PUSH, Some(State::InHostFirstDomain), 0, None),
                    a(Test::Eq(b'@'), PUSH, Some(State::InEmail), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InHostDomain => {
            const {
                &[
                    a(Test::Eof, BINGO | CLRALL, Some(State::Base), HOST, None),
                    a(Test::AsciiLetter, 0, Some(State::InHostDomain), 0, None),
                    a(Test::Digit, PUSH, Some(State::InHost), 0, None),
                    a(Test::Eq(b':'), PUSH, Some(State::InPortFirst), 0, None),
                    a(Test::Eq(b'-'), PUSH, Some(State::InHostFirstAN), 0, None),
                    a(Test::Eq(b'_'), PUSH, Some(State::InHostFirstAN), 0, None),
                    a(Test::Eq(b'.'), PUSH, Some(State::InHostFirstDomain), 0, None),
                    a(Test::Eq(b'@'), PUSH, Some(State::InEmail), 0, None),
                    a(Test::Digit, POP, None, 0, None),
                    a(Test::StopHost, BINGO | CLRALL, Some(State::InURLPathStart), HOST, None),
                    a(Test::Eq(b'/'), PUSH, Some(State::InFURL), 0, None),
                    a(Test::Any, BINGO | CLRALL, Some(State::Base), HOST, None),
                ]
            }
        }
        State::InPortFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Digit, 0, Some(State::InPort), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InPort => {
            const {
                &[
                    a(Test::Eof, BINGO | CLRALL, Some(State::Base), HOST, None),
                    a(Test::Digit, 0, Some(State::InPort), 0, None),
                    a(Test::StopHost, BINGO | CLRALL, Some(State::InURLPathStart), HOST, None),
                    a(Test::Eq(b'/'), PUSH, Some(State::InFURL), 0, None),
                    a(Test::Any, BINGO | CLRALL, Some(State::Base), HOST, None),
                ]
            }
        }
        State::InHostFirstAN => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Digit, 0, Some(State::InHost), 0, None),
                    a(Test::AsciiLetter, 0, Some(State::InHost), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InHost => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Digit, 0, Some(State::InHost), 0, None),
                    a(Test::AsciiLetter, 0, Some(State::InHost), 0, None),
                    a(Test::Eq(b'@'), PUSH, Some(State::InEmail), 0, None),
                    a(Test::Eq(b'.'), PUSH, Some(State::InHostFirstDomain), 0, None),
                    a(Test::Eq(b'-'), PUSH, Some(State::InHostFirstAN), 0, None),
                    a(Test::Eq(b'_'), PUSH, Some(State::InHostFirstAN), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InEmail => {
            const {
                &[
                    a(Test::StopHost, POP, None, 0, None),
                    a(Test::Host, BINGO | CLRALL, Some(State::Base), EMAIL, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InFileFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::AsciiLetter, 0, Some(State::InFile), 0, None),
                    a(Test::Digit, 0, Some(State::InFile), 0, None),
                    a(Test::Eq(b'.'), 0, Some(State::InPathFirst), 0, None),
                    a(Test::Eq(b'_'), 0, Some(State::InFile), 0, None),
                    a(Test::Eq(b'~'), PUSH, Some(State::InFileTwiddle), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InFileTwiddle => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::AsciiLetter, 0, Some(State::InFile), 0, None),
                    a(Test::Digit, 0, Some(State::InFile), 0, None),
                    a(Test::Eq(b'_'), 0, Some(State::InFile), 0, None),
                    a(Test::Eq(b'/'), 0, Some(State::InFileFirst), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InPathFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::AsciiLetter, 0, Some(State::InFile), 0, None),
                    a(Test::Digit, 0, Some(State::InFile), 0, None),
                    a(Test::Eq(b'_'), 0, Some(State::InFile), 0, None),
                    a(Test::Eq(b'.'), 0, Some(State::InPathSecond), 0, None),
                    a(Test::Eq(b'/'), 0, Some(State::InFileFirst), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InPathFirstFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Eq(b'.'), 0, Some(State::InPathSecond), 0, None),
                    a(Test::Eq(b'/'), 0, Some(State::InFileFirst), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InPathSecond => {
            const {
                &[
                    a(Test::Eof, BINGO | CLEAR, Some(State::Base), FILE, None),
                    a(Test::Eq(b'/'), PUSH, Some(State::InFileFirst), 0, None),
                    a(Test::Eq(b'/'), BINGO | CLEAR, Some(State::Base), FILE, None),
                    a(Test::Space, BINGO | CLEAR, Some(State::Base), FILE, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InFile => {
            const {
                &[
                    a(Test::Eof, BINGO, Some(State::Base), FILE, None),
                    a(Test::AsciiLetter, 0, Some(State::InFile), 0, None),
                    a(Test::Digit, 0, Some(State::InFile), 0, None),
                    a(Test::Eq(b'.'), PUSH, Some(State::InFileNext), 0, None),
                    a(Test::Eq(b'_'), 0, Some(State::InFile), 0, None),
                    a(Test::Eq(b'-'), 0, Some(State::InFile), 0, None),
                    a(Test::Eq(b'/'), PUSH, Some(State::InFileFirst), 0, None),
                    a(Test::Any, BINGO, Some(State::Base), FILE, None),
                ]
            }
        }
        State::InFileNext => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::AsciiLetter, CLEAR, Some(State::InFile), 0, None),
                    a(Test::Digit, CLEAR, Some(State::InFile), 0, None),
                    a(Test::Eq(b'_'), CLEAR, Some(State::InFile), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InURLPathFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::UrlChar, 0, Some(State::InURLPath), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InURLPathStart => const { &[a(Test::Any, 0, Some(State::InURLPath), 0, None)] },
        State::InURLPath => {
            const {
                &[
                    a(Test::Eof, BINGO, Some(State::Base), URL_PATH, None),
                    a(Test::UrlChar, 0, Some(State::InURLPath), 0, None),
                    a(Test::Any, BINGO, Some(State::Base), URL_PATH, None),
                ]
            }
        }
        State::InFURL => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::UrlPath, BINGO | CLRALL, Some(State::Base), URL, Some(Special::FullUrl)),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InProtocolFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Eq(b'/'), 0, Some(State::InProtocolSecond), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InProtocolSecond => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Eq(b'/'), 0, Some(State::InProtocolEnd), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InProtocolEnd => const { &[a(Test::Any, BINGO | CLRALL, Some(State::Base), PROTOCOL, None)] },
        State::InHyphenAsciiWordFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::AsciiLetter, 0, Some(State::InHyphenAsciiWord), 0, None),
                    a(Test::Alpha, 0, Some(State::InHyphenWord), 0, None),
                    a(Test::Digit, 0, Some(State::InHyphenDigitLookahead), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InHyphenAsciiWord => {
            const {
                &[
                    a(Test::Eof, BINGO | CLRALL, Some(State::InParseHyphen), ASCII_HWORD, Some(Special::Hyphen)),
                    a(Test::AsciiLetter, 0, Some(State::InHyphenAsciiWord), 0, None),
                    a(Test::Alpha, 0, Some(State::InHyphenWord), 0, None),
                    a(Test::Digit, 0, Some(State::InHyphenNumWord), 0, None),
                    a(Test::Eq(b'-'), PUSH, Some(State::InHyphenAsciiWordFirst), 0, None),
                    a(Test::Any, BINGO | CLRALL, Some(State::InParseHyphen), ASCII_HWORD, Some(Special::Hyphen)),
                ]
            }
        }
        State::InHyphenWordFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Alpha, 0, Some(State::InHyphenWord), 0, None),
                    a(Test::Digit, 0, Some(State::InHyphenDigitLookahead), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InHyphenWord => {
            const {
                &[
                    a(Test::Eof, BINGO | CLRALL, Some(State::InParseHyphen), HWORD, Some(Special::Hyphen)),
                    a(Test::Alpha, 0, Some(State::InHyphenWord), 0, None),
                    a(Test::Digit, 0, Some(State::InHyphenNumWord), 0, None),
                    a(Test::Eq(b'-'), PUSH, Some(State::InHyphenWordFirst), 0, None),
                    a(Test::Any, BINGO | CLRALL, Some(State::InParseHyphen), HWORD, Some(Special::Hyphen)),
                ]
            }
        }
        State::InHyphenNumWordFirst => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Alpha, 0, Some(State::InHyphenNumWord), 0, None),
                    a(Test::Digit, 0, Some(State::InHyphenDigitLookahead), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InHyphenNumWord => {
            const {
                &[
                    a(Test::Eof, BINGO | CLRALL, Some(State::InParseHyphen), NUM_HWORD, Some(Special::Hyphen)),
                    a(Test::Alnum, 0, Some(State::InHyphenNumWord), 0, None),
                    a(Test::Eq(b'-'), PUSH, Some(State::InHyphenNumWordFirst), 0, None),
                    a(Test::Any, BINGO | CLRALL, Some(State::InParseHyphen), NUM_HWORD, Some(Special::Hyphen)),
                ]
            }
        }
        State::InHyphenDigitLookahead => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Digit, 0, Some(State::InHyphenDigitLookahead), 0, None),
                    a(Test::Alpha, 0, Some(State::InHyphenNumWord), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InParseHyphen => {
            const {
                &[
                    a(Test::Eof, RERUN, Some(State::Base), 0, None),
                    a(Test::AsciiLetter, 0, Some(State::InHyphenAsciiWordPart), 0, None),
                    a(Test::Alpha, 0, Some(State::InHyphenWordPart), 0, None),
                    a(Test::Digit, PUSH, Some(State::InHyphenUnsignedInt), 0, None),
                    a(Test::Eq(b'-'), PUSH, Some(State::InParseHyphenHyphen), 0, None),
                    a(Test::Any, RERUN, Some(State::Base), 0, None),
                ]
            }
        }
        State::InParseHyphenHyphen => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Alnum, BINGO | CLEAR, Some(State::InParseHyphen), SPACE, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
        State::InHyphenWordPart => {
            const {
                &[
                    a(Test::Eof, BINGO, Some(State::Base), PART_HWORD, None),
                    a(Test::Alpha, 0, Some(State::InHyphenWordPart), 0, None),
                    a(Test::Digit, 0, Some(State::InHyphenNumWordPart), 0, None),
                    a(Test::Any, BINGO, Some(State::InParseHyphen), PART_HWORD, None),
                ]
            }
        }
        State::InHyphenAsciiWordPart => {
            const {
                &[
                    a(Test::Eof, BINGO, Some(State::Base), ASCII_PART_HWORD, None),
                    a(Test::AsciiLetter, 0, Some(State::InHyphenAsciiWordPart), 0, None),
                    a(Test::Alpha, 0, Some(State::InHyphenWordPart), 0, None),
                    a(Test::Digit, 0, Some(State::InHyphenNumWordPart), 0, None),
                    a(Test::Any, BINGO, Some(State::InParseHyphen), ASCII_PART_HWORD, None),
                ]
            }
        }
        State::InHyphenNumWordPart => {
            const {
                &[
                    a(Test::Eof, BINGO, Some(State::Base), NUM_PART_HWORD, None),
                    a(Test::Alnum, 0, Some(State::InHyphenNumWordPart), 0, None),
                    a(Test::Any, BINGO, Some(State::InParseHyphen), NUM_PART_HWORD, None),
                ]
            }
        }
        State::InHyphenUnsignedInt => {
            const {
                &[
                    a(Test::Eof, POP, None, 0, None),
                    a(Test::Digit, 0, None, 0, None),
                    a(Test::Alpha, CLEAR, Some(State::InHyphenNumWordPart), 0, None),
                    a(Test::Any, POP, None, 0, None),
                ]
            }
        }
    }
}

/// Position is a saved state of the parser: where it is in bytes, the length of the current character, the length of
/// the token so far in bytes and characters, the state, and the action a PUSH saved it at.
#[derive(Clone)]
struct Position {
    byte: usize,
    char_len: usize,
    token_bytes: usize,
    token_chars: usize,
    state: State,
    pushed_at: Option<usize>,
}

/// Parser is Postgres' TParser over one text.
pub struct Parser<'t> {
    text: &'t str,
    stack: Vec<Position>,
    ignore: bool,
    want_host: bool,
    token_start: usize,
}

impl<'t> Parser<'t> {
    /// new returns a parser at the start of the text.
    pub fn new(text: &'t str) -> Parser<'t> {
        Parser::starting(text, State::Base)
    }

    /// starting returns a parser at the start of the text in a state, saved above the base state as Postgres'
    /// p_isURLPath does.
    fn starting(text: &'t str, state: State) -> Parser<'t> {
        let base =
            Position { byte: 0, char_len: 0, token_bytes: 0, token_chars: 0, state: State::Base, pushed_at: None };
        let mut stack = vec![base.clone()];
        if state != State::Base {
            stack.push(Position { state, ..base });
        }
        Parser { text, stack, ignore: false, want_host: false, token_start: 0 }
    }

    /// top returns the current state.
    fn top(&mut self) -> &mut Position {
        self.stack.last_mut().expect("a parser state")
    }

    /// current returns the current character.
    fn current(&self) -> Option<char> {
        let top = self.stack.last()?;
        self.text[top.byte..].chars().next().filter(|_| top.char_len > 0)
    }

    /// ascii returns the current character when it is ASCII.
    fn ascii(&self) -> Option<u8> {
        self.current().filter(char::is_ascii).map(|c| c as u8)
    }

    /// test runs an action's test on the current character.
    fn test(&mut self, test: Test) -> bool {
        let non_ascii = self.current().is_some_and(|c| !c.is_ascii());
        match test {
            Test::Any => true,
            Test::Eof => self.current().is_none(),
            Test::Eq(c) => self.ascii() == Some(c),
            Test::Ignore => self.ignore,
            Test::AsciiLetter => self.ascii().is_some_and(|c| c.is_ascii_alphabetic()),
            Test::Alpha => non_ascii || self.ascii().is_some_and(|c| c.is_ascii_alphabetic()),
            Test::Digit => self.ascii().is_some_and(|c| c.is_ascii_digit()),
            Test::Alnum => non_ascii || self.ascii().is_some_and(|c| c.is_ascii_alphanumeric()),
            Test::NotAlnum => !non_ascii && !self.ascii().is_some_and(|c| c.is_ascii_alphanumeric()),
            Test::Space => self.ascii().is_some_and(|c| matches!(c, b' ' | b'\t' | b'\n' | 0x0b | 0x0c | b'\r')),
            Test::HexDigit => self.ascii().is_some_and(|c| c.is_ascii_hexdigit()),
            Test::UrlChar => self.ascii().is_some_and(|c| c > 0x20 && c < 0x7f && !b"\"<>\\^`{|}".contains(&c)),
            Test::StopHost => std::mem::take(&mut self.want_host),
            Test::Host => self.take_subtoken(State::Base, HOST),
            Test::UrlPath => self.take_subtoken(State::InURLPathFirst, URL_PATH),
        }
    }

    /// take_subtoken parses the rest of the text from the current character with a new parser, as Postgres'
    /// p_ishost and p_isURLPath do, and takes its first token into the current one when it has the wanted type.
    fn take_subtoken(&mut self, state: State, wanted: u8) -> bool {
        let at = self.stack.last().map_or(0, |p| p.byte);
        let mut sub = Parser::starting(&self.text[at..], state);
        sub.want_host = wanted == HOST;
        let Some((kind, token)) = sub.next_token() else { return false };
        if kind != wanted {
            return false;
        }
        let (bytes, chars, char_len) = (token.len(), token.chars().count(), sub.stack.last().map_or(0, |p| p.char_len));
        let top = self.top();
        top.byte += bytes;
        top.token_bytes += bytes;
        top.token_chars += chars;
        top.char_len = char_len;
        true
    }

    /// special runs an action's special handling.
    fn special(&mut self, special: Special) {
        if let Special::Tags = special {
            let token = &self.text.as_bytes()[self.token_start..];
            let starts =
                |tag: &str| token.len() >= tag.len() && token[..tag.len()].eq_ignore_ascii_case(tag.as_bytes());
            match self.top().token_chars {
                8 if starts("</script") => self.ignore = false,
                7 if starts("</style") => self.ignore = false,
                7 if starts("<script") => self.ignore = true,
                6 if starts("<style") => self.ignore = true,
                _ => {}
            }
            return;
        }
        let top = self.top();
        top.byte -= top.token_bytes;
        if let Special::VerVersion = special {
            top.token_bytes = 0;
            top.token_chars = 0;
        }
        if let Special::FullUrl = special {
            self.want_host = true;
        }
    }

    /// next_token returns the type and text of the next token, as Postgres' TParserGet does, or None at the end.
    pub fn next_token(&mut self) -> Option<(u8, &'t str)> {
        let len = self.text.len();
        if self.top().byte >= len {
            return None;
        }
        self.token_start = self.top().byte;
        self.top().pushed_at = None;
        let mut found = None;
        loop {
            let byte = self.top().byte;
            self.top().char_len = self.text[byte..].chars().next().map_or(0, char::len_utf8);
            let table = actions(self.top().state);
            let mut index = match self.top().pushed_at.take() {
                Some(pushed) => pushed + 1,
                None => 0,
            };
            while !self.test(table[index].test) {
                index += 1;
            }
            let action = &table[index];
            if let Some(special) = action.special {
                self.special(special);
            }
            if action.flags & BINGO != 0 {
                let top = self.top();
                found = Some((action.token, top.token_bytes));
                top.token_bytes = 0;
                top.token_chars = 0;
            }
            if action.flags & POP != 0 {
                self.stack.pop();
            } else if action.flags & PUSH != 0 {
                self.top().pushed_at = Some(index);
                let saved = Position { pushed_at: None, ..self.top().clone() };
                self.stack.push(saved);
            } else if action.flags & CLEAR != 0 {
                let below = self.stack.len() - 2;
                self.stack.remove(below);
            } else if action.flags & CLRALL != 0 {
                self.stack.drain(..self.stack.len() - 1);
            } else if action.flags & MERGE != 0 {
                let merged = self.stack.pop().expect("a merged state");
                let top = self.top();
                top.byte = merged.byte;
                top.char_len = merged.char_len;
                top.token_bytes = merged.token_bytes;
                top.token_chars = merged.token_chars;
            }
            if let Some(to) = action.to {
                self.top().state = to;
            }
            if action.flags & BINGO != 0 || (self.top().byte >= len && action.flags & RERUN == 0) {
                break;
            }
            if action.flags & (RERUN | POP) != 0 {
                continue;
            }
            let top = self.top();
            if top.char_len > 0 {
                top.byte += top.char_len;
                top.token_bytes += top.char_len;
                top.token_chars += 1;
            }
        }
        let (kind, bytes) = found?;
        Some((kind, &self.text[self.token_start..self.token_start + bytes]))
    }
}
