// cmd/router.go

package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	apiH "gusseynov/GO-Quiz/api"
	mdw "gusseynov/GO-Quiz/middleware"
	ssoPkg "gusseynov/GO-Quiz/sso"
	webH "gusseynov/GO-Quiz/web"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Router() http.Handler {
	r := chi.NewRouter()

	// Мидлвары
	// r.Use(mdw.Authorize)
	// r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// r.Use(mdw.SlogLogger)
	r.Use(mdw.Metrics)

	r.Handle("/metrics", promhttp.Handler())
	// Health
	r.Post("/ping", ssoPkg.Alive())
	// Статика
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Настройки пользователя
	// -----------------------------
	// WEB UI (внутренний интерфейс)
	// -----------------------------
	// WEB UI (Рендеринг полных HTML-страниц)
	r.Group(func(r chi.Router) {
		r.Use(mdw.PageContext)
		// r.Get("/", webH.ViewRootGet())
		// r.Get("/language/{lang}", webH.ChangeLangHandler())
		// r.Get("/theme/{theme}", webH.ChangeThemeHandler())

		// Три главные страницы
		r.Get("/quiz/start", webH.QuizStartPagetGet())  // Шаблон стартового экрана
		r.Get("/quiz/process", webH.QuizTestingGet())   // Шаблон самого процесса (один на весь тест)
		r.Post("/quiz/process", webH.QuizTestingPost()) // Принимает ответ, двигает очередь в Oracle и снова показывает вопрос
		// r.Get("/quiz/result", webH.QuizResultGet())   // Шаблон результатов
	})

	// API (JSON)
	r.Route("/api", func(api chi.Router) {
		api.Use(mdw.ApiContext)
		api.Post("/start", apiH.QuizStart)
		// api.Get("/question", apiH.GetQuestion)
		// api.Post("/answer", apiH.SubmitAnswer)
		// api.Post("/finish", apiH.FinishQuiz)
	})

	return r
}
