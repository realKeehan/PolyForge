import './styles.scss';
import { createApp } from './ui/App';

document.addEventListener('DOMContentLoaded', () => {
  const root = document.getElementById('app');
  if (!root) {
    throw new Error('App root not found');
  }
  createApp(root);
});
