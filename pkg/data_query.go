/*
Copyright 2025 API Testing Authors.

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
package pkg

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/linuxsuren/api-testing/pkg/server"
	"gorm.io/gorm"
)

func (s *dbserver) Query(ctx context.Context, query *server.DataQuery) (result *server.DataQueryResult, err error) {
	var db *gorm.DB
	var dbQuery DataQuery
	if dbQuery, err = s.getClientWithDatabase(ctx, query.Key); err != nil {
		return
	}

	db = dbQuery.GetClient()

	result = &server.DataQueryResult{
		Data:  []*server.Pair{},
		Items: make([]*server.Pairs, 0),
		Meta:  &server.DataMeta{},
	}

	// query database and tables
	if result.Meta.Databases, err = dbQuery.GetDatabases(ctx); err != nil {
		log.Printf("failed to query databases: %v\n", err)
	}

	if result.Meta.CurrentDatabase = query.Key; query.Key == "" {
		if result.Meta.CurrentDatabase, err = dbQuery.GetCurrentDatabase(); err != nil {
			log.Printf("failed to query current database: %v\n", err)
		}
	}

	if result.Meta.Tables, err = dbQuery.GetTables(ctx, result.Meta.CurrentDatabase); err != nil {
		log.Printf("failed to query tables: %v\n", err)
	}

	// query data
	if query.Sql == "" {
		return
	}

	var dataResult *server.DataQueryResult
	now := time.Now()
	if dataResult, err = sqlQuery(ctx, query.Sql, db); err == nil {
		result.Items = dataResult.Items
		result.Meta.Duration = time.Since(now).String()
	}
	return
}

func sqlQuery(ctx context.Context, sql string, db *gorm.DB) (result *server.DataQueryResult, err error) {
	rows, err := db.Raw(sql).Rows()
	if err != nil {
		return
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	result = &server.DataQueryResult{
		Data:  []*server.Pair{},
		Items: make([]*server.Pairs, 0),
		Meta:  &server.DataMeta{},
	}

	if rows == nil {
		if rows, err = db.ConnPool.QueryContext(ctx, sql); err != nil {
			return
		} else if rows == nil {
			fmt.Println("no rows found")
			return
		}
	}

	columns, err := rows.Columns()
	if err != nil {
		return
	}

	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columnsData := make([]interface{}, len(columns))
		columnsPointers := make([]interface{}, len(columns))
		for i := range columnsData {
			columnsPointers[i] = &columnsData[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnsPointers...); err != nil {
			return nil, err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		for i, colName := range columns {
			rowData := &server.Pair{}
			val := columnsData[i]

			rowData.Key = colName
			switch v := val.(type) {
			case []byte:
				rowData.Value = string(v)
			case string:
				rowData.Value = v
			case int, uint64, uint32, int32, int64:
				rowData.Value = fmt.Sprintf("%d", v)
			case float32, float64:
				rowData.Value = fmt.Sprintf("%f", v)
			case time.Time:
				rowData.Value = v.String()
			case bool:
				rowData.Value = fmt.Sprintf("%t", v)
			case nil:
				rowData.Value = "null"
			case []int, []uint64, []uint32, []int32, []int64:
				rowData.Value = fmt.Sprintf("%v", v)
			case []float32, []float64:
				rowData.Value = fmt.Sprintf("%v", v)
			case []string:
				rowData.Value = fmt.Sprintf("%v", v)
			default:
				fmt.Println("column", colName, "type", reflect.TypeOf(v))
			}

			// Append the map to our slice of maps.
			result.Data = append(result.Data, rowData)
		}
		result.Items = append(result.Items, &server.Pairs{
			Data: result.Data,
		})
	}
	return
}

const queryDatabaseSql = "show databases"

type DataQuery interface {
	GetDatabases(context.Context) (databases []string, err error)
	GetTables(ctx context.Context, currentDatabase string) (tables []string, err error)
	GetCurrentDatabase() (string, error)
	GetClient() *gorm.DB
}

type commonDataQuery struct {
	showDatabases, showTables, currentDatabase string
	db                                         *gorm.DB
}

var _ DataQuery = &commonDataQuery{}

func NewCommonDataQuery(showDatabases, showTables, currentDatabase string, db *gorm.DB) DataQuery {
	return &commonDataQuery{
		showDatabases:   showDatabases,
		showTables:      showTables,
		currentDatabase: currentDatabase,
		db:              db,
	}
}

func (q *commonDataQuery) GetDatabases(ctx context.Context) (databases []string, err error) {
	var databaseResult *server.DataQueryResult
	if databaseResult, err = sqlQuery(ctx, q.showDatabases, q.db); err == nil {
		for _, table := range databaseResult.Items {
			for _, item := range table.GetData() {
				if item.Key == "Database" || item.Key == "name" {
					var found bool
					for _, name := range databases {
						if name == item.Value {
							found = true
						}
					}
					if !found {
						databases = append(databases, item.Value)
					}
				}
			}
		}
		sort.Strings(databases)
	}
	return
}

func (q *commonDataQuery) GetTables(ctx context.Context, currentDatabase string) (tables []string, err error) {
	showTables := q.showTables
	if strings.Contains(showTables, "%s") {
		showTables = fmt.Sprintf(showTables, currentDatabase)
	}

	var tableResult *server.DataQueryResult
	if tableResult, err = sqlQuery(ctx, showTables, q.db); err == nil {
		for _, table := range tableResult.Items {
			for _, item := range table.GetData() {
				if item.Key == fmt.Sprintf("Tables_in_%s", currentDatabase) || item.Key == "table_name" ||
					item.Key == "Tables" || item.Key == "tablename" {
					var found bool
					for _, name := range tables {
						if name == item.Value {
							found = true
						}
					}
					if !found {
						tables = append(tables, item.Value)
					}
				}
			}
		}
		sort.Strings(tables)
	}
	return
}
func (q *commonDataQuery) GetCurrentDatabase() (current string, err error) {
	var row *sql.Row
	if row = q.db.Raw(q.currentDatabase).Row(); row != nil {
		err = row.Scan(&current)
	}
	return
}

func (q *commonDataQuery) GetClient() *gorm.DB {
	return q.db
}
