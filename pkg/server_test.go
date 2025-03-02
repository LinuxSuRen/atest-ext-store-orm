/*
Copyright 2023-2024 API Testing Authors.

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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/linuxsuren/api-testing/pkg/server"
	atest "github.com/linuxsuren/api-testing/pkg/testing"
	"github.com/linuxsuren/api-testing/pkg/testing/remote"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewRemoteServer(t *testing.T) {
	remoteServer := NewRemoteServer(10)
	assert.NotNil(t, remoteServer)
	defaultCtx := context.Background()

	t.Run("ListTestSuite", func(t *testing.T) {
		_, err := remoteServer.ListTestSuite(defaultCtx, nil)
		assert.Error(t, err)
	})

	t.Run("CreateTestSuite", func(t *testing.T) {
		_, err := remoteServer.CreateTestSuite(defaultCtx, nil)
		assert.Error(t, err)
	})

	t.Run("GetTestSuite", func(t *testing.T) {
		_, err := remoteServer.GetTestSuite(defaultCtx, nil)
		assert.Error(t, err)
	})

	t.Run("UpdateTestSuite", func(t *testing.T) {
		_, err := remoteServer.UpdateTestSuite(defaultCtx, &remote.TestSuite{})
		assert.Error(t, err)
	})

	t.Run("DeleteTestSuite", func(t *testing.T) {
		_, err := remoteServer.DeleteTestSuite(defaultCtx, nil)
		assert.Error(t, err)
	})

	t.Run("ListTestCases", func(t *testing.T) {
		_, err := remoteServer.ListTestCases(defaultCtx, nil)
		assert.Error(t, err)
	})

	t.Run("CreateTestCase", func(t *testing.T) {
		_, err := remoteServer.CreateTestCase(defaultCtx, &server.TestCase{})
		assert.Error(t, err)
	})

	t.Run("GetTestCase", func(t *testing.T) {
		_, err := remoteServer.GetTestCase(defaultCtx, nil)
		assert.Error(t, err)
	})

	t.Run("UpdateTestCase", func(t *testing.T) {
		_, err := remoteServer.UpdateTestCase(defaultCtx, &server.TestCase{})
		assert.Error(t, err)
	})

	t.Run("DeleteTestCase", func(t *testing.T) {
		_, err := remoteServer.DeleteTestCase(defaultCtx, &server.TestCase{})
		assert.Error(t, err)
	})

	t.Run("Verify", func(t *testing.T) {
		reply, err := remoteServer.Verify(defaultCtx, nil)
		assert.NoError(t, err)
		assert.False(t, reply.Ready)
	})

	t.Run("CreateTestCaseHistory", func(t *testing.T) {
		_, err := remoteServer.CreateTestCaseHistory(defaultCtx, nil)
		assert.Error(t, err)
	})

	t.Run("ListHistoryTestSuite", func(t *testing.T) {
		_, err := remoteServer.ListHistoryTestSuite(defaultCtx, nil)
		assert.Error(t, err)
	})

	t.Run("GetHistoryTestCase", func(t *testing.T) {
		_, err := remoteServer.GetHistoryTestCase(defaultCtx, nil)
		assert.Error(t, err)
	})

	t.Run("GetTestCaseAllHistory", func(t *testing.T) {
		_, err := remoteServer.GetTestCaseAllHistory(defaultCtx, nil)
		assert.Error(t, err)
	})

	t.Run("DeleteHistoryTestCase", func(t *testing.T) {
		_, err := remoteServer.DeleteHistoryTestCase(defaultCtx, &server.HistoryTestCase{})
		assert.Error(t, err)
	})

	t.Run("DeleteAllHistoryTestCase", func(t *testing.T) {
		_, err := remoteServer.DeleteAllHistoryTestCase(defaultCtx, &server.HistoryTestCase{})
		assert.Error(t, err)
	})

	t.Run("invalid orm driver", func(t *testing.T) {
		remoteServer := NewRemoteServer(10)
		assert.NotNil(t, remoteServer)
		defaultCtx := remote.WithIncomingStoreContext(context.TODO(), &atest.Store{
			Properties: map[string]string{
				"driver": "invalid",
			},
		})
		_, err := remoteServer.ListTestSuite(defaultCtx, &server.Empty{})
		assert.Error(t, err)
	})

	t.Run("invalid mysql config", func(t *testing.T) {
		remoteServer := NewRemoteServer(10)
		assert.NotNil(t, remoteServer)
		defaultCtx := remote.WithIncomingStoreContext(context.TODO(), &atest.Store{
			Properties: map[string]string{
				"driver": "mysql",
			},
		})
		_, err := remoteServer.ListTestSuite(defaultCtx, &server.Empty{})
		assert.Error(t, err)
	})

	t.Run("invalid postgres config", func(t *testing.T) {
		remoteServer := NewRemoteServer(10)
		assert.NotNil(t, remoteServer)
		defaultCtx := remote.WithIncomingStoreContext(context.TODO(), &atest.Store{
			Properties: map[string]string{
				"driver": "postgres",
			},
			URL: "0.0.0.0:-123",
		})
		_, err := remoteServer.ListTestSuite(defaultCtx, &server.Empty{})
		assert.Error(t, err)
	})
}

func TestSQLite(t *testing.T) {
	remoteServer := NewRemoteServer(10)
	assert.NotNil(t, remoteServer)
	defaultCtx := remote.WithIncomingStoreContext(context.TODO(), &atest.Store{
		Properties: map[string]string{
			"driver":   "sqlite",
			"database": "atest",
		},
	})
	defer func() {
		_ = os.Remove("atest.db")
	}()

	t.Run("CreateTestSuite", func(t *testing.T) {
		_, err := remoteServer.CreateTestSuite(defaultCtx, &remote.TestSuite{
			Name: "test",
		})
		assert.NoError(t, err)
	})

	t.Run("ListTestSuite", func(t *testing.T) {
		result, err := remoteServer.ListTestSuite(defaultCtx, nil)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(result.Data))
		assert.Equal(t, "test", result.Data[0].Name)
	})

	t.Run("UpdateTestSuite", func(t *testing.T) {
		_, err := remoteServer.UpdateTestSuite(defaultCtx, &remote.TestSuite{
			Name: "test",
			Api:  "fake",
		})
		assert.NoError(t, err)

		var suite *remote.TestSuite
		suite, err = remoteServer.GetTestSuite(defaultCtx, &remote.TestSuite{
			Name: "test",
		})
		assert.NoError(t, err)
		assert.Equal(t, "fake", suite.Api)
	})

	t.Run("CreateTestCase", func(t *testing.T) {
		_, err := remoteServer.CreateTestCase(defaultCtx, &server.TestCase{
			SuiteName: "test",
			Name:      "test",
		})
		assert.NoError(t, err)

		var testcases *server.TestCases
		testcases, err = remoteServer.ListTestCases(defaultCtx, &remote.TestSuite{
			Name: "test",
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(testcases.Data))
	})

	t.Run("UpdateTestCase", func(t *testing.T) {
		_, err := remoteServer.UpdateTestCase(defaultCtx, &server.TestCase{
			SuiteName: "test",
			Name:      "test",
			Request: &server.Request{
				Api: "api",
			},
		})
		assert.NoError(t, err)

		var testcase *server.TestCase
		testcase, err = remoteServer.GetTestCase(defaultCtx, &server.TestCase{
			SuiteName: "test",
			Name:      "test",
		})
		assert.NoError(t, err)
		assert.Equal(t, "api", testcase.Request.Api)
	})

	t.Run("DeleteTestCase", func(t *testing.T) {
		_, err := remoteServer.DeleteTestCase(defaultCtx, &server.TestCase{
			SuiteName: "test",
			Name:      "test",
		})
		assert.NoError(t, err)
	})

	t.Run("DeleteTestSuite", func(t *testing.T) {
		_, err := remoteServer.DeleteTestSuite(defaultCtx, &remote.TestSuite{
			Name: "test",
		})
		assert.NoError(t, err)
	})

	t.Run("PProf", func(t *testing.T) {
		_, err := remoteServer.PProf(defaultCtx, &server.PProfRequest{})
		assert.NoError(t, err)
	})

	t.Run("GetVersion", func(t *testing.T) {
		_, err := remoteServer.GetVersion(defaultCtx, &server.Empty{})
		assert.NoError(t, err)
	})

	now := time.Now()
	t.Run("CreateTestCaseHistory", func(t *testing.T) {
		_, err := remoteServer.CreateTestCaseHistory(defaultCtx, &server.HistoryTestResult{
			CreateTime: timestamppb.New(now),
			Data: &server.HistoryTestCase{
				CaseName:  "test",
				SuiteName: "test",
			},
			TestCaseResult: []*server.TestCaseResult{
				{Output: "test output"},
			},
		})
		assert.NoError(t, err)
	})
	id := fmt.Sprintf("%s_test_test", now.Local().Format("2006-01-02T15:04:05.999999999"))

	t.Run("ListHistoryTestSuite", func(t *testing.T) {
		result, err := remoteServer.ListHistoryTestSuite(defaultCtx, nil)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(result.Data))
		assert.Equal(t, "test", result.Data[0].Items[0].SuiteName)
	})

	t.Run("GetHistoryTestCase", func(t *testing.T) {
		result, err := remoteServer.GetHistoryTestCase(defaultCtx, &server.HistoryTestCase{
			ID: id,
		})
		assert.NoError(t, err)
		assert.Equal(t, "test", result.CaseName)
	})

	t.Run("GetHistoryTestCaseWithResult", func(t *testing.T) {
		result, err := remoteServer.GetHistoryTestCaseWithResult(defaultCtx, &server.HistoryTestCase{
			ID: id,
		})
		assert.NoError(t, err)
		assert.Equal(t, "test", result.Data.CaseName)
		assert.Equal(t, "test output", result.TestCaseResult[0].Output)
	})

	t.Run("GetTestCaseAllHistory", func(t *testing.T) {
		result, err := remoteServer.GetTestCaseAllHistory(defaultCtx, &server.TestCase{
			Name:      "test",
			SuiteName: "test",
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(result.Data))
	})

	t.Run("DeleteHistoryTestCase", func(t *testing.T) {
		_, err := remoteServer.DeleteHistoryTestCase(defaultCtx, &server.HistoryTestCase{
			ID: id,
		})
		assert.NoError(t, err)
	})

	t.Run("DeleteAllHistoryTestCase", func(t *testing.T) {
		_, err := remoteServer.DeleteAllHistoryTestCase(defaultCtx, &server.HistoryTestCase{
			CaseName:  "test",
			SuiteName: "test",
		})
		assert.NoError(t, err)
	})
}

func TestTdEngine(t *testing.T) {
	remoteServer := NewRemoteServer(10)
	assert.NotNil(t, remoteServer)
	defaultCtx := remote.WithIncomingStoreContext(context.TODO(), &atest.Store{
		URL:      "127.0.0.1:6041",
		Username: "root",
		Password: "taosdata",
		Properties: map[string]string{
			"driver": "tdengine",
		},
	})
	_, err := remoteServer.Query(defaultCtx, &server.DataQuery{})
	assert.Error(t, err)
}
