# PremierPro MCP Server -- Comprehensive Feature Plan

Master plan for every MCP tool the PremierPro MCP Server will expose. Each tool maps to one or more Adobe Premiere Pro ExtendScript/QE DOM operations, orchestrated through the four-language stack (Go orchestration, Rust media processing, Python intelligence, TypeScript ExtendScript bridge).

**Current state:** 16 tools implemented (Phase 0). This plan adds ~230 additional tools across 7 phases.

---

## Language Responsibility Key

| Abbreviation | Language | Role |
|---|---|---|
| **Go** | Go Orchestrator | MCP protocol handler, request routing, parameter validation, task coordination |
| **TS** | TypeScript Bridge | ExtendScript/CEP execution inside Premiere Pro, QE DOM access |
| **Rust** | Rust Engine | Media file analysis, waveform processing, thumbnail generation, asset indexing |
| **Python** | Python Intelligence | NLP, AI-driven edit decisions, script parsing, shot matching, pacing analysis |

---

## Phase 0 -- Foundation (IMPLEMENTED)

Already shipped. These tools form the baseline.

| # | Tool Name | Description | Primary Language |
|---|---|---|---|
| 1 | `premiere_open` | Launch Premiere Pro, optionally open a .prproj file | Go (OS commands) |
| 2 | `premiere_close` | Quit Premiere Pro gracefully or force-kill | Go (OS commands) |
| 3 | `premiere_is_running` | Check if Premiere Pro process is active | Go (OS commands) |
| 4 | `premiere_ping` | Verify Premiere Pro is reachable and get version/bridge info | Go -> TS |
| 5 | `premiere_get_project` | Get project name, path, sequences, bins, save state | Go -> TS |
| 6 | `premiere_create_sequence` | Create a new sequence with resolution/fps/track counts | Go -> TS |
| 7 | `premiere_import_media` | Import a single media file into the project | Go -> TS |
| 8 | `premiere_place_clip` | Place a clip on timeline at track/position with speed | Go -> TS |
| 9 | `premiere_remove_clip` | Remove a clip from a sequence | Go -> TS |
| 10 | `premiere_add_transition` | Add a transition at a position on a track | Go -> TS |
| 11 | `premiere_add_text` | Add a text overlay via .mogrt | Go -> TS |
| 12 | `premiere_set_audio_level` | Set audio level of a clip in dB | Go -> TS |
| 13 | `premiere_get_timeline` | Get full timeline state (tracks, clips, durations) | Go -> TS |
| 14 | `premiere_export` | Export a sequence with a preset | Go -> TS |
| 15 | `premiere_scan_assets` | Scan directory for media assets with metadata | Go -> Rust |
| 16 | `premiere_parse_script` | Parse script file/text into structured segments | Go -> Python |
| 17 | `premiere_auto_edit` | Full automated edit: scan, parse, match, assemble, export | Go -> Rust + Python + TS |

---

## Phase 1 -- Project Management & Media Management

Core project lifecycle and media organization operations.

### 1.1 Project Lifecycle

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 1 | `premiere_new_project` | Create a new empty Premiere Pro project | `project_path` (string, required): absolute path for the new .prproj file; `name` (string): project name | Go -> TS |
| 2 | `premiere_open_project` | Open an existing .prproj file in Premiere Pro | `project_path` (string, required): path to .prproj file | Go (OS) -> TS |
| 3 | `premiere_save_project` | Save the current project | (none) | Go -> TS |
| 4 | `premiere_save_project_as` | Save the current project to a new path | `output_path` (string, required): new .prproj file path | Go -> TS |
| 5 | `premiere_close_project` | Close the current project | `save_first` (bool): save before closing (default: true); `prompt_if_dirty` (bool): prompt on unsaved changes (default: true) | Go -> TS |
| 6 | `premiere_import_project` | Import sequences from another .prproj file | `project_path` (string, required): path to source .prproj; `sequence_ids` (string[]): specific sequence IDs to import (empty = all) | Go -> TS |
| 7 | `premiere_get_project_settings` | Get current project settings (scratch disks, renderer, etc.) | (none) | Go -> TS |
| 8 | `premiere_set_scratch_disks` | Set project scratch disk paths | `captured_video` (string): path for captured video; `captured_audio` (string): path for captured audio; `preview_video` (string): path for preview renders; `preview_audio` (string): path for preview audio | Go -> TS |

### 1.2 Project Metadata

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 9 | `premiere_get_project_metadata` | Get private project metadata for an item | `item_id` (string, required): project item ID | Go -> TS |
| 10 | `premiere_set_project_metadata` | Set private project metadata on an item | `item_id` (string, required): project item ID; `metadata_xml` (string, required): metadata XML string; `field_updates` (object[]): array of {field, value} pairs | Go -> TS |
| 11 | `premiere_get_xmp_metadata` | Get XMP metadata for a project item | `item_id` (string, required): project item ID | Go -> TS |
| 12 | `premiere_set_xmp_metadata` | Set XMP metadata on a project item | `item_id` (string, required): project item ID; `metadata_xml` (string, required): XMP metadata XML string | Go -> TS |
| 13 | `premiere_add_metadata_field` | Add a custom field to project metadata schema | `field_name` (string, required): internal name; `field_label` (string, required): display label; `field_type` (int, required): field type constant (0=int, 1=float, 2=string, 3=bool) | Go -> TS |

### 1.3 Media Import & Organization

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 14 | `premiere_import_files` | Batch import multiple files | `file_paths` (string[], required): array of absolute file paths; `target_bin` (string): destination bin path; `suppress_ui` (bool): suppress import dialog (default: true); `as_numbered_stills` (bool): treat as numbered still sequence | Go -> TS |
| 15 | `premiere_import_folder` | Import all media from a folder | `folder_path` (string, required): absolute folder path; `target_bin` (string): destination bin path; `recursive` (bool): include subfolders (default: true); `extensions` (string[]): filter by extensions | Go -> TS |
| 16 | `premiere_create_bin` | Create a new bin (folder) in the project panel | `name` (string, required): bin name; `parent_bin` (string): parent bin path (default: root) | Go -> TS |
| 17 | `premiere_rename_bin` | Rename an existing bin | `bin_id` (string, required): bin ID; `new_name` (string, required): new bin name | Go -> TS |
| 18 | `premiere_move_item` | Move a project item into a different bin | `item_id` (string, required): project item ID; `target_bin` (string, required): destination bin path or ID | Go -> TS |
| 19 | `premiere_delete_item` | Delete a project item (clip or bin) | `item_id` (string, required): project item ID; `delete_from_disk` (bool): also delete source file (default: false) | Go -> TS |
| 20 | `premiere_find_items` | Search project items by name, path, or metadata | `query` (string, required): search string; `search_type` (string): one of "name", "path", "metadata" (default: "name"); `media_type` (string): filter by "video", "audio", "image", "all" | Go -> TS |
| 21 | `premiere_find_items_by_path` | Find all project items referencing a specific media path | `media_path` (string, required): absolute file path to match | Go -> TS |
| 22 | `premiere_set_label_color` | Set the label color of a project item | `item_id` (string, required): project item ID; `color_index` (int, required): label color index (0-15) | Go -> TS |
| 23 | `premiere_list_bins` | List all bins in the project with hierarchy | `parent_bin` (string): start from this bin (default: root); `recursive` (bool): include nested bins (default: true) | Go -> TS |
| 24 | `premiere_list_project_items` | List all items in a bin with metadata | `bin_id` (string): bin to list (default: root); `recursive` (bool): include nested bins (default: false); `media_type` (string): filter by type | Go -> TS |

