function matcher(rows, exp, exceptionKeys, getExceptionIsValid) {
  // Row lengths match
  if (rows.length !== exp.length) {
    console.log("row lengths don't match", rows.length, exp.length);
    return false;
  }
  for (let i = 0; i < rows.length; i++) {
    const rowKeys = Object.keys(rows[i]);
    const expKeys = Object.keys(exp[i]);
    // Row key lengths match
    if (rowKeys.length !== expKeys.length) {
      return false;
    }
    // Row key values match
    for (let j = 0; j < rowKeys.length; j++) {
      const rowKey = rowKeys[j];
      // Check if key has an exception function
      if (exceptionKeys.includes(rowKey)) {
        const isValid = getExceptionIsValid(rows[i], rowKey, exp[i]);
        if (!isValid) {
          console.log("exception was not valid for key:", rowKey);
          return false;
        }
      } else {
        // Compare cell values
        const cellVal = JSON.stringify(rows[i][rowKey]);
        const expCellVal = JSON.stringify(exp[i][rowKey]);
        if (cellVal !== expCellVal) {
          console.log("values don't match", cellVal, expCellVal);
          return false;
        }
      }
    }
  }
  return true;
}

function commitHashIsValid(commit) {
  return commit === "STAGED" || commit === "WORKING" || commit.length === 32;
}

function dateIsValid(date) {
  return JSON.stringify(date).length > 0;
}

export function branchesMatcher(data, exp) {
  const exceptionKeys = ["hash", "latest_commit_date"];

  function getExceptionIsValid(row, key) {
    const val = row[key];
    switch (key) {
      case "hash":
        return commitHashIsValid(val);
      case "latest_commit_date":
        return dateIsValid(val);
      default:
        return false;
    }
  }

  return matcher(data.rows, exp.rows, exceptionKeys, getExceptionIsValid);
}

export function logsMatcher(data, exp) {
  const exceptionKeys = ["commit_hash", "date", "parents"];

  function getExceptionIsValid(row, key, expRow) {
    const val = row[key];
    switch (key) {
      case "commit_hash":
        return commitHashIsValid(val);
      case "date":
        return dateIsValid(val);
      case "parents":
        const numParents = val.split(", ").filter((v) => !!v.length).length;
        const expParents = expRow.parents.length;
        return numParents === expParents;
      default:
        return false;
    }
  }

  return matcher(data.rows, exp.rows, exceptionKeys, getExceptionIsValid);
}

export function mergeBaseMatcher(data) {
  if (data.rows.length !== 1) {
    return false;
  }
  return commitHashIsValid(data.rows[0].dolt_merge_base);
}

export function mergeMatcher(data, exp) {
  if (data.rows.length !== 1) {
    console.log("Rows length not 1", data.rows.length);
    return false;
  }

  const row = data.rows[0].dolt_merge;
  const expRow = exp.rows[0].dolt_merge;

  // Check valid commit hash
  if (!commitHashIsValid(row[0])) {
    console.log("Invalid commit hash", row[0]);
    return false;
  }
  // Check the rest of the fields
  for (let i = 1; i < row.length; i++) {
    if (row[i] !== expRow[i]) {
      console.log("Values don't match", row[i], expRow[i]);
      return false;
    }
  }

  return true;
}
