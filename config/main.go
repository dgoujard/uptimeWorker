package config

type TomlConfig struct {
	Database DatabaseConfig
	Amq AmqConfig
}
type AmqConfig struct {
	Uri string
	QueueName string
	ConcurentRuntime int
}
type DatabaseConfig struct {
	Server string
	Port int
	User string
	Password string
	Database string
}