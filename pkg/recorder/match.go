package recorder

import (
	"regexp"

	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
)

type MatchConfig struct {
	User  *regexp.Regexp
	Group *regexp.Regexp
	Extra map[string]*regexp.Regexp
}

func MustNewMatchConfig(user, group string, info map[string]string) (cfg *MatchConfig) {
	cfg = &MatchConfig{}

	if user != "" {
		cfg.User = regexp.MustCompile(user)
	}
	if group != "" {
		cfg.Group = regexp.MustCompile(group)
	}
	if info != nil {
		cfg.Extra = make(map[string]*regexp.Regexp)
		for k, v := range info {
			cfg.Extra[k] = regexp.MustCompile(v)
		}
	}
	return
}

func (m *MatchConfig) Match(req *admissionv1.AdmissionRequest) bool {
	return m.matchUser(req.UserInfo.Username) &&
		m.matchGroup(req.UserInfo.Groups) &&
		m.matchExtra(req.UserInfo.Extra)
}

func (m *MatchConfig) matchUser(user string) bool {
	if m.User == nil {
		return true
	}
	return m.User.MatchString(user)
}

func (m *MatchConfig) matchGroup(groups []string) bool {
	if m.Group == nil {
		return true
	}
	for _, group := range groups {
		if m.Group.MatchString(group) {
			return true
		}
	}
	return false
}

func (m *MatchConfig) matchExtra(extraMap map[string]authenticationv1.ExtraValue) bool {
	if m.Extra == nil {
		return true
	}
	// one of the extra []string must match
	// but all present keys must be there
	for k, v := range m.Extra {
		extraValues, ok := extraMap[k]
		if !ok {
			return false
		}
		var matched bool
		for _, extraValue := range extraValues {
			if v.MatchString(extraValue) {
				matched = true
			}
		}
		if !matched {
			return false
		}
	}
	return true
}
