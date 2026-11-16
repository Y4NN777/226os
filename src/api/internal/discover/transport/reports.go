package transport

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/discover/domain"
	"github.com/Y4NN777/226os/src/api/internal/discover/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

// CreateResourceErrorReport serves POST /resources/{resourceId}/error-reports. The reporter is the
// session subject; the response never carries a reporter field.
func (h *Handler) CreateResourceErrorReport(w http.ResponseWriter, r *http.Request, resourceID apigen.ResourceId) {
	var body apigen.ErrorReportCreate
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	req := service.ReportRequest{ResourceID: resourceID.String(), Kind: string(body.Kind), Body: body.Body}
	if body.SourceId != nil {
		req.SourceID = body.SourceId.String()
	}
	report, err := h.reports.Create(r.Context(), subject.From(r.Context()), req)
	if err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	out, err := reportJSON(report)
	if err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	writeStatusJSON(w, http.StatusCreated, out)
}

func reportJSON(report domain.ErrorReport) (apigen.ErrorReport, error) {
	id, err := uuid.Parse(report.ID)
	if err != nil {
		return apigen.ErrorReport{}, err
	}
	resourceID, err := uuid.Parse(report.ResourceID)
	if err != nil {
		return apigen.ErrorReport{}, err
	}
	return apigen.ErrorReport{
		Id: id, ResourceId: resourceID, Kind: apigen.ErrorReportKind(report.Kind),
		Status: apigen.ErrorReportStatus(report.Status), CreatedAt: report.CreatedAt.UTC(),
	}, nil
}
