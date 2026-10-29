package context

import (
	"context"

	"gorm.io/gorm"
)

var txCtxKey = contextKey{name: "tx"}

func NewTXCtx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txCtxKey, tx)
}

func TXFromCtx(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(txCtxKey).(*gorm.DB)
	return tx, ok
}

// GetDB returns TX within provided context,
// if it does not find any it will return provided db argument.
func GetDB(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := TXFromCtx(ctx); ok {
		return tx
	}

	return db
}
