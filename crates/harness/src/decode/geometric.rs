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

use crate::decode::numeric::float8_text;
use crate::decode::reader::Reader;

/// point_text reads a point and renders it as "(x,y)".
fn point_text(r: &mut Reader<'_>) -> Result<String, String> {
    let x = r.f64()?;
    let y = r.f64()?;
    Ok(format!("({},{})", float8_text(x), float8_text(y)))
}

/// points_text reads a count of points and renders them separated by commas.
fn points_text(r: &mut Reader<'_>) -> Result<String, String> {
    let count = r.i32()?;
    if count < 0 {
        return Err(format!("invalid point count {count}"));
    }
    let points = (0..count).map(|_| point_text(r)).collect::<Result<Vec<_>, _>>()?;
    Ok(points.join(","))
}

/// geometric_text renders a binary point, lseg, path, box, polygon, line, or circle.
pub(crate) fn geometric_text(oid: u32, r: &mut Reader<'_>) -> Result<String, String> {
    Ok(match oid {
        600 => point_text(r)?,
        601 => format!("[{},{}]", point_text(r)?, point_text(r)?),
        602 => {
            let closed = r.u8()? != 0;
            let points = points_text(r)?;
            if closed { format!("({points})") } else { format!("[{points}]") }
        }
        603 => format!("{},{}", point_text(r)?, point_text(r)?),
        604 => format!("({})", points_text(r)?),
        628 => {
            let a = r.f64()?;
            let b = r.f64()?;
            let c = r.f64()?;
            format!("{{{},{},{}}}", float8_text(a), float8_text(b), float8_text(c))
        }
        718 => {
            let center = point_text(r)?;
            let radius = r.f64()?;
            format!("<{center},{}>", float8_text(radius))
        }
        _ => return Err(format!("type {oid} is not geometric")),
    })
}
