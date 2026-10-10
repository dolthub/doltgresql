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

//! The geometric functions and operators, ported from Postgres' geo_ops.c with its tolerance for comparisons and its
//! overflow checks on arithmetic.

use std::cmp::Ordering;

use super::Function;
use crate::error::{PgError, Result, code};
use crate::oid::{BOOL, FLOAT8, INT4};
use crate::types::{BaseValue, Value};

/// The geometric types' OIDs.
const POINT: u32 = 600;
const LSEG: u32 = 601;
const PATH: u32 = 602;
const BOX: u32 = 603;
const POLYGON: u32 = 604;
const LINE: u32 = 628;
const CIRCLE: u32 = 718;

/// EPSILON is the tolerance of geometric comparisons.
const EPSILON: f64 = 1.0e-06;

/// overflow returns Postgres' error for a float result too large to represent.
fn overflow() -> PgError {
    PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "value out of range: overflow")
}

/// underflow returns Postgres' error for a float result too small to represent.
fn underflow() -> PgError {
    PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "value out of range: underflow")
}

/// pl adds floats, failing on overflow as Postgres' float8_pl does.
fn pl(a: f64, b: f64) -> Result<f64> {
    let r = a + b;
    if r.is_infinite() && !a.is_infinite() && !b.is_infinite() {
        return Err(overflow());
    }
    Ok(r)
}

/// mi subtracts floats, failing on overflow as Postgres' float8_mi does.
fn mi(a: f64, b: f64) -> Result<f64> {
    let r = a - b;
    if r.is_infinite() && !a.is_infinite() && !b.is_infinite() {
        return Err(overflow());
    }
    Ok(r)
}

/// mul multiplies floats, failing on overflow and underflow as Postgres' float8_mul does.
fn mul(a: f64, b: f64) -> Result<f64> {
    let r = a * b;
    if r.is_infinite() && !a.is_infinite() && !b.is_infinite() {
        return Err(overflow());
    }
    if r == 0.0 && a != 0.0 && b != 0.0 {
        return Err(underflow());
    }
    Ok(r)
}

/// div divides floats, failing on division by zero, overflow, and underflow as Postgres' float8_div does.
fn div(a: f64, b: f64) -> Result<f64> {
    if b == 0.0 && !a.is_nan() {
        return Err(PgError::new(code::DIVISION_BY_ZERO, "division by zero"));
    }
    let r = a / b;
    if r.is_infinite() && !a.is_infinite() {
        return Err(overflow());
    }
    if r == 0.0 && a != 0.0 && !b.is_infinite() {
        return Err(underflow());
    }
    Ok(r)
}

/// hypot returns the length of a vector as Postgres' pg_hypot computes it.
fn hypot(x: f64, y: f64) -> Result<f64> {
    if x.is_infinite() || y.is_infinite() {
        return Ok(f64::INFINITY);
    }
    if x.is_nan() || y.is_nan() {
        return Ok(f64::NAN);
    }
    let (mut x, mut y) = (x.abs(), y.abs());
    if x < y {
        std::mem::swap(&mut x, &mut y);
    }
    if y == 0.0 {
        return Ok(x);
    }
    let yx = y / x;
    let result = x * (1.0 + yx * yx).sqrt();
    if result.is_infinite() {
        return Err(overflow());
    }
    if result == 0.0 {
        return Err(underflow());
    }
    Ok(result)
}

/// cmp orders floats as Postgres' float8 comparisons do, with NaN above everything and equal to itself.
fn cmp(a: f64, b: f64) -> Ordering {
    a.partial_cmp(&b).unwrap_or_else(|| a.is_nan().cmp(&b.is_nan()))
}

/// float_min and float_max return the smaller and larger float as Postgres' float8_min and float8_max do.
fn float_min(a: f64, b: f64) -> f64 {
    if cmp(a, b).is_lt() { a } else { b }
}

/// float_max returns the larger float, treating NaN as the largest.
fn float_max(a: f64, b: f64) -> f64 {
    if cmp(a, b).is_gt() { a } else { b }
}

/// fp_zero reports whether a float is zero within the tolerance.
fn fp_zero(a: f64) -> bool {
    a.abs() <= EPSILON
}

/// fp_eq reports whether floats are equal within the tolerance.
fn fp_eq(a: f64, b: f64) -> bool {
    a == b || (a - b).abs() <= EPSILON
}

/// fp_lt reports whether a float is less than another by more than the tolerance.
fn fp_lt(a: f64, b: f64) -> bool {
    a + EPSILON < b
}

/// fp_le reports whether a float is at most another, within the tolerance.
fn fp_le(a: f64, b: f64) -> bool {
    a <= b + EPSILON
}

/// fp_gt reports whether a float is greater than another by more than the tolerance.
fn fp_gt(a: f64, b: f64) -> bool {
    a > b + EPSILON
}

/// fp_ge reports whether a float is at least another, within the tolerance.
fn fp_ge(a: f64, b: f64) -> bool {
    a + EPSILON >= b
}

/// Point is a point.
#[derive(Clone, Copy, Debug)]
struct Point {
    x: f64,
    y: f64,
}

/// Lseg is a line segment.
#[derive(Clone, Copy, Debug)]
struct Lseg {
    p: [Point; 2],
}

/// Line is a line of the equation Ax + By + C = 0.
#[derive(Clone, Copy, Debug)]
struct Line {
    a: f64,
    b: f64,
    c: f64,
}

/// GeoBox is a box by its upper right and lower left corners.
#[derive(Clone, Copy, Debug)]
struct GeoBox {
    high: Point,
    low: Point,
}

/// Path is a path of points, open or closed.
#[derive(Clone, Debug)]
struct Path {
    closed: bool,
    p: Vec<Point>,
}

/// Polygon is a polygon of points with its bounding box.
#[derive(Clone, Debug)]
struct Polygon {
    p: Vec<Point>,
    bound: GeoBox,
}

/// Circle is a circle by its center and radius.
#[derive(Clone, Copy, Debug)]
struct Circle {
    center: Point,
    radius: f64,
}

/// floats reads a geometric value's floats, which it stores in network order.
fn floats(bytes: &[u8]) -> Vec<f64> {
    bytes.as_chunks::<8>().0.iter().map(|c| f64::from_be_bytes(*c)).collect()
}

/// data returns the stored bytes of a geometric value.
fn data(value: &Value) -> Result<&[u8]> {
    match value {
        Value::Base(base) => Ok(&base.data),
        _ => Err(PgError::internal("a geometric value that is not one")),
    }
}

/// points reads pairs of floats as points.
fn points(values: &[f64]) -> Vec<Point> {
    values.as_chunks::<2>().0.iter().map(|[x, y]| Point { x: *x, y: *y }).collect()
}

/// value returns a geometric value of a type from its floats, prefixed by any header bytes.
fn value(type_oid: u32, header: &[u8], values: &[f64]) -> Value {
    let mut data = header.to_vec();
    data.extend(values.iter().flat_map(|v| v.to_be_bytes()));
    Value::Base(Box::new(BaseValue { type_oid, data }))
}

/// zero_sign turns negative zero into zero, as Postgres does for computed coordinates.
fn zero_sign(a: f64) -> f64 {
    if a == 0.0 { 0.0 } else { a }
}

impl Point {
    /// read reads a point value.
    fn read(v: &Value) -> Result<Point> {
        let f = floats(data(v)?);
        Ok(Point { x: f[0], y: f[1] })
    }

    /// value returns the point as a value.
    fn value(self) -> Value {
        value(POINT, &[], &[self.x, self.y])
    }

    /// eq compares points within the tolerance, or exactly when NaN is involved, as Postgres' point_eq_point does.
    fn eq(&self, other: &Point) -> bool {
        if self.x.is_nan() || self.y.is_nan() || other.x.is_nan() || other.y.is_nan() {
            return cmp(self.x, other.x).is_eq() && cmp(self.y, other.y).is_eq();
        }
        fp_eq(self.x, other.x) && fp_eq(self.y, other.y)
    }

    /// dt returns the distance between points.
    fn dt(&self, other: &Point) -> Result<f64> {
        hypot(mi(self.x, other.x)?, mi(self.y, other.y)?)
    }

    /// sl returns the slope between points.
    fn sl(&self, other: &Point) -> Result<f64> {
        if fp_eq(self.x, other.x) {
            return Ok(f64::INFINITY);
        }
        if fp_eq(self.y, other.y) {
            return Ok(0.0);
        }
        div(mi(self.y, other.y)?, mi(self.x, other.x)?)
    }

    /// invsl returns the slope perpendicular to the one between points.
    fn invsl(&self, other: &Point) -> Result<f64> {
        if fp_eq(self.x, other.x) {
            return Ok(0.0);
        }
        if fp_eq(self.y, other.y) {
            return Ok(f64::INFINITY);
        }
        div(mi(self.x, other.x)?, mi(other.y, self.y)?)
    }

    /// add adds points as vectors.
    fn add(&self, other: &Point) -> Result<Point> {
        Ok(Point { x: pl(self.x, other.x)?, y: pl(self.y, other.y)? })
    }

    /// sub subtracts points as vectors.
    fn sub(&self, other: &Point) -> Result<Point> {
        Ok(Point { x: mi(self.x, other.x)?, y: mi(self.y, other.y)? })
    }

    /// mul multiplies points as complex numbers.
    fn mul(&self, other: &Point) -> Result<Point> {
        Ok(Point {
            x: mi(mul(self.x, other.x)?, mul(self.y, other.y)?)?,
            y: pl(mul(self.x, other.y)?, mul(self.y, other.x)?)?,
        })
    }

    /// div divides points as complex numbers.
    fn div(&self, other: &Point) -> Result<Point> {
        let d = pl(mul(other.x, other.x)?, mul(other.y, other.y)?)?;
        Ok(Point {
            x: div(pl(mul(self.x, other.x)?, mul(self.y, other.y)?)?, d)?,
            y: div(mi(mul(self.y, other.x)?, mul(self.x, other.y)?)?, d)?,
        })
    }
}

impl Lseg {
    /// read reads a line segment value.
    fn read(v: &Value) -> Result<Lseg> {
        let p = points(&floats(data(v)?));
        Ok(Lseg { p: [p[0], p[1]] })
    }

    /// value returns the segment as a value.
    fn value(self) -> Value {
        value(LSEG, &[], &[self.p[0].x, self.p[0].y, self.p[1].x, self.p[1].y])
    }

    /// sl returns the segment's slope.
    fn sl(&self) -> Result<f64> {
        self.p[0].sl(&self.p[1])
    }

    /// invsl returns the slope perpendicular to the segment.
    fn invsl(&self) -> Result<f64> {
        self.p[0].invsl(&self.p[1])
    }

    /// length returns the segment's length.
    fn length(&self) -> Result<f64> {
        self.p[0].dt(&self.p[1])
    }

    /// center returns the segment's midpoint.
    fn center(&self) -> Result<Point> {
        Ok(Point { x: div(pl(self.p[0].x, self.p[1].x)?, 2.0)?, y: div(pl(self.p[0].y, self.p[1].y)?, 2.0)? })
    }
}

impl Line {
    /// read reads a line value.
    fn read(v: &Value) -> Result<Line> {
        let f = floats(data(v)?);
        Ok(Line { a: f[0], b: f[1], c: f[2] })
    }

    /// value returns the line as a value.
    fn value(self) -> Value {
        value(LINE, &[], &[self.a, self.b, self.c])
    }

    /// construct returns the line through a point with a slope, as Postgres' line_construct does.
    fn construct(pt: &Point, m: f64) -> Result<Line> {
        if m.is_infinite() {
            return Ok(Line { a: -1.0, b: 0.0, c: pt.x });
        }
        if m == 0.0 {
            return Ok(Line { a: 0.0, b: -1.0, c: pt.y });
        }
        Ok(Line { a: m, b: -1.0, c: zero_sign(mi(pt.y, mul(m, pt.x)?)?) })
    }

    /// sl returns the line's slope.
    fn sl(&self) -> Result<f64> {
        if fp_zero(self.a) {
            return Ok(0.0);
        }
        if fp_zero(self.b) {
            return Ok(f64::INFINITY);
        }
        div(self.a, -self.b)
    }

    /// invsl returns the slope perpendicular to the line.
    fn invsl(&self) -> Result<f64> {
        if fp_zero(self.a) {
            return Ok(f64::INFINITY);
        }
        if fp_zero(self.b) {
            return Ok(0.0);
        }
        div(self.b, self.a)
    }

    /// interpt returns where lines cross, or None when they are parallel, as Postgres' line_interpt_line does.
    fn interpt(&self, other: &Line) -> Result<Option<Point>> {
        let (l1, l2) = (self, other);
        let (x, y) = if !fp_zero(l1.b) {
            if fp_eq(l2.a, mul(l1.a, div(l2.b, l1.b)?)?) {
                return Ok(None);
            }
            let x = div(mi(mul(l1.b, l2.c)?, mul(l2.b, l1.c)?)?, mi(mul(l1.a, l2.b)?, mul(l2.a, l1.b)?)?)?;
            (x, div(-pl(mul(l1.a, x)?, l1.c)?, l1.b)?)
        } else if !fp_zero(l2.b) {
            if fp_eq(l1.a, mul(l2.a, div(l1.b, l2.b)?)?) {
                return Ok(None);
            }
            let x = div(mi(mul(l2.b, l1.c)?, mul(l1.b, l2.c)?)?, mi(mul(l2.a, l1.b)?, mul(l1.a, l2.b)?)?)?;
            (x, div(-pl(mul(l2.a, x)?, l2.c)?, l2.b)?)
        } else {
            return Ok(None);
        };
        Ok(Some(Point { x: zero_sign(x), y: zero_sign(y) }))
    }
}

impl GeoBox {
    /// read reads a box value.
    fn read(v: &Value) -> Result<GeoBox> {
        let p = points(&floats(data(v)?));
        Ok(GeoBox { high: p[0], low: p[1] })
    }

