# Adobe Premiere Pro Scripting API Reference

Comprehensive reference for the ExtendScript, CEP, and standalone APIs used by the PremierPro MCP Server.

## 1. ExtendScript Object Model

```
app (Application)
  .project (Project)
    .activeSequence (Sequence)
      .videoTracks[] → Track → .clips[] → TrackItem → .components[] → Component → .properties[]
      .audioTracks[] → Track → .clips[] → TrackItem
      .markers → MarkerCollection
    .rootItem (ProjectItem) → .children[]
    .sequences[] → SequenceCollection
  .encoder (Encoder)
  .sourceMonitor (SourceMonitor)
```

## 2. Key Methods

### Project
- `app.project.createNewSequence(name, id)` — default settings only
- `app.project.importFiles(pathsArray, suppressUI, targetBin, asStills)` — batch import
- `app.project.save()` / `saveAs(path)`
- `app.enableQE()` — enables undocumented QE DOM

### Sequence
- `sequence.insertClip(projectItem, time, vTrack, aTrack)` — ripple insert
- `sequence.overwriteClip(projectItem, time, vTrack, aTrack)` — overwrite
- `sequence.importMGT(mogrtPath, timeTicks, vOffset, aOffset)` — text/titles via .mogrt
- `sequence.exportAsMediaDirect(outputPath, presetPath, workAreaType)` — sync export
- `sequence.getSettings()` / `setSettings(obj)` — frame rate, resolution, etc.

### Track
- `track.insertClip(projectItem, timeInSeconds)`
- `track.overwriteClip(projectItem, timeInSeconds)`
- `track.isMuted()` / `setMute(0|1)`

### TrackItem (Clip)
- `clip.start` / `end` / `inPoint` / `outPoint` — Time objects (R/W)
- `clip.components` — effects (index 0=Motion, 1=Opacity, 2+=applied effects)
- `clip.remove(inRipple, inAlignToVideo)`
- `clip.getMGTComponent()` — Motion Graphics Template params

### Audio Levels
```javascript
var volumeComp = audioClip.components[0]; // "Volume"
var levelParam = volumeComp.properties[0]; // "Level"
levelParam.setValue(0.5, true); // linear scale, not dB
```

### Encoder
- `app.encoder.encodeSequence(seq, outPath, presetPath, workArea, removeOnDone)`
- `app.encoder.startBatch()` — start AME render queue

## 3. QE DOM (Undocumented, Required for Transitions)

```javascript
app.enableQE();
var qeSeq = qe.project.getActiveSequence();
var qeClip = qeSeq.getVideoTrackAt(0).getItemAt(clipIndex);
qeClip.addTransition(qe.project.getVideoTransitionByName("Cross Dissolve"), true, duration);
```

- `qe.project.newSequence(name, presetPath)` — create sequence with specific settings

## 4. Key Limitations

| Limitation | Workaround |
|---|---|
| Transitions — no standard DOM method | QE DOM (`qe.project.getVideoTransitionByName()`) |
| Text/Titles — no create-from-scratch API | Motion Graphics Templates (`.mogrt`) via `sequence.importMGT()` |
| Effects — no apply-by-name in standard DOM | QE DOM or modify existing component params |
| Sequence creation with settings | QE DOM `newSequence()` with `.sqpreset` file |
| No undo grouping | N/A — each operation is a separate undo step |
| Sync execution blocks UI | Split operations, use `$.sleep()` |

## 5. Time System

- **Ticks**: 254,016,000,000 ticks = 1 second
- `new Time(); t.seconds = 5.0;`
- `time.ticks` — string, `time.seconds` — number

## 6. CEP Panel Communication

```javascript
// Panel → Premiere Pro
csInterface.evalScript('hostFunction()', function(result) { ... });

// Node.js in CEP (prefix with cep_node)
const ws = cep_node.require('ws');

// Event system
csInterface.addEventListener("eventType", callback);
```

## 7. Installation Paths

| Platform | Path |
|---|---|
| macOS (user) | `~/Library/Application Support/Adobe/CEP/extensions/` |
| macOS (system) | `/Library/Application Support/Adobe/CEP/extensions/` |
| Windows (user) | `%APPDATA%/Adobe/CEP/extensions/` |

Enable unsigned extensions: `defaults write com.adobe.CSXS.11 PlayerDebugMode 1`

## 8. Standalone Control (macOS)

```bash
osascript -e 'tell application "Adobe Premiere Pro 2025" to do script "app.project.name"'
```

- macOS only, higher latency, string-only returns
- CEP panel is preferred; standalone is fallback

## 9. Common Transition Names

Cross Dissolve, Dip to Black, Dip to White, Film Dissolve, Wipe, Barn Doors, Push, Slide, Morph Cut, Constant Power (audio), Constant Gain (audio)

## 10. Future: UXP

- ExtendScript/CEP supported through September 2026
- UXP is the future but currently beta-only for Premiere Pro
- CEP + ExtendScript is the correct choice for now
