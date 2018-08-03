import * as React from 'react';
import Designer, { } from 'react-designer';
import { IApi } from '../Api';
import { BarLight, PointLight } from './lights';
import './Map.css';

interface IProps {
  api: IApi;
}

class Map extends React.Component<IProps> {

  public state = {
    objects: [
      {type: "pointLight", x: 50, y: 70, width: 30, height: 40, fill: "red"},
      {type: "barLight", x: 50, y: 70, width: 30, height: 40, fill: "red"}
    ]
  };

  public render() {
    return (
      <div className="Map">
        <Designer width={800} height={400}
          objectTypes={{
            'pointLight': PointLight,
            'barLight': BarLight,
          }}
          onUpdate={(objects: any) => this.setState({objects})}
          objects={this.state.objects} />
      </div>
    );
  }
}

export default Map;
