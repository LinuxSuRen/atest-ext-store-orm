/*
Copyright 2023-2025 API Testing Authors.

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
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
)

type dbserver struct {
	remote.UnimplementedLoaderServer
	defaultHistoryLimit int
}

// NewRemoteServer creates a remote server instance
func NewRemoteServer(defaultHistoryLimit int) (s remote.LoaderServer) {
	s = &dbserver{
		defaultHistoryLimit: defaultHistoryLimit,
	}
	return
}

func createDB(user, password, address, database, driver string) (db *gorm.DB, err error) {
	var dialector gorm.Dialector
	var dsn string
	switch driver {
	case "mysql", "":
		if !strings.Contains(address, ":") {
			address = fmt.Sprintf("%s:%d", address, 3306)
		}
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
	case "tdengine":
		dsn = fmt.Sprintf("%s:%s@ws(%s)/%s", user, password, address, database)
		dialector = NewTDengineDialector(dsn)
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

	if driver != "tdengine" {
		err = errors.Join(err, db.AutoMigrate(&TestCase{}))
		err = errors.Join(err, db.AutoMigrate(&TestSuite{}))
		err = errors.Join(err, db.AutoMigrate(&HistoryTestResult{}))
	}
	return
}

var dbCache = make(map[string]*gorm.DB)
var dbNameCache = make(map[string]string)

func (s *dbserver) getClientWithDatabase(ctx context.Context, dbName string) (db *gorm.DB, err error) {
	store := remote.GetStoreFromContext(ctx)
	if store == nil {
		err = errors.New("no connect to database")
	} else {
		database := dbName
		if database == "" {
			if v, ok := store.Properties["database"]; ok && v != "" {
				database = v
			}
		}

		driver := "mysql"
		if v, ok := store.Properties["driver"]; ok && v != "" {
			driver = v
		}
		log.Printf("get client from driver[%s] in database [%s]", driver, database)

		var ok bool
		if db, ok = dbCache[store.Name]; ok && db != nil && dbNameCache[store.Name] == database {
			return
		}

		if db, err = createDB(store.Username, store.Password, store.URL, database, driver); err == nil {
			dbCache[store.Name] = db
			dbNameCache[store.Name] = database
		}
	}
	return
}
func (s *dbserver) getClient(ctx context.Context) (db *gorm.DB, err error) {
	db, err = s.getClientWithDatabase(ctx, "")
	return
}

func (s *dbserver) ListTestSuite(ctx context.Context, _ *server.Empty) (suites *remote.TestSuites, err error) {
	items := make([]*TestSuite, 0)

	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}

	if err = db.Find(&items).Error; err == nil {
		suites = &remote.TestSuites{}
		for i := range items {
			suite := ConvertToGRPCTestSuite(items[i])
			suites.Data = append(suites.Data, suite)

			suite.Full = true
			if suiteWithCases, dErr := s.GetTestSuite(ctx, suite); dErr == nil {
				suites.Data[i] = suiteWithCases
			}
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

	err = db.Create(ConvertToDBTestSuite(testSuite)).Error
	return
}

const (
	nameQuery      = `name = ?`
	suiteNameQuery = "suite_name = ?"
	idQuery        = "id = ?"
)

func (s *dbserver) GetTestSuite(ctx context.Context, suite *remote.TestSuite) (reply *remote.TestSuite, err error) {
	query := &TestSuite{}
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}

	if err = db.Find(&query, nameQuery, suite.Name).Error; err == nil {
		reply = ConvertToGRPCTestSuite(query)
		if suite.Full {
			var testcases *server.TestCases
			if testcases, err = s.ListTestCases(ctx, &remote.TestSuite{
				Name: suite.Name,
			}); err == nil && testcases != nil {
				reply.Items = testcases.Data
			}
		}
	}
	return
}

func (s *dbserver) UpdateTestSuite(ctx context.Context, suite *remote.TestSuite) (reply *remote.TestSuite, err error) {
	reply = &remote.TestSuite{}
	input := ConvertToDBTestSuite(suite)
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}

	err = testSuiteIdentity(db, input).Updates(input).Error
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

	err = db.Transaction(func(tx *gorm.DB) (err error) {
		err = db.Delete(TestSuite{}, nameQuery, suite.Name).Error
		if err == nil {
			err = db.Delete(TestCase{}, suiteNameQuery, suite.Name).Error
		}
		return
	})
	return
}

func (s *dbserver) ListTestCases(ctx context.Context, suite *remote.TestSuite) (result *server.TestCases, err error) {
	items := make([]*TestCase, 0)
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}
	if err = db.Find(&items, suiteNameQuery, suite.Name).Error; err == nil {
		result = &server.TestCases{}
		for i := range items {
			result.Data = append(result.Data, ConvertToRemoteTestCase(items[i]))
		}
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
	err = db.Create(&payload).Error
	return
}

func (s *dbserver) CreateTestCaseHistory(ctx context.Context, historyTestResult *server.HistoryTestResult) (reply *server.Empty, err error) {
	reply = &server.Empty{}
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}

	store := remote.GetStoreFromContext(ctx)
	historyLimit := s.defaultHistoryLimit
	if v, ok := store.Properties["historyLimit"]; ok {
		if parsedHistoryLimit, parseErr := strconv.Atoi(v); parseErr == nil {
			historyLimit = parsedHistoryLimit
		} else {
			log.Printf("failed to parse history limit: %v\n", parseErr)
		}
	}

	var count int64
	if err = db.Model(&HistoryTestResult{}).Count(&count).Error; err != nil {
		return
	}

	if count >= int64(historyLimit) {
		var oldestRecord HistoryTestResult
		if err = db.Order("create_time").First(&oldestRecord).Error; err != nil {
			log.Printf("Error find oldest record: %v\n", err)
			return
		}

		if err = db.Delete(&oldestRecord).Error; err != nil {
			log.Printf("Error delete oldest record: %v\n", err)
			return
		}
		log.Printf("Existing count: %d, limit: %d\nmaximum number of entries reached.\n", count, historyLimit)
	}

	err = db.Create(ConvertToDBHistoryTestResult(historyTestResult)).Error
	return
}

func (s *dbserver) ListHistoryTestSuite(ctx context.Context, _ *server.Empty) (suites *remote.HistoryTestSuites, err error) {
	items := make([]*HistoryTestResult, 0)

	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}

	if err = db.Find(&items).Error; err != nil {
		return
	}

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
	if err = db.Find(&item, "suite_name = ? AND name = ?", testcase.SuiteName, testcase.Name).Error; err == nil {
		result = ConvertToRemoteTestCase(item)
	}
	return
}

func (s *dbserver) GetHistoryTestCaseWithResult(ctx context.Context, testcase *server.HistoryTestCase) (result *server.HistoryTestResult, err error) {
	item := &HistoryTestResult{}
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}
	if err = db.Find(&item, idQuery, testcase.ID).Error; err == nil {
		result = ConvertToRemoteHistoryTestResult(item)
	}
	return
}

func (s *dbserver) GetHistoryTestCase(ctx context.Context, testcase *server.HistoryTestCase) (result *server.HistoryTestCase, err error) {
	item := &HistoryTestResult{}
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}

	if err = db.Find(&item, idQuery, testcase.ID).Error; err == nil {
		result = ConvertToGRPCHistoryTestCase(item)
	}
	return
}

func (s *dbserver) GetTestCaseAllHistory(ctx context.Context, testcase *server.TestCase) (result *server.HistoryTestCases, err error) {
	items := make([]*HistoryTestResult, 0)
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}
	if err = db.Find(&items, "suite_name = ? AND case_name = ? ", testcase.SuiteName, testcase.Name).Error; err == nil {
		result = &server.HistoryTestCases{}
		for i := range items {
			result.Data = append(result.Data, ConvertToGRPCHistoryTestCase(items[i]))
		}
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
	if err = testCaseIdentity(db, input).Updates(input).Error; err != nil {
		return
	}

	data := make(map[string]interface{})
	if input.ExpectBody == "" {
		data["expect_body"] = ""
	}
	if input.ExpectSchema == "" {
		data["expect_schema"] = ""
	}

	if len(data) > 0 {
		err = testCaseIdentity(db, input).Updates(data).Error
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
	err = testCaseIdentity(db, input).Delete(input).Error
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
	var historyTestResult HistoryTestResult
	if err = historyTestCaseIdentity(db, input).Find(&historyTestResult).Error; err != nil {
		return nil, err
	}
	fileName := historyTestResult.Body
	if strings.HasPrefix(fileName, "isFilePath-") {
		tempDir := os.TempDir()
		fullFilePath := filepath.Join(tempDir, fileName)
		if err = os.Remove(fullFilePath); err != nil {
			log.Printf("Failed to delete file: %s, error: %v\n", fullFilePath, err)
			return
		}
	}
	err = db.Delete(&historyTestResult).Error
	return
}

func (s *dbserver) DeleteAllHistoryTestCase(ctx context.Context, historyTestCase *server.HistoryTestCase) (reply *server.Empty, err error) {
	reply = &server.Empty{}
	input := &HistoryTestResult{
		SuiteName: historyTestCase.SuiteName,
		CaseName:  historyTestCase.CaseName,
	}
	var db *gorm.DB
	if db, err = s.getClient(ctx); err != nil {
		return
	}

	var historyTestResults []HistoryTestResult
	if err = allHistoryTestCaseIdentity(db, input).Find(&historyTestResults).Error; err != nil {
		return nil, err
	}
	for _, historyTestResult := range historyTestResults {
		fileName := historyTestResult.Body
		if strings.HasPrefix(fileName, "isFilePath-") {
			tempDir := os.TempDir()
			fullFilePath := filepath.Join(tempDir, fileName)

			if err = os.Remove(fullFilePath); err != nil {
				log.Printf("Failed to delete file: %s, error: %v\n", fullFilePath, err)
				continue
			}
		}
		db.Delete(&historyTestResult)
	}
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

func (s *dbserver) GetVersion(context.Context, *server.Empty) (ver *server.Version, err error) {
	ver = &server.Version{
		Version: version.GetVersion(),
		Commit:  version.GetCommit(),
		Date:    version.GetDate(),
	}
	return
}

func (s *dbserver) PProf(ctx context.Context, in *server.PProfRequest) (data *server.PProfData, err error) {
	log.Println("pprof", in.Name)

	data = &server.PProfData{
		Data: extension.LoadPProf(in.Name),
	}
	return
}

func testCaseIdentity(db *gorm.DB, testcase *TestCase) *gorm.DB {
	return db.Model(testcase).Where(fmt.Sprintf("suite_name = '%s' AND name = '%s'", testcase.SuiteName, testcase.Name))
}

func historyTestCaseIdentity(db *gorm.DB, historyTestResult *HistoryTestResult) *gorm.DB {
	return db.Model(historyTestResult).Where(fmt.Sprintf("id = '%s'", historyTestResult.ID))
}

func allHistoryTestCaseIdentity(db *gorm.DB, historyTestResult *HistoryTestResult) *gorm.DB {
	return db.Model(historyTestResult).Where(fmt.Sprintf("suite_name = '%s' AND case_name = '%s'", historyTestResult.SuiteName, historyTestResult.CaseName))
}
