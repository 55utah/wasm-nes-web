# wasm-fc
> 基于go语言实现的fc模拟器项目创建的项目，项目最终编译生成wasm，并运行在浏览器上，实现fc模拟器web化；
> 建议优先使用safari浏览器/火狐浏览器，chrome浏览器某些情况下更容易卡顿；

### 在线体验地址
https://55utah.github.io/wasm-nes/index.html

### 效果
<img src="https://user-images.githubusercontent.com/17704150/147459628-64bc0fe8-8f28-45fd-9f12-39ceedd6fdbb.png" width="300" />

### 音效
支持音效，部分游戏音效效果不太好，待优化。

### 组成
wasm下是go文件，用于编译生成wasm。
web下是web项目。

### 编译
`GOOS=js GOARCH=wasm go build -o nes.wasm wasm/wasm.go`
会生成nes.wasm，放在web/src内给web项目使用。

### 调试
`go run web/server.go`