### 1.4 Proxy & Offline Media

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 25 | `premiere_attach_proxy` | Attach a proxy file to a project item | `item_id` (string, required): project item ID; `proxy_path` (string, required): path to proxy media; `is_hires` (bool): treat as hi-res replacement (default: false) | Go -> TS |
| 26 | `premiere_check_proxy` | Check if a project item can have a proxy attached | `item_id` (string, required): project item ID | Go -> TS |
| 27 | `premiere_toggle_proxy` | Toggle between proxy and full-res playback | `enable_proxy` (bool, required): true for proxy, false for full-res | Go -> TS |
| 28 | `premiere_relink_media` | Relink offline media to a new file path | `item_id` (string, required): project item ID; `new_path` (string, required): new absolute file path; `overwrite_checks` (bool): skip can-change-media check (default: false) | Go -> TS |
| 29 | `premiere_check_offline` | List all offline/unlinked media in the project | (none) | Go -> TS |
| 30 | `premiere_set_scale_to_frame` | Enable auto-scale to frame size for a project item | `item_id` (string, required): project item ID; `enabled` (bool, required): whether to enable scale-to-frame | Go -> TS |

---

## Phase 2 -- Sequence & Timeline Operations

Sequence lifecycle, settings, and timeline structure.

### 2.1 Sequence Lifecycle

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 31 | `premiere_create_sequence_from_preset` | Create a sequence from a .sqpreset file (via QE DOM) | `name` (string, required): sequence name; `preset_path` (string, required): absolute path to .sqpreset file | Go -> TS (QE) |
| 32 | `premiere_create_sequence_from_clip` | Create a sequence matching a clip's settings | `name` (string, required): sequence name; `item_ids` (string[], required): project item IDs to insert; `destination_bin` (string): target bin | Go -> TS |
| 33 | `premiere_duplicate_sequence` | Clone an existing sequence | `sequence_id` (string, required): source sequence ID; `new_name` (string): name for the clone | Go -> TS |
| 34 | `premiere_delete_sequence` | Delete a sequence from the project | `sequence_id` (string, required): sequence ID to delete | Go -> TS |
| 35 | `premiere_open_sequence` | Open a sequence in the timeline panel | `sequence_id` (string, required): sequence ID to open | Go -> TS |
| 36 | `premiere_close_sequence` | Close a sequence in the timeline panel | `sequence_id` (string, required): sequence ID to close | Go -> TS |
| 37 | `premiere_list_sequences` | List all sequences in the project with their settings | (none) | Go -> TS |
| 38 | `premiere_get_active_sequence` | Get the currently active sequence info | (none) | Go -> TS |
| 39 | `premiere_set_active_sequence` | Set the active sequence | `sequence_id` (string, required): sequence ID to activate | Go -> TS |

### 2.2 Sequence Settings

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 40 | `premiere_get_sequence_settings` | Get all settings of a sequence | `sequence_id` (string): sequence ID (default: active) | Go -> TS |
| 41 | `premiere_set_sequence_settings` | Modify sequence settings | `sequence_id` (string): sequence ID (default: active); `width` (int): frame width; `height` (int): frame height; `frame_rate` (float): fps; `pixel_aspect_ratio` (string): PAR; `field_type` (string): field order; `audio_sample_rate` (int): audio sample rate; `audio_display_format` (string): audio time display; `video_display_format` (string): timecode display | Go -> TS |
| 42 | `premiere_set_sequence_zero_point` | Set the starting timecode of a sequence | `sequence_id` (string): sequence ID (default: active); `start_time_ticks` (string, required): start time in ticks | Go -> TS |

### 2.3 Nested Sequences & Work Area

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 43 | `premiere_nest_clips` | Nest selected clips into a subsequence | `sequence_id` (string, required): parent sequence ID; `clip_ids` (string[], required): clip IDs to nest; `nested_name` (string): name for the nested sequence | Go -> TS |
| 44 | `premiere_set_work_area` | Set the work area in/out points | `sequence_id` (string): sequence ID (default: active); `in_seconds` (float, required): work area in point; `out_seconds` (float, required): work area out point | Go -> TS |
| 45 | `premiere_render_previews` | Render preview files for the work area | `sequence_id` (string): sequence ID (default: active) | Go -> TS |
| 46 | `premiere_delete_render_files` | Delete rendered preview files for a sequence | `sequence_id` (string): sequence ID (default: active) | Go -> TS |

---

## Phase 3 -- Clip Operations & Trimming

Placing, manipulating, and trimming clips on the timeline.

### 3.1 Clip Placement

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 47 | `premiere_insert_clip` | Insert a clip (ripple insert) at the playhead or time | `item_id` (string, required): project item ID; `time_seconds` (float, required): insertion point in seconds; `video_track` (int): video track index (default: 0); `audio_track` (int): audio track index (default: 0) | Go -> TS |
| 48 | `premiere_overwrite_clip` | Overwrite-place a clip at a position | `item_id` (string, required): project item ID; `time_seconds` (float, required): position in seconds; `video_track` (int): video track index (default: 0); `audio_track` (int): audio track index (default: 0) | Go -> TS |
| 49 | `premiere_lift_clip` | Lift a clip from the timeline (leave gap) | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID | Go -> TS |
| 50 | `premiere_extract_clip` | Extract a clip from the timeline (close gap) | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID | Go -> TS |
| 51 | `premiere_ripple_delete` | Delete a clip and close the resulting gap | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID | Go -> TS |
| 52 | `premiere_copy_clips` | Copy clips to the internal clipboard | `clip_ids` (string[], required): array of clip IDs; `sequence_id` (string, required): sequence ID | Go -> TS |
| 53 | `premiere_paste_clips` | Paste clips from clipboard at a position | `sequence_id` (string, required): target sequence ID; `time_seconds` (float, required): paste position; `track_index` (int): target track; `track_type` (string): "video" or "audio" | Go -> TS |
| 54 | `premiere_paste_insert` | Paste-insert clips (ripple) from clipboard | `sequence_id` (string, required): target sequence ID; `time_seconds` (float, required): paste position | Go -> TS |
| 55 | `premiere_paste_attributes` | Paste clip attributes (effects, motion, etc.) to target clips | `source_clip_id` (string, required): clip to copy attributes from; `target_clip_ids` (string[], required): clips to paste onto; `sequence_id` (string, required): sequence ID; `attributes` (string[]): specific attributes to paste ("motion", "opacity", "effects", "audio") | Go -> TS |
| 56 | `premiere_move_clip` | Move a clip to a different position or track | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `new_position_seconds` (float): new timeline position; `new_track_index` (int): new track index; `new_track_type` (string): "video" or "audio" | Go -> TS |
| 57 | `premiere_disable_clip` | Enable or disable a clip on the timeline | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `disabled` (bool, required): true to disable, false to enable | Go -> TS |

