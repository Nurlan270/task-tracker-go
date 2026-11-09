package context

type contextKey struct {
	name string
}

var (
	UserCtxKey         = &contextKey{name: "user"}
	RefreshTokenCtxKey = &contextKey{name: "refresh_token"}
)
