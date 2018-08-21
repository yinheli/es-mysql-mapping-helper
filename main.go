package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gobwas/glob"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

var (
	Build = "devel"

	logger = log.New(os.Stderr, "", log.LstdFlags)

	V = flag.Bool("version", false, "show version")
	f = flag.String("c", "config.yml", "config file")
	t = flag.Bool("t", false, "index use current as index suffix (version)")
)

func main() {
	flag.Parse()
	if !flag.Parsed() {
		flag.Usage()
		return
	}

	if *V {
		fmt.Println("es-mysql-mapping-helper")
		fmt.Println("  version   :", Build)
		fmt.Println("  go version:", runtime.Version())
		return
	}

	// load config
	config, err := loadConfig(*f)
	if err != nil {
		logger.Fatal(err)
	}

	db, err := openDB(config.Database)
	if err != nil {
		logger.Fatal(err)
	}

	now := time.Now().Format("200601021504")

	suffix := ""
	if *t {
		suffix = fmt.Sprintf("%s", now)
	}

	// read tables
	tables, err := getTables(db, config)
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Println("// build via es-mysql-mapping-helper, version:", Build)
	fmt.Println("// time: ", now)
	fmt.Println("")

	// read table columns
	for _, table := range tables {
		cols, err := getColumns(db, table)

		mappings := &Mappings{
			Settings: &Settings{
				Index: &Index{
					NumberOfShards:   config.Index.Shards,
					NumberOfReplicas: config.Index.Replicas,
				},
			},
			Mappings: make(map[string]*Columns),
		}

		rule := getRule(config.Rules, table)

		columns := &Columns{
			Properties: make(map[string]*Column),
		}

		index := table

		if rule != nil {
			if rule.Index != "" {
				index = rule.Index
			}

			if rule.Shards != 0 && rule.Shards != config.Index.Shards {
				mappings.Settings.Index.NumberOfShards = rule.Shards
			}

			if rule.Replicas > config.Index.Replicas {
				mappings.Settings.Index.NumberOfReplicas = rule.Replicas
			}
		}

		for _, c := range cols {
			col := getRuleCol(rule, c)
			columns.Properties[c.Name] = col
		}

		mappings.Mappings[index] = columns

		indexName := fmt.Sprintf("%s%s", config.Index.Prefix, table)
		if suffix != "" {
			indexName = fmt.Sprintf("%s_%s", indexName, suffix)
		}

		b, err := json.MarshalIndent(mappings, "", "  ")
		if err != nil {
			logger.Fatal(err)
		}
		result := fmt.Sprintf("PUT /%s\n%s\n\n", indexName, string(b))

		fmt.Println(result)
	}
}

func getRuleCol(rule *Rule, dbColumn *DbColumn) *Column {
	var s *ColumnSetting
	if rule != nil {
		if len(rule.SearchableColumns) > 0 {
			for _, c := range rule.SearchableColumns {
				if strings.ToLower(c) == dbColumn.Name {
					return &Column{
						Type:           "text",
						Analyzer:       "ik_max_word",
						SearchAnalyzer: "ik_max_word",
					}
				}
			}
		}

		for _, c := range rule.Columns {
			if strings.ToLower(c.Name) == dbColumn.Name {
				s = c
				break
			}
		}
	}

	if s != nil {
		return &Column{
			Type:           s.Type,
			Analyzer:       s.Analyzer,
			SearchAnalyzer: s.SearchAnalyzer,
		}
	}

	return &Column{
		Type: mapType(dbColumn.Type),
	}
}
func mapType(dbType string) string {
	switch dbType {
	case "varchar", "char", "tinytext", "text", "mediumtext", "longtext", "enum", "set":
		return "text"
	case "bigint":
		return "long"
	case "int", "tinyint", "smallint", "mediumint", "bit":
		return "integer"
	case "float", "double", "decimal":
		return "double"
	case "date", "datetime", "timestamp", "time":
		return "date"
	case "blob", "binary", "longblob":
		return "binary"
	}

	return "text"
}

func getRule(rules []*Rule, table string) *Rule {
	for _, r := range rules {
		if r.Table == table {
			return r
		}
	}

	return nil
}

func openDB(config *Database) (*sql.DB, error) {
	return sql.Open("mysql", config.Uri)
}

func getTables(db *sql.DB, config *Config) ([]string, error) {
	rows, err := db.Query("select t.table_name name from information_schema.tables t where t.table_schema = database()")
	if err != nil {
		return nil, err
	}

	var tables []string
	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		if err != nil {
			return nil, err
		}

		if len(config.Tables) > 0 {
			if isTableMatch(table, config.TablesGlob) {
				tables = append(tables, table)
			}
		} else {
			tables = append(tables, table)
		}
	}

	return tables, nil
}

func getColumns(db *sql.DB, table string) ([]*DbColumn, error) {
	rows, err := db.Query("select lower(column_name), lower(data_type) from information_schema.Columns t where t.table_schema=database() and t.table_name=?", table)
	if err != nil {
		return nil, err
	}
	var cols []*DbColumn
	for rows.Next() {
		var col DbColumn
		err := rows.Scan(&col.Name, &col.Type)
		if err != nil {
			return nil, err
		}
		cols = append(cols, &col)
	}

	return cols, nil
}

func isTableMatch(table string, filters []glob.Glob) bool {
	for _, v := range filters {
		if v.Match(table) {
			return true
		}
	}
	return false
}
