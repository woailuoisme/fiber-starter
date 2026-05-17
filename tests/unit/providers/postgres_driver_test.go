package providers_test

import (
	"testing"

	"fiber-starter/configs"
	databaseDrivers "fiber-starter/internal/providers/database/drivers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresConnector(t *testing.T) {
	connector := &databaseDrivers.PostgresConnector{}
	assert.Equal(t, "pgx", connector.DriverName())

	_, err := connector.Connect(configs.DBConnection{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "postgres database name is empty")

	db, err := connector.Connect(configs.DBConnection{
		Host:     "127.0.0.1",
		Port:     "5432",
		Username: "postgres",
		Password: "secret",
		Database: "demo",
	})
	require.NoError(t, err)
	require.NoError(t, db.Close())
}
