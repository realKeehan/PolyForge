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

const OPTION_ICONS: Record<Mode, string> = {
  install: `
    <svg viewBox="0 0 50 50" fill="none" aria-hidden="true">
      <path d="M15 37.76H34.32" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
      <path d="M24.6599 12V31.32M24.6599 31.32L30.2949 25.685M24.6599 31.32L19.0249 25.685" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
    </svg>
  `,
  update: `
    <svg viewBox="0 0 50 50" fill="none" aria-hidden="true">
      <path d="M36.053 20.2347C34.1922 15.9761 29.9429 13 24.9983 13C18.7459 13 13.605 17.7589 13 23.8521" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
      <path d="M31.0273 20.2347H36.3328C36.7324 20.2347 37.0563 19.9108 37.0563 19.5113V14.2058" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
      <path d="M14.0034 29.881C15.8641 34.1396 20.1135 37.1158 25.058 37.1158C31.3104 37.1158 36.4514 32.3569 37.0563 26.2637" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
      <path d="M19.0289 29.881H13.7235C13.3239 29.881 13 30.2049 13 30.6045V35.91" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
    </svg>
  `,
  uninstall: `
    <svg viewBox="0 0 50 50" fill="none" aria-hidden="true">
      <path d="M34.3833 23.6917V36.2544C34.3833 36.6973 34.0243 37.0563 33.5814 37.0563H16.4747C16.0319 37.0563 15.6729 36.6973 15.6729 36.2544V23.6917" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
      <path d="M22.3555 31.7104V23.6917" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
      <path d="M27.7012 31.7104V23.6917" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
      <path d="M37.0563 18.3458H30.374M30.374 18.3458V13.8019C30.374 13.359 30.015 13 29.5721 13H20.4842C20.0413 13 19.6823 13.359 19.6823 13.8019V18.3458M30.374 18.3458H19.6823M13 18.3458H19.6823" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
    </svg>
  `,
  repair: `
    <svg viewBox="0 0 50 50" fill="none" aria-hidden="true">
      <path d="M22.3313 22.9589L14.3587 30.9316C13.4781 31.8122 13.4781 33.2399 14.3587 34.1206C15.2393 35.0012 16.6671 35.0012 17.5478 34.1206L25.5204 26.148" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
      <path d="M22.3308 22.9589C21.3794 20.5314 21.5654 17.3463 23.5268 15.385C25.488 13.4236 29.1075 12.9932 31.1007 14.1891L27.6723 17.6174L27.3537 21.1251L30.8614 20.8065L34.2898 17.3781C35.4857 19.3713 35.0552 22.9909 33.0938 24.9521C31.1325 26.9134 27.9473 27.0995 25.5198 26.148" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
    </svg>
  `,
};

const MODE_CARDS: Array<{ id: Mode; label: string; description: string }> = [
  { id: 'install', label: 'Install Modpack', description: 'Clean install of the latest Turtel SMP build.' },
  { id: 'update', label: 'Update Modpack', description: 'Fetch the newest release and patch your instance.' },
  { id: 'uninstall', label: 'Uninstall Modpack', description: 'Remove PolyForge files from the selected launcher.' },
  { id: 'repair', label: 'Repair Installation', description: 'Reapply the modpack files if something looks off.' },
];

export function renderMode(store: Store): HTMLElement {
  const container = document.createElement('section');
  container.className = 'screen screen--mode';

  const header = document.createElement('div');
  header.className = 'stage__header';
  header.innerHTML = `
    <span class="stage__icon">${HERO_ICON}</span>
    <div>
      <h2 class="stage__title">Choose an Option</h2>
      <p class="stage__subtitle">Install, update, repair, or remove your PolyForge modpack.</p>
    </div>
  `;

  const hint = document.createElement('p');
  hint.className = 'stage__hint';
  hint.textContent = 'Auto-detects supported launchers on 1.20.1 and keeps your files tidy.';

  const list = document.createElement('div');
  list.className = 'select-list';

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

  container.append(header, hint, list, footer);

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
    button.className = 'select-card';
    button.dataset.mode = card.id;
    button.innerHTML = `
      <span class="select-card__icon">${OPTION_ICONS[card.id]}</span>
      <span class="select-card__body">
        <span class="select-card__title">${card.label}</span>
        <span class="select-card__description">${card.description}</span>
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
