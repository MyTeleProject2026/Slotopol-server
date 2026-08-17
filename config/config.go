package cfg

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"xorm.io/xorm"
)

var (
	// compiled binary version, sets by compiler with command
	//    go build -ldflags="-X 'github.com/slotopol/server/config.BuildVers=%buildvers%'"
	BuildVers string

	// compiled binary build date, sets by compiler with command
	//    go build -ldflags="-X 'github.com/slotopol/server/config.BuildTime=%buildtime%'"
	BuildTime string
)

// Default master RTP if no others found.
const DefMRTP = 95.0

// Program paths.
var (
	ExePath string
	CfgPath string
	SqlPath string
)

type CfgJwtAuth struct {
	AccessTTL    time.Duration `json:"access-ttl" yaml:"access-ttl" mapstructure:"access-ttl"`
	RefreshTTL   time.Duration `json:"refresh-ttl" yaml:"refresh-ttl" mapstructure:"refresh-ttl"`
	AccessKey    string        `json:"access-key" yaml:"access-key" mapstructure:"access-key"`
	RefreshKey   string        `json:"refresh-key" yaml:"refresh-key" mapstructure:"refresh-key"`
	NonceTimeout time.Duration `json:"nonce-timeout" yaml:"nonce-timeout" mapstructure:"nonce-timeout"`
}

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

type CfgGameplay struct {
	AdjunctLimit    float64 `json:"adjunct-limit" yaml:"adjunct-limit" mapstructure:"adjunct-limit"`
	MinJackpot      float64 `json:"min-jackpot" yaml:"min-jackpot" mapstructure:"min-jackpot"`
	MaxSpinAttempts int     `json:"max-spin-attempts" yaml:"max-spin-attempts" mapstructure:"max-spin-attempts"`
}

// Cloudinary configuration
type CfgCloudinary struct {
	CloudName    string `json:"cloud_name" yaml:"cloud_name" mapstructure:"cloud_name"`
	APIKey       string `json:"api_key" yaml:"api_key" mapstructure:"api_key"`
	APISecret    string `json:"api_secret" yaml:"api_secret" mapstructure:"api_secret"`
	UploadFolder string `json:"upload_folder" yaml:"upload_folder" mapstructure:"upload_folder"`
}

// Uploads configuration
type CfgUploads struct {
	AllowedTypes []string `json:"allowed_types" yaml:"allowed_types" mapstructure:"allowed_types"`
	MaxFileSize  int64    `json:"max_file_size" yaml:"max_file_size" mapstructure:"max_file_size"`
	Storage      string   `json:"storage" yaml:"storage" mapstructure:"storage"`
}

// Config is common service settings.
type Config struct {
	CfgJwtAuth  `json:"authentication" yaml:"authentication" mapstructure:"authentication"`
	CfgSendCode `json:"activation" yaml:"activation" mapstructure:"activation"`
	CfgWebServ  `json:"web-server" yaml:"web-server" mapstructure:"web-server"`
	CfgXormDrv  `json:"database" yaml:"database" mapstructure:"database"`
	CfgGameplay `json:"gameplay" yaml:"gameplay" mapstructure:"gameplay"`

	// NEW: Cloudinary
	Cloudinary CfgCloudinary `json:"cloudinary" yaml:"cloudinary" mapstructure:"cloudinary"`

	// NEW: Uploads
	Uploads CfgUploads `json:"uploads" yaml:"uploads" mapstructure:"uploads"`
}

// Instance of common service settings with defaults.
var Cfg = &Config{
	CfgJwtAuth: CfgJwtAuth{
		AccessTTL:    24 * time.Hour,
		RefreshTTL:   72 * time.Hour,
		AccessKey:    "skJgM4NsbP3fs4k7vh0gfdkgGl8dJTszdLxZ1sQ9ksFnxbgvw2RsGH8xxddUV479",
		RefreshKey:   "zxK4dUnuq3Lhd1Gzhpr3usI5lAzgvy2t3fmxld2spzz7a5nfv0hsksm9cheyutie",
		NonceTimeout: 150 * time.Second,
	},
	CfgSendCode: CfgSendCode{
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
	CfgWebServ: CfgWebServ{
		TrustedProxies:    []string{"127.0.0.0/8"},
		PortHTTP:          []string{":8080"},
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
		ShutdownTimeout:   15 * time.Second,
	},
	CfgXormDrv: CfgXormDrv{
		DriverName:       "sqlite3",
		UseSpinLog:       true,
		ClubSourceName:   "slot-club.sqlite",
		SpinSourceName:   "slot-spin.sqlite",
		SqlFlushTick:     2500 * time.Millisecond,
		ClubUpdateBuffer: 200,
		ClubInsertBuffer: 150,
		SpinInsertBuffer: 250,
	},
	CfgGameplay: CfgGameplay{
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
		AllowedTypes: []string{
			"image/jpeg",
			"image/png",
			"image/webp",
			"image/gif",
			"image/svg+xml",
		},
		MaxFileSize: 5 * 1024 * 1024, // 5MB
		Storage:     "cloudinary",
	},
}

// LoadConfig reads configuration from YAML file.
func LoadConfig(path string) error {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("SLOTOPOL")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(nil)

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	if err := v.Unmarshal(Cfg); err != nil {
		return err
	}
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
