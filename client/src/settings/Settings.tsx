import * as React from 'react';

interface IProps {
  apiServer: string;
}

class Settings extends React.Component<IProps> {
  public render() {
    return (
      <div className="Settings">
        Settings
      </div>
    );
  }
}

export default Settings;
