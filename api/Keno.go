package api

import (
	"encoding/json"
	"encoding/xml"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/MyTeleProject2026/Slotopol-server/game/keno"
	"github.com/MyTeleProject2026/Slotopol-server/util"
)

func kenoGame(c *gin.Context, gid uint64) (*Scene, keno.KenoGame, error) {
	scene, err := GetScene(gid)
	if err != nil {
		return nil, nil, err
	}
	g, ok := scene.Game.(keno.KenoGame)
	if !ok {
		return scene, nil, ErrNotKeno
	}
	return scene, g, nil
}

func kenoAccess(c *gin.Context, scene *Scene) bool {
	admin, al := MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_keno_spin_noaccess, ErrNoAccess)
		return false
	}
	return true
}

func ApiKenoBetGet(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Bet     float64  `json:"bet" yaml:"bet" xml:"bet"`
	}
	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_keno_betget_nobind, err)
		return
	}
	scene, g, err := kenoGame(c, arg.GID)
	if err != nil {
		if scene == nil {
			Ret404(c, AEC_keno_betget_noscene, err)
		} else {
			Ret403(c, AEC_keno_betget_notslot, err)
		}
		return
	}
	admin, al := MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_keno_betget_noaccess, ErrNoAccess)
		return
	}
	ret.Bet = g.GetBet()
	RetOk(c, ret)
}

func ApiKenoBetSet(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" binding:"required"`
		Bet     float64  `json:"bet" yaml:"bet" xml:"bet" binding:"required"`
	}
	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_keno_betset_nobind, err)
		return
	}
	scene, g, err := kenoGame(c, arg.GID)
	if err != nil {
		if scene == nil {
			Ret404(c, AEC_keno_betset_noscene, err)
		} else {
			Ret403(c, AEC_keno_betset_notslot, err)
		}
		return
	}
	admin, al := MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_keno_betset_noaccess, ErrNoAccess)
		return
	}
	if err := g.SetBet(arg.Bet); err != nil {
		Ret403(c, AEC_keno_betset_badbet, err)
		return
	}
	Ret204(c)
}

func ApiKenoSelGet(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name    `json:"-" yaml:"-" xml:"ret"`
		Sel     keno.Bitset `json:"sel" yaml:"sel" xml:"sel"`
	}
	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_keno_selget_nobind, err)
		return
	}
	scene, g, err := kenoGame(c, arg.GID)
	if err != nil {
		if scene == nil {
			Ret404(c, AEC_keno_selget_noscene, err)
		} else {
			Ret403(c, AEC_keno_selget_notslot, err)
		}
		return
	}
	admin, al := MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_keno_selget_noaccess, ErrNoAccess)
		return
	}
	ret.Sel = g.GetSel()
	RetOk(c, ret)
}

func ApiKenoSelGetSlice(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Sel     []int    `json:"sel" yaml:"sel" xml:"sel"`
	}
	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_keno_selgetslice_nobind, err)
		return
	}
	scene, g, err := kenoGame(c, arg.GID)
	if err != nil {
		if scene == nil {
			Ret404(c, AEC_keno_selgetslice_noscene, err)
		} else {
			Ret403(c, AEC_keno_selgetslice_notslot, err)
		}
		return
	}
	admin, al := MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_keno_selgetslice_noaccess, ErrNoAccess)
		return
	}
	sel := g.GetSel()
	ret.Sel = (&sel).Expand()
	RetOk(c, ret)
}

func ApiKenoSelSet(c *gin.Context) {
	var arg struct {
		XMLName xml.Name    `json:"-" yaml:"-" xml:"arg"`
		GID     uint64      `json:"gid" yaml:"gid" xml:"gid,attr" binding:"required"`
		Sel     keno.Bitset `json:"sel" yaml:"sel" xml:"sel" binding:"required"`
	}
	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_keno_selset_nobind, err)
		return
	}
	scene, g, err := kenoGame(c, arg.GID)
	if err != nil {
		if scene == nil {
			Ret404(c, AEC_keno_selset_noscene, err)
		} else {
			Ret403(c, AEC_keno_selset_notslot, err)
		}
		return
	}
	admin, al := MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_keno_selset_noaccess, ErrNoAccess)
		return
	}
	if err := g.SetSel(arg.Sel); err != nil {
		Ret403(c, AEC_keno_selset_badsel, err)
		return
	}
	Ret204(c)
}

