import renderOption from '../../templates/optionTemplate';
import type { Store } from '../../app/state';
import { Step, type OptionDescriptor } from '../../app/types';
import { browseForDirectory, runInstaller } from '../../app/ipc';

export function renderInstaller(store: Store): HTMLElement {
  const container = document.createElement('section');
  container.className = 'screen screen--installer';

  const optionsHost = document.createElement('div');
  optionsHost.className = 'installer-grid';

  const current = store.getState().selectedInstaller;

  const optionButtons = new Map<string, HTMLButtonElement>();

  store.getState().options.forEach((option) => {
    const html = renderOption(option);
    const template = document.createElement('template');
    template.innerHTML = html.trim();
    const button = template.content.firstElementChild as HTMLButtonElement;
    if (current && current.id === option.id) {
      button.classList.add('is-active');
    }
    optionButtons.set(option.id, button);
    optionsHost.appendChild(button);
  });

  const details = document.createElement('div');
  details.className = 'installer-details';
  details.innerHTML = `
    <div class="field" data-role="path-field" hidden>
      <label class="field__label" data-role="path-label">Installation path</label>
      <div class="field__row">
        <input type="text" class="field__input" data-role="path-input" placeholder="Select a folder" />
        <button type="button" class="btn btn--ghost" data-role="browse">Browse…</button>
      </div>
    </div>
    <div class="alert" data-role="error" hidden></div>
    <footer class="screen__actions">
      <button type="button" class="btn btn--ghost" data-role="back">Back</button>
      <button type="button" class="btn btn--primary" data-role="run" disabled>Start Install</button>
    </footer>
  `;

  const pathField = details.querySelector('[data-role="path-field"]') as HTMLDivElement;
  const pathLabel = details.querySelector('[data-role="path-label"]') as HTMLLabelElement;
  const pathInput = details.querySelector('[data-role="path-input"]') as HTMLInputElement;
  const browseBtn = details.querySelector('[data-role="browse"]') as HTMLButtonElement;
  const errorBox = details.querySelector('[data-role="error"]') as HTMLDivElement;
  const runBtn = details.querySelector('[data-role="run"]') as HTMLButtonElement;
  const backBtn = details.querySelector('[data-role="back"]') as HTMLButtonElement;

  const setError = (message: string | null) => {
    if (!message) {
      errorBox.hidden = true;
      errorBox.textContent = '';
      return;
    }
    errorBox.textContent = message;
    errorBox.hidden = false;
  };

  const selectOption = (option: OptionDescriptor | undefined) => {
    store.setInstaller(option);
    setError(null);
    optionButtons.forEach((button) => button.classList.remove('is-active'));
    if (option) {
      const btn = optionButtons.get(option.id);
      if (btn) btn.classList.add('is-active');
    }
    if (option && option.requiresPath) {
      pathField.hidden = false;
      pathLabel.textContent = option.pathLabel ?? 'Installation path';
      runBtn.disabled = pathInput.value.trim() === '';
    } else {
      pathField.hidden = true;
      pathInput.value = '';
      store.setPath(undefined);
      runBtn.disabled = !option;
    }
  };

  optionButtons.forEach((button, id) => {
    button.addEventListener('click', () => {
      const option = store.getState().options.find((item) => item.id === id);
      selectOption(option);
    });
  });

  pathInput.addEventListener('input', () => {
    store.setPath(pathInput.value.trim() || undefined);
    runBtn.disabled = pathInput.value.trim() === '';
  });

  browseBtn.addEventListener('click', async () => {
    const selected = store.getState().selectedInstaller;
    const path = await browseForDirectory(selected?.pathLabel ?? 'Select folder');
    if (path) {
      pathInput.value = path;
      store.setPath(path);
      runBtn.disabled = false;
    }
  });

  backBtn.addEventListener('click', () => {
    store.setStep(Step.Modpack);
  });

  runBtn.addEventListener('click', async () => {
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
    }

    setError(null);
    store.clearLogs();
    store.setBusy(true);

    try {
      const payload = option.requiresPath
        ? { path: store.getState().selectedPath }
        : {};
      const result = await runInstaller(option.id, payload);
      store.appendLogs(result.messages);
      store.setResult(result);
      store.setStep(Step.Status);
    } catch (error) {
      console.error('Failed to run installer', error);
      setError('Something went wrong while running the installer. See console for details.');
      store.setResult(null);
    } finally {
      store.setBusy(false);
    }
  });

  if (current) {
    selectOption(current);
    if (current.requiresPath && store.getState().selectedPath) {
      pathInput.value = store.getState().selectedPath ?? '';
      runBtn.disabled = false;
    }
  }

  container.innerHTML = `
    <header class="screen__header">
      <h2 class="screen__title">Choose Installer</h2>
      <p class="screen__subtitle">Pick the launcher you want to configure.</p>
    </header>
  `;
  container.appendChild(optionsHost);
  container.appendChild(details);

  return container;
}
