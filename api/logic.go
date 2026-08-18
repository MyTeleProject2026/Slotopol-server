package api

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"

	"github.com/MyTeleProject2026/Slotopol-server/config"
	"github.com/MyTeleProject2026/Slotopol-server/game"
	"github.com/MyTeleProject2026/Slotopol-server/util"
)

type ClubData struct {
	CID   uint64    `xorm:"pk autoincr" json:"cid" yaml:"cid" xml:"cid,attr"`
	CTime time.Time `xorm:"created 'ctime' notnull default CURRENT_TIMESTAMP" json:"ctime" yaml:"ctime" xml:"ctime"`
	UTime time.Time `xorm:"updated 'utime' notnull default CURRENT_TIMESTAMP" json:"utime" yaml:"utime" xml:"utime"`
	Name  string    `xorm:"notnull" json:"name,omitempty" yaml:"name,omitempty" xml:"name,omitempty"`
	Bank  float64   `xorm:"notnull default 0" json:"bank" yaml:"bank" xml:"bank"`
	Fund  float64   `xorm:"notnull default 0" json:"fund" yaml:"fund" xml:"fund"`
	Lock  float64   `xorm:"notnull default 0" json:"lock" yaml:"lock" xml:"lock"`
	Rate  float64   `xorm:"'rate' notnull default 2.5" json:"rate" yaml:"rate" xml:"rate"`
	MRTP  float64   `xorm:"'mrtp' notnull default 0" json:"mrtp" yaml:"mrtp" xml:"mrtp"`
}

func (ClubData) TableName() string {
	return "club"
}

type Club struct {
	data ClubData
	mux  sync.RWMutex
}

type UF uint

const (
	UFactivated UF = 1 << iota
	UFsigncode
)

type User struct {
	UID    uint64    `xorm:"pk autoincr" json:"uid" yaml:"uid" xml:"uid,attr"`
	CTime  time.Time `xorm:"created 'ctime' notnull default CURRENT_TIMESTAMP" json:"ctime" yaml:"ctime" xml:"ctime"`
	UTime  time.Time `xorm:"updated 'utime' notnull default CURRENT_TIMESTAMP" json:"utime" yaml:"utime" xml:"utime"`
	Email  string    `xorm:"notnull unique index" json:"email" yaml:"email" xml:"email"`
	Secret string    `xorm:"notnull" json:"secret" yaml:"secret" xml:"secret"`
	Name   string    `xorm:"notnull" json:"name,omitempty" yaml:"name,omitempty" xml:"name,omitempty"`
	Code   uint32    `xorm:"notnull default 0" json:"code,omitempty" yaml:"code,omitempty" xml:"code,omitempty"`
	Status UF        `xorm:"notnull default 0" json:"status,omitempty" yaml:"status,omitempty" xml:"status,omitempty"`
	GAL    AL        `xorm:"notnull default 0" json:"gal,omitempty" yaml:"gal,omitempty" xml:"gal,omitempty"`
	props  util.RWMap[uint64, *Props]
}

type Story struct {
	GID   uint64    `xorm:"pk" json:"gid" yaml:"gid" xml:"gid,attr"`
	CID   uint64    `xorm:"notnull" json:"cid" yaml:"cid" xml:"cid,attr"`
	UID   uint64    `xorm:"notnull" json:"uid" yaml:"uid" xml:"uid,attr"`
	Alias string    `xorm:"notnull" json:"alias" yaml:"alias" xml:"alias"`
	CTime time.Time `xorm:"created 'ctime' notnull default CURRENT_TIMESTAMP" json:"ctime" yaml:"ctime" xml:"ctime"`
}

var StoryCounter atomic.Uint64

type Scene struct {
	Story `yaml:",inline"`
	SID   uint64      `json:"sid" yaml:"sid" xml:"sid,attr"`
	Game  game.Gamble `json:"game" yaml:"game" xml:"game"`
}

type AL uint

const (
	ALmember AL = 1 << iota
	ALdealer
	ALbooker
	ALmaster
	ALadmin
	ALall = ALmember | ALdealer | ALbooker | ALmaster | ALadmin
)

type Props struct {
	CID    uint64    `xorm:"notnull index(bid)" json:"cid" yaml:"cid" xml:"cid,attr"`
	UID    uint64    `xorm:"notnull index(bid)" json:"uid" yaml:"uid" xml:"uid,attr"`
	CTime  time.Time `xorm:"created 'ctime' notnull default CURRENT_TIMESTAMP" json:"ctime" yaml:"ctime" xml:"ctime"`
	UTime  time.Time `xorm:"updated 'utime' notnull default CURRENT_TIMESTAMP" json:"utime" yaml:"utime" xml:"utime"`
	Wallet float64   `xorm:"notnull default 0" json:"wallet" yaml:"wallet" xml:"wallet"`
	Access AL        `xorm:"notnull default 0" json:"access" yaml:"access" xml:"access"`
	MRTP   float64   `xorm:"notnull default 0" json:"mrtp" yaml:"mrtp" xml:"mrtp"`
}

