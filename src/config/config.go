package config

type Config struct {
	Character struct {
		Path string `mapstructure:"path"`
	} `mapstructure:"character"`

	Database struct {
		Type string `mapstructure:"type"`
		Path string `mapstructure:"path"`
	} `mapstructure:"database"`

	LLM struct {
		Provider string `mapstructure:"provider"`
		APIKey   string `mapstructure:"api_key"`
		BaseURL  string `mapstructure:"base_url"`
	} `mapstructure:"llm"`

	Data struct {
		CarvID struct {
			URL    string `mapstructure:"url"`
			APIKey string `mapstructure:"api_key"`
		} `mapstructure:"carvid"`
	} `mapstructure:"data"`
}
