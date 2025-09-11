package configs

type ServiceConfig struct {
	API              apiConfig              `yaml:"api"`
	ServiceDiscovery serviceDiscoveryConfig `yaml:"serviceDiscovery"`
	Jaeger           jaegerConfig           `yaml:"jaeger"`
	Prometheus       prometheusConfig       `yaml:"prometheus"`
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

type jaegerConfig struct {
	URL string `yaml:"url"`
}

type prometheusConfig struct {
	MetricsPort int `yaml:"metricsPort"`
}
