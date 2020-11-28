package api

import (
	"github.com/ehpalumbo/go-diff/domain"
	"github.com/gin-gonic/gin"
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

// GetRouter returns a ready-to-use Gin engine for this Application
func (app Application) GetRouter() *gin.Engine {
	router := gin.Default()

	diff := router.Group("/v1/diff")

	// POST endpoint to upload sides to diff
	diff.POST("/:id/:side", app.saveSide)

	// GET endpoint to get diff results
	diff.GET("/:id", app.getReport)

	return router
}

func (app Application) saveSide(ctx *gin.Context) {
	id := ctx.Param("id")

	// check side is valid
	side, err := domain.ParseDiffSide(ctx.Param("side"))
	if err != nil {
		ctx.Status(404)
		return
	}

	// parse request body
	var requestBody PayloadRequestBody
	err = ctx.BindJSON(&requestBody)
	if err != nil {
		ctx.JSON(400, &ErrorResponseBody{id, "invalid body", err.Error()})
		return
	}

	// save side data
	payload := domain.DiffPayload{
		ID:    id,
		Side:  side,
		Value: requestBody.Data,
	}
	err = app.service.Save(payload)
	if err != nil {
		ctx.JSON(500, &ErrorResponseBody{id, "save operation failed", err.Error()})
		return
	}

	ctx.Status(204)
}

func (app Application) getReport(ctx *gin.Context) {
	id := ctx.Param("id")

	report, err := app.service.GetDiffReport(id)

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
		ctx.JSON(status, &ErrorResponseBody{id, message, err.Error()})
	} else {
		ctx.JSON(200, toDiffReportResponseBody(&report))
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
