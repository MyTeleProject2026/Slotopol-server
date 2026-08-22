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
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		List    []gin.H  `json:"list" yaml:"list" xml:"list>gi"`
		AlgNum  int      `json:"algnum" yaml:"algnum" xml:"algnum"`
		PrvNum  int      `json:"prvnum" yaml:"prvnum" xml:"prvnum"`
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

	include := strings.Fields(arg.Include)
	exclude := strings.Fields(arg.Exclude)

	var finclist, fexclist [][]game.Filter
	var f game.Filter
	var flist []game.Filter

	for _, inc := range include {
		keys := strings.Split(inc, "&")
		flist = nil

		for _, key := range keys {
			if f = game.GetFilter(key); f == nil {
				Ret400(
					c,
					0,
					fmt.Errorf(
						"filter with name '%s' does not recognized",
						key,
					),
				)
				return
			}

			flist = append(flist, f)
		}

		finclist = append(finclist, flist)
	}

	for _, exc := range exclude {
		keys := strings.Split(exc, "&")
		flist = nil

		for _, key := range keys {
			if f = game.GetFilter(key); f == nil {
				Ret400(
					c,
					0,
					fmt.Errorf(
						"filter with name '%s' does not recognized",
						key,
					),
				)
				return
			}

			flist = append(flist, f)
		}

		fexclist = append(fexclist, flist)
	}

	alg := map[*game.AlgDescr]int{}
	prov := map[string]int{}

	gamelist := make([]*game.GameInfo, 0, 256)

	for _, gi := range game.InfoMap {
		if game.Passes(gi, finclist, fexclist) {
			alg[gi.AlgDescr]++
			prov[util.ToID(gi.Prov)]++
			gamelist = append(gamelist, gi)
		}
	}

	sort.Slice(gamelist, func(i, j int) bool {
		gii := gamelist[i]
		gij := gamelist[j]

		if arg.Sort {
			if gii.Prov == gij.Prov {
				return gii.Name < gij.Name
			}

			return gii.Prov < gij.Prov
		}

		if gii.Name == gij.Name {
			return gii.Prov < gij.Prov
		}

		return gii.Name < gij.Name
	})

	permissionMap := make(map[string]bool)

	if arg.CID != 0 {
		if permissionMap, err = GetGamePermissionMap(arg.CID); err != nil {
			Ret500(c, 0, err)
			return
		}
	}

	responseList := make([]gin.H, 0, len(gamelist))

	for _, gi := range gamelist {
		gameID := util.ToID(gi.Prov + "/" + gi.Name)

		responseList = append(responseList, gin.H{
			"game_id": gameID,
			"prov":    gi.Prov,
			"name":    gi.Name,
			"lnum":    gi.LNum,
			"date":    gi.Date,
			"gt":      gi.GT,
			"gp":      gi.GP,
			"sx":      gi.SX,
			"sy":      gi.SY,
			"sn":      gi.SN,
			"ln":      gi.LN,
			"wn":      gi.WN,
			"bn":      gi.BN,
			"rtp":     gi.RTP,
			"enabled": permissionMap[gameID],
		})
	}

	ret.List = responseList
	ret.AlgNum = len(alg)
	ret.PrvNum = len(prov)

	RetOk(c, ret)
}
var (
	SpinBuf util.SqlBuf[Spinlog]
	MultBuf util.SqlBuf[Multlog]
	BankBat = map[uint64]*SqlBank{}
	JoinBuf = SqlStory{}
)

func ApiGameNew(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
		UID     uint64   `json:"uid" yaml:"uid" xml:"uid,attr" form:"uid" binding:"required"`
		Alias   string   `json:"alias" yaml:"alias" xml:"alias" form:"alias" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr"`
		Game    any      `json:"game" yaml:"game" xml:"game"`
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
	if (al&ALmember == 0) || (admin != user && al&ALdealer == 0) {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	var scene *Scene
	var alias = util.ToID(arg.Alias)
	var maker, has = game.GameFactory[alias]
	if !has {
		Ret400(c, 0, ErrNoAliase)
		return
	}

	var anygame = maker()
	var gid = StoryCounter.Inc()
	scene = &Scene{
		Story: Story{
			GID:   gid,
			CID:   arg.CID,
			UID:   arg.UID,
			Alias: alias,
		},
		Game: anygame,
	}

	if config.Cfg.ClubInsertBuffer > 1 {
		go JoinBuf.Join(XormStorage, &scene.Story)
	} else if err = JoinBuf.Join(XormStorage, &scene.Story); err != nil {
		Ret500(c, 0, err)
		return
	}

	Scenes.Set(scene.GID, scene)

	ret.GID = scene.GID
	ret.Game = scene.Game
	ret.Wallet = user.GetWallet(arg.CID)

	RetOk(c, ret)
}

func ApiGameJoin(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
		UID     uint64   `json:"uid" yaml:"uid" xml:"uid,attr" form:"uid" binding:"required"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr"`
		Game    any      `json:"game" yaml:"game" xml:"game"`
		Wallet  float64  `json:"wallet" yaml:"wallet" xml:"wallet"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	var user *User
	if user, ok = Users.Get(arg.UID); !ok {
		Ret404(c, 0, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, arg.CID)
	if (al&ALmember == 0) || (admin != user && al&ALdealer == 0) {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, 0, err)
		return
	}

	ret.GID = scene.GID
	ret.Game = scene.Game
	ret.Wallet = user.GetWallet(arg.CID)

	RetOk(c, ret)
}

func ApiGameInfo(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr"`
		Alias   string   `json:"alias" yaml:"alias" xml:"alias"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr"`
		UID     uint64   `json:"uid" yaml:"uid" xml:"uid,attr"`
		SID     uint64   `json:"sid" yaml:"sid" xml:"sid,attr"`
		Game    any      `json:"game" yaml:"game" xml:"game"`
		Wallet  float64  `json:"wallet" yaml:"wallet" xml:"wallet"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, 0, err)
		return
	}

	var user *User
	if user, ok = Users.Get(scene.UID); !ok {
		Ret500(c, 0, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin != user && al&ALdealer == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	var props *Props
	if props, ok = user.props.Get(scene.CID); !ok {
		Ret500(c, 0, ErrNoProps)
		return
	}

	ret.GID = arg.GID
	ret.Alias = scene.Alias
	ret.CID = scene.CID
	ret.UID = scene.UID
	ret.SID = scene.SID
	ret.Game = scene.Game
	ret.Wallet = props.Wallet

	RetOk(c, ret)
}

func ApiGameRtpGet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		MRTP    float64  `json:"mrtp" yaml:"mrtp" xml:"mrtp"`
		RTP     float64  `json:"rtp" yaml:"rtp" xml:"rtp"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, 0, err)
		return
	}

	var gi *game.GameInfo
	if gi, ok = game.InfoMap[scene.Alias]; !ok {
		Ret500(c, 0, ErrNoAliase)
		return
	}

	var club *Club
	if club, ok = Clubs.Get(scene.CID); !ok {
		Ret500(c, 0, ErrNoClub)
		return
	}

	var user *User
	if user, ok = Users.Get(scene.UID); !ok {
		Ret500(c, 0, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin != user && al&ALdealer == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	ret.MRTP = GetRTP(user, club)
	ret.RTP = gi.FindClosest(ret.MRTP)

	RetOk(c, ret)
}