    /// value returns the box as a value.
    fn value(self) -> Value {
        value(BOX, &[], &[self.high.x, self.high.y, self.low.x, self.low.y])
    }

    /// construct returns the box with two corners, as Postgres' box_construct does.
    fn construct(p1: &Point, p2: &Point) -> GeoBox {
        let (hx, lx) = if cmp(p1.x, p2.x).is_gt() { (p1.x, p2.x) } else { (p2.x, p1.x) };
        let (hy, ly) = if cmp(p1.y, p2.y).is_gt() { (p1.y, p2.y) } else { (p2.y, p1.y) };
        GeoBox { high: Point { x: hx, y: hy }, low: Point { x: lx, y: ly } }
    }

    /// ov reports whether boxes overlap.
    fn ov(&self, other: &GeoBox) -> bool {
        fp_le(self.low.x, other.high.x)
            && fp_le(other.low.x, self.high.x)
            && fp_le(self.low.y, other.high.y)
            && fp_le(other.low.y, self.high.y)
    }

    /// contains_box reports whether the box contains another.
    fn contains_box(&self, other: &GeoBox) -> bool {
        fp_ge(self.high.x, other.high.x)
            && fp_le(self.low.x, other.low.x)
            && fp_ge(self.high.y, other.high.y)
            && fp_le(self.low.y, other.low.y)
    }

    /// wd returns the box's width.
    fn wd(&self) -> Result<f64> {
        mi(self.high.x, self.low.x)
    }

    /// ht returns the box's height.
    fn ht(&self) -> Result<f64> {
        mi(self.high.y, self.low.y)
    }

    /// ar returns the box's area.
    fn ar(&self) -> Result<f64> {
        mul(self.wd()?, self.ht()?)
    }

    /// cn returns the box's center.
    fn cn(&self) -> Result<Point> {
        Ok(Point { x: div(pl(self.high.x, self.low.x)?, 2.0)?, y: div(pl(self.high.y, self.low.y)?, 2.0)? })
    }
}

impl Path {
    /// read reads a path value: whether it is closed, its point count, and its points.
    fn read(v: &Value) -> Result<Path> {
        let d = data(v)?;
        Ok(Path { closed: d[0] != 0, p: points(&floats(&d[5..])) })
    }

    /// value returns the path as a value.
    fn value(&self) -> Value {
        let mut header = vec![u8::from(self.closed)];
        header.extend((self.p.len() as i32).to_be_bytes());
        value(PATH, &header, &self.p.iter().flat_map(|p| [p.x, p.y]).collect::<Vec<_>>())
    }

    /// segments returns the path's segments, including the closing one of a closed path, as Postgres' loops over
    /// `iprev` visit them.
    fn segments(&self) -> Vec<Lseg> {
        let n = self.p.len();
        (0..n)
            .filter_map(|i| match i {
                0 if !self.closed => None,
                0 => Some(Lseg { p: [self.p[n - 1], self.p[0]] }),
                _ => Some(Lseg { p: [self.p[i - 1], self.p[i]] }),
            })
            .collect()
    }
}

/// bounding_box returns the box around points, as Postgres' path_inter and make_bound_box compute it.
fn bounding_box(points: &[Point]) -> GeoBox {
    let first = points[0];
    let mut b = GeoBox { high: first, low: first };
    for p in &points[1..] {
        b.high.x = float_max(p.x, b.high.x);
        b.high.y = float_max(p.y, b.high.y);
        b.low.x = float_min(p.x, b.low.x);
        b.low.y = float_min(p.y, b.low.y);
    }
    b
}

impl Polygon {
    /// read reads a polygon value: its point count and points.
    fn read(v: &Value) -> Result<Polygon> {
        Ok(Polygon::new(points(&floats(&data(v)?[4..]))))
    }

    /// new returns a polygon of points with its bounding box, as Postgres' make_bound_box computes it.
    fn new(p: Vec<Point>) -> Polygon {
        let (mut x1, mut y1) = (p[0].x, p[0].y);
        let (mut x2, mut y2) = (x1, y1);
        for q in &p[1..] {
            if cmp(q.x, x1).is_lt() {
                x1 = q.x;
            }
            if cmp(q.x, x2).is_gt() {
                x2 = q.x;
            }
            if cmp(q.y, y1).is_lt() {
                y1 = q.y;
            }
            if cmp(q.y, y2).is_gt() {
                y2 = q.y;
            }
        }
        let bound = GeoBox { high: Point { x: x2, y: y2 }, low: Point { x: x1, y: y1 } };
        Polygon { p, bound }
    }

    /// value returns the polygon as a value.
    fn value(&self) -> Value {
        let header = (self.p.len() as i32).to_be_bytes();
        value(POLYGON, &header, &self.p.iter().flat_map(|p| [p.x, p.y]).collect::<Vec<_>>())
    }

    /// edges returns the polygon's edges, starting with the one from the last point to the first.
    fn edges(&self) -> Vec<Lseg> {
        let n = self.p.len();
        (0..n).map(|i| Lseg { p: [self.p[if i == 0 { n - 1 } else { i - 1 }], self.p[i]] }).collect()
    }

    /// to_circle returns the circle around the polygon's points, as Postgres' poly_to_circle does.
    fn to_circle(&self) -> Result<Circle> {
        let n = self.p.len() as f64;
        let mut center = Point { x: 0.0, y: 0.0 };
        for p in &self.p {
            center = center.add(p)?;
        }
        center = Point { x: div(center.x, n)?, y: div(center.y, n)? };
        let mut radius = 0.0;
        for p in &self.p {
            radius = pl(radius, p.dt(&center)?)?;
        }
        Ok(Circle { center, radius: div(radius, n)? })
    }
}

impl Circle {
    /// read reads a circle value.
    fn read(v: &Value) -> Result<Circle> {
        let f = floats(data(v)?);
        Ok(Circle { center: Point { x: f[0], y: f[1] }, radius: f[2] })
    }

    /// value returns the circle as a value.
    fn value(self) -> Value {
        value(CIRCLE, &[], &[self.center.x, self.center.y, self.radius])
    }

    /// ar returns the circle's area.
    fn ar(&self) -> Result<f64> {
        mul(mul(self.radius, self.radius)?, std::f64::consts::PI)
    }
}

/// lseg_contain_point reports whether a point lies on a segment.
fn lseg_contain_point(lseg: &Lseg, pt: &Point) -> Result<bool> {
    Ok(fp_eq(pt.dt(&lseg.p[0])? + pt.dt(&lseg.p[1])?, lseg.p[0].dt(&lseg.p[1])?))
}

/// line_contain_point reports whether a point lies on a line.
fn line_contain_point(line: &Line, pt: &Point) -> Result<bool> {
    Ok(fp_zero(pl(pl(mul(line.a, pt.x)?, mul(line.b, pt.y)?)?, line.c)?))
}

/// box_contain_point reports whether a point lies in a box.
fn box_contain_point(b: &GeoBox, pt: &Point) -> bool {
    b.high.x >= pt.x && b.low.x <= pt.x && b.high.y >= pt.y && b.low.y <= pt.y
}

/// lseg_interpt_line returns where a segment meets a line, preferring the segment's own end points.
fn lseg_interpt_line(lseg: &Lseg, line: &Line) -> Result<Option<Point>> {
    let tmp = Line::construct(&lseg.p[0], lseg.sl()?)?;
    let Some(interpt) = tmp.interpt(line)? else { return Ok(None) };
    if !lseg_contain_point(lseg, &interpt)? {
        return Ok(None);
    }
    Ok(Some(if lseg.p[0].eq(&interpt) {
        lseg.p[0]
    } else if lseg.p[1].eq(&interpt) {
        lseg.p[1]
    } else {
        interpt
    }))
}

/// lseg_interpt_lseg returns where segments cross.
fn lseg_interpt_lseg(l1: &Lseg, l2: &Lseg) -> Result<Option<Point>> {
    let tmp = Line::construct(&l2.p[0], l2.sl()?)?;
    let Some(interpt) = lseg_interpt_line(l1, &tmp)? else { return Ok(None) };
    Ok(lseg_contain_point(l2, &interpt)?.then_some(interpt))
}

/// line_closept_point returns the point of a line closest to a point and their distance, which is NaN when the
/// perpendicular cannot be found.
fn line_closept_point(line: &Line, pt: &Point) -> Result<(Point, f64)> {
    let tmp = Line::construct(pt, line.invsl()?)?;
    match tmp.interpt(line)? {
        Some(closept) => Ok((closept, closept.dt(pt)?)),
        None => Ok((*pt, f64::NAN)),
    }
}

/// lseg_closept_line returns the point of a segment closest to a line and their distance.
fn lseg_closept_line(lseg: &Lseg, line: &Line) -> Result<(Point, f64)> {
    if let Some(p) = lseg_interpt_line(lseg, line)? {
        return Ok((p, 0.0));
    }
    let d1 = line_closept_point(line, &lseg.p[0])?.1;
    let d2 = line_closept_point(line, &lseg.p[1])?.1;
    Ok(if d1 < d2 { (lseg.p[0], d1) } else { (lseg.p[1], d2) })
}

/// lseg_closept_point returns the point of a segment closest to a point and their distance.
fn lseg_closept_point(lseg: &Lseg, pt: &Point) -> Result<(Point, f64)> {
    let tmp = Line::construct(pt, lseg.p[0].invsl(&lseg.p[1])?)?;
    let closept = lseg_closept_line(lseg, &tmp)?.0;
    Ok((closept, closept.dt(pt)?))
}

/// lseg_closept_lseg returns the point of one segment closest to another and their distance.
fn lseg_closept_lseg(on: &Lseg, to: &Lseg) -> Result<(Point, f64)> {
    if let Some(p) = lseg_interpt_lseg(on, to)? {
        return Ok((p, 0.0));
    }
    let (mut result, mut dist) = lseg_closept_point(on, &to.p[0])?;
    let (point, d) = lseg_closept_point(on, &to.p[1])?;
    if cmp(d, dist).is_lt() {
        (result, dist) = (point, d);
    }
    for end in on.p {
        let d = lseg_closept_point(to, &end)?.1;
        if cmp(d, dist).is_lt() {
            (result, dist) = (end, d);
        }
    }
    Ok((result, dist))
}

/// box_sides returns a box's sides in the order Postgres' box_closept functions visit them.
fn box_sides(b: &GeoBox) -> [Lseg; 4] {
    let upper_left = Point { x: b.low.x, y: b.high.y };
    let lower_right = Point { x: b.high.x, y: b.low.y };
    [
        Lseg { p: [b.low, upper_left] },
        Lseg { p: [b.high, upper_left] },
        Lseg { p: [b.low, lower_right] },
        Lseg { p: [b.high, lower_right] },
    ]
}

/// box_closept_point returns the point of a box closest to a point and their distance.
fn box_closept_point(b: &GeoBox, pt: &Point) -> Result<(Point, f64)> {
    if box_contain_point(b, pt) {
        return Ok((*pt, 0.0));
    }
    let sides = box_sides(b);
    let (mut result, mut dist) = lseg_closept_point(&sides[0], pt)?;
    for side in &sides[1..] {
        let (closept, d) = lseg_closept_point(side, pt)?;
        if cmp(d, dist).is_lt() {
            (result, dist) = (closept, d);
        }
    }
    Ok((result, dist))
}

/// box_interpt_lseg reports whether a segment meets a box, with the point of the segment closest to the box's center.
fn box_interpt_lseg(b: &GeoBox, lseg: &Lseg) -> Result<Option<Point>> {
    let lbox = GeoBox {
        low: Point { x: float_min(lseg.p[0].x, lseg.p[1].x), y: float_min(lseg.p[0].y, lseg.p[1].y) },
        high: Point { x: float_max(lseg.p[0].x, lseg.p[1].x), y: float_max(lseg.p[0].y, lseg.p[1].y) },
    };
    if !lbox.ov(b) {
        return Ok(None);
    }
    let closest = lseg_closept_point(lseg, &b.cn()?)?.0;
    if box_contain_point(b, &lseg.p[0]) || box_contain_point(b, &lseg.p[1]) {
        return Ok(Some(closest));
    }
    for side in box_sides(b) {
        if lseg_interpt_lseg(&side, lseg)?.is_some() {
            return Ok(Some(closest));
        }
    }
    Ok(None)
}

/// box_closept_lseg returns the point of a box closest to a segment and their distance.
fn box_closept_lseg(b: &GeoBox, lseg: &Lseg) -> Result<(Point, f64)> {
    if let Some(p) = box_interpt_lseg(b, lseg)? {
        return Ok((p, 0.0));
    }
    let sides = box_sides(b);
    let (mut result, mut dist) = lseg_closept_lseg(&sides[0], lseg)?;
    for side in &sides[1..] {
        let (closept, d) = lseg_closept_lseg(side, lseg)?;
        if cmp(d, dist).is_lt() {
            (result, dist) = (closept, d);
        }
    }
    Ok((result, dist))
}

/// lseg_crossing returns how a polygon edge crosses the positive x axis from a point, as Postgres' lseg_crossing
/// does, or None when the point lies on the edge.
fn lseg_crossing(x: f64, y: f64, prev_x: f64, prev_y: f64) -> Result<Option<i32>> {
    if fp_zero(y) {
        if fp_zero(x) {
            return Ok(None);
        }
        if fp_gt(x, 0.0) {
            if fp_zero(prev_y) {
                return Ok(if fp_gt(prev_x, 0.0) { Some(0) } else { None });
            }
            return Ok(Some(if fp_lt(prev_y, 0.0) { 1 } else { -1 }));
        }
        if fp_zero(prev_y) {
            return Ok(if fp_lt(prev_x, 0.0) { Some(0) } else { None });
        }
        return Ok(Some(0));
    }
    let y_sign = if fp_gt(y, 0.0) { 1 } else { -1 };
    if fp_zero(prev_y) {
        return Ok(Some(if fp_lt(prev_x, 0.0) { 0 } else { y_sign }));
    }
    if (y_sign < 0 && fp_lt(prev_y, 0.0)) || (y_sign > 0 && fp_gt(prev_y, 0.0)) {
        return Ok(Some(0));
    }
    if fp_ge(x, 0.0) && fp_gt(prev_x, 0.0) {
        return Ok(Some(2 * y_sign));
    }
    if fp_lt(x, 0.0) && fp_le(prev_x, 0.0) {
        return Ok(Some(0));
    }
    let z = mi(mul(mi(x, prev_x)?, y)?, mul(mi(y, prev_y)?, x)?)?;
    if fp_zero(z) {
        return Ok(None);
    }
    if (y_sign < 0 && fp_lt(z, 0.0)) || (y_sign > 0 && fp_gt(z, 0.0)) {
        return Ok(Some(0));
    }
    Ok(Some(2 * y_sign))
}

