"""Хранилище бота поверх FamilyOS API (заметки инвентаря). Чистые функции + тонкий слой."""
import json
from datetime import datetime, timezone


def user_title(tg_id):
    return 'myastaya:' + str(tg_id) + ':user'


def pet_title(tg_id, pet_id):
    return 'myastaya:' + str(tg_id) + ':pet:' + str(pet_id)


def rem_title(tg_id, rem_id):
    return 'myastaya:' + str(tg_id) + ':rem:' + str(rem_id)


def parse_doc(note):
    try:
        doc = json.loads(note.get('body') or '{}')
    except (ValueError, AttributeError):
        doc = {}
    doc['_note_id'] = note.get('id')
    return doc


def ensure_user(fos, tg_id):
    found = [n for n in fos.notes(user_title(tg_id)) if n.get('title') == user_title(tg_id)]
    if found:
        return found[0]
    body = json.dumps({'kind': 'user', 'tg_id': tg_id,
                       'registered_at': datetime.now(timezone.utc).isoformat()},
                      ensure_ascii=False)
    return fos.note_create(user_title(tg_id), body)


def _user_docs(fos, tg_id, kind):
    prefix = 'myastaya:' + str(tg_id) + ':'
    out = []
    for n in fos.notes(prefix):
        if not str(n.get('title', '')).startswith(prefix):
            continue
        doc = parse_doc(n)
        if doc.get('kind') == kind:
            out.append(doc)
    return out


def list_pets(fos, tg_id):
    return sorted(_user_docs(fos, tg_id, 'pet'), key=lambda d: d.get('pet_id', 0))


def add_pet(fos, tg_id, name, species):
    pets = list_pets(fos, tg_id)
    pid = max([p.get('pet_id', 0) for p in pets] + [0]) + 1
    body = json.dumps({'kind': 'pet', 'pet_id': pid, 'name': name, 'species': species},
                      ensure_ascii=False)
    fos.note_create(pet_title(tg_id, pid), body)
    return pid


def find_pet(fos, tg_id, pet_id):
    for p in list_pets(fos, tg_id):
        if p.get('pet_id') == pet_id:
            return p
    return None


def list_reminders(fos, tg_id, include_sent=False):
    docs = [d for d in _user_docs(fos, tg_id, 'rem')
            if include_sent or not d.get('sent')]
    return sorted(docs, key=lambda d: d.get('due_at', ''))


def add_reminder(fos, tg_id, pet_id, text, due_iso):
    rems = _user_docs(fos, tg_id, 'rem')
    rid = max([r.get('rem_id', 0) for r in rems] + [0]) + 1
    body = json.dumps({'kind': 'rem', 'rem_id': rid, 'pet_id': pet_id,
                       'text': text, 'due_at': due_iso, 'sent': False},
                      ensure_ascii=False)
    fos.note_create(rem_title(tg_id, rid), body)
    return rid


def all_pending(fos):
    """Все неотправленные напоминания всех пользователей (для крона)."""
    out = []
    for n in fos.notes('myastaya:'):
        doc = parse_doc(n)
        if doc.get('kind') == 'rem' and not doc.get('sent'):
            doc['_tg_id'] = str(n.get('title', '')).split(':')[1] if str(n.get('title', '')).count(':') >= 2 else ''
            out.append(doc)
    return out


def mark_sent(fos, doc):
    doc['sent'] = True
    clean = {k: v for k, v in doc.items() if not k.startswith('_')}
    fos.note_patch(doc['_note_id'], json.dumps(clean, ensure_ascii=False))
