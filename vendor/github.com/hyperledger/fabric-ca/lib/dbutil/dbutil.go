/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

                 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dbutil

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudflare/cfssl/log"
	"github.com/go-sql-driver/mysql"
	"github.com/hyperledger/fabric-ca/lib/tls"
	"github.com/jmoiron/sqlx"
)

// NewUserRegistrySQLLite3 returns a pointer to a sqlite database
func NewUserRegistrySQLLite3(datasource string) (*sqlx.DB, bool, error) {
	log.Debugf("Using sqlite database, connect to database in home (%s) directory", datasource)

	datasource = filepath.Join(datasource)
	var exists bool

	if datasource != "" {
		// Check if database exists if not create it and bootstrap it based on the config file
		if _, err := os.Stat(datasource); err != nil {
			if os.IsNotExist(err) {
				log.Debugf("Database (%s) does not exist", datasource)
				exists = false
				err2 := createSQLiteDBTables(datasource)
				if err2 != nil {
					return nil, false, err2
				}
			} else {
				log.Debug("Database (%s) exists", datasource)
				exists = true
			}
		}
	}

	db, err := sqlx.Open("sqlite3", datasource)
	if err != nil {
		return nil, false, err
	}

	log.Debug("Successfully opened sqlite3 DB")

	return db, exists, nil
}

func createSQLiteDBTables(datasource string) error {
	log.Debug("Creating SQLite Database...")
	log.Debug("Database location: ", datasource)
	db, err := sqlx.Open("sqlite3", datasource)
	if err != nil {
		return fmt.Errorf("Failed to open database: %s", err)
	}

	log.Debug("Creating tables...")
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS users (id VARCHAR(64), token bytea, type VARCHAR(64), affiliation VARCHAR(64), attributes VARCHAR(256), state INTEGER,  max_enrollments INTEGER)"); err != nil {
		return err
	}
	log.Debug("Created users table")

	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS affiliations (name VARCHAR(64), prekey VARCHAR(64))"); err != nil {
		return err
	}
	log.Debug("Created affiliation table")

	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS certificates (id VARCHAR(64), serial_number bytea NOT NULL, authority_key_identifier bytea NOT NULL, ca_label bytea, status bytea NOT NULL, reason int, expiry timestamp, revoked_at timestamp, pem bytea NOT NULL, PRIMARY KEY(serial_number, authority_key_identifier))"); err != nil {
		return err
	}
	log.Debug("Created certificates table")

	return nil
}

// NewUserRegistryPostgres opens a connecton to a postgres database
func NewUserRegistryPostgres(datasource string, clientTLSConfig *tls.ClientTLSConfig) (*sqlx.DB, bool, error) {
	log.Debugf("Using postgres database, connecting to database...")

	var exists bool
	dbName := getDBName(datasource)
	log.Debug("Database Name: ", dbName)

	if strings.Contains(dbName, "-") || strings.HasSuffix(dbName, ".db") {
		return nil, false, fmt.Errorf("Database name %s cannot contain any '-' or end with '.db'", dbName)
	}

	connStr := getConnStr(datasource)

	if clientTLSConfig.Enabled {
		if len(clientTLSConfig.CertFilesList) > 0 {
			root := clientTLSConfig.CertFilesList[0]
			connStr = fmt.Sprintf("%s sslrootcert=%s", connStr, root)
		}

		cert := clientTLSConfig.Client.CertFile
		key := clientTLSConfig.Client.KeyFile
		connStr = fmt.Sprintf("%s sslcert=%s sslkey=%s", connStr, cert, key)
	}

	log.Debug("Connection String: ", connStr)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, false, fmt.Errorf("Failed to open database: %s", err)
	}

	err = db.Ping()
	if err != nil {
		log.Errorf("Failed to connect to Postgres database [error: %s]", err)
		return nil, false, err
	}

	// Check if database exists
	r, err2 := db.Exec("SELECT * FROM pg_catalog.pg_database where datname=$1", dbName)
	if err2 != nil {
		return nil, false, fmt.Errorf("Failed to query 'pg_database' table: %s", err2)
	}

	found, _ := r.RowsAffected()
	if found == 0 {
		log.Debugf("Database (%s) does not exist", dbName)
		exists = false
		connStr = connStr + " dbname=" + dbName
		err = createPostgresDBTables(connStr, dbName, db)
		if err != nil {
			return nil, false, err
		}
	} else {
		log.Debugf("Database (%s) exists", dbName)
		exists = true
	}

	db, err = sqlx.Open("postgres", datasource)
	if err != nil {
		return nil, false, err
	}

	return db, exists, nil
}

