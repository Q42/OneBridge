import * as React from 'react';

interface IProps {
  apiServer: string;
}

class Map extends React.Component<IProps> {
  public render() {
    return (
      <div className="Map">
        <img className="App-logo" alt="logo" />
      </div>
    );
  }
}

export default Map;
