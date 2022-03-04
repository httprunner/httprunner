from concurrent import futures
import sys
import time

import grpc

import debugtalk_pb2
import debugtalk_pb2_grpc

from grpc_health.v1.health import HealthServicer
from grpc_health.v1 import health_pb2, health_pb2_grpc

class DebugTalkServicer(debugtalk_pb2_grpc.DebugTalkServicer):
    """Implementation of DebugTalk service."""

    def GetNames(self, request, context):
        result = debugtalk_pb2.GetNamesResponse()
        return result

    def Call(self, request, context):
        return

def serve():
    # We need to build a health service to work with go-plugin
    health = HealthServicer()
    health.set("plugin", health_pb2.HealthCheckResponse.ServingStatus.Value('SERVING'))

    # Start the server.
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    debugtalk_pb2_grpc.add_DebugTalkServicer_to_server(DebugTalkServicer(), server)
    health_pb2_grpc.add_HealthServicer_to_server(health, server)
    server.add_insecure_port('127.0.0.1:1234')
    server.start()

    # Output information
    print("1|1|tcp|127.0.0.1:1234|grpc")
    sys.stdout.flush()

    try:
        while True:
            time.sleep(60 * 60 * 24)
    except KeyboardInterrupt:
        server.stop(0)

if __name__ == '__main__':
    serve()