var PropMaster []Props

type Spinlog struct {
	SID    uint64    `xorm:"pk" json:"sid" yaml:"sid" xml:"sid,attr"`
	GID    uint64    `xorm:"notnull" json:"gid" yaml:"gid" xml:"gid,attr"`
	CTime  time.Time `xorm:"created 'ctime' notnull default CURRENT_TIMESTAMP" json:"ctime" yaml:"ctime" xml:"ctime"`
	MRTP   float64   `xorm:"notnull" json:"mrtp" yaml:"mrtp" xml:"mrtp,attr"`
	Game   string    `xorm:"notnull" json:"game" yaml:"game" xml:"game"`
	Wins   string    `xorm:"text" json:"wins,omitempty" yaml:"wins,omitempty" xml:"wins,omitempty"`
	Gain   float64   `xorm:"notnull" json:"gain" yaml:"gain" xml:"gain"`
	Wallet float64   `xorm:"notnull" json:"wallet" yaml:"wallet" xml:"wallet"`
}

var SpinCounter atomic.Uint64

type Multlog struct {
	ID     uint64    `xorm:"pk" json:"id" yaml:"id" xml:"id,attr"`
	GID    uint64    `xorm:"notnull" json:"gid" yaml:"gid" xml:"gid,attr"`
	CTime  time.Time `xorm:"created 'ctime' notnull default CURRENT_TIMESTAMP" json:"ctime" yaml:"ctime" xml:"ctime"`
	MRTP   float64   `xorm:"notnull" json:"mrtp" yaml:"mrtp" xml:"mrtp,attr"`
	Mult   float64   `xorm:"notnull" json:"mult" yaml:"mult" xml:"mult"`
	Risk   float64   `xorm:"notnull" json:"risk" yaml:"risk" xml:"risk"`
	Win    bool      `xorm:"notnull" json:"win" yaml:"win" xml:"win"`
	Gain   float64   `xorm:"notnull" json:"gain" yaml:"gain" xml:"gain"`
	Wallet float64   `xorm:"notnull" json:"wallet" yaml:"wallet" xml:"wallet"`
}

var MultCounter atomic.Uint64

type Walletlog struct {
	ID     uint64    `xorm:"pk autoincr" json:"id" yaml:"id" xml:"id,attr"`
	CID    uint64    `xorm:"notnull index(bid)" json:"cid" yaml:"cid" xml:"cid,attr"`
	UID    uint64    `xorm:"notnull index(bid)" json:"uid" yaml:"uid" xml:"uid,attr"`
	AID    uint64    `xorm:"notnull" json:"aid" yaml:"aid" xml:"aid"`
	CTime  time.Time `xorm:"created 'ctime' notnull default CURRENT_TIMESTAMP" json:"ctime" yaml:"ctime" xml:"ctime"`
	Wallet float64   `xorm:"notnull" json:"wallet" yaml:"wallet" xml:"wallet"`
	Sum    float64   `xorm:"notnull" json:"sum" yaml:"sum" xml:"sum"`
}

type Banklog struct {
	ID      uint64    `xorm:"pk autoincr" json:"id" yaml:"id" xml:"id,attr"`
	CTime   time.Time `xorm:"created 'ctime' notnull default CURRENT_TIMESTAMP" json:"ctime" yaml:"ctime" xml:"ctime"`
	Bank    float64   `xorm:"notnull 'bank'" json:"bank" yaml:"bank" xml:"bank"`
	Fund    float64   `xorm:"notnull 'fund'" json:"fund" yaml:"fund" xml:"fund"`
	Lock    float64   `xorm:"notnull 'lock'" json:"lock" yaml:"lock" xml:"lock"`
	BankSum float64   `xorm:"notnull 'banksum'" json:"banksum" yaml:"banksum" xml:"banksum" form:"banksum"`
	FundSum float64   `xorm:"notnull 'fundsum'" json:"fundsum" yaml:"fundsum" xml:"fundsum" form:"fundsum"`
	LockSum float64   `xorm:"notnull 'locksum'" json:"locksum" yaml:"locksum" xml:"locksum" form:"locksum"`
}

