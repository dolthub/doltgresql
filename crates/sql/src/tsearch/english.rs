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

//! The Snowball English (Porter2) stemmer, ported from the stem_UTF_8_english.c that Snowball 2.2.0 generates and
//! Postgres ships, with the Snowball runtime it calls.

/// Among is a table of strings that a step looks for, each with the number of the action it selects, where -1 selects
/// none.
type Among = &'static [(&'static [u8], i32)];

/// A_0 to A_10 are the stemmer's tables, named as the generated C names them.
const A_0: Among = &[(b"arsen", -1), (b"commun", -1), (b"gener", -1)];
const A_1: Among = &[(b"'", 1), (b"'s'", 1), (b"'s", 1)];
const A_2: Among = &[(b"ied", 2), (b"s", 3), (b"ies", 2), (b"sses", 1), (b"ss", -1), (b"us", -1)];
const A_3: Among = &[
    (b"", 3),
    (b"bb", 2),
    (b"dd", 2),
    (b"ff", 2),
    (b"gg", 2),
    (b"bl", 1),
    (b"mm", 2),
    (b"nn", 2),
    (b"pp", 2),
    (b"rr", 2),
    (b"at", 1),
    (b"tt", 2),
    (b"iz", 1),
];
const A_4: Among = &[(b"ed", 2), (b"eed", 1), (b"ing", 2), (b"edly", 2), (b"eedly", 1), (b"ingly", 2)];
const A_5: Among = &[
    (b"anci", 3),
    (b"enci", 2),
    (b"ogi", 13),
    (b"li", 15),
    (b"bli", 12),
    (b"abli", 4),
    (b"alli", 8),
    (b"fulli", 9),
    (b"lessli", 14),
    (b"ousli", 10),
    (b"entli", 5),
    (b"aliti", 8),
    (b"biliti", 12),
    (b"iviti", 11),
    (b"tional", 1),
    (b"ational", 7),
    (b"alism", 8),
    (b"ation", 7),
    (b"ization", 6),
    (b"izer", 6),
    (b"ator", 7),
    (b"iveness", 11),
    (b"fulness", 9),
    (b"ousness", 10),
];
const A_6: Among = &[
    (b"icate", 4),
    (b"ative", 6),
    (b"alize", 3),
    (b"iciti", 4),
    (b"ical", 4),
    (b"tional", 1),
    (b"ational", 2),
    (b"ful", 5),
    (b"ness", 5),
];
const A_7: Among = &[
    (b"ic", 1),
    (b"ance", 1),
    (b"ence", 1),
    (b"able", 1),
    (b"ible", 1),
    (b"ate", 1),
    (b"ive", 1),
    (b"ize", 1),
    (b"iti", 1),
    (b"al", 1),
    (b"ism", 1),
    (b"ion", 2),
    (b"er", 1),
    (b"ous", 1),
    (b"ant", 1),
    (b"ent", 1),
    (b"ment", 1),
    (b"ement", 1),
];
const A_8: Among = &[(b"e", 1), (b"l", 2)];
const A_9: Among = &[
    (b"succeed", -1),
    (b"proceed", -1),
    (b"exceed", -1),
    (b"canning", -1),
    (b"inning", -1),
    (b"earring", -1),
    (b"herring", -1),
    (b"outing", -1),
];
const A_10: Among = &[
    (b"andes", -1),
    (b"atlas", -1),
    (b"bias", -1),
    (b"cosmos", -1),
    (b"dying", 3),
    (b"early", 9),
    (b"gently", 7),
    (b"howe", -1),
    (b"idly", 6),
    (b"lying", 4),
    (b"news", -1),
    (b"only", 10),
    (b"singly", 11),
    (b"skies", 2),
    (b"skis", 1),
    (b"sky", -1),
    (b"tying", 5),
    (b"ugly", 8),
];

/// V, V_WXY, and VALID_LI are the stemmer's groupings of characters: the vowels, the vowels with w, x, and Y, and
/// the letters that may come before a deleted li.
const V: &[u8] = b"aeiouy";
const V_WXY: &[u8] = b"Yaeiouwxy";
const VALID_LI: &[u8] = b"cdeghkmnrt";

