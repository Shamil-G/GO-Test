package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"

	mdw "gusseynov/GO-Quiz/middleware"
	"gusseynov/GO-Quiz/service/i18n"
	"gusseynov/GO-Quiz/storage"
)

// Структуры для навигационных кнопок хедера
type NavItem struct {
	Num        int `json:"num"`
	IsAnswered int `json:"is_answered"` // ИСПРАВЛЕНО: теперь int вместо bool
	IsCurrent  int `json:"is_current"`  // ИСПРАВЛЕНО: теперь int вместо bool
}

type AnswerVD struct {
	IDAnswer int    `json:"id_answer"`
	Text     string `json:"text"`
}

type QuestionVD struct {
	IDQuest        int        `json:"id_quest"`
	Text           string     `json:"text"`
	ImageURL       string     `json:"image_url"`
	Answers        []AnswerVD `json:"answers"`
	SelectedAnswer int        `json:"selected_answer"`
	EndUnix        int        `json:"expires_at"`
}

// ====================================================================
// 1. ХЕНДЛЕР GET: ОТОБРАЖЕНИЕ ТЕКУЩЕГО ВОПРОСА
// ====================================================================
func QuizTestingGet() http.HandlerFunc {
	tmplPath := filepath.Join("templates", "testing.html")
	tmpl, err := template.New("testing.html").Funcs(
		template.FuncMap{
			"res_value": func(key string) string { return "" },
		}).ParseFiles(tmplPath)

	if err != nil {
		slog.Error("Критическая ошибка компиляции базового шаблона", "err", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if tmpl == nil {
			http.Error(w, "Шаблон не инициализирован", http.StatusInternalServerError)
			return
		}

		ctx_service := r.Context()
		ctxPage := mdw.GetOrCreatePageCtx(ctx_service)
		userIP := ctxPage.IP

		// 1. Запрашиваем из Oracle данные текущего вопроса в формате JSON
		var jsonResult string
		queryQuestion := `SELECT test.question(:1) FROM dual`
		err = storage.DBSelectOne(ctx_service, "GetQuestionJSON", &jsonResult, queryQuestion, userIP)
		if err != nil {
			slog.Error("Ошибка получения вопроса из Oracle", "ip", userIP, "err", err)
			http.Error(w, "Ошибка получения данных из БД", http.StatusInternalServerError)
			return
		}

		var oracleQuestion QuestionVD
		_ = json.Unmarshal([]byte(jsonResult), &oracleQuestion)

		// 2. Запрашиваем из Oracle массив плитки навигации хедера
		var jsonNavResult string
		queryNav := `SELECT test.question_nav(:1) FROM dual`
		err = storage.DBSelectOne(ctx_service, "GetQuestionNavJSON", &jsonNavResult, queryNav, userIP)
		if err != nil {
			slog.Error("Ошибка получения навигации из Oracle", "ip", userIP, "err", err)
			http.Error(w, "Ошибка получения навигации из БД", http.StatusInternalServerError)
			return
		}

		var oracleNav []NavItem
		_ = json.Unmarshal([]byte(jsonNavResult), &oracleNav)

		// 3. Формируем итоговые данные для передачи в HTML
		data := QuizPageData{
			Lang:               ctxPage.Lang,
			IDReg:              123,
			QuizName:           "Тестирование сотрудников",
			CurrentQuestionNum: 2,
			RemainTimeStr:      "25:00",
			QuestionNav:        oracleNav,
			Question:           oracleQuestion,
			Answers:            oracleQuestion.Answers,
		}

		if data.Answers == nil {
			data.Answers = []AnswerVD{}
		}

		// Динамически изолируем локализацию под язык конкретного запроса
		requestTmpl, _ := tmpl.Clone()
		requestTmpl.Funcs(template.FuncMap{
			"res_value": func(key string) string {
				return i18n.Get(ctxPage.Lang, key)
			},
		})

		var buf bytes.Buffer
		if err := requestTmpl.Execute(&buf, data); err != nil {
			slog.Error("Ошибка сборки HTML страницы", "err", err)
			http.Error(w, "Ошибка генерации интерфейса", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
		w.WriteHeader(http.StatusOK)
		buf.WriteTo(w)
	}
}

// ====================================================================
// 2. ХЕНДЛЕР POST: ПРИЕМ ОТВЕТА И СДВИГ ОЧЕРЕДИ В ORACLE
// ====================================================================
func QuizTestingPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx_service := r.Context()
		ctxPage := mdw.GetOrCreatePageCtx(ctx_service)
		userIP := ctxPage.IP

		// Читаем данные из отправленной HTML-формы
		if err := r.ParseForm(); err != nil {
			slog.Error("Ошибка парсинга POST-формы", "err", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		idQuestRaw := r.FormValue("id_quest")
		answerIdRaw := r.FormValue("answer_id")
		moveAction := r.FormValue("move") // Команды: "next", "prev", "goto_3" и т.д.

		idQuest, _ := strconv.Atoi(idQuestRaw)
		answerID, _ := strconv.Atoi(answerIdRaw)

		slog.Info("Навигация", "IP", userIP, "idQuest", idQuest, "answerID", answerID, "moveAction", moveAction)

		// Кликнули кнопку "Сохранить" -> остаемся на текущем вопросе
		if moveAction == "" {
			moveAction = "stay"
		}

		// Дергаем вашу процедуру в Oracle: она сохраняет ответ и меняет current_question в сессии
		queryNavUpdate := `BEGIN test.navigate_quiz(:1, :2, :3, :4); END;`
		errUpdate := storage.DBExecNamed(ctx_service, queryNavUpdate,
			"test.navigate_quiz",
			userIP,
			idQuest,
			answerID,
			moveAction,
		)

		if errUpdate != nil {
			slog.Error("Ошибка вызова процедуры навигации в Oracle", "ip", userIP, "err", errUpdate)
			http.Error(w, "Ошибка сохранения данных в СУБД", http.StatusInternalServerError)
			return
		}

		// КЛЮЧЕВОЙ МОМЕНТ: Перенаправляем пользователя обратно на GET-эндпоинт страницы.
		// Браузер мгновенно выполнит GET-запрос, и функция QuizTesting()
		// отобразит уже новый, измененный базой данных вопрос!
		http.Redirect(w, r, "/quiz/process", http.StatusSeeOther)
	}
}
