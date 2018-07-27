import * as React from 'react';
import { IApi } from '../Api';

interface IProps {
  api: IApi;
}

class Map extends React.Component<IProps> {
  public render() {
    return (
      <div className="Map">
        <img className="App-logo" />
      </div>
    );
  }
}

export default Map;
