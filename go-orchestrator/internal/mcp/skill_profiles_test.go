package mcp

import (
	"testing"

	"go.uber.org/zap"
)

func TestWorkflowSkillProfilesExposeRequiredTools(t *testing.T) {
	tests := map[string][]string{
		"dialogue": {
			"premiere_ping",
			"premiere_get_project",
			"premiere_get_timeline",
			"premiere_get_sequence_list",
			"premiere_duplicate_sequence",
			"premiere_set_active_sequence",
			"premiere_get_sequence_hash",
			"premiere_get_clips_on_track",
			"premiere_get_clip_info",
			"premiere_analyze_waveform",
			"premiere_razor_clip",
			"premiere_remove_clip_from_track",
			"premiere_ripple_delete_gap",
			"premiere_ripple_trim",
			"premiere_add_audio_transition",
		},
		"captions": {
			"premiere_ping",
			"premiere_get_project",
			"premiere_get_timeline",
			"premiere_get_sequence_list",
			"premiere_duplicate_sequence",
			"premiere_set_active_sequence",
			"premiere_add_subtitles_from_srt",
			"premiere_get_captions",
			"premiere_validate_closed_captions",
			"premiere_export_captions",
		},
		"social": {
			"premiere_get_sequence_list",
			"premiere_create_vertical_version",
			"premiere_create_square_version",
			"premiere_auto_reframe",
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
			"premiere_get_sequence_list",
			"premiere_duplicate_sequence",
			"premiere_set_active_sequence",
			"premiere_get_available_transitions",
			"premiere_get_installed_transitions",
			"premiere_get_transitions",
			"premiere_add_video_transition",
			"premiere_add_audio_transition",
			"premiere_remove_transition",
		},
		"effects": {
			"premiere_get_sequence_list",
			"premiere_duplicate_sequence",
			"premiere_set_active_sequence",
			"premiere_get_sequence_hash",
			"premiere_get_installed_effects",
			"premiere_get_clip_effects",
			"premiere_apply_video_effect",
			"premiere_set_effect_parameter",
		},
		"proxies": {
			"premiere_get_project_items",
			"premiere_get_media_info",
			"premiere_get_offline_items",
			"premiere_create_proxy",
			"premiere_attach_proxy",
			"premiere_has_proxy",
			"premiere_get_proxy_path",
			"premiere_toggle_proxies",
			"premiere_probe_media",
		},
		"delivery": {
			"premiere_get_sequence_list",
			"premiere_set_active_sequence",
			"premiere_export_direct",
			"premiere_export_via_ame",
			"premiere_list_export_presets_disk",
			"premiere_validate_closed_captions",
			"premiere_export_captions",
			"premiere_probe_media",
		},
	}

	for profile, requiredTools := range tests {
		t.Run(profile, func(t *testing.T) {
			t.Setenv("MCP_TOOL_PROFILE", profile)
			tools := NewMCPServer(nil, "test", zap.NewNop()).ListTools()
			for _, name := range requiredTools {
				if tools[name] == nil {
					t.Errorf("MCP_TOOL_PROFILE=%s omits skill dependency %q", profile, name)
				}
			}
		})
	}
}

func TestStandardProfileSkillCoverageClaims(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "standard")
	tools := NewMCPServer(nil, "test", zap.NewNop()).ListTools()

	for _, name := range []string{
		"premiere_analyze_waveform",
		"premiere_get_clips_on_track",
		"premiere_razor_clip",
		"premiere_ripple_delete_gap",
		"premiere_add_subtitles_from_srt",
		"premiere_get_captions",
		"premiere_validate_closed_captions",
		"premiere_export_captions",
	} {
		if tools[name] == nil {
			t.Errorf("standard profile omits documented dialogue/captions dependency %q", name)
		}
	}

	for workflow, specializedTool := range map[string]string{
		"social":  "premiere_auto_reframe",
		"proxies": "premiere_create_proxy",
	} {
		if tools[specializedTool] != nil {
			t.Errorf("standard profile unexpectedly exposes %s-only dependency %q", workflow, specializedTool)
		}
	}
}

func TestVerifiedWorkflowProfilesOmitUnverifiableLegacyTools(t *testing.T) {
	tests := map[string][]string{
		"social": {
			"premiere_add_safe_zone_guides",
			"premiere_create_social_cuts",
		},
		"effects": {
			"premiere_enable_effect",
			"premiere_apply_vignette",
		},
		"transitions": {
			"premiere_add_audio_crossfade",
		},
	}

	for profile, omittedTools := range tests {
		t.Run(profile, func(t *testing.T) {
			t.Setenv("MCP_TOOL_PROFILE", profile)
			tools := NewMCPServer(nil, "test", zap.NewNop()).ListTools()
			for _, name := range omittedTools {
				if tools[name] != nil {
					t.Errorf("MCP_TOOL_PROFILE=%s exposes unverifiable legacy tool %q", profile, name)
				}
			}
		})
	}
}
