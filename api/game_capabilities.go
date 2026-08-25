package api

import (
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/MyTeleProject2026/Slotopol-server/game"
    "github.com/MyTeleProject2026/Slotopol-server/util"
)

// ApiGameCapabilities exposes renderer-neutral metadata for a single Slotopol game.
// It intentionally contains no provider artwork, audio, or proprietary client assets.
// The game server remains authoritative for mathematics and spin results.
func ApiGameCapabilities(c *gin.Context) {
    requested := strings.TrimSpace(c.Param("id"))
    if requested == "" {
        Ret400(c, 0, ErrNoAliase)
        return
    }

    requested = util.ToID(requested)
    var found *game.GameInfo
    for _, gi := range game.InfoMap {
        if util.ToID(gi.Prov+"/"+gi.Name) == requested {
            found = gi
            break
        }
    }

    if found == nil {
        Ret404(c, 0, ErrNoAliase)
        return
    }

    RetOk(c, gin.H{
        "success": true,
        "game": gin.H{
            "game_id": util.ToID(found.Prov + "/" + found.Name),
            "provider": found.Prov,
            "name": found.Name,
            "date": found.Date,
            "game_type": found.GT,
            "game_provider_id": found.GP,
            "reels": found.SX,
            "rows": found.SY,
            "symbols": found.SN,
            "lines": found.LN,
            "ways": found.WN,
            "bet_options": found.BN,
            "max_lines": found.LNum,
            "rtp": found.RTP,
            "server_authoritative": true,
            "features": gin.H{
                "spin": true,
                "bet": true,
                "line_selection": found.LN > 0,
                "ways": found.WN > 0,
                "double_up": true,
                "collect": true,
                "free_spins": true,
                "cascades": false,
                "progressive_jackpot": false,
            },
        },
    })
}
