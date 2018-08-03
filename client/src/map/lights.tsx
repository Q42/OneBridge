import * as React from "react";
import { Vector } from "react-designer";
import Icon from "react-designer/lib/Icon";

export class PointLight extends Vector {
  public static meta = {
    icon: <Icon icon={'circle'} size={30} />,
    initial: {
      height: 5,
      width: 5,
      strokeWidth: 0,
      fill: "yellow",
      radius: 5,
      blendMode: "normal"
    }
  };

  public render() {
    const { object } = this.props;
    return React.createElement(
      "rect", {
        style: this.getStyle(),
        ...this.getObjectAttributes(),
        rx: object.radius,
        width: object.width,
        height: object.height
      });
  }
}

// tslint:disable-next-line:max-classes-per-file
export class BarLight extends Vector {
  public static meta = {
    icon: <Icon icon={'rectangle'} size={30} />,
    initial: {
      height: 5,
      width: 5,
      strokeWidth: 0,
      fill: "yellow",
      radius: 5,
      blendMode: "normal"
    }
  };

  public render() {
    const { object } = this.props;
    return React.createElement(
      "rect", {
        style: this.getStyle(),
        ...this.getObjectAttributes(),
        rx: object.radius,
        width: object.width,
        height: object.height
      });
  }
}
