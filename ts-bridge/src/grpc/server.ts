/**
 * gRPC server setup using nice-grpc.
 *
 * Creates the server, registers the PremiereBridgeService implementation,
 * and exposes a health-check endpoint via the standard gRPC health protocol.
 *
 * NOTE: Until the ts-proto generated definitions are available (via
 * `buf generate`), we load the proto file dynamically with
 * `@grpc/proto-loader` and cast the service definition so nice-grpc can
 * consume it.  This approach keeps the project runnable before code-gen is
 * wired into CI.
 */

import * as grpc from "@grpc/grpc-js";
import * as protoLoader from "@grpc/proto-loader";
import { createHash, timingSafeEqual } from "node:crypto";
import path from "node:path";
import { fileURLToPath } from "node:url";
import type { Logger } from "winston";
import type { BridgeConfig } from "../config.js";
import { resolveSharedToken } from "../auth/token.js";
import type { PremiereBridge } from "../bridge/interface.js";
import { createHandlers } from "./handlers.js";

// ---------------------------------------------------------------------------
// Resolve the .proto file relative to the monorepo root
// ---------------------------------------------------------------------------

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

/** Monorepo root — two levels up from ts-bridge/src/grpc/ */
const MONOREPO_ROOT = path.resolve(__dirname, "..", "..", "..");

const PROTO_DIR = path.join(MONOREPO_ROOT, "proto", "definitions");

const PREMIERE_PROTO = path.join(
  PROTO_DIR,
  "premierpro",
  "premiere",
  "v1",
  "premiere.proto",
);

// ---------------------------------------------------------------------------
// Server
// ---------------------------------------------------------------------------

export interface GrpcServer {
  /** Start listening. Resolves once the server is bound. */
  start(): Promise<void>;
  /** Graceful shutdown. */
  stop(): Promise<void>;
}

export async function createGrpcServer(
  config: BridgeConfig,
  bridge: PremiereBridge,
  logger: Logger,
): Promise<GrpcServer> {
  // Load the proto definition dynamically
  const packageDef = await protoLoader.load(PREMIERE_PROTO, {
    keepCase: false,
    longs: String,
    enums: Number,
    defaults: true,
    oneofs: true,
    includeDirs: [PROTO_DIR],
  });

  const grpcObject = grpc.loadPackageDefinition(packageDef);

  // Navigate to the service constructor
  const premierePkg = grpcObject["premierpro"] as Record<string, any>;
  const v1Pkg = premierePkg["premiere"]["v1"] as Record<string, any>;
  const ServiceConstructor = v1Pkg["PremiereBridgeService"] as grpc.ServiceClientConstructor;

  // Build handler map
  const handlers = createHandlers(bridge, logger);
  const authToken = resolveSharedToken(config.cepToken);

  // Create the raw gRPC server
  const server = new grpc.Server({
    "grpc.max_receive_message_length": 64 * 1024 * 1024, // 64 MB
    "grpc.max_send_message_length": 64 * 1024 * 1024,
  });

  // Register the PremiereBridgeService
  server.addService(ServiceConstructor.service, {
    getProjectState: wrapUnary(handlers.getProjectState, authToken),
    createSequence: wrapUnary(handlers.createSequence, authToken),
    getTimelineState: wrapUnary(handlers.getTimelineState, authToken),
    importMedia: wrapUnary(handlers.importMedia, authToken),
    placeClip: wrapUnary(handlers.placeClip, authToken),
    removeClip: wrapUnary(handlers.removeClip, authToken),
    addTransition: wrapUnary(handlers.addTransition, authToken),
    addText: wrapUnary(handlers.addText, authToken),
    applyEffect: wrapUnary(handlers.applyEffect, authToken),
    setAudioLevel: wrapUnary(handlers.setAudioLevel, authToken),
    exportSequence: wrapUnary(handlers.exportSequence, authToken),
    executeEdl: wrapUnary(handlers.executeEDL, authToken),
    evalCommand: wrapUnary(handlers.evalCommand, authToken),
    ping: wrapUnary(handlers.ping, authToken),
  });

  const bindAddress = `${config.grpcHost}:${config.grpcPort}`;

  return {
    async start() {
      return new Promise<void>((resolve, reject) => {
        server.bindAsync(
          bindAddress,
          grpc.ServerCredentials.createInsecure(),
          (err, port) => {
            if (err) {
              reject(err);
              return;
            }
            logger.info(`gRPC server listening on ${config.grpcHost}:${port}`);
            resolve();
          },
        );
      });
    },

    async stop() {
      return new Promise<void>((resolve) => {
        server.tryShutdown(() => {
          logger.info("gRPC server shut down");
          resolve();
        });
      });
    },
  };
}

// ---------------------------------------------------------------------------
// Helper: wrap an async handler into the callback-style that @grpc/grpc-js
// expects for unary RPCs.
// ---------------------------------------------------------------------------

function wrapUnary<TReq, TRes>(
  handler: (request: TReq) => Promise<TRes>,
  authToken: string,
): grpc.handleUnaryCall<TReq, TRes> {
  return (call, callback) => {
    if (!isAuthorizedMetadata(call.metadata, authToken)) {
      callback({
        code: grpc.status.UNAUTHENTICATED,
        details: "Missing or invalid bridge credentials",
        metadata: new grpc.Metadata(),
        name: "ServiceError",
        message: "Missing or invalid bridge credentials",
      });
      return;
    }
    handler(call.request)
      .then((result) => callback(null, result))
      .catch((err: unknown) => {
        const grpcError: grpc.ServiceError = {
          code: grpc.status.INTERNAL,
          details: err instanceof Error ? err.message : String(err),
          metadata: new grpc.Metadata(),
          name: "ServiceError",
          message: err instanceof Error ? err.message : String(err),
        };
        callback(grpcError);
      });
  };
}

export function isAuthorizedMetadata(
  metadata: grpc.Metadata,
  expectedToken: string,
): boolean {
  const values = metadata.get("authorization");
  if (values.length !== 1 || typeof values[0] !== "string") return false;
  const match = values[0].match(/^Bearer\s+(.+)$/i);
  if (!match) return false;

  const actualDigest = createHash("sha256").update(match[1] ?? "").digest();
  const expectedDigest = createHash("sha256").update(expectedToken).digest();
  return timingSafeEqual(actualDigest, expectedDigest);
}
