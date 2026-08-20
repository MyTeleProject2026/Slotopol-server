package keno

import (
	"context"
	"errors"
	"math/rand/v2"

	"github.com/MyTeleProject2026/Slotopol-server/util"
)

const InitRTP = 95.0

// Bitset is an 80-cell selection bitset backed by the project's 128-bit utility.
type Bitset = util.Bitset128

// KS is the state of one Keno cell.
type KS byte

const (
	KSempty  KS = 0
	KSsel    KS = 0x1
	KShit    KS = 0x2
	KSselhit KS = KSsel | KShit
)

// Grid is the 80-number Keno board.
type Grid [80]KS

// Wins describes the result of a Keno spin.
type Wins struct {
	Sel int     `json:"sel" yaml:"sel" xml:"sel,attr"`
	Num int     `json:"num" yaml:"num" xml:"num,attr"`
	Pay float64 `json:"pay" yaml:"pay" xml:"pay,attr"`
}

// Keno80 is the common 80-number Keno state.
type Keno80 struct {
	Grid Grid    `json:"grid" yaml:"grid" xml:"grid"`
	Bet  float64 `json:"bet" yaml:"bet" xml:"bet"`
	Sel  Bitset  `json:"sel" yaml:"sel" xml:"sel"`
}

var (
	ErrBadBet = errors.New("wrong bet value")
	ErrBadSel = errors.New("wrong Keno selection")
)

// KenoGame is the common Keno game interface.
type KenoGame interface {
	Scanner(*Wins) error
	Spin(float64)
	GetBet() float64
	SetBet(float64) error
	GetSel() Bitset
	SetSel(Bitset) error
}

// GetBet returns the current bet.
func (g *Keno80) GetBet() float64 { return g.Bet }

// SetBet changes the bet.
func (g *Keno80) SetBet(bet float64) error {
	if bet <= 0 {
		return ErrBadBet
	}
	g.Bet = bet
	return nil
}

// GetSel returns the current selected numbers.
func (g *Keno80) GetSel() Bitset { return g.Sel }

// CheckSel validates and stores a selection against a paytable.
func (g *Keno80) CheckSel(sel Bitset, kp *Paytable) error {
	n := sel.Num()
	if n < 1 || n > 10 || !kp.HasSel(n) {
		return ErrBadSel
	}
	g.Sel = sel
	return nil
}

// Spin generates 20 unique hits on the 80-number board.
func (g *Keno80) Spin(_ float64) {
	g.Grid = Grid{}
	for i := range 80 {
		if g.Sel.Is(i) {
			g.Grid[i] = KSsel
		}
	}

	perm := rand.Perm(80)
	for _, n := range perm[:20] {
		g.Grid[n] |= KShit
	}
}

// Paytable contains payouts indexed by selected spots and hits.
type Paytable [11][11]float64

func (kp *Paytable) Pay(sel, hit int) float64 {
	if sel < 0 || sel >= len(kp) || hit < 0 || hit >= len(kp[sel]) {
		return 0
	}
	return kp[sel][hit]
}

func (kp *Paytable) HasSel(sel int) bool {
	if sel < 0 || sel >= len(kp) {
		return false
	}
	for _, pay := range kp[sel] {
		if pay > 0 {
			return true
		}
	}
	return false
}

func (kp *Paytable) Scanner(grid *Grid, wins *Wins, bet float64) error {
	wins.Sel = 0
	wins.Num = 0
	for i := range 80 {
		if grid[i]&KSsel != 0 {
			wins.Sel++
			if grid[i]&KShit != 0 {
				wins.Num++
			}
		}
	}
	wins.Pay = kp.Pay(wins.Sel, wins.Num) * bet
	return nil
}

// Combin returns n choose r.
func Combin(n, r int) float64 {
	if r < 0 || r > n {
		return 0
	}
	if r > n-r {
		r = n - r
	}
	result := 1.0
	for i := 1; i <= r; i++ {
		result *= float64(n-r+i) / float64(i)
	}
	return result
}

// Prob returns the probability of exactly r hits among n selected numbers
// when 20 numbers are drawn from an 80-number Keno board.
func Prob(n, r int) float64 {
	den := Combin(80, 20)
	if den == 0 {
		return 0
	}
	return Combin(n, r) * Combin(80-n, 20-r) / den
}

// CalcStat calculates the theoretical RTP of a paytable.
func (kp *Paytable) CalcStat(ctx context.Context) float64 {
	var rtp float64
	for sel := 1; sel <= 10; sel++ {
		if ctx != nil {
			select {
			case <-ctx.Done():
				return rtp
			default:
			}
		}
		if !kp.HasSel(sel) {
			continue
		}
		for hit := 0; hit <= sel; hit++ {
			rtp += Prob(sel, hit) * kp.Pay(sel, hit)
		}
	}
	// The statistic above is a sum over selections. The paytable's normal
	// per-bet RTP is the average of the configured selection counts.
	var count float64
	for sel := 1; sel <= 10; sel++ {
		if kp.HasSel(sel) {
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return rtp / count
}
