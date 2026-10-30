import { defineConfig } from "mygo-cli";

export default defineConfig({
  name: "cuteimg",
  identifier: "io.github.rafalfilipek.cuteimg",
  version: "0.3.2",
  // Signed updates from this repository's GitHub releases. The secret key
  // that signs them is the MYGO_UPDATER_PRIVATE_KEY secret of the repository.
  updates: {
    publicKey: "+gGq209bSdf04cxnVKJ0EgQKWv2UMt7dhbqa0DHeNKs=",
    github: "RafalFilipek/cuteimg",
  },
  // `mygo dev` runs devCommand and loads devUrl; `mygo build` runs
  // buildCommand and embeds frontendDist into the app.
  devUrl: "http://localhost:5173",
  devCommand: "bun run dev:web",
  buildCommand: "bun run build:web",
  frontendDist: "dist",
  bindings: "src/mygo.ts",
  out: "build",
});
