package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/MyTeleProject2026/Slotopol-server/api"
	"github.com/MyTeleProject2026/Slotopol-server/config"
	"github.com/MyTeleProject2026/Slotopol-server/game"
	"github.com/MyTeleProject2026/Slotopol-server/util"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

var (
	Cfg = config.Cfg // shortcut
)

var (
	ErrNoClubName = errors.New("name of 'club' database does not provided at data source name")
	ErrNoSpinName = errors.New("name of 'spin' database does not provided at data source name")
)

func Startup() (exitctx context.Context) {
	exitctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()

		var sigint = make(chan os.Signal, 1)
		var sigterm = make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT)
		signal.Notify(sigterm, syscall.SIGTERM)
		select {
		case <-exitctx.Done():
			if errors.Is(exitctx.Err(), context.DeadlineExceeded) {
				log.Println("shutting down by timeout")
			} else if errors.Is(exitctx.Err(), context.Canceled) {
				log.Println("shutting down by cancel")
			} else {
				log.Printf("shutting down by %s\n", exitctx.Err().Error())
			}
		case <-sigint:
			log.Println("shutting down by break")
		case <-sigterm:
			log.Println("shutting down by process termination")
		}
		signal.Stop(sigint)
		signal.Stop(sigterm)
	}()
	return
}

func LoadInternalYaml(ctx context.Context) {
	var t0 = time.Now()
	var size int
	for _, b := range game.LoadMap {
		if ctx.Err() != nil {
			return
		}
		game.MustReadChain(bytes.NewReader(b))
		size += len(b)
	}
	var d = time.Since(t0)
	if config.Verbose {
		log.Printf("loaded %d embedded yaml files in %s on %d bytes\n", len(game.LoadMap), d.String(), size)
	}
}

func LoadYamlFromFile(fullpath string) (err error) {
	if ext := util.ToLower(filepath.Ext(fullpath)); ext != ".yaml" && ext != ".yml" {
		return nil
	}
	var r io.ReadCloser
	if r, err = os.Open(fullpath); err != nil {
		return err
	}
	defer r.Close()
	if err = game.ReadChain(r); err != nil {
		return fmt.Errorf("can not read data from %s: %w", fullpath, err)
	}
	if config.Verbose {
		log.Printf("loaded data from: %s\n", fullpath)
	}
	return nil
}

func LoadExternalYaml(ctx context.Context) (err error) {
	for _, root := range config.ObjPath {
		root, _ = util.ExpandHomePath(root)
		var isdir bool
		if isdir, err = config.DirExists(root); err != nil {
			return
		}
		if !isdir {
			return LoadYamlFromFile(root)
		}
		err = fs.WalkDir(os.DirFS(root), ".", func(fpath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if err = ctx.Err(); err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			var fullpath = filepath.Join(root, fpath)
			return LoadYamlFromFile(fullpath)
		})
		if err != nil {
			return
		}
	}
	return
}

func UpdateAlgList() {
	for _, ai := range game.AlgList {
		if ai.Update != nil {
			ai.Update(ai)
		}
	}
}

func CheckAlgList() {
	for _, ai := range game.AlgList {
		if len(ai.RTP) == 0 {
			var id string
			if len(ai.Aliases) > 0 {
				id = ai.Aliases[0].ID()
			}
			panic(fmt.Errorf("RTP list does not complete for %s", id))
		}
	}
}

