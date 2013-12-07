package main

import (
	"errors"
	"fmt"
	"github.com/Ferguzz/glam"
	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	"io/ioutil"
	"runtime"
	"time"
)

// Window stuff.
const (
	height   = 640.0
	width    = 480.0
	aspect   = width / height
	gridSize = 20
	speed    = 2
)

// Controls.
const (
	down = iota
	left
	right
)

// The falling block.
var shaderProgram gl.Program
var generateNewBlock chan bool = make(chan bool, 1)

func run() {
	// All OpenGL related stuff needs to be called from the same goroutine.  The mainThread function achieves this.
	mainThread(func() {
		glfw.SetErrorCallback(errorCallback)

		if !glfw.Init() {
			panic("Can't init glfw!")
		}
		defer glfw.Terminate()

		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 2)
		glfw.WindowHint(glfw.OpenglProfile, glfw.OpenglCoreProfile)
		glfw.WindowHint(glfw.OpenglForwardCompatible, gl.TRUE)

		window, err := glfw.CreateWindow(width, height, "Tetris", nil, nil)
		if err != nil {
			panic(err)
		}

		window.MakeContextCurrent()
		window.SetKeyCallback(keyCallback)

		if gl.Init() != 0 {
			panic("Can't initialise OpenGL.")
		}

		vertexShader, err := loadShader(gl.VERTEX_SHADER, "block_shader.vert")
		if err != nil {
			panic(err)
		}
		defer vertexShader.Delete()

		fragmentShader, err := loadShader(gl.FRAGMENT_SHADER, "block_shader.frag")
		if err != nil {
			panic(err)
		}
		defer fragmentShader.Delete()

		shaderProgram = gl.CreateProgram()
		shaderProgram.AttachShader(vertexShader)
		shaderProgram.AttachShader(fragmentShader)
		shaderProgram.BindFragDataLocation(0, "outColor")
		shaderProgram.Link()
		shaderProgram.Use()
		defer shaderProgram.Delete()

		projection := glam.Orthographic(height, aspect, -1, 1)
		shaderProgram.GetUniformLocation("projection").UniformMatrix4fv(false, projection)
		shaderProgram.GetUniformLocation("scale").Uniform1f(gridSize)
		gl.ClearColor(0, 0, 0, 1)

		// Create the first block.
		NewBlock()

		// Set off a thread to cause the active block to fall.
		go func() {
			for _ = range time.Tick(time.Second / speed) {
				blocks[len(blocks)-1].Move(down)
			}
		}()

		for !window.ShouldClose() {
			gl.Clear(gl.COLOR_BUFFER_BIT)

			// Check for messages requesting a new block.
			select {
			case <-generateNewBlock:
				NewBlock()
			default:
			}

			for _, block := range blocks {
				block.Draw()
			}

			window.SwapBuffers()
			glfw.PollEvents()
		}
	})
	close(mainfunc)
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press || action == glfw.Repeat {
		switch key {
		case glfw.KeyQ:
			window.SetShouldClose(true)
		case glfw.KeyLeft:
			blocks[len(blocks)-1].Move(left)
		case glfw.KeyRight:
			blocks[len(blocks)-1].Move(right)
		case glfw.KeyDown:
			blocks[len(blocks)-1].Move(down)
		case glfw.KeyR:
			blocks[len(blocks)-1].Rotate()
		}
	}
}

func loadShader(shaderType gl.GLenum, source string) (gl.Shader, error) {
	shader := gl.CreateShader(shaderType)

	shaderSource, err := ioutil.ReadFile(source)
	if err != nil {
		return shader, err
	}

	shader.Source(string(shaderSource))
	shader.Compile()

	if shader.Get(gl.COMPILE_STATUS) == 0 {
		return shader, errors.New(fmt.Sprintf("Shader (%v) did not compile.", source))
	}

	return shader, nil
}

func errorCallback(err glfw.ErrorCode, desc string) {
	fmt.Printf("%v: %v\n", err, desc)
}

func init() {
	runtime.LockOSThread()
}

func Main() {
	for f := range mainfunc {
		f()
	}
}

var mainfunc = make(chan func())

func mainThread(f func()) {
	done := make(chan bool, 1)
	mainfunc <- func() {
		f()
		done <- true
	}
	<-done
}

func main() {
	go run()
	Main()
}
