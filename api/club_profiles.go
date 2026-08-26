package api

import (
    "encoding/xml"
    "fmt"
    "strings"

    "github.com/gin-gonic/gin"
)

// ClubProfile stores provider-side country/currency defaults for a club.
type ClubProfile struct {
    ClubID      uint64 `xorm:"pk 'club_id'" json:"club_id"`
    CountryCode string `xorm:"varchar(8) notnull default ''" json:"country_code"`
    Currency    string `xorm:"varchar(8) notnull default ''" json:"currency"`
}
func (ClubProfile) TableName() string { return "club_profiles" }

// CountryGameProfile supplies safe default bet limits when a club does not override them.
type CountryGameProfile struct {
    ID          uint64  `xorm:"pk autoincr" json:"id"`
    CountryCode string  `xorm:"varchar(8) notnull" json:"country_code"`
    Currency    string  `xorm:"varchar(8) notnull" json:"currency"`
    MinBet      float64 `xorm:"notnull default 0" json:"min_bet"`
    MaxBet      float64 `xorm:"notnull default 0" json:"max_bet"`
    BetStep     float64 `xorm:"notnull default 0" json:"bet_step"`
    Enabled     bool    `xorm:"notnull default true" json:"enabled"`
}
func (CountryGameProfile) TableName() string { return "country_game_profiles" }

func ensureClubProfiles() error { return XormStorage.Sync(new(ClubProfile), new(CountryGameProfile)) }

func ApiClubProfileSet(c *gin.Context) {
    var arg struct { XMLName xml.Name `json:"-"`; ClubID uint64 `json:"club_id" binding:"required"`; CountryCode string `json:"country_code" binding:"required"`; Currency string `json:"currency" binding:"required"` }
    if err := c.ShouldBind(&arg); err != nil { Ret400(c, 0, err); return }
    _, al := GetAdmin(c, 0); if al&ALadmin == 0 { Ret403(c, 0, ErrNoAccess); return }
    if _, ok := Clubs.Get(arg.ClubID); !ok { Ret404(c, 0, ErrNoClub); return }
    if err := ensureClubProfiles(); err != nil { Ret500(c, 0, err); return }
    p := ClubProfile{ClubID: arg.ClubID, CountryCode: strings.ToUpper(strings.TrimSpace(arg.CountryCode)), Currency: strings.ToUpper(strings.TrimSpace(arg.Currency))}
    n, err := XormStorage.Where("club_id=?", p.ClubID).AllCols().Update(&p)
    if err != nil { Ret500(c, 0, err); return }
    if n == 0 { if _, err = XormStorage.InsertOne(&p); err != nil { Ret500(c, 0, err); return } }
    recordAdminAudit(c, p.ClubID, "club.profile.update", "club_profile", fmt.Sprintf("country=%s currency=%s", p.CountryCode, p.Currency))
    RetOk(c, p)
}

func ApiClubProfileGet(c *gin.Context) {
    var arg struct { CID uint64 `form:"cid" binding:"required"` }; if err := c.ShouldBindQuery(&arg); err != nil { Ret400(c, 0, err); return }
    if err := ensureClubProfiles(); err != nil { Ret500(c, 0, err); return }; var p ClubProfile
    has, err := XormStorage.Where("club_id=?", arg.CID).Get(&p); if err != nil { Ret500(c, 0, err); return }; if !has { Ret404(c, 0, ErrNoClub); return }; RetOk(c, p)
}

func ApiCountryGameProfileSet(c *gin.Context) {
    var p CountryGameProfile; if err := c.ShouldBind(&p); err != nil { Ret400(c, 0, err); return }; _, al := GetAdmin(c, 0); if al&ALadmin == 0 { Ret403(c, 0, ErrNoAccess); return }
    if p.MinBet < 0 || p.MaxBet < p.MinBet || p.BetStep < 0 { Ret400(c, 0, ErrNoAccess); return }; if err := ensureClubProfiles(); err != nil { Ret500(c, 0, err); return }
    p.CountryCode = strings.ToUpper(strings.TrimSpace(p.CountryCode)); p.Currency = strings.ToUpper(strings.TrimSpace(p.Currency))
    var old CountryGameProfile; has, err := XormStorage.Where("country_code=? AND currency=?", p.CountryCode, p.Currency).Get(&old); if err != nil { Ret500(c, 0, err); return }
    if has { p.ID = old.ID; _, err = XormStorage.ID(old.ID).AllCols().Update(&p) } else { _, err = XormStorage.InsertOne(&p) }; if err != nil { Ret500(c, 0, err); return }
    recordAdminAudit(c, 0, "country-game-profile.update", "country_game_profile", fmt.Sprintf("country=%s currency=%s min=%g max=%g step=%g enabled=%t", p.CountryCode, p.Currency, p.MinBet, p.MaxBet, p.BetStep, p.Enabled))
    RetOk(c, p)
}

func ApiCountryGameProfileList(c *gin.Context) { if err := ensureClubProfiles(); err != nil { Ret500(c, 0, err); return }; var rows []CountryGameProfile; if err := XormStorage.Find(&rows); err != nil { Ret500(c, 0, err); return }; RetOk(c, gin.H{"profiles": rows}) }
