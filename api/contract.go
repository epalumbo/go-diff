package api

// PayloadRequestBody is the definition of the JSON request body for uploading side data
type PayloadRequestBody struct {
	Data string `json:"data"`
}

// DiffInsightResponse contains information about a single difference in the data
type DiffInsightResponse struct {
	Offset uint `json:"offset"`
	Length uint `json:"length"`
}

// DiffReportResponseBody contains information about differences in the data
type DiffReportResponseBody struct {
	Result   string                `json:"result"`
	Insights []DiffInsightResponse `json:"insights,omitempty"`
}

// ErrorResponseBody is the definition of JSON response body returned in case of errors
type ErrorResponseBody struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
	Cause  string `json:"cause"`
}
