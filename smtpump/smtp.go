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

// SMTP server implementation for mailpump.
package smtpump

import (
	"expvar"
	"io"
	"net"
	"net/textproto"
	"strings"
	"time"

	"ancient-solutions.com/mailpump"
)

var smtp_dialog_errors = expvar.NewMap("smtp-dialog-errors")
var smtp_command_timeouts = expvar.NewInt("smtp-command-timeouts")
var smtp_bytes_in = expvar.NewInt("smtp-bytes-in")
var smtp_bytes_out = expvar.NewInt("smtp-bytes-out")
var smtp_active_connections = expvar.NewInt("smtp-active-connections")

// Generic SMTP return code; indicates what the server should respond
// to the client.
type SmtpReturnCode struct {
	// SMTP error/success code to return to the client.
	Code int

	// Text representation of the message the client should get.
	Message string

	// Whether or not to terminate the connection after the command.
	Terminate bool
}

// Callback implementation for SMTP servers.
type SmtpReceiver interface {
	// Invoked when a new connection is opened.
	ConnectionOpened(conn *SmtpConnection, peer net.Addr) SmtpReturnCode

	// Invoked when the connection has been closed.
	ConnectionClosed(conn *SmtpConnection)

	// Invoked when a HELO is received from the server.
	Helo(conn *SmtpConnection, hostname string) SmtpReturnCode

	// Invoked when a MAIL From command is received.
	MailFrom(conn *SmtpConnection, sender string) SmtpReturnCode

	// Invoked when a RCPT To command is received.
	RcptTo(conn *SmtpConnection, recipient string) SmtpReturnCode

	// Invoked when a DATA command is received.
	Data(conn *SmtpConnection, contents io.ReadCloser) SmtpReturnCode

	// Invoked when the DATA command finished.
	DataEnd(conn *SmtpConnection) SmtpReturnCode

	// Invoked when an RSET command was received.
	Reset(conn *SmtpConnection) SmtpReturnCode

	// Invoked when a QUIT command was received.
	Quit(conn *SmtpConnection) SmtpReturnCode
}

// An ongoing SMTP connection with all required state.
type SmtpConnection struct {
	active   bool
	cb       SmtpReceiver
	conn     *textproto.Conn
	origconn net.Conn
}

// Create a new SMTP connection by doing the SMTP server-side handshake
// on the socket given as conn. This will spawn a new thread which will
// handle any callbacks to "cb".
func newSmtpConnection(conn net.Conn, cb SmtpReceiver) {
	var txt = textproto.NewConn(conn)
	var ret = SmtpConnection{
		active:   true,
		cb:       cb,
		conn:     txt,
		origconn: conn,
	}
	go ret.handle()
}

// Send an error message back to the client with the given SMTP error
// code and explanation text. This will NOT terminate the connection for
// you; if you want that, you'll have to do it yourself.
func (self *SmtpConnection) RespondWithError(code int, explanation string) {
	self.Respond(code, false, "Error: "+explanation)
}

// Send a message back to the client with the given SMTP response code and
// text.
func (self *SmtpConnection) Respond(code int, continued bool, text string) {
	var sep string
	if continued {
		sep = "-"
	} else {
		sep = " "
	}
	self.conn.PrintfLine("%03d%s%s", code, sep, text)
	smtp_bytes_out.Add(int64(len(text) + 6))
}

// Send the given SMTP return code back to the client. This will NOT
// terminate the connection for you; if you want that, you'll have to do
// it yourself.
func (self *SmtpConnection) RespondWithRCode(code *SmtpReturnCode) {
	self.Respond(code.Code, false, code.Message)
}

