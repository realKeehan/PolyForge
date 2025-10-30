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

const ICONS: Record<string, string> = {
  vanilla: `
    <svg viewBox="0 0 48 48" fill="none" aria-hidden="true">
      <path d="M12 18L24 12L36 18V30L24 36L12 30V18Z" stroke="currentColor" stroke-width="2.5" stroke-linejoin="round"></path>
      <path d="M12 18L24 24L36 18" stroke="currentColor" stroke-width="2.5" stroke-linejoin="round"></path>
      <path d="M24 24V36" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"></path>
    </svg>
  `,
  multimc: `
    <svg viewBox="0 0 48 48" fill="none" aria-hidden="true">
      <rect x="10" y="10" width="10" height="10" rx="2" stroke="currentColor" stroke-width="2.5"></rect>
      <rect x="18" y="18" width="14" height="14" rx="3" stroke="currentColor" stroke-width="2.5"></rect>
      <circle cx="34" cy="14" r="4" stroke="currentColor" stroke-width="2.5"></circle>
    </svg>
  `,
  curseforge: `
    <svg viewBox="0 0 48 48" fill="none" aria-hidden="true">
      <path d="M14 30H20L22 26H30L34 20H18L14 14" stroke="currentColor" stroke-width="2.5" stroke-linejoin="round" stroke-linecap="round"></path>
      <path d="M22 12V18" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"></path>
    </svg>
  `,
  modrinth: `
    <svg viewBox="0 0 48 48" fill="none" aria-hidden="true">
      <path d="M14 32V16H19L24 24L29 16H34V32" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
      <path d="M34 32L29 24L24 32L19 24L14 32" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"></path>
    </svg>
  `,
  gdlauncher: `
    <svg viewBox="0 0 48 48" fill="none" aria-hidden="true">
      <path d="M18 34L22 24L16 22L24 12L32 20L26 22L30 30L18 34Z" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
      <circle cx="24" cy="24" r="3" stroke="currentColor" stroke-width="2.5"></circle>
    </svg>
  `,
  atlauncher: `
    <svg viewBox="0 0 48 48" fill="none" aria-hidden="true">
      <path d="M24 12L36 30H28L24 36L20 30H12L24 12Z" stroke="currentColor" stroke-width="2.5" stroke-linejoin="round"></path>
      <path d="M24 36V30" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"></path>
    </svg>
  `,
  prismlauncher: `
    <svg viewBox="0 0 48 48" fill="none" aria-hidden="true">
      <path d="M24 10L38 34H10L24 10Z" stroke="currentColor" stroke-width="2.5" stroke-linejoin="round"></path>
      <path d="M24 16L32 30H16L24 16Z" stroke="currentColor" stroke-width="2.5" stroke-linejoin="round"></path>
    </svg>
  `,
  bakaxl: `
    <svg viewBox="0 0 48 48" fill="none" aria-hidden="true">
      <path d="M16 12H26C31 12 34 16 34 20C34 24 31 28 26 28H16V12Z" stroke="currentColor" stroke-width="2.5" stroke-linejoin="round"></path>
      <path d="M16 28H28C32 28 36 31 36 36H20" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"></path>
    </svg>
  `,
  feather: `
    <svg viewBox="0 0 48 48" fill="none" aria-hidden="true">
      <path d="M32 12C25 14 16 22 16 32C22 30 28 24 28 24C26 30 22 32 20 34C24 36 30 34 34 28C36 24 36 16 32 12Z" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
    </svg>
  `,
  technic: `
    <svg viewBox="0 0 48 48" fill="none" aria-hidden="true">
      <circle cx="20" cy="20" r="7" stroke="currentColor" stroke-width="2.5"></circle>
      <path d="M20 12L22 16L26 18L22 20L20 24L18 20L14 18L18 16L20 12Z" stroke="currentColor" stroke-width="2.5" stroke-linejoin="round"></path>
      <rect x="26" y="22" width="8" height="12" rx="2" stroke="currentColor" stroke-width="2.5"></rect>
    </svg>
  `,
  polymc: `
    <svg viewBox="0 0 48 48" fill="none" aria-hidden="true">
      <path d="M24 10L36 18V32L24 40L12 32V18L24 10Z" stroke="currentColor" stroke-width="2.5" stroke-linejoin="round"></path>
      <path d="M24 16L30 20V26L24 30L18 26V20L24 16Z" stroke="currentColor" stroke-width="2.5" stroke-linejoin="round"></path>
    </svg>
  `,
  custom: `
    <svg viewBox="0 0 48 48" fill="none" aria-hidden="true">
      <path d="M14 16H28L32 20V34H14V16Z" stroke="currentColor" stroke-width="2.5" stroke-linejoin="round"></path>
      <path d="M28 16V22H34" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"></path>
    </svg>
  `,
  manual: `
    <svg viewBox="0 0 48 48" fill="none" aria-hidden="true">
      <path d="M16 12H32V36H16C13.7909 36 12 34.2091 12 32V16C12 13.7909 13.7909 12 16 12Z" stroke="currentColor" stroke-width="2.5" stroke-linejoin="round"></path>
      <path d="M16 16H30M16 24H30M16 32H26" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"></path>
    </svg>
  `,
  about: `
    <svg viewBox="0 0 40 40" fill="none" aria-hidden="true">
      <circle cx="20" cy="20" r="17" stroke="currentColor" stroke-width="3"></circle>
      <path d="M20 28V18" stroke="currentColor" stroke-width="3" stroke-linecap="round"></path>
      <circle cx="20" cy="13" r="2.5" fill="currentColor"></circle>
    </svg>
  `,
  cake: `
    <svg viewBox="0 0 40 40" fill="none" aria-hidden="true">
      <path d="M10 20H30V30C30 31.1046 29.1046 32 28 32H12C10.8954 32 10 31.1046 10 30V20Z" stroke="currentColor" stroke-width="2.5" stroke-linejoin="round"></path>
      <path d="M10 20C13 20 14 18 16 18C18 18 19 20 22 20C25 20 26 18 28 18C30 18 31 20 34 20" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"></path>
      <path d="M20 10V16" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"></path>
      <path d="M18 12L20 10L22 12" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"></path>
    </svg>
  `,
};

