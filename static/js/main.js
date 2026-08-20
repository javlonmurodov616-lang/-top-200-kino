/* ==========================================================================
   Top 200 eng zo'r kinolar - Frontend logikasi
   Funksiyalar:
     - Intro animatsiyasini tugatish
     - Kartalarning ketma-ket fade-up animatsiyasi
     - Poster yuklanmasa fallback (SVG plaseholder)
     - Janr bo'yicha filtr (smooth transition bilan)
     - Qidiruv (kino nomi bo'yicha)
     - Modal oynani ochish/yopish
   ========================================================================== */

(function () {
    'use strict';

    var grid = document.getElementById('grid');
    var cards = Array.prototype.slice.call(grid.querySelectorAll('.card'));
    var searchInput = document.getElementById('search');
    var genreButtons = Array.prototype.slice.call(document.querySelectorAll('.filter-btn'));
    var resultInfo = document.getElementById('resultInfo');

    // Joriy filtr holati
    var state = { genre: 'all', query: '' };
    // Yashirish animatsiyasi tugagach display:none qo'yish uchun vaqtlar
    var hideTimers = [];

    /* ------------------------------------------------------------------ */
    /* 1. Intro (kirish) animatsiyasi tugashi                              */
    /* ------------------------------------------------------------------ */

    var intro = document.getElementById('intro');
    if (intro) {
        setTimeout(function () {
            intro.classList.add('done');
            // Fade tugagach intro elementni butunlay olib tashlaymiz
            setTimeout(function () {
                if (intro.parentNode) intro.parentNode.removeChild(intro);
            }, 800);
        }, 3800);
    }

    /* ------------------------------------------------------------------ */
    /* 2. Kartalarning ketma-ket chiqish animatsiyasi                      */
    /* ------------------------------------------------------------------ */

    cards.forEach(function (card, i) {
        // Har bir karta o'zidan oldingisidan 55ms keyin chiqadi
        card.style.animationDelay = (i * 55) + 'ms';
        card.classList.add('animate-in');
    });

    /* ------------------------------------------------------------------ */
    /* 3. Poster yuklanmasa fallback                                       */
    /* ------------------------------------------------------------------ */

    // Poster o'rnida gradient fonli SVG plaseholder yaratamiz
    function placeholderSvg(title) {
        var letter = (title || 'KINO').charAt(0).toUpperCase();
        var svg =
            '<svg xmlns="http://www.w3.org/2000/svg" width="600" height="900" viewBox="0 0 600 900">' +
            '<rect width="600" height="900" fill="#15151f"/>' +
            '<circle cx="300" cy="320" r="160" fill="#f5c518" opacity="0.16"/>' +
            '<text x="300" y="410" font-family="Segoe UI, Arial, sans-serif" font-size="230" font-weight="800" fill="#f5c518" text-anchor="middle">' + letter + '</text>' +
            '<text x="300" y="790" font-family="Segoe UI, Arial, sans-serif" font-size="38" letter-spacing="8" fill="#8a92a6" text-anchor="middle">TOP 200</text>' +
            '</svg>';
        return 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(svg);
    }

    // Template'dan chaqiriladigan global funksiya (onerror atributida)
    window.posterFallback = function (img) {
        if (img.dataset.fallback) return; // faqat bir marta
        img.dataset.fallback = '1';
        img.src = placeholderSvg(img.alt);
        img.classList.add('poster-fallback');
    };

    /* ------------------------------------------------------------------ */
    /* 4. Janr filtri + qidiruv (birgalikda ishlaydi)                      */
    /* ------------------------------------------------------------------ */

    // Karta tanlangan janr va qidiruv so'ziga mos keladimi?
    function matches(card) {
        if (state.genre !== 'all') {
            var genres = (card.getAttribute('data-genre') || '').split(',').map(function (s) { return s.trim(); });
            if (genres.indexOf(state.genre) === -1) return false;
        }
        if (state.query) {
            var title = card.querySelector('.card-title').textContent.toLowerCase();
            if (title.indexOf(state.query) === -1) return false;
        }
        return true;
    }

    // Kartani silliq yashirish / ko'rsatish
    function setVisible(card, show) {
        var gone = card.classList.contains('card-gone');

        if (show) {
            if (gone) {
                // display:none bo'lgan kartani qayta ko'rsatamiz.
                // Reflow qilamiz, shunda fade-in transition ishlaydi.
                card.classList.remove('card-gone');
                void card.offsetWidth;
            }
            card.classList.remove('card-hide');
        } else {
            // Hali yashirilmagan kartaga fade-out klassini beramiz,
            // transition tugagach display:none qo'yamiz.
            if (!gone && !card.classList.contains('card-hide')) {
                card.classList.add('card-hide');
                hideTimers.push(setTimeout(function () {
                    card.classList.add('card-gone');
                }, 350));
            }
        }
    }

    function applyFilters() {
        // Eski vaqtlarni bekor qilamiz (tez-tez filtr almashtirilsa)
        hideTimers.forEach(clearTimeout);
        hideTimers = [];

        var visible = 0;
        cards.forEach(function (card) {
            var show = matches(card);
            if (show) visible++;
            setVisible(card, show);
        });

        // Natijalar haqidagi xabar
        var parts = [];
        if (state.genre !== 'all') parts.push('Janr: ' + state.genre);
        if (state.query) parts.push('Qidiruv: "' + state.query + '"');
        resultInfo.textContent = visible + " ta kino ko'rsatilmoqda" +
            (parts.length ? ' • ' + parts.join(' • ') : '');
    }

    // Janr tugmalariga bosish
    genreButtons.forEach(function (btn) {
        btn.addEventListener('click', function () {
            genreButtons.forEach(function (b) { b.classList.remove('active'); });
            btn.classList.add('active');
            state.genre = btn.getAttribute('data-genre');
            applyFilters();
        });
    });

    // Qidiruv maydoni
    searchInput.addEventListener('input', function () {
        state.query = this.value.trim().toLowerCase();
        applyFilters();
    });

    /* ------------------------------------------------------------------ */
    /* 5. Modal oyna                                                       */
    /* ------------------------------------------------------------------ */

    var modal = document.getElementById('modal');
    var backdrop = document.getElementById('modalBackdrop');
    var closeBtn = document.getElementById('modalClose');
    var lastFocused = null;

    // Reyting uchun yulduzlar (5 dan to'liq yulduzlar)
    function starsHtml(rating) {
        var full = Math.round(rating / 2);
        var html = '';
        for (var i = 1; i <= 5; i++) {
            html += i <= full ? '★' : '<span class="empty">☆</span>';
        }
        return html;
    }

    // Rank bo'yicha kino ma'lumotini topamiz (window.MOVIES JSON'da)
    function findMovie(rank) {
        if (!Array.isArray(window.MOVIES)) return null;
        for (var i = 0; i < window.MOVIES.length; i++) {
            if (window.MOVIES[i].rank === rank) return window.MOVIES[i];
        }
        return null;
    }

    function openModal(rank) {
        var movie = findMovie(rank);
        if (!movie) return;

        document.getElementById('modalRank').textContent = '#' + movie.rank;
        document.getElementById('modalTitle').textContent = movie.title;
        document.getElementById('modalYear').textContent = movie.year;
        document.getElementById('modalGenre').textContent = movie.genre.join(', ');
        document.getElementById('modalDuration').textContent = movie.duration;
        document.getElementById('modalStars').innerHTML = starsHtml(movie.rating);
        document.getElementById('modalRating').textContent = movie.rating.toFixed(1);
        document.getElementById('modalDesc').textContent = movie.description;
        document.getElementById('modalDirector').textContent = movie.director;

        var poster = document.getElementById('modalPoster');
        poster.dataset.fallback = '';
        poster.classList.remove('poster-fallback');
        poster.src = movie.poster;

        lastFocused = document.activeElement;
        modal.classList.add('open');
        backdrop.classList.add('open');
        modal.setAttribute('aria-hidden', 'false');
        closeBtn.focus();
    }

    function closeModal() {
        modal.classList.remove('open');
        backdrop.classList.remove('open');
        modal.setAttribute('aria-hidden', 'true');
        if (lastFocused) lastFocused.focus();
    }

    // Kartaga bosish va klaviatura (Enter/Prob) bilan ochish
    cards.forEach(function (card) {
        card.addEventListener('click', function () {
            openModal(parseInt(card.getAttribute('data-rank'), 10));
        });
        card.addEventListener('keydown', function (e) {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                openModal(parseInt(card.getAttribute('data-rank'), 10));
            }
        });
    });

    closeBtn.addEventListener('click', closeModal);
    backdrop.addEventListener('click', closeModal);

    // Esc tugmasi bilan yopish
    document.addEventListener('keydown', function (e) {
        if (e.key === 'Escape' && modal.classList.contains('open')) closeModal();
    });

    /* ------------------------------------------------------------------ */
    /* 6. Dastlabki ishga tushirish                                        */
    /* ------------------------------------------------------------------ */

    applyFilters();
})();
