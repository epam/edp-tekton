package interceptor

import (
	"context"
	"fmt"
	"time"

	triggersApi "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	certresources "knative.dev/pkg/webhook/certificates/resources"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// SecretCertsName is name of secret where ca-cert, server-cert, server-key will be stored after generation.
	SecretCertsName = "tekton-edp-interceptor-certs"
	// secretServerKey is the name of the key associated with the secret's private key.
	secretServerKey = "server-key.pem"
	// SecretServerCert is the name of the key associated with the secret's public key.
	secretServerCert = "server-cert.pem"
	// SecretCACert is the name of the key associated with the certificate of the CA for the keypair.
	secretCACert = "ca-cert.pem"
	decade       = 100 * 365 * 24 * time.Hour
)

type CertData struct {
	ServerKey  []byte
	ServerCert []byte
	CaCert     []byte
}

func NewCertData(serverKey, serverCert, caCert []byte) *CertData {
	return &CertData{ServerKey: serverKey, ServerCert: serverCert, CaCert: caCert}
}

type SecretService struct {
	client ctrlClient.Client
}

func NewSecretService(client ctrlClient.Client) *SecretService {
	return &SecretService{client: client}
}

// CreateCertsSecret creates and returns a CA certificate and certificate and key for the server.
// serverKey and serverCert are used by the server to establish trust for clients, CA certificate is used by the
// client to verify the server authentication chain. Certificates are based on Interceptor spec.
// After generation all certificates are stored in secret: SecretCertsName.
func (s *SecretService) CreateCertsSecret(
	ctx context.Context,
	namespace,
	interceptorName string,
) (*CertData, error) {
	interceptor := &triggersApi.Interceptor{}

	err := s.client.Get(ctx, ctrlClient.ObjectKey{Namespace: namespace, Name: interceptorName}, interceptor)
	if err != nil {
		return nil, fmt.Errorf("failed to get Interceptor: %w", err)
	}

	serKey, serCert, cacert, err := certresources.CreateCerts(
		ctx,
		interceptor.Spec.ClientConfig.Service.Name,
		interceptor.Spec.ClientConfig.Service.Namespace,
		time.Now().Add(decade),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create certs: %w", err)
	}

	certData := NewCertData(serKey, serCert, cacert)

	secret := &corev1.Secret{}

	err = s.client.Get(ctx, ctrlClient.ObjectKey{Namespace: namespace, Name: SecretCertsName}, secret)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			secret.ObjectMeta = metav1.ObjectMeta{
				Namespace: namespace,
				Name:      SecretCertsName,
			}
			secret.Data = map[string][]byte{
				secretServerKey:  serKey,
				secretServerCert: serCert,
				secretCACert:     cacert,
			}
			secret.Type = corev1.SecretTypeOpaque

			if err = s.client.Create(ctx, secret); err != nil {
				return nil, fmt.Errorf("failed to create secret: %w", err)
			}

			return certData, nil
		}

		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	secret.Data = map[string][]byte{
		secretServerKey:  serKey,
		secretServerCert: serCert,
		secretCACert:     cacert,
	}
	if err = s.client.Update(ctx, secret); err != nil {
		return nil, fmt.Errorf("failed to update secret: %w", err)
	}

	return certData, nil
}

// UpdateCABundle updates Interceptor CaBundle spec with CA certificate.
func (s *SecretService) UpdateCABundle(
	ctx context.Context,
	namespace,
	interceptorName string,
	ca []byte,
) error {
	interceptor := &triggersApi.Interceptor{}

	err := s.client.Get(ctx, ctrlClient.ObjectKey{Namespace: namespace, Name: interceptorName}, interceptor)
	if err != nil {
		return fmt.Errorf("failed to get Interceptor: %w", err)
	}

	interceptor.Spec.ClientConfig.CaBundle = ca

	if err = s.client.Update(ctx, interceptor); err != nil {
		return fmt.Errorf("failed to update Interceptor caBundle: %w", err)
	}

	return nil
}
