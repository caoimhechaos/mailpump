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
	"io"
	"net"

	"ancient-solutions.com/mailpump/smtpump"
)

type smtpCallback struct {
	smtpump.SmtpReceiver
}

// FIXME: STUB.
func (self smtpCallback) ConnectionOpened(
	conn *smtpump.SmtpConnection, peer net.Addr) (
	ret smtpump.SmtpReturnCode) {
	return
}

// FIXME: STUB.
func (self smtpCallback) ConnectionClosed(conn *smtpump.SmtpConnection) {
}

// FIXME: STUB.
func (self smtpCallback) Helo(
	conn *smtpump.SmtpConnection, hostname string) (
	ret smtpump.SmtpReturnCode) {
	return
}

// FIXME: STUB.
func (self smtpCallback) MailFrom(
	conn *smtpump.SmtpConnection, sender string) (
	ret smtpump.SmtpReturnCode) {
	return
}

// FIXME: STUB.
func (self smtpCallback) RcptTo(
	conn *smtpump.SmtpConnection, recipient string) (
	ret smtpump.SmtpReturnCode) {
	return
}

// FIXME: STUB.
func (self smtpCallback) Data(
	conn *smtpump.SmtpConnection, contents io.ReadCloser) (
	ret smtpump.SmtpReturnCode) {
	return
}

// FIXME: STUB.
func (self smtpCallback) DataEnd(conn *smtpump.SmtpConnection) (
	ret smtpump.SmtpReturnCode) {
	return
}

// FIXME: STUB.
func (self smtpCallback) Reset(conn *smtpump.SmtpConnection) (
	ret smtpump.SmtpReturnCode) {
	return
}

// FIXME: STUB.
func (self smtpCallback) Quit(conn *smtpump.SmtpConnection) (
	ret smtpump.SmtpReturnCode) {
	return
}
