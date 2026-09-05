import * as pdfjsLib from 'https://cdnjs.cloudflare.com/ajax/libs/pdf.js/4.3.136/pdf.min.mjs';
pdfjsLib.GlobalWorkerOptions.workerSrc =
  'https://cdnjs.cloudflare.com/ajax/libs/pdf.js/4.3.136/pdf.worker.min.mjs';

const cfg = window.READER_CONFIG;
const title = cfg.title;
const vol = cfg.vol;
const CSRF = cfg.csrf;
const MANGADEX_ID = cfg.mangadexId;

const readerContainer = document.getElementById('readerContainer');
const wrapper = document.getElementById('pagesWrapper');
const pageInput = document.getElementById('pageInput');
const totalPagesEl = document.getElementById('totalPages');
const zoomInBtn = document.getElementById('zoomIn');
const zoomOutBtn = document.getElementById('zoomOut');
const zoomLevelEl = document.getElementById('zoomLevel');
const progressBar = document.getElementById('progressBar');
const progressTrack = document.getElementById('progressTrack');
const loadingText = document.getElementById('loadingText');

const ZOOM_STEPS = [50, 60, 70, 80, 90, 100, 110, 120, 140, 160, 200];
const DEFAULT_ZOOM = 5;
let zoomIndex = DEFAULT_ZOOM;

function applyZoom() {
  const pct = ZOOM_STEPS[zoomIndex];
  zoomLevelEl.textContent = `${pct}%`;

  const naturalWidth = Math.min(720, readerContainer.clientWidth);
  const targetWidth = Math.round(naturalWidth * pct / 100);
  wrapper.style.width = `${targetWidth}px`;

  readerContainer.style.alignItems =
    targetWidth > readerContainer.clientWidth ? 'flex-start' : 'center';

  const isMin = zoomIndex === 0;
  const isMax = zoomIndex === ZOOM_STEPS.length - 1;

  zoomOutBtn.disabled = isMin;
  zoomInBtn.disabled = isMax;
  zoomOutBtn.classList.toggle('opacity-30', isMin);
  zoomOutBtn.classList.toggle('cursor-not-allowed', isMin);
  zoomInBtn.classList.toggle('opacity-30', isMax);
  zoomInBtn.classList.toggle('cursor-not-allowed', isMax);
}

const IS_MOBILE = matchMedia('(pointer: coarse)').matches || navigator.maxTouchPoints > 0;
const MAX_CANVAS_PIXELS = IS_MOBILE ? 5_500_000 : 16_000_000;
const RENDER_KEEP_RADIUS = 2;

const containers = new Map();
const renderedPages = new Map();
const renderTasks = new Map();
let saveTimer;
let resizeTimer;
let currentPage = 1;
let pdf;

setLoading(true);
try {
  pdf = await pdfjsLib.getDocument(
    `/pdf?title=${encodeURIComponent(title)}&vol=${encodeURIComponent(vol)}`
  ).promise;

  applyZoom();

  const savedPage = await fetchProgress();
  await showPage(parseInt(savedPage) || 1);
  setLoading(false);
} catch (err) {
  console.error('Error loading PDF:', err);
  showError();
}

function getOrCreateContainer(num) {
  let div = containers.get(num);
  if (div) return div;

  div = document.createElement('div');
  div.id = `page-container-${num}`;
  div.dataset.page = num;
  div.className = 'w-full hidden';
  div.style.contain = 'layout paint';

  const canvas = document.createElement('canvas');
  canvas.id = `canvas-${num}`;
  canvas.className = 'w-full block';
  div.appendChild(canvas);

  wrapper.appendChild(div);
  containers.set(num, div);
  return div;
}

function removePage(num) {
  renderTasks.get(num)?.cancel();
  renderTasks.delete(num);
  renderedPages.delete(num);
  containers.get(num)?.remove();
  containers.delete(num);
}

async function triggerRender(num) {
  if (renderedPages.has(num) || renderTasks.has(num)) return;
  renderTasks.set(num, null);
  try {
    const page = await pdf.getPage(num);
    const container = getOrCreateContainer(num);
    const canvas = container.querySelector('canvas');

    const dpr = Math.min(window.devicePixelRatio || 1, 2);
    const rawVp = page.getViewport({ scale: 1 });
    const maxZoom = ZOOM_STEPS[ZOOM_STEPS.length - 1] / 100;
    let scale = (720 / rawVp.width) * dpr * maxZoom;
    const area = rawVp.width * rawVp.height * scale * scale;
    if (area > MAX_CANVAS_PIXELS) scale *= Math.sqrt(MAX_CANVAS_PIXELS / area);
    const viewport = page.getViewport({ scale });

    canvas.width = viewport.width;
    canvas.height = viewport.height;
    container.style.aspectRatio = `${rawVp.width} / ${rawVp.height}`;

    const task = page.render({ canvasContext: canvas.getContext('2d'), viewport });
    renderTasks.set(num, task);
    await task.promise;

    if (Math.abs(num - currentPage) > RENDER_KEEP_RADIUS) {
      removePage(num);
    } else {
      renderedPages.set(num, 'rendered');
    }
  } catch (err) {
    if (err?.name !== 'RenderingCancelledException') {
      console.warn(`Render failed on page ${num}`, err);
    }
  } finally {
    renderTasks.delete(num);
  }
}

