# Native transition recipes

The `standard` and `transitions` MCP profiles include `premiere_list_transition_recipes` and `premiere_apply_transition_recipe`. The recipes use exact installed native Premiere names and original duration suggestions. They do not install or redistribute vendor preset files.

## Library

| Recipe ID | Native transition | Default frames | Use |
| --- | --- | ---: | --- |
| cross_dissolve | Cross Dissolve | 12 | Clean continuity |
| blur_dissolve | Blur Dissolve | 12 | Soft change between detailed shots |
| push | Push | 8 | Navigation between views |
| slide | Slide | 8 | Reveal the next screen |
| whip | Whip | 6 | Fast movement |
| zoom_blur | Zoom Blur | 6 | A short change in focus |
| film_dissolve | Film Dissolve | 12 | Gentle scene change |
| luma_fade | Luma Fade | 12 | Brightness-led reveal |
| light_leak | Light Leak | 12 | A brief light wash |
| glitch | Glitch | 6 | Digital accent |
| flash | Flash | 6 | Beat or launch accent |
| directional_blur | Directional Blur | 6 | Carry movement through a cut |

Durations are frames at the active sequence's actual frame rate: 6 frames is 0.25 seconds at 24 fps, or 0.2 seconds at 30 fps. Omit `duration_frames` to use the recipe default.

## Usage

1. Inspect the intended project and active sequence. List recipes; `available=true` means the exact native video name was discovered, not that an application or render was tested. A disconnected or malformed catalog reports unknown availability.
2. Plan the transition at the outgoing cut of `clip_index`. Dry run is the default and creates no sequence copy:

   ```json
   {"recipe_id":"whip","track_index":0,"clip_index":0}
   ```

3. Inspect the plan's cut, duration, source sequence identity, and proven media handles. Handles are unused source frames outside the clip trims. The conservative preflight requires a full requested duration after the outgoing source out point and before the incoming source in point; it does not fabricate extra frames.
4. Apply explicitly:

   ```json
   {"recipe_id":"whip","track_index":0,"clip_index":0,"dry_run":false}
   ```

   This makes and activates a distinct sequence copy by default. The source remains the recovery sequence. Set `duplicate_sequence=false` only when intentionally editing the active sequence directly.
5. Treat the operation as successful only when `status=applied` and `verified=true`. Inspect `sourceSequence`, `appliedSequence`, `requested`, and `actual`, then read the transitions back or preview/export the copy.

## Safety and limits

- The two clips must form an adjacent video cut, use normal playback speed, and have proven handles. Locked tracks, nested/offline sources, missing bounds, existing transitions at the selected cut, or ambiguous QE clip mappings fail before application.
- QE collections contain gaps and transitions as well as clips. The recipe maps the outgoing DOM clip by unique name and timeline bounds, not by assuming the DOM clip index equals the QE item index.
- Success requires exactly one added transition, exact name equality, centered cut placement, duration within one frame, and unchanged existing structural timeline content.
- Structural comparison covers video/audio track counts, clip/source identities, trims, positions, speeds, and transition names/times. It does not certify component parameters, keyframes, or sequence markers as a complete backup.
- Errors remain MCP errors; recovery identities and mutation diagnostics identify a copy that may need inspection. Never retry blindly after a post-application verification failure.
- The public `premiere_add_video_transition` flag `apply_to_end=true` now targets the clip end; QE's internal boolean has the opposite meaning. `false` targets the start.
- Direction, intensity, color, synthesized animation stacks, and third-party template import are outside this recipe tool's scope.

## Native acceptance evidence

[Recorded native results](verification/native-transition-recipes-2026-10-01.json) cover all 12 names at a 12-frame duration and all 12 recipes at their 6-, 8-, or 12-frame defaults, including dry-run and copy verification, exact transition readback, H.264 export and complete decoding. Five live MCP argument guards, one native occupied-cut guard, and both clip-edge flags also passed. Tests used dedicated fixture footage and did not edit a production project. Full raw reports and rendered previews were retained locally; the committed manifest records their hashes and per-recipe readbacks.

The acceptance environment is Premiere 26.5.2, macOS 27.0.1 arm64, en_US, 1280×720, 24 fps. Other versions, platforms, frame rates, vertical formats, and performance remain unverified. Export decoding is not a perceptual-quality certificate.

A fresh MCP connection exposes 74 tools in `standard`. Clients caching the old catalog must reconnect the MCP server to discover the two new recipe tools.
