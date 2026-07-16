package replytarget

import (
	"testing"

	"agent-telegram/telegram/types"
)

func TestBuild(t *testing.T) {
	tests := []struct {
		name               string
		in                 types.ThreadTarget
		wantReply, wantTop int
		wantNil            bool
	}{
		{name: "empty", wantNil: true},
		{name: "thread root", in: types.ThreadTarget{ThreadID: 77}, wantReply: 77},
		{name: "ordinary reply", in: types.ThreadTarget{ReplyTo: 88}, wantReply: 88},
		{name: "nested topic reply", in: types.ThreadTarget{ThreadID: 77, ReplyTo: 88}, wantReply: 88, wantTop: 77},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Build(tt.in)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("Build() = %#v, want nil", got)
				}
				return
			}
			if got == nil || got.ReplyToMsgID != tt.wantReply || got.TopMsgID != tt.wantTop {
				t.Fatalf("Build() = %#v, want reply=%d top=%d", got, tt.wantReply, tt.wantTop)
			}
		})
	}
}
