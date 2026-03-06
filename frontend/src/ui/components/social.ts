import { APP_VERSION } from '../../app/constants';

function createAboutIcon(): string {
  return `
    <svg viewBox="0 0 40 40" fill="none" aria-hidden="true">
      <circle cx="20" cy="20" r="17" stroke="currentColor" stroke-width="3"></circle>
      <path d="M20 28V18" stroke="currentColor" stroke-width="3" stroke-linecap="round"></path>
      <circle cx="20" cy="13" r="2.5" fill="currentColor"></circle>
    </svg>
  `;
}

function createGlobeIcon(idPrefix: string): string {
  return `
    <svg viewBox="0 0 40 40" fill="none" aria-hidden="true">
      <g clip-path="url(#${idPrefix}-globe)">
        <path d="M11 19.75C11 24.5824 14.9175 28.5 19.75 28.5C24.5824 28.5 28.5 24.5824 28.5 19.75C28.5 14.9175 24.5824 11 19.75 11C14.9175 11 11 14.9175 11 19.75Z" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"></path>
        <path d="M20.625 11.043C20.625 11.043 23.25 14.4998 23.25 19.7498C23.25 24.9998 20.625 28.4567 20.625 28.4567" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"></path>
        <path d="M18.875 28.4567C18.875 28.4567 16.25 24.9998 16.25 19.7498C16.25 14.4998 18.875 11.043 18.875 11.043" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"></path>
        <path d="M11.5508 22.8126H27.9489" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"></path>
        <path d="M11.5508 16.6874H27.9489" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"></path>
      </g>
      <defs>
        <clipPath id="${idPrefix}-globe">
          <rect width="40" height="40" fill="white"></rect>
        </clipPath>
      </defs>
    </svg>
  `;
}

function createGithubIcon(idPrefix: string): string {
  return `
    <svg viewBox="0 0 40 40" fill="none" aria-hidden="true">
      <g clip-path="url(#${idPrefix}-github)">
        <path d="M23.5642 28.5V25.9921C23.5979 25.5755 23.5401 25.1566 23.3946 24.7633C23.2491 24.3701 23.0194 24.0115 22.7206 23.7114C25.5386 23.4056 28.5001 22.3657 28.5001 17.5946C28.4998 16.3746 28.0179 15.2014 27.154 14.3177C27.563 13.2504 27.5341 12.0706 27.0732 11.0234C27.0732 11.0234 26.0142 10.7176 23.5642 12.3167C21.5073 11.7739 19.3391 11.7739 17.2822 12.3167C14.8322 10.7176 13.7732 11.0234 13.7732 11.0234C13.3122 12.0706 13.2833 13.2504 13.6924 14.3177C12.822 15.2079 12.3395 16.3918 12.3463 17.6208C12.3463 22.357 15.3078 23.3968 18.1257 23.7376C17.8305 24.0347 17.6028 24.389 17.4574 24.7774C17.3121 25.1658 17.2524 25.5798 17.2822 25.9921V28.5" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"></path>
        <path d="M17.2821 26.7523C14.5897 27.6027 12.3462 26.7523 11 24.1308" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"></path>
      </g>
      <defs>
        <clipPath id="${idPrefix}-github">
          <rect width="40" height="40" fill="white"></rect>
        </clipPath>
      </defs>
    </svg>
  `;
}

function createHeartIcon(idPrefix: string): string {
  return `
    <svg viewBox="0 0 40 40" fill="none" aria-hidden="true">
      <g clip-path="url(#${idPrefix}-heart)">
        <path d="M28.5 16.6994C28.5 18.2029 27.9804 19.647 27.0526 20.7153C24.9168 23.1751 22.8452 25.7402 20.6296 28.1108C20.1218 28.6463 19.3162 28.6268 18.8302 28.0671L12.447 20.7153C10.5177 18.4931 10.5177 14.9056 12.447 12.6835C14.3954 10.4395 17.5694 10.4395 19.5178 12.6835L19.7498 12.9507L19.9817 12.6836C20.9158 11.6072 22.1881 11 23.5171 11C24.8462 11 26.1183 11.6071 27.0526 12.6835C27.9805 13.7518 28.5 15.1959 28.5 16.6994Z" stroke="currentColor" stroke-width="3" stroke-linejoin="round"></path>
      </g>
      <defs>
        <clipPath id="${idPrefix}-heart">
          <rect width="40" height="40" fill="white"></rect>
        </clipPath>
      </defs>
    </svg>
  `;
}