func InitStorage() (err error) {
	switch Cfg.DriverName {
	case "sqlite3":
		var fpath string
		if Cfg.ClubSourceName != ":memory:" {
			fpath = util.JoinPath(config.SqlPath, Cfg.ClubSourceName)
		} else {
			fpath = Cfg.ClubSourceName
		}
		if api.XormStorage, err = xorm.NewEngine(Cfg.DriverName, fpath); err != nil {
			return
		}
		if Cfg.ClubSourceName != ":memory:" {
			log.Println("club db: sqlite")
		} else {
			log.Println("club db: memory")
		}

	case "mysql", "postgres":
		if api.XormStorage, err = xorm.NewEngine(Cfg.DriverName, Cfg.ClubSourceName); err != nil {
			return
		}
		log.Printf("club db: %s\n", Cfg.DriverName)
	}
	api.XormStorage.SetMapper(names.GonicMapper{})

	var session = api.XormStorage.NewSession()
	defer session.Close()

	if err = session.Sync(
		&api.ClubData{}, &api.User{}, &api.Props{},
		&api.Story{}, api.Walletlog{}, api.Banklog{},
	); err != nil {
		return
	}

	// --- FIX: use absolute path to appdata for initialization files ---
	appdataDir := "appdata"
	// If running in Render, the binary is in the root, appdata is a subdirectory.
	// If local, it might be in the same directory as binary; we'll check both.
	if _, err := os.Stat(appdataDir); os.IsNotExist(err) {
		// Try relative to executable path
		appdataDir = filepath.Join(config.ExePath, "appdata")
	}

	// Check if club table is empty; if so, run the full init SQL
	var clubEmpty bool
	if clubEmpty, err = session.IsTableEmpty(&api.ClubData{}); err != nil {
		return
	}
	if clubEmpty {
		// Load the SQL file
		sqlPath := filepath.Join(appdataDir, "slot-clubinit.sql")
		var body []byte
		if body, err = os.ReadFile(sqlPath); err != nil {
			log.Printf("can not open SQL-file with initial settings: %s", err.Error())
			// Allow continue, but we will also try to insert defaults if missing
		} else {
			var list = bytes.Split(body, []byte{';'})
			for _, cmd := range list {
				if cmd = bytes.TrimSpace(cmd); len(cmd) > 0 {
					if _, err = session.Exec(util.B2S(cmd)); err != nil {
						return
					}
				}
			}
			log.Println("initialized club table with SQL")
		}
	}

	// --- NEW: Ensure default users exist even if club table was not empty ---
	// This is to fix the "props without user" error when users table is empty.
	var userCount int64
	if userCount, err = session.Count(&api.User{}); err != nil {
		return
	}
	if userCount == 0 {
		log.Println("users table is empty, inserting default users and props")

		// Insert default users
		defaultUsers := []api.User{
			{UID: 1, Email: "admin@slotopol.com", Secret: "admin123", Name: "Administrator", Status: api.UFactivated, GAL: 31},
			{UID: 2, Email: "dealer@slotopol.com", Secret: "dealer123", Name: "Dealer", Status: api.UFactivated, GAL: 3},
			{UID: 3, Email: "player@slotopol.com", Secret: "player123", Name: "Player", Status: api.UFactivated, GAL: 1},
		}
		for _, u := range defaultUsers {
			if _, err = session.Insert(&u); err != nil {
				return
			}
		}

		// Insert default props (assuming club 1 exists)
		defaultProps := []api.Props{
			{CID: 1, UID: 1, Wallet: 100000, Access: 31, MRTP: 0},
			{CID: 2, UID: 1, Wallet: 0, Access: 0, MRTP: 0},
			{CID: 1, UID: 2, Wallet: 10000, Access: 13, MRTP: 0},
			{CID: 2, UID: 2, Wallet: 0, Access: 0, MRTP: 0},
			{CID: 1, UID: 3, Wallet: 1000, Access: 1, MRTP: 98},
			{CID: 2, UID: 3, Wallet: 0, Access: 0, MRTP: 98},
		}
		for _, p := range defaultProps {
			if _, err = session.Insert(&p); err != nil {
				return
			}
		}

		log.Println("default users and props inserted")
	}

	// Load properties master for new registered user (from YAML)
	yamlPath := filepath.Join(appdataDir, "slot-newuser.yaml")
	var body []byte
	if body, err = os.ReadFile(yamlPath); err != nil {
		log.Printf("can not open YAML-file with properties initialization for new user: %s", err.Error())
		// Not fatal – we already inserted default users
		err = nil // remove error
	} else if err = yaml.Unmarshal(body, &api.PropMaster); err != nil {
		log.Printf("can not unmarshal 'slot-newuser.yaml': %s", err.Error())
		err = nil // remove error
	}

	const limit = 256

	var offset = 0
	for {
		var chunk []api.ClubData
		if err = session.Limit(limit, offset).Find(&chunk); err != nil {
			return
		}
		offset += limit
		for _, cd := range chunk {
			api.Clubs.Set(cd.CID, api.MakeClub(cd))
			var bat = &api.SqlBank{}
			bat.Init(cd.CID, Cfg.ClubUpdateBuffer, Cfg.ClubInsertBuffer)
			api.BankBat[cd.CID] = bat
		}
		if limit > len(chunk) {
			break
		}
	}
	log.Printf("loaded %d clubs\n", api.Clubs.Len())

	offset = 0
	for {
		var chunk []*api.User
		if err = session.Limit(limit, offset).Find(&chunk); err != nil {
			return
		}
		offset += limit
		for _, user := range chunk {
			user.Init()
			api.Users.Set(user.UID, user)
		}
		if limit > len(chunk) {
			break
		}
	}
	log.Printf("loaded %d users\n", api.Users.Len())

	offset = 0
	for {
		var chunk []*api.Props
		if err = session.Limit(limit, offset).Find(&chunk); err != nil {
			return
		}
		offset += limit
		for _, props := range chunk {
			if !api.Clubs.Has(props.CID) {
				return fmt.Errorf("found props without club linkage, UID=%d, CID=%d, value=%g", props.UID, props.CID, props.Wallet)
			}
			var user, ok = api.Users.Get(props.UID)
			if !ok {
				// Skip props for missing users (shouldn't happen now, but safety)
				log.Printf("warning: skipping props for missing user UID=%d, CID=%d", props.UID, props.CID)
				continue
			}
			user.InsertProps(props)
		}
		if limit > len(chunk) {
			break
		}
	}

	var i64 int64
	if i64, err = session.Count(&api.Story{}); err != nil {
		return
	}
	api.StoryCounter.Store(uint64(i64))

	api.JoinBuf.Init(Cfg.ClubInsertBuffer)
	return
}

