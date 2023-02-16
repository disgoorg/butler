package butler

import (
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/oauth2"

	"github.com/disgoorg/disgo-butler/db"
)

var _ oauth2.Session = (*Session)(nil)

func newSession(contributor db.Contributor) Session {
	return Session{
		accessToken:  contributor.AccessToken,
		refreshToken: contributor.RefreshToken,
		scopes:       contributor.Scopes,
		tokenType:    discord.TokenType(contributor.TokenType),
		expiration:   contributor.Expiration,
	}
}

type Session struct {
	accessToken  string
	refreshToken string
	scopes       []discord.OAuth2Scope
	tokenType    discord.TokenType
	expiration   time.Time
}

func (s Session) AccessToken() string {
	return s.accessToken
}

func (s Session) RefreshToken() string {
	return s.refreshToken
}

func (s Session) Scopes() []discord.OAuth2Scope {
	return s.scopes
}

func (s Session) TokenType() discord.TokenType {
	return s.tokenType
}

func (s Session) Expiration() time.Time {
	return s.expiration
}

func (s Session) Webhook() *discord.IncomingWebhook {
	return nil
}