/// stem returns the stem of a lowercase word.
pub fn stem(word: &str) -> String {
    let mut z = Env { p: word.as_bytes().to_vec(), c: 0, l: word.len(), lb: 0, bra: 0, ket: word.len(), i: [0; 3] };
    z.stem();
    String::from_utf8_lossy(&z.p).into_owned()
}

/// Env is the Snowball runtime's SN_env: the word, the cursor, the limits, the slice to replace, and the stemmer's
/// integer variables p2, p1, and Y_found.
struct Env {
    p: Vec<u8>,
    c: usize,
    l: usize,
    lb: usize,
    bra: usize,
    ket: usize,
    i: [usize; 3],
}

/// decode returns the character at the start of bytes and its length, as the Snowball runtime's get_utf8 does.
fn decode(bytes: &[u8]) -> Option<(char, usize)> {
    let text =
        std::str::from_utf8(&bytes[..bytes.len().min(4)]).or_else(|e| std::str::from_utf8(&bytes[..e.valid_up_to()]));
    text.ok().and_then(|t| t.chars().next()).map(|c| (c, c.len_utf8()))
}

impl Env {
    /// grouping is the Snowball runtime's in_grouping_U and out_grouping_U, forward or backward: it returns None at the
    /// limit, the length of a character that is in `group` when `inside` is false, or out of it when `inside` is
    /// true, and otherwise moves past such characters, once or while `repeat` holds, and returns zero.
    fn grouping(&mut self, group: &[u8], inside: bool, backward: bool, repeat: bool) -> Option<usize> {
        loop {
            let (ch, width) = match backward {
                true => {
                    let start = (self.lb..self.c).rev().find(|&i| self.p[i] & 0xC0 != 0x80)?;
                    decode(&self.p[start..self.c])?
                }
                false => decode(&self.p[self.c..self.l])?,
            };
            let member = ch.is_ascii() && group.contains(&(ch as u8));
            if member != inside {
                return Some(width);
            }
            match backward {
                true => self.c -= width,
                false => self.c += width,
            }
            if !repeat {
                return Some(0);
            }
        }
    }

    /// in_grouping succeeds, moving past one character, when the next character is in a group.
    fn in_grouping(&mut self, group: &[u8]) -> bool {
        self.grouping(group, true, false, false) == Some(0)
    }

    /// in_grouping_b succeeds, moving back past one character, when the previous character is in a group.
    fn in_grouping_b(&mut self, group: &[u8]) -> bool {
        self.grouping(group, true, true, false) == Some(0)
    }

    /// out_grouping_b succeeds, moving back past one character, when the previous character is out of a group.
    fn out_grouping_b(&mut self, group: &[u8]) -> bool {
        self.grouping(group, false, true, false) == Some(0)
    }

    /// gopast moves past the next character that is in a group, when `inside` is true, or out of it, failing at the end.
    fn gopast(&mut self, group: &[u8], inside: bool) -> bool {
        match self.grouping(group, !inside, false, true) {
            Some(width) => {
                self.c += width;
                true
            }
            None => false,
        }
    }

    /// gopast_b moves back past the previous character that is in a group, failing at the start.
    fn gopast_b(&mut self, group: &[u8]) -> bool {
        match self.grouping(group, false, true, true) {
            Some(width) => {
                self.c -= width;
                true
            }
            None => false,
        }
    }

    /// skip moves forward past n characters, as skip_utf8 does, failing at the end.
    fn skip(&mut self, n: usize) -> bool {
        let mut c = self.c;
        for _ in 0..n {
            if c >= self.l {
                return false;
            }
            c += 1;
            while c < self.l && self.p[c] & 0xC0 == 0x80 {
                c += 1;
            }
        }
        self.c = c;
        true
    }

    /// skip_b moves back past n characters, as skip_b_utf8 does, failing at the start.
    fn skip_b(&mut self, n: usize) -> bool {
        let mut c = self.c;
        for _ in 0..n {
            if c <= self.lb {
                return false;
            }
            c -= 1;
            while c > self.lb && self.p[c] & 0xC0 == 0x80 {
                c -= 1;
            }
        }
        self.c = c;
        true
    }

