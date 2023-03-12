package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"time"
)

// From https://golang.org/src/net/http/server.go
// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
	period time.Duration
}

// Accept TCP connection and set up keepalive
func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(ln.period)
	return tc, nil
}

// genX509KeyPair generates the TLS keypair for the server
// see: https://gist.github.com/shivakar/cd52b5594d4912fbeb46
func genX509KeyPair(cn, c, o, ou string, expiry int) (tls.Certificate, error) {
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(now.Unix()),
		Subject: pkix.Name{
			CommonName:         cn,
			Country:            []string{c},
			Organization:       []string{o},
			OrganizationalUnit: []string{ou},
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(0, 0, expiry),
		SubjectKeyId:          []byte{113, 117, 105, 99, 107, 115, 101, 114, 118, 101},
		BasicConstraintsValid: true,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, template, template,
		priv.Public(), priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	var outCert tls.Certificate
	outCert.Certificate = append(outCert.Certificate, cert)
	outCert.PrivateKey = priv

	return outCert, nil
}

func newTlsListener(addr, cn, c, o, ou string, expiry int) (net.Listener, error) {
	tlsCfg := &tls.Config{}
	tlsCfg.NextProtos = []string{"http/1.1"}
	tlsCfg.Certificates = make([]tls.Certificate, 1)
	cert, err := genX509KeyPair(cn, c, o, ou, expiry)
	if err != nil {
		return nil, err
	}
	tlsCfg.Certificates[0] = cert
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	listener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener), 60 * time.Minute}, tlsCfg)
	return listener, nil
}