func ApiKenoSelSetSlice(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" binding:"required"`
		Sel     []int    `json:"sel" yaml:"sel" xml:"sel" binding:"required"`
	}
	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_keno_selsetslice_nobind, err)
		return
	}
	var sel keno.Bitset
	for _, n := range arg.Sel {
		if n < 0 || n >= 80 {
			Ret403(c, AEC_keno_selsetslice_badsel, keno.ErrBadSel)
			return
		}
		sel.Set(n)
	}
	scene, g, err := kenoGame(c, arg.GID)
	if err != nil {
		if scene == nil {
			Ret404(c, AEC_keno_selsetslice_noscene, err)
		} else {
			Ret403(c, AEC_keno_selsetslice_notslot, err)
		}
		return
	}
	admin, al := MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_keno_selsetslice_noaccess, ErrNoAccess)
		return
	}
	if err := g.SetSel(sel); err != nil {
		Ret403(c, AEC_keno_selsetslice_badsel, err)
		return
	}
	Ret204(c)
}

func ApiKenoSpin(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
		Bet     float64  `json:"bet,omitempty" yaml:"bet,omitempty" xml:"bet,omitempty"`
		Sel     []int    `json:"sel,omitempty" yaml:"sel,omitempty" xml:"sel,omitempty"`
	}
	var ret struct {
		XMLName xml.Name  `json:"-" yaml:"-" xml:"ret"`
		SID     uint64    `json:"sid" yaml:"sid,attr" xml:"sid,attr"`
		Game    any       `json:"game" yaml:"game" xml:"game"`
		Wins    keno.Wins `json:"wins" yaml:"wins" xml:"wins"`
		Wallet  float64   `json:"wallet" yaml:"wallet" xml:"wallet"`
	}
	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_keno_spin_nobind, err)
		return
	}
	scene, g, err := kenoGame(c, arg.GID)
	if err != nil {
		if scene == nil {
			Ret404(c, AEC_keno_spin_noscene, err)
		} else {
			Ret403(c, AEC_keno_spin_notslot, err)
		}
		return
	}
	club, ok := Clubs.Get(scene.CID)
	if !ok {
		Ret500(c, AEC_keno_spin_noclub, ErrNoClub)
		return
	}
	user, ok := Users.Get(scene.UID)
	if !ok {
		Ret500(c, AEC_keno_spin_nouser, ErrNoUser)
		return
	}
	admin, al := MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_keno_spin_noaccess, ErrNoAccess)
		return
	}
	props, ok := user.props.Get(scene.CID)
	if !ok {
		Ret500(c, AEC_keno_spin_noprops, ErrNoProps)
		return
	}

	if arg.Bet != 0 {
		if err := g.SetBet(arg.Bet); err != nil {
			Ret403(c, AEC_keno_spin_badbet, err)
			return
		}
	}
	if len(arg.Sel) > 0 {
		var sel keno.Bitset
		for _, n := range arg.Sel {
			if n < 0 || n >= 80 {
				Ret403(c, AEC_keno_spin_badsel, keno.ErrBadSel)
				return
			}
			sel.Set(n)
		}
		if err := g.SetSel(sel); err != nil {
			Ret403(c, AEC_keno_spin_badsel, err)
			return
		}
	}

	cost := g.GetBet()
	if props.Wallet < cost {
		Ret403(c, AEC_keno_spin_nomoney, ErrNoMoney)
		return
	}

	bank, _, _ := club.GetCash()
	var wins keno.Wins
	var gain, debit float64
	for n := 0; ; n++ {
		if n >= Cfg.MaxSpinAttempts {
			Ret500(c, AEC_keno_spin_badbank, ErrBadBank)
			return
		}
		g.Spin(GetRTP(user, club))
		if err := g.Scanner(&wins); err != nil {
			continue
		}
		gain = wins.Pay
		debit = cost - gain
		if bank+debit >= 0 {
			break
		}
	}

	if Cfg.ClubUpdateBuffer > 1 {
		go BankBat[scene.CID].Put(XormStorage, scene.UID, debit)
	} else if err := BankBat[scene.CID].Put(XormStorage, scene.UID, debit); err != nil {
		Ret500(c, AEC_keno_spin_sqlbank, err)
		return
	}

	club.AddBank(debit)
	props.Wallet += gain - cost

	sid := SpinCounter.Inc()
	scene.SID = sid
	rec := Spinlog{
		SID: sid, GID: arg.GID, MRTP: GetRTP(user, club),
		Gain: gain, Wallet: props.Wallet,
	}
	if b, err := json.Marshal(scene.Game); err != nil {
		Ret500(c, AEC_keno_spin_sqlbank, err)
		return
	} else {
		rec.Game = util.B2S(b)
	}
	if b, err := json.Marshal(wins); err == nil {
		rec.Wins = util.B2S(b)
	}
	if Cfg.UseSpinLog {
		go func(rec Spinlog) {
			if err := SpinBuf.Put(XormSpinlog, rec); err != nil {
				log.Printf("can not write Keno spin log: %s", err)
			}
		}(rec)
	}

	ret.SID = sid
	ret.Game = scene.Game
	ret.Wins = wins
	ret.Wallet = props.Wallet
	RetOk(c, ret)
}
