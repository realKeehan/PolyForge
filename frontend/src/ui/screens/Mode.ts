import type { Store } from '../../app/state';
import { Step, type Mode } from '../../app/types';

const MODES: { id: Mode; label: string; description: string }[] = [
  { id: 'install', label: 'Install Modpack', description: 'Clean install of the latest Turtel SMP build.' },
  { id: 'repair', label: 'Repair Installation', description: 'Reapply the modpack files if something looks off.' },
  { id: 'update', label: 'Update Modpack', description: 'Fetch the newest release and patch your instance.' },
  { id: 'uninstall', label: 'Uninstall Modpack', description: 'Remove PolyForge files from the selected launcher.' },
];

export function renderMode(store: Store): HTMLElement {
  const container = document.createElement('section');
  container.className = 'screen screen--mode';

  const current = store.getState().selectedMode;

  const cards = MODES.map((mode) => {
    const button = document.createElement('button');
    button.type = 'button';
    button.className = 'card card--mode';
    button.dataset.mode = mode.id;
    button.innerHTML = `
      <span class="card__title">${mode.label}</span>
      <span class="card__description">${mode.description}</span>
    `;
    if (mode.id === current) {
      button.classList.add('is-active');
    }
    return button;
  });

  const actions = document.createElement('footer');
  actions.className = 'screen__actions';

  const next = document.createElement('button');
  next.type = 'button';
  next.className = 'btn btn--primary';
  next.textContent = 'Next';
  next.disabled = !current;

  actions.appendChild(next);

  const list = document.createElement('div');
  list.className = 'mode-grid';
  cards.forEach((card) => list.appendChild(card));

  container.innerHTML = `
    <header class="screen__header">
      <h2 class="screen__title">Choose an Action</h2>
      <p class="screen__subtitle">What would you like PolyForge to do?</p>
    </header>
  `;
  container.appendChild(list);
  container.appendChild(actions);

  cards.forEach((card) => {
    card.addEventListener('click', () => {
      cards.forEach((btn) => btn.classList.remove('is-active'));
      card.classList.add('is-active');
      const mode = card.dataset.mode as Mode;
      store.setMode(mode);
      next.disabled = false;
    });
  });

  next.addEventListener('click', () => {
    if (!store.getState().selectedMode) return;
    store.setStep(Step.Modpack);
  });

  return container;
}
