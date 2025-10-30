import type { Store } from '../../app/state';
import { Step } from '../../app/types';
import { createSocialLinks } from '../components/social';

const LICENSE_ICON = `
  <svg viewBox="0 0 40 40" fill="none" aria-hidden="true">
    <path d="M11 27.4875V12.1859C11 10.9787 11.9787 10 13.1859 10H27.8318C28.194 10 28.4875 10.2936 28.4875 10.6558V24.9893" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round"></path>
    <path d="M15.3721 10V18.7438L18.1045 16.995L20.8369 18.7438V10" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"></path>
    <path d="M13.186 25.3016H28.4876" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round"></path>
    <path d="M13.186 29.6735H28.4876" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round"></path>
    <path d="M13.1859 29.6735C11.9787 29.6735 11 28.6948 11 27.4875C11 26.2802 11.9787 25.3016 13.1859 25.3016" stroke="#8F00FF" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"></path>
  </svg>
`;

const CHECK_ICON = `
  <svg viewBox="0 0 18 14" fill="none" aria-hidden="true">
    <path d="M2 8.084L6.056 12.14L16.196 2" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"></path>
  </svg>
`;

const LICENSE_COPY = `
End-User License Agreement for PolyForge.

PLEASE READ THIS AGREEMENT CAREFULLY. IT CONTAINS IMPORTANT TERMS THAT AFFECT YOU AND YOUR USE OF THE SOFTWARE. BY INSTALLING, COPYING OR USING THE SOFTWARE, YOU AGREE TO BE BOUND BY THE TERMS OF THIS AGREEMENT. IF YOU DO NOT AGREE TO THESE TERMS, DO NOT INSTALL, COPY, OR USE THE SOFTWARE.

This End-User License Agreement (EULA) is a legal agreement between you--either an individual or a single entity--and the author(s) of this Software for the product identified above, which includes computer software and may include associated media, and online or electronic documentation ("Software").

By installing, copying, or otherwise using the Software, you agree to be bounded by the terms of this EULA. If you do not agree to the terms of this EULA, do not install or use the Software.

1. GRANT OF LICENSE.
This EULA grants you a non-exclusive, non-sublicensable, non-transferable license to install and use the Software. You may install and use an unlimited number of copies of the Software.

2. COPYRIGHT.
All title and copyrights in and to the Software (including but not limited to any images, libraries, and examples incorporated into the Software), the accompanying documentation, and any copies of the Software are owned by the author(s) of this Software.

3. NO WARRANTIES.
The author(s) of this Software expressly disclaims any warranty for the Software. The Software and any related documentation is provided "as is" without warranty of any kind, either express or implied, including, without limitation, the implied warranties or merchantability, fitness for a particular purpose, or noninfringement. The entire risk arising out of use or performance of the Software remains with you.

4. NO LIABILITY FOR DAMAGES.
In no event shall the author(s) of this Software be liable for any special, consequential, incidental or indirect damages whatsoever (including, without limitation, damages for loss of business profits, business interruption, loss of business information, or any other pecuniary loss) arising out of the use of or inability to use this product, even if the author(s) of this Software is aware of the possibility of such damages and known defects.
`.trim();

function copyToClipboard(text: string): Promise<void> {
  if (navigator.clipboard?.writeText) {
    return navigator.clipboard.writeText(text);
  }
  const textarea = document.createElement('textarea');
  textarea.value = text;
  textarea.style.position = 'fixed';
  textarea.style.opacity = '0';
  document.body.appendChild(textarea);
  textarea.select();
  document.execCommand('copy');
  document.body.removeChild(textarea);
  return Promise.resolve();
}

