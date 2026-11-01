"""Диалог бота. Чистая логика без сети: handle() по сообщению возвращает ответы.
Состояние разговора хранится у вызывающей стороны (dict user_id -> state)."""
from datetime import datetime
from zoneinfo import ZoneInfo

HELP = ('Команды: /start, /addpet (добавить питомца), /pets (мои питомцы), '
        '/remind (напомнить), /list (ближайшие напоминания), /cancel (отмена).')


def parse_when(text, tzname):
    """ГГГГ-ММ-ДД ЧЧ:ММ в часовом поясе бота. Пусто при ошибке."""
    try:
        naive = datetime.strptime(text.strip(), '%Y-%m-%d %H:%M')
    except ValueError:
        return ''
    tz = ZoneInfo(tzname)
    return naive.replace(tzinfo=tz).isoformat()


def fmt_rem(r, pets):
    name = next((p.get('name', '?') for p in pets if p.get('pet_id') == r.get('pet_id')), '?')
    return '#' + str(r.get('rem_id')) + ' ' + name + ': ' + r.get('text', '') + ' (' + r.get('due_at', '') + ')'


def handle(store_calls, user_id, text, state, tzname):
    """store_calls: dict колбэков хранилища. Возвращает (replies[list], new_state)."""
    text = (text or '').strip()
    cmd, _, arg = text.partition(' ')
    if cmd.startswith('/'):
        state = {}
        if cmd == '/start':
            store_calls['ensure_user'](user_id)
            return [(
                'Моя стая: напоминания о прививках, таблетках и визитах. '
                'Сначала добавьте питомца: /addpet. ' + HELP
            )], state
        if cmd == '/cancel':
            return ['Отменено. ' + HELP], state
        if cmd == '/addpet':
            return ['Как зовут питомца?'], {'step': 'pet_name'}
        if cmd == '/pets':
            pets = store_calls['list_pets'](user_id)
            if not pets:
                return ['Питомцев пока нет. Добавьте: /addpet'], state
            lines = [str(p.get('pet_id')) + '. ' + p.get('name') + ' (' + p.get('species') + ')' for p in pets]
            return ['Мои питомцы:\n' + '\n'.join(lines)], state
        if cmd == '/remind':
            pets = store_calls['list_pets'](user_id)
            if not pets:
                return ['Сначала добавьте питомца: /addpet'], state
            lines = [str(p.get('pet_id')) + '. ' + p.get('name') for p in pets]
            return ['Номер питомца?\n' + '\n'.join(lines)], {'step': 'rem_pet'}
        if cmd == '/list':
            pets = store_calls['list_pets'](user_id)
            rems = store_calls['list_reminders'](user_id)
            if not rems:
                return ['Напоминаний нет. Создать: /remind'], state
            return ['Ближайшие:\n' + '\n'.join(fmt_rem(r, pets) for r in rems)], state
        return ['Не знаю такую команду. ' + HELP], state

    step = state.get('step')
    if step == 'pet_name':
        if not text:
            return ['Имя не должно быть пустым. Как зовут питомца?'], state
        return ['Вид: собака, кошка или другое?'], {'step': 'pet_species', 'name': text[:60]}
    if step == 'pet_species':
        low = text.lower()
        species = 'собака' if low.startswith('собак') else 'кошка' if low.startswith('кош') else 'другое'
        pid = store_calls['add_pet'](user_id, state.get('name', '?'), species)
        return ['Записано: ' + state.get('name', '?') + ' (' + species + '), номер ' + str(pid) + '. Напомнить: /remind'], {}
    if step == 'rem_pet':
        try:
            pid = int(text)
        except ValueError:
            return ['Нужен номер питомца из списка.'], state
        if not store_calls['find_pet'](user_id, pid):
            return ['Такого номера нет. Проверьте /pets.'], state
        return ['Текст напоминания?'], {'step': 'rem_text', 'pet_id': pid}
    if step == 'rem_text':
        if not text:
            return ['Текст не должен быть пустым.'], state
        return ['Дата и время? Формат: ГГГГ-ММ-ДД ЧЧ:ММ'], {'step': 'rem_when', 'pet_id': state.get('pet_id'), 'text': text[:200]}
    if step == 'rem_when':
        due = parse_when(text, tzname)
        if not due:
            return ['Не разобрал дату. Формат: ГГГГ-ММ-ДД ЧЧ:ММ'], state
        rid = store_calls['add_reminder'](user_id, state.get('pet_id'), state.get('text'), due)
        return ['Готово, напоминание #' + str(rid) + ' на ' + due + '. Список: /list'], {}
    return ['Напишите команду. ' + HELP], {}
