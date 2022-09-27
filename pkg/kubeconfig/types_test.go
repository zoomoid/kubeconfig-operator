package kubeconfig

import (
	"bytes"
	"reflect"
	"testing"
)

func TestConfigMarshalUnmarshal(t *testing.T) {
	testCases := []struct {
		desc     string
		config   *Config
		expected []byte
	}{
		{
			desc:   "empty",
			config: &Config{},
			expected: []byte(`apiVersion: v1
clusters: []
contexts: []
current-context: ""
kind: Config
preferences: null
users: []
`),
		},
		{
			desc: "non-empty",
			config: &Config{
				CurrentContext: "default",
				Clusters: map[string]Cluster{
					"test": {
						CertificateAuthority: "ca1",
						Server:               "server1",
					},
				},
				Contexts: map[string]Context{
					"test": {
						Cluster:   "default",
						Namespace: "default",
						User:      "user1",
					},
				},
				Users: map[string]User{
					"test": {
						ClientCertificate: "cc1",
						ClientKey:         "key1",
					},
				},
			},
			expected: []byte(`apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: ca1
    server: server1
  name: test
contexts:
- context:
    cluster: default
    namespace: default
    user: user1
  name: test
current-context: default
kind: Config
preferences: null
users:
- name: test
  user:
    client-certificate-data: cc1
    client-key-data: key1
`),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// test marshal
			got := tC.config.Marshal()
			if !bytes.Equal(got, tC.expected) {
				t.Errorf("got %s, expected: %s", got, tC.expected)
			}

			// test unmarshal
			gotConfig, err := Unmarshal(tC.expected)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if reflect.DeepEqual(gotConfig, tC.config) {
				t.Errorf("got %v, expected: %v", gotConfig, tC.config)
			}
		})
	}
}