/// point_inside reports whether a point lies inside or on a polygon's outline, as Postgres' point_inside does.
fn point_inside(p: &Point, list: &[Point]) -> Result<bool> {
    let (x0, y0) = (mi(list[0].x, p.x)?, mi(list[0].y, p.y)?);
    let (mut prev_x, mut prev_y) = (x0, y0);
    let mut total = 0;
    for q in &list[1..] {
        let (x, y) = (mi(q.x, p.x)?, mi(q.y, p.y)?);
        let Some(cross) = lseg_crossing(x, y, prev_x, prev_y)? else { return Ok(true) };
        total += cross;
        (prev_x, prev_y) = (x, y);
    }
    let Some(cross) = lseg_crossing(x0, y0, prev_x, prev_y)? else { return Ok(true) };
    Ok(total + cross != 0)
}

/// plist_same reports whether point lists are the same cycle of points, in either direction.
fn plist_same(p1: &[Point], p2: &[Point]) -> bool {
    let n = p1.len();
    (0..n).any(|i| {
        p2[i].eq(&p1[0])
            && ((1..n).all(|k| p2[(i + k) % n].eq(&p1[k])) || (1..n).all(|k| p2[(i + n - k) % n].eq(&p1[k])))
    })
}

/// poly_overlap reports whether polygons overlap: their edges cross, or one holds the other's first point.
fn poly_overlap(a: &Polygon, b: &Polygon) -> Result<bool> {
    if !a.bound.ov(&b.bound) {
        return Ok(false);
    }
    for sa in a.edges() {
        for sb in b.edges() {
            if lseg_interpt_lseg(&sa, &sb)?.is_some() {
                return Ok(true);
            }
        }
    }
    Ok(point_inside(&a.p[0], &b.p)? || point_inside(&b.p[0], &a.p)?)
}

/// touched_lseg_inside_poly decides whether a segment touching a polygon edge stays inside, as Postgres'
/// touched_lseg_inside_poly does.
fn touched_lseg_inside_poly(a: &Point, b: &Point, s: &Lseg, poly: &Polygon, start: usize) -> Result<bool> {
    let t = Lseg { p: [*a, *b] };
    if a.eq(&s.p[0]) {
        if lseg_contain_point(&t, &s.p[1])? {
            return lseg_inside_poly(b, &s.p[1], poly, start);
        }
    } else if a.eq(&s.p[1]) {
        if lseg_contain_point(&t, &s.p[0])? {
            return lseg_inside_poly(b, &s.p[0], poly, start);
        }
    } else if lseg_contain_point(&t, &s.p[0])? {
        return lseg_inside_poly(b, &s.p[0], poly, start);
    } else if lseg_contain_point(&t, &s.p[1])? {
        return lseg_inside_poly(b, &s.p[1], poly, start);
    }
    Ok(true)
}

/// lseg_inside_poly reports whether a segment lies inside a polygon, checking the edges from `start`, as Postgres'
/// lseg_inside_poly does.
fn lseg_inside_poly(a: &Point, b: &Point, poly: &Polygon, start: usize) -> Result<bool> {
    let t = Lseg { p: [*a, *b] };
    let n = poly.p.len();
    let mut s = Lseg { p: [poly.p[if start == 0 { n - 1 } else { start - 1 }], poly.p[0]] };
    let (mut res, mut intersection) = (true, false);
    let mut i = start;
    while i < n && res {
        s.p[1] = poly.p[i];
        if lseg_contain_point(&s, &t.p[0])? {
            if lseg_contain_point(&s, &t.p[1])? {
                return Ok(true);
            }
            res = touched_lseg_inside_poly(&t.p[0], &t.p[1], &s, poly, i + 1)?;
        } else if lseg_contain_point(&s, &t.p[1])? {
            res = touched_lseg_inside_poly(&t.p[1], &t.p[0], &s, poly, i + 1)?;
        } else if let Some(interpt) = lseg_interpt_lseg(&t, &s)? {
            intersection = true;
            res = lseg_inside_poly(&t.p[0], &interpt, poly, i + 1)?;
            if res {
                res = lseg_inside_poly(&t.p[1], &interpt, poly, i + 1)?;
            }
        }
        s.p[0] = s.p[1];
        i += 1;
    }
    if res && !intersection {
        let p = Point { x: div(pl(t.p[0].x, t.p[1].x)?, 2.0)?, y: div(pl(t.p[0].y, t.p[1].y)?, 2.0)? };
        res = point_inside(&p, &poly.p)?;
    }
    Ok(res)
}

/// poly_contain_poly reports whether one polygon contains another.
fn poly_contain_poly(outer: &Polygon, inner: &Polygon) -> Result<bool> {
    if !outer.bound.contains_box(&inner.bound) {
        return Ok(false);
    }
    for s in inner.edges() {
        if !lseg_inside_poly(&s.p[0], &s.p[1], outer, 0)? {
            return Ok(false);
        }
    }
    Ok(true)
}

/// dist_ppoly returns the distance from a point to a polygon, zero inside it.
fn dist_ppoly(pt: &Point, poly: &Polygon) -> Result<f64> {
    if point_inside(pt, &poly.p)? {
        return Ok(0.0);
    }
    let n = poly.p.len();
    let mut result = lseg_closept_point(&Lseg { p: [poly.p[0], poly.p[n - 1]] }, pt)?.1;
    for i in 0..n - 1 {
        let d = lseg_closept_point(&Lseg { p: [poly.p[i], poly.p[i + 1]] }, pt)?.1;
        if cmp(d, result).is_lt() {
            result = d;
        }
    }
    Ok(result)
}

/// minimum returns the smallest of distances, or None without any.
fn minimum(distances: impl IntoIterator<Item = Result<f64>>) -> Result<Option<f64>> {
    let mut min: Option<f64> = None;
    for d in distances {
        let d = d?;
        if min.is_none_or(|m| cmp(d, m).is_lt()) {
            min = Some(d);
        }
    }
    Ok(min)
}

/// dist_ppath returns the distance from a point to a path.
fn dist_ppath(pt: &Point, path: &Path) -> Result<f64> {
    Ok(minimum(path.segments().iter().map(|s| lseg_closept_point(s, pt).map(|r| r.1)))?.unwrap_or(0.0))
}

/// on_ppath reports whether a point lies on an open path, or inside a closed one.
fn on_ppath(pt: &Point, path: &Path) -> Result<bool> {
    if path.closed {
        return point_inside(pt, &path.p);
    }
    let mut a = pt.dt(&path.p[0])?;
    for i in 0..path.p.len() - 1 {
        let b = pt.dt(&path.p[i + 1])?;
        if fp_eq(pl(a, b)?, path.p[i].dt(&path.p[i + 1])?) {
            return Ok(true);
        }
        a = b;
    }
    Ok(false)
}

/// line_eq reports whether lines are the same, as Postgres' line_eq does.
fn line_eq(l1: &Line, l2: &Line) -> Result<bool> {
    if [l1.a, l1.b, l1.c, l2.a, l2.b, l2.c].iter().any(|f| f.is_nan()) {
        return Ok(cmp(l1.a, l2.a).is_eq() && cmp(l1.b, l2.b).is_eq() && cmp(l1.c, l2.c).is_eq());
    }
    let ratio = if !fp_zero(l2.a) {
        div(l1.a, l2.a)?
    } else if !fp_zero(l2.b) {
        div(l1.b, l2.b)?
    } else if !fp_zero(l2.c) {
        div(l1.c, l2.c)?
    } else {
        1.0
    };
    Ok(fp_eq(l1.a, mul(ratio, l2.a)?) && fp_eq(l1.b, mul(ratio, l2.b)?) && fp_eq(l1.c, mul(ratio, l2.c)?))
}

/// line_perp reports whether lines are perpendicular.
fn line_perp(l1: &Line, l2: &Line) -> Result<bool> {
    if fp_zero(l1.a) {
        return Ok(fp_zero(l2.b));
    }
    if fp_zero(l2.a) {
        return Ok(fp_zero(l1.b));
    }
    if fp_zero(l1.b) {
        return Ok(fp_zero(l2.a));
    }
    if fp_zero(l2.b) {
        return Ok(fp_zero(l1.a));
    }
    Ok(fp_eq(div(mul(l1.a, l2.a)?, mul(l1.b, l2.b)?)?, -1.0))
}

/// line_distance returns the distance between parallel lines, or zero for crossing ones.
fn line_distance(l1: &Line, l2: &Line) -> Result<f64> {
    if l1.interpt(l2)?.is_some() {
        return Ok(0.0);
    }
    let usable = |a: f64| !fp_zero(a) && !a.is_nan();
    let ratio = if usable(l1.a) && usable(l2.a) {
        div(l1.a, l2.a)?
    } else if usable(l1.b) && usable(l2.b) {
        div(l1.b, l2.b)?
    } else {
        1.0
    };
    div(mi(l1.c, mul(ratio, l2.c)?)?.abs(), hypot(l1.a, l1.b)?)
}

/// path_inter reports whether paths cross.
fn path_inter(p1: &Path, p2: &Path) -> Result<bool> {
    if !bounding_box(&p1.p).ov(&bounding_box(&p2.p)) {
        return Ok(false);
    }
    for s1 in p1.segments() {
        for s2 in p2.segments() {
            if lseg_interpt_lseg(&s1, &s2)?.is_some() {
                return Ok(true);
            }
        }
    }
    Ok(false)
}

/// circle_poly returns a polygon of points spaced evenly around a circle.
fn circle_poly(npts: i32, circle: &Circle) -> Result<Polygon> {
    if fp_zero(circle.radius) {
        return Err(PgError::new(code::FEATURE_NOT_SUPPORTED, "cannot convert circle with radius zero to polygon"));
    }
    if npts < 2 {
        return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "must request at least 2 points"));
    }
    let step = div(2.0 * std::f64::consts::PI, f64::from(npts))?;
    let mut p = Vec::with_capacity(npts as usize);
    for i in 0..npts {
        let angle = mul(step, f64::from(i))?;
        p.push(Point {
            x: mi(circle.center.x, mul(circle.radius, angle.cos())?)?,
            y: pl(circle.center.y, mul(circle.radius, angle.sin())?)?,
        });
    }
    Ok(Polygon::new(p))
}

/// map_path applies a point operation to each point of a path.
fn map_path(path: &Value, pt: &Value, op: fn(&Point, &Point) -> Result<Point>) -> Result<Value> {
    let (mut path, pt) = (Path::read(path)?, Point::read(pt)?);
    for p in &mut path.p {
        *p = op(p, &pt)?;
    }
    Ok(path.value())
}

/// optional returns a point value, or NULL when the distance that found it is NaN.
fn optional((point, distance): (Point, f64)) -> Value {
    if distance.is_nan() { Value::Null } else { point.value() }
}

/// nonnegative clamps a negative distance to zero.
fn nonnegative(d: f64) -> f64 {
    if d < 0.0 { 0.0 } else { d }
}

/// boolean wraps a result as a bool value.
fn boolean(b: Result<bool>) -> Result<Value> {
    b.map(Value::Bool)
}

/// float wraps a result as a float8 value.
fn float(f: Result<f64>) -> Result<Value> {
    f.map(Value::Float8)
}

