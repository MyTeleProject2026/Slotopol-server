package api

import (
	"encoding/xml"

	"github.com/gin-gonic/gin"
)

// ApiPropsGet returns all properties for a user at a club.
func ApiPropsGet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
		UID     uint64   `json:"uid" yaml:"uid" xml:"uid,attr" form:"uid" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Wallet  float64  `json:"wallet" yaml:"wallet" xml:"wallet"`
		Access  AL       `json:"access" yaml:"access" xml:"access"`
		MRTP    float64  `json:"mrtp" yaml:"mrtp" xml:"mrtp"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	var club *Club
	if club, ok = Clubs.Get(arg.CID); !ok {
		Ret404(c, 0, ErrNoClub)
		return
	}
	_ = club

	var user *User
	if user, ok = Users.Get(arg.UID); !ok {
		Ret404(c, 0, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, arg.CID)
	if admin != user && al&ALbooker == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	if props, ok := user.props.Get(arg.CID); ok {
		ret.Wallet = props.Wallet
		ret.Access = props.Access
		ret.MRTP = props.MRTP
	}

	RetOk(c, ret)
}

// ApiPropsWalletGet returns balance at wallet for a user at a club.
func ApiPropsWalletGet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
		UID     uint64   `json:"uid" yaml:"uid" xml:"uid,attr" form:"uid" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Wallet  float64  `json:"wallet" yaml:"wallet" xml:"wallet"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	var club *Club
	if club, ok = Clubs.Get(arg.CID); !ok {
		Ret404(c, 0, ErrNoClub)
		return
	}
	_ = club

	var user *User
	if user, ok = Users.Get(arg.UID); !ok {
		Ret404(c, 0, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, arg.CID)
	if admin != user && al&ALbooker == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	ret.Wallet = user.GetWallet(arg.CID)

	RetOk(c, ret)
}

// ApiPropsWalletAdd adds (or removes) coins from a user's wallet.
func ApiPropsWalletAdd(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
		UID     uint64   `json:"uid" yaml:"uid" xml:"uid,attr" form:"uid" binding:"required"`
		Sum     float64  `json:"sum" yaml:"sum" xml:"sum" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Wallet  float64  `json:"wallet" yaml:"wallet" xml:"wallet"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}
	if arg.Sum > Cfg.AdjunctLimit || arg.Sum < -Cfg.AdjunctLimit {
		Ret400(c, 0, ErrTooBig)
		return
	}

	var club *Club
	if club, ok = Clubs.Get(arg.CID); !ok {
		Ret404(c, 0, ErrNoClub)
		return
	}
	_ = club

	var user *User
	if user, ok = Users.Get(arg.UID); !ok {
		Ret404(c, 0, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, arg.CID)
	if al&ALbooker == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	var props *Props
	if props, ok = user.props.Get(arg.CID); !ok {
		Ret500(c, 0, ErrNoProps)
		return
	}
	if props.Wallet+arg.Sum < 0 {
		Ret403(c, 0, ErrNoMoney)
		return
	}

	// Update wallet as transaction
	if Cfg.ClubInsertBuffer > 1 {
		go BankBat[arg.CID].Add(XormStorage, arg.UID, admin.UID, props.Wallet+arg.Sum, arg.Sum)
	} else if err = BankBat[arg.CID].Add(XormStorage, arg.UID, admin.UID, props.Wallet+arg.Sum, arg.Sum); err != nil {
		Ret500(c, 0, err)
		return
	}

	// Make changes to memory data
	props.Wallet += arg.Sum

	ret.Wallet = props.Wallet

	RetOk(c, ret)
}

