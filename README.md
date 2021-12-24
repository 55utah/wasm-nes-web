# wasm-fc
> go语言实现的fc模拟器编译的wasm项目，在浏览器运行fc模拟器。

### 组成
wasm下是go文件，用于编译生成wasm。
web下是web项目。
### 编译
`GOOS=js GOARCH=wasm go build -o nes.wasm wasm/wasm.go`
会生成nes.wasm，放在web/src内。

### 调试
`go run web/server.go`
