import { APP_VERSION } from '../../app/constants';
import type { Store } from '../../app/state';
import splashImage from '../../assets/splash.png';

const DEFAULT_DELAY = 45;
const FINAL_DELAY = 1000;

interface LoadingStep {
  text: string;
  delay?: number;
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
    { text: 'PolyForge launcher waking up...' },
    { text: `Runtime version ${APP_VERSION} initialised at ${timestamp} (${timezone})`, delay: 140 },
    { text: 'Seeding installer pipeline' },
    { text: 'Scanning launchers' },
    { text: 'Inspecting profile manifests' },
    { text: 'Warming up renderer', delay: 120 },
    { text: 'Fetching manifest signatures' },
    { text: 'Syncing trusted mirrors' },
    { text: 'Linking shared assets' },
    { text: 'Finalising splash sequence', delay: 160 },
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
  const container = document.createElement('section');
  container.className = 'screen screen--startup';
  container.innerHTML = `
    <div class="loading-splash">
      <img class="loading-splash__image" src="${splashImage}" alt="PolyForge bootstrap splash" draggable="false" />
    </div>
  `;

  if (store.startLoading()) {
    runLoadingSequence(store);
  }

  return container;
}
