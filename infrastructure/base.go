package infrastructure

import (
	"database/sql"
	"github.com/spf13/viper"
)

func (c *Context) ConnectToDatabase(config *viper.Viper) (*sql.DB, error) {
	return c.database.Connect(config)
}
