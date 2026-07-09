import type { Store } from '../../app/state';
import { LOCAL_PACK_ID, Step, type RemotePack } from '../../app/types';
import { inspectPolyPack, selectPackFile, verifyPackAccess } from '../../app/ipc';
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

const LOCK_ICON = `<svg viewBox="0 0 16 16" fill="none" aria-hidden="true"><path d="M4.5 7V5a3.5 3.5 0 117 0v2m-8.5 0h10a1 1 0 011 1v5a1 1 0 01-1 1H3a1 1 0 01-1-1V8a1 1 0 011-1z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>`;

type ModpackDef = RemotePack;

// Fallback list used only when the remote content manifest is unavailable
// (offline first run). The live list comes from polyforge.dev/api/manifest.json
// so packs can be added or changed without shipping a new binary.
// Passwords are verified server-side via /api/pack-access; the local
// passwordHash here is only an offline fallback for the built-in pack.
const DEFAULT_MODPACKS: ModpackDef[] = [
  { id: 'turtel-smp', name: 'Turtel SMP' },
  {
    id: 'event-pack',
    name: 'Event Pack',
    requiresPassword: true,
    // SHA-256 of the pack password (offline fallback only)
    passwordHash: '908baa40ef565d0d30fab71f76b9e73d4cf88101984c4f57c6c674804dc4006a',
  },
];

let cleanInstallPreference = false;

function radioDot(): string {
  return `<span class="radio-dot"><span class="radio-dot__inner"></span></span>`;
}

/** Hash a string with SHA-256 and return hex */
async function sha256(input: string): Promise<string> {
  const data = new TextEncoder().encode(input);
  const hash = await crypto.subtle.digest('SHA-256', data);
  return Array.from(new Uint8Array(hash))
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('');
}