### 3.2 Subclips & Match Frame

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 58 | `premiere_make_subclip` | Create a subclip from a project item | `item_id` (string, required): source project item ID; `name` (string, required): subclip name; `in_seconds` (float, required): subclip in point; `out_seconds` (float, required): subclip out point; `restrict_trims` (bool): lock subclip boundaries (default: false) | Go -> TS |
| 59 | `premiere_match_frame` | Find the source frame matching the current playhead position | `sequence_id` (string): sequence ID (default: active); `track_index` (int): target video track (default: topmost with clip) | Go -> TS |
| 60 | `premiere_reverse_match_frame` | Find the timeline position of the current source monitor frame | `item_id` (string, required): project item ID; `time_seconds` (float, required): source time position | Go -> TS |

### 3.3 Razor / Split

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 61 | `premiere_razor_at_time` | Split clip(s) at a specific time on specified tracks | `sequence_id` (string): sequence ID (default: active); `time_seconds` (float, required): cut point in seconds; `track_index` (int): specific track (-1 for all targeted tracks); `track_type` (string): "video", "audio", or "both" (default: "both") | Go -> TS (QE) |
| 62 | `premiere_razor_all_tracks` | Razor (split) all tracks at a specific time | `sequence_id` (string): sequence ID (default: active); `time_seconds` (float, required): cut point in seconds | Go -> TS (QE) |

### 3.4 Trimming

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 63 | `premiere_ripple_trim` | Ripple trim a clip edge (adjusts subsequent clips) | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `edge` (string, required): "in" or "out"; `delta_seconds` (float, required): trim amount (positive = extend, negative = shorten) | Go -> TS |
| 64 | `premiere_roll_trim` | Roll trim (move edit point between two clips) | `clip_id` (string, required): clip ID at the edit point; `sequence_id` (string, required): sequence ID; `edge` (string, required): "in" or "out"; `delta_seconds` (float, required): amount to roll | Go -> TS |
| 65 | `premiere_slip_clip` | Slip a clip (change source in/out without moving it on timeline) | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `delta_seconds` (float, required): slip amount (positive = slip source right) | Go -> TS |
| 66 | `premiere_slide_clip` | Slide a clip (move it while adjusting adjacent clips) | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `delta_seconds` (float, required): slide amount | Go -> TS |
| 67 | `premiere_extend_edit` | Extend edit point to the playhead position | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `edge` (string, required): "in" or "out" | Go -> TS |
| 68 | `premiere_trim_to_playhead` | Trim clip in or out point to current playhead | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `edge` (string, required): "in" or "out"; `playhead_seconds` (float, required): playhead position | Go -> TS |
| 69 | `premiere_set_clip_in_out` | Directly set a clip's in and out points | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `in_seconds` (float): new source in point; `out_seconds` (float): new source out point | Go -> TS |

---

## Phase 4 -- Effects, Transitions & Motion

Visual effects, transitions, audio effects, motion properties, and keyframing.

### 4.1 Video Transitions

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 70 | `premiere_apply_video_transition` | Apply a video transition to a clip edge or between clips | `sequence_id` (string, required): sequence ID; `track_index` (int, required): video track index; `clip_index` (int, required): clip index on the track; `transition_name` (string, required): transition name (e.g., "Cross Dissolve", "Dip to Black"); `duration_seconds` (float): transition duration (default: 1.0); `alignment` (string): "center", "start", "end" (default: "center"); `apply_to_end` (bool): apply to end of clip (default: true) | Go -> TS (QE) |
| 71 | `premiere_apply_audio_transition` | Apply an audio transition between clips | `sequence_id` (string, required): sequence ID; `track_index` (int, required): audio track index; `clip_index` (int, required): clip index on the track; `transition_name` (string, required): e.g., "Constant Power", "Constant Gain", "Exponential Fade"; `duration_seconds` (float): duration (default: 1.0) | Go -> TS (QE) |
| 72 | `premiere_set_default_transition` | Set the default video or audio transition | `transition_name` (string, required): transition name; `type` (string, required): "video" or "audio"; `duration_seconds` (float): default duration | Go -> TS (QE) |
| 73 | `premiere_apply_transition_to_multiple` | Apply a transition to multiple clip edit points at once | `sequence_id` (string, required): sequence ID; `track_index` (int, required): track index; `clip_indices` (int[], required): array of clip indices; `transition_name` (string, required): transition name; `duration_seconds` (float): duration; `apply_to` (string): "in", "out", or "both" | Go -> TS (QE) |
| 74 | `premiere_remove_transition` | Remove a transition from a clip | `sequence_id` (string, required): sequence ID; `track_index` (int, required): track index; `clip_index` (int, required): clip index; `edge` (string): "in" or "out" | Go -> TS (QE) |
| 75 | `premiere_get_transition_list` | List all available video and audio transitions | `type` (string): "video", "audio", or "all" (default: "all") | Go -> TS (QE) |

### 4.2 Video Effects

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 76 | `premiere_apply_effect` | Apply a video or audio effect to a clip by name | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `effect_name` (string, required): effect name (e.g., "Gaussian Blur", "Lumetri Color"); `is_audio` (bool): true for audio effect (default: false) | Go -> TS (QE) |
| 77 | `premiere_remove_effect` | Remove an applied effect from a clip | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `effect_index` (int, required): index of the effect component to remove | Go -> TS |
| 78 | `premiere_enable_effect` | Enable or disable an effect on a clip | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `effect_index` (int, required): effect component index; `enabled` (bool, required): true to enable | Go -> TS |
| 79 | `premiere_get_clip_effects` | List all effects applied to a clip with their parameters | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID | Go -> TS |
| 80 | `premiere_set_effect_param` | Set a specific effect parameter value | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `component_index` (int, required): effect component index; `param_index` (int, required): parameter index within the component; `value` (any, required): new parameter value; `update_ui` (bool): refresh the UI (default: true) | Go -> TS |
| 81 | `premiere_get_effect_params` | Get all parameter values for a specific effect component | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `component_index` (int, required): effect component index | Go -> TS |
| 82 | `premiere_copy_effects` | Copy effects from one clip to another | `source_clip_id` (string, required): source clip ID; `target_clip_ids` (string[], required): target clip IDs; `sequence_id` (string, required): sequence ID; `include_intrinsics` (bool): include Motion/Opacity (default: false) | Go -> TS |
| 83 | `premiere_get_effect_list` | List all available video or audio effects | `type` (string): "video", "audio", or "all" (default: "all"); `category` (string): filter by effect category | Go -> TS (QE) |
| 84 | `premiere_create_adjustment_layer` | Create an adjustment layer in the project | `name` (string): adjustment layer name (default: "Adjustment Layer"); `width` (int): frame width (default: sequence width); `height` (int): frame height (default: sequence height); `frame_rate` (float): fps (default: sequence fps); `duration_seconds` (float): duration (default: 10.0) | Go -> TS |

