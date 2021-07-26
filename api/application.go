package api

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ehpalumbo/go-diff/domain"
)

// DiffService provides access to the service layer operations
type DiffService interface {
	Save(domain.DiffPayload) error
	GetDiffReport(string) (domain.DiffReport, error)
}

// Application is the entry point for starting this API
type Application struct {
	service DiffService
}

// NewApplication creates a new Application with the provided service dependency
func NewApplication(s DiffService) Application {
	return Application{s}
}

// Handle in the Lambda handler implementation
func (app Application) Handle(req events.APIGatewayProxyRequest) (res events.APIGatewayProxyResponse) {
	switch req.HTTPMethod {
	case "GET":
		res = app.getReport(req)
	case "POST":
		res = app.saveSide(req)
	default:
		res = events.APIGatewayProxyResponse{StatusCode: 404}
	}
	decorate(&res)
	return
}

func (app Application) saveSide(req events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	ID := req.PathParameters["id"]

	// check side is valid
	side, err := domain.ParseDiffSide(req.PathParameters["side"])
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
		}
	}

	// parse request body
	var requestBody PayloadRequestBody
	if err := json.Unmarshal([]byte(req.Body), &requestBody); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body: toJSON(&ErrorResponseBody{
				ID:     ID,
				Reason: "invalid body",
				Cause:  err.Error(),
			}),
		}
	}
	if len(requestBody.Data) == 0 {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body: toJSON(&ErrorResponseBody{
				ID:     ID,
				Reason: "missing data",
			}),
		}
	}

	// save side data
	payload := domain.DiffPayload{
		ID:    ID,
		Side:  side,
		Value: requestBody.Data,
	}
	err = app.service.Save(payload)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body: toJSON(&ErrorResponseBody{
				ID:     ID,
				Reason: "save operation failed",
				Cause:  err.Error(),
			}),
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 204,
	}
}

func (app Application) getReport(req events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	ID := req.PathParameters["id"]
	report, err := app.service.GetDiffReport(ID)
	if err != nil {
		var status int
		var message string
		if _, ok := err.(domain.DiffNotFoundError); ok {
			status = 404
			message = "diff not found"
		} else {
			status = 500
			message = "get diff failed"
		}
		return events.APIGatewayProxyResponse{
			StatusCode: status,
			Body: toJSON(&ErrorResponseBody{
				ID:     ID,
				Reason: message,
				Cause:  err.Error(),
			}),
		}
	} else {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       toJSON(toDiffReportResponseBody(&report)),
		}
	}
}

func toDiffReportResponseBody(report *domain.DiffReport) *DiffReportResponseBody {
	var insightResponses []DiffInsightResponse

	if len(report.Insights) > 0 {
		insightResponses = make([]DiffInsightResponse, len(report.Insights))
		for i, insight := range report.Insights {
			insightResponses[i] = DiffInsightResponse{
				Length: insight.Length,
				Offset: insight.Offset,
			}
		}
	}

	return &DiffReportResponseBody{
		Result:   report.Result.String(),
		Insights: insightResponses,
	}
}

func toJSON(v interface{}) string {
	buf, _ := json.Marshal(v)
	return string(buf)
}

func decorate(res *events.APIGatewayProxyResponse) {
	if res.Headers == nil {
		res.Headers = make(map[string]string)
	}
	res.Headers["Content-Type"] = "application/json"
}
