import * as React from 'react';
import { Redirect, Route, Switch } from 'react-router';
import { IApi } from './Api';
import './App.css';
import Map from './map/Map';
import Menu from './Menu';
import Settings from './settings/Settings';

interface IProps {
  api: IApi;
}

class App extends React.Component<IProps> {

  public state = { apiHost: this.props.api.host };

  constructor(props: Readonly<IProps>) {
    super(props);
    this.urlDialog = this.urlDialog.bind(this);
  }

  public urlDialog() {
    localStorage.setItem("apiHost", window.prompt("New API url") || this.state.apiHost);
    window.location.href = window.location.href;
  }

  public render() {
    // tslint:disable:jsx-no-lambda
    return (
      <div className="App">
        <header className="App-header">
          <h1>OneBridge</h1>
          <h2 onClick={this.urlDialog}>{this.props.api.host}</h2>
          <Menu />
        </header>
        <Switch>
          <Route exact={true} path='/'><Redirect from="/" to="/map" /></Route>
          <Route path='/map' render={(props) => (<Map {...props} api={this.props.api} />)} />
          <Route path='/settings' render={(props) => (<Settings {...props} api={this.props.api} />)} />
        </Switch>
      </div>
    );
  }
}

export default App;
