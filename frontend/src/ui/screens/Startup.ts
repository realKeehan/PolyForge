import type { Store } from '../../app/state';
import { Step } from '../../app/types';
import splashImage from '../../assets/splash.png';

const STARTUP_DISPLAY_MS = 2500;

export function renderStartup(store: Store): HTMLElement {
  const container = document.createElement('section');
  container.className = 'screen screen--startup-splash';

  const splash = document.createElement('div');
  splash.className = 'startup-splash';

  const img = document.createElement('img');
  img.className = 'startup-splash__image';
  img.src = splashImage;
  img.alt = 'PolyForge';
  img.draggable = false;

  splash.appendChild(img);
  container.appendChild(splash);

  // Auto-advance to License after display time
  window.setTimeout(() => {
    if (store.getState().step === Step.Startup) {
      store.setStep(Step.License);
    }
  }, STARTUP_DISPLAY_MS);

  return container;
}
