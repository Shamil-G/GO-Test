document.getElementById('startQuizForm').addEventListener('submit', async function (e) {
    e.preventDefault();

    const submitBtn = e.target.querySelector('button[type="submit"]');
    const errorContainer = document.getElementById('errorContainer');
    
    // Сбрасываем старую ошибку при повторном клике
    errorContainer.style.display = 'none';
    errorContainer.innerText = '';

    submitBtn.disabled = true;
    submitBtn.innerText = 'Запуск...';

    const requestData = {
        id_quiz: parseInt(document.getElementById('idQuiz').value, 10),
        fio: document.getElementById('fioInput').value.trim(),
        dep_name: document.getElementById('depInput').value.trim()
    };

    try {
        const response = await fetch('/api/start', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(requestData)
        });

        // Если бэкенд упал с 500 ошибкой, пробуем прочитать JSON с текстом ошибки
        if (!response.ok) {
            const errData = await response.json().catch(() => ({}));
            throw new Error(errData.error || 'Ошибка сервера при инициализации сессии');
        }

        const data = await response.json();

        if (data.error) {
            throw new Error(data.error);
        }

        // Переход на страницу процесса теста
        window.location.href = `/quiz/process`;

    } catch (err) {
        // ИСПРАВЛЕНИЕ: Выводим ошибку красиво в элемент интерфейса, а не в alert
        errorContainer.innerText = '❌ Не удалось начать тест: ' + err.message;
        errorContainer.style.display = 'block';
        
        submitBtn.disabled = false;
        submitBtn.innerText = 'Начать тестирование';
    }
});
