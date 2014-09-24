#!/usr/bin/env node
var fs = require('fs');
var fsPath = require('path');
var slugid = require('slugid');
var open = require('open');
var loadStdinOrFile = require('../lib/stdin_or_file');

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

var args = yargs.argv

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


loadStdinOrFile(args._[0]).then(function(contents) {
  return runGraph(contents);
}).catch(function(err) {
  console.error(err.toString());
  console.error(yargs.help());
  process.exit(1);
});
