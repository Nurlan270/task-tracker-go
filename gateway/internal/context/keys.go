package context

type contextKey struct {
	name string
}

var (
	UserCtxKey = &contextKey{name: "user"}
)
