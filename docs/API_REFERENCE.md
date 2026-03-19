# PremierPro MCP -- API Reference

Complete reference for the 50 most-used MCP tools organized by category. Each tool is callable via the MCP protocol through the Go orchestrator.

---

## Table of Contents

- [Project Management](#project-management)
- [Sequence](#sequence)
- [Clips](#clips)
- [Effects & Color](#effects--color)
- [Audio & Export](#audio--export)

---

## Project Management

### `premiere_open`

Launch Adobe Premiere Pro on macOS. If Premiere is already running and a project_path is provided, the project is opened in the existing instance. When wait is true (default), the call blocks until Premiere finishes launching (up to 60 seconds).

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `project_path` | string | No | — | Absolute path to a .prproj file to open on launch (e.g. `/Users/me/Projects/MyEdit.prproj`). If omitted, Premiere opens with no project. |
| `wait` | boolean | No | `true` | If true, block until Premiere Pro is confirmed running or 60 seconds elapse. Set to false for fire-and-forget launch. |

**Returns:**
```json
{
  "status": "launched",
  "message": "Adobe Premiere Pro launched successfully (took ~6s).",
  "project": "/Users/me/Projects/MyEdit.prproj"
}
```

**Example:**
```
"Launch Premiere and open my project"
-> premiere_open { project_path: "/Users/me/Projects/MyEdit.prproj" }
```

---

### `premiere_close`

Quit Adobe Premiere Pro. By default sends a graceful quit via AppleScript, which may trigger a save dialog. Use force=true to kill the process immediately without saving.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `force` | boolean | No | `false` | If true, force-kill the process (SIGKILL) without saving. If false, send a graceful quit that allows Premiere to prompt for unsaved changes. |

**Returns:**
```json
{ "status": "closed", "message": "Adobe Premiere Pro has been closed." }
```

**Example:**
```
"Force quit Premiere"
-> premiere_close { force: true }
```

---

### `premiere_is_running`

Check whether Adobe Premiere Pro is currently running as a macOS process. Returns running status and, when running, the process IDs.

**Parameters:**

_None._

**Returns:**
```json
{ "running": true, "pids": ["12345"] }
```

**Example:**
```
"Is Premiere running?"
-> premiere_is_running {}
```

---

### `premiere_open_project`

Open an existing Premiere Pro project file (.prproj). Closes any currently open project. If there are unsaved changes, Premiere may prompt to save.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `path` | string | Yes | — | Absolute path to the .prproj file to open (e.g. `/Users/me/Projects/MyEdit.prproj`). The file must exist. |

**Returns:**
```json
{ "status": "ok", "project_name": "MyEdit", "path": "/Users/me/Projects/MyEdit.prproj" }
```

**Example:**
```
"Open my interview project"
-> premiere_open_project { path: "/Users/me/Projects/Interview.prproj" }
```

---

### `premiere_save_project`

Save the currently open Premiere Pro project to its existing file path. Equivalent to Cmd+S / Ctrl+S. No parameters required.

**Parameters:**

_None._

**Returns:**
```json
{ "status": "ok", "message": "Project saved." }
```

**Example:**
```
"Save the project"
-> premiere_save_project {}
```

---

### `premiere_get_project_info`

Retrieve detailed information about the currently open project, including project name, file path, all sequences (with indices, names, and resolutions), bin structure, total item counts, and the active sequence.

**Parameters:**

_None._

**Returns:**
```json
{
  "name": "MyEdit",
  "path": "/Users/me/Projects/MyEdit.prproj",
  "sequences": [
    { "index": 0, "name": "Main Edit", "width": 1920, "height": 1080 }
  ],
  "bins": ["Footage", "Audio", "Graphics"],
  "total_items": 47,
  "active_sequence_index": 0
}
```

**Example:**
```
"What's in this project?"
-> premiere_get_project_info {}
```

---

### `premiere_import_files`

Import multiple media files into the Premiere Pro project in a single operation. Faster than calling premiere_import_media repeatedly. Optionally import into a specific bin (created automatically if it does not exist).

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `file_paths` | string[] | Yes | — | Array of absolute file paths to import (e.g. `["/Users/me/footage/clip01.mp4", "/Users/me/audio/narration.wav"]`). |
| `target_bin` | string | No | — | Slash-separated bin path to import into (e.g. `Footage/Raw`). The bin hierarchy is created automatically if it does not exist. If omitted, files go to the project root. |

**Returns:**
```json
{ "status": "ok", "imported_count": 3, "items": ["clip01.mp4", "clip02.mp4", "narration.wav"] }
```

**Example:**
```
"Import these clips into the B-Roll bin"
-> premiere_import_files { file_paths: ["/Users/me/footage/broll1.mp4", "/Users/me/footage/broll2.mp4"], target_bin: "Footage/B-Roll" }
```

---

### `premiere_create_bin`

Create a new bin (folder) in the Premiere Pro project panel for organizing media. Bins can be nested. If the parent bin does not exist, the operation fails.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `name` | string | Yes | — | Name for the new bin (e.g. `B-Roll`, `Interview Footage`, `Music`). Must be unique within the parent bin. |
| `parent_bin` | string | No | — | Slash-separated path to the parent bin (e.g. `Footage/Day1`). If omitted, the bin is created at the project root. |

**Returns:**
```json
{ "status": "ok", "bin_name": "B-Roll", "path": "Footage/B-Roll" }
```

**Example:**
```
"Create a Music bin inside Audio"
-> premiere_create_bin { name: "Music", parent_bin: "Audio" }
```

---

### `premiere_find_project_items`

Search for project items by name across all bins in the project (recursive). Returns matching items with their bin paths, types, and metadata.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `query` | string | Yes | — | Search query string (case-insensitive substring match). Examples: `interview`, `.mp4`, `B-roll`. |

**Returns:**
```json
{
  "results": [
    { "name": "interview_01.mp4", "path": "Footage/Interviews/interview_01.mp4", "type": "clip", "duration": 342.5 },
    { "name": "interview_02.mp4", "path": "Footage/Interviews/interview_02.mp4", "type": "clip", "duration": 198.2 }
  ]
}
```

**Example:**
```
"Find all interview clips"
-> premiere_find_project_items { query: "interview" }
```

---

### `premiere_get_project_items`

List all items (clips, sequences, bins) in a specific bin of the project panel. Returns item names, types, indices, and metadata.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `bin_path` | string | No | — | Slash-separated path to the bin to list (e.g. `Footage/Interviews`). If omitted, lists items at the project root. |

**Returns:**
```json
{
  "items": [
    { "index": 0, "name": "clip01.mp4", "type": "clip" },
    { "index": 1, "name": "Interviews", "type": "bin" }
  ]
}
```

**Example:**
```
"List everything in the Footage bin"
-> premiere_get_project_items { bin_path: "Footage" }
```

---

## Sequence

### `premiere_create_sequence`

Create a new empty sequence in the active Premiere Pro project with the specified resolution, frame rate, and track layout. The sequence becomes the active sequence after creation.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `name` | string | Yes | — | Display name for the new sequence (e.g. `Main Edit`, `Social Cut 9x16`) |
| `width` | number | No | `1920` | Frame width in pixels. Common values: 1920 (1080p), 3840 (4K UHD), 1080 (vertical 9:16). |
| `height` | number | No | `1080` | Frame height in pixels. Common values: 1080 (1080p/vertical), 2160 (4K UHD), 1920 (vertical 9:16). |
| `frame_rate` | number | No | `24` | Frame rate in fps. Common values: 23.976, 24, 25 (PAL), 29.97, 30, 50, 59.94, 60. |
| `video_tracks` | number | No | `3` | Number of video tracks to create. Minimum 1. |
| `audio_tracks` | number | No | `2` | Number of stereo audio tracks to create. Minimum 1. |

**Returns:**
```json
{ "name": "Interview", "id": "seq-abc123", "width": 3840, "height": 2160, "frame_rate": 30 }
```

**Example:**
```
"Create a 4K 30fps sequence called Interview"
-> premiere_create_sequence { name: "Interview", width: 3840, height: 2160, frame_rate: 30 }
```

---

### `premiere_get_active_sequence`

Get details of the currently active sequence including name, resolution, frame rate, track counts, and duration.

**Parameters:**

_None._

**Returns:**
```json
{
  "name": "Main Edit",
  "width": 1920,
  "height": 1080,
  "frame_rate": 24,
  "video_track_count": 3,
  "audio_track_count": 2,
  "duration_seconds": 145.5
}
```

**Example:**
```
"What sequence am I working on?"
-> premiere_get_active_sequence {}
```

---

### `premiere_get_sequence_list`

Retrieve a list of all sequences in the current project with their indices, names, resolutions, and durations.

**Parameters:**

_None._

**Returns:**
```json
{
  "sequences": [
    { "index": 0, "name": "Main Edit", "width": 1920, "height": 1080 },
    { "index": 1, "name": "Social Cut", "width": 1080, "height": 1920 }
  ]
}
```

**Example:**
```
"Show me all sequences"
-> premiere_get_sequence_list {}
```

---

### `premiere_get_playhead_position`

Get the current playhead position in the active sequence as seconds, timecode, and frame number.

**Parameters:**

_None. Uses the currently active sequence._

**Returns:**
```json
{ "seconds": 42.5, "timecode": "00:00:42:12", "frame": 1020 }
```

**Example:**
```
"Where is the playhead?"
-> premiere_get_playhead_position {}
```

---

### `premiere_set_playhead_position`

Move the playhead to a specific position in the active sequence.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `seconds` | number | Yes | — | Target position in seconds (e.g. `10.5` moves to the 10.5-second mark). |

**Returns:**
```json
{ "status": "ok", "seconds": 10.5, "timecode": "00:00:10:12" }
```

**Example:**
```
"Go to the 30-second mark"
-> premiere_set_playhead_position { seconds: 30 }
```

---

### `premiere_set_in_point`

Set the sequence in-point at the current playhead position or at a specified time. The in-point defines the start of a work area for playback, export, or lift/extract operations.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `seconds` | number | No | — | Time in seconds to set the in-point. If omitted, uses the current playhead position. |

**Returns:**
```json
{ "status": "ok", "in_point_seconds": 5.0 }
```

**Example:**
```
"Set in-point at 5 seconds"
-> premiere_set_in_point { seconds: 5 }
```

---

### `premiere_set_out_point`

Set the sequence out-point at the current playhead position or at a specified time. The out-point defines the end of a work area for playback, export, or lift/extract operations.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `seconds` | number | No | — | Time in seconds to set the out-point. If omitted, uses the current playhead position. |

**Returns:**
```json
{ "status": "ok", "out_point_seconds": 15.0 }
```

**Example:**
```
"Set out-point at 15 seconds"
-> premiere_set_out_point { seconds: 15 }
```

---

### `premiere_get_timeline`

Retrieve the full state of a sequence's timeline, including every video and audio track, all clips on each track (with names, positions, durations, in/out points), and applied effects.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `sequence_id` | string | No | — | Unique identifier of the sequence to inspect. If omitted, uses the currently active sequence. |

**Returns:**
```json
{
  "sequence_name": "Main Edit",
  "video_tracks": [
    {
      "index": 0,
      "clips": [
        { "id": "clip-001", "name": "interview.mp4", "start": 0.0, "end": 10.5, "duration": 10.5 }
      ]
    }
  ],
  "audio_tracks": [
    {
      "index": 0,
      "clips": [
        { "id": "clip-002", "name": "narration.wav", "start": 0.0, "end": 30.0 }
      ]
    }
  ]
}
```

**Example:**
```
"Show me the full timeline"
-> premiere_get_timeline {}
```

---

### `premiere_add_sequence_marker`

Add a marker to the active sequence timeline at a specified position. Markers are used for annotation, navigation, and chapter points.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `time` | number | Yes | — | Marker position on the timeline in seconds. |
| `name` | string | No | `"Marker"` | Display name for the marker. |
| `comment` | string | No | — | Comment text attached to the marker. |
| `color` | number | No | `0` | Marker color index (0-7). |

**Returns:**
```json
{ "status": "ok", "marker_index": 3, "time": 42.0, "name": "Chapter 2" }
```

**Example:**
```
"Add a green marker at 42 seconds called Chapter 2"
-> premiere_add_sequence_marker { time: 42, name: "Chapter 2", color: 3 }
```

---

### `premiere_get_sequence_settings`

Get the full settings of a sequence including resolution, frame rate, pixel aspect ratio, field order, preview codec, and audio sample rate.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `sequence_index` | number | Yes | — | Zero-based index of the sequence. Obtain from premiere_get_sequence_list. |

**Returns:**
```json
{
  "name": "Main Edit",
  "width": 1920,
  "height": 1080,
  "frame_rate": 23.976,
  "pixel_aspect_ratio": "1.0",
  "field_order": "progressive",
  "audio_sample_rate": 48000
}
```

**Example:**
```
"Get the settings for the first sequence"
-> premiere_get_sequence_settings { sequence_index: 0 }
```

---

## Clips

### `premiere_insert_clip`

Insert a clip onto the timeline at the playhead position using insert mode, which pushes all subsequent clips forward. The clip must already be imported into the project.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `source_path` | string | Yes | — | Absolute path to the source media file on disk, or the clip's project item ID. The file must already be imported. |
| `track_type` | string | No | `"video"` | Type of track to place the clip on. Values: `video`, `audio`. |
| `track_index` | number | No | `0` | Zero-based track index. Track 0 is the bottom-most track. |
| `position_seconds` | number | No | `0` | Timeline position in seconds where the clip is inserted. |

**Returns:**
```json
{ "status": "ok", "clip_id": "clip-abc123", "position": 0.0, "duration": 15.3 }
```

**Example:**
```
"Insert clip01.mp4 at the 10-second mark on video track 1"
-> premiere_insert_clip { source_path: "/Users/me/footage/clip01.mp4", track_index: 1, position_seconds: 10 }
```

---

### `premiere_overwrite_clip`

Place (overwrite) a media clip onto the active sequence timeline at a specified position and track. Overwrites any existing material at that position. The clip must already be imported into the project.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `source_path` | string | Yes | — | Absolute path to the source media file on disk, or the clip's project item ID. |
| `track_type` | string | No | `"video"` | Type of track: `video` or `audio`. |
| `track_index` | number | No | `0` | Zero-based track index. |
| `position_seconds` | number | No | `0` | Timeline position in seconds where the clip starts. |
| `in_point_seconds` | number | No | — | Source in-point in seconds (the point in the original media where playback begins). |
| `out_point_seconds` | number | No | — | Source out-point in seconds (the point in the original media where playback ends). |
| `speed` | number | No | `1.0` | Playback speed multiplier. 2.0 = double speed, 0.5 = half speed. |

**Returns:**
```json
{ "status": "ok", "clip_id": "clip-xyz789", "position": 5.0, "duration": 10.0 }
```

**Example:**
```
"Place the interview clip at 5 seconds, using only seconds 30-60 of the source"
-> premiere_overwrite_clip { source_path: "/Users/me/footage/interview.mp4", position_seconds: 5, in_point_seconds: 30, out_point_seconds: 60 }
```

---

### `premiere_remove_clip`

Remove a clip from the timeline by its clip ID. This performs a lift edit (leaves a gap). For ripple delete (closing the gap), use premiere_ripple_delete_gap instead.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `clip_id` | string | Yes | — | Unique identifier of the clip to remove. Obtain from premiere_get_timeline or premiere_get_all_clips. |
| `sequence_id` | string | Yes | — | Unique identifier of the sequence containing the clip. |

**Returns:**
```json
{ "status": "ok" }
```

**Example:**
```
"Remove clip-001 from the main sequence"
-> premiere_remove_clip { clip_id: "clip-001", sequence_id: "seq-abc" }
```

---

### `premiere_razor_clip`

Split (razor) a clip at a specified time position on the timeline, creating two independent clips at the cut point.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `track_type` | string | Yes | — | Track type: `video` or `audio`. |
| `track_index` | number | Yes | — | Zero-based track index. |
| `time_seconds` | number | Yes | — | Timeline position in seconds where the cut is made. |

**Returns:**
```json
{ "status": "ok", "cut_at": 15.0 }
```

**Example:**
```
"Split the clip on video track 0 at the 15-second mark"
-> premiere_razor_clip { track_type: "video", track_index: 0, time_seconds: 15 }
```

---

### `premiere_get_clip_info`

Get detailed information about a specific clip on the timeline, including source media path, in/out points, position, duration, speed, and applied effects.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `track_type` | string | Yes | — | Track type: `video` or `audio`. |
| `track_index` | number | Yes | — | Zero-based track index. |
| `clip_index` | number | Yes | — | Zero-based clip index on the track. |

**Returns:**
```json
{
  "name": "interview.mp4",
  "source_path": "/Users/me/footage/interview.mp4",
  "start": 5.0,
  "end": 65.0,
  "duration": 60.0,
  "in_point": 30.0,
  "out_point": 90.0,
  "speed": 1.0,
  "effects": ["Lumetri Color"]
}
```

**Example:**
```
"Get info about the first clip on video track 0"
-> premiere_get_clip_info { track_type: "video", track_index: 0, clip_index: 0 }
```

---

### `premiere_set_clip_speed`

Change the playback speed of a clip on the timeline. Adjusts the clip's duration accordingly.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `track_type` | string | Yes | — | Track type: `video` or `audio`. |
| `track_index` | number | Yes | — | Zero-based track index. |
| `clip_index` | number | Yes | — | Zero-based clip index on the track. |
| `speed` | number | Yes | — | Speed multiplier. 1.0 = normal, 2.0 = double speed, 0.5 = half speed, -1.0 = reverse. |

**Returns:**
```json
{ "status": "ok", "speed": 2.0, "new_duration": 5.0 }
```

**Example:**
```
"Make the second clip on track 0 play at half speed"
-> premiere_set_clip_speed { track_type: "video", track_index: 0, clip_index: 1, speed: 0.5 }
```

---

### `premiere_trim_clip_start`

Trim the start (in-point) of a clip on the timeline by shifting it forward or backward.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `track_type` | string | Yes | — | Track type: `video` or `audio`. |
| `track_index` | number | Yes | — | Zero-based track index. |
| `clip_index` | number | Yes | — | Zero-based clip index on the track. |
| `delta_seconds` | number | Yes | — | Seconds to trim (positive = shorten from start, negative = extend from start). |

**Returns:**
```json
{ "status": "ok", "new_start": 2.0, "new_duration": 8.0 }
```

**Example:**
```
"Trim 2 seconds off the beginning of the first clip"
-> premiere_trim_clip_start { track_type: "video", track_index: 0, clip_index: 0, delta_seconds: 2 }
```

---

### `premiere_trim_clip_end`

Trim the end (out-point) of a clip on the timeline by extending or shortening it.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `track_type` | string | Yes | — | Track type: `video` or `audio`. |
| `track_index` | number | Yes | — | Zero-based track index. |
| `clip_index` | number | Yes | — | Zero-based clip index on the track. |
| `delta_seconds` | number | Yes | — | Seconds to trim (positive = extend end, negative = shorten end). |

**Returns:**
```json
{ "status": "ok", "new_end": 12.0, "new_duration": 12.0 }
```

**Example:**
```
"Extend the first clip by 3 seconds at the end"
-> premiere_trim_clip_end { track_type: "video", track_index: 0, clip_index: 0, delta_seconds: 3 }
```

---

### `premiere_select_clip`

Select a specific clip on the timeline by its track and index.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `track_type` | string | Yes | — | Track type: `video` or `audio`. |
| `track_index` | number | Yes | — | Zero-based track index. |
| `clip_index` | number | Yes | — | Zero-based clip index on the track. |
| `add_to_selection` | boolean | No | `false` | If true, add to existing selection rather than replacing it. |

**Returns:**
```json
{ "status": "ok", "selected_clip": "interview.mp4" }
```

**Example:**
```
"Select the third clip on video track 1"
-> premiere_select_clip { track_type: "video", track_index: 1, clip_index: 2 }
```

---

### `premiere_deselect_all`

Deselect all currently selected clips on the timeline.

**Parameters:**

_None._

**Returns:**
```json
{ "status": "ok" }
```

**Example:**
```
"Deselect everything"
-> premiere_deselect_all {}
```

---

## Effects & Color

### `premiere_add_video_transition`

Add a video transition effect at a cut point on the timeline. The transition is placed at the specified position on the given track.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `sequence_id` | string | Yes | — | Unique identifier of the target sequence. |
| `track_index` | number | No | `0` | Zero-based video track index where the transition is applied. |
| `position_seconds` | number | No | `0` | Timeline position in seconds. Should align with a cut point between two adjacent clips. |
| `type` | string | No | `"cross_dissolve"` | Transition type. Values: `cross_dissolve`, `dip_to_black`, `dip_to_white`, `film_dissolve`, `morph_cut`. |
| `duration_seconds` | number | No | `1.0` | Duration of the transition in seconds. Typical range: 0.25 to 2.0. |

**Returns:**
```json
{ "status": "ok" }
```

**Example:**
```
"Add a 0.5-second cross dissolve at the 10-second mark"
-> premiere_add_video_transition { sequence_id: "seq-abc", position_seconds: 10, type: "cross_dissolve", duration_seconds: 0.5 }
```

---

### `premiere_apply_video_effect`

Apply a video effect to a clip on the timeline by effect name.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `track_index` | number | Yes | — | Zero-based video track index. |
| `clip_index` | number | Yes | — | Zero-based clip index on the track. |
| `effect_name` | string | Yes | — | Name of the effect to apply (e.g. `Gaussian Blur`, `Lumetri Color`, `Warp Stabilizer`). |

**Returns:**
```json
{ "status": "ok", "effect": "Gaussian Blur" }
```

**Example:**
```
"Apply Lumetri Color to the first clip"
-> premiere_apply_video_effect { track_index: 0, clip_index: 0, effect_name: "Lumetri Color" }
```

---

### `premiere_set_effect_parameter`

Set a specific parameter value on an effect applied to a clip.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `track_index` | number | Yes | — | Zero-based video track index. |
| `clip_index` | number | Yes | — | Zero-based clip index on the track. |
| `effect_name` | string | Yes | — | Name of the effect containing the parameter. |
| `parameter_name` | string | Yes | — | Name of the parameter to set. |
| `value` | number | Yes | — | New value for the parameter. |

**Returns:**
```json
{ "status": "ok", "effect": "Gaussian Blur", "parameter": "Blurriness", "value": 15.0 }
```

**Example:**
```
"Set the blurriness to 15 on the Gaussian Blur effect"
-> premiere_set_effect_parameter { track_index: 0, clip_index: 0, effect_name: "Gaussian Blur", parameter_name: "Blurriness", value: 15 }
```

---

### `premiere_set_position`

Set the position (X, Y) of a video clip's Motion effect. Coordinates are in pixels relative to the sequence frame.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `track_index` | number | Yes | — | Zero-based video track index. |
| `clip_index` | number | Yes | — | Zero-based clip index on the track. |
| `x` | number | Yes | — | Horizontal position in pixels. 960 = center of 1920-wide frame. |
| `y` | number | Yes | — | Vertical position in pixels. 540 = center of 1080-tall frame. |

**Returns:**
```json
{ "status": "ok", "x": 960, "y": 540 }
```

**Example:**
```
"Move the clip to the top-left quadrant"
-> premiere_set_position { track_index: 0, clip_index: 0, x: 480, y: 270 }
```

---

### `premiere_set_scale`

Set the scale percentage of a video clip's Motion effect.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `track_index` | number | Yes | — | Zero-based video track index. |
| `clip_index` | number | Yes | — | Zero-based clip index on the track. |
| `scale` | number | Yes | — | Scale percentage. 100 = original size, 50 = half, 200 = double. |

**Returns:**
```json
{ "status": "ok", "scale": 150 }
```

**Example:**
```
"Scale the clip up to 150%"
-> premiere_set_scale { track_index: 0, clip_index: 0, scale: 150 }
```

---

### `premiere_set_opacity`

Set the opacity of a video clip (0 = fully transparent, 100 = fully opaque).

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `track_index` | number | Yes | — | Zero-based video track index. |
| `clip_index` | number | Yes | — | Zero-based clip index on the track. |
| `opacity` | number | Yes | — | Opacity percentage. 0 = transparent, 100 = opaque. |

**Returns:**
```json
{ "status": "ok", "opacity": 75 }
```

**Example:**
```
"Set the overlay clip to 50% opacity"
-> premiere_set_opacity { track_index: 1, clip_index: 0, opacity: 50 }
```

---

### `premiere_lumetri_set_exposure`

Set the exposure value on a clip's Lumetri Color effect. The Lumetri effect is auto-applied if not already present.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `track_index` | number | Yes | — | Zero-based video track index. |
| `clip_index` | number | Yes | — | Zero-based clip index on the track. |
| `exposure` | number | Yes | — | Exposure value in stops. 0 = no change, positive = brighter, negative = darker. Typical range: -4.0 to +4.0. |

**Returns:**
```json
{ "status": "ok", "exposure": 1.5 }
```

**Example:**
```
"Brighten the clip by 1.5 stops"
-> premiere_lumetri_set_exposure { track_index: 0, clip_index: 0, exposure: 1.5 }
```

---

### `premiere_lumetri_set_contrast`

Set the contrast value on a clip's Lumetri Color effect. The Lumetri effect is auto-applied if not already present.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `track_index` | number | Yes | — | Zero-based video track index. |
| `clip_index` | number | Yes | — | Zero-based clip index on the track. |
| `contrast` | number | Yes | — | Contrast value. 0 = no change, positive = more contrast, negative = less. Typical range: -100 to +100. |

**Returns:**
```json
{ "status": "ok", "contrast": 25 }
```

**Example:**
```
"Increase contrast by 25"
-> premiere_lumetri_set_contrast { track_index: 0, clip_index: 0, contrast: 25 }
```

---

### `premiere_lumetri_set_saturation`

Set the saturation value on a clip's Lumetri Color effect. The Lumetri effect is auto-applied if not already present.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `track_index` | number | Yes | — | Zero-based video track index. |
| `clip_index` | number | Yes | — | Zero-based clip index on the track. |
| `saturation` | number | Yes | — | Saturation value. 100 = no change, 0 = fully desaturated (grayscale), 200 = double saturation. |

**Returns:**
```json
{ "status": "ok", "saturation": 120 }
```

**Example:**
```
"Desaturate the clip to grayscale"
-> premiere_lumetri_set_saturation { track_index: 0, clip_index: 0, saturation: 0 }
```

---

### `premiere_add_keyframe`

Add a keyframe at a specific time on a clip's effect parameter. Enables animation of effect values over time.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `track_index` | number | Yes | — | Zero-based video track index. |
| `clip_index` | number | Yes | — | Zero-based clip index on the track. |
| `effect_name` | string | Yes | — | Name of the effect containing the parameter. |
| `parameter_name` | string | Yes | — | Name of the parameter to keyframe. |
| `time_seconds` | number | Yes | — | Time in seconds relative to clip start for the keyframe. |
| `value` | number | Yes | — | Parameter value at this keyframe. |

**Returns:**
```json
{ "status": "ok", "keyframe_time": 2.0, "value": 50 }
```

**Example:**
```
"Add an opacity keyframe at 0s=0% and 1s=100% to fade in"
-> premiere_add_keyframe { track_index: 0, clip_index: 0, effect_name: "Opacity", parameter_name: "Opacity", time_seconds: 0, value: 0 }
-> premiere_add_keyframe { track_index: 0, clip_index: 0, effect_name: "Opacity", parameter_name: "Opacity", time_seconds: 1, value: 100 }
```

---

## Audio & Export

### `premiere_set_audio_level`

Set the audio volume level (in decibels) for a specific clip on the timeline.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `clip_id` | string | Yes | — | Unique identifier of the audio clip. Obtain from premiere_get_timeline. |
| `sequence_id` | string | Yes | — | Unique identifier of the sequence containing the clip. |
| `level_db` | number | Yes | — | Audio level in decibels. 0 = unity gain (no change), -6 = half volume, -96 = silence, +6 = double volume, +15 = maximum boost. |

**Returns:**
```json
{ "status": "ok" }
```

**Example:**
```
"Reduce the narration clip to -6 dB"
-> premiere_set_audio_level { clip_id: "clip-audio-001", sequence_id: "seq-abc", level_db: -6 }
```

---

### `premiere_normalize_audio`

Normalize audio levels on a clip to a target peak or loudness level.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `track_index` | number | Yes | — | Zero-based audio track index. |
| `clip_index` | number | Yes | — | Zero-based clip index on the track. |
| `target_db` | number | No | `-3` | Target peak level in dB (default: -3). Common targets: -1 for broadcast, -3 for online, -6 for dialogue. |

**Returns:**
```json
{ "status": "ok", "original_peak": -12.5, "applied_gain": 9.5, "new_peak": -3.0 }
```

**Example:**
```
"Normalize the interview audio to -3 dB"
-> premiere_normalize_audio { track_index: 0, clip_index: 0, target_db: -3 }
```

---

### `premiere_mute_audio_track`

Mute or unmute an entire audio track.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `track_index` | number | Yes | — | Zero-based audio track index. |
| `muted` | boolean | Yes | — | `true` to mute, `false` to unmute. |

**Returns:**
```json
{ "status": "ok", "track_index": 1, "muted": true }
```

**Example:**
```
"Mute audio track 2"
-> premiere_mute_audio_track { track_index: 1, muted: true }
```

---

### `premiere_export`

Export a sequence to a media file using a built-in preset. This is the simplest export option.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `sequence_id` | string | No | — | Unique identifier of the sequence to export. If omitted, exports the active sequence. |
| `output_path` | string | Yes | — | Absolute file path for the exported output (e.g. `/Users/me/exports/final.mp4`). |
| `preset` | string | No | `"h264_1080p"` | Export preset name. Values: `h264_1080p`, `h264_4k`, `prores_422`, `prores_4444`, `dnxhd`, `gif`. |

**Returns:**
```json
{ "status": "ok", "output_path": "/Users/me/exports/final.mp4", "file_size_mb": 245.3 }
```

**Example:**
```
"Export the sequence as a 4K MP4"
-> premiere_export { output_path: "/Users/me/exports/final_4k.mp4", preset: "h264_4k" }
```

---

### `premiere_export_direct`

Export a sequence with an Adobe Media Encoder preset file (.epr) for full control over encoding settings. Supports work area selection.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `sequence_id` | string | No | — | Unique identifier of the sequence to export. If omitted, exports the active sequence. |
| `output_path` | string | Yes | — | Absolute file path for the exported output. |
| `preset_path` | string | Yes | — | Absolute path to an Adobe Media Encoder preset file (.epr). |
| `work_area_only` | boolean | No | `false` | If true, export only the work area (between in/out points). |

**Returns:**
```json
{ "status": "ok", "output_path": "/Users/me/exports/final.mp4" }
```

**Example:**
```
"Export using my custom YouTube preset"
-> premiere_export_direct { output_path: "/Users/me/exports/youtube.mp4", preset_path: "/Users/me/presets/YouTube_1080p.epr" }
```

---

### `premiere_export_frame`

Export the current frame at the playhead as a still image file (PNG, JPEG, or TIFF).

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `output_path` | string | Yes | — | Absolute file path for the exported frame (e.g. `/Users/me/frames/thumbnail.png`). |
| `format` | string | No | `"png"` | Image format. Values: `png`, `jpeg`, `tiff`. |

**Returns:**
```json
{ "status": "ok", "output_path": "/Users/me/frames/thumbnail.png", "timecode": "00:01:23:12" }
```

**Example:**
```
"Export the current frame as a PNG thumbnail"
-> premiere_export_frame { output_path: "/Users/me/frames/thumb.png" }
```

---

### `premiere_export_for_youtube`

Export a sequence optimized for YouTube upload with recommended settings (H.264, AAC audio, appropriate bitrate).

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `output_path` | string | Yes | — | Absolute file path for the exported output. |
| `resolution` | string | No | `"1080p"` | Target resolution. Values: `720p`, `1080p`, `4k`. |

**Returns:**
```json
{ "status": "ok", "output_path": "/Users/me/exports/youtube_video.mp4", "resolution": "1080p" }
```

**Example:**
```
"Export for YouTube in 4K"
-> premiere_export_for_youtube { output_path: "/Users/me/exports/yt_upload.mp4", resolution: "4k" }
```

---

### `premiere_capture_frame_base64`

Capture the current frame at the playhead position and return it as a base64-encoded PNG image. The image is returned as an MCP image content block that the AI can visually inspect.

**Parameters:**

_None._

**Returns:**

An MCP image content block with metadata:
```json
{
  "format": "png",
  "width": 1920,
  "height": 1080,
  "timecode": "00:01:23:12"
}
```

**Example:**
```
"Show me what the current frame looks like"
-> premiere_capture_frame_base64 {}
```

---

### `premiere_execute_extendscript`

Execute an arbitrary ExtendScript snippet with security validation. Dangerous operations (system calls, file deletion, infinite loops, app.quit) are blocked by default.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `script` | string | Yes | — | The ExtendScript code to execute. |
| `validate` | boolean | No | `true` | When true, runs security checks to block dangerous operations. Set to false to bypass (use with caution). |

**Returns:**
```json
{ "status": "ok", "result": "Return value from the script" }
```

**Example:**
```
"Get the Premiere Pro version via script"
-> premiere_execute_extendscript { script: "app.version" }
```

---

### `premiere_auto_edit`

Perform a fully automated edit: scan an assets directory, parse a script, match script segments to media files, and assemble a complete sequence with clips, transitions, and text overlays.

**Parameters:**
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `script_path` | string | No | — | Absolute path to the script file. Provide this or `script_text`, not both. |
| `script_text` | string | No | — | Raw script text. Provide this or `script_path`, not both. |
| `assets_directory` | string | Yes | — | Absolute path to the directory containing media assets to match against the script. Scanned recursively. |
| `output_name` | string | No | — | Name for the generated sequence. Auto-generated from script content if omitted. |
| `resolution` | string | No | `"1080p"` | Output resolution: `1080p` (1920x1080) or `4k` (3840x2160). |
| `pacing` | string | No | `"moderate"` | Editing pace that controls cut rhythm. Values: `slow`, `moderate`, `fast`, `dynamic`. |

**Returns:**
```json
{
  "status": "ok",
  "sequence_name": "Episode 1 Assembly",
  "clip_count": 12,
  "duration_seconds": 185.3,
  "transitions_added": 11
}
```

**Example:**
```
"Auto-edit my podcast with fast pacing"
-> premiere_auto_edit { script_path: "/Users/me/scripts/ep1.fountain", assets_directory: "/Users/me/footage", pacing: "fast" }
```
