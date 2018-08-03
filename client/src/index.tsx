import * as React from 'react';
import * as ReactDOM from 'react-dom';
import { BrowserRouter } from 'react-router-dom';
import { makeApi } from './Api';
import App from './App';
import './index.css';
import registerServiceWorker from './registerServiceWorker';

const apiHost = localStorage.getItem("apiHost") || process.env.REACT_APP_API || process.env.API || window.location.host;
const publicUrl = process.env.PUBLIC_URL;

ReactDOM.render(
  <BrowserRouter basename={publicUrl}>
    <App api={makeApi(apiHost)} />
  </BrowserRouter>,
  document.getElementById('root') as HTMLElement
);
registerServiceWorker();