    /// find_among returns the action of the longest string of a table that follows the cursor, moving past it, or 0.
    fn find_among(&mut self, among: Among) -> i32 {
        let rest = &self.p[self.c..self.l];
        match among.iter().filter(|(s, _)| rest.starts_with(s)).max_by_key(|(s, _)| s.len()) {
            Some((s, action)) => {
                self.c += s.len();
                *action
            }
            None => 0,
        }
    }

    /// find_among_b returns the action of the longest string of a table that precedes the cursor, moving back past it,
    /// or 0.
    fn find_among_b(&mut self, among: Among) -> i32 {
        let before = &self.p[self.lb..self.c];
        match among.iter().filter(|(s, _)| before.ends_with(s)).max_by_key(|(s, _)| s.len()) {
            Some((s, action)) => {
                self.c -= s.len();
                *action
            }
            None => 0,
        }
    }

    /// replace replaces the bytes between two positions, moving the cursor and limit as replace_s does, and returns
    /// how much longer the word became.
    fn replace(&mut self, bra: usize, ket: usize, s: &[u8]) -> isize {
        let adjustment = s.len() as isize - (ket - bra) as isize;
        self.p.splice(bra..ket, s.iter().copied());
        if adjustment != 0 {
            self.l = (self.l as isize + adjustment) as usize;
            if self.c >= ket {
                self.c = (self.c as isize + adjustment) as usize;
            } else if self.c > bra {
                self.c = bra;
            }
        }
        adjustment
    }

    /// slice_from replaces the slice between `bra` and `ket`.
    fn slice_from(&mut self, s: &[u8]) {
        self.replace(self.bra, self.ket, s);
    }

    /// insert inserts bytes at the cursor, leaving the cursor where it was, as insert_s does.
    fn insert(&mut self, s: &[u8]) {
        let (saved, at) = (self.c, self.c);
        let adjustment = self.replace(at, at, s);
        if at <= self.bra {
            self.bra = (self.bra as isize + adjustment) as usize;
        }
        if at <= self.ket {
            self.ket = (self.ket as isize + adjustment) as usize;
        }
        self.c = saved;
    }

    /// next_is moves past the next byte when it is one of the given ones.
    fn next_is(&mut self, bytes: &[u8]) -> bool {
        let found = self.c < self.l && bytes.contains(&self.p[self.c]);
        if found {
            self.c += 1;
        }
        found
    }

    /// previous_is moves back past the previous byte when it is one of the given ones.
    fn previous_is(&mut self, bytes: &[u8]) -> bool {
        let found = self.c > self.lb && bytes.contains(&self.p[self.c - 1]);
        if found {
            self.c -= 1;
        }
        found
    }

    /// prelude drops a leading apostrophe and marks each y that starts the word or follows a vowel as Y.
    fn prelude(&mut self) {
        self.i[2] = 0;
        let c1 = self.c;
        self.bra = self.c;
        if self.next_is(b"'") {
            self.ket = self.c;
            self.slice_from(b"");
        }
        self.c = c1;
        self.bra = self.c;
        if self.next_is(b"y") {
            self.ket = self.c;
            self.slice_from(b"Y");
            self.i[2] = 1;
        }
        self.c = c1;
        loop {
            let c4 = self.c;
            let found = loop {
                let c5 = self.c;
                if self.in_grouping(V) {
                    self.bra = self.c;
                    if self.next_is(b"y") {
                        self.ket = self.c;
                        self.c = c5;
                        break true;
                    }
                }
                self.c = c5;
                if !self.skip(1) {
                    break false;
                }
            };
            if !found {
                self.c = c4;
                break;
            }
            self.slice_from(b"Y");
            self.i[2] = 1;
        }
        self.c = c1;
    }

