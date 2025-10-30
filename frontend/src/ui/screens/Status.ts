import type { Store } from '../../app/state';
import { Step } from '../../app/types';
import { createSocialLinks } from '../components/social';

const STATUS_ICON = `
  <svg viewBox="0 0 40 40" fill="none" aria-hidden="true">
    <path d="M11 27.4875V12.1859C11 10.9787 11.9787 10 13.1859 10H27.8318C28.194 10 28.4875 10.2936 28.4875 10.6558V24.9893" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round"></path>
    <path d="M15.3721 10V18.7438L18.1045 16.995L20.8369 18.7438V10" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"></path>
    <path d="M13.1855 25.3016H28.4871" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round"></path>
    <path d="M13.1855 29.6735H28.4871" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round"></path>
    <path d="M13.1859 29.6735C11.9787 29.6735 11 28.6948 11 27.4875C11 26.2802 11.9787 25.3016 13.1859 25.3016" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"></path>
  </svg>
`;

function buildLogText(store: Store): string {
  const logEntries = store.getState().logs;
  if (logEntries.length === 0) {
    return 'Logs will appear here once an installer has run.';
  }
  return logEntries
    .map((entry) => `[${entry.level.toUpperCase()}] ${entry.message}`)
    .join('\n');
}

function copyLogs(text: string): Promise<void> {
  if (navigator.clipboard?.writeText) {
    return navigator.clipboard.writeText(text);
  }
  const node = document.createElement('textarea');
  node.value = text;
  node.style.position = 'fixed';
  node.style.opacity = '0';
  document.body.appendChild(node);
  node.select();
  document.execCommand('copy');
  document.body.removeChild(node);
  return Promise.resolve();
}

export function renderStatus(store: Store): HTMLElement {
  const container = document.createElement('section');
  container.className = 'screen screen--status';

  const header = document.createElement('div');
  header.className = 'stage__header';
  header.innerHTML = `
    <span class="stage__icon">${STATUS_ICON}</span>
    <div>
      <h2 class="stage__title">Status Log</h2>
      <p class="stage__subtitle">Review the output from the previous action.</p>
    </div>
  `;

  const result = store.getState().lastResult;
  const isSuccess = result?.success ?? false;

  const banner = document.createElement('div');
  banner.className = `status-banner ${isSuccess ? '' : 'status-banner--error'}`.trim();
  banner.innerHTML = isSuccess
    ? '<span>Installer completed successfully.</span>'
    : '<span>Installer finished with warnings or errors.</span>';

  const logPanel = document.createElement('pre');
  logPanel.className = 'log-panel';
  logPanel.textContent = buildLogText(store);

  const copyButton = document.createElement('button');
  copyButton.type = 'button';
  copyButton.className = 'copy-button log-panel__copy';
  copyButton.textContent = 'Copy';

  const footer = document.createElement('footer');
  footer.className = 'screen-footer';
  const social = createSocialLinks();
  const actions = document.createElement('div');
  actions.className = 'screen-footer__actions';

  const backButton = document.createElement('button');
  backButton.type = 'button';
  backButton.className = 'btn btn--ghost';
  backButton.textContent = 'Back';

  const closeButton = document.createElement('button');
  closeButton.type = 'button';
  closeButton.className = 'btn btn--primary';
  closeButton.textContent = isSuccess ? 'Done' : 'Dismiss';

  actions.append(backButton, closeButton);
  footer.append(social, actions);

  container.append(header, banner, logPanel, footer);
  logPanel.appendChild(copyButton);

  copyButton.addEventListener('click', async () => {
    const original = copyButton.textContent;
    try {
      await copyLogs(logPanel.textContent ?? '');
      copyButton.textContent = 'Copied!';
    } catch (error) {
      console.error('Failed to copy logs', error);
      copyButton.textContent = 'Failed';
    } finally {
      window.setTimeout(() => {
        copyButton.textContent = original ?? 'Copy';
      }, 1500);
    }
  });

  backButton.addEventListener('click', () => {
    store.setStep(Step.Installer);
  });

  closeButton.addEventListener('click', () => {
    store.setStep(Step.Mode);
  });

  return container;
}
