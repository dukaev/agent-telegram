// Package replytarget builds MTProto reply targets for messages and forum topics.
package replytarget

import (
	"agent-telegram/telegram/types"
	"github.com/gotd/td/tg"
)

// Build converts the shared thread target into Telegram's reply representation.
func Build(target types.ThreadTarget) *tg.InputReplyToMessage {
	if target.ThreadID == 0 && target.ReplyTo == 0 {
		return nil
	}

	replyID := target.ReplyTo
	if replyID == 0 {
		replyID = target.ThreadID
	}
	reply := &tg.InputReplyToMessage{ReplyToMsgID: int(replyID)}
	if target.ThreadID != 0 && target.ReplyTo != 0 {
		reply.SetTopMsgID(int(target.ThreadID))
	}
	return reply
}
