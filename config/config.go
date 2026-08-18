package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"xorm.io/xorm"
)

// ============================================================
// Command-line flags and exports for cmd package
// ============================================================

// AppName is the application name used in CLI.
var AppName = "slotopol"

// Command-line flags (set by cobra)
var (
	CfgFile string   // path to config file (ignored now, but kept for compatibility)
	SqlPath string   // path to sqlite databases
	ObjPath []string // additional yaml paths
	Verbose bool     // verbose logging
)

// ============================================================
// Configuration Structures – Flat Fields
// ============================================================

// Config holds all configuration – all fields are flat (not nested)
// to match the existing code that uses Cfg.AccessKey, Cfg.SenderName, etc.
type Config struct {
	// Authentication
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
	AccessKey    string
	RefreshKey   string
	NonceTimeout time.Duration

	// Activation / Email
	UseActivation      bool
	BrevoApiKey        string
	BrevoEmailEndpoint string
	SenderName         string
	SenderEmail        string
	ReplytoEmail       string
	EmailSubject       string
	EmailHtmlContent   string
	CodeTimeout        time.Duration

	// Web Server
	TrustedProxies    []string
	PortHTTP          []string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
	ShutdownTimeout   time.Duration

	// Database
	DriverName       string
	UseSpinLog       bool
	ClubSourceName   string
	SpinSourceName   string
	SqlFlushTick     time.Duration
	ClubUpdateBuffer int
	ClubInsertBuffer int
	SpinInsertBuffer int

	// Gameplay
	AdjunctLimit    float64
	MinJackpot      float64
	MaxSpinAttempts int

	// Cloudinary
	CloudName    string
	APIKey       string
	APISecret    string
	UploadFolder string

	// Uploads
	AllowedTypes []string
	MaxFileSize  int64
	Storage      string
}

// Cfg is the global config instance.
var Cfg = &Config{}

// Program paths.
var (
	BuildVers string
	BuildTime string
	ExePath   string
	CfgPath   string
)

// Default master RTP if no others found.
const DefMRTP = 95.0

// InitConfig initializes the configuration from environment variables.
// This is a simple, robust implementation that does not use Viper.
func InitConfig() {
	// Set defaults first (so env vars can override)
	setDefaults()

	// Override with environment variables (if set)
	// Authentication
	if v := os.Getenv("SLOTOPOL_ACCESS_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			Cfg.AccessTTL = d
		}
	}
	if v := os.Getenv("SLOTOPOL_REFRESH_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			Cfg.RefreshTTL = d
		}
	}
	if v := os.Getenv("SLOTOPOL_ACCESS_KEY"); v != "" {
		Cfg.AccessKey = v
	}
	if v := os.Getenv("SLOTOPOL_REFRESH_KEY"); v != "" {
		Cfg.RefreshKey = v
	}
	if v := os.Getenv("SLOTOPOL_NONCE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			Cfg.NonceTimeout = d
		}
	}

	// Activation
	if v := os.Getenv("SLOTOPOL_USE_ACTIVATION"); v != "" {
		Cfg.UseActivation = strings.ToLower(v) == "true"
	}
	if v := os.Getenv("SLOTOPOL_BREVO_API_KEY"); v != "" {
		Cfg.BrevoApiKey = v
	}
	if v := os.Getenv("SLOTOPOL_BREVO_EMAIL_ENDPOINT"); v != "" {
		Cfg.BrevoEmailEndpoint = v
	}
	if v := os.Getenv("SLOTOPOL_SENDER_NAME"); v != "" {
		Cfg.SenderName = v
	}
	if v := os.Getenv("SLOTOPOL_SENDER_EMAIL"); v != "" {
		Cfg.SenderEmail = v
	}
	if v := os.Getenv("SLOTOPOL_REPLYTO_EMAIL"); v != "" {
		Cfg.ReplytoEmail = v
	}
	if v := os.Getenv("SLOTOPOL_EMAIL_SUBJECT"); v != "" {
		Cfg.EmailSubject = v
	}
	if v := os.Getenv("SLOTOPOL_EMAIL_HTML_CONTENT"); v != "" {
		Cfg.EmailHtmlContent = v
	}
	if v := os.Getenv("SLOTOPOL_CODE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			Cfg.CodeTimeout = d
		}
	}

	// Web Server
	if v := os.Getenv("SLOTOPOL_TRUSTED_PROXIES"); v != "" {
		Cfg.TrustedProxies = strings.Split(v, ",")
	}
	if v := os.Getenv("SLOTOPOL_PORT_HTTP"); v != "" {
		Cfg.PortHTTP = strings.Split(v, ",")
	}
	if v := os.Getenv("SLOTOPOL_READ_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			Cfg.ReadTimeout = d
		}
	}
	if v := os.Getenv("SLOTOPOL_READ_HEADER_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			Cfg.ReadHeaderTimeout = d
		}
	}
	if v := os.Getenv("SLOTOPOL_WRITE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			Cfg.WriteTimeout = d
		}
	}
	if v := os.Getenv("SLOTOPOL_IDLE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			Cfg.IdleTimeout = d
		}
	}
	if v := os.Getenv("SLOTOPOL_MAX_HEADER_BYTES"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			Cfg.MaxHeaderBytes = i
		}
	}
	if v := os.Getenv("SLOTOPOL_SHUTDOWN_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			Cfg.ShutdownTimeout = d
		}
	}

	// Database
	if v := os.Getenv("SLOTOPOL_DBDRIVER"); v != "" {
		Cfg.DriverName = v
	}
	if v := os.Getenv("SLOTOPOL_USE_SPIN_LOG"); v != "" {
		Cfg.UseSpinLog = strings.ToLower(v) == "true"
	}
	if v := os.Getenv("SLOTOPOL_CLUBDSN"); v != "" {
		Cfg.ClubSourceName = v
	}
	if v := os.Getenv("SLOTOPOL_SPINDSN"); v != "" {
		Cfg.SpinSourceName = v
	}
	if v := os.Getenv("SLOTOPOL_SQL_FLUSH_TICK"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			Cfg.SqlFlushTick = d
		}
	}
	if v := os.Getenv("SLOTOPOL_CLUB_UPDATE_BUFFER"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			Cfg.ClubUpdateBuffer = i
		}
	}
	if v := os.Getenv("SLOTOPOL_CLUB_INSERT_BUFFER"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			Cfg.ClubInsertBuffer = i
		}
	}
	if v := os.Getenv("SLOTOPOL_SPIN_INSERT_BUFFER"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			Cfg.SpinInsertBuffer = i
		}
	}

	// Gameplay
	if v := os.Getenv("SLOTOPOL_ADJUNCT_LIMIT"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			Cfg.AdjunctLimit = f
		}
	}
	if v := os.Getenv("SLOTOPOL_MIN_JACKPOT"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			Cfg.MinJackpot = f
		}
	}
	if v := os.Getenv("SLOTOPOL_MAX_SPIN_ATTEMPTS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			Cfg.MaxSpinAttempts = i
		}
	}

	// Cloudinary
	if v := os.Getenv("SLOTOPOL_CLOUDINARY_CLOUD_NAME"); v != "" {
		Cfg.CloudName = v
	}
	if v := os.Getenv("SLOTOPOL_CLOUDINARY_API_KEY"); v != "" {
		Cfg.APIKey = v
	}
	if v := os.Getenv("SLOTOPOL_CLOUDINARY_API_SECRET"); v != "" {
		Cfg.APISecret = v
	}
	if v := os.Getenv("SLOTOPOL_CLOUDINARY_UPLOAD_FOLDER"); v != "" {
		Cfg.UploadFolder = v
	}

	// Uploads
	if v := os.Getenv("SLOTOPOL_ALLOWED_TYPES"); v != "" {
		Cfg.AllowedTypes = strings.Split(v, ",")
	}
	if v := os.Getenv("SLOTOPOL_MAX_FILE_SIZE"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			Cfg.MaxFileSize = i
		}
	}
	if v := os.Getenv("SLOTOPOL_STORAGE"); v != "" {
		Cfg.Storage = v
	}
}

