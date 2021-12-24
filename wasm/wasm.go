package main

import (
	"image"
	"syscall/js"
	"time"

	"github.com/55utah/fc-simulator/nes"
)

var console *nes.Console

var timestamp float64

// dom canvas ctx对象
var ctx js.Value
var canvas js.Value
var document js.Value
var ratio = 1
var width = 256
var height = 240

var ctrl1 [8]bool
var ctrl2 [8]bool

// 写死一个最大值
const FrameSecond float64 = 0.0167

func floatSecond() float64 {
	return float64(time.Now().Nanosecond()) * float64(1e-9)
}

func setSize(ratio int) {
	if ratio > 1 && ratio < 5 {
		ratio = ratio
	}
	canvas.Set("width", width*ratio)
	canvas.Set("height", height*ratio)
}

func resize(source *image.RGBA, w int, h int, ratio int) *image.RGBA {

	tw := w * ratio
	th := h * ratio

	var target *image.RGBA = image.NewRGBA(image.Rect(0, 0, tw, th))

	for y := 0; y < th; y++ {
		for x := 0; x < tw; x++ {
			sx := x / ratio
			sy := y / ratio
			target.SetRGBA(x, y, source.RGBAAt(sx, sy))
		}
	}

	return target
}

// 将js更新canvas改为go修改dom，执行更快，cpu占用更小！！
func render(value []byte, width int, height int) {
	imageData := ctx.Call("getImageData", 0, 0, width, height)
	buf := js.Global().Get("Uint8ClampedArray").New(width * height * 4)

	dst := js.Global().Get("Uint8Array").New(len(value))
	js.CopyBytesToJS(dst, value)
	buf.Call("set", dst)
	imageData.Get("data").Call("set", buf)
	ctx.Call("putImageData", imageData, 0, 0)
}

func onFrame() {
	current := floatSecond()
	cost := current - timestamp
	if cost > FrameSecond {
		cost = FrameSecond
	} else if cost < 0 {
		cost = 0
	}
	timestamp = current
	console.StepSeconds(cost)
	buffer := console.Buffer()
	newBuffer := resize(buffer, width, height, ratio)
	render(newBuffer.Pix, width*ratio, height*ratio)
}

func global() js.Value {
	return js.Global().Get("customConsole")
}

// func outputAudio(v float32) {
// 	audio := global().Get("outputAudio")
// 	audio.Invoke(v)
// }

func newConsole(file js.Value) {
	game := make([]byte, file.Get("length").Int())
	js.CopyBytesToGo(game, file)
	console, _ = nes.NewConsole(game)
	// console.SetAudioOutputWork(outputAudio)

	document = js.Global().Get("document")
	canvas = document.Call("querySelector", "canvas")
	ctx = canvas.Call("getContext", "2d")

	handleKeyEvent()
}

func handleKeyEvent() {
	onkeydownCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		handleKey(e.Get("code").String(), true)
		return nil
	})
	document.Set("onkeydown", onkeydownCallback)

	onkeyupCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		handleKey(e.Get("code").String(), false)
		return nil
	})
	document.Set("onkeyup", onkeyupCallback)
}

func handleKey(code string, down bool) {

	if console == nil {
		return
	}

	if down {
		keyParseSys(code, func() {
			canvas.Set("width", width*ratio)
			canvas.Set("height", height*ratio)
		})
	}

	index1 := keyPress1(code)
	index2 := keyPress2(code)

	if index1 >= 0 {
		ctrl1[index1] = down
		console.SetButton1(ctrl1)
	}
	if index2 >= 0 {
		ctrl2[index2] = down
		console.SetButton2(ctrl2)
	}
}

func reset() {
	console.Reset()
}

// func setController(ctrl js.Value, btnIndex int) {
// 	if console == nil {
// 		return
// 	}
// 	if ctrl.Get("length").Int() != 8 {
// 		return
// 	}
// 	list := make([]byte, 8)
// 	js.CopyBytesToGo(list, ctrl)

// 	btn := [8]bool{}
// 	for i, k := range list {
// 		if k > 0 {
// 			btn[i] = true
// 		} else {
// 			btn[i] = false
// 		}
// 	}

// 	if btnIndex == 0 {
// 		console.SetButton1(btn)
// 	} else {
// 		console.SetButton2(btn)
// 	}
// }

func main() {
	println("Hello, fc!")

	// newConsole
	global().Set("newConsole", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		newConsole(args[0])
		return nil
	}))

	// // reset
	// global().Set("reset", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	// 	reset()
	// 	return nil
	// }))

	// onFrame
	global().Set("frame", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		onFrame()
		return nil
	}))

	// setSampleRate
	// global().Set("setSampleRate", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	// 	sampleRate := args[0].Int()
	// 	fmt.Printf("sampleRate %f \n", float64(sampleRate))
	// 	console.SetAudioSampleRate(float64(sampleRate))
	// 	return nil
	// }))

	// setController1
	// global().Set("setController1", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	// 	setController(args[0], 0)
	// 	return nil
	// }))

	// // setController2
	// global().Set("setController2", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	// 	setController(args[0], 1)
	// 	return nil
	// }))

	channel := make(chan int, 0)
	<-channel
	println("finished!")
}

func keyParseSys(keyName string, resizeWindow func()) {
	switch keyName {
	// 重置游戏
	case "KeyQ":
		console.Reset()
	// 缩小屏幕
	case "Minus":
		if ratio > 1 {
			ratio--
			resizeWindow()
		}
	// 放大屏幕
	case "Equal":
		if ratio < 4 {
			ratio++
			resizeWindow()
		}
	}
}

func keyPress1(code string) int {
	index := -1
	switch code {
	// A
	case "KeyF":
		index = 0
		break
	// B
	case "KeyG":
		index = 1
		break
	// Select
	case "KeyR":
		index = 2
		break
	// Start
	case "KeyT":
		index = 3
		break
	// 上
	case "KeyW":
		index = 4
		break
	// 下
	case "KeyS":
		index = 5
		break
	// 左
	case "KeyA":
		index = 6
		break
	// 右
	case "KeyD":
		index = 7
		break
	}
	return index
}

func keyPress2(code string) int {
	index := -1
	switch code {
	// A
	case "KeyJ":
		index = 0
		break
	// B
	case "KeyK":
		index = 1
		break
	// Select
	case "KeyU":
		index = 2
		break
	// Start
	case "KeyI":
		index = 3
		break
	// 上
	case "ArrowUp":
		index = 4
		break
	// 下
	case "ArrowDown":
		index = 5
		break
	// 左
	case "ArrowLeft":
		index = 6
		break
	// 右
	case "ArrowRight":
		index = 7
		break
	}
	return index
}
