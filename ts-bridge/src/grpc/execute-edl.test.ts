import * as grpc from "@grpc/grpc-js";
import * as protoLoader from "@grpc/proto-loader";
import assert from "node:assert/strict";
import { once } from "node:events";
import { createServer, type AddressInfo } from "node:net";
import path from "node:path";
import test, { type TestContext } from "node:test";
import { fileURLToPath } from "node:url";
import winston from "winston";

import { CepBridge } from "../bridge/cep-bridge.js";
import type { BridgeConfig } from "../config.js";
import { createGrpcServer } from "./server.js";

const token = "execute-edl-test-token-0123456789abcdef";
const protoRoot = fileURLToPath(new URL("../../../proto/definitions/", import.meta.url));

interface EdlClient extends grpc.Client {
  executeEdl(
    request: object,
    metadata: grpc.Metadata,
    callback: (error: grpc.ServiceError | null, response: Record<string, unknown>) => void,
  ): void;
}

async function unusedPort(): Promise<number> {
  const server = createServer();
  server.listen(0, "127.0.0.1");
  await once(server, "listening");
  const port = (server.address() as AddressInfo).port;
  await new Promise<void>((resolve, reject) => server.close(error => error ? reject(error) : resolve()));
  return port;
}

async function executeEdlHarness(t: TestContext): Promise<{
  calls: Array<{ functionName: string; args: Record<string, unknown> }>;
  invoke(request: object): Promise<Record<string, unknown>>;
}> {
  const config: BridgeConfig = {
    grpcPort: await unusedPort(), grpcHost: "127.0.0.1", premierePath: "/test/Premiere.app",
    bridgeMode: "cep", logLevel: "error", cepWsPort: 9801, cepToken: token,
  };
  const calls: Array<{ functionName: string; args: Record<string, unknown> }> = [];
  const bridge = new CepBridge(config);
  // This is the last transport boundary before CEP calls the native dispatcher.
  // Enforce the real host wrapper contract; never create or mutate a project.
  bridge.evalCommand = async (functionName, argsJson) => {
    const args = JSON.parse(argsJson) as Record<string, unknown>;
    calls.push({ functionName, args });
    const edl = args["edl"] as { entries?: unknown[] } | undefined;
    if (!Array.isArray(edl?.entries) || edl.entries.length === 0) {
      return { resultJson: "", isError: true, errorMessage: "mcpExecuteEDL requires edl.entries" };
    }
    return {
      resultJson: JSON.stringify({ success: true, data: {
        sequenceId: "transport-test-sequence", status: "completed", clipsPlaced: edl.entries.length,
        transitionsAdded: 0, errors: [], warnings: ["transport test only"],
      } }),
      isError: false, errorMessage: "",
    };
  };
  const logger = winston.createLogger({ silent: true });
  const server = await createGrpcServer(config, bridge, logger);
  await server.start();
  const definition = protoLoader.loadSync(path.join(protoRoot, "premierpro/premiere/v1/premiere.proto"), {
    keepCase: false, longs: String, enums: Number, defaults: true, oneofs: true, includeDirs: [protoRoot],
  });
  const packages = grpc.loadPackageDefinition(definition) as unknown as {
    premierpro: { premiere: { v1: { PremiereBridgeService: new (
      address: string, credentials: grpc.ChannelCredentials,
    ) => EdlClient } } };
  };
  const client = new packages.premierpro.premiere.v1.PremiereBridgeService(
    `${config.grpcHost}:${config.grpcPort}`, grpc.credentials.createInsecure(),
  );
  t.after(async () => {
    client.close();
    await server.stop();
    await bridge.disconnect();
  });
  const metadata = new grpc.Metadata();
  metadata.set("authorization", `Bearer ${token}`);
  return {
    calls,
    invoke: request => new Promise((resolve, reject) => {
      client.executeEdl(request, metadata, (error, response) => error ? reject(error) : resolve(response));
    }),
  };
}

test("protobuf ExecuteEDL preserves the CEP wrapper, entries and distinct clip ranges", async t => {
  const { calls, invoke } = await executeEdlHarness(t);
  const sourceRange = {
    inPoint: { hours: 1, minutes: 2, seconds: 3, frames: 7, frameRate: 23.976 },
    outPoint: { hours: 1, minutes: 2, seconds: 7, frames: 8, frameRate: 23.976 },
  };
  const timelineRange = {
    inPoint: { hours: 0, minutes: 0, seconds: 20, frames: 9, frameRate: 29.97 },
    outPoint: { hours: 0, minutes: 0, seconds: 24, frames: 10, frameRate: 29.97 },
  };
  const response = await invoke({
    edl: {
      id: "transport-edl", name: "EDL with a source path",
      sequenceResolution: { width: 320, height: 180 }, sequenceFrameRate: 24,
      entries: [{ index: 9, sourceAssetId: "/tmp/source with spaces.wav", sourceRange, timelineRange,
        track: { type: 2, trackIndex: 1 }, effects: [], notes: "exact clip range" }],
    },
    autoImport: true, autoCreateSequence: true,
  });
  assert.equal(calls.length, 1);
  assert.equal(calls[0]?.functionName, "mcpExecuteEDL");
  const args = calls[0]!.args;
  assert.equal(args["autoImport"], true);
  assert.equal(args["autoCreateSequence"], true);
  const edl = args["edl"] as { entries: Array<Record<string, unknown>>; sequenceFrameRate: number };
  assert.equal(edl.sequenceFrameRate, 24);
  assert.equal(edl.entries.length, 1);
  assert.equal(edl.entries[0]?.["sourceAssetId"], "/tmp/source with spaces.wav");
  assert.deepEqual(edl.entries[0]?.["sourceRange"], sourceRange);
  assert.deepEqual(edl.entries[0]?.["timelineRange"], timelineRange);
  assert.deepEqual(edl.entries[0]?.["track"], { type: 2, trackIndex: 1 });
  assert.equal(response["sequenceId"], "transport-test-sequence");
  assert.equal(response["clipsPlaced"], 1);
  assert.equal(response["status"], 3); // OPERATION_STATUS_COMPLETED
  assert.deepEqual(response["errors"], []);
});

test("an empty protobuf EDL surfaces the host error rather than a completed response", async t => {
  const { calls, invoke } = await executeEdlHarness(t);
  await assert.rejects(invoke({ edl: { name: "Empty edit", entries: [] }, autoImport: true }),
    (error: unknown) => error instanceof Error && error.message.includes("requires edl.entries"));
  assert.equal(calls.length, 1);
});