// setDefaults fills Cfg with sensible defaults.
func setDefaults() {
	Cfg.AccessTTL = 24 * time.Hour
	Cfg.RefreshTTL = 72 * time.Hour
	Cfg.AccessKey = "skJgM4NsbP3fs4k7vh0gfdkgGl8dJTszdLxZ1sQ9ksFnxbgvw2RsGH8xxddUV479"
	Cfg.RefreshKey = "zxK4dUnuq3Lhd1Gzhpr3usI5lAzgvy2t3fmxld2spzz7a5nfv0hsksm9cheyutie"
	Cfg.NonceTimeout = 150 * time.Second

	Cfg.UseActivation = false
	Cfg.BrevoApiKey = ""
	Cfg.BrevoEmailEndpoint = "https://api.brevo.com/v3/smtp/email"
	Cfg.SenderName = "Slotopol server"
	Cfg.SenderEmail = "noreply@slotopol.com"
	Cfg.ReplytoEmail = "noreply@slotopol.com"
	Cfg.EmailSubject = "Slotopol verification code"
	Cfg.EmailHtmlContent = "<html><head></head><body><p>Your Slotopol verification code is: <b>%06d</b></p></body></html>"
	Cfg.CodeTimeout = 15 * time.Minute

	Cfg.TrustedProxies = []string{"127.0.0.0/8"}
	Cfg.PortHTTP = []string{":8080"}
	Cfg.ReadTimeout = 15 * time.Second
	Cfg.ReadHeaderTimeout = 15 * time.Second
	Cfg.WriteTimeout = 15 * time.Second
	Cfg.IdleTimeout = 60 * time.Second
	Cfg.MaxHeaderBytes = 1 << 20
	Cfg.ShutdownTimeout = 15 * time.Second

	Cfg.DriverName = "sqlite3"
	Cfg.UseSpinLog = true
	Cfg.ClubSourceName = "slot-club.sqlite"
	Cfg.SpinSourceName = "slot-spin.sqlite"
	Cfg.SqlFlushTick = 2500 * time.Millisecond
	Cfg.ClubUpdateBuffer = 200
	Cfg.ClubInsertBuffer = 150
	Cfg.SpinInsertBuffer = 250

	Cfg.AdjunctLimit = 100000
	Cfg.MinJackpot = 10000
	Cfg.MaxSpinAttempts = 300

	Cfg.CloudName = ""
	Cfg.APIKey = ""
	Cfg.APISecret = ""
	Cfg.UploadFolder = "slotopol"

	Cfg.AllowedTypes = []string{"image/jpeg", "image/png", "image/webp", "image/gif", "image/svg+xml"}
	Cfg.MaxFileSize = 5 * 1024 * 1024 // 5MB
	Cfg.Storage = "cloudinary"
}

// LoadConfig is kept for compatibility (not used).
func LoadConfig(path string) error {
	return nil
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