// ApiPropsAlGet returns personal access level for a user at a club.
func ApiPropsAlGet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid"`
		UID     uint64   `json:"uid" yaml:"uid" xml:"uid,attr" form:"uid" binding:"required"`
		All     bool     `json:"all" yaml:"all" xml:"all,attr" form:"all"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Access  AL       `json:"access" yaml:"access" xml:"access"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	var club *Club
	if club, ok = Clubs.Get(arg.CID); !ok && !arg.All {
		Ret404(c, 0, ErrNoClub)
		return
	}
	_ = club

	var user *User
	if user, ok = Users.Get(arg.UID); !ok {
		Ret404(c, 0, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, arg.CID)
	if admin != user && al&(ALbooker+ALadmin) == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	ret.Access = user.GetAL(arg.CID)
	if arg.All {
		ret.Access |= user.GAL
	}

	RetOk(c, ret)
}

// ApiPropsAlSet sets personal access level for a user at a club.
func ApiPropsAlSet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
		UID     uint64   `json:"uid" yaml:"uid" xml:"uid,attr" form:"uid" binding:"required"`
		Access  AL       `json:"access" yaml:"access" xml:"access"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	var club *Club
	if club, ok = Clubs.Get(arg.CID); !ok {
		Ret404(c, 0, ErrNoClub)
		return
	}
	_ = club

	var user *User
	if user, ok = Users.Get(arg.UID); !ok {
		Ret404(c, 0, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, arg.CID)
	if al&ALadmin == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}
	_ = admin

	var props *Props
	if props, ok = user.props.Get(arg.CID); !ok {
		Ret500(c, 0, ErrNoProps)
		return
	}
	if al&arg.Access != arg.Access {
		Ret403(c, 0, ErrNoLevel)
		return
	}

	// Update access level as transaction
	if Cfg.ClubInsertBuffer > 1 {
		go BankBat[arg.CID].Access(XormStorage, arg.UID, arg.Access)
	} else if err = BankBat[arg.CID].Access(XormStorage, arg.UID, arg.Access); err != nil {
		Ret500(c, 0, err)
		return
	}

	// Make changes to memory data
	props.Access = arg.Access

	Ret204(c)
}

// ApiPropsRtpGet returns master RTP for a user at a club.
func ApiPropsRtpGet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
		UID     uint64   `json:"uid" yaml:"uid" xml:"uid,attr" form:"uid"`
		All     bool     `json:"all" yaml:"all" xml:"all,attr" form:"all"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		MRTP    float64  `json:"mrtp" yaml:"mrtp" xml:"mrtp"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	var club *Club
	if club, ok = Clubs.Get(arg.CID); !ok {
		Ret404(c, 0, ErrNoClub)
		return
	}
	_ = club

	var user *User
	if user, ok = Users.Get(arg.UID); !ok && !arg.All {
		Ret404(c, 0, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, arg.CID)
	if admin != user && al&ALbooker == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	if arg.All {
		ret.MRTP = GetRTP(user, club)
	} else {
		ret.MRTP = user.GetRTP(arg.CID)
	}

	RetOk(c, ret)
}

// ApiPropsRtpSet sets personal master RTP for a user at a club.
func ApiPropsRtpSet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid"`
		UID     uint64   `json:"uid" yaml:"uid" xml:"uid,attr" form:"uid"`
		MRTP    float64  `json:"mrtp" yaml:"mrtp" xml:"mrtp"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	var club *Club
	if club, ok = Clubs.Get(arg.CID); !ok {
		Ret404(c, 0, ErrNoClub)
		return
	}
	_ = club

	var user *User
	if user, ok = Users.Get(arg.UID); !ok {
		Ret404(c, 0, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, arg.CID)
	if al&ALbooker == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}
	_ = admin

	var props *Props
	if props, ok = user.props.Get(arg.CID); !ok {
		Ret500(c, 0, ErrNoProps)
		return
	}

	// Update master RTP as transaction
	if Cfg.ClubInsertBuffer > 1 {
		go BankBat[arg.CID].MRTP(XormStorage, arg.UID, arg.MRTP)
	} else if err = BankBat[arg.CID].MRTP(XormStorage, arg.UID, arg.MRTP); err != nil {
		Ret500(c, 0, err)
		return
	}

	// Make changes to memory data
	props.MRTP = arg.MRTP

	Ret204(c)
}
