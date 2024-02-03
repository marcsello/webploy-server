package hooks

import "webploy-server/config"

type HookID string

const (
	HookPreCreate  HookID = "pre_create"
	HookPreFinish  HookID = "pre_finish"
	HookPostFinish HookID = "post_finish"
	HookPreLive    HookID = "pre_live"
	HookPostLive   HookID = "post_live"
)

func getHookPathFromConfig(hook HookID, config config.HooksConfig) string {
	switch hook {
	case HookPreCreate:
		return config.PreCreate
	case HookPreFinish:
		return config.PreFinish
	case HookPostFinish:
		return config.PostFinish
	case HookPreLive:
		return config.PreLive
	case HookPostLive:
		return config.PostLive
	}
	panic("invalid hook")
}
