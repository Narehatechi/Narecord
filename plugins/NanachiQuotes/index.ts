import { addChatBarButton, ChatBarButton, removeChatBarButton } from "@api/ChatButtons";
import { sendMessage } from "@utils/discord";
import definePlugin from "@utils/types";
import { createElement } from "react";
import { SelectedChannelStore } from "@webpack/common";

const quotes = [
    "Nnaa~ The Abyss is beautiful, but it never gives anything back.",
    "Don't look at me like that. I know what I'm doing.",
    "If you want to survive down here, keep your head and your supplies.",
    "The best discoveries are the ones you live to tell someone about.",
    "Nnaa. Take a breath, then take the next step."
];

function Ears() {
    return createElement("span", { "aria-hidden": true }, "⌁");
}

async function sendQuote() {
    const channelId = SelectedChannelStore.getChannelId();
    if (!channelId) return;
    await sendMessage(channelId, { content: quotes[Math.floor(Math.random() * quotes.length)]! });
}

export default definePlugin({
    name: "NanachiQuotes",
    description: "Chat bar button that sends a random Nanachi quote or bit of wisdom.",
    authors: [{ name: "Narecord", id: 0n }],
    dependencies: ["ChatInputButtonAPI"],
    start() {
        addChatBarButton("narecord-nanachi-quotes", () => createElement(
            ChatBarButton,
            { tooltip: "Nanachi wisdom", onClick: () => void sendQuote() },
            createElement(Ears)
        ), Ears);
    },
    stop() {
        removeChatBarButton("narecord-nanachi-quotes");
    }
});
