package zmiddle

type Options struct {
	Cors    *CorsOptions    `yaml:"cors"`
	Limit   *LimitOptions   `yaml:"limit"`
	Logger  *LoggerOptions  `yaml:"logger"`
	Timeout *TimeoutOptions `yaml:"timeout"`
}
