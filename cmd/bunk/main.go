package main

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/styles"
)

func main() {
	b := core.NewBody().SetTitle("bunk - nostr bunker signer")
	ts := core.NewTabs(b)
	start, _ := ts.NewTab("start")
	start.Styler(func(s *styles.Style) {
		s.CenterAll()
	})
	core.NewText(start).SetType(core.TextHeadlineMedium).SetText("welcome")
	core.NewText(start).SetType(core.TextBodyMedium).SetText("enter your secret key (nsec or hex):")
	type nsec struct {
		Secret string
	}
	n := &nsec{}
	core.NewForm(start).SetStruct(n)
	signer, _ := ts.NewTab("signer")
	signer.Styler(func(s *styles.Style) {
		s.CenterAll()
	})
	core.NewText(signer).SetType(core.TextHeadlineMedium).SetText("signer")
	config, _ := ts.NewTab("config")
	config.Styler(func(s *styles.Style) {
		s.CenterAll()
	})
	core.NewText(config).SetType(core.TextHeadlineMedium).SetText("config")
	b.RunMainWindow()
}
