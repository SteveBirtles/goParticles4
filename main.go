package main

import (
	"fmt"
	_ "image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"runtime"
	"strings"
	"time"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	windowWidth  = 1280
	windowHeight = 720
	numParticles = 1000000
)

var (
	frames            = 0
	second            = time.Tick(time.Second)
	windowTitlePrefix = "Particles"
	vao               uint32
)

func init() {

	runtime.LockOSThread()

}

func LoadShader(path string, shaderType uint32) uint32 {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	source := string(bytes) + "\x00"

	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		panic(fmt.Errorf("failed to compile %v: %v", source, log))
	}

	return shader
}

func main() {

	var err error
	if err = glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(windowWidth, windowHeight, windowTitlePrefix, nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	if err = gl.Init(); err != nil {
		panic(err)
	}
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	glfw.SwapInterval(0)

	particleShader := LoadShader("shaders/particles.glsl", gl.COMPUTE_SHADER)
	vertexShader := LoadShader("shaders/vert.glsl", gl.VERTEX_SHADER)
	fragmentShader := LoadShader("shaders/frag.glsl", gl.FRAGMENT_SHADER)

	particleProg := gl.CreateProgram()
	gl.AttachShader(particleProg, particleShader)
	gl.LinkProgram(particleProg)
	gl.UseProgram(particleProg)

	var positionBuffer, velocityBuffer, colourBuffer, attractorBuffer uint32

	var points, velocities, colours, attractors []mgl32.Vec4

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < numParticles; i++ {
		x := (rand.Float32()*2 - 1) * float32(32)
		y := (rand.Float32()*2 - 1) * float32(32)
		z := (rand.Float32()*2 - 1) * float32(32)
		points = append(points, mgl32.Vec4{x, y, z, 1})
		velocities = append(velocities, mgl32.Vec4{})
		colours = append(colours, mgl32.Vec4{})
	}

	for i := 0; i < 6; i++ {
		x := (rand.Float32()*2 - 1) * float32(50)
		y := (rand.Float32()*2 - 1) * float32(50)
		z := (rand.Float32()*2 - 1) * float32(50)
		g := -rand.Float32() * float32(10)
		attractors = append(attractors, mgl32.Vec4{x, y, z, g})
	}

	gl.GenBuffers(1, &positionBuffer)
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, positionBuffer)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, numParticles*16, gl.Ptr(points), gl.DYNAMIC_DRAW)
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 0, positionBuffer)

	gl.GenBuffers(1, &velocityBuffer)
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, velocityBuffer)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, numParticles*16, gl.Ptr(velocities), gl.DYNAMIC_DRAW)
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 1, velocityBuffer)

	gl.GenBuffers(1, &colourBuffer)
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, colourBuffer)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, numParticles*16, gl.Ptr(colours), gl.DYNAMIC_DRAW)
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 2, colourBuffer)

	gl.GenBuffers(1, &attractorBuffer)
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, attractorBuffer)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, 6*16, gl.Ptr(attractors), gl.DYNAMIC_DRAW)
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 3, attractorBuffer)

	quadProg := gl.CreateProgram()
	gl.AttachShader(quadProg, vertexShader)
	gl.AttachShader(quadProg, fragmentShader)
	gl.LinkProgram(quadProg)

	gl.UseProgram(quadProg)

	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, positionBuffer)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 0, nil)

	gl.EnableVertexAttribArray(3)
	gl.BindBuffer(gl.ARRAY_BUFFER, colourBuffer)
	gl.VertexAttribPointer(3, 4, gl.FLOAT, false, 0, nil)

	position := mgl32.Vec3{0, 0, 100}
	target := mgl32.Vec3{0, 0, 0}
	up := mgl32.Vec3{0, 1, 0}
	view := mgl32.LookAtV(position, target, up)
	projection := mgl32.Perspective(mgl32.DegToRad(60), float32(windowWidth)/float32(windowHeight), 0.1, 1000.0)

	projUniform := int32(1)
	gl.UniformMatrix4fv(projUniform, 1, false, &projection[0])

	viewUniform := int32(2)
	gl.UniformMatrix4fv(viewUniform, 1, false, &view[0])

	for !window.ShouldClose() {

		if window.GetKey(glfw.KeyEscape) == glfw.Press {
			window.SetShouldClose(true)
		}

		/* --------------------------- */

		gl.UseProgram(particleProg)
		gl.DispatchCompute(1024, 1, 1)
		gl.MemoryBarrier(gl.VERTEX_ATTRIB_ARRAY_BARRIER_BIT)

		gl.UseProgram(quadProg)
		gl.ClearColor(0, 0, 0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.POINTS, 0, numParticles)

		/* --------------------------- */

		gl.UseProgram(0)
		window.SwapBuffers()

		glfw.PollEvents()
		frames++
		select {
		case <-second:
			window.SetTitle(fmt.Sprintf("%s | FPS: %d", windowTitlePrefix, frames))
			frames = 0
		default:
		}

	}

	glfw.Terminate()
}
