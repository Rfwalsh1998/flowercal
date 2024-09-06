package handler

import (
	"errors"

	"m1k1o/neko/pkg/types"
	"m1k1o/neko/pkg/types/message"
)

func (h *MessageHandlerCtx) clipboardSet(session types.Session, payload *message.ClipboardData) error {
	if !session.Profile().CanAccessClipboard {
		return errors.New("cannot access clipboard")
	}

	if !session.IsHost() {
		return errors.New("is not the host")
	}

	return h.desktop.ClipboardSetText(types.ClipboardText{
		Text: payload.Text,
		// TODO: Send HTML?
	})
}