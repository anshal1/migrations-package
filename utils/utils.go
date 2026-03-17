package utils

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

type MigrationsConfig struct {
	InputDirectory string   `json:"inputDirectory"`
	Exclude        []string `json:"exclude"`
}

func GetConfig() MigrationsConfig {
	var config MigrationsConfig
	path := filepath.Join(".", "m-config.json")
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return config
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Println(err)
		return config
	}
	return config
}

func createMigrationTable(db *sql.DB) error {
	_, err := db.Exec("create table if not exists migration (version varchar(255) not null, applied_at timestamp not null default current_timestamp);")
	return err
}

func isMigrationApplied(migrationFileName string, db *sql.DB) (bool, error) {
	version := strings.Split(migrationFileName, ".")[0]
	if version == "" {
		return true, errors.New("No file name found")
	}

	var versionExists string
	err := db.QueryRow("select applied_at from migration where version = $1", version).Scan(&versionExists)
	if err != sql.ErrNoRows {
		return true, err
	}

	if versionExists != "" {
		return true, nil
	}
	return false, nil
}

func applyMigration(version string, sql string, db *sql.DB) error {
	if version == "" {
		return errors.New("No version provided")
	}
	_, err := db.Exec("INSERT INTO migration (version) VALUES ($1)", version)
	if err != nil {
		fmt.Println(1)
		return err
	}
	_, err = db.Exec(sql)
	if err != nil {
		_, err := db.Exec("delete from migration where version ($1)", version)
		if err != nil {
			fmt.Println(2)
			return err
		}
		fmt.Println(3)
		return err
	}
	return nil
}

func connectToDB(databaseURI string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURI)
	if err != nil {
		return db, err
	}
	error := db.Ping()
	if error != nil {
		return db, error
	}

	return db, err
}
func readDir(config MigrationsConfig) ([]os.DirEntry, error) {
	path := filepath.Join(".", config.InputDirectory)
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	allFiles, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	return allFiles, err
}

func CreateMigrations(config MigrationsConfig, dbUri string) error {
	db, err := connectToDB(dbUri)
	if err != nil {
		return err
	}
	err = createMigrationTable(db)
	if err != nil {
		return err
	}
	files, err := readDir(config)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		path := filepath.Join(".", config.InputDirectory, name)
		sql, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		applied, err := isMigrationApplied(name, db)
		if err != nil {
			fmt.Println("From is applied")
			return err
		}
		if applied {
			fmt.Println("Migration for", name, "skipped as its already applied")
			continue
		}
		sqlString := string(sql)
		err = applyMigration(strings.Split(name, ".")[0], sqlString, db)
		if err != nil {
			fmt.Println("from apply")
			return err
		}
	}
	fmt.Println("Migrations applied")
	return nil
}
