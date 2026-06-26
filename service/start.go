package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	mdw "gusseynov/GO-Quiz/middleware"
	"gusseynov/GO-Quiz/storage"
)

func GetQuiz(idQuiz string, lang string, fio string, depName string) (int, error) {
	// Создаем контекст таймаута прямо здесь внутри, автономно!
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// query := `BEGIN testing.get_quiz(:1, :2, :3, :4); END;`

	err := storage.DBExec(ctx, "test.registration", idQuiz, fio, depName)
	if err != nil {
		return 0, err
	}

	slog.Debug("ADD USE_FILE_STATISTIC", "employee", fio, "dep_name", depName)
	return 0, nil
}

func StartQuiz(ctx_service context.Context, idQuiz int) error {
	ctx := mdw.GetOrCreatePageCtx(ctx_service)

	outErr := strings.Repeat(" ", 2000)

	slog.Debug("StartQuiz", "idQuiz", idQuiz, "IP", ctx.IP, "Lang", ctx.Lang, "fio", ctx.FIO, "dep", ctx.DepName)

	// Используем именованные параметры вместо :1, :2...
	// 1. Формируем чистый PL/SQL блок с понятными именами для go-ora
	query := `BEGIN test.registration(:id, :ip, :lang, :fio, :dep, :out_err); END;`

	// 2. Вызываем DBExecNamed вместо DBExec
	err := storage.DBExecNamed(ctx_service,
		query,               // Сама строка запроса
		"test.registration", // Имя для метрик Прометея (procName)
		sql.Named("id", idQuiz),
		sql.Named("ip", ctx.IP),
		sql.Named("lang", ctx.Lang),
		sql.Named("fio", ctx.FIO),
		sql.Named("dep", ctx.DepName),
		sql.Named("out_err", sql.Out{Dest: &outErr}),
	)
	if err != nil {
		return fmt.Errorf("oracle error: %w", err)
	}

	// ИСПРАВЛЕНИЕ: Сначала переводим байты в строку, отсекая пустые символы концевого буфера
	outErr = strings.TrimSpace(outErr)

	if outErr != "" {
		slog.Warn("Процедура test.registration вернула ошибку", "outErr", outErr)
		return fmt.Errorf("%s", outErr)
	}

	return nil
}
