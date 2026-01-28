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

	frustum [6]mgl32.Vec4

	// GodMode flags
	FrustumFrozen bool
}

func NewCamera(width, height int) *Camera {
	c := &Camera{
		Position:         mgl32.Vec3{0, 70, 0},
		WorldUp:          mgl32.Vec3{0, 1, 0},
		Yaw:              -90.0,
		Pitch:            0.0,
		MovementSpeed:    15.0,
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

	// [GodMode] Only update frustum if NOT frozen
	if !c.FrustumFrozen {
		c.updateFrustum()
	}
	// Update Frustum Planes whenever camera moves
	c.updateFrustum()
}

func (c *Camera) SetSize(width, height int) {
	c.width = width
	c.height = height
}

// Extract the 6 planes of the view frustum
func (c *Camera) updateFrustum() {
	proj := c.GetProjectionMatrix()
	view := c.GetViewMatrix()
	clip := proj.Mul4(view)

	// Left
	c.frustum[0] = mgl32.Vec4{
		clip[3] + clip[0],
		clip[7] + clip[4],
		clip[11] + clip[8],
		clip[15] + clip[12],
	}
	// Right
	c.frustum[1] = mgl32.Vec4{
		clip[3] - clip[0],
		clip[7] - clip[4],
		clip[11] - clip[8],
		clip[15] - clip[12],
	}
	// Bottom
	c.frustum[2] = mgl32.Vec4{
		clip[3] + clip[1],
		clip[7] + clip[5],
		clip[11] + clip[9],
		clip[15] + clip[13],
	}
	// Top
	c.frustum[3] = mgl32.Vec4{
		clip[3] - clip[1],
		clip[7] - clip[5],
		clip[11] - clip[9],
		clip[15] - clip[13],
	}
	// Near
	c.frustum[4] = mgl32.Vec4{
		clip[3] + clip[2],
		clip[7] + clip[6],
		clip[11] + clip[10],
		clip[15] + clip[14],
	}
	// Far
	c.frustum[5] = mgl32.Vec4{
		clip[3] - clip[2],
		clip[7] - clip[6],
		clip[11] - clip[10],
		clip[15] - clip[14],
	}

	// Normalize planes
	for i := 0; i < 6; i++ {
		length := float32(math.Sqrt(float64(
			c.frustum[i][0]*c.frustum[i][0] +
				c.frustum[i][1]*c.frustum[i][1] +
				c.frustum[i][2]*c.frustum[i][2])))
		c.frustum[i] = c.frustum[i].Mul(1.0 / length)
	}
}

func (c *Camera) IsChunkVisible(chunkX, chunkZ int, chunkSize int) bool {
	// Chunk AABB (Axis Aligned Bounding Box)
	minX := float32(chunkX * chunkSize)
	minY := float32(0)
	minZ := float32(chunkZ * chunkSize)

	maxX := minX + float32(chunkSize)
	maxY := float32(256) // Height limit
	maxZ := minZ + float32(chunkSize)

	// Check box against all 6 planes
	for i := 0; i < 6; i++ {
		// If the box is completely behind any plane, it's invisible
		if c.frustum[i][0]*minX+c.frustum[i][1]*minY+c.frustum[i][2]*minZ+c.frustum[i][3] > 0 {
			continue
		}
		if c.frustum[i][0]*maxX+c.frustum[i][1]*minY+c.frustum[i][2]*minZ+c.frustum[i][3] > 0 {
			continue
		}
		if c.frustum[i][0]*minX+c.frustum[i][1]*maxY+c.frustum[i][2]*minZ+c.frustum[i][3] > 0 {
			continue
		}
		if c.frustum[i][0]*maxX+c.frustum[i][1]*maxY+c.frustum[i][2]*minZ+c.frustum[i][3] > 0 {
			continue
		}
		if c.frustum[i][0]*minX+c.frustum[i][1]*minY+c.frustum[i][2]*maxZ+c.frustum[i][3] > 0 {
			continue
		}
		if c.frustum[i][0]*maxX+c.frustum[i][1]*minY+c.frustum[i][2]*maxZ+c.frustum[i][3] > 0 {
			continue
		}
		if c.frustum[i][0]*minX+c.frustum[i][1]*maxY+c.frustum[i][2]*maxZ+c.frustum[i][3] > 0 {
			continue
		}
		if c.frustum[i][0]*maxX+c.frustum[i][1]*maxY+c.frustum[i][2]*maxZ+c.frustum[i][3] > 0 {
			continue
		}

		return false
	}
	return true
}
