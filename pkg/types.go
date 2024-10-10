package pkg

type TestCase struct {
	SuiteName string `json:"suiteName" gorm:"type:varchar(200);uniqueIndex:idx_name_and_suite_name"`
	Name      string `gorm:"type:varchar(200);uniqueIndex:idx_name_and_suite_name"`
	API       string
	Method    string
	Body      string
	Header    string
	Cookie    string
	Query     string
	Form      string

	ExpectStatusCode int
	ExpectBody       string
	ExpectSchema     string
	ExpectHeader     string
	ExpectBodyFields string
	ExpectVerify     string
}

type TestSuite struct {
	Name     string `gorm:"primaryKey"`
	API      string
	SpecKind string
	SpecURL  string
	Param    string
}

type HistoryTestResult struct {
	ID               string `gorm:"primaryKey"`
	HistorySuiteName string
	CreateTime       string

	//suite information
	SuiteName string `json:"suiteName"`
	SuiteAPI  string
	SpecKind  string
	SpecURL   string
	Param     string

	//case information
	CaseName      string `json:"caseName"`
	CaseAPI       string
	Method        string
	Body          string
	Header        string
	HistoryHeader string
	Cookie        string
	Query         string
	Form          string

	ExpectStatusCode int
	ExpectBody       string
	ExpectSchema     string
	ExpectHeader     string
	ExpectBodyFields string
	ExpectVerify     string

	//result information
	Message    string `json:"message"`
	Error      string `json:"error"`
	StatusCode int32  `json:"statusCode"`
	Output     string `json:"output"`
}
