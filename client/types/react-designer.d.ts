declare module "react-designer" {

	interface DesignerProps {
		width: number;
		height: number;
		objectTypes: { [key: string]: Vector }
		onUpdate: (object: any[]) => void
		objects: any[]
	}

	export default class Designer extends React.Component<DesignerProps, {}> {

	}

	export class Text implements Vector {}
	export class Rect implements Vector {}
	export interface Vector {}

}