function showAboutDialog() {
  // Remove existing dialog if any
  const existing = document.querySelector('.about-dialog-overlay');
  if (existing) {
    existing.remove();
    return;
  }

  const overlay = document.createElement('div');
  overlay.className = 'about-dialog-overlay';
  overlay.innerHTML = `
    <div class="about-dialog">
      <h3 class="about-dialog__title">About PolyForge</h3>
      <p class="about-dialog__version">Version ${APP_VERSION} <button type="button" class="about-dialog__cake" aria-label="Cake easter egg" title="Cake?"><svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16"><path d="m3.494.013-.595.79A.747.747 0 0 0 3 1.814v2.683q-.224.051-.432.107c-.702.187-1.305.418-1.745.696C.408 5.56 0 5.954 0 6.5v7c0 .546.408.94.823 1.201.44.278 1.043.51 1.745.696C3.978 15.773 5.898 16 8 16s4.022-.227 5.432-.603c.701-.187 1.305-.418 1.745-.696.415-.261.823-.655.823-1.201v-7c0-.546-.408-.94-.823-1.201-.44-.278-1.043-.51-1.745-.696A12 12 0 0 0 13 4.496v-2.69a.747.747 0 0 0 .092-1.004l-.598-.79-.595.792A.747.747 0 0 0 12 1.813V4.3a22 22 0 0 0-2-.23V1.806a.747.747 0 0 0 .092-1.004l-.598-.79-.595.792A.747.747 0 0 0 9 1.813v2.204a29 29 0 0 0-2 0V1.806A.747.747 0 0 0 7.092.802l-.598-.79-.595.792A.747.747 0 0 0 6 1.813V4.07c-.71.05-1.383.129-2 .23V1.806A.747.747 0 0 0 4.092.802zm-.668 5.556L3 5.524v.967q.468.111 1 .201V5.315a21 21 0 0 1 2-.242v1.855q.488.036 1 .054V5.018a28 28 0 0 1 2 0v1.964q.512-.018 1-.054V5.073c.72.054 1.393.137 2 .242v1.377q.532-.09 1-.201v-.967l.175.045c.655.175 1.15.374 1.469.575.344.217.356.35.356.356s-.012.139-.356.356c-.319.2-.814.4-1.47.575C11.87 7.78 10.041 8 8 8c-2.04 0-3.87-.221-5.174-.569-.656-.175-1.151-.374-1.47-.575C1.012 6.639 1 6.506 1 6.5s.012-.139.356-.356c.319-.2.814-.4 1.47-.575M15 7.806v1.027l-.68.907a.94.94 0 0 1-1.17.276 1.94 1.94 0 0 0-2.236.363l-.348.348a1 1 0 0 1-1.307.092l-.06-.044a2 2 0 0 0-2.399 0l-.06.044a1 1 0 0 1-1.306-.092l-.35-.35a1.935 1.935 0 0 0-2.233-.362.935.935 0 0 1-1.168-.277L1 8.82V7.806c.42.232.956.428 1.568.591C3.978 8.773 5.898 9 8 9s4.022-.227 5.432-.603c.612-.163 1.149-.36 1.568-.591m0 2.679V13.5c0 .006-.012.139-.356.355-.319.202-.814.401-1.47.576C11.87 14.78 10.041 15 8 15c-2.04 0-3.87-.221-5.174-.569-.656-.175-1.151-.374-1.47-.575-.344-.217-.356-.35-.356-.356v-3.02a1.935 1.935 0 0 0 2.298.43.935.935 0 0 1 1.08.175l.348.349a2 2 0 0 0 2.615.185l.059-.044a1 1 0 0 1 1.2 0l.06.044a2 2 0 0 0 2.613-.185l.348-.348a.94.94 0 0 1 1.082-.175c.781.39 1.718.208 2.297-.426"/></svg></button></p>
      <p class="about-dialog__desc">
        PolyForge is a modpack installer and launcher manager for Minecraft.
        It simplifies the process of installing, updating, and managing modpacks
        across multiple launcher platforms.
      </p>
      <p class="about-dialog__desc">
        Built with care for the community. Select your preferred launcher,
        choose a modpack, and PolyForge handles the rest.
      </p>
      <div class="about-dialog__actions">
        <button type="button" class="btn btn--primary about-dialog__close">Close</button>
      </div>
    </div>
  `;

  overlay.addEventListener('click', (event) => {
    if (event.target === overlay) {
      overlay.remove();
    }
  });

  const closeBtn = overlay.querySelector('.about-dialog__close') as HTMLButtonElement;
  closeBtn.addEventListener('click', () => {
    overlay.remove();
  });

  const cakeBtn = overlay.querySelector('.about-dialog__cake') as HTMLButtonElement;
  cakeBtn.addEventListener('click', () => {
    overlay.remove();
    const shell = document.querySelector('.app-window') as HTMLElement;
    if (shell) {
      // Trigger the easter egg video (same as Konami code)
      const EASTER_EGG_VIDEO = 'https://keehan.co/KUMI_Files/NiceComputer.mp4';
      const eeOverlay = document.createElement('div');
      eeOverlay.className = 'easter-egg-overlay';
      eeOverlay.innerHTML = `
        <video class="easter-egg-video" autoplay>
          <source src="${EASTER_EGG_VIDEO}" type="video/mp4" />
        </video>
      `;
      eeOverlay.addEventListener('click', (e) => {
        if (e.target === eeOverlay) {
          const video = eeOverlay.querySelector('video');
          if (video) { video.pause(); video.src = ''; }
          eeOverlay.remove();
        }
      });
      const video = eeOverlay.querySelector('video') as HTMLVideoElement;
      video.addEventListener('ended', () => { eeOverlay.remove(); });
      shell.appendChild(eeOverlay);
    }
  });

  document.querySelector('.app-window')?.appendChild(overlay);
}

export function createSocialLinks(): HTMLElement {
  const unique = `pf-${Math.random().toString(36).slice(2, 8)}`;
  const container = document.createElement('div');
  container.className = 'social-links';

  container.innerHTML = `
    <button type="button" class="social-links__button social-links__button--primary" data-action="about" aria-label="About">
      ${createAboutIcon()}
    </button>
    <a class="social-links__button" href="https://polyforge.dev" target="_blank" rel="noopener noreferrer" aria-label="Website">
      ${createGlobeIcon(unique)}
    </a>
    <a class="social-links__button" href="https://github.com/realKeehan/PolyForge" target="_blank" rel="noopener noreferrer" aria-label="GitHub">
      ${createGithubIcon(unique)}
    </a>
    <a class="social-links__button" href="https://keehan.co/donate" target="_blank" rel="noopener noreferrer" aria-label="Donate">
      ${createHeartIcon(unique)}
    </a>
  `;

  const aboutButton = container.querySelector('[data-action="about"]') as HTMLButtonElement;
  aboutButton.addEventListener('click', () => {
    showAboutDialog();
  });

  return container;
}