/// f declares a strict geometric function.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}
/// FUNCTIONS are the geometric functions and operators.
pub const FUNCTIONS: &[Function] = &[
    f(ELEMENT, &[POINT, INT4], FLOAT8, |_, a| element(a)),
    f(ELEMENT, &[LINE, INT4], FLOAT8, |_, a| element(a)),
    f(ELEMENT, &[LSEG, INT4], POINT, |_, a| element(a)),
    f(ELEMENT, &[BOX, INT4], POINT, |_, a| element(a)),
    f("box_same", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(b1.high.eq(&b2.high) && b1.low.eq(&b2.low)))
    }),
    f("~=", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(b1.high.eq(&b2.high) && b1.low.eq(&b2.low)))
    }),
    f("box_overlap", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(b1.ov(&b2)))
    }),
    f("&&", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(b1.ov(&b2)))
    }),
    f("?#", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(b1.ov(&b2)))
    }),
    f("box_left", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_lt(b1.high.x, b2.low.x)))
    }),
    f("<<", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_lt(b1.high.x, b2.low.x)))
    }),
    f("box_overleft", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_le(b1.high.x, b2.high.x)))
    }),
    f("&<", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_le(b1.high.x, b2.high.x)))
    }),
    f("box_right", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_gt(b1.low.x, b2.high.x)))
    }),
    f(">>", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_gt(b1.low.x, b2.high.x)))
    }),
    f("box_overright", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_ge(b1.low.x, b2.low.x)))
    }),
    f("&>", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_ge(b1.low.x, b2.low.x)))
    }),
    f("box_below", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_lt(b1.high.y, b2.low.y)))
    }),
    f("<<|", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_lt(b1.high.y, b2.low.y)))
    }),
    f("box_overbelow", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_le(b1.high.y, b2.high.y)))
    }),
    f("&<|", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_le(b1.high.y, b2.high.y)))
    }),
    f("box_above", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_gt(b1.low.y, b2.high.y)))
    }),
    f("|>>", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_gt(b1.low.y, b2.high.y)))
    }),
    f("box_overabove", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_ge(b1.low.y, b2.low.y)))
    }),
    f("|&>", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_ge(b1.low.y, b2.low.y)))
    }),
    f("box_contained", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(b2.contains_box(&b1)))
    }),
    f("<@", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(b2.contains_box(&b1)))
    }),
    f("box_contain", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(b1.contains_box(&b2)))
    }),
    f("@>", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(b1.contains_box(&b2)))
    }),
    f("box_below_eq", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_le(b1.high.y, b2.low.y)))
    }),
    f("<^", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_le(b1.high.y, b2.low.y)))
    }),
    f("box_above_eq", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_ge(b1.low.y, b2.high.y)))
    }),
    f(">^", &[BOX, BOX], BOOL, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(fp_ge(b1.low.y, b2.high.y)))
    }),
    f("box_lt", &[BOX, BOX], BOOL, |_, a| {
        Ok(Value::Bool(fp_lt(GeoBox::read(&a[0])?.ar()?, GeoBox::read(&a[1])?.ar()?)))
    }),
    f("<", &[BOX, BOX], BOOL, |_, a| Ok(Value::Bool(fp_lt(GeoBox::read(&a[0])?.ar()?, GeoBox::read(&a[1])?.ar()?)))),
    f("box_gt", &[BOX, BOX], BOOL, |_, a| {
        Ok(Value::Bool(fp_gt(GeoBox::read(&a[0])?.ar()?, GeoBox::read(&a[1])?.ar()?)))
    }),
    f(">", &[BOX, BOX], BOOL, |_, a| Ok(Value::Bool(fp_gt(GeoBox::read(&a[0])?.ar()?, GeoBox::read(&a[1])?.ar()?)))),
    f("box_eq", &[BOX, BOX], BOOL, |_, a| {
        Ok(Value::Bool(fp_eq(GeoBox::read(&a[0])?.ar()?, GeoBox::read(&a[1])?.ar()?)))
    }),
    f("=", &[BOX, BOX], BOOL, |_, a| Ok(Value::Bool(fp_eq(GeoBox::read(&a[0])?.ar()?, GeoBox::read(&a[1])?.ar()?)))),
    f("box_le", &[BOX, BOX], BOOL, |_, a| {
        Ok(Value::Bool(fp_le(GeoBox::read(&a[0])?.ar()?, GeoBox::read(&a[1])?.ar()?)))
    }),
    f("<=", &[BOX, BOX], BOOL, |_, a| Ok(Value::Bool(fp_le(GeoBox::read(&a[0])?.ar()?, GeoBox::read(&a[1])?.ar()?)))),
    f("box_ge", &[BOX, BOX], BOOL, |_, a| {
        Ok(Value::Bool(fp_ge(GeoBox::read(&a[0])?.ar()?, GeoBox::read(&a[1])?.ar()?)))
    }),
    f(">=", &[BOX, BOX], BOOL, |_, a| Ok(Value::Bool(fp_ge(GeoBox::read(&a[0])?.ar()?, GeoBox::read(&a[1])?.ar()?)))),
    f("box_area", &[BOX], FLOAT8, |_, a| float(GeoBox::read(&a[0])?.ar())),
    f("area", &[BOX], FLOAT8, |_, a| float(GeoBox::read(&a[0])?.ar())),
    f("box_width", &[BOX], FLOAT8, |_, a| float(GeoBox::read(&a[0])?.wd())),
    f("width", &[BOX], FLOAT8, |_, a| float(GeoBox::read(&a[0])?.wd())),
    f("box_height", &[BOX], FLOAT8, |_, a| float(GeoBox::read(&a[0])?.ht())),
    f("height", &[BOX], FLOAT8, |_, a| float(GeoBox::read(&a[0])?.ht())),
    f("box_distance", &[BOX, BOX], FLOAT8, |_, a| float(GeoBox::read(&a[0])?.cn()?.dt(&GeoBox::read(&a[1])?.cn()?))),
    f("<->", &[BOX, BOX], FLOAT8, |_, a| float(GeoBox::read(&a[0])?.cn()?.dt(&GeoBox::read(&a[1])?.cn()?))),
    f("box_center", &[BOX], POINT, |_, a| Ok(GeoBox::read(&a[0])?.cn()?.value())),
    f("center", &[BOX], POINT, |_, a| Ok(GeoBox::read(&a[0])?.cn()?.value())),
    f("point", &[BOX], POINT, |_, a| point_of_box(a)),
    f("@@", &[BOX], POINT, |_, a| Ok(GeoBox::read(&a[0])?.cn()?.value())),
    f("box_intersect", &[BOX, BOX], BOX, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        if !b1.ov(&b2) {
            return Ok(Value::Null);
        }
        Ok(GeoBox {
            high: Point { x: float_min(b1.high.x, b2.high.x), y: float_min(b1.high.y, b2.high.y) },
            low: Point { x: float_max(b1.low.x, b2.low.x), y: float_max(b1.low.y, b2.low.y) },
        }
        .value())
    }),
    f("#", &[BOX, BOX], BOX, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        if !b1.ov(&b2) {
            return Ok(Value::Null);
        }
        Ok(GeoBox {
            high: Point { x: float_min(b1.high.x, b2.high.x), y: float_min(b1.high.y, b2.high.y) },
            low: Point { x: float_max(b1.low.x, b2.low.x), y: float_max(b1.low.y, b2.low.y) },
        }
        .value())
    }),
    f("box_diagonal", &[BOX], LSEG, |_, a| {
        let b = GeoBox::read(&a[0])?;
        Ok(Lseg { p: [b.high, b.low] }.value())
    }),
    f("diagonal", &[BOX], LSEG, |_, a| {
        let b = GeoBox::read(&a[0])?;
        Ok(Lseg { p: [b.high, b.low] }.value())
    }),
    f("lseg", &[BOX], LSEG, |_, a| lseg_of_box(a)),
    f("bound_box", &[BOX, BOX], BOX, |_, a| {
        let (b1, b2) = (GeoBox::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(GeoBox {
            high: Point { x: float_max(b1.high.x, b2.high.x), y: float_max(b1.high.y, b2.high.y) },
            low: Point { x: float_min(b1.low.x, b2.low.x), y: float_min(b1.low.y, b2.low.y) },
        }
        .value())
    }),
    f("box_add", &[BOX, POINT], BOX, |_, a| {
        let (b, p) = (GeoBox::read(&a[0])?, Point::read(&a[1])?);
        Ok(GeoBox { high: b.high.add(&p)?, low: b.low.add(&p)? }.value())
    }),
    f("+", &[BOX, POINT], BOX, |_, a| {
        let (b, p) = (GeoBox::read(&a[0])?, Point::read(&a[1])?);
        Ok(GeoBox { high: b.high.add(&p)?, low: b.low.add(&p)? }.value())
    }),
    f("box_sub", &[BOX, POINT], BOX, |_, a| {
        let (b, p) = (GeoBox::read(&a[0])?, Point::read(&a[1])?);
        Ok(GeoBox { high: b.high.sub(&p)?, low: b.low.sub(&p)? }.value())
    }),
    f("-", &[BOX, POINT], BOX, |_, a| {
        let (b, p) = (GeoBox::read(&a[0])?, Point::read(&a[1])?);
        Ok(GeoBox { high: b.high.sub(&p)?, low: b.low.sub(&p)? }.value())
    }),
    f("box_mul", &[BOX, POINT], BOX, |_, a| {
        let (b, p) = (GeoBox::read(&a[0])?, Point::read(&a[1])?);
        Ok(GeoBox::construct(&b.high.mul(&p)?, &b.low.mul(&p)?).value())
    }),
    f("*", &[BOX, POINT], BOX, |_, a| {
        let (b, p) = (GeoBox::read(&a[0])?, Point::read(&a[1])?);
        Ok(GeoBox::construct(&b.high.mul(&p)?, &b.low.mul(&p)?).value())
    }),
    f("box_div", &[BOX, POINT], BOX, |_, a| {
        let (b, p) = (GeoBox::read(&a[0])?, Point::read(&a[1])?);
        Ok(GeoBox::construct(&b.high.div(&p)?, &b.low.div(&p)?).value())
    }),
    f("/", &[BOX, POINT], BOX, |_, a| {
        let (b, p) = (GeoBox::read(&a[0])?, Point::read(&a[1])?);
        Ok(GeoBox::construct(&b.high.div(&p)?, &b.low.div(&p)?).value())
    }),
    f("box", &[POINT, POINT], BOX, |_, a| Ok(GeoBox::construct(&Point::read(&a[0])?, &Point::read(&a[1])?).value())),
    f("box", &[POINT], BOX, |_, a| box_of_point(a)),
    f("box", &[POLYGON], BOX, |_, a| box_of_polygon(a)),
    f("box", &[CIRCLE], BOX, |_, a| box_of_circle(a)),
    f("point_left", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(fp_lt(p1.x, p2.x)))
    }),
    f("<<", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(fp_lt(p1.x, p2.x)))
    }),
    f("point_right", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(fp_gt(p1.x, p2.x)))
    }),
    f(">>", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(fp_gt(p1.x, p2.x)))
    }),
    f("point_above", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(fp_gt(p1.y, p2.y)))
    }),
    f(">^", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(fp_gt(p1.y, p2.y)))
    }),
    f("|>>", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(fp_gt(p1.y, p2.y)))
    }),
    f("point_below", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(fp_lt(p1.y, p2.y)))
    }),
    f("<^", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(fp_lt(p1.y, p2.y)))
    }),
    f("<<|", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(fp_lt(p1.y, p2.y)))
    }),
    f("point_vert", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(fp_eq(p1.x, p2.x)))
    }),
    f("isvertical", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(fp_eq(p1.x, p2.x)))
    }),
    f("?|", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(fp_eq(p1.x, p2.x)))
    }),
    f("point_horiz", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(fp_eq(p1.y, p2.y)))
    }),
    f("ishorizontal", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(fp_eq(p1.y, p2.y)))
    }),
    f("?-", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(fp_eq(p1.y, p2.y)))
    }),
    f("point_eq", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(p1.eq(&p2)))
    }),
    f("~=", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(p1.eq(&p2)))
    }),
    f("point_ne", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(!p1.eq(&p2)))
    }),
    f("<>", &[POINT, POINT], BOOL, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        Ok(Value::Bool(!p1.eq(&p2)))
    }),
    f("point_distance", &[POINT, POINT], FLOAT8, |_, a| float(Point::read(&a[0])?.dt(&Point::read(&a[1])?))),
    f("<->", &[POINT, POINT], FLOAT8, |_, a| float(Point::read(&a[0])?.dt(&Point::read(&a[1])?))),
    f("slope", &[POINT, POINT], FLOAT8, |_, a| float(Point::read(&a[0])?.sl(&Point::read(&a[1])?))),
    f("point_add", &[POINT, POINT], POINT, |_, a| Ok(Point::read(&a[0])?.add(&Point::read(&a[1])?)?.value())),
    f("+", &[POINT, POINT], POINT, |_, a| Ok(Point::read(&a[0])?.add(&Point::read(&a[1])?)?.value())),
    f("point_sub", &[POINT, POINT], POINT, |_, a| Ok(Point::read(&a[0])?.sub(&Point::read(&a[1])?)?.value())),
    f("-", &[POINT, POINT], POINT, |_, a| Ok(Point::read(&a[0])?.sub(&Point::read(&a[1])?)?.value())),
    f("point_mul", &[POINT, POINT], POINT, |_, a| Ok(Point::read(&a[0])?.mul(&Point::read(&a[1])?)?.value())),
    f("*", &[POINT, POINT], POINT, |_, a| Ok(Point::read(&a[0])?.mul(&Point::read(&a[1])?)?.value())),
    f("point_div", &[POINT, POINT], POINT, |_, a| Ok(Point::read(&a[0])?.div(&Point::read(&a[1])?)?.value())),
    f("/", &[POINT, POINT], POINT, |_, a| Ok(Point::read(&a[0])?.div(&Point::read(&a[1])?)?.value())),
    f("point", &[FLOAT8, FLOAT8], POINT, |_, a| {
        let (Value::Float8(x), Value::Float8(y)) = (&a[0], &a[1]) else {
            return Err(PgError::internal("point coordinates that are not float8"));
        };
        Ok(Point { x: *x, y: *y }.value())
    }),
    f("on_pb", &[POINT, BOX], BOOL, |_, a| {
        Ok(Value::Bool(box_contain_point(&GeoBox::read(&a[1])?, &Point::read(&a[0])?)))
    }),
    f("<@", &[POINT, BOX], BOOL, |_, a| {
        Ok(Value::Bool(box_contain_point(&GeoBox::read(&a[1])?, &Point::read(&a[0])?)))
    }),
    f("box_contain_pt", &[BOX, POINT], BOOL, |_, a| {
        Ok(Value::Bool(box_contain_point(&GeoBox::read(&a[0])?, &Point::read(&a[1])?)))
    }),
    f("@>", &[BOX, POINT], BOOL, |_, a| {
        Ok(Value::Bool(box_contain_point(&GeoBox::read(&a[0])?, &Point::read(&a[1])?)))
    }),
    f("on_ppath", &[POINT, PATH], BOOL, |_, a| boolean(on_ppath(&Point::read(&a[0])?, &Path::read(&a[1])?))),
    f("<@", &[POINT, PATH], BOOL, |_, a| boolean(on_ppath(&Point::read(&a[0])?, &Path::read(&a[1])?))),
    f("path_contain_pt", &[PATH, POINT], BOOL, |_, a| boolean(on_ppath(&Point::read(&a[1])?, &Path::read(&a[0])?))),
    f("@>", &[PATH, POINT], BOOL, |_, a| boolean(on_ppath(&Point::read(&a[1])?, &Path::read(&a[0])?))),
    f("lseg_center", &[LSEG], POINT, |_, a| Ok(Lseg::read(&a[0])?.center()?.value())),
    f("point", &[LSEG], POINT, |_, a| point_of_lseg(a)),
    f("@@", &[LSEG], POINT, |_, a| Ok(Lseg::read(&a[0])?.center()?.value())),
    f("lseg", &[POINT, POINT], LSEG, |_, a| Ok(Lseg { p: [Point::read(&a[0])?, Point::read(&a[1])?] }.value())),
    f("lseg_length", &[LSEG], FLOAT8, |_, a| float(Lseg::read(&a[0])?.length())),
    f("length", &[LSEG], FLOAT8, |_, a| float(Lseg::read(&a[0])?.length())),
    f("@-@", &[LSEG], FLOAT8, |_, a| float(Lseg::read(&a[0])?.length())),
    f("lseg_intersect", &[LSEG, LSEG], BOOL, |_, a| {
        Ok(Value::Bool(lseg_interpt_lseg(&Lseg::read(&a[0])?, &Lseg::read(&a[1])?)?.is_some()))
    }),
    f("?#", &[LSEG, LSEG], BOOL, |_, a| {
        Ok(Value::Bool(lseg_interpt_lseg(&Lseg::read(&a[0])?, &Lseg::read(&a[1])?)?.is_some()))
    }),
    f("lseg_parallel", &[LSEG, LSEG], BOOL, |_, a| {
        Ok(Value::Bool(fp_eq(Lseg::read(&a[0])?.sl()?, Lseg::read(&a[1])?.sl()?)))
    }),
    f("isparallel", &[LSEG, LSEG], BOOL, |_, a| {
        Ok(Value::Bool(fp_eq(Lseg::read(&a[0])?.sl()?, Lseg::read(&a[1])?.sl()?)))
    }),
    f("?||", &[LSEG, LSEG], BOOL, |_, a| Ok(Value::Bool(fp_eq(Lseg::read(&a[0])?.sl()?, Lseg::read(&a[1])?.sl()?)))),
    f("lseg_perp", &[LSEG, LSEG], BOOL, |_, a| {
        Ok(Value::Bool(fp_eq(Lseg::read(&a[0])?.sl()?, Lseg::read(&a[1])?.invsl()?)))
    }),
    f("isperp", &[LSEG, LSEG], BOOL, |_, a| {
        Ok(Value::Bool(fp_eq(Lseg::read(&a[0])?.sl()?, Lseg::read(&a[1])?.invsl()?)))
    }),
    f("?-|", &[LSEG, LSEG], BOOL, |_, a| Ok(Value::Bool(fp_eq(Lseg::read(&a[0])?.sl()?, Lseg::read(&a[1])?.invsl()?)))),
    f("lseg_vertical", &[LSEG], BOOL, |_, a| {
        let l = Lseg::read(&a[0])?;
        Ok(Value::Bool(fp_eq(l.p[0].x, l.p[1].x)))
    }),
    f("isvertical", &[LSEG], BOOL, |_, a| {
        let l = Lseg::read(&a[0])?;
        Ok(Value::Bool(fp_eq(l.p[0].x, l.p[1].x)))
    }),
    f("?|", &[LSEG], BOOL, |_, a| {
        let l = Lseg::read(&a[0])?;
        Ok(Value::Bool(fp_eq(l.p[0].x, l.p[1].x)))
    }),
    f("lseg_horizontal", &[LSEG], BOOL, |_, a| {
        let l = Lseg::read(&a[0])?;
        Ok(Value::Bool(fp_eq(l.p[0].y, l.p[1].y)))
    }),
    f("ishorizontal", &[LSEG], BOOL, |_, a| {
        let l = Lseg::read(&a[0])?;
        Ok(Value::Bool(fp_eq(l.p[0].y, l.p[1].y)))
    }),
    f("?-", &[LSEG], BOOL, |_, a| {
        let l = Lseg::read(&a[0])?;
        Ok(Value::Bool(fp_eq(l.p[0].y, l.p[1].y)))
    }),
    f("lseg_eq", &[LSEG, LSEG], BOOL, |_, a| {
        let (l1, l2) = (Lseg::read(&a[0])?, Lseg::read(&a[1])?);
        Ok(Value::Bool(l1.p[0].eq(&l2.p[0]) && l1.p[1].eq(&l2.p[1])))
    }),
    f("=", &[LSEG, LSEG], BOOL, |_, a| {
        let (l1, l2) = (Lseg::read(&a[0])?, Lseg::read(&a[1])?);
        Ok(Value::Bool(l1.p[0].eq(&l2.p[0]) && l1.p[1].eq(&l2.p[1])))
    }),
    f("lseg_ne", &[LSEG, LSEG], BOOL, |_, a| {
        let (l1, l2) = (Lseg::read(&a[0])?, Lseg::read(&a[1])?);
        Ok(Value::Bool(!l1.p[0].eq(&l2.p[0]) || !l1.p[1].eq(&l2.p[1])))
    }),
    f("<>", &[LSEG, LSEG], BOOL, |_, a| {
        let (l1, l2) = (Lseg::read(&a[0])?, Lseg::read(&a[1])?);
        Ok(Value::Bool(!l1.p[0].eq(&l2.p[0]) || !l1.p[1].eq(&l2.p[1])))
    }),
    f("lseg_lt", &[LSEG, LSEG], BOOL, |_, a| {
        Ok(Value::Bool(fp_lt(Lseg::read(&a[0])?.length()?, Lseg::read(&a[1])?.length()?)))
    }),
    f("<", &[LSEG, LSEG], BOOL, |_, a| {
        Ok(Value::Bool(fp_lt(Lseg::read(&a[0])?.length()?, Lseg::read(&a[1])?.length()?)))
    }),
    f("lseg_le", &[LSEG, LSEG], BOOL, |_, a| {
        Ok(Value::Bool(fp_le(Lseg::read(&a[0])?.length()?, Lseg::read(&a[1])?.length()?)))
    }),
    f("<=", &[LSEG, LSEG], BOOL, |_, a| {
        Ok(Value::Bool(fp_le(Lseg::read(&a[0])?.length()?, Lseg::read(&a[1])?.length()?)))
    }),
    f("lseg_gt", &[LSEG, LSEG], BOOL, |_, a| {
        Ok(Value::Bool(fp_gt(Lseg::read(&a[0])?.length()?, Lseg::read(&a[1])?.length()?)))
    }),
    f(">", &[LSEG, LSEG], BOOL, |_, a| {
        Ok(Value::Bool(fp_gt(Lseg::read(&a[0])?.length()?, Lseg::read(&a[1])?.length()?)))
    }),
    f("lseg_ge", &[LSEG, LSEG], BOOL, |_, a| {
        Ok(Value::Bool(fp_ge(Lseg::read(&a[0])?.length()?, Lseg::read(&a[1])?.length()?)))
    }),
    f(">=", &[LSEG, LSEG], BOOL, |_, a| {
        Ok(Value::Bool(fp_ge(Lseg::read(&a[0])?.length()?, Lseg::read(&a[1])?.length()?)))
    }),
    f("lseg_distance", &[LSEG, LSEG], FLOAT8, |_, a| {
        float(lseg_closept_lseg(&Lseg::read(&a[0])?, &Lseg::read(&a[1])?).map(|r| r.1))
    }),
    f("<->", &[LSEG, LSEG], FLOAT8, |_, a| {
        float(lseg_closept_lseg(&Lseg::read(&a[0])?, &Lseg::read(&a[1])?).map(|r| r.1))
    }),
    f("lseg_interpt", &[LSEG, LSEG], POINT, |_, a| {
        Ok(lseg_interpt_lseg(&Lseg::read(&a[0])?, &Lseg::read(&a[1])?)?.map_or(Value::Null, Point::value))
    }),
    f("#", &[LSEG, LSEG], POINT, |_, a| {
        Ok(lseg_interpt_lseg(&Lseg::read(&a[0])?, &Lseg::read(&a[1])?)?.map_or(Value::Null, Point::value))
    }),
    f("dist_pl", &[POINT, LINE], FLOAT8, |_, a| {
        float(line_closept_point(&Line::read(&a[1])?, &Point::read(&a[0])?).map(|r| r.1))
    }),
    f("<->", &[POINT, LINE], FLOAT8, |_, a| {
        float(line_closept_point(&Line::read(&a[1])?, &Point::read(&a[0])?).map(|r| r.1))
    }),
    f("dist_lp", &[LINE, POINT], FLOAT8, |_, a| {
        float(line_closept_point(&Line::read(&a[0])?, &Point::read(&a[1])?).map(|r| r.1))
    }),
    f("<->", &[LINE, POINT], FLOAT8, |_, a| {
        float(line_closept_point(&Line::read(&a[0])?, &Point::read(&a[1])?).map(|r| r.1))
    }),
    f("dist_ps", &[POINT, LSEG], FLOAT8, |_, a| {
        float(lseg_closept_point(&Lseg::read(&a[1])?, &Point::read(&a[0])?).map(|r| r.1))
    }),
    f("<->", &[POINT, LSEG], FLOAT8, |_, a| {
        float(lseg_closept_point(&Lseg::read(&a[1])?, &Point::read(&a[0])?).map(|r| r.1))
    }),
    f("dist_sp", &[LSEG, POINT], FLOAT8, |_, a| {
        float(lseg_closept_point(&Lseg::read(&a[0])?, &Point::read(&a[1])?).map(|r| r.1))
    }),
    f("<->", &[LSEG, POINT], FLOAT8, |_, a| {
        float(lseg_closept_point(&Lseg::read(&a[0])?, &Point::read(&a[1])?).map(|r| r.1))
    }),
    f("dist_pb", &[POINT, BOX], FLOAT8, |_, a| {
        float(box_closept_point(&GeoBox::read(&a[1])?, &Point::read(&a[0])?).map(|r| r.1))
    }),
    f("<->", &[POINT, BOX], FLOAT8, |_, a| {
        float(box_closept_point(&GeoBox::read(&a[1])?, &Point::read(&a[0])?).map(|r| r.1))
    }),
    f("dist_bp", &[BOX, POINT], FLOAT8, |_, a| {
        float(box_closept_point(&GeoBox::read(&a[0])?, &Point::read(&a[1])?).map(|r| r.1))
    }),
    f("<->", &[BOX, POINT], FLOAT8, |_, a| {
        float(box_closept_point(&GeoBox::read(&a[0])?, &Point::read(&a[1])?).map(|r| r.1))
    }),
    f("dist_sl", &[LSEG, LINE], FLOAT8, |_, a| {
        float(lseg_closept_line(&Lseg::read(&a[0])?, &Line::read(&a[1])?).map(|r| r.1))
    }),
    f("<->", &[LSEG, LINE], FLOAT8, |_, a| {
        float(lseg_closept_line(&Lseg::read(&a[0])?, &Line::read(&a[1])?).map(|r| r.1))
    }),
    f("dist_ls", &[LINE, LSEG], FLOAT8, |_, a| {
        float(lseg_closept_line(&Lseg::read(&a[1])?, &Line::read(&a[0])?).map(|r| r.1))
    }),
    f("<->", &[LINE, LSEG], FLOAT8, |_, a| {
        float(lseg_closept_line(&Lseg::read(&a[1])?, &Line::read(&a[0])?).map(|r| r.1))
    }),
    f("dist_sb", &[LSEG, BOX], FLOAT8, |_, a| {
        float(box_closept_lseg(&GeoBox::read(&a[1])?, &Lseg::read(&a[0])?).map(|r| r.1))
    }),
    f("<->", &[LSEG, BOX], FLOAT8, |_, a| {
        float(box_closept_lseg(&GeoBox::read(&a[1])?, &Lseg::read(&a[0])?).map(|r| r.1))
    }),
    f("dist_bs", &[BOX, LSEG], FLOAT8, |_, a| {
        float(box_closept_lseg(&GeoBox::read(&a[0])?, &Lseg::read(&a[1])?).map(|r| r.1))
    }),
    f("<->", &[BOX, LSEG], FLOAT8, |_, a| {
        float(box_closept_lseg(&GeoBox::read(&a[0])?, &Lseg::read(&a[1])?).map(|r| r.1))
    }),
    f("dist_ppath", &[POINT, PATH], FLOAT8, |_, a| float(dist_ppath(&Point::read(&a[0])?, &Path::read(&a[1])?))),
    f("<->", &[POINT, PATH], FLOAT8, |_, a| float(dist_ppath(&Point::read(&a[0])?, &Path::read(&a[1])?))),
    f("dist_pathp", &[PATH, POINT], FLOAT8, |_, a| float(dist_ppath(&Point::read(&a[1])?, &Path::read(&a[0])?))),
    f("<->", &[PATH, POINT], FLOAT8, |_, a| float(dist_ppath(&Point::read(&a[1])?, &Path::read(&a[0])?))),
    f("dist_ppoly", &[POINT, POLYGON], FLOAT8, |_, a| float(dist_ppoly(&Point::read(&a[0])?, &Polygon::read(&a[1])?))),
    f("<->", &[POINT, POLYGON], FLOAT8, |_, a| float(dist_ppoly(&Point::read(&a[0])?, &Polygon::read(&a[1])?))),
    f("dist_polyp", &[POLYGON, POINT], FLOAT8, |_, a| float(dist_ppoly(&Point::read(&a[1])?, &Polygon::read(&a[0])?))),
    f("<->", &[POLYGON, POINT], FLOAT8, |_, a| float(dist_ppoly(&Point::read(&a[1])?, &Polygon::read(&a[0])?))),
    f("dist_cpoly", &[CIRCLE, POLYGON], FLOAT8, |_, a| {
        let c = Circle::read(&a[0])?;
        float(dist_ppoly(&c.center, &Polygon::read(&a[1])?).and_then(|d| mi(d, c.radius)).map(nonnegative))
    }),
    f("<->", &[CIRCLE, POLYGON], FLOAT8, |_, a| {
        let c = Circle::read(&a[0])?;
        float(dist_ppoly(&c.center, &Polygon::read(&a[1])?).and_then(|d| mi(d, c.radius)).map(nonnegative))
    }),
    f("dist_polyc", &[POLYGON, CIRCLE], FLOAT8, |_, a| {
        let c = Circle::read(&a[1])?;
        float(dist_ppoly(&c.center, &Polygon::read(&a[0])?).and_then(|d| mi(d, c.radius)).map(nonnegative))
    }),
    f("<->", &[POLYGON, CIRCLE], FLOAT8, |_, a| {
        let c = Circle::read(&a[1])?;
        float(dist_ppoly(&c.center, &Polygon::read(&a[0])?).and_then(|d| mi(d, c.radius)).map(nonnegative))
    }),
    f("dist_pc", &[POINT, CIRCLE], FLOAT8, |_, a| {
        let c = Circle::read(&a[1])?;
        float(Point::read(&a[0])?.dt(&c.center).and_then(|d| mi(d, c.radius)).map(nonnegative))
    }),
    f("<->", &[POINT, CIRCLE], FLOAT8, |_, a| {
        let c = Circle::read(&a[1])?;
        float(Point::read(&a[0])?.dt(&c.center).and_then(|d| mi(d, c.radius)).map(nonnegative))
    }),
    f("dist_cpoint", &[CIRCLE, POINT], FLOAT8, |_, a| {
        let c = Circle::read(&a[0])?;
        float(Point::read(&a[1])?.dt(&c.center).and_then(|d| mi(d, c.radius)).map(nonnegative))
    }),
    f("<->", &[CIRCLE, POINT], FLOAT8, |_, a| {
        let c = Circle::read(&a[0])?;
        float(Point::read(&a[1])?.dt(&c.center).and_then(|d| mi(d, c.radius)).map(nonnegative))
    }),
    f("close_pl", &[POINT, LINE], POINT, |_, a| {
        Ok(optional(line_closept_point(&Line::read(&a[1])?, &Point::read(&a[0])?)?))
    }),
    f("##", &[POINT, LINE], POINT, |_, a| Ok(optional(line_closept_point(&Line::read(&a[1])?, &Point::read(&a[0])?)?))),
    f("close_ps", &[POINT, LSEG], POINT, |_, a| {
        Ok(optional(lseg_closept_point(&Lseg::read(&a[1])?, &Point::read(&a[0])?)?))
    }),
    f("##", &[POINT, LSEG], POINT, |_, a| Ok(optional(lseg_closept_point(&Lseg::read(&a[1])?, &Point::read(&a[0])?)?))),
    f("close_pb", &[POINT, BOX], POINT, |_, a| {
        Ok(optional(box_closept_point(&GeoBox::read(&a[1])?, &Point::read(&a[0])?)?))
    }),
    f("##", &[POINT, BOX], POINT, |_, a| Ok(optional(box_closept_point(&GeoBox::read(&a[1])?, &Point::read(&a[0])?)?))),
    f("close_sb", &[LSEG, BOX], POINT, |_, a| {
        Ok(optional(box_closept_lseg(&GeoBox::read(&a[1])?, &Lseg::read(&a[0])?)?))
    }),
    f("##", &[LSEG, BOX], POINT, |_, a| Ok(optional(box_closept_lseg(&GeoBox::read(&a[1])?, &Lseg::read(&a[0])?)?))),
    f("close_ls", &[LINE, LSEG], POINT, |_, a| {
        let (line, lseg) = (Line::read(&a[0])?, Lseg::read(&a[1])?);
        if lseg.sl()? == line.sl()? {
            return Ok(Value::Null);
        }
        Ok(optional(lseg_closept_line(&lseg, &line)?))
    }),
    f("##", &[LINE, LSEG], POINT, |_, a| {
        let (line, lseg) = (Line::read(&a[0])?, Lseg::read(&a[1])?);
        if lseg.sl()? == line.sl()? {
            return Ok(Value::Null);
        }
        Ok(optional(lseg_closept_line(&lseg, &line)?))
    }),
    f("close_lseg", &[LSEG, LSEG], POINT, |_, a| {
        let (l1, l2) = (Lseg::read(&a[0])?, Lseg::read(&a[1])?);
        if l1.sl()? == l2.sl()? {
            return Ok(Value::Null);
        }
        Ok(optional(lseg_closept_lseg(&l2, &l1)?))
    }),
    f("##", &[LSEG, LSEG], POINT, |_, a| {
        let (l1, l2) = (Lseg::read(&a[0])?, Lseg::read(&a[1])?);
        if l1.sl()? == l2.sl()? {
            return Ok(Value::Null);
        }
        Ok(optional(lseg_closept_lseg(&l2, &l1)?))
    }),
    f("on_pl", &[POINT, LINE], BOOL, |_, a| boolean(line_contain_point(&Line::read(&a[1])?, &Point::read(&a[0])?))),
    f("<@", &[POINT, LINE], BOOL, |_, a| boolean(line_contain_point(&Line::read(&a[1])?, &Point::read(&a[0])?))),
    f("on_ps", &[POINT, LSEG], BOOL, |_, a| boolean(lseg_contain_point(&Lseg::read(&a[1])?, &Point::read(&a[0])?))),
    f("<@", &[POINT, LSEG], BOOL, |_, a| boolean(lseg_contain_point(&Lseg::read(&a[1])?, &Point::read(&a[0])?))),
    f("on_sl", &[LSEG, LINE], BOOL, |_, a| {
        let (l, line) = (Lseg::read(&a[0])?, Line::read(&a[1])?);
        Ok(Value::Bool(line_contain_point(&line, &l.p[0])? && line_contain_point(&line, &l.p[1])?))
    }),
    f("<@", &[LSEG, LINE], BOOL, |_, a| {
        let (l, line) = (Lseg::read(&a[0])?, Line::read(&a[1])?);
        Ok(Value::Bool(line_contain_point(&line, &l.p[0])? && line_contain_point(&line, &l.p[1])?))
    }),
    f("on_sb", &[LSEG, BOX], BOOL, |_, a| {
        let (l, b) = (Lseg::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(box_contain_point(&b, &l.p[0]) && box_contain_point(&b, &l.p[1])))
    }),
    f("<@", &[LSEG, BOX], BOOL, |_, a| {
        let (l, b) = (Lseg::read(&a[0])?, GeoBox::read(&a[1])?);
        Ok(Value::Bool(box_contain_point(&b, &l.p[0]) && box_contain_point(&b, &l.p[1])))
    }),
    f("inter_sl", &[LSEG, LINE], BOOL, |_, a| {
        Ok(Value::Bool(lseg_interpt_line(&Lseg::read(&a[0])?, &Line::read(&a[1])?)?.is_some()))
    }),
    f("?#", &[LSEG, LINE], BOOL, |_, a| {
        Ok(Value::Bool(lseg_interpt_line(&Lseg::read(&a[0])?, &Line::read(&a[1])?)?.is_some()))
    }),
    f("inter_sb", &[LSEG, BOX], BOOL, |_, a| {
        Ok(Value::Bool(box_interpt_lseg(&GeoBox::read(&a[1])?, &Lseg::read(&a[0])?)?.is_some()))
    }),
    f("?#", &[LSEG, BOX], BOOL, |_, a| {
        Ok(Value::Bool(box_interpt_lseg(&GeoBox::read(&a[1])?, &Lseg::read(&a[0])?)?.is_some()))
    }),
    f("inter_lb", &[LINE, BOX], BOOL, |_, a| {
        let (line, b) = (Line::read(&a[0])?, GeoBox::read(&a[1])?);
        let corners = [b.low, Point { x: b.low.x, y: b.high.y }, b.high, Point { x: b.high.x, y: b.low.y }];
        for i in 0..4 {
            if lseg_interpt_line(&Lseg { p: [corners[(i + 3) % 4], corners[i]] }, &line)?.is_some() {
                return Ok(Value::Bool(true));
            }
        }
        Ok(Value::Bool(false))
    }),
    f("?#", &[LINE, BOX], BOOL, |_, a| {
        let (line, b) = (Line::read(&a[0])?, GeoBox::read(&a[1])?);
        let corners = [b.low, Point { x: b.low.x, y: b.high.y }, b.high, Point { x: b.high.x, y: b.low.y }];
        for i in 0..4 {
            if lseg_interpt_line(&Lseg { p: [corners[(i + 3) % 4], corners[i]] }, &line)?.is_some() {
                return Ok(Value::Bool(true));
            }
        }
        Ok(Value::Bool(false))
    }),
    f("line", &[POINT, POINT], LINE, |_, a| {
        let (p1, p2) = (Point::read(&a[0])?, Point::read(&a[1])?);
        if p1.eq(&p2) {
            return Err(PgError::new(
                code::INVALID_PARAMETER_VALUE,
                "invalid line specification: must be two distinct points",
            ));
        }
        Ok(Line::construct(&p1, p1.sl(&p2)?)?.value())
    }),
    f("line_intersect", &[LINE, LINE], BOOL, |_, a| {
        Ok(Value::Bool(Line::read(&a[0])?.interpt(&Line::read(&a[1])?)?.is_some()))
    }),
    f("?#", &[LINE, LINE], BOOL, |_, a| Ok(Value::Bool(Line::read(&a[0])?.interpt(&Line::read(&a[1])?)?.is_some()))),
    f("line_parallel", &[LINE, LINE], BOOL, |_, a| {
        Ok(Value::Bool(Line::read(&a[0])?.interpt(&Line::read(&a[1])?)?.is_none()))
    }),
    f("isparallel", &[LINE, LINE], BOOL, |_, a| {
        Ok(Value::Bool(Line::read(&a[0])?.interpt(&Line::read(&a[1])?)?.is_none()))
    }),
    f("?||", &[LINE, LINE], BOOL, |_, a| Ok(Value::Bool(Line::read(&a[0])?.interpt(&Line::read(&a[1])?)?.is_none()))),
    f("line_perp", &[LINE, LINE], BOOL, |_, a| boolean(line_perp(&Line::read(&a[0])?, &Line::read(&a[1])?))),
    f("isperp", &[LINE, LINE], BOOL, |_, a| boolean(line_perp(&Line::read(&a[0])?, &Line::read(&a[1])?))),
    f("?-|", &[LINE, LINE], BOOL, |_, a| boolean(line_perp(&Line::read(&a[0])?, &Line::read(&a[1])?))),
    f("line_vertical", &[LINE], BOOL, |_, a| Ok(Value::Bool(fp_zero(Line::read(&a[0])?.b)))),
    f("isvertical", &[LINE], BOOL, |_, a| Ok(Value::Bool(fp_zero(Line::read(&a[0])?.b)))),
    f("?|", &[LINE], BOOL, |_, a| Ok(Value::Bool(fp_zero(Line::read(&a[0])?.b)))),
    f("line_horizontal", &[LINE], BOOL, |_, a| Ok(Value::Bool(fp_zero(Line::read(&a[0])?.a)))),
    f("ishorizontal", &[LINE], BOOL, |_, a| Ok(Value::Bool(fp_zero(Line::read(&a[0])?.a)))),
    f("?-", &[LINE], BOOL, |_, a| Ok(Value::Bool(fp_zero(Line::read(&a[0])?.a)))),
    f("line_eq", &[LINE, LINE], BOOL, |_, a| boolean(line_eq(&Line::read(&a[0])?, &Line::read(&a[1])?))),
    f("=", &[LINE, LINE], BOOL, |_, a| boolean(line_eq(&Line::read(&a[0])?, &Line::read(&a[1])?))),
    f("line_distance", &[LINE, LINE], FLOAT8, |_, a| float(line_distance(&Line::read(&a[0])?, &Line::read(&a[1])?))),
    f("<->", &[LINE, LINE], FLOAT8, |_, a| float(line_distance(&Line::read(&a[0])?, &Line::read(&a[1])?))),
    f("line_interpt", &[LINE, LINE], POINT, |_, a| {
        Ok(Line::read(&a[0])?.interpt(&Line::read(&a[1])?)?.map_or(Value::Null, Point::value))
    }),
    f("#", &[LINE, LINE], POINT, |_, a| {
        Ok(Line::read(&a[0])?.interpt(&Line::read(&a[1])?)?.map_or(Value::Null, Point::value))
    }),
    f("path_area", &[PATH], FLOAT8, |_, a| {
        let p = Path::read(&a[0])?;
        if !p.closed {
            return Ok(Value::Null);
        }
        let n = p.p.len();
        let mut area = 0.0;
        for i in 0..n {
            let j = (i + 1) % n;
            area = pl(area, mul(p.p[i].x, p.p[j].y)?)?;
            area = mi(area, mul(p.p[i].y, p.p[j].x)?)?;
        }
        float(div(area.abs(), 2.0))
    }),
    f("area", &[PATH], FLOAT8, |_, a| {
        let p = Path::read(&a[0])?;
        if !p.closed {
            return Ok(Value::Null);
        }
        let n = p.p.len();
        let mut area = 0.0;
        for i in 0..n {
            let j = (i + 1) % n;
            area = pl(area, mul(p.p[i].x, p.p[j].y)?)?;
            area = mi(area, mul(p.p[i].y, p.p[j].x)?)?;
        }
        float(div(area.abs(), 2.0))
    }),
    f("path_n_lt", &[PATH, PATH], BOOL, |_, a| {
        Ok(Value::Bool(Path::read(&a[0])?.p.len() < Path::read(&a[1])?.p.len()))
    }),
    f("<", &[PATH, PATH], BOOL, |_, a| Ok(Value::Bool(Path::read(&a[0])?.p.len() < Path::read(&a[1])?.p.len()))),
    f("path_n_gt", &[PATH, PATH], BOOL, |_, a| {
        Ok(Value::Bool(Path::read(&a[0])?.p.len() > Path::read(&a[1])?.p.len()))
    }),
    f(">", &[PATH, PATH], BOOL, |_, a| Ok(Value::Bool(Path::read(&a[0])?.p.len() > Path::read(&a[1])?.p.len()))),
    f("path_n_eq", &[PATH, PATH], BOOL, |_, a| {
        Ok(Value::Bool(Path::read(&a[0])?.p.len() == Path::read(&a[1])?.p.len()))
    }),
    f("=", &[PATH, PATH], BOOL, |_, a| Ok(Value::Bool(Path::read(&a[0])?.p.len() == Path::read(&a[1])?.p.len()))),
    f("path_n_le", &[PATH, PATH], BOOL, |_, a| {
        Ok(Value::Bool(Path::read(&a[0])?.p.len() <= Path::read(&a[1])?.p.len()))
    }),
    f("<=", &[PATH, PATH], BOOL, |_, a| Ok(Value::Bool(Path::read(&a[0])?.p.len() <= Path::read(&a[1])?.p.len()))),
    f("path_n_ge", &[PATH, PATH], BOOL, |_, a| {
        Ok(Value::Bool(Path::read(&a[0])?.p.len() >= Path::read(&a[1])?.p.len()))
    }),
    f(">=", &[PATH, PATH], BOOL, |_, a| Ok(Value::Bool(Path::read(&a[0])?.p.len() >= Path::read(&a[1])?.p.len()))),
    f("isclosed", &[PATH], BOOL, |_, a| Ok(Value::Bool(Path::read(&a[0])?.closed))),
    f("isopen", &[PATH], BOOL, |_, a| Ok(Value::Bool(!Path::read(&a[0])?.closed))),
    f("path_npoints", &[PATH], INT4, |_, a| Ok(Value::Int4(Path::read(&a[0])?.p.len() as i32))),
    f("npoints", &[PATH], INT4, |_, a| Ok(Value::Int4(Path::read(&a[0])?.p.len() as i32))),
    f("#", &[PATH], INT4, |_, a| Ok(Value::Int4(Path::read(&a[0])?.p.len() as i32))),
    f("pclose", &[PATH], PATH, |_, a| Ok(Path { closed: true, ..Path::read(&a[0])? }.value())),
    f("popen", &[PATH], PATH, |_, a| Ok(Path { closed: false, ..Path::read(&a[0])? }.value())),
    f("path_inter", &[PATH, PATH], BOOL, |_, a| boolean(path_inter(&Path::read(&a[0])?, &Path::read(&a[1])?))),
    f("?#", &[PATH, PATH], BOOL, |_, a| boolean(path_inter(&Path::read(&a[0])?, &Path::read(&a[1])?))),
    f("path_distance", &[PATH, PATH], FLOAT8, |_, a| {
        let (p1, p2) = (Path::read(&a[0])?, Path::read(&a[1])?);
        let pairs = p1
            .segments()
            .into_iter()
            .flat_map(|s1| p2.segments().into_iter().map(move |s2| lseg_closept_lseg(&s1, &s2).map(|r| r.1)));
        Ok(minimum(pairs)?.map_or(Value::Null, Value::Float8))
    }),
    f("<->", &[PATH, PATH], FLOAT8, |_, a| {
        let (p1, p2) = (Path::read(&a[0])?, Path::read(&a[1])?);
        let pairs = p1
            .segments()
            .into_iter()
            .flat_map(|s1| p2.segments().into_iter().map(move |s2| lseg_closept_lseg(&s1, &s2).map(|r| r.1)));
        Ok(minimum(pairs)?.map_or(Value::Null, Value::Float8))
    }),
    f("path_length", &[PATH], FLOAT8, |_, a| {
        let mut total = 0.0;
        for s in Path::read(&a[0])?.segments() {
            total = pl(total, s.length()?)?;
        }
        Ok(Value::Float8(total))
    }),
    f("length", &[PATH], FLOAT8, |_, a| {
        let mut total = 0.0;
        for s in Path::read(&a[0])?.segments() {
            total = pl(total, s.length()?)?;
        }
        Ok(Value::Float8(total))
    }),
    f("@-@", &[PATH], FLOAT8, |_, a| {
        let mut total = 0.0;
        for s in Path::read(&a[0])?.segments() {
            total = pl(total, s.length()?)?;
        }
        Ok(Value::Float8(total))
    }),
    f("path_add", &[PATH, PATH], PATH, |_, a| {
        let (p1, p2) = (Path::read(&a[0])?, Path::read(&a[1])?);
        if p1.closed || p2.closed {
            return Ok(Value::Null);
        }
        Ok(Path { closed: false, p: [p1.p, p2.p].concat() }.value())
    }),
    f("+", &[PATH, PATH], PATH, |_, a| {
        let (p1, p2) = (Path::read(&a[0])?, Path::read(&a[1])?);
        if p1.closed || p2.closed {
            return Ok(Value::Null);
        }
        Ok(Path { closed: false, p: [p1.p, p2.p].concat() }.value())
    }),
    f("path_add_pt", &[PATH, POINT], PATH, |_, a| map_path(&a[0], &a[1], Point::add)),
    f("+", &[PATH, POINT], PATH, |_, a| map_path(&a[0], &a[1], Point::add)),
    f("path_sub_pt", &[PATH, POINT], PATH, |_, a| map_path(&a[0], &a[1], Point::sub)),
    f("-", &[PATH, POINT], PATH, |_, a| map_path(&a[0], &a[1], Point::sub)),
    f("path_mul_pt", &[PATH, POINT], PATH, |_, a| map_path(&a[0], &a[1], Point::mul)),
    f("*", &[PATH, POINT], PATH, |_, a| map_path(&a[0], &a[1], Point::mul)),
    f("path_div_pt", &[PATH, POINT], PATH, |_, a| map_path(&a[0], &a[1], Point::div)),
    f("/", &[PATH, POINT], PATH, |_, a| map_path(&a[0], &a[1], Point::div)),
    f("path", &[POLYGON], PATH, |_, a| path_of_polygon(a)),
    f("poly_left", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.bound.high.x < p2.bound.low.x))
    }),
    f("<<", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.bound.high.x < p2.bound.low.x))
    }),
    f("poly_overleft", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.bound.high.x <= p2.bound.high.x))
    }),
    f("&<", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.bound.high.x <= p2.bound.high.x))
    }),
    f("poly_right", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.bound.low.x > p2.bound.high.x))
    }),
    f(">>", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.bound.low.x > p2.bound.high.x))
    }),
    f("poly_overright", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.bound.low.x >= p2.bound.low.x))
    }),
    f("&>", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.bound.low.x >= p2.bound.low.x))
    }),
    f("poly_below", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.bound.high.y < p2.bound.low.y))
    }),
    f("<<|", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.bound.high.y < p2.bound.low.y))
    }),
    f("poly_overbelow", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.bound.high.y <= p2.bound.high.y))
    }),
    f("&<|", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.bound.high.y <= p2.bound.high.y))
    }),
    f("poly_above", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.bound.low.y > p2.bound.high.y))
    }),
    f("|>>", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.bound.low.y > p2.bound.high.y))
    }),
    f("poly_overabove", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.bound.low.y >= p2.bound.low.y))
    }),
    f("|&>", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.bound.low.y >= p2.bound.low.y))
    }),
    f("poly_same", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.p.len() == p2.p.len() && plist_same(&p1.p, &p2.p)))
    }),
    f("~=", &[POLYGON, POLYGON], BOOL, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        Ok(Value::Bool(p1.p.len() == p2.p.len() && plist_same(&p1.p, &p2.p)))
    }),
    f("poly_overlap", &[POLYGON, POLYGON], BOOL, |_, a| {
        boolean(poly_overlap(&Polygon::read(&a[0])?, &Polygon::read(&a[1])?))
    }),
    f("&&", &[POLYGON, POLYGON], BOOL, |_, a| boolean(poly_overlap(&Polygon::read(&a[0])?, &Polygon::read(&a[1])?))),
    f("poly_contain", &[POLYGON, POLYGON], BOOL, |_, a| {
        boolean(poly_contain_poly(&Polygon::read(&a[0])?, &Polygon::read(&a[1])?))
    }),
    f("@>", &[POLYGON, POLYGON], BOOL, |_, a| {
        boolean(poly_contain_poly(&Polygon::read(&a[0])?, &Polygon::read(&a[1])?))
    }),
    f("poly_contained", &[POLYGON, POLYGON], BOOL, |_, a| {
        boolean(poly_contain_poly(&Polygon::read(&a[1])?, &Polygon::read(&a[0])?))
    }),
    f("<@", &[POLYGON, POLYGON], BOOL, |_, a| {
        boolean(poly_contain_poly(&Polygon::read(&a[1])?, &Polygon::read(&a[0])?))
    }),
    f("poly_contain_pt", &[POLYGON, POINT], BOOL, |_, a| {
        boolean(point_inside(&Point::read(&a[1])?, &Polygon::read(&a[0])?.p))
    }),
    f("@>", &[POLYGON, POINT], BOOL, |_, a| boolean(point_inside(&Point::read(&a[1])?, &Polygon::read(&a[0])?.p))),
    f("pt_contained_poly", &[POINT, POLYGON], BOOL, |_, a| {
        boolean(point_inside(&Point::read(&a[0])?, &Polygon::read(&a[1])?.p))
    }),
    f("<@", &[POINT, POLYGON], BOOL, |_, a| boolean(point_inside(&Point::read(&a[0])?, &Polygon::read(&a[1])?.p))),
    f("poly_distance", &[POLYGON, POLYGON], FLOAT8, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        if poly_overlap(&p1, &p2)? {
            return Ok(Value::Float8(0.0));
        }
        let pairs = p1
            .edges()
            .into_iter()
            .flat_map(|s1| p2.edges().into_iter().map(move |s2| lseg_closept_lseg(&s1, &s2).map(|r| r.1)));
        Ok(minimum(pairs)?.map_or(Value::Null, Value::Float8))
    }),
    f("<->", &[POLYGON, POLYGON], FLOAT8, |_, a| {
        let (p1, p2) = (Polygon::read(&a[0])?, Polygon::read(&a[1])?);
        if poly_overlap(&p1, &p2)? {
            return Ok(Value::Float8(0.0));
        }
        let pairs = p1
            .edges()
            .into_iter()
            .flat_map(|s1| p2.edges().into_iter().map(move |s2| lseg_closept_lseg(&s1, &s2).map(|r| r.1)));
        Ok(minimum(pairs)?.map_or(Value::Null, Value::Float8))
    }),
    f("poly_npoints", &[POLYGON], INT4, |_, a| Ok(Value::Int4(Polygon::read(&a[0])?.p.len() as i32))),
    f("npoints", &[POLYGON], INT4, |_, a| Ok(Value::Int4(Polygon::read(&a[0])?.p.len() as i32))),
    f("#", &[POLYGON], INT4, |_, a| Ok(Value::Int4(Polygon::read(&a[0])?.p.len() as i32))),
    f("poly_center", &[POLYGON], POINT, |_, a| Ok(Polygon::read(&a[0])?.to_circle()?.center.value())),
    f("point", &[POLYGON], POINT, |_, a| point_of_polygon(a)),
    f("@@", &[POLYGON], POINT, |_, a| Ok(Polygon::read(&a[0])?.to_circle()?.center.value())),
    f("polygon", &[PATH], POLYGON, |_, a| polygon_of_path(a)),
    f("polygon", &[BOX], POLYGON, |_, a| polygon_of_box(a)),
    f("polygon", &[INT4, CIRCLE], POLYGON, |_, a| {
        let Value::Int4(n) = a[0] else { return Err(PgError::internal("a point count that is not int4")) };
        Ok(circle_poly(n, &Circle::read(&a[1])?)?.value())
    }),
    f("polygon", &[CIRCLE], POLYGON, |_, a| polygon_of_circle(a)),
    f("circle_same", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(
            ((c1.radius.is_nan() && c2.radius.is_nan()) || fp_eq(c1.radius, c2.radius)) && c1.center.eq(&c2.center),
        ))
    }),
    f("~=", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(
            ((c1.radius.is_nan() && c2.radius.is_nan()) || fp_eq(c1.radius, c2.radius)) && c1.center.eq(&c2.center),
        ))
    }),
    f("circle_overlap", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_le(c1.center.dt(&c2.center)?, pl(c1.radius, c2.radius)?)))
    }),
    f("&&", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_le(c1.center.dt(&c2.center)?, pl(c1.radius, c2.radius)?)))
    }),
    f("circle_overleft", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_le(pl(c1.center.x, c1.radius)?, pl(c2.center.x, c2.radius)?)))
    }),
    f("&<", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_le(pl(c1.center.x, c1.radius)?, pl(c2.center.x, c2.radius)?)))
    }),
    f("circle_left", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_lt(pl(c1.center.x, c1.radius)?, mi(c2.center.x, c2.radius)?)))
    }),
    f("<<", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_lt(pl(c1.center.x, c1.radius)?, mi(c2.center.x, c2.radius)?)))
    }),
    f("circle_right", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_gt(mi(c1.center.x, c1.radius)?, pl(c2.center.x, c2.radius)?)))
    }),
    f(">>", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_gt(mi(c1.center.x, c1.radius)?, pl(c2.center.x, c2.radius)?)))
    }),
    f("circle_overright", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_ge(mi(c1.center.x, c1.radius)?, mi(c2.center.x, c2.radius)?)))
    }),
    f("&>", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_ge(mi(c1.center.x, c1.radius)?, mi(c2.center.x, c2.radius)?)))
    }),
    f("circle_contained", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_le(c1.center.dt(&c2.center)?, mi(c2.radius, c1.radius)?)))
    }),
    f("<@", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_le(c1.center.dt(&c2.center)?, mi(c2.radius, c1.radius)?)))
    }),
    f("circle_contain", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_le(c1.center.dt(&c2.center)?, mi(c1.radius, c2.radius)?)))
    }),
    f("@>", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_le(c1.center.dt(&c2.center)?, mi(c1.radius, c2.radius)?)))
    }),
    f("circle_below", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_lt(pl(c1.center.y, c1.radius)?, mi(c2.center.y, c2.radius)?)))
    }),
    f("<<|", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_lt(pl(c1.center.y, c1.radius)?, mi(c2.center.y, c2.radius)?)))
    }),
    f("circle_above", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_gt(mi(c1.center.y, c1.radius)?, pl(c2.center.y, c2.radius)?)))
    }),
    f("|>>", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_gt(mi(c1.center.y, c1.radius)?, pl(c2.center.y, c2.radius)?)))
    }),
    f("circle_overbelow", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_le(pl(c1.center.y, c1.radius)?, pl(c2.center.y, c2.radius)?)))
    }),
    f("&<|", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_le(pl(c1.center.y, c1.radius)?, pl(c2.center.y, c2.radius)?)))
    }),
    f("circle_overabove", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_ge(mi(c1.center.y, c1.radius)?, mi(c2.center.y, c2.radius)?)))
    }),
    f("|&>", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_ge(mi(c1.center.y, c1.radius)?, mi(c2.center.y, c2.radius)?)))
    }),
    f("circle_eq", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_eq(c1.ar()?, c2.ar()?)))
    }),
    f("=", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_eq(c1.ar()?, c2.ar()?)))
    }),
    f("circle_lt", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_lt(c1.ar()?, c2.ar()?)))
    }),
    f("<", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_lt(c1.ar()?, c2.ar()?)))
    }),
    f("circle_gt", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_gt(c1.ar()?, c2.ar()?)))
    }),
    f(">", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_gt(c1.ar()?, c2.ar()?)))
    }),
    f("circle_le", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_le(c1.ar()?, c2.ar()?)))
    }),
    f("<=", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_le(c1.ar()?, c2.ar()?)))
    }),
    f("circle_ge", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_ge(c1.ar()?, c2.ar()?)))
    }),
    f(">=", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        Ok(Value::Bool(fp_ge(c1.ar()?, c2.ar()?)))
    }),
    f("circle_ne", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (x, y) = (Circle::read(&a[0])?.ar()?, Circle::read(&a[1])?.ar()?);
        Ok(Value::Bool(x != y && (x - y).abs() > EPSILON))
    }),
    f("<>", &[CIRCLE, CIRCLE], BOOL, |_, a| {
        let (x, y) = (Circle::read(&a[0])?.ar()?, Circle::read(&a[1])?.ar()?);
        Ok(Value::Bool(x != y && (x - y).abs() > EPSILON))
    }),
    f("circle_add_pt", &[CIRCLE, POINT], CIRCLE, |_, a| {
        let c = Circle::read(&a[0])?;
        Ok(Circle { center: c.center.add(&Point::read(&a[1])?)?, ..c }.value())
    }),
    f("+", &[CIRCLE, POINT], CIRCLE, |_, a| {
        let c = Circle::read(&a[0])?;
        Ok(Circle { center: c.center.add(&Point::read(&a[1])?)?, ..c }.value())
    }),
    f("circle_sub_pt", &[CIRCLE, POINT], CIRCLE, |_, a| {
        let c = Circle::read(&a[0])?;
        Ok(Circle { center: c.center.sub(&Point::read(&a[1])?)?, ..c }.value())
    }),
    f("-", &[CIRCLE, POINT], CIRCLE, |_, a| {
        let c = Circle::read(&a[0])?;
        Ok(Circle { center: c.center.sub(&Point::read(&a[1])?)?, ..c }.value())
    }),
    f("circle_mul_pt", &[CIRCLE, POINT], CIRCLE, |_, a| {
        let (c, p) = (Circle::read(&a[0])?, Point::read(&a[1])?);
        Ok(Circle { center: c.center.mul(&p)?, radius: mul(c.radius, hypot(p.x, p.y)?)? }.value())
    }),
    f("*", &[CIRCLE, POINT], CIRCLE, |_, a| {
        let (c, p) = (Circle::read(&a[0])?, Point::read(&a[1])?);
        Ok(Circle { center: c.center.mul(&p)?, radius: mul(c.radius, hypot(p.x, p.y)?)? }.value())
    }),
    f("circle_div_pt", &[CIRCLE, POINT], CIRCLE, |_, a| {
        let (c, p) = (Circle::read(&a[0])?, Point::read(&a[1])?);
        Ok(Circle { center: c.center.div(&p)?, radius: div(c.radius, hypot(p.x, p.y)?)? }.value())
    }),
    f("/", &[CIRCLE, POINT], CIRCLE, |_, a| {
        let (c, p) = (Circle::read(&a[0])?, Point::read(&a[1])?);
        Ok(Circle { center: c.center.div(&p)?, radius: div(c.radius, hypot(p.x, p.y)?)? }.value())
    }),
    f("circle_area", &[CIRCLE], FLOAT8, |_, a| float(Circle::read(&a[0])?.ar())),
    f("area", &[CIRCLE], FLOAT8, |_, a| float(Circle::read(&a[0])?.ar())),
    f("diameter", &[CIRCLE], FLOAT8, |_, a| float(mul(Circle::read(&a[0])?.radius, 2.0))),
    f("radius", &[CIRCLE], FLOAT8, |_, a| Ok(Value::Float8(Circle::read(&a[0])?.radius))),
    f("circle_distance", &[CIRCLE, CIRCLE], FLOAT8, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        float(mi(c1.center.dt(&c2.center)?, pl(c1.radius, c2.radius)?).map(nonnegative))
    }),
    f("<->", &[CIRCLE, CIRCLE], FLOAT8, |_, a| {
        let (c1, c2) = (Circle::read(&a[0])?, Circle::read(&a[1])?);
        float(mi(c1.center.dt(&c2.center)?, pl(c1.radius, c2.radius)?).map(nonnegative))
    }),
    f("circle_contain_pt", &[CIRCLE, POINT], BOOL, |_, a| {
        let c = Circle::read(&a[0])?;
        Ok(Value::Bool(c.center.dt(&Point::read(&a[1])?)? <= c.radius))
    }),
    f("@>", &[CIRCLE, POINT], BOOL, |_, a| {
        let c = Circle::read(&a[0])?;
        Ok(Value::Bool(c.center.dt(&Point::read(&a[1])?)? <= c.radius))
    }),
    f("pt_contained_circle", &[POINT, CIRCLE], BOOL, |_, a| {
        let c = Circle::read(&a[1])?;
        Ok(Value::Bool(c.center.dt(&Point::read(&a[0])?)? <= c.radius))
    }),
    f("<@", &[POINT, CIRCLE], BOOL, |_, a| {
        let c = Circle::read(&a[1])?;
        Ok(Value::Bool(c.center.dt(&Point::read(&a[0])?)? <= c.radius))
    }),
    f("circle_center", &[CIRCLE], POINT, |_, a| Ok(Circle::read(&a[0])?.center.value())),
    f("center", &[CIRCLE], POINT, |_, a| Ok(Circle::read(&a[0])?.center.value())),
    f("point", &[CIRCLE], POINT, |_, a| point_of_circle(a)),
    f("@@", &[CIRCLE], POINT, |_, a| Ok(Circle::read(&a[0])?.center.value())),
    f("circle", &[POINT, FLOAT8], CIRCLE, |_, a| {
        let Value::Float8(radius) = a[1] else { return Err(PgError::internal("a radius that is not float8")) };
        Ok(Circle { center: Point::read(&a[0])?, radius }.value())
    }),
    f("circle", &[BOX], CIRCLE, |_, a| circle_of_box(a)),
    f("circle", &[POLYGON], CIRCLE, |_, a| circle_of_polygon(a)),
];

