import type { Store } from '../../app/state';
import { Step } from '../../app/types';

export function renderLicense(store: Store): HTMLElement {
  const container = document.createElement('section');
  container.className = 'screen screen--license';

  container.innerHTML = `
    <header class="screen__header">
      <h2 class="screen__title">License Agreement</h2>
    </header>
    <article class="license">
      <p>Before the installer can continue we just need to confirm you&apos;re ready for take-off. This wizard will configure your chosen launcher for the Turtel SMP 5 modpack.</p>
      <p>By continuing you acknowledge that modded Minecraft launches may adjust game settings, install additional dependencies such as Quilt Loader, and download required resources.</p>
      <label class="license__accept">
        <input type="checkbox" data-role="accept" />
        <span>I understand and accept the above.</span>
      </label>
    </article>
    <footer class="screen__actions">
      <button type="button" class="btn btn--primary" data-role="next" disabled>Next</button>
    </footer>
  `;

  const accept = container.querySelector('[data-role="accept"]') as HTMLInputElement;
  const next = container.querySelector('[data-role="next"]') as HTMLButtonElement;

  accept.addEventListener('change', () => {
    next.disabled = !accept.checked;
  });

  next.addEventListener('click', () => {
    if (!accept.checked) return;
    store.setStep(Step.Mode);
  });

  return container;
}
