// Доска потеряшек: фильтры, карточки, форма с клиентской проверкой и черновиками в localStorage.
(function () {
  'use strict';

  var items = window.__LOST__ || [];
  var state = { kind: '', animal: '', district: '' };
  var ANIMALS = { dog: 'Собака', cat: 'Кошка', other: 'Другое' };
  var DRAFT_KEY = 'staya-lost-drafts-v1';

  function esc(value) {
    return String(value == null ? '' : value).replace(/[&<>"']/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
    });
  }

  var list = document.getElementById('lost-list');
  var count = document.getElementById('result-count');
  var empty = document.getElementById('lost-empty');

  function card(item) {
    var title = item.name || ANIMALS[item.animal] + ', кличка неизвестна';
    return (
      '<li class="chat-item' +
      (item.kind === 'found' ? ' chat-item--found' : '') +
      '">' +
      '<span class="chat-item__kind">' +
      (item.kind === 'lost' ? 'Поиск' : 'Находка') +
      '</span>' +
      (item.test ? '<span class="chat-item__demo">пример</span>' : '') +
      '<h3>' +
      esc(title) +
      '</h3>' +
      '<p class="chat-item__meta">' +
      esc(ANIMALS[item.animal] || item.animal) +
      ' · ' +
      esc(item.district || 'округ не указан') +
      ' · ' +
      esc(item.date) +
      '</p>' +
      (item.place ? '<p>' + esc(item.place) + '</p>' : '') +
      '<p>' +
      esc(item.description) +
      '</p>' +
      '<p class="chat-item__contact">Контакт: ' +
      esc(item.contact) +
      '</p></li>'
    );
  }

  function render() {
    var shown = items.filter(function (item) {
      return (
        (!state.kind || item.kind === state.kind) &&
        (!state.animal || item.animal === state.animal) &&
        (!state.district || item.district === state.district)
      );
    });
    if (list) list.innerHTML = shown.map(card).join('');
    if (count) count.textContent = 'Объявлений: ' + shown.length + ' из ' + items.length;
    if (empty) empty.hidden = shown.length !== 0;
  }

  function chipRow(id, key) {
    var row = document.getElementById(id);
    if (!row) return;
    row.addEventListener('click', function (event) {
      var btn = event.target.closest('button[data-' + key + ']');
      if (!btn) return;
      state[key] = btn.dataset[key];
      var chips = row.querySelectorAll('.chip');
      for (var i = 0; i < chips.length; i++) chips[i].setAttribute('aria-pressed', String(chips[i] === btn));
      render();
    });
  }

  chipRow('kind-row', 'kind');
  chipRow('animal-row', 'animal');
  var districtSel = document.getElementById('f-district');
  if (districtSel) {
    districtSel.addEventListener('change', function () {
      state.district = districtSel.value;
      render();
    });
  }
  render();

  // --- форма ---
  var form = document.getElementById('lost-form');
  if (!form) return;
  var success = document.getElementById('lost-success');
  var draftsBox = document.getElementById('drafts');
  var draftsList = document.getElementById('drafts-list');
  var fields = {
    kind: document.getElementById('lf-kind'),
    animal: document.getElementById('lf-animal'),
    district: document.getElementById('lf-district'),
    place: document.getElementById('lf-place'),
    contact: document.getElementById('lf-contact'),
    desc: document.getElementById('lf-desc'),
    consent: document.getElementById('lf-consent'),
  };

  function setError(name, message) {
    var node = form.querySelector('[data-error="' + name + '"]');
    if (node) {
      node.textContent = message || '';
      node.dataset.show = message ? 'true' : 'false';
    }
    var field = fields[name];
    if (field && 'setAttribute' in field) {
      if (message) field.setAttribute('aria-invalid', 'true');
      else field.removeAttribute('aria-invalid');
    }
  }

  function readDrafts() {
    try {
      var raw = localStorage.getItem(DRAFT_KEY);
      var parsed = raw ? JSON.parse(raw) : [];
      return Array.isArray(parsed) ? parsed : [];
    } catch (error) {
      return [];
    }
  }

  function saveDrafts(list_) {
    try {
      localStorage.setItem(DRAFT_KEY, JSON.stringify(list_.slice(-20)));
      return true;
    } catch (error) {
      return false;
    }
  }

  function renderDrafts() {
    if (!draftsBox || !draftsList) return;
    var drafts = readDrafts();
    draftsBox.hidden = drafts.length === 0;
    draftsList.innerHTML = drafts
      .slice()
      .reverse()
      .map(function (item) {
        return card(item);
      })
      .join('');
  }

  form.addEventListener('submit', function (event) {
    event.preventDefault();
    var contact = fields.contact.value.trim();
    var desc = fields.desc.value.trim();
    var contactOk = /^[+0-9][0-9 ()-]{5,}$/.test(contact) || /^[^@\s]+@[^@\s]+\.[^@\s]{2,}$/.test(contact);

    setError('kind', fields.kind.value ? '' : 'Выберите, что случилось');
    setError('animal', fields.animal.value ? '' : 'Выберите вид животного');
    setError('district', fields.district.value ? '' : 'Выберите округ');
    setError('place', '');
    setError('contact', contactOk ? '' : 'Нужен телефон или почта');
    setError('desc', desc.length >= 20 ? '' : 'Опишите подробнее: минимум 20 символов');
    setError('consent', fields.consent.checked ? '' : 'Без согласия публиковать нельзя');

    var invalid = form.querySelector('[aria-invalid="true"]');
    if (invalid) {
      invalid.focus();
      return;
    }

    var draft = {
      id: 'draft-' + Date.now(),
      kind: fields.kind.value,
      animal: fields.animal.value,
      district: fields.district.value,
      place: fields.place.value.trim(),
      contact: contact,
      description: desc,
      date: new Date().toISOString().slice(0, 10),
      test: false,
    };

    var saved = saveDrafts(readDrafts().concat([draft]));
    if (success) {
      success.hidden = false;
      success.querySelector('h3').textContent = saved
        ? 'Черновик сохранён на этом устройстве'
        : 'Браузер не дал сохранить черновик';
      success.focus();
    }
    form.reset();
    renderDrafts();
  });

  Object.keys(fields).forEach(function (name) {
    var field = fields[name];
    if (!field || !field.addEventListener) return;
    ['input', 'change'].forEach(function (type) {
      field.addEventListener(type, function () {
        if (field.getAttribute('aria-invalid') === 'true') setError(name, '');
      });
    });
  });

  renderDrafts();
})();
