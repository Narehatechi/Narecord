import definePlugin from "@utils/types";
import "./style.css";

const ID = "nr-rainbow-stars";

export default definePlugin({
    name: "RainbowStars",
    description: "Rainbow 5-point stars add quiet static color to the den.",
    authors: [{ name: "Narecord", id: 0n }],
    start() {
        if (document.getElementById(ID)) return;
        const sky = document.createElement("div");
        sky.id = ID;
        sky.className = "nr-rainbow-stars";
        sky.setAttribute("aria-hidden", "true");
        for (let i = 0; i < 7; i++) {
            const star = document.createElement("span");
            star.className = "nr-rainbow-star";
            star.style.left = `${8 + i * 13}%`;
            star.style.top = `${12 + (i * 29) % 76}%`;
            star.style.color = `hsl(${i * 52}, 95%, 70%)`;
            sky.appendChild(star);
        }
        document.body.appendChild(sky);
    },
    stop() {
        document.getElementById(ID)?.remove();
    }
});
