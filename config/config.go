package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"xorm.io/xorm"
)

// ============================================================
// Command-line flags and exports for cmd package
// ============================================================

// AppName is the application name used in CLI.
var AppName = "slotopol"

// Command-line flags (set by cobra)
var (
	CfgFile string   // path to config file
	SqlPath string   // path to sqlite databases
	ObjPath []string // additional yaml paths
	Verbose bool     // verbose logging
)

// ============================================================
// Config Structures – Compatible with Original Code
// ============================================================

// CfgJwtAuth is the authentication configuration.
type CfgJwtAuth struct {
	AccessTTL    time.Duration `json:"access-ttl" yaml:"access-ttl" mapstructure:"access-ttl"`
	RefreshTTL   time.Duration `json:"refresh-ttl" yaml:"refresh-ttl" mapstructure:"refresh-ttl"`
	AccessKey    string        `json:"access-key" yaml:"access-key" mapstructure:"access-key"`
	RefreshKey   string        `json:"refresh-key" yaml:"refresh-key" mapstructure:"refresh-key"`
	NonceTimeout time.Duration `json:"nonce-timeout" yaml:"nonce-timeout" mapstructure:"nonce-timeout"`
}

// CfgSendCode is the activation/email configuration.
type CfgSendCode struct {
	UseActivation      bool          `json:"use-activation" yaml:"use-activation" mapstructure:"use-activation"`
	BrevoApiKey        string        `json:"brevo-api-key" yaml:"brevo-api-key" mapstructure:"brevo-api-key"`
	BrevoEmailEndpoint string        `json:"brevo-email-endpoint" yaml:"brevo-email-endpoint" mapstructure:"brevo-email-endpoint"`
	SenderName         string        `json:"sender-name" yaml:"sender-name" mapstructure:"sender-name"`
	SenderEmail        string        `json:"sender-email" yaml:"sender-email" mapstructure:"sender-email"`
	ReplytoEmail       string        `json:"replyto-email" yaml:"replyto-email" mapstructure:"replyto-email"`
	EmailSubject       string        `json:"email-subject" yaml:"email-subject" mapstructure:"email-subject"`
	EmailHtmlContent   string        `json:"email-html-content" yaml:"email-html-content" mapstructure:"email-html-content"`
	CodeTimeout        time.Duration `json:"code-timeout" yaml:"code-timeout" mapstructure:"code-timeout"`
}

// CfgWebServ is the web server configuration.
type CfgWebServ struct {
	TrustedProxies    []string      `json:"trusted-proxies" yaml:"trusted-proxies" mapstructure:"trusted-proxies"`
	PortHTTP          []string      `json:"port-http" yaml:"port-http" mapstructure:"port-http"`
	ReadTimeout       time.Duration `json:"read-timeout" yaml:"read-timeout" mapstructure:"read-timeout"`
	ReadHeaderTimeout time.Duration `json:"read-header-timeout" yaml:"read-header-timeout" mapstructure:"read-header-timeout"`
	WriteTimeout      time.Duration `json:"write-timeout" yaml:"write-timeout" mapstructure:"write-timeout"`
	IdleTimeout       time.Duration `json:"idle-timeout" yaml:"idle-timeout" mapstructure:"idle-timeout"`
	MaxHeaderBytes    int           `json:"max-header-bytes" yaml:"max-header-bytes" mapstructure:"max-header-bytes"`
	ShutdownTimeout   time.Duration `json:"shutdown-timeout" yaml:"shutdown-timeout" mapstructure:"shutdown-timeout"`
}