    /// mark_regions sets p1 and p2, the starts of the regions R1 and R2.
    fn mark_regions(&mut self) {
        self.i[1] = self.l;
        self.i[0] = self.l;
        let c1 = self.c;
        'regions: {
            let c2 = self.c;
            if self.find_among(A_0) == 0 {
                self.c = c2;
                if !self.gopast(V, true) || !self.gopast(V, false) {
                    break 'regions;
                }
            }
            self.i[1] = self.c;
            if !self.gopast(V, true) || !self.gopast(V, false) {
                break 'regions;
            }
            self.i[0] = self.c;
        }
        self.c = c1;
    }

    /// shortv reports whether the word ends in a short syllable.
    fn shortv(&mut self) -> bool {
        let m1 = self.l - self.c;
        if self.out_grouping_b(V_WXY) && self.in_grouping_b(V) && self.out_grouping_b(V) {
            return true;
        }
        self.c = self.l - m1;
        self.out_grouping_b(V) && self.in_grouping_b(V) && self.c <= self.lb
    }

    /// r1 reports whether the cursor is in R1.
    fn r1(&self) -> bool {
        self.i[1] <= self.c
    }

    /// r2 reports whether the cursor is in R2.
    fn r2(&self) -> bool {
        self.i[0] <= self.c
    }

    /// step_1a removes possessives and plural endings.
    fn step_1a(&mut self) -> bool {
        let m1 = self.l - self.c;
        self.ket = self.c;
        if self.find_among_b(A_1) != 0 {
            self.bra = self.c;
            self.slice_from(b"");
        } else {
            self.c = self.l - m1;
        }
        self.ket = self.c;
        let among = self.find_among_b(A_2);
        if among == 0 {
            return false;
        }
        self.bra = self.c;
        match among {
            1 => self.slice_from(b"ss"),
            2 => {
                let m2 = self.l - self.c;
                if self.skip_b(2) {
                    self.slice_from(b"i");
                } else {
                    self.c = self.l - m2;
                    self.slice_from(b"ie");
                }
            }
            3 => {
                if !self.skip_b(1) || !self.gopast_b(V) {
                    return false;
                }
                self.slice_from(b"");
            }
            _ => {}
        }
        true
    }

    /// step_1b removes ed, ing, and their forms, then fixes up the stem.
    fn step_1b(&mut self) -> bool {
        self.ket = self.c;
        let among = self.find_among_b(A_4);
        if among == 0 {
            return false;
        }
        self.bra = self.c;
        match among {
            1 => {
                if !self.r1() {
                    return false;
                }
                self.slice_from(b"ee");
            }
            2 => {
                let m_test1 = self.l - self.c;
                if !self.gopast_b(V) {
                    return false;
                }
                self.c = self.l - m_test1;
                self.slice_from(b"");
                let m_test2 = self.l - self.c;
                let among = self.find_among_b(A_3);
                if among == 0 {
                    return false;
                }
                self.c = self.l - m_test2;
                match among {
                    1 => self.insert(b"e"),
                    2 => {
                        self.ket = self.c;
                        if !self.skip_b(1) {
                            return false;
                        }
                        self.bra = self.c;
                        self.slice_from(b"");
                    }
                    3 => {
                        if self.c != self.i[1] {
                            return false;
                        }
                        let m_test3 = self.l - self.c;
                        if !self.shortv() {
                            return false;
                        }
                        self.c = self.l - m_test3;
                        self.insert(b"e");
                    }
                    _ => {}
                }
            }
            _ => {}
        }
        true
    }

    /// step_1c turns a final y or Y after a consonant into i.
    fn step_1c(&mut self) -> bool {
        self.ket = self.c;
        let m1 = self.l - self.c;
        if !self.previous_is(b"y") {
            self.c = self.l - m1;
            if !self.previous_is(b"Y") {
                return false;
            }
        }
        self.bra = self.c;
        if !self.out_grouping_b(V) || self.c <= self.lb {
            return false;
        }
        self.slice_from(b"i");
        true
    }

    /// step_2 replaces the double suffixes in R1.
    fn step_2(&mut self) -> bool {
        self.ket = self.c;
        let among = self.find_among_b(A_5);
        if among == 0 {
            return false;
        }
        self.bra = self.c;
        if !self.r1() {
            return false;
        }
        let to: &[u8] = match among {
            1 => b"tion",
            2 => b"ence",
            3 => b"ance",
            4 => b"able",
            5 => b"ent",
            6 => b"ize",
            7 => b"ate",
            8 => b"al",
            9 => b"ful",
            10 => b"ous",
            11 => b"ive",
            12 => b"ble",
            13 => {
                if !self.previous_is(b"l") {
                    return false;
                }
                b"og"
            }
            14 => b"less",
            15 => {
                if !self.in_grouping_b(VALID_LI) {
                    return false;
                }
                b""
            }
            _ => return true,
        };
        self.slice_from(to);
        true
    }

    /// step_3 replaces the suffixes in R1 that step 2 leaves.
    fn step_3(&mut self) -> bool {
        self.ket = self.c;
        let among = self.find_among_b(A_6);
        if among == 0 {
            return false;
        }
        self.bra = self.c;
        if !self.r1() {
            return false;
        }
        let to: &[u8] = match among {
            1 => b"tion",
            2 => b"ate",
            3 => b"al",
            4 => b"ic",
            5 => b"",
            6 => {
                if !self.r2() {
                    return false;
                }
                b""
            }
            _ => return true,
        };
        self.slice_from(to);
        true
    }

    /// step_4 deletes the suffixes in R2.
    fn step_4(&mut self) -> bool {
        self.ket = self.c;
        let among = self.find_among_b(A_7);
        if among == 0 {
            return false;
        }
        self.bra = self.c;
        if !self.r2() {
            return false;
        }
        if among == 2 && !self.previous_is(b"st") {
            return false;
        }
        if among == 1 || among == 2 {
            self.slice_from(b"");
        }
        true
    }

    /// step_5 deletes a final e or the second l of a final ll.
    fn step_5(&mut self) -> bool {
        self.ket = self.c;
        let among = self.find_among_b(A_8);
        if among == 0 {
            return false;
        }
        self.bra = self.c;
        match among {
            1 => {
                let m1 = self.l - self.c;
                if !self.r2() {
                    self.c = self.l - m1;
                    if !self.r1() {
                        return false;
                    }
                    let m2 = self.l - self.c;
                    if self.shortv() {
                        return false;
                    }
                    self.c = self.l - m2;
                }
                self.slice_from(b"");
            }
            2 => {
                if !self.r2() || !self.previous_is(b"l") {
                    return false;
                }
                self.slice_from(b"");
            }
            _ => {}
        }
        true
    }

    /// exception2 reports whether the word is one that step 1a leaves as it is.
    fn exception2(&mut self) -> bool {
        self.ket = self.c;
        if self.find_among_b(A_9) == 0 {
            return false;
        }
        self.bra = self.c;
        self.c <= self.lb
    }

    /// exception1 replaces the words with special stems, reporting whether the word was one.
    fn exception1(&mut self) -> bool {
        self.bra = self.c;
        let among = self.find_among(A_10);
        if among == 0 {
            return false;
        }
        self.ket = self.c;
        if self.c < self.l {
            return false;
        }
        let to: &[u8] = match among {
            1 => b"ski",
            2 => b"sky",
            3 => b"die",
            4 => b"lie",
            5 => b"tie",
            6 => b"idl",
            7 => b"gentl",
            8 => b"ugli",
            9 => b"earli",
            10 => b"onli",
            11 => b"singl",
            _ => return true,
        };
        self.slice_from(to);
        true
    }

    /// postlude turns every Y back into y.
    fn postlude(&mut self) {
        if self.i[2] == 0 {
            return;
        }
        loop {
            let c1 = self.c;
            let found = loop {
                let c2 = self.c;
                self.bra = self.c;
                if self.next_is(b"Y") {
                    self.ket = self.c;
                    self.c = c2;
                    break true;
                }
                self.c = c2;
                if !self.skip(1) {
                    break false;
                }
            };
            if !found {
                self.c = c1;
                break;
            }
            self.slice_from(b"y");
        }
    }

    /// stem runs the stemmer, as english_UTF_8_stem does.
    fn stem(&mut self) {
        let c1 = self.c;
        if self.exception1() {
            return;
        }
        self.c = c1;
        if !self.skip(3) {
            return;
        }
        self.c = c1;
        self.prelude();
        self.mark_regions();
        self.lb = self.c;
        self.c = self.l;
        let m3 = self.l - self.c;
        self.step_1a();
        self.c = self.l - m3;
        let m4 = self.l - self.c;
        if !self.exception2() {
            self.c = self.l - m4;
            for step in [Env::step_1b, Env::step_1c, Env::step_2, Env::step_3, Env::step_4, Env::step_5] {
                let m = self.l - self.c;
                step(self);
                self.c = self.l - m;
            }
        }
        self.c = self.lb;
        let c11 = self.c;
        self.postlude();
        self.c = c11;
    }
}
