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
	"errors"
	"fmt"
	"github.com/linuxsuren/api-testing/pkg/extension"
	"github.com/linuxsuren/api-testing/pkg/server"
	"github.com/linuxsuren/api-testing/pkg/testing/remote"
	"github.com/linuxsuren/api-testing/pkg/util"
	"github.com/linuxsuren/api-testing/pkg/version"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"strings"
)

type dbserver struct {
	remote.UnimplementedLoaderServer
}

// NewRemoteServer creates a remote server instance
func NewRemoteServer() (s remote.LoaderServer) {
	s = &dbserver{}
	return
}

func createDB(user, password, address, database, driver string) (db *gorm.DB, err error) {
	var dialector gorm.Dialector
	var dsn string
	switch driver {
	case "mysql", "":
		dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true", user, password, address, database)
		dialector = mysql.Open(dsn)
	case "sqlite":
		dsn = fmt.Sprintf("%s.db", database)
		dialector = sqlite.Open(dsn)
	case "postgres":
		obj := strings.Split(address, ":")
		host, port := obj[0], "5432"
		if len(obj) > 1 {
			port = obj[1]
		}
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", host, user, password, database, port)
		dialector = postgres.Open(dsn)
	default:
		err = fmt.Errorf("invalid database driver %q", driver)
		return
	}

	log.Printf("try to connect to %q", dsn)
	db, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		err = fmt.Errorf("failed to connect to %q %v", dsn, err)
		return
	}

	db.AutoMigrate(&TestCase{})
	db.AutoMigrate(&TestSuite{})
	db.AutoMigrate(&HistoryTestResult{})
	return
}

var dbCache map[string]*gorm.DB = make(map[string]*gorm.DB)

func (s *dbserver) getClient(ctx context.Context) (db *gorm.DB, err error) {
	store := remote.GetStoreFromContext(ctx)
	if store == nil {
		err = errors.New("no connect to database")
	} else {
		var ok bool
		if db, ok = dbCache[store.Name]; ok && db != nil {
			return
		}

		database := "atest"
		driver := "mysql"
		if v, ok := store.Properties["database"]; ok && v != "" {
			database = v
		}
		if v, ok := store.Properties["driver"]; ok && v != "" {
			driver = v
		}

		if db, err = createDB(store.Username, store.Password, store.URL, database, driver); err == nil {
			dbCache[store.Name] = db
		}
	}
	return
}

func (s *dbserver) ListTestSuite(ctx context.Context, _ *server.Empty) (suites *remote.TestSuites, err error) {
	items := make([]*TestSuite, 0)

	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}

	db.Find(&items)
	suites = &remote.TestSuites{}
	for i := range items {
		suite := ConvertToGRPCTestSuite(items[i])
		suites.Data = append(suites.Data, suite)

		suite.Full = true
		if suiteWithCases, dErr := s.GetTestSuite(ctx, suite); dErr == nil {
			suites.Data[i] = suiteWithCases
		}
	}
	return
}

func (s *dbserver) CreateTestSuite(ctx context.Context, testSuite *remote.TestSuite) (reply *server.Empty, err error) {
	reply = &server.Empty{}
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}

	db.Create(ConvertToDBTestSuite(testSuite))
	return
}

const nameQuery = `name = ?`

func (s *dbserver) GetTestSuite(ctx context.Context, suite *remote.TestSuite) (reply *remote.TestSuite, err error) {
	query := &TestSuite{}
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}

	db.Find(&query, nameQuery, suite.Name)

	reply = ConvertToGRPCTestSuite(query)
	if suite.Full {
		var testcases *server.TestCases
		if testcases, err = s.ListTestCases(ctx, &remote.TestSuite{
			Name: suite.Name,
		}); err == nil && testcases != nil {
			reply.Items = testcases.Data
		}
	}
	return
}

func (s *dbserver) GetHistoryTestSuite(ctx context.Context, suite *remote.HistoryTestSuite) (reply *remote.HistoryTestSuite, err error) {
	query := &HistoryTestResult{}
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}

	db.Find(&query, nameQuery, suite.HistorySuiteName)

	reply = ConvertToGRPCHistoryTestSuite(query)
	return
}

func (s *dbserver) UpdateTestSuite(ctx context.Context, suite *remote.TestSuite) (reply *remote.TestSuite, err error) {
	reply = &remote.TestSuite{}
	input := ConvertToDBTestSuite(suite)
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}

	testSuiteIdentity(db, input).Updates(input)
	return
}

func testSuiteIdentity(db *gorm.DB, suite *TestSuite) *gorm.DB {
	return db.Model(suite).Where(nameQuery, suite.Name)
}

func (s *dbserver) DeleteTestSuite(ctx context.Context, suite *remote.TestSuite) (reply *server.Empty, err error) {
	reply = &server.Empty{}
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}

	err = db.Delete(TestSuite{}, nameQuery, suite.Name).Error
	return
}

func (s *dbserver) ListTestCases(ctx context.Context, suite *remote.TestSuite) (result *server.TestCases, err error) {
	items := make([]*TestCase, 0)
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}
	db.Find(&items, "suite_name = ?", suite.Name)

	result = &server.TestCases{}
	for i := range items {
		result.Data = append(result.Data, ConvertToRemoteTestCase(items[i]))
	}
	return
}

