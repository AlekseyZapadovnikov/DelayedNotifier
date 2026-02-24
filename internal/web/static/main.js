document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('notificationForm');
    const messageInput = document.getElementById('message');
    const sendAtInput = document.getElementById('send_at');
    const sendChanInput = document.getElementById('sendChan');
    const toInput = document.getElementById('to');
    const notificationsList = document.getElementById('notificationsList');
    const FINAL_STATUSES = new Set(['sended', 'redused']);

    if (!form || !notificationsList) {
        return;
    }

    form.addEventListener('submit', async (event) => {
        event.preventDefault();

        const message = messageInput?.value.trim() ?? '';
        const sendAt = sendAtInput?.value ?? '';
        const sendChan = sendChanInput?.value ?? '';
        const to = (toInput?.value ?? '')
            .split(',')
            .map((value) => value.trim())
            .filter(Boolean);

        if (!message || !sendAt || !sendChan || to.length === 0) {
            alert('Fill in all fields.');
            return;
        }

        const payload = {
            message,
            dateTime: new Date(sendAt).toISOString(),
            sendChan,
            to,
        };

        try {
            const response = await fetch('/notify', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                const errorText = await safeReadText(response);
                throw new Error(errorText || `Server error: ${response.status}`);
            }

            const created = await safeReadJSON(response);
            addNotificationToDOM({
                id: created?.id ?? `local-${Date.now()}`,
                message: created?.message ?? message,
                dateTime: created?.dateTime ?? payload.dateTime,
                status: created?.status ?? 'waiting',
                sendChan: created?.sendChan ?? sendChan,
                to: created?.to ?? to,
            });

            form.reset();
        } catch (error) {
            console.error('Failed to create notification:', error);
            alert(`Failed to create notification: ${error.message}`);
        }
    });

    notificationsList.addEventListener('click', async (event) => {
        const button = event.target.closest('.notification-delete');
        if (!button || button.disabled) {
            return;
        }

        const item = button.closest('.notification-item');
        const id = item?.dataset.notificationId;
        if (!item || !id || item.dataset.trackable !== 'true') {
            return;
        }

        button.disabled = true;
        button.textContent = 'Deleting...';

        try {
            const response = await fetch(`/notify/${id}/`, { method: 'DELETE' });
            if (!response.ok) {
                const errorText = await safeReadText(response);
                throw new Error(errorText || `Delete failed: ${response.status}`);
            }

            item.remove();
        } catch (error) {
            console.error(`Failed to delete notification ${id}:`, error);
            alert(`Failed to delete notification: ${error.message}`);
            button.disabled = false;
            button.textContent = 'Delete';
        }
    });

    function addNotificationToDOM(notification) {
        const item = document.createElement('div');
        item.className = 'notification-item';

        const id = String(notification.id ?? `local-${Date.now()}`);
        const status = String(notification.status ?? 'waiting');
        const sendTimeRaw = notification.dateTime ?? notification.send_at;
        const sendTime = sendTimeRaw ? new Date(sendTimeRaw).toLocaleString() : 'unknown';
        const isTrackable = /^\d+$/.test(id);

        item.dataset.notificationId = id;
        item.dataset.trackable = String(isTrackable);

        item.innerHTML = `
            <div class="notification-content">
                <p class="message">${escapeHtml(notification.message ?? '')}</p>
                <small>Send at: ${escapeHtml(sendTime)}</small>
                ${isTrackable ? '' : '<small> (status polling disabled: server did not return id)</small>'}
            </div>
            <div class="notification-meta">
                <span class="status status-${escapeHtml(status)}">${escapeHtml(status)}</span>
                <button type="button" class="notification-delete" ${isTrackable ? '' : 'disabled'}>Delete</button>
            </div>
        `;

        notificationsList.prepend(item);
    }

    async function updateStatuses() {
        const items = notificationsList.querySelectorAll('.notification-item');

        for (const item of items) {
            if (item.dataset.trackable !== 'true') {
                continue;
            }

            const id = item.dataset.notificationId;
            const statusElement = item.querySelector('.status');
            if (!id || !statusElement) {
                continue;
            }

            const currentStatus = statusElement.textContent.trim();
            if (FINAL_STATUSES.has(currentStatus)) {
                continue;
            }

            try {
                const response = await fetch(`/notify/${id}/`);
                if (!response.ok) {
                    continue;
                }

                const nextStatus = (await response.text()).trim();
                if (!nextStatus || currentStatus === nextStatus) {
                    continue;
                }

                statusElement.textContent = nextStatus;
                statusElement.className = `status status-${nextStatus}`;
            } catch (error) {
                console.error(`Failed to update status for ID ${id}:`, error);
            }
        }
    }

    setInterval(updateStatuses, 5000);
});

async function safeReadText(response) {
    try {
        return (await response.text()).trim();
    } catch {
        return '';
    }
}

async function safeReadJSON(response) {
    try {
        const text = await response.text();
        if (!text.trim()) {
            return null;
        }
        return JSON.parse(text);
    } catch {
        return null;
    }
}

function escapeHtml(value) {
    return String(value)
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#39;');
}
