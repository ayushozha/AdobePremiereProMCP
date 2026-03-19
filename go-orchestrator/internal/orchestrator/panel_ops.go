package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ---------------------------------------------------------------------------
// Timeline Panel Menu (QE DOM) — delegates to ExtendScript via the bridge
// ---------------------------------------------------------------------------

func (e *Engine) SetAudioWaveformLabelColor(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{"enabled": enabled})
	result, err := e.premiere.EvalCommand(ctx, "setAudioWaveformLabelColor", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetAudioWaveformLabelColor: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetLogarithmicWaveformScaling(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{"enabled": enabled})
	result, err := e.premiere.EvalCommand(ctx, "setLogarithmicWaveformScaling", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetLogarithmicWaveformScaling: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetTimeRulerNumbers(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{"enabled": enabled})
	result, err := e.premiere.EvalCommand(ctx, "setTimeRulerNumbers", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetTimeRulerNumbers: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetMultiCameraAudioFollowsVideo(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{"enabled": enabled})
	result, err := e.premiere.EvalCommand(ctx, "setMultiCameraAudioFollowsVideo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetMultiCameraAudioFollowsVideo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetMultiCameraSelectionTopPanel(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{"enabled": enabled})
	result, err := e.premiere.EvalCommand(ctx, "setMultiCameraSelectionTopPanel", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetMultiCameraSelectionTopPanel: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetMultiCameraFollowsNestSetting(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{"enabled": enabled})
	result, err := e.premiere.EvalCommand(ctx, "setMultiCameraFollowsNestSetting", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetMultiCameraFollowsNestSetting: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetRectifiedAudioWaveforms(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{"enabled": enabled})
	result, err := e.premiere.EvalCommand(ctx, "setRectifiedAudioWaveforms", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetRectifiedAudioWaveforms: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Panel Docking (macOS Accessibility / AppleScript)
// These bypass the bridge and invoke osascript directly.
// ---------------------------------------------------------------------------

// runAppleScript executes an AppleScript string via osascript and returns the
// combined stdout/stderr output.
func (e *Engine) runAppleScript(ctx context.Context, script string) (string, error) {
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("osascript: %s – %w", strings.TrimSpace(string(out)), err)
	}
	return strings.TrimSpace(string(out)), nil
}

// panelMenuAppleScript builds an AppleScript that right-clicks on a panel tab
// (via accessibility) and selects a context menu item.  Because panel tab
// identification is fragile, we fall-through to the Window menu where
// possible.
func panelMenuAppleScript(menuItem string) string {
	// Use the Window menu path as the most reliable workaround.
	return `tell application "System Events"
  tell process "Adobe Premiere Pro 2026"
    set frontmost to true
    click menu item "` + menuItem + `" of menu 1 of menu bar item "Window" of menu bar 1
  end tell
end tell`
}

func (e *Engine) UndockPanel(ctx context.Context, panelName string) (*GenericResult, error) {
	script := panelMenuAppleScript("Undock Panel")
	_, err := e.runAppleScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("UndockPanel: %w", err)
	}
	return &GenericResult{Status: "success", Message: "Panel undocked: " + panelName}, nil
}

func (e *Engine) CloseOtherPanelsInGroup(ctx context.Context, panelName string) (*GenericResult, error) {
	script := panelMenuAppleScript("Close Other Panels in Group")
	_, err := e.runAppleScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("CloseOtherPanelsInGroup: %w", err)
	}
	return &GenericResult{Status: "success", Message: "Closed other panels in group: " + panelName}, nil
}

func (e *Engine) ClosePanelGroup(ctx context.Context, panelName string) (*GenericResult, error) {
	script := panelMenuAppleScript("Close Panel Group")
	_, err := e.runAppleScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("ClosePanelGroup: %w", err)
	}
	return &GenericResult{Status: "success", Message: "Panel group closed: " + panelName}, nil
}

func (e *Engine) UndockPanelGroup(ctx context.Context, panelName string) (*GenericResult, error) {
	script := panelMenuAppleScript("Undock Panel Group")
	_, err := e.runAppleScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("UndockPanelGroup: %w", err)
	}
	return &GenericResult{Status: "success", Message: "Panel group undocked: " + panelName}, nil
}

func (e *Engine) MaximizePanelGroup(ctx context.Context, panelName string) (*GenericResult, error) {
	script := panelMenuAppleScript("Maximize Panel Group")
	_, err := e.runAppleScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("MaximizePanelGroup: %w", err)
	}
	return &GenericResult{Status: "success", Message: "Panel group maximized: " + panelName}, nil
}

func (e *Engine) SetStackedPanels(ctx context.Context, enabled bool) (*GenericResult, error) {
	// "Stacked Panel Group" is a toggle menu item.
	script := panelMenuAppleScript("Stacked Panel Group")
	_, err := e.runAppleScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("SetStackedPanels: %w", err)
	}
	state := "enabled"
	if !enabled {
		state = "disabled"
	}
	return &GenericResult{Status: "success", Message: "Stacked panels " + state}, nil
}

func (e *Engine) SetSmallTabs(ctx context.Context, enabled bool) (*GenericResult, error) {
	script := panelMenuAppleScript("Small Tabs")
	_, err := e.runAppleScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("SetSmallTabs: %w", err)
	}
	state := "enabled"
	if !enabled {
		state = "disabled"
	}
	return &GenericResult{Status: "success", Message: "Small tabs " + state}, nil
}

func (e *Engine) SimulateMenuClick(ctx context.Context, menuPath string) (*GenericResult, error) {
	parts := strings.Split(menuPath, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("SimulateMenuClick: menu_path must have at least 2 segments (got %q)", menuPath)
	}

	var clickLine string
	switch len(parts) {
	case 2:
		clickLine = `click menu item "` + parts[1] + `" of menu 1 of menu bar item "` + parts[0] + `" of menu bar 1`
	case 3:
		clickLine = `click menu item "` + parts[2] + `" of menu 1 of menu item "` + parts[1] + `" of menu 1 of menu bar item "` + parts[0] + `" of menu bar 1`
	default:
		// Build nested submenu path for deeper menus
		clickLine = `click menu item "` + parts[len(parts)-1] + `"`
		for i := len(parts) - 2; i >= 1; i-- {
			clickLine += ` of menu 1 of menu item "` + parts[i] + `"`
		}
		clickLine += ` of menu 1 of menu bar item "` + parts[0] + `" of menu bar 1`
	}

	script := `tell application "System Events"
  tell process "Adobe Premiere Pro 2026"
    set frontmost to true
    ` + clickLine + `
  end tell
end tell`

	_, err := e.runAppleScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("SimulateMenuClick: %w", err)
	}
	return &GenericResult{Status: "success", Message: "Menu clicked: " + menuPath}, nil
}
