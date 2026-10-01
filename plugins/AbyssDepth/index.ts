import definePlugin from "@utils/types";
import "./style.css";
import { SelectedChannelStore } from "@webpack/common";

const LAYERS = [
    ["1st Layer", "Curiosity"],
    ["2nd Layer", "Temptation"],
    ["3rd Layer", "Vertigo"],
    ["4th Layer", "Longing"],
    ["5th Layer", "Silence"],
    ["6th Layer", "Unreturned"],
    ["7th Layer", "Finality"]
];
const ID = "nr-abyss-depth";
let timer: ReturnType<typeof setInterval> | undefined;
let observer: MutationObserver | undefined;

function layerIndex(channelId: string) {
    let value = 0;
    for (let i = 0; i < channelId.length; i++)
        value = (value + channelId.charCodeAt(i) * (i + 1)) % LAYERS.length;
    return value;
}

function update() {
    const titlebar = document.querySelector<HTMLElement>('[class*="titleBar"]');
    if (!titlebar) return;
    let indicator = document.getElementById(ID);
    if (!indicator) {
        indicator = document.createElement("span");
        indicator.id = ID;
        indicator.className = "nr-abyss-depth";
    }
    if (indicator.parentElement !== titlebar)
        titlebar.appendChild(indicator);
    const channelId = SelectedChannelStore.getChannelId() || "";
    const [layer, mood] = LAYERS[layerIndex(channelId)]!;
    const label = `${layer} · ${mood}`;
    if (indicator.textContent !== label)
        indicator.textContent = label;
    indicator.title = `Current channel mood: ${mood}`;
}

export default definePlugin({
    name: "AbyssDepth",
    description: "Thin Abyss depth strip shows the selected channel's layer mood.",
    authors: [{ name: "Narecord", id: 0n }],
    start() {
        observer = new MutationObserver(update);
        observer.observe(document.body, { childList: true, subtree: true });
        timer = setInterval(update, 3000);
        update();
    },
    stop() {
        if (timer) clearInterval(timer);
        timer = undefined;
        observer?.disconnect();
        observer = undefined;
        document.getElementById(ID)?.remove();
    }
});
