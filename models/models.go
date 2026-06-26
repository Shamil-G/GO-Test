package models

import "time"

type Quiz struct {
	IDQuiz          int    `db:"ID_QUIZ"`
	Active          string `db:"ACTIVE"` // 'Y' или 'N'
	DurationMinutes int    `db:"DURATION_MINUTES"`
	QuestionsCount  int    `db:"QUESTIONS_COUNT"`
	PassingScore    int    `db:"PASSING_SCORE"`
	NameRu          string `db:"NAME_RU"`
	NameKZ          string `db:"NAME_KZ"`
}

type TestSession struct {
	IDTest             int       `db:"ID_TEST"`
	IDQuiz             int       `db:"ID_QUIZ"`
	StartedAt          time.Time `db:"STARTED_AT"`
	ExpiresAt          time.Time `db:"EXPIRES_AT"`
	CurrentQuestionIdx int       `db:"CURRENT_QUESTION_IDX"`
	IsFinished         string    `db:"IS_FINISHED"` // 'Y' или 'N'
	FinalScore         int       `db:"FINAL_SCORE"`
	FIO                string    `db:"FIO"`
	DepName            string    `db:"DEP_NAME"`
}

// QuestionStep — данные самого вопроса для текущего шага
type QuestionStep struct {
	IDQuestion     int          `db:"ID_QUESTION" json:"id_question"`
	IDQuest        int          `db:"ID_QUEST" json:"id_quest"`
	QuestionTxt    string       `db:"QUESTION_TXT" json:"question_txt"`
	SortOrder      int          `db:"SORT_ORDER" json:"sort_order"`
	PhotoURL       string       `db:"PHOTO_URL" json:"photo_url"`
	SelectedAnswer int          `db:"SELECTED_ANSWER" json:"selected_answer"`
	Answers        []AnswerStep `db:"-" json:"answers"` // Сюда складываем варианты ответов для этого вопроса
}

type AnswerStep struct {
	IDAnswer   int    `db:"ID_ANSWER" json:"id_answer"`
	AnswerTxt  string `db:"ANSWER_TXT" json:"answer_txt"`
	OrderNum   int    `db:"ORDER_NUM" json:"order_num"`
	IsSelected bool   `db:"-" json:"is_selected"`
}
