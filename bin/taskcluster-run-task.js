#!/usr/bin/env node
var fs = require('fs');
var slugid = require('slugid');
var readLog = require('../lib/read_log');
var loadStdinOrFile = require('../lib/stdin_or_file');

var taskcluster = require('taskcluster-client');
var debug = require('debug')('taskcluster-cli:run');
var config = require('../lib/config');
config.load();

var Listener = taskcluster.WebListener;
var queueEvents = new taskcluster.QueueEvents;
var queue = new taskcluster.Queue();

var listener;
var taskComplete = false;

var LOG_NAME = 'public/logs/live.log';

var yargs = require('yargs')
  .usage('Run a task within the task cluster')
  .options('verbose', {
    boolean: true,
    default: false,
    describe: 'Log additional details'
  })
  .example('taskcluster run-task', 'task.json')
  .example('taskcluster run-task', 'task.yml')
  .example('cat task.json | taskcluster run-task')

var args = yargs.argv;

var verbose = (args.verbose) ? console.error : function(){};

function endWhenTasksComplete(stream) {
  if (!stream && taskComplete) {
    listener.close();
  } else {
    setTimeout(function (stream) { endWhenTasksComplete(stream); }, 1);
  }
}

var showingLog = false;
function displayLog(taskId, runId) {
  // The live log url is updated frequently as we update it from the live log
  // endpoint to the actual backing file on s3. We only care about it the first
  // time we log it.
  if (showingLog) return
  showingLog = true; // don't show more then one log at once...
  var url = queue.buildUrl(queue.getArtifact, taskId, runId, LOG_NAME);

  return readLog(url).then(function(stream) {
    stream.pipe(process.stdout);
    endWhenTasksComplete(stream);
  }).catch(function(err) {
    console.error('Could not open stream: ', err);
  });
}

function handleEvent(message) {
  var payload = message.payload;
  var taskId = payload.status.taskId;
  var runId = payload.runId;

  switch (payload.status.state) {
    case 'pending':
      verbose('Task Created.\nTask ID: %s\nTask State: Pending', taskId);
      break;
    case 'running':
      if (
        message.exchange.indexOf('artifact-created') > -1 &&
        payload.artifact.name == LOG_NAME
      ) {
        displayLog(taskId, runId);
      }
      break;
    case 'completed':
      if (payload.success) {
        verbose('Task Completed');
        taskComplete = true;
        break;
      }
    case 'failed':
      if (payload.status.state == 'completed') {
        result = 'Task Completed Unsuccessfully';
      } else {
        taskComplete = true;
        result = 'Task Failed';
      }
      verbose(result);
      debug('Message: %j', message);
  }
}

function runTask(task) {
  var taskId = slugid.v4();

  listener = new Listener();
  listener.bind(queueEvents.taskPending({taskId: taskId}));
  listener.bind(queueEvents.taskRunning({taskId: taskId}));
  listener.bind(queueEvents.taskCompleted({taskId: taskId}));
  listener.bind(queueEvents.artifactCreated({taskId: taskId}));
  listener.on('message', handleEvent);

  listener.resume().then(function () {
    return queue.createTask(taskId, task);
  }).catch(function(error) {
    if (error.body) {
      console.error('Message: %s', error.body.message);
      console.error('Error: %j', error.body.error);
    } else {
      console.log(error);
    }
    if (error.stack) {;
      console.error(error.stack);
    }
    process.exit(1);
  });
}

loadStdinOrFile(args._[0]).then(function(task) {
  return runTask(task);
}).catch(function(err) {
  console.error(err.toString());
  console.error(yargs.help());
  process.exit(1);
});
