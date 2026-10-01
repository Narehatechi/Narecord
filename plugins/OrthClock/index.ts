import definePlugin from "@utils/types";
import "./style.css";

const ID = "nr-orth-clock";
let timer: ReturnType<typeof setInterval> | undefined;

function updateClock() {
    let clock = document.getElementById(ID);
    if (!clock) {
        clock = document.createElement("time");
        clock.id = ID;
        clock.className = "nr-orth-clock";
        clock.setAttribute("aria-label", "Local Orth time");
        document.body.appendChild(clock);
    }
    clock.textContent = `ORTH · ${new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}`;
}

export default definePlugin({
    name: "OrthClock",
    description: "Small Orth-time clock widget displays your local time.",
    authors: [{ name: "Narecord", id: 0n }],
    start() {
        updateClock();
        timer = setInterval(updateClock, 15000);
    },
    stop() {
        if (timer) clearInterval(timer);
        timer = undefined;
        document.getElementById(ID)?.remove();
    }
});