export function renderModpack(store: Store): HTMLElement {
  const container = document.createElement('section');
  container.className = 'screen screen--modpack';

  const header = document.createElement('div');
  header.className = 'stage__header';
  header.innerHTML = `
    <span class="stage__icon">${MODPACK_ICON}</span>
    <div>
      <h2 class="stage__title">Choose Modpack</h2>
    </div>
  `;

  const list = document.createElement('div');
  list.className = 'radio-list';

  const toggle = document.createElement('label');
  toggle.className = 'toggle';
  toggle.innerHTML = `
    <input type="checkbox" class="toggle__input" hidden aria-hidden="true" ${cleanInstallPreference ? 'checked' : ''} />
    <span class="toggle__control" aria-hidden="true"></span>
    <span class="toggle__label">Clean install</span>
  `;

  // Password field (shown only when a locked pack is selected)
  const passwordField = document.createElement('div');
  passwordField.className = 'password-field';
  passwordField.hidden = true;
  passwordField.innerHTML = `
    <label class="password-field__label">
      ${LOCK_ICON}
      <span>Password required</span>
    </label>
    <div class="password-field__row">
      <input type="password" class="password-field__input" placeholder="Enter pack password" autocomplete="off" />
    </div>
    <span class="password-field__error" hidden></span>
  `;

  const passwordInput = passwordField.querySelector('.password-field__input') as HTMLInputElement;
  const passwordError = passwordField.querySelector('.password-field__error') as HTMLSpanElement;

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

  container.append(header, list, toggle, passwordField, footer);

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

  // Prefer the remotely managed pack list; fall back to built-ins offline.
  const remotePacks = store.getState().modpacks;
  const packs: ModpackDef[] = remotePacks?.length ? remotePacks : DEFAULT_MODPACKS;

  const buttons: HTMLButtonElement[] = [];
  const selected = store.getState().selectedModpack ?? packs[0]?.id;
  let currentPack: ModpackDef | undefined;
  let passwordUnlocked = false;

  const showPasswordField = (pack: ModpackDef) => {
    if (pack.requiresPassword && !passwordUnlocked) {
      passwordField.hidden = false;
      passwordInput.value = '';
      passwordError.hidden = true;
      passwordInput.focus();
    } else {
      passwordField.hidden = true;
    }
  };

  const activate = (modpackId: string) => {
    // Re-clicking the current pack should not wipe a typed password
    if (currentPack?.id === modpackId && store.getState().selectedModpack === modpackId) return;

    buttons.forEach((button) => {
      button.classList.toggle('is-active', button.dataset.modpack === modpackId);
    });
    currentPack = packs.find((p) => p.id === modpackId);
    store.setModpack(modpackId);

    // Reset password state when switching packs
    passwordUnlocked = false;
    if (currentPack) {
      showPasswordField(currentPack);
    } else {
      passwordField.hidden = true;
    }
  };

  packs.forEach((pack) => {
    const button = document.createElement('button');
    button.type = 'button';
    button.className = 'radio-item radio-item--card radio-item--has-bg';
    button.dataset.modpack = pack.id;
    button.innerHTML = `
      ${radioDot()}
      <span class="radio-item__body">
        <span class="radio-item__label">${pack.name}${pack.requiresPassword ? `<span class="radio-item__lock" title="Password protected">${LOCK_ICON}</span>` : ''}</span>
      </span>
    `;
    if (pack.id === selected) {
      button.classList.add('is-active');
      currentPack = pack;
    }
    button.addEventListener('click', () => activate(pack.id));
    buttons.push(button);
    list.appendChild(button);
  });

  // ── Local pack (manual profile mode) ──────────
  // Lets the user load a .polypack from disk instead of a hosted pack.
  const localPack = store.getState().localPack;
  const localBtn = document.createElement('button');
  localBtn.type = 'button';
  localBtn.className = 'radio-item radio-item--card radio-item--has-bg';
  localBtn.dataset.modpack = LOCAL_PACK_ID;
  const localLabel = localPack
    ? `${localPack.name} <span style="color:var(--text-muted, #888);font-size:.78em">v${localPack.version} · ${localPack.modCount} mods · local file</span>`
    : `Load local pack&hellip; <span style="color:var(--text-muted, #888);font-size:.78em">(.polypack)</span>`;
  localBtn.innerHTML = `
    ${radioDot()}
    <span class="radio-item__body">
      <span class="radio-item__label">${localLabel}</span>
    </span>
  `;
  if (selected === LOCAL_PACK_ID && localPack) {
    localBtn.classList.add('is-active');
  }
  localBtn.addEventListener('click', async () => {
    // Already loaded → just select it; otherwise open the file picker.
    if (store.getState().localPack && store.getState().selectedModpack === LOCAL_PACK_ID) return;
    if (!store.getState().localPack) {
      const path = await selectPackFile();
      if (!path) return;
      try {
        const info = await inspectPolyPack(path);
        store.setModpack(LOCAL_PACK_ID);
        store.setLocalPack(info); // triggers re-render showing the pack
        return;
      } catch (error) {
        console.error('Failed to read pack', error);
        passwordError.textContent = 'That file is not a valid PolyForge pack.';
        passwordError.hidden = false;
        passwordField.hidden = false;
        return;
      }
    }
    activate(LOCAL_PACK_ID);
  });
  buttons.push(localBtn);
  list.appendChild(localBtn);

  // Show password field if the initially selected pack requires it
  if (currentPack?.requiresPassword) {
    showPasswordField(currentPack);
  }

  if (!store.getState().selectedModpack && packs[0]) {
    activate(packs[0].id);
  }

  backButton.addEventListener('click', () => {
    store.setStep(Step.Mode);
  });

  const showPasswordError = (message: string, clearInput: boolean) => {
    passwordError.textContent = message;
    passwordError.hidden = false;
    if (clearInput) passwordInput.value = '';
    passwordInput.focus();
  };

  nextButton.addEventListener('click', async () => {
    if (!store.getState().selectedModpack && packs[0]) {
      activate(packs[0].id);
    }

    // Validate password if needed
    if (currentPack?.requiresPassword && !passwordUnlocked) {
      const value = passwordInput.value.trim();
      if (!value) {
        showPasswordError('Please enter the pack password.', false);
        return;
      }

      nextButton.disabled = true;
      nextButton.textContent = 'Checking...';
      let granted = false;
      let message = 'Incorrect password.';
      try {
        // Server-side check — the hash never ships with the app
        const result = await verifyPackAccess(currentPack.id, value);
        if (result.granted) {
          granted = true;
          // Capture the download URL revealed on success; the password is not
          // retained, so this is the only chance to grab it for the installer.
          store.setPackDownload(result.url, currentPack.name);
        } else if (result.offline && currentPack.passwordHash) {
          // Server unreachable: fall back to the locally cached hash
          granted = (await sha256(value)) === currentPack.passwordHash;
        } else if (result.offline) {
          message = 'Could not reach the verification server. Check your connection.';
        } else if (result.error) {
          message = result.error;
        }
      } catch (error) {
        console.error('Pack verification failed', error);
        if (currentPack.passwordHash) {
          granted = (await sha256(value)) === currentPack.passwordHash;
        } else {
          message = 'Password verification failed. Please try again.';
        }
      } finally {
        nextButton.disabled = false;
        nextButton.textContent = 'Next';
      }

      if (!granted) {
        showPasswordError(message, true);
        return;
      }

      passwordUnlocked = true;
      passwordField.hidden = true;
    }

    store.setStep(Step.Installer);
  });

  return container;
}
