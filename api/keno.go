package api

import (
	"encoding/json"
	"encoding/xml"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/MyTeleProject2026/Slotopol-server/config"
	"github.com/MyTeleProject2026/Slotopol-server/game/keno"
	"github.com/MyTeleProject2026/Slotopol-server/util"
)

// Returns bet value.
func ApiKenoBetGet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Bet     float64  `json:"bet" yaml:"bet" xml:"bet"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_keno_betget_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_keno_betget_noscene, err)
		return
	}
	var game keno.KenoGame
	if game, ok = scene.Game.(keno.KenoGame); !ok {
		Ret403(c, AEC_keno_betget_notslot, ErrNotKeno)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_keno_betget_noaccess, ErrNoAccess)
		return
	}

	ret.Bet = game.GetBet()

	RetOk(c, ret)
}
