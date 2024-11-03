package config

import "github.com/spf13/viper"

var AppConfig *Config

type Config struct {
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     int    `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASS"`
	DBName     string `mapstructure:"DB_NAME"`
	DBSSL      string `mapstructure:"DB_SSL"`

	RMQHost            string `mapstructure:"RABBITMQ_HOST"`
	RMQPort            int    `mapstructure:"RABBITMQ_PORT"`
	RMQUser            string `mapstructure:"RABBITMQ_USER"`
	RMQPass            string `mapstructure:"RABBITMQ_PASS"`
	RMQVHost           string `mapstructure:"RABBITMQ_VHOST"`
	RMQChannelPoolSize int    `mapstructure:"RABBITMQ_CHANNEL_POOL_SIZE"`

	ListenPort      int    `mapstructure:"LISTEN_PORT"`
	ListenHost      string `mapstructure:"LISTEN_HOST"`
	StaticPath      string `mapstructure:"STATIC_PATH"`
	APIServcieToken string `mapstructure:"API_SERVICE_TOKEN"`
	APIMagicPath    string `mapstructure:"API_MAGIC_PATH"`

	// Session timeout in seconds
	SessionTimeout int `mapstructure:"SESSION_TIMEOUT"`

	AceDataAPIKey  string `mapstructure:"ACE_DATA_API_KEY"`
	UserUploadPath string `mapstructure:"USER_UPLOAD_PATH"`

	// AWS
	AWSRegion          string `mapstructure:"AWS_REGION"` // OR CF AccountID
	AWSAccessKeyID     string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `mapstructure:"AWS_SECRET_KEY"`
	AWSBucketName      string `mapstructure:"AWS_BUCKET_NAME"`
	AWSBucketURL       string `mapstructure:"AWS_BUCKET_URL"`
	AWSBucketPath      string `mapstructure:"AWS_BUCKET_PATH"`
	AWSBucketPublicURL string `mapstructure:"AWS_BUCKET_PUBLIC_URL"`

	// Instance Public URL
	InstancePublicURL string `mapstructure:"INSTANCE_PUBLIC_URL"`
	ClusterPublicURL  string `mapstructure:"CLUSTER_PUBLIC_URL"`
}

func NewConfig() *Config {
	config := &Config{}
	viper.AddConfigPath(".")
	viper.AddConfigPath("../")
	viper.AddConfigPath("../../")
	viper.AddConfigPath("/etc/onedotnet/asynctasks/")
	viper.AddConfigPath("$HOME/.onedotnet/asynctasks/")
	viper.SetConfigFile(".env")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := viper.Unmarshal(config); err != nil {
		panic(err)
	}

	return config
}

func init() {
	AppConfig = NewConfig()
}
