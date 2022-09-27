package interceptor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	triggersApi "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSecretService_CreateCertsSecret(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(triggersApi.AddToScheme(scheme))

	type args struct {
		namespace              string
		clusterInterceptorName string
	}

	tests := []struct {
		name                    string
		objects                 []runtime.Object
		args                    args
		wantCertData            assert.ValueAssertionFunc
		wantErr                 assert.ErrorAssertionFunc
		wantGetCreatedSecretErr assert.ErrorAssertionFunc
	}{
		{
			name: "success, secret created",
			objects: []runtime.Object{
				&triggersApi.ClusterInterceptor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "edp",
					},
					Spec: triggersApi.ClusterInterceptorSpec{
						ClientConfig: triggersApi.ClientConfig{
							Service: &triggersApi.ServiceReference{
								Name:      "edp-service",
								Namespace: "edp-ns",
							},
						},
					},
				},
			},
			args: args{
				namespace:              "default",
				clusterInterceptorName: "edp",
			},
			wantCertData:            assert.NotNil,
			wantErr:                 assert.NoError,
			wantGetCreatedSecretErr: assert.NoError,
		},
		{
			name: "success, secret updated",
			objects: []runtime.Object{
				&triggersApi.ClusterInterceptor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "edp",
					},
					Spec: triggersApi.ClusterInterceptorSpec{
						ClientConfig: triggersApi.ClientConfig{
							Service: &triggersApi.ServiceReference{
								Name:      "edp-service",
								Namespace: "edp-ns",
							},
						},
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      SecretCertsName,
					},
				},
			},
			args: args{
				namespace:              "default",
				clusterInterceptorName: "edp",
			},
			wantCertData:            assert.NotNil,
			wantErr:                 assert.NoError,
			wantGetCreatedSecretErr: assert.NoError,
		},
		{
			name: "failed, ClusterInterceptor not found",
			args: args{
				namespace:              "default",
				clusterInterceptorName: "edp",
			},
			wantCertData:            assert.Nil,
			wantErr:                 assert.Error,
			wantGetCreatedSecretErr: assert.Error,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(tt.objects...).Build()

			s := NewSecretService(fakeClient)

			gotCertData, err := s.CreateCertsSecret(context.Background(), tt.args.namespace, tt.args.clusterInterceptorName)

			tt.wantErr(t, err)
			tt.wantCertData(t, gotCertData)

			secret := &corev1.Secret{}
			err = fakeClient.Get(context.Background(), client.ObjectKey{Namespace: tt.args.namespace, Name: SecretCertsName}, secret)
			tt.wantGetCreatedSecretErr(t, err)
		})
	}
}

func TestSecretService_UpdateCABundle(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	utilruntime.Must(triggersApi.AddToScheme(scheme))

	type args struct {
		namespace              string
		clusterInterceptorName string
		ca                     []byte
	}

	tests := []struct {
		name                         string
		objects                      []runtime.Object
		args                         args
		wantErr                      assert.ErrorAssertionFunc
		wantGetClusterInterceptorErr assert.ErrorAssertionFunc
		wantCaBundle                 []byte
	}{
		{
			name: "success",
			objects: []runtime.Object{
				&triggersApi.ClusterInterceptor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "edp",
					},
				},
			},
			args: args{
				namespace:              "default",
				clusterInterceptorName: "edp",
				ca:                     []byte("test-ca"),
			},
			wantErr:                      assert.NoError,
			wantGetClusterInterceptorErr: assert.NoError,
			wantCaBundle:                 []byte("test-ca"),
		},
		{
			name: "failed, ClusterInterceptor not found",
			args: args{
				namespace:              "default",
				clusterInterceptorName: "edp",
				ca:                     []byte("test-ca"),
			},
			wantErr:                      assert.Error,
			wantGetClusterInterceptorErr: assert.Error,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(tt.objects...).Build()

			s := NewSecretService(fakeClient)

			err := s.UpdateCABundle(context.Background(), tt.args.namespace, tt.args.clusterInterceptorName, tt.args.ca)

			tt.wantErr(t, err)

			clusterInterceptor := &triggersApi.ClusterInterceptor{}
			err = s.client.Get(context.Background(), client.ObjectKey{Namespace: tt.args.namespace, Name: tt.args.clusterInterceptorName}, clusterInterceptor)
			tt.wantGetClusterInterceptorErr(t, err)
			assert.Equal(t, tt.wantCaBundle, clusterInterceptor.Spec.ClientConfig.CaBundle)
		})
	}
}
