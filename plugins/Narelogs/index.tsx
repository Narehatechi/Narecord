import { ApplicationCommandInputType, ApplicationCommandOptionType, findOption } from "@api/Commands";
import * as DataStore from "@api/DataStore";
import type { MessageObject } from "@api/MessageEvents";
import { definePluginSettings } from "@api/Settings";
import { Button } from "@components/Button";
import definePlugin, { OptionType } from "@utils/types";
import { MessageStore, SelectedChannelStore, showToast, Toasts } from "@webpack/common";

import "./style.css";

const STORE = "narecord-narelogs";
const Narehatechi = { name: "Narehatechi", id: 1326338080696832010n };
const DAY_MS = 86400000;
const MAX_CACHE = 500;
const PAINT_DEBOUNCE_MS = 150;
/** Discord mounts the new channel's scroller asynchronously after CHANNEL_SELECT fires, so we re-attach once more shortly after to catch it. */
const CHANNEL_SWITCH_REWATCH_MS = 400;
const MAX_JOURNAL_ENTRIES = 80;

type Row = {
    id: string;
    kind: "deleted" | "edited" | "sent";
    content: string;
    before?: string;
    channelId?: string;
    at: number;
};

/** Minimal shapes for the Flux events Narelogs cares about (real payloads carry more). */
interface FluxMessageLike {
    id: string;
    channel_id: string;
    content?: string;
}
interface MessageCreateEvent {
    message: FluxMessageLike;
}
interface MessageDeleteEvent {
    id: string;
    channelId: string;
    guildId?: string;
    mlDeleted?: boolean;
}
interface MessageUpdateEvent {
    message: FluxMessageLike;
}

let journal: Row[] = [];
/** Tracks the last-known content per message so edits can show a before/after diff. */
const contentCache = new Map<string, string>();
const KIND_ICON: Record<Row["kind"], string> = { deleted: "🗑️", edited: "✏️", sent: "📤" };

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

async function exportJournal() {
    try {
        const blob = new Blob([JSON.stringify(journal, null, 2)], { type: "application/json" });
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `narelogs-${Date.now()}.json`;
        document.body.appendChild(a);
        a.click();
        a.remove();
        URL.revokeObjectURL(url);
        showToast("Nnaa. Exported the journal.", Toasts.Type.SUCCESS);
    } catch {
        showToast("Nnaa. Couldn't export the journal.", Toasts.Type.FAILURE);
    }
}

function NarelogsAbout() {
    return (
        <Button onClick={exportJournal}>
            Export journal as JSON
        </Button>
    );
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
    persistSends: { type: OptionType.BOOLEAN, description: "Also keep sent messages in /narelogs", default: false },
    maxAgeDays: {
        type: OptionType.NUMBER,
        description: "Keep entries for N days (0 disables expiry)",
        default: 14
    }
});

function cacheKey(channelId: string | undefined, id: string) {
    return `${channelId ?? ""}:${id}`;
}

function cacheSet(key: string, value: string) {
    if (!contentCache.has(key) && contentCache.size >= MAX_CACHE) {
        const oldest = contentCache.keys().next().value;
        if (oldest !== undefined) contentCache.delete(oldest);
    }
    contentCache.set(key, value);
}

function pruneOld() {
    const maxDays = Number(settings.store.maxAgeDays) || 0;
    if (maxDays <= 0) return;
    const cutoff = Date.now() - maxDays * DAY_MS;
    journal = journal.filter(r => r.at >= cutoff);
}

async function remember(row: Row) {
    journal = [row, ...journal].slice(0, MAX_JOURNAL_ENTRIES);
    await DataStore.set(STORE, journal);
}

function formatRelative(ts: number) {
    const diff = Date.now() - ts;
    const sec = Math.floor(diff / 1000);
    if (sec < 60) return "just now";
    const min = Math.floor(sec / 60);
    if (min < 60) return `${min}m ago`;
    const hr = Math.floor(min / 60);
    if (hr < 24) return `${hr}h ago`;
    const day = Math.floor(hr / 24);
    if (day < 30) return `${day}d ago`;
    return new Date(ts).toLocaleDateString();
}

function truncate(text: string, max = 240) {
    const t = text || "(no text)";
    return t.length > max ? t.slice(0, max) + "…" : t;
}

