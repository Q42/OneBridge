import * as React from 'react';
import Designer, {Rect, Text} from 'react-designer';
import { IApi } from '../Api';

interface IProps {
  api: IApi;
}

class Map extends React.Component<IProps> {

  public state = {
    objects: [
      {type: "text", x: 10, y: 20, text: "Hello!", fill: "red"},
      {type: "rect", x: 50, y: 70, width: 30, height: 40, fill: "red"}
    ]
  };

  public render() {
    return (
      <div className="Map">
        <Designer width={500} height={500}
          objectTypes={{
            'rect': Rect,
            'text': Text,
          }}
          onUpdate={(objects: any) => this.setState({objects})}
          objects={this.state.objects} />
      </div>
    );
  }
}

export default Map;
