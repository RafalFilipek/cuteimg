// Connects the editor in index.html to Go: copying and saving go through the
// system, screen captures from the global shortcut arrive as the captured
// event, and the shortcut can be recorded and tested in a settings dialog.
import { currentWindow } from "mygo-runtime";
import { Editor, events } from "./mygo";

declare global {
  interface Window {
    cuteimg: { loadBlob(blob: Blob): Promise<void> };
    cuteimgNative?: { copy(png: Blob): Promise<void>; save(png: Blob, name: string): Promise<string> };
  }
}

const $ = <T extends HTMLElement>(selector: string) => document.querySelector<T>(selector)!;

async function toBase64(blob: Blob): Promise<string> {
  const bytes = new Uint8Array(await blob.arrayBuffer());
  let bin = "";
  for (let i = 0; i < bytes.length; i += 0x8000) bin += String.fromCharCode(...bytes.subarray(i, i + 0x8000));
  return btoa(bin);
}

function fromBase64(b64: string): Blob {
  const bin = atob(b64);
  const bytes = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
  return new Blob([bytes], { type: "image/png" });
}

window.cuteimgNative = {
  copy: async (png) => Editor.copy(await toBase64(png)),
  save: async (png, name) => Editor.save(await toBase64(png), name),
};

events.captured.on((png) => {
  if (png) window.cuteimg.loadBlob(fromBase64(png));
});

// ---------- title bar ----------
// The window is frameless: the toolbar drags it, and these buttons replace
// the system's.

const chrome = document.createElement("style");
chrome.textContent = `
  body .bar { --app-region: drag; user-select: none; -webkit-user-select: none; }
  body .bar input, .bar button { --app-region: no-drag; }
  body .winctl { display: flex; align-items: center; gap: 2px; margin-left: 4px; padding-left: 10px; border-left: 1px solid var(--line); }
  body .winctl button { width: 34px; height: 30px; border: 0; border-radius: 8px; background: transparent; color: var(--muted); display: grid; place-items: center; cursor: pointer; }
  body .winctl button:hover { background: var(--hover); color: var(--fg); }
  body .winctl button.close:hover { background: #e5484d; color: #fff; }
  body .winctl svg { width: 14px; height: 14px; }
  body .bar { flex-wrap: nowrap; }
  @media (max-width: 1420px) { body .bar .tbtn kbd { display: none; } }
  @media (max-width: 1260px) { body .bar .tbtn:not(.primary) .lbl { display: none; } }
  @media (max-width: 1080px) { body .bar .brand { font-size: 0; gap: 0; } .bar .primary .lbl { display: none; } }
`;
document.head.append(chrome);

const winctl = document.createElement("div");
winctl.className = "winctl";
winctl.innerHTML = `
  <button id="winMin" title="Minimize" aria-label="Minimize"><svg viewBox="0 0 14 14"><path d="M2.5 7h9" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/></svg></button>
  <button id="winMax" title="Maximize" aria-label="Maximize"><svg viewBox="0 0 14 14"><rect x="2.5" y="2.5" width="9" height="9" rx="2.2" fill="none" stroke="currentColor" stroke-width="1.4"/></svg></button>
  <button id="winClose" class="close" title="Close to tray" aria-label="Close"><svg viewBox="0 0 14 14"><path d="M3 3l8 8M11 3l-8 8" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/></svg></button>
`;
$(".bar").append(winctl);
$("#winMin").onclick = () => currentWindow.minimize();
$("#winMax").onclick = () => currentWindow.toggleMaximize();
$("#winClose").onclick = () => currentWindow.close();

// ---------- toolbar ----------

const icon = (d: string) =>
  `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">${d}</svg>`;

const captureBtn = document.createElement("button");
captureBtn.className = "tbtn";
captureBtn.innerHTML = icon('<path d="M4 9V5h4M16 5h4v4M20 15v4h-4M8 19H4v-4"/>') + '<span class="lbl">Capture</span>';
captureBtn.onclick = () => Editor.capture();

const settingsBtn = document.createElement("button");
settingsBtn.className = "tbtn";
settingsBtn.title = "Capture shortcut";
settingsBtn.setAttribute("aria-label", "Capture shortcut settings");
settingsBtn.innerHTML = icon(
  '<rect x="2" y="6" width="20" height="12" rx="2"/><path d="M6 10h.01M10 10h.01M14 10h.01M18 10h.01M8 14h8"/>',
);
settingsBtn.onclick = () => openDialog();

