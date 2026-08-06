package mcp

import (
	"os"
	"path"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// Tool profiles let clients load a workflow-sized tool catalog instead of the
// complete low-level surface. Every specialized profile includes core tools.
// The default is the compact "standard" profile.
var toolProfilePatterns = map[string][]string{
	"core": {
		"premiere_ping",
		"premiere_is_running",
		"premiere_get_premiere_version",
		"premiere_get_project",
		"premiere_get_timeline",
		"premiere_get_sequence_list",
		"premiere_get_sequence_hash",
		"premiere_get_state_snapshot",
		"premiere_snapshot_timeline",
		"premiere_save_sequence_version",
		"premiere_duplicate_sequence",
		"premiere_set_active_sequence",
		"premiere_save_project",
	},
	"standard": {
		"premiere_open",
		"premiere_scan_assets",
		"premiere_probe_media",
		"premiere_generate_thumbnail",
		"premiere_analyze_waveform",
		"premiere_detect_scenes",
		"premiere_import_files",
		"premiere_import_media",
		"premiere_parse_script",
		"premiere_get_project_items",
		"premiere_find_project_items",
		"premiere_get_offline_items",
		"premiere_relink_media",
		"premiere_create_sequence",
		"premiere_duplicate_sequence",
		"premiere_rename_sequence",
		"premiere_set_active_sequence",
		"premiere_get_video_tracks",
		"premiere_get_audio_tracks",
		"premiere_get_clips_on_track",
		"premiere_get_clip_info",
		"premiere_get_playhead_position",
		"premiere_set_playhead_position",
		"premiere_insert_clip",
		"premiere_overwrite_clip",
		"premiere_place_clip",
		"premiere_remove_clip",
		"premiere_remove_clip_from_track",
		"premiere_move_clip",
		"premiere_trim_clip_start",
		"premiere_trim_clip_end",
		"premiere_razor_clip",
		"premiere_ripple_trim",
		"premiere_ripple_delete_gap",
		"premiere_add_sequence_marker",
		"premiere_get_sequence_markers",
		"premiere_add_video_transition",
		"premiere_add_audio_transition",
		"premiere_get_installed_transitions",
		"premiere_get_transitions",
		"premiere_get_installed_effects",
		"premiere_apply_video_effect",
		"premiere_apply_audio_effect",
		"premiere_get_clip_effects",
		"premiere_set_effect_parameter",
		"premiere_get_audio_mixer_state",
		"premiere_set_audio_level",
		"premiere_get_audio_level",
		"premiere_normalize_audio",
		"premiere_add_subtitles_from_srt",
		"premiere_get_captions",
		"premiere_export_captions",
		"premiere_validate_closed_captions",
		"premiere_import_mogrt",
		"premiere_get_mogrt_properties",
		"premiere_set_mogrt_text",
		"premiere_set_mogrt_property",
		"premiere_export",
		"premiere_export_direct",
		"premiere_export_via_ame",
		"premiere_list_export_presets_disk",
	},
	"dialogue": {
		"premiere_analyze_waveform",
		"premiere_probe_media",
		"premiere_*trim*",
		"premiere_*razor*",
		"premiere_*gap*",
		"premiere_add_audio_transition",
		"premiere_*audio_level*",
		"premiere_remove_clip*",
		"premiere_get_clips_on_track",
		"premiere_get_clip_info",
	},
	"captions": {
		"premiere_add_subtitles_from_srt",
		"premiere_get_captions",
		"premiere_export_captions",
		"premiere_validate_closed_captions",
	},
	"social": {
		"premiere_auto_reframe",
		"premiere_create_vertical_version",
		"premiere_create_square_version",
		"premiere_duplicate_sequence",
		"premiere_get_sequence_markers",
		"premiere_get_captions",
		"premiere_set_position",
		"premiere_set_scale",
		"premiere_export_direct",
		"premiere_export_via_ame",
		"premiere_list_export_presets_disk",
		"premiere_probe_media",
	},
	"transitions": {
		"premiere_get_available_transitions",
		"premiere_get_installed_transitions",
		"premiere_get_transitions",
		"premiere_get_transition_properties",
		"premiere_get_duration_of_transition",
		"premiere_add_video_transition",
		"premiere_add_audio_transition",
		"premiere_remove_transition",
	},
	"effects": {
		"premiere_get_installed_effects",
		"premiere_get_clip_effects",
		"premiere_apply_video_effect",
		"premiere_apply_audio_effect",
		"premiere_set_effect_parameter",
	},
	"proxies": {
		"premiere_*proxy*",
		"premiere_toggle_proxies",
		"premiere_get_project_items",
		"premiere_get_media_path",
		"premiere_get_media_info",
		"premiere_get_offline_items",
		"premiere_probe_media",
	},
	"delivery": {
		"premiere_export_direct",
		"premiere_export_via_ame",
		"premiere_list_export_presets_disk",
		"premiere_export_captions",
		"premiere_validate_closed_captions",
		"premiere_probe_media",
	},
	"unsafe": {
		"premiere_evaluate_expression",
		"premiere_execute_script",
		"premiere_execute_script_with_args",
		"premiere_execute_extendscript",
		"premiere_execute_qe_script",
		"premiere_run_extend_script",
		"premiere_if_clip_exists",
		"premiere_if_sequence_open",
		"premiere_if_project_open",
		"premiere_while_condition",
		"premiere_execute_batch",
		"premiere_execute_parallel",
		"premiere_execute_with_retry",
		"premiere_execute_with_timeout",
		"premiere_schedule_script",
		"premiere_schedule_repeating",
		"premiere_list_available_scripts",
		"premiere_execute_system_command",
		"premiere_read_json_file",
		"premiere_write_json_file",
		"premiere_read_csv_file",
		"premiere_write_csv_file",
		"premiere_read_text_file",
		"premiere_write_text_file",
		"premiere_append_text_file",
		"premiere_move_media_file",
		"premiere_copy_media_file",
		"premiere_rename_media_file",
		"premiere_open_url",
		"premiere_copy_to_clipboard",
		"premiere_open_in_external_editor",
	},
}

func applyToolProfiles(s *server.MCPServer, logger *zap.Logger) {
	raw := strings.TrimSpace(os.Getenv("MCP_TOOL_PROFILE"))
	if raw == "" {
		raw = "standard"
	}

	selected := []string{"core"}
	unknown := make([]string, 0)
	explicitCore := false
	allSelected := false
	unsafeSelected := false
	for _, value := range strings.Split(raw, ",") {
		profile := strings.ToLower(strings.TrimSpace(value))
		if profile == "" {
			continue
		}
		if profile == "core" {
			explicitCore = true
			continue
		}
		if profile == "all" {
			allSelected = true
			continue
		}
		if profile == "unsafe" {
			unsafeSelected = true
		}
		if _, ok := toolProfilePatterns[profile]; !ok {
			unknown = append(unknown, profile)
			continue
		}
		selected = append(selected, profile)
	}

	if len(unknown) > 0 {
		sort.Strings(unknown)
		logger.Warn("ignoring unknown MCP tool profiles", zap.Strings("profiles", unknown))
	}
	if len(selected) == 1 && len(unknown) > 0 && !explicitCore && !unsafeSelected {
		// A typo should not unexpectedly reduce the default catalog to core.
		selected = append(selected, "standard")
	}

	if allSelected {
		if unsafeSelected {
			return
		}
		removeToolsMatching(s, toolProfilePatterns["unsafe"])
		logger.Info("applied safe all-tools profile", zap.Int("tools", len(s.ListTools())))
		return
	}

	patterns := make([]string, 0)
	for _, profile := range selected {
		patterns = append(patterns, toolProfilePatterns[profile]...)
	}

	tools := s.ListTools()
	remove := make([]string, 0, len(tools))
	for name := range tools {
		if !matchesAnyToolPattern(name, patterns) {
			remove = append(remove, name)
		}
	}
	s.DeleteTools(remove...)
	logger.Info("applied MCP tool profiles",
		zap.Strings("profiles", selected),
		zap.Int("tools", len(tools)-len(remove)),
	)
}

func removeToolsMatching(s *server.MCPServer, patterns []string) {
	remove := make([]string, 0)
	for name := range s.ListTools() {
		if matchesAnyToolPattern(name, patterns) {
			remove = append(remove, name)
		}
	}
	s.DeleteTools(remove...)
}

func matchesAnyToolPattern(name string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := path.Match(pattern, name)
		if err == nil && matched {
			return true
		}
	}
	return false
}
