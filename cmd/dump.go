package cmd

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3" // SQLite driver usage
	"github.com/spf13/cobra"
)

// https://stackoverflow.com/a/10485970/2610955
func contains(tables []string, target string) bool {
	for _, table := range tables {
		if table == target {
			return true
		}
	}
	return false
}

var dumpCmd = &cobra.Command{
	Use:   "dump [tables]",
	Short: "Dump sqlite database metadata and table",
	Long:  `Dump sqlite database metadata and table`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		filePath, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				err := errors.New("Path doesn't exist")
				return err
			}
			return err
		}

		if filePath.IsDir() {
			err := errors.New("Path should be a file")
			return err
		}

		output, _ := cmd.Flags().GetString("output")
		_, err = os.Stat(output)
		if err != nil {
			if os.IsNotExist(err) {
				err := os.MkdirAll(output, os.ModePerm)
				if err != nil {
					return err
				}
			}
		}

		outputPath, err := os.Stat(output)
		if err != nil {
			if os.IsNotExist(err) {
				return err
			}
		}

		if !outputPath.IsDir() {
			err := errors.New("Output should be a directory")
			return err
		}

		db, err := sql.Open("sqlite3", path)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		allTables := []string{}
		queryTables := []string{}
		all, _ := cmd.Flags().GetBool("all")

		if !all && len(args) == 0 {
			msg := "You must pass --all or specify some tables"
			return errors.New(msg)
		}

		allTablesQuery := "SELECT name FROM sqlite_master WHERE type='table'"
		rows, _ := db.Query(allTablesQuery)
		for rows.Next() {
			var table string
			rows.Scan(&table)
			allTables = append(allTables, table)
		}

		if all {
			queryTables = allTables
		} else {
			for _, arg := range args {
				if contains(allTables, arg) {
					queryTables = append(queryTables, arg)
				}
			}
		}

		for _, table := range queryTables {
			rows, err := db.Query("select * from " + table)
			if err != nil {
				return err
			}

			if rows == nil {
				err := errors.New("No rows found for table " + table)
				return err
			}

			// https://stackoverflow.com/questions/17845619/how-to-call-the-scan-variadic-function-using-reflection/17885636#17885636
			columns, _ := rows.Columns()
			count := len(columns)
			values := make([]interface{}, count)
			valuePtrs := make([]interface{}, count)
			lines := []string{}

			schemaResult := db.QueryRow("SELECT sql FROM sqlite_schema WHERE name = ?", table)
			var schema string
			schemaResult.Scan(&schema)

			metadata := map[string]interface{}{
				"name":    table,
				"columns": columns,
				"schema":  schema,
			}
			metadataJSON, _ := json.MarshalIndent(metadata, "", "    ")

			for rows.Next() {
				for i := range columns {
					valuePtrs[i] = &values[i]
				}

				rows.Scan(valuePtrs...)
				var entries = []interface{}{}

				for i := range columns {
					val := values[i]

					b, ok := val.([]byte)
					var v interface{}
					if ok {
						v = string(b)
					} else {
						v = val
					}

					entries = append(entries, v)
				}
				data, _ := json.Marshal(entries)
				jsonString := string(data)
				lines = append(lines, jsonString)
			}

			metadataPath := filepath.Join(output, table+".metadata.json")
			tablePath := filepath.Join(output, table+".ndjson")
			err = ioutil.WriteFile(metadataPath,
				[]byte(metadataJSON), 0644)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(tablePath,
				[]byte(strings.Join(lines, "\n")), 0644)
			if err != nil {
				return err
			}
		}

		return nil

	},
}

func init() {
	dumpCmd.Flags().StringP("path", "p", "", "Path to sqlite database")
	dumpCmd.Flags().StringP("output", "o", "", "Output directory")
	dumpCmd.Flags().BoolP("all", "", false, "Dump all tables")

	dumpCmd.MarkFlagRequired("path")
	dumpCmd.MarkFlagRequired("output")
}
