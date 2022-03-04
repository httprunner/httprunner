
## Updating the Protocol

If you update the protocol buffers file, you can regenerate the file using the following command from the project root directory. You do not need to run this if you're just using the plugin.

For Go:

```bash
$ protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative plugin/proto/debugtalk.proto
```