/// box_of_point converts a point to a box.
fn box_of_point(a: &[Value]) -> Result<Value> {
    let p = Point::read(&a[0])?;
    Ok(GeoBox { high: p, low: p }.value())
}

/// point_of_lseg converts a lseg to a point.
fn point_of_lseg(a: &[Value]) -> Result<Value> {
    Ok(Lseg::read(&a[0])?.center()?.value())
}

/// polygon_of_path converts a path to a polygon.
fn polygon_of_path(a: &[Value]) -> Result<Value> {
    let p = Path::read(&a[0])?;
    if !p.closed {
        return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "open path cannot be converted to polygon"));
    }
    Ok(Polygon::new(p.p).value())
}

/// point_of_box converts a box to a point.
fn point_of_box(a: &[Value]) -> Result<Value> {
    Ok(GeoBox::read(&a[0])?.cn()?.value())
}

/// lseg_of_box converts a box to a lseg.
fn lseg_of_box(a: &[Value]) -> Result<Value> {
    let b = GeoBox::read(&a[0])?;
    Ok(Lseg { p: [b.high, b.low] }.value())
}

/// polygon_of_box converts a box to a polygon.
fn polygon_of_box(a: &[Value]) -> Result<Value> {
    let b = GeoBox::read(&a[0])?;
    Ok(Polygon {
        p: vec![b.low, Point { x: b.low.x, y: b.high.y }, b.high, Point { x: b.high.x, y: b.low.y }],
        bound: GeoBox::construct(&b.high, &b.low),
    }
    .value())
}

