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

	"github.com/linuxsuren/api-testing/pkg/server"
	"gorm.io/gorm"
)

func (s *dbserver) Query(ctx context.Context, query *server.DataQuery) (result *server.DataQueryResult, err error) {
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}

	rows, err := db.Raw(query.Sql).Rows()
	if err != nil {
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return
	}

	result = &server.DataQueryResult{
		Data:  []*server.Pair{},
		Items: make([]*server.Pairs, 0),
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
			b, ok := val.([]byte)
			if ok {
				rowData.Key = colName
				rowData.Value = string(b)
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