export function renderLicense(store: Store): HTMLElement {
  const container = document.createElement('section');
  container.className = 'screen screen--license';

  const header = document.createElement('div');
  header.className = 'stage__header';
  header.innerHTML = `
    <span class="stage__icon">${LICENSE_ICON}</span>
    <div>
      <h2 class="stage__title">License Agreement</h2>
      <p class="stage__subtitle">Please review before continuing.</p>
    </div>
  `;

  const stageBody = document.createElement('div');
  stageBody.className = 'stage__body';

  const licensePanel = document.createElement('div');
  licensePanel.className = 'scroll-panel';
  licensePanel.setAttribute('role', 'document');
  licensePanel.innerHTML = `
    <div class="license-prose">
      <p>End-User License Agreement for PolyForge.</p>
      <p>PLEASE READ THIS AGREEMENT CAREFULLY. IT CONTAINS IMPORTANT TERMS THAT AFFECT YOU AND YOUR USE OF THE SOFTWARE. BY INSTALLING, COPYING OR USING THE SOFTWARE, YOU AGREE TO BE BOUND BY THE TERMS OF THIS AGREEMENT. IF YOU DO NOT AGREE TO THESE TERMS, DO NOT INSTALL, COPY, OR USE THE SOFTWARE.</p>
      <p>This End-User License Agreement (EULA) is a legal agreement between you--either an individual or a single entity--and the author(s) of this Software for the product identified above, which includes computer software and may include associated media, and online or electronic documentation ("Software").</p>
      <p>By installing, copying, or otherwise using the Software, you agree to be bounded by the terms of this EULA. If you do not agree to the terms of this EULA, do not install or use the Software.</p>
      <p><strong>1. GRANT OF LICENSE.</strong></p>
      <p>This EULA grants you a non-exclusive, non-sublicensable, non-transferable license to install and use the Software. You may install and use an unlimited number of copies of the Software.</p>
      <p><strong>2. COPYRIGHT.</strong></p>
      <p>All title and copyrights in and to the Software (including but not limited to any images, libraries, and examples incorporated into the Software), the accompanying documentation, and any copies of the Software are owned by the author(s) of this Software.</p>
      <p><strong>3. NO WARRANTIES.</strong></p>
      <p>The author(s) of this Software expressly disclaims any warranty for the Software. The Software and any related documentation is provided "as is" without warranty of any kind, either express or implied, including, without limitation, the implied warranties or merchantability, fitness for a particular purpose, or noninfringement. The entire risk arising out of use or performance of the Software remains with you.</p>
      <p><strong>4. NO LIABILITY FOR DAMAGES.</strong></p>
      <p>In no event shall the author(s) of this Software be liable for any special, consequential, incidental or indirect damages whatsoever (including, without limitation, damages for loss of business profits, business interruption, loss of business information, or any other pecuniary loss) arising out of the use of or inability to use this product, even if the author(s) of this Software is aware of the possibility of such damages and known defects.</p>
    </div>
    <div class="scroll-panel__copy">
      <button type="button" class="copy-button" data-role="copy">Copy</button>
    </div>
  `;

  const toggle = document.createElement('label');
  toggle.className = 'toggle';
  toggle.innerHTML = `
    <input type="checkbox" class="toggle__input" hidden aria-hidden="true" />
    <span class="toggle__control" aria-hidden="true"></span>
    <span class="toggle__label">I accept the license agreement.</span>
  `;

  stageBody.appendChild(licensePanel);
  stageBody.appendChild(toggle);

  const footer = document.createElement('footer');
  footer.className = 'screen-footer';
  const social = createSocialLinks();
  const actions = document.createElement('div');
  actions.className = 'screen-footer__actions';
  const nextButton = document.createElement('button');
  nextButton.type = 'button';
  nextButton.className = 'btn btn--primary';
  nextButton.textContent = 'Next';
  nextButton.disabled = true;
  actions.appendChild(nextButton);
  footer.append(social, actions);

  container.append(header, stageBody, footer);

  const copyBtn = licensePanel.querySelector('[data-role="copy"]') as HTMLButtonElement;
  const toggleInput = toggle.querySelector('.toggle__input') as HTMLInputElement;
  const toggleControl = toggle.querySelector('.toggle__control') as HTMLSpanElement;

  const updateToggleState = () => {
    if (toggleInput.checked) {
      toggleControl.classList.add('is-active');
      toggleControl.innerHTML = CHECK_ICON;
    } else {
      toggleControl.classList.remove('is-active');
      toggleControl.innerHTML = '';
    }
    nextButton.disabled = !toggleInput.checked;
  };

  toggle.addEventListener('click', (event) => {
    if (event.target === toggleInput) return;
    event.preventDefault();
    toggleInput.checked = !toggleInput.checked;
    updateToggleState();
  });

  toggleInput.addEventListener('change', updateToggleState);
  updateToggleState();

  copyBtn.addEventListener('click', async () => {
    const original = copyBtn.textContent;
    try {
      await copyToClipboard(LICENSE_COPY);
      copyBtn.textContent = 'Copied!';
      window.setTimeout(() => {
        copyBtn.textContent = original ?? 'Copy';
      }, 1500);
    } catch (error) {
      console.error('Failed to copy license', error);
      copyBtn.textContent = 'Failed';
      window.setTimeout(() => {
        copyBtn.textContent = original ?? 'Copy';
      }, 1800);
    }
  });

  nextButton.addEventListener('click', () => {
    if (!toggleInput.checked) return;
    store.setStep(Step.Mode);
  });

  return container;
}