func InitSpinlog() (err error) {
	switch Cfg.DriverName {
	case "sqlite3":
		var fpath string
		if Cfg.SpinSourceName != ":memory:" {
			fpath = util.JoinPath(config.SqlPath, Cfg.SpinSourceName)
		} else {
			fpath = Cfg.SpinSourceName
		}
		if api.XormSpinlog, err = xorm.NewEngine(Cfg.DriverName, fpath); err != nil {
			return
		}
		if Cfg.ClubSourceName != ":memory:" {
			log.Println("spin db: sqlite")
		} else {
			log.Println("spin db: memory")
		}

	case "mysql", "postgres":
		if api.XormSpinlog, err = xorm.NewEngine(Cfg.DriverName, Cfg.SpinSourceName); err != nil {
			return
		}
		log.Printf("spin db: %s\n", Cfg.DriverName)
	}
	api.XormSpinlog.SetMapper(names.GonicMapper{})

	var session = api.XormSpinlog.NewSession()
	defer session.Close()

	if err = session.Sync(&api.Spinlog{}, &api.Multlog{}); err != nil {
		return
	}
	var i64 int64
	if i64, err = session.Count(&api.Spinlog{}); err != nil {
		return
	}
	api.SpinCounter.Store(uint64(i64))
	if i64, err = session.Count(&api.Multlog{}); err != nil {
		return
	}
	api.MultCounter.Store(uint64(i64))

	api.SpinBuf.Init(Cfg.SpinInsertBuffer)
	api.MultBuf.Init(Cfg.SpinInsertBuffer)
	return
}

func SqlLoop(exitctx context.Context) {
	var fd = Cfg.SqlFlushTick
	var flush = time.Tick(fd)
	var passers = time.Tick(time.Hour * 8)
	for {
		select {
		case <-flush:
			for cid, bat := range api.BankBat {
				if err := bat.Flush(api.XormStorage, fd); err != nil {
					log.Printf("can not update bank for cid=%d: %s", cid, err.Error())
				}
			}
			if err := api.JoinBuf.Flush(api.XormStorage, fd); err != nil {
				log.Printf("can not write to story log: %s", err.Error())
			}
			if Cfg.UseSpinLog {
				if err := api.SpinBuf.Flush(api.XormSpinlog, fd); err != nil {
					log.Printf("can not write to spin log: %s", err.Error())
				}
				if err := api.MultBuf.Flush(api.XormSpinlog, fd); err != nil {
					log.Printf("can not write to mult log: %s", err.Error())
				}
			}
		case <-passers:
			api.XormStorage.Where("ctime<? AND status=0", time.Now().Add(-time.Hour*3*24).Format(time.DateTime)).Delete(&api.User{})
		case <-exitctx.Done():
			return
		}
	}
}

func InitSQL() (err error) {
	if err = InitStorage(); err != nil {
		err = fmt.Errorf("can not init XORM records storage: %w", err)
		return
	}
	if Cfg.SpinSourceName == "" {
		Cfg.UseSpinLog = false
	}
	if Cfg.UseSpinLog {
		if err = InitSpinlog(); err != nil {
			err = fmt.Errorf("can not init XORM spins log storage: %w", err)
			return
		}
	}
	return
}

func DoneSQL() (err error) {
	var errs []error
	for _, bat := range api.BankBat {
		errs = append(errs, bat.Flush(api.XormStorage, 0))
	}
	errs = append(errs, api.JoinBuf.Flush(api.XormStorage, 0))
	errs = append(errs, api.XormStorage.Close())

	if Cfg.UseSpinLog {
		errs = append(errs, api.SpinBuf.Flush(api.XormSpinlog, 0))
		errs = append(errs, api.MultBuf.Flush(api.XormSpinlog, 0))
		errs = append(errs, api.XormSpinlog.Close())
	}
	return errors.Join(errs...)
}
