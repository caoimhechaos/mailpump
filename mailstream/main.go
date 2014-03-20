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

	"ancient-solutions.com/doozer/exportedservice"
	"ancient-solutions.com/mailpump"
	"code.google.com/p/goprotobuf/proto"
)

func main() {
	var service *MailSubmissionService
	var conf *mailpump.MailPumpConfiguration
	var l net.Listener
	var config_contents []byte
	var configpath string
	var err error

	flag.StringVar(&configpath, "config-path", "mailstream.cfg",
		"Path to the mailstream configuration file "+
			"(an ascii protocol buffer).")
	flag.Parse()

	config_contents, err = ioutil.ReadFile(configpath)
	if err != nil {
		log.Fatal("Unable to read ", configpath, ": ", err)
	}

	conf = new(mailpump.MailPumpConfiguration)
	err = proto.UnmarshalText(string(config_contents), conf)
	if err != nil {
		log.Fatal("Error parsing ", configpath, ": ", err)
	}

	if conf.GetInsecure() {
		l, err = net.Listen("tcp", conf.GetBindTo())
	} else {
		var tlscert tls.Certificate
		var config *tls.Config = new(tls.Config)
		var certdata []byte

		config.ClientAuth = tls.VerifyClientCertIfGiven
		config.MinVersion = tls.VersionTLS12

		tlscert, err = tls.LoadX509KeyPair(conf.GetX509Cert(),
			conf.GetX509Key())
		if err != nil {
			log.Fatal("Unable to load X.509 key pair: ", err)
		}
		config.Certificates = append(config.Certificates, tlscert)
		config.BuildNameToCertificate()

		config.ClientCAs = x509.NewCertPool()
		certdata, err = ioutil.ReadFile(conf.GetX509CaCert())
		if err != nil {
			log.Fatal("Error reading ", conf.GetX509CaCert(), ": ", err)
		}
		if !config.ClientCAs.AppendCertsFromPEM(certdata) {
			log.Fatal("Unable to load the X.509 certificates from ",
				conf.GetX509CaCert())
		}

		if conf.DoozerUri != nil && len(conf.GetDoozerUri()) > 0 {
			var exporter *exportedservice.ServiceExporter
			exporter, err = exportedservice.NewExporter(
				conf.GetDoozerUri(), conf.GetDoozerBootUri())
			if err != nil {
				log.Fatal("Error contacting lock service: ", err)
			}
			l, err = exporter.NewExportedTLSPort("tcp",
				conf.GetBindTo(), conf.GetServiceName(), config)
			if err != nil {
				log.Fatal("Error creating exported TLS port: ", err)
			}
			defer exporter.UnexportPort()
		} else {
			l, err = tls.Listen("tcp", conf.GetBindTo(), config)
			if err != nil {
				log.Fatal("Error creating TLS listener port: ", err)
			}
		}
	}

	// Create server-side service object and register with the HTTP server.
	service = &MailSubmissionService{
		config: conf,
	}

	if err = rpc.Register(service); err != nil {
		log.Fatal("Unable to register RPC handler: ", err)
	}

	log.Print("Listening on ", l.Addr())
	rpc.Accept(l)
}
