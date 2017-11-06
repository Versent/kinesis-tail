const pj = require('../package.json');
const fs = require('fs');

var ver = process.argv[2];
if (!ver) {
	console.log("You need to specify a version number.")
	return
}

pj.version = ver;
pj.scripts.preinstall = "node node/install.js " + ver;
fs.writeFileSync("package.json", JSON.stringify(pj, null, 2));