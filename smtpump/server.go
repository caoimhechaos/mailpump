/**
 * (c) 2014, Caoimhe Chaos <caoimhechaos@protonmail.com>,
 *	     Ancient Solutions. All rights reserved.
 *
 * Redistribution and use in source  and binary forms, with or without
 * modification, are permitted  provided that the following conditions
 * are met:
 *
 * * Redistributions of  source code  must retain the  above copyright
 *   notice, this list of conditions and the following disclaimer.
 * * Redistributions in binary form must reproduce the above copyright
 *   notice, this  list of conditions and the  following disclaimer in
 *   the  documentation  and/or  other  materials  provided  with  the
 *   distribution.
 * * Neither  the  name  of  Ancient Solutions  nor  the  name  of its
 *   contributors may  be used to endorse or  promote products derived
 *   from this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * "AS IS"  AND ANY EXPRESS  OR IMPLIED WARRANTIES  OF MERCHANTABILITY
 * AND FITNESS  FOR A PARTICULAR  PURPOSE ARE DISCLAIMED. IN  NO EVENT
 * SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT,
 * INDIRECT, INCIDENTAL, SPECIAL,  EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED  TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 * SERVICES; LOSS OF USE,  DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
 * HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
 * STRICT  LIABILITY,  OR  TORT  (INCLUDING NEGLIGENCE  OR  OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED
 * OF THE POSSIBILITY OF SUCH DAMAGE.
 */

// Connection handling part of an SMTP server for mailpump.
package smtpump

import (
	"expvar"
	"net"
)

var smtp_num_accepts = expvar.NewInt("smtp-num-accepts")
var smtp_accept_errors = expvar.NewMap("smtp-accept-errors")
var smtp_recent_accept_errors = expvar.NewInt("smtp-recent-accept-errors")

// Structure to hold all data required for an active server.
type SMTPServer struct {
	callback SmtpReceiver
	listener net.Listener
}

// Create a new SMTP server listening on the address "laddr" with the
// protocol "net". Any callbacks will be done on "callback".
func NewSMTPServer(netname, laddr string, callback SmtpReceiver) (
	*SMTPServer, error) {
	var srv *SMTPServer
	var l net.Listener
	var err error

	l, err = net.Listen(netname, laddr)
	if err != nil {
		return nil, err
	}
	srv = &SMTPServer{
		callback: callback,
		listener: l,
	}

	go srv.waitForConnections()
	return srv, nil
}

// Accept new connections and start processing input on them.
func (self *SMTPServer) waitForConnections() {
	for {
		var c net.Conn
		var err error

		c, err = self.listener.Accept()
		if err == nil {
			smtp_num_accepts.Add(1)
			smtp_recent_accept_errors.Set(0)
			newSmtpConnection(c, self.callback)
		} else {
			smtp_accept_errors.Add(err.Error(), 1)
			smtp_recent_accept_errors.Add(1)
		}
	}
}
