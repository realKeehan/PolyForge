import { APP_VERSION } from '../../app/constants';
import type { Store } from '../../app/state';
import stdoutSfx from '../../assets/audio/stdout.wav';

const DEFAULT_DELAY = 45;
const FINAL_DELAY = 1000;

interface LoadingStep {
  text: string;
  delay?: number;
  highlight?: boolean;
}

function playStdout() {
  try {
    const audio = new Audio(stdoutSfx);
    audio.volume = 0.3;
    audio.play().catch(() => {});
  } catch {}
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
    { text: 'Welcome to PolyForge!', highlight: true },
    { text: `PolyForge version ${APP_VERSION} boot at ${timestamp} ${timezone}`, delay: 120 },
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
      if (step.text.length > 0) {
        playStdout();
      }
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
