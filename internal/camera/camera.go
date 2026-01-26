package camera

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

type Camera struct {
	Position mgl32.Vec3
	Front    mgl32.Vec3
	Up       mgl32.Vec3
	Right    mgl32.Vec3
	WorldUp  mgl32.Vec3

	Yaw   float32
	Pitch float32

	MovementSpeed    float32
	MouseSensitivity float32
	Fov              float32

	width  int
	height int
}

func NewCamera(width, height int) *Camera {
	c := &Camera{
		Position:         mgl32.Vec3{0, 70, 0},
		WorldUp:          mgl32.Vec3{0, 1, 0},
		Yaw:              -90.0,
		Pitch:            0.0,
		MovementSpeed:    10.0,
		MouseSensitivity: 0.1,
		Fov:              45.0,
		width:            width,
		height:           height,
	}
	c.updateCameraVectors()
	return c
}

func (c *Camera) GetViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(c.Position, c.Position.Add(c.Front), c.Up)
}

func (c *Camera) GetProjectionMatrix() mgl32.Mat4 {
	return mgl32.Perspective(
		mgl32.DegToRad(c.Fov),
		float32(c.width)/float32(c.height),
		0.1,
		1000.0,
	)
}

func (c *Camera) ProcessMouseMovement(xoffset, yoffset float32) {
	xoffset *= c.MouseSensitivity
	yoffset *= c.MouseSensitivity

	c.Yaw += xoffset
	c.Pitch += yoffset

	// Constrain pitch
	if c.Pitch > 89.0 {
		c.Pitch = 89.0
	}
	if c.Pitch < -89.0 {
		c.Pitch = -89.0
	}

	c.updateCameraVectors()
}

func (c *Camera) updateCameraVectors() {
	// Calculate new Front vector
	front := mgl32.Vec3{
		float32(math.Cos(float64(mgl32.DegToRad(c.Yaw))) * math.Cos(float64(mgl32.DegToRad(c.Pitch)))),
		float32(math.Sin(float64(mgl32.DegToRad(c.Pitch)))),
		float32(math.Sin(float64(mgl32.DegToRad(c.Yaw))) * math.Cos(float64(mgl32.DegToRad(c.Pitch)))),
	}
	c.Front = front.Normalize()

	// Recalculate Right and Up vectors
	c.Right = c.Front.Cross(c.WorldUp).Normalize()
	c.Up = c.Right.Cross(c.Front).Normalize()
}

func (c *Camera) SetSize(width, height int) {
	c.width = width
	c.height = height
}