var Clubs util.RWMap[uint64, *Club]
var Users util.RWMap[uint64, *User]
var Scenes util.RWMap[uint64, *Scene]

func MakeClub(cd ClubData) *Club {
	return &Club{data: cd}
}

func (club *Club) Get() ClubData {
	club.mux.RLock()
	defer club.mux.RUnlock()
	return club.data
}

func (club *Club) CID() uint64 {
	return club.data.CID
}

func (club *Club) Name() string {
	club.mux.RLock()
	defer club.mux.RUnlock()
	return club.data.Name
}

func (club *Club) SetName(name string) {
	club.mux.Lock()
	defer club.mux.Unlock()
	club.data.Name = name
}

func (club *Club) Bank() float64 {
	club.mux.RLock()
	defer club.mux.RUnlock()
	return club.data.Bank
}

func (club *Club) Fund() float64 {
	club.mux.RLock()
	defer club.mux.RUnlock()
	return club.data.Fund
}

func (club *Club) Deposit() float64 {
	club.mux.RLock()
	defer club.mux.RUnlock()
	return club.data.Lock
}

func (club *Club) GetCash() (bank, fund, deposit float64) {
	club.mux.RLock()
	defer club.mux.RUnlock()
	return club.data.Bank, club.data.Fund, club.data.Lock
}

func (club *Club) AddCash(bank, fund, deposit float64) {
	club.mux.Lock()
	defer club.mux.Unlock()
	club.data.Bank += bank
	club.data.Fund += fund
	club.data.Lock += deposit
}

func (club *Club) AddBank(bank float64) {
	club.mux.Lock()
	defer club.mux.Unlock()
	club.data.Bank += bank
}

func (club *Club) AddFund(fund float64) {
	club.mux.Lock()
	defer club.mux.Unlock()
	club.data.Fund += fund
}

func (club *Club) AddDeposit(deposit float64) {
	club.mux.Lock()
	defer club.mux.Unlock()
	club.data.Lock += deposit
}

func (club *Club) Rate() float64 {
	club.mux.RLock()
	defer club.mux.RUnlock()
	return club.data.Rate
}

func (club *Club) MRTP() float64 {
	club.mux.RLock()
	defer club.mux.RUnlock()
	return club.data.MRTP
}

func (user *User) Init() {
	user.props.Init(0)
}

func (user *User) GetWallet(cid uint64) float64 {
	if props, ok := user.props.Get(cid); ok {
		return props.Wallet
	}
	return 0
}

func (user *User) GetAL(cid uint64) AL {
	if props, ok := user.props.Get(cid); ok {
		return props.Access
	}
	return 0
}

func (user *User) GetRTP(cid uint64) float64 {
	if props, ok := user.props.Get(cid); ok {
		return props.MRTP
	}
	return 0
}

func (user *User) InsertProps(props *Props) {
	user.props.Set(props.CID, props)
}

func GetAdmin(c *gin.Context, cid uint64) (*User, AL) {
	if value, exists := c.Get(userKey); exists {
		var admin = value.(*User)
		return admin, admin.GAL | admin.GetAL(cid)
	}
	return nil, 0
}

func MustAdmin(c *gin.Context, cid uint64) (*User, AL) {
	var admin = c.MustGet(userKey).(*User)
	return admin, admin.GAL | admin.GetAL(cid)
}

func GetRTP(user *User, club *Club) float64 {
	if user != nil {
		if props, ok := user.props.Get(club.CID()); ok && props.MRTP != 0 {
			return props.MRTP
		}
	}
	if mrtp := club.MRTP(); mrtp != 0 {
		return mrtp
	}
	return config.DefMRTP // default master RTP
}

func GetScene(gid uint64) (scene *Scene, err error) {
	var ok bool
	if scene, ok = Scenes.Get(gid); ok {
		return
	}

	var tmp Scene
	if ok, _ = XormStorage.ID(gid).Get(&tmp.Story); !ok {
		err = ErrNotOpened
		return
	}
	var maker func() game.Gamble
	if maker, ok = game.GameFactory[tmp.Alias]; !ok {
		err = ErrNoAliase
		return
	}
	tmp.Game = maker()

	scene = &tmp
	Scenes.Set(gid, scene)

	if !Cfg.UseSpinLog {
		return
	}

	var rec Spinlog
	if ok, _ = XormSpinlog.Where("gid = ?", gid).Desc("ctime").Get(&rec); !ok {
		return
	}
	scene.SID = rec.SID
	err = json.Unmarshal(util.S2B(rec.Game), scene.Game)
	return
}

func init() {
	Clubs.Init(0)
	Users.Init(0)
	Scenes.Init(0)
}
