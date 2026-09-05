const CSRF = window.CSRF_TOKEN;

const searchInput = document.getElementById('search-input');
const mangaItems  = Array.from(document.querySelectorAll('.manga-item'));
const noResults   = document.getElementById('no-results');
const countEl     = document.getElementById('library-count');
const moreWrap    = document.getElementById('see-more-wrap');
const seeMoreBtn  = document.getElementById('see-more');
const seeAllBtn   = document.getElementById('see-all');

const PAGE_SIZE = 20;
let shown = PAGE_SIZE;

function normalizeText(text) {
    return text.normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLowerCase();
}

function getMatchScore(query, target) {
    if (!query) return 1;
    const q = normalizeText(query);
    const t = normalizeText(target);

    if (t === q) return 100;
    if (t.startsWith(q)) return 80;

    const index = t.indexOf(q);
    if (index !== -1) return 50 - index;

    let score = 0, qIdx = 0;
    for (let i = 0; i < t.length; i++) {
        if (t[i] === q[qIdx]) {
            score++;
            if (++qIdx === q.length) return score + 10;
        }
    }
    return 0;
}

function scoreItem(query, el) {
    return Math.max(
        getMatchScore(query, el.dataset.title || '') * 2,
        getMatchScore(query, el.dataset.tags || ''),
        getMatchScore(query, el.dataset.authors || ''),
        getMatchScore(query, el.querySelector('.manga-desc')?.textContent || '')
    );
}

function setCount(visible) {
    if (!countEl) return;
    countEl.textContent = mangaItems.length ? `${visible}/${mangaItems.length}` : '';
}

function render() {
    const query = searchInput.value.trim();

    // no query: plain library order, capped at `shown`
    if (!query) {
        mangaItems.forEach((el, i) => {
            el.style.order   = '';
            el.style.display = i < shown ? '' : 'none';
        });

        noResults.classList.add('hidden');

        const remaining = mangaItems.length - shown;
        moreWrap.classList.toggle('hidden', remaining <= 0);
        if (remaining > 0) seeMoreBtn.textContent = `see more (${remaining}) ⌄`;

        setCount(Math.min(shown, mangaItems.length));
        return;
    }

    // query: every match is shown, best score first
    let visible = 0;
    for (const el of mangaItems) {
        const score = scoreItem(query, el);
        if (score > 0) {
            el.style.display = '';
            el.style.order   = -Math.round(score);
            visible++;
        } else {
            el.style.display = 'none';
            el.style.order   = '';
        }
    }

    noResults.classList.toggle('hidden', visible > 0);
    moreWrap.classList.add('hidden');
    setCount(visible);
}

searchInput.addEventListener('input', render);

seeMoreBtn?.addEventListener('click', () => {
    shown = Math.min(shown + PAGE_SIZE, mangaItems.length);
    render();
});

seeAllBtn?.addEventListener('click', () => {
    shown = mangaItems.length;
    render();
});

render();

const readingList    = document.getElementById('reading-now-list');
const readingSection = document.getElementById('reading-now');
const readingCount   = document.getElementById('reading-now-count');

const mangaById = Object.fromEntries(
    mangaItems.map(el => [el.dataset.mangadexId, el.dataset])
);

function refreshReadingCount() {
    const n = readingList.children.length;
    readingCount.textContent = n ? `· ${n}` : '';
    if (!n) readingSection.classList.add('hidden');
}

function buildReadingCard(manga, entry) {
    const { title, cover } = manga;
    const vol  = entry.volume ?? entry.Volume;
    const page = entry.page   ?? entry.Page;
    const id   = entry.mangadex_id ?? entry.MangadexID;

    const link     = document.createElement('a');
    link.href      = `/manga?title=${encodeURIComponent(title)}`;
    link.className = 'shrink-0 flex flex-col gap-1.5 w-24 group';

    const thumb     = document.createElement('div');
    thumb.className = 'relative w-24 h-32 bg-neutral-800/40 border border-neutral-800 rounded overflow-hidden group-hover:border-neutral-600 transition';

    if (cover) {
        const img     = document.createElement('img');
        img.src       = cover;
        img.alt       = '';
        img.className = 'w-full h-full object-cover';
        thumb.appendChild(img);
    } else {
        thumb.textContent = 'no cover';
        thumb.className  += ' flex items-center justify-center text-neutral-600 text-[10px]';
    }

    const del       = document.createElement('button');
    del.className   = 'absolute top-1 right-1 w-5 h-5 bg-neutral-900/80 border border-neutral-700 rounded-sm text-neutral-400 hover:text-white hover:border-neutral-500 flex items-center justify-center transition text-[10px] leading-none';
    del.textContent = '✕';
    del.title       = 'remove from history';
    del.addEventListener('click', async e => {
        e.preventDefault();
        e.stopPropagation();
        try {
            await fetch(`/api/progress/${encodeURIComponent(id)}`, {
                method: 'DELETE',
                headers: { 'X-CSRF-Token': CSRF },
            });
            link.remove();
            refreshReadingCount();
        } catch { }
    });
    thumb.appendChild(del);

    const titleEl       = document.createElement('p');
    titleEl.className   = 'text-[10px] text-neutral-400 leading-tight truncate group-hover:text-neutral-200 transition';
    titleEl.textContent = title;

    const progressEl       = document.createElement('p');
    progressEl.className   = 'text-[10px] text-neutral-600';
    progressEl.textContent = `vol.${vol} · p.${page}`;

    link.append(thumb, titleEl, progressEl);
    return link;
}

(async () => {
    try {
        const res = await fetch('/api/progress');
        if (!res.ok) return;

        const entries = await res.json();
        if (!Array.isArray(entries) || !entries.length) return;

        for (const entry of entries) {
            const id    = entry.mangadex_id ?? entry.MangadexID;
            const manga = mangaById[id];
            if (!manga) continue;

            readingList.appendChild(buildReadingCard(manga, entry));
        }

        if (readingList.children.length) {
            readingCount.textContent = `· ${readingList.children.length}`;
            readingSection.classList.remove('hidden');
        }
    } catch { /* best-effort, doesn't break the dashboard */ }
})();