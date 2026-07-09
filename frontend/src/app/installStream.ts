// Live install progress, shared between the run controller (which receives
// backend events) and the Status screen (which renders the bar). Progress is
// kept OUT of the store on purpose: byte-level download events fire far too
// often to drive a full screen re-render each time, so we snapshot the latest
// value here and patch the on-screen bar directly. Log lines are low-frequency
// and still flow through the store as normal.

export interface InstallProgress {
  active: boolean;
  percent: number;
  label: string;
  indeterminate: boolean;
  failed: boolean;
}

let progress: InstallProgress = {
  active: false,
  percent: 0,
  label: '',
  indeterminate: false,
  failed: false,
};

export function getInstallProgress(): InstallProgress {
  return progress;
}

/** Begin a fresh install: an active, indeterminate bar with a starting label. */
export function beginInstallProgress(label = 'Starting…'): void {
  progress = { active: true, percent: 0, label, indeterminate: true, failed: false };
  paint();
}

/** Apply a backend progress event (percent < 0 or indeterminate => animated). */
export function updateInstallProgress(input: { percent?: number; label?: string; indeterminate?: boolean }): void {
  const indeterminate = input.indeterminate === true || (input.percent ?? 0) < 0;
  progress = {
    active: true,
    failed: false,
    indeterminate,
    percent: indeterminate ? progress.percent : Math.max(0, Math.min(100, input.percent ?? 0)),
    label: input.label ?? progress.label,
  };
  paint();
}

/** Mark the install finished; the bar stays visible as a done/failed summary. */
export function finishInstallProgress(success: boolean): void {
  progress = {
    active: true,
    percent: 100,
    indeterminate: false,
    failed: !success,
    label: success ? 'Done' : 'Failed',
  };
  paint();
}

/** Hide the bar entirely (e.g. leaving the status screen for a new run). */
export function clearInstallProgress(): void {
  progress = { active: false, percent: 0, label: '', indeterminate: false, failed: false };
  paint();
}

function escapeText(value: string): string {
  const div = document.createElement('div');
  div.textContent = value;
  return div.innerHTML;
}

/** Renders the current progress into a `.install-progress` container element. */
export function renderInstallProgress(root: HTMLElement): void {
  root.hidden = !progress.active;
  root.classList.toggle('install-progress--indeterminate', progress.indeterminate);
  root.classList.toggle('install-progress--error', progress.failed);
  const pct = Math.round(progress.percent);
  root.innerHTML = `
    <div class="install-progress__row">
      <span class="install-progress__label">${escapeText(progress.label)}</span>
      <span class="install-progress__pct">${progress.indeterminate ? '' : `${pct}%`}</span>
    </div>
    <div class="install-progress__track">
      <div class="install-progress__bar" style="width:${progress.indeterminate ? 100 : pct}%"></div>
    </div>
  `;
}

// Patches the live bar in place, if one is currently mounted, without going
// through the store (avoids a full re-render per byte-progress event).
function paint(): void {
  const root = document.querySelector('.install-progress');
  if (root instanceof HTMLElement) {
    renderInstallProgress(root);
  }
}
