package api

import (
	"errors"

	"github.com/gin-gonic/gin"
)

var ErrGameDisabled = errors.New("game is disabled for this club")

// RequireGameEnabledForClub is the central authorization gate for gameplay.
//
// Every endpoint that can create or advance gameplay should use this.
func RequireGameEnabledForClub(
	c *gin.Context,
	clubID uint64,
	gameID string,
) bool {
	enabled, err := IsGameEnabledForClub(clubID, gameID)

	if err != nil {
		Ret500(c, 0, err)
		return false
	}

	if !enabled {
		Ret403(c, 0, ErrGameDisabled)
		return false
	}

	return true
}

// RequireSceneGameEnabled verifies permission for an existing game session.
func RequireSceneGameEnabled(c *gin.Context, scene *Scene) bool {
	if scene == nil {
		Ret404(c, 0, errors.New("game scene does not exist"))
		return false
	}

	return RequireGameEnabledForClub(
		c,
		scene.CID,
		scene.Alias,
	)
}
