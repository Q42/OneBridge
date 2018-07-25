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
    const { send, receiveOnce } = this.props.api;
    send(`{ "type": "request", "paths": ["bridges"] }`)
    receiveOnce((msg) => Array.isArray(msg) && (msg.length === 0 || msg[0].id && msg[0].internalipaddress))
      .then(
        msg => this.setState({ ...this.state, bridges: msg }),
        (error) => this.setState({ ...this.state, error })
      )
  }

  public render() {
    return (
      <div className="Settings">
        Settings {this.props.api.host}
        {(this.state.bridges || []).map(renderBridge(this.props.api))}
      </div>
    );
  }
}

function renderBridge(api: IApi) {
  return ({ id, internalipaddress }: { id: string, internalipaddress: string }) => <div key={id}>{id} <button onClick={connect(api, { id, internalipaddress })}>connect</button></div>
}

function connect(api: IApi, { id, internalipaddress }: { id: string, internalipaddress: string }) {
  return () => {
    api.send(JSON.stringify({ type: "link", id, internalipaddress }));
  }
}

export default Settings;
