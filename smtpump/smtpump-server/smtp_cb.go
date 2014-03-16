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

// SMTP handler callback.
package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"regexp"

	"ancient-solutions.com/mailpump"
	"ancient-solutions.com/mailpump/smtpump"
)

type smtpCallback struct {
	smtpump.SmtpReceiver
}

// String representation of an email regular expression.
var email_re string = "([\\w\\+-\\.]+(?:%[\\w\\+-\\.]+)?@[\\w\\+-\\.]+)"

// RE match to extract the mail address from a MAIL From command.
var from_re *regexp.Regexp = regexp.MustCompile(
	"^[Ff][Rr][Oo][Mm]:\\s*(?:<" + email_re + ">|" + email_re + ")$")

// RE match to extract the mail address from a RCPT To command.
var rcpt_re *regexp.Regexp = regexp.MustCompile(
	"^[Tt][Oo]:\\s*(?:<" + email_re + ">|" + email_re + ")$")

func getConnectionData(conn *smtpump.SmtpConnection) *mailpump.MailMessage {
	var msg *mailpump.MailMessage
	var val reflect.Value
	var ud interface{}
	var ok bool

	ud = conn.GetUserdata()
	val = reflect.ValueOf(ud)
	if !val.IsValid() || val.IsNil() {
		msg = new(mailpump.MailMessage)
		conn.SetUserdata(msg)
		return msg
	}

	msg, ok = ud.(*mailpump.MailMessage)
	if !ok {
		log.Print("Connection userdata is not a MailMessage!")
		return nil
	}

	if msg == nil {
		msg = new(mailpump.MailMessage)
		conn.SetUserdata(msg)
	}

	return msg
}

// Store all available information about the peer in the message structure
// for SPAM analysis.
func (self smtpCallback) ConnectionOpened(
	conn *smtpump.SmtpConnection, peer net.Addr) (
	ret smtpump.SmtpReturnCode) {
	var host string
	var msg *mailpump.MailMessage = getConnectionData(conn)
	var err error

	if msg == nil {
		ret.Code = smtpump.SMTP_LOCALERR
		ret.Message = "Unable to allocate connection structures."
		ret.Terminate = true
		return
	}

	host, _, err = net.SplitHostPort(peer.String())
	if err == nil {
		msg.SmtpPeer = &host
	} else {
		host = peer.String()
		msg.SmtpPeer = &host
	}
	msg.SmtpPeerRevdns, _ = net.LookupAddr(host)
	return
}

// Ignore disconnections.
func (self smtpCallback) ConnectionClosed(conn *smtpump.SmtpConnection) {
}

// Just save the host name and respond.
func (self smtpCallback) Helo(
	conn *smtpump.SmtpConnection, hostname string) (
	ret smtpump.SmtpReturnCode) {
	var msg *mailpump.MailMessage = getConnectionData(conn)
	msg.SmtpHelo = &hostname

	ret.Code = smtpump.SMTP_COMPLETED
	ret.Message = fmt.Sprintf("Hello, %s! Nice to meet you.", hostname)
	return
}

// Ensure HELO has been set, then record From.
func (self smtpCallback) MailFrom(
	conn *smtpump.SmtpConnection, sender string) (
	ret smtpump.SmtpReturnCode) {
	var msg *mailpump.MailMessage = getConnectionData(conn)
	var matches []string
	var addr string

	if msg.SmtpHelo == nil {
		ret.Code = smtpump.SMTP_BAD_SEQUENCE
		ret.Message = "Polite people say Hello first!"
		return
	}

	matches = from_re.FindStringSubmatch(sender)
	if len(matches) == 0 {
		if len(sender) > 0 {
			log.Print("Received unparseable address: ", sender)
		}
		ret.Code = smtpump.SMTP_PARAMETER_NOT_IMPLEMENTED
		ret.Message = "Address not understood, sorry."
		return
	}

	for _, addr = range matches {
		if len(addr) > 0 {
			msg.SmtpFrom = new(string)
			*msg.SmtpFrom = addr
		}
	}
	ret.Code = smtpump.SMTP_COMPLETED
	ret.Message = "Ok."
	return
}

// Ensure HELO and MAIL have been set, then record To.
func (self smtpCallback) RcptTo(
	conn *smtpump.SmtpConnection, recipient string) (
	ret smtpump.SmtpReturnCode) {
	var msg *mailpump.MailMessage = getConnectionData(conn)
	var matches []string
	var addr string
	var realaddr string

	if msg.SmtpHelo == nil {
		ret.Code = smtpump.SMTP_BAD_SEQUENCE
		ret.Message = "Polite people say Hello first!"
		return
	}

	if msg.SmtpFrom == nil {
		ret.Code = smtpump.SMTP_BAD_SEQUENCE
		ret.Message = "Need MAIL command before RCPT."
	}

	matches = rcpt_re.FindStringSubmatch(recipient)
	if len(matches) == 0 {
		if len(recipient) > 0 {
			log.Print("Received unparseable address: ", recipient)
		}
		ret.Code = smtpump.SMTP_PARAMETER_NOT_IMPLEMENTED
		ret.Message = "Address not understood, sorry."
		return
	}

	for _, addr = range matches {
		if len(addr) > 0 {
			realaddr = addr
		}
	}
	msg.SmtpTo = append(msg.SmtpTo, realaddr)
	ret.Code = smtpump.SMTP_COMPLETED
	ret.Message = "Ok."
	return
}

// FIXME: STUB.
func (self smtpCallback) Data(
	conn *smtpump.SmtpConnection, contents io.Reader) (
	ret smtpump.SmtpReturnCode) {
	ret.Code = smtpump.SMTP_NOT_IMPLEMENTED
	ret.Message = "Not yet implemented."
	return
}

// FIXME: STUB.
func (self smtpCallback) DataEnd(conn *smtpump.SmtpConnection) (
	ret smtpump.SmtpReturnCode) {
	ret.Code = smtpump.SMTP_NOT_IMPLEMENTED
	ret.Message = "Not yet implemented."
	return
}

// FIXME: STUB.
func (self smtpCallback) Etrn(conn *smtpump.SmtpConnection, domain string) (
	ret smtpump.SmtpReturnCode) {
	ret.Code = smtpump.SMTP_NOT_IMPLEMENTED
	ret.Message = "Not yet implemented."
	return
}

// FIXME: STUB.
func (self smtpCallback) Reset(conn *smtpump.SmtpConnection) (
	ret smtpump.SmtpReturnCode) {
	ret.Code = smtpump.SMTP_NOT_IMPLEMENTED
	ret.Message = "Not yet implemented."
	return
}

// FIXME: STUB.
func (self smtpCallback) Quit(conn *smtpump.SmtpConnection) (
	ret smtpump.SmtpReturnCode) {
	ret.Code = smtpump.SMTP_CLOSING
	ret.Message = "See you later!"
	ret.Terminate = true
	return
}
