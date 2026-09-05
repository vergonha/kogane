const CSRF = window.CSRF_TOKEN;

const searchInput = document.getElementById('search-input');
const mangaItems  = Array.from(document.querySelectorAll('.manga-item'));
const noResults   = document.getElementById('no-results');

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

searchInput.addEventListener('input', e => {
    const query = e.target.value.trim();
    const container = document.getElementById('manga-list');

    if (!query) {
        mangaItems.forEach(el => { el.style.display = ''; el.style.order = ''; });
        noResults.classList.add('hidden');
        container.classList.remove('flex', 'flex-col');
        return;
    }

    let hasVisible = false;
    mangaItems
        .map(el => ({
            el,
            score: Math.max(
                getMatchScore(query, el.dataset.title || '') * 2,
                getMatchScore(query, el.querySelector('.manga-desc')?.textContent || '')
            ),
        }))
        .forEach(({ el, score }) => {
            if (score > 0) {
                el.style.display = '';
                el.style.order   = -Math.round(score);
                hasVisible       = true;
            } else {
                el.style.display = 'none';
            }
        });

    container.classList.toggle('flex', hasVisible);
    container.classList.toggle('flex-col', hasVisible);
    noResults.classList.toggle('hidden', hasVisible);
});

const mangaById = Object.fromEntries(
    mangaItems.map(el => [el.dataset.mangadexId, el.dataset])
);

function buildReadingCard(manga, entry) {
    const { title, cover } = manga;
    const vol  = entry.volume  ?? entry.Volume;
    const page = entry.page    ?? entry.Page;
    const id   = entry.mangadex_id ?? entry.MangadexID;

    const link       = document.createElement('a');
    link.href        = `/read?title=${encodeURIComponent(title)}&vol=${encodeURIComponent(vol)}&mangadex_id=${encodeURIComponent(id)}`;
    link.className   = 'shrink-0 flex flex-col gap-1.5 w-24 group';

    const thumb      = document.createElement('div');
    thumb.className  = 'relative w-24 h-32 bg-neutral-800/40 border border-neutral-800 rounded overflow-hidden group-hover:border-neutral-600 transition';

    if (cover) {
        const img    = document.createElement('img');
        img.src      = cover;
        img.alt      = '';
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
            if (!list.children.length) section.classList.add('hidden');
        } catch { }
    });
    thumb.appendChild(del);

    const titleEl      = document.createElement('p');
    titleEl.className  = 'text-[10px] text-neutral-400 leading-tight truncate group-hover:text-neutral-200 transition';
    titleEl.textContent = title;

    const progressEl      = document.createElement('p');
    progressEl.className  = 'text-[10px] text-neutral-600';
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

        const list    = document.getElementById('reading-now-list');
        const section = document.getElementById('reading-now');

        for (const entry of entries) {

            const id    = entry.mangadex_id ?? entry.MangadexID;
            const manga = mangaById[id];
            if (!manga) continue;

            list.appendChild(buildReadingCard(manga, entry));
        }

        if (list.children.length) section.classList.remove('hidden');
    } catch { /* best-effort, doesn't break the dashboard */ }
})();
