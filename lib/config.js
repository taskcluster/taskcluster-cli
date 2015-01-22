var AppDirectory = require('appdirectory');
var fs = require('fs');
var path = require('path');
var taskcluster = require('taskcluster-client');
var mkdirp = require('mkdirp');
var debug = require('debug')('taskcluster-cli:config');

var dirs = new AppDirectory('taskcluster-cli');
exports.cfgFile = path.join(dirs.userConfig(), 'credentials.json');

exports.load = function() {
  try {
    if (fs.existsSync(exports.cfgFile)) {
      config = JSON.parse(fs.readFileSync(exports.cfgFile));
      taskcluster.config({credentials: config});
      if (config.certificate) {
        var cert = config.certificate;
        if (typeof(cert) === 'string') {
          cert = JSON.parse(cert);
        }
        if (cert.expiry < new Date().getTime()) {
          console.log("Temporary credentials expired");
          console.log("Login with `taskcluster login`");
          process.exit(1);
        }
      }
    }
  }
  catch(err) {
    debug("Failed to loading configuration");
  }
};

exports.save = function(credentials) {
  mkdirp.sync(dirs.userConfig());
  fs.writeFileSync(exports.cfgFile, JSON.stringify(credentials, null, 2), {
    encoding: 'utf8'
  });
};