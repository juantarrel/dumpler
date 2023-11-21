package infrastructure

import (
	"database/sql"
	"github.com/spf13/viper"
)

type SQL struct {
	db *sql.DB
}

func (c *Context) ConnectToDatabase(config *viper.Viper) (*sql.DB, error) {
	return c.database.Connect(config)
}
