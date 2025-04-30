package main

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"

	"realy.lol/bech32encoding"
	"realy.lol/chk"
	"realy.lol/ec/secp256k1"
)

func main() {
	b := core.NewBody().SetTitle("bunk - nostr bunker signer")
	b.Styler(func(s *styles.Style) {
		s.Min = units.XY{
			X: units.Dp(480),
			Y: units.Dp(640),
		}
	})
	// b.Styler(func(s *styles.Style) {
	// 	s.CenterAll()
	// })
	pgs := core.NewPages(b)
	pgs.AddPage("start", StartInput)

	b.RunMainWindow()
}

func StartInput(pg *core.Pages) {
	pg.Styler(func(s *styles.Style) {
		s.CenterAll()
	})
	core.NewText().SetType(core.TextHeadlineMedium).SetText("welcome")
	core.NewText(pg).SetType(core.TextBodyMedium).SetText("enter your secret key (nsec or hex):")
	tf := core.NewTextField(pg)
	tf.Styler(func(s *styles.Style) {
		s.Min = units.XY{
			X: units.Dp(400),
			Y: units.Dp(32),
		}
	})
	next := core.NewButton(pg).SetText("next").SetEnabled(false)
	next.OnClick(func(e events.Event) {

	})
	_ = next
	tf.On(events.Input, func(e events.Event) {
		t := tf.Text()
		var err error
		var sk []byte
		if sk, err = bech32encoding.NsecToBytes([]byte(t)); chk.E(err) {
			return
		}
		if len(sk) == secp256k1.SecKeyBytesLen {
			next.SetEnabled(true)
		}
	})
}