### 4.3 Lumetri Color (via effect parameters)

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 85 | `premiere_lumetri_basic_correction` | Apply/modify basic Lumetri Color correction | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `temperature` (float): white balance temperature; `tint` (float): white balance tint; `exposure` (float): exposure; `contrast` (float): contrast; `highlights` (float): highlights; `shadows` (float): shadows; `whites` (float): whites; `blacks` (float): blacks; `saturation` (float): saturation; `vibrance` (float): vibrance | Go -> TS |
| 86 | `premiere_lumetri_creative` | Apply Lumetri creative look/LUT | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `lut_path` (string): path to .cube or .look file; `intensity` (float): look intensity 0-1; `faded_film` (float): faded film amount; `sharpen` (float): sharpen amount | Go -> TS |
| 87 | `premiere_lumetri_curves` | Modify Lumetri curves (RGB, Hue vs Sat, etc.) | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `curve_type` (string, required): "rgb", "red", "green", "blue", "hue_vs_sat", "hue_vs_hue", "hue_vs_luma", "luma_vs_sat", "sat_vs_sat"; `control_points` (object[], required): array of {x, y} normalized control points | Go -> TS |
| 88 | `premiere_lumetri_color_wheels` | Modify Lumetri color wheels | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `shadow_tint` (object): {x, y} shadow color offset; `midtone_tint` (object): {x, y} midtone color offset; `highlight_tint` (object): {x, y} highlight color offset; `shadow_level` (float); `midtone_level` (float); `highlight_level` (float) | Go -> TS |
| 89 | `premiere_lumetri_vignette` | Apply Lumetri vignette | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `amount` (float, required): vignette amount (-5 to 5); `midpoint` (float): midpoint; `roundness` (float): roundness; `feather` (float): feather | Go -> TS |

### 4.4 Motion & Transform

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 90 | `premiere_set_motion` | Set Motion effect properties (position, scale, rotation) | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `position_x` (float): horizontal position (pixels); `position_y` (float): vertical position (pixels); `scale` (float): uniform scale percentage (100 = 100%); `scale_width` (float): width scale (if non-uniform); `rotation` (float): rotation in degrees; `anchor_x` (float): anchor point X; `anchor_y` (float): anchor point Y; `anti_flicker` (float): anti-flicker filter amount | Go -> TS |
| 91 | `premiere_get_motion` | Get current Motion effect values for a clip | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID | Go -> TS |
| 92 | `premiere_set_opacity` | Set clip opacity and blend mode | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `opacity` (float, required): opacity 0-100; `blend_mode` (string): blend mode name (e.g., "Normal", "Multiply", "Screen", "Overlay", "Add", "Subtract", "Difference") | Go -> TS |
| 93 | `premiere_get_opacity` | Get current opacity and blend mode of a clip | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID | Go -> TS |

### 4.5 Keyframing

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 94 | `premiere_add_keyframe` | Add a keyframe to a component parameter at a time | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `component_index` (int, required): effect/component index; `param_index` (int, required): parameter index; `time_seconds` (float, required): keyframe time relative to clip start; `value` (any): value at this keyframe | Go -> TS |
| 95 | `premiere_remove_keyframe` | Remove a keyframe at a specific time | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `component_index` (int, required): effect/component index; `param_index` (int, required): parameter index; `time_seconds` (float, required): keyframe time | Go -> TS |
| 96 | `premiere_remove_keyframe_range` | Remove all keyframes within a time range | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `component_index` (int, required): effect/component index; `param_index` (int, required): parameter index; `start_seconds` (float, required): range start; `end_seconds` (float, required): range end | Go -> TS |
| 97 | `premiere_get_keyframes` | Get all keyframes for a parameter | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `component_index` (int, required): effect/component index; `param_index` (int, required): parameter index | Go -> TS |
| 98 | `premiere_set_keyframe_value` | Set the value of an existing keyframe | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `component_index` (int, required): effect/component index; `param_index` (int, required): parameter index; `time_seconds` (float, required): keyframe time; `value` (any, required): new value | Go -> TS |
| 99 | `premiere_set_keyframe_interpolation` | Set keyframe interpolation type | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `component_index` (int, required): effect/component index; `param_index` (int, required): parameter index; `time_seconds` (float, required): keyframe time; `interpolation` (string, required): "linear", "bezier", "hold", "ease_in", "ease_out", "ease_in_out" | Go -> TS |
| 100 | `premiere_enable_time_varying` | Enable or disable keyframing (time-varying) on a parameter | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `component_index` (int, required): effect/component index; `param_index` (int, required): parameter index; `enabled` (bool, required): true to enable keyframing | Go -> TS |
| 101 | `premiere_check_keyframes_supported` | Check if a parameter supports keyframing | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `component_index` (int, required): effect/component index; `param_index` (int, required): parameter index | Go -> TS |

---

## Phase 5 -- Audio, Speed, Markers & Playback

Audio control, speed changes, marker management, playback navigation, and in/out points.

### 5.1 Audio

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 102 | `premiere_set_clip_volume` | Set clip audio volume | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `level_db` (float, required): volume in dB (0 = unity); `use_keyframe` (bool): set as keyframe at current position (default: false) | Go -> TS |
| 103 | `premiere_set_audio_gain` | Set audio gain on a project item (source level) | `item_id` (string, required): project item ID; `gain_db` (float, required): gain in dB; `normalize_peak` (bool): normalize to peak (default: false); `normalize_target_db` (float): normalization target (default: -3) | Go -> TS |
| 104 | `premiere_set_audio_channel_mapping` | Set audio channel mapping for a clip | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `channel_type` (string, required): "mono", "stereo", "5.1", "adaptive" | Go -> TS |
| 105 | `premiere_mute_track` | Mute or unmute an audio track | `sequence_id` (string): sequence ID (default: active); `track_index` (int, required): audio track index; `muted` (bool, required): true to mute | Go -> TS |
| 106 | `premiere_solo_track` | Solo an audio track | `sequence_id` (string): sequence ID (default: active); `track_index` (int, required): audio track index; `solo` (bool, required): true to solo | Go -> TS |
| 107 | `premiere_add_audio_keyframe` | Add audio level keyframe at a specific time | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `time_seconds` (float, required): keyframe time; `level_db` (float, required): volume in dB | Go -> TS |
| 108 | `premiere_set_essential_sound` | Tag a clip with Essential Sound panel type | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `sound_type` (string, required): "dialogue", "music", "sfx", "ambience"; `loudness_auto` (bool): enable auto loudness (default: false); `ducking_target` (string): ducking against type | Go -> TS |
| 109 | `premiere_apply_auto_ducking` | Apply auto-ducking to lower music under dialogue | `sequence_id` (string): sequence ID (default: active); `duck_amount_db` (float): how much to duck (default: -12); `fade_duration_seconds` (float): fade in/out duration (default: 0.5); `sensitivity` (float): detection sensitivity 0-1 (default: 0.5) | Go -> Python + TS |

### 5.2 Speed & Time

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 110 | `premiere_set_clip_speed` | Change clip playback speed | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `speed` (float, required): speed multiplier (1.0 = normal, 2.0 = double, 0.5 = half); `ripple` (bool): ripple subsequent clips (default: false); `maintain_pitch` (bool): maintain audio pitch (default: true) | Go -> TS |
| 111 | `premiere_get_clip_speed` | Get the current speed of a clip | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID | Go -> TS |
| 112 | `premiere_time_remap` | Enable time remapping on a clip and set speed keyframes | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `enabled` (bool, required): enable or disable time remapping; `speed_points` (object[]): array of {time_seconds, speed} keyframe pairs | Go -> TS |
| 113 | `premiere_reverse_clip` | Reverse a clip's playback direction | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID | Go -> TS |
| 114 | `premiere_freeze_frame` | Create a freeze frame at a specific time | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `time_seconds` (float, required): frame to freeze at; `duration_seconds` (float): freeze duration (default: 2.0) | Go -> TS |
| 115 | `premiere_speed_ramp` | Create a speed ramp (gradual speed change) on a clip | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `ramp_points` (object[], required): array of {time_seconds, speed, interpolation} (supports "linear" and "bezier"); `ripple` (bool): adjust timeline duration | Go -> TS |

