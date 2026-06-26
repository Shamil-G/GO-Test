package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"

	mdw "gusseynov/GO-Quiz/middleware"
	"gusseynov/GO-Quiz/service/i18n"
	"gusseynov/GO-Quiz/storage"
)

type QuizPageData struct {
	// Добавляем техническое поле, чтобы функция внутри шаблона знала, какой язык сейчас у пользователя
	Lang               string
	IDReg              int
	QuizName           string
	CurrentQuestionNum int
	RemainTimeStr      string
	QuestionNav        []NavItem
	Question           QuestionVD
	Answers            []AnswerVD
}

func QuizTesting() http.HandlerFunc {
	// Мы убираем глобальный парсинг при старте, чтобы исключить конфликты путей,
	// и будем парсить шаблон прямо внутри запроса (для веб-интерфейса это незаметно по скорости)
	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Получаем языковой контекст текущего пользователя
		ctx_service := r.Context()
		ctxPage := mdw.GetOrCreatePageCtx(ctx_service)
		userIP := ctxPage.IP

		// 2. СОЗДАЕМ И ПАРСИМ ШАБЛОН ПРЯМО ТУТ (Потокобезопасно и динамически)
		tmplPath := filepath.Join("templates", "testing.html")
		tmpl, err := template.New("testing.html").Funcs(
			template.FuncMap{
				"res_value": func(key string) string {
					// ИСПОЛЬЗУЕМ КЛЮЧ НАПРЯМУЮ: Теперь ctxPage доступен без трюков с Clone!
					return i18n.Get(ctxPage.Lang, key)
				},
			}).ParseFiles(tmplPath)

		// Если файл не найден или в HTML ошибка синтаксиса — мы увидим это НА ЭКРАНЕ
		if err != nil {
			slog.Error("Ошибка загрузки или парсинга файла шаблона", "path", tmplPath, "err", err)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "❌ ОШИБКА ИНИЦИАЛИЗАЦИИ ШАБЛОНА:\nПуть: %s\nОшибка: %v\n\nПроверьте, существует ли файл по этому пути относительно места запуска приложения.", tmplPath, err)
			return
		}

		// 3. ИДЕМ В БАЗУ ORACLE ЗА JSON-ДАННЫМИ ВОПРОСА
		var jsonResult string
		query := `SELECT test.question(:1) FROM dual`

		err = storage.DBSelectOne(ctx_service, "GetQuestionJSON", &jsonResult, query, userIP)
		if err != nil {
			slog.Error("Ошибка вызова функции test.question из Oracle", "ip", userIP, "err", err)
			http.Error(w, "Ошибка получения данных из БД", http.StatusInternalServerError)
			return
		}

		// 4. ДЕКОДИРУЕМ JSON ИЗ ORACLE
		var oracleQuestion QuestionVD
		if err := json.Unmarshal([]byte(jsonResult), &oracleQuestion); err != nil {
			slog.Error("Ошибка парсинга JSON от Oracle", "json", jsonResult, "err", err)
			http.Error(w, "Ошибка обработки данных вопроса", http.StatusInternalServerError)
			return
		}

		// ====================================================================
		// ЗДЕСЬ ИДЕМ В БАЗУ ЗА ПЛИТКОЙ НАВИГАЦИИ (QuestionNav)
		// ====================================================================
		var jsonNavResult string
		queryNav := `SELECT test.question_nav(:1) FROM dual`

		err = storage.DBSelectOne(ctx_service, "GetQuestionNavJSON", &jsonNavResult, queryNav, userIP)
		if err != nil {
			slog.Error("Ошибка вызова функции test.question_nav из Oracle", "ip", userIP, "err", err)
			http.Error(w, "Ошибка получения навигации из БД", http.StatusInternalServerError)
			return
		}
		// ПРЯМО ПЕРЕД json.Unmarshal:
		if jsonNavResult == "" || jsonNavResult == "[]" {
			slog.Warn("Oracle вернул пустой массив навигации", "json", jsonNavResult)
		} else {
			// Выведем в консоль, чтобы глазами увидеть битый JSON или ошибку СУБД
			fmt.Printf("\n=== НАВИГАЦИЯ ИЗ ORACLE ===\n%s\n===========================\n", jsonNavResult)
		}

		var oracleNav []NavItem
		if err := json.Unmarshal([]byte(jsonNavResult), &oracleNav); err != nil {
			slog.Error("Ошибка парсинга JSON навигации от Oracle", "json", jsonNavResult, "err", err)
			http.Error(w, "Ошибка обработки навигационной панели", http.StatusInternalServerError)
			return
		}

		// 5. НАПОЛНЯЕМ ДАННЫЕ СТРАНИЦЫ
		// 5. НАПОЛНЯЕМ ДАННЫЕ СТРАНИЦЫ
		data := QuizPageData{
			Lang:               ctxPage.Lang,
			IDReg:              123, // В будущем вытащим из сессии registrations
			QuizName:           "Тестирование сотрудников",
			CurrentQuestionNum: 2, // В будущем вытащим из r.current_question
			RemainTimeStr:      "25:00",

			// ИСПРАВЛЕНИЕ: Передаем реальный массив из базы вместо захардкоженного
			QuestionNav: oracleNav,

			Question: oracleQuestion,
			Answers:  oracleQuestion.Answers,
		}

		if data.Answers == nil {
			data.Answers = []AnswerVD{}
		}

		// 6. РЕНДЕРИМ СТРАНИЦУ ЧЕРЕЗ БУФЕР
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			slog.Error("Ошибка сборки (Execute) HTML страницы", "err", err)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "❌ ОШИБКА ВНУТРИ HTML ШАБЛОНА (Execute):\n%v", err)
			return
		}

		// 7. ОТПРАВЛЯЕМ СФОРМИРОВАННЫЙ HTML В СЕТЬ
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
		w.WriteHeader(http.StatusOK)
		buf.WriteTo(w)
	}
}
