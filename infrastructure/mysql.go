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

var batchSize = 45000

func (m *SQL) Connect(config *viper.Viper) (*sql.DB, error) {
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

func (m *SQL) Ping() {
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

func (m *SQL) Dump() {
	start := time.Now()
	log.Printf("[info] [source] start at %s\n", start.Format("2006-01-02 15:04:05"))
	defer func() {
		end := time.Now()
		log.Printf("[info] [source] end at %s, cost %s\n", end.Format("2006-01-02 15:04:05"), end.Sub(start))
	}()

	tables, err := m.getTables()
	if err != nil {
		slog.Error("Error: ", err)
		return
	}

	slog.Debug(fmt.Sprintf("Dumping %d tables", len(tables)))

	maxConcurrent := runtime.NumCPU()
	maxConcurrentToUse := int(math.Floor(float64(maxConcurrent) * 0.4))
	slog.Debug(fmt.Sprintf("Num CPU: %d", maxConcurrent))
	slog.Debug(fmt.Sprintf("Num CPU to use: %d", maxConcurrentToUse))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrentToUse)

	if err != nil {
		slog.Error("Error: ", err)
		return
	}
	defer m.db.Close()

	var errorsMu sync.Mutex
	var allErrors []error

	filePath, err := fs.GetAppCacheDir()
	if err != nil {
		errorsMu.Lock()
		allErrors = append(allErrors, err)
		errorsMu.Unlock()
		return
	}
	counter := 0
	for _, name := range tables {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()

			file, err := os.Create(fmt.Sprintf("%s/%s.sql", filePath, name))
			if err != nil {
				errorsMu.Lock()
				allErrors = append(allErrors, err)
				errorsMu.Unlock()
			}
			defer file.Close()

			err = m.createTableDML(name, m.db, file)
			if err != nil {
				errorsMu.Lock()
				allErrors = append(allErrors, err)
				errorsMu.Unlock()
			}
			counter++
			slog.Debug(fmt.Sprintf("%d/%d", counter, len(tables)))
		}(name)
	}

	wg.Wait()

	if len(allErrors) > 0 {
		for _, err := range allErrors {
			slog.Error("Error: ", err)
		}
	}
}

func (m *SQL) getTables() ([]string, error) {
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

func (m *SQL) CreateTableDDL(name string) (string, error) {
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
func (m *SQL) createTableDML(name string, db *sql.DB, file *os.File) error {
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
	data := make([]string, len(columns))
	pointersData := make([]interface{}, len(columns))

	for i := range data {
		pointersData[i] = &data[i]
	}

	for rows.Next() {
		err = rows.Scan(pointersData...)

		for i, value := range data {
			comma := ","
			if i == len(data)-1 {
				comma = ""
			}
			if value != "NULL" {
				dataBuffer.WriteString("'" + strings.Trim(value, "\r") + "'" + comma)
			} else {
				dataBuffer.WriteString(value + comma)

			}
		}
		dataBuffer.WriteString("\n")
		batchCount++

		if batchCount >= batchSize {
			if _, err := file.WriteString(dataBuffer.String()); err != nil {
				return err
			}
			dataBuffer.Reset()
			batchCount = 0
		}
	}

	if dataBuffer.Len() > 0 {
		if _, err := file.WriteString(dataBuffer.String()); err != nil {
			return err
		}
	}

	slog.Debug(fmt.Sprintf("DML for table %s finished", name))
	return rows.Err()
}

func (m *SQL) Bulk(viper *viper.Viper) error {
	return nil
}
