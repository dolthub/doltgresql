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

//! XPath 1.0 over a parsed XML document: compiling expressions and evaluating them as libxml2 does.

use std::collections::HashMap;

use super::parse::{self, Element, Name, Node};

/// Kind is what a node of a document is.
#[derive(Clone, Debug)]
pub enum Kind {
    Root,
    Element(Box<Element>),
    Attribute(parse::Attribute),
    Text(String),
    CData(String),
    Comment(String),
    Pi(String, String),
}

/// DocNode is a node of a document with its parent, children, and attributes, by their positions in document order.
#[derive(Clone, Debug)]
pub struct DocNode {
    pub kind: Kind,
    pub parent: Option<usize>,
    pub children: Vec<usize>,
    pub attributes: Vec<usize>,
}

/// Document is a parsed document's nodes in document order, the root first.
#[derive(Clone, Debug)]
pub struct Document {
    pub nodes: Vec<DocNode>,
}

impl Document {
    /// new builds a document from the top-level nodes of a parse.
    pub fn new(top: Vec<Node>) -> Document {
        let mut document = Document {
            nodes: vec![DocNode { kind: Kind::Root, parent: None, children: Vec::new(), attributes: Vec::new() }],
        };
        for node in top {
            let index = document.add(node, 0);
            document.nodes[0].children.push(index);
        }
        document
    }

    /// add adds a node and its descendants under a parent, returning its position.
    fn add(&mut self, node: Node, parent: usize) -> usize {
        let index = self.nodes.len();
        let kind = match node {
            Node::Element(mut element) => {
                let children = std::mem::take(&mut element.children);
                let attributes = element.attributes.clone();
                self.nodes.push(DocNode {
                    kind: Kind::Element(element),
                    parent: Some(parent),
                    children: Vec::new(),
                    attributes: Vec::new(),
                });
                for attribute in attributes {
                    let position = self.nodes.len();
                    self.nodes.push(DocNode {
                        kind: Kind::Attribute(attribute),
                        parent: Some(index),
                        children: Vec::new(),
                        attributes: Vec::new(),
                    });
                    self.nodes[index].attributes.push(position);
                }
                for child in children {
                    let position = self.add(child, index);
                    self.nodes[index].children.push(position);
                }
                return index;
            }
            Node::Text(text) => Kind::Text(text),
            Node::CData(text) => Kind::CData(text),
            Node::Comment(text) => Kind::Comment(text),
            Node::Pi(target, data) => Kind::Pi(target, data),
        };
        self.nodes.push(DocNode { kind, parent: Some(parent), children: Vec::new(), attributes: Vec::new() });
        index
    }

    /// string_value returns the XPath string value of a node.
    pub fn string_value(&self, index: usize) -> String {
        match &self.nodes[index].kind {
            Kind::Root | Kind::Element(_) => {
                let mut out = String::new();
                self.collect_text(index, &mut out);
                out
            }
            Kind::Attribute(attribute) => attribute.value.clone(),
            Kind::Text(text) | Kind::CData(text) | Kind::Comment(text) => text.clone(),
            Kind::Pi(_, data) => data.clone(),
        }
    }

    /// collect_text appends the text of a node's descendants.
    fn collect_text(&self, index: usize, out: &mut String) {
        for &child in &self.nodes[index].children {
            match &self.nodes[child].kind {
                Kind::Text(text) | Kind::CData(text) => out.push_str(text),
                Kind::Element(_) => self.collect_text(child, out),
                _ => {}
            }
        }
    }

    /// tree rebuilds the parsed node at a position with its descendants.
    pub fn tree(&self, index: usize) -> Option<Node> {
        Some(match &self.nodes[index].kind {
            Kind::Root | Kind::Attribute(_) => return None,
            Kind::Element(element) => {
                let mut element = (**element).clone();
                element.children = self.nodes[index].children.iter().filter_map(|&c| self.tree(c)).collect();
                Node::Element(Box::new(element))
            }
            Kind::Text(text) => Node::Text(text.clone()),
            Kind::CData(text) => Node::CData(text.clone()),
            Kind::Comment(text) => Node::Comment(text.clone()),
            Kind::Pi(target, data) => Node::Pi(target.clone(), data.clone()),
        })
    }

    /// serialize_copy writes the node at a position as libxml2 writes a copy of it, which declares at its top the
    /// namespaces its names use that its ancestors declared.
    pub fn serialize_copy(&self, index: usize) -> String {
        let mut out = String::new();
        if let Kind::Root = self.nodes[index].kind {
            for &child in &self.nodes[index].children {
                if let Some(tree) = self.tree(child) {
                    parse::serialize(&tree, &mut out);
                    out.push('\n');
                }
            }
            return out;
        }
        let Some(mut tree) = self.tree(index) else { return out };
        if let Node::Element(element) = &mut tree {
            let mut declared: Vec<Option<String>> = element.namespaces.iter().map(|(p, _)| p.clone()).collect();
            let mut added = Vec::new();
            reconcile(element, &mut declared, &mut added, true);
            element.namespaces.extend(added);
        }
        parse::serialize(&tree, &mut out);
        out
    }
}

