#!/usr/bin/env node
var spawn = require('child_process').spawn;
var path = require('path');

var yargs = require('yargs')
  .usage(
'Usage: taskcluster COMMAND [arg...]\n\n' +
'Commands:\n' +
'  run        Run a task via a docker run like interface\n' +
'  run-graph  Run a task graph within taskcluster\n' +
'  run-task   Run a task within taskcluster\n' +
'  login      Login with taskcluster\n'
);

var argv = yargs.argv;
var allowedCommands = ['run', 'run-graph', 'run-task', 'login', 'ssh'];

if (!argv._.length) {
  console.log(yargs.help());
  process.exit(1);
}

var command = argv._[0];
if (allowedCommands.indexOf(command) == -1) {
  console.error('Error: Command not found: %s\n', command);
  console.log(yargs.help());
  process.exit(1);
}

var bin = argv['$0'] + '-' + argv._[0];
var args = process.argv.slice(3);

var proc = spawn(bin, args, {stdio: 'inherit', customFds: [0,1,2] });

proc.on('error', function(err){
  if (err.code == "ENOENT") {
    console.error('\n  %s(1) does not exist, try --help\n', bin);
  } else if (err.code == "EACCES") {
    console.error('\n  %s(1) not executable. try chmod or run with root\n', bin);
  }
});



