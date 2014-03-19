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

package smtpump

const (
	SMTP_STATUS                    = 211
	SMTP_HELP                      = 214
	SMTP_READY                     = 220
	SMTP_CLOSING                   = 221
	SMTP_COMPLETED                 = 250
	SMTP_NONLOCAL_USER             = 251
	SMTP_PROCEED                   = 354
	SMTP_UNAVAIL                   = 421
	SMTP_MAILBOX_UNAVAIL           = 450
	SMTP_LOCALERR                  = 451
	SMTP_SERVER_FULL               = 452
	SMTP_SYNTAX_ERROR              = 500
	SMTP_PARAMETER_ERROR           = 501
	SMTP_NOT_IMPLEMENTED           = 502
	SMTP_BAD_SEQUENCE              = 503
	SMTP_PARAMETER_NOT_IMPLEMENTED = 504
	SMTP_NONMAIL_DOMAIN            = 521
	SMTP_ACCESS_DENIED             = 530
	SMTP_BAD_AUTH                  = 535
	SMTP_NO_ACTION_MAILBOX_UNAVAIl = 550
	SMTP_PLEASE_FORWARD            = 551
	SMTP_MESSAGE_TOO_BIG           = 552
	SMTP_ILLEGAL_MAILBOX_NAME      = 553
	SMTP_TRANSACTION_FAILED        = 554
)
