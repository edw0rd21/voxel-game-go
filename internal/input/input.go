package input

import (
	"voxel-game/internal/camera"
	"voxel-game/internal/player"
	"voxel-game/internal/world"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type InputManager struct {
	window *glfw.Window
	camera *camera.Camera
	player *player.Player

	firstMouse bool
	lastX      float64
	lastY      float64

	selectedBlock world.BlockType
	cursorLocked  bool
	wireframe     *bool
}

func NewInputManager(window *glfw.Window, cam *camera.Camera, p *player.Player, wireframe *bool) *InputManager {
	im := &InputManager{
		window:        window,
		camera:        cam,
		player:        p,
		firstMouse:    true,
		selectedBlock: world.BlockDirt,
		cursorLocked:  true,
		wireframe:     wireframe,
	}

	// Set up callbacks
	window.SetCursorPosCallback(im.mouseCallback)
	window.SetMouseButtonCallback(im.mouseButtonCallback)
	window.SetKeyCallback(im.keyCallback)

	return im
}

func (im *InputManager) GetSelectedBlock() world.BlockType {
	return im.selectedBlock
}

func (im *InputManager) Update(deltaTime float32) {
	// Movement
	var moveDir mgl32.Vec3

	if im.window.GetKey(glfw.KeyW) == glfw.Press {
		forward := im.camera.Front
		forward[1] = 0 // Don't move up/down with camera pitch
		forward = forward.Normalize()
		moveDir = moveDir.Add(forward)
	}
	if im.window.GetKey(glfw.KeyS) == glfw.Press {
		forward := im.camera.Front
		forward[1] = 0
		forward = forward.Normalize()
		moveDir = moveDir.Sub(forward)
	}
	if im.window.GetKey(glfw.KeyA) == glfw.Press {
		moveDir = moveDir.Sub(im.camera.Right)
	}
	if im.window.GetKey(glfw.KeyD) == glfw.Press {
		moveDir = moveDir.Add(im.camera.Right)
	}

	if moveDir.Len() > 0 {
		moveDir = moveDir.Normalize()
		im.player.Move(moveDir)
	}

	// Jump
	if im.window.GetKey(glfw.KeySpace) == glfw.Press {
		im.player.Jump()
	}

	// Close window
	if im.window.GetKey(glfw.KeyEscape) == glfw.Press {
		im.window.SetShouldClose(true)
	}
}

func (im *InputManager) mouseCallback(w *glfw.Window, xpos, ypos float64) {
	if !im.cursorLocked {
		return
	}
	if im.firstMouse {
		im.lastX = xpos
		im.lastY = ypos
		im.firstMouse = false
	}

	xoffset := xpos - im.lastX
	yoffset := im.lastY - ypos // Reversed since y-coordinates go from bottom to top

	im.lastX = xpos
	im.lastY = ypos

	im.camera.ProcessMouseMovement(float32(xoffset), float32(yoffset))
}

func (im *InputManager) mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		if button == glfw.MouseButtonLeft {
			// Break block
			im.player.BreakBlock()
		} else if button == glfw.MouseButtonRight {
			// Place block
			im.player.PlaceBlock(im.selectedBlock)
		}
	}
}

func (im *InputManager) keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		// Number keys to select block type
		switch key {
		case glfw.Key1:
			im.selectedBlock = world.BlockDirt
		case glfw.Key2:
			im.selectedBlock = world.BlockGrass
		case glfw.Key3:
			im.selectedBlock = world.BlockStone
		case glfw.KeyTab:
			im.cursorLocked = !im.cursorLocked
			if im.cursorLocked {
				w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
			} else {
				w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
			}
		case glfw.KeyF:
			*im.wireframe = !*im.wireframe
			if *im.wireframe {
				gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
			} else {
				gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
			}
		case glfw.KeyN:
			im.player.NoClip()
		}
	}
}
