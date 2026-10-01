import { definePluginSettings } from "@api/Settings";
import definePlugin, { OptionType } from "@utils/types";
import "./style.css";

const settings = definePluginSettings({
    keywords: {
        type: OptionType.STRING,
        description: "Comma-separated relic names and Abyss words to highlight.",
        default: "Star Compass, Blaze Reap, Reg, White Whistle, narehate, curse, blessing",
        onChange() { refresh(); }
    }
});

let observer: MutationObserver | undefined;

function clearHighlights() {
    document.querySelectorAll<HTMLElement>(".nr-relic-ping").forEach(mark => {
        mark.replaceWith(document.createTextNode(mark.textContent || ""));
    });
}

function keywordPattern() {
    const words = settings.store.keywords.split(",").map(word => word.trim()).filter(Boolean);
    if (!words.length) return null;
    const escaped = words.sort((a, b) => b.length - a.length).map(word => word.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"));
    return new RegExp(`\\b(${escaped.join("|")})\\b`, "gi");
}

function highlightMessage(message: Element) {
    const pattern = keywordPattern();
    if (!pattern) return;
    const walker = document.createTreeWalker(message, NodeFilter.SHOW_TEXT);
    const nodes: Text[] = [];
    while (walker.nextNode()) {
        const node = walker.currentNode as Text;
        const parent = node.parentElement;
        if (parent?.closest(".nr-relic-ping, code, pre, a, [contenteditable=true]")) continue;
        pattern.lastIndex = 0;
        if (pattern.test(node.data)) {
            pattern.lastIndex = 0;
            nodes.push(node);
        }
    }
    for (const node of nodes) {
        pattern.lastIndex = 0;
        const fragment = document.createDocumentFragment();
        let last = 0;
        for (const match of node.data.matchAll(pattern)) {
            const index = match.index!;
            if (index > last) fragment.appendChild(document.createTextNode(node.data.slice(last, index)));
            const mark = document.createElement("mark");
            mark.className = "nr-relic-ping";
            mark.textContent = match[0];
            fragment.appendChild(mark);
            last = index + match[0].length;
        }
        if (last) {
            if (last < node.data.length) fragment.appendChild(document.createTextNode(node.data.slice(last)));
            node.replaceWith(fragment);
        }
    }
}

function paint() {
    document.querySelectorAll('[class*="messageContent"]').forEach(highlightMessage);
}

function paintAdded(records: MutationRecord[]) {
    for (const record of records) {
        const parent = record.target instanceof Element ? record.target : record.target.parentElement;
        const message = parent?.closest('[class*="messageContent"]');
        if (message) {
            highlightMessage(message);
            continue;
        }
        for (const node of record.addedNodes) {
            if (!(node instanceof Element)) continue;
            if (node.matches('[class*="messageContent"]'))
                highlightMessage(node);
            else
                node.querySelectorAll('[class*="messageContent"]').forEach(highlightMessage);
        }
    }
}

function refresh() {
    clearHighlights();
    paint();
}

export default definePlugin({
    name: "RelicPing",
    description: "Highlights configured Abyss keywords such as relic names and curse words.",
    authors: [{ name: "Narecord", id: 0n }],
    settings,
    start() {
        observer = new MutationObserver(paintAdded);
        observer.observe(document.body, { childList: true, subtree: true, characterData: true });
        paint();
    },
    stop() {
        observer?.disconnect();
        observer = undefined;
        clearHighlights();
    }
});
