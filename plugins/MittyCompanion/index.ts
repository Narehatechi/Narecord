import definePlugin from "@utils/types";
import "./style.css";

const ID = "nr-mitty-companion";

export default definePlugin({
    name: "MittyCompanion",
    description: "Small Mitty flair keeps a cute companion indicator nearby.",
    authors: [{ name: "Narecord", id: 0n }],
    start() {
        if (document.getElementById(ID)) return;
        const companion = document.createElement("div");
        companion.id = ID;
        companion.className = "nr-mitty-companion";
        companion.setAttribute("aria-label", "Mitty is here");
        companion.textContent = "♡";
        document.body.appendChild(companion);
    },
    stop() {
        document.getElementById(ID)?.remove();
    }
});