$("#open").before(captureBtn, settingsBtn);

// The capture shortcut leads the list of keys in the side panel.
const captureRow = document.createElement("div");
captureRow.className = "featured";
$("#keys").prepend(captureRow);

function showShortcut(sc: string) {
  captureBtn.title = sc ? `Capture area (${sc})` : "Capture area";
  captureRow.innerHTML = `<dt>Capture area</dt><dd>${sc ? keys(sc) : '<span class="hold">not set</span>'}</dd>`;
  const tag = document.querySelector("#sampleTag");
  if (tag) tag.innerHTML = `Example. <b>${sc ? sc + " captures the screen, " : ""}Ctrl+V pastes</b>`;
}

const keys = (sc: string) =>
  sc
    .split("+")
    .map((k) => `<kbd>${k}</kbd>`)
    .join("");

const chips = (sc: string) =>
  sc
    .split("+")
    .map((k) => `<kbd>${k}</kbd>`)
    .join("+");

// ---------- shortcut dialog ----------

const style = document.createElement("style");
style.textContent = `
  #scDialog { border: 1px solid var(--line); border-radius: 14px; background: var(--panel); color: var(--fg); padding: 22px; width: min(420px, calc(100vw - 32px)); box-shadow: 0 24px 60px rgba(0,0,0,.25); }
  #scDialog::backdrop { background: rgba(10,12,18,.45); }
  #scDialog h3 { margin: 0 0 4px; font-size: 17px; }
  #scDialog p { margin: 0; color: var(--muted); font-size: 13px; line-height: 1.5; }
  .sc-box { margin: 16px 0 12px; min-height: 64px; border: 1.5px dashed var(--line); border-radius: 12px; display: grid; place-items: center; font-size: 15px; padding: 10px; text-align: center; }
  .sc-box.rec { border-color: var(--accent); }
  .sc-box kbd { font-family: var(--font-num); font-size: 14px; background: var(--hover); border: 1px solid var(--line); border-bottom-width: 3px; border-radius: 6px; padding: 3px 9px; margin: 0 3px; color: var(--fg); }
  .sc-actions { display: flex; gap: 8px; flex-wrap: wrap; }
  .sc-actions .tbtn, .sc-foot .tbtn { border: 1px solid var(--line); padding: 0 12px; }
  .sc-status { min-height: 20px; margin-top: 12px; font-size: 13px; line-height: 1.5; }
  .sc-status.ok { color: #1f9d55; } .sc-status.err { color: #d93a2b; } .sc-status.wait { color: var(--muted); }
  .sc-foot { display: flex; justify-content: flex-end; margin-top: 16px; }
`;
document.head.append(style);

const dialog = document.createElement("dialog");
dialog.id = "scDialog";
dialog.setAttribute("aria-labelledby", "scTitle");
dialog.innerHTML = `
  <h3 id="scTitle">Capture shortcut</h3>
  <p>Press it anywhere to capture an area of the screen. Record a new combination, then test that it reaches cuteimg.</p>
  <div class="sc-box" id="scBox"></div>
  <div class="sc-actions">
    <button class="tbtn primary" id="scRecord">Record new</button>
    <button class="tbtn" id="scTest">Test</button>
    <button class="tbtn" id="scReset">Reset to default</button>
  </div>
  <div class="sc-status" id="scStatus" role="status" aria-live="polite"></div>
  <div class="sc-foot"><button class="tbtn" id="scDone">Done</button></div>
`;
document.body.append(dialog);

const box = $("#scBox"), status = $("#scStatus");
const isWindows = navigator.userAgent.includes("Windows");
const isMac = navigator.userAgent.includes("Mac");
let current = "";
let recording = false;
let capsHeld = false;
let testTimer = 0;

function setStatus(text: string, kind: "ok" | "err" | "wait" | "" = "") {
  status.textContent = text;
  status.className = "sc-status " + kind;
}

function renderBox(live = "") {
  box.classList.toggle("rec", recording);
  if (recording) box.innerHTML = live ? chips(live) + "<span>+…</span>" : "Press the keys you want…";
  else box.innerHTML = current ? chips(current) : "<span>No shortcut</span>";
}