### 5.3 Markers

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 116 | `premiere_add_sequence_marker` | Add a marker to the active sequence | `sequence_id` (string): sequence ID (default: active); `time_seconds` (float, required): marker position; `name` (string): marker name; `comments` (string): marker comments; `duration_seconds` (float): marker duration (0 for point marker); `color_index` (int): color index 0-7; `type` (string): "comment", "chapter", "segmentation", "web_link" (default: "comment") | Go -> TS |
| 117 | `premiere_add_clip_marker` | Add a marker to a project item (source clip) | `item_id` (string, required): project item ID; `time_seconds` (float, required): marker position; `name` (string): marker name; `comments` (string): marker comments; `duration_seconds` (float): marker duration; `color_index` (int): color index | Go -> TS |
| 118 | `premiere_edit_marker` | Edit an existing marker's properties | `sequence_id` (string): sequence or item ID; `marker_index` (int, required): index of the marker; `name` (string): new name; `comments` (string): new comments; `color_index` (int): new color; `duration_seconds` (float): new duration; `type` (string): new marker type | Go -> TS |
| 119 | `premiere_delete_marker` | Delete a marker from a sequence or clip | `sequence_id` (string): sequence or item ID; `marker_index` (int, required): index of the marker to delete | Go -> TS |
| 120 | `premiere_get_markers` | Get all markers from a sequence or clip | `sequence_id` (string): sequence ID (default: active); `item_id` (string): project item ID (if querying clip markers) | Go -> TS |
| 121 | `premiere_navigate_to_marker` | Move playhead to a specific marker | `sequence_id` (string): sequence ID (default: active); `marker_index` (int): go to marker by index; `direction` (string): "next" or "previous" relative to playhead | Go -> TS |
| 122 | `premiere_add_markers_from_regions` | Batch-add markers from silence detection regions | `sequence_id` (string): sequence ID (default: active); `regions` (object[], required): array of {time_seconds, name, color_index} | Go -> Rust + TS |
| 123 | `premiere_export_markers` | Export all markers as JSON, CSV, or EDL | `sequence_id` (string): sequence ID (default: active); `format` (string): "json", "csv", "edl" (default: "json"); `output_path` (string): write to file (otherwise return in response) | Go -> TS |

### 5.4 Playback & Navigation

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 124 | `premiere_get_playhead_position` | Get the current playhead position | `sequence_id` (string): sequence ID (default: active) | Go -> TS |
| 125 | `premiere_set_playhead_position` | Set the playhead to a specific time | `sequence_id` (string): sequence ID (default: active); `time_seconds` (float, required): target time in seconds | Go -> TS |
| 126 | `premiere_step_forward` | Step forward by one frame | `sequence_id` (string): sequence ID (default: active); `frames` (int): number of frames to step (default: 1) | Go -> TS |
| 127 | `premiere_step_backward` | Step backward by one frame | `sequence_id` (string): sequence ID (default: active); `frames` (int): number of frames to step back (default: 1) | Go -> TS |
| 128 | `premiere_go_to_start` | Move playhead to the start of the sequence | `sequence_id` (string): sequence ID (default: active) | Go -> TS |
| 129 | `premiere_go_to_end` | Move playhead to the end of the sequence | `sequence_id` (string): sequence ID (default: active) | Go -> TS |

### 5.5 In/Out Points

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 130 | `premiere_set_in_point` | Set the sequence in point | `sequence_id` (string): sequence ID (default: active); `time_seconds` (float, required): in point time | Go -> TS |
| 131 | `premiere_set_out_point` | Set the sequence out point | `sequence_id` (string): sequence ID (default: active); `time_seconds` (float, required): out point time | Go -> TS |
| 132 | `premiere_clear_in_out` | Clear the sequence in and/or out points | `sequence_id` (string): sequence ID (default: active); `clear` (string): "in", "out", or "both" (default: "both") | Go -> TS |
| 133 | `premiere_get_in_out_points` | Get current in/out point times | `sequence_id` (string): sequence ID (default: active) | Go -> TS |
| 134 | `premiere_mark_clip` | Set in/out points around a clip at the playhead | `sequence_id` (string): sequence ID (default: active); `track_index` (int): target track | Go -> TS |
| 135 | `premiere_mark_selection` | Set in/out points around the current selection | `sequence_id` (string): sequence ID (default: active) | Go -> TS |

---

## Phase 6 -- Titles, Captions, Track Management & Export

Text/graphics, captions, track operations, and advanced export.

### 6.1 Titles & Graphics

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 136 | `premiere_add_mogrt` | Import and place a Motion Graphics Template (.mogrt) | `mogrt_path` (string, required): path to .mogrt file; `sequence_id` (string): sequence ID (default: active); `time_seconds` (float): position on timeline (default: 0); `track_index` (int): video track; `duration_seconds` (float): duration; `video_offset` (int): video track offset; `audio_offset` (int): audio track offset | Go -> TS |
| 137 | `premiere_modify_mogrt_text` | Modify text within a placed Motion Graphics Template | `clip_id` (string, required): clip ID of the mogrt on timeline; `sequence_id` (string, required): sequence ID; `param_name` (string, required): parameter name in the .mogrt; `value` (string, required): new text value | Go -> TS |
| 138 | `premiere_create_text_layer` | Create a text clip (via Essential Graphics .mogrt template) | `sequence_id` (string): sequence ID (default: active); `text` (string, required): text content; `font_family` (string): font name (default: "Arial"); `font_size` (float): size in points (default: 60); `color` (string): hex color (default: "#FFFFFF"); `position_x` (float): normalized X 0-1 (default: 0.5); `position_y` (float): normalized Y 0-1 (default: 0.5); `alignment` (string): "left", "center", "right" (default: "center"); `track_index` (int): video track; `time_seconds` (float): start time; `duration_seconds` (float): duration (default: 5.0) | Go -> TS |
| 139 | `premiere_create_lower_third` | Create a lower-third title from template | `sequence_id` (string): sequence ID (default: active); `title_text` (string, required): main title text; `subtitle_text` (string): subtitle text; `mogrt_path` (string): custom lower-third template path; `time_seconds` (float): start time; `duration_seconds` (float): duration (default: 5.0); `track_index` (int): video track | Go -> TS |
| 140 | `premiere_create_title_card` | Create a full-screen title card from template | `sequence_id` (string): sequence ID (default: active); `title_text` (string, required): title text; `subtitle_text` (string): optional subtitle; `mogrt_path` (string): custom title card template; `time_seconds` (float): start position; `duration_seconds` (float): duration (default: 5.0); `background_color` (string): hex background color | Go -> TS |
| 141 | `premiere_create_shape` | Create a shape layer (solid color mat) | `name` (string): layer name (default: "Color Matte"); `width` (int): width in pixels; `height` (int): height in pixels; `color_r` (int): red 0-255; `color_g` (int): green 0-255; `color_b` (int): blue 0-255 | Go -> TS |
| 142 | `premiere_modify_text_properties` | Modify text properties on an existing text clip | `clip_id` (string, required): clip ID; `sequence_id` (string, required): sequence ID; `text` (string): new text content; `font_family` (string): new font; `font_size` (float): new size; `color` (string): new hex color; `stroke_color` (string): stroke hex color; `stroke_width` (float): stroke width; `shadow_enabled` (bool): enable drop shadow; `shadow_color` (string): shadow hex color; `shadow_opacity` (float): shadow opacity; `shadow_angle` (float): shadow angle; `shadow_distance` (float): shadow distance | Go -> TS |