// Parse a line as a command and run the appropriate handlers.
// This will block until all appropriate handlers have finished.
func (self *SmtpConnection) handleCommand(command string) (ret SmtpReturnCode) {
	var cmd, params string
	var splitdata []string = strings.SplitN(command, " ", 2)

	if len(splitdata) < 1 {
		smtp_dialog_errors.Add("empty-command", 1)
	}
	cmd = strings.ToUpper(splitdata[0])
	if len(splitdata) > 1 {
		params = splitdata[1]
	}

	switch cmd {
	case "HELO":
	case "EHLO":
		{
			ret.Code = SMTP_NOT_IMPLEMENTED
			ret.Message = "Your command was: HELO, parameter was: " + params
			return
		}
	case "MAIL":
		{
			// TODO: Parse from part.
			ret.Code = SMTP_NOT_IMPLEMENTED
			ret.Message = "Your command was: MAIL, parameter was: " + params
			return
		}
	case "RCPT":
		{
			// TODO: Parse from part.
			ret.Code = SMTP_NOT_IMPLEMENTED
			ret.Message = "Your command was: MAIL, parameter was: " + params
			return
		}
	case "DATA":
		{
			// TODO: Parse from part.
			ret.Code = SMTP_NOT_IMPLEMENTED
			ret.Message = "Your command was: DATA, parameter was: " + params
			return
		}
	case "ETRN":
		{
			// TODO: Parse from part.
			ret.Code = SMTP_NOT_IMPLEMENTED
			ret.Message = "Your command was: ETRN, parameter was: " + params
			return
		}
	case "RSET":
		{
			// TODO: Parse from part.
			ret.Code = SMTP_NOT_IMPLEMENTED
			ret.Message = "Your command was: RSET, parameter was: " + params
			return
		}
	case "QUIT":
		{
			ret.Code = SMTP_CLOSING
			ret.Message = "Your command was: QUIT, parameter was: " + params
			ret.Terminate = true
			return
		}
	default:
		{
			ret.Code = SMTP_NOT_IMPLEMENTED
			ret.Message = "Command " + cmd + " is not supported."
		}
	}

	return ret
}

// Do a server-side SMTP handshake on the wrapped connection and handle
// any incomming commands. This method will block until the connection
// is terminated.
func (self *SmtpConnection) handle() {
	var deadline time.Time = time.Now().Add(1 * time.Second)
	var rc SmtpReturnCode
	var err error

	smtp_active_connections.Add(1)

	// When we get out of here, do some cleanup.
	defer self.conn.Close()
	defer self.setInactive()
	defer self.cb.ConnectionClosed(self)
	defer smtp_active_connections.Add(-1)

	self.origconn.SetReadDeadline(deadline)
	for time.Now().Before(deadline) {
		var cmd string
		cmd, err = self.conn.ReadLine()
		if len(cmd) > 0 {
			self.RespondWithError(SMTP_CLOSING,
				"I can break rules, too. Goodbye.")
			smtp_dialog_errors.Add("unauth-pipelining", 1)
			smtp_bytes_in.Add(int64(len(cmd)))
			return
		} else if err != nil {
			var neterr net.Error
			var ok bool
			neterr, ok = err.(net.Error)
			if !ok || (!neterr.Timeout() && !neterr.Temporary()) {
				self.RespondWithError(SMTP_UNAVAIL, err.Error())
				smtp_dialog_errors.Add(err.Error(), 1)
				return
			}
		}
	}

	self.origconn.SetReadDeadline(time.Unix(0, 0))

	rc = self.cb.ConnectionOpened(self, self.origconn.RemoteAddr())
	if rc.Code != 0 {
		self.Respond(rc.Code, false, rc.Message)
		if rc.Terminate {
			return
		}
	} else {
		self.Respond(SMTP_READY, false, "MailPump "+mailpump.MAILPUMP_VERSION+
			" ready.")
	}

	// By this, the connection is established. Start looking for commands.
	for {
		var cmd string
		deadline = time.Now().Add(time.Minute)
		self.origconn.SetReadDeadline(deadline)
		cmd, err = self.conn.ReadLine()
		self.origconn.SetReadDeadline(time.Unix(0, 0))
		smtp_bytes_in.Add(int64(len(cmd)))
		if err != nil {
			var neterr net.Error
			var ok bool
			neterr, ok = err.(net.Error)
			if ok && neterr.Timeout() {
				self.Respond(SMTP_CLOSING, false,
					"Timeout; closing connection")
				smtp_command_timeouts.Add(1)
				return
			}
			if !ok || !neterr.Temporary() {
				self.RespondWithError(SMTP_UNAVAIL, err.Error())
				return
			} else {
				// Try reading again.
				continue
			}
		}

		rc = self.handleCommand(cmd)
		if rc.Code > 0 {
			self.RespondWithRCode(&rc)
		}
		if rc.Terminate {
			return
		}
	}
}

// Called to end the active period of the connection.
func (self *SmtpConnection) setInactive() {
	self.active = false
}

// Indicate to the caller whether the connection is still active.
func (self *SmtpConnection) IsActive() bool {
	return self.active
}