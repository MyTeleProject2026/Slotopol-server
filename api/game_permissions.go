package api

import (
	"encoding/xml"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/MyTeleProject2026/Slotopol-server/util"
)

var gamePermissionOnce sync.Once
var gamePermissionInitErr error

// ClubGamePermission intentionally keeps the existing database column
// "game_alias" for backward compatibility.
//
// The application/API uses GameID as the canonical game identifier.
// Example:
//   agt/doubleice
//
// Existing database table:
//   club_game_permissions
//     club_id
//     game_alias
//     enabled
type ClubGamePermission struct {
	ClubID uint64 `xorm:"pk 'club_id'" json:"club_id" yaml:"club_id" xml:"club_id"`

	GameID string `xorm:"pk 'game_alias'" json:"game_id" yaml:"game_id" xml:"game_id"`

	Enabled bool `xorm:"notnull default false" json:"enabled" yaml:"enabled" xml:"enabled"`
}

func (ClubGamePermission) TableName() string {
	return "club_game_permissions"
}

// EnsureGamePermissionStore creates/synchronizes the permission table once.
//
// It also migrates legacy display aliases such as:
//
//   AGT / Double Ice
//
// into canonical game IDs:
//
//   agt/doubleice
//
// The old rows are replaced with canonical rows.
func EnsureGamePermissionStore() error {
	gamePermissionOnce.Do(func() {
		if XormStorage == nil {
			gamePermissionInitErr = fmt.Errorf("game permission storage is not initialized")
			return
		}

		if err := XormStorage.Sync(new(ClubGamePermission)); err != nil {
			gamePermissionInitErr = err
			return
		}

		var permissions []ClubGamePermission
		if err := XormStorage.Find(&permissions); err != nil {
			gamePermissionInitErr = err
			return
		}

		for _, permission := range permissions {
			canonicalID := util.ToID(permission.GameID)

			if canonicalID == "" || canonicalID == permission.GameID {
				continue
			}

			canonical := ClubGamePermission{
				ClubID:  permission.ClubID,
				GameID:  canonicalID,
				Enabled: permission.Enabled,
			}

			if _, err := XormStorage.
				Where("club_id=? AND game_alias=?", canonical.ClubID, canonical.GameID).
				Delete(&ClubGamePermission{}); err != nil {
				gamePermissionInitErr = err
				return
			}

			if _, err := XormStorage.InsertOne(&canonical); err != nil {
				gamePermissionInitErr = err
				return
			}

			if _, err := XormStorage.
				Where("club_id=? AND game_alias=?", permission.ClubID, permission.GameID).
				Delete(&ClubGamePermission{}); err != nil {
				gamePermissionInitErr = err
				return
			}
		}
	})

	return gamePermissionInitErr
}

func NormalizeGameID(gameID string) string {
	return util.ToID(gameID)
}

// IsGameEnabledForClub is the authoritative permission check.
//
// IMPORTANT:
// No permission record means the game is DISABLED.
//
// This gives the platform administrator complete control over
// which games each club can use.
func IsGameEnabledForClub(clubID uint64, gameID string) (bool, error) {
	if clubID == 0 {
		return false, nil
	}

	if err := EnsureGamePermissionStore(); err != nil {
		return false, err
	}

	gameID = NormalizeGameID(gameID)

	if gameID == "" {
		return false, nil
	}

	var permission ClubGamePermission

	has, err := XormStorage.
		Where(
			"club_id=? AND game_alias=? AND enabled=?",
			clubID,
			gameID,
			true,
		).
		Get(&permission)

	if err != nil {
		return false, err
	}

	return has, nil
}

// GetGamePermissionMap returns:
//
// map[canonicalGameID]enabled
func GetGamePermissionMap(clubID uint64) (map[string]bool, error) {
	result := make(map[string]bool)

	if clubID == 0 {
		return result, nil
	}

	if err := EnsureGamePermissionStore(); err != nil {
		return nil, err
	}

	var permissions []ClubGamePermission

	if err := XormStorage.
		Where("club_id=?", clubID).
		Find(&permissions); err != nil {
		return nil, err
	}

	for _, permission := range permissions {
		result[NormalizeGameID(permission.GameID)] = permission.Enabled
	}

	return result, nil
}

