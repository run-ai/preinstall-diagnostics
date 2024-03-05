package external_cluster_tests

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/run-ai/preinstall-diagnostics/internal/k8sclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func CertificateIsValid(clusterFQDN string) error {
	if clusterFQDN == "" {
		return fmt.Errorf("no cluster domain specified, please provide it using --cluster-domain")
	}

	secretName := "runai-cluster-domain-tls-secret"

	k8s, err := k8sclient.ClientSet()
	if err != nil {
		return err
	}

	secret, err := k8s.CoreV1().Secrets("runai").Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	//keyKey := "tls.key"
	crtKey := "tls.crt"

	//key := secret.Data[keyKey]
	crt := secret.Data[crtKey]

	crts := []*x509.Certificate{}

	block, rest := pem.Decode(crt)

	for block != nil {
		if block.Type == "CERTIFICATE" {
			crt, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return err
			}

			crts = append(crts, crt)
		}

		block, rest = pem.Decode(rest)
	}

	clusterCertFound := false

	for _, crt := range crts {
		if time.Now().After(crt.NotAfter) {
			return fmt.Errorf("cert %s is expired", crt.Subject.String())
		}

		err := crt.VerifyHostname(clusterFQDN)
		if err == nil {
			clusterCertFound = true
		}
	}

	if !clusterCertFound {
		return fmt.Errorf("no certificate found for the DNS record %s", clusterFQDN)
	}

	return nil
}
