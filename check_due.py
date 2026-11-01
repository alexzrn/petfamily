#!/usr/bin/env python3
"""Крон каждые 5 минут: отправляет наступившие напоминания, помечает sent.
Строка крона: */5 * * * * /usr/bin/python3 /opt/staya-bot/check_due.py
Секреты из окружения процесса, не из репозитория."""
import sys
from datetime import datetime, timezone

sys.path.insert(0, __file__.rsplit('/', 1)[0] + '/..' if '/' in __file__ else '.')

from bot import config
from bot.clients import ApiError, FamilyOS, Telegram
from bot import store as store_mod


def main():
    tg = Telegram(config.TELEGRAM_API, config.TELEGRAM_TOKEN)
    fos = FamilyOS(config.FAMILYOS_API_URL, config.FAMILYOS_API_KEY)
    now = datetime.now(timezone.utc)
    sent = 0
    try:
        pending = store_mod.all_pending(fos)
    except ApiError as e:
        print('familyos error: ' + str(e), flush=True)
        raise SystemExit(1)
    for doc in pending:
        try:
            due = datetime.fromisoformat(doc.get('due_at', ''))
        except ValueError:
            continue
        if due.tzinfo is None:
            due = due.replace(tzinfo=timezone.utc)
        if due > now:
            continue
        chat_id = doc.get('_tg_id')
        if not chat_id:
            continue
        try:
            tg.send(int(chat_id), 'Напоминание: ' + doc.get('text', ''))
            store_mod.mark_sent(fos, doc)
            sent += 1
        except (ApiError, ValueError) as e:
            print('send error: ' + str(e), flush=True)
    print('отправлено: ' + str(sent), flush=True)


if __name__ == '__main__':
    main()
