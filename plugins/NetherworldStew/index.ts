import { addChatBarButton, ChatBarButton, removeChatBarButton } from "@api/ChatButtons";
import { sendMessage } from "@utils/discord";
import definePlugin from "@utils/types";
import { createElement } from "react";
import { SelectedChannelStore } from "@webpack/common";

const dishes = [
    "Hellbroom Broth",
    "Madwort Soup",
    "Hammerbeak Egg Stew",
    "Demonfish Roast",
    "Yellow-shining Grass Porridge",
    "Sun-Sphere Skewers",
    "Crimson Splitjaw Jerky"
];

function Bowl() {
    return createElement("span", { "aria-hidden": true }, "♨");
}

async function serve() {
    const channelId = SelectedChannelStore.getChannelId();
    if (!channelId) return;
    const dish = dishes[Math.floor(Math.random() * dishes.length)]!;
    await sendMessage(channelId, { content: `Tonight's Netherworld meal: **${dish}** 🍲` });
}

export default definePlugin({
    name: "NetherworldStew",
    description: "Chat button that posts a random Abyss food name from the field menu.",
    authors: [{ name: "Narecord", id: 0n }],
    dependencies: ["ChatInputButtonAPI"],
    start() {
        addChatBarButton("narecord-netherworld-stew", () => createElement(
            ChatBarButton,
            { tooltip: "Serve a Netherworld dish", onClick: () => void serve() },
            createElement(Bowl)
        ), Bowl);
    },
    stop() {
        removeChatBarButton("narecord-netherworld-stew");
    }
});
