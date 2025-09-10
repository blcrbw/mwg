package configs

type ServiceConfig struct {
	API              apiConfig              `yaml:"api"`
	ServiceDiscovery serviceDiscoveryConfig `yaml:"serviceDiscovery"`
	Jaeger           JaegerConfig           `yaml:"jaeger"`
}

type apiConfig struct {
	Port int `yaml:"port"`
}

type serviceDiscoveryConfig struct {
	Consul consulConfig `yaml:"consul"`
}
type consulConfig struct {
	Address string `yaml:"address"`
}

type JaegerConfig struct {
	URL string `yaml:"url"`
}