// CfgXormDrv is the database configuration.
type CfgXormDrv struct {
	DriverName       string        `json:"driver-name" yaml:"driver-name" mapstructure:"driver-name"`
	UseSpinLog       bool          `json:"use-spin-log" yaml:"use-spin-log" mapstructure:"use-spin-log"`
	ClubSourceName   string        `json:"club-source-name" yaml:"club-source-name" mapstructure:"club-source-name"`
	SpinSourceName   string        `json:"spin-source-name" yaml:"spin-source-name" mapstructure:"spin-source-name"`
	SqlFlushTick     time.Duration `json:"sql-flush-tick" yaml:"sql-flush-tick" mapstructure:"sql-flush-tick"`
	ClubUpdateBuffer int           `json:"club-update-buffer" yaml:"club-update-buffer" mapstructure:"club-update-buffer"`
	ClubInsertBuffer int           `json:"club-insert-buffer" yaml:"club-insert-buffer" mapstructure:"club-insert-buffer"`
	SpinInsertBuffer int           `json:"spin-insert-buffer" yaml:"spin-insert-buffer" mapstructure:"spin-insert-buffer"`
}

// CfgGameplay is the gameplay configuration.
type CfgGameplay struct {
	AdjunctLimit    float64 `json:"adjunct-limit" yaml:"adjunct-limit" mapstructure:"adjunct-limit"`
	MinJackpot      float64 `json:"min-jackpot" yaml:"min-jackpot" mapstructure:"min-jackpot"`
	MaxSpinAttempts int     `json:"max-spin-attempts" yaml:"max-spin-attempts" mapstructure:"max-spin-attempts"`
}

// CfgCloudinary is the Cloudinary configuration.
type CfgCloudinary struct {
	CloudName    string `json:"cloud_name" yaml:"cloud_name" mapstructure:"cloud_name"`
	APIKey       string `json:"api_key" yaml:"api_key" mapstructure:"api_key"`
	APISecret    string `json:"api_secret" yaml:"api_secret" mapstructure:"api_secret"`
	UploadFolder string `json:"upload_folder" yaml:"upload_folder" mapstructure:"upload_folder"`
}

// CfgUploads is the uploads configuration.
type CfgUploads struct {
	AllowedTypes []string `json:"allowed_types" yaml:"allowed_types" mapstructure:"allowed_types"`
	MaxFileSize  int64    `json:"max_file_size" yaml:"max_file_size" mapstructure:"max_file_size"`
	Storage      string   `json:"storage" yaml:"storage" mapstructure:"storage"`
}

