import { browseForDirectory, runInstaller } from '../../app/ipc';
import type { Store } from '../../app/state';
import { Step, type OptionDescriptor } from '../../app/types';
import { createSocialLinks } from '../components/social';

const LAUNCHER_ICON = `
  <svg viewBox="0 0 40 40" fill="none" aria-hidden="true">
    <path d="M11 27.4875V12.1859C11 10.9787 11.9787 10 13.1859 10H27.8318C28.194 10 28.4875 10.2936 28.4875 10.6558V24.9893" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round"></path>
    <path d="M15.3721 10V18.7438L18.1045 16.995L20.8369 18.7438V10" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"></path>
    <path d="M13.186 25.3016H28.4876" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round"></path>
    <path d="M13.186 29.6735H28.4876" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round"></path>
    <path d="M13.1859 29.6735C11.9787 29.6735 11 28.6948 11 27.4875C11 26.2802 11.9787 25.3016 13.1859 25.3016" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"></path>
  </svg>
`;

const FOLDER_ICON = `<svg viewBox="0 0 20 20" fill="none" aria-hidden="true"><path d="M3 5a2 2 0 012-2h3.172a2 2 0 011.414.586l1.828 1.828A2 2 0 0012.828 6H15a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2V5z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/></svg>`;

// Test dummy launcher option
const TEST_DUMMY: OptionDescriptor = {
  id: '__test_dummy__',
  title: 'Test',
  description: 'Dummy test launcher for development',
  requiresPath: false,
};

// Dummy log messages for test
const DUMMY_LOG_MESSAGES = [
  { level: 'info' as const, message: 'Starting Install...' },
  { level: 'info' as const, message: '' },
  { level: 'info' as const, message: 'Creating required directories...' },
  { level: 'info' as const, message: '✅ Directory exists: C:\\Users\\USER\\AppData\\Roaming\\BetterMinecraft' },
  { level: 'info' as const, message: '✅ Directory exists: C:\\Users\\USER\\AppData\\Roaming\\BetterMinecraft\\data' },
  { level: 'info' as const, message: '✅ Directory exists: C:\\Users\\USER\\AppData\\Roaming\\BetterMinecraft\\themes' },
  { level: 'info' as const, message: '✅ Directory exists: C:\\Users\\USER\\AppData\\Roaming\\BetterMinecraft\\plugins' },
  { level: 'info' as const, message: '✅ Directories created' },
  { level: 'info' as const, message: '' },
  { level: 'info' as const, message: 'Downloading asar file' },
  { level: 'info' as const, message: '✅ Downloaded BetterMinecraft version undefined from the official website' },
  { level: 'info' as const, message: '✅ Package downloaded' },
  { level: 'info' as const, message: '' },
  { level: 'info' as const, message: 'Injecting shims...' },
  { level: 'info' as const, message: 'Injecting into: C:\\Users\\USER\\AppData\\Local\\Minecraft\\app-1.0.9211\\modules\\Minecraft_desktop_core-1\\Minecraft_desktop_core' },
  { level: 'info' as const, message: '✅ Injection successful' },
  { level: 'info' as const, message: '✅ Shims injected' },
  { level: 'info' as const, message: '' },
  { level: 'info' as const, message: 'Restarting Minecraft...' },
  { level: 'info' as const, message: 'Attempting to kill Minecraft' },
  { level: 'warning' as const, message: '✅ Minecraft not running' },
  { level: 'info' as const, message: '✅ Minecraft restarted' },
  { level: 'info' as const, message: '' },
  { level: 'info' as const, message: 'Install completed!' },
];

// Options to exclude from launcher list
const EXCLUDED_OPTIONS = ['about', 'cake'];

