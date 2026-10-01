# UXP migration and inspection prototype

Status: partial implementation of [issue #8](https://github.com/ayushozha/AdobePremiereProMCP/issues/8), checked 2026-09-30. The production MCP bridge still uses CEP or standalone ExtendScript. The UXP prototype is separate and has not been loaded in Premiere during this validation run.

## Current Adobe support

Adobe documents UXP support in Premiere 25.6. It is no longer accurate to describe all Premiere UXP support as beta-only, or to divide CEP and UXP support solely by calendar year. Use the installed version and available APIs to select a bridge.

Adobe's [September 24, 2026 transition announcement](https://blog.developer.adobe.com/en/publish/2026/09/investing-in-the-future-of-creative-cloud-extensibility-uxp-comes-to-our-flagship-applications) specifies that Premiere stops accepting new CEP Marketplace submissions in December 2027, disables CEP by default in December 2028, and removes CEP from new flagship-app versions starting December 2029. ExtendScript itself is unaffected. These are different milestones; issue #8's original approximate 2027 deprecation claim should not be used as a removal deadline. Future Adobe changes should be checked before release planning.

- [Adobe introduction and host support](https://developer.adobe.com/premiere-pro/uxp/introduction/)
- [Manifest requirements](https://developer.adobe.com/premiere-pro/uxp/plugins/concepts/manifest/)
- [Project API](https://developer.adobe.com/premiere-pro/uxp/ppro-reference/classes/project)
- [Sequence API](https://developer.adobe.com/premiere-pro/uxp/ppro-reference/classes/sequence)
- [Host version information](https://developer.adobe.com/premiere-pro/uxp/uxp-api/reference-js/modules/uxp/host-information/host)

## Prototype included here

`uxp-panel/` is a development panel with an **Inspect project** button. It reads the host version, active project, sequence names, and video/audio/caption track counts through documented UXP methods. Missing methods and failed reads produce unknown values and warnings; they never become zero counts. The panel requests no network or filesystem permission.

To evaluate it in Premiere 25.6 or newer, add `uxp-panel/manifest.json` in UXP Developer Tool 2.2 or newer, load it into Premiere, and open the **MCP UXP Inspection** panel. Inspect a disposable project with known video, audio, and caption tracks. Compare its output with the timeline. Record the Premiere version, OS, counts, warnings, and an application screenshot. The development plugin ID is not a Marketplace registration.

The offline check is:

```sh
node --test uxp-panel/test/inspect.test.js
```

These tests use host fixtures. They validate reporting and error behavior; they do not establish application loading, Adobe API compatibility, or MCP connectivity.

## Bridge design

| Concern | Current CEP implementation | UXP work required |
|---|---|---|
| Host execution | ExtendScript string dispatch | Explicit handlers calling the asynchronous `premierepro` API |
| Transport | CEP hosts the WebSocket server | Node must host it; UXP initiates the connection |
| Authentication | Shared bearer token on handshake | Pair the plugin with a local authenticated transport; never put tokens in URL query strings |
| Mutations and undo | DOM/QE calls | Construct supported actions and execute undoable project transactions |
| Capability gaps | Large ExtendScript command catalog | Advertise only implemented, verified UXP commands; reject others explicitly |
| Selection | `BRIDGE_MODE=cep` or `standalone` | Add `uxp` only when the adapter and transport are implemented and tested |

Adobe's [network recipe](https://developer.adobe.com/premiere-pro/uxp/resources/recipes/network/) specifies client-only WebSockets. Reusing the CEP server inside a UXP plugin will not work. The authenticated transport must also be validated against host-specific network permissions and macOS restrictions before choosing loopback HTTP/WS versus TLS. No network listener is included in this inspection prototype.

## Remaining acceptance criteria for #8

1. Load the panel in Premiere and record native readback evidence for project, sequence, and caption inspection.
2. Implement authenticated Node-to-UXP request correlation, disconnect handling, timeouts, and duplicate-request handling.
3. Add the UXP bridge adapter and capability filtering, retaining CEP as an explicit option.
4. Validate import, insertion, transitions, effects, captions, export, and undo behavior on disposable projects. Unsupported APIs must remain explicit failures.
5. Run the live validation checklist through both adapters on named Premiere/OS versions. A manifest minimum version alone is not a compatibility claim.

Issue #8 remains open until these criteria are met.
