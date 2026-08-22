package api

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/MyTeleProject2026/Slotopol-server/config"
	"github.com/MyTeleProject2026/Slotopol-server/game"
	"github.com/MyTeleProject2026/Slotopol-server/util"
)

func ApiGameAlgs(c *gin.Context) {
	RetOk(c, game.AlgList)
}

func ApiGameList(c *gin.Context) {
	var err error
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		Include string   `json:"include" yaml:"include" xml:"include" form:"inc"`
		Exclude string   `json:"exclude" yaml:"exclude" xml:"exclude" form:"exc"`
		Sort    bool     `json:"sort" yaml:"sort" xml:"sort" form:"sort"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid" form:"cid"`
	}
	var ret struct {
		XMLName xml.Name         `json:"-" yaml:"-" xml:"ret"`
		List    []*game.GameInfo `json:"list" yaml:"list" xml:"list>gi"`
		AlgNum  int              `json:"algnum" yaml:"algnum" xml:"algnum"`
		PrvNum  int              `json:"prvnum" yaml:"prvnum" xml:"prvnum"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	arg.Include = strings.TrimSpace(arg.Include)
	arg.Exclude = strings.TrimSpace(arg.Exclude)
	if arg.Include == "" {
		arg.Include = "all"
	}
	var include = strings.Fields(arg.Include)
	var exclude = strings.Fields(arg.Exclude)

	var finclist, fexclist [][]game.Filter
	var f game.Filter
	var flist []game.Filter
	for _, inc := range include {
		var keys = strings.Split(inc, "&")
		flist = nil
		for _, key := range keys {
			if f = game.GetFilter(key); f == nil {
				Ret400(c, 0, fmt.Errorf("filter with name '%s' does not recognized", key))
				return
			}
			flist = append(flist, f)
		}
		finclist = append(finclist, flist)
	}
	for _, exc := range exclude {
		var keys = strings.Split(exc, "&")
		flist = nil
		for _, key := range keys {
			if f = game.GetFilter(key); f == nil {
				Ret400(c, 0, fmt.Errorf("filter with name '%s' does not recognized", key))
				return
			}
			flist = append(flist, f)
		}
		fexclist = append(fexclist, flist)
	}

	var alg = map[*game.AlgDescr]int{}
	var prov = map[string]int{}
	var gamelist = make([]*game.GameInfo, 0, 256)
	for _, gi := range game.InfoMap {
		if game.Passes(gi, finclist, fexclist) {
			alg[gi.AlgDescr]++
			prov[util.ToID(gi.Prov)]++
			gamelist = append(gamelist, gi)
		}
	}

	sort.Slice(gamelist, func(i, j int) bool {
		var gii, gij = gamelist[i], gamelist[j]
		if arg.Sort {
			if gii.Prov == gij.Prov {
				return gii.Name < gij.Name
			}
			return gii.Prov < gij.Prov
		} else {
			if gii.Name == gij.Name {
				return gii.Prov < gij.Prov
			}
			return gii.Name < gij.Name
		}
	})

	// Filter by club permissions
	if arg.CID != 0 {
		var enabled []string
		err := XormStorage.Table("club_game_permissions").
			Where("club_id=? AND enabled=1", arg.CID).
			Select("game_alias").Find(&enabled)

		// If table doesn't exist or any error, just show all games
		if err != nil {
			ret.List = gamelist
			ret.AlgNum = len(alg)
			ret.PrvNum = len(prov)
			RetOk(c, ret)
			return
		}

		enabledSet := make(map[string]bool)
		for _, alias := range enabled {
			enabledSet[alias] = true
		}

		// Force enable all games if no permissions exist
		if len(enabled) == 0 {
			for _, gi := range gamelist {
				enabledSet[gi.GameAlias.ID()] = true
			}
		}

		filtered := make([]*game.GameInfo, 0, len(gamelist))
		for _, gi := range gamelist {
			if enabledSet[gi.GameAlias.ID()] {
				filtered = append(filtered, gi)
			}
		}
		gamelist = filtered
	}

	ret.List = gamelist
	ret.AlgNum = len(alg)
	ret.PrvNum = len(prov)

	RetOk(c, ret)
}

// ... rest of file unchanged
