import { ApplicationCommandInputType, ApplicationCommandOptionType, findOption } from "@api/Commands";
import * as DataStore from "@api/DataStore";
import { definePluginSettings } from "@api/Settings";
import definePlugin, { OptionType } from "@utils/types";
import { showToast, Toasts } from "@webpack/common";

import "./style.css";

const STORE = "narecord-narelogs";
const Narehatechi = { name: "Narehatechi", id: 1326338080696832010n };

type Row = { id: string; kind: "deleted" | "edited" | "sent"; content: string; at: number; };

let journal: Row[] = [];

function hex(v: string, fallback: string) {
    const t = (v || "").trim();
    return /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.test(t) ? t : fallback;
}

function sync() {
    const r = document.documentElement;
    const s = settings.store;
    r.classList.add("narelogs-root");
    r.classList.toggle("narelogs-strobe", s.motion && s.strobe);
    r.classList.toggle("narelogs-quake", s.motion && s.quake);
    r.classList.toggle("narelogs-boom", s.motion && s.boom);
    r.classList.toggle("narelogs-neon", s.motion && s.neon);
    r.classList.toggle("narelogs-ring", s.motion && s.ring);
    r.classList.toggle("narelogs-scan", s.motion && s.scan);
    r.classList.toggle("narelogs-glint", s.motion && s.glint);
    r.classList.toggle("narelogs-hide-box", !s.showBox);
    r.style.setProperty("--narelogs-gold", hex(s.gold, "#e4c04a"));
    r.style.setProperty("--narelogs-rose", hex(s.rose, "#8b3a44"));
    r.style.setProperty("--narelogs-cream", hex(s.cream, "#fff6e4"));
    r.style.setProperty("--narelogs-speed", String(s.speed || 1));
}

const settings = definePluginSettings({
    motion: { type: OptionType.BOOLEAN, description: "Master motion switch", default: true, onChange: sync },
    quake: { type: OptionType.BOOLEAN, description: "Shake the deleted message", default: true, onChange: sync },
    strobe: { type: OptionType.BOOLEAN, description: "Gold/rose flash on delete", default: true, onChange: sync },
    boom: { type: OptionType.BOOLEAN, description: "Pop/scale on delete", default: true, onChange: sync },
    neon: { type: OptionType.BOOLEAN, description: "Neon glow on text and label", default: true, onChange: sync },
    ring: { type: OptionType.BOOLEAN, description: "Ring pulse on avatars", default: true, onChange: sync },
    scan: { type: OptionType.BOOLEAN, description: "Scanline on the gone box", default: true, onChange: sync },
    glint: { type: OptionType.BOOLEAN, description: "Glint sweep on the gone box", default: true, onChange: sync },
    speed: {
        type: OptionType.SLIDER,
        description: "Animation speed (higher is faster)",
        default: 1,
        markers: [0.5, 1, 1.5, 2],
        onChange: sync
    },
    gold: { type: OptionType.STRING, description: "Gold hex", default: "#e4c04a", onChange: sync },
    rose: { type: OptionType.STRING, description: "Rose hex", default: "#8b3a44", onChange: sync },
    cream: { type: OptionType.STRING, description: "Cream hex", default: "#fff6e4", onChange: sync },
    kicker: { type: OptionType.STRING, description: "Label on deleted messages", default: "gone" },
    goneText: { type: OptionType.STRING, description: "Line under the label", default: "Nnaa. This one was deleted." },
    showBox: { type: OptionType.BOOLEAN, description: "Show the gone box under deleted messages", default: true, onChange: sync },
    toastDeletes: { type: OptionType.BOOLEAN, description: "Toast when a message is deleted", default: false },
    persistSends: { type: OptionType.BOOLEAN, description: "Also keep sent messages in /narelogs", default: false }
});

async function remember(row: Row) {
    journal = [row, ...journal].slice(0, 80);
    await DataStore.set(STORE, journal);
}

function goneBox() {
    const box = document.createElement("div");
    box.className = "narelogs-gone";
    const bar = document.createElement("span");
    bar.className = "narelogs-gone-bar";
    const kicker = document.createElement("span");
    kicker.className = "narelogs-gone-kicker";
    kicker.textContent = settings.store.kicker || "gone";
    const line = document.createElement("span");
    line.className = "narelogs-gone-text";
    line.textContent = settings.store.goneText || "Nnaa. This one was deleted.";
    box.append(bar, kicker, line);
    return box;
}

function markDeleted(el: Element) {
    el.classList.add("narelogs-deleted");
    if (!settings.store.showBox) return;
    if (el.querySelector(":scope > .narelogs-gone")) return;
    el.appendChild(goneBox());
}

function paint() {
    document.querySelectorAll(".messagelogger-deleted, [class*='deleted']").forEach(el => {
        if (el.closest("[class*='messagesWrapper'], [class*='scrollerInner']")) markDeleted(el);
    });
}

let obs: MutationObserver | null = null;

function watch() {
    obs?.disconnect();
    obs = new MutationObserver(() => paint());
    obs.observe(document.body, { childList: true, subtree: true, attributes: true, attributeFilter: ["class"] });
    paint();
}

export default definePlugin({
    name: "Narelogs",
    description: "Nanachi journal and skin on top of MessageLogger. Motion, colors, and gone-text are all in settings.",
    authors: [Narehatechi],
    settings,
    dependencies: ["CommandsAPI"],
    commands: [
        {
            name: "narelogs",
            description: "Dump recent deleted/edited log",
            inputType: ApplicationCommandInputType.BUILT_IN,
            options: [{ name: "count", description: "How many", type: ApplicationCommandOptionType.INTEGER, required: false }],
            execute: opts => {
                const n = Math.max(1, Math.min(20, Number(findOption(opts, "count")) || 8));
                if (!journal.length) return { content: "Nnaa. Journal's empty." };
                const lines = journal.slice(0, n).map((r, i) => `${i + 1}. [${r.kind}] ${r.content || "(no text)"}`);
                return { content: "**Narelogs**\n" + lines.join("\n") };
            }
        }
    ],
    flux: {
        MESSAGE_DELETE: ({ id, message }: any) => {
            const content = message?.content || "";
            void remember({ id: id || message?.id || String(Date.now()), kind: "deleted", content, at: Date.now() });
            if (settings.store.toastDeletes) showToast("Nnaa. This one was deleted.", Toasts.Type.MESSAGE);
            requestAnimationFrame(paint);
        },
        MESSAGE_UPDATE: ({ message }: any) => {
            if (!message?.id) return;
            void remember({ id: message.id, kind: "edited", content: message.content || "", at: Date.now() });
        }
    },
    onBeforeMessageSend(_id, message) {
        if (!settings.store.persistSends || !message?.content) return;
        void remember({ id: String(Date.now()), kind: "sent", content: message.content, at: Date.now() });
    },
    async start() {
        journal = (await DataStore.get(STORE)) ?? [];
        sync();
        watch();
    },
    stop() {
        obs?.disconnect();
        obs = null;
        const r = document.documentElement;
        r.classList.remove("narelogs-root", "narelogs-strobe", "narelogs-quake", "narelogs-boom", "narelogs-neon", "narelogs-ring", "narelogs-scan", "narelogs-glint", "narelogs-hide-box");
        r.style.removeProperty("--narelogs-gold");
        r.style.removeProperty("--narelogs-rose");
        r.style.removeProperty("--narelogs-cream");
        r.style.removeProperty("--narelogs-speed");
        document.querySelectorAll(".narelogs-gone").forEach(el => el.remove());
    }
});
