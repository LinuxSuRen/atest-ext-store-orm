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

import "strings"

type InnerSQL interface {
	ToNativeSQL(query string) string
}

const (
	innerSelectTable_  = "@selectTable_"
	innerShowDatabases = "@showDatabases"
	innerShowTables    = "@showTables"
	innerCurrentDB     = "@currentDB"
)

func GetInnerSQL(dialect string) InnerSQL {
	switch dialect {
	case "postgres":
		return &postgresDialect{}
	default:
		return &mysqlDialect{}
	}
}

type mysqlDialect struct {
}

func (m *mysqlDialect) ToNativeSQL(query string) (sql string) {
	if strings.HasPrefix(query, innerSelectTable_) {
		sql = "SELECT * FROM " + strings.ReplaceAll(query, innerSelectTable_, "")
	} else if query == innerShowDatabases {
		sql = "SHOW DATABASES"
	} else if query == innerShowTables {
		sql = "SHOW TABLES"
	} else if query == innerCurrentDB {
		sql = "SELECT DATABASE() as name"
	} else {
		sql = query
	}
	return
}

type postgresDialect struct {
}

func (p *postgresDialect) ToNativeSQL(query string) (sql string) {
	if strings.HasPrefix(query, innerSelectTable_) {
		sql = `SELECT * FROM "` + strings.ReplaceAll(query, innerSelectTable_, "") + `"`
	} else if query == innerShowDatabases {
		sql = "SELECT table_catalog as name FROM information_schema.tables"
	} else if query == innerShowTables {
		sql = `SELECT table_name FROM information_schema.tables WHERE table_catalog = '%s' and table_schema != 'pg_catalog' and table_schema != 'information_schema'`
	} else if query == innerCurrentDB {
		sql = "SELECT current_database() as name"
	} else {
		sql = query
	}
	return
}
