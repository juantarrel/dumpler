package infrastructure

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/juantarrel/dumpler/fs"
	"github.com/spf13/viper"
	"log"
	"log/slog"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type MySQL struct {
	db *sql.DB
}

type table struct {
	Name   string
	SQL    string
	Values string
}

type dump struct {
	DumpVersion   string
	ServerVersion string
	Tables        []*table
	CompleteTime  string
}

var batchSize = 10000

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
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(m.db)
	err := m.db.Ping()
	if err != nil {
		slog.Error("Panic occurred", "error", err)
	}
	fmt.Println("Database exists")
}

func (m *MySQL) Dump() {
	start := time.Now()
	log.Printf("[info] [source] start at %s\n", start.Format("2006-01-02 15:04:05"))
	defer func() {
		end := time.Now()
		log.Printf("[info] [source] end at %s, cost %s\n", end.Format("2006-01-02 15:04:05"), end.Sub(start))
	}()

	tables, err := m.getTables()
	if err != nil {
		slog.Error("Error: ", err)
	}

	slog.Debug(fmt.Sprintf("Dumping %d tables", len(tables)))

	maxConcurrent := runtime.NumCPU()
	maxConcurrentToUse := int(math.Floor(float64(maxConcurrent / 2)))
	slog.Debug(fmt.Sprintf("Num CPU: %d", maxConcurrent))
	slog.Debug(fmt.Sprintf("Num CPU: %d", maxConcurrentToUse))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrentToUse)
	// Get sql for each table
	for _, name := range tables {
		wg.Add(1)
		var file *os.File
		func() {
			filePath, err := fs.GetAppCacheDir()
			file, err = os.Create(fmt.Sprintf("%s/%s.sql", filePath, name))
			if err != nil {
				log.Fatal(err)
			}
		}()

		go func(name string, file *os.File) {
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()
			err = m.createTableDML(name, m.db, file)
			if err != nil {
				log.Fatal(err)
			}
			wg.Done()
		}(name, file)
	}
	wg.Wait()
}

// getTables @TODO needs to be in strategy pattern
func (m *MySQL) getTables() ([]string, error) {
	var tables []string

	rows, err := m.db.Query("SHOW TABLES")
	if err != nil {
		slog.Error("Error: ", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	for rows.Next() {
		var table sql.NullString
		if err := rows.Scan(&table); err != nil {
			return tables, err
		}
		tables = append(tables, table.String)
	}
	return tables, rows.Err()
}

// createTableDML @TODO needs to be in strategy pattern
func (m *MySQL) CreateTableDDL(name string) (string, error) {
	slog.Debug(fmt.Sprintf("DDL for table %s started", name))
	var tableReturn, tableSql sql.NullString
	err := m.db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE %s", name)).Scan(&tableReturn, &tableSql)

	if err != nil {
		slog.Error("Error: ", err)
		return "", err
	}

	if tableReturn.String != name {
		return "", errors.New("Returned table is not the same as requested table")
	}
	slog.Debug(fmt.Sprintf("DDL for table %s finished", name))
	return tableSql.String, nil
}

// createTableDML  @TODO needs to be in strategy pattern
func (m *MySQL) createTableDML(name string, db *sql.DB, file *os.File) error {
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	slog.Debug(fmt.Sprintf("DML for table %s started", name))

	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", name))
	if err != nil {
		slog.Error("Error: ", err)
		return err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	if len(columns) == 0 {
		return errors.New(fmt.Sprintf("No columns in table %s", name))
	}

	batchCount := 0
	var dataBuffer strings.Builder
	data := make([]*sql.NullString, len(columns))
	ptrs := make([]interface{}, len(columns))

	for i := range data {
		ptrs[i] = &data[i]
	}

	for rows.Next() {
		err = rows.Scan(ptrs...)

		rowString := ""
		for _, value := range data {
			if value != nil {
				rowString += "'" + strings.Trim(value.String, "\r") + "',"
			} else {
				rowString += "'NULL',"
			}
		}
		dataBuffer.WriteString(rowString)
		batchCount++

		if batchCount >= batchSize {
			file.WriteString(dataBuffer.String())
			dataBuffer.Reset()
			batchCount = 0
		}
	}

	if dataBuffer.Len() > 0 {
		file.WriteString(dataBuffer.String())
	}

	slog.Debug(fmt.Sprintf("DML for table %s finished", name))
	return rows.Err()
}

func (m *MySQL) getDatabaseSize(database string) (string, error) {
	var tableReturn, tableSql sql.NullString
	query := `
		SELECT
			table_schema "Database Name",
			sum(data_length + index_length) / 1024 / 1024 / 1024 "Size (GB)"
		FROM
			information_schema.tables
		GROUP BY
			table_schema;
`

	err := m.db.QueryRow(query).Scan(&tableReturn, &tableSql)

	if err != nil {
		slog.Error("Error: ", err)
		return "", err
	}

	return tableSql.String, nil
}

func (m *MySQL) getTableSize(database, table string) (string, error) {
	var tableReturn, tableSql sql.NullString
	query := `
		SELECT
			table_name "Table Name",
			table_schema "Database Name",
			round((data_length + index_length) / 1024 / 1024 / 1024, 2) "Size (GB)"
		FROM
			information_schema.tables
		WHERE
			table_schema = 'database'
			AND table_name = 'table';
`
	err := m.db.QueryRow(query).Scan(&tableReturn, &tableSql)

	if err != nil {
		slog.Error("Error: ", err)
		return "", err
	}

	return tableSql.String, nil
}
