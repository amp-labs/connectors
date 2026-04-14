package subscriber

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"
)

// GenerateGraphEncryptionCert returns:
// - base64‑encoded DER X.509 certificate (for EncryptionCertificate)
// - a simple cert ID you can use (for EncryptionCertificateId)
func generateGraphEncryptionCert() (base64Cert string, certId string, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate RSA key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "graph.notification.test",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0), // 1 year validity

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// self‑signed: template signs itself
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate: %w", err)
	}

	return base64.StdEncoding.EncodeToString(derBytes), "test-encryption-cert-1", nil
}