### 6.2 Captions & Subtitles

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 143 | `premiere_import_captions` | Import SRT, VTT, or SCC caption file | `file_path` (string, required): path to caption file; `sequence_id` (string): sequence ID (default: active); `format` (string): "srt", "vtt", "scc" (auto-detected if omitted) | Go -> TS |
| 144 | `premiere_export_captions` | Export captions to SRT, VTT, or SCC | `sequence_id` (string): sequence ID (default: active); `output_path` (string, required): output file path; `format` (string, required): "srt", "vtt", "scc" | Go -> TS |
| 145 | `premiere_create_caption_track` | Create a new caption track on the sequence | `sequence_id` (string): sequence ID (default: active); `name` (string): track name (default: "Subtitles"); `format` (string): "subtitle", "open_caption" | Go -> TS |
| 146 | `premiere_add_caption` | Add a single caption entry | `sequence_id` (string): sequence ID (default: active); `start_seconds` (float, required): caption start time; `end_seconds` (float, required): caption end time; `text` (string, required): caption text; `track_index` (int): caption track index (default: 0) | Go -> TS |
| 147 | `premiere_edit_caption` | Edit an existing caption's text or timing | `sequence_id` (string): sequence ID (default: active); `caption_index` (int, required): index of caption to edit; `text` (string): new text; `start_seconds` (float): new start time; `end_seconds` (float): new end time | Go -> TS |
| 148 | `premiere_style_captions` | Apply styling to captions (font, size, color, background) | `sequence_id` (string): sequence ID (default: active); `font_family` (string): font name; `font_size` (float): size; `color` (string): hex text color; `background_color` (string): hex background color; `background_opacity` (float): background opacity 0-1; `position` (string): "bottom", "top", "center" | Go -> TS |
| 149 | `premiere_generate_captions_from_audio` | Auto-transcribe audio to captions (via Rust waveform + Python NLP) | `sequence_id` (string): sequence ID (default: active); `language` (string): language code (default: "en"); `max_chars_per_line` (int): max characters per caption line (default: 42); `max_lines` (int): max lines per caption (default: 2) | Go -> Rust + Python + TS |
| 150 | `premiere_burn_in_captions` | Render captions as burned-in text (bake into video) | `sequence_id` (string): sequence ID (default: active); `font_family` (string); `font_size` (float); `color` (string); `position` (string): "bottom", "top", "center" | Go -> TS |

### 6.3 Track Management

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 151 | `premiere_add_track` | Add a new video or audio track to a sequence | `sequence_id` (string): sequence ID (default: active); `track_type` (string, required): "video" or "audio"; `count` (int): number of tracks to add (default: 1) | Go -> TS |
| 152 | `premiere_delete_track` | Delete a track from a sequence | `sequence_id` (string): sequence ID (default: active); `track_type` (string, required): "video" or "audio"; `track_index` (int, required): track to delete | Go -> TS |
| 153 | `premiere_lock_track` | Lock or unlock a track | `sequence_id` (string): sequence ID (default: active); `track_type` (string, required): "video" or "audio"; `track_index` (int, required): track index; `locked` (bool, required): true to lock | Go -> TS |
| 154 | `premiere_mute_video_track` | Toggle track output/visibility for a video track | `sequence_id` (string): sequence ID (default: active); `track_index` (int, required): video track index; `muted` (bool, required): true to mute | Go -> TS |
| 155 | `premiere_set_track_targets` | Set which tracks are targeted for editing operations | `sequence_id` (string): sequence ID (default: active); `video_target` (int): video track target index (-1 to untarget); `audio_target` (int): audio track target index (-1 to untarget) | Go -> TS |
| 156 | `premiere_get_track_info` | Get detailed info about all tracks in a sequence | `sequence_id` (string): sequence ID (default: active) | Go -> TS |

### 6.4 Export & Render

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 157 | `premiere_export_direct` | Export directly from Premiere (synchronous, blocks UI) | `sequence_id` (string): sequence ID (default: active); `output_path` (string, required): output file path; `preset_path` (string, required): path to .epr export preset file; `work_area_type` (int): 0=entire sequence, 1=in-out, 2=work area | Go -> TS |
| 158 | `premiere_export_via_encoder` | Send to Adobe Media Encoder queue | `sequence_id` (string): sequence ID (default: active); `output_path` (string, required): output file path; `preset_path` (string, required): path to .epr preset; `remove_on_complete` (bool): remove from queue after encode (default: true) | Go -> TS |
| 159 | `premiere_start_encoder_batch` | Start the Adobe Media Encoder render queue | (none) | Go -> TS |
| 160 | `premiere_export_frame` | Export a single frame as an image | `sequence_id` (string): sequence ID (default: active); `time_seconds` (float): frame time (default: playhead position); `output_path` (string, required): output image path (.png, .jpg, .tiff); `quality` (int): JPEG quality 1-100 (default: 100) | Go -> TS |
| 161 | `premiere_export_audio_only` | Export only the audio from a sequence | `sequence_id` (string): sequence ID (default: active); `output_path` (string, required): output file path; `format` (string): "wav", "aac", "mp3" (default: "wav"); `sample_rate` (int): sample rate (default: 48000); `bit_depth` (int): bit depth (default: 24) | Go -> TS |
| 162 | `premiere_get_export_presets` | List available export presets | `format` (string): filter by format ("h264", "prores", "dnxhr", etc.) | Go -> TS |
| 163 | `premiere_batch_export` | Export multiple sequences in a batch | `exports` (object[], required): array of {sequence_id, output_path, preset}; `use_encoder` (bool): use Media Encoder (default: true) | Go -> TS |
| 164 | `premiere_export_with_settings` | Export with custom bitrate, resolution, and frame rate overrides | `sequence_id` (string): sequence ID (default: active); `output_path` (string, required): output file path; `preset` (string, required): base preset name; `width` (int): override width; `height` (int): override height; `frame_rate` (float): override fps; `bitrate_mbps` (float): target bitrate in Mbps; `max_bitrate_mbps` (float): max bitrate (VBR); `encoding_mode` (string): "cbr", "vbr1", "vbr2" | Go -> TS |

---

## Phase 7 -- Multicam, Auto Features, Workspace & Preferences

Advanced workflows, AI-powered features, and application settings.

