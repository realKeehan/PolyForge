export function renderLoading(): HTMLElement {
  const container = document.createElement('section');
  container.className = 'screen screen--loading';
  container.innerHTML = `
    <div class="loading">
      <div class="loading__spinner"></div>
      <p class="loading__text">Preparing installer…</p>
    </div>
  `;
  return container;
}
