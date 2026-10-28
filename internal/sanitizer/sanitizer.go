package sanitizer

import (
	"encoding/json"

	"github.com/microcosm-cc/bluemonday"
)

var policy = bluemonday.StrictPolicy()

type SafeString string

func (s *SafeString) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*s = SafeString(policy.Sanitize(raw))
	return nil
}

func (s *SafeString) String() string {
	if s == nil {
		return ""
	}
	return string(*s)
}

func Ptr(s *SafeString) *string {
	if s == nil {
		return nil
	}
	str := string(*s)
	return &str
}