/// reconcile finds the namespaces an element's subtree uses without declaring, as libxml2's xmlCopyNode adds them to
/// the top of a copy.
fn reconcile(
    element: &Element,
    declared: &mut Vec<Option<String>>,
    added: &mut Vec<(Option<String>, String)>,
    top: bool,
) {
    let scope = declared.len();
    if !top {
        declared.extend(element.namespaces.iter().map(|(p, _)| p.clone()));
    }
    let mut need = |name: &Name, element_name: bool| {
        let Some(uri) = &name.uri else { return };
        if name.prefix.as_deref() == Some("xml") {
            return;
        }
        if (element_name || name.prefix.is_some())
            && !declared.contains(&name.prefix)
            && !added.iter().any(|(p, _)| *p == name.prefix)
        {
            added.push((name.prefix.clone(), uri.clone()));
        }
    };
    need(&element.name, true);
    for attribute in &element.attributes {
        need(&attribute.name, false);
    }
    for child in &element.children {
        if let Node::Element(child) = child {
            reconcile(child, declared, added, false);
        }
    }
    declared.truncate(scope);
}

/// Value is the value of an XPath expression.
#[derive(Clone, Debug)]
pub enum Value {
    Nodes(Vec<usize>),
    Bool(bool),
    Number(f64),
    String(String),
}

/// Error is an XPath compile or evaluation failure, with libxml2's message.
#[derive(Debug)]
pub struct Error(pub String);

/// invalid returns libxml2's error for an expression that does not compile.
fn invalid() -> Error {
    Error("Invalid expression".into())
}

/// Axis is an XPath axis.
#[derive(Clone, Copy, Debug, PartialEq)]
pub enum Axis {
    Child,
    Descendant,
    DescendantOrSelf,
    Parent,
    Ancestor,
    AncestorOrSelf,
    FollowingSibling,
    PrecedingSibling,
    Following,
    Preceding,
    Attribute,
    SelfAxis,
    Namespace,
}

/// Test is a node test.
#[derive(Clone, Debug)]
pub enum Test {
    /// A name with an optional prefix, where a local name of `*` matches any.
    Name(Option<String>, String),
    Node,
    Text,
    Comment,
    Pi(Option<String>),
}

/// Step is a location step.
#[derive(Clone, Debug)]
pub struct Step {
    axis: Axis,
    test: Test,
    predicates: Vec<Expr>,
}

