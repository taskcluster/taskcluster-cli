/**
Load a task or graph file with support for yaml, yml and json...
*/

var fsPath = require('path');
var fs = require('fs');
var yaml = require('js-yaml');

/**
@param {String} pathName to graph or task file.
@return {Object} task or graph.
*/
function loadGraph(pathName) {
  pathName = fsPath.resolve(pathName);

  switch (fsPath.extname(pathName)) {
    // YAML files...
    case '.yaml':
    case '.yml':
      return yaml.safeLoad(fs.readFileSync(pathName, 'utf8'));

    // JSON or JS files...
    default:
      return require(pathName);
  }
}

module.exports = loadGraph;
