import { APP_VERSION } from '../../app/constants';
import type { Store } from '../../app/state';

const DEFAULT_DELAY = 45;
const FINAL_DELAY = 1000;

interface LoadingStep {
  text: string;
  delay?: number;
  highlight?: boolean;
}

function formatTimestamp(date: Date): string {
  const pad = (value: number) => value.toString().padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(
    date.getMinutes(),
  )}:${pad(date.getSeconds())}`;
}

function resolveTimezoneName(date: Date): string {
  const formatter = new Intl.DateTimeFormat(undefined, { timeZoneName: 'long' });
  const part = formatter
    .formatToParts(date)
    .find((item) => item.type === 'timeZoneName');
  return part?.value ?? Intl.DateTimeFormat().resolvedOptions().timeZone ?? 'Local Time';
}

function buildSteps(): LoadingStep[] {
  const now = new Date();
  const timestamp = formatTimestamp(now);
  const timezone = resolveTimezoneName(now);

  return [
    { text: `Welcome to PolyForge!`, highlight: true },
    { text: `PolyForge version ${APP_VERSION} boot at ${timestamp}`, delay: 100 },
    { text: `Timezone: ${timezone}` },
    { text: 'Loading active theme' },
    { text: 'Locating directories' },
    { text: 'Inspecting manifests' },
    { text: 'Checking for updates...' },
    { text: 'Download time slicing stopwatch started', delay: 80 },
    { text: 'rsubs bonds bkr, process started' },
    { text: 'Linking assets...' },
    { text: 'Bootleggers Bootleg' },
    { text: 'Boot User to Server in Timeouts', delay: 100 },
    { text: 'Reticulating splines' },
    { text: 'Configuring arcane runes' },
    { text: 'Alt-clicking nether portals' },
    { text: 'cake mode' },
    { text: '' },
    { text: 'Sort pigs best to the bongoing' },
    { text: 'Reading the quantum bookmarks', delay: 80 },
    { text: 'Deploying anti-monster 1s to kill the duo', delay: 160 },
  ];
}

function runLoadingSequence(store: Store) {
  const steps = buildSteps();
  let accumulated = 0;

  steps.forEach((step, index) => {
    accumulated += step.delay ?? DEFAULT_DELAY;
    window.setTimeout(() => {
      store.appendLoadingMessage(step.text);
      if (index === steps.length - 1) {
        window.setTimeout(() => {
          store.markLoadingComplete();
        }, FINAL_DELAY);
      }
    }, accumulated);
  });
}

export function renderLoading(store: Store): HTMLElement {
  const state = store.getState();
  const container = document.createElement('section');
  container.className = 'screen screen--startup';

  const terminal = document.createElement('div');
  terminal.className = 'terminal-log';

  state.loadingMessages.forEach((msg, index) => {
    const line = document.createElement('div');
    line.className = 'terminal-log__line';
    if (index === 0) {
      line.classList.add('terminal-log__line--highlight');
    }
    line.textContent = msg;
    terminal.appendChild(line);
  });

  // Add blinking cursor at the end
  if (!state.loadingComplete) {
    const cursorLine = document.createElement('div');
    cursorLine.className = 'terminal-log__line';
    const cursor = document.createElement('span');
    cursor.className = 'terminal-log__cursor';
    cursorLine.appendChild(cursor);
    terminal.appendChild(cursorLine);
  }

  container.appendChild(terminal);

  // Auto-scroll to bottom
  requestAnimationFrame(() => {
    terminal.scrollTop = terminal.scrollHeight;
  });

  if (store.startLoading()) {
    runLoadingSequence(store);
  }

  return container;
}
