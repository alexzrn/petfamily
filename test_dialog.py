import unittest
from bot import dialog


class Calls:
    def __init__(self):
        self.pets = {}
        self.rems = []

    def ensure_user(self, u):
        return True

    def list_pets(self, u):
        return self.pets.get(u, [])

    def add_pet(self, u, n, s):
        pid = len(self.pets.get(u, [])) + 1
        self.pets.setdefault(u, []).append({'pet_id': pid, 'name': n, 'species': s})
        return pid

    def find_pet(self, u, pid):
        return next((p for p in self.pets.get(u, []) if p['pet_id'] == pid), None)

    def list_reminders(self, u):
        return self.rems

    def add_reminder(self, u, p, t, d):
        self.rems.append({'rem_id': len(self.rems) + 1, 'pet_id': p, 'text': t, 'due_at': d})
        return len(self.rems)


def run(calls, uid, texts):
    state = {}
    out = []
    for t in texts:
        replies, state = dialog.handle({k: getattr(calls, k) for k in
            ('ensure_user', 'list_pets', 'add_pet', 'find_pet', 'list_reminders', 'add_reminder')}, uid, t, state, 'Europe/Moscow')
        out.extend(replies)
    return out


class TestDialog(unittest.TestCase):
    def test_full_flow(self):
        c = Calls()
        out = run(c, 1, ['/start', '/addpet', 'Бублик', 'собака', '/remind', '1', 'Прививка', '2026-12-01 10:00', '/list'])
        text = '\n'.join(out)
        self.assertIn('Бублик', text)
        self.assertIn('2026-12-01', text)
        self.assertIn('#1', text)

    def test_bad_date(self):
        c = Calls()
        out = run(c, 1, ['/addpet', 'Кот', 'кошка', '/remind', '1', 'Таблетка', 'завтра'])
        self.assertTrue(any('ГГГГ-ММ-ДД' in r for r in out))

    def test_unknown_pet(self):
        c = Calls()
        out = run(c, 1, ['/addpet', 'Кот', 'кошка', '/remind', '9'])
        self.assertTrue(any('Такого номера нет' in r for r in out))

    def test_cancel(self):
        c = Calls()
        out = run(c, 1, ['/addpet', '/cancel'])
        self.assertTrue(any('Отменено' in r for r in out))


if __name__ == '__main__':
    unittest.main()
