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
// Configuration Structures
// ============================================================

type CfgJwtAuth struct {
	AccessTTL    time.Duration `mapstructure:"access-ttl"`
	RefreshTTL   time.Duration `mapstructure:"refresh-ttl"`
	AccessKey    string        `mapstructure:"access-key"`
	RefreshKey   string        `mapstructure:"refresh-key"`
	NonceTimeout time.Duration `mapstructure:"nonce-timeout"`
}

type CfgSendCode struct {
	UseActivation      bool          `mapstructure:"use-activation"`
	BrevoApiKey        string        `mapstructure:"brevo-api-key"`
	BrevoEmailEndpoint string        `mapstructure:"brevo-email-endpoint"`
	SenderName         string        `mapstructure:"sender-name"`
	SenderEmail        string        `mapstructure:"sender-email"`
	ReplytoEmail       string        `mapstructure:"replyto-email"`
	EmailSubject       string        `mapstructure:"email-subject"`
	EmailHtmlContent   string        `mapstructure:"email-html-content"`
	CodeTimeout        time.Duration `mapstructure:"code-timeout"`
}

type CfgWebServ struct {
	TrustedProxies    []string      `mapstructure:"trusted-proxies"`
	PortHTTP          []string      `mapstructure:"port-http"`
	ReadTimeout       time.Duration `mapstructure:"read-timeout"`
	ReadHeaderTimeout time.Duration `mapstructure:"read-header-timeout"`
	WriteTimeout      time.Duration `mapstructure:"write-timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle-timeout"`
	MaxHeaderBytes    int           `mapstructure:"max-header-bytes"`
	ShutdownTimeout   time.Duration `mapstructure:"shutdown-timeout"`
}

type CfgXormDrv struct {
	DriverName       string        `mapstructure:"driver-name"`
	UseSpinLog       bool          `mapstructure:"use-spin-log"`
	ClubSourceName   string        `mapstructure:"club-source-name"`
	SpinSourceName   string        `mapstructure:"spin-source-name"`
	SqlFlushTick     time.Duration `mapstructure:"sql-flush-tick"`
	ClubUpdateBuffer int           `mapstructure:"club-update-buffer"`
	ClubInsertBuffer int           `mapstructure:"club-insert-buffer"`
	SpinInsertBuffer int           `mapstructure:"spin-insert-buffer"`
}

type CfgGameplay struct {
	AdjunctLimit    float64 `mapstructure:"adjunct-limit"`
	MinJackpot      float64 `mapstructure:"min-jackpot"`
	MaxSpinAttempts int     `mapstructure:"max-spin-attempts"`
}

type CfgCloudinary struct {
	CloudName    string `mapstructure:"cloud_name"`
	APIKey       string `mapstructure:"api_key"`
	APISecret    string `mapstructure:"api_secret"`
	UploadFolder string `mapstructure:"upload_folder"`
}

type CfgUploads struct {
	AllowedTypes []string `mapstructure:"allowed_types"`
	MaxFileSize  int64    `mapstructure:"max_file_size"`
	Storage      string   `mapstructure:"storage"`
}

// Config is common service settings.
type Config struct {
	Authentication CfgJwtAuth  `mapstructure:"authentication"`
	Activation     CfgSendCode `mapstructure:"activation"`
	WebServer      CfgWebServ  `mapstructure:"web-server"`
	Database       CfgXormDrv  `mapstructure:"database"`
	Gameplay       CfgGameplay `mapstructure:"gameplay"`
	Cloudinary     CfgCloudinary `mapstructure:"cloudinary"`
	Uploads        CfgUploads    `mapstructure:"uploads"`
}

// Global config instance with defaults
var Cfg = &Config{
	Authentication: CfgJwtAuth{
		AccessTTL:    24 * time.Hour,
		RefreshTTL:   72 * time.Hour,
		AccessKey:    "skJgM4NsbP3fs4k7vh0gfdkgGl8dJTszdLxZ1sQ9ksFnxbgvw2RsGH8xxddUV479",
		RefreshKey:   "zxK4dUnuq3Lhd1Gzhpr3usI5lAzgvy2t3fmxld2spzz7a5nfv0hsksm9cheyutie",
		NonceTimeout: 150 * time.Second,
	},
	Activation: CfgSendCode{
		UseActivation:      false,
		BrevoApiKey:        "",
		BrevoEmailEndpoint: "https://api.brevo.com/v3/smtp/email",
		SenderName:         "Slotopol server",
		SenderEmail:        "noreply@slotopol.com",
		ReplytoEmail:       "noreply@slotopol.com",
		EmailSubject:       "Slotopol verification code",
		EmailHtmlContent:   "<html><head></head><body><p>Your Slotopol verification code is: <b>%06d</b></p></body></html>",
		CodeTimeout:        15 * time.Minute,
	},
	WebServer: CfgWebServ{
		TrustedProxies:    []string{"127.0.0.0/8"},
		PortHTTP:          []string{":8080"},
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
		ShutdownTimeout:   15 * time.Second,
	},
	Database: CfgXormDrv{
		DriverName:       "sqlite3",
		UseSpinLog:       true,
		ClubSourceName:   "slot-club.sqlite",
		SpinSourceName:   "slot-spin.sqlite",
		SqlFlushTick:     2500 * time.Millisecond,
		ClubUpdateBuffer: 200,
		ClubInsertBuffer: 150,
		SpinInsertBuffer: 250,
	},
	Gameplay: CfgGameplay{
		AdjunctLimit:    100000,
		MinJackpot:      10000,
		MaxSpinAttempts: 300,
	},
	Cloudinary: CfgCloudinary{
		CloudName:    "",
		APIKey:       "",
		APISecret:    "",
		UploadFolder: "slotopol",
	},
	Uploads: CfgUploads{
		AllowedTypes: []string{"image/jpeg", "image/png", "image/webp", "image/gif", "image/svg+xml"},
		MaxFileSize:  5 * 1024 * 1024,
		Storage:      "cloudinary",
	},
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
	// Bind environment variables to config struct
	if err := v.Unmarshal(Cfg); err != nil {
		// Fallback: use defaults, environment variables will be read directly
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

// LoadConfig is kept for compatibility but not used directly.
func LoadConfig(path string) error {
	return nil
}

// GetClubDB returns a xorm engine for club database.
func (cfg *Config) GetClubDB() (*xorm.Engine, error) {
	return xorm.NewEngine(cfg.Database.DriverName, cfg.Database.ClubSourceName)
}

// GetSpinDB returns a xorm engine for spin database.
func (cfg *Config) GetSpinDB() (*xorm.Engine, error) {
	return xorm.NewEngine(cfg.Database.DriverName, cfg.Database.SpinSourceName)
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
