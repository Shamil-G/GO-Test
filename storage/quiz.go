package storage

import (
	"context"
	"fmt"
	"log/slog"

	"gusseynov/GO-Quiz/metrics"
	"gusseynov/GO-Quiz/models"
	"time"
)

// GetQuizByID вытягивает данные теста из таблицы QUIZ по его ID
func GetQuizByID(ctx context.Context, idQuiz int) (models.Quiz, error) {
	var quiz models.Quiz

	// Пишем SQL-запрос для Oracle (в Oracle плейсхолдеры идут через двоеточие и порядковый номер :1)
	query := `
		SELECT 
			ID_QUIZ, 
			ACTIVE, 
			DURATION_MINUTES, 
			QUESTIONS_COUNT, 
			PASSING_SCORE, 
			NAME_RU, 
			NAME_KZ 
		FROM quiz 
		WHERE ID_QUIZ = :1 AND ACTIVE = 'Y'
	`

	start := time.Now()
	// Используем встроенный в sqlx метод GetContext, чтобы маппить db-теги
	err := DB.GetContext(ctx, &quiz, query, idQuiz)
	slog.Info("GetQuizByID", "idQuiz", idQuiz, "err", err)

	// Фиксируем метрики аналогично вашим хелперам
	elapsed := time.Since(start).Milliseconds()
	metrics.DBSelectDuration.WithLabelValues("GetQuizByID").Observe(float64(elapsed))
	metrics.DBSelectTotal.WithLabelValues("GetQuizByID").Inc()

	if err != nil {
		metrics.DBSelectErrors.WithLabelValues("GetQuizByID", err.Error()).Inc()
		return quiz, fmt.Errorf("ошибка получения теста ID %d из Oracle: %w", idQuiz, err)
	}

	return quiz, nil
}
