package upgrade_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/vmware/cluster-api-upgrade-tool/pkg/upgrade"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type secrets struct {
	secret *v1.Secret
	err    error
}

func (s *secrets) Get(_ string, _ metav1.GetOptions) (*v1.Secret, error) {
	return s.secret, s.err
}

func TestNewRestConfigFromCASecretRef(t *testing.T) {
	secret := &secrets{
		secret: &v1.Secret{
			Data: map[string][]byte{
				"cert": []byte(caCertificate),
				"key":  []byte(caPrivateKey),
			},
		},
	}
	config, err := upgrade.NewRestConfigFromCASecretRef(secret, "name", "clustername", "https://example.com:6443")
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if config == nil {
		t.Fatal("config should not be nil")
	}
}

func TestNewRestConfigFromCASecretRefError(t *testing.T) {
	testcases := []struct {
		name    string
		secrets *secrets
	}{
		{
			name: "secret fails to get a secret",
			secrets: &secrets{
				err: errors.New("some error"),
			},
		},
		{
			name: "no cert in the secret",
			secrets: &secrets{
				secret: &v1.Secret{},
			},
		},
		{
			name: "valid cert but no key",
			secrets: &secrets{
				secret: &v1.Secret{
					Data: map[string][]byte{
						"cert": []byte(caCertificate),
					},
				},
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := upgrade.NewRestConfigFromCASecretRef(tc.secrets, "name", "clustername", "https://example.com:6443")
			if err == nil {
				t.Fatal("expected an error but didn't get one")
			}
			if config != nil {
				t.Fatal("config should be nil but it's not")
			}

		})
	}
}

func TestNewRestConfigFromClusterField(t *testing.T) {
	cluster := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ProviderSpec: v1alpha1.ProviderSpec{
				Value: &runtime.RawExtension{
					Raw: []byte(fmt.Sprintf(`{"test": {"cert": "%s", "key": "%s"}}`, caCertificate, caPrivateKey)),
				},
			},
		},
	}
	cfg, err := upgrade.NewRestConfigFromCAClusterField(cluster, "spec.providerSpec.value.test", "https://example.com:8888")
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if cfg == nil {
		t.Fatal("*rest.Config should not be nil")
	}
}

func TestNewRestConfigFromClusterFieldErrors(t *testing.T) {
	testcases := []struct {
		name    string
		cluster *v1alpha1.Cluster
		field   string
	}{
		{
			name: "an empty cluster object",
		},
		{
			name:    "a bad field path",
			cluster: &v1alpha1.Cluster{},
			field:   "some field",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := upgrade.NewRestConfigFromCAClusterField(tc.cluster, tc.field, "https://example.com:8888")
			if err == nil {
				t.Fatal("expected an error but did not get one")
			}
			if cfg != nil {
				t.Fatal("*rest.Config should always be nil")
			}

		})
	}
}

func TestNewRestConfigFromKubeconfigSecretRef(t *testing.T) {
	secret := &secrets{
		secret: &v1.Secret{
			Data: map[string][]byte{
				"kubeconfig": []byte(kubeconfigBytes),
			},
		},
	}
	config, err := upgrade.NewRestConfigFromKubeconfigSecretRef(secret, "")
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if config == nil {
		t.Fatal("nil config is not expected")
	}
}

func TestNewRestConfigFromKubeconfigSecretRefErrors(t *testing.T) {
	testcases := []struct {
		name   string
		secret *secrets
	}{
		{
			name: "empty map in the secret/kubeconfig doesn't exist as a key",
			secret: &secrets{
				secret: &v1.Secret{
					Data: map[string][]byte{
						"unknown key": []byte(""),
					},
				},
			},
		},
		{
			name: "k8s secret client returned an error",
			secret: &secrets{
				err: errors.New("an error."),
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := upgrade.NewRestConfigFromKubeconfigSecretRef(tc.secret, "")
			if err == nil {
				t.Fatal("expected an error but did not get one")
			}
			if config != nil {
				t.Fatalf("expected config to be nil but it is not: %v", config)
			}
		})
	}

}

