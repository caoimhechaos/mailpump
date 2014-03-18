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
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"

	"ancient-solutions.com/doozer/exportedservice"
)

func main() {
	var service *MailSubmissionService
	var l net.Listener
	var insecure bool
	var bind string
	var cert, key string
	var cacert string
	var uri, buri string
	var srvname string
	var spamd string
	var err error

	flag.StringVar(&bind, "bind", "[::]:0", "Port to bind the RPC server "+
		"to. By default, a port is picked at random.")
	flag.StringVar(&cert, "cert", "mailstream.crt",
		"Path to the X.509 certificate of this service.")
	flag.StringVar(&key, "key", "mailstream.key",
		"Path to the X.509 key of this service.")
	flag.StringVar(&cacert, "ca-certificate", "cacert.crt",
		"Path to the CA certificate clients will be checked against.")
	flag.StringVar(&uri, "doozer-uri", os.Getenv("DOOZER_URI"),
		"Doozer URI for lock services.")
	flag.StringVar(&buri, "doozer-boot-uri", os.Getenv("DOOZER_BOOT_URI"),
		"Doozer boot URI for finding the right lock service cluster.")
	flag.StringVar(&srvname, "service-name", "mailstream",
		"Name of the exported port on the lock service.")
	flag.BoolVar(&insecure, "insecure", false,
		"Disable the use of client certificates (for debugging).")

	// Flags for dependencies.
	flag.StringVar(&spamd, "spamd", "localhost",
		"host name or host:port pair of a SpamAssassin instance.")
	flag.Parse()

	if insecure {
		l, err = net.Listen("tcp", bind)
	} else {
		var tlscert tls.Certificate
		var config *tls.Config = new(tls.Config)
		var certdata []byte

		config.ClientAuth = tls.VerifyClientCertIfGiven
		config.MinVersion = tls.VersionTLS12

		tlscert, err = tls.LoadX509KeyPair(cert, key)
		if err != nil {
			log.Fatal("Unable to load X.509 key pair: ", err)
		}
		config.Certificates = append(config.Certificates, tlscert)
		config.BuildNameToCertificate()

		config.ClientCAs = x509.NewCertPool()
		certdata, err = ioutil.ReadFile(cacert)
		if err != nil {
			log.Fatal("Error reading ", cacert, ": ", err)
		}
		if !config.ClientCAs.AppendCertsFromPEM(certdata) {
			log.Fatal("Unable to load the X.509 certificates from ", cacert)
		}

		if len(uri) > 0 {
			var exporter *exportedservice.ServiceExporter
			exporter, err = exportedservice.NewExporter(uri, buri)
			if err != nil {
				log.Fatal("Error contacting lock service: ", err)
			}
			l, err = exporter.NewExportedTLSPort("tcp", bind, srvname, config)
			if err != nil {
				log.Fatal("Error creating exported TLS port: ", err)
			}
			defer exporter.UnexportPort()
		} else {
			l, err = tls.Listen("tcp", bind, config)
			if err != nil {
				log.Fatal("Error creating TLS listener port: ", err)
			}
		}
	}

	// Create server-side service object and register with the HTTP server.
	service = &MailSubmissionService{
		insecure:   insecure,
		spamd_peer: spamd,
	}

	if err = rpc.Register(service); err != nil {
		log.Fatal("Unable to register RPC handler: ", err)
	}

	log.Print("Listening on ", l.Addr())
	rpc.Accept(l)
}
