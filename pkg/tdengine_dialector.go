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
	"database/sql"
	"fmt"
	_ "github.com/taosdata/driver-go/v3/taosWS"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var _ gorm.Dialector = &tdengineDialector{}

type tdengineDialector struct {
	DSN string
}

func (d tdengineDialector) Name() string {
	return "tdengine"
}

func (d tdengineDialector) Initialize(db *gorm.DB) (err error) {
	// Initialize the TDengine connection here
	fmt.Println("init", d.DSN)
	if db.ConnPool == nil {
		db.ConnPool, err = sql.Open("taosWS", d.DSN)
	}
	return
}

func (d tdengineDialector) Migrator(db *gorm.DB) gorm.Migrator {
	// Return the TDengine migrator here
	return &mysql.Migrator{}
}

func (d tdengineDialector) DataTypeOf(field *schema.Field) string {
	// Return the TDengine data type for the given field
	switch field.DataType {
	case schema.Bool:
		return "bool"
	case schema.Int, schema.Uint:
		sqlType := "bigint"
		switch {
		case field.Size <= 8:
			sqlType = "tinyint"
		case field.Size <= 16:
			sqlType = "smallint"
		case field.Size <= 32:
			sqlType = "int"
		}
		return sqlType
	case schema.Float:
		if field.Size <= 32 {
			return "float"
		}
		return "double"
	case schema.String:
		size := field.Size
		if size == 0 {
			size = 64
		}
		return fmt.Sprintf("NCHAR(%d)", size)
	case schema.Time:
		return "TIMESTAMP"
	case schema.Bytes:
		size := field.Size
		if size == 0 {
			size = 64
		}
		return fmt.Sprintf("BINARY(%d)", size)
	}

	return string(field.DataType)
}

func (d tdengineDialector) DefaultValueOf(field *schema.Field) clause.Expression {
	return clause.Expr{SQL: "NULL"}
}

func (d tdengineDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
	// Bind variables for TDengine
	switch v.(type) {
	case string:
		writer.WriteString("'?'")
	default:
		writer.WriteByte('?')
	}
}

func (d tdengineDialector) QuoteTo(writer clause.Writer, str string) {
	// Quote identifiers for TDengine
	writer.WriteString(str)
}

func (d tdengineDialector) Explain(sql string, vars ...interface{}) string {
	// Explain the SQL query for TDengine
	return logger.ExplainSQL(sql, nil, "'", vars...)
}

func NewTDengineDialector(dsn string) gorm.Dialector {
	return tdengineDialector{DSN: dsn}
}
