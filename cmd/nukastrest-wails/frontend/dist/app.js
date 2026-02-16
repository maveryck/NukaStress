const stressGifMap = {
  waiting: 'assets/esperandostart.gif',
  running: 'assets/stressejecutionwhite.gif',
  stopped: 'assets/testdetenido.gif',
};

let stressState = 'waiting';
let hasStartedOnce = false;
let applyTimer = null;

function wireAnimationFallback() {
  const gif = document.getElementById('pip-anim');
  const cfgGif = document.getElementById('cfg-anim');
  [gif, cfgGif].forEach((img) => {
    if (!img) return;
    img.addEventListener('error', () => {
      img.style.opacity = '0.18';
    });
    img.addEventListener('load', () => {
      img.style.opacity = '0.92';
    });
  });
}

function setStressState(next) {
  stressState = next;
  const gif = document.getElementById('pip-anim');
  if (!gif) return;
  const src = stressGifMap[next] || stressGifMap.waiting;
  if (!gif.src.endsWith(src)) gif.src = src;
}

function updateQuickButtons() {
  document.querySelectorAll('.preset-btn').forEach((btn) => {
    const key = btn.dataset.preset;
    btn.classList.toggle('active', !!toggles[key]);
  });
}

function wireWindowControls() {
  const rt = window.runtime;
  const btnMin = document.getElementById('win-min');
  const btnMax = document.getElementById('win-max');
  const btnClose = document.getElementById('win-close');

  if (btnMin) btnMin.addEventListener('click', () => rt?.WindowMinimise?.());
  if (btnMax) btnMax.addEventListener('click', () => rt?.WindowToggleMaximise?.());
  if (btnClose) btnClose.addEventListener('click', () => {
    if (rt?.Quit) rt.Quit();
    else window.close();
  });
}

const appApi = () => window?.go?.wailsapp?.App;

const toggles = {
  cpu: true,
  ram: true,
  gpu: false,
  disk: false,
};

const series = {
  cpu: [],
  ram: [],
  disk: [],
  temp: [],
};
const maxPoints = 90;

function pushPoint(key, value) {
  series[key].push(Number.isFinite(value) ? value : 0);
  if (series[key].length > maxPoints) series[key].shift();
}

function drawChart(canvasId, data, max = 100) {
  const canvas = document.getElementById(canvasId);
  if (!canvas) return;
  const ctx = canvas.getContext('2d');
  const w = canvas.width;
  const h = canvas.height;
  ctx.clearRect(0, 0, w, h);

  ctx.fillStyle = '#0d150d';
  ctx.fillRect(0, 0, w, h);

  ctx.strokeStyle = 'rgba(121,234,144,.22)';
  ctx.lineWidth = 1;
  for (let i = 1; i <= 4; i++) {
    const y = (h * i) / 5;
    ctx.beginPath();
    ctx.moveTo(0, y);
    ctx.lineTo(w, y);
    ctx.stroke();
  }

  if (!data.length) return;

  ctx.strokeStyle = '#89f3a0';
  ctx.lineWidth = 2;
  ctx.beginPath();
  data.forEach((v, i) => {
    const x = (i / Math.max(1, data.length - 1)) * w;
    const n = Math.max(0, Math.min(1, v / max));
    const y = h - n * h;
    if (i === 0) ctx.moveTo(x, y);
    else ctx.lineTo(x, y);
  });
  ctx.stroke();
}

function drawAllCharts() {
  drawChart('chart-cpu', series.cpu, 100);
  drawChart('chart-ram', series.ram, 100);
  drawChart('chart-disk', series.disk, 100);
  drawChart('chart-temp', series.temp, 120);
}

function renderFindings(findings) {
  const ul = document.getElementById('findings-list');
  if (!ul) return;
  ul.textContent = '';
  const rows = Array.isArray(findings) && findings.length
    ? findings
    : ['Sin fallos detectados en la ultima corrida'];

  rows.forEach((line) => {
    const li = document.createElement('li');
    li.textContent = line;
    ul.appendChild(li);
  });
}

function setStatusTheme(statusText = '', running = false) {
  const root = document.documentElement;
  const text = (statusText || '').toLowerCase();
  if (text.includes('critical') || text.includes('error') || text.includes('fail') || text.includes('[')) {
    root.style.setProperty('--status-bg', '#3a1515');
    root.style.setProperty('--status-fg', '#ffb6b6');
    return;
  }
  if (running) {
    root.style.setProperty('--status-bg', '#13311a');
    root.style.setProperty('--status-fg', '#a7ffba');
    return;
  }
  root.style.setProperty('--status-bg', '#143018');
  root.style.setProperty('--status-fg', '#9bf0ad');
}

function showScreen(id) {
  document.querySelectorAll('.screen').forEach((s) => s.classList.remove('active-screen'));
  document.querySelectorAll('.nav-link').forEach((l) => l.classList.remove('active'));
  const target = document.querySelector(id);
  if (target) target.classList.add('active-screen');
  const nav = document.querySelector(`a[href="${id}"]`);
  if (nav) nav.classList.add('active');
}

document.querySelectorAll('.nav-link').forEach((link) => {
  link.addEventListener('click', (e) => {
    e.preventDefault();
    showScreen(link.getAttribute('href'));
  });
});

function updateToggleUI() {
  document.querySelectorAll('.toggle-card').forEach((btn) => {
    const key = (btn.dataset.toggle || '').replace('cfg-', '');
    btn.classList.toggle('active', !!toggles[key]);
  });
  updateQuickButtons();
}

