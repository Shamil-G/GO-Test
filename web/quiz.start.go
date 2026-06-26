package web

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	mdw "gusseynov/GO-Quiz/middleware"
	"gusseynov/GO-Quiz/service/i18n"
	"gusseynov/GO-Quiz/storage"
)

type ViewStartPage struct {
	// 💡 Встраиваем базовый контекст (подтянет FIO, DepName, Lang, Theme)
	*mdw.BasePageContext
	QuizName        string
	IDQuiz          int
	Active          string
	DurationMinutes int
	QuestionsCount  int
	PassingScore    int
}

// QuizStartGet отображает стартовую страницу теста с данными из Oracle
func QuizStartPagetGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем контекст пользователя (ФИО, Департамент, Язык, Тема), созданный middleware
		pageCtx := mdw.GetOrCreatePageCtx(r.Context())

		// Берём ID теста из query-параметра: /quiz/start?id=12
		quizIDStr := r.URL.Query().Get("id")
		quizID, err := strconv.Atoi(quizIDStr)
		if err != nil {
			slog.Error("Неверный или отсутствующий ID теста", "IP", pageCtx.IP, "FIO", pageCtx.FIO, "DepName", pageCtx.DepName)
			http.Error(w, "Неверный или отсутствующий ID теста (?id=)", http.StatusBadRequest)
			return
		}

		// Вытягиваем живые данные теста из базы данных Oracle
		quiz, err := storage.GetQuizByID(r.Context(), quizID)
		if err != nil {
			slog.Error("Тест не найден или деактивирован в системе", "quizID", quizID, "IP", pageCtx.IP, "FIO", pageCtx.FIO, "DepName", pageCtx.DepName)
			http.Error(w, "Тест не найден или деактивирован в системе", http.StatusNotFound)
			return
		}

		// Выбираем локализацию названия в зависимости от выбранного языка пользователя
		quizName := quiz.NameRu
		if pageCtx.Lang == "kz" && quiz.NameKZ != "" {
			quizName = quiz.NameKZ
		}

		// Собираем монолитную структуру под ваш HTML-шаблон
		data := ViewStartPage{
			BasePageContext: pageCtx,
			IDQuiz:          quiz.IDQuiz,
			Active:          quiz.Active,
			DurationMinutes: quiz.DurationMinutes,
			QuestionsCount:  quiz.QuestionsCount,
			PassingScore:    quiz.PassingScore,
			QuizName:        quizName,
		}

		// Рендерим ваш HTML-шаблон
		tmpl, err := template.New("quiz_start.html").Funcs(
			template.FuncMap{
				"res_value": func(key string) string {
					return i18n.Get(pageCtx.Lang, key)
				},
			}).ParseFiles("templates/quiz_start.html")
		if err != nil {
			http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Ошибка рендеринга страницы: "+err.Error(), http.StatusInternalServerError)
		}
	}
}