var caPrivateKey = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBb2Z6TGY3YndJQURKMDBpbkdEWHFGUVRUaDRxZEdFRk96UVRxMjNsL0NxUENacnFoCmI4N0pEelFCeUhhZkxtUVdUcW1JTGdtTFBlQlhkTEkwcHQ3N05haFVGVVBlRWFtYkJVYUhndERkK09ibnJEVUsKMi90anVyYnFaMndPZlEwaUJqb3ovVDhzeUw4OTQra2U2RENnRHRNMDgveUYxMWtUN2h3OEhsVWxUMWZGNE84VwpKbTFsbVBXVVhxWkpDRU5Uc3h1NjFKRmcwdkNUWEdsajVNdHNORU9YaGtFeU9xMndNRVpsR0d4R0dDK1lTaGFsCjByUWQwNTM0eGVlVG96cXVGS0k5MkdQZHVDZ1NJemJ5a2l5ejJxelB1QTFFSXczV3hUK1JGL2ZuS25ZeGN0SVgKNDZoS2cwYnNTVUtTdS9CV2xTQk1ienlsRmdWT0paOW1qSG00M1FJREFRQUJBb0lCQUFDMDk3Wmc5LzlMd1pXNApkWEkzbWdQcGQzclo4Z0dQRjJieHBJeThwUDlJNDZwTEdqVkFzOFByT0M3RnhjQVFFOERZSUk0QzRLYXVlZk02CmE5eG1rTFlWTy9penlSNU9uU3lTdURpcjdLd1BaZWRzZTlXbDdUTjlaTng2cmoyQmR5cUx5bnBLY2ttVi9CRW8KalVmNkhsaXlOMEx4VVN3MWFVN2F0TEIxTXFwNzk4dWlKL3dXZFpuamVGVEN1UThScnltOERoMGEzZ3RMeUkxRwpYN1RtNnZHK2tMSzE4QVFQd0RndUpmWExEazc5SjZNWG5RMFlRUzJVVHpnSGREdk1RUjR4dHV6eXRzUFNWYjdTCjl5L1JlMHlyZTFpVHBTMkRlTDZMQVMyZzVlUmFtaVlaYXlCTnBmblJsUVcwOGdGRjkzYlB0UUlwTzYrSDh0dnYKWEZzSkgwa0NnWUVBMVFKMityU3MzcDVCN1YwQW9ZT0ZOS3Nka0dla21JUFZZZUNwMExYYmVSVnd6L01IaXpXSQpibjd6ZHdCeUJJYTJQcE9UMmJoNCtsZXNWeXBPTnVYS2ZrSHlPdHB5Q1kzNVlhVGRWU2M5RkZLVDVHaHhUdHFRClFyOWxPUjJsNkx2Z3V0cXJBbVUzTDRSM2t2aEQ4M1pjUW81TFBpMDBXT2FOczBkVlBLbk5GZmNDZ1lFQXdxNHQKUXB4UjBCRUVBK1FQY0RWTzY2S21sTXkzMW53RURvV3BOcjJZWGZDM2g4YVNnemQ2NzJlWW4yQjNldENjbUNZcwp2U1dYL0QrSkZhc2J6L3VvcDEyVUovNUViMjhFSEtpdHZycmxxKzF3b0JIYTFjLzNGT2VrNHd3bmRVRmk0eXRXClVwd3VuaWhoWkxHeDVxODU2RjJCRTc4aGYycmZGbjgyTmcvbG9zc0NnWUVBcnNSOThFY0xTd0FYNFhPY0QrakUKQXltZWNSdklYV1pWVGlBeDFFOVJpbkJBQmk1RmN6OXgrQTdySFNsZFl6OVFDZG0xeGozbjdLYkFmU2YxMG04SgpqRHY1VGJack9GR25XaWtWZkVkY2d1OFo3cDZPMFA3Y3ZCY2pLeENiVG0vUC9COXJqZVNUdWNYN0FiZjJzS3ZkCkdMSjlJNytkSW8vUGxWZWlwTXBBdlpNQ2dZQmFBcWdnZlNBQ2dHdUgxUUVpVXpOckZTZko4cUVwQk92blB2dE8KdVBoaXJySmNqMzRjTnlHYTRSNGF5a0pUd1hJMWtxanF4eC92Vy96b3pOVXVJMkFHQ2VrL1dIdVJ2aFY3bnEyKwpXckZuL1g4dU16TW4ybUNXQk1naXhmTFVidWZtdXBuTjFqSmpvNjNzSFpCd1pTSDBBbzkwYnRGeEZSdVNUanpsCllCSS9Zd0tCZ0hqRmtzQ0p2R3NWeUhPS21BZXBnVkhmSnlrQzFBNzJ5UFJSWnBQSlVaMFdmMEs0ZjJjQ1BnVGkKWkZ3UjJsMWZpRkdpR21lY0hYMGJCM3VzazdIUENkdXJtRWMrVi94VDRFSnliOUFPTGN6WXplY1FscWtkMWNyTwpCUSszMnViQWtBRzFTUERJdEQvRTJ2ZHRUcU0xTk9QVGpZVy90TEFzMHNhWUJNUUUzaGJKCi0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg=="
var caCertificate = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM4akNDQWRxZ0F3SUJBZ0lJS21pRWJsOW1ISmN3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB4T1RBMk1UY3hNelEyTlRSYUZ3MHlNREEyTVRZeE16UTJOVGRhTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQW9mekxmN2J3SUFESjAwaW4KR0RYcUZRVFRoNHFkR0VGT3pRVHEyM2wvQ3FQQ1pycWhiODdKRHpRQnlIYWZMbVFXVHFtSUxnbUxQZUJYZExJMApwdDc3TmFoVUZVUGVFYW1iQlVhSGd0RGQrT2JuckRVSzIvdGp1cmJxWjJ3T2ZRMGlCam96L1Q4c3lMODk0K2tlCjZEQ2dEdE0wOC95RjExa1Q3aHc4SGxVbFQxZkY0TzhXSm0xbG1QV1VYcVpKQ0VOVHN4dTYxSkZnMHZDVFhHbGoKNU10c05FT1hoa0V5T3Eyd01FWmxHR3hHR0MrWVNoYWwwclFkMDUzNHhlZVRvenF1RktJOTJHUGR1Q2dTSXpieQpraXl6MnF6UHVBMUVJdzNXeFQrUkYvZm5Lbll4Y3RJWDQ2aEtnMGJzU1VLU3UvQldsU0JNYnp5bEZnVk9KWjltCmpIbTQzUUlEQVFBQm95Y3dKVEFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFBNmpGN0wwbUhqU2lqL2lLV3p6S2ljVDN2RW5iODArdllBUApWdVVVZE8yR1pNVWlFdUZUMTcyVDVESmhaTEx5LzRRUlpXRkNyUlZpVDJIaUhSQmF1dUQrS0MrYTN2L2V0M3YzCm5mQk1RYWw3c2ZyR2Z0TFBpUlFVZ2dJSitaeFNqWHZySm9CQndQZTIzUTMyRENDSEVnMEUreGVzM3FWUzNZcmQKRVE2K3BRR25hc2VtQ29TOTdrUnQwa1lia2hXa0w4dUdnQWRZK29vNHN5QVkxUS9Fak01cnAvNGVyakxiaWpaNgptR0NHQjBEZlZ5YnV3bUlXY3JtcnY4b0YvSXR5ZjE2c2tWWEZzOU9WU045aUorM3FXRDM2d0NNQTliK0dhQ09NCnVTVTlLNnFVMHlLUHBkU21Qanc1NHVsNlN3elY0c053SGlrZ3hlRGlhSkJmU3ZwS3Z5OD0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
var kubeconfigBytes = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRFNU1EWXhPREUzTVRjek1Wb1hEVEk1TURZeE5URTNNVGN6TVZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTTdMCkhPSzBIekpjRDBpV1N3cEl2Wk5kdVYxRVI2UlgwWnJuZW1odERwMTgzd3Z0VlMycGpkem0rWlY1VVlwT3ZRbEsKTThIK09PTkdIaFA0cHllTjJabkxSaWVnZkd1cE93QlVhTkRya3RrYVVqUzgwZHdSdzlsWnJSSEhkSXI0TitBMApzSGQ3MU5kejBzSGtmeCtEKzY5N1R0T0FkeFE5Y2F5Vk53QklxbGV4Z2NZRkExS0ttc2xVQ0ZtdFJ2Y24wdnd6Cm95UCtRYm9iRUZqUHVncVRFK1QvMTFaaEZ0RXhJYzd6bmdsVmI5NlFFMVNlc0Z0WFhiOHFEZEJvL3hlNGdURksKYk1xOGJBZzQ1SjBRem9lOVJra0pGdHNaQTJtdE1qNEdmeDBablZDQXpSVUhDVzVWeld5a0plclFYbzN1SWJZSQpwMWZVOTViN3haZCtiZ1FVbDQwQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFMZTU3SFlYdGdNRllpY0NmVXhwMGlZUVh3VlUKMjlLV2lna2IyVGJBWThQNlpwVGR5STZGS1NVN1hQWWZmc00vcUNuTGJTdjg2VWdSV3BqWXpjYkpSak9jSnIwQwovOURXMm5CV0oyeFdZM3U2S3g5QVI0ZmQ4eTdqY1hlQXNjMWxqeTc1dFBMVSs4VS9XMVU3UHBLWXowUlFOczZuCmgyUTJmRlNJOW9DR1pveFZQZDIrMXRDR0JyVHRVUkVRYzNNa0luM2orOEFQVndlczZVTjhHYkM1TWJPay9DT3oKVWpCbXBZWVg0aUVCN3VKaytmdTJtMXpBK0JpNjRqU1BVK0FSd3ZvNWY4VFN6eHlMNEYrbzErT3dBNCtMelRNSApja1M4Y2JZS3VHencxOGRUTi9nWmlBek5rMFpFTEdoVnNvVE8wOWVjN3h6ZnJVMG1ubyt5T1F1Vk1rUT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    server: https://127.0.0.1:38499
  name: my-cluster
