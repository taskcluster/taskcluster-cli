#!/usr/bin/env node
var fs = require('fs');
var Promise = require('promise');
var slugid = require('slugid');
var dotenv = require('dotenv');
var readLog = require('../lib/read_log');

var taskcluster = require('taskcluster-client');
var debug = require('debug')('taskcluster-cli:run');
var TaskFactory = require('taskcluster-task-factory/task');
var LogStream = require('taskcluster-logstream');
var Promise = require('promise');
var URL = require('url');

var Listener = taskcluster.Listener;
var queueEvents = new taskcluster.QueueEvents;
var queue = new taskcluster.Queue();

var listener;
var taskComplete = false;

var LOG_NAME = 'public/logs/live.log';
var taskOwner = process.env.TASKCLUSTER_TASK_OWNER || process.env.EMAIL;

var yargs = require('yargs')
  .usage('Run a task within the task cluster')
  .example('taskcluster run', '-e BUILD_OPTS=debug ubuntu:14.04 \'make -j=5\'')
  .demand(['provisioner-id', 'worker-type'])
  .alias('e', 'env')
  .describe('e', 'Environment variable to apply on worker')
  .describe('env-file', 'Environment variable file to apply on worker')
  .describe('provisioner-id', 'Provisioner ID. Example: aws-provisioner')
  .describe('worker-type', 'Worker Type. Example: \'cli\'')

var args = yargs.argv

if(args._.length < 2) {
  console.error('Error: Must supply an image and command\n')
  console.log(yargs.help());
  process.exit(1);
} else {
  // Docker image.
  var image = args._[0];
  // Arguments for docker cmd (note that we do not specify a shell here very
  // similar to how docker run does not specify a default shell for commands)
  var command = args._.slice(1);
}

if(!taskOwner) {
  console.error('Error: TASKCLUSTER_TASK_OWNER environment variable not configured.\n');
  console.log(yargs.help());
  process.exit(1);
}

// Command line arguments should take precedence over those listed in the env file.
// Allows selective overwriting of env variables
var env = {};
if (args.envFile) {
  if (fs.existsSync(args.envFile)) {
    env = dotenv.parse(fs.readFileSync(args.envFile));
  } else {
    console.error('Environment file does not exist at the location provided.');
    process.exit(1);
  }
}
if (args.env) {
  var cli_envs = args.env;
  if (args.env instanceof Array) {
    cli_envs = args.env.join('\n');
  }
  var parsed_envs = dotenv.parse(cli_envs);
  for (var env_name in parsed_envs) {
    env[env_name] = parsed_envs[env_name];
  }
}

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
  var url = queue.buildUrl(queue.getArtifact, taskId, runId, LOG_NAME);

  return readLog(url).then(function(stream) {
    stream.pipe(process.stdout);
    endWhenTasksComplete(stream);
  }).catch(function(err) {
    console.error("Could not open stream: ", err)
  });
}

function handleEvent(message) {
  var payload = message.payload;
  var taskId = payload.status.taskId;
  var runId = payload.runId;

  switch (payload.status.state) {
    case 'pending':
      console.log("Task Created.\nTask ID: %s\nTask State: Pending", taskId);
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
        console.log('Task Completed');
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
      console.log(result);
      debug("Message: %j", message);
  }
}

function buildTaskRequest(taskId) {
    var creationDate = new Date();
    var deadlineDate = new Date(creationDate)
    deadlineDate.setHours(deadlineDate.getHours() + 24);

    var task = {
      'provisionerId': args['provisioner-id'],
      'workerType': args['worker-type'],
      'schedulerId': 'taskcluster-cli',
      'taskGroupId': taskId,
      'routes': [],
      'retries': 1,
      'priority': 5,
      'created': creationDate,
      'deadline': deadlineDate,
      'scopes': [],
      'payload': {
        'image': image,
        'command': command,
        'env': env,
        'maxRunTime': 7200 // two hours...
      },
      'metadata': {
        'owner': taskOwner,
        'name': '',
        'description': '',
        'source': 'http://localhost'
      },
      'tags': {
        'misc_info': 'task created by cli'
      }
    };
    debug("Task Payload: %j", task);

  return task;
}

function createListener(connectionString, taskId, eventHandler) {
  listener = new Listener({
    connectionString: connectionString,
  });
  listener.bind(queueEvents.taskPending({taskId: taskId}));
  listener.bind(queueEvents.taskRunning({taskId: taskId}));
  listener.bind(queueEvents.taskCompleted({taskId: taskId}));
  listener.bind(queueEvents.artifactCreated({taskId: taskId}));
  listener.on('message', handleEvent);
  return listener;
}

queue.getAMQPConnectionString().then(function(result) {
  var taskId = slugid.v4();
  var task = buildTaskRequest(taskId);
  var listener = createListener(result.url, taskId, handleEvent);

  return listener.resume().then(function () {
    return queue.createTask(taskId, task)
  });
}).catch(function(error) {
  if(error.body) {
    console.error("Message: %s", error.body.message);
    console.error("Error: %j", error.body.error);
  } else {
    console.log(error);
  }
  if (error.stack) {;
    console.error(error.stack);
  }
  process.exit(1);
});
