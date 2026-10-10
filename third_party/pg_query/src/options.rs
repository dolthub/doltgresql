use std::ops::{BitOr, BitOrAssign};

/// Options that control how a query string is parsed, mirroring the parser
/// options of the C `pg_query_parse_opts`/`pg_query_fingerprint_opts` API.
///
/// The parse mode (which grammar to use) occupies the lower 4 bits, the
/// GUC-style flags occupy the bits above them, and both can be combined:
///
/// ```rust
/// let options = pg_query::ParserOptions::TYPE_NAME | pg_query::ParserOptions::DISABLE_BACKSLASH_QUOTE;
/// assert_eq!(options.bits(), 17);
/// ```
///
/// Functions taking parser options accept these constants or raw integers;
/// pass `0` when no specific options are needed:
///
/// ```rust
/// let result = pg_query::parse("SELECT * FROM contacts", 0).unwrap();
/// assert_eq!(result.tables(), vec!["contacts"]);
/// ```
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq, Hash)]
pub struct ParserOptions(i32);

impl ParserOptions {
    /// Parse using the standard top-level SQL grammar (the default)
    pub const DEFAULT: Self = Self(0);
    /// Parse a type name (e.g. `integer`, `character varying(32)`)
    pub const TYPE_NAME: Self = Self(1);
    /// Parse a PL/pgSQL expression
    pub const PLPGSQL_EXPR: Self = Self(2);
    /// Parse a PL/pgSQL assignment (target 1)
    pub const PLPGSQL_ASSIGN1: Self = Self(3);
    /// Parse a PL/pgSQL assignment (target 2)
    pub const PLPGSQL_ASSIGN2: Self = Self(4);
    /// Parse a PL/pgSQL assignment (target 3)
    pub const PLPGSQL_ASSIGN3: Self = Self(5);

    /// `backslash_quote = off` (default is `safe_encoding`, which is effectively on)
    pub const DISABLE_BACKSLASH_QUOTE: Self = Self(16);
    /// `standard_conforming_strings = off` (default is on)
    pub const DISABLE_STANDARD_CONFORMING_STRINGS: Self = Self(32);
    /// `escape_string_warning = off` (default is on)
    pub const DISABLE_ESCAPE_STRING_WARNING: Self = Self(64);

    /// The raw options value as passed to the C API
    pub fn bits(self) -> i32 {
        self.0
    }
}

impl BitOr for ParserOptions {
    type Output = Self;

    fn bitor(self, rhs: Self) -> Self {
        Self(self.0 | rhs.0)
    }
}

impl BitOrAssign for ParserOptions {
    fn bitor_assign(&mut self, rhs: Self) {
        self.0 |= rhs.0;
    }
}

// Allow passing raw parser option integers wherever `Into<ParserOptions>` is accepted
impl From<i32> for ParserOptions {
    fn from(bits: i32) -> Self {
        Self(bits)
    }
}

impl From<u32> for ParserOptions {
    fn from(bits: u32) -> Self {
        Self(bits as i32)
    }
}

/// Options that control how fingerprints are calculated, mirroring the
/// fingerprint options of the C `pg_query_fingerprint_opts` API.
///
/// Functions taking fingerprint options accept these constants or raw
/// integers; pass `0` when no specific options are needed. Flags can also be
/// combined:
///
/// ```rust
/// let options = pg_query::FingerprintOptions::RANGEVAR_IGNORE_ALIASES | pg_query::FingerprintOptions::FULL_RELNAME;
/// assert_eq!(options.bits(), 17);
/// ```
///
/// By default (matching Postgres 18 query ID behavior, see Postgres commit
/// 787514b30bb) relation references in SELECT/DML statements are
/// fingerprinted by their alias when one is present, and schema names are
/// ignored; additionally, sequences of two or more consecutive digits in
/// relation names are ignored (so queries on date/number-suffixed partitions
/// such as `orders_2024_01` and `orders_2024_02` share a fingerprint).
///
/// Pass [`FingerprintOptions::RANGEVAR_PG17_COMPAT`] to instead fingerprint
/// relation references the way Postgres 17 and earlier (and libpg_query 17
/// and earlier) did: schema names significant, aliases always ignored.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq, Hash)]
pub struct FingerprintOptions(i32);

