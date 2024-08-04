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
	"encoding/json"
	"fmt"
	"time"

	"github.com/linuxsuren/api-testing/pkg/server"
	"github.com/linuxsuren/api-testing/pkg/testing/remote"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConverToDBTestCase(testcase *server.TestCase) (result *TestCase) {
	result = &TestCase{
		Name:      testcase.Name,
		SuiteName: testcase.SuiteName,
	}
	request := testcase.Request
	if request != nil {
		result.API = request.Api
		result.Method = request.Method
		result.Body = request.Body
		result.Header = pairToJSON(request.Header)
		result.Cookie = pairToJSON(request.Cookie)
		result.Form = pairToJSON(request.Form)
		result.Query = pairToJSON(request.Query)
	}

	resp := testcase.Response
	if resp != nil {
		result.ExpectBody = resp.Body
		result.ExpectSchema = resp.Schema
		result.ExpectStatusCode = int(resp.StatusCode)
		result.ExpectHeader = pairToJSON(resp.Header)
		result.ExpectBodyFields = pairToJSON(resp.BodyFieldsExpect)
		result.ExpectVerify = SliceToJSON(resp.Verify)
	}
	return
}

func ConvertToRemoteTestCase(testcase *TestCase) (result *server.TestCase) {
	result = &server.TestCase{
		Name: testcase.Name,

		Request: &server.Request{
			Api:    testcase.API,
			Method: testcase.Method,
			Body:   testcase.Body,
			Header: jsonToPair(testcase.Header),
			Cookie: jsonToPair(testcase.Cookie),
			Query:  jsonToPair(testcase.Query),
			Form:   jsonToPair(testcase.Form),
		},

		Response: &server.Response{
			StatusCode:       int32(testcase.ExpectStatusCode),
			Body:             testcase.ExpectBody,
			Schema:           testcase.ExpectSchema,
			Verify:           jsonToSlice(testcase.ExpectVerify),
			BodyFieldsExpect: jsonToPair(testcase.ExpectBodyFields),
			Header:           jsonToPair(testcase.ExpectHeader),
		},
	}
	return
}

func ConvertToDBTestSuite(suite *remote.TestSuite) (result *TestSuite) {
	result = &TestSuite{
		Name: suite.Name,
		API:  suite.Api,
	}
	if suite.Spec != nil {
		result.SpecKind = suite.Spec.Kind
		result.SpecURL = suite.Spec.Url
	}
	if suite.Param != nil {
		result.Param = pairToJSON(suite.Param)
	}
	return
}

func ConvertToDBHistoryTestResult(historyTestResult *server.HistoryTestResult) (result *HistoryTestResult) {
	id := fmt.Sprintf("%s_%s_%s", historyTestResult.CreateTime.AsTime().Local().Format("2006-01-02T15:04:05.999999999"), historyTestResult.Data.SuiteName, historyTestResult.Data.CaseName)
	result = &HistoryTestResult{
		ID:               id,
		Message:          historyTestResult.Message,
		Error:            historyTestResult.Error,
		HistorySuiteName: historyTestResult.CreateTime.AsTime().Local().Format("2006-1-2"),
		CaseName:         historyTestResult.Data.CaseName,
		SuiteName:        historyTestResult.Data.SuiteName,
		SuiteAPI:         historyTestResult.Data.SuiteApi,
		Param:            pairToJSON(historyTestResult.Data.SuiteParam),
		CreateTime:       historyTestResult.CreateTime.AsTime().Local().Format("2006-01-02T15:04:05.999999999"),
	}
	for _, testCase := range historyTestResult.TestCaseResult {
		result.StatusCode = int32(testCase.StatusCode)
		result.Output = testCase.Output
		result.Body = testCase.Body
	}

	request := historyTestResult.Data.Request
	if request != nil {
		result.CaseAPI = request.Api
		result.Method = request.Method
		result.Header = pairToJSON(request.Header)
		result.Cookie = pairToJSON(request.Cookie)
		result.Form = pairToJSON(request.Form)
		result.Query = pairToJSON(request.Query)
	}

	resp := historyTestResult.Data.Response
	if resp != nil {
		result.ExpectBody = resp.Body
		result.ExpectSchema = resp.Schema
		result.ExpectStatusCode = int(resp.StatusCode)
		result.ExpectHeader = pairToJSON(resp.Header)
		result.ExpectBodyFields = pairToJSON(resp.BodyFieldsExpect)
		result.ExpectVerify = SliceToJSON(resp.Verify)
	}

	if historyTestResult.Data.SuiteSpec != nil {
		result.SpecKind = historyTestResult.Data.SuiteSpec.Kind
		result.SpecURL = historyTestResult.Data.SuiteSpec.Url
	}

	return
}

