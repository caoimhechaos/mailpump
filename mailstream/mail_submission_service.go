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

package main

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"ancient-solutions.com/mailpump"
	"ancient-solutions.com/mailpump/smtpump"
	"github.com/saintienn/go-spamc"
)

type MailSubmissionService struct {
	insecure     bool
	spamd_peer   string
	spamd_client *spamc.Client
	spamd_mtx    sync.Mutex
}

func fillSmtpError(result *mailpump.MailSubmissionResult, code int32,
	text string) {
	result.ErrorCode = new(int32)
	*result.ErrorCode = code
	result.ErrorText = new(string)
	*result.ErrorText = text
}

func (self *MailSubmissionService) Send(
	msg mailpump.MailMessage, ret *mailpump.MailSubmissionResult) error {
	var res *spamc.SpamDOut
	var rawmessage string
	var spam_verdict *mailpump.QualityVerdict
	var hdr *mailpump.MailMessage_MailHeader
	var spamresult, ok bool
	var err error

	if self.spamd_client != nil {
		res, err = self.spamd_client.Ping()
		if err != nil {
			log.Print("Error: ", err)
		}
	}
	if res == nil || res.Code != spamc.EX_OK {
		self.spamd_mtx.Lock()
		// TODO(caoimhe): this will reconnect multiple times in case of
		// lock contention, it's just good enough for testing.
		self.spamd_client = spamc.New(self.spamd_peer, 5)
		self.spamd_mtx.Unlock()
	}

	for _, hdr = range msg.Headers {
		rawmessage += fmt.Sprintf("%s: %s\r\n", hdr.GetName(), hdr.GetValue())
	}
	rawmessage += "\r\n" + string(msg.Body)

	// TODO(caoimhe): invoke spamd asynchronously and gather the result via
	// a channel.
	res, err = self.spamd_client.Check(rawmessage)
	if err == nil && res.Code != spamc.EX_OK {
		err = errors.New(spamc.SpamDError[res.Code])
	}
	if err != nil {
		fillSmtpError(ret, smtpump.SMTP_LOCALERR,
			"Error communicating with backend")
		return nil
	}

	spam_verdict = new(mailpump.QualityVerdict)
	spam_verdict.Source = new(string)
	*spam_verdict.Source = "SpamAssassin"
	spam_verdict.Score = new(float64)
	*spam_verdict.Score, ok = res.Vars["spamScore"].(float64)
	if !ok {
		fillSmtpError(ret, smtpump.SMTP_LOCALERR,
			"Error communicating with backend")
		return nil
	}

	spamresult, ok = res.Vars["isSpam"].(bool)
	if !ok {
		fillSmtpError(ret, smtpump.SMTP_LOCALERR,
			"Error communicating with backend")
		return nil
	}
	spam_verdict.Verdict = new(mailpump.QualityVerdict_VerdictType)
	if spamresult {
		*spam_verdict.Verdict = mailpump.QualityVerdict_SPAM
	} else {
		*spam_verdict.Verdict = mailpump.QualityVerdict_OK
	}

	msg.Verdicts = append(msg.Verdicts, spam_verdict)

	// TODO(caoimhe): Log the message structure somewhere.
	if spamresult {
		fillSmtpError(ret, smtpump.SMTP_TRANSACTION_FAILED,
			"Reject, please keep your SPAM to yourself!")
		return nil
	}
	log.Print("Result: ", msg.String())

	fillSmtpError(ret, smtpump.SMTP_UNAVAIL,
		"Hello from MailSubmissionService!")
	return nil
}