// Config – FLAT STRUCTURE with embedded types
// This is compatible with the original code that expects
// fields like Cfg.AccessKey, Cfg.SenderName, etc.
type Config struct {
	// Authentication fields (from CfgJwtAuth)
	AccessTTL    time.Duration `json:"access-ttl" yaml:"access-ttl" mapstructure:"access-ttl"`
	RefreshTTL   time.Duration `json:"refresh-ttl" yaml:"refresh-ttl" mapstructure:"refresh-ttl"`
	AccessKey    string        `json:"access-key" yaml:"access-key" mapstructure:"access-key"`
	RefreshKey   string        `json:"refresh-key" yaml:"refresh-key" mapstructure:"refresh-key"`
	NonceTimeout time.Duration `json:"nonce-timeout" yaml:"nonce-timeout" mapstructure:"nonce-timeout"`

	// Activation fields (from CfgSendCode)
	UseActivation      bool          `json:"use-activation" yaml:"use-activation" mapstructure:"use-activation"`
	BrevoApiKey        string        `json:"brevo-api-key" yaml:"brevo-api-key" mapstructure:"brevo-api-key"`
	BrevoEmailEndpoint string        `json:"brevo-email-endpoint" yaml:"brevo-email-endpoint" mapstructure:"brevo-email-endpoint"`
	SenderName         string        `json:"sender-name" yaml:"sender-name" mapstructure:"sender-name"`
	SenderEmail        string        `json:"sender-email" yaml:"sender-email" mapstructure:"sender-email"`
	ReplytoEmail       string        `json:"replyto-email" yaml:"replyto-email" mapstructure:"replyto-email"`
	EmailSubject       string        `json:"email-subject" yaml:"email-subject" mapstructure:"email-subject"`
	EmailHtmlContent   string        `json:"email-html-content" yaml:"email-html-content" mapstructure:"email-html-content"`
	CodeTimeout        time.Duration `json:"code-timeout" yaml:"code-timeout" mapstructure:"code-timeout"`

	// Web Server fields (from CfgWebServ)
	TrustedProxies    []string      `json:"trusted-proxies" yaml:"trusted-proxies" mapstructure:"trusted-proxies"`
	PortHTTP          []string      `json:"port-http" yaml:"port-http" mapstructure:"port-http"`
	ReadTimeout       time.Duration `json:"read-timeout" yaml:"read-timeout" mapstructure:"read-timeout"`
	ReadHeaderTimeout time.Duration `json:"read-header-timeout" yaml:"read-header-timeout" mapstructure:"read-header-timeout"`
	WriteTimeout      time.Duration `json:"write-timeout" yaml:"write-timeout" mapstructure:"write-timeout"`
	IdleTimeout       time.Duration `json:"idle-timeout" yaml:"idle-timeout" mapstructure:"idle-timeout"`
	MaxHeaderBytes    int           `json:"max-header-bytes" yaml:"max-header-bytes" mapstructure:"max-header-bytes"`
	ShutdownTimeout   time.Duration `json:"shutdown-timeout" yaml:"shutdown-timeout" mapstructure:"shutdown-timeout"`

	// Database fields (from CfgXormDrv)
	DriverName       string        `json:"driver-name" yaml:"driver-name" mapstructure:"driver-name"`
	UseSpinLog       bool          `json:"use-spin-log" yaml:"use-spin-log" mapstructure:"use-spin-log"`
	ClubSourceName   string        `json:"club-source-name" yaml:"club-source-name" mapstructure:"club-source-name"`
	SpinSourceName   string        `json:"spin-source-name" yaml:"spin-source-name" mapstructure:"spin-source-name"`
	SqlFlushTick     time.Duration `json:"sql-flush-tick" yaml:"sql-flush-tick" mapstructure:"sql-flush-tick"`
	ClubUpdateBuffer int           `json:"club-update-buffer" yaml:"club-update-buffer" mapstructure:"club-update-buffer"`
	ClubInsertBuffer int           `json:"club-insert-buffer" yaml:"club-insert-buffer" mapstructure:"club-insert-buffer"`
	SpinInsertBuffer int           `json:"spin-insert-buffer" yaml:"spin-insert-buffer" mapstructure:"spin-insert-buffer"`

	// Gameplay fields (from CfgGameplay)
	AdjunctLimit    float64 `json:"adjunct-limit" yaml:"adjunct-limit" mapstructure:"adjunct-limit"`
	MinJackpot      float64 `json:"min-jackpot" yaml:"min-jackpot" mapstructure:"min-jackpot"`
	MaxSpinAttempts int     `json:"max-spin-attempts" yaml:"max-spin-attempts" mapstructure:"max-spin-attempts"`

	// Cloudinary fields
	CloudName    string `json:"cloud_name" yaml:"cloud_name" mapstructure:"cloud_name"`
	APIKey       string `json:"api_key" yaml:"api_key" mapstructure:"api_key"`
	APISecret    string `json:"api_secret" yaml:"api_secret" mapstructure:"api_secret"`
	UploadFolder string `json:"upload_folder" yaml:"upload_folder" mapstructure:"upload_folder"`

	// Uploads fields
	AllowedTypes []string `json:"allowed_types" yaml:"allowed_types" mapstructure:"allowed_types"`
	MaxFileSize  int64    `json:"max_file_size" yaml:"max_file_size" mapstructure:"max_file_size"`
	Storage      string   `json:"storage" yaml:"storage" mapstructure:"storage"`
}

