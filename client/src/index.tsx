import * as React from 'react';
import * as ReactDOM from 'react-dom';
import { BrowserRouter } from 'react-router-dom';
import App from './App';
import './index.css';
import registerServiceWorker from './registerServiceWorker';

ReactDOM.render(
  <BrowserRouter>
    <App apiServer="192.168.178.103:8080" />
  </BrowserRouter>,
  document.getElementById('root') as HTMLElement
);
registerServiceWorker();
