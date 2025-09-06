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
	"github.com/linuxsuren/api-testing/pkg/testing"
	"github.com/linuxsuren/api-testing/pkg/testing/remote"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mcpServer struct {
	store *testing.Store
}

type DBQuery struct {
	SQL string `json:"sql" jsonschema:"the sql to be executed"`
}

type DatabaseQuery interface {
	Query(ctx context.Context, request *mcp.CallToolRequest, query DBQuery) (
		result *mcp.CallToolResult, a any, err error)
}

func NewMcpServer(store *testing.Store) DatabaseQuery {
	return &mcpServer{store: store}
}

func (s *mcpServer) Query(ctx context.Context, request *mcp.CallToolRequest, query DBQuery) (
	result *mcp.CallToolResult, a any, err error) {
	db := &dbserver{}
	ctx = remote.WithIncomingStoreContext(ctx, s.store)
	result = &mcp.CallToolResult{}

	var queryResult *server.DataQueryResult
	if queryResult, err = db.Query(ctx, &server.DataQuery{
		Sql: query.SQL,
	}); err == nil {
		result.StructuredContent = queryResult
		result.Content = []mcp.Content{
			&mcp.TextContent{Text: queryResult.String()},
		}
	}
	return
}