// createPostgresDB creates postgres database
func createPostgresDBTables(datasource string, dbName string, db *sqlx.DB) error {
	log.Debugf("Creating Postgres Database (%s)...", dbName)
	query := "CREATE DATABASE " + dbName
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("Failed to create Postgres database: %s", err)
	}

	database, err := sqlx.Open("postgres", datasource)
	if err != nil {
		log.Errorf("Failed to open database (%s)", dbName)
	}

	log.Debug("Creating Tables...")
	if _, err := database.Exec("CREATE TABLE users (id VARCHAR(64), token bytea, type VARCHAR(64), affiliation VARCHAR(64), attributes VARCHAR(256), state INTEGER,  max_enrollments INTEGER)"); err != nil {
		log.Errorf("Error creating users table [error: %s] ", err)
		return err
	}
	if _, err := database.Exec("CREATE TABLE affiliations (name VARCHAR(64), prekey VARCHAR(64))"); err != nil {
		log.Errorf("Error creating affiliations table [error: %s] ", err)
		return err
	}
	if _, err := database.Exec("CREATE TABLE certificates (id VARCHAR(64), serial_number bytea NOT NULL, authority_key_identifier bytea NOT NULL, ca_label bytea, status bytea NOT NULL, reason int, expiry timestamp, revoked_at timestamp, pem bytea NOT NULL, PRIMARY KEY(serial_number, authority_key_identifier))"); err != nil {
		log.Errorf("Error creating certificates table [error: %s] ", err)
		return err
	}
	return nil
}

// NewUserRegistryMySQL opens a connecton to a postgres database
func NewUserRegistryMySQL(datasource string, clientTLSConfig *tls.ClientTLSConfig) (*sqlx.DB, bool, error) {
	log.Debugf("Using MySQL database, connecting to database...")

	var exists bool
	dbName := getDBName(datasource)
	log.Debug("Database Name: ", dbName)

	re := regexp.MustCompile(`\/([a-zA-z]+)`)
	connStr := re.ReplaceAllString(datasource, "/")

	if clientTLSConfig.Enabled {
		tlsConfig, err := tls.GetClientTLSConfig(clientTLSConfig)
		if err != nil {
			log.Errorf("Failed to create TLS configuration [error: %s]", err)
		}

		mysql.RegisterTLSConfig("custom", tlsConfig)
	}

	log.Debug("Connection String: ", connStr)
	db, err := sqlx.Open("mysql", connStr)
	if err != nil {
		return nil, false, fmt.Errorf("Failed to open database: %s", err)
	}

	err = db.Ping()
	if err != nil {
		log.Errorf("Failed to connect to MySQL database [error: %s]", err)
	}

	// Check if database exists
	var name string
	err = db.QueryRow("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", dbName).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debugf("Database (%s) does not exist", dbName)
			exists = false
		} else {
			return nil, false, fmt.Errorf("Failed to query 'INFORMATION_SCHEMA.SCHEMATA table: %s", err)
		}
	}

	if name == "" {
		createMySQLTables(datasource, dbName, db)
	} else {
		log.Debugf("Database (%s) exists", dbName)
		exists = true
	}

	db, err = sqlx.Open("mysql", datasource)
	if err != nil {
		return nil, false, err
	}

	return db, exists, nil
}

func createMySQLTables(datasource string, dbName string, db *sqlx.DB) error {
	log.Debugf("Creating MySQL Database (%s)...", dbName)

	_, err := db.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		panic(err)
	}

	database, err := sqlx.Open("mysql", datasource)
	if err != nil {
		log.Errorf("Failed to open database (%s), err: %s", dbName, err)
	}
	log.Debug("Creating Tables...")
	if _, err := database.Exec("CREATE TABLE users (id VARCHAR(64) NOT NULL, token blob, type VARCHAR(64), affiliation VARCHAR(64), attributes VARCHAR(256), state INTEGER, max_enrollments INTEGER, PRIMARY KEY (id))"); err != nil {
		log.Errorf("Error creating users table [error: %s] ", err)
		return err
	}
	if _, err := database.Exec("CREATE TABLE affiliations (name VARCHAR(64), prekey VARCHAR(64))"); err != nil {
		log.Errorf("Error creating affiliations table [error: %s] ", err)
		return err
	}
	if _, err := database.Exec("CREATE TABLE certificates (id VARCHAR(64), serial_number varbinary(128) NOT NULL, authority_key_identifier varbinary(128) NOT NULL, ca_label varbinary(128), status varbinary(128) NOT NULL, reason int, expiry timestamp DEFAULT '1970-01-01 00:00:01', revoked_at timestamp DEFAULT '1970-01-01 00:00:01', pem varbinary(4096) NOT NULL, PRIMARY KEY(serial_number, authority_key_identifier))"); err != nil {
		log.Errorf("Error creating certificates table [error: %s] ", err)
		return err
	}

	return nil
}

// GetDBName gets database name from connection string
func getDBName(datasource string) string {
	var dbName string
	datasource = strings.ToLower(datasource)

	re := regexp.MustCompile(`(?:\/([^\/?]+))|(?:dbname=([^\s]+))`)
	getName := re.FindStringSubmatch(datasource)
	if getName != nil {
		dbName = getName[1]
		if dbName == "" {
			dbName = getName[2]
		}
	}

	return dbName
}

// GetConnStr gets connection string without database
func getConnStr(datasource string) string {
	re := regexp.MustCompile(`(dbname=)([^\s]+)`)
	connStr := re.ReplaceAllString(datasource, "")
	return connStr
}