### 7.1 Multicam

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 165 | `premiere_create_multicam_sequence` | Create a multi-camera source sequence | `name` (string, required): sequence name; `clip_ids` (string[], required): array of project item IDs for camera angles; `sync_method` (string): "timecode", "audio", "in_point", "clip_marker" (default: "audio"); `audio_sequence_settings` (object): audio channel preset | Go -> TS |
| 166 | `premiere_switch_multicam_angle` | Switch the active camera angle at the playhead | `sequence_id` (string, required): multicam sequence ID; `angle_index` (int, required): camera angle index (0-based); `time_seconds` (float): time to switch at (default: playhead); `cut_type` (string): "cut", "dissolve" (default: "cut") | Go -> TS |
| 167 | `premiere_flatten_multicam` | Flatten a multicam sequence into a standard sequence | `sequence_id` (string, required): multicam sequence ID | Go -> TS |
| 168 | `premiere_get_multicam_angles` | List all camera angles in a multicam sequence | `sequence_id` (string, required): multicam sequence ID | Go -> TS |

### 7.2 AI-Powered Auto Features

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 169 | `premiere_auto_reframe` | Automatically reframe a sequence for a different aspect ratio | `sequence_id` (string): sequence ID (default: active); `target_aspect` (string, required): target aspect ratio ("9:16", "1:1", "4:5", "16:9"); `motion_tracking` (string): "default", "slower", "faster" (default: "default") | Go -> TS |
| 170 | `premiere_scene_edit_detection` | Detect scene edits/cuts in a clip and add markers or cut points | `item_id` (string, required): project item ID of the source clip; `sensitivity` (float): detection sensitivity 0-1 (default: 0.5); `action` (string): "markers" or "cuts" (default: "markers") | Go -> Rust + TS |
| 171 | `premiere_auto_color_match` | Automatically match color between clips | `reference_clip_id` (string, required): reference clip ID; `target_clip_ids` (string[], required): clips to color match; `sequence_id` (string, required): sequence ID; `intensity` (float): match intensity 0-1 (default: 0.8) | Go -> TS |
| 172 | `premiere_detect_silence` | Detect silence in audio clips and optionally remove | `sequence_id` (string): sequence ID (default: active); `threshold_db` (float): silence threshold (default: -40); `min_duration_seconds` (float): minimum silence duration (default: 0.5); `action` (string): "detect", "mark", "ripple_delete" (default: "detect") | Go -> Rust + TS |
| 173 | `premiere_auto_levels` | Automatically normalize audio levels across all clips | `sequence_id` (string): sequence ID (default: active); `target_db` (float): target loudness (default: -23 LUFS); `standard` (string): "ebu_r128", "atsc_a85" (default: "ebu_r128") | Go -> Rust + TS |
| 174 | `premiere_smart_trim` | AI-powered smart trim -- remove dead air, ums, and pauses | `sequence_id` (string): sequence ID (default: active); `remove_silence` (bool): remove silent gaps (default: true); `remove_filler_words` (bool): remove um/uh/like (default: false); `padding_seconds` (float): breathing room around cuts (default: 0.1) | Go -> Rust + Python + TS |
| 175 | `premiere_generate_edl` | Generate an EDL from matched script segments and assets | `segments` (object[], required): parsed script segments; `assets` (object[], required): scanned asset metadata; `settings` (object): EDL settings (resolution, fps, pacing, transitions) | Go -> Python |
| 176 | `premiere_match_assets` | Match script segments to available media assets via AI | `segments` (object[], required): parsed script segments; `assets` (object[], required): scanned assets; `strategy` (string): "filename", "embedding", "hybrid" (default: "hybrid") | Go -> Python |
| 177 | `premiere_analyze_pacing` | Analyze the pacing of an edit and suggest adjustments | `edl` (object, required): edit decision list; `target_mood` (string): "energetic", "calm", "dramatic", "comedic" | Go -> Python |

### 7.3 Source Monitor

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 178 | `premiere_open_in_source_monitor` | Open a project item in the Source Monitor | `item_id` (string, required): project item ID | Go -> TS |
| 179 | `premiere_get_source_monitor_clip` | Get info about the clip currently in the Source Monitor | (none) | Go -> TS |

### 7.4 Workspace

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 180 | `premiere_save_workspace` | Save the current workspace layout | `name` (string, required): workspace name | Go -> TS |
| 181 | `premiere_load_workspace` | Load a saved workspace layout | `name` (string, required): workspace name (e.g., "Editing", "Color", "Audio", "Effects") | Go -> TS |
| 182 | `premiere_reset_workspace` | Reset the current workspace to its default layout | `name` (string): workspace name (default: current) | Go -> TS |

### 7.5 Project Panel Views

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 183 | `premiere_sort_project_items` | Sort items in the Project Panel | `sort_by` (string, required): "name", "type", "date_created", "date_modified", "duration", "frame_rate"; `ascending` (bool): sort direction (default: true) | Go -> TS |
| 184 | `premiere_search_project_panel` | Search the Project Panel with filters | `query` (string, required): search query; `filter` (string): "all", "video", "audio", "image", "sequence", "bin" (default: "all") | Go -> TS |

### 7.6 Undo/Redo

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 185 | `premiere_undo` | Undo the last operation | `steps` (int): number of undos (default: 1) | Go -> TS |
| 186 | `premiere_redo` | Redo the last undone operation | `steps` (int): number of redos (default: 1) | Go -> TS |

### 7.7 Clipboard & Selection

| # | Tool Name | Description | Parameters | Language |
|---|---|---|---|---|
| 187 | `premiere_get_selection` | Get the current clip selection in the active sequence | `sequence_id` (string): sequence ID (default: active) | Go -> TS |
| 188 | `premiere_select_clips` | Select specific clips on the timeline | `clip_ids` (string[], required): clip IDs to select; `sequence_id` (string, required): sequence ID; `add_to_selection` (bool): add to current selection (default: false) | Go -> TS |
| 189 | `premiere_deselect_all` | Deselect all clips in the active sequence | `sequence_id` (string): sequence ID (default: active) | Go -> TS |

---

## Phase Summary

| Phase | Focus Area | New Tools | Cumulative |
|---|---|---|---|
| 0 | Foundation (implemented) | 17 | 17 |
| 1 | Project Management & Media Management | 30 | 47 |
| 2 | Sequence & Timeline Operations | 16 | 63 |
| 3 | Clip Operations & Trimming | 23 | 86 |
| 4 | Effects, Transitions & Motion | 32 | 118 |
| 5 | Audio, Speed, Markers & Playback | 34 | 152 |
| 6 | Titles, Captions, Track Management & Export | 29 | 181 |
| 7 | Multicam, Auto Features, Workspace & Preferences | 25 | 206 |

**Total: 206 MCP tools** covering the complete Premiere Pro automation surface.

---

## Implementation Notes

### ExtendScript API vs QE DOM

| API Surface | Access Method | Stability |
|---|---|---|
| Standard ExtendScript DOM | `app.project`, `sequence`, `trackItem`, `component`, `componentParam`, `marker` | Stable, documented |
| QE DOM (undocumented) | `app.enableQE()`, `qe.project`, `qeSeq.getVideoTrackAt()` | Unstable, undocumented, but required for transitions and effects |

