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

function radioDot(): string {
  return `<span class="radio-dot"><span class="radio-dot__inner"></span></span>`;
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
    </div>
  `;

  const list = document.createElement('div');
  list.className = 'radio-list';

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
    button.className = 'radio-item';
    button.dataset.optionId = option.id;
    button.innerHTML = `
      ${radioDot()}
      <span class="radio-item__label">${option.title}</span>
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
