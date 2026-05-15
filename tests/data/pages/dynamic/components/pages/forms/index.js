const e = React.createElement;

export default class FormsPage extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            textInput: "",
            select: "",
            multiSelect: "",
            textarea: "",
            checkbox: "unchecked",
            windowScroll: "0",
            containerScroll: "0"
        };

        this.handleWindowScroll = () => {
            this.setState({
                windowScroll: String(Math.round(window.scrollY || window.pageYOffset || 0))
            });
        };

        this.handleTextInput = (evt) => {
            evt.preventDefault();

            this.setState({
                textInput: evt.target.value
            });
        };

        this.handleSelect = (evt) => {
            evt.preventDefault();

            this.setState({
                select: evt.target.value
            });
        };

        this.handleMultiSelect = (evt) => {
            evt.preventDefault();

            this.setState({
                multiSelect: Array.prototype.map.call(evt.target.selectedOptions, i => i.value).join(", ")
            });
        };

        this.handleTextarea = (evt) => {
            evt.preventDefault();

            this.setState({
                textarea: evt.target.value
            });
        };

        this.handleCheckbox = (evt) => {
            this.setState({
                checkbox: evt.target.checked ? "checked" : "unchecked"
            });
        };

        this.handleContainerScroll = (evt) => {
            this.setState({
                containerScroll: String(Math.round(evt.target.scrollTop))
            });
        };
    }

    componentDidMount() {
        window.addEventListener("scroll", this.handleWindowScroll);
    }

    componentWillUnmount() {
        window.removeEventListener("scroll", this.handleWindowScroll);
    }

    render() {
        return e("form", { id: "page-form" }, [
            e("div", { className: "form-group" }, [
                e("label", null, "Text input"),
                e("input", {
                    id: "text_input",
                    type: "text",
                    className: "form-control",
                    onChange: this.handleTextInput
                }),
                e("small", {
                    id: "text_output",
                    className: "form-text text-muted"
                },
                    this.state.textInput
                )
            ]),
            e("div", { className: "form-group" }, [
                e("label", null, "Select"),
                e("select", {
                    id: "select_input",
                    className: "form-control",
                    onChange: this.handleSelect
                    },
                    [
                        e("option", null, 1),
                        e("option", null, 2),
                        e("option", null, 3),
                        e("option", null, 4),
                        e("option", null, 5),
                    ]
                ),
                e("small", {
                        id: "select_output",
                        className: "form-text text-muted"
                    }, this.state.select
                )
            ]),
            e("div", { className: "form-group" }, [
                e("label", null, "Multi select"),
                e("select", {
                        id: "multi_select_input",
                        multiple: true,
                        className: "form-control",
                        onChange: this.handleMultiSelect
                    },
                    [
                        e("option", null, 1),
                        e("option", null, 2),
                        e("option", null, 3),
                        e("option", null, 4),
                        e("option", null, 5),
                    ]
                ),
                e("small", {
                        id: "multi_select_output",
                        className: "form-text text-muted"
                    }, this.state.multiSelect
                )
            ]),
            e("div", { className: "form-group" }, [
                e("label", null, "Textarea"),
                e("textarea", {
                        id: "textarea_input",
                        rows:"5",
                        className: "form-control",
                        onChange: this.handleTextarea
                    }
                ),
                e("small", {
                        id: "textarea_output",
                        className: "form-text text-muted"
                    }, this.state.textarea
                )
            ]),
            e("div", { className: "form-group" }, [
                e("label", null, [
                    e("input", {
                        id: "checkbox_input",
                        type: "checkbox",
                        onChange: this.handleCheckbox
                    }),
                    " Checkbox"
                ]),
                e("small", {
                        id: "checkbox_output",
                        className: "form-text text-muted"
                    }, this.state.checkbox
                )
            ]),
            e("div", { className: "form-group" }, [
                e("small", {
                    id: "scroll_output",
                    className: "form-text text-muted"
                }, this.state.windowScroll),
                e("div", {
                    id: "scroll_container",
                    onScroll: this.handleContainerScroll,
                    style: {
                        height: "80px",
                        overflowY: "auto",
                        border: "1px solid #ddd"
                    }
                }, [
                    e("div", {
                        id: "scroll_container_inner",
                        style: {
                            height: "500px"
                        }
                    }, "Scrollable content")
                ]),
                e("small", {
                    id: "scroll_container_output",
                    className: "form-text text-muted"
                }, this.state.containerScroll)
            ]),
            e("div", {
                id: "scroll_spacer",
                style: {
                    height: "1800px"
                }
            })
        ])
    }
}
