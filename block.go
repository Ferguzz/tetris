package main

import (
	"github.com/Ferguzz/glam"
	"github.com/go-gl/gl"
	"math"
	"math/rand"
	"time"
)

type Block struct {
	vao           gl.VertexArray
	vbo           gl.Buffer
	elementBuffer gl.Buffer
	numElements   int
	position      glam.Vec3
	orientation   int
	color         Color
}

var blocks []Block

type Shape struct {
	vertices []gl.GLfloat
	elements []gl.GLushort
}

type Color []float32

var square Shape = Shape{[]gl.GLfloat{-1, -1, 1, -1, -1, 1, 1, 1}, []gl.GLushort{0, 1, 2, 2, 3, 1}}
var L Shape = Shape{[]gl.GLfloat{-2, 0, -2, -1, 2, -1, 2, 0, 2, 1, 1, 1, 1, 0}, []gl.GLushort{0, 1, 2, 2, 3, 0, 3, 4, 5, 5, 6, 3}}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomColor() Color {
	color := Color{1, 1, 1}
	switch rand.Intn(5) {
	case 0:
		// Red.
		color = Color{1, 0, 0}
	case 1:
		// Green.
		color = Color{0, 1, 0}
	case 3:
		// Blue.
		color = Color{0, 0, 1}
	case 4:
		// White.
		color = Color{1, 1, 1}
	}

	return color
}

func randomShape() Shape {
	var shape Shape
	switch rand.Intn(2) {
	case 0:
		shape = square
	case 1:
		shape = L
	}
	return shape
}

func NewBlock() {

	// I only need one vao per block type, not per actual block.

	var block Block
	shape := randomShape()
	block.vao = gl.GenVertexArray()
	block.vao.Bind()
	block.vbo = gl.GenBuffer()
	block.vbo.Bind(gl.ARRAY_BUFFER)
	gl.BufferData(gl.ARRAY_BUFFER, len(shape.vertices)*4, shape.vertices, gl.STATIC_DRAW)
	block.elementBuffer = gl.GenBuffer()
	block.elementBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(shape.elements)*2, shape.elements, gl.STATIC_DRAW)
	block.numElements = len(shape.elements)

	vertexAttribArray := shaderProgram.GetAttribLocation("position")
	vertexAttribArray.AttribPointer(2, gl.FLOAT, false, 0, uintptr(0))
	vertexAttribArray.EnableArray()

	// Initialise a random X position.
	block.position = glam.Vec3{0, 0, 0}
	// Pick a random orientation.
	block.orientation = rand.Intn(4)
	// Finally, a random color.
	// block.color = Color{1, 0, 0}
	block.color = randomColor()

	blocks = append(blocks, block)
}

func (block *Block) Delete() {
	block.vao.Delete()
	block.vbo.Delete()
	block.elementBuffer.Delete()
}

func (block *Block) Draw() {
	block.vao.Bind()
	position := glam.Translation(block.position)
	rotation := glam.Rotation(float32(block.orientation)*math.Pi/2.0, glam.Vec3{0, 0, 1})
	var model glam.Mat4
	model.Multiply(&rotation, &position)
	shaderProgram.GetUniformLocation("model").UniformMatrix4fv(false, model)
	shaderProgram.GetUniformLocation("inColor").Uniform3fv(1, block.color)
	gl.DrawElements(gl.TRIANGLES, block.numElements, gl.UNSIGNED_SHORT, uintptr(0))
}

func (block *Block) Move(direction int) {
	// Do collision detection in here.
	switch direction {
	case down:
		if block.position.Y >= -(height/2)+40 {
			block.position.Y -= gridSize
		} else {
			// We've hit the bottom.  Generate a new block.
			// Since block creation involves OpenGL related calls, it needs to be done in the main thread.
			// Since this function can be called from other threads, we need to signal to the OpenGL thread to make a block.
			generateNewBlock <- true
		}
	case left:
		if block.position.X >= -width/2+40 {
			block.position.X -= gridSize
		}
	case right:
		if block.position.X <= width/2-40 {
			block.position.X += gridSize
		}
	}
}

func (block *Block) Rotate() {
	block.orientation += 1
}
