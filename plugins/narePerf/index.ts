/*
 * SPDX-License-Identifier: GPL-3.0
 * Narecord, a Nanachi-themed Equicord/Vencord fork
 * Copyright (c) 2026 Narehatechi and Narecord contributors
 *
 * narePerf — first-party den userplugin.
 *
 * This repo is the Narelotl installer. Stock installs download Equicord's
 * desktop.asar, which cannot load a new plugin until Equicord is rebuilt with
 * this folder in src/userplugins. Narelotl still ships the source, enables
 * narePerf in settings.json, and injects the same default-ON CSS via QuickCSS
 * so fluff-kill works immediately.
 *
 * Enable: Narecord Settings → Plugins → narePerf (Narecord Plugin card).
 * Source-build: copy this folder to Equicord's src/userplugins/narePerf and
 * run `pnpm build`, then restart Discord.
 */

import { isPluginEnabled } from "@api/PluginManager";
import { definePluginSettings, Settings } from "@api/Settings";
import { Logger } from "@utils/Logger";
import { isObject } from "@utils/misc";
import definePlugin, { OptionType } from "@utils/types";
import { findAll } from "@webpack";

const log = new Logger("narePerf");
const STYLE_ID = "nr-nareperf-style";
const ROOT_CLASS = "nr-perf-plugin";

const Narehatechi = { name: "Narehatechi", id: 0n };

interface SpringModule {
    Globals: { assign(options: { skipAnimation: boolean; }): void; };
    Springs: object;
}

let springs: SpringModule[] = [];
let started = false;
let savedWindowsMaterial: string | undefined;
let savedMacosVibrancy: string | undefined;
let savedMaterials = false;

const isSpringGlobals = (v: unknown): v is SpringModule["Globals"] =>
    isObject(v) && "assign" in v && typeof (v as { assign?: unknown; }).assign === "function";

const isSpringModule = (v: unknown): v is SpringModule => {
    if (!isObject(v)) return false;
    const m = v as Partial<SpringModule>;
    return isSpringGlobals(m.Globals) && isObject(m.Springs);
};

function loadSprings() {
    try {
        springs = findAll(isSpringModule);
    } catch (err) {
        log.warn("spring modules not found", err);
        springs = [];
    }
}

function applySpringSkip(skip: boolean) {
    for (const s of springs) {
        try {
            s.Globals.assign({ skipAnimation: skip });
        } catch (err) {
            log.warn("spring skip failed", err);
        }
    }
}

function applyVibrancy(kill: boolean) {
    try {
        if (!savedMaterials) {
            savedWindowsMaterial = Settings.windowsMaterial;
            savedMacosVibrancy = Settings.macosVibrancyStyle as string | undefined;
            savedMaterials = true;
        }
        if (kill) {
            Settings.windowsMaterial = "none";
            Settings.macosVibrancyStyle = undefined;
            return;
        }
        if (savedWindowsMaterial !== undefined)
            Settings.windowsMaterial = savedWindowsMaterial as typeof Settings.windowsMaterial;
        Settings.macosVibrancyStyle = savedMacosVibrancy as typeof Settings.macosVibrancyStyle;
    } catch (err) {
        log.warn("vibrancy setting skipped", err);
    }
}

function featureClass(name: string, on: boolean) {
    document.documentElement.classList.toggle(name, on);
}

function syncRootClasses() {
    const root = document.documentElement;
    root.classList.add(ROOT_CLASS);
    featureClass("nr-perf-kill-blur", settings.store.killBlur);
    featureClass("nr-perf-kill-motion", settings.store.killAnimations);
    featureClass("nr-perf-kill-acrylic", settings.store.killAcrylic);
    featureClass("nr-perf-kill-decor", settings.store.killDecorations);
    featureClass("nr-perf-quiet-gifs", settings.store.quietGifs);
}

function pluginCss(): string {
    return `
/* narePerf — fluff first. Do not restyle message text. */
html.nr-perf-kill-blur [class*="backdrop"],
html.nr-perf-kill-blur [class*="layer"] [class*="focusLock"],
html.nr-perf-kill-blur [class*="modal"],
html.nr-perf-kill-blur [class*="popout"],
html.nr-perf-kill-blur [class*="menu"],
html.nr-perf-kill-blur [style*="backdrop-filter"] {
    backdrop-filter: none !important;
    -webkit-backdrop-filter: none !important;
}
html.nr-perf-kill-blur [class*="backdrop"] {
    background-color: rgb(0 0 0 / 48%) !important;
}

html.nr-perf-kill-motion,
html.nr-perf-kill-motion *:not([class*="messageContent"]):not([class*="markup"]),
html.nr-perf-kill-motion *:not([class*="messageContent"]):not([class*="markup"])::before,
html.nr-perf-kill-motion *:not([class*="messageContent"]):not([class*="markup"])::after {
    animation-duration: 0.001ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.001ms !important;
    scroll-behavior: auto !important;
}

html.nr-perf-kill-acrylic [class*="acrylic"],
html.nr-perf-kill-acrylic [class*="Acrylic"] {
    background-color: var(--background-secondary, #2b2d31) !important;
    backdrop-filter: none !important;
    -webkit-backdrop-filter: none !important;
}

html.nr-perf-kill-decor [class*="avatarDecoration"],
html.nr-perf-kill-decor [class*="AvatarDecoration"],
html.nr-perf-kill-decor [class*="profileEffect"],
html.nr-perf-kill-decor [class*="ProfileEffect"],
html.nr-perf-kill-decor [class*="confetti"],
html.nr-perf-kill-decor [class*="Confetti"],
html.nr-perf-kill-decor [class*="giftAnimation"],
html.nr-perf-kill-decor [class*="GiftAnimation"],
html.nr-perf-kill-decor [class*="hoverParticle"],
html.nr-perf-kill-decor [class*="HoverParticle"] {
    display: none !important;
    animation: none !important;
    visibility: hidden !important;
}

html.nr-perf-quiet-gifs img[class*="emoji"][src*="gif"],
html.nr-perf-quiet-gifs [class*="lottieIcon"],
html.nr-perf-quiet-gifs [class*="sticker"][class*="lottie"] {
    animation: none !important;
}
`.trim();
}