/// circle_of_box converts a box to a circle.
fn circle_of_box(a: &[Value]) -> Result<Value> {
    let b = GeoBox::read(&a[0])?;
    let center = b.cn()?;
    Ok(Circle { center, radius: center.dt(&b.high)? }.value())
}

/// point_of_polygon converts a polygon to a point.
fn point_of_polygon(a: &[Value]) -> Result<Value> {
    Ok(Polygon::read(&a[0])?.to_circle()?.center.value())
}

/// path_of_polygon converts a polygon to a path.
fn path_of_polygon(a: &[Value]) -> Result<Value> {
    Ok(Path { closed: true, p: Polygon::read(&a[0])?.p }.value())
}

/// box_of_polygon converts a polygon to a box.
fn box_of_polygon(a: &[Value]) -> Result<Value> {
    Ok(Polygon::read(&a[0])?.bound.value())
}

/// circle_of_polygon converts a polygon to a circle.
fn circle_of_polygon(a: &[Value]) -> Result<Value> {
    Ok(Polygon::read(&a[0])?.to_circle()?.value())
}

/// point_of_circle converts a circle to a point.
fn point_of_circle(a: &[Value]) -> Result<Value> {
    Ok(Circle::read(&a[0])?.center.value())
}

/// box_of_circle converts a circle to a box.
fn box_of_circle(a: &[Value]) -> Result<Value> {
    let c = Circle::read(&a[0])?;
    let delta = div(c.radius, 2f64.sqrt())?;
    Ok(GeoBox {
        high: Point { x: pl(c.center.x, delta)?, y: pl(c.center.y, delta)? },
        low: Point { x: mi(c.center.x, delta)?, y: mi(c.center.y, delta)? },
    }
    .value())
}

