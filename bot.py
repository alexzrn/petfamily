#!/usr/bin/env python3
"""Длинный опрос Telegram. Состояние диалогов в памяти процесса, данные в FamilyOS."""
import sys
import time

sys.path.insert(0, __file__.rsplit('/', 1)[0] + '/..' if '/' in __file__ else '.')

from bot import config
from bot.clients import ApiError, FamilyOS, Telegram
from bot.dialog import handle
from bot import store as store_mod


def main():
    tg = Telegram(config.TELEGRAM_API, config.TELEGRAM_TOKEN)
    fos = FamilyOS(config.FAMILYOS_API_URL, config.FAMILYOS_API_KEY)
    states = {}
    calls = {
        'ensure_user': lambda u: store_mod.ensure_user(fos, u),
        'list_pets': lambda u: store_mod.list_pets(fos, u),
        'add_pet': lambda u, n, s: store_mod.add_pet(fos, u, n, s),
        'find_pet': lambda u, p: store_mod.find_pet(fos, u, p),
        'list_reminders': lambda u: store_mod.list_reminders(fos, u),
        'add_reminder': lambda u, p, t, d: store_mod.add_reminder(fos, u, p, t, d),
    }
    offset = 0
    print('опрос запущен', flush=True)
    while True:
        try:
            batch = tg.get_updates(offset, config.POLL_TIMEOUT)
        except ApiError as e:
            print('poll error: ' + str(e), flush=True)
            time.sleep(5)
            continue
        for upd in (batch or {}).get('result', []):
            offset = max(offset, upd.get('update_id', 0) + 1)
            msg = upd.get('message') or {}
            chat = msg.get('chat') or {}
            text = msg.get('text', '')
            uid = chat.get('id')
            if uid is None:
                continue
            try:
                replies, states[uid] = handle(calls, uid, text, states.get(uid, {}), config.TIMEZONE)
            except ApiError as e:
                replies, states[uid] = ['FamilyOS недоступен: ' + str(e)[:120]], {}
            for r in replies:
                try:
                    tg.send(uid, r)
                except ApiError as e:
                    print('send error: ' + str(e), flush=True)


if __name__ == '__main__':
    main()
