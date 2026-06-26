// static/js/quiz_testing.js

// 1. Анимация и подсветка карточек ответов при клике
function initRadioAnimation() {
    const cards = document.querySelectorAll('.answer-label-card');
    cards.forEach(card => {
        const radio = card.querySelector('.custom-radio');
        if (!radio) return;

        card.addEventListener('click', (e) => {
            if (e.target !== radio) {
                radio.checked = true;
            }
            cards.forEach(c => c.classList.remove('selected-active'));
            card.classList.add('selected-active');
        });

        radio.addEventListener('change', () => {
            if (radio.checked) {
                cards.forEach(c => c.classList.remove('selected-active'));
                card.classList.add('selected-active');
            }
        });
    });
}

// 2. Логика досрочного завершения тестирования
function initFinishButton() {
    const finishBtn = document.getElementById('btnFinishQuiz');
    if (!finishBtn) return;
    
    finishBtn.addEventListener('click', () => {
        if (confirm("Вы действительно хотите досрочно завершить тест?\nНеотвеченные вопросы будут зачитаны как неверные.")) {
            window.location.href = "/quiz/result";
        }
    });
}

// 3. Перехват кликов по элементам навигации и отправка их через POST форму
function initNavigationInterceptor() {
    const quizForm = document.getElementById('quizForm');
    if (!quizForm) {
        console.error("❌ Форма quizForm не найдена на странице!");
        return;
    }

    // Находим ТОЛЬКО ссылки (плитка хедера и стрелочки футера)
    // Кнопку BUTTON из селектора убираем, у неё своя нативная валидация!
    const navLinks = document.querySelectorAll('.nav-wrapper a, .footer-nav-move a');

    console.log(`Проинициализировано ссылок навигации: ${navLinks.length}`);

    navLinks.forEach(element => {
        element.addEventListener('click', (e) => {
            e.preventDefault(); // Отменяем переход по ссылке

            console.log("Успешный клик по ссылке навигации:", element);

            let moveAction = 'save';

            const url = new URL(element.href, window.location.origin);
            const moveParam = url.searchParams.get('move');
            const numParam = url.searchParams.get('num');
            
            if (numParam) {
                moveAction = 'goto_' + numParam;
            } else if (moveParam) {
                moveAction = moveParam;
            }

            console.log("Вычисленная команда движения (move):", moveAction);

            // Обновляем скрытое поле move
            const oldInput = quizForm.querySelector('input[name="move"]');
            if (oldInput) oldInput.remove();

            const moveInput = document.createElement('input');
            moveInput.type = 'hidden';
            moveInput.name = 'move';
            moveInput.value = moveAction;
            quizForm.appendChild(moveInput);
            
            // Если переходим по ссылкам, отключаем обязательность выбора радиокнопки
            const radioInputs = quizForm.querySelectorAll('.custom-radio');
            radioInputs.forEach(radio => radio.removeAttribute('required'));

            console.log("🚀 Отправка формы через навигацию POST /quiz/process...");
            quizForm.submit(); 
        });
    });

    // Отдельно обрабатываем нативную отправку формы по кнопке "Сохранить"
    quizForm.addEventListener('submit', (e) => {
        // Проверяем, что отправка вызвана именно кнопкой "Сохранить"
        // Нам нужно просто добавить hidden-поле move=save перед уходом данных
        const oldInput = quizForm.querySelector('input[name="move"]');
        if (oldInput) oldInput.remove();

        const moveInput = document.createElement('input');
        moveInput.type = 'hidden';
        moveInput.name = 'move';
        moveInput.value = 'save';
        quizForm.appendChild(moveInput);

        console.log("🚀 Нативная отправка формы (Сохранить) POST /quiz/process...");
        // e.preventDefault() НЕ вызываем, чтобы сработал required браузера
    });
}

// 4. Таймер обратного отсчета с автоматическим завершением теста
function initQuizTimer() {
    const timerElement = document.getElementById('timer');
    if (!timerElement) return;

    // Читаем Unix Timestamp (в секундах), который прилетел из Oracle через Go
    const endTimestamp = parseInt(timerElement.getAttribute('data-end'), 10);
    
    // Если база данных ничего не вернула, останавливаем скрипт
    if (!endTimestamp || isNaN(endTimestamp)) return;

    function updateTimer() {
        // Текущее время компьютера в секундах
        const nowSeconds = Math.floor(Date.now() / 1000);
        // Вычисляем разницу
        const timeLeft = endTimestamp - nowSeconds;

        // Если время вышло
        if (timeLeft <= 0) {
            timerElement.textContent = "00:00";
            timerElement.style.color = "#e11d48"; // Красный цвет
            clearInterval(timerInterval);
            
            alert("Время, отведенное на тестирование, истекло!");
            window.location.href = "/quiz/result"; // Автоматический редирект
            return;
        }

        // Переводим секунды в минуты и остаток секунд
        const minutes = Math.floor(timeLeft / 60);
        const seconds = timeLeft % 60;

        // Форматируем в красивый вид 05:09 вместо 5:9
        const displayMinutes = String(minutes).padStart(2, '0');
        const displaySeconds = String(seconds).padStart(2, '0');

        // Обновляем текст на странице
        timerElement.textContent = `${displayMinutes}:${displaySeconds}`;

        // Если осталась последняя минута — подсвечиваем красным и включаем пульсацию
        if (timeLeft < 60) {
            timerElement.style.color = "#e11d48";
            timerElement.classList.add('pulse-animation');
        }
    }

    // Запускаем первый раз сразу при загрузке
    updateTimer();
    
    // Включаем счетчик каждую секунду (1000 миллисекунд)
    const timerInterval = setInterval(updateTimer, 1000);
}

// Запуск всех модулей после полной загрузки страницы
document.addEventListener('DOMContentLoaded', () => {
    initRadioAnimation();
    initFinishButton();
    initNavigationInterceptor();
    initQuizTimer(); // <-- ВОТ ТЕПЕРЬ ОНО БУДЕТ ТИКАТЬ!
});