function getIconForOption(optionId: string): string {
  return ICONS[optionId] ?? `
    <svg viewBox="0 0 40 40" fill="none" aria-hidden="true">
      <rect x="11" y="11" width="18" height="18" rx="4" stroke="currentColor" stroke-width="2.5"></rect>
      <path d="M11 19.5H29" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"></path>
      <path d="M20 11V29" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"></path>
    </svg>
  `;
}

export function renderInstaller(store: Store): HTMLElement {
  const container = document.createElement('section');
  container.className = 'screen screen--installer';

  const header = document.createElement('div');
  header.className = 'stage__header';
  header.innerHTML = `
    <span class="stage__icon">${LAUNCHER_ICON}</span>
    <div>
      <h2 class="stage__title">Choose Launcher</h2>
      <p class="stage__subtitle">Pick the launcher you want PolyForge to configure.</p>
    </div>
  `;

  const list = document.createElement('div');
  list.className = 'select-list';

  const pathField = document.createElement('div');
  pathField.className = 'field';
  pathField.hidden = true;
  pathField.innerHTML = `
    <label class="field__label" data-role="path-label">Installation path</label>
    <div class="field__row">
      <input class="field__input" type="text" data-role="path-input" placeholder="Select a folder" />
      <button type="button" class="btn btn--plain" data-role="browse">Browse</button>
    </div>
  `;

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
  runButton.disabled = !store.getState().selectedInstaller;

  actions.append(backButton, runButton);
  footer.append(social, actions);

  container.append(header, list, pathField, errorBox, footer);

  const optionButtons = new Map<string, HTMLButtonElement>();
  const pathLabel = pathField.querySelector('[data-role="path-label"]') as HTMLLabelElement;
  const pathInput = pathField.querySelector('[data-role="path-input"]') as HTMLInputElement;
  const browseButton = pathField.querySelector('[data-role="browse"]') as HTMLButtonElement;

  const setError = (message: string | null) => {
    if (!message) {
      errorBox.hidden = true;
      errorBox.textContent = '';
      return;
    }
    errorBox.hidden = false;
    errorBox.textContent = message;
  };

  const selectOption = (option: OptionDescriptor | undefined) => {
    store.setInstaller(option);
    setError(null);
    optionButtons.forEach((button) => button.classList.remove('is-active'));
    if (option) {
      const button = optionButtons.get(option.id);
      if (button) {
        button.classList.add('is-active');
      }
    }
    if (option && option.requiresPath) {
      pathField.hidden = false;
      pathLabel.textContent = option.pathLabel ?? 'Installation path';
      const existing = store.getState().selectedPath ?? '';
      if (existing) {
        pathInput.value = existing;
      }
      runButton.disabled = pathInput.value.trim() === '';
    } else {
      pathField.hidden = true;
      pathInput.value = '';
      store.setPath(undefined);
      runButton.disabled = !option;
    }
  };

  store.getState().options.forEach((option) => {
    const button = document.createElement('button');
    button.type = 'button';
    button.className = 'select-card';
    button.dataset.optionId = option.id;
    button.innerHTML = `
      <span class="select-card__icon">${getIconForOption(option.id)}</span>
      <span class="select-card__body">
        <span class="select-card__title">${option.title}</span>
        <span class="select-card__description">${option.description}</span>
      </span>
      ${option.requiresPath ? '<span class="select-card__badge">Path</span>' : ''}
    `;
    if (store.getState().selectedInstaller?.id === option.id) {
      button.classList.add('is-active');
    }
    button.addEventListener('click', () => {
      selectOption(option);
    });
    optionButtons.set(option.id, button);
    list.appendChild(button);
  });

  pathInput.addEventListener('input', () => {
    const value = pathInput.value.trim();
    store.setPath(value || undefined);
    runButton.disabled = value.length === 0;
  });

  browseButton.addEventListener('click', async () => {
    const selected = store.getState().selectedInstaller;
    const chosen = await browseForDirectory(selected?.pathLabel ?? 'Select folder');
    if (chosen) {
      pathInput.value = chosen;
      store.setPath(chosen);
      runButton.disabled = false;
    }
  });

  backButton.addEventListener('click', () => {
    store.setStep(Step.Modpack);
  });

  runButton.addEventListener('click', async () => {
    const option = store.getState().selectedInstaller;
    if (!option) {
      setError('Select an installer option to continue.');
      return;
    }
    if (option.requiresPath) {
      const path = pathInput.value.trim();
      if (!path) {
        setError('Please choose a directory for this launcher.');
        return;
      }
      store.setPath(path);
    } else {
      store.setPath(undefined);
    }

    setError(null);
    store.clearLogs();
    store.setBusy(true);

    try {
      const payload = option.requiresPath ? { path: store.getState().selectedPath } : {};
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

  const current = store.getState().selectedInstaller;
  if (current) {
    selectOption(current);
    if (current.requiresPath && store.getState().selectedPath) {
      pathInput.value = store.getState().selectedPath ?? '';
      runButton.disabled = false;
    }
  }

  return container;
}
