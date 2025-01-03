package chat

import (
	"image"
	"realy.lol/images"
	"bytes"
)

type ChatButton struct {
	Username, Message st
	Avatar            *image.Image
}

func (cb *ChatButton) Init(img by, username, msg st) *ChatButton {
	if img == nil {
		// set default goofy rly owl image
		buf := bytes.NewBuffer(images.Realy)
		var i image.Image
		var err er
		if i, _, err = image.Decode(buf); chk.E(err) {
			return nil
		}
		cb.Avatar = &i
	}
	cb.Username = username
	cb.Message = msg
	return cb
}

func (cb *ChatButton) Layout(g Gx) (d Dim) {

	return
}
