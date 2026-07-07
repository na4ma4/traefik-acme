package traefik_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/na4ma4/traefik-acme/traefik"
)

var (
	//nolint:gochecknoglobals // acmeDatav1 is a test variable
	acmeDatav1 []byte

	//nolint:gochecknoglobals // acmeDatav2 is a test variable
	acmeDatav2 []byte

	//nolint:gochecknoglobals,golines // acmeDatav3 is a test variable
	acmeDatav3 = []byte(`{"acme":{"Account":{"Email":"na4ma4@noreply.users.github.com","Registration":{"body":{"status":"valid","contact":["mailto:na4ma4@noreply.users.github.com"]},"uri":"https://acme-v02.api.letsencrypt.org/acme/acct/123456789"},"PrivateKey":"c2VjcmV0LXByaXZhdGUta2V5LWZvci0xMjM0NTY3ODkK","KeyType":"4096"},"Certificates":[{"domain":{"main":"example.com","sans":["*.example.com"]},"certificate":"Y2VydGlmaWNhdGUtZm9yLWV4YW1wbGUuY29tCg==","key":"a2V5LWZvci1leGFtcGxlLmNvbQo=","Store":"default"}]}}`)

	//nolint:gochecknoglobals,golines // acmeDatav4 is a test variable
	acmeDatav4 = []byte(`{"acme":{"Account":{"Email":"na4ma4@noreply.users.github.com","Registration":{"body":{"status":"valid","contact":["mailto:na4ma4@noreply.users.github.com"]},"uri":"https://acme-v02.api.letsencrypt.org/acme/acct/123456789"},"PrivateKey":"c2VjcmV0LXByaXZhdGUta2V5LWZvci0xMjM0NTY3ODkK","KeyType":"4096"},"Certificates":[{"domain":{"main":"*.example.com"},"certificate":"Y2VydGlmaWNhdGUtZm9yLWV4YW1wbGUuY29tCg==","key":"a2V5LWZvci1leGFtcGxlLmNvbQo=","Store":"default"}]}}`)

	//nolint:gochecknoglobals,golines // acmeDatav5 is a test variable
	acmeDatav5 = []byte(`{"acme-different":{"Account":{"Email":"na4ma4@noreply.users.github.com","Registration":{"body":{"status":"valid","contact":["mailto:na4ma4@noreply.users.github.com"]},"uri":"https://acme-v02.api.letsencrypt.org/acme/acct/123456789"},"PrivateKey":"c2VjcmV0LXByaXZhdGUta2V5LWZvci0xMjM0NTY3ODkK","KeyType":"4096"},"Certificates":[{"domain":{"main":"example.com","sans":["*.example.com"]},"certificate":"Y2VydGlmaWNhdGUtZm9yLWV4YW1wbGUuY29tCg==","key":"a2V5LWZvci1leGFtcGxlLmNvbQo=","Store":"default"}]}}`)
)

//nolint:gochecknoinits // init is used to generate test data
func init() {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	notBefore := time.Now().Add(-time.Hour)
	notAfter := time.Now().Add(time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		panic(err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
		DNSNames:  []string{"test.example.com", "another-test.example.com"},

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		panic(err)
	}

	certBuf := &bytes.Buffer{}
	keyBuf := &bytes.Buffer{}

	err = pem.Encode(certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		panic(err)
	}
	err = pem.Encode(keyBuf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	if err != nil {
		panic(err)
	}

	acmeTemp := traefik.LocalNamedStore{
		Certificates: []*traefik.Certificate{
			{
				Key:         keyBuf.Bytes(),
				Certificate: certBuf.Bytes(),
				Domain: traefik.Domain{
					Main: "test.example.com",
					SANs: []string{"another-test.example.com"},
				},
			},
		},
	}

	buf := &bytes.Buffer{}
	err = json.NewEncoder(buf).Encode(acmeTemp)
	if err != nil {
		panic(err)
	}

	acmeDatav1 = buf.Bytes()
	if len(acmeDatav1) == 0 {
		panic("acmeDatav1 is empty")
	}

	buf = &bytes.Buffer{}

	acmeTempv2 := traefik.LocalStore{
		"acme": &acmeTemp,
	}
	err = json.NewEncoder(buf).Encode(acmeTempv2)
	if err != nil {
		panic(err)
	}

	acmeDatav2 = buf.Bytes()
	if len(acmeDatav2) == 0 {
		panic("acmeDatav2 is empty")
	}
}
