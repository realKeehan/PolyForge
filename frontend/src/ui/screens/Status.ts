import type { Store } from '../../app/state';
import { Step } from '../../app/types';

export function renderStatus(store: Store): HTMLElement {
  const container = document.createElement('section');
  container.className = 'screen screen--status';

  const state = store.getState();
  const success = state.lastResult?.success ?? false;

  const statusClass = success ? 'status-banner--success' : 'status-banner--error';
  const statusText = success ? 'Installer completed successfully.' : 'Installer finished with warnings or errors.';

  const logs = document.createElement('ul');
  logs.className = 'log-view';
  state.logs.forEach((entry) => {
    const item = document.createElement('li');
    item.className = `log-view__entry log-view__entry--${entry.level}`;
    item.innerHTML = `
      <span class="log-view__badge">${entry.level}</span>
      <span class="log-view__message">${entry.message}</span>
    `;
    logs.appendChild(item);
  });

  const restart = document.createElement('button');
  restart.type = 'button';
  restart.className = 'btn btn--primary';
  restart.textContent = success ? 'Run another installer' : 'Try again';
  restart.addEventListener('click', () => {
    store.setStep(Step.Installer);
  });

  const close = document.createElement('button');
  close.type = 'button';
  close.className = 'btn btn--ghost';
  close.textContent = 'Back to action selection';
  close.addEventListener('click', () => {
    store.setStep(Step.Mode);
  });

  const actions = document.createElement('footer');
  actions.className = 'screen__actions';
  actions.append(close, restart);

  container.innerHTML = `
    <header class="screen__header">
      <h2 class="screen__title">Status Log</h2>
      <div class="status-banner ${statusClass}">
        <span class="status-banner__icon"></span>
        <span class="status-banner__text">${statusText}</span>
      </div>
    </header>
  `;
  container.appendChild(logs);
  container.appendChild(actions);

  return container;
}
