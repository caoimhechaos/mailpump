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
	"io/ioutil"
	"log"
	"net"
	"net/mail"
	"reflect"
	"regexp"
	"time"

	"ancient-solutions.com/mailpump"
	"ancient-solutions.com/mailpump/smtpump"
)

type smtpCallback struct {
	smtpump.SmtpReceiver
	maxContentLength int64
}

var features = []string{"ETRN", "8BITMIME", "DSN"}

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
	conn *smtpump.SmtpConnection, hostname string, esmtp bool) (
	ret smtpump.SmtpReturnCode) {
	var msg *mailpump.MailMessage = getConnectionData(conn)
	var response string = fmt.Sprintf("Hello, %s! Nice to meet you.",
		hostname)
	msg.SmtpHelo = &hostname

	if esmtp {
		var pos int
		var capa string
		conn.Respond(smtpump.SMTP_COMPLETED, true, response)

		for pos, capa = range features {
			conn.Respond(smtpump.SMTP_COMPLETED,
				pos < (len(features)-1), capa)
		}
		return
	}

	ret.Code = smtpump.SMTP_COMPLETED
	ret.Message = response
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
		return
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

// Read the data following the DATA command, up to the configured limit.
func (self smtpCallback) Data(conn *smtpump.SmtpConnection) (
	ret smtpump.SmtpReturnCode) {
	var msg *mailpump.MailMessage = getConnectionData(conn)
	var hdr string
	var vals []string
	var addrs []*mail.Address
	var addr *mail.Address
	var dotreader io.Reader
	var contentsreader *io.LimitedReader
	var message *mail.Message
	var tm time.Time
	var err error

	if msg.SmtpHelo == nil {
		ret.Code = smtpump.SMTP_BAD_SEQUENCE
		ret.Message = "Polite people say Hello first!"
		return
	}

	if msg.SmtpFrom == nil {
		ret.Code = smtpump.SMTP_BAD_SEQUENCE
		ret.Message = "Need MAIL command before DATA."
		return
	}

	if len(msg.SmtpTo) == 0 {
		ret.Code = smtpump.SMTP_BAD_SEQUENCE
		ret.Message = "Need RCPT command before DATA."
		return
	}

	conn.Respond(smtpump.SMTP_PROCEED, false, "Proceed with message.")

	dotreader = conn.GetDotReader()
	contentsreader = &io.LimitedReader{
		R: dotreader,
		N: self.maxContentLength + 1,
	}
	message, err = mail.ReadMessage(contentsreader)
	if err != nil {
		ret.Code = smtpump.SMTP_LOCALERR
		ret.Message = "Unable to read message: " + err.Error()
		// Consume all remaining output before returning an error.
		ioutil.ReadAll(dotreader)
		return
	}

	// See if we ran out of bytes to our limit
	if contentsreader.N <= 0 {
		ret.Code = smtpump.SMTP_MESSAGE_TOO_BIG
		ret.Message = "Size limit exceeded. Thanks for playing."
		ret.Terminate = true
		return
	}

	msg.Body, err = ioutil.ReadAll(message.Body)
	if err != nil {
		ret.Code = smtpump.SMTP_LOCALERR
		ret.Message = "Unable to parse message: " + err.Error()
		return
	}

	for hdr, vals = range message.Header {
		var header = new(mailpump.MailMessage_MailHeader)
		header.Name = new(string)
		*header.Name = hdr
		header.Value = make([]string, len(vals))
		copy(header.Value, vals)
		msg.Headers = append(msg.Headers, header)
	}

	tm, err = message.Header.Date()
	if err == nil {
		msg.DateHdr = new(int64)
		*msg.DateHdr = tm.Unix()
	}

	addrs, _ = message.Header.AddressList("From")
	if len(addrs) > 0 {
		msg.FromHdr = new(string)
		*msg.FromHdr = addrs[0].String()
	}

	addrs, _ = message.Header.AddressList("To")
	for _, addr = range addrs {
		msg.ToHdr = append(msg.ToHdr, addr.String())
	}

	addrs, _ = message.Header.AddressList("Cc")
	for _, addr = range addrs {
		msg.CcHdrs = append(msg.CcHdrs, addr.String())
	}

	addrs, _ = message.Header.AddressList("Sender")
	if len(addrs) > 0 {
		msg.SenderHdr = new(string)
		*msg.SenderHdr = addrs[0].String()
	}

	hdr = message.Header.Get("Message-Id")
	if len(hdr) <= 0 {
		hdr = message.Header.Get("Message-ID")
	}
	if len(hdr) > 0 {
		msg.MsgidHdr = new(string)
		*msg.MsgidHdr = hdr
	}

	// TODO(caoimhe): send this to some server.
	ret.Code = smtpump.SMTP_NOT_IMPLEMENTED
	ret.Message = "Ok, but this doesn't go anywhere yet."
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
