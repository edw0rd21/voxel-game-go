package main

import (
	"log"
	"runtime"

	"voxel-game/internal/camera"
	"voxel-game/internal/input"
	"voxel-game/internal/player"
	"voxel-game/internal/render"
	"voxel-game/internal/ui"
	"voxel-game/internal/world"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	windowWidth  = 1280
	windowHeight = 720
	windowTitle  = "Voxel Game"
)

func init() {
	// GLFW requires this to run on main thread
	runtime.LockOSThread()
}

func main() {
	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	// Configure GLFW
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Create window
	window, err := glfw.CreateWindow(windowWidth, windowHeight, windowTitle, nil, nil)
	if err != nil {
		log.Fatalln("failed to create window:", err)
	}
	window.MakeContextCurrent()

	// Control VSync
	glfw.SwapInterval(1) // 1 = VSync on, 0 = VSync off

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		log.Fatalln("failed to initialize OpenGL:", err)
	}

	// Configure OpenGL
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.ClearColor(0.53, 0.81, 0.92, 1.0) // Sky blue

	// Print OpenGL version
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version:", version)

	// Load Fonts
	pixelFont, err := ui.LoadFont("assets/fonts/PixelifySans-Regular.ttf", 24, false)
	if err != nil {
		log.Fatalf("Failed to load font: %v. Make sure assets/fonts/PixelifySans-Regular.ttf exists!", err)
	}
	cleanFont, err := ui.LoadFont("assets/fonts/Roboto-Bold.ttf", 24, true)
	if err != nil {
		log.Fatalf("Failed to load font: %v. Make sure assets/fonts/Roboto-Bold exists!", err)
	}

	// Load Texture Atlas (World)
	atlas, err := render.LoadTexture("assets/atlas.png")
	if err != nil {
		log.Fatalf("Failed to load texture atlas: %v", err)
	}
	log.Printf("Loaded atlas.png (ID: %d)", atlas.ID)

	// Initialize camera
	cam := camera.NewCamera(windowWidth, windowHeight)

	// Initialize renderer
	renderer, err := render.NewRenderer()
	if err != nil {
		log.Fatalln("failed to create renderer:", err)
	}

	// Initialize UI renderer
	uiRenderer, err := ui.NewUIRenderer(windowWidth, windowHeight)
	if err != nil {
		log.Fatalln("failed to create UI renderer:", err)
	}
	defer uiRenderer.Cleanup()

	// Add UI elements
	notifications := ui.NewNotificationSystem(cleanFont, windowWidth, windowHeight)
	if err := uiRenderer.AddElement(notifications); err != nil {
		log.Fatalln("failed to add notification system:", err)
	}
	notifications.Add("Welcome to Voxel Engine!")

	debugLayer := ui.NewDebugLayer(pixelFont, windowWidth, windowHeight)
	uiRenderer.AddElement(debugLayer)

	crosshair := ui.NewCrosshair(windowWidth, windowHeight)
	if err := uiRenderer.AddElement(crosshair); err != nil {
		log.Fatalln("failed to add crosshair:", err)
	}

	hotbar := ui.NewHotbar(windowWidth, windowHeight)
	if err := uiRenderer.AddElement(hotbar); err != nil {
		log.Fatalln("failed to add hotbar:", err)
	}

	// Window resize callback
	window.SetFramebufferSizeCallback(func(w *glfw.Window, width, height int) {
		gl.Viewport(0, 0, int32(width), int32(height))
		cam.SetSize(width, height)
		const targetUIHeight = 720.0
		uiScale := float32(height) / targetUIHeight

		logicalWidth := int(float32(width) / uiScale)
		logicalHeight := int(float32(height) / uiScale)
		uiRenderer.Resize(logicalWidth, logicalHeight)

		// Update UI elements with new size
		screenSize := &ui.ScreenSize{Width: logicalWidth, Height: logicalHeight}
		//notifications.Update(nil)
		crosshair.Update(screenSize)
		hotbar.Update(screenSize)
	})

	// Initialize world
	gameWorld := world.NewWorld()

	// Initialize player
	p := player.NewPlayer(cam, gameWorld)

	wireframeMode := false

	// Initialize input manager
	inputMgr := input.NewInputManager(window, cam, p, &wireframeMode)

	// Capture cursor
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	// Track delta time
	lastTime := glfw.GetTime()

	// FPS tracking
	frameCount := 0
	fpsTime := glfw.GetTime()
	currentFPS := 0.0

	// Chunk update throttling
	lastChunkUpdate := glfw.GetTime()
	chunkUpdateInterval := 0.5

	// Track selected block for hotbar
	var lastSelectedBlock world.BlockType = world.BlockAir

	// Game loop
	for !window.ShouldClose() {
		glfw.PollEvents()

		// Calculate delta time
		currentTime := glfw.GetTime()
		deltaTime := float32(currentTime - lastTime)
		lastTime = currentTime

		// FPS calculation
		frameCount++
		if currentTime-fpsTime >= 1.0 {
			currentFPS = float64(frameCount) / (currentTime - fpsTime)
			frameCount = 0
			fpsTime = currentTime
		}

		// Handle input
		inputMgr.Update(deltaTime)

		if inputMgr.IsActionJustPressed("TOGGLE_DEBUG") {
			// Toggle Persistent HUD
			isVisible := debugLayer.Toggle()

			// Trigger Transient Notification
			if isVisible {
				notifications.Add("Debug Mode: ON")
			} else {
				notifications.Add("Debug Mode: OFF")
			}
		}

		debugLayer.UpdateInfo(
			currentFPS,
			cam.Position,
			int(cam.Position[0])>>4,
			int(cam.Position[2])>>4,
			cam.Front,
		)
		debugLayer.Update(nil)

		notifications.Update(nil)

		// Update player - ONLY update player physics if NOT in debug mode
		if !inputMgr.IsDebugMode() {
			p.Update(deltaTime)
		}

		// Update world chunks based on player position
		if currentTime-lastChunkUpdate >= chunkUpdateInterval {
			gameWorld.UpdateChunks(cam.Position[0], cam.Position[2])
			lastChunkUpdate = currentTime
		}

		// Update hotbar if selected block changed
		selectedBlock := inputMgr.GetSelectedBlock()
		if selectedBlock != lastSelectedBlock {
			hotbar.Update(selectedBlock)
			lastSelectedBlock = selectedBlock
		}

		// Clear screen
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Render world
		renderer.RenderWorld(gameWorld, cam, atlas.ID)

		// Render block highlight
		target := p.TargetBlock()
		if target.Hit {
			renderer.DrawBlockHighlight(
				target.Pos,
				cam,
				mgl32.Vec3{1.0, 1.0, 1.0},
			)
		}

		gl.Disable(gl.DEPTH_TEST)
		gl.DepthMask(false)
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		gl.Disable(gl.CULL_FACE)
		// Render UI
		uiRenderer.Render()

		gl.Enable(gl.CULL_FACE)
		gl.DepthMask(true)
		gl.Enable(gl.DEPTH_TEST)

		// Swap buffers and poll events
		window.SwapBuffers()
	}
}
