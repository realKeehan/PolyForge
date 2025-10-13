import type { Store } from '../../app/state';
import { Step, type Mode } from '../../app/types';
import renderModeOption from '../../templates/modeOptionTemplate';

const HERO_GLYPH = `
  <svg width="52" height="52" viewBox="0 0 52 52" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <circle cx="26" cy="26" r="24" stroke="#C8A4FF" stroke-opacity="0.25" stroke-width="2.5"/>
    <path d="M26.0002 28.9997C28.7616 28.9997 31.0002 26.7612 31.0002 23.9997C31.0002 21.2383 28.7616 18.9997 26.0002 18.9997C23.2387 18.9997 21.0002 21.2383 21.0002 23.9997C21.0002 26.7612 23.2387 28.9997 26.0002 28.9997Z" stroke="#EBD7FF" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
    <path d="M38.9995 40.9998C38.9995 35.4768 33.5224 30.9998 27.9995 30.9998H23.9995C18.4765 30.9998 12.9995 35.4768 12.9995 40.9998" stroke="#EBD7FF" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
  </svg>
`;

const OPTION_ICONS: Record<Mode, string> = {
  install: `
    <svg width="50" height="50" viewBox="0 0 50 50" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M15 37.76H34.32" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
      <path d="M24.6599 12V31.32M24.6599 31.32L30.2949 25.685M24.6599 31.32L19.0249 25.685" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
    </svg>
  `,
  update: `
    <svg width="50" height="50" viewBox="0 0 50 50" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M36.053 20.2347C34.1922 15.9761 29.9429 13 24.9983 13C18.7459 13 13.605 17.7589 13 23.8521" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
      <path d="M31.0273 20.2347H36.3328C36.7324 20.2347 37.0563 19.9108 37.0563 19.5113V14.2058" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
      <path d="M14.0034 29.881C15.8641 34.1396 20.1135 37.1158 25.058 37.1158C31.3104 37.1158 36.4514 32.3569 37.0563 26.2637" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
      <path d="M19.0289 29.881H13.7235C13.3239 29.881 13 30.2049 13 30.6045V35.91" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
    </svg>
  `,
  uninstall: `
    <svg width="50" height="50" viewBox="0 0 50 50" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M34.3833 23.6917V36.2544C34.3833 36.6973 34.0243 37.0563 33.5814 37.0563H16.4747C16.0319 37.0563 15.6729 36.6973 15.6729 36.2544V23.6917" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
      <path d="M22.3555 31.7104V23.6917" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
      <path d="M27.7012 31.7104V23.6917" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
      <path d="M37.0563 18.3458H30.374M30.374 18.3458V13.8019C30.374 13.359 30.015 13 29.5721 13H20.4842C20.0413 13 19.6823 13.359 19.6823 13.8019V18.3458M30.374 18.3458H19.6823M13 18.3458H19.6823" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
    </svg>
  `,
  repair: `
    <svg width="50" height="50" viewBox="0 0 50 50" fill="none" xmlns="http://www.w3.org/2000/svg">
      <mask id="mask0" style="mask-type:luminance" maskUnits="userSpaceOnUse" x="11" y="11" width="28" height="28">
        <path d="M38.06 11H11V38.06H38.06V11Z" fill="white"/>
      </mask>
      <g mask="url(#mask0)">
        <path d="M22.3313 22.9589L14.3587 30.9316C13.4781 31.8122 13.4781 33.2399 14.3587 34.1206C15.2393 35.0012 16.6671 35.0012 17.5478 34.1206L25.5204 26.148" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
        <path d="M22.3308 22.9589C21.3794 20.5314 21.5654 17.3463 23.5268 15.385C25.488 13.4236 29.1075 12.9932 31.1007 14.1891L27.6723 17.6174L27.3537 21.1251L30.8614 20.8065L34.2898 17.3781C35.4857 19.3713 35.0552 22.9909 33.0938 24.9521C31.1325 26.9134 27.9473 27.0995 25.5198 26.148" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
      </g>
    </svg>
  `,
};

type ModeCard = {
  id: Mode;
  label: string;
  description: string;
  variant: string;
};

const MODE_CARDS: ModeCard[] = [
  { id: 'install', label: 'Install Modpack', description: 'Clean install of the latest Turtel SMP build.', variant: 'install' },
  { id: 'update', label: 'Update Modpack', description: 'Fetch the newest release and patch your instance.', variant: 'update' },
  { id: 'uninstall', label: 'Uninstall Modpack', description: 'Remove PolyForge files from the selected launcher.', variant: 'uninstall' },
  { id: 'repair', label: 'Repair Installation', description: 'Reapply the modpack files if something looks off.', variant: 'repair' },
];

export function renderMode(store: Store): HTMLElement {
  const container = document.createElement('section');
  container.className = 'screen screen--mode';

  const current = store.getState().selectedMode;

  container.innerHTML = `
    <div class="mode-stage">
      <header class="mode-stage__header">
        <span class="mode-stage__glyph">${HERO_GLYPH}</span>
        <div class="mode-stage__headline">
          <h2 class="mode-stage__title">Choose an Option</h2>
          <p class="mode-stage__subtitle">PolyForge can install, update, repair, or remove your modpack.</p>
        </div>
      </header>
      <p class="mode-stage__hint">Auto-detects compatible launchers on 1.20.1 and keeps your files tidy.</p>
      <div class="mode-stage__options" role="list"></div>
      <footer class="mode-stage__actions">
        <button type="button" class="btn btn--ghost mode-stage__back">Back</button>
        <button type="button" class="btn btn--primary mode-stage__next" ${current ? '' : 'disabled'}>Next</button>
      </footer>
    </div>
  `;

  const optionsHost = container.querySelector('.mode-stage__options') as HTMLDivElement;
  const backButton = container.querySelector('.mode-stage__back') as HTMLButtonElement;
  const nextButton = container.querySelector('.mode-stage__next') as HTMLButtonElement;

  const buttons: HTMLButtonElement[] = [];

  MODE_CARDS.forEach((card) => {
    const markup = renderModeOption({
      id: card.id,
      title: card.label,
      description: card.description,
      variant: card.variant,
      icon: OPTION_ICONS[card.id],
      active: card.id === current,
    });
    const template = document.createElement('template');
    template.innerHTML = markup.trim();
    const button = template.content.firstElementChild as HTMLButtonElement;
    optionsHost.appendChild(button);
    buttons.push(button);
  });

  const activate = (mode: Mode) => {
    buttons.forEach((btn) => {
      btn.classList.toggle('is-active', btn.dataset.mode === mode);
    });
    store.setMode(mode);
    nextButton.disabled = false;
  };

  buttons.forEach((button) => {
    button.addEventListener('click', () => {
      const mode = button.dataset.mode as Mode;
      activate(mode);
    });
  });

  nextButton.addEventListener('click', () => {
    const mode = store.getState().selectedMode;
    if (!mode) {
      activate(MODE_CARDS[0].id);
    }
    store.setStep(Step.Modpack);
  });

  backButton.addEventListener('click', () => {
    store.setStep(Step.License);
  });

  return container;
}