/// polygon_of_circle converts a circle to a polygon.
fn polygon_of_circle(a: &[Value]) -> Result<Value> {
    Ok(circle_poly(12, &Circle::read(&a[0])?)?.value())
}

/// Conversion converts a geometric value to another geometric type.
type Conversion = fn(&[Value]) -> Result<Value>;

/// CASTS are the casts between geometric types: the source, the target, the conversion, and whether an assignment may
/// use it without an explicit cast.
const CASTS: &[(u32, u32, Conversion, bool)] = &[
    (POINT, BOX, box_of_point, true),
    (LSEG, POINT, point_of_lseg, false),
    (PATH, POLYGON, polygon_of_path, true),
    (BOX, POINT, point_of_box, false),
    (BOX, LSEG, lseg_of_box, false),
    (BOX, POLYGON, polygon_of_box, true),
    (BOX, CIRCLE, circle_of_box, false),
    (POLYGON, POINT, point_of_polygon, false),
    (POLYGON, PATH, path_of_polygon, true),
    (POLYGON, BOX, box_of_polygon, false),
    (POLYGON, CIRCLE, circle_of_polygon, false),
    (CIRCLE, POINT, point_of_circle, false),
    (CIRCLE, BOX, box_of_circle, false),
    (CIRCLE, POLYGON, polygon_of_circle, false),
];