// Cfg is the global config instance with defaults.
var Cfg = &Config{
	// Authentication defaults
	AccessTTL:    24 * time.Hour,
	RefreshTTL:   72 * time.Hour,
	AccessKey:    "skJgM4NsbP3fs4k7vh0gfdkgGl8dJTszdLxZ1sQ9ksFnxbgvw2RsGH8xxddUV479",
	RefreshKey:   "zxK4dUnuq3Lhd1Gzhpr3usI5lAzgvy2t3fmxld2spzz7a5nfv0hsksm9cheyutie",
	NonceTimeout: 150 * time.Second,

	// Activation defaults
	UseActivation:      false,
	BrevoApiKey:        "",
	BrevoEmailEndpoint: "https://api.brevo.com/v3/smtp/email",
	SenderName:         "Slotopol server",
	SenderEmail:        "noreply@slotopol.com",
	ReplytoEmail:       "noreply@slotopol.com",
	EmailSubject:       "Slotopol verification code",
	EmailHtmlContent:   "<html><head></head><body><p>Your Slotopol verification code is: <b>%06d</b></p></body></html>",
	CodeTimeout:        15 * time.Minute,

	// Web Server defaults
	TrustedProxies:    []string{"127.0.0.0/8"},
	PortHTTP:          []string{":8080"},
	ReadTimeout:       15 * time.Second,
	ReadHeaderTimeout: 15 * time.Second,
	WriteTimeout:      15 * time.Second,
	IdleTimeout:       60 * time.Second,
	MaxHeaderBytes:    1 << 20,
	ShutdownTimeout:   15 * time.Second,

	// Database defaults
	DriverName:       "sqlite3",
	UseSpinLog:       true,
	ClubSourceName:   "slot-club.sqlite",
	SpinSourceName:   "slot-spin.sqlite",
	SqlFlushTick:     2500 * time.Millisecond,
	ClubUpdateBuffer: 200,
	ClubInsertBuffer: 150,
	SpinInsertBuffer: 250,

	// Gameplay defaults
	AdjunctLimit:    100000,
	MinJackpot:      10000,
	MaxSpinAttempts: 300,

	// Cloudinary defaults
	CloudName:    "",
	APIKey:       "",
	APISecret:    "",
	UploadFolder: "slotopol",

	// Uploads defaults
	AllowedTypes: []string{"image/jpeg", "image/png", "image/webp", "image/gif", "image/svg+xml"},
	MaxFileSize:  5 * 1024 * 1024,
	Storage:      "cloudinary",
}

// Program paths.
var (
	BuildVers string
	BuildTime string
	ExePath   string
	CfgPath   string
)

// Default master RTP if no others found.
const DefMRTP = 95.0

// InitConfig initializes the configuration.
func InitConfig() {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix("SLOTOPOL")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(nil)

	// Try to read config file if specified
	if CfgFile != "" {
		v.SetConfigFile(CfgFile)
		if err := v.ReadInConfig(); err == nil {
			if err := v.Unmarshal(Cfg); err == nil {
				return
			}
		}
	}

	// Try default path
	defaultPath := filepath.Join(".", "appdata", "slot-app.yaml")
	if _, err := os.Stat(defaultPath); err == nil {
		v.SetConfigFile(defaultPath)
		if err := v.ReadInConfig(); err == nil {
			if err := v.Unmarshal(Cfg); err == nil {
				return
			}
		}
	}

	// No config file found - use environment variables only
	if err := v.Unmarshal(Cfg); err != nil {
		// Fallback: use defaults
	}
}

// DirExists checks if a directory exists.
func DirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// LoadConfig is kept for compatibility.
func LoadConfig(path string) error {
	return nil
}

// GetClubDB returns a xorm engine for club database.
func (cfg *Config) GetClubDB() (*xorm.Engine, error) {
	return xorm.NewEngine(cfg.DriverName, cfg.ClubSourceName)
}

// GetSpinDB returns a xorm engine for spin database.
func (cfg *Config) GetSpinDB() (*xorm.Engine, error) {
	return xorm.NewEngine(cfg.DriverName, cfg.SpinSourceName)
}

func init() {
	var err error
	if ExePath, err = os.Executable(); err != nil {
		panic(err)
	}
	ExePath = filepath.Dir(ExePath)

	if CfgPath, err = filepath.Abs("."); err != nil {
		panic(err)
	}
}
