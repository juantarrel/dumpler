package infrastructure

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"log/slog"
)

type MySQL struct {
	db *sql.DB
}

func (m *MySQL) Connect(config *viper.Viper) (*sql.DB, error) {
	var host string
	if config.IsSet("mysql-host") {
		host = config.GetString("mysql-host")
	}

	var port string
	if config.IsSet("mysql-port") {
		port = config.GetString("mysql-port")
	}

	var username string
	if config.IsSet("mysql-username") {
		username = config.GetString("mysql-username")
	}

	var password string
	if config.IsSet("mysql-password") {
		password = config.GetString("mysql-password")
	}

	var dbase string
	if config.IsSet("mysql-database") {
		dbase = config.GetString("mysql-database")
	}

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbase)
	slog.Debug("ConnectionString: " + connectionString)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		slog.Error("Panic occurred when is trying to connect to database", "error", err)
	}

	m.db = db

	slog.Debug("Connecting to MySQL Database")
	return db, nil
}

func (m *MySQL) Ping() {
	defer m.db.Close()
	err := m.db.Ping()
	if err != nil {
		slog.Error("Panic occurred", "error", err)
	}
	fmt.Println("Database exists")
}