Tools that require the QE DOM are marked with `(QE)` in the Language column. These include:
- Applying transitions (`qeClip.addTransition()`)
- Applying effects by name (`qe.project.getVideoEffectByName()`, `qeClip.addVideoEffect()`)
- Creating sequences from presets (`qe.project.newSequence()`)
- Razor tool operations
- Listing available effects and transitions

### Time System

All tools accept time in **seconds** (float) at the MCP interface level. The TypeScript bridge converts to Premiere's internal tick system (254,016,000,000 ticks per second) or frame-based timecodes as needed.

### Language Routing

Each tool call follows this flow:

```
MCP Client (Claude) -> Go Orchestrator -> Service Router
                                           |
                     +---------------------+---------------------+
                     |                     |                     |
                 Rust Engine          Python Intel          TS Bridge
              (media analysis)    (AI/NLP decisions)    (Premiere Pro)
```

- **Go** handles parameter validation, error wrapping, retry logic, and response formatting for every tool.
- **TS** handles all direct Premiere Pro interaction via ExtendScript/CEP.
- **Rust** handles file-system-level media analysis (probing, waveforms, scene detection, thumbnails).
- **Python** handles AI-driven operations (script parsing, asset matching, EDL generation, pacing analysis, filler word detection).

### Error Handling Strategy

Every tool returns one of:
- Success JSON with result data
- Error with human-readable message and error code
- Partial success with warnings array (for batch operations)

The Go orchestrator wraps all downstream errors with context and applies circuit-breaker patterns for the gRPC connections to each service.

### Limitations & Workarounds

| Limitation | Impact | Workaround |
|---|---|---|
| No `setSpeed()` in standard DOM | Cannot programmatically change clip speed via documented API | QE DOM `changeSpeed()` or manipulate clip in/out + duration |
| No apply-effect-by-name in standard DOM | Must use QE DOM for applying new effects | `qe.project.getVideoEffectByName()` + `qeClip.addVideoEffect()` |
| No create-text-from-scratch | Cannot generate text layers without a template | Use bundled .mogrt templates, modify via `getMGTComponent()` |
| No undo grouping | Each scripted operation is a separate undo step | Document this limitation; consider batching in single evalScript calls |
| Sync execution blocks Premiere UI | Long scripts freeze the app | Split into small atomic operations with `$.sleep()` between them |
| No caption API in ExtendScript | Caption creation/editing not in standard DOM | Import .srt files; use UI automation as last resort |
| ExtendScript EOL September 2026 | API will eventually be replaced by UXP | Architecture supports future UXP bridge module swap |

---

## Appendix A: ExtendScript Object Model Quick Reference

```
app (Application)
  .project (Project)
    .name, .path
    .rootItem (ProjectItem, the root bin)
      .children[] -> ProjectItem
        .name, .treePath, .type (CLIP, BIN, ROOT, FILE)
        .getMarkers() -> MarkerCollection
        .getProjectMetadata() / .setProjectMetadata()
        .getXMPMetadata() / .setXMPMetadata()
        .attachProxy() / .canProxy()
        .changeMediaPath() / .canChangeMediaPath()
        .setScaleToFrameSize()
        .findItemsMatchingMediaPath()
        .moveBin() / .renameBin()
        .createBin() / .createSmartBin()
    .sequences[] -> SequenceCollection -> Sequence
    .activeSequence -> Sequence
      .name, .sequenceID, .id
      .videoTracks[] -> Track
        .clips[] -> TrackItem
          .name, .start, .end, .inPoint, .outPoint (Time)
          .duration, .type, .mediaType
          .components[] -> Component
            .displayName, .matchName
            .properties[] -> ComponentParam
              .displayName, .getValue(), .setValue()
              .addKey(), .removeKey(), .getKeys()
              .setTimeVarying(), .isTimeVarying()
              .areKeyframesSupported()
              .setInterpolationTypeAtKey()
              .getValueAtKey(), .setValueAtKey()
              .findNearestKey(), .findNextKey(), .findPreviousKey()
              .removeKeyRange(), .keyExistsAtTime()
          .remove(inRipple, inAlignToVideo)
          .disabled (read/write)
          .getSpeed()
          .getMGTComponent() -> Component (for .mogrt clips)
          .projectItem -> ProjectItem
        .isMuted() / .setMute()
        .isLocked() / .setLocked()
        .insertClip() / .overwriteClip()
      .audioTracks[] -> Track (same structure)
      .markers -> MarkerCollection
        .createMarker(time) -> Marker
          .name, .comments, .end, .type
          .setTypeAsComment() / .setTypeAsChapter()
          .getColorByIndex() / .setColorByIndex()
      .getSettings() / .setSettings()
      .insertClip() / .overwriteClip()
      .importMGT()
      .clone()
      .getPlayerPosition() / .setPlayerPosition()
      .setInPoint() / .setOutPoint()
      .getInPointAsTime() / .getOutPointAsTime()
      .setWorkAreaInPoint() / .setWorkAreaOutPoint()
      .setZeroPoint()
      .exportAsMediaDirect()
      .getSelection()
    .createNewSequence()
    .createNewSequenceFromClips()
    .importFiles()
    .importSequences()
    .save() / .saveAs()
    .closeDocument()
    .addPropertyToProjectMetadataSchema()
  .encoder (Encoder)
    .encodeSequence()
    .startBatch()
  .sourceMonitor (SourceMonitor)
    .openFilePath()
  .enableQE() -> unlocks qe.* namespace
```

## Appendix B: QE DOM Quick Reference

```
qe (QE Application, after app.enableQE())
  .project
    .getActiveSequence() -> QESequence
      .getVideoTrackAt(index) -> QETrack
        .getItemAt(index) -> QETrackItem
          .addTransition(transition, applyToEnd, duration)
          .addVideoEffect(effect)
          .addAudioEffect(effect)
          .removeTransition(fromEnd)
          .changeSpeed(speedStr, boolStr, boolStr, ...)
      .getAudioTrackAt(index) -> QETrack
      .razor(timeTicks)
      .addTracks(numVideo, videoType, numAudio, audioType)
    .getVideoEffectByName(name) -> QEEffect
    .getVideoTransitionByName(name) -> QETransition
    .getAudioEffectByName(name) -> QEEffect
    .getAudioTransitionByName(name) -> QETransition
    .newSequence(name, presetPath)
    .getVideoEffectList() -> string (pipe-delimited)
    .getAudioEffectList() -> string
    .getVideoTransitionList() -> string
    .getAudioTransitionList() -> string
```

## Appendix C: Common Effect Component Indices

For a standard video clip on the timeline:

| Component Index | Name | Key Properties |
|---|---|---|
| 0 | Opacity | Opacity (0-100), Blend Mode |
| 1 | Motion | Position [x,y], Scale, Scale Width, Rotation, Anchor Point, Anti-flicker Filter |
| 2+ | Applied Effects | Varies per effect (Lumetri Color, Gaussian Blur, etc.) |

For a standard audio clip:

| Component Index | Name | Key Properties |
|---|---|---|
| 0 | Volume | Level (linear scale, not dB -- convert with `Math.pow(10, dB/20)`) |
| 1 | Channel Volume | Per-channel levels |
| 2 | Panner | Pan value |
| 3+ | Applied Effects | Varies per effect |
