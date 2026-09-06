const cfg = window.MANGA_CONFIG;
const TITLE = cfg.title;
const MANGADEX_ID = cfg.mangadexId;

const searchInput = document.getElementById('volume-search');
const listEl = document.getElementById('volume-list');
const noResultsEl = document.getElementById('volume-no-results');

function volumeNumber(filename) {
    const match = filename.match(/vol(?:ume)?\.?\s*0*(\d+)/i);
    return match ? parseInt(match[1], 10) : null;
}

function sortKey(filename) {
    return (filename.match(/\d+/g) || []).map(n => parseInt(n, 10));
}

function compareSortKeys(a, b) {
    const len = Math.max(a.length, b.length);
    for (let i = 0; i < len; i++) {
        const diff = (a[i] ?? -1) - (b[i] ?? -1);
        if (diff !== 0) return diff;
    }
    return 0;
}

function volumeLabel(filename, num) {
    // if (num !== null) return `Vol. ${String(num).padStart(2, '0')}`;
    // return filename.replace(/\.pdf$/i, '');

    return filename
}

function readerLink(filename) {
    return `/read?title=${encodeURIComponent(TITLE)}&vol=${encodeURIComponent(filename)}`;
}

const volumes = (cfg.volumes || [])
    .map(filename => {
        const num = volumeNumber(filename);
        return { filename, num, label: volumeLabel(filename, num) };
    })
    .sort((a, b) => compareSortKeys(sortKey(a.filename), sortKey(b.filename)));

function renderVolumes(progress) {
    const query = searchInput.value.trim().toLowerCase();

    listEl.innerHTML = '';
    let visibleCount = 0;

    for (const v of volumes) {
        if (query && !v.label.toLowerCase().includes(query)) continue;
        visibleCount++;

        const row = document.createElement('a');
        row.href = readerLink(v.filename);
        row.className = 'flex items-center justify-between gap-3 py-2.5 text-sm text-neutral-300 hover:text-white hover:bg-neutral-800/20 transition px-1 -mx-1 rounded';

        const left = document.createElement('span');
        left.className = 'flex items-center gap-3 min-w-0';

        const num = document.createElement('span');
        num.className = 'text-neutral-600 text-xs w-8 shrink-0';
        num.textContent = v.num !== null ? `#${String(v.num).padStart(2, '0')}` : '#—';

        const label = document.createElement('span');
        label.className = 'truncate';
        label.textContent = v.label;

        left.append(num, label);

        const right = document.createElement('span');
        right.className = 'flex items-center gap-3 shrink-0';

        if (progress && String(progress.Volume) === String(v.num)) {
            const status = document.createElement('span');
            status.className = 'text-[10px] text-neutral-500';
            status.textContent = progress.Completed ? 'Completed' : `Page ${progress.Page}`;
            right.appendChild(status);
        }

        const arrow = document.createElement('span');
        arrow.className = 'text-neutral-600';
        arrow.textContent = '>';
        right.appendChild(arrow);

        row.append(left, right);
        listEl.appendChild(row);
    }

    noResultsEl.classList.toggle('hidden', visibleCount > 0);
}

function renderContinueReading(progress) {
    const section = document.getElementById('continue-reading-section');
    const label = document.getElementById('continue-reading-label');
    const sub = document.getElementById('continue-reading-sub');
    const link = document.getElementById('continue-reading-link');

    const matched = progress && volumes.find(v => String(v.filename) === String(progress.Volume));

    if (matched) {
        label.textContent = matched.label;
        sub.textContent = progress.Completed ? 'Completed' : `Page ${progress.Page}`;
        link.href = readerLink(matched.filename);
        link.textContent = 'Continue reading';
    } else if (volumes.length) {
        const first = volumes[0];
        label.textContent = first.label;
        sub.textContent = '';
        link.href = readerLink(first.filename);
        link.textContent = 'start reading';
    } else {
        return;
    }

    section.classList.remove('hidden');
}

function setupMangaHeader() {
    const skeleton = document.getElementById('header-skeleton');
    const content = document.getElementById('header-content');
    if (!content) return;

    let revealed = false;
    function reveal() {
        if (revealed) return;
        revealed = true;
        content.classList.remove('invisible');
        if (skeleton) skeleton.classList.add('hidden');
    }

    const failsafe = setTimeout(reveal, 4000);

    const cover = document.getElementById('cover-img');
    const block = document.getElementById('description-block');
    const desc = document.getElementById('manga-description');
    const toggle = document.getElementById('description-toggle');
    const chevron = document.getElementById('description-chevron');

    if (!cover || !block || !desc || !toggle) {
        clearTimeout(failsafe);
        reveal();
        return;
    }

    const desktop = window.matchMedia('(min-width: 640px)');
    let expanded = false;
    let coverFailed = false;

    const coverReady = () => coverFailed || (cover.complete && cover.naturalHeight > 0);

    function unclamp() {
        desc.style.removeProperty('display');
        desc.style.removeProperty('-webkit-box-orient');
        desc.style.removeProperty('-webkit-line-clamp');
        desc.style.removeProperty('overflow');
    }

    function clampTo(lines) {
        desc.style.setProperty('display', '-webkit-box');
        desc.style.setProperty('-webkit-box-orient', 'vertical');
        desc.style.setProperty('-webkit-line-clamp', String(lines));
        desc.style.setProperty('overflow', 'hidden');
    }

    function done() {
        if (coverReady()) {
            clearTimeout(failsafe);
            reveal();
        }
    }

    function apply() {
        if (!desktop.matches || expanded) {
            unclamp();
            toggle.classList.toggle('hidden', !desktop.matches);
            chevron.style.transform = expanded ? 'rotate(180deg)' : '';
            done();
            return;
        }

        unclamp();
        chevron.style.transform = '';
        toggle.classList.remove('hidden', 'invisible');

        const lh = parseFloat(getComputedStyle(desc).lineHeight) || 16;
        const gap = parseFloat(getComputedStyle(block).rowGap) || 0;
        const available = cover.getBoundingClientRect().bottom
            - desc.getBoundingClientRect().top
            - toggle.getBoundingClientRect().height - gap;

        const lines = Math.floor(available / lh);

        if (lines < 2 || desc.scrollHeight <= lines * lh + 1) {
            toggle.classList.add('invisible');
        } else {
            clampTo(lines);
        }

        done();
    }

    let frame = null;
    const schedule = () => {
        if (frame) cancelAnimationFrame(frame);
        frame = requestAnimationFrame(apply);
    };

    toggle.addEventListener('click', () => {
        expanded = !expanded;
        toggle.setAttribute('aria-expanded', String(expanded));
        toggle.setAttribute('aria-label', expanded ? 'Collapse description' : 'Expand description');
        apply();
    });

    cover.addEventListener('load', schedule);
    cover.addEventListener('error', () => {
        coverFailed = true;
        unclamp();
        toggle.classList.add('hidden');
        clearTimeout(failsafe);
        reveal();
    });

    new ResizeObserver(schedule).observe(cover);
    window.addEventListener('resize', schedule);
    desktop.addEventListener('change', schedule);
    if (document.fonts?.ready) document.fonts.ready.then(schedule);
    schedule();
}

setupMangaHeader();

searchInput.addEventListener('input', () => renderVolumes(window.__progress));

window.__progress = cfg.progress || null;
renderVolumes(window.__progress);
renderContinueReading(window.__progress);
