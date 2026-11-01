"""HTTP-клиент Telegram Bot API и FamilyOS API. Только стандартная библиотека."""
import json
import urllib.error
import urllib.parse
import urllib.request


class ApiError(Exception):
    pass


def _call(base, headers, method, path, body=None, timeout=40):
    data = None
    req_headers = dict(headers)
    if body is not None:
        data = json.dumps(body, ensure_ascii=False).encode('utf-8')
        req_headers['Content-Type'] = 'application/json'
    req = urllib.request.Request(base + path, data=data, headers=req_headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as r:
            raw = r.read().decode('utf-8')
            return json.loads(raw) if raw else None
    except urllib.error.HTTPError as e:
        raise ApiError('http ' + str(e.code) + ': ' + e.read().decode('utf-8', 'replace')[:300])


class Telegram:
    def __init__(self, api_base, token):
        self.base = api_base + '/bot' + token

    def get_updates(self, offset, timeout):
        return _call(self.base, {}, 'POST', '/getUpdates',
                      {'offset': offset, 'timeout': timeout, 'allowed_updates': ['message']},
                      timeout=timeout + 10)

    def send(self, chat_id, text):
        return _call(self.base, {}, 'POST', '/sendMessage',
                      {'chat_id': chat_id, 'text': text})


class FamilyOS:
    """Питомцы и напоминания лежат в заметках инвентаря с префиксом myastaya:.
    Выделенных endpoints питомцев и напоминаний в API нет (факт, см. swagger.json),
    заметки используются как key-value хранилище. Это решение проекта для MVP."""

    PREFIX = 'myastaya:'

    def __init__(self, api_url, api_key):
        self.api_url = api_url
        self.headers = {'X-API-Key': api_key}

    def _get(self, path, params=None):
        if params:
            path += '?' + urllib.parse.urlencode(params)
        return _call(self.api_url, self.headers, 'GET', path)

    def notes(self, query):
        res = self._get('/inventory/notes', {'q': query})
        return res if isinstance(res, list) else (res or {}).get('items', res or [])

    def note_create(self, title, body):
        return _call(self.api_url, self.headers, 'POST', '/inventory/notes',
                      {'title': title, 'body': body})

    def note_patch(self, note_id, body):
        return _call(self.api_url, self.headers, 'PATCH',
                      '/inventory/notes/' + str(note_id), {'body': body})

    def note_delete(self, note_id):
        return _call(self.api_url, self.headers, 'DELETE',
                      '/inventory/notes/' + str(note_id))
