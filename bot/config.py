"""Конфиг бота. Секреты только из окружения, в репозитории их нет."""
import os


def _req(name):
    v = os.environ.get(name, '').strip()
    if not v:
        raise SystemExit('нет переменной окружения ' + name)
    return v


TELEGRAM_TOKEN = _req('TELEGRAM_BOT_TOKEN')
FAMILYOS_API_URL = os.environ.get('FAMILYOS_API_URL', 'https://familyos.beliberde.net/api/v1').rstrip('/')
FAMILYOS_API_KEY = _req('FAMILYOS_API_KEY')
TELEGRAM_API = os.environ.get('TELEGRAM_API', 'https://api.telegram.org').rstrip('/')
TIMEZONE = os.environ.get('BOT_TIMEZONE', 'Europe/Moscow')
POLL_TIMEOUT = int(os.environ.get('BOT_POLL_TIMEOUT', '25'))
