// Портал «Моя стая»: карта, поиск и фильтры. Без внешних зависимостей, кроме локального Leaflet.
// Данные приходят из window.__SPOTS__ (инлайн JSON, собранный на этапе build).
(function () {
  'use strict';

  var spots = window.__SPOTS__ || [];
  var mapEl = document.getElementById('map');
  if (!mapEl || !window.L) return;

  var LABELS = {
    vet: 'Клиники',
    park: 'Площадки',
    shelter: 'Приюты',
    boarding: 'Передержка',
    grooming: 'Груминг',
    shop: 'Магазины',
  };

  var mode = mapEl.dataset.mode || 'full';
  var state = { cat: '', q: '', open_24h: false, home_visit: false, inpatient: false };

  var map = L.map(mapEl, { scrollWheelZoom: false, attributionControl: true }).setView([55.7558, 37.6173], 10);
  L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
  }).addTo(map);

  var layer = L.layerGroup().addTo(map);
  var items = [];
  var byId = {};

  function esc(value) {
    return String(value == null ? '' : value).replace(/[&<>"']/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
    });
  }

  function norm(value) {
    return String(value == null ? '' : value)
      .toLowerCase()
      .replace(/ё/g, 'е')
      .replace(/\s+/g, ' ')
      .trim();
  }

  function telHref(phone) {
    var digits = String(phone).replace(/[^\d+]/g, '');
    if (digits.charAt(0) === '8') return 'tel:+7' + digits.slice(1);
    if (digits.charAt(0) === '+') return 'tel:' + digits;
    return 'tel:+7' + digits;
  }

  function matches(spot) {
    if (state.cat && spot.category !== state.cat) return false;
    if (state.open_24h && spot.open_24h !== true) return false;
    if (state.home_visit && spot.home_visit !== true) return false;
    if (state.inpatient && spot.inpatient !== true) return false;
    if (state.q) {
      var hay = norm(spot.name + ' ' + spot.address + ' ' + (spot.district || ''));
      var words = state.q.split(' ').filter(Boolean);
      for (var i = 0; i < words.length; i++) {
        if (hay.indexOf(words[i]) === -1) return false;
      }
    }
    return true;
  }

  function flags(spot) {
    var out = [];
    if (spot.open_24h) out.push('круглосуточно');
    if (spot.home_visit) out.push('выезд');
    if (spot.inpatient) out.push('стационар');
    return out;
  }

  function popupHtml(spot) {
    return (
      '<strong>' +
      esc(spot.name) +
      '</strong><br>' +
      '<span class="popup__cat">' +
      esc(LABELS[spot.category] || spot.category) +
      '</span><br>' +
      esc(spot.address) +
      (spot.phone ? '<br><a href="' + esc(telHref(spot.phone)) + '">' + esc(spot.phone) + '</a>' : '') +
      '<br><a href="/spot/' + esc(spot.id) + '">Карточка точки →</a>'
    );
  }

  var all = spots.map(function (spot) {
    var marker = L.marker([spot.lat, spot.lon], {
      title: spot.name,
      icon: L.divIcon({
        className: 'pin-wrap',
        html: '<span class="pin pin--' + esc(spot.category) + '"></span>',
        iconSize: [18, 18],
        iconAnchor: [9, 9],
        popupAnchor: [0, -8],
      }),
    });
    marker.bindPopup(popupHtml(spot), { closeButton: true, maxWidth: 260 });
    byId[spot.id] = marker;
    return { spot: spot, marker: marker };
  });

  var listEl = document.getElementById('spot-list');
  var countEl = document.getElementById('result-count');
  var emptyEl = document.getElementById('empty');
  var openMapEl = document.getElementById('open-map');

  function cardHtml(spot) {
    var tagList = flags(spot);
    return (
      '<li class="spot-card" data-id="' +
      esc(spot.id) +
      '">' +
      '<a class="spot-card__link" href="/spot/' + esc(spot.id) + '">' +
      '<span class="spot-card__cat spot-card__cat--' +
      esc(spot.category) +
      '">' +
      esc(LABELS[spot.category] || spot.category) +
      '</span>' +
      '<span class="spot-card__name">' +
      esc(spot.name) +
      '</span>' +
      '<span class="spot-card__addr">' +
      esc(spot.address) +
      '</span>' +
      (tagList.length ? '<span class="spot-card__tags">' + esc(tagList.join(' · ')) + '</span>' : '') +
      (spot.hours && !(spot.open_24h && norm(spot.hours).indexOf('круглосуточно') === 0)
        ? '<span class="spot-card__tags">' + esc(spot.hours) + '</span>'
        : '') +
      '</a></li>'
    );
  }

  var lastFitKey = '';

  function render() {
    var shown = [];
    layer.clearLayers();
    for (var i = 0; i < all.length; i++) {
      if (!matches(all[i].spot)) continue;
      shown.push(all[i].spot);
      layer.addLayer(all[i].marker);
    }

    if (countEl) {
      countEl.textContent = 'Показано точек: ' + shown.length + ' из ' + spots.length;
    }
    if (listEl) {
      listEl.innerHTML = shown.map(cardHtml).join('');
    }
    if (emptyEl) emptyEl.hidden = shown.length !== 0;
    if (openMapEl) openMapEl.href = '/map' + (state.cat ? '?cat=' + encodeURIComponent(state.cat) : '');

    var fitKey = shown.length + ':' + state.cat + ':' + state.q + ':' + state.open_24h + state.home_visit + state.inpatient;
    if (shown.length && fitKey !== lastFitKey) {
      lastFitKey = fitKey;
      var bounds = shown.map(function (s) {
        return [s.lat, s.lon];
      });
      if (shown.length === 1) {
        map.setView(bounds[0], 15);
      } else {
        map.fitBounds(bounds, { padding: [24, 24], maxZoom: mode === 'home' ? 11 : 13 });
      }
    }
  }

  function setChip(row, active) {
    if (!row) return;
    var chips = row.querySelectorAll('.chip');
    for (var i = 0; i < chips.length; i++) {
      chips[i].setAttribute('aria-pressed', String(chips[i] === active));
    }
  }

  var catRow = document.getElementById('cat-row');
  if (catRow) {
    catRow.addEventListener('click', function (event) {
      var btn = event.target.closest('button[data-cat]');
      if (!btn) return;
      state.cat = btn.dataset.cat;
      setChip(catRow, btn);
      render();
    });
  }

  var flagRow = document.getElementById('flag-row');
  if (flagRow) {
    flagRow.addEventListener('click', function (event) {
      var btn = event.target.closest('button[data-flag]');
      if (!btn) return;
      var key = btn.dataset.flag;
      state[key] = !state[key];
      btn.setAttribute('aria-pressed', String(state[key]));
      render();
    });
  }

  var searchEl = document.getElementById('q');
  var timer = null;
  if (searchEl) {
    searchEl.addEventListener('input', function () {
      clearTimeout(timer);
      timer = setTimeout(function () {
        state.q = norm(searchEl.value);
        render();
      }, 140);
    });
    var form = document.getElementById('search-form');
    if (form) {
      form.addEventListener('submit', function (event) {
        event.preventDefault();
        state.q = norm(searchEl.value);
        render();
      });
    }
  }

  // ?cat= и ?spot= — переход с главной и из карточки точки
  var params = new URLSearchParams(window.location.search);
  var catParam = params.get('cat');
  if (catParam && LABELS[catParam] && catRow) {
    state.cat = catParam;
    var target = catRow.querySelector('button[data-cat="' + catParam + '"]');
    if (target) setChip(catRow, target);
  }

  render();

  var spotParam = params.get('spot');
  if (spotParam && byId[spotParam]) {
    var found = spots.filter(function (s) {
      return s.id === spotParam;
    })[0];
    if (found) {
      if (state.cat && found.category !== state.cat) {
        state.cat = '';
        state.q = '';
        if (catRow) {
          var allChip = catRow.querySelector('button[data-cat=""]');
          if (allChip) setChip(catRow, allChip);
        }
        render();
      }
      map.setView([found.lat, found.lon], 16);
      byId[spotParam].openPopup();
      var card = listEl && listEl.querySelector('.spot-card[data-id="' + spotParam + '"]');
      if (card) card.classList.add('spot-card--active');
    }
  }

  if (listEl) {
    listEl.addEventListener('click', function (event) {
      var card = event.target.closest('.spot-card');
      if (!card) return;
      var id = card.dataset.id;
      if (byId[id]) {
        map.setView(byId[id].getLatLng(), 16);
        byId[id].openPopup();
      }
    });
  }
})();
