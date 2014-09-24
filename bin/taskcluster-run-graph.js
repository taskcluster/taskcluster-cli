#!/usr/bin/env node
var fs = require('fs');
var fsPath = require('path');
var slugid = require('slugid');
var open = require('open');
var yaml = require('js-yaml');

var taskcluster = require('taskcluster-client');
var debug = require('debug')('taskcluster-cli:run');

var scheduler = new taskcluster.Scheduler();

var yargs = require('yargs')
  .usage('Create a taskcluster graph from a file')
  .example('taskcluster run-graph', 'graph.json')
  .example('taskcluster run-graph --no-open', 'graph.json')
  .options('open', {
    boolean: true,
    default: true,
    describe: 'When true open the graph in the task graph inspector',
  })
  .options('dump', {
    boolean: true,
    default: false,
    describe: 'When true dump the task graph to stdout',
  });

/**
Figure out what format the graph is then load it and return the object.
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

function runGraph(graphContent) {
  // Most fields can be directly embedded in the graph but some must be
  // dynamically generated such as the taskId and deadline...
  graphContent.tasks = graphContent.tasks.map(function(node) {
    node.taskId = node.taskId || slugid.v4();
    node.task.created = node.created || new Date().toJSON();
    if (!node.task.deadline) {
      node.task.deadline = new Date();
      node.task.deadline.setHours(node.task.deadline.getHours() + 24);
    }
    return node;
  });

  // Schedule the task
  var graphId = slugid.v4();
  var inspectorUrl = 'http://docs.taskcluster.net/tools/task-graph-inspector/#';
  inspectorUrl += graphId;

  // Dump the graph first so its easy to inspect for errors...
  if (args.dump) {
    process.stderr.write(JSON.stringify(graphContent, null, 2));
  }

  scheduler.createTaskGraph(graphId, graphContent).then(function() {
    if (args.open) {
      open(inspectorUrl);
    }
    process.stdout.write(graphId);
  }).catch(function(err) {
    console.error('Taskgraph creation error', err.toString());
    process.stderr.write(JSON.stringify(err.body, null, 2));
    process.exit(1);
  });
}

var args = yargs.argv
var graph = args._[0];

if (graph && !fs.existsSync(graph)) {
  console.error('Error: graph file "%s" does not exist', graph);
  console.error(yargs.help());
  process.exit(1);
}

if (graph) {
  // Run graph from a file...
  runGraph(loadGraph(graph));
} else {
  var graphContent = '';
  // Run graph from stdin...
  process.stdin.setEncoding('utf8');
  process.stdin.on('readable', function() {
    var chunk;
    while (chunk = process.stdin.read()) graphContent += chunk;
  });
  process.stdin.once('end', function() {
    var graph;
    try {
      graph = JSON.parse(graphContent);
      runGraph(graph);
    } catch(e) {
      console.error("Failed to read JSON from stdin: %j", graph);
      console.error(yargs.help());
      process.exit(1);
    }
  });
}
