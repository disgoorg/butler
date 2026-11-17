package dbtype

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/disgoorg/disgo/discord"
)

// OAuthScopes scans the jsonb scopes column into Discord OAuth scopes.
type OAuthScopes []discord.OAuth2Scope

func (s *OAuthScopes) Scan(src any) error {
	if src == nil {
		*s = nil
		return nil
	}
	var data []byte
	switch value := src.(type) {
	case []byte:
		data = value
	case string:
		data = []byte(value)
	default:
		return fmt.Errorf("scopes: unsupported type %T", src)
	}
	if len(data) == 0 {
		*s = nil
		return nil
	}
	return json.Unmarshal(data, s)
}

func (s OAuthScopes) Value() (driver.Value, error) {
	if s == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(s)
}
