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
	"expvar"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"ancient-solutions.com/mailpump"
	"ancient-solutions.com/mailpump/smtpump"
	"github.com/saintienn/go-spamc"
)

// Counts the SMTP errors which have been returned.
var smtp_return_codes = expvar.NewMap("smtp-return-codes")

// Statistics for spamd.
var spamd_ping_errors = expvar.NewMap("spamd-ping-errors")
var spamd_ping_requests = expvar.NewInt("spamd-ping-requests")
var spamd_ping_timing = expvar.NewFloat("spamd-ping-timing")
var spamd_num_reconnects = expvar.NewInt("spamd-num-reconnects")
var spamd_num_evaluations = expvar.NewInt("spamd-num-evaluations")
var spamd_eval_timing = expvar.NewFloat("spamd-evaluation-timing")
var spamd_eval_errors = expvar.NewMap("spamd-evaluation-errors")
var spamd_result_parsing_errors = expvar.NewInt("spamd-result-parsing-errors")
var num_spam_mails = expvar.NewInt("num-mails-rejected-for-spam")

// Total message processing statistics.
var total_num_messages = expvar.NewInt("num-messages-total")
var total_timing = expvar.NewFloat("message-timing-total")

// Implementation class of the submission service itself.
type MailSubmissionService struct {
	config       *mailpump.MailPumpConfiguration
	spamd_client *spamc.Client
	spamd_mtx    sync.Mutex
}

// Put the code and text inside the submission result and do expvar
// bookkeeping.
func fillSmtpError(result *mailpump.MailSubmissionResult, code int32,
	text string) {
	smtp_return_codes.Add(strconv.Itoa(int(code)), 1)
	result.ErrorCode = new(int32)
	*result.ErrorCode = code
	result.ErrorText = new(string)
	*result.ErrorText = text
}

// Submit a message which hasn't previously been checked for validity.
// This will run SPAM and SPF filter as well as policies before
// attempting to deliver the mail.
func (self *MailSubmissionService) Send(
	msg mailpump.MailMessage, ret *mailpump.MailSubmissionResult) error {
	var start, total_start time.Time
	var res *spamc.SpamDOut
	var rawmessage string
	var spam_verdict *mailpump.QualityVerdict
	var hdr *mailpump.MailMessage_MailHeader
	var spamresult, ok bool
	var err error

	total_start = time.Now()

	if self.spamd_client != nil {
		start = time.Now()
		res, err = self.spamd_client.Ping()
		spamd_ping_timing.Add(time.Now().Sub(start).Seconds())
		spamd_ping_requests.Add(1)
		if err != nil {
			spamd_ping_errors.Add(err.Error(), 1)
			log.Print("Error: ", err)
		}
	}
	if res == nil || res.Code != spamc.EX_OK {
		spamd_num_reconnects.Add(1)
		self.spamd_mtx.Lock()
		// TODO(caoimhe): this will reconnect multiple times in case of
		// lock contention, it's just good enough for testing.
		self.spamd_client = spamc.New(self.config.GetSpamdHost(), 5)
		self.spamd_mtx.Unlock()
	}

	for _, hdr = range msg.Headers {
		rawmessage += fmt.Sprintf("%s: %s\r\n", hdr.GetName(), hdr.GetValue())
	}
	rawmessage += "\r\n" + string(msg.Body)

	// TODO(caoimhe): invoke spamd asynchronously and gather the result via
	// a channel.
	start = time.Now()
	res, err = self.spamd_client.Check(rawmessage)
	spamd_eval_timing.Add(time.Now().Sub(start).Seconds())
	spamd_num_evaluations.Add(1)
	if err == nil && res.Code != spamc.EX_OK {
		err = errors.New(spamc.SpamDError[res.Code])
	} else if err != nil {
		log.Print("Error talking to spamd: ", err)
	}
	if err != nil {
		spamd_eval_errors.Add(err.Error(), 1)
		fillSmtpError(ret, smtpump.SMTP_LOCALERR,
			"Error communicating with backend")
		total_num_messages.Add(1)
		total_timing.Add(time.Now().Sub(total_start).Seconds())
		return nil
	}

	spam_verdict = new(mailpump.QualityVerdict)
	spam_verdict.Source = new(string)
	*spam_verdict.Source = "SpamAssassin"
	spam_verdict.Score = new(float64)
	*spam_verdict.Score, ok = res.Vars["spamScore"].(float64)
	if !ok {
		spamd_result_parsing_errors.Add(1)
		log.Print("Unable to determine SPAM score (", res.Vars, ")")
		fillSmtpError(ret, smtpump.SMTP_LOCALERR,
			"Error communicating with backend")
		total_num_messages.Add(1)
		total_timing.Add(time.Now().Sub(total_start).Seconds())
		return nil
	}

	spamresult, ok = res.Vars["isSpam"].(bool)
	if !ok {
		spamd_result_parsing_errors.Add(1)
		log.Print("Unable to determine SPAM flag (", res.Vars, ")")
		fillSmtpError(ret, smtpump.SMTP_LOCALERR,
			"Error communicating with backend")
		total_num_messages.Add(1)
		total_timing.Add(time.Now().Sub(total_start).Seconds())
		return nil
	}
	spam_verdict.Verdict = new(mailpump.QualityVerdict_VerdictType)
	if spamresult {
		num_spam_mails.Add(1)
		*spam_verdict.Verdict = mailpump.QualityVerdict_SPAM
	} else {
		*spam_verdict.Verdict = mailpump.QualityVerdict_OK
	}

	msg.Verdicts = append(msg.Verdicts, spam_verdict)

	// TODO(caoimhe): Log the message structure somewhere.
	if spamresult {
		fillSmtpError(ret, smtpump.SMTP_TRANSACTION_FAILED,
			"Reject, please keep your SPAM to yourself!")
		total_num_messages.Add(1)
		total_timing.Add(time.Now().Sub(total_start).Seconds())
		return nil
	}
	log.Print("Result: ", msg.String())

	fillSmtpError(ret, smtpump.SMTP_UNAVAIL,
		"Hello from MailSubmissionService!")
	total_num_messages.Add(1)
	total_timing.Add(time.Now().Sub(total_start).Seconds())
	return nil
}
