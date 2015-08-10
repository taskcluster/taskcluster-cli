var taskcluster = require('taskcluster-client');
var Promise = require('promise');
var DockerClient = require('docker-exec-websocket-server').DockerExecClient;
var https = require('https');
var url = require('url');
var _ = require('lodash');
var assert = require('assert');
require('../lib/config').load();

var INTERACTIVE_SOCKET_NAME = 'private/docker-worker/interactive.sock';

var yargs = require('yargs')
  .usage('SSH into task container')
  .example('taskcluster ssh', 'taskId')

var taskId = yargs.argv._[0];
assert(taskId, "Missing taskId");

var queue = new taskcluster.Queue();

queue.status(taskId).then(function(result) {
  // Get run and check state
  var run = _.last(result.status.runs);
  console.log("Latest run: " + run.runId);
  if (run.state !== 'running') {
    console.log("Task must be running for us to connect!");
    process.exit(1);
  }

  // Build URL for the interactive socket
  var signedUrl = queue.buildSignedUrl(
    queue.getArtifact, taskId, run.runId, INTERACTIVE_SOCKET_NAME
  );

  var req = https.get(url.parse(signedUrl));
  return new Promise(function(accept, reject) {
    req.on('response', accept);
    req.on('error', reject);
  });
}).then(function(res) {
  assert(res.statusCode === 303, "Expected a 303 redirect");
  var socketUrl = res.headers.location;

  var client = new DockerClient({
    url: socketUrl,
    tty: true,
    command: ['/.taskclusterutils/busybox', 'sh'],
  });

  return client.execute().then(function() {
    // Set terminal size and update when terminal is resized
    client.resize(process.stdout.rows, process.stdout.columns);
    process.stdout.on('resize', function() {
      client.resize(process.stdout.rows, process.stdout.columns);
    });

    // Set raw mode and pipe to/from client
    process.stdin.setRawMode(true);
    process.stdin.pipe(client.stdin);
    client.stdout.pipe(process.stdout);
    client.stderr.pipe(process.stderr);

    return new Promise(function(accept, reject) {
      client.on('exit', accept);
      client.on('error', reject);
    });
  }).then(function(code) {
    client.close();
    process.exit(code);
  });
}).catch(function(err) {
  console.log("Error: " + err.stack);
  process.exit(1);
});