impl FingerprintOptions {
    /// Fingerprint using the default Postgres 18+ behaviour: relation
    /// references follow the Postgres 18+ query ID behavior (see Postgres
    /// commit 787514b30bb), and sequences of two or more digits in relation
    /// names are ignored.
    pub const DEFAULT: Self = Self(0);

    /// Relation names are always fingerprinted, aliases are ignored
    pub const RANGEVAR_IGNORE_ALIASES: Self = Self(1 << 0);
    /// Schema names are also fingerprinted in SELECT/DML statements (they are
    /// always fingerprinted in utility statements)
    pub const RANGEVAR_INCLUDE_SCHEMA: Self = Self(1 << 1);
    /// Convenience combination that matches how Postgres 17 and earlier
    /// calculated query IDs, and how libpg_query 17 and earlier calculated
    /// fingerprints
    pub const RANGEVAR_PG17_COMPAT: Self = Self(Self::RANGEVAR_IGNORE_ALIASES.0 | Self::RANGEVAR_INCLUDE_SCHEMA.0);

    /// Fingerprint the full relation name, instead of ignoring sequences of
    /// two or more digits (so that e.g. partitions like `orders_2024_01` and
    /// `orders_2024_02` get different fingerprints)
    pub const FULL_RELNAME: Self = Self(1 << 4);

    /// The raw options value as passed to the C API
    pub fn bits(self) -> i32 {
        self.0
    }
}

impl BitOr for FingerprintOptions {
    type Output = Self;

    fn bitor(self, rhs: Self) -> Self {
        Self(self.0 | rhs.0)
    }
}

impl BitOrAssign for FingerprintOptions {
    fn bitor_assign(&mut self, rhs: Self) {
        self.0 |= rhs.0;
    }
}

// Allow passing raw fingerprint option integers wherever `Into<FingerprintOptions>` is accepted
impl From<i32> for FingerprintOptions {
    fn from(bits: i32) -> Self {
        Self(bits)
    }
}

impl From<u32> for FingerprintOptions {
    fn from(bits: u32) -> Self {
        Self(bits as i32)
    }
}

/// Options that control how a parse tree is deparsed, mirroring the C
/// `PostgresDeparseOpts` struct. Use struct update syntax for a partial
/// configuration:
///
/// ```rust
/// let result = pg_query::parse("SELECT a, b FROM t WHERE x = 1", 0).unwrap();
/// let pretty = pg_query::deparse(
///     &result.protobuf,
///     pg_query::DeparseOptions {
///         pretty_print: true,
///         indent_size: 2,
///         ..Default::default()
///     },
/// )
/// .unwrap();
/// assert!(pretty.contains('\n'));
/// ```
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct DeparseOptions {
    /// Comments to insert into the output, e.g. as returned by
    /// [`deparse_comments_for_query`](crate::deparse_comments_for_query)
    pub comments: Vec<DeparseComment>,
    /// Pretty-print the output with newlines and indentation
    pub pretty_print: bool,
    /// Indentation size when pretty printing (0 lets the C library use its default of 4)
    pub indent_size: i32,
    /// Restricts the line length of certain lists of items when pretty
    /// printing (0 lets the C library use its default of 80)
    pub max_line_length: i32,
    /// Add a trailing newline at the end of the output (default off)
    pub trailing_newline: bool,
    /// Place separating commas at the start of the line when pretty printing (default off)
    pub commas_start_of_line: bool,
}

/// A comment to re-insert into deparsed output, mirroring the C
/// `PostgresDeparseComment` struct. Comments are produced by
/// [`deparse_comments_for_query`](crate::deparse_comments_for_query) and fed
/// back into [`deparse`](crate::deparse) via [`DeparseOptions::comments`].
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct DeparseComment {
    /// Insert the comment before the first node whose location is equal or higher than this location
    pub match_location: i32,
    /// Insert newlines before the comment (non-zero if the source comment was separated from the prior token by at least one newline)
    pub newlines_before_comment: i32,
    /// Insert newlines after the comment (non-zero if the source comment was separated from the next token by at least one newline)
    pub newlines_after_comment: i32,
    /// The actual comment string, including comment start/end tokens and newline characters (if any)
    pub text: String,
}
