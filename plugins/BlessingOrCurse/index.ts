import { addChatBarButton, ChatBarButton, removeChatBarButton } from "@api/ChatButtons";
import { sendMessage } from "@utils/discord";
import definePlugin from "@utils/types";
import { createElement } from "react";
import { SelectedChannelStore } from "@webpack/common";

const messages = [
    "The Abyss has blessed this expedition. 🌟",
    "A gentle breeze from the surface follows you.",
    "The curse has noticed you. Keep climbing carefully.",
    "Your relic hums ominously. That cannot be good.",
    "Blessing: your next meal is probably edible.",
    "Curse: something is watching from the dark."
];

function Sigil() {
    return createElement("span", { "aria-hidden": true }, "☯");
}

async function sendFate() {
    const channelId = SelectedChannelStore.getChannelId();
    if (!channelId) return;
    await sendMessage(channelId, { content: messages[Math.floor(Math.random() * messages.length)]! });
}

export default definePlugin({
    name: "BlessingOrCurse",
    description: "Chat button: random Blessing or Curse message from the Abyss.",
    authors: [{ name: "Narecord", id: 0n }],
    dependencies: ["ChatInputButtonAPI"],
    start() {
        addChatBarButton("narecord-blessing-or-curse", () => createElement(
            ChatBarButton,
            { tooltip: "Ask the Abyss: blessing or curse?", onClick: () => void sendFate() },
            createElement(Sigil)
        ), Sigil);
    },
    stop() {
        removeChatBarButton("narecord-blessing-or-curse");
    }
});
