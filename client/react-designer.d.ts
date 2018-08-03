import { Component } from "react";

declare module "react-designer" {

	interface DesignerProps {
		width: number;
		height: number;
		objectTypes: { [key: string]: any }
		onUpdate: (object: any[]) => void
		objects: any[]
	}

  interface IVectorProps {
    object: any;
    index: number;
  }
	interface IIconProps {
		icon: string;
		size: number;
	}

	export default class Designer extends Component<DesignerProps, {}> {

	}

	export class Vector<P extends IVectorProps = IVectorProps, S = any> extends Component<P, S> {
    getStyle(): any;
    getObjectAttributes(): any;
  }
	export class Text extends Vector {}
	export class Rect extends Vector {}

	export class Icon extends Component<IIconProps> {}

}
