const e = React.createElement;

export default class ObservableComponent extends React.PureComponent {
    dispatchElementEvent() {
        const targetId = `${this.props.id}-element-target`;

        setTimeout(() => {
            const target = document.getElementById(targetId);

            if (target == null) {
                return;
            }

            target.dispatchEvent(new CustomEvent("ferret:element", {
                bubbles: true,
                detail: {
                    scope: "element"
                }
            }));
        }, 50);
    }

    dispatchDocumentEvent() {
        setTimeout(() => {
            document.dispatchEvent(new CustomEvent("ferret:document", {
                detail: {
                    scope: "document"
                }
            }));
        }, 50);
    }

    dispatchSequenceEvent() {
        const targetId = `${this.props.id}-element-target`;

        const emit = (index, delay) => {
            setTimeout(() => {
                const target = document.getElementById(targetId);

                if (target == null) {
                    return;
                }

                target.dispatchEvent(new CustomEvent("ferret:sequence", {
                    bubbles: true,
                    detail: {
                        index
                    }
                }));
            }, delay);
        };

        emit(1, 50);
        emit(2, 100);
    }

    dispatchDelegateEvent() {
        const rootId = `${this.props.id}-delegate-root`;

        setTimeout(() => {
            const root = document.getElementById(rootId);

            if (root == null) {
                return;
            }

            const other = root.querySelector(".observable-delegate-other");
            const match = root.querySelector(".observable-delegate-item");

            if (other != null) {
                other.dispatchEvent(new CustomEvent("ferret:delegate", {
                    bubbles: true,
                    detail: {
                        scope: "other"
                    }
                }));
            }

            if (match != null) {
                match.dispatchEvent(new CustomEvent("ferret:delegate", {
                    bubbles: true,
                    detail: {
                        scope: "match"
                    }
                }));
            }
        }, 50);
    }

    dispatchTargetSelectorEvent() {
        const rootId = `${this.props.id}-target-selector-root`;

        setTimeout(() => {
            const root = document.getElementById(rootId);

            if (root == null) {
                return;
            }

            const target = root.querySelector(".observable-target-selector-child");

            if (target == null) {
                return;
            }

            target.dispatchEvent(new CustomEvent("ferret:target-selector", {
                detail: {
                    scope: "child"
                }
            }));
        }, 50);
    }

    dispatchPropsEvent() {
        const targetId = `${this.props.id}-element-target`;

        setTimeout(() => {
            const target = document.getElementById(targetId);

            if (target == null) {
                return;
            }

            target.dispatchEvent(new CustomEvent("ferret:props", {
                bubbles: true,
                detail: {
                    scope: "props",
                    nested: {
                        enabled: true
                    }
                }
            }));
        }, 50);
    }

    dispatchDepthEvent() {
        const targetId = `${this.props.id}-element-target`;

        setTimeout(() => {
            const target = document.getElementById(targetId);

            if (target == null) {
                return;
            }

            target.dispatchEvent(new CustomEvent("ferret:depth", {
                bubbles: true,
                detail: {
                    level1: {
                        level2: {
                            level3: {
                                leaf: true
                            }
                        }
                    }
                }
            }));
        }, 50);
    }

    render() {
        const elementButtonId = `${this.props.id}-element-btn`;
        const documentButtonId = `${this.props.id}-document-btn`;
        const sequenceButtonId = `${this.props.id}-sequence-btn`;
        const delegateButtonId = `${this.props.id}-delegate-btn`;
        const targetSelectorButtonId = `${this.props.id}-target-selector-btn`;
        const propsButtonId = `${this.props.id}-props-btn`;
        const depthButtonId = `${this.props.id}-depth-btn`;
        const targetId = `${this.props.id}-element-target`;
        const delegateRootId = `${this.props.id}-delegate-root`;
        const targetSelectorRootId = `${this.props.id}-target-selector-root`;

        return e("div", { id: this.props.id, className: "card observable" }, [
            e("div", { className: "card-header" }, [
                e("button", {
                    id: elementButtonId,
                    className: "btn btn-primary mr-2",
                    onClick: this.dispatchElementEvent.bind(this)
                }, [
                    "Element event"
                ]),
                e("button", {
                    id: documentButtonId,
                    className: "btn btn-secondary mr-2",
                    onClick: this.dispatchDocumentEvent.bind(this)
                }, [
                    "Document event"
                ]),
                e("button", {
                    id: sequenceButtonId,
                    className: "btn btn-info mr-2 mb-2",
                    onClick: this.dispatchSequenceEvent.bind(this)
                }, [
                    "Sequence event"
                ]),
                e("button", {
                    id: delegateButtonId,
                    className: "btn btn-warning mr-2 mb-2",
                    onClick: this.dispatchDelegateEvent.bind(this)
                }, [
                    "Delegate event"
                ]),
                e("button", {
                    id: targetSelectorButtonId,
                    className: "btn btn-dark mr-2 mb-2",
                    onClick: this.dispatchTargetSelectorEvent.bind(this)
                }, [
                    "Target selector event"
                ]),
                e("button", {
                    id: propsButtonId,
                    className: "btn btn-outline-primary mr-2 mb-2",
                    onClick: this.dispatchPropsEvent.bind(this)
                }, [
                    "Props event"
                ]),
                e("button", {
                    id: depthButtonId,
                    className: "btn btn-outline-secondary mb-2",
                    onClick: this.dispatchDepthEvent.bind(this)
                }, [
                    "Depth event"
                ])
            ]),
            e("div", { className: "card-body" }, [
                e("div", {
                    id: targetId,
                    className: "alert alert-success mb-3"
                }, [
                    "Observable target"
                ]),
                e("div", {
                    id: delegateRootId,
                    className: "alert alert-warning mb-3"
                }, [
                    e("button", {
                        type: "button",
                        className: "btn btn-sm btn-outline-primary observable-delegate-item mr-2"
                    }, [
                        "Delegate match"
                    ]),
                    e("button", {
                        type: "button",
                        className: "btn btn-sm btn-outline-secondary observable-delegate-other"
                    }, [
                        "Delegate other"
                    ])
                ]),
                e("div", {
                    id: targetSelectorRootId,
                    className: "alert alert-info mb-0"
                }, [
                    e("button", {
                        type: "button",
                        className: "btn btn-sm btn-dark observable-target-selector-child"
                    }, [
                        "Target selector child"
                    ])
                ])
            ])
        ]);
    }
}
