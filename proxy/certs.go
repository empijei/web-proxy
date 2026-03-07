package proxy

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	certBase = "cert.pem"
	keyBase  = "key.pem"
)

// LoadCA loads a Certificate Authority from disc, expecting cert.pem and key.pem
// to be in the specified directory.
func LoadCA(certDir string) (caCert, caKey []byte, err error) {
	certPath := filepath.Join(certDir, certBase)
	caCert, err = os.ReadFile(certPath) //nolint: gosec // We are on a local machine.
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", certPath, err)
	}

	keyPath := filepath.Join(certDir, keyBase)
	caKey, err = os.ReadFile(keyPath) //nolint: gosec // We are on a local machine.
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", keyPath, err)
	}

	return caCert, caKey, nil
}

// ParseCA parses the given cert and key.
func ParseCA(caCert, caKey []byte) (*tls.Certificate, error) {
	parsedCert, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return nil, err
	}
	if parsedCert.Leaf, err = x509.ParseCertificate(parsedCert.Certificate[0]); err != nil {
		return nil, err
	}
	return &parsedCert, nil
}

// StoreCA stores the given cert and key on disk as "cert.pem" and "key.pem".
func StoreCA(certDir string, caCert, caKey []byte) error {
	certPath := filepath.Join(certDir, certBase)
	if err := os.WriteFile(certPath, caCert, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", certPath, err)
	}
	keyPath := filepath.Join(certDir, keyBase)
	if err := os.WriteFile(keyPath, caKey, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", keyPath, err)
	}
	return nil
}

// GenerateCA generates the cert.pem and key.pem files to create a new Certificate Authority.
func GenerateCA() (caCert, caKey []byte, _ error) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	{ // Cert (Public)
		pub := &priv.PublicKey
		now := time.Now()
		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128) //nolint: mnd // It's to generate the cert.
		serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)
		template := x509.Certificate{
			SerialNumber:          serialNumber,
			Subject:               pkix.Name{Organization: []string{"WAPTY Suite"}},
			NotBefore:             now,
			NotAfter:              now.Add(365 * 24 * time.Hour),
			IsCA:                  true,
			KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
		}

		derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, pub, priv)
		if err != nil {
			return nil, nil, fmt.Errorf("create certificate: %w", err)
		}
		var certBuf bytes.Buffer
		if err := pem.Encode(&certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
			return nil, nil, fmt.Errorf("write data to cert file: %w", err)
		}
		caCert = certBuf.Bytes()
	}

	{ // Key (Private)
		privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal private key: %w", err)
		}
		var keyBuf bytes.Buffer
		if err := pem.Encode(&keyBuf, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
			return nil, nil, fmt.Errorf("write data to key file: %w", err)
		}
		caKey = keyBuf.Bytes()
	}

	return caCert, caKey, nil
}
