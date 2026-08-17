package config

type CloudinaryConfig struct {
	CloudName    string `mapstructure:"cloud_name"`
	APIKey       string `mapstructure:"api_key"`
	APISecret    string `mapstructure:"api_secret"`
	UploadFolder string `mapstructure:"upload_folder"`
}

type UploadsConfig struct {
	AllowedTypes  []string `mapstructure:"allowed_types"`
	MaxFileSize   int64    `mapstructure:"max_file_size"`
	Storage       string   `mapstructure:"storage"`
}
