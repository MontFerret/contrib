import AppComponent from "./components/app.js";
import {parse} from "./utils/qs.js";

const qs = parse(location.search.charAt(0) === "?" ? location.search.slice(1) : location.search);

ReactDOM.render(
    React.createElement(AppComponent, qs),
    document.getElementById("root")
);
