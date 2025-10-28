package log

import (
	"encoding/json"
	"net/http"

	usecase "github.com/rubensantoniorosa2704/LoggingSSE/internal/application/log"
	"github.com/rubensantoniorosa2704/LoggingSSE/internal/application/log/dto"
)

type LogController struct {
	Usecase usecase.LogUsecaseInterface
}

func NewLogController(uc usecase.LogUsecaseInterface) *LogController {
	return &LogController{
		Usecase: uc,
	}
}

// @Summary      Create a new log entry
// @Description  Creates a new application log entry associated with an application and user, recording the message and severity level.
// @Tags         Logs
// @Accept       json
// @Produce      json
// @Param        log  body  dto.CreateLogInput  true  "Log creation data including ApplicationID and UserID."
// @Success      201  {object} dto.CreateLogOutput
// @Failure      400  {string} string "Invalid request body format, or missing/invalid ApplicationID/UserID."
// @Failure      500  {string} string "An internal error occurred while processing the log."
// @Router       /logs [post]
func (c *LogController) CreateLogHandler(w http.ResponseWriter, r *http.Request) {
	var input dto.CreateLogInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body format."})
		return
	}

	output, err := c.Usecase.CreateLog(r.Context(), input)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "An internal error occurred while creating the log."})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(output)
}
