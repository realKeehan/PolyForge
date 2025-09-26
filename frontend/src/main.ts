import './styles.scss';
import { createApp } from './ui/App';

document.addEventListener('DOMContentLoaded', () => {
  const root = document.getElementById('app');
  if (!root) {
    throw new Error('App root not found');
  }

  createApp(root).catch((error) => {
    console.error('Failed to bootstrap PolyForge', error);
    root.innerHTML = `
      <section class="boot-error">
        <h1 class="boot-error__title">Something went wrong</h1>
        <p class="boot-error__message">The installer UI could not be initialised. Check the developer tools console for details and restart the application.</p>
      </section>
    `;
  });
});
