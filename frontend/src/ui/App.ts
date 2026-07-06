import { APP_VERSION } from '../app/constants';
import { Quit, WindowMinimise } from '@wailsapp/runtime';
import { createStore } from '../app/state';
import { Step, type OptionDescriptor, type RemoteContentResult } from '../app/types';
import { fetchMenuOptions, fetchRemoteContent, inspectPolyPack, launchedPackPath } from '../app/ipc';
import { LOCAL_PACK_ID } from '../app/types';
import { renderLoading } from './screens/Loading';
import { renderStartup } from './screens/Startup';
import { renderLicense } from './screens/License';
import { renderMode } from './screens/Mode';
import { renderModpack } from './screens/Modpack';
import { renderInstaller } from './screens/Installer';
import { renderStatus } from './screens/Status';
import brandIcon from '../assets/app-icon.ico';

const KONAMI_CODE = [
  'ArrowUp', 'ArrowUp', 'ArrowDown', 'ArrowDown',
  'ArrowLeft', 'ArrowRight', 'ArrowLeft', 'ArrowRight',
  'KeyB', 'KeyA',
];

const EASTER_EGG_VIDEO = 'https://keehan.co/KUMI_Files/NiceComputer.mp4';

function setupKonamiCode(shell: HTMLElement) {
  let konamiIndex = 0;

  document.addEventListener('keydown', (event) => {
    if (event.code === KONAMI_CODE[konamiIndex]) {
      konamiIndex++;
      if (konamiIndex === KONAMI_CODE.length) {
        konamiIndex = 0;
        showEasterEgg(shell);
      }
    } else {
      konamiIndex = 0;
    }
  });
}

function showEasterEgg(shell: HTMLElement) {
  const overlay = document.createElement('div');
  overlay.className = 'easter-egg-overlay';
  overlay.innerHTML = `
    <video class="easter-egg-video" autoplay disablepictureinpicture disableremoteplayback playsinline>
      <source src="${EASTER_EGG_VIDEO}" type="video/mp4" />
    </video>
  `;

  overlay.addEventListener('click', (event) => {
    if (event.target === overlay) {
      const video = overlay.querySelector('video');
      if (video) {
        video.pause();
        video.src = '';
      }
      overlay.remove();
    }
  });

  const video = overlay.querySelector('video') as HTMLVideoElement;
  video.addEventListener('ended', () => {
    overlay.remove();
  });

  shell.appendChild(overlay);
}

