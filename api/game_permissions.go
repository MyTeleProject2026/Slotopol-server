package api

import (
	"encoding/xml"

	"github.com/gin-gonic/gin"
)

type ClubGamePermission struct {
	ClubID    uint64 `xorm:"club_id" json:"club_id" yaml:"club_id" xml:"club_id"`
	GameAlias string `xorm:"game_alias" json:"game_alias" yaml:"game_alias" xml:"game_alias"`
	Enabled   bool   `xorm:"enabled" json:"enabled" yaml:"enabled" xml:"enabled"`
}

func ApiGamePermissionSet(c *gin.Context) {
	var arg struct {
		XMLName   xml.Name `json:"-" yaml:"-" xml:"arg"`
		ClubID    uint64   `json:"club_id" binding:"required"`
		GameAlias string   `json:"game_alias" binding:"required"`
		Enabled   bool     `json:"enabled"`
	}
	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}
	admin, al := MustAdmin(c, arg.ClubID)
	if admin == nil || al&ALadmin == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	// ✅ Auto-create table if missing (so you never need to run SQL!)
	_, _ = XormStorage.Exec(`CREATE TABLE IF NOT EXISTS club_game_permissions (
		club_id BIGINT UNSIGNED NOT NULL,
		game_alias VARCHAR(128) NOT NULL,
		enabled BOOLEAN NOT NULL DEFAULT TRUE,
		PRIMARY KEY (club_id, game_alias)
	)`)

	perm := ClubGamePermission{
		ClubID:    arg.ClubID,
		GameAlias: arg.GameAlias,
		Enabled:   arg.Enabled,
	}

	_, err := XormStorage.InsertOne(&perm)
	if err != nil {
		if _, err2 := XormStorage.Where("club_id=? AND game_alias=?", arg.ClubID, arg.GameAlias).
			Cols("enabled").Update(&perm); err2 != nil {
			Ret500(c, 0, err2)
			return
		}
	}
	Ret204(c)
}
