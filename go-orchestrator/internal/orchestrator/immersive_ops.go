package orchestrator

// ---------------------------------------------------------------------------
// Immersive Video, VR/360, HDR & Advanced Format Operations
// ---------------------------------------------------------------------------
//
// All 30 immersive/advanced-format tools route through EvalImmersiveCommand
// (defined in orchestrator.go) which forwards ExtendScript function calls to
// the Premiere bridge. No additional Engine methods are required because each
// MCP tool uses makeImmersiveHandler with a command name that maps directly
// to an ExtendScript function in premiere.jsx.
//
// Command mapping (MCP tool -> ExtendScript function):
//
// VR / 360 Video:
//   1.  setVRProjection          -> setVRProjection(sequenceIndex, projection)
//   2.  getVRProjection          -> getVRProjection(sequenceIndex)
//   3.  setVRFieldOfView         -> setVRFieldOfView(trackIndex, clipIndex, fov)
//   4.  rotateVRView             -> rotateVRView(trackIndex, clipIndex, pan, tilt, roll)
//   5.  createVRSequence         -> createVRSequence(name, width, height, fps, projection)
//
// HDR:
//   6.  setHDRSettings           -> setHDRSettings(sequenceIndex, colorSpace, maxLuminance)
//   7.  getHDRSettings           -> getHDRSettings(sequenceIndex)
//   8.  isHDRSequence            -> isHDRSequence(sequenceIndex)
//
// Stereoscopic 3D:
//   9.  setStereoscopicMode      -> setStereoscopicMode(sequenceIndex, mode)
//   10. getStereoscopicMode      -> getStereoscopicMode(sequenceIndex)
//
// Frame Rate:
//   11. setSequenceFrameRate     -> setSequenceFrameRate(sequenceIndex, fps)
//   12. interpretFootageFrameRate -> interpretFootageFrameRate(projectItemIndex, fps)
//   13. getAvailableFrameRates   -> getAvailableFrameRates()
//
// Aspect Ratio:
//   14. setPixelAspectRatio      -> setPixelAspectRatio(sequenceIndex, num, den)
//   15. getPixelAspectRatio      -> getPixelAspectRatio(sequenceIndex)
//   16. addLetterboxing          -> addLetterboxing(trackIndex, clipIndex, targetAspect)
//   17. addPillarboxing          -> addPillarboxing(trackIndex, clipIndex, targetAspect)
//
// Timecode:
//   18. setTimecodeOffset        -> setTimecodeOffset(sequenceIndex, offset)
//   19. getTimecodeOffset        -> getTimecodeOffset(sequenceIndex)
//   20. setDropFrame             -> setDropFrame(sequenceIndex, enabled)
//   21. convertTimecode          -> convertTimecode(timecode, fromFps, toFps)
//
// Render Settings:
//   22. getMaxRenderQuality      -> getMaxRenderQuality()
//   23. setMaxRenderQuality      -> setMaxRenderQuality(enabled)
//   24. setMaxBitDepth           -> setMaxBitDepth(enabled)
//   25. getGPURenderStatus       -> getGPURenderStatus()
//
// Closed Captions (extended):
//   26. getCaptionFormats        -> getCaptionFormats()
//   27. setCaptionPosition       -> setCaptionPosition(trackIndex, captionIndex, x, y)
//   28. setCaptionBackground     -> setCaptionBackground(trackIndex, captionIndex, color, opacity)
//   29. alignCaptionToSpeech     -> alignCaptionToSpeech(trackIndex)
//   30. splitLongCaptions        -> splitLongCaptions(trackIndex, maxChars)
