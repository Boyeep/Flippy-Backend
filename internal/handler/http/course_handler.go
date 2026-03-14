package http

import (
	"net/http"

	"github.com/Boyeep/flippy-backend/internal/service"
)

type CourseHandler struct {
	service service.CourseService
}

func NewCourseHandler(service service.CourseService) CourseHandler {
	return CourseHandler{service: service}
}

func (h CourseHandler) List(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"data": h.service.List(),
	})
}
