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
    .sort((a, b) => (b.num ?? -1) - (a.num ?? -1));

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
        const first = volumes[volumes.length - 1];
        label.textContent = first.label;
        sub.textContent = '';
        link.href = readerLink(first.filename);
        link.textContent = 'start reading';
    } else {
        return;
    }

    section.classList.remove('hidden');
}

searchInput.addEventListener('input', () => renderVolumes(window.__progress));

window.__progress = cfg.progress || null;
renderVolumes(window.__progress);
renderContinueReading(window.__progress);