function scheduleApplyConfig(statusMsg = 'Config autoaplicada') {
  if (applyTimer) clearTimeout(applyTimer);
  applyTimer = setTimeout(async () => {
    await applyConfig();
    document.getElementById('status').textContent = statusMsg;
    setStatusTheme(statusMsg, false);
  }, 180);
}

function confirmDiskEnable() {
  return window.confirm('Advertencia: el test de DISK hace I/O intenso y puede degradar SSD/HDD o afectar datos temporales. No recomendado en equipo de uso diario. Continuar?');
}

function wireToggleCards() {
  document.querySelectorAll('.toggle-card').forEach((btn) => {
    btn.addEventListener('click', () => {
      const key = (btn.dataset.toggle || '').replace('cfg-', '');
      if (!key) return;

      const next = !toggles[key];
      if (key === 'disk' && next && !confirmDiskEnable()) {
        return;
      }

      toggles[key] = next;
      btn.classList.toggle('active', toggles[key]);
      updateQuickButtons();
      scheduleApplyConfig();
    });
  });
}

function wireAutoApplyControls() {
  const load = document.getElementById('cfg-load');
  const mins = document.getElementById('cfg-minutes');
  [load, mins].forEach((el) => {
    if (!el) return;
    el.addEventListener('input', () => scheduleApplyConfig());
    el.addEventListener('change', () => scheduleApplyConfig());
  });
}

document.querySelectorAll('.preset-btn').forEach((btn) => {
  btn.addEventListener('click', () => {
    const key = btn.dataset.preset;
    if (!key || !(key in toggles)) return;
    toggles[key] = !toggles[key];
    updateQuickButtons();
    updateToggleUI();
    scheduleApplyConfig(`${key.toUpperCase()} ${toggles[key] ? 'ON' : 'OFF'}`);
  });
});

async function refreshStatus() {
  const api = appApi();
  if (!api) return;
  try {
    const st = await api.GetStatus();
    const statusText = st.statusText || 'Idle';
    const running = !!st.running;
    document.getElementById('status').textContent = statusText;
    setStatusTheme(statusText, running);

    if (running) {
      hasStartedOnce = true;
      setStressState('running');
    } else if (!hasStartedOnce) {
      setStressState('waiting');
    } else if (stressState === 'running') {
      setStressState('stopped');
    }

    const cpu = Number(st.snapshot.cpuPercent || 0);
    const ram = Number(st.snapshot.memoryPercent || 0);
    const disk = Number(st.snapshot.diskPercent || 0);

    document.getElementById('results').textContent = st.lastResults || '';
    renderFindings(st.findings);
    document.getElementById('cpu').textContent = `${cpu.toFixed(1)}%`;
    document.getElementById('ram').textContent = `${ram.toFixed(1)}%`;
    document.getElementById('disk').textContent = `${disk.toFixed(1)}%`;

    document.getElementById('f-cpu').textContent = `CPU ${cpu.toFixed(1)}%`;
    document.getElementById('f-ram').textContent = `RAM ${ram.toFixed(1)}%`;
    document.getElementById('f-disk').textContent = `DISK ${disk.toFixed(1)}%`;

    pushPoint('cpu', cpu);
    pushPoint('ram', ram);
    pushPoint('disk', disk);

    if (st.snapshot.tempSupported) {
      const temp = Number(st.snapshot.temperatureC || 0);
      document.getElementById('temp').textContent = `${temp.toFixed(1)} C`;
      document.getElementById('temp-note').textContent = '';
      document.getElementById('f-temp').textContent = `TEMP ${temp.toFixed(1)}C`;
      pushPoint('temp', temp);
    } else {
      document.getElementById('temp').textContent = 'N/A';
      document.getElementById('temp-note').textContent = 'Temperatura no compatible en este equipo';
      document.getElementById('f-temp').textContent = 'TEMP N/A';
      pushPoint('temp', 0);
    }

    drawAllCharts();
  } catch (e) {
    console.error(e);
  }
}

async function applyConfig() {
  const api = appApi();
  if (!api) return;
  await api.SetConfig(
    toggles.cpu,
    toggles.gpu,
    toggles.ram,
    toggles.disk,
    false,
    false,
    parseInt(document.getElementById('cfg-load').value || '70', 10),
    parseInt(document.getElementById('cfg-minutes').value || '10', 10)
  );
}

document.getElementById('start-btn').addEventListener('click', async () => {
  const api = appApi();
  if (!api) return;
  await applyConfig();
  const msg = await api.StartStress();
  hasStartedOnce = true;
  setStressState('running');
  document.getElementById('status').textContent = msg;
  setStatusTheme(msg, true);
});

document.getElementById('stop-btn').addEventListener('click', async () => {
  const api = appApi();
  if (!api) return;
  const msg = await api.StopStress();
  hasStartedOnce = true;
  setStressState('stopped');
  document.getElementById('status').textContent = msg;
  setStatusTheme(msg, false);
});

document.querySelectorAll('.export-btn').forEach((btn) => {
  btn.addEventListener('click', async () => {
    const api = appApi();
    if (!api) return;
    const format = btn.dataset.format;
    try {
      const out = await api.ExportReport(format);
      document.getElementById('export-status').textContent = `Exportado: ${out}`;
    } catch (e) {
      document.getElementById('export-status').textContent = `Error: ${e}`;
    }
  });
});

wireWindowControls();
wireAnimationFallback();
wireToggleCards();
wireAutoApplyControls();
updateToggleUI();
setStressState('waiting');
setInterval(refreshStatus, 1500);
refreshStatus();
showScreen('#screen-stress');
