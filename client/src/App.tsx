import * as React from 'react';
import { Redirect, Route, Switch } from 'react-router';
import './App.css';
import Map from './map/Map';
import Menu from './Menu';
import Settings from './settings/Settings';

interface IProps {
  apiServer: string;
}

class App extends React.Component<IProps> {
  public render() {
    return (
      <div className="App">
        <header className="App-header">
          <h1>OneBridge</h1>
          <h2>{this.props.apiServer}</h2>
          <Menu />
        </header>
        <Switch>
          <Route exact={true} path='/'><Redirect from="/" to="/map" /></Route>
          <Route path='/map' component={Map}/>
          <Route path='/settings' component={Settings}/>
        </Switch>
      </div>
    );
  }
}

export default App;