function radioDot(): string {
  return `<span class="radio-dot"><span class="radio-dot__inner"></span></span>`;
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

// Per-option paths chosen via Browse. Module-scoped so the user's choices
// survive screen re-renders (navigating back and forth) within a session.
const overriddenPaths = new Map<string, string>();

export function renderInstaller(store: Store): HTMLElement {
  const container = document.createElement('section');
  container.className = 'screen screen--installer';

  // Track selection locally to avoid triggering re-renders (which reset scroll)
  let localSelected: OptionDescriptor | undefined = store.getState().selectedInstaller;

  const header = document.createElement('div');
  header.className = 'stage__header';
  header.innerHTML = `
    <span class="stage__icon">${LAUNCHER_ICON}</span>
    <div>
      <h2 class="stage__title">Choose Launcher</h2>
    </div>
  `;

  const list = document.createElement('div');
  list.className = 'radio-list';

  const errorBox = document.createElement('div');
  errorBox.className = 'alert';
  errorBox.hidden = true;

  const footer = document.createElement('footer');
  footer.className = 'screen-footer';
  const social = createSocialLinks();
  const actions = document.createElement('div');
  actions.className = 'screen-footer__actions';

  const backButton = document.createElement('button');
  backButton.type = 'button';
  backButton.className = 'btn btn--ghost';
  backButton.textContent = 'Back';

  const runButton = document.createElement('button');
  runButton.type = 'button';
  runButton.className = 'btn btn--primary';
  runButton.textContent = 'Install';
  runButton.disabled = !localSelected;

  actions.append(backButton, runButton);
  footer.append(social, actions);

  // Old-style target folder chooser for custom/manual
  const pathChooser = document.createElement('div');
  pathChooser.className = 'field';
  pathChooser.hidden = true;
  pathChooser.innerHTML = `
    <label class="field__label">Target Folder</label>
    <div class="field__row">
      <input type="text" class="field__input" placeholder="Select a folder" readonly />
      <button type="button" class="btn btn--plain">Browse</button>
    </div>
  `;
  const pathInput = pathChooser.querySelector('.field__input') as HTMLInputElement;
  const pathBrowseBtn = pathChooser.querySelector('.btn--plain') as HTMLButtonElement;

  container.append(header, list, pathChooser, errorBox, footer);

  const optionButtons = new Map<string, HTMLButtonElement>();

  const setError = (message: string | null) => {
    if (!message) {
      errorBox.hidden = true;
      errorBox.textContent = '';
      return;
    }
    errorBox.hidden = false;
    errorBox.textContent = message;
  };

  // Local-only selection update — does NOT touch the store, so no re-render
  const selectOption = (option: OptionDescriptor | undefined) => {
    localSelected = option;
    setError(null);

    optionButtons.forEach((button) => button.classList.remove('is-active'));
    if (option) {
      const button = optionButtons.get(option.id);
      if (button) button.classList.add('is-active');
    }

    // Show/hide old-style path chooser for custom/manual
    const isCustomManual = option && (option.id === 'custom' || option.id === 'manual');
    pathChooser.hidden = !isCustomManual;
    if (isCustomManual) {
      const existingPath = overriddenPaths.get(option!.id) || '';
      pathInput.value = existingPath;
    }

    // Enable install if option selected and either doesn't need path, or has a path
    if (option) {
      if (option.requiresPath || (option.id === 'custom' || option.id === 'manual')) {
        const path = overriddenPaths.get(option.id) || option.detectedPath || '';
        runButton.disabled = path.trim() === '';
      } else {
        runButton.disabled = false;
      }
    } else {
      runButton.disabled = true;
    }
  };

  // Path chooser browse button handler
  pathBrowseBtn.addEventListener('click', async () => {
    if (!localSelected) return;
    const chosen = await browseForDirectory(localSelected.pathLabel ?? 'Select folder');
    if (chosen) {
      overriddenPaths.set(localSelected.id, chosen);
      pathInput.value = chosen;
      runButton.disabled = false;
    }
  });

  // Filter out excluded options and add real backend options
  const filteredOptions = store.getState().options.filter(
    (opt) => !EXCLUDED_OPTIONS.includes(opt.id.toLowerCase()),
  );

  // Build combined list: backend options + test dummy
  const allOptions: OptionDescriptor[] = [...filteredOptions, TEST_DUMMY];

  // Sort: found options first, then not-found, preserving original order within each group
  const sortedOptions = [
    ...allOptions.filter((o) => o.found),
    ...allOptions.filter((o) => !o.found),
  ];

  sortedOptions.forEach((option) => {
    const isFound = !!option.found;
    const displayPath = option.detectedPath || '';
    const isSpecial = option.id === 'custom' || option.id === 'manual' || option.id === '__test_dummy__';

    // Restore any path the user already browsed to this session
    const restoredPath = overriddenPaths.get(option.id);
    const missing = !isFound && !isSpecial && !restoredPath;

    const button = document.createElement('button');
    button.type = 'button';
    button.className = `radio-item radio-item--launcher radio-item--has-bg${missing ? ' radio-item--not-found' : ''}`;
    button.dataset.optionId = option.id;

    const pathText = isSpecial
      ? ''
      : restoredPath || (isFound ? displayPath : 'Not Found');

    // The browse control is a span[role=button] because interactive content
    // cannot be nested inside the row's <button> in valid HTML.
    button.innerHTML = `
      ${radioDot()}
      <span class="radio-item__body">
        <span class="radio-item__label">${escapeHtml(option.title)}</span>
        ${pathText ? `<span class="radio-item__path" data-path-for="${option.id}">${escapeHtml(pathText)}</span>` : ''}
      </span>
      ${!isSpecial ? `<span role="button" tabindex="0" class="radio-item__browse" data-browse-for="${option.id}" aria-label="Browse for ${escapeHtml(option.title)}" title="Browse">${FOLDER_ICON}</span>` : ''}
    `;

    if (localSelected?.id === option.id) {
      button.classList.add('is-active');
    }

    // Main click selects the option
    button.addEventListener('click', (e) => {
      // Don't select if the browse button was clicked
      if ((e.target as HTMLElement).closest('.radio-item__browse')) return;
      selectOption(option);
    });

    // Browse control (span[role=button]) — handle both click and keyboard
    const browseBtn = button.querySelector(`[data-browse-for="${option.id}"]`) as HTMLSpanElement | null;
    if (browseBtn) {
      const browse = async (e: Event) => {
        e.stopPropagation();
        const chosen = await browseForDirectory(option.pathLabel ?? 'Select folder');
        if (chosen) {
          overriddenPaths.set(option.id, chosen);
          // Update path display
          const pathEl = button.querySelector(`[data-path-for="${option.id}"]`);
          if (pathEl) {
            pathEl.textContent = chosen;
            pathEl.classList.remove('radio-item__path--not-found');
          }
          // Remove not-found style
          button.classList.remove('radio-item--not-found');
          // Select this option
          selectOption(option);
        }
      };
      browseBtn.addEventListener('click', browse);
      browseBtn.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          void browse(e);
        }
      });
    }

    const pathEl = button.querySelector(`[data-path-for="${option.id}"]`) as HTMLSpanElement | null;
    if (pathEl && missing) {
      pathEl.classList.add('radio-item__path--not-found');
    }

    optionButtons.set(option.id, button);
    list.appendChild(button);
  });

  backButton.addEventListener('click', () => {
    // Commit selection to store before leaving
    store.setInstaller(localSelected);
    store.setStep(Step.Modpack);
  });

  runButton.addEventListener('click', async () => {
    const option = localSelected;
    if (!option) {
      setError('Select an installer option to continue.');
      return;
    }

    // Commit selection to store
    store.setInstaller(option);

    // Handle test dummy launcher
    if (option.id === TEST_DUMMY.id) {
      store.clearLogs();
      store.appendLogs(DUMMY_LOG_MESSAGES);
      store.setResult({ success: true, messages: DUMMY_LOG_MESSAGES });
      store.setStep(Step.Status);
      return;
    }

    // Determine path
    const overridden = overriddenPaths.get(option.id);
    const effectivePath = overridden || option.detectedPath || '';
    const needsPath = option.requiresPath || option.id === 'custom' || option.id === 'manual';

    if (needsPath) {
      if (!effectivePath) {
        setError('Please choose a directory for this launcher.');
        return;
      }
      store.setPath(effectivePath);
    } else {
      store.setPath(undefined);
    }

    setError(null);
    store.clearLogs();
    store.setBusy(true);

    try {
      const payload = needsPath ? { path: store.getState().selectedPath } : {};
      const result = await runInstaller(option.id, payload);
      store.appendLogs(result.messages);
      store.setResult(result);
      store.setStep(Step.Status);
    } catch (error) {
      console.error('Failed to run installer', error);
      setError('Something went wrong while running the installer. Check the logs for details.');
      store.setResult(null);
    } finally {
      store.setBusy(false);
    }
  });

  // Restore selection state (local only, no store emit)
  if (localSelected) {
    selectOption(localSelected);
  }

  return container;
}
