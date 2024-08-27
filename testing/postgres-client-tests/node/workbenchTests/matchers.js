function matcher(rows, exp, exceptionKeys, getExceptionIsValid) {
  // Row lengths match
  if (rows.length !== exp.length) {
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
          console.log("exception was not valid for key", rowKey);
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

export function branchesMatcher(rows, exp) {
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

  return matcher(rows, exp, exceptionKeys, getExceptionIsValid);
}
