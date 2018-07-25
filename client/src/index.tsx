import * as React from 'react';
import * as ReactDOM from 'react-dom';
import { BrowserRouter } from 'react-router-dom';
import App from './App';
import './index.css';
import registerServiceWorker from './registerServiceWorker';

const apiServer = process.env.REACT_APP_API || process.env.API || window.location.host;
// tslint:disable-next-line:no-console
console.log(apiServer);

const ws = new WebSocket(`ws://${apiServer}/ws`);
// tslint:disable:no-console
ws.onclose = console.log.bind(console, "close")
ws.onerror = console.log.bind(console, "error")
ws.onopen = console.log.bind(console, "open")
ws.onmessage = console.log.bind(console, "message")

ReactDOM.render(
  <BrowserRouter>
    <App apiServer={apiServer} />
  </BrowserRouter>,
  document.getElementById('root') as HTMLElement
);
registerServiceWorker();
