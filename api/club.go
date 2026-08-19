package api

import (
	"encoding/xml"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/MyTeleProject2026/Slotopol-server/config"
)

// ApiClubList returns list of all clubs.
func ApiClubList(c *gin.Context) {
	var ret struct {
		XMLName xml.Name           `json:"-" yaml:"-" xml:"ret"`
		Clubs   []*ClubData `json:"clubs" yaml:"clubs" xml:"clubs>club"`
	}

	var admin, al = MustAdmin(c, 0)
	if admin == nil || al&ALadmin == 0 {
		Ret403(c, AEC_club_list_noaccess, ErrNoAccess)
		return
	}

	for _, club := range Clubs.Items() {
		ret.Clubs = append(ret.Clubs, &club.data)
	}

	RetOk(c, ret)
}

// ApiClubIs checks if a club exists.
func ApiClubIs(c *gin.Context) {
	var err error
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Exists  bool     `json:"exists" yaml:"exists" xml:"exists"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_club_is_nobind, err)
		return
	}

	var admin, al = MustAdmin(c, arg.CID)
	if admin == nil || al&ALadmin == 0 {
		Ret403(c, AEC_club_is_noaccess, ErrNoAccess)
		return
	}

	_, ret.Exists = Clubs.Get(arg.CID)
	RetOk(c, ret)
}

// ApiClubInfo returns detailed information about a club.
func ApiClubInfo(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Club    ClubData `json:"club" yaml:"club" xml:"club"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_club_info_nobind, err)
		return
	}

	var club *Club
	if club, ok = Clubs.Get(arg.CID); !ok {
		Ret404(c, AEC_club_info_noclub, ErrNoClub)
		return
	}

	var admin, al = MustAdmin(c, arg.CID)
	if admin == nil || al&ALadmin == 0 {
		Ret403(c, AEC_club_info_noaccess, ErrNoAccess)
		return
	}

	ret.Club = club.Get()
	RetOk(c, ret)
}

// ApiClubJpfund updates the jackpot fund for a club.
func ApiClubJpfund(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
		Fund    float64  `json:"fund" yaml:"fund" xml:"fund" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_club_jpfund_nobind, err)
		return
	}

	var club *Club
	if club, ok = Clubs.Get(arg.CID); !ok {
		Ret404(c, AEC_club_jpfund_noclub, ErrNoClub)
		return
	}

	var admin, al = MustAdmin(c, arg.CID)
	if admin == nil || al&ALmaster == 0 {
		Ret403(c, AEC_club_jpfund_noaccess, ErrNoAccess)
		return
	}

	// Update fund in memory
	club.AddFund(arg.Fund - club.Fund())

	// Persist to database via BankBat
	if Cfg.ClubInsertBuffer > 1 {
		go BankBat[arg.CID].Fund(XormStorage, arg.Fund)
	} else if err = BankBat[arg.CID].Fund(XormStorage, arg.Fund); err != nil {
		Ret500(c, AEC_club_jpfund_sql, err)
		return
	}

	Ret204(c)
}

// ApiClubRename changes the name of a club.
func ApiClubRename(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
		Name    string   `json:"name" yaml:"name" xml:"name" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_club_rename_nobind, err)
		return
	}

	var club *Club
	if club, ok = Clubs.Get(arg.CID); !ok {
		Ret404(c, AEC_club_rename_noclub, ErrNoClub)
		return
	}

	var admin, al = MustAdmin(c, arg.CID)
	if admin == nil || al&ALadmin == 0 {
		Ret403(c, AEC_club_rename_noaccess, ErrNoAccess)
		return
	}

	// Update name in memory
	club.SetName(arg.Name)

	// Update in database
	if _, err = XormStorage.ID(arg.CID).Cols("name").Update(&ClubData{Name: arg.Name}); err != nil {
		Ret500(c, AEC_club_rename_sql, err)
		return
	}

	Ret204(c)
}

// ApiClubCashin moves money to/from club deposit.
func ApiClubCashin(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
		Sum     float64  `json:"sum" yaml:"sum" xml:"sum" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_club_cashin_nobind, err)
		return
	}
	if arg.Sum > Cfg.AdjunctLimit || arg.Sum < -Cfg.AdjunctLimit {
		Ret400(c, AEC_club_cashin_limit, ErrTooBig)
		return
	}

	var club *Club
	if club, ok = Clubs.Get(arg.CID); !ok {
		Ret404(c, AEC_club_cashin_noclub, ErrNoClub)
		return
	}

	var admin, al = MustAdmin(c, arg.CID)
	if admin == nil || al&ALmaster == 0 {
		Ret403(c, AEC_club_cashin_noaccess, ErrNoAccess)
		return
	}

	// Update memory
	club.AddDeposit(arg.Sum)
	club.AddBank(-arg.Sum) // deposit taken from bank

	// Persist via BankBat (we need to define BankBat.Cashin or use direct SQL)
	if err = SafeTransaction(XormStorage, func(session *Session) error {
		_, err := session.Exec("UPDATE club SET bank=bank-?, lock=lock+? WHERE cid=?", arg.Sum, arg.Sum, arg.CID)
		return err
	}); err != nil {
		Ret500(c, AEC_club_cashin_sql, err)
		return
	}

	Ret204(c)
}
