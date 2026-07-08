package user

import (
	"context"
	"errors"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"

	"agent-telegram/telegram/client"
	"agent-telegram/telegram/types"
)

func TestClientMethodsRequireInitialization(t *testing.T) {
	c := NewClient(nil)
	ctx := context.Background()
	check := func(name string, err error) {
		t.Helper()
		if !errors.Is(err, client.ErrNotInitialized) {
			t.Fatalf("%s err = %v, want ErrNotInitialized", name, err)
		}
	}

	_, err := c.GetContacts(ctx, types.GetContactsParams{})
	check("GetContacts", err)
	_, err = c.AddContact(ctx, types.AddContactParams{})
	check("AddContact", err)
	_, err = c.DeleteContact(ctx, types.DeleteContactParams{})
	check("DeleteContact", err)
	_, err = c.UpdateProfile(ctx, types.UpdateProfileParams{})
	check("UpdateProfile", err)
	_, err = c.UpdateAvatar(ctx, types.UpdateAvatarParams{})
	check("UpdateAvatar", err)
	_, err = c.BlockPeer(ctx, types.BlockPeerParams{})
	check("BlockPeer", err)
	_, err = c.UnblockPeer(ctx, types.UnblockPeerParams{})
	check("UnblockPeer", err)
	_, err = c.GetPrivacy(ctx, types.GetPrivacyParams{})
	check("GetPrivacy", err)
	_, err = c.SetPrivacy(ctx, types.SetPrivacyParams{})
	check("SetPrivacy", err)
	_, err = c.GetUserInfo(ctx, types.GetUserInfoParams{Username: "@user"})
	check("GetUserInfo", err)
}

func TestContactHelpers(t *testing.T) {
	if got := trimUsernamePrefix("@ada"); got != "ada" {
		t.Fatalf("trim = %q", got)
	}
	if got := trimUsernamePrefix("ada"); got != "ada" {
		t.Fatalf("trim bare = %q", got)
	}
	contact := convertUserToContact(tgUser{
		ID:         42,
		AccessHash: 10,
		FirstName:  "Ada",
		LastName:   "Lovelace",
		Username:   "ada",
		Phone:      "+1",
		Bot:        true,
		Verified:   true,
	}.User())
	if contact.ID != 42 || contact.FirstName != "Ada" || contact.LastName != "Lovelace" ||
		contact.Peer != "@ada" || !contact.Bot || !contact.Verified {
		t.Fatalf("contact = %+v", contact)
	}
	if !matchesQuery(contact, "lovelace") || !matchesQuery(contact, "ada") || !matchesQuery(contact, "+1") {
		t.Fatalf("contact should match query: %+v", contact)
	}
	if matchesQuery(contact, "missing") {
		t.Fatalf("contact should not match missing query")
	}
}

func TestGetContactsWithFakeAPI(t *testing.T) {
	c := NewClient(nil)
	c.SetAPI(tg.NewClient(tgmock.Invoker(func(input bin.Encoder) (bin.Encoder, error) {
		switch input.(type) {
		case *tg.ContactsGetContactsRequest:
			return &tg.ContactsContacts{Users: []tg.UserClass{
				&tg.User{ID: 1, FirstName: "Ada", LastName: "Lovelace", Username: "ada", Phone: "+1"},
				&tg.User{ID: 2, FirstName: "Grace", Username: "grace", Phone: "+2"},
			}}, nil
		default:
			t.Fatalf("unexpected request %T", input)
			return nil, nil
		}
	})))

	result, err := c.GetContacts(context.Background(), types.GetContactsParams{Query: "ada", Limit: 0})
	if err != nil {
		t.Fatal(err)
	}
	if result.Count != 1 || result.Contacts[0].Username != "ada" || result.Query != "ada" {
		t.Fatalf("contacts = %+v", result)
	}
}

type tgUser struct {
	ID         int64
	AccessHash int64
	FirstName  string
	LastName   string
	Username   string
	Phone      string
	Bot        bool
	Verified   bool
}

func (u tgUser) User() *tg.User {
	return &tg.User{
		ID:         u.ID,
		AccessHash: u.AccessHash,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Username:   u.Username,
		Phone:      u.Phone,
		Bot:        u.Bot,
		Verified:   u.Verified,
	}
}
