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

// Main caller of the SMTP server.
package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"ancient-solutions.com/mailpump/smtpump"
	"github.com/caoimhechaos/go-urlconnection"
)

func main() {
	var mailstream_conn net.Conn
	var netname, laddr, webaddr string
	var maxlen int64
	var insecure_backends bool
	var callback *smtpCallback
	var uri, buri string
	var mailstream_uri string
	var cert, key, cacert string
	var err error

	flag.StringVar(&netname, "network-type", "tcp",
		"Type of network connection (tcp, tcp4, tcp6, etc).")
	flag.StringVar(&laddr, "bind", "[::]:2525",
		"IP address and port to bind to (e.g. [::]:25).")
	flag.StringVar(&webaddr, "web-port", "[::]:8025",
		"IP address and port to bind the web server to (e.g. [::]:8025).")
	flag.StringVar(&uri, "doozer-uri", os.Getenv("DOOZER_URI"),
		"Doozer URI for lock services.")
	flag.StringVar(&buri, "doozer-boot-uri", os.Getenv("DOOZER_BOOT_URI"),
		"Doozer boot URI for finding the right lock service cluster.")
	flag.Int64Var(&maxlen, "max-length-mb", 20,
		"Maximum length (in megabytes) acceptable for mails to be accepted.")
	flag.StringVar(&cert, "cert", "mailstream.crt",
		"Path to the X.509 certificate of this service.")
	flag.StringVar(&key, "key", "mailstream.key",
		"Path to the X.509 key of this service.")
	flag.StringVar(&cacert, "ca-certificate", "cacert.crt",
		"Path to the CA certificate clients will be checked against.")

	// Backend connections.
	flag.StringVar(&mailstream_uri, "mailstream-uri", "",
		"URI to connect to mailstream (e.g. tcp://localhost:1234).")
	flag.BoolVar(&insecure_backends, "insecure-backends", false,
		"Use insecure connections to backends. Do NOT use this for "+
			"production! Mails with user data will be transmitted unencrypted!")
	flag.Parse()

	if maxlen < 1 {
		log.Fatal("Maximum length of a mail must be 1MB or greater.")
	}

	if len(uri) > 0 {
		err = urlconnection.SetupDoozer(buri, uri)
		if err != nil {
			log.Print("Unable to connect to Doozer: ", err, " Disabling.")
		}
	}

	// Establish a connection to the mailstream backend.
	mailstream_conn, err = urlconnection.ConnectTimeout(
		mailstream_uri, time.Second)
	if err != nil {
		log.Fatal("Unable to connect to mailstream on ", mailstream_uri,
			": ", err)
	}

	if !insecure_backends {
		var tlscert tls.Certificate
		var config *tls.Config = new(tls.Config)
		var certdata []byte

		config.MinVersion = tls.VersionTLS12

		tlscert, err = tls.LoadX509KeyPair(cert, key)
		if err != nil {
			log.Fatal("Unable to load X.509 key pair: ", err)
		}
		config.Certificates = append(config.Certificates, tlscert)
		config.BuildNameToCertificate()

		config.RootCAs = x509.NewCertPool()
		certdata, err = ioutil.ReadFile(cert)
		if err != nil {
			log.Fatal("Error reading ", cacert, ": ", err)
		}
		if !config.ClientCAs.AppendCertsFromPEM(certdata) {
			log.Fatal("Unable to load the X.509 certificates from ", cacert)
		}

		mailstream_conn = tls.Client(mailstream_conn, config)
	}

	callback = &smtpCallback{
		maxContentLength: maxlen * 1048576,
		mailstreamConn:   mailstream_conn,
	}
	_, err = smtpump.NewSMTPServer(netname, laddr, *callback)
	if err != nil {
		log.Fatal(err)
	}

	http.ListenAndServe(webaddr, nil)
}
