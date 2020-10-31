package config

import (
	"github.com/BurntSushi/toml"
	"github.com/joeshaw/envdecode"
	"github.com/kardianos/osext"
	"log"
	"os"
	"strings"
	"time"
)

type Conf struct {
	Debug bool `env:"DEBUG"`
	ApiServer ApiserverConf
	Database DatabaseConfig
	Postgres PostegresConfig
	Amq AmqConfig
	Aws AwsConfig
	Worker WorkerConfig
	Alert AlertConfig
	Realtime RealtimeConfig
}
type ApiserverConf struct {
	Port         int           `env:"SERVER_PORT,required"`
	TimeoutRead  duration `env:"SERVER_TIMEOUT_READ,required"`
	TimeoutWrite duration `env:"SERVER_TIMEOUT_WRITE,required"`
	TimeoutIdle  duration `env:"SERVER_TIMEOUT_IDLE,required"`
}
type AlertConfig struct {
	EmailFrom string `env:"ALERT_EMAIL_FROM,required"`
	Realtimechannel string `env:"ALERT_RT_CHANNEL,required"`
}
type RealtimeConfig struct {
	Type string `env:"RT_TYPE,required"`
	Apiurl string `env:"RT_API_URL"`
	Apikey string `env:"RT_API_KEY"`
}

type AmqConfig struct {
	Uri string  `env:"AMQ_URI,required"`
	QueueName string  `env:"AMQ_CHECK_QUEUE,required"`
	QueueAlertName string  `env:"AMQ_ALERT_QUEUE,required"`
	ConcurentRuntime int  `env:"AMQ_WATCHERS_NUMBER,required"`
}
type AwsConfig struct {
	AccessKey string  `env:"AWS_ACCESS,required"`
	SecretKey string  `env:"AWS_SECRET,required"`
	Region string  `env:"AWS_REGION,required"`
}
type WorkerConfig struct {
	EnableLambdaRemoteCheck bool  `env:"LAMBDA_CHECK_ENABLED,required"`
	LambdaArn string  `env:"LAMBDA_ARN"`
	EnableInfluxDb2Reporting bool  `env:"INFLUXDB2_REPORT_ENABLED,required"`
	InfluxDb2Token string  `env:"INFLUXDB2_TOKEN"`
	InfluxDb2Bucket string  `env:"INFLUXDB2_BUCKET"`
	InfluxDb2Org string  `env:"INFLUXDB2_ORIG"`
	InfluxDb2Url string  `env:"INFLUXDB2_URL"`
}
type DatabaseConfig struct {
	Server string  `env:"MONGODB_SERVEUR,required"`
	Port int  `env:"MONGODB_PORT,required"`
	User string  `env:"MONGODB_USER,required"`
	Password string  `env:"MONGODB_PASSWORD,required"`
	Database string  `env:"MONGODB_DATABASE,required"`
}
type PostegresConfig struct {
	Server string  `env:"POSTGRES_SERVEUR,required"`
	Port int  `env:"POSTGRES_PORT,required"`
	User string  `env:"POSTGRES_USER,required"`
	Password string  `env:"POSTGRES_PASSWORD,required"`
	Database string  `env:"POSTGRES_DATABASE,required"`
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

func AppConfig() *Conf {
	var c Conf

	folderPath, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatal(err)
	}
	//Fix for local dev with GoLand
	if strings.HasPrefix(folderPath, "/private/") || strings.HasPrefix(folderPath, "/var/folders/") {
		folderPath = "/Users/damien/uptimeWorker"
	}

	if fileExists(folderPath+"/config.toml") {
		if _, err := toml.DecodeFile(folderPath+"/config.toml", &c); err != nil {
			log.Println(err)
		}
		return &c
	}

	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}