/// ELEMENT is the hidden function that subscripts a point, segment, box, or line as Postgres subscripts these
/// fixed-length types: as arrays of their coordinates or corners, indexed from zero.
pub(crate) const ELEMENT: &str = "__doltgres_geometry_element";

/// element_type returns the type of the elements that subscripting a geometric type returns, or None for a type
/// that does not support subscripting.
pub(crate) fn element_type(type_oid: u32) -> Option<u32> {
    match type_oid {
        POINT | LINE => Some(FLOAT8),
        LSEG | BOX => Some(POINT),
        _ => None,
    }
}

/// element returns the element of a point, segment, box, or line at an index, or NULL outside it.
fn element(a: &[Value]) -> Result<Value> {
    let Value::Int4(index) = a[1] else { return Err(PgError::internal("a subscript that is not int4")) };
    let Ok(index) = usize::try_from(index) else { return Ok(Value::Null) };
    let Value::Base(base) = &a[0] else { return Err(PgError::internal("a geometric value that is not one")) };
    let values = floats(&base.data);
    Ok(match element_type(base.type_oid) {
        Some(FLOAT8) => values.get(index).map_or(Value::Null, |f| Value::Float8(*f)),
        _ => points(&values).get(index).map_or(Value::Null, |p| p.value()),
    })
}

/// castable reports whether a geometric type has a cast to another that the context allows.
pub(crate) fn castable(from: u32, to: u32, explicit: bool) -> bool {
    CASTS.iter().any(|(source, target, _, assignment)| *source == from && *target == to && (explicit || *assignment))
}

/// cast converts a geometric value to another geometric type as Postgres' casts between them do, returning None when
/// there is no such cast, or none that the context allows.
pub(crate) fn cast(value: &Value, to: u32, explicit: bool) -> Option<Result<Value>> {
    let Value::Base(base) = value else { return None };
    if !castable(base.type_oid, to, explicit) {
        return None;
    }
    let (_, _, conversion, _) = CASTS.iter().find(|(from, target, _, _)| *from == base.type_oid && *target == to)?;
    Some(conversion(std::slice::from_ref(value)))
}