func ConvertToRemoteHistoryTestResult(historyTestResult *HistoryTestResult) (result *server.HistoryTestResult) {
	createTime, err := time.Parse("2006-01-02T15:04:05.999999999", historyTestResult.CreateTime)
	if err != nil {
		fmt.Println("Error parsing time:", err)
	}
	result = &server.HistoryTestResult{
		Message:    historyTestResult.Message,
		Error:      historyTestResult.Error,
		CreateTime: timestamppb.New(createTime),
	}
	TestCaseResult := &server.TestCaseResult{
		StatusCode: historyTestResult.StatusCode,
		Body:       historyTestResult.Body,
		Output:     historyTestResult.Output,
		Error:      historyTestResult.Error,
		Header:     jsonToPair(historyTestResult.Header),
	}
	result.TestCaseResult = append(result.TestCaseResult, TestCaseResult)
	result.Data = ConvertToGRPCHistoryTestCase(historyTestResult)
	return
}

func ConvertToGRPCTestSuite(suite *TestSuite) (result *remote.TestSuite) {
	result = &remote.TestSuite{
		Name: suite.Name,
		Api:  suite.API,
		Spec: &server.APISpec{
			Kind: suite.SpecKind,
			Url:  suite.SpecURL,
		},
		Param: jsonToPair(suite.Param),
	}
	return
}

func ConvertToGRPCHistoryTestSuite(historyTestResult *HistoryTestResult) (result *remote.HistoryTestSuite) {
	result = &remote.HistoryTestSuite{
		HistorySuiteName: historyTestResult.HistorySuiteName,
	}

	item := ConvertToGRPCHistoryTestCase(historyTestResult)
	result.Items = append(result.Items, item)
	return
}

func ConvertToGRPCHistoryTestCase(historyTestResult *HistoryTestResult) (result *server.HistoryTestCase) {
	createTime, err := time.Parse("2006-01-02T15:04:05.999999999", historyTestResult.CreateTime)
	if err != nil {
		fmt.Println("Error parsing time:", err)
	}
	result = &server.HistoryTestCase{
		ID:               historyTestResult.ID,
		SuiteName:        historyTestResult.SuiteName,
		CaseName:         historyTestResult.CaseName,
		SuiteApi:         historyTestResult.SuiteAPI,
		SuiteParam:       jsonToPair(historyTestResult.Param),
		HistorySuiteName: historyTestResult.HistorySuiteName,
		CreateTime:       timestamppb.New(createTime),

		SuiteSpec: &server.APISpec{
			Kind: historyTestResult.SpecKind,
			Url:  historyTestResult.SpecURL,
		},

		Request: &server.Request{
			Api:    historyTestResult.CaseAPI,
			Method: historyTestResult.Method,
			Body:   historyTestResult.Body,
			Header: jsonToPair(historyTestResult.Header),
			Cookie: jsonToPair(historyTestResult.Cookie),
			Query:  jsonToPair(historyTestResult.Query),
			Form:   jsonToPair(historyTestResult.Form),
		},

		Response: &server.Response{
			StatusCode:       int32(historyTestResult.ExpectStatusCode),
			Body:             historyTestResult.ExpectBody,
			Schema:           historyTestResult.ExpectSchema,
			Verify:           jsonToSlice(historyTestResult.ExpectVerify),
			BodyFieldsExpect: jsonToPair(historyTestResult.ExpectBodyFields),
			Header:           jsonToPair(historyTestResult.ExpectHeader),
		},
	}
	return
}

func SliceToJSON(slice []string) (result string) {
	var data []byte
	var err error
	if slice != nil {
		if data, err = json.Marshal(slice); err == nil {
			result = string(data)
		}
	}
	if result == "" {
		result = "[]"
	}
	return
}

func pairToJSON(pair []*server.Pair) (result string) {
	var obj = make(map[string]string)
	for i := range pair {
		k := pair[i].Key
		v := pair[i].Value
		obj[k] = v
	}

	var data []byte
	var err error
	if data, err = json.Marshal(obj); err == nil {
		result = string(data)
	}
	return
}

func jsonToPair(jsonStr string) (pairs []*server.Pair) {
	pairMap := make(map[string]string, 0)
	err := json.Unmarshal([]byte(jsonStr), &pairMap)
	if err == nil {
		for k, v := range pairMap {
			pairs = append(pairs, &server.Pair{
				Key: k, Value: v,
			})
		}
	}
	return
}

func jsonToSlice(jsonStr string) (result []string) {
	_ = json.Unmarshal([]byte(jsonStr), &result)
	return
}
