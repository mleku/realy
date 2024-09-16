package relay

const AUTH_CONTEXT_KEY = iota

func GetAuthStatus(ctx Ctx) (pubkey S, ok bool) {
	authedPubkey := ctx.Value(AUTH_CONTEXT_KEY)
	if authedPubkey == nil {
		return "", false
	}
	return authedPubkey.(S), true
}
