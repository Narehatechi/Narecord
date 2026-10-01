import definePlugin from "@utils/types";
import "./style.css";

const ID = "nr-rainbow-fall";

export default definePlugin({
    name: "RainbowFall",
    description: "Rainbow stars and orbs fall softly through the hideout background.",
    authors: [{ name: "Narecord", id: 0n }],
    start() {
        if (document.getElementById(ID)) return;
        const sky = document.createElement("div");
        sky.id = ID;
        sky.className = "nr-rainbow-fall";
        sky.setAttribute("aria-hidden", "true");
        for (let i = 0; i < 18; i++) {
            const orb = document.createElement("span");
            orb.className = "nr-rainbow-orb";
            orb.style.setProperty("--nr-x", `${Math.random() * 100}%`);
            orb.style.setProperty("--nr-hue", `${Math.floor(Math.random() * 360)}`);
            orb.style.setProperty("--nr-delay", `${Math.random() * -18}s`);
            orb.style.setProperty("--nr-duration", `${12 + Math.random() * 14}s`);
            orb.style.setProperty("--nr-size", `${3 + Math.random() * 6}px`);
            sky.appendChild(orb);
        }
        document.body.appendChild(sky);
    },
    stop() {
        document.getElementById(ID)?.remove();
    }
});
