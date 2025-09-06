package configs

type ServiceConfig struct {
	API              apiConfig              `yaml:"api"`
	ServiceDiscovery serviceDiscoveryConfig `yaml:"serviceDiscovery"`
	MessengerConfig  MessengerConfig        `yaml:"messenger"`
	DatabaseConfig   DatabaseConfig         `yaml:"database"`
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

type MessengerConfig struct {
	Kafka kafkaConfig `yaml:"kafka"`
}

type kafkaConfig struct {
	Address string `yaml:"address" default:"localhost"`
	Port    int    `yaml:"port" default:"9092"`
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
