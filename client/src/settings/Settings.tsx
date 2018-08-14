import * as React from 'react';
import { IApi } from '../Api';

interface IProps {
  api: IApi;
}

interface IState {
  bridges?: any[];
  error?: any;
}

class Settings extends React.Component<IProps,IState> {
  public componentWillMount(){
    this.setState({})
    this.refresh()
  }

  public refresh() {
    const { send, receiveOnce } = this.props.api;
    send(`{ "type": "request", "paths": ["bridges"] }`)
    receiveOnce((msg) => Array.isArray(msg) && (msg.length === 0 || msg[0] && msg[0].Bridge))
      .then(
        msg => this.setState({ ...this.state, bridges: msg }),
        (error) => this.setState({ ...this.state, error })
      )
  }

  public render() {
    return (
      <div className="Settings">
        <h2>Bridges</h2>
        {(this.state.bridges || []).map((args) => <div key={args.id}>{<Bridge api={this.props.api} {...args } />}</div>)}
      </div>
    );
  }
}

function connect(api: IApi, data: IApiBridge) {
  return () => {
    api.send(JSON.stringify({ type: "link", bridgeId: data.ID, bridgeMac: data.Mac, bridgeIp: data.IP }));
  }
}

// tslint:disable-next-line:max-classes-per-file
class Bridge extends React.Component<{Bridge: IApiBridge, Name: string, Linked: boolean, api: IApi}> {

  public render() {
    return (
      <div>
        {this.props.Name || this.props.Bridge.ID}
        {(!this.props.Linked) ? 
          (<button onClick={connect(this.props.api, this.props.Bridge)}>connect</button>)
        : (<div/>)}
      </div>
    );
  }
}

export default Settings;

interface IApiBridge {
  ID: string, IP: string, Mac: string;
}