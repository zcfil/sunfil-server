package config

type Lotus struct {
	Host  string `mapstructure:"host" json:"host" yaml:"host"`
	Token string `mapstructure:"token" json:"token" yaml:"token"`
}
