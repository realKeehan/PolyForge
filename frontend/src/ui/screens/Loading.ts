import { APP_VERSION } from '../../app/constants';
import type { Store } from '../../app/state';

const DEFAULT_DELAY = 35;
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
    { text: 'Welcome to KUMI!' },
    { text: `KUMI version ${APP_VERSION} boot at ${timestamp} ${timezone}`, delay: 120 },
    { text: 'loading startup sequence' },
    { text: 'Locating directories' },
    { text: 'Loading splash images' },
    { text: 'Checking for updates...', delay: 160 },
    { text: 'Quantumn time slicing stopwatch started', delay: 80 },
    { text: 'yadda yadda etc. process started', delay: 140 },
    { text: 'Checked on Jerry', delay: 90 },
    { text: 'Bootloaders Found', delay: 120 },
    { text: 'Bootleggers Spotted', delay: 140 },
    { text: 'Sent data to Server in Timbuktu', delay: 180 },
    { text: 'Reticulating splines' },
    { text: 'Configuring arcane runes' },
    { text: 'Polishing magic orbs' },
    { text: 'Re-aligning nether portals' },
    { text: 'John Died' },
    { text: '', delay: 500 },
    { text: 'Sent pipe bomb to the bungalow' },
    { text: 'Feeding the quantum hamsters' },
    { text: 'Whispering sweet nothings to the CPU fan' },
    { text: 'Tempering cosmic rays' },
    { text: 'Enchanting progress bars with glitter' },
    { text: 'Untangling spaghetti code' },
    { text: 'Defragmenting existential dread' },
    { text: 'Calibrating ducks in a row' },
    { text: 'Warming up flux capacitor' },
    { text: 'Compressing logs with duct tape' },
    { text: 'Negotiating with unpaid interns' },
    { text: 'Smoothing jagged timelines' },
    { text: 'Priming coffee machine responsibly', delay: 90 },
    { text: 'Counting turtles (and also turtels)' },
    { text: 'Checking insomnia logs for phantoms' },
    { text: 'Re-seating loose bolts' },
    { text: '' },
    { text: 'Optimizing interdimensional aerodynamics' },
    { text: 'Applying owl-based compression' },
    { text: 'Spinning up conspiracy API' },
    { text: 'Brewing speed potion (IRL edition)' },
    { text: "Consulting magic 8-ball: 'Signs point to yes'", delay: 110 },
    { text: 'Taming stray threads' },
    { text: 'Loading gremlins (do not feed after midnight)' },
    { text: 'Rehydrating dehydrated water' },
    { text: 'Installing extra RAM stickers' },
    { text: 'Teaching rubber ducks conflict resolution' },
    { text: 'Preheating furnace to 9001°', delay: 100 },
    { text: 'Aligning quantum foam with bedrock reality' },
    { text: "Verifying 'it works on my machine' certificate" },
    { text: 'Hydrating developers' },
    { text: 'Petting penguins' },
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
  container.className = 'screen screen--loading';
  container.innerHTML = `
    <header class="screen__header">
      <h2 class="screen__title">Booting PolyForge</h2>
      <p class="screen__subtitle">Hold tight while we warm up the installer.</p>
    </header>
    <div class="loading-console" data-role="console" aria-live="polite"></div>
  `;

  const consoleHost = container.querySelector('[data-role="console"]') as HTMLDivElement;
  const state = store.getState();

  state.loadingMessages.forEach((message) => {
    const line = document.createElement('div');
    line.className = 'loading-console__line';
    if (message.trim().length === 0) {
      line.classList.add('loading-console__line--empty');
      line.textContent = '\u00A0';
    } else {
      line.textContent = message;
    }
    consoleHost.appendChild(line);
  });

  if (state.loadingComplete) {
    consoleHost.classList.add('loading-console--complete');
  }

  queueMicrotask(() => {
    consoleHost.scrollTop = consoleHost.scrollHeight;
  });

  if (store.startLoading()) {
    runLoadingSequence(store);
  }

  return container;
}
