package orchestrator

// ---------------------------------------------------------------------------
// Advanced Audio Processing & Mixing Operations
// ---------------------------------------------------------------------------
//
// All 30 advanced audio tools route through EvalAudioCommand (defined in
// orchestrator.go) which forwards ExtendScript function calls to the Premiere
// bridge. No additional Engine methods are required because each MCP tool uses
// makeAudioHandler with a command name that maps directly to an ExtendScript
// function in premiere.jsx.
//
// Command mapping (MCP tool -> ExtendScript function):
//
// Audio Mixing:
//   1.  getAudioMixerState       -> getAudioMixerState()
//   2.  setTrackPanning           -> setTrackPanning(trackIndex, panValue)
//   3.  setClipPanning            -> setClipPanning(trackIndex, clipIndex, panValue)
//   4.  getClipPanning            -> getClipPanning(trackIndex, clipIndex)
//
// Audio Keyframes (extended):
//   5.  addVolumeKeyframe         -> addVolumeKeyframe(trackIndex, clipIndex, time, levelDb)
//   6.  addPanningKeyframe        -> addPanningKeyframe(trackIndex, clipIndex, time, panValue)
//   7.  getVolumeKeyframes        -> getVolumeKeyframes(trackIndex, clipIndex)
//   8.  getPanningKeyframes       -> getPanningKeyframes(trackIndex, clipIndex)
//   9.  removeAllAudioKeyframes   -> removeAllAudioKeyframes(trackIndex, clipIndex)
//
// Audio Routing:
//   10. setTrackOutput            -> setTrackOutput(trackIndex, outputChannels)
//   11. getTrackOutput            -> getTrackOutput(trackIndex)
//   12. createSubmix              -> createSubmix(name, channelType)
//
// Audio Effects (extended):
//   13. applyEQ                   -> applyEQ(trackIndex, clipIndex, bands)
//   14. applyCompressor           -> applyCompressor(trackIndex, clipIndex, threshold, ratio, attack, release)
//   15. applyLimiter              -> applyLimiter(trackIndex, clipIndex, ceiling)
//   16. applyDeEsser              -> applyDeEsser(trackIndex, clipIndex, frequency, reduction)
//   17. getAudioEffectPresets     -> getAudioEffectPresets()
//   18. applyAudioPreset          -> applyAudioPreset(trackIndex, clipIndex, presetName)
//
// Audio Analysis (extended):
//   19. getAudioWaveformData      -> getAudioWaveformData(trackIndex, clipIndex, samples)
//   20. getLoudnessInfo           -> getLoudnessInfo(trackIndex, clipIndex)
//   21. getSequenceLoudness       -> getSequenceLoudness()
//   22. findAudioPeaks            -> findAudioPeaks(trackIndex, clipIndex, thresholdDb)
//   23. detectClipping            -> detectClipping(trackIndex, clipIndex)
//
// Voiceover:
//   24. prepareVoiceoverTrack     -> prepareVoiceoverTrack(trackIndex)
//   25. setVoiceoverDucking       -> setVoiceoverDucking(voTrackIndex, musicTrackIndex, duckAmount, sensitivity)
//
// Audio Sync:
//   26. syncAudioToVideo          -> syncAudioToVideo(audioTrackIndex, audioClipIndex, videoTrackIndex, videoClipIndex)
//   27. detectAudioDrift          -> detectAudioDrift(trackIndex, clipIndex)
//
// Channel Operations:
//   28. convertStereoToMono       -> convertStereoToMono(projectItemIndex)
//   29. swapAudioChannels         -> swapAudioChannels(projectItemIndex)
//   30. extractAudioFromVideo     -> extractAudioFromVideo(projectItemIndex)
