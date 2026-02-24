import type { Store } from '../../app/state';
import { Step, type Mode } from '../../app/types';
import { createSocialLinks } from '../components/social';

const HERO_ICON = `
  <svg viewBox="0 0 52 52" fill="none" aria-hidden="true">
    <circle cx="26" cy="26" r="24" stroke="#C8A4FF" stroke-opacity="0.25" stroke-width="2.5"></circle>
    <path d="M26.0002 28.9997C28.7616 28.9997 31.0002 26.7612 31.0002 23.9997C31.0002 21.2383 28.7616 18.9997 26.0002 18.9997C23.2387 18.9997 21.0002 21.2383 21.0002 23.9997C21.0002 26.7612 23.2387 28.9997 26.0002 28.9997Z" stroke="#EBD7FF" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
    <path d="M38.9995 40.9998C38.9995 35.4768 33.5224 30.9998 27.9995 30.9998H23.9995C18.4765 30.9998 12.9995 35.4768 12.9995 40.9998" stroke="#EBD7FF" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
  </svg>
`;

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
  nextButton.disabled = !store.getState().selectedMode;
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

  MODE_CARDS.forEach((card) => {
    const button = document.createElement('button');
    button.type = 'button';
    button.className = 'radio-item radio-item--card';
    button.dataset.mode = card.id;
    button.innerHTML = `
      ${radioDot()}
      <span class="radio-item__body">
        <span class="radio-item__label">${card.label}</span>
      </span>
    `;
    if (store.getState().selectedMode === card.id) {
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
    const selected = store.getState().selectedMode ?? MODE_CARDS[0].id;
    activate(selected);
    store.setStep(Step.Modpack);
  });

  return container;
}