function modifiers(e: KeyboardEvent): string[] {
  const m: string[] = [];
  if (capsHeld && isWindows) m.push("CapsLock");
  if (e.ctrlKey) m.push("Ctrl");
  if (e.altKey) m.push("Alt");
  if (e.shiftKey) m.push("Shift");
  if (e.metaKey) m.push(isMac ? "Cmd" : "Super");
  return m;
}

function keyName(e: KeyboardEvent): string {
  const c = e.code;
  if (/^Key[A-Z]$/.test(c)) return c.slice(3);
  if (/^Digit\d$/.test(c)) return c.slice(5);
  if (/^F\d{1,2}$/.test(c)) return c;
  if (c === "Space") return "Space";
  return "";
}

async function startRecording() {
  stopTest();
  recording = true;
  capsHeld = false;
  await Editor.pauseShortcut(true);
  setStatus(isWindows ? "Hold Ctrl, Alt, Win or CapsLock and press a key. A Raycast hyper key works too. Esc cancels." : "Hold Ctrl, Alt or Cmd and press a key. Esc cancels.", "wait");
  renderBox();
}

async function stopRecording() {
  recording = false;
  capsHeld = false;
  await Editor.pauseShortcut(false);
  renderBox();
}

async function apply(sc: string) {
  try {
    await Editor.setShortcut(sc);
    current = sc;
    showShortcut(sc);
    renderBox();
    setStatus(`Saved. Press Test to check that ${sc} works.`, "ok");
  } catch (err) {
    renderBox();
    setStatus(String((err as Error)?.message ?? err), "err");
  }
}

window.addEventListener(
  "keydown",
  async (e) => {
    if (!dialog.open) return;
    e.stopImmediatePropagation(); // keep the editor's own shortcuts out of the dialog
    if (!recording) return;
    e.preventDefault();
    if (e.key === "Escape") {
      await stopRecording();
      setStatus("");
      return;
    }
    if (e.key === "CapsLock") capsHeld = true;
    const mods = modifiers(e);
    const key = keyName(e);
    if (!key) {
      renderBox(mods.join("+"));
      return;
    }
    const sc = [...mods, key].join("+");
    recording = false;
    capsHeld = false;
    await apply(sc); // registering the new shortcut also ends the pause
  },
  true,
);
window.addEventListener(
  "keyup",
  (e) => {
    if (!dialog.open) return;
    e.stopImmediatePropagation();
    if (e.key === "CapsLock") capsHeld = false;
    if (recording) renderBox(modifiersFromState(e));
  },
  true,
);
const modifiersFromState = (e: KeyboardEvent) => modifiers(e).join("+");

function stopTest() {
  if (!testTimer) return;
  clearTimeout(testTimer);
  testTimer = 0;
  Editor.testShortcut(false);
}

async function startTest() {
  if (recording) await stopRecording();
  if (!current) {
    setStatus("Record a shortcut first.", "err");
    return;
  }
  stopTest();
  await Editor.testShortcut(true);
  setStatus(`Press ${current} now…`, "wait");
  testTimer = window.setTimeout(() => {
    stopTest();
    setStatus(`Nothing arrived. Another app may catch ${current} first. Try another combination.`, "err");
  }, 15000);
}

events.shortcutPressed.on((sc) => {
  if (!testTimer) return;
  stopTest();
  setStatus(`It works. cuteimg received ${sc}.`, "ok");
});

$("#scRecord").onclick = startRecording;
$("#scTest").onclick = startTest;
$("#scReset").onclick = async () => {
  if (recording) {
    recording = false;
    capsHeld = false;
  }
  await apply(await Editor.defaultShortcut());
};
$("#scDone").onclick = () => dialog.close();
dialog.addEventListener("cancel", (e) => {
  if (recording) e.preventDefault(); // Esc ends the recording, not the dialog
});
dialog.addEventListener("close", async () => {
  stopTest();
  if (recording) await stopRecording();
  setStatus("");
});

function openDialog() {
  renderBox();
  setStatus(current ? "" : "No shortcut is active. Record one.", current ? "" : "err");
  dialog.showModal();
}

current = await Editor.shortcut();
showShortcut(current);