contexts:
- context:
    cluster: my-cluster
    user: kubernetes-admin
  name: kubernetes-admin@my-cluster
current-context: kubernetes-admin@my-cluster
kind: Config
preferences: {}
users:
- name: kubernetes-admin
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM4akNDQWRxZ0F3SUJBZ0lJZEMxZmRMelI4aEV3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB4T1RBMk1UZ3hOekUzTXpGYUZ3MHlNREEyTVRjeE56RTNNelJhTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXozMVVtYXd2dVgxN0tyNGMKVndkaTJnYlZEOWRiQmZvcG44U2gvbnJkZnJHWG10VW8rbGN0bVhTZVhyZmJWcEZtNXo3VG9mbXhERmNCMmZDMQpEbVY3RzVnVVhzSG9SYW1wNjJvaVFIZk41amVXOFM4eXVBLzRVZUtRbFdSUWFuZlUwN2ZTMHl1V05PL2FCbUM5Cm5NRkhkdS81TUxkMlJIc2Z0TFFTUkUzL3dvZUFsbUNwRTBLMTBtVkhoUGdIeDhFeTIvYytFT2pQS082N2dZWDQKRHFOT2hndUxWNHNCUkVsOFo2NkJ6b2hrZTEveEwyQVEreHJrYnNQeTZkcXpPZkxrRDI5aGdCaXZ3Nm9zanF0VwpkZ3lSZS9GYmdmcjlIcjBmRC9BRkRNdWNZSXFLVU1FSEtpczRyL3dWRnUyTnUzUk1PSWxUcVRSc1FTYndycnRnCkI0VG43d0lEQVFBQm95Y3dKVEFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFIZndnVElYK2hETWxmQVE3NkJsbjd0b1RwM3ZkdXRRREg4dApkZ2hqV1JIZUVqa0RHWWdsRktYTlhMdjQzY1NlZnhKZE9WU1pqWEVjRDkwSW4rcHpwTW5nNzJmYk9KTnBLZ3N6ClRvYm9DR1JTUEpKeGllWEloOUtld0hFTVFONWMvZEZmeFVSUFdVU0ZWYktJbVZZMklOZmxQQS9rN3RHcXBXZ00KODVoZG1ZcldLa0x6d0tkYUVaT3JtbERmVmUzSDNPcmVyWmR6QlFTMThadlFsUkNzd0xSZXV2ckJmc056SGlMVAo3czdJbTJNQ0FuNGd2QUlzQWpIRjJqNTlsakxxTE9TdU1vMjNLYjRnTVNQdHlLNzdDamd1aUhaMzkzZnpoaTQ4CmNHU0cydWx6ekFHQThKNEJybFZRVCswQkdWNCtOV084RDNLNTRaSHpnUXdUUzMrdDVmOD0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBejMxVW1hd3Z1WDE3S3I0Y1Z3ZGkyZ2JWRDlkYkJmb3BuOFNoL25yZGZyR1htdFVvCitsY3RtWFNlWHJmYlZwRm01ejdUb2ZteERGY0IyZkMxRG1WN0c1Z1VYc0hvUmFtcDYyb2lRSGZONWplVzhTOHkKdUEvNFVlS1FsV1JRYW5mVTA3ZlMweXVXTk8vYUJtQzluTUZIZHUvNU1MZDJSSHNmdExRU1JFMy93b2VBbG1DcApFMEsxMG1WSGhQZ0h4OEV5Mi9jK0VPalBLTzY3Z1lYNERxTk9oZ3VMVjRzQlJFbDhaNjZCem9oa2UxL3hMMkFRCit4cmtic1B5NmRxek9mTGtEMjloZ0Jpdnc2b3NqcXRXZGd5UmUvRmJnZnI5SHIwZkQvQUZETXVjWUlxS1VNRUgKS2lzNHIvd1ZGdTJOdTNSTU9JbFRxVFJzUVNid3JydGdCNFRuN3dJREFRQUJBb0lCQVFERUY0Mkp3anBFVW52Qgp0SFB5SisvYlg5T2lxZ1BEVFY1ak9TRmo4Tmc5OFRiM1JIYjZ5TU0yb1FrL09RRlkrZ2ZIaWcvV3A3VVVsWElSCmQ3U1ZTNEVpWGdMNlhzWFdlSGMzSGxJS25XOEJJUTVOR0M4VjF6bjJvV25GVmszTm9UeUpidFFYY0x6L2dLS0wKbm9mMGlwR3dyVERUUXIvS0RwYXpYKzlYa0dPamdsajhxZDlCVy9IbDVKOXpmdTd4TEpjSGJHRy9Eckt1WFpoZgpCc0JKN0szdzdweGVVYzhOQ1R1azgxVVI0b2R5SnYzZ3dMMlhiNXdRUDVCVzFXMFRMZm5Ub0ljTVFDUmdkK1F0CnlLdFVDRU1uZDM3d2x2WlE1N000NmtuTVYwUW14YXovOTRVQjhIR2w5TUhMaFFJNXUycC9jSXFUdm1vaW9vL1EKSTBxcEtqOGhBb0dCQU5Rd0h2UG1xS2xidml3ck9nOWxPQzIzeGc2YVJzbUQvTG5lenR5QUtqdDA0MzdBV05XMQpib1I1U0FBb2RBTnlOUmtxVTVzTElLcklTcTQxaWRjMW4zVWdwWHBubEpBTHpVMGIzdTBEc2tkdGZaTFVOMGM5CkErSm5jN08wVkxkVGtIS1I2ZXNuVExTTlJESEZhd1RYeUZtOURPRHZQTHdBYUdHaC9EekdKUHh4QW9HQkFQcFUKM0VoY251M1E5U1VjM2NIVFowRS9MU0pXclZMSlpTbllZb0RTY3I0d1JmUzg3b0RqcGIzMlArQ09ITmFodUJGbgpCajR0Njk1UWtPWlhUWFZLaW9JMUtleFZYMzhxRjRFVFlBdmVEdUxPS0R3cE1oT1E3Vlgzcnp0cWllNHNMZTIyClJBdUVpUHdZemg0Vlk0TUJNM0lzejFUdjRyVnRjVURLMjd5RGc5cGZBb0dBU2FhYVY5YU1YSnkzbWVVM05malEKVXk0aTFSYS8wMXE0OGx0Z09qRlNkMmpQUGtQTmtnQno2QStnNmswZ1Y4SGdwR2VJdFp4YkxteHZYNkM5dzdHZApNNjZ0UVp1S2VhVmZFWkRIQkVYd0w5TFZiNDJ4MGt4ZmVNcW40b3lKaTBpNkxzcHZBMnlVdDJjQmNMVXh2SENaCjNtdzhlQ2NTVWI5aWUrRWFDSTVPY0VFQ2dZRUF5MGJSd2lrRUhaWExHNkgvS1gvam53WVFQb3dvSHN2UHpGVWMKV0FPTWpZaGhwa1V0WDVoOFpHOXNtNHFMUEhjQ0k0K0hjRUtXQUdkbjRzUU44Q3Job3E1TkpzNkV4NXlFalpvUQpLbExkdjZzczNQMk8zbmlYWVhISjUvT3hvYWhhZTJmQWhhSFFJdlo1bWRCQWlJY1hJYVhsanRGbFJYSmp2dnQ3CngrNzd5UDBDZ1lFQXdGZ2dLWVdHTUxKczhaa0JyT21wNm1NZHA3RzBpSS8xamxhTFVCQU1oV2RWWHlhSElESGwKSnBETGw1QUpKcmxwT3VVaXlBZSsxNHVkblQxcWFNanE0czVlN28xUFRpMW1SUjA5eWFCTjF5NEpNZHB3eTRMOAptY1VhcTRmeTc5YUZCdlF0ZG9qKy8zOWhVWU5lZHg3bUEvUzVhY29iME9kM3VlWFhpY1dOYlpBPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=`