// ApiGamePermissionSet changes one game's permission for one club.
//
// Requires global administrator permission.
//
// Request:
//
// {
//   "club_id": 5,
//   "game_id": "agt/doubleice",
//   "enabled": true
// }
func ApiGamePermissionSet(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`

		ClubID uint64 `json:"club_id" yaml:"club_id" xml:"club_id" binding:"required"`

		GameID string `json:"game_id" yaml:"game_id" xml:"game_id" binding:"required"`

		Enabled bool `json:"enabled" yaml:"enabled" xml:"enabled"`
	}

	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	// Platform/global administrator check.
	//
	// This allows Slotopol Admin to manage permissions for OTHER clubs,
	// instead of requiring the admin user to belong to every target club.
	_, accessLevel := GetAdmin(c, 0)

	if accessLevel&ALadmin == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	if _, exists := Clubs.Get(arg.ClubID); !exists {
		Ret404(c, 0, ErrNoClub)
		return
	}

	gameID := NormalizeGameID(arg.GameID)

	if gameID == "" {
		Ret400(c, 0, ErrNoAliase)
		return
	}

	if err := EnsureGamePermissionStore(); err != nil {
		Ret500(c, 0, err)
		return
	}

	permission := ClubGamePermission{
		ClubID:  arg.ClubID,
		GameID:  gameID,
		Enabled: arg.Enabled,
	}

	updated, err := XormStorage.
		Where(
			"club_id=? AND game_alias=?",
			arg.ClubID,
			gameID,
		).
		Cols("enabled").
		Update(&permission)

	if err != nil {
		Ret500(c, 0, err)
		return
	}

	if updated == 0 {
		if _, err = XormStorage.InsertOne(&permission); err != nil {
			Ret500(c, 0, err)
			return
		}
	}

	Ret204(c)
}

// ApiGamePermissionsBulkSet changes multiple games in one request.
//
// Request:
//
// {
//   "club_id": 5,
//   "game_ids": [
//     "agt/doubleice",
//     "netent/gonzosquest"
//   ],
//   "enabled": true
// }
func ApiGamePermissionsBulkSet(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`

		ClubID uint64 `json:"club_id" yaml:"club_id" xml:"club_id" binding:"required"`

		GameIDs []string `json:"game_ids" yaml:"game_ids" xml:"game_ids>game_id" binding:"required,min=1"`

		Enabled bool `json:"enabled" yaml:"enabled" xml:"enabled"`
	}

	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, 0, err)
		return
	}

	_, accessLevel := GetAdmin(c, 0)

	if accessLevel&ALadmin == 0 {
		Ret403(c, 0, ErrNoAccess)
		return
	}

	if _, exists := Clubs.Get(arg.ClubID); !exists {
		Ret404(c, 0, ErrNoClub)
		return
	}

	if err := EnsureGamePermissionStore(); err != nil {
		Ret500(c, 0, err)
		return
	}

	session := XormStorage.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		Ret500(c, 0, err)
		return
	}

	for _, rawGameID := range arg.GameIDs {
		gameID := NormalizeGameID(rawGameID)

		if gameID == "" {
			_ = session.Rollback()
			Ret400(c, 0, ErrNoAliase)
			return
		}

		permission := ClubGamePermission{
			ClubID:  arg.ClubID,
			GameID:  gameID,
			Enabled: arg.Enabled,
		}

		updated, err := session.
			Where(
				"club_id=? AND game_alias=?",
				arg.ClubID,
				gameID,
			).
			Cols("enabled").
			Update(&permission)

		if err != nil {
			_ = session.Rollback()
			Ret500(c, 0, err)
			return
		}

		if updated == 0 {
			if _, err := session.InsertOne(&permission); err != nil {
				_ = session.Rollback()
				Ret500(c, 0, err)
				return
			}
		}
	}

	if err := session.Commit(); err != nil {
		Ret500(c, 0, err)
		return
	}

	Ret204(c)
}
