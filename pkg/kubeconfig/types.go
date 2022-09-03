package kubeconfig

import "sigs.k8s.io/yaml"

type ObjectMeta struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

type Config struct {
	ObjectMeta
	Preferences    interface{}        `json:"preferences"`
	CurrentContext string             `json:"current-context"`
	Clusters       map[string]Cluster `json:"clusters"`
	Contexts       map[string]Context `json:"contexts"`
	Users          map[string]User    `json:"users"`
}

type serializedConfig struct {
	ObjectMeta
	Preferences    interface{}    `json:"preferences"`
	CurrentContext string         `json:"current-context"`
	Clusters       []clusterEntry `json:"clusters"`
	Contexts       []contextEntry `json:"contexts"`
	Users          []userEntry    `json:"users"`
}

type clusterEntry struct {
	Name    string  `json:"name"`
	Cluster Cluster `json:"cluster"`
}

type contextEntry struct {
	Name    string  `json:"name"`
	Context Context `json:"context"`
}

type userEntry struct {
	Name string `json:"name"`
	User User   `json:"user"`
}

type Cluster struct {
	CertificateAuthority string `json:"certificate-authority-data"`
	Server               string `json:"server"`
}

type Context struct {
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace,omitempty"`
	User      string `json:"user" yaml:"user"`
}

type User struct {
	ClientCertificate string `json:"client-certificate-data"`
	ClientKey         string `json:"client-key-data"`
}

func (c *Config) Marshal() []byte {

	clusters := []clusterEntry{}
	for name, cluster := range c.Clusters {
		clusters = append(clusters, clusterEntry{
			Name:    name,
			Cluster: cluster,
		})
	}

	contexts := []contextEntry{}
	for name, context := range c.Contexts {
		contexts = append(contexts, contextEntry{
			Name:    name,
			Context: context,
		})
	}

	users := []userEntry{}
	for name, user := range c.Users {
		users = append(users, userEntry{
			Name: name,
			User: user,
		})
	}

	s := serializedConfig{
		ObjectMeta: ObjectMeta{
			APIVersion: "v1",
			Kind:       "Config",
		},
		Preferences:    c.Preferences,
		CurrentContext: c.CurrentContext,
		Clusters:       clusters,
		Contexts:       contexts,
		Users:          users,
	}

	o, _ := yaml.Marshal(s)

	return o
}

func Unmarshal(config []byte) (*Config, error) {
	sc := &serializedConfig{}
	err := yaml.Unmarshal(config, sc)
	if err != nil {
		return nil, err
	}

	// Convert lists of clusters/contexts/users back to map[string]

	clusters := make(map[string]Cluster, len(sc.Clusters))
	for _, cluster := range sc.Clusters {
		clusters[cluster.Name] = cluster.Cluster
	}

	contexts := make(map[string]Context, len(sc.Contexts))
	for _, context := range sc.Contexts {
		contexts[context.Name] = context.Context
	}

	users := make(map[string]User, len(sc.Users))
	for _, user := range sc.Users {
		users[user.Name] = user.User
	}

	cfg := &Config{
		ObjectMeta: ObjectMeta{
			APIVersion: "v1",
			Kind:       "Config",
		},
		Preferences:    sc.Preferences,
		CurrentContext: sc.CurrentContext,
		Clusters:       clusters,
		Contexts:       contexts,
		Users:          users,
	}

	return cfg, nil
}

func NewBareConfig() *Config {
	return &Config{
		ObjectMeta: ObjectMeta{
			APIVersion: "v1",
			Kind:       "Config",
		},
	}
}
