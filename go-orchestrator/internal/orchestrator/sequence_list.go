package orchestrator

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// Normalize complete older camelCase host responses without changing the
// public snake_case schema. Reject unreadable inventories instead of zeros.
func decodeSequenceList(raw string, out *SequenceListResult) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &object); err != nil {
		return err
	}
	if object == nil {
		return fmt.Errorf("expected sequence-list object")
	}
	if value, present := object["success"]; present {
		var success bool
		if err := json.Unmarshal(value, &success); err != nil {
			return err
		}
		if !success {
			var message string
			_ = json.Unmarshal(object["error"], &message)
			return fmt.Errorf("host sequence-list failure: %s", message)
		}
		if len(object) != 2 || object["data"] == nil {
			return fmt.Errorf("invalid sequence-list envelope")
		}
		return decodeSequenceList(string(object["data"]), out)
	}
	if err := sequenceListAliases(object, map[string]string{"activeSequenceID": "active_sequence_id"}); err != nil {
		return err
	}
	if object["active_sequence_id"] == nil || string(object["active_sequence_id"]) == "null" {
		return fmt.Errorf("missing readable active sequence metadata")
	}
	entries, exists := object["sequences"]
	if !exists || string(entries) == "null" {
		return fmt.Errorf("missing readable sequences array")
	}
	var sequences []map[string]json.RawMessage
	if err := json.Unmarshal(entries, &sequences); err != nil {
		return err
	}
	for _, item := range sequences {
		if item == nil {
			return fmt.Errorf("unreadable sequence entry")
		}
		if err := sequenceListAliases(item, map[string]string{
			"sequenceID": "sequence_id", "id": "sequence_id",
			"frameSizeHorizontal": "frame_size_horizontal", "frameSizeVertical": "frame_size_vertical",
			"videoTrackCount": "video_track_count", "audioTrackCount": "audio_track_count", "isActive": "is_active",
		}); err != nil {
			return err
		}
		for _, key := range []string{"index", "name", "sequence_id", "frame_size_horizontal", "frame_size_vertical", "timebase", "video_track_count", "audio_track_count", "is_active"} {
			if item[key] == nil || string(item[key]) == "null" {
				return fmt.Errorf("incomplete sequence metadata: %s; reload the CEP panel", key)
			}
		}
	}
	normalized, err := json.Marshal(sequences)
	if err != nil {
		return err
	}
	object["sequences"] = normalized
	normalized, err = json.Marshal(object)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(normalized, out); err != nil {
		return err
	}
	if object["count"] == nil || out.Count != len(out.Sequences) {
		return fmt.Errorf("sequence count disagrees with inventory")
	}
	seen := make(map[string]bool)
	activeCount := 0
	for index, item := range out.Sequences {
		if item == nil || item.SequenceID == "" || seen[item.SequenceID] {
			return fmt.Errorf("missing or duplicate sequence identity")
		}
		if item.VideoTrackCount < 0 || item.AudioTrackCount < 0 {
			return fmt.Errorf("negative sequence track count")
		}
		if item.Index != index || item.FrameSizeHorizontal <= 0 || item.FrameSizeVertical <= 0 {
			return fmt.Errorf("invalid sequence index or dimensions")
		}
		ticks, err := strconv.ParseUint(item.Timebase, 10, 64)
		if err != nil || ticks == 0 {
			return fmt.Errorf("invalid sequence timebase")
		}
		seen[item.SequenceID] = true
		if item.IsActive {
			activeCount++
			if item.SequenceID != out.ActiveSequenceID {
				return fmt.Errorf("active sequence identity disagrees with inventory")
			}
		}
	}
	if activeCount > 1 || (out.ActiveSequenceID != "" && (!seen[out.ActiveSequenceID] || activeCount != 1)) {
		return fmt.Errorf("unverifiable active sequence identity")
	}
	return nil
}

func sequenceListAliases(object map[string]json.RawMessage, aliases map[string]string) error {
	for old, current := range aliases {
		value, present := object[old]
		if !present {
			continue
		}
		if existing, conflict := object[current]; conflict {
			var a, b any
			if err := json.Unmarshal(existing, &a); err != nil {
				return err
			}
			if err := json.Unmarshal(value, &b); err != nil {
				return err
			}
			if !reflect.DeepEqual(a, b) {
				return fmt.Errorf("conflicting sequence field %s/%s", old, current)
			}
		}
		object[current] = value
		delete(object, old)
	}
	return nil
}
