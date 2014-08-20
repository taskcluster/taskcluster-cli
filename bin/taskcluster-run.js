var fs = require('fs');
var Promise = require('promise');
var slugid = require('slugid');
var dotenv = require('dotenv');
var _ = require('lodash');

var taskcluster = require('taskcluster-client');
var debug = require('debug')('taskcluster-cli:run');
var TaskFactory = require('taskcluster-task-factory/task');
var LogStream = require('taskcluster-logstream');

var Listener = taskcluster.Listener;
var queueEvents = new taskcluster.QueueEvents;
var queue = new taskcluster.Queue();

var listener;
var taskComplete = false;

var LOG_NAME = 'public/logs/terminal_live.log';
var taskOwner = process.env.TASKCLUSTER_TASK_OWNER;

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
  var image = args._[0];
  var command = ['/bin/bash', '-cvex'].concat(args._.slice(1));
}

if(!taskOwner) {
  console.error('Error: TASKCLUSTER_TASK_OWNER not specified\n');
  console.log(yargs.help());
  process.exit(1);
}

// Command line arguments should take precedence over those listed in the env file.
// Allows selective overwriting of env variables
var env = {};
if (args.env) {
  var joinedEnvs = args.env.join('\n')
  env = dotenv.parse(joinedEnvs);
}

if (args.envFile) {
  if (fs.existsSync(args.envFile)) {
    var file = fs.readFileSync(args.envFile);
    env = _.defaults(env, dotenv.parse(file));
  } else {
    console.error('Environment file does not exist at the location provided.');
    process.exit(1);
  }
}

function endWhenTasksComplete(stream) {
  if (!stream && taskComplete) {
    listener.close();
  } else {
    setTimeout(function (stream) { endWhenTasksComplete(stream); }, 1);
  }
}

function displayLog(taskId, runId) {
  var signedUrl = queue.buildSignedUrl(
      queue.getArtifact, taskId, runId, LOG_NAME, {expiration: 60 * 100}
  );
  var stream = new LogStream(signedUrl);
  stream.pipe(process.stdout);
  endWhenTasksComplete(stream);
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
      if (message.exchange.indexOf('artifact-created') > -1 && payload.artifact.name == LOG_NAME) {
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
        'maxRunTime': 600
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
