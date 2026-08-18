package keno

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"sort"

	"github.com/MyTeleProject2026/Slotopol-server/game"
	"github.com/MyTeleProject2026/Slotopol-server/util"
)

type Paytable [11][11]float64

func (kp *Paytable) Pay(sel, hit int) float64 {
	return kp[sel][hit]
}

func (kp *Paytable) HasSel(sel int) bool {
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
		if grid[i]&KSsel > 0 {
			wins.Sel++
			if grid[i]&KShit > 0 {
				wins.Num++
			}
		}
	}
	wins.Pay = kp[wins.Sel][wins.Num] * bet
	return nil
}
