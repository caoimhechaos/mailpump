mailpump
========

mailpump is a Go implementation of a distributed mail system. It consists of
a number of different components:

 * smtpump is a thin layer to receive SMTP connections and pass the request
   on to a backend speaking the appropriate mail RPC protocol.

More components will be added at the time they are required.


Requirements
------------

In order to install mailpump, you need to install a few prerequisites:

* golang-go


Installation
------------

Once you have these, just run

    go build

 to complete the build process. You will then receive a number of binaries
 called smtpump-server, etc. which implement the different components of
 the mailpump service.

Once this is done, just invoke the binaries like:

    ./smtpump-service --bind="[::]:25" --web-port="[::1]:8025"

Point your browser to http://localhost:8025/debug/vars to see if it worked.
You can also test the SMTP server by using the command:

    telnet localhost 25


Performance
-----------

No performance tests have been performed yet.


Monitoring
----------

Like any good Go program, pasten exports a few variables under
/debug/vars:

* smtp-num-accepts: total number of connections accepted in the lifetime of
  the SMTP server.
* smtp-accept-errors: map of the different errors which ocurred when
  accepting a request.
* smtp-recent-accept-errors: number of accept errors which ocurred since the
  most recent successful acceptance of a connection.
* smtp-dialog-errors: map of the total number of errors during the SMTP
  dialog, by error type.
* smtp-command-timeouts: total number of times a connection has been
  terminated due to lack of activity from the client.
* smtp-bytes-in: total number of bytes recieved by the SMTP server.
* smtp-bytes-out: total number of bytes sent by the SMTP server.
* smtp-active-connections: number of SMTP connections currently open for
  the server.


Roadmap
-------

For future releases, we are planning to add the following features:

Version 1.0 will feature a full end-to-end implementation of mail delivery
from SMTP to IMAP.

Version 1.1 will feature support for mailing lists.


BUGS
----

Bugs for this project are tracked using ditz in the source code tree itself.
The current state of bug squashing can be viewed at
<http://mailpump.ancient-solutions.com/bugtracker/>

To report bugs, please send a pull request with the ditz bug added to the
project
[caoimhechaos/mailpump](https://github.com/caoimhechaos/mailpump/) on GitHub.
