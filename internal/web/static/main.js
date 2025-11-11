// Ждем, пока весь HTML-документ будет загружен и готов к взаимодействию
document.addEventListener('DOMContentLoaded', () => {

    // Находим ключевые элементы на странице
    const form = document.getElementById('notificationForm');
    const messageInput = document.getElementById('message');
    const sendAtInput = document.getElementById('send_at');
    const notificationsList = document.getElementById('notificationsList');

    // --- 1. Обработка отправки формы для создания уведомления ---
    form.addEventListener('submit', async (event) => {
        // Предотвращаем стандартное поведение формы (перезагрузку страницы)
        event.preventDefault();

        // Собираем данные из полей ввода
        const message = messageInput.value;
        const sendAt = sendAtInput.value;

        // Простая валидация
        if (!message || !sendAt) {
            alert('Пожалуйста, заполните все поля.');
            return;
        }

        try {
            // Отправляем POST-запрос на сервер с помощью fetch API
            const response = await fetch('/notify', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                // Преобразуем данные в JSON-строку
                body: JSON.stringify({
                    message: message,
                    // Приводим дату к стандартному формату ISO, который понимают большинство серверов
                    send_at: new Date(sendAt).toISOString(),
                }),
            });

            if (!response.ok) {
                // Если сервер ответил ошибкой, выбрасываем исключение
                throw new Error(`Ошибка сервера: ${response.status}`);
            }

            // Получаем созданное уведомление из ответа сервера
            const newNotification = await response.json();
            
            // Добавляем новое уведомление в список на странице
            addNotificationToDOM(newNotification);

            // Очищаем форму
            form.reset();

        } catch (error) {
            console.error('Ошибка при создании уведомления:', error);
            alert('Не удалось создать уведомление. Проверьте консоль для деталей.');
        }
    });


    // --- 2. Функция для добавления элемента уведомления на страницу ---
    function addNotificationToDOM(notification) {
        // Создаем новый div для уведомления
        const item = document.createElement('div');
        item.className = 'notification-item';
        // Устанавливаем ID, чтобы легко находить и обновлять этот элемент
        item.id = `notification-${notification.id}`;

        // Форматируем дату для более читаемого вида
        const sendTime = new Date(notification.send_at).toLocaleString();

        // Заполняем HTML-содержимое элемента
        item.innerHTML = `
            <div>
                <p class="message">${notification.message}</p>
                <small>Отправить: ${sendTime}</small>
            </div>
            <span class="status status-${notification.status}">${notification.status}</span>
        `;

        // Добавляем новый элемент в начало списка
        notificationsList.prepend(item);
    }


    // --- 3. Периодическое обновление статусов существующих уведомлений ---
    async function updateStatuses() {
        // Находим все элементы уведомлений на странице
        const items = document.querySelectorAll('.notification-item');

        for (const item of items) {
            const id = item.id.split('-')[1];
            const statusElement = item.querySelector('.status');

            // Пропускаем обновление финальных статусов, чтобы не делать лишних запросов
            if (statusElement.textContent === 'sent' || statusElement.textContent === 'cancelled') {
                continue;
            }

            try {
                // Отправляем GET-запрос для получения актуального статуса
                const response = await fetch(`/notify/${id}`);
                if (response.ok) {
                    const data = await response.json();
                    
                    // Обновляем текст и класс статуса, если он изменился
                    if (statusElement.textContent !== data.status) {
                        statusElement.textContent = data.status;
                        statusElement.className = `status status-${data.status}`;
                    }
                }
            } catch (error) {
                // Ошибки обновления одного элемента не должны ломать весь цикл
                console.error(`Ошибка при обновлении статуса для ID ${id}:`, error);
            }
        }
    }

    // Запускаем функцию обновления статусов каждые 5 секунд (5000 миллисекунд)
    setInterval(updateStatuses, 5000);
});