func (s *dbserver) CreateTestCase(ctx context.Context, testcase *server.TestCase) (reply *server.Empty, err error) {
	payload := ConverToDBTestCase(testcase)
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}
	reply = &server.Empty{}
	db.Create(&payload)
	return
}

func (s *dbserver) CreateTestCaseHistory(ctx context.Context, historyTestResult *server.HistoryTestResult) (reply *server.Empty, err error) {
	reply = &server.Empty{}
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}

	db.Create(ConvertToDBHistoryTestResult(historyTestResult))
	return
}

func (s *dbserver) ListHistoryTestSuite(ctx context.Context, _ *server.Empty) (suites *remote.HistoryTestSuites, err error) {
	items := make([]*HistoryTestResult, 0)

	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}

	db.Find(&items)

	groupedItems := make(map[string][]*HistoryTestResult)
	for _, item := range items {
		groupedItems[item.HistorySuiteName] = append(groupedItems[item.HistorySuiteName], item)
	}

	suites = &remote.HistoryTestSuites{}

	for historySuiteName, group := range groupedItems {
		suite := &remote.HistoryTestSuite{
			HistorySuiteName: historySuiteName,
		}
		for _, item := range group {
			converted := ConvertToGRPCHistoryTestSuite(item)
			suite.Items = append(suite.Items, converted.Items[0])
		}
		suites.Data = append(suites.Data, suite)
	}
	return
}

func (s *dbserver) GetTestCase(ctx context.Context, testcase *server.TestCase) (result *server.TestCase, err error) {
	item := &TestCase{}
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}
	db.Find(&item, "suite_name = ? AND name = ?", testcase.SuiteName, testcase.Name)

	result = ConvertToRemoteTestCase(item)
	return
}

func (s *dbserver) GetHistoryTestCaseWithResult(ctx context.Context, testcase *server.HistoryTestCase) (result *server.HistoryTestResult, err error) {
	item := &HistoryTestResult{}
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}
	db.Find(&item, "id = ? ", testcase.ID)

	result = ConvertToRemoteHistoryTestResult(item)
	return
}

func (s *dbserver) GetHistoryTestCase(ctx context.Context, testcase *server.HistoryTestCase) (result *server.HistoryTestCase, err error) {
	item := &HistoryTestResult{}
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}
	db.Find(&item, "id = ? ", testcase.ID)

	result = ConvertToGRPCHistoryTestCase(item)
	return
}

func (s *dbserver) GetTestCaseAllHistory(ctx context.Context, testcase *server.TestCase) (result *server.HistoryTestCases, err error) {
	items := make([]*HistoryTestResult, 0)
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}
	db.Find(&items, "suite_name = ? AND case_name = ? ", testcase.SuiteName, testcase.Name)

	result = &server.HistoryTestCases{}
	for i := range items {
		result.Data = append(result.Data, ConvertToGRPCHistoryTestCase(items[i]))
	}
	return
}

func (s *dbserver) UpdateTestCase(ctx context.Context, testcase *server.TestCase) (reply *server.TestCase, err error) {
	reply = &server.TestCase{}
	input := ConverToDBTestCase(testcase)
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}
	testCaseIdentiy(db, input).Updates(input)

	data := make(map[string]interface{})
	if input.ExpectBody == "" {
		data["expect_body"] = ""
	}
	if input.ExpectSchema == "" {
		data["expect_schema"] = ""
	}

	if len(data) > 0 {
		testCaseIdentiy(db, input).Updates(data)
	}
	return
}

func (s *dbserver) DeleteTestCase(ctx context.Context, testcase *server.TestCase) (reply *server.Empty, err error) {
	reply = &server.Empty{}
	input := ConverToDBTestCase(testcase)
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}
	testCaseIdentiy(db, input).Delete(input)
	return
}

func (s *dbserver) DeleteHistoryTestCase(ctx context.Context, historyTestCase *server.HistoryTestCase) (reply *server.Empty, err error) {
	reply = &server.Empty{}
	input := &HistoryTestResult{
		ID: historyTestCase.ID,
	}
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}
	historyTestCaseIdentiy(db, input).Delete(input)
	return
}

func (s *dbserver) Verify(ctx context.Context, in *server.Empty) (reply *server.ExtensionStatus, err error) {
	_, vErr := s.ListTestSuite(ctx, in)
	reply = &server.ExtensionStatus{
		Ready:   vErr == nil,
		Message: util.OKOrErrorMessage(vErr),
		Version: version.GetVersion(),
	}
	return
}

// func (s *dbserver) GetVersion(context.Context, *server.Empty) (ver *server.Version, err error) {
// 	ver = &server.Version{
// 		Version: version.GetVersion(),
// 		Commit:  version.GetCommit(),
// 		Date:    version.GetDate(),
// 	}
// 	return
// }

func (s *dbserver) PProf(ctx context.Context, in *server.PProfRequest) (data *server.PProfData, err error) {
	log.Println("pprof", in.Name)

	data = &server.PProfData{
		Data: extension.LoadPProf(in.Name),
	}
	return
}

func testCaseIdentiy(db *gorm.DB, testcase *TestCase) *gorm.DB {
	return db.Model(testcase).Where(fmt.Sprintf("suite_name = '%s' AND name = '%s'", testcase.SuiteName, testcase.Name))
}

func historyTestCaseIdentiy(db *gorm.DB, historyTestResult *HistoryTestResult) *gorm.DB {
	return db.Model(historyTestResult).Where(fmt.Sprintf("id = '%s'", historyTestResult.ID))
}