function formatRow(r: Row, i: number) {
    const icon = KIND_ICON[r.kind] ?? "•";
    const when = formatRelative(r.at);
    if (r.kind === "edited" && r.before) {
        return `${i + 1}. ${icon} edited · ${when}\n~~${truncate(r.before)}~~ → ${truncate(r.content)}`;
    }
    return `${i + 1}. ${icon} ${r.kind} · ${when} — ${truncate(r.content)}`;
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

/** Elements we've already classified, so repeated mutations don't re-scan/re-animate them. */
const processedNodes = new WeakSet<Element>();

function markDeleted(el: Element) {
    processedNodes.add(el);
    el.classList.add("narelogs-deleted");
    if (!settings.store.showBox) return;
    if (el.querySelector(":scope > .narelogs-gone")) return;
    el.appendChild(goneBox());
}

function findScrollerRoot(): ParentNode {
    return document.querySelector("[class*='scrollerInner']")
        ?? document.querySelector("[class*='chatContent']")
        ?? document.body;
}

function paint() {
    const root = findScrollerRoot();
    root.querySelectorAll(".messagelogger-deleted, [class*='deleted']").forEach(el => {
        if (processedNodes.has(el)) return;
        if (el.closest("[class*='messagesWrapper'], [class*='scrollerInner']")) markDeleted(el);
    });
}

let paintTimer: ReturnType<typeof setTimeout> | null = null;

function schedulePaint() {
    if (paintTimer) return;
    paintTimer = setTimeout(() => {
        paintTimer = null;
        paint();
    }, PAINT_DEBOUNCE_MS);
}

let obs: MutationObserver | null = null;
let watchRoot: ParentNode | null = null;

function watch() {
    const root = findScrollerRoot();
    if (root === watchRoot && obs) {
        paint();
        return;
    }
    obs?.disconnect();
    watchRoot = root;
    obs = new MutationObserver(schedulePaint);
    obs.observe(root, { childList: true, subtree: true, attributes: true, attributeFilter: ["class"] });
    paint();
}

export default definePlugin({
    name: "Narelogs",
    description: "Nanachi journal and skin on top of MessageLogger. Motion, colors, and gone-text are all in settings.",
    authors: [Narehatechi],
    settings,
    dependencies: ["CommandsAPI", "MessageLogger"],
    settingsAboutComponent: NarelogsAbout,
    commands: [
        {
            name: "narelogs",
            description: "Dump recent deleted/edited/sent log",
            inputType: ApplicationCommandInputType.BUILT_IN,
            options: [
                { name: "count", description: "How many", type: ApplicationCommandOptionType.INTEGER, required: false },
                { name: "here", description: "Only show this channel's entries", type: ApplicationCommandOptionType.BOOLEAN, required: false }
            ],
            execute: opts => {
                const n = Math.max(1, Math.min(20, Number(findOption(opts, "count")) || 8));
                const onlyThisChannel = Boolean(findOption(opts, "here", false));
                const currentChannelId = onlyThisChannel ? SelectedChannelStore.getChannelId() : undefined;
                const rows = currentChannelId ? journal.filter(r => r.channelId === currentChannelId) : journal;
                if (!rows.length) return { content: "Nnaa. Journal's empty." };
                const lines = rows.slice(0, n).map(formatRow);
                return { content: "**Narelogs**\n" + lines.join("\n") };
            }
        },
        {
            name: "narelogs-clear",
            description: "Wipe the Narelogs journal",
            inputType: ApplicationCommandInputType.BUILT_IN,
            options: [],
            execute: async () => {
                journal = [];
                await DataStore.set(STORE, journal);
                return { content: "Nnaa. Journal cleared." };
            }
        }
    ],
    flux: {
        MESSAGE_CREATE: (e: MessageCreateEvent) => {
            const m = e.message;
            if (!m?.id) return;
            cacheSet(cacheKey(m.channel_id, m.id), m.content ?? "");
        },
        MESSAGE_DELETE: (e: MessageDeleteEvent) => {
            const { id, channelId } = e;
            if (!id) return;
            const content = MessageStore.getMessage(channelId, id)?.content
                ?? contentCache.get(cacheKey(channelId, id))
                ?? "";
            void remember({ id, kind: "deleted", content, channelId, at: Date.now() });
            if (settings.store.toastDeletes) showToast(settings.store.goneText || "Nnaa. This one was deleted.", Toasts.Type.MESSAGE);
            schedulePaint();
        },
        MESSAGE_UPDATE: (e: MessageUpdateEvent) => {
            const m = e.message;
            if (!m?.id) return;
            const key = cacheKey(m.channel_id, m.id);
            const before = contentCache.get(key) ?? "";
            const after = m.content ?? "";
            cacheSet(key, after);
            if (before === after) return;
            void remember({ id: m.id, kind: "edited", content: after, before: before || undefined, channelId: m.channel_id, at: Date.now() });
        },
        CHANNEL_SELECT: () => {
            watch();
            setTimeout(watch, CHANNEL_SWITCH_REWATCH_MS);
        }
    },
    onBeforeMessageSend(channelId, message: MessageObject) {
        if (!settings.store.persistSends || !message?.content) return;
        const withId = message as MessageObject & { id?: string; nonce?: string | number; };
        const rawId = withId.nonce ?? withId.id;
        const id = rawId != null ? String(rawId) : Date.now().toString(36) + Math.random().toString(36).slice(2, 8);
        void remember({ id, kind: "sent", content: message.content, channelId, at: Date.now() });
    },
    async start() {
        journal = (await DataStore.get(STORE)) ?? [];
        const before = journal.length;
        pruneOld();
        if (journal.length !== before) await DataStore.set(STORE, journal);
        sync();
        watch();
    },
    stop() {
        obs?.disconnect();
        obs = null;
        watchRoot = null;
        if (paintTimer) {
            clearTimeout(paintTimer);
            paintTimer = null;
        }
        contentCache.clear();
        const r = document.documentElement;
        r.classList.remove("narelogs-root", "narelogs-strobe", "narelogs-quake", "narelogs-boom", "narelogs-neon", "narelogs-ring", "narelogs-scan", "narelogs-glint", "narelogs-hide-box");
        r.style.removeProperty("--narelogs-gold");
        r.style.removeProperty("--narelogs-rose");
        r.style.removeProperty("--narelogs-cream");
        r.style.removeProperty("--narelogs-speed");
        document.querySelectorAll(".narelogs-gone").forEach(el => el.remove());
    }
});
