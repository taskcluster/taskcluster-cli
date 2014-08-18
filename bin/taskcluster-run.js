var taskcluster = require('taskcluster-client');
var slugid = require('slugid');
var debug = require('debug')('taskcluster-client');
var co = require('co');

var LogStream = require('taskcluster-logstream');
var Listener = require('taskcluster-client').Listener;
var Promise = require('promise');
var TaskFactory = require('taskcluster-task-factory/task');

var queueEvents = new (require('taskcluster-client').QueueEvents);
var schedulerEvents = new (require('taskcluster-client').SchedulerEvents);
var queue = new taskcluster.Queue();
var image = process.argv[2];
var commands = process.argv.slice(3);

var LOG_NAME = 'public/logs/terminal_live.log';

function waitForEvent(listener, event) {
  return new Promise(function(accept, reject) {
    listener.on(event, function(message) {
      accept(message);
    });
  });
}

function* createTask(taskId, taskConfig) {
  // XXX: This is just a hack really so the validator does not complain.
  taskConfig.taskGroupId = taskId;
  var task = TaskFactory.create(taskConfig);
  debug('post to queue %j', task);
  return yield queue.createTask(taskId, task);
}

function displayLog(taskId, runId) {
  var signedUrl = queue.buildSignedUrl(
    queue.getArtifact, taskId, runId, LOG_NAME, { expiration: 60 * 100 }
  );
  var stream = new LogStream(signedUrl);
  stream.pipe(process.stdout);
}

function* hasLog(taskId, runId) {
  var list = yield queue.listArtifacts(taskId, runId);
  return list.artifacts.some(function(item) {
    return item.name === LOG_NAME;
  });
}

function* handleEvents(listener) {
  var event;
  while (event = yield waitForEvent(listener, 'message')) {
    var payload = event.payload;
    var taskId = payload.status.taskId;
    var runId = payload.runId;

    switch (payload.status.state) {
      case 'completed':
        yield listener.close();
        break;
      case 'running':
        var list;
        // wait until we get artifacts...
        while (!(yield hasLog(taskId, runId))) {}
        displayLog(taskId, runId);
    }
  }
}

co(function* () {
  // Be lazy and terrible and assume you want shell style...
  if (commands) commands = ['/bin/sh', '-c'].concat(commands);

  var taskId = slugid.v4();

  // Create and bind the listener which will notify us when the worker
  // completes a task.
  var listener = new Listener({
    connectionString: (yield queue.getAMQPConnectionString()).url
  });

  yield listener.bind(queueEvents.taskRunning({
    taskId: taskId
  }));

  yield listener.bind(queueEvents.taskCompleted({
    taskId: taskId
  }));

  yield listener.connect();
  yield listener.resume();

  // Begin listening at the same time we create the task to ensure we get the
  // message at the correct time.
  var creation = yield [
    handleEvents(listener),
    createTask(taskId, {
      workerType: 'v2',
      provisionerId: 'aws-provisioner',
      schedulerId: 'taskcluster-cli',

      metadata: {
        owner: 'jlal@mozilla.com'
      },

      payload: {
        image: image,
        command: commands
      }
    }),
  ];

  try {
    yield listener.close();
  } catch(e) {
    console.log('error during close:', e);
  }
})(function(err) {
  console.log('ERR: ', err);
  process.exit(1);
});
