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

    render() {
        const elementButtonId = `${this.props.id}-element-btn`;
        const documentButtonId = `${this.props.id}-document-btn`;
        const sequenceButtonId = `${this.props.id}-sequence-btn`;
        const targetId = `${this.props.id}-element-target`;

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
                    className: "btn btn-info",
                    onClick: this.dispatchSequenceEvent.bind(this)
                }, [
                    "Sequence event"
                ])
            ]),
            e("div", { className: "card-body" }, [
                e("div", {
                    id: targetId,
                    className: "alert alert-success"
                }, [
                    "Observable target"
                ])
            ])
        ]);
    }
}
