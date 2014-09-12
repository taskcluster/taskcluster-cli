/**
Live logging / http streaming utilities.
*/

var Promise = require('promise');
var URL = require('url');
var debug = require('debug')('taskcluster-cli:lib:read_log');

// Under normal load the docker-worker can take up to 2s to start the live
// logging server and bind it to a port.
var MAX_REDIRECTS = 20;
var REDIRECT_DELAY = 100;

/**
Open a livelog file for reading... The logic here is fairly specific to how
taskcluster queue returns artifacts.
*/
function openLog(url) {
  return new Promise(function(accept, reject) {
    debug('open log', url);
    var redirects = 0;

    // Attempt to open request...
    function openRequest(url) {
      var client = URL.parse(url).protocol === 'http:' ?
        require('http') :
        require('https');

      client.get(url, function(res) {
        switch (res.statusCode) {
          // Handle redirects by the queue...
          case 303:
            debug('live log redirect | %s', res.headers.location);
            // Keep track of the redirects...
            redirects++;
            if (redirects >= MAX_REDIRECTS) {
              return reject(new Error('Max redirects reached.'))
            }
            // Ensure we purge the body and then setup our redirect...
            res.resume();
            setTimeout(openRequest, REDIRECT_DELAY, res.headers.location);
            break;

          case 200:
            debug('live log success | %s', url);
            // TODO: For maximum robustness we can retry failed requests and
            //       keep track of our current byte offset to send range queries
            //       when the live logging falls over.
            return accept(res);

          default:
            debug('live log error | %s', url);
            return reject(new Error('Unknown server error: ' + res.statusCode));
        }
      });
    }

    openRequest(url);
  });
}

module.exports = openLog;
