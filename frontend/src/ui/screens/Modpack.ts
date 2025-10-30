import type { Store } from '../../app/state';
import { Step } from '../../app/types';
import { createSocialLinks } from '../components/social';

const MODPACK_ICON = `
  <svg viewBox="0 0 40 40" fill="none" aria-hidden="true">
    <path d="M21.1345 11.0384L25.1729 7L32.24 14.0672L28.2016 18.1056M21.1345 11.0384L7.41819 24.7547C7.15043 25.0224 7 25.3856 7 25.7642V32.24H13.4758C13.8545 32.24 14.2176 32.0897 14.4854 31.8218L28.2016 18.1056M21.1345 11.0384L28.2016 18.1056" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"></path>
  </svg>
`;

const CHECK_ICON = `
  <svg viewBox="0 0 18 14" fill="none" aria-hidden="true">
    <path d="M2 8.084L6.056 12.14L16.196 2" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
  </svg>
`;

const MODPACKS = [
  { id: 'turtel-smp5', name: 'Turtel SMP Season 5', tagline: 'Official experience with Quilt, QoL tweaks, and parity improvements.' },
];

let cleanInstallPreference = true;

export function renderModpack(store: Store): HTMLElement {
  const container = document.createElement('section');
  container.className = 'screen screen--modpack';

  const header = document.createElement('div');
  header.className = 'stage__header';
  header.innerHTML = `
    <span class="stage__icon">${MODPACK_ICON}</span>
    <div>
      <h2 class="stage__title">Choose Modpack</h2>
      <p class="stage__subtitle">Pick the pack you want to manage today.</p>
    </div>
  `;

  const list = document.createElement('div');
  list.className = 'select-list';

  const toggle = document.createElement('label');
  toggle.className = 'toggle';
  toggle.innerHTML = `
    <input type="checkbox" class="toggle__input" hidden aria-hidden="true" ${cleanInstallPreference ? 'checked' : ''} />
    <span class="toggle__control" aria-hidden="true"></span>
    <span class="toggle__label">Clean install</span>
  `;

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

  actions.append(backButton, nextButton);
  footer.append(social, actions);

  container.append(header, list, toggle, footer);

  const toggleInput = toggle.querySelector('.toggle__input') as HTMLInputElement;
  const toggleControl = toggle.querySelector('.toggle__control') as HTMLSpanElement;

  const updateToggle = () => {
    cleanInstallPreference = toggleInput.checked;
    if (toggleInput.checked) {
      toggleControl.classList.add('is-active');
      toggleControl.innerHTML = CHECK_ICON;
    } else {
      toggleControl.classList.remove('is-active');
      toggleControl.innerHTML = '';
    }
  };

  toggle.addEventListener('click', (event) => {
    if (event.target === toggleInput) return;
    event.preventDefault();
    toggleInput.checked = !toggleInput.checked;
    updateToggle();
  });

  toggleInput.addEventListener('change', updateToggle);
  updateToggle();

  const buttons: HTMLButtonElement[] = [];
  const selected = store.getState().selectedModpack ?? MODPACKS[0]?.id;

  const activate = (modpackId: string) => {
    buttons.forEach((button) => {
      button.classList.toggle('is-active', button.dataset.modpack === modpackId);
    });
    store.setModpack(modpackId);
  };

  MODPACKS.forEach((pack) => {
    const button = document.createElement('button');
    button.type = 'button';
    button.className = 'select-card';
    button.dataset.modpack = pack.id;
    button.innerHTML = `
      <span class="select-card__icon">
        <svg viewBox="0 0 40 40" fill="none" aria-hidden="true">
          <path d="M21.1345 11.0384L25.1729 7L32.24 14.0672L28.2016 18.1056M21.1345 11.0384L7.41819 24.7547C7.15043 25.0224 7 25.3856 7 25.7642V32.24H13.4758C13.8545 32.24 14.2176 32.0897 14.4854 31.8218L28.2016 18.1056M21.1345 11.0384L28.2016 18.1056" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"></path>
        </svg>
      </span>
      <span class="select-card__body">
        <span class="select-card__title">${pack.name}</span>
        <span class="select-card__description">${pack.tagline}</span>
      </span>
    `;
    if (pack.id === selected) {
      button.classList.add('is-active');
    }
    button.addEventListener('click', () => activate(pack.id));
    buttons.push(button);
    list.appendChild(button);
  });

  if (!store.getState().selectedModpack && MODPACKS[0]) {
    activate(MODPACKS[0].id);
  }

  backButton.addEventListener('click', () => {
    store.setStep(Step.Mode);
  });

  nextButton.addEventListener('click', () => {
    if (!store.getState().selectedModpack && MODPACKS[0]) {
      activate(MODPACKS[0].id);
    }
    store.setStep(Step.Installer);
  });

  return container;
}
