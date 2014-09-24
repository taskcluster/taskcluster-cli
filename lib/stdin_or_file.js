/**
Utility which loads a file from stdin or processed through the `load` utility.
*/

var load = require('./load');
var fs = require('fs');
var Promise = require('promise');
var util = require('util');

function handleStdinOrFile(pathName) {
  return new Promise(function(accept, reject) {
    // First check to see if we can load a file...
    if (pathName) {
      if (!fs.existsSync(pathName)) {
        return reject(
          new Error('Task or graph not found at path: ' + pathName)
        );
      }
      return accept(load(pathName));
    }

    // Load the file from stdin...
    var content = '';
    process.stdin.setEncoding('utf8');
    process.stdin.on('readable', function() {
      var chunk;
      while (chunk = process.stdin.read()) content += chunk;
    });
    process.stdin.once('error', reject)
    process.stdin.once('end', function() {
      var object;
      try {
        accept(JSON.parse(content));
      } catch(e) {
        reject(new Error(
          util.format("Failed to read JSON from stdin: %s", e)
        ));
      }
    });
  });
}

module.exports = handleStdinOrFile;
