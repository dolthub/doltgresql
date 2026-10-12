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

//! Splits the output of psql run with `-a` into the output of each unit of its script.

use crate::script::Unit;

/// split returns the output of each unit, found between the unit's echoed lines and the next unit's. Output before
/// the first unit's echo belongs to it, and the last unit's output runs to the end. A unit has no output when its echo
/// or the next unit's is missing, as when psql stopped early.
pub fn split(output: &str, units: &[Unit]) -> Vec<Option<String>> {
    let echoes: Vec<String> = units.iter().map(|u| u.echo.join("\n")).collect();
    let mut outputs = vec![None; units.len()];
    let Some(mut position) = echoes.first().and_then(|echo| find(output, echo, 0)) else { return outputs };
    let prefix = &output[..position];
    for (k, echo) in echoes.iter().enumerate() {
        let start = (position + echo.len() + 1).min(output.len());
        let next = match echoes.get(k + 1) {
            Some(next) => find(output, next, start),
            None => Some(output.len()),
        };
        let Some(next) = next else { break };
        let body = &output[start..next.max(start)];
        outputs[k] = Some(if k == 0 { format!("{prefix}{body}") } else { body.to_string() });
        position = next;
    }
    outputs
}

/// find returns where echoed lines next appear in the output, at or after the start, ending a line. They may begin
/// within a line, after output that ended without a newline, as `\echo -n` prints.
fn find(output: &str, echo: &str, start: usize) -> Option<usize> {
    let ends_line = |q: usize| matches!(output.as_bytes().get(q + echo.len()), None | Some(b'\n'));
    output[start..].match_indices(echo).map(|(i, _)| start + i).find(|&q| ends_line(q))
}
