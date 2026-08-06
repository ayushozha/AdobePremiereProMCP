#!/usr/bin/env node

/**
 * Keep literal Go EvalCommand names and the CEP ExtendScript host in sync.
 *
 * This is intentionally static: Premiere's ExtendScript runtime is not
 * available in Node, but a missing top-level function is always a bridge
 * failure and can be caught without launching Premiere Pro.
 */

var fs = require("fs");
var path = require("path");
var vm = require("vm");

var repositoryRoot = path.resolve(__dirname, "..", "..");
var orchestratorDirectory = path.join(repositoryRoot, "go-orchestrator", "internal", "orchestrator");
var hostPath = path.join(repositoryRoot, "cep-panel", "src", "host", "premiere.jsx");
var panelPath = path.join(repositoryRoot, "cep-panel", "src", "panel.js");

var commandPattern = /EvalCommand\(ctx,\s*"([A-Za-z_$][A-Za-z0-9_$]*)"/g;
var functionPattern = /(?:^|\n)function\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*\(/g;
var commands = {};

fs.readdirSync(orchestratorDirectory).filter(function (name) {
    return /\.go$/.test(name);
}).forEach(function (name) {
    var source = fs.readFileSync(path.join(orchestratorDirectory, name), "utf8");
    var match;
    while ((match = commandPattern.exec(source)) !== null) {
        if (!commands[match[1]]) commands[match[1]] = [];
        commands[match[1]].push(name);
    }
});

var hostSource = fs.readFileSync(hostPath, "utf8");
var panelSource = fs.readFileSync(panelPath, "utf8");
if (!/exportSequence:\s*function\s*\(p\)[\s\S]*?p\.sequenceId[\s\S]*?p\.outputPath[\s\S]*?p\.presetPath[\s\S]*?\)";\s*\},/.test(panelSource)) {
    throw new Error("panel exportSequence action must forward sequenceId, outputPath, and presetPath");
}
var functions = {};
var functionMatch;
while ((functionMatch = functionPattern.exec(hostSource)) !== null) {
    functions[functionMatch[1]] = true;
}

var requiredTypedCommands = [
    "mcpDispatch",
    "mcpGetTimelineState",
    "mcpPlaceClip",
    "mcpRemoveClip",
    "mcpAddTransition",
    "mcpAddText",
    "mcpApplyEffect",
    "mcpSetAudioLevel",
    "mcpExecuteEDL"
];

var missing = Object.keys(commands).filter(function (name) {
    return !functions[name];
}).sort();
var missingTyped = requiredTypedCommands.filter(function (name) {
    return !functions[name];
});

console.log("Go EvalCommand names: " + Object.keys(commands).length);
console.log("Premiere JSX functions: " + Object.keys(functions).length);

if (missing.length) {
    console.error("\nMissing Go command functions:");
    missing.forEach(function (name) {
        console.error("  - " + name + " (" + commands[name].join(", ") + ")");
    });
}

if (missingTyped.length) {
    console.error("\nMissing typed bridge functions:");
    missingTyped.forEach(function (name) { console.error("  - " + name); });
}

if (missing.length || missingTyped.length) process.exit(1);
console.log("Symbol coverage complete: no missing host functions.");

// Exercise the generic object-to-positional adapter with representative
// contracts. Symbol presence alone cannot detect these cross-layer failures.
var context = vm.createContext({ console: console, JSON: JSON, Date: Date, Math: Math });
vm.runInContext(hostSource, context, { filename: hostPath });
vm.runInContext([
    "function __mcpPositional(trackIndex, clipIndex, durationSeconds) {",
    "  return [trackIndex, clipIndex, durationSeconds];",
    "}",
    "function __mcpArgs(argsJson) { return JSON.parse(argsJson); }",
    "function __mcpEDL(edlJSON) { return edlJSON; }",
    "function __mcpClips(clipsJson) { return clipsJson; }"
].join("\n"), context);

var positional = context.mcpDispatch(
    "__mcpPositional",
    JSON.stringify({ track_index: 2, clipIndex: 4, duration: 1.5 })
);
var completeObject = context.mcpDispatch(
    "__mcpArgs",
    JSON.stringify({ trackIndex: 3, enabled: false })
);
var wrappedJson = context.mcpDispatch(
    "__mcpEDL",
    JSON.stringify({ edlJSON: "[{\"index\":1}]" })
);
var baseJson = context.mcpDispatch(
    "__mcpClips",
    JSON.stringify({ clips: [{ trackIndex: 0, clipIndex: 1 }] })
);

if (JSON.stringify(positional) !== JSON.stringify([2, 4, 1.5])) {
    throw new Error("dispatcher positional mapping regression: " + JSON.stringify(positional));
}
if (JSON.stringify(completeObject) !== JSON.stringify({ trackIndex: 3, enabled: false })) {
    throw new Error("dispatcher full-object mapping regression: " + JSON.stringify(completeObject));
}
if (wrappedJson !== "[{\"index\":1}]") {
    throw new Error("dispatcher wrapped JSON regression: " + JSON.stringify(wrappedJson));
}
if (baseJson !== JSON.stringify([{ trackIndex: 0, clipIndex: 1 }])) {
    throw new Error("dispatcher base JSON regression: " + JSON.stringify(baseJson));
}

// Cover the real cross-layer shapes that previously defaulted destructive
// operations to track/clip zero or dropped wrapped export parameters.
vm.runInContext([
    "function copyEffects(srcTrackType,srcTrackIndex,srcClipIndex){return [srcTrackType,srcTrackIndex,srcClipIndex];}",
    "function exportDirect(sequenceIndex,outputPath,presetPath,workAreaType){return [sequenceIndex,outputPath,presetPath,workAreaType];}",
    "function createMontage(clipRefsJson){return clipRefsJson;}",
    "function batchSetMetadata(itemIndicesStr,field,value){return [itemIndicesStr,field,value];}",
    "function exportAAF(sequenceIndex,outputPath,optionsJson){return [sequenceIndex,outputPath,optionsJson];}",
    "function showInputDialog(title,promptText,defaultValue){return [title,promptText,defaultValue];}"
].join("\n"), context);
var copiedEffects = context.mcpDispatch("copyEffects", JSON.stringify({trackType:"audio",trackIndex:3,clipIndex:7}));
var nestedExport = context.mcpDispatch("exportDirect", JSON.stringify({params:{sequence_index:2,output_path:"/tmp/out.mov",preset_path:"/tmp/preset.epr",work_area_type:1}}));
var acronymJson = context.mcpDispatch("createMontage", JSON.stringify({clipRefsJSON:[{trackIndex:1,clipIndex:2}]}));
var joinedIndices = context.mcpDispatch("batchSetMetadata", JSON.stringify({itemIndices:[1,4,8],field:"Scene",value:"12"}));
var aafOptions = context.mcpDispatch("exportAAF", JSON.stringify({params:{sequence_index:3,output_path:"/tmp/out.aaf",mixdown:true,sample_rate:48000,bits_per_sample:24}}));
var promptAlias = context.mcpDispatch("showInputDialog", JSON.stringify({title:"Name",prompt:"Enter a name",defaultValue:"Cut"}));
if (JSON.stringify(copiedEffects) !== JSON.stringify(["audio",3,7])) {
    throw new Error("dispatcher source-alias regression: " + JSON.stringify(copiedEffects));
}
if (JSON.stringify(nestedExport) !== JSON.stringify([2,"/tmp/out.mov","/tmp/preset.epr",1])) {
    throw new Error("dispatcher nested params regression: " + JSON.stringify(nestedExport));
}
if (acronymJson !== JSON.stringify([{trackIndex:1,clipIndex:2}])) {
    throw new Error("dispatcher acronym JSON regression: " + JSON.stringify(acronymJson));
}
if (JSON.stringify(joinedIndices) !== JSON.stringify(["1,4,8","Scene","12"])) {
    throw new Error("dispatcher array/string adapter regression: " + JSON.stringify(joinedIndices));
}
var parsedAAFOptions = JSON.parse(aafOptions[2]);
if (aafOptions[0] !== 3 || aafOptions[1] !== "/tmp/out.aaf" || parsedAAFOptions.sample_rate !== 48000 || parsedAAFOptions.bits_per_sample !== 24) {
    throw new Error("dispatcher export options regression: " + JSON.stringify(aafOptions));
}
if (JSON.stringify(promptAlias) !== JSON.stringify(["Name","Enter a name","Cut"])) {
    throw new Error("dispatcher semantic alias regression: " + JSON.stringify(promptAlias));
}

// Typed placement must mark the source before an overwrite edit, support an
// independently omitted range boundary, and restore the project-item marks.
vm.runInContext([
    "function Time(){this.seconds=0;this.ticks='0';}",
    "var __placeMarks={inPoint:0,outPoint:10};var __placedRanges=[];var __overwriteCount=0;var __insertCount=0;var __placedClip=null;",
    "var __placeItem={nodeId:'place-item',name:'Source',getMediaPath:function(){return '/tmp/source.mov';},",
    " getInPoint:function(){return {seconds:__placeMarks.inPoint};},getOutPoint:function(){return {seconds:__placeMarks.outPoint};},",
    " setInPoint:function(value){__placeMarks.inPoint=parseFloat(value);return true;},setOutPoint:function(value){__placeMarks.outPoint=parseFloat(value);return true;}};",
    "var __placeTrack={clips:[] ,insertClip:function(){__insertCount++;},overwriteClip:function(item,time){__overwriteCount++;",
    " __placedRanges.push({inPoint:__placeMarks.inPoint,outPoint:__placeMarks.outPoint});",
    " __placedClip={nodeId:'placed-'+__overwriteCount,projectItem:item,start:{seconds:time.seconds},",
    "  inPoint:{seconds:__placeMarks.inPoint},outPoint:{seconds:__placeMarks.outPoint},getSpeed:function(){return 1;},remove:function(){}};}};",
    "__placeTrack.clips.numItems=0;var __placeTracks=[__placeTrack];__placeTracks.numTracks=1;var __emptyTracks=[];__emptyTracks.numTracks=0;",
    "var __placeSequence={sequenceID:'place-seq',videoTracks:__placeTracks,audioTracks:__emptyTracks,getSettings:function(){return {videoFrameRate:{ticks:'10584000000'}};}};",
    "app={project:{activeSequence:__placeSequence,sequences:[__placeSequence]}};app.project.sequences.numSequences=1;",
    "var __originalFindClipAtPosition=_mcpFindClipAtPosition;_mcpFindProjectItem=function(){return __placeItem;};_mcpFindClipAtPosition=function(){return {clip:__placedClip,clipIndex:0};};",
    "var __tcUnset={hours:0,minutes:0,seconds:0,frames:0,frameRate:0};var __tc0={hours:0,minutes:0,seconds:0,frames:0,frameRate:24};var __tc2={hours:0,minutes:0,seconds:2,frames:0,frameRate:24};var __tc4={hours:0,minutes:0,seconds:4,frames:0,frameRate:24};",
    "var __placeInOnly=JSON.parse(mcpPlaceClip({sourcePath:'/tmp/source.mov',track:{type:'video',trackIndex:0},position:__tc0,sourceRange:{inPoint:__tc2,outPoint:__tcUnset},speed:1}));",
    "var __placeOutOnly=JSON.parse(mcpPlaceClip({sourcePath:'/tmp/source.mov',track:{type:'video',trackIndex:0},position:__tc0,sourceRange:{inPoint:__tcUnset,outPoint:__tc4},speed:1}));",
    "__placeMarks.inPoint=3;var __placeExplicitZero=JSON.parse(mcpPlaceClip({sourcePath:'/tmp/source.mov',track:{type:'video',trackIndex:0},position:__tc0,sourceRange:{inPoint:__tc0,outPoint:__tcUnset},speed:1}));",
    "_mcpFindClipAtPosition=__originalFindClipAtPosition;"
].join("\n"), context);
if (!context.__placeInOnly.success || !context.__placeOutOnly.success || !context.__placeExplicitZero.success || context.__overwriteCount !== 3 || context.__insertCount !== 0 ||
    context.__placedRanges[0].inPoint !== 2 || context.__placedRanges[0].outPoint !== 10 ||
    context.__placedRanges[1].inPoint !== 0 || context.__placedRanges[1].outPoint !== 4 ||
    context.__placedRanges[2].inPoint !== 0 || context.__placedRanges[2].outPoint !== 10 ||
    context.__placeMarks.inPoint !== 3 || context.__placeMarks.outPoint !== 10) {
    throw new Error("typed overwrite/source-range regression: " + JSON.stringify({first:context.__placeInOnly,second:context.__placeOutOnly,explicitZero:context.__placeExplicitZero,ranges:context.__placedRanges,marks:context.__placeMarks}));
}
var minusSixAmplitude = context._dbToAmplitude(-6);
if (Math.abs(minusSixAmplitude - 0.501187) > 0.00001 ||
    Math.abs(context._amplitudeToDb(minusSixAmplitude) - (-6)) > 0.001) {
    throw new Error("audio dB/amplitude conversion regression");
}
vm.runInContext([
    "var __clip = {nodeId:'track-item-1', name:'Interview'};",
    "var __clips = [__clip]; __clips.numItems = 1;",
    "var __videoTracks = [{clips:__clips}]; __videoTracks.numTracks = 1;",
    "var __audioTracks = []; __audioTracks.numTracks = 0;",
    "var __sequence = {videoTracks:__videoTracks, audioTracks:__audioTracks};",
    "var __clipId = _mcpCanonicalClipId('video', 0, 0, __clip);",
    "var __clipResolved = _mcpResolveClipId(__sequence, __clipId);"
].join("\n"), context);
if (!/:ntrack-item-1$/.test(context.__clipId) || context.__clipResolved.clip !== context.__clip) {
    throw new Error("typed clip identity regression");
}
vm.runInContext([
    "var __removedClip={nodeId:'removed-node'};var __remainingClip={nodeId:'remaining-node'};",
    "var __remainingClips=[__remainingClip];__remainingClips.numItems=1;",
    "var __removeVerified=_mcpVerifyClipRemoved({clips:__remainingClips},_mcpClipIdentity(__removedClip),2);",
    "var __staleClips=[__removedClip];__staleClips.numItems=1;",
    "var __removeRejected=_mcpVerifyClipRemoved({clips:__staleClips},_mcpClipIdentity(__removedClip),2);"
].join("\n"), context);
if (context.__removeVerified.error || !context.__removeRejected.error) {
    throw new Error("clip removal readback regression");
}
console.log("Dispatcher contract smoke tests passed.");

// Standard-profile clip and audio mutations must reject malformed inputs and
// verify the values Premiere actually accepted before reporting success.
vm.runInContext([
    "function Time(){this.seconds=0;this.ticks='0';}",
    "function __mutationClip(name,start,end,inPoint,outPoint,mediaDuration){return {name:name,nodeId:name,",
    " start:{seconds:start},end:{seconds:end},inPoint:{seconds:inPoint},outPoint:{seconds:outPoint},",
    " projectItem:{getMediaDuration:function(){return String(mediaDuration*254016000000);}},getSpeed:function(){return 1;}};}",
    "function __resetMutationTimeline(){",
    " __mutationA=__mutationClip('A',2,6,1,5,20);__mutationB=__mutationClip('B',6,9,0,3,20);",
    " __mutationClips=[__mutationA,__mutationB];__mutationClips.numItems=2;",
    " __mutationTrack={clips:__mutationClips,isLocked:function(){return false;}};",
    " __mutationVideoTracks=[__mutationTrack];__mutationVideoTracks.numTracks=1;",
    " __mutationAudioTracks=[];__mutationAudioTracks.numTracks=0;",
    " __mutationSequence={sequenceID:'mutation-sequence',videoTracks:__mutationVideoTracks,audioTracks:__mutationAudioTracks,",
    "  getSettings:function(){return {videoFrameRate:{ticks:'10584000000'}};}};",
    " app={project:{activeSequence:__mutationSequence}};}",
    "__resetMutationTimeline();var __badMove=JSON.parse(moveClip('video',0,0,'NaN'));",
    "var __moved=JSON.parse(moveClip('video',0,0,8));",
    "__resetMutationTimeline();var __fixedStart=__mutationA.start;Object.defineProperty(__mutationA,'start',{get:function(){return __fixedStart;},set:function(){},configurable:true});",
    "var __moveReadbackRejected=JSON.parse(moveClip('video',0,0,8));",
    "__resetMutationTimeline();var __trimmedStart=JSON.parse(trimClipStart('video',0,0,3));",
    "__resetMutationTimeline();var __trimmedEnd=JSON.parse(trimClipEnd('video',0,0,8));",
    "__resetMutationTimeline();var __rippleTail=JSON.parse(rippleTrim('video',0,0,true,2));",
    "var __rippleTailState={aStart:__mutationA.start.seconds,aEnd:__mutationA.end.seconds,aIn:__mutationA.inPoint.seconds,aOut:__mutationA.outPoint.seconds,",
    " bStart:__mutationB.start.seconds,bEnd:__mutationB.end.seconds};",
    "__resetMutationTimeline();var __rippleHead=JSON.parse(rippleTrim('video',0,0,false,1));",
    "var __rippleHeadState={aStart:__mutationA.start.seconds,aEnd:__mutationA.end.seconds,aIn:__mutationA.inPoint.seconds,aOut:__mutationA.outPoint.seconds,",
    " bStart:__mutationB.start.seconds,bEnd:__mutationB.end.seconds};",
    "__resetMutationTimeline();var __ripplePastSource=JSON.parse(rippleTrim('video',0,0,false,2));"
].join("\n"), context);
if (context.__badMove.success || !context.__moved.success || !context.__moved.data.verified || context.__moveReadbackRejected.success ||
    context.__moved.data.newStartTime !== 8 || context.__moved.data.newEndTime !== 12 ||
    !context.__trimmedStart.success || context.__trimmedStart.data.newStart !== 3 || context.__trimmedStart.data.newInPoint !== 2 ||
    !context.__trimmedEnd.success || context.__trimmedEnd.data.newEnd !== 8 || context.__trimmedEnd.data.newOutPoint !== 7 ||
    !context.__rippleTail.success || context.__rippleTailState.aStart !== 2 || context.__rippleTailState.aEnd !== 8 ||
    context.__rippleTailState.aIn !== 1 || context.__rippleTailState.aOut !== 7 || context.__rippleTailState.bStart !== 8 || context.__rippleTailState.bEnd !== 11 ||
    !context.__rippleHead.success || context.__rippleHeadState.aStart !== 2 || context.__rippleHeadState.aEnd !== 7 ||
    context.__rippleHeadState.aIn !== 0 || context.__rippleHeadState.aOut !== 5 || context.__rippleHeadState.bStart !== 7 || context.__rippleHeadState.bEnd !== 10 ||
    context.__ripplePastSource.success) {
    throw new Error("verified clip-mutation regression: " + JSON.stringify({
        badMove:context.__badMove,moved:context.__moved,moveReadback:context.__moveReadbackRejected,trimmedStart:context.__trimmedStart,trimmedEnd:context.__trimmedEnd,
        rippleTail:context.__rippleTail,rippleTailState:context.__rippleTailState,rippleHead:context.__rippleHead,
        rippleHeadState:context.__rippleHeadState,ripplePastSource:context.__ripplePastSource
    }));
}

vm.runInContext([
    "var __audioAmplitude=1;var __audioSetStatus=0;var __audioWrite=true;var __audioKeys=[];var __audioRemoveStatus=0;",
    "var __audioParam={getValue:function(){return __audioAmplitude;},setValue:function(value){if(__audioSetStatus===0&&__audioWrite)__audioAmplitude=value;return __audioSetStatus;},",
    " getKeys:function(){return __audioKeys;},getValueAtKey:function(key){return key.value;},",
    " addKey:function(key){if(__audioKeys.indexOf(key)<0)__audioKeys.push(key);return 0;},setValueAtKey:function(key,value){key.value=value;return 0;},",
    " removeKey:function(key){if(__audioRemoveStatus!==0)return __audioRemoveStatus;var index=__audioKeys.indexOf(key);if(index<0)return 1;__audioKeys.splice(index,1);return 0;}};",
    "var __audioProperties={numItems:1,getParamForDisplayName:function(){return __audioParam;}};",
    "var __audioComponent={displayName:'Volume',properties:__audioProperties};var __audioComponents=[__audioComponent];__audioComponents.numItems=1;",
    "var __audioClip={components:__audioComponents};var __audioClips=[__audioClip];__audioClips.numItems=1;",
    "var __audioTrack={clips:__audioClips};var __mutationAudioTracks=[__audioTrack];__mutationAudioTracks.numTracks=1;",
    "app={project:{activeSequence:{audioTracks:__mutationAudioTracks}}};",
    "var __audioRangeRejected=JSON.parse(setAudioLevel(0,0,-120));",
    "__audioSetStatus=1;var __audioStatusRejected=JSON.parse(setAudioLevel(0,0,-6));",
    "__audioSetStatus=0;__audioWrite=false;var __audioReadbackRejected=JSON.parse(setAudioLevel(0,0,-6));",
    "__audioWrite=true;var __audioSet=JSON.parse(setAudioLevel(0,0,-6));",
    "__audioKeys=[{seconds:1,value:0.75}];__audioRemoveStatus=1;var __audioKeyRemovalRejected=JSON.parse(normalizeAudio(0,0,-3));",
    "__audioRemoveStatus=0;",
    "__audioKeys=[{seconds:1,value:0.75},{seconds:2,value:0.5}];var __audioNormalized=JSON.parse(normalizeAudio(0,0,-3));"
].join("\n"), context);
if (context.__audioRangeRejected.success || context.__audioStatusRejected.success || context.__audioReadbackRejected.success || context.__audioKeyRemovalRejected.success ||
    !context.__audioSet.success || !context.__audioSet.data.verified || !context.__audioNormalized.success ||
    !context.__audioNormalized.data.verified || context.__audioNormalized.data.keyframesRemoved !== 2 || context.__audioKeys.length !== 0 ||
    Math.abs(context.__audioNormalized.data.actualLevelDb - (-3)) > 0.001) {
    throw new Error("verified audio-mutation regression: " + JSON.stringify({
        range:context.__audioRangeRejected,status:context.__audioStatusRejected,readback:context.__audioReadbackRejected,keyRemoval:context.__audioKeyRemovalRejected,
        set:context.__audioSet,normalized:context.__audioNormalized,keys:context.__audioKeys
    }));
}
console.log("Verified clip and audio mutation smoke tests passed.");

// Verify that EDL assembly marks the project item before overwrite, restores
// those marks afterward, and never inserts a full-length source clip first.
vm.runInContext([
    "function Time() { this.seconds = 0; this.ticks = '0'; }",
    "var __sourceMarks = {inPoint:0, outPoint:10};",
    "var __overwriteMarks = null;",
    "var __overwriteCalls = 0;",
    "var __edlItem = {",
    "  nodeId:'edl-item', name:'EDL Source',",
    "  getMediaPath:function(){return '/tmp/edl-source.mov';},",
    "  getInPoint:function(){return {seconds:__sourceMarks.inPoint};},",
    "  getOutPoint:function(){return {seconds:__sourceMarks.outPoint};},",
    "  setInPoint:function(value){__sourceMarks.inPoint=parseFloat(value); return true;},",
    "  setOutPoint:function(value){__sourceMarks.outPoint=parseFloat(value); return true;}",
    "};",
    "var __edlClips = []; __edlClips.numItems = 0;",
    "var __edlTrack = {clips:__edlClips, isLocked:function(){return false;}, overwriteClip:function(item,time){",
    "  __overwriteCalls++; __overwriteMarks={inPoint:__sourceMarks.inPoint,outPoint:__sourceMarks.outPoint};",
    "  var duration=__sourceMarks.outPoint-__sourceMarks.inPoint;",
    "  var clip={nodeId:'edl-clip',projectItem:item,start:{seconds:time.seconds},end:{seconds:time.seconds+duration},",
    "    inPoint:{seconds:__sourceMarks.inPoint},outPoint:{seconds:__sourceMarks.outPoint},getSpeed:function(){return 1;},",
    "    remove:function(){__edlClips.length=0;__edlClips.numItems=0;}};",
    "  __edlClips.length=0;__edlClips.push(clip);__edlClips.numItems=1;",
    "}};",
    "var __edlVideoTracks=[__edlTrack];__edlVideoTracks.numTracks=1;",
    "var __edlAudioTracks=[];__edlAudioTracks.numTracks=0;",
    "var __edlSequence={sequenceID:'edl-sequence',videoTracks:__edlVideoTracks,audioTracks:__edlAudioTracks,",
    "  getSettings:function(){return {videoFrameWidth:1920,videoFrameHeight:1080,videoFrameRate:{ticks:'10584000000'}};}};",
    "var app={project:{activeSequence:__edlSequence}};",
    "_mcpFindProjectItem=function(){return __edlItem;};",
    "var __assembled=JSON.parse(assembleFromEDL({clips:[{file:'/tmp/edl-source.mov',trackType:'video',",
    "  trackIndex:0,position:5,inPoint:2,outPoint:4,expectedTimelineOut:7,speed:1}]}));"
].join("\n"), context);
if (!context.__assembled.success || context.__assembled.data.placed !== 1 || context.__overwriteCalls !== 1 ||
    context.__overwriteMarks.inPoint !== 2 || context.__overwriteMarks.outPoint !== 4 ||
    context.__sourceMarks.inPoint !== 0 || context.__sourceMarks.outPoint !== 10) {
    throw new Error("EDL marked-source overwrite regression: " + JSON.stringify(context.__assembled));
}

// A partial lower-level assembly must surface as a failed typed EDL result,
// never as "completed" merely because one clip was placed.
vm.runInContext([
    "assembleFromEDL=function(){return _ok({placed:1,total:2,transitionsAdded:0,errors:['forced partial'],warnings:[]});};",
    "var __typedRangeA={inPoint:{hours:0,minutes:0,seconds:0,frames:0,frameRate:24},",
    "  outPoint:{hours:0,minutes:0,seconds:1,frames:0,frameRate:24}};",
    "var __typedRangeB={inPoint:{hours:0,minutes:0,seconds:1,frames:0,frameRate:24},",
    "  outPoint:{hours:0,minutes:0,seconds:2,frames:0,frameRate:24}};",
    "var __typedEDL={name:'Typed EDL',sequenceResolution:{width:1920,height:1080},sequenceFrameRate:24,entries:[",
    "  {sourceAssetId:'edl-item',sourceRange:__typedRangeA,timelineRange:__typedRangeA,track:{type:1,trackIndex:0},effects:[]},",
    "  {sourceAssetId:'edl-item',sourceRange:__typedRangeA,timelineRange:__typedRangeB,track:{type:1,trackIndex:0},effects:[]}",
    "]};",
    "var __typedResult=JSON.parse(mcpExecuteEDL({edl:__typedEDL,autoImport:false,autoCreateSequence:false}));"
].join("\n"), context);
if (!context.__typedResult.success || context.__typedResult.data.status !== "failed" ||
    context.__typedResult.data.clipsPlaced !== 1 || context.__typedResult.data.errors.length !== 1) {
    throw new Error("typed EDL partial-status regression: " + JSON.stringify(context.__typedResult));
}
vm.runInContext([
    "var __mismatchEDL={name:'Mismatch EDL',sequenceResolution:{width:1080,height:1920},sequenceFrameRate:30,entries:[",
    "  {sourceAssetId:'edl-item',sourceRange:__typedRangeA,timelineRange:__typedRangeA,track:{type:1,trackIndex:0},effects:[]}",
    "]};",
    "var __mismatchResult=JSON.parse(mcpExecuteEDL({edl:__mismatchEDL,autoImport:false,autoCreateSequence:false}));",
    "var __alignmentEDL={name:'Alignment EDL',sequenceResolution:{width:1920,height:1080},sequenceFrameRate:24,entries:[",
    "  {sourceAssetId:'edl-item',sourceRange:__typedRangeA,timelineRange:__typedRangeA,track:{type:1,trackIndex:0},",
    "   transition:{type:'cross dissolve',durationSeconds:0.5,alignment:'center'},effects:[]}",
    "]};",
    "var __alignmentResult=JSON.parse(mcpExecuteEDL({edl:__alignmentEDL,autoImport:false,autoCreateSequence:false}));"
].join("\n"), context);
if (!context.__mismatchResult.success || context.__mismatchResult.data.status !== "failed" ||
    context.__mismatchResult.data.clipsPlaced !== 0 || context.__mismatchResult.data.errors.length !== 2 ||
    !context.__alignmentResult.success || context.__alignmentResult.data.status !== "failed" ||
    context.__alignmentResult.data.errors.join(" ").indexOf("alignment") < 0) {
    throw new Error("typed EDL settings/alignment preflight regression");
}

// Creation must fail and remove the sequence when Premiere ignores requested
// settings instead of echoing those requested values as if they were applied.
vm.runInContext([
    "function __tracks(count){var result=[];for(var i=0;i<count;i++)result.push({clips:(function(){var c=[];c.numItems=0;return c;})()});result.numTracks=count;return result;}",
    "var __badSequence={sequenceID:'bad-sequence',name:'Bad Sequence',videoTracks:__tracks(1),audioTracks:__tracks(1),timebase:'10584000000',",
    "  getSettings:function(){return {videoFrameWidth:1280,videoFrameHeight:720,videoFrameRate:{ticks:'10584000000'}};},",
    "  setSettings:function(){return true;}};",
    "var __badSequences=[];__badSequences.numSequences=0;",
    "app={enableQE:function(){},project:{sequences:__badSequences,activeSequence:null,",
    "  createNewSequence:function(){__badSequences.push(__badSequence);__badSequences.numSequences=1;this.activeSequence=__badSequence;return 'bad-sequence';},",
    "  deleteSequence:function(sequence){var index=__badSequences.indexOf(sequence);if(index>=0)__badSequences.splice(index,1);__badSequences.numSequences=__badSequences.length;this.activeSequence=null;}}};",
    "qe={project:{getSequenceAt:function(){return {setFrameSize:function(){}};},getActiveSequence:function(){return {}}}};",
    "var __badCreate=JSON.parse(createSequence({name:'Bad Sequence',width:1920,height:1080,fps:24,videoTracks:1,audioTracks:1}));"
].join("\n"), context);
if (context.__badCreate.success || context.__badSequences.numSequences !== 0 ||
    context.__badCreate.error.indexOf("invalid sequence was deleted") < 0) {
    throw new Error("createSequence verification/rollback regression: " + JSON.stringify(context.__badCreate));
}

// Social versions must be content-preserving clones with verified dimensions.
vm.runInContext([
    "function __clipList(count){var list=[];for(var i=0;i<count;i++)list.push({nodeId:'clip-'+i});list.numItems=count;return list;}",
    "function __socialTracks(count,clips){var result=[];for(var i=0;i<count;i++)result.push({clips:__clipList(clips)});result.numTracks=count;return result;}",
    "var __socialSequences=[];__socialSequences.numSequences=0;",
    "function __makeSocialSequence(id,name,width,height){var settings={videoFrameWidth:width,videoFrameHeight:height,videoFrameRate:{ticks:'10584000000'}};",
    "  return {sequenceID:id,name:name,videoTracks:__socialTracks(1,1),audioTracks:__socialTracks(1,1),timebase:'10584000000',",
    "    getSettings:function(){return settings;},setSettings:function(next){settings=next;return true;}};}",
    "var __socialSource=__makeSocialSequence('social-source','Main Edit',1920,1080);",
    "__socialSource.clone=function(){var clone=__makeSocialSequence('social-clone','Main Edit Copy',1920,1080);__socialSequences.push(clone);",
    "  __socialSequences.numSequences=__socialSequences.length;app.project.activeSequence=clone;};",
    "__socialSequences.push(__socialSource);__socialSequences.numSequences=1;",
    "app={enableQE:function(){},project:{sequences:__socialSequences,activeSequence:__socialSource,",
    "  openSequence:function(id){for(var i=0;i<__socialSequences.numSequences;i++){if(__socialSequences[i].sequenceID===id){this.activeSequence=__socialSequences[i];return true;}}return false;},",
    "  deleteSequence:function(sequence){var index=__socialSequences.indexOf(sequence);if(index>=0)__socialSequences.splice(index,1);__socialSequences.numSequences=__socialSequences.length;}}};",
    "var __vertical=JSON.parse(createVerticalVersion(0,'Vertical Cut'));"
].join("\n"), context);
if (!context.__vertical.success || !context.__vertical.data.cloned || context.__vertical.data.clipCount !== 2 ||
    context.__vertical.data.width !== 1080 || context.__vertical.data.height !== 1920 || context.__socialSequences.numSequences !== 2 ||
    context.__socialSource.getSettings().videoFrameWidth !== 1920 || context.__socialSource.getSettings().videoFrameHeight !== 1080) {
    throw new Error("social sequence clone regression: " + JSON.stringify(context.__vertical));
}
console.log("EDL and verified-sequence workflow smoke tests passed.");

// QE calls are not success by themselves: effects and transitions must appear
// in the public DOM before the host reports them as applied.
vm.runInContext([
    "var __effectComponents=[{displayName:'Motion'}];__effectComponents.numItems=1;",
    "var __effectClip={components:__effectComponents};",
    "var __effectClips=[__effectClip];__effectClips.numItems=1;",
    "var __effectTrack={clips:__effectClips};",
    "var __effectTracks=[__effectTrack];__effectTracks.numTracks=1;",
    "app={enableQE:function(){},project:{activeSequence:{videoTracks:__effectTracks}}};",
    "var __qeEffectClip={addVideoEffect:function(){}};",
    "qe={project:{getActiveSequence:function(){return {getVideoTrackAt:function(){return {getItemAt:function(){return __qeEffectClip;}};}};},",
    "  getVideoEffectByName:function(){return {name:'Gaussian Blur'};}}};",
    "var __effectNoop=JSON.parse(applyVideoEffect(0,0,'Gaussian Blur'));",
    "__qeEffectClip.addVideoEffect=function(){__effectComponents.push({displayName:'Gaussian Blur',properties:[]});__effectComponents.numItems=__effectComponents.length;};",
    "var __effectApplied=JSON.parse(applyVideoEffect(0,0,'Gaussian Blur'));"
].join("\n"), context);
if (context.__effectNoop.success || !context.__effectApplied.success || !context.__effectApplied.data.verified ||
    context.__effectComponents.numItems !== 2) {
    throw new Error("effect readback verification regression");
}

vm.runInContext([
    "__effectComponents.length=1;__effectComponents.numItems=1;",
    "__effectClip.nodeId='effect-clip';",
    "var __effectAudioTracks=[];__effectAudioTracks.numTracks=0;",
    "app.project.activeSequence.sequenceID='effect-sequence';app.project.activeSequence.audioTracks=__effectAudioTracks;",
    "__qeEffectClip.addVideoEffect=function(){var added={displayName:'Gaussian Blur',properties:[],remove:function(){",
    "  __effectComponents.splice(__effectComponents.indexOf(added),1);__effectComponents.numItems=__effectComponents.length;}};",
    "  added.properties.numItems=0;__effectComponents.push(added);__effectComponents.numItems=__effectComponents.length;};",
    "var __rollbackClipId=_mcpCanonicalClipId('video',0,0,__effectClip);",
    "var __effectRollback=JSON.parse(mcpApplyEffect(JSON.stringify({clipId:__rollbackClipId,",
    "  effect:{name:'Gaussian Blur',parameters:{Blurriness:10}}})));"
].join("\n"), context);
if (context.__effectRollback.success || context.__effectComponents.numItems !== 1 ||
    context.__effectRollback.error.indexOf("rolled back") < 0) {
    throw new Error("effect parameter-failure rollback regression: " + JSON.stringify(context.__effectRollback));
}

vm.runInContext([
    "var __paramStored=5;var __paramStatus=1;var __verifiedParam={setValue:function(value){__paramStored=value;return __paramStatus;},getValue:function(){return __paramStored;}};",
    "var __paramRejected=_setComponentParamAndVerify(__verifiedParam,10);",
    "__paramStatus=0;__verifiedParam.getValue=function(){return 9;};var __paramMismatch=_setComponentParamAndVerify(__verifiedParam,10);",
    "__verifiedParam.getValue=function(){return __paramStored;};var __paramAccepted=_setComponentParamAndVerify(__verifiedParam,10);",
    "var __mogrtProperties=[__verifiedParam];__mogrtProperties.numItems=1;__verifiedParam.displayName='Headline';",
    "var __mogrtClips=[{getMGTComponent:function(){return {properties:__mogrtProperties};}}];__mogrtClips.numItems=1;",
    "var __mogrtTracks=[{clips:__mogrtClips}];__mogrtTracks.numTracks=1;app={project:{activeSequence:{videoTracks:__mogrtTracks}}};",
    "var __mogrtWrite=JSON.parse(setMOGRTText(0,0,0,'Verified title'));"
].join("\n"), context);
if (!context.__paramRejected.error || !context.__paramMismatch.error || context.__paramAccepted.error ||
    !context.__mogrtWrite.success || !context.__mogrtWrite.data.verified || context.__paramStored !== "Verified title") {
    throw new Error("component parameter status/readback regression");
}

vm.runInContext([
    "var __transitions=[];__transitions.numItems=0;",
    "var __transitionClip={end:{seconds:2}};var __transitionClips=[__transitionClip];__transitionClips.numItems=1;",
    "var __transitionTrack={clips:__transitionClips,transitions:__transitions};",
    "var __transitionTracks=[__transitionTrack];__transitionTracks.numTracks=1;",
    "app={enableQE:function(){},project:{activeSequence:{videoTracks:__transitionTracks}}};",
    "var __qeTransitionClip={addTransition:function(){}};",
    "qe={project:{getActiveSequence:function(){return {getVideoTrackAt:function(){return {getItemAt:function(){return __qeTransitionClip;}};}};},",
    "  getVideoTransitionByName:function(){return {name:'Cross Dissolve'};}}};",
    "var __transitionNoop=JSON.parse(addVideoTransition(0,0,'Cross Dissolve',1,true));",
    "__qeTransitionClip.addTransition=function(){__transitions.push({nodeId:'transition-1',displayName:'Cross Dissolve',start:{seconds:1.6},end:{seconds:2.4}});__transitions.numItems=__transitions.length;};",
    "var __transitionApplied=JSON.parse(addVideoTransition(0,0,'Cross Dissolve',1,true));",
    "__transitions[0].remove=function(){__transitions.splice(0,1);__transitions.numItems=__transitions.length;return 0;};",
    "var __transitionRemoved=JSON.parse(removeTransition('video',0,0));",
    "var __staleTransition={nodeId:'stale-transition',displayName:'Dip to Black',start:{seconds:3},end:{seconds:4},remove:function(){return 0;}};",
    "__transitions.push(__staleTransition);__transitions.numItems=__transitions.length;",
    "var __transitionRemovalRejected=JSON.parse(removeTransition('video',0,0));"
].join("\n"), context);
if (context.__transitionNoop.success || !context.__transitionApplied.success ||
    !context.__transitionApplied.data.verified || Math.abs(context.__transitionApplied.data.duration - 0.8) > 0.0001 ||
    !context.__transitionRemoved.success || !context.__transitionRemoved.data.verified || context.__transitionRemovalRejected.success ||
    context.__transitions.numItems !== 1) {
    throw new Error("transition readback verification regression");
}

vm.runInContext([
    "var __qeEnableCount=0;var __mixerComponents=[];__mixerComponents.numItems=0;",
    "var __mixerAudioTracks=[{name:'Dialogue',clips:{numItems:2},components:__mixerComponents,isMuted:function(){return false;},isLocked:function(){return true;}}];__mixerAudioTracks.numTracks=1;",
    "app={enableQE:function(){__qeEnableCount++;},project:{activeSequence:{name:'Mix',audioTracks:__mixerAudioTracks}}};",
    "qe={project:{getActiveSequence:function(){return {getAudioTrackAt:function(){return {isSolo:function(){return true;}};}};},",
    " getVideoEffectList:function(){return 'Gaussian Blur|Lumetri Color|';},getAudioEffectList:function(){return 'Dynamics|';},",
    " getVideoTransitionList:function(){return 'Cross Dissolve|Dip to Black|';},getAudioTransitionList:function(){return 'Constant Power|';}}};",
    "var __installedEffects=JSON.parse(getInstalledEffects());var __installedTransitions=JSON.parse(getInstalledTransitions());var __mixerState=JSON.parse(getAudioMixerState());"
].join("\n"), context);
if (!context.__installedEffects.success || context.__installedEffects.data.totalCount !== 3 ||
    context.__installedEffects.data.effects[1].name !== "Lumetri Color" || !context.__installedTransitions.success ||
    context.__installedTransitions.data.totalCount !== 3 || !context.__mixerState.success ||
    context.__mixerState.data.tracks[0].soloed !== true || context.__mixerState.data.tracks[0].volume !== null ||
    context.__qeEnableCount !== 3) {
    throw new Error("QE discovery/mixer initialization regression");
}
console.log("Effect and transition readback smoke tests passed.");

// Timed-caption import must use Sequence.createCaptionTrack and verify the
// created segment count. Legacy helpers that used to truncate caption text
// now fail explicitly.
vm.runInContext([
    "function File(filePath){this.fsName=filePath;this.exists=true;this.open=function(){};this.close=function(){};",
    "  this.read=function(){return '1\\n00:00:00,000 --> 00:00:01,000\\nHello\\n\\n2\\n00:00:01,000 --> 00:00:02,000\\nWorld\\n';};}",
    "var __captionChildren=[];__captionChildren.numItems=0;",
    "var __captionTracks=[];__captionTracks.numTracks=0;",
    "var __captionSequence={captionTracks:__captionTracks,createCaptionTrack:function(){",
    "  var clips=[{name:'Hello',start:{seconds:0},end:{seconds:1}},{name:'World',start:{seconds:1},end:{seconds:2}}];clips.numItems=2;",
    "  __captionTracks.push({clips:clips});__captionTracks.numTracks=__captionTracks.length;return true;}};",
    "app={project:{activeSequence:__captionSequence,rootItem:{children:__captionChildren},importFiles:function(paths){",
    "  __captionChildren.push({name:'captions.srt',getMediaPath:function(){return paths[0];}});",
    "  __captionChildren.numItems=__captionChildren.length;return true;}}};",
    "var __captionImported=JSON.parse(addSubtitlesFromSRT('/tmp/captions.srt',0));",
    "var __captionSplit=JSON.parse(splitLongCaptions(0,42));",
    "var __captionAlign=JSON.parse(alignCaptionToSpeech(0));",
    "var __captionAuto=JSON.parse(autoGenerateSubtitles('en','default'));",
    "var __captionTranslate=JSON.parse(translateSubtitles(0,'es'));",
    "var __captionFormat=JSON.parse(formatSubtitles(0,42,2));",
    "var __captionBurn=JSON.parse(burnInSubtitles(0));",
    "var __fakeGraphics=[",
    "  JSON.parse(addScrollingTitle('Title',0,0,5,100)),",
    "  JSON.parse(addTypewriterText('Title',0,0,5,50)),",
    "  JSON.parse(addTextWithBackground('Title',0,0,5,'#000000',10)),",
    "  JSON.parse(addRectangle(0,0,5,0,0,100,100,'#ffffff',0)),",
    "  JSON.parse(addWatermark('/tmp/logo.png','top-left',50,25)),",
    "  JSON.parse(createSplitScreen('2-up','[]')),",
    "  JSON.parse(createCollage('[]',2,2,0))",
    "];"
].join("\n"), context);
if (!context.__captionImported.success || !context.__captionImported.data.verified ||
    context.__captionImported.data.captionsCreated !== 2 || context.__captionSplit.success || context.__captionAlign.success ||
    context.__captionAuto.success || context.__captionTranslate.success || context.__captionFormat.success || context.__captionBurn.success ||
    context.__fakeGraphics.some(function (result) { return result.success; })) {
    throw new Error("caption import/unsupported-operation regression: " + JSON.stringify(context.__captionImported));
}
console.log("Caption import and safety smoke tests passed.");
