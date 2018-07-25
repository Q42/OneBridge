import * as React from 'react';
import { Link } from 'react-router-dom';

class Menu extends React.Component {
  public render() {
    return (
      <div className="Menu">
        <Link to={`/map`}>Floorplan</Link>
        <Link to={`/settings`}>Settings</Link>
      </div>
    );
  }
}

export default Menu;
