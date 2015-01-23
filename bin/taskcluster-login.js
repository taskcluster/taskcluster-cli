#!/usr/bin/env node
var http = require('http');
var fs = require('fs');
var open = require('open');
var url = require('url');
var querystring = require('querystring');
var config = require('../lib/config');

var server = http.createServer(function (req, res) {
  res.writeHead(200, {'Content-Type': 'text/html'});
  var query = url.parse(req.url, true).query;
  config.save({
    clientId:     query.clientId,
    accessToken:  query.accessToken,
    certificate:  query.certificate
  });
  console.log("Saved credentials in: " + config.cfgFile);
  res.end("<h1>Login Successful</h1><br>You can close this window now...");
  req.connection.destroy();
  server.close();
}).listen(0, function() {
  var port = server.address().port;
  console.log("Listening for credentials on port: " + port);
  var params = {
    target:       "http://localhost:" + port,
    description:  "`taskcluster-cli` is currently listening on port `" +
                  port + "` for credentials, and will save them in " +
                  "`" + config.cfgFile + "`."
  };
  open("https://auth.taskcluster.net/?" + querystring.stringify(params));
});
