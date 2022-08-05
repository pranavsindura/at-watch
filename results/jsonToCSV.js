const fs = require("fs")
const path = require("path")

function pivot(arr) {
  const mp = new Map();

  function setValue(a, path, val) {
    if (Object(val) !== val) {
      // primitive value
      const pathStr = path.join(".");
      const i = (mp.has(pathStr) ? mp : mp.set(pathStr, mp.size)).get(pathStr);
      a[i] = val;
    } else {
      for (const key of Object.keys(val)) {
        setValue(a, key === "0" ? path : path.concat(key), val[key]);
      }
    }
    return a;
  }

  const result = arr.map((obj) => setValue([], [], obj));
  return [[...mp.keys()], ...result];
}

function toCsv(arr) {
  return arr
    .map((row) =>
      row.map((val) => (isNaN(val) ? JSON.stringify(val) : +val)).join(",")
    )
    .join("\n");
}

function main() {
  const fileName = process.argv[2]
  if(!fileName) {
    console.log('filename not provided')
    process.exit(1)
  }
  const filePath = path.join(__dirname, fileName);
  console.log("reading", filePath);
  const trades = JSON.parse(fs.readFileSync(filePath).toString());
  
  const csvPath = path.join(__dirname, fileName.split('.')[0] + '.csv')
  console.log("writing", csvPath);
  fs.writeFileSync(csvPath, toCsv(pivot(trades)));
}

main()