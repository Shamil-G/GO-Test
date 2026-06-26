-- ====================================================================
-- 1. ПОСЛЕДОВАТЕЛЬНОСТИ (SEQUENCES)
-- ====================================================================
CREATE SEQUENCE seq_quizzes START WITH 1 INCREMENT BY 1 NOCACHE;
CREATE SEQUENCE seq_questions_pool START WITH 1 INCREMENT BY 1 NOCACHE;
CREATE SEQUENCE seq_answer_options START WITH 1 INCREMENT BY 1 NOCACHE;
CREATE SEQUENCE seq_tests START WITH 1 INCREMENT BY 1 NOCACHE;
CREATE SEQUENCE seq_questions START WITH 1 INCREMENT BY 1 NOCACHE;
CREATE SEQUENCE seq_answers START WITH 1 INCREMENT BY 1 NOCACHE;

-- ====================================================================
-- 2. ТАБЛИЦЫ
-- ====================================================================

-- Справочник тестов
CREATE TABLE quizzes (
    id_quiz NUMBER NOT NULL,
    title VARCHAR2(255) NOT NULL,
    description VARCHAR2(1000) NOT NULL,
    question_count NUMBER NOT NULL,      
    duration_minutes NUMBER NOT NULL,   
    passing_score NUMBER NOT NULL,
    CONSTRAINT pk_quizzes PRIMARY KEY (id_quiz)
);

-- Банк вопросов (Глобальный пул)
CREATE TABLE questions_pool (
    id_question NUMBER NOT NULL,
    id_quiz NUMBER NOT NULL,
    num_order NUMBER NOT NULL,
    text VARCHAR2(4000) NOT NULL,
    CONSTRAINT pk_questions_pool PRIMARY KEY (id_question),
    CONSTRAINT fk_q_pool_id_quiz FOREIGN KEY (id_quiz) REFERENCES quizzes(id_quiz) ON DELETE CASCADE
);

-- Варианты ответов (Глобальный пул)
CREATE TABLE answers_pool (
    id_answer NUMBER NOT NULL,
    id_question NUMBER NOT NULL,
    is_correct CHAR(1) DEFAULT 'N' NOT NULL, 
    text VARCHAR2(1000) NOT NULL,
    CONSTRAINT pk_answers_pool PRIMARY KEY (id_answer),
    CONSTRAINT fk_a_pool_id_question FOREIGN KEY (id_question) REFERENCES questions_pool(id_question) ON DELETE CASCADE,
    CONSTRAINT chk_a_pool_is_correct CHECK (is_correct IN ('Y', 'N'))
);

-- Активные сессии прохождения (Контекст теста)
CREATE TABLE tests (
    id_test NUMBER NOT NULL,
    id_quiz NUMBER NOT NULL,             
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,    
    current_question_idx NUMBER DEFAULT 0 NOT NULL, 
    is_finished CHAR(1) DEFAULT 'N' NOT NULL,
    final_score NUMBER DEFAULT 0,
    fio VARCHAR2(255) NOT NULL,
    dep_name VARCHAR2(255) NOT NULL,
    CONSTRAINT pk_tests PRIMARY KEY (id_test),
    CONSTRAINT fk_tests_quiz_id FOREIGN KEY (id_quiz) REFERENCES quizzes(id_quiz) ON DELETE CASCADE,
    CONSTRAINT chk_tests_is_finished CHECK (is_finished IN ('Y', 'N'))
);

-- Персональный рандомный пул вопросов для конкретной сессии
CREATE TABLE questions (
    id_question NUMBER NOT NULL,       -- Уникальный ID строки в текущей сессии (из seq_questions)
    id_test NUMBER NOT NULL,           -- Ссылка на сессию теста
    id_quest NUMBER NOT NULL,          -- Ссылка на реальный вопрос из questions_pool
    sort_order NUMBER NOT NULL,          
    selected_answer NUMBER,            -- ID выбранного ответа из answers_pool (или из таблицы answers)
    is_correct CHAR(1) DEFAULT 'N', 
    CONSTRAINT pk_questions PRIMARY KEY (id_question),
    CONSTRAINT fk_q_id_test FOREIGN KEY (id_test) REFERENCES tests(id_test) ON DELETE CASCADE,
    CONSTRAINT fk_q_id_quest_pool FOREIGN KEY (id_quest) REFERENCES questions_pool(id_question) ON DELETE CASCADE,
    CONSTRAINT chk_q_is_correct CHECK (is_correct IN ('Y', 'N'))
);

-- Персональный рандомный пул ответов для конкретной сессии (чтобы мешать их порядок)
CREATE TABLE answers (
    id_session_answer NUMBER NOT NULL, -- Суррогатный PK для этой таблицы (из seq_answers)
    id_test NUMBER NOT NULL,           -- Ссылка на тест
    id_question NUMBER NOT NULL,       -- Ссылка на персональный вопрос из таблицы questions
    id_answer NUMBER NOT NULL,         -- Ссылка на реальный ответ из answers_pool
    order_num NUMBER NOT NULL,          
    CONSTRAINT pk_answers PRIMARY KEY (id_session_answer),
    CONSTRAINT fk_a_id_test FOREIGN KEY (id_test) REFERENCES tests(id_test) ON DELETE CASCADE,
    CONSTRAINT fk_a_id_question FOREIGN KEY (id_question) REFERENCES questions(id_question) ON DELETE CASCADE,
    CONSTRAINT fk_a_id_answer_pool FOREIGN KEY (id_answer) REFERENCES answers_pool(id_answer) ON DELETE CASCADE
);
