import './styles/main.scss';
import pug from 'pug';
import optionTemplateSource from './templates/option.pug?raw';
import type { ActionResult, ExecutionPayload, LogEntry, OptionDescriptor } from './types';

const renderOption = pug.compile(optionTemplateSource, { filename: 'option.pug' });

interface PathBinding {
  option: OptionDescriptor;
  input?: HTMLInputElement;
}

const appRoot = document.getElementById('app');
if (!appRoot) {
  throw new Error('App root not found');
}

const layout = document.createElement('main');
layout.className = 'layout';
layout.innerHTML = `
  <section class="menu">
    <header class="menu__header">
      <h1 class="menu__title">PolyForge</h1>
      <p class="menu__subtitle">Keehan's Universal Modpack Installer</p>
    </header>
    <div class="menu__options" data-role="options"></div>
  </section>
  <section class="details">
    <header class="details__header">
      <h2 class="details__title">Activity Log</h2>
      <div class="log-controls">
        <button class="log-controls__clear" type="button" data-role="clear-log">Clear log</button>
      </div>
    </header>
    <div class="log-container">
      <ul class="log-list" data-role="log"></ul>
    </div>
  </section>
`;

appRoot.appendChild(layout);

const optionsHost = layout.querySelector('[data-role="options"]') as HTMLDivElement;
const logList = layout.querySelector('[data-role="log"]') as HTMLUListElement;
const clearButton = layout.querySelector('[data-role="clear-log"]') as HTMLButtonElement;

const pathBindings = new Map<string, PathBinding>();

clearButton.addEventListener('click', () => {
  logList.innerHTML = '';
});

function appendLog(entry: LogEntry) {
  const item = document.createElement('li');
  item.className = `log-entry ${entry.level}`;
  item.innerHTML = `
    <span class="log-entry__level">${entry.level}</span>
    <span class="log-entry__message">${entry.message}</span>
  `;
  logList.appendChild(item);
  item.scrollIntoView({ behavior: 'smooth', block: 'end' });
}

function appendDivider(message: string) {
  appendLog({ level: 'info', message });
}

function createOptionElement(option: OptionDescriptor) {
  const html = renderOption(option);
  const template = document.createElement('template');
  template.innerHTML = html.trim();
  const button = template.content.firstElementChild as HTMLButtonElement;
  const card = document.createElement('div');
  card.className = 'option-card';
  card.appendChild(button);

  let binding: PathBinding = { option };

  if (option.requiresPath) {
    const wrapper = document.createElement('div');
    wrapper.className = 'option-card__path';
    const label = document.createElement('label');
    label.textContent = option.pathLabel ?? 'Path';
    const input = document.createElement('input');
    input.type = 'text';
    input.placeholder = 'Select a directory…';
    label.appendChild(input);
    const chooser = document.createElement('button');
    chooser.type = 'button';
    chooser.textContent = 'Browse…';
    chooser.addEventListener('click', async () => {
      try {
        const selected = await window.go.app.App.SelectDirectory(option.pathLabel ?? 'Select directory');
        if (selected) {
          input.value = selected;
        }
      } catch (err) {
        console.error(err);
      }
    });
    wrapper.appendChild(label);
    wrapper.appendChild(chooser);
    card.appendChild(wrapper);
    binding = { option, input };
  }

  pathBindings.set(option.id, binding);

  button.addEventListener('click', async () => {
    await executeOption(option);
  });

  return card;
}

async function executeOption(option: OptionDescriptor) {
  appendDivider(`▶ Running ${option.title}`);

  const binding = pathBindings.get(option.id);
  const payload: ExecutionPayload = {};
  if (binding?.input) {
    payload.path = binding.input.value.trim();
  }

  try {
    const result = await window.go.app.App.Execute(option.id, payload);
    handleResult(result);
  } catch (err) {
    appendLog({ level: 'error', message: err instanceof Error ? err.message : String(err) });
  }
}

function handleResult(result: ActionResult) {
  if (!result) {
    appendLog({ level: 'error', message: 'No response from backend' });
    return;
  }
  result.messages.forEach((msg) => appendLog(msg));
  if (result.success) {
    appendLog({ level: 'info', message: '✔ Operation completed successfully.' });
  } else {
    appendLog({ level: 'warning', message: '⚠ Operation reported issues.' });
  }
}

async function bootstrap() {
  try {
    const options = await window.go.app.App.GetMenuOptions();
    options.forEach((option) => {
      const element = createOptionElement(option);
      optionsHost.appendChild(element);
    });
    appendLog({ level: 'info', message: 'Ready. Select an action to begin.' });
  } catch (err) {
    appendLog({ level: 'error', message: err instanceof Error ? err.message : String(err) });
  }
}

bootstrap();
