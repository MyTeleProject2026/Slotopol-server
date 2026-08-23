package api

import "github.com/gin-gonic/gin"

// RequireGameEnabledForClub verifies that a game is currently enabled
// for the specified club.
//
// Permission policy:
//
//   - enabled permission record  -> game can be used
//   - disabled permission record -> game cannot be used
//   - no permission record       -> game cannot be used
//
// This makes the database permission table the authoritative source
// for determining whether a club is allowed to use a game.
//
// The function writes the HTTP error response itself and returns false
// when access should be denied.
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
		Ret403(c, 0, ErrNoAccess)
		return false
	}

	return true
}

// RequireSceneGameEnabled verifies that the game associated with an
// already-created scene is still enabled for that scene's club.
//
// This is important because a platform administrator may disable a game
// after a player has already created a game session.
//
// Therefore, every gameplay action that changes game state or money
// should verify the current permission again.
//
// The function writes the HTTP error response itself and returns false
// when access should be denied.
func RequireSceneGameEnabled(
	c *gin.Context,
	scene *Scene,
) bool {
	if scene == nil {
		Ret403(c, 0, ErrNotOpened)
		return false
	}

	return RequireGameEnabledForClub(
		c,
		scene.CID,
		scene.Alias,
	)
}
