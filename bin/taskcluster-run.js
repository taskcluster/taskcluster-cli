#!/usr/bin/env node
var fs = require('fs');
var slugid = require('slugid');
var dotenv = require('dotenv');

var debug = require('debug')('taskcluster-cli:run');

var yargs = require('yargs')
  .usage('Run a task within the task cluster')
  .example('taskcluster run', '-e BUILD_OPTS=debug ubuntu:14.04 \'make -j=5\'')
  .demand(['provisioner-id', 'worker-type', 'owner'])
  .alias('e', 'env')
  .describe('e', 'Environment variable to apply on worker')
  .describe('env-file', 'Environment variable file to apply on worker')
  .describe('provisioner-id', 'Provisioner ID. Example: aws-provisioner')
  .describe('worker-type', 'Worker Type. Example: \'cli\'')
  .options('verbose', {
    boolean: true,
    default: true,
    describe: 'Log additional details'
  })
  .options('owner', {
    default: process.env.TASKCLUSTER_TASK_OWNER || process.env.EMAIL,
    describe: 'Who owns this task (email address)'
  });

var args = yargs.argv

if (args._.length < 2) {
  console.error('Error: Must supply an image and command\n')
  console.log(yargs.help());
  process.exit(1);
} else {
  // Docker image.
  var image = args._[0];
  // Arguments for docker cmd (note that we do not specify a shell here very
  // similar to how docker run does not specify a default shell for commands)
  var command = args._.slice(1).toString().split(" ");
}

// Command line arguments should take precedence over those listed in the env 
// file. Allows selective overwriting of env variables.
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
  var cliEnvs = args.env;
  if (args.env instanceof Array) {
    cliEnvs = args.env.join('\n');
  }
  var parsedEnvs = dotenv.parse(cliEnvs);
  for (var envName in parsedEnvs) {
    env[envName] = parsedEnvs[envName];
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
      'owner': args.owner,
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

var taskId = slugid.v4();
var task = buildTaskRequest(taskId);
var procArgs = ['run-task'];

if (args.verbose) {
  procArgs.push('--verbose');
}

var proc = require('child_process').spawn(
  'taskcluster',
  procArgs,
  { stdio: 'pipe', env: process.env }
);

// Yield ownership of this process to the result of the child.
proc.stdout.pipe(process.stdout);
proc.stderr.pipe(process.stderr);
proc.once('exit', process.exit);

proc.stdin.write(JSON.stringify(task));
proc.stdin.end();
