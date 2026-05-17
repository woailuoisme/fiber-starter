package drivers

import (
	"database/sql"

	"fiber-starter/configs"
)

// Connector is the interface for all database connectors (similar to Laravel's Connector)
type Connector interface {
	Connect(cfg configs.DBConnection) (*sql.DB, error)
	DriverName() string
}
