import assert from "node:assert/strict";
import test from "node:test";

import { loadConfig, type BridgeConfig } from "../config.js";
import { CepBridge, CepCommandError } from "./cep-bridge.js";
import type { EvalCommandResult } from "./interface.js";

const config: BridgeConfig = {
  grpcPort: 50054,
  grpcHost: "127.0.0.1",
  premierePath: "/Applications/Adobe Premiere Pro.app",
  bridgeMode: "cep",
  logLevel: "error",
  cepWsPort: 9801,
  cepToken: "test-only-cep-token-0123456789abcdef",
};

interface RecordedCall {
  functionName: string;
  args: Record<string, unknown>;
}

function bridgeWithHostResult(
  result: unknown,
  bridgeConfig: BridgeConfig = config,
): { bridge: CepBridge; calls: RecordedCall[] } {
  const bridge = new CepBridge(bridgeConfig);
  const calls: RecordedCall[] = [];
  bridge.evalCommand = async (
    functionName: string,
    argsJson: string,
  ): Promise<EvalCommandResult> => {
    calls.push({
      functionName,
      args: JSON.parse(argsJson) as Record<string, unknown>,
    });
    return {
      resultJson: JSON.stringify(result),
      isError: false,
      errorMessage: "",
    };
  };
  return { bridge, calls };
}

test("normalizes host project state into the bridge contract", async () => {
  const { bridge } = bridgeWithHostResult({
    name: "Demo Project",
    path: "/tmp/demo.prproj",
    binCount: 4,
    sequences: [
      {
        sequenceID: "sequence-1",
        name: "Master",
        frameSizeHorizontal: 1920,
        frameSizeVertical: 1080,
        fps: 23.976,
        outPoint: 12.5,
        videoTrackCount: 3,
        audioTrackCount: 2,
      },
    ],
  });

  const state = await bridge.getProjectState();

  assert.equal(state.projectName, "Demo Project");
  assert.equal(state.projectPath, "/tmp/demo.prproj");
  assert.equal(state.binCount, 4);
  assert.equal(state.isSaved, false);
  assert.deepEqual(state.sequences[0], {
    id: "sequence-1",
    name: "Master",
    resolution: { width: 1920, height: 1080 },
    frameRate: 23.976,
    durationSeconds: 12.5,
    videoTrackCount: 3,
    audioTrackCount: 2,
  });
});

test("adapts core sequence and clip calls to explicit host commands", async () => {
  const { bridge, calls } = bridgeWithHostResult({
    sequenceID: "created-1",
    name: "Vertical",
    clipId: "clip-1",
  });

  const created = await bridge.createSequence({
    name: "Vertical",
    resolution: { width: 1080, height: 1920 },
    frameRate: 30,
    videoTracks: 2,
    audioTracks: 2,
  });
  await bridge.placeClip({
    sourcePath: "/tmp/source.mov",
    track: { type: "video", trackIndex: 1 },
    position: { hours: 0, minutes: 0, seconds: 3, frames: 0, frameRate: 30 },
    sourceRange: {
      inPoint: { hours: 0, minutes: 0, seconds: 1, frames: 0, frameRate: 30 },
      outPoint: { hours: 0, minutes: 0, seconds: 5, frames: 0, frameRate: 30 },
    },
    speed: 1.25,
  });

  assert.deepEqual(created, { sequenceId: "created-1", name: "Vertical" });
  assert.deepEqual(calls[0], {
    functionName: "createSequence",
    args: {
      name: "Vertical",
      width: 1080,
      height: 1920,
      fps: 30,
      videoTracks: 2,
      audioTracks: 2,
    },
  });
  assert.equal(calls[1]?.functionName, "mcpPlaceClip");
  assert.deepEqual(
    (calls[1]?.args["sourceRange"] as Record<string, unknown>)["inPoint"],
    { hours: 0, minutes: 0, seconds: 1, frames: 0, frameRate: 30 },
  );
  assert.equal(calls[1]?.args["speed"], 1.25);
});

test("turns host command failures into typed bridge errors", async () => {
  const bridge = new CepBridge(config);
  bridge.evalCommand = async (): Promise<EvalCommandResult> => ({
    resultJson: "",
    isError: true,
    errorMessage: "effect is unavailable",
  });

  await assert.rejects(
    bridge.applyEffect({
      clipId: "clip-1",
      sequenceId: "sequence-1",
      effect: { name: "Missing Effect", parameters: {} },
    }),
    (error: unknown) =>
      error instanceof CepCommandError &&
      error.message.includes("effect is unavailable"),
  );
});

test("uses a loopback-only gRPC bind by default", () => {
  const previous = process.env["BRIDGE_GRPC_HOST"];
  delete process.env["BRIDGE_GRPC_HOST"];
  try {
    assert.equal(loadConfig().grpcHost, "127.0.0.1");
  } finally {
    if (previous === undefined) delete process.env["BRIDGE_GRPC_HOST"];
    else process.env["BRIDGE_GRPC_HOST"] = previous;
  }
});

test("keeps the CEP token out of the WebSocket URL", () => {
  const bridge = new CepBridge(config) as unknown as {
    wsEndpoint: string;
    wsHeaders: Readonly<Record<string, string>>;
  };

  assert.equal(bridge.wsEndpoint, "ws://127.0.0.1:9801");
  assert.equal(bridge.wsEndpoint.includes("test-only-cep-token-0123456789abcdef"), false);
  assert.equal(
    bridge.wsHeaders["Authorization"],
    "Bearer test-only-cep-token-0123456789abcdef",
  );
});

test("resolves every typed export preset to its configured epr path", async () => {
  const presetPaths = {
    h264_1080p: "/presets/h264-1080.epr",
    h264_4k: "/presets/h264-4k.epr",
    prores_422: "/presets/prores-422.epr",
    prores_4444: "/presets/prores-4444.epr",
    dnx_hr: "/presets/dnx-hr.epr",
    custom: "/presets/custom.epr",
  } as const;
  const { bridge, calls } = bridgeWithHostResult(
    { jobID: "job-1", status: "export_queued", outputPath: "/tmp/out.mov" },
    { ...config, exportPresetPaths: presetPaths },
  );

  for (const preset of Object.keys(presetPaths) as Array<keyof typeof presetPaths>) {
    await bridge.exportSequence({
      sequenceId: "sequence-1",
      outputPath: "/tmp/out.mov",
      preset,
    });
  }

  assert.equal(calls.length, 6);
  calls.forEach((call, index) => {
    const preset = Object.keys(presetPaths)[index] as keyof typeof presetPaths;
    assert.equal(call.functionName, "exportSequence");
    assert.equal(call.args["sequenceId"], "sequence-1");
    assert.equal(call.args["presetPath"], presetPaths[preset]);
  });
});

test("rejects an unconfigured typed export preset", async () => {
  const { bridge, calls } = bridgeWithHostResult({});
  await assert.rejects(
    bridge.exportSequence({
      sequenceId: "sequence-1",
      outputPath: "/tmp/out.mp4",
      preset: "h264_1080p",
    }),
    /PREMIERE_EXPORT_PRESET_H264_1080P/,
  );
  assert.equal(calls.length, 0);
});
