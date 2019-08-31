package config

type TomlConfig struct {
	Database DatabaseConfig
	Amq AmqConfig
	Aws AwsConfig
}
type AmqConfig struct {
	Uri string
	QueueName string
	ConcurentRuntime int
}
type AwsConfig struct {
	AccessKey string
	SecretKey string
	Region string
}
type DatabaseConfig struct {
	Server string
	Port int
	User string
	Password string
	Database string
}