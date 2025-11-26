package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func ParseCert(data string) (string, error) {
	info := ""
	blocks := []*pem.Block{}
	rest := []byte(data)
	for {
		block, newRest := pem.Decode(rest)
		if block == nil {
			return info, fmt.Errorf("failed to parse certificate PEM")
		}
		blocks = append(blocks, block)
		if len(newRest) == 0 {
			break
		}
		rest = newRest
	}

	for i, block := range blocks {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return info, fmt.Errorf("failed to parse certificate: %w", err)
		}
		info += fmt.Sprintf("Certificate %d\n", i)
		info += fmt.Sprintf("\tIssuer: %s\n", cert.Issuer)
		info += fmt.Sprintf("\tSubject: %s\n", cert.Subject)
		info += fmt.Sprintf("\tNotBefore: %s\n", cert.NotBefore)
		info += fmt.Sprintf("\tNotAfter: %s\n", cert.NotAfter)
		info += fmt.Sprintf("\tSignatureAlgorithm: %s\n", cert.SignatureAlgorithm)
		info += fmt.Sprintf("\tPublicKeyAlgorithm: %s\n", cert.PublicKeyAlgorithm)
		info += fmt.Sprintf("\tSerialNumber: %s\n", cert.SerialNumber)
		info += fmt.Sprintf("\tKeyUsage: %s\n", KeyUsage(cert.KeyUsage))
		if len(cert.OCSPServer) > 0 {
			info += fmt.Sprintf("\tOCSPServer: %v\n", cert.OCSPServer)
		}
		if len(cert.IssuingCertificateURL) > 0 {
			info += fmt.Sprintf("\tIssuingCertificateURL: %v\n", cert.IssuingCertificateURL)
		}
		if len(cert.DNSNames) > 0 {
			info += fmt.Sprintf("\tDNSNames: %v\n", cert.DNSNames)
		}
		if len(cert.EmailAddresses) > 0 {
			info += fmt.Sprintf("\tEmailAddresses: %v\n", cert.EmailAddresses)
		}
		if len(cert.IPAddresses) > 0 {
			info += fmt.Sprintf("\tIPAddresses: %v\n", cert.IPAddresses)
		}
		if len(cert.URIs) > 0 {
			info += fmt.Sprintf("\tURIs: %v\n", cert.URIs)
		}
		if len(cert.CRLDistributionPoints) > 0 {
			info += fmt.Sprintf("\tCRLDistributionPoints: %v\n", cert.CRLDistributionPoints)
		}
	}
	return info, nil
}

func KeyUsage(k x509.KeyUsage) string {
	switch k {
	case x509.KeyUsageDigitalSignature:
		return "DigitalSignature"
	case x509.KeyUsageContentCommitment:
		return "ContentCommitment"
	case x509.KeyUsageKeyEncipherment:
		return "KeyEncipherment"
	case x509.KeyUsageDataEncipherment:
		return "DataEncipherment"
	case x509.KeyUsageKeyAgreement:
		return "KeyAgreement"
	case x509.KeyUsageCertSign:
		return "CertSign"
	case x509.KeyUsageCRLSign:
		return "CRLSign"
	case x509.KeyUsageEncipherOnly:
		return "EncipherOnly"
	case x509.KeyUsageDecipherOnly:
		return "DecipherOnly"
	}
	return ""
}
