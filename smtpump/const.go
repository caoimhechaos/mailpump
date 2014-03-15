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

const SMTP_STATUS = 211
const SMTP_HELP = 214
const SMTP_READY = 220
const SMTP_CLOSING = 221
const SMTP_COMPLETED = 250
const SMTP_NONLOCAL_USER = 251
const SMTP_PROCEED = 354
const SMTP_UNAVAIL = 421
const SMTP_MAILBOX_UNAVAIL = 450
const SMTP_LOCALERR = 451
const SMTP_SERVER_FULL = 452
const SMTP_SYNTAX_ERROR = 500
const SMTP_PARAMETER_ERROR = 501
const SMTP_NOT_IMPLEMENTED = 502
const SMTP_BAD_SEQUENCE = 503
const SMTP_PARAMETER_NOT_IMPLEMENTED = 504
const SMTP_NONMAIL_DOMAIN = 521
const SMTP_ACCESS_DENIED = 530
const SMTP_BAD_AUTH = 535
const SMTP_NO_ACTION_MAILBOX_UNAVAIl = 550
const SMTP_PLEASE_FORWARD = 551
const SMTP_MESSAGE_TOO_BIG = 552
const SMTP_ILLEGAL_MAILBOX_NAME = 553
const SMTP_TRANSACTION_FAILED = 554
