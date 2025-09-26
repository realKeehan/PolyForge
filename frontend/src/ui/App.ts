import { APP_VERSION } from '../app/constants';
import { createStore } from '../app/state';
import { Step } from '../app/types';
import { fetchMenuOptions } from '../app/ipc';
import { renderLoading } from './screens/Loading';
import { renderLicense } from './screens/License';
import { renderMode } from './screens/Mode';
import { renderModpack } from './screens/Modpack';
import { renderInstaller } from './screens/Installer';
import { renderStatus } from './screens/Status';

export async function createApp(root: HTMLElement) {
  const store = createStore();

  const shell = document.createElement('div');
  shell.className = 'app-shell';

  const header = document.createElement('header');
  header.className = 'app-shell__header';
  header.innerHTML = `
    <div class="brand">
      <span class="brand__mark">KUMI</span>
      <span class="brand__version">v${APP_VERSION}</span>
    </div>
    <div class="brand__subtitle">Keehan&apos;s Universal Modpack Installer</div>
  `;

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