/// Expr is a compiled XPath expression.
#[derive(Clone, Debug)]
pub enum Expr {
    Number(f64),
    Literal(String),
    Binary(&'static str, Box<Expr>, Box<Expr>),
    Negate(Box<Expr>),
    Union(Box<Expr>, Box<Expr>),
    Call(String, Vec<Expr>),
    /// A path from the root, or from the context node when not absolute.
    Path {
        absolute: bool,
        steps: Vec<Step>,
    },
    /// A filter expression with predicates, then steps from its nodes.
    Filter {
        primary: Box<Expr>,
        predicates: Vec<Expr>,
        steps: Vec<Step>,
    },
}

/// Token is a lexical token of an XPath expression.
#[derive(Clone, Debug, PartialEq)]
enum Token {
    Number(f64),
    Literal(String),
    Name(String),
    Symbol(&'static str),
    /// The `*` name test, or a prefixed one like `p:*`.
    Star(Option<String>),
    Operator(&'static str),
}

/// tokenize splits an expression into tokens, telling operators from names as the XPath grammar does.
fn tokenize(text: &str) -> Result<Vec<Token>, Error> {
    let chars: Vec<char> = text.chars().collect();
    let mut tokens: Vec<Token> = Vec::new();
    let mut i = 0;
    let operator_context = |tokens: &[Token]| match tokens.last() {
        None => false,
        Some(Token::Symbol(s)) => !matches!(*s, "@" | "::" | "(" | "[" | "," | "/" | "//"),
        Some(Token::Operator(_)) => false,
        _ => true,
    };
    while i < chars.len() {
        let c = chars[i];
        if c.is_whitespace() {
            i += 1;
            continue;
        }
        let two: String = chars[i..chars.len().min(i + 2)].iter().collect();
        match two.as_str() {
            "//" | "::" | ".." | "!=" | "<=" | ">=" => {
                let symbol = match two.as_str() {
                    "//" => Token::Symbol("//"),
                    "::" => Token::Symbol("::"),
                    ".." => Token::Symbol(".."),
                    "!=" => Token::Operator("!="),
                    "<=" => Token::Operator("<="),
                    _ => Token::Operator(">="),
                };
                tokens.push(symbol);
                i += 2;
                continue;
            }
            _ => {}
        }
        if c.is_ascii_digit() || (c == '.' && chars.get(i + 1).is_some_and(|d| d.is_ascii_digit())) {
            let start = i;
            while i < chars.len() && (chars[i].is_ascii_digit() || chars[i] == '.') {
                i += 1;
            }
            let text: String = chars[start..i].iter().collect();
            tokens.push(Token::Number(text.parse().map_err(|_| invalid())?));
            continue;
        }
        match c {
            '"' | '\'' => {
                let end = chars[i + 1..].iter().position(|&d| d == c).ok_or_else(invalid)?;
                tokens.push(Token::Literal(chars[i + 1..i + 1 + end].iter().collect()));
                i += end + 2;
            }
            '/' => {
                tokens.push(Token::Symbol("/"));
                i += 1;
            }
            '(' | ')' | '[' | ']' | '@' | ',' | '.' => {
                let symbol = match c {
                    '(' => "(",
                    ')' => ")",
                    '[' => "[",
                    ']' => "]",
                    '@' => "@",
                    ',' => ",",
                    _ => ".",
                };
                tokens.push(Token::Symbol(symbol));
                i += 1;
            }
            '|' | '+' | '-' | '=' | '<' | '>' => {
                let operator = match c {
                    '|' => "|",
                    '+' => "+",
                    '-' => "-",
                    '=' => "=",
                    '<' => "<",
                    _ => ">",
                };
                tokens.push(Token::Operator(operator));
                i += 1;
            }
            '*' => {
                if operator_context(&tokens) {
                    tokens.push(Token::Operator("*"));
                } else {
                    tokens.push(Token::Star(None));
                }
                i += 1;
            }
            '$' => return Err(Error("Undefined variable".into())),
            c if c.is_alphabetic() || c == '_' => {
                let start = i;
                while i < chars.len()
                    && (chars[i].is_alphanumeric() || matches!(chars[i], '_' | '-' | '.'))
                    && !(chars[i] == '.' && chars.get(i + 1) == Some(&'.'))
                {
                    i += 1;
                }
                let mut name: String = chars[start..i].iter().collect();
                if chars.get(i) == Some(&':') && chars.get(i + 1) != Some(&':') {
                    if chars.get(i + 1) == Some(&'*') {
                        tokens.push(Token::Star(Some(name)));
                        i += 2;
                        continue;
                    }
                    let rest_start = i + 1;
                    let mut j = rest_start;
                    while j < chars.len() && (chars[j].is_alphanumeric() || matches!(chars[j], '_' | '-' | '.')) {
                        j += 1;
                    }
                    if j == rest_start {
                        return Err(invalid());
                    }
                    name = format!("{name}:{}", chars[rest_start..j].iter().collect::<String>());
                    i = j;
                }
                if operator_context(&tokens) {
                    let operator = match name.as_str() {
                        "and" => "and",
                        "or" => "or",
                        "mod" => "mod",
                        "div" => "div",
                        _ => return Err(invalid()),
                    };
                    tokens.push(Token::Operator(operator));
                } else {
                    tokens.push(Token::Name(name));
                }
            }
            _ => return Err(invalid()),
        }
    }
    Ok(tokens)
}

/// Compiler parses tokens into an expression.
struct Compiler {
    tokens: Vec<Token>,
    at: usize,
}

impl Compiler {
    /// peek returns the token being read.
    fn peek(&self) -> Option<&Token> {
        self.tokens.get(self.at)
    }

    /// peek_at returns a token ahead of the one being read.
    fn peek_at(&self, ahead: usize) -> Option<&Token> {
        self.tokens.get(self.at + ahead)
    }

    /// eat reads a symbol when it is next.
    fn eat(&mut self, symbol: &str) -> bool {
        if matches!(self.peek(), Some(Token::Symbol(s)) if *s == symbol) {
            self.at += 1;
            return true;
        }
        false
    }

    /// eat_operator reads an operator when it is next, returning it.
    fn eat_operator(&mut self, operators: &[&'static str]) -> Option<&'static str> {
        if let Some(Token::Operator(o)) = self.peek()
            && let Some(found) = operators.iter().find(|op| *op == o)
        {
            self.at += 1;
            return Some(found);
        }
        None
    }

    /// binary reads a left-associative chain of operators over operands that `next` reads.
    fn binary(
        &mut self,
        operators: &[&'static str],
        next: fn(&mut Compiler) -> Result<Expr, Error>,
    ) -> Result<Expr, Error> {
        let mut left = next(self)?;
        while let Some(operator) = self.eat_operator(operators) {
            let right = next(self)?;
            left = Expr::Binary(operator, Box::new(left), Box::new(right));
        }
        Ok(left)
    }

    /// or reads an or-expression.
    fn or(&mut self) -> Result<Expr, Error> {
        self.binary(&["or"], Compiler::and)
    }

    /// and reads an and-expression.
    fn and(&mut self) -> Result<Expr, Error> {
        self.binary(&["and"], Compiler::equality)
    }

    /// equality reads an equality expression.
    fn equality(&mut self) -> Result<Expr, Error> {
        self.binary(&["=", "!="], Compiler::relational)
    }

    /// relational reads a relational expression.
    fn relational(&mut self) -> Result<Expr, Error> {
        self.binary(&["<", ">", "<=", ">="], Compiler::additive)
    }

    /// additive reads an additive expression.
    fn additive(&mut self) -> Result<Expr, Error> {
        self.binary(&["+", "-"], Compiler::multiplicative)
    }

    /// multiplicative reads a multiplicative expression.
    fn multiplicative(&mut self) -> Result<Expr, Error> {
        self.binary(&["*", "div", "mod"], Compiler::unary)
    }

    /// unary reads a negation or a union expression.
    fn unary(&mut self) -> Result<Expr, Error> {
        if self.eat_operator(&["-"]).is_some() {
            return Ok(Expr::Negate(Box::new(self.unary()?)));
        }
        let mut left = self.path()?;
        while self.eat_operator(&["|"]).is_some() {
            let right = self.path()?;
            left = Expr::Union(Box::new(left), Box::new(right));
        }
        Ok(left)
    }

    /// path reads a location path or a filter expression with steps after it.
    fn path(&mut self) -> Result<Expr, Error> {
        let primary_start = match self.peek() {
            Some(Token::Number(_) | Token::Literal(_)) => true,
            Some(Token::Symbol("(")) => true,
            Some(Token::Name(name)) => {
                matches!(self.peek_at(1), Some(Token::Symbol("(")))
                    && !matches!(name.as_str(), "node" | "text" | "comment" | "processing-instruction")
            }
            _ => false,
        };
        if !primary_start {
            return self.location_path();
        }
        let primary = self.primary()?;
        let mut predicates = Vec::new();
        while self.eat("[") {
            predicates.push(self.or()?);
            if !self.eat("]") {
                return Err(invalid());
            }
        }
        let mut steps = Vec::new();
        loop {
            if self.eat("//") {
                steps.push(Step { axis: Axis::DescendantOrSelf, test: Test::Node, predicates: Vec::new() });
                steps.push(self.step()?);
            } else if self.eat("/") {
                steps.push(self.step()?);
            } else {
                break;
            }
        }
        if predicates.is_empty() && steps.is_empty() {
            return Ok(primary);
        }
        Ok(Expr::Filter { primary: Box::new(primary), predicates, steps })
    }

    /// primary reads a number, literal, parenthesized expression, or function call.
    fn primary(&mut self) -> Result<Expr, Error> {
        match self.peek().cloned() {
            Some(Token::Number(n)) => {
                self.at += 1;
                Ok(Expr::Number(n))
            }
            Some(Token::Literal(s)) => {
                self.at += 1;
                Ok(Expr::Literal(s))
            }
            Some(Token::Symbol("(")) => {
                self.at += 1;
                let inner = self.or()?;
                if !self.eat(")") {
                    return Err(invalid());
                }
                Ok(inner)
            }
            Some(Token::Name(name)) => {
                self.at += 2;
                let mut args = Vec::new();
                if !self.eat(")") {
                    loop {
                        args.push(self.or()?);
                        if self.eat(")") {
                            break;
                        }
                        if !self.eat(",") {
                            return Err(invalid());
                        }
                    }
                }
                Ok(Expr::Call(name, args))
            }
            _ => Err(invalid()),
        }
    }

    /// location_path reads an absolute or relative location path.
    fn location_path(&mut self) -> Result<Expr, Error> {
        let mut steps = Vec::new();
        let absolute = if self.eat("//") {
            steps.push(Step { axis: Axis::DescendantOrSelf, test: Test::Node, predicates: Vec::new() });
            steps.push(self.step()?);
            true
        } else if self.eat("/") {
            if self.starts_step() {
                steps.push(self.step()?);
            }
            true
        } else {
            steps.push(self.step()?);
            false
        };
        loop {
            if self.eat("//") {
                steps.push(Step { axis: Axis::DescendantOrSelf, test: Test::Node, predicates: Vec::new() });
                steps.push(self.step()?);
            } else if self.eat("/") {
                steps.push(self.step()?);
            } else {
                break;
            }
        }
        Ok(Expr::Path { absolute, steps })
    }

    /// starts_step reports whether a location step comes next.
    fn starts_step(&self) -> bool {
        matches!(self.peek(), Some(Token::Name(_) | Token::Star(_) | Token::Symbol("@" | "." | "..")))
    }

    /// step reads a location step.
    fn step(&mut self) -> Result<Step, Error> {
        if self.eat(".") {
            return Ok(Step { axis: Axis::SelfAxis, test: Test::Node, predicates: Vec::new() });
        }
        if self.eat("..") {
            return Ok(Step { axis: Axis::Parent, test: Test::Node, predicates: Vec::new() });
        }
        let mut axis = Axis::Child;
        if self.eat("@") {
            axis = Axis::Attribute;
        } else if let (Some(Token::Name(name)), Some(Token::Symbol("::"))) = (self.peek().cloned(), self.peek_at(1)) {
            axis = match name.as_str() {
                "child" => Axis::Child,
                "descendant" => Axis::Descendant,
                "descendant-or-self" => Axis::DescendantOrSelf,
                "parent" => Axis::Parent,
                "ancestor" => Axis::Ancestor,
                "ancestor-or-self" => Axis::AncestorOrSelf,
                "following-sibling" => Axis::FollowingSibling,
                "preceding-sibling" => Axis::PrecedingSibling,
                "following" => Axis::Following,
                "preceding" => Axis::Preceding,
                "attribute" => Axis::Attribute,
                "self" => Axis::SelfAxis,
                "namespace" => Axis::Namespace,
                _ => return Err(invalid()),
            };
            self.at += 2;
        }
        let test = match self.peek().cloned() {
            Some(Token::Star(prefix)) => {
                self.at += 1;
                Test::Name(prefix, "*".into())
            }
            Some(Token::Name(name)) if matches!(self.peek_at(1), Some(Token::Symbol("("))) => {
                self.at += 2;
                let test = match name.as_str() {
                    "node" => Test::Node,
                    "text" => Test::Text,
                    "comment" => Test::Comment,
                    "processing-instruction" => match self.peek().cloned() {
                        Some(Token::Literal(target)) => {
                            self.at += 1;
                            Test::Pi(Some(target))
                        }
                        _ => Test::Pi(None),
                    },
                    _ => return Err(invalid()),
                };
                if !self.eat(")") {
                    return Err(invalid());
                }
                test
            }
            Some(Token::Name(name)) => {
                self.at += 1;
                match name.split_once(':') {
                    Some((prefix, local)) => Test::Name(Some(prefix.into()), local.into()),
                    None => Test::Name(None, name),
                }
            }
            _ => return Err(invalid()),
        };
        let mut predicates = Vec::new();
        while self.eat("[") {
            predicates.push(self.or()?);
            if !self.eat("]") {
                return Err(invalid());
            }
        }
        Ok(Step { axis, test, predicates })
    }
}

/// compile compiles an XPath expression.
pub fn compile(text: &str) -> Result<Expr, Error> {
    let tokens = tokenize(text)?;
    if tokens.is_empty() {
        return Err(invalid());
    }
    let mut compiler = Compiler { tokens, at: 0 };
    let expr = compiler.or()?;
    if compiler.at != compiler.tokens.len() {
        return Err(invalid());
    }
    Ok(expr)
}

/// Context is where an expression evaluates: the document, the namespaces its prefixes name, and the context node
/// with its position and the size of its node set.
struct Context<'a> {
    document: &'a Document,
    namespaces: &'a HashMap<String, String>,
    node: usize,
    position: usize,
    size: usize,
}

/// number_of converts a string to a number as XPath's number function does.
pub fn number_of(text: &str) -> f64 {
    let trimmed = text.trim_matches(|c: char| matches!(c, ' ' | '\t' | '\n' | '\r'));
    let valid = {
        let body = trimmed.strip_prefix('-').unwrap_or(trimmed);
        !body.is_empty()
            && body.chars().all(|c| c.is_ascii_digit() || c == '.')
            && body.matches('.').count() <= 1
            && body != "."
    };
    if valid { trimmed.parse().unwrap_or(f64::NAN) } else { f64::NAN }
}

/// number_text writes a number as XPath's string function does.
pub fn number_text(n: f64) -> String {
    if n.is_nan() {
        "NaN".into()
    } else if n.is_infinite() {
        if n > 0.0 { "Infinity".into() } else { "-Infinity".into() }
    } else if n == n.trunc() && n.abs() < 1e15 {
        format!("{}", n as i64)
    } else {
        format!("{n}")
    }
}

impl Value {
    /// to_bool converts the value as XPath's boolean function does.
    fn to_bool(&self) -> bool {
        match self {
            Value::Nodes(nodes) => !nodes.is_empty(),
            Value::Bool(b) => *b,
            Value::Number(n) => *n != 0.0 && !n.is_nan(),
            Value::String(s) => !s.is_empty(),
        }
    }

    /// to_number converts the value as XPath's number function does.
    fn to_number(&self, document: &Document) -> f64 {
        match self {
            Value::Nodes(_) => number_of(&self.to_text(document)),
            Value::Bool(b) => *b as u8 as f64,
            Value::Number(n) => *n,
            Value::String(s) => number_of(s),
        }
    }

    /// to_text converts the value as XPath's string function does.
    pub fn to_text(&self, document: &Document) -> String {
        match self {
            Value::Nodes(nodes) => nodes.first().map_or(String::new(), |&n| document.string_value(n)),
            Value::Bool(b) => b.to_string(),
            Value::Number(n) => number_text(*n),
            Value::String(s) => s.clone(),
        }
    }
}

impl Context<'_> {
    /// at returns a context at another node, with its position and the size of its node set.
    fn at(&self, node: usize, position: usize, size: usize) -> Context<'_> {
        Context { document: self.document, namespaces: self.namespaces, node, position, size }
    }

    /// matches reports whether a node passes a step's node test on its axis.
    fn matches(&self, index: usize, axis: Axis, test: &Test) -> Result<bool, Error> {
        let node = &self.document.nodes[index];
        Ok(match test {
            Test::Node => true,
            Test::Text => matches!(node.kind, Kind::Text(_) | Kind::CData(_)),
            Test::Comment => matches!(node.kind, Kind::Comment(_)),
            Test::Pi(target) => match &node.kind {
                Kind::Pi(t, _) => target.as_ref().is_none_or(|target| target == t),
                _ => false,
            },
            Test::Name(prefix, local) => {
                let name = match (&node.kind, axis) {
                    (Kind::Attribute(attribute), Axis::Attribute) => &attribute.name,
                    (Kind::Element(element), axis) if axis != Axis::Attribute => &element.name,
                    _ => return Ok(false),
                };
                if prefix.is_none() && local == "*" {
                    return Ok(true);
                }
                let uri = match prefix {
                    Some(prefix) => Some(
                        self.namespaces
                            .get(prefix)
                            .cloned()
                            .ok_or_else(|| Error("Undefined namespace prefix".into()))?,
                    ),
                    None => None,
                };
                name.uri == uri && (local == "*" || name.local == *local)
            }
        })
    }

    /// axis_nodes returns the nodes on an axis from a node, in document order.
    fn axis_nodes(&self, index: usize, axis: Axis) -> Vec<usize> {
        let nodes = &self.document.nodes;
        let mut out = Vec::new();
        let descendants = |start: usize, out: &mut Vec<usize>| {
            let mut stack: Vec<usize> = nodes[start].children.iter().rev().copied().collect();
            while let Some(n) = stack.pop() {
                out.push(n);
                stack.extend(nodes[n].children.iter().rev());
            }
        };
        match axis {
            Axis::Child => out.extend(&nodes[index].children),
            Axis::Attribute => out.extend(&nodes[index].attributes),
            Axis::SelfAxis => out.push(index),
            Axis::Descendant => descendants(index, &mut out),
            Axis::DescendantOrSelf => {
                out.push(index);
                descendants(index, &mut out);
            }
            Axis::Parent => out.extend(nodes[index].parent),
            Axis::Ancestor | Axis::AncestorOrSelf => {
                if axis == Axis::AncestorOrSelf {
                    out.push(index);
                }
                let mut current = nodes[index].parent;
                while let Some(p) = current {
                    out.push(p);
                    current = nodes[p].parent;
                }
                out.reverse();
            }
            Axis::FollowingSibling | Axis::PrecedingSibling => {
                if let Some(parent) = nodes[index].parent
                    && !matches!(nodes[index].kind, Kind::Attribute(_))
                {
                    let siblings = &nodes[parent].children;
                    let position = siblings.iter().position(|&s| s == index).unwrap_or(0);
                    if axis == Axis::FollowingSibling {
                        out.extend(&siblings[position + 1..]);
                    } else {
                        out.extend(&siblings[..position]);
                    }
                }
            }
            Axis::Following => {
                let mut current = index;
                while let Some(parent) = nodes[current].parent {
                    let siblings = &nodes[parent].children;
                    if let Some(position) = siblings.iter().position(|&s| s == current) {
                        for &sibling in &siblings[position + 1..] {
                            out.push(sibling);
                            descendants(sibling, &mut out);
                        }
                    }
                    current = parent;
                }
                out.sort_unstable();
            }
            Axis::Preceding => {
                let ancestors: Vec<usize> = self.axis_nodes(index, Axis::Ancestor);
                out.extend(
                    (1..index).filter(|n| !ancestors.contains(n) && !matches!(nodes[*n].kind, Kind::Attribute(_))),
                );
            }
            Axis::Namespace => {}
        }
        out
    }

    /// step applies a location step to a node set.
    fn step(&self, input: &[usize], step: &Step) -> Result<Vec<usize>, Error> {
        let mut result: Vec<usize> = Vec::new();
        for &node in input {
            let mut selected = Vec::new();
            for candidate in self.axis_nodes(node, step.axis) {
                if self.matches(candidate, step.axis, &step.test)? {
                    selected.push(candidate);
                }
            }
            let reverse =
                matches!(step.axis, Axis::Ancestor | Axis::AncestorOrSelf | Axis::PrecedingSibling | Axis::Preceding);
            for predicate in &step.predicates {
                let ordered: Vec<usize> =
                    if reverse { selected.iter().rev().copied().collect() } else { selected.clone() };
                let size = ordered.len();
                let mut kept = Vec::new();
                for (i, &n) in ordered.iter().enumerate() {
                    if self.at(n, i + 1, size).predicate(predicate)? {
                        kept.push(n);
                    }
                }
                if reverse {
                    kept.reverse();
                }
                selected = kept;
            }
            result.extend(selected);
        }
        result.sort_unstable();
        result.dedup();
        Ok(result)
    }

    /// predicate evaluates a predicate, where a number tests the context position.
    fn predicate(&self, predicate: &Expr) -> Result<bool, Error> {
        Ok(match self.eval(predicate)? {
            Value::Number(n) => n == self.position as f64,
            other => other.to_bool(),
        })
    }

    /// compare compares two values as XPath's comparison operators do, existentially over node sets.
    fn compare(&self, operator: &str, left: &Value, right: &Value) -> bool {
        let document = self.document;
        let strings = |nodes: &[usize]| nodes.iter().map(|&n| document.string_value(n)).collect::<Vec<_>>();
        let relational = !matches!(operator, "=" | "!=");
        let numbers = |a: f64, b: f64| match operator {
            "=" => a == b,
            "!=" => a != b,
            "<" => a < b,
            "<=" => a <= b,
            ">" => a > b,
            _ => a >= b,
        };
        match (left, right) {
            (Value::Nodes(a), Value::Nodes(b)) => {
                let (a, b) = (strings(a), strings(b));
                a.iter().any(|x| {
                    b.iter().any(|y| {
                        if relational { numbers(number_of(x), number_of(y)) } else { (x == y) == (operator == "=") }
                    })
                })
            }
            (Value::Nodes(nodes), other) | (other, Value::Nodes(nodes)) => {
                let flipped = matches!(right, Value::Nodes(_)) && !matches!(left, Value::Nodes(_));
                let ordered = |x: f64, y: f64| if flipped { numbers(y, x) } else { numbers(x, y) };
                strings(nodes).iter().any(|s| match other {
                    Value::Number(n) => ordered(number_of(s), *n),
                    Value::String(t) if !relational => (s == t) == (operator == "="),
                    Value::Bool(b) => {
                        let a = !nodes.is_empty();
                        if relational {
                            ordered(a as u8 as f64, *b as u8 as f64)
                        } else {
                            (a == *b) == (operator == "=")
                        }
                    }
                    other => ordered(number_of(s), other.to_number(document)),
                })
            }
            (a, b) if relational => numbers(a.to_number(document), b.to_number(document)),
            (Value::Bool(_), _) | (_, Value::Bool(_)) => (left.to_bool() == right.to_bool()) == (operator == "="),
            (Value::Number(_), _) | (_, Value::Number(_)) => {
                numbers(left.to_number(document), right.to_number(document))
            }
            (a, b) => (a.to_text(document) == b.to_text(document)) == (operator == "="),
        }
    }

    /// eval evaluates an expression.
    fn eval(&self, expr: &Expr) -> Result<Value, Error> {
        let document = self.document;
        Ok(match expr {
            Expr::Number(n) => Value::Number(*n),
            Expr::Literal(s) => Value::String(s.clone()),
            Expr::Negate(inner) => Value::Number(-self.eval(inner)?.to_number(document)),
            Expr::Binary(operator, left, right) => {
                let l = self.eval(left)?;
                match *operator {
                    "or" => return Ok(Value::Bool(l.to_bool() || self.eval(right)?.to_bool())),
                    "and" => return Ok(Value::Bool(l.to_bool() && self.eval(right)?.to_bool())),
                    _ => {}
                }
                let r = self.eval(right)?;
                match *operator {
                    "=" | "!=" | "<" | "<=" | ">" | ">=" => Value::Bool(self.compare(operator, &l, &r)),
                    _ => {
                        let (a, b) = (l.to_number(document), r.to_number(document));
                        Value::Number(match *operator {
                            "+" => a + b,
                            "-" => a - b,
                            "*" => a * b,
                            "div" => a / b,
                            _ => a % b,
                        })
                    }
                }
            }
            Expr::Union(left, right) => {
                let (Value::Nodes(mut a), Value::Nodes(b)) = (self.eval(left)?, self.eval(right)?) else {
                    return Err(Error("Invalid type".into()));
                };
                a.extend(b);
                a.sort_unstable();
                a.dedup();
                Value::Nodes(a)
            }
            Expr::Path { absolute, steps } => {
                let mut nodes = vec![if *absolute { 0 } else { self.node }];
                for step in steps {
                    nodes = self.step(&nodes, step)?;
                }
                Value::Nodes(nodes)
            }
            Expr::Filter { primary, predicates, steps } => {
                let value = self.eval(primary)?;
                if predicates.is_empty() && steps.is_empty() {
                    return Ok(value);
                }
                let Value::Nodes(mut nodes) = value else { return Err(Error("Invalid type".into())) };
                for predicate in predicates {
                    let size = nodes.len();
                    let mut kept = Vec::new();
                    for (i, &n) in nodes.iter().enumerate() {
                        if self.at(n, i + 1, size).predicate(predicate)? {
                            kept.push(n);
                        }
                    }
                    nodes = kept;
                }
                for step in steps {
                    nodes = self.step(&nodes, step)?;
                }
                Value::Nodes(nodes)
            }
            Expr::Call(name, args) => self.call(name, args)?,
        })
    }

    /// call evaluates a core function call.
    fn call(&self, name: &str, args: &[Expr]) -> Result<Value, Error> {
        let document = self.document;
        let values = args.iter().map(|a| self.eval(a)).collect::<Result<Vec<_>, _>>()?;
        let text = |i: usize| -> String {
            values.get(i).map_or_else(|| document.string_value(self.node), |v| v.to_text(document))
        };
        let arity = |min: usize, max: usize| {
            if values.len() < min || values.len() > max {
                Err(Error("Invalid number of arguments".into()))
            } else {
                Ok(())
            }
        };
        let node_name = |local: bool| -> String {
            let node = match values.first() {
                Some(Value::Nodes(nodes)) => nodes.first().copied(),
                Some(_) => None,
                None => Some(self.node),
            };
            match node.map(|n| &document.nodes[n].kind) {
                Some(Kind::Element(element)) => {
                    if local {
                        element.name.local.clone()
                    } else {
                        element.name.qualified()
                    }
                }
                Some(Kind::Attribute(attribute)) => {
                    if local {
                        attribute.name.local.clone()
                    } else {
                        attribute.name.qualified()
                    }
                }
                Some(Kind::Pi(target, _)) => target.clone(),
                _ => String::new(),
            }
        };
        Ok(match name {
            "last" => Value::Number(self.size as f64),
            "position" => Value::Number(self.position as f64),
            "count" => {
                arity(1, 1)?;
                match &values[0] {
                    Value::Nodes(nodes) => Value::Number(nodes.len() as f64),
                    _ => return Err(Error("Invalid type".into())),
                }
            }
            "local-name" => Value::String(node_name(true)),
            "name" => Value::String(node_name(false)),
            "namespace-uri" => {
                let node = match values.first() {
                    Some(Value::Nodes(nodes)) => nodes.first().copied(),
                    _ => Some(self.node),
                };
                Value::String(match node.map(|n| &document.nodes[n].kind) {
                    Some(Kind::Element(element)) => element.name.uri.clone().unwrap_or_default(),
                    Some(Kind::Attribute(attribute)) => attribute.name.uri.clone().unwrap_or_default(),
                    _ => String::new(),
                })
            }
            "string" => {
                arity(0, 1)?;
                Value::String(text(0))
            }
            "concat" => {
                if values.len() < 2 {
                    return Err(Error("Invalid number of arguments".into()));
                }
                Value::String(values.iter().map(|v| v.to_text(document)).collect())
            }
            "starts-with" => {
                arity(2, 2)?;
                Value::Bool(text(0).starts_with(&text(1)))
            }
            "contains" => {
                arity(2, 2)?;
                Value::Bool(text(0).contains(&text(1)))
            }
            "substring-before" => {
                arity(2, 2)?;
                let (s, t) = (text(0), text(1));
                Value::String(s.find(&t).map_or(String::new(), |i| s[..i].to_string()))
            }
            "substring-after" => {
                arity(2, 2)?;
                let (s, t) = (text(0), text(1));
                Value::String(s.find(&t).map_or(String::new(), |i| s[i + t.len()..].to_string()))
            }
            "substring" => {
                arity(2, 3)?;
                let s: Vec<char> = text(0).chars().collect();
                let start = values[1].to_number(document).round();
                let end = match values.get(2) {
                    Some(length) => start + length.to_number(document).round(),
                    None => f64::INFINITY,
                };
                Value::String(
                    s.iter()
                        .enumerate()
                        .filter(|(i, _)| {
                            let p = (*i + 1) as f64;
                            p >= start && p < end
                        })
                        .map(|(_, c)| c)
                        .collect(),
                )
            }
            "string-length" => {
                arity(0, 1)?;
                Value::Number(text(0).chars().count() as f64)
            }
            "normalize-space" => {
                arity(0, 1)?;
                Value::String(text(0).split_whitespace().collect::<Vec<_>>().join(" "))
            }
            "translate" => {
                arity(3, 3)?;
                let (from, to): (Vec<char>, Vec<char>) = (text(1).chars().collect(), text(2).chars().collect());
                Value::String(
                    text(0)
                        .chars()
                        .filter_map(|c| match from.iter().position(|&f| f == c) {
                            Some(i) => to.get(i).copied(),
                            None => Some(c),
                        })
                        .collect(),
                )
            }
            "boolean" => {
                arity(1, 1)?;
                Value::Bool(values[0].to_bool())
            }
            "not" => {
                arity(1, 1)?;
                Value::Bool(!values[0].to_bool())
            }
            "true" => Value::Bool(true),
            "false" => Value::Bool(false),
            "number" => {
                arity(0, 1)?;
                Value::Number(match values.first() {
                    Some(v) => v.to_number(document),
                    None => number_of(&document.string_value(self.node)),
                })
            }
            "sum" => {
                arity(1, 1)?;
                match &values[0] {
                    Value::Nodes(nodes) => {
                        Value::Number(nodes.iter().map(|&n| number_of(&document.string_value(n))).sum())
                    }
                    _ => return Err(Error("Invalid type".into())),
                }
            }
            "floor" => {
                arity(1, 1)?;
                Value::Number(values[0].to_number(document).floor())
            }
            "ceiling" => {
                arity(1, 1)?;
                Value::Number(values[0].to_number(document).ceil())
            }
            "round" => {
                arity(1, 1)?;
                let n = values[0].to_number(document);
                Value::Number(if n.is_finite() { (n + 0.5).floor() } else { n })
            }
            _ => return Err(Error("Unregistered function".into())),
        })
    }
}

/// evaluate evaluates a compiled expression at a node of a document, with the namespaces its prefixes name.
pub fn evaluate(
    expr: &Expr,
    document: &Document,
    node: usize,
    namespaces: &HashMap<String, String>,
) -> Result<Value, Error> {
    Context { document, namespaces, node, position: 1, size: 1 }.eval(expr)
}
