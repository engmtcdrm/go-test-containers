package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupPostgres starts a Postgres container, connects via GORM, runs AutoMigrate
// for Product, and returns a context containing the *gorm.DB.
func setupPostgres(t *testing.T) context.Context {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		tcpostgres.BasicWaitStrategies(),
	)
	testcontainers.CleanupContainer(t, pgContainer)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := gorm.Open(pgdriver.New(pgdriver.Config{
		DSN: connStr,
	}), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&DBTableProduct{})
	require.NoError(t, err)

	return contextWithDatabase(ctx, db)
}

func TestPostgresWithGorm(t *testing.T) {
	ctx := setupPostgres(t)
	db := databaseFromContext(ctx)

	// Insert records
	products := []DBTableProduct{
		{Name: "Widget", Price: 9.99},
		{Name: "Gadget", Price: 24.50},
	}
	for _, p := range products {
		result := db.Create(&p)
		require.NoError(t, result.Error)
	}

	// Verify record count
	var count int64
	db.Model(&DBTableProduct{}).Count(&count)
	assert.Equal(t, int64(2), count)

	// Verify individual records
	var widget DBTableProduct
	err := db.Where("name = ?", "Widget").First(&widget).Error
	require.NoError(t, err)
	assert.Equal(t, "Widget", widget.Name)
	assert.Equal(t, 9.99, widget.Price)

	var gadget DBTableProduct
	err = db.Where("name = ?", "Gadget").First(&gadget).Error
	require.NoError(t, err)
	assert.Equal(t, "Gadget", gadget.Name)
	assert.Equal(t, 24.50, gadget.Price)

	// Verify all records via query
	var allProducts []DBTableProduct
	err = db.Order("price asc").Find(&allProducts).Error
	require.NoError(t, err)
	require.Len(t, allProducts, 2)
	assert.Equal(t, "Widget", allProducts[0].Name)
	assert.Equal(t, "Gadget", allProducts[1].Name)

	fmt.Println("All assertions passed!")
}

// GoStructProduct is a Go struct that represents the products table but includes an extra field not present in the actual database table.
type GoStructProduct struct {
	gorm.Model
	Name        string  `gorm:"column:name"`
	Price       float64 `gorm:"column:price"`
	RequiredCol string  `gorm:"column:required_col;not null"`
	ExtraField  string  `gorm:"column:extra_field"`
}

func (GoStructProduct) TableName() string {
	return "products"
}

func TestPostgresWithMissingColumns(t *testing.T) {
	ctx := setupPostgres(t)
	db := databaseFromContext(ctx)

	// Insert using the mismatched struct that references a column not in the table
	p := GoStructProduct{Name: "Widget", Price: 9.99, ExtraField: "should fail"}
	result := db.Create(&p)

	require.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "extra_field")
	fmt.Println("Schema mismatch error caught:", result.Error)
}
