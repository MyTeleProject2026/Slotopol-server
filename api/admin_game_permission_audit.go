package api

import "github.com/gin-gonic/gin"

// ApiAdminGamePermissionSetAudited preserves the existing game-permission API
// while ensuring every privileged mutation is recorded in the server audit log.
func ApiAdminGamePermissionSetAudited(c *gin.Context) {
	ApiGamePermissionSet(c)
	recordAdminAudit(c, 0, "game.permission.set", "club_game_permissions", "game permission mutation")
}

// ApiAdminGamePermissionsBulkSetAudited preserves the existing bulk API while
// recording the privileged operation as one auditable change.
func ApiAdminGamePermissionsBulkSetAudited(c *gin.Context) {
	ApiGamePermissionsBulkSet(c)
	recordAdminAudit(c, 0, "game.permission.bulk_set", "club_game_permissions", "bulk game permission mutation")
}
