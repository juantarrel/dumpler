package infrastructure

import (
	"database/sql"
	"github.com/spf13/viper"
)

type Database interface {
	Connect(*viper.Viper) (*sql.DB, error)
	Dump()
	Bulk(*viper.Viper) error
}

type Context struct {
	database Database
}

func NewContext(database Database) *Context {
	return &Context{
		database: database,
	}
}
