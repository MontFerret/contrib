const e = React.createElement;

export default class PressableComponent extends React.PureComponent {
    constructor(props) {
        super(props);

        this.inputRef = React.createRef();
        this.handleKeyDown = this.handleKeyDown.bind(this);
        this.handleReset = this.handleReset.bind(this);

        this.state = {
            key: ''
        };
    }

    componentDidMount() {
        if (this.inputRef.current) {
            this.inputRef.current.addEventListener('keydown', this.handleKeyDown);
        }
    }

    componentWillUnmount() {
        if (this.inputRef.current) {
            this.inputRef.current.removeEventListener('keydown', this.handleKeyDown);
        }
    }

    handleKeyDown(e) {
        if (e.key === 'Unidentified') {
            return;
        }

        this.setState((prevState) => ({
            key: prevState.key ? prevState.key + ' + ' + e.key : e.key
        }))
    }

    handleReset() {
        this.setState({ key: '' })
    }

    render() {
        const inputId = `${this.props.id}-input`;
        const contentId = `${this.props.id}-content`;
        const classNames = ["alert", "alert-success"];

        return e("div", { id: this.props.id, className: "card clickable"}, [
            e("div", { className: "card-header"}, [
                e("div", { className: "form-group" }, [
                    e("label", null, "Pressable"),
                    e("input", {
                        id: inputId,
                        type: "text",
                        className: "form-control",
                        ref: this.inputRef
                    }),
                    e("input", {
                        type: "button",
                        className: "btn btn-primary",
                        onClick: this.handleReset,
                        value: "Reset"
                    },
                    )
                ]),
            ]),
            e("div", { className: "card-body"}, [
                e("div", { className: classNames.join(" ")}, [
                    e("p", {
                        id: contentId
                    }, [
                        this.state.key
                    ])
                ])
            ])
        ]);
    }
}
