import { APP_VERSION } from '../app/constants';
import { Quit, WindowMinimise } from '@wailsapp/runtime';
import { createStore } from '../app/state';
import { Step } from '../app/types';
import { fetchMenuOptions } from '../app/ipc';
import { renderLoading } from './screens/Loading';
import { renderLicense } from './screens/License';
import { renderMode } from './screens/Mode';
import { renderModpack } from './screens/Modpack';
import { renderInstaller } from './screens/Installer';
import { renderStatus } from './screens/Status';
import brandIcon from '../assets/app-icon.png';

export async function createApp(root: HTMLElement) {
  const store = createStore();

  const shell = document.createElement('div');
  shell.className = 'app-shell';

  const header = document.createElement('header');
  header.className = 'app-shell__topbar';
  header.innerHTML = `
    <div class="topbar__brand" role="presentation">
      <span class="topbar__glyph" aria-hidden="true" style="background-image: url('${brandIcon}');"></span>
      <span class="topbar__title">KUMI</span>
      <span class="topbar__version">v${APP_VERSION}</span>
    </div>
    <div class="topbar__controls" role="toolbar" aria-label="Window controls">
      <button type="button" class="window-btn window-btn--min" aria-label="Minimize window">
        <svg class="window-btn__icon" viewBox="0 0 12 12" width="12" height="12" aria-hidden="true" focusable="false">
          <rect x="2" y="5.3" width="8" height="1.4" rx="0.7" />
        </svg>
      </button>
      <button type="button" class="window-btn window-btn--close" aria-label="Close window">
        <svg class="window-btn__icon" viewBox="0 0 12 12" width="12" height="12" aria-hidden="true" focusable="false">
          <path d="M3.2 3.2l5.6 5.6M8.8 3.2l-5.6 5.6" stroke-width="1.6" stroke-linecap="round" />
        </svg>
      </button>
    </div>
  `;

  const minimiseBtn = header.querySelector('.window-btn--min') as HTMLButtonElement;
  const closeBtn = header.querySelector('.window-btn--close') as HTMLButtonElement;

  minimiseBtn.addEventListener('click', () => {
    try {
      WindowMinimise();
    } catch (error) {
      console.error('Failed to minimise window', error);
    }
  });

  closeBtn.addEventListener('click', () => {
    try {
      Quit();
    } catch (error) {
      console.error('Failed to close window', error);
    }
  });


  const contentHost = document.createElement('main');
  contentHost.className = 'app-shell__content';

  const overlay = document.createElement('div');
  overlay.className = 'app-shell__overlay';
  overlay.innerHTML = '<div class="overlay__spinner"></div>';
  overlay.hidden = true;

  shell.append(header, contentHost, overlay);
  root.appendChild(shell);

  const render = () => {
    const state = store.getState();
    overlay.hidden = !state.busy;

    let screen: HTMLElement;
    switch (state.step) {
      case Step.Loading:
        screen = renderLoading(store);
        break;
      case Step.License:
        screen = renderLicense(store);
        break;
      case Step.Mode:
        screen = renderMode(store);
        break;
      case Step.Modpack:
        screen = renderModpack(store);
        break;
      case Step.Installer:
        screen = renderInstaller(store);
        break;
      case Step.Status:
      default:
        screen = renderStatus(store);
        break;
    }

    contentHost.replaceChildren(screen);
  };

  store.subscribe(render);

  const loadingReady = store.waitForLoadingComplete();

  try {
    const options = await fetchMenuOptions();
    store.setOptions(options);
    await loadingReady;
    store.setStep(Step.License);
  } catch (error) {
    console.error('Failed to load menu options', error);
    store.setOptions([]);
    store.appendLogs([
      { level: 'error', message: 'Unable to load installer options from backend. Please restart the application.' },
    ]);
    store.setResult({ success: false, messages: store.getState().logs });
    await loadingReady;
    store.setStep(Step.Status);
  }
}
