import * as React from 'react';
import * as ReactDOM from 'react-dom';
import { BrowserRouter } from 'react-router-dom';
import { makeApi } from './Api';
import App from './App';
import './index.css';
import registerServiceWorker from './registerServiceWorker';

const host = process.env.REACT_APP_API || process.env.API || window.location.host;
// tslint:disable-next-line:no-console
console.log(host);

ReactDOM.render(
  <BrowserRouter>
    <App api={makeApi(host)} />
  </BrowserRouter>,
  document.getElementById('root') as HTMLElement
);
registerServiceWorker();
