package api

import (
	"encoding/xml"

	"github.com/gin-gonic/gin"
)

// ApiClubList returns list of all clubs.
func ApiClubList(c *gin.Context) {
	var ret struct {
		XMLName xml.Name    `json:"-" yaml:"-" xml:"ret"`
		Clubs   []*ClubData `json:"clubs" yaml:"clubs" xml:"clubs>club"`
	}

	admin, al := MustAdmin(c, 0)
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
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
	}

	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Exists  bool     `json:"exists" yaml:"exists" xml:"exists"`
	}

	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_club_is_nobind, err)
		return
	}

	admin, al := MustAdmin(c, arg.CID)
	if admin == nil || al&ALadmin == 0 {
		Ret403(c, AEC_club_is_noaccess, ErrNoAccess)
		return
	}

	_, ret.Exists = Clubs.Get(arg.CID)
	RetOk(c, ret)
}

// ApiClubInfo returns detailed information about a club.
func ApiClubInfo(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
	}

	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Club    ClubData `json:"club" yaml:"club" xml:"club"`
	}

	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_club_info_nobind, err)
		return
	}

	club, ok := Clubs.Get(arg.CID)
	if !ok {
		Ret404(c, AEC_club_info_noclub, ErrNoClub)
		return
	}

	admin, al := MustAdmin(c, arg.CID)
	if admin == nil || al&ALadmin == 0 {
		Ret403(c, AEC_club_info_noaccess, ErrNoAccess)
		return
	}

	ret.Club = club.Get()
	RetOk(c, ret)
}

// ApiClubJpfund updates the jackpot fund for a club.
func ApiClubJpfund(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
		Fund    float64  `json:"fund" yaml:"fund" xml:"fund" binding:"required"`
	}

	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_club_jpfund_nobind, err)
		return
	}

	club, ok := Clubs.Get(arg.CID)
	if !ok {
		Ret404(c, AEC_club_jpfund_noclub, ErrNoClub)
		return
	}

	admin, al := MustAdmin(c, arg.CID)
	if admin == nil || al&ALmaster == 0 {
		Ret403(c, AEC_club_jpfund_noaccess, ErrNoAccess)
		return
	}

	// Update the in-memory club fund.
	club.AddFund(arg.Fund - club.Fund())

	// Persist the fund directly to the club table.
	// BankBat/SqlBank does not provide a Fund method.
	if _, err := XormStorage.
		ID(arg.CID).
		Cols("fund").
		Update(&ClubData{Fund: arg.Fund}); err != nil {
		Ret500(c, AEC_club_jpfund_sql, err)
		return
	}

	Ret204(c)
}

// ApiClubRename changes the name of a club.
func ApiClubRename(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
		Name    string   `json:"name" yaml:"name" xml:"name" binding:"required"`
	}

	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_club_rename_nobind, err)
		return
	}

	club, ok := Clubs.Get(arg.CID)
	if !ok {
		Ret404(c, AEC_club_rename_noclub, ErrNoClub)
		return
	}

	admin, al := MustAdmin(c, arg.CID)
	if admin == nil || al&ALadmin == 0 {
		Ret403(c, AEC_club_rename_noaccess, ErrNoAccess)
		return
	}

	// Update the in-memory club name.
	club.SetName(arg.Name)

	// Persist the name to the club table.
	if _, err := XormStorage.
		ID(arg.CID).
		Cols("name").
		Update(&ClubData{Name: arg.Name}); err != nil {
		Ret500(c, AEC_club_rename_sql, err)
		return
	}

	Ret204(c)
}

// ApiClubCashin moves money to/from club deposit.
func ApiClubCashin(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
		Sum     float64  `json:"sum" yaml:"sum" xml:"sum" binding:"required"`
	}

	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_club_cashin_nobind, err)
		return
	}

	if arg.Sum > Cfg.AdjunctLimit || arg.Sum < -Cfg.AdjunctLimit {
		Ret400(c, AEC_club_cashin_limit, ErrTooBig)
		return
	}

	club, ok := Clubs.Get(arg.CID)
	if !ok {
		Ret404(c, AEC_club_cashin_noclub, ErrNoClub)
		return
	}

	admin, al := MustAdmin(c, arg.CID)
	if admin == nil || al&ALmaster == 0 {
		Ret403(c, AEC_club_cashin_noaccess, ErrNoAccess)
		return
	}

	// Update the in-memory club state.
	club.AddDeposit(arg.Sum)
	club.AddBank(-arg.Sum)

	// Persist the bank/deposit change atomically.
	if err := SafeTransaction(XormStorage, func(session *Session) error {
		_, err := session.Exec(
			"UPDATE club SET bank=bank-?, lock=lock+? WHERE cid=?",
			arg.Sum,
			arg.Sum,
			arg.CID,
		)
		return err
	}); err != nil {
		Ret500(c, AEC_club_cashin_sql, err)
		return
	}

	Ret204(c)
}
