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
package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/linuxsuren/api-testing/pkg/testing"
	"github.com/linuxsuren/atest-ext-store-orm/pkg"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

func newMCPCommand() (c *cobra.Command) {
	opt := &mcpOption{}
	c = &cobra.Command{
		Use:     "mcp",
		Short:   "Multi-Cluster-Platform related commands",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}
	flags := c.Flags()
	flags.StringVarP(&opt.mode, "mode", "", "http", "Server mode, one of http/stdio/sse")
	flags.IntVarP(&opt.port, "port", "", 7072, "Server port for http or sse mode")
	flags.StringVarP(&opt.url, "url", "", "", "Database URL")
	flags.StringVarP(&opt.username, "username", "", "", "Database username")
	flags.StringVarP(&opt.password, "password", "", "", "Database password")
	flags.StringVarP(&opt.database, "database", "", "", "Database name")
	flags.StringVarP(&opt.driver, "driver", "", "mysql", "Database driver, one of mysql/postgres/sqlite")
	return
}

type mcpOption struct {
	mode     string
	port     int
	url      string
	username string
	password string
	database string
	driver   string
}

func (o *mcpOption) preRunE(c *cobra.Command, args []string) (err error) {
	if o.url = getValueOrEnv(o.url, "DB_URL"); o.url == "" {
		err = fmt.Errorf("database url is required")
		return
	}
	o.username = getValueOrEnv(o.username, "DB_USERNAME")
	o.password = getValueOrEnv(o.password, "DB_PASSWORD")
	o.database = getValueOrEnv(o.database, "DB_DATABASE")
	o.driver = getValueOrEnv(o.driver, "DB_DRIVER")
	return
}

func getValueOrEnv(value, envKey string) (result string) {
	if value != "" {
		result = value
	} else {
		result = os.Getenv(envKey)
	}
	return
}

func (o *mcpOption) runE(c *cobra.Command, args []string) (err error) {
	opts := &mcp.ServerOptions{
		Instructions: "Database query mcp server",
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:  "database-mcp-server",
		Title: "ORM Database MCP Server",
	}, opts)

	store := &testing.Store{
		URL:      o.url,
		Username: o.username,
		Password: o.password,
		Properties: map[string]string{
			"database": o.database,
			"driver":   o.driver,
		},
	}

	dbServer := pkg.NewMcpServer(store)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "database-query",
		Description: "Query the database by SQL",
	}, dbServer.Query)

	switch o.mode {
	case "sse":
		handler := mcp.NewSSEHandler(func(request *http.Request) *mcp.Server {
			return server
		})
		c.Println("Starting SSE server on port:", o.port)
		err = http.ListenAndServe(fmt.Sprintf(":%d", o.port), handler)
	case "stdio":
		err = server.Run(c.Context(), &mcp.StdioTransport{})
	case "http":
		fallthrough
	default:
		handler := mcp.NewStreamableHTTPHandler(func(request *http.Request) *mcp.Server {
			return server
		}, nil)
		c.Println("Starting HTTP server on port:", o.port)
		err = http.ListenAndServe(fmt.Sprintf(":%d", o.port), handler)
	}
	return
}
