import type { Store } from '../../app/state';
import { Step, type Mode } from '../../app/types';
import { createSocialLinks } from '../components/social';

const HERO_ICON = `
  <svg viewBox="0 0 40 40" fill="none" aria-hidden="true">
    <path d="M20 22C22.7614 22 25 19.7614 25 17C25 14.2386 22.7614 12 20 12C17.2386 12 15 14.2386 15 17C15 19.7614 17.2386 22 20 22Z" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"></path>
    <path d="M33 34C33 28.477 27.523 24 22 24H18C12.477 24 7 28.477 7 34" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"></path>
  </svg>
`;

const MODE_ICONS: Record<string, string> = {
  install: `<svg viewBox="0 0 24 24" fill="none" width="22" height="22"><path d="M12 3v12m0 0l-4-4m4 4l4-4M4 17v2a2 2 0 002 2h12a2 2 0 002-2v-2" stroke="#8F00FF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>`,
  update: `<svg viewBox="0 0 24 24" fill="none" width="22" height="22"><path d="M4 4v5h5M20 20v-5h-5" stroke="#8F00FF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M20.49 9A9 9 0 005.64 5.64L4 4m16 16l-1.64-1.64A9 9 0 013.51 15" stroke="#8F00FF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>`,
  uninstall: `<svg viewBox="0 0 24 24" fill="none" width="22" height="22"><path d="M3 6h18M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2m3 0v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6h14z" stroke="#8F00FF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>`,
  repair: `<svg viewBox="0 0 24 24" fill="none" width="22" height="22"><path d="M14.7 6.3a1 1 0 000 1.4l1.6 1.6a1 1 0 001.4 0l3.77-3.77a6 6 0 01-7.94 7.94l-6.91 6.91a2.12 2.12 0 01-3-3l6.91-6.91a6 6 0 017.94-7.94l-3.76 3.76z" stroke="#8F00FF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>`,
};

const MODE_CARDS: Array<{ id: Mode; label: string }> = [
  { id: 'install', label: 'Install Modpack' },
  { id: 'update', label: 'Update Modpack' },
  { id: 'uninstall', label: 'Uninstall Modpack' },
  { id: 'repair', label: 'Repair Modpack' },
];

function radioDot(): string {
  return `<span class="radio-dot"><span class="radio-dot__inner"></span></span>`;
}

export function renderMode(store: Store): HTMLElement {
  const container = document.createElement('section');
  container.className = 'screen screen--mode';

  const header = document.createElement('div');
  header.className = 'stage__header';
  header.innerHTML = `
    <span class="stage__icon">${HERO_ICON}</span>
    <div>
      <h2 class="stage__title">Choose an Option</h2>
    </div>
  `;

  const list = document.createElement('div');
  list.className = 'radio-list';

  const footer = document.createElement('footer');
  footer.className = 'screen-footer';
  const social = createSocialLinks();
  const actions = document.createElement('div');
  actions.className = 'screen-footer__actions';
  const backButton = document.createElement('button');
  backButton.type = 'button';
  backButton.className = 'btn btn--ghost';
  backButton.textContent = 'Back';
  const nextButton = document.createElement('button');
  nextButton.type = 'button';
  nextButton.className = 'btn btn--primary';
  nextButton.textContent = 'Next';
  // Install is selected by default so always enabled
  nextButton.disabled = false;
  actions.append(backButton, nextButton);
  footer.append(social, actions);

  container.append(header, list, footer);

  const buttons: HTMLButtonElement[] = [];

  const activate = (mode: Mode) => {
    buttons.forEach((btn) => {
      const isActive = btn.dataset.mode === mode;
      btn.classList.toggle('is-active', isActive);
    });
    store.setMode(mode);
    nextButton.disabled = false;
  };

  const currentMode = store.getState().selectedMode ?? 'install';

  MODE_CARDS.forEach((card) => {
    const button = document.createElement('button');
    button.type = 'button';
    button.className = 'radio-item radio-item--card radio-item--has-bg';
    button.dataset.mode = card.id;
    button.innerHTML = `
      ${radioDot()}
      <span class="radio-item__icon">${MODE_ICONS[card.id] ?? ''}</span>
      <span class="radio-item__body">
        <span class="radio-item__label">${card.label}</span>
      </span>
    `;
    if (currentMode === card.id) {
      button.classList.add('is-active');
    }
    button.addEventListener('click', () => {
      activate(card.id);
    });
    buttons.push(button);
    list.appendChild(button);
  });

  backButton.addEventListener('click', () => {
    store.setStep(Step.License);
  });

  nextButton.addEventListener('click', () => {
    const selected = store.getState().selectedMode ?? 'install';
    activate(selected);
    store.setStep(Step.Modpack);
  });

  return container;
}
