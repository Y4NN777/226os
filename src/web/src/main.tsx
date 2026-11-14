import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { App } from './app/App';
import { removeLegacyDisplayName } from './app/storage';
import './styles/index.css';

removeLegacyDisplayName();

const container = document.getElementById('root');
if (!container) throw new Error('#root element not found');

createRoot(container).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
