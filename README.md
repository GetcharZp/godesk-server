# GoDesk Server

> Remote Desktop Server

### Core Dependencies

+ [grpc](https://github.com/grpc/grpc-go)

### Dev

+ 项目运行

```shell
go run main.go
```

+ proto文件生成

```shell
protoc -I ./proto --go_out=./proto/ --go_opt=paths=source_relative \
 --go-grpc_out=./proto/ --go-grpc_opt=require_unimplemented_servers=false \
 --go-grpc_opt=paths=source_relative ./proto/*.proto
```

### Build

```shell
go build main.go
```
