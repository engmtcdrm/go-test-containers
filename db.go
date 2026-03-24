package main

import (
	"context"

	"gorm.io/gorm"
)

type contextKey string

const dbKey contextKey = "gormDB"

// contextWithDatabase returns a new context with the *gorm.DB attached.
func contextWithDatabase(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, dbKey, db)
}

// databaseFromContext retrieves the *gorm.DB from the context.
func databaseFromContext(ctx context.Context) *gorm.DB {
	return ctx.Value(dbKey).(*gorm.DB)
}

// DBTableProduct is the actual table structure for products in the database
type DBTableProduct struct {
	gorm.Model
	Name        string  `gorm:"column:name"`
	Price       float64 `gorm:"column:price"`
	RequiredCol string  `gorm:"column:required_col;not null"`
}

func (DBTableProduct) TableName() string {
	return "products"
}