function evictFarPages() {
  for (const num of containers.keys()) {
    if (Math.abs(num - currentPage) > RENDER_KEEP_RADIUS) removePage(num);
  }
}

async function showPage(num) {
  if (!pdf || num < 1 || num > pdf.numPages) return;

  containers.get(currentPage)?.classList.add('hidden');
  currentPage = num;
  getOrCreateContainer(currentPage).classList.remove('hidden');

  readerContainer.scrollTop = 0;
  readerContainer.scrollLeft = 0;

  updateProgress(currentPage);

  evictFarPages();
  await triggerRender(currentPage);
  if (currentPage < pdf.numPages) triggerRender(currentPage + 1);
}


async function fetchProgress() {
  if (!MANGADEX_ID) return 1;
  try {
    const res = await fetch(`/api/progress/${MANGADEX_ID}`);
    if (!res.ok) return 1;
    const data = await res.json();
    return String(data.Volume) === String(vol) ? (data.Page ?? 1) : 1;
  } catch {
    return 1;
  }
}

function sendProgress(page) {
  fetch('/api/progress', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': CSRF },
    body: JSON.stringify({ mangadex_id: MANGADEX_ID, volume: vol, page }),
    keepalive: true,
  }).catch(() => { });
}

function saveProgress(page) {
  if (!MANGADEX_ID) return;
  clearTimeout(saveTimer);
  saveTimer = setTimeout(() => sendProgress(page), 1500);
}

function flushProgress() {
  if (!MANGADEX_ID) return;
  clearTimeout(saveTimer);
  sendProgress(currentPage);
}

document.addEventListener('visibilitychange', () => {
  if (document.visibilityState === 'hidden') flushProgress();
});
window.addEventListener('pagehide', flushProgress);

function updateProgress(num) {
  pageInput.value = num;
  totalPagesEl.textContent = pdf.numPages;
  saveProgress(num);
  const pct = Math.max(0, Math.min(100, (num / pdf.numPages) * 100));
  progressBar.style.width = `${pct}%`;
}

function setLoading(on) {
  loadingText.classList.toggle('hidden', !on);
  progressTrack.classList.toggle('progress-loading', on);
  pageInput.disabled = on;
  if (on) {
    pageInput.value = '';
    totalPagesEl.textContent = '...';
  } else {
    pageInput.max = pdf.numPages;
  }
}

function showError() {
  loadingText.textContent = 'ERROR LOADING PDF';
  loadingText.classList.remove('hidden');
  progressTrack.classList.remove('progress-loading');
  progressBar.style.width = '100%';
  progressBar.classList.replace('bg-neutral-300', 'bg-red-400');
}

function handlePageInput() {
  const val = parseInt(pageInput.value, 10);
  if (isNaN(val)) { pageInput.value = currentPage; return; }
  showPage(Math.max(1, Math.min(pdf.numPages, val)));
}

zoomInBtn.addEventListener('click', () => {
  if (zoomIndex < ZOOM_STEPS.length - 1) { zoomIndex++; applyZoom(); }
});
zoomOutBtn.addEventListener('click', () => {
  if (zoomIndex > 0) { zoomIndex--; applyZoom(); }
});

window.addEventListener('resize', () => {
  clearTimeout(resizeTimer);
  resizeTimer = setTimeout(applyZoom, 150);
});

readerContainer.addEventListener('click', (e) => {
  if (e.target.closest('button, a, input')) return;
  showPage(e.clientX < window.innerWidth / 2 ? currentPage - 1 : currentPage + 1);
});

pageInput.addEventListener('keydown', (e) => {
  if (e.key === 'Enter') { handlePageInput(); pageInput.blur(); }
});
pageInput.addEventListener('blur', handlePageInput);

document.addEventListener('keydown', (e) => {
  if (e.target.tagName === 'INPUT') return;
  const k = e.key;
  if (k === 'x' || k === 'X') zoomInBtn.click();
  if (k === 'z' || k === 'Z') zoomOutBtn.click();
  if (k === 'ArrowLeft') showPage(currentPage - 1);
  if (k === 'ArrowRight') showPage(currentPage + 1);
});