function injectCss() {
    let el = document.getElementById(STYLE_ID) as HTMLStyleElement | null;
    if (!el) {
        el = document.createElement("style");
        el.id = STYLE_ID;
        document.documentElement.appendChild(el);
    }
    el.textContent = pluginCss();
    syncRootClasses();
}

function removeCss() {
    document.getElementById(STYLE_ID)?.remove();
    const root = document.documentElement;
    root.classList.remove(
        ROOT_CLASS,
        "nr-perf-kill-blur",
        "nr-perf-kill-motion",
        "nr-perf-kill-acrylic",
        "nr-perf-kill-decor",
        "nr-perf-quiet-gifs",
    );
}

function refreshLive() {
    if (!started) return;
    injectCss();
    if (settings.store.killAnimations) {
        if (springs.length === 0) loadSprings();
        applySpringSkip(settings.store.skipSprings);
    } else {
        applySpringSkip(false);
    }
    applyVibrancy(settings.store.killAcrylic);
}

const settings = definePluginSettings({
    killBlur: {
        type: OptionType.BOOLEAN,
        description: "Kill backdrop-filter / heavy blur on layers and modals. Replaces blur with a solid dim so chat contrast stays readable.",
        default: true,
        onChange() { refreshLive(); }
    },
    killAnimations: {
        type: OptionType.BOOLEAN,
        description: "Disable UI animations and transitions (prefers-reduced-motion style). Message text spacing and contrast are left alone.",
        default: true,
        onChange() { refreshLive(); }
    },
    killAcrylic: {
        type: OptionType.BOOLEAN,
        description: "Strip window vibrancy / mica / acrylic (Equicord windowsMaterial + CSS). Discord-native acrylic only where Equicord exposes it.",
        default: true,
        onChange() { refreshLive(); }
    },
    killDecorations: {
        type: OptionType.BOOLEAN,
        description: "Hide nitro confetti, gift animations, hover particles, avatar decorations, and profile effects when those nodes exist.",
        default: true,
        onChange() { refreshLive(); }
    },
    quietTyping: {
        type: OptionType.BOOLEAN,
        description: "Disable the CPU-heavy typing-dot animation. The typing indicator still appears. Uses Equicord's NoTypingAnimation patch if that plugin is on.",
        default: true,
        restartNeeded: true
    },
    quietGifs: {
        type: OptionType.BOOLEAN,
        description: "Lower GIF / sticker / emoji animation aggressiveness (canAnimate inverse of AlwaysAnimate, plus CSS). Hover is not required for chat readability.",
        default: true,
        restartNeeded: true,
        onChange() { refreshLive(); }
    },
    skipSprings: {
        type: OptionType.BOOLEAN,
        description: "Skip Discord spring physics (known Vencord/FastDiscord pattern). Does not throttle chat message list re-renders.",
        default: true,
        onChange() { refreshLive(); }
    },
});

export default definePlugin({
    name: "narePerf",
    description: "Quiet den: strips Discord fluff so the hideout stays snappy. Nnaa~ blur, motion, and acrylic first; chat text stays readable.",
    authors: [Narehatechi],
    tags: ["Utility", "Appearance"],
    searchTerms: ["performance", "blur", "animation", "acrylic", "vibrancy", "fps", "lag"],
    enabledByDefault: true,
    settings,

    patches: [
        {
            find: "dotCycle",
            predicate: () => settings.store.quietTyping && !isPluginEnabled("NoTypingAnimation"),
            replacement: {
                match: /focused:(\i)/g,
                replace: (_, focused) => `_focused:${focused}=false`
            }
        },
        {
            find: "canAnimate:",
            all: true,
            noWarn: true,
            predicate: () => settings.store.quietGifs,
            replacement: {
                match: /canAnimate:.+?([,}].*?\))/g,
                replace: (m, rest) => {
                    const destructuringMatch = rest.match(/}=.+/);
                    if (destructuringMatch == null) return `canAnimate:!1${rest}`;
                    return m;
                }
            }
        }
    ],

    start() {
        started = true;
        injectCss();
        if (settings.store.killAnimations && settings.store.skipSprings) {
            loadSprings();
            applySpringSkip(true);
        }
        if (settings.store.killAcrylic) applyVibrancy(true);
        log.info("narePerf started (fluff-kill first; chat left readable)");
    },

    stop() {
        started = false;
        applySpringSkip(false);
        springs = [];
        applyVibrancy(false);
        removeCss();
        log.info("narePerf stopped");
    }
});
