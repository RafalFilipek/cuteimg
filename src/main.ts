// Connects the editor in index.html to Go: copying and saving go through the
// system, and screen captures from the global shortcut arrive as the
// captured event.
import { Editor, events } from "./mygo";

declare global {
  interface Window {
    cuteimg: { loadBlob(blob: Blob): Promise<void> };
    cuteimgNative?: { copy(png: Blob): Promise<void>; save(png: Blob, name: string): Promise<string> };
  }
}

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

const shortcut = await Editor.shortcut();

// A "Capture" button next to Open, the same as the global shortcut.
const open = document.querySelector("#open");
if (open) {
  const btn = document.createElement("button");
  btn.className = "tbtn";
  btn.title = `Capture area (${shortcut})`;
  btn.innerHTML =
    '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M4 9V5h4M16 5h4v4M20 15v4h-4M8 19H4v-4"/></svg><span class="lbl">Capture</span>';
  btn.onclick = () => Editor.capture();
  open.before(btn);
}

const tag = document.querySelector("#sampleTag");
if (tag) tag.innerHTML = `Example. <b>${shortcut} captures the screen, Ctrl+V pastes</b>`;
const hint = document.querySelector(".hint");
if (hint) hint.firstElementChild?.parentElement?.insertAdjacentHTML("afterbegin", `<kbd>${shortcut}</kbd> captures an area of the screen from anywhere.<br>`);
