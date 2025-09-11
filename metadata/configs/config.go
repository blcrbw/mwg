package configs

type ServiceConfig struct {
	API              apiConfig              `yaml:"api"`
	ServiceDiscovery serviceDiscoveryConfig `yaml:"serviceDiscovery"`
	DatabaseConfig   DatabaseConfig         `yaml:"database"`
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

type DatabaseConfig struct {
	Mysql MysqlConfig `yaml:"mysql"`
}

type MysqlConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port" default:"3306"`
	User string `yaml:"user"`
	Pass string `yaml:"password"`
	Name string `yaml:"db_name"`
}

type jaegerConfig struct {
	URL string `yaml:"url"`
}

type prometheusConfig struct {
	MetricsPort int `yaml:"metricsPort"`
}
