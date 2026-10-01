import { addChatBarButton, ChatBarButton, removeChatBarButton } from "@api/ChatButtons";
import * as DataStore from "@api/DataStore";
import definePlugin from "@utils/types";
import { createElement } from "react";
import { showToast, Toasts } from "@webpack/common";

const KEY = "narecord-whistle-rank";
const RANKS = ["White", "Red", "Blue", "Yellow", "Black", "Sovereign"];
let rankIndex = 0;
let loaded = false;

async function loadRank() {
    const saved = await DataStore.get(KEY);
    if (typeof saved === "number" && Number.isInteger(saved) && saved >= 0 && saved < RANKS.length)
        rankIndex = saved;
    loaded = true;
}

function Whistle() {
    return createElement("span", { "aria-hidden": true }, "♢");
}

async function advanceRank() {
    if (!loaded) await loadRank();
    rankIndex = (rankIndex + 1) % RANKS.length;
    await DataStore.set(KEY, rankIndex);
    showToast(`Whistle rank: ${RANKS[rankIndex]} · keep exploring!`, Toasts.Type.SUCCESS);
}

export default definePlugin({
    name: "WhistleRank",
    description: "Tracks a playful local whistle rank from White through Sovereign.",
    authors: [{ name: "Narecord", id: 0n }],
    dependencies: ["ChatInputButtonAPI"],
    async start() {
        await loadRank();
        addChatBarButton("narecord-whistle-rank", () => createElement(
            ChatBarButton,
            { tooltip: `Whistle rank: ${RANKS[rankIndex]} · click to advance`, onClick: () => void advanceRank() },
            createElement(Whistle)
        ), Whistle);
    },
    stop() {
        removeChatBarButton("narecord-whistle-rank");
        loaded = false;
    }
});
