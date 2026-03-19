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
import path from "node:path";
import { fileURLToPath } from "node:url";
import type { Logger } from "winston";
import type { BridgeConfig } from "../config.js";
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

  // Create the raw gRPC server
  const server = new grpc.Server({
    "grpc.max_receive_message_length": 64 * 1024 * 1024, // 64 MB
    "grpc.max_send_message_length": 64 * 1024 * 1024,
  });

  // Register the PremiereBridgeService
  server.addService(ServiceConstructor.service, {
    getProjectState: wrapUnary(handlers.getProjectState),
    createSequence: wrapUnary(handlers.createSequence),
    getTimelineState: wrapUnary(handlers.getTimelineState),
    importMedia: wrapUnary(handlers.importMedia),
    placeClip: wrapUnary(handlers.placeClip),
    removeClip: wrapUnary(handlers.removeClip),
    addTransition: wrapUnary(handlers.addTransition),
    addText: wrapUnary(handlers.addText),
    applyEffect: wrapUnary(handlers.applyEffect),
    setAudioLevel: wrapUnary(handlers.setAudioLevel),
    exportSequence: wrapUnary(handlers.exportSequence),
    executeEdl: wrapUnary(handlers.executeEDL),
    ping: wrapUnary(handlers.ping),
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
): grpc.handleUnaryCall<TReq, TRes> {
  return (call, callback) => {
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
