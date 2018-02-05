package proxy

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
)

type certFaker struct {
	ca  *x509.Certificate
	key crypto.PrivateKey
}

func newCertFaker(caPath, keyPath string) (*certFaker, error) {
	certs, err := tls.LoadX509KeyPair(caPath, keyPath)
	if err != nil {
		return nil, err
	}
	ca, err := x509.ParseCertificate(certs.Certificate[0])
	if err != nil {
		return nil, err
	}
	return &certFaker{
		ca:  ca,
		key: certs.PrivateKey,
	}, nil
}

func (cf *certFaker) FakeCert(original *x509.Certificate) (*tls.Certificate, error) {
	template := cf.createTemplate(original)
	fakeCertData, err := x509.CreateCertificate(nil, template, cf.ca, cf.ca.PublicKey, cf.key)
	return &tls.Certificate{
		Certificate: [][]byte{fakeCertData},
		PrivateKey:  cf.key,
	}, err
}

func (cf *certFaker) createTemplate(cert *x509.Certificate) *x509.Certificate {
	template := &x509.Certificate{}
	*template = *cert
	template.SignatureAlgorithm = cf.ca.SignatureAlgorithm
	return template
}
