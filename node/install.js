const os = require('os');
const fs = require('fs');
const https = require('https');
const zlib = require('zlib');

const repo = 'Versent';
const project = 'kinesis-tail';

const ver = process.argv[2];
const platform = os.platform();
const arch = os.arch();
const zipfile = `${project}-${ver}-${platform}-${arch}.gz`;
const repoBaseURL = `https://github.com/${repo}/${project}/releases/download`;
const url = `${repoBaseURL}/${ver}/${zipfile}`;
let retry = 0;
const writeto = `./node/bin/${project}`;

// create a bin directory if one doesn't exist
fs.access('./node/bin', (err) => {
  if (err) {
    if (err.code === 'ENOENT') {
      console.log('creating bin directory');
      fs.mkdirSync('./node/bin');
    } else {
      throw err;
    }
  } else {
    console.log('node/bin directory already exists');
  }
});

function saveAndUnzip(response, uri, dest) {
  console.log('Extracting ...');

  const file = fs.createWriteStream(dest);
  response.pipe(zlib.createGunzip()).pipe(file);

  file.on('finish', () => {
    // close the file
    file.close(() => {
      console.log('Done!');
    });
  });

  // something went wrong.  unlink the file.
  file.on('error', () => {
    fs.unlink(file);
    console.log('Something went wrong while downloading or unzipping ...Retrying');
    retry += 1;
    download(uri, dest);
  });
}

function download(uri, dest) {
  if (retry === 3) {
    console.log('Retried 3 times.  Sorry.');
    return;
  }
  console.log(`Downloading ${platform} ${arch} executable...`);
  https.get(uri, (res) => {
    // github will redirect
    let newURI = uri;

    if (res.statusCode > 300 && res.statusCode < 400 && res.headers.location) {
      newURI = res.headers.location;
      console.log('Redirecting and retrying ...');
      download(newURI, dest);
    } else if (res.statusCode === 200) {
      saveAndUnzip(res, newURI, dest);
    } else {
      // something bad happened, return the status code and exit.
      console.log(`could not download zip archive- ${res.statusCode} ...Retrying`);
      retry += 1;
      download(newURI, dest);
    }
  });
}

download(url, writeto);
