import type { Store } from '../../app/state';
import { Step } from '../../app/types';

const MODPACKS = [
  { id: 'turtel-smp5', name: 'Turtel SMP Season 5', tagline: 'The official Turtel SMP experience with Quilt and curated mods.' },
];

export function renderModpack(store: Store): HTMLElement {
  const container = document.createElement('section');
  container.className = 'screen screen--modpack';

  const current = store.getState().selectedModpack ?? MODPACKS[0].id;

  const list = document.createElement('div');
  list.className = 'modpack-grid';

  MODPACKS.forEach((pack) => {
    const button = document.createElement('button');
    button.type = 'button';
    button.className = 'card card--modpack';
    button.dataset.id = pack.id;
    button.innerHTML = `
      <span class="card__title">${pack.name}</span>
      <span class="card__description">${pack.tagline}</span>
    `;
    if (pack.id === current) {
      button.classList.add('is-active');
    }
    button.addEventListener('click', () => {
      list.querySelectorAll('.card').forEach((node) => node.classList.remove('is-active'));
      button.classList.add('is-active');
      store.setModpack(pack.id);
    });
    list.appendChild(button);
  });

  const next = document.createElement('button');
  next.type = 'button';
  next.className = 'btn btn--primary';
  next.textContent = 'Next';
  next.addEventListener('click', () => {
    if (!store.getState().selectedModpack) {
      store.setModpack(MODPACKS[0].id);
    }
    store.setStep(Step.Installer);
  });

  const back = document.createElement('button');
  back.type = 'button';
  back.className = 'btn btn--ghost';
  back.textContent = 'Back';
  back.addEventListener('click', () => store.setStep(Step.Mode));

  const actions = document.createElement('footer');
  actions.className = 'screen__actions';
  actions.append(back, next);

  container.innerHTML = `
    <header class="screen__header">
      <h2 class="screen__title">Choose Modpack</h2>
      <p class="screen__subtitle">We&apos;ll apply the selected pack to your launcher.</p>
    </header>
  `;
  container.appendChild(list);
  container.appendChild(actions);

  return container;
}