/** Merge remote option overrides / disabled options into the backend list. */
function applyRemoteOverrides(
  options: OptionDescriptor[],
  remote: RemoteContentResult | null,
): OptionDescriptor[] {
  const manifest = remote?.manifest;
  if (!manifest) return options;

  const disabled = new Set(manifest.disabledOptions ?? []);
  const overrides = new Map((manifest.optionOverrides ?? []).map((o) => [o.id, o]));

  return options
    .filter((option) => !disabled.has(option.id))
    .map((option) => {
      const override = overrides.get(option.id);
      if (!override) return option;
      return {
        ...option,
        title: override.title ?? option.title,
        description: override.description ?? option.description,
      };
    });
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

/** Prompt for a binary update. Mandatory updates cannot be dismissed. */
function showUpdateDialog(shell: HTMLElement, remote: RemoteContentResult) {
  const app = remote.manifest?.app;
  if (!app) return;

  const overlay = document.createElement('div');
  overlay.className = 'update-dialog-overlay';
  overlay.innerHTML = `
    <div class="update-dialog" role="alertdialog" aria-modal="true" aria-labelledby="updateTitle">
      <h2 class="update-dialog__title" id="updateTitle">
        ${remote.mandatory ? 'Update required' : 'Update available'}
      </h2>
      <p class="update-dialog__text">
        PolyForge v${escapeHtml(app.latestVersion)} is available - you are on v${escapeHtml(remote.currentVersion)}.
        ${remote.mandatory ? 'This version is no longer supported; please update to continue.' : ''}
      </p>
      ${app.notes ? `<p class="update-dialog__notes">${escapeHtml(app.notes)}</p>` : ''}
      <div class="update-dialog__actions">
        <button type="button" class="btn btn--primary" data-action="download">Get update</button>
        ${remote.mandatory ? '' : '<button type="button" class="btn btn--ghost" data-action="later">Later</button>'}
      </div>
    </div>
  `;

  const downloadBtn = overlay.querySelector('[data-action="download"]') as HTMLButtonElement;
  downloadBtn.addEventListener('click', () => {
    const url = app.downloadUrl || 'https://polyforge.dev/downloads';
    const runtimeApi = (window as unknown as { runtime?: { BrowserOpenURL?: (u: string) => void } }).runtime;
    if (runtimeApi?.BrowserOpenURL) {
      runtimeApi.BrowserOpenURL(url);
    } else {
      window.open(url, '_blank');
    }
  });

  const laterBtn = overlay.querySelector('[data-action="later"]');
  laterBtn?.addEventListener('click', () => overlay.remove());

  shell.appendChild(overlay);
}

export async function createApp(root: HTMLElement) {
  const store = createStore();

  // Default to install mode
  store.setMode('install');

  const frame = document.createElement('div');
  frame.className = 'app-root';

  const shell = document.createElement('div');
  shell.className = 'app-window';

  const header = document.createElement('header');
  header.className = 'app-header';
  header.innerHTML = `
    <div class="app-header__side" role="presentation">
      <img class="app-header__logo" src="${brandIcon}" alt="PolyForge logo" draggable="false" />
    </div>
    <div class="app-header__center" role="presentation">
      <span class="app-header__title">PolyForge v${APP_VERSION}</span>
    </div>
    <div class="app-header__controls" role="toolbar" aria-label="Window controls">
      <button type="button" class="window-control window-control--minimise" data-action="minimise" aria-label="Minimise window">
        <svg viewBox="0 0 26 2" width="18" height="18" aria-hidden="true" focusable="false">
          <path d="M1 1H25" stroke="currentColor" stroke-width="2" stroke-linecap="round"></path>
        </svg>
      </button>
      <button type="button" class="window-control window-control--close" data-action="close" aria-label="Close window">
        <svg viewBox="0 0 14 14" width="16" height="16" aria-hidden="true" focusable="false">
          <path d="M1 13L13 1M13 13L1 1" stroke="currentColor" stroke-width="2" stroke-linecap="round"></path>
        </svg>
      </button>
    </div>
  `;

  const minimiseBtn = header.querySelector('[data-action="minimise"]') as HTMLButtonElement;
  const closeBtn = header.querySelector('[data-action="close"]') as HTMLButtonElement;

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
  contentHost.className = 'app-content';

  const overlay = document.createElement('div');
  overlay.className = 'app-overlay';
  overlay.innerHTML = `
    <div class="overlay__panel" role="status" aria-live="polite">
      <span class="overlay__spinner" aria-hidden="true"></span>
      <span class="overlay__label">Working...</span>
    </div>
  `;
  overlay.hidden = true;

  shell.append(header, contentHost, overlay);
  frame.appendChild(shell);
  root.appendChild(frame);

  // Setup Konami code easter egg (works on any screen)
  setupKonamiCode(shell);

  const render = () => {
    const state = store.getState();
    overlay.hidden = !state.busy;

    let screen: HTMLElement;
    switch (state.step) {
      case Step.Loading:
        screen = renderLoading(store);
        break;
      case Step.Startup:
        screen = renderStartup(store);
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
    const [options, remote] = await Promise.all([
      fetchMenuOptions(),
      fetchRemoteContent().catch((error): RemoteContentResult | null => {
        console.warn('Remote content unavailable, using built-in defaults', error);
        return null;
      }),
    ]);

    store.setOptions(applyRemoteOverrides(options, remote));
    if (remote?.manifest?.modpacks?.length) {
      store.setModpacks(remote.manifest.modpacks);
    }

    // If launched by double-clicking a .polypack, pre-load it so the
    // modpack screen shows it selected and ready to install.
    const packPath = await launchedPackPath();
    if (packPath) {
      try {
        const info = await inspectPolyPack(packPath);
        store.setLocalPack(info);
        store.setModpack(LOCAL_PACK_ID);
      } catch (error) {
        console.error('Could not read the opened pack', error);
      }
    }

    await loadingReady;
    store.setStep(Step.Startup);

    if (remote?.updateAvailable || remote?.mandatory) {
      showUpdateDialog(shell, remote);
    }
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
