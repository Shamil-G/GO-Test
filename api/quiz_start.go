package api

import (
	"encoding/json"
	"net/http"

	"gusseynov/GO-Quiz/service"
)

func QuizStart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDQuiz int `json:"id_quiz"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	err := service.StartQuiz(r.Context(), req.IDQuiz)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"ok": true,
	